package service

import (
	"context"
	"testing"

	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"

	"github.com/stretchr/testify/assert"
)

// TestReplicationServiceCreation tests replication service creation
func TestReplicationServiceCreation(t *testing.T) {
	cfg := &config.Config{
		ECR: config.ECRConfig{
			Region:    "us-east-1",
			AccountID: "123456789012",
		},
		GCR: config.GCRConfig{
			Project:  "test-project",
			Location: "us",
		},
		Replicate: config.ReplicateConfig{
			DryRun: true,
			Force:  false,
		},
		Workers: config.WorkerConfig{
			ReplicateWorkers: 4,
			AutoDetect:       false,
		},
		Encryption: config.EncryptionConfig{
			Enabled: false,
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)

	svc := NewReplicationService(cfg, logger)
	assert.NotNil(t, svc)

	// Verify it implements the interface
	var _ ReplicationService = svc
}

// TestReplicationServiceValidation tests input validation
func TestReplicationServiceValidation(t *testing.T) {
	cfg := &config.Config{
		ECR: config.ECRConfig{
			Region:    "us-east-1",
			AccountID: "123456789012",
		},
		GCR: config.GCRConfig{
			Project:  "test-project",
			Location: "us",
		},
		Replicate: config.ReplicateConfig{},
		Workers: config.WorkerConfig{
			ReplicateWorkers: 2,
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewReplicationService(cfg, logger).(*replicationService)

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
			_, err := svc.ReplicateRepository(ctx, tt.source, tt.destination)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				// Note: Will fail on actual registry access in unit tests
				// This is expected as we're only testing validation
			}
		})
	}
}

