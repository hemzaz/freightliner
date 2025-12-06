# TEST COVERAGE ACTION PLAN
## Freightliner - Path to 85% Coverage

**Current:** 30.7% | **Target:** 85% | **Gap:** 54.3%

---

## WEEK 1: CRITICAL SYNC ENGINE (Target: +20%)

### Day 1-2: pkg/sync/batch_test.go (NEW FILE)

**Priority:** 🔴 CRITICAL - Core sync functionality untested

**Tests to Write (50 tests, ~2500 lines):**

```go
package sync_test

import (
    "context"
    "testing"
    "time"

    "freightliner/pkg/sync"
    "freightliner/pkg/client"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// === CONSTRUCTOR TESTS ===
func TestNewBatchExecutor_ValidConfig(t *testing.T) {
    config := &sync.Config{BatchSize: 10, Parallel: 3}
    executor := sync.NewBatchExecutor(config, testLogger())
    assert.NotNil(t, executor)
}

func TestNewBatchExecutorWithFactory_NilFactory(t *testing.T) {
    // Test behavior with nil factory
}

// === BATCH CREATION TESTS ===
func TestCreateBatches_EmptyTasks(t *testing.T) {
    // Should return empty slice
}

func TestCreateBatches_SingleTask(t *testing.T) {
    // Should create one batch with one task
}

func TestCreateBatches_ExactBatchSize(t *testing.T) {
    // 10 tasks, batch size 10 → 1 batch
}

func TestCreateBatches_MultipleBatches(t *testing.T) {
    // 25 tasks, batch size 10 → 3 batches (10, 10, 5)
}

func TestCreateBatches_ZeroBatchSize(t *testing.T) {
    // Should use default batch size (10)
}

func TestCreateBatches_NegativeBatchSize(t *testing.T) {
    // Should use default batch size
}

// === EXECUTE TESTS ===
func TestExecute_SingleBatch_Success(t *testing.T) {
    // Setup mock registry clients
    // Execute 5 tasks in single batch
    // Verify all tasks complete successfully
}

func TestExecute_MultipleBatches_Parallel(t *testing.T) {
    // Execute 30 tasks with batch size 10, parallel 3
    // Verify batches run in parallel
    // Check execution time is ~3x batch time, not 30x
}

func TestExecute_ContextCancellation(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    // Start execution
    // Cancel context mid-execution
    // Verify graceful shutdown
    // Verify no goroutine leaks
}

func TestExecute_PartialFailure_ContinueOnError(t *testing.T) {
    // Config with ContinueOnError=true
    // Some tasks fail
    // Verify execution continues
    // Verify failed tasks in results
}

func TestExecute_PartialFailure_StopOnError(t *testing.T) {
    // Config with ContinueOnError=false
    // First batch has failure
    // Verify execution stops
    // Verify error returned
}

func TestExecute_AllTasksFail(t *testing.T) {
    // All registry operations fail
    // Verify all results marked failed
    // Verify error returned
}

// === RETRY TESTS ===
func TestExecuteTask_FirstAttemptSuccess(t *testing.T) {
    // Task succeeds on first try
    // Verify retries=0
    // Verify success=true
}

func TestExecuteTask_SucceedsOnSecondAttempt(t *testing.T) {
    // First attempt fails (network error)
    // Second attempt succeeds
    // Verify retries=1
    // Verify exponential backoff applied
}

func TestExecuteTask_ExhaustRetries(t *testing.T) {
    // All retry attempts fail
    // Verify retries=config.RetryAttempts
    // Verify success=false
    // Verify error in result
}

func TestExecuteTask_RetryBackoff(t *testing.T) {
    // Verify exponential backoff timing
    // First retry: 2s, second: 4s, third: 8s
}

// === SYNC IMAGE TESTS (CRITICAL) ===
func TestSyncImage_ValidReferences_Success(t *testing.T) {
    // Mock source and dest registries
    // Sync single-arch image
    // Verify bytes copied
    // Verify manifest and layers copied
}

func TestSyncImage_InvalidSourceReference(t *testing.T) {
    // Invalid source image name
    // Verify error returned
    // Verify no partial state
}

func TestSyncImage_InvalidDestReference(t *testing.T) {
    // Invalid dest image name
    // Verify error
}

func TestSyncImage_SourceNotFound(t *testing.T) {
    // Source image doesn't exist
    // Verify proper error
}

func TestSyncImage_SourceRegistryUnreachable(t *testing.T) {
    // Source registry offline
    // Verify network error
    // Verify retryable
}

func TestSyncImage_DestRegistryUnreachable(t *testing.T) {
    // Dest registry offline
    // Verify error
}

func TestSyncImage_AuthenticationFailure_Source(t *testing.T) {
    // Invalid credentials for source
    // Verify 401 error
}

func TestSyncImage_AuthenticationFailure_Dest(t *testing.T) {
    // Invalid credentials for dest
    // Verify 401 error
}

func TestSyncImage_InsufficientPermissions_Dest(t *testing.T) {
    // No push permission on dest
    // Verify 403 error
}

func TestSyncImage_ManifestNotFound(t *testing.T) {
    // Tag doesn't exist
    // Verify 404 error
}

func TestSyncImage_MultiArchImage(t *testing.T) {
    // Sync image with multiple platforms
    // Verify all platforms copied
}

func TestSyncImage_LargeImage(t *testing.T) {
    // Image with many layers (100+)
    // Verify all layers copied
    // Verify no timeout
}

func TestSyncImage_HugeBlob(t *testing.T) {
    // Layer > 1GB
    // Verify streaming works
    // Verify memory usage bounded
}

func TestSyncImage_DigestMismatch(t *testing.T) {
    // Corrupted blob
    // Verify integrity check fails
    // Verify error returned
}

func TestSyncImage_PartialTransfer_Retry(t *testing.T) {
    // Transfer interrupted
    // Verify retry works
    // Verify resume from checkpoint
}

func TestSyncImage_DuplicateBlobs(t *testing.T) {
    // Image with shared layers
    // Verify deduplication works
    // Verify each blob copied once
}

func TestSyncImage_AlreadyExists(t *testing.T) {
    // Image already in dest
    // Verify handled gracefully
    // Verify no redundant copy
}

// === CONCURRENCY TESTS ===
func TestExecute_RaceConditions(t *testing.T) {
    // Run with -race flag
    // Execute many tasks concurrently
    // Verify no data races
}

func TestExecute_GoroutineLeaks(t *testing.T) {
    // Count goroutines before
    // Execute tasks
    // Wait for completion
    // Verify goroutine count returns to baseline
}

func TestExecute_MemoryBounded(t *testing.T) {
    // Execute 1000 tasks
    // Monitor memory usage
    // Verify bounded growth
    // Verify no memory leak
}

func TestExecute_SemaphoreRespected(t *testing.T) {
    // Set parallelism=3
    // Verify max 3 batches running concurrently
    // Use channel blocking to test
}

func TestExecute_BatchOrdering(t *testing.T) {
    // Verify batches processed in order
    // Verify results match input order
}

// === OPTIMIZATION TESTS ===
func TestOptimizeBatches_EmptyTasks(t *testing.T) {
    result := sync.OptimizeBatches([]sync.SyncTask{})
    assert.Empty(t, result)
}

func TestOptimizeBatches_SingleTask(t *testing.T) {
    // One task unchanged
}

func TestOptimizeBatches_PriorityOrdering(t *testing.T) {
    tasks := []sync.SyncTask{
        {Priority: 1, SourceRegistry: "reg1"},
        {Priority: 3, SourceRegistry: "reg1"},
        {Priority: 2, SourceRegistry: "reg1"},
    }
    result := sync.OptimizeBatches(tasks)
    // Verify sorted: 3, 2, 1
}

func TestOptimizeBatches_RegistryGrouping(t *testing.T) {
    tasks := []sync.SyncTask{
        {Priority: 1, SourceRegistry: "reg2"},
        {Priority: 1, SourceRegistry: "reg1"},
        {Priority: 1, SourceRegistry: "reg2"},
        {Priority: 1, SourceRegistry: "reg1"},
    }
    result := sync.OptimizeBatches(tasks)
    // Verify grouped by registry
    // reg1, reg1, reg2, reg2 or reg2, reg2, reg1, reg1
}

// === STATISTICS TESTS ===
func TestCalculateStatistics_EmptyResults(t *testing.T) {
    stats := sync.CalculateStatistics([]sync.SyncResult{})
    assert.Equal(t, 0, stats.TotalTasks)
}

func TestCalculateStatistics_AllSuccess(t *testing.T) {
    results := []sync.SyncResult{
        {Success: true, BytesCopied: 1000, Duration: 100},
        {Success: true, BytesCopied: 2000, Duration: 200},
    }
    stats := sync.CalculateStatistics(results)
    assert.Equal(t, 2, stats.CompletedTasks)
    assert.Equal(t, 0, stats.FailedTasks)
    assert.Equal(t, 100.0, stats.SuccessRate)
    assert.Equal(t, int64(3000), stats.TotalBytes)
}

func TestCalculateStatistics_AllFailure(t *testing.T) {
    results := []sync.SyncResult{
        {Success: false, Error: errors.New("fail1")},
        {Success: false, Error: errors.New("fail2")},
    }
    stats := sync.CalculateStatistics(results)
    assert.Equal(t, 0, stats.CompletedTasks)
    assert.Equal(t, 2, stats.FailedTasks)
    assert.Equal(t, 0.0, stats.SuccessRate)
}

func TestCalculateStatistics_MixedResults(t *testing.T) {
    results := []sync.SyncResult{
        {Success: true, BytesCopied: 1000, Duration: 100},
        {Success: false},
        {Success: true, BytesCopied: 2000, Duration: 200},
        {Skipped: true},
    }
    stats := sync.CalculateStatistics(results)
    assert.Equal(t, 4, stats.TotalTasks)
    assert.Equal(t, 2, stats.CompletedTasks)
    assert.Equal(t, 1, stats.FailedTasks)
    assert.Equal(t, 1, stats.SkippedTasks)
    assert.Equal(t, 50.0, stats.SuccessRate)
}

func TestCalculateStatistics_ThroughputCalculation(t *testing.T) {
    // 10MB in 1 second = 10 MB/s
    results := []sync.SyncResult{
        {Success: true, BytesCopied: 10 * 1024 * 1024, Duration: 1000},
    }
    stats := sync.CalculateStatistics(results)
    assert.InDelta(t, 10.0, stats.ThroughputMBps, 0.1)
}

func TestCalculateStatistics_ZeroDuration(t *testing.T) {
    results := []sync.SyncResult{
        {Success: true, BytesCopied: 1000, Duration: 0},
    }
    stats := sync.CalculateStatistics(results)
    // Should handle division by zero
    assert.GreaterOrEqual(t, stats.ThroughputMBps, 0.0)
}

// === ESTIMATION TESTS ===
func TestEstimateDuration_EmptyTasks(t *testing.T) {
    duration := sync.EstimateDuration([]sync.SyncTask{}, 3, 10)
    assert.Equal(t, time.Duration(0), duration)
}

func TestEstimateDuration_SingleBatch(t *testing.T) {
    tasks := make([]sync.SyncTask, 5)
    duration := sync.EstimateDuration(tasks, 3, 10)
    // 1 batch * 10 tasks * 30s = 300s
    assert.Equal(t, 5*30*time.Second, duration)
}

func TestEstimateDuration_MultipleBatches(t *testing.T) {
    tasks := make([]sync.SyncTask, 25)
    // 25 tasks, batch=10, parallel=2
    // 3 batches (10, 10, 5)
    // 2 rounds: [batch1, batch2], [batch3]
    duration := sync.EstimateDuration(tasks, 2, 10)
    // 2 rounds * 10 tasks * 30s = 600s
    assert.Equal(t, 2*10*30*time.Second, duration)
}
```

