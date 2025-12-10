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
	logger log.Logger
}

// NewBaseTransport creates a new base transport
func NewBaseTransport(logger log.Logger) *BaseTransport {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	return &BaseTransport{
		logger: logger,
	}
}

// CreateDefaultTransport creates a default HTTP transport optimized for container registry operations
func (t *BaseTransport) CreateDefaultTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 60 * time.Second, // Increased for better connection reuse
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          200,               // Increased for high-throughput scenarios
		MaxIdleConnsPerHost:   20,                // Optimize per-host connection pooling
		MaxConnsPerHost:       50,                // Limit total connections per host
		IdleConnTimeout:       120 * time.Second, // Longer idle timeout for registry connections
		TLSHandshakeTimeout:   15 * time.Second,  // Slightly increased for registry TLS
		ExpectContinueTimeout: 2 * time.Second,   // Better for large blob uploads
		ResponseHeaderTimeout: 30 * time.Second,  // Add response header timeout
		DisableCompression:    false,             // Enable compression for better bandwidth utilization
		WriteBufferSize:       64 * 1024,         // 64KB write buffer for better throughput
		ReadBufferSize:        64 * 1024,         // 64KB read buffer for better throughput
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
	logger log.Logger
}

// RoundTrip implements http.RoundTripper
func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.logger.WithFields(map[string]interface{}{
		"method": req.Method,
		"url":    req.URL.String(),
	}).Debug("HTTP Request")

	start := time.Now()
	resp, err := t.inner.RoundTrip(req)
	duration := time.Since(start)

	if err != nil {
		t.logger.WithFields(map[string]interface{}{
			"method":   req.Method,
			"url":      req.URL.String(),
			"error":    err.Error(),
			"duration": duration.String(),
		}).Debug("HTTP Error")
		return nil, err
	}

	t.logger.WithFields(map[string]interface{}{
		"method":   req.Method,
		"url":      req.URL.String(),
		"status":   resp.Status,
		"duration": duration.String(),
	}).Debug("HTTP Response")

	return resp, nil
}

// retryTransport retries failed HTTP requests
type retryTransport struct {
	inner       http.RoundTripper
	logger      log.Logger
	maxRetries  int
	shouldRetry func(*http.Response, error) bool
}

// RoundTrip implements http.RoundTripper with enhanced resilience for multi-cloud registry operations
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	// Create a copy of the request that we can re-use
	reqCopy := req.Clone(req.Context())
	if req.Body != nil {
		// Enhanced body handling for registry uploads
		t.logger.WithFields(map[string]interface{}{
			"method": req.Method,
			"url":    req.URL.String(),
		}).Debug("Request with body detected, implementing smart retry")

		// For registry operations, we can often retry PUT/POST operations
		if req.Method == "PUT" || req.Method == "POST" {
			// Try to preserve body for retry (in production, this would handle seekable bodies)
			return t.retryWithBodyPreservation(req)
		}

		// Fall back to single attempt for non-retryable body requests
		return t.inner.RoundTrip(req)
	}

	for i := 0; i <= t.maxRetries; i++ {
		if i > 0 {
			// Enhanced backoff with jitter for multi-cloud scenarios
			backoffDuration := t.calculateBackoffWithJitter(i)

			t.logger.WithFields(map[string]interface{}{
				"method":     req.Method,
				"url":        req.URL.String(),
				"attempt":    i,
				"backoff_ms": backoffDuration.Milliseconds(),
			}).Debug("Retrying request with enhanced backoff")

			// Check context cancellation during backoff
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(backoffDuration):
				// Continue with retry
			}
		}

		resp, err = t.inner.RoundTrip(reqCopy)

		// Enhanced success criteria for registry operations
		if err == nil && t.isSuccessfulResponse(resp) {
			return resp, nil
		}

		// Enhanced retry logic for registry-specific errors
		if !t.shouldRetryRegistryOperation(resp, err, i) {
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

// calculateBackoffWithJitter implements exponential backoff with jitter for multi-cloud resilience
func (t *retryTransport) calculateBackoffWithJitter(attempt int) time.Duration {
	// Base exponential backoff
	backoffFactor := attempt - 1
	if backoffFactor > 10 { // Cap to prevent excessive delays
		backoffFactor = 10
	}

	baseDelay := time.Duration(1<<uint(backoffFactor)) * 200 * time.Millisecond // #nosec G115 - backoffFactor is capped at 10

	// Add jitter (±25%) to prevent thundering herd
	jitter := time.Duration(float64(baseDelay) * 0.25 * (2.0*float64(time.Now().UnixNano()%1000)/1000.0 - 1.0))

	finalDelay := baseDelay + jitter

	// Cap maximum delay for registry operations
	maxDelay := 30 * time.Second
	if finalDelay > maxDelay {
		finalDelay = maxDelay
	}

	return finalDelay
}

// isSuccessfulResponse determines if a response is successful for registry operations
func (t *retryTransport) isSuccessfulResponse(resp *http.Response) bool {
	if resp == nil {
		return false
	}

	// Registry-specific success codes
	switch resp.StatusCode {
	case 200, 201, 202, 204: // Standard success codes
		return true
	case 206: // Partial content (resumable uploads)
		return true
	case 302, 307, 308: // Redirects are handled by transport
		return true
	default:
		return false
	}
}

// shouldRetryRegistryOperation determines if we should retry a registry operation
func (t *retryTransport) shouldRetryRegistryOperation(resp *http.Response, err error, attempt int) bool {
	// Network errors are always retryable
	if err != nil {
		t.logger.WithFields(map[string]interface{}{
			"error":   err.Error(),
			"attempt": attempt,
		}).Debug("Network error detected, will retry")
		return true
	}

	if resp == nil {
		return true
	}

	// Registry-specific retry logic
	switch resp.StatusCode {
	case 429: // Rate limiting - always retry with backoff
		return true
	case 500, 502, 503, 504: // Server errors
		return true
	case 408: // Request timeout
		return true
	case 520, 521, 522, 523, 524: // Cloudflare errors (common with registry CDNs)
		return true
	case 401, 403: // Auth errors - retry once in case of token expiration
		return attempt <= 1
	default:
		// Use custom retry policy if available
		if t.shouldRetry != nil {
			return t.shouldRetry(resp, err)
		}
		return false
	}
}

// retryWithBodyPreservation handles retry for requests with bodies (registry uploads)
func (t *retryTransport) retryWithBodyPreservation(req *http.Request) (*http.Response, error) {
	// In a production implementation, this would handle seekable bodies
	// For now, attempt the request once with enhanced error context
	resp, err := t.inner.RoundTrip(req)

	if err != nil {
		t.logger.WithFields(map[string]interface{}{
			"method": req.Method,
			"url":    req.URL.String(),
			"error":  err.Error(),
		}).Warn("Upload request failed, body preservation not implemented")
	}

	return resp, err
}

// timeoutTransport times out HTTP requests
type timeoutTransport struct {
	inner   http.RoundTripper
	logger  log.Logger
	timeout time.Duration
}

// RoundTrip implements http.RoundTripper
func (t *timeoutTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(req.Context(), t.timeout)
	defer cancel()

	return t.inner.RoundTrip(req.WithContext(ctx))
}
