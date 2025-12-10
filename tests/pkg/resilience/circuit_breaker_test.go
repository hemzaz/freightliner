package resilience_test

import (
	"errors"
	"testing"
	"time"

	"freightliner/pkg/helper/log"
	"freightliner/pkg/resilience"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreaker_ClosedState(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	settings := resilience.DefaultCircuitBreakerSettings("test")
	cb := resilience.NewCircuitBreaker(settings, logger)

	// Initially closed
	assert.Equal(t, resilience.StateClosed, cb.State())

	// Successful requests should keep it closed
	for i := 0; i < 10; i++ {
		err := cb.Execute(func() error {
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, resilience.StateClosed, cb.State())
	}
}

func TestCircuitBreaker_TripsOnFailures(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	settings := resilience.CircuitBreakerSettings{
		Name:             "test",
		MaxRequests:      3,
		Interval:         10 * time.Second,
		Timeout:          100 * time.Millisecond,
		FailureThreshold: 0.6,
		MinRequests:      3,
	}
	cb := resilience.NewCircuitBreaker(settings, logger)

	// Generate failures to trip circuit
	for i := 0; i < 5; i++ {
		err := cb.Execute(func() error {
			return errors.New("simulated failure")
		})
		assert.Error(t, err)
	}

	// Circuit should now be open
	assert.Equal(t, resilience.StateOpen, cb.State())

	// Further requests should be rejected
	err := cb.Execute(func() error {
		return nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker")
}

func TestCircuitBreaker_RecoveryToHalfOpen(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	settings := resilience.CircuitBreakerSettings{
		Name:             "test",
		MaxRequests:      2,
		Interval:         10 * time.Second,
		Timeout:          50 * time.Millisecond,
		FailureThreshold: 0.6,
		MinRequests:      3,
	}
	cb := resilience.NewCircuitBreaker(settings, logger)

	// Trip the circuit
	for i := 0; i < 5; i++ {
		cb.Execute(func() error {
			return errors.New("failure")
		})
	}
	assert.Equal(t, resilience.StateOpen, cb.State())

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Next request should transition to half-open
	err := cb.Execute(func() error {
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, resilience.StateHalfOpen, cb.State())
}

func TestCircuitBreaker_RecoveryToClosed(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	settings := resilience.CircuitBreakerSettings{
		Name:             "test",
		MaxRequests:      2,
		Interval:         10 * time.Second,
		Timeout:          50 * time.Millisecond,
		FailureThreshold: 0.6,
		MinRequests:      3,
	}
	cb := resilience.NewCircuitBreaker(settings, logger)

	// Trip the circuit
	for i := 0; i < 5; i++ {
		cb.Execute(func() error {
			return errors.New("failure")
		})
	}

	// Wait and recover
	time.Sleep(60 * time.Millisecond)

	// Successful requests in half-open should close circuit
	for i := 0; i < 3; i++ {
		err := cb.Execute(func() error {
			return nil
		})
		assert.NoError(t, err)
	}

	assert.Equal(t, resilience.StateClosed, cb.State())
}

func TestCircuitBreakerManager(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	manager := resilience.NewCircuitBreakerManager(logger)

	// Create multiple circuit breakers
	cb1 := manager.GetOrCreate("registry1", resilience.DefaultCircuitBreakerSettings("registry1"))
	cb2 := manager.GetOrCreate("registry2", resilience.DefaultCircuitBreakerSettings("registry2"))

	assert.NotNil(t, cb1)
	assert.NotNil(t, cb2)
	assert.NotEqual(t, cb1, cb2)

	// Getting same breaker should return same instance
	cb1Again := manager.GetOrCreate("registry1", resilience.DefaultCircuitBreakerSettings("registry1"))
	assert.Equal(t, cb1, cb1Again)

	// Test execution through manager
	err := manager.Execute("registry1", func() error {
		return nil
	})
	assert.NoError(t, err)
}

func TestCircuitBreaker_StateCallback(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)

	stateChanges := make([]string, 0)
	settings := resilience.CircuitBreakerSettings{
		Name:             "test",
		MaxRequests:      2,
		Interval:         10 * time.Second,
		Timeout:          50 * time.Millisecond,
		FailureThreshold: 0.6,
		MinRequests:      3,
		OnStateChange: func(name string, from, to resilience.State) {
			stateChanges = append(stateChanges, to.String())
		},
	}
	cb := resilience.NewCircuitBreaker(settings, logger)

	// Trip circuit
	for i := 0; i < 5; i++ {
		cb.Execute(func() error {
			return errors.New("failure")
		})
	}

	// Wait for state change callback
	time.Sleep(10 * time.Millisecond)
	require.Contains(t, stateChanges, "open")
}
