# Mission Control Protocol

Status: Draft v0.1

## 1. Purpose

Mission Control needs a protocol that cleanly separates:

- AI agents, which plan and decide
- the orchestrator/gateway, which authorizes, schedules, and routes work
- rigs, which execute telescope/imaging operations
- external astronomy control systems such as NINA, ASCOM, and later INDI

The protocol must be:

- platform-agnostic at the Mission Control layer
- explicit about authority boundaries
- safe for remote execution
- versioned and evolvable
- suitable for both human-issued commands and AI-generated plans

This document defines the first protocol model for Mission Control.

---

## 2. Core Principle

**Agents do not control hardware directly.**

The chain of authority is:

1. User expresses intent in the app
2. AI agent turns intent into a proposed plan
3. Gateway/orchestrator validates and approves the plan
4. Gateway issues structured rig commands
5. Rig executes those commands against NINA/ASCOM/INDI adapters
6. Rig reports state, events, and results back to the gateway

So:

- **agent** = planner / assistant / tool-calling intelligence
- **rig** = telescope node / hardware executor
- **gateway** = source of orchestration truth

---

## 3. System Roles

### 3.1 App

The user-facing client.

Responsibilities:

- chat and operator UI
- reviewing plans
- approving or rejecting execution
- viewing rig state and history

The app never talks to rigs directly.

### 3.2 Agent

An AI planning component.

Responsibilities:

- interpret user intent
- generate structured imaging plans
- call orchestrator tools
- explain proposed actions

The agent never talks to rigs directly.

### 3.3 Gateway / Orchestrator

The control plane.

Responsibilities:

- authentication and authorization
- rig registry
- agent/tool interface
- plan validation
- approval workflow
- scheduling and routing
- audit log and execution history
- state projection for app/UI

The gateway is the only authority allowed to issue executable rig commands.

### 3.4 Rig

A telescope execution node.

Responsibilities:

- advertise capabilities
- accept structured commands from gateway
- translate commands into local adapter operations
- report state, events, failures, and artifacts

The rig does not plan. The rig executes.

### 3.5 Adapter

A local integration layer inside the rig.

Examples:

- NINA adapter
- ASCOM adapter
- INDI adapter

Adapters are implementation details behind the rig protocol.

---

## 4. Protocol Layers

Mission Control should be treated as three separate protocols.

## 4.1 Agent Protocol

Between agent and gateway.

Purpose:

- propose plans
- inspect rigs and capabilities
- request validation
- request execution after approval

This is a tool/API contract, not a hardware protocol.

## 4.2 Control Protocol

Between gateway and rigs.

Purpose:

- register rigs
- exchange health/state
- deliver executable commands
- report progress and results

This is the main Mission Control protocol.

## 4.3 Data Protocol

For larger artifacts and streaming state.

Purpose:

- thumbnails
- latest sub exposure
- FITS metadata
- logs
- event streams

This should remain separate from the core command channel.

---

## 5. Transport Recommendation

### 5.1 Agent <-> Gateway

- HTTPS JSON API
- optionally SSE or WebSocket for streaming plan updates

### 5.2 App <-> Gateway

- HTTPS JSON API
- WebSocket or SSE for live status updates

### 5.3 Gateway <-> Rig

Recommended first transport:

- HTTPS JSON
- rig-initiated outbound connection model

Reason:

- easier across NAT/firewalls
- safer for observatory PCs
- gateway remains central authority

Recommended interaction pattern:

- rig registers with gateway
- rig sends heartbeat/state updates
- rig polls or long-polls for commands
- rig posts command acknowledgements and results

Later, this can evolve to WebSocket if needed.

---

## 6. Identity Model

Every message must carry stable identities.

### 6.1 Principal IDs

- `agent_id`: AI planner identity
- `session_id`: app/chat session
- `plan_id`: proposed plan
- `execution_id`: approved execution instance
- `rig_id`: logical rig identity
- `command_id`: individual executable command
- `event_id`: emitted state/event record

### 6.2 Rule

`rig_id` identifies the logical rig, not the transport connection and not the AI agent.

---

