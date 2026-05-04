package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rbrick/mission-control/protocol"
	"github.com/rbrick/mission-control/rig/adapter"
	"github.com/rbrick/mission-control/rig/config"
	"github.com/spf13/cobra"
)

func newConnectCmd() *cobra.Command {
	var configPath string
	var gatewayURL string
	var token string
	var keepAliveInterval time.Duration
	cmd := &cobra.Command{Use: "connect", Short: "Connect this rig to a Mission Control gateway", RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		return runConnect(ctx, configPath, gatewayURL, token, keepAliveInterval)
	}}
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to a rig config YAML file")
	cmd.Flags().StringVar(&gatewayURL, "gateway", defaultGatewayURL(), "Gateway rig WebSocket URL")
	cmd.Flags().StringVar(&token, "token", os.Getenv("GATEWAY_TOKEN"), "Gateway bearer token")
	cmd.Flags().DurationVar(&keepAliveInterval, "keep-alive", 5*time.Second, "Keep-alive interval")
	return cmd
}

func defaultGatewayURL() string {
	if value := os.Getenv("GATEWAY_URL"); value != "" {
		return value
	}
	return "ws://localhost:8080/v1/ws/rig"
}

func runConnect(ctx context.Context, configPath, gatewayURL, token string, keepAliveInterval time.Duration) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	adapters, err := adapter.FromConfig(cfg)
	if err != nil {
		return err
	}
	rig := newRigRuntime(adapters)
	u, err := url.Parse(gatewayURL)
	if err != nil {
		return fmt.Errorf("parse gateway url: %w", err)
	}
	q := u.Query()
	if q.Get("id") == "" {
		q.Set("id", cfg.ID)
	}
	u.RawQuery = q.Encode()
	headers := http.Header{}
	if token != "" {
		headers.Set("Authorization", "Bearer "+token)
	}
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), headers)
	if err != nil {
		return fmt.Errorf("connect gateway: %w", err)
	}
	defer conn.Close()
	client := &gatewayClient{conn: conn}
	log.Printf("connected rig %s to gateway %s with %d adapter(s)", cfg.ID, u.String(), len(adapters))
	if err := client.WriteJSON(registerPacket(cfg, adapters)); err != nil {
		return fmt.Errorf("send register: %w", err)
	}
	errCh := make(chan error, 2)
	go func() { errCh <- readCommands(ctx, client, rig) }()
	go func() { errCh <- writeKeepAlives(ctx, client, rig, keepAliveInterval) }()
	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		if errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}
}

func registerPacket(cfg *config.Config, adapters []adapter.Adapter) protocol.Packet {
	caps := []protocol.Capability{}
	adapterType := ""
	for _, a := range adapters {
		if adapterType == "" {
			adapterType = a.Type()
		}
		caps = append(caps, a.Capabilities()...)
	}
	return protocol.Packet{V: protocol.Version, Action: "register", ID: cfg.ID, TS: time.Now().UTC(), Adapter: adapterType, Capabilities: caps}
}

type rigRuntime struct {
	adapters map[string]adapter.Adapter
	fallback adapter.Adapter
}

func newRigRuntime(adapters []adapter.Adapter) *rigRuntime {
	r := &rigRuntime{adapters: map[string]adapter.Adapter{}}
	for _, a := range adapters {
		if r.fallback == nil {
			r.fallback = a
		}
		r.adapters[a.ID()] = a
	}
	return r
}
func (r *rigRuntime) adapterFor(target *protocol.Target) adapter.Adapter {
	if target != nil && target.AdapterID != "" {
		if a := r.adapters[target.AdapterID]; a != nil {
			return a
		}
	}
	return r.fallback
}
func (r *rigRuntime) status(ctx context.Context) map[string]interface{} {
	out := map[string]interface{}{"connected": true, "safety": "safe", "active": []interface{}{}}
	adapters := map[string]interface{}{}
	for id, a := range r.adapters {
		status, err := a.Status(ctx)
		if err != nil {
			adapters[id] = map[string]interface{}{"error": err.Error()}
		} else {
			adapters[id] = status
		}
	}
	out["adapters"] = adapters
	return out
}

type gatewayClient struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (c *gatewayClient) WriteJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(v)
}

func writeKeepAlives(ctx context.Context, client *gatewayClient, rig *rigRuntime, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			state, _ := json.Marshal(rig.status(ctx))
			pkt := protocol.Packet{V: protocol.Version, Action: "keep_alive", TS: time.Now().UTC(), IntervalMS: int(interval / time.Millisecond), State: state}
			if err := client.WriteJSON(pkt); err != nil {
				return err
			}
		}
	}
}

func readCommands(ctx context.Context, client *gatewayClient, rig *rigRuntime) error {
	for {
		var pkt protocol.Packet
		if err := client.conn.ReadJSON(&pkt); err != nil {
			return err
		}
		if pkt.Action == "send" && pkt.Phase == "command" {
			go handleCommand(ctx, client, rig, pkt)
		}
	}
}

func handleCommand(ctx context.Context, client *gatewayClient, rig *rigRuntime, cmd protocol.Packet) {
	log.Printf("command %s: %s.%s", cmd.ID, cmd.Namespace, cmd.Command)
	a := rig.adapterFor(cmd.Target)
	if a == nil {
		writeSend(client, cmd, "error", nil, adapter.ProtocolError("UNAVAILABLE", "no adapter configured"))
		return
	}
	progress := func(phase string, data map[string]interface{}, packetErr *protocol.Error) error {
		writeSend(client, cmd, phase, data, packetErr)
		return nil
	}
	result, packetErr := a.Handle(ctx, cmd.Namespace, cmd.Command, cmd.Data, progress)
	if packetErr != nil {
		writeSend(client, cmd, "error", nil, packetErr)
		return
	}
	writeSend(client, cmd, "result", result, nil)
}

func writeSend(client *gatewayClient, cmd protocol.Packet, phase string, data map[string]interface{}, packetErr *protocol.Error) {
	raw, _ := json.Marshal(data)
	pkt := protocol.Packet{V: protocol.Version, Action: "send", ID: cmd.ID, TS: time.Now().UTC(), Namespace: cmd.Namespace, Command: cmd.Command, Phase: phase, Data: raw, Error: packetErr}
	if err := client.WriteJSON(pkt); err != nil {
		log.Printf("write %s failed: %v", phase, err)
	}
}
