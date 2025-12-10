package sync

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"freightliner/pkg/client"
	copyutil "freightliner/pkg/copy"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/replication"
	"freightliner/pkg/service"

	"github.com/google/go-containerregistry/pkg/name"
)

// BatchExecutor executes sync tasks in optimized batches
type BatchExecutor struct {
	config      *Config
	logger      log.Logger
	scheduler   *replication.Scheduler
	results     []SyncResult
	mu          sync.Mutex
	factory     *client.Factory
	clientCache map[string]service.RegistryClient // Cache clients by registry URL
	cacheMu     sync.RWMutex                      // Protect client cache

	// Adaptive batching state
	currentBatchSize int        // Current batch size (adjusted dynamically)
	batchStats       batchStat  // Statistics from previous batches
	statsMu          sync.Mutex // Protect batch statistics
}

// batchStat tracks statistics for adaptive batch sizing
type batchStat struct {
	successRate      float64 // Success rate of last batch (0.0 to 1.0)
	avgDuration      int64   // Average task duration in milliseconds
	lastAdjustment   time.Time
	consecutiveFails int // Number of consecutive batches with failures
}

// NewBatchExecutor creates a new batch executor
func NewBatchExecutor(config *Config, logger log.Logger) *BatchExecutor {
	return &BatchExecutor{
		config:           config,
		logger:           logger,
		results:          make([]SyncResult, 0),
		currentBatchSize: config.BatchSize, // Initialize with configured batch size
	}
}

// NewBatchExecutorWithFactory creates a new batch executor with a client factory
func NewBatchExecutorWithFactory(config *Config, logger log.Logger, factory *client.Factory) *BatchExecutor {
	return &BatchExecutor{
		config:           config,
		logger:           logger,
		results:          make([]SyncResult, 0),
		factory:          factory,
		clientCache:      make(map[string]service.RegistryClient),
		currentBatchSize: config.BatchSize, // Initialize with configured batch size
	}
}

// Execute executes sync tasks in parallel batches
func (be *BatchExecutor) Execute(ctx context.Context, tasks []SyncTask) ([]SyncResult, error) {
	if len(tasks) == 0 {
		return []SyncResult{}, nil
	}

	be.logger.WithFields(map[string]interface{}{
		"total_tasks": len(tasks),
		"batch_size":  be.config.BatchSize,
		"parallelism": be.config.Parallel,
		"adaptive":    be.config.EnableAdaptiveBatching,
	}).Info("Starting batch execution")

	// Initialize results
	be.results = make([]SyncResult, len(tasks))

	// Adjust batch size based on previous performance (if adaptive batching enabled)
	be.adjustBatchSize()

	// Create batches
	batches := be.createBatches(tasks)

	be.logger.WithFields(map[string]interface{}{
		"num_batches": len(batches),
	}).Info("Created task batches")

	// Execute batches in parallel
	var wg sync.WaitGroup
	sem := make(chan struct{}, be.config.Parallel)
	errChan := make(chan error, len(batches))

	// Track cumulative start index for adaptive batch sizes
	cumulativeIdx := 0
	for batchIdx, batch := range batches {
		wg.Add(1)
		startIdx := cumulativeIdx
		cumulativeIdx += len(batch)

		go func(idx int, start int, b []SyncTask) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Execute batch with actual start index
			if err := be.executeBatch(ctx, idx, start, b); err != nil {
				be.logger.WithFields(map[string]interface{}{
					"batch": idx,
				}).Error("Batch execution failed", err)
				errChan <- err
			}
		}(batchIdx, startIdx, batch)
	}

	// Wait for all batches
	wg.Wait()
	close(errChan)

	// Check for errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 && !be.config.ContinueOnError {
		return be.results, fmt.Errorf("batch execution failed: %d errors", len(errs))
	}

	return be.results, nil
}

// createBatches groups tasks into batches using adaptive or fixed batch size
func (be *BatchExecutor) createBatches(tasks []SyncTask) [][]SyncTask {
	// Use adaptive batch size if enabled, otherwise use configured size
	batchSize := be.currentBatchSize
	if batchSize <= 0 {
		batchSize = be.config.BatchSize
		if batchSize <= 0 {
			batchSize = 10
		}
	}

	numBatches := (len(tasks) + batchSize - 1) / batchSize
	batches := make([][]SyncTask, numBatches)

	for i := 0; i < len(tasks); i += batchSize {
		end := i + batchSize
		if end > len(tasks) {
			end = len(tasks)
		}
		batches[i/batchSize] = tasks[i:end]
	}

	return batches
}

