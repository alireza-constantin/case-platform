package assembly_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/alireza-constantin/case-platform/server/internal/assembly"
	"github.com/alireza-constantin/case-platform/server/internal/cases/phone"
	"github.com/alireza-constantin/case-platform/server/internal/kernel"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Existing Phone identity and callbacks supply the rare neutral completion-only
// transition. This wrapper is test-only and registers no third runnable case.
type completionOnlyPhone struct{ phone.Module }

func (module completionOnlyPhone) Act(state json.RawMessage, tag string, payload json.RawMessage) (kernel.Transition, error) {
	transition, err := module.Module.Act(state, tag, payload)
	if err == nil && tag == "gallery.open" {
		transition.State = append(json.RawMessage(nil), state...)
	}
	return transition, err
}

type durableOutcomes struct {
	state            json.RawMessage
	events, receipts int
}

func inspectDurableOutcomes(t *testing.T, pool *pgxpool.Pool, id string) durableOutcomes {
	t.Helper()
	var result durableOutcomes
	// Read-only acceptance evidence of durable outcomes, not assertions on SQL shapes.
	err := pool.QueryRow(context.Background(), `SELECT state,
 (SELECT count(*) FROM playthrough_events WHERE playthrough_id=$1),
 (SELECT count(*) FROM action_receipts WHERE playthrough_id=$1)
 FROM playthroughs WHERE playthrough_id=$1`, id).Scan(&result.state, &result.events, &result.receipts)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestCompletionOnlyConcurrentReplayAndMonotonicContinuation(t *testing.T) {
	pool := database(t)
	srv := httptest.NewServer(kernel.NewHandler(pool, "http://localhost:5173", false, []kernel.Case{completionOnlyPhone{}}))
	defer srv.Close()
	owner := guest(t, srv.URL)
	fresh := decode(t, owner.call("POST", "/api/playthroughs", createPhone, 201))
	path := "/api/playthroughs/" + fresh.Playthrough.ID
	actions := path + "/actions"
	unlockBody := actionBody("unlock", 0, "gallery.attempt", map[string]string{"password": "light" + "house"})
	unlock := owner.call("POST", actions, unlockBody, 200)
	before := inspectDurableOutcomes(t, pool, fresh.Playthrough.ID)
	if before.events != 1 || before.receipts != 1 {
		t.Fatal("unlock must durably record its one relevant event and receipt")
	}
	completionBody := actionBody("complete-only", 1, "gallery.open", map[string]any{})
	responses := owner.concurrent(actions, []string{completionBody, completionBody, completionBody, completionBody})
	var recorded []byte
	for _, res := range responses {
		if res.status != 200 {
			t.Fatalf("completion-only duplicate returned %d", res.status)
		}
		if recorded == nil {
			recorded = res.body
		} else if !bytes.Equal(recorded, res.body) {
			t.Fatal("concurrent completion-only duplicates must return one recorded result")
		}
	}
	completed := decode(t, recorded)
	after := inspectDurableOutcomes(t, pool, fresh.Playthrough.ID)
	if completed.Playthrough.Revision != 2 || completed.Playthrough.Completed == nil || !completed.Outcome["opened"] {
		t.Fatal("first completion alone must advance once and return completion metadata")
	}
	if !bytes.Equal(before.state, after.state) || after.events != 2 || after.receipts != 2 {
		t.Fatal("completion-only must preserve case JSON while committing exactly one event and receipt")
	}
	// The original base is now stale, but its committed identity wins before revision rejection.
	replay := owner.call("POST", actions, completionBody, 200)
	if !bytes.Equal(replay, recorded) {
		t.Fatal("first-completion-only result must be replayable")
	}
	old := owner.call("POST", actions, unlockBody, 200)
	if !bytes.Equal(old, unlock) || decode(t, old).Playthrough.Revision != 1 {
		t.Fatal("older replay keeps its own safe committed snapshot")
	}
	owner.reject(actions, actionBody("complete-only", 2, "gallery.open", map[string]any{}), 409, "request_id_conflict")
	noOp := decode(t, owner.call("POST", actions, actionBody("repeat-completion", 2, "gallery.open", map[string]any{}), 200))
	continued := decode(t, owner.call("POST", actions, actionBody("post-completion", 2, "gallery.attempt", map[string]string{"password": "wrong"}), 200))
	if noOp.Playthrough.Revision != 2 || continued.Playthrough.Revision != 2 || continued.Playthrough.Completed == nil || *continued.Playthrough.Completed != *completed.Playthrough.Completed || !continued.Outcome["accepted"] {
		t.Fatal("case-allowed post-completion actions preserve the first timestamp, including a false completion signal")
	}
	retained := inspectDurableOutcomes(t, pool, fresh.Playthrough.ID)
	if !bytes.Equal(after.state, retained.state) || retained.events != 2 || retained.receipts != 2 {
		t.Fatal("old replays, no-ops and conflicts must not duplicate events, receipts or private state")
	}
	srv.Close()
	srv = httptest.NewServer(kernel.NewHandler(pool, "http://localhost:5173", false, []kernel.Case{completionOnlyPhone{}}))
	defer srv.Close()
	owner.url = srv.URL
	if !bytes.Equal(owner.call("POST", actions, completionBody, 200), recorded) {
		t.Fatal("completion-only receipt survives API restart")
	}
}

func TestCompletedPinRemovalAndRestorationPreservesDurableHistory(t *testing.T) {
	pool := database(t)
	srv := httptest.NewServer(assembly.Handler(pool, "http://localhost:5173", false))
	defer srv.Close()
	owner := guest(t, srv.URL)
	fresh := decode(t, owner.call("POST", "/api/playthroughs", createPhone, 201))
	path := "/api/playthroughs/" + fresh.Playthrough.ID
	actions := path + "/actions"
	owner.call("POST", actions, actionBody("unlock", 0, "gallery.attempt", map[string]string{"password": "light" + "house"}), 200)
	completionBody := actionBody("open", 1, "gallery.open", map[string]any{})
	completion := owner.call("POST", actions, completionBody, 200)
	completed := decode(t, completion)
	if completed.Playthrough.Completed == nil || completed.Playthrough.Revision != 2 {
		t.Fatal("Phone must be completed before removing its registration")
	}
	before := inspectDurableOutcomes(t, pool, fresh.Playthrough.ID)
	absent := httptest.NewServer(kernel.NewHandler(pool, "http://localhost:5173", false, nil))
	defer absent.Close()
	owner.url = absent.URL
	var listing struct {
		Playthroughs []kernel.Summary `json:"playthroughs"`
	}
	if err := json.Unmarshal(owner.call("GET", "/api/playthroughs", "", 200), &listing); err != nil {
		t.Fatal(err)
	}
	if len(listing.Playthroughs) != 1 {
		t.Fatal("missing registration must retain the completed run")
	}
	summary := listing.Playthroughs[0]
	if summary.ID != fresh.Playthrough.ID || summary.CaseID != "phone-demo" || summary.CaseVersion != "m0-v1" || summary.Revision != 2 || summary.CompletedAt == nil || summary.CompletedAt.Format("2006-01-02T15:04:05.999999999Z07:00") != *completed.Playthrough.Completed || summary.Availability != "unavailable" {
		t.Fatal("unavailable summary must preserve identity, pin, revision and first completion")
	}
	owner.call("GET", path, "", 409)
	owner.call("GET", path+"/assets/coastal-photo", "", 409)
	owner.reject(actions, actionBody("unavailable", 2, "gallery.open", map[string]any{}), 409, "case_version_unavailable")
	missing := inspectDurableOutcomes(t, pool, fresh.Playthrough.ID)
	if !bytes.Equal(before.state, missing.state) || missing.events != before.events || missing.receipts != before.receipts {
		t.Fatal("availability changes must not rewrite durable case state or history")
	}
	owner.url = srv.URL
	restored := decode(t, owner.call("GET", path, "", 200))
	if restored.Playthrough.ID != completed.Playthrough.ID || restored.Playthrough.CaseID != completed.Playthrough.CaseID || restored.Playthrough.Version != completed.Playthrough.Version || restored.Playthrough.Revision != 2 || restored.Playthrough.Availability != "available" || restored.Playthrough.Completed == nil || *restored.Playthrough.Completed != *completed.Playthrough.Completed || !restored.View.Unlocked {
		t.Fatal("restoring the compatible pin must resume original completion and progress")
	}
	owner.call("GET", path+"/assets/coastal-photo", "", 200)
	if !bytes.Equal(owner.call("POST", actions, completionBody, 200), completion) {
		t.Fatal("completed receipt remains intact across registration absence")
	}
	post := decode(t, owner.call("POST", actions, actionBody("after-restore", 2, "gallery.open", map[string]any{}), 200))
	if post.Playthrough.Revision != 2 || post.Playthrough.Completed == nil || *post.Playthrough.Completed != *completed.Playthrough.Completed {
		t.Fatal("restored post-completion exploration must remain a monotonic no-op")
	}
	after := inspectDurableOutcomes(t, pool, fresh.Playthrough.ID)
	if !bytes.Equal(before.state, after.state) || after.events != before.events || after.receipts != before.receipts {
		t.Fatal("restoration, receipt replay and allowed exploration retain durable history")
	}
}
