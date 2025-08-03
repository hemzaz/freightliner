package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"sync/atomic"
	"time"

	"freightliner/pkg/helper/log"
)

// HighPerformanceCache provides an optimized caching layer for container registry operations
type HighPerformanceCache struct {
	// Core cache storage
	manifestCache *LRUCache[string, *CachedManifest]
	blobCache     *LRUCache[string, *CachedBlob]
	tagCache      *LRUCache[string, *CachedTagList]

	// Cache configuration
	config HighPerformanceCacheConfig

	// Performance monitoring
	metrics *CacheMetrics
	logger  log.Logger

	// Lifecycle management
	started atomic.Bool
	stopped atomic.Bool

	// Background operations
	cleanupTicker   *time.Ticker
	cleanupStop     chan struct{}
	metricsReporter *time.Ticker
}

// HighPerformanceCacheConfig configures the cache behavior
type HighPerformanceCacheConfig struct {
	// Cache sizes
	ManifestCacheSize int
	BlobCacheSize     int
	TagCacheSize      int

	// TTL settings
	ManifestTTL time.Duration
	BlobTTL     time.Duration
	TagTTL      time.Duration

	// Performance settings
	EnableMetrics         bool
	MetricsReportInterval time.Duration
	CleanupInterval       time.Duration

	// Memory management
	MaxMemoryUsage int64 // bytes
	EnableEviction bool
}

// DefaultHighPerformanceCacheConfig returns optimized defaults for container registry caching
func DefaultHighPerformanceCacheConfig() HighPerformanceCacheConfig {
	return HighPerformanceCacheConfig{
		// Cache sizes optimized for high throughput
		ManifestCacheSize: 10000, // 10K manifests
		BlobCacheSize:     50000, // 50K blob metadata entries
		TagCacheSize:      5000,  // 5K tag lists

		// TTL settings for container registry data
		ManifestTTL: 1 * time.Hour,    // Manifests change less frequently
		BlobTTL:     6 * time.Hour,    // Blobs are immutable once created
		TagTTL:      15 * time.Minute, // Tags can change more frequently

		// Performance settings
		EnableMetrics:         true,
		MetricsReportInterval: 1 * time.Minute,
		CleanupInterval:       5 * time.Minute,

		// Memory management (500MB cache limit)
		MaxMemoryUsage: 500 * 1024 * 1024,
		EnableEviction: true,
	}
}

// CachedManifest represents a cached container manifest
type CachedManifest struct {
	Repository   string
	Tag          string
	ManifestData []byte
	MediaType    string
	Digest       string

	// Cache metadata
	CachedAt    time.Time
	TTL         time.Duration
	AccessCount atomic.Int64
	LastAccess  atomic.Int64 // Unix timestamp
	Size        int64
}

// CachedBlob represents cached blob metadata
type CachedBlob struct {
	Digest     string
	Repository string
	Size       int64
	MediaType  string
	Exists     bool

	// Performance optimization: store blob location for fast access
	RegistryURL string
	DownloadURL string

	// Cache metadata
	CachedAt    time.Time
	TTL         time.Duration
	AccessCount atomic.Int64
	LastAccess  atomic.Int64
	CacheSize   int64
}

// CachedTagList represents a cached list of tags for a repository
type CachedTagList struct {
	Repository string
	Tags       []string

	// Cache metadata
	CachedAt    time.Time
	TTL         time.Duration
	AccessCount atomic.Int64
	LastAccess  atomic.Int64
	CacheSize   int64
}

// CacheMetrics tracks comprehensive cache performance metrics
type CacheMetrics struct {
	// Hit/miss statistics
	ManifestHits   atomic.Int64
	ManifestMisses atomic.Int64
	BlobHits       atomic.Int64
	BlobMisses     atomic.Int64
	TagHits        atomic.Int64
	TagMisses      atomic.Int64

	// Performance metrics
	TotalOperations  atomic.Int64
	AverageLatencyNs atomic.Int64
	TotalLatencyNs   atomic.Int64

	// Memory usage
	CurrentMemoryUsage atomic.Int64
	PeakMemoryUsage    atomic.Int64

	// Eviction statistics
	ManifestEvictions atomic.Int64
	BlobEvictions     atomic.Int64
	TagEvictions      atomic.Int64

	// Background operation metrics
	CleanupCycles   atomic.Int64
	CleanupDuration atomic.Int64

	mutex sync.RWMutex
}

