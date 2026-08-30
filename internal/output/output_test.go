package output

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"gork/internal/config"
	"gork/internal/metrics"
)

func TestPrintTerminal(t *testing.T) {
	cfg := &config.Config{
		URL:         "http://example.com/api",
		Threads:     2,
		Connections: 10,
		Duration:    5 * time.Second,
		Rate:        500,
	}

	summary := &metrics.Summary{
		TotalRequests:  1000,
		Successful:     980,
		Failed:         20,
		Duration:       5 * time.Second,
		RequestsPerSec: 200.0,
		BytesRead:      1024000,
		BytesPerSec:    204800,
		Latency: metrics.LatencySummary{
			Min:    500 * time.Microsecond,
			Avg:    2 * time.Millisecond,
			Max:    50 * time.Millisecond,
			StdDev: 1 * time.Millisecond,
			P50:    1500 * time.Microsecond,
			P75:    2 * time.Millisecond,
			P90:    3 * time.Millisecond,
			P95:    5 * time.Millisecond,
			P99:    10 * time.Millisecond,
			P999:   30 * time.Millisecond,
		},
		StatusCodes: map[int]int64{
			200: 980,
			500: 20,
		},
		Errors: map[string]int64{
			"Timeout": 20,
		},
	}

	out := PrintTerminal(cfg, summary)

	if !strings.Contains(out, "Running 5s test @ http://example.com/api") {
		t.Errorf("expected header text in terminal output")
	}
	if !strings.Contains(out, "Latency Statistics:") {
		t.Errorf("expected Latency Statistics section")
	}
	if !strings.Contains(out, "Requests/sec:  200.00") {
		t.Errorf("expected requests/sec formatting")
	}
	if !strings.Contains(out, "[200 OK]: 980") {
		t.Errorf("expected 200 OK status code formatting")
	}
	if !strings.Contains(out, "Timeout: 20") {
		t.Errorf("expected error breakdown in output")
	}
}

func TestPrintJSON(t *testing.T) {
	summary := &metrics.Summary{
		TotalRequests:  500,
		Successful:     500,
		Failed:         0,
		Duration:       2 * time.Second,
		RequestsPerSec: 250.0,
		BytesRead:      50000,
		BytesPerSec:    25000,
		Latency: metrics.LatencySummary{
			Min:    100 * time.Microsecond,
			Avg:    1 * time.Millisecond,
			Max:    10 * time.Millisecond,
			StdDev: 500 * time.Microsecond,
			P50:    800 * time.Microsecond,
			P75:    1200 * time.Microsecond,
			P90:    2 * time.Millisecond,
			P95:    3 * time.Millisecond,
			P99:    5 * time.Millisecond,
			P999:   8 * time.Millisecond,
		},
		StatusCodes: map[int]int64{
			200: 500,
		},
	}

	jsonStr, err := PrintJSON(summary)
	if err != nil {
		t.Fatalf("unexpected JSON error: %v", err)
	}

	var report JSONReport
	if err := json.Unmarshal([]byte(jsonStr), &report); err != nil {
		t.Fatalf("invalid JSON generated: %v", err)
	}

	if report.TotalRequests != 500 {
		t.Errorf("expected total_requests 500, got %d", report.TotalRequests)
	}
	if report.RequestsPerSec != 250.0 {
		t.Errorf("expected requests_per_sec 250.0, got %f", report.RequestsPerSec)
	}
	if report.LatencyMs["min"] <= 0 {
		t.Errorf("expected positive min latency in ms, got %f", report.LatencyMs["min"])
	}
}
