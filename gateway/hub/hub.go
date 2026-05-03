package hub

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rbrick/mission-control/protocol"
)

type RigSnapshot struct {
	ID           string                  `json:"id"`
	Online       bool                    `json:"online"`
	Adapter      string                  `json:"adapter,omitempty"`
	Capabilities []protocol.Capability   `json:"capabilities,omitempty"`
	State        json.RawMessage         `json:"state,omitempty"`
	LastSeen     time.Time               `json:"last_seen,omitempty"`
	ConnectedAt  time.Time               `json:"connected_at,omitempty"`
	Active       map[string]CommandState `json:"active,omitempty"`
}

type CommandState struct {
	ID        string          `json:"id"`
	RigID     string          `json:"rig_id"`
	Namespace string          `json:"namespace"`
	Command   string          `json:"command"`
	Phase     string          `json:"phase"`
	Data      json.RawMessage `json:"data,omitempty"`
	Error     *protocol.Error `json:"error,omitempty"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type RigHub struct {
	mu       sync.RWMutex
	upgrader websocket.Upgrader
	token    string
	rigs     map[string]*connectedRig
	commands map[string]CommandState
}

type connectedRig struct {
	id       string
	conn     *websocket.Conn
	send     chan protocol.Packet
	snapshot RigSnapshot
}

func New(options ...Option) *RigHub {
	h := &RigHub{
		upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		rigs:     map[string]*connectedRig{},
		commands: map[string]CommandState{},
	}
	for _, option := range options {
		option(h)
	}
	return h
}

type Option func(*RigHub)

func WithToken(token string) Option {
	return func(h *RigHub) {
		h.token = token
	}
}

func (h *RigHub) ServeWS(w http.ResponseWriter, r *http.Request) error {
	if !h.authorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return errors.New("unauthorized rig websocket")
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	_, raw, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return err
	}

	var pkt protocol.Packet
	if err := json.Unmarshal(raw, &pkt); err != nil {
		conn.Close()
		return err
	}
	if pkt.V != protocol.Version || pkt.Action != "register" {
		conn.Close()
		return errors.New("first packet must be mc.v1 register")
	}

	rigID := r.URL.Query().Get("id")
	if rigID == "" {
		rigID = pkt.ID
	}
	if rigID == "" {
		rigID = randomID("rig_")
	}

	cr := &connectedRig{id: rigID, conn: conn, send: make(chan protocol.Packet, 32)}
	cr.snapshot = RigSnapshot{ID: rigID, Online: true, Adapter: pkt.Adapter, Capabilities: pkt.Capabilities, LastSeen: time.Now().UTC(), ConnectedAt: time.Now().UTC(), Active: map[string]CommandState{}}

	h.mu.Lock()
	if old := h.rigs[rigID]; old != nil {
		old.conn.Close()
		close(old.send)
	}
	h.rigs[rigID] = cr
	h.mu.Unlock()

	go h.writeLoop(cr)
	h.readLoop(cr)
	return nil
}

func (h *RigHub) ListRigs() []RigSnapshot {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]RigSnapshot, 0, len(h.rigs))
	for _, r := range h.rigs {
		out = append(out, r.snapshot)
	}
	return out
}

func (h *RigHub) GetRig(id string) (RigSnapshot, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	r := h.rigs[id]
	if r == nil {
		return RigSnapshot{}, false
	}
	return r.snapshot, true
}

func (h *RigHub) ListCommands() []CommandState {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]CommandState, 0, len(h.commands))
	for _, command := range h.commands {
		out = append(out, command)
	}
	return out
}

func (h *RigHub) GetCommand(id string) (CommandState, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	command, ok := h.commands[id]
	return command, ok
}

func (h *RigHub) SendCommand(rigID, namespace, command string, target *protocol.Target, data json.RawMessage) (CommandState, error) {
	h.mu.RLock()
	r := h.rigs[rigID]
	h.mu.RUnlock()
	if r == nil {
		return CommandState{}, errors.New("rig not connected")
	}
	id := randomID("op_")
	pkt := protocol.NewCommand(id, namespace, command, target, data)
	state := CommandState{ID: id, RigID: rigID, Namespace: namespace, Command: command, Phase: "command", Data: data, UpdatedAt: time.Now().UTC()}
	h.mu.Lock()
	h.commands[id] = state
	h.mu.Unlock()
	r.send <- pkt
	return state, nil
}

func (h *RigHub) readLoop(r *connectedRig) {
	defer func() {
		r.conn.Close()
		h.mu.Lock()
		if h.rigs[r.id] == r {
			r.snapshot.Online = false
			r.snapshot.LastSeen = time.Now().UTC()
			delete(h.rigs, r.id)
		}
		h.mu.Unlock()
	}()
	for {
		var pkt protocol.Packet
		if err := r.conn.ReadJSON(&pkt); err != nil {
			return
		}
		h.handlePacket(r, pkt)
	}
}

func (h *RigHub) writeLoop(r *connectedRig) {
	for pkt := range r.send {
		if err := r.conn.WriteJSON(pkt); err != nil {
			return
		}
	}
}

func (h *RigHub) handlePacket(r *connectedRig, pkt protocol.Packet) {
	now := time.Now().UTC()
	h.mu.Lock()
	defer h.mu.Unlock()
	r.snapshot.LastSeen = now
	switch pkt.Action {
	case "register":
		r.snapshot.Adapter = pkt.Adapter
		r.snapshot.Capabilities = pkt.Capabilities
	case "keep_alive":
		r.snapshot.State = pkt.State
	case "send":
		if pkt.ID != "" {
			cs := CommandState{ID: pkt.ID, RigID: r.id, Namespace: pkt.Namespace, Command: pkt.Command, Phase: pkt.Phase, Data: pkt.Data, Error: pkt.Error, UpdatedAt: now}
			h.commands[pkt.ID] = cs
			switch pkt.Phase {
			case "progress":
				r.snapshot.Active[pkt.ID] = cs
			case "result", "error":
				delete(r.snapshot.Active, pkt.ID)
			}
		}
	}
}

func (h *RigHub) authorized(r *http.Request) bool {
	if h.token == "" {
		return true
	}
	if r.Header.Get("Authorization") == "Bearer "+h.token {
		return true
	}
	return r.URL.Query().Get("token") == h.token
}

func randomID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return prefix + hex.EncodeToString(b)
}
