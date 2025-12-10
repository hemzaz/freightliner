package network

import (
	"bufio"
	"io"
	"sync"

	"freightliner/pkg/helper/util"
)

// StreamingBufferPool provides high-performance buffer pools for network streaming operations
type StreamingBufferPool struct {
	bufferManager    *util.BufferManager
	readWriteBuffers *util.ObjectPool // Pool of bufio.ReadWriter
	readerBuffers    *util.ObjectPool // Pool of bufio.Reader
	writerBuffers    *util.ObjectPool // Pool of bufio.Writer
}

// NewStreamingBufferPool creates a new streaming buffer pool optimized for network operations
func NewStreamingBufferPool() *StreamingBufferPool {
	return &StreamingBufferPool{
		bufferManager: util.NewBufferManager(),
		readWriteBuffers: util.NewObjectPool(func() interface{} {
			// Create bufio.ReadWriter with optimal buffer sizes for network operations
			reader := bufio.NewReaderSize(nil, 65536) // 64KB read buffer
			writer := bufio.NewWriterSize(nil, 65536) // 64KB write buffer
			return bufio.NewReadWriter(reader, writer)
		}),
		readerBuffers: util.NewObjectPool(func() interface{} {
			return bufio.NewReaderSize(nil, 65536) // 64KB read buffer
		}),
		writerBuffers: util.NewObjectPool(func() interface{} {
			return bufio.NewWriterSize(nil, 65536) // 64KB write buffer
		}),
	}
}

// GetOptimizedReader returns a buffered reader from the pool for streaming operations
func (sbp *StreamingBufferPool) GetOptimizedReader(r io.Reader) *OptimizedReader {
	reader := sbp.readerBuffers.Get().(*bufio.Reader)
	reader.Reset(r)
	return &OptimizedReader{
		Reader: reader,
		pool:   sbp.readerBuffers,
	}
}

// GetOptimizedWriter returns a buffered writer from the pool for streaming operations
func (sbp *StreamingBufferPool) GetOptimizedWriter(w io.Writer) *OptimizedWriter {
	writer := sbp.writerBuffers.Get().(*bufio.Writer)
	writer.Reset(w)
	return &OptimizedWriter{
		Writer: writer,
		pool:   sbp.writerBuffers,
	}
}

// GetOptimizedReadWriter returns a buffered read-writer from the pool
func (sbp *StreamingBufferPool) GetOptimizedReadWriter(rw io.ReadWriter) *OptimizedReadWriter {
	readWriter := sbp.readWriteBuffers.Get().(*bufio.ReadWriter)
	readWriter.Reader.Reset(rw)
	readWriter.Writer.Reset(rw)
	return &OptimizedReadWriter{
		ReadWriter: readWriter,
		pool:       sbp.readWriteBuffers,
	}
}

// OptimizedReader wraps a buffered reader with automatic pool return
type OptimizedReader struct {
	*bufio.Reader
	pool *util.ObjectPool
	once sync.Once
}

// Release returns the reader to the pool (can be called multiple times safely)
func (or *OptimizedReader) Release() {
	or.once.Do(func() {
		or.pool.Put(or.Reader)
	})
}

// OptimizedWriter wraps a buffered writer with automatic pool return
type OptimizedWriter struct {
	*bufio.Writer
	pool *util.ObjectPool
	once sync.Once
}

// Release flushes and returns the writer to the pool (can be called multiple times safely)
func (ow *OptimizedWriter) Release() {
	ow.once.Do(func() {
		_ = ow.Writer.Flush() // Ensure data is flushed before returning to pool
		ow.pool.Put(ow.Writer)
	})
}

// OptimizedReadWriter wraps a buffered read-writer with automatic pool return
type OptimizedReadWriter struct {
	*bufio.ReadWriter
	pool *util.ObjectPool
	once sync.Once
}

// Release flushes and returns the read-writer to the pool (can be called multiple times safely)
func (orw *OptimizedReadWriter) Release() {
	orw.once.Do(func() {
		_ = orw.Writer.Flush() // Ensure data is flushed before returning to pool
		orw.pool.Put(orw.ReadWriter)
	})
}

// StreamingCopier provides high-performance copying with buffer reuse
type StreamingCopier struct {
	pool       *StreamingBufferPool
	bufferSize int64
}

