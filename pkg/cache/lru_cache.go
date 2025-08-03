package cache

import (
	"sync"
)

// LRUCache provides a thread-safe LRU (Least Recently Used) cache implementation
type LRUCache[K comparable, V any] struct {
	capacity int
	items    map[K]*lruNode[K, V]
	head     *lruNode[K, V]
	tail     *lruNode[K, V]
	mutex    sync.RWMutex
}

// lruNode represents a node in the LRU cache's doubly-linked list
type lruNode[K comparable, V any] struct {
	key   K
	value V
	prev  *lruNode[K, V]
	next  *lruNode[K, V]
}

// NewLRUCache creates a new LRU cache with the specified capacity
func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	if capacity <= 0 {
		capacity = 1
	}

	cache := &LRUCache[K, V]{
		capacity: capacity,
		items:    make(map[K]*lruNode[K, V]),
	}

	// Initialize sentinel nodes
	cache.head = &lruNode[K, V]{}
	cache.tail = &lruNode[K, V]{}
	cache.head.next = cache.tail
	cache.tail.prev = cache.head

	return cache
}

// Get retrieves a value from the cache and marks it as recently used
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if node, exists := c.items[key]; exists {
		// Move to front (most recently used)
		c.moveToFront(node)
		return node.value, true
	}

	var zero V
	return zero, false
}

// Put adds or updates a value in the cache
func (c *LRUCache[K, V]) Put(key K, value V) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if node, exists := c.items[key]; exists {
		// Update existing node
		node.value = value
		c.moveToFront(node)
		return
	}

	// Create new node
	newNode := &lruNode[K, V]{
		key:   key,
		value: value,
	}

	// Add to front
	c.addToFront(newNode)
	c.items[key] = newNode

	// Check capacity and evict if necessary
	if len(c.items) > c.capacity {
		c.evictLRU()
	}
}

// Remove removes a key from the cache
func (c *LRUCache[K, V]) Remove(key K) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if node, exists := c.items[key]; exists {
		c.removeNode(node)
		delete(c.items, key)
		return true
	}

	return false
}

// Size returns the current number of items in the cache
func (c *LRUCache[K, V]) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.items)
}

// Clear removes all items from the cache
func (c *LRUCache[K, V]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[K]*lruNode[K, V])
	c.head.next = c.tail
	c.tail.prev = c.head
}

// IterateAll iterates over all items in the cache (from most to least recently used)
func (c *LRUCache[K, V]) IterateAll(fn func(key K, value V) bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	current := c.head.next
	for current != c.tail {
		if !fn(current.key, current.value) {
			break
		}
		current = current.next
	}
}

// IterateOldest iterates over items starting from the least recently used
func (c *LRUCache[K, V]) IterateOldest(fn func(key K, value V) bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	current := c.tail.prev
	for current != c.head {
		if !fn(current.key, current.value) {
			break
		}
		current = current.prev
	}
}

// GetOldest returns the least recently used item without removing it
func (c *LRUCache[K, V]) GetOldest() (K, V, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.tail.prev != c.head {
		node := c.tail.prev
		return node.key, node.value, true
	}

	var zeroK K
	var zeroV V
	return zeroK, zeroV, false
}

// GetNewest returns the most recently used item without removing it
func (c *LRUCache[K, V]) GetNewest() (K, V, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.head.next != c.tail {
		node := c.head.next
		return node.key, node.value, true
	}

	var zeroK K
	var zeroV V
	return zeroK, zeroV, false
}

// Keys returns all keys in the cache (from most to least recently used)
func (c *LRUCache[K, V]) Keys() []K {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make([]K, 0, len(c.items))
	current := c.head.next
	for current != c.tail {
		keys = append(keys, current.key)
		current = current.next
	}

	return keys
}

// Contains checks if a key exists in the cache without updating its position
func (c *LRUCache[K, V]) Contains(key K) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	_, exists := c.items[key]
	return exists
}

// Internal methods

// moveToFront moves a node to the front of the list (most recently used)
func (c *LRUCache[K, V]) moveToFront(node *lruNode[K, V]) {
	c.removeNode(node)
	c.addToFront(node)
}

// addToFront adds a node to the front of the list
func (c *LRUCache[K, V]) addToFront(node *lruNode[K, V]) {
	node.prev = c.head
	node.next = c.head.next
	c.head.next.prev = node
	c.head.next = node
}

// removeNode removes a node from the list
func (c *LRUCache[K, V]) removeNode(node *lruNode[K, V]) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

// evictLRU removes the least recently used item
func (c *LRUCache[K, V]) evictLRU() {
	if c.tail.prev != c.head {
		lru := c.tail.prev
		c.removeNode(lru)
		delete(c.items, lru.key)
	}
}
