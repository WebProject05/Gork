package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gork/internal/benchmark"
	"gork/internal/cli"
	"gork/internal/output"
)

func main() {
	cfg, err := cli.Parse()
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Trap OS interrupt signals (Ctrl+C) for graceful benchmark cancellation
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	summary := benchmark.RunWithContext(ctx, cfg)

	var resultString string
	if cfg.JSONOutput {
		resultString, err = output.PrintJSON(summary)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating JSON output: %v\n", err)
			os.Exit(1)
		}
	} else {
		resultString = output.PrintTerminal(cfg, summary)
	}

	// Output routing
	if cfg.OutFile != "" {
		if err := os.WriteFile(cfg.OutFile, []byte(resultString), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to file %s: %v\n", cfg.OutFile, err)
			os.Exit(1)
		}
		if !cfg.JSONOutput {
			fmt.Printf("Benchmark complete. Results written to %s\n", cfg.OutFile)
		}
	} else {
		fmt.Print(resultString)
	}
}