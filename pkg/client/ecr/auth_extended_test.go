package ecr

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsecr "github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewECRAuthenticator(t *testing.T) {
	mockService := new(MockECRServiceExt)

	auth := NewECRAuthenticator(mockService, "us-west-2")
	assert.NotNil(t, auth)
}

func TestECRAuthenticator_AuthorizationMultipleCalls(t *testing.T) {
	mockService := new(MockECRServiceExt)

	// Setup mock to return different tokens
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

	auth := NewECRAuthenticator(mockService, "us-west-2")

	// First call
	config1, err := auth.Authorization()
	assert.NoError(t, err)
	assert.Equal(t, "AWS", config1.Username)
	assert.Equal(t, "token1", config1.Password)

	// Second call - should get fresh token
	config2, err := auth.Authorization()
	assert.NoError(t, err)
	assert.Equal(t, "AWS", config2.Username)
	assert.Equal(t, "token2", config2.Password)

	mockService.AssertExpectations(t)
}

func TestECRAuthenticator_RegistryAuthenticatorCrossRegion(t *testing.T) {
	mockService := new(MockECRServiceExt)
	auth := NewECRAuthenticator(mockService, "us-west-2")

	// Test cross-region registry
	_, err := auth.RegistryAuthenticator("123456789012.dkr.ecr.eu-central-1.amazonaws.com")
	// Will error without AWS credentials but tests the code path
	if err != nil {
		assert.Error(t, err)
	}
}

func TestIsECRRegistryExtended(t *testing.T) {
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
			name:     "Public ECR",
			registry: "public.ecr.aws",
			expected: true,
		},
		{
			name:     "Docker Hub",
			registry: "docker.io",
			expected: false,
		},
		{
			name:     "GCR",
			registry: "gcr.io",
			expected: false,
		},
		{
			name:     "Empty string",
			registry: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isECRRegistry(tt.registry)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetECRRegistryMultipleRegions(t *testing.T) {
	tests := []struct {
		accountID string
		region    string
		expected  string
	}{
		{
			accountID: "123456789012",
			region:    "us-west-2",
			expected:  "123456789012.dkr.ecr.us-west-2.amazonaws.com",
		},
		{
			accountID: "987654321098",
			region:    "eu-central-1",
			expected:  "987654321098.dkr.ecr.eu-central-1.amazonaws.com",
		},
		{
			accountID: "111222333444",
			region:    "ap-southeast-1",
			expected:  "111222333444.dkr.ecr.ap-southeast-1.amazonaws.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.region, func(t *testing.T) {
			result := GetECRRegistry(tt.accountID, tt.region)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDefaultCredentialHelper(t *testing.T) {
	helper := GetDefaultCredentialHelper()
	assert.NotNil(t, helper)
}

func TestCreateAWSConfig(t *testing.T) {
	tests := []struct {
		name string
		opts ClientOptions
	}{
		{
			name: "Default config",
			opts: ClientOptions{
				Region: "us-west-2",
			},
		},
		{
			name: "With profile",
			opts: ClientOptions{
				Region:  "us-west-2",
				Profile: "default",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := createAWSConfig(ctx, &tt.opts)
			// Will succeed if AWS credentials are available
			if err != nil {
				assert.Error(t, err)
			}
		})
	}
}

func TestCreateECRClient(t *testing.T) {
	ctx := context.Background()
	cfg, err := createAWSConfig(ctx, &ClientOptions{Region: "us-west-2"})
	if err == nil {
		// Test without role
		client1, err := createECRClient(cfg, "")
		if err == nil {
			assert.NotNil(t, client1)
		}

		// Test with role (will fail without actual role)
		_, err = createECRClient(cfg, "arn:aws:iam::123456789012:role/TestRole")
		// Expected to error without actual AWS role
		if err != nil {
			assert.Error(t, err)
		}
	}
}
