package server

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Log the request
		duration := time.Since(start)
		s.logger.WithFields(map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"status":     wrapped.statusCode,
			"duration":   duration.String(),
			"remote_ip":  s.getRealIP(r),
			"user_agent": r.UserAgent(),
		}).Info("HTTP request")
	})
}

// metricsMiddleware records HTTP metrics
func (s *Server) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Record metrics
		duration := time.Since(start)

		// Get route pattern for better metrics grouping
		route := s.getRoutePattern(r)

		// Record HTTP request metrics
		s.metricsRegistry.RecordHTTPRequest(
			r.Method,
			route,
			fmt.Sprintf("%d", wrapped.statusCode),
			duration.Seconds(),
		)
	})
}

// recoveryMiddleware recovers from panics
func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				s.logger.WithFields(map[string]interface{}{
					"method":    r.Method,
					"path":      r.URL.Path,
					"remote_ip": s.getRealIP(r),
					"stack":     string(debug.Stack()),
				}).Error("HTTP handler panic", fmt.Errorf("panic: %v", err))

				// Record panic metric
				s.metricsRegistry.RecordPanic("http_handler")

				// Return 500 error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// corsMiddleware handles CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Set CORS headers
		if s.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(s.cfg.Server.AllowedOrigins) == 0 || s.cfg.Server.AllowedOrigins[0] == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// authMiddleware validates API key authentication
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health checks and metrics
		if strings.HasPrefix(r.URL.Path, "/health") || strings.HasPrefix(r.URL.Path, "/metrics") {
			next.ServeHTTP(w, r)
			return
		}

		// Check if API key auth is enabled
		if !s.cfg.Server.APIKeyAuth {
			next.ServeHTTP(w, r)
			return
		}

		// Get API key from header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			apiKey = r.Header.Get("Authorization")
			if strings.HasPrefix(apiKey, "Bearer ") {
				apiKey = strings.TrimPrefix(apiKey, "Bearer ")
			}
		}

		// Validate API key
		if apiKey == "" || apiKey != s.cfg.Server.APIKey {
			s.logger.WithFields(map[string]interface{}{
				"method":    r.Method,
				"path":      r.URL.Path,
				"remote_ip": s.getRealIP(r),
			}).Warn("Unauthorized API request")

			s.metricsRegistry.RecordAuthFailure("api_key")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			if _, err := w.Write([]byte(`{"error":"Unauthorized","message":"Valid API key required"}`)); err != nil {
				s.logger.WithFields(map[string]interface{}{
					"error": err.Error(),
				}).Error("Failed to write unauthorized response")
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

// rateLimitMiddleware implements basic rate limiting
func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple rate limiting based on IP
		clientIP := s.getRealIP(r)

		// TODO: Implement proper rate limiting with Redis or in-memory store
		// For now, just pass through
		_ = clientIP

		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// getRealIP extracts the real client IP from various headers
func (s *Server) getRealIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP from the comma-separated list
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Check CF-Connecting-IP header (Cloudflare)
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		return cfip
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}

	return r.RemoteAddr
}

// getRoutePattern extracts the route pattern for metrics
func (s *Server) getRoutePattern(r *http.Request) string {
	if route := mux.CurrentRoute(r); route != nil {
		if template, err := route.GetPathTemplate(); err == nil {
			return template
		}
	}

	// Fallback to path
	path := r.URL.Path

	// Normalize common patterns for better metrics grouping
	if strings.HasPrefix(path, "/api/v1/") {
		return strings.Replace(path, "/api/v1", "/api/v1", 1)
	}

	return path
}

// isOriginAllowed checks if an origin is allowed by CORS policy
func (s *Server) isOriginAllowed(origin string) bool {
	if len(s.cfg.Server.AllowedOrigins) == 0 {
		return true
	}

	for _, allowed := range s.cfg.Server.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}

		// Support wildcard subdomains (e.g., *.example.com)
		if strings.HasPrefix(allowed, "*.") {
			domain := allowed[2:] // Remove "*."
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true
			}
		}
	}

	return false
}
