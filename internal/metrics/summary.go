package metrics

import (
	"math"
	"time"
)

// LatencySummary contains statistical distributions of request latencies.
type LatencySummary struct {
	Min    time.Duration `json:"min"`
	Avg    time.Duration `json:"avg"`
	Max    time.Duration `json:"max"`
	StdDev time.Duration `json:"stddev"`
	P50    time.Duration `json:"p50"`
	P75    time.Duration `json:"p75"`
	P90    time.Duration `json:"p90"`
	P95    time.Duration `json:"p95"`
	P99    time.Duration `json:"p99"`
	P999   time.Duration `json:"p99_9"`
}

// PhaseSummary holds min/avg/max timing for an individual network phase.
type PhaseSummary struct {
	Min time.Duration `json:"min"`
	Avg time.Duration `json:"avg"`
	Max time.Duration `json:"max"`
}

// LatencyPhaseSummary encapsulates breakdown of latency across network lifecycle stages.
type LatencyPhaseSummary struct {
	DNS     PhaseSummary `json:"dns"`
	TCP     PhaseSummary `json:"tcp"`
	TLS     PhaseSummary `json:"tls"`
	TTFB    PhaseSummary `json:"ttfb"`
	Content PhaseSummary `json:"content"`
}

// ConnectionSummary represents TCP keep-alive connection reuse metrics.
type ConnectionSummary struct {
	TotalAttempts int64   `json:"total_attempts"`
	Reused        int64   `json:"reused"`
	New           int64   `json:"new"`
	ReusedPct     float64 `json:"reused_pct"`
}

// TransferSummary tracks bidirectional payload transfer throughput.
type TransferSummary struct {
	BytesSent    int64   `json:"bytes_sent"`
	BytesRead    int64   `json:"bytes_read"`
	TotalBytes   int64   `json:"total_bytes"`
	UploadRate   float64 `json:"upload_bytes_per_sec"`
	DownloadRate float64 `json:"download_bytes_per_sec"`
	MinBodyBytes int64   `json:"min_body_bytes"`
	AvgBodyBytes int64   `json:"avg_body_bytes"`
	MaxBodyBytes int64   `json:"max_body_bytes"`
}

// AvailabilitySummary categorizes request success and status code classes.
type AvailabilitySummary struct {
	SuccessRate float64 `json:"success_rate_pct"`
	ErrorRate   float64 `json:"error_rate_pct"`
	Count2xx    int64   `json:"count_2xx"`
	Count3xx    int64   `json:"count_3xx"`
	Count4xx    int64   `json:"count_4xx"`
	Count5xx    int64   `json:"count_5xx"`
}

// StageSummary records performance metrics for a single stage in a multi-stage benchmark.
type StageSummary struct {
	Name        string        `json:"name"`
	Duration    time.Duration `json:"duration"`
	Concurrency string        `json:"concurrency"`
	Requests    int64         `json:"requests"`
	RPS         float64       `json:"rps"`
	AvgLatency  time.Duration `json:"avg_latency"`
	P95Latency  time.Duration `json:"p95_latency"`
	Errors      int64         `json:"errors"`
}

// BreakpointSummary documents the saturation / breaking point discovered in step-load testing.
type BreakpointSummary struct {
	Triggered         bool          `json:"triggered"`
	BreakingConns     int           `json:"breaking_conns,omitempty"`
	BreakingRPS       float64       `json:"breaking_rps,omitempty"`
	BreakingLatency   time.Duration `json:"breaking_latency,omitempty"`
	BreakingErrorRate float64       `json:"breaking_error_rate,omitempty"`
	Reason            string        `json:"reason,omitempty"`
}

// Summary represents the comprehensive benchmark report.
type Summary struct {
	TotalRequests  int64         `json:"total_requests"`
	Successful     int64         `json:"successful"`
	Failed         int64         `json:"failed"`
	Duration       time.Duration `json:"duration"`
	RequestsPerSec float64       `json:"requests_per_sec"`
	BytesRead      int64         `json:"bytes_read"`
	BytesPerSec    float64       `json:"bytes_per_sec"`

	// Extended Metrics
	Latency      LatencySummary      `json:"latency"`
	Phases       LatencyPhaseSummary `json:"phases"`
	Connections  ConnectionSummary   `json:"connections"`
	Transfer     TransferSummary     `json:"transfer"`
	Availability AvailabilitySummary `json:"availability"`

	// Stage and Breakpoint Telemetry
	Stages     []StageSummary     `json:"stages,omitempty"`
	Breakpoint *BreakpointSummary `json:"breakpoint,omitempty"`

	StatusCodes map[int]int64    `json:"status_codes"`
	Errors      map[string]int64 `json:"errors,omitempty"`
}

func (c *Collector) latencyCount() int64 {
	var count int64
	for i := 0; i < numBuckets; i++ {
		count += c.Buckets[i]
	}
	return count
}

