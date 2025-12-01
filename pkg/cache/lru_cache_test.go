package cache

import (
	"testing"
)

// TestNewLRUCache tests LRU cache creation
func TestNewLRUCache(t *testing.T) {
	cache := NewLRUCache[string, string](10)

	if cache == nil {
		t.Fatal("Expected cache to be created")
	}

	if cache.capacity != 10 {
		t.Errorf("Expected capacity 10, got %d", cache.capacity)
	}

	if cache.Size() != 0 {
		t.Errorf("Expected initial size 0, got %d", cache.Size())
	}
}

// TestNewLRUCacheZeroCapacity tests cache with zero capacity
func TestNewLRUCacheZeroCapacity(t *testing.T) {
	cache := NewLRUCache[string, string](0)

	if cache.capacity != 1 {
		t.Errorf("Expected minimum capacity 1, got %d", cache.capacity)
	}
}

// TestNewLRUCacheNegativeCapacity tests cache with negative capacity
func TestNewLRUCacheNegativeCapacity(t *testing.T) {
	cache := NewLRUCache[string, string](-5)

	if cache.capacity != 1 {
		t.Errorf("Expected minimum capacity 1, got %d", cache.capacity)
	}
}

// TestPutAndGet tests basic put and get operations
func TestPutAndGet(t *testing.T) {
	cache := NewLRUCache[string, int](10)

	// Put values
	cache.Put("key1", 100)
	cache.Put("key2", 200)
	cache.Put("key3", 300)

	// Get values
	val, found := cache.Get("key1")
	if !found {
		t.Error("Expected key1 to be found")
	}
	if val != 100 {
		t.Errorf("Expected value 100, got %d", val)
	}

	val, found = cache.Get("key2")
	if !found {
		t.Error("Expected key2 to be found")
	}
	if val != 200 {
		t.Errorf("Expected value 200, got %d", val)
	}

	// Check size
	if cache.Size() != 3 {
		t.Errorf("Expected size 3, got %d", cache.Size())
	}
}

// TestGetNonExistent tests getting non-existent key
func TestGetNonExistent(t *testing.T) {
	cache := NewLRUCache[string, int](10)

	_, found := cache.Get("nonexistent")
	if found {
		t.Error("Expected key not to be found")
	}
}

// TestPutUpdate tests updating existing key
func TestPutUpdate(t *testing.T) {
	cache := NewLRUCache[string, int](10)

	cache.Put("key1", 100)
	cache.Put("key1", 200)

	val, found := cache.Get("key1")
	if !found {
		t.Error("Expected key1 to be found")
	}
	if val != 200 {
		t.Errorf("Expected updated value 200, got %d", val)
	}

	// Size should still be 1
	if cache.Size() != 1 {
		t.Errorf("Expected size 1, got %d", cache.Size())
	}
}

// TestLRUEviction tests LRU eviction policy
func TestLRUEviction(t *testing.T) {
	cache := NewLRUCache[string, int](3)

	// Fill cache to capacity
	cache.Put("key1", 1)
	cache.Put("key2", 2)
	cache.Put("key3", 3)

	// Add one more (should evict key1 as least recently used)
	cache.Put("key4", 4)

	// key1 should be evicted
	_, found := cache.Get("key1")
	if found {
		t.Error("Expected key1 to be evicted")
	}

	// Others should still exist
	if _, found := cache.Get("key2"); !found {
		t.Error("Expected key2 to exist")
	}
	if _, found := cache.Get("key3"); !found {
		t.Error("Expected key3 to exist")
	}
	if _, found := cache.Get("key4"); !found {
		t.Error("Expected key4 to exist")
	}

	// Size should be at capacity
	if cache.Size() != 3 {
		t.Errorf("Expected size 3, got %d", cache.Size())
	}
}

// TestLRUEvictionWithAccess tests that accessing items prevents eviction
func TestLRUEvictionWithAccess(t *testing.T) {
	cache := NewLRUCache[string, int](3)

	cache.Put("key1", 1)
	cache.Put("key2", 2)
	cache.Put("key3", 3)

	// Access key1 to make it most recently used
	cache.Get("key1")

	// Add key4 (should evict key2, not key1)
	cache.Put("key4", 4)

	// key1 should still exist
	if _, found := cache.Get("key1"); !found {
		t.Error("Expected key1 to exist after access")
	}

	// key2 should be evicted
	if _, found := cache.Get("key2"); found {
		t.Error("Expected key2 to be evicted")
	}
}