## 7. Capability Model

Rigs must advertise capabilities, not implementation brands.

Examples:

- `mount.slew`
- `mount.park`
- `camera.expose`
- `camera.cool`
- `filterwheel.select`
- `focuser.move`
- `focuser.autofocus`
- `sequence.load`
- `sequence.start`
- `sequence.stop`
- `image.preview.latest`
- `weather.read`

Optional implementation metadata may also be reported:

- adapter type: `nina`, `indi`, `custom`
- local platform: `windows`, `linux`
- external systems: `ASCOM`, `INDI`, `Alpaca`

But orchestrator decisions should target capabilities first.

---

## 8. State Model

## 8.1 Rig State

Minimum canonical rig states:

- `offline`
- `idle`
- `preparing`
- `ready`
- `running`
- `paused`
- `stopping`
- `error`
- `maintenance`

## 8.2 Command State

- `queued`
- `dispatched`
- `acknowledged`
- `running`
- `succeeded`
- `failed`
- `canceled`
- `timed_out`

## 8.3 Execution State

- `draft`
- `awaiting_approval`
- `approved`
- `scheduled`
- `running`
- `completed`
- `failed`
- `canceled`

---

## 9. Message Families

## 9.1 Rig Registration

Purpose: declare a rig exists and can participate.

Example shape:

```json
{
	"protocol_version": "0.1",
	"rig_id": "rig-west-1",
	"display_name": "West Observatory",
	"adapter": {
		"type": "nina",
		"version": "3.x"
	},
	"platform": {
		"os": "windows"
	},
	"capabilities": [
		"mount.slew",
		"camera.expose",
		"sequence.load",
		"sequence.start",
		"sequence.stop"
	]
}
```

## 9.2 Rig Heartbeat

Purpose: liveness and summary health.

Example:

```json
{
	"rig_id": "rig-west-1",
	"ts": "2026-03-15T01:02:03Z",
	"state": "ready",
	"health": {
		"ok": true,
		"issues": []
	},
	"sequence": {
		"active": false
	}
}
```

## 9.3 Rig Status Snapshot

Purpose: a richer state document for UI and orchestration.

Should include:

- mount state
- camera state
- filter state
- guider state
- focuser state
- weather/safety state
- active target
- active sequence summary
- latest error summary

## 9.4 Rig Command

Purpose: executable structured request from gateway to rig.

Example envelope:

```json
{
	"command_id": "cmd_123",
	"execution_id": "exec_456",
	"rig_id": "rig-west-1",
	"type": "sequence.start",
	"payload": {
		"sequence_ref": "seq_abc",
		"skip_validation": false
	}
}
```

## 9.5 Rig Command Result

Purpose: terminal or progress result for a command.

Example:

```json
{
	"command_id": "cmd_123",
	"rig_id": "rig-west-1",
	"state": "succeeded",
	"message": "sequence started",
	"ts": "2026-03-15T01:04:10Z"
}
```

## 9.6 Rig Event

Purpose: append-only event stream.

Examples:

- `mount.slew.started`
- `mount.slew.completed`
- `camera.exposure.started`
- `camera.exposure.completed`
- `sequence.item.started`
- `sequence.item.failed`
- `safety.abort.triggered`

Events should be immutable and timestamped.

## 9.7 Artifact Notification

Purpose: report that a new image or artifact exists.

Examples:

- latest sub thumbnail
- FITS file metadata
- plate solve result
- autofocus run output

---

## 10. Agent Protocol

The AI agent should work in terms of tools exposed by the gateway.

Recommended tool surface:

- `list_rigs`
- `get_rig_status`
- `list_rig_capabilities`
- `propose_imaging_plan`
- `validate_imaging_plan`
- `estimate_execution`
- `request_execution_approval`
- `start_approved_execution`
- `cancel_execution`
- `get_execution_status`

Important rule:

The agent should emit **plans**, not raw hardware instructions.

Bad:

- "call NINA endpoint X"

Good:

- "load sequence template LRGB_DSO"
- "set target Orion Nebula"
- "start execution on rig-west-1"

---

## 11. Imaging Plan Model

