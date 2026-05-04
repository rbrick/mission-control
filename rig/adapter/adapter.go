package adapter

import (
	"context"
	"encoding/json"

	"github.com/rbrick/mission-control/protocol"
)

type Error = protocol.Error
type Capability = protocol.Capability

type ProgressFunc func(phase string, data map[string]interface{}, packetErr *protocol.Error) error

type Adapter interface {
	ID() string
	Type() string
	Capabilities() []protocol.Capability
	Status(ctx context.Context) (map[string]interface{}, error)
	Handle(ctx context.Context, namespace, command string, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, *protocol.Error)
}

func ProtocolError(code, message string) *protocol.Error {
	return &protocol.Error{Code: code, Message: message}
}
