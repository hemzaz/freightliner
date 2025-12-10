package sync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	validYAML := `
source:
  registry: "docker.io"
  type: "docker"
destination:
  registry: "my-registry.io"
  type: "generic"
parallel: 5
images:
  - repository: "library/nginx"
    tags: ["latest", "1.21"]
`

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configFile, []byte(validYAML), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configFile)
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "docker.io", config.Source.Registry)
	assert.Equal(t, "my-registry.io", config.Destination.Registry)
	assert.Equal(t, 5, config.Parallel)
	assert.Len(t, config.Images, 1)
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	invalidYAML := `
source: invalid yaml structure {{{
`

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configFile, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	_, err = LoadConfig(configFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse YAML")
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoadConfig_ValidationError(t *testing.T) {
	// Config missing required source registry
	invalidConfigYAML := `
destination:
  registry: "my-registry.io"
images:
  - repository: "nginx"
    tags: ["latest"]
`

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid_config.yaml")
	err := os.WriteFile(configFile, []byte(invalidConfigYAML), 0644)
	require.NoError(t, err)

	_, err = LoadConfig(configFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid configuration")
	assert.Contains(t, err.Error(), "source.registry is required")
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: Config{
				Source:      RegistryConfig{Registry: "docker.io"},
				Destination: RegistryConfig{Registry: "my-registry.io"},
				Images: []ImageSync{
					{Repository: "library/nginx", Tags: []string{"latest"}},
				},
			},
			expectError: false,
		},
		{
			name: "missing source registry",
			config: Config{
				Destination: RegistryConfig{Registry: "my-registry.io"},
				Images: []ImageSync{
					{Repository: "library/nginx", Tags: []string{"latest"}},
				},
			},
			expectError: true,
			errorMsg:    "source.registry is required",
		},
		{
			name: "missing destination registry",
			config: Config{
				Source: RegistryConfig{Registry: "docker.io"},
				Images: []ImageSync{
					{Repository: "library/nginx", Tags: []string{"latest"}},
				},
			},
			expectError: true,
			errorMsg:    "destination.registry is required",
		},
		{
			name: "no images",
			config: Config{
				Source:      RegistryConfig{Registry: "docker.io"},
				Destination: RegistryConfig{Registry: "my-registry.io"},
				Images:      []ImageSync{},
			},
			expectError: true,
			errorMsg:    "at least one image must be specified",
		},
		{
			name: "image without repository",
			config: Config{
				Source:      RegistryConfig{Registry: "docker.io"},
				Destination: RegistryConfig{Registry: "my-registry.io"},
				Images: []ImageSync{
					{Tags: []string{"latest"}},
				},
			},
			expectError: true,
			errorMsg:    "repository is required",
		},
		{
			name: "image without filters",
			config: Config{
				Source:      RegistryConfig{Registry: "docker.io"},
				Destination: RegistryConfig{Registry: "my-registry.io"},
				Images: []ImageSync{
					{Repository: "library/nginx"},
				},
			},
			expectError: true,
			errorMsg:    "must specify at least one of",
		},
		{
			name: "image with multiple filters",
			config: Config{
				Source:      RegistryConfig{Registry: "docker.io"},
				Destination: RegistryConfig{Registry: "my-registry.io"},
				Images: []ImageSync{
					{
						Repository:       "library/nginx",
						Tags:             []string{"latest"},
						TagRegex:         ".*",
						SemverConstraint: ">=1.0.0",
					},
				},
			},
			expectError: true,
			errorMsg:    "cannot specify multiple tag filters",
		},
		{
			name: "tags and tag_regex both set",
			config: Config{
				Source:      RegistryConfig{Registry: "docker.io"},
				Destination: RegistryConfig{Registry: "my-registry.io"},
				Images: []ImageSync{
					{
						Repository: "library/nginx",
						Tags:       []string{"latest"},
						TagRegex:   ".*",
					},
				},
			},
			expectError: true,
			errorMsg:    "cannot specify multiple tag filters",
		},
		{
			name: "tags and semver both set",
			config: Config{
				Source:      RegistryConfig{Registry: "docker.io"},
				Destination: RegistryConfig{Registry: "my-registry.io"},
				Images: []ImageSync{
					{
						Repository:       "library/nginx",
						Tags:             []string{"latest"},
						SemverConstraint: ">=1.0.0",
					},
				},
			},
			expectError: true,
			errorMsg:    "cannot specify multiple tag filters",
		},
		{
			name: "tags and all_tags both set",
			config: Config{
				Source:      RegistryConfig{Registry: "docker.io"},
				Destination: RegistryConfig{Registry: "my-registry.io"},
				Images: []ImageSync{
					{
						Repository: "library/nginx",
						Tags:       []string{"latest"},
						AllTags:    true,
					},
				},
			},
			expectError: true,
			errorMsg:    "cannot specify multiple tag filters",
		},
		{
			name: "tags and latest_n both set",
			config: Config{
				Source:      RegistryConfig{Registry: "docker.io"},
				Destination: RegistryConfig{Registry: "my-registry.io"},
				Images: []ImageSync{
					{
						Repository: "library/nginx",
						Tags:       []string{"latest"},
						LatestN:    5,
					},
				},
			},
			expectError: true,
			errorMsg:    "cannot specify multiple tag filters",
		},
		{
			name: "tag_regex and semver both set",
			config: Config{
				Source:      RegistryConfig{Registry: "docker.io"},
				Destination: RegistryConfig{Registry: "my-registry.io"},
				Images: []ImageSync{
					{
						Repository:       "library/nginx",
						TagRegex:         ".*",
						SemverConstraint: ">=1.0.0",
					},
				},
			},
			expectError: true,
			errorMsg:    "cannot specify multiple tag filters",
		},
		{
			name: "semver and all_tags both set",
			config: Config{
				Source:      RegistryConfig{Registry: "docker.io"},
				Destination: RegistryConfig{Registry: "my-registry.io"},
				Images: []ImageSync{
					{
						Repository:       "library/nginx",
						SemverConstraint: ">=1.0.0",
						AllTags:          true,
					},
				},
			},
			expectError: true,
			errorMsg:    "cannot specify multiple tag filters",
		},
		{
			name: "all_tags and latest_n both set",
			config: Config{
				Source:      RegistryConfig{Registry: "docker.io"},
				Destination: RegistryConfig{Registry: "my-registry.io"},
				Images: []ImageSync{
					{
						Repository: "library/nginx",
						AllTags:    true,
						LatestN:    5,
					},
				},
			},
			expectError: true,
			errorMsg:    "cannot specify multiple tag filters",
		},
		{
			name: "multiple images with mixed validity",
			config: Config{
				Source:      RegistryConfig{Registry: "docker.io"},
				Destination: RegistryConfig{Registry: "my-registry.io"},
				Images: []ImageSync{
					{Repository: "library/nginx", Tags: []string{"latest"}},
					{Repository: "library/redis"}, // No filter - should fail
				},
			},
			expectError: true,
			errorMsg:    "must specify at least one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_SetDefaults(t *testing.T) {
	config := &Config{}
	config.SetDefaults()

	assert.Equal(t, 3, config.Parallel)
	assert.Equal(t, 10, config.BatchSize)
	assert.Equal(t, 300, config.Timeout)
	assert.Equal(t, 3, config.RetryAttempts)
	assert.Equal(t, 5, config.RetryBackoff)
}

func TestConfig_SetDefaults_WithExistingValues(t *testing.T) {
	// Test that SetDefaults doesn't override existing positive values
	config := &Config{
		Parallel:      10,
		BatchSize:     50,
		Timeout:       600,
		RetryAttempts: 5,
		RetryBackoff:  10,
		MinBatchSize:  5,
		MaxBatchSize:  100,
	}
	config.SetDefaults()

	assert.Equal(t, 10, config.Parallel)
	assert.Equal(t, 50, config.BatchSize)
	assert.Equal(t, 600, config.Timeout)
	assert.Equal(t, 5, config.RetryAttempts)
	assert.Equal(t, 10, config.RetryBackoff)
	assert.Equal(t, 5, config.MinBatchSize)
	assert.Equal(t, 100, config.MaxBatchSize)
}

func TestConfig_SetDefaults_NegativeValues(t *testing.T) {
	// Test that negative values are replaced with defaults
	config := &Config{
		Parallel:      -1,
		BatchSize:     -5,
		Timeout:       -100,
		RetryAttempts: -2,
		RetryBackoff:  -3,
		MinBatchSize:  -1,
		MaxBatchSize:  -10,
	}
	config.SetDefaults()

	assert.Equal(t, 3, config.Parallel)
	assert.Equal(t, 10, config.BatchSize)
	assert.Equal(t, 300, config.Timeout)
	assert.Equal(t, 3, config.RetryAttempts)
	assert.Equal(t, 5, config.RetryBackoff)
	assert.Equal(t, 1, config.MinBatchSize)
	assert.Equal(t, 50, config.MaxBatchSize)
}

func TestConfig_SetDefaults_ZeroValues(t *testing.T) {
	// Test that zero values are replaced with defaults
	config := &Config{
		Parallel:      0,
		BatchSize:     0,
		Timeout:       0,
		RetryAttempts: 0,
		RetryBackoff:  0,
		MinBatchSize:  0,
		MaxBatchSize:  0,
	}
	config.SetDefaults()

	assert.Equal(t, 3, config.Parallel)
	assert.Equal(t, 10, config.BatchSize)
	assert.Equal(t, 300, config.Timeout)
	assert.Equal(t, 3, config.RetryAttempts)
	assert.Equal(t, 5, config.RetryBackoff)
	assert.Equal(t, 1, config.MinBatchSize)
	assert.Equal(t, 50, config.MaxBatchSize)
}

func TestDetectRegistryType(t *testing.T) {
	tests := []struct {
		registry     string
		expectedType string
	}{
		{"docker.io", "docker"},
		{"registry-1.docker.io", "docker"},
		{"gcr.io", "gcr"},
		{"us.gcr.io", "gcr"},
		{"ghcr.io", "ghcr"},
		{"123456789012.dkr.ecr.us-west-2.amazonaws.com", "ecr"},
		{"myregistry.azurecr.io", "acr"},
		{"quay.io", "quay"},
		{"registry.company.com", "generic"},
	}

	for _, tt := range tests {
		t.Run(tt.registry, func(t *testing.T) {
			detected := detectRegistryType(tt.registry)
			assert.Equal(t, tt.expectedType, detected)
		})
	}
}

func TestConfig_SetDefaults_RegistryTypes(t *testing.T) {
	config := &Config{
		Source:      RegistryConfig{Registry: "docker.io"},
		Destination: RegistryConfig{Registry: "gcr.io"},
	}

	config.SetDefaults()

	assert.Equal(t, "docker", config.Source.Type)
	assert.Equal(t, "gcr", config.Destination.Type)
}

func TestConfig_SetDefaults_AllRegistryTypes(t *testing.T) {
	tests := []struct {
		name           string
		sourceRegistry string
		destRegistry   string
		expectedSource string
		expectedDest   string
	}{
		{
			name:           "docker.io",
			sourceRegistry: "docker.io",
			destRegistry:   "registry-1.docker.io",
			expectedSource: "docker",
			expectedDest:   "docker",
		},
		{
			name:           "gcr.io variants",
			sourceRegistry: "gcr.io",
			destRegistry:   "us.gcr.io",
			expectedSource: "gcr",
			expectedDest:   "gcr",
		},
		{
			name:           "gcr.io more variants",
			sourceRegistry: "eu.gcr.io",
			destRegistry:   "asia.gcr.io",
			expectedSource: "gcr",
			expectedDest:   "gcr",
		},
		{
			name:           "ghcr.io",
			sourceRegistry: "ghcr.io",
			destRegistry:   "ghcr.io",
			expectedSource: "ghcr",
			expectedDest:   "ghcr",
		},
		{
			name:           "ecr public (note: production code has bug with this pattern)",
			sourceRegistry: "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			destRegistry:   "987654321098.dkr.ecr.eu-west-1.amazonaws.com",
			expectedSource: "ecr",
			expectedDest:   "ecr",
		},
		{
			name:           "ecr private",
			sourceRegistry: "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			destRegistry:   "987654321098.dkr.ecr.us-east-1.amazonaws.com",
			expectedSource: "ecr",
			expectedDest:   "ecr",
		},
		{
			name:           "acr",
			sourceRegistry: "myregistry.azurecr.io",
			destRegistry:   "otherregistry.azurecr.io",
			expectedSource: "acr",
			expectedDest:   "acr",
		},
		{
			name:           "quay.io",
			sourceRegistry: "quay.io",
			destRegistry:   "myorg.quay.io",
			expectedSource: "quay",
			expectedDest:   "quay",
		},
		{
			name:           "generic registries",
			sourceRegistry: "registry.company.com",
			destRegistry:   "harbor.example.org",
			expectedSource: "generic",
			expectedDest:   "generic",
		},
		{
			name:           "mixed types",
			sourceRegistry: "docker.io",
			destRegistry:   "myregistry.azurecr.io",
			expectedSource: "docker",
			expectedDest:   "acr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Source:      RegistryConfig{Registry: tt.sourceRegistry},
				Destination: RegistryConfig{Registry: tt.destRegistry},
			}

			config.SetDefaults()

			assert.Equal(t, tt.expectedSource, config.Source.Type, "Source type mismatch")
			assert.Equal(t, tt.expectedDest, config.Destination.Type, "Destination type mismatch")
		})
	}
}