The protocol should define a gateway-owned, structured imaging plan.

Suggested shape:

- target
- rig selection
- instrument/profile selection
- framing intent
- filters
- exposure settings per filter
- counts / duration goals
- constraints
- calibration requirements
- safety policy
- operator notes

Example conceptual plan:

```json
{
	"target": {
		"name": "Orion Nebula"
	},
	"rig_id": "rig-west-1",
	"profile": "Redcat 91",
	"sequence": {
		"mode": "LRGB",
		"filters": [
			{"name": "L", "exposure_s": 60, "sub_count": 24},
			{"name": "R", "exposure_s": 30, "sub_count": 10},
			{"name": "G", "exposure_s": 30, "sub_count": 10},
			{"name": "B", "exposure_s": 30, "sub_count": 10}
		]
	}
}
```

The gateway then translates this into adapter-specific rig commands.

---

## 12. Command Semantics

Commands must be:

- structured
- idempotent where possible
- auditable
- attributable to an execution and approval chain

Initial command families:

- `rig.sync_state`
- `sequence.load_template`
- `sequence.apply_plan`
- `sequence.start`
- `sequence.pause`
- `sequence.resume`
- `sequence.stop`
- `mount.park`
- `mount.unpark`
- `session.abort`

Avoid exposing low-level vendor-specific commands in the public Mission Control protocol.

---

## 13. Safety Model

The protocol must support explicit safety interlocks.

Examples:

- weather unsafe
- roof/dome unsafe
- mount not parked when required
- camera not cooled
- guider unavailable
- manual lock by operator

The gateway may refuse to dispatch.
The rig may refuse to execute.

Both refusals must be represented as structured failures.

---

## 14. Error Model

All protocol errors should include:

- machine code
- human message
- retriable flag
- source

Example:

```json
{
	"error": {
		"code": "RIG_CAPABILITY_MISSING",
		"message": "Rig does not support sequence.apply_plan",
		"retriable": false,
		"source": "gateway"
	}
}
```

Sources:

- `app`
- `agent`
- `gateway`
- `rig`
- `adapter`
- `external_system`

---

## 15. Versioning

Every protocol exchange should include:

- `protocol_version`
- sender identity
- timestamp

Rules:

- additive fields are preferred
- receivers must ignore unknown fields
- breaking changes require a new protocol version

Suggested initial version: `0.1`

---

## 16. Security

Minimum requirements:

- authenticated agents
- authenticated rigs
- TLS for remote transport
- signed or token-authenticated rig requests
- audit trail from user -> agent -> gateway -> rig command

Desired audit linkage:

- `session_id`
- `agent_id`
- `plan_id`
- `execution_id`
- `command_id`

---

## 17. Recommended First Milestone

Define and implement only these first:

### Agent <-> Gateway

- `list_rigs`
- `get_rig_status`
- `propose_imaging_plan`
- `validate_imaging_plan`

### Gateway <-> Rig

- `register`
- `heartbeat`
- `status.snapshot`
- `command.dispatch`
- `command.result`

### Commands

- `sequence.load_template`
- `sequence.start`
- `sequence.stop`

Do not start with:

- direct vendor-specific procedure calls
- arbitrary agent-issued hardware actions
- image streaming in the same protocol channel

---

## 18. Naming Rules

To avoid confusion:

- use **agent** only for AI planning/execution intelligence
- use **rig** only for telescope execution nodes
- use **adapter** for local NINA/ASCOM/INDI integration
- use **gateway** or **orchestrator** for the central control plane

Never use `agent` to mean `rig`.

---

## 19. Summary

Mission Control should define:

1. an **agent protocol** for planning
2. a **gateway-to-rig protocol** for execution
3. a separate **data/artifact protocol** for images and previews

The key architectural rule is:

**AI agents propose and reason. Rigs execute. The gateway authorizes and orchestrates.**

---

## 20. Next Draft Targets

The next revision of this document should define:

- exact JSON schemas
- transport auth headers
- command retry semantics
- event taxonomy
- execution approval states
- artifact delivery model
- adapter capability mapping for NINA vs INDI
