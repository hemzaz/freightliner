package testing

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"freightliner/pkg/helper/log"
)

// TestCacheSystem provides intelligent test result caching and parallelization
type TestCacheSystem struct {
	logger          log.Logger
	cacheDir        string
	maxCacheAge     time.Duration
	enabledPackages map[string]bool
	cacheStats      *CacheStatistics
	mu              sync.RWMutex
}

// CacheStatistics tracks cache performance metrics
type CacheStatistics struct {
	TotalRequests int64
	CacheHits     int64
	CacheMisses   int64
	TimesSaved    time.Duration
	StorageUsed   int64
	LastCleanup   time.Time
	HitRate       float64
}

// TestCacheEntry represents a cached test result
type TestCacheEntry struct {
	PackagePath  string            `json:"package_path"`
	TestCommand  string            `json:"test_command"`
	ContentHash  string            `json:"content_hash"`
	Success      bool              `json:"success"`
	Output       string            `json:"output"`
	Duration     time.Duration     `json:"duration"`
	Timestamp    time.Time         `json:"timestamp"`
	GoVersion    string            `json:"go_version"`
	Environment  map[string]string `json:"environment"`
	Dependencies []string          `json:"dependencies"`
	CoverageData []byte            `json:"coverage_data,omitempty"`
}

// PackageTestContext holds information needed for test execution and caching
type PackageTestContext struct {
	PackagePath  string
	TestType     string
	ContentHash  string
	Dependencies []string
	TestCommand  []string
	Environment  map[string]string
	CacheEnabled bool
}

// ParallelExecutionResult holds results from parallel test execution
type ParallelExecutionResult struct {
	PackagePath  string
	Success      bool
	Duration     time.Duration
	Output       string
	Error        error
	CacheHit     bool
	RetryCount   int
	CoverageFile string
}

// NewTestCacheSystem creates a new test caching system
func NewTestCacheSystem(logger log.Logger, cacheDir string) *TestCacheSystem {
	if logger == nil {
		logger = log.NewLogger()
	}

	if cacheDir == "" {
		cacheDir = filepath.Join(".", ".test-cache")
	}

	// Ensure cache directory exists
	_ = os.MkdirAll(cacheDir, 0755)

	system := &TestCacheSystem{
		logger:          logger,
		cacheDir:        cacheDir,
		maxCacheAge:     24 * time.Hour, // Default cache TTL
		enabledPackages: make(map[string]bool),
		cacheStats: &CacheStatistics{
			LastCleanup: time.Now(),
		},
	}

	// Enable caching for all packages by default
	system.enableCachingForAllPackages()

	return system
}

// EnableCaching enables caching for specific packages
func (tcs *TestCacheSystem) EnableCaching(packages []string) {
	tcs.mu.Lock()
	defer tcs.mu.Unlock()

	for _, pkg := range packages {
		tcs.enabledPackages[pkg] = true
	}

	tcs.logger.Info(fmt.Sprintf("Enabled caching for %d packages", len(packages)))
}

// IsCacheValid checks if a cached result is still valid
func (tcs *TestCacheSystem) IsCacheValid(ctx *PackageTestContext) (*TestCacheEntry, bool) {
	if !tcs.isCachingEnabled(ctx.PackagePath) {
		return nil, false
	}

	tcs.mu.Lock()
	tcs.cacheStats.TotalRequests++
	tcs.mu.Unlock()

	cacheFile := tcs.getCacheFilePath(ctx)
	entry, err := tcs.loadCacheEntry(cacheFile)
	if err != nil {
		tcs.recordCacheMiss()
		return nil, false
	}

	// Validate cache entry
	if !tcs.isEntryValid(entry, ctx) {
		tcs.recordCacheMiss()
		return nil, false
	}

	tcs.recordCacheHit(entry.Duration)
	return entry, true
}

// StoreCacheEntry saves a test result to cache
func (tcs *TestCacheSystem) StoreCacheEntry(ctx *PackageTestContext, result *ParallelExecutionResult) error {
	if !tcs.isCachingEnabled(ctx.PackagePath) {
		return nil
	}

	entry := &TestCacheEntry{
		PackagePath:  ctx.PackagePath,
		TestCommand:  strings.Join(ctx.TestCommand, " "),
		ContentHash:  ctx.ContentHash,
		Success:      result.Success,
		Output:       result.Output,
		Duration:     result.Duration,
		Timestamp:    time.Now(),
		GoVersion:    tcs.getGoVersion(),
		Environment:  ctx.Environment,
		Dependencies: ctx.Dependencies,
	}

	// Load coverage data if available
	if result.CoverageFile != "" {
		if coverageData, err := os.ReadFile(result.CoverageFile); err == nil {
			entry.CoverageData = coverageData
		}
	}

	cacheFile := tcs.getCacheFilePath(ctx)
	return tcs.saveCacheEntry(cacheFile, entry)
}