// adjustBatchSize adjusts the current batch size based on previous batch performance
func (be *BatchExecutor) adjustBatchSize() {
	if !be.config.EnableAdaptiveBatching {
		return // Adaptive batching disabled
	}

	be.statsMu.Lock()
	defer be.statsMu.Unlock()

	stats := be.batchStats

	// Don't adjust too frequently (wait at least 5 seconds between adjustments)
	if time.Since(stats.lastAdjustment) < 5*time.Second {
		return
	}

	oldSize := be.currentBatchSize
	newSize := oldSize

	// Adjust based on success rate
	if stats.successRate < 0.5 {
		// High failure rate - reduce batch size aggressively
		newSize = oldSize / 2
		stats.consecutiveFails++
	} else if stats.successRate < 0.8 {
		// Moderate failure rate - reduce batch size slightly
		newSize = oldSize - (oldSize / 4)
		stats.consecutiveFails++
	} else if stats.successRate >= 0.95 && stats.consecutiveFails == 0 {
		// High success rate and no recent failures - increase batch size
		newSize = oldSize + (oldSize / 4)
	} else if stats.successRate >= 0.9 {
		// Good success rate - reset consecutive failures
		stats.consecutiveFails = 0
	}

	// Enforce bounds
	if newSize < be.config.MinBatchSize {
		newSize = be.config.MinBatchSize
	}
	if newSize > be.config.MaxBatchSize {
		newSize = be.config.MaxBatchSize
	}

	// Apply adjustment
	if newSize != oldSize {
		be.currentBatchSize = newSize
		stats.lastAdjustment = time.Now()
		be.batchStats = stats

		be.logger.WithFields(map[string]interface{}{
			"old_size":     oldSize,
			"new_size":     newSize,
			"success_rate": fmt.Sprintf("%.1f%%", stats.successRate*100),
			"avg_duration": stats.avgDuration,
		}).Info("Adjusted batch size based on performance")
	}
}

// updateBatchStats updates batch statistics after a batch completes
func (be *BatchExecutor) updateBatchStats(batchResults []SyncResult) {
	if !be.config.EnableAdaptiveBatching {
		return
	}

	be.statsMu.Lock()
	defer be.statsMu.Unlock()

	if len(batchResults) == 0 {
		return
	}

	// Calculate success rate
	successes := 0
	var totalDuration int64
	for _, result := range batchResults {
		if result.Success {
			successes++
		}
		totalDuration += result.Duration
	}

	successRate := float64(successes) / float64(len(batchResults))
	avgDuration := totalDuration / int64(len(batchResults))

	// Update statistics
	be.batchStats.successRate = successRate
	be.batchStats.avgDuration = avgDuration

	be.logger.WithFields(map[string]interface{}{
		"success_rate": fmt.Sprintf("%.1f%%", successRate*100),
		"successes":    successes,
		"failures":     len(batchResults) - successes,
		"avg_duration": avgDuration,
	}).Debug("Updated batch statistics")
}

// executeBatch executes a single batch of tasks
func (be *BatchExecutor) executeBatch(ctx context.Context, batchIdx int, startIdx int, tasks []SyncTask) error {
	be.logger.WithFields(map[string]interface{}{
		"batch": batchIdx,
		"size":  len(tasks),
	}).Info("Executing batch")

	// Execute tasks concurrently within batch
	var wg sync.WaitGroup
	taskChan := make(chan struct {
		idx  int
		task SyncTask
	}, len(tasks))

	// Use passed startIdx (already calculated correctly for adaptive batching)
	// Start workers
	for i := 0; i < len(tasks); i++ {
		taskChan <- struct {
			idx  int
			task SyncTask
		}{idx: startIdx + i, task: tasks[i]}
	}
	close(taskChan)

	// Process tasks
	for taskInfo := range taskChan {
		wg.Add(1)
		go func(ti struct {
			idx  int
			task SyncTask
		}) {
			defer wg.Done()

			result := be.executeTask(ctx, ti.task)

			// Store result
			be.mu.Lock()
			be.results[ti.idx] = result
			be.mu.Unlock()
		}(taskInfo)
	}

	wg.Wait()

	// Collect results for this batch to update statistics
	batchResults := make([]SyncResult, len(tasks))
	be.mu.Lock()
	for i := 0; i < len(tasks); i++ {
		batchResults[i] = be.results[startIdx+i]
	}
	be.mu.Unlock()

	// Update batch statistics for adaptive sizing
	be.updateBatchStats(batchResults)

	be.logger.WithFields(map[string]interface{}{
		"batch": batchIdx,
	}).Info("Batch execution completed")

	return nil
}

