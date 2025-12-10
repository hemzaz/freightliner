package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"freightliner/pkg/helper/log"
)

// RateLimiterSettings configures rate limiter behavior
type RateLimiterSettings struct {
	// RequestsPerSecond is the sustained request rate
	RequestsPerSecond float64
	// BurstSize allows temporary bursts above the rate
	BurstSize int
	// WaitTimeout is the maximum time to wait for a token
	WaitTimeout time.Duration
}

// DefaultRateLimiterSettings returns sensible defaults
func DefaultRateLimiterSettings() RateLimiterSettings {
	return RateLimiterSettings{
		RequestsPerSecond: 100,
		BurstSize:         200,
		WaitTimeout:       5 * time.Second,
	}
}

// RateLimiter manages rate limiting for a specific resource
type RateLimiter struct {
	name     string
	settings RateLimiterSettings
	limiter  *rate.Limiter
	logger   log.Logger
	stats    *rateLimiterStats
}

type rateLimiterStats struct {
	totalRequests   int64
	allowedRequests int64
	deniedRequests  int64
	waitedRequests  int64
	mu              sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(name string, settings RateLimiterSettings, logger log.Logger) *RateLimiter {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	limiter := rate.NewLimiter(
		rate.Limit(settings.RequestsPerSecond),
		settings.BurstSize,
	)

	return &RateLimiter{
		name:     name,
		settings: settings,
		limiter:  limiter,
		logger:   logger,
		stats:    &rateLimiterStats{},
	}
}

// Allow checks if a request should be allowed (non-blocking)
func (r *RateLimiter) Allow() bool {
	r.stats.incrementTotal()
	allowed := r.limiter.Allow()

	if allowed {
		r.stats.incrementAllowed()
	} else {
		r.stats.incrementDenied()
		r.logger.WithFields(map[string]interface{}{
			"rateLimiter": r.name,
		}).Warn("Rate limit exceeded, request denied")
	}

	return allowed
}

// Wait waits until a request can be allowed (blocking with timeout)
func (r *RateLimiter) Wait(ctx context.Context) error {
	r.stats.incrementTotal()

	// Create context with timeout
	waitCtx := ctx
	if r.settings.WaitTimeout > 0 {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, r.settings.WaitTimeout)
		defer cancel()
	}

	// Wait for token
	err := r.limiter.Wait(waitCtx)
	if err != nil {
		r.stats.incrementDenied()
		return fmt.Errorf("rate limiter '%s' wait failed: %w", r.name, err)
	}

	r.stats.incrementAllowed()
	r.stats.incrementWaited()
	return nil
}

// Reserve reserves a token for future use
func (r *RateLimiter) Reserve() *rate.Reservation {
	r.stats.incrementTotal()
	reservation := r.limiter.Reserve()

	if reservation.OK() {
		r.stats.incrementAllowed()
	} else {
		r.stats.incrementDenied()
	}

	return reservation
}

// Stats returns current rate limiter statistics
func (r *RateLimiter) Stats() RateLimiterStats {
	return RateLimiterStats{
		Name:              r.name,
		RequestsPerSecond: r.settings.RequestsPerSecond,
		BurstSize:         r.settings.BurstSize,
		TotalRequests:     r.stats.getTotal(),
		AllowedRequests:   r.stats.getAllowed(),
		DeniedRequests:    r.stats.getDenied(),
		WaitedRequests:    r.stats.getWaited(),
	}
}

// RateLimiterStats represents rate limiter statistics
type RateLimiterStats struct {
	Name              string
	RequestsPerSecond float64
	BurstSize         int
	TotalRequests     int64
	AllowedRequests   int64
	DeniedRequests    int64
	WaitedRequests    int64
}

// Helper methods for stats
func (s *rateLimiterStats) incrementTotal() {
	s.mu.Lock()
	s.totalRequests++
	s.mu.Unlock()
}

func (s *rateLimiterStats) incrementAllowed() {
	s.mu.Lock()
	s.allowedRequests++
	s.mu.Unlock()
}

func (s *rateLimiterStats) incrementDenied() {
	s.mu.Lock()
	s.deniedRequests++
	s.mu.Unlock()
}

func (s *rateLimiterStats) incrementWaited() {
	s.mu.Lock()
	s.waitedRequests++
	s.mu.Unlock()
}

func (s *rateLimiterStats) getTotal() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.totalRequests
}

func (s *rateLimiterStats) getAllowed() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.allowedRequests
}

func (s *rateLimiterStats) getDenied() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.deniedRequests
}

func (s *rateLimiterStats) getWaited() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.waitedRequests
}

// RateLimiterManager manages multiple rate limiters
type RateLimiterManager struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
	logger   log.Logger
}

// NewRateLimiterManager creates a new rate limiter manager
func NewRateLimiterManager(logger log.Logger) *RateLimiterManager {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	return &RateLimiterManager{
		limiters: make(map[string]*RateLimiter),
		logger:   logger,
	}
}

// GetOrCreate gets an existing rate limiter or creates a new one
func (m *RateLimiterManager) GetOrCreate(name string, settings RateLimiterSettings) *RateLimiter {
	m.mu.RLock()
	limiter, exists := m.limiters[name]
	m.mu.RUnlock()

	if exists {
		return limiter
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists := m.limiters[name]; exists {
		return limiter
	}

	// Create new rate limiter
	limiter = NewRateLimiter(name, settings, m.logger)
	m.limiters[name] = limiter

	return limiter
}

// Get retrieves a rate limiter by name
func (m *RateLimiterManager) Get(name string) (*RateLimiter, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	limiter, exists := m.limiters[name]
	return limiter, exists
}

// Allow checks if a request should be allowed for a named limiter
func (m *RateLimiterManager) Allow(name string) bool {
	settings := DefaultRateLimiterSettings()
	limiter := m.GetOrCreate(name, settings)
	return limiter.Allow()
}

// Wait waits for a token from a named limiter
func (m *RateLimiterManager) Wait(ctx context.Context, name string) error {
	settings := DefaultRateLimiterSettings()
	limiter := m.GetOrCreate(name, settings)
	return limiter.Wait(ctx)
}

// GetAllStats returns statistics for all rate limiters
func (m *RateLimiterManager) GetAllStats() []RateLimiterStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make([]RateLimiterStats, 0, len(m.limiters))
	for _, limiter := range m.limiters {
		stats = append(stats, limiter.Stats())
	}
	return stats
}

// UpdateLimit updates the rate limit for a named limiter
func (m *RateLimiterManager) UpdateLimit(name string, requestsPerSecond float64, burstSize int) error {
	m.mu.RLock()
	limiter, exists := m.limiters[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("rate limiter '%s' not found", name)
	}

	limiter.limiter.SetLimit(rate.Limit(requestsPerSecond))
	limiter.limiter.SetBurst(burstSize)
	limiter.settings.RequestsPerSecond = requestsPerSecond
	limiter.settings.BurstSize = burstSize

	m.logger.WithFields(map[string]interface{}{
		"rateLimiter":       name,
		"requestsPerSecond": requestsPerSecond,
		"burstSize":         burstSize,
	}).Info("Updated rate limiter settings")

	return nil
}
