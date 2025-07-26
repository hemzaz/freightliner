package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"freightliner/pkg/client/ecr"
	"freightliner/pkg/client/gcr"
	freightlinerConfig "freightliner/pkg/config"
	"freightliner/pkg/copy"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/helper/util"
	"freightliner/pkg/helper/validation"
	"freightliner/pkg/replication"
	"freightliner/pkg/secrets"
	"freightliner/pkg/security/encryption"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/google/go-containerregistry/pkg/name"
	"google.golang.org/api/option"
)

// ReplicationService handles repository replication
type ReplicationService struct {
	cfg    *freightlinerConfig.Config
	logger *log.Logger
}

// NewReplicationService creates a new replication service
func NewReplicationService(cfg *freightlinerConfig.Config, logger *log.Logger) *ReplicationService {
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

// RepositoryReplicationOptions holds configuration for repository replication
type RepositoryReplicationOptions struct {
	// Source and destination registries
	Source      string
	Destination string

	// Specific tags to replicate (empty for all tags)
	Tags []string

	// Operation behavior
	DryRun         bool
	ForceOverwrite bool

	// Worker count for parallel operations
	WorkerCount int

	// Encryption settings
	EnableEncryption bool
}

// ReplicateRepository replicates a repository from source to destination
func (s *ReplicationService) ReplicateRepository(ctx context.Context, source, destination string) (*ReplicationResult, error) {
	// Create options from configuration
	options := RepositoryReplicationOptions{
		Source:           source,
		Destination:      destination,
		Tags:             s.cfg.Replicate.Tags,
		DryRun:           s.cfg.Replicate.DryRun,
		ForceOverwrite:   s.cfg.Replicate.Force,
		WorkerCount:      s.cfg.Workers.ReplicateWorkers,
		EnableEncryption: s.cfg.Encryption.Enabled,
	}

	// Parse source and destination
	sourceRegistry, sourceRepo, err := parseRegistryPath(options.Source)
	if err != nil {
		return nil, err
	}

	destRegistry, destRepo, err := parseRegistryPath(options.Destination)
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
	if initErr := s.initializeCredentials(ctx); initErr != nil {
		return nil, initErr
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
		creator, ok := destClient.(RepositoryCreator)
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

	// Auto-detect worker count if needed
	if options.WorkerCount == 0 && s.cfg.Workers.AutoDetect {
		options.WorkerCount = freightlinerConfig.GetOptimalWorkerCount()
		s.logger.Info("Auto-detected worker count", map[string]interface{}{
			"workers": options.WorkerCount,
		})
	}

	// Create copier
	copier := copy.NewCopier(s.logger)

	// Configure the copier if encryption is enabled
	if encManager != nil {
		copier = copier.WithEncryptionManager(encManager)
	}

	// If specific tags were provided, copy them individually
	if len(options.Tags) > 0 {
		var copyErrors []string
		tagsCopied := 0

		for _, tagName := range options.Tags {
			// Parse source and destination references
			srcRef, srcErr := name.NewTag(sourceRepository.GetName() + ":" + tagName)
			if srcErr != nil {
				copyErrors = append(copyErrors, fmt.Sprintf("invalid source tag %s: %s", tagName, srcErr))
				continue
			}

			destRef, destErr := name.NewTag(destRepository.GetName() + ":" + tagName)
			if destErr != nil {
				copyErrors = append(copyErrors, fmt.Sprintf("invalid destination tag %s: %s", tagName, destErr))
				continue
			}

			// Set copy options
			copyOpts := copy.CopyOptions{
				Source:         srcRef,
				Destination:    destRef,
				ForceOverwrite: options.ForceOverwrite,
				DryRun:         options.DryRun,
			}

			// Execute the copy
			result, copyErr := copier.CopyImage(ctx, srcRef, destRef, nil, nil, copyOpts)
			if copyErr != nil {
				copyErrors = append(copyErrors, fmt.Sprintf("failed to copy tag %s: %s", tagName, copyErr))
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

	// Get all tags from the source repository
	s.logger.Info("Listing all tags for full repository replication", map[string]interface{}{
		"source_repository":      sourceRepo,
		"destination_repository": destRepo,
	})

	sourceTags, err := sourceRepository.ListTags(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list tags in source repository")
	}

	if len(sourceTags) == 0 {
		s.logger.Info("No tags found in source repository", map[string]interface{}{
			"repository": sourceRepo,
		})
		return &ReplicationResult{
			TagsCopied:  0,
			TagsSkipped: 0,
			Errors:      0,
		}, nil
	}

	s.logger.Info("Starting full repository replication", map[string]interface{}{
		"source_repository":      sourceRepo,
		"destination_repository": destRepo,
		"tag_count":              len(sourceTags),
		"dry_run":                options.DryRun,
		"force_overwrite":        options.ForceOverwrite,
	})

	// Create a results collector for metrics
	results := util.NewResults()

	// Create a limited error group with the worker count as concurrency limit
	g := util.NewLimitedErrGroup(ctx, options.WorkerCount)

	// Process each tag
	for _, tag := range sourceTags {
		// Create local variable for tag to avoid closure issues
		currentTag := tag

		g.Go(func() error {
			// Create source and destination references
			srcRef, err := sourceRepository.GetImageReference(currentTag)
			if err != nil {
				s.logger.Error("Failed to get source image reference", err, map[string]interface{}{
					"tag": currentTag,
				})
				return err
			}

			destRef, err := destRepository.GetImageReference(currentTag)
			if err != nil {
				s.logger.Error("Failed to get destination image reference", err, map[string]interface{}{
					"tag": currentTag,
				})
				return err
			}

			// Check if tag already exists at destination and has same digest
			if !options.ForceOverwrite {
				skipTag, skipErr := s.shouldSkipTag(ctx, currentTag, sourceRepository, destRepository)
				if skipErr != nil {
					s.logger.Warn("Error checking if tag should be skipped, will attempt to copy", map[string]interface{}{
						"tag":   currentTag,
						"error": skipErr.Error(),
					})
				} else if skipTag {
					results.AddMetric("tagsSkipped", 1)
					return nil
				}
			}

			// Setup copy options
			copyOpts := copy.CopyOptions{
				Source:         srcRef,
				Destination:    destRef,
				ForceOverwrite: options.ForceOverwrite,
				DryRun:         options.DryRun,
			}

			// Get remote options
			srcOpts, err := sourceRepository.GetRemoteOptions()
			if err != nil {
				return errors.Wrap(err, "failed to get source remote options")
			}

			destOpts, err := destRepository.GetRemoteOptions()
			if err != nil {
				return errors.Wrap(err, "failed to get destination remote options")
			}

			// Execute copy
			result, err := copier.CopyImage(ctx, srcRef, destRef, srcOpts, destOpts, copyOpts)
			if err != nil {
				s.logger.Error("Failed to copy tag", err, map[string]interface{}{
					"tag": currentTag,
				})
				return err
			}

			// Update stats
			results.AddMetric("tagsCopied", 1)
			results.AddMetric("bytesTransferred", result.Stats.BytesTransferred)

			s.logger.Info("Tag copied successfully", map[string]interface{}{
				"tag":    currentTag,
				"bytes":  result.Stats.BytesTransferred,
				"layers": result.Stats.Layers,
			})

			return nil
		})
	}

	// Wait for all jobs to complete and collect any errors
	if err := g.Wait(); err != nil {
		// If there was an error, we still continue and return the results
		// but also log the error
		s.logger.Error("Error during tag replication", err, map[string]interface{}{
			"source_repository":      sourceRepo,
			"destination_repository": destRepo,
		})

		// Count this as an error
		results.AddMetric("errorCount", 1)
	}

	// Get metrics from results collector
	tagsCopied := int(results.GetMetric("tagsCopied"))
	tagsSkipped := int(results.GetMetric("tagsSkipped"))
	errorCount := int(results.GetMetric("errorCount"))
	bytesTransferred := results.GetMetric("bytesTransferred")

	s.logger.Info("Repository replication completed", map[string]interface{}{
		"source_repository":      sourceRepo,
		"destination_repository": destRepo,
		"tags_copied":            tagsCopied,
		"tags_skipped":           tagsSkipped,
		"errors":                 errorCount,
		"bytes_transferred":      bytesTransferred,
	})

	return &ReplicationResult{
		TagsCopied:       tagsCopied,
		TagsSkipped:      tagsSkipped,
		Errors:           errorCount,
		BytesTransferred: bytesTransferred,
	}, nil
}

// createWorkerPool creates a worker pool for parallel processing
func (s *ReplicationService) createWorkerPool(workerCount int) *replication.WorkerPool {
	if workerCount <= 0 {
		workerCount = 1
	}
	return replication.NewWorkerPool(workerCount, s.logger)
}

// shouldSkipTag checks if a tag should be skipped during replication
func (s *ReplicationService) shouldSkipTag(
	ctx context.Context,
	tag string,
	sourceRepo Repository,
	destRepo Repository,
) (bool, error) {
	// Get source manifest
	sourceManifest, err := sourceRepo.GetManifest(ctx, tag)
	if err != nil {
		return false, errors.Wrap(err, "failed to get source manifest")
	}

	// Try to get destination manifest
	destManifest, err := destRepo.GetManifest(ctx, tag)
	if err != nil {
		// If the destination manifest doesn't exist, we need to copy it
		return false, nil
	}

	// If both manifests have the same digest, we can skip copying
	if sourceManifest.Digest == destManifest.Digest {
		s.logger.Debug("Skipping tag, already exists with same digest", map[string]interface{}{
			"tag":           tag,
			"source_digest": sourceManifest.Digest,
			"dest_digest":   destManifest.Digest,
		})
		return true, nil
	}

	s.logger.Debug("Tag exists but has different digest, will re-copy", map[string]interface{}{
		"tag":           tag,
		"source_digest": sourceManifest.Digest,
		"dest_digest":   destManifest.Digest,
	})

	return false, nil
}

// Helper functions

// parseRegistryPath parses a registry path into registry type and repository name
func parseRegistryPath(path string) (string, string, error) {
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
func (s *ReplicationService) createRegistryClients(ctx context.Context, registries ...string) (map[string]RegistryClient, error) {
	registrySet := make(map[string]bool)
	for _, r := range registries {
		registrySet[r] = true
	}

	registryClients := make(map[string]RegistryClient)

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

// Local type alias for secrets.Provider to avoid breaking existing code
type SecretsProvider = secrets.Provider

// awsSecretsProvider implements the SecretsProvider interface using AWS Secrets Manager
// Note: This is a simplified implementation that should be replaced with pkg/secrets/aws
type awsSecretsProvider struct {
	client    *secretsmanager.Client
	logger    *log.Logger
	validator *validation.SecretsValidator
}

// GetSecret retrieves a secret from AWS Secrets Manager
func (p *awsSecretsProvider) GetSecret(ctx context.Context, name string) (string, error) {
	// Create the request to get the secret value
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(name),
	}

	// Call AWS Secrets Manager to get the secret value
	result, err := p.client.GetSecretValue(ctx, input)
	if err != nil {
		var resourceNotFound *types.ResourceNotFoundException
		if errors.As(err, &resourceNotFound) {
			return "", errors.NotFoundf("secret not found: %s", name)
		}
		return "", errors.Wrap(err, "failed to get secret from AWS Secrets Manager")
	}

	// The secret value can be either a SecretString or SecretBinary
	var secretValue string
	if result.SecretString != nil {
		secretValue = *result.SecretString
	} else if result.SecretBinary != nil {
		// For binary secrets, decode from base64
		decodedBinarySecret := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
		n, err := base64.StdEncoding.Decode(decodedBinarySecret, result.SecretBinary)
		if err != nil {
			return "", errors.Wrap(err, "failed to decode binary secret")
		}
		secretValue = string(decodedBinarySecret[:n])
	} else {
		return "", errors.InvalidInputf("secret value is empty for secret: %s", name)
	}

	return secretValue, nil
}

// GetJSONSecret retrieves a JSON-formatted secret and unmarshal it into the provided struct
func (p *awsSecretsProvider) GetJSONSecret(ctx context.Context, secretName string, v interface{}) error {
	secretValue, err := p.GetSecret(ctx, secretName)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(secretValue), v)
}

// PutSecret creates or updates a secret value
func (p *awsSecretsProvider) PutSecret(ctx context.Context, secretName, secretValue string) error {
	// Validate the secret operation using the comprehensive validation framework
	if err := p.validator.ValidateSecretOperation("create", "aws", secretName, secretValue, nil); err != nil {
		p.logger.Error("AWS secret validation failed", err, map[string]interface{}{
			"secret_name": secretName,
		})
		return err
	}

	// First try to create the secret
	createInput := &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(secretValue),
		Description:  aws.String("Secret managed by Freightliner"),
	}

	_, err := p.client.CreateSecret(ctx, createInput)
	if err != nil {
		// If secret already exists, update it instead
		var resourceExists *types.ResourceExistsException
		if errors.As(err, &resourceExists) {
			updateInput := &secretsmanager.UpdateSecretInput{
				SecretId:     aws.String(secretName),
				SecretString: aws.String(secretValue),
			}

			_, updateErr := p.client.UpdateSecret(ctx, updateInput)
			if updateErr != nil {
				p.logger.Error("failed to update AWS secret", err, map[string]interface{}{
					"secret_name": secretName,
					"error":       updateErr.Error(),
				})
				return errors.Wrap(updateErr, "failed to update secret in AWS Secrets Manager")
			}

			p.logger.Info("successfully updated AWS secret", map[string]interface{}{
				"secret_name": secretName,
			})
			return nil
		}

		// Handle permission errors
		var accessDenied *types.InvalidRequestException
		if errors.As(err, &accessDenied) {
			return errors.Forbiddenf("insufficient permissions to create secret %s: %v", secretName, err)
		}

		p.logger.Error("failed to create AWS secret", err, map[string]interface{}{
			"secret_name": secretName,
			"error":       err.Error(),
		})
		return errors.Wrap(err, "failed to create secret in AWS Secrets Manager")
	}

	p.logger.Info("successfully created AWS secret", map[string]interface{}{
		"secret_name": secretName,
	})
	return nil
}

// PutJSONSecret marshals a struct to JSON and stores it as a secret
func (p *awsSecretsProvider) PutJSONSecret(ctx context.Context, secretName string, v interface{}) error {
	// Validate the JSON secret using the comprehensive validation framework
	if err := p.validator.ValidateJSONSecret("aws", secretName, v, nil); err != nil {
		p.logger.Error("AWS JSON secret validation failed", err, map[string]interface{}{
			"secret_name": secretName,
			"error":       err.Error(),
		})
		return err
	}

	// Marshal the struct to JSON (validation already checked this will succeed)
	jsonData, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "failed to marshal struct to JSON")
	}

	// Store the JSON as a secret
	return p.PutSecret(ctx, secretName, string(jsonData))
}

// DeleteSecret deletes a secret
func (p *awsSecretsProvider) DeleteSecret(ctx context.Context, secretName string) error {
	// Validate the delete operation using the comprehensive validation framework
	if err := p.validator.ValidateSecretOperation("delete", "aws", secretName, "", nil); err != nil {
		p.logger.Error("AWS secret delete validation failed", err, map[string]interface{}{
			"secret_name": secretName,
			"error":       err.Error(),
		})
		return err
	}

	// Delete the secret with immediate deletion (no recovery window)
	deleteInput := &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(secretName),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	}

	_, err := p.client.DeleteSecret(ctx, deleteInput)
	if err != nil {
		var resourceNotFound *types.ResourceNotFoundException
		if errors.As(err, &resourceNotFound) {
			p.logger.Warn("attempted to delete non-existent AWS secret", map[string]interface{}{
				"secret_name": secretName,
			})
			return errors.NotFoundf("secret not found: %s", secretName)
		}

		// Handle permission errors
		var accessDenied *types.InvalidRequestException
		if errors.As(err, &accessDenied) {
			return errors.Forbiddenf("insufficient permissions to delete secret %s: %v", secretName, err)
		}

		p.logger.Error("failed to delete AWS secret", err, map[string]interface{}{
			"secret_name": secretName,
			"error":       err.Error(),
		})
		return errors.Wrap(err, "failed to delete secret from AWS Secrets Manager")
	}

	p.logger.Info("successfully deleted AWS secret", map[string]interface{}{
		"secret_name": secretName,
	})
	return nil
}

// gcpSecretsProvider implements the SecretsProvider interface using Google Secret Manager
// Note: This is a simplified implementation that should be replaced with pkg/secrets/gcp
type gcpSecretsProvider struct {
	client    *secretmanager.Client
	project   string
	logger    *log.Logger
	validator *validation.SecretsValidator
}

// GetSecret retrieves a secret from Google Secret Manager
func (p *gcpSecretsProvider) GetSecret(ctx context.Context, name string) (string, error) {
	// Construct the full resource name for the secret
	secretName := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", p.project, name)

	// Create the access request for the secret
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretName,
	}

	// Call Google Secret Manager to access the secret
	result, err := p.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", errors.Wrap(err, "failed to get secret from Google Secret Manager")
	}

	// Extract the payload data
	secretValue := string(result.Payload.Data)
	return secretValue, nil
}

