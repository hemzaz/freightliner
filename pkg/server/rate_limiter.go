package server

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu            sync.RWMutex
	clients       map[string]*bucket
	rate          int           // requests per window
	window        time.Duration // time window
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

// bucket represents a token bucket for rate limiting
type bucket struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerWindow int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients:     make(map[string]*bucket),
		rate:        requestsPerWindow,
		window:      window,
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup goroutine to remove old entries
	rl.cleanupTicker = time.NewTicker(window)
	go rl.cleanup()

	return rl
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.RLock()
	b, exists := rl.clients[clientID]
	rl.mu.RUnlock()

	if !exists {
		b = &bucket{
			tokens:     rl.rate,
			lastRefill: time.Now(),
		}
		rl.mu.Lock()
		rl.clients[clientID] = b
		rl.mu.Unlock()
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens based on time passed
	now := time.Now()
	elapsed := now.Sub(b.lastRefill)
	if elapsed >= rl.window {
		b.tokens = rl.rate
		b.lastRefill = now
	}

	// Check if we have tokens available
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// GetRemainingTokens returns remaining tokens for a client
func (rl *RateLimiter) GetRemainingTokens(clientID string) int {
	rl.mu.RLock()
	b, exists := rl.clients[clientID]
	rl.mu.RUnlock()

	if !exists {
		return rl.rate
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens based on time passed
	now := time.Now()
	elapsed := now.Sub(b.lastRefill)
	if elapsed >= rl.window {
		return rl.rate
	}

	return b.tokens
}

// GetResetTime returns when the rate limit will reset for a client
func (rl *RateLimiter) GetResetTime(clientID string) time.Time {
	rl.mu.RLock()
	b, exists := rl.clients[clientID]
	rl.mu.RUnlock()

	if !exists {
		return time.Now()
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	return b.lastRefill.Add(rl.window)
}

// cleanup removes old bucket entries
func (rl *RateLimiter) cleanup() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.mu.Lock()
			now := time.Now()
			for clientID, b := range rl.clients {
				b.mu.Lock()
				if now.Sub(b.lastRefill) > rl.window*2 {
					delete(rl.clients, clientID)
				}
				b.mu.Unlock()
			}
			rl.mu.Unlock()
		case <-rl.stopCleanup:
			return
		}
	}
}

// Stop stops the rate limiter cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.stopCleanup)
	rl.cleanupTicker.Stop()
}

// getClientID extracts a client identifier from the request
func (s *Server) getClientID(r *http.Request) string {
	// Prefer API key if present
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return "key:" + apiKey
	}

	// Fall back to IP address
	ip := r.RemoteAddr
	// Strip port if present
	if idx := len(ip) - 1; idx >= 0 {
		for i := idx; i >= 0; i-- {
			if ip[i] == ':' {
				ip = ip[:i]
				break
			}
		}
	}
	return "ip:" + ip
}

// PerRegistryRateLimiter implements rate limiting per registry
type PerRegistryRateLimiter struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
}

// NewPerRegistryRateLimiter creates a new per-registry rate limiter
func NewPerRegistryRateLimiter() *PerRegistryRateLimiter {
	return &PerRegistryRateLimiter{
		limiters: make(map[string]*RateLimiter),
	}
}

// SetLimit sets the rate limit for a specific registry
func (prl *PerRegistryRateLimiter) SetLimit(registry string, requestsPerWindow int, window time.Duration) {
	prl.mu.Lock()
	defer prl.mu.Unlock()

	prl.limiters[registry] = NewRateLimiter(requestsPerWindow, window)
}

// Allow checks if a request to a registry should be allowed
func (prl *PerRegistryRateLimiter) Allow(registry, clientID string) bool {
	prl.mu.RLock()
	limiter, exists := prl.limiters[registry]
	prl.mu.RUnlock()

	if !exists {
		// No rate limit configured for this registry
		return true
	}

	return limiter.Allow(clientID)
}

// Stop stops all rate limiters
func (prl *PerRegistryRateLimiter) Stop() {
	prl.mu.Lock()
	defer prl.mu.Unlock()

	for _, limiter := range prl.limiters {
		limiter.Stop()
	}
}