**Estimated Impact:** +15% coverage

---

### Day 3: pkg/sync/size_estimator_test.go (NEW FILE)

**Priority:** 🔴 CRITICAL - Size estimation affects batch optimization

**Tests to Write (30 tests, ~1500 lines):**

```go
package sync_test

// === BASIC SIZE ESTIMATION ===
func TestEstimateImageSize_OCIManifest_SingleLayer(t *testing.T)
func TestEstimateImageSize_OCIManifest_MultipleLayers(t *testing.T)
func TestEstimateImageSize_DockerV2Manifest_Success(t *testing.T)
func TestEstimateImageSize_InvalidManifest_Error(t *testing.T)
func TestEstimateImageSize_EmptyManifest_Error(t *testing.T)
func TestEstimateImageSize_ManifestFetchError(t *testing.T)

// === MULTI-ARCH MANIFESTS ===
func TestEstimateMultiArchManifestSize_OCIIndex(t *testing.T)
func TestEstimateMultiArchManifestSize_DockerList(t *testing.T)
func TestEstimateMultiArchManifestSize_MultiplePlatforms(t *testing.T)
func TestEstimateMultiArchManifestSize_InvalidFormat(t *testing.T)

// === EDGE CASES ===
func TestEstimateImageSize_ZeroSizeLayers(t *testing.T)
func TestEstimateImageSize_HugeImage_NoOverflow(t *testing.T)
func TestEstimateImageSize_NegativeSize_Error(t *testing.T)
func TestEstimateImageSize_MissingConfigSize(t *testing.T)

// === DOCKER V1 (DEPRECATED) ===
func TestEstimateDockerV1ManifestSize_NotSupported(t *testing.T)
func TestEstimateDockerV1ManifestSize_ReturnsError(t *testing.T)

// === BATCH OPTIMIZATION WITH SIZE ===
func TestOptimizeBatchesWithSizeEstimation_SmallFirst(t *testing.T)
func TestOptimizeBatchesWithSizeEstimation_PriorityFirst(t *testing.T)
func TestOptimizeBatchesWithSizeEstimation_RegistryGrouping(t *testing.T)
func TestOptimizeBatchesWithSizeEstimation_EstimationErrors(t *testing.T)
func TestOptimizeBatchesWithSizeEstimation_NilEstimator(t *testing.T)
func TestOptimizeBatchesWithSizeEstimation_MixedSizes(t *testing.T)

// === BATCH SIZE CALCULATION ===
func TestEstimateBatchSize_EmptyTasks(t *testing.T)
func TestEstimateBatchSize_SingleTask(t *testing.T)
func TestEstimateBatchSize_MultipleTasks(t *testing.T)
func TestEstimateBatchSize_PartialEstimationFailure(t *testing.T)
func TestEstimateBatchSize_AllEstimationsFail(t *testing.T)

// === BATCH SIZES MAP ===
func TestEstimateBatchSizes_ReturnsMap(t *testing.T)
func TestEstimateBatchSizes_SkipsFailures(t *testing.T)
func TestEstimateBatchSizes_CorrectIndices(t *testing.T)
```

