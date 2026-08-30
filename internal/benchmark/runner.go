package benchmark

import (
	"context"
	"net/http"
	"runtime"
	"sync"
	"time"

	"gork/internal/config"
	"gork/internal/httpclient"
	"gork/internal/metrics"
)


// Run executes the benchmark with default background context and returns collected metrics.
func Run(cfg *config.Config) *metrics.Summary {
	return RunWithContext(context.Background(), cfg)
}

// RunWithContext executes the benchmark with support for external cancellation (e.g. Ctrl+C).
func RunWithContext(parentCtx context.Context, cfg *config.Config) *metrics.Summary {
	// Configure OS thread parallelism
	if cfg.CPUs > 0 {
		runtime.GOMAXPROCS(cfg.CPUs)
	} else if cfg.Threads > 0 {
		runtime.GOMAXPROCS(cfg.Threads)
	}

	client := httpclient.NewClient(cfg)
	reqBuilder := httpclient.NewBuilder(cfg)

	// Calculate rate limit per worker connection if enabled
	var workerRateInterval time.Duration
	if cfg.Rate > 0 && cfg.Connections > 0 {
		intervalNs := int64(time.Second) * int64(cfg.Connections) / int64(cfg.Rate)
		if intervalNs > 0 {
			workerRateInterval = time.Duration(intervalNs)
		}
	}

	// Warmup Phase
	if cfg.Warmup > 0 {
		warmupCtx, cancelWarmup := context.WithTimeout(parentCtx, cfg.Warmup)
		runWorkers(warmupCtx, cfg, client, reqBuilder, workerRateInterval)
		cancelWarmup()
	}

	// Benchmark Phase
	benchCtx, cancelBench := context.WithTimeout(parentCtx, cfg.Duration)
	defer cancelBench()

	start := time.Now()
	collector := runWorkers(benchCtx, cfg, client, reqBuilder, workerRateInterval)
	totalDuration := time.Since(start)

	return collector.CalculateSummary(totalDuration)
}

// runWorkers coordinates concurrent worker execution and merges results.
func runWorkers(ctx context.Context, cfg *config.Config, client *http.Client, reqBuilder *httpclient.Builder, rateInterval time.Duration) *metrics.Collector {
	collector := metrics.NewCollector()

	threads := cfg.Threads
	if threads > cfg.Connections {
		threads = cfg.Connections
	}
	if threads <= 0 {
		threads = 1
	}

	connsPerThread := cfg.Connections / threads
	remainder := cfg.Connections % threads

	resultsCh := make(chan *metrics.Collector, cfg.Connections)
	var wg sync.WaitGroup
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
					workerCollector := runWorker(ctx, client, reqBuilder, rateInterval)
					resultsCh <- workerCollector
				}()
			}

			threadWg.Wait()
		}(connsForThisThread)
	}

	// Close results channel once all workers finish
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Merge all worker collectors
	for workerCol := range resultsCh {
		collector.Merge(workerCol)
	}

	return collector
}
