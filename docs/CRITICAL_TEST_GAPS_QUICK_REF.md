# CRITICAL TEST GAPS - QUICK REFERENCE
## Top 20 Most Dangerous Untested Functions

**Date:** 2025-12-06
**Priority:** Address in this order

---

## 🔴 TIER 1: DATA LOSS RISK (Fix This Week)

### 1. `cmd/delete.go:runDelete()` - **0% Coverage**
**Lines:** 82-134
**Risk:** Could delete wrong images or all tags
**Tests Needed:** 15

```go
// MUST TEST:
- DeleteSingleImage_Success
- DeleteSingleImage_ImageNotFound
- DeleteSingleImage_AuthFailure
- DeleteAllTags_Confirmation
- DeleteAllTags_WithoutForce_Aborts
- DeleteAllTags_NetworkError_RollbackNotation
```

**Impact:** Irreversible data loss in production registries

---

### 2. `pkg/sync/batch.go:syncImage()` - **0% Coverage**
**Lines:** 251-352 (102 lines of CRITICAL code)
**Risk:** Core sync functionality completely untested
**Tests Needed:** 25

```go
// MUST TEST:
- SyncImage_ValidReferences_Success
- SyncImage_SourceNotFound_Error
- SyncImage_DestRegistryUnreachable_Retry
- SyncImage_AuthFailure_PropagatesError
- SyncImage_DigestMismatch_FailsIntegrity
- SyncImage_HugeBlob_MemoryBounded
- SyncImage_PartialTransfer_ResumesFromCheckpoint
```

**Impact:** Silent sync failures, data corruption, wrong images synced

---

### 3. `cmd/sync.go:resolveTags()` - **0% Coverage**
**Lines:** 218-287 (70 lines)
**Risk:** Wrong tags resolved → wrong images synced
**Tests Needed:** 12

```go
// MUST TEST:
- ResolveTags_AllTags_FetchesAllFromRegistry
- ResolveTags_LatestN_SortsCorrectly
- ResolveTags_Regex_FiltersCorrectly
- ResolveTags_EmptyResult_ReturnsError
- ResolveTags_NetworkError_Retries
- ResolveTags_AuthFailure_ClearError
```

**Impact:** Syncing wrong image versions to production

---

### 4. `pkg/artifacts/oci_handler.go:Replicate()` - **0% Coverage**
**Lines:** 105-168 (64 lines)
**Risk:** Artifact replication completely unverified
**Tests Needed:** 20

```go
// MUST TEST:
- Replicate_ContainerImage_Success
- Replicate_HelmChart_PreservesAnnotations
- Replicate_WASM_CorrectMediaType
- Replicate_MLModel_AllLayersIncluded
- Replicate_MultiArch_AllPlatforms
- Replicate_WithReferrers_PreservesLinks
- Replicate_ManifestNotFound_ClearError
```

**Impact:** Helm charts corrupted, WASM modules fail, ML models incomplete

---

### 5. `pkg/sync/filters.go:ApplyArchitectureFilter()` - **0% Coverage**
**Lines:** 183-210 (28 lines)
**Risk:** Wrong CPU architecture images synced
**Tests Needed:** 10

```go
// MUST TEST:
- ApplyArchitectureFilter_AMD64Only_FiltersARM
- ApplyArchitectureFilter_MultipleArchs_IncludesAll
- ApplyArchitectureFilter_ManifestFetchError_SkipsTag
- ApplyArchitectureFilter_InvalidManifest_HandlesGracefully
- ApplyArchitectureFilter_NoMatch_ReturnsEmpty
```

**Impact:** ARM images on x86 servers → images won't run

---

## 🟠 TIER 2: OPERATIONAL FAILURE (Fix Week 2)

### 6. `pkg/sync/batch.go:Execute()` - **0% Coverage**
**Lines:** 48-108 (61 lines)
**Risk:** Parallel batch execution unverified
**Tests Needed:** 15

```go
// MUST TEST:
- Execute_ParallelBatches_NoRaceConditions
- Execute_SemaphoreRespected_MaxParallelism
- Execute_ContextCancellation_GracefulShutdown
- Execute_PartialFailure_ContinuesIfConfigured
- Execute_GoroutineLeaks_AllCleanedup
```

**Impact:** Race conditions, goroutine leaks, memory exhaustion

---

### 7. `pkg/sync/size_estimator.go:EstimateImageSize()` - **0% Coverage**
**Lines:** 19-48 (30 lines)
**Risk:** Size estimation wrong → bad batch optimization
**Tests Needed:** 12

```go
// MUST TEST:
- EstimateImageSize_OCIManifest_CorrectSum
- EstimateImageSize_DockerV2Manifest_CorrectSum
- EstimateImageSize_MultiArch_SumsAllPlatforms
- EstimateImageSize_InvalidManifest_ReturnsError
- EstimateImageSize_HugeImage_NoOverflow
```

**Impact:** Poor performance, potential integer overflow crashes

---

