package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"freightliner/pkg/manifest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSizeEstimator is a mock implementation of SizeEstimator interface
type MockSizeEstimator struct {
	mock.Mock
}

// GetManifest mocks the GetManifest method
func (m *MockSizeEstimator) GetManifest(ctx context.Context, repository, tag string) ([]byte, string, error) {
	args := m.Called(ctx, repository, tag)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).([]byte), args.String(1), args.Error(2)
}

// Test helper functions to create manifest data

func createOCIManifestJSON(configSize int64, layerSizes []int64) []byte {
	layers := make([]manifest.OCIDescriptor, len(layerSizes))
	for i, size := range layerSizes {
		layers[i] = manifest.OCIDescriptor{
			MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
			Size:      size,
			Digest:    fmt.Sprintf("sha256:layer%d", i),
		}
	}

	m := manifest.OCIManifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config: manifest.OCIDescriptor{
			MediaType: "application/vnd.oci.image.config.v1+json",
			Size:      configSize,
			Digest:    "sha256:config",
		},
		Layers: layers,
	}

	data, _ := json.Marshal(m)
	return data
}

func createDockerV2ManifestJSON(configSize int64, layerSizes []int64) []byte {
	layers := make([]manifest.DockerDescriptor, len(layerSizes))
	for i, size := range layerSizes {
		layers[i] = manifest.DockerDescriptor{
			MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
			Size:      size,
			Digest:    fmt.Sprintf("sha256:layer%d", i),
		}
	}

	m := manifest.DockerV2Schema2Manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.docker.distribution.manifest.v2+json",
		Config: manifest.DockerDescriptor{
			MediaType: "application/vnd.docker.container.image.v1+json",
			Size:      configSize,
			Digest:    "sha256:config",
		},
		Layers: layers,
	}

	data, _ := json.Marshal(m)
	return data
}

func createOCIImageIndexJSON(manifestSizes []int64) []byte {
	manifests := make([]manifest.OCIDescriptor, len(manifestSizes))
	for i, size := range manifestSizes {
		manifests[i] = manifest.OCIDescriptor{
			MediaType: "application/vnd.oci.image.manifest.v1+json",
			Size:      size,
			Digest:    fmt.Sprintf("sha256:manifest%d", i),
			Platform: &manifest.Platform{
				OS:           "linux",
				Architecture: fmt.Sprintf("arch%d", i),
			},
		}
	}

	index := manifest.OCIImageIndex{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.index.v1+json",
		Manifests:     manifests,
	}

	data, _ := json.Marshal(index)
	return data
}

func createDockerManifestListJSON(manifestSizes []int64) []byte {
	manifests := make([]manifest.DockerManifestListDescriptor, len(manifestSizes))
	for i, size := range manifestSizes {
		manifests[i] = manifest.DockerManifestListDescriptor{
			MediaType: "application/vnd.docker.distribution.manifest.v2+json",
			Size:      size,
			Digest:    fmt.Sprintf("sha256:manifest%d", i),
			Platform: &manifest.Platform{
				OS:           "linux",
				Architecture: fmt.Sprintf("arch%d", i),
			},
		}
	}

	list := manifest.DockerManifestList{
		SchemaVersion: 2,
		MediaType:     "application/vnd.docker.distribution.manifest.list.v2+json",
		Manifests:     manifests,
	}

	data, _ := json.Marshal(list)
	return data
}

func createDockerV1ManifestJSON() []byte {
	type V1FSLayer struct {
		BlobSum string `json:"blobSum"`
	}
	type V1Manifest struct {
		SchemaVersion int         `json:"schemaVersion"`
		FSLayers      []V1FSLayer `json:"fsLayers"`
	}

	m := V1Manifest{
		SchemaVersion: 1,
		FSLayers: []V1FSLayer{
			{BlobSum: "sha256:layer1"},
			{BlobSum: "sha256:layer2"},
		},
	}

	data, _ := json.Marshal(m)
	return data
}

