package metrics

import (
	"errors"
	"testing"
	"time"
)

func TestCollectorRecordAndMerge(t *testing.T) {
	c1 := NewCollector()
	c2 := NewCollector()

	// Record to c1
	c1.Record(Result{
		StatusCode:      200,
		Latency:         10 * time.Millisecond,
		BytesRead:       500,
		BytesSent:       150,
		ConnReused:      false,
		DNSDuration:     2 * time.Millisecond,
		TCPDuration:     3 * time.Millisecond,
		TTFBDuration:    7 * time.Millisecond,
		ContentDuration: 3 * time.Millisecond,
	})
	c1.Record(Result{
		StatusCode:      200,
		Latency:         20 * time.Millisecond,
		BytesRead:       600,
		BytesSent:       150,
		ConnReused:      true,
		TTFBDuration:    15 * time.Millisecond,
		ContentDuration: 5 * time.Millisecond,
	})

	// Record to c2
	c2.Record(Result{
		StatusCode:      500,
		Latency:         30 * time.Millisecond,
		BytesRead:       200,
		BytesSent:       150,
		ConnReused:      true,
		TTFBDuration:    25 * time.Millisecond,
		ContentDuration: 5 * time.Millisecond,
	})
	c2.Record(Result{
		Latency:    5 * time.Millisecond,
		BytesSent:  150,
		ConnReused: false,
		Error:      errors.New("connection refused"),
	})

	// Merge c2 into c1
	c1.Merge(c2)

	if c1.TotalRequests != 4 {
		t.Errorf("expected 4 total requests, got %d", c1.TotalRequests)
	}
	if c1.Successful != 2 {
		t.Errorf("expected 2 successful requests, got %d", c1.Successful)
	}
	if c1.Failed != 2 {
		t.Errorf("expected 2 failed requests, got %d", c1.Failed)
	}
	if c1.BytesRead != 1300 {
		t.Errorf("expected 1300 bytes read, got %d", c1.BytesRead)
	}
	if c1.BytesSent != 600 {
		t.Errorf("expected 600 bytes sent, got %d", c1.BytesSent)
	}
	if c1.ConnsReused != 2 {
		t.Errorf("expected 2 reused connections, got %d", c1.ConnsReused)
	}
	if c1.ConnsNew != 2 {
		t.Errorf("expected 2 new connections, got %d", c1.ConnsNew)
	}
	if c1.Count2xx != 2 {
		t.Errorf("expected 2 2xx responses, got %d", c1.Count2xx)
	}
	if c1.Count5xx != 1 {
		t.Errorf("expected 1 5xx response, got %d", c1.Count5xx)
	}
	if c1.MinLatency != 5*time.Millisecond {
		t.Errorf("expected MinLatency 5ms, got %v", c1.MinLatency)
	}
	if c1.MaxLatency != 30*time.Millisecond {
		t.Errorf("expected MaxLatency 30ms, got %v", c1.MaxLatency)
	}
	if c1.TTFBStats.Count != 3 {
		t.Errorf("expected 3 TTFB measurements, got %d", c1.TTFBStats.Count)
	}
	if c1.StatusCodes[200] != 2 {
		t.Errorf("expected 2 status 200, got %d", c1.StatusCodes[200])
	}
	if c1.StatusCodes[500] != 1 {
		t.Errorf("expected 1 status 500, got %d", c1.StatusCodes[500])
	}
	if c1.Errors["Connection Refused"] != 1 {
		t.Errorf("expected 1 Connection Refused error, got %d", c1.Errors["Connection Refused"])
	}
}

func BenchmarkCollectorRecord(b *testing.B) {
	c := NewCollector()
	res := Result{
		StatusCode: 200,
		Latency:    5 * time.Millisecond,
		BytesRead:  1024,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.Record(res)
	}
}
