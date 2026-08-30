# Gork

[![Go Report Card](https://goreportcard.com/badge/gork)](https://goreportcard.com/report/gork)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://golang.org)

**Gork** is a high-performance, `wrk`-inspired HTTP/HTTPS API benchmarking and load generation tool written in Go. Designed for maximum throughput, sub-millisecond precision, and minimal memory overhead, Gork enables developers and SREs to stress-test APIs, measure latency distributions, and evaluate server performance under intense concurrent load.

---

##  Highlights & Key Features

- **Ultra-High Throughput & Keep-Alive Pooling**: Fully tuned HTTP/1.1 and HTTP/2 transport with automated response body draining and connection reuse, preventing socket exhaustion and TLS renegotiation bottlenecks.
- **Constant $O(1)$ Memory HDR Histogram**: Online logarithmic latency histogram binning (1 µs to 100 s with ~1.9% resolution) calculates high-precision percentiles (P50, P75, P90, P95, P99, P99.9) and Standard Deviation (`StdDev`) without allocating millions of in-memory structs or triggering GC pauses.
- **Lock-Free Per-Worker Aggregation**: Each worker thread records metrics into an independent collector and merges results in $O(1)$ time upon completion, eliminating lock contention.
- **Pre-Parsed Zero-Allocation Request Templates**: Headers and body reader templates are pre-parsed once at startup, eliminating string splitting and repetitive allocations inside the hot request loop.
- **Graceful Interrupt Handling**: Traps `SIGINT`/`SIGTERM` (`Ctrl+C`) to terminate workers cleanly and output accurate partial benchmark statistics.
- **QPS Rate Limiting & Warmup**: Control load with target requests-per-second pacing (`-r` / `--rate`) and eliminate cold-start noise using warmup periods (`-w` / `--warmup`).
- **Dual Output Formats**: Human-friendly terminal table and machine-readable JSON format with sub-millisecond floating-point precision for CI/CD integration.
- **Insecure TLS & Custom Hosts**: Benchmark local and staging endpoints with self-signed certificates (`-k` / `--insecure`) and custom `Host` headers.

---

##  Installation

### From Source (Go 1.23+)

```bash
git clone https://github.com/WebProject05/Gork
cd gork
go build -o gork cmd/gork/main.go
```

Or using the Makefile:

```bash
make build
```

### Install with `go install`

```bash
go install ./cmd/gork
```

---

##  Quick Start

### 1. Basic Benchmark
Run a 10-second benchmark against an endpoint using 2 worker threads and 50 concurrent connections:

```bash
gork -c 50 -t 2 -d 10s https://api.example.com/health
```

### 2. POST Request with JSON Payload
Benchmark an API creation endpoint with custom headers and inline JSON body:

```bash
gork -c 100 -d 30s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer my-secret-token" \
  -b '{"action":"ping","timestamp":1700000000}' \
  https://api.example.com/v1/events
```

### 3. Using a Request Body File
Load payloads directly from an external file:

```bash
gork -c 100 -d 30s -X POST \
  -H "Content-Type: application/json" \
  --body-file ./payload.json \
  https://api.example.com/v1/data
```

### 4. Rate-Paced Benchmark with Warmup
Warm up connections for 3 seconds, then benchmark at a target rate of 1,000 requests/sec for 1 minute:

```bash
gork -c 50 -d 1m -w 3s -r 1000 https://api.example.com/search
```

### 5. Insecure TLS (Self-Signed Certificates)
Benchmark local HTTPS development servers without certificate validation errors:

```bash
gork -k -c 20 -d 15s https://localhost:8443/test
```

### 6. JSON Export for CI/CD Pipelines
Output machine-readable JSON metrics directly to a file:

```bash
gork -c 50 -d 20s --json -o benchmark_results.json https://api.example.com
```

---

##  CLI Reference

```
Usage:
  gork [options] <url>
```

| Flag | Short | Type | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--connections` | `-c` | `int` | `10` | Number of concurrent HTTP connections |
| `--threads` | `-t` | `int` | `2` | Number of worker threads / goroutine dispatchers |
| `--duration` | `-d` | `duration` | `10s` | Total benchmark duration (e.g. `10s`, `1m`, `30s`) |
| `--warmup` | `-w` | `duration` | `0s` | Warmup duration before recording metrics |
| `--rate` | `-r` | `int` | `0` | Target rate limit in requests per second (`0` = unlimited) |
| `--method` | `-X` | `string` | `"GET"` | HTTP method (`GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `HEAD`) |
| `--header` | `-H` | `string` | `[]` | Custom HTTP header in `'Key: Value'` format (repeatable) |
| `--body` | `-b` | `string` | `""` | Inline request payload string |
| `--body-file` | | `string` | `""` | Path to file containing request payload |
| `--timeout` | | `duration` | `5s` | Per-request timeout |
| `--insecure` | `-k` | `bool` | `false` | Skip TLS certificate verification |
| `--no-keepalive` | | `bool` | `false` | Disable HTTP Keep-Alive connection reuse |
| `--cpus` | | `int` | `0` | Explicit `GOMAXPROCS` override (default: number of CPUs) |
| `--json` | | `bool` | `false` | Output results in machine-readable JSON format |
| `--out` | `-o` | `string` | `""` | File path to write benchmark results |
| `--version` | `-v` | `bool` | `false` | Print version and exit |
| `--help` | `-h` | `bool` | `false` | Show CLI help message |

---

##  Output Formats

### Terminal Output
```
Running 10s test @ https://api.example.com/v1/status
  4 thread(s) and 100 connection(s)

Latency Statistics:
  Min:       240.50µs
  Avg:       1.85ms
  Max:       18.42ms
  StdDev:    940.12µs

Latency Percentiles:
  50% (P50):  1.52ms
  75% (P75):  2.10ms
  90% (P90):  2.88ms
  95% (P95):  3.64ms
  99% (P99):  5.92ms
  99.9%:     12.15ms

Throughput:
  Requests/sec:  54120.35
  Transfer/sec:  28.45 MB/s
  Total Read:    284.50 MB

Requests:
  Total:         541204
  Successful:    541204
  Failed:        0

Status Codes:
  [200 OK]: 541204
```

### JSON Output
```json
{
  "total_requests": 541204,
  "successful": 541204,
  "failed": 0,
  "duration_seconds": 10.00001,
  "requests_per_sec": 54120.35,
  "bytes_read": 298319808,
  "bytes_per_sec": 29831950.96,
  "transfer_rate": "28.45 MB/s",
  "latency_ms": {
    "min": 0.2405,
    "avg": 1.85,
    "max": 18.42,
    "stddev": 0.94012,
    "p50": 1.52,
    "p75": 2.10,
    "p90": 2.88,
    "p95": 3.64,
    "p99": 5.92,
    "p99_9": 12.15
  },
  "latency_readable": {
    "min": "240.50µs",
    "avg": "1.85ms",
    "max": "18.42ms",
    "stddev": "940.12µs",
    "p50": "1.52ms",
    "p75": "2.10ms",
    "p90": "2.88ms",
    "p95": "3.64ms",
    "p99": "5.92ms",
    "p99_9": "12.15ms"
  },
  "status_codes": {
    "200": 541204
  }
}
```

---

##  Architecture & Internal Design

```
┌────────────────────────────────────────────────────────┐
│                        CLI / Main                      │
│     (Flags Parsing, Signal Trapping, Validation)       │
└───────────────────────────┬────────────────────────────┘
                            │
              ┌─────────────▼─────────────┐
              │      Benchmark Runner     │
              │  (GOMAXPROCS, Warmup, Wg) │
              └─────────────┬─────────────┘
                            │ Spawns
          ┌─────────────────┼─────────────────┐
          ▼                 ▼                 ▼
   ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
   │   Worker 1   │  │   Worker 2   │  │   Worker N   │
   │  Keep-Alive  │  │  Keep-Alive  │  │  Keep-Alive  │
   │  Collector   │  │  Collector   │  │  Collector   │
   └──────┬───────┘  └──────┬───────┘  └──────┬───────┘
          │                 │                 │
          └─────────────────┼─────────────────┘
                            │ Merge O(1)
              ┌─────────────▼─────────────┐
              │     Summary Generator     │
              │ (HDR Percentiles, StdDev) │
              └─────────────┬─────────────┘
                            │
              ┌─────────────▼─────────────┐
              │    Terminal / JSON Output │
              └───────────────────────────┘
```

1. **Keep-Alive Connection Lifecycle**:
   In standard Go HTTP clients, failing to read the response body to `EOF` causes `net/http` to close the underlying TCP connection. Gork automatically drains all response bodies into `io.Discard`, maintaining open TCP keep-alive connections and enabling realistic microservice load simulation.
2. **Logarithmic Online Histogram**:
   Traditional tools often store individual latency samples in a dynamic slice and sort them at the end. For a 100k RPS benchmark, this consumes gigabytes of RAM and triggers major GC pauses. Gork utilizes 1,000 logarithmic buckets spanning 1 µs to 100 s, guaranteeing $O(1)$ memory usage (~8 KB per worker) and instant percentile calculations.
3. **Lock-Free Concurrency**:
   Workers maintain isolated metric counters without acquiring mutexes on individual HTTP responses. Upon test completion, collectors are merged into the main aggregator via atomic addition.

---

##  Development & Testing

Run all unit and integration tests:

```bash
go test -v ./internal/...
```

Run benchmarks on the metrics aggregation engine:

```bash
go test -bench=. -benchmem ./internal/metrics
```

---

##  License

This project is licensed under the [MIT License](LICENSE).
