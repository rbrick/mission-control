#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PORT="${PORT:-8099}"
HOST="127.0.0.1:${PORT}"
GATEWAY_URL="http://${HOST}"
RIG_WS_URL="ws://${HOST}/v1/ws/rig"
TOKEN="${GATEWAY_RIG_TOKEN:-test-token}"
KEEP_ALIVE="${KEEP_ALIVE:-500ms}"
GW_LOG="${TMPDIR:-/tmp}/mission-control-gateway-mvp.log"
RIG_LOG="${TMPDIR:-/tmp}/mission-control-rig-mvp.log"
GW_PID=""
RIG_PID=""

cleanup() {
  if [[ -n "${RIG_PID}" ]]; then kill "${RIG_PID}" 2>/dev/null || true; fi
  if [[ -n "${GW_PID}" ]]; then kill "${GW_PID}" 2>/dev/null || true; fi
}
trap cleanup EXIT

require() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

wait_for() {
  local name="$1"
  local command="$2"
  for _ in $(seq 1 50); do
    if eval "$command" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.2
  done
  echo "timed out waiting for ${name}" >&2
  echo "--- gateway log ---" >&2
  tail -50 "${GW_LOG}" >&2 || true
  echo "--- rig log ---" >&2
  tail -50 "${RIG_LOG}" >&2 || true
  exit 1
}

json_get() {
  python3 -c 'import json,sys; data=json.load(sys.stdin); print(eval(sys.argv[1], {}, {"data": data}))' "$1"
}

require curl
require python3
require go

rm -f "${GW_LOG}" "${RIG_LOG}"

echo "starting gateway on ${HOST}"
(
  cd "${ROOT_DIR}/gateway"
  HOST="${HOST}" GATEWAY_RIG_TOKEN="${TOKEN}" COMMAND_TIMEOUT="5s" go run .
) >"${GW_LOG}" 2>&1 &
GW_PID=$!

wait_for "gateway" "curl -sf '${GATEWAY_URL}/v1/rigs'"

echo "starting simulated rig"
(
  cd "${ROOT_DIR}/rig"
  go run . connect --gateway "${RIG_WS_URL}" --token "${TOKEN}" --keep-alive "${KEEP_ALIVE}"
) >"${RIG_LOG}" 2>&1 &
RIG_PID=$!

wait_for "rig registration" "curl -sf '${GATEWAY_URL}/v1/rigs' | grep -q 'sfro-rc-91'"

echo "connected rigs:"
curl -sf "${GATEWAY_URL}/v1/rigs"
echo

echo "sending rig.get_status"
COMMAND_RESPONSE="$(curl -sf -X POST "${GATEWAY_URL}/v1/rigs/sfro-rc-91/commands" \
  -H 'content-type: application/json' \
  -d '{"namespace":"rig","command":"get_status"}')"
COMMAND_ID="$(printf '%s' "${COMMAND_RESPONSE}" | json_get 'data["id"]')"

wait_for "command result" "curl -sf '${GATEWAY_URL}/v1/commands/${COMMAND_ID}' | grep -q '\"phase\":\"result\"'"

echo "command result:"
curl -sf "${GATEWAY_URL}/v1/commands/${COMMAND_ID}"
echo

echo "sending mount.goto_radec"
GOTO_RESPONSE="$(curl -sf -X POST "${GATEWAY_URL}/v1/rigs/sfro-rc-91/commands" \
  -H 'content-type: application/json' \
  -d '{"namespace":"mount","command":"goto_radec","data":{"ra_hours":10.684,"dec_degrees":41.269,"epoch":"J2000"}}')"
GOTO_ID="$(printf '%s' "${GOTO_RESPONSE}" | json_get 'data["id"]')"

wait_for "goto result" "curl -sf '${GATEWAY_URL}/v1/commands/${GOTO_ID}' | grep -q '\"phase\":\"result\"'"

echo "goto result:"
curl -sf "${GATEWAY_URL}/v1/commands/${GOTO_ID}"
echo

echo "MVP smoke test passed"
