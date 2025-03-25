package ecr

import (
	"encoding/base64"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/awslabs/amazon-ecr-credential-helper/ecr-login"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for ECR Auth API
type mockECRAuthAPI struct {
	mock.Mock
}

func (m *mockECRAuthAPI) GetAuthorizationToken(ctx interface{}, params *ecr.GetAuthorizationTokenInput, optFns ...interface{}) (*ecr.GetAuthorizationTokenOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ecr.GetAuthorizationTokenOutput), args.Error(1)
}

func TestECRAuthenticatorAuthorization(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*mockECRAuthAPI)
		expectedAuth  *authn.AuthConfig
		expectedErr   bool
	}{
		{
			name: "Successful authentication",
			mockSetup: func(mockECR *mockECRAuthAPI) {
				// Create a base64 encoded "username:password" string
				authToken := base64.StdEncoding.EncodeToString([]byte("AWS:password123"))
				mockECR.On("GetAuthorizationToken", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.GetAuthorizationTokenOutput{
						AuthorizationData: []types.AuthorizationData{
							{
								AuthorizationToken: aws.String(authToken),
								ProxyEndpoint:      aws.String("https://123456789012.dkr.ecr.us-west-2.amazonaws.com"),
							},
						},
					}, nil)
			},
			expectedAuth: &authn.AuthConfig{
				Username: "AWS",
				Password: "password123",
			},
			expectedErr: false,
		},
		{
			name: "API error",
			mockSetup: func(mockECR *mockECRAuthAPI) {
				mockECR.On("GetAuthorizationToken", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.GetAuthorizationTokenOutput{}, errors.New("API error"))
			},
			expectedAuth: nil,
			expectedErr:  true,
		},
		{
			name: "Empty authorization data",
			mockSetup: func(mockECR *mockECRAuthAPI) {
				mockECR.On("GetAuthorizationToken", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.GetAuthorizationTokenOutput{
						AuthorizationData: []types.AuthorizationData{},
					}, nil)
			},
			expectedAuth: nil,
			expectedErr:  true,
		},
		{
			name: "Invalid token format",
			mockSetup: func(mockECR *mockECRAuthAPI) {
				// Not a valid base64 encoded string
				mockECR.On("GetAuthorizationToken", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.GetAuthorizationTokenOutput{
						AuthorizationData: []types.AuthorizationData{
							{
								AuthorizationToken: aws.String("not-base64"),
								ProxyEndpoint:      aws.String("https://123456789012.dkr.ecr.us-west-2.amazonaws.com"),
							},
						},
					}, nil)
			},
			expectedAuth: nil,
			expectedErr:  true,
		},
		{
			name: "Invalid credentials format",
			mockSetup: func(mockECR *mockECRAuthAPI) {
				// Base64 encoded string that doesn't contain a colon
				authToken := base64.StdEncoding.EncodeToString([]byte("no-colon-separator"))
				mockECR.On("GetAuthorizationToken", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.GetAuthorizationTokenOutput{
						AuthorizationData: []types.AuthorizationData{
							{
								AuthorizationToken: aws.String(authToken),
								ProxyEndpoint:      aws.String("https://123456789012.dkr.ecr.us-west-2.amazonaws.com"),
							},
						},
					}, nil)
			},
			expectedAuth: nil,
			expectedErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockECR := &mockECRAuthAPI{}
			tc.mockSetup(mockECR)

			auth := &ECRAuthenticator{
				registry:  "123456789012.dkr.ecr.us-west-2.amazonaws.com",
				ecrClient: mockECR,
			}

			config, err := auth.Authorization()
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedAuth.Username, config.Username)
				assert.Equal(t, tc.expectedAuth.Password, config.Password)
			}

			mockECR.AssertExpectations(t)
		})
	}
}

// Mock for ECR credential helper
type mockCredentialHelper struct {
	mock.Mock
}

func (m *mockCredentialHelper) Get(serverURL string) (ecrapi.Auth, error) {
	args := m.Called(serverURL)
	return args.Get(0).(ecrapi.Auth), args.Error(1)
}

