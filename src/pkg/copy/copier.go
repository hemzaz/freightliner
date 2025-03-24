package copy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/hemzaz/freightliner/src/internal/log"
	"github.com/hemzaz/freightliner/src/internal/util"
	"github.com/hemzaz/freightliner/src/pkg/client/common"
	"github.com/hemzaz/freightliner/src/pkg/metrics"
)

// Copier handles copying images between registries
type Copier struct {
	logger     *log.Logger
	workerPool int
	metrics    metrics.Metrics
}

// CopyOptions contains options for copying images
type CopyOptions struct {
	// SourceTag is the tag to copy from
	SourceTag string

	// DestinationTag is the tag to copy to (if empty, uses SourceTag)
	DestinationTag string

	// ForceOverwrite forces overwriting existing tags
	ForceOverwrite bool
}

// NewCopier creates a new image copier
func NewCopier(logger *log.Logger) *Copier {
	return &Copier{
		logger:     logger,
		workerPool: 4, // Default to 4 concurrent workers for layer operations
		metrics:    &metrics.NoopMetrics{}, // Default to no-op metrics
	}
}

// WithMetrics sets the metrics collector
func (c *Copier) WithMetrics(m metrics.Metrics) *Copier {
	c.metrics = m
	return c
}

// WithWorkerPool sets the number of workers for parallel operations
func (c *Copier) WithWorkerPool(workers int) *Copier {
	if workers > 0 {
		c.workerPool = workers
	}
	return c
}

