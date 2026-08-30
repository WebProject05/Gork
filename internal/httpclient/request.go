package httpclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"gork/internal/config"
)

// Builder pre-compiles request metadata for zero-allocation request construction in hot worker loops.
type Builder struct {
	cfg        *config.Config
	headers    http.Header
	customHost string
}

// NewBuilder pre-parses headers and initializes the request template.
func NewBuilder(cfg *config.Config) *Builder {
	headers := make(http.Header)
	var customHost string

	for _, h := range cfg.Headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if strings.EqualFold(key, "Host") {
				customHost = val
			} else {
				headers.Add(key, val)
			}
		}
	}

	return &Builder{
		cfg:        cfg,
		headers:    headers,
		customHost: customHost,
	}
}

// BuildWithContext creates a fresh http.Request bound to the worker context.
func (b *Builder) BuildWithContext(ctx context.Context) (*http.Request, error) {
	var bodyReader io.Reader
	if len(b.cfg.Body) > 0 {
		bodyReader = bytes.NewReader(b.cfg.Body)
	}

	req, err := http.NewRequestWithContext(ctx, b.cfg.Method, b.cfg.URL, bodyReader)
	if err != nil {
		return nil, err
	}

	if b.customHost != "" {
		req.Host = b.customHost
	}

	// Efficient header cloning
	for k, vv := range b.headers {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}

	if len(b.cfg.Body) > 0 {
		req.ContentLength = int64(len(b.cfg.Body))
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(b.cfg.Body)), nil
		}
	}

	return req, nil
}

// Build creates a fresh http.Request with context.Background().
func (b *Builder) Build() (*http.Request, error) {
	return b.BuildWithContext(context.Background())
}
