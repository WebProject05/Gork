package cli

import (
	"flag"
	"fmt"
	"gork/internal/config"
	"gork/internal/version"
	"os"
	"time"
)

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

// Parsing CLI arguments into a Config Struct
func Parse() (*config.Config, error) {
	cfg := &config.Config{}
	var headers stringSlice
	var bodyFile, inlineBody string
	var showVersion bool

	// Defining the Flags
	flag.IntVar(&cfg.Threads, "t", 2, "Number of worker threads")
	flag.IntVar(&cfg.Threads, "threads", 2, "Number of worker threads")

	flag.IntVar(&cfg.Connections, "c", 10, "Number of concurrent connections")
	flag.IntVar(&cfg.Connections, "connections", 10, "Number of concurrent connections")

	flag.DurationVar(&cfg.Duration, "d", 10*time.Second, "Duration of benchmark (e.g., 30s)")
	flag.DurationVar(&cfg.Duration, "duration", 10*time.Second, "Duration of benchmark")

	flag.StringVar(&cfg.Method, "X", "GET", "HTTP method")
	flag.StringVar(&cfg.Method, "method", "GET", "HTTP method")

	flag.Var(&headers, "H", "HTTP Headers (can be repeated)")
	flag.Var(&headers, "header", "HTTP Headers (can be repeated)")

	flag.StringVar(&inlineBody, "b", "", "Inline request body")
	flag.StringVar(&inlineBody, "body", "", "Inline request body")
	flag.StringVar(&bodyFile, "body-file", "", "Path to file containing request body")

	flag.DurationVar(&cfg.Timeout, "timeout", 5*time.Second, "Per-request timeout")

	flag.BoolVar(&cfg.JSONOutput, "json", false, "Output results in JSON format")
	flag.StringVar(&cfg.OutFile, "out", "", "Save output to file")

	flag.BoolVar(&showVersion, "version", false, "Print version and exit")

	// Custom usage helper
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: gork [options] <url> \n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if showVersion {
		fmt.Printf("gork v%s\n:", version.Version)
		os.Exit(0)
	}

	cfg.Headers = headers
	args := flag.Args()
	if len(args) > 0 {
		cfg.URL = args[len(args)-1]
	}

	// Body Params handling
	if inlineBody != "" && bodyFile != "" {
		return nil, fmt.Errorf("Cannot use both -b/--body and --body-file")
	}

	if inlineBody != "" {
		cfg.Body = []byte(inlineBody)
	} else if bodyFile != "" {
		b, err := os.ReadFile(bodyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read body file: %w", err)
		}
		cfg.Body = b
	}

	return cfg, Validate(cfg)
}
