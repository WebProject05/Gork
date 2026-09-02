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

	// 1. Breakpoint Saturation Banner (if triggered or run)
	if s.Breakpoint != nil {
		b.WriteString("================================================================================\n")
		if s.Breakpoint.Triggered {
			b.WriteString("🚨 BREAKPOINT SATURATION DETECTED\n")
			b.WriteString(fmt.Sprintf("  Breaking Concurrency: %d connections\n", s.Breakpoint.BreakingConns))
			b.WriteString(fmt.Sprintf("  Throughput at Peak:   %.2f req/sec\n", s.Breakpoint.BreakingRPS))
			b.WriteString(fmt.Sprintf("  P95 Latency at Peak:  %s\n", formatDuration(s.Breakpoint.BreakingLatency)))
			b.WriteString(fmt.Sprintf("  Error Rate at Peak:   %.2f%%\n", s.Breakpoint.BreakingErrorRate))
			b.WriteString(fmt.Sprintf("  Trigger Reason:       %s\n", s.Breakpoint.Reason))
		} else {
			b.WriteString("✅ STEP-LOAD COMPLETED WITHOUT SATURATION (Max ceiling reached)\n")
		}
		b.WriteString("================================================================================\n\n")
	}

	// 2. Stage Progression Table (if multi-stage test)
	if len(s.Stages) > 0 {
		b.WriteString("Stage Progression:\n")
		b.WriteString(fmt.Sprintf("  %-28s %-10s %-12s %-10s %-12s %-12s %-12s %s\n",
			"Stage", "Duration", "Concurrency", "Requests", "RPS", "Avg Latency", "P95 Latency", "Errors"))
		b.WriteString(fmt.Sprintf("  %s\n", strings.Repeat("-", 106)))
		for _, st := range s.Stages {
			b.WriteString(fmt.Sprintf("  %-28s %-10s %-12s %-10d %-12.2f %-12s %-12s %d\n",
				st.Name, st.Duration.Round(time.Millisecond), st.Concurrency, st.Requests, st.RPS,
				formatDuration(st.AvgLatency), formatDuration(st.P95Latency), st.Errors))
		}
		b.WriteString("\n")
	}

	// 3. Latency Statistics
	b.WriteString("Latency Statistics:\n")
	b.WriteString(fmt.Sprintf("  Min:       %s\n", formatDuration(s.Latency.Min)))
	b.WriteString(fmt.Sprintf("  Avg:       %s\n", formatDuration(s.Latency.Avg)))
	b.WriteString(fmt.Sprintf("  Max:       %s\n", formatDuration(s.Latency.Max)))
	b.WriteString(fmt.Sprintf("  StdDev:    %s\n\n", formatDuration(s.Latency.StdDev)))

	// 4. Latency Percentiles
	b.WriteString("Latency Percentiles:\n")
	b.WriteString(fmt.Sprintf("  50%% (P50):  %s\n", formatDuration(s.Latency.P50)))
	b.WriteString(fmt.Sprintf("  75%% (P75):  %s\n", formatDuration(s.Latency.P75)))
	b.WriteString(fmt.Sprintf("  90%% (P90):  %s\n", formatDuration(s.Latency.P90)))

	b.WriteString(fmt.Sprintf("  95%% (P95):  %s\n", formatDuration(s.Latency.P95)))
	b.WriteString(fmt.Sprintf("  99%% (P99):  %s\n", formatDuration(s.Latency.P99)))
	b.WriteString(fmt.Sprintf("  99.9%%:     %s\n\n", formatDuration(s.Latency.P999)))

	// 5. HTTP Timing Lifecycle Phase Breakdown
	if s.Phases.TTFB.Avg > 0 || s.Phases.DNS.Avg > 0 || s.Phases.TCP.Avg > 0 {
		b.WriteString("HTTP Lifecycle Phase Breakdown:\n")
		b.WriteString(fmt.Sprintf("  %-26s %-14s %-14s %s\n", "Phase", "Avg", "Min", "Max"))
		b.WriteString(fmt.Sprintf("  %s\n", strings.Repeat("-", 64)))
		if s.Phases.DNS.Avg > 0 {
			b.WriteString(fmt.Sprintf("  %-26s %-14s %-14s %s\n", "DNS Lookup:",
				formatDuration(s.Phases.DNS.Avg), formatDuration(s.Phases.DNS.Min), formatDuration(s.Phases.DNS.Max)))
		}
		if s.Phases.TCP.Avg > 0 {
			b.WriteString(fmt.Sprintf("  %-26s %-14s %-14s %s\n", "TCP Connect:",
				formatDuration(s.Phases.TCP.Avg), formatDuration(s.Phases.TCP.Min), formatDuration(s.Phases.TCP.Max)))
		}
		if s.Phases.TLS.Avg > 0 {
			b.WriteString(fmt.Sprintf("  %-26s %-14s %-14s %s\n", "TLS Handshake:",
				formatDuration(s.Phases.TLS.Avg), formatDuration(s.Phases.TLS.Min), formatDuration(s.Phases.TLS.Max)))
		}
		if s.Phases.TTFB.Avg > 0 {
			b.WriteString(fmt.Sprintf("  %-26s %-14s %-14s %s\n", "Server Processing (TTFB):",
				formatDuration(s.Phases.TTFB.Avg), formatDuration(s.Phases.TTFB.Min), formatDuration(s.Phases.TTFB.Max)))
		}
		if s.Phases.Content.Avg > 0 {
			b.WriteString(fmt.Sprintf("  %-26s %-14s %-14s %s\n", "Content Download:",
				formatDuration(s.Phases.Content.Avg), formatDuration(s.Phases.Content.Min), formatDuration(s.Phases.Content.Max)))
		}
		b.WriteString("\n")
	}

	// 6. Connection & Transport Health
	if s.Connections.TotalAttempts > 0 {
		b.WriteString("Connection & Keep-Alive Health:\n")
		b.WriteString(fmt.Sprintf("  Reuse Rate:    %.2f%%\n", s.Connections.ReusedPct))
		b.WriteString(fmt.Sprintf("  Reused Sockets:%d\n", s.Connections.Reused))
		b.WriteString(fmt.Sprintf("  New Sockets:   %d\n", s.Connections.New))
		b.WriteString(fmt.Sprintf("  Total Attempts:%d\n\n", s.Connections.TotalAttempts))
	}

	// 7. Bidirectional Throughput
	b.WriteString("Data Transfer & Throughput:\n")
	b.WriteString(fmt.Sprintf("  Requests/sec:  %.2f\n", s.RequestsPerSec))
	b.WriteString(fmt.Sprintf("  Download Rate: %s/s (Total: %s)\n",
		formatBytes(int64(s.Transfer.DownloadRate)), formatBytes(s.Transfer.BytesRead)))
	if s.Transfer.BytesSent > 0 {
		b.WriteString(fmt.Sprintf("  Upload Rate:   %s/s (Total: %s)\n",
			formatBytes(int64(s.Transfer.UploadRate)), formatBytes(s.Transfer.BytesSent)))
		b.WriteString(fmt.Sprintf("  Total Traffic: %s\n", formatBytes(s.Transfer.TotalBytes)))
	}
	if s.Transfer.MaxBodyBytes > 0 {
		b.WriteString(fmt.Sprintf("  Payload Size:  Min: %s | Avg: %s | Max: %s\n",
			formatBytes(s.Transfer.MinBodyBytes), formatBytes(s.Transfer.AvgBodyBytes), formatBytes(s.Transfer.MaxBodyBytes)))
	}
	b.WriteString("\n")

	// 8. Requests & Availability
	b.WriteString("Requests & Availability:\n")
	b.WriteString(fmt.Sprintf("  Total:         %d\n", s.TotalRequests))
	b.WriteString(fmt.Sprintf("  Success Rate:  %.2f%% (%d successful)\n", s.Availability.SuccessRate, s.Successful))
	b.WriteString(fmt.Sprintf("  Error Rate:    %.2f%% (%d failed)\n", s.Availability.ErrorRate, s.Failed))
	b.WriteString(fmt.Sprintf("  Status Classes:2xx: %d | 3xx: %d | 4xx: %d | 5xx: %d\n\n",
		s.Availability.Count2xx, s.Availability.Count3xx, s.Availability.Count4xx, s.Availability.Count5xx))

	// 9. Status Codes
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

	// 10. Errors
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
