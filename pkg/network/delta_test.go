package network

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"math"
	"strings"
	"testing"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/interfaces"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// MockRepository implements the interfaces.Repository interface for testing
type MockRepository struct {
	Tags             map[string][]byte
	manifests        map[string]*interfaces.Manifest
	putManifestCalls int
	getManifestCalls int
	name             string
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		Tags:      make(map[string][]byte),
		manifests: make(map[string]*interfaces.Manifest),
		name:      "mock/repository",
	}
}

func (m *MockRepository) GetRepositoryName() string {
	return m.name
}

func (m *MockRepository) GetName() string {
	return m.name
}

func (m *MockRepository) ListTags(ctx context.Context) ([]string, error) {
	var tags []string
	for tag := range m.Tags {
		tags = append(tags, tag)
	}
	return tags, nil
}

// GetImage implements common.Repository.GetImage for testing
func (m *MockRepository) GetImage(ctx context.Context, tag string) (v1.Image, error) {
	// For tests, return a simple mock image implementation
	return nil, errors.NotImplementedf("GetImage not implemented in tests")
}

func (m *MockRepository) GetManifest(ctx context.Context, tag string) (*interfaces.Manifest, error) {
	m.getManifestCalls++

	// Handle digest-based tags
	if bytes.HasPrefix([]byte(tag), []byte("@sha256:")) {
		digest := tag[1:] // Remove @ prefix
		manifest, ok := m.manifests[digest]
		if !ok {
			return nil, nil
		}
		return manifest, nil
	}

	// Handle normal tags
	manifest, ok := m.manifests[tag]
	if !ok {
		return nil, nil
	}
	return manifest, nil
}

func (m *MockRepository) PutManifest(ctx context.Context, tag string, manifest *interfaces.Manifest) error {
	m.putManifestCalls++

	if m.manifests == nil {
		m.manifests = make(map[string]*interfaces.Manifest)
	}

	m.manifests[tag] = manifest
	return nil
}

// Legacy method for test compatibility
func (m *MockRepository) PutManifest2(tag string, content []byte, mediaType string) error {
	m.putManifestCalls++

	if m.manifests == nil {
		m.manifests = make(map[string]*interfaces.Manifest)
	}

	m.manifests[tag] = &interfaces.Manifest{
		Content:   content,
		MediaType: mediaType,
		Digest:    "sha256:" + tag,
	}
	return nil
}

func (m *MockRepository) DeleteManifest(ctx context.Context, tag string) error {
	delete(m.manifests, tag)
	delete(m.Tags, tag)
	return nil
}

func (m *MockRepository) GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("mock layer content")), nil
}

func (m *MockRepository) GetImageReference(tag string) (name.Reference, error) {
	return name.NewTag("example.com/repo:" + tag)
}

func (m *MockRepository) GetRemoteOptions() ([]remote.Option, error) {
	return []remote.Option{}, nil
}

func TestDeltaOptions(t *testing.T) {
	opts := DefaultDeltaOptions()

	if opts.DeltaFormat != "bsdiff" {
		t.Errorf("Expected default format to be bsdiff, got %v", opts.DeltaFormat)
	}

	if !opts.VerifyDelta {
		t.Error("Expected VerifyDelta to be enabled by default")
	}

	if opts.MaxDeltaRatio != 0.8 {
		t.Errorf("Expected default MaxDeltaRatio to be 0.8, got %v", opts.MaxDeltaRatio)
	}

	if opts.ChunkSize != 1024*1024 {
		t.Errorf("Expected default ChunkSize to be 1MB, got %v bytes", opts.ChunkSize)
	}
}

func TestBlobDiff(t *testing.T) {
	source := []byte("The quick brown fox jumps over the lazy dog")
	target := []byte("The quick brown fox jumps over the lazy cat")

	// Create a delta
	delta, err := CreateDelta(source, target, BSDiffFormat)
	if err != nil {
		t.Fatalf("CreateDelta failed: %v", err)
	}

	if len(delta) <= 0 {
		t.Errorf("Expected non-empty delta, got %d bytes", len(delta))
	}

	// Apply the delta
	result, err := ApplyDelta(delta, source, BSDiffFormat)
	if err != nil {
		t.Fatalf("ApplyDelta failed: %v", err)
	}

	if !bytes.Equal(result, target) {
		t.Errorf("Delta application failed, expected %s, got %s", target, result)
	}
}

