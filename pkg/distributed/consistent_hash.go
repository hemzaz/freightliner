package distributed

import (
	"context"
	"hash/crc32"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
)

// DistributedCache implements a distributed cache using consistent hashing
type DistributedCache struct {
	ring        *ConsistentHashRing
	nodes       map[string]*CacheNode
	logger      log.Logger
	mu          sync.RWMutex
	replication int // Number of replicas for fault tolerance
	metrics     *CacheMetrics
}

// ConsistentHashRing implements consistent hashing
type ConsistentHashRing struct {
	circle     map[uint32]string // Hash -> NodeID
	sortedKeys []uint32
	nodes      map[string]*NodeInfo
	vnodes     int // Virtual nodes per physical node
	mu         sync.RWMutex
}

// CacheNode represents a cache node
type CacheNode struct {
	ID       string
	Address  string
	cache    map[string]*CacheEntry
	capacity int64
	size     atomic.Int64
	client   CacheClient
	mu       sync.RWMutex
	healthy  atomic.Bool
}

// CacheEntry represents a cached value
type CacheEntry struct {
	Key        string
	Value      []byte
	Size       int64
	CreatedAt  int64
	ExpiresAt  int64
	AccessTime atomic.Int64
	HitCount   atomic.Int64
}

// NodeInfo contains node metadata
type NodeInfo struct {
	ID       string
	Address  string
	Weight   int
	Capacity int64
}

// CacheClient defines the interface for cache communication
type CacheClient interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl int64) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// CacheMetrics tracks cache performance
type CacheMetrics struct {
	Gets        atomic.Uint64
	Sets        atomic.Uint64
	Deletes     atomic.Uint64
	Hits        atomic.Uint64
	Misses      atomic.Uint64
	Evictions   atomic.Uint64
	Relocations atomic.Uint64 // Keys moved due to node changes
}

// CacheConfig holds distributed cache configuration
type CacheConfig struct {
	Logger       log.Logger
	VirtualNodes int
	Replication  int
}

// NewDistributedCache creates a new distributed cache
func NewDistributedCache(config CacheConfig) *DistributedCache {
	if config.Logger == nil {
		config.Logger = log.NewBasicLogger(log.InfoLevel)
	}

	if config.VirtualNodes == 0 {
		config.VirtualNodes = 150 // Default: 150 virtual nodes per physical node
	}

	if config.Replication == 0 {
		config.Replication = 3 // Default: 3 replicas
	}

	return &DistributedCache{
		ring:        NewConsistentHashRing(config.VirtualNodes),
		nodes:       make(map[string]*CacheNode),
		logger:      config.Logger,
		replication: config.Replication,
		metrics:     &CacheMetrics{},
	}
}

// NewConsistentHashRing creates a new consistent hash ring
func NewConsistentHashRing(vnodes int) *ConsistentHashRing {
	return &ConsistentHashRing{
		circle:     make(map[uint32]string),
		sortedKeys: make([]uint32, 0),
		nodes:      make(map[string]*NodeInfo),
		vnodes:     vnodes,
	}
}

// AddNode adds a node to the distributed cache
func (dc *DistributedCache) AddNode(node *CacheNode) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if _, exists := dc.nodes[node.ID]; exists {
		return errors.AlreadyExistsf("node already exists: %s", node.ID)
	}

	dc.nodes[node.ID] = node
	node.healthy.Store(true)

	nodeInfo := &NodeInfo{
		ID:       node.ID,
		Address:  node.Address,
		Weight:   1,
		Capacity: node.capacity,
	}

	if err := dc.ring.AddNode(nodeInfo); err != nil {
		delete(dc.nodes, node.ID)
		return err
	}

	dc.logger.WithFields(map[string]interface{}{
		"node_id":  node.ID,
		"address":  node.Address,
		"capacity": node.capacity,
	}).Info("Added node to distributed cache")

	// Trigger key redistribution
	go dc.redistributeKeys(node.ID)

	return nil
}

