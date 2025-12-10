package throttle

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestRaceLimiter_ConcurrentAcquire tests concurrent requests with race detection
func TestRaceLimiter_ConcurrentAcquire(t *testing.T) {
	// Create a limiter with moderate capacity
	limiter := NewRateLimiter(50, time.Second)

	// Run a large number of goroutines that all try to acquire tokens
	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	var successCount atomic.Int32
	var errorCount atomic.Int32

	// Launch goroutines
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			// Try to acquire token
			err := limiter.Acquire(ctx)
			if err != nil {
				errorCount.Add(1)
			} else {
				successCount.Add(1)

				// Do some "work" with the token
				time.Sleep(time.Duration(id%10) * time.Millisecond)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify that we got approximately the right number of successes
	// (should be around the limiter capacity)
	if successCount.Load() < 20 || successCount.Load() > 80 {
		t.Errorf("Expected around 50 successful acquisitions, got %d successes and %d errors",
			successCount.Load(), errorCount.Load())
	}

	t.Logf("Acquired %d tokens successfully, %d failed due to timeout",
		successCount.Load(), errorCount.Load())
}

// TestRaceLimiter_ConcurrentAcquireAndReplenish tests the race detection for
// concurrent acquires while the replenishment goroutine is running
func TestRaceLimiter_ConcurrentAcquireAndReplenish(t *testing.T) {
	// Create a limiter with a small capacity and fast replenishment
	limiter := NewRateLimiter(5, 50*time.Millisecond)

	// Run for a fixed duration
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Track successful acquisitions
	var acquisitions atomic.Int32

	// Launch multiple goroutines that continuously try to acquire tokens
	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			for {
				// Check if we should exit
				select {
				case <-ctx.Done():
					return
				default:
					// Continue
				}

				// Try to acquire with a short timeout
				acquireCtx, acquireCancel := context.WithTimeout(ctx, 10*time.Millisecond)
				err := limiter.Acquire(acquireCtx)
				acquireCancel()

				if err == nil {
					acquisitions.Add(1)
					// Simulate using the resource briefly
					time.Sleep(time.Millisecond)
				}
			}
		}()
	}

	// Wait for the test duration to complete
	wg.Wait()

	// We should have been able to acquire a reasonable number of tokens
	// due to replenishment, but exact numbers depend on timing
	acquired := acquisitions.Load()
	if acquired < 10 {
		t.Errorf("Expected at least 10 successful acquisitions due to replenishment, got %d", acquired)
	}

	t.Logf("Acquired %d tokens over the test duration with replenishment", acquired)
}

// TestRaceLimiter_MultipleTokenStreams tests multiple token consumers and the token
// replenishment goroutine for race conditions
func TestRaceLimiter_MultipleTokenStreams(t *testing.T) {
	// Create limiters with different capacities and rates
	fastLimiter := NewRateLimiter(10, 100*time.Millisecond) // 100 tokens/sec
	slowLimiter := NewRateLimiter(5, 500*time.Millisecond)  // 10 tokens/sec

	// Run for a fixed duration
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Track acquisitions for each limiter
	var fastAcquisitions atomic.Int32
	var slowAcquisitions atomic.Int32

	// Consumer function that tries to acquire from a limiter
	consumeTokens := func(limiter *RateLimiter, counter *atomic.Int32, wg *sync.WaitGroup) {
		defer wg.Done()

		for {
			// Check if we should exit
			select {
			case <-ctx.Done():
				return
			default:
				// Continue
			}

			// Try to acquire with a short timeout
			acquireCtx, acquireCancel := context.WithTimeout(ctx, 5*time.Millisecond)
			err := limiter.Acquire(acquireCtx)
			acquireCancel()

			if err == nil {
				counter.Add(1)
			}
		}
	}

	// Launch consumers for both limiters
	var wg sync.WaitGroup
	const consumersPerLimiter = 5
	wg.Add(consumersPerLimiter * 2)

	for i := 0; i < consumersPerLimiter; i++ {
		go consumeTokens(fastLimiter, &fastAcquisitions, &wg)
		go consumeTokens(slowLimiter, &slowAcquisitions, &wg)
	}

	// Wait for the test duration to complete
	wg.Wait()

	// The fast limiter should have allowed more acquisitions than the slow one
	fast := fastAcquisitions.Load()
	slow := slowAcquisitions.Load()

	if fast <= slow {
		t.Errorf("Expected fast limiter (%d) to allow more acquisitions than slow limiter (%d)",
			fast, slow)
	}

	t.Logf("Fast limiter acquisitions: %d, Slow limiter acquisitions: %d", fast, slow)
}

