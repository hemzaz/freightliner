package throttle

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	tests := []struct {
		name         string
		requestLimit int
		timeWindow   time.Duration
	}{
		{
			name:         "Basic limiter",
			requestLimit: 10,
			timeWindow:   time.Second,
		},
		{
			name:         "High rate limiter",
			requestLimit: 100,
			timeWindow:   time.Second,
		},
		{
			name:         "Long window limiter",
			requestLimit: 5,
			timeWindow:   time.Minute,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			limiter := NewRateLimiter(tc.requestLimit, tc.timeWindow)

			if limiter == nil {
				t.Fatal("Expected non-nil rate limiter")
			}

			// Check token channel capacity
			if cap(limiter.tokens) != tc.requestLimit {
				t.Errorf("Expected token capacity %d, got %d", tc.requestLimit, cap(limiter.tokens))
			}

			// Check fields are properly set
			if limiter.requestLimit != tc.requestLimit {
				t.Errorf("Expected request limit %d, got %d", tc.requestLimit, limiter.requestLimit)
			}

			if limiter.timeWindow != tc.timeWindow {
				t.Errorf("Expected time window %v, got %v", tc.timeWindow, limiter.timeWindow)
			}
		})
	}
}

func TestAcquireWithinLimit(t *testing.T) {
	// Create a rate limiter with 5 requests per second
	limiter := NewRateLimiter(5, time.Second)

	// Should be able to make 5 requests immediately without blocking
	for i := 0; i < 5; i++ {
		ctx := context.Background()
		err := limiter.Acquire(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	}
}

func TestAcquireExceedingLimit(t *testing.T) {
	// Create a rate limiter with 3 requests per second
	limiter := NewRateLimiter(3, time.Second)

	// Make 3 requests that should succeed immediately
	for i := 0; i < 3; i++ {
		ctx := context.Background()
		err := limiter.Acquire(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	}

	// The 4th request should block, so use a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := limiter.Acquire(ctx)
	if err == nil {
		t.Error("Expected context deadline exceeded error, got nil")
	}
}

func TestTokenReplenishment(t *testing.T) {
	// Create a rate limiter with 2 requests per 100ms
	limiter := NewRateLimiter(2, 100*time.Millisecond)

	// Consume all tokens
	for i := 0; i < 2; i++ {
		err := limiter.Acquire(context.Background())
		if err != nil {
			t.Fatalf("Failed to acquire token: %v", err)
		}
	}

	// Wait for token replenishment (a bit more than the reset interval)
	time.Sleep(60 * time.Millisecond)

	// Should be able to make another request now
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := limiter.Acquire(ctx)
	if err != nil {
		t.Errorf("Expected token to be replenished, but got error: %v", err)
	}
}

func TestContextCancellation(t *testing.T) {
	// Create a rate limiter with 1 request per second
	limiter := NewRateLimiter(1, time.Second)

	// Consume the only token
	err := limiter.Acquire(context.Background())
	if err != nil {
		t.Fatalf("Failed to acquire token: %v", err)
	}

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Start a goroutine that will attempt to acquire a token
	var wg sync.WaitGroup
	wg.Add(1)

	var acquireErr error
	go func() {
		defer wg.Done()
		acquireErr = limiter.Acquire(ctx)
	}()

	// Cancel the context after a short delay
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Wait for the goroutine to finish
	wg.Wait()

	// Check that we got a context cancelled error
	if acquireErr == nil || acquireErr.Error() != "context canceled" {
		t.Errorf("Expected 'context canceled' error, got: %v", acquireErr)
	}
}

func TestConcurrentAcquire(t *testing.T) {
	// Create a rate limiter with 10 requests per second
	limiter := NewRateLimiter(10, time.Second)

	// Launch 20 goroutines to simultaneously acquire tokens
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			err := limiter.Acquire(ctx)
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Should have successfully acquired exactly 10 tokens
	if successCount != 10 {
		t.Errorf("Expected 10 successful token acquisitions, got %d", successCount)
	}
}

func TestPerformanceUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Create a high-throughput rate limiter
	requestsPerSecond := 100
	limiter := NewRateLimiter(requestsPerSecond, time.Second)

	// Track successful acquisitions
	successCount := 0
	var mu sync.Mutex

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start time
	startTime := time.Now()

	// Launch a bunch of goroutines that continuously try to acquire tokens
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					err := limiter.Acquire(ctx)
					if err == nil {
						mu.Lock()
						successCount++
						mu.Unlock()
					} else if err.Error() != "context canceled" && err.Error() != "context deadline exceeded" {
						t.Errorf("Unexpected error: %v", err)
						return
					}
				}
			}
		}()
	}

	// Wait for context cancellation
	<-ctx.Done()

	// Wait for all goroutines to finish
	wg.Wait()

	// Calculate elapsed time
	elapsed := time.Since(startTime)
	elapsedSeconds := elapsed.Seconds()

	// Calculate rate
	rate := float64(successCount) / elapsedSeconds

	// The rate should be roughly around the configured rate limit
	// Allow for some margin of error (±50%) since test environments can be unpredictable
	minAcceptableRate := float64(requestsPerSecond) * 0.5
	maxAcceptableRate := float64(requestsPerSecond) * 1.5

	t.Logf("Rate limit: %d/s, Achieved: %.2f/s over %.2f seconds (%d requests)",
		requestsPerSecond, rate, elapsedSeconds, successCount)

	if rate < minAcceptableRate || rate > maxAcceptableRate {
		t.Errorf("Expected rate around %d/s (±50%%), got %.2f/s", requestsPerSecond, rate)
	}
}