// CopyImage copies an image from one repository to another
// Implementation based on patterns from skopeo and go-containerregistry
func (c *Copier) CopyImage(ctx context.Context,
	sourceRepo common.Repository,
	destRepo common.Repository,
	options CopyOptions) error {
	
	// Resolve tags
	sourceTag := options.SourceTag
	destTag := options.DestinationTag
	if destTag == "" {
		destTag = sourceTag
	}

	start := time.Now()
	c.logger.Info("Starting image copy", map[string]interface{}{
		"source_tag":      sourceTag,
		"destination_tag": destTag,
	})
	
	// Record metrics for the start of the replication
	c.metrics.ReplicationStarted(sourceTag, destTag)

	// Step 1: Get manifest from source repository
	manifest, mediaType, err := sourceRepo.GetManifest(sourceTag)
	if err != nil {
		// Record metrics for the failed replication
		c.metrics.ReplicationFailed()
		return fmt.Errorf("failed to get source manifest: %w", err)
	}

	// Check if we need to handle a manifest list (multi-arch image)
	if isManifestList(mediaType) {
		return c.copyManifestList(ctx, sourceRepo, destRepo, sourceTag, destTag, manifest, mediaType, options.ForceOverwrite)
	}

	// If not a manifest list, handle as a single image manifest
	if !options.ForceOverwrite {
		// Check if the destination already has this tag
		existingTags, err := destRepo.ListTags()
		if err != nil {
			c.logger.Warn("Failed to check destination tags", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			for _, tag := range existingTags {
				if tag == destTag {
					if !options.ForceOverwrite {
						c.logger.Info("Destination tag already exists, skipping", map[string]interface{}{
							"tag": destTag,
						})
						return nil
					}
					break
				}
			}
		}
	}

	// Step 2: Handle the manifest
	// We first need to copy all the layers referenced in the manifest
	blobs, err := c.copyLayers(ctx, sourceRepo, destRepo, manifest, mediaType)
	if err != nil {
		// Record metrics for the failed replication
		c.metrics.ReplicationFailed()
		return fmt.Errorf("failed to copy layers: %w", err)
	}

	c.logger.Debug("Copied image layers", map[string]interface{}{
		"count": len(blobs),
	})

	// Step 3: Upload the manifest to the destination
	err = destRepo.PutManifest(destTag, manifest, mediaType)
	if err != nil {
		// Record metrics for the failed replication
		c.metrics.ReplicationFailed()
		return fmt.Errorf("failed to put manifest: %w", err)
	}

	duration := time.Since(start)
	
	// Calculate total bytes copied (approximate)
	var totalBytes int64
	for _, blob := range blobs {
		// This is just for metrics - we're not tracking actual size in this implementation
		// In a real implementation, we would track the actual size of each blob
		totalBytes += 1024 * 1024 // Assume 1MB per blob for metrics
	}
	
	// Record metrics for the completion of the replication
	c.metrics.ReplicationCompleted(duration, len(blobs), totalBytes)
	
	c.logger.Info("Image copy completed", map[string]interface{}{
		"source_tag":      sourceTag,
		"destination_tag": destTag,
		"duration_ms":     duration.Milliseconds(),
		"layers":          len(blobs),
	})

	return nil
}

// isManifestList checks if the media type is a manifest list
func isManifestList(mediaType string) bool {
	return mediaType == string(types.OCIImageIndex) || mediaType == string(types.DockerManifestList)
}

// copyManifestList handles copying a manifest list (multi-arch image)
func (c *Copier) copyManifestList(
	ctx context.Context,
	sourceRepo common.Repository,
	destRepo common.Repository,
	sourceTag string,
	destTag string,
	manifestList []byte,
	mediaType string,
	forceOverwrite bool,
) error {
	start := time.Now()
	c.logger.Info("Copying manifest list (multi-arch image)", map[string]interface{}{
		"media_type": mediaType,
	})

	// Parse the manifest list
	var index struct {
		Manifests []struct {
			Digest    string `json:"digest"`
			MediaType string `json:"mediaType"`
			Platform  struct {
				Architecture string `json:"architecture"`
				OS           string `json:"os"`
				Variant      string `json:"variant,omitempty"`
			} `json:"platform"`
		} `json:"manifests"`
		SchemaVersion int `json:"schemaVersion"`
	}

	if err := json.Unmarshal(manifestList, &index); err != nil {
		return fmt.Errorf("failed to parse manifest list: %w", err)
	}

	// Copy each manifest referenced in the list
	// We'll use a waitgroup to track completion
	var wg sync.WaitGroup
	errors := make(chan error, len(index.Manifests))

	for _, manifest := range index.Manifests {
		wg.Add(1)
		
		// Create a copy of the loop variables to use in the goroutine
		digest := manifest.Digest
		platform := fmt.Sprintf("%s/%s", manifest.Platform.OS, manifest.Platform.Architecture)
		if manifest.Platform.Variant != "" {
			platform += "/" + manifest.Platform.Variant
		}
		
		go func() {
			defer wg.Done()
			
			// Each platform-specific manifest will be referenced by digest, not tag
			c.logger.Debug("Copying platform-specific manifest", map[string]interface{}{
				"digest":   digest,
				"platform": platform,
			})
			
			// We don't have a direct way to get a manifest by digest via our interface,
			// but we can create a pseudo-tag based on the digest to copy
			digestTag := "@" + digest // This is a pseudo-tag, not a real tag
			
			// The destination will receive this as part of the overall index,
			// so we just need to ensure the blobs are copied
			manifestBytes, manifestMediaType, err := sourceRepo.GetManifest(digestTag)
			if err != nil {
				errors <- fmt.Errorf("failed to get manifest for platform %s: %w", platform, err)
				return
			}
			
			// Copy the layers for this platform's manifest
			_, err = c.copyLayers(ctx, sourceRepo, destRepo, manifestBytes, manifestMediaType)
			if err != nil {
				errors <- fmt.Errorf("failed to copy layers for platform %s: %w", platform, err)
				return
			}
		}()
	}

	// Wait for all copies to complete
	wg.Wait()
	close(errors)

	// Check for errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		// Record metrics for the failed replication
		c.metrics.ReplicationFailed()
		return fmt.Errorf("failed to copy one or more platform manifests: %v", errs)
	}

	// All platform manifests copied successfully, now upload the manifest list
	err := destRepo.PutManifest(destTag, manifestList, mediaType)
	if err != nil {
		// Record metrics for the failed replication
		c.metrics.ReplicationFailed()
		return fmt.Errorf("failed to put manifest list: %w", err)
	}

	// Record metrics for successful multi-arch image copy
	duration := time.Since(start)
	
	// For multi-arch images, calculate approximate layers and bytes
	// We're counting each platform manifest as one layer for simplicity
	totalLayers := len(index.Manifests)
	// Rough estimate - 50MB per platform image
	totalBytes := int64(totalLayers) * 50 * 1024 * 1024
	
	c.metrics.ReplicationCompleted(duration, totalLayers, totalBytes)
	
	c.logger.Info("Manifest list copy completed", map[string]interface{}{
		"source_tag":      sourceTag,
		"destination_tag": destTag,
		"duration_ms":     duration.Milliseconds(),
		"platforms":       totalLayers,
	})

	return nil
}

