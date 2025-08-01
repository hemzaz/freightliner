package load

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// runHighVolumeReplication executes the high-volume replication scenario
func (sr *ScenarioRunner) runHighVolumeReplication() (*LoadTestResults, error) {
	sr.logger.Info("Running high-volume replication scenario", map[string]interface{}{
		"total_repositories": len(sr.config.Images),
		"target_completion":  "90% within 2 hours vs current 6-8 hours",
		"concurrent_workers": sr.config.ConcurrentWorkers,
	})

	// Track progress for completion rate calculation
	var processedCount atomic.Int64
	var failureCount atomic.Int64
	var totalThroughput atomic.Int64 // KB/s
	var peakThroughput atomic.Int64  // KB/s

	// Create worker semaphore to limit concurrent operations
	workerSem := make(chan struct{}, sr.config.ConcurrentWorkers)

	// Process repositories in parallel with controlled concurrency
	var wg sync.WaitGroup
	startTime := time.Now()

	// Metrics collection goroutine
	go sr.collectThroughputMetrics(&totalThroughput, &peakThroughput)

	for i, image := range sr.config.Images {
		select {
		case <-sr.ctx.Done():
			sr.logger.Info("Scenario cancelled, stopping workers")
			break
		default:
		}

		wg.Add(1)
		go func(idx int, img ContainerImage) {
			defer wg.Done()

			// Acquire worker slot
			workerSem <- struct{}{}
			defer func() { <-workerSem }()

			// Simulate container replication with realistic timing
			processingTime := sr.calculateProcessingTime(img)
			throughputKBps := sr.simulateReplication(img, processingTime)

			// Update metrics
			processedCount.Add(1)
			totalThroughput.Add(throughputKBps)

			// Check for peak throughput
			for {
				current := peakThroughput.Load()
				if throughputKBps <= current || peakThroughput.CompareAndSwap(current, throughputKBps) {
					break
				}
			}

			// Simulate failure rate based on network conditions
			if sr.shouldSimulateFailure() {
				failureCount.Add(1)
				sr.logger.Debug("Simulated replication failure", map[string]interface{}{
					"repository": img.Repository,
					"tag":        img.Tag,
					"size_mb":    img.SizeMB,
				})
			}

			if idx%100 == 0 {
				elapsed := time.Since(startTime)
				processed := processedCount.Load()
				completionRate := float64(processed) / float64(len(sr.config.Images)) * 100

				sr.logger.Info("High-volume replication progress", map[string]interface{}{
					"processed":       processed,
					"total":           len(sr.config.Images),
					"completion_rate": fmt.Sprintf("%.1f%%", completionRate),
					"elapsed":         elapsed.String(),
					"avg_throughput": fmt.Sprintf("%.1f MB/s",
						float64(totalThroughput.Load())/float64(processed)/1000),
				})
			}
		}(i, image)

		// Rate limiting to prevent overwhelming the system
		if i > 0 && i%sr.config.ConcurrentWorkers == 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Wait for all workers to complete
	wg.Wait()
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// Calculate final metrics
	processed := processedCount.Load()
	failed := failureCount.Load()
	avgThroughputMBps := float64(totalThroughput.Load()) / float64(processed) / 1000
	peakThroughputMBps := float64(peakThroughput.Load()) / 1000

	results := &LoadTestResults{
		ScenarioName:          sr.config.Name,
		Duration:              duration,
		TotalImages:           int64(len(sr.config.Images)),
		ProcessedImages:       processed,
		FailedImages:          failed,
		AverageThroughputMBps: avgThroughputMBps,
		PeakThroughputMBps:    peakThroughputMBps,
		MemoryUsageMB:         sr.memoryUsageMB.Load(),
		ConnectionReuseRate:   sr.connectionStats.GetConnectionReuseRate(),
		FailureRate:           float64(failed) / float64(processed),
	}

	// Validate against criteria
	results.ValidationPassed, results.ValidationErrors = sr.validateResults(results)

	sr.logger.Info("High-volume replication completed", map[string]interface{}{
		"duration":          duration.String(),
		"processed":         processed,
		"failed":            failed,
		"completion_rate":   fmt.Sprintf("%.1f%%", float64(processed)/float64(len(sr.config.Images))*100),
		"avg_throughput":    fmt.Sprintf("%.1f MB/s", avgThroughputMBps),
		"peak_throughput":   fmt.Sprintf("%.1f MB/s", peakThroughputMBps),
		"validation_passed": results.ValidationPassed,
	})

	return results, nil
}

// runLargeImageStress executes the large image stress testing scenario
func (sr *ScenarioRunner) runLargeImageStress() (*LoadTestResults, error) {
	sr.logger.Info("Running large image stress test", map[string]interface{}{
		"large_images":      len(sr.config.Images),
		"memory_limit_mb":   sr.config.MemoryLimitMB,
		"target_throughput": fmt.Sprintf("%.1f MB/s", sr.config.ExpectedThroughput),
	})

	var processedCount atomic.Int64
	var failureCount atomic.Int64
	var totalThroughput atomic.Int64
	var peakThroughput atomic.Int64
	var peakMemoryMB atomic.Int64

	// Lower concurrency for large images to control memory usage
	workerSem := make(chan struct{}, sr.config.ConcurrentWorkers)

	var wg sync.WaitGroup
	startTime := time.Now()

	// Enhanced memory monitoring for large images
	go sr.monitorMemoryUsage(&peakMemoryMB)
	go sr.collectThroughputMetrics(&totalThroughput, &peakThroughput)

	for i, image := range sr.config.Images {
		select {
		case <-sr.ctx.Done():
			break
		default:
		}

		wg.Add(1)
		go func(idx int, img ContainerImage) {
			defer wg.Done()

			workerSem <- struct{}{}
			defer func() { <-workerSem }()

			// Memory-conscious processing for large images
			beforeMemory := sr.getCurrentMemoryMB()

			// Simulate streaming buffer management (50MB chunks as per memory-profiler)
			chunkSize := int64(50) // 50MB chunks
			chunks := (img.SizeMB + chunkSize - 1) / chunkSize

			totalProcessingTime := time.Duration(0)
			totalThroughputForImage := int64(0)

			for chunk := int64(0); chunk < chunks; chunk++ {
				currentChunkSize := chunkSize
				if chunk == chunks-1 {
					currentChunkSize = img.SizeMB - (chunk * chunkSize)
				}

				// Simulate chunk processing with memory efficiency
				chunkProcessingTime := time.Duration(currentChunkSize*8) * time.Millisecond // 8ms per MB
				time.Sleep(chunkProcessingTime)
				totalProcessingTime += chunkProcessingTime

				// Calculate throughput for this chunk
				chunkThroughputKBps := int64(float64(currentChunkSize*1000) / chunkProcessingTime.Seconds())
				totalThroughputForImage += chunkThroughputKBps

				// Update peak throughput
				for {
					current := peakThroughput.Load()
					if chunkThroughputKBps <= current || peakThroughput.CompareAndSwap(current, chunkThroughputKBps) {
						break
					}
				}

				// Simulate GC and memory cleanup between chunks
				if chunk%3 == 0 {
					runtime.GC()
					time.Sleep(10 * time.Millisecond)
				}
			}

			afterMemory := sr.getCurrentMemoryMB()
			memoryDelta := afterMemory - beforeMemory

			// Update memory tracking
			for {
				current := peakMemoryMB.Load()
				if afterMemory <= current || peakMemoryMB.CompareAndSwap(current, afterMemory) {
					break
				}
			}

			processedCount.Add(1)
			totalThroughput.Add(totalThroughputForImage / chunks) // Average throughput

			// Check memory constraint violations
			if afterMemory > sr.config.MemoryLimitMB {
				failureCount.Add(1)
				sr.logger.Warn("Memory limit exceeded during large image processing", map[string]interface{}{
					"repository":   img.Repository,
					"size_mb":      img.SizeMB,
					"memory_mb":    afterMemory,
					"memory_delta": memoryDelta,
					"limit_mb":     sr.config.MemoryLimitMB,
				})
			}

			if idx%5 == 0 {
				processed := processedCount.Load()
				avgThroughput := float64(totalThroughput.Load()) / float64(processed) / 1000

				sr.logger.Info("Large image stress progress", map[string]interface{}{
					"processed":      processed,
					"total":          len(sr.config.Images),
					"avg_throughput": fmt.Sprintf("%.1f MB/s", avgThroughput),
					"peak_memory":    fmt.Sprintf("%d MB", peakMemoryMB.Load()),
					"current_memory": fmt.Sprintf("%d MB", afterMemory),
				})
			}
		}(i, image)

		// Stagger large image processing to control memory pressure
		time.Sleep(200 * time.Millisecond)
	}

	wg.Wait()
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	processed := processedCount.Load()
	failed := failureCount.Load()
	avgThroughputMBps := float64(totalThroughput.Load()) / float64(processed) / 1000
	peakThroughputMBps := float64(peakThroughput.Load()) / 1000

	results := &LoadTestResults{
		ScenarioName:          sr.config.Name,
		Duration:              duration,
		TotalImages:           int64(len(sr.config.Images)),
		ProcessedImages:       processed,
		FailedImages:          failed,
		AverageThroughputMBps: avgThroughputMBps,
		PeakThroughputMBps:    peakThroughputMBps,
		MemoryUsageMB:         peakMemoryMB.Load(),
		ConnectionReuseRate:   sr.connectionStats.GetConnectionReuseRate(),
		FailureRate:           float64(failed) / float64(processed),
		DetailedMetrics: map[string]interface{}{
			"peak_memory_mb":    peakMemoryMB.Load(),
			"memory_limit_mb":   sr.config.MemoryLimitMB,
			"memory_efficiency": fmt.Sprintf("%.1f%%", float64(sr.config.MemoryLimitMB-peakMemoryMB.Load())/float64(sr.config.MemoryLimitMB)*100),
		},
	}

	results.ValidationPassed, results.ValidationErrors = sr.validateResults(results)

	sr.logger.Info("Large image stress test completed", map[string]interface{}{
		"duration":          duration.String(),
		"avg_throughput":    fmt.Sprintf("%.1f MB/s", avgThroughputMBps),
		"peak_memory":       fmt.Sprintf("%d MB", peakMemoryMB.Load()),
		"memory_efficient":  peakMemoryMB.Load() <= sr.config.MemoryLimitMB,
		"validation_passed": results.ValidationPassed,
	})

	return results, nil
}

// runNetworkResilience executes the network resilience testing scenario
func (sr *ScenarioRunner) runNetworkResilience() (*LoadTestResults, error) {
	sr.logger.Info("Running network resilience test", map[string]interface{}{
		"packet_loss_rate":      fmt.Sprintf("%.1f%%", sr.config.NetworkConditions.PacketLossRate*100),
		"service_interruptions": len(sr.config.NetworkConditions.ServiceInterruptions),
		"target_final_failure":  "<1% despite poor conditions",
	})

	var processedCount atomic.Int64
	var failureCount atomic.Int64
	var recoveredCount atomic.Int64 // Failures recovered by retry logic
	var totalThroughput atomic.Int64
	var peakThroughput atomic.Int64

	workerSem := make(chan struct{}, sr.config.ConcurrentWorkers)

	var wg sync.WaitGroup
	startTime := time.Now()

	// Schedule service interruptions
	go sr.simulateServiceInterruptions()
	go sr.collectThroughputMetrics(&totalThroughput, &peakThroughput)

	for i, image := range sr.config.Images {
		select {
		case <-sr.ctx.Done():
			break
		default:
		}

		wg.Add(1)
		go func(idx int, img ContainerImage) {
			defer wg.Done()

			workerSem <- struct{}{}
			defer func() { <-workerSem }()

			// Simulate network resilience with retry logic
			success, recovered := sr.simulateResilientReplication(img)

			if success {
				processedCount.Add(1)

				// Calculate throughput (reduced due to network conditions)
				baseProcessingTime := sr.calculateProcessingTime(img)
				networkDelayFactor := 1.0 + sr.config.NetworkConditions.PacketLossRate*3.0 // 3x delay per 1% packet loss
				actualProcessingTime := time.Duration(float64(baseProcessingTime) * networkDelayFactor)

				throughputKBps := int64(float64(img.SizeMB*1000) / actualProcessingTime.Seconds())
				totalThroughput.Add(throughputKBps)

				// Update peak throughput
				for {
					current := peakThroughput.Load()
					if throughputKBps <= current || peakThroughput.CompareAndSwap(current, throughputKBps) {
						break
					}
				}

				if recovered {
					recoveredCount.Add(1)
				}
			} else {
				failureCount.Add(1)
			}

			if idx%25 == 0 {
				processed := processedCount.Load()
				failed := failureCount.Load()
				recovered := recoveredCount.Load()

				sr.logger.Info("Network resilience progress", map[string]interface{}{
					"processed":            processed,
					"failed":               failed,
					"recovered":            recovered,
					"recovery_rate":        fmt.Sprintf("%.1f%%", float64(recovered)/float64(processed+failed)*100),
					"current_failure_rate": fmt.Sprintf("%.2f%%", float64(failed)/float64(processed+failed)*100),
				})
			}
		}(i, image)

		// Simulate varied request patterns
		if i%10 == 0 {
			time.Sleep(50 * time.Millisecond)
		}
	}

	wg.Wait()
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	processed := processedCount.Load()
	failed := failureCount.Load()
	recovered := recoveredCount.Load()
	avgThroughputMBps := float64(totalThroughput.Load()) / float64(processed) / 1000
	peakThroughputMBps := float64(peakThroughput.Load()) / 1000

	results := &LoadTestResults{
		ScenarioName:          sr.config.Name,
		Duration:              duration,
		TotalImages:           int64(len(sr.config.Images)),
		ProcessedImages:       processed,
		FailedImages:          failed,
		AverageThroughputMBps: avgThroughputMBps,
		PeakThroughputMBps:    peakThroughputMBps,
		MemoryUsageMB:         sr.memoryUsageMB.Load(),
		ConnectionReuseRate:   sr.connectionStats.GetConnectionReuseRate(),
		FailureRate:           float64(failed) / float64(processed+failed),
		DetailedMetrics: map[string]interface{}{
			"recovered_failures":  recovered,
			"recovery_rate":       fmt.Sprintf("%.1f%%", float64(recovered)/float64(processed+failed)*100),
			"retry_effectiveness": float64(recovered) / float64(recovered+failed),
		},
	}

	results.ValidationPassed, results.ValidationErrors = sr.validateResults(results)

	sr.logger.Info("Network resilience test completed", map[string]interface{}{
		"duration":           duration.String(),
		"final_failure_rate": fmt.Sprintf("%.2f%%", results.FailureRate*100),
		"recovered_failures": recovered,
		"retry_success":      fmt.Sprintf("%.1f%%", float64(recovered)/float64(recovered+failed)*100),
		"validation_passed":  results.ValidationPassed,
	})

	return results, nil
}

// runBurstReplication executes the burst replication scenario
func (sr *ScenarioRunner) runBurstReplication() (*LoadTestResults, error) {
	sr.logger.Info("Running burst replication test", map[string]interface{}{
		"burst_workers":     sr.config.ConcurrentWorkers,
		"images_per_burst":  len(sr.config.Images) / 5, // 5 bursts
		"target_throughput": fmt.Sprintf("%.1f MB/s", sr.config.ExpectedThroughput),
	})

	var processedCount atomic.Int64
	var failureCount atomic.Int64
	var totalThroughput atomic.Int64
	var peakThroughput atomic.Int64

	// Higher concurrency for burst scenarios
	workerSem := make(chan struct{}, sr.config.ConcurrentWorkers)

	var wg sync.WaitGroup
	startTime := time.Now()

	go sr.collectThroughputMetrics(&totalThroughput, &peakThroughput)

	// Create 5 bursts of work
	batchSize := len(sr.config.Images) / 5
	for batch := 0; batch < 5; batch++ {
		startIdx := batch * batchSize
		endIdx := startIdx + batchSize
		if batch == 4 { // Last batch gets remaining images
			endIdx = len(sr.config.Images)
		}

		sr.logger.Info("Starting burst", map[string]interface{}{
			"burst_number": batch + 1,
			"batch_size":   endIdx - startIdx,
			"start_time":   time.Since(startTime).String(),
		})

		// Submit entire batch at once (burst pattern)
		for i := startIdx; i < endIdx; i++ {
			select {
			case <-sr.ctx.Done():
				break
			default:
			}

			wg.Add(1)
			go func(idx int, img ContainerImage) {
				defer wg.Done()

				workerSem <- struct{}{}
				defer func() { <-workerSem }()

				processingTime := sr.calculateProcessingTime(img)
				throughputKBps := sr.simulateReplication(img, processingTime)

				processedCount.Add(1)
				totalThroughput.Add(throughputKBps)

				// Track peak throughput during bursts
				for {
					current := peakThroughput.Load()
					if throughputKBps <= current || peakThroughput.CompareAndSwap(current, throughputKBps) {
						break
					}
				}

				if sr.shouldSimulateFailure() {
					failureCount.Add(1)
				}
			}(i, sr.config.Images[i])
		}

		// Brief pause between bursts
		time.Sleep(30 * time.Second)

		processed := processedCount.Load()
		avgThroughput := float64(totalThroughput.Load()) / float64(processed) / 1000

		sr.logger.Info("Burst completed", map[string]interface{}{
			"burst_number":    batch + 1,
			"processed":       processed,
			"avg_throughput":  fmt.Sprintf("%.1f MB/s", avgThroughput),
			"peak_throughput": fmt.Sprintf("%.1f MB/s", float64(peakThroughput.Load())/1000),
		})
	}

	wg.Wait()
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	processed := processedCount.Load()
	failed := failureCount.Load()
	avgThroughputMBps := float64(totalThroughput.Load()) / float64(processed) / 1000
	peakThroughputMBps := float64(peakThroughput.Load()) / 1000

	results := &LoadTestResults{
		ScenarioName:          sr.config.Name,
		Duration:              duration,
		TotalImages:           int64(len(sr.config.Images)),
		ProcessedImages:       processed,
		FailedImages:          failed,
		AverageThroughputMBps: avgThroughputMBps,
		PeakThroughputMBps:    peakThroughputMBps,
		MemoryUsageMB:         sr.memoryUsageMB.Load(),
		ConnectionReuseRate:   sr.connectionStats.GetConnectionReuseRate(),
		FailureRate:           float64(failed) / float64(processed),
	}

	results.ValidationPassed, results.ValidationErrors = sr.validateResults(results)

	sr.logger.Info("Burst replication completed", map[string]interface{}{
		"duration":          duration.String(),
		"avg_throughput":    fmt.Sprintf("%.1f MB/s", avgThroughputMBps),
		"peak_throughput":   fmt.Sprintf("%.1f MB/s", peakThroughputMBps),
		"burst_efficiency":  fmt.Sprintf("%.1f%%", peakThroughputMBps/avgThroughputMBps*100),
		"validation_passed": results.ValidationPassed,
	})

	return results, nil
}

// runSustainedThroughput executes the sustained throughput scenario
func (sr *ScenarioRunner) runSustainedThroughput() (*LoadTestResults, error) {
	sr.logger.Info("Running sustained throughput test", map[string]interface{}{
		"duration":          sr.config.Duration.String(),
		"target_throughput": fmt.Sprintf("%.1f MB/s sustained", sr.config.ExpectedThroughput),
		"total_images":      len(sr.config.Images),
	})

	var processedCount atomic.Int64
	var failureCount atomic.Int64
	var totalThroughput atomic.Int64
	var peakThroughput atomic.Int64
	var throughputSamples []float64
	var samplesMutex sync.Mutex

	workerSem := make(chan struct{}, sr.config.ConcurrentWorkers)

	var wg sync.WaitGroup
	startTime := time.Now()

	// Enhanced throughput monitoring for sustained test
	go sr.collectSustainedThroughputMetrics(&totalThroughput, &peakThroughput, &throughputSamples, &samplesMutex)

	// Evenly distribute work over the test duration
	imageInterval := sr.config.Duration / time.Duration(len(sr.config.Images))

	for i, image := range sr.config.Images {
		select {
		case <-sr.ctx.Done():
			break
		default:
		}

		wg.Add(1)
		go func(idx int, img ContainerImage) {
			defer wg.Done()

			workerSem <- struct{}{}
			defer func() { <-workerSem }()

			processingTime := sr.calculateProcessingTime(img)
			throughputKBps := sr.simulateReplication(img, processingTime)
			throughputMBps := float64(throughputKBps) / 1000

			processedCount.Add(1)
			totalThroughput.Add(throughputKBps)

			// Add to sustained throughput samples
			samplesMutex.Lock()
			throughputSamples = append(throughputSamples, throughputMBps)
			samplesMutex.Unlock()

			// Update peak throughput
			for {
				current := peakThroughput.Load()
				if throughputKBps <= current || peakThroughput.CompareAndSwap(current, throughputKBps) {
					break
				}
			}

			if sr.shouldSimulateFailure() {
				failureCount.Add(1)
			}
		}(i, image)

		// Rate limiting to maintain sustained rather than burst pattern
		time.Sleep(imageInterval)
	}

	wg.Wait()
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	processed := processedCount.Load()
	failed := failureCount.Load()
	avgThroughputMBps := float64(totalThroughput.Load()) / float64(processed) / 1000
	peakThroughputMBps := float64(peakThroughput.Load()) / 1000

	// Calculate sustained throughput statistics
	samplesMutex.Lock()
	sustainedStats := sr.calculateSustainedStats(throughputSamples)
	samplesMutex.Unlock()

	results := &LoadTestResults{
		ScenarioName:          sr.config.Name,
		Duration:              duration,
		TotalImages:           int64(len(sr.config.Images)),
		ProcessedImages:       processed,
		FailedImages:          failed,
		AverageThroughputMBps: avgThroughputMBps,
		PeakThroughputMBps:    peakThroughputMBps,
		MemoryUsageMB:         sr.memoryUsageMB.Load(),
		ConnectionReuseRate:   sr.connectionStats.GetConnectionReuseRate(),
		FailureRate:           float64(failed) / float64(processed),
		DetailedMetrics:       sustainedStats,
	}

	results.ValidationPassed, results.ValidationErrors = sr.validateResults(results)

	sr.logger.Info("Sustained throughput test completed", map[string]interface{}{
		"duration":            duration.String(),
		"avg_throughput":      fmt.Sprintf("%.1f MB/s", avgThroughputMBps),
		"sustained_min":       fmt.Sprintf("%.1f MB/s", sustainedStats["min_throughput_mbps"]),
		"sustained_max":       fmt.Sprintf("%.1f MB/s", sustainedStats["max_throughput_mbps"]),
		"throughput_variance": fmt.Sprintf("%.1f%%", sustainedStats["throughput_variance_percent"]),
		"validation_passed":   results.ValidationPassed,
	})

	return results, nil
}

// runMixedContainerSizes executes the mixed container sizes scenario
func (sr *ScenarioRunner) runMixedContainerSizes() (*LoadTestResults, error) {
	sr.logger.Info("Running mixed container sizes test", map[string]interface{}{
		"total_images":      len(sr.config.Images),
		"size_distribution": "40% small, 35% medium, 20% large, 5% extra-large",
		"target_throughput": fmt.Sprintf("%.1f MB/s", sr.config.ExpectedThroughput),
	})

	var processedCount atomic.Int64
	var failureCount atomic.Int64
	var totalThroughput atomic.Int64
	var peakThroughput atomic.Int64

	// Track performance by size category
	sizeCategoryStats := make(map[string]*CategoryStats)
	sizeCategoryStats["small"] = &CategoryStats{}
	sizeCategoryStats["medium"] = &CategoryStats{}
	sizeCategoryStats["large"] = &CategoryStats{}
	sizeCategoryStats["extra-large"] = &CategoryStats{}
	var statsMutex sync.Mutex

	workerSem := make(chan struct{}, sr.config.ConcurrentWorkers)

	var wg sync.WaitGroup
	startTime := time.Now()

	go sr.collectThroughputMetrics(&totalThroughput, &peakThroughput)

	for i, image := range sr.config.Images {
		select {
		case <-sr.ctx.Done():
			break
		default:
		}

		wg.Add(1)
		go func(idx int, img ContainerImage) {
			defer wg.Done()

			workerSem <- struct{}{}
			defer func() { <-workerSem }()

			category := sr.getSizeCategory(img.SizeMB)
			processingTime := sr.calculateProcessingTime(img)
			throughputKBps := sr.simulateReplication(img, processingTime)
			throughputMBps := float64(throughputKBps) / 1000

			// Update category statistics
			statsMutex.Lock()
			stats := sizeCategoryStats[category]
			stats.Count++
			stats.TotalThroughput += throughputMBps
			stats.TotalSize += img.SizeMB
			if throughputMBps > stats.MaxThroughput {
				stats.MaxThroughput = throughputMBps
			}
			if stats.MinThroughput == 0 || throughputMBps < stats.MinThroughput {
				stats.MinThroughput = throughputMBps
			}
			statsMutex.Unlock()

			processedCount.Add(1)
			totalThroughput.Add(throughputKBps)

			// Update peak throughput
			for {
				current := peakThroughput.Load()
				if throughputKBps <= current || peakThroughput.CompareAndSwap(current, throughputKBps) {
					break
				}
			}

			if sr.shouldSimulateFailure() {
				failureCount.Add(1)
			}
		}(i, image)

		// Varied pacing based on image size
		if image.SizeMB > 2000 { // Large images
			time.Sleep(100 * time.Millisecond)
		} else {
			time.Sleep(20 * time.Millisecond)
		}
	}

	wg.Wait()
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	processed := processedCount.Load()
	failed := failureCount.Load()
	avgThroughputMBps := float64(totalThroughput.Load()) / float64(processed) / 1000
	peakThroughputMBps := float64(peakThroughput.Load()) / 1000

	// Compile category statistics
	statsMutex.Lock()
	categoryMetrics := make(map[string]interface{})
	for category, stats := range sizeCategoryStats {
		if stats.Count > 0 {
			categoryMetrics[category] = map[string]interface{}{
				"count":          stats.Count,
				"avg_throughput": stats.TotalThroughput / float64(stats.Count),
				"max_throughput": stats.MaxThroughput,
				"min_throughput": stats.MinThroughput,
				"total_size_gb":  float64(stats.TotalSize) / 1024,
			}
		}
	}
	statsMutex.Unlock()

	results := &LoadTestResults{
		ScenarioName:          sr.config.Name,
		Duration:              duration,
		TotalImages:           int64(len(sr.config.Images)),
		ProcessedImages:       processed,
		FailedImages:          failed,
		AverageThroughputMBps: avgThroughputMBps,
		PeakThroughputMBps:    peakThroughputMBps,
		MemoryUsageMB:         sr.memoryUsageMB.Load(),
		ConnectionReuseRate:   sr.connectionStats.GetConnectionReuseRate(),
		FailureRate:           float64(failed) / float64(processed),
		DetailedMetrics:       categoryMetrics,
	}

	results.ValidationPassed, results.ValidationErrors = sr.validateResults(results)

	sr.logger.Info("Mixed container sizes test completed", map[string]interface{}{
		"duration":           duration.String(),
		"avg_throughput":     fmt.Sprintf("%.1f MB/s", avgThroughputMBps),
		"category_breakdown": categoryMetrics,
		"validation_passed":  results.ValidationPassed,
	})

	return results, nil
}

// CategoryStats tracks statistics for container size categories
type CategoryStats struct {
	Count           int64
	TotalThroughput float64
	MaxThroughput   float64
	MinThroughput   float64
	TotalSize       int64
}

// Helper methods for scenario implementation...

// calculateProcessingTime estimates processing time based on image characteristics
func (sr *ScenarioRunner) calculateProcessingTime(image ContainerImage) time.Duration {
	// Base processing time: 5ms per MB + 2ms per layer
	baseTime := time.Duration(image.SizeMB*5+int64(image.LayerCount)*2) * time.Millisecond

	// Network conditions impact
	networkFactor := 1.0 + sr.config.NetworkConditions.PacketLossRate*2.0
	latencyPenalty := time.Duration(sr.config.NetworkConditions.LatencyMs) * time.Millisecond

	return time.Duration(float64(baseTime)*networkFactor) + latencyPenalty
}

// simulateReplication simulates container replication and returns throughput in KB/s
func (sr *ScenarioRunner) simulateReplication(image ContainerImage, processingTime time.Duration) int64 {
	// Simulate the actual processing time
	time.Sleep(processingTime)

	// Calculate throughput based on size and time
	throughputKBps := int64(float64(image.SizeMB*1000) / processingTime.Seconds())

	// Connection reuse tracking
	sr.connectionStats.TotalConnections.Add(1)
	if rand.Float64() < 0.80 { // 80% reuse rate target
		sr.connectionStats.ReuseConnections.Add(1)
	} else {
		sr.connectionStats.NewConnections.Add(1)
	}

	return throughputKBps
}

// shouldSimulateFailure determines if a failure should be simulated
func (sr *ScenarioRunner) shouldSimulateFailure() bool {
	return rand.Float64() < sr.config.MaxFailureRate
}

// Additional helper methods for monitoring and statistics...
func (sr *ScenarioRunner) collectThroughputMetrics(totalThroughput, peakThroughput *atomic.Int64) {
	// Implementation for throughput metrics collection
}

func (sr *ScenarioRunner) monitorMemoryUsage(peakMemoryMB *atomic.Int64) {
	// Implementation for memory monitoring
}

func (sr *ScenarioRunner) getCurrentMemoryMB() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.Alloc / 1024 / 1024)
}

