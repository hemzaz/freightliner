package config

import (
	"freightliner/pkg/helper/errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadFromFile loads configuration from a file
func LoadFromFile(configPath string) (*Config, error) {
	// Set default configuration
	config := NewDefaultConfig()

	// If configPath is provided, load config from file
	if configPath != "" {
		expandedPath := ExpandHomeDir(configPath)

		// Check if file exists
		if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
			return nil, errors.NotFoundf("configuration file not found: %s", expandedPath)
		}

		// Read file
		data, err := os.ReadFile(expandedPath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read configuration file")
		}

		// Unmarshal YAML
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, errors.Wrap(err, "failed to parse configuration file")
		}
	}

	// Load from environment variables
	if err := loadFromEnv(config); err != nil {
		return nil, err
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) error {
	// Map of environment variables to configuration fields
	envVars := map[string]*string{
		"FREIGHTLINER_LOG_LEVEL":      &config.LogLevel,
		"FREIGHTLINER_ECR_REGION":     &config.ECR.Region,
		"FREIGHTLINER_ECR_ACCOUNT_ID": &config.ECR.AccountID,
		"FREIGHTLINER_GCR_PROJECT":    &config.GCR.Project,
		"FREIGHTLINER_GCR_LOCATION":   &config.GCR.Location,

		// Add more mappings as needed
	}

	// Load environment variables
	for env, field := range envVars {
		if value, exists := os.LookupEnv(env); exists && value != "" {
			*field = value
		}
	}

	// Handle boolean and numeric environment variables
	if value, exists := os.LookupEnv("FREIGHTLINER_ENCRYPTION_ENABLED"); exists {
		config.Encryption.Enabled = strings.ToLower(value) == "true" || value == "1"
	}

	if value, exists := os.LookupEnv("FREIGHTLINER_REPLICATE_WORKERS"); exists {
		if n, err := strconv.Atoi(value); err == nil {
			config.Workers.ReplicateWorkers = n
		}
	}

	// More boolean and numeric environment variables can be handled here

	return nil
}

// SaveToFile saves the configuration to a file
func (c *Config) SaveToFile(filePath string) error {
	expandedPath := ExpandHomeDir(filePath)

	// Create directory if it doesn't exist
	dir := filepath.Dir(expandedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(err, "failed to create directory")
	}

	// Create or truncate file
	file, err := os.Create(expandedPath)
	if err != nil {
		return errors.Wrap(err, "failed to create file")
	}
	defer file.Close()

	// Create encoder and encode config
	encoder := yaml.NewEncoder(file)
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
