package metrics

import (
	"testing"
	"time"
)

func TestCalculateSummary(t *testing.T) {
	c := NewCollector()

	// Add 100 requests with increasing latencies: 1ms, 2ms, ... 100ms
	for i := 1; i <= 100; i++ {
		c.Record(Result{
			StatusCode: 200,
			Latency:    time.Duration(i) * time.Millisecond,
			BytesRead:  1000,
		})
	}

	summary := c.CalculateSummary(10 * time.Second)

	if summary.TotalRequests != 100 {
		t.Errorf("expected 100 total requests, got %d", summary.TotalRequests)
	}
	if summary.Successful != 100 {
		t.Errorf("expected 100 successful requests, got %d", summary.Successful)
	}
	if summary.Failed != 0 {
		t.Errorf("expected 0 failed requests, got %d", summary.Failed)
	}
	if summary.RequestsPerSec != 10.0 {
		t.Errorf("expected 10.0 RPS, got %f", summary.RequestsPerSec)
	}
	if summary.BytesRead != 100000 {
		t.Errorf("expected 100000 bytes read, got %d", summary.BytesRead)
	}

	// Exact min/max/avg
	if summary.Latency.Min != 1*time.Millisecond {
		t.Errorf("expected MinLatency 1ms, got %v", summary.Latency.Min)
	}
	if summary.Latency.Max != 100*time.Millisecond {
		t.Errorf("expected MaxLatency 100ms, got %v", summary.Latency.Max)
	}
	if summary.Latency.Avg != 50500*time.Microsecond { // (100 * 101 / 2) * 1000 / 100 = 50500 µs = 50.5ms
		t.Errorf("expected AvgLatency 50.5ms, got %v", summary.Latency.Avg)
	}

	// StdDev should be approximately 28.86 ms for uniform distribution 1..100
	if summary.Latency.StdDev < 28*time.Millisecond || summary.Latency.StdDev > 30*time.Millisecond {
		t.Errorf("expected StdDev ~29ms, got %v", summary.Latency.StdDev)
	}

	// Percentiles (within logarithmic histogram tolerance ~2-3%)
	if summary.Latency.P50 < 45*time.Millisecond || summary.Latency.P50 > 55*time.Millisecond {
		t.Errorf("expected P50 ~50ms, got %v", summary.Latency.P50)
	}
	if summary.Latency.P95 < 90*time.Millisecond || summary.Latency.P95 > 100*time.Millisecond {
		t.Errorf("expected P95 ~95ms, got %v", summary.Latency.P95)
	}
	if summary.Latency.P99 < 95*time.Millisecond || summary.Latency.P99 > 100*time.Millisecond {
		t.Errorf("expected P99 ~99ms, got %v", summary.Latency.P99)
	}

	// Availability & Transfer
	if summary.Availability.SuccessRate != 100.0 {
		t.Errorf("expected 100%% success rate, got %f", summary.Availability.SuccessRate)
	}
	if summary.Availability.Count2xx != 100 {
		t.Errorf("expected 100 2xx responses, got %d", summary.Availability.Count2xx)
	}
	if summary.Transfer.BytesRead != 100000 {
		t.Errorf("expected 100000 download bytes, got %d", summary.Transfer.BytesRead)
	}
}