**Estimated Impact:** +5% coverage

---

### Day 4-5: pkg/artifacts/ tests (NEW FILES)

**Priority:** 🔴 CRITICAL - Artifact handling completely untested

**Files to Create:**
- `pkg/artifacts/oci_handler_test.go` (40 tests, ~2000 lines)
- `pkg/artifacts/types_test.go` (30 tests, ~1500 lines)

**Key Tests:**

```go
// oci_handler_test.go
func TestNewHandler_ValidOptions(t *testing.T)
func TestNewHandler_NilRegistry_Error(t *testing.T)
func TestReplicate_ContainerImage_Success(t *testing.T)
func TestReplicate_ContainerImage_ManifestError(t *testing.T)
func TestReplicate_MultiArchIndex_AllPlatforms(t *testing.T)
func TestReplicate_HelmChart_Success(t *testing.T)
func TestReplicate_WASMModule_Success(t *testing.T)
func TestReplicate_MLModel_Success(t *testing.T)
func TestReplicateWithReferrers_PreservesLinks(t *testing.T)
func TestReplicateWithReferrers_NoReferrers(t *testing.T)
func TestCopyBlob_Success(t *testing.T)
func TestCopyBlob_AlreadyExists_SkipsNot(t *testing.T)
func TestCopyBlob_NetworkError_Retry(t *testing.T)
func TestCopyBlob_DigestMismatch_Error(t *testing.T)
func TestMountBlob_CrossRepo_Success(t *testing.T)
func TestMountBlob_NotSupported_FallbackToCopy(t *testing.T)
func TestVerifyDigest_Match(t *testing.T)
func TestVerifyDigest_Mismatch_Error(t *testing.T)
func TestArtifactExists_True(t *testing.T)
func TestArtifactExists_False(t *testing.T)
func TestListReferrers_Success(t *testing.T)
func TestListReferrers_Empty(t *testing.T)

// types_test.go
func TestDetectArtifactType_ContainerImage(t *testing.T)
func TestDetectArtifactType_HelmChart(t *testing.T)
func TestDetectArtifactType_WASMModule(t *testing.T)
func TestDetectArtifactType_MLModel(t *testing.T)
func TestDetectArtifactType_SBOM(t *testing.T)
func TestDetectArtifactType_Signature(t *testing.T)
func TestDetectArtifactType_Unknown(t *testing.T)
func TestIsContainerImage_DockerMediaType(t *testing.T)
func TestIsContainerImage_OCIMediaType(t *testing.T)
func TestIsMultiArch_Index(t *testing.T)
func TestIsMultiArch_ManifestList(t *testing.T)
func TestIsMultiArch_SingleImage_False(t *testing.T)
func TestIsHelm_True(t *testing.T)
func TestIsHelm_False(t *testing.T)
func TestIsWASM_True(t *testing.T)
func TestIsWASM_False(t *testing.T)
```