// TestEstimateImageSize tests the main entry point with different manifest types
func TestEstimateImageSize_OCIManifest(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	manifestData := createOCIManifestJSON(1024, []int64{2048, 3072, 4096})
	mockEstimator.On("GetManifest", ctx, "test/repo", "v1.0").
		Return(manifestData, "application/vnd.oci.image.manifest.v1+json", nil)

	size, err := EstimateImageSize(ctx, mockEstimator, "test/repo", "v1.0")

	require.NoError(t, err)
	assert.Equal(t, int64(1024+2048+3072+4096), size)
	mockEstimator.AssertExpectations(t)
}

func TestEstimateImageSize_DockerV2Manifest(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	manifestData := createDockerV2ManifestJSON(512, []int64{1024, 2048})
	mockEstimator.On("GetManifest", ctx, "test/repo", "latest").
		Return(manifestData, "application/vnd.docker.distribution.manifest.v2+json", nil)

	size, err := EstimateImageSize(ctx, mockEstimator, "test/repo", "latest")

	require.NoError(t, err)
	assert.Equal(t, int64(512+1024+2048), size)
	mockEstimator.AssertExpectations(t)
}

func TestEstimateImageSize_OCIImageIndex(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	manifestData := createOCIImageIndexJSON([]int64{5000, 6000, 7000})
	mockEstimator.On("GetManifest", ctx, "test/multiarch", "v2.0").
		Return(manifestData, "application/vnd.oci.image.index.v1+json", nil)

	size, err := EstimateImageSize(ctx, mockEstimator, "test/multiarch", "v2.0")

	require.NoError(t, err)
	assert.Equal(t, int64(5000+6000+7000), size)
	mockEstimator.AssertExpectations(t)
}

func TestEstimateImageSize_DockerManifestList(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	manifestData := createDockerManifestListJSON([]int64{10000, 20000})
	mockEstimator.On("GetManifest", ctx, "test/multiarch", "v3.0").
		Return(manifestData, "application/vnd.docker.distribution.manifest.list.v2+json", nil)

	size, err := EstimateImageSize(ctx, mockEstimator, "test/multiarch", "v3.0")

	require.NoError(t, err)
	assert.Equal(t, int64(10000+20000), size)
	mockEstimator.AssertExpectations(t)
}

func TestEstimateImageSize_DockerV1Manifest(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	manifestData := createDockerV1ManifestJSON()
	mockEstimator.On("GetManifest", ctx, "test/legacy", "old").
		Return(manifestData, "application/vnd.docker.distribution.manifest.v1+json", nil)

	size, err := EstimateImageSize(ctx, mockEstimator, "test/legacy", "old")

	require.Error(t, err)
	assert.Equal(t, int64(0), size)
	assert.Contains(t, err.Error(), "Docker V1 manifests don't include size information")
	mockEstimator.AssertExpectations(t)
}

func TestEstimateImageSize_UnsupportedMediaType(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	mockEstimator.On("GetManifest", ctx, "test/repo", "v1.0").
		Return([]byte("{}"), "application/unsupported.type+json", nil)

	size, err := EstimateImageSize(ctx, mockEstimator, "test/repo", "v1.0")

	require.Error(t, err)
	assert.Equal(t, int64(0), size)
	assert.Contains(t, err.Error(), "unsupported manifest media type")
	mockEstimator.AssertExpectations(t)
}

func TestEstimateImageSize_GetManifestError(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	expectedErr := fmt.Errorf("network error")
	mockEstimator.On("GetManifest", ctx, "test/repo", "v1.0").
		Return(nil, "", expectedErr)

	size, err := EstimateImageSize(ctx, mockEstimator, "test/repo", "v1.0")

	require.Error(t, err)
	assert.Equal(t, int64(0), size)
	assert.Contains(t, err.Error(), "failed to get manifest")
	assert.Contains(t, err.Error(), "network error")
	mockEstimator.AssertExpectations(t)
}

// TestEstimateOCIManifestSize tests OCI manifest size calculation
func TestEstimateOCIManifestSize_Success(t *testing.T) {
	manifestData := createOCIManifestJSON(1000, []int64{2000, 3000, 4000})

	size, err := estimateOCIManifestSize(manifestData)

	require.NoError(t, err)
	assert.Equal(t, int64(1000+2000+3000+4000), size)
}