// ExecuteWithCache executes tests with intelligent caching
func (tcs *TestCacheSystem) ExecuteWithCache(contexts []*PackageTestContext, maxConcurrency int) ([]*ParallelExecutionResult, error) {
	tcs.logger.Info(fmt.Sprintf("Executing %d packages with cache optimization (concurrency=%d)", len(contexts), maxConcurrency))

	results := make([]*ParallelExecutionResult, len(contexts))
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for i, ctx := range contexts {
		wg.Add(1)
		go func(index int, context *PackageTestContext) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			result := tcs.executePackageWithCache(context)
			results[index] = result

		}(i, ctx)
	}

	wg.Wait()

	// Log cache statistics
	tcs.logCacheStatistics()

	return results, nil
}

// executePackageWithCache executes a single package with cache checking
func (tcs *TestCacheSystem) executePackageWithCache(ctx *PackageTestContext) *ParallelExecutionResult {
	startTime := time.Now()

	// Check cache first
	if entry, valid := tcs.IsCacheValid(ctx); valid {
		tcs.logger.Info(fmt.Sprintf("Cache hit for %s (saved %v)", ctx.PackagePath, entry.Duration))

		// Restore coverage file if cached
		if len(entry.CoverageData) > 0 {
			coverageFile := fmt.Sprintf("coverage_%s.out", tcs.sanitizePackageName(ctx.PackagePath))
			_ = os.WriteFile(coverageFile, entry.CoverageData, 0600)
		}

		return &ParallelExecutionResult{
			PackagePath:  ctx.PackagePath,
			Success:      entry.Success,
			Duration:     time.Since(startTime), // Actual time to retrieve from cache
			Output:       entry.Output,
			CacheHit:     true,
			CoverageFile: fmt.Sprintf("coverage_%s.out", tcs.sanitizePackageName(ctx.PackagePath)),
		}
	}

	// Execute test
	result := tcs.executeTestPackage(ctx)
	result.CacheHit = false

	// Store result in cache if successful or if it's a consistent failure
	if result.Success || tcs.shouldCacheFailure(ctx, result) {
		if err := tcs.StoreCacheEntry(ctx, result); err != nil {
			tcs.logger.Warn(fmt.Sprintf("Failed to cache result for %s: %v", ctx.PackagePath, err))
		}
	}

	return result
}

// CreateTestContext generates test context with content hashing
func (tcs *TestCacheSystem) CreateTestContext(packagePath, testType string, testCommand []string) (*PackageTestContext, error) {
	ctx := &PackageTestContext{
		PackagePath:  packagePath,
		TestType:     testType,
		TestCommand:  testCommand,
		Environment:  tcs.getRelevantEnvironment(),
		CacheEnabled: tcs.isCachingEnabled(packagePath),
	}

	// Calculate content hash
	hash, err := tcs.calculatePackageHash(packagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate package hash: %w", err)
	}
	ctx.ContentHash = hash

	// Get dependencies
	deps, err := tcs.getPackageDependencies(packagePath)
	if err != nil {
		tcs.logger.Warn(fmt.Sprintf("Failed to get dependencies for %s: %v", packagePath, err))
	}
	ctx.Dependencies = deps

	return ctx, nil
}

// CleanupCache removes old cache entries
func (tcs *TestCacheSystem) CleanupCache() error {
	tcs.logger.Info("Starting cache cleanup")

	removed := 0
	totalSize := int64(0)

	err := filepath.Walk(tcs.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".json") {
			// Check if cache entry is too old
			if time.Since(info.ModTime()) > tcs.maxCacheAge {
				if removeErr := os.Remove(path); removeErr == nil {
					removed++
				}
			} else {
				totalSize += info.Size()
			}
		}

		return nil
	})

	tcs.mu.Lock()
	tcs.cacheStats.StorageUsed = totalSize
	tcs.cacheStats.LastCleanup = time.Now()
	tcs.mu.Unlock()

	tcs.logger.Info(fmt.Sprintf("Cache cleanup completed: removed %d entries, %d bytes in use", removed, totalSize))

	return err
}

// GetCacheStatistics returns current cache performance metrics
func (tcs *TestCacheSystem) GetCacheStatistics() CacheStatistics {
	tcs.mu.RLock()
	defer tcs.mu.RUnlock()

	stats := *tcs.cacheStats
	if stats.TotalRequests > 0 {
		stats.HitRate = float64(stats.CacheHits) / float64(stats.TotalRequests) * 100
	}

	return stats
}

// Helper methods

func (tcs *TestCacheSystem) enableCachingForAllPackages() {
	// This would typically scan the project for packages
	commonPackages := []string{
		"./pkg/client/...",
		"./pkg/tree/...",
		"./pkg/network/...",
		"./pkg/metrics/...",
		"./pkg/security/...",
		"./pkg/secrets/...",
		"./pkg/replication/...",
	}

	for _, pkg := range commonPackages {
		tcs.enabledPackages[pkg] = true
	}
}

