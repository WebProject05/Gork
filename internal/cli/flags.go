package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
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
	var bodyFile, inlineBody, stagesRaw string
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

	// Ramping and Multi-Stage Profiles
	fs.DurationVar(&cfg.RampUp, "ramp-up", 0, "Ramp-up duration from 1 to connections (e.g. 10s)")
	fs.DurationVar(&cfg.RampDown, "ramp-down", 0, "Ramp-down duration from connections to 1 (e.g. 5s)")
	fs.StringVar(&stagesRaw, "stages", "", "Multi-stage profile, e.g. '10s:10->50, 30s:50, 10s:50->0'")

	// Breakpoint / Step-Load Saturation Testing
	fs.BoolVar(&cfg.StepLoad, "step-load", false, "Enable automated step-load saturation / breakpoint testing")
	fs.IntVar(&cfg.StepConns, "step-conns", 10, "Concurrency increment per step (default: 10)")
	fs.DurationVar(&cfg.StepDuration, "step-duration", 5*time.Second, "Duration to hold each step (default: 5s)")
	fs.DurationVar(&cfg.MaxLatency, "max-latency", 500*time.Millisecond, "Saturation threshold: max acceptable P95 latency (default: 500ms)")
	fs.Float64Var(&cfg.MaxErrorRate, "max-error-rate", 1.0, "Saturation threshold: max acceptable error rate % (default: 1.0 = 1%)")
	fs.IntVar(&cfg.MaxConns, "max-conns", 1000, "Safety ceiling concurrency for step load (default: 1000)")

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

Ramping & Multi-Stage Profiles:
      --ramp-up <duration>      Ramp-up concurrency from 1 to connections over duration
      --ramp-down <duration>    Ramp-down concurrency from connections to 1 over duration
      --stages <string>         Multi-stage execution profile, e.g. '10s:10->50, 30s:50, 10s:50->0'

Breakpoint / Saturation Testing:
      --step-load               Enable automated step-load saturation / breakpoint testing
      --step-conns <int>        Concurrency increment per step (default: 10)
      --step-duration <dur>     Duration to hold each step (default: 5s)
      --max-latency <duration>  Stop threshold: max acceptable P95 latency (default: 500ms)
      --max-error-rate <float>  Stop threshold: max acceptable error rate % (default: 1.0)
      --max-conns <int>         Max concurrency ceiling for step load (default: 1000)

HTTP Request Options:
  -X, --method <string>         HTTP method (default: "GET")
  -H, --header <string>         HTTP header 'Key: Value' (can be repeated)
  -b, --body <string>           Inline request body string
      --body-file <path>        Path to file containing request body
      --timeout <duration>      Per-request timeout (default: 5s)
  -k, --insecure                Skip TLS certificate verification
      --no-keepalive            Disable HTTP Keep-Alive connection reuse
      --cpus <int>              Set GOMAXPROCS (default: number of CPUs)

Output Options:
      --json                    Output results in JSON format
  -o, --out <path>              Save output to file
  -v, --version                 Print version and exit
  -h, --help                    Show this help message

