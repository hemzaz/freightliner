package copy

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/network"
	"freightliner/pkg/security/encryption"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
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

			// Transfer the blob with proper implementation
			transferred, err := c.transferBlob(ctx, layer, sourceRef, destRef, srcOpts, destOpts)
			if err != nil {
				return nil, errors.Wrap(err, "failed to transfer blob")
			}

			// Update transfer statistics
			stats.BytesTransferred += transferred
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

	// Parse the manifest to get proper media type
	var mediaType types.MediaType
	if bytes.Contains(manifest, []byte("schemaVersion")) {
		if bytes.Contains(manifest, []byte("mediaType")) {
			// Docker Image Manifest V2 Schema 2
			mediaType = types.DockerManifestSchema2
		} else {
			// Docker Image Manifest V2 Schema 1
			mediaType = types.DockerManifestSchema1
		}
	} else {
		// OCI Image Manifest
		mediaType = types.OCIManifestSchema1
	}

	// Create manifest descriptor
	manifestHash, err := v1.NewHash(fmt.Sprintf("sha256:%x", sha256.Sum256(manifest)))
	if err != nil {
		return errors.Wrap(err, "failed to calculate manifest hash")
	}

	// Upload manifest using go-containerregistry
	err = remote.Put(destRef, &manifestDescriptor{
		mediaType: mediaType,
		data:      manifest,
		hash:      manifestHash,
	}, destOpts...)

	if err != nil {
		return errors.Wrap(err, "failed to push manifest to destination")
	}

	c.logger.Info("Successfully pushed manifest to destination", map[string]interface{}{
		"destination": destRef.String(),
		"size":        len(manifest),
		"media_type":  string(mediaType),
		"digest":      manifestHash.String(),
	})

	return nil
}

// manifestDescriptor implements the remote.Taggable interface for manifest uploads
type manifestDescriptor struct {
	mediaType types.MediaType
	data      []byte
	hash      v1.Hash
}

func (m *manifestDescriptor) MediaType() (types.MediaType, error) {
	return m.mediaType, nil
}

func (m *manifestDescriptor) RawManifest() ([]byte, error) {
	return m.data, nil
}

func (m *manifestDescriptor) Digest() (v1.Hash, error) {
	return m.hash, nil
}

// transferBlob handles the actual blob transfer between registries
func (c *Copier) transferBlob(
	ctx context.Context,
	layer v1.Layer,
	sourceRef name.Reference,
	destRef name.Reference,
	srcOpts []remote.Option,
	destOpts []remote.Option,
) (int64, error) {
	// Get layer properties
	digest, err := layer.Digest()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get layer digest")
	}

	size, err := layer.Size()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get layer size")
	}

	c.logger.Debug("Transferring blob", map[string]interface{}{
		"digest": digest.String(),
		"size":   size,
		"source": sourceRef.String(),
		"dest":   destRef.String(),
	})

	// Check if blob already exists at destination
	if exists, checkErr := c.checkBlobExists(ctx, destRef, digest, destOpts); checkErr == nil && exists {
		c.logger.Debug("Blob already exists at destination, skipping", map[string]interface{}{
			"digest": digest.String(),
		})
		return 0, nil // Already exists, no bytes transferred
	}

	// Get layer reader from source
	reader, err := layer.Compressed()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get layer reader")
	}
	defer reader.Close()

	// Apply compression if needed
	processedReader := reader
	if c.shouldCompress(size) {
		processedReader, err = c.compressStream(reader)
		if err != nil {
			return 0, errors.Wrap(err, "failed to compress stream")
		}
		defer processedReader.Close()
	}

	// Apply encryption if configured
	if c.encryptionMgr != nil {
		processedReader, err = c.encryptBlob(ctx, processedReader, destRef.Context().RegistryStr())
		if err != nil {
			return 0, errors.Wrap(err, "failed to encrypt blob")
		}
		defer processedReader.Close()
	}

	// Upload blob to destination
	err = c.uploadBlob(ctx, destRef, digest, processedReader, destOpts)
	if err != nil {
		return 0, errors.Wrap(err, "failed to upload blob")
	}

	c.logger.Debug("Successfully transferred blob", map[string]interface{}{
		"digest": digest.String(),
		"size":   size,
	})

	return size, nil
}

