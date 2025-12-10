package throttle

import (
	"context"
	"time"
)

// RateLimiter provides rate limiting capabilities for API calls
type RateLimiter struct {
	tokens        chan struct{}
	requestLimit  int
	timeWindow    time.Duration
	resetInterval time.Duration
}

// NewRateLimiter creates a new rate limiter with the specified parameters
func NewRateLimiter(requestLimit int, timeWindow time.Duration) *RateLimiter {
	tokens := make(chan struct{}, requestLimit)

	// Fill the token bucket
	for i := 0; i < requestLimit; i++ {
		tokens <- struct{}{}
	}

	limiter := &RateLimiter{
		tokens:        tokens,
		requestLimit:  requestLimit,
		timeWindow:    timeWindow,
		resetInterval: timeWindow / time.Duration(requestLimit),
	}

	// Start token replenishment
	go limiter.replenishTokens()

	return limiter
}

// Acquire acquires a token for an API call, blocking if necessary
func (r *RateLimiter) Acquire(ctx context.Context) error {
	select {
	case <-r.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// replenishTokens continuously replenishes tokens at the appropriate rate
func (r *RateLimiter) replenishTokens() {
	ticker := time.NewTicker(r.resetInterval)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case r.tokens <- struct{}{}:
			// Successfully replenished a token
		default:
			// Token bucket is full
		}
	}
}
