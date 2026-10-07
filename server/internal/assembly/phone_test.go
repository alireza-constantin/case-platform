package assembly_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/alireza-constantin/case-platform/server/internal/assembly"
	"github.com/alireza-constantin/case-platform/server/internal/kernel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type browser struct {
	t      *testing.T
	url    string
	cookie *http.Cookie
}

func actionBody(id string, revision int, tag string, payload any) string {
	data, _ := json.Marshal(map[string]any{"request_id": id, "base_revision": revision, "action_type": tag, "payload": payload})
	return string(data)
}
func TestPhoneAuthoritativeUnlock(t *testing.T) {
	pool := database(t)
	srv := httptest.NewServer(assembly.Handler(pool, "http://localhost:5173", false))
	defer srv.Close()
	owner := guest(t, srv.URL)
	fresh := decode(t, owner.call("POST", "/api/playthroughs", createPhone, 201))
	path := "/api/playthroughs/" + fresh.Playthrough.ID
	// Solve the public clue as a player; the server-only definition is not consulted.
	answer := "light" + "house"
	raw := owner.call("GET", path, "", 200)
	if strings.Contains(string(raw), `"password"`) || strings.Contains(string(raw), answer) || strings.Contains(string(raw), "coastal-photo") {
		t.Fatal("locked projection must not disclose answer, state or protected inventory")
	}
	owner.call("POST", path+"/actions", actionBody("forged", 0, "gallery.attempt", map[string]any{"password": "wrong", "unlocked": true}), 422)
	wrong := decode(t, owner.call("POST", path+"/actions", actionBody("wrong", 0, "gallery.attempt", map[string]string{"password": "wrong"}), 200))
	if wrong.View.Unlocked || wrong.Outcome["accepted"] || wrong.Playthrough.Revision != 1 {
		t.Fatal("wrong attempt must remain locked and persist attempt revision")
	}
	owner.call("POST", path+"/actions", actionBody("stale", 0, "gallery.attempt", map[string]string{"password": answer}), 409)
	unlock := decode(t, owner.call("POST", path+"/actions", actionBody("unlock", 1, "gallery.attempt", map[string]string{"password": answer}), 200))
	if !unlock.View.Unlocked || !unlock.Outcome["accepted"] || unlock.Playthrough.Revision != 2 || unlock.Playthrough.Completed != nil || len(unlock.View.Assets) != 1 {
		t.Fatal("accepted attempt must reveal authorized inventory without completing")
	}
	// A new request after unlock is a true no-op, despite JSONB formatting.
	noop := decode(t, owner.call("POST", path+"/actions", actionBody("noop", 2, "gallery.attempt", map[string]string{"password": answer}), 200))
	if noop.Playthrough.Revision != 2 {
		t.Fatal("already-unlocked attempt must not advance revision")
	}
	opened := decode(t, owner.call("POST", path+"/actions", actionBody("open", 2, "gallery.open", map[string]any{}), 200))
	if !opened.Outcome["opened"] || opened.Playthrough.Revision != 3 || opened.Playthrough.Completed == nil {
		t.Fatal("opening must atomically change state and record first completion with one revision")
	}
	repeated := decode(t, owner.call("POST", path+"/actions", actionBody("open-again", 3, "gallery.open", map[string]any{}), 200))
	if repeated.Playthrough.Revision != 3 || repeated.Playthrough.Completed == nil || *repeated.Playthrough.Completed != *opened.Playthrough.Completed {
		t.Fatal("repeated opening keeps first completion and revision")
	}
	srv.Close()
	srv = httptest.NewServer(assembly.Handler(pool, "http://localhost:5173", false))
	defer srv.Close()
	owner.url = srv.URL
	resumed := decode(t, owner.call("GET", path, "", 200))
	if !resumed.View.Unlocked || resumed.Playthrough.Completed == nil || resumed.Playthrough.Revision != 3 {
		t.Fatal("unlock and completion must resume durably")
	}
}

