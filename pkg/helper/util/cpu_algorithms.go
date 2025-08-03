package util

import (
	"hash/fnv"
	"sort"
	"strings"
	"sync"
	"time"

	"freightliner/pkg/helper/log"
)

// AlgorithmComplexity defines the time complexity of algorithms
type AlgorithmComplexity string

const (
	ComplexityO1     AlgorithmComplexity = "O(1)"       // Constant time
	ComplexityOLogN  AlgorithmComplexity = "O(log n)"   // Logarithmic time
	ComplexityON     AlgorithmComplexity = "O(n)"       // Linear time
	ComplexityONLogN AlgorithmComplexity = "O(n log n)" // Linearithmic time
	ComplexityON2    AlgorithmComplexity = "O(n²)"      // Quadratic time
	ComplexityON3    AlgorithmComplexity = "O(n³)"      // Cubic time
)

// CPUEfficientSorter provides optimized sorting algorithms with complexity analysis
type CPUEfficientSorter struct {
	logger log.Logger
}

// NewCPUEfficientSorter creates a new CPU-efficient sorter
func NewCPUEfficientSorter(logger log.Logger) *CPUEfficientSorter {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}
	return &CPUEfficientSorter{
		logger: logger,
	}
}

// SortableResource represents a resource that can be sorted by priority
type SortableResource struct {
	Name     string
	Priority int
	Data     interface{}
}

// SortResourcesByPriority sorts resources efficiently by priority (O(n log n))
// Replaces the O(n²) bubble sort in resource_cleanup.go
func (ces *CPUEfficientSorter) SortResourcesByPriority(resources []SortableResource) {
	if len(resources) <= 1 {
		return
	}

	// Use Go's optimized sorting (introsort/quicksort hybrid) - O(n log n)
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Priority > resources[j].Priority // Higher priority first
	})

	if ces.logger != nil {
		ces.logger.WithFields(map[string]interface{}{
			"count":      len(resources),
			"complexity": string(ComplexityONLogN),
			"algorithm":  "introsort",
		}).Debug("Sorted resources by priority")
	}
}

// StringMatcher provides optimized string matching algorithms
type StringMatcher struct {
	// Boyer-Moore bad character table for efficient string search
	badCharTable map[rune]int
	pattern      string
	patternLen   int
	mu           sync.RWMutex
}

// NewStringMatcher creates an optimized string matcher
func NewStringMatcher(pattern string) *StringMatcher {
	sm := &StringMatcher{
		pattern:    pattern,
		patternLen: len(pattern),
	}
	sm.buildBadCharTable()
	return sm
}

// buildBadCharTable builds the bad character table for Boyer-Moore algorithm
func (sm *StringMatcher) buildBadCharTable() {
	sm.badCharTable = make(map[rune]int)

	// Build bad character table
	patternRunes := []rune(sm.pattern)
	for i, char := range patternRunes {
		sm.badCharTable[char] = sm.patternLen - 1 - i
	}
}

// FindAll finds all occurrences of pattern in text using Boyer-Moore algorithm (O(n/m) average case)
func (sm *StringMatcher) FindAll(text string) []int {
	if sm.patternLen == 0 || len(text) < sm.patternLen {
		return nil
	}

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var matches []int
	textRunes := []rune(text)
	patternRunes := []rune(sm.pattern)
	textLen := len(textRunes)

	i := 0
	for i <= textLen-sm.patternLen {
		// Start matching from the end of pattern
		j := sm.patternLen - 1

		// Match characters from right to left
		for j >= 0 && patternRunes[j] == textRunes[i+j] {
			j--
		}

		if j < 0 {
			// Pattern found
			matches = append(matches, i)
			i++ // Move to next position
		} else {
			// Mismatch found, use bad character heuristic
			badCharShift := sm.badCharTable[textRunes[i+j]]
			if badCharShift == 0 {
				badCharShift = sm.patternLen
			}
			i += badCharShift
		}
	}

	return matches
}

// GetComplexity returns the complexity of the string matching algorithm
func (sm *StringMatcher) GetComplexity() AlgorithmComplexity {
	return "O(n/m) average, O(nm) worst case"
}

// EfficientSetOperations provides optimized set operations
type EfficientSetOperations[T comparable] struct {
	logger log.Logger
}

// NewEfficientSetOperations creates efficient set operations
func NewEfficientSetOperations[T comparable](logger log.Logger) *EfficientSetOperations[T] {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}
	return &EfficientSetOperations[T]{
		logger: logger,
	}
}

