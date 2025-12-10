package config

import (
	"testing"
)

func TestRegistryConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  RegistryConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid ECR config",
			config: RegistryConfig{
				Name:   "test-ecr",
				Type:   RegistryTypeECR,
				Region: "us-east-1",
				Auth: AuthConfig{
					Type: AuthTypeAWS,
				},
			},
			wantErr: false,
		},
		{
			name: "valid GCR config",
			config: RegistryConfig{
				Name:    "test-gcr",
				Type:    RegistryTypeGCR,
				Project: "my-project",
				Auth: AuthConfig{
					Type: AuthTypeGCP,
				},
			},
			wantErr: false,
		},
		{
			name: "valid DockerHub config",
			config: RegistryConfig{
				Name:     "test-dockerhub",
				Type:     RegistryTypeDockerHub,
				Endpoint: "https://registry-1.docker.io",
				Auth: AuthConfig{
					Type:     AuthTypeBasic,
					Username: "testuser",
					Password: "testpass",
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			config: RegistryConfig{
				Type:   RegistryTypeECR,
				Region: "us-east-1",
			},
			wantErr: true,
			errMsg:  "registry name is required",
		},
		{
			name: "missing type",
			config: RegistryConfig{
				Name: "test",
			},
			wantErr: true,
			errMsg:  "registry type is required",
		},
		{
			name: "ECR without region",
			config: RegistryConfig{
				Name: "test-ecr",
				Type: RegistryTypeECR,
			},
			wantErr: true,
			errMsg:  "region is required for ECR",
		},
		{
			name: "GCR without project",
			config: RegistryConfig{
				Name: "test-gcr",
				Type: RegistryTypeGCR,
			},
			wantErr: true,
			errMsg:  "project is required for GCR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("RegistryConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg && !contains(err.Error(), tt.errMsg) {
					t.Errorf("RegistryConfig.Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestAuthConfig_Validate(t *testing.T) {
	tests := []struct {
		name         string
		authConfig   AuthConfig
		registryType RegistryType
		wantErr      bool
		wantAuthType AuthType
	}{
		{
			name: "basic auth with credentials",
			authConfig: AuthConfig{
				Type:     AuthTypeBasic,
				Username: "testuser",
				Password: "testpass",
			},
			registryType: RegistryTypeDockerHub,
			wantErr:      false,
			wantAuthType: AuthTypeBasic,
		},
		{
			name: "AWS auth for ECR",
			authConfig: AuthConfig{
				Type: AuthTypeAWS,
			},
			registryType: RegistryTypeECR,
			wantErr:      false,
			wantAuthType: AuthTypeAWS,
		},
		{
			name:       "auto-detect AWS for ECR",
			authConfig: AuthConfig{
				// No type specified, should auto-detect
			},
			registryType: RegistryTypeECR,
			wantErr:      false,
			wantAuthType: AuthTypeAWS,
		},
		{
			name:       "auto-detect GCP for GCR",
			authConfig: AuthConfig{
				// No type specified, should auto-detect
			},
			registryType: RegistryTypeGCR,
			wantErr:      false,
			wantAuthType: AuthTypeGCP,
		},
		{
			name: "basic auth without username",
			authConfig: AuthConfig{
				Type:     AuthTypeBasic,
				Password: "testpass",
			},
			registryType: RegistryTypeDockerHub,
			wantErr:      true,
		},
		{
			name: "token auth without token",
			authConfig: AuthConfig{
				Type: AuthTypeToken,
			},
			registryType: RegistryTypeQuay,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.authConfig.Validate(tt.registryType)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.wantAuthType != "" {
				if tt.authConfig.Type != tt.wantAuthType {
					t.Errorf("AuthConfig.Validate() authType = %v, want %v", tt.authConfig.Type, tt.wantAuthType)
				}
			}
		})
	}
}

func TestRegistryConfig_GetDefaultEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		config   RegistryConfig
		expected string
	}{
		{
			name: "DockerHub default",
			config: RegistryConfig{
				Type: RegistryTypeDockerHub,
			},
			expected: "https://registry-1.docker.io",
		},
		{
			name: "Quay default",
			config: RegistryConfig{
				Type: RegistryTypeQuay,
			},
			expected: "https://quay.io",
		},
		{
			name: "GitHub default",
			config: RegistryConfig{
				Type: RegistryTypeGitHub,
			},
			expected: "https://ghcr.io",
		},
		{
			name: "ECR with region and account",
			config: RegistryConfig{
				Type:      RegistryTypeECR,
				Region:    "us-east-1",
				AccountID: "123456789012",
			},
			expected: "https://123456789012.dkr.ecr.us-east-1.amazonaws.com",
		},
		{
			name: "GCR with us location",
			config: RegistryConfig{
				Type:   RegistryTypeGCR,
				Region: "us",
			},
			expected: "https://us.gcr.io",
		},
		{
			name: "GCR default location",
			config: RegistryConfig{
				Type: RegistryTypeGCR,
			},
			expected: "https://us.gcr.io",
		},
		{
			name: "Azure ACR",
			config: RegistryConfig{
				Type:    RegistryTypeAzure,
				Project: "myregistry",
			},
			expected: "https://myregistry.azurecr.io",
		},
		{
			name: "Custom endpoint",
			config: RegistryConfig{
				Type:     RegistryTypeGeneric,
				Endpoint: "https://custom.registry.io",
			},
			expected: "https://custom.registry.io",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetDefaultEndpoint()
			if result != tt.expected {
				t.Errorf("GetDefaultEndpoint() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRegistryConfig_GetRegistryHost(t *testing.T) {
	tests := []struct {
		name     string
		config   RegistryConfig
		expected string
		wantErr  bool
	}{
		{
			name: "DockerHub host",
			config: RegistryConfig{
				Type:     RegistryTypeDockerHub,
				Endpoint: "https://registry-1.docker.io",
			},
			expected: "registry-1.docker.io",
			wantErr:  false,
		},
		{
			name: "ECR host",
			config: RegistryConfig{
				Type:      RegistryTypeECR,
				Region:    "us-east-1",
				AccountID: "123456789012",
			},
			expected: "123456789012.dkr.ecr.us-east-1.amazonaws.com",
			wantErr:  false,
		},
		{
			name: "invalid endpoint",
			config: RegistryConfig{
				Type:     RegistryTypeGeneric,
				Endpoint: "://invalid-url",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.config.GetRegistryHost()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRegistryHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("GetRegistryHost() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRegistryConfig_GetImageReference(t *testing.T) {
	tests := []struct {
		name       string
		config     RegistryConfig
		repository string
		tag        string
		expected   string
		wantErr    bool
	}{
		{
			name: "DockerHub with tag",
			config: RegistryConfig{
				Type:     RegistryTypeDockerHub,
				Endpoint: "https://registry-1.docker.io",
			},
			repository: "library/nginx",
			tag:        "latest",
			expected:   "registry-1.docker.io/library/nginx:latest",
			wantErr:    false,
		},
		{
			name: "without tag",
			config: RegistryConfig{
				Type:     RegistryTypeDockerHub,
				Endpoint: "https://registry-1.docker.io",
			},
			repository: "library/nginx",
			tag:        "",
			expected:   "registry-1.docker.io/library/nginx",
			wantErr:    false,
		},
		{
			name: "repository with leading slash",
			config: RegistryConfig{
				Type:     RegistryTypeDockerHub,
				Endpoint: "https://registry-1.docker.io",
			},
			repository: "/library/nginx",
			tag:        "1.21",
			expected:   "registry-1.docker.io/library/nginx:1.21",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.config.GetImageReference(tt.repository, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetImageReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("GetImageReference() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRegistriesConfig_GetByName(t *testing.T) {
	config := &RegistriesConfig{
		Registries: []RegistryConfig{
			{Name: "ecr-prod", Type: RegistryTypeECR, Region: "us-east-1"},
			{Name: "gcr-dev", Type: RegistryTypeGCR, Project: "dev-project"},
		},
	}

	tests := []struct {
		name     string
		findName string
		wantErr  bool
	}{
		{
			name:     "find existing ECR",
			findName: "ecr-prod",
			wantErr:  false,
		},
		{
			name:     "find existing GCR",
			findName: "gcr-dev",
			wantErr:  false,
		},
		{
			name:     "find non-existing",
			findName: "not-found",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := config.GetByName(tt.findName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.Name != tt.findName {
				t.Errorf("GetByName() = %v, want %v", result.Name, tt.findName)
			}
		})
	}
}

func TestRegistriesConfig_GetByType(t *testing.T) {
	config := &RegistriesConfig{
		Registries: []RegistryConfig{
			{Name: "ecr-prod", Type: RegistryTypeECR, Region: "us-east-1"},
			{Name: "ecr-dev", Type: RegistryTypeECR, Region: "us-west-2"},
			{Name: "gcr-prod", Type: RegistryTypeGCR, Project: "prod-project"},
		},
	}

	tests := []struct {
		name         string
		registryType RegistryType
		expectedLen  int
	}{
		{
			name:         "find ECR registries",
			registryType: RegistryTypeECR,
			expectedLen:  2,
		},
		{
			name:         "find GCR registries",
			registryType: RegistryTypeGCR,
			expectedLen:  1,
		},
		{
			name:         "find non-existing type",
			registryType: RegistryTypeDockerHub,
			expectedLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.GetByType(tt.registryType)
			if len(result) != tt.expectedLen {
				t.Errorf("GetByType() returned %d registries, want %d", len(result), tt.expectedLen)
			}
		})
	}
}

func TestRegistriesConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  RegistriesConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: RegistriesConfig{
				DefaultSource:      "ecr-prod",
				DefaultDestination: "gcr-prod",
				Registries: []RegistryConfig{
					{Name: "ecr-prod", Type: RegistryTypeECR, Region: "us-east-1"},
					{Name: "gcr-prod", Type: RegistryTypeGCR, Project: "prod-project"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty registries",
			config: RegistriesConfig{
				Registries: []RegistryConfig{},
			},
			wantErr: true,
			errMsg:  "no registries configured",
		},
		{
			name: "duplicate names",
			config: RegistriesConfig{
				Registries: []RegistryConfig{
					{Name: "my-registry", Type: RegistryTypeECR, Region: "us-east-1"},
					{Name: "my-registry", Type: RegistryTypeGCR, Project: "my-project"},
				},
			},
			wantErr: true,
			errMsg:  "duplicate registry name",
		},
		{
			name: "default source not found",
			config: RegistriesConfig{
				DefaultSource: "non-existing",
				Registries: []RegistryConfig{
					{Name: "ecr-prod", Type: RegistryTypeECR, Region: "us-east-1"},
				},
			},
			wantErr: true,
			errMsg:  "default source registry not found",
		},
		{
			name: "default destination not found",
			config: RegistriesConfig{
				DefaultDestination: "non-existing",
				Registries: []RegistryConfig{
					{Name: "ecr-prod", Type: RegistryTypeECR, Region: "us-east-1"},
				},
			},
			wantErr: true,
			errMsg:  "default destination registry not found",
		},
		{
			name: "invalid registry config",
			config: RegistriesConfig{
				Registries: []RegistryConfig{
					{Name: "invalid", Type: RegistryTypeECR}, // Missing region
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("RegistriesConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("RegistriesConfig.Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestCreateRegistryConfigFromLegacy(t *testing.T) {
	ecr := ECRConfig{
		Region:    "us-east-1",
		AccountID: "123456789012",
	}
	gcr := GCRConfig{
		Project:  "my-project",
		Location: "us",
	}

	config := CreateRegistryConfigFromLegacy(ecr, gcr)

	if len(config.Registries) != 2 {
		t.Errorf("CreateRegistryConfigFromLegacy() created %d registries, want 2", len(config.Registries))
	}

	// Check ECR registry
	ecrReg, err := config.GetByName("default-ecr")
	if err != nil {
		t.Errorf("ECR registry not found: %v", err)
	} else {
		if ecrReg.Type != RegistryTypeECR {
			t.Errorf("ECR registry type = %v, want %v", ecrReg.Type, RegistryTypeECR)
		}
		if ecrReg.Region != ecr.Region {
			t.Errorf("ECR region = %v, want %v", ecrReg.Region, ecr.Region)
		}
		if ecrReg.AccountID != ecr.AccountID {
			t.Errorf("ECR accountID = %v, want %v", ecrReg.AccountID, ecr.AccountID)
		}
	}

	// Check GCR registry
	gcrReg, err := config.GetByName("default-gcr")
	if err != nil {
		t.Errorf("GCR registry not found: %v", err)
	} else {
		if gcrReg.Type != RegistryTypeGCR {
			t.Errorf("GCR registry type = %v, want %v", gcrReg.Type, RegistryTypeGCR)
		}
		if gcrReg.Project != gcr.Project {
			t.Errorf("GCR project = %v, want %v", gcrReg.Project, gcr.Project)
		}
	}

	// Check defaults
	if config.DefaultSource != "default-ecr" {
		t.Errorf("DefaultSource = %v, want default-ecr", config.DefaultSource)
	}
	if config.DefaultDestination != "default-gcr" {
		t.Errorf("DefaultDestination = %v, want default-gcr", config.DefaultDestination)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
