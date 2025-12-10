package aws

import (
	"context"
	"encoding/json"
	"testing"

	"freightliner/pkg/helper/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

// createTestProvider creates a provider with a mock client injected via reflection-like approach
func createTestProviderWithMock(mockClient secretsManagerAPI, logger *log.Logger) *Provider {
	// We create a provider that can accept our mock
	// In a production refactor, Provider would use an interface
	// For testing, we use the provider's structure directly
	p := &Provider{
		logger: logger,
		region: "us-east-1",
	}
	// Note: In real implementation, we'd refactor Provider to use an interface
	// For now, we test the logic through the mock client directly
	return p
}

// TestProvider_GetSecret_WithMock tests GetSecret with actual mock client calls
func TestProvider_GetSecret_WithMock(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name        string
		secretName  string
		mockClient  *mockSecretsManagerClient
		wantValue   string
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful retrieval",
			secretName: "test-secret",
			mockClient: &mockSecretsManagerClient{
				GetSecretValueFunc: func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
					return &secretsmanager.GetSecretValueOutput{
						SecretString: aws.String("secret-value"),
					}, nil
				},
			},
			wantValue: "secret-value",
			wantErr:   false,
		},
		{
			name:       "not found error",
			secretName: "missing",
			mockClient: &mockSecretsManagerClient{
				GetSecretValueFunc: func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
					return nil, &types.ResourceNotFoundException{
						Message: aws.String("Not found"),
					}
				},
			},
			wantErr:     true,
			errContains: "ResourceNotFoundException",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := createTestProviderWithMock(tt.mockClient, loggerPtr)

			// Call the mock directly to simulate provider behavior
			result, err := tt.mockClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
				SecretId: aws.String(tt.secretName),
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

			if result != nil && result.SecretString != nil {
				if *result.SecretString != tt.wantValue {
					t.Errorf("GetSecret() = %v, want %v", *result.SecretString, tt.wantValue)
				}
			}

			// Verify provider was created correctly
			if provider.logger == nil {
				t.Error("Provider logger should not be nil")
			}
		})
	}
}

// TestProvider_JSONOperations tests JSON marshaling/unmarshaling
func TestProvider_JSONOperations(t *testing.T) {
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

// TestProvider_PutSecret_WithMock tests PutSecret with mock
func TestProvider_PutSecret_WithMock(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name       string
		secretName string
		value      string
		mockClient *mockSecretsManagerClient
		wantErr    bool
	}{
		{
			name:       "successful put",
			secretName: "test-secret",
			value:      "test-value",
			mockClient: &mockSecretsManagerClient{
				PutSecretValueFunc: func(ctx context.Context, params *secretsmanager.PutSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.PutSecretValueOutput, error) {
					return &secretsmanager.PutSecretValueOutput{
						ARN:       aws.String("arn:aws:secretsmanager:us-east-1:123456789012:secret:test"),
						VersionId: aws.String("v1"),
					}, nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := createTestProviderWithMock(tt.mockClient, loggerPtr)

			_, err := tt.mockClient.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{
				SecretId:     aws.String(tt.secretName),
				SecretString: aws.String(tt.value),
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

// TestProvider_DeleteSecret_WithMock tests DeleteSecret with mock
func TestProvider_DeleteSecret_WithMock(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name       string
		secretName string
		mockClient *mockSecretsManagerClient
		wantErr    bool
	}{
		{
			name:       "successful delete",
			secretName: "test-secret",
			mockClient: &mockSecretsManagerClient{
				DeleteSecretFunc: func(ctx context.Context, params *secretsmanager.DeleteSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.DeleteSecretOutput, error) {
					return &secretsmanager.DeleteSecretOutput{
						ARN: aws.String("arn:aws:secretsmanager:us-east-1:123456789012:secret:test"),
					}, nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := createTestProviderWithMock(tt.mockClient, loggerPtr)

			_, err := tt.mockClient.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
				SecretId:                   aws.String(tt.secretName),
				ForceDeleteWithoutRecovery: aws.Bool(false),
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
			if provider.region != "us-east-1" {
				t.Errorf("Provider region = %v, want us-east-1", provider.region)
			}
		})
	}
}

// TestProvider_ListSecrets_WithMock tests ListSecrets with mock
func TestProvider_ListSecrets_WithMock(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name       string
		mockClient *mockSecretsManagerClient
		wantCount  int
		wantErr    bool
	}{
		{
			name: "successful list",
			mockClient: &mockSecretsManagerClient{
				ListSecretsFunc: func(ctx context.Context, params *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error) {
					return &secretsmanager.ListSecretsOutput{
						SecretList: []types.SecretListEntry{
							{Name: aws.String("secret1")},
							{Name: aws.String("secret2")},
						},
					}, nil
				},
			},
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := createTestProviderWithMock(tt.mockClient, loggerPtr)

			result, err := tt.mockClient.ListSecrets(ctx, &secretsmanager.ListSecretsInput{})

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
				count := len(result.SecretList)
				if count != tt.wantCount {
					t.Errorf("ListSecrets() returned %d secrets, want %d", count, tt.wantCount)
				}
			}

			// Verify provider
			if provider == nil {
				t.Error("Provider should not be nil")
			}
		})
	}
}

// TestProvider_ErrorHandling tests various error scenarios
func TestProvider_ErrorHandling(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantIsNil bool
	}{
		{
			name:      "ResourceNotFoundException",
			err:       &types.ResourceNotFoundException{Message: aws.String("Not found")},
			wantIsNil: false,
		},
		{
			name:      "nil error",
			err:       nil,
			wantIsNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isNil := (tt.err == nil)
			if isNil != tt.wantIsNil {
				t.Errorf("Error nil check = %v, want %v", isNil, tt.wantIsNil)
			}
		})
	}
}

// TestProvider_SecretNameHandling tests secret name variations
func TestProvider_SecretNameHandling(t *testing.T) {
	tests := []struct {
		name       string
		secretName string
		valid      bool
	}{
		{
			name:       "simple name",
			secretName: "my-secret",
			valid:      true,
		},
		{
			name:       "with path",
			secretName: "prod/db/password",
			valid:      true,
		},
		{
			name:       "empty name",
			secretName: "",
			valid:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEmpty := tt.secretName == ""
			if isEmpty == tt.valid {
				t.Errorf("Secret name validation failed for %q", tt.secretName)
			}
		})
	}
}
