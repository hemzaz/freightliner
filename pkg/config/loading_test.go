package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestLoadFromFile tests configuration loading from file
func TestLoadFromFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping file I/O test in short mode")
	}

	tests := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name: "valid config",
			content: `
loglevel: debug
ecr:
  region: us-east-1
  accountid: "123456789012"
server:
  port: 9090
  host: 0.0.0.0
`,
			wantError: false,
		},
		{
			name:      "empty file",
			content:   "",
			wantError: false, // Should use defaults
		},
		{
			name: "invalid yaml",
			content: `
invalid: [yaml
  missing: bracket
`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			// Always write the file, even if content is empty
			err := os.WriteFile(configPath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			config, err := LoadFromFile(configPath)
			if (err != nil) != tt.wantError {
				t.Errorf("LoadFromFile() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && config == nil {
				t.Error("Expected config to be non-nil")
			}
		})
	}
}

// TestLoadFromFileNotFound tests loading non-existent file
func TestLoadFromFileNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping file I/O test in short mode")
	}

	_, err := LoadFromFile("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// TestLoadFromFileEmpty tests loading with empty path
func TestLoadFromFileEmpty(t *testing.T) {
	config, err := LoadFromFile("")
	if err != nil {
		t.Fatalf("LoadFromFile(\"\") failed: %v", err)
	}

	if config == nil {
		t.Error("Expected default config for empty path")
	}
}

// TestLoadFromEnv tests environment variable loading
func TestLoadFromEnv(t *testing.T) {
	// Save original env vars
	originalEnv := make(map[string]string)
	envVars := []string{
		"FREIGHTLINER_LOG_LEVEL",
		"FREIGHTLINER_ECR_REGION",
		"FREIGHTLINER_SERVER_PORT",
		"FREIGHTLINER_REPLICATE_WORKERS",
		"FREIGHTLINER_ENCRYPTION_ENABLED",
		"FREIGHTLINER_TLS_ENABLED",
		"FREIGHTLINER_SERVER_READ_TIMEOUT",
		"FREIGHTLINER_SERVER_ALLOWED_ORIGINS",
		"FREIGHTLINER_TREE_EXCLUDE_REPOS",
	}

	for _, env := range envVars {
		originalEnv[env] = os.Getenv(env)
	}

	// Clean up after test
	defer func() {
		for _, env := range envVars {
			if val, exists := originalEnv[env]; exists {
				os.Setenv(env, val)
			} else {
				os.Unsetenv(env)
			}
		}
	}()

	// Set test environment variables
	os.Setenv("FREIGHTLINER_LOG_LEVEL", "debug")
	os.Setenv("FREIGHTLINER_ECR_REGION", "eu-west-1")
	os.Setenv("FREIGHTLINER_SERVER_PORT", "9090")
	os.Setenv("FREIGHTLINER_REPLICATE_WORKERS", "10")
	os.Setenv("FREIGHTLINER_ENCRYPTION_ENABLED", "true")
	os.Setenv("FREIGHTLINER_TLS_ENABLED", "yes")
	os.Setenv("FREIGHTLINER_SERVER_READ_TIMEOUT", "45s")
	os.Setenv("FREIGHTLINER_SERVER_ALLOWED_ORIGINS", "http://localhost,https://example.com")
	os.Setenv("FREIGHTLINER_TREE_EXCLUDE_REPOS", "test-*,tmp-*")

	config := NewDefaultConfig()
	err := loadFromEnv(config)
	if err != nil {
		t.Fatalf("loadFromEnv() failed: %v", err)
	}

	// Test string values
	if config.LogLevel != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", config.LogLevel)
	}
	if config.ECR.Region != "eu-west-1" {
		t.Errorf("Expected ECR region 'eu-west-1', got '%s'", config.ECR.Region)
	}

	// Test integer values
	if config.Server.Port != 9090 {
		t.Errorf("Expected server port 9090, got %d", config.Server.Port)
	}
	if config.Workers.ReplicateWorkers != 10 {
		t.Errorf("Expected replicate workers 10, got %d", config.Workers.ReplicateWorkers)
	}

	// Test boolean values
	if !config.Encryption.Enabled {
		t.Error("Expected encryption enabled to be true")
	}
	if !config.Server.TLSEnabled {
		t.Error("Expected TLS enabled to be true")
	}

	// Test duration values
	if config.Server.ReadTimeout != 45*time.Second {
		t.Errorf("Expected read timeout 45s, got %v", config.Server.ReadTimeout)
	}

	// Test slice values
	if len(config.Server.AllowedOrigins) != 2 {
		t.Errorf("Expected 2 allowed origins, got %d", len(config.Server.AllowedOrigins))
	}
	if len(config.TreeReplicate.ExcludeRepos) != 2 {
		t.Errorf("Expected 2 exclude repos, got %d", len(config.TreeReplicate.ExcludeRepos))
	}
}