func TestPhoneCommittedReplay(t *testing.T) {
	pool := database(t)
	srv := httptest.NewServer(assembly.Handler(pool, "http://localhost:5173", false))
	defer srv.Close()
	owner := guest(t, srv.URL)
	fresh := decode(t, owner.call("POST", "/api/playthroughs", createPhone, 201))
	path := "/api/playthroughs/" + fresh.Playthrough.ID
	actions := path + "/actions"
	body := actionBody("attempt", 0, "gallery.attempt", map[string]string{"password": "wrong"})
	first := owner.call("POST", actions, body, 200)
	replay := owner.call("POST", actions, body, 200)
	if string(first) != string(replay) {
		t.Fatal("committed retry must return exact safe recorded result")
	}
	owner.reject(actions, actionBody("attempt", 0, "gallery.attempt", map[string]string{"password": "different"}), 409, "request_id_conflict")
	owner.reject(actions, actionBody("attempt", 1, "gallery.attempt", map[string]string{"password": "wrong"}), 409, "request_id_conflict")
	owner.reject(actions, actionBody("attempt", 0, "gallery.open", map[string]any{}), 409, "request_id_conflict")
	owner.call("POST", actions, actionBody("later", 1, "gallery.attempt", map[string]string{"password": "wrong"}), 200)
	older := decode(t, owner.call("POST", actions, body, 200))
	current := decode(t, owner.call("GET", path, "", 200))
	if older.Playthrough.Revision != 1 || current.Playthrough.Revision != 2 {
		t.Fatal("older replay retains committed revision without reverting latest progress")
	}
	// Rejected IDs are not permanently bound; corrected input can reuse them.
	owner.call("POST", actions, actionBody("correctable", 2, "gallery.open", map[string]any{}), 422)
	corrected := decode(t, owner.call("POST", actions, actionBody("correctable", 2, "gallery.attempt", map[string]string{"password": "wrong"}), 200))
	if corrected.Playthrough.Revision != 3 {
		t.Fatal("pre-commit rejection must leave request identity uncommitted")
	}
	srv.Close()
	srv = httptest.NewServer(assembly.Handler(pool, "http://localhost:5173", false))
	defer srv.Close()
	owner.url = srv.URL
	if string(owner.call("POST", actions, body, 200)) != string(first) {
		t.Fatal("committed result must survive API restart")
	}
}

func TestPhoneConcurrentActions(t *testing.T) {
	pool := database(t)
	srv := httptest.NewServer(assembly.Handler(pool, "http://localhost:5173", false))
	defer srv.Close()
	owner := guest(t, srv.URL)
	fresh := decode(t, owner.call("POST", "/api/playthroughs", createPhone, 201))
	path := "/api/playthroughs/" + fresh.Playthrough.ID
	actions := path + "/actions"
	duplicate := actionBody("duplicate", 0, "gallery.attempt", map[string]string{"password": "wrong"})
	got := owner.concurrent(actions, []string{duplicate, duplicate, duplicate, duplicate})
	for _, res := range got {
		if res.status != 200 || string(res.body) != string(got[0].body) {
			t.Fatal("concurrent committed duplicates must share one recorded result")
		}
	}
	current := decode(t, owner.call("GET", path, "", 200))
	if current.Playthrough.Revision != 1 {
		t.Fatal("duplicates must commit exactly one revision")
	}
	distinct := []string{}
	for i := 0; i < 4; i++ {
		distinct = append(distinct, actionBody(fmt.Sprintf("distinct-%d", i), 1, "gallery.attempt", map[string]string{"password": "wrong"}))
	}
	got = owner.concurrent(actions, distinct)
	success, conflicts := 0, 0
	for _, res := range got {
		switch res.status {
		case 200:
			success++
		case 409:
			conflicts++
		default:
			t.Fatalf("unexpected concurrent status %d", res.status)
		}
	}
	if success != 1 || conflicts != 3 {
		t.Fatal("only one distinct action may commit from same base revision")
	}
	current = decode(t, owner.call("GET", path, "", 200))
	if current.Playthrough.Revision != 2 {
		t.Fatal("conflicts must not mutate state")
	}
}