// executeTask executes a single sync task with retries and timeout enforcement
func (be *BatchExecutor) executeTask(ctx context.Context, task SyncTask) SyncResult {
	startTime := time.Now()
	var lastErr error

	srcRef := fmt.Sprintf("%s/%s:%s", task.SourceRegistry, task.SourceRepository, task.SourceTag)
	dstRef := fmt.Sprintf("%s/%s:%s", task.DestRegistry, task.DestRepository, task.DestTag)

	be.logger.WithFields(map[string]interface{}{
		"source": srcRef,
		"dest":   dstRef,
	}).Debug("Starting sync task")

	// Create timeout context for the entire task (all retry attempts)
	// Use configured timeout, default 5 minutes
	timeout := time.Duration(be.config.Timeout) * time.Second
	taskCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Check for cancellation before starting
	select {
	case <-taskCtx.Done():
		return SyncResult{
			Task:     task,
			Success:  false,
			Error:    fmt.Errorf("task cancelled before execution: %w", taskCtx.Err()),
			Duration: time.Since(startTime).Milliseconds(),
		}
	default:
	}

	// Retry loop
	for attempt := 0; attempt <= be.config.RetryAttempts; attempt++ {
		// Check for context cancellation/timeout at start of each attempt
		select {
		case <-taskCtx.Done():
			duration := time.Since(startTime).Milliseconds()
			be.logger.WithFields(map[string]interface{}{
				"source":   srcRef,
				"dest":     dstRef,
				"attempts": attempt,
				"elapsed":  duration,
			}).Warn("Sync task cancelled or timed out")

			return SyncResult{
				Task:     task,
				Success:  false,
				Error:    fmt.Errorf("sync cancelled/timed out after %d attempts: %w", attempt, taskCtx.Err()),
				Duration: duration,
				Retries:  attempt,
			}
		default:
		}
		if attempt > 0 {
			// Exponential backoff with context-aware sleep
			backoff := time.Duration(be.config.RetryBackoff*(1<<(attempt-1))) * time.Second
			be.logger.WithFields(map[string]interface{}{
				"attempt": attempt,
				"backoff": backoff,
			}).Debug("Retrying sync task")

			// Use context-aware sleep to allow immediate cancellation
			timer := time.NewTimer(backoff)
			select {
			case <-taskCtx.Done():
				timer.Stop()
				duration := time.Since(startTime).Milliseconds()
				return SyncResult{
					Task:     task,
					Success:  false,
					Error:    fmt.Errorf("sync cancelled during retry backoff: %w", taskCtx.Err()),
					Duration: duration,
					Retries:  attempt,
				}
			case <-timer.C:
				// Backoff complete, continue to next attempt
			}
		}

		// Execute sync with timeout context
		bytesCopied, err := be.syncImage(taskCtx, task)
		if err == nil {
			duration := time.Since(startTime).Milliseconds()
			be.logger.WithFields(map[string]interface{}{
				"source":       srcRef,
				"dest":         dstRef,
				"bytes_copied": bytesCopied,
				"duration_ms":  duration,
				"attempts":     attempt + 1,
			}).Info("Sync task completed successfully")

			return SyncResult{
				Task:        task,
				Success:     true,
				BytesCopied: bytesCopied,
				Duration:    duration,
				Retries:     attempt,
			}
		}

		lastErr = err
		be.logger.WithFields(map[string]interface{}{
			"source":  srcRef,
			"dest":    dstRef,
			"attempt": attempt + 1,
			"error":   err.Error(),
		}).Warn("Sync task failed")
	}

	// All retries failed
	duration := time.Since(startTime).Milliseconds()
	return SyncResult{
		Task:     task,
		Success:  false,
		Error:    lastErr,
		Duration: duration,
		Retries:  be.config.RetryAttempts,
	}
}

