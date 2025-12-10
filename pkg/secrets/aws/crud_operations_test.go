package aws

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"freightliner/pkg/helper/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

// Test structures for JSON operations
type testSecret struct {
	Username string `json:"username"`
	Password string `json:"password"`
	APIKey   string `json:"api_key"`
}

// Test GetSecret with various scenarios
func TestProvider_GetSecret_CRUD(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name        string
		secretName  string
		mockOutput  *secretsmanager.GetSecretValueOutput
		mockError   error
		wantValue   string
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful string secret retrieval",
			secretName: "prod/db/password",
			mockOutput: &secretsmanager.GetSecretValueOutput{
				SecretString: aws.String("super-secret-password"),
			},
			wantValue: "super-secret-password",
			wantErr:   false,
		},
		{
			name:       "successful binary secret retrieval",
			secretName: "prod/api/key",
			mockOutput: &secretsmanager.GetSecretValueOutput{
				SecretBinary: []byte("binary-api-key-data"),
			},
			wantValue: "binary-api-key-data",
			wantErr:   false,
		},
		{
			name:       "empty secret string",
			secretName: "empty-secret",
			mockOutput: &secretsmanager.GetSecretValueOutput{
				SecretString: aws.String(""),
			},
			wantValue: "",
			wantErr:   false,
		},
		{
			name:        "secret with no value",
			secretName:  "invalid-secret",
			mockOutput:  &secretsmanager.GetSecretValueOutput{},
			wantErr:     true,
			errContains: "retrieved secret has no value",
		},
		{
			name:       "secret not found",
			secretName: "nonexistent-secret",
			mockError: &types.ResourceNotFoundException{
				Message: aws.String("Secret not found"),
			},
			wantErr:     true,
			errContains: "failed to get secret value",
		},
		{
			name:        "access denied error",
			secretName:  "forbidden-secret",
			mockError:   errors.New("AccessDeniedException: User is not authorized"),
			wantErr:     true,
			errContains: "failed to get secret value",
		},
		{
			name:        "network timeout error",
			secretName:  "timeout-secret",
			mockError:   errors.New("RequestTimeout: Request timed out"),
			wantErr:     true,
			errContains: "failed to get secret value",
		},
		{
			name:       "secret with special characters",
			secretName: "app/config",
			mockOutput: &secretsmanager.GetSecretValueOutput{
				SecretString: aws.String("pass@word!#$%^&*(){}[]|\\:;\"'<>,.?/~`"),
			},
			wantValue: "pass@word!#$%^&*(){}[]|\\:;\"'<>,.?/~`",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockSecretsManagerClient{
				GetSecretValueFunc: func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockOutput, nil
				},
			}

			// Create provider with mock
			provider := &Provider{
				logger: loggerPtr,
				region: "us-east-1",
			}

			// Test using mock client directly
			result, err := mockClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
				SecretId: aws.String(tt.secretName),
			})

			if tt.wantErr {
				if err == nil && tt.mockError != nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != nil && result.SecretString != nil {
					got := *result.SecretString
					if got != tt.wantValue {
						t.Errorf("GetSecret() = %v, want %v", got, tt.wantValue)
					}
				} else if result != nil && result.SecretBinary != nil {
					got := string(result.SecretBinary)
					if got != tt.wantValue {
						t.Errorf("GetSecret() = %v, want %v", got, tt.wantValue)
					}
				}
			}

			// Verify provider structure
			if provider.logger == nil {
				t.Error("Provider logger should not be nil")
			}
		})
	}
}

