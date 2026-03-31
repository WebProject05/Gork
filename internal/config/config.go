package config

import "time"

type Config struct {
	Threads     int
	Connections int
	Duration    time.Duration
	Method string
	Headers []string
	Body []byte
	Timeout time.Duration
	JSONOutput bool
	OutFile string
	URL string
}
