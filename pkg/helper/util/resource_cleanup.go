package util

import (
	"context"
	"io"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
)

// ResourceCleaner provides centralized resource cleanup with proper error handling
type ResourceCleaner struct {
	resources []CleanupResource
	mutex     sync.Mutex
	logger    log.Logger
	cleaned   atomic.Bool
}

// CleanupResource represents a resource that needs cleanup
type CleanupResource struct {
	Name     string
	Cleanup  func() error
	Priority int // Higher priority resources are cleaned first
}

// NewResourceCleaner creates a new resource cleaner
func NewResourceCleaner(logger log.Logger) *ResourceCleaner {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}
	return &ResourceCleaner{
		resources: make([]CleanupResource, 0),
		logger:    logger,
	}
}

// AddResource adds a resource for cleanup
func (rc *ResourceCleaner) AddResource(name string, cleanup func() error, priority int) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	if rc.cleaned.Load() {
		rc.logger.WithField("resource", name).Warn("Attempted to add resource to already cleaned cleaner")
		return
	}

	rc.resources = append(rc.resources, CleanupResource{
		Name:     name,
		Cleanup:  cleanup,
		Priority: priority,
	})
}

// AddCloser adds an io.Closer for cleanup
func (rc *ResourceCleaner) AddCloser(name string, closer io.Closer, priority int) {
	if closer == nil {
		return
	}
	rc.AddResource(name, func() error {
		return closer.Close()
	}, priority)
}

// AddCancelFunc adds a context cancel function for cleanup
func (rc *ResourceCleaner) AddCancelFunc(name string, cancel context.CancelFunc, priority int) {
	if cancel == nil {
		return
	}
	rc.AddResource(name, func() error {
		cancel()
		return nil
	}, priority)
}

// CleanupAll performs cleanup of all resources in priority order
func (rc *ResourceCleaner) CleanupAll() error {
	if !rc.cleaned.CompareAndSwap(false, true) {
		return nil // Already cleaned
	}

	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	// Sort resources by priority using efficient O(n log n) algorithm instead of O(n²) bubble sort
	resources := make([]CleanupResource, len(rc.resources))
	copy(resources, rc.resources)

	// Use Go's optimized sorting (introsort/quicksort hybrid) - O(n log n)
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Priority > resources[j].Priority // Higher priority first
	})

	var cleanupErrors []error

	for _, resource := range resources {
		if resource.Cleanup != nil {
			if err := resource.Cleanup(); err != nil {
				rc.logger.WithFields(map[string]interface{}{
					"resource": resource.Name,
					"priority": resource.Priority,
				}).WithError(err).Error("Resource cleanup failed", err)
				cleanupErrors = append(cleanupErrors, errors.Wrapf(err, "cleanup failed for %s", resource.Name))
			} else {
				rc.logger.WithFields(map[string]interface{}{
					"resource": resource.Name,
					"priority": resource.Priority,
				}).Debug("Resource cleaned successfully")
			}
		}
	}

	if len(cleanupErrors) > 0 {
		return errors.Multiple(cleanupErrors...)
	}

	return nil
}

// DeferCleanupAll sets up cleanup to run when the function returns (use with defer)
func (rc *ResourceCleaner) DeferCleanupAll() {
	if err := rc.CleanupAll(); err != nil {
		rc.logger.WithError(err).Error("Deferred cleanup failed", err)
	}
}

// OptimizedResourceManager provides advanced resource management with automatic cleanup
type OptimizedResourceManager struct {
	cleaner   *ResourceCleaner
	bufferMgr *BufferManager
	logger    log.Logger

	// Resource tracking for leak detection
	activeReaders sync.Map
	activeWriters sync.Map
	activeBuffers sync.Map

	// Statistics
	stats ResourceStats
}

// ResourceStats tracks resource usage statistics
type ResourceStats struct {
	ReadersCreated   atomic.Int64
	WritersCreated   atomic.Int64
	BuffersAllocated atomic.Int64
	ResourcesCleaned atomic.Int64
	CleanupFailures  atomic.Int64
}

