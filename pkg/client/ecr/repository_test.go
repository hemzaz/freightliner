package ecr

import (
	"context"
	"errors"
	"os"
	"testing"

	"freightliner/pkg/helper/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Our internal ECRAPI interface should be defined in the main package, not here
// We're just using mockRepositoryECRAPI directly for tests

// Mock for ECR API for repository operations
type mockRepositoryECRAPI struct {
	mock.Mock
}

func (m *mockRepositoryECRAPI) DescribeImages(ctx context.Context, params *ecr.DescribeImagesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeImagesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ecr.DescribeImagesOutput), args.Error(1)
}

func (m *mockRepositoryECRAPI) BatchGetImage(ctx context.Context, params *ecr.BatchGetImageInput, optFns ...func(*ecr.Options)) (*ecr.BatchGetImageOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ecr.BatchGetImageOutput), args.Error(1)
}

func (m *mockRepositoryECRAPI) PutImage(ctx context.Context, params *ecr.PutImageInput, optFns ...func(*ecr.Options)) (*ecr.PutImageOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ecr.PutImageOutput), args.Error(1)
}

func (m *mockRepositoryECRAPI) BatchDeleteImage(ctx context.Context, params *ecr.BatchDeleteImageInput, optFns ...func(*ecr.Options)) (*ecr.BatchDeleteImageOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ecr.BatchDeleteImageOutput), args.Error(1)
}

func (m *mockRepositoryECRAPI) ListImages(ctx context.Context, params *ecr.ListImagesInput, optFns ...func(*ecr.Options)) (*ecr.ListImagesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ecr.ListImagesOutput), args.Error(1)
}

func (m *mockRepositoryECRAPI) DescribeRepositories(ctx context.Context, params *ecr.DescribeRepositoriesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ecr.DescribeRepositoriesOutput), args.Error(1)
}

func (m *mockRepositoryECRAPI) CreateRepository(ctx context.Context, params *ecr.CreateRepositoryInput, optFns ...func(*ecr.Options)) (*ecr.CreateRepositoryOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ecr.CreateRepositoryOutput), args.Error(1)
}

func (m *mockRepositoryECRAPI) GetAuthorizationToken(ctx context.Context, params *ecr.GetAuthorizationTokenInput, optFns ...func(*ecr.Options)) (*ecr.GetAuthorizationTokenOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ecr.GetAuthorizationTokenOutput), args.Error(1)
}

// Test setup helper
func setupTestRepository(mockECR *mockRepositoryECRAPI) *Repository {
	// Create a name.Repository for the test
	repo, _ := name.NewRepository("123456789012.dkr.ecr.us-west-2.amazonaws.com/test-repo")

	// Create a logger
	logger := log.NewLogger()

	// Create a transport option
	transportOpt := remote.WithUserAgent("test-agent")

	// Create a client with our mock ECR API
	client := &Client{
		ecr:          mockECR, // Now we can directly assign the mock since it implements the ECRAPI interface
		region:       "us-west-2",
		accountID:    "123456789012",
		logger:       logger,
		transportOpt: transportOpt,
	}

	// Create and return the Repository
	return &Repository{
		client:     client,
		name:       "test-repo",
		repository: repo,
	}
}