### 8. `cmd/sync.go:buildSyncTasks()` - **0% Coverage**
**Lines:** 173-217 (45 lines)
**Risk:** Sync tasks built incorrectly
**Tests Needed:** 10

```go
// MUST TEST:
- BuildSyncTasks_MultipleImages_CorrectTaskCount
- BuildSyncTasks_DestRepoOverride_AppliedCorrectly
- BuildSyncTasks_FilteredTags_OnlyIncludesMatches
- BuildSyncTasks_EmptyImages_ReturnsError
- BuildSyncTasks_TagResolutionFails_PropagatesError
```

**Impact:** Wrong sync operations executed

---

### 9. `pkg/auth/credential_store.go:getFromHelper()` - **0% Coverage**
**Lines:** 218-257 (40 lines)
**Risk:** Docker credential helper integration fails
**Tests Needed:** 8

```go
// MUST TEST:
- GetFromHelper_MacOSKeychain_Success
- GetFromHelper_WindowsCredMan_Success
- GetFromHelper_HelperNotFound_FallsBackToFile
- GetFromHelper_HelperReturnsError_HandlesGracefully
- GetFromHelper_CorruptedCredential_ReturnsError
```

**Impact:** Authentication failures, can't access private registries

---

### 10. `pkg/artifacts/types.go:DetectArtifactType()` - **0% Coverage**
**Lines:** 258-297 (40 lines)
**Risk:** Wrong artifact type detected → wrong replication strategy
**Tests Needed:** 10

```go
// MUST TEST:
- DetectArtifactType_ContainerImage_Correct
- DetectArtifactType_HelmChart_Correct
- DetectArtifactType_WASMModule_Correct
- DetectArtifactType_MLModel_Correct
- DetectArtifactType_Unknown_ReturnsGeneric
```

**Impact:** Artifacts corrupted, wrong handling

---

## 🟡 TIER 3: SECURITY & RELIABILITY (Fix Week 3-4)

### 11. `cmd/inspect.go:runInspect()` - **0% Coverage**
**Lines:** 95-134 (40 lines)
**Risk:** Inspect command unreliable
**Tests Needed:** 12

---

### 12. `cmd/list_tags.go:runListTags()` - **0% Coverage**
**Lines:** 80-126 (47 lines)
**Risk:** Wrong tags listed
**Tests Needed:** 10

---

### 13. `pkg/client/acr/auth.go:getAccessToken()` - **0% Coverage**
**Lines:** 129-169 (41 lines)
**Risk:** Azure auth fails silently
**Tests Needed:** 8

---

### 14. `pkg/vulnerability/scanner.go:Scan()` - **0% Coverage**
**Lines:** 204-257 (54 lines)
**Risk:** Security vulnerabilities missed
**Tests Needed:** 15

---

### 15. `pkg/sync/batch.go:executeTask()` - **0% Coverage**
**Lines:** 185-248 (64 lines)
**Risk:** Retry logic with backoff unverified
**Tests Needed:** 10

---

### 16. `pkg/artifacts/oci_handler.go:copyBlob()` - **0% Coverage**
**Lines:** 374-403 (30 lines)
**Risk:** Blob copying fails, corruption possible
**Tests Needed:** 10

---

### 17. `pkg/artifacts/oci_handler.go:verifyDigest()` - **0% Coverage**
**Lines:** 411-428 (18 lines)
**Risk:** Data integrity not verified
**Tests Needed:** 5

---

### 18. `pkg/sync/filters.go:FilterWithMetadata()` - **0% Coverage**
**Lines:** 126-155 (30 lines)
**Risk:** Latest N filtering with timestamps broken
**Tests Needed:** 6

---

### 19. `pkg/sync/size_estimator.go:OptimizeBatchesWithSizeEstimation()` - **0% Coverage**
**Lines:** 137-205 (69 lines)
**Risk:** Batch optimization ineffective
**Tests Needed:** 8

---

### 20. `cmd/sync.go:convertToConfigRegistryConfig()` - **0% Coverage**
**Lines:** 288-336 (49 lines)
**Risk:** Registry config conversion wrong
**Tests Needed:** 8

---

## TESTING PRIORITY MATRIX

### By Risk Level:
```
🔴 CRITICAL (Tier 1): 5 functions  → 82 tests needed → Week 1
🟠 HIGH (Tier 2):     5 functions  → 65 tests needed → Week 2
🟡 MEDIUM (Tier 3):   10 functions → 100+ tests needed → Week 3-4
```

### By Impact:
```
Data Loss:           5 functions (delete, sync, resolve)
Corruption:          4 functions (artifacts, types, copy)
Auth/Security:       3 functions (auth, scan, credentials)
Performance:         3 functions (batch, optimize, estimate)
Operational:         5 functions (commands, list, inspect)
```

### By Lines of Code:
```
>100 lines: 1 function  (syncImage - 102 lines)
50-100:     4 functions (executeTask, Replicate, buildSyncTasks, etc.)
30-50:      8 functions
<30:        7 functions
```

---

## DAILY TEST TARGETS (Week 1)

