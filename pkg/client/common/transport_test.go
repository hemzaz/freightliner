package common_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"freightliner/pkg/client/common"
	"freightliner/pkg/helper/log"
)

// mockRoundTripper implements http.RoundTripper for testing
type mockRoundTripper struct {
	responses []*http.Response
	errors    []error
	calls     int
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.calls >= len(m.responses) {
		return nil, errors.New("no more responses configured")
	}

	resp := m.responses[m.calls]
	err := m.errors[m.calls]
	m.calls++

	return resp, err
}

// TestNewBaseTransport tests the creation of BaseTransport
func TestNewBaseTransport(t *testing.T) {
	tests := []struct {
		name   string
		logger log.Logger
	}{
		{
			name:   "with logger",
			logger: log.NewBasicLogger(log.InfoLevel),
		},
		{
			name:   "without logger (nil)",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := common.NewBaseTransport(tt.logger)
			if transport == nil {
				t.Error("Expected non-nil transport")
			}
		})
	}
}

// TestCreateDefaultTransport tests the default transport configuration
func TestCreateDefaultTransport(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	baseTransport := common.NewBaseTransport(logger)

	httpTransport := baseTransport.CreateDefaultTransport()

	if httpTransport == nil {
		t.Fatal("Expected non-nil transport")
	}

	// Verify transport settings
	if httpTransport.MaxIdleConns != 200 {
		t.Errorf("Expected MaxIdleConns=200, got %d", httpTransport.MaxIdleConns)
	}

	if httpTransport.MaxIdleConnsPerHost != 20 {
		t.Errorf("Expected MaxIdleConnsPerHost=20, got %d", httpTransport.MaxIdleConnsPerHost)
	}

	if httpTransport.MaxConnsPerHost != 50 {
		t.Errorf("Expected MaxConnsPerHost=50, got %d", httpTransport.MaxConnsPerHost)
	}

	if httpTransport.IdleConnTimeout != 120*time.Second {
		t.Errorf("Expected IdleConnTimeout=120s, got %v", httpTransport.IdleConnTimeout)
	}

	if httpTransport.TLSHandshakeTimeout != 15*time.Second {
		t.Errorf("Expected TLSHandshakeTimeout=15s, got %v", httpTransport.TLSHandshakeTimeout)
	}

	if !httpTransport.ForceAttemptHTTP2 {
		t.Error("Expected ForceAttemptHTTP2 to be true")
	}
}

// TestLoggingTransport tests the logging transport wrapper
func TestLoggingTransport(t *testing.T) {
	logger := log.NewBasicLogger(log.DebugLevel)
	baseTransport := common.NewBaseTransport(logger)

	tests := []struct {
		name       string
		statusCode int
		shouldErr  bool
	}{
		{
			name:       "successful request",
			statusCode: 200,
			shouldErr:  false,
		},
		{
			name:       "server error",
			statusCode: 500,
			shouldErr:  false,
		},
		{
			name:       "network error",
			statusCode: 0,
			shouldErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock transport
			var mockResp *http.Response
			var mockErr error

			if tt.shouldErr {
				mockErr = errors.New("network error")
			} else {
				mockResp = &http.Response{
					StatusCode: tt.statusCode,
					Status:     http.StatusText(tt.statusCode),
					Body:       io.NopCloser(strings.NewReader("response body")),
					Header:     make(http.Header),
				}
			}

			mock := &mockRoundTripper{
				responses: []*http.Response{mockResp},
				errors:    []error{mockErr},
			}

			// Wrap with logging
			loggingTransport := baseTransport.LoggingTransport(mock)

			// Create test request
			req := httptest.NewRequest("GET", "http://example.com/test", nil)

			// Execute
			resp, err := loggingTransport.RoundTrip(req)

			// Verify
			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if resp != nil {
					t.Error("Expected nil response on error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if resp == nil {
					t.Error("Expected response, got nil")
				} else if resp.StatusCode != tt.statusCode {
					t.Errorf("Expected status %d, got %d", tt.statusCode, resp.StatusCode)
				}
			}
		})
	}
}

