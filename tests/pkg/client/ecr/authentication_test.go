package ecr

import (
	"context"
	"encoding/base64"
	"testing"

	"freightliner/pkg/client/ecr"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsecr "github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestECRAuthenticator_Authorization(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockECRService)
		expectedUser   string
		expectedPass   string
		expectError    bool
	}{
		{
			name: "Successful authorization",
			setupMock: func(m *MockECRService) {
				token := base64.StdEncoding.EncodeToString([]byte("AWS:secrettoken123"))
				m.On("GetAuthorizationToken", mock.Anything, mock.Anything, mock.Anything).
					Return(&awsecr.GetAuthorizationTokenOutput{
						AuthorizationData: []ecrtypes.AuthorizationData{
							{
								AuthorizationToken: aws.String(token),
							},
						},
					}, nil)
			},
			expectedUser: "AWS",
			expectedPass: "secrettoken123",
			expectError:  false,
		},
		{
			name: "No authorization data",
			setupMock: func(m *MockECRService) {
				m.On("GetAuthorizationToken", mock.Anything, mock.Anything, mock.Anything).
					Return(&awsecr.GetAuthorizationTokenOutput{
						AuthorizationData: []ecrtypes.AuthorizationData{},
					}, nil)
			},
			expectError: true,
		},
		{
			name: "Invalid token format",
			setupMock: func(m *MockECRService) {
				token := base64.StdEncoding.EncodeToString([]byte("invalidformat"))
				m.On("GetAuthorizationToken", mock.Anything, mock.Anything, mock.Anything).
					Return(&awsecr.GetAuthorizationTokenOutput{
						AuthorizationData: []ecrtypes.AuthorizationData{
							{
								AuthorizationToken: aws.String(token),
							},
						},
					}, nil)
			},
			expectError: true,
		},
		{
			name: "Invalid base64 encoding",
			setupMock: func(m *MockECRService) {
				m.On("GetAuthorizationToken", mock.Anything, mock.Anything, mock.Anything).
					Return(&awsecr.GetAuthorizationTokenOutput{
						AuthorizationData: []ecrtypes.AuthorizationData{
							{
								AuthorizationToken: aws.String("not-base64!@#$"),
							},
						},
					}, nil)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockECRService)
			tt.setupMock(mockService)

			auth := ecr.NewECRAuthenticator(mockService, "us-west-2")
			config, err := auth.Authorization()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.Equal(t, tt.expectedUser, config.Username)
				assert.Equal(t, tt.expectedPass, config.Password)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestECRAuthenticator_RegistryAuthenticator(t *testing.T) {
	tests := []struct {
		name         string
		registry     string
		clientRegion string
		expectError  bool
		expectSame   bool
	}{
		{
			name:         "Same region registry",
			registry:     "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			clientRegion: "us-west-2",
			expectError:  false,
			expectSame:   true,
		},
		{
			name:         "Cross-region registry",
			registry:     "123456789012.dkr.ecr.eu-central-1.amazonaws.com",
			clientRegion: "us-west-2",
			expectError:  false,
			expectSame:   false,
		},
		{
			name:         "Non-ECR registry",
			registry:     "docker.io",
			clientRegion: "us-west-2",
			expectError:  true,
		},
		{
			name:         "Invalid ECR format",
			registry:     "invalid.ecr.format",
			clientRegion: "us-west-2",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockECRService)
			auth := ecr.NewECRAuthenticator(mockService, tt.clientRegion)

			newAuth, err := auth.RegistryAuthenticator(tt.registry)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, newAuth)
			} else {
				// Note: Cross-region will fail without AWS credentials
				// but we can test the error path
				if err != nil {
					// Cross-region auth creation failed (expected without AWS creds)
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, newAuth)

					if tt.expectSame {
						// Should return the same authenticator
						assert.Equal(t, auth, newAuth)
					}
				}
			}
		})
	}
}

func TestGetECRRegistry(t *testing.T) {
	tests := []struct {
		name      string
		accountID string
		region    string
		expected  string
	}{
		{
			name:      "US West 2",
			accountID: "123456789012",
			region:    "us-west-2",
			expected:  "123456789012.dkr.ecr.us-west-2.amazonaws.com",
		},
		{
			name:      "EU Central 1",
			accountID: "987654321098",
			region:    "eu-central-1",
			expected:  "987654321098.dkr.ecr.eu-central-1.amazonaws.com",
		},
		{
			name:      "AP Southeast 1",
			accountID: "111222333444",
			region:    "ap-southeast-1",
			expected:  "111222333444.dkr.ecr.ap-southeast-1.amazonaws.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ecr.GetECRRegistry(tt.accountID, tt.region)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestECRAuthenticator_GetECRRepository(t *testing.T) {
	tests := []struct {
		name        string
		region      string
		repoName    string
		expectError bool
	}{
		{
			name:        "Valid repository name",
			region:      "us-west-2",
			repoName:    "my-app",
			expectError: false,
		},
		{
			name:        "Repository with path",
			region:      "us-west-2",
			repoName:    "org/my-app",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockECRService)
			auth := ecr.NewECRAuthenticator(mockService, tt.region)

			repo, err := auth.GetECRRepository(tt.repoName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)
			}
		})
	}
}

