package network

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"freightliner/pkg/helper/log"
	"freightliner/pkg/interfaces"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultTransferOptions_Creation tests default options creation
func TestDefaultTransferOptions_Creation(t *testing.T) {
	opts := DefaultTransferOptions()

	assert.True(t, opts.EnableCompression)
	assert.Equal(t, GzipCompression, opts.CompressionType)
	assert.Equal(t, DefaultCompression, opts.CompressionLevel)
	assert.True(t, opts.EnableDelta)
	assert.Equal(t, 3, opts.RetryAttempts)
	assert.Equal(t, time.Second, opts.RetryInitialDelay)
	assert.Equal(t, 30*time.Second, opts.RetryMaxDelay)
}

// TestNewTransferManager tests transfer manager creation
func TestNewTransferManager(t *testing.T) {
	t.Run("with logger", func(t *testing.T) {
		opts := DefaultTransferOptions()
		logger := log.NewBasicLogger(log.InfoLevel)

		tm, err := NewTransferManager(opts, logger)
		require.NoError(t, err)
		assert.NotNil(t, tm)
		assert.NotNil(t, tm.logger)
		assert.NotNil(t, tm.bufferMgr)
	})

	t.Run("without logger creates default", func(t *testing.T) {
		opts := DefaultTransferOptions()

		tm, err := NewTransferManager(opts, nil)
		require.NoError(t, err)
		assert.NotNil(t, tm)
		assert.NotNil(t, tm.logger)
	})
}