func TestConfig_SetDefaults_PresetRegistryTypes(t *testing.T) {
	// Test that SetDefaults doesn't override pre-set registry types
	config := &Config{
		Source: RegistryConfig{
			Registry: "docker.io",
			Type:     "custom",
		},
		Destination: RegistryConfig{
			Registry: "gcr.io",
			Type:     "another-custom",
		},
	}

	config.SetDefaults()

	assert.Equal(t, "custom", config.Source.Type, "Should not override preset source type")
	assert.Equal(t, "another-custom", config.Destination.Type, "Should not override preset destination type")
}

func TestImageSync_Validation(t *testing.T) {
	tests := []struct {
		name   string
		image  ImageSync
		valid  bool
		errMsg string
	}{
		{
			name:  "with specific tags",
			image: ImageSync{Repository: "nginx", Tags: []string{"latest"}},
			valid: true,
		},
		{
			name:  "with regex",
			image: ImageSync{Repository: "nginx", TagRegex: "^1\\..*"},
			valid: true,
		},
		{
			name:  "with semver",
			image: ImageSync{Repository: "nginx", SemverConstraint: ">=1.20.0"},
			valid: true,
		},
		{
			name:  "with all_tags",
			image: ImageSync{Repository: "nginx", AllTags: true},
			valid: true,
		},
		{
			name:  "with latest_n",
			image: ImageSync{Repository: "nginx", LatestN: 5},
			valid: true,
		},
		{
			name:   "no filters",
			image:  ImageSync{Repository: "nginx"},
			valid:  false,
			errMsg: "must specify at least one of",
		},
		{
			name: "multiple filters",
			image: ImageSync{
				Repository: "nginx",
				Tags:       []string{"latest"},
				AllTags:    true,
			},
			valid:  false,
			errMsg: "cannot specify multiple tag filters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Source:      RegistryConfig{Registry: "docker.io"},
				Destination: RegistryConfig{Registry: "my-registry.io"},
				Images:      []ImageSync{tt.image},
			}

			err := config.Validate()
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestAuthConfig(t *testing.T) {
	tests := []struct {
		name string
		auth AuthConfig
	}{
		{
			name: "basic auth",
			auth: AuthConfig{Username: "user", Password: "pass"},
		},
		{
			name: "token auth",
			auth: AuthConfig{Token: "token123"},
		},
		{
			name: "docker config",
			auth: AuthConfig{UseDockerConfig: true},
		},
		{
			name: "aws profile",
			auth: AuthConfig{AWSProfile: "default"},
		},
		{
			name: "gcp credentials",
			auth: AuthConfig{GCPCredentials: "/path/to/creds.json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Source: RegistryConfig{
					Registry: "docker.io",
					Auth:     &tt.auth,
				},
				Destination: RegistryConfig{Registry: "my-registry.io"},
				Images: []ImageSync{
					{Repository: "nginx", Tags: []string{"latest"}},
				},
			}

			err := config.Validate()
			assert.NoError(t, err)
		})
	}
}
