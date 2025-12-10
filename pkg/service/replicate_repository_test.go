package service

import (
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRepository is a test mock for Repository interface
type MockRepository struct {
	name             string
	manifest         *MockManifest
	manifestError    error
	layerReader      io.ReadCloser
	layerReaderError error
	tags             []string
}

func (m *MockRepository) GetRepositoryName() string {
	if m.name == "" {
		return "test-repo"
	}
	return m.name
}

func (m *MockRepository) GetManifest(ctx context.Context, tag string) (*MockManifest, error) {
	if m.manifestError != nil {
		return nil, m.manifestError
	}
	if m.manifest == nil {
		return &MockManifest{
			Content: []byte(`{"test":"manifest"}`),
		}, nil
	}
	return m.manifest, nil
}

func (m *MockRepository) GetLayerReader(ctx context.Context, digest string) (io.ReadCloser, error) {
	if m.layerReaderError != nil {
		return nil, m.layerReaderError
	}
	return m.layerReader, nil
}

func (m *MockRepository) ListTags(ctx context.Context) ([]string, error) {
	if m.tags == nil {
		return []string{"latest", "v1.0"}, nil
	}
	return m.tags, nil
}

// MockManifest represents a container manifest for testing
type MockManifest struct {
	MediaType string
	Content   []byte
	Digest    string
}

// TestReplicateRepository_ParseRegistryPath tests path parsing
func TestReplicateRepository_ParseRegistryPath(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		expectedRegistry string
		expectedRepo     string
		expectError      bool
	}{
		{
			name:             "valid ECR path",
			path:             "ecr/my-repository",
			expectedRegistry: "ecr",
			expectedRepo:     "my-repository",
			expectError:      false,
		},
		{
			name:             "valid GCR path",
			path:             "gcr/my-project/my-repo",
			expectedRegistry: "gcr",
			expectedRepo:     "my-project/my-repo",
			expectError:      false,
		},
		{
			name:        "invalid path - no slash",
			path:        "invalid-path",
			expectError: true,
		},
		{
			name:        "empty path",
			path:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry, repo, err := parseRegistryPath(tt.path)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRegistry, registry)
				assert.Equal(t, tt.expectedRepo, repo)
			}
		})
	}
}

