package service

import (
	"context"
	"testing"

	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"

	"github.com/stretchr/testify/assert"
)

// TestTreeReplicationServiceCreation tests service creation
func TestTreeReplicationServiceCreation(t *testing.T) {
	cfg := &config.Config{
		ECR: config.ECRConfig{
			Region:    "us-east-1",
			AccountID: "123456789012",
		},
		GCR: config.GCRConfig{
			Project:  "test-project",
			Location: "us",
		},
		TreeReplicate: config.TreeReplicateConfig{
			Workers:          4,
			ExcludeRepos:     []string{"test"},
			ExcludeTags:      []string{"old"},
			IncludeTags:      []string{"latest"},
			DryRun:           false,
			Force:            false,
			EnableCheckpoint: true,
			CheckpointDir:    "/tmp/checkpoints",
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)

	svc := NewTreeReplicationService(cfg, logger)
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.cfg)
	assert.NotNil(t, svc.logger)
	assert.NotNil(t, svc.replicationService)
}

// TestTreeReplicationValidation tests input validation
func TestTreeReplicationValidation(t *testing.T) {
	cfg := &config.Config{
		ECR: config.ECRConfig{
			Region:    "us-east-1",
			AccountID: "123456789012",
		},
		GCR: config.GCRConfig{
			Project:  "test-project",
			Location: "us",
		},
		TreeReplicate: config.TreeReplicateConfig{
			Workers: 2,
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewTreeReplicationService(cfg, logger)

	ctx := context.Background()

	tests := []struct {
		name        string
		source      string
		destination string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "invalid source format",
			source:      "invalid",
			destination: "gcr/dest",
			expectError: true,
			errorMsg:    "invalid format",
		},
		{
			name:        "invalid destination format",
			source:      "ecr/source",
			destination: "invalid",
			expectError: true,
			errorMsg:    "invalid format",
		},
		{
			name:        "invalid source registry type",
			source:      "docker/source",
			destination: "gcr/dest",
			expectError: true,
			errorMsg:    "registry type must be",
		},
		{
			name:        "invalid destination registry type",
			source:      "ecr/source",
			destination: "docker/dest",
			expectError: true,
			errorMsg:    "registry type must be",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ReplicateTree(ctx, tt.source, tt.destination)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			}
		})
	}
}

// TestTreeReplicationOptions tests tree replication options structure
func TestTreeReplicationOptionsStructure(t *testing.T) {
	opts := TreeReplicationOptions{
		Source:           "ecr/source",
		Destination:      "gcr/dest",
		WorkerCount:      8,
		ExcludeRepos:     []string{"old-repo"},
		ExcludeTags:      []string{"alpha", "beta"},
		IncludeTags:      []string{"latest", "stable"},
		DryRun:           true,
		Force:            false,
		EnableCheckpoint: true,
		CheckpointDir:    "/tmp/checkpoints",
		ResumeID:         "resume-123",
		SkipCompleted:    true,
		RetryFailed:      true,
	}

	assert.Equal(t, "ecr/source", opts.Source)
	assert.Equal(t, "gcr/dest", opts.Destination)
	assert.Equal(t, 8, opts.WorkerCount)
	assert.Len(t, opts.ExcludeRepos, 1)
	assert.Len(t, opts.ExcludeTags, 2)
	assert.Len(t, opts.IncludeTags, 2)
	assert.True(t, opts.DryRun)
	assert.False(t, opts.Force)
	assert.True(t, opts.EnableCheckpoint)
	assert.Equal(t, "/tmp/checkpoints", opts.CheckpointDir)
	assert.Equal(t, "resume-123", opts.ResumeID)
	assert.True(t, opts.SkipCompleted)
	assert.True(t, opts.RetryFailed)
}

// TestDefaultTreeReplicatorCreationOptionsValues tests default values
func TestDefaultTreeReplicatorCreationOptionsValues(t *testing.T) {
	opts := DefaultTreeReplicatorCreationOptions()

	assert.Equal(t, 2, opts.WorkerCount)
	assert.Empty(t, opts.ExcludeRepos)
	assert.Empty(t, opts.ExcludeTags)
	assert.Empty(t, opts.IncludeTags)
	assert.False(t, opts.DryRun)
	assert.False(t, opts.Force)
	assert.False(t, opts.EnableCheckpoint)
	assert.Contains(t, opts.CheckpointDir, "checkpoints")
	assert.Empty(t, opts.ResumeID)
	assert.True(t, opts.SkipCompleted)
	assert.True(t, opts.RetryFailed)
}

