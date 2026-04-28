#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCENARIO_FILE="${SCENARIO_FILE:-$ROOT_DIR/bench/scenarios/header-auth.env}"

if [[ ! -f "$SCENARIO_FILE" ]]; then
  echo "error: scenario file not found: $SCENARIO_FILE" >&2
  exit 1
fi

# shellcheck disable=SC1090
source "$SCENARIO_FILE"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "error: required command '$1' not found" >&2
    exit 1
  fi
}

require_cmd curl
require_cmd python3
require_cmd vegeta
require_cmd go

WASMFORGE_BIN="${WASMFORGE_BIN:-$ROOT_DIR/bin/wasmforge}"
UPSTREAM_BIN="${UPSTREAM_BIN:-$ROOT_DIR/bench/bin/upstream}"
NATIVE_BIN="${NATIVE_BIN:-$ROOT_DIR/bench/bin/native-gateway}"
PLUGIN_WASM_FILE="${PLUGIN_WASM_FILE:-$ROOT_DIR/bench/artifacts/header_auth_bench.wasm}"

for file in "$WASMFORGE_BIN" "$UPSTREAM_BIN" "$NATIVE_BIN"; do
  if [[ ! -x "$file" ]]; then
    echo "error: executable not found: $file" >&2
    echo "hint: run 'make build && make bench-build'" >&2
    exit 1
  fi
done

if [[ ! -f "$PLUGIN_WASM_FILE" ]]; then
  echo "error: plugin file not found: $PLUGIN_WASM_FILE" >&2
  echo "hint: run ./bench/scripts/build_plugin.sh" >&2
  exit 1
fi

STAMP="$(date +%Y%m%d-%H%M%S)"
STAMP_SAFE="${STAMP//-/_}"
BENCH_PATH="${REQUEST_PATH}-${STAMP}"
PLUGIN_NAME="header_auth_bench_${STAMP_SAFE}"
PLUGIN_FILENAME="header_auth_bench_${STAMP_SAFE}.wasm"
RESULT_DIR="$ROOT_DIR/bench/results/$STAMP"
TMP_DIR="$ROOT_DIR/bench/tmp/$STAMP"
UPLOADS_DIR="$TMP_DIR/uploads"
LOGS_DIR="$TMP_DIR/logs"
CERTS_DIR="$TMP_DIR/certs"
DB_PATH="$TMP_DIR/wasmforge.db"

mkdir -p "$RESULT_DIR" "$TMP_DIR" "$UPLOADS_DIR" "$LOGS_DIR" "$CERTS_DIR"

cleanup() {
  local pids=("${WASMFORGE_PID:-}" "${UPSTREAM_PID:-}" "${NATIVE_PID:-}")
  for pid in "${pids[@]}"; do
    if [[ -n "$pid" ]] && kill -0 "$pid" >/dev/null 2>&1; then
      kill "$pid" >/dev/null 2>&1 || true
      wait "$pid" >/dev/null 2>&1 || true
    fi
  done
}
trap cleanup EXIT

wait_http_ok() {
  local url="$1"
  local retries="${2:-60}"
  local delay="${3:-0.25}"
  for _ in $(seq 1 "$retries"); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep "$delay"
  done
  echo "error: timed out waiting for $url" >&2
  exit 1
}

extract_json_field() {
  local field="$1"
  local payload="$2"
  python3 - "$field" "$payload" <<'PY'
import json
import sys

field = sys.argv[1]
payload = sys.argv[2]
try:
    data = json.loads(payload)
except json.JSONDecodeError as exc:
    raise SystemExit(f"failed to decode JSON response while extracting '{field}': {exc}; payload={payload!r}")
cur = data
for part in field.split("."):
    cur = cur[part]
if isinstance(cur, (dict, list)):
    print(json.dumps(cur))
else:
    print(cur)
PY
}

api_post_json() {
  local path="$1"
  local payload="$2"
  curl -fsS -X POST "http://127.0.0.1:${WASMFORGE_ADMIN_PORT}/api${path}" \
    -H "Content-Type: application/json" \
    -d "$payload"
}

api_patch_json() {
  local path="$1"
  local payload="$2"
  curl -fsS -X PATCH "http://127.0.0.1:${WASMFORGE_ADMIN_PORT}/api${path}" \
    -H "Content-Type: application/json" \
    -d "$payload"
}

vegeta_run() {
  local mode="$1"
  local url="$2"
  local attack_bin="$RESULT_DIR/${mode}.bin"
  local report_json="$RESULT_DIR/${mode}.report.json"
  local report_txt="$RESULT_DIR/${mode}.report.txt"

  echo "Running mode=$mode url=$url rate=$RATE duration=$DURATION"
  echo "GET $url" | vegeta attack \
    -rate="$RATE" \
    -duration="$DURATION" \
    -header "${AUTH_HEADER_NAME}: ${AUTH_HEADER_VALUE}" \
    >"$attack_bin"

  vegeta report "$attack_bin" >"$report_txt"
  vegeta report -type=json "$attack_bin" >"$report_json"
}