func (c *Collector) percentile(p float64, totalCount int64) time.Duration {
	if totalCount == 0 {
		return 0
	}
	target := int64(math.Ceil(float64(totalCount) * p))
	var count int64
	for i := 0; i < numBuckets; i++ {
		count += c.Buckets[i]
		if count >= target {
			lat := bucketToLatency(i)
			if c.MinLatency > 0 && lat < c.MinLatency {
				return c.MinLatency
			}
			if c.MaxLatency > 0 && lat > c.MaxLatency {
				return c.MaxLatency
			}
			return lat
		}
	}
	return c.MaxLatency
}

// CalculateSummary aggregates collected metrics over the specified test duration.
func (c *Collector) CalculateSummary(duration time.Duration) *Summary {
	s := &Summary{
		TotalRequests: c.TotalRequests,
		Successful:    c.Successful,
		Failed:        c.Failed,
		Duration:      duration,
		BytesRead:     c.BytesRead,
		StatusCodes:   c.StatusCodes,
		Errors:        c.Errors,
	}

	durationSec := duration.Seconds()
	if durationSec > 0 {
		s.RequestsPerSec = float64(s.TotalRequests) / durationSec
		s.BytesPerSec = float64(s.BytesRead) / durationSec
	}

	// 1. Overall Latency Distributions
	latCount := c.latencyCount()
	if latCount > 0 {
		s.Latency.Min = c.MinLatency
		s.Latency.Max = c.MaxLatency
		s.Latency.Avg = time.Duration(c.TotalLatency.Nanoseconds() / latCount)

		if latCount > 1 {
			meanNs := float64(c.TotalLatency.Nanoseconds()) / float64(latCount)
			meanSq := c.SumSqLatency / float64(latCount)
			variance := meanSq - (meanNs * meanNs)
			if variance > 0 {
				s.Latency.StdDev = time.Duration(math.Sqrt(variance))
			}
		}

		s.Latency.P50 = c.percentile(0.50, latCount)
		s.Latency.P75 = c.percentile(0.75, latCount)
		s.Latency.P90 = c.percentile(0.90, latCount)
		s.Latency.P95 = c.percentile(0.95, latCount)
		s.Latency.P99 = c.percentile(0.99, latCount)
		s.Latency.P999 = c.percentile(0.999, latCount)
	}

	// 2. HTTP Timing Phase Breakdown
	s.Phases.DNS = PhaseSummary{Min: c.DNSStats.Min, Avg: c.DNSStats.Avg(), Max: c.DNSStats.Max}
	s.Phases.TCP = PhaseSummary{Min: c.TCPStats.Min, Avg: c.TCPStats.Avg(), Max: c.TCPStats.Max}
	s.Phases.TLS = PhaseSummary{Min: c.TLSStats.Min, Avg: c.TLSStats.Avg(), Max: c.TLSStats.Max}
	s.Phases.TTFB = PhaseSummary{Min: c.TTFBStats.Min, Avg: c.TTFBStats.Avg(), Max: c.TTFBStats.Max}
	s.Phases.Content = PhaseSummary{Min: c.ContentStats.Min, Avg: c.ContentStats.Avg(), Max: c.ContentStats.Max}

	// 3. Connection Reuse Metrics
	totalAttempts := c.ConnsReused + c.ConnsNew
	s.Connections.TotalAttempts = totalAttempts
	s.Connections.Reused = c.ConnsReused
	s.Connections.New = c.ConnsNew
	if totalAttempts > 0 {
		s.Connections.ReusedPct = (float64(c.ConnsReused) / float64(totalAttempts)) * 100.0
	}

	// 4. Bidirectional Data Transfer & Response Body Sizes
	s.Transfer.BytesSent = c.BytesSent
	s.Transfer.BytesRead = c.BytesRead
	s.Transfer.TotalBytes = c.BytesSent + c.BytesRead
	if durationSec > 0 {
		s.Transfer.UploadRate = float64(c.BytesSent) / durationSec
		s.Transfer.DownloadRate = float64(c.BytesRead) / durationSec
	}
	s.Transfer.MinBodyBytes = c.MinBodySize
	s.Transfer.MaxBodyBytes = c.MaxBodySize
	if c.TotalRequests > 0 {
		s.Transfer.AvgBodyBytes = c.BytesRead / c.TotalRequests
	}

	// 5. Availability & Status Code Classes
	if c.TotalRequests > 0 {
		s.Availability.SuccessRate = (float64(c.Successful) / float64(c.TotalRequests)) * 100.0
		s.Availability.ErrorRate = (float64(c.Failed) / float64(c.TotalRequests)) * 100.0
	}
	s.Availability.Count2xx = c.Count2xx
	s.Availability.Count3xx = c.Count3xx
	s.Availability.Count4xx = c.Count4xx
	s.Availability.Count5xx = c.Count5xx

	return s
}