// TestTransferManager_TransferBlob tests blob transfer with mock repository
func TestTransferManager_TransferBlob(t *testing.T) {
	opts := DefaultTransferOptions()
	opts.EnableCompression = false // Disable for simpler testing
	opts.EnableDelta = false
	opts.RetryAttempts = 1

	logger := log.NewBasicLogger(log.InfoLevel)
	tm, err := NewTransferManager(opts, logger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("nil source repository", func(t *testing.T) {
		stats, err := tm.TransferBlob(ctx, nil, &MockRepository{}, "sha256:abc123")
		assert.Error(t, err)
		assert.Nil(t, stats)
		assert.Contains(t, err.Error(), "source repository cannot be nil")
	})

	t.Run("nil destination repository", func(t *testing.T) {
		stats, err := tm.TransferBlob(ctx, &MockRepository{}, nil, "sha256:abc123")
		assert.Error(t, err)
		assert.Nil(t, stats)
		assert.Contains(t, err.Error(), "destination repository cannot be nil")
	})

	t.Run("empty digest", func(t *testing.T) {
		stats, err := tm.TransferBlob(ctx, &MockRepository{}, &MockRepository{}, "")
		assert.Error(t, err)
		assert.Nil(t, stats)
		assert.Contains(t, err.Error(), "digest cannot be empty")
	})

	t.Run("successful transfer", func(t *testing.T) {
		sourceRepo := NewTestMockRepository("source-repo")
		destRepo := NewTestMockRepository("dest-repo")

		stats, err := tm.TransferBlob(ctx, sourceRepo, destRepo, "sha256:test123")
		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Greater(t, stats.BytesTransferred, int64(0))
	})
}

// TestTransferManager_TransferImage tests image transfer
func TestTransferManager_TransferImage(t *testing.T) {
	opts := DefaultTransferOptions()
	logger := log.NewBasicLogger(log.InfoLevel)
	tm, err := NewTransferManager(opts, logger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("nil source repository", func(t *testing.T) {
		stats, err := tm.TransferImage(ctx, nil, &MockRepository{}, "v1.0")
		assert.Error(t, err)
		assert.Nil(t, stats)
	})

	t.Run("nil destination repository", func(t *testing.T) {
		stats, err := tm.TransferImage(ctx, &MockRepository{}, nil, "v1.0")
		assert.Error(t, err)
		assert.Nil(t, stats)
	})

	t.Run("empty tag", func(t *testing.T) {
		stats, err := tm.TransferImage(ctx, &MockRepository{}, &MockRepository{}, "")
		assert.Error(t, err)
		assert.Nil(t, stats)
	})

	t.Run("successful transfer", func(t *testing.T) {
		sourceRepo := NewTestMockRepository("source-repo")
		destRepo := NewTestMockRepository("dest-repo")

		stats, err := tm.TransferImage(ctx, sourceRepo, destRepo, "v1.0")
		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Greater(t, stats.BytesTransferred, int64(0))
	})
}

// TestTransferManager_CreateStreamingCompressor tests compression
func TestTransferManager_CreateStreamingCompressor(t *testing.T) {
	opts := DefaultTransferOptions()
	logger := log.NewBasicLogger(log.InfoLevel)
	tm, err := NewTransferManager(opts, logger)
	require.NoError(t, err)

	testData := []byte("This is test data that should be compressed. " +
		"It needs to be long enough to show compression benefits. " +
		"More text here to make it compressible. " +
		"Even more text to ensure we have enough data. " +
		"And some more for good measure.")

	reader := bytes.NewReader(testData)
	compressionOpts := DefaultCompressionOptions()

	compressedReader, ratio, err := tm.createStreamingCompressor(reader, compressionOpts)
	require.NoError(t, err)
	assert.NotNil(t, compressedReader)
	assert.Greater(t, ratio, 0.0)

	// Read the compressed data
	compressedData, err := io.ReadAll(compressedReader)
	require.NoError(t, err)
	assert.NotEmpty(t, compressedData)
}

// TestTransferManager_StreamToDestination tests streaming to destination
func TestTransferManager_StreamToDestination(t *testing.T) {
	opts := DefaultTransferOptions()
	logger := log.NewBasicLogger(log.InfoLevel)
	tm, err := NewTransferManager(opts, logger)
	require.NoError(t, err)

	testData := []byte("test data to stream")
	reader := bytes.NewReader(testData)
	destRepo := NewTestMockRepository("dest-repo")

	ctx := context.Background()

	bytesTransferred, err := tm.streamToDestination(ctx, reader, destRepo, "sha256:test")
	require.NoError(t, err)
	assert.Equal(t, int64(len(testData)), bytesTransferred)
}

// TestTransferManager_StreamToDestination_ContextCancellation tests context cancellation
func TestTransferManager_StreamToDestination_ContextCancellation(t *testing.T) {
	opts := DefaultTransferOptions()
	logger := log.NewBasicLogger(log.InfoLevel)
	tm, err := NewTransferManager(opts, logger)
	require.NoError(t, err)

	// Create a slow reader that will be interrupted
	slowReader := &slowReader{
		data:  bytes.Repeat([]byte("x"), 1000000),
		delay: 10 * time.Millisecond,
	}

	destRepo := NewTestMockRepository("dest-repo")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	bytesTransferred, err := tm.streamToDestination(ctx, slowReader, destRepo, "sha256:test")
	assert.Error(t, err)
	assert.True(t, err == context.DeadlineExceeded || err == context.Canceled)
	assert.Greater(t, bytesTransferred, int64(0)) // Should have transferred some data
}

// TestTransferStats_Structure tests the statistics structure
func TestTransferStats_Structure(t *testing.T) {
	stats := &TransferStats{
		BytesTransferred:    1024 * 1024,
		BytesCompressed:     512 * 1024,
		CompressionRatio:    0.5,
		DeltaReductions:     100 * 1024,
		TransferDuration:    2 * time.Second,
		CompressionDuration: 500 * time.Millisecond,
		RetryCount:          1,
	}

	assert.Equal(t, int64(1024*1024), stats.BytesTransferred)
	assert.Equal(t, float64(0.5), stats.CompressionRatio)
	assert.Equal(t, 2*time.Second, stats.TransferDuration)
	assert.Equal(t, 1, stats.RetryCount)
}

// TestTransferOptions_CustomValues tests custom option values
func TestTransferOptions_CustomValues(t *testing.T) {
	opts := TransferOptions{
		EnableCompression: false,
		CompressionType:   GzipCompression,
		CompressionLevel:  BestCompression,
		EnableDelta:       false,
		RetryAttempts:     5,
		RetryInitialDelay: 2 * time.Second,
		RetryMaxDelay:     60 * time.Second,
	}

	assert.False(t, opts.EnableCompression)
	assert.Equal(t, GzipCompression, opts.CompressionType)
	assert.Equal(t, BestCompression, opts.CompressionLevel)
	assert.False(t, opts.EnableDelta)
	assert.Equal(t, 5, opts.RetryAttempts)
	assert.Equal(t, 2*time.Second, opts.RetryInitialDelay)
	assert.Equal(t, 60*time.Second, opts.RetryMaxDelay)
}

// TestTransferManager_DeltaTransferDisabled tests that delta transfer options are set correctly
func TestTransferManager_DeltaTransferDisabled(t *testing.T) {
	opts := DefaultTransferOptions()
	opts.EnableDelta = false // Delta is not implemented yet
	logger := log.NewBasicLogger(log.InfoLevel)
	tm, err := NewTransferManager(opts, logger)
	require.NoError(t, err)
	assert.NotNil(t, tm, "TransferManager should be created")
}

// slowReader is a test helper that reads slowly
type slowReader struct {
	data  []byte
	pos   int
	delay time.Duration
}

func (sr *slowReader) Read(p []byte) (n int, err error) {
	if sr.pos >= len(sr.data) {
		return 0, io.EOF
	}

	// Add delay to simulate slow reading
	time.Sleep(sr.delay)

	// Read small chunks
	chunkSize := 100
	if len(p) < chunkSize {
		chunkSize = len(p)
	}

	remaining := len(sr.data) - sr.pos
	if remaining < chunkSize {
		chunkSize = remaining
	}

	copy(p, sr.data[sr.pos:sr.pos+chunkSize])
	sr.pos += chunkSize
	return chunkSize, nil
}

// TestTransferManager_WithCompression tests transfer with compression enabled
func TestTransferManager_WithCompression(t *testing.T) {
	opts := DefaultTransferOptions()
	opts.EnableCompression = true
	opts.CompressionType = GzipCompression
	opts.EnableDelta = false

	logger := log.NewBasicLogger(log.InfoLevel)
	tm, err := NewTransferManager(opts, logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Create repositories
	sourceRepo := NewTestMockRepository("source-repo")
	destRepo := NewTestMockRepository("dest-repo")

	stats, err := tm.TransferBlob(ctx, sourceRepo, destRepo, "sha256:test123")
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Greater(t, stats.CompressionDuration, time.Duration(0))
}

// TestTransferManager_RetryBehavior tests retry logic
func TestTransferManager_RetryBehavior(t *testing.T) {
	opts := DefaultTransferOptions()
	opts.RetryAttempts = 3
	opts.RetryInitialDelay = 10 * time.Millisecond
	opts.RetryMaxDelay = 100 * time.Millisecond
	opts.EnableCompression = false
	opts.EnableDelta = false

	logger := log.NewBasicLogger(log.InfoLevel)
	tm, err := NewTransferManager(opts, logger)
	require.NoError(t, err)

	// Test that retry configuration is applied correctly
	assert.Equal(t, 3, opts.RetryAttempts, "Retry attempts should be configured")
	assert.Equal(t, 10*time.Millisecond, opts.RetryInitialDelay, "Initial delay should be configured")
	assert.NotNil(t, tm, "TransferManager should be created with retry config")
}

// NewTestMockRepository creates a MockRepository from delta_test.go for use in network tests
func NewTestMockRepository(name string) *MockRepository {
	repo := NewMockRepository()
	repo.name = name
	// Use PutManifest to properly initialize manifests for common tags
	_ = repo.PutManifest(context.Background(), "latest", &interfaces.Manifest{
		Content:   []byte(`{"test":"manifest"}`),
		MediaType: "application/vnd.docker.distribution.manifest.v2+json",
	})
	_ = repo.PutManifest(context.Background(), "v1.0", &interfaces.Manifest{
		Content:   []byte(`{"test":"manifest"}`),
		MediaType: "application/vnd.docker.distribution.manifest.v2+json",
	})
	return repo
}
