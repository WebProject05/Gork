package httpclient

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"gork/internal/config"
)

// NewClient constructs an optimized http.Client tailored for high-throughput benchmarking.
func NewClient(cfg *config.Config) *http.Client {
	var transport *http.Transport

	if defaultTrans, ok := http.DefaultTransport.(*http.Transport); ok {
		transport = defaultTrans.Clone()
	} else {
		transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	}

	// Concurrency-tuned pooling
	transport.MaxIdleConns = cfg.Connections * 2
	transport.MaxIdleConnsPerHost = cfg.Connections * 2
	transport.MaxConnsPerHost = cfg.Connections * 2
	transport.DisableKeepAlives = cfg.DisableKeepAlives
	transport.ResponseHeaderTimeout = cfg.Timeout

	if cfg.Insecure {
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.InsecureSkipVerify = true
	}

	return &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}
}