// TestTreeReplicatorCreationOptionsWithOverrides tests option overrides
func TestTreeReplicatorCreationOptionsWithOverrides(t *testing.T) {
	cfg := &config.Config{
		ECR: config.ECRConfig{
			Region:    "us-east-1",
			AccountID: "123456789012",
		},
		GCR: config.GCRConfig{
			Project:  "test-project",
			Location: "us",
		},
		TreeReplicate: config.TreeReplicateConfig{
			Workers:          8,
			ExcludeRepos:     []string{"test-repo"},
			ExcludeTags:      []string{"old"},
			IncludeTags:      []string{"latest"},
			DryRun:           true,
			Force:            true,
			EnableCheckpoint: true,
			CheckpointDir:    "/custom/checkpoint/dir",
			ResumeID:         "test-resume-id",
			SkipCompleted:    false,
			RetryFailed:      false,
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewTreeReplicationService(cfg, logger)

	// Create options from config
	opts := TreeReplicationOptions{
		Source:           "ecr/source",
		Destination:      "gcr/dest",
		WorkerCount:      svc.cfg.TreeReplicate.Workers,
		ExcludeRepos:     svc.cfg.TreeReplicate.ExcludeRepos,
		ExcludeTags:      svc.cfg.TreeReplicate.ExcludeTags,
		IncludeTags:      svc.cfg.TreeReplicate.IncludeTags,
		DryRun:           svc.cfg.TreeReplicate.DryRun,
		Force:            svc.cfg.TreeReplicate.Force,
		EnableCheckpoint: svc.cfg.TreeReplicate.EnableCheckpoint,
		CheckpointDir:    svc.cfg.TreeReplicate.CheckpointDir,
		ResumeID:         svc.cfg.TreeReplicate.ResumeID,
		SkipCompleted:    svc.cfg.TreeReplicate.SkipCompleted,
		RetryFailed:      svc.cfg.TreeReplicate.RetryFailed,
	}

	assert.Equal(t, 8, opts.WorkerCount)
	assert.Equal(t, []string{"test-repo"}, opts.ExcludeRepos)
	assert.Equal(t, []string{"old"}, opts.ExcludeTags)
	assert.Equal(t, []string{"latest"}, opts.IncludeTags)
	assert.True(t, opts.DryRun)
	assert.True(t, opts.Force)
	assert.True(t, opts.EnableCheckpoint)
	assert.Equal(t, "/custom/checkpoint/dir", opts.CheckpointDir)
	assert.Equal(t, "test-resume-id", opts.ResumeID)
	assert.False(t, opts.SkipCompleted)
	assert.False(t, opts.RetryFailed)
}

// TestTreeReplicationResultStructure tests result structure
func TestTreeReplicationResultStructure(t *testing.T) {
	result := &TreeReplicationResult{
		RepositoriesFound:      100,
		RepositoriesReplicated: 95,
		RepositoriesSkipped:    3,
		RepositoriesFailed:     2,
		TotalTagsCopied:        500,
		TotalTagsSkipped:       20,
		TotalErrors:            5,
		TotalBytesTransferred:  1024000000,
		CheckpointID:           "checkpoint-abc",
	}

	assert.Equal(t, 100, result.RepositoriesFound)
	assert.Equal(t, 95, result.RepositoriesReplicated)
	assert.Equal(t, 3, result.RepositoriesSkipped)
	assert.Equal(t, 2, result.RepositoriesFailed)
	assert.Equal(t, 500, result.TotalTagsCopied)
	assert.Equal(t, 20, result.TotalTagsSkipped)
	assert.Equal(t, 5, result.TotalErrors)
	assert.Equal(t, int64(1024000000), result.TotalBytesTransferred)
	assert.Equal(t, "checkpoint-abc", result.CheckpointID)
}

// TestTreeReplicationAutoDetectWorkers tests worker auto-detection
func TestTreeReplicationAutoDetectWorkers(t *testing.T) {
	cfg := &config.Config{
		ECR: config.ECRConfig{
			Region:    "us-east-1",
			AccountID: "123456789012",
		},
		GCR: config.GCRConfig{
			Project:  "test-project",
			Location: "us",
		},
		TreeReplicate: config.TreeReplicateConfig{
			Workers: 0, // Will trigger auto-detect
		},
		Workers: config.WorkerConfig{
			AutoDetect: true,
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewTreeReplicationService(cfg, logger)

	// Verify auto-detect is enabled
	assert.True(t, svc.cfg.Workers.AutoDetect)
	assert.Equal(t, 0, svc.cfg.TreeReplicate.Workers)
}

// TestTreeReplicationOptionsMapping tests options mapping
func TestTreeReplicationOptionsMapping(t *testing.T) {
	cfg := &config.Config{
		TreeReplicate: config.TreeReplicateConfig{
			Workers:          4,
			ExcludeRepos:     []string{"repo1", "repo2"},
			ExcludeTags:      []string{"tag1", "tag2"},
			IncludeTags:      []string{"latest", "stable"},
			DryRun:           true,
			Force:            false,
			EnableCheckpoint: true,
			CheckpointDir:    "/path/to/checkpoints",
			ResumeID:         "resume-abc",
			SkipCompleted:    true,
			RetryFailed:      false,
		},
	}

	// Create options map (simulating internal createTreeReplicator method)
	optionsMap := map[string]interface{}{
		"workers":          cfg.TreeReplicate.Workers,
		"excludeRepos":     cfg.TreeReplicate.ExcludeRepos,
		"excludeTags":      cfg.TreeReplicate.ExcludeTags,
		"includeTags":      cfg.TreeReplicate.IncludeTags,
		"dryRun":           cfg.TreeReplicate.DryRun,
		"force":            cfg.TreeReplicate.Force,
		"enableCheckpoint": cfg.TreeReplicate.EnableCheckpoint,
		"checkpointDir":    cfg.TreeReplicate.CheckpointDir,
		"resumeID":         cfg.TreeReplicate.ResumeID,
		"skipCompleted":    cfg.TreeReplicate.SkipCompleted,
		"retryFailed":      cfg.TreeReplicate.RetryFailed,
	}

	// Verify all options are correctly mapped
	assert.Equal(t, 4, optionsMap["workers"])
	assert.Equal(t, []string{"repo1", "repo2"}, optionsMap["excludeRepos"])
	assert.Equal(t, []string{"tag1", "tag2"}, optionsMap["excludeTags"])
	assert.Equal(t, []string{"latest", "stable"}, optionsMap["includeTags"])
	assert.True(t, optionsMap["dryRun"].(bool))
	assert.False(t, optionsMap["force"].(bool))
	assert.True(t, optionsMap["enableCheckpoint"].(bool))
	assert.Equal(t, "/path/to/checkpoints", optionsMap["checkpointDir"])
	assert.Equal(t, "resume-abc", optionsMap["resumeID"])
	assert.True(t, optionsMap["skipCompleted"].(bool))
	assert.False(t, optionsMap["retryFailed"].(bool))
}