**Estimated Impact:** +8% coverage

---

## WEEK 2: COMMAND LINE & AUTH (Target: +15%)

### Day 1-2: cmd/sync_test.go (NEW FILE)

**Priority:** 🔴 CRITICAL - Main sync command untested

**Tests to Write (40 tests, ~2000 lines):**

```go
// Basic execution
func TestRunSync_ValidConfig_Success(t *testing.T)
func TestRunSync_InvalidConfig_Error(t *testing.T)
func TestRunSync_ConfigFileNotFound_Error(t *testing.T)
func TestRunSync_MalformedYAML_Error(t *testing.T)
func TestRunSync_DryRun_NoActualSync(t *testing.T)
func TestRunSync_ParallelismOverride(t *testing.T)

// Task building
func TestBuildSyncTasks_EmptyImages_Error(t *testing.T)
func TestBuildSyncTasks_SingleImage_SingleTag(t *testing.T)
func TestBuildSyncTasks_MultipleImages_MultipleTags(t *testing.T)
func TestBuildSyncTasks_TagResolutionError(t *testing.T)
func TestBuildSyncTasks_DestinationRepositoryOverride(t *testing.T)
func TestBuildSyncTasks_ArchitectureFiltering(t *testing.T)

// Tag resolution
func TestResolveTags_SpecificTags(t *testing.T)
func TestResolveTags_AllTags(t *testing.T)
func TestResolveTags_TagRegex(t *testing.T)
func TestResolveTags_SemverConstraint(t *testing.T)
func TestResolveTags_LatestN(t *testing.T)
func TestResolveTags_EmptyResult_Error(t *testing.T)
func TestResolveTags_NetworkError_Retry(t *testing.T)
func TestResolveTags_AuthenticationFailure(t *testing.T)

// Registry config conversion
func TestConvertToConfigRegistryConfig_Docker(t *testing.T)
func TestConvertToConfigRegistryConfig_ACR(t *testing.T)
func TestConvertToConfigRegistryConfig_GHCR(t *testing.T)
func TestConvertToConfigRegistryConfig_Harbor(t *testing.T)
func TestConvertToConfigRegistryConfig_Generic(t *testing.T)

// Type mapping
func TestMapRegistryType_ValidTypes(t *testing.T)
func TestMapRegistryType_UnknownType(t *testing.T)

// Result display
func TestDisplaySyncResults_AllSuccess(t *testing.T)
func TestDisplaySyncResults_PartialFailure(t *testing.T)
func TestDisplaySyncResults_AllFailure(t *testing.T)
func TestDisplaySyncResults_WithSkipped(t *testing.T)

// Formatting
func TestFormatBytes_Various(t *testing.T)
```

