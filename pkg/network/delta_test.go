package network

import (
	"bytes"
	"context"
	"freightliner/pkg/client/common"
	"freightliner/pkg/helper/log"
	"io"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// MockRepository implements the common.Repository interface for testing
type MockRepository struct {
	Tags             map[string][]byte
	manifests        map[string]*common.Manifest
	putManifestCalls int
	getManifestCalls int
	name             string
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		Tags:      make(map[string][]byte),
		manifests: make(map[string]*common.Manifest),
		name:      "mock/repository",
	}
}

func (m *MockRepository) GetRepositoryName() string {
	return m.name
}

func (m *MockRepository) GetName() string {
	return m.name
}

func (m *MockRepository) ListTags() ([]string, error) {
	var tags []string
	for tag := range m.Tags {
		tags = append(tags, tag)
	}
	return tags, nil
}

func (m *MockRepository) GetManifest(ctx context.Context, tag string) (*common.Manifest, error) {
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

func (m *MockRepository) PutManifest(ctx context.Context, tag string, manifest *common.Manifest) error {
	m.putManifestCalls++

	if m.manifests == nil {
		m.manifests = make(map[string]*common.Manifest)
	}

	m.manifests[tag] = manifest
	return nil
}

// Legacy method for test compatibility
func (m *MockRepository) PutManifest2(tag string, content []byte, mediaType string) error {
	m.putManifestCalls++

	if m.manifests == nil {
		m.manifests = make(map[string]*common.Manifest)
	}

	m.manifests[tag] = &common.Manifest{
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
	result, err := ApplyDelta(source, delta, BSDiffFormat)
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
	result, err := ApplyDelta(source, delta, SimpleDeltaFormat)
	if err != nil {
		t.Fatalf("ApplyDelta failed: %v", err)
	}

	if !bytes.Equal(result, target) {
		t.Errorf("Delta application failed, expected %s, got %s", target, result)
	}
}

func TestDeltaManager(t *testing.T) {
	logger := log.NewLogger(log.InfoLevel)

	manager, err := NewDeltaManager(DefaultDeltaOptions(), logger)
	if err != nil {
		t.Fatalf("NewDeltaManager failed: %v", err)
	}

	sourceData := []byte(`{"layers":[{"digest":"sha256:layer1"},{"digest":"sha256:layer2"}]}`)
	targetData := []byte(`{"layers":[{"digest":"sha256:layer1"},{"digest":"sha256:layer3"}]}`)

	// Try to find a delta between the two manifests
	delta, savings, err := manager.getDelta(sourceData, targetData)
	if err != nil {
		t.Fatalf("getDelta failed: %v", err)
	}

	// Delta should be smaller than the target
	if len(delta) >= len(targetData) {
		t.Errorf("Expected delta to be smaller than target, got %d vs %d bytes", len(delta), len(targetData))
	}

	// Savings should be positive
	if savings <= 0 {
		t.Errorf("Expected positive savings, got %d bytes", savings)
	}

	// Apply the delta to reconstruct the target
	reconstructed, err := ApplyDelta(sourceData, delta, manager.options.Format)
	if err != nil {
		t.Fatalf("ApplyDelta failed: %v", err)
	}

	if !bytes.Equal(reconstructed, targetData) {
		t.Errorf("Reconstructed data doesn't match target")
	}
}

func TestOptimizeTransfer(t *testing.T) {
	logger := log.NewLogger(log.InfoLevel)
	manager, _ := NewDeltaManager(DefaultDeltaOptions(), logger)

	// Create source repository with a manifest
	sourceRepo := NewMockRepository()
	manifest := []byte(`{"schemaVersion":2,"config":{"digest":"sha256:abc"},"layers":[{"digest":"layer1"},{"digest":"layer2"}]}`)
	sourceRepo.PutManifest2("@sha256:manifest1", manifest, "application/json")

	// Create destination repository with the same data
	destRepo := NewMockRepository()
	destRepo.PutManifest2("@sha256:manifest1", manifest, "application/json")

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

	// Create a different manifest
	manifest2 := []byte(`{"schemaVersion":2,"config":{"digest":"sha256:abc"},"layers":[{"digest":"layer1"},{"digest":"layer3"}]}`)
	sourceRepo.PutManifest2("@sha256:manifest2", manifest2, "application/json")

	// Test delta transfer - should optimize
	summary, wasOptimized, err = manager.OptimizeTransfer(sourceRepo, destRepo, "sha256:manifest2", "application/json")
	if err != nil {
		t.Fatalf("OptimizeTransfer failed: %v", err)
	}

	if !wasOptimized {
		t.Errorf("Expected wasOptimized=true for delta optimization")
	}

	if summary.DeltaSize <= 0 {
		t.Errorf("Expected positive DeltaSize, got %d", summary.DeltaSize)
	}

	if summary.OriginalSize <= 0 {
		t.Errorf("Expected positive OriginalSize, got %d", summary.OriginalSize)
	}
}

func TestNoDelta(t *testing.T) {
	logger := log.NewLogger(log.InfoLevel)

	// Disable delta generation
	opts := DefaultDeltaOptions()
	opts.MinimumSize = 1000000 // Set a high threshold

	manager, _ := NewDeltaManager(opts, logger)

	// Create source and destination repositories
	sourceRepo := NewMockRepository()
	destRepo := NewMockRepository()

	// Create a small manifest (below threshold)
	manifest := []byte(`{"schemaVersion":2,"config":{"digest":"sha256:abc"},"layers":[{"digest":"layer1"}]}`)
	sourceRepo.PutManifest2("@sha256:smallmanifest", manifest, "application/json")

	// Test - should not use delta due to size
	summary, wasOptimized, err := manager.OptimizeTransfer(sourceRepo, destRepo, "sha256:smallmanifest", "application/json")
	if err != nil {
		t.Fatalf("OptimizeTransfer failed: %v", err)
	}

	if wasOptimized {
		t.Errorf("Expected wasOptimized=false for below threshold size")
	}

	if summary.DeltaSize > 0 {
		t.Errorf("Expected DeltaSize=0, got %d", summary.DeltaSize)
	}
}