// GetJSONSecret retrieves a JSON-formatted secret and unmarshal it into the provided struct
func (p *gcpSecretsProvider) GetJSONSecret(ctx context.Context, secretName string, v interface{}) error {
	secretValue, err := p.GetSecret(ctx, secretName)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(secretValue), v)
}

// PutSecret creates or updates a secret value
func (p *gcpSecretsProvider) PutSecret(ctx context.Context, secretName, secretValue string) error {
	// Validate the secret operation using the comprehensive validation framework
	if err := p.validator.ValidateSecretOperation("create", "gcp", secretName, secretValue, nil); err != nil {
		p.logger.Error("GCP secret validation failed", err, map[string]interface{}{
			"secret_name": secretName,
			"project":     p.project,
			"error":       err.Error(),
		})
		return err
	}

	// Check if secret exists first
	fullSecretName := fmt.Sprintf("projects/%s/secrets/%s", p.project, secretName)

	// Try to get secret metadata to see if it exists
	getReq := &secretmanagerpb.GetSecretRequest{
		Name: fullSecretName,
	}

	_, err := p.client.GetSecret(ctx, getReq)
	if err != nil {
		// Secret doesn't exist, create it
		createReq := &secretmanagerpb.CreateSecretRequest{
			Parent:   fmt.Sprintf("projects/%s", p.project),
			SecretId: secretName,
			Secret: &secretmanagerpb.Secret{
				Replication: &secretmanagerpb.Replication{
					Replication: &secretmanagerpb.Replication_Automatic_{
						Automatic: &secretmanagerpb.Replication_Automatic{},
					},
				},
				Labels: map[string]string{
					"managed-by": "freightliner",
				},
			},
		}

		_, createErr := p.client.CreateSecret(ctx, createReq)
		if createErr != nil {
			p.logger.Error("failed to create GCP secret", err, map[string]interface{}{
				"secret_name": secretName,
				"project":     p.project,
				"error":       createErr.Error(),
			})
			return errors.Wrap(createErr, "failed to create secret in Google Secret Manager")
		}
	}

	// Add secret version with the provided value
	addVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: fullSecretName,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(secretValue),
		},
	}

	_, err = p.client.AddSecretVersion(ctx, addVersionReq)
	if err != nil {
		p.logger.Error("failed to add secret version in GCP", err, map[string]interface{}{
			"secret_name": secretName,
			"project":     p.project,
			"error":       err.Error(),
		})
		return errors.Wrap(err, "failed to add secret version in Google Secret Manager")
	}

	p.logger.Info("successfully created/updated GCP secret", map[string]interface{}{
		"secret_name": secretName,
		"project":     p.project,
	})
	return nil
}