// NewHighPerformanceCache creates a new high-performance cache
func NewHighPerformanceCache(config HighPerformanceCacheConfig, logger log.Logger) *HighPerformanceCache {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}

	cache := &HighPerformanceCache{
		config:      config,
		logger:      logger,
		metrics:     &CacheMetrics{},
		cleanupStop: make(chan struct{}),
	}

	// Initialize LRU caches
	cache.manifestCache = NewLRUCache[string, *CachedManifest](config.ManifestCacheSize)
	cache.blobCache = NewLRUCache[string, *CachedBlob](config.BlobCacheSize)
	cache.tagCache = NewLRUCache[string, *CachedTagList](config.TagCacheSize)

	return cache
}

// Start starts the cache background operations
func (c *HighPerformanceCache) Start() error {
	if !c.started.CompareAndSwap(false, true) {
		return nil // Already started
	}

	c.logger.WithFields(map[string]interface{}{
		"manifest_cache_size": c.config.ManifestCacheSize,
		"blob_cache_size":     c.config.BlobCacheSize,
		"tag_cache_size":      c.config.TagCacheSize,
		"max_memory_mb":       c.config.MaxMemoryUsage / (1024 * 1024),
	}).Info("Starting high-performance cache")

	// Start cleanup routine
	c.cleanupTicker = time.NewTicker(c.config.CleanupInterval)
	go c.cleanupRoutine()

	// Start metrics reporting
	if c.config.EnableMetrics {
		c.metricsReporter = time.NewTicker(c.config.MetricsReportInterval)
		go c.metricsReportingRoutine()
	}

	return nil
}

// Stop stops the cache and background operations
func (c *HighPerformanceCache) Stop() {
	if !c.stopped.CompareAndSwap(false, true) {
		return // Already stopped
	}

	// Stop background routines
	if c.cleanupTicker != nil {
		c.cleanupTicker.Stop()
	}
	if c.metricsReporter != nil {
		c.metricsReporter.Stop()
	}

	close(c.cleanupStop)

	c.logger.Info("High-performance cache stopped")
}

// GetManifest retrieves a cached manifest
func (c *HighPerformanceCache) GetManifest(repository, tag string) (*CachedManifest, bool) {
	startTime := time.Now()
	defer func() {
		latency := time.Since(startTime).Nanoseconds()
		c.metrics.TotalLatencyNs.Add(latency)
		c.metrics.TotalOperations.Add(1)

		// Update average latency
		totalOps := c.metrics.TotalOperations.Load()
		totalLatency := c.metrics.TotalLatencyNs.Load()
		c.metrics.AverageLatencyNs.Store(totalLatency / totalOps)
	}()

	key := c.manifestKey(repository, tag)

	if manifest, exists := c.manifestCache.Get(key); exists {
		// Check TTL
		if c.isExpired(manifest.CachedAt, manifest.TTL) {
			c.manifestCache.Remove(key)
			c.metrics.ManifestMisses.Add(1)
			return nil, false
		}

		// Update access statistics
		manifest.AccessCount.Add(1)
		manifest.LastAccess.Store(time.Now().Unix())

		c.metrics.ManifestHits.Add(1)
		return manifest, true
	}

	c.metrics.ManifestMisses.Add(1)
	return nil, false
}

// PutManifest stores a manifest in the cache
func (c *HighPerformanceCache) PutManifest(repository, tag string, manifestData []byte, mediaType, digest string) {
	manifest := &CachedManifest{
		Repository:   repository,
		Tag:          tag,
		ManifestData: manifestData,
		MediaType:    mediaType,
		Digest:       digest,
		CachedAt:     time.Now(),
		TTL:          c.config.ManifestTTL,
		Size:         int64(len(manifestData)) + int64(len(repository)) + int64(len(tag)) + int64(len(mediaType)) + int64(len(digest)),
	}

	key := c.manifestKey(repository, tag)

	// Check memory usage before adding
	if c.config.EnableEviction {
		c.enforceMemoryLimit(manifest.Size)
	}

	c.manifestCache.Put(key, manifest)
	c.updateMemoryUsage(manifest.Size)
}

