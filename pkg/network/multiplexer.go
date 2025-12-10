package network

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// StreamMultiplexer manages parallel downloads using HTTP/3 streams
type StreamMultiplexer struct {
	transport  *HTTP3Transport
	maxStreams int
	stats      *MultiplexerStats
}

// MultiplexerStats tracks multiplexer performance metrics
type MultiplexerStats struct {
	TotalLayers     int64
	CompletedLayers int64
	FailedLayers    int64
	TotalBytes      int64
	ParallelStreams int64
	AverageLatency  time.Duration
	mu              sync.RWMutex
}

// LayerDescriptor describes a layer to download
type LayerDescriptor struct {
	URL      string
	Digest   string
	Size     int64
	Writer   io.Writer
	Priority int // Higher priority downloads first
}

// MultiplexerConfig configures stream multiplexer
type MultiplexerConfig struct {
	MaxStreams     int
	StreamTimeout  time.Duration
	RetryAttempts  int
	BufferSize     int
	EnablePriority bool
}

// DefaultMultiplexerConfig returns sensible defaults
func DefaultMultiplexerConfig() *MultiplexerConfig {
	return &MultiplexerConfig{
		MaxStreams:     100,
		StreamTimeout:  30 * time.Second,
		RetryAttempts:  3,
		BufferSize:     64 * 1024, // 64 KB
		EnablePriority: true,
	}
}

// NewStreamMultiplexer creates a new stream multiplexer
func NewStreamMultiplexer(transport *HTTP3Transport, config *MultiplexerConfig) *StreamMultiplexer {
	if config == nil {
		config = DefaultMultiplexerConfig()
	}

	return &StreamMultiplexer{
		transport:  transport,
		maxStreams: config.MaxStreams,
		stats:      &MultiplexerStats{},
	}
}

// DownloadLayers downloads multiple layers in parallel using HTTP/3 streams
func (m *StreamMultiplexer) DownloadLayers(ctx context.Context, layers []LayerDescriptor) error {
	if len(layers) == 0 {
		return nil
	}

	atomic.AddInt64(&m.stats.TotalLayers, int64(len(layers)))

	// Sort by priority if enabled
	sortedLayers := m.prioritizeLayers(layers)

	// Create semaphore to limit concurrent streams
	sem := make(chan struct{}, m.maxStreams)
	errChan := make(chan error, len(sortedLayers))
	var wg sync.WaitGroup

	atomic.StoreInt64(&m.stats.ParallelStreams, int64(len(sortedLayers)))

	for _, layer := range sortedLayers {
		wg.Add(1)

		// Acquire semaphore
		sem <- struct{}{}

		go func(l LayerDescriptor) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			start := time.Now()
			err := m.downloadLayer(ctx, l)
			latency := time.Since(start)

			if err != nil {
				atomic.AddInt64(&m.stats.FailedLayers, 1)
				errChan <- fmt.Errorf("download layer %s: %w", l.Digest, err)
			} else {
				atomic.AddInt64(&m.stats.CompletedLayers, 1)
				atomic.AddInt64(&m.stats.TotalBytes, l.Size)
				m.updateLatency(latency)
			}
		}(layer)
	}

	// Wait for all downloads to complete
	wg.Wait()
	close(errChan)

	// Collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("parallel download errors: %v", errs)
	}

	return nil
}

// downloadLayer downloads a single layer
func (m *StreamMultiplexer) downloadLayer(ctx context.Context, layer LayerDescriptor) error {
	config := DefaultMultiplexerConfig()

	var lastErr error
	for attempt := 0; attempt < config.RetryAttempts; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(attempt*attempt) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		// Create context with timeout
		downloadCtx, cancel := context.WithTimeout(ctx, config.StreamTimeout)

		_, err := m.transport.StreamDownload(downloadCtx, layer.URL, layer.Writer)
		cancel()

		if err == nil {
			return nil
		}

		lastErr = err
	}

	return fmt.Errorf("failed after %d attempts: %w", config.RetryAttempts, lastErr)
}

