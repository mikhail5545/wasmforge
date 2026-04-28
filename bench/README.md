# Benchmark Toolkit

This folder gives you a repeatable benchmark workflow to compare:

1. **Baseline**: WasmForge route without plugin
2. **WASM**: Same route with a WASM header-auth plugin
3. **Native**: A Go native middleware gateway with equivalent auth logic

## What is included

- `cmd/upstream/` – simple upstream service used by both gateways
- `cmd/native-gateway/` – comparison gateway with native auth middleware
- `plugins/header-auth-rust/` – sample WASM plugin source
- `scripts/build_plugin.sh` – builds the sample plugin to `.wasm`
- `scripts/run_suite.sh` – runs full benchmark suite and stores results
- `scripts/compare.py` – generates a markdown summary from result JSON
- `scenarios/header-auth.env` – default benchmark scenario config

## Prerequisites

- Go 1.25+
- Rust toolchain (for sample plugin build), target `wasm32-unknown-unknown`
- `curl`
- `python3`
- `vegeta` load tool

Install vegeta (example):

```bash
go install github.com/tsenart/vegeta/v12@latest
```

Make sure `$GOPATH/bin` is in your `PATH`.

## Quick start

1. Build WasmForge binary:

```bash
make build
```

2. Build benchmark helper binaries:

```bash
make bench-build
```

3. Build sample WASM plugin:

```bash
./bench/scripts/build_plugin.sh
```

4. Run benchmark suite:

```bash
make bench-run
```

Results are written under:

```text
bench/results/<timestamp>/
```

## Result files

- `baseline.report.json`
- `wasm.report.json`
- `native.report.json`
- `summary.md`

`summary.md` includes p50/p95/p99 latency, throughput, success rate, and overhead vs baseline.

## Scenario tuning

Edit:

```text
bench/scenarios/header-auth.env
```

You can tune:

- rate
- duration
- request path
- benchmark header key/value
- listen ports for upstream, WasmForge proxy/admin, native gateway
