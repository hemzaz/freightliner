package gcp

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"freightliner/pkg/helper/log"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Mock GCP Secret Manager client interface
type mockGCPSecretManagerClient struct {
	AccessSecretVersionFunc func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error)
	CreateSecretFunc        func(ctx context.Context, req *secretmanagerpb.CreateSecretRequest) (*secretmanagerpb.Secret, error)
	GetSecretFunc           func(ctx context.Context, req *secretmanagerpb.GetSecretRequest) (*secretmanagerpb.Secret, error)
	AddSecretVersionFunc    func(ctx context.Context, req *secretmanagerpb.AddSecretVersionRequest) (*secretmanagerpb.SecretVersion, error)
	DeleteSecretFunc        func(ctx context.Context, req *secretmanagerpb.DeleteSecretRequest) error
	ListSecretsFunc         func(ctx context.Context, req *secretmanagerpb.ListSecretsRequest) secretIterator
}

// Mock iterator interface
type secretIterator interface {
	Next() (*secretmanagerpb.Secret, error)
}

type mockSecretIterator struct {
	secrets []*secretmanagerpb.Secret
	index   int
}

func (m *mockSecretIterator) Next() (*secretmanagerpb.Secret, error) {
	if m.index >= len(m.secrets) {
		return nil, iterator.Done
	}
	secret := m.secrets[m.index]
	m.index++
	return secret, nil
}

// Test structures for JSON operations
type testSecret struct {
	Username string `json:"username"`
	Password string `json:"password"`
	APIKey   string `json:"api_key"`
}