func TestEstimateOCIManifestSize_EmptyLayers(t *testing.T) {
	manifestData := createOCIManifestJSON(1000, []int64{})

	size, err := estimateOCIManifestSize(manifestData)

	require.NoError(t, err)
	assert.Equal(t, int64(1000), size)
}

func TestEstimateOCIManifestSize_ZeroSizes(t *testing.T) {
	manifestData := createOCIManifestJSON(0, []int64{0, 0})

	size, err := estimateOCIManifestSize(manifestData)

	require.NoError(t, err)
	assert.Equal(t, int64(0), size)
}

func TestEstimateOCIManifestSize_InvalidJSON(t *testing.T) {
	invalidJSON := []byte("{invalid json")

	size, err := estimateOCIManifestSize(invalidJSON)

	require.Error(t, err)
	assert.Equal(t, int64(0), size)
	assert.Contains(t, err.Error(), "failed to parse OCI manifest")
}

func TestEstimateOCIManifestSize_LargeSizes(t *testing.T) {
	// Test with realistic large image sizes
	manifestData := createOCIManifestJSON(1024*1024, []int64{
		100 * 1024 * 1024, // 100MB
		200 * 1024 * 1024, // 200MB
		150 * 1024 * 1024, // 150MB
	})

	size, err := estimateOCIManifestSize(manifestData)

	require.NoError(t, err)
	expectedSize := int64(1024*1024 + 100*1024*1024 + 200*1024*1024 + 150*1024*1024)
	assert.Equal(t, expectedSize, size)
}

// TestEstimateDockerV2ManifestSize tests Docker V2 manifest size calculation
func TestEstimateDockerV2ManifestSize_Success(t *testing.T) {
	manifestData := createDockerV2ManifestJSON(500, []int64{1000, 2000})

	size, err := estimateDockerV2ManifestSize(manifestData)

	require.NoError(t, err)
	assert.Equal(t, int64(500+1000+2000), size)
}

func TestEstimateDockerV2ManifestSize_EmptyLayers(t *testing.T) {
	manifestData := createDockerV2ManifestJSON(500, []int64{})

	size, err := estimateDockerV2ManifestSize(manifestData)

	require.NoError(t, err)
	assert.Equal(t, int64(500), size)
}

func TestEstimateDockerV2ManifestSize_InvalidJSON(t *testing.T) {
	invalidJSON := []byte("not json")

	size, err := estimateDockerV2ManifestSize(invalidJSON)

	require.Error(t, err)
	assert.Equal(t, int64(0), size)
	assert.Contains(t, err.Error(), "failed to parse Docker V2 manifest")
}

func TestEstimateDockerV2ManifestSize_LargeSizes(t *testing.T) {
	// Test with multiple large layers
	manifestData := createDockerV2ManifestJSON(2*1024*1024, []int64{
		500 * 1024 * 1024,
		300 * 1024 * 1024,
		100 * 1024 * 1024,
	})

	size, err := estimateDockerV2ManifestSize(manifestData)

	require.NoError(t, err)
	expectedSize := int64(2*1024*1024 + 500*1024*1024 + 300*1024*1024 + 100*1024*1024)
	assert.Equal(t, expectedSize, size)
}

// TestEstimateMultiArchManifestSize tests multi-arch manifest size calculation
func TestEstimateMultiArchManifestSize_OCIIndex(t *testing.T) {
	manifestData := createOCIImageIndexJSON([]int64{10000, 20000, 30000})

	size, err := estimateMultiArchManifestSize(manifestData)

	require.NoError(t, err)
	assert.Equal(t, int64(10000+20000+30000), size)
}

func TestEstimateMultiArchManifestSize_DockerList(t *testing.T) {
	manifestData := createDockerManifestListJSON([]int64{15000, 25000})

	size, err := estimateMultiArchManifestSize(manifestData)

	require.NoError(t, err)
	assert.Equal(t, int64(15000+25000), size)
}

func TestEstimateMultiArchManifestSize_EmptyManifests(t *testing.T) {
	manifestData := createOCIImageIndexJSON([]int64{})

	size, err := estimateMultiArchManifestSize(manifestData)

	require.NoError(t, err)
	assert.Equal(t, int64(0), size)
}