echo "Starting upstream service on :$UPSTREAM_PORT ..."
"$UPSTREAM_BIN" \
  --listen ":${UPSTREAM_PORT}" \
  --path "$BENCH_PATH" \
  >"$TMP_DIR/upstream.log" 2>&1 &
UPSTREAM_PID=$!
wait_http_ok "http://127.0.0.1:${UPSTREAM_PORT}/health"

echo "Starting WasmForge admin on :$WASMFORGE_ADMIN_PORT ..."
"$WASMFORGE_BIN" \
  --admin-port "$WASMFORGE_ADMIN_PORT" \
  --plugins-uploads-dir "$UPLOADS_DIR" \
  --logs-dir "$LOGS_DIR" \
  >"$TMP_DIR/wasmforge.log" 2>&1 &
WASMFORGE_PID=$!
wait_http_ok "http://127.0.0.1:${WASMFORGE_ADMIN_PORT}/api/health"

echo "Configuring WasmForge proxy listen port :$WASMFORGE_PROXY_PORT ..."
api_patch_json "/proxy/config" "{\"listen_port\":${WASMFORGE_PROXY_PORT},\"read_header_timeout\":5}" >/dev/null
api_post_json "/proxy/server/start" "{}" >/dev/null

route_payload="$(cat <<JSON
{
  "path":"${BENCH_PATH}",
  "target_url":"http://127.0.0.1:${UPSTREAM_PORT}${BENCH_PATH}",
  "idle_conn_timeout":30,
  "tls_handshake_timeout":5,
  "expect_continue_timeout":1,
  "response_header_timeout":5
}
JSON
)"

route_resp="$(api_post_json "/routes" "$route_payload")"
ROUTE_ID="$(extract_json_field route.id "$route_resp")"
api_post_json "/routes/${ROUTE_ID}/enable" "{}" >/dev/null

# 1) Baseline
wait_http_ok "http://127.0.0.1:${WASMFORGE_PROXY_PORT}${BENCH_PATH}"
vegeta_run "baseline" "http://127.0.0.1:${WASMFORGE_PROXY_PORT}${BENCH_PATH}"

# 2) WASM plugin mode
metadata="$(
  python3 - "$PLUGIN_NAME" "$PLUGIN_FILENAME" <<'PY'
import json
import sys

name = sys.argv[1]
filename = sys.argv[2]
print(json.dumps({"name": name, "version": "0.1.0", "filename": filename}))
PY
)"
plugin_resp="$(curl -fsS -X POST "http://127.0.0.1:${WASMFORGE_ADMIN_PORT}/api/plugins" \
  -F "wasm_file=@${PLUGIN_WASM_FILE}" \
  -F "metadata=${metadata}")"
PLUGIN_ID="$(extract_json_field plugin.id "$plugin_resp")"

route_plugin_payload="$(cat <<JSON
{
  "route_id":"${ROUTE_ID}",
  "plugin_id":"${PLUGIN_ID}",
  "version_constraint":"*",
  "execution_order":10,
  "config":"{}"
}
JSON
)"
api_post_json "/route-plugins" "$route_plugin_payload" >/dev/null
vegeta_run "wasm" "http://127.0.0.1:${WASMFORGE_PROXY_PORT}${BENCH_PATH}"

# 3) Native gateway mode
echo "Starting native gateway on :$NATIVE_GATEWAY_PORT ..."
"$NATIVE_BIN" \
  --listen ":${NATIVE_GATEWAY_PORT}" \
  --upstream "http://127.0.0.1:${UPSTREAM_PORT}${BENCH_PATH}" \
  --path "$BENCH_PATH" \
  --auth-header-name "$AUTH_HEADER_NAME" \
  --auth-header-value "$AUTH_HEADER_VALUE" \
  >"$TMP_DIR/native.log" 2>&1 &
NATIVE_PID=$!
wait_http_ok "http://127.0.0.1:${NATIVE_GATEWAY_PORT}/health"
vegeta_run "native" "http://127.0.0.1:${NATIVE_GATEWAY_PORT}${BENCH_PATH}"

python3 "$ROOT_DIR/bench/scripts/compare.py" "$RESULT_DIR" >"$RESULT_DIR/summary.md"

echo
echo "Benchmark complete."
echo "Results: $RESULT_DIR"
echo "Summary: $RESULT_DIR/summary.md"
