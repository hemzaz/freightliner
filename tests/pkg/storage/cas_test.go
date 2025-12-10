package storage_test

import (
	"context"
	"testing"
	"time"

	"freightliner/pkg/helper/log"
	"freightliner/pkg/storage"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCAS_Store(t *testing.T) {
	cas := storage.NewContentAddressableStore(storage.CASConfig{
		Logger: log.NewBasicLogger(log.InfoLevel),
	})
	defer cas.Stop()

	ctx := context.Background()
	data := []byte("test blob data")

	d, err := cas.Store(ctx, data)
	require.NoError(t, err)
	assert.NotEmpty(t, d)

	// Verify digest is correct
	expectedDigest := digest.SHA256.FromBytes(data)
	assert.Equal(t, expectedDigest, d)
}

func TestCAS_Get(t *testing.T) {
	cas := storage.NewContentAddressableStore(storage.CASConfig{
		Logger: log.NewBasicLogger(log.InfoLevel),
	})
	defer cas.Stop()

	ctx := context.Background()
	data := []byte("test blob data")

	// Store
	d, err := cas.Store(ctx, data)
	require.NoError(t, err)

	// Get
	retrieved, err := cas.Get(ctx, d)
	require.NoError(t, err)
	assert.Equal(t, data, retrieved)
}

func TestCAS_Deduplication(t *testing.T) {
	cas := storage.NewContentAddressableStore(storage.CASConfig{
		Logger: log.NewBasicLogger(log.InfoLevel),
	})
	defer cas.Stop()

	ctx := context.Background()
	data := []byte("duplicate data")

	// Store first time
	d1, err := cas.Store(ctx, data)
	require.NoError(t, err)

	// Store same data again
	d2, err := cas.Store(ctx, data)
	require.NoError(t, err)

	// Should have same digest
	assert.Equal(t, d1, d2)

	// Should count as dedup hit
	metrics := cas.GetMetrics()
	assert.Greater(t, metrics.DedupHits.Load(), uint64(0))
}

func TestCAS_Exists(t *testing.T) {
	cas := storage.NewContentAddressableStore(storage.CASConfig{
		Logger: log.NewBasicLogger(log.InfoLevel),
	})
	defer cas.Stop()

	ctx := context.Background()
	data := []byte("test data")

	// Should not exist initially
	d := digest.SHA256.FromBytes(data)
	assert.False(t, cas.Exists(ctx, d))

	// Store
	cas.Store(ctx, data)

	// Should exist now
	assert.True(t, cas.Exists(ctx, d))
}

func TestCAS_Delete(t *testing.T) {
	cas := storage.NewContentAddressableStore(storage.CASConfig{
		Logger: log.NewBasicLogger(log.InfoLevel),
	})
	defer cas.Stop()

	ctx := context.Background()
	data := []byte("test data")

	// Store
	d, err := cas.Store(ctx, data)
	require.NoError(t, err)

	// Delete
	err = cas.Delete(ctx, d)
	require.NoError(t, err)

	// Should not exist
	assert.False(t, cas.Exists(ctx, d))
}

func TestCAS_ReferenceCount(t *testing.T) {
	cas := storage.NewContentAddressableStore(storage.CASConfig{
		Logger: log.NewBasicLogger(log.InfoLevel),
	})
	defer cas.Stop()

	ctx := context.Background()
	data := []byte("ref counted data")

	// Store multiple times (increments ref count)
	d1, _ := cas.Store(ctx, data)
	d2, _ := cas.Store(ctx, data)
	d3, _ := cas.Store(ctx, data)

	assert.Equal(t, d1, d2)
	assert.Equal(t, d2, d3)

	// Delete once (decrements ref count)
	err := cas.Delete(ctx, d1)
	require.NoError(t, err)

	// Should still exist (ref count > 0)
	assert.True(t, cas.Exists(ctx, d1))

	// Delete twice more
	cas.Delete(ctx, d1)
	cas.Delete(ctx, d1)

	// Now should not exist (ref count = 0)
	assert.False(t, cas.Exists(ctx, d1))
}

func TestCAS_List(t *testing.T) {
	cas := storage.NewContentAddressableStore(storage.CASConfig{
		Logger: log.NewBasicLogger(log.InfoLevel),
	})
	defer cas.Stop()

	ctx := context.Background()

	// Store multiple blobs
	blobs := []string{"blob1", "blob2", "blob3"}
	for _, data := range blobs {
		_, err := cas.Store(ctx, []byte(data))
		require.NoError(t, err)
	}

	// List
	digests, err := cas.List(ctx)
	require.NoError(t, err)
	assert.Len(t, digests, 3)
}

func TestCAS_Metrics(t *testing.T) {
	cas := storage.NewContentAddressableStore(storage.CASConfig{
		Logger: log.NewBasicLogger(log.InfoLevel),
	})
	defer cas.Stop()

	ctx := context.Background()
	data := []byte("metrics test data")

	// Store
	d, _ := cas.Store(ctx, data)

	// Get (cache hit)
	_, err := cas.Get(ctx, d)
	require.NoError(t, err)

	// Get again (another cache hit)
	_, err = cas.Get(ctx, d)
	require.NoError(t, err)

	// Check metrics
	metrics := cas.GetMetrics()
	assert.Equal(t, uint64(1), metrics.BlobsStored.Load())
	// BlobsRetrieved only counts backend retrievals, not cache hits
	// Cache hits are tracked separately
	assert.GreaterOrEqual(t, metrics.CacheHits.Load(), uint64(1))
}

func TestCAS_GetStats(t *testing.T) {
	cas := storage.NewContentAddressableStore(storage.CASConfig{
		Logger: log.NewBasicLogger(log.InfoLevel),
	})
	defer cas.Stop()

	ctx := context.Background()

	// Store some data
	cas.Store(ctx, []byte("data1"))
	cas.Store(ctx, []byte("data2"))
	cas.Store(ctx, []byte("data1")) // Duplicate

	stats := cas.GetStats()
	assert.Contains(t, stats, "blob_count")
	assert.Contains(t, stats, "dedup_rate")
	assert.Contains(t, stats, "cache_hit_rate")
}

func TestCAS_GetReader(t *testing.T) {
	cas := storage.NewContentAddressableStore(storage.CASConfig{
		Logger: log.NewBasicLogger(log.InfoLevel),
	})
	defer cas.Stop()

	ctx := context.Background()
	data := []byte("reader test data")

	// Store
	d, err := cas.Store(ctx, data)
	require.NoError(t, err)

	// Get reader
	reader, err := cas.GetReader(ctx, d)
	require.NoError(t, err)
	defer reader.Close()

	// Read data
	buf := make([]byte, len(data))
	n, err := reader.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, data, buf)
}

func TestCAS_GarbageCollection(t *testing.T) {
	// Fast GC interval for testing
	cas := storage.NewContentAddressableStore(storage.CASConfig{
		Logger:     log.NewBasicLogger(log.InfoLevel),
		GCInterval: 100 * time.Millisecond,
	})
	defer cas.Stop()

	ctx := context.Background()
	data := []byte("gc test data")

	// Store and immediately delete
	d, _ := cas.Store(ctx, data)
	cas.Delete(ctx, d)

	// Wait for GC
	time.Sleep(200 * time.Millisecond)

	// Should be cleaned up
	assert.False(t, cas.Exists(ctx, d))
}
