package terminal

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/alireza-constantin/case-platform/server/internal/kernel"
)

type Module struct{}
type state struct {
	Cwd            string `json:"cwd"`
	PowerOn        bool   `json:"power_on"`
	LinkVerified   bool   `json:"link_verified"`
	ArchiveMounted bool   `json:"archive_mounted"`
	ReportRead     bool   `json:"report_read"`
}
type entry struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}
type view struct {
	Cwd           string         `json:"cwd"`
	Entries       []entry        `json:"entries"`
	Ready         bool           `json:"workstation_ready"`
	ProtectedFile *fileReference `json:"protected_file,omitempty"`
}
type fileReference struct {
	ID   string `json:"asset_id"`
	Name string `json:"name"`
}
type outcome struct {
	Lines []string `json:"lines"`
	OK    bool     `json:"ok"`
}

func (Module) Metadata() kernel.CaseMetadata {
	return kernel.CaseMetadata{CaseID: "terminal-demo", CaseVersion: "m0-v1", Title: "The Harbour Workstation"}
}
func (Module) Initialize() (json.RawMessage, error) { return json.Marshal(state{Cwd: "/"}) }
func decodeObject(data json.RawMessage, value any) bool {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return len(data) > 0 && data[0] == '{' && decoder.Decode(value) == nil && decoder.Decode(new(any)) == io.EOF
}
func readState(data json.RawMessage) (state, error) {
	var saved struct {
		Cwd            *string `json:"cwd"`
		PowerOn        *bool   `json:"power_on"`
		LinkVerified   *bool   `json:"link_verified"`
		ArchiveMounted *bool   `json:"archive_mounted"`
		ReportRead     *bool   `json:"report_read"`
	}
	if !decodeObject(data, &saved) || saved.Cwd == nil || saved.PowerOn == nil || saved.LinkVerified == nil || saved.ArchiveMounted == nil || saved.ReportRead == nil {
		return state{}, errors.New("invalid Terminal state")
	}
	s := state{*saved.Cwd, *saved.PowerOn, *saved.LinkVerified, *saved.ArchiveMounted, *saved.ReportRead}
	if (s.Cwd != "/" && s.Cwd != "/maintenance" && s.Cwd != "/archive") || (s.LinkVerified && !s.PowerOn) || (s.ArchiveMounted && !s.LinkVerified) || (s.ReportRead && !s.ArchiveMounted) || (s.Cwd == "/archive" && !s.ArchiveMounted) {
		return state{}, errors.New("invalid Terminal state")
	}
	return s, nil
}
func (Module) Project(data json.RawMessage) (json.RawMessage, error) {
	s, err := readState(data)
	if err != nil {
		return nil, err
	}
	v := view{Cwd: s.Cwd, Ready: s.ArchiveMounted, Entries: visibleEntries(s)}
	if s.ArchiveMounted {
		v.ProtectedFile = &fileReference{"harbour-report", "report.txt"}
	}
	return json.Marshal(v)
}

func visibleEntries(s state) []entry {
	entries := []entry{}
	switch s.Cwd {
	case "/":
		entries = []entry{{"readme.txt", "file"}, {"maintenance", "directory"}}
	case "/maintenance":
		entries = []entry{{"panel.txt", "file"}}
	case "/archive":
		entries = []entry{{"report.txt", "file"}}
	}
	if s.ArchiveMounted && s.Cwd == "/" {
		entries = append(entries, entry{"archive", "directory"})
	}
	return entries
}

var publicInstructions = []string{"This is a simulated harbour workstation. No operating-system command is executed.", "Navigate to /maintenance and read panel.txt to restore the archive."}

