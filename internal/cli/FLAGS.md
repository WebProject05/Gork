# Gork CLI Flags Reference & Usage Guide

This document provides a comprehensive, exhaustive reference for all command-line interface (CLI) flags, options, syntax rules, and real-world usage patterns supported by **Gork**.

---

## 📑 Table of Contents

1. [Synopsis](#synopsis)
2. [Quick Reference Table](#quick-reference-table)
3. [Flag Categories & Detailed Usage](#flag-categories--detailed-usage)
   - [1. Core Concurrency & Timing](#1-core-concurrency--timing)
   - [2. Ramping & Multi-Stage Profiles](#2-ramping--multi-stage-profiles)
   - [3. Breakpoint & Saturation Testing](#3-breakpoint--saturation-testing)
   - [4. HTTP Request & Transport Configuration](#4-http-request--transport-configuration)
   - [5. Reporting & Output Formatting](#5-reporting--output-formatting)
   - [6. Informational Flags](#6-informational-flags)
4. [Positional Arguments](#positional-arguments)
5. [Real-World Usage Recipes](#real-world-usage-recipes)
6. [Validation Rules & Mutual Exclusions](#validation-rules--mutual-exclusions)

---

## 📌 Synopsis

```text
gork [options] <url>
```

- Target `<url>` is passed as the final positional argument.
- Both short flags (`-c 50`) and GNU-style long flags (`--connections 50` or `--connections=50`) are supported.
- Boolean flags do not require an argument (e.g., `-k`, `--json`, `--step-load`).
- Repeatable flags (such as `-H` / `--header`) can be specified multiple times.

---

## 📊 Quick Reference Table

| Flag | Short | Type | Default | Category | Description |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `--connections` | `-c` | `int` | `10` | Core | Number of concurrent HTTP connections / sockets |
| `--threads` | `-t` | `int` | `2` | Core | Number of worker goroutine dispatcher threads |
| `--duration` | `-d` | `duration` | `10s` | Core | Total duration of benchmark run (e.g., `10s`, `1m`, `45s`) |
| `--warmup` | `-w` | `duration` | `0s` | Core | Warmup duration before metrics recording begins |
| `--rate` | `-r` | `int` | `0` | Core | Target global throughput limit in requests per second (`0` = unlimited) |
| `--cpus` | | `int` | `0` | Core | Override `GOMAXPROCS` (`0` = auto-detect runtime CPUs) |
| `--ramp-up` | | `duration` | `0s` | Ramping | Smoothly scale concurrency from 1 to target connections |
| `--ramp-down` | | `duration` | `0s` | Ramping | Smoothly scale concurrency from target connections down to 1 |
| `--stages` | | `string` | `""` | Ramping | Multi-stage load profile string (e.g., `'10s:10->50, 30s:50, 10s:50->0'`) |
| `--step-load` | | `bool` | `false` | Breakpoint | Enable automated step-load saturation testing |
| `--step-conns` | | `int` | `10` | Breakpoint | Concurrency increment per step in step-load testing |
| `--step-duration` | | `duration` | `5s` | Breakpoint | Duration to sustain each concurrency step |
| `--max-latency` | | `duration` | `500ms` | Breakpoint | Saturation threshold: maximum tolerable P95 latency |
| `--max-error-rate`| | `float` | `1.0` | Breakpoint | Saturation threshold: maximum tolerable error rate % (`1.0` = 1%) |
| `--max-conns` | | `int` | `1000` | Breakpoint | Safety ceiling concurrency for step-load testing |
| `--method` | `-X` | `string` | `"GET"` | HTTP | HTTP method (`GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `HEAD`) |
| `--header` | `-H` | `string` | `[]` | HTTP | Custom HTTP header `'Key: Value'` (repeatable) |
| `--body` | `-b` | `string` | `""` | HTTP | Inline request body string |
| `--body-file` | | `string` | `""` | HTTP | Path to file containing request payload |
| `--timeout` | | `duration` | `5s` | HTTP | Per-request round-trip timeout |
| `--insecure` | `-k` | `bool` | `false` | HTTP | Skip TLS certificate verification |
| `--no-keepalive` | | `bool` | `false` | HTTP | Disable HTTP Keep-Alive connection pooling |
| `--json` | | `bool` | `false` | Output | Output metrics in structured, machine-readable JSON format |
| `--out` | `-o` | `string` | `""` | Output | Write terminal or JSON output directly to file |
| `--version` | `-v` | `bool` | `false` | Info | Display binary version and exit |
| `--help` | `-h` | `bool` | `false` | Info | Print CLI usage guide and exit |

---

## 🔍 Flag Categories & Detailed Usage

### 1. Core Concurrency & Timing

#### `-c, --connections <int>`
- **Default**: `10`
- **Description**: Sets the maximum number of concurrent HTTP connections maintained against the target server.
- **Example**:
  ```bash
  gork -c 100 https://api.example.com
  ```

#### `-t, --threads <int>`
- **Default**: `2`
- **Description**: Number of worker dispatchers across which connections are divided. Connections are evenly partitioned among threads.
- **Example**:
  ```bash
  gork -c 200 -t 4 https://api.example.com
  ```

#### `-d, --duration <duration>`
- **Default**: `10s`
- **Supported Units**: `ms`, `s`, `m`, `h` (e.g., `500ms`, `30s`, `2m`, `1h`).
- **Description**: Total duration of the benchmark. When stages or ramping are configured, duration is automatically computed from the stage definitions.
- **Example**:
  ```bash
  gork -c 50 -d 45s https://api.example.com
  ```

#### `-w, --warmup <duration>`
- **Default**: `0s`
- **Description**: Runs an initial unrecorded benchmark phase for the specified duration. Allows JIT compilers, database connection pools, and TLS sessions to initialize so cold-start anomalies do not skew your benchmark metrics.
- **Example**:
  ```bash
  gork -c 50 -w 5s -d 30s https://api.example.com
  ```

#### `-r, --rate <int>`
- **Default**: `0` (unlimited / max throughput)
- **Description**: Global target rate limit in requests per second across all workers. Each worker paces itself using high-resolution timers.
- **Example**:
  ```bash
  # Limit throughput to exactly 2,500 requests/sec
  gork -c 50 -r 2500 -d 1m https://api.example.com
  ```

#### `--cpus <int>`
- **Default**: `0` (Go runtime CPU count)
- **Description**: Overrides Go's runtime OS thread allocation (`GOMAXPROCS`). Useful for benchmarking how software scales across specific core counts.
- **Example**:
  ```bash
  gork --cpus 4 -c 100 -d 30s https://api.example.com
  ```

---

### 2. Ramping & Multi-Stage Profiles

#### `--ramp-up <duration>`
- **Default**: `0s`
- **Description**: Automatically prepends a linear ramp-up stage from `1` connection to `-c` connections over the given duration before running the sustained benchmark.
- **Example**:
  ```bash
  # 5s ramp (1->50), 20s sustained at 50 conns
  gork -c 50 --ramp-up 5s -d 20s https://api.example.com
  ```

#### `--ramp-down <duration>`
- **Default**: `0s`
- **Description**: Automatically appends a linear ramp-down stage from `-c` connections down to `1` connection over the given duration after the sustained benchmark.
- **Example**:
  ```bash
  # 20s sustained at 50 conns, followed by 5s ramp-down (50->1)
  gork -c 50 -d 20s --ramp-down 5s https://api.example.com
  ```

#### Combined `--ramp-up` + `-d` + `--ramp-down`
- Automatically generates a 3-stage load curve (Ramp Up $\to$ Sustained $\to$ Ramp Down).
- **Example**:
  ```bash
  gork -c 100 --ramp-up 10s -d 30s --ramp-down 10s https://api.example.com
  ```

#### `--stages <spec>`
- **Default**: `""`
- **Description**: Defines a custom multi-stage execution plan using comma-separated stage expressions.
- **Grammar**:
  - Linear Ramp: `<duration>:<start_conns>-><target_conns>`
  - Steady Hold: `<duration>:<conns>`
- **Syntax Examples**:
  - Single ramp: `"10s:1->50"`
  - Steady hold: `"30s:50"`
  - Full profile: `"10s:10->50, 30s:50, 20s:50->150, 1m:150, 10s:150->0"`
- **Example**:
  ```bash
  gork --stages "10s:5->25, 20s:25, 10s:25->100, 30s:100, 10s:100->0" https://api.example.com
  ```

---

### 3. Breakpoint & Saturation Testing

#### `--step-load`
- **Default**: `false`
- **Description**: Switches Gork into automated saturation discovery mode. Instead of running a fixed-duration test, Gork continuously increases concurrency step-by-step until the server breaks (threshold tripped) or the safety ceiling is reached.
- **Example**:
  ```bash
  gork --step-load https://api.example.com
  ```

#### `--step-conns <int>`
- **Default**: `10`
- **Description**: Number of connections added at each step.
- **Example**:
  ```bash
  gork --step-load --step-conns 25 https://api.example.com
  ```

#### `--step-duration <duration>`
- **Default**: `5s`
- **Description**: Time held at each concurrency step before evaluating performance and escalating.
- **Example**:
  ```bash
  gork --step-load --step-duration 10s https://api.example.com
  ```

#### `--max-latency <duration>`
- **Default**: `500ms`
- **Description**: Saturation trigger threshold for latency. If the step's **P95 Latency** exceeds this value, the benchmark immediately halts and marks the breakpoint.
- **Example**:
  ```bash
  # Break when P95 latency crosses 150ms
  gork --step-load --max-latency 150ms https://api.example.com
  ```

#### `--max-error-rate <float>`
- **Default**: `1.0` (meaning 1.0%)
- **Description**: Saturation trigger threshold for errors. If the percentage of failed requests (`5xx`, timeouts, connection drops) exceeds this percentage, the benchmark immediately halts.
- **Example**:
  ```bash
  # Break when failure rate exceeds 2.5%
  gork --step-load --max-error-rate 2.5 https://api.example.com
  ```

#### `--max-conns <int>`
- **Default**: `1000`
- **Description**: Upper safety ceiling for concurrency during step-load testing to prevent overloading test harness infrastructure.
- **Example**:
  ```bash
  gork --step-load --max-conns 500 https://api.example.com
  ```

---

### 4. HTTP Request & Transport Configuration

#### `-X, --method <string>`
- **Default**: `"GET"`
- **Supported Methods**: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `HEAD`, `OPTIONS`
- **Description**: HTTP request method. Automatically converted to uppercase.
- **Example**:
  ```bash
  gork -X POST -b '{"name":"test"}' https://api.example.com/items
  ```

#### `-H, --header <string>`
- **Default**: `[]`
- **Format**: `'Header-Name: Header-Value'`
- **Repeatable**: Yes. Specify multiple `-H` flags for multiple headers.
- **Special Headers**:
  - Setting `-H 'Host: custom.domain.com'` automatically overrides the HTTP `Host` request header without modifying the target dial address.
- **Example**:
  ```bash
  gork -H "Content-Type: application/json" \
       -H "Authorization: Bearer token-12345" \
       -H "X-Trace-ID: gork-perf-test" \
       https://api.example.com
  ```

#### `-b, --body <string>`
- **Default**: `""`
- **Description**: Inline string payload for requests (`POST`, `PUT`, `PATCH`). Mutually exclusive with `--body-file`.
- **Example**:
  ```bash
  gork -X POST -H "Content-Type: application/json" -b '{"ping":true}' https://api.example.com/ping
  ```

#### `--body-file <path>`
- **Default**: `""`
- **Description**: Reads request payload from an external file on disk. Mutually exclusive with `-b`/`--body`.
- **Example**:
  ```bash
  gork -X POST -H "Content-Type: application/json" --body-file ./payload.json https://api.example.com/data
  ```

#### `--timeout <duration>`
- **Default**: `5s`
- **Description**: Maximum timeout for a single HTTP round trip (dial + TLS + send + TTFB + body download).
- **Example**:
  ```bash
  gork --timeout 2s https://api.example.com
  ```

#### `-k, --insecure`
- **Default**: `false`
- **Description**: Skips TLS certificate verification (`InsecureSkipVerify: true`). Essential when testing local microservices, staging environments, or self-signed internal certificates.
- **Example**:
  ```bash
  gork -k https://localhost:8443/api
  ```

#### `--no-keepalive`
- **Default**: `false`
- **Description**: Disables HTTP Keep-Alive connection pooling (`DisableKeepAlives: true`). Forces a fresh TCP handshake and TLS negotiation for every individual request. Useful for testing edge gateways under connection storms.
- **Example**:
  ```bash
  gork --no-keepalive -c 20 -d 10s https://api.example.com
  ```

---

### 5. Reporting & Output Formatting

#### `--json`
- **Default**: `false`
- **Description**: Emits benchmark results in structured JSON format with floating-point millisecond precision instead of the human-readable terminal table. Ideal for CI/CD assertions and Grafana/Datadog metric ingestion.
- **Example**:
  ```bash
  gork -c 50 -d 10s --json https://api.example.com
  ```

#### `-o, --out <path>`
- **Default**: `""` (stdout)
- **Description**: Saves the benchmark report directly to a file on disk.
- **Example**:
  ```bash
  gork -c 50 -d 30s --json -o ./reports/benchmark.json https://api.example.com
  ```

---

### 6. Informational Flags

#### `-v, --version`
- **Description**: Prints the Gork binary version and exits.
- **Example**:
  ```bash
  gork -v
  ```

#### `-h, --help`
- **Description**: Displays the CLI usage guide and options summary.
- **Example**:
  ```bash
  gork --help
  ```

---

## 🌐 Positional Arguments

### `<url>`
- **Requirement**: **Mandatory** positional argument placed after options.
- **Supported Schemes**: `http://` or `https://`
- **Validation**:
  - Must include scheme and valid host (e.g. `http://localhost:8080` or `https://api.mycorp.internal/v1`).
  - Omitting the URL or providing unsupported schemes (e.g. `ftp://`) results in an immediate validation error before any connections are established.

---

## 💡 Real-World Usage Recipes

### 1. Rapid API Health Check
```bash
gork -c 10 -t 2 -d 5s https://api.example.com/healthz
```

### 2. High-Concurrency Stress Test
```bash
gork -c 500 -t 8 -d 1m https://api.example.com/catalog
```

### 3. Rate-Paced Load Test (Fixed QPS)
Test how latencies behave under a steady 5,000 req/sec load:
```bash
gork -c 100 -r 5000 -d 2m -w 5s https://api.example.com/search
```

### 4. Smooth Concurrency Ramping (Soak Test)
Ramp from 1 to 100 connections over 15 seconds, hold for 1 minute, and ramp down over 10 seconds:
```bash
gork -c 100 --ramp-up 15s -d 1m --ramp-down 10s https://api.example.com/checkout
```

### 5. Custom Multi-Stage Traffic Curve
Simulate morning traffic surge, lunchtime peak, and evening taper:
```bash
gork --stages "30s:10->50, 1m:50, 30s:50->200, 2m:200, 30s:200->20" https://api.example.com
```

### 6. Automated Breakpoint Discovery
Find the exact concurrency where the service exceeds 250ms latency or 1% error rate:
```bash
gork --step-load \
     --step-conns 20 \
     --step-duration 5s \
     --max-latency 250ms \
     --max-error-rate 1.0 \
     --max-conns 800 \
     https://api.example.com/order
```

### 7. POST API with External JSON Payload & Bearer Authentication
```bash
gork -X POST \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer s3cr3t-t0k3n" \
     --body-file ./requests/create_user.json \
     -c 50 -d 30s \
     https://api.example.com/v1/users
```

### 8. Testing Local TLS Services with Self-Signed Certificates
```bash
gork -k -c 25 -d 15s https://localhost:8443/api/v1/ping
```

### 9. CI/CD Automated Regression Test
Run benchmark, output JSON to file, and verify thresholds in CI pipeline:
```bash
gork -c 50 -d 20s --json -o ./benchmark-result.json https://staging-api.example.com
```

---

## ⚠️ Validation Rules & Mutual Exclusions

Gork enforces strict configuration validation before initiating network requests:

1. **Payload Mutual Exclusivity**: You cannot specify both `-b`/`--body` and `--body-file`. Doing so produces an immediate error:
   ```text
   error: cannot use both -b/--body and --body-file
   ```
2. **URL Scheme Enforcement**: Only `http` and `https` URLs are supported. Missing scheme or invalid hosts trigger:
   ```text
   error: unsupported URL scheme "ftp": only http and https are supported
   ```
3. **Thread Safety Ceiling**: Threads cannot exceed `10,000` to protect the host machine against accidental resource exhaustion.
4. **Step-Load Parameter Constraints**: When `--step-load` is enabled:
   - `--step-conns` must be $> 0$
   - `--step-duration` must be $> 0$
   - `--max-latency` must be $> 0$
   - `--max-error-rate` must be $\ge 0.0$
   - `--max-conns` must be $> 0$
5. **Stage Duration Constraints**: In `--stages`, every stage duration must be $> 0$, and concurrency must have at least one positive value.

