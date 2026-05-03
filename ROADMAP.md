# Mission Control Roadmap

## Project vision

Mission Control is intended to become an ultimate telescope orchestration platform.

The long-term goal is to coordinate one or more telescope rigs, each potentially running multiple hardware-control interfaces such as NINA, PHD2, ASTAP, PlateSolve 3, and eventually other astronomy systems.

The system should support both direct user control and agentic/AI control so users do not need to manually configure and coordinate every subsystem.

## Project layout

- `gateway/` — control plane, rig registry, frontend/API surface
- `rig/` — node/server process that runs near telescope hardware and talks to local astronomy software/hardware
- `protocol/` — simple WebSocket protocol between rig and gateway
- `nina-api/` — generated client used to communicate with NINA's API
- `app/` — React frontend application

## Current architecture direction

The preferred architecture is:

1. The `gateway` acts as the central control plane.
2. Each `rig` connects outward to the gateway over WebSocket.
3. The frontend talks only to the gateway.
4. The gateway routes commands to connected rigs.
5. The rig translates protocol commands into local adapter calls, such as NINA, PHD2, ASTAP, or simulated hardware.

This outbound rig connection model works well for observatories, home networks, NAT, remote rigs, and multiple distributed nodes.

## Current mismatch to resolve

The protocol currently describes this model:

- rig connects outward to gateway over WebSocket
- rig sends `register`
- rig sends periodic `keep_alive`
- gateway sends command packets using `send`
- rig replies with progress/result/error packets using `send`

The current code is not fully aligned yet:

- `rig` currently runs as an HTTP server
- `gateway` currently exposes simple REST stubs for rig register/send/status
- the WebSocket rig hub is not implemented yet

The MVP should align implementation with the protocol.

## MVP 1: Gateway + rig live connection

The first useful MVP should prove the control-plane loop without requiring telescope hardware.

### Gateway MVP

Implement:

- `GET /v1/rigs`
- `GET /v1/rigs/:id`
- `POST /v1/rigs/:id/commands`
- `GET /v1/ws/rig` for rig WebSocket connections

The gateway should:

- accept outbound WebSocket connections from rigs
- receive `register` packets
- store connected rig state in memory
- receive `keep_alive` packets
- mark rigs online/offline
- route commands to a connected rig
- track command progress/result/error by operation id

### Rig MVP

The rig should:

- read config
- connect to the gateway WebSocket
- send `register`
- send periodic `keep_alive`
- receive `send` command packets
- handle simulated commands
- return fake `progress` and `result` packets

### Simulated command set

Initial simulated commands:

- `rig.get_status`
- `mount.goto_radec`
- `mount.park`
- `camera.capture`
- `sequence.start`
- `sequence.stop`

This allows the gateway, protocol, and UI to be tested without telescope hardware.

## MVP 2: React app

The frontend should talk only to the gateway.

Initial app functionality:

- list registered/connected rigs
- show online/offline state
- show latest keep-alive state
- show capabilities
- send test commands
- show command progress/result/error

A minimal useful flow:

1. Start gateway.
2. Start one simulated rig.
3. Open app.
4. See the rig online.
5. Send `rig.get_status`.
6. See the result.
7. Send a simulated mount/camera command.
8. See progress and completion.

## MVP 3: NINA adapter

After the simulated rig loop works, add real NINA integration.

The rig should support one or more NINA instances.

Example future config:

```yaml
id: backyard-node
display_name: Backyard Observatory Node

gateway:
  url: ws://localhost:8080/v1/ws/rig
  token: dev-token

adapters:
  - id: nina_left
    type: nina
    host: http://localhost:5000

  - id: nina_right
    type: nina
    host: http://localhost:5001
```

The rig should expose capabilities based on configured adapters.

Example capabilities:

- `nina_left.mount.goto_radec`
- `nina_left.mount.park`
- `nina_left.camera.capture`
- `nina_left.sequence.start`
- `nina_right.mount.goto_radec`
- `nina_right.camera.capture`

## Protocol improvement: adapter targets

For dual-rig and multi-adapter setups, distinguish between:

- rig — physical/server node
- adapter instance — one NINA/PHD2/ASTAP/etc. connection under that rig

Recommended capability shape:

```json
{
  "adapter_id": "nina_1",
  "adapter": "nina",
  "namespace": "mount",
  "commands": ["goto_radec", "park", "unpark"]
}
```

Recommended command target shape:

```json
{
  "v": "mc.v1",
  "action": "send",
  "id": "op_123",
  "ts": "2026-04-24T18:00:01Z",
  "namespace": "mount",
  "command": "park",
  "phase": "command",
  "target": {
    "adapter_id": "nina_1"
  },
  "data": {}
}
```

This keeps the protocol small while allowing:

- one rig with one NINA instance
- one rig with multiple NINA instances
- gateway umbrella commands across multiple adapters
- future adapters like PHD2, ASTAP, PlateSolve 3, ASCOM/Alpaca, INDI, etc.

## Future integrations

### Plate solving

Potential adapters:

- ASTAP
- PlateSolve 3
- NINA's plate-solving facilities

Possible commands:

- `platesolver.solve_image`
- `platesolver.solve_and_sync`
- `platesolver.center_target`

### Guiding

Potential adapter:

- PHD2

Possible commands/events:

- `guider.connect`
- `guider.start_guiding`
- `guider.stop_guiding`
- `guider.dither`
- `guider.get_graph`
- guide graph telemetry events

### Sequencing

Possible commands:

- `sequence.start`
- `sequence.stop`
- `sequence.pause`
- `sequence.resume`
- `sequence.status`

### Safety and environment

Possible namespaces:

- `weather`
- `safety`
- `dome`
- `power`
- `environment`

Possible events:

- unsafe weather
- rain detected
- wind threshold exceeded
- dome disconnected
- camera disconnected
- mount fault
- sequence completed

## Agentic / AI control direction

The AI agent should interact with the gateway API, not directly with rig internals.

The gateway should expose enough structured state for an agent to:

- list rigs
- inspect capabilities
- inspect current state
- understand active operations
- propose an imaging plan
- execute commands
- monitor progress
- recover from failures

The agent should eventually be able to answer and execute requests like:

- "Image M31 tonight with the RedCat rig."
- "Check if the rig is safe to open."
- "Center this target and start guiding."
- "Diagnose why the sequence failed."
- "Pause imaging if clouds roll in."

## Implementation order

Recommended order:

1. Add protocol structs for `register`, `keep_alive`, and `send`.
2. Implement gateway WebSocket rig hub.
3. Implement rig outbound WebSocket client mode.
4. Add simulated rig command handling.
5. Add gateway REST endpoints for listing rigs and sending commands.
6. Add in-memory command tracking.
7. Build minimal React app against gateway REST API.
8. Add multi-adapter config model to rig.
9. Add NINA adapter using `nina-api`.
10. Add PHD2 guiding adapter.
11. Add plate-solving adapter support.
12. Add orchestration workflows.
13. Add AI/agent API layer.

## Local MVP goal

A good first local test should be:

```bash
cd gateway && go run .
cd rig && go run . connect --gateway ws://localhost:8080/v1/ws/rig
```

Then use the gateway API or frontend to:

1. list connected rigs
2. inspect rig capabilities
3. send `rig.get_status`
4. send a simulated `mount.goto_radec`
5. receive progress/result packets

Once this works, hardware-specific integrations can be layered in without changing the overall architecture.
