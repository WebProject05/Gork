package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"gork/internal/config"
	"gork/internal/version"
)

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

// Parse parses CLI arguments from os.Args[1:] into a Config struct.
func Parse() (*config.Config, error) {
	return ParseArgs(os.Args[1:], os.Stdout, os.Stderr)
}

// ParseArgs parses arguments from a provided string slice, allowing unit testing.
func ParseArgs(args []string, stdout, stderr io.Writer) (*config.Config, error) {
	cfg := &config.Config{}
	var headers stringSlice
	var bodyFile, inlineBody string
	var showVersion, showHelp bool

	fs := flag.NewFlagSet("gork", flag.ContinueOnError)
	fs.SetOutput(stderr)

	// Short & Long Flag Bindings
	fs.IntVar(&cfg.Threads, "t", 2, "Number of worker threads")
	fs.IntVar(&cfg.Threads, "threads", 2, "Number of worker threads")

	fs.IntVar(&cfg.Connections, "c", 10, "Number of concurrent connections")
	fs.IntVar(&cfg.Connections, "connections", 10, "Number of concurrent connections")

	fs.DurationVar(&cfg.Duration, "d", 10*time.Second, "Duration of benchmark (e.g. 10s, 1m)")
	fs.DurationVar(&cfg.Duration, "duration", 10*time.Second, "Duration of benchmark")

	fs.DurationVar(&cfg.Warmup, "w", 0, "Warmup duration before recording metrics (e.g. 2s)")
	fs.DurationVar(&cfg.Warmup, "warmup", 0, "Warmup duration before recording metrics")

	fs.IntVar(&cfg.Rate, "r", 0, "Target requests per second (0 = unlimited)")
	fs.IntVar(&cfg.Rate, "rate", 0, "Target requests per second")

	fs.StringVar(&cfg.Method, "X", "GET", "HTTP method (GET, POST, PUT, DELETE, etc.)")
	fs.StringVar(&cfg.Method, "method", "GET", "HTTP method")

	fs.Var(&headers, "H", "HTTP header in 'Key: Value' format (repeatable)")
	fs.Var(&headers, "header", "HTTP header in 'Key: Value' format (repeatable)")

	fs.StringVar(&inlineBody, "b", "", "Inline request body string")
	fs.StringVar(&inlineBody, "body", "", "Inline request body string")
	fs.StringVar(&bodyFile, "body-file", "", "Path to file containing request body")

	fs.DurationVar(&cfg.Timeout, "timeout", 5*time.Second, "Per-request timeout")
	fs.BoolVar(&cfg.Insecure, "k", false, "Skip TLS certificate verification")
	fs.BoolVar(&cfg.Insecure, "insecure", false, "Skip TLS certificate verification")

	fs.BoolVar(&cfg.DisableKeepAlives, "no-keepalive", false, "Disable HTTP Keep-Alive connection reuse")
	fs.IntVar(&cfg.CPUs, "cpus", 0, "Set GOMAXPROCS (0 = default/threads)")

	fs.BoolVar(&cfg.JSONOutput, "json", false, "Output results in machine-readable JSON format")
	fs.StringVar(&cfg.OutFile, "o", "", "Save output to file")
	fs.StringVar(&cfg.OutFile, "out", "", "Save output to file")

	fs.BoolVar(&showVersion, "v", false, "Print version and exit")
	fs.BoolVar(&showVersion, "version", false, "Print version and exit")
	fs.BoolVar(&showHelp, "h", false, "Show help message")
	fs.BoolVar(&showHelp, "help", false, "Show help message")

	// Custom formatted usage helper
	fs.Usage = func() {
		usageText := `gork - A fast, wrk-style HTTP/HTTPS API benchmarking tool

Usage:
  gork [options] <url>

Options:
  -c, --connections <int>       Number of concurrent HTTP connections (default: 10)
  -t, --threads <int>           Number of worker threads (default: 2)
  -d, --duration <duration>     Duration of benchmark (e.g. 10s, 1m, 30s) (default: 10s)
  -w, --warmup <duration>       Warmup duration before recording metrics (default: 0s)
  -r, --rate <int>              Target requests per second (0 = unlimited) (default: 0)
  -X, --method <string>         HTTP method (default: "GET")
  -H, --header <string>         HTTP header 'Key: Value' (can be repeated)
  -b, --body <string>           Inline request body string
      --body-file <path>        Path to file containing request body
      --timeout <duration>      Per-request timeout (default: 5s)
  -k, --insecure                Skip TLS certificate verification
      --no-keepalive            Disable HTTP Keep-Alive connection reuse
      --cpus <int>              Set GOMAXPROCS (default: number of CPUs)
      --json                    Output results in JSON format
  -o, --out <path>              Save output to file
  -v, --version                 Print version and exit
  -h, --help                    Show this help message

Examples:
  gork -c 100 -d 30s https://api.example.com/users
  gork -c 50 -d 10s -X POST -H "Content-Type: application/json" -b '{"name":"benchmark"}' http://localhost:8080/api
  gork -c 20 -d 1m -r 500 --json -o results.json https://api.example.com
`
		fmt.Fprint(stderr, usageText)
	}

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if showHelp {
		fs.Usage()
		return nil, flag.ErrHelp
	}

	if showVersion {
		fmt.Fprintf(stdout, "gork v%s\n", version.Version)
		return nil, flag.ErrHelp
	}

	cfg.Headers = headers
	parsedArgs := fs.Args()
	if len(parsedArgs) > 0 {
		cfg.URL = parsedArgs[len(parsedArgs)-1]
	}

	// Body parameters mutual exclusivity
	if inlineBody != "" && bodyFile != "" {
		return nil, fmt.Errorf("cannot use both -b/--body and --body-file")
	}

	if inlineBody != "" {
		cfg.Body = []byte(inlineBody)
	} else if bodyFile != "" {
		b, err := os.ReadFile(bodyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read body file %q: %w", bodyFile, err)
		}
		cfg.Body = b
	}

	return cfg, Validate(cfg)
}
