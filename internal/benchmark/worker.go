package benchmark

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptrace"
	"sync/atomic"
	"time"

	"gork/internal/httpclient"
	"gork/internal/metrics"
)

// runWorker executes requests concurrently and records metrics into a local lock-free collector.
func runWorker(ctx context.Context, client *http.Client, builder *httpclient.Builder, rateInterval time.Duration, workerIndex int, activeConns *atomic.Int64) *metrics.Collector {
	collector := metrics.NewCollector()

	var ticker *time.Ticker
	if rateInterval > 0 {
		ticker = time.NewTicker(rateInterval)
		defer ticker.Stop()
	}

	for {
		if ticker != nil {
			select {
			case <-ctx.Done():
				return collector
			case <-ticker.C:
			}
		} else {
			select {
			case <-ctx.Done():
				return collector
			default:
			}
		}

		// Dynamic concurrency gating for ramping stages
		if activeConns != nil {
			if int64(workerIndex) >= activeConns.Load() {
				// Sleep briefly to yield CPU until target concurrency scales up or stage ends
				time.Sleep(10 * time.Millisecond)
				continue
			}
		}

		var dnsStart, dnsEnd time.Time
		var connectStart, connectEnd time.Time
		var tlsStart, tlsEnd time.Time
		var firstByteTime time.Time
		var connReused bool

		trace := &httptrace.ClientTrace{
			DNSStart: func(_ httptrace.DNSStartInfo) { dnsStart = time.Now() },
			DNSDone:  func(_ httptrace.DNSDoneInfo) { dnsEnd = time.Now() },
			ConnectStart: func(_, _ string) { connectStart = time.Now() },
			ConnectDone:  func(_, _ string, _ error) { connectEnd = time.Now() },
			TLSHandshakeStart: func() { tlsStart = time.Now() },
			TLSHandshakeDone:  func(_ tls.ConnectionState, _ error) { tlsEnd = time.Now() },
			GotConn: func(info httptrace.GotConnInfo) {
				connReused = info.Reused
			},
			GotFirstResponseByte: func() {
				firstByteTime = time.Now()
			},
		}

		req, err := builder.BuildWithContext(httptrace.WithClientTrace(ctx, trace))
		if err != nil {
			collector.Record(metrics.Result{Error: err})
			continue
		}

		// Estimate request bytes sent (Method, Path, Headers, Body)
		bytesSent := int64(len(req.Method) + len(req.URL.RequestURI()) + 12)
		for k, vv := range req.Header {
			for _, v := range vv {
				bytesSent += int64(len(k) + len(v) + 4)
			}
		}
		if req.ContentLength > 0 {
			bytesSent += req.ContentLength
		}

		start := time.Now()
		resp, err := client.Do(req)
		latency := time.Since(start)

		if err != nil {
			// If context was canceled or deadline exceeded during request execution,
			// the benchmark timer has expired. Do not record as a server error.
			if ctx.Err() != nil {
				return collector
			}
			collector.Record(metrics.Result{
				Error:      err,
				Latency:    latency,
				BytesSent:  bytesSent,
				ConnReused: connReused,
			})
			continue
		}

		// Drain response body to EOF so the underlying TCP connection can be returned to the pool
		bytesRead, _ := io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		// Calculate phase timings
		var dnsDur, tcpDur, tlsDur, ttfbDur, contentDur time.Duration
		if !dnsEnd.IsZero() && !dnsStart.IsZero() {
			dnsDur = dnsEnd.Sub(dnsStart)
		}
		if !connectEnd.IsZero() && !connectStart.IsZero() {
			tcpDur = connectEnd.Sub(connectStart)
		}
		if !tlsEnd.IsZero() && !tlsStart.IsZero() {
			tlsDur = tlsEnd.Sub(tlsStart)
		}
		if !firstByteTime.IsZero() {
			ttfbDur = firstByteTime.Sub(start)
			if latency > ttfbDur {
				contentDur = latency - ttfbDur
			}
		}

		collector.Record(metrics.Result{
			StatusCode:      resp.StatusCode,
			Latency:         latency,
			BytesRead:       bytesRead,
			BytesSent:       bytesSent,
			ConnReused:      connReused,
			DNSDuration:     dnsDur,
			TCPDuration:     tcpDur,
			TLSDuration:     tlsDur,
			TTFBDuration:    ttfbDur,
			ContentDuration: contentDur,
		})
	}
}

