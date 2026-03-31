package output

import (
	"encoding/json"

	"gork/internal/metrics"
)

// PrintJSON converts the summary to machine-readable JSON.
func PrintJSON(s *metrics.Summary) (string, error) {
	// Use an anonymous struct to format latencies neatly in string formats
	data := struct {
		TotalRequests  int            `json:"total_requests"`
		Successful     int            `json:"successful"`
		Failed         int            `json:"failed"`
		DurationSec    float64        `json:"duration_seconds"`
		RequestsPerSec float64        `json:"requests_per_sec"`
		Latency        map[string]any `json:"latency_ms"`
		StatusCodes    map[int]int    `json:"status_codes"`
	}{
		TotalRequests:  s.TotalRequests,
		Successful:     s.Successful,
		Failed:         s.Failed,
		DurationSec:    s.Duration.Seconds(),
		RequestsPerSec: s.RequestsPerSec,
		Latency: map[string]any{
			"min": s.MinLatency.Milliseconds(),
			"avg": s.AvgLatency.Milliseconds(),
			"max": s.MaxLatency.Milliseconds(),
			"p50": s.P50Latency.Milliseconds(),
			"p95": s.P95Latency.Milliseconds(),
			"p99": s.P99Latency.Milliseconds(),
		},
		StatusCodes: s.StatusCodes,
	}

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}