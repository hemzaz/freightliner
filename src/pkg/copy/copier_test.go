package copy

import (
	"context"
	"testing"
	"time"

	"github.com/hemzaz/freightliner/src/internal/log"
	"github.com/hemzaz/freightliner/src/pkg/client/common"
)

// MockRepository is a mock implementation of the common.Repository interface
type MockRepository struct {
	Tags     []string
	Manifest []byte
	MediaType string
	PutCalled bool
	GetCalled bool
}

func (m *MockRepository) ListTags() ([]string, error) {
	return m.Tags, nil
}

func (m *MockRepository) GetManifest(tag string) ([]byte, string, error) {
	m.GetCalled = true
	return m.Manifest, m.MediaType, nil
}

func (m *MockRepository) PutManifest(tag string, manifest []byte, mediaType string) error {
	m.PutCalled = true
	return nil
}

func (m *MockRepository) DeleteManifest(tag string) error {
	return nil
}

// MockMetrics is a mock implementation of the metrics.Metrics interface
type MockMetrics struct {
	StartCalled    bool
	CompletedCalled bool
	FailedCalled   bool
}

func (m *MockMetrics) ReplicationStarted(source, destination string) {
	m.StartCalled = true
}

func (m *MockMetrics) ReplicationCompleted(duration time.Duration, layerCount int, byteCount int64) {
	m.CompletedCalled = true
}

func (m *MockMetrics) ReplicationFailed() {
	m.FailedCalled = true
}

func TestCopyImage(t *testing.T) {
	// Create a logger
	logger := log.NewLogger(log.InfoLevel)

	// Create a copier
	copier := NewCopier(logger)

	// Create a source repository with a manifest
	sourceRepo := &MockRepository{
		Tags:      []string{"v1.0"},
		Manifest:  []byte(`{"schemaVersion":2,"mediaType":"application/vnd.docker.distribution.manifest.v2+json","config":{"mediaType":"application/vnd.docker.container.image.v1+json","size":1472,"digest":"sha256:c8be1b8f4d60d99c281fc2db75e0f56df42a83ad2f0b091621ce19357e19d853"},"layers":[{"mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip","size":2802957,"digest":"sha256:c9b1b535fdd91a9855fb7f82348177e5f019329a58c53c47272962dd60f71fc9"},{"mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip","size":108,"digest":"sha256:6b0937e234ce911b75630b744fb12836fe01bda5f7db203927edbb1390bc7e21"}]}`),
		MediaType: "application/vnd.docker.distribution.manifest.v2+json",
	}

	// Create a destination repository
	destRepo := &MockRepository{
		Tags: []string{},
	}

	// Create mock metrics
	mockMetrics := &MockMetrics{}
	copier.WithMetrics(mockMetrics)

	// Set up copy options
	options := CopyOptions{
		SourceTag:      "v1.0",
		DestinationTag: "latest",
		ForceOverwrite: false,
	}

	// Copy the image
	err := copier.CopyImage(context.Background(), sourceRepo, destRepo, options)
	if err != nil {
		t.Fatalf("CopyImage failed: %v", err)
	}

	// Check that the source repository was queried
	if !sourceRepo.GetCalled {
		t.Errorf("Expected GetManifest to be called on source repository")
	}

	// Check that the destination repository was updated
	if !destRepo.PutCalled {
		t.Errorf("Expected PutManifest to be called on destination repository")
	}

	// Check that metrics were recorded
	if !mockMetrics.StartCalled {
		t.Errorf("Expected ReplicationStarted to be called")
	}
	if !mockMetrics.CompletedCalled {
		t.Errorf("Expected ReplicationCompleted to be called")
	}
	if mockMetrics.FailedCalled {
		t.Errorf("ReplicationFailed was called unexpectedly")
	}
}

func TestCopyImage_AlreadyExists(t *testing.T) {
	// Create a logger
	logger := log.NewLogger(log.InfoLevel)

	// Create a copier
	copier := NewCopier(logger)

	// Create a source repository with a manifest
	sourceRepo := &MockRepository{
		Tags:      []string{"v1.0"},
		Manifest:  []byte(`{"schemaVersion":2,"mediaType":"application/vnd.docker.distribution.manifest.v2+json","config":{"mediaType":"application/vnd.docker.container.image.v1+json","size":1472,"digest":"sha256:c8be1b8f4d60d99c281fc2db75e0f56df42a83ad2f0b091621ce19357e19d853"},"layers":[{"mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip","size":2802957,"digest":"sha256:c9b1b535fdd91a9855fb7f82348177e5f019329a58c53c47272962dd60f71fc9"},{"mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip","size":108,"digest":"sha256:6b0937e234ce911b75630b744fb12836fe01bda5f7db203927edbb1390bc7e21"}]}`),
		MediaType: "application/vnd.docker.distribution.manifest.v2+json",
	}

	// Create a destination repository with the tag already existing
	destRepo := &MockRepository{
		Tags: []string{"latest"},
	}

	// Create mock metrics
	mockMetrics := &MockMetrics{}
	copier.WithMetrics(mockMetrics)

	// Set up copy options
	options := CopyOptions{
		SourceTag:      "v1.0",
		DestinationTag: "latest",
		ForceOverwrite: false,
	}

	// Copy the image
	err := copier.CopyImage(context.Background(), sourceRepo, destRepo, options)
	if err != nil {
		t.Fatalf("CopyImage failed: %v", err)
	}

	// Check that the manifest wasn't put (since tag already exists)
	if destRepo.PutCalled {
		t.Errorf("Expected PutManifest not to be called when tag already exists")
	}
}

func TestCopyImage_ForceOverwrite(t *testing.T) {
	// Create a logger
	logger := log.NewLogger(log.InfoLevel)

	// Create a copier
	copier := NewCopier(logger)

	// Create a source repository with a manifest
	sourceRepo := &MockRepository{
		Tags:      []string{"v1.0"},
		Manifest:  []byte(`{"schemaVersion":2,"mediaType":"application/vnd.docker.distribution.manifest.v2+json","config":{"mediaType":"application/vnd.docker.container.image.v1+json","size":1472,"digest":"sha256:c8be1b8f4d60d99c281fc2db75e0f56df42a83ad2f0b091621ce19357e19d853"},"layers":[{"mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip","size":2802957,"digest":"sha256:c9b1b535fdd91a9855fb7f82348177e5f019329a58c53c47272962dd60f71fc9"},{"mediaType":"application/vnd.docker.image.rootfs.diff.tar.gzip","size":108,"digest":"sha256:6b0937e234ce911b75630b744fb12836fe01bda5f7db203927edbb1390bc7e21"}]}`),
		MediaType: "application/vnd.docker.distribution.manifest.v2+json",
	}

	// Create a destination repository with the tag already existing
	destRepo := &MockRepository{
		Tags: []string{"latest"},
	}

	// Create mock metrics
	mockMetrics := &MockMetrics{}
	copier.WithMetrics(mockMetrics)

	// Set up copy options with force overwrite
	options := CopyOptions{
		SourceTag:      "v1.0",
		DestinationTag: "latest",
		ForceOverwrite: true,
	}

	// Copy the image
	err := copier.CopyImage(context.Background(), sourceRepo, destRepo, options)
	if err != nil {
		t.Fatalf("CopyImage failed: %v", err)
	}

	// Check that the manifest was put even though tag already exists
	if !destRepo.PutCalled {
		t.Errorf("Expected PutManifest to be called when force overwrite is true")
	}
}