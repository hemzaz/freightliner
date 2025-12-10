package config

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"freightliner/pkg/helper/errors"

	"gopkg.in/yaml.v3"
)

// LoadFromFile loads configuration from a file or URL
func LoadFromFile(configPath string) (*Config, error) {
	// Set default configuration
	config := NewDefaultConfig()

	// If configPath is provided, load config from file or URL
	if configPath != "" {
		var data []byte
		var err error

		// Check if configPath is a URL
		if strings.HasPrefix(configPath, "http://") || strings.HasPrefix(configPath, "https://") {
			// Load from URL
			data, err = loadFromURL(configPath)
			if err != nil {
				return nil, errors.Wrap(err, "failed to load configuration from URL")
			}
		} else {
			// Load from file
			expandedPath := ExpandHomeDir(configPath)

			// Check if file exists
			if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
				return nil, errors.NotFoundf("configuration file not found: %s", expandedPath)
			}

			// Read file
			data, err = os.ReadFile(expandedPath)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read configuration file")
			}
		}

		// Unmarshal YAML
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, errors.Wrap(err, "failed to parse configuration")
		}
	}

	// Load from environment variables - these override file settings
	if err := loadFromEnv(config); err != nil {
		return nil, err
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// loadFromURL loads configuration data from an HTTP/HTTPS URL
func loadFromURL(url string) ([]byte, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Make GET request
	resp, err := client.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch URL")
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, errors.InvalidInputf("HTTP request failed with status: %d %s", resp.StatusCode, resp.Status)
	}

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	return data, nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) error {
	// Process string configuration values
	processStringEnvVars(config)

	// Process boolean configuration values
	processBooleanEnvVars(config)

	// Process integer configuration values
	processIntEnvVars(config)

	// Process slice configuration values
	processSliceEnvVars(config)

	// Process duration configuration values
	processDurationEnvVars(config)

	return nil
}

// processStringEnvVars loads string configuration from environment variables
func processStringEnvVars(config *Config) {
	// Map of environment variables to configuration fields
	envVars := map[string]*string{
		// General configuration
		"FREIGHTLINER_LOG_LEVEL": &config.LogLevel,

		// ECR configuration
		"FREIGHTLINER_ECR_REGION":     &config.ECR.Region,
		"FREIGHTLINER_ECR_ACCOUNT_ID": &config.ECR.AccountID,

		// GCR configuration
		"FREIGHTLINER_GCR_PROJECT":  &config.GCR.Project,
		"FREIGHTLINER_GCR_LOCATION": &config.GCR.Location,

		// Encryption configuration
		"FREIGHTLINER_AWS_KMS_KEY_ID": &config.Encryption.AWSKMSKeyID,
		"FREIGHTLINER_GCP_KMS_KEY_ID": &config.Encryption.GCPKMSKeyID,
		"FREIGHTLINER_GCP_KEY_RING":   &config.Encryption.GCPKeyRing,
		"FREIGHTLINER_GCP_KEY_NAME":   &config.Encryption.GCPKeyName,

		// Secrets configuration
		"FREIGHTLINER_SECRETS_MANAGER_TYPE":   &config.Secrets.SecretsManagerType,
		"FREIGHTLINER_AWS_SECRET_REGION":      &config.Secrets.AWSSecretRegion,
		"FREIGHTLINER_GCP_SECRET_PROJECT":     &config.Secrets.GCPSecretProject,
		"FREIGHTLINER_GCP_CREDENTIALS_FILE":   &config.Secrets.GCPCredentialsFile,
		"FREIGHTLINER_REGISTRY_CREDS_SECRET":  &config.Secrets.RegistryCredsSecret,
		"FREIGHTLINER_ENCRYPTION_KEYS_SECRET": &config.Secrets.EncryptionKeysSecret,

		// Server configuration
		"FREIGHTLINER_TLS_CERT_FILE":       &config.Server.TLSCertFile,
		"FREIGHTLINER_TLS_KEY_FILE":        &config.Server.TLSKeyFile,
		"FREIGHTLINER_API_KEY":             &config.Server.APIKey,
		"FREIGHTLINER_HEALTH_CHECK_PATH":   &config.Server.HealthCheckPath,
		"FREIGHTLINER_METRICS_PATH":        &config.Server.MetricsPath,
		"FREIGHTLINER_REPLICATE_PATH":      &config.Server.ReplicatePath,
		"FREIGHTLINER_TREE_REPLICATE_PATH": &config.Server.TreeReplicatePath,
		"FREIGHTLINER_STATUS_PATH":         &config.Server.StatusPath,

		// Checkpoint configuration
		"FREIGHTLINER_CHECKPOINT_DIRECTORY": &config.Checkpoint.Directory,
		"FREIGHTLINER_CHECKPOINT_ID":        &config.Checkpoint.ID,

		// Tree replication configuration
		"FREIGHTLINER_TREE_CHECKPOINT_DIR": &config.TreeReplicate.CheckpointDir,
		"FREIGHTLINER_TREE_RESUME_ID":      &config.TreeReplicate.ResumeID,
	}

	// Load environment variables
	for env, field := range envVars {
		if value, exists := os.LookupEnv(env); exists && value != "" {
			*field = value
		}
	}
}

