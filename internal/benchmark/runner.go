package benchmark

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
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
		runStage(warmupCtx, cfg.Threads, config.StageConfig{
			Name:        "Warmup",
			Duration:    cfg.Warmup,
			StartConns:  cfg.Connections,
			TargetConns: cfg.Connections,
		}, client, reqBuilder, workerRateInterval)
		cancelWarmup()
	}

	// Step-load / Breakpoint Saturation Mode
	if cfg.StepLoad {
		return runStepLoad(parentCtx, cfg, client, reqBuilder, workerRateInterval)
	}

	// Staged / Ramping Profiles
	if len(cfg.Stages) > 0 {
		return runStages(parentCtx, cfg, client, reqBuilder, workerRateInterval)
	}

	// Standard Benchmark Run
	stage := config.StageConfig{
		Name:        "Standard",
		Duration:    cfg.Duration,
		StartConns:  cfg.Connections,
		TargetConns: cfg.Connections,
	}

	benchCtx, cancelBench := context.WithTimeout(parentCtx, cfg.Duration)
	defer cancelBench()

	start := time.Now()
	collector := runStage(benchCtx, cfg.Threads, stage, client, reqBuilder, workerRateInterval)
	totalDuration := time.Since(start)

	return collector.CalculateSummary(totalDuration)
}

// runStages executes multiple stages sequentially, collecting stage-by-stage and overall telemetry.
func runStages(parentCtx context.Context, cfg *config.Config, client *http.Client, reqBuilder *httpclient.Builder, rateInterval time.Duration) *metrics.Summary {
	overallCollector := metrics.NewCollector()
	var stageSummaries []metrics.StageSummary

	startTotal := time.Now()

	for _, stage := range cfg.Stages {
		if parentCtx.Err() != nil {
			break
		}

		stageCtx, cancelStage := context.WithTimeout(parentCtx, stage.Duration)
		startStage := time.Now()
		stageCol := runStage(stageCtx, cfg.Threads, stage, client, reqBuilder, rateInterval)
		stageDuration := time.Since(startStage)
		cancelStage()

		stageSummary := stageCol.CalculateSummary(stageDuration)

		concurrencyLabel := fmt.Sprintf("%d", stage.TargetConns)
		if stage.StartConns != stage.TargetConns {
			concurrencyLabel = fmt.Sprintf("%d->%d", stage.StartConns, stage.TargetConns)
		}

		stageSummaries = append(stageSummaries, metrics.StageSummary{
			Name:        stage.Name,
			Duration:    stageDuration,
			Concurrency: concurrencyLabel,
			Requests:    stageSummary.TotalRequests,
			RPS:         stageSummary.RequestsPerSec,
			AvgLatency:  stageSummary.Latency.Avg,
			P95Latency:  stageSummary.Latency.P95,
			Errors:      stageSummary.Failed,
		})

		overallCollector.Merge(stageCol)
	}

	totalDuration := time.Since(startTotal)
	summary := overallCollector.CalculateSummary(totalDuration)
	summary.Stages = stageSummaries
	return summary
}