// NewStreamingCopier creates a new streaming copier with optimal buffer management
func NewStreamingCopier(bufferSize int64) *StreamingCopier {
	if bufferSize <= 0 {
		bufferSize = 65536 // 64KB default
	}
	return &StreamingCopier{
		pool:       NewStreamingBufferPool(),
		bufferSize: bufferSize,
	}
}

// CopyWithOptimizedBuffer performs memory-efficient copy operations using buffer pools
func (sc *StreamingCopier) CopyWithOptimizedBuffer(dst io.Writer, src io.Reader) (int64, error) {
	// Get optimized reader and writer from pools
	optimizedReader := sc.pool.GetOptimizedReader(src)
	defer optimizedReader.Release()

	optimizedWriter := sc.pool.GetOptimizedWriter(dst)
	defer optimizedWriter.Release()

	// Use buffered copy for optimal performance
	return io.Copy(optimizedWriter, optimizedReader)
}

// StreamThroughPipe creates an optimized pipe for streaming operations
func (sc *StreamingCopier) StreamThroughPipe(src io.Reader, transform func(io.Reader) io.Reader) io.Reader {
	pr, pw := io.Pipe()

	go func() {
		defer func() {
			if err := pw.Close(); err != nil {
				// Log close error but don't propagate since we're in a goroutine
				// and the main operation may have already failed
			}
		}()

		// Apply transformation and use optimized copying
		transformedReader := transform(src)
		optimizedReader := sc.pool.GetOptimizedReader(transformedReader)
		defer optimizedReader.Release()

		optimizedWriter := sc.pool.GetOptimizedWriter(pw)
		defer optimizedWriter.Release()

		_, err := io.Copy(optimizedWriter, optimizedReader)
		if err != nil {
			pw.CloseWithError(err)
		}
	}()

	return pr
}

// GlobalStreamingPool is a singleton instance for application-wide streaming optimization
var GlobalStreamingPool = NewStreamingBufferPool()

// OptimizedStreamCopy provides a global function for optimized streaming copy
func OptimizedStreamCopy(dst io.Writer, src io.Reader) (int64, error) {
	copier := &StreamingCopier{pool: GlobalStreamingPool}
	return copier.CopyWithOptimizedBuffer(dst, src)
}

// ChunkedStreamProcessor processes data in chunks with memory-efficient buffering
type ChunkedStreamProcessor struct {
	pool      *StreamingBufferPool
	chunkSize int64
	bufferMgr *util.BufferManager
}

// NewChunkedStreamProcessor creates a processor for chunk-based stream processing
func NewChunkedStreamProcessor(chunkSize int64) *ChunkedStreamProcessor {
	if chunkSize <= 0 {
		chunkSize = 1048576 // 1MB default chunk size
	}
	return &ChunkedStreamProcessor{
		pool:      GlobalStreamingPool,
		chunkSize: chunkSize,
		bufferMgr: util.NewBufferManager(),
	}
}

// ProcessInChunks processes a stream in memory-efficient chunks
func (csp *ChunkedStreamProcessor) ProcessInChunks(
	src io.Reader,
	processor func(chunk []byte) ([]byte, error),
	dst io.Writer,
) (int64, error) {
	// Get reusable buffer for chunk processing
	reusableBuffer := csp.bufferMgr.GetOptimalBuffer(csp.chunkSize, "copy")
	defer reusableBuffer.Release()

	optimizedReader := csp.pool.GetOptimizedReader(src)
	defer optimizedReader.Release()

	optimizedWriter := csp.pool.GetOptimizedWriter(dst)
	defer optimizedWriter.Release()

	buffer := reusableBuffer.Bytes()
	var totalBytes int64

	for {
		n, readErr := optimizedReader.Read(buffer)
		if n > 0 {
			// Process the chunk
			processedChunk, processErr := processor(buffer[:n])
			if processErr != nil {
				return totalBytes, processErr
			}

			// Write processed chunk
			written, writeErr := optimizedWriter.Write(processedChunk)
			if writeErr != nil {
				return totalBytes, writeErr
			}
			totalBytes += int64(written)
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return totalBytes, readErr
		}
	}

	return totalBytes, nil
}
