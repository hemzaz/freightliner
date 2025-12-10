package common

import (
	"context"
	"net/http"
	"testing"

	"freightliner/pkg/helper/log"

	"github.com/stretchr/testify/assert"
)

func TestNewRegistryUtil(t *testing.T) {
	tests := []struct {
		name   string
		logger log.Logger
	}{
		{
			name:   "With logger",
			logger: log.NewBasicLogger(log.InfoLevel),
		},
		{
			name:   "Without logger (should create default)",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			util := NewRegistryUtil(tt.logger)
			assert.NotNil(t, util)
			assert.NotNil(t, util.logger)
		})
	}
}

func TestRegistryUtil_ParseRegistryPath(t *testing.T) {
	util := NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	tests := []struct {
		name               string
		path               string
		expectedRegistry   string
		expectedRepository string
		shouldError        bool
	}{
		{
			name:               "Valid path",
			path:               "ecr/my-repo",
			expectedRegistry:   "ecr",
			expectedRepository: "my-repo",
			shouldError:        false,
		},
		{
			name:               "Path with namespace",
			path:               "gcr/org/repo",
			expectedRegistry:   "gcr",
			expectedRepository: "org/repo",
			shouldError:        false,
		},
		{
			name:               "Invalid path - no separator",
			path:               "just-a-repo",
			expectedRegistry:   "",
			expectedRepository: "",
			shouldError:        true,
		},
		{
			name:               "Empty path",
			path:               "",
			expectedRegistry:   "",
			expectedRepository: "",
			shouldError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry, repository, err := util.ParseRegistryPath(tt.path)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRegistry, registry)
				assert.Equal(t, tt.expectedRepository, repository)
			}
		})
	}
}

func TestRegistryUtil_ValidateRepositoryName(t *testing.T) {
	util := NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	tests := []struct {
		name        string
		repoName    string
		shouldError bool
	}{
		{
			name:        "Valid repository name",
			repoName:    "my-repo",
			shouldError: false,
		},
		{
			name:        "Valid with namespace",
			repoName:    "org/my-repo",
			shouldError: false,
		},
		{
			name:        "Empty repository name",
			repoName:    "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := util.ValidateRepositoryName(tt.repoName)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegistryUtil_CreateRepositoryReference(t *testing.T) {
	util := NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	tests := []struct {
		name        string
		registry    string
		repoName    string
		shouldError bool
	}{
		{
			name:        "Valid repository",
			registry:    "example.com",
			repoName:    "my-repo",
			shouldError: false,
		},
		{
			name:        "Empty repository name",
			registry:    "example.com",
			repoName:    "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := util.CreateRepositoryReference(tt.registry, tt.repoName)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, ref.String(), tt.registry)
				assert.Contains(t, ref.String(), tt.repoName)
			}
		})
	}
}

func TestRegistryUtil_GetRemoteOptions(t *testing.T) {
	util := NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	// Test with nil transport
	opts := util.GetRemoteOptions(nil)
	assert.Len(t, opts, 0)

	// Test with mock transport
	mockTransport := &http.Transport{}
	opts = util.GetRemoteOptions(mockTransport)
	assert.Len(t, opts, 1)
}

func TestRegistryUtil_IsValidRegistryType(t *testing.T) {
	util := NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	tests := []struct {
		name         string
		registryType string
		expected     bool
	}{
		{
			name:         "ECR is valid",
			registryType: "ecr",
			expected:     true,
		},
		{
			name:         "GCR is valid",
			registryType: "gcr",
			expected:     true,
		},
		{
			name:         "Docker Hub is not valid",
			registryType: "dockerhub",
			expected:     false,
		},
		{
			name:         "Empty string is not valid",
			registryType: "",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.IsValidRegistryType(tt.registryType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRegistryUtil_FormatRepositoryURI(t *testing.T) {
	util := NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	tests := []struct {
		name         string
		registryType string
		accountID    string
		region       string
		repoName     string
		expected     string
	}{
		{
			name:         "ECR URI",
			registryType: "ecr",
			accountID:    "123456789012",
			region:       "us-west-2",
			repoName:     "my-repo",
			expected:     "123456789012.dkr.ecr.us-west-2.amazonaws.com/my-repo",
		},
		{
			name:         "GCR URI",
			registryType: "gcr",
			accountID:    "my-project",
			region:       "",
			repoName:     "my-repo",
			expected:     "gcr.io/my-project/my-repo",
		},
		{
			name:         "Unknown registry type",
			registryType: "custom",
			accountID:    "account",
			region:       "region",
			repoName:     "my-repo",
			expected:     "custom/my-repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.FormatRepositoryURI(tt.registryType, tt.accountID, tt.region, tt.repoName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRegistryUtil_LogRegistryOperation(t *testing.T) {
	util := NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	ctx := context.Background()

	// Should not panic
	assert.NotPanics(t, func() {
		util.LogRegistryOperation(ctx, "pull", "example.com", "my-repo", map[string]interface{}{
			"tag": "latest",
		})
	})

	// Test with nil extra fields
	assert.NotPanics(t, func() {
		util.LogRegistryOperation(ctx, "push", "example.com", "my-repo", nil)
	})

	// Test with empty extra fields
	assert.NotPanics(t, func() {
		util.LogRegistryOperation(ctx, "delete", "example.com", "my-repo", map[string]interface{}{})
	})
}

func TestFormRegistryPath(t *testing.T) {
	tests := []struct {
		name     string
		registry string
		repoName string
		expected string
	}{
		{
			name:     "ECR path",
			registry: "123456789012.dkr.ecr.us-west-2.amazonaws.com",
			repoName: "my-repo",
			expected: "123456789012.dkr.ecr.us-west-2.amazonaws.com/my-repo",
		},
		{
			name:     "GCR path",
			registry: "gcr.io",
			repoName: "project/my-repo",
			expected: "gcr.io/project/my-repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormRegistryPath(tt.registry, tt.repoName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseRegistryPath_Standalone(t *testing.T) {
	tests := []struct {
		name               string
		path               string
		expectedRegistry   string
		expectedRepository string
		shouldError        bool
	}{
		{
			name:               "Valid ECR path",
			path:               "ecr/my-repo",
			expectedRegistry:   "ecr",
			expectedRepository: "my-repo",
			shouldError:        false,
		},
		{
			name:               "Valid GCR path with namespace",
			path:               "gcr/project/repo",
			expectedRegistry:   "gcr",
			expectedRepository: "project/repo",
			shouldError:        false,
		},
		{
			name:               "Invalid path",
			path:               "invalid",
			expectedRegistry:   "",
			expectedRepository: "",
			shouldError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry, repository, err := ParseRegistryPath(tt.path)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRegistry, registry)
				assert.Equal(t, tt.expectedRepository, repository)
			}
		})
	}
}

func TestGetRemoteOptions_WithTransport(t *testing.T) {
	util := NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	transport := &http.Transport{
		MaxIdleConns: 10,
	}

	opts := util.GetRemoteOptions(transport)
	assert.NotNil(t, opts)
	assert.Len(t, opts, 1)
}
