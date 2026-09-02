package output

import (
	"encoding/json"
	"time"

	"gork/internal/metrics"
)

// JSONPhaseDetails holds timing for an HTTP lifecycle phase in milliseconds.
type JSONPhaseDetails struct {
	MinMs float64 `json:"min_ms"`
	AvgMs float64 `json:"avg_ms"`
	MaxMs float64 `json:"max_ms"`
}

// JSONReport represents the structured machine-readable benchmark report with extended telemetry.
type JSONReport struct {
	TotalRequests  int64              `json:"total_requests"`
	Successful     int64              `json:"successful"`
	Failed         int64              `json:"failed"`
	DurationSec    float64            `json:"duration_seconds"`
	RequestsPerSec float64            `json:"requests_per_sec"`
	LatencyMs      map[string]float64 `json:"latency_ms"`
	LatencyReadable map[string]string `json:"latency_readable"`

	// HTTP Timing Lifecycle Phases
	Phases map[string]JSONPhaseDetails `json:"phases_ms"`

	// Connection Health & Keep-Alive
	Connections metrics.ConnectionSummary `json:"connections"`

	// Data Transfer & Throughput
	Transfer struct {
		BytesSent      int64   `json:"bytes_sent"`
		BytesRead      int64   `json:"bytes_read"`
		TotalBytes     int64   `json:"total_bytes"`
		UploadRate     float64 `json:"upload_bytes_per_sec"`
		DownloadRate   float64 `json:"download_bytes_per_sec"`
		UploadRateFmt  string  `json:"upload_rate"`
		DownloadRateFmt string `json:"download_rate"`
		MinBodyBytes   int64   `json:"min_body_bytes"`
		AvgBodyBytes   int64   `json:"avg_body_bytes"`
		MaxBodyBytes   int64   `json:"max_body_bytes"`
	} `json:"transfer"`

	// Availability & Status Classes
	Availability metrics.AvailabilitySummary `json:"availability"`

	// Multi-Stage & Breakpoint Telemetry
	Stages     []metrics.StageSummary     `json:"stages,omitempty"`
	Breakpoint *metrics.BreakpointSummary `json:"breakpoint,omitempty"`

	StatusCodes map[int]int64    `json:"status_codes"`
	Errors      map[string]int64 `json:"errors,omitempty"`
}

// PrintJSON converts the summary to machine-readable indented JSON.
func PrintJSON(s *metrics.Summary) (string, error) {
	durationSec := s.Duration.Seconds()

	toMs := func(d time.Duration) float64 {
		return float64(d.Nanoseconds()) / 1e6
	}

	phaseDetails := func(p metrics.PhaseSummary) JSONPhaseDetails {
		return JSONPhaseDetails{
			MinMs: toMs(p.Min),
			AvgMs: toMs(p.Avg),
			MaxMs: toMs(p.Max),
		}
	}

	report := JSONReport{
		TotalRequests:  s.TotalRequests,
		Successful:     s.Successful,
		Failed:         s.Failed,
		DurationSec:    durationSec,
		RequestsPerSec: s.RequestsPerSec,
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
		Phases: map[string]JSONPhaseDetails{
			"dns":     phaseDetails(s.Phases.DNS),
			"tcp":     phaseDetails(s.Phases.TCP),
			"tls":     phaseDetails(s.Phases.TLS),
			"ttfb":    phaseDetails(s.Phases.TTFB),
			"content": phaseDetails(s.Phases.Content),
		},
		Connections: s.Connections,
		Availability: s.Availability,
		Stages:       s.Stages,
		Breakpoint:   s.Breakpoint,
		StatusCodes:  s.StatusCodes,
		Errors:       s.Errors,
	}

	report.Transfer.BytesSent = s.Transfer.BytesSent
	report.Transfer.BytesRead = s.Transfer.BytesRead
	report.Transfer.TotalBytes = s.Transfer.TotalBytes
	report.Transfer.UploadRate = s.Transfer.UploadRate
	report.Transfer.DownloadRate = s.Transfer.DownloadRate
	report.Transfer.UploadRateFmt = formatBytes(int64(s.Transfer.UploadRate)) + "/s"
	report.Transfer.DownloadRateFmt = formatBytes(int64(s.Transfer.DownloadRate)) + "/s"
	report.Transfer.MinBodyBytes = s.Transfer.MinBodyBytes
	report.Transfer.AvgBodyBytes = s.Transfer.AvgBodyBytes
	report.Transfer.MaxBodyBytes = s.Transfer.MaxBodyBytes

	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