func TestEstimateMultiArchManifestSize_InvalidJSON(t *testing.T) {
	invalidJSON := []byte("{not valid json at all")

	size, err := estimateMultiArchManifestSize(invalidJSON)

	require.Error(t, err)
	assert.Equal(t, int64(0), size)
	assert.Contains(t, err.Error(), "failed to parse multi-arch manifest")
}

func TestEstimateMultiArchManifestSize_SingleManifest(t *testing.T) {
	manifestData := createOCIImageIndexJSON([]int64{50000})

	size, err := estimateMultiArchManifestSize(manifestData)

	require.NoError(t, err)
	assert.Equal(t, int64(50000), size)
}

func TestEstimateMultiArchManifestSize_ManyPlatforms(t *testing.T) {
	// Test with many platforms
	sizes := make([]int64, 10)
	for i := range sizes {
		sizes[i] = int64((i + 1) * 5000)
	}
	manifestData := createOCIImageIndexJSON(sizes)

	size, err := estimateMultiArchManifestSize(manifestData)

	require.NoError(t, err)
	var expectedSize int64
	for _, s := range sizes {
		expectedSize += s
	}
	assert.Equal(t, expectedSize, size)
}

func TestEstimateMultiArchManifestSize_DockerListFallback(t *testing.T) {
	// Create a Docker manifest list (the function should try OCI first, then Docker)
	// Both formats are actually compatible JSON, so this mainly tests the logic flow
	manifestData := createDockerManifestListJSON([]int64{12000, 24000, 36000})

	size, err := estimateMultiArchManifestSize(manifestData)

	require.NoError(t, err)
	assert.Equal(t, int64(12000+24000+36000), size)
}

func TestEstimateMultiArchManifestSize_BothParsesFail(t *testing.T) {
	// Test with completely invalid JSON that will fail both OCI and Docker parsing
	// The function tries OCI first, then Docker, then returns error
	invalidManifest := []byte(`this is not valid JSON at all { broken }`)

	size, err := estimateMultiArchManifestSize(invalidManifest)

	require.Error(t, err)
	assert.Equal(t, int64(0), size)
	assert.Contains(t, err.Error(), "failed to parse multi-arch manifest")
}

// TestEstimateDockerV1ManifestSize tests Docker V1 manifest size calculation
func TestEstimateDockerV1ManifestSize_ReturnsError(t *testing.T) {
	manifestData := createDockerV1ManifestJSON()

	size, err := estimateDockerV1ManifestSize(manifestData)

	require.Error(t, err)
	assert.Equal(t, int64(0), size)
	assert.Contains(t, err.Error(), "Docker V1 manifests don't include size information")
}

func TestEstimateDockerV1ManifestSize_InvalidJSON(t *testing.T) {
	invalidJSON := []byte("invalid")

	size, err := estimateDockerV1ManifestSize(invalidJSON)

	require.Error(t, err)
	assert.Equal(t, int64(0), size)
	assert.Contains(t, err.Error(), "failed to parse Docker V1 manifest")
}

// TestOptimizeBatchesWithSizeEstimation tests batch optimization
func TestOptimizeBatchesWithSizeEstimation_PrioritySorting(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	tasks := []SyncTask{
		{SourceRegistry: "reg1", SourceRepository: "repo1", SourceTag: "tag1", Priority: 1},
		{SourceRegistry: "reg1", SourceRepository: "repo2", SourceTag: "tag2", Priority: 3},
		{SourceRegistry: "reg1", SourceRepository: "repo3", SourceTag: "tag3", Priority: 2},
	}

	// Mock size estimation to return same size for all
	for _, task := range tasks {
		manifestData := createOCIManifestJSON(1000, []int64{2000})
		mockEstimator.On("GetManifest", ctx, task.SourceRepository, task.SourceTag).
			Return(manifestData, "application/vnd.oci.image.manifest.v1+json", nil)
	}

	optimized := OptimizeBatchesWithSizeEstimation(ctx, tasks, mockEstimator)

	require.Len(t, optimized, 3)
	// Should be sorted by priority (descending)
	assert.Equal(t, 3, optimized[0].Priority)
	assert.Equal(t, 2, optimized[1].Priority)
	assert.Equal(t, 1, optimized[2].Priority)
	mockEstimator.AssertExpectations(t)
}

