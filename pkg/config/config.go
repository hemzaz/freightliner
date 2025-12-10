package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Config represents the main application configuration
type Config struct {
	// General configuration
	LogLevel string `yaml:"log_level" json:"log_level"`

	// Registry configuration
	ECR        ECRConfig        `yaml:"ecr" json:"ecr"`
	GCR        GCRConfig        `yaml:"gcr" json:"gcr"`
	Registries RegistriesConfig `yaml:"registries" json:"registries"`

	// Worker configuration
	Workers WorkerConfig `yaml:"workers" json:"workers"`

	// Encryption configuration
	Encryption EncryptionConfig `yaml:"encryption" json:"encryption"`

	// Secrets configuration
	Secrets SecretsConfig `yaml:"secrets" json:"secrets"`

	// Server configuration
	Server ServerConfig `yaml:"server" json:"server"`

	// Metrics configuration
	Metrics MetricsConfig `yaml:"metrics" json:"metrics"`

	// Checkpoint configuration
	Checkpoint CheckpointConfig `yaml:"checkpoint" json:"checkpoint"`

	// Tree replication configuration
	TreeReplicate TreeReplicateConfig `yaml:"tree_replicate" json:"tree_replicate"`

	// Replicate configuration
	Replicate ReplicateConfig `yaml:"replicate" json:"replicate"`
}

// ECRConfig contains AWS ECR specific configuration
type ECRConfig struct {
	Region    string `yaml:"region" json:"region"`
	AccountID string `yaml:"account_id" json:"account_id"`
}

// GCRConfig contains Google Container Registry specific configuration
type GCRConfig struct {
	Project  string `yaml:"project" json:"project"`
	Location string `yaml:"location" json:"location"`
}

// WorkerConfig contains worker pool configuration
type WorkerConfig struct {
	ReplicateWorkers int  `yaml:"replicate_workers" json:"replicate_workers"`
	ServeWorkers     int  `yaml:"serve_workers" json:"serve_workers"`
	AutoDetect       bool `yaml:"auto_detect" json:"auto_detect"`
}

// EncryptionConfig contains encryption related configuration
type EncryptionConfig struct {
	Enabled             bool   `yaml:"enabled" json:"enabled"`
	CustomerManagedKeys bool   `yaml:"customer_managed_keys" json:"customer_managed_keys"`
	AWSKMSKeyID         string `yaml:"aws_kms_key_id" json:"aws_kms_key_id"`
	GCPKMSKeyID         string `yaml:"gcp_kms_key_id" json:"gcp_kms_key_id"`
	GCPKeyRing          string `yaml:"gcp_key_ring" json:"gcp_key_ring"`
	GCPKeyName          string `yaml:"gcp_key_name" json:"gcp_key_name"`
	EnvelopeEncryption  bool   `yaml:"envelope_encryption" json:"envelope_encryption"`
}

// SecretsConfig contains secrets management configuration
type SecretsConfig struct {
	UseSecretsManager    bool   `yaml:"use_secrets_manager" json:"use_secrets_manager"`
	SecretsManagerType   string `yaml:"secrets_manager_type" json:"secrets_manager_type"`
	AWSSecretRegion      string `yaml:"aws_secret_region" json:"aws_secret_region"`
	GCPSecretProject     string `yaml:"gcp_secret_project" json:"gcp_secret_project"`
	GCPCredentialsFile   string `yaml:"gcp_credentials_file" json:"gcp_credentials_file"`
	RegistryCredsSecret  string `yaml:"registry_creds_secret" json:"registry_creds_secret"`
	EncryptionKeysSecret string `yaml:"encryption_keys_secret" json:"encryption_keys_secret"`
}

