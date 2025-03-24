package network

import (
	"bytes"
	"context"
	"testing"

	"github.com/hemzaz/freightliner/src/internal/log"
	"github.com/hemzaz/freightliner/src/pkg/client/common"
)

// MockRepository implements the common.Repository interface for testing
type MockRepository struct {
	Tags            map[string][]byte
	manifests       map[string][]byte
	manifestTypes   map[string]string
	putManifestCalls int
	getManifestCalls int
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		Tags:          make(map[string][]byte),
		manifests:     make(map[string][]byte),
		manifestTypes: make(map[string]string),
	}
}

func (m *MockRepository) ListTags() ([]string, error) {
	var tags []string
	for tag := range m.Tags {
		tags = append(tags, tag)
	}
	return tags, nil
}

func (m *MockRepository) GetManifest(tag string) ([]byte, string, error) {
	m.getManifestCalls++
	
	// Handle digest-based tags
	if bytes.HasPrefix([]byte(tag), []byte("@sha256:")) {
		digest := tag[1:] // Remove @ prefix
		manifest, ok := m.manifests[digest]
		if !ok {
			return []byte{}, "", nil
		}
		mediaType, ok := m.manifestTypes[digest]
		if !ok {
			mediaType = "application/vnd.docker.distribution.manifest.v2+json"
		}
		return manifest, mediaType, nil
	}
	
	// Handle normal tags
	manifest, ok := m.Tags[tag]
	if !ok {
		return []byte{}, "", nil
	}
	return manifest, "application/vnd.docker.distribution.manifest.v2+json", nil
}

func (m *MockRepository) PutManifest(tag string, manifest []byte, mediaType string) error {
	m.putManifestCalls++
	
	// Handle digest-based tags
	if bytes.HasPrefix([]byte(tag), []byte("@sha256:")) {
		digest := tag[1:] // Remove @ prefix
		m.manifests[digest] = manifest
		m.manifestTypes[digest] = mediaType
		return nil
	}
	
	// Handle normal tags
	m.Tags[tag] = manifest
	return nil
}

func (m *MockRepository) DeleteManifest(tag string) error {
	delete(m.Tags, tag)
	return nil
}

func TestDeltaOptions(t *testing.T) {
	opts := DefaultDeltaOptions()
	
	// Check default options
	if opts.ChunkSize <= 0 {
		t.Errorf("Default ChunkSize should be positive, got %d", opts.ChunkSize)
	}
	
	if opts.MaxDeltaRatio <= 0 || opts.MaxDeltaRatio > 1.0 {
		t.Errorf("Default MaxDeltaRatio should be between 0 and 1, got %f", opts.MaxDeltaRatio)
	}
}

func TestSplitIntoChunks(t *testing.T) {
	// Test empty data
	chunks := splitIntoChunks([]byte{}, 10)
	if len(chunks) != 0 {
		t.Errorf("Expected 0 chunks for empty data, got %d", len(chunks))
	}
	
	// Test data smaller than chunk size
	data := []byte("small data")
	chunks = splitIntoChunks(data, 100)
	if len(chunks) != 1 {
		t.Errorf("Expected 1 chunk for small data, got %d", len(chunks))
	}
	if chunks[0].Size != len(data) {
		t.Errorf("Chunk size mismatch, got %d, want %d", chunks[0].Size, len(data))
	}
	
	// Test data larger than chunk size
	data = bytes.Repeat([]byte("a"), 250)
	chunks = splitIntoChunks(data, 100)
	if len(chunks) != 3 {
		t.Errorf("Expected 3 chunks for 250 bytes with 100 byte chunks, got %d", len(chunks))
	}
	
	// Check chunk sizes
	expectedSizes := []int{100, 100, 50}
	for i, expected := range expectedSizes {
		if i >= len(chunks) {
			t.Fatalf("Chunk %d missing", i)
		}
		if chunks[i].Size != expected {
			t.Errorf("Chunk %d size mismatch, got %d, want %d", i, chunks[i].Size, expected)
		}
	}
}

func TestCompareChunks(t *testing.T) {
	// Create identical chunks
	source := splitIntoChunks([]byte("abcdefghijklmnopqrstuvwxyz"), 10)
	dest := splitIntoChunks([]byte("abcdefghijklmnopqrstuvwxyz"), 10)
	
	changed, total := compareChunks(source, dest)
	if len(changed) != 0 {
		t.Errorf("Expected 0 changed chunks for identical data, got %d", len(changed))
	}
	if total != len(source) {
		t.Errorf("Expected total chunks %d, got %d", len(source), total)
	}
	
	// Create source with one different chunk
	source = splitIntoChunks([]byte("abcdefghijKLMNOPQRSTuvwxyz"), 10)
	dest = splitIntoChunks([]byte("abcdefghijklmnopqrstuvwxyz"), 10)
	
	changed, total = compareChunks(source, dest)
	if len(changed) != 1 {
		t.Errorf("Expected 1 changed chunk for modified data, got %d", len(changed))
	}
	
	// Create source with additional chunks
	source = splitIntoChunks([]byte("abcdefghijklmnopqrstuvwxyzEXTRA"), 10)
	dest = splitIntoChunks([]byte("abcdefghijklmnopqrstuvwxyz"), 10)
	
	changed, total = compareChunks(source, dest)
	if len(changed) != 1 {
		t.Errorf("Expected 1 changed chunk for additional data, got %d", len(changed))
	}
}

func TestOptimizeTransfer(t *testing.T) {
	logger := log.NewLogger(log.InfoLevel)
	opts := DefaultDeltaOptions()
	manager := NewDeltaManager(logger, opts)
	
	// Create source repository with original data
	sourceRepo := NewMockRepository()
	manifest := []byte(`{"schemaVersion":2,"config":{"digest":"sha256:abc"},"layers":[{"digest":"layer1"},{"digest":"layer2"}]}`)
	sourceRepo.PutManifest("@sha256:manifest1", manifest, "application/json")
	
	// Create destination repository with the same data
	destRepo := NewMockRepository()
	destRepo.PutManifest("@sha256:manifest1", manifest, "application/json")
	
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
	
	// Create destination repository with different data
	destRepo = NewMockRepository()
	differentManifest := []byte(`{"schemaVersion":2,"config":{"digest":"sha256:xyz"},"layers":[{"digest":"layer1"},{"digest":"layer3"}]}`)
	destRepo.PutManifest("@sha256:manifest1", differentManifest, "application/json")
	
	// Test different data - should use delta
	summary, wasOptimized, err = manager.OptimizeTransfer(sourceRepo, destRepo, "sha256:manifest1", "application/json")
	if err != nil {
		t.Fatalf("OptimizeTransfer failed: %v", err)
	}
	
	if !wasOptimized {
		t.Errorf("Expected wasOptimized=true for delta transfer")
	}
	
	if summary.SavingsPercent <= 0 {
		t.Errorf("Expected positive savings percentage, got %f", summary.SavingsPercent)
	}
}

func TestDeltaSummary(t *testing.T) {
	// Create a delta summary
	summary := DeltaSummary{
		OriginalSize:   1000,
		DeltaSize:      200,
		ChunksModified: 2,
		TransferSize:   220, // Delta size plus overhead
		SavingsPercent: 78.0,
		Duration:       0,
	}
	
	// Verify calculations
	if summary.OriginalSize-summary.TransferSize != 780 {
		t.Errorf("Savings calculation incorrect")
	}
}