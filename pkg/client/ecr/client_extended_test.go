package ecr

import (
	"context"
	"testing"

	"freightliner/pkg/helper/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsecr "github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockECRServiceExt extends the existing mock for additional testing
type MockECRServiceExt struct {
	mock.Mock
}

func (m *MockECRServiceExt) ListImages(ctx context.Context, params *awsecr.ListImagesInput, optFns ...func(*awsecr.Options)) (*awsecr.ListImagesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.ListImagesOutput), args.Error(1)
}

func (m *MockECRServiceExt) BatchGetImage(ctx context.Context, params *awsecr.BatchGetImageInput, optFns ...func(*awsecr.Options)) (*awsecr.BatchGetImageOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.BatchGetImageOutput), args.Error(1)
}

func (m *MockECRServiceExt) PutImage(ctx context.Context, params *awsecr.PutImageInput, optFns ...func(*awsecr.Options)) (*awsecr.PutImageOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.PutImageOutput), args.Error(1)
}

func (m *MockECRServiceExt) BatchDeleteImage(ctx context.Context, params *awsecr.BatchDeleteImageInput, optFns ...func(*awsecr.Options)) (*awsecr.BatchDeleteImageOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.BatchDeleteImageOutput), args.Error(1)
}

func (m *MockECRServiceExt) DescribeImages(ctx context.Context, params *awsecr.DescribeImagesInput, optFns ...func(*awsecr.Options)) (*awsecr.DescribeImagesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.DescribeImagesOutput), args.Error(1)
}

func (m *MockECRServiceExt) DescribeRepositories(ctx context.Context, params *awsecr.DescribeRepositoriesInput, optFns ...func(*awsecr.Options)) (*awsecr.DescribeRepositoriesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.DescribeRepositoriesOutput), args.Error(1)
}

func (m *MockECRServiceExt) CreateRepository(ctx context.Context, params *awsecr.CreateRepositoryInput, optFns ...func(*awsecr.Options)) (*awsecr.CreateRepositoryOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.CreateRepositoryOutput), args.Error(1)
}

func (m *MockECRServiceExt) GetAuthorizationToken(ctx context.Context, params *awsecr.GetAuthorizationTokenInput, optFns ...func(*awsecr.Options)) (*awsecr.GetAuthorizationTokenOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*awsecr.GetAuthorizationTokenOutput), args.Error(1)
}

