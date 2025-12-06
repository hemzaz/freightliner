package resilience_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"freightliner/pkg/helper/log"
	"freightliner/pkg/resilience"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthChecker_InitialState(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	checker := resilience.NewHealthChecker(logger)

	// With no checks, system should be healthy
	assert.Equal(t, resilience.HealthStatusHealthy, checker.GetStatus())
	assert.True(t, checker.IsHealthy())
}

func TestHealthChecker_RegisterCheck(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	checker := resilience.NewHealthChecker(logger)

	check := resilience.HealthCheck{
		Name: "test-check",
		Check: func(ctx context.Context) error {
			return nil
		},
		Interval: 100 * time.Millisecond,
		Timeout:  50 * time.Millisecond,
	}

	err := checker.RegisterCheck(check)
	assert.NoError(t, err)

	// Registering same check again should fail
	err = checker.RegisterCheck(check)
	assert.Error(t, err)
}

func TestHealthChecker_HealthyCheck(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	checker := resilience.NewHealthChecker(logger)

	check := resilience.HealthCheck{
		Name: "healthy-check",
		Check: func(ctx context.Context) error {
			return nil
		},
		Interval: 50 * time.Millisecond,
		Timeout:  25 * time.Millisecond,
		Critical: true,
	}

	err := checker.RegisterCheck(check)
	require.NoError(t, err)

	checker.Start()
	defer checker.Stop()

	// Wait for check to run
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, resilience.HealthStatusHealthy, checker.GetStatus())

	result, exists := checker.GetCheckResult("healthy-check")
	assert.True(t, exists)
	assert.Equal(t, resilience.HealthStatusHealthy, result.Status)
	assert.Equal(t, 0, result.ConsecutiveFailures)
}

func TestHealthChecker_UnhealthyCheck(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	checker := resilience.NewHealthChecker(logger)

	check := resilience.HealthCheck{
		Name: "unhealthy-check",
		Check: func(ctx context.Context) error {
			return errors.New("service unavailable")
		},
		Interval: 50 * time.Millisecond,
		Timeout:  25 * time.Millisecond,
		Critical: true,
	}

	err := checker.RegisterCheck(check)
	require.NoError(t, err)

	checker.Start()
	defer checker.Stop()

	// Wait for check to run
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, resilience.HealthStatusUnhealthy, checker.GetStatus())

	result, exists := checker.GetCheckResult("unhealthy-check")
	assert.True(t, exists)
	assert.Equal(t, resilience.HealthStatusUnhealthy, result.Status)
	assert.Greater(t, result.ConsecutiveFailures, 0)
	assert.NotEmpty(t, result.Error)
}

func TestHealthChecker_DegradedState(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	checker := resilience.NewHealthChecker(logger)

	// Critical check passes
	criticalCheck := resilience.HealthCheck{
		Name: "critical-check",
		Check: func(ctx context.Context) error {
			return nil
		},
		Interval: 50 * time.Millisecond,
		Timeout:  25 * time.Millisecond,
		Critical: true,
	}

	// Non-critical check fails
	nonCriticalCheck := resilience.HealthCheck{
		Name: "non-critical-check",
		Check: func(ctx context.Context) error {
			return errors.New("non-critical failure")
		},
		Interval: 50 * time.Millisecond,
		Timeout:  25 * time.Millisecond,
		Critical: false,
	}

	checker.RegisterCheck(criticalCheck)
	checker.RegisterCheck(nonCriticalCheck)

	checker.Start()
	defer checker.Stop()

	// Wait for checks to run
	time.Sleep(100 * time.Millisecond)

	// System should be degraded (critical healthy, non-critical unhealthy)
	assert.Equal(t, resilience.HealthStatusDegraded, checker.GetStatus())
	assert.False(t, checker.IsHealthy())
}

func TestHealthChecker_OnFailureCallback(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	checker := resilience.NewHealthChecker(logger)

	failureCalled := false
	check := resilience.HealthCheck{
		Name: "callback-check",
		Check: func(ctx context.Context) error {
			return errors.New("failure")
		},
		Interval: 50 * time.Millisecond,
		Timeout:  25 * time.Millisecond,
		OnFailure: func(name string, err error) {
			failureCalled = true
		},
	}

	checker.RegisterCheck(check)
	checker.Start()
	defer checker.Stop()

	// Wait for check to run and callback to fire
	time.Sleep(100 * time.Millisecond)

	assert.True(t, failureCalled)
}

func TestHealthChecker_OnRecoveryCallback(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	checker := resilience.NewHealthChecker(logger)

	failures := 0
	recoveryCalled := false

	check := resilience.HealthCheck{
		Name: "recovery-check",
		Check: func(ctx context.Context) error {
			failures++
			if failures <= 2 {
				return errors.New("failure")
			}
			return nil
		},
		Interval: 50 * time.Millisecond,
		Timeout:  25 * time.Millisecond,
		OnRecovery: func(name string) {
			recoveryCalled = true
		},
	}

	checker.RegisterCheck(check)
	checker.Start()
	defer checker.Stop()

	// Wait for recovery
	time.Sleep(200 * time.Millisecond)

	assert.True(t, recoveryCalled)
}

func TestHealthChecker_GetAllResults(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	checker := resilience.NewHealthChecker(logger)

	for i := 0; i < 3; i++ {
		check := resilience.HealthCheck{
			Name: "check-" + string(rune(i)),
			Check: func(ctx context.Context) error {
				return nil
			},
			Interval: 100 * time.Millisecond,
			Timeout:  50 * time.Millisecond,
		}
		checker.RegisterCheck(check)
	}

	checker.Start()
	defer checker.Stop()

	time.Sleep(150 * time.Millisecond)

	results := checker.GetAllResults()
	assert.Len(t, results, 3)

	for _, result := range results {
		assert.Equal(t, resilience.HealthStatusHealthy, result.Status)
	}
}

func TestHealthChecker_CheckTimeout(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	checker := resilience.NewHealthChecker(logger)

	check := resilience.HealthCheck{
		Name: "timeout-check",
		Check: func(ctx context.Context) error {
			// Wait for context timeout
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(200 * time.Millisecond):
				return nil
			}
		},
		Interval: 200 * time.Millisecond,
		Timeout:  20 * time.Millisecond, // Shorter than check duration
		Critical: true,
	}

	checker.RegisterCheck(check)
	checker.Start()
	defer checker.Stop()

	// Wait for check to run and timeout
	time.Sleep(300 * time.Millisecond)

	// Check should timeout and be marked unhealthy
	result, exists := checker.GetCheckResult("timeout-check")
	assert.True(t, exists)
	assert.Equal(t, resilience.HealthStatusUnhealthy, result.Status)
	assert.Contains(t, result.Error, "context deadline exceeded")
}
