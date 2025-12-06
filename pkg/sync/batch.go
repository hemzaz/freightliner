package sync

import (
	"context"
	"fmt"
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
	config    *Config
	logger    log.Logger
	scheduler *replication.Scheduler
	results   []SyncResult
	mu        sync.Mutex
	factory   *client.Factory
}

// NewBatchExecutor creates a new batch executor
func NewBatchExecutor(config *Config, logger log.Logger) *BatchExecutor {
	return &BatchExecutor{
		config:  config,
		logger:  logger,
		results: make([]SyncResult, 0),
	}
}

// NewBatchExecutorWithFactory creates a new batch executor with a client factory
func NewBatchExecutorWithFactory(config *Config, logger log.Logger, factory *client.Factory) *BatchExecutor {
	return &BatchExecutor{
		config:  config,
		logger:  logger,
		results: make([]SyncResult, 0),
		factory: factory,
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
	}).Info("Starting batch execution")

	// Initialize results
	be.results = make([]SyncResult, len(tasks))

	// Create batches
	batches := be.createBatches(tasks)

	be.logger.WithFields(map[string]interface{}{
		"num_batches": len(batches),
	}).Info("Created task batches")

	// Execute batches in parallel
	var wg sync.WaitGroup
	sem := make(chan struct{}, be.config.Parallel)
	errChan := make(chan error, len(batches))

	for batchIdx, batch := range batches {
		wg.Add(1)
		go func(idx int, b []SyncTask) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Execute batch
			if err := be.executeBatch(ctx, idx, b); err != nil {
				be.logger.WithFields(map[string]interface{}{
					"batch": idx,
				}).Error("Batch execution failed", err)
				errChan <- err
			}
		}(batchIdx, batch)
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

// createBatches groups tasks into batches
func (be *BatchExecutor) createBatches(tasks []SyncTask) [][]SyncTask {
	batchSize := be.config.BatchSize
	if batchSize <= 0 {
		batchSize = 10
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

// executeBatch executes a single batch of tasks
func (be *BatchExecutor) executeBatch(ctx context.Context, batchIdx int, tasks []SyncTask) error {
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

	// Find indices in original results array
	startIdx := batchIdx * be.config.BatchSize

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

	be.logger.WithFields(map[string]interface{}{
		"batch": batchIdx,
	}).Info("Batch execution completed")

	return nil
}

// executeTask executes a single sync task with retries
func (be *BatchExecutor) executeTask(ctx context.Context, task SyncTask) SyncResult {
	startTime := time.Now()
	var lastErr error

	srcRef := fmt.Sprintf("%s/%s:%s", task.SourceRegistry, task.SourceRepository, task.SourceTag)
	dstRef := fmt.Sprintf("%s/%s:%s", task.DestRegistry, task.DestRepository, task.DestTag)

	be.logger.WithFields(map[string]interface{}{
		"source": srcRef,
		"dest":   dstRef,
	}).Debug("Starting sync task")

	// Retry loop
	for attempt := 0; attempt <= be.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(be.config.RetryBackoff*(1<<(attempt-1))) * time.Second
			be.logger.WithFields(map[string]interface{}{
				"attempt": attempt,
				"backoff": backoff,
			}).Debug("Retrying sync task")
			time.Sleep(backoff)
		}

		// Execute sync
		bytesCopied, err := be.syncImage(ctx, task)
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

	// Create or get source registry client
	var srcClient service.RegistryClient
	if be.factory != nil {
		srcClient, err = be.factory.CreateClientForRegistry(ctx, task.SourceRegistry)
		if err != nil {
			return 0, fmt.Errorf("failed to create source registry client: %w", err)
		}
	} else {
		return 0, fmt.Errorf("client factory not initialized")
	}

	// Create or get destination registry client
	var destClient service.RegistryClient
	if be.factory != nil {
		destClient, err = be.factory.CreateClientForRegistry(ctx, task.DestRegistry)
		if err != nil {
			return 0, fmt.Errorf("failed to create destination registry client: %w", err)
		}
	} else {
		return 0, fmt.Errorf("client factory not initialized")
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

	// Stable sort to preserve relative order within same priority/registry
	for i := 0; i < len(optimized)-1; i++ {
		for j := i + 1; j < len(optimized); j++ {
			ki := keys[i]
			kj := keys[j]

			// Higher priority first
			if kj.priority > ki.priority {
				optimized[i], optimized[j] = optimized[j], optimized[i]
				keys[i], keys[j] = keys[j], keys[i]
			} else if kj.priority == ki.priority && kj.registry < ki.registry {
				// Same priority, group by registry alphabetically
				optimized[i], optimized[j] = optimized[j], optimized[i]
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

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
