package phone

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/alireza-constantin/case-platform/server/internal/kernel"
)

type Module struct{}
type state struct {
	Attempts int  `json:"attempts"`
	Unlocked bool `json:"unlocked"`
	Opened   bool `json:"opened"`
}

func (Module) Metadata() kernel.CaseMetadata {
	return kernel.CaseMetadata{CaseID: "phone-demo", CaseVersion: "m0-v1", Title: "The Last Photograph"}
}
func (Module) Initialize() (json.RawMessage, error) {
	return json.RawMessage(`{"attempts":0,"unlocked":false,"opened":false}`), nil
}
func readState(data json.RawMessage) (state, error) {
	var stored struct {
		Attempts *int  `json:"attempts"`
		Unlocked *bool `json:"unlocked"`
		Opened   *bool `json:"opened"`
	}
	if !payloadObject(data, &stored) || stored.Attempts == nil || stored.Unlocked == nil || stored.Opened == nil || *stored.Attempts < 0 || (*stored.Opened && !*stored.Unlocked) {
		return state{}, errors.New("invalid Phone state")
	}
	return state{*stored.Attempts, *stored.Unlocked, *stored.Opened}, nil
}
func (Module) Project(data json.RawMessage) (json.RawMessage, error) {
	s, err := readState(data)
	if err != nil {
		return nil, err
	}
	view := struct {
		Messages []message        `json:"messages"`
		Unlocked bool             `json:"gallery_unlocked"`
		Assets   []assetReference `json:"gallery_assets,omitempty"`
	}{
		Messages: []message{{"arrival", "Mara", "Meet me where the shoreline light guides the fishing boats. The gallery password is that building's single lowercase name."}, {"reminder", "Jon", "The photos from the coast are in the locked gallery. Please look at them when you arrive."}}, Unlocked: s.Unlocked}
	if s.Unlocked {
		view.Assets = []assetReference{{"coastal-photo", "The last photograph"}}
	}
	return json.Marshal(view)
}

type message struct {
	ID     string `json:"id"`
	Sender string `json:"sender"`
	Text   string `json:"text"`
}
type assetReference struct {
	ID    string `json:"asset_id"`
	Label string `json:"label"`
}

func payloadObject(payload json.RawMessage, target any) bool {
	if len(payload) == 0 || payload[0] != '{' {
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if decoder.Decode(target) != nil {
		return false
	}
	return decoder.Decode(new(any)) == io.EOF
}
func (Module) Act(data json.RawMessage, tag string, payload json.RawMessage) (kernel.Transition, error) {
	s, err := readState(data)
	if err != nil {
		return kernel.Transition{}, err
	}
	transition := kernel.Transition{}
	switch tag {
	case "gallery.attempt":
		var input struct {
			Password *string `json:"password"`
		}
		if !payloadObject(payload, &input) || input.Password == nil || len(*input.Password) > 200 {
			return transition, kernel.ErrActionRejected
		}
		accepted := s.Unlocked || *input.Password == "lighthouse"
		if !s.Unlocked {
			s.Attempts++
			s.Unlocked = accepted
			transition.Events = []json.RawMessage{json.RawMessage(`{"type":"gallery.attempted"}`)}
		}
		transition.Outcome, _ = json.Marshal(struct {
			Accepted bool `json:"accepted"`
		}{accepted})
	case "gallery.open":
		if !payloadObject(payload, &struct{}{}) || !s.Unlocked {
			return transition, kernel.ErrActionRejected
		}
		if !s.Opened {
			s.Opened = true
			transition.Events = []json.RawMessage{json.RawMessage(`{"type":"gallery.opened"}`)}
		}
		transition.Complete = true
		transition.Outcome = json.RawMessage(`{"opened":true}`)
	default:
		return transition, kernel.ErrActionRejected
	}
	transition.State, err = json.Marshal(s)
	return transition, err
}

func (Module) Asset(data json.RawMessage, id string) (kernel.Asset, bool, error) {
	s, err := readState(data)
	if err != nil {
		return kernel.Asset{}, false, err
	}
	if !s.Unlocked || id != "coastal-photo" {
		return kernel.Asset{}, false, nil
	}
	// The unrevealed photograph is server-owned; it is never a public/static fixture.
	return kernel.Asset{ContentType: "image/svg+xml", Bytes: []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="960" height="640" viewBox="0 0 960 640"><rect width="960" height="640" fill="#182d3c"/><rect y="330" width="960" height="310" fill="#325c73"/><path d="M150 400h620l-110 125H260z" fill="#87adc1"/><rect x="320" y="290" width="300" height="100" fill="#dae2df"/><path d="M540 90v200M540 90l150 160H540" stroke="#d4be7a" fill="#527d92" stroke-width="8"/><text x="60" y="580" fill="#fff" font-family="sans-serif" font-size="27">Mara left aboard the blue ferry at dawn.</text><text x="60" y="620" fill="#a5c1d0" font-family="sans-serif" font-size="20">The last photograph — north pier, 06:10</text></svg>`)}, true, nil
}
