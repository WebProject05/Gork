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

// Summary represents the comprehensive benchmark report.
type Summary struct {
	TotalRequests  int64         `json:"total_requests"`
	Successful     int64         `json:"successful"`
	Failed         int64         `json:"failed"`
	Duration       time.Duration `json:"duration"`
	RequestsPerSec float64       `json:"requests_per_sec"`
	BytesRead      int64         `json:"bytes_read"`
	BytesPerSec    float64       `json:"bytes_per_sec"`

	Latency     LatencySummary   `json:"latency"`
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

	if duration > 0 {
		s.RequestsPerSec = float64(s.TotalRequests) / duration.Seconds()
		s.BytesPerSec = float64(s.BytesRead) / duration.Seconds()
	}

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

	return s
}
