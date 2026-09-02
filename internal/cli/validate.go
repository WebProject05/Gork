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

	// Step load / Breakpoint validation
	if cfg.StepLoad {
		if cfg.StepConns <= 0 {
			return fmt.Errorf("step-conns must be greater than 0")
		}
		if cfg.StepDuration <= 0 {
			return fmt.Errorf("step-duration must be greater than 0")
		}
		if cfg.MaxLatency <= 0 {
			return fmt.Errorf("max-latency must be greater than 0")
		}
		if cfg.MaxErrorRate < 0 {
			return fmt.Errorf("max-error-rate cannot be negative")
		}
		if cfg.MaxConns <= 0 {
			return fmt.Errorf("max-conns must be greater than 0")
		}
	} else if len(cfg.Stages) > 0 {
		// Stage validation
		for i, s := range cfg.Stages {
			if s.Duration <= 0 {
				return fmt.Errorf("stage %d %q duration must be greater than 0", i+1, s.Name)
			}
			if s.StartConns < 0 || s.TargetConns < 0 || (s.StartConns == 0 && s.TargetConns == 0) {
				return fmt.Errorf("stage %d %q concurrency must have at least one positive value", i+1, s.Name)
			}
		}
	} else {
		// Standard connection and duration validation
		if cfg.Connections <= 0 {
			return fmt.Errorf("connections (-c) must be greater than 0")
		}
		if cfg.Duration <= 0 {
			return fmt.Errorf("duration (-d) must be greater than 0")
		}
	}

	// Ramp duration checks
	if cfg.RampUp < 0 {
		return fmt.Errorf("ramp-up cannot be negative")
	}
	if cfg.RampDown < 0 {
		return fmt.Errorf("ramp-down cannot be negative")
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
