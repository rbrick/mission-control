# Mission Control App

Early React/Vite frontend for the Mission Control MVP.

## Run locally

Start the gateway first:

```bash
cd ../gateway
go run .
```

Start a rig:

```bash
cd ../rig
go run . connect
```

Start the app:

```bash
cd ../app
npm install
npm run dev
```

Open `http://localhost:5173`.

## Gateway URL

In development, Vite proxies `/v1` to `http://127.0.0.1:8080` by default.

If the gateway runs elsewhere:

```bash
VITE_GATEWAY_URL=http://127.0.0.1:8090 npm run dev
```

If you do not want to use the dev proxy, set an explicit browser-facing API base:

```bash
VITE_API_BASE=http://127.0.0.1:8080 npm run dev
```

A `502` from the app usually means the Vite proxy cannot reach the gateway. Make sure the gateway is running and the port matches `VITE_GATEWAY_URL`.

## Scope

This frontend is intentionally minimal until the rig/gateway APIs settle. The page structure reserves areas for:

- dashboard
- imaging/image preview
- guiding/guide graph
- agent chat