// RemoveNode removes a node from the distributed cache
func (dc *DistributedCache) RemoveNode(nodeID string) error {
	dc.mu.Lock()
	node, exists := dc.nodes[nodeID]
	if !exists {
		dc.mu.Unlock()
		return errors.NotFoundf("node not found: %s", nodeID)
	}

	node.healthy.Store(false)
	delete(dc.nodes, nodeID)
	dc.mu.Unlock()

	if err := dc.ring.RemoveNode(nodeID); err != nil {
		return err
	}

	dc.logger.WithFields(map[string]interface{}{
		"node_id": nodeID,
	}).Info("Removed node from distributed cache")

	// Trigger key redistribution
	go dc.redistributeKeys("")

	return nil
}

// Get retrieves a value from the cache
func (dc *DistributedCache) Get(ctx context.Context, key string) ([]byte, error) {
	dc.metrics.Gets.Add(1)

	// Find primary node
	nodeID, err := dc.ring.GetNode(key)
	if err != nil {
		return nil, err
	}

	dc.mu.RLock()
	node, exists := dc.nodes[nodeID]
	dc.mu.RUnlock()

	if !exists || !node.healthy.Load() {
		// Try replicas
		return dc.getFromReplicas(ctx, key)
	}

	// Try primary node first
	value, err := node.Get(ctx, key)
	if err == nil {
		dc.metrics.Hits.Add(1)
		dc.logger.WithFields(map[string]interface{}{
			"key":     key,
			"node_id": nodeID,
			"source":  "primary",
		}).Debug("Cache hit on primary node")
		return value, nil
	}

	dc.metrics.Misses.Add(1)

	// Try replicas on miss or error
	return dc.getFromReplicas(ctx, key)
}

// getFromReplicas tries to get value from replica nodes
func (dc *DistributedCache) getFromReplicas(ctx context.Context, key string) ([]byte, error) {
	replicaNodes := dc.ring.GetReplicaNodes(key, dc.replication)

	dc.mu.RLock()
	defer dc.mu.RUnlock()

	for _, nodeID := range replicaNodes {
		node, exists := dc.nodes[nodeID]
		if !exists || !node.healthy.Load() {
			continue
		}

		value, err := node.Get(ctx, key)
		if err == nil {
			dc.metrics.Hits.Add(1)
			dc.logger.WithFields(map[string]interface{}{
				"key":     key,
				"node_id": nodeID,
				"source":  "replica",
			}).Debug("Cache hit on replica node")
			return value, nil
		}
	}

	return nil, errors.NotFoundf("key not found in cache: %s", key)
}

// Set stores a value in the cache
func (dc *DistributedCache) Set(ctx context.Context, key string, value []byte, ttl int64) error {
	dc.metrics.Sets.Add(1)

	// Find primary node
	nodeID, err := dc.ring.GetNode(key)
	if err != nil {
		return err
	}

	dc.mu.RLock()
	node, exists := dc.nodes[nodeID]
	dc.mu.RUnlock()

	if !exists || !node.healthy.Load() {
		return errors.NotFoundf("primary node not available: %s", nodeID)
	}

	// Store on primary node
	if err := node.Set(ctx, key, value, ttl); err != nil {
		return err
	}

	// Replicate to replica nodes
	go dc.replicateToNodes(ctx, key, value, ttl)

	dc.logger.WithFields(map[string]interface{}{
		"key":     key,
		"size":    len(value),
		"node_id": nodeID,
	}).Debug("Value stored in cache")

	return nil
}

// replicateToNodes replicates value to replica nodes
func (dc *DistributedCache) replicateToNodes(ctx context.Context, key string, value []byte, ttl int64) {
	replicaNodes := dc.ring.GetReplicaNodes(key, dc.replication)

	dc.mu.RLock()
	defer dc.mu.RUnlock()

	for _, nodeID := range replicaNodes {
		node, exists := dc.nodes[nodeID]
		if !exists || !node.healthy.Load() {
			continue
		}

		if err := node.Set(ctx, key, value, ttl); err != nil {
			dc.logger.WithFields(map[string]interface{}{
				"key":     key,
				"node_id": nodeID,
				"error":   err.Error(),
			}).Warn("Failed to replicate to node")
		}
	}
}

