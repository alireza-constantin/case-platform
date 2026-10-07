package assembly_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alireza-constantin/case-platform/server/internal/assembly"
	"github.com/alireza-constantin/case-platform/server/internal/kernel"
)

const createTerminal = `{"case_id":"terminal-demo","case_version":"m0-v1"}`

type terminalSnapshot struct {
	Playthrough kernel.Summary `json:"playthrough"`
	View        struct {
		Cwd     string `json:"cwd"`
		Entries []struct {
			Name string `json:"name"`
			Kind string `json:"kind"`
		} `json:"entries"`
		Ready         bool `json:"workstation_ready"`
		ProtectedFile *struct {
			ID   string `json:"asset_id"`
			Name string `json:"name"`
		} `json:"protected_file"`
	} `json:"view"`
	Outcome struct {
		Lines []string `json:"lines"`
		OK    bool     `json:"ok"`
	} `json:"outcome"`
}

func terminalDecode(t *testing.T, data []byte) terminalSnapshot {
	t.Helper()
	var s terminalSnapshot
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatal(err)
	}
	return s
}
func terminalCommand(owner browser, path, id string, revision int, command string) terminalSnapshot {
	owner.t.Helper()
	return terminalDecode(owner.t, owner.call("POST", path+"/actions", actionBody(id, revision, "workstation.command", map[string]string{"command": command}), 200))
}
func TestTerminalExploreSimulatedWorkstation(t *testing.T) {
	pool := database(t)
	srv := httptest.NewServer(assembly.Handler(pool, "http://localhost:5173", false))
	defer srv.Close()
	owner := guest(t, srv.URL)
	fresh := terminalDecode(t, owner.call("POST", "/api/playthroughs", createTerminal, 201))
	path := "/api/playthroughs/" + fresh.Playthrough.ID
	if fresh.View.Cwd != "/" || fresh.View.Ready || fresh.View.ProtectedFile != nil || len(fresh.View.Entries) != 2 {
		t.Fatal("initial Terminal must expose safe root navigation without protected inventory")
	}
	help := terminalCommand(owner, path, "help", 0, "help")
	if !help.Outcome.OK || len(help.Outcome.Lines) == 0 || help.Playthrough.Revision != 0 {
		t.Fatal("help is an informative no-op")
	}
	listing := terminalCommand(owner, path, "list", 0, "ls")
	if !listing.Outcome.OK || strings.Join(listing.Outcome.Lines, ",") != "readme.txt,maintenance" || listing.Playthrough.Revision != 0 {
		t.Fatal("listing exposes only safe current-directory entries")
	}
	pwd := terminalCommand(owner, path, "location", 0, "pwd")
	if !pwd.Outcome.OK || strings.Join(pwd.Outcome.Lines, "") != "/" || pwd.Playthrough.Revision != 0 {
		t.Fatal("location command reads the current directory without mutation")
	}
	readme := terminalCommand(owner, path, "readme", 0, "cat readme.txt")
	if !readme.Outcome.OK || !strings.Contains(strings.Join(readme.Outcome.Lines, " "), "maintenance") || readme.Playthrough.Revision != 0 {
		t.Fatal("public readme must expose discoverable instructions without changing progress")
	}
	moved := terminalCommand(owner, path, "navigate", 0, "cd /maintenance")
	if moved.View.Cwd != "/maintenance" || moved.Playthrough.Revision != 1 {
		t.Fatal("navigation must persist Terminal-owned cwd")
	}
	denied := terminalCommand(owner, path, "bad-path", 1, "cd /does-not-exist")
	if denied.Outcome.OK || denied.Playthrough.Revision != 1 || denied.View.Cwd != "/maintenance" {
		t.Fatal("unknown paths must not mutate the workstation")
	}
	owner.reject(path+"/actions", actionBody("forged", 1, "workstation.command", map[string]any{"command": "help", "workstation_ready": true}), 422, "action_rejected")
	resumed := terminalDecode(t, owner.call("GET", path, "", 200))
	if resumed.View.Cwd != "/maintenance" || resumed.Playthrough.Revision != 1 {
		t.Fatal("current view must resume server-owned navigation")
	}
}