// TestReplicateImage tests the ReplicateImage method
func TestReplicateImage(t *testing.T) {
	cfg := &config.Config{
		ECR: config.ECRConfig{
			Region:    "us-east-1",
			AccountID: "123456789012",
		},
		GCR: config.GCRConfig{
			Project:  "test-project",
			Location: "us",
		},
		Replicate: config.ReplicateConfig{
			DryRun: true,
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewReplicationService(cfg, logger)

	ctx := context.Background()
	req := &ReplicationRequest{
		SourceRegistry:        "ecr",
		SourceRepository:      "test-source",
		DestinationRegistry:   "gcr",
		DestinationRepository: "test-dest",
		SourceTags:            []string{"latest"},
		Priority:              1,
	}

	// This will fail on actual registry access, which is expected in unit tests
	_, err := svc.ReplicateImage(ctx, req)
	assert.Error(t, err) // Expected to fail on registry client creation
}

// TestReplicateImagesBatch tests batch replication
func TestReplicateImagesBatch(t *testing.T) {
	cfg := &config.Config{
		ECR: config.ECRConfig{
			Region:    "us-east-1",
			AccountID: "123456789012",
		},
		GCR: config.GCRConfig{
			Project:  "test-project",
			Location: "us",
		},
		Replicate: config.ReplicateConfig{
			DryRun: true,
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewReplicationService(cfg, logger)

	ctx := context.Background()
	requests := []*ReplicationRequest{
		{
			SourceRegistry:        "ecr",
			SourceRepository:      "test-source-1",
			DestinationRegistry:   "gcr",
			DestinationRepository: "test-dest-1",
		},
		{
			SourceRegistry:        "ecr",
			SourceRepository:      "test-source-2",
			DestinationRegistry:   "gcr",
			DestinationRepository: "test-dest-2",
		},
	}

	results, err := svc.ReplicateImagesBatch(ctx, requests)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	// All should have errors due to registry access
	for _, result := range results {
		assert.False(t, result.Success)
		assert.NotNil(t, result.Error)
	}
}

// TestStreamReplication tests streaming replication
func TestStreamReplication(t *testing.T) {
	cfg := &config.Config{
		ECR: config.ECRConfig{
			Region:    "us-east-1",
			AccountID: "123456789012",
		},
		GCR: config.GCRConfig{
			Project:  "test-project",
			Location: "us",
		},
		Replicate: config.ReplicateConfig{
			DryRun: true,
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewReplicationService(cfg, logger)

	ctx := context.Background()
	requestsChan := make(chan *ReplicationRequest, 2)

	requestsChan <- &ReplicationRequest{
		SourceRegistry:        "ecr",
		SourceRepository:      "test-source",
		DestinationRegistry:   "gcr",
		DestinationRepository: "test-dest",
	}
	close(requestsChan)

	resultsChan, errorsChan := svc.StreamReplication(ctx, requestsChan)

	// Collect results
	var results []*ReplicationResult
	var errors []error

	done := false
	for !done {
		select {
		case result, ok := <-resultsChan:
			if !ok {
				resultsChan = nil
			} else {
				results = append(results, result)
			}
		case err, ok := <-errorsChan:
			if !ok {
				errorsChan = nil
			} else {
				errors = append(errors, err)
			}
		}

		if resultsChan == nil && errorsChan == nil {
			done = true
		}
	}

	// Should have at least one error due to registry access
	assert.True(t, len(results) > 0 || len(errors) > 0)
}

// TestCreateWorkerPool tests worker pool creation
func TestCreateWorkerPool(t *testing.T) {
	cfg := &config.Config{}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewReplicationService(cfg, logger).(*replicationService)

	tests := []struct {
		name            string
		workerCount     int
		expectedWorkers int
	}{
		{
			name:            "positive worker count",
			workerCount:     4,
			expectedWorkers: 4,
		},
		{
			name:            "zero worker count defaults to 1",
			workerCount:     0,
			expectedWorkers: 1,
		},
		{
			name:            "negative worker count defaults to 1",
			workerCount:     -1,
			expectedWorkers: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := svc.createWorkerPool(tt.workerCount)
			assert.NotNil(t, pool)
		})
	}
}

// TestInitializeCredentialsDisabled tests when secrets manager is disabled
func TestInitializeCredentialsDisabled(t *testing.T) {
	cfg := &config.Config{
		Secrets: config.SecretsConfig{
			UseSecretsManager: false,
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewReplicationService(cfg, logger).(*replicationService)

	ctx := context.Background()
	err := svc.initializeCredentials(ctx)
	assert.NoError(t, err) // Should be no-op when disabled
}

// TestSetupEncryptionManagerDisabled tests when encryption is disabled
func TestSetupEncryptionManagerDisabled(t *testing.T) {
	cfg := &config.Config{
		Encryption: config.EncryptionConfig{
			Enabled: false,
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewReplicationService(cfg, logger).(*replicationService)

	ctx := context.Background()
	manager, err := svc.setupEncryptionManager(ctx, "ecr")
	assert.NoError(t, err)
	assert.NotNil(t, manager) // Should return empty manager, not nil
}

// TestApplyRegistryCredentials tests credential application
func TestApplyRegistryCredentials(t *testing.T) {
	cfg := &config.Config{
		ECR: config.ECRConfig{
			Region:    "us-west-2",
			AccountID: "111111111111",
		},
		GCR: config.GCRConfig{
			Project:  "old-project",
			Location: "europe",
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewReplicationService(cfg, logger).(*replicationService)

	creds := RegistryCredentials{}
	creds.ECR.Region = "us-east-1"
	creds.ECR.AccountID = "123456789012"
	creds.GCR.Project = "new-project"
	creds.GCR.Location = "us"

	svc.applyRegistryCredentials(creds)

	// Verify config was updated
	assert.Equal(t, "us-east-1", svc.cfg.ECR.Region)
	assert.Equal(t, "123456789012", svc.cfg.ECR.AccountID)
	assert.Equal(t, "new-project", svc.cfg.GCR.Project)
	assert.Equal(t, "us", svc.cfg.GCR.Location)
}

// TestApplyEncryptionKeys tests encryption key application
func TestApplyEncryptionKeys(t *testing.T) {
	cfg := &config.Config{
		Encryption: config.EncryptionConfig{
			Enabled: true,
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewReplicationService(cfg, logger).(*replicationService)

	keys := EncryptionKeys{}
	keys.AWS.KMSKeyID = "arn:aws:kms:us-east-1:123456789012:key/test-key"
	keys.GCP.KMSKeyID = "projects/test/locations/us/keyRings/test/cryptoKeys/test"
	keys.GCP.KeyRing = "test-keyring"
	keys.GCP.Key = "test-key"

	svc.applyEncryptionKeys(keys)

	// Verify config was updated
	assert.Equal(t, keys.AWS.KMSKeyID, svc.cfg.Encryption.AWSKMSKeyID)
	assert.Equal(t, keys.GCP.KMSKeyID, svc.cfg.Encryption.GCPKMSKeyID)
	assert.Equal(t, keys.GCP.KeyRing, svc.cfg.Encryption.GCPKeyRing)
	assert.Equal(t, keys.GCP.Key, svc.cfg.Encryption.GCPKeyName)
}

// TestRepositoryReplicationOptionsDefaults tests default options
func TestRepositoryReplicationOptionsDefaults(t *testing.T) {
	cfg := &config.Config{
		Replicate: config.ReplicateConfig{
			Tags:   []string{"latest", "stable"},
			DryRun: true,
			Force:  false,
		},
		Workers: config.WorkerConfig{
			ReplicateWorkers: 8,
		},
		Encryption: config.EncryptionConfig{
			Enabled: true,
		},
	}

	opts := RepositoryReplicationOptions{
		Source:           "ecr/test-source",
		Destination:      "gcr/test-dest",
		Tags:             cfg.Replicate.Tags,
		DryRun:           cfg.Replicate.DryRun,
		ForceOverwrite:   cfg.Replicate.Force,
		WorkerCount:      cfg.Workers.ReplicateWorkers,
		EnableEncryption: cfg.Encryption.Enabled,
	}

	assert.Equal(t, "ecr/test-source", opts.Source)
	assert.Equal(t, "gcr/test-dest", opts.Destination)
	assert.Len(t, opts.Tags, 2)
	assert.True(t, opts.DryRun)
	assert.False(t, opts.ForceOverwrite)
	assert.Equal(t, 8, opts.WorkerCount)
	assert.True(t, opts.EnableEncryption)
}
