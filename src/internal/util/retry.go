package util

import (
	"context"
	"errors"
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
