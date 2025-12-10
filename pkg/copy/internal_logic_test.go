package copy

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"freightliner/pkg/helper/log"
	"freightliner/pkg/helper/util"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestShouldCompressInternal tests the shouldCompress logic
func TestShouldCompressInternal(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)

	tests := []struct {
		size     int64
		expected bool
	}{
		{0, false},
		{512, false},
		{1023, false},
		{1024, false}, // At threshold
		{1025, true},  // Above threshold
		{10000, true},
	}

	for _, tt := range tests {
		result := copier.shouldCompress(tt.size)
		assert.Equal(t, tt.expected, result, "size: %d", tt.size)
	}
}

// TestEncryptBlobNoManager tests encrypt blob passthrough
func TestEncryptBlobNoManager(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)
	ctx := context.Background()

	data := []byte("test data")
	reader := io.NopCloser(bytes.NewReader(data))

	result, err := copier.encryptBlob(ctx, reader, "registry.io")
	assert.NoError(t, err)
	assert.Equal(t, reader, result, "should pass through when no encryption manager")
}

// TestProcessManifestStub tests the stub method
func TestProcessManifestStub(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)
	ctx := context.Background()

	result, err := copier.processManifest(ctx, nil, nil, nil, nil, nil, false, nil)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

// TestManifestDescriptorImplementation tests manifestDescriptor
func TestManifestDescriptorImplementation(t *testing.T) {
	data := []byte(`{"test": "manifest"}`)
	hash, _ := v1.NewHash("sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	mediaType := types.DockerManifestSchema2

	desc := &manifestDescriptor{
		mediaType: mediaType,
		data:      data,
		hash:      hash,
	}

	t.Run("MediaType", func(t *testing.T) {
		mt, err := desc.MediaType()
		assert.NoError(t, err)
		assert.Equal(t, mediaType, mt)
	})

	t.Run("RawManifest", func(t *testing.T) {
		raw, err := desc.RawManifest()
		assert.NoError(t, err)
		assert.Equal(t, data, raw)
	})

	t.Run("Digest", func(t *testing.T) {
		digest, err := desc.Digest()
		assert.NoError(t, err)
		assert.Equal(t, hash, digest)
	})
}

// TestBlobLayerImplementation tests blobLayer
func TestBlobLayerImplementation(t *testing.T) {
	data := []byte("test blob data")
	hash, _ := v1.NewHash("sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

	layer := &blobLayer{
		digestHash: hash,
		data:       data,
	}

	t.Run("Digest", func(t *testing.T) {
		digest, err := layer.Digest()
		assert.NoError(t, err)
		assert.Equal(t, hash, digest)
	})

	t.Run("DiffID", func(t *testing.T) {
		diffID, err := layer.DiffID()
		assert.NoError(t, err)
		assert.Equal(t, hash, diffID)
	})

	t.Run("Size", func(t *testing.T) {
		size, err := layer.Size()
		assert.NoError(t, err)
		assert.Equal(t, int64(len(data)), size)
	})

	t.Run("MediaType", func(t *testing.T) {
		mt, err := layer.MediaType()
		assert.NoError(t, err)
		assert.Equal(t, types.DockerLayer, mt)
	})

	t.Run("Compressed", func(t *testing.T) {
		reader, err := layer.Compressed()
		assert.NoError(t, err)
		defer reader.Close()

		read, _ := io.ReadAll(reader)
		assert.Equal(t, data, read)
	})

	t.Run("Uncompressed", func(t *testing.T) {
		reader, err := layer.Uncompressed()
		assert.NoError(t, err)
		defer reader.Close()

		read, _ := io.ReadAll(reader)
		assert.Equal(t, data, read)
	})
}

// TestStreamingBlobLayerImplementation tests streamingBlobLayer
func TestStreamingBlobLayerImplementation(t *testing.T) {
	data := []byte("streaming test data")
	hash, _ := v1.NewHash("sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	reader := bytes.NewReader(data)
	bufferMgr := util.NewBufferManager()

	layer := &streamingBlobLayer{
		digestHash: hash,
		reader:     reader,
		bufferMgr:  bufferMgr,
		cachedSize: int64(len(data)),
	}

	t.Run("Digest", func(t *testing.T) {
		digest, err := layer.Digest()
		assert.NoError(t, err)
		assert.Equal(t, hash, digest)
	})

	t.Run("DiffID", func(t *testing.T) {
		diffID, err := layer.DiffID()
		assert.NoError(t, err)
		assert.Equal(t, hash, diffID)
	})

	t.Run("Size with cache", func(t *testing.T) {
		size, err := layer.Size()
		assert.NoError(t, err)
		assert.Equal(t, int64(len(data)), size)
	})

	t.Run("Size without cache", func(t *testing.T) {
		layerNoCache := &streamingBlobLayer{
			digestHash: hash,
			reader:     bytes.NewReader(data),
			bufferMgr:  bufferMgr,
			cachedSize: 0,
		}

		size, err := layerNoCache.Size()
		assert.NoError(t, err)
		assert.Greater(t, size, int64(0))
	})

	t.Run("MediaType", func(t *testing.T) {
		mt, err := layer.MediaType()
		assert.NoError(t, err)
		assert.Equal(t, types.DockerLayer, mt)
	})

	t.Run("Compressed", func(t *testing.T) {
		// Reset reader
		layer.reader = bytes.NewReader(data)
		compressed, err := layer.Compressed()
		assert.NoError(t, err)
		defer compressed.Close()

		buf := make([]byte, 5)
		n, _ := compressed.Read(buf)
		assert.Greater(t, n, 0)
	})

	t.Run("Uncompressed", func(t *testing.T) {
		// Reset reader
		layer.reader = bytes.NewReader(data)
		uncompressed, err := layer.Uncompressed()
		assert.NoError(t, err)
		defer uncompressed.Close()

		buf := make([]byte, 5)
		n, _ := uncompressed.Read(buf)
		assert.Greater(t, n, 0)
	})
}

// TestOptimizedReadCloserImplementation tests optimizedReadCloser
func TestOptimizedReadCloserImplementation(t *testing.T) {
	data := []byte("test read closer data")
	reader := bytes.NewReader(data)
	bufferMgr := util.NewBufferManager()

	t.Run("Read without buffer", func(t *testing.T) {
		orc := &optimizedReadCloser{
			reader:    bytes.NewReader(data),
			bufferMgr: bufferMgr,
		}

		buf := make([]byte, len(data))
		n, err := orc.Read(buf)
		assert.True(t, err == nil || err == io.EOF)
		assert.Equal(t, len(data), n)
		assert.Equal(t, data, buf)
	})

	t.Run("Close without buffer", func(t *testing.T) {
		orc := &optimizedReadCloser{
			reader:    reader,
			bufferMgr: bufferMgr,
		}

		err := orc.Close()
		assert.NoError(t, err)
	})

	t.Run("Close with buffer", func(t *testing.T) {
		buffer := bufferMgr.GetOptimalBuffer(1024, "test")
		orc := &optimizedReadCloser{
			reader:    bytes.NewReader(data),
			bufferMgr: bufferMgr,
			buffer:    buffer,
		}

		err := orc.Close()
		assert.NoError(t, err)
	})

	t.Run("Close with closable reader", func(t *testing.T) {
		closableReader := io.NopCloser(bytes.NewReader(data))
		orc := &optimizedReadCloser{
			reader:    closableReader,
			bufferMgr: bufferMgr,
		}

		err := orc.Close()
		assert.NoError(t, err)
	})

	t.Run("Multiple closes", func(t *testing.T) {
		orc := &optimizedReadCloser{
			reader:    bytes.NewReader(data),
			bufferMgr: bufferMgr,
		}

		err1 := orc.Close()
		assert.NoError(t, err1)

		err2 := orc.Close()
		assert.NoError(t, err2)
	})
}

// TestCopyStatsStructure tests CopyStats
func TestCopyStatsStructure(t *testing.T) {
	stats := &CopyStats{
		BytesTransferred: 1024,
		CompressedBytes:  512,
		PullDuration:     1 * time.Second,
		PushDuration:     2 * time.Second,
		Layers:           3,
		ManifestSize:     256,
	}

	assert.Equal(t, int64(1024), stats.BytesTransferred)
	assert.Equal(t, int64(512), stats.CompressedBytes)
	assert.Equal(t, 1*time.Second, stats.PullDuration)
	assert.Equal(t, 2*time.Second, stats.PushDuration)
	assert.Equal(t, 3, stats.Layers)
	assert.Equal(t, int64(256), stats.ManifestSize)
}

// TestCopyResultStructure tests CopyResult
func TestCopyResultStructure(t *testing.T) {
	stats := CopyStats{BytesTransferred: 2048}

	t.Run("Success result", func(t *testing.T) {
		result := &CopyResult{
			Success: true,
			Stats:   stats,
			Error:   nil,
		}

		assert.True(t, result.Success)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(2048), result.Stats.BytesTransferred)
	})

	t.Run("Failure result", func(t *testing.T) {
		testErr := assert.AnError
		result := &CopyResult{
			Success: false,
			Stats:   stats,
			Error:   testErr,
		}

		assert.False(t, result.Success)
		assert.Error(t, result.Error)
	})
}

// TestCopyOptionsStructure tests CopyOptions
func TestCopyOptionsStructure(t *testing.T) {
	options := CopyOptions{
		DryRun:         true,
		ForceOverwrite: false,
	}

	assert.True(t, options.DryRun)
	assert.False(t, options.ForceOverwrite)
}

// TestBlobTransferFuncType tests BlobTransferFunc
func TestBlobTransferFuncType(t *testing.T) {
	called := false
	var capturedSrc, capturedDest string

	transferFunc := BlobTransferFunc(func(ctx context.Context, src, dest string) error {
		called = true
		capturedSrc = src
		capturedDest = dest
		return nil
	})

	ctx := context.Background()
	err := transferFunc(ctx, "source-url", "dest-url")

	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "source-url", capturedSrc)
	assert.Equal(t, "dest-url", capturedDest)
}

// TestBuilderPattern tests the builder pattern
func TestBuilderPattern(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)

	t.Run("Single builder call", func(t *testing.T) {
		copier := NewCopier(logger).WithEncryptionManager(nil)
		assert.NotNil(t, copier)
	})

	t.Run("Chained builder calls", func(t *testing.T) {
		copier := NewCopier(logger).
			WithEncryptionManager(nil).
			WithBlobTransferFunc(nil).
			WithMetrics(nil)
		assert.NotNil(t, copier)
	})

	t.Run("WithBlobTransferFunc non-nil", func(t *testing.T) {
		transferFunc := func(ctx context.Context, src, dest string) error {
			return nil
		}
		copier := NewCopier(logger).WithBlobTransferFunc(transferFunc)
		assert.NotNil(t, copier)
	})
}

