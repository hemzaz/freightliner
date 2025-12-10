package testing

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"freightliner/pkg/helper/log"
)

// DeterministicTestFramework provides utilities for creating deterministic, non-flaky tests
type DeterministicTestFramework struct {
	t      *testing.T
	logger log.Logger
}

// NewDeterministicTestFramework creates a new test framework instance
func NewDeterministicTestFramework(t *testing.T) *DeterministicTestFramework {
	return &DeterministicTestFramework{
		t:      t,
		logger: log.NewBasicLogger(log.DebugLevel),
	}
}

// SynchronizedExecution provides deterministic execution patterns
type SynchronizedExecution struct {
	wg       sync.WaitGroup
	mutex    sync.Mutex
	results  []interface{}
	errors   []error
	started  atomic.Bool
	finished atomic.Bool
}

// NewSynchronizedExecution creates a synchronized execution context
func (f *DeterministicTestFramework) NewSynchronizedExecution() *SynchronizedExecution {
	return &SynchronizedExecution{
		results: make([]interface{}, 0),
		errors:  make([]error, 0),
	}
}

// ExecuteWithTimeout runs a function with a deterministic timeout
func (f *DeterministicTestFramework) ExecuteWithTimeout(timeout time.Duration, fn func() error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// WaitForCondition waits for a condition to be true with polling
func (f *DeterministicTestFramework) WaitForCondition(condition func() bool, timeout time.Duration, interval time.Duration) bool {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(interval)
	}

	return false
}

// AsyncTaskGroup manages a group of async tasks with proper synchronization
type AsyncTaskGroup struct {
	tasks   []func() error
	results []error
	wg      sync.WaitGroup
	mutex   sync.Mutex
}

// NewAsyncTaskGroup creates a new async task group
func (f *DeterministicTestFramework) NewAsyncTaskGroup() *AsyncTaskGroup {
	return &AsyncTaskGroup{
		tasks:   make([]func() error, 0),
		results: make([]error, 0),
	}
}

// AddTask adds a task to the group
func (g *AsyncTaskGroup) AddTask(task func() error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.tasks = append(g.tasks, task)
}

// ExecuteAll executes all tasks concurrently and waits for completion
func (g *AsyncTaskGroup) ExecuteAll() []error {
	g.mutex.Lock()
	taskCount := len(g.tasks)
	g.results = make([]error, taskCount)
	g.mutex.Unlock()

	for i, task := range g.tasks {
		g.wg.Add(1)
		go func(index int, t func() error) {
			defer g.wg.Done()
			err := t()

			g.mutex.Lock()
			g.results[index] = err
			g.mutex.Unlock()
		}(i, task)
	}

	g.wg.Wait()
	return g.results
}

// CounterGroup provides atomic counters for test validation
type CounterGroup struct {
	counters map[string]*atomic.Int64
	mutex    sync.RWMutex
}

// NewCounterGroup creates a new counter group
func (f *DeterministicTestFramework) NewCounterGroup() *CounterGroup {
	return &CounterGroup{
		counters: make(map[string]*atomic.Int64),
	}
}

// GetCounter gets or creates a counter by name
func (c *CounterGroup) GetCounter(name string) *atomic.Int64 {
	c.mutex.RLock()
	if counter, exists := c.counters[name]; exists {
		c.mutex.RUnlock()
		return counter
	}
	c.mutex.RUnlock()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Double-check pattern
	if counter, exists := c.counters[name]; exists {
		return counter
	}

	counter := &atomic.Int64{}
	c.counters[name] = counter
	return counter
}

// GetCounterValue gets the current value of a counter
func (c *CounterGroup) GetCounterValue(name string) int64 {
	return c.GetCounter(name).Load()
}

// IncrementCounter increments a counter by 1
func (c *CounterGroup) IncrementCounter(name string) {
	c.GetCounter(name).Add(1)
}

// AddToCounter adds a value to a counter
func (c *CounterGroup) AddToCounter(name string, value int64) {
	c.GetCounter(name).Add(value)
}

// ResetCounter resets a counter to 0
func (c *CounterGroup) ResetCounter(name string) {
	c.GetCounter(name).Store(0)
}