// processBooleanEnvVars loads boolean configuration from environment variables
func processBooleanEnvVars(config *Config) {
	// Map of environment variables to configuration fields
	envVars := map[string]*bool{
		// Workers configuration
		"FREIGHTLINER_AUTO_DETECT_WORKERS": &config.Workers.AutoDetect,

		// Encryption configuration
		"FREIGHTLINER_ENCRYPTION_ENABLED":    &config.Encryption.Enabled,
		"FREIGHTLINER_CUSTOMER_MANAGED_KEYS": &config.Encryption.CustomerManagedKeys,
		"FREIGHTLINER_ENVELOPE_ENCRYPTION":   &config.Encryption.EnvelopeEncryption,

		// Secrets configuration
		"FREIGHTLINER_USE_SECRETS_MANAGER": &config.Secrets.UseSecretsManager,

		// Server configuration
		"FREIGHTLINER_TLS_ENABLED":  &config.Server.TLSEnabled,
		"FREIGHTLINER_API_KEY_AUTH": &config.Server.APIKeyAuth,

		// Tree replication configuration
		"FREIGHTLINER_TREE_DRY_RUN":           &config.TreeReplicate.DryRun,
		"FREIGHTLINER_TREE_FORCE":             &config.TreeReplicate.Force,
		"FREIGHTLINER_TREE_ENABLE_CHECKPOINT": &config.TreeReplicate.EnableCheckpoint,
		"FREIGHTLINER_TREE_SKIP_COMPLETED":    &config.TreeReplicate.SkipCompleted,
		"FREIGHTLINER_TREE_RETRY_FAILED":      &config.TreeReplicate.RetryFailed,

		// Replication configuration
		"FREIGHTLINER_REPLICATE_FORCE":   &config.Replicate.Force,
		"FREIGHTLINER_REPLICATE_DRY_RUN": &config.Replicate.DryRun,
	}

	// Load environment variables
	for env, field := range envVars {
		if value, exists := os.LookupEnv(env); exists {
			*field = strings.ToLower(value) == "true" || value == "1" || value == "yes" || value == "y"
		}
	}
}

// processIntEnvVars loads integer configuration from environment variables
func processIntEnvVars(config *Config) {
	// Map of environment variables to configuration fields
	envVars := map[string]*int{
		// Workers configuration
		"FREIGHTLINER_REPLICATE_WORKERS": &config.Workers.ReplicateWorkers,
		"FREIGHTLINER_SERVE_WORKERS":     &config.Workers.ServeWorkers,

		// Server configuration
		"FREIGHTLINER_SERVER_PORT": &config.Server.Port,

		// Tree replication configuration
		"FREIGHTLINER_TREE_WORKERS": &config.TreeReplicate.Workers,
	}

	// Load environment variables
	for env, field := range envVars {
		if value, exists := os.LookupEnv(env); exists && value != "" {
			if n, err := strconv.Atoi(value); err == nil {
				*field = n
			}
		}
	}
}

// processSliceEnvVars loads slice configuration from environment variables
func processSliceEnvVars(config *Config) {
	// Process string slice environment variables
	stringSliceEnvs := map[string]*[]string{
		"FREIGHTLINER_SERVER_ALLOWED_ORIGINS": &config.Server.AllowedOrigins,
		"FREIGHTLINER_TREE_EXCLUDE_REPOS":     &config.TreeReplicate.ExcludeRepos,
		"FREIGHTLINER_TREE_EXCLUDE_TAGS":      &config.TreeReplicate.ExcludeTags,
		"FREIGHTLINER_TREE_INCLUDE_TAGS":      &config.TreeReplicate.IncludeTags,
		"FREIGHTLINER_REPLICATE_TAGS":         &config.Replicate.Tags,
	}

	for env, field := range stringSliceEnvs {
		if value, exists := os.LookupEnv(env); exists && value != "" {
			// Split by comma, trim whitespace
			values := strings.Split(value, ",")
			trimmedValues := make([]string, 0, len(values))

			for _, v := range values {
				trimmed := strings.TrimSpace(v)
				if trimmed != "" {
					trimmedValues = append(trimmedValues, trimmed)
				}
			}

			if len(trimmedValues) > 0 {
				*field = trimmedValues
			}
		}
	}
}

