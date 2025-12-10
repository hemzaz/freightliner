package network

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"runtime"
	"sync"
	"sync/atomic"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/helper/util"
)

const (
	// ParallelDefaultChunkSize is the default size for parallel compression chunks
	// Optimized for L2 cache (512KB per chunk)
	ParallelDefaultChunkSize = 512 * 1024

	// MinChunkSize is the minimum chunk size for parallel compression
	MinChunkSize = 64 * 1024

	// MaxChunkSize is the maximum chunk size
	MaxChunkSize = 4 * 1024 * 1024
)

// ParallelCompressor compresses data in parallel chunks for maximum throughput
type ParallelCompressor struct {
	config ParallelCompressionConfig
	logger log.Logger

	// Worker pool for parallel compression
	workers int
	jobChan chan *compressionJob
	wg      sync.WaitGroup

	// Statistics
	stats *ParallelCompressionStats
}

// ParallelCompressionConfig configures parallel compression behavior
type ParallelCompressionConfig struct {
	Workers          int              // Number of compression workers
	ChunkSize        int              // Size of each compression chunk
	CompressionLevel CompressionLevel // gzip compression level
	BufferPool       *util.BufferPool // Buffer pool for memory efficiency
}

// DefaultParallelCompressionConfig returns optimized defaults
func DefaultParallelCompressionConfig() ParallelCompressionConfig {
	return ParallelCompressionConfig{
		Workers:          runtime.NumCPU(), // One worker per CPU core
		ChunkSize:        ParallelDefaultChunkSize,
		CompressionLevel: BestSpeed, // Prioritize speed for network transfers
		BufferPool:       util.GlobalBufferPool,
	}
}

// ParallelCompressionStats tracks compression performance metrics
type ParallelCompressionStats struct {
	BytesProcessed   atomic.Int64
	BytesCompressed  atomic.Int64
	ChunksProcessed  atomic.Int64
	CompressionRatio atomic.Uint64 // Stored as uint64 representing float64 bits
}

// compressionJob represents a single chunk compression task
type compressionJob struct {
	id         int
	data       []byte
	level      int
	resultChan chan *compressionResult
	ctx        context.Context
}

// compressionResult contains the compressed chunk and metadata
type compressionResult struct {
	id             int
	compressed     []byte
	originalSize   int
	compressedSize int
	err            error
}

// NewParallelCompressor creates a new parallel compressor
func NewParallelCompressor(config ParallelCompressionConfig, logger log.Logger) *ParallelCompressor {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	if config.Workers <= 0 {
		config.Workers = runtime.NumCPU()
	}

	if config.ChunkSize < MinChunkSize {
		config.ChunkSize = MinChunkSize
	} else if config.ChunkSize > MaxChunkSize {
		config.ChunkSize = MaxChunkSize
	}

	if config.BufferPool == nil {
		config.BufferPool = util.GlobalBufferPool
	}

	pc := &ParallelCompressor{
		config:  config,
		logger:  logger,
		workers: config.Workers,
		jobChan: make(chan *compressionJob, config.Workers*2),
		stats:   &ParallelCompressionStats{},
	}

	// Start worker pool
	for i := 0; i < pc.workers; i++ {
		pc.wg.Add(1)
		go pc.compressionWorker(i)
	}

	pc.logger.WithFields(map[string]interface{}{
		"workers":    config.Workers,
		"chunk_size": config.ChunkSize,
		"level":      config.CompressionLevel,
	}).Info("Parallel compressor initialized")

	return pc
}

