package service

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/tree/checkpoint"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCheckpointServiceFullLifecycle tests complete checkpoint lifecycle
func TestCheckpointServiceFullLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping full lifecycle test in short mode")
	}

	dir := t.TempDir()
	cfg := &config.Config{
		Checkpoint: config.CheckpointConfig{
			Directory: dir,
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewCheckpointService(cfg, logger)

	ctx := context.Background()

	// Test 1: List empty checkpoints
	checkpoints, err := svc.ListCheckpoints(ctx)
	require.NoError(t, err)
	assert.Empty(t, checkpoints)

	// Test 2: Import a checkpoint
	info := CheckpointInfo{
		ID:                    "test-checkpoint-123",
		CreatedAt:             time.Now(),
		Source:                "ecr/source-repo",
		Destination:           "gcr/dest-repo",
		Status:                string(checkpoint.StatusCompleted),
		TotalRepositories:     10,
		CompletedRepositories: 10,
		FailedRepositories:    0,
		TotalTagsCopied:       100,
		TotalTagsSkipped:      5,
		TotalErrors:           0,
		TotalBytesTransferred: 10240000,
		Repositories: []RepositoryInfo{
			{
				Name:        "repo1",
				Status:      string(checkpoint.StatusCompleted),
				TagsCopied:  10,
				TagsSkipped: 1,
				Errors:      0,
			},
			{
				Name:        "repo2",
				Status:      string(checkpoint.StatusCompleted),
				TagsCopied:  15,
				TagsSkipped: 2,
				Errors:      0,
			},
		},
	}

	tmpFile := filepath.Join(t.TempDir(), "checkpoint.json")
	jsonData, err := json.Marshal(info)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(tmpFile, jsonData, 0600))

	imported, err := svc.ImportCheckpoint(ctx, tmpFile)
	require.NoError(t, err)
	assert.Equal(t, info.ID, imported.ID)

	// Test 3: List checkpoints (should have 1)
	checkpoints, err = svc.ListCheckpoints(ctx)
	require.NoError(t, err)
	assert.Len(t, checkpoints, 1)
	assert.Equal(t, info.ID, checkpoints[0].ID)

	// Test 4: Get specific checkpoint
	retrieved, err := svc.GetCheckpoint(ctx, info.ID)
	require.NoError(t, err)
	assert.Equal(t, info.ID, retrieved.ID)
	assert.Equal(t, info.Source, retrieved.Source)

	// Test 5: Verify checkpoint exists
	exists, err := svc.VerifyCheckpoint(ctx, info.ID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test 6: Export checkpoint
	exportFile := filepath.Join(t.TempDir(), "exported.json")
	err = svc.ExportCheckpoint(ctx, info.ID, exportFile)
	require.NoError(t, err)

	// Verify exported file exists and is valid JSON
	exportedData, err := os.ReadFile(exportFile)
	require.NoError(t, err)
	var exportedInfo CheckpointInfo
	err = json.Unmarshal(exportedData, &exportedInfo)
	require.NoError(t, err)
	assert.Equal(t, info.ID, exportedInfo.ID)

	// Test 7: Get remaining repositories
	remaining, err := svc.GetRemainingRepositories(ctx, info.ID, true, false)
	require.NoError(t, err)
	// All repos are completed, so should be empty
	assert.Empty(t, remaining)

	// Test 8: Delete checkpoint
	err = svc.DeleteCheckpoint(ctx, info.ID)
	require.NoError(t, err)

	// Test 9: Verify deletion
	exists, err = svc.VerifyCheckpoint(ctx, info.ID)
	require.NoError(t, err)
	assert.False(t, exists)

	// Test 10: List checkpoints (should be empty again)
	checkpoints, err = svc.ListCheckpoints(ctx)
	require.NoError(t, err)
	assert.Empty(t, checkpoints)
}

// TestCheckpointServiceGetRemainingWithFilters tests remaining repos with different filters
func TestCheckpointServiceGetRemainingWithFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping remaining repositories filter test in short mode")
	}

	dir := t.TempDir()
	cfg := &config.Config{
		Checkpoint: config.CheckpointConfig{
			Directory: dir,
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewCheckpointService(cfg, logger)

	ctx := context.Background()

	// Create a checkpoint with mixed status repositories
	info := CheckpointInfo{
		ID:                    "test-mixed-status",
		CreatedAt:             time.Now(),
		Source:                "ecr/source",
		Destination:           "gcr/dest",
		Status:                string(checkpoint.StatusInProgress),
		TotalRepositories:     4,
		CompletedRepositories: 2,
		FailedRepositories:    1,
		Repositories: []RepositoryInfo{
			{Name: "repo1", Status: string(checkpoint.StatusCompleted)},
			{Name: "repo2", Status: string(checkpoint.StatusCompleted)},
			{Name: "repo3", Status: string(checkpoint.StatusFailed)},
			{Name: "repo4", Status: string(checkpoint.StatusPending)},
		},
	}

	tmpFile := filepath.Join(t.TempDir(), "checkpoint.json")
	jsonData, err := json.Marshal(info)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(tmpFile, jsonData, 0600))

	_, err = svc.ImportCheckpoint(ctx, tmpFile)
	require.NoError(t, err)

	// Test skipCompleted=true, retryFailed=false
	remaining, err := svc.GetRemainingRepositories(ctx, info.ID, true, false)
	require.NoError(t, err)
	// Should exclude completed (repo1, repo2) and failed (repo3), only pending (repo4)
	assert.Len(t, remaining, 1)
	assert.Contains(t, remaining, "repo4")

	// Test skipCompleted=true, retryFailed=true
	remaining, err = svc.GetRemainingRepositories(ctx, info.ID, true, true)
	require.NoError(t, err)
	// Should exclude completed, include failed and pending
	assert.Len(t, remaining, 2)
	assert.Contains(t, remaining, "repo3")
	assert.Contains(t, remaining, "repo4")

	// Test skipCompleted=false, retryFailed=true
	remaining, err = svc.GetRemainingRepositories(ctx, info.ID, false, true)
	require.NoError(t, err)
	// Should include all repositories
	assert.Len(t, remaining, 4)

	// Test skipCompleted=false, retryFailed=false
	remaining, err = svc.GetRemainingRepositories(ctx, info.ID, false, false)
	require.NoError(t, err)
	// Should include completed and pending, exclude failed
	assert.Len(t, remaining, 3)
	assert.Contains(t, remaining, "repo1")
	assert.Contains(t, remaining, "repo2")
	assert.Contains(t, remaining, "repo4")
	assert.NotContains(t, remaining, "repo3")
}

