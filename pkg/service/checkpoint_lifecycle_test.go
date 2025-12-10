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

// TestCheckpointServiceInitStore tests checkpoint store initialization
func TestCheckpointServiceInitStore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping checkpoint store initialization test in short mode")
	}

	tests := []struct {
		name        string
		setupDir    func(t *testing.T) string
		expectError bool
	}{
		{
			name: "creates directory if not exists",
			setupDir: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "new-checkpoints")
			},
			expectError: false,
		},
		{
			name: "uses existing directory",
			setupDir: func(t *testing.T) string {
				dir := filepath.Join(t.TempDir(), "existing-checkpoints")
				require.NoError(t, os.MkdirAll(dir, 0700))
				return dir
			},
			expectError: false,
		},
		{
			name: "fixes insecure permissions",
			setupDir: func(t *testing.T) string {
				dir := filepath.Join(t.TempDir(), "insecure-checkpoints")
				require.NoError(t, os.MkdirAll(dir, 0755))
				return dir
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setupDir(t)
			cfg := &config.Config{
				Checkpoint: config.CheckpointConfig{
					Directory: dir,
				},
			}
			logger := log.NewBasicLogger(log.InfoLevel)
			svc := NewCheckpointService(cfg, logger)

			ctx := context.Background()
			err := svc.initStore(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, svc.store)

				// Verify directory exists with correct permissions
				info, statErr := os.Stat(dir)
				require.NoError(t, statErr)
				assert.Equal(t, os.FileMode(0700), info.Mode().Perm())
			}
		})
	}
}

