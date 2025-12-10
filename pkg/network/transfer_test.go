package network

import (
	"context"
	"io"
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

// TransferMockRepository for testing TransferManager
type TransferMockRepository struct {
	name string
	tags map[string][]byte
}

func NewTransferMockRepository(name string) *TransferMockRepository {
	return &TransferMockRepository{
		name: name,
		tags: make(map[string][]byte),
	}
}

func (m *TransferMockRepository) GetRepositoryName() string {
	return m.name
}

func (m *TransferMockRepository) GetName() string {
	return m.name
}

func (m *TransferMockRepository) ListTags(ctx context.Context) ([]string, error) {
	tags := make([]string, 0, len(m.tags))
	for t := range m.tags {
		tags = append(tags, t)
	}
	return tags, nil
}

// GetImage implements common.Repository.GetImage for testing
func (m *TransferMockRepository) GetImage(ctx context.Context, tag string) (v1.Image, error) {
	// For tests, return a simple mock image implementation
	return nil, errors.NotImplementedf("GetImage not implemented in tests")
}

func (m *TransferMockRepository) GetManifest(ctx context.Context, tag string) (*interfaces.Manifest, error) {
	manifest, ok := m.tags[tag]
	if !ok {
		return nil, nil
	}
	return &interfaces.Manifest{
		Content:   manifest,
		MediaType: "application/vnd.docker.distribution.manifest.v2+json",
		Digest:    "sha256:" + tag,
	}, nil
}

func (m *TransferMockRepository) PutManifest(ctx context.Context, tag string, manifest *interfaces.Manifest) error {
	m.tags[tag] = manifest.Content
	return nil
}

func (m *TransferMockRepository) DeleteManifest(ctx context.Context, tag string) error {
	delete(m.tags, tag)
	return nil
}

func (m *TransferMockRepository) GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("mock layer data")), nil
}

func (m *TransferMockRepository) GetImageReference(tag string) (name.Reference, error) {
	// Use correct local registry port for testing
	return name.NewTag("localhost:5100/" + m.name + ":" + tag)
}

func (m *TransferMockRepository) GetRemoteOptions() ([]remote.Option, error) {
	return []remote.Option{}, nil
}

func TestTransferManager(t *testing.T) {
	logger := log.NewLogger()
	opts := DefaultTransferOptions()

	// Create transfer manager
	manager, err := NewTransferManager(opts, logger)
	if err != nil {
		t.Fatalf("Failed to create TransferManager: %v", err)
	}

	// Check that the manager was initialized properly
	if manager.options.RetryAttempts <= 0 {
		t.Errorf("RetryAttempts should be positive, got %d", manager.options.RetryAttempts)
	}

	if manager.options.RetryInitialDelay <= 0 {
		t.Errorf("RetryInitialDelay should be positive, got %v", manager.options.RetryInitialDelay)
	}

	if manager.options.RetryMaxDelay <= 0 {
		t.Errorf("RetryMaxDelay should be positive, got %v", manager.options.RetryMaxDelay)
	}
}

func TestDefaultTransferOptions(t *testing.T) {
	opts := DefaultTransferOptions()

	// Verify reasonable defaults
	if !opts.EnableCompression {
		t.Errorf("EnableCompression should default to true")
	}

	if !opts.EnableDelta {
		t.Errorf("EnableDelta should default to true")
	}

	if opts.RetryAttempts < 1 {
		t.Errorf("RetryAttempts should be at least 1, got %d", opts.RetryAttempts)
	}

	if opts.RetryInitialDelay < time.Millisecond {
		t.Errorf("RetryInitialDelay should be at least 1ms, got %v", opts.RetryInitialDelay)
	}
}

func TestTransferBlob(t *testing.T) {
	logger := log.NewLogger()
	opts := DefaultTransferOptions()

	manager, err := NewTransferManager(opts, logger)
	if err != nil {
		t.Fatalf("Failed to create TransferManager: %v", err)
	}

	// Create source and destination repositories
	sourceRepo := NewTransferMockRepository("source/repo")
	destRepo := NewTransferMockRepository("dest/repo")

	// Test transfer
	ctx := context.Background()
	digest := "sha256:test-digest"

	stats, err := manager.TransferBlob(ctx, sourceRepo, destRepo, digest)
	if err != nil {
		t.Fatalf("TransferBlob failed: %v", err)
	}

	// Verify the stats
	if stats == nil {
		t.Fatal("Expected non-nil TransferStats")
	}
}

func TestTransferImage(t *testing.T) {
	logger := log.NewLogger()
	opts := DefaultTransferOptions()

	manager, err := NewTransferManager(opts, logger)
	if err != nil {
		t.Fatalf("Failed to create TransferManager: %v", err)
	}

	// Create source and destination repositories
	sourceRepo := NewTransferMockRepository("source/repo")
	destRepo := NewTransferMockRepository("dest/repo")

	// Add a manifest to the source
	manifest := []byte(`{"schemaVersion":2,"config":{"digest":"sha256:abc"},"layers":[{"digest":"layer1"},{"digest":"layer2"}]}`)
	ctx := context.Background()
	_ = sourceRepo.PutManifest(ctx, "latest", &interfaces.Manifest{
		Content:   manifest,
		MediaType: "application/json",
		Digest:    "sha256:latest",
	})

	// Test transfer
	stats, err := manager.TransferImage(ctx, sourceRepo, destRepo, "latest")
	if err != nil {
		t.Fatalf("TransferImage failed: %v", err)
	}

	// Verify the stats
	if stats == nil {
		t.Fatal("Expected non-nil TransferStats")
	}

	if stats.BytesTransferred == 0 {
		t.Errorf("Expected BytesTransferred > 0")
	}
}

func TestTransferStats(t *testing.T) {
	// Create a TransferStats
	stats := &TransferStats{
		BytesTransferred:    1000,
		BytesCompressed:     800,
		CompressionRatio:    0.8,
		DeltaReductions:     100,
		TransferDuration:    100 * time.Millisecond,
		CompressionDuration: 20 * time.Millisecond,
		RetryCount:          1,
	}

	// Verify calculations
	if stats.BytesTransferred != 1000 {
		t.Errorf("Expected BytesTransferred=1000, got %d", stats.BytesTransferred)
	}

	// Verify compression ratio is reasonable
	if stats.CompressionRatio < 0 || stats.CompressionRatio > 1.0 {
		t.Errorf("CompressionRatio should be between 0 and 1, got %f", stats.CompressionRatio)
	}
}
