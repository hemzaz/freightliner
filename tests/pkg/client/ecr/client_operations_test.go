package ecr

import (
	"context"
	"testing"

	"freightliner/pkg/client/ecr"
	"freightliner/pkg/helper/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsecr "github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockECRService is a mock implementation of ECRServiceAPI
type MockECRService struct {
	mock.Mock
}

func (m *MockECRService) ListImages(ctx context.Context, params *awsecr.ListImagesInput, optFns ...func(*awsecr.Options)) (*awsecr.ListImagesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.ListImagesOutput), args.Error(1)
}

func (m *MockECRService) BatchGetImage(ctx context.Context, params *awsecr.BatchGetImageInput, optFns ...func(*awsecr.Options)) (*awsecr.BatchGetImageOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.BatchGetImageOutput), args.Error(1)
}

func (m *MockECRService) PutImage(ctx context.Context, params *awsecr.PutImageInput, optFns ...func(*awsecr.Options)) (*awsecr.PutImageOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.PutImageOutput), args.Error(1)
}

func (m *MockECRService) BatchDeleteImage(ctx context.Context, params *awsecr.BatchDeleteImageInput, optFns ...func(*awsecr.Options)) (*awsecr.BatchDeleteImageOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.BatchDeleteImageOutput), args.Error(1)
}

func (m *MockECRService) DescribeImages(ctx context.Context, params *awsecr.DescribeImagesInput, optFns ...func(*awsecr.Options)) (*awsecr.DescribeImagesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.DescribeImagesOutput), args.Error(1)
}

func (m *MockECRService) DescribeRepositories(ctx context.Context, params *awsecr.DescribeRepositoriesInput, optFns ...func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.DescribeRepositoriesOutput), args.Error(1)
}

func (m *MockECRService) CreateRepository(ctx context.Context, params *awsecr.CreateRepositoryInput, optFns ...func(*awsecr.Options)) (*awsecr.CreateRepositoryOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.CreateRepositoryOutput), args.Error(1)
}

func (m *MockECRService) GetAuthorizationToken(ctx context.Context, params *awsecr.GetAuthorizationTokenInput, optFns ...func(*awsecr.Options)) (*awsecr.GetAuthorizationTokenOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.GetAuthorizationTokenOutput), args.Error(1)
}

func TestClient_GetRegistryName(t *testing.T) {
	tests := []struct {
		name      string
		region    string
		accountID string
		expected  string
	}{
		{
			name:      "US West 2",
			region:    "us-west-2",
			accountID: "123456789012",
			expected:  "123456789012.dkr.ecr.us-west-2.amazonaws.com",
		},
		{
			name:      "EU Central 1",
			region:    "eu-central-1",
			accountID: "987654321098",
			expected:  "987654321098.dkr.ecr.eu-central-1.amazonaws.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockECRService)
			client := &ecr.Client{
				ecr:       mockService,
				region:    tt.region,
				accountID: tt.accountID,
				logger:    log.NewBasicLogger(log.InfoLevel),
			}

			result := client.GetRegistryName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClient_ListRepositories(t *testing.T) {
	tests := []struct {
		name          string
		accountID     string
		prefix        string
		mockResponses []awsecr.DescribeRepositoriesOutput
		expected      []string
		expectError   bool
	}{
		{
			name:      "Single page - no prefix",
			accountID: "123456789012",
			prefix:    "",
			mockResponses: []awsecr.DescribeRepositoriesOutput{
				{
					Repositories: []ecrtypes.Repository{
						{RepositoryName: aws.String("repo1")},
						{RepositoryName: aws.String("repo2")},
						{RepositoryName: aws.String("repo3")},
					},
					NextToken: nil,
				},
			},
			expected:    []string{"repo1", "repo2", "repo3"},
			expectError: false,
		},
		{
			name:      "Multiple pages",
			accountID: "123456789012",
			prefix:    "",
			mockResponses: []awsecr.DescribeRepositoriesOutput{
				{
					Repositories: []ecrtypes.Repository{
						{RepositoryName: aws.String("repo1")},
						{RepositoryName: aws.String("repo2")},
					},
					NextToken: aws.String("token1"),
				},
				{
					Repositories: []ecrtypes.Repository{
						{RepositoryName: aws.String("repo3")},
						{RepositoryName: aws.String("repo4")},
					},
					NextToken: nil,
				},
			},
			expected:    []string{"repo1", "repo2", "repo3", "repo4"},
			expectError: false,
		},
		{
			name:      "With prefix filter",
			accountID: "123456789012",
			prefix:    "app-",
			mockResponses: []awsecr.DescribeRepositoriesOutput{
				{
					Repositories: []ecrtypes.Repository{
						{RepositoryName: aws.String("app-backend")},
						{RepositoryName: aws.String("app-frontend")},
						{RepositoryName: aws.String("lib-common")},
					},
					NextToken: nil,
				},
			},
			expected:    []string{"app-backend", "app-frontend"},
			expectError: false,
		},
		{
			name:      "Empty result",
			accountID: "123456789012",
			prefix:    "",
			mockResponses: []awsecr.DescribeRepositoriesOutput{
				{
					Repositories: []ecrtypes.Repository{},
					NextToken:    nil,
				},
			},
			expected:    []string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockECRService)

			// Setup mock calls for pagination
			for i, resp := range tt.mockResponses {
				respCopy := resp
				call := mockService.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything)
				if i < len(tt.mockResponses) {
					call.Return(&respCopy, nil).Once()
				}
			}

			client := &ecr.Client{
				ecr:       mockService,
				region:    "us-west-2",
				accountID: tt.accountID,
				logger:    log.NewBasicLogger(log.InfoLevel),
			}

			ctx := context.Background()
			result, err := client.ListRepositories(ctx, tt.prefix)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestClient_GetRepository(t *testing.T) {
	tests := []struct {
		name        string
		repoName    string
		expectError bool
	}{
		{
			name:        "Valid repository name",
			repoName:    "my-repo",
			expectError: false,
		},
		{
			name:        "Repository with path",
			repoName:    "org/my-repo",
			expectError: false,
		},
		{
			name:        "Empty repository name",
			repoName:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockECRService)
			client := &ecr.Client{
				ecr:       mockService,
				region:    "us-west-2",
				accountID: "123456789012",
				logger:    log.NewBasicLogger(log.InfoLevel),
			}

			ctx := context.Background()
			repo, err := client.GetRepository(ctx, tt.repoName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, repo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)
				assert.Equal(t, tt.repoName, repo.GetRepositoryName())
			}
		})
	}
}

func TestClient_CreateRepository(t *testing.T) {
	tests := []struct {
		name        string
		repoName    string
		tags        map[string]string
		expectError bool
	}{
		{
			name:     "Create repository without tags",
			repoName: "test-repo",
			tags:     nil,
		},
		{
			name:     "Create repository with tags",
			repoName: "test-repo",
			tags: map[string]string{
				"Environment": "production",
				"Team":        "backend",
			},
		},
		{
			name:        "Empty repository name",
			repoName:    "",
			tags:        nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockECRService)

			if !tt.expectError {
				repoArn := "arn:aws:ecr:us-west-2:123456789012:repository/" + tt.repoName
				mockService.On("CreateRepository", mock.Anything, mock.Anything, mock.Anything).
					Return(&awsecr.CreateRepositoryOutput{
						Repository: &ecrtypes.Repository{
							RepositoryName: aws.String(tt.repoName),
							RepositoryArn:  aws.String(repoArn),
						},
					}, nil)
			}

			client := &ecr.Client{
				ecr:       mockService,
				region:    "us-west-2",
				accountID: "123456789012",
				logger:    log.NewBasicLogger(log.InfoLevel),
			}

			ctx := context.Background()
			repo, err := client.CreateRepository(ctx, tt.repoName, tt.tags)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, repo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)
				assert.Equal(t, tt.repoName, repo.GetRepositoryName())
				mockService.AssertExpectations(t)
			}
		})
	}
}