// TestCheckpointServiceGetCheckpoint tests retrieving a checkpoint
func TestCheckpointServiceGetCheckpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping checkpoint get test in short mode")
	}

	cfg := &config.Config{
		Checkpoint: config.CheckpointConfig{
			Directory: t.TempDir(),
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewCheckpointService(cfg, logger)

	ctx := context.Background()

	// Test with empty ID
	_, err := svc.GetCheckpoint(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checkpoint ID is required")

	// Test with non-existent checkpoint
	_, err = svc.GetCheckpoint(ctx, "non-existent-id")
	assert.Error(t, err)
}

// TestCheckpointServiceDeleteCheckpoint tests deleting a checkpoint
func TestCheckpointServiceDeleteCheckpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping checkpoint delete test in short mode")
	}

	cfg := &config.Config{
		Checkpoint: config.CheckpointConfig{
			Directory: t.TempDir(),
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewCheckpointService(cfg, logger)

	ctx := context.Background()

	// Test with empty ID
	err := svc.DeleteCheckpoint(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checkpoint ID is required")

	// Test with non-existent checkpoint
	err = svc.DeleteCheckpoint(ctx, "non-existent-id")
	assert.Error(t, err)
}

// TestCheckpointServiceExportImport tests exporting and importing checkpoints
func TestCheckpointServiceExportImport(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping checkpoint export/import test in short mode")
	}

	cfg := &config.Config{
		Checkpoint: config.CheckpointConfig{
			Directory: t.TempDir(),
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewCheckpointService(cfg, logger)

	ctx := context.Background()

	t.Run("export with empty ID", func(t *testing.T) {
		err := svc.ExportCheckpoint(ctx, "", "/tmp/test.json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "checkpoint ID is required")
	})

	t.Run("export with empty path", func(t *testing.T) {
		err := svc.ExportCheckpoint(ctx, "test-id", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "export file path is required")
	})

	t.Run("import with empty path", func(t *testing.T) {
		_, err := svc.ImportCheckpoint(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "import file path is required")
	})

	t.Run("import non-existent file", func(t *testing.T) {
		_, err := svc.ImportCheckpoint(ctx, "/non/existent/file.json")
		assert.Error(t, err)
	})

	t.Run("import invalid JSON", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "invalid.json")
		require.NoError(t, os.WriteFile(tmpFile, []byte("invalid json"), 0600))

		_, err := svc.ImportCheckpoint(ctx, tmpFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse checkpoint from JSON")
	})

	t.Run("successful import", func(t *testing.T) {
		// Create a valid checkpoint JSON
		info := CheckpointInfo{
			ID:                    "test-checkpoint",
			CreatedAt:             time.Now(),
			Source:                "ecr/test-source",
			Destination:           "gcr/test-dest",
			Status:                "completed",
			TotalRepositories:     5,
			CompletedRepositories: 5,
			FailedRepositories:    0,
			TotalTagsCopied:       20,
			TotalTagsSkipped:      2,
			TotalErrors:           0,
			TotalBytesTransferred: 1024000,
			Repositories: []RepositoryInfo{
				{
					Name:        "repo1",
					Status:      "completed",
					TagsCopied:  10,
					TagsSkipped: 1,
					Errors:      0,
				},
			},
		}

		tmpFile := filepath.Join(t.TempDir(), "valid.json")
		jsonData, err := json.Marshal(info)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(tmpFile, jsonData, 0600))

		imported, err := svc.ImportCheckpoint(ctx, tmpFile)
		assert.NoError(t, err)
		assert.NotNil(t, imported)
		assert.Equal(t, info.ID, imported.ID)
		assert.Equal(t, info.Source, imported.Source)
		assert.Equal(t, info.Destination, imported.Destination)
	})
}

// TestCheckpointServiceVerifyCheckpoint tests checkpoint verification
func TestCheckpointServiceVerifyCheckpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping checkpoint verification test in short mode")
	}

	cfg := &config.Config{
		Checkpoint: config.CheckpointConfig{
			Directory: t.TempDir(),
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewCheckpointService(cfg, logger)

	ctx := context.Background()

	t.Run("empty ID", func(t *testing.T) {
		_, err := svc.VerifyCheckpoint(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "checkpoint ID is required")
	})

	t.Run("non-existent checkpoint", func(t *testing.T) {
		exists, err := svc.VerifyCheckpoint(ctx, "non-existent-id")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

// TestCheckpointServiceGetRemainingRepositories tests getting remaining repositories
func TestCheckpointServiceGetRemainingRepositories(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping get remaining repositories test in short mode")
	}

	cfg := &config.Config{
		Checkpoint: config.CheckpointConfig{
			Directory: t.TempDir(),
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewCheckpointService(cfg, logger)

	ctx := context.Background()

	t.Run("empty ID", func(t *testing.T) {
		_, err := svc.GetRemainingRepositories(ctx, "", true, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "checkpoint ID is required")
	})

	t.Run("non-existent checkpoint", func(t *testing.T) {
		_, err := svc.GetRemainingRepositories(ctx, "non-existent-id", true, true)
		assert.Error(t, err)
	})
}

// TestConvertCheckpointMethods tests conversion between checkpoint formats
func TestConvertCheckpointMethods(t *testing.T) {
	cfg := &config.Config{
		Checkpoint: config.CheckpointConfig{
			Directory: "/tmp/test-checkpoints",
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewCheckpointService(cfg, logger)

	t.Run("convertCheckpointToInfo with repositories", func(t *testing.T) {
		cp := &checkpoint.TreeCheckpoint{
			ID:             "test-123",
			StartTime:      time.Now(),
			LastUpdated:    time.Now(),
			SourceRegistry: "ecr",
			SourcePrefix:   "source-repo",
			DestRegistry:   "gcr",
			DestPrefix:     "dest-repo",
			Status:         checkpoint.StatusInProgress,
			Repositories: map[string]checkpoint.RepoStatus{
				"repo1": {
					Status:      checkpoint.StatusCompleted,
					SourceRepo:  "repo1",
					DestRepo:    "repo1",
					LastUpdated: time.Now(),
				},
				"repo2": {
					Status:      checkpoint.StatusFailed,
					SourceRepo:  "repo2",
					DestRepo:    "repo2",
					LastUpdated: time.Now(),
				},
			},
			CompletedRepositories: []string{"repo1"},
			Progress:              50.0,
		}

		info := svc.convertCheckpointToInfo(cp)
		assert.Equal(t, cp.ID, info.ID)
		assert.Equal(t, "ecr/source-repo", info.Source)
		assert.Equal(t, "gcr/dest-repo", info.Destination)
		assert.Equal(t, string(checkpoint.StatusInProgress), info.Status)
		assert.Equal(t, 2, info.TotalRepositories)
		assert.Equal(t, 1, info.CompletedRepositories)
		assert.Equal(t, 1, info.FailedRepositories)
		assert.Len(t, info.Repositories, 2)
	})

	t.Run("convertInfoToCheckpoint", func(t *testing.T) {
		info := CheckpointInfo{
			ID:                    "test-456",
			CreatedAt:             time.Now(),
			Source:                "ecr/source-repo",
			Destination:           "gcr/dest-repo",
			Status:                string(checkpoint.StatusCompleted),
			TotalRepositories:     3,
			CompletedRepositories: 3,
			FailedRepositories:    0,
			Repositories: []RepositoryInfo{
				{
					Name:        "repo1",
					Status:      string(checkpoint.StatusCompleted),
					TagsCopied:  10,
					TagsSkipped: 1,
					Errors:      0,
				},
			},
		}

		cp := svc.convertInfoToCheckpoint(info)
		assert.Equal(t, info.ID, cp.ID)
		assert.Equal(t, "ecr", cp.SourceRegistry)
		assert.Equal(t, "source-repo", cp.SourcePrefix)
		assert.Equal(t, "gcr", cp.DestRegistry)
		assert.Equal(t, "dest-repo", cp.DestPrefix)
		assert.Equal(t, checkpoint.Status(info.Status), cp.Status)
		assert.Len(t, cp.Repositories, 1)
		assert.Equal(t, 100.0, cp.Progress)
	})

	t.Run("round trip conversion", func(t *testing.T) {
		original := CheckpointInfo{
			ID:                    "round-trip-test",
			CreatedAt:             time.Now().Truncate(time.Second),
			Source:                "ecr/test-source",
			Destination:           "gcr/test-dest",
			Status:                "running",
			TotalRepositories:     5,
			CompletedRepositories: 3,
			FailedRepositories:    1,
			Repositories: []RepositoryInfo{
				{Name: "repo1", Status: "completed"},
				{Name: "repo2", Status: "failed"},
			},
		}

		cp := svc.convertInfoToCheckpoint(original)
		converted := svc.convertCheckpointToInfo(cp)

		assert.Equal(t, original.ID, converted.ID)
		assert.Equal(t, original.Source, converted.Source)
		assert.Equal(t, original.Destination, converted.Destination)
		assert.Equal(t, original.Status, converted.Status)
	})
}
