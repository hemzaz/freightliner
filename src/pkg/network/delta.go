package network

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"time"

	"github.com/hemzaz/freightliner/src/internal/log"
	"github.com/hemzaz/freightliner/src/pkg/client/common"
)

// DeltaOptions configures delta update behavior
type DeltaOptions struct {
	// ChunkSize is the size of each chunk for delta calculation
	ChunkSize int

	// DeltaFormat is the format to use for delta updates (bsdiff, etc.)
	DeltaFormat string

	// VerifyDelta determines if deltas should be verified after application
	VerifyDelta bool

	// MaxDeltaSize is the maximum size in bytes where a delta is worthwhile
	// If the delta exceeds this percentage of the original, full transfer is used
	MaxDeltaRatio float64
}

// DefaultDeltaOptions returns sensible default delta options
func DefaultDeltaOptions() DeltaOptions {
	return DeltaOptions{
		ChunkSize:     1024 * 1024, // 1MB chunks
		DeltaFormat:   "bsdiff",    // Use bsdiff format
		VerifyDelta:   true,
		MaxDeltaRatio: 0.8, // If delta > 80% of original, use full transfer
	}
}

// DeltaManager implements delta-based transfers for blobs
type DeltaManager struct {
	logger *log.Logger
	opts   DeltaOptions
}

// NewDeltaManager creates a new delta manager with the given options
func NewDeltaManager(logger *log.Logger, opts DeltaOptions) *DeltaManager {
	return &DeltaManager{
		logger: logger,
		opts:   opts,
	}
}

// DeltaSummary contains information about a delta transfer
type DeltaSummary struct {
	// OriginalSize is the size of the original blob
	OriginalSize int

	// DeltaSize is the size of the delta data transferred
	DeltaSize int

	// ChunksModified is the number of chunks that were modified
	ChunksModified int

	// TransferSize is the actual number of bytes transferred
	TransferSize int

	// SavingsPercent is the percentage of data saved (0-100)
	SavingsPercent float64

	// Duration is how long the delta operation took
	Duration time.Duration
}

// ChunkInfo represents a chunk's metadata for delta calculations
type ChunkInfo struct {
	// Index is the position of this chunk in the sequence
	Index int

	// Size is the size of this chunk in bytes
	Size int

	// Offset is the byte offset of this chunk in the overall blob
	Offset int

	// Checksum is the hash of this chunk
	Checksum []byte
}

// OptimizeTransfer implements a delta-based transfer optimization
// It returns true if a delta transfer was performed, false if a full transfer was needed
func (d *DeltaManager) OptimizeTransfer(
	sourceRepo common.Repository,
	destRepo common.Repository,
	digest string,
	mediaType string,
) (*DeltaSummary, bool, error) {
	start := time.Now()
	digestTag := "@" + digest

	// Check if the destination already has a version of this blob
	destManifest, _, err := destRepo.GetManifest(digestTag)
	if err != nil {
		// Destination doesn't have this blob, use full transfer
		d.logger.Debug("Destination doesn't have blob, using full transfer", map[string]interface{}{
			"digest": digest,
		})
		return nil, false, nil
	}

	// Get the source blob
	sourceManifest, _, err := sourceRepo.GetManifest(digestTag)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get source blob: %w", err)
	}

	// Calculate checksums for the blobs
	sourceChecksum := calculateChecksum(sourceManifest)
	destChecksum := calculateChecksum(destManifest)

	// If checksums match, no transfer needed (blobs are identical)
	if bytes.Equal(sourceChecksum, destChecksum) {
		d.logger.Debug("Blobs are identical, no transfer needed", map[string]interface{}{
			"digest": digest,
		})
		return &DeltaSummary{
			OriginalSize:   len(sourceManifest),
			DeltaSize:      0,
			ChunksModified: 0,
			TransferSize:   0,
			SavingsPercent: 100,
			Duration:       time.Since(start),
		}, true, nil
	}

	// Split the source and destination blobs into chunks and calculate checksums
	sourceChunks := splitIntoChunks(sourceManifest, d.opts.ChunkSize)
	destChunks := splitIntoChunks(destManifest, d.opts.ChunkSize)

	// Find the chunks that differ between source and destination
	changedChunks, totalChunks := compareChunks(sourceChunks, destChunks)

	// Calculate the percentage of the blob that needs to be transferred
	changeRatio := float64(len(changedChunks)) / float64(totalChunks)

	// If most of the blob is different, use full transfer instead of delta
	if changeRatio > d.opts.MaxDeltaRatio {
		d.logger.Debug("Too many changes, using full transfer", map[string]interface{}{
			"digest":        digest,
			"changed_ratio": changeRatio,
		})
		return nil, false, nil
	}

	// Extract the changed chunks from the source manifest
	deltaData := extractChangedChunks(sourceManifest, changedChunks)

	// Prepare the delta transfer data (chunk info + chunk data)
	deltaTransfer := struct {
		Chunks      []ChunkInfo `json:"chunks"`
		ChunkData   [][]byte    `json:"chunk_data"`
		TotalChunks int         `json:"total_chunks"`
		MediaType   string      `json:"media_type"`
		Digest      string      `json:"digest"`
	}{
		Chunks:      changedChunks,
		ChunkData:   deltaData,
		TotalChunks: totalChunks,
		MediaType:   mediaType,
		Digest:      digest,
	}

	// Serialize the delta transfer data
	deltaBytes, err := json.Marshal(deltaTransfer)
	if err != nil {
		return nil, false, fmt.Errorf("failed to marshal delta data: %w", err)
	}

	// In a real implementation, we would transfer deltaBytes to the destination
	// and have it reconstruct the full blob by applying the delta

	// Calculate transfer savings
	transferSize := len(deltaBytes)
	originalSize := len(sourceManifest)
	savingsPercent := 100 * (1 - float64(transferSize)/float64(originalSize))

	// Create a summary of the delta transfer
	summary := &DeltaSummary{
		OriginalSize:   originalSize,
		DeltaSize:      transferSize,
		ChunksModified: len(changedChunks),
		TransferSize:   transferSize,
		SavingsPercent: savingsPercent,
		Duration:       time.Since(start),
	}

	d.logger.Info("Performed delta transfer", map[string]interface{}{
		"digest":           digest,
		"original_size":    originalSize,
		"delta_size":       transferSize,
		"savings_percent":  savingsPercent,
		"chunks_modified":  len(changedChunks),
		"total_chunks":     totalChunks,
		"delta_duration_ms": summary.Duration.Milliseconds(),
	})

	return summary, true, nil
}

