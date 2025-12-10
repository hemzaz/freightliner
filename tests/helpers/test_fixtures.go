package helpers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"testing"

	"github.com/opencontainers/go-digest"
	specs "github.com/opencontainers/image-spec/specs-go"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

// TestImage represents a test container image
type TestImage struct {
	Repository string
	Tag        string
	Manifest   *v1.Manifest
	Config     *v1.Image
	Layers     []TestLayer
	TotalSize  int64
}

// TestLayer represents an image layer for testing
type TestLayer struct {
	Digest    digest.Digest
	Size      int64
	Data      []byte
	MediaType string
}

// ImageSize constants for test images
const (
	ImageSizeSmall  = 5 * 1024 * 1024        // 5MB
	ImageSizeMedium = 100 * 1024 * 1024      // 100MB
	ImageSizeLarge  = 1 * 1024 * 1024 * 1024 // 1GB
)

// GenerateTestImage creates a test image with specified size and layers
func GenerateTestImage(t *testing.T, totalSize int64, layerCount int) *TestImage {
	t.Helper()

	if layerCount <= 0 {
		layerCount = 1
	}

	layerSize := totalSize / int64(layerCount)
	layers := make([]TestLayer, layerCount)

	for i := 0; i < layerCount; i++ {
		layers[i] = GenerateTestLayer(t, layerSize)
	}

	// Create config
	config := &v1.Image{
		Platform: v1.Platform{
			Architecture: "amd64",
			OS:           "linux",
		},
		RootFS: v1.RootFS{
			Type:    "layers",
			DiffIDs: make([]digest.Digest, layerCount),
		},
	}

	for i, layer := range layers {
		config.RootFS.DiffIDs[i] = layer.Digest
	}

	// Create manifest
	manifest := &v1.Manifest{
		Versioned: specs.Versioned{
			SchemaVersion: 2,
		},
		MediaType: v1.MediaTypeImageManifest,
		Config: v1.Descriptor{
			MediaType: v1.MediaTypeImageConfig,
			Size:      int64(len([]byte("config"))),
			Digest:    digest.FromString("config"),
		},
		Layers: make([]v1.Descriptor, layerCount),
	}

	for i, layer := range layers {
		manifest.Layers[i] = v1.Descriptor{
			MediaType: layer.MediaType,
			Size:      layer.Size,
			Digest:    layer.Digest,
		}
	}

	return &TestImage{
		Repository: "test-repo",
		Tag:        "test-tag",
		Manifest:   manifest,
		Config:     config,
		Layers:     layers,
		TotalSize:  totalSize,
	}
}

// GenerateTestLayer creates a test layer with random data
func GenerateTestLayer(t *testing.T, size int64) TestLayer {
	t.Helper()

	// Generate random data
	data := make([]byte, size)
	_, err := rand.Read(data)
	require.NoError(t, err, "Failed to generate random data")

	// Calculate digest
	hash := sha256.Sum256(data)
	dgst := digest.NewDigestFromBytes(digest.SHA256, hash[:])

	return TestLayer{
		Digest:    dgst,
		Size:      size,
		Data:      data,
		MediaType: v1.MediaTypeImageLayerGzip,
	}
}

// GenerateRandomData generates random bytes of specified size
func GenerateRandomData(t *testing.T, size int64) []byte {
	t.Helper()

	data := make([]byte, size)
	_, err := rand.Read(data)
	require.NoError(t, err, "Failed to generate random data")

	return data
}

// GenerateDigest generates a digest from data
func GenerateDigest(data []byte) digest.Digest {
	hash := sha256.Sum256(data)
	return digest.NewDigestFromBytes(digest.SHA256, hash[:])
}

// GenerateRandomDigest generates a random digest
func GenerateRandomDigest() digest.Digest {
	hash := make([]byte, 32)
	rand.Read(hash)
	return digest.NewDigestFromBytes(digest.SHA256, hash)
}

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

// MockRegistryServer represents a mock registry for testing
type MockRegistryServer struct {
	Images    map[string]*TestImage
	Blobs     map[digest.Digest][]byte
	Manifests map[string][]byte
	Tags      map[string]map[string]digest.Digest // repo -> tag -> digest
}

// NewMockRegistryServer creates a new mock registry server
func NewMockRegistryServer() *MockRegistryServer {
	return &MockRegistryServer{
		Images:    make(map[string]*TestImage),
		Blobs:     make(map[digest.Digest][]byte),
		Manifests: make(map[string][]byte),
		Tags:      make(map[string]map[string]digest.Digest),
	}
}

// AddImage adds a test image to the mock registry
func (m *MockRegistryServer) AddImage(image *TestImage) {
	key := fmt.Sprintf("%s:%s", image.Repository, image.Tag)
	m.Images[key] = image

	// Add layers as blobs
	for _, layer := range image.Layers {
		m.Blobs[layer.Digest] = layer.Data
	}

	// Add tags
	if m.Tags[image.Repository] == nil {
		m.Tags[image.Repository] = make(map[string]digest.Digest)
	}
	m.Tags[image.Repository][image.Tag] = image.Manifest.Config.Digest
}