**Estimated Impact:** +5% coverage

---

### Day 3: Other Command Tests

**Files to Create:**
- `cmd/delete_test.go` (20 tests, ~1000 lines)
- `cmd/inspect_test.go` (25 tests, ~1250 lines)
- `cmd/list_tags_test.go` (20 tests, ~1000 lines)

**Estimated Impact:** +5% coverage

---

### Day 4-5: Authentication Tests

**Enhance existing and add new:**
- `pkg/auth/credential_store_test.go` (add helper tests)
- `pkg/client/acr/auth_test.go` (NEW - 30 tests)

**Estimated Impact:** +5% coverage

---

## WEEK 3-4: REGISTRY CLIENTS (Target: +20%)

**Create test files for each registry client:**

### Week 3:
- `pkg/client/dockerhub/client_test.go` (25 tests)
- `pkg/client/ghcr/client_test.go` (25 tests)
- `pkg/client/acr/client_test.go` (enhance existing)

### Week 4:
- `pkg/client/harbor/client_test.go` (25 tests)
- `pkg/client/quay/client_test.go` (25 tests)
- `pkg/client/generic/client_test.go` (enhance existing)

**Estimated Impact:** +20% coverage

---

## WEEK 5: INTEGRATION TESTS (Target: +10%)

**Create:** `tests/integration/sync_integration_test.go`

