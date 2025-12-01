package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestLoggingMiddleware tests the logging middleware
func TestLoggingMiddleware(t *testing.T) {
	server := createTestServer(t)

	handler := server.loggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())
}

// TestMetricsMiddleware tests the metrics middleware
func TestMetricsMiddleware(t *testing.T) {
	server := createTestServer(t)

	handler := server.metricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Greater(t, server.metricsRegistry.GetTotalRequests(), uint64(0))
}

// TestRecoveryMiddleware tests panic recovery
func TestRecoveryMiddleware(t *testing.T) {
	server := createTestServer(t)

	handler := server.recoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Greater(t, server.metricsRegistry.GetPanicCount(), uint64(0))
}

// TestCORSMiddleware tests CORS headers
func TestCORSMiddleware(t *testing.T) {
	server := createTestServer(t)
	server.cfg.Server.EnableCORS = true
	server.cfg.Server.AllowedOrigins = []string{"https://example.com"}

	handler := server.corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name         string
		origin       string
		method       string
		expectOrigin string
	}{
		{
			name:         "allowed origin",
			origin:       "https://example.com",
			method:       "GET",
			expectOrigin: "https://example.com",
		},
		{
			name:         "preflight request",
			origin:       "https://example.com",
			method:       "OPTIONS",
			expectOrigin: "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			req.Header.Set("Origin", tt.origin)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if tt.method == "OPTIONS" {
				assert.Equal(t, http.StatusOK, w.Code)
			}
			assert.Equal(t, tt.expectOrigin, w.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}

// TestAuthMiddleware tests authentication middleware
func TestAuthMiddleware(t *testing.T) {
	server := createTestServer(t)
	server.cfg.Server.APIKeyAuth = true
	server.cfg.Server.APIKey = "secret-key"

	handler := server.authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name           string
		path           string
		apiKey         string
		expectedStatus int
	}{
		{
			name:           "valid api key",
			path:           "/api/v1/test",
			apiKey:         "secret-key",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid api key",
			path:           "/api/v1/test",
			apiKey:         "wrong-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "health check bypass",
			path:           "/health",
			apiKey:         "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "metrics bypass",
			path:           "/metrics",
			apiKey:         "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "bearer token",
			path:           "/api/v1/test",
			apiKey:         "Bearer secret-key",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.apiKey != "" {
				if tt.apiKey == "Bearer secret-key" {
					req.Header.Set("Authorization", tt.apiKey)
				} else {
					req.Header.Set("X-API-Key", tt.apiKey)
				}
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestRateLimitMiddleware tests rate limiting
func TestRateLimitMiddleware(t *testing.T) {
	server := createTestServer(t)

	handler := server.rateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make many requests to trigger rate limit
	for i := 0; i < 105; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if i < 100 {
			assert.Equal(t, http.StatusOK, w.Code)
		} else {
			// After 100 requests, should be rate limited
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
		}
	}
}

// TestRateLimiter tests the rate limiter implementation
func TestRateLimiter(t *testing.T) {
	limiter := newRateLimiter(10, time.Minute)

	// Allow first 10 requests
	for i := 0; i < 10; i++ {
		assert.True(t, limiter.allow("test-client"))
	}

	// 11th request should be denied
	assert.False(t, limiter.allow("test-client"))

	// Different client should still be allowed
	assert.True(t, limiter.allow("another-client"))

	// Cleanup should work
	limiter.cleanup()
}

// TestResponseWriter tests the response writer wrapper
func TestResponseWriter(t *testing.T) {
	w := httptest.NewRecorder()
	wrapped := &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	wrapped.WriteHeader(http.StatusCreated)
	assert.Equal(t, http.StatusCreated, wrapped.statusCode)
	assert.Equal(t, http.StatusCreated, w.Code)
}

// TestGetRealIP tests real IP extraction
func TestGetRealIP(t *testing.T) {
	server := createTestServer(t)

	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name: "X-Forwarded-For",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 10.0.0.1",
			},
			remoteAddr: "172.16.0.1:1234",
			expected:   "192.168.1.1",
		},
		{
			name: "X-Real-IP",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.2",
			},
			remoteAddr: "172.16.0.1:1234",
			expected:   "192.168.1.2",
		},
		{
			name: "CF-Connecting-IP",
			headers: map[string]string{
				"CF-Connecting-IP": "192.168.1.3",
			},
			remoteAddr: "172.16.0.1:1234",
			expected:   "192.168.1.3",
		},
		{
			name:       "RemoteAddr fallback",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.4:5678",
			expected:   "192.168.1.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			ip := server.getRealIP(req)
			assert.Equal(t, tt.expected, ip)
		})
	}
}

// TestIsOriginAllowed tests origin validation for CORS
func TestIsOriginAllowed(t *testing.T) {
	server := createTestServer(t)

	tests := []struct {
		name           string
		allowedOrigins []string
		testOrigin     string
		expected       bool
	}{
		{
			name:           "exact match",
			allowedOrigins: []string{"https://example.com"},
			testOrigin:     "https://example.com",
			expected:       true,
		},
		{
			name:           "wildcard all",
			allowedOrigins: []string{"*"},
			testOrigin:     "https://any-origin.com",
			expected:       true,
		},
		{
			name:           "wildcard subdomain",
			allowedOrigins: []string{"*.example.com"},
			testOrigin:     "https://sub.example.com",
			expected:       true,
		},
		{
			name:           "wildcard base domain",
			allowedOrigins: []string{"*.example.com"},
			testOrigin:     "https://example.com",
			expected:       false, // *.example.com doesn't match the base domain with protocol
		},
		{
			name:           "no match",
			allowedOrigins: []string{"https://example.com"},
			testOrigin:     "https://evil.com",
			expected:       false,
		},
		{
			name:           "empty allowed origins",
			allowedOrigins: []string{},
			testOrigin:     "https://any.com",
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server.cfg.Server.AllowedOrigins = tt.allowedOrigins
			result := server.isOriginAllowed(tt.testOrigin)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetRoutePattern tests route pattern extraction
func TestGetRoutePattern(t *testing.T) {
	server := createTestServer(t)

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	pattern := server.getRoutePattern(req)

	// Should normalize the path
	assert.Contains(t, pattern, "/api/v1")
}
