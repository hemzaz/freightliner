package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"src/internal/log"
	"src/pkg/secrets"
)

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

// SigningKeys represents the structure of signing keys stored in the secrets manager
type SigningKeys struct {
	KeyPath string `json:"key_path"`
	KeyID   string `json:"key_id"`
	KeyData string `json:"key_data,omitempty"` // Base64 encoded key data
}

// initializeSecretsManager creates a secrets provider based on CLI flags
func initializeSecretsManager(ctx context.Context, logger *log.Logger) (secrets.Provider, error) {
	if !useSecretsManager {
		return nil, fmt.Errorf("secrets manager not enabled")
	}

	// Validate provider type
	var providerType secrets.ProviderType
	switch secretsManagerType {
	case "aws":
		providerType = secrets.AWSProvider
	case "gcp":
		providerType = secrets.GCPProvider
	default:
		return nil, fmt.Errorf("unsupported secrets manager type: %s", secretsManagerType)
	}

	// Set default regions/projects if not specified
	secretRegion := awsSecretRegion
	if secretRegion == "" {
		secretRegion = ecrRegion
	}

	secretProject := gcpSecretProject
	if secretProject == "" {
		secretProject = gcrProject
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
	var creds RegistryCredentials
	if err := provider.GetJSONSecret(ctx, registryCredsSecret, &creds); err != nil {
		return nil, fmt.Errorf("failed to load registry credentials: %w", err)
	}
	return &creds, nil
}

// loadEncryptionKeys loads encryption keys from secrets manager
func loadEncryptionKeys(ctx context.Context, provider secrets.Provider) (*EncryptionKeys, error) {
	var keys EncryptionKeys
	if err := provider.GetJSONSecret(ctx, encryptionKeysSecret, &keys); err != nil {
		return nil, fmt.Errorf("failed to load encryption keys: %w", err)
	}
	return &keys, nil
}

// loadSigningKeys loads signing keys from secrets manager
func loadSigningKeys(ctx context.Context, provider secrets.Provider) (*SigningKeys, error) {
	var keys SigningKeys
	if err := provider.GetJSONSecret(ctx, signingKeysSecret, &keys); err != nil {
		return nil, fmt.Errorf("failed to load signing keys: %w", err)
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
		ecrRegion = creds.ECR.Region
	}
	
	if creds.ECR.AccountID != "" {
		ecrAccountID = creds.ECR.AccountID
	}
	
	if creds.GCR.Project != "" {
		gcrProject = creds.GCR.Project
	}
	
	if creds.GCR.Location != "" {
		gcrLocation = creds.GCR.Location
	}
	
	// Handle GCP credentials if provided
	if creds.GCR.Credentials != "" {
		// Create a temporary file for the credentials
		tmpFile, err := os.CreateTemp("", "gcp-credentials-*.json")
		if err == nil {
			defer tmpFile.Close()
			
			// Decode and write credentials to the temp file
			decoded, err := base64.StdEncoding.DecodeString(creds.GCR.Credentials)
			if err == nil {
				if _, err := tmpFile.Write(decoded); err == nil {
					// Set environment variable to use this credentials file
					os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tmpFile.Name())
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

// applySigningKeys applies loaded signing keys to the current configuration
func applySigningKeys(keys *SigningKeys) {
	if keys.KeyID != "" {
		signKeyID = keys.KeyID
	}
	
	// If key data is provided, create a temporary file for it
	if keys.KeyData != "" {
		tmpFile, err := os.CreateTemp("", "signing-key-*")
		if err == nil {
			defer tmpFile.Close()
			
			// Decode and write key data to the temp file
			decoded, err := base64.StdEncoding.DecodeString(keys.KeyData)
			if err == nil {
				if _, err := tmpFile.Write(decoded); err == nil {
					signKeyPath = tmpFile.Name()
				}
			}
		}
	} else if keys.KeyPath != "" {
		signKeyPath = keys.KeyPath
	}
}