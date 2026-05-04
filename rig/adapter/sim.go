package adapter

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rbrick/mission-control/protocol"
)

type SimAdapter struct {
	id       string
	registry *Registry
}

func NewSim(id string) *SimAdapter {
	if id == "" {
		id = "sim"
	}
	a := &SimAdapter{id: id}
	a.registry = NewRegistry(
		Command{Namespace: "rig", Name: "get_status", Handler: a.getStatus},
		Command{Namespace: "mount", Name: "goto_radec", Handler: a.gotoRADec},
		Command{Namespace: "mount", Name: "park", Handler: a.park},
		Command{Namespace: "mount", Name: "unpark", Handler: a.unpark},
		Command{Namespace: "mount", Name: "abort", Handler: a.abort},
		Command{Namespace: "camera", Name: "capture", Handler: a.capture},
		Command{Namespace: "sequence", Name: "start", Handler: a.startSequence},
		Command{Namespace: "sequence", Name: "stop", Handler: a.stopSequence},
	)
	return a
}

func (a *SimAdapter) ID() string   { return a.id }
func (a *SimAdapter) Type() string { return "sim" }
func (a *SimAdapter) Capabilities() []protocol.Capability {
	return a.registry.Capabilities(a.id, a.Type())
}
func (a *SimAdapter) Status(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{"connected": true, "safety": "safe", "mode": "simulated"}, nil
}
func (a *SimAdapter) Handle(ctx context.Context, namespace, command string, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, *protocol.Error) {
	return a.registry.Handle(ctx, namespace, command, data, progress)
}

func (a *SimAdapter) getStatus(ctx context.Context, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, error) {
	return a.Status(ctx)
}

func (a *SimAdapter) gotoRADec(ctx context.Context, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, error) {
	_ = progress("progress", map[string]interface{}{"state": "slewing", "progress": 0.25}, nil)
	time.Sleep(500 * time.Millisecond)
	_ = progress("progress", map[string]interface{}{"state": "slewing", "progress": 0.75}, nil)
	time.Sleep(500 * time.Millisecond)
	return map[string]interface{}{"arrived": true}, nil
}

func (a *SimAdapter) park(ctx context.Context, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, error) {
	return map[string]interface{}{"parked": true}, nil
}

func (a *SimAdapter) unpark(ctx context.Context, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, error) {
	return map[string]interface{}{"unparked": true}, nil
}

func (a *SimAdapter) abort(ctx context.Context, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, error) {
	return map[string]interface{}{"aborted": true}, nil
}

func (a *SimAdapter) capture(ctx context.Context, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, error) {
	return map[string]interface{}{"image_id": "simulated-image", "saved": true}, nil
}

func (a *SimAdapter) startSequence(ctx context.Context, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, error) {
	return map[string]interface{}{"started": true}, nil
}

func (a *SimAdapter) stopSequence(ctx context.Context, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, error) {
	return map[string]interface{}{"stopped": true}, nil
}