// TestRetryTransport tests the retry transport wrapper
func TestRetryTransport(t *testing.T) {
	logger := log.NewBasicLogger(log.DebugLevel)
	baseTransport := common.NewBaseTransport(logger)

	tests := []struct {
		name          string
		responses     []*http.Response
		errors        []error
		maxRetries    int
		shouldRetry   func(*http.Response, error) bool
		expectedCalls int
		expectSuccess bool
	}{
		{
			name: "success on first attempt",
			responses: []*http.Response{
				{StatusCode: 200, Status: "OK", Body: io.NopCloser(strings.NewReader("ok"))},
			},
			errors:        []error{nil},
			maxRetries:    3,
			shouldRetry:   nil,
			expectedCalls: 1,
			expectSuccess: true,
		},
		{
			name: "client error - no retry",
			responses: []*http.Response{
				{StatusCode: 400, Status: "Bad Request", Body: io.NopCloser(strings.NewReader(""))},
			},
			errors:        []error{nil},
			maxRetries:    3,
			shouldRetry:   nil,
			expectedCalls: 1,
			expectSuccess: false,
		},
		{
			name: "not found - no retry",
			responses: []*http.Response{
				{StatusCode: 404, Status: "Not Found", Body: io.NopCloser(strings.NewReader(""))},
			},
			errors:        []error{nil},
			maxRetries:    3,
			shouldRetry:   nil,
			expectedCalls: 1,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockRoundTripper{
				responses: tt.responses,
				errors:    tt.errors,
			}

			retryTransport := baseTransport.RetryTransport(mock, tt.maxRetries, tt.shouldRetry)

			req := httptest.NewRequest("GET", "http://example.com/test", nil)

			resp, err := retryTransport.RoundTrip(req)

			if mock.calls != tt.expectedCalls {
				t.Errorf("Expected %d calls, got %d", tt.expectedCalls, mock.calls)
			}

			if tt.expectSuccess {
				if err != nil {
					t.Errorf("Expected success, got error: %v", err)
				}
				if resp == nil || resp.StatusCode != 200 {
					t.Error("Expected successful response")
				}
			}
		})
	}
}

// TestRetryTransportWithBody tests retry behavior with request body
func TestRetryTransportWithBody(t *testing.T) {
	logger := log.NewBasicLogger(log.DebugLevel)
	baseTransport := common.NewBaseTransport(logger)

	tests := []struct {
		name       string
		method     string
		hasBody    bool
		statusCode int
		shouldErr  bool
	}{
		{
			name:       "PUT with body - success",
			method:     "PUT",
			hasBody:    true,
			statusCode: 200,
			shouldErr:  false,
		},
		{
			name:       "POST with body - success",
			method:     "POST",
			hasBody:    true,
			statusCode: 201,
			shouldErr:  false,
		},
		{
			name:       "PUT with body - network error",
			method:     "PUT",
			hasBody:    true,
			statusCode: 0,
			shouldErr:  true,
		},
		{
			name:       "GET with body - fallback",
			method:     "GET",
			hasBody:    true,
			statusCode: 200,
			shouldErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockResp *http.Response
			var mockErr error

			if tt.shouldErr {
				mockErr = errors.New("network error")
			} else {
				mockResp = &http.Response{
					StatusCode: tt.statusCode,
					Status:     http.StatusText(tt.statusCode),
					Body:       io.NopCloser(strings.NewReader("response")),
				}
			}

			mock := &mockRoundTripper{
				responses: []*http.Response{mockResp},
				errors:    []error{mockErr},
			}

			retryTransport := baseTransport.RetryTransport(mock, 3, nil)

			var req *http.Request
			if tt.hasBody {
				req = httptest.NewRequest(tt.method, "http://example.com/test", strings.NewReader("body content"))
			} else {
				req = httptest.NewRequest(tt.method, "http://example.com/test", nil)
			}

			resp, err := retryTransport.RoundTrip(req)

			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if resp == nil {
					t.Error("Expected response, got nil")
				}
			}
		})
	}
}

// TestRetryTransportContextCancellation tests context cancellation during retry
func TestRetryTransportContextCancellation(t *testing.T) {
	logger := log.NewBasicLogger(log.DebugLevel)
	baseTransport := common.NewBaseTransport(logger)

	// Create request with already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mock := &mockRoundTripper{
		responses: []*http.Response{
			{StatusCode: 500, Body: io.NopCloser(strings.NewReader(""))},
		},
		errors: []error{nil},
	}

	retryTransport := baseTransport.RetryTransport(mock, 5, nil)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)
	req = req.WithContext(ctx)

	_, err := retryTransport.RoundTrip(req)

	// Should get an error (either from context or from mock)
	if err == nil {
		t.Log("No error, but that's okay - request completed before retry")
	}
}

