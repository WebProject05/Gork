package config

import "time"

// StageConfig defines a discrete execution stage with concurrency progression.
type StageConfig struct {
	Name        string        `json:"name"`
	Duration    time.Duration `json:"duration"`
	StartConns  int           `json:"start_conns"`
	TargetConns int           `json:"target_conns"`
}

// Config holds all runtime configuration settings for Gork.
type Config struct {
	// Concurrency & System
	Threads     int // Number of OS threads / scheduler parallelism
	Connections int // Number of concurrent HTTP connections / workers
	CPUs        int // GOMAXPROCS override (0 = default/threads)

	// Benchmark Timing
	Duration time.Duration // Benchmark execution duration
	Warmup   time.Duration // Warmup duration before metrics collection

	// Ramping & Stepped Load Profiles
	RampUp   time.Duration // Ramp-up duration from 1 to Connections
	RampDown time.Duration // Ramp-down duration from Connections to 1
	Stages   []StageConfig // Custom multi-stage execution profile

	// Breakpoint / Step-Load Saturation Testing
	StepLoad     bool          // Enable automated step-load saturation testing
	StepConns    int           // Concurrency increment per step
	StepDuration time.Duration // Duration to hold each step
	MaxLatency   time.Duration // Stop threshold: max acceptable P95 latency
	MaxErrorRate float64       // Stop threshold: max acceptable error rate % (e.g. 1.0 = 1%)
	MaxConns     int           // Safety ceiling concurrency for step load

	// HTTP Request
	URL               string        // Target URL
	Method            string        // HTTP method (GET, POST, PUT, etc.)
	Headers           []string      // Custom HTTP headers (key: value)
	Body              []byte        // Request payload
	Timeout           time.Duration // Per-request HTTP timeout
	Insecure          bool          // Skip TLS certificate verification
	DisableKeepAlives bool          // Disable HTTP Keep-Alive connection reuse
	Rate              int           // Overall rate limit in requests per second (0 = unlimited)

	// Output
	JSONOutput bool   // Format output as JSON
	OutFile    string // Destination file for benchmark results
}