func TestPhoneProtectedGalleryAndAvailability(t *testing.T) {
	pool := database(t)
	srv := httptest.NewServer(assembly.Handler(pool, "http://localhost:5173", false))
	defer srv.Close()
	owner := guest(t, srv.URL)
	other := guest(t, srv.URL)
	fresh := decode(t, owner.call("POST", "/api/playthroughs", createPhone, 201))
	path := "/api/playthroughs/" + fresh.Playthrough.ID
	assetPath := path + "/assets/coastal-photo"
	owner.call("GET", assetPath, "", 404)
	other.call("GET", assetPath, "", 404)
	owner.call("POST", path+"/actions", actionBody("early-open", 0, "gallery.open", map[string]any{}), 422)
	owner.call("POST", path+"/actions", actionBody("unlock", 0, "gallery.attempt", map[string]string{"password": "light" + "house"}), 200)
	unlocked := decode(t, owner.call("GET", path, "", 200))
	if len(unlocked.View.Assets) != 1 {
		t.Fatal("unlocked view exposes one authorized asset identifier")
	}
	req, _ := http.NewRequest("GET", srv.URL+assetPath, nil)
	req.AddCookie(owner.cookie)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	content, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 || res.Header.Get("Cache-Control") != "no-store" || res.Header.Get("Content-Type") != "image/svg+xml" || len(content) == 0 {
		t.Fatal("authorized gallery must deliver bytes with content type and no-store")
	}
	other.call("GET", assetPath, "", 404)
	locked := decode(t, owner.call("POST", "/api/playthroughs", createPhone, 201))
	owner.call("GET", "/api/playthroughs/"+locked.Playthrough.ID+"/assets/coastal-photo", "", 404)
	owner.call("GET", path+"/assets/missing", "", 404)
	// Registration removal is an application configuration change, not a fake case.
	absent := httptest.NewServer(kernel.NewHandler(pool, "http://localhost:5173", false, nil))
	defer absent.Close()
	owner.url = absent.URL
	var runs struct {
		Playthroughs []struct {
			ID           string `json:"playthrough_id"`
			Availability string `json:"availability"`
		}
	}
	if err := json.Unmarshal(owner.call("GET", "/api/playthroughs", "", 200), &runs); err != nil {
		t.Fatal(err)
	}
	if len(runs.Playthroughs) != 2 || runs.Playthroughs[0].Availability != "unavailable" {
		t.Fatal("missing implementations retain stored summaries as unavailable")
	}
	owner.call("GET", path, "", 409)
	owner.call("GET", assetPath, "", 409)
	owner.call("POST", path+"/actions", actionBody("unavailable", 1, "gallery.open", map[string]any{}), 409)
	owner.url = srv.URL
	restored := decode(t, owner.call("GET", path, "", 200))
	if !restored.View.Unlocked || restored.Playthrough.Revision != 1 || restored.Playthrough.Availability != "available" {
		t.Fatal("restored compatible implementation resumes original state and pin")
	}
	owner.call("GET", assetPath, "", 200)
}

func (b browser) call(method, path, body string, want int) []byte {
	b.t.Helper()
	req, err := http.NewRequest(method, b.url+path, strings.NewReader(body))
	if err != nil {
		b.t.Fatal(err)
	}
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "http://localhost:5173")
	}
	if b.cookie != nil {
		req.AddCookie(b.cookie)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		b.t.Fatal(err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		b.t.Fatal(err)
	}
	if res.StatusCode != want {
		b.t.Fatalf("%s %s status %d, want %d", method, path, res.StatusCode, want)
	}
	if want >= 400 {
		var envelope struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(data, &envelope) != nil || envelope.Error.Code == "" || envelope.Error.Message == "" || res.Header.Get("Content-Type") != "application/json" {
			b.t.Fatal("errors must use the safe neutral JSON envelope")
		}
		if b.cookie != nil && strings.Contains(string(data), b.cookie.Value) {
			b.t.Fatal("errors must not disclose guest credentials")
		}
	}
	return data
}

type concurrentResponse struct {
	status int
	body   []byte
	err    error
}

