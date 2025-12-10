package copy

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"freightliner/pkg/helper/log"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// MockMetrics implements the Metrics interface for testing
type MockMetricsImpl struct {
	StartCalled     bool
	CompletedCalled bool
	FailedCalled    bool
	LastSourceRef   string
	LastDestRef     string
	LastDuration    time.Duration
	LastLayerCount  int
	LastByteCount   int64
}

func (m *MockMetricsImpl) ReplicationStarted(source, destination string) {
	m.StartCalled = true
	m.LastSourceRef = source
	m.LastDestRef = destination
}

func (m *MockMetricsImpl) ReplicationCompleted(duration time.Duration, layerCount int, byteCount int64) {
	m.CompletedCalled = true
	m.LastDuration = duration
	m.LastLayerCount = layerCount
	m.LastByteCount = byteCount
}

func (m *MockMetricsImpl) ReplicationFailed() {
	m.FailedCalled = true
}

// MockLayer implements v1.Layer interface for testing
type MockLayer struct {
	digest     v1.Hash
	size       int64
	compressed []byte
}

func (m *MockLayer) Digest() (v1.Hash, error) {
	return m.digest, nil
}

func (m *MockLayer) DiffID() (v1.Hash, error) {
	return m.digest, nil
}

func (m *MockLayer) Compressed() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(m.compressed)), nil
}

func (m *MockLayer) Uncompressed() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(m.compressed)), nil
}

func (m *MockLayer) Size() (int64, error) {
	return m.size, nil
}

func (m *MockLayer) MediaType() (types.MediaType, error) {
	return types.DockerLayer, nil
}

// TestManifestDescriptor tests the manifestDescriptor implementation
func TestManifestDescriptor(t *testing.T) {
	data := []byte(`{"test": "manifest"}`)
	hash, _ := v1.NewHash("sha256:abc123def456")
	mediaType := types.DockerManifestSchema2

	desc := &manifestDescriptor{
		mediaType: mediaType,
		data:      data,
		hash:      hash,
	}

	// Test MediaType
	mt, err := desc.MediaType()
	if err != nil {
		t.Errorf("MediaType() error: %v", err)
	}
	if mt != mediaType {
		t.Errorf("Expected media type %s, got %s", mediaType, mt)
	}

	// Test RawManifest
	manifest, err := desc.RawManifest()
	if err != nil {
		t.Errorf("RawManifest() error: %v", err)
	}
	if !bytes.Equal(manifest, data) {
		t.Error("RawManifest data mismatch")
	}

	// Test Digest
	digest, err := desc.Digest()
	if err != nil {
		t.Errorf("Digest() error: %v", err)
	}
	if digest != hash {
		t.Errorf("Expected digest %s, got %s", hash, digest)
	}
}

// TestBlobLayer tests the blobLayer implementation
func TestBlobLayer(t *testing.T) {
	data := []byte("test blob data")
	hash, _ := v1.NewHash("sha256:test123")

	layer := &blobLayer{
		digestHash: hash,
		data:       data,
	}

	// Test Digest
	digest, err := layer.Digest()
	if err != nil {
		t.Errorf("Digest() error: %v", err)
	}
	if digest != hash {
		t.Errorf("Expected digest %s, got %s", hash, digest)
	}

	// Test DiffID
	diffID, err := layer.DiffID()
	if err != nil {
		t.Errorf("DiffID() error: %v", err)
	}
	if diffID != hash {
		t.Errorf("Expected diffID %s, got %s", hash, diffID)
	}

	// Test Size
	size, err := layer.Size()
	if err != nil {
		t.Errorf("Size() error: %v", err)
	}
	if size != int64(len(data)) {
		t.Errorf("Expected size %d, got %d", len(data), size)
	}

	// Test MediaType
	mt, err := layer.MediaType()
	if err != nil {
		t.Errorf("MediaType() error: %v", err)
	}
	if mt != types.DockerLayer {
		t.Errorf("Expected media type %s, got %s", types.DockerLayer, mt)
	}

	// Test Compressed
	reader, err := layer.Compressed()
	if err != nil {
		t.Errorf("Compressed() error: %v", err)
	}
	defer reader.Close()

	compressed, _ := io.ReadAll(reader)
	if !bytes.Equal(compressed, data) {
		t.Error("Compressed data mismatch")
	}

	// Test Uncompressed
	reader2, err := layer.Uncompressed()
	if err != nil {
		t.Errorf("Uncompressed() error: %v", err)
	}
	defer reader2.Close()

	uncompressed, _ := io.ReadAll(reader2)
	if !bytes.Equal(uncompressed, data) {
		t.Error("Uncompressed data mismatch")
	}
}