// copyLayers copies all layers referenced in a manifest
func (c *Copier) copyLayers(
	ctx context.Context,
	sourceRepo common.Repository,
	destRepo common.Repository,
	manifest []byte,
	mediaType string,
) ([]string, error) {
	// Parse the manifest to extract layer info
	type layerInfo struct {
		MediaType string `json:"mediaType"`
		Size      int64  `json:"size"`
		Digest    string `json:"digest"`
	}

	// Parse the manifest based on its type
	var layers []layerInfo
	var configInfo layerInfo

	// First, handle different manifest schema versions
	if mediaType == string(types.DockerManifestSchema1) || 
	   mediaType == string(types.DockerManifestSchema1Signed) {
		// Schema 1 manifests have a different structure (legacy format)
		var schema1 struct {
			FSLayers []struct {
				BlobSum string `json:"blobSum"`
			} `json:"fsLayers"`
			History []struct {
				V1Compatibility string `json:"v1Compatibility"`
			} `json:"history"`
			SchemaVersion int `json:"schemaVersion"`
		}
		
		if err := json.Unmarshal(manifest, &schema1); err != nil {
			return nil, fmt.Errorf("failed to parse schema 1 manifest: %w", err)
		}
		
		// Convert to a common format for our processing
		for _, layer := range schema1.FSLayers {
			layers = append(layers, layerInfo{
				MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				Digest:    layer.BlobSum,
			})
		}
	} else {
		// Schema 2 and OCI manifests have a similar structure
		var schema2 struct {
			Config layerInfo   `json:"config"`
			Layers []layerInfo `json:"layers"`
		}
		
		if err := json.Unmarshal(manifest, &schema2); err != nil {
			return nil, fmt.Errorf("failed to parse manifest: %w", err)
		}
		
		layers = schema2.Layers
		configInfo = schema2.Config
	}

	// Add the config blob to the list of items to copy
	allBlobs := []layerInfo{}
	if configInfo.Digest != "" {
		allBlobs = append(allBlobs, configInfo)
	}
	allBlobs = append(allBlobs, layers...)

	// Copy each blob using a worker pool pattern
	sem := make(chan struct{}, c.workerPool)
	var wg sync.WaitGroup
	errors := make(chan error, len(allBlobs))
	copiedBlobs := make(chan string, len(allBlobs))

	for _, blob := range allBlobs {
		wg.Add(1)
		
		// Create a copy of the loop variables to use in the goroutine
		digest := blob.Digest
		mediaType := blob.MediaType
		
		go func() {
			defer wg.Done()
			sem <- struct{}{} // Acquire token
			defer func() { <-sem }() // Release token
			
			// Create a pseudo-tag from the digest
			digestTag := "@" + digest
			
			c.logger.Debug("Copying blob", map[string]interface{}{
				"digest":     digest,
				"media_type": mediaType,
			})
			
			// Apply retry logic for blob transfers
			err := util.RetryWithBackoff(ctx, 3, time.Second, 10*time.Second, func() error {
				// Get the blob from the source
				blobData, _, err := sourceRepo.GetManifest(digestTag)
				if err != nil {
					return fmt.Errorf("failed to get blob %s: %w", digest, err)
				}
				
				// Put the blob in the destination
				err = destRepo.PutManifest(digestTag, blobData, mediaType)
				if err != nil {
					return fmt.Errorf("failed to put blob %s: %w", digest, err)
				}
				
				return nil
			})
			
			if err != nil {
				errors <- err
				return
			}
			
			copiedBlobs <- digest
		}()
	}

	// Wait for all blob copies to complete
	wg.Wait()
	close(errors)
	close(copiedBlobs)

	// Check for errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to copy one or more blobs: %v", errs)
	}

	// Collect all the copied blob digests
	var blobs []string
	for blob := range copiedBlobs {
		blobs = append(blobs, blob)
	}

	return blobs, nil
}
