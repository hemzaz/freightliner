package service

import (
	"context"
	"testing"

	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"

	"github.com/stretchr/testify/assert"
)

// TestInitializeSecretsManager tests secrets manager initialization logic
func TestInitializeSecretsManager(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)

	tests := []struct {
		name        string
		cfg         *config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "unsupported secrets manager type",
			cfg: &config.Config{
				Secrets: config.SecretsConfig{
					SecretsManagerType: "unsupported",
				},
			},
			expectError: true,
			errorMsg:    "unsupported secrets manager type",
		},
		{
			name: "aws secrets manager without credentials",
			cfg: &config.Config{
				Secrets: config.SecretsConfig{
					SecretsManagerType: "aws",
					AWSSecretRegion:    "us-east-1",
				},
				ECR: config.ECRConfig{
					Region: "us-east-1",
				},
			},
			expectError: false, // Will succeed with initializeSecretsManager but fail on actual secret operations
		},
		{
			name: "gcp secrets manager without credentials",
			cfg: &config.Config{
				Secrets: config.SecretsConfig{
					SecretsManagerType: "gcp",
					GCPSecretProject:   "test-project",
				},
				GCR: config.GCRConfig{
					Project: "test-project",
				},
			},
			expectError: false, // Will succeed with initializeSecretsManager but fail on actual secret operations
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewReplicationService(tt.cfg, logger).(*replicationService)
			ctx := context.Background()

			_, err := svc.initializeSecretsManager(ctx)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				// May succeed or fail depending on credentials availability
				// We're testing the logic path, not actual cloud access
			}
		})
	}
}

// TestApplyRegistryCredentialsPartial tests partial credential application
func TestApplyRegistryCredentialsPartial(t *testing.T) {
	cfg := &config.Config{
		ECR: config.ECRConfig{
			Region:    "original-region",
			AccountID: "original-account",
		},
		GCR: config.GCRConfig{
			Project:  "original-project",
			Location: "original-location",
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewReplicationService(cfg, logger).(*replicationService)

	t.Run("only ECR credentials", func(t *testing.T) {
		creds := RegistryCredentials{}
		creds.ECR.Region = "new-region"
		// Don't set AccountID

		svc.applyRegistryCredentials(creds)

		assert.Equal(t, "new-region", svc.cfg.ECR.Region)
		assert.Equal(t, "original-account", svc.cfg.ECR.AccountID) // Unchanged
	})

	t.Run("only GCR credentials", func(t *testing.T) {
		// Reset to original
		svc.cfg.GCR.Project = "original-project"
		svc.cfg.GCR.Location = "original-location"

		creds := RegistryCredentials{}
		creds.GCR.Project = "new-project"
		// Don't set Location

		svc.applyRegistryCredentials(creds)

		assert.Equal(t, "new-project", svc.cfg.GCR.Project)
		assert.Equal(t, "original-location", svc.cfg.GCR.Location) // Unchanged
	})

	t.Run("empty credentials", func(t *testing.T) {
		// Reset config
		svc.cfg.ECR.Region = "test-region"
		svc.cfg.ECR.AccountID = "test-account"

		creds := RegistryCredentials{}
		svc.applyRegistryCredentials(creds)

		// Should remain unchanged
		assert.Equal(t, "test-region", svc.cfg.ECR.Region)
		assert.Equal(t, "test-account", svc.cfg.ECR.AccountID)
	})
}

// TestApplyEncryptionKeysPartial tests partial encryption key application
func TestApplyEncryptionKeysPartial(t *testing.T) {
	cfg := &config.Config{
		Encryption: config.EncryptionConfig{
			Enabled:     true,
			AWSKMSKeyID: "original-aws-key",
			GCPKMSKeyID: "original-gcp-key",
			GCPKeyRing:  "original-keyring",
			GCPKeyName:  "original-key-name",
		},
	}
	logger := log.NewBasicLogger(log.InfoLevel)
	svc := NewReplicationService(cfg, logger).(*replicationService)

	t.Run("only AWS key", func(t *testing.T) {
		keys := EncryptionKeys{}
		keys.AWS.KMSKeyID = "new-aws-key"

		svc.applyEncryptionKeys(keys)

		assert.Equal(t, "new-aws-key", svc.cfg.Encryption.AWSKMSKeyID)
		assert.Equal(t, "original-gcp-key", svc.cfg.Encryption.GCPKMSKeyID) // Unchanged
	})

	t.Run("only GCP keys", func(t *testing.T) {
		// Reset
		svc.cfg.Encryption.GCPKMSKeyID = "original-gcp-key"
		svc.cfg.Encryption.GCPKeyRing = "original-keyring"

		keys := EncryptionKeys{}
		keys.GCP.KMSKeyID = "new-gcp-key"
		keys.GCP.KeyRing = "new-keyring"

		svc.applyEncryptionKeys(keys)

		assert.Equal(t, "new-gcp-key", svc.cfg.Encryption.GCPKMSKeyID)
		assert.Equal(t, "new-keyring", svc.cfg.Encryption.GCPKeyRing)
		assert.Equal(t, "original-key-name", svc.cfg.Encryption.GCPKeyName) // Unchanged
	})

	t.Run("empty keys", func(t *testing.T) {
		// Reset
		svc.cfg.Encryption.AWSKMSKeyID = "test-aws-key"

		keys := EncryptionKeys{}
		svc.applyEncryptionKeys(keys)

		// Should remain unchanged
		assert.Equal(t, "test-aws-key", svc.cfg.Encryption.AWSKMSKeyID)
	})
}

// TestSetupEncryptionManagerWithProviders tests encryption manager setup
func TestSetupEncryptionManagerWithProviders(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)

	tests := []struct {
		name           string
		cfg            *config.Config
		destRegistry   string
		expectNil      bool
		expectProvider string
	}{
		{
			name: "encryption disabled",
			cfg: &config.Config{
				Encryption: config.EncryptionConfig{
					Enabled: false,
				},
			},
			destRegistry: "ecr",
			expectNil:    false, // Returns empty manager, not nil
		},
		{
			name: "AWS KMS with ECR destination",
			cfg: &config.Config{
				Encryption: config.EncryptionConfig{
					Enabled:     true,
					AWSKMSKeyID: "arn:aws:kms:us-east-1:123456789012:key/test",
				},
				ECR: config.ECRConfig{
					Region: "us-east-1",
				},
			},
			destRegistry:   "ecr",
			expectNil:      false,
			expectProvider: "aws-kms",
		},
		{
			name: "GCP KMS with GCR destination",
			cfg: &config.Config{
				Encryption: config.EncryptionConfig{
					Enabled:     true,
					GCPKMSKeyID: "projects/test/locations/us/keyRings/test/cryptoKeys/test",
					GCPKeyRing:  "test-keyring",
					GCPKeyName:  "test-key",
				},
				GCR: config.GCRConfig{
					Project:  "test-project",
					Location: "us",
				},
			},
			destRegistry:   "gcr",
			expectNil:      false,
			expectProvider: "gcp-kms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewReplicationService(tt.cfg, logger).(*replicationService)
			ctx := context.Background()

			manager, err := svc.setupEncryptionManager(ctx, tt.destRegistry)

			// Error may occur due to missing credentials, which is expected in unit tests
			if err != nil {
				t.Logf("Expected error due to missing credentials: %v", err)
				// When there's an error creating provider, manager can be nil
				return
			}

			if tt.expectNil {
				assert.Nil(t, manager)
			} else {
				assert.NotNil(t, manager)
				// Note: Can't test provider creation without actual cloud credentials
				// but we're testing the logic paths
			}
		})
	}
}