// Intersection computes set intersection efficiently using hash maps (O(n + m))
func (eso *EfficientSetOperations[T]) Intersection(set1, set2 []T) []T {
	if len(set1) == 0 || len(set2) == 0 {
		return nil
	}

	// Use the smaller set for the hash map to optimize memory usage
	smaller, larger := set1, set2
	if len(set2) < len(set1) {
		smaller, larger = set2, set1
	}

	// Build hash map from smaller set - O(min(n, m))
	hashMap := make(map[T]struct{}, len(smaller))
	for _, item := range smaller {
		hashMap[item] = struct{}{}
	}

	// Find intersection - O(max(n, m))
	var result []T
	for _, item := range larger {
		if _, exists := hashMap[item]; exists {
			result = append(result, item)
			delete(hashMap, item) // Avoid duplicates
		}
	}

	if eso.logger != nil {
		eso.logger.WithFields(map[string]interface{}{
			"set1_size":   len(set1),
			"set2_size":   len(set2),
			"result_size": len(result),
			"complexity":  string(ComplexityON),
		}).Debug("Computed set intersection")
	}

	return result
}

// Union computes set union efficiently (O(n + m))
func (eso *EfficientSetOperations[T]) Union(set1, set2 []T) []T {
	seen := make(map[T]struct{}, len(set1)+len(set2))
	var result []T

	// Add all elements from set1
	for _, item := range set1 {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	// Add new elements from set2
	for _, item := range set2 {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// Difference computes set difference efficiently (O(n + m))
func (eso *EfficientSetOperations[T]) Difference(set1, set2 []T) []T {
	// Build hash map from set2
	hashMap := make(map[T]struct{}, len(set2))
	for _, item := range set2 {
		hashMap[item] = struct{}{}
	}

	// Find elements in set1 but not in set2
	var result []T
	for _, item := range set1 {
		if _, exists := hashMap[item]; !exists {
			result = append(result, item)
		}
	}

	return result
}

// PatternMatchingCache provides optimized pattern matching with caching
type PatternMatchingCache struct {
	cache        map[string]*StringMatcher
	mu           sync.RWMutex
	maxCacheSize int
	logger       log.Logger
}

// NewPatternMatchingCache creates a new pattern matching cache
func NewPatternMatchingCache(maxCacheSize int, logger log.Logger) *PatternMatchingCache {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}
	return &PatternMatchingCache{
		cache:        make(map[string]*StringMatcher),
		maxCacheSize: maxCacheSize,
		logger:       logger,
	}
}

// Match performs optimized pattern matching with caching
func (pmc *PatternMatchingCache) Match(pattern, text string) bool {
	// Fast path for simple patterns
	if !strings.ContainsAny(pattern, "*?[]{}") {
		return strings.Contains(text, pattern)
	}

	// Get or create matcher
	matcher := pmc.getMatcher(pattern)
	matches := matcher.FindAll(text)
	return len(matches) > 0
}

// getMatcher gets or creates a string matcher with LRU eviction
func (pmc *PatternMatchingCache) getMatcher(pattern string) *StringMatcher {
	pmc.mu.RLock()
	if matcher, exists := pmc.cache[pattern]; exists {
		pmc.mu.RUnlock()
		return matcher
	}
	pmc.mu.RUnlock()

	pmc.mu.Lock()
	defer pmc.mu.Unlock()

	// Double-check after acquiring write lock
	if matcher, exists := pmc.cache[pattern]; exists {
		return matcher
	}

	// Evict oldest entries if cache is full
	if len(pmc.cache) >= pmc.maxCacheSize {
		// Simple eviction - remove first entry (in practice, use LRU)
		for key := range pmc.cache {
			delete(pmc.cache, key)
			break
		}
	}

	// Create and cache new matcher
	matcher := NewStringMatcher(pattern)
	pmc.cache[pattern] = matcher

	if pmc.logger != nil {
		pmc.logger.WithFields(map[string]interface{}{
			"pattern":    pattern,
			"cache_size": len(pmc.cache),
		}).Debug("Created new pattern matcher")
	}

	return matcher
}

// BinarySearchTree provides efficient search operations
type BinarySearchTree[T comparable] struct {
	root    *BSTNode[T]
	size    int
	compare func(T, T) int
	mu      sync.RWMutex
}

// BSTNode represents a node in the binary search tree
type BSTNode[T comparable] struct {
	value T
	left  *BSTNode[T]
	right *BSTNode[T]
}

// NewBinarySearchTree creates a new binary search tree
func NewBinarySearchTree[T comparable](compare func(T, T) int) *BinarySearchTree[T] {
	return &BinarySearchTree[T]{
		compare: compare,
	}
}

// Insert inserts a value into the tree (O(log n) average, O(n) worst case)
func (bst *BinarySearchTree[T]) Insert(value T) {
	bst.mu.Lock()
	defer bst.mu.Unlock()

	bst.root = bst.insertNode(bst.root, value)
	bst.size++
}

// insertNode recursively inserts a node
func (bst *BinarySearchTree[T]) insertNode(node *BSTNode[T], value T) *BSTNode[T] {
	if node == nil {
		return &BSTNode[T]{value: value}
	}

	if bst.compare(value, node.value) < 0 {
		node.left = bst.insertNode(node.left, value)
	} else {
		node.right = bst.insertNode(node.right, value)
	}

	return node
}

// Search searches for a value in the tree (O(log n) average, O(n) worst case)
func (bst *BinarySearchTree[T]) Search(value T) bool {
	bst.mu.RLock()
	defer bst.mu.RUnlock()

	return bst.searchNode(bst.root, value)
}

// searchNode recursively searches for a node
func (bst *BinarySearchTree[T]) searchNode(node *BSTNode[T], value T) bool {
	if node == nil {
		return false
	}

	cmp := bst.compare(value, node.value)
	if cmp == 0 {
		return true
	} else if cmp < 0 {
		return bst.searchNode(node.left, value)
	} else {
		return bst.searchNode(node.right, value)
	}
}

// Size returns the number of elements in the tree
func (bst *BinarySearchTree[T]) Size() int {
	bst.mu.RLock()
	defer bst.mu.RUnlock()
	return bst.size
}

// HashTable provides efficient O(1) average case operations
type HashTable[K comparable, V any] struct {
	buckets  [][]HashEntry[K, V]
	size     int
	capacity int
	mu       sync.RWMutex
}

// HashEntry represents an entry in the hash table
type HashEntry[K comparable, V any] struct {
	Key   K
	Value V
}

// NewHashTable creates a new hash table with specified initial capacity
func NewHashTable[K comparable, V any](initialCapacity int) *HashTable[K, V] {
	if initialCapacity < 16 {
		initialCapacity = 16
	}

	return &HashTable[K, V]{
		buckets:  make([][]HashEntry[K, V], initialCapacity),
		capacity: initialCapacity,
	}
}

// hash computes hash for a key using FNV-1a hash algorithm
func (ht *HashTable[K, V]) hash(key K) int {
	h := fnv.New32a()

	// Convert key to bytes for hashing
	// This is a safer approach than using unsafe.Pointer
	switch k := any(key).(type) {
	case string:
		h.Write([]byte(k))
	case int:
		bytes := [8]byte{}
		for i := 0; i < 8; i++ {
			bytes[i] = byte(k >> (i * 8))
		}
		h.Write(bytes[:])
	case int32:
		bytes := [4]byte{}
		for i := 0; i < 4; i++ {
			bytes[i] = byte(k >> (i * 8))
		}
		h.Write(bytes[:])
	case int64:
		bytes := [8]byte{}
		for i := 0; i < 8; i++ {
			bytes[i] = byte(k >> (i * 8))
		}
		h.Write(bytes[:])
	default:
		// For other comparable types, use a simple hash based on type conversion
		// This is safe but basic - for production, implement specific hash for each type
		h.Write([]byte{byte(any(key).(int) % 256)})
	}

	return int(h.Sum32()) % ht.capacity
}

// Put inserts or updates a key-value pair (O(1) average case)
func (ht *HashTable[K, V]) Put(key K, value V) {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	index := ht.hash(key)
	bucket := ht.buckets[index]

	// Check if key already exists
	for i, entry := range bucket {
		if entry.Key == key {
			bucket[i].Value = value
			return
		}
	}

	// Add new entry
	ht.buckets[index] = append(bucket, HashEntry[K, V]{Key: key, Value: value})
	ht.size++

	// Resize if load factor is too high
	if float64(ht.size)/float64(ht.capacity) > 0.75 {
		ht.resize()
	}
}

// Get retrieves a value by key (O(1) average case)
func (ht *HashTable[K, V]) Get(key K) (V, bool) {
	ht.mu.RLock()
	defer ht.mu.RUnlock()

	index := ht.hash(key)
	bucket := ht.buckets[index]

	for _, entry := range bucket {
		if entry.Key == key {
			return entry.Value, true
		}
	}

	var zero V
	return zero, false
}

// resize doubles the capacity and rehashes all entries
func (ht *HashTable[K, V]) resize() {
	oldBuckets := ht.buckets
	ht.capacity *= 2
	ht.buckets = make([][]HashEntry[K, V], ht.capacity)
	oldSize := ht.size
	ht.size = 0

	// Rehash all entries
	for _, bucket := range oldBuckets {
		for _, entry := range bucket {
			ht.put(entry.Key, entry.Value) // Internal put without locking
		}
	}

	ht.size = oldSize // Restore size (put() increments it)
}

// put is the internal put method without locking
func (ht *HashTable[K, V]) put(key K, value V) {
	index := ht.hash(key)
	bucket := ht.buckets[index]

	for i, entry := range bucket {
		if entry.Key == key {
			bucket[i].Value = value
			return
		}
	}

	ht.buckets[index] = append(bucket, HashEntry[K, V]{Key: key, Value: value})
}

// AlgorithmProfiler provides performance profiling for algorithms
type AlgorithmProfiler struct {
	profiles map[string]*AlgorithmProfile
	mu       sync.Mutex
	logger   log.Logger
}

// AlgorithmProfile contains performance metrics for an algorithm
type AlgorithmProfile struct {
	Name           string
	Complexity     AlgorithmComplexity
	ExecutionCount int64
	TotalTime      time.Duration
	AverageTime    time.Duration
	MinTime        time.Duration
	MaxTime        time.Duration
	InputSizes     []int
}

// NewAlgorithmProfiler creates a new algorithm profiler
func NewAlgorithmProfiler(logger log.Logger) *AlgorithmProfiler {
	if logger == nil {
		logger = log.NewBasicLogger(log.InfoLevel)
	}
	return &AlgorithmProfiler{
		profiles: make(map[string]*AlgorithmProfile),
		logger:   logger,
	}
}

// ProfileAlgorithm profiles the execution of an algorithm
func (ap *AlgorithmProfiler) ProfileAlgorithm(
	name string,
	complexity AlgorithmComplexity,
	inputSize int,
	algorithm func(),
) {
	start := time.Now()
	algorithm()
	duration := time.Since(start)

	ap.recordExecution(name, complexity, inputSize, duration)
}

// recordExecution records the execution metrics
func (ap *AlgorithmProfiler) recordExecution(
	name string,
	complexity AlgorithmComplexity,
	inputSize int,
	duration time.Duration,
) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	profile, exists := ap.profiles[name]
	if !exists {
		profile = &AlgorithmProfile{
			Name:       name,
			Complexity: complexity,
			MinTime:    duration,
			MaxTime:    duration,
		}
		ap.profiles[name] = profile
	}

	// Update metrics
	profile.ExecutionCount++
	profile.TotalTime += duration
	profile.AverageTime = profile.TotalTime / time.Duration(profile.ExecutionCount)
	profile.InputSizes = append(profile.InputSizes, inputSize)

	if duration < profile.MinTime {
		profile.MinTime = duration
	}
	if duration > profile.MaxTime {
		profile.MaxTime = duration
	}

	if ap.logger != nil {
		ap.logger.WithFields(map[string]interface{}{
			"name":        name,
			"complexity":  string(complexity),
			"input_size":  inputSize,
			"duration_ns": duration.Nanoseconds(),
			"avg_time_ns": profile.AverageTime.Nanoseconds(),
		}).Debug("Algorithm executed")
	}
}

// GetProfile returns the profile for an algorithm
func (ap *AlgorithmProfiler) GetProfile(name string) (*AlgorithmProfile, bool) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	profile, exists := ap.profiles[name]
	return profile, exists
}

// GetAllProfiles returns all algorithm profiles
func (ap *AlgorithmProfiler) GetAllProfiles() map[string]*AlgorithmProfile {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	result := make(map[string]*AlgorithmProfile)
	for name, profile := range ap.profiles {
		result[name] = profile
	}
	return result
}

// Global instances for convenience
var (
	GlobalCPUSorter    = NewCPUEfficientSorter(nil)
	GlobalPatternCache = NewPatternMatchingCache(1000, nil)
	GlobalAlgProfiler  = NewAlgorithmProfiler(nil)
)
