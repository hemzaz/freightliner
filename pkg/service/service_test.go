package service

import (
	"context"
	"testing"
	"time"

	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewCheckpointService tests checkpoint service creation
func TestNewCheckpointService(t *testing.T) {
	cfg := &config.Config{
		Checkpoint: config.CheckpointConfig{
			Directory: "/tmp/test-checkpoints",
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)

	svc := NewCheckpointService(cfg, logger)
	assert.NotNil(t, svc)
	assert.Equal(t, cfg, svc.cfg)
	assert.NotNil(t, svc.logger)
}

// TestCheckpointServiceListEmpty tests listing when no checkpoints exist
func TestCheckpointServiceListEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping checkpoint service test in short mode")
	}

	cfg := &config.Config{
		Checkpoint: config.CheckpointConfig{
			Directory: t.TempDir() + "/checkpoints",
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewCheckpointService(cfg, logger)

	ctx := context.Background()
	checkpoints, err := svc.ListCheckpoints(ctx)

	require.NoError(t, err)
	assert.Empty(t, checkpoints)
}

// TestCheckpointInfoConversion tests checkpoint info conversion
func TestCheckpointInfoConversion(t *testing.T) {
	cfg := &config.Config{
		Checkpoint: config.CheckpointConfig{
			Directory: "/tmp/test-checkpoints",
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewCheckpointService(cfg, logger)

	// Create test checkpoint info
	info := CheckpointInfo{
		ID:                    "test-123",
		CreatedAt:             time.Now(),
		Source:                "ecr/test-source",
		Destination:           "gcr/test-dest",
		Status:                "running",
		TotalRepositories:     10,
		CompletedRepositories: 5,
		FailedRepositories:    1,
		TotalTagsCopied:       50,
		TotalTagsSkipped:      5,
		TotalErrors:           2,
		TotalBytesTransferred: 1024000,
	}

	// Convert to checkpoint and back
	cp := svc.convertInfoToCheckpoint(info)
	assert.NotNil(t, cp)
	assert.Equal(t, info.ID, cp.ID)

	converted := svc.convertCheckpointToInfo(cp)
	assert.Equal(t, info.ID, converted.ID)
	assert.Equal(t, info.Source, converted.Source)
	assert.Equal(t, info.Destination, converted.Destination)
}

// TestSplitPath tests path splitting utility
func TestSplitPath(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		expectedRegistry string
		expectedPrefix   string
	}{
		{
			name:             "full path",
			path:             "ecr/my-repo",
			expectedRegistry: "ecr",
			expectedPrefix:   "my-repo",
		},
		{
			name:             "path with multiple slashes",
			path:             "gcr/org/repo",
			expectedRegistry: "gcr",
			expectedPrefix:   "org/repo",
		},
		{
			name:             "single component",
			path:             "registry",
			expectedRegistry: "registry",
			expectedPrefix:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry, prefix := splitPath(tt.path)
			assert.Equal(t, tt.expectedRegistry, registry)
			assert.Equal(t, tt.expectedPrefix, prefix)
		})
	}
}

// TestNewReplicationService tests replication service creation
func TestNewReplicationService(t *testing.T) {
	cfg := &config.Config{
		ECR: config.ECRConfig{
			Region:    "us-east-1",
			AccountID: "123456789012",
		},
		GCR: config.GCRConfig{
			Project:  "test-project",
			Location: "us",
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)

	svc := NewReplicationService(cfg, logger)
	assert.NotNil(t, svc)
}

// TestParseRegistryPath tests registry path parsing
func TestParseRegistryPath(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		wantRegistry string
		wantRepo     string
		wantErr      bool
	}{
		{
			name:         "valid ecr path",
			path:         "ecr/my-repository",
			wantRegistry: "ecr",
			wantRepo:     "my-repository",
			wantErr:      false,
		},
		{
			name:         "valid gcr path",
			path:         "gcr/project/repo",
			wantRegistry: "gcr",
			wantRepo:     "project/repo",
			wantErr:      false,
		},
		{
			name:         "invalid path - no separator",
			path:         "invalid",
			wantRegistry: "",
			wantRepo:     "",
			wantErr:      true,
		},
		{
			name:         "invalid path - empty",
			path:         "",
			wantRegistry: "",
			wantRepo:     "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry, repo, err := parseRegistryPath(tt.path)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRegistry, registry)
				assert.Equal(t, tt.wantRepo, repo)
			}
		})
	}
}

