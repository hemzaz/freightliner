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
	LogLevel string

	// Registry configuration
	ECR ECRConfig
	GCR GCRConfig

	// Worker configuration
	Workers WorkerConfig

	// Encryption configuration
	Encryption EncryptionConfig

	// Secrets configuration
	Secrets SecretsConfig

	// Server configuration
	Server ServerConfig

	// Checkpoint configuration
	Checkpoint CheckpointConfig

	// Tree replication configuration
	TreeReplicate TreeReplicateConfig

	// Replicate configuration
	Replicate ReplicateConfig
}

// ECRConfig contains AWS ECR specific configuration
type ECRConfig struct {
	Region    string
	AccountID string
}

// GCRConfig contains Google Container Registry specific configuration
type GCRConfig struct {
	Project  string
	Location string
}

// WorkerConfig contains worker pool configuration
type WorkerConfig struct {
	ReplicateWorkers int
	ServeWorkers     int
	AutoDetect       bool
}

// EncryptionConfig contains encryption related configuration
type EncryptionConfig struct {
	Enabled             bool
	CustomerManagedKeys bool
	AWSKMSKeyID         string
	GCPKMSKeyID         string
	GCPKeyRing          string
	GCPKeyName          string
	EnvelopeEncryption  bool
}

// SecretsConfig contains secrets management configuration
type SecretsConfig struct {
	UseSecretsManager    bool
	SecretsManagerType   string
	AWSSecretRegion      string
	GCPSecretProject     string
	GCPCredentialsFile   string
	RegistryCredsSecret  string
	EncryptionKeysSecret string
}

// ServerConfig contains server related configuration
type ServerConfig struct {
	Port              int
	TLSEnabled        bool
	TLSCertFile       string
	TLSKeyFile        string
	APIKeyAuth        bool
	APIKey            string
	AllowedOrigins    []string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	ShutdownTimeout   time.Duration
	HealthCheckPath   string
	MetricsPath       string
	ReplicatePath     string
	TreeReplicatePath string
	StatusPath        string
}

// CheckpointConfig contains checkpoint related configuration
type CheckpointConfig struct {
	Directory string
	ID        string
}

// TreeReplicateConfig contains tree replication options
type TreeReplicateConfig struct {
	Workers          int
	ExcludeRepos     []string
	ExcludeTags      []string
	IncludeTags      []string
	DryRun           bool
	Force            bool
	EnableCheckpoint bool
	CheckpointDir    string
	ResumeID         string
	SkipCompleted    bool
	RetryFailed      bool
}

// ReplicateConfig contains single repository replication options
type ReplicateConfig struct {
	Force  bool
	DryRun bool
	Tags   []string
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
			Port:              8080,
			TLSEnabled:        false,
			TLSCertFile:       "",
			TLSKeyFile:        "",
			APIKeyAuth:        false,
			APIKey:            "",
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
	cmd.Flags().IntVar(&c.Server.Port, "port", c.Server.Port, "Server listening port")
	cmd.Flags().BoolVar(&c.Server.TLSEnabled, "tls", c.Server.TLSEnabled, "Enable TLS")
	cmd.Flags().StringVar(&c.Server.TLSCertFile, "tls-cert", c.Server.TLSCertFile, "TLS certificate file")
	cmd.Flags().StringVar(&c.Server.TLSKeyFile, "tls-key", c.Server.TLSKeyFile, "TLS key file")
	cmd.Flags().BoolVar(&c.Server.APIKeyAuth, "api-key-auth", c.Server.APIKeyAuth, "Enable API key authentication")
	cmd.Flags().StringVar(&c.Server.APIKey, "api-key", c.Server.APIKey, "API key for authentication")
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
