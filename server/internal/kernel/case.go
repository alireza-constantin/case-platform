package kernel

import (
	"encoding/json"
	"errors"
)

var ErrActionRejected = errors.New("case rejected action")

type Transition struct {
	State    json.RawMessage
	Outcome  json.RawMessage
	Events   []json.RawMessage
	Complete bool
}
type Asset struct {
	ContentType string
	Bytes       []byte
}

// Case is trusted synchronous case logic. It receives no persistence or network capabilities.
type Case interface {
	Metadata() CaseMetadata
	Initialize() (json.RawMessage, error)
	Project(json.RawMessage) (json.RawMessage, error)
	Act(json.RawMessage, string, json.RawMessage) (Transition, error)
	Asset(json.RawMessage, string) (Asset, bool, error)
}