// NewOptimizedResourceManager creates a new optimized resource manager
func NewOptimizedResourceManager(logger log.Logger) *OptimizedResourceManager {
	return &OptimizedResourceManager{
		cleaner:   NewResourceCleaner(logger),
		bufferMgr: NewBufferManager(),
		logger:    logger,
	}
}

// ManagedReader wraps an io.Reader with automatic cleanup tracking
type ManagedReader struct {
	io.Reader
	name    string
	manager *OptimizedResourceManager
	cleaned atomic.Bool
}

// Close implements io.Closer for automatic cleanup
func (mr *ManagedReader) Close() error {
	if !mr.cleaned.CompareAndSwap(false, true) {
		return nil // Already closed
	}

	mr.manager.activeReaders.Delete(mr.name)
	mr.manager.stats.ResourcesCleaned.Add(1)

	if closer, ok := mr.Reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// CreateManagedReader creates a reader with automatic cleanup tracking
func (orm *OptimizedResourceManager) CreateManagedReader(name string, reader io.Reader) *ManagedReader {
	managedReader := &ManagedReader{
		Reader:  reader,
		name:    name,
		manager: orm,
	}

	orm.activeReaders.Store(name, managedReader)
	orm.stats.ReadersCreated.Add(1)

	// Add to cleanup list
	orm.cleaner.AddResource(name+"_reader", func() error {
		return managedReader.Close()
	}, 5) // Medium priority

	return managedReader
}

// ManagedWriter wraps an io.Writer with automatic cleanup tracking
type ManagedWriter struct {
	io.Writer
	name    string
	manager *OptimizedResourceManager
	cleaned atomic.Bool
}

// Close implements io.Closer for automatic cleanup
func (mw *ManagedWriter) Close() error {
	if !mw.cleaned.CompareAndSwap(false, true) {
		return nil // Already closed
	}

	mw.manager.activeWriters.Delete(mw.name)
	mw.manager.stats.ResourcesCleaned.Add(1)

	if closer, ok := mw.Writer.(io.Closer); ok {
		if flusher, ok := mw.Writer.(interface{ Flush() error }); ok {
			_ = flusher.Flush() // Best effort flush before close
		}
		return closer.Close()
	}
	return nil
}

// CreateManagedWriter creates a writer with automatic cleanup tracking
func (orm *OptimizedResourceManager) CreateManagedWriter(name string, writer io.Writer) *ManagedWriter {
	managedWriter := &ManagedWriter{
		Writer:  writer,
		name:    name,
		manager: orm,
	}

	orm.activeWriters.Store(name, managedWriter)
	orm.stats.WritersCreated.Add(1)

	// Add to cleanup list
	orm.cleaner.AddResource(name+"_writer", func() error {
		return managedWriter.Close()
	}, 5) // Medium priority

	return managedWriter
}

// ManagedBuffer provides automatic buffer cleanup
type ManagedBuffer struct {
	*ReusableBuffer
	name    string
	manager *OptimizedResourceManager
	cleaned atomic.Bool
}

// Release implements proper buffer cleanup with tracking
func (mb *ManagedBuffer) Release() {
	if !mb.cleaned.CompareAndSwap(false, true) {
		return // Already released
	}

	mb.manager.activeBuffers.Delete(mb.name)
	mb.manager.stats.ResourcesCleaned.Add(1)
	mb.ReusableBuffer.Release()
}

// CreateManagedBuffer creates a buffer with automatic cleanup tracking
func (orm *OptimizedResourceManager) CreateManagedBuffer(name string, size int64, operation string) *ManagedBuffer {
	reusableBuffer := orm.bufferMgr.GetOptimalBuffer(size, operation)

	managedBuffer := &ManagedBuffer{
		ReusableBuffer: reusableBuffer,
		name:           name,
		manager:        orm,
	}

	orm.activeBuffers.Store(name, managedBuffer)
	orm.stats.BuffersAllocated.Add(1)

	// Add to cleanup list with high priority (buffers should be cleaned first)
	orm.cleaner.AddResource(name+"_buffer", func() error {
		managedBuffer.Release()
		return nil
	}, 10) // High priority

	return managedBuffer
}

// PerformWithTimeout executes a function with resource cleanup and timeout
func (orm *OptimizedResourceManager) PerformWithTimeout(
	ctx context.Context,
	timeout time.Duration,
	operation func(context.Context, *OptimizedResourceManager) error,
) error {
	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Add cancel function to cleanup
	orm.cleaner.AddCancelFunc("timeout_context", cancel, 8) // High priority

	// Ensure cleanup happens regardless of how function exits
	defer orm.cleaner.DeferCleanupAll()

	// Execute operation
	return operation(timeoutCtx, orm)
}

// GetStats returns current resource usage statistics
func (orm *OptimizedResourceManager) GetStats() *ResourceStats {
	return &orm.stats
}

// DetectLeaks logs any resources that haven't been properly cleaned up
func (orm *OptimizedResourceManager) DetectLeaks() {
	leaks := 0

	orm.activeReaders.Range(func(key, value interface{}) bool {
		orm.logger.WithField("reader", key).Warn("Potential reader leak detected")
		leaks++
		return true
	})

	orm.activeWriters.Range(func(key, value interface{}) bool {
		orm.logger.WithField("writer", key).Warn("Potential writer leak detected")
		leaks++
		return true
	})

	orm.activeBuffers.Range(func(key, value interface{}) bool {
		orm.logger.WithField("buffer", key).Warn("Potential buffer leak detected")
		leaks++
		return true
	})

	if leaks > 0 {
		orm.logger.WithField("total_leaks", leaks).Error("Resource leaks detected", errors.New("resource leak detection"))
	}
}

// SafeCleanupFunc provides a helper for creating safe cleanup functions
func SafeCleanupFunc(name string, cleanup func() error, logger log.Logger) func() error {
	return func() error {
		defer func() {
			if r := recover(); r != nil {
				if logger != nil {
					logger.WithField("resource", name).Error("Panic during cleanup", errors.Newf("panic: %v", r))
				}
			}
		}()

		if cleanup != nil {
			return cleanup()
		}
		return nil
	}
}

// DeferSafeCleanup sets up safe cleanup with panic recovery
func DeferSafeCleanup(name string, cleanup func() error, logger log.Logger) {
	defer func() {
		if r := recover(); r != nil {
			if logger != nil {
				logger.WithField("resource", name).Error("Panic during deferred cleanup", errors.Newf("panic: %v", r))
			}
		}

		if cleanup != nil {
			if err := cleanup(); err != nil && logger != nil {
				logger.WithField("resource", name).WithError(err).Error("Deferred cleanup failed", err)
			}
		}
	}()
}

// ResourceMonitor provides monitoring and alerts for resource usage
type ResourceMonitor struct {
	manager  *OptimizedResourceManager
	logger   log.Logger
	ticker   *time.Ticker
	stopChan chan struct{}
	started  atomic.Bool
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor(manager *OptimizedResourceManager, checkInterval time.Duration) *ResourceMonitor {
	return &ResourceMonitor{
		manager:  manager,
		logger:   manager.logger,
		ticker:   time.NewTicker(checkInterval),
		stopChan: make(chan struct{}),
	}
}

// Start starts the resource monitor
func (rm *ResourceMonitor) Start() {
	if !rm.started.CompareAndSwap(false, true) {
		return // Already started
	}

	go func() {
		defer rm.ticker.Stop()

		for {
			select {
			case <-rm.ticker.C:
				rm.checkResources()
			case <-rm.stopChan:
				return
			}
		}
	}()
}

// Stop stops the resource monitor
func (rm *ResourceMonitor) Stop() {
	if rm.started.CompareAndSwap(true, false) {
		close(rm.stopChan)
	}
}

// checkResources performs periodic resource checks
func (rm *ResourceMonitor) checkResources() {
	stats := rm.manager.GetStats()

	rm.logger.WithFields(map[string]interface{}{
		"readers_created":   stats.ReadersCreated.Load(),
		"writers_created":   stats.WritersCreated.Load(),
		"buffers_allocated": stats.BuffersAllocated.Load(),
		"resources_cleaned": stats.ResourcesCleaned.Load(),
		"cleanup_failures":  stats.CleanupFailures.Load(),
	}).Debug("Resource usage stats")

	// Check for potential leaks
	rm.manager.DetectLeaks()
}