// TestMetricsInterface tests the Metrics interface
func TestMetricsInterface(t *testing.T) {
	metrics := &testMetrics{}

	metrics.ReplicationStarted("src", "dest")
	assert.True(t, metrics.startCalled)
	assert.Equal(t, "src", metrics.lastSrc)
	assert.Equal(t, "dest", metrics.lastDest)

	metrics.ReplicationCompleted(1*time.Second, 5, 1024)
	assert.True(t, metrics.completedCalled)
	assert.Equal(t, 1*time.Second, metrics.lastDuration)
	assert.Equal(t, 5, metrics.lastLayers)
	assert.Equal(t, int64(1024), metrics.lastBytes)

	metrics.ReplicationFailed()
	assert.True(t, metrics.failedCalled)
}

// testMetrics implements Metrics for testing
type testMetrics struct {
	startCalled     bool
	completedCalled bool
	failedCalled    bool
	lastSrc         string
	lastDest        string
	lastDuration    time.Duration
	lastLayers      int
	lastBytes       int64
}

func (m *testMetrics) ReplicationStarted(source, destination string) {
	m.startCalled = true
	m.lastSrc = source
	m.lastDest = destination
}

func (m *testMetrics) ReplicationCompleted(duration time.Duration, layerCount int, byteCount int64) {
	m.completedCalled = true
	m.lastDuration = duration
	m.lastLayers = layerCount
	m.lastBytes = byteCount
}

func (m *testMetrics) ReplicationFailed() {
	m.failedCalled = true
}

// TestCompressStreamLogic tests compress stream creation
func TestCompressStreamLogic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping compression test in short mode")
	}

	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)

	data := []byte("test data for compression")
	reader := io.NopCloser(bytes.NewReader(data))

	compressed, err := copier.compressStream(reader)
	require.NoError(t, err)
	defer compressed.Close()

	// Should be able to read from compressed stream
	buf := make([]byte, 100)
	n, _ := compressed.Read(buf)
	assert.Greater(t, n, 0)
}

// TestCopyBlobDeprecatedMethod tests the deprecated method
func TestCopyBlobDeprecatedMethod(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	copier := NewCopier(logger)
	ctx := context.Background()

	_, err := copier.copyBlob(ctx, "src", "dest", "gzip", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deprecated")
}
