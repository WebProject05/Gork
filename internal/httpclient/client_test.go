package httpclient

import (
	"net/http"
	"testing"
	"time"

	"gork/internal/config"
)

func TestNewClient(t *testing.T) {
	cfg := &config.Config{
		Connections:       50,
		Timeout:           3 * time.Second,
		Insecure:          true,
		DisableKeepAlives: true,
	}

	client := NewClient(cfg)
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	if client.Timeout != 3*time.Second {
		t.Errorf("expected timeout 3s, got %v", client.Timeout)
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}

	if transport.MaxIdleConns != 100 {
		t.Errorf("expected MaxIdleConns 100, got %d", transport.MaxIdleConns)
	}
	if transport.MaxIdleConnsPerHost != 100 {
		t.Errorf("expected MaxIdleConnsPerHost 100, got %d", transport.MaxIdleConnsPerHost)
	}
	if transport.MaxConnsPerHost != 100 {
		t.Errorf("expected MaxConnsPerHost 100, got %d", transport.MaxConnsPerHost)
	}
	if !transport.DisableKeepAlives {
		t.Errorf("expected DisableKeepAlives true")
	}
	if transport.TLSClientConfig == nil || !transport.TLSClientConfig.InsecureSkipVerify {
		t.Errorf("expected TLS InsecureSkipVerify true")
	}
}
