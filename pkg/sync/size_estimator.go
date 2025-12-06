package sync

import (
	"context"
	"encoding/json"
	"fmt"

	"freightliner/pkg/manifest"
)

// SizeEstimator provides image size estimation capabilities
type SizeEstimator interface {
	// GetManifest returns the manifest for a given repository and tag
	GetManifest(ctx context.Context, repository, tag string) ([]byte, string, error)
}

// EstimateImageSize estimates the size of an image by inspecting its manifest
// Returns the total size in bytes (config + all layers)
func EstimateImageSize(ctx context.Context, estimator SizeEstimator, repository, tag string) (int64, error) {
	// Get the manifest
	manifestData, mediaType, err := estimator.GetManifest(ctx, repository, tag)
	if err != nil {
		return 0, fmt.Errorf("failed to get manifest: %w", err)
	}

	// Calculate size based on manifest type
	switch mediaType {
	case "application/vnd.oci.image.index.v1+json",
		"application/vnd.docker.distribution.manifest.list.v2+json":
		// Multi-arch manifest - estimate size of all platform manifests
		return estimateMultiArchManifestSize(manifestData)

	case "application/vnd.oci.image.manifest.v1+json":
		// OCI single-arch manifest
		return estimateOCIManifestSize(manifestData)

	case "application/vnd.docker.distribution.manifest.v2+json":
		// Docker V2 single-arch manifest
		return estimateDockerV2ManifestSize(manifestData)

	case "application/vnd.docker.distribution.manifest.v1+json":
		// Docker V1 manifest (deprecated but might still encounter)
		return estimateDockerV1ManifestSize(manifestData)

	default:
		return 0, fmt.Errorf("unsupported manifest media type: %s", mediaType)
	}
}

// estimateOCIManifestSize estimates size from OCI manifest
func estimateOCIManifestSize(manifestData []byte) (int64, error) {
	var m manifest.OCIManifest
	if err := json.Unmarshal(manifestData, &m); err != nil {
		return 0, fmt.Errorf("failed to parse OCI manifest: %w", err)
	}

	// Sum config size and all layer sizes
	totalSize := m.Config.Size
	for _, layer := range m.Layers {
		totalSize += layer.Size
	}

	return totalSize, nil
}

// estimateDockerV2ManifestSize estimates size from Docker V2 manifest
func estimateDockerV2ManifestSize(manifestData []byte) (int64, error) {
	var m manifest.DockerV2Schema2Manifest
	if err := json.Unmarshal(manifestData, &m); err != nil {
		return 0, fmt.Errorf("failed to parse Docker V2 manifest: %w", err)
	}

	// Sum config size and all layer sizes
	totalSize := m.Config.Size
	for _, layer := range m.Layers {
		totalSize += layer.Size
	}

	return totalSize, nil
}

// estimateMultiArchManifestSize estimates size from multi-arch manifest
// For manifest lists, we sum all platform manifests
func estimateMultiArchManifestSize(manifestData []byte) (int64, error) {
	// Try OCI Image Index first
	var ociIndex manifest.OCIImageIndex
	if err := json.Unmarshal(manifestData, &ociIndex); err == nil {
		totalSize := int64(0)
		for _, desc := range ociIndex.Manifests {
			totalSize += desc.Size
		}
		return totalSize, nil
	}

	// Try Docker Manifest List
	var dockerList manifest.DockerManifestList
	if err := json.Unmarshal(manifestData, &dockerList); err == nil {
		totalSize := int64(0)
		for _, desc := range dockerList.Manifests {
			totalSize += desc.Size
		}
		return totalSize, nil
	}

	return 0, fmt.Errorf("failed to parse multi-arch manifest")
}