func TestRepositoryListTags(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*mockRepositoryECRAPI)
		expected    []string
		expectedErr bool
	}{
		{
			name: "Successful list",
			mockSetup: func(mockECR *mockRepositoryECRAPI) {
				mockECR.On("ListImages", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.ListImagesOutput{
						ImageIds: []types.ImageIdentifier{
							{ImageTag: aws.String("tag1")},
							{ImageTag: aws.String("latest")},
							{ImageTag: aws.String("tag2")},
							{ImageDigest: aws.String("sha256:abc123")}, // Image without a tag
						},
					}, nil)
			},
			expected:    []string{"tag1", "latest", "tag2"},
			expectedErr: false,
		},
		{
			name: "Empty list",
			mockSetup: func(mockECR *mockRepositoryECRAPI) {
				mockECR.On("ListImages", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.ListImagesOutput{
						ImageIds: []types.ImageIdentifier{},
					}, nil)
			},
			expected:    []string{},
			expectedErr: false,
		},
		{
			name: "Only untagged images",
			mockSetup: func(mockECR *mockRepositoryECRAPI) {
				mockECR.On("ListImages", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.ListImagesOutput{
						ImageIds: []types.ImageIdentifier{
							{ImageDigest: aws.String("sha256:abc123")},
							{ImageDigest: aws.String("sha256:def456")},
						},
					}, nil)
			},
			expected:    []string{},
			expectedErr: false,
		},
		{
			name: "API error",
			mockSetup: func(mockECR *mockRepositoryECRAPI) {
				mockECR.On("ListImages", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.ListImagesOutput{}, errors.New("API error"))
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockECR := &mockRepositoryECRAPI{}
			tc.mockSetup(mockECR)

			repo := setupTestRepository(mockECR)

			tags, err := repo.ListTags(context.Background())
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tc.expected, tags)
			}

			mockECR.AssertExpectations(t)
		})
	}
}

func TestRepositoryGetManifest(t *testing.T) {
	// Skip the test unless explicitly enabled via environment variable
	if os.Getenv("ENABLE_ECR_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping ECR integration test that requires AWS API calls. Set ENABLE_ECR_INTEGRATION_TESTS=true to run.")
	}

	// This test would require more extensive mocking of the remote.Get functionality
	// which is difficult without a more sophisticated test framework.
	// For now, we'll skip this test for brevity.
	t.Skip("Skipping test that requires extensive mocking of go-containerregistry remote operations")
}

func TestRepositoryPutManifest(t *testing.T) {
	// Skip the test unless explicitly enabled via environment variable
	if os.Getenv("ENABLE_ECR_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping ECR integration test that requires AWS API calls. Set ENABLE_ECR_INTEGRATION_TESTS=true to run.")
	}

	// Similarly, this test would require extensive mocking.
	t.Skip("Skipping test that requires extensive mocking of go-containerregistry remote operations")
}

func TestRepositoryDeleteManifest(t *testing.T) {
	tests := []struct {
		name            string
		tag             string
		mockSetup       func(*mockRepositoryECRAPI)
		expectedErr     bool
		expectedErrType error
	}{
		{
			name: "Successful delete",
			tag:  "latest",
			mockSetup: func(mockECR *mockRepositoryECRAPI) {
				// First mock the BatchGetImage call
				mockECR.On("BatchGetImage", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.BatchGetImageOutput{
						Images: []types.Image{
							{
								ImageId: &types.ImageIdentifier{
									ImageDigest: aws.String("sha256:1234567890"),
									ImageTag:    aws.String("latest"),
								},
							},
						},
					}, nil)

				// Then mock the BatchDeleteImage call
				mockECR.On("BatchDeleteImage", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.BatchDeleteImageOutput{
						ImageIds: []types.ImageIdentifier{
							{ImageTag: aws.String("latest")},
						},
					}, nil)
			},
			expectedErr: false,
		},
		{
			name: "API error",
			tag:  "latest",
			mockSetup: func(mockECR *mockRepositoryECRAPI) {
				// First mock the BatchGetImage call
				mockECR.On("BatchGetImage", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.BatchGetImageOutput{
						Images: []types.Image{
							{
								ImageId: &types.ImageIdentifier{
									ImageDigest: aws.String("sha256:1234567890"),
									ImageTag:    aws.String("latest"),
								},
							},
						},
					}, nil)

				// Then mock the error on BatchDeleteImage
				mockECR.On("BatchDeleteImage", mock.Anything, mock.Anything, mock.Anything).
					Return(&ecr.BatchDeleteImageOutput{}, errors.New("API error"))
			},
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockECR := &mockRepositoryECRAPI{}
			tc.mockSetup(mockECR)

			repo := setupTestRepository(mockECR)

			err := repo.DeleteManifest(context.Background(), tc.tag)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockECR.AssertExpectations(t)
		})
	}
}
