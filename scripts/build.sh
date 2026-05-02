#!/usr/bin/env bash
#
# Copyright (c) 2026. Mikhail Kulik.
#
# Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
UI_DIR="$ROOT_DIR/ui/adminv2"
UI_BUILD_DIR="$UI_DIR/apps/web"
EMBED_DIR="$ROOT_DIR/pkg/ui/out"
UI_OUT_DIR="$UI_DIR/apps/web/out"
BIN_DIR="$ROOT_DIR/bin"
BIN_PATH="$BIN_DIR/wasmforge"

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Error: '$1' command is required but not found in PATH."
    exit 1
  }
}

require_cmd npm
require_cmd npx
require_cmd go

echo "==> Building Admin UI"
pushd "$UI_DIR" > /dev/null
npm install
popd > /dev/null

pushd "$UI_BUILD_DIR" > /dev/null
npx next build
popd > /dev/null

echo "==> Preparing embedded UI output"
rm -rf "$EMBED_DIR"
mkdir -p "$(dirname "$EMBED_DIR")"

if command -v rsync >/dev/null 2>&1; then
  rsync -a --delete "$UI_OUT_DIR"/ "$EMBED_DIR"/
else
  mkdir -p "$EMBED_DIR"
  cp -R "$UI_OUT_DIR"/. "$EMBED_DIR"/
fi

echo "==> Building Go binary"
mkdir -p "$BIN_DIR"
go build -o "$BIN_PATH" "$ROOT_DIR/cmd/gateway/main.go"

echo "Build completed successfully. Binary available at: $BIN_PATH"