// TestProcessBooleanEnvVars tests boolean environment variable parsing
func TestProcessBooleanEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"true lowercase", "true", true},
		{"TRUE uppercase", "TRUE", true},
		{"1 numeric", "1", true},
		{"yes", "yes", true},
		{"y single char", "y", true},
		{"false", "false", false},
		{"0 numeric", "0", false},
		{"no", "no", false},
		{"n single char", "n", false},
		{"empty", "", false},
		{"invalid", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig()
			envVar := "FREIGHTLINER_ENCRYPTION_ENABLED"

			// Save and restore
			original := os.Getenv(envVar)
			defer func() {
				if original != "" {
					os.Setenv(envVar, original)
				} else {
					os.Unsetenv(envVar)
				}
			}()

			os.Setenv(envVar, tt.envValue)
			processBooleanEnvVars(config)

			if config.Encryption.Enabled != tt.expected {
				t.Errorf("For value '%s', expected %v, got %v", tt.envValue, tt.expected, config.Encryption.Enabled)
			}
		})
	}
}

// TestProcessIntEnvVars tests integer environment variable parsing
func TestProcessIntEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected int
	}{
		{"valid integer", "42", 42},
		{"zero", "0", 0},
		{"negative", "-10", -10},
		{"invalid", "invalid", 0}, // Should use default (0)
		{"empty", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig()
			config.Server.Port = 0 // Reset to zero
			envVar := "FREIGHTLINER_SERVER_PORT"

			// Save and restore
			original := os.Getenv(envVar)
			defer func() {
				if original != "" {
					os.Setenv(envVar, original)
				} else {
					os.Unsetenv(envVar)
				}
			}()

			if tt.envValue != "" {
				os.Setenv(envVar, tt.envValue)
			} else {
				os.Unsetenv(envVar)
			}

			processIntEnvVars(config)

			if config.Server.Port != tt.expected {
				t.Errorf("For value '%s', expected %d, got %d", tt.envValue, tt.expected, config.Server.Port)
			}
		})
	}
}

// TestProcessSliceEnvVars tests slice environment variable parsing
func TestProcessSliceEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected int
	}{
		{"single value", "value1", 1},
		{"multiple values", "value1,value2,value3", 3},
		{"values with spaces", "value1 , value2 , value3", 3},
		{"empty value", "", 0},
		{"only commas", ",,,", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig()
			envVar := "FREIGHTLINER_TREE_EXCLUDE_REPOS"

			// Save and restore
			original := os.Getenv(envVar)
			defer func() {
				if original != "" {
					os.Setenv(envVar, original)
				} else {
					os.Unsetenv(envVar)
				}
			}()

			if tt.envValue != "" {
				os.Setenv(envVar, tt.envValue)
			} else {
				os.Unsetenv(envVar)
			}

			processSliceEnvVars(config)

			if len(config.TreeReplicate.ExcludeRepos) != tt.expected {
				t.Errorf("For value '%s', expected %d items, got %d", tt.envValue, tt.expected, len(config.TreeReplicate.ExcludeRepos))
			}
		})
	}
}

// TestProcessDurationEnvVars tests duration environment variable parsing
func TestProcessDurationEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected time.Duration
	}{
		{"seconds", "30s", 30 * time.Second},
		{"minutes", "5m", 5 * time.Minute},
		{"hours", "2h", 2 * time.Hour},
		{"combined", "1h30m", 90 * time.Minute},
		{"invalid", "invalid", 30 * time.Second}, // Should keep default
		{"empty", "", 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDefaultConfig()
			envVar := "FREIGHTLINER_SERVER_READ_TIMEOUT"

			// Save and restore
			original := os.Getenv(envVar)
			defer func() {
				if original != "" {
					os.Setenv(envVar, original)
				} else {
					os.Unsetenv(envVar)
				}
			}()

			if tt.envValue != "" {
				os.Setenv(envVar, tt.envValue)
			} else {
				os.Unsetenv(envVar)
			}

			processDurationEnvVars(config)

			if config.Server.ReadTimeout != tt.expected {
				t.Errorf("For value '%s', expected %v, got %v", tt.envValue, tt.expected, config.Server.ReadTimeout)
			}
		})
	}
}

// TestProcessStringEnvVars tests string environment variable processing
func TestProcessStringEnvVars(t *testing.T) {
	config := NewDefaultConfig()
	envVar := "FREIGHTLINER_LOG_LEVEL"

	// Save and restore
	original := os.Getenv(envVar)
	defer func() {
		if original != "" {
			os.Setenv(envVar, original)
		} else {
			os.Unsetenv(envVar)
		}
	}()

	// Test setting value
	os.Setenv(envVar, "debug")
	processStringEnvVars(config)

	if config.LogLevel != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", config.LogLevel)
	}

	// Test empty value doesn't override
	os.Setenv(envVar, "")
	processStringEnvVars(config)

	if config.LogLevel != "debug" {
		t.Errorf("Expected log level to remain 'debug', got '%s'", config.LogLevel)
	}
}