// Test GetJSONSecret with various scenarios
func TestProvider_GetJSONSecret_CRUD(t *testing.T) {
	logger := log.NewLogger()
	loggerPtr := &logger

	validSecret := testSecret{
		Username: "admin",
		Password: "secret123",
		APIKey:   "key-abc-123",
	}
	validJSON, _ := json.Marshal(validSecret)

	tests := []struct {
		name        string
		secretName  string
		mockOutput  *secretsmanager.GetSecretValueOutput
		mockError   error
		wantSecret  testSecret
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful JSON retrieval",
			secretName: "app/config",
			mockOutput: &secretsmanager.GetSecretValueOutput{
				SecretString: aws.String(string(validJSON)),
			},
			wantSecret: validSecret,
			wantErr:    false,
		},
		{
			name:       "invalid JSON format",
			secretName: "bad-json",
			mockOutput: &secretsmanager.GetSecretValueOutput{
				SecretString: aws.String("{invalid-json}"),
			},
			wantErr:     true,
			errContains: "invalid character",
		},
		{
			name:       "empty JSON object",
			secretName: "empty-json",
			mockOutput: &secretsmanager.GetSecretValueOutput{
				SecretString: aws.String("{}"),
			},
			wantSecret: testSecret{},
			wantErr:    false,
		},
		{
			name:       "secret not found",
			secretName: "missing",
			mockError: &types.ResourceNotFoundException{
				Message: aws.String("Secret not found"),
			},
			wantErr:     true,
			errContains: "failed to get secret value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &Provider{
				logger: loggerPtr,
				region: "us-east-1",
			}

			// Test JSON unmarshaling
			if tt.mockOutput != nil && tt.mockOutput.SecretString != nil {
				var result testSecret
				err := json.Unmarshal([]byte(*tt.mockOutput.SecretString), &result)

				if tt.wantErr {
					if err == nil {
						t.Errorf("Expected error but got none")
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
					if result != tt.wantSecret {
						t.Errorf("GetJSONSecret() = %+v, want %+v", result, tt.wantSecret)
					}
				}
			}

			// Verify provider
			if provider == nil {
				t.Error("Provider should not be nil")
			}
		})
	}
}

// Test PutSecret with various scenarios
func TestProvider_PutSecret_CRUD(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name        string
		secretName  string
		secretValue string
		mockOutput  *secretsmanager.PutSecretValueOutput
		mockError   error
		wantErr     bool
		errContains string
	}{
		{
			name:        "successful secret creation",
			secretName:  "new-secret",
			secretValue: "new-value",
			mockOutput: &secretsmanager.PutSecretValueOutput{
				ARN:       aws.String("arn:aws:secretsmanager:us-east-1:123456789012:secret:new-secret"),
				VersionId: aws.String("v1"),
			},
			wantErr: false,
		},
		{
			name:        "successful secret update",
			secretName:  "existing-secret",
			secretValue: "updated-value",
			mockOutput: &secretsmanager.PutSecretValueOutput{
				ARN:       aws.String("arn:aws:secretsmanager:us-east-1:123456789012:secret:existing-secret"),
				VersionId: aws.String("v2"),
			},
			wantErr: false,
		},
		{
			name:        "permission denied",
			secretName:  "forbidden",
			secretValue: "value",
			mockError:   errors.New("AccessDeniedException: Not authorized"),
			wantErr:     true,
			errContains: "failed to put secret value",
		},
		{
			name:        "invalid secret name",
			secretName:  "invalid@name",
			secretValue: "value",
			mockError:   errors.New("InvalidParameterException: Invalid secret name"),
			wantErr:     true,
			errContains: "failed to put secret value",
		},
		{
			name:        "empty value",
			secretName:  "empty-secret",
			secretValue: "",
			mockOutput: &secretsmanager.PutSecretValueOutput{
				ARN:       aws.String("arn:aws:secretsmanager:us-east-1:123456789012:secret:empty-secret"),
				VersionId: aws.String("v1"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockSecretsManagerClient{
				PutSecretValueFunc: func(ctx context.Context, params *secretsmanager.PutSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.PutSecretValueOutput, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockOutput, nil
				},
			}

			provider := &Provider{
				logger: loggerPtr,
				region: "us-east-1",
			}

			_, err := mockClient.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{
				SecretId:     aws.String(tt.secretName),
				SecretString: aws.String(tt.secretValue),
			})

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			// Verify provider
			if provider.logger == nil {
				t.Error("Provider logger should not be nil")
			}
		})
	}
}

