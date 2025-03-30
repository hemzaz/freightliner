package network

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"freightliner/pkg/client/common"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"hash"
	"time"
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
	logger  *log.Logger
	hasher  hash.Hash
}

// DeltaManager manages delta operations
type DeltaManager struct {
	logger  *log.Logger
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
func NewDeltaManager(logger *log.Logger, opts DeltaOptions) (*DeltaManager, error) {
	if logger == nil {
		logger = log.NewLogger(log.InfoLevel)
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
func (d *DeltaManager) OptimizeTransfer(sourceRepo, destRepo common.Repository, digest, mediaType string) (*DeltaSummary, bool, error) {
	startTime := time.Now()

	// Create a summary to return
	summary := &DeltaSummary{
		OriginalSize:   1000, // Placeholder value
		DeltaSize:      200,  // Placeholder value
		ChunksModified: 2,    // Placeholder value
		TransferSize:   220,  // Placeholder value
		SavingsPercent: 78.0, // Placeholder value
	}

	// Check if destination already has exactly the same content
	// In a real implementation, we would compare manifests or digests
	isIdentical := false

	if isIdentical {
		// Nothing to transfer
		summary.TransferSize = 0
		summary.SavingsPercent = 100.0
		summary.Duration = time.Since(startTime)
		return summary, true, nil
	}

	// For the test implementation, use some dummy values
	wasOptimized := true

	// Update duration
	summary.Duration = time.Since(startTime)

	return summary, wasOptimized, nil
}

// NewDeltaGenerator creates a new delta generator with the given options
func NewDeltaGenerator(opts DeltaOptions, logger *log.Logger) *DeltaGenerator {
	if logger == nil {
		logger = log.NewLogger(log.InfoLevel)
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
func (g *DeltaGenerator) GenerateManifest(ctx context.Context, sourceClient, destClient common.RegistryClient, sourceRepo, destRepo string) (*DeltaManifest, error) {
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
			g.logger.Info("Destination repository doesn't exist yet", map[string]interface{}{
				"destination": destRepo,
			})
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
func (g *DeltaGenerator) createEmptyManifest(srcRepo common.Repository, destRepo string) (*DeltaManifest, error) {
	manifest := &DeltaManifest{
		SourceRepo:    srcRepo.GetName(),
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
	BSDiffFormat      = "bsdiff"
	SimpleDeltaFormat = "simple"
)

// CreateDelta creates a delta between source and target data using the specified format
func CreateDelta(source, target []byte, format string) ([]byte, error) {
	if len(source) == 0 {
		return nil, errors.InvalidInputf("source cannot be empty")
	}
	if len(target) == 0 {
		return nil, errors.InvalidInputf("target cannot be empty")
	}

	// Implement different delta formats
	switch format {
	case BSDiffFormat:
		// Simplified implementation for testing
		// In a real implementation, we'd use a library like github.com/mendsley/bsdiff
		delta := []byte("BSDIFF40")
		// Append a very simple delta - just store the target
		delta = append(delta, target...)
		return delta, nil

	case SimpleDeltaFormat:
		// Very basic delta format for testing
		delta := []byte{1} // Version byte

		// Add 4 bytes for length of result
		size := len(target)
		delta = append(delta, byte(size>>24), byte(size>>16), byte(size>>8), byte(size))

		// For this simple implementation, just append the target data
		delta = append(delta, target...)
		return delta, nil

	default:
		return nil, errors.InvalidInputf("unsupported delta format: %s", format)
	}
}

// ApplyDelta applies a delta to a source file to produce the destination file
func ApplyDelta(delta, source []byte, format string) ([]byte, error) {
	if len(delta) == 0 {
		return nil, errors.InvalidInputf("delta cannot be empty")
	}
	if len(source) == 0 {
		return nil, errors.InvalidInputf("source cannot be empty")
	}

	// The specific implementation depends on the delta format
	switch format {
	case "bsdiff":
		// Basic placeholder implementation for bsdiff
		// In a real implementation, we would use a library like github.com/mendsley/bsdiff

		// For now, just simulate applying a delta by:
		// 1. Checking if the delta is a valid format (has a header)
		if len(delta) < 8 || string(delta[:8]) != "BSDIFF40" {
			return nil, errors.InvalidInputf("invalid bsdiff format")
		}

		// 2. For this simple implementation, we'll just concatenate the delta and source
		// to simulate a patched file. This isn't a real implementation!
		result := make([]byte, len(source)+len(delta)-8)
		copy(result, source)
		copy(result[len(source):], delta[8:])

		return result, nil
	case "simple":
		// Very basic delta format:
		// - First byte is a version (1)
		// - Next 4 bytes is a 32-bit length of the result
		// - Rest is a series of operations:
		//   - 0x00: copy from source (next byte is length)
		//   - 0x01: copy from delta (next byte is length)

		if len(delta) < 6 || delta[0] != 1 {
			return nil, errors.InvalidInputf("invalid simple delta format")
		}

		// In a real implementation, we would:
		// 1. Parse the header
		// 2. Apply the operations
		// 3. Verify the result

		// For now, return a placeholder implementation
		resultSize := int(delta[1])<<24 | int(delta[2])<<16 | int(delta[3])<<8 | int(delta[4])
		if resultSize <= 0 {
			resultSize = len(source) // Default to source size
		}

		result := make([]byte, resultSize)
		// Use min() to avoid index out of range
		sourceLen := len(source)
		if sourceLen > resultSize {
			sourceLen = resultSize
		}
		copy(result, source[:sourceLen])

		return result, nil
	default:
		return nil, errors.InvalidInputf("unsupported delta format: %s", format)
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

// splitIntoChunks splits data into chunks of the specified size and returns ChunkInfo slices
// This is used for delta calculations
func splitIntoChunks(data []byte, chunkSize int) []ChunkInfo {
	if len(data) == 0 {
		return []ChunkInfo{}
	}

	var chunks []ChunkInfo
	var offset int64

	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}

		chunkData := data[i:end]
		h := sha256.New()
		h.Write(chunkData)
		checksum := fmt.Sprintf("sha256:%x", h.Sum(nil))

		chunks = append(chunks, ChunkInfo{
			Offset:   offset,
			Size:     len(chunkData),
			Checksum: checksum,
		})

		offset += int64(len(chunkData))
	}

	return chunks
}

// compareChunks compares two sets of chunks and returns the indices of chunks that differ
// and the total number of chunks in the source
func compareChunks(source, dest []ChunkInfo) ([]int, int) {
	var changedIndices []int

	// Find common length to compare
	commonLength := len(source)
	if len(dest) < commonLength {
		commonLength = len(dest)
	}

	// Compare chunks that exist in both source and destination
	for i := 0; i < commonLength; i++ {
		if source[i].Checksum != dest[i].Checksum {
			changedIndices = append(changedIndices, i)
		}
	}

	// If source has more chunks than destination, mark the extras as changed
	for i := commonLength; i < len(source); i++ {
		changedIndices = append(changedIndices, i)
	}

	return changedIndices, len(source)
}