// TestStreamingBlobLayer tests the streamingBlobLayer implementation
func TestStreamingBlobLayer(t *testing.T) {
	data := []byte("streaming test data")
	hash, _ := v1.NewHash("sha256:stream123")
	reader := bytes.NewReader(data)

	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)

	layer := &streamingBlobLayer{
		digestHash: hash,
		reader:     reader,
		bufferMgr:  copier.bufferMgr,
		cachedSize: int64(len(data)),
	}

	// Test Digest
	digest, err := layer.Digest()
	if err != nil {
		t.Errorf("Digest() error: %v", err)
	}
	if digest != hash {
		t.Errorf("Expected digest %s, got %s", hash, digest)
	}

	// Test DiffID
	diffID, err := layer.DiffID()
	if err != nil {
		t.Errorf("DiffID() error: %v", err)
	}
	if diffID != hash {
		t.Error("DiffID mismatch")
	}

	// Test Size
	size, err := layer.Size()
	if err != nil {
		t.Errorf("Size() error: %v", err)
	}
	if size != int64(len(data)) {
		t.Errorf("Expected size %d, got %d", len(data), size)
	}

	// Test MediaType
	mt, err := layer.MediaType()
	if err != nil {
		t.Errorf("MediaType() error: %v", err)
	}
	if mt != types.DockerLayer {
		t.Errorf("Expected media type %s, got %s", types.DockerLayer, mt)
	}

	// Test Compressed
	compressed, err := layer.Compressed()
	if err != nil {
		t.Errorf("Compressed() error: %v", err)
	}
	defer compressed.Close()

	// Should be able to read from it
	buf := make([]byte, 10)
	n, _ := compressed.Read(buf)
	if n == 0 {
		t.Error("Expected to read data from compressed reader")
	}
}

// TestOptimizedReadCloser tests the optimizedReadCloser
func TestOptimizedReadCloser(t *testing.T) {
	data := []byte("test read closer data")
	reader := bytes.NewReader(data)

	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)

	orc := &optimizedReadCloser{
		reader:    reader,
		bufferMgr: copier.bufferMgr,
	}

	// Test Read
	buf := make([]byte, len(data))
	n, err := orc.Read(buf)
	if err != nil && err != io.EOF {
		t.Errorf("Read() error: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to read %d bytes, got %d", len(data), n)
	}

	// Test Close
	err = orc.Close()
	if err != nil {
		t.Errorf("Close() error: %v", err)
	}

	// Test multiple closes (should not panic)
	err = orc.Close()
	if err != nil {
		t.Errorf("Second Close() error: %v", err)
	}
}

// TestShouldCompress tests compression decision logic
func TestShouldCompress(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)

	tests := []struct {
		name     string
		size     int64
		expected bool
	}{
		{"very small", 512, false},
		{"at threshold", 1024, false},
		{"above threshold", 2048, true},
		{"large", 1024 * 1024, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := copier.shouldCompress(tt.size)
			if result != tt.expected {
				t.Errorf("shouldCompress(%d) = %v, expected %v", tt.size, result, tt.expected)
			}
		})
	}
}

// TestCompressStream tests stream compression
func TestCompressStream(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping compression test in short mode")
	}

	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)

	data := []byte("test data for compression")
	reader := io.NopCloser(bytes.NewReader(data))

	compressed, err := copier.compressStream(reader)
	if err != nil {
		t.Fatalf("compressStream() error: %v", err)
	}
	defer compressed.Close()

	// Read compressed data
	compressedData, err := io.ReadAll(compressed)
	if err != nil {
		t.Fatalf("Failed to read compressed data: %v", err)
	}

	// Compressed data should be different from original
	if bytes.Equal(compressedData, data) {
		t.Error("Expected compressed data to differ from original")
	}
}

