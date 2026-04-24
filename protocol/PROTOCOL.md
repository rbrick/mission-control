# Mission Control Protocol

## Overview

Mission Control operates over a WebSocket connection between each rig and the gateway.

The protocol has exactly three actions:

- `register`
- `keep_alive`
- `send`

That is the full protocol surface.

`register` is how a rig establishes its identity and declares its capabilities.

`keep_alive` is how a rig announces that it is still online and publishes current summary state.

`send` is how commands, progress, results, errors, and unsolicited events move across the connection.

## Roles

There are two roles on this connection:

- `rig`
- `gateway`

The rig connects outward to the gateway over WebSocket.

The normal flow is:

1. The rig connects to the gateway.
2. The rig sends `register` once for the connection.
3. The rig sends `keep_alive` periodically.
4. The gateway sends `send` packets to issue commands.
5. The rig sends `send` packets back with progress, results, or errors.

The UI is out of scope for this document. The gateway may expose a different protocol to the UI, or it may reuse the same message model internally.

## Transport

- Transport: WebSocket
- Encoding: JSON
- Protocol version: `mc.v1`
- Connection health: WebSocket ping/pong

The application protocol does not define any extra transport heartbeat beyond `keep_alive`.

`register` exists to publish rig identity and capabilities.

`keep_alive` exists to publish rig presence and current state to the gateway.

## Common Fields

Every packet must include these fields:

```json
{
  "v": "mc.v1",
  "action": "register"
}
```

### Common packet fields

- `v`: protocol version
- `action`: `register`, `keep_alive`, or `send`

Most packets will also include:

- `ts`: UTC timestamp in RFC3339 format

Rig identity is connection-scoped.

The gateway determines which internal rig record a connection belongs to through its authentication or registration context, not from a `rig_id` carried in every packet.

`display_name` is also gateway-owned metadata and is not part of the wire protocol.

## Action: `register`

`register` is sent by the rig immediately after the WebSocket connection is established.

It has three jobs:

1. Identify the rig.
2. Declare the adapter in use.
3. Declare the command surface exposed by the rig.

Each connection must begin with a `register` packet before any `keep_alive` or `send` traffic.

If the rig's capabilities change while connected, the rig must send a new `register` packet with the updated capability set.

`register` does not carry a gateway-owned rig identifier. The gateway binds the connection to its internal rig record out of band.

### Shape

```json
{
  "v": "mc.v1",
  "action": "register",
  "ts": "2026-04-24T18:00:00Z",
  "adapter": "nina",
  "capabilities": [
    {
      "namespace": "mount",
      "commands": ["goto_radec", "goto_altaz", "park", "unpark", "abort"]
    },
    {
      "namespace": "camera",
      "commands": ["capture"]
    },
    {
      "namespace": "focuser",
      "commands": ["move", "run_autofocus"]
    }
  ]
}
```

### Required fields

- `ts`
- `capabilities`

### Registration rules

- The gateway must not issue `send` commands to a rig until it has received `register`.
- The latest `register` packet is the source of truth for capabilities.
- A reconnect requires a new `register` packet.

## Action: `keep_alive`

`keep_alive` is sent by the rig to the gateway on a fixed interval.

It has two jobs:

1. Confirm the rig is still online.
2. Publish a compact snapshot of current state.

### Shape

```json
{
  "v": "mc.v1",
  "action": "keep_alive",
  "ts": "2026-04-24T18:00:00Z",
  "interval_ms": 5000,
  "state": {
    "connected": true,
    "safety": "safe",
    "active": [
      {
        "id": "op_123",
        "namespace": "mount",
        "command": "goto_radec",
        "phase": "progress"
      }
    ]
  }
}
```

### Required fields

- `ts`
- `interval_ms`
- `state`

### `state`

`state` is a compact snapshot, not a full telemetry dump.

It should contain only the status the gateway needs to display availability and route commands safely.

Recommended fields:

- `connected`
- `safety`
- `active`
- `faults`

### Keep-alive timeout

If the gateway does not receive a `keep_alive` from a rig within `interval_ms * 3`, the rig should be considered offline.

## Action: `send`

`send` is the general-purpose message packet.

It is used for:

- gateway to rig commands
- rig to gateway progress updates
- rig to gateway final results
- rig to gateway errors
- rig to gateway unsolicited events

The protocol still has only one message action here: `send`.

The meaning of a `send` packet is determined by its `phase`.

## `send` Shape

```json
{
  "v": "mc.v1",
  "action": "send",
  "id": "op_123",
  "ts": "2026-04-24T18:00:01Z",
  "namespace": "mount",
  "command": "goto_radec",
  "phase": "command",
  "data": {
    "ra_hours": 10.684,
    "dec_degrees": 41.269,
    "epoch": "J2000"
  }
}
```

### Required fields

- `id`
- `ts`
- `namespace`
- `command`
- `phase`

### Field meanings

- `id`: correlation identifier for one command lifecycle
- `namespace`: logical subsystem, such as `mount` or `camera`
- `command`: operation name inside the namespace
- `phase`: the meaning of this message
- `data`: payload for the message
- `error`: structured error object when `phase` is `error`