func (sr *ScenarioRunner) simulateServiceInterruptions() {
	// Implementation for service interruption simulation
}

func (sr *ScenarioRunner) simulateResilientReplication(image ContainerImage) (success bool, recovered bool) {
	// Implementation for resilient replication simulation with retry logic
	maxRetries := 3
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if sr.shouldSimulateFailure() {
			if attempt == maxRetries {
				return false, attempt > 0
			}
			// Simulate retry delay
			time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
			continue
		}

		// Simulate successful replication
		processingTime := sr.calculateProcessingTime(image)
		time.Sleep(processingTime)
		return true, attempt > 0
	}
	return false, false
}

func (sr *ScenarioRunner) collectSustainedThroughputMetrics(totalThroughput, peakThroughput *atomic.Int64, samples *[]float64, mutex *sync.Mutex) {
	// Implementation for sustained throughput metrics collection
}

func (sr *ScenarioRunner) calculateSustainedStats(samples []float64) map[string]interface{} {
	if len(samples) == 0 {
		return map[string]interface{}{}
	}

	var sum, min, max float64
	min = samples[0]
	max = samples[0]

	for _, sample := range samples {
		sum += sample
		if sample < min {
			min = sample
		}
		if sample > max {
			max = sample
		}
	}

	avg := sum / float64(len(samples))
	variance := (max - min) / avg * 100

	return map[string]interface{}{
		"min_throughput_mbps":         min,
		"max_throughput_mbps":         max,
		"avg_throughput_mbps":         avg,
		"throughput_variance_percent": variance,
		"sample_count":                len(samples),
	}
}

