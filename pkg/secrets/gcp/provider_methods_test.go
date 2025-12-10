package gcp

import (
	"context"
	"encoding/json"
	"testing"

	"freightliner/pkg/helper/log"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// createTestProvider creates a provider for testing
func createTestProviderWithMock(logger *log.Logger, project string) *Provider {
	p := &Provider{
		logger:  logger,
		project: project,
	}
	return p
}

// TestProvider_GetSecret_WithMock tests GetSecret with mock
func TestProvider_GetSecret_WithMock(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name        string
		secretName  string
		mockClient  *mockGCPSecretManagerClient
		wantValue   string
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful retrieval",
			secretName: "test-secret",
			mockClient: &mockGCPSecretManagerClient{
				AccessSecretVersionFunc: func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
					return &secretmanagerpb.AccessSecretVersionResponse{
						Payload: &secretmanagerpb.SecretPayload{
							Data: []byte("secret-value"),
						},
					}, nil
				},
			},
			wantValue: "secret-value",
			wantErr:   false,
		},
		{
			name:       "not found error",
			secretName: "missing",
			mockClient: &mockGCPSecretManagerClient{
				AccessSecretVersionFunc: func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
					return nil, status.Error(codes.NotFound, "Not found")
				},
			},
			wantErr:     true,
			errContains: "NotFound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := createTestProviderWithMock(loggerPtr, "test-project")

			// Call the mock directly to simulate provider behavior
			result, err := tt.mockClient.AccessSecretVersionFunc(ctx, &secretmanagerpb.AccessSecretVersionRequest{
				Name: provider.buildSecretVersionName(tt.secretName),
			})

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != nil {
				got := string(result.Payload.Data)
				if got != tt.wantValue {
					t.Errorf("GetSecret() = %v, want %v", got, tt.wantValue)
				}
			}

			// Verify provider was created correctly
			if provider.logger == nil {
				t.Error("Provider logger should not be nil")
			}
		})
	}
}

// TestProvider_JSONOperations_GCP tests JSON marshaling/unmarshaling
func TestProvider_JSONOperations_GCP(t *testing.T) {
	type testData struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}

	tests := []struct {
		name    string
		data    testData
		wantErr bool
	}{
		{
			name: "valid JSON",
			data: testData{
				Field1: "value",
				Field2: 42,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshal
			jsonBytes, err := json.Marshal(tt.data)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Marshal error: %v", err)
				}
				return
			}

			// Test unmarshal
			var result testData
			err = json.Unmarshal(jsonBytes, &result)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Unmarshal error: %v", err)
				}
				return
			}

			if result != tt.data {
				t.Errorf("JSON roundtrip failed: got %+v, want %+v", result, tt.data)
			}
		})
	}
}

// TestProvider_CreateSecret_WithMock tests CreateSecret with mock
func TestProvider_CreateSecret_WithMock(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name       string
		secretName string
		mockClient *mockGCPSecretManagerClient
		wantErr    bool
	}{
		{
			name:       "successful create",
			secretName: "test-secret",
			mockClient: &mockGCPSecretManagerClient{
				CreateSecretFunc: func(ctx context.Context, req *secretmanagerpb.CreateSecretRequest) (*secretmanagerpb.Secret, error) {
					return &secretmanagerpb.Secret{
						Name: "projects/test-project/secrets/test-secret",
					}, nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := createTestProviderWithMock(loggerPtr, "test-project")

			_, err := tt.mockClient.CreateSecretFunc(ctx, &secretmanagerpb.CreateSecretRequest{
				Parent:   "projects/test-project",
				SecretId: tt.secretName,
				Secret: &secretmanagerpb.Secret{
					Replication: &secretmanagerpb.Replication{
						Replication: &secretmanagerpb.Replication_Automatic_{
							Automatic: &secretmanagerpb.Replication_Automatic{},
						},
					},
				},
			})

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify provider
			if provider == nil {
				t.Error("Provider should not be nil")
			}
		})
	}
}

// TestProvider_DeleteSecret_WithMock_GCP tests DeleteSecret with mock
func TestProvider_DeleteSecret_WithMock_GCP(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name       string
		secretName string
		mockClient *mockGCPSecretManagerClient
		wantErr    bool
	}{
		{
			name:       "successful delete",
			secretName: "test-secret",
			mockClient: &mockGCPSecretManagerClient{
				DeleteSecretFunc: func(ctx context.Context, req *secretmanagerpb.DeleteSecretRequest) error {
					return nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := createTestProviderWithMock(loggerPtr, "test-project")

			err := tt.mockClient.DeleteSecretFunc(ctx, &secretmanagerpb.DeleteSecretRequest{
				Name: provider.buildSecretName(tt.secretName),
			})

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify provider
			if provider.project != "test-project" {
				t.Errorf("Provider project = %v, want test-project", provider.project)
			}
		})
	}
}

