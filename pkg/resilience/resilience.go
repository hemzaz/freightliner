// Package resilience provides battle-tested reliability patterns for distributed systems
package resilience

import (
	"context"
	"fmt"

	"freightliner/pkg/helper/log"
)

// Manager coordinates all resilience patterns
type Manager struct {
	circuitBreakers *CircuitBreakerManager
	bulkheads       *BulkheadManager
	rateLimiters    *RateLimiterManager
	retryManager    *RetryManager
	degradation     *DegradationManager
	healthChecker   *HealthChecker
	logger          log.Logger
}

// NewManager creates a new resilience manager with all patterns
func NewManager(logger log.Logger) *Manager {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	return &Manager{
		circuitBreakers: NewCircuitBreakerManager(logger),
		bulkheads:       NewBulkheadManager(logger),
		rateLimiters:    NewRateLimiterManager(logger),
		retryManager:    NewRetryManager(logger),
		degradation:     NewDegradationManager(logger),
		healthChecker:   NewHealthChecker(logger),
		logger:          logger,
	}
}

// ExecuteWithResilience executes a function with all resilience patterns
func (m *Manager) ExecuteWithResilience(ctx context.Context, name string, fn func() error) error {
	// 1. Check rate limit
	rateLimiter := m.rateLimiters.GetOrCreate(name, DefaultRateLimiterSettings())
	if !rateLimiter.Allow() {
		return fmt.Errorf("rate limit exceeded for '%s'", name)
	}

	// 2. Check circuit breaker
	circuitBreaker := m.circuitBreakers.GetOrCreate(name, DefaultCircuitBreakerSettings(name))

	// 3. Execute with bulkhead isolation and retry
	return circuitBreaker.Execute(func() error {
		bulkhead := m.bulkheads.GetOrCreate(name, DefaultBulkheadSettings())

		return bulkhead.Execute(ctx, func() error {
			// Execute with retry policy
			policy := m.retryManager.GetPolicy(name)
			return policy.RetryWithLogger(ctx, fn, m.logger)
		})
	})
}

// CircuitBreakers returns the circuit breaker manager
func (m *Manager) CircuitBreakers() *CircuitBreakerManager {
	return m.circuitBreakers
}

// Bulkheads returns the bulkhead manager
func (m *Manager) Bulkheads() *BulkheadManager {
	return m.bulkheads
}

// RateLimiters returns the rate limiter manager
func (m *Manager) RateLimiters() *RateLimiterManager {
	return m.rateLimiters
}

// Retry returns the retry manager
func (m *Manager) Retry() *RetryManager {
	return m.retryManager
}

// Degradation returns the degradation manager
func (m *Manager) Degradation() *DegradationManager {
	return m.degradation
}

// Health returns the health checker
func (m *Manager) Health() *HealthChecker {
	return m.healthChecker
}

// Start starts all resilience components
func (m *Manager) Start() {
	m.healthChecker.Start()
	m.logger.Info("Resilience manager started")
}

// Stop stops all resilience components
func (m *Manager) Stop() {
	m.healthChecker.Stop()
	m.logger.Info("Resilience manager stopped")
}

// GetSystemHealth returns overall system health
func (m *Manager) GetSystemHealth() SystemHealth {
	return SystemHealth{
		Status:          m.healthChecker.GetStatus(),
		HealthChecks:    m.healthChecker.GetAllResults(),
		CircuitBreakers: m.circuitBreakers.GetAllStates(),
		Bulkheads:       m.bulkheads.GetAllStats(),
		RateLimiters:    m.rateLimiters.GetAllStats(),
	}
}

// SystemHealth represents the overall health of the system
type SystemHealth struct {
	Status          HealthStatus
	HealthChecks    []HealthCheckResult
	CircuitBreakers map[string]State
	Bulkheads       []BulkheadStats
	RateLimiters    []RateLimiterStats
}

// IsHealthy returns true if the system is healthy
func (s SystemHealth) IsHealthy() bool {
	return s.Status == HealthStatusHealthy
}

// GetUnhealthyComponents returns names of unhealthy components
func (s SystemHealth) GetUnhealthyComponents() []string {
	unhealthy := make([]string, 0)

	// Check health checks
	for _, check := range s.HealthChecks {
		if check.Status == HealthStatusUnhealthy {
			unhealthy = append(unhealthy, fmt.Sprintf("health:%s", check.Name))
		}
	}

	// Check circuit breakers
	for name, state := range s.CircuitBreakers {
		if state == StateOpen {
			unhealthy = append(unhealthy, fmt.Sprintf("circuit:%s", name))
		}
	}

	// Check bulkheads
	for _, bulkhead := range s.Bulkheads {
		utilizationPct := float64(bulkhead.ActiveCount) / float64(bulkhead.MaxConcurrent) * 100
		if utilizationPct > 90 {
			unhealthy = append(unhealthy, fmt.Sprintf("bulkhead:%s", bulkhead.Name))
		}
	}

	// Check rate limiters
	for _, limiter := range s.RateLimiters {
		if limiter.TotalRequests > 0 {
			deniedPct := float64(limiter.DeniedRequests) / float64(limiter.TotalRequests) * 100
			if deniedPct > 10 {
				unhealthy = append(unhealthy, fmt.Sprintf("ratelimit:%s", limiter.Name))
			}
		}
	}

	return unhealthy
}