// PutJSONSecret marshals a struct to JSON and stores it as a secret
func (p *gcpSecretsProvider) PutJSONSecret(ctx context.Context, secretName string, v interface{}) error {
	// Validate the JSON secret using the comprehensive validation framework
	if err := p.validator.ValidateJSONSecret("gcp", secretName, v, nil); err != nil {
		p.logger.Error("GCP JSON secret validation failed", err, map[string]interface{}{
			"secret_name": secretName,
			"project":     p.project,
			"error":       err.Error(),
		})
		return err
	}

	// Marshal the struct to JSON (validation already checked this will succeed)
	jsonData, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "failed to marshal struct to JSON")
	}

	// Store the JSON as a secret
	return p.PutSecret(ctx, secretName, string(jsonData))
}

// DeleteSecret deletes a secret
func (p *gcpSecretsProvider) DeleteSecret(ctx context.Context, secretName string) error {
	// Validate the delete operation using the comprehensive validation framework
	if err := p.validator.ValidateSecretOperation("delete", "gcp", secretName, "", nil); err != nil {
		p.logger.Error("GCP secret delete validation failed", err, map[string]interface{}{
			"secret_name": secretName,
			"project":     p.project,
			"error":       err.Error(),
		})
		return err
	}

	// Construct the full resource name for the secret
	fullSecretName := fmt.Sprintf("projects/%s/secrets/%s", p.project, secretName)

	// Delete the secret
	deleteReq := &secretmanagerpb.DeleteSecretRequest{
		Name: fullSecretName,
	}

	err := p.client.DeleteSecret(ctx, deleteReq)
	if err != nil {
		// Check if the secret doesn't exist
		if strings.Contains(err.Error(), "not found") {
			p.logger.Warn("attempted to delete non-existent GCP secret", map[string]interface{}{
				"secret_name": secretName,
				"project":     p.project,
			})
			return errors.NotFoundf("secret not found: %s", secretName)
		}

		// Check for permission errors
		if strings.Contains(err.Error(), "permission denied") {
			return errors.Forbiddenf("insufficient permissions to delete secret %s: %v", secretName, err)
		}

		p.logger.Error("failed to delete GCP secret", err, map[string]interface{}{
			"secret_name": secretName,
			"project":     p.project,
			"error":       err.Error(),
		})
		return errors.Wrap(err, "failed to delete secret from Google Secret Manager")
	}

	p.logger.Info("successfully deleted GCP secret", map[string]interface{}{
		"secret_name": secretName,
		"project":     p.project,
	})
	return nil
}