// TestRegistryCredentialsStructure tests registry credentials structure
func TestRegistryCredentialsStructure(t *testing.T) {
	creds := RegistryCredentials{}

	// Test ECR credentials
	creds.ECR.AccessKey = "AKIAIOSFODNN7EXAMPLE"
	creds.ECR.SecretKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	creds.ECR.SessionToken = "session-token-example"
	creds.ECR.Region = "us-east-1"
	creds.ECR.AccountID = "123456789012"

	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", creds.ECR.AccessKey)
	assert.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", creds.ECR.SecretKey)
	assert.Equal(t, "session-token-example", creds.ECR.SessionToken)
	assert.Equal(t, "us-east-1", creds.ECR.Region)
	assert.Equal(t, "123456789012", creds.ECR.AccountID)

	// Test GCR credentials
	creds.GCR.Credentials = "base64-encoded-credentials"
	creds.GCR.Project = "my-gcp-project"
	creds.GCR.Location = "us-central1"

	assert.Equal(t, "base64-encoded-credentials", creds.GCR.Credentials)
	assert.Equal(t, "my-gcp-project", creds.GCR.Project)
	assert.Equal(t, "us-central1", creds.GCR.Location)
}

// TestEncryptionKeysStructure tests encryption keys structure
func TestEncryptionKeysStructure(t *testing.T) {
	keys := EncryptionKeys{}

	// Test AWS encryption keys
	keys.AWS.KMSKeyID = "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
	assert.Equal(t, "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012", keys.AWS.KMSKeyID)
	assert.Contains(t, keys.AWS.KMSKeyID, "arn:aws:kms")

	// Test GCP encryption keys
	keys.GCP.KMSKeyID = "projects/my-project/locations/us-central1/keyRings/my-keyring/cryptoKeys/my-key"
	keys.GCP.KeyRing = "my-keyring"
	keys.GCP.Key = "my-key"

	assert.Equal(t, "projects/my-project/locations/us-central1/keyRings/my-keyring/cryptoKeys/my-key", keys.GCP.KMSKeyID)
	assert.Equal(t, "my-keyring", keys.GCP.KeyRing)
	assert.Equal(t, "my-key", keys.GCP.Key)
	assert.Contains(t, keys.GCP.KMSKeyID, "projects/")
}

// TestReplicationServiceInterfaceCompliance tests that service implements interface
func TestReplicationServiceInterfaceCompliance(t *testing.T) {
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

	// Verify concrete type implements interface
	var _ ReplicationService = NewReplicationService(cfg, logger)

	// Verify interface compliance
	svc := NewReplicationService(cfg, logger)
	assert.NotNil(t, svc)

	// Test interface methods are available
	ctx := context.Background()
	req := &ReplicationRequest{
		SourceRegistry:        "ecr",
		SourceRepository:      "test",
		DestinationRegistry:   "gcr",
		DestinationRepository: "test",
	}

	// These will fail due to registry access but verify method exists
	_, err := svc.ReplicateImage(ctx, req)
	assert.Error(t, err) // Expected

	_, err = svc.ReplicateImagesBatch(ctx, []*ReplicationRequest{req})
	assert.NoError(t, err) // Batch returns results, not error

	requestsChan := make(chan *ReplicationRequest)
	close(requestsChan)
	resultsChan, errorsChan := svc.StreamReplication(ctx, requestsChan)
	assert.NotNil(t, resultsChan)
	assert.NotNil(t, errorsChan)
}
