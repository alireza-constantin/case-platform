package kernel

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type actionRequest struct {
	RequestID    string          `json:"request_id"`
	BaseRevision *int64          `json:"base_revision"`
	Tag          string          `json:"action_type"`
	Payload      json.RawMessage `json:"payload"`
}
type ActionResult struct {
	Snapshot
	RequestID string          `json:"request_id"`
	Outcome   json.RawMessage `json:"outcome"`
}

func registerActions(mux *http.ServeMux, db *pgxpool.Pool, origin string, registry map[casePin]Case) {
	mux.HandleFunc("POST /api/playthroughs/{playthrough_id}/actions", func(w http.ResponseWriter, r *http.Request) {
		r, cancel := boundedRequest(r)
		defer cancel()
		var body actionRequest
		if !mutationJSON(w, r, origin, &body) {
			return
		}
		if body.RequestID == "" || len(body.RequestID) > 200 || body.BaseRevision == nil || *body.BaseRevision < 0 || body.Tag == "" || len(body.Tag) > 200 || len(body.Payload) == 0 || string(body.Payload) == "null" {
			writeError(w, 400, "invalid_request", "Invalid action envelope.")
			return
		}
		token, ok := guestToken(w, r, db)
		if !ok {
			return
		}
		tx, err := db.Begin(r.Context())
		if err != nil {
			internalError(w)
			return
		}
		defer tx.Rollback(r.Context())
		run, err := scanRun(tx.QueryRow(r.Context(), "SELECT "+runColumns+" FROM playthroughs WHERE playthrough_id=$1 AND guest_token=$2 FOR UPDATE", r.PathValue("playthrough_id"), token))
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, 404, "not_found", "Resource unavailable.")
			return
		}
		if err != nil {
			internalError(w)
			return
		}
		module := run.module(registry)
		if module == nil {
			writeError(w, 409, "case_version_unavailable", "Case version unavailable.")
			return
		}
		// Canonicalize JSON in Go: action input need not be representable in PostgreSQL JSONB.
		var payload any
		decoder := json.NewDecoder(bytes.NewReader(body.Payload))
		decoder.UseNumber()
		if err := decoder.Decode(&payload); err != nil {
			writeError(w, 400, "invalid_request", "Invalid action payload.")
			return
		}
		input, err := json.Marshal(struct {
			BaseRevision int64  `json:"base_revision"`
			Tag          string `json:"action_type"`
			Payload      any    `json:"payload"`
		}{*body.BaseRevision, body.Tag, payload})
		if err != nil {
			internalError(w)
			return
		}
		identity := sha256.Sum256(input)
		var recordedIdentity, recordedResult []byte
		err = tx.QueryRow(r.Context(), "SELECT input_hash,result FROM action_receipts WHERE playthrough_id=$1 AND request_id=$2", run.summary.ID, body.RequestID).Scan(&recordedIdentity, &recordedResult)
		if err == nil {
			if !bytes.Equal(recordedIdentity, identity[:]) {
				writeError(w, 409, "request_id_conflict", "Request ID already committed with different input.")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "no-store")
			w.WriteHeader(200)
			w.Write(recordedResult)
			return
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			internalError(w)
			return
		}
		if *body.BaseRevision != run.summary.Revision {
			writeError(w, 409, "revision_conflict", "Fetch the current view before submitting a new action.")
			return
		}
		transition, err := module.Act(run.state, body.Tag, body.Payload)
		if errors.Is(err, ErrActionRejected) {
			writeError(w, 422, "action_rejected", "Case action rejected.")
			return
		}
		if err != nil || !json.Valid(transition.State) || !json.Valid(transition.Outcome) {
			internalError(w)
			return
		}
		var changed bool
		if err := tx.QueryRow(r.Context(), "SELECT $1::jsonb IS DISTINCT FROM $2::jsonb", run.state, transition.State).Scan(&changed); err != nil {
			internalError(w)
			return
		}
		changed = changed || (transition.Complete && run.summary.CompletedAt == nil)
		if changed {
			run, err = scanRun(tx.QueryRow(r.Context(), "UPDATE playthroughs SET state=$1, revision=revision+1, completed_at=CASE WHEN $2 THEN COALESCE(completed_at,clock_timestamp()) ELSE completed_at END,updated_at=clock_timestamp() WHERE playthrough_id=$3 RETURNING "+runColumns, transition.State, transition.Complete, run.summary.ID))
			if err != nil {
				internalError(w)
				return
			}
			for _, event := range transition.Events {
				if !json.Valid(event) {
					internalError(w)
					return
				}
				if _, err := tx.Exec(r.Context(), "INSERT INTO playthrough_events(playthrough_id,revision,event) VALUES($1,$2,$3)", run.summary.ID, run.summary.Revision, event); err != nil {
					internalError(w)
					return
				}
			}
		}
		snapshot, err := run.snapshot(registry)
		if err != nil {
			internalError(w)
			return
		}
		result := ActionResult{snapshot, body.RequestID, transition.Outcome}
		if changed {
			encoded, err := json.Marshal(result)
			if err != nil {
				internalError(w)
				return
			}
			if _, err := tx.Exec(r.Context(), "INSERT INTO action_receipts(playthrough_id,request_id,input_hash,result) VALUES($1,$2,$3,$4)", run.summary.ID, body.RequestID, identity[:], encoded); err != nil {
				internalError(w)
				return
			}
		}
		if err := tx.Commit(r.Context()); err != nil {
			internalError(w)
			return
		}
		writeJSON(w, 200, result)
	})
}