// runStepLoad automatically steps up concurrency to identify the server breaking point.
func runStepLoad(parentCtx context.Context, cfg *config.Config, client *http.Client, reqBuilder *httpclient.Builder, rateInterval time.Duration) *metrics.Summary {
	overallCollector := metrics.NewCollector()
	var stageSummaries []metrics.StageSummary
	var breakpoint *metrics.BreakpointSummary

	startTotal := time.Now()
	stepNum := 1
	conns := cfg.StepConns
	if conns <= 0 {
		conns = 10
	}

	for conns <= cfg.MaxConns {
		if parentCtx.Err() != nil {
			break
		}

		stage := config.StageConfig{
			Name:        fmt.Sprintf("Step %d (%d conns)", stepNum, conns),
			Duration:    cfg.StepDuration,
			StartConns:  conns,
			TargetConns: conns,
		}

		stepCtx, cancelStep := context.WithTimeout(parentCtx, cfg.StepDuration)
		startStep := time.Now()
		stepCol := runStage(stepCtx, cfg.Threads, stage, client, reqBuilder, rateInterval)
		stepDuration := time.Since(startStep)
		cancelStep()

		stepSummary := stepCol.CalculateSummary(stepDuration)

		errorRatePct := 0.0
		if stepSummary.TotalRequests > 0 {
			errorRatePct = (float64(stepSummary.Failed) / float64(stepSummary.TotalRequests)) * 100.0
		}

		stageSummaries = append(stageSummaries, metrics.StageSummary{
			Name:        stage.Name,
			Duration:    stepDuration,
			Concurrency: fmt.Sprintf("%d", conns),
			Requests:    stepSummary.TotalRequests,
			RPS:         stepSummary.RequestsPerSec,
			AvgLatency:  stepSummary.Latency.Avg,
			P95Latency:  stepSummary.Latency.P95,
			Errors:      stepSummary.Failed,
		})

		overallCollector.Merge(stepCol)

		// Check Breakpoint thresholds
		if cfg.MaxLatency > 0 && stepSummary.Latency.P95 > cfg.MaxLatency {
			breakpoint = &metrics.BreakpointSummary{
				Triggered:         true,
				BreakingConns:     conns,
				BreakingRPS:       stepSummary.RequestsPerSec,
				BreakingLatency:   stepSummary.Latency.P95,
				BreakingErrorRate: errorRatePct,
				Reason:            fmt.Sprintf("P95 latency of %v exceeded threshold of %v", stepSummary.Latency.P95, cfg.MaxLatency),
			}
			break
		}

		if cfg.MaxErrorRate >= 0 && errorRatePct > cfg.MaxErrorRate {
			breakpoint = &metrics.BreakpointSummary{
				Triggered:         true,
				BreakingConns:     conns,
				BreakingRPS:       stepSummary.RequestsPerSec,
				BreakingLatency:   stepSummary.Latency.P95,
				BreakingErrorRate: errorRatePct,
				Reason:            fmt.Sprintf("Error rate of %.2f%% exceeded threshold of %.2f%%", errorRatePct, cfg.MaxErrorRate),
			}
			break
		}

		stepNum++
		conns += cfg.StepConns
	}

	totalDuration := time.Since(startTotal)
	summary := overallCollector.CalculateSummary(totalDuration)
	summary.Stages = stageSummaries
	summary.Breakpoint = breakpoint
	return summary
}

// runStage executes a single stage with dynamic worker concurrency management.
func runStage(ctx context.Context, numThreads int, stage config.StageConfig, client *http.Client, reqBuilder *httpclient.Builder, rateInterval time.Duration) *metrics.Collector {
	collector := metrics.NewCollector()

	maxConns := stage.TargetConns
	if stage.StartConns > maxConns {
		maxConns = stage.StartConns
	}
	if maxConns <= 0 {
		maxConns = 1
	}

	threads := numThreads
	if threads > maxConns {
		threads = maxConns
	}
	if threads <= 0 {
		threads = 1
	}

	activeConns := &atomic.Int64{}
	activeConns.Store(int64(stage.StartConns))

	// Dynamic worker scaling routine for linear ramping stages
	if stage.StartConns != stage.TargetConns {
		go func() {
			ticker := time.NewTicker(50 * time.Millisecond)
			defer ticker.Stop()
			start := time.Now()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					progress := float64(time.Since(start)) / float64(stage.Duration)
					if progress > 1.0 {
						progress = 1.0
					}
					cur := float64(stage.StartConns) + float64(stage.TargetConns-stage.StartConns)*progress
					target := int64(math.Round(cur))
					if target < 1 {
						target = 1
					}
					activeConns.Store(target)
				}
			}
		}()
	}

	connsPerThread := maxConns / threads
	remainder := maxConns % threads

	resultsCh := make(chan *metrics.Collector, maxConns)
	var wg sync.WaitGroup
	wg.Add(threads)

	workerIndexCounter := 0
	for t := 0; t < threads; t++ {
		connsForThisThread := connsPerThread
		if t == threads-1 {
			connsForThisThread += remainder
		}

		baseIndex := workerIndexCounter
		workerIndexCounter += connsForThisThread

		go func(numConns, startIdx int) {
			defer wg.Done()

			var threadWg sync.WaitGroup
			threadWg.Add(numConns)

			for c := 0; c < numConns; c++ {
				idx := startIdx + c
				go func(wIdx int) {
					defer threadWg.Done()
					workerCol := runWorker(ctx, client, reqBuilder, rateInterval, wIdx, activeConns)
					resultsCh <- workerCol
				}(idx)
			}

			threadWg.Wait()
		}(connsForThisThread, baseIndex)
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