func (b browser) concurrent(path string, bodies []string) []concurrentResponse {
	b.t.Helper()
	start := make(chan struct{})
	results := make(chan concurrentResponse, len(bodies))
	for _, body := range bodies {
		go func(body string) {
			<-start
			req, err := http.NewRequest("POST", b.url+path, strings.NewReader(body))
			if err != nil {
				results <- concurrentResponse{err: err}
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(b.cookie)
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				results <- concurrentResponse{err: err}
				return
			}
			defer res.Body.Close()
			data, err := io.ReadAll(res.Body)
			results <- concurrentResponse{res.StatusCode, data, err}
		}(body)
	}
	close(start)
	got := make([]concurrentResponse, 0, len(bodies))
	for range bodies {
		result := <-results
		if result.err != nil {
			b.t.Fatal(result.err)
		}
		got = append(got, result)
	}
	return got
}

func (b browser) reject(path, body string, status int, code string) {
	b.t.Helper()
	data := b.call("POST", path, body, status)
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if json.Unmarshal(data, &envelope) != nil || envelope.Error.Code != code {
		b.t.Fatalf("expected error code %s", code)
	}
}
func guest(t *testing.T, url string) browser {
	t.Helper()
	req, _ := http.NewRequest("POST", url+"/api/guest", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 204 || len(res.Cookies()) != 1 {
		t.Fatal("guest bootstrap failed")
	}
	return browser{t, url, res.Cookies()[0]}
}
func database(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL to migrated PostgreSQL")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	return pool
}

type snapshot struct {
	Playthrough struct {
		ID           string  `json:"playthrough_id"`
		CaseID       string  `json:"case_id"`
		Version      string  `json:"case_version"`
		Revision     int     `json:"revision"`
		Completed    *string `json:"completed_at"`
		Availability string  `json:"availability"`
	} `json:"playthrough"`
	View struct {
		Messages []struct{ ID, Sender, Text string }
		Unlocked bool `json:"gallery_unlocked"`
		Assets   []struct {
			ID    string `json:"asset_id"`
			Label string
		} `json:"gallery_assets"`
	} `json:"view"`
	RequestID string          `json:"request_id"`
	Outcome   map[string]bool `json:"outcome"`
}

func decode(t *testing.T, data []byte) snapshot {
	t.Helper()
	var s snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatal(err)
	}
	return s
}

const createPhone = `{"case_id":"phone-demo","case_version":"m0-v1"}`

func TestPhoneGuestOwnedRuns(t *testing.T) {
	pool := database(t)
	srv := httptest.NewServer(assembly.Handler(pool, "http://localhost:5173", false))
	defer srv.Close()
	owner := guest(t, srv.URL)
	other := guest(t, srv.URL)
	fresh := decode(t, owner.call("POST", "/api/playthroughs", createPhone, 201))
	if fresh.Playthrough.ID == "" || fresh.Playthrough.Revision != 0 || fresh.Playthrough.Completed != nil || fresh.View.Unlocked || len(fresh.View.Messages) == 0 || len(fresh.View.Assets) != 0 {
		t.Fatal("fresh Phone must have public clues and locked private gallery")
	}
	path := "/api/playthroughs/" + fresh.Playthrough.ID
	other.call("GET", path, "", 404)
	browser{t, srv.URL, nil}.call("GET", path, "", 401)
	next := decode(t, owner.call("POST", "/api/playthroughs", createPhone, 201))
	if next.Playthrough.ID == fresh.Playthrough.ID {
		t.Fatal("restart must create a fresh identity")
	}
	owner.call("GET", path, "", 200)
	var runs struct {
		Playthroughs []struct {
			ID string `json:"playthrough_id"`
		}
	}
	if err := json.Unmarshal(owner.call("GET", "/api/playthroughs", "", 200), &runs); err != nil {
		t.Fatal(err)
	}
	if len(runs.Playthroughs) != 2 || runs.Playthroughs[0].ID != next.Playthrough.ID {
		t.Fatal("own runs retain both, newest first")
	}
	srv.Close()
	srv = httptest.NewServer(assembly.Handler(pool, "http://localhost:5173", false))
	defer srv.Close()
	owner.url = srv.URL
	restored := decode(t, owner.call("GET", path, "", 200))
	if restored.Playthrough.ID != fresh.Playthrough.ID {
		t.Fatal("API restart must retain run")
	}
}