func TestOptimizeBatchesWithSizeEstimation_RegistryGrouping(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	tasks := []SyncTask{
		{SourceRegistry: "reg2", SourceRepository: "repo1", SourceTag: "tag1", Priority: 1},
		{SourceRegistry: "reg1", SourceRepository: "repo2", SourceTag: "tag2", Priority: 1},
		{SourceRegistry: "reg1", SourceRepository: "repo3", SourceTag: "tag3", Priority: 1},
	}

	for _, task := range tasks {
		manifestData := createOCIManifestJSON(1000, []int64{2000})
		mockEstimator.On("GetManifest", ctx, task.SourceRepository, task.SourceTag).
			Return(manifestData, "application/vnd.oci.image.manifest.v1+json", nil)
	}

	optimized := OptimizeBatchesWithSizeEstimation(ctx, tasks, mockEstimator)

	require.Len(t, optimized, 3)
	// Same priority, should group by registry
	assert.Equal(t, "reg1", optimized[0].SourceRegistry)
	assert.Equal(t, "reg1", optimized[1].SourceRegistry)
	assert.Equal(t, "reg2", optimized[2].SourceRegistry)
	mockEstimator.AssertExpectations(t)
}

func TestOptimizeBatchesWithSizeEstimation_SizeSorting(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	tasks := []SyncTask{
		{SourceRegistry: "reg1", SourceRepository: "large", SourceTag: "v1", Priority: 1},
		{SourceRegistry: "reg1", SourceRepository: "small", SourceTag: "v1", Priority: 1},
		{SourceRegistry: "reg1", SourceRepository: "medium", SourceTag: "v1", Priority: 1},
	}

	// Mock different sizes
	mockEstimator.On("GetManifest", ctx, "large", "v1").
		Return(createOCIManifestJSON(1000, []int64{50000}), "application/vnd.oci.image.manifest.v1+json", nil)
	mockEstimator.On("GetManifest", ctx, "small", "v1").
		Return(createOCIManifestJSON(1000, []int64{10000}), "application/vnd.oci.image.manifest.v1+json", nil)
	mockEstimator.On("GetManifest", ctx, "medium", "v1").
		Return(createOCIManifestJSON(1000, []int64{30000}), "application/vnd.oci.image.manifest.v1+json", nil)

	optimized := OptimizeBatchesWithSizeEstimation(ctx, tasks, mockEstimator)

	require.Len(t, optimized, 3)
	// Same priority and registry, should sort by size (smaller first)
	assert.Equal(t, "small", optimized[0].SourceRepository)
	assert.Equal(t, "medium", optimized[1].SourceRepository)
	assert.Equal(t, "large", optimized[2].SourceRepository)
	mockEstimator.AssertExpectations(t)
}

func TestOptimizeBatchesWithSizeEstimation_NilEstimator(t *testing.T) {
	ctx := context.Background()

	tasks := []SyncTask{
		{SourceRegistry: "reg1", SourceRepository: "repo1", SourceTag: "tag1", Priority: 1},
		{SourceRegistry: "reg1", SourceRepository: "repo2", SourceTag: "tag2", Priority: 2},
	}

	optimized := OptimizeBatchesWithSizeEstimation(ctx, tasks, nil)

	require.Len(t, optimized, 2)
	// Should still sort by priority even without estimator
	assert.Equal(t, 2, optimized[0].Priority)
	assert.Equal(t, 1, optimized[1].Priority)
}

func TestOptimizeBatchesWithSizeEstimation_EstimationErrors(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	tasks := []SyncTask{
		{SourceRegistry: "reg1", SourceRepository: "repo1", SourceTag: "tag1", Priority: 1},
		{SourceRegistry: "reg1", SourceRepository: "repo2", SourceTag: "tag2", Priority: 1},
	}

	// First succeeds, second fails
	mockEstimator.On("GetManifest", ctx, "repo1", "tag1").
		Return(createOCIManifestJSON(1000, []int64{5000}), "application/vnd.oci.image.manifest.v1+json", nil)
	mockEstimator.On("GetManifest", ctx, "repo2", "tag2").
		Return(nil, "", fmt.Errorf("estimation error"))

	optimized := OptimizeBatchesWithSizeEstimation(ctx, tasks, mockEstimator)

	require.Len(t, optimized, 2)
	// Should continue even if estimation fails
	mockEstimator.AssertExpectations(t)
}