// ServerConfig contains server related configuration
type ServerConfig struct {
	Host              string        `yaml:"host" json:"host"`                 // Bind address: "localhost", "0.0.0.0", or specific IP
	Port              int           `yaml:"port" json:"port"`                 // Bind port
	ExternalURL       string        `yaml:"external_url" json:"external_url"` // External URL for API access (e.g., "https://api.example.com")
	TLSEnabled        bool          `yaml:"tls_enabled" json:"tls_enabled"`
	TLSCertFile       string        `yaml:"tls_cert_file" json:"tls_cert_file"`
	TLSKeyFile        string        `yaml:"tls_key_file" json:"tls_key_file"`
	APIKeyAuth        bool          `yaml:"api_key_auth" json:"api_key_auth"`
	APIKey            string        `yaml:"api_key" json:"api_key"`
	EnableCORS        bool          `yaml:"enable_cors" json:"enable_cors"`         // Enable CORS middleware
	AllowedOrigins    []string      `yaml:"allowed_origins" json:"allowed_origins"` // CORS allowed origins
	ReadTimeout       time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout" json:"write_timeout"`
	ShutdownTimeout   time.Duration `yaml:"shutdown_timeout" json:"shutdown_timeout"`
	HealthCheckPath   string        `yaml:"health_check_path" json:"health_check_path"` // Health check endpoint path
	MetricsPath       string        `yaml:"metrics_path" json:"metrics_path"`           // Metrics endpoint path
	ReplicatePath     string        `yaml:"replicate_path" json:"replicate_path"`
	TreeReplicatePath string        `yaml:"tree_replicate_path" json:"tree_replicate_path"`
	StatusPath        string        `yaml:"status_path" json:"status_path"`
}

// CheckpointConfig contains checkpoint related configuration
type CheckpointConfig struct {
	Directory string `yaml:"directory" json:"directory"`
	ID        string `yaml:"id" json:"id"`
}

// TreeReplicateConfig contains tree replication options
type TreeReplicateConfig struct {
	Workers          int      `yaml:"workers" json:"workers"`
	ExcludeRepos     []string `yaml:"exclude_repos" json:"exclude_repos"`
	ExcludeTags      []string `yaml:"exclude_tags" json:"exclude_tags"`
	IncludeTags      []string `yaml:"include_tags" json:"include_tags"`
	DryRun           bool     `yaml:"dry_run" json:"dry_run"`
	Force            bool     `yaml:"force" json:"force"`
	EnableCheckpoint bool     `yaml:"enable_checkpoint" json:"enable_checkpoint"`
	CheckpointDir    string   `yaml:"checkpoint_dir" json:"checkpoint_dir"`
	ResumeID         string   `yaml:"resume_id" json:"resume_id"`
	SkipCompleted    bool     `yaml:"skip_completed" json:"skip_completed"`
	RetryFailed      bool     `yaml:"retry_failed" json:"retry_failed"`
}

// ReplicateConfig contains single repository replication options
type ReplicateConfig struct {
	Force  bool     `yaml:"force" json:"force"`
	DryRun bool     `yaml:"dry_run" json:"dry_run"`
	Tags   []string `yaml:"tags" json:"tags"`
}

