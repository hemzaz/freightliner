package mocks

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	ststypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/stretchr/testify/mock"
)

// MockECRClient implements a mock ECR client for testing
type MockECRClient struct {
	mock.Mock
}

func (m *MockECRClient) DescribeRepositories(ctx context.Context, params *ecr.DescribeRepositoriesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ecr.DescribeRepositoriesOutput), args.Error(1)
}

func (m *MockECRClient) CreateRepository(ctx context.Context, params *ecr.CreateRepositoryInput, optFns ...func(*ecr.Options)) (*ecr.CreateRepositoryOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ecr.CreateRepositoryOutput), args.Error(1)
}

func (m *MockECRClient) GetAuthorizationToken(ctx context.Context, params *ecr.GetAuthorizationTokenInput, optFns ...func(*ecr.Options)) (*ecr.GetAuthorizationTokenOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ecr.GetAuthorizationTokenOutput), args.Error(1)
}

func (m *MockECRClient) DescribeImages(ctx context.Context, params *ecr.DescribeImagesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeImagesOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ecr.DescribeImagesOutput), args.Error(1)
}

func (m *MockECRClient) PutImage(ctx context.Context, params *ecr.PutImageInput, optFns ...func(*ecr.Options)) (*ecr.PutImageOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ecr.PutImageOutput), args.Error(1)
}

func (m *MockECRClient) BatchGetImage(ctx context.Context, params *ecr.BatchGetImageInput, optFns ...func(*ecr.Options)) (*ecr.BatchGetImageOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*ecr.BatchGetImageOutput), args.Error(1)
}

// MockSTSClient implements a mock STS client for testing
type MockSTSClient struct {
	mock.Mock
}

func (m *MockSTSClient) GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*sts.GetCallerIdentityOutput), args.Error(1)
}

func (m *MockSTSClient) AssumeRole(ctx context.Context, params *sts.AssumeRoleInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*sts.AssumeRoleOutput), args.Error(1)
}

// Helper functions to create mock responses

// CreateMockECRRepositories creates sample ECR repositories for testing
func CreateMockECRRepositories(count int) []types.Repository {
	repos := make([]types.Repository, count)
	now := time.Now()

	for i := 0; i < count; i++ {
		repos[i] = types.Repository{
			RegistryId:     aws.String("123456789012"),
			RepositoryName: aws.String(fmt.Sprintf("test-repo-%d", i+1)),
			RepositoryUri:  aws.String(fmt.Sprintf("123456789012.dkr.ecr.us-east-1.amazonaws.com/test-repo-%d", i+1)),
			CreatedAt:      &now,
			ImageScanningConfiguration: &types.ImageScanningConfiguration{
				ScanOnPush: false,
			},
			ImageTagMutability: types.ImageTagMutabilityMutable,
			EncryptionConfiguration: &types.EncryptionConfiguration{
				EncryptionType: types.EncryptionTypeAes256,
			},
		}
	}

	return repos
}

// CreateMockAuthToken creates a mock ECR authorization token
func CreateMockAuthToken() *ecr.GetAuthorizationTokenOutput {
	token := "dGVzdC11c2VyOnRlc3QtcGFzc3dvcmQ=" // base64 encoded "test-user:test-password"
	endpoint := "https://123456789012.dkr.ecr.us-east-1.amazonaws.com"
	expiresAt := time.Now().Add(12 * time.Hour)

	return &ecr.GetAuthorizationTokenOutput{
		AuthorizationData: []types.AuthorizationData{
			{
				AuthorizationToken: &token,
				ExpiresAt:          &expiresAt,
				ProxyEndpoint:      &endpoint,
			},
		},
	}
}

// CreateMockCallerIdentity creates a mock STS caller identity
func CreateMockCallerIdentity(accountID string) *sts.GetCallerIdentityOutput {
	arn := fmt.Sprintf("arn:aws:iam::%s:user/test-user", accountID)
	userID := "AIDACKCEVSQ6C2EXAMPLE"

	return &sts.GetCallerIdentityOutput{
		Account: &accountID,
		Arn:     &arn,
		UserId:  &userID,
	}
}

// CreateMockSTSError creates a mock STS error for testing error scenarios
func CreateMockSTSError() error {
	return &ststypes.InvalidAuthorizationMessageException{
		Message: aws.String("Mock STS error for testing"),
	}
}

// CreateMockImages creates sample ECR images for testing
func CreateMockImages(repoName string, count int) []types.ImageDetail {
	images := make([]types.ImageDetail, count)
	now := time.Now()

	for i := 0; i < count; i++ {
		tag := fmt.Sprintf("v1.0.%d", i+1)
		digest := fmt.Sprintf("sha256:abcdef%024d", i)
		sizeBytes := int64(1024 * 1024 * (i + 1)) // Varying sizes

		images[i] = types.ImageDetail{
			ImageDigest:      &digest,
			ImageTags:        []string{tag},
			ImageSizeInBytes: &sizeBytes,
			ImagePushedAt:    &now,
			RegistryId:       aws.String("123456789012"),
			RepositoryName:   &repoName,
		}
	}

	return images
}

// MockECRClientBuilder provides a fluent interface for setting up ECR mock expectations
type MockECRClientBuilder struct {
	client *MockECRClient
}

// NewMockECRClient creates a new mock ECR client builder
func NewMockECRClient() *MockECRClientBuilder {
	return &MockECRClientBuilder{
		client: &MockECRClient{},
	}
}

// ExpectDescribeRepositories sets up expectations for DescribeRepositories calls
func (b *MockECRClientBuilder) ExpectDescribeRepositories(repos []types.Repository, err error) *MockECRClientBuilder {
	output := &ecr.DescribeRepositoriesOutput{
		Repositories: repos,
	}
	b.client.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).Return(output, err)
	return b
}

// ExpectGetAuthorizationToken sets up expectations for GetAuthorizationToken calls
func (b *MockECRClientBuilder) ExpectGetAuthorizationToken(token *ecr.GetAuthorizationTokenOutput, err error) *MockECRClientBuilder {
	b.client.On("GetAuthorizationToken", mock.Anything, mock.Anything, mock.Anything).Return(token, err)
	return b
}

// ExpectDescribeImages sets up expectations for DescribeImages calls
func (b *MockECRClientBuilder) ExpectDescribeImages(images []types.ImageDetail, err error) *MockECRClientBuilder {
	output := &ecr.DescribeImagesOutput{
		ImageDetails: images,
	}
	b.client.On("DescribeImages", mock.Anything, mock.Anything, mock.Anything).Return(output, err)
	return b
}

// Build returns the configured mock client
func (b *MockECRClientBuilder) Build() *MockECRClient {
	return b.client
}

// MockSTSClientBuilder provides a fluent interface for setting up STS mock expectations
type MockSTSClientBuilder struct {
	client *MockSTSClient
}

// NewMockSTSClient creates a new mock STS client builder
func NewMockSTSClient() *MockSTSClientBuilder {
	return &MockSTSClientBuilder{
		client: &MockSTSClient{},
	}
}

// ExpectGetCallerIdentity sets up expectations for GetCallerIdentity calls
func (b *MockSTSClientBuilder) ExpectGetCallerIdentity(identity *sts.GetCallerIdentityOutput, err error) *MockSTSClientBuilder {
	b.client.On("GetCallerIdentity", mock.Anything, mock.Anything, mock.Anything).Return(identity, err)
	return b
}

// Build returns the configured mock client
func (b *MockSTSClientBuilder) Build() *MockSTSClient {
	return b.client
}
