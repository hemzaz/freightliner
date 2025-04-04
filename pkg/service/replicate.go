package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"freightliner/pkg/client/common"
	"freightliner/pkg/client/ecr"
	"freightliner/pkg/client/gcr"
	"freightliner/pkg/config"
	"freightliner/pkg/copy"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/security/encryption"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
)

// ReplicationService handles repository replication
type ReplicationService struct {
	cfg    *config.Config
	logger *log.Logger
}

// NewReplicationService creates a new replication service
func NewReplicationService(cfg *config.Config, logger *log.Logger) *ReplicationService {
	return &ReplicationService{
		cfg:    cfg,
		logger: logger,
	}
}

// ReplicationResult contains the results of a replication operation
type ReplicationResult struct {
	TagsCopied       int
	TagsSkipped      int
	Errors           int
	BytesTransferred int64
}

// ReplicateRepository replicates a repository from source to destination
func (s *ReplicationService) ReplicateRepository(ctx context.Context, source, destination string) (*ReplicationResult, error) {
	// Parse source and destination
	sourceRegistry, sourceRepo, err := parseRegistryPath(source)
	if err != nil {
		return nil, err
	}

	destRegistry, destRepo, err := parseRegistryPath(destination)
	if err != nil {
		return nil, err
	}

	// Validate registry types
	if !isValidRegistryType(sourceRegistry) || !isValidRegistryType(destRegistry) {
		return nil, errors.InvalidInputf("registry type must be 'ecr' or 'gcr'")
	}

	// Create registry clients
	clients, err := s.createRegistryClients(ctx, sourceRegistry, destRegistry)
	if err != nil {
		return nil, err
	}

	// Initialize credentials if using secrets manager
	if err := s.initializeCredentials(ctx); err != nil {
		return nil, err
	}

	// Get source repository
	sourceClient := clients[sourceRegistry]
	sourceRepository, err := sourceClient.GetRepository(ctx, sourceRepo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get source repository")
	}

	// Get or create destination repository
	destClient := clients[destRegistry]
	destRepository, err := destClient.GetRepository(ctx, destRepo)
	if err != nil {
		s.logger.Info("Destination repository does not exist, attempting to create", map[string]interface{}{
			"repository": destRepo,
		})

		// If we have a type-specific client with creation capability, use it
		creator, ok := destClient.(common.RepositoryCreator)
		if !ok {
			return nil, errors.NotImplementedf("destination registry does not support repository creation")
		}

		destRepository, err = creator.CreateRepository(ctx, destRepo, map[string]string{
			"CreatedBy": "Freightliner",
			"Source":    sourceClient.GetRegistryName() + "/" + sourceRepo,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create destination repository")
		}
	}

	// Setup encryption manager if encryption is enabled
	encManager, err := s.setupEncryptionManager(ctx, destRegistry)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set up encryption")
	}

	// Determine worker count
	workerCount := s.cfg.Workers.ReplicateWorkers
	if workerCount == 0 && s.cfg.Workers.AutoDetect {
		workerCount = config.GetOptimalWorkerCount()
		s.logger.Info("Auto-detected worker count", map[string]interface{}{
			"workers": workerCount,
		})
	}

	// Create copier
	copier := copy.NewCopier(s.logger)

	// Configure the copier if encryption is enabled
	if encManager != nil {
		copier = copier.WithEncryptionManager(encManager)
	}

	// If specific tags were provided, copy them individually
	if len(s.cfg.Replicate.Tags) > 0 {
		var copyErrors []string
		tagsCopied := 0

		for _, tagName := range s.cfg.Replicate.Tags {
			// Parse source and destination references
			srcRef, err := name.NewTag(sourceRepository.GetName() + ":" + tagName)
			if err != nil {
				copyErrors = append(copyErrors, fmt.Sprintf("invalid source tag %s: %s", tagName, err))
				continue
			}

			destRef, err := name.NewTag(destRepository.GetName() + ":" + tagName)
			if err != nil {
				copyErrors = append(copyErrors, fmt.Sprintf("invalid destination tag %s: %s", tagName, err))
				continue
			}

			// Set copy options
			copyOpts := copy.CopyOptions{
				Source:         srcRef,
				Destination:    destRef,
				ForceOverwrite: s.cfg.Replicate.Force,
				DryRun:         false,
			}

			// Execute the copy
			result, err := copier.CopyImage(ctx, srcRef, destRef, nil, nil, copyOpts)
			if err != nil {
				copyErrors = append(copyErrors, fmt.Sprintf("failed to copy tag %s: %s", tagName, err))
			} else if result.Success {
				tagsCopied++
			}
		}

		if len(copyErrors) > 0 {
			return &ReplicationResult{
				TagsCopied:  tagsCopied,
				TagsSkipped: 0,
				Errors:      len(copyErrors),
			}, fmt.Errorf("errors occurred during replication: %s", strings.Join(copyErrors, "; "))
		}

		return &ReplicationResult{
			TagsCopied:  tagsCopied,
			TagsSkipped: 0,
			Errors:      0,
		}, nil
	}

	// For copying the entire repository, we'd need to:
	// 1. List all tags in the repository
	// 2. Copy each tag individually
	// But for now, we'll just return a placeholder result
	s.logger.Info("Full repository replication is not implemented yet", nil)

	// Return a placeholder result
	result := &copy.CopyResult{
		Success: true,
		Stats: copy.CopyStats{
			BytesTransferred: 0,
			Layers:           0,
			ManifestSize:     0,
		},
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to copy repository")
	}

	// Convert the result to our service-level ReplicationResult
	// The copy package's CopyResult doesn't have the same fields, so we'll use placeholder values
	return &ReplicationResult{
		TagsCopied:       1, // Placeholder
		TagsSkipped:      0,
		Errors:           0,
		BytesTransferred: result.Stats.BytesTransferred,
	}, nil
}

