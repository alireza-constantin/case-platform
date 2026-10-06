package kernel

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CaseMetadata struct {
	CaseID      string `json:"case_id"`
	CaseVersion string `json:"case_version"`
	Title       string `json:"title"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewHandler receives safe catalogue registrations only from application assembly.
func NewHandler(db *pgxpool.Pool, origin string, secureCookie bool, modules []Case) http.Handler {
	cases := []CaseMetadata{}
	registrations := make(map[casePin]Case)
	for _, module := range modules {
		metadata := module.Metadata()
		pin := casePin{metadata.CaseID, metadata.CaseVersion}
		if metadata.CaseID == "" || metadata.CaseVersion == "" || registrations[pin] != nil {
			panic("invalid or duplicate case registration")
		}
		registrations[pin] = module
		cases = append(cases, metadata)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/guest", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if supplied := r.Header.Get("Origin"); supplied != "" && supplied != origin {
			writeError(w, 400, "invalid_request", "Use the application origin.")
			return
		}
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			writeError(w, 400, "invalid_request", "Send an empty JSON object.")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		decoder := json.NewDecoder(r.Body)
		var body map[string]json.RawMessage
		if err := decoder.Decode(&body); err != nil || body == nil || len(body) != 0 {
			writeError(w, 400, "invalid_request", "Send an empty JSON object.")
			return
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			writeError(w, 400, "invalid_request", "Send one empty JSON object.")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		token := ""
		if cookie, err := r.Cookie("guest_session"); err == nil && len(cookie.Value) == 64 {
			var exists bool
			err := db.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM guests WHERE session_token = $1)", cookie.Value).Scan(&exists)
			if err != nil {
				writeError(w, 500, "internal_error", "Guest access could not be established.")
				return
			}
			if exists {
				token = cookie.Value
			}
		}
		if token == "" {
			bytes := make([]byte, 32)
			if _, err := rand.Read(bytes); err != nil {
				writeError(w, 500, "internal_error", "Guest access could not be established.")
				return
			}
			token = hex.EncodeToString(bytes)
			if _, err := db.Exec(ctx, "INSERT INTO guests (session_token) VALUES ($1)", token); err != nil {
				writeError(w, 500, "internal_error", "Guest access could not be established.")
				return
			}
		}
		http.SetCookie(w, &http.Cookie{Name: "guest_session", Value: token, Path: "/", HttpOnly: true, Secure: secureCookie, SameSite: http.SameSiteLaxMode, MaxAge: 365 * 24 * 60 * 60, Expires: time.Now().Add(365 * 24 * time.Hour)})
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /api/cases", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		json.NewEncoder(w).Encode(struct {
			Cases []CaseMetadata `json:"cases"`
		}{Cases: cases})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, 404, "not_found", "Resource unavailable.")
	})
	registerPlaythroughs(mux, db, origin, registrations)
	return mux
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(struct {
		Error errorDetail `json:"error"`
	}{Error: errorDetail{Code: code, Message: message}})
}