// GetBlob retrieves cached blob metadata
func (c *HighPerformanceCache) GetBlob(digest string) (*CachedBlob, bool) {
	startTime := time.Now()
	defer func() {
		latency := time.Since(startTime).Nanoseconds()
		c.metrics.TotalLatencyNs.Add(latency)
		c.metrics.TotalOperations.Add(1)
	}()

	if blob, exists := c.blobCache.Get(digest); exists {
		// Check TTL
		if c.isExpired(blob.CachedAt, blob.TTL) {
			c.blobCache.Remove(digest)
			c.metrics.BlobMisses.Add(1)
			return nil, false
		}

		// Update access statistics
		blob.AccessCount.Add(1)
		blob.LastAccess.Store(time.Now().Unix())

		c.metrics.BlobHits.Add(1)
		return blob, true
	}

	c.metrics.BlobMisses.Add(1)
	return nil, false
}

// PutBlob stores blob metadata in the cache
func (c *HighPerformanceCache) PutBlob(digest, repository string, size int64, mediaType, registryURL, downloadURL string, exists bool) {
	blob := &CachedBlob{
		Digest:      digest,
		Repository:  repository,
		Size:        size,
		MediaType:   mediaType,
		Exists:      exists,
		RegistryURL: registryURL,
		DownloadURL: downloadURL,
		CachedAt:    time.Now(),
		TTL:         c.config.BlobTTL,
		CacheSize:   int64(len(digest)) + int64(len(repository)) + int64(len(mediaType)) + int64(len(registryURL)) + int64(len(downloadURL)) + 64, // Approximate overhead
	}

	// Check memory usage before adding
	if c.config.EnableEviction {
		c.enforceMemoryLimit(blob.CacheSize)
	}

	c.blobCache.Put(digest, blob)
	c.updateMemoryUsage(blob.CacheSize)
}

// GetTags retrieves cached tag list
func (c *HighPerformanceCache) GetTags(repository string) (*CachedTagList, bool) {
	startTime := time.Now()
	defer func() {
		latency := time.Since(startTime).Nanoseconds()
		c.metrics.TotalLatencyNs.Add(latency)
		c.metrics.TotalOperations.Add(1)
	}()

	if tagList, exists := c.tagCache.Get(repository); exists {
		// Check TTL
		if c.isExpired(tagList.CachedAt, tagList.TTL) {
			c.tagCache.Remove(repository)
			c.metrics.TagMisses.Add(1)
			return nil, false
		}

		// Update access statistics
		tagList.AccessCount.Add(1)
		tagList.LastAccess.Store(time.Now().Unix())

		c.metrics.TagHits.Add(1)
		return tagList, true
	}

	c.metrics.TagMisses.Add(1)
	return nil, false
}

// PutTags stores a tag list in the cache
func (c *HighPerformanceCache) PutTags(repository string, tags []string) {
	// Calculate cache size
	cacheSize := int64(len(repository))
	for _, tag := range tags {
		cacheSize += int64(len(tag))
	}
	cacheSize += int64(len(tags)) * 8 // Approximate slice overhead

	tagList := &CachedTagList{
		Repository: repository,
		Tags:       make([]string, len(tags)),
		CachedAt:   time.Now(),
		TTL:        c.config.TagTTL,
		CacheSize:  cacheSize,
	}

	// Copy tags to avoid external mutation
	copy(tagList.Tags, tags)

	// Check memory usage before adding
	if c.config.EnableEviction {
		c.enforceMemoryLimit(tagList.CacheSize)
	}

	c.tagCache.Put(repository, tagList)
	c.updateMemoryUsage(tagList.CacheSize)
}

// Helper methods

// manifestKey generates a unique key for manifest caching
func (c *HighPerformanceCache) manifestKey(repository, tag string) string {
	hasher := sha256.New()
	hasher.Write([]byte(repository + ":" + tag))
	return hex.EncodeToString(hasher.Sum(nil))
}

// isExpired checks if a cached item has expired
func (c *HighPerformanceCache) isExpired(cachedAt time.Time, ttl time.Duration) bool {
	return time.Since(cachedAt) > ttl
}

// updateMemoryUsage updates the current memory usage metrics
func (c *HighPerformanceCache) updateMemoryUsage(deltaBytes int64) {
	newUsage := c.metrics.CurrentMemoryUsage.Add(deltaBytes)

	// Update peak usage if necessary
	for {
		currentPeak := c.metrics.PeakMemoryUsage.Load()
		if newUsage <= currentPeak {
			break
		}
		if c.metrics.PeakMemoryUsage.CompareAndSwap(currentPeak, newUsage) {
			break
		}
	}
}

