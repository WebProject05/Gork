package benchmark

import (
	"context"
	"sync"
	"time"

	"gork/internal/config"
	"gork/internal/httpclient"
	"gork/internal/metrics"
)

// Run executes the benchmark and returns the collected metrics.
func Run(cfg *config.Config) *metrics.Summary {
	client := httpclient.NewClient(cfg)
	reqBuilder := httpclient.NewBuilder(cfg)
	collector := metrics.NewCollector()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Duration)
	defer cancel()

	var wg sync.WaitGroup
	start := time.Now()

	threads := cfg.Threads
	if threads > cfg.Connections {
		threads = cfg.Connections
	}

	connsPerThread := cfg.Connections / threads
	remainder := cfg.Connections % threads

	// Channel to collect result batches from each worker
	resultsCh := make(chan []metrics.Result, cfg.Connections)

	wg.Add(threads)

	for t := 0; t < threads; t++ {
		connsForThisThread := connsPerThread
		if t == threads-1 {
			connsForThisThread += remainder
		}

		go func(numConns int) {
			defer wg.Done()

			var threadWg sync.WaitGroup
			threadWg.Add(numConns)

			for c := 0; c < numConns; c++ {
				go func() {
					defer threadWg.Done()
					// Worker runs and returns its local batch of results
					batch := runWorker(ctx, client, reqBuilder)
					resultsCh <- batch
				}()
			}

			threadWg.Wait()
		}(connsForThisThread)
	}

	// Wait for all workers to finish, then safely close the channel
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Drain the channel and merge all batches into the main collector
	for batch := range resultsCh {
		collector.AddBatch(batch)
	}

	totalDuration := time.Since(start)
	return collector.CalculateSummary(totalDuration)
}