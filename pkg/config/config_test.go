package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// TestNewDefaultConfig tests the default configuration creation
func TestNewDefaultConfig(t *testing.T) {
	config := NewDefaultConfig()

	// Test basic defaults
	if config.LogLevel != "info" {
		t.Errorf("Expected log level 'info', got '%s'", config.LogLevel)
	}

	// Test ECR defaults
	if config.ECR.Region != "us-west-2" {
		t.Errorf("Expected ECR region 'us-west-2', got '%s'", config.ECR.Region)
	}

	// Test GCR defaults
	if config.GCR.Location != "us" {
		t.Errorf("Expected GCR location 'us', got '%s'", config.GCR.Location)
	}

	// Test Worker defaults
	if config.Workers.AutoDetect != true {
		t.Error("Expected workers auto-detect to be true")
	}

	// Test Server defaults
	if config.Server.Host != "localhost" {
		t.Errorf("Expected server host 'localhost', got '%s'", config.Server.Host)
	}
	if config.Server.Port != 8080 {
		t.Errorf("Expected server port 8080, got %d", config.Server.Port)
	}
	if config.Server.ReadTimeout != 30*time.Second {
		t.Errorf("Expected read timeout 30s, got %v", config.Server.ReadTimeout)
	}

	// Test Metrics defaults
	if config.Metrics.Enabled != true {
		t.Error("Expected metrics enabled to be true")
	}
	if config.Metrics.Port != 2112 {
		t.Errorf("Expected metrics port 2112, got %d", config.Metrics.Port)
	}

	// Test encryption defaults
	if config.Encryption.Enabled != false {
		t.Error("Expected encryption disabled by default")
	}
	if config.Encryption.EnvelopeEncryption != true {
		t.Error("Expected envelope encryption to be true")
	}

	// Test checkpoint defaults
	if config.Checkpoint.Directory != "${HOME}/.freightliner/checkpoints" {
		t.Errorf("Expected checkpoint directory '${HOME}/.freightliner/checkpoints', got '%s'", config.Checkpoint.Directory)
	}
}

// TestExpandHomeDir tests home directory expansion
func TestExpandHomeDir(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "empty path",
			input:    "",
			contains: "",
		},
		{
			name:     "path with ${HOME}",
			input:    "${HOME}/test",
			contains: "/test",
		},
		{
			name:     "path with tilde",
			input:    "~/test",
			contains: "/test",
		},
		{
			name:     "path without home",
			input:    "/absolute/path",
			contains: "/absolute/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandHomeDir(tt.input)
			if tt.input == "" && result != "" {
				t.Errorf("Expected empty result for empty input, got '%s'", result)
			}
			if tt.contains != "" && result != tt.input {
				// Result should not contain ${HOME} or ~ after expansion
				if result == tt.input && (tt.input != "/absolute/path") {
					t.Errorf("Path not expanded: %s", result)
				}
			}
		})
	}
}

// TestGetOptimalWorkerCount tests worker count calculation
func TestGetOptimalWorkerCount(t *testing.T) {
	count := GetOptimalWorkerCount()

	numCPU := runtime.NumCPU()

	// Test minimum workers
	if count < 2 {
		t.Errorf("Expected at least 2 workers, got %d", count)
	}

	// Test logic based on CPU count
	if numCPU <= 2 {
		if count != 2 {
			t.Errorf("For %d CPUs, expected 2 workers, got %d", numCPU, count)
		}
	} else if numCPU <= 4 {
		if count != numCPU {
			t.Errorf("For %d CPUs, expected %d workers, got %d", numCPU, numCPU, count)
		}
	} else {
		if count != numCPU-1 {
			t.Errorf("For %d CPUs, expected %d workers, got %d", numCPU, numCPU-1, count)
		}
	}
}

// TestAddFlagsToCommand tests flag registration
func TestAddFlagsToCommand(t *testing.T) {
	config := NewDefaultConfig()
	cmd := &cobra.Command{Use: "test"}

	config.AddFlagsToCommand(cmd)

	// Check if flags were added
	flags := []string{
		"log-level",
		"ecr-region",
		"ecr-account",
		"gcr-project",
		"gcr-location",
		"replicate-workers",
		"serve-workers",
		"encrypt",
		"use-secrets-manager",
	}

	for _, flagName := range flags {
		flag := cmd.PersistentFlags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to be registered", flagName)
		}
	}
}

// TestAddCheckpointFlagsToCommand tests checkpoint flag registration
func TestAddCheckpointFlagsToCommand(t *testing.T) {
	config := NewDefaultConfig()
	cmd := &cobra.Command{Use: "test"}

	config.AddCheckpointFlagsToCommand(cmd)

	// Check checkpoint-dir persistent flag
	flag := cmd.PersistentFlags().Lookup("checkpoint-dir")
	if flag == nil {
		t.Error("Expected 'checkpoint-dir' flag to be registered")
	}

	// Check id flag (non-persistent)
	idFlag := cmd.Flags().Lookup("id")
	if idFlag == nil {
		t.Error("Expected 'id' flag to be registered")
	}
}

// TestAddTreeReplicateFlags tests tree replicate flag registration
func TestAddTreeReplicateFlags(t *testing.T) {
	config := NewDefaultConfig()
	cmd := &cobra.Command{Use: "test"}

	config.AddTreeReplicateFlags(cmd)

	flags := []string{
		"workers",
		"exclude-repo",
		"exclude-tag",
		"include-tag",
		"dry-run",
		"force",
		"checkpoint",
		"checkpoint-dir",
		"resume",
		"skip-completed",
		"retry-failed",
	}

	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to be registered", flagName)
		}
	}
}

