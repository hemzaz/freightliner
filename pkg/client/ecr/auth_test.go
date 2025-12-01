package ecr

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsecr "github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for ECR Auth API
type mockECRAuthAPI struct {
	mock.Mock
}

func (m *mockECRAuthAPI) GetAuthorizationToken(ctx context.Context, params *awsecr.GetAuthorizationTokenInput, optFns ...func(*awsecr.Options)) (*awsecr.GetAuthorizationTokenOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.GetAuthorizationTokenOutput), args.Error(1)
}

// Implement the rest of the ECRAPI interface

func (m *mockECRAuthAPI) ListImages(ctx context.Context, params *awsecr.ListImagesInput, optFns ...func(*awsecr.Options)) (*awsecr.ListImagesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.ListImagesOutput), args.Error(1)
}

func (m *mockECRAuthAPI) BatchGetImage(ctx context.Context, params *awsecr.BatchGetImageInput, optFns ...func(*awsecr.Options)) (*awsecr.BatchGetImageOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.BatchGetImageOutput), args.Error(1)
}

func (m *mockECRAuthAPI) PutImage(ctx context.Context, params *awsecr.PutImageInput, optFns ...func(*awsecr.Options)) (*awsecr.PutImageOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.PutImageOutput), args.Error(1)
}

func (m *mockECRAuthAPI) BatchDeleteImage(ctx context.Context, params *awsecr.BatchDeleteImageInput, optFns ...func(*awsecr.Options)) (*awsecr.BatchDeleteImageOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.BatchDeleteImageOutput), args.Error(1)
}

func (m *mockECRAuthAPI) DescribeImages(ctx context.Context, params *awsecr.DescribeImagesInput, optFns ...func(*awsecr.Options)) (*awsecr.DescribeImagesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.DescribeImagesOutput), args.Error(1)
}

func (m *mockECRAuthAPI) DescribeRepositories(ctx context.Context, params *awsecr.DescribeRepositoriesInput, optFns ...func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.DescribeRepositoriesOutput), args.Error(1)
}

func (m *mockECRAuthAPI) CreateRepository(ctx context.Context, params *awsecr.CreateRepositoryInput, optFns ...func(*awsecr.Options)) (*awsecr.CreateRepositoryOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.CreateRepositoryOutput), args.Error(1)
}

// getTestAuthToken returns a test token from environment or generates one
func getTestAuthToken() string {
	token := os.Getenv("TEST_ECR_AUTH_TOKEN")
	if token == "" {
		// Generate a test token dynamically (not a real credential)
		token = base64.StdEncoding.EncodeToString([]byte("AWS:password123"))
	}
	return token
}

func TestECRAuthenticatorAuthorization(t *testing.T) {
	tests := []struct {
		name         string
		mockSetup    func(*mockECRAuthAPI)
		expectedAuth *authn.AuthConfig
		expectedErr  bool
	}{
		{
			name: "Successful authentication",
			mockSetup: func(mockECR *mockECRAuthAPI) {
				authToken := getTestAuthToken()
				mockECR.On("GetAuthorizationToken", mock.Anything, mock.Anything, mock.Anything).
					Return(&awsecr.GetAuthorizationTokenOutput{
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
					Return(&awsecr.GetAuthorizationTokenOutput{}, errors.New("API error"))
			},
			expectedAuth: nil,
			expectedErr:  true,
		},
		{
			name: "Empty authorization data",
			mockSetup: func(mockECR *mockECRAuthAPI) {
				mockECR.On("GetAuthorizationToken", mock.Anything, mock.Anything, mock.Anything).
					Return(&awsecr.GetAuthorizationTokenOutput{
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
					Return(&awsecr.GetAuthorizationTokenOutput{
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
					Return(&awsecr.GetAuthorizationTokenOutput{
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
				registry: "123456789012.dkr.ecr.us-west-2.amazonaws.com",
				client:   mockECR,
				region:   "us-west-2",
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

// Auth is a simple struct to hold ECR auth credentials
type Auth struct {
	Username string
	Password string
}

func (m *mockCredentialHelper) Get(serverURL string) (ClientAuth, error) {
	args := m.Called(serverURL)
	return args.Get(0).(ClientAuth), args.Error(1)
}

// Mock Resource that implements authn.Resource
type resourceMock struct {
	registry string
}

func (r resourceMock) RegistryStr() string {
	return r.registry
}

func (r resourceMock) String() string {
	return r.registry
}

func TestECRKeychainResolve(t *testing.T) {
	tests := []struct {
		name         string
		resource     resourceMock
		mockSetup    func(*mockCredentialHelper)
		expectedAuth *authn.AuthConfig
		expectedErr  bool
	}{
		{
			name: "Successful ECR resolution",
			resource: resourceMock{
				registry: "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			},
			mockSetup: func(mockHelper *mockCredentialHelper) {
				mockHelper.On("Get", "https://123456789012.dkr.ecr.us-west-2.amazonaws.com").
					Return(ClientAuth{
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
			resource: resourceMock{
				registry: "docker.io",
			},
			mockSetup: func(mockHelper *mockCredentialHelper) {
				// Should not be called
			},
			expectedAuth: nil,
			expectedErr:  true,
		},
		{
			name: "Credential helper error",
			resource: resourceMock{
				registry: "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			},
			mockSetup: func(mockHelper *mockCredentialHelper) {
				mockHelper.On("Get", "https://123456789012.dkr.ecr.us-west-2.amazonaws.com").
					Return(ClientAuth{}, errors.New("credential helper error"))
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
		name      string
		accountID string
		region    string
		expected  string
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
		name     string
		registry string
		repoName string
		expected string
	}{
		{
			name:     "Simple repository name",
			registry: "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			repoName: "test-repo",
			expected: "123456789012.dkr.ecr.us-west-2.amazonaws.com/test-repo",
		},
		{
			name:     "Repository with path",
			registry: "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			repoName: "org/test-repo",
			expected: "123456789012.dkr.ecr.us-west-2.amazonaws.com/org/test-repo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := fmt.Sprintf("%s/%s", tc.registry, tc.repoName)
			assert.Equal(t, tc.expected, result)
		})
	}
}