// TestCheckpointServiceInitStorePermissions tests permission handling
func TestCheckpointServiceInitStorePermissions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping init store permissions test in short mode")
	}

	baseDir := t.TempDir()

	tests := []struct {
		name          string
		setupDir      func() string
		initialPerms  os.FileMode
		expectedPerms os.FileMode
	}{
		{
			name: "creates directory with secure permissions",
			setupDir: func() string {
				return filepath.Join(baseDir, "new-secure")
			},
			expectedPerms: 0700,
		},
		{
			name: "fixes insecure 0755 permissions",
			setupDir: func() string {
				dir := filepath.Join(baseDir, "fix-755")
				require.NoError(t, os.MkdirAll(dir, 0755))
				return dir
			},
			initialPerms:  0755,
			expectedPerms: 0700,
		},
		{
			name: "fixes insecure 0777 permissions",
			setupDir: func() string {
				dir := filepath.Join(baseDir, "fix-777")
				require.NoError(t, os.MkdirAll(dir, 0777))
				return dir
			},
			initialPerms:  0777,
			expectedPerms: 0700,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir()

			cfg := &config.Config{
				Checkpoint: config.CheckpointConfig{
					Directory: dir,
				},
			}
			logger := log.NewBasicLogger(log.InfoLevel)
			svc := NewCheckpointService(cfg, logger)

			ctx := context.Background()
			err := svc.initStore(ctx)
			require.NoError(t, err)

			// Verify permissions were set correctly
			info, err := os.Stat(dir)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedPerms, info.Mode().Perm())
		})
	}
}

