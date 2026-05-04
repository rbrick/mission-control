package adapter

import (
	"context"
	"encoding/json"
)

type Command struct {
	Namespace   string
	Name        string
	Description string
	Handler     Handler
}

type Handler func(ctx context.Context, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, error)

type Registry struct {
	commands map[string]Command
}

func NewRegistry(commands ...Command) *Registry {
	r := &Registry{commands: map[string]Command{}}
	for _, command := range commands {
		r.Register(command)
	}
	return r
}

func (r *Registry) Register(command Command) {
	r.commands[key(command.Namespace, command.Name)] = command
}

func (r *Registry) Handle(ctx context.Context, namespace, command string, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, *Error) {
	registered, ok := r.commands[key(namespace, command)]
	if !ok {
		return nil, ProtocolError("NOT_SUPPORTED", "command not supported")
	}
	result, err := registered.Handler(ctx, data, progress)
	if err != nil {
		if protocolErr, ok := err.(*CommandError); ok {
			return nil, protocolErr.Packet
		}
		return nil, ProtocolError("INTERNAL", err.Error())
	}
	return result, nil
}

func (r *Registry) Capabilities(adapterID, adapterType string) []Capability {
	byNamespace := map[string][]string{}
	for _, command := range r.commands {
		byNamespace[command.Namespace] = append(byNamespace[command.Namespace], command.Name)
	}
	capabilities := make([]Capability, 0, len(byNamespace))
	for namespace, commands := range byNamespace {
		capabilities = append(capabilities, Capability{AdapterID: adapterID, Adapter: adapterType, Namespace: namespace, Commands: commands})
	}
	return capabilities
}

func key(namespace, command string) string { return namespace + "." + command }

type CommandError struct{ Packet *Error }

func (e *CommandError) Error() string { return e.Packet.Message }

func Fail(code, message string) error { return &CommandError{Packet: ProtocolError(code, message)} }