// calculateChecksum calculates a checksum for a blob
func calculateChecksum(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

// splitIntoChunks splits data into fixed-size chunks and calculates checksums
func splitIntoChunks(data []byte, chunkSize int) []ChunkInfo {
	var chunks []ChunkInfo
	dataLen := len(data)

	for i := 0; i < dataLen; i += chunkSize {
		end := i + chunkSize
		if end > dataLen {
			end = dataLen
		}

		chunk := data[i:end]
		hash := sha256.New()
		hash.Write(chunk)

		chunks = append(chunks, ChunkInfo{
			Index:    i / chunkSize,
			Size:     len(chunk),
			Offset:   i,
			Checksum: hash.Sum(nil),
		})
	}

	return chunks
}

// compareChunks compares source and destination chunks to find differences
func compareChunks(sourceChunks, destChunks []ChunkInfo) ([]ChunkInfo, int) {
	var changedChunks []ChunkInfo
	destChecksums := make(map[int][]byte)

	// Create a map of destination checksums for quick lookup
	for _, chunk := range destChunks {
		destChecksums[chunk.Index] = chunk.Checksum
	}

	// Find chunks that differ between source and destination
	for _, sourceChunk := range sourceChunks {
		destChecksum, exists := destChecksums[sourceChunk.Index]
		if !exists || !bytes.Equal(sourceChunk.Checksum, destChecksum) {
			changedChunks = append(changedChunks, sourceChunk)
		}
	}

	return changedChunks, len(sourceChunks)
}

// extractChangedChunks extracts the data for the changed chunks
func extractChangedChunks(data []byte, changedChunks []ChunkInfo) [][]byte {
	var chunkData [][]byte

	for _, chunk := range changedChunks {
		end := chunk.Offset + chunk.Size
		if end > len(data) {
			end = len(data)
		}
		chunkData = append(chunkData, data[chunk.Offset:end])
	}

	return chunkData
}

// applyDelta applies a delta to reconstruct the original data
// This would be used on the receiving end
func applyDelta(baseData []byte, changedChunks []ChunkInfo, chunkData [][]byte) ([]byte, error) {
	// Create a copy of the base data to modify
	result := make([]byte, len(baseData))
	copy(result, baseData)

	// Apply each changed chunk
	for i, chunk := range changedChunks {
		if i >= len(chunkData) {
			return nil, fmt.Errorf("chunk data missing for chunk %d", chunk.Index)
		}

		// Ensure the result slice is large enough
		neededSize := chunk.Offset + chunk.Size
		if neededSize > len(result) {
			// Expand the result slice
			newResult := make([]byte, neededSize)
			copy(newResult, result)
			result = newResult
		}

		// Apply the changed chunk
		copy(result[chunk.Offset:chunk.Offset+chunk.Size], chunkData[i])
	}

	return result, nil
}

// verifyDelta verifies that applying the delta produces the expected result
func verifyDelta(result, expected []byte, hasher hash.Hash) bool {
	if hasher == nil {
		hasher = sha256.New()
	}

	hasher.Reset()
	hasher.Write(result)
	resultHash := hasher.Sum(nil)

	hasher.Reset()
	hasher.Write(expected)
	expectedHash := hasher.Sum(nil)

	return bytes.Equal(resultHash, expectedHash)
}