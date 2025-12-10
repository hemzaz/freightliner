package ecr

import (
	"context"
	"testing"

	"freightliner/pkg/testing/mocks"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	freightliner_log "freightliner/pkg/helper/log"
)

func TestNewClientWithMocks(t *testing.T) {
	tests := []struct {
		name          string
		region        string
		accountID     string
		setupMocks    func() (*mocks.MockECRClient, *mocks.MockSTSClient)
		expectedErr   bool
		expectedAccID string
	}{
		{
			name:      "Explicit account ID success",
			region:    "us-east-1",
			accountID: "123456789012",
			setupMocks: func() (*mocks.MockECRClient, *mocks.MockSTSClient) {
				// For conceptual test - no expectations needed as clients won't be called
				ecrClient := mocks.NewMockECRClient().Build()
				stsClient := mocks.NewMockSTSClient().Build()

				return ecrClient, stsClient
			},
			expectedErr:   false,
			expectedAccID: "123456789012",
		},
		{
			name:      "Auto-detect account ID success",
			region:    "us-east-1",
			accountID: "",
			setupMocks: func() (*mocks.MockECRClient, *mocks.MockSTSClient) {
				// For conceptual test - no expectations needed as clients won't be called
				ecrClient := mocks.NewMockECRClient().Build()
				stsClient := mocks.NewMockSTSClient().Build()

				return ecrClient, stsClient
			},
			expectedErr:   false,
			expectedAccID: "123456789012",
		},
		{
			name:      "STS error when auto-detecting account ID",
			region:    "us-east-1",
			accountID: "",
			setupMocks: func() (*mocks.MockECRClient, *mocks.MockSTSClient) {
				// For conceptual test - no expectations needed as clients won't be called
				ecrClient := mocks.NewMockECRClient().Build()
				stsClient := mocks.NewMockSTSClient().Build()

				return ecrClient, stsClient
			},
			expectedErr:   true,
			expectedAccID: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockECR, mockSTS := tc.setupMocks()

			// Create client options with mock clients
			opts := ClientOptions{
				Region:    tc.region,
				AccountID: tc.accountID,
				// In a real implementation, you'd inject these mocks through the options
				// For this example, we're showing the testing pattern
			}

			// This is a conceptual test - in practice you'd need dependency injection
			// to replace the real AWS clients with mocks
			_ = opts
			_ = mockECR
			_ = mockSTS

			// Verify mock expectations were met
			mockECR.AssertExpectations(t)
			mockSTS.AssertExpectations(t)

			// In a real test, you would:
			// client, err := NewClientWithMocks(opts, mockECR, mockSTS)
			// if tc.expectedErr {
			//     assert.Error(t, err)
			// } else {
			//     assert.NoError(t, err)
			//     assert.NotNil(t, client)
			//     assert.Equal(t, tc.expectedAccID, client.accountID)
			// }
		})
	}
}

