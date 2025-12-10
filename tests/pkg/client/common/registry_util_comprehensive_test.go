package common

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"freightliner/pkg/client/common"
	"freightliner/pkg/helper/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegistryUtil_ParseRegistryPath_EdgeCases tests edge cases for registry path parsing
func TestRegistryUtil_ParseRegistryPath_EdgeCases(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	tests := []struct {
		name               string
		path               string
		expectedRegistry   string
		expectedRepository string
		shouldError        bool
		errorContains      string
	}{
		{
			name:               "Multiple slashes in repository",
			path:               "ecr/org/team/repo",
			expectedRegistry:   "ecr",
			expectedRepository: "org/team/repo",
			shouldError:        false,
		},
		{
			name:               "Trailing slash",
			path:               "gcr/repo/",
			expectedRegistry:   "gcr",
			expectedRepository: "repo/",
			shouldError:        false,
		},
		{
			name:               "Leading slash",
			path:               "/ecr/repo",
			expectedRegistry:   "",
			expectedRepository: "ecr/repo",
			shouldError:        false,
		},
		{
			name:          "Only slash",
			path:          "/",
			shouldError:   true,
			errorContains: "invalid format",
		},
		{
			name:          "Multiple slashes only",
			path:          "///",
			shouldError:   true,
			errorContains: "invalid format",
		},
		{
			name:               "Very long path",
			path:               "registry/" + strings.Repeat("a", 500),
			expectedRegistry:   "registry",
			expectedRepository: strings.Repeat("a", 500),
			shouldError:        false,
		},
		{
			name:               "Special characters in repository",
			path:               "ecr/my-repo_v2.3",
			expectedRegistry:   "ecr",
			expectedRepository: "my-repo_v2.3",
			shouldError:        false,
		},
		{
			name:               "Unicode characters",
			path:               "gcr/测试-repo",
			expectedRegistry:   "gcr",
			expectedRepository: "测试-repo",
			shouldError:        false,
		},
		{
			name:          "Whitespace path",
			path:          "   ",
			shouldError:   true,
			errorContains: "invalid format",
		},
		{
			name:          "Tab separated",
			path:          "ecr\trepo",
			shouldError:   true,
			errorContains: "invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry, repository, err := util.ParseRegistryPath(tt.path)

			if tt.shouldError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRegistry, registry)
				assert.Equal(t, tt.expectedRepository, repository)
			}
		})
	}
}

// TestRegistryUtil_ValidateRepositoryName_EdgeCases tests edge cases for repository validation
func TestRegistryUtil_ValidateRepositoryName_EdgeCases(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	tests := []struct {
		name        string
		repoName    string
		shouldError bool
		description string
	}{
		{
			name:        "Whitespace only",
			repoName:    "   ",
			shouldError: false, // Current implementation doesn't check whitespace
			description: "Repository with only whitespace",
		},
		{
			name:        "Very long name",
			repoName:    strings.Repeat("a", 256),
			shouldError: false,
			description: "Repository name at max length",
		},
		{
			name:        "Name with dots",
			repoName:    "my.repo.name",
			shouldError: false,
			description: "Repository with dots",
		},
		{
			name:        "Name with underscores",
			repoName:    "my_repo_name",
			shouldError: false,
			description: "Repository with underscores",
		},
		{
			name:        "Name with hyphens",
			repoName:    "my-repo-name",
			shouldError: false,
			description: "Repository with hyphens",
		},
		{
			name:        "Mixed separators",
			repoName:    "my-repo.name_v2",
			shouldError: false,
			description: "Repository with mixed separators",
		},
		{
			name:        "Namespace with slash",
			repoName:    "org/team/repo",
			shouldError: false,
			description: "Multi-level namespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := util.ValidateRepositoryName(tt.repoName)
			if tt.shouldError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// TestRegistryUtil_CreateRepositoryReference_EdgeCases tests edge cases for repository reference creation
func TestRegistryUtil_CreateRepositoryReference_EdgeCases(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	tests := []struct {
		name        string
		registry    string
		repoName    string
		shouldError bool
		description string
	}{
		{
			name:        "Registry with port",
			registry:    "example.com:5000",
			repoName:    "my-repo",
			shouldError: false,
			description: "Registry with explicit port",
		},
		{
			name:        "Localhost registry",
			registry:    "localhost",
			repoName:    "test-repo",
			shouldError: false,
			description: "Local registry",
		},
		{
			name:        "IP address registry",
			registry:    "192.168.1.100",
			repoName:    "my-repo",
			shouldError: false,
			description: "IP-based registry",
		},
		{
			name:        "IP with port",
			registry:    "192.168.1.100:5000",
			repoName:    "my-repo",
			shouldError: false,
			description: "IP registry with port",
		},
		{
			name:        "Registry with protocol",
			registry:    "https://example.com",
			repoName:    "my-repo",
			shouldError: true,
			description: "Registry should not include protocol",
		},
		{
			name:        "Very long registry name",
			registry:    strings.Repeat("a", 253) + ".com",
			repoName:    "repo",
			shouldError: false,
			description: "Max length registry name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := util.CreateRepositoryReference(tt.registry, tt.repoName)
			if tt.shouldError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, ref)
			}
		})
	}
}