func (tcs *TestCacheSystem) isCachingEnabled(packagePath string) bool {
	tcs.mu.RLock()
	defer tcs.mu.RUnlock()

	// Check exact match first
	if enabled, exists := tcs.enabledPackages[packagePath]; exists {
		return enabled
	}

	// Check wildcard patterns
	for pattern := range tcs.enabledPackages {
		if strings.HasSuffix(pattern, "...") {
			prefix := strings.TrimSuffix(pattern, "...")
			if strings.HasPrefix(packagePath, prefix) {
				return tcs.enabledPackages[pattern]
			}
		}
	}

	return false
}

func (tcs *TestCacheSystem) getCacheFilePath(ctx *PackageTestContext) string {
	sanitized := tcs.sanitizePackageName(ctx.PackagePath)
	hashStr := ctx.ContentHash[:12] // Use first 12 chars of hash
	filename := fmt.Sprintf("%s_%s_%s.json", sanitized, ctx.TestType, hashStr)
	return filepath.Join(tcs.cacheDir, filename)
}

func (tcs *TestCacheSystem) sanitizePackageName(packagePath string) string {
	// Replace path separators and special characters with underscores
	sanitized := strings.ReplaceAll(packagePath, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, ".", "_")
	sanitized = strings.ReplaceAll(sanitized, "-", "_")
	return sanitized
}

func (tcs *TestCacheSystem) loadCacheEntry(cacheFile string) (*TestCacheEntry, error) {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}

	var entry TestCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (tcs *TestCacheSystem) saveCacheEntry(cacheFile string, entry *TestCacheEntry) error {
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, data, 0600)
}

func (tcs *TestCacheSystem) isEntryValid(entry *TestCacheEntry, ctx *PackageTestContext) bool {
	// Check age
	if time.Since(entry.Timestamp) > tcs.maxCacheAge {
		return false
	}

	// Check content hash
	if entry.ContentHash != ctx.ContentHash {
		return false
	}

	// Check Go version
	if entry.GoVersion != tcs.getGoVersion() {
		return false
	}

	// Check command
	currentCommand := strings.Join(ctx.TestCommand, " ")
	return entry.TestCommand == currentCommand
}

func (tcs *TestCacheSystem) calculatePackageHash(packagePath string) (string, error) {
	h := sha256.New()

	// Include Go source files in hash calculation
	err := filepath.Walk(strings.TrimPrefix(packagePath, "./"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".go") {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(h, file); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (tcs *TestCacheSystem) getPackageDependencies(packagePath string) ([]string, error) {
	// This would use `go list -deps` to get actual dependencies
	// For now, return a simplified list
	return []string{"github.com/stretchr/testify"}, nil
}

func (tcs *TestCacheSystem) getRelevantEnvironment() map[string]string {
	env := make(map[string]string)

	relevantVars := []string{
		"GO_VERSION",
		"GOOS",
		"GOARCH",
		"CGO_ENABLED",
		"TEST_TIMEOUT",
	}

	for _, v := range relevantVars {
		if value := os.Getenv(v); value != "" {
			env[v] = value
		}
	}

	return env
}

func (tcs *TestCacheSystem) getGoVersion() string {
	// This would execute `go version` and parse the result
	return "go1.21.0" // Mock version
}

func (tcs *TestCacheSystem) executeTestPackage(ctx *PackageTestContext) *ParallelExecutionResult {
	// This would execute the actual test command
	// For now, return a mock result
	return &ParallelExecutionResult{
		PackagePath: ctx.PackagePath,
		Success:     true,
		Duration:    time.Second,
		Output:      "PASS",
	}
}

func (tcs *TestCacheSystem) shouldCacheFailure(ctx *PackageTestContext, result *ParallelExecutionResult) bool {
	// Cache consistent failures to avoid repeated execution
	// This would implement logic to determine if a failure is worth caching
	return false
}

func (tcs *TestCacheSystem) recordCacheHit(savedTime time.Duration) {
	tcs.mu.Lock()
	defer tcs.mu.Unlock()

	tcs.cacheStats.CacheHits++
	tcs.cacheStats.TimesSaved += savedTime
}

func (tcs *TestCacheSystem) recordCacheMiss() {
	tcs.mu.Lock()
	defer tcs.mu.Unlock()

	tcs.cacheStats.CacheMisses++
}

func (tcs *TestCacheSystem) logCacheStatistics() {
	stats := tcs.GetCacheStatistics()

	tcs.logger.Info(fmt.Sprintf("Cache Statistics: %.1f%% hit rate, %d hits, %d misses, %v time saved",
		stats.HitRate, stats.CacheHits, stats.CacheMisses, stats.TimesSaved))
}