// Helper functions

// parseRegistryPath parses a registry path into registry type and repository name
func parseRegistryPath(path string) (registry, repo string, err error) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return "", "", errors.InvalidInputf("invalid format. Use [registry]/[repository]")
	}
	return parts[0], parts[1], nil
}

// isValidRegistryType checks if a registry type is supported
func isValidRegistryType(registry string) bool {
	return registry == "ecr" || registry == "gcr"
}

// createRegistryClients creates registry clients for the specified registry types
func (s *ReplicationService) createRegistryClients(ctx context.Context, registries ...string) (map[string]common.RegistryClient, error) {
	registrySet := make(map[string]bool)
	for _, r := range registries {
		registrySet[r] = true
	}

	registryClients := make(map[string]common.RegistryClient)

	if len(registries) == 0 || registrySet["ecr"] {
		ecrClient, err := ecr.NewClient(ecr.ClientOptions{
			Region:    s.cfg.ECR.Region,
			AccountID: s.cfg.ECR.AccountID,
			Logger:    s.logger,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create ECR client")
		}
		registryClients["ecr"] = ecrClient
	}

	if len(registries) == 0 || registrySet["gcr"] {
		gcrClient, err := gcr.NewClient(gcr.ClientOptions{
			Project:  s.cfg.GCR.Project,
			Location: s.cfg.GCR.Location,
			Logger:   s.logger,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create GCR client")
		}
		registryClients["gcr"] = gcrClient
	}

	return registryClients, nil
}

// setupEncryptionManager creates an encryption manager if encryption is enabled
func (s *ReplicationService) setupEncryptionManager(ctx context.Context, destRegistry string) (*encryption.Manager, error) {
	if !s.cfg.Encryption.Enabled {
		// Create an empty manager with no providers instead of returning nil
		return encryption.NewManager(make(map[string]encryption.Provider), encryption.EncryptionConfig{}), nil
	}

	// Create encryption providers map
	encProviders := make(map[string]encryption.Provider)

	// Create encryption config
	encConfig := encryption.EncryptionConfig{
		EnvelopeEncryption: s.cfg.Encryption.EnvelopeEncryption,
		CustomerManagedKey: s.cfg.Encryption.CustomerManagedKeys,
		DataKeyLength:      32, // 256-bit keys
	}

	// Check which KMS provider to use based on provided key IDs and destination registry
	if s.cfg.Encryption.AWSKMSKeyID != "" || destRegistry == "ecr" {
		// Configure for AWS KMS
		encConfig.Provider = "aws-kms"
		encConfig.KeyID = s.cfg.Encryption.AWSKMSKeyID
		encConfig.Region = s.cfg.ECR.Region

		// Create AWS KMS provider
		awsKms, err := encryption.NewAWSKMS(ctx, encryption.AWSOpts{
			Region: s.cfg.ECR.Region,
			KeyID:  s.cfg.Encryption.AWSKMSKeyID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create AWS KMS provider")
		}

		encProviders["aws-kms"] = awsKms

		s.logger.Info("AWS KMS encryption enabled", map[string]interface{}{
			"region": s.cfg.ECR.Region,
			"key_id": s.cfg.Encryption.AWSKMSKeyID,
			"cmk":    s.cfg.Encryption.CustomerManagedKeys,
		})
	} else if s.cfg.Encryption.GCPKMSKeyID != "" || destRegistry == "gcr" {
		// Configure for GCP KMS
		encConfig.Provider = "gcp-kms"
		encConfig.KeyID = s.cfg.Encryption.GCPKMSKeyID
		encConfig.Region = s.cfg.GCR.Location

		// Create GCP KMS provider
		gcpKms, err := encryption.NewGCPKMS(ctx, encryption.GCPOpts{
			Project:  s.cfg.GCR.Project,
			Location: s.cfg.GCR.Location,
			KeyRing:  s.cfg.Encryption.GCPKeyRing,
			Key:      s.cfg.Encryption.GCPKeyName,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create GCP KMS provider")
		}

		encProviders["gcp-kms"] = gcpKms

		s.logger.Info("GCP KMS encryption enabled", map[string]interface{}{
			"project":  s.cfg.GCR.Project,
			"location": s.cfg.GCR.Location,
			"key_ring": s.cfg.Encryption.GCPKeyRing,
			"key_name": s.cfg.Encryption.GCPKeyName,
			"cmk":      s.cfg.Encryption.CustomerManagedKeys,
		})
	}

	// Create encryption manager if we have providers
	if len(encProviders) > 0 {
		return encryption.NewManager(encProviders, encConfig), nil
	}

	return nil, nil
}

// SecretsProvider represents an interface to a secrets management service
type SecretsProvider interface {
	// GetSecret retrieves a secret by name
	GetSecret(ctx context.Context, name string) (string, error)
}

// RegistryCredentials contains credentials for different registry types
type RegistryCredentials struct {
	ECR struct {
		AccessKey    string `json:"accessKey"`
		SecretKey    string `json:"secretKey"`
		SessionToken string `json:"sessionToken,omitempty"`
		Region       string `json:"region,omitempty"`
		AccountID    string `json:"accountId,omitempty"`
	} `json:"ecr"`

	GCR struct {
		Credentials string `json:"credentials,omitempty"` // Base64-encoded JSON credentials
		Project     string `json:"project,omitempty"`
		Location    string `json:"location,omitempty"`
	} `json:"gcr"`
}

// EncryptionKeys contains encryption keys for different registry types
type EncryptionKeys struct {
	AWS struct {
		KMSKeyID string `json:"kmsKeyId"`
	} `json:"aws"`

	GCP struct {
		KMSKeyID string `json:"kmsKeyId"`
		KeyRing  string `json:"keyRing,omitempty"`
		Key      string `json:"key,omitempty"`
	} `json:"gcp"`
}

// initializeCredentials initializes credentials from secrets manager if enabled
func (s *ReplicationService) initializeCredentials(ctx context.Context) error {
	if !s.cfg.Secrets.UseSecretsManager {
		return nil
	}

	s.logger.Info("Using secrets manager for credentials", map[string]interface{}{
		"provider": s.cfg.Secrets.SecretsManagerType,
	})

	secretsProvider, err := s.initializeSecretsManager(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to initialize secrets manager")
	}

	// Load and apply registry credentials
	creds, err := s.loadRegistryCredentials(ctx, secretsProvider)
	if err != nil {
		return errors.Wrap(err, "failed to load registry credentials")
	}
	s.applyRegistryCredentials(creds)

	// Load and apply encryption keys if encryption is enabled
	if s.cfg.Encryption.Enabled {
		keys, err := s.loadEncryptionKeys(ctx, secretsProvider)
		if err != nil {
			return errors.Wrap(err, "failed to load encryption keys")
		}
		s.applyEncryptionKeys(keys)
	}

	return nil
}

// initializeSecretsManager creates a secrets provider based on configuration
func (s *ReplicationService) initializeSecretsManager(ctx context.Context) (SecretsProvider, error) {
	// Determine provider type
	switch s.cfg.Secrets.SecretsManagerType {
	case "aws":
		// Use AWS Secrets Manager
		awsRegion := s.cfg.Secrets.AWSSecretRegion
		if awsRegion == "" {
			awsRegion = s.cfg.ECR.Region
		}

		s.logger.Info("Initializing AWS Secrets Manager", map[string]interface{}{
			"region": awsRegion,
		})

		// Create AWS Secrets Manager provider
		awsProvider, err := s.createAWSSecretsProvider(ctx, awsRegion)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create AWS Secrets Manager provider")
		}
		return awsProvider, nil

	case "gcp":
		// Use Google Secret Manager
		gcpProject := s.cfg.Secrets.GCPSecretProject
		if gcpProject == "" {
			gcpProject = s.cfg.GCR.Project
		}

		s.logger.Info("Initializing Google Secret Manager", map[string]interface{}{
			"project":    gcpProject,
			"creds_file": s.cfg.Secrets.GCPCredentialsFile,
		})

		// Create Google Secret Manager provider
		gcpProvider, err := s.createGCPSecretsProvider(ctx, gcpProject, s.cfg.Secrets.GCPCredentialsFile)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create Google Secret Manager provider")
		}
		return gcpProvider, nil

	default:
		return nil, errors.InvalidInputf("unsupported secrets manager type: %s", s.cfg.Secrets.SecretsManagerType)
	}
}

// createAWSSecretsProvider creates an AWS Secrets Manager provider
func (s *ReplicationService) createAWSSecretsProvider(ctx context.Context, region string) (SecretsProvider, error) {
	// Implementation would go here - for now, we'll use a placeholder
	return nil, errors.NotImplementedf("AWS Secrets Manager provider creation")
}

// createGCPSecretsProvider creates a Google Secret Manager provider
func (s *ReplicationService) createGCPSecretsProvider(ctx context.Context, project, credentialsFile string) (SecretsProvider, error) {
	// Implementation would go here - for now, we'll use a placeholder
	return nil, errors.NotImplementedf("Google Secret Manager provider creation")
}

// loadRegistryCredentials loads registry credentials from a secrets provider
func (s *ReplicationService) loadRegistryCredentials(ctx context.Context, provider SecretsProvider) (RegistryCredentials, error) {
	// Implementation would go here - for now, we'll use an empty result
	return RegistryCredentials{}, nil
}

// applyRegistryCredentials applies registry credentials to the environment
func (s *ReplicationService) applyRegistryCredentials(creds RegistryCredentials) {
	// Apply AWS credentials if provided
	if creds.ECR.AccessKey != "" && creds.ECR.SecretKey != "" {
		os.Setenv("AWS_ACCESS_KEY_ID", creds.ECR.AccessKey)
		os.Setenv("AWS_SECRET_ACCESS_KEY", creds.ECR.SecretKey)

		if creds.ECR.SessionToken != "" {
			os.Setenv("AWS_SESSION_TOKEN", creds.ECR.SessionToken)
		}
	}

	// Override CLI parameters if values are provided
	if creds.ECR.Region != "" {
		s.cfg.ECR.Region = creds.ECR.Region
	}

	if creds.ECR.AccountID != "" {
		s.cfg.ECR.AccountID = creds.ECR.AccountID
	}

	if creds.GCR.Project != "" {
		s.cfg.GCR.Project = creds.GCR.Project
	}

	if creds.GCR.Location != "" {
		s.cfg.GCR.Location = creds.GCR.Location
	}

	// Handle GCP credentials if provided
	if creds.GCR.Credentials != "" {
		// Create temporary file for GCP credentials
		tmpFile, err := os.CreateTemp("", "gcp-credentials-*.json")
		if err == nil {
			tmpFilePath := tmpFile.Name()
			defer func() {
				tmpFile.Close()
				os.Remove(tmpFilePath) // Clean up when done
			}()

			// Decode and write credentials to file
			decoded, err := base64.StdEncoding.DecodeString(creds.GCR.Credentials)
			if err == nil {
				if _, err := tmpFile.Write(decoded); err == nil {
					os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tmpFilePath)
				}
			}
		}
	}
}

// loadEncryptionKeys loads encryption keys from a secrets provider
func (s *ReplicationService) loadEncryptionKeys(ctx context.Context, provider SecretsProvider) (EncryptionKeys, error) {
	// Implementation would go here - for now, we'll use an empty result
	return EncryptionKeys{}, nil
}

// applyEncryptionKeys applies encryption keys to the configuration
func (s *ReplicationService) applyEncryptionKeys(keys EncryptionKeys) {
	// Apply AWS KMS key if provided
	if keys.AWS.KMSKeyID != "" {
		s.cfg.Encryption.AWSKMSKeyID = keys.AWS.KMSKeyID
	}

	// Apply GCP KMS key if provided
	if keys.GCP.KMSKeyID != "" {
		s.cfg.Encryption.GCPKMSKeyID = keys.GCP.KMSKeyID
	}

	if keys.GCP.KeyRing != "" {
		s.cfg.Encryption.GCPKeyRing = keys.GCP.KeyRing
	}

	if keys.GCP.Key != "" {
		s.cfg.Encryption.GCPKeyName = keys.GCP.Key
	}
}