// getOrCreateClient gets a cached client or creates a new one for the registry
func (be *BatchExecutor) getOrCreateClient(ctx context.Context, registryURL string) (service.RegistryClient, error) {
	// Check if factory is initialized
	if be.factory == nil {
		return nil, fmt.Errorf("client factory not initialized")
	}

	// Try to get from cache first (read lock)
	be.cacheMu.RLock()
	if client, exists := be.clientCache[registryURL]; exists {
		be.cacheMu.RUnlock()
		return client, nil
	}
	be.cacheMu.RUnlock()

	// Not in cache, create new client (write lock)
	be.cacheMu.Lock()
	defer be.cacheMu.Unlock()

	// Double-check after acquiring write lock (another goroutine might have created it)
	if client, exists := be.clientCache[registryURL]; exists {
		return client, nil
	}

	// Create new client
	client, err := be.factory.CreateClientForRegistry(ctx, registryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for registry %s: %w", registryURL, err)
	}

	// Store in cache
	be.clientCache[registryURL] = client

	be.logger.WithFields(map[string]interface{}{
		"registry": registryURL,
	}).Debug("Created and cached new registry client")

	return client, nil
}

// syncImage performs the actual image synchronization using freightliner's copy infrastructure
func (be *BatchExecutor) syncImage(ctx context.Context, task SyncTask) (int64, error) {
	// Create source registry reference
	srcImageRef := fmt.Sprintf("%s/%s:%s", task.SourceRegistry, task.SourceRepository, task.SourceTag)

	// Create destination registry reference
	dstImageRef := fmt.Sprintf("%s/%s:%s", task.DestRegistry, task.DestRepository, task.DestTag)

	be.logger.WithFields(map[string]interface{}{
		"source": srcImageRef,
		"dest":   dstImageRef,
	}).Debug("Starting image synchronization")

	// Parse source reference using go-containerregistry
	sourceRef, err := name.ParseReference(srcImageRef)
	if err != nil {
		return 0, fmt.Errorf("failed to parse source reference: %w", err)
	}

	// Parse destination reference
	destRef, err := name.ParseReference(dstImageRef)
	if err != nil {
		return 0, fmt.Errorf("failed to parse destination reference: %w", err)
	}

	// Get or create source registry client (with caching)
	srcClient, err := be.getOrCreateClient(ctx, task.SourceRegistry)
	if err != nil {
		return 0, fmt.Errorf("failed to get source registry client: %w", err)
	}

	// Get or create destination registry client (with caching)
	destClient, err := be.getOrCreateClient(ctx, task.DestRegistry)
	if err != nil {
		return 0, fmt.Errorf("failed to get destination registry client: %w", err)
	}

	// Get source repository
	sourceRepo, err := srcClient.GetRepository(ctx, task.SourceRepository)
	if err != nil {
		return 0, fmt.Errorf("failed to get source repository: %w", err)
	}

	// Get destination repository
	destRepo, err := destClient.GetRepository(ctx, task.DestRepository)
	if err != nil {
		return 0, fmt.Errorf("failed to get destination repository: %w", err)
	}

	// Get remote options for authentication
	srcOpts, err := sourceRepo.GetRemoteOptions()
	if err != nil {
		return 0, fmt.Errorf("failed to get source remote options: %w", err)
	}

	destOpts, err := destRepo.GetRemoteOptions()
	if err != nil {
		return 0, fmt.Errorf("failed to get destination remote options: %w", err)
	}

	// Create copier instance
	copier := copyutil.NewCopier(be.logger)

	// Prepare copy options
	copyOptions := copyutil.CopyOptions{
		DryRun:         false, // Always perform actual copy for sync operations
		ForceOverwrite: true,  // Sync should overwrite by default
		Source:         sourceRef,
		Destination:    destRef,
	}

	// Execute the image copy operation
	result, err := copier.CopyImage(ctx, sourceRef, destRef, srcOpts, destOpts, copyOptions)
	if err != nil {
		return 0, fmt.Errorf("failed to copy image: %w", err)
	}

	if !result.Success {
		return 0, fmt.Errorf("image copy reported failure")
	}

	be.logger.WithFields(map[string]interface{}{
		"source":            srcImageRef,
		"dest":              dstImageRef,
		"bytes_transferred": result.Stats.BytesTransferred,
		"layers":            result.Stats.Layers,
		"pull_duration_ms":  result.Stats.PullDuration.Milliseconds(),
		"push_duration_ms":  result.Stats.PushDuration.Milliseconds(),
	}).Debug("Image synchronization completed")

	// Return the number of bytes transferred
	return result.Stats.BytesTransferred, nil
}

