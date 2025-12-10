// Package sync provides YAML-based synchronization between container registries.
package sync

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the complete sync configuration
type Config struct {
	// Source registry configuration
	Source RegistryConfig `yaml:"source"`

	// Destination registry configuration
	Destination RegistryConfig `yaml:"destination"`

	// Images to sync with filtering rules
	Images []ImageSync `yaml:"images"`

	// Parallel specifies number of concurrent sync operations
	Parallel int `yaml:"parallel,omitempty"`

	// SkipExisting skips images that already exist in destination
	SkipExisting bool `yaml:"skip_existing,omitempty"`

	// ContinueOnError continues syncing even if some images fail
	ContinueOnError bool `yaml:"continue_on_error,omitempty"`

	// Timeout for each sync operation in seconds
	Timeout int `yaml:"timeout,omitempty"`

	// BatchSize for optimized batch operations (default: 10)
	BatchSize int `yaml:"batch_size,omitempty"`

	// EnableAdaptiveBatching enables dynamic batch size adjustment based on workload
	EnableAdaptiveBatching bool `yaml:"enable_adaptive_batching,omitempty"`

	// MinBatchSize minimum batch size for adaptive batching (default: 1)
	MinBatchSize int `yaml:"min_batch_size,omitempty"`

	// MaxBatchSize maximum batch size for adaptive batching (default: 50)
	MaxBatchSize int `yaml:"max_batch_size,omitempty"`

	// EnableDeduplication enables content-addressable storage deduplication
	EnableDeduplication bool `yaml:"enable_deduplication,omitempty"`

	// EnableHTTP3 enables HTTP/3 with QUIC protocol for faster transfers
	EnableHTTP3 bool `yaml:"enable_http3,omitempty"`

	// RetryAttempts specifies number of retry attempts for failed syncs (default: 3)
	RetryAttempts int `yaml:"retry_attempts,omitempty"`

	// RetryBackoff specifies retry backoff in seconds (default: 5)
	RetryBackoff int `yaml:"retry_backoff,omitempty"`
}