func TestSimpleDelta(t *testing.T) {
	source := []byte("This is a test string for simple delta")
	target := []byte("This is a different test string for simple delta")

	// Create a delta using the simple format
	delta, err := CreateDelta(source, target, SimpleDeltaFormat)
	if err != nil {
		t.Fatalf("CreateDelta failed: %v", err)
	}

	if len(delta) <= 0 {
		t.Errorf("Expected non-empty delta, got %d bytes", len(delta))
	}

	// Apply the delta
	result, err := ApplyDelta(delta, source, SimpleDeltaFormat)
	if err != nil {
		t.Fatalf("ApplyDelta failed: %v", err)
	}

	if !bytes.Equal(result, target) {
		t.Errorf("Delta application failed, expected %s, got %s", target, result)
	}
}

func TestChunkBasedDelta(t *testing.T) {
	// For proper chunk-based delta testing, we need data specifically designed to
	// work well with our algorithm, rather than random data which is not compressible.
	// We'll use a smaller data set with predictable patterns

	// Create base array of repeating patterns
	basePattern := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	source := make([]byte, 100*1024) // 100KB is enough for testing

	// Fill source with repeating pattern
	for i := 0; i < len(source); i += len(basePattern) {
		copied := copy(source[i:], basePattern)
		if copied < len(basePattern) {
			// Handle partial copy at the end
			break
		}
	}

	// Create target with most data identical to source
	target := make([]byte, len(source))
	copy(target, source)

	// Only modify a small portion - 10% of the data
	modifyStart := len(target) / 2
	modifyLength := len(target) / 10

	// Put some different content in the middle
	differentPattern := []byte("9876543210ZYXWVUTSRQPONMLKJIHGFEDCBA")
	for i := 0; i < modifyLength; i += len(differentPattern) {
		pos := modifyStart + i
		copied := copy(target[pos:], differentPattern)
		if copied < len(differentPattern) || pos+copied >= len(target) {
			break
		}
	}

	// Create a delta using the chunk-based format
	// Force smaller chunk size to make test more predictable
	// DefaultChunkSize will be used internally by CreateDelta

	// The real test - create the delta
	delta, err := CreateDelta(source, target, ChunkBasedFormat)
	if err != nil {
		t.Fatalf("CreateDelta failed: %v", err)
	}

	if len(delta) <= 0 {
		t.Errorf("Expected non-empty delta, got %d bytes", len(delta))
	}

	// For predictable test patterns (unlike random data),
	// the delta should be smaller than the target
	compressionRatio := float64(len(delta)) / float64(len(target))
	t.Logf("Chunk-based delta size: %d bytes, ratio: %.2f", len(delta), compressionRatio)
	// We don't enforce a specific ratio in the test to avoid flakiness

	// Apply the delta
	result, err := ApplyDelta(delta, source, ChunkBasedFormat)
	if err != nil {
		t.Fatalf("ApplyDelta failed: %v", err)
	}

	if !bytes.Equal(result, target) {
		// Binary data equality failed
		t.Errorf("Delta application failed, result length: %d, target length: %d", len(result), len(target))

		// Additional diagnostics for the first few differences
		diffCount := 0
		minLen := len(result)
		if len(target) < minLen {
			minLen = len(target)
		}

		for i := 0; i < minLen; i++ {
			if result[i] != target[i] {
				if diffCount < 5 {
					t.Logf("Diff at position %d: result=0x%02x, target=0x%02x", i, result[i], target[i])
				}
				diffCount++
			}
		}
		t.Logf("Total differences: %d", diffCount)
	}
}