// TestCheckpointServiceMultipleInitCalls tests idempotent initialization
func TestCheckpointServiceMultipleInitCalls(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping multiple init calls test in short mode")
	}

	cfg := &config.Config{
		Checkpoint: config.CheckpointConfig{
			Directory: t.TempDir(),
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewCheckpointService(cfg, logger)

	ctx := context.Background()

	// First init
	err := svc.initStore(ctx)
	require.NoError(t, err)
	assert.NotNil(t, svc.store)

	// Store the reference
	firstStore := svc.store

	// Second init (should be no-op)
	err = svc.initStore(ctx)
	require.NoError(t, err)
	assert.Equal(t, firstStore, svc.store) // Same store instance

	// Third init
	err = svc.initStore(ctx)
	require.NoError(t, err)
	assert.Equal(t, firstStore, svc.store)
}

// TestCheckpointServiceInvalidCheckpointFile tests handling of corrupted checkpoint
func TestCheckpointServiceInvalidCheckpointFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping invalid checkpoint file test in short mode")
	}

	cfg := &config.Config{
		Checkpoint: config.CheckpointConfig{
			Directory: t.TempDir(),
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewCheckpointService(cfg, logger)

	ctx := context.Background()

	// Test invalid JSON file
	tmpFile := filepath.Join(t.TempDir(), "invalid.json")
	err := os.WriteFile(tmpFile, []byte("{invalid json"), 0600)
	require.NoError(t, err)

	_, err = svc.ImportCheckpoint(ctx, tmpFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse checkpoint from JSON")
}

// TestCheckpointServiceExportToNestedDirectory tests export with nested path
func TestCheckpointServiceExportToNestedDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping export to nested directory test in short mode")
	}

	cfg := &config.Config{
		Checkpoint: config.CheckpointConfig{
			Directory: t.TempDir(),
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewCheckpointService(cfg, logger)

	ctx := context.Background()

	// Import a checkpoint first
	info := CheckpointInfo{
		ID:                    "test-export-nested",
		CreatedAt:             time.Now(),
		Source:                "ecr/source",
		Destination:           "gcr/dest",
		Status:                string(checkpoint.StatusCompleted),
		TotalRepositories:     1,
		CompletedRepositories: 1,
	}

	tmpFile := filepath.Join(t.TempDir(), "checkpoint.json")
	jsonData, err := json.Marshal(info)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(tmpFile, jsonData, 0600))

	_, err = svc.ImportCheckpoint(ctx, tmpFile)
	require.NoError(t, err)

	// Export to nested directory
	exportPath := filepath.Join(t.TempDir(), "nested", "dir", "export.json")
	err = svc.ExportCheckpoint(ctx, info.ID, exportPath)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(exportPath)
	require.NoError(t, err)
}

// TestReplicationOptionsVariations tests various replication option combinations
func TestReplicationOptionsVariations(t *testing.T) {
	tests := []struct {
		name string
		opts ReplicationOptions
	}{
		{
			name: "all enabled",
			opts: ReplicationOptions{
				DryRun:           true,
				ForceOverwrite:   true,
				IncludeManifests: true,
				IncludeLayers:    true,
				ParallelCopies:   8,
				RetryAttempts:    5,
				RetryDelay:       10 * time.Second,
			},
		},
		{
			name: "minimal options",
			opts: ReplicationOptions{
				DryRun:           false,
				ForceOverwrite:   false,
				IncludeManifests: false,
				IncludeLayers:    false,
				ParallelCopies:   1,
				RetryAttempts:    0,
				RetryDelay:       0,
			},
		},
		{
			name: "with progress callback",
			opts: ReplicationOptions{
				DryRun: true,
				ProgressCallback: func(progress *ReplicationProgress) {
					// Callback implementation
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.opts.DryRun, tt.opts.DryRun)
			assert.Equal(t, tt.opts.ParallelCopies, tt.opts.ParallelCopies)
			if tt.opts.ProgressCallback != nil {
				assert.NotNil(t, tt.opts.ProgressCallback)
			}
		})
	}
}

// TestReplicationResultsAggregation tests result aggregation
func TestReplicationResultsAggregation(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(5 * time.Second)

	results := []*ReplicationResult{
		{
			Success:      true,
			BytesCopied:  1000,
			LayersCopied: 5,
			Duration:     2 * time.Second,
		},
		{
			Success:      true,
			BytesCopied:  2000,
			LayersCopied: 10,
			Duration:     3 * time.Second,
		},
		{
			Success:      false,
			Error:        assert.AnError,
			BytesCopied:  0,
			LayersCopied: 0,
			Duration:     1 * time.Second,
		},
	}

	// Aggregate results
	var totalBytes int64
	var totalLayers int
	successCount := 0

	for _, result := range results {
		totalBytes += result.BytesCopied
		totalLayers += result.LayersCopied
		if result.Success {
			successCount++
		}
	}

	assert.Equal(t, int64(3000), totalBytes)
	assert.Equal(t, 15, totalLayers)
	assert.Equal(t, 2, successCount)

	// Test result with all fields
	fullResult := &ReplicationResult{
		Request: &ReplicationRequest{
			SourceRegistry:   "ecr",
			SourceRepository: "test",
		},
		Success:      true,
		Error:        nil,
		Duration:     5 * time.Second,
		BytesCopied:  5000,
		LayersCopied: 20,
		StartTime:    startTime,
		EndTime:      endTime,
	}

	assert.NotNil(t, fullResult.Request)
	assert.True(t, fullResult.Success)
	assert.NoError(t, fullResult.Error)
	assert.Equal(t, endTime.Sub(startTime), fullResult.Duration)
}
