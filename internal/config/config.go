package config

import "time"

// Config holds all runtime configuration settings for Gork.
type Config struct {
	// Concurrency & System
	Threads     int // Number of OS threads / scheduler parallelism
	Connections int // Number of concurrent HTTP connections / workers
	CPUs        int // GOMAXPROCS override (0 = default/threads)

	// Benchmark Timing
	Duration time.Duration // Benchmark execution duration
	Warmup   time.Duration // Warmup duration before metrics collection

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