// TestEncryptBlob tests blob encryption passthrough
func TestEncryptBlob(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)

	ctx := context.Background()
	data := []byte("test data")
	reader := io.NopCloser(bytes.NewReader(data))

	// Without encryption manager, should pass through
	result, err := copier.encryptBlob(ctx, reader, "registry.example.com")
	if err != nil {
		t.Fatalf("encryptBlob() error: %v", err)
	}

	if result != reader {
		t.Error("Expected reader to pass through when no encryption manager")
	}
}

// TestCopyBlobDeprecated tests the deprecated copyBlob method
func TestCopyBlobDeprecated(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)

	ctx := context.Background()
	_, err := copier.copyBlob(ctx, "src", "dest", "gzip", false)

	if err == nil {
		t.Error("Expected copyBlob to return error (deprecated)")
	}

	if !strings.Contains(err.Error(), "deprecated") {
		t.Errorf("Expected deprecation error, got: %v", err)
	}
}

// TestCopyOptionsValidation tests copy options structure
func TestCopyOptionsValidation(t *testing.T) {
	srcRef, _ := name.ParseReference("registry.example.com/source:tag")
	destRef, _ := name.ParseReference("registry.example.com/dest:tag")

	options := CopyOptions{
		DryRun:         true,
		ForceOverwrite: false,
		Source:         srcRef,
		Destination:    destRef,
	}

	if options.Source != srcRef {
		t.Error("Source reference mismatch")
	}

	if options.Destination != destRef {
		t.Error("Destination reference mismatch")
	}

	if !options.DryRun {
		t.Error("Expected DryRun to be true")
	}
}

// TestCopyStats tests copy statistics structure
func TestCopyStats(t *testing.T) {
	stats := CopyStats{
		BytesTransferred: 1024,
		CompressedBytes:  512,
		PullDuration:     1 * time.Second,
		PushDuration:     2 * time.Second,
		Layers:           5,
		ManifestSize:     256,
	}

	if stats.BytesTransferred != 1024 {
		t.Errorf("Expected BytesTransferred 1024, got %d", stats.BytesTransferred)
	}

	if stats.Layers != 5 {
		t.Errorf("Expected Layers 5, got %d", stats.Layers)
	}

	if stats.PullDuration != 1*time.Second {
		t.Errorf("Expected PullDuration 1s, got %v", stats.PullDuration)
	}
}

// TestCopyResult tests copy result structure
func TestCopyResult(t *testing.T) {
	stats := CopyStats{
		BytesTransferred: 2048,
		Layers:           3,
	}

	result := CopyResult{
		Success: true,
		Stats:   stats,
		Error:   nil,
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}

	if result.Stats.BytesTransferred != 2048 {
		t.Errorf("Expected BytesTransferred 2048, got %d", result.Stats.BytesTransferred)
	}

	if result.Error != nil {
		t.Error("Expected Error to be nil")
	}
}

// TestBlobTransferFunc tests blob transfer function type
func TestBlobTransferFunc(t *testing.T) {
	called := false
	var receivedSrc, receivedDest string

	transferFunc := BlobTransferFunc(func(ctx context.Context, src, dest string) error {
		called = true
		receivedSrc = src
		receivedDest = dest
		return nil
	})

	ctx := context.Background()
	err := transferFunc(ctx, "source-url", "dest-url")

	if err != nil {
		t.Errorf("transferFunc error: %v", err)
	}

	if !called {
		t.Error("Expected transfer function to be called")
	}

	if receivedSrc != "source-url" {
		t.Errorf("Expected source 'source-url', got '%s'", receivedSrc)
	}

	if receivedDest != "dest-url" {
		t.Errorf("Expected destination 'dest-url', got '%s'", receivedDest)
	}
}
