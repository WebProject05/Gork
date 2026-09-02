package cli

import (
	"bytes"
	"testing"
	"time"
)

func TestParseStages(t *testing.T) {
	tests := []struct {
		input       string
		wantErr     bool
		expectedLen int
	}{
		{
			input:       "10s:10->50, 30s:50, 10s:50->0",
			wantErr:     false,
			expectedLen: 3,
		},
		{
			input:       "10s:20, 20s:50",
			wantErr:     false,
			expectedLen: 2,
		},
		{
			input:       "1m:100",
			wantErr:     false,
			expectedLen: 1,
		},
		{
			input:   "invalid_duration:50",
			wantErr: true,
		},
		{
			input:   "10s:invalid_conns",
			wantErr: true,
		},
		{
			input:   "10s:50->invalid",
			wantErr: true,
		},
		{
			input:   "missing_colon",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			stages, err := parseStages(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseStages(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && len(stages) != tt.expectedLen {
				t.Errorf("expected %d stages, got %d", tt.expectedLen, len(stages))
			}
		})
	}
}

func TestParseRampUpAndDownFlags(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	args := []string{
		"-c", "50",
		"--ramp-up", "5s",
		"-d", "20s",
		"--ramp-down", "5s",
		"http://example.com",
	}

	cfg, err := ParseArgs(args, stdout, stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Stages) != 3 {
		t.Fatalf("expected 3 auto-generated stages, got %d", len(cfg.Stages))
	}

	// Stage 1: Ramp Up
	if cfg.Stages[0].Duration != 5*time.Second || cfg.Stages[0].StartConns != 1 || cfg.Stages[0].TargetConns != 50 {
		t.Errorf("unexpected Stage 1: %+v", cfg.Stages[0])
	}
	// Stage 2: Sustained
	if cfg.Stages[1].Duration != 20*time.Second || cfg.Stages[1].StartConns != 50 || cfg.Stages[1].TargetConns != 50 {
		t.Errorf("unexpected Stage 2: %+v", cfg.Stages[1])
	}
	// Stage 3: Ramp Down
	if cfg.Stages[2].Duration != 5*time.Second || cfg.Stages[2].StartConns != 50 || cfg.Stages[2].TargetConns != 1 {
		t.Errorf("unexpected Stage 3: %+v", cfg.Stages[2])
	}

	// Overall duration should be 30s
	if cfg.Duration != 30*time.Second {
		t.Errorf("expected overall duration 30s, got %v", cfg.Duration)
	}
}

func TestStepLoadFlags(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	args := []string{
		"--step-load",
		"--step-conns", "25",
		"--step-duration", "3s",
		"--max-latency", "350ms",
		"--max-error-rate", "2.5",
		"--max-conns", "500",
		"http://example.com",
	}

	cfg, err := ParseArgs(args, stdout, stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.StepLoad {
		t.Errorf("expected StepLoad true")
	}
	if cfg.StepConns != 25 {
		t.Errorf("expected StepConns 25, got %d", cfg.StepConns)
	}
	if cfg.StepDuration != 3*time.Second {
		t.Errorf("expected StepDuration 3s, got %v", cfg.StepDuration)
	}
	if cfg.MaxLatency != 350*time.Millisecond {
		t.Errorf("expected MaxLatency 350ms, got %v", cfg.MaxLatency)
	}
	if cfg.MaxErrorRate != 2.5 {
		t.Errorf("expected MaxErrorRate 2.5, got %f", cfg.MaxErrorRate)
	}
	if cfg.MaxConns != 500 {
		t.Errorf("expected MaxConns 500, got %d", cfg.MaxConns)
	}
}
