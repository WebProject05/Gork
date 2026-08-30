package output

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"gork/internal/config"
	"gork/internal/metrics"
)

// PrintTerminal formats the benchmark summary into a clean, human-readable terminal report.
func PrintTerminal(cfg *config.Config, s *metrics.Summary) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("\nRunning %s test @ %s\n", cfg.Duration, cfg.URL))
	b.WriteString(fmt.Sprintf("  %d thread(s) and %d connection(s)\n", cfg.Threads, cfg.Connections))
	if cfg.Rate > 0 {
		b.WriteString(fmt.Sprintf("  Target rate: %d req/sec\n", cfg.Rate))
	}
	if cfg.Warmup > 0 {
		b.WriteString(fmt.Sprintf("  Warmup: %s\n", cfg.Warmup))
	}
	b.WriteString("\n")

	b.WriteString("Latency Statistics:\n")
	b.WriteString(fmt.Sprintf("  Min:       %s\n", formatDuration(s.Latency.Min)))
	b.WriteString(fmt.Sprintf("  Avg:       %s\n", formatDuration(s.Latency.Avg)))
	b.WriteString(fmt.Sprintf("  Max:       %s\n", formatDuration(s.Latency.Max)))
	b.WriteString(fmt.Sprintf("  StdDev:    %s\n\n", formatDuration(s.Latency.StdDev)))

	b.WriteString("Latency Percentiles:\n")
	b.WriteString(fmt.Sprintf("  50%% (P50):  %s\n", formatDuration(s.Latency.P50)))
	b.WriteString(fmt.Sprintf("  75%% (P75):  %s\n", formatDuration(s.Latency.P75)))
	b.WriteString(fmt.Sprintf("  90%% (P90):  %s\n", formatDuration(s.Latency.P90)))
	b.WriteString(fmt.Sprintf("  95%% (P95):  %s\n", formatDuration(s.Latency.P95)))
	b.WriteString(fmt.Sprintf("  99%% (P99):  %s\n", formatDuration(s.Latency.P99)))
	b.WriteString(fmt.Sprintf("  99.9%%:     %s\n\n", formatDuration(s.Latency.P999)))

	b.WriteString("Throughput:\n")
	b.WriteString(fmt.Sprintf("  Requests/sec:  %.2f\n", s.RequestsPerSec))
	b.WriteString(fmt.Sprintf("  Transfer/sec:  %s/s\n", formatBytes(int64(s.BytesPerSec))))
	b.WriteString(fmt.Sprintf("  Total Read:    %s\n\n", formatBytes(s.BytesRead)))

	b.WriteString("Requests:\n")
	b.WriteString(fmt.Sprintf("  Total:         %d\n", s.TotalRequests))
	b.WriteString(fmt.Sprintf("  Successful:    %d\n", s.Successful))
	b.WriteString(fmt.Sprintf("  Failed:        %d\n\n", s.Failed))

	if len(s.StatusCodes) > 0 {
		b.WriteString("Status Codes:\n")
		var codes []int
		for code := range s.StatusCodes {
			codes = append(codes, code)
		}
		sort.Ints(codes)

		for _, code := range codes {
			statusText := http.StatusText(code)
			if statusText != "" {
				b.WriteString(fmt.Sprintf("  [%d %s]: %d\n", code, statusText, s.StatusCodes[code]))
			} else {
				b.WriteString(fmt.Sprintf("  [%d]: %d\n", code, s.StatusCodes[code]))
			}
		}
		b.WriteString("\n")
	}

	if len(s.Errors) > 0 {
		b.WriteString("Error Breakdown:\n")
		var errTypes []string
		for errType := range s.Errors {
			errTypes = append(errTypes, errType)
		}
		sort.Strings(errTypes)

		for _, errType := range errTypes {
			b.WriteString(fmt.Sprintf("  %s: %d\n", errType, s.Errors[errType]))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(d.Nanoseconds())/1000.0)
	}
	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Nanoseconds())/1000000.0)
	}
	return fmt.Sprintf("%.3fs", d.Seconds())
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}