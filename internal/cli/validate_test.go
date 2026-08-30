package cli

import (
	"testing"
	"time"

	"gork/internal/config"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &config.Config{
				URL:         "http://localhost:8080/test",
				Threads:     4,
				Connections: 50,
				Duration:    10 * time.Second,
				Timeout:     5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid https config",
			cfg: &config.Config{
				URL:         "https://example.com/api",
				Threads:     1,
				Connections: 1,
				Duration:    1 * time.Second,
				Timeout:     1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing URL",
			cfg: &config.Config{
				URL:         "",
				Threads:     1,
				Connections: 1,
				Duration:    1 * time.Second,
				Timeout:     1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid URL scheme",
			cfg: &config.Config{
				URL:         "ftp://example.com",
				Threads:     1,
				Connections: 1,
				Duration:    1 * time.Second,
				Timeout:     1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "zero threads",
			cfg: &config.Config{
				URL:         "http://example.com",
				Threads:     0,
				Connections: 10,
				Duration:    5 * time.Second,
				Timeout:     1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "threads exceeding ceiling",
			cfg: &config.Config{
				URL:         "http://example.com",
				Threads:     20000,
				Connections: 10,
				Duration:    5 * time.Second,
				Timeout:     1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "zero connections",
			cfg: &config.Config{
				URL:         "http://example.com",
				Threads:     2,
				Connections: 0,
				Duration:    5 * time.Second,
				Timeout:     1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "negative duration",
			cfg: &config.Config{
				URL:         "http://example.com",
				Threads:     2,
				Connections: 10,
				Duration:    -1 * time.Second,
				Timeout:     1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "negative warmup",
			cfg: &config.Config{
				URL:         "http://example.com",
				Threads:     2,
				Connections: 10,
				Duration:    5 * time.Second,
				Warmup:      -1 * time.Second,
				Timeout:     1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "negative rate",
			cfg: &config.Config{
				URL:         "http://example.com",
				Threads:     2,
				Connections: 10,
				Duration:    5 * time.Second,
				Rate:        -10,
				Timeout:     1 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
