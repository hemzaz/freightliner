package util

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// RetryOptions configures the retry behavior
type RetryOptions struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
	Factor      float64
	Retryable   func(error) bool
}

// DefaultRetryOptions returns sensible default retry options
func DefaultRetryOptions() RetryOptions {
	return RetryOptions{
		MaxRetries:  5,
		InitialWait: 1 * time.Second,
		MaxWait:     60 * time.Second,
		Factor:      2.0,
		Retryable:   func(err error) bool { return true },
	}
}

// RetryableFunc represents a function that can be retried
type RetryableFunc func() error

// RetryWithContext retries the given function with exponential backoff
func RetryWithContext(ctx context.Context, fn RetryableFunc, opts RetryOptions) error {
	var err error
	wait := opts.InitialWait

	for attempt := 0; attempt <= opts.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait with exponential backoff
			select {
			case <-time.After(wait):
				// Continue with retry
			case <-ctx.Done():
				return errors.New("retry aborted by context cancellation")
			}

			// Increase wait time for next attempt
			wait = time.Duration(float64(wait) * opts.Factor)
			if wait > opts.MaxWait {
				wait = opts.MaxWait
			}
		}

		err = fn()
		if err == nil {
			return nil // Success
		}

		if !opts.Retryable(err) {
			return err // Non-retryable error
		}
	}

	return err // Return the last error after all retries
}

// RetryWithBackoff provides a simpler interface for retrying operations with backoff
// It uses sensible defaults and allows specifying just the key parameters
func RetryWithBackoff(ctx context.Context, maxRetries int, initialWait, maxWait time.Duration, fn RetryableFunc) error {
	return RetryWithContext(ctx, fn, RetryOptions{
		MaxRetries:  maxRetries,
		InitialWait: initialWait,
		MaxWait:     maxWait,
		Factor:      2.0,                                  // Standard exponential backoff multiplier
		Retryable:   func(err error) bool { return true }, // Retry all errors by default
	})
}

// RetryWithBackoffAndLogger wraps RetryWithBackoff and adds logging
// This is useful for operations where you want to log each retry attempt
type Logger interface {
	Debug(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
}

func RetryWithBackoffAndLogger(
	ctx context.Context,
	maxRetries int,
	initialWait,
	maxWait time.Duration,
	logger Logger,
	operationName string,
	fn RetryableFunc,
) error {
	var lastErr error

	err := RetryWithContext(ctx, func() error {
		err := fn()
		if err != nil {
			lastErr = err
			logger.Warn(fmt.Sprintf("Operation '%s' failed, will retry", operationName), map[string]interface{}{
				"error": err.Error(),
			})
			return err
		}
		return nil
	}, RetryOptions{
		MaxRetries:  maxRetries,
		InitialWait: initialWait,
		MaxWait:     maxWait,
		Factor:      2.0,
		Retryable:   func(err error) bool { return true },
	})

	if err != nil && err == lastErr {
		// Only log the final failure if we've exhausted all retries
		logger.Warn(fmt.Sprintf("Operation '%s' failed permanently after %d retries", operationName, maxRetries), map[string]interface{}{
			"error": err.Error(),
		})
	} else if err == nil && lastErr != nil {
		// Log success after previous failures
		logger.Debug(fmt.Sprintf("Operation '%s' succeeded after retries", operationName), nil)
	}

	return err
}