// OptimizeBatches optimizes batch ordering for efficiency
// Groups tasks by:
// - Same source registry (reduce connection overhead)
// - Image size (process smaller images first)
// - Priority (high priority first)
func OptimizeBatches(tasks []SyncTask) []SyncTask {
	// Create optimized ordering
	type taskWithMetrics struct {
		task     SyncTask
		priority int
		size     int64
	}

	// For now, simple priority-based sorting
	// TODO: Add actual size estimation from manifests
	optimized := make([]SyncTask, len(tasks))
	copy(optimized, tasks)

	// Sort by priority (descending), then by registry (grouping)
	// This reduces connection overhead and processes critical images first
	type sortKey struct {
		priority int
		registry string
	}

	keys := make(map[int]sortKey)
	for i, task := range optimized {
		keys[i] = sortKey{
			priority: task.Priority,
			registry: task.SourceRegistry,
		}
	}

	// Stable sort using O(n log n) algorithm to preserve relative order within same priority/registry
	sort.SliceStable(optimized, func(i, j int) bool {
		ki := keys[i]
		kj := keys[j]

		// Higher priority first
		if ki.priority != kj.priority {
			return ki.priority > kj.priority
		}

		// Same priority, group by registry alphabetically
		return ki.registry < kj.registry
	})

	return optimized
}

// EstimateDuration estimates batch execution duration
func EstimateDuration(tasks []SyncTask, parallelism int, batchSize int) time.Duration {
	if len(tasks) == 0 {
		return 0
	}

	// Average sync time per image (rough estimate)
	avgSyncTime := 30 * time.Second

	// Calculate total batches
	numBatches := (len(tasks) + batchSize - 1) / batchSize

	// Calculate parallel execution time
	batchesPerRound := parallelism
	numRounds := (numBatches + batchesPerRound - 1) / batchesPerRound

	// Estimate: rounds * batch_size * avg_sync_time
	return time.Duration(numRounds*batchSize) * avgSyncTime
}

// ReplicationOptions contains options for native replication integration
type ReplicationOptions struct {
	// SourceClient is the source registry client
	SourceClient service.RegistryClient

	// DestClient is the destination registry client
	DestClient service.RegistryClient

	// EnableDeduplication enables CAS-based deduplication
	EnableDeduplication bool

	// EnableHTTP3 enables HTTP/3 with QUIC protocol
	EnableHTTP3 bool

	// CheckpointInterval for resumable transfers
	CheckpointInterval time.Duration

	// VerifySignatures verifies image signatures
	VerifySignatures bool

	// WorkerCount for parallel layer transfers
	WorkerCount int
}

// BatchStatistics tracks batch execution statistics
type BatchStatistics struct {
	TotalTasks      int
	CompletedTasks  int
	FailedTasks     int
	SkippedTasks    int
	TotalBytes      int64
	TotalDuration   time.Duration
	AverageDuration time.Duration
	SuccessRate     float64
	ThroughputMBps  float64
}

// CalculateStatistics calculates execution statistics
func CalculateStatistics(results []SyncResult) BatchStatistics {
	stats := BatchStatistics{
		TotalTasks: len(results),
	}

	for _, result := range results {
		if result.Success {
			stats.CompletedTasks++
			stats.TotalBytes += result.BytesCopied
		} else if result.Skipped {
			stats.SkippedTasks++
		} else {
			stats.FailedTasks++
		}
		stats.TotalDuration += time.Duration(result.Duration) * time.Millisecond
	}

	if stats.TotalTasks > 0 {
		stats.AverageDuration = stats.TotalDuration / time.Duration(stats.TotalTasks)
		stats.SuccessRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	if stats.TotalDuration > 0 {
		totalSeconds := stats.TotalDuration.Seconds()
		totalMB := float64(stats.TotalBytes) / (1024 * 1024)
		stats.ThroughputMBps = totalMB / totalSeconds
	}

	return stats
}