// NewDefaultConfig creates a new configuration with default values
func NewDefaultConfig() *Config {
	return &Config{
		LogLevel: "info",
		ECR: ECRConfig{
			Region:    "us-west-2",
			AccountID: "",
		},
		GCR: GCRConfig{
			Project:  "",
			Location: "us",
		},
		Registries: RegistriesConfig{
			DefaultSource:      "",
			DefaultDestination: "",
			Registries:         []RegistryConfig{},
		},
		Workers: WorkerConfig{
			ReplicateWorkers: 0,
			ServeWorkers:     0,
			AutoDetect:       true,
		},
		Encryption: EncryptionConfig{
			Enabled:             false,
			CustomerManagedKeys: false,
			AWSKMSKeyID:         "",
			GCPKMSKeyID:         "",
			GCPKeyRing:          "freightliner",
			GCPKeyName:          "image-encryption",
			EnvelopeEncryption:  true,
		},
		Secrets: SecretsConfig{
			UseSecretsManager:    false,
			SecretsManagerType:   "aws",
			AWSSecretRegion:      "",
			GCPSecretProject:     "",
			GCPCredentialsFile:   "",
			RegistryCredsSecret:  "freightliner-registry-credentials",
			EncryptionKeysSecret: "freightliner-encryption-keys",
		},
		Server: ServerConfig{
			Host:              "localhost", // Default to localhost for security
			Port:              8080,
			ExternalURL:       "", // Empty means use Host:Port
			TLSEnabled:        false,
			TLSCertFile:       "",
			TLSKeyFile:        "",
			APIKeyAuth:        false,
			APIKey:            "",
			EnableCORS:        true, // Enable CORS by default
			AllowedOrigins:    []string{"*"},
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      60 * time.Second,
			ShutdownTimeout:   15 * time.Second,
			HealthCheckPath:   "/health",
			MetricsPath:       "/metrics",
			ReplicatePath:     "/api/v1/replicate",
			TreeReplicatePath: "/api/v1/replicate-tree",
			StatusPath:        "/api/v1/status",
		},
		Metrics: MetricsConfig{
			Enabled:   true,
			Port:      2112,
			Path:      "/metrics",
			Namespace: "freightliner",
		},
		Checkpoint: CheckpointConfig{
			Directory: "${HOME}/.freightliner/checkpoints",
			ID:        "",
		},
		TreeReplicate: TreeReplicateConfig{
			Workers:          0,
			ExcludeRepos:     []string{},
			ExcludeTags:      []string{},
			IncludeTags:      []string{},
			DryRun:           false,
			Force:            false,
			EnableCheckpoint: false,
			CheckpointDir:    "${HOME}/.freightliner/checkpoints",
			ResumeID:         "",
			SkipCompleted:    true,
			RetryFailed:      true,
		},
		Replicate: ReplicateConfig{
			Force:  false,
			DryRun: false,
			Tags:   []string{},
		},
	}
}

// AddFlagsToCommand adds configuration flags to a cobra command
func (c *Config) AddFlagsToCommand(cmd *cobra.Command) {
	// Add global flags
	cmd.PersistentFlags().StringVar(&c.LogLevel, "log-level", c.LogLevel, "Log level (debug, info, warn, error, fatal)")
	cmd.PersistentFlags().StringVar(&c.ECR.Region, "ecr-region", c.ECR.Region, "AWS region for ECR")
	cmd.PersistentFlags().StringVar(&c.ECR.AccountID, "ecr-account", c.ECR.AccountID, "AWS account ID for ECR (empty uses default from credentials)")
	cmd.PersistentFlags().StringVar(&c.GCR.Project, "gcr-project", c.GCR.Project, "GCP project for GCR")
	cmd.PersistentFlags().StringVar(&c.GCR.Location, "gcr-location", c.GCR.Location, "GCR location (us, eu, asia)")

	// Add worker configuration flags
	cmd.PersistentFlags().IntVar(&c.Workers.ReplicateWorkers, "replicate-workers", c.Workers.ReplicateWorkers, "Number of concurrent workers for replication (0 = auto-detect)")
	cmd.PersistentFlags().IntVar(&c.Workers.ServeWorkers, "serve-workers", c.Workers.ServeWorkers, "Number of concurrent workers for server mode (0 = auto-detect)")
	cmd.PersistentFlags().BoolVar(&c.Workers.AutoDetect, "auto-detect-workers", c.Workers.AutoDetect, "Auto-detect optimal worker count based on system resources")

	// Add encryption-related global flags
	cmd.PersistentFlags().BoolVar(&c.Encryption.Enabled, "encrypt", c.Encryption.Enabled, "Enable image encryption")
	cmd.PersistentFlags().BoolVar(&c.Encryption.CustomerManagedKeys, "customer-key", c.Encryption.CustomerManagedKeys, "Use customer-managed encryption keys")
	cmd.PersistentFlags().StringVar(&c.Encryption.AWSKMSKeyID, "aws-kms-key", c.Encryption.AWSKMSKeyID, "AWS KMS key ID for encryption")
	cmd.PersistentFlags().StringVar(&c.Encryption.GCPKMSKeyID, "gcp-kms-key", c.Encryption.GCPKMSKeyID, "GCP KMS key ID for encryption")
	cmd.PersistentFlags().StringVar(&c.Encryption.GCPKeyRing, "gcp-key-ring", c.Encryption.GCPKeyRing, "GCP KMS key ring name")
	cmd.PersistentFlags().StringVar(&c.Encryption.GCPKeyName, "gcp-key-name", c.Encryption.GCPKeyName, "GCP KMS key name")
	cmd.PersistentFlags().BoolVar(&c.Encryption.EnvelopeEncryption, "envelope-encryption", c.Encryption.EnvelopeEncryption, "Use envelope encryption")

	// Add secrets manager related flags
	cmd.PersistentFlags().BoolVar(&c.Secrets.UseSecretsManager, "use-secrets-manager", c.Secrets.UseSecretsManager, "Use cloud provider secrets manager for credentials")
	cmd.PersistentFlags().StringVar(&c.Secrets.SecretsManagerType, "secrets-manager-type", c.Secrets.SecretsManagerType, "Type of secrets manager to use (aws, gcp)")
	cmd.PersistentFlags().StringVar(&c.Secrets.AWSSecretRegion, "aws-secret-region", c.Secrets.AWSSecretRegion, "AWS region for Secrets Manager (defaults to --ecr-region if not specified)")
	cmd.PersistentFlags().StringVar(&c.Secrets.GCPSecretProject, "gcp-secret-project", c.Secrets.GCPSecretProject, "GCP project for Secret Manager (defaults to --gcr-project if not specified)")
	cmd.PersistentFlags().StringVar(&c.Secrets.GCPCredentialsFile, "gcp-credentials-file", c.Secrets.GCPCredentialsFile, "GCP credentials file path for Secret Manager")
	cmd.PersistentFlags().StringVar(&c.Secrets.RegistryCredsSecret, "registry-creds-secret", c.Secrets.RegistryCredsSecret, "Secret name for registry credentials")
	cmd.PersistentFlags().StringVar(&c.Secrets.EncryptionKeysSecret, "encryption-keys-secret", c.Secrets.EncryptionKeysSecret, "Secret name for encryption keys")
}

