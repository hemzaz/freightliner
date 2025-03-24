package network

import (
	"context"
	"fmt"
	"time"

	"github.com/hemzaz/freightliner/src/internal/log"
	"github.com/hemzaz/freightliner/src/internal/util"
	"github.com/hemzaz/freightliner/src/pkg/client/common"
)

// TransferOptions configures the network optimization behavior
type TransferOptions struct {
	// EnableCompression enables data compression during transfers
	EnableCompression bool

	// CompressionOptions configures compression behavior
	CompressionOptions CompressionOptions

	// EnableDelta enables delta-based transfers
	EnableDelta bool

	// DeltaOptions configures delta behavior
	DeltaOptions DeltaOptions

	// RetryAttempts is the number of times to retry failed transfers
	RetryAttempts int

	// RetryInitialDelay is the initial delay before retrying (increases with backoff)
	RetryInitialDelay time.Duration

	// RetryMaxDelay is the maximum delay between retry attempts
	RetryMaxDelay time.Duration
}

// DefaultTransferOptions returns sensible default transfer options
func DefaultTransferOptions() TransferOptions {
	return TransferOptions{
		EnableCompression:   true,
		CompressionOptions:  DefaultCompressionOptions(),
		EnableDelta:         true,
		DeltaOptions:        DefaultDeltaOptions(),
		RetryAttempts:       3,
		RetryInitialDelay:   time.Second,
		RetryMaxDelay:       30 * time.Second,
	}
}

// TransferManager optimizes network transfers between registries
type TransferManager struct {
	logger     *log.Logger
	opts       TransferOptions
	deltaMan   *DeltaManager
}

// NewTransferManager creates a new transfer manager
func NewTransferManager(logger *log.Logger, opts TransferOptions) *TransferManager {
	return &TransferManager{
		logger:   logger,
		opts:     opts,
		deltaMan: NewDeltaManager(logger, opts.DeltaOptions),
	}
}

// TransferResult contains information about a transfer operation
type TransferResult struct {
	// Digest is the digest of the transferred blob
	Digest string

	// Size is the original size of the blob
	Size int

	// TransferSize is the actual number of bytes transferred
	TransferSize int

	// CompressionSavings is the percentage saved by compression
	CompressionSavings float64

	// DeltaSavings is the percentage saved by delta transfer
	DeltaSavings float64

	// TotalSavings is the overall percentage saved
	TotalSavings float64

	// Duration is how long the transfer took
	Duration time.Duration

	// UsedDelta indicates if delta transfer was used
	UsedDelta bool

	// UsedCompression indicates if compression was used
	UsedCompression bool
}

// TransferBlob transfers a blob between repositories with optimizations
func (t *TransferManager) TransferBlob(
	ctx context.Context,
	sourceRepo common.Repository,
	destRepo common.Repository,
	digest string,
	mediaType string,
) (*TransferResult, error) {
	start := time.Now()
	digestTag := "@" + digest

	result := &TransferResult{
		Digest: digest,
		UsedDelta: false,
		UsedCompression: false,
	}

	// Try delta transfer if enabled
	if t.opts.EnableDelta {
		deltaSummary, deltaUsed, err := t.deltaMan.OptimizeTransfer(sourceRepo, destRepo, digest, mediaType)
		if err != nil {
			t.logger.Warn("Delta transfer failed, falling back to full transfer", map[string]interface{}{
				"digest": digest,
				"error":  err.Error(),
			})
		} else if deltaUsed {
			result.UsedDelta = true
			result.Size = deltaSummary.OriginalSize
			result.TransferSize = deltaSummary.TransferSize
			result.DeltaSavings = deltaSummary.SavingsPercent
			result.TotalSavings = deltaSummary.SavingsPercent
			result.Duration = time.Since(start)
			
			// Delta transfer was successful, no need for full transfer
			return result, nil
		}
	}

	// If we get here, we need to do a full transfer
	var transferErr error
	var blobData []byte
	var originalSize int
	var transferSize int

	// Get the blob from the source with retry logic
	transferErr = util.RetryWithBackoff(ctx, t.opts.RetryAttempts, 
		t.opts.RetryInitialDelay, t.opts.RetryMaxDelay, func() error {
		var err error
		blobData, _, err = sourceRepo.GetManifest(digestTag)
		if err != nil {
			return fmt.Errorf("failed to get blob %s: %w", digest, err)
		}
		return nil
	})

	if transferErr != nil {
		return nil, transferErr
	}

	originalSize = len(blobData)
	transferSize = originalSize

	// Compress the data if enabled
	if t.opts.EnableCompression && originalSize > t.opts.CompressionOptions.MinSize {
		compressedData, err := Compress(blobData, t.opts.CompressionOptions)
		if err != nil {
			t.logger.Warn("Compression failed, using uncompressed data", map[string]interface{}{
				"digest": digest,
				"error":  err.Error(),
			})
		} else if len(compressedData) < len(blobData) {
			// Use compressed data only if it's smaller
			blobData = compressedData
			transferSize = len(compressedData)
			result.UsedCompression = true
			result.CompressionSavings = 100 * (1 - float64(transferSize)/float64(originalSize))
		}
	}

	// Write to destination with retry logic
	transferErr = util.RetryWithBackoff(ctx, t.opts.RetryAttempts, 
		t.opts.RetryInitialDelay, t.opts.RetryMaxDelay, func() error {
		// If we used compression, we need to decompress before writing to destination
		dataToWrite := blobData
		if result.UsedCompression {
			var err error
			dataToWrite, err = Decompress(blobData, t.opts.CompressionOptions.Type)
			if err != nil {
				return fmt.Errorf("failed to decompress data: %w", err)
			}
		}
		
		// Put the blob in the destination
		return destRepo.PutManifest(digestTag, dataToWrite, mediaType)
	})

	if transferErr != nil {
		return nil, transferErr
	}

	// Calculate total savings and update result
	result.Size = originalSize
	result.TransferSize = transferSize
	result.TotalSavings = 100 * (1 - float64(transferSize)/float64(originalSize))
	result.Duration = time.Since(start)

	t.logger.Debug("Blob transfer completed", map[string]interface{}{
		"digest":         digest,
		"size":           originalSize,
		"transfer_size":  transferSize,
		"total_savings":  result.TotalSavings,
		"used_delta":     result.UsedDelta,
		"used_compression": result.UsedCompression,
		"duration_ms":    result.Duration.Milliseconds(),
	})

	return result, nil
}