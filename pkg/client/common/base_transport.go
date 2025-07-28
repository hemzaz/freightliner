package common

import (
	"context"
	"net"
	"net/http"
	"time"

	"freightliner/pkg/helper/log"
)

// BaseTransport provides common HTTP transport functionality
type BaseTransport struct {
	logger *log.Logger
}

// NewBaseTransport creates a new base transport
func NewBaseTransport(logger *log.Logger) *BaseTransport {
	if logger == nil {
		logger = log.NewLogger(log.InfoLevel)
	}

	return &BaseTransport{
		logger: logger,
	}
}

// CreateDefaultTransport creates a default HTTP transport with reasonable timeouts
func (t *BaseTransport) CreateDefaultTransport() *http.Transport {
	return &http.Transport{
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

// LoggingTransport wraps a transport with request/response logging
func (t *BaseTransport) LoggingTransport(inner http.RoundTripper) http.RoundTripper {
	return &loggingTransport{
		inner:  inner,
		logger: t.logger,
	}
}

// RetryTransport creates a transport that retries failed requests
func (t *BaseTransport) RetryTransport(inner http.RoundTripper, maxRetries int, shouldRetry func(*http.Response, error) bool) http.RoundTripper {
	return &retryTransport{
		inner:       inner,
		logger:      t.logger,
		maxRetries:  maxRetries,
		shouldRetry: shouldRetry,
	}
}

// TimeoutTransport creates a transport that times out requests
func (t *BaseTransport) TimeoutTransport(inner http.RoundTripper, timeout time.Duration) http.RoundTripper {
	return &timeoutTransport{
		inner:   inner,
		logger:  t.logger,
		timeout: timeout,
	}
}

// loggingTransport logs HTTP requests and responses
type loggingTransport struct {
	inner  http.RoundTripper
	logger *log.Logger
}

// RoundTrip implements http.RoundTripper
func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.logger.Debug("HTTP Request", map[string]interface{}{
		"method": req.Method,
		"url":    req.URL.String(),
	})

	start := time.Now()
	resp, err := t.inner.RoundTrip(req)
	duration := time.Since(start)

	if err != nil {
		t.logger.Debug("HTTP Error", map[string]interface{}{
			"method":   req.Method,
			"url":      req.URL.String(),
			"error":    err.Error(),
			"duration": duration.String(),
		})
		return nil, err
	}

	t.logger.Debug("HTTP Response", map[string]interface{}{
		"method":   req.Method,
		"url":      req.URL.String(),
		"status":   resp.Status,
		"duration": duration.String(),
	})

	return resp, nil
}

// retryTransport retries failed HTTP requests
type retryTransport struct {
	inner       http.RoundTripper
	logger      *log.Logger
	maxRetries  int
	shouldRetry func(*http.Response, error) bool
}

// RoundTrip implements http.RoundTripper
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	// Create a copy of the request that we can re-use
	reqCopy := req.Clone(req.Context())
	if req.Body != nil {
		// This is a limitation - we can't retry requests with a body
		t.logger.Warn("Cannot retry request with body", map[string]interface{}{
			"method": req.Method,
			"url":    req.URL.String(),
		})
		return t.inner.RoundTrip(req)
	}

	for i := 0; i <= t.maxRetries; i++ {
		if i > 0 {
			t.logger.Debug("Retrying request", map[string]interface{}{
				"method":  req.Method,
				"url":     req.URL.String(),
				"attempt": i,
			})
			// Use exponential backoff with overflow protection
			backoffFactor := i - 1
			if backoffFactor > 20 { // Cap to prevent overflow
				backoffFactor = 20
			}
			time.Sleep(time.Duration(1<<uint(backoffFactor)) * 100 * time.Millisecond) // #nosec G115 - backoffFactor is capped at 20
		}

		resp, err = t.inner.RoundTrip(reqCopy)

		if err == nil && resp.StatusCode < 500 {
			// Success or non-server error
			return resp, nil
		}

		if t.shouldRetry != nil && !t.shouldRetry(resp, err) {
			// Don't retry based on custom logic
			return resp, err
		}

		if i < t.maxRetries {
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}
		}
	}

	return resp, err
}

// timeoutTransport times out HTTP requests
type timeoutTransport struct {
	inner   http.RoundTripper
	logger  *log.Logger
	timeout time.Duration
}

// RoundTrip implements http.RoundTripper
func (t *timeoutTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(req.Context(), t.timeout)
	defer cancel()

	return t.inner.RoundTrip(req.WithContext(ctx))
}
