# Gork

[![Go Report Card](https://goreportcard.com/badge/gork)](https://goreportcard.com/report/gork)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://golang.org)

**Gork** is a high-performance, `wrk`-inspired HTTP/HTTPS API benchmarking and load generation tool written in Go. Designed for maximum throughput, sub-millisecond precision, and minimal memory overhead, Gork enables developers and SREs to stress-test APIs, measure latency distributions, and evaluate server performance under intense concurrent load.

---

## ⚡ Highlights & Key Features

- **🚀 Ultra-High Throughput & Keep-Alive Pooling**: Fully tuned HTTP/1.1 and HTTP/2 transport with automated response body draining and connection reuse, preventing socket exhaustion and TLS renegotiation bottlenecks.
- **📈 Ramping & Multi-Stage Profiles**: Smoothly ramp concurrency up and down (`--ramp-up`, `--ramp-down`) to eliminate connection stampedes, or define complex multi-stage test plans (`--stages "10s:10->50, 30s:50, 10s:50->0"`).
- **🚨 Breakpoint / Saturation Testing**: Automatically increment concurrency step-by-step (`--step-load`) until latency or error rate thresholds are breached, pinpointing the exact server breaking point.
- **⏱️ Detailed HTTP Latency Breakdown (`httptrace`)**: Isolates where latency is spent: **DNS Lookup**, **TCP Dial**, **TLS Handshake**, **TTFB (Time to First Byte / Server Processing)**, and **Content Download**.
- **🔌 Connection Health & Reuse Tracking**: Measures **Connection Reuse Rate (%)**, reused sockets, and newly dialed sockets to verify Keep-Alive efficiency.
- **📊 Constant $O(1)$ Memory HDR Histogram**: Online logarithmic latency histogram binning (1 µs to 100 s with ~1.9% resolution) calculates high-precision percentiles (P50, P75, P90, P95, P99, P99.9) and Standard Deviation (`StdDev`) without allocating millions of in-memory structs or triggering GC pauses.
- **🔄 Bidirectional Throughput**: Measures both **Upload Rate (Bytes Sent)** and **Download Rate (Bytes Read)**, alongside payload size distributions (Min, Avg, Max).
- **🔒 Lock-Free Per-Worker Aggregation**: Each worker thread records metrics into an independent collector and merges results in $O(1)$ time upon completion, eliminating lock contention.
- **🛡️ Graceful Interrupt Handling**: Traps `SIGINT`/`SIGTERM` (`Ctrl+C`) to terminate workers cleanly and output accurate partial benchmark statistics.
- **🎯 QPS Rate Limiting & Warmup**: Control load with target requests-per-second pacing (`-r` / `--rate`) and eliminate cold-start noise using warmup periods (`-w` / `--warmup`).
- **📑 Dual Output Formats**: Human-friendly terminal table and machine-readable JSON format with sub-millisecond floating-point precision for CI/CD integration.
- **🔑 Insecure TLS & Custom Hosts**: Benchmark local and staging endpoints with self-signed certificates (`-k` / `--insecure`) and custom `Host` headers.

---

## 📦 Installation

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

## 🚀 Quick Start

### 1. Basic Benchmark
Run a 10-second benchmark against an endpoint using 2 worker threads and 50 concurrent connections:

```bash
gork -c 50 -t 2 -d 10s https://api.example.com/health
```

### 2. Smooth Concurrency Ramping (Stress / Soak Testing)
Ramp up from 1 to 50 connections over 10s, sustain for 30s, and ramp down over 5s:

```bash
gork -c 50 --ramp-up 10s -d 30s --ramp-down 5s https://api.example.com/items
```

### 3. Custom Multi-Stage Load Profile
Define complex traffic stages with linear ramps and steady-state holds:

```bash
gork --stages "10s:10->50, 30s:50, 20s:50->150, 1m:150, 10s:150->0" https://api.example.com
```

### 4. Automated Breakpoint / Saturation Testing
Automatically step up load (+10 connections every 5s) until P95 latency exceeds 200ms or error rate exceeds 1%:

```bash
gork --step-load --step-conns 10 --step-duration 5s --max-latency 200ms --max-error-rate 1.0 https://api.example.com
```

### 5. POST Request with JSON Payload
Benchmark an API creation endpoint with custom headers and inline JSON body:

```bash
gork -c 100 -d 30s -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer my-secret-token" \
  -b '{"action":"ping","timestamp":1700000000}' \
  https://api.example.com/v1/events
```

### 6. Using a Request Body File
Load payloads directly from an external file:

```bash
gork -c 100 -d 30s -X POST \
  -H "Content-Type: application/json" \
  --body-file ./payload.json \
  https://api.example.com/v1/data
```

### 7. Rate-Paced Benchmark with Warmup
Warm up connections for 3 seconds, then benchmark at a target rate of 1,000 requests/sec for 1 minute:

```bash
gork -c 50 -d 1m -w 3s -r 1000 https://api.example.com/search
```

### 8. Insecure TLS (Self-Signed Certificates)
Benchmark local HTTPS development servers without certificate validation errors:

```bash
gork -k -c 20 -d 15s https://localhost:8443/test
```

### 9. JSON Export for CI/CD Pipelines
Output machine-readable JSON metrics directly to a file:

```bash
gork -c 50 -d 20s --json -o benchmark_results.json https://api.example.com
```

---

## ⚙️ CLI Reference

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
| `--ramp-up` | | `duration` | `0s` | Ramp-up concurrency from 1 to connections over duration |
| `--ramp-down` | | `duration` | `0s` | Ramp-down concurrency from connections to 1 over duration |
| `--stages` | | `string` | `""` | Multi-stage execution profile, e.g. `'10s:10->50, 30s:50, 10s:50->0'` |
| `--step-load` | | `bool` | `false` | Enable automated step-load saturation / breakpoint testing |
| `--step-conns` | | `int` | `10` | Concurrency increment per step in step-load testing |
| `--step-duration` | | `duration` | `5s` | Duration to hold each step in step-load testing |
| `--max-latency` | | `duration` | `500ms` | Saturation stop threshold: max acceptable P95 latency |
| `--max-error-rate` | | `float` | `1.0` | Saturation stop threshold: max acceptable error rate % (1.0 = 1%) |
| `--max-conns` | | `int` | `1000` | Safety ceiling concurrency for step load |
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

## 📊 Output Formats

### Terminal Output
```
Running 6s test @ https://api.example.com/v1/status
  2 thread(s) and 5 connection(s)

Stage Progression:
  Stage                        Duration   Concurrency  Requests   RPS          Avg Latency  P95 Latency  Errors
  ----------------------------------------------------------------------------------------------------------
  Stage 1 (Ramp Up 1->5)       2s         1->5         91         45.49        63.94ms      74.99ms      0
  Stage 2 (Sustained 5)        2s         5            159        79.48        61.24ms      68.13ms      0
  Stage 3 (Ramp Down 5->1)     2s         5->1         95         47.25        63.88ms      76.44ms      0

Latency Statistics:
  Min:       51.38ms
  Avg:       62.68ms
  Max:       260.42ms
  StdDev:    12.15ms

Latency Percentiles:
  50% (P50):  59.57ms
  75% (P75):  61.90ms
  90% (P90):  68.13ms
  95% (P95):  73.56ms
  99% (P99):  82.54ms
  99.9%:     256.06ms

HTTP Lifecycle Phase Breakdown:
  Phase                      Avg            Min            Max
  ----------------------------------------------------------------
  DNS Lookup:                16.68ms        16.68ms        16.68ms
  TCP Connect:               63.71ms        63.71ms        63.71ms
  TLS Handshake:             123.67ms       123.67ms       123.67ms
  Server Processing (TTFB):  62.40ms        51.38ms        260.42ms
  Content Download:          1.53ms         324.10µs       7.22ms

Connection & Keep-Alive Health:
  Reuse Rate:    99.71%
  Reused Sockets:344
  New Sockets:   1
  Total Attempts:345

Data Transfer & Throughput:
  Requests/sec:  57.39
  Download Rate: 31.33 KB/s (Total: 188.33 KB)
  Upload Rate:   918 B/s (Total: 5.39 KB)
  Total Traffic: 193.73 KB
  Payload Size:  Min: 559 B | Avg: 559 B | Max: 559 B

Requests & Availability:
  Total:         345
  Success Rate:  100.00% (345 successful)
  Error Rate:    0.00% (0 failed)
  Status Classes:2xx: 345 | 3xx: 0 | 4xx: 0 | 5xx: 0

Status Codes:
  [200 OK]: 345
```

### Breakpoint Detection Banner Example
```
================================================================================
🚨 BREAKPOINT SATURATION DETECTED
  Breaking Concurrency: 80 connections
  Throughput at Peak:   4,512.45 req/sec
  P95 Latency at Peak:  245.20ms
  Error Rate at Peak:   1.80%
  Trigger Reason:       P95 latency of 245.20ms exceeded threshold of 200ms
================================================================================
```

### JSON Output
```json
{
  "total_requests": 345,
  "successful": 345,
  "failed": 0,
  "duration_seconds": 6.0001,
  "requests_per_sec": 57.39,
  "latency_ms": {
    "min": 51.38,
    "avg": 62.68,
    "max": 260.42,
    "stddev": 12.15,
    "p50": 59.57,
    "p75": 61.90,
    "p90": 68.13,
    "p95": 73.56,
    "p99": 82.54,
    "p99_9": 256.06
  },
  "phases_ms": {
    "dns": { "min_ms": 16.68, "avg_ms": 16.68, "max_ms": 16.68 },
    "tcp": { "min_ms": 63.71, "avg_ms": 63.71, "max_ms": 63.71 },
    "tls": { "min_ms": 123.67, "avg_ms": 123.67, "max_ms": 123.67 },
    "ttfb": { "min_ms": 51.38, "avg_ms": 62.40, "max_ms": 260.42 },
    "content": { "min_ms": 0.32, "avg_ms": 1.53, "max_ms": 7.22 }
  },
  "connections": {
    "total_attempts": 345,
    "reused": 344,
    "new": 1,
    "reused_pct": 99.71
  },
  "transfer": {
    "bytes_sent": 5520,
    "bytes_read": 192855,
    "total_bytes": 198375,
    "upload_bytes_per_sec": 918.0,
    "download_bytes_per_sec": 32080.0,
    "upload_rate": "918 B/s",
    "download_rate": "31.33 KB/s",
    "min_body_bytes": 559,
    "avg_body_bytes": 559,
    "max_body_bytes": 559
  },
  "availability": {
    "success_rate_pct": 100,
    "error_rate_pct": 0,
    "count_2xx": 345,
    "count_3xx": 0,
    "count_4xx": 0,
    "count_5xx": 0
  },
  "stages": [
    {
      "name": "Stage 1 (Ramp Up 1->5)",
      "duration": 2000000000,
      "concurrency": "1->5",
      "requests": 91,
      "rps": 45.49,
      "avg_latency": 63940000,
      "p95_latency": 74990000,
      "errors": 0
    },
    {
      "name": "Stage 2 (Sustained 5)",
      "duration": 2000000000,
      "concurrency": "5",
      "requests": 159,
      "rps": 79.48,
      "avg_latency": 61240000,
      "p95_latency": 68130000,
      "errors": 0
    },
    {
      "name": "Stage 3 (Ramp Down 5->1)",
      "duration": 2000000000,
      "concurrency": "5->1",
      "requests": 95,
      "rps": 47.25,
      "avg_latency": 63880000,
      "p95_latency": 76440000,
      "errors": 0
    }
  ],
  "status_codes": {
    "200": 345
  }
}
```

---

## 🏗️ Architecture & Internal Design

```
┌────────────────────────────────────────────────────────┐
│                        CLI / Main                      │
│     (Flags Parsing, Signal Trapping, Validation)       │
└───────────────────────────┬────────────────────────────┘
                            │
              ┌─────────────▼─────────────┐
              │      Benchmark Runner     │
              │ (Stages, Step-Load, GOMAX)│
              └─────────────┬─────────────┘
                            │ Dynamic Worker Scaling
          ┌─────────────────┼─────────────────┐
          ▼                 ▼                 ▼
   ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
   │   Worker 1   │  │   Worker 2   │  │   Worker N   │
   │  httptrace   │  │  httptrace   │  │  httptrace   │
   │  Phase Stats │  │  Phase Stats │  │  Phase Stats │
   └──────┬───────┘  └──────┬───────┘  └──────┬───────┘
          │                 │                 │
          └─────────────────┼─────────────────┘
                            │ Merge O(1)
              ┌─────────────▼─────────────┐
              │     Summary Generator     │
              │(Phases, Transfer, Avail)  │
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
4. **Dynamic Stage & Breakpoint Scaling**:
   Workers gate requests using atomic target concurrency counters, allowing smooth linear ramping transitions between concurrency levels without spawning and killing thousands of goroutines.

---

## 🧪 Development & Testing

Run all unit and integration tests:

```bash
go test -v ./internal/...
```

Run benchmarks on the metrics aggregation engine:

```bash
go test -bench=. -benchmem ./internal/metrics
```

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).

