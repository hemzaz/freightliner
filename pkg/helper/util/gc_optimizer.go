package util

import (
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"time"
	"unsafe"

	"freightliner/pkg/helper/log"
)

// GCOptimizer provides advanced garbage collection optimization strategies
type GCOptimizer struct {
	logger            log.Logger
	enabled           atomic.Bool
	originalGCPercent int

	// Object lifecycle management
	shortLivedObjects  *ObjectPool
	mediumLivedObjects *ObjectPool
	longLivedObjects   *ObjectPool

	// Memory pressure monitoring
	memoryPressure atomic.Int64
	lastGCTime     atomic.Int64
	gcStats        GCStats

	// Configuration
	config GCOptimizerConfig
}

// GCOptimizerConfig holds configuration for GC optimization
type GCOptimizerConfig struct {
	// GC tuning parameters
	LowMemoryGCPercent  int   // GC percent under low memory pressure
	HighMemoryGCPercent int   // GC percent under high memory pressure
	MemoryPressureLimit int64 // Memory threshold for pressure detection (bytes)

	// Object lifecycle thresholds
	ShortLifeThreshold  time.Duration // Objects expected to live less than this
	MediumLifeThreshold time.Duration // Objects expected to live less than this

	// Monitoring intervals
	MonitoringInterval time.Duration // How often to check memory pressure
	GCStatsInterval    time.Duration // How often to collect GC stats

	// Pool sizes
	ShortLivedPoolSize  int
	MediumLivedPoolSize int
	LongLivedPoolSize   int
}

// GCStats tracks garbage collection statistics
type GCStats struct {
	NumGC      atomic.Uint32
	PauseTotal atomic.Int64 // Total pause time in nanoseconds
	LastPause  atomic.Int64 // Last pause time in nanoseconds
	HeapSize   atomic.Int64 // Current heap size
	HeapInUse  atomic.Int64 // Current heap in use
	StackInUse atomic.Int64 // Current stack in use
	NextGC     atomic.Int64 // Next GC threshold
}

// DefaultGCOptimizerConfig returns sensible defaults for GC optimization
func DefaultGCOptimizerConfig() GCOptimizerConfig {
	return GCOptimizerConfig{
		LowMemoryGCPercent:  200,               // More relaxed GC under low pressure
		HighMemoryGCPercent: 50,                // Aggressive GC under high pressure
		MemoryPressureLimit: 500 * 1024 * 1024, // 500MB threshold
		ShortLifeThreshold:  100 * time.Millisecond,
		MediumLifeThreshold: 10 * time.Second,
		MonitoringInterval:  5 * time.Second,
		GCStatsInterval:     10 * time.Second,
		ShortLivedPoolSize:  1000,
		MediumLivedPoolSize: 100,
		LongLivedPoolSize:   10,
	}
}

// NewGCOptimizer creates a new GC optimizer
func NewGCOptimizer(config GCOptimizerConfig, logger log.Logger) *GCOptimizer {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	optimizer := &GCOptimizer{
		logger:            logger,
		originalGCPercent: debug.SetGCPercent(-1), // Get current value
		config:            config,
	}

	// Restore original GC percent
	debug.SetGCPercent(optimizer.originalGCPercent)

	// Initialize object pools
	optimizer.shortLivedObjects = NewObjectPool(func() interface{} {
		return &ObjectLifecycleWrapper{
			createdAt: time.Now(),
			category:  ShortLived,
		}
	})

	optimizer.mediumLivedObjects = NewObjectPool(func() interface{} {
		return &ObjectLifecycleWrapper{
			createdAt: time.Now(),
			category:  MediumLived,
		}
	})

	optimizer.longLivedObjects = NewObjectPool(func() interface{} {
		return &ObjectLifecycleWrapper{
			createdAt: time.Now(),
			category:  LongLived,
		}
	})

	return optimizer
}

// ObjectLifecycleCategory defines different object lifecycle categories
type ObjectLifecycleCategory int

const (
	ShortLived ObjectLifecycleCategory = iota
	MediumLived
	LongLived
)

