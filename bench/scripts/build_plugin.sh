#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PLUGIN_DIR="$ROOT_DIR/bench/plugins/header-auth-rust"
OUT_DIR="$ROOT_DIR/bench/artifacts"
OUT_FILE="$OUT_DIR/header_auth_bench.wasm"

if ! command -v cargo >/dev/null 2>&1; then
  echo "error: cargo is required to build benchmark plugin" >&2
  exit 1
fi

rustup target add wasm32-unknown-unknown >/dev/null 2>&1 || true

mkdir -p "$OUT_DIR"

echo "Building benchmark WASM plugin..."
cargo build --manifest-path "$PLUGIN_DIR/Cargo.toml" --release --target wasm32-unknown-unknown

cp "$PLUGIN_DIR/target/wasm32-unknown-unknown/release/header_auth_bench.wasm" "$OUT_FILE"
echo "Plugin built: $OUT_FILE"
