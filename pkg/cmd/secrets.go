package cmd

import (
	"context"
	"encoding/base64"
	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/secrets"
	"os"
)

// Configuration options for secrets management
var (
	// useSecretsManager indicates whether to use a secrets manager
	useSecretsManager bool

	// secretsManagerType indicates which secrets manager to use (aws, gcp)
	secretsManagerType string

	// registryCredsSecret is the name of the secret containing registry credentials
	registryCredsSecret string

	// encryptionKeysSecret is the name of the secret containing encryption keys
	encryptionKeysSecret string

	// awsSecretRegion is the AWS region for secrets manager
	awsSecretRegion string

	// gcpSecretProject is the GCP project for secrets manager
	gcpSecretProject string

	// gcpCredentialsFile is the path to GCP credentials file
	gcpCredentialsFile string

	// awsKmsKeyID is the AWS KMS key ID
	awsKmsKeyID string

	// gcpKmsKeyID is the GCP KMS key ID
	gcpKmsKeyID string

	// main reference to the main CLI configuration
	main *Config
)

// Config represents the main CLI configuration
type Config struct {
	// ECR specific configuration
	ecrRegion    string
	ecrAccountID string

	// GCR specific configuration
	gcrProject  string
	gcrLocation string
}

// RegistryCredentials represents the structure of registry credentials stored in the secrets manager
type RegistryCredentials struct {
	ECR struct {
		AccessKey    string `json:"access_key"`
		SecretKey    string `json:"secret_key"`
		AccountID    string `json:"account_id"`
		Region       string `json:"region"`
		SessionToken string `json:"session_token,omitempty"`
	} `json:"ecr"`

	GCR struct {
		Project      string `json:"project"`
		Location     string `json:"location"`
		Credentials  string `json:"credentials,omitempty"` // Base64 encoded JSON credentials
		TokenSource  string `json:"token_source,omitempty"`
		ClientEmail  string `json:"client_email,omitempty"`
		PrivateKeyID string `json:"private_key_id,omitempty"`
		PrivateKey   string `json:"private_key,omitempty"`
	} `json:"gcr"`
}

// EncryptionKeys represents the structure of encryption keys stored in the secrets manager
type EncryptionKeys struct {
	AWS struct {
		KMSKeyID string `json:"kms_key_id"`
		Region   string `json:"region"`
	} `json:"aws"`

	GCP struct {
		KMSKeyID    string `json:"kms_key_id"`
		Project     string `json:"project"`
		Location    string `json:"location"`
		KeyRing     string `json:"key_ring"`
		Key         string `json:"key"`
		Credentials string `json:"credentials,omitempty"` // Base64 encoded JSON credentials
	} `json:"gcp"`
}

// initializeSecretsManager creates a secrets provider based on CLI flags
func initializeSecretsManager(ctx context.Context, logger *log.Logger) (secrets.Provider, error) {
	if ctx == nil {
		return nil, errors.InvalidInputf("context cannot be nil")
	}

	if logger == nil {
		return nil, errors.InvalidInputf("logger cannot be nil")
	}

	if !useSecretsManager {
		return nil, errors.InvalidInputf("secrets manager not enabled")
	}

	// Validate provider type
	var providerType secrets.ProviderType
	switch secretsManagerType {
	case "aws":
		providerType = secrets.AWSProvider
	case "gcp":
		providerType = secrets.GCPProvider
	default:
		return nil, errors.InvalidInputf("unsupported secrets manager type: %s", secretsManagerType)
	}

	// Set default regions/projects if not specified
	secretRegion := awsSecretRegion
	if secretRegion == "" {
		secretRegion = main.ecrRegion
	}

	secretProject := gcpSecretProject
	if secretProject == "" {
		secretProject = main.gcrProject
	}

	// Create the provider
	opts := secrets.ManagerOptions{
		Provider:           providerType,
		Logger:             logger,
		AWSRegion:          secretRegion,
		GCPProject:         secretProject,
		GCPCredentialsFile: gcpCredentialsFile,
	}

	return secrets.GetProvider(ctx, opts)
}

