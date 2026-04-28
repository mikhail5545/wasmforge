#!/usr/bin/env python3
import json
import pathlib
import sys


def ns_to_ms(value: float) -> float:
    return value / 1_000_000.0


def load_report(path: pathlib.Path):
    with path.open("r", encoding="utf-8") as f:
        data = json.load(f)
    return {
        "throughput": float(data.get("throughput", 0.0)),
        "success": float(data.get("success", 0.0)),
        "lat_p50_ms": ns_to_ms(float(data["latencies"]["50th"])),
        "lat_p95_ms": ns_to_ms(float(data["latencies"]["95th"])),
        "lat_p99_ms": ns_to_ms(float(data["latencies"]["99th"])),
    }


def pct_delta(current: float, baseline: float) -> float:
    if baseline == 0:
        return 0.0
    return ((current - baseline) / baseline) * 100.0


def fmt(num: float, digits=2):
    return f"{num:.{digits}f}"


def main():
    if len(sys.argv) != 2:
        print("usage: compare.py <result_dir>", file=sys.stderr)
        sys.exit(1)

    result_dir = pathlib.Path(sys.argv[1])
    baseline = load_report(result_dir / "baseline.report.json")
    wasm = load_report(result_dir / "wasm.report.json")
    native = load_report(result_dir / "native.report.json")

    rows = [
        ("baseline", baseline),
        ("wasm", wasm),
        ("native", native),
    ]

    print("# Benchmark Summary")
    print()
    print("| mode | throughput (req/s) | success (%) | p50 (ms) | p95 (ms) | p99 (ms) |")
    print("| --- | ---: | ---: | ---: | ---: | ---: |")
    for mode, r in rows:
        print(
            f"| {mode} | {fmt(r['throughput'])} | {fmt(r['success']*100)} | {fmt(r['lat_p50_ms'])} | {fmt(r['lat_p95_ms'])} | {fmt(r['lat_p99_ms'])} |"
        )

    print()
    print("## Overhead vs baseline")
    print()
    print("| mode | throughput delta (%) | p95 delta (%) | p99 delta (%) |")
    print("| --- | ---: | ---: | ---: |")
    for mode, r in [("wasm", wasm), ("native", native)]:
        print(
            f"| {mode} | {fmt(pct_delta(r['throughput'], baseline['throughput']))} | {fmt(pct_delta(r['lat_p95_ms'], baseline['lat_p95_ms']))} | {fmt(pct_delta(r['lat_p99_ms'], baseline['lat_p99_ms']))} |"
        )


if __name__ == "__main__":
    main()