func TestClientExtended_GetRegistryName(t *testing.T) {
	tests := []struct {
		name      string
		region    string
		accountID string
		expected  string
	}{
		{
			name:      "Standard registry",
			region:    "us-west-2",
			accountID: "123456789012",
			expected:  "123456789012.dkr.ecr.us-west-2.amazonaws.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockECRServiceExt)
			client := &Client{
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

func TestClientExtended_ListRepositoriesWithPagination(t *testing.T) {
	mockService := new(MockECRServiceExt)

	// First page
	mockService.On("DescribeRepositories", mock.Anything, mock.MatchedBy(func(input *awsecr.DescribeRepositoriesInput) bool {
		return input.NextToken == nil
	}), mock.Anything).
		Return(&awsecr.DescribeRepositoriesOutput{
			Repositories: []ecrtypes.Repository{
				{RepositoryName: aws.String("repo1")},
				{RepositoryName: aws.String("repo2")},
			},
			NextToken: aws.String("token1"),
		}, nil).Once()

	// Second page
	mockService.On("DescribeRepositories", mock.Anything, mock.MatchedBy(func(input *awsecr.DescribeRepositoriesInput) bool {
		return input.NextToken != nil && *input.NextToken == "token1"
	}), mock.Anything).
		Return(&awsecr.DescribeRepositoriesOutput{
			Repositories: []ecrtypes.Repository{
				{RepositoryName: aws.String("repo3")},
			},
			NextToken: nil,
		}, nil).Once()

	client := &Client{
		ecr:       mockService,
		region:    "us-west-2",
		accountID: "123456789012",
		logger:    log.NewBasicLogger(log.InfoLevel),
	}

	ctx := context.Background()
	repos, err := client.ListRepositories(ctx, "")

	assert.NoError(t, err)
	assert.Len(t, repos, 3)
	assert.Equal(t, []string{"repo1", "repo2", "repo3"}, repos)
	mockService.AssertExpectations(t)
}

func TestClientExtended_ListRepositoriesWithPrefix(t *testing.T) {
	mockService := new(MockECRServiceExt)

	mockService.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).
		Return(&awsecr.DescribeRepositoriesOutput{
			Repositories: []ecrtypes.Repository{
				{RepositoryName: aws.String("app-backend")},
				{RepositoryName: aws.String("app-frontend")},
				{RepositoryName: aws.String("lib-common")},
			},
			NextToken: nil,
		}, nil)

	client := &Client{
		ecr:       mockService,
		region:    "us-west-2",
		accountID: "123456789012",
		logger:    log.NewBasicLogger(log.InfoLevel),
	}

	ctx := context.Background()
	repos, err := client.ListRepositories(ctx, "app-")

	assert.NoError(t, err)
	assert.Len(t, repos, 2)
	assert.Contains(t, repos, "app-backend")
	assert.Contains(t, repos, "app-frontend")
	assert.NotContains(t, repos, "lib-common")
	mockService.AssertExpectations(t)
}

func TestClientExtended_GetRemoteOptions(t *testing.T) {
	mockService := new(MockECRServiceExt)
	client := &Client{
		ecr:       mockService,
		region:    "us-west-2",
		accountID: "123456789012",
		logger:    log.NewBasicLogger(log.InfoLevel),
	}

	opts := client.GetRemoteOptions()
	assert.NotNil(t, opts)
	assert.Len(t, opts, 1)
}

func TestClientExtended_CreateRepositoryWithAccountID(t *testing.T) {
	mockService := new(MockECRServiceExt)

	repoName := "test-repo"
	repoArn := "arn:aws:ecr:us-west-2:123456789012:repository/test-repo"

	mockService.On("CreateRepository", mock.Anything, mock.MatchedBy(func(input *awsecr.CreateRepositoryInput) bool {
		return *input.RepositoryName == repoName &&
			input.RegistryId != nil &&
			*input.RegistryId == "123456789012"
	}), mock.Anything).
		Return(&awsecr.CreateRepositoryOutput{
			Repository: &ecrtypes.Repository{
				RepositoryName: aws.String(repoName),
				RepositoryArn:  aws.String(repoArn),
			},
		}, nil)

	client := &Client{
		ecr:       mockService,
		region:    "us-west-2",
		accountID: "123456789012",
		logger:    log.NewBasicLogger(log.InfoLevel),
	}

	ctx := context.Background()
	repo, err := client.CreateRepository(ctx, repoName, nil)

	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, repoName, repo.GetRepositoryName())
	mockService.AssertExpectations(t)
}

func TestClientExtended_CreateRepositoryWithTags(t *testing.T) {
	mockService := new(MockECRServiceExt)

	repoName := "test-repo"
	tags := map[string]string{
		"Environment": "production",
		"Team":        "backend",
	}

	repoArn := "arn:aws:ecr:us-west-2:123456789012:repository/" + repoName
	mockService.On("CreateRepository", mock.Anything, mock.MatchedBy(func(input *awsecr.CreateRepositoryInput) bool {
		if *input.RepositoryName != repoName {
			return false
		}
		// Verify tags are present
		return len(input.Tags) == 2
	}), mock.Anything).
		Return(&awsecr.CreateRepositoryOutput{
			Repository: &ecrtypes.Repository{
				RepositoryName: aws.String(repoName),
				RepositoryArn:  aws.String(repoArn),
			},
		}, nil)

	client := &Client{
		ecr:       mockService,
		region:    "us-west-2",
		accountID: "123456789012",
		logger:    log.NewBasicLogger(log.InfoLevel),
	}

	ctx := context.Background()
	_, err := client.CreateRepository(ctx, repoName, tags)

	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

func TestNormalizeClientOptions(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
	}{
		{
			name: "Valid ClientOptions",
			input: ClientOptions{
				Region:    "us-west-2",
				AccountID: "123456789012",
			},
			expectError: false,
		},
		{
			name: "Valid LegacyClientOptions",
			input: LegacyClientOptions{
				Region:    "us-west-2",
				AccountID: "123456789012",
			},
			expectError: false,
		},
		{
			name:        "Invalid type",
			input:       "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := normalizeClientOptions(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, opts.Region)
			}
		})
	}
}

func TestValidateClientOptions(t *testing.T) {
	tests := []struct {
		name        string
		opts        ClientOptions
		expectError bool
	}{
		{
			name: "Valid options",
			opts: ClientOptions{
				Region:    "us-west-2",
				AccountID: "123456789012",
				Logger:    log.NewBasicLogger(log.InfoLevel),
			},
			expectError: false,
		},
		{
			name: "Missing region",
			opts: ClientOptions{
				AccountID: "123456789012",
			},
			expectError: true,
		},
		{
			name: "Nil logger gets set",
			opts: ClientOptions{
				Region:    "us-west-2",
				AccountID: "123456789012",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateClientOptions(&tt.opts)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tt.opts.Logger)
			}
		})
	}
}