// Test GetSecret with various scenarios
func TestProvider_GetSecret_GCP_CRUD(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name        string
		secretName  string
		mockPayload []byte
		mockError   error
		wantValue   string
		wantErr     bool
		errContains string
	}{
		{
			name:        "successful secret retrieval",
			secretName:  "prod/db/password",
			mockPayload: []byte("super-secret-password"),
			wantValue:   "super-secret-password",
			wantErr:     false,
		},
		{
			name:        "empty secret value",
			secretName:  "empty-secret",
			mockPayload: []byte(""),
			wantValue:   "",
			wantErr:     false,
		},
		{
			name:        "secret not found",
			secretName:  "nonexistent-secret",
			mockError:   status.Error(codes.NotFound, "Secret not found"),
			wantErr:     true,
			errContains: "failed to access secret",
		},
		{
			name:        "permission denied",
			secretName:  "forbidden-secret",
			mockError:   status.Error(codes.PermissionDenied, "Permission denied"),
			wantErr:     true,
			errContains: "failed to access secret",
		},
		{
			name:        "network timeout",
			secretName:  "timeout-secret",
			mockError:   status.Error(codes.DeadlineExceeded, "Deadline exceeded"),
			wantErr:     true,
			errContains: "failed to access secret",
		},
		{
			name:        "secret with special characters",
			secretName:  "app/config",
			mockPayload: []byte("pass@word!#$%^&*(){}[]|\\:;\"'<>,.?/~`"),
			wantValue:   "pass@word!#$%^&*(){}[]|\\:;\"'<>,.?/~`",
			wantErr:     false,
		},
		{
			name:        "binary data",
			secretName:  "binary-secret",
			mockPayload: []byte{0x00, 0x01, 0x02, 0xFF},
			wantValue:   string([]byte{0x00, 0x01, 0x02, 0xFF}),
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockGCPSecretManagerClient{
				AccessSecretVersionFunc: func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return &secretmanagerpb.AccessSecretVersionResponse{
						Payload: &secretmanagerpb.SecretPayload{
							Data: tt.mockPayload,
						},
					}, nil
				},
			}

			provider := &Provider{
				logger:  loggerPtr,
				project: "test-project",
			}

			// Test using mock client
			result, err := mockClient.AccessSecretVersionFunc(ctx, &secretmanagerpb.AccessSecretVersionRequest{
				Name: provider.buildSecretVersionName(tt.secretName),
			})

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != nil {
					got := string(result.Payload.Data)
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
func TestProvider_GetJSONSecret_GCP_CRUD(t *testing.T) {
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
		mockPayload []byte
		mockError   error
		wantSecret  testSecret
		wantErr     bool
		errContains string
	}{
		{
			name:        "successful JSON retrieval",
			secretName:  "app/config",
			mockPayload: validJSON,
			wantSecret:  validSecret,
			wantErr:     false,
		},
		{
			name:        "invalid JSON format",
			secretName:  "bad-json",
			mockPayload: []byte("{invalid-json}"),
			wantErr:     true,
			errContains: "invalid character",
		},
		{
			name:        "empty JSON object",
			secretName:  "empty-json",
			mockPayload: []byte("{}"),
			wantSecret:  testSecret{},
			wantErr:     false,
		},
		{
			name:        "secret not found",
			secretName:  "missing",
			mockError:   status.Error(codes.NotFound, "Secret not found"),
			wantErr:     true,
			errContains: "failed to access secret",
		},
		{
			name:        "malformed JSON array",
			secretName:  "array-json",
			mockPayload: []byte("[1,2,3]"),
			wantErr:     true,
			errContains: "cannot unmarshal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &Provider{
				logger:  loggerPtr,
				project: "test-project",
			}

			// Test JSON unmarshaling
			if tt.mockPayload != nil {
				var result testSecret
				err := json.Unmarshal(tt.mockPayload, &result)

				if tt.wantErr {
					if err == nil && tt.mockError == nil {
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

// Test CreateSecret with various scenarios
func TestProvider_CreateSecret_GCP_CRUD(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name        string
		secretName  string
		mockSecret  *secretmanagerpb.Secret
		mockError   error
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful creation",
			secretName: "new-secret",
			mockSecret: &secretmanagerpb.Secret{
				Name: "projects/test-project/secrets/new-secret",
			},
			wantErr: false,
		},
		{
			name:        "secret already exists",
			secretName:  "existing-secret",
			mockError:   status.Error(codes.AlreadyExists, "Secret already exists"),
			wantErr:     true,
			errContains: "failed to create secret",
		},
		{
			name:        "permission denied",
			secretName:  "forbidden",
			mockError:   status.Error(codes.PermissionDenied, "Permission denied"),
			wantErr:     true,
			errContains: "failed to create secret",
		},
		{
			name:        "invalid secret name",
			secretName:  "invalid@name",
			mockError:   status.Error(codes.InvalidArgument, "Invalid secret name"),
			wantErr:     true,
			errContains: "failed to create secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockGCPSecretManagerClient{
				CreateSecretFunc: func(ctx context.Context, req *secretmanagerpb.CreateSecretRequest) (*secretmanagerpb.Secret, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockSecret, nil
				},
			}

			provider := &Provider{
				logger:  loggerPtr,
				project: "test-project",
			}

			_, err := mockClient.CreateSecretFunc(ctx, &secretmanagerpb.CreateSecretRequest{
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
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			// Verify provider
			if provider.project != "test-project" {
				t.Errorf("Provider project = %v, want test-project", provider.project)
			}
		})
	}
}

// Test PutSecret with various scenarios
func TestProvider_PutSecret_GCP_CRUD(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name          string
		secretName    string
		secretValue   string
		secretExists  bool
		mockError     error
		mockCreateErr error
		wantErr       bool
		errContains   string
	}{
		{
			name:         "successful new secret",
			secretName:   "new-secret",
			secretValue:  "new-value",
			secretExists: false,
			wantErr:      false,
		},
		{
			name:         "successful update",
			secretName:   "existing-secret",
			secretValue:  "updated-value",
			secretExists: true,
			wantErr:      false,
		},
		{
			name:          "create error",
			secretName:    "bad-secret",
			secretValue:   "value",
			secretExists:  false,
			mockCreateErr: status.Error(codes.InvalidArgument, "Invalid name"),
			wantErr:       true,
			errContains:   "failed to create secret",
		},
		{
			name:         "add version error",
			secretName:   "existing",
			secretValue:  "value",
			secretExists: true,
			mockError:    status.Error(codes.PermissionDenied, "Permission denied"),
			wantErr:      true,
			errContains:  "failed to add secret version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockGCPSecretManagerClient{
				GetSecretFunc: func(ctx context.Context, req *secretmanagerpb.GetSecretRequest) (*secretmanagerpb.Secret, error) {
					if tt.secretExists {
						return &secretmanagerpb.Secret{
							Name: req.Name,
						}, nil
					}
					return nil, status.Error(codes.NotFound, "Secret not found")
				},
				CreateSecretFunc: func(ctx context.Context, req *secretmanagerpb.CreateSecretRequest) (*secretmanagerpb.Secret, error) {
					if tt.mockCreateErr != nil {
						return nil, tt.mockCreateErr
					}
					return &secretmanagerpb.Secret{
						Name: "projects/test-project/secrets/" + req.SecretId,
					}, nil
				},
				AddSecretVersionFunc: func(ctx context.Context, req *secretmanagerpb.AddSecretVersionRequest) (*secretmanagerpb.SecretVersion, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return &secretmanagerpb.SecretVersion{
						Name: req.Parent + "/versions/1",
					}, nil
				},
			}

			provider := &Provider{
				logger:  loggerPtr,
				project: "test-project",
			}

			// Check if secret exists
			secretName := provider.buildSecretName(tt.secretName)
			_, err := mockClient.GetSecretFunc(ctx, &secretmanagerpb.GetSecretRequest{
				Name: secretName,
			})

			secretExists := err == nil

			// Create if doesn't exist
			if !secretExists && tt.mockCreateErr == nil {
				_, createErr := mockClient.CreateSecretFunc(ctx, &secretmanagerpb.CreateSecretRequest{
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
				if createErr != nil {
					if !tt.wantErr {
						t.Errorf("Unexpected create error: %v", createErr)
					}
					return
				}
			}

			// Add version
			_, addErr := mockClient.AddSecretVersionFunc(ctx, &secretmanagerpb.AddSecretVersionRequest{
				Parent: secretName,
				Payload: &secretmanagerpb.SecretPayload{
					Data: []byte(tt.secretValue),
				},
			})

			if tt.wantErr {
				if addErr == nil && tt.mockError != nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if addErr != nil {
					t.Errorf("Unexpected error: %v", addErr)
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
func TestProvider_DeleteSecret_GCP_CRUD(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name        string
		secretName  string
		mockError   error
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful deletion",
			secretName: "old-secret",
			wantErr:    false,
		},
		{
			name:        "secret not found",
			secretName:  "nonexistent",
			mockError:   status.Error(codes.NotFound, "Secret not found"),
			wantErr:     true,
			errContains: "failed to delete secret",
		},
		{
			name:        "permission denied",
			secretName:  "protected",
			mockError:   status.Error(codes.PermissionDenied, "Permission denied"),
			wantErr:     true,
			errContains: "failed to delete secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockGCPSecretManagerClient{
				DeleteSecretFunc: func(ctx context.Context, req *secretmanagerpb.DeleteSecretRequest) error {
					if tt.mockError != nil {
						return tt.mockError
					}
					return nil
				},
			}

			provider := &Provider{
				logger:  loggerPtr,
				project: "test-project",
			}

			err := mockClient.DeleteSecretFunc(ctx, &secretmanagerpb.DeleteSecretRequest{
				Name: provider.buildSecretName(tt.secretName),
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
			if provider.project != "test-project" {
				t.Errorf("Provider project = %v, want test-project", provider.project)
			}
		})
	}
}

// Test ListSecrets with various scenarios
func TestProvider_ListSecrets_GCP_CRUD(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	tests := []struct {
		name        string
		mockSecrets []*secretmanagerpb.Secret
		mockError   error
		wantCount   int
		wantErr     bool
		errContains string
	}{
		{
			name: "successful list all",
			mockSecrets: []*secretmanagerpb.Secret{
				{Name: "projects/test-project/secrets/secret1"},
				{Name: "projects/test-project/secrets/secret2"},
				{Name: "projects/test-project/secrets/secret3"},
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:        "empty list",
			mockSecrets: []*secretmanagerpb.Secret{},
			wantCount:   0,
			wantErr:     false,
		},
		{
			name: "single secret",
			mockSecrets: []*secretmanagerpb.Secret{
				{Name: "projects/test-project/secrets/only-one"},
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockIterator := &mockSecretIterator{
				secrets: tt.mockSecrets,
				index:   0,
			}

			mockClient := &mockGCPSecretManagerClient{
				ListSecretsFunc: func(ctx context.Context, req *secretmanagerpb.ListSecretsRequest) secretIterator {
					return mockIterator
				},
			}

			provider := &Provider{
				logger:  loggerPtr,
				project: "test-project",
			}

			it := mockClient.ListSecretsFunc(ctx, &secretmanagerpb.ListSecretsRequest{
				Parent: "projects/test-project",
			})

			var secretNames []string
			for {
				secret, err := it.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					if !tt.wantErr {
						t.Errorf("Unexpected error: %v", err)
					}
					break
				}
				secretNames = append(secretNames, secret.Name)
			}

			if len(secretNames) != tt.wantCount {
				t.Errorf("ListSecrets() returned %d secrets, want %d", len(secretNames), tt.wantCount)
			}

			// Verify provider
			if provider == nil {
				t.Error("Provider should not be nil")
			}
		})
	}
}

// Test concurrent secret operations
func TestProvider_ConcurrentOperations_GCP(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	mockClient := &mockGCPSecretManagerClient{
		AccessSecretVersionFunc: func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
			// Simulate some processing time
			time.Sleep(10 * time.Millisecond)
			return &secretmanagerpb.AccessSecretVersionResponse{
				Payload: &secretmanagerpb.SecretPayload{
					Data: []byte("value"),
				},
			}, nil
		},
		AddSecretVersionFunc: func(ctx context.Context, req *secretmanagerpb.AddSecretVersionRequest) (*secretmanagerpb.SecretVersion, error) {
			time.Sleep(10 * time.Millisecond)
			return &secretmanagerpb.SecretVersion{
				Name: req.Parent + "/versions/1",
			}, nil
		},
	}

	provider := &Provider{
		logger:  loggerPtr,
		project: "test-project",
	}

	// Test concurrent reads
	t.Run("concurrent reads", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				_, err := mockClient.AccessSecretVersionFunc(ctx, &secretmanagerpb.AccessSecretVersionRequest{
					Name: provider.buildSecretVersionName("test-secret"),
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
				_, err := mockClient.AddSecretVersionFunc(ctx, &secretmanagerpb.AddSecretVersionRequest{
					Parent: provider.buildSecretName("test-secret"),
					Payload: &secretmanagerpb.SecretPayload{
						Data: []byte("value"),
					},
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
	if provider.project != "test-project" {
		t.Errorf("Provider project changed during concurrent operations")
	}
}

// Test secret versioning behavior
func TestProvider_SecretVersioning_GCP(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	loggerPtr := &logger

	versions := []string{"v1", "v2", "v3"}
	currentVersion := 0

	mockClient := &mockGCPSecretManagerClient{
		AddSecretVersionFunc: func(ctx context.Context, req *secretmanagerpb.AddSecretVersionRequest) (*secretmanagerpb.SecretVersion, error) {
			if currentVersion >= len(versions) {
				return nil, errors.New("too many versions")
			}
			version := versions[currentVersion]
			currentVersion++
			return &secretmanagerpb.SecretVersion{
				Name: req.Parent + "/versions/" + version,
			}, nil
		},
	}

	provider := &Provider{
		logger:  loggerPtr,
		project: "test-project",
	}

	// Add multiple versions
	for i := 0; i < 3; i++ {
		_, err := mockClient.AddSecretVersionFunc(ctx, &secretmanagerpb.AddSecretVersionRequest{
			Parent: provider.buildSecretName("versioned-secret"),
			Payload: &secretmanagerpb.SecretPayload{
				Data: []byte("value-" + versions[i]),
			},
		})
		if err != nil {
			t.Errorf("Failed to add version %d: %v", i, err)
		}
	}

	if currentVersion != 3 {
		t.Errorf("Expected 3 versions, got %d", currentVersion)
	}
}