func TestTerminalProtectedFlowAndDurableCompletion(t *testing.T) {
	pool := database(t)
	srv := httptest.NewServer(assembly.Handler(pool, "http://localhost:5173", false))
	defer srv.Close()
	owner := guest(t, srv.URL)
	other := guest(t, srv.URL)
	initial := owner.call("POST", "/api/playthroughs", createTerminal, 201)
	fresh := terminalDecode(t, initial)
	path := "/api/playthroughs/" + fresh.Playthrough.ID
	assetPath := path + "/assets/harbour-report"
	if strings.Contains(string(initial), "harbour-report") || strings.Contains(string(initial), "archive_mounted") || strings.Contains(string(initial), "report_read") {
		t.Fatal("initial view must not disclose protected inventory or private schema")
	}
	owner.call("GET", assetPath, "", 404)
	other.call("GET", path, "", 404)
	premature := terminalCommand(owner, path, "early-read", 0, "cat report.txt")
	if premature.Outcome.OK || premature.Playthrough.Revision != 0 {
		t.Fatal("direct premature read must deny without progress")
	}
	outside := terminalCommand(owner, path, "outside-mount", 0, "mount archive")
	if outside.Outcome.OK || outside.View.Ready || outside.Playthrough.Revision != 0 {
		t.Fatal("workstation operations require the maintenance panel")
	}
	terminalCommand(owner, path, "navigate", 0, "cd /maintenance")
	outOfOrder := terminalCommand(owner, path, "early-mount", 1, "mount archive")
	if outOfOrder.Outcome.OK || outOfOrder.View.Ready || outOfOrder.Playthrough.Revision != 1 {
		t.Fatal("archive needs prerequisites before mounting")
	}
	powered := terminalCommand(owner, path, "power", 1, "power on")
	if !powered.Outcome.OK || powered.Playthrough.Revision != 2 || powered.View.Ready {
		t.Fatal("power changes a durable prerequisite without granting archive permission")
	}
	unverified := terminalCommand(owner, path, "unverified", 2, "mount archive")
	if unverified.Outcome.OK || unverified.Playthrough.Revision != 2 {
		t.Fatal("mount must still require link verification")
	}
	verified := terminalCommand(owner, path, "verify", 2, "verify link")
	if !verified.Outcome.OK || verified.Playthrough.Revision != 3 || verified.View.Ready {
		t.Fatal("link verification is a distinct durable prerequisite")
	}
	mounted := terminalCommand(owner, path, "mount", 3, "mount archive")
	if !mounted.Outcome.OK || !mounted.View.Ready || mounted.View.ProtectedFile == nil || mounted.Playthrough.Revision != 4 || mounted.Playthrough.CompletedAt != nil {
		t.Fatal("mount grants protected-file permission without completing")
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
	if res.StatusCode != 200 || res.Header.Get("Cache-Control") != "no-store" || res.Header.Get("Content-Type") != "text/plain; charset=utf-8" || len(content) == 0 {
		t.Fatal("authorized report must use playthrough-scoped no-store delivery")
	}
	if strings.Contains(string(initial), string(content)) {
		t.Fatal("protected content must not be public projection material")
	}
	other.call("GET", assetPath, "", 404)
	locked := terminalDecode(t, owner.call("POST", "/api/playthroughs", createTerminal, 201))
	owner.call("GET", "/api/playthroughs/"+locked.Playthrough.ID+"/assets/harbour-report", "", 404)
	terminalCommand(owner, path, "archive", 4, "cd /archive")
	read := terminalCommand(owner, path, "read", 5, "cat report.txt")
	if !read.Outcome.OK || read.Playthrough.Revision != 6 || read.Playthrough.CompletedAt == nil {
		t.Fatal("reaching report changes case state and first completion in one revision")
	}
	if strings.Contains(strings.Join(read.Outcome.Lines, " "), string(content)) {
		t.Fatal("protected bytes belong to asset delivery rather than transport outcomes")
	}
	again := terminalCommand(owner, path, "read-again", 6, "cat report.txt")
	if again.Playthrough.Revision != 6 || again.Playthrough.CompletedAt == nil || !again.Playthrough.CompletedAt.Equal(*read.Playthrough.CompletedAt) {
		t.Fatal("repeated read preserves first completion without new mutation")
	}
	srv.Close()
	srv = httptest.NewServer(assembly.Handler(pool, "http://localhost:5173", false))
	defer srv.Close()
	owner.url = srv.URL
	resumed := terminalDecode(t, owner.call("GET", path, "", 200))
	if !resumed.View.Ready || resumed.View.Cwd != "/archive" || resumed.Playthrough.Revision != 6 || resumed.Playthrough.CompletedAt == nil {
		t.Fatal("workstation prerequisites, navigation and completion must resume after API restart")
	}
	owner.call("GET", assetPath, "", 200)
}

func TestTerminalReplayStaleAndNoopWithPhoneIsolation(t *testing.T) {
	pool := database(t)
	srv := httptest.NewServer(assembly.Handler(pool, "http://localhost:5173", false))
	defer srv.Close()
	owner := guest(t, srv.URL)
	other := guest(t, srv.URL)
	fresh := terminalDecode(t, owner.call("POST", "/api/playthroughs", createTerminal, 201))
	path := "/api/playthroughs/" + fresh.Playthrough.ID
	actions := path + "/actions"
	phone := decode(t, owner.call("POST", "/api/playthroughs", createPhone, 201))
	phonePath := "/api/playthroughs/" + phone.Playthrough.ID
	body := actionBody("navigate", 0, "workstation.command", map[string]string{"command": "cd /maintenance"})
	first := owner.call("POST", actions, body, 200)
	replay := owner.call("POST", actions, body, 200)
	if string(first) != string(replay) {
		t.Fatal("Terminal committed retry must return exact recorded safe result")
	}
	owner.reject(actions, actionBody("navigate", 0, "workstation.command", map[string]string{"command": "cd /"}), 409, "request_id_conflict")
	owner.reject(actions, actionBody("stale", 0, "workstation.command", map[string]string{"command": "power on"}), 409, "revision_conflict")
	terminalCommand(owner, path, "power", 1, "power on")
	noop := terminalCommand(owner, path, "power-noop", 2, "power on")
	if noop.Playthrough.Revision != 2 {
		t.Fatal("repeated prerequisite does not advance revision")
	}
	old := terminalDecode(t, owner.call("POST", actions, body, 200))
	current := terminalDecode(t, owner.call("GET", path, "", 200))
	if old.Playthrough.Revision != 1 || current.Playthrough.Revision != 2 {
		t.Fatal("older replay retains committed snapshot without reverting current progress")
	}
	other.reject(actions, actionBody("other", 2, "workstation.command", map[string]string{"command": "verify link"}), 404, "not_found")
	owner.reject(actions, actionBody("wrong-case-tag", 2, "gallery.open", map[string]any{}), 422, "action_rejected")
	unchangedPhone := decode(t, owner.call("GET", phonePath, "", 200))
	if unchangedPhone.View.Unlocked || unchangedPhone.Playthrough.Revision != 0 || unchangedPhone.Playthrough.Completed != nil {
		t.Fatal("Terminal transitions cannot mutate Phone run")
	}
	restart := terminalDecode(t, owner.call("POST", "/api/playthroughs", createTerminal, 201))
	if restart.Playthrough.ID == fresh.Playthrough.ID || restart.View.Ready || restart.View.Cwd != "/" {
		t.Fatal("Terminal restart creates a distinct fresh run")
	}
	retained := terminalDecode(t, owner.call("GET", path, "", 200))
	if retained.Playthrough.Revision != 2 || retained.View.Cwd != "/maintenance" {
		t.Fatal("restart must retain previous Terminal progress")
	}
	srv.Close()
	srv = httptest.NewServer(assembly.Handler(pool, "http://localhost:5173", false))
	defer srv.Close()
	owner.url = srv.URL
	if string(owner.call("POST", actions, body, 200)) != string(first) {
		t.Fatal("Terminal receipt survives API-instance restart")
	}
}

func TestTerminalCommandValidationAndSimulatedOnly(t *testing.T) {
	pool := database(t)
	srv := httptest.NewServer(assembly.Handler(pool, "http://localhost:5173", false))
	defer srv.Close()
	owner := guest(t, srv.URL)
	fresh := terminalDecode(t, owner.call("POST", "/api/playthroughs", createTerminal, 201))
	path := "/api/playthroughs/" + fresh.Playthrough.ID
	actions := path + "/actions"
	for _, payload := range []any{map[string]any{}, map[string]any{"command": nil}, map[string]any{"command": 5}, map[string]any{"command": "help", "state": map[string]bool{"archive_mounted": true}}, map[string]any{"command": strings.Repeat("x", 201)}} {
		owner.reject(actions, actionBody("invalid", 0, "workstation.command", payload), 422, "action_rejected")
	}
	for i, command := range []string{"", "cat /archive/report.txt", "cat ../../report.txt", "help; whoami", "curl https://example.test", "power on", "verify link", "mount archive", "cd /archive"} {
		result := terminalCommand(owner, path, fmt.Sprintf("denied-%d", i), 0, command)
		if result.Outcome.OK || result.Playthrough.Revision != 0 || result.View.Ready || result.Playthrough.CompletedAt != nil {
			t.Fatal("unrecognized or premature simulated input must not execute or grant progression")
		}
	}
	// A malformed attempt does not permanently reserve the request ID.
	help := terminalCommand(owner, path, "invalid", 0, "help")
	if !help.Outcome.OK || help.Playthrough.Revision != 0 {
		t.Fatal("corrected command must remain usable")
	}
}
