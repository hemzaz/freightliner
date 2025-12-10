package network

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash"
	"math"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/interfaces"
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
		VerifyDelta:   true,        // Verify deltas by default
		MaxDeltaRatio: 0.8,         // If delta is > 80% of original, use full transfer
	}
}

// DeltaGenerator creates deltas between source and destination files
type DeltaGenerator struct {
	options DeltaOptions
	logger  log.Logger
	hasher  hash.Hash
}

// DeltaManager manages delta operations
type DeltaManager struct {
	logger  log.Logger
	options DeltaOptions
}

// getDelta calculates a delta between source and target
func (d *DeltaManager) getDelta(source, target []byte) ([]byte, int64, error) {
	// Create a delta using the configured format
	delta, err := CreateDelta(source, target, d.options.DeltaFormat)
	if err != nil {
		return nil, 0, err
	}

	// Calculate savings in bytes
	savings := int64(len(target)) - int64(len(delta))
	if savings < 0 {
		savings = 0
	}

	return delta, savings, nil
}

// NewDeltaManager creates a new delta manager
func NewDeltaManager(logger log.Logger, opts DeltaOptions) (*DeltaManager, error) {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}
	return &DeltaManager{
		logger:  logger,
		options: opts,
	}, nil
}

// DeltaSummary contains information about a delta optimization
type DeltaSummary struct {
	OriginalSize   int64         // Original size in bytes
	DeltaSize      int64         // Size of the delta in bytes
	ChunksModified int           // Number of chunks that were modified
	TransferSize   int64         // Actual size transferred after delta
	SavingsPercent float64       // Percentage saved vs full transfer
	Duration       time.Duration // Time taken to calculate and apply delta
}