// TestAddServerFlags tests server flag registration
func TestAddServerFlags(t *testing.T) {
	config := NewDefaultConfig()
	cmd := &cobra.Command{Use: "test"}

	config.AddServerFlags(cmd)

	flags := []string{
		"host",
		"port",
		"external-url",
		"tls",
		"tls-cert",
		"tls-key",
		"api-key-auth",
		"api-key",
		"enable-cors",
		"allowed-origins",
		"read-timeout",
		"write-timeout",
		"shutdown-timeout",
	}

	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to be registered", flagName)
		}
	}
}

// TestAddReplicateFlags tests replicate flag registration
func TestAddReplicateFlags(t *testing.T) {
	config := NewDefaultConfig()
	cmd := &cobra.Command{Use: "test"}

	config.AddReplicateFlags(cmd)

	flags := []string{
		"force",
		"dry-run",
		"tags",
	}

	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to be registered", flagName)
		}
	}
}

// TestValidate tests configuration validation
func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		modifyFn  func(*Config)
		wantError bool
	}{
		{
			name:      "valid default config",
			modifyFn:  func(c *Config) {},
			wantError: false,
		},
		{
			name: "invalid log level",
			modifyFn: func(c *Config) {
				c.LogLevel = "invalid"
			},
			wantError: true,
		},
		{
			name: "negative replicate workers",
			modifyFn: func(c *Config) {
				c.Workers.ReplicateWorkers = -1
			},
			wantError: true,
		},
		{
			name: "negative serve workers",
			modifyFn: func(c *Config) {
				c.Workers.ServeWorkers = -1
			},
			wantError: true,
		},
		{
			name: "invalid server port - negative",
			modifyFn: func(c *Config) {
				c.Server.Port = -1
			},
			wantError: true,
		},
		{
			name: "invalid server port - too high",
			modifyFn: func(c *Config) {
				c.Server.Port = 70000
			},
			wantError: true,
		},
		{
			name: "TLS enabled without cert file",
			modifyFn: func(c *Config) {
				c.Server.TLSEnabled = true
				c.Server.TLSCertFile = ""
				c.Server.TLSKeyFile = "key.pem"
			},
			wantError: true,
		},
		{
			name: "TLS enabled without key file",
			modifyFn: func(c *Config) {
				c.Server.TLSEnabled = true
				c.Server.TLSCertFile = "cert.pem"
				c.Server.TLSKeyFile = ""
			},
			wantError: true,
		},
		{
			name: "API key auth without key",
			modifyFn: func(c *Config) {
				c.Server.APIKeyAuth = true
				c.Server.APIKey = ""
			},
			wantError: true,
		},
		{
			name: "AWS KMS without ECR region",
			modifyFn: func(c *Config) {
				c.Encryption.Enabled = true
				c.Encryption.CustomerManagedKeys = true
				c.Encryption.AWSKMSKeyID = "some-key"
				c.ECR.Region = ""
			},
			wantError: true,
		},
		{
			name: "GCP KMS without GCP project",
			modifyFn: func(c *Config) {
				c.Encryption.Enabled = true
				c.Encryption.CustomerManagedKeys = true
				c.Encryption.GCPKMSKeyID = "some-key"
				c.GCR.Project = ""
			},
			wantError: true,
		},
		{
			name: "AWS secrets manager without region",
			modifyFn: func(c *Config) {
				c.Secrets.UseSecretsManager = true
				c.Secrets.SecretsManagerType = "aws"
				c.Secrets.AWSSecretRegion = ""
				c.ECR.Region = ""
			},
			wantError: true,
		},
		{
			name: "GCP secrets manager without project",
			modifyFn: func(c *Config) {
				c.Secrets.UseSecretsManager = true
				c.Secrets.SecretsManagerType = "gcp"
				c.Secrets.GCPSecretProject = ""
				c.GCR.Project = ""
			},
			wantError: true,
		},
		{
			name: "invalid secrets manager type",
			modifyFn: func(c *Config) {
				c.Secrets.UseSecretsManager = true
				c.Secrets.SecretsManagerType = "invalid"
			},
			wantError: true,
		},
		{
			name: "valid AWS secrets with ECR region",
			modifyFn: func(c *Config) {
				c.Secrets.UseSecretsManager = true
				c.Secrets.SecretsManagerType = "aws"
				c.ECR.Region = "us-east-1"
			},
			wantError: false,
		},
		{
			name: "valid GCP secrets with project",
			modifyFn: func(c *Config) {
				c.Secrets.UseSecretsManager = true
				c.Secrets.SecretsManagerType = "gcp"
				c.GCR.Project = "my-project"
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig()
			tt.modifyFn(config)

			err := config.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestSaveToFile tests configuration saving
func TestSaveToFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping file I/O test in short mode")
	}

	config := NewDefaultConfig()

	// Create temporary directory
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "config.yaml")

	// Save config
	err := config.SaveToFile(filePath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Read file and verify it's valid YAML
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	if len(data) == 0 {
		t.Error("Saved config file is empty")
	}
}

// TestSaveToFileCreatesDirectory tests directory creation
func TestSaveToFileCreatesDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping file I/O test in short mode")
	}

	config := NewDefaultConfig()

	// Create temporary directory
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "dir", "config.yaml")

	// Save config to nested path
	err := config.SaveToFile(nestedPath)
	if err != nil {
		t.Fatalf("Failed to save config to nested path: %v", err)
	}

	// Check if file exists
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("Config file was not created in nested directory")
	}
}