// TestRaceLimiter_StressTest subjects the rate limiter to high concurrency stress
func TestRaceLimiter_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Create a limiter with moderate capacity
	limiter := NewRateLimiter(100, time.Second)

	// Run for a longer duration
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Track operation counts
	var attempts atomic.Int32
	var successes atomic.Int32
	var timeouts atomic.Int32
	var cancellations atomic.Int32

	// Launch many goroutines with different patterns
	var wg sync.WaitGroup
	const goroutineCount = 200

	// Shared cancellation to simulate app shutdown
	stressCtx, stressCancel := context.WithCancel(ctx)

	wg.Add(goroutineCount)
	for i := 0; i < goroutineCount; i++ {
		go func(id int) {
			defer wg.Done()

			// Different goroutines have different patterns
			switch id % 5 {
			case 0:
				// Fast polling with short timeouts
				for {
					select {
					case <-stressCtx.Done():
						return
					default:
						// Continue
					}

					attempts.Add(1)
					acquireCtx, acquireCancel := context.WithTimeout(stressCtx, 5*time.Millisecond)
					err := limiter.Acquire(acquireCtx)
					acquireCancel()

					if err == nil {
						successes.Add(1)
					} else if err == context.DeadlineExceeded {
						timeouts.Add(1)
					} else {
						cancellations.Add(1)
					}

					time.Sleep(time.Millisecond)
				}

			case 1:
				// Slow polling with longer timeouts
				for {
					select {
					case <-stressCtx.Done():
						return
					default:
						// Continue
					}

					attempts.Add(1)
					acquireCtx, acquireCancel := context.WithTimeout(stressCtx, 50*time.Millisecond)
					err := limiter.Acquire(acquireCtx)
					acquireCancel()

					if err == nil {
						successes.Add(1)
					} else if err == context.DeadlineExceeded {
						timeouts.Add(1)
					} else {
						cancellations.Add(1)
					}

					time.Sleep(10 * time.Millisecond)
				}

			case 2:
				// Burst traffic pattern
				for {
					select {
					case <-stressCtx.Done():
						return
					default:
						// Continue
					}

					// Burst of 5 quick requests
					for j := 0; j < 5; j++ {
						attempts.Add(1)
						acquireCtx, acquireCancel := context.WithTimeout(stressCtx, 1*time.Millisecond)
						err := limiter.Acquire(acquireCtx)
						acquireCancel()

						if err == nil {
							successes.Add(1)
						} else if err == context.DeadlineExceeded {
							timeouts.Add(1)
						} else {
							cancellations.Add(1)
						}
					}

					// Rest period
					time.Sleep(50 * time.Millisecond)
				}

			case 3:
				// Long operations
				for {
					select {
					case <-stressCtx.Done():
						return
					default:
						// Continue
					}

					attempts.Add(1)
					acquireCtx, acquireCancel := context.WithTimeout(stressCtx, 100*time.Millisecond)
					err := limiter.Acquire(acquireCtx)
					acquireCancel()

					if err == nil {
						successes.Add(1)
						// Simulate a long operation that holds the token
						time.Sleep(20 * time.Millisecond)
					} else if err == context.DeadlineExceeded {
						timeouts.Add(1)
					} else {
						cancellations.Add(1)
					}
				}

			case 4:
				// Mix of durations
				for {
					select {
					case <-stressCtx.Done():
						return
					default:
						// Continue
					}

					timeout := time.Duration(id%20+1) * time.Millisecond
					attempts.Add(1)
					acquireCtx, acquireCancel := context.WithTimeout(stressCtx, timeout)
					err := limiter.Acquire(acquireCtx)
					acquireCancel()

					if err == nil {
						successes.Add(1)
					} else if err == context.DeadlineExceeded {
						timeouts.Add(1)
					} else {
						cancellations.Add(1)
					}

					time.Sleep(time.Duration(id%5+1) * time.Millisecond)
				}
			}
		}(i)
	}

	// After 1 second, simulate app shutdown by cancelling the context
	time.AfterFunc(1*time.Second, func() {
		stressCancel()
	})

	// Wait for all goroutines to exit
	wg.Wait()

	// Log the results
	t.Logf("Stress test results: %d attempts, %d successes, %d timeouts, %d cancellations",
		attempts.Load(), successes.Load(), timeouts.Load(), cancellations.Load())

	// We don't make specific assertions about the exact numbers, just that the limiter didn't crash
	// and that we got some successes
	if successes.Load() < 100 {
		t.Errorf("Expected at least 100 successful acquisitions, got %d", successes.Load())
	}
}
