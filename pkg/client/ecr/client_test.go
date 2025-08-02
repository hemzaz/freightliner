package ecr

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"freightliner/pkg/helper/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsecr "github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for ECR API
type mockECRAPI struct {
	mock.Mock
}

func (m *mockECRAPI) DescribeRepositories(ctx context.Context, params *awsecr.DescribeRepositoriesInput, optFns ...func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*awsecr.DescribeRepositoriesOutput), args.Error(1)
}

func (m *mockECRAPI) GetAuthorizationToken(ctx context.Context, params *awsecr.GetAuthorizationTokenInput, optFns ...func(*awsecr.Options)) (*awsecr.GetAuthorizationTokenOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*awsecr.GetAuthorizationTokenOutput), args.Error(1)
}

// Mock for STS API
type mockSTSAPI struct {
	mock.Mock
}

func (m *mockSTSAPI) GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*sts.GetCallerIdentityOutput), args.Error(1)
}

// Mock for ECR credential helper
type mockECRCredentialHelper struct {
	mock.Mock
}

// TestClientAuth is a simple struct to hold ECR auth credentials for tests
type TestClientAuth struct {
	Username string
	Password string
}

func (m *mockECRCredentialHelper) GetCredentials(serverURL string) (TestClientAuth, error) {
	args := m.Called(serverURL)
	return args.Get(0).(TestClientAuth), args.Error(1)
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		region      string
		accountID   string
		registry    string
		mockSetup   func(*mockSTSAPI)
		expectedErr bool
	}{
		{
			name:      "With explicit account ID",
			region:    "us-west-2",
			accountID: "123456789012",
			registry:  "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			mockSetup: func(mockSTS *mockSTSAPI) {
				// No STS calls expected with explicit account ID
			},
			expectedErr: false,
		},
		{
			name:      "Auto-detect account ID",
			region:    "us-west-2",
			accountID: "",
			registry:  "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			mockSetup: func(mockSTS *mockSTSAPI) {
				mockSTS.On("GetCallerIdentity", mock.Anything, mock.Anything, mock.Anything).
					Return(&sts.GetCallerIdentityOutput{
						Account: aws.String("123456789012"),
					}, nil)
			},
			expectedErr: false,
		},
		{
			name:      "STS error",
			region:    "us-west-2",
			accountID: "",
			registry:  "",
			mockSetup: func(mockSTS *mockSTSAPI) {
				mockSTS.On("GetCallerIdentity", mock.Anything, mock.Anything, mock.Anything).
					Return(&sts.GetCallerIdentityOutput{}, errors.New("STS error"))
			},
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSTS := &mockSTSAPI{}
			tc.mockSetup(mockSTS)

			// Skip the test unless explicitly enabled via environment variable
			if os.Getenv("ENABLE_ECR_INTEGRATION_TESTS") != "true" {
				t.Skip("Skipping ECR integration test. Set ENABLE_ECR_INTEGRATION_TESTS=true to run.")
			}

			client, err := NewClient(ClientOptions{
				Region:    tc.region,
				AccountID: tc.accountID,
				Logger:    log.NewLogger(),
			})

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tc.registry, client.GetRegistryName())
				assert.Equal(t, tc.region, client.region)
			}

			mockSTS.AssertExpectations(t)
		})
	}
}

// parseECRRepository parses ECR repository references
func parseECRRepository(input string) (string, string, error) {
	// Split by '/' to separate registry from repository
	parts := strings.Split(input, "/")

	// If no '/' or only one part, it's just a repository name
	if len(parts) == 1 {
		return "", input, nil
	}

	// Check if first part looks like an ECR registry
	if strings.Contains(parts[0], "dkr.ecr") && strings.Contains(parts[0], ".amazonaws.com") {
		// It's a full ECR URI
		registry := parts[0]
		repository := strings.Join(parts[1:], "/")
		return registry, repository, nil
	}

	// Otherwise it's just a repository path
	return "", input, nil
}

func TestParseECRRepository(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedRegistry   string
		expectedRepository string
		expectedErr        bool
	}{
		{
			name:               "Full ECR URI",
			input:              "123456789012.dkr.ecr.us-west-2.amazonaws.com/repo-name",
			expectedRegistry:   "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			expectedRepository: "repo-name",
			expectedErr:        false,
		},
		{
			name:               "Simple repository name",
			input:              "repo-name",
			expectedRegistry:   "",
			expectedRepository: "repo-name",
			expectedErr:        false,
		},
		{
			name:               "Repository with path",
			input:              "org/repo-name",
			expectedRegistry:   "",
			expectedRepository: "org/repo-name",
			expectedErr:        false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			registry, repository, err := parseECRRepository(tc.input)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedRegistry, registry)
				assert.Equal(t, tc.expectedRepository, repository)
			}
		})
	}
}
