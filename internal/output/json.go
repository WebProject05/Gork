package output

import (
	"encoding/json"
	"time"

	"gork/internal/metrics"
)

// JSONReport represents the structured machine-readable benchmark report.
type JSONReport struct {
	TotalRequests  int64            `json:"total_requests"`
	Successful     int64            `json:"successful"`
	Failed         int64            `json:"failed"`
	DurationSec    float64          `json:"duration_seconds"`
	RequestsPerSec float64          `json:"requests_per_sec"`
	BytesRead      int64            `json:"bytes_read"`
	BytesPerSec    float64          `json:"bytes_per_sec"`
	TransferRate   string           `json:"transfer_rate"`
	LatencyMs      map[string]float64 `json:"latency_ms"`
	LatencyReadable map[string]string `json:"latency_readable"`
	StatusCodes    map[int]int64    `json:"status_codes"`
	Errors         map[string]int64 `json:"errors,omitempty"`
}

// PrintJSON converts the summary to machine-readable indented JSON.
func PrintJSON(s *metrics.Summary) (string, error) {
	durationSec := s.Duration.Seconds()

	toMs := func(d time.Duration) float64 {
		return float64(d.Nanoseconds()) / 1e6
	}

	report := JSONReport{
		TotalRequests:  s.TotalRequests,
		Successful:     s.Successful,
		Failed:         s.Failed,
		DurationSec:    durationSec,
		RequestsPerSec: s.RequestsPerSec,
		BytesRead:      s.BytesRead,
		BytesPerSec:    s.BytesPerSec,
		TransferRate:   formatBytes(int64(s.BytesPerSec)) + "/s",
		LatencyMs: map[string]float64{
			"min":    toMs(s.Latency.Min),
			"avg":    toMs(s.Latency.Avg),
			"max":    toMs(s.Latency.Max),
			"stddev": toMs(s.Latency.StdDev),
			"p50":    toMs(s.Latency.P50),
			"p75":    toMs(s.Latency.P75),
			"p90":    toMs(s.Latency.P90),
			"p95":    toMs(s.Latency.P95),
			"p99":    toMs(s.Latency.P99),
			"p99_9":  toMs(s.Latency.P999),
		},
		LatencyReadable: map[string]string{
			"min":    formatDuration(s.Latency.Min),
			"avg":    formatDuration(s.Latency.Avg),
			"max":    formatDuration(s.Latency.Max),
			"stddev": formatDuration(s.Latency.StdDev),
			"p50":    formatDuration(s.Latency.P50),
			"p75":    formatDuration(s.Latency.P75),
			"p90":    formatDuration(s.Latency.P90),
			"p95":    formatDuration(s.Latency.P95),
			"p99":    formatDuration(s.Latency.P99),
			"p99_9":  formatDuration(s.Latency.P999),
		},
		StatusCodes: s.StatusCodes,
		Errors:      s.Errors,
	}

	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}