package httpclient

import (
	"gork/internal/config"
	"net/http"
)

func NewClient(cfg *config.Config) *http.Client {
	transport := &http.Transport{
		MaxIdleConns: cfg.Connections,
		MaxIdleConnsPerHost: cfg.Connections,
		MaxConnsPerHost: cfg.Connections,
		DisableKeepAlives: false,
	}

	return &http.Client{
		Transport: transport,
		Timeout: cfg.Timeout,
	}
}