// processDurationEnvVars loads time.Duration configuration from environment variables
func processDurationEnvVars(config *Config) {
	// Map of environment variables to configuration fields
	envVars := map[string]*time.Duration{
		"FREIGHTLINER_SERVER_READ_TIMEOUT":     &config.Server.ReadTimeout,
		"FREIGHTLINER_SERVER_WRITE_TIMEOUT":    &config.Server.WriteTimeout,
		"FREIGHTLINER_SERVER_SHUTDOWN_TIMEOUT": &config.Server.ShutdownTimeout,
	}

	// Load environment variables
	for env, field := range envVars {
		if value, exists := os.LookupEnv(env); exists && value != "" {
			if duration, err := time.ParseDuration(value); err == nil {
				*field = duration
			}
		}
	}
}

// SaveToFile saves the configuration to a file
func (c *Config) SaveToFile(filePath string) error {
	expandedPath := ExpandHomeDir(filePath)

	// Create directory if it doesn't exist
	dir := filepath.Dir(expandedPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return errors.Wrap(err, "failed to create directory")
	}

	// Create or truncate file
	file, err := os.Create(expandedPath)
	if err != nil {
		return errors.Wrap(err, "failed to create file")
	}
	defer func() {
		_ = file.Close()
	}()

	// Create encoder and encode config
	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2) // Makes the output more readable
	if err := encoder.Encode(c); err != nil {
		return errors.Wrap(err, "failed to encode configuration")
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Check if ECR region is provided when AWS KMS is enabled
	if c.Encryption.Enabled && c.Encryption.CustomerManagedKeys && c.Encryption.AWSKMSKeyID != "" && c.ECR.Region == "" {
		return errors.InvalidInputf("ECR region must be specified when using AWS KMS for encryption")
	}

	// Check if GCP project is provided when GCP KMS is enabled
	if c.Encryption.Enabled && c.Encryption.CustomerManagedKeys && c.Encryption.GCPKMSKeyID != "" && c.GCR.Project == "" {
		return errors.InvalidInputf("GCP project must be specified when using GCP KMS for encryption")
	}

	// Validate log level
	logLevel := strings.ToLower(c.LogLevel)
	if logLevel != "debug" && logLevel != "info" && logLevel != "warn" && logLevel != "error" && logLevel != "fatal" {
		return errors.InvalidInputf("invalid log level: %s (must be one of: debug, info, warn, error, fatal)", c.LogLevel)
	}

	// Validate worker counts
	if c.Workers.ReplicateWorkers < 0 {
		return errors.InvalidInputf("replicate workers must be non-negative")
	}
	if c.Workers.ServeWorkers < 0 {
		return errors.InvalidInputf("serve workers must be non-negative")
	}

	// Validate server configuration
	if c.Server.Port < 0 || c.Server.Port > 65535 {
		return errors.InvalidInputf("server port must be between 0 and 65535")
	}
	if c.Server.TLSEnabled && (c.Server.TLSCertFile == "" || c.Server.TLSKeyFile == "") {
		return errors.InvalidInputf("TLS certificate and key files must be provided when TLS is enabled")
	}
	if c.Server.APIKeyAuth && c.Server.APIKey == "" {
		return errors.InvalidInputf("API key must be provided when API key authentication is enabled")
	}

	// Validate secrets configuration
	if c.Secrets.UseSecretsManager {
		if c.Secrets.SecretsManagerType != "aws" && c.Secrets.SecretsManagerType != "gcp" {
			return errors.InvalidInputf("invalid secrets manager type: %s (must be one of: aws, gcp)", c.Secrets.SecretsManagerType)
		}
		if c.Secrets.SecretsManagerType == "aws" && c.Secrets.AWSSecretRegion == "" && c.ECR.Region == "" {
			return errors.InvalidInputf("AWS region must be specified when using AWS Secrets Manager")
		}
		if c.Secrets.SecretsManagerType == "gcp" && c.Secrets.GCPSecretProject == "" && c.GCR.Project == "" {
			return errors.InvalidInputf("GCP project must be specified when using Google Secret Manager")
		}
	}

	return nil
}
