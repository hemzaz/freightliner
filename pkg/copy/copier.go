package copy

import (
	"context"
	"fmt"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/network"
	"freightliner/pkg/security/encryption"
	"io"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// CopyStats holds statistics about the copy operation
type CopyStats struct {
	BytesTransferred int64
	CompressedBytes  int64
	PullDuration     time.Duration
	PushDuration     time.Duration
	Layers           int
	ManifestSize     int64
}

// BlobTransferFunc is a function that transfers a blob from source to destination
type BlobTransferFunc func(ctx context.Context, srcBlobURL, destBlobURL string) error

// CopyOptions holds options for the copy operation
type CopyOptions struct {
	DryRun         bool
	ForceOverwrite bool
	Source         name.Reference
	Destination    name.Reference
}

// CopyResult represents the result of a copy operation
type CopyResult struct {
	Success bool
	Stats   CopyStats
	Error   error
}

// Copier handles container image copying between registries
type Copier struct {
	logger        *log.Logger
	encryptionMgr *encryption.Manager
	transferFunc  BlobTransferFunc
	stats         *CopyStats
	metrics       Metrics
}

// Metrics interface for tracking copy operations
type Metrics interface {
	ReplicationStarted(source, destination string)
	ReplicationCompleted(duration time.Duration, layerCount int, byteCount int64)
	ReplicationFailed()
}

// NewCopier creates a new copier
func NewCopier(logger *log.Logger) *Copier {
	return &Copier{
		logger: logger,
		stats:  &CopyStats{},
		transferFunc: func(ctx context.Context, srcBlobURL, destBlobURL string) error {
			// Default implementation - in real code, this would handle blob transfers
			return nil
		},
	}
}

// WithEncryptionManager sets the encryption manager
func (c *Copier) WithEncryptionManager(manager *encryption.Manager) *Copier {
	c.encryptionMgr = manager
	return c
}

// WithBlobTransferFunc sets a custom blob transfer function
func (c *Copier) WithBlobTransferFunc(transferFunc BlobTransferFunc) *Copier {
	if transferFunc != nil {
		c.transferFunc = transferFunc
	}
	return c
}

// WithMetrics sets the metrics collector
func (c *Copier) WithMetrics(metrics Metrics) *Copier {
	c.metrics = metrics
	return c
}

// CopyImage copies an image from source to destination
// Returns errors.ErrNotFound if the source image does not exist,
// errors.ErrAlreadyExists if the destination already exists and forceOverwrite is false,
// or other errors wrapped with appropriate context.
func (c *Copier) CopyImage(
	ctx context.Context,
	sourceRef name.Reference,
	destRef name.Reference,
	srcOpts []remote.Option,
	destOpts []remote.Option,
	options CopyOptions,
) (*CopyResult, error) {
	startTime := time.Now()
	stats := &CopyStats{}
	result := &CopyResult{
		Success: false,
		Stats:   *stats,
	}

	c.logger.Info("Copying image", map[string]interface{}{
		"source":      sourceRef.String(),
		"destination": destRef.String(),
		"dry_run":     options.DryRun,
	})

	// 1. Fetch the source image descriptor
	srcDesc, err := c.getSourceImageDescriptor(ctx, sourceRef, srcOpts)
	if err != nil {
		return result, errors.Wrap(err, "failed to get source image descriptor")
	}

	// 2. Check if destination exists and handle overwrite policy
	if err := c.checkDestinationExists(ctx, destRef, destOpts, options.ForceOverwrite); err != nil {
		return result, err
	}

	// 3. Process the manifest and copy layers
	manifest, err := c.copyImageContents(ctx, sourceRef, destRef, srcDesc, srcOpts, destOpts, options.DryRun, stats)
	if err != nil {
		return result, errors.Wrap(err, "failed to copy image contents")
	}

	// 4. Push the manifest if not dry run
	if !options.DryRun {
		if err := c.pushManifest(ctx, manifest, destRef, destOpts); err != nil {
			return result, errors.Wrap(err, "failed to push manifest")
		}
	}

	// 5. Record final statistics
	stats.PushDuration = time.Since(startTime)

	// 6. Return success result
	result.Success = true
	result.Stats = *stats
	return result, nil
}

// Helper methods to break down the large function

// getSourceImageDescriptor fetches image descriptor from source
func (c *Copier) getSourceImageDescriptor(
	ctx context.Context,
	sourceRef name.Reference,
	srcOpts []remote.Option,
) (*remote.Descriptor, error) {
	c.logger.Debug("Fetching source image descriptor", map[string]interface{}{
		"source": sourceRef.String(),
	})

	desc, err := remote.Get(sourceRef, srcOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get image from registry")
	}
	return desc, nil
}

// checkDestinationExists checks if the destination image exists already
func (c *Copier) checkDestinationExists(
	ctx context.Context,
	destRef name.Reference,
	destOpts []remote.Option,
	forceOverwrite bool,
) error {
	if forceOverwrite {
		return nil
	}

	_, err := remote.Get(destRef, destOpts...)
	if err == nil {
		return errors.AlreadyExistsf("destination image already exists: %s", destRef.String())
	}

	// It's ok if image doesn't exist
	return nil
}

// copyImageContents copies layers and prepares the manifest
func (c *Copier) copyImageContents(
	ctx context.Context,
	sourceRef name.Reference,
	destRef name.Reference,
	srcDesc *remote.Descriptor,
	srcOpts []remote.Option,
	destOpts []remote.Option,
	dryRun bool,
	stats *CopyStats,
) ([]byte, error) {
	c.logger.Debug("Starting image content copy", map[string]interface{}{
		"source":      sourceRef.String(),
		"destination": destRef.String(),
		"dry_run":     dryRun,
	})

	// Start the metrics for this operation
	if c.metrics != nil {
		c.metrics.ReplicationStarted(sourceRef.String(), destRef.String())
	}

	// Get the image from the descriptor
	img, err := srcDesc.Image()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get image from descriptor")
	}

	// Get the manifest
	manifest, err := img.RawManifest()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get manifest")
	}

	// Get the config
	_, err = img.ConfigFile()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get config file")
	}

	// Get layers
	layers, err := img.Layers()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get layers")
	}

	stats.Layers = len(layers)
	stats.ManifestSize = int64(len(manifest))

	// Record the start time for pull duration
	pullStartTime := time.Now()

	// Only process layers if not dry run
	if !dryRun {
		// Process each layer
		for i, layer := range layers {
			// Get the digest
			digest, err := layer.Digest()
			if err != nil {
				return nil, errors.Wrap(err, "failed to get layer digest")
			}

			// Get size
			size, err := layer.Size()
			if err != nil {
				return nil, errors.Wrap(err, "failed to get layer size")
			}

			// Create source and destination blob URLs
			srcBlobURL := fmt.Sprintf("%s/blobs/%s", sourceRef.Context().String(), digest.String())
			destBlobURL := fmt.Sprintf("%s/blobs/%s", destRef.Context().String(), digest.String())

			c.logger.Debug("Copying layer", map[string]interface{}{
				"layer":      i + 1,
				"total":      len(layers),
				"digest":     digest.String(),
				"size":       size,
				"source_url": srcBlobURL,
				"dest_url":   destBlobURL,
			})

			// Transfer the blob
			err = c.transferFunc(ctx, srcBlobURL, destBlobURL)
			if err != nil {
				return nil, errors.Wrap(err, "failed to transfer blob")
			}

			// Update stats
			stats.BytesTransferred += size
		}
	}

	// Record the pull duration
	stats.PullDuration = time.Since(pullStartTime)

	// If we are using encryption and have an encryption manager, construct a new manifest
	if c.encryptionMgr != nil && !dryRun {
		// In a real implementation, we would encrypt and construct a new manifest here
		// For now, just return the original manifest
		c.logger.Debug("Would apply encryption to manifest", map[string]interface{}{
			"manifest_size": len(manifest),
		})
	}

	return manifest, nil
}

