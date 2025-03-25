package copy

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/go-containerregistry/pkg/v1/types"
	"src/internal/log"
	"src/internal/util"
	"src/pkg/client/common"
	"src/pkg/metrics"
	"src/pkg/network"
	"src/pkg/security/encryption"
	"src/pkg/security/signing"
)

// Copier handles copying images between registries
type Copier struct {
	logger         *log.Logger
	workerPool     int
	metrics        metrics.Metrics
	transferMgr    *network.TransferManager
	enableOptimize bool
	signManager    *signing.Manager
	encryptManager *encryption.Manager
}

// CopyOptions contains options for copying images
type CopyOptions struct {
	// SourceTag is the tag to copy from
	SourceTag string

	// DestinationTag is the tag to copy to (if empty, uses SourceTag)
	DestinationTag string

	// ForceOverwrite forces overwriting existing tags
	ForceOverwrite bool

	// EnableNetworkOptimization enables network optimization features
	EnableNetworkOptimization bool

	// TransferOptions configures network transfer optimizations
	TransferOptions network.TransferOptions

	// SigningOptions configures image signing
	SigningOptions *signing.SignManagerOptions

	// EncryptionOptions configures image encryption
	EncryptionOptions *encryption.EncryptionConfig

	// VerifySignatures determines if signatures should be verified during copy
	VerifySignatures bool

	// UseCustomerManagedKeys enables use of customer-managed keys for encryption
	UseCustomerManagedKeys bool
}

// NewCopier creates a new image copier
func NewCopier(logger *log.Logger) *Copier {
	// Create default transfer manager
	transferOpts := network.DefaultTransferOptions()
	transferMgr := network.NewTransferManager(logger, transferOpts)

	return &Copier{
		logger:         logger,
		workerPool:     4,                      // Default to 4 concurrent workers for layer operations
		metrics:        &metrics.NoopMetrics{}, // Default to no-op metrics
		transferMgr:    transferMgr,
		enableOptimize: true, // Enable optimization by default
	}
}

// WithNetworkOptimization enables or disables network optimization
func (c *Copier) WithNetworkOptimization(enable bool) *Copier {
	c.enableOptimize = enable
	return c
}

