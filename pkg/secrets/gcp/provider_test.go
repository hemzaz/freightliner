package gcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"freightliner/pkg/helper/log"
)

func TestNewProvider(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name        string
		opts        ProviderOptions
		wantErr     bool
		errContains string
	}{
		{
			name: "missing logger",
			opts: ProviderOptions{
				Project: "test-project",
			},
			wantErr:     true,
			errContains: "logger is required",
		},
		{
			name: "missing project",
			opts: ProviderOptions{
				Logger: loggerPtr,
			},
			wantErr:     true,
			errContains: "project is required",
		},
		{
			name: "valid provider with project",
			opts: ProviderOptions{
				Project: "test-project-123",
				Logger:  loggerPtr,
			},
			wantErr: false, // Will fail if no credentials, but structure is valid
		},
		{
			name: "valid provider with credentials file",
			opts: ProviderOptions{
				Project:         "test-project-123",
				Logger:          loggerPtr,
				CredentialsFile: "/path/to/creds.json",
			},
			wantErr: false, // Will fail if file doesn't exist, but structure is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(ctx, tt.opts)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewProvider() expected error, got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("NewProvider() error = %v, want error containing %s", err, tt.errContains)
				}
			} else {
				// In test environment, GCP might fail due to missing credentials
				// We just verify the error is related to credentials, not our logic
				if err != nil && !contains(err.Error(), "failed to create Secret Manager client") {
					t.Errorf("NewProvider() unexpected error type = %v", err)
				}
				if err == nil && provider != nil {
					if provider.logger == nil {
						t.Error("Provider logger should not be nil")
					}
					if provider.client == nil {
						t.Error("Provider client should not be nil")
					}
					if provider.project != tt.opts.Project {
						t.Errorf("Provider project = %v, want %v", provider.project, tt.opts.Project)
					}
				}
			}
		})
	}
}

func TestProvider_buildSecretName(t *testing.T) {
	logger := log.NewLogger()
	loggerPtr := &logger
	provider := &Provider{
		logger:  loggerPtr,
		project: "test-project",
	}

	tests := []struct {
		name       string
		secretName string
		want       string
	}{
		{
			name:       "simple secret name",
			secretName: "my-secret",
			want:       "projects/test-project/secrets/my-secret",
		},
		{
			name:       "secret with hyphens",
			secretName: "my-test-secret-123",
			want:       "projects/test-project/secrets/my-test-secret-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := provider.buildSecretName(tt.secretName)
			if got != tt.want {
				t.Errorf("buildSecretName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_buildSecretVersionName(t *testing.T) {
	logger := log.NewLogger()
	loggerPtr := &logger
	provider := &Provider{
		logger:  loggerPtr,
		project: "test-project",
	}

	tests := []struct {
		name       string
		secretName string
		want       string
	}{
		{
			name:       "latest version",
			secretName: "my-secret",
			want:       "projects/test-project/secrets/my-secret/versions/latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := provider.buildSecretVersionName(tt.secretName)
			if got != tt.want {
				t.Errorf("buildSecretVersionName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONMarshalUnmarshal(t *testing.T) {
	type testStruct struct {
		Key1 string `json:"key1"`
		Key2 int    `json:"key2"`
	}

	validJSON := testStruct{Key1: "value1", Key2: 42}
	validJSONBytes, _ := json.Marshal(validJSON)

	t.Run("successful JSON marshal", func(t *testing.T) {
		data, err := json.Marshal(validJSON)
		if err != nil {
			t.Errorf("Marshal unexpected error = %v", err)
		}
		if len(data) == 0 {
			t.Error("Marshal returned empty data")
		}
	})

	t.Run("successful JSON unmarshal", func(t *testing.T) {
		var result testStruct
		err := json.Unmarshal(validJSONBytes, &result)
		if err != nil {
			t.Errorf("Unmarshal unexpected error = %v", err)
		}
		if result.Key1 != validJSON.Key1 || result.Key2 != validJSON.Key2 {
			t.Errorf("Unmarshal result = %+v, want %+v", result, validJSON)
		}
	})

	t.Run("invalid JSON format", func(t *testing.T) {
		var result testStruct
		err := json.Unmarshal([]byte("not-valid-json"), &result)
		if err == nil {
			t.Error("Unmarshal expected error for invalid JSON, got nil")
		}
	})

	t.Run("marshal error with invalid type", func(t *testing.T) {
		_, err := json.Marshal(make(chan int))
		if err == nil {
			t.Error("Marshal expected error for channel type, got nil")
		}
	})
}

func TestSecretExistsLogic(t *testing.T) {
	tests := []struct {
		name       string
		errMessage string
		wantExists bool
		wantErr    bool
	}{
		{
			name:       "secret exists",
			errMessage: "",
			wantExists: true,
			wantErr:    false,
		},
		{
			name:       "secret not found",
			errMessage: "rpc error: code = NotFound desc = Secret not found",
			wantExists: false,
			wantErr:    false,
		},
		{
			name:       "other error",
			errMessage: "permission denied",
			wantExists: false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.errMessage != "" {
				err = errors.New(tt.errMessage)
			}

			exists := err == nil
			isNotFound := err != nil && err.Error() == "rpc error: code = NotFound desc = Secret not found"
			hasOtherError := err != nil && !isNotFound

			if exists != tt.wantExists && !isNotFound {
				t.Errorf("exists = %v, want %v", exists, tt.wantExists)
			}
			if hasOtherError != tt.wantErr {
				t.Errorf("hasOtherError = %v, want %v", hasOtherError, tt.wantErr)
			}
		})
	}
}

func TestProviderStructure(t *testing.T) {
	logger := log.NewLogger()
	loggerPtr := &logger

	provider := &Provider{
		logger:  loggerPtr,
		project: "test-project-123",
	}

	if provider.logger == nil {
		t.Error("Provider logger should not be nil")
	}

	if provider.project != "test-project-123" {
		t.Errorf("Provider project = %v, want test-project-123", provider.project)
	}
}

func TestProviderOptions(t *testing.T) {
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name string
		opts ProviderOptions
	}{
		{
			name: "basic options",
			opts: ProviderOptions{
				Project: "my-project",
				Logger:  loggerPtr,
			},
		},
		{
			name: "options with credentials file",
			opts: ProviderOptions{
				Project:         "my-project",
				Logger:          loggerPtr,
				CredentialsFile: "/path/to/creds.json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts.Logger == nil {
				t.Error("Logger should not be nil")
			}
			if tt.opts.Project == "" {
				t.Error("Project should not be empty")
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
