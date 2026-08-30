package cli

import (
	"fmt"
	"net/url"
	"strings"

	"gork/internal/config"
)

const (
	// MaxThreads is the safety ceiling to prevent system resource exhaustion.
	MaxThreads = 10000
)

// Validate checks the configuration for fatal errors.
func Validate(cfg *config.Config) error {
	if cfg.URL == "" {
		return fmt.Errorf("target URL is required as a positional argument (e.g., http://localhost:8080/api)")
	}

	u, err := url.ParseRequestURI(cfg.URL)
	if err != nil || u.Host == "" {
		return fmt.Errorf("invalid URL %q: must include scheme and host (e.g. http://example.com)", cfg.URL)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("unsupported URL scheme %q: only http and https are supported", u.Scheme)
	}

	// Thread validation
	if cfg.Threads <= 0 {
		return fmt.Errorf("threads (-t) must be greater than 0")
	}
	if cfg.Threads > MaxThreads {
		return fmt.Errorf("threads (-t) cannot exceed %d", MaxThreads)
	}

	// Connection validation
	if cfg.Connections <= 0 {
		return fmt.Errorf("connections (-c) must be greater than 0")
	}

	// Duration validation
	if cfg.Duration <= 0 {
		return fmt.Errorf("duration (-d) must be greater than 0")
	}

	// Warmup validation
	if cfg.Warmup < 0 {
		return fmt.Errorf("warmup (-w) cannot be negative")
	}

	// Timeout validation
	if cfg.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}

	// Rate validation
	if cfg.Rate < 0 {
		return fmt.Errorf("rate (-r) cannot be negative")
	}

	// CPU validation
	if cfg.CPUs < 0 {
		return fmt.Errorf("cpus cannot be negative")
	}

	// Method validation
	if cfg.Method == "" {
		cfg.Method = "GET"
	} else {
		cfg.Method = strings.ToUpper(cfg.Method)
	}

	return nil
}