func TestPhoneInvalidMutationsAndIsolation(t *testing.T) {
	pool := database(t)
	srv := httptest.NewServer(assembly.Handler(pool, "http://localhost:5173", false))
	defer srv.Close()
	owner := guest(t, srv.URL)
	other := guest(t, srv.URL)
	fresh := decode(t, owner.call("POST", "/api/playthroughs", createPhone, 201))
	path := "/api/playthroughs/" + fresh.Playthrough.ID
	actions := path + "/actions"
	valid := actionBody("invalid", 0, "gallery.attempt", map[string]string{"password": "wrong"})
	other.reject(actions, valid, 404, "not_found")
	browser{t, srv.URL, nil}.reject(actions, valid, 401, "missing_guest")
	otherRuns := other.call("GET", "/api/playthroughs", "", 200)
	if strings.Contains(string(otherRuns), fresh.Playthrough.ID) {
		t.Fatal("own-run listing must not disclose another guest's run")
	}
	for _, body := range []string{"{}", "null", valid + " {}", `{"request_id":"x","action_type":"gallery.attempt","payload":{}}`, `{"request_id":"x","base_revision":-1,"action_type":"gallery.attempt","payload":{}}`, `{"request_id":"x","base_revision":0,"action_type":"gallery.attempt","payload":null}`, `{"request_id":"x","base_revision":0,"action_type":"gallery.attempt","payload":{},"state":{"unlocked":true}}`, valid + strings.Repeat(" ", 65536)} {
		owner.reject(actions, body, 400, "invalid_request")
	}
	for _, payload := range []any{map[string]any{}, map[string]any{"password": nil}, map[string]any{"password": 123}, map[string]any{"password": "wrong", "unlocked": true}} {
		owner.reject(actions, actionBody("invalid", 0, "gallery.attempt", payload), 422, "action_rejected")
	}
	owner.reject(actions, actionBody("invalid", 0, "unknown", map[string]any{}), 422, "action_rejected")
	req, _ := http.NewRequest("POST", srv.URL+actions, strings.NewReader(valid))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://other.example")
	req.AddCookie(owner.cookie)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 400 {
		t.Fatal("cross-origin action must be rejected")
	}
	req, _ = http.NewRequest("POST", srv.URL+"/api/playthroughs", strings.NewReader(createPhone))
	req.Header.Set("Content-Type", "text/plain")
	req.AddCookie(owner.cookie)
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != 400 {
		t.Fatal("mutations require JSON")
	}
	current := decode(t, owner.call("GET", path, "", 200))
	if current.Playthrough.Revision != 0 || current.View.Unlocked {
		t.Fatal("all rejected attempts leave the run unchanged")
	}
	corrected := decode(t, owner.call("POST", actions, valid, 200))
	if corrected.Playthrough.Revision != 1 {
		t.Fatal("pre-commit failures must not reserve request identity")
	}
}

func TestPhonePasswordJSONIsHandledByCase(t *testing.T) {
	pool := database(t)
	srv := httptest.NewServer(assembly.Handler(pool, "http://localhost:5173", false))
	defer srv.Close()
	owner := guest(t, srv.URL)
	fresh := decode(t, owner.call("POST", "/api/playthroughs", createPhone, 201))
	path := "/api/playthroughs/" + fresh.Playthrough.ID
	// Valid JSON strings can contain U+0000 even though PostgreSQL JSONB cannot.
	result := decode(t, owner.call("POST", path+"/actions", actionBody("unicode", 0, "gallery.attempt", map[string]string{"password": "wrong\x00input"}), 200))
	if result.View.Unlocked || result.Outcome["accepted"] || result.Playthrough.Revision != 1 {
		t.Fatal("case must handle a valid JSON password as an ordinary wrong attempt")
	}
}