// RegistryCredentials contains credentials for different registry types
type RegistryCredentials struct {
	ECR struct {
		AccessKey    string `json:"accessKey"`
		SecretKey    string `json:"secretKey"`
		SessionToken string `json:"sessionToken,omitempty"`
		Region       string `json:"region,omitempty"`
		AccountID    string `json:"accountID,omitempty"`
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
	// Configure AWS SDK options
	configOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
	}

	// Load the default AWS configuration
	cfg, err := awsconfig.LoadDefaultConfig(ctx, configOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load AWS config")
	}

	// Create the Secrets Manager client
	client := secretsmanager.NewFromConfig(cfg)

	// Return the AWS secrets provider implementation
	return &awsSecretsProvider{
		client:    client,
		logger:    s.logger,
		validator: validation.NewSecretsValidator(),
	}, nil
}

// createGCPSecretsProvider creates a Google Secret Manager provider
func (s *ReplicationService) createGCPSecretsProvider(ctx context.Context, project, credentialsFile string) (SecretsProvider, error) {
	// Configure client options
	var clientOpts []option.ClientOption
	if credentialsFile != "" {
		clientOpts = append(clientOpts, option.WithCredentialsFile(credentialsFile))
	}

	// Create the Secret Manager client
	client, err := secretmanager.NewClient(ctx, clientOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Secret Manager client")
	}

	// Return the GCP secrets provider implementation
	return &gcpSecretsProvider{
		client:    client,
		project:   project,
		logger:    s.logger,
		validator: validation.NewSecretsValidator(),
	}, nil
}

