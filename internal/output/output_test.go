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
		Phases: metrics.LatencyPhaseSummary{
			DNS:     metrics.PhaseSummary{Min: 100 * time.Microsecond, Avg: 1 * time.Millisecond, Max: 5 * time.Millisecond},
			TCP:     metrics.PhaseSummary{Min: 200 * time.Microsecond, Avg: 2 * time.Millisecond, Max: 8 * time.Millisecond},
			TLS:     metrics.PhaseSummary{Min: 500 * time.Microsecond, Avg: 3 * time.Millisecond, Max: 10 * time.Millisecond},
			TTFB:    metrics.PhaseSummary{Min: 1 * time.Millisecond, Avg: 5 * time.Millisecond, Max: 25 * time.Millisecond},
			Content: metrics.PhaseSummary{Min: 100 * time.Microsecond, Avg: 1 * time.Millisecond, Max: 6 * time.Millisecond},
		},
		Connections: metrics.ConnectionSummary{
			TotalAttempts: 1000,
			Reused:        950,
			New:           50,
			ReusedPct:     95.0,
		},
		Transfer: metrics.TransferSummary{
			BytesSent:    150000,
			BytesRead:    1024000,
			TotalBytes:   1174000,
			UploadRate:   30000,
			DownloadRate: 204800,
			MinBodyBytes: 500,
			AvgBodyBytes: 1024,
			MaxBodyBytes: 2048,
		},
		Availability: metrics.AvailabilitySummary{
			SuccessRate: 98.0,
			ErrorRate:   2.0,
			Count2xx:    980,
			Count5xx:    20,
		},
		Stages: []metrics.StageSummary{
			{Name: "Stage 1", Duration: 2 * time.Second, Concurrency: "10", Requests: 400, RPS: 200.0, AvgLatency: 2 * time.Millisecond, P95Latency: 4 * time.Millisecond, Errors: 5},
			{Name: "Stage 2", Duration: 3 * time.Second, Concurrency: "20", Requests: 600, RPS: 200.0, AvgLatency: 2 * time.Millisecond, P95Latency: 5 * time.Millisecond, Errors: 15},
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
	if !strings.Contains(out, "Stage Progression:") {
		t.Errorf("expected Stage Progression section")
	}
	if !strings.Contains(out, "HTTP Lifecycle Phase Breakdown:") {
		t.Errorf("expected HTTP Lifecycle Phase Breakdown section")
	}
	if !strings.Contains(out, "Connection & Keep-Alive Health:") {
		t.Errorf("expected Connection & Keep-Alive Health section")
	}
	if !strings.Contains(out, "Data Transfer & Throughput:") {
		t.Errorf("expected Data Transfer & Throughput section")
	}
	if !strings.Contains(out, "Requests & Availability:") {
		t.Errorf("expected Requests & Availability section")
	}
	if !strings.Contains(out, "Success Rate:  98.00%") {
		t.Errorf("expected success rate formatting")
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
