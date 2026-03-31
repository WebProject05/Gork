package httpclient

import (
	"bytes"
	"net/http"
	"strings"

	"gork/internal/config"
)

// Builder helps construct HTTP requests efficiently for workers.
type Builder struct {
	cfg *config.Config
}

func NewBuilder(cfg *config.Config) *Builder {
	return &Builder{cfg: cfg}
}

// Build creates a fresh http.Request. We must do this per request
// because the Body (io.ReadCloser) is consumed.
func (b *Builder) Build() (*http.Request, error) {
	var bodyReader *bytes.Reader
	if b.cfg.Body != nil {
		bodyReader = bytes.NewReader(b.cfg.Body)
	}

	var req *http.Request
	var err error
	if bodyReader != nil {
		req, err = http.NewRequest(b.cfg.Method, b.cfg.URL, bodyReader)
	} else {
		req, err = http.NewRequest(b.cfg.Method, b.cfg.URL, nil)
	}

	if err != nil {
		return nil, err
	}

	for _, h := range b.cfg.Headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			req.Header.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	return req, nil
}
