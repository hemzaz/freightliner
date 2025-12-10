package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"

	"freightliner/pkg/helper/log"
)

// BulkheadSettings configures bulkhead behavior
type BulkheadSettings struct {
	// MaxConcurrent is the maximum number of concurrent operations
	MaxConcurrent int64
	// MaxQueueDepth is the maximum number of queued operations
	MaxQueueDepth int
	// Timeout for acquiring semaphore
	Timeout time.Duration
}

// DefaultBulkheadSettings returns sensible defaults
func DefaultBulkheadSettings() BulkheadSettings {
	return BulkheadSettings{
		MaxConcurrent: 100,
		MaxQueueDepth: 500,
		Timeout:       30 * time.Second,
	}
}

// Bulkhead implements the bulkhead pattern for resource isolation
type Bulkhead struct {
	name      string
	settings  BulkheadSettings
	semaphore *semaphore.Weighted
	queue     chan struct{}
	mu        sync.RWMutex
	logger    log.Logger
	stats     *bulkheadStats
}

type bulkheadStats struct {
	totalExecutions int64
	totalRejections int64
	totalTimeouts   int64
	currentActive   int64
	currentQueued   int64
	mu              sync.RWMutex
}

// NewBulkhead creates a new bulkhead
func NewBulkhead(name string, settings BulkheadSettings, logger log.Logger) *Bulkhead {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	return &Bulkhead{
		name:      name,
		settings:  settings,
		semaphore: semaphore.NewWeighted(settings.MaxConcurrent),
		queue:     make(chan struct{}, settings.MaxQueueDepth),
		logger:    logger,
		stats:     &bulkheadStats{},
	}
}

// Execute runs a function with bulkhead protection
func (b *Bulkhead) Execute(ctx context.Context, fn func() error) error {
	// Try to add to queue
	select {
	case b.queue <- struct{}{}:
		defer func() { <-b.queue }()
		b.stats.incrementQueued()
		defer b.stats.decrementQueued()
	default:
		// Queue is full
		b.stats.incrementRejections()
		b.logger.WithFields(map[string]interface{}{
			"bulkhead": b.name,
		}).Warn("Bulkhead queue full, rejecting request")
		return fmt.Errorf("bulkhead '%s' queue full", b.name)
	}

	// Create context with timeout
	acquireCtx := ctx
	if b.settings.Timeout > 0 {
		var cancel context.CancelFunc
		acquireCtx, cancel = context.WithTimeout(ctx, b.settings.Timeout)
		defer cancel()
	}

	// Try to acquire semaphore
	if err := b.semaphore.Acquire(acquireCtx, 1); err != nil {
		b.stats.incrementTimeouts()
		b.logger.WithFields(map[string]interface{}{
			"bulkhead": b.name,
		}).Warn("Bulkhead semaphore acquisition timeout")
		return fmt.Errorf("bulkhead '%s' timeout: %w", b.name, err)
	}
	defer b.semaphore.Release(1)

	// Execute the function
	b.stats.incrementActive()
	b.stats.incrementExecutions()
	defer b.stats.decrementActive()

	return fn()
}

// Stats returns current bulkhead statistics
func (b *Bulkhead) Stats() BulkheadStats {
	return BulkheadStats{
		Name:            b.name,
		MaxConcurrent:   b.settings.MaxConcurrent,
		MaxQueueDepth:   b.settings.MaxQueueDepth,
		ActiveCount:     b.stats.getActive(),
		QueuedCount:     b.stats.getQueued(),
		TotalExecutions: b.stats.getExecutions(),
		TotalRejections: b.stats.getRejections(),
		TotalTimeouts:   b.stats.getTimeouts(),
	}
}

// BulkheadStats represents bulkhead statistics
type BulkheadStats struct {
	Name            string
	MaxConcurrent   int64
	MaxQueueDepth   int
	ActiveCount     int64
	QueuedCount     int64
	TotalExecutions int64
	TotalRejections int64
	TotalTimeouts   int64
}

// Helper methods for stats
func (s *bulkheadStats) incrementActive() {
	s.mu.Lock()
	s.currentActive++
	s.mu.Unlock()
}

func (s *bulkheadStats) decrementActive() {
	s.mu.Lock()
	s.currentActive--
	s.mu.Unlock()
}

func (s *bulkheadStats) incrementQueued() {
	s.mu.Lock()
	s.currentQueued++
	s.mu.Unlock()
}

func (s *bulkheadStats) decrementQueued() {
	s.mu.Lock()
	s.currentQueued--
	s.mu.Unlock()
}

func (s *bulkheadStats) incrementExecutions() {
	s.mu.Lock()
	s.totalExecutions++
	s.mu.Unlock()
}

func (s *bulkheadStats) incrementRejections() {
	s.mu.Lock()
	s.totalRejections++
	s.mu.Unlock()
}

func (s *bulkheadStats) incrementTimeouts() {
	s.mu.Lock()
	s.totalTimeouts++
	s.mu.Unlock()
}

func (s *bulkheadStats) getActive() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentActive
}

func (s *bulkheadStats) getQueued() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentQueued
}

func (s *bulkheadStats) getExecutions() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.totalExecutions
}

func (s *bulkheadStats) getRejections() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.totalRejections
}

func (s *bulkheadStats) getTimeouts() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.totalTimeouts
}

// BulkheadManager manages multiple bulkheads
type BulkheadManager struct {
	bulkheads map[string]*Bulkhead
	mu        sync.RWMutex
	logger    log.Logger
}

// NewBulkheadManager creates a new bulkhead manager
func NewBulkheadManager(logger log.Logger) *BulkheadManager {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	return &BulkheadManager{
		bulkheads: make(map[string]*Bulkhead),
		logger:    logger,
	}
}

// GetOrCreate gets an existing bulkhead or creates a new one
func (m *BulkheadManager) GetOrCreate(name string, settings BulkheadSettings) *Bulkhead {
	m.mu.RLock()
	bulkhead, exists := m.bulkheads[name]
	m.mu.RUnlock()

	if exists {
		return bulkhead
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if bulkhead, exists := m.bulkheads[name]; exists {
		return bulkhead
	}

	// Create new bulkhead
	bulkhead = NewBulkhead(name, settings, m.logger)
	m.bulkheads[name] = bulkhead

	return bulkhead
}

// Get retrieves a bulkhead by name
func (m *BulkheadManager) Get(name string) (*Bulkhead, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	bulkhead, exists := m.bulkheads[name]
	return bulkhead, exists
}

// Execute runs a function with bulkhead protection
func (m *BulkheadManager) Execute(ctx context.Context, name string, fn func() error) error {
	settings := DefaultBulkheadSettings()
	bulkhead := m.GetOrCreate(name, settings)
	return bulkhead.Execute(ctx, fn)
}

// GetAllStats returns statistics for all bulkheads
func (m *BulkheadManager) GetAllStats() []BulkheadStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make([]BulkheadStats, 0, len(m.bulkheads))
	for _, bulkhead := range m.bulkheads {
		stats = append(stats, bulkhead.Stats())
	}
	return stats
}
