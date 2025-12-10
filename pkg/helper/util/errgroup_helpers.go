package util

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// LimitedErrGroup wraps errgroup with a semaphore to limit concurrency
type LimitedErrGroup struct {
	group *errgroup.Group
	ctx   context.Context
	sem   *semaphore.Weighted
}

// NewLimitedErrGroup creates a new error group with limited concurrency
func NewLimitedErrGroup(ctx context.Context, maxConcurrency int) *LimitedErrGroup {
	g, ctx := errgroup.WithContext(ctx)

	// If maxConcurrency is <= 0, use unlimited concurrency (no semaphore)
	var sem *semaphore.Weighted
	if maxConcurrency > 0 {
		sem = semaphore.NewWeighted(int64(maxConcurrency))
	}

	return &LimitedErrGroup{
		group: g,
		ctx:   ctx,
		sem:   sem,
	}
}

// Go runs the given function in a new goroutine, respecting concurrency limits
func (g *LimitedErrGroup) Go(f func() error) {
	g.group.Go(func() error {
		// If no semaphore was created (unlimited concurrency), just run the function
		if g.sem == nil {
			return f()
		}

		// Acquire semaphore (blocks if max concurrency reached)
		if err := g.sem.Acquire(g.ctx, 1); err != nil {
			return err
		}

		// Release semaphore when done
		defer g.sem.Release(1)

		// Run the function
		return f()
	})
}

// Wait waits for all goroutines to complete and returns the first error
func (g *LimitedErrGroup) Wait() error {
	return g.group.Wait()
}

// Results collects results from multiple goroutines while respecting errgroup error handling
type Results struct {
	mu      sync.Mutex
	items   []interface{}
	metrics map[string]int64
}

// NewResults creates a new Results collector
func NewResults() *Results {
	return &Results{
		items:   make([]interface{}, 0),
		metrics: make(map[string]int64),
	}
}

// Add adds an item to the results
func (r *Results) Add(item interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, item)
}

// AddMetric adds to a numeric metric
func (r *Results) AddMetric(name string, value int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics[name] += value
}

// GetItems returns all collected items
func (r *Results) GetItems() []interface{} {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Return a copy to avoid concurrent modification issues
	result := make([]interface{}, len(r.items))
	copy(result, r.items)
	return result
}

// GetMetric gets the value of a metric
func (r *Results) GetMetric(name string) int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.metrics[name]
}

// GetAllMetrics returns all metrics
func (r *Results) GetAllMetrics() map[string]int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Return a copy to avoid concurrent modification issues
	result := make(map[string]int64, len(r.metrics))
	for k, v := range r.metrics {
		result[k] = v
	}
	return result
}