// CompressParallel compresses data using parallel workers
func (pc *ParallelCompressor) CompressParallel(ctx context.Context, reader io.Reader) (io.Reader, error) {
	// Read all data into memory first (for parallel processing)
	// In production, use streaming with overlapping reads
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read input data")
	}

	originalSize := len(data)
	if originalSize == 0 {
		return bytes.NewReader(nil), nil
	}

	pc.stats.BytesProcessed.Add(int64(originalSize))

	// If data is small, use single-threaded compression
	if originalSize < pc.config.ChunkSize*2 {
		return pc.compressSingleThreaded(data)
	}

	// Split data into chunks
	chunks := pc.splitIntoChunks(data)
	numChunks := len(chunks)

	// Create result channel
	resultChan := make(chan *compressionResult, numChunks)

	// Submit compression jobs
	for i, chunk := range chunks {
		job := &compressionJob{
			id:         i,
			data:       chunk,
			level:      int(pc.config.CompressionLevel),
			resultChan: resultChan,
			ctx:        ctx,
		}

		select {
		case pc.jobChan <- job:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Collect results in order
	results := make([]*compressionResult, numChunks)
	for i := 0; i < numChunks; i++ {
		select {
		case result := <-resultChan:
			if result.err != nil {
				return nil, errors.Wrap(result.err, "chunk compression failed")
			}
			results[result.id] = result
			pc.stats.ChunksProcessed.Add(1)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Combine compressed chunks
	var totalCompressed int64
	var combined bytes.Buffer

	for _, result := range results {
		combined.Write(result.compressed)
		totalCompressed += int64(result.compressedSize)
	}

	pc.stats.BytesCompressed.Add(totalCompressed)

	// Calculate compression ratio
	if originalSize > 0 {
		ratio := float64(totalCompressed) / float64(originalSize)
		pc.stats.CompressionRatio.Store(uint64(ratio * 1000000)) // Store as fixed-point
	}

	pc.logger.WithFields(map[string]interface{}{
		"original_size_mb":   float64(originalSize) / (1024 * 1024),
		"compressed_size_mb": float64(totalCompressed) / (1024 * 1024),
		"compression_ratio":  float64(totalCompressed) / float64(originalSize),
		"chunks":             numChunks,
	}).Debug("Parallel compression completed")

	return bytes.NewReader(combined.Bytes()), nil
}

// compressSingleThreaded compresses small data using single thread
func (pc *ParallelCompressor) compressSingleThreaded(data []byte) (io.Reader, error) {
	var buf bytes.Buffer

	w, err := gzip.NewWriterLevel(&buf, int(pc.config.CompressionLevel))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gzip writer")
	}

	if _, err := w.Write(data); err != nil {
		return nil, errors.Wrap(err, "failed to compress data")
	}

	if err := w.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to finalize compression")
	}

	pc.stats.BytesCompressed.Add(int64(buf.Len()))
	pc.stats.ChunksProcessed.Add(1)

	return bytes.NewReader(buf.Bytes()), nil
}

// splitIntoChunks divides data into optimal-sized chunks for parallel processing
func (pc *ParallelCompressor) splitIntoChunks(data []byte) [][]byte {
	totalSize := len(data)
	chunkSize := pc.config.ChunkSize

	// Calculate number of chunks
	numChunks := (totalSize + chunkSize - 1) / chunkSize
	chunks := make([][]byte, 0, numChunks)

	for i := 0; i < totalSize; i += chunkSize {
		end := i + chunkSize
		if end > totalSize {
			end = totalSize
		}
		chunks = append(chunks, data[i:end])
	}

	return chunks
}

// compressionWorker processes compression jobs from the job channel
func (pc *ParallelCompressor) compressionWorker(workerID int) {
	defer pc.wg.Done()

	for job := range pc.jobChan {
		result := pc.compressChunk(job)

		select {
		case job.resultChan <- result:
		case <-job.ctx.Done():
			return
		}
	}
}

// compressChunk compresses a single chunk of data
func (pc *ParallelCompressor) compressChunk(job *compressionJob) *compressionResult {
	result := &compressionResult{
		id:           job.id,
		originalSize: len(job.data),
	}

	// Use buffer pool for compression
	var buf bytes.Buffer

	w, err := gzip.NewWriterLevel(&buf, job.level)
	if err != nil {
		result.err = errors.Wrap(err, "failed to create compressor")
		return result
	}

	if _, err := w.Write(job.data); err != nil {
		result.err = errors.Wrap(err, "failed to compress chunk")
		return result
	}

	if err := w.Close(); err != nil {
		result.err = errors.Wrap(err, "failed to finalize compression")
		return result
	}

	result.compressed = buf.Bytes()
	result.compressedSize = len(result.compressed)

	return result
}

// Close shuts down the parallel compressor
func (pc *ParallelCompressor) Close() error {
	close(pc.jobChan)
	pc.wg.Wait()

	pc.logger.WithFields(map[string]interface{}{
		"bytes_processed_mb":  float64(pc.stats.BytesProcessed.Load()) / (1024 * 1024),
		"bytes_compressed_mb": float64(pc.stats.BytesCompressed.Load()) / (1024 * 1024),
		"chunks_processed":    pc.stats.ChunksProcessed.Load(),
	}).Info("Parallel compressor closed")

	return nil
}

// GetStats returns current compression statistics
func (pc *ParallelCompressor) GetStats() *ParallelCompressionStats {
	return pc.stats
}

// CompressParallelStream provides a streaming interface for parallel compression
func CompressParallelStream(ctx context.Context, reader io.Reader, level CompressionLevel, workers int) (io.Reader, error) {
	config := DefaultParallelCompressionConfig()
	config.CompressionLevel = level
	if workers > 0 {
		config.Workers = workers
	}

	logger := log.NewBasicLogger(log.InfoLevel)
	compressor := NewParallelCompressor(config, logger)
	defer compressor.Close()

	return compressor.CompressParallel(ctx, reader)
}
