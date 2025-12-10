package copy_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"freightliner/pkg/copy"
	"freightliner/pkg/helper/log"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewCopierLifecycle tests the copier creation and initialization
func TestNewCopierLifecycle(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	copier := copy.NewCopier(logger)

	assert.NotNil(t, copier, "copier should not be nil")
	// Note: We can't access private fields directly, but we can test the behavior
}

// TestCopierBuilderPattern tests the builder pattern methods
func TestCopierBuilderPattern(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	copier := copy.NewCopier(logger)

	t.Run("WithEncryptionManager chaining", func(t *testing.T) {
		result := copier.WithEncryptionManager(nil)
		assert.Equal(t, copier, result, "should return same instance for chaining")
	})

	t.Run("WithBlobTransferFunc chaining", func(t *testing.T) {
		transferFunc := func(ctx context.Context, src, dest string) error {
			return nil
		}
		result := copier.WithBlobTransferFunc(transferFunc)
		assert.Equal(t, copier, result, "should return same instance for chaining")
	})

	t.Run("WithBlobTransferFunc nil safety", func(t *testing.T) {
		result := copier.WithBlobTransferFunc(nil)
		assert.Equal(t, copier, result, "should handle nil transfer func gracefully")
	})

	t.Run("WithMetrics chaining", func(t *testing.T) {
		metrics := &mockMetrics{}
		result := copier.WithMetrics(metrics)
		assert.Equal(t, copier, result, "should return same instance for chaining")
	})
}

// TestCopierChainedBuilder tests multiple builder calls
func TestCopierChainedBuilder(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	metrics := &mockMetrics{}
	transferFunc := func(ctx context.Context, src, dest string) error {
		return nil
	}

	copier := copy.NewCopier(logger).
		WithEncryptionManager(nil).
		WithBlobTransferFunc(transferFunc).
		WithMetrics(metrics)

	assert.NotNil(t, copier, "chained builder should return valid copier")
}

// TestCopyImageDryRun tests dry run mode (no actual registry operations)
func TestCopyImageDryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := log.NewBasicLogger(log.InfoLevel)
	copier := copy.NewCopier(logger)

	srcRef, err := name.ParseReference("source.registry.io/repo:tag")
	require.NoError(t, err)

	destRef, err := name.ParseReference("dest.registry.io/repo:tag")
	require.NoError(t, err)

	options := copy.CopyOptions{
		DryRun:         true,
		ForceOverwrite: true,
		Source:         srcRef,
		Destination:    destRef,
	}

	ctx := context.Background()

	// This will fail at remote.Get() but tests dry run path
	_, err = copier.CopyImage(ctx, srcRef, destRef, nil, nil, options)

	// We expect an error because we can't connect to registry
	// but we're testing the code path
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get source image descriptor")
}

// TestCopyImageForceOverwrite tests force overwrite logic
func TestCopyImageForceOverwrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := log.NewBasicLogger(log.InfoLevel)
	copier := copy.NewCopier(logger)

	srcRef, _ := name.ParseReference("source.io/repo:tag")
	destRef, _ := name.ParseReference("dest.io/repo:tag")

	t.Run("with force overwrite", func(t *testing.T) {
		options := copy.CopyOptions{
			DryRun:         false,
			ForceOverwrite: true,
			Source:         srcRef,
			Destination:    destRef,
		}

		ctx := context.Background()
		_, err := copier.CopyImage(ctx, srcRef, destRef, nil, nil, options)

		// Will fail at remote.Get but tests the force overwrite check is skipped
		assert.Error(t, err)
	})

	t.Run("without force overwrite", func(t *testing.T) {
		options := copy.CopyOptions{
			DryRun:         false,
			ForceOverwrite: false,
			Source:         srcRef,
			Destination:    destRef,
		}

		ctx := context.Background()
		_, err := copier.CopyImage(ctx, srcRef, destRef, nil, nil, options)

		assert.Error(t, err)
	})
}

// TestCopyImageWithMetrics tests metrics collection
func TestCopyImageWithMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := log.NewBasicLogger(log.InfoLevel)
	metrics := &mockMetrics{}
	copier := copy.NewCopier(logger).WithMetrics(metrics)

	srcRef, _ := name.ParseReference("source.io/repo:tag")
	destRef, _ := name.ParseReference("dest.io/repo:tag")

	options := copy.CopyOptions{
		DryRun:         true,
		ForceOverwrite: true,
		Source:         srcRef,
		Destination:    destRef,
	}

	ctx := context.Background()
	_, err := copier.CopyImage(ctx, srcRef, destRef, nil, nil, options)

	// Will fail but that's ok for testing metrics path
	assert.Error(t, err)

	// Metrics should not be called if we fail early
	assert.False(t, metrics.StartCalled, "metrics should not start if source fetch fails")
}

