package cli

import (
	"fmt"
	"net/url"

	"gork/internal/config"
)

const (
	// MaxThreads is the safety ceiling to prevent system resource exhaustion.
	MaxThreads = 1000
)

// Validate checks the configuration for fatal errors.
func Validate(cfg *config.Config) error {
	if cfg.URL == "" {
		return fmt.Errorf("URL is required as the final positional argument")
	}
	if _, err := url.ParseRequestURI(cfg.URL); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	
	// Thread validation
	if cfg.Threads <= 0 {
		return fmt.Errorf("threads (-t) must be greater than 0")
	}
	if cfg.Threads > MaxThreads {
		return fmt.Errorf("threads (-t) cannot exceed %d to prevent system crash", MaxThreads)
	}
	
	// Connection validation
	if cfg.Connections <= 0 {
		return fmt.Errorf("connections (-c) must be greater than 0")
	}
	
	// Duration validation
	if cfg.Duration <= 0 {
		return fmt.Errorf("duration (-d) must be greater than 0")
	}
	
	return nil
}