// ObjectLifecycleWrapper wraps objects with lifecycle information
type ObjectLifecycleWrapper struct {
	object    interface{}
	createdAt time.Time
	category  ObjectLifecycleCategory
	inUse     atomic.Bool
}

// Get returns the wrapped object and marks it as in use
func (olw *ObjectLifecycleWrapper) Get() interface{} {
	olw.inUse.Store(true)
	return olw.object
}

// Release marks the object as no longer in use and eligible for pooling
func (olw *ObjectLifecycleWrapper) Release() {
	olw.inUse.Store(false)
}

// IsInUse returns whether the object is currently in use
func (olw *ObjectLifecycleWrapper) IsInUse() bool {
	return olw.inUse.Load()
}

// Age returns how long the object has been alive
func (olw *ObjectLifecycleWrapper) Age() time.Duration {
	return time.Since(olw.createdAt)
}

// Start enables GC optimization and starts monitoring
func (gco *GCOptimizer) Start() {
	if !gco.enabled.CompareAndSwap(false, true) {
		return // Already started
	}

	gco.logger.WithFields(map[string]interface{}{
		"original_gc_percent":   gco.originalGCPercent,
		"low_memory_gc":         gco.config.LowMemoryGCPercent,
		"high_memory_gc":        gco.config.HighMemoryGCPercent,
		"memory_pressure_limit": gco.config.MemoryPressureLimit,
	}).Info("Starting GC optimizer")

	// Start monitoring goroutines
	go gco.memoryPressureMonitor()
	go gco.gcStatsCollector()

	// Set initial GC parameters
	gco.adjustGCParameters()
}

// Stop disables GC optimization and restores defaults
func (gco *GCOptimizer) Stop() {
	if !gco.enabled.CompareAndSwap(true, false) {
		return // Already stopped
	}

	// Restore original GC percent
	debug.SetGCPercent(gco.originalGCPercent)

	gco.logger.WithField("restored_gc_percent", gco.originalGCPercent).Info("GC optimizer stopped")
}

// memoryPressureMonitor continuously monitors memory pressure
func (gco *GCOptimizer) memoryPressureMonitor() {
	ticker := time.NewTicker(gco.config.MonitoringInterval)
	defer ticker.Stop()

	for gco.enabled.Load() {
		<-ticker.C
		gco.updateMemoryPressure()
		gco.adjustGCParameters()
	}
}

// updateMemoryPressure calculates current memory pressure
func (gco *GCOptimizer) updateMemoryPressure() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Calculate memory pressure based on heap size and allocation rate
	heapInUse := int64(m.HeapInuse)
	pressure := heapInUse * 100 / gco.config.MemoryPressureLimit

	gco.memoryPressure.Store(pressure)
	gco.gcStats.HeapSize.Store(int64(m.HeapSys))
	gco.gcStats.HeapInUse.Store(heapInUse)
	gco.gcStats.StackInUse.Store(int64(m.StackInuse))
	gco.gcStats.NextGC.Store(int64(m.NextGC))

	gco.logger.WithFields(map[string]interface{}{
		"heap_in_use_mb":   heapInUse / (1024 * 1024),
		"heap_sys_mb":      m.HeapSys / (1024 * 1024),
		"stack_in_use_mb":  m.StackInuse / (1024 * 1024),
		"next_gc_mb":       m.NextGC / (1024 * 1024),
		"pressure_percent": pressure,
		"num_gc":           m.NumGC,
	}).Debug("Memory pressure updated")
}

// adjustGCParameters adjusts GC parameters based on memory pressure
func (gco *GCOptimizer) adjustGCParameters() {
	pressure := gco.memoryPressure.Load()

	var newGCPercent int
	if pressure > 80 { // High memory pressure
		newGCPercent = gco.config.HighMemoryGCPercent
	} else if pressure < 20 { // Low memory pressure
		newGCPercent = gco.config.LowMemoryGCPercent
	} else { // Medium pressure - interpolate
		ratio := float64(pressure-20) / 60.0 // Scale 20-80 to 0-1
		newGCPercent = int(float64(gco.config.LowMemoryGCPercent) +
			ratio*float64(gco.config.HighMemoryGCPercent-gco.config.LowMemoryGCPercent))
	}

	// Set the new GC percent
	oldPercent := debug.SetGCPercent(newGCPercent)

	if oldPercent != newGCPercent {
		gco.logger.WithFields(map[string]interface{}{
			"old_gc_percent":  oldPercent,
			"new_gc_percent":  newGCPercent,
			"memory_pressure": pressure,
		}).Debug("Adjusted GC parameters")
	}
}

