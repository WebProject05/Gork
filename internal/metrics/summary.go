package metrics

import (
	"sort"
	"time"
)

type Summary struct {
	TotalRequests  int
	Successful     int
	Failed         int
	Duration       time.Duration
	RequestsPerSec float64

	MinLatency time.Duration
	AvgLatency time.Duration
	MaxLatency time.Duration
	P50Latency time.Duration
	P95Latency time.Duration
	P99Latency time.Duration

	StatusCodes map[int]int
}

func (c *Collector) CalculateSummary(duration time.Duration) *Summary {
	s := &Summary{
		TotalRequests: len(c.results),
		Duration:      duration,
		StatusCodes:   make(map[int]int),
	}

	if s.TotalRequests == 0 {
		return s
	}

	s.RequestsPerSec = float64(s.TotalRequests) / duration.Seconds()

	var latencies []time.Duration
	var totalLatency time.Duration

	for _, r := range c.results {
		if r.Error != nil || r.StatusCode >= 400 {
			s.Failed++
		} else {
			s.Successful++
		}

		if r.StatusCode > 0 {
			s.StatusCodes[r.StatusCode]++
		}

		if r.Latency > 0 {
			latencies = append(latencies, r.Latency)
			totalLatency += r.Latency
		}
	}
	if len(latencies) > 0 {
		sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

		s.MinLatency = latencies[0]
		s.MaxLatency = latencies[len(latencies)-1]
		s.AvgLatency = totalLatency / time.Duration(len(latencies))
		s.P50Latency = latencies[int(float64(len(latencies))*0.50)]
		s.P95Latency = latencies[int(float64(len(latencies))*0.95)]
		s.P99Latency = latencies[int(float64(len(latencies))*0.99)]
	}

	return s
}