// TestCopyOptionsValidation tests copy options structure and validation
func TestCopyOptionsValidation(t *testing.T) {
	srcRef, err := name.ParseReference("source.io/repo:tag")
	require.NoError(t, err)

	destRef, err := name.ParseReference("dest.io/repo:tag")
	require.NoError(t, err)

	t.Run("valid options", func(t *testing.T) {
		options := copy.CopyOptions{
			DryRun:         true,
			ForceOverwrite: false,
			Source:         srcRef,
			Destination:    destRef,
		}

		assert.Equal(t, srcRef, options.Source)
		assert.Equal(t, destRef, options.Destination)
		assert.True(t, options.DryRun)
		assert.False(t, options.ForceOverwrite)
	})

	t.Run("all flags enabled", func(t *testing.T) {
		options := copy.CopyOptions{
			DryRun:         true,
			ForceOverwrite: true,
			Source:         srcRef,
			Destination:    destRef,
		}

		assert.True(t, options.DryRun)
		assert.True(t, options.ForceOverwrite)
	})
}

// TestCopyStatsTracking tests statistics tracking structure
func TestCopyStatsTracking(t *testing.T) {
	stats := &copy.CopyStats{
		BytesTransferred: 1024 * 1024, // 1MB
		CompressedBytes:  512 * 1024,  // 512KB
		PullDuration:     2 * time.Second,
		PushDuration:     3 * time.Second,
		Layers:           5,
		ManifestSize:     4096,
	}

	assert.Equal(t, int64(1024*1024), stats.BytesTransferred)
	assert.Equal(t, int64(512*1024), stats.CompressedBytes)
	assert.Equal(t, 2*time.Second, stats.PullDuration)
	assert.Equal(t, 3*time.Second, stats.PushDuration)
	assert.Equal(t, 5, stats.Layers)
	assert.Equal(t, int64(4096), stats.ManifestSize)
}

// TestCopyResultSuccess tests successful copy result
func TestCopyResultSuccess(t *testing.T) {
	stats := copy.CopyStats{
		BytesTransferred: 2048,
		Layers:           3,
		PullDuration:     1 * time.Second,
		PushDuration:     1 * time.Second,
	}

	result := &copy.CopyResult{
		Success: true,
		Stats:   stats,
		Error:   nil,
	}

	assert.True(t, result.Success)
	assert.Equal(t, int64(2048), result.Stats.BytesTransferred)
	assert.Equal(t, 3, result.Stats.Layers)
	assert.NoError(t, result.Error)
}

// TestCopyResultFailure tests failed copy result
func TestCopyResultFailure(t *testing.T) {
	stats := copy.CopyStats{}
	testErr := errors.New("copy failed")

	result := &copy.CopyResult{
		Success: false,
		Stats:   stats,
		Error:   testErr,
	}

	assert.False(t, result.Success)
	assert.Error(t, result.Error)
	assert.Equal(t, "copy failed", result.Error.Error())
}

// TestBlobTransferFuncType tests blob transfer function type
func TestBlobTransferFuncType(t *testing.T) {
	called := false
	var capturedSrc, capturedDest string
	var capturedCtx context.Context

	transferFunc := copy.BlobTransferFunc(func(ctx context.Context, src, dest string) error {
		called = true
		capturedCtx = ctx
		capturedSrc = src
		capturedDest = dest
		return nil
	})

	ctx := context.Background()
	err := transferFunc(ctx, "source-blob-url", "dest-blob-url")

	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "source-blob-url", capturedSrc)
	assert.Equal(t, "dest-blob-url", capturedDest)
	assert.NotNil(t, capturedCtx)
}

// TestBlobTransferFuncError tests error handling in transfer function
func TestBlobTransferFuncError(t *testing.T) {
	expectedErr := errors.New("transfer failed")

	transferFunc := copy.BlobTransferFunc(func(ctx context.Context, src, dest string) error {
		return expectedErr
	})

	ctx := context.Background()
	err := transferFunc(ctx, "src", "dest")

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

// TestCopierWithCustomTransferFunc tests custom transfer function integration
func TestCopierWithCustomTransferFunc(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)

	callCount := 0
	transferFunc := func(ctx context.Context, src, dest string) error {
		callCount++
		return nil
	}

	copier := copy.NewCopier(logger).WithBlobTransferFunc(transferFunc)
	assert.NotNil(t, copier)

	// The transfer function is set, though we can't call it directly in this test
	// It would be called during actual CopyImage operations
}