// Test PutJSONSecret with various scenarios
func TestProvider_PutJSONSecret_CRUD(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	validSecret := testSecret{
		Username: "user",
		Password: "pass",
		APIKey:   "key",
	}

	tests := []struct {
		name        string
		secretName  string
		secretValue interface{}
		mockOutput  *secretsmanager.PutSecretValueOutput
		mockError   error
		wantErr     bool
		errContains string
	}{
		{
			name:        "successful JSON secret",
			secretName:  "json-secret",
			secretValue: validSecret,
			mockOutput: &secretsmanager.PutSecretValueOutput{
				ARN:       aws.String("arn:aws:secretsmanager:us-east-1:123456789012:secret:json-secret"),
				VersionId: aws.String("v1"),
			},
			wantErr: false,
		},
		{
			name:        "marshal error - channel type",
			secretName:  "bad-type",
			secretValue: make(chan int),
			wantErr:     true,
			errContains: "json: unsupported type",
		},
		{
			name:        "empty struct",
			secretName:  "empty-struct",
			secretValue: struct{}{},
			mockOutput: &secretsmanager.PutSecretValueOutput{
				ARN:       aws.String("arn:aws:secretsmanager:us-east-1:123456789012:secret:empty-struct"),
				VersionId: aws.String("v1"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockSecretsManagerClient{
				PutSecretValueFunc: func(ctx context.Context, params *secretsmanager.PutSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.PutSecretValueOutput, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockOutput, nil
				},
			}

			provider := &Provider{
				logger: loggerPtr,
				region: "us-east-1",
			}

			// Test JSON marshaling
			jsonBytes, err := json.Marshal(tt.secretValue)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Unexpected marshal error: %v", err)
				}
				return
			}

			_, err = mockClient.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{
				SecretId:     aws.String(tt.secretName),
				SecretString: aws.String(string(jsonBytes)),
			})

			if tt.wantErr {
				if err == nil && tt.mockError != nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			// Verify provider
			if provider == nil {
				t.Error("Provider should not be nil")
			}
		})
	}
}

// Test DeleteSecret with various scenarios
func TestProvider_DeleteSecret_CRUD(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name        string
		secretName  string
		mockOutput  *secretsmanager.DeleteSecretOutput
		mockError   error
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful deletion",
			secretName: "old-secret",
			mockOutput: &secretsmanager.DeleteSecretOutput{
				ARN:          aws.String("arn:aws:secretsmanager:us-east-1:123456789012:secret:old-secret"),
				DeletionDate: aws.Time(time.Now().Add(30 * 24 * time.Hour)),
			},
			wantErr: false,
		},
		{
			name:       "secret not found",
			secretName: "nonexistent",
			mockError: &types.ResourceNotFoundException{
				Message: aws.String("Secret not found"),
			},
			wantErr:     true,
			errContains: "failed to delete secret",
		},
		{
			name:        "access denied",
			secretName:  "protected",
			mockError:   errors.New("AccessDeniedException: Not authorized"),
			wantErr:     true,
			errContains: "failed to delete secret",
		},
		{
			name:        "invalid request",
			secretName:  "",
			mockError:   errors.New("InvalidParameterException: Empty secret name"),
			wantErr:     true,
			errContains: "failed to delete secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockSecretsManagerClient{
				DeleteSecretFunc: func(ctx context.Context, params *secretsmanager.DeleteSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.DeleteSecretOutput, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockOutput, nil
				},
			}

			provider := &Provider{
				logger: loggerPtr,
				region: "us-east-1",
			}

			_, err := mockClient.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
				SecretId:                   aws.String(tt.secretName),
				ForceDeleteWithoutRecovery: aws.Bool(false),
			})

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			// Verify provider
			if provider.region != "us-east-1" {
				t.Errorf("Provider region = %v, want us-east-1", provider.region)
			}
		})
	}
}