// TestRemove tests removing items
func TestRemove(t *testing.T) {
	cache := NewLRUCache[string, int](10)

	cache.Put("key1", 100)
	cache.Put("key2", 200)

	// Remove existing key
	removed := cache.Remove("key1")
	if !removed {
		t.Error("Expected Remove to return true")
	}

	// Try to get removed key
	_, found := cache.Get("key1")
	if found {
		t.Error("Expected key1 to be removed")
	}

	// Size should be 1
	if cache.Size() != 1 {
		t.Errorf("Expected size 1, got %d", cache.Size())
	}

	// Remove non-existent key
	removed = cache.Remove("nonexistent")
	if removed {
		t.Error("Expected Remove to return false for non-existent key")
	}
}

// TestClear tests clearing the cache
func TestClear(t *testing.T) {
	cache := NewLRUCache[string, int](10)

	cache.Put("key1", 1)
	cache.Put("key2", 2)
	cache.Put("key3", 3)

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}

	// All keys should be gone
	_, found := cache.Get("key1")
	if found {
		t.Error("Expected all keys to be cleared")
	}
}

// TestContains tests key existence check
func TestContains(t *testing.T) {
	cache := NewLRUCache[string, int](10)

	cache.Put("key1", 100)

	if !cache.Contains("key1") {
		t.Error("Expected Contains to return true for existing key")
	}

	if cache.Contains("nonexistent") {
		t.Error("Expected Contains to return false for non-existent key")
	}

	// Contains should not update LRU order
	cache.Put("key2", 200)
	cache.Put("key3", 300)
	cache.Contains("key1") // Check without affecting order

	// Access should move to front, but Contains shouldn't
	oldest, _, _ := cache.GetOldest()
	if oldest != "key1" {
		t.Error("Expected Contains not to affect LRU order")
	}
}

// TestKeys tests retrieving all keys
func TestKeys(t *testing.T) {
	cache := NewLRUCache[string, int](10)

	cache.Put("key1", 1)
	cache.Put("key2", 2)
	cache.Put("key3", 3)

	keys := cache.Keys()

	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Keys should be in MRU order (most recent first)
	if keys[0] != "key3" {
		t.Errorf("Expected most recent key to be key3, got %s", keys[0])
	}
}

// TestGetOldest tests getting least recently used item
func TestGetOldest(t *testing.T) {
	cache := NewLRUCache[string, int](10)

	// Empty cache
	_, _, found := cache.GetOldest()
	if found {
		t.Error("Expected GetOldest to return false for empty cache")
	}

	cache.Put("key1", 1)
	cache.Put("key2", 2)
	cache.Put("key3", 3)

	key, val, found := cache.GetOldest()
	if !found {
		t.Error("Expected GetOldest to return true")
	}
	if key != "key1" {
		t.Errorf("Expected oldest key to be key1, got %s", key)
	}
	if val != 1 {
		t.Errorf("Expected oldest value to be 1, got %d", val)
	}
}

// TestGetNewest tests getting most recently used item
func TestGetNewest(t *testing.T) {
	cache := NewLRUCache[string, int](10)

	// Empty cache
	_, _, found := cache.GetNewest()
	if found {
		t.Error("Expected GetNewest to return false for empty cache")
	}

	cache.Put("key1", 1)
	cache.Put("key2", 2)
	cache.Put("key3", 3)

	key, val, found := cache.GetNewest()
	if !found {
		t.Error("Expected GetNewest to return true")
	}
	if key != "key3" {
		t.Errorf("Expected newest key to be key3, got %s", key)
	}
	if val != 3 {
		t.Errorf("Expected newest value to be 3, got %d", val)
	}
}

