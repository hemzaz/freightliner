package resilience

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"freightliner/pkg/helper/log"
)

// RetryPolicy defines retry behavior with exponential backoff and jitter
type RetryPolicy struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
	// InitialWait is the initial wait time before the first retry
	InitialWait time.Duration
	// MaxWait is the maximum wait time between retries
	MaxWait time.Duration
	// Multiplier for exponential backoff (typically 2.0)
	Multiplier float64
	// Jitter adds randomization to prevent thundering herd (0.0-1.0)
	Jitter float64
	// RetryableErrors is a function that determines if an error is retryable
	RetryableErrors func(error) bool
	// OnRetry is called before each retry attempt
	OnRetry func(attempt int, err error, wait time.Duration)
}

// DefaultRetryPolicy returns a sensible default retry policy
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:  3,
		InitialWait: 100 * time.Millisecond,
		MaxWait:     30 * time.Second,
		Multiplier:  2.0,
		Jitter:      0.5,
		RetryableErrors: func(err error) bool {
			// By default, retry all errors
			return err != nil
		},
	}
}

// AggressiveRetryPolicy returns a retry policy for critical operations
func AggressiveRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:  5,
		InitialWait: 50 * time.Millisecond,
		MaxWait:     1 * time.Minute,
		Multiplier:  2.0,
		Jitter:      0.5,
		RetryableErrors: func(err error) bool {
			return err != nil
		},
	}
}

// ConservativeRetryPolicy returns a retry policy for non-critical operations
func ConservativeRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:  2,
		InitialWait: 500 * time.Millisecond,
		MaxWait:     10 * time.Second,
		Multiplier:  2.0,
		Jitter:      0.3,
		RetryableErrors: func(err error) bool {
			return err != nil
		},
	}
}

// Retry executes the operation with retry logic
func (r *RetryPolicy) Retry(ctx context.Context, operation func() error) error {
	return r.RetryWithLogger(ctx, operation, nil)
}

// RetryWithLogger executes the operation with retry logic and logging
func (r *RetryPolicy) RetryWithLogger(ctx context.Context, operation func() error, logger log.Logger) error {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	var lastErr error

	for attempt := 0; attempt <= r.MaxRetries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		default:
		}

		// Execute the operation
		err := operation()
		if err == nil {
			// Success!
			if attempt > 0 {
				logger.WithFields(map[string]interface{}{
					"attempt": attempt + 1,
				}).Info("Operation succeeded after retries")
			}
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if r.RetryableErrors != nil && !r.RetryableErrors(err) {
			logger.WithFields(map[string]interface{}{
				"attempt": attempt + 1,
			}).Debug("Error is not retryable, giving up")
			return fmt.Errorf("non-retryable error: %w", err)
		}

		// Check if we've exhausted retries
		if attempt >= r.MaxRetries {
			logger.WithError(err).WithFields(map[string]interface{}{
				"attempts": attempt + 1,
			}).Error("Maximum retries exhausted", err)
			return fmt.Errorf("max retries exceeded (%d attempts): %w", attempt+1, err)
		}

		// Calculate wait time with exponential backoff and jitter
		wait := r.calculateBackoff(attempt)

		logger.WithError(err).WithFields(map[string]interface{}{
			"attempt":  attempt + 1,
			"waitTime": wait.String(),
		}).Warn("Operation failed, retrying")

		// Call OnRetry callback if set
		if r.OnRetry != nil {
			r.OnRetry(attempt+1, err, wait)
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled while waiting: %w", ctx.Err())
		case <-time.After(wait):
			// Continue to next attempt
		}
	}

	return lastErr
}

// calculateBackoff calculates the wait time for a given attempt
func (r *RetryPolicy) calculateBackoff(attempt int) time.Duration {
	// Calculate exponential backoff: InitialWait * (Multiplier ^ attempt)
	backoff := float64(r.InitialWait) * math.Pow(r.Multiplier, float64(attempt))

	// Cap at MaxWait
	if backoff > float64(r.MaxWait) {
		backoff = float64(r.MaxWait)
	}

	// Add jitter to prevent thundering herd
	if r.Jitter > 0 {
		// Random jitter between -Jitter and +Jitter
		jitterRange := backoff * r.Jitter
		jitter := (rand.Float64() * 2 * jitterRange) - jitterRange
		backoff += jitter
	}

	// Ensure non-negative
	if backoff < 0 {
		backoff = 0
	}

	return time.Duration(backoff)
}

// RetryWithResult executes an operation that returns a result and an error
func RetryWithResult[T any](ctx context.Context, policy *RetryPolicy, operation func() (T, error), logger log.Logger) (T, error) {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	var result T
	var lastErr error

	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return result, fmt.Errorf("retry cancelled: %w", ctx.Err())
		default:
		}

		// Execute the operation
		res, err := operation()
		if err == nil {
			if attempt > 0 {
				logger.WithFields(map[string]interface{}{
					"attempt": attempt + 1,
				}).Info("Operation succeeded after retries")
			}
			return res, nil
		}

		lastErr = err

		// Check if error is retryable
		if policy.RetryableErrors != nil && !policy.RetryableErrors(err) {
			logger.WithFields(map[string]interface{}{
				"attempt": attempt + 1,
			}).Debug("Error is not retryable, giving up")
			return result, fmt.Errorf("non-retryable error: %w", err)
		}

		// Check if we've exhausted retries
		if attempt >= policy.MaxRetries {
			logger.WithError(err).WithFields(map[string]interface{}{
				"attempts": attempt + 1,
			}).Error("Maximum retries exhausted", err)
			return result, fmt.Errorf("max retries exceeded (%d attempts): %w", attempt+1, err)
		}

		// Calculate wait time
		wait := policy.calculateBackoff(attempt)

		logger.WithError(err).WithFields(map[string]interface{}{
			"attempt":  attempt + 1,
			"waitTime": wait.String(),
		}).Warn("Operation failed, retrying")

		// Call OnRetry callback if set
		if policy.OnRetry != nil {
			policy.OnRetry(attempt+1, err, wait)
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			return result, fmt.Errorf("retry cancelled while waiting: %w", ctx.Err())
		case <-time.After(wait):
			// Continue to next attempt
		}
	}

	return result, lastErr
}

// RetryManager manages retry policies for different operations
type RetryManager struct {
	policies map[string]*RetryPolicy
	logger   log.Logger
}

// NewRetryManager creates a new retry manager
func NewRetryManager(logger log.Logger) *RetryManager {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	return &RetryManager{
		policies: make(map[string]*RetryPolicy),
		logger:   logger,
	}
}

// SetPolicy sets a retry policy for a specific operation
func (m *RetryManager) SetPolicy(name string, policy *RetryPolicy) {
	m.policies[name] = policy
}

// GetPolicy retrieves a retry policy by name
func (m *RetryManager) GetPolicy(name string) *RetryPolicy {
	if policy, exists := m.policies[name]; exists {
		return policy
	}
	return DefaultRetryPolicy()
}

// Retry executes an operation with the specified retry policy
func (m *RetryManager) Retry(ctx context.Context, policyName string, operation func() error) error {
	policy := m.GetPolicy(policyName)
	return policy.RetryWithLogger(ctx, operation, m.logger)
}
