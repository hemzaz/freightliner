package copy

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"freightliner/pkg/helper/log"
)

// MockRepository is a mock implementation of the common.Repository interface
type MockRepository struct {
	Tags      []string
	Manifest  []byte
	MediaType string
	PutCalled bool
	GetCalled bool
	Name      string
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

func (m *MockRepository) GetRepositoryName() string {
	return m.Name
}

func (m *MockRepository) GetName() string {
	return m.Name
}

func (m *MockRepository) GetLayerReader(digest string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("mock layer content")), nil
}

// MockMetrics is a mock implementation of the Metrics interface
type MockMetrics struct {
	StartCalled     bool
	CompletedCalled bool
	FailedCalled    bool
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

func TestNewCopier(t *testing.T) {
	// Create a logger
	logger := log.NewBasicLogger(log.InfoLevel)

	// Create a copier
	copier := NewCopier(logger)

	// Ensure the copier was created properly
	if copier.logger == nil {
		t.Error("Expected logger to be set")
	}

	if copier.stats == nil {
		t.Error("Expected stats to be initialized")
	}

	if copier.transferFunc == nil {
		t.Error("Expected transferFunc to be initialized")
	}
}

func TestWithEncryptionManager(t *testing.T) {
	// Create a logger
	logger := log.NewBasicLogger(log.InfoLevel)

	// Create a copier
	copier := NewCopier(logger)

	// Set an encryption manager (nil for test)
	returnedCopier := copier.WithEncryptionManager(nil)

	// Test method chaining
	if returnedCopier != copier {
		t.Error("Method should return the same copier for chaining")
	}
}

func TestWithMetrics(t *testing.T) {
	// Create a logger
	logger := log.NewBasicLogger(log.InfoLevel)

	// Create a copier
	copier := NewCopier(logger)

	// Set metrics
	mockMetrics := &MockMetrics{}
	returnedCopier := copier.WithMetrics(mockMetrics)

	// Test method chaining
	if returnedCopier != copier {
		t.Error("Method should return the same copier for chaining")
	}

	// Check the metrics was set
	if copier.metrics != mockMetrics {
		t.Error("Metrics were not set properly")
	}
}

func TestWithBlobTransferFunc(t *testing.T) {
	// Create a logger
	logger := log.NewBasicLogger(log.InfoLevel)

	// Create a copier
	copier := NewCopier(logger)

	// Create a test transfer function
	testFunc := func(ctx context.Context, srcBlob, destBlob string) error {
		return nil
	}

	// Set the transfer function
	returnedCopier := copier.WithBlobTransferFunc(testFunc)

	// Test method chaining
	if returnedCopier != copier {
		t.Error("Method should return the same copier for chaining")
	}
}