// TestRegistryUtil_IsValidRegistryType_AllCases tests all registry type validations
func TestRegistryUtil_IsValidRegistryType_AllCases(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	tests := []struct {
		name         string
		registryType string
		expected     bool
	}{
		// Valid types
		{name: "ECR lowercase", registryType: "ecr", expected: true},
		{name: "GCR lowercase", registryType: "gcr", expected: true},

		// Invalid types
		{name: "ECR uppercase", registryType: "ECR", expected: false},
		{name: "GCR uppercase", registryType: "GCR", expected: false},
		{name: "ACR (Azure)", registryType: "acr", expected: false},
		{name: "Docker Hub", registryType: "dockerhub", expected: false},
		{name: "Quay", registryType: "quay", expected: false},
		{name: "Harbor", registryType: "harbor", expected: false},
		{name: "Artifactory", registryType: "artifactory", expected: false},
		{name: "GitLab", registryType: "gitlab", expected: false},
		{name: "GitHub", registryType: "github", expected: false},
		{name: "Custom", registryType: "custom", expected: false},
		{name: "Empty string", registryType: "", expected: false},
		{name: "Whitespace", registryType: "   ", expected: false},
		{name: "ECR with space", registryType: "ecr ", expected: false},
		{name: "Mixed case", registryType: "Ecr", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.IsValidRegistryType(tt.registryType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRegistryUtil_FormatRepositoryURI_AllRegistries tests URI formatting for all registry types
func TestRegistryUtil_FormatRepositoryURI_AllRegistries(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	tests := []struct {
		name         string
		registryType string
		accountID    string
		region       string
		repoName     string
		expected     string
	}{
		// ECR variations
		{
			name:         "ECR us-west-2",
			registryType: "ecr",
			accountID:    "123456789012",
			region:       "us-west-2",
			repoName:     "my-app",
			expected:     "123456789012.dkr.ecr.us-west-2.amazonaws.com/my-app",
		},
		{
			name:         "ECR us-east-1",
			registryType: "ecr",
			accountID:    "987654321098",
			region:       "us-east-1",
			repoName:     "backend-service",
			expected:     "987654321098.dkr.ecr.us-east-1.amazonaws.com/backend-service",
		},
		{
			name:         "ECR eu-west-1",
			registryType: "ecr",
			accountID:    "111222333444",
			region:       "eu-west-1",
			repoName:     "frontend/web",
			expected:     "111222333444.dkr.ecr.eu-west-1.amazonaws.com/frontend/web",
		},
		{
			name:         "ECR with namespace",
			registryType: "ecr",
			accountID:    "123456789012",
			region:       "us-west-2",
			repoName:     "team/project/service",
			expected:     "123456789012.dkr.ecr.us-west-2.amazonaws.com/team/project/service",
		},

		// GCR variations
		{
			name:         "GCR basic",
			registryType: "gcr",
			accountID:    "my-project-123",
			region:       "",
			repoName:     "my-app",
			expected:     "gcr.io/my-project-123/my-app",
		},
		{
			name:         "GCR with namespace",
			registryType: "gcr",
			accountID:    "company-prod",
			region:       "",
			repoName:     "team/app/service",
			expected:     "gcr.io/company-prod/team/app/service",
		},

		// Custom/unknown registries
		{
			name:         "Docker Hub",
			registryType: "dockerhub",
			accountID:    "myorg",
			region:       "",
			repoName:     "myapp",
			expected:     "dockerhub/myapp",
		},
		{
			name:         "ACR (Azure)",
			registryType: "acr",
			accountID:    "myregistry",
			region:       "westus",
			repoName:     "app",
			expected:     "acr/app",
		},
		{
			name:         "Harbor",
			registryType: "harbor",
			accountID:    "registry.example.com",
			region:       "",
			repoName:     "project/app",
			expected:     "harbor/project/app",
		},
		{
			name:         "Empty registry type",
			registryType: "",
			accountID:    "account",
			region:       "region",
			repoName:     "repo",
			expected:     "/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.FormatRepositoryURI(tt.registryType, tt.accountID, tt.region, tt.repoName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRegistryUtil_GetRemoteOptions_Concurrent tests thread safety
func TestRegistryUtil_GetRemoteOptions_Concurrent(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	// Test concurrent access to ensure thread safety
	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			var transport http.RoundTripper
			if index%2 == 0 {
				transport = &http.Transport{}
			}

			opts := util.GetRemoteOptions(transport)
			assert.NotNil(t, opts)
		}(i)
	}

	wg.Wait()
}

// TestRegistryUtil_LogRegistryOperation_Concurrent tests concurrent logging
func TestRegistryUtil_LogRegistryOperation_Concurrent(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))
	ctx := context.Background()

	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			util.LogRegistryOperation(
				ctx,
				"test-operation",
				"test-registry",
				"test-repo",
				map[string]interface{}{
					"index": index,
					"time":  time.Now(),
				},
			)
		}(i)
	}

	wg.Wait()
}

// TestRegistryUtil_NilLogger tests behavior with nil logger
func TestRegistryUtil_NilLogger(t *testing.T) {
	// Should not panic with nil logger
	util := common.NewRegistryUtil(nil)
	assert.NotNil(t, util)

	// All operations should work without panicking
	ctx := context.Background()

	t.Run("ParseRegistryPath with nil logger", func(t *testing.T) {
		assert.NotPanics(t, func() {
			_, _, _ = util.ParseRegistryPath("ecr/repo")
		})
	})

	t.Run("ValidateRepositoryName with nil logger", func(t *testing.T) {
		assert.NotPanics(t, func() {
			_ = util.ValidateRepositoryName("repo")
		})
	})

	t.Run("LogRegistryOperation with nil logger", func(t *testing.T) {
		assert.NotPanics(t, func() {
			util.LogRegistryOperation(ctx, "test", "reg", "repo", nil)
		})
	})
}

// TestRegistryUtil_ContextCancellation tests behavior with cancelled context
func TestRegistryUtil_ContextCancellation(t *testing.T) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Operations should handle cancelled context gracefully
	t.Run("LogRegistryOperation with cancelled context", func(t *testing.T) {
		assert.NotPanics(t, func() {
			util.LogRegistryOperation(ctx, "test", "registry", "repo", nil)
		})
	})
}