func (Module) Act(data json.RawMessage, tag string, payload json.RawMessage) (kernel.Transition, error) {
	s, err := readState(data)
	if err != nil {
		return kernel.Transition{}, err
	}
	var input struct {
		Command *string `json:"command"`
	}
	if tag != "workstation.command" || !decodeObject(payload, &input) || input.Command == nil || len(*input.Command) > 200 {
		return kernel.Transition{}, kernel.ErrActionRejected
	}
	command := strings.TrimSpace(*input.Command)
	result := outcome{Lines: []string{}, OK: true}
	transition := kernel.Transition{}
	switch command {
	case "help":
		result.Lines = []string{"Simulated commands: help, pwd, ls, cd /, cd /maintenance, cat readme.txt, cat panel.txt.", "Read the public instructions to discover workstation operations."}
	case "pwd":
		result.Lines = []string{s.Cwd}
	case "ls":
		for _, item := range visibleEntries(s) {
			result.Lines = append(result.Lines, item.Name)
		}
	case "cat readme.txt":
		if s.Cwd != "/" {
			result = outcome{[]string{"File not found in this directory."}, false}
		} else {
			result.Lines = publicInstructions
		}
	case "cat panel.txt":
		if s.Cwd != "/maintenance" {
			result = outcome{[]string{"File not found in this directory."}, false}
		} else {
			result.Lines = []string{"Restore sequence: power on, then verify link, then mount archive.", "After mounting, cd /archive and cat report.txt to reach the authorized report."}
		}
	case "cd /":
		s.Cwd = "/"
		result.Lines = []string{s.Cwd}
	case "cd /maintenance", "cd maintenance":
		if command == "cd maintenance" && s.Cwd != "/" {
			result = outcome{[]string{"Directory not found."}, false}
		} else {
			s.Cwd = "/maintenance"
			result.Lines = []string{s.Cwd}
		}
	case "power on":
		if s.Cwd != "/maintenance" {
			result = outcome{[]string{"Use the maintenance panel first."}, false}
		} else {
			if !s.PowerOn {
				s.PowerOn = true
				transition.Events = []json.RawMessage{json.RawMessage(`{"type":"workstation.powered"}`)}
			}
			result.Lines = []string{"Power is on. The communication link can now be verified."}
		}
	case "verify link":
		if s.Cwd != "/maintenance" || !s.PowerOn {
			result = outcome{[]string{"Verification denied: restore panel power first."}, false}
		} else {
			if !s.LinkVerified {
				s.LinkVerified = true
				transition.Events = []json.RawMessage{json.RawMessage(`{"type":"workstation.link_verified"}`)}
			}
			result.Lines = []string{"Link verified. The archive can now be mounted."}
		}
	case "mount archive":
		if s.Cwd != "/maintenance" || !s.PowerOn || !s.LinkVerified {
			result = outcome{[]string{"Mount denied: panel power and link verification are required."}, false}
		} else {
			if !s.ArchiveMounted {
				s.ArchiveMounted = true
				transition.Events = []json.RawMessage{json.RawMessage(`{"type":"workstation.archive_mounted"}`)}
			}
			result.Lines = []string{"Archive mounted. Navigate to /archive and read report.txt."}
		}
	case "cd /archive", "cd archive":
		if !s.ArchiveMounted {
			result = outcome{[]string{"Access denied: archive unavailable."}, false}
		} else if command == "cd archive" && s.Cwd != "/" {
			result = outcome{[]string{"Directory not found."}, false}
		} else {
			s.Cwd = "/archive"
			result.Lines = []string{s.Cwd}
		}
	case "cat report.txt":
		if s.Cwd != "/archive" || !s.ArchiveMounted {
			result = outcome{[]string{"Access denied: reach the mounted archive first."}, false}
		} else {
			if !s.ReportRead {
				s.ReportRead = true
				transition.Events = []json.RawMessage{json.RawMessage(`{"type":"workstation.report_read"}`)}
			}
			transition.Complete = true
			result.Lines = []string{"Authorized report reached. Open the protected file to read its contents."}
		}
	default:
		result = outcome{[]string{"Unknown simulated command. Use help."}, false}
	}
	transition.State, err = json.Marshal(s)
	if err != nil {
		return transition, err
	}
	transition.Outcome, err = json.Marshal(result)
	return transition, err
}
func (Module) Asset(data json.RawMessage, id string) (kernel.Asset, bool, error) {
	s, err := readState(data)
	if err != nil {
		return kernel.Asset{}, false, err
	}
	if !s.ArchiveMounted || id != "harbour-report" {
		return kernel.Asset{}, false, nil
	}
	return kernel.Asset{ContentType: "text/plain; charset=utf-8", Bytes: []byte("HARBOUR ARCHIVE — RESTRICTED REPORT\nThe missing survey vessel returned to the eastern berth at 05:40.\nIts recorder contains the evidence needed to close the harbour investigation.\n")}, true, nil
}
