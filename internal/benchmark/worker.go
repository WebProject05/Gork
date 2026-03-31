/*
Instead of thousands of workers fighting over a single sync.Mutex 
or spamming a channel after every single HTTP request, each 
worker will keep a local slice of its own results. When the 
benchmark duration ends, the worker will send its entire batch 
of results down a channel just once.
*/


package benchmark

import (
	"context"
	"net/http"
	"time"

	"gork/internal/httpclient"
	"gork/internal/metrics"
)

// runWorker now collects its own results locally and returns the batch when done.
func runWorker(ctx context.Context, client *http.Client, builder *httpclient.Builder) []metrics.Result {
	// Pre-allocate a reasonable capacity to avoid reallocation overhead during the run
	results := make([]metrics.Result, 0, 1000)

	for {
		select {
		case <-ctx.Done():
			// Time is up, return the local batch
			return results
		default:
			req, err := builder.Build()
			if err != nil {
				results = append(results, metrics.Result{Error: err})
				continue
			}

			start := time.Now()
			resp, err := client.Do(req.WithContext(ctx))
			duration := time.Since(start)

			if err != nil {
				results = append(results, metrics.Result{Error: err, Latency: duration})
				continue
			}

			resp.Body.Close()

			results = append(results, metrics.Result{
				StatusCode: resp.StatusCode,
				Latency:    duration,
			})
		}
	}
}