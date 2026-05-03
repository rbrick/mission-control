package protocol

import (
	"encoding/json"
	"time"
)

const Version = "mc.v1"

type Capability struct {
	AdapterID string   `json:"adapter_id,omitempty"`
	Adapter   string   `json:"adapter,omitempty"`
	Namespace string   `json:"namespace"`
	Commands  []string `json:"commands"`
}

type Error struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type Target struct {
	AdapterID string `json:"adapter_id,omitempty"`
}

type Packet struct {
	V            string          `json:"v"`
	Action       string          `json:"action"`
	ID           string          `json:"id,omitempty"`
	TS           time.Time       `json:"ts,omitempty"`
	Adapter      string          `json:"adapter,omitempty"`
	Capabilities []Capability    `json:"capabilities,omitempty"`
	IntervalMS   int             `json:"interval_ms,omitempty"`
	State        json.RawMessage `json:"state,omitempty"`
	Namespace    string          `json:"namespace,omitempty"`
	Command      string          `json:"command,omitempty"`
	Phase        string          `json:"phase,omitempty"`
	Target       *Target         `json:"target,omitempty"`
	Data         json.RawMessage `json:"data,omitempty"`
	Error        *Error          `json:"error,omitempty"`
}