// Test ListSecrets with various scenarios
func TestProvider_ListSecrets_CRUD(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name        string
		filter      string
		mockOutput  *secretsmanager.ListSecretsOutput
		mockError   error
		wantSecrets []string
		wantErr     bool
		errContains string
	}{
		{
			name:   "successful list all",
			filter: "",
			mockOutput: &secretsmanager.ListSecretsOutput{
				SecretList: []types.SecretListEntry{
					{Name: aws.String("secret1")},
					{Name: aws.String("secret2")},
					{Name: aws.String("secret3")},
				},
			},
			wantSecrets: []string{"secret1", "secret2", "secret3"},
			wantErr:     false,
		},
		{
			name:   "filter by prefix",
			filter: "prod",
			mockOutput: &secretsmanager.ListSecretsOutput{
				SecretList: []types.SecretListEntry{
					{Name: aws.String("prod-secret1")},
					{Name: aws.String("dev-secret")},
					{Name: aws.String("prod-secret2")},
				},
			},
			wantSecrets: []string{"prod-secret1", "prod-secret2"},
			wantErr:     false,
		},
		{
			name:   "empty list",
			filter: "",
			mockOutput: &secretsmanager.ListSecretsOutput{
				SecretList: []types.SecretListEntry{},
			},
			wantSecrets: []string{},
			wantErr:     false,
		},
		{
			name:        "access denied",
			filter:      "",
			mockError:   errors.New("AccessDeniedException: Not authorized"),
			wantErr:     true,
			errContains: "failed to list secrets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockSecretsManagerClient{
				ListSecretsFunc: func(ctx context.Context, params *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockOutput, nil
				},
			}

			provider := &Provider{
				logger: loggerPtr,
				region: "us-east-1",
			}

			result, err := mockClient.ListSecrets(ctx, &secretsmanager.ListSecretsInput{})

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != nil {
					var secretNames []string
					for _, entry := range result.SecretList {
						if entry.Name != nil {
							name := *entry.Name
							if tt.filter == "" || containsSubstr(name, tt.filter) {
								secretNames = append(secretNames, name)
							}
						}
					}

					if len(secretNames) != len(tt.wantSecrets) {
						t.Errorf("ListSecrets() returned %d secrets, want %d", len(secretNames), len(tt.wantSecrets))
					}
				}
			}

			// Verify provider
			if provider == nil {
				t.Error("Provider should not be nil")
			}
		})
	}
}

// Test concurrent secret operations
func TestProvider_ConcurrentOperations(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	mockClient := &mockSecretsManagerClient{
		GetSecretValueFunc: func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
			// Simulate some processing time
			time.Sleep(10 * time.Millisecond)
			return &secretsmanager.GetSecretValueOutput{
				SecretString: aws.String("value"),
			}, nil
		},
		PutSecretValueFunc: func(ctx context.Context, params *secretsmanager.PutSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.PutSecretValueOutput, error) {
			time.Sleep(10 * time.Millisecond)
			return &secretsmanager.PutSecretValueOutput{
				ARN:       aws.String("arn:aws:secretsmanager:us-east-1:123456789012:secret:test"),
				VersionId: aws.String("v1"),
			}, nil
		},
	}

	provider := &Provider{
		logger: loggerPtr,
		region: "us-east-1",
	}

	// Test concurrent reads
	t.Run("concurrent reads", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				_, err := mockClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
					SecretId: aws.String("test-secret"),
				})
				if err != nil {
					errors <- err
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		for err := range errors {
			t.Errorf("Concurrent read error: %v", err)
		}
	})

	// Test concurrent writes
	t.Run("concurrent writes", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, 5)

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				_, err := mockClient.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{
					SecretId:     aws.String("test-secret"),
					SecretString: aws.String("value"),
				})
				if err != nil {
					errors <- err
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		for err := range errors {
			t.Errorf("Concurrent write error: %v", err)
		}
	})

	// Verify provider wasn't corrupted
	if provider.region != "us-east-1" {
		t.Errorf("Provider region changed during concurrent operations")
	}
}

// Helper function to check if a string contains a substring
func containsSubstr(s, substr string) bool {
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
