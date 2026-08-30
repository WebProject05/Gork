package cli

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseArgsDefaults(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cfg, err := ParseArgs([]string{"http://example.com"}, stdout, stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.URL != "http://example.com" {
		t.Errorf("expected URL http://example.com, got %s", cfg.URL)
	}
	if cfg.Threads != 2 {
		t.Errorf("expected 2 threads, got %d", cfg.Threads)
	}
	if cfg.Connections != 10 {
		t.Errorf("expected 10 connections, got %d", cfg.Connections)
	}
	if cfg.Duration != 10*time.Second {
		t.Errorf("expected 10s duration, got %v", cfg.Duration)
	}
	if cfg.Method != "GET" {
		t.Errorf("expected GET method, got %s", cfg.Method)
	}
	if cfg.Timeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", cfg.Timeout)
	}
	if cfg.Insecure != false {
		t.Errorf("expected insecure false, got %v", cfg.Insecure)
	}
}

func TestParseArgsCustomFlags(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	args := []string{
		"-c", "50",
		"-t", "4",
		"-d", "30s",
		"-w", "5s",
		"-r", "1000",
		"-X", "POST",
		"-H", "Content-Type: application/json",
		"-H", "Authorization: Bearer token123",
		"-b", `{"key":"val"}`,
		"-k",
		"--no-keepalive",
		"--json",
		"-o", "output.json",
		"https://api.test.com/v1",
	}

	cfg, err := ParseArgs(args, stdout, stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.URL != "https://api.test.com/v1" {
		t.Errorf("expected URL https://api.test.com/v1, got %s", cfg.URL)
	}
	if cfg.Connections != 50 {
		t.Errorf("expected 50 connections, got %d", cfg.Connections)
	}
	if cfg.Threads != 4 {
		t.Errorf("expected 4 threads, got %d", cfg.Threads)
	}
	if cfg.Duration != 30*time.Second {
		t.Errorf("expected 30s duration, got %v", cfg.Duration)
	}
	if cfg.Warmup != 5*time.Second {
		t.Errorf("expected 5s warmup, got %v", cfg.Warmup)
	}
	if cfg.Rate != 1000 {
		t.Errorf("expected rate 1000, got %d", cfg.Rate)
	}
	if cfg.Method != "POST" {
		t.Errorf("expected POST method, got %s", cfg.Method)
	}
	if len(cfg.Headers) != 2 {
		t.Errorf("expected 2 headers, got %d", len(cfg.Headers))
	}
	if string(cfg.Body) != `{"key":"val"}` {
		t.Errorf("expected body to match, got %s", string(cfg.Body))
	}
	if !cfg.Insecure {
		t.Errorf("expected Insecure true, got %v", cfg.Insecure)
	}
	if !cfg.DisableKeepAlives {
		t.Errorf("expected DisableKeepAlives true, got %v", cfg.DisableKeepAlives)
	}
	if !cfg.JSONOutput {
		t.Errorf("expected JSONOutput true, got %v", cfg.JSONOutput)
	}
	if cfg.OutFile != "output.json" {
		t.Errorf("expected OutFile output.json, got %s", cfg.OutFile)
	}
}

func TestParseArgsBodyFile(t *testing.T) {
	tempDir := t.TempDir()
	bodyPath := filepath.Join(tempDir, "body.json")
	if err := os.WriteFile(bodyPath, []byte(`{"hello":"world"}`), 0644); err != nil {
		t.Fatalf("failed to write body file: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cfg, err := ParseArgs([]string{"--body-file", bodyPath, "http://example.com"}, stdout, stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(cfg.Body) != `{"hello":"world"}` {
		t.Errorf("expected body file content, got %s", string(cfg.Body))
	}
}

func TestParseArgsMutualExclusiveBody(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_, err := ParseArgs([]string{"-b", "inline", "--body-file", "file.json", "http://example.com"}, stdout, stderr)
	if err == nil {
		t.Fatalf("expected error for mutually exclusive body args, got nil")
	}
}

func TestParseArgsVersionAndHelp(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_, err := ParseArgs([]string{"--version"}, stdout, stderr)
	if !errors.Is(err, flag.ErrHelp) {
		t.Errorf("expected ErrHelp on version flag, got %v", err)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("gork v")) {
		t.Errorf("expected version output in stdout, got %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	_, err = ParseArgs([]string{"--help"}, stdout, stderr)
	if !errors.Is(err, flag.ErrHelp) {
		t.Errorf("expected ErrHelp on help flag, got %v", err)
	}
	if !bytes.Contains(stderr.Bytes(), []byte("Usage:")) {
		t.Errorf("expected usage output in stderr, got %s", stderr.String())
	}
}