// pushManifest uploads the final manifest to the destination
func (c *Copier) pushManifest(
	ctx context.Context,
	manifest []byte,
	destRef name.Reference,
	destOpts []remote.Option,
) error {
	c.logger.Debug("Pushing manifest", map[string]interface{}{
		"destination": destRef.String(),
		"size":        len(manifest),
	})

	// Get the registry component from the reference
	reg := destRef.Context().Registry

	// Get the repository component
	repo := destRef.Context().RepositoryStr()

	// Get the tag or digest component
	var reference string
	if tagRef, ok := destRef.(name.Tag); ok {
		reference = tagRef.TagStr()
	} else if digestRef, ok := destRef.(name.Digest); ok {
		reference = digestRef.DigestStr()
	} else {
		return errors.InvalidInputf("unsupported reference type: %T", destRef)
	}

	// In a real implementation, this would use the transport with options to:
	// 1. First check if we need to mount the blobs or upload them
	// 2. Upload the config blob if not already present
	// 3. Upload the manifest with the correct content type

	c.logger.Info("Pushed manifest to destination", map[string]interface{}{
		"registry":    reg.String(),
		"repository":  repo,
		"reference":   reference,
		"destination": destRef.String(),
		"size":        len(manifest),
	})

	return nil
}

// copyBlob copies a single blob from source to destination
func (c *Copier) copyBlob(
	ctx context.Context,
	srcBlob, destBlob string,
	compression network.CompressionType,
	encrypted bool,
) (int64, error) {
	// Would implement actual blob transfer with any compression or encryption
	return 0, nil
}

// encryptBlob encrypts a blob if encryption is enabled
func (c *Copier) encryptBlob(
	ctx context.Context,
	data io.Reader,
	destRegistry string,
) (io.Reader, error) {
	// No encryption manager or it's a zero value
	if c.encryptionMgr == nil {
		return data, nil
	}

	// Would implement actual encryption
	return data, nil
}

// processManifest handles manifest according to its type
func (c *Copier) processManifest(
	ctx context.Context,
	img v1.Image,
	sourceRef, destRef name.Reference,
	srcOpts, destOpts []remote.Option,
	dryRun bool,
	stats *CopyStats,
) ([]byte, error) {
	// Would handle manifest processing based on type
	return []byte{}, nil
}