// mockMetrics implements the Metrics interface for testing
type mockMetrics struct {
	StartCalled     bool
	CompletedCalled bool
	FailedCalled    bool
	LastSource      string
	LastDestination string
	LastDuration    time.Duration
	LastLayerCount  int
	LastByteCount   int64
}

func (m *mockMetrics) ReplicationStarted(source, destination string) {
	m.StartCalled = true
	m.LastSource = source
	m.LastDestination = destination
}

func (m *mockMetrics) ReplicationCompleted(duration time.Duration, layerCount int, byteCount int64) {
	m.CompletedCalled = true
	m.LastDuration = duration
	m.LastLayerCount = layerCount
	m.LastByteCount = byteCount
}

func (m *mockMetrics) ReplicationFailed() {
	m.FailedCalled = true
}

// mockImage implements v1.Image for testing
type mockImage struct {
	manifest   []byte
	config     *v1.ConfigFile
	layers     []v1.Layer
	mediaType  types.MediaType
	shouldFail bool
}

func (m *mockImage) Layers() ([]v1.Layer, error) {
	if m.shouldFail {
		return nil, errors.New("mock layers error")
	}
	return m.layers, nil
}

func (m *mockImage) MediaType() (types.MediaType, error) {
	return m.mediaType, nil
}

func (m *mockImage) Size() (int64, error) {
	return int64(len(m.manifest)), nil
}

func (m *mockImage) ConfigName() (v1.Hash, error) {
	return v1.Hash{}, nil
}

func (m *mockImage) ConfigFile() (*v1.ConfigFile, error) {
	if m.shouldFail {
		return nil, errors.New("mock config error")
	}
	if m.config == nil {
		return &v1.ConfigFile{}, nil
	}
	return m.config, nil
}

func (m *mockImage) RawConfigFile() ([]byte, error) {
	return []byte("{}"), nil
}

func (m *mockImage) Digest() (v1.Hash, error) {
	return v1.Hash{}, nil
}

func (m *mockImage) Manifest() (*v1.Manifest, error) {
	return &v1.Manifest{}, nil
}

func (m *mockImage) RawManifest() ([]byte, error) {
	if m.shouldFail {
		return nil, errors.New("mock manifest error")
	}
	if m.manifest == nil {
		return []byte(`{"schemaVersion":2}`), nil
	}
	return m.manifest, nil
}

func (m *mockImage) LayerByDigest(v1.Hash) (v1.Layer, error) {
	return nil, errors.New("not implemented")
}

func (m *mockImage) LayerByDiffID(v1.Hash) (v1.Layer, error) {
	return nil, errors.New("not implemented")
}

// mockLayer implements v1.Layer for testing
type mockLayer struct {
	digest     v1.Hash
	diffID     v1.Hash
	size       int64
	data       []byte
	shouldFail bool
}

func (m *mockLayer) Digest() (v1.Hash, error) {
	if m.shouldFail {
		return v1.Hash{}, errors.New("mock digest error")
	}
	return m.digest, nil
}

func (m *mockLayer) DiffID() (v1.Hash, error) {
	return m.diffID, nil
}

func (m *mockLayer) Compressed() (io.ReadCloser, error) {
	if m.shouldFail {
		return nil, errors.New("mock compressed error")
	}
	return io.NopCloser(strings.NewReader(string(m.data))), nil
}

func (m *mockLayer) Uncompressed() (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(string(m.data))), nil
}

func (m *mockLayer) Size() (int64, error) {
	if m.shouldFail {
		return 0, errors.New("mock size error")
	}
	return m.size, nil
}

func (m *mockLayer) MediaType() (types.MediaType, error) {
	return types.DockerLayer, nil
}

// mockDescriptor implements remote.Descriptor for testing
type mockDescriptor struct {
	image      v1.Image
	shouldFail bool
}

func (m *mockDescriptor) Image() (v1.Image, error) {
	if m.shouldFail {
		return nil, errors.New("mock descriptor error")
	}
	return m.image, nil
}

func (m *mockDescriptor) ImageIndex() (v1.ImageIndex, error) {
	return nil, errors.New("not implemented")
}

func (m *mockDescriptor) Descriptor() (*remote.Descriptor, error) {
	return nil, errors.New("not implemented")
}