// TestIterateAll tests iterating over all items
func TestIterateAll(t *testing.T) {
	cache := NewLRUCache[string, int](10)

	cache.Put("key1", 1)
	cache.Put("key2", 2)
	cache.Put("key3", 3)

	count := 0
	keys := []string{}

	cache.IterateAll(func(key string, value int) bool {
		count++
		keys = append(keys, key)
		return true
	})

	if count != 3 {
		t.Errorf("Expected to iterate over 3 items, got %d", count)
	}

	// Should iterate in MRU order
	if keys[0] != "key3" {
		t.Errorf("Expected first key to be key3, got %s", keys[0])
	}
}

// TestIterateAllEarlyExit tests early exit from iteration
func TestIterateAllEarlyExit(t *testing.T) {
	cache := NewLRUCache[string, int](10)

	cache.Put("key1", 1)
	cache.Put("key2", 2)
	cache.Put("key3", 3)

	count := 0
	cache.IterateAll(func(key string, value int) bool {
		count++
		return count < 2 // Stop after 2 iterations
	})

	if count != 2 {
		t.Errorf("Expected to iterate 2 times, got %d", count)
	}
}

// TestIterateOldest tests iterating from oldest to newest
func TestIterateOldest(t *testing.T) {
	cache := NewLRUCache[string, int](10)

	cache.Put("key1", 1)
	cache.Put("key2", 2)
	cache.Put("key3", 3)

	keys := []string{}
	cache.IterateOldest(func(key string, value int) bool {
		keys = append(keys, key)
		return true
	})

	// Should iterate from oldest to newest
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}
	if keys[0] != "key1" {
		t.Errorf("Expected first key to be key1, got %s", keys[0])
	}
	if keys[2] != "key3" {
		t.Errorf("Expected last key to be key3, got %s", keys[2])
	}
}

// TestIterateOldestEarlyExit tests early exit from oldest iteration
func TestIterateOldestEarlyExit(t *testing.T) {
	cache := NewLRUCache[string, int](10)

	cache.Put("key1", 1)
	cache.Put("key2", 2)
	cache.Put("key3", 3)

	count := 0
	cache.IterateOldest(func(key string, value int) bool {
		count++
		return count < 2
	})

	if count != 2 {
		t.Errorf("Expected to iterate 2 times, got %d", count)
	}
}

// TestConcurrentAccess tests basic thread safety
func TestConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	cache := NewLRUCache[int, int](100)

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cache.Put(i, i*2)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cache.Get(i)
		}
		done <- true
	}()

	// Wait for both to complete
	<-done
	<-done

	// Cache should be consistent
	if cache.Size() > 100 {
		t.Errorf("Expected size <= 100, got %d", cache.Size())
	}
}

// TestDifferentTypes tests LRU cache with different types
func TestDifferentTypes(t *testing.T) {
	// String to struct
	type TestStruct struct {
		Name  string
		Value int
	}

	cache := NewLRUCache[string, TestStruct](10)
	cache.Put("test", TestStruct{Name: "Test", Value: 42})

	val, found := cache.Get("test")
	if !found {
		t.Error("Expected to find test key")
	}
	if val.Name != "Test" || val.Value != 42 {
		t.Error("Expected struct values to match")
	}

	// Int to string
	intCache := NewLRUCache[int, string](10)
	intCache.Put(1, "one")
	intCache.Put(2, "two")

	str, found := intCache.Get(1)
	if !found || str != "one" {
		t.Error("Expected to find int key with string value")
	}
}

// TestLRUOrderingAfterUpdate tests that updating moves item to front
func TestLRUOrderingAfterUpdate(t *testing.T) {
	cache := NewLRUCache[string, int](3)

	cache.Put("key1", 1)
	cache.Put("key2", 2)
	cache.Put("key3", 3)

	// Update key1 (should move to front)
	cache.Put("key1", 10)

	// Add key4 (should evict key2, not key1)
	cache.Put("key4", 4)

	// key1 should still exist with updated value
	val, found := cache.Get("key1")
	if !found {
		t.Error("Expected key1 to exist after update")
	}
	if val != 10 {
		t.Errorf("Expected updated value 10, got %d", val)
	}

	// key2 should be evicted
	_, found = cache.Get("key2")
	if found {
		t.Error("Expected key2 to be evicted")
	}
}