// TestIsValidRegistryType tests registry type validation
func TestIsValidRegistryType(t *testing.T) {
	tests := []struct {
		registry   string
		configRegs []config.RegistryConfig
		expected   bool
	}{
		{
			registry:   "ecr",
			configRegs: []config.RegistryConfig{},
			expected:   true,
		},
		{
			registry:   "gcr",
			configRegs: []config.RegistryConfig{},
			expected:   true,
		},
		{
			registry: "docker",
			configRegs: []config.RegistryConfig{
				{Name: "docker", Type: config.RegistryTypeGeneric},
			},
			expected: true,
		},
		{
			registry:   "docker",
			configRegs: []config.RegistryConfig{},
			expected:   false,
		},
		{
			registry:   "",
			configRegs: []config.RegistryConfig{},
			expected:   false,
		},
		{
			registry:   "ECR",
			configRegs: []config.RegistryConfig{},
			expected:   false, // case sensitive
		},
		{
			registry:   "GCR",
			configRegs: []config.RegistryConfig{},
			expected:   false,
		},
		{
			registry:   "invalid",
			configRegs: []config.RegistryConfig{},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.registry, func(t *testing.T) {
			cfg := &config.Config{
				Registries: config.RegistriesConfig{
					Registries: tt.configRegs,
				},
			}
			svc := &replicationService{
				cfg:    cfg,
				logger: log.NewBasicLogger(log.InfoLevel),
			}
			result := svc.isValidRegistryType(tt.registry)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestReplicationRequest tests replication request structure
func TestReplicationRequest(t *testing.T) {
	req := &ReplicationRequest{
		SourceRegistry:        "ecr",
		SourceRepository:      "source-repo",
		SourceTags:            []string{"latest", "v1.0"},
		DestinationRegistry:   "gcr",
		DestinationRepository: "dest-repo",
		DestinationTags:       []string{"latest", "v1.0"},
		Options: &ReplicationOptions{
			DryRun:           true,
			ForceOverwrite:   false,
			IncludeManifests: true,
			IncludeLayers:    true,
			ParallelCopies:   4,
			RetryAttempts:    3,
			RetryDelay:       5 * time.Second,
		},
		Priority: 1,
	}

	assert.Equal(t, "ecr", req.SourceRegistry)
	assert.Equal(t, "source-repo", req.SourceRepository)
	assert.Len(t, req.SourceTags, 2)
	assert.True(t, req.Options.DryRun)
	assert.Equal(t, 4, req.Options.ParallelCopies)
}

// TestReplicationResult tests replication result structure
func TestReplicationResult(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(5 * time.Second)

	result := &ReplicationResult{
		Request: &ReplicationRequest{
			SourceRegistry:   "ecr",
			SourceRepository: "test",
		},
		Success:      true,
		Error:        nil,
		Duration:     5 * time.Second,
		BytesCopied:  1024000,
		LayersCopied: 10,
		StartTime:    startTime,
		EndTime:      endTime,
	}

	assert.True(t, result.Success)
	assert.NoError(t, result.Error)
	assert.Equal(t, int64(1024000), result.BytesCopied)
	assert.Equal(t, 10, result.LayersCopied)
	assert.Equal(t, 5*time.Second, result.Duration)
}

// TestNewTreeReplicationService tests tree replication service creation
func TestNewTreeReplicationService(t *testing.T) {
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

// TestTreeReplicationOptions tests tree replication options
func TestTreeReplicationOptions(t *testing.T) {
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
	assert.True(t, opts.EnableCheckpoint)
	assert.True(t, opts.SkipCompleted)
	assert.True(t, opts.RetryFailed)
}

// TestTreeReplicationResult tests tree replication result structure
func TestTreeReplicationResult(t *testing.T) {
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
	assert.Equal(t, int64(1024000000), result.TotalBytesTransferred)
	assert.Equal(t, "checkpoint-abc", result.CheckpointID)
}

// TestDefaultTreeReplicatorCreationOptions tests default options
func TestDefaultTreeReplicatorCreationOptions(t *testing.T) {
	opts := DefaultTreeReplicatorCreationOptions()

	assert.Equal(t, 2, opts.WorkerCount)
	assert.Empty(t, opts.ExcludeRepos)
	assert.Empty(t, opts.ExcludeTags)
	assert.Empty(t, opts.IncludeTags)
	assert.False(t, opts.DryRun)
	assert.False(t, opts.Force)
	assert.False(t, opts.EnableCheckpoint)
	assert.Contains(t, opts.CheckpointDir, "checkpoints")
	assert.True(t, opts.SkipCompleted)
	assert.True(t, opts.RetryFailed)
}

// TestReplicationProgress tests replication progress structure
func TestReplicationProgress(t *testing.T) {
	req := &ReplicationRequest{
		SourceRegistry:   "ecr",
		SourceRepository: "test",
	}

	progress := &ReplicationProgress{
		Request:          req,
		Stage:            "copying layers",
		Completed:        50,
		Total:            100,
		BytesTransferred: 512000,
		TotalBytes:       1024000,
		CurrentImage:     "test:latest",
	}

	assert.Equal(t, req, progress.Request)
	assert.Equal(t, "copying layers", progress.Stage)
	assert.Equal(t, 50, progress.Completed)
	assert.Equal(t, 100, progress.Total)
	assert.Equal(t, int64(512000), progress.BytesTransferred)
	assert.Equal(t, int64(1024000), progress.TotalBytes)
	assert.Equal(t, "test:latest", progress.CurrentImage)
}

// TestHealthStatus tests health status types
func TestHealthStatus(t *testing.T) {
	tests := []struct {
		status   HealthStatus
		expected string
	}{
		{HealthyStatus, "healthy"},
		{DegradedStatus, "degraded"},
		{UnhealthyStatus, "unhealthy"},
		{UnknownStatus, "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

// TestServiceHealth tests service health structure
func TestServiceHealth(t *testing.T) {
	health := &ServiceHealth{
		Status:  HealthyStatus,
		Version: "1.0.0",
		Uptime:  24 * time.Hour,
		Dependencies: map[string]*DependencyHealth{
			"database": {
				Name:         "database",
				Status:       HealthyStatus,
				ResponseTime: 50 * time.Millisecond,
				LastChecked:  time.Now(),
			},
		},
		Checks: map[string]*HealthCheck{
			"api": {
				Name:      "api",
				Status:    HealthyStatus,
				Message:   "API is responsive",
				Duration:  10 * time.Millisecond,
				Timestamp: time.Now(),
			},
		},
		Timestamp: time.Now(),
	}

	assert.Equal(t, HealthyStatus, health.Status)
	assert.Equal(t, "1.0.0", health.Version)
	assert.Equal(t, 24*time.Hour, health.Uptime)
	assert.Len(t, health.Dependencies, 1)
	assert.Len(t, health.Checks, 1)

	dbHealth := health.Dependencies["database"]
	assert.Equal(t, "database", dbHealth.Name)
	assert.Equal(t, HealthyStatus, dbHealth.Status)
}

// TestDependencyHealth tests dependency health structure
func TestDependencyHealth(t *testing.T) {
	dep := &DependencyHealth{
		Name:         "redis",
		Status:       HealthyStatus,
		ResponseTime: 5 * time.Millisecond,
		ErrorMessage: "",
		LastChecked:  time.Now(),
	}

	assert.Equal(t, "redis", dep.Name)
	assert.Equal(t, HealthyStatus, dep.Status)
	assert.Equal(t, 5*time.Millisecond, dep.ResponseTime)
	assert.Empty(t, dep.ErrorMessage)
}

// TestHealthCheck tests health check structure
func TestHealthCheck(t *testing.T) {
	check := &HealthCheck{
		Name:      "storage",
		Status:    HealthyStatus,
		Message:   "Storage is accessible",
		Duration:  20 * time.Millisecond,
		Timestamp: time.Now(),
	}

	assert.Equal(t, "storage", check.Name)
	assert.Equal(t, HealthyStatus, check.Status)
	assert.Equal(t, "Storage is accessible", check.Message)
	assert.Equal(t, 20*time.Millisecond, check.Duration)
}

// TestServiceInfo tests service info structure
func TestServiceInfo(t *testing.T) {
	buildTime := time.Now()

	info := &ServiceInfo{
		Name:        "freightliner",
		Version:     "1.0.0",
		BuildTime:   buildTime,
		GitCommit:   "abc123",
		Environment: "production",
		Features:    []string{"replication", "checkpoints", "tree-sync"},
	}

	assert.Equal(t, "freightliner", info.Name)
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, buildTime, info.BuildTime)
	assert.Equal(t, "abc123", info.GitCommit)
	assert.Equal(t, "production", info.Environment)
	assert.Len(t, info.Features, 3)
}

// TestRegistryCredentials tests registry credentials structure
func TestRegistryCredentials(t *testing.T) {
	creds := RegistryCredentials{}
	creds.ECR.AccessKey = "test-access-key"
	creds.ECR.SecretKey = "test-secret-key"
	creds.ECR.Region = "us-east-1"
	creds.ECR.AccountID = "123456789012"

	creds.GCR.Project = "test-project"
	creds.GCR.Location = "us"
	creds.GCR.Credentials = "base64-encoded-creds"

	assert.Equal(t, "test-access-key", creds.ECR.AccessKey)
	assert.Equal(t, "test-secret-key", creds.ECR.SecretKey)
	assert.Equal(t, "us-east-1", creds.ECR.Region)
	assert.Equal(t, "123456789012", creds.ECR.AccountID)
	assert.Equal(t, "test-project", creds.GCR.Project)
	assert.Equal(t, "us", creds.GCR.Location)
}

// TestEncryptionKeys tests encryption keys structure
func TestEncryptionKeys(t *testing.T) {
	keys := EncryptionKeys{}
	keys.AWS.KMSKeyID = "arn:aws:kms:us-east-1:123456789012:key/test"
	keys.GCP.KMSKeyID = "projects/test/locations/us/keyRings/test/cryptoKeys/test"
	keys.GCP.KeyRing = "test-keyring"
	keys.GCP.Key = "test-key"

	assert.Contains(t, keys.AWS.KMSKeyID, "arn:aws:kms")
	assert.Contains(t, keys.GCP.KMSKeyID, "projects/test")
	assert.Equal(t, "test-keyring", keys.GCP.KeyRing)
	assert.Equal(t, "test-key", keys.GCP.Key)
}

// TestRepositoryReplicationOptions tests repository replication options
func TestRepositoryReplicationOptions(t *testing.T) {
	opts := RepositoryReplicationOptions{
		Source:           "ecr/source-repo",
		Destination:      "gcr/dest-repo",
		Tags:             []string{"latest", "v1.0", "stable"},
		DryRun:           true,
		ForceOverwrite:   false,
		WorkerCount:      8,
		EnableEncryption: true,
	}

	assert.Equal(t, "ecr/source-repo", opts.Source)
	assert.Equal(t, "gcr/dest-repo", opts.Destination)
	assert.Len(t, opts.Tags, 3)
	assert.True(t, opts.DryRun)
	assert.False(t, opts.ForceOverwrite)
	assert.Equal(t, 8, opts.WorkerCount)
	assert.True(t, opts.EnableEncryption)
}

// TestTreeReplicatorCreationOptions tests tree replicator options
func TestTreeReplicatorCreationOptions(t *testing.T) {
	opts := TreeReplicatorCreationOptions{
		WorkerCount:      4,
		ExcludeRepos:     []string{"old-repo", "test-repo"},
		ExcludeTags:      []string{"alpha", "beta", "rc"},
		IncludeTags:      []string{"latest", "stable", "prod"},
		DryRun:           false,
		Force:            true,
		EnableCheckpoint: true,
		CheckpointDir:    "/var/lib/checkpoints",
		ResumeID:         "resume-abc-123",
		SkipCompleted:    true,
		RetryFailed:      true,
	}

	assert.Equal(t, 4, opts.WorkerCount)
	assert.Len(t, opts.ExcludeRepos, 2)
	assert.Len(t, opts.ExcludeTags, 3)
	assert.Len(t, opts.IncludeTags, 3)
	assert.False(t, opts.DryRun)
	assert.True(t, opts.Force)
	assert.True(t, opts.EnableCheckpoint)
	assert.Equal(t, "/var/lib/checkpoints", opts.CheckpointDir)
	assert.Equal(t, "resume-abc-123", opts.ResumeID)
	assert.True(t, opts.SkipCompleted)
	assert.True(t, opts.RetryFailed)
}

// TestCheckpointInfoStructure tests checkpoint info structure
func TestCheckpointInfoStructure(t *testing.T) {
	now := time.Now()

	info := CheckpointInfo{
		ID:                    "checkpoint-123",
		CreatedAt:             now,
		Source:                "ecr/test-source",
		Destination:           "gcr/test-dest",
		Status:                "running",
		TotalRepositories:     100,
		CompletedRepositories: 75,
		FailedRepositories:    5,
		TotalTagsCopied:       1000,
		TotalTagsSkipped:      50,
		TotalErrors:           10,
		TotalBytesTransferred: 10240000000,
		Repositories: []RepositoryInfo{
			{
				Name:        "repo1",
				Status:      "completed",
				TagsCopied:  10,
				TagsSkipped: 1,
				Errors:      0,
			},
			{
				Name:        "repo2",
				Status:      "failed",
				TagsCopied:  5,
				TagsSkipped: 2,
				Errors:      3,
			},
		},
	}

	assert.Equal(t, "checkpoint-123", info.ID)
	assert.Equal(t, now, info.CreatedAt)
	assert.Equal(t, "ecr/test-source", info.Source)
	assert.Equal(t, "gcr/test-dest", info.Destination)
	assert.Equal(t, "running", info.Status)
	assert.Equal(t, 100, info.TotalRepositories)
	assert.Equal(t, 75, info.CompletedRepositories)
	assert.Equal(t, 5, info.FailedRepositories)
	assert.Len(t, info.Repositories, 2)
	assert.Equal(t, int64(10240000000), info.TotalBytesTransferred)
}

// TestRepositoryInfo tests repository info structure
func TestRepositoryInfo(t *testing.T) {
	repo := RepositoryInfo{
		Name:        "test-repository",
		Status:      "completed",
		TagsCopied:  25,
		TagsSkipped: 5,
		Errors:      2,
	}

	assert.Equal(t, "test-repository", repo.Name)
	assert.Equal(t, "completed", repo.Status)
	assert.Equal(t, 25, repo.TagsCopied)
	assert.Equal(t, 5, repo.TagsSkipped)
	assert.Equal(t, 2, repo.Errors)
}