// gcStatsCollector collects detailed GC statistics
func (gco *GCOptimizer) gcStatsCollector() {
	ticker := time.NewTicker(gco.config.GCStatsInterval)
	defer ticker.Stop()

	var lastNumGC uint32
	var lastPauseTotal time.Duration

	for gco.enabled.Load() {
		<-ticker.C
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// Update GC statistics
		gco.gcStats.NumGC.Store(m.NumGC)
		gco.gcStats.PauseTotal.Store(int64(m.PauseTotalNs))

		if m.NumGC > 0 {
			gco.gcStats.LastPause.Store(int64(m.PauseNs[(m.NumGC+255)%256]))
		}

		// Log GC activity if there were new collections
		if m.NumGC > lastNumGC {
			newCollections := m.NumGC - lastNumGC
			newPauseTime := time.Duration(m.PauseTotalNs) - lastPauseTotal
			avgPause := newPauseTime / time.Duration(newCollections)

			gco.logger.WithFields(map[string]interface{}{
				"new_collections":   newCollections,
				"total_collections": m.NumGC,
				"avg_pause_ms":      avgPause.Nanoseconds() / 1000000,
				"total_pause_ms":    newPauseTime.Nanoseconds() / 1000000,
			}).Info("GC activity")

			lastNumGC = m.NumGC
			lastPauseTotal = time.Duration(m.PauseTotalNs)
		}
	}
}

// GetObject returns an object from the appropriate lifecycle pool
func (gco *GCOptimizer) GetObject(category ObjectLifecycleCategory) *ObjectLifecycleWrapper {
	switch category {
	case ShortLived:
		return gco.shortLivedObjects.Get().(*ObjectLifecycleWrapper)
	case MediumLived:
		return gco.mediumLivedObjects.Get().(*ObjectLifecycleWrapper)
	case LongLived:
		return gco.longLivedObjects.Get().(*ObjectLifecycleWrapper)
	default:
		return gco.shortLivedObjects.Get().(*ObjectLifecycleWrapper)
	}
}

// ReturnObject returns an object to the appropriate lifecycle pool
func (gco *GCOptimizer) ReturnObject(wrapper *ObjectLifecycleWrapper) {
	wrapper.Release()

	switch wrapper.category {
	case ShortLived:
		gco.shortLivedObjects.Put(wrapper)
	case MediumLived:
		gco.mediumLivedObjects.Put(wrapper)
	case LongLived:
		gco.longLivedObjects.Put(wrapper)
	}
}

// ForceGC triggers garbage collection with logging
func (gco *GCOptimizer) ForceGC(reason string) {
	start := time.Now()
	var beforeStats, afterStats runtime.MemStats

	runtime.ReadMemStats(&beforeStats)
	runtime.GC()
	runtime.ReadMemStats(&afterStats)

	duration := time.Since(start)
	freed := int64(beforeStats.HeapInuse - afterStats.HeapInuse)

	gco.logger.WithFields(map[string]interface{}{
		"reason":         reason,
		"duration_ms":    duration.Nanoseconds() / 1000000,
		"freed_bytes":    freed,
		"freed_mb":       freed / (1024 * 1024),
		"heap_before_mb": beforeStats.HeapInuse / (1024 * 1024),
		"heap_after_mb":  afterStats.HeapInuse / (1024 * 1024),
	}).Info("Forced garbage collection")
}

// GetStats returns current GC optimization statistics
func (gco *GCOptimizer) GetStats() *GCOptimizerStats {
	return &GCOptimizerStats{
		Enabled:        gco.enabled.Load(),
		MemoryPressure: gco.memoryPressure.Load(),
		GCStats:        &gco.gcStats,
	}
}