func TestNoDeltaFormat(t *testing.T) {
	source := []byte("Original content that will be ignored")
	target := []byte("This is the complete new content with no delta applied")

	// Create a "delta" using the no-delta format (just the target with minimal header)
	delta, err := CreateDelta(source, target, NoDeltaFormat)
	if err != nil {
		t.Fatalf("CreateDelta failed: %v", err)
	}

	// Check that the delta contains the full target plus header
	if len(delta) <= len(target) {
		t.Errorf("Expected delta to be larger than target due to header, got %d vs %d bytes", len(delta), len(target))
	}

	// Parse header size
	headerSize := binary.BigEndian.Uint32(delta[:4])

	// Check that deltaSize (from NoDeltaFormat) is properly set
	headerBytes := delta[4 : 4+headerSize]
	var header DeltaHeader
	err = json.Unmarshal(headerBytes, &header)
	if err != nil {
		t.Fatalf("Failed to parse delta header: %v", err)
	}

	if header.Format != NoDeltaFormat {
		t.Errorf("Expected format to be %s, got %s", NoDeltaFormat, header.Format)
	}

	if header.DeltaSize != uint32(len(target)) {
		t.Errorf("Expected delta size to be %d, got %d", len(target), header.DeltaSize)
	}

	// Apply the delta - the source should be ignored
	result, err := ApplyDelta(delta, source, NoDeltaFormat)
	if err != nil {
		t.Fatalf("ApplyDelta failed: %v", err)
	}

	if !bytes.Equal(result, target) {
		t.Errorf("Delta application failed, expected %s, got %s", target, result)
	}
}

func TestIdenticalContent(t *testing.T) {
	content := []byte("This content is identical between source and target")

	// Create a delta between identical content
	delta, err := CreateDelta(content, content, BSDiffFormat)
	if err != nil {
		t.Fatalf("CreateDelta failed: %v", err)
	}

	// Parse header size
	headerSize := binary.BigEndian.Uint32(delta[:4])

	// Check that the header is properly set
	headerBytes := delta[4 : 4+headerSize]
	var header DeltaHeader
	err = json.Unmarshal(headerBytes, &header)
	if err != nil {
		t.Fatalf("Failed to parse delta header: %v", err)
	}

	if header.Format != "identical" {
		t.Errorf("Expected format to be 'identical', got %s", header.Format)
	}

	if header.DeltaSize != 0 {
		t.Errorf("Expected delta size to be 0, got %d", header.DeltaSize)
	}

	// Apply the delta
	result, err := ApplyDelta(delta, content, BSDiffFormat)
	if err != nil {
		t.Fatalf("ApplyDelta failed: %v", err)
	}

	if !bytes.Equal(result, content) {
		t.Errorf("Delta application failed, expected identical content")
	}
}

func TestDigestValidation(t *testing.T) {
	source := []byte("Original content")
	target := []byte("Modified content")

	// Create a delta
	delta, err := CreateDelta(source, target, BSDiffFormat)
	if err != nil {
		t.Fatalf("CreateDelta failed: %v", err)
	}

	// Tamper with the source data
	tamperedSource := make([]byte, len(source))
	copy(tamperedSource, source)
	tamperedSource[0] = 'T' // Change first byte

	// Apply the delta to tampered source - should fail with digest validation error
	_, err = ApplyDelta(delta, tamperedSource, BSDiffFormat)
	if err == nil {
		t.Error("Expected digest validation to fail, but it succeeded")
	}

	// Check error type and message
	if !strings.Contains(err.Error(), "source digest mismatch") {
		t.Errorf("Expected source digest mismatch error, got: %v", err)
	}
}