// prioritizeLayers sorts layers by priority (higher first)
func (m *StreamMultiplexer) prioritizeLayers(layers []LayerDescriptor) []LayerDescriptor {
	config := DefaultMultiplexerConfig()
	if !config.EnablePriority {
		return layers
	}

	// Create a copy to avoid modifying original
	sorted := make([]LayerDescriptor, len(layers))
	copy(sorted, layers)

	// Simple bubble sort by priority (sufficient for small layer counts)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Priority < sorted[j+1].Priority {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// updateLatency updates average latency
func (m *StreamMultiplexer) updateLatency(latency time.Duration) {
	m.stats.mu.Lock()
	defer m.stats.mu.Unlock()

	// Simple moving average
	if m.stats.AverageLatency == 0 {
		m.stats.AverageLatency = latency
	} else {
		m.stats.AverageLatency = (m.stats.AverageLatency + latency) / 2
	}
}

// GetStats returns current multiplexer statistics
func (m *StreamMultiplexer) GetStats() MultiplexerStats {
	m.stats.mu.RLock()
	defer m.stats.mu.RUnlock()

	return MultiplexerStats{
		TotalLayers:     atomic.LoadInt64(&m.stats.TotalLayers),
		CompletedLayers: atomic.LoadInt64(&m.stats.CompletedLayers),
		FailedLayers:    atomic.LoadInt64(&m.stats.FailedLayers),
		TotalBytes:      atomic.LoadInt64(&m.stats.TotalBytes),
		ParallelStreams: atomic.LoadInt64(&m.stats.ParallelStreams),
		AverageLatency:  m.stats.AverageLatency,
	}
}

// Reset resets statistics
func (m *StreamMultiplexer) Reset() {
	atomic.StoreInt64(&m.stats.TotalLayers, 0)
	atomic.StoreInt64(&m.stats.CompletedLayers, 0)
	atomic.StoreInt64(&m.stats.FailedLayers, 0)
	atomic.StoreInt64(&m.stats.TotalBytes, 0)
	atomic.StoreInt64(&m.stats.ParallelStreams, 0)

	m.stats.mu.Lock()
	m.stats.AverageLatency = 0
	m.stats.mu.Unlock()
}

// StreamBatch represents a batch of streams to download
type StreamBatch struct {
	Layers   []LayerDescriptor
	Priority int
	Context  context.Context
}

// BatchMultiplexer manages multiple batches of parallel downloads
type BatchMultiplexer struct {
	multiplexer *StreamMultiplexer
	batches     chan StreamBatch
	workers     int
}

// NewBatchMultiplexer creates a batch multiplexer
func NewBatchMultiplexer(transport *HTTP3Transport, workers int) *BatchMultiplexer {
	return &BatchMultiplexer{
		multiplexer: NewStreamMultiplexer(transport, nil),
		batches:     make(chan StreamBatch, workers*2),
		workers:     workers,
	}
}

// Start starts the batch processor
func (bm *BatchMultiplexer) Start(ctx context.Context) {
	for i := 0; i < bm.workers; i++ {
		go bm.worker(ctx)
	}
}

// worker processes batches
func (bm *BatchMultiplexer) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case batch, ok := <-bm.batches:
			if !ok {
				return
			}

			// Process batch
			if err := bm.multiplexer.DownloadLayers(batch.Context, batch.Layers); err != nil {
				// Log error (would use proper logger in production)
				fmt.Printf("batch download error: %v\n", err)
			}
		}
	}
}

// SubmitBatch submits a batch for processing
func (bm *BatchMultiplexer) SubmitBatch(batch StreamBatch) error {
	select {
	case bm.batches <- batch:
		return nil
	default:
		return fmt.Errorf("batch queue full")
	}
}

// Close closes the batch multiplexer
func (bm *BatchMultiplexer) Close() {
	close(bm.batches)
}