// loadRegistryCredentials loads registry credentials from secrets manager
func loadRegistryCredentials(ctx context.Context, provider secrets.Provider) (*RegistryCredentials, error) {
	if ctx == nil {
		return nil, errors.InvalidInputf("context cannot be nil")
	}

	if provider == nil {
		return nil, errors.InvalidInputf("secrets provider cannot be nil")
	}

	if registryCredsSecret == "" {
		return nil, errors.InvalidInputf("registry credentials secret name cannot be empty")
	}

	var creds RegistryCredentials
	if err := provider.GetJSONSecret(ctx, registryCredsSecret, &creds); err != nil {
		return nil, errors.Wrap(err, "failed to load registry credentials")
	}
	return &creds, nil
}

// loadEncryptionKeys loads encryption keys from secrets manager
func loadEncryptionKeys(ctx context.Context, provider secrets.Provider) (*EncryptionKeys, error) {
	if ctx == nil {
		return nil, errors.InvalidInputf("context cannot be nil")
	}

	if provider == nil {
		return nil, errors.InvalidInputf("secrets provider cannot be nil")
	}

	if encryptionKeysSecret == "" {
		return nil, errors.InvalidInputf("encryption keys secret name cannot be empty")
	}

	var keys EncryptionKeys
	if err := provider.GetJSONSecret(ctx, encryptionKeysSecret, &keys); err != nil {
		return nil, errors.Wrap(err, "failed to load encryption keys")
	}
	return &keys, nil
}

// applyRegistryCredentials applies loaded credentials to the current configuration
func applyRegistryCredentials(creds *RegistryCredentials) {
	// Set environment variables for AWS credentials
	if creds.ECR.AccessKey != "" && creds.ECR.SecretKey != "" {
		os.Setenv("AWS_ACCESS_KEY_ID", creds.ECR.AccessKey)
		os.Setenv("AWS_SECRET_ACCESS_KEY", creds.ECR.SecretKey)

		if creds.ECR.SessionToken != "" {
			os.Setenv("AWS_SESSION_TOKEN", creds.ECR.SessionToken)
		}
	}

	// Override CLI parameters if values are provided in secrets
	if creds.ECR.Region != "" {
		main.ecrRegion = creds.ECR.Region
	}

	if creds.ECR.AccountID != "" {
		main.ecrAccountID = creds.ECR.AccountID
	}

	if creds.GCR.Project != "" {
		main.gcrProject = creds.GCR.Project
	}

	if creds.GCR.Location != "" {
		main.gcrLocation = creds.GCR.Location
	}

	// Handle GCP credentials if provided
	if creds.GCR.Credentials != "" {
		// Create a temporary file for the credentials
		tmpFile, err := os.CreateTemp("", "gcp-credentials-*.json")
		if err == nil {
			// Ensure file is closed and deleted when function returns
			tmpFilePath := tmpFile.Name()
			defer func() {
				tmpFile.Close()
				os.Remove(tmpFilePath) // Securely delete the file when done
			}()

			// Decode and write credentials to the temp file
			decoded, err := base64.StdEncoding.DecodeString(creds.GCR.Credentials)
			if err == nil {
				if _, err := tmpFile.Write(decoded); err == nil {
					// Set environment variable to use this credentials file
					os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tmpFilePath)
				}
			}
		}
	}
}

// applyEncryptionKeys applies loaded encryption keys to the current configuration
func applyEncryptionKeys(keys *EncryptionKeys) {
	if keys.AWS.KMSKeyID != "" {
		awsKmsKeyID = keys.AWS.KMSKeyID
	}

	if keys.GCP.KMSKeyID != "" {
		gcpKmsKeyID = keys.GCP.KMSKeyID
	}
}
