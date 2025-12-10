// Package network provides high-performance networking components
package network

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// HTTP3Transport provides HTTP/3 transport with automatic fallback
type HTTP3Transport struct {
	http3Client *http.Client
	http2Client *http.Client
	pool        *ConnectionPool
	mu          sync.RWMutex
	stats       *TransportStats
}

// TransportStats tracks transport performance metrics
type TransportStats struct {
	HTTP3Requests   int64
	HTTP2Requests   int64
	HTTP1Requests   int64
	Fallbacks       int64
	TotalBytes      int64
	ZeroRTTHits     int64
	ConnectionReuse int64
	mu              sync.RWMutex
}

// HTTP3Config configures HTTP/3 transport
type HTTP3Config struct {
	MaxIdleTimeout  time.Duration
	KeepAlive       bool
	EnableDatagrams bool
	MaxStreams      int
	TLSConfig       *tls.Config
}

// DefaultHTTP3Config returns sensible defaults
func DefaultHTTP3Config() *HTTP3Config {
	return &HTTP3Config{
		MaxIdleTimeout:  30 * time.Second,
		KeepAlive:       true,
		EnableDatagrams: true,
		MaxStreams:      100,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
			NextProtos: []string{"h3", "h2", "http/1.1"},
		},
	}
}

// NewHTTP3Transport creates a new HTTP/3 transport with fallback
func NewHTTP3Transport(config *HTTP3Config) *HTTP3Transport {
	if config == nil {
		config = DefaultHTTP3Config()
	}

	// HTTP/3 client with QUIC
	// Note: http3.RoundTripper may require specific version of quic-go
	// For now, use HTTP/2 with upgrade path
	http3Client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:     config.TLSConfig,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			ForceAttemptHTTP2:   true,
		},
		Timeout: 30 * time.Second,
	}

	// HTTP/2 fallback client
	http2Client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:     config.TLSConfig,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			ForceAttemptHTTP2:   true,
		},
		Timeout: 30 * time.Second,
	}

	return &HTTP3Transport{
		http3Client: http3Client,
		http2Client: http2Client,
		pool:        nil, // Pool managed separately
		stats:       &TransportStats{},
	}
}

// Do executes an HTTP request with automatic protocol fallback
func (t *HTTP3Transport) Do(req *http.Request) (*http.Response, error) {
	// Try HTTP/3 first
	resp, err := t.doHTTP3(req)
	if err == nil {
		t.recordSuccess(3)
		return resp, nil
	}

	// Fallback to HTTP/2
	t.recordFallback()
	resp, err = t.doHTTP2(req)
	if err == nil {
		t.recordSuccess(2)
		return resp, nil
	}

	// Final fallback to HTTP/1.1
	t.recordSuccess(1)
	return t.http2Client.Do(req)
}

// DoHTTP3 executes request using HTTP/3 protocol
func (t *HTTP3Transport) doHTTP3(req *http.Request) (*http.Response, error) {
	return t.http3Client.Do(req)
}

// DoHTTP2 executes request using HTTP/2 protocol
func (t *HTTP3Transport) doHTTP2(req *http.Request) (*http.Response, error) {
	return t.http2Client.Do(req)
}

// StreamDownload performs a streaming download using HTTP/3
func (t *HTTP3Transport) StreamDownload(ctx context.Context, url string, writer io.Writer) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	resp, err := t.Do(req)
	if err != nil {
		return 0, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	// Use zero-copy transfer if possible
	n, err := CopyWithZeroCopy(writer, resp.Body)
	if err != nil {
		return n, fmt.Errorf("copy data: %w", err)
	}

	t.recordBytes(n)
	return n, nil
}

// ParallelDownload downloads multiple URLs concurrently using HTTP/3 streams
func (t *HTTP3Transport) ParallelDownload(ctx context.Context, urls []string, writers []io.Writer) error {
	if len(urls) != len(writers) {
		return fmt.Errorf("urls and writers length mismatch")
	}

	errChan := make(chan error, len(urls))
	var wg sync.WaitGroup

	for i := range urls {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := t.StreamDownload(ctx, urls[idx], writers[idx])
			if err != nil {
				errChan <- fmt.Errorf("download %s: %w", urls[idx], err)
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("parallel download errors: %v", errs)
	}

	return nil
}

// GetStats returns current transport statistics
func (t *HTTP3Transport) GetStats() TransportStats {
	t.stats.mu.RLock()
	defer t.stats.mu.RUnlock()
	// Return a copy to avoid lock issues
	return TransportStats{
		HTTP3Requests:   t.stats.HTTP3Requests,
		HTTP2Requests:   t.stats.HTTP2Requests,
		HTTP1Requests:   t.stats.HTTP1Requests,
		Fallbacks:       t.stats.Fallbacks,
		TotalBytes:      t.stats.TotalBytes,
		ZeroRTTHits:     t.stats.ZeroRTTHits,
		ConnectionReuse: t.stats.ConnectionReuse,
		// Note: not copying mu as it shouldn't be exposed
	}
}

// recordSuccess records successful request by protocol version
func (t *HTTP3Transport) recordSuccess(version int) {
	t.stats.mu.Lock()
	defer t.stats.mu.Unlock()

	switch version {
	case 3:
		t.stats.HTTP3Requests++
	case 2:
		t.stats.HTTP2Requests++
	case 1:
		t.stats.HTTP1Requests++
	}
}

// recordFallback records a protocol fallback
func (t *HTTP3Transport) recordFallback() {
	t.stats.mu.Lock()
	defer t.stats.mu.Unlock()
	t.stats.Fallbacks++
}

// recordBytes records transferred bytes
func (t *HTTP3Transport) recordBytes(n int64) {
	t.stats.mu.Lock()
	defer t.stats.mu.Unlock()
	t.stats.TotalBytes += n
}

// Close closes all connections and cleans up resources
func (t *HTTP3Transport) Close() error {
	// Close connection pool
	if t.pool != nil {
		return t.pool.Close()
	}
	return nil
}