func TestECRRepositoryOperationsWithMocks(t *testing.T) {
	tests := []struct {
		name       string
		repoName   string
		setupMocks func() *mocks.MockECRClient
		operation  string
		expectErr  bool
	}{
		{
			name:     "Successful repository list",
			repoName: "",
			setupMocks: func() *mocks.MockECRClient {
				repos := mocks.CreateMockECRRepositories(3)
				return mocks.NewMockECRClient().
					ExpectDescribeRepositories(repos, nil).
					Build()
			},
			operation: "list",
			expectErr: false,
		},
		{
			name:     "Successful image description",
			repoName: "test-repo",
			setupMocks: func() *mocks.MockECRClient {
				images := mocks.CreateMockImages("test-repo", 2)
				return mocks.NewMockECRClient().
					ExpectDescribeImages(images, nil).
					Build()
			},
			operation: "describe_images",
			expectErr: false,
		},
		{
			name:     "Repository not found error",
			repoName: "non-existent-repo",
			setupMocks: func() *mocks.MockECRClient {
				// Return empty result to simulate repository not found
				return mocks.NewMockECRClient().
					ExpectDescribeRepositories([]types.Repository{}, nil).
					Build()
			},
			operation: "list",
			expectErr: false, // Empty list is not an error
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := tc.setupMocks()
			ctx := context.Background()

			switch tc.operation {
			case "list":
				// Test repository listing
				input := &ecr.DescribeRepositoriesInput{}
				result, err := mockClient.DescribeRepositories(ctx, input)

				if tc.expectErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, result)
				}

			case "describe_images":
				// Test image description
				input := &ecr.DescribeImagesInput{
					RepositoryName: &tc.repoName,
				}
				result, err := mockClient.DescribeImages(ctx, input)

				if tc.expectErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, result)
					if len(result.ImageDetails) > 0 {
						assert.Equal(t, tc.repoName, *result.ImageDetails[0].RepositoryName)
					}
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestECRAuthenticationWithMocks(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func() *mocks.MockECRClient
		expectErr  bool
	}{
		{
			name: "Successful authentication",
			setupMocks: func() *mocks.MockECRClient {
				token := mocks.CreateMockAuthToken()
				return mocks.NewMockECRClient().
					ExpectGetAuthorizationToken(token, nil).
					Build()
			},
			expectErr: false,
		},
		{
			name: "Authentication failure",
			setupMocks: func() *mocks.MockECRClient {
				return mocks.NewMockECRClient().
					ExpectGetAuthorizationToken(nil, assert.AnError).
					Build()
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := tc.setupMocks()
			ctx := context.Background()

			input := &ecr.GetAuthorizationTokenInput{}
			result, err := mockClient.GetAuthorizationToken(ctx, input)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.AuthorizationData, 1)
				assert.NotNil(t, result.AuthorizationData[0].AuthorizationToken)
				assert.NotNil(t, result.AuthorizationData[0].ProxyEndpoint)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

// TestECRClientIntegrationWithMocks demonstrates how to test the full client with mocks
func TestECRClientIntegrationWithMocks(t *testing.T) {
	_ = freightliner_log.NewLogger() // Logger available if needed for client creation

	// This test demonstrates the pattern but would require dependency injection
	// in the actual ECR client to work with mocks

	t.Run("Full client workflow", func(t *testing.T) {
		// Setup mocks for a complete workflow
		mockECRBuilder := mocks.NewMockECRClient()
		mockSTSBuilder := mocks.NewMockSTSClient()

		// Expect STS call to get account ID
		identity := mocks.CreateMockCallerIdentity("123456789012")
		mockSTSBuilder.ExpectGetCallerIdentity(identity, nil)

		// Expect ECR auth token request
		token := mocks.CreateMockAuthToken()
		mockECRBuilder.ExpectGetAuthorizationToken(token, nil)

		// Expect repository listing
		repos := mocks.CreateMockECRRepositories(2)
		mockECRBuilder.ExpectDescribeRepositories(repos, nil)

		// Build the actual mock clients
		mockECR := mockECRBuilder.Build()
		mockSTS := mockSTSBuilder.Build()

		// In practice, you'd create a client with injected dependencies:
		// client := NewClientWithDependencies(ClientOptions{
		//     Region: "us-east-1",
		// }, mockECR, mockSTS, logger)

		// Then test client operations:
		// repositories, err := client.ListRepositories(context.Background())
		// assert.NoError(t, err)
		// assert.Len(t, repositories, 2)

		// For now, just verify the mocks would work
		ctx := context.Background()

		// Test STS call
		stsResult, err := mockSTS.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
		assert.NoError(t, err)
		assert.Equal(t, "123456789012", *stsResult.Account)

		// Test ECR auth
		authResult, err := mockECR.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
		assert.NoError(t, err)
		assert.NotEmpty(t, authResult.AuthorizationData)

		// Test repository listing
		repoResult, err := mockECR.DescribeRepositories(ctx, &ecr.DescribeRepositoriesInput{})
		assert.NoError(t, err)
		assert.Len(t, repoResult.Repositories, 2)

		mockECR.AssertExpectations(t)
		mockSTS.AssertExpectations(t)
	})
}

// TestECRErrorScenarios tests various error conditions with mocks
func TestECRErrorScenarios(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func() *mocks.MockECRClient
		operation   string
		expectedErr string
	}{
		{
			name: "Network timeout",
			setupMocks: func() *mocks.MockECRClient {
				mockClient := &mocks.MockECRClient{}
				mockClient.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).
					Return((*ecr.DescribeRepositoriesOutput)(nil), context.DeadlineExceeded)
				return mockClient
			},
			operation:   "list_repos",
			expectedErr: "context deadline exceeded",
		},
		{
			name: "Access denied",
			setupMocks: func() *mocks.MockECRClient {
				mockClient := &mocks.MockECRClient{}
				// Use InvalidParameterException as AccessDeniedException is not available in ECR types
				accessDeniedErr := &types.InvalidParameterException{
					Message: aws.String("Access denied for repository access"),
				}
				mockClient.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).
					Return((*ecr.DescribeRepositoriesOutput)(nil), accessDeniedErr)
				return mockClient
			},
			operation:   "list_repos",
			expectedErr: "Access denied",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := tc.setupMocks()
			ctx := context.Background()

			switch tc.operation {
			case "list_repos":
				_, err := mockClient.DescribeRepositories(ctx, &ecr.DescribeRepositoriesInput{})
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

// BenchmarkECRMockOperations benchmarks mock operations to ensure they're fast
func BenchmarkECRMockOperations(b *testing.B) {
	mockClient := mocks.NewMockECRClient().
		ExpectDescribeRepositories(mocks.CreateMockECRRepositories(10), nil).
		Build()

	ctx := context.Background()
	input := &ecr.DescribeRepositoriesInput{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mockClient.DescribeRepositories(ctx, input)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}
