package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"

	"github.com/opencontainers/go-digest"
)

// ContentAddressableStore implements content-addressable storage with automatic deduplication
type ContentAddressableStore struct {
	storage    map[digest.Digest]*Blob
	index      *BlobIndex
	backend    StorageBackend
	logger     log.Logger
	mu         sync.RWMutex
	metrics    *CASMetrics
	gcInterval time.Duration
	stopGC     chan struct{}
}

// Blob represents a stored blob with metadata
type Blob struct {
	Digest      digest.Digest
	Data        []byte
	Size        int64
	RefCount    atomic.Int64
	CreatedAt   time.Time
	LastAccess  time.Time
	Compressed  bool
	ContentType string
	Tags        []string
	mu          sync.RWMutex
}

// BlobIndex provides fast lookup for blobs
type BlobIndex struct {
	digestIndex map[digest.Digest]*Blob
	tagIndex    map[string][]digest.Digest
	sizeIndex   map[int64][]digest.Digest
	mu          sync.RWMutex
}

// StorageBackend defines the interface for storage backends
type StorageBackend interface {
	Put(ctx context.Context, d digest.Digest, data []byte) error
	Get(ctx context.Context, d digest.Digest) ([]byte, error)
	Exists(ctx context.Context, d digest.Digest) bool
	Delete(ctx context.Context, d digest.Digest) error
	List(ctx context.Context) ([]digest.Digest, error)
}

// CASMetrics tracks CAS performance metrics
type CASMetrics struct {
	BlobsStored     atomic.Uint64
	BlobsRetrieved  atomic.Uint64
	BlobsDeleted    atomic.Uint64
	DedupHits       atomic.Uint64
	TotalBytes      atomic.Uint64
	DedupSavedBytes atomic.Uint64
	CacheHits       atomic.Uint64
	CacheMisses     atomic.Uint64
	AvgGetLatency   atomic.Uint64 // in microseconds
	AvgPutLatency   atomic.Uint64 // in microseconds
}

// CASConfig holds configuration for CAS
type CASConfig struct {
	Backend      StorageBackend
	Logger       log.Logger
	GCInterval   time.Duration
	EnableCache  bool
	MaxCacheSize int64
}

// NewContentAddressableStore creates a new CAS
func NewContentAddressableStore(config CASConfig) *ContentAddressableStore {
	if config.Logger == nil {
		config.Logger = log.NewBasicLogger(log.InfoLevel)
	}

	if config.GCInterval == 0 {
		config.GCInterval = 1 * time.Hour
	}

	cas := &ContentAddressableStore{
		storage:    make(map[digest.Digest]*Blob),
		index:      newBlobIndex(),
		backend:    config.Backend,
		logger:     config.Logger,
		metrics:    &CASMetrics{},
		gcInterval: config.GCInterval,
		stopGC:     make(chan struct{}),
	}

	// Start garbage collection
	go cas.runGarbageCollection()

	return cas
}

// newBlobIndex creates a new blob index
func newBlobIndex() *BlobIndex {
	return &BlobIndex{
		digestIndex: make(map[digest.Digest]*Blob),
		tagIndex:    make(map[string][]digest.Digest),
		sizeIndex:   make(map[int64][]digest.Digest),
	}
}

// Store stores data with automatic deduplication
func (cas *ContentAddressableStore) Store(ctx context.Context, data []byte) (digest.Digest, error) {
	startTime := time.Now()
	defer func() {
		latency := time.Since(startTime).Microseconds()
		cas.metrics.AvgPutLatency.Store(uint64(latency))
	}()

	// Calculate content hash (SHA256)
	d := digest.SHA256.FromBytes(data)

	// Check if already exists (deduplication!)
	if cas.Exists(ctx, d) {
		cas.metrics.DedupHits.Add(1)
		cas.metrics.DedupSavedBytes.Add(uint64(len(data)))

		// Increment reference count
		cas.mu.RLock()
		if blob, exists := cas.storage[d]; exists {
			blob.RefCount.Add(1)
			blob.UpdateLastAccess()
		}
		cas.mu.RUnlock()

		cas.logger.WithFields(map[string]interface{}{
			"digest": d.String(),
			"size":   len(data),
		}).Debug("Blob already exists, deduplicated")

		return d, nil
	}

	// Create new blob
	blob := &Blob{
		Digest:     d,
		Data:       data,
		Size:       int64(len(data)),
		CreatedAt:  time.Now(),
		LastAccess: time.Now(),
	}
	blob.RefCount.Store(1)

	// Store in memory cache
	cas.mu.Lock()
	cas.storage[d] = blob
	cas.mu.Unlock()

	// Store in backend
	if cas.backend != nil {
		if err := cas.backend.Put(ctx, d, data); err != nil {
			// Remove from memory cache on backend failure
			cas.mu.Lock()
			delete(cas.storage, d)
			cas.mu.Unlock()
			return "", errors.Wrap(err, "failed to store blob in backend")
		}
	}

	// Update index
	cas.index.Add(blob)

	// Update metrics
	cas.metrics.BlobsStored.Add(1)
	cas.metrics.TotalBytes.Add(uint64(len(data)))

	cas.logger.WithFields(map[string]interface{}{
		"digest": d.String(),
		"size":   len(data),
	}).Debug("Blob stored successfully")

	return d, nil
}