// GetImage retrieves an image from the mock registry
func (m *MockRegistryServer) GetImage(repository, tag string) (*TestImage, bool) {
	key := fmt.Sprintf("%s:%s", repository, tag)
	img, ok := m.Images[key]
	return img, ok
}

// GetBlob retrieves a blob from the mock registry
func (m *MockRegistryServer) GetBlob(dgst digest.Digest) ([]byte, bool) {
	blob, ok := m.Blobs[dgst]
	return blob, ok
}

// HasImage checks if an image exists in the mock registry
func (m *MockRegistryServer) HasImage(repository, tag string) bool {
	_, ok := m.GetImage(repository, tag)
	return ok
}

// ListTags lists all tags for a repository
func (m *MockRegistryServer) ListTags(repository string) []string {
	tags := m.Tags[repository]
	result := make([]string, 0, len(tags))
	for tag := range tags {
		result = append(result, tag)
	}
	return result
}

// TestRegistry represents a test registry configuration
type TestRegistry struct {
	Name     string
	Type     string
	Endpoint string
	Username string
	Password string
	Insecure bool
}

// NewTestRegistry creates a test registry configuration
func NewTestRegistry(name, endpoint string) *TestRegistry {
	return &TestRegistry{
		Name:     name,
		Type:     "generic",
		Endpoint: endpoint,
		Insecure: true,
	}
}

// WithAuth sets authentication for the test registry
func (r *TestRegistry) WithAuth(username, password string) *TestRegistry {
	r.Username = username
	r.Password = password
	return r
}

// WithType sets the registry type
func (r *TestRegistry) WithType(regType string) *TestRegistry {
	r.Type = regType
	return r
}

// TestContext provides a test context with cancellation
type TestContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// NewTestContext creates a new test context
func NewTestContext() *TestContext {
	ctx, cancel := context.WithCancel(context.Background())
	return &TestContext{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Context returns the underlying context
func (tc *TestContext) Context() context.Context {
	return tc.ctx
}

// Cancel cancels the context
func (tc *TestContext) Cancel() {
	tc.cancel()
}

// Cleanup registers a cleanup function
func (tc *TestContext) Cleanup(fn func()) {
	// Call cleanup when context is cancelled
	go func() {
		<-tc.ctx.Done()
		fn()
	}()
}

// BlobGenerator generates blobs for testing
type BlobGenerator struct {
	reader io.Reader
	size   int64
}

// NewBlobGenerator creates a new blob generator
func NewBlobGenerator(size int64) *BlobGenerator {
	return &BlobGenerator{
		reader: io.LimitReader(rand.Reader, size),
		size:   size,
	}
}

// Read reads from the blob generator
func (b *BlobGenerator) Read(p []byte) (n int, err error) {
	return b.reader.Read(p)
}

// Size returns the blob size
func (b *BlobGenerator) Size() int64 {
	return b.size
}

// Close closes the blob generator
func (b *BlobGenerator) Close() error {
	return nil
}

// ReplicationTestCase represents a replication test case
type ReplicationTestCase struct {
	Name             string
	SourceRegistry   string
	SourceRepository string
	DestRegistry     string
	DestRepository   string
	Tags             []string
	ExpectedSuccess  bool
	ExpectedError    string
	SetupFunc        func(*testing.T)
	VerifyFunc       func(*testing.T)
}

// RunReplicationTestCase runs a replication test case
func RunReplicationTestCase(t *testing.T, tc ReplicationTestCase) {
	t.Helper()

	if tc.SetupFunc != nil {
		tc.SetupFunc(t)
	}

	// Run test logic here
	// This would be implemented based on the actual replication service

	if tc.VerifyFunc != nil {
		tc.VerifyFunc(t)
	}
}

// PerformanceMetrics holds performance test metrics
type PerformanceMetrics struct {
	Throughput     float64 // MB/s
	Latency        float64 // seconds
	MemoryUsage    int64   // bytes
	CPUUtilization float64 // percentage
	ErrorRate      float64 // percentage
}

// NewPerformanceMetrics creates a new performance metrics struct
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{}
}

// RecordThroughput records throughput in MB/s
func (p *PerformanceMetrics) RecordThroughput(bytes int64, duration float64) {
	p.Throughput = float64(bytes) / duration / 1024 / 1024
}

// RecordLatency records latency in seconds
func (p *PerformanceMetrics) RecordLatency(duration float64) {
	p.Latency = duration
}

// RecordMemory records memory usage
func (p *PerformanceMetrics) RecordMemory(bytes int64) {
	p.MemoryUsage = bytes
}

// RecordCPU records CPU utilization
func (p *PerformanceMetrics) RecordCPU(utilization float64) {
	p.CPUUtilization = utilization
}

// RecordErrors records error rate
func (p *PerformanceMetrics) RecordErrors(errors, total int) {
	if total > 0 {
		p.ErrorRate = float64(errors) / float64(total) * 100
	}
}

// String returns a string representation of metrics
func (p *PerformanceMetrics) String() string {
	return fmt.Sprintf(
		"Throughput: %.2f MB/s, Latency: %.3fs, Memory: %d bytes, CPU: %.2f%%, Errors: %.2f%%",
		p.Throughput, p.Latency, p.MemoryUsage, p.CPUUtilization, p.ErrorRate,
	)
}