func TestOptimizeBatchesWithSizeEstimation_EmptyTasks(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	tasks := []SyncTask{}

	optimized := OptimizeBatchesWithSizeEstimation(ctx, tasks, mockEstimator)

	require.Len(t, optimized, 0)
}

func TestOptimizeBatchesWithSizeEstimation_SingleTask(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	tasks := []SyncTask{
		{SourceRegistry: "reg1", SourceRepository: "repo1", SourceTag: "tag1", Priority: 1},
	}

	manifestData := createOCIManifestJSON(1000, []int64{2000})
	mockEstimator.On("GetManifest", ctx, "repo1", "tag1").
		Return(manifestData, "application/vnd.oci.image.manifest.v1+json", nil)

	optimized := OptimizeBatchesWithSizeEstimation(ctx, tasks, mockEstimator)

	require.Len(t, optimized, 1)
	assert.Equal(t, tasks[0], optimized[0])
	mockEstimator.AssertExpectations(t)
}

func TestOptimizeBatchesWithSizeEstimation_ComplexScenario(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	tasks := []SyncTask{
		// Priority 3, reg1, large
		{SourceRegistry: "reg1", SourceRepository: "p3-large", SourceTag: "v1", Priority: 3},
		// Priority 2, reg1, small
		{SourceRegistry: "reg1", SourceRepository: "p2-small", SourceTag: "v1", Priority: 2},
		// Priority 3, reg2, small
		{SourceRegistry: "reg2", SourceRepository: "p3-small", SourceTag: "v1", Priority: 3},
		// Priority 1, reg1, medium
		{SourceRegistry: "reg1", SourceRepository: "p1-medium", SourceTag: "v1", Priority: 1},
		// Priority 2, reg1, medium
		{SourceRegistry: "reg1", SourceRepository: "p2-medium", SourceTag: "v1", Priority: 2},
	}

	// Mock sizes
	mockEstimator.On("GetManifest", ctx, "p3-large", "v1").
		Return(createOCIManifestJSON(1000, []int64{100000}), "application/vnd.oci.image.manifest.v1+json", nil)
	mockEstimator.On("GetManifest", ctx, "p2-small", "v1").
		Return(createOCIManifestJSON(1000, []int64{10000}), "application/vnd.oci.image.manifest.v1+json", nil)
	mockEstimator.On("GetManifest", ctx, "p3-small", "v1").
		Return(createOCIManifestJSON(1000, []int64{15000}), "application/vnd.oci.image.manifest.v1+json", nil)
	mockEstimator.On("GetManifest", ctx, "p1-medium", "v1").
		Return(createOCIManifestJSON(1000, []int64{50000}), "application/vnd.oci.image.manifest.v1+json", nil)
	mockEstimator.On("GetManifest", ctx, "p2-medium", "v1").
		Return(createOCIManifestJSON(1000, []int64{40000}), "application/vnd.oci.image.manifest.v1+json", nil)

	optimized := OptimizeBatchesWithSizeEstimation(ctx, tasks, mockEstimator)

	require.Len(t, optimized, 5)

	// Verify priority 3 items come first
	assert.Equal(t, 3, optimized[0].Priority)
	assert.Equal(t, 3, optimized[1].Priority)

	// Within priority 3: reg1 before reg2 (alphabetically)
	// Within reg1: smaller size first

	// Then priority 2 items
	assert.Equal(t, 2, optimized[2].Priority)
	assert.Equal(t, 2, optimized[3].Priority)

	// Finally priority 1
	assert.Equal(t, 1, optimized[4].Priority)

	mockEstimator.AssertExpectations(t)
}

