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
	"github.com/rbrick/mission-control/rig/config"
	"github.com/spf13/cobra"
)

func newConnectCmd() *cobra.Command {
	var configPath string
	var gatewayURL string
	var token string
	var keepAliveInterval time.Duration

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect this rig to a Mission Control gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return runConnect(ctx, configPath, gatewayURL, token, keepAliveInterval)
		},
	}

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

	log.Printf("connected rig %s to gateway %s", cfg.ID, u.String())

	if err := client.WriteJSON(registerPacket(cfg)); err != nil {
		return fmt.Errorf("send register: %w", err)
	}

	errCh := make(chan error, 2)
	go func() { errCh <- readCommands(ctx, client) }()
	go func() { errCh <- writeKeepAlives(ctx, client, keepAliveInterval) }()

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

func registerPacket(cfg *config.Config) protocol.Packet {
	adapterID := cfg.Adapter.Type
	if adapterID == "" {
		adapterID = "sim"
	}
	return protocol.Packet{
		V:       protocol.Version,
		Action:  "register",
		ID:      cfg.ID,
		TS:      time.Now().UTC(),
		Adapter: cfg.Adapter.Type,
		Capabilities: []protocol.Capability{
			{AdapterID: adapterID, Adapter: cfg.Adapter.Type, Namespace: "rig", Commands: []string{"get_status"}},
			{AdapterID: adapterID, Adapter: cfg.Adapter.Type, Namespace: "mount", Commands: []string{"goto_radec", "park"}},
			{AdapterID: adapterID, Adapter: cfg.Adapter.Type, Namespace: "camera", Commands: []string{"capture"}},
			{AdapterID: adapterID, Adapter: cfg.Adapter.Type, Namespace: "sequence", Commands: []string{"start", "stop"}},
		},
	}
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

func writeKeepAlives(ctx context.Context, client *gatewayClient, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			state, _ := json.Marshal(map[string]interface{}{"connected": true, "safety": "safe", "active": []interface{}{}})
			pkt := protocol.Packet{V: protocol.Version, Action: "keep_alive", TS: time.Now().UTC(), IntervalMS: int(interval / time.Millisecond), State: state}
			if err := client.WriteJSON(pkt); err != nil {
				return err
			}
		}
	}
}

func readCommands(ctx context.Context, client *gatewayClient) error {
	for {
		var pkt protocol.Packet
		if err := client.conn.ReadJSON(&pkt); err != nil {
			return err
		}
		if pkt.Action != "send" || pkt.Phase != "command" {
			continue
		}
		go handleCommand(client, pkt)
	}
}

func handleCommand(client *gatewayClient, cmd protocol.Packet) {
	log.Printf("command %s: %s.%s", cmd.ID, cmd.Namespace, cmd.Command)
	writeSend(client, cmd, "progress", map[string]interface{}{"state": "running", "progress": 0.25}, nil)
	time.Sleep(500 * time.Millisecond)

	switch cmd.Namespace + "." + cmd.Command {
	case "rig.get_status":
		writeSend(client, cmd, "result", map[string]interface{}{"connected": true, "safety": "safe", "mode": "simulated"}, nil)
	case "mount.goto_radec":
		writeSend(client, cmd, "progress", map[string]interface{}{"state": "slewing", "progress": 0.75}, nil)
		time.Sleep(500 * time.Millisecond)
		writeSend(client, cmd, "result", map[string]interface{}{"arrived": true}, nil)
	case "mount.park":
		writeSend(client, cmd, "result", map[string]interface{}{"parked": true}, nil)
	case "camera.capture":
		writeSend(client, cmd, "result", map[string]interface{}{"image_id": "simulated-image", "saved": true}, nil)
	case "sequence.start":
		writeSend(client, cmd, "result", map[string]interface{}{"started": true}, nil)
	case "sequence.stop":
		writeSend(client, cmd, "result", map[string]interface{}{"stopped": true}, nil)
	default:
		writeSend(client, cmd, "error", nil, &protocol.Error{Code: "NOT_SUPPORTED", Message: "command not supported by simulated rig"})
	}
}

func writeSend(client *gatewayClient, cmd protocol.Packet, phase string, data map[string]interface{}, packetErr *protocol.Error) {
	raw, _ := json.Marshal(data)
	pkt := protocol.Packet{V: protocol.Version, Action: "send", ID: cmd.ID, TS: time.Now().UTC(), Namespace: cmd.Namespace, Command: cmd.Command, Phase: phase, Data: raw, Error: packetErr}
	if err := client.WriteJSON(pkt); err != nil {
		log.Printf("write %s failed: %v", phase, err)
	}
}