Examples:
  gork -c 100 -d 30s https://api.example.com/users
  gork -c 50 --ramp-up 10s -d 30s --ramp-down 5s https://api.example.com
  gork --stages "10s:10->50, 30s:50, 10s:50->0" https://api.example.com
  gork --step-load --step-conns 10 --step-duration 5s --max-latency 200ms https://api.example.com
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

	// Parse custom stages if provided
	if stagesRaw != "" {
		stages, err := parseStages(stagesRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid --stages format: %w", err)
		}
		cfg.Stages = stages
	} else if cfg.RampUp > 0 || cfg.RampDown > 0 {
		// Auto-construct stages from --ramp-up / -d / --ramp-down
		stageNum := 1
		if cfg.RampUp > 0 {
			cfg.Stages = append(cfg.Stages, config.StageConfig{
				Name:        fmt.Sprintf("Stage %d (Ramp Up 1->%d)", stageNum, cfg.Connections),
				Duration:    cfg.RampUp,
				StartConns:  1,
				TargetConns: cfg.Connections,
			})
			stageNum++
		}
		if cfg.Duration > 0 {
			cfg.Stages = append(cfg.Stages, config.StageConfig{
				Name:        fmt.Sprintf("Stage %d (Sustained %d)", stageNum, cfg.Connections),
				Duration:    cfg.Duration,
				StartConns:  cfg.Connections,
				TargetConns: cfg.Connections,
			})
			stageNum++
		}
		if cfg.RampDown > 0 {
			cfg.Stages = append(cfg.Stages, config.StageConfig{
				Name:        fmt.Sprintf("Stage %d (Ramp Down %d->1)", stageNum, cfg.Connections),
				Duration:    cfg.RampDown,
				StartConns:  cfg.Connections,
				TargetConns: 1,
			})
		}
	}

	// Update overall duration if custom stages are active
	if len(cfg.Stages) > 0 {
		var totalStageDuration time.Duration
		maxConns := 0
		for _, s := range cfg.Stages {
			totalStageDuration += s.Duration
			if s.StartConns > maxConns {
				maxConns = s.StartConns
			}
			if s.TargetConns > maxConns {
				maxConns = s.TargetConns
			}
		}
		cfg.Duration = totalStageDuration
		if maxConns > cfg.Connections {
			cfg.Connections = maxConns
		}
	}

	return cfg, Validate(cfg)
}

// parseStages parses a comma-separated stage specification into []config.StageConfig.
// Syntax examples: "10s:10->50, 30s:50, 10s:50->0" or "10s:20, 20s:50"
func parseStages(stagesStr string) ([]config.StageConfig, error) {
	rawParts := strings.Split(stagesStr, ",")
	var stages []config.StageConfig

	for i, part := range rawParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		colonIdx := strings.Index(part, ":")
		if colonIdx == -1 {
			return nil, fmt.Errorf("stage %d %q: missing colon separator between duration and concurrency (format: <duration>:<conns> or <duration>:<start>-><target>)", i+1, part)
		}

		durStr := strings.TrimSpace(part[:colonIdx])
		connsStr := strings.TrimSpace(part[colonIdx+1:])

		dur, err := time.ParseDuration(durStr)
		if err != nil || dur <= 0 {
			return nil, fmt.Errorf("stage %d %q: invalid stage duration %q", i+1, part, durStr)
		}

		var startConns, targetConns int
		if strings.Contains(connsStr, "->") {
			arrowParts := strings.Split(connsStr, "->")
			if len(arrowParts) != 2 {
				return nil, fmt.Errorf("stage %d %q: invalid ramp syntax %q", i+1, part, connsStr)
			}
			sVal, err1 := strconv.Atoi(strings.TrimSpace(arrowParts[0]))
			tVal, err2 := strconv.Atoi(strings.TrimSpace(arrowParts[1]))
			if err1 != nil || err2 != nil || sVal < 0 || tVal < 0 || (sVal == 0 && tVal == 0) {
				return nil, fmt.Errorf("stage %d %q: concurrency values must be non-negative integers", i+1, part)
			}
			startConns = sVal
			targetConns = tVal
		} else {
			val, err := strconv.Atoi(connsStr)
			if err != nil || val <= 0 {
				return nil, fmt.Errorf("stage %d %q: concurrency must be a positive integer", i+1, part)
			}
			startConns = val
			targetConns = val
		}

		name := fmt.Sprintf("Stage %d", i+1)
		if startConns != targetConns {
			name += fmt.Sprintf(" (Ramp %d->%d)", startConns, targetConns)
		} else {
			name += fmt.Sprintf(" (Hold %d)", targetConns)
		}

		stages = append(stages, config.StageConfig{
			Name:        name,
			Duration:    dur,
			StartConns:  startConns,
			TargetConns: targetConns,
		})
	}

	if len(stages) == 0 {
		return nil, fmt.Errorf("stages string contained no valid stages")
	}

	return stages, nil
}