// WithTransferOptions configures the transfer options
func (c *Copier) WithTransferOptions(options network.TransferOptions) *Copier {
	c.transferMgr = network.NewTransferManager(c.logger, options)
	return c
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

// WithSigningManager sets the signing manager for image signing
func (c *Copier) WithSigningManager(manager *signing.Manager) *Copier {
	c.signManager = manager
	return c
}

// WithEncryptionManager sets the encryption manager for image encryption
func (c *Copier) WithEncryptionManager(manager *encryption.Manager) *Copier {
	c.encryptManager = manager
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

	// If signature verification is enabled, verify the image signature
	if c.signManager != nil && options.VerifySignatures {
		if err := c.verifyImageSignature(ctx, sourceRepo, sourceTag, manifest); err != nil {
			c.metrics.ReplicationFailed()
			return fmt.Errorf("image signature verification failed: %w", err)
		}
		c.logger.Debug("Image signature verification succeeded", map[string]interface{}{
			"repository": sourceRepo.GetRepositoryName(),
			"tag":        sourceTag,
		})
	}

	// Check if we need to handle a manifest list (multi-arch image)
	if isManifestList(mediaType) {
		return c.copyManifestList(ctx, sourceRepo, destRepo, sourceTag, destTag, manifest, mediaType, options)
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
	blobs, err := c.copyLayers(ctx, sourceRepo, destRepo, manifest, mediaType, options)
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

	// If signing is enabled, sign the destination image
	if c.signManager != nil && c.signManager.IsSigningEnabled() {
		if err := c.signImage(ctx, destRepo, destTag, manifest); err != nil {
			c.logger.Warn("Failed to sign image, but copy was successful", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			c.logger.Debug("Image signed successfully", map[string]interface{}{
				"repository": destRepo.GetRepositoryName(),
				"tag":        destTag,
			})
		}
	}

	duration := time.Since(start)

	// Calculate total bytes copied (approximate)
	var totalBytes int64
	// This is just for metrics - we're not tracking actual size in this implementation
	// In a real implementation, we would track the actual size of each blob
	totalBytes = int64(len(blobs)) * 1024 * 1024 // Assume 1MB per blob for metrics

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

// signImage signs an image and stores the signature
func (c *Copier) signImage(ctx context.Context, repo common.Repository, tag string, manifest []byte) error {
	// Skip if signing manager is not configured
	if c.signManager == nil {
		return nil
	}

	// Calculate the manifest digest
	manifestDigest, err := util.CalculateDigest(manifest)
	if err != nil {
		return fmt.Errorf("failed to calculate manifest digest: %w", err)
	}

	// Create the signature payload
	payload := &signing.SignaturePayload{
		ManifestDigest: manifestDigest,
		Repository:     repo.GetRepositoryName(),
		Tag:            tag,
		AdditionalData: map[string]string{
			"timestamp": fmt.Sprintf("%d", time.Now().Unix()),
		},
	}

	// Sign the image
	_, err = c.signManager.SignImage(ctx, payload)
	return err
}

// verifyImageSignature verifies an image's signature
func (c *Copier) verifyImageSignature(ctx context.Context, repo common.Repository, tag string, manifest []byte) error {
	// Skip if signing manager is not configured
	if c.signManager == nil {
		return nil
	}

	// Calculate the manifest digest
	manifestDigest, err := util.CalculateDigest(manifest)
	if err != nil {
		return fmt.Errorf("failed to calculate manifest digest: %w", err)
	}

	// Create the signature payload
	payload := &signing.SignaturePayload{
		ManifestDigest: manifestDigest,
		Repository:     repo.GetRepositoryName(),
		Tag:            tag,
	}

	// Try to get the signature from storage
	signature, err := c.signManager.GetSignatureFromStorage(ctx, repo.GetRepositoryName(), tag)
	if err != nil {
		return fmt.Errorf("failed to get signature: %w", err)
	}

	// Verify the signature
	valid, err := c.signManager.VerifyImageSignature(ctx, payload, signature)
	if err != nil {
		return fmt.Errorf("signature verification error: %w", err)
	}

	if !valid {
		return fmt.Errorf("invalid signature for image %s:%s", repo.GetRepositoryName(), tag)
	}

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
	options CopyOptions,
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
			_, err = c.copyLayers(ctx, sourceRepo, destRepo, manifestBytes, manifestMediaType, options)
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

	// If signing is enabled, sign the multi-arch image
	if c.signManager != nil && c.signManager.IsSigningEnabled() {
		if err := c.signImage(ctx, destRepo, destTag, manifestList); err != nil {
			c.logger.Warn("Failed to sign multi-arch image, but copy was successful", map[string]interface{}{
				"error": err.Error(),
			})
		}
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
	options CopyOptions,
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
	networkStats := make(chan *network.TransferResult, len(allBlobs))

	// Total transfer statistics
	var totalOriginalSize int64
	var totalTransferSize int64
	var totalSavingsBytes int64

	// Determine if we should use encryption for blobs
	useEncryption := c.encryptManager != nil && options.EncryptionOptions != nil

	// Check for customer-managed keys if specified
	if useEncryption && options.UseCustomerManagedKeys {
		isCMK, err := c.encryptManager.IsCustomerManagedKeyEnabled(ctx)
		if err != nil {
			c.logger.Warn("Failed to check if customer-managed key is enabled", map[string]interface{}{
				"error": err.Error(),
			})
		} else if !isCMK {
			c.logger.Warn("Customer-managed key was requested but is not enabled", nil)
		} else {
			c.logger.Info("Using customer-managed encryption key", nil)
		}
	}

	for _, blob := range allBlobs {
		wg.Add(1)

		// Create a copy of the loop variables to use in the goroutine
		digest := blob.Digest
		mediaType := blob.MediaType

		go func() {
			defer wg.Done()
			sem <- struct{}{}        // Acquire token
			defer func() { <-sem }() // Release token

			c.logger.Debug("Copying blob", map[string]interface{}{
				"digest":     digest,
				"media_type": mediaType,
			})

			// Check if network optimization is enabled
			if c.enableOptimize && options.EnableNetworkOptimization {
				// Use the transfer manager to optimize the transfer
				result, err := c.transferMgr.TransferBlob(ctx, sourceRepo, destRepo, digest, mediaType)
				if err != nil {
					errors <- err
					return
				}

				networkStats <- result
				copiedBlobs <- digest
				return
			}

			// If network optimization is disabled, use the original approach
			// Create a pseudo-tag from the digest
			digestTag := "@" + digest

			// Apply retry logic for blob transfers
			err := util.RetryWithBackoff(ctx, 3, time.Second, 10*time.Second, func() error {
				// Get the blob from the source
				blobData, _, err := sourceRepo.GetManifest(digestTag)
				if err != nil {
					return fmt.Errorf("failed to get blob %s: %w", digest, err)
				}

				// If encryption is enabled, encrypt the blob
				if useEncryption {
					encryptedData, err := c.encryptManager.Encrypt(ctx, blobData)
					if err != nil {
						return fmt.Errorf("failed to encrypt blob %s: %w", digest, err)
					}
					blobData = encryptedData
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
	close(networkStats)

	// Check for errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to copy one or more blobs: %v", errs)
	}

	// Process network statistics if optimization was enabled
	if c.enableOptimize && options.EnableNetworkOptimization {
		for stat := range networkStats {
			totalOriginalSize += int64(stat.Size)
			totalTransferSize += int64(stat.TransferSize)

			// Log detailed stats for each blob
			c.logger.Debug("Blob transfer statistics", map[string]interface{}{
				"digest":            stat.Digest,
				"original_size":     stat.Size,
				"transfer_size":     stat.TransferSize,
				"compression_used":  stat.UsedCompression,
				"delta_used":        stat.UsedDelta,
				"total_savings_pct": stat.TotalSavings,
			})
		}

		// Calculate overall savings
		if totalOriginalSize > 0 {
			totalSavingsBytes = totalOriginalSize - totalTransferSize
			savingsPercent := 100 * (float64(totalSavingsBytes) / float64(totalOriginalSize))

			c.logger.Info("Network optimization summary", map[string]interface{}{
				"total_original_size":   totalOriginalSize,
				"total_transfer_size":   totalTransferSize,
				"total_savings_bytes":   totalSavingsBytes,
				"total_savings_percent": savingsPercent,
				"blob_count":            len(allBlobs),
			})
		}
	}

	// Collect all the copied blob digests
	var blobs []string
	for blob := range copiedBlobs {
		blobs = append(blobs, blob)
	}

	return blobs, nil
}