func TestECRKeychainResolve(t *testing.T) {
	tests := []struct {
		name        string
		resource    authn.Resource
		mockSetup   func(*mockCredentialHelper)
		expectedAuth *authn.AuthConfig
		expectedErr bool
	}{
		{
			name: "Successful ECR resolution",
			resource: &authn.Resource{
				RegistryStr: "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			},
			mockSetup: func(mockHelper *mockCredentialHelper) {
				mockHelper.On("Get", "https://123456789012.dkr.ecr.us-west-2.amazonaws.com").
					Return(ecrapi.Auth{
						Username: "AWS",
						Password: "password123",
					}, nil)
			},
			expectedAuth: &authn.AuthConfig{
				Username: "AWS",
				Password: "password123",
			},
			expectedErr: false,
		},
		{
			name: "Non-ECR registry",
			resource: &authn.Resource{
				RegistryStr: "docker.io",
			},
			mockSetup: func(mockHelper *mockCredentialHelper) {
				// Should not be called
			},
			expectedAuth: nil,
			expectedErr:  true,
		},
		{
			name: "Credential helper error",
			resource: &authn.Resource{
				RegistryStr: "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			},
			mockSetup: func(mockHelper *mockCredentialHelper) {
				mockHelper.On("Get", "https://123456789012.dkr.ecr.us-west-2.amazonaws.com").
					Return(ecrapi.Auth{}, errors.New("credential helper error"))
			},
			expectedAuth: nil,
			expectedErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockHelper := &mockCredentialHelper{}
			tc.mockSetup(mockHelper)

			keychain := &ECRKeychain{
				helper: mockHelper,
			}

			authenticator, err := keychain.Resolve(tc.resource)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Get the auth config from the authenticator
				config, err := authenticator.Authorization()
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedAuth.Username, config.Username)
				assert.Equal(t, tc.expectedAuth.Password, config.Password)
			}

			mockHelper.AssertExpectations(t)
		})
	}
}

func TestIsECRRegistry(t *testing.T) {
	tests := []struct {
		name     string
		registry string
		expected bool
	}{
		{
			name:     "Valid ECR registry",
			registry: "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			expected: true,
		},
		{
			name:     "Valid ECR registry in another region",
			registry: "123456789012.dkr.ecr.eu-central-1.amazonaws.com",
			expected: true,
		},
		{
			name:     "Valid ECR registry with public suffix",
			registry: "public.ecr.aws",
			expected: true,
		},
		{
			name:     "Docker Hub",
			registry: "docker.io",
			expected: false,
		},
		{
			name:     "GCR registry",
			registry: "gcr.io",
			expected: false,
		},
		{
			name:     "Empty string",
			registry: "",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isECRRegistry(tc.registry)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetECRRegistry(t *testing.T) {
	tests := []struct {
		name       string
		accountID  string
		region     string
		expected   string
	}{
		{
			name:      "Valid account and region",
			accountID: "123456789012",
			region:    "us-west-2",
			expected:  "123456789012.dkr.ecr.us-west-2.amazonaws.com",
		},
		{
			name:      "Different region",
			accountID: "123456789012",
			region:    "eu-central-1",
			expected:  "123456789012.dkr.ecr.eu-central-1.amazonaws.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := GetECRRegistry(tc.accountID, tc.region)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExampleECRRepository(t *testing.T) {
	tests := []struct {
		name       string
		registry   string
		repoName   string
		expected   string
	}{
		{
			name:      "Simple repository name",
			registry:  "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			repoName:  "test-repo",
			expected:  "123456789012.dkr.ecr.us-west-2.amazonaws.com/test-repo",
		},
		{
			name:      "Repository with path",
			registry:  "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			repoName:  "org/test-repo",
			expected:  "123456789012.dkr.ecr.us-west-2.amazonaws.com/org/test-repo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := fmt.Sprintf("%s/%s", tc.registry, tc.repoName)
			assert.Equal(t, tc.expected, result)
		})
	}
}