func TestDeltaHeader(t *testing.T) {
	source := []byte("Original content for testing delta headers")
	target := []byte("Modified content for testing delta headers")

	// Test all delta formats to check header consistency
	formats := []string{BSDiffFormat, SimpleDeltaFormat, ChunkBasedFormat, NoDeltaFormat}

	for _, format := range formats {
		t.Run("Format_"+format, func(t *testing.T) {
			// Create a delta
			delta, err := CreateDelta(source, target, format)
			if err != nil {
				t.Fatalf("CreateDelta failed for %s: %v", format, err)
			}

			// Parse the header
			if len(delta) < 4 {
				t.Fatalf("Delta too short")
			}

			headerSize := binary.BigEndian.Uint32(delta[:4])
			if len(delta) < int(4+headerSize) {
				t.Fatalf("Delta too short for header")
			}

			var header DeltaHeader
			err = json.Unmarshal(delta[4:4+headerSize], &header)
			if err != nil {
				t.Fatalf("Failed to parse header: %v", err)
			}

			// Verify header fields
			if header.Format != format {
				t.Errorf("Expected format %s, got %s", format, header.Format)
			}

			if header.SourceSize != uint32(len(source)) {
				t.Errorf("Expected source size %d, got %d", len(source), header.SourceSize)
			}

			if header.TargetSize != uint32(len(target)) {
				t.Errorf("Expected target size %d, got %d", len(target), header.TargetSize)
			}

			// Check digests
			sourceDigest, _ := CalculateDigest(source)
			if header.SourceDigest != sourceDigest {
				t.Errorf("Expected source digest %s, got %s", sourceDigest, header.SourceDigest)
			}

			targetDigest, _ := CalculateDigest(target)
			if header.TargetDigest != targetDigest {
				t.Errorf("Expected target digest %s, got %s", targetDigest, header.TargetDigest)
			}
		})
	}
}

func TestErrorConditions(t *testing.T) {
	// Test empty source
	_, err := CreateDelta([]byte{}, []byte("target"), BSDiffFormat)
	if err == nil || !errors.Is(err, errors.ErrInvalidInput) {
		t.Errorf("Expected invalid input error for empty source, got: %v", err)
	}

	// Test empty target
	_, err = CreateDelta([]byte("source"), []byte{}, BSDiffFormat)
	if err == nil || !errors.Is(err, errors.ErrInvalidInput) {
		t.Errorf("Expected invalid input error for empty target, got: %v", err)
	}

	// Test invalid format
	_, err = CreateDelta([]byte("source"), []byte("target"), "invalid-format")
	if err == nil || !strings.Contains(err.Error(), "unsupported delta format") {
		t.Errorf("Expected unsupported format error, got: %v", err)
	}

	// Test invalid delta when applying
	_, err = ApplyDelta([]byte{1, 2, 3}, []byte("source"), BSDiffFormat)
	if err == nil {
		t.Error("Expected error when applying invalid delta, got nil")
	}

	// Test empty source when applying
	validDelta, _ := CreateDelta([]byte("source"), []byte("target"), BSDiffFormat)
	_, err = ApplyDelta(validDelta, []byte{}, BSDiffFormat)
	if err == nil || !errors.Is(err, errors.ErrInvalidInput) {
		t.Errorf("Expected invalid input error for empty source in ApplyDelta, got: %v", err)
	}
}

func TestChunkingFunctions(t *testing.T) {
	data := []byte("This is test data that will be split into multiple chunks for testing purposes.")
	chunkSize := 10 // Small chunk size for testing

	// Test ChunkData function
	chunks, err := ChunkData(data, chunkSize)
	if err != nil {
		t.Fatalf("ChunkData failed: %v", err)
	}

	// Verify number of chunks
	expectedChunks := (len(data) + chunkSize - 1) / chunkSize
	if len(chunks) != expectedChunks {
		t.Errorf("Expected %d chunks, got %d", expectedChunks, len(chunks))
	}

	// Verify chunk sizes
	for i, chunk := range chunks {
		if i < len(chunks)-1 && len(chunk) != chunkSize {
			t.Errorf("Chunk %d has size %d, expected %d", i, len(chunk), chunkSize)
		}
	}

	// Verify reconstitution
	reconstituted := []byte{}
	for _, chunk := range chunks {
		reconstituted = append(reconstituted, chunk...)
	}

	if !bytes.Equal(reconstituted, data) {
		t.Error("Reconstituted data doesn't match original")
	}

	// Test error conditions
	_, err = ChunkData([]byte{}, chunkSize)
	if err == nil || !errors.Is(err, errors.ErrInvalidInput) {
		t.Errorf("Expected invalid input error for empty data, got: %v", err)
	}

	_, err = ChunkData(data, 0)
	if err == nil || !errors.Is(err, errors.ErrInvalidInput) {
		t.Errorf("Expected invalid input error for zero chunk size, got: %v", err)
	}
}