// OptimizeTransfer uses delta techniques to optimize a transfer between repositories
func (d *DeltaManager) OptimizeTransfer(sourceRepo, destRepo interfaces.Repository, digest, mediaType string) (*DeltaSummary, bool, error) {
	startTime := time.Now()
	ctx := context.Background()

	// Try to get the manifests from both source and destination
	sourceManifest, err := sourceRepo.GetManifest(ctx, digest)
	if err != nil {
		return nil, false, errors.Wrap(err, "failed to get source manifest")
	}

	// Try to get the destination manifest
	destManifest, err := destRepo.GetManifest(ctx, digest)

	// Create a summary to return
	summary := &DeltaSummary{
		OriginalSize:   int64(len(sourceManifest.Content)),
		TransferSize:   int64(len(sourceManifest.Content)), // Default to full transfer
		ChunksModified: 0,
		DeltaSize:      0,
		SavingsPercent: 0.0,
	}

	// Check if destination already has exactly the same content
	if err == nil && destManifest != nil {
		// Compare the manifests by digest
		if sourceManifest.Digest == destManifest.Digest {
			// Already identical - nothing to transfer
			d.logger.WithFields(map[string]interface{}{
				"digest": digest,
				"source": sourceRepo.GetRepositoryName(),
				"dest":   destRepo.GetRepositoryName(),
			}).Info("Manifests are identical, skipping transfer")

			summary.TransferSize = 0
			summary.SavingsPercent = 100.0
			summary.Duration = time.Since(startTime)
			return summary, true, nil
		}
	}

	// Check if delta is disabled via MaxDeltaRatio
	if d.options.MaxDeltaRatio <= 0.0 {
		// Delta optimization is disabled
		d.logger.Debug("Delta optimization disabled (MaxDeltaRatio = 0)")
		summary.Duration = time.Since(startTime)
		return summary, false, nil
	}

	// For small manifests below a certain threshold, delta might not be worthwhile
	// Threshold could be configurable, but 1KB is a reasonable default for manifests
	minDeltaSize := 1024 // 1KB threshold
	if len(sourceManifest.Content) < minDeltaSize {
		d.logger.WithFields(map[string]interface{}{
			"size":      len(sourceManifest.Content),
			"threshold": minDeltaSize,
		}).Debug("Manifest too small for delta optimization")
		summary.Duration = time.Since(startTime)
		return summary, false, nil
	}

	// If we get here, we need to calculate a delta to see if it's worthwhile
	// First, we need the destination content if it exists
	targetContent := sourceManifest.Content
	destContent := []byte{}

	if err == nil && destManifest != nil {
		destContent = destManifest.Content
	}

	// Determine the best delta format based on manifest size and content
	var deltaFormat string

	// Start with bsdiff by default for best compression
	deltaFormat = BSDiffFormat

	// For very large manifests, chunk-based might be better
	if len(sourceManifest.Content) > 10*1024*1024 { // 10MB
		deltaFormat = ChunkBasedFormat
	}

	// Try to create a delta
	delta, err := CreateDelta(destContent, targetContent, deltaFormat)
	if err != nil {
		d.logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Warn("Failed to create delta, falling back to full transfer")
		summary.Duration = time.Since(startTime)
		return summary, false, nil
	}

	// Check if the delta is smaller than the MaxDeltaRatio threshold
	deltaRatio := float64(len(delta)) / float64(len(targetContent))
	if deltaRatio > d.options.MaxDeltaRatio {
		d.logger.WithFields(map[string]interface{}{
			"ratio":       deltaRatio,
			"threshold":   d.options.MaxDeltaRatio,
			"delta_size":  len(delta),
			"target_size": len(targetContent),
		}).Debug("Delta too large, using full transfer")
		summary.Duration = time.Since(startTime)
		return summary, false, nil
	}

	// Delta is worth using
	savings := int64(len(targetContent)) - int64(len(delta))
	savingsPercent := float64(savings) / float64(len(targetContent)) * 100.0

	// Update the summary with real values
	summary.DeltaSize = int64(len(delta))
	summary.TransferSize = int64(len(delta))
	summary.SavingsPercent = savingsPercent

	// For chunk-based delta, estimate number of modified chunks
	if deltaFormat == ChunkBasedFormat {
		// Parse the delta header to get chunk info
		if len(delta) >= 4 {
			headerSize := binary.BigEndian.Uint32(delta[:4])
			if len(delta) > int(4+headerSize) {
				var header DeltaHeader
				err := json.Unmarshal(delta[4:4+headerSize], &header)
				if err == nil && header.ChunkCount > 0 {
					// Each chunk map entry is 4 bytes
					chunkMapSize := header.ChunkCount * 4
					if len(delta) > int(4+headerSize+chunkMapSize) {
						// Count how many chunks were modified (have -1 in the chunk map)
						modifiedChunks := 0
						deltaData := delta[4+headerSize:]

						for i := uint32(0); i < header.ChunkCount; i++ {
							offset := i * 4
							if offset+4 <= uint32(len(deltaData)) { // nolint:gosec // safe conversion, len() returns non-negative int
								chunkRef := int32(binary.BigEndian.Uint32(deltaData[offset : offset+4])) // #nosec G115 - safe conversion within bounds
								if chunkRef == -1 {
									modifiedChunks++
								}
							}
						}

						summary.ChunksModified = modifiedChunks
					}
				}
			}
		}
	} else {
		// For non-chunk delta, estimate modified chunks based on delta size
		if len(destContent) > 0 {
			chunkSize := d.options.ChunkSize
			if chunkSize <= 0 {
				chunkSize = DefaultChunkSize
			}

			// Rough estimate of affected chunks based on delta size ratio
			totalChunks := (len(targetContent) + chunkSize - 1) / chunkSize
			summary.ChunksModified = int(math.Ceil(float64(totalChunks) * deltaRatio))
		} else {
			// If no destination content, all chunks are modified
			chunkSize := d.options.ChunkSize
			if chunkSize <= 0 {
				chunkSize = DefaultChunkSize
			}
			summary.ChunksModified = (len(targetContent) + chunkSize - 1) / chunkSize
		}
	}

	d.logger.WithFields(map[string]interface{}{
		"source_size":     len(targetContent),
		"delta_size":      len(delta),
		"savings_percent": savingsPercent,
		"format":          deltaFormat,
		"chunks_modified": summary.ChunksModified,
	}).Info("Created delta for transfer optimization")

	// Update duration
	summary.Duration = time.Since(startTime)

	return summary, true, nil
}

// NewDeltaGenerator creates a new delta generator with the given options
func NewDeltaGenerator(opts DeltaOptions, logger log.Logger) *DeltaGenerator {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	// Normalize options
	if opts.ChunkSize <= 0 {
		opts.ChunkSize = DefaultDeltaOptions().ChunkSize
	}
	if opts.DeltaFormat == "" {
		opts.DeltaFormat = DefaultDeltaOptions().DeltaFormat
	}
	if opts.MaxDeltaRatio <= 0 {
		opts.MaxDeltaRatio = DefaultDeltaOptions().MaxDeltaRatio
	}

	return &DeltaGenerator{
		options: opts,
		logger:  logger,
		hasher:  sha256.New(),
	}
}