// estimateDockerV1ManifestSize estimates size from Docker V1 manifest
func estimateDockerV1ManifestSize(manifestData []byte) (int64, error) {
	// Docker V1 manifest structure (deprecated)
	type V1Layer struct {
		BlobSum string `json:"blobSum"`
	}
	type V1FSLayer struct {
		BlobSum string `json:"blobSum"`
	}
	type V1Manifest struct {
		FSLayers []V1FSLayer `json:"fsLayers"`
	}

	var m V1Manifest
	if err := json.Unmarshal(manifestData, &m); err != nil {
		return 0, fmt.Errorf("failed to parse Docker V1 manifest: %w", err)
	}

	// V1 manifests don't include size information in the manifest itself
	// We'd need to make separate HEAD requests for each layer
	// For now, return 0 to indicate we can't estimate without additional API calls
	return 0, fmt.Errorf("Docker V1 manifests don't include size information, cannot estimate")
}

// OptimizeBatchesWithSizeEstimation optimizes batch ordering with size estimation
// Groups tasks by:
// - Priority (high priority first)
// - Same source registry (reduce connection overhead)
// - Image size (process smaller images first for faster completion)
func OptimizeBatchesWithSizeEstimation(ctx context.Context, tasks []SyncTask, estimator SizeEstimator) []SyncTask {
	// Create optimized ordering with size estimation
	type taskWithMetrics struct {
		task     SyncTask
		priority int
		size     int64
		registry string
	}

	// Estimate sizes for all tasks
	tasksWithMetrics := make([]taskWithMetrics, len(tasks))
	for i, task := range tasks {
		size := int64(0)
		if estimator != nil {
			// Try to estimate size, but don't fail if it errors
			estimatedSize, err := EstimateImageSize(ctx, estimator, task.SourceRepository, task.SourceTag)
			if err == nil {
				size = estimatedSize
			}
		}

		tasksWithMetrics[i] = taskWithMetrics{
			task:     task,
			priority: task.Priority,
			size:     size,
			registry: task.SourceRegistry,
		}
	}

	// Sort by multiple criteria:
	// 1. Priority (descending - higher priority first)
	// 2. Registry (group by registry to reduce connection overhead)
	// 3. Size (ascending - smaller images first for faster completion)
	for i := 0; i < len(tasksWithMetrics)-1; i++ {
		for j := i + 1; j < len(tasksWithMetrics); j++ {
			ti := tasksWithMetrics[i]
			tj := tasksWithMetrics[j]

			shouldSwap := false

			// Priority comparison (higher priority first)
			if tj.priority > ti.priority {
				shouldSwap = true
			} else if tj.priority == ti.priority {
				// Same priority, group by registry
				if tj.registry < ti.registry {
					shouldSwap = true
				} else if tj.registry == ti.registry {
					// Same registry, sort by size (smaller first)
					if tj.size > 0 && ti.size > 0 && tj.size < ti.size {
						shouldSwap = true
					}
				}
			}

			if shouldSwap {
				tasksWithMetrics[i], tasksWithMetrics[j] = tasksWithMetrics[j], tasksWithMetrics[i]
			}
		}
	}

	// Extract optimized task list
	optimized := make([]SyncTask, len(tasksWithMetrics))
	for i, tm := range tasksWithMetrics {
		optimized[i] = tm.task
	}

	return optimized
}

// EstimateBatchSize estimates the total size of a batch of tasks
func EstimateBatchSize(ctx context.Context, tasks []SyncTask, estimator SizeEstimator) (int64, error) {
	totalSize := int64(0)

	for _, task := range tasks {
		size, err := EstimateImageSize(ctx, estimator, task.SourceRepository, task.SourceTag)
		if err != nil {
			// Continue even if we can't estimate size for some images
			continue
		}
		totalSize += size
	}

	return totalSize, nil
}

// EstimateBatchSizes estimates sizes for all tasks in a batch
// Returns a map of task index to estimated size
func EstimateBatchSizes(ctx context.Context, tasks []SyncTask, estimator SizeEstimator) map[int]int64 {
	sizes := make(map[int]int64)

	for i, task := range tasks {
		size, err := EstimateImageSize(ctx, estimator, task.SourceRepository, task.SourceTag)
		if err == nil {
			sizes[i] = size
		}
	}

	return sizes
}