### End-to-End Scenarios (30 tests):

```go
func TestE2E_DockerHubToPrivateRegistry(t *testing.T)
func TestE2E_ACRToHarbor_MultiArch(t *testing.T)
func TestE2E_GHCRToQuay_WithAuth(t *testing.T)
func TestE2E_SyncWithTagFiltering(t *testing.T)
func TestE2E_SyncWithArchitectureFilter(t *testing.T)
func TestE2E_SyncWithRetries_NetworkFailure(t *testing.T)
func TestE2E_SyncWithCheckpoint_Resume(t *testing.T)
func TestE2E_ConcurrentSyncs_NoConflict(t *testing.T)
func TestE2E_LargeImageSync_MemoryBounded(t *testing.T)
func TestE2E_HelmChartReplication(t *testing.T)
func TestE2E_MultiArchImage_AllPlatforms(t *testing.T)
func TestE2E_RateLimiting_BackoffRetry(t *testing.T)
```

**Estimated Impact:** +10% coverage

---

## WEEK 6: VULNERABILITY & SBOM (Target: +5%)

### Files to Create:
- `pkg/vulnerability/scanner_test.go` (enhance)
- `pkg/vulnerability/grype_integration_test.go` (enhance)
- `pkg/sbom/generator_test.go` (NEW)

**Estimated Impact:** +5% coverage

---

## TRACKING PROGRESS

### Daily Coverage Check:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tail -1
```

### Weekly Goals:
- **End of Week 1:** 50% (from 30.7%)
- **End of Week 2:** 65% (from 50%)
- **End of Week 3:** 75% (from 65%)
- **End of Week 4:** 80% (from 75%)
- **End of Week 5:** 85% (from 80%)
- **End of Week 6:** 87%+ (maintenance)

---

## CI INTEGRATION

### Add to `.github/workflows/ci.yml`:

```yaml
- name: Test Coverage Gate
  run: |
    go test -coverprofile=coverage.out ./...
    coverage=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')

    if (( $(echo "$coverage < 50" | bc -l) )); then
      echo "Coverage $coverage% is below 50% threshold"
      exit 1
    fi

    echo "Coverage: $coverage%"

- name: Upload Coverage Report
  uses: codecov/codecov-action@v3
  with:
    file: ./coverage.out
```

---

## SUCCESS CRITERIA

### Week 1:
- ✅ All sync engine functions >70% coverage
- ✅ All artifact handler functions >60% coverage
- ✅ At least 100 new passing tests

### Week 2:
- ✅ All command functions >50% coverage
- ✅ Auth helper functions >70% coverage
- ✅ At least 150 new passing tests

### Week 3-4:
- ✅ All registry clients >70% coverage
- ✅ At least 200 new passing tests

### Week 5:
- ✅ 20+ E2E integration tests passing
- ✅ Critical user journeys covered

### Week 6:
- ✅ Overall coverage >85%
- ✅ All packages >70% coverage
- ✅ CI gates enforced

---

## RESOURCES NEEDED

### Team:
- 2 senior developers (full-time, 6 weeks)
- 1 QA engineer (part-time, testing integration tests)

### Tools:
- Test coverage tools (already have: `go test -cover`)
- Mocking library (already have: testify/mock)
- Integration test infrastructure (Docker Compose for registries)

### Estimated Cost:
- **Full plan:** $80,000 (6 weeks, 2 devs)
- **Minimum viable (60%):** $30,000 (4 weeks, 1 dev)

---

## NEXT IMMEDIATE ACTION

**TODAY:**
```bash
cd /Users/elad/PROJ/freightliner

# Create test file structure
touch pkg/sync/batch_test.go
touch pkg/sync/size_estimator_test.go
touch pkg/artifacts/oci_handler_test.go
touch pkg/artifacts/types_test.go
touch cmd/sync_test.go

# Write first 5 smoke tests
# Run tests
go test ./pkg/sync -v
go test ./pkg/artifacts -v
go test ./cmd -v

# Check coverage improvement
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tail -1
```

**Goal:** Add 10-15% coverage in first 2 days.
