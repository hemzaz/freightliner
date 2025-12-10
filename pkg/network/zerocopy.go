package network

import (
	"io"
	"net"
	"runtime"
	"sync"

	"github.com/valyala/bytebufferpool"
)

// Global buffer pool for zero-copy transfers
var bufferPool = bytebufferpool.Pool{}

// ZeroCopyConfig configures zero-copy behavior
type ZeroCopyConfig struct {
	BufferSize     int
	EnableSplice   bool // Linux-specific optimization
	EnableSendfile bool // Unix-specific optimization
	ReuseBuffers   bool
}

// DefaultZeroCopyConfig returns sensible defaults
func DefaultZeroCopyConfig() *ZeroCopyConfig {
	return &ZeroCopyConfig{
		BufferSize:     64 * 1024, // 64 KB
		EnableSplice:   runtime.GOOS == "linux",
		EnableSendfile: runtime.GOOS == "linux" || runtime.GOOS == "darwin",
		ReuseBuffers:   true,
	}
}

// CopyWithZeroCopy performs optimized data copy with minimal allocations
func CopyWithZeroCopy(dst io.Writer, src io.Reader) (int64, error) {
	config := DefaultZeroCopyConfig()
	return CopyWithZeroCopyConfig(dst, src, config)
}

// CopyWithZeroCopyConfig performs zero-copy with custom configuration
func CopyWithZeroCopyConfig(dst io.Writer, src io.Reader, config *ZeroCopyConfig) (int64, error) {
	// Try kernel-level optimizations first (Linux splice, Unix sendfile)
	if config.EnableSplice || config.EnableSendfile {
		if n, err := tryKernelCopy(dst, src); err == nil {
			return n, nil
		}
	}

	// Fall back to optimized userspace copy
	if config.ReuseBuffers {
		return copyWithBufferPool(dst, src, config.BufferSize)
	}

	// Standard copy with fixed buffer
	buf := make([]byte, config.BufferSize)
	return io.CopyBuffer(dst, src, buf)
}

// tryKernelCopy attempts to use kernel-level zero-copy
func tryKernelCopy(dst io.Writer, src io.Reader) (int64, error) {
	// Check if source is TCP connection (splice/sendfile eligible)
	if tc, ok := src.(*net.TCPConn); ok {
		if runtime.GOOS == "linux" {
			// On Linux, io.Copy uses splice() for TCPConn
			return io.Copy(dst, tc)
		}
	}

	// Not eligible for kernel-level optimization
	return 0, io.ErrShortWrite
}

// copyWithBufferPool uses buffer pool to minimize allocations
func copyWithBufferPool(dst io.Writer, src io.Reader, bufSize int) (int64, error) {
	// Get buffer from pool
	buf := bufferPool.Get()
	defer bufferPool.Put(buf)

	// Ensure buffer has sufficient capacity
	if cap(buf.B) < bufSize {
		buf.B = make([]byte, bufSize)
	} else {
		buf.B = buf.B[:bufSize]
	}

	// Perform copy
	return io.CopyBuffer(dst, src, buf.B)
}

// MultiCopy performs parallel copy from multiple sources to multiple destinations
func MultiCopy(pairs []CopyPair) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(pairs))

	for _, pair := range pairs {
		wg.Add(1)
		go func(p CopyPair) {
			defer wg.Done()
			_, err := CopyWithZeroCopy(p.Dst, p.Src)
			if err != nil {
				errChan <- err
			}
		}(pair)
	}

	wg.Wait()
	close(errChan)

	// Return first error if any
	for err := range errChan {
		return err
	}

	return nil
}

// CopyPair represents source and destination for copy operation
type CopyPair struct {
	Src io.Reader
	Dst io.Writer
}

// BufferedWriter wraps a writer with buffering using buffer pool
type BufferedWriter struct {
	w   io.Writer
	buf *bytebufferpool.ByteBuffer
	mu  sync.Mutex
}

// NewBufferedWriter creates a buffered writer using buffer pool
func NewBufferedWriter(w io.Writer) *BufferedWriter {
	return &BufferedWriter{
		w:   w,
		buf: bufferPool.Get(),
	}
}

// Write writes data to buffer
func (bw *BufferedWriter) Write(p []byte) (int, error) {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	return bw.buf.Write(p)
}

// Flush flushes buffered data to underlying writer
func (bw *BufferedWriter) Flush() error {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	if bw.buf.Len() == 0 {
		return nil
	}

	_, err := bw.w.Write(bw.buf.B)
	if err != nil {
		return err
	}

	bw.buf.Reset()
	return nil
}

// Close flushes and returns buffer to pool
func (bw *BufferedWriter) Close() error {
	if err := bw.Flush(); err != nil {
		return err
	}

	bufferPool.Put(bw.buf)
	bw.buf = nil
	return nil
}

// StreamCopier handles multiple concurrent copy operations
type StreamCopier struct {
	workers int
	jobs    chan CopyJob
	results chan CopyResult
	wg      sync.WaitGroup
}

// CopyJob represents a copy job
type CopyJob struct {
	ID  string
	Src io.Reader
	Dst io.Writer
}

// CopyResult represents copy operation result
type CopyResult struct {
	ID    string
	Bytes int64
	Error error
}

// NewStreamCopier creates a stream copier with worker pool
func NewStreamCopier(workers int) *StreamCopier {
	sc := &StreamCopier{
		workers: workers,
		jobs:    make(chan CopyJob, workers*2),
		results: make(chan CopyResult, workers*2),
	}

	// Start workers
	for i := 0; i < workers; i++ {
		sc.wg.Add(1)
		go sc.worker()
	}

	return sc
}

// worker processes copy jobs
func (sc *StreamCopier) worker() {
	defer sc.wg.Done()

	for job := range sc.jobs {
		n, err := CopyWithZeroCopy(job.Dst, job.Src)
		sc.results <- CopyResult{
			ID:    job.ID,
			Bytes: n,
			Error: err,
		}
	}
}

// Submit submits a copy job
func (sc *StreamCopier) Submit(job CopyJob) {
	sc.jobs <- job
}

// Results returns the results channel
func (sc *StreamCopier) Results() <-chan CopyResult {
	return sc.results
}

// Close closes the stream copier
func (sc *StreamCopier) Close() {
	close(sc.jobs)
	sc.wg.Wait()
	close(sc.results)
}

// GetBufferPoolStats returns buffer pool statistics
func GetBufferPoolStats() BufferPoolStats {
	// Note: bytebufferpool doesn't expose stats, so we track separately
	// This is a placeholder for custom buffer pool implementation
	return BufferPoolStats{
		BuffersInUse:  0,
		BuffersInPool: 0,
		TotalGets:     0,
		TotalPuts:     0,
	}
}

// BufferPoolStats tracks buffer pool statistics
type BufferPoolStats struct {
	BuffersInUse  int
	BuffersInPool int
	TotalGets     int64
	TotalPuts     int64
}
