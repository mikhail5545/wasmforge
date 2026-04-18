# WasmForge

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/license-Apache_2.0-green)](https://www.apache.org/licenses/LICENSE-2.0)
[![WASM](https://img.shields.io/badge/WASM-Powered-654FF0?style=flat&logo=webassembly)](https://webassembly.org/)
[![Next.js](https://img.shields.io/badge/Frontend-Next.js-black?style=flat&logo=next.js)](https://nextjs.org/)

**WasmForge** is an API Gateway written in Go that lets you attach **WASM plugins** to routes for auth, rate limits, request/response transforms, and other middleware logic.

Build plugins in your language of choice, compile to WASM, and manage everything from the embedded admin UI.

## Why WasmForge?

- **Fast data plane** built on `net/http`
- **Extensible middleware** with sandboxed WASM modules
- **Route-level control** with plugin ordering and config
- **Dynamic management** via REST API + embedded Next.js dashboard
- **Persistent state** with SQLite + GORM

## Architecture at a glance

WasmForge is split into:

1. **Data plane (proxy):** Handles incoming traffic and executes route middleware chains.
2. **Control plane (admin API + UI):** Manages routes, plugins, and runtime configuration.

This separation keeps request handling efficient while making configuration simple and safe.

## Installation and quick start

### Prerequisites

- Go 1.25+
- Node.js 18+
- `make`

### 1. Clone

```bash
git clone https://github.com/mikhail5545/wasmforge.git
cd wasmforge
```

### 2. Build

This builds the admin UI, embeds it into the Go binary, and compiles the gateway.

```bash
make build
```

### 3. Run

```bash
./bin/wasmforge
```

Default ports:

- **Proxy:** `:8000`
- **Admin UI/API:** `:8080`

### 4. Open the dashboard

Go to:

```text
http://localhost:8080
```

From there, you can create routes, upload WASM plugins, and attach plugins to routes.

## Benchmarking

For repeatable performance benchmarks (baseline vs WASM plugin vs native middleware), use the toolkit in [`bench/`](./bench/README.md).

## Project status

WasmForge is actively evolving. See:

- [CONTRIBUTING.md](./CONTRIBUTING.md) for contribution workflow
- [ROADMAP.md](./ROADMAP.md) for planned improvements
- [WEBSITE_CONTENT.md](./WEBSITE_CONTENT.md) for website copy draft