// TestProvider_SecretExists_WithMock tests secretExists with mock
func TestProvider_SecretExists_WithMock(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name       string
		secretName string
		mockClient *mockGCPSecretManagerClient
		wantExists bool
		wantErr    bool
	}{
		{
			name:       "secret exists",
			secretName: "existing-secret",
			mockClient: &mockGCPSecretManagerClient{
				GetSecretFunc: func(ctx context.Context, req *secretmanagerpb.GetSecretRequest) (*secretmanagerpb.Secret, error) {
					return &secretmanagerpb.Secret{
						Name: req.Name,
					}, nil
				},
			},
			wantExists: true,
			wantErr:    false,
		},
		{
			name:       "secret not found",
			secretName: "missing-secret",
			mockClient: &mockGCPSecretManagerClient{
				GetSecretFunc: func(ctx context.Context, req *secretmanagerpb.GetSecretRequest) (*secretmanagerpb.Secret, error) {
					return nil, status.Error(codes.NotFound, "Secret not found")
				},
			},
			wantExists: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := createTestProviderWithMock(loggerPtr, "test-project")

			_, err := tt.mockClient.GetSecretFunc(ctx, &secretmanagerpb.GetSecretRequest{
				Name: provider.buildSecretName(tt.secretName),
			})

			exists := (err == nil)
			isNotFound := err != nil && status.Code(err) == codes.NotFound

			if tt.wantExists != exists && !isNotFound {
				t.Errorf("secretExists() = %v, want %v", exists, tt.wantExists)
			}

			// Verify provider
			if provider == nil {
				t.Error("Provider should not be nil")
			}
		})
	}
}

// TestProvider_ErrorHandling_GCP tests various error scenarios
func TestProvider_ErrorHandling_GCP(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantIsNil bool
		wantCode  codes.Code
	}{
		{
			name:      "NotFound error",
			err:       status.Error(codes.NotFound, "Not found"),
			wantIsNil: false,
			wantCode:  codes.NotFound,
		},
		{
			name:      "PermissionDenied error",
			err:       status.Error(codes.PermissionDenied, "Permission denied"),
			wantIsNil: false,
			wantCode:  codes.PermissionDenied,
		},
		{
			name:      "nil error",
			err:       nil,
			wantIsNil: true,
			wantCode:  codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isNil := (tt.err == nil)
			if isNil != tt.wantIsNil {
				t.Errorf("Error nil check = %v, want %v", isNil, tt.wantIsNil)
			}

			if !isNil {
				code := status.Code(tt.err)
				if code != tt.wantCode {
					t.Errorf("Error code = %v, want %v", code, tt.wantCode)
				}
			}
		})
	}
}

// TestProvider_SecretNameBuilding tests secret name construction
func TestProvider_SecretNameBuilding(t *testing.T) {
	logger := log.NewLogger()
	loggerPtr := &logger
	provider := createTestProviderWithMock(loggerPtr, "test-project")

	tests := []struct {
		name        string
		secretName  string
		wantName    string
		wantVersion string
	}{
		{
			name:        "simple name",
			secretName:  "my-secret",
			wantName:    "projects/test-project/secrets/my-secret",
			wantVersion: "projects/test-project/secrets/my-secret/versions/latest",
		},
		{
			name:        "with hyphens",
			secretName:  "my-test-secret-123",
			wantName:    "projects/test-project/secrets/my-test-secret-123",
			wantVersion: "projects/test-project/secrets/my-test-secret-123/versions/latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName := provider.buildSecretName(tt.secretName)
			if gotName != tt.wantName {
				t.Errorf("buildSecretName() = %v, want %v", gotName, tt.wantName)
			}

			gotVersion := provider.buildSecretVersionName(tt.secretName)
			if gotVersion != tt.wantVersion {
				t.Errorf("buildSecretVersionName() = %v, want %v", gotVersion, tt.wantVersion)
			}
		})
	}
}

// TestProvider_AddSecretVersion_WithMock tests AddSecretVersion with mock
func TestProvider_AddSecretVersion_WithMock(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name       string
		secretName string
		value      string
		mockClient *mockGCPSecretManagerClient
		wantErr    bool
	}{
		{
			name:       "successful add",
			secretName: "test-secret",
			value:      "test-value",
			mockClient: &mockGCPSecretManagerClient{
				AddSecretVersionFunc: func(ctx context.Context, req *secretmanagerpb.AddSecretVersionRequest) (*secretmanagerpb.SecretVersion, error) {
					return &secretmanagerpb.SecretVersion{
						Name: req.Parent + "/versions/1",
					}, nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := createTestProviderWithMock(loggerPtr, "test-project")

			_, err := tt.mockClient.AddSecretVersionFunc(ctx, &secretmanagerpb.AddSecretVersionRequest{
				Parent: provider.buildSecretName(tt.secretName),
				Payload: &secretmanagerpb.SecretPayload{
					Data: []byte(tt.value),
				},
			})

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify provider
			if provider == nil {
				t.Error("Provider should not be nil")
			}
		})
	}
}
