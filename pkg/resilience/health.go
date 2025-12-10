package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"

	"freightliner/pkg/helper/log"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthCheckFunc is a function that performs a health check
type HealthCheckFunc func(context.Context) error

// HealthCheck represents a health check configuration
type HealthCheck struct {
	// Name of the health check
	Name string
	// Check function to execute
	Check HealthCheckFunc
	// Interval between checks
	Interval time.Duration
	// Timeout for each check
	Timeout time.Duration
	// Critical indicates if failure should mark overall system as unhealthy
	Critical bool
	// OnFailure callback when check fails
	OnFailure func(name string, err error)
	// OnRecovery callback when check recovers
	OnRecovery func(name string)
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Name                string
	Status              HealthStatus
	LastCheck           time.Time
	LastSuccess         time.Time
	LastFailure         time.Time
	ConsecutiveFailures int
	Error               string
	Duration            time.Duration
}

// HealthChecker manages health checks for various components
type HealthChecker struct {
	checks  map[string]*healthCheckState
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
	logger  log.Logger
	started bool
}

type healthCheckState struct {
	config              HealthCheck
	status              HealthStatus
	lastCheck           time.Time
	lastSuccess         time.Time
	lastFailure         time.Time
	consecutiveFailures int
	lastError           error
	lastDuration        time.Duration
	mu                  sync.RWMutex
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(logger log.Logger) *HealthChecker {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &HealthChecker{
		checks: make(map[string]*healthCheckState),
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
	}
}

// RegisterCheck registers a new health check
func (h *HealthChecker) RegisterCheck(check HealthCheck) error {
	if check.Name == "" {
		return fmt.Errorf("health check name cannot be empty")
	}
	if check.Check == nil {
		return fmt.Errorf("health check function cannot be nil")
	}
	if check.Interval <= 0 {
		check.Interval = 30 * time.Second
	}
	if check.Timeout <= 0 {
		check.Timeout = 10 * time.Second
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.checks[check.Name]; exists {
		return fmt.Errorf("health check '%s' already registered", check.Name)
	}

	state := &healthCheckState{
		config: check,
		status: HealthStatusUnknown,
	}

	h.checks[check.Name] = state

	// Start check if health checker is already running
	if h.started {
		go h.runCheck(check.Name, state)
	}

	h.logger.WithFields(map[string]interface{}{
		"checkName": check.Name,
		"interval":  check.Interval.String(),
		"critical":  check.Critical,
	}).Info("Registered health check")

	return nil
}

// Start starts all health checks
func (h *HealthChecker) Start() {
	h.mu.Lock()
	if h.started {
		h.mu.Unlock()
		return
	}
	h.started = true
	checks := make(map[string]*healthCheckState, len(h.checks))
	for name, state := range h.checks {
		checks[name] = state
	}
	h.mu.Unlock()

	h.logger.WithFields(map[string]interface{}{
		"checkCount": len(checks),
	}).Info("Starting health checks")

	// Start all health checks
	for name, state := range checks {
		go h.runCheck(name, state)
	}
}

// Stop stops all health checks
func (h *HealthChecker) Stop() {
	h.mu.Lock()
	h.started = false
	h.mu.Unlock()

	h.cancel()
	h.logger.Info("Stopped health checks")
}

// runCheck runs a single health check in a loop
func (h *HealthChecker) runCheck(name string, state *healthCheckState) {
	ticker := time.NewTicker(state.config.Interval)
	defer ticker.Stop()

	// Run immediately on start
	h.executeCheck(name, state)

	for {
		select {
		case <-ticker.C:
			h.executeCheck(name, state)
		case <-h.ctx.Done():
			return
		}
	}
}

// executeCheck performs a single health check
func (h *HealthChecker) executeCheck(name string, state *healthCheckState) {
	checkCtx, cancel := context.WithTimeout(h.ctx, state.config.Timeout)
	defer cancel()

	start := time.Now()
	err := state.config.Check(checkCtx)
	duration := time.Since(start)

	state.mu.Lock()
	state.lastCheck = time.Now()
	state.lastDuration = duration

	previousStatus := state.status

	if err != nil {
		state.consecutiveFailures++
		state.lastFailure = time.Now()
		state.lastError = err
		state.status = HealthStatusUnhealthy

		h.logger.WithFields(map[string]interface{}{
			"checkName":           name,
			"duration":            duration.String(),
			"consecutiveFailures": state.consecutiveFailures,
		}).Error("Health check failed", err)

		// Call OnFailure callback
		if state.config.OnFailure != nil && previousStatus != HealthStatusUnhealthy {
			go state.config.OnFailure(name, err)
		}
	} else {
		wasUnhealthy := state.consecutiveFailures > 0
		state.consecutiveFailures = 0
		state.lastSuccess = time.Now()
		state.lastError = nil
		state.status = HealthStatusHealthy

		h.logger.WithFields(map[string]interface{}{
			"checkName": name,
			"duration":  duration.String(),
		}).Debug("Health check passed")

		// Call OnRecovery callback
		if state.config.OnRecovery != nil && wasUnhealthy {
			go state.config.OnRecovery(name)
		}
	}

	state.mu.Unlock()
}

// GetStatus returns the current health status
func (h *HealthChecker) GetStatus() HealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// If no checks registered, consider healthy
	if len(h.checks) == 0 {
		return HealthStatusHealthy
	}

	hasUnhealthy := false
	hasDegraded := false

	for _, state := range h.checks {
		state.mu.RLock()
		status := state.status
		critical := state.config.Critical
		state.mu.RUnlock()

		switch status {
		case HealthStatusUnhealthy:
			if critical {
				return HealthStatusUnhealthy
			}
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		case HealthStatusUnknown:
			hasDegraded = true
		}
	}

	if hasUnhealthy || hasDegraded {
		return HealthStatusDegraded
	}

	return HealthStatusHealthy
}

// GetCheckResult returns the result of a specific health check
func (h *HealthChecker) GetCheckResult(name string) (HealthCheckResult, bool) {
	h.mu.RLock()
	state, exists := h.checks[name]
	h.mu.RUnlock()

	if !exists {
		return HealthCheckResult{}, false
	}

	state.mu.RLock()
	defer state.mu.RUnlock()

	result := HealthCheckResult{
		Name:                name,
		Status:              state.status,
		LastCheck:           state.lastCheck,
		LastSuccess:         state.lastSuccess,
		LastFailure:         state.lastFailure,
		ConsecutiveFailures: state.consecutiveFailures,
		Duration:            state.lastDuration,
	}

	if state.lastError != nil {
		result.Error = state.lastError.Error()
	}

	return result, true
}

// GetAllResults returns results for all health checks
func (h *HealthChecker) GetAllResults() []HealthCheckResult {
	h.mu.RLock()
	defer h.mu.RUnlock()

	results := make([]HealthCheckResult, 0, len(h.checks))
	for name, state := range h.checks {
		state.mu.RLock()
		result := HealthCheckResult{
			Name:                name,
			Status:              state.status,
			LastCheck:           state.lastCheck,
			LastSuccess:         state.lastSuccess,
			LastFailure:         state.lastFailure,
			ConsecutiveFailures: state.consecutiveFailures,
			Duration:            state.lastDuration,
		}
		if state.lastError != nil {
			result.Error = state.lastError.Error()
		}
		state.mu.RUnlock()
		results = append(results, result)
	}

	return results
}

// IsHealthy returns true if all critical checks are healthy
func (h *HealthChecker) IsHealthy() bool {
	return h.GetStatus() == HealthStatusHealthy
}
