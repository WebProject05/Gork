package output

import (
	"fmt"
	"sort"

	"gork/internal/config"
	"gork/internal/metrics"
)

func PrintTerminal(cfg *config.Config, s *metrics.Summary) string {
	out := fmt.Sprintf("Running %s test @ %s\n", cfg.Duration, cfg.URL)
	out += fmt.Sprintf("  %d threads and %d connections\n\n", cfg.Threads, cfg.Connections)

	out += "Latency\n"
	out += fmt.Sprintf("  Min      %v\n", s.MinLatency)
	out += fmt.Sprintf("  Avg      %v\n", s.AvgLatency)
	out += fmt.Sprintf("  Max      %v\n", s.MaxLatency)
	out += fmt.Sprintf("  P50      %v\n", s.P50Latency)
	out += fmt.Sprintf("  P95      %v\n", s.P95Latency)
	out += fmt.Sprintf("  P99      %v\n\n", s.P99Latency)

	out += fmt.Sprintf("Requests/sec:  %.2f\n\n", s.RequestsPerSec)

	out += fmt.Sprintf("Total requests: %d\n", s.TotalRequests)
	out += fmt.Sprintf("Successful:     %d\n", s.Successful)
	out += fmt.Sprintf("Failed:         %d\n\n", s.Failed)

	out += "Status codes:\n"
	
	// Sort status codes for deterministic output
	var codes []int
	for code := range s.StatusCodes {
		codes = append(codes, code)
	}
	sort.Ints(codes)

	for _, code := range codes {
		out += fmt.Sprintf("  %d: %d\n", code, s.StatusCodes[code])
	}

	return out
}