// TestEstimateBatchSize tests batch size estimation
func TestEstimateBatchSize_Success(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	tasks := []SyncTask{
		{SourceRepository: "repo1", SourceTag: "tag1"},
		{SourceRepository: "repo2", SourceTag: "tag2"},
		{SourceRepository: "repo3", SourceTag: "tag3"},
	}

	mockEstimator.On("GetManifest", ctx, "repo1", "tag1").
		Return(createOCIManifestJSON(1000, []int64{5000}), "application/vnd.oci.image.manifest.v1+json", nil)
	mockEstimator.On("GetManifest", ctx, "repo2", "tag2").
		Return(createOCIManifestJSON(2000, []int64{6000}), "application/vnd.oci.image.manifest.v1+json", nil)
	mockEstimator.On("GetManifest", ctx, "repo3", "tag3").
		Return(createOCIManifestJSON(3000, []int64{7000}), "application/vnd.oci.image.manifest.v1+json", nil)

	totalSize, err := EstimateBatchSize(ctx, tasks, mockEstimator)

	require.NoError(t, err)
	expectedSize := int64((1000 + 5000) + (2000 + 6000) + (3000 + 7000))
	assert.Equal(t, expectedSize, totalSize)
	mockEstimator.AssertExpectations(t)
}

func TestEstimateBatchSize_EmptyBatch(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	tasks := []SyncTask{}

	totalSize, err := EstimateBatchSize(ctx, tasks, mockEstimator)

	require.NoError(t, err)
	assert.Equal(t, int64(0), totalSize)
}

func TestEstimateBatchSize_SomeEstimationsFail(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	tasks := []SyncTask{
		{SourceRepository: "repo1", SourceTag: "tag1"},
		{SourceRepository: "repo2", SourceTag: "tag2"},
		{SourceRepository: "repo3", SourceTag: "tag3"},
	}

	// First and third succeed, second fails
	mockEstimator.On("GetManifest", ctx, "repo1", "tag1").
		Return(createOCIManifestJSON(1000, []int64{5000}), "application/vnd.oci.image.manifest.v1+json", nil)
	mockEstimator.On("GetManifest", ctx, "repo2", "tag2").
		Return(nil, "", fmt.Errorf("estimation error"))
	mockEstimator.On("GetManifest", ctx, "repo3", "tag3").
		Return(createOCIManifestJSON(3000, []int64{7000}), "application/vnd.oci.image.manifest.v1+json", nil)

	totalSize, err := EstimateBatchSize(ctx, tasks, mockEstimator)

	require.NoError(t, err)
	// Should only count successful estimations
	expectedSize := int64((1000 + 5000) + (3000 + 7000))
	assert.Equal(t, expectedSize, totalSize)
	mockEstimator.AssertExpectations(t)
}

func TestEstimateBatchSize_AllEstimationsFail(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	tasks := []SyncTask{
		{SourceRepository: "repo1", SourceTag: "tag1"},
		{SourceRepository: "repo2", SourceTag: "tag2"},
	}

	mockEstimator.On("GetManifest", ctx, "repo1", "tag1").
		Return(nil, "", fmt.Errorf("error1"))
	mockEstimator.On("GetManifest", ctx, "repo2", "tag2").
		Return(nil, "", fmt.Errorf("error2"))

	totalSize, err := EstimateBatchSize(ctx, tasks, mockEstimator)

	require.NoError(t, err)
	assert.Equal(t, int64(0), totalSize)
	mockEstimator.AssertExpectations(t)
}

// TestEstimateBatchSizes tests individual batch sizes estimation
func TestEstimateBatchSizes_Success(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	tasks := []SyncTask{
		{SourceRepository: "repo1", SourceTag: "tag1"},
		{SourceRepository: "repo2", SourceTag: "tag2"},
		{SourceRepository: "repo3", SourceTag: "tag3"},
	}

	mockEstimator.On("GetManifest", ctx, "repo1", "tag1").
		Return(createOCIManifestJSON(1000, []int64{5000}), "application/vnd.oci.image.manifest.v1+json", nil)
	mockEstimator.On("GetManifest", ctx, "repo2", "tag2").
		Return(createOCIManifestJSON(2000, []int64{6000}), "application/vnd.oci.image.manifest.v1+json", nil)
	mockEstimator.On("GetManifest", ctx, "repo3", "tag3").
		Return(createOCIManifestJSON(3000, []int64{7000}), "application/vnd.oci.image.manifest.v1+json", nil)

	sizes := EstimateBatchSizes(ctx, tasks, mockEstimator)

	require.Len(t, sizes, 3)
	assert.Equal(t, int64(1000+5000), sizes[0])
	assert.Equal(t, int64(2000+6000), sizes[1])
	assert.Equal(t, int64(3000+7000), sizes[2])
	mockEstimator.AssertExpectations(t)
}