// checkBlobExists checks if a blob already exists at the destination
func (c *Copier) checkBlobExists(
	ctx context.Context,
	destRef name.Reference,
	digest v1.Hash,
	destOpts []remote.Option,
) (bool, error) {
	// Create blob reference
	blobRef := destRef.Context().Digest(digest.String())

	// Try to get the blob descriptor
	_, err := remote.Get(blobRef, destOpts...)
	if err != nil {
		// If error contains "not found" or similar, blob doesn't exist
		return false, nil
	}

	// No error means blob exists
	return true, nil
}

// shouldCompress determines if a layer should be compressed based on size
func (c *Copier) shouldCompress(size int64) bool {
	// Only compress layers larger than 1KB to avoid overhead
	const minCompressionSize = 1024
	return size > minCompressionSize
}

// compressStream applies compression to a stream
func (c *Copier) compressStream(reader io.ReadCloser) (io.ReadCloser, error) {
	// Use gzip compression by default
	opts := network.DefaultCompressorOptions()

	// Create a pipe for streaming compression
	pr, pw := io.Pipe()

	// Start compression in a goroutine
	go func() {
		defer pw.Close()
		defer reader.Close()

		// Create compressing writer
		compressor, err := network.NewCompressingWriter(pw, opts)
		if err != nil {
			pw.CloseWithError(errors.Wrap(err, "failed to create compressor"))
			return
		}
		defer compressor.Close()

		// Copy data through compressor
		if _, err := io.Copy(compressor, reader); err != nil {
			pw.CloseWithError(errors.Wrap(err, "compression failed"))
			return
		}
	}()

	return pr, nil
}

// uploadBlob uploads a blob to the destination registry
func (c *Copier) uploadBlob(
	ctx context.Context,
	destRef name.Reference,
	digest v1.Hash,
	reader io.Reader,
	destOpts []remote.Option,
) error {
	// For production implementation, we would use the registry's blob upload API
	// This involves:
	// 1. POST to /v2/{name}/blobs/uploads/ to initiate upload
	// 2. PUT to upload URL with digest parameter
	// 3. Handle chunked uploads for large blobs

	// For now, we'll use go-containerregistry's remote package
	// In a real implementation, we'd implement the full registry v2 API

	c.logger.Debug("Uploading blob to destination", map[string]interface{}{
		"destination": destRef.String(),
		"digest":      digest.String(),
	})

	// Read all data (in production, we'd stream this)
	data, err := io.ReadAll(reader)
	if err != nil {
		return errors.Wrap(err, "failed to read blob data")
	}

	// Create a layer from the data
	layer := &blobLayer{
		digestHash: digest,
		data:       data,
	}

	// Upload using remote.WriteLayer
	err = remote.WriteLayer(destRef.Context(), layer, destOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to write layer to destination")
	}

	return nil
}

// blobLayer implements v1.Layer interface for uploading arbitrary data
type blobLayer struct {
	digestHash v1.Hash
	data       []byte
}

func (b *blobLayer) Digest() (v1.Hash, error) {
	return b.digestHash, nil
}

func (b *blobLayer) DiffID() (v1.Hash, error) {
	return b.digestHash, nil
}

func (b *blobLayer) Compressed() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(b.data)), nil
}

func (b *blobLayer) Uncompressed() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(b.data)), nil
}

func (b *blobLayer) Size() (int64, error) {
	return int64(len(b.data)), nil
}

func (b *blobLayer) MediaType() (types.MediaType, error) {
	return types.DockerLayer, nil
}

// copyBlob is the old method - keeping for backwards compatibility but not used
func (c *Copier) copyBlob(
	ctx context.Context,
	srcBlob, destBlob string,
	compression network.CompressionType,
	encrypted bool,
) (int64, error) {
	// This method is deprecated in favor of transferBlob
	// Keeping for backwards compatibility
	return 0, errors.New("copyBlob is deprecated, use transferBlob instead")
}

// encryptBlob encrypts a blob if encryption is enabled
func (c *Copier) encryptBlob(
	ctx context.Context,
	data io.ReadCloser,
	destRegistry string,
) (io.ReadCloser, error) {
	// No encryption manager or it's a zero value
	if c.encryptionMgr == nil {
		return data, nil
	}

	// Would implement actual encryption using the encryption manager
	// For now, just pass through the data
	c.logger.Debug("Encryption manager available but encryption not implemented", map[string]interface{}{
		"registry": destRegistry,
	})
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