// loadRegistryCredentials loads registry credentials from a secrets provider
func (s *ReplicationService) loadRegistryCredentials(ctx context.Context, provider SecretsProvider) (RegistryCredentials, error) {
	// Get the registry credentials from the secrets provider
	registryCredsJson, err := provider.GetSecret(ctx, s.cfg.Secrets.RegistryCredsSecret)
	if err != nil {
		return RegistryCredentials{}, errors.Wrap(err, "failed to get registry credentials from secrets provider")
	}

	if registryCredsJson == "" {
		return RegistryCredentials{}, errors.InvalidInputf("empty registry credentials retrieved from secrets provider")
	}

	// Parse the credentials JSON
	var creds RegistryCredentials
	if err := json.Unmarshal([]byte(registryCredsJson), &creds); err != nil {
		return RegistryCredentials{}, errors.Wrap(err, "failed to unmarshal registry credentials")
	}

	// Log successful retrieval
	s.logger.Info("Successfully loaded registry credentials from secrets provider", map[string]interface{}{
		"secret_name": s.cfg.Secrets.RegistryCredsSecret,
	})

	return creds, nil
}

// applyRegistryCredentials applies registry credentials to the environment
func (s *ReplicationService) applyRegistryCredentials(creds RegistryCredentials) {
	// Apply AWS credentials if provided
	if creds.ECR.AccessKey != "" && creds.ECR.SecretKey != "" {
		if err := os.Setenv("AWS_ACCESS_KEY_ID", creds.ECR.AccessKey); err != nil {
			s.logger.Warn("Failed to set AWS_ACCESS_KEY_ID environment variable", map[string]interface{}{"error": err.Error()})
		}
		if err := os.Setenv("AWS_SECRET_ACCESS_KEY", creds.ECR.SecretKey); err != nil {
			s.logger.Warn("Failed to set AWS_SECRET_ACCESS_KEY environment variable", map[string]interface{}{"error": err.Error()})
		}

		if creds.ECR.SessionToken != "" {
			if err := os.Setenv("AWS_SESSION_TOKEN", creds.ECR.SessionToken); err != nil {
				s.logger.Warn("Failed to set AWS_SESSION_TOKEN environment variable", map[string]interface{}{"error": err.Error()})
			}
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
				_ = tmpFile.Close()
				_ = os.Remove(tmpFilePath) // Clean up when done
			}()

			// Decode and write credentials to file
			decoded, err := base64.StdEncoding.DecodeString(creds.GCR.Credentials)
			if err == nil {
				if _, err := tmpFile.Write(decoded); err == nil {
					if setEnvErr := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tmpFilePath); setEnvErr != nil {
						s.logger.Warn("Failed to set GOOGLE_APPLICATION_CREDENTIALS environment variable", map[string]interface{}{"error": setEnvErr.Error()})
					}
				}
			}
		}
	}
}

// loadEncryptionKeys loads encryption keys from a secrets provider
func (s *ReplicationService) loadEncryptionKeys(ctx context.Context, provider SecretsProvider) (EncryptionKeys, error) {
	// Get the encryption keys from the secrets provider
	encryptionKeysJson, err := provider.GetSecret(ctx, s.cfg.Secrets.EncryptionKeysSecret)
	if err != nil {
		return EncryptionKeys{}, errors.Wrap(err, "failed to get encryption keys from secrets provider")
	}

	if encryptionKeysJson == "" {
		return EncryptionKeys{}, errors.InvalidInputf("empty encryption keys retrieved from secrets provider")
	}

	// Parse the encryption keys JSON
	var keys EncryptionKeys
	if err := json.Unmarshal([]byte(encryptionKeysJson), &keys); err != nil {
		return EncryptionKeys{}, errors.Wrap(err, "failed to unmarshal encryption keys")
	}

	// Log successful retrieval
	s.logger.Info("Successfully loaded encryption keys from secrets provider", map[string]interface{}{
		"secret_name": s.cfg.Secrets.EncryptionKeysSecret,
	})

	return keys, nil
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