// AddCheckpointFlagsToCommand adds checkpoint-specific flags to a command
func (c *Config) AddCheckpointFlagsToCommand(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&c.Checkpoint.Directory, "checkpoint-dir", c.Checkpoint.Directory, "Directory for checkpoint files")
	cmd.Flags().StringVar(&c.Checkpoint.ID, "id", c.Checkpoint.ID, "Checkpoint ID for operations")
}

// AddTreeReplicateFlags adds tree replication-specific flags to a command
func (c *Config) AddTreeReplicateFlags(cmd *cobra.Command) {
	cmd.Flags().IntVar(&c.TreeReplicate.Workers, "workers", c.TreeReplicate.Workers, "Number of concurrent worker threads (0 = auto-detect)")
	cmd.Flags().StringSliceVar(&c.TreeReplicate.ExcludeRepos, "exclude-repo", c.TreeReplicate.ExcludeRepos, "Repository patterns to exclude (e.g. 'helper-*')")
	cmd.Flags().StringSliceVar(&c.TreeReplicate.ExcludeTags, "exclude-tag", c.TreeReplicate.ExcludeTags, "Tag patterns to exclude (e.g. 'dev-*')")
	cmd.Flags().StringSliceVar(&c.TreeReplicate.IncludeTags, "include-tag", c.TreeReplicate.IncludeTags, "Tag patterns to include (e.g. 'v*')")
	cmd.Flags().BoolVar(&c.TreeReplicate.DryRun, "dry-run", c.TreeReplicate.DryRun, "Perform a dry run without actually copying images")
	cmd.Flags().BoolVar(&c.TreeReplicate.Force, "force", c.TreeReplicate.Force, "Force overwrite of existing images")
	cmd.Flags().BoolVar(&c.TreeReplicate.EnableCheckpoint, "checkpoint", c.TreeReplicate.EnableCheckpoint, "Enable checkpointing for interrupted replications")
	cmd.Flags().StringVar(&c.TreeReplicate.CheckpointDir, "checkpoint-dir", c.TreeReplicate.CheckpointDir, "Directory for storing checkpoint files")
	cmd.Flags().StringVar(&c.TreeReplicate.ResumeID, "resume", c.TreeReplicate.ResumeID, "Resume replication from a checkpoint ID")
	cmd.Flags().BoolVar(&c.TreeReplicate.SkipCompleted, "skip-completed", c.TreeReplicate.SkipCompleted, "Skip completed repositories when resuming")
	cmd.Flags().BoolVar(&c.TreeReplicate.RetryFailed, "retry-failed", c.TreeReplicate.RetryFailed, "Retry failed repositories when resuming")
}