// RegistryConfig represents registry connection configuration
type RegistryConfig struct {
	// Registry URL (e.g., "docker.io", "gcr.io")
	Registry string `yaml:"registry"`

	// Type of registry (docker, ecr, gcr, acr, ghcr, harbor, quay, generic)
	Type string `yaml:"type,omitempty"`

	// Auth configuration
	Auth *AuthConfig `yaml:"auth,omitempty"`

	// Insecure allows insecure connections (skip TLS verification)
	Insecure bool `yaml:"insecure,omitempty"`

	// Region for cloud registries (ECR, GCR)
	Region string `yaml:"region,omitempty"`

	// Project for GCR
	Project string `yaml:"project,omitempty"`

	// Account for ECR
	Account string `yaml:"account,omitempty"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	// Username for basic auth
	Username string `yaml:"username,omitempty"`

	// Password for basic auth
	Password string `yaml:"password,omitempty"`

	// Token for token-based auth
	Token string `yaml:"token,omitempty"`

	// UseDockerConfig uses credentials from ~/.docker/config.json
	UseDockerConfig bool `yaml:"use_docker_config,omitempty"`

	// AWSProfile for ECR authentication
	AWSProfile string `yaml:"aws_profile,omitempty"`

	// GCPCredentials path to GCP credentials file
	GCPCredentials string `yaml:"gcp_credentials,omitempty"`
}

// ImageSync represents a single image synchronization rule
type ImageSync struct {
	// Repository path (e.g., "library/nginx")
	Repository string `yaml:"repository"`

	// Tags is a list of specific tags to sync
	Tags []string `yaml:"tags,omitempty"`

	// TagRegex is a regex pattern for matching tags
	TagRegex string `yaml:"tag_regex,omitempty"`

	// SemverConstraint is a semantic version constraint (e.g., ">=1.2.3", "^2.0.0")
	SemverConstraint string `yaml:"semver_constraint,omitempty"`

	// AllTags syncs all tags in the repository
	AllTags bool `yaml:"all_tags,omitempty"`

	// LatestN syncs only the latest N tags (by creation date)
	LatestN int `yaml:"latest_n,omitempty"`

	// DestinationRepository overrides the destination repository path
	DestinationRepository string `yaml:"destination_repository,omitempty"`

	// DestinationPrefix adds a prefix to destination tags
	DestinationPrefix string `yaml:"destination_prefix,omitempty"`

	// DestinationSuffix adds a suffix to destination tags
	DestinationSuffix string `yaml:"destination_suffix,omitempty"`

	// Limit limits the number of tags to sync
	Limit int `yaml:"limit,omitempty"`

	// Architectures to sync (e.g., ["amd64", "arm64"])
	Architectures []string `yaml:"architectures,omitempty"`

	// SignVerification requires signature verification (Cosign)
	SignVerification *SignatureConfig `yaml:"sign_verification,omitempty"`

	// SkipLayers allows skipping specific layers (advanced)
	SkipLayers []string `yaml:"skip_layers,omitempty"`
}

// SignatureConfig represents signature verification configuration
type SignatureConfig struct {
	// Enabled enables signature verification
	Enabled bool `yaml:"enabled"`

	// PublicKey path to public key for verification
	PublicKey string `yaml:"public_key,omitempty"`

	// KeylessVerification enables keyless verification using Fulcio
	KeylessVerification bool `yaml:"keyless_verification,omitempty"`

	// CertificateIdentity for keyless verification
	CertificateIdentity string `yaml:"certificate_identity,omitempty"`

	// Issuer for keyless verification
	Issuer string `yaml:"issuer,omitempty"`
}

// LoadConfig loads and validates a sync configuration from a YAML file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Set defaults
	config.SetDefaults()

	return &config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate source registry
	if c.Source.Registry == "" {
		return fmt.Errorf("source.registry is required")
	}

	// Validate destination registry
	if c.Destination.Registry == "" {
		return fmt.Errorf("destination.registry is required")
	}

	// Validate images
	if len(c.Images) == 0 {
		return fmt.Errorf("at least one image must be specified")
	}

	for i, img := range c.Images {
		if img.Repository == "" {
			return fmt.Errorf("images[%d].repository is required", i)
		}

		// Validate filter exclusivity
		filterCount := 0
		if len(img.Tags) > 0 {
			filterCount++
		}
		if img.TagRegex != "" {
			filterCount++
		}
		if img.SemverConstraint != "" {
			filterCount++
		}
		if img.AllTags {
			filterCount++
		}
		if img.LatestN > 0 {
			filterCount++
		}

		if filterCount == 0 {
			return fmt.Errorf("images[%d]: must specify at least one of: tags, tag_regex, semver_constraint, all_tags, or latest_n", i)
		}
		if filterCount > 1 {
			return fmt.Errorf("images[%d]: cannot specify multiple tag filters (tags, tag_regex, semver_constraint, all_tags, latest_n)", i)
		}
	}

	return nil
}

// SetDefaults sets default values for optional fields
func (c *Config) SetDefaults() {
	if c.Parallel <= 0 {
		c.Parallel = 3
	}
	if c.BatchSize <= 0 {
		c.BatchSize = 10
	}
	if c.Timeout <= 0 {
		c.Timeout = 300 // 5 minutes
	}
	if c.RetryAttempts <= 0 {
		c.RetryAttempts = 3
	}
	if c.RetryBackoff <= 0 {
		c.RetryBackoff = 5
	}
	if c.MinBatchSize <= 0 {
		c.MinBatchSize = 1
	}
	if c.MaxBatchSize <= 0 {
		c.MaxBatchSize = 50
	}
	// EnableAdaptiveBatching defaults to false for backward compatibility

	// Set default registry types if not specified
	if c.Source.Type == "" {
		c.Source.Type = detectRegistryType(c.Source.Registry)
	}
	if c.Destination.Type == "" {
		c.Destination.Type = detectRegistryType(c.Destination.Registry)
	}
}

// detectRegistryType detects registry type from URL
func detectRegistryType(registry string) string {
	switch {
	case registry == "docker.io" || registry == "registry-1.docker.io":
		return "docker"
	case registry == "gcr.io" || registry == "us.gcr.io" || registry == "eu.gcr.io" || registry == "asia.gcr.io":
		return "gcr"
	case registry == "ghcr.io":
		return "ghcr"
	case len(registry) >= 12 && registry[:12] == "public.ecr.aws":
		return "ecr"
	case len(registry) >= 23 && strings.Contains(registry, ".dkr.ecr."):
		return "ecr"
	case len(registry) >= 11 && registry[len(registry)-11:] == ".azurecr.io":
		return "acr"
	case len(registry) >= 8 && registry[len(registry)-8:] == ".quay.io" || registry == "quay.io":
		return "quay"
	default:
		return "generic"
	}
}

// SyncTask represents a resolved sync task ready for execution
type SyncTask struct {
	// Source
	SourceRegistry   string
	SourceRepository string
	SourceTag        string

	// Destination
	DestRegistry   string
	DestRepository string
	DestTag        string

	// Metadata
	Architecture     string
	SignVerification *SignatureConfig
	Priority         int
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	Task        SyncTask
	Success     bool
	Error       error
	BytesCopied int64
	Duration    int64 // milliseconds
	Retries     int
	Skipped     bool
	SkipReason  string
}
