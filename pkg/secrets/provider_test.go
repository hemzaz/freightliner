package secrets

import (
	"context"
	"errors"
	"testing"

	"freightliner/pkg/helper/log"
)

// mockProvider implements the Provider interface for testing
type mockProvider struct {
	getSecretFunc     func(ctx context.Context, secretName string) (string, error)
	getJSONSecretFunc func(ctx context.Context, secretName string, v interface{}) error
	putSecretFunc     func(ctx context.Context, secretName, secretValue string) error
	putJSONSecretFunc func(ctx context.Context, secretName string, v interface{}) error
	deleteSecretFunc  func(ctx context.Context, secretName string) error
}

func (m *mockProvider) GetSecret(ctx context.Context, secretName string) (string, error) {
	if m.getSecretFunc != nil {
		return m.getSecretFunc(ctx, secretName)
	}
	return "", errors.New("not implemented")
}

func (m *mockProvider) GetJSONSecret(ctx context.Context, secretName string, v interface{}) error {
	if m.getJSONSecretFunc != nil {
		return m.getJSONSecretFunc(ctx, secretName, v)
	}
	return errors.New("not implemented")
}

func (m *mockProvider) PutSecret(ctx context.Context, secretName, secretValue string) error {
	if m.putSecretFunc != nil {
		return m.putSecretFunc(ctx, secretName, secretValue)
	}
	return errors.New("not implemented")
}

func (m *mockProvider) PutJSONSecret(ctx context.Context, secretName string, v interface{}) error {
	if m.putJSONSecretFunc != nil {
		return m.putJSONSecretFunc(ctx, secretName, v)
	}
	return errors.New("not implemented")
}

func (m *mockProvider) DeleteSecret(ctx context.Context, secretName string) error {
	if m.deleteSecretFunc != nil {
		return m.deleteSecretFunc(ctx, secretName)
	}
	return errors.New("not implemented")
}

func TestProviderConstants(t *testing.T) {
	tests := []struct {
		name     string
		provider ProviderType
		expected string
	}{
		{
			name:     "AWS provider constant",
			provider: AWSProvider,
			expected: "aws",
		},
		{
			name:     "GCP provider constant",
			provider: GCPProvider,
			expected: "gcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.provider) != tt.expected {
				t.Errorf("Provider constant mismatch: got %s, want %s", tt.provider, tt.expected)
			}
		})
	}
}

func TestGetProvider(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger // Convert interface to pointer to interface

	tests := []struct {
		name        string
		opts        ManagerOptions
		wantErr     bool
		errContains string
	}{
		{
			name: "missing logger",
			opts: ManagerOptions{
				Provider: AWSProvider,
			},
			wantErr:     true,
			errContains: "logger is required",
		},
		{
			name: "unsupported provider",
			opts: ManagerOptions{
				Provider: "invalid",
				Logger:   loggerPtr,
			},
			wantErr:     true,
			errContains: "unsupported provider type",
		},
		{
			name: "GCP provider with missing project",
			opts: ManagerOptions{
				Provider:   GCPProvider,
				Logger:     loggerPtr,
				GCPProject: "",
			},
			wantErr:     true,
			errContains: "project is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := GetProvider(ctx, tt.opts)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetProvider() expected error, got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("GetProvider() error = %v, want error containing %s", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("GetProvider() unexpected error = %v", err)
				}
				if provider == nil {
					t.Errorf("GetProvider() returned nil provider")
				}
			}
		})
	}
}

func TestGetProviderAWS(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()

	opts := ManagerOptions{
		Provider:  AWSProvider,
		Logger:    &logger,
		AWSRegion: "us-east-1",
	}

	// This test verifies provider creation without requiring AWS credentials
	// The AWS SDK will use default config which may or may not work in CI
	// but the provider structure should be created correctly
	provider, err := GetProvider(ctx, opts)

	// In short mode or without credentials, we just verify structure
	if err != nil && !contains(err.Error(), "failed to load AWS configuration") {
		t.Fatalf("GetProvider() unexpected error type = %v", err)
	}

	// If provider was created successfully, verify structure
	if err == nil && provider == nil {
		t.Fatal("GetProvider() returned nil provider without error")
	}
}

func TestGetProviderGCP(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()

	opts := ManagerOptions{
		Provider:   GCPProvider,
		Logger:     &logger,
		GCPProject: "test-project-123",
	}

	// This test verifies provider creation logic
	// GCP will fail without credentials, but we verify error handling
	provider, err := GetProvider(ctx, opts)

	// Expected to fail without credentials, but error should be about client creation
	if err != nil && !contains(err.Error(), "failed to create Secret Manager client") {
		t.Fatalf("GetProvider() unexpected error type = %v", err)
	}

	// If provider was created (has credentials), verify structure
	if err == nil && provider == nil {
		t.Fatal("GetProvider() returned nil provider without error")
	}
}