// AddServerFlagsToCommand adds server-specific flags to a command
func (c *Config) AddServerFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&c.Server.Host, "host", c.Server.Host, "Server bind address (localhost, 0.0.0.0, or specific IP)")
	cmd.Flags().IntVar(&c.Server.Port, "port", c.Server.Port, "Server listening port")
	cmd.Flags().StringVar(&c.Server.ExternalURL, "external-url", c.Server.ExternalURL, "External URL for API access (e.g., https://api.example.com)")
	cmd.Flags().BoolVar(&c.Server.TLSEnabled, "tls", c.Server.TLSEnabled, "Enable TLS")
	cmd.Flags().StringVar(&c.Server.TLSCertFile, "tls-cert", c.Server.TLSCertFile, "TLS certificate file")
	cmd.Flags().StringVar(&c.Server.TLSKeyFile, "tls-key", c.Server.TLSKeyFile, "TLS key file")
	cmd.Flags().BoolVar(&c.Server.APIKeyAuth, "api-key-auth", c.Server.APIKeyAuth, "Enable API key authentication")
	cmd.Flags().StringVar(&c.Server.APIKey, "api-key", c.Server.APIKey, "API key for authentication")
	cmd.Flags().BoolVar(&c.Server.EnableCORS, "enable-cors", c.Server.EnableCORS, "Enable CORS middleware")
	cmd.Flags().StringSliceVar(&c.Server.AllowedOrigins, "allowed-origins", c.Server.AllowedOrigins, "Allowed CORS origins")
	cmd.Flags().DurationVar(&c.Server.ReadTimeout, "read-timeout", c.Server.ReadTimeout, "HTTP server read timeout")
	cmd.Flags().DurationVar(&c.Server.WriteTimeout, "write-timeout", c.Server.WriteTimeout, "HTTP server write timeout")
	cmd.Flags().DurationVar(&c.Server.ShutdownTimeout, "shutdown-timeout", c.Server.ShutdownTimeout, "HTTP server shutdown timeout")
}

// AddReplicateFlags adds single repository replication-specific flags to a command
func (c *Config) AddReplicateFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&c.Replicate.Force, "force", c.Replicate.Force, "Force overwrite of existing images")
	cmd.Flags().BoolVar(&c.Replicate.DryRun, "dry-run", c.Replicate.DryRun, "Perform a dry run without actually copying images")
	cmd.Flags().StringSliceVar(&c.Replicate.Tags, "tags", c.Replicate.Tags, "Specific tags to replicate (if empty, all tags will be replicated)")
}

// ExpandHomeDir expands the ~ or $HOME at the beginning of a directory path
func ExpandHomeDir(path string) string {
	if path == "" {
		return path
	}

	// Replace ${HOME} with actual home directory
	if strings.Contains(path, "${HOME}") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = strings.ReplaceAll(path, "${HOME}", homeDir)
		}
	}

	// Replace ~ with actual home directory
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[1:])
		}
	}

	return path
}

// GetOptimalWorkerCount determines the optimal number of worker threads
func GetOptimalWorkerCount() int {
	numCPU := runtime.NumCPU()

	// Simple heuristic:
	// - Minimum of 2 workers
	// - For small machines, use one worker per core
	// - For larger machines, leave one core free for system tasks

	if numCPU <= 2 {
		return 2 // Always have at least 2 workers
	} else if numCPU <= 4 {
		return numCPU // For small machines, use one worker per core
	} else {
		return numCPU - 1 // For larger machines, leave one core free for system tasks
	}
}
