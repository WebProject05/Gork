package metrics

import (
	"math"
	"net/url"
	"strings"
	"time"
)

const (
	numBuckets       = 1000
	bucketsPerDecade = 120.0
)

// Result represents a single HTTP transaction outcome.
type Result struct {
	StatusCode int
	Latency    time.Duration
	BytesRead  int64
	Error      error
}

// Collector records metrics online with constant O(1) memory using a high-precision logarithmic histogram.
type Collector struct {
	TotalRequests int64
	Successful    int64
	Failed        int64
	BytesRead     int64

	MinLatency   time.Duration
	MaxLatency   time.Duration
	TotalLatency time.Duration
	SumSqLatency float64 // sum of squared nanoseconds for StdDev calculation

	StatusCodes map[int]int64
	Errors      map[string]int64

	Buckets [numBuckets]int64
}

// NewCollector initializes a fresh metrics collector.
func NewCollector() *Collector {
	return &Collector{
		StatusCodes: make(map[int]int64),
		Errors:      make(map[string]int64),
	}
}

func latencyToBucket(d time.Duration) int {
	if d <= 0 {
		return 0
	}
	micros := float64(d.Nanoseconds()) / 1000.0
	if micros < 1.0 {
		return 0
	}
	// log10(1) = 0, log10(100_000_000) = 8 (100 seconds)
	b := int(math.Log10(micros) * bucketsPerDecade)
	if b < 0 {
		return 0
	}
	if b >= numBuckets {
		return numBuckets - 1
	}
	return b
}

func bucketToLatency(bucket int) time.Duration {
	if bucket <= 0 {
		return time.Microsecond
	}
	micros := math.Pow(10, float64(bucket)/bucketsPerDecade)
	return time.Duration(micros * float64(time.Microsecond))
}

// Record adds a single HTTP transaction result to the collector.
func (c *Collector) Record(res Result) {
	c.TotalRequests++
	c.BytesRead += res.BytesRead

	if res.Error != nil || res.StatusCode >= 400 || res.StatusCode == 0 {
		c.Failed++
		if res.Error != nil {
			errCategory := categorizeError(res.Error)
			c.Errors[errCategory]++
		}
	} else {
		c.Successful++
	}

	if res.StatusCode > 0 {
		c.StatusCodes[res.StatusCode]++
	}

	if res.Latency > 0 {
		if c.MinLatency == 0 || res.Latency < c.MinLatency {
			c.MinLatency = res.Latency
		}
		if res.Latency > c.MaxLatency {
			c.MaxLatency = res.Latency
		}

		c.TotalLatency += res.Latency
		ns := float64(res.Latency.Nanoseconds())
		c.SumSqLatency += ns * ns

		b := latencyToBucket(res.Latency)
		c.Buckets[b]++
	}
}

// Merge combines another collector's aggregated metrics in O(1) time.
func (c *Collector) Merge(other *Collector) {
	if other == nil || other.TotalRequests == 0 {
		return
	}

	c.TotalRequests += other.TotalRequests
	c.Successful += other.Successful
	c.Failed += other.Failed
	c.BytesRead += other.BytesRead
	c.TotalLatency += other.TotalLatency
	c.SumSqLatency += other.SumSqLatency

	if other.MinLatency > 0 {
		if c.MinLatency == 0 || other.MinLatency < c.MinLatency {
			c.MinLatency = other.MinLatency
		}
	}
	if other.MaxLatency > c.MaxLatency {
		c.MaxLatency = other.MaxLatency
	}

	for code, count := range other.StatusCodes {
		c.StatusCodes[code] += count
	}

	for errType, count := range other.Errors {
		c.Errors[errType] += count
	}

	for i := 0; i < numBuckets; i++ {
		c.Buckets[i] += other.Buckets[i]
	}
}

// categorizeError groups network and protocol errors into clean human-readable labels.
func categorizeError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()

	// Unwrap url.Error if present
	if urlErr, ok := err.(*url.Error); ok {
		msg = urlErr.Err.Error()
	}

	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "context deadline exceeded") || strings.Contains(lower, "i/o timeout") || strings.Contains(lower, "client.timeout"):
		return "Timeout"
	case strings.Contains(lower, "context canceled"):
		return "Canceled"
	case strings.Contains(lower, "connection refused"):
		return "Connection Refused"
	case strings.Contains(lower, "connection reset by peer"):
		return "Connection Reset"
	case strings.Contains(lower, "no such host"):
		return "DNS Lookup Failed"
	case strings.Contains(lower, "certificate") || strings.Contains(lower, "tls:"):
		return "TLS Handshake Error"
	case strings.Contains(lower, "broken pipe"):
		return "Broken Pipe"
	case strings.Contains(lower, "too many open files"):
		return "Socket Exhaustion (Too many open files)"
	default:
		if len(msg) > 40 {
			return msg[:40] + "..."
		}
		return msg
	}
}