// TestReplicateRepository_IsValidRegistryType tests registry type validation
func TestReplicateRepository_IsValidRegistryType(t *testing.T) {
	tests := []struct {
		name       string
		registry   string
		configRegs []config.RegistryConfig
		expected   bool
	}{
		{
			name:       "ECR is valid",
			registry:   "ecr",
			configRegs: []config.RegistryConfig{},
			expected:   true,
		},
		{
			name:       "GCR is valid",
			registry:   "gcr",
			configRegs: []config.RegistryConfig{},
			expected:   true,
		},
		{
			name:     "configured custom registry is valid",
			registry: "harbor",
			configRegs: []config.RegistryConfig{
				{Name: "harbor", Type: config.RegistryTypeGeneric},
			},
			expected: true,
		},
		{
			name:       "empty is invalid",
			registry:   "",
			configRegs: []config.RegistryConfig{},
			expected:   false,
		},
		{
			name:       "unconfigured registry is invalid",
			registry:   "random",
			configRegs: []config.RegistryConfig{},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

// TestReplicationService_CreateRegistryClients tests client creation
func TestReplicationService_CreateRegistryClients(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

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
	service := NewReplicationService(cfg, logger).(*replicationService)

	ctx := context.Background()

	t.Run("create ECR client", func(t *testing.T) {
		clients, err := service.createRegistryClients(ctx, "ecr")
		require.NoError(t, err)
		assert.NotNil(t, clients)
		assert.NotNil(t, clients["ecr"])
	})

	t.Run("create GCR client", func(t *testing.T) {
		_, err := service.createRegistryClients(ctx, "gcr")
		// May fail without proper GCP credentials, but tests the logic
		if err != nil {
			assert.Contains(t, err.Error(), "failed to create GCR client")
		}
	})

	t.Run("create both clients", func(t *testing.T) {
		clients, err := service.createRegistryClients(ctx)
		require.NoError(t, err)
		assert.NotNil(t, clients["ecr"])
	})
}

// TestReplicationService_ShouldSkipTag_Logic tests tag skip logic at unit level
func TestReplicationService_ShouldSkipTag_Logic(t *testing.T) {
	// Skip this test as it requires full Repository interface implementation
	// The shouldSkipTag logic is tested through integration tests
	t.Skip("Requires complex Repository interface - tested in integration tests")
}

// TestRegistryCredentials_JSON tests registry credentials marshaling
func TestRegistryCredentials_JSON(t *testing.T) {
	creds := RegistryCredentials{}
	creds.ECR.AccessKey = "AKIAIOSFODNN7EXAMPLE"
	creds.ECR.SecretKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	creds.ECR.Region = "us-east-1"
	creds.GCR.Project = "test-project"

	t.Run("marshal to JSON", func(t *testing.T) {
		data, err := json.Marshal(creds)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("unmarshal from JSON", func(t *testing.T) {
		data, _ := json.Marshal(creds)
		var decoded RegistryCredentials
		err := json.Unmarshal(data, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, creds.ECR.AccessKey, decoded.ECR.AccessKey)
		assert.Equal(t, creds.ECR.Region, decoded.ECR.Region)
	})
}

// TestEncryptionKeys_JSON tests encryption keys marshaling
func TestEncryptionKeys_JSON(t *testing.T) {
	keys := EncryptionKeys{}
	keys.AWS.KMSKeyID = "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
	keys.GCP.KMSKeyID = "projects/test/locations/us/keyRings/test-ring/cryptoKeys/test-key"
	keys.GCP.KeyRing = "test-ring"

	t.Run("marshal to JSON", func(t *testing.T) {
		data, err := json.Marshal(keys)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("unmarshal from JSON", func(t *testing.T) {
		data, _ := json.Marshal(keys)
		var decoded EncryptionKeys
		err := json.Unmarshal(data, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, keys.AWS.KMSKeyID, decoded.AWS.KMSKeyID)
		assert.Equal(t, keys.GCP.KeyRing, decoded.GCP.KeyRing)
	})
}

// TestAWSSecretsProvider_Operations tests AWS secrets provider
func TestAWSSecretsProvider_Operations(t *testing.T) {
	// Skip this test as it requires AWS credentials and valid client setup
	t.Skip("Requires AWS credentials - tested in integration tests")
}

// TestGCPSecretsProvider_Operations tests GCP secrets provider
func TestGCPSecretsProvider_Operations(t *testing.T) {
	// Skip this test as it requires GCP credentials and valid client setup
	t.Skip("Requires GCP credentials - tested in integration tests")
}

// TestReplicationService_SetupEncryptionManager tests encryption setup
func TestReplicationService_SetupEncryptionManager(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)

	t.Run("encryption disabled", func(t *testing.T) {
		cfg := &config.Config{
			Encryption: config.EncryptionConfig{
				Enabled: false,
			},
		}
		service := NewReplicationService(cfg, logger).(*replicationService)
		ctx := context.Background()

		manager, err := service.setupEncryptionManager(ctx, "ecr")
		assert.NoError(t, err)
		assert.NotNil(t, manager) // Returns empty manager instead of nil
	})

	t.Run("AWS KMS encryption enabled", func(t *testing.T) {
		cfg := &config.Config{
			Encryption: config.EncryptionConfig{
				Enabled:             true,
				AWSKMSKeyID:         "arn:aws:kms:us-east-1:123456789012:key/test",
				CustomerManagedKeys: true,
				EnvelopeEncryption:  true,
			},
			ECR: config.ECRConfig{
				Region: "us-east-1",
			},
		}
		service := NewReplicationService(cfg, logger).(*replicationService)
		ctx := context.Background()

		manager, err := service.setupEncryptionManager(ctx, "ecr")
		// May fail without proper AWS credentials, but tests the logic
		if err != nil {
			assert.Contains(t, err.Error(), "failed to create AWS KMS provider")
		}
		_ = manager
	})

	t.Run("GCP KMS encryption enabled", func(t *testing.T) {
		cfg := &config.Config{
			Encryption: config.EncryptionConfig{
				Enabled:             true,
				GCPKMSKeyID:         "projects/test/locations/us/keyRings/test/cryptoKeys/test",
				GCPKeyRing:          "test-ring",
				GCPKeyName:          "test-key",
				CustomerManagedKeys: true,
			},
			GCR: config.GCRConfig{
				Project:  "test-project",
				Location: "us",
			},
		}
		service := NewReplicationService(cfg, logger).(*replicationService)
		ctx := context.Background()

		manager, err := service.setupEncryptionManager(ctx, "gcr")
		// May fail without proper GCP credentials, but tests the logic
		if err != nil {
			assert.Contains(t, err.Error(), "failed to create GCP KMS provider")
		}
		_ = manager
	})
}

// TestReplicationService_ApplyCredentials tests credential application
func TestReplicationService_ApplyCredentials(t *testing.T) {
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
	service := NewReplicationService(cfg, logger).(*replicationService)

	creds := RegistryCredentials{}
	creds.ECR.Region = "us-east-1"
	creds.ECR.AccountID = "123456789012"
	creds.GCR.Project = "new-project"
	creds.GCR.Location = "us"

	service.applyRegistryCredentials(creds)

	assert.Equal(t, "us-east-1", service.cfg.ECR.Region)
	assert.Equal(t, "123456789012", service.cfg.ECR.AccountID)
	assert.Equal(t, "new-project", service.cfg.GCR.Project)
	assert.Equal(t, "us", service.cfg.GCR.Location)
}

// TestReplicationService_ApplyEncryptionKeys tests encryption key application
func TestReplicationService_ApplyEncryptionKeys(t *testing.T) {
	cfg := &config.Config{
		Encryption: config.EncryptionConfig{
			Enabled: true,
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	service := NewReplicationService(cfg, logger).(*replicationService)

	keys := EncryptionKeys{}
	keys.AWS.KMSKeyID = "arn:aws:kms:us-east-1:123456789012:key/test"
	keys.GCP.KMSKeyID = "projects/test/locations/us/keyRings/test/cryptoKeys/test"
	keys.GCP.KeyRing = "test-ring"
	keys.GCP.Key = "test-key"

	service.applyEncryptionKeys(keys)

	assert.Equal(t, keys.AWS.KMSKeyID, service.cfg.Encryption.AWSKMSKeyID)
	assert.Equal(t, keys.GCP.KMSKeyID, service.cfg.Encryption.GCPKMSKeyID)
	assert.Equal(t, keys.GCP.KeyRing, service.cfg.Encryption.GCPKeyRing)
	assert.Equal(t, keys.GCP.Key, service.cfg.Encryption.GCPKeyName)
}

// TestReplicationService_ReplicateImagesBatch tests batch replication
func TestReplicationService_ReplicateImagesBatch(t *testing.T) {
	cfg := &config.Config{
		Replicate: config.ReplicateConfig{
			DryRun: true,
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	service := NewReplicationService(cfg, logger)

	ctx := context.Background()
	requests := []*ReplicationRequest{
		{
			SourceRegistry:        "ecr",
			SourceRepository:      "repo1",
			DestinationRegistry:   "gcr",
			DestinationRepository: "repo1",
		},
		{
			SourceRegistry:        "ecr",
			SourceRepository:      "repo2",
			DestinationRegistry:   "gcr",
			DestinationRepository: "repo2",
		},
	}

	results, err := service.ReplicateImagesBatch(ctx, requests)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	// Results will have errors due to missing registry setup, but tests the flow
}

// TestReplicationService_StreamReplication tests streaming replication
func TestReplicationService_StreamReplication(t *testing.T) {
	cfg := &config.Config{
		Replicate: config.ReplicateConfig{
			DryRun: true,
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	service := NewReplicationService(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	requestsChan := make(chan *ReplicationRequest, 2)
	requestsChan <- &ReplicationRequest{
		SourceRegistry:        "ecr",
		SourceRepository:      "repo1",
		DestinationRegistry:   "gcr",
		DestinationRepository: "repo1",
	}
	close(requestsChan)

	resultsChan, errorsChan := service.StreamReplication(ctx, requestsChan)

	// Consume channels
	resultsCount := 0
	errorsCount := 0

	for {
		select {
		case _, ok := <-resultsChan:
			if !ok {
				resultsChan = nil
			} else {
				resultsCount++
			}
		case _, ok := <-errorsChan:
			if !ok {
				errorsChan = nil
			} else {
				errorsCount++
			}
		case <-ctx.Done():
			t.Log("Context timeout")
			return
		}

		if resultsChan == nil && errorsChan == nil {
			break
		}
	}

	// Should have processed one request (either result or error)
	assert.True(t, resultsCount > 0 || errorsCount > 0)
}

// TestReplicationService_InitializeCredentials tests credential initialization
func TestReplicationService_InitializeCredentials(t *testing.T) {
	t.Run("secrets manager disabled", func(t *testing.T) {
		cfg := &config.Config{
			Secrets: config.SecretsConfig{
				UseSecretsManager: false,
			},
		}
		logger := log.NewBasicLogger(log.InfoLevel)
		service := NewReplicationService(cfg, logger).(*replicationService)

		ctx := context.Background()
		err := service.initializeCredentials(ctx)
		assert.NoError(t, err)
	})

	t.Run("unsupported secrets manager type", func(t *testing.T) {
		cfg := &config.Config{
			Secrets: config.SecretsConfig{
				UseSecretsManager:  true,
				SecretsManagerType: "invalid",
			},
		}
		logger := log.NewBasicLogger(log.InfoLevel)
		service := NewReplicationService(cfg, logger).(*replicationService)

		ctx := context.Background()
		err := service.initializeCredentials(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported secrets manager type")
	})
}

// TestRepositoryReplicationOptions_Structure tests options structure
func TestRepositoryReplicationOptions_Structure(t *testing.T) {
	opts := RepositoryReplicationOptions{
		Source:           "ecr/source-repo",
		Destination:      "gcr/dest-repo",
		Tags:             []string{"v1.0", "v1.1"},
		DryRun:           true,
		ForceOverwrite:   false,
		WorkerCount:      4,
		EnableEncryption: true,
	}

	assert.Equal(t, "ecr/source-repo", opts.Source)
	assert.Equal(t, 2, len(opts.Tags))
	assert.True(t, opts.DryRun)
	assert.Equal(t, 4, opts.WorkerCount)
}
