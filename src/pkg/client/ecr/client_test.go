package ecr

import (
	"context"
	"errors"
	"src/pkg/client/common"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/awslabs/amazon-ecr-credential-helper/ecr-login"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for ECR API
type mockECRAPI struct {
	mock.Mock
}

func (m *mockECRAPI) DescribeRepositories(ctx context.Context, params *ecr.DescribeRepositoriesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ecr.DescribeRepositoriesOutput), args.Error(1)
}

func (m *mockECRAPI) GetAuthorizationToken(ctx context.Context, params *ecr.GetAuthorizationTokenInput, optFns ...func(*ecr.Options)) (*ecr.GetAuthorizationTokenOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ecr.GetAuthorizationTokenOutput), args.Error(1)
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

func (m *mockECRCredentialHelper) GetCredentials(serverURL string) (ecrapi.Auth, error) {
	args := m.Called(serverURL)
	return args.Get(0).(ecrapi.Auth), args.Error(1)
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

			client, err := NewClient(Options{
				Region:    tc.region,
				AccountID: tc.accountID,
				stsClient: mockSTS,
			})

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tc.registry, client.registry)
				assert.Equal(t, tc.region, client.region)
			}

			mockSTS.AssertExpectations(t)
		})
	}
}

func TestClientListRepositories(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*mockECRAPI)
		expected    []string
		expectedErr bool
	}{
		{
			name: "Successful list",
			mockSetup: func(mockECR *mockECRAPI) {
				mockECR.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.DescribeRepositoriesOutput{
						Repositories: []types.Repository{
							{RepositoryName: aws.String("repo1")},
							{RepositoryName: aws.String("repo2")},
						},
					}, nil)
			},
			expected:    []string{"repo1", "repo2"},
			expectedErr: false,
		},
		{
			name: "Empty list",
			mockSetup: func(mockECR *mockECRAPI) {
				mockECR.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.DescribeRepositoriesOutput{
						Repositories: []types.Repository{},
					}, nil)
			},
			expected:    []string{},
			expectedErr: false,
		},
		{
			name: "API error",
			mockSetup: func(mockECR *mockECRAPI) {
				mockECR.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.DescribeRepositoriesOutput{}, errors.New("API error"))
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockECR := &mockECRAPI{}
			tc.mockSetup(mockECR)

			client := &Client{
				region:    "us-west-2",
				registry:  "123456789012.dkr.ecr.us-west-2.amazonaws.com",
				ecrClient: mockECR,
			}

			repos, err := client.ListRepositories()
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, repos)
			}

			mockECR.AssertExpectations(t)
		})
	}
}

func TestClientGetRepository(t *testing.T) {
	tests := []struct {
		name             string
		repoName         string
		mockSetup        func(*mockECRAPI)
		expectedRepoName string
		expectedErr      bool
		expectedErrType  error
	}{
		{
			name:     "Existing repository",
			repoName: "existing-repo",
			mockSetup: func(mockECR *mockECRAPI) {
				mockECR.On("DescribeRepositories", mock.Anything, &ecr.DescribeRepositoriesInput{
					RepositoryNames: []string{"existing-repo"},
				}, mock.Anything).
					Return(&ecr.DescribeRepositoriesOutput{
						Repositories: []types.Repository{
							{
								RepositoryName: aws.String("existing-repo"),
								RepositoryUri:  aws.String("123456789012.dkr.ecr.us-west-2.amazonaws.com/existing-repo"),
							},
						},
					}, nil)
			},
			expectedRepoName: "existing-repo",
			expectedErr:      false,
		},
		{
			name:     "Non-existent repository",
			repoName: "non-existent-repo",
			mockSetup: func(mockECR *mockECRAPI) {
				mockECR.On("DescribeRepositories", mock.Anything, &ecr.DescribeRepositoriesInput{
					RepositoryNames: []string{"non-existent-repo"},
				}, mock.Anything).
					Return(&ecr.DescribeRepositoriesOutput{}, &types.RepositoryNotFoundException{})
			},
			expectedRepoName: "",
			expectedErr:      true,
			expectedErrType:  common.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockECR := &mockECRAPI{}
			tc.mockSetup(mockECR)

			client := &Client{
				region:    "us-west-2",
				registry:  "123456789012.dkr.ecr.us-west-2.amazonaws.com",
				ecrClient: mockECR,
			}

			repo, err := client.GetRepository(tc.repoName)
			if tc.expectedErr {
				assert.Error(t, err)
				if tc.expectedErrType != nil {
					assert.True(t, errors.Is(err, tc.expectedErrType))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)
				assert.Equal(t, tc.expectedRepoName, repo.GetRepositoryName())
			}

			mockECR.AssertExpectations(t)
		})
	}
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