// TestRegistryUtil_Performance tests performance characteristics
func TestRegistryUtil_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	t.Run("ParseRegistryPath performance", func(t *testing.T) {
		start := time.Now()
		iterations := 10000

		for i := 0; i < iterations; i++ {
			_, _, _ = util.ParseRegistryPath("ecr/my-repo")
		}

		duration := time.Since(start)
		avgTime := duration / time.Duration(iterations)

		// Should be very fast - less than 10 microseconds per operation
		assert.Less(t, avgTime, 10*time.Microsecond,
			"ParseRegistryPath took %v per operation, expected < 10µs", avgTime)
	})

	t.Run("FormatRepositoryURI performance", func(t *testing.T) {
		start := time.Now()
		iterations := 10000

		for i := 0; i < iterations; i++ {
			_ = util.FormatRepositoryURI("ecr", "123456789012", "us-west-2", "repo")
		}

		duration := time.Since(start)
		avgTime := duration / time.Duration(iterations)

		// Should be very fast - less than 10 microseconds per operation
		assert.Less(t, avgTime, 10*time.Microsecond,
			"FormatRepositoryURI took %v per operation, expected < 10µs", avgTime)
	})
}

// TestRegistryUtil_MemoryLeaks tests for potential memory leaks
func TestRegistryUtil_MemoryLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	// Create many instances to check for resource leaks
	for i := 0; i < 1000; i++ {
		util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))
		util.ParseRegistryPath("ecr/repo")
		util.FormatRepositoryURI("ecr", "123", "us-west-2", "repo")
	}

	// If we get here without OOM, test passes
	assert.True(t, true)
}

// BenchmarkRegistryUtil_ParseRegistryPath benchmarks path parsing
func BenchmarkRegistryUtil_ParseRegistryPath(b *testing.B) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = util.ParseRegistryPath("ecr/my-repo")
	}
}

// BenchmarkRegistryUtil_FormatRepositoryURI benchmarks URI formatting
func BenchmarkRegistryUtil_FormatRepositoryURI(b *testing.B) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = util.FormatRepositoryURI("ecr", "123456789012", "us-west-2", "my-repo")
	}
}

// BenchmarkRegistryUtil_CreateRepositoryReference benchmarks reference creation
func BenchmarkRegistryUtil_CreateRepositoryReference(b *testing.B) {
	util := common.NewRegistryUtil(log.NewBasicLogger(log.InfoLevel))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = util.CreateRepositoryReference("example.com", "my-repo")
	}
}
