package kernel

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type casePin struct{ id, version string }
type Summary struct {
	ID           string     `json:"playthrough_id"`
	CaseID       string     `json:"case_id"`
	CaseVersion  string     `json:"case_version"`
	Revision     int64      `json:"revision"`
	CompletedAt  *time.Time `json:"completed_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Availability string     `json:"availability"`
}
type Snapshot struct {
	Playthrough Summary         `json:"playthrough"`
	View        json.RawMessage `json:"view"`
}
type storedRun struct {
	summary Summary
	state   json.RawMessage
}

const runColumns = "playthrough_id, case_id, case_version, revision, completed_at, created_at, updated_at, state"

func scanRun(row pgx.Row) (storedRun, error) {
	var run storedRun
	s := &run.summary
	err := row.Scan(&s.ID, &s.CaseID, &s.CaseVersion, &s.Revision, &s.CompletedAt, &s.CreatedAt, &s.UpdatedAt, &run.state)
	s.CreatedAt = s.CreatedAt.UTC()
	s.UpdatedAt = s.UpdatedAt.UTC()
	if s.CompletedAt != nil {
		utc := s.CompletedAt.UTC()
		s.CompletedAt = &utc
	}
	return run, err
}
func (run storedRun) module(registry map[casePin]Case) Case {
	return registry[casePin{run.summary.CaseID, run.summary.CaseVersion}]
}
func (run storedRun) snapshot(registry map[casePin]Case) (Snapshot, error) {
	module := run.module(registry)
	if module == nil {
		return Snapshot{}, errors.New("unavailable")
	}
	view, err := module.Project(run.state)
	run.summary.Availability = "available"
	return Snapshot{run.summary, view}, err
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		writeError(w, 500, "internal_error", "Request could not be completed.")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	w.Write(data)
}
func mutationJSON(w http.ResponseWriter, r *http.Request, origin string, value any) bool {
	if supplied := r.Header.Get("Origin"); supplied != "" && supplied != origin {
		writeError(w, 400, "invalid_request", "Use the application origin.")
		return false
	}
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		writeError(w, 400, "invalid_request", "Send a JSON object.")
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		writeError(w, 400, "invalid_request", "Invalid JSON request.")
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		writeError(w, 400, "invalid_request", "Send one JSON object.")
		return false
	}
	return true
}
func guestToken(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool) (string, bool) {
	cookie, err := r.Cookie("guest_session")
	if err != nil || len(cookie.Value) != 64 {
		writeError(w, 401, "missing_guest", "Establish a guest first.")
		return "", false
	}
	var exists bool
	if err := db.QueryRow(r.Context(), "SELECT EXISTS (SELECT 1 FROM guests WHERE session_token=$1)", cookie.Value).Scan(&exists); err != nil {
		writeError(w, 500, "internal_error", "Request could not be completed.")
		return "", false
	}
	if !exists {
		writeError(w, 401, "missing_guest", "Establish a guest first.")
		return "", false
	}
	return cookie.Value, true
}
func internalError(w http.ResponseWriter) {
	writeError(w, 500, "internal_error", "Request could not be completed.")
}
func ownedRun(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool, token string) (storedRun, bool) {
	run, err := scanRun(db.QueryRow(r.Context(), "SELECT "+runColumns+" FROM playthroughs WHERE playthrough_id=$1 AND guest_token=$2", r.PathValue("playthrough_id"), token))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, 404, "not_found", "Resource unavailable.")
		return run, false
	}
	if err != nil {
		internalError(w)
		return run, false
	}
	return run, true
}
func registerPlaythroughs(mux *http.ServeMux, db *pgxpool.Pool, origin string, registry map[casePin]Case) {
	registerActions(mux, db, origin, registry)
	mux.HandleFunc("GET /api/playthroughs/{playthrough_id}/assets/{asset_id}", func(w http.ResponseWriter, r *http.Request) {
		r, cancel := boundedRequest(r)
		defer cancel()
		token, ok := guestToken(w, r, db)
		if !ok {
			return
		}
		run, ok := ownedRun(w, r, db, token)
		if !ok {
			return
		}
		module := run.module(registry)
		if module == nil {
			writeError(w, 409, "case_version_unavailable", "Case version unavailable.")
			return
		}
		asset, allowed, err := module.Asset(run.state, r.PathValue("asset_id"))
		if err != nil {
			internalError(w)
			return
		}
		if !allowed {
			writeError(w, 404, "not_found", "Resource unavailable.")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", asset.ContentType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(200)
		w.Write(asset.Bytes)
	})
	mux.HandleFunc("POST /api/playthroughs", func(w http.ResponseWriter, r *http.Request) {
		r, cancel := boundedRequest(r)
		defer cancel()
		var body struct {
			CaseID  string `json:"case_id"`
			Version string `json:"case_version"`
		}
		if !mutationJSON(w, r, origin, &body) {
			return
		}
		if body.CaseID == "" || body.Version == "" {
			writeError(w, 400, "invalid_request", "Choose a case and version.")
			return
		}
		token, ok := guestToken(w, r, db)
		if !ok {
			return
		}
		module := registry[casePin{body.CaseID, body.Version}]
		if module == nil {
			writeError(w, 409, "case_version_unavailable", "Case version unavailable.")
			return
		}
		state, err := module.Initialize()
		if err != nil || !json.Valid(state) {
			internalError(w)
			return
		}
		bytes := make([]byte, 16)
		if _, err := rand.Read(bytes); err != nil {
			internalError(w)
			return
		}
		id := hex.EncodeToString(bytes)
		run, err := scanRun(db.QueryRow(r.Context(), "INSERT INTO playthroughs (playthrough_id,guest_token,case_id,case_version,state) VALUES ($1,$2,$3,$4,$5) RETURNING "+runColumns, id, token, body.CaseID, body.Version, state))
		if err != nil {
			internalError(w)
			return
		}
		snapshot, err := run.snapshot(registry)
		if err != nil {
			internalError(w)
			return
		}
		writeJSON(w, 201, snapshot)
	})
	mux.HandleFunc("GET /api/playthroughs", func(w http.ResponseWriter, r *http.Request) {
		r, cancel := boundedRequest(r)
		defer cancel()
		token, ok := guestToken(w, r, db)
		if !ok {
			return
		}
		rows, err := db.Query(r.Context(), "SELECT "+runColumns+" FROM playthroughs WHERE guest_token=$1 ORDER BY created_at DESC,playthrough_id DESC", token)
		if err != nil {
			internalError(w)
			return
		}
		defer rows.Close()
		summaries := []Summary{}
		for rows.Next() {
			run, err := scanRun(rows)
			if err != nil {
				internalError(w)
				return
			}
			run.summary.Availability = "available"
			if run.module(registry) == nil {
				run.summary.Availability = "unavailable"
			}
			summaries = append(summaries, run.summary)
		}
		if rows.Err() != nil {
			internalError(w)
			return
		}
		writeJSON(w, 200, struct {
			Playthroughs []Summary `json:"playthroughs"`
		}{summaries})
	})
	mux.HandleFunc("GET /api/playthroughs/{playthrough_id}", func(w http.ResponseWriter, r *http.Request) {
		r, cancel := boundedRequest(r)
		defer cancel()
		token, ok := guestToken(w, r, db)
		if !ok {
			return
		}
		run, ok := ownedRun(w, r, db, token)
		if !ok {
			return
		}
		if run.module(registry) == nil {
			writeError(w, 409, "case_version_unavailable", "Case version unavailable.")
			return
		}
		snapshot, err := run.snapshot(registry)
		if err != nil {
			internalError(w)
			return
		}
		writeJSON(w, 200, snapshot)
	})
}

// boundedRequest supplies cooperative cancellation to database work.
func boundedRequest(r *http.Request) (*http.Request, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	return r.WithContext(ctx), cancel
}