func (sr *ScenarioRunner) getSizeCategory(sizeMB int64) string {
	if sizeMB <= 100 {
		return "small"
	} else if sizeMB <= 500 {
		return "medium"
	} else if sizeMB <= 2000 {
		return "large"
	}
	return "extra-large"
}

func (sr *ScenarioRunner) validateResults(results *LoadTestResults) (bool, []string) {
	var errors []string

	// Validate throughput
	if results.AverageThroughputMBps < sr.config.ValidationCriteria.MinThroughputMBps {
		errors = append(errors, fmt.Sprintf("Average throughput %.1f MB/s below minimum %.1f MB/s",
			results.AverageThroughputMBps, sr.config.ValidationCriteria.MinThroughputMBps))
	}

	// Validate memory usage
	if results.MemoryUsageMB > sr.config.ValidationCriteria.MaxMemoryUsageMB {
		errors = append(errors, fmt.Sprintf("Memory usage %d MB exceeds limit %d MB",
			results.MemoryUsageMB, sr.config.ValidationCriteria.MaxMemoryUsageMB))
	}

	// Validate failure rate
	if results.FailureRate > sr.config.ValidationCriteria.MaxFailureRate {
		errors = append(errors, fmt.Sprintf("Failure rate %.3f%% exceeds maximum %.3f%%",
			results.FailureRate*100, sr.config.ValidationCriteria.MaxFailureRate*100))
	}

	// Validate connection reuse
	if results.ConnectionReuseRate < sr.config.ValidationCriteria.MinConnectionReuse {
		errors = append(errors, fmt.Sprintf("Connection reuse rate %.1f%% below minimum %.1f%%",
			results.ConnectionReuseRate*100, sr.config.ValidationCriteria.MinConnectionReuse*100))
	}

	return len(errors) == 0, errors
}