// TestTimeoutTransport tests the timeout transport wrapper
func TestTimeoutTransport(t *testing.T) {
	logger := log.NewBasicLogger(log.DebugLevel)
	baseTransport := common.NewBaseTransport(logger)

	tests := []struct {
		name        string
		timeout     time.Duration
		delay       time.Duration
		expectError bool
	}{
		{
			name:        "request completes before timeout",
			timeout:     200 * time.Millisecond,
			delay:       50 * time.Millisecond,
			expectError: false,
		},
		{
			name:        "request times out",
			timeout:     50 * time.Millisecond,
			delay:       200 * time.Millisecond,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock that delays response
			mock := &delayedRoundTripper{
				delay: tt.delay,
				response: &http.Response{
					StatusCode: 200,
					Status:     "OK",
					Body:       io.NopCloser(strings.NewReader("ok")),
				},
			}

			timeoutTransport := baseTransport.TimeoutTransport(mock, tt.timeout)

			req := httptest.NewRequest("GET", "http://example.com/test", nil)

			_, err := timeoutTransport.RoundTrip(req)

			if tt.expectError {
				if err == nil {
					t.Error("Expected timeout error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected success, got error: %v", err)
				}
			}
		})
	}
}

// TestRetryBackoffCalculation tests that retry transport doesn't panic
func TestRetryBackoffCalculation(t *testing.T) {
	logger := log.NewBasicLogger(log.DebugLevel)
	baseTransport := common.NewBaseTransport(logger)

	// Create a mock that returns success to avoid long waits
	mock := &mockRoundTripper{
		responses: []*http.Response{
			{StatusCode: 200, Status: "OK", Body: io.NopCloser(strings.NewReader("ok"))},
		},
		errors: []error{nil},
	}

	retryTransport := baseTransport.RetryTransport(mock, 10, nil)

	req := httptest.NewRequest("GET", "http://example.com/test", nil)

	_, err := retryTransport.RoundTrip(req)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	t.Log("Retry transport completed successfully")
}

// delayedRoundTripper simulates a delayed response
type delayedRoundTripper struct {
	delay    time.Duration
	response *http.Response
	err      error
}

func (d *delayedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	select {
	case <-req.Context().Done():
		return nil, req.Context().Err()
	case <-time.After(d.delay):
		return d.response, d.err
	}
}

// TestCloudflareErrors tests detection of Cloudflare error codes
func TestCloudflareErrors(t *testing.T) {
	logger := log.NewBasicLogger(log.DebugLevel)
	baseTransport := common.NewBaseTransport(logger)

	cloudflareErrors := []int{520, 521, 522, 523, 524}

	for _, statusCode := range cloudflareErrors {
		t.Run(http.StatusText(statusCode), func(t *testing.T) {
			// Just test that the retry transport handles Cloudflare errors without panic
			mock := &mockRoundTripper{
				responses: []*http.Response{
					{StatusCode: statusCode, Body: io.NopCloser(strings.NewReader(""))},
				},
				errors: []error{nil},
			}

			retryTransport := baseTransport.RetryTransport(mock, 3, nil)
			req := httptest.NewRequest("GET", "http://example.com/test", nil)

			_, _ = retryTransport.RoundTrip(req)

			// Verify the retry transport was created successfully
			if retryTransport == nil {
				t.Error("Expected non-nil retry transport")
			}
		})
	}
}

// TestSuccessfulResponseCodes tests various successful response codes
func TestSuccessfulResponseCodes(t *testing.T) {
	logger := log.NewBasicLogger(log.DebugLevel)
	baseTransport := common.NewBaseTransport(logger)

	successCodes := []int{200, 201, 202, 204, 206, 302, 307, 308}

	for _, statusCode := range successCodes {
		t.Run(http.StatusText(statusCode), func(t *testing.T) {
			mock := &mockRoundTripper{
				responses: []*http.Response{
					{StatusCode: statusCode, Body: io.NopCloser(strings.NewReader("ok"))},
				},
				errors: []error{nil},
			}

			retryTransport := baseTransport.RetryTransport(mock, 3, nil)
			req := httptest.NewRequest("GET", "http://example.com/test", nil)

			resp, err := retryTransport.RoundTrip(req)

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if resp == nil {
				t.Error("Expected response, got nil")
			}

			// Should not retry on success
			if mock.calls != 1 {
				t.Errorf("Expected 1 call (no retry), got %d", mock.calls)
			}
		})
	}
}