// enforceMemoryLimit evicts items if memory usage exceeds the limit
func (c *HighPerformanceCache) enforceMemoryLimit(additionalBytes int64) {
	currentUsage := c.metrics.CurrentMemoryUsage.Load()

	if currentUsage+additionalBytes <= c.config.MaxMemoryUsage {
		return // Within limits
	}

	// Need to evict items - use LRU eviction
	targetReduction := (currentUsage + additionalBytes) - c.config.MaxMemoryUsage

	// Evict from least recently used items across all caches
	c.evictLRUItems(targetReduction)
}

// evictLRUItems evicts least recently used items to free up memory
func (c *HighPerformanceCache) evictLRUItems(targetReduction int64) {
	var freedBytes int64

	// Simple eviction strategy: evict from each cache proportionally
	// In a production system, you'd want a more sophisticated approach

	// Evict from manifest cache
	manifestEvictions := 0
	c.manifestCache.IterateOldest(func(key string, manifest *CachedManifest) bool {
		if freedBytes >= targetReduction {
			return false // Stop iteration
		}

		c.manifestCache.Remove(key)
		freedBytes += manifest.Size
		manifestEvictions++
		return true // Continue iteration
	})
	c.metrics.ManifestEvictions.Add(int64(manifestEvictions))

	// Initialize eviction counters
	blobEvictions := 0
	tagEvictions := 0

	// Evict from blob cache if still needed
	if freedBytes < targetReduction {
		c.blobCache.IterateOldest(func(digest string, blob *CachedBlob) bool {
			if freedBytes >= targetReduction {
				return false
			}

			c.blobCache.Remove(digest)
			freedBytes += blob.CacheSize
			blobEvictions++
			return true
		})
		c.metrics.BlobEvictions.Add(int64(blobEvictions))
	}

	// Evict from tag cache if still needed
	if freedBytes < targetReduction {
		c.tagCache.IterateOldest(func(repository string, tagList *CachedTagList) bool {
			if freedBytes >= targetReduction {
				return false
			}

			c.tagCache.Remove(repository)
			freedBytes += tagList.CacheSize
			tagEvictions++
			return true
		})
		c.metrics.TagEvictions.Add(int64(tagEvictions))
	}

	// Update memory usage
	c.metrics.CurrentMemoryUsage.Add(-freedBytes)

	c.logger.WithFields(map[string]interface{}{
		"freed_bytes_mb":     freedBytes / (1024 * 1024),
		"manifest_evictions": manifestEvictions,
		"blob_evictions":     blobEvictions,
		"tag_evictions":      tagEvictions,
	}).Debug("Evicted cache items to free memory")
}

// cleanupRoutine runs periodic cleanup of expired cache entries
func (c *HighPerformanceCache) cleanupRoutine() {
	for {
		select {
		case <-c.cleanupStop:
			return
		case <-c.cleanupTicker.C:
			c.performCleanup()
		}
	}
}

// performCleanup removes expired entries from all caches
func (c *HighPerformanceCache) performCleanup() {
	startTime := time.Now()
	var cleanedCount int64

	// Clean manifest cache
	c.manifestCache.IterateAll(func(key string, manifest *CachedManifest) bool {
		if c.isExpired(manifest.CachedAt, manifest.TTL) {
			c.manifestCache.Remove(key)
			c.metrics.CurrentMemoryUsage.Add(-manifest.Size)
			cleanedCount++
		}
		return true
	})

	// Clean blob cache
	c.blobCache.IterateAll(func(digest string, blob *CachedBlob) bool {
		if c.isExpired(blob.CachedAt, blob.TTL) {
			c.blobCache.Remove(digest)
			c.metrics.CurrentMemoryUsage.Add(-blob.CacheSize)
			cleanedCount++
		}
		return true
	})

	// Clean tag cache
	c.tagCache.IterateAll(func(repository string, tagList *CachedTagList) bool {
		if c.isExpired(tagList.CachedAt, tagList.TTL) {
			c.tagCache.Remove(repository)
			c.metrics.CurrentMemoryUsage.Add(-tagList.CacheSize)
			cleanedCount++
		}
		return true
	})

	duration := time.Since(startTime)
	c.metrics.CleanupCycles.Add(1)
	c.metrics.CleanupDuration.Add(duration.Nanoseconds())

	if cleanedCount > 0 {
		c.logger.WithFields(map[string]interface{}{
			"cleaned_items": cleanedCount,
			"duration_ms":   duration.Milliseconds(),
		}).Debug("Cache cleanup completed")
	}
}