// Delete removes a value from the cache
func (dc *DistributedCache) Delete(ctx context.Context, key string) error {
	dc.metrics.Deletes.Add(1)

	// Find all nodes that might have this key (primary + replicas)
	nodeID, _ := dc.ring.GetNode(key)
	replicaNodes := dc.ring.GetReplicaNodes(key, dc.replication)

	allNodes := append([]string{nodeID}, replicaNodes...)

	dc.mu.RLock()
	defer dc.mu.RUnlock()

	var lastErr error
	for _, nid := range allNodes {
		node, exists := dc.nodes[nid]
		if !exists {
			continue
		}

		if err := node.Delete(ctx, key); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// GetMetrics returns cache metrics
func (dc *DistributedCache) GetMetrics() *CacheMetrics {
	return dc.metrics
}

// GetStats returns cache statistics
func (dc *DistributedCache) GetStats() map[string]interface{} {
	dc.mu.RLock()
	nodeCount := len(dc.nodes)
	dc.mu.RUnlock()

	hits := dc.metrics.Hits.Load()
	total := dc.metrics.Gets.Load()
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	return map[string]interface{}{
		"nodes":         nodeCount,
		"gets":          total,
		"sets":          dc.metrics.Sets.Load(),
		"hits":          hits,
		"misses":        dc.metrics.Misses.Load(),
		"hit_rate":      hitRate,
		"evictions":     dc.metrics.Evictions.Load(),
		"relocations":   dc.metrics.Relocations.Load(),
		"replication":   dc.replication,
		"virtual_nodes": dc.ring.vnodes,
	}
}

// redistributeKeys redistributes keys when nodes change
func (dc *DistributedCache) redistributeKeys(newNodeID string) {
	dc.logger.WithFields(map[string]interface{}{
		"new_node": newNodeID,
	}).Info("Starting key redistribution")

	relocated := 0

	// This is a placeholder - in production, you'd iterate through all keys
	// and move them to the new correct nodes based on the hash ring

	dc.metrics.Relocations.Add(uint64(relocated))

	dc.logger.WithFields(map[string]interface{}{
		"relocated": relocated,
	}).Info("Key redistribution completed")
}

// ConsistentHashRing methods

// AddNode adds a node to the hash ring
func (hr *ConsistentHashRing) AddNode(node *NodeInfo) error {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if _, exists := hr.nodes[node.ID]; exists {
		return errors.AlreadyExistsf("node already in ring: %s", node.ID)
	}

	hr.nodes[node.ID] = node

	// Add virtual nodes
	for i := 0; i < hr.vnodes; i++ {
		hash := hr.hashKey(node.ID, i)
		hr.circle[hash] = node.ID
		hr.sortedKeys = append(hr.sortedKeys, hash)
	}

	// Sort keys
	sort.Slice(hr.sortedKeys, func(i, j int) bool {
		return hr.sortedKeys[i] < hr.sortedKeys[j]
	})

	return nil
}

// RemoveNode removes a node from the hash ring
func (hr *ConsistentHashRing) RemoveNode(nodeID string) error {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if _, exists := hr.nodes[nodeID]; !exists {
		return errors.NotFoundf("node not in ring: %s", nodeID)
	}

	delete(hr.nodes, nodeID)

	// Remove virtual nodes
	newKeys := make([]uint32, 0, len(hr.sortedKeys))
	for _, key := range hr.sortedKeys {
		if hr.circle[key] != nodeID {
			newKeys = append(newKeys, key)
		} else {
			delete(hr.circle, key)
		}
	}

	hr.sortedKeys = newKeys

	return nil
}

// GetNode returns the node responsible for a key
func (hr *ConsistentHashRing) GetNode(key string) (string, error) {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if len(hr.sortedKeys) == 0 {
		return "", errors.New("no nodes in hash ring")
	}

	hash := hr.hashKey(key, 0)

	// Binary search for the first node >= hash
	idx := sort.Search(len(hr.sortedKeys), func(i int) bool {
		return hr.sortedKeys[i] >= hash
	})

	// Wrap around if necessary
	if idx >= len(hr.sortedKeys) {
		idx = 0
	}

	return hr.circle[hr.sortedKeys[idx]], nil
}

// GetReplicaNodes returns replica nodes for a key
func (hr *ConsistentHashRing) GetReplicaNodes(key string, count int) []string {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if len(hr.sortedKeys) == 0 {
		return nil
	}

	hash := hr.hashKey(key, 0)
	idx := sort.Search(len(hr.sortedKeys), func(i int) bool {
		return hr.sortedKeys[i] >= hash
	})

	replicas := make([]string, 0, count)
	seen := make(map[string]bool)

	for len(replicas) < count && len(replicas) < len(hr.nodes) {
		if idx >= len(hr.sortedKeys) {
			idx = 0
		}

		nodeID := hr.circle[hr.sortedKeys[idx]]
		if !seen[nodeID] {
			replicas = append(replicas, nodeID)
			seen[nodeID] = true
		}

		idx++
	}

	return replicas
}

// hashKey generates a hash for a key
func (hr *ConsistentHashRing) hashKey(key string, vnode int) uint32 {
	data := []byte(key)
	if vnode > 0 {
		data = append(data, byte(vnode))
	}
	return crc32.ChecksumIEEE(data)
}

// CacheNode methods

// NewCacheNode creates a new cache node
func NewCacheNode(id, address string, capacity int64, client CacheClient) *CacheNode {
	node := &CacheNode{
		ID:       id,
		Address:  address,
		cache:    make(map[string]*CacheEntry),
		capacity: capacity,
		client:   client,
	}
	node.healthy.Store(true)
	return node
}

// Get retrieves a value from the node
func (cn *CacheNode) Get(ctx context.Context, key string) ([]byte, error) {
	// Check local cache first
	cn.mu.RLock()
	entry, exists := cn.cache[key]
	cn.mu.RUnlock()

	if exists {
		entry.HitCount.Add(1)
		entry.AccessTime.Store(time.Now().Unix())
		return entry.Value, nil
	}

	// Fetch from remote node if client is available
	if cn.client != nil {
		return cn.client.Get(ctx, key)
	}

	return nil, errors.NotFoundf("key not found: %s", key)
}

// Set stores a value in the node
func (cn *CacheNode) Set(ctx context.Context, key string, value []byte, ttl int64) error {
	now := time.Now().Unix()
	entry := &CacheEntry{
		Key:       key,
		Value:     value,
		Size:      int64(len(value)),
		CreatedAt: now,
		ExpiresAt: now + ttl,
	}
	entry.AccessTime.Store(now)

	cn.mu.Lock()
	cn.cache[key] = entry
	cn.size.Add(entry.Size)
	cn.mu.Unlock()

	// Store to remote node if client is available
	if cn.client != nil {
		return cn.client.Set(ctx, key, value, ttl)
	}

	return nil
}

// Delete removes a value from the node
func (cn *CacheNode) Delete(ctx context.Context, key string) error {
	cn.mu.Lock()
	if entry, exists := cn.cache[key]; exists {
		delete(cn.cache, key)
		cn.size.Add(-entry.Size)
	}
	cn.mu.Unlock()

	// Delete from remote node if client is available
	if cn.client != nil {
		return cn.client.Delete(ctx, key)
	}

	return nil
}

// GetSize returns the current size of the cache
func (cn *CacheNode) GetSize() int64 {
	return cn.size.Load()
}

// IsHealthy returns the health status
func (cn *CacheNode) IsHealthy() bool {
	return cn.healthy.Load()
}