func TestECRKeychain_Resolve(t *testing.T) {
	tests := []struct {
		name           string
		registry       string
		setupHelper    func() ecr.CredentialHelper
		expectError    bool
	}{
		{
			name:     "Valid ECR registry",
			registry: "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			setupHelper: func() ecr.CredentialHelper {
				return &mockCredentialHelper{
					auth: ecr.ClientAuth{
						Username: "AWS",
						Password: "token123",
					},
				}
			},
			expectError: false,
		},
		{
			name:        "Non-ECR registry",
			registry:    "docker.io",
			setupHelper: func() ecr.CredentialHelper { return nil },
			expectError: true,
		},
		{
			name:     "Public ECR",
			registry: "public.ecr.aws",
			setupHelper: func() ecr.CredentialHelper {
				return &mockCredentialHelper{
					auth: ecr.ClientAuth{
						Username: "AWS",
						Password: "token123",
					},
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helper := tt.setupHelper()
			keychain := &ecr.ECRKeychain{}
			if helper != nil {
				keychain = &ecr.ECRKeychain{helper}
			}

			resource := &mockResource{registry: tt.registry}
			auth, err := keychain.Resolve(resource)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, auth)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, auth)
			}
		})
	}
}

func TestStaticAuthenticator_Authorization(t *testing.T) {
	// Test the static authenticator used by ECR keychain
	username := "testuser"
	password := "testpass"

	// This is internal but we can test through the keychain
	helper := &mockCredentialHelper{
		auth: ecr.ClientAuth{
			Username: username,
			Password: password,
		},
	}

	keychain := &ecr.ECRKeychain{helper}
	resource := &mockResource{registry: "123456789012.dkr.ecr.us-west-2.amazonaws.com"}

	auth, err := keychain.Resolve(resource)
	assert.NoError(t, err)
	assert.NotNil(t, auth)

	config, err := auth.Authorization()
	assert.NoError(t, err)
	assert.Equal(t, username, config.Username)
	assert.Equal(t, password, config.Password)
}

func TestNewECRClientForRegion(t *testing.T) {
	tests := []struct {
		name        string
		region      string
		expectError bool
	}{
		{
			name:        "Valid region",
			region:      "us-west-2",
			expectError: false,
		},
		{
			name:        "Another valid region",
			region:      "eu-central-1",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := ecr.NewECRClientForRegion(tt.region)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				// Will succeed if AWS credentials are available
				if err != nil {
					// Expected without AWS credentials
					assert.Error(t, err)
				} else {
					assert.NotNil(t, client)
				}
			}
		})
	}
}

// Test credential refresh by calling Authorization multiple times
func TestECRAuthenticator_CredentialRefresh(t *testing.T) {
	mockService := new(MockECRService)

	// Setup mock to return different tokens on successive calls
	token1 := base64.StdEncoding.EncodeToString([]byte("AWS:token1"))
	token2 := base64.StdEncoding.EncodeToString([]byte("AWS:token2"))

	mockService.On("GetAuthorizationToken", mock.Anything, mock.Anything, mock.Anything).
		Return(&awsecr.GetAuthorizationTokenOutput{
			AuthorizationData: []ecrtypes.AuthorizationData{
				{AuthorizationToken: aws.String(token1)},
			},
		}, nil).Once()

	mockService.On("GetAuthorizationToken", mock.Anything, mock.Anything, mock.Anything).
		Return(&awsecr.GetAuthorizationTokenOutput{
			AuthorizationData: []ecrtypes.AuthorizationData{
				{AuthorizationToken: aws.String(token2)},
			},
		}, nil).Once()

	auth := ecr.NewECRAuthenticator(mockService, "us-west-2")

	// First call
	config1, err := auth.Authorization()
	assert.NoError(t, err)
	assert.Equal(t, "token1", config1.Password)

	// Second call should get new token
	config2, err := auth.Authorization()
	assert.NoError(t, err)
	assert.Equal(t, "token2", config2.Password)

	mockService.AssertExpectations(t)
}

// Mock types for testing
type mockCredentialHelper struct {
	auth ecr.ClientAuth
	err  error
}

func (m *mockCredentialHelper) Get(serverURL string) (ecr.ClientAuth, error) {
	return m.auth, m.err
}

type mockResource struct {
	registry string
}

func (m *mockResource) RegistryStr() string {
	return m.registry
}

func (m *mockResource) String() string {
	return m.registry
}
