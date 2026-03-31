package main

import (
	"fmt"
	"os"

	"gork/internal/benchmark"
	"gork/internal/cli"
	"gork/internal/output"
)

func main() {
	cfg, err := cli.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing configuration: %v\n", err)
		os.Exit(1)
	}

	summary := benchmark.Run(cfg)

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
			fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
			os.Exit(1)
		}
		// If saving to file, also print standard terminal to stdout to keep user informed 
		// (unless JSON was requested, then we just wrote JSON to the file).
		if !cfg.JSONOutput {
			fmt.Println("Benchmark complete. Results written to", cfg.OutFile)
		}
	} else {
		fmt.Print(resultString)
	}
}