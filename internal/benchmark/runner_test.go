package benchmark

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"gork/internal/config"
)

func TestBenchmarkRunAgainstMockServer(t *testing.T) {
	var requestCount int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok"}`)
	}))
	defer server.Close()

	cfg := &config.Config{
		URL:         server.URL,
		Threads:     2,
		Connections: 10,
		Duration:    500 * time.Millisecond,
		Timeout:     2 * time.Second,
		Method:      "GET",
	}

	summary := Run(cfg)

	if summary.TotalRequests == 0 {
		t.Fatal("expected > 0 total requests")
	}
	if summary.Successful == 0 {
		t.Fatal("expected > 0 successful requests")
	}
	if summary.Failed != 0 {
		t.Errorf("expected 0 failed requests, got %d", summary.Failed)
	}
	if summary.StatusCodes[200] != summary.TotalRequests {
		t.Errorf("expected all requests to be 200 OK, got %d vs %d", summary.StatusCodes[200], summary.TotalRequests)
	}
	if summary.RequestsPerSec <= 0 {
		t.Errorf("expected positive requests per second, got %f", summary.RequestsPerSec)
	}
	if summary.Latency.Min <= 0 {
		t.Errorf("expected positive min latency, got %v", summary.Latency.Min)
	}
	if summary.BytesRead <= 0 {
		t.Errorf("expected positive bytes read, got %d", summary.BytesRead)
	}
}

func TestBenchmarkRunWithContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.Config{
		URL:         server.URL,
		Threads:     1,
		Connections: 2,
		Duration:    10 * time.Second, // Long duration
		Timeout:     2 * time.Second,
		Method:      "GET",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	summary := RunWithContext(ctx, cfg)
	elapsed := time.Since(start)

	if elapsed > 2*time.Second {
		t.Errorf("expected fast cancellation, took %v", elapsed)
	}
	if summary == nil {
		t.Fatal("expected non-nil summary on cancellation")
	}
}