func TestEstimateBatchSizes_EmptyBatch(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	tasks := []SyncTask{}

	sizes := EstimateBatchSizes(ctx, tasks, mockEstimator)

	require.Empty(t, sizes)
}

func TestEstimateBatchSizes_SomeEstimationsFail(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	tasks := []SyncTask{
		{SourceRepository: "repo1", SourceTag: "tag1"},
		{SourceRepository: "repo2", SourceTag: "tag2"},
		{SourceRepository: "repo3", SourceTag: "tag3"},
	}

	mockEstimator.On("GetManifest", ctx, "repo1", "tag1").
		Return(createOCIManifestJSON(1000, []int64{5000}), "application/vnd.oci.image.manifest.v1+json", nil)
	mockEstimator.On("GetManifest", ctx, "repo2", "tag2").
		Return(nil, "", fmt.Errorf("error"))
	mockEstimator.On("GetManifest", ctx, "repo3", "tag3").
		Return(createOCIManifestJSON(3000, []int64{7000}), "application/vnd.oci.image.manifest.v1+json", nil)

	sizes := EstimateBatchSizes(ctx, tasks, mockEstimator)

	require.Len(t, sizes, 2)
	assert.Equal(t, int64(1000+5000), sizes[0])
	// Index 1 should be missing (estimation failed)
	_, exists := sizes[1]
	assert.False(t, exists)
	assert.Equal(t, int64(3000+7000), sizes[2])
	mockEstimator.AssertExpectations(t)
}

func TestEstimateBatchSizes_AllEstimationsFail(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	tasks := []SyncTask{
		{SourceRepository: "repo1", SourceTag: "tag1"},
		{SourceRepository: "repo2", SourceTag: "tag2"},
	}

	mockEstimator.On("GetManifest", ctx, "repo1", "tag1").
		Return(nil, "", fmt.Errorf("error1"))
	mockEstimator.On("GetManifest", ctx, "repo2", "tag2").
		Return(nil, "", fmt.Errorf("error2"))

	sizes := EstimateBatchSizes(ctx, tasks, mockEstimator)

	require.Empty(t, sizes)
	mockEstimator.AssertExpectations(t)
}

func TestEstimateBatchSizes_SingleTask(t *testing.T) {
	ctx := context.Background()
	mockEstimator := new(MockSizeEstimator)

	tasks := []SyncTask{
		{SourceRepository: "repo1", SourceTag: "tag1"},
	}

	mockEstimator.On("GetManifest", ctx, "repo1", "tag1").
		Return(createOCIManifestJSON(1000, []int64{5000}), "application/vnd.oci.image.manifest.v1+json", nil)

	sizes := EstimateBatchSizes(ctx, tasks, mockEstimator)

	require.Len(t, sizes, 1)
	assert.Equal(t, int64(1000+5000), sizes[0])
	mockEstimator.AssertExpectations(t)
}

// Edge case tests
func TestEstimateImageSize_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	mockEstimator := new(MockSizeEstimator)
	mockEstimator.On("GetManifest", mock.Anything, "test/repo", "v1.0").
		Return(nil, "", context.Canceled)

	size, err := EstimateImageSize(ctx, mockEstimator, "test/repo", "v1.0")

	require.Error(t, err)
	assert.Equal(t, int64(0), size)
	mockEstimator.AssertExpectations(t)
}

func TestEstimateOCIManifestSize_EmptyJSON(t *testing.T) {
	emptyJSON := []byte("{}")

	size, err := estimateOCIManifestSize(emptyJSON)

	// Should succeed with zero size as the manifest has no config or layers
	require.NoError(t, err)
	assert.Equal(t, int64(0), size)
}

func TestEstimateDockerV2ManifestSize_EmptyJSON(t *testing.T) {
	emptyJSON := []byte("{}")

	size, err := estimateDockerV2ManifestSize(emptyJSON)

	require.NoError(t, err)
	assert.Equal(t, int64(0), size)
}