// ManifestEntry represents an entry in a delta manifest
type ManifestEntry struct {
	Path      string            `json:"path"`
	Size      int64             `json:"size"`
	Checksum  string            `json:"checksum"`
	Chunks    []ChunkInfo       `json:"chunks,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// ChunkInfo contains information about a chunk in a file
type ChunkInfo struct {
	Offset   int64  `json:"offset"`
	Size     int    `json:"size"`
	Checksum string `json:"checksum"`
}

// DeltaManifest describes the differences between source and destination
type DeltaManifest struct {
	SourceRepo      string          `json:"source_repo"`
	DestRepo        string          `json:"dest_repo"`
	Entries         []ManifestEntry `json:"entries"`
	CreatedAt       time.Time       `json:"created_at"`
	FormatVersion   string          `json:"format_version"`
	TotalSize       int64           `json:"total_size"`
	TotalDeltaSize  int64           `json:"total_delta_size"`
	CompressionType string          `json:"compression_type,omitempty"`
}

// GenerateManifest creates a delta manifest between source and destination
func (g *DeltaGenerator) GenerateManifest(ctx context.Context, sourceClient, destClient interfaces.RegistryClient, sourceRepo, destRepo string) (*DeltaManifest, error) {
	if sourceClient == nil {
		return nil, errors.InvalidInputf("source client cannot be nil")
	}
	if destClient == nil {
		return nil, errors.InvalidInputf("destination client cannot be nil")
	}
	if sourceRepo == "" {
		return nil, errors.InvalidInputf("source repository cannot be empty")
	}
	if destRepo == "" {
		return nil, errors.InvalidInputf("destination repository cannot be empty")
	}

	// Get source and destination repositories
	srcRepo, err := sourceClient.GetRepository(ctx, sourceRepo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get source repository")
	}

	_, err = destClient.GetRepository(ctx, destRepo)
	if err != nil {
		// If the destination doesn't exist yet, that's okay - empty manifest
		if errors.Is(err, errors.ErrNotFound) {
			g.logger.WithFields(map[string]interface{}{
				"destination": destRepo,
			}).Info("Destination repository doesn't exist yet")
			// Create empty manifest with just the source items
			return g.createEmptyManifest(srcRepo, destRepo)
		}
		return nil, errors.Wrap(err, "failed to get destination repository")
	}

	// Create manifest
	manifest := &DeltaManifest{
		SourceRepo:    sourceRepo,
		DestRepo:      destRepo,
		CreatedAt:     time.Now(),
		FormatVersion: "1.0",
		Entries:       []ManifestEntry{},
	}

	// This is just a sample implementation - real code would:
	// 1. List tags in source and destination
	// 2. For each tag in source that needs to be copied:
	//    a. Check if it exists in destination
	//    b. If not, mark for full copy
	//    c. If yes, compare layers and mark changed ones for delta copy
	// 3. Generate the manifest with this information

	// Return the manifest
	return manifest, nil
}

// createEmptyManifest creates a manifest with no existing destination content
func (g *DeltaGenerator) createEmptyManifest(srcRepo interfaces.Repository, destRepo string) (*DeltaManifest, error) {
	manifest := &DeltaManifest{
		SourceRepo:    srcRepo.GetRepositoryName(),
		DestRepo:      destRepo,
		CreatedAt:     time.Now(),
		FormatVersion: "1.0",
		Entries:       []ManifestEntry{},
	}

	// In a real implementation, we would:
	// 1. List all tags in srcRepo
	// 2. Add entries for each tag that needs to be copied (full copy)

	return manifest, nil
}

// Serialize converts a delta manifest to JSON
func (m *DeltaManifest) Serialize() ([]byte, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize delta manifest")
	}
	return data, nil
}

// ParseManifest parses a JSON delta manifest
func ParseManifest(data []byte) (*DeltaManifest, error) {
	if len(data) == 0 {
		return nil, errors.InvalidInputf("manifest data cannot be empty")
	}

	var manifest DeltaManifest
	err := json.Unmarshal(data, &manifest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse delta manifest")
	}

	// Validate manifest
	if manifest.SourceRepo == "" {
		return nil, errors.InvalidInputf("manifest missing source repository")
	}
	if manifest.DestRepo == "" {
		return nil, errors.InvalidInputf("manifest missing destination repository")
	}

	return &manifest, nil
}

// Constants for delta formats
const (
	BSDiffFormat      = "bsdiff" // Use bsdiff binary delta format (best compression)
	SimpleDeltaFormat = "simple" // Use simple format (faster but larger deltas)
	ChunkBasedFormat  = "chunk"  // Chunked format for partial updates
	NoDeltaFormat     = "none"   // No delta, just direct copy

	// Buffer sizes
	DefaultChunkSize = 1024 * 1024 // 1MB default chunk size
	BufferSize       = 32 * 1024   // 32KB buffer for I/O
)

// DeltaHeader contains information about the delta format and contents
type DeltaHeader struct {
	Format       string // The format used (bsdiff, simple, chunk)
	SourceSize   uint32 // Size of the original source in bytes
	TargetSize   uint32 // Size of the resulting target in bytes
	DeltaSize    uint32 // Size of the delta data (excluding header)
	ChunkSize    uint32 // For chunked format, size of each chunk
	ChunkCount   uint32 // For chunked format, number of chunks
	SourceDigest string // SHA256 digest of the source
	TargetDigest string // SHA256 digest of the expected target
}

// CreateDelta creates a delta between source and target data using the specified format
func CreateDelta(source, target []byte, format string) ([]byte, error) {
	if len(source) == 0 {
		return nil, errors.InvalidInputf("source cannot be empty")
	}
	if len(target) == 0 {
		return nil, errors.InvalidInputf("target cannot be empty")
	}

	// Check if source and target are identical
	if bytes.Equal(source, target) {
		// Create a special "empty delta" that indicates no changes
		return createIdenticalDelta(source), nil
	}

	// Calculate source and target digests
	sourceDigest, err := CalculateDigest(source)
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate source digest")
	}

	targetDigest, err := CalculateDigest(target)
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate target digest")
	}

	// Implement different delta formats
	switch format {
	case BSDiffFormat:
		// Simplified bsdiff-like implementation
		var delta bytes.Buffer

		// Create delta header
		header := DeltaHeader{
			Format:       BSDiffFormat,
			SourceSize:   uint32(len(source)), // #nosec G115 - safe conversion, len() returns non-negative int
			TargetSize:   uint32(len(target)), // #nosec G115 - safe conversion, len() returns non-negative int
			SourceDigest: sourceDigest,
			TargetDigest: targetDigest,
		}

		// Serialize header
		headerBytes, err := json.Marshal(header)
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialize delta header")
		}

		// Write header size as uint32 (4 bytes)
		headerSize := uint32(len(headerBytes)) // #nosec G115 - safe conversion, len() returns non-negative int
		if err := binary.Write(&delta, binary.BigEndian, headerSize); err != nil {
			return nil, errors.Wrap(err, "failed to write header size")
		}

		// Write header
		delta.Write(headerBytes)

		// Find common prefix
		commonPrefixLen := 0
		minLen := len(source)
		if len(target) < minLen {
			minLen = len(target)
		}

		for i := 0; i < minLen; i++ {
			if source[i] == target[i] {
				commonPrefixLen++
			} else {
				break
			}
		}

		// Find common suffix for the remaining parts
		commonSuffixLen := 0
		for i := 1; i <= minLen-commonPrefixLen; i++ {
			if source[len(source)-i] == target[len(target)-i] {
				commonSuffixLen++
			} else {
				break
			}
		}

		// Create a control structure:
		// - Prefix length (uint32)
		// - Suffix length (uint32)
		// - Middle data (the differing part)
		if err := binary.Write(&delta, binary.BigEndian, uint32(commonPrefixLen)); err != nil { // #nosec G115 - safe conversion, commonPrefixLen is bounded
			return nil, errors.Wrap(err, "failed to write prefix length")
		}
		if err := binary.Write(&delta, binary.BigEndian, uint32(commonSuffixLen)); err != nil { // #nosec G115 - safe conversion, commonSuffixLen is bounded
			return nil, errors.Wrap(err, "failed to write suffix length")
		}

		// Write the different middle section
		middleStart := commonPrefixLen
		middleEnd := len(target) - commonSuffixLen
		diffData := make([]byte, 0)
		if middleStart < middleEnd {
			diffData = target[middleStart:middleEnd]
			delta.Write(diffData)
		}

		// Update header with actual delta size
		header.DeltaSize = uint32(8 + len(diffData)) // 8 bytes for the prefix/suffix lengths + diff data

		return delta.Bytes(), nil

	case SimpleDeltaFormat:
		// Simple delta format - good for small files or when bsdiff is overkill
		var delta bytes.Buffer

		// Create delta header
		header := DeltaHeader{
			Format:       SimpleDeltaFormat,
			SourceSize:   uint32(len(source)),
			TargetSize:   uint32(len(target)),
			SourceDigest: sourceDigest,
			TargetDigest: targetDigest,
		}

		// Serialize header
		headerBytes, err := json.Marshal(header)
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialize delta header")
		}

		// Write header size as uint32 (4 bytes)
		headerSize := uint32(len(headerBytes)) // #nosec G115 - safe conversion, len() returns non-negative int
		if err := binary.Write(&delta, binary.BigEndian, headerSize); err != nil {
			return nil, errors.Wrap(err, "failed to write header size")
		}

		// Write header
		delta.Write(headerBytes)

		// Find common prefix
		commonPrefixLen := 0
		minLen := len(source)
		if len(target) < minLen {
			minLen = len(target)
		}

		for i := 0; i < minLen; i++ {
			if source[i] == target[i] {
				commonPrefixLen++
			} else {
				break
			}
		}

		// Find common suffix for the remaining parts
		suffixLen := 0
		for i := 1; i <= minLen-commonPrefixLen; i++ {
			if source[len(source)-i] == target[len(target)-i] {
				suffixLen++
			} else {
				break
			}
		}

		// Write the common prefix length as uint32
		if err := binary.Write(&delta, binary.BigEndian, uint32(commonPrefixLen)); err != nil { // #nosec G115 - safe conversion, commonPrefixLen is bounded
			return nil, errors.Wrap(err, "failed to write prefix length")
		}

		// Write the common suffix length as uint32
		if err := binary.Write(&delta, binary.BigEndian, uint32(suffixLen)); err != nil {
			return nil, errors.Wrap(err, "failed to write suffix length")
		}

		// Write only the middle different part
		middleStart := commonPrefixLen
		middleEnd := len(target) - suffixLen

		if middleStart < middleEnd {
			delta.Write(target[middleStart:middleEnd])
		}

		// Update header with delta size
		header.DeltaSize = uint32(delta.Len() - int(headerSize) - 4)

		return delta.Bytes(), nil

	case ChunkBasedFormat:
		// Chunk-based format - useful for very large files that need partial updates
		var delta bytes.Buffer
		chunkSize := DefaultChunkSize

		// Split source into chunks
		sourceChunks, err := ChunkData(source, chunkSize)
		if err != nil {
			return nil, errors.Wrap(err, "failed to chunk source data")
		}

		// Split target into chunks
		targetChunks, err := ChunkData(target, chunkSize)
		if err != nil {
			return nil, errors.Wrap(err, "failed to chunk target data")
		}

		// Calculate checksums for source chunks
		sourceChecksums := make([]string, len(sourceChunks))
		for i, chunk := range sourceChunks {
			checksum, _ := CalculateDigest(chunk)
			sourceChecksums[i] = checksum
		}

		// Find matching chunks
		chunkMatches := make([]int, len(targetChunks)) // -1 means no match
		for i := range chunkMatches {
			chunkMatches[i] = -1 // Initialize with no match
		}

		for i, targetChunk := range targetChunks {
			targetChecksum, _ := CalculateDigest(targetChunk)

			// Look for matching chunk in source
			for j, sourceChecksum := range sourceChecksums {
				if targetChecksum == sourceChecksum && bytes.Equal(targetChunk, sourceChunks[j]) {
					chunkMatches[i] = j
					break
				}
			}
		}

		// Create delta header
		header := DeltaHeader{
			Format:       ChunkBasedFormat,
			SourceSize:   uint32(len(source)),
			TargetSize:   uint32(len(target)),
			ChunkSize:    uint32(chunkSize),
			ChunkCount:   uint32(len(targetChunks)),
			SourceDigest: sourceDigest,
			TargetDigest: targetDigest,
		}

		// Serialize header
		headerBytes, err := json.Marshal(header)
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialize delta header")
		}

		// Write header size as uint32 (4 bytes)
		headerSize := uint32(len(headerBytes)) // #nosec G115 - safe conversion, len() returns non-negative int
		if err := binary.Write(&delta, binary.BigEndian, headerSize); err != nil {
			return nil, errors.Wrap(err, "failed to write header size")
		}

		// Write header
		delta.Write(headerBytes)

		// Write chunk match map - for each target chunk, either:
		// - Source chunk index to copy from (≥0)
		// - -1 if no match (needs new chunk data)
		for _, matchIdx := range chunkMatches {
			if err := binary.Write(&delta, binary.BigEndian, int32(matchIdx)); err != nil { // #nosec G115 - safe conversion, matchIdx within bounds
				return nil, errors.Wrap(err, "failed to write chunk match index")
			}
		}

		// Write new chunk data for non-matching chunks
		for i, matchIdx := range chunkMatches {
			if matchIdx == -1 && i < len(targetChunks) {
				// No match, write the chunk data
				delta.Write(targetChunks[i])
			}
		}

		// Update header with delta size
		header.DeltaSize = uint32(delta.Len() - int(headerSize) - 4)

		return delta.Bytes(), nil

	case NoDeltaFormat:
		// No delta format - just store the target directly with minimal overhead
		var delta bytes.Buffer

		// Create delta header
		header := DeltaHeader{
			Format:       NoDeltaFormat,
			SourceSize:   uint32(len(source)),
			TargetSize:   uint32(len(target)),
			DeltaSize:    uint32(len(target)),
			SourceDigest: sourceDigest,
			TargetDigest: targetDigest,
		}

		// Serialize header
		headerBytes, err := json.Marshal(header)
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialize delta header")
		}

		// Write header size as uint32 (4 bytes)
		headerSize := uint32(len(headerBytes)) // #nosec G115 - safe conversion, len() returns non-negative int
		if err := binary.Write(&delta, binary.BigEndian, headerSize); err != nil {
			return nil, errors.Wrap(err, "failed to write header size")
		}

		// Write header
		delta.Write(headerBytes)

		// Write the full target data
		delta.Write(target)

		return delta.Bytes(), nil

	default:
		return nil, errors.InvalidInputf("unsupported delta format: %s", format)
	}
}

// createIdenticalDelta creates a special delta indicating source and target are identical
func createIdenticalDelta(source []byte) []byte {
	var delta bytes.Buffer

	// Calculate source digest
	sourceDigest, _ := CalculateDigest(source)

	// Create delta header for identical files
	header := DeltaHeader{
		Format:       "identical",
		SourceSize:   uint32(len(source)),
		TargetSize:   uint32(len(source)),
		DeltaSize:    0, // No delta data needed
		SourceDigest: sourceDigest,
		TargetDigest: sourceDigest, // Same as source
	}

	// Serialize header
	headerBytes, _ := json.Marshal(header)

	// Write header size as uint32 (4 bytes)
	headerSize := uint32(len(headerBytes))
	if err := binary.Write(&delta, binary.BigEndian, headerSize); err != nil {
		// Return original delta on write error
		return delta.Bytes()
	}

	// Write header
	delta.Write(headerBytes)

	return delta.Bytes()
}

// ApplyDelta applies a delta to a source file to produce the destination file
func ApplyDelta(delta, source []byte, format string) ([]byte, error) {
	if len(delta) == 0 {
		return nil, errors.InvalidInputf("delta cannot be empty")
	}
	if len(source) == 0 {
		return nil, errors.InvalidInputf("source cannot be empty")
	}

	// Read and parse the delta header
	if len(delta) < 4 {
		return nil, errors.InvalidInputf("delta too short - missing header size")
	}

	// Read header size
	headerSize := binary.BigEndian.Uint32(delta[:4])
	if len(delta) < int(4+headerSize) {
		return nil, errors.InvalidInputf("delta too short - incomplete header")
	}

	// Parse header JSON
	var header DeltaHeader
	err := json.Unmarshal(delta[4:4+headerSize], &header)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse delta header")
	}

	// Verify source digest if available
	if header.SourceDigest != "" {
		sourceDigest, err := CalculateDigest(source)
		if err != nil {
			return nil, errors.Wrap(err, "failed to calculate source digest")
		}

		if sourceDigest != header.SourceDigest {
			return nil, errors.InvalidInputf("source digest mismatch: expected %s, got %s",
				header.SourceDigest, sourceDigest)
		}
	}

	// Start of delta data
	deltaStart := 4 + headerSize
	deltaData := delta[deltaStart:]

	// Special case for identical files
	if header.Format == "identical" {
		return source, nil
	}

	// The specific implementation depends on the delta format
	switch header.Format {
	case BSDiffFormat:
		// Simple bsdiff format that just stores prefix/suffix/middle
		if len(deltaData) < 8 {
			return nil, errors.InvalidInputf("invalid bsdiff format - missing prefix/suffix lengths")
		}

		// Read prefix and suffix lengths
		prefixLen := binary.BigEndian.Uint32(deltaData[0:4])
		suffixLen := binary.BigEndian.Uint32(deltaData[4:8])

		// Validate lengths
		if prefixLen > header.SourceSize || suffixLen > header.SourceSize {
			return nil, errors.InvalidInputf("invalid prefix/suffix lengths in bsdiff format")
		}

		// Calculate sizes
		middleLen := header.TargetSize - prefixLen - suffixLen
		middleData := deltaData[8:]

		if uint32(len(middleData)) < middleLen {
			return nil, errors.InvalidInputf("missing middle data in bsdiff format")
		}

		// Create the result buffer
		result := make([]byte, header.TargetSize)

		// Copy prefix from source
		if prefixLen > 0 {
			copy(result[:prefixLen], source[:prefixLen])
		}

		// Copy middle from delta
		if middleLen > 0 {
			copy(result[prefixLen:prefixLen+middleLen], middleData[:middleLen])
		}

		// Copy suffix from source
		if suffixLen > 0 {
			copy(result[prefixLen+middleLen:], source[len(source)-int(suffixLen):])
		}

		// Verify the result size
		if header.TargetSize > 0 && uint32(len(result)) != header.TargetSize {
			return nil, errors.InvalidInputf(
				"target size mismatch after bsdiff: expected %d bytes, got %d",
				header.TargetSize, len(result))
		}

		// Verify target digest if available
		if header.TargetDigest != "" {
			resultDigest, err := CalculateDigest(result)
			if err != nil {
				return nil, errors.Wrap(err, "failed to calculate result digest")
			}

			if resultDigest != header.TargetDigest {
				return nil, errors.InvalidInputf(
					"target digest mismatch after bsdiff: expected %s, got %s",
					header.TargetDigest, resultDigest)
			}
		}

		return result, nil

	case SimpleDeltaFormat:
		if len(deltaData) < 8 {
			return nil, errors.InvalidInputf("invalid simple delta format - missing prefix/suffix lengths")
		}

		// Read prefix and suffix lengths
		prefixLen := binary.BigEndian.Uint32(deltaData[0:4])
		suffixLen := binary.BigEndian.Uint32(deltaData[4:8])

		// Validate lengths
		if prefixLen > header.SourceSize || suffixLen > header.SourceSize {
			return nil, errors.InvalidInputf("invalid prefix/suffix lengths in simple delta")
		}

		// Calculate sizes
		middleLen := header.TargetSize - prefixLen - suffixLen
		middleData := deltaData[8:]

		if uint32(len(middleData)) < middleLen {
			return nil, errors.InvalidInputf("missing middle data in simple delta")
		}

		// Create the result buffer
		result := make([]byte, header.TargetSize)

		// Copy prefix from source
		if prefixLen > 0 {
			copy(result[:prefixLen], source[:prefixLen])
		}

		// Copy middle from delta
		if middleLen > 0 {
			copy(result[prefixLen:prefixLen+middleLen], middleData[:middleLen])
		}

		// Copy suffix from source
		if suffixLen > 0 {
			copy(result[prefixLen+middleLen:], source[len(source)-int(suffixLen):])
		}

		// Verify target digest if available
		if header.TargetDigest != "" {
			resultDigest, err := CalculateDigest(result)
			if err != nil {
				return nil, errors.Wrap(err, "failed to calculate result digest")
			}

			if resultDigest != header.TargetDigest {
				return nil, errors.InvalidInputf(
					"target digest mismatch after simple patch: expected %s, got %s",
					header.TargetDigest, resultDigest)
			}
		}

		return result, nil

	case ChunkBasedFormat:
		// Chunk-based delta application
		if header.ChunkSize == 0 || header.ChunkCount == 0 {
			return nil, errors.InvalidInputf("invalid chunk-based delta: missing chunk information")
		}

		// Size of the chunk map in bytes
		chunkMapSize := header.ChunkCount * 4 // Each chunk reference is an int32

		if uint32(len(deltaData)) < chunkMapSize {
			return nil, errors.InvalidInputf("invalid chunk-based delta: missing chunk map")
		}

		// Read chunk map
		chunkMap := make([]int32, header.ChunkCount)
		for i := uint32(0); i < header.ChunkCount; i++ {
			offset := i * 4
			chunkMap[i] = int32(binary.BigEndian.Uint32(deltaData[offset : offset+4])) // #nosec G115 - safe conversion, reading from validated data
		}

		// Get source chunks
		sourceChunks, err := ChunkData(source, int(header.ChunkSize))
		if err != nil {
			return nil, errors.Wrap(err, "failed to chunk source data")
		}

		// New chunk data starts after the chunk map
		newChunkData := deltaData[chunkMapSize:]
		newChunkOffset := 0

		// Create result buffer
		result := new(bytes.Buffer)
		result.Grow(int(header.TargetSize)) // Pre-allocate for efficiency

		// Apply chunk map to reconstruct target
		for _, chunkRef := range chunkMap {
			if chunkRef >= 0 && int(chunkRef) < len(sourceChunks) {
				// Copy from source chunk
				result.Write(sourceChunks[chunkRef])
			} else {
				// Copy new chunk from delta
				chunkSize := int(header.ChunkSize)
				if newChunkOffset+chunkSize > len(newChunkData) {
					chunkSize = len(newChunkData) - newChunkOffset
				}

				if chunkSize <= 0 {
					return nil, errors.InvalidInputf(
						"invalid chunk-based delta: missing chunk data at offset %d",
						newChunkOffset)
				}

				result.Write(newChunkData[newChunkOffset : newChunkOffset+chunkSize])
				newChunkOffset += chunkSize
			}
		}

		// Verify the result size
		resultBytes := result.Bytes()
		if uint32(len(resultBytes)) != header.TargetSize {
			return nil, errors.InvalidInputf(
				"target size mismatch after chunk reassembly: expected %d bytes, got %d",
				header.TargetSize, len(resultBytes))
		}

		// Verify target digest if available
		if header.TargetDigest != "" {
			resultDigest, err := CalculateDigest(resultBytes)
			if err != nil {
				return nil, errors.Wrap(err, "failed to calculate result digest")
			}

			if resultDigest != header.TargetDigest {
				return nil, errors.InvalidInputf(
					"target digest mismatch after chunk reassembly: expected %s, got %s",
					header.TargetDigest, resultDigest)
			}
		}

		return resultBytes, nil

	case NoDeltaFormat:
		// No delta format - just return the data directly
		result := deltaData

		// Verify the result size
		if header.TargetSize > 0 && uint32(len(result)) != header.TargetSize {
			return nil, errors.InvalidInputf(
				"target size mismatch in no-delta format: expected %d bytes, got %d",
				header.TargetSize, len(result))
		}

		// Verify target digest if available
		if header.TargetDigest != "" {
			resultDigest, err := CalculateDigest(result)
			if err != nil {
				return nil, errors.Wrap(err, "failed to calculate result digest")
			}

			if resultDigest != header.TargetDigest {
				return nil, errors.InvalidInputf(
					"target digest mismatch in no-delta format: expected %s, got %s",
					header.TargetDigest, resultDigest)
			}
		}

		return result, nil

	default:
		return nil, errors.InvalidInputf("unsupported delta format: %s", header.Format)
	}
}

// CalculateDigest calculates a digest for the given data
func CalculateDigest(data []byte) (string, error) {
	if len(data) == 0 {
		return "", errors.InvalidInputf("data cannot be empty")
	}

	h := sha256.New()
	_, err := h.Write(data)
	if err != nil {
		return "", errors.Wrap(err, "failed to calculate digest")
	}

	return fmt.Sprintf("sha256:%x", h.Sum(nil)), nil
}

// VerifyDigest checks if the given data matches the expected digest
func VerifyDigest(data []byte, expectedDigest string) error {
	if len(data) == 0 {
		return errors.InvalidInputf("data cannot be empty")
	}
	if expectedDigest == "" {
		return errors.InvalidInputf("expected digest cannot be empty")
	}

	actualDigest, err := CalculateDigest(data)
	if err != nil {
		return err
	}

	if actualDigest != expectedDigest {
		return errors.InvalidInputf("digest mismatch: expected %s, got %s", expectedDigest, actualDigest)
	}

	return nil
}

// ChunkData splits data into chunks of the specified size
func ChunkData(data []byte, chunkSize int) ([][]byte, error) {
	if len(data) == 0 {
		return nil, errors.InvalidInputf("data cannot be empty")
	}
	if chunkSize <= 0 {
		return nil, errors.InvalidInputf("chunk size must be positive")
	}

	var chunks [][]byte
	reader := bytes.NewReader(data)
	buffer := make([]byte, chunkSize)

	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			// Make a copy of the buffer to avoid reusing the same slice
			chunk := make([]byte, n)
			copy(chunk, buffer[:n])
			chunks = append(chunks, chunk)
		}
		if err == nil {
			continue
		}
		if err.Error() == "EOF" {
			break
		}
		return nil, errors.Wrap(err, "failed to read data for chunking")
	}

	return chunks, nil
}