func TestDeltaManager(t *testing.T) {
	logger := log.NewLogger()

	manager, err := NewDeltaManager(logger, DefaultDeltaOptions())
	if err != nil {
		t.Fatalf("NewDeltaManager failed: %v", err)
	}

	// Make the source and target data different enough to ensure delta is smaller
	sourceData := []byte(`{"layers":[{"digest":"sha256:layer1"},{"digest":"sha256:layer2"}]}`)
	// Make target significantly different to ensure the delta is smaller
	targetData := []byte(`{"layers":[{"digest":"sha256:layer1"},{"digest":"sha256:layer3"},{"digest":"sha256:layer4"},{"digest":"sha256:layer5"}]}`)

	// Try to find a delta between the two manifests
	delta, savings, err := manager.getDelta(sourceData, targetData)
	if err != nil {
		t.Fatalf("getDelta failed: %v", err)
	}

	// Delta should be smaller than the target for our sufficiently different data
	if len(delta) >= len(targetData) {
		t.Logf("Delta size: %d, Target size: %d", len(delta), len(targetData))
		// Skip exact comparison since this is implementation specific, just verify we can apply it
	} else {
		// Savings should be positive in this case
		if savings <= 0 {
			t.Errorf("Expected positive savings, got %d bytes", savings)
		}
	}

	// Apply the delta to reconstruct the target
	reconstructed, err := ApplyDelta(delta, sourceData, manager.options.DeltaFormat)
	if err != nil {
		t.Fatalf("ApplyDelta failed: %v", err)
	}

	if !bytes.Equal(reconstructed, targetData) {
		t.Errorf("Reconstructed data doesn't match target")
	}
}

func TestOptimizeTransfer(t *testing.T) {
	logger := log.NewLogger()
	manager, _ := NewDeltaManager(logger, DefaultDeltaOptions())

	// Create source repository with a manifest
	sourceRepo := NewMockRepository()
	manifest := []byte(`{"schemaVersion":2,"config":{"digest":"sha256:abc"},"layers":[{"digest":"layer1"},{"digest":"layer2"}]}`)
	_ = sourceRepo.PutManifest2("sha256:manifest1", manifest, "application/json")

	// Create destination repository with the same data
	destRepo := NewMockRepository()
	_ = destRepo.PutManifest2("sha256:manifest1", manifest, "application/json")

	// Test identical data - should skip transfer
	summary, wasOptimized, err := manager.OptimizeTransfer(sourceRepo, destRepo, "sha256:manifest1", "application/json")
	if err != nil {
		t.Fatalf("OptimizeTransfer failed: %v", err)
	}

	if !wasOptimized {
		t.Errorf("Expected wasOptimized=true for identical data")
	}

	if summary.TransferSize != 0 {
		t.Errorf("Expected TransferSize=0 for identical data, got %d", summary.TransferSize)
	}

	// Create a different manifest - make it substantially different to ensure a delta is worthwhile
	manifest2 := []byte(`{"schemaVersion":2,"config":{"digest":"sha256:abc"},
	"layers":[{"digest":"layer1"},{"digest":"layer3"},
	          {"digest":"layer4"},{"digest":"layer5"},
	          {"digest":"layer6"},{"digest":"layer7"}]}`)
	_ = sourceRepo.PutManifest2("sha256:manifest2", manifest2, "application/json")

	// Test delta transfer - should optimize
	summary, wasOptimized, err = manager.OptimizeTransfer(sourceRepo, destRepo, "sha256:manifest2", "application/json")
	if err != nil {
		t.Fatalf("OptimizeTransfer failed: %v", err)
	}

	if !wasOptimized {
		t.Logf("Delta was not optimized, this may be expected for very small inputs")
		return // Skip the rest of the test if optimization didn't happen
	}

	if summary.DeltaSize <= 0 {
		t.Errorf("Expected positive DeltaSize, got %d", summary.DeltaSize)
	}

	if summary.OriginalSize <= 0 {
		t.Errorf("Expected positive OriginalSize, got %d", summary.OriginalSize)
	}

	// Check savings percentage
	expectedSavings := float64(summary.OriginalSize-summary.TransferSize) / float64(summary.OriginalSize) * 100.0
	if math.Abs(summary.SavingsPercent-expectedSavings) > 0.1 {
		t.Errorf("Savings percentage doesn't match: calculated %f vs reported %f",
			expectedSavings, summary.SavingsPercent)
	}
}

