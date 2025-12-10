package resilience

import (
	"fmt"
	"sync"
	"time"

	"freightliner/pkg/helper/log"
)

// State represents the state of a circuit breaker
type State int

const (
	// StateClosed - circuit is closed, requests are allowed
	StateClosed State = iota
	// StateOpen - circuit is open, requests are blocked
	StateOpen
	// StateHalfOpen - circuit is testing if service recovered
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerSettings configures circuit breaker behavior
type CircuitBreakerSettings struct {
	// Name of the circuit breaker (usually registry name)
	Name string
	// MaxRequests in half-open state before transitioning
	MaxRequests uint32
	// Interval to reset failure counts
	Interval time.Duration
	// Timeout before transitioning from open to half-open
	Timeout time.Duration
	// FailureThreshold - ratio of failures to trip circuit (0.0-1.0)
	FailureThreshold float64
	// MinRequests before considering failure threshold
	MinRequests uint32
	// OnStateChange callback when state changes
	OnStateChange func(name string, from State, to State)
}

// DefaultCircuitBreakerSettings returns sensible defaults
func DefaultCircuitBreakerSettings(name string) CircuitBreakerSettings {
	return CircuitBreakerSettings{
		Name:             name,
		MaxRequests:      3,
		Interval:         10 * time.Second,
		Timeout:          30 * time.Second,
		FailureThreshold: 0.6,
		MinRequests:      3,
	}
}

// Counts holds circuit breaker statistics
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	settings CircuitBreakerSettings
	state    State
	counts   Counts
	expiry   time.Time
	mu       sync.RWMutex
	logger   log.Logger
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(settings CircuitBreakerSettings, logger log.Logger) *CircuitBreaker {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	cb := &CircuitBreaker{
		settings: settings,
		state:    StateClosed,
		expiry:   time.Now().Add(settings.Interval),
		logger:   logger,
	}

	return cb
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// Check if request is allowed
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	// Execute the function
	err := fn()

	// Record the result
	cb.afterRequest(err == nil)

	return err
}

// beforeRequest checks if the request should be allowed
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state := cb.state

	// Check if we need to reset counts based on interval
	if state == StateClosed && cb.expiry.Before(now) {
		cb.resetCounts(now)
	}

	switch state {
	case StateOpen:
		// Check if timeout has elapsed
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
			return nil
		}
		return fmt.Errorf("circuit breaker '%s' is open", cb.settings.Name)

	case StateHalfOpen:
		// Allow limited requests in half-open state
		if cb.counts.Requests >= cb.settings.MaxRequests {
			return fmt.Errorf("circuit breaker '%s' is half-open and at max requests", cb.settings.Name)
		}
		return nil

	case StateClosed:
		return nil

	default:
		return fmt.Errorf("circuit breaker '%s' in unknown state", cb.settings.Name)
	}
}

// afterRequest records the result of a request
func (cb *CircuitBreaker) afterRequest(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state := cb.state

	cb.counts.Requests++

	if success {
		cb.counts.TotalSuccesses++
		cb.counts.ConsecutiveSuccesses++
		cb.counts.ConsecutiveFailures = 0

		if state == StateHalfOpen {
			// Enough successful requests in half-open, close the circuit
			if cb.counts.ConsecutiveSuccesses >= cb.settings.MaxRequests {
				cb.setState(StateClosed, now)
			}
		}
	} else {
		cb.counts.TotalFailures++
		cb.counts.ConsecutiveFailures++
		cb.counts.ConsecutiveSuccesses = 0

		if state == StateHalfOpen {
			// Failed in half-open, reopen the circuit
			cb.setState(StateOpen, now)
		} else if state == StateClosed {
			// Check if we should trip the circuit
			if cb.shouldTrip() {
				cb.setState(StateOpen, now)
			}
		}
	}
}

// shouldTrip determines if the circuit should trip based on failure rate
func (cb *CircuitBreaker) shouldTrip() bool {
	counts := cb.counts
	settings := cb.settings

	// Need minimum requests before we can trip
	if counts.Requests < settings.MinRequests {
		return false
	}

	// Calculate failure ratio
	failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)

	return failureRatio >= settings.FailureThreshold
}

// setState transitions the circuit breaker to a new state
func (cb *CircuitBreaker) setState(state State, now time.Time) {
	if cb.state == state {
		return
	}

	oldState := cb.state
	cb.state = state

	// Set expiry based on new state
	switch state {
	case StateClosed:
		cb.expiry = now.Add(cb.settings.Interval)
		cb.resetCounts(now)
	case StateOpen:
		cb.expiry = now.Add(cb.settings.Timeout)
	case StateHalfOpen:
		cb.expiry = now.Add(cb.settings.Timeout)
		cb.counts.Requests = 0
		cb.counts.ConsecutiveSuccesses = 0
		cb.counts.ConsecutiveFailures = 0
	}

	// Log state change
	cb.logger.WithFields(map[string]interface{}{
		"circuitBreaker": cb.settings.Name,
		"oldState":       oldState.String(),
		"newState":       state.String(),
	}).Info("Circuit breaker state changed")

	// Call callback if set
	if cb.settings.OnStateChange != nil {
		go cb.settings.OnStateChange(cb.settings.Name, oldState, state)
	}
}

// resetCounts resets the circuit breaker counts
func (cb *CircuitBreaker) resetCounts(now time.Time) {
	cb.counts = Counts{}
	cb.expiry = now.Add(cb.settings.Interval)
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Counts returns a copy of current counts
func (cb *CircuitBreaker) Counts() Counts {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.counts
}

// Name returns the circuit breaker name
func (cb *CircuitBreaker) Name() string {
	return cb.settings.Name
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex
	logger   log.Logger
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(logger log.Logger) *CircuitBreakerManager {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
		logger:   logger,
	}
}

// GetOrCreate gets an existing circuit breaker or creates a new one
func (m *CircuitBreakerManager) GetOrCreate(name string, settings CircuitBreakerSettings) *CircuitBreaker {
	m.mu.RLock()
	breaker, exists := m.breakers[name]
	m.mu.RUnlock()

	if exists {
		return breaker
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if breaker, exists := m.breakers[name]; exists {
		return breaker
	}

	// Create new circuit breaker
	settings.Name = name
	breaker = NewCircuitBreaker(settings, m.logger)
	m.breakers[name] = breaker

	return breaker
}

// Get retrieves a circuit breaker by name
func (m *CircuitBreakerManager) Get(name string) (*CircuitBreaker, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	breaker, exists := m.breakers[name]
	return breaker, exists
}

// Execute runs a function with circuit breaker protection
func (m *CircuitBreakerManager) Execute(name string, fn func() error) error {
	settings := DefaultCircuitBreakerSettings(name)
	breaker := m.GetOrCreate(name, settings)
	return breaker.Execute(fn)
}

// GetAllStates returns the state of all circuit breakers
func (m *CircuitBreakerManager) GetAllStates() map[string]State {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states := make(map[string]State, len(m.breakers))
	for name, breaker := range m.breakers {
		states[name] = breaker.State()
	}
	return states
}

// GetAllCounts returns counts for all circuit breakers
func (m *CircuitBreakerManager) GetAllCounts() map[string]Counts {
	m.mu.RLock()
	defer m.mu.RUnlock()

	counts := make(map[string]Counts, len(m.breakers))
	for name, breaker := range m.breakers {
		counts[name] = breaker.Counts()
	}
	return counts
}