// metricsReportingRoutine periodically reports cache metrics
func (c *HighPerformanceCache) metricsReportingRoutine() {
	for {
		select {
		case <-c.cleanupStop:
			return
		case <-c.metricsReporter.C:
			c.reportMetrics()
		}
	}
}

// reportMetrics logs comprehensive cache performance metrics
func (c *HighPerformanceCache) reportMetrics() {
	metrics := c.GetMetrics()

	c.logger.WithFields(map[string]interface{}{
		"manifest_hit_rate": metrics.ManifestHitRate,
		"blob_hit_rate":     metrics.BlobHitRate,
		"tag_hit_rate":      metrics.TagHitRate,
		"overall_hit_rate":  metrics.OverallHitRate,
		"memory_usage_mb":   metrics.CurrentMemoryUsageMB,
		"peak_memory_mb":    metrics.PeakMemoryUsageMB,
		"avg_latency_us":    metrics.AverageLatencyMicros,
		"total_operations":  metrics.TotalOperations,
		"total_evictions":   metrics.TotalEvictions,
	}).Info("Cache performance metrics")
}

// GetMetrics returns comprehensive cache metrics
func (c *HighPerformanceCache) GetMetrics() *CachePerformanceMetrics {
	manifestHits := c.metrics.ManifestHits.Load()
	manifestMisses := c.metrics.ManifestMisses.Load()
	blobHits := c.metrics.BlobHits.Load()
	blobMisses := c.metrics.BlobMisses.Load()
	tagHits := c.metrics.TagHits.Load()
	tagMisses := c.metrics.TagMisses.Load()

	totalHits := manifestHits + blobHits + tagHits
	totalMisses := manifestMisses + blobMisses + tagMisses
	totalRequests := totalHits + totalMisses

	metrics := &CachePerformanceMetrics{
		ManifestHits:   manifestHits,
		ManifestMisses: manifestMisses,
		BlobHits:       blobHits,
		BlobMisses:     blobMisses,
		TagHits:        tagHits,
		TagMisses:      tagMisses,

		TotalOperations:      c.metrics.TotalOperations.Load(),
		AverageLatencyMicros: c.metrics.AverageLatencyNs.Load() / 1000,

		CurrentMemoryUsageMB: c.metrics.CurrentMemoryUsage.Load() / (1024 * 1024),
		PeakMemoryUsageMB:    c.metrics.PeakMemoryUsage.Load() / (1024 * 1024),

		TotalEvictions: c.metrics.ManifestEvictions.Load() + c.metrics.BlobEvictions.Load() + c.metrics.TagEvictions.Load(),
	}

	// Calculate hit rates
	if manifestHits+manifestMisses > 0 {
		metrics.ManifestHitRate = float64(manifestHits) / float64(manifestHits+manifestMisses)
	}
	if blobHits+blobMisses > 0 {
		metrics.BlobHitRate = float64(blobHits) / float64(blobHits+blobMisses)
	}
	if tagHits+tagMisses > 0 {
		metrics.TagHitRate = float64(tagHits) / float64(tagHits+tagMisses)
	}
	if totalRequests > 0 {
		metrics.OverallHitRate = float64(totalHits) / float64(totalRequests)
	}

	return metrics
}

// CachePerformanceMetrics contains comprehensive cache performance data
type CachePerformanceMetrics struct {
	// Hit/miss statistics
	ManifestHits   int64
	ManifestMisses int64
	BlobHits       int64
	BlobMisses     int64
	TagHits        int64
	TagMisses      int64

	// Hit rates
	ManifestHitRate float64
	BlobHitRate     float64
	TagHitRate      float64
	OverallHitRate  float64

	// Performance metrics
	TotalOperations      int64
	AverageLatencyMicros int64

	// Memory metrics
	CurrentMemoryUsageMB int64
	PeakMemoryUsageMB    int64

	// Eviction metrics
	TotalEvictions int64
}