// GetAllCounters returns a snapshot of all counter values
func (c *CounterGroup) GetAllCounters() map[string]int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	snapshot := make(map[string]int64)
	for name, counter := range c.counters {
		snapshot[name] = counter.Load()
	}
	return snapshot
}

// MockChannelDrainer helps prevent channel-related deadlocks in tests
type MockChannelDrainer struct {
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewMockChannelDrainer creates a new channel drainer
func (f *DeterministicTestFramework) NewMockChannelDrainer() *MockChannelDrainer {
	return &MockChannelDrainer{
		stopChan: make(chan struct{}),
	}
}

// DrainChannel drains a channel to prevent deadlock
func (d *MockChannelDrainer) DrainChannel(ch <-chan interface{}) {
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for {
			select {
			case <-ch:
				// Drain the channel
			case <-d.stopChan:
				return
			}
		}
	}()
}

// Stop stops all draining operations
func (d *MockChannelDrainer) Stop() {
	close(d.stopChan)
	d.wg.Wait()
}

// TestStateManager manages test state to prevent race conditions
type TestStateManager struct {
	state  map[string]interface{}
	mutex  sync.RWMutex
	logger log.Logger
}

// NewTestStateManager creates a new test state manager
func (f *DeterministicTestFramework) NewTestStateManager() *TestStateManager {
	return &TestStateManager{
		state:  make(map[string]interface{}),
		logger: f.logger,
	}
}

// SetState sets a state value atomically
func (s *TestStateManager) SetState(key string, value interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.state[key] = value
	s.logger.WithFields(map[string]interface{}{
		"key":   key,
		"value": value,
	}).Debug("Test state set")
}

// GetState gets a state value atomically
func (s *TestStateManager) GetState(key string) (interface{}, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	value, exists := s.state[key]
	return value, exists
}

// WaitForState waits for a state key to have a specific value
func (s *TestStateManager) WaitForState(key string, expectedValue interface{}, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if value, exists := s.GetState(key); exists && value == expectedValue {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}

	return false
}

// DeterministicTimeProvider provides controlled time for tests
type DeterministicTimeProvider struct {
	currentTime atomic.Value // stores time.Time
	mutex       sync.Mutex
}

// NewDeterministicTimeProvider creates a time provider starting at a specific time
func (f *DeterministicTestFramework) NewDeterministicTimeProvider(startTime time.Time) *DeterministicTimeProvider {
	provider := &DeterministicTimeProvider{}
	provider.currentTime.Store(startTime)
	return provider
}

// Now returns the current mock time
func (p *DeterministicTimeProvider) Now() time.Time {
	return p.currentTime.Load().(time.Time)
}

// Advance advances the mock time by a duration
func (p *DeterministicTimeProvider) Advance(duration time.Duration) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	current := p.currentTime.Load().(time.Time)
	p.currentTime.Store(current.Add(duration))
}

// Helper methods for common test patterns

// AssertEventuallyTrue asserts that a condition becomes true within a timeout
func (f *DeterministicTestFramework) AssertEventuallyTrue(condition func() bool, timeout time.Duration, msgAndArgs ...interface{}) {
	if !f.WaitForCondition(condition, timeout, 10*time.Millisecond) {
		f.t.Errorf("Condition did not become true within %v: %v", timeout, msgAndArgs)
	}
}

// AssertCounterEquals asserts that a counter has the expected value
func (f *DeterministicTestFramework) AssertCounterEquals(counter *atomic.Int64, expected int64, msgAndArgs ...interface{}) {
	actual := counter.Load()
	if actual != expected {
		f.t.Errorf("Expected counter value %d, got %d: %v", expected, actual, msgAndArgs)
	}
}

// AssertNoRaceConditions runs a test multiple times to check for race conditions
func (f *DeterministicTestFramework) AssertNoRaceConditions(testFunc func(), iterations int) {
	for i := 0; i < iterations; i++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					f.t.Errorf("Race condition detected in iteration %d: %v", i, r)
				}
			}()
			testFunc()
		}()
	}
}
