package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadRegistriesConfig loads registry configurations from a YAML file
func LoadRegistriesConfig(path string) (*RegistriesConfig, error) {
	// Expand home directory if needed
	path = ExpandHomeDir(path)

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read registries config file: %w", err)
	}

	// Parse YAML
	var config RegistriesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse registries config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid registries configuration: %w", err)
	}

	return &config, nil
}

// LoadRegistriesConfigFromBytes loads registry configurations from YAML bytes
func LoadRegistriesConfigFromBytes(data []byte) (*RegistriesConfig, error) {
	var config RegistriesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse registries config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid registries configuration: %w", err)
	}

	return &config, nil
}

// SaveRegistriesConfig saves registry configurations to a YAML file
func SaveRegistriesConfig(path string, config *RegistriesConfig) error {
	// Validate configuration before saving
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid registries configuration: %w", err)
	}

	// Expand home directory if needed
	path = ExpandHomeDir(path)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal registries config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write registries config file: %w", err)
	}

	return nil
}

// GetDefaultRegistriesConfigPath returns the default path for registries configuration
func GetDefaultRegistriesConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./registries.yaml"
	}
	return filepath.Join(homeDir, ".freightliner", "registries.yaml")
}

// MergeRegistriesConfig merges a registry configuration into the main Config struct
// This provides backward compatibility with the existing ECR and GCR configuration
func (c *Config) MergeRegistriesConfig(registries *RegistriesConfig) error {
	if registries == nil {
		return nil
	}

	// Find ECR and GCR registries and update the legacy config
	for _, reg := range registries.Registries {
		switch reg.Type {
		case RegistryTypeECR:
			// Update ECR config with first ECR registry found
			if c.ECR.Region == "" && reg.Region != "" {
				c.ECR.Region = reg.Region
			}
			if c.ECR.AccountID == "" && reg.AccountID != "" {
				c.ECR.AccountID = reg.AccountID
			}

		case RegistryTypeGCR:
			// Update GCR config with first GCR registry found
			if c.GCR.Project == "" && reg.Project != "" {
				c.GCR.Project = reg.Project
			}
			if c.GCR.Location == "" && reg.Region != "" {
				c.GCR.Location = reg.Region
			}
		}
	}

	return nil
}

// CreateRegistryConfigFromLegacy creates a RegistryConfig from legacy ECR configuration
func CreateRegistryConfigFromLegacy(ecr ECRConfig, gcr GCRConfig) *RegistriesConfig {
	config := &RegistriesConfig{
		Registries: []RegistryConfig{},
	}

	// Add ECR registry if configured
	if ecr.Region != "" {
		ecrReg := RegistryConfig{
			Name:      "default-ecr",
			Type:      RegistryTypeECR,
			Region:    ecr.Region,
			AccountID: ecr.AccountID,
			Auth: AuthConfig{
				Type: AuthTypeAWS,
			},
		}
		config.Registries = append(config.Registries, ecrReg)
		config.DefaultSource = "default-ecr"
	}

	// Add GCR registry if configured
	if gcr.Project != "" {
		gcrReg := RegistryConfig{
			Name:    "default-gcr",
			Type:    RegistryTypeGCR,
			Project: gcr.Project,
			Region:  gcr.Location,
			Auth: AuthConfig{
				Type: AuthTypeGCP,
			},
		}
		config.Registries = append(config.Registries, gcrReg)
		if config.DefaultDestination == "" {
			config.DefaultDestination = "default-gcr"
		}
	}

	return config
}

// LoadOrCreateRegistriesConfig loads registries config from file or creates from legacy config
func (c *Config) LoadOrCreateRegistriesConfig(path string) (*RegistriesConfig, error) {
	// Try to load from file first
	if path != "" {
		config, err := LoadRegistriesConfig(path)
		if err == nil {
			return config, nil
		}
		// If file doesn't exist, continue to create from legacy
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	// Try default path
	defaultPath := GetDefaultRegistriesConfigPath()
	config, err := LoadRegistriesConfig(defaultPath)
	if err == nil {
		return config, nil
	}

	// If no config file exists, create from legacy ECR/GCR config
	if os.IsNotExist(err) {
		return CreateRegistryConfigFromLegacy(c.ECR, c.GCR), nil
	}

	return nil, err
}
