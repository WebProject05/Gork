package benchmark

import (
	"context"
	"io"
	"net/http"
	"time"

	"gork/internal/httpclient"
	"gork/internal/metrics"
)

// runWorker executes requests concurrently and records metrics into a local lock-free collector.
func runWorker(ctx context.Context, client *http.Client, builder *httpclient.Builder, rateInterval time.Duration) *metrics.Collector {
	collector := metrics.NewCollector()

	var ticker *time.Ticker
	if rateInterval > 0 {
		ticker = time.NewTicker(rateInterval)
		defer ticker.Stop()
	}

	for {
		if ticker != nil {
			select {
			case <-ctx.Done():
				return collector
			case <-ticker.C:
			}
		} else {
			select {
			case <-ctx.Done():
				return collector
			default:
			}
		}

		req, err := builder.BuildWithContext(ctx)
		if err != nil {
			collector.Record(metrics.Result{Error: err})
			continue
		}

		start := time.Now()
		resp, err := client.Do(req)
		latency := time.Since(start)

		if err != nil {
			// If context was canceled or deadline exceeded during request execution,
			// the benchmark timer has expired. Do not record as a server error.
			if ctx.Err() != nil {
				return collector
			}
			collector.Record(metrics.Result{
				Error:   err,
				Latency: latency,
			})
			continue
		}

		// Drain response body to EOF so the underlying TCP connection can be returned to the pool
		bytesRead, _ := io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		collector.Record(metrics.Result{
			StatusCode: resp.StatusCode,
			Latency:    latency,
			BytesRead:  bytesRead,
		})
	}
}
