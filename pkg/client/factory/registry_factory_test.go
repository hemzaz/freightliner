package factory

import (
	"context"
	"testing"

	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"
)

func TestNewRegistryClientFactory(t *testing.T) {
	logger := log.NewLogger()
	factory := NewRegistryClientFactory(logger)

	if factory == nil {
		t.Fatal("NewRegistryClientFactory() returned nil")
	}

	if factory.logger == nil {
		t.Error("Factory logger is nil")
	}
}

func TestNewRegistryClientFactory_NilLogger(t *testing.T) {
	factory := NewRegistryClientFactory(nil)

	if factory == nil {
		t.Fatal("NewRegistryClientFactory() returned nil")
	}

	if factory.logger == nil {
		t.Error("Factory should create default logger when nil is passed")
	}
}

func TestRegistryClientFactory_GetSupportedRegistryTypes(t *testing.T) {
	factory := NewRegistryClientFactory(log.NewLogger())
	types := factory.GetSupportedRegistryTypes()

	if len(types) == 0 {
		t.Error("GetSupportedRegistryTypes() returned empty list")
	}

	// Check that key types are present
	expectedTypes := []config.RegistryType{
		config.RegistryTypeECR,
		config.RegistryTypeGCR,
		config.RegistryTypeDockerHub,
	}

	for _, expected := range expectedTypes {
		found := false
		for _, t := range types {
			if t == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected registry type %s not found in supported types", expected)
		}
	}
}

func TestRegistryClientFactory_IsRegistryTypeSupported(t *testing.T) {
	factory := NewRegistryClientFactory(log.NewLogger())

	tests := []struct {
		name         string
		registryType config.RegistryType
		want         bool
	}{
		{
			name:         "ECR is supported",
			registryType: config.RegistryTypeECR,
			want:         true,
		},
		{
			name:         "GCR is supported",
			registryType: config.RegistryTypeGCR,
			want:         true,
		},
		{
			name:         "DockerHub is supported",
			registryType: config.RegistryTypeDockerHub,
			want:         true,
		},
		{
			name:         "Invalid type is not supported",
			registryType: config.RegistryType("invalid"),
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := factory.IsRegistryTypeSupported(tt.registryType)
			if got != tt.want {
				t.Errorf("IsRegistryTypeSupported(%v) = %v, want %v", tt.registryType, got, tt.want)
			}
		})
	}
}

func TestRegistryClientFactory_CreateClient_NilConfig(t *testing.T) {
	factory := NewRegistryClientFactory(log.NewLogger())
	ctx := context.Background()

	_, err := factory.CreateClient(ctx, nil)
	if err == nil {
		t.Error("CreateClient() with nil config should return error")
	}
}

func TestRegistryClientFactory_CreateClient_InvalidConfig(t *testing.T) {
	factory := NewRegistryClientFactory(log.NewLogger())
	ctx := context.Background()

	// Config without name
	invalidConfig := &config.RegistryConfig{
		Type: config.RegistryTypeECR,
	}

	_, err := factory.CreateClient(ctx, invalidConfig)
	if err == nil {
		t.Error("CreateClient() with invalid config should return error")
	}
}

func TestRegistryClientFactory_CreateClient_UnsupportedType(t *testing.T) {
	factory := NewRegistryClientFactory(log.NewLogger())
	ctx := context.Background()

	unsupportedConfig := &config.RegistryConfig{
		Name: "test",
		Type: config.RegistryType("unsupported"),
	}

	_, err := factory.CreateClient(ctx, unsupportedConfig)
	if err == nil {
		t.Error("CreateClient() with unsupported type should return error")
	}
}

