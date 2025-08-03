package util

import (
	"bytes"
	"errors"
	"sync"
)

// BufferPool manages a pool of reusable byte buffers for memory optimization
type BufferPool struct {
	pools map[int]*sync.Pool
	mutex sync.RWMutex
}

// GlobalBufferPool is a singleton instance for application-wide buffer reuse
var GlobalBufferPool = NewBufferPool()

// NewBufferPool creates a new buffer pool with standard sizes
func NewBufferPool() *BufferPool {
	bp := &BufferPool{
		pools: make(map[int]*sync.Pool),
	}

	// Pre-create pools for common buffer sizes (powers of 2 for optimal memory alignment)
	// Optimized for container registry operations targeting 100-150 MB/s throughput
	standardSizes := []int{
		1024,      // 1KB - small operations
		4096,      // 4KB - page size
		16384,     // 16KB - medium operations
		65536,     // 64KB - network buffers (optimal for TCP)
		262144,    // 256KB - large operations
		1048576,   // 1MB - very large operations
		4194304,   // 4MB - chunk processing
		16777216,  // 16MB - large layer processing
		52428800,  // 50MB - coordinated with transfer.go buffer size
		104857600, // 100MB - high-throughput operations
		209715200, // 200MB - very large layer processing
	}

	for _, size := range standardSizes {
		bp.createPoolForSize(size)
	}

	return bp
}

// createPoolForSize creates a new pool for a specific buffer size
func (bp *BufferPool) createPoolForSize(size int) {
	bp.pools[size] = &sync.Pool{
		New: func() interface{} {
			return make([]byte, size)
		},
	}
}

// Get retrieves a buffer of at least the specified size
// Returns the actual buffer size which may be larger than requested
func (bp *BufferPool) Get(size int) ([]byte, int) {
	bp.mutex.RLock()
	defer bp.mutex.RUnlock()

	// Find the smallest buffer size that fits the request
	actualSize := bp.findOptimalSize(size)

	pool, exists := bp.pools[actualSize]
	if !exists {
		// If no pool exists for this size, create one
		bp.mutex.RUnlock()
		bp.mutex.Lock()
		if _, stillNotExists := bp.pools[actualSize]; stillNotExists {
			bp.createPoolForSize(actualSize)
		}
		pool = bp.pools[actualSize]
		bp.mutex.Unlock()
		bp.mutex.RLock()
	}

	buffer := pool.Get().([]byte)
	// Reset buffer length to avoid stale data
	if cap(buffer) >= size {
		buffer = buffer[:size]
	}

	return buffer, actualSize
}

// Put returns a buffer to the pool for reuse
func (bp *BufferPool) Put(buffer []byte, originalSize int) {
	bp.mutex.RLock()
	defer bp.mutex.RUnlock()

	if pool, exists := bp.pools[originalSize]; exists {
		// Reset buffer to full capacity and clear content for security
		buffer = buffer[:cap(buffer)]
		for i := range buffer {
			buffer[i] = 0
		}
		pool.Put(buffer)
	}
	// If pool doesn't exist, let buffer be garbage collected
}

// findOptimalSize finds the smallest standard size that fits the request
func (bp *BufferPool) findOptimalSize(requestedSize int) int {
	// Check against our standard sizes first for optimal performance
	standardSizes := []int{
		1024, 4096, 16384, 65536, 262144, 1048576,
		4194304, 16777216, 52428800, 104857600, 209715200,
	}

	for _, size := range standardSizes {
		if size >= requestedSize {
			return size
		}
	}

	// For very large requests beyond our standard sizes, use the exact size
	return requestedSize
}

// GetStats returns usage statistics for monitoring
func (bp *BufferPool) GetStats() map[int]int {
	bp.mutex.RLock()
	defer bp.mutex.RUnlock()

	stats := make(map[int]int)
	for size := range bp.pools {
		stats[size] = 0 // In a real implementation, we'd track usage counts
	}
	return stats
}

// BytesBufferPool manages a pool of bytes.Buffer for text operations
type BytesBufferPool struct {
	pool sync.Pool
}

// GlobalBytesBufferPool is a singleton for application-wide bytes.Buffer reuse
var GlobalBytesBufferPool = &BytesBufferPool{
	pool: sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	},
}

// Get retrieves a clean bytes.Buffer from the pool
func (bbp *BytesBufferPool) Get() *bytes.Buffer {
	buffer := bbp.pool.Get().(*bytes.Buffer)
	buffer.Reset() // Ensure buffer is clean
	return buffer
}

// Put returns a bytes.Buffer to the pool
func (bbp *BytesBufferPool) Put(buffer *bytes.Buffer) {
	// Only return buffers that aren't too large to avoid memory bloat
	const maxBufferSize = 1024 * 1024 // 1MB max retained size
	if buffer.Cap() <= maxBufferSize {
		bbp.pool.Put(buffer)
	}
	// Buffers larger than maxBufferSize are allowed to be garbage collected
}

// ObjectPool provides a generic pool for arbitrary objects
type ObjectPool struct {
	pool sync.Pool
}

// NewObjectPool creates a new generic object pool
func NewObjectPool(factory func() interface{}) *ObjectPool {
	return &ObjectPool{
		pool: sync.Pool{
			New: factory,
		},
	}
}

// Get retrieves an object from the pool
func (op *ObjectPool) Get() interface{} {
	return op.pool.Get()
}

// Put returns an object to the pool
func (op *ObjectPool) Put(obj interface{}) {
	op.pool.Put(obj)
}

// MemoryOptimizer provides helpers for memory-efficient operations
type MemoryOptimizer struct {
	bufferPool      *BufferPool
	bytesBufferPool *BytesBufferPool
}

// NewMemoryOptimizer creates a new memory optimizer
func NewMemoryOptimizer() *MemoryOptimizer {
	return &MemoryOptimizer{
		bufferPool:      GlobalBufferPool,
		bytesBufferPool: GlobalBytesBufferPool,
	}
}

// OptimizedCopy performs memory-efficient copying with buffer reuse
func (mo *MemoryOptimizer) OptimizedCopy(dst, src []byte) int {
	n := copy(dst, src)
	return n
}

// StreamingCopy performs streaming copy with optimized buffering
func (mo *MemoryOptimizer) StreamingCopy(dst interface{}, src interface{}, size int64) (int64, error) {
	// This would implement optimized streaming copy
	// For now, return placeholder
	return size, nil
}

// ReusableBuffer provides a wrapper for managing buffer lifecycle
type ReusableBuffer struct {
	buffer       []byte
	originalSize int
	pool         *BufferPool
}

// NewReusableBuffer creates a new reusable buffer
func NewReusableBuffer(size int) *ReusableBuffer {
	buffer, actualSize := GlobalBufferPool.Get(size)
	return &ReusableBuffer{
		buffer:       buffer,
		originalSize: actualSize,
		pool:         GlobalBufferPool,
	}
}

// Bytes returns the underlying byte slice
func (rb *ReusableBuffer) Bytes() []byte {
	return rb.buffer
}

// Len returns the current length of the buffer
func (rb *ReusableBuffer) Len() int {
	return len(rb.buffer)
}

// Cap returns the capacity of the buffer
func (rb *ReusableBuffer) Cap() int {
	return cap(rb.buffer)
}

// Release returns the buffer to the pool
func (rb *ReusableBuffer) Release() {
	if rb.buffer != nil {
		rb.pool.Put(rb.buffer, rb.originalSize)
		rb.buffer = nil
	}
}

// Resize adjusts the buffer length (within capacity)
func (rb *ReusableBuffer) Resize(newLen int) error {
	if newLen > cap(rb.buffer) {
		return ErrBufferTooSmall
	}
	rb.buffer = rb.buffer[:newLen]
	return nil
}

// Custom errors for buffer management
var (
	ErrBufferTooSmall = errors.New("buffer too small for requested operation")
)

// BufferManager provides high-level buffer management operations
type BufferManager struct {
	optimizer *MemoryOptimizer
}

// NewBufferManager creates a new buffer manager
func NewBufferManager() *BufferManager {
	return &BufferManager{
		optimizer: NewMemoryOptimizer(),
	}
}

// GetOptimalBuffer returns an optimally-sized buffer for the operation
func (bm *BufferManager) GetOptimalBuffer(estimatedSize int64, operation string) *ReusableBuffer {
	// Adjust buffer size based on operation type
	var targetSize int

	switch operation {
	case "compress":
		// Compression benefits from larger buffers
		targetSize = int(min(estimatedSize*2, 4194304)) // Up to 4MB
	case "network":
		// Network operations work well with 64KB buffers
		targetSize = int(min(estimatedSize, 65536))
	case "copy":
		// Copy operations benefit from 1MB buffers for large data
		targetSize = int(min(estimatedSize, 1048576))
	default:
		// Default to estimated size
		targetSize = int(estimatedSize)
	}

	return NewReusableBuffer(targetSize)
}

// min returns the minimum of two int64 values
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