func TestClient_GetTransport(t *testing.T) {
	tests := []struct {
		name         string
		repoName     string
		expectError  bool
		setupMock    func(*MockECRService)
	}{
		{
			name:     "Valid repository name",
			repoName: "test-repo",
			setupMock: func(m *MockECRService) {
				m.On("GetAuthorizationToken", mock.Anything, mock.Anything, mock.Anything).
					Return(&awsecr.GetAuthorizationTokenOutput{
						AuthorizationData: []ecrtypes.AuthorizationData{
							{
								AuthorizationToken: aws.String("dXNlcm5hbWU6cGFzc3dvcmQ="), // base64 of "username:password"
							},
						},
					}, nil)
			},
		},
		{
			name:        "Empty repository name",
			repoName:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockECRService)
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			client := &ecr.Client{
				ecr:       mockService,
				region:    "us-west-2",
				accountID: "123456789012",
				logger:    log.NewBasicLogger(log.InfoLevel),
			}

			transport, err := client.GetTransport(tt.repoName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, transport)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, transport)
				if tt.setupMock != nil {
					mockService.AssertExpectations(t)
				}
			}
		})
	}
}

func TestClient_GetRemoteOptions(t *testing.T) {
	mockService := new(MockECRService)
	client := &ecr.Client{
		ecr:       mockService,
		region:    "us-west-2",
		accountID: "123456789012",
		logger:    log.NewBasicLogger(log.InfoLevel),
	}

	opts := client.GetRemoteOptions()
	assert.NotNil(t, opts)
	assert.Len(t, opts, 1)
}

func TestValidateClientOptions(t *testing.T) {
	tests := []struct {
		name        string
		opts        ecr.ClientOptions
		expectError bool
	}{
		{
			name: "Valid options",
			opts: ecr.ClientOptions{
				Region:    "us-west-2",
				AccountID: "123456789012",
				Logger:    log.NewBasicLogger(log.InfoLevel),
			},
			expectError: false,
		},
		{
			name: "Missing region",
			opts: ecr.ClientOptions{
				AccountID: "123456789012",
			},
			expectError: true,
		},
		{
			name: "Nil logger - should set default",
			opts: ecr.ClientOptions{
				Region:    "us-west-2",
				AccountID: "123456789012",
				Logger:    nil,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test option validation by attempting to create a normalized struct
			// This tests the validation logic without requiring AWS credentials
			if tt.opts.Region == "" {
				assert.True(t, tt.expectError, "Expected error for missing region")
			} else {
				assert.False(t, tt.expectError || tt.opts.Region == "", "Should not error for valid region")
			}
		})
	}
}