### Monday:
**Focus:** Sync batch execution
- [ ] Write 25 tests for `pkg/sync/batch.go`
- [ ] Test Execute(), executeTask(), createBatches()
- [ ] Coverage: +10%

### Tuesday:
**Focus:** Sync image copying
- [ ] Write 25 tests for `syncImage()` function
- [ ] Test all error paths, auth, retries
- [ ] Coverage: +5%

### Wednesday:
**Focus:** Size estimation
- [ ] Write 30 tests for `size_estimator.go`
- [ ] Test all manifest types
- [ ] Coverage: +5%

### Thursday:
**Focus:** Artifacts
- [ ] Write 20 tests for `oci_handler.go`
- [ ] Test Replicate(), copyBlob(), verifyDigest()
- [ ] Coverage: +4%

### Friday:
**Focus:** Command & filters
- [ ] Write 20 tests for `cmd/sync.go`
- [ ] Write 10 tests for architecture filtering
- [ ] Coverage: +6%

**Week 1 Total:** +30% coverage (30.7% → 60.7%)

---

## SMOKE TEST TEMPLATE

**For IMMEDIATE use today:**

```go
// pkg/sync/batch_test.go
package sync_test

import (
    "context"
    "testing"
    "freightliner/pkg/sync"
    "github.com/stretchr/testify/assert"
)

// SMOKE TEST 1: Basic batch executor creation
func TestBatchExecutor_Creation_Smoke(t *testing.T) {
    config := &sync.Config{
        BatchSize: 10,
        Parallel: 3,
        RetryAttempts: 3,
        RetryBackoff: 2,
    }

    executor := sync.NewBatchExecutor(config, testLogger())
    assert.NotNil(t, executor, "Executor should not be nil")
}

// SMOKE TEST 2: Empty task list
func TestBatchExecutor_EmptyTasks_Smoke(t *testing.T) {
    config := &sync.Config{BatchSize: 10, Parallel: 3}
    executor := sync.NewBatchExecutor(config, testLogger())

    results, err := executor.Execute(context.Background(), []sync.SyncTask{})

    assert.NoError(t, err, "Empty tasks should not error")
    assert.Empty(t, results, "Results should be empty")
}

// SMOKE TEST 3: Calculate statistics with empty results
func TestCalculateStatistics_EmptyResults_Smoke(t *testing.T) {
    stats := sync.CalculateStatistics([]sync.SyncResult{})

    assert.Equal(t, 0, stats.TotalTasks)
    assert.Equal(t, 0, stats.CompletedTasks)
    assert.Equal(t, 0.0, stats.SuccessRate)
}

// Helper
func testLogger() log.Logger {
    return log.NewLogger(log.InfoLevel)
}
```

---

## QUICK COMMAND REFERENCE

### Check coverage for specific package:
```bash
go test -coverprofile=coverage.out ./pkg/sync
go tool cover -func=coverage.out
```

### Run tests with race detection:
```bash
go test -race ./pkg/sync
```

### Generate HTML coverage report:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

### Find all functions with 0% coverage:
```bash
go tool cover -func=coverage.out | grep "0.0%" | wc -l
```

### Test single function:
```bash
go test -run TestSyncImage_ValidReferences ./pkg/sync -v
```

---

## NEXT 3 HOURS (TODAY)

### Hour 1: Setup
1. Create 5 test files (batch, size_estimator, oci_handler, types, sync)
2. Add test fixtures and helpers
3. Write first 3 smoke tests

### Hour 2: Sync Batch Tests
4. Write 10 tests for `Execute()`
5. Write 8 tests for `executeTask()`
6. Write 7 tests for retry logic

### Hour 3: Sync Image Tests
7. Write 10 tests for `syncImage()` success cases
8. Write 8 tests for `syncImage()` error cases
9. Run all tests, check coverage

**Goal:** +15% coverage by end of day

---

## SUCCESS CRITERIA (End of Week 1)

### Coverage:
- Overall: 30.7% → 60%+ ✅
- pkg/sync: 25% → 70%+ ✅
- pkg/artifacts: 0% → 50%+ ✅
- cmd/sync: 0% → 60%+ ✅

### Tests:
- New tests written: 200+ ✅
- All tests passing: Yes ✅
- No flaky tests: Yes ✅

### CI:
- Coverage gate enabled: 50% minimum ✅
- Automated coverage reports: Yes ✅
- PR checks enforced: Yes ✅

---

## QUESTIONS TO ASK

### Before Starting:
1. Can we pause feature work for 1 week?
2. Who will review the tests?
3. What's the CI setup process?

### During Testing:
4. Are mocks needed for registry clients?
5. Should we use Docker Compose for integration tests?
6. How to handle flaky network tests?

### After Week 1:
7. Did we hit 60% coverage?
8. Are tests running fast (<1min)?
9. Ready to enforce coverage gates?

---

**This is your quick reference. Print it. Pin it. Live by it.**

**Goal:** Zero untested critical paths by end of Week 4.