// GCOptimizerStats contains comprehensive GC optimization statistics
type GCOptimizerStats struct {
	Enabled        bool
	MemoryPressure int64
	GCStats        *GCStats
}

// MemoryEfficientProcessor provides memory-efficient data processing patterns
type MemoryEfficientProcessor struct {
	optimizer *GCOptimizer
	bufferMgr *BufferManager
	logger    log.Logger
}

// NewMemoryEfficientProcessor creates a processor optimized for memory efficiency
func NewMemoryEfficientProcessor(optimizer *GCOptimizer, logger log.Logger) *MemoryEfficientProcessor {
	return &MemoryEfficientProcessor{
		optimizer: optimizer,
		bufferMgr: NewBufferManager(),
		logger:    logger,
	}
}

// ProcessWithMinimalAllocation processes data with minimal memory allocation
func (mep *MemoryEfficientProcessor) ProcessWithMinimalAllocation(
	data []byte,
	processor func([]byte) ([]byte, error),
) ([]byte, error) {
	// Use object pooling for short-lived processing objects
	wrapper := mep.optimizer.GetObject(ShortLived)
	defer mep.optimizer.ReturnObject(wrapper)

	// Get reusable buffer for processing
	reusableBuffer := mep.bufferMgr.GetOptimalBuffer(int64(len(data)*2), "copy")
	defer reusableBuffer.Release()

	// Process data with minimal allocation
	return processor(data)
}

// BatchProcessWithGCControl processes data in batches with GC pressure monitoring
func (mep *MemoryEfficientProcessor) BatchProcessWithGCControl(
	items []interface{},
	processor func(interface{}) error,
	batchSize int,
) error {
	if batchSize <= 0 {
		batchSize = 100
	}

	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}

		// Process batch
		for j := i; j < end; j++ {
			if err := processor(items[j]); err != nil {
				return err
			}
		}

		// Check memory pressure and force GC if needed
		stats := mep.optimizer.GetStats()
		if stats.MemoryPressure > 90 {
			mep.optimizer.ForceGC("high memory pressure during batch processing")
		}

		// Yield to other goroutines
		runtime.Gosched()
	}

	return nil
}

// ObjectSizeCalculator provides utilities for calculating object memory usage
type ObjectSizeCalculator struct{}

// SizeOf calculates the approximate memory size of an object
func (osc *ObjectSizeCalculator) SizeOf(obj interface{}) uintptr {
	// This is a simplified implementation - in production you'd want more sophisticated analysis
	switch v := obj.(type) {
	case string:
		return unsafe.Sizeof(v) + uintptr(len(v))
	case []byte:
		return unsafe.Sizeof(v) + uintptr(len(v))
	case []interface{}:
		size := unsafe.Sizeof(v)
		for _, item := range v {
			size += osc.SizeOf(item)
		}
		return size
	default:
		return unsafe.Sizeof(obj)
	}
}

// GlobalGCOptimizer provides a singleton GC optimizer for application-wide use
var GlobalGCOptimizer = NewGCOptimizer(DefaultGCOptimizerConfig(), nil)

// OptimizeForContainerRegistry configures GC optimization specifically for container registry workloads
func OptimizeForContainerRegistry() *GCOptimizer {
	config := GCOptimizerConfig{
		LowMemoryGCPercent:  300,                    // Very relaxed for large image processing
		HighMemoryGCPercent: 25,                     // Aggressive for memory pressure
		MemoryPressureLimit: 2 * 1024 * 1024 * 1024, // 2GB threshold
		ShortLifeThreshold:  500 * time.Millisecond, // Blob processing
		MediumLifeThreshold: 30 * time.Second,       // Image processing
		MonitoringInterval:  3 * time.Second,
		GCStatsInterval:     15 * time.Second,
		ShortLivedPoolSize:  5000, // Many blob operations
		MediumLivedPoolSize: 500,  // Image operations
		LongLivedPoolSize:   50,   // Long-lived connections
	}

	return NewGCOptimizer(config, nil)
}