// Get retrieves data by digest
func (cas *ContentAddressableStore) Get(ctx context.Context, d digest.Digest) ([]byte, error) {
	startTime := time.Now()
	defer func() {
		latency := time.Since(startTime).Microseconds()
		cas.metrics.AvgGetLatency.Store(uint64(latency))
	}()

	// Check memory cache first
	cas.mu.RLock()
	blob, exists := cas.storage[d]
	cas.mu.RUnlock()

	if exists {
		cas.metrics.CacheHits.Add(1)
		blob.UpdateLastAccess()
		blob.RefCount.Add(1)

		cas.logger.WithFields(map[string]interface{}{
			"digest": d.String(),
			"size":   blob.Size,
			"source": "cache",
		}).Debug("Blob retrieved from cache")

		return blob.Data, nil
	}

	cas.metrics.CacheMisses.Add(1)

	// Fetch from backend
	if cas.backend == nil {
		return nil, errors.NotFoundf("blob not found: %s", d.String())
	}

	data, err := cas.backend.Get(ctx, d)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve blob from backend")
	}

	// Verify digest
	if digest.SHA256.FromBytes(data) != d {
		return nil, errors.New("digest mismatch: data corruption detected")
	}

	// Add to cache
	blob = &Blob{
		Digest:     d,
		Data:       data,
		Size:       int64(len(data)),
		CreatedAt:  time.Now(),
		LastAccess: time.Now(),
	}
	blob.RefCount.Store(1)

	cas.mu.Lock()
	cas.storage[d] = blob
	cas.mu.Unlock()

	cas.index.Add(blob)
	cas.metrics.BlobsRetrieved.Add(1)

	cas.logger.WithFields(map[string]interface{}{
		"digest": d.String(),
		"size":   len(data),
		"source": "backend",
	}).Debug("Blob retrieved from backend")

	return data, nil
}