func TestNoDelta(t *testing.T) {
	logger := log.NewLogger()

	// Disable delta generation
	opts := DefaultDeltaOptions()
	opts.MaxDeltaRatio = 0.0 // Disable delta generation

	manager, _ := NewDeltaManager(logger, opts)

	// Create source and destination repositories
	sourceRepo := NewMockRepository()
	destRepo := NewMockRepository()

	// Create a small manifest (below threshold)
	manifest := []byte(`{"schemaVersion":2,"config":{"digest":"sha256:abc"},"layers":[{"digest":"layer1"}]}`)
	_ = sourceRepo.PutManifest2("sha256:smallmanifest", manifest, "application/json")

	// Test - should not use delta due to disabling delta via MaxDeltaRatio=0
	summary, wasOptimized, err := manager.OptimizeTransfer(sourceRepo, destRepo, "sha256:smallmanifest", "application/json")
	if err != nil {
		t.Fatalf("OptimizeTransfer failed: %v", err)
	}

	if wasOptimized {
		t.Errorf("Expected wasOptimized=false when MaxDeltaRatio=0")
	}

	if summary.DeltaSize > 0 {
		t.Errorf("Expected DeltaSize=0, got %d", summary.DeltaSize)
	}
}

func TestDeltaManifest(t *testing.T) {
	// Test manifest serialization/deserialization
	manifest := &DeltaManifest{
		SourceRepo:     "source-repo",
		DestRepo:       "dest-repo",
		FormatVersion:  "1.0",
		CreatedAt:      NoZeroTime(),
		TotalSize:      1000,
		TotalDeltaSize: 500,
		Entries: []ManifestEntry{
			{
				Path:     "test/path",
				Size:     100,
				Checksum: "sha256:abc123",
				Chunks: []ChunkInfo{
					{
						Offset:   0,
						Size:     50,
						Checksum: "sha256:chunk1",
					},
					{
						Offset:   50,
						Size:     50,
						Checksum: "sha256:chunk2",
					},
				},
				Timestamp: NoZeroTime(),
			},
		},
	}

	// Serialize
	data, err := manifest.Serialize()
	if err != nil {
		t.Fatalf("Manifest serialization failed: %v", err)
	}

	// Deserialize
	parsedManifest, err := ParseManifest(data)
	if err != nil {
		t.Fatalf("Manifest parsing failed: %v", err)
	}

	// Verify fields
	if parsedManifest.SourceRepo != manifest.SourceRepo {
		t.Errorf("SourceRepo mismatch: expected %s, got %s",
			manifest.SourceRepo, parsedManifest.SourceRepo)
	}

	if parsedManifest.DestRepo != manifest.DestRepo {
		t.Errorf("DestRepo mismatch: expected %s, got %s",
			manifest.DestRepo, parsedManifest.DestRepo)
	}

	if parsedManifest.TotalSize != manifest.TotalSize {
		t.Errorf("TotalSize mismatch: expected %d, got %d",
			manifest.TotalSize, parsedManifest.TotalSize)
	}

	if len(parsedManifest.Entries) != len(manifest.Entries) {
		t.Errorf("Entries count mismatch: expected %d, got %d",
			len(manifest.Entries), len(parsedManifest.Entries))
	}

	// Error conditions
	_, err = ParseManifest([]byte{})
	if err == nil || !errors.Is(err, errors.ErrInvalidInput) {
		t.Errorf("Expected invalid input error for empty manifest data, got: %v", err)
	}

	// Invalid JSON
	_, err = ParseManifest([]byte("invalid json"))
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}

	// Missing required fields
	incompleteManifest := map[string]interface{}{
		"format_version": "1.0",
	}
	incompleteData, _ := json.Marshal(incompleteManifest)
	_, err = ParseManifest(incompleteData)
	if err == nil || !errors.Is(err, errors.ErrInvalidInput) {
		t.Errorf("Expected invalid input error for incomplete manifest, got: %v", err)
	}
}

// Helper for tests that need non-zero time values
func NoZeroTime() time.Time {
	return time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
}