func TestRegistryClientFactory_CreateClientByName(t *testing.T) {
	factory := NewRegistryClientFactory(log.NewLogger())
	ctx := context.Background()

	registriesConfig := &config.RegistriesConfig{
		Registries: []config.RegistryConfig{
			{
				Name:   "test-ecr",
				Type:   config.RegistryTypeECR,
				Region: "us-east-1",
			},
		},
	}

	tests := []struct {
		name         string
		registryName string
		wantErr      bool
	}{
		{
			name:         "existing registry",
			registryName: "test-ecr",
			wantErr:      false, // Will fail in actual creation due to AWS credentials, but should pass lookup
		},
		{
			name:         "non-existing registry",
			registryName: "not-found",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := factory.CreateClientByName(ctx, tt.registryName, registriesConfig)
			// We expect errors for both cases in unit tests (no AWS credentials)
			// But the error messages should be different
			if err == nil && tt.wantErr {
				t.Error("CreateClientByName() expected error but got none")
			}
			if err != nil && !tt.wantErr {
				// Check if it's a "not found" error vs. creation error
				if !contains(err.Error(), "failed to find registry") {
					// It's a creation error, which is expected in unit tests
					// This test validates the lookup logic works
					t.Logf("CreateClientByName() error = %v (expected in unit test without credentials)", err)
				} else {
					t.Errorf("CreateClientByName() unexpected not found error = %v", err)
				}
			}
		})
	}
}

func TestRegistryClientFactory_CreateClientByName_NilConfig(t *testing.T) {
	factory := NewRegistryClientFactory(log.NewLogger())
	ctx := context.Background()

	_, err := factory.CreateClientByName(ctx, "test", nil)
	if err == nil {
		t.Error("CreateClientByName() with nil config should return error")
	}
}

func TestRegistryClientFactory_CreateSourceAndDestClients_MissingRegistries(t *testing.T) {
	factory := NewRegistryClientFactory(log.NewLogger())
	ctx := context.Background()

	registriesConfig := &config.RegistriesConfig{
		Registries: []config.RegistryConfig{},
	}

	tests := []struct {
		name           string
		sourceRegistry string
		destRegistry   string
		wantErr        bool
		errContains    string
	}{
		{
			name:           "both missing",
			sourceRegistry: "",
			destRegistry:   "",
			wantErr:        true,
			errContains:    "source registry not specified",
		},
		{
			name:           "destination missing",
			sourceRegistry: "source",
			destRegistry:   "",
			wantErr:        true,
			errContains:    "destination registry not specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := factory.CreateSourceAndDestClients(ctx, tt.sourceRegistry, tt.destRegistry, registriesConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSourceAndDestClients() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("CreateSourceAndDestClients() error = %v, want error containing %v", err, tt.errContains)
				}
			}
		})
	}
}

func TestRegistryClientFactory_CreateSourceAndDestClients_WithDefaults(t *testing.T) {
	factory := NewRegistryClientFactory(log.NewLogger())
	ctx := context.Background()

	registriesConfig := &config.RegistriesConfig{
		DefaultSource:      "source-ecr",
		DefaultDestination: "dest-gcr",
		Registries: []config.RegistryConfig{
			{
				Name:   "source-ecr",
				Type:   config.RegistryTypeECR,
				Region: "us-east-1",
			},
			{
				Name:    "dest-gcr",
				Type:    config.RegistryTypeGCR,
				Project: "my-project",
			},
		},
	}

	// Test with empty registry names - should use defaults
	_, _, err := factory.CreateSourceAndDestClients(ctx, "", "", registriesConfig)
	// We expect an error due to missing AWS/GCP credentials in unit tests
	// but the error should not be about missing registry names
	if err != nil {
		if contains(err.Error(), "source registry not specified") || contains(err.Error(), "destination registry not specified") {
			t.Errorf("CreateSourceAndDestClients() should use defaults, got error: %v", err)
		}
		// Other errors (like AWS credential errors) are expected in unit tests
		t.Logf("CreateSourceAndDestClients() error = %v (expected in unit test without credentials)", err)
	}
}

func TestRegistryClientFactory_ValidateRegistryConnection_InvalidConfig(t *testing.T) {
	factory := NewRegistryClientFactory(log.NewLogger())
	ctx := context.Background()

	invalidConfig := &config.RegistryConfig{
		Name: "invalid",
		Type: config.RegistryTypeECR,
		// Missing required region
	}

	err := factory.ValidateRegistryConnection(ctx, invalidConfig)
	if err == nil {
		t.Error("ValidateRegistryConnection() with invalid config should return error")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