func TestManagerOptions(t *testing.T) {
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name string
		opts ManagerOptions
	}{
		{
			name: "AWS options",
			opts: ManagerOptions{
				Provider:  AWSProvider,
				Logger:    loggerPtr,
				AWSRegion: "us-west-2",
			},
		},
		{
			name: "GCP options",
			opts: ManagerOptions{
				Provider:           GCPProvider,
				Logger:             loggerPtr,
				GCPProject:         "my-project",
				GCPCredentialsFile: "/path/to/creds.json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts.Logger == nil {
				t.Error("Logger should not be nil")
			}
			if tt.opts.Provider == "" {
				t.Error("Provider should not be empty")
			}

			switch tt.opts.Provider {
			case AWSProvider:
				if tt.opts.AWSRegion == "" {
					t.Error("AWS region should not be empty")
				}
			case GCPProvider:
				if tt.opts.GCPProject == "" {
					t.Error("GCP project should not be empty")
				}
			}
		})
	}
}

func TestProviderInterface(t *testing.T) {
	ctx := context.Background()

	mock := &mockProvider{
		getSecretFunc: func(ctx context.Context, secretName string) (string, error) {
			return "test-value", nil
		},
		putSecretFunc: func(ctx context.Context, secretName, secretValue string) error {
			return nil
		},
		deleteSecretFunc: func(ctx context.Context, secretName string) error {
			return nil
		},
	}

	// Verify mock implements Provider interface
	var _ Provider = mock

	// Test GetSecret
	val, err := mock.GetSecret(ctx, "test-secret")
	if err != nil {
		t.Errorf("GetSecret() error = %v", err)
	}
	if val != "test-value" {
		t.Errorf("GetSecret() = %v, want test-value", val)
	}

	// Test PutSecret
	err = mock.PutSecret(ctx, "test-secret", "new-value")
	if err != nil {
		t.Errorf("PutSecret() error = %v", err)
	}

	// Test DeleteSecret
	err = mock.DeleteSecret(ctx, "test-secret")
	if err != nil {
		t.Errorf("DeleteSecret() error = %v", err)
	}
}

func TestProviderInterfaceErrors(t *testing.T) {
	ctx := context.Background()

	mock := &mockProvider{
		getSecretFunc: func(ctx context.Context, secretName string) (string, error) {
			return "", errors.New("secret not found")
		},
		putSecretFunc: func(ctx context.Context, secretName, secretValue string) error {
			return errors.New("permission denied")
		},
		deleteSecretFunc: func(ctx context.Context, secretName string) error {
			return errors.New("secret is protected")
		},
		getJSONSecretFunc: func(ctx context.Context, secretName string, v interface{}) error {
			return errors.New("invalid JSON")
		},
		putJSONSecretFunc: func(ctx context.Context, secretName string, v interface{}) error {
			return errors.New("marshal error")
		},
	}

	// Test GetSecret error
	_, err := mock.GetSecret(ctx, "test-secret")
	if err == nil || err.Error() != "secret not found" {
		t.Errorf("GetSecret() expected 'secret not found' error, got %v", err)
	}

	// Test PutSecret error
	err = mock.PutSecret(ctx, "test-secret", "value")
	if err == nil || err.Error() != "permission denied" {
		t.Errorf("PutSecret() expected 'permission denied' error, got %v", err)
	}

	// Test DeleteSecret error
	err = mock.DeleteSecret(ctx, "test-secret")
	if err == nil || err.Error() != "secret is protected" {
		t.Errorf("DeleteSecret() expected 'secret is protected' error, got %v", err)
	}

	// Test GetJSONSecret error
	var result map[string]string
	err = mock.GetJSONSecret(ctx, "test-secret", &result)
	if err == nil || err.Error() != "invalid JSON" {
		t.Errorf("GetJSONSecret() expected 'invalid JSON' error, got %v", err)
	}

	// Test PutJSONSecret error
	err = mock.PutJSONSecret(ctx, "test-secret", map[string]string{"key": "value"})
	if err == nil || err.Error() != "marshal error" {
		t.Errorf("PutJSONSecret() expected 'marshal error' error, got %v", err)
	}
}

func TestProviderTypeString(t *testing.T) {
	tests := []struct {
		provider ProviderType
		want     string
	}{
		{AWSProvider, "aws"},
		{GCPProvider, "gcp"},
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			if string(tt.provider) != tt.want {
				t.Errorf("ProviderType string = %v, want %v", string(tt.provider), tt.want)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
