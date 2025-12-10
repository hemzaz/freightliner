package resilience_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"freightliner/pkg/helper/log"
	"freightliner/pkg/resilience"

	"github.com/stretchr/testify/assert"
)

func TestRetryPolicy_SuccessFirstAttempt(t *testing.T) {
	policy := resilience.DefaultRetryPolicy()
	ctx := context.Background()

	attempts := 0
	err := policy.Retry(ctx, func() error {
		attempts++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, attempts)
}

func TestRetryPolicy_SuccessAfterRetries(t *testing.T) {
	policy := resilience.DefaultRetryPolicy()
	ctx := context.Background()

	attempts := 0
	err := policy.Retry(ctx, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("transient error")
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

func TestRetryPolicy_ExhaustRetries(t *testing.T) {
	policy := &resilience.RetryPolicy{
		MaxRetries:  2,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     100 * time.Millisecond,
		Multiplier:  2.0,
		Jitter:      0.1,
	}
	ctx := context.Background()

	attempts := 0
	err := policy.Retry(ctx, func() error {
		attempts++
		return errors.New("persistent error")
	})

	assert.Error(t, err)
	assert.Equal(t, 3, attempts) // Initial + 2 retries
	assert.Contains(t, err.Error(), "max retries exceeded")
}

func TestRetryPolicy_ContextCancellation(t *testing.T) {
	policy := resilience.DefaultRetryPolicy()
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after first attempt
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	attempts := 0
	err := policy.Retry(ctx, func() error {
		attempts++
		time.Sleep(50 * time.Millisecond)
		return errors.New("error")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "retry cancelled")
}

func TestRetryPolicy_NonRetryableError(t *testing.T) {
	policy := &resilience.RetryPolicy{
		MaxRetries:  3,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     100 * time.Millisecond,
		Multiplier:  2.0,
		RetryableErrors: func(err error) bool {
			return err.Error() != "non-retryable"
		},
	}
	ctx := context.Background()

	attempts := 0
	err := policy.Retry(ctx, func() error {
		attempts++
		return errors.New("non-retryable")
	})

	assert.Error(t, err)
	assert.Equal(t, 1, attempts) // Only first attempt
	assert.Contains(t, err.Error(), "non-retryable error")
}

func TestRetryPolicy_OnRetryCallback(t *testing.T) {
	callbackCount := 0
	policy := &resilience.RetryPolicy{
		MaxRetries:  2,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     100 * time.Millisecond,
		Multiplier:  2.0,
		OnRetry: func(attempt int, err error, wait time.Duration) {
			callbackCount++
		},
	}
	ctx := context.Background()

	policy.Retry(ctx, func() error {
		return errors.New("error")
	})

	assert.Equal(t, 2, callbackCount) // Called for each retry
}

func TestRetryWithResult(t *testing.T) {
	policy := resilience.DefaultRetryPolicy()
	ctx := context.Background()
	logger := log.NewBasicLogger(log.InfoLevel)

	attempts := 0
	result, err := resilience.RetryWithResult(ctx, policy, func() (string, error) {
		attempts++
		if attempts < 2 {
			return "", errors.New("error")
		}
		return "success", nil
	}, logger)

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, 2, attempts)
}

func TestRetryManager(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	manager := resilience.NewRetryManager(logger)

	// Set custom policy
	customPolicy := &resilience.RetryPolicy{
		MaxRetries:  1,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     100 * time.Millisecond,
		Multiplier:  2.0,
	}
	manager.SetPolicy("custom", customPolicy)

	// Use custom policy
	attempts := 0
	err := manager.Retry(context.Background(), "custom", func() error {
		attempts++
		return errors.New("error")
	})

	assert.Error(t, err)
	assert.Equal(t, 2, attempts) // 1 retry

	// Unknown policy should use default
	attempts = 0
	err = manager.Retry(context.Background(), "unknown", func() error {
		attempts++
		if attempts < 3 {
			return errors.New("error")
		}
		return nil
	})

	assert.NoError(t, err)
}

func TestAggressiveRetryPolicy(t *testing.T) {
	policy := resilience.AggressiveRetryPolicy()
	assert.Equal(t, 5, policy.MaxRetries)
	assert.Equal(t, 50*time.Millisecond, policy.InitialWait)
}

func TestConservativeRetryPolicy(t *testing.T) {
	policy := resilience.ConservativeRetryPolicy()
	assert.Equal(t, 2, policy.MaxRetries)
	assert.Equal(t, 500*time.Millisecond, policy.InitialWait)
}
