package httpclient

import (
	"context"
	"io"
	"testing"

	"gork/internal/config"
)

func TestBuilderHeadersAndHost(t *testing.T) {
	cfg := &config.Config{
		Method: "POST",
		URL:    "http://example.com/api/test",
		Headers: []string{
			"Content-Type: application/json",
			"Authorization: Bearer mytoken",
			"Host: custom.host.com",
			"X-Custom-Header: value with spaces ",
		},
		Body: []byte(`{"message":"hello"}`),
	}

	builder := NewBuilder(cfg)
	req, err := builder.BuildWithContext(context.Background())
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	if req.Method != "POST" {
		t.Errorf("expected POST, got %s", req.Method)
	}
	if req.URL.String() != "http://example.com/api/test" {
		t.Errorf("expected URL http://example.com/api/test, got %s", req.URL.String())
	}
	if req.Host != "custom.host.com" {
		t.Errorf("expected Host custom.host.com, got %s", req.Host)
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", req.Header.Get("Content-Type"))
	}
	if req.Header.Get("Authorization") != "Bearer mytoken" {
		t.Errorf("expected Authorization Bearer mytoken, got %s", req.Header.Get("Authorization"))
	}
	if req.Header.Get("X-Custom-Header") != "value with spaces" {
		t.Errorf("expected X-Custom-Header 'value with spaces', got %s", req.Header.Get("X-Custom-Header"))
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}
	if string(bodyBytes) != `{"message":"hello"}` {
		t.Errorf("expected body %s, got %s", `{"message":"hello"}`, string(bodyBytes))
	}
}