## `send.phase`

`send` supports these phases:

- `command`
- `progress`
- `result`
- `error`
- `event`

### `command`

Sent by the gateway to the rig to start an operation.

Example:

```json
{
  "v": "mc.v1",
  "action": "send",
  "id": "op_123",
  "ts": "2026-04-24T18:00:01Z",
  "namespace": "mount",
  "command": "goto_radec",
  "phase": "command",
  "data": {
    "ra_hours": 10.684,
    "dec_degrees": 41.269,
    "epoch": "J2000"
  }
}
```

### `progress`

Sent by the rig to report that an operation is underway.

The gateway should expect zero or more `progress` packets for a command.

Example:

```json
{
  "v": "mc.v1",
  "action": "send",
  "id": "op_123",
  "ts": "2026-04-24T18:00:04Z",
  "namespace": "mount",
  "command": "goto_radec",
  "phase": "progress",
  "data": {
    "state": "slewing",
    "progress": 0.42,
    "mount": {
      "ra_hours": 10.12,
      "dec_degrees": 39.85
    }
  }
}
```

### `result`

Sent by the rig to indicate successful completion.

`result` is terminal for the command identified by `id`.

Example:

```json
{
  "v": "mc.v1",
  "action": "send",
  "id": "op_123",
  "ts": "2026-04-24T18:00:08Z",
  "namespace": "mount",
  "command": "goto_radec",
  "phase": "result",
  "data": {
    "arrived": true
  }
}
```

### `error`

Sent by the rig when a command fails.

`error` is terminal for the command identified by `id`.

Example:

```json
{
  "v": "mc.v1",
  "action": "send",
  "id": "op_123",
  "ts": "2026-04-24T18:00:03Z",
  "namespace": "mount",
  "command": "goto_radec",
  "phase": "error",
  "error": {
    "code": "INVALID_ARGUMENT",
    "message": "target below horizon"
  }
}
```

### `event`

Sent by the rig for unsolicited notifications not tied to a gateway-issued command.

Examples:

- safety state changed
- camera disconnected
- sequence completed locally
- weather warning raised

Example:

```json
{
  "v": "mc.v1",
  "action": "send",
  "id": "evt_9001",
  "ts": "2026-04-24T18:01:00Z",
  "namespace": "rig",
  "command": "state_changed",
  "phase": "event",
  "data": {
    "safety": "unsafe",
    "reason": "rain"
  }
}
```

## Correlation Rules

For a gateway-issued command:

1. The gateway creates `id`.
2. Every rig `progress`, `result`, or `error` packet for that command must reuse the same `id`.
3. `result` and `error` are terminal.
4. No further packets may be sent for that `id` after a terminal packet.

This keeps command tracking simple and preserves the original request-response matching idea.

## Command Model

The protocol does not define separate packet types for each telescope function.

Instead, the gateway sends commands into a namespace.

Examples:

- `mount.goto_radec`
- `mount.goto_altaz`
- `mount.park`
- `mount.unpark`
- `mount.abort`
- `camera.capture`
- `sequence.start`
- `sequence.stop`
- `focuser.move`
- `focuser.run_autofocus`

This keeps the protocol small while still allowing a rich command surface.

## Error Model

When `phase` is `error`, the packet must include an `error` object.

```json
{
  "code": "NOT_SUPPORTED",
  "message": "command not supported by adapter",
  "details": {}
}
```

Recommended error codes:

- `UNAUTHENTICATED`
- `FORBIDDEN`
- `NOT_FOUND`
- `NOT_SUPPORTED`
- `INVALID_ARGUMENT`
- `BUSY`
- `TIMEOUT`
- `CANCELLED`
- `HARDWARE_FAULT`
- `INTERNAL`
- `UNAVAILABLE`

## Capability Rules

Capabilities are declared by the rig, not assumed by the gateway.

The gateway must treat the latest `register.capabilities` payload as the source of truth.

If a namespace or command is absent from the latest `register`, the gateway must assume it is unavailable.

## State Rules

The protocol splits rig updates into two categories:

1. `register` for identity and capability declaration.
2. `keep_alive` for periodic presence and summary state.
3. `send` for command-specific progress and discrete events.

This prevents the gateway from needing to infer rig health from command traffic.

## Ordering Rules

- Packets are processed in receive order on a single WebSocket connection.
- Command lifecycles are correlated by `id`.
- Different command IDs may interleave freely.

## Minimal Command Set

A minimal useful rig should probably expose at least some of these namespaces:

- `rig`
- `mount`
- `camera`
- `focuser`
- `sequence`

Example base commands:

- `rig.get_status`
- `mount.goto_radec`
- `mount.abort`
- `camera.capture`
- `focuser.move`

## Design Intent

The protocol is intentionally small:

- one packet for registration and capability advertisement
- one packet for periodic presence and summary state
- one packet for actual message exchange

That keeps the wire format simple, while the namespace plus command model keeps the behavior powerful.