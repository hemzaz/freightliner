package util

import (
	"io"
	"sync"
	"sync/atomic"
	"unsafe"
)

// ZeroCopyBufferPool provides zero-copy buffer operations for maximum performance
type ZeroCopyBufferPool struct {
	pools   map[int]*sync.Pool
	mu      sync.RWMutex
	metrics *ZeroCopyMetrics
}

// ZeroCopyMetrics tracks zero-copy operations
type ZeroCopyMetrics struct {
	ZeroCopyOperations atomic.Int64
	StandardOperations atomic.Int64
	BytesTransferred   atomic.Int64
	MemorySaved        atomic.Int64
}

// NewZeroCopyBufferPool creates a new zero-copy buffer pool
func NewZeroCopyBufferPool() *ZeroCopyBufferPool {
	pool := &ZeroCopyBufferPool{
		pools:   make(map[int]*sync.Pool),
		metrics: &ZeroCopyMetrics{},
	}

	// Pre-create pools for standard sizes optimized for registry operations
	sizes := []int{
		4096,      // 4KB - small metadata
		16384,     // 16KB - medium operations
		65536,     // 64KB - optimal for network operations
		262144,    // 256KB - large chunks
		1048576,   // 1MB - very large operations
		4194304,   // 4MB - blob chunks
		16777216,  // 16MB - large blobs
		52428800,  // 50MB - coordinated with transfer operations
		104857600, // 100MB - high-throughput transfers
	}

	for _, size := range sizes {
		pool.createPool(size)
	}

	return pool
}

// createPool creates a sync.Pool for a specific buffer size
func (z *ZeroCopyBufferPool) createPool(size int) {
	z.pools[size] = &sync.Pool{
		New: func() interface{} {
			buf := make([]byte, size)
			return &buf
		},
	}
}

// GetBuffer retrieves a buffer with zero-copy semantics
func (z *ZeroCopyBufferPool) GetBuffer(size int) *ZeroCopyBuffer {
	optimalSize := z.findOptimalSize(size)

	z.mu.RLock()
	pool, exists := z.pools[optimalSize]
	z.mu.RUnlock()

	if !exists {
		z.mu.Lock()
		if _, stillNotExists := z.pools[optimalSize]; stillNotExists {
			z.createPool(optimalSize)
		}
		pool = z.pools[optimalSize]
		z.mu.Unlock()
	}

	bufPtr := pool.Get().(*[]byte)
	buf := *bufPtr

	return &ZeroCopyBuffer{
		buf:          buf[:size],
		capacity:     optimalSize,
		pool:         pool,
		originalSize: optimalSize,
		metrics:      z.metrics,
	}
}

// findOptimalSize finds the smallest pool size that fits the request
func (z *ZeroCopyBufferPool) findOptimalSize(requested int) int {
	standardSizes := []int{
		4096, 16384, 65536, 262144, 1048576,
		4194304, 16777216, 52428800, 104857600,
	}

	for _, size := range standardSizes {
		if size >= requested {
			return size
		}
	}

	// Round up to next power of 2 for very large requests
	size := 1
	for size < requested {
		size <<= 1
	}
	return size
}

// ZeroCopyBuffer provides a buffer with zero-copy operations
type ZeroCopyBuffer struct {
	buf          []byte
	capacity     int
	pool         *sync.Pool
	originalSize int
	metrics      *ZeroCopyMetrics
	released     atomic.Bool
}

// Bytes returns the underlying byte slice (zero-copy)
func (z *ZeroCopyBuffer) Bytes() []byte {
	return z.buf
}

// Len returns the current length
func (z *ZeroCopyBuffer) Len() int {
	return len(z.buf)
}

// Cap returns the buffer capacity
func (z *ZeroCopyBuffer) Cap() int {
	return z.capacity
}

// Reset resets the buffer length to zero without reallocating
func (z *ZeroCopyBuffer) Reset() {
	z.buf = z.buf[:0]
}

// Resize changes the buffer length within capacity (zero-copy)
func (z *ZeroCopyBuffer) Resize(newLen int) error {
	if newLen > z.capacity {
		return ErrBufferTooSmall
	}
	z.buf = z.buf[:newLen]
	return nil
}

// ZeroCopyWriteTo writes data to writer without copying (implements io.WriterTo)
func (z *ZeroCopyBuffer) ZeroCopyWriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(z.buf)
	if err != nil {
		return int64(n), err
	}

	z.metrics.ZeroCopyOperations.Add(1)
	z.metrics.BytesTransferred.Add(int64(n))
	return int64(n), nil
}

// ZeroCopyReadFrom reads data from reader without extra copying (implements io.ReaderFrom)
func (z *ZeroCopyBuffer) ZeroCopyReadFrom(r io.Reader) (int64, error) {
	// Reset buffer to use full capacity
	z.buf = z.buf[:cap(z.buf)]

	n, err := r.Read(z.buf)
	if err != nil && err != io.EOF {
		return int64(n), err
	}

	z.buf = z.buf[:n]
	z.metrics.ZeroCopyOperations.Add(1)
	z.metrics.BytesTransferred.Add(int64(n))
	return int64(n), nil
}

// UnsafeString converts buffer to string without copying (use with caution!)
func (z *ZeroCopyBuffer) UnsafeString() string {
	return *(*string)(unsafe.Pointer(&z.buf))
}

// UnsafeBytes converts string to []byte without copying (use with caution!)
func UnsafeBytesFromString(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			int
		}{s, len(s)},
	))
}

// Release returns the buffer to the pool
func (z *ZeroCopyBuffer) Release() {
	if z.released.CompareAndSwap(false, true) {
		// Clear buffer for security
		for i := range z.buf[:cap(z.buf)] {
			z.buf[i] = 0
		}

		// Return to pool
		bufPtr := &z.buf
		z.pool.Put(bufPtr)
	}
}

// GetMetrics returns zero-copy operation metrics
func (z *ZeroCopyBufferPool) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"zero_copy_operations": z.metrics.ZeroCopyOperations.Load(),
		"standard_operations":  z.metrics.StandardOperations.Load(),
		"bytes_transferred_mb": float64(z.metrics.BytesTransferred.Load()) / (1024 * 1024),
		"memory_saved_mb":      float64(z.metrics.MemorySaved.Load()) / (1024 * 1024),
	}
}

// StreamCopier provides optimized stream copying with minimal allocations
type StreamCopier struct {
	bufferPool *ZeroCopyBufferPool
	bufferSize int
}

// NewStreamCopier creates a new optimized stream copier
func NewStreamCopier(bufferSize int) *StreamCopier {
	if bufferSize <= 0 {
		bufferSize = 64 * 1024 // 64KB default
	}

	return &StreamCopier{
		bufferPool: NewZeroCopyBufferPool(),
		bufferSize: bufferSize,
	}
}

// Copy performs an optimized copy between reader and writer
func (s *StreamCopier) Copy(dst io.Writer, src io.Reader) (int64, error) {
	// Try zero-copy path first
	if wt, ok := src.(io.WriterTo); ok {
		return wt.WriteTo(dst)
	}
	if rf, ok := dst.(io.ReaderFrom); ok {
		return rf.ReadFrom(src)
	}

	// Use buffered copy with pooled buffers
	buf := s.bufferPool.GetBuffer(s.bufferSize)
	defer buf.Release()

	var total int64
	for {
		nr, er := src.Read(buf.Bytes())
		if nr > 0 {
			buf.Resize(nr)
			nw, ew := dst.Write(buf.Bytes())
			if nw > 0 {
				total += int64(nw)
			}
			if ew != nil {
				return total, ew
			}
			if nr != nw {
				return total, io.ErrShortWrite
			}
		}
		if er != nil {
			if er != io.EOF {
				return total, er
			}
			break
		}
	}

	return total, nil
}

// CopyBuffer performs an optimized copy with a provided buffer
func (s *StreamCopier) CopyBuffer(dst io.Writer, src io.Reader, buf []byte) (int64, error) {
	if buf == nil || len(buf) == 0 {
		return s.Copy(dst, src)
	}

	// Try zero-copy path first
	if wt, ok := src.(io.WriterTo); ok {
		return wt.WriteTo(dst)
	}
	if rf, ok := dst.(io.ReaderFrom); ok {
		return rf.ReadFrom(src)
	}

	// Use provided buffer
	var total int64
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[:nr])
			if nw > 0 {
				total += int64(nw)
			}
			if ew != nil {
				return total, ew
			}
			if nr != nw {
				return total, io.ErrShortWrite
			}
		}
		if er != nil {
			if er != io.EOF {
				return total, er
			}
			break
		}
	}

	return total, nil
}

// GlobalZeroCopyPool is a global instance for application-wide use
var GlobalZeroCopyPool = NewZeroCopyBufferPool()

// GetZeroCopyBuffer is a convenience function to get a zero-copy buffer from the global pool
func GetZeroCopyBuffer(size int) *ZeroCopyBuffer {
	return GlobalZeroCopyPool.GetBuffer(size)
}