// GetReader returns a reader for the blob
func (cas *ContentAddressableStore) GetReader(ctx context.Context, d digest.Digest) (io.ReadCloser, error) {
	data, err := cas.Get(ctx, d)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

// Exists checks if a blob exists
func (cas *ContentAddressableStore) Exists(ctx context.Context, d digest.Digest) bool {
	// Check memory cache
	cas.mu.RLock()
	_, exists := cas.storage[d]
	cas.mu.RUnlock()

	if exists {
		return true
	}

	// Check backend
	if cas.backend != nil {
		return cas.backend.Exists(ctx, d)
	}

	return false
}

// Delete removes a blob (decrements ref count)
func (cas *ContentAddressableStore) Delete(ctx context.Context, d digest.Digest) error {
	cas.mu.Lock()
	blob, exists := cas.storage[d]
	if exists {
		refCount := blob.RefCount.Add(-1)
		if refCount <= 0 {
			delete(cas.storage, d)
			cas.index.Remove(d)
			cas.metrics.BlobsDeleted.Add(1)

			// Delete from backend
			if cas.backend != nil {
				if err := cas.backend.Delete(ctx, d); err != nil {
					cas.logger.WithFields(map[string]interface{}{
						"digest": d.String(),
						"error":  err.Error(),
					}).Warn("Failed to delete blob from backend")
				}
			}

			cas.logger.WithFields(map[string]interface{}{
				"digest": d.String(),
			}).Debug("Blob deleted")
		} else {
			cas.logger.WithFields(map[string]interface{}{
				"digest":    d.String(),
				"ref_count": refCount,
			}).Debug("Decremented blob reference count")
		}
	}
	cas.mu.Unlock()

	if !exists {
		return errors.NotFoundf("blob not found: %s", d.String())
	}

	return nil
}

// List returns all blob digests
func (cas *ContentAddressableStore) List(ctx context.Context) ([]digest.Digest, error) {
	cas.mu.RLock()
	defer cas.mu.RUnlock()

	digests := make([]digest.Digest, 0, len(cas.storage))
	for d := range cas.storage {
		digests = append(digests, d)
	}

	return digests, nil
}

// GetMetrics returns CAS metrics
func (cas *ContentAddressableStore) GetMetrics() *CASMetrics {
	return cas.metrics
}

// GetStats returns statistics about the CAS
func (cas *ContentAddressableStore) GetStats() map[string]interface{} {
	cas.mu.RLock()
	blobCount := len(cas.storage)
	cas.mu.RUnlock()

	dedupRate := float64(0)
	if total := cas.metrics.BlobsStored.Load(); total > 0 {
		dedupRate = float64(cas.metrics.DedupHits.Load()) / float64(total) * 100
	}

	return map[string]interface{}{
		"blob_count":        blobCount,
		"total_bytes":       cas.metrics.TotalBytes.Load(),
		"dedup_saved_bytes": cas.metrics.DedupSavedBytes.Load(),
		"dedup_rate":        fmt.Sprintf("%.2f%%", dedupRate),
		"cache_hit_rate":    cas.getCacheHitRate(),
		"avg_get_latency":   fmt.Sprintf("%dµs", cas.metrics.AvgGetLatency.Load()),
		"avg_put_latency":   fmt.Sprintf("%dµs", cas.metrics.AvgPutLatency.Load()),
	}
}

// getCacheHitRate calculates cache hit rate
func (cas *ContentAddressableStore) getCacheHitRate() string {
	hits := cas.metrics.CacheHits.Load()
	misses := cas.metrics.CacheMisses.Load()
	total := hits + misses

	if total == 0 {
		return "0.00%"
	}

	rate := float64(hits) / float64(total) * 100
	return fmt.Sprintf("%.2f%%", rate)
}

// runGarbageCollection runs periodic garbage collection
func (cas *ContentAddressableStore) runGarbageCollection() {
	ticker := time.NewTicker(cas.gcInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cas.garbageCollect()
		case <-cas.stopGC:
			return
		}
	}
}

// garbageCollect removes unreferenced blobs
func (cas *ContentAddressableStore) garbageCollect() {
	cas.mu.Lock()
	defer cas.mu.Unlock()

	before := len(cas.storage)
	collected := 0

	for d, blob := range cas.storage {
		if blob.RefCount.Load() <= 0 {
			delete(cas.storage, d)
			cas.index.Remove(d)
			collected++
		}
	}

	if collected > 0 {
		cas.logger.WithFields(map[string]interface{}{
			"before":    before,
			"after":     len(cas.storage),
			"collected": collected,
		}).Info("Garbage collection completed")
	}
}

// Stop stops the CAS
func (cas *ContentAddressableStore) Stop() {
	close(cas.stopGC)
}

// BlobIndex methods

// Add adds a blob to the index
func (idx *BlobIndex) Add(blob *Blob) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.digestIndex[blob.Digest] = blob

	// Index by size
	idx.sizeIndex[blob.Size] = append(idx.sizeIndex[blob.Size], blob.Digest)

	// Index by tags
	for _, tag := range blob.Tags {
		idx.tagIndex[tag] = append(idx.tagIndex[tag], blob.Digest)
	}
}

// Remove removes a blob from the index
func (idx *BlobIndex) Remove(d digest.Digest) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	blob, exists := idx.digestIndex[d]
	if !exists {
		return
	}

	delete(idx.digestIndex, d)

	// Remove from size index
	sizeDigests := idx.sizeIndex[blob.Size]
	for i, sd := range sizeDigests {
		if sd == d {
			idx.sizeIndex[blob.Size] = append(sizeDigests[:i], sizeDigests[i+1:]...)
			break
		}
	}

	// Remove from tag index
	for _, tag := range blob.Tags {
		tagDigests := idx.tagIndex[tag]
		for i, td := range tagDigests {
			if td == d {
				idx.tagIndex[tag] = append(tagDigests[:i], tagDigests[i+1:]...)
				break
			}
		}
	}
}

// FindByTag finds blobs by tag
func (idx *BlobIndex) FindByTag(tag string) []digest.Digest {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	digests := idx.tagIndex[tag]
	result := make([]digest.Digest, len(digests))
	copy(result, digests)
	return result
}

// Blob methods

// UpdateLastAccess updates the last access time
func (b *Blob) UpdateLastAccess() {
	b.mu.Lock()
	b.LastAccess = time.Now()
	b.mu.Unlock()
}

// GetRefCount returns the current reference count
func (b *Blob) GetRefCount() int64 {
	return b.RefCount.Load()
}
