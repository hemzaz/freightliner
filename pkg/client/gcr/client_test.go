package gcr

import (
	"context"
	"errors"
	"os"
	"testing"

	freightliner_log "freightliner/pkg/helper/log"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/stretchr/testify/assert"
)

// Note: Mock types have been moved to pkg/testing/mocks for reuse
// Use mocks.MockGoogleCatalogClient and mocks.MockArtifactRegistryClient for testing

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		location    string
		project     string
		expectedErr bool
	}{
		{
			name:        "Valid GCR registry",
			location:    "us",
			project:     "test-project",
			expectedErr: false,
		},
		{
			name:        "Valid Artifact Registry",
			location:    "us-central1",
			project:     "test-project",
			expectedErr: false,
		},
		{
			name:        "Invalid registry with path",
			location:    "us",
			project:     "",    // Empty project should cause an error
			expectedErr: false, // The client creation should still work, but operations would fail
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClient(ClientOptions{
				Location: tc.location,
				Project:  tc.project,
				// Use nil for auth to use default auth
			})

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tc.location, client.location)
				assert.Equal(t, tc.project, client.project)
			}
		})
	}
}

func TestClientListRepositories(t *testing.T) {
	// Skip this test unless integration tests are explicitly enabled
	// This test makes real API calls to GCR and requires valid GCP credentials
	if os.Getenv("ENABLE_GCR_INTEGRATION_TESTS") != "true" {
		t.Skip("GCR integration tests disabled. Set ENABLE_GCR_INTEGRATION_TESTS=true to enable")
	}

	// This function uses real GCR API calls and requires valid credentials
	tests := []struct {
		name        string
		location    string
		prefix      string
		expected    []string
		expectedErr bool
	}{
		{
			name:        "List all repositories",
			location:    "us",
			prefix:      "",
			expected:    []string{"repo1", "repo2", "testing/repo3", "testing/repo4"},
			expectedErr: false,
		},
		{
			name:        "List with prefix",
			location:    "us",
			prefix:      "testing",
			expected:    []string{"testing/repo3", "testing/repo4"},
			expectedErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Import the logger package
			logger := freightliner_log.NewLogger()

			// Mock client with required fields for the test
			client := &Client{
				project:  "mock-project",
				location: tc.location,
				logger:   logger,
			}

			repos, err := client.ListRepositories(context.Background(), tc.prefix)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tc.expected, repos)
			}
		})
	}
}

func TestClientGetRepository(t *testing.T) {
	tests := []struct {
		name             string
		registry         string
		repoName         string
		expectedRepoName string
		expectedErr      bool
		expectedErrType  error
	}{
		{
			name:             "Valid repository",
			registry:         "gcr.io",
			repoName:         "project/repo",
			expectedRepoName: "project/repo",
			expectedErr:      false,
		},
		{
			name:             "Empty repository name",
			registry:         "gcr.io",
			repoName:         "",
			expectedRepoName: "",
			expectedErr:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Registry variable not used currently
			_, err := name.NewRegistry(tc.registry)
			assert.NoError(t, err)

			// Mock client with required fields for the test
			client := &Client{
				project:  "mock-project",
				location: tc.registry,
				logger:   nil,
			}

			repo, err := client.GetRepository(context.Background(), tc.repoName)
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
		})
	}
}

func TestParseGCRRepository(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		registry           string
		expectedRegistry   string
		expectedRepository string
		expectedErr        bool
	}{
		{
			name:               "Full GCR URI",
			input:              "gcr.io/project/repo-name",
			registry:           "gcr.io",
			expectedRegistry:   "gcr.io",
			expectedRepository: "project/repo-name",
			expectedErr:        false,
		},
		{
			name:               "Repository with suffix",
			input:              "gcr.io/project/repo-name:latest",
			registry:           "gcr.io",
			expectedRegistry:   "gcr.io",
			expectedRepository: "project/repo-name",
			expectedErr:        false,
		},
		{
			name:               "Repository with deep path",
			input:              "gcr.io/project/group/repo-name",
			registry:           "gcr.io",
			expectedRegistry:   "gcr.io",
			expectedRepository: "project/group/repo-name",
			expectedErr:        false,
		},
		{
			name:               "Simple repository (error expected)",
			input:              "repo-name",
			registry:           "gcr.io",
			expectedRegistry:   "",
			expectedRepository: "",
			expectedErr:        true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// This registry is actually used by parseGCRRepository
			reg, _ := name.NewRegistry(tc.registry)

			registry, repository, err := parseGCRRepository(tc.input, reg)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedRegistry, registry.String())
				assert.Equal(t, tc.expectedRepository, repository)
			}
		})
	}
}

func TestExtractProjectFromRepository(t *testing.T) {
	tests := []struct {
		name            string
		repository      string
		expectedProject string
		expectedErr     bool
	}{
		{
			name:            "Valid project/repo path",
			repository:      "project/repo",
			expectedProject: "project",
			expectedErr:     false,
		},
		{
			name:            "Valid project/path/repo path",
			repository:      "project/path/repo",
			expectedProject: "project",
			expectedErr:     false,
		},
		{
			name:            "Invalid repo without project",
			repository:      "repo",
			expectedProject: "",
			expectedErr:     true,
		},
		{
			name:            "Empty string",
			repository:      "",
			expectedProject: "",
			expectedErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			project, err := extractProjectFromRepository(tc.repository)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedProject, project)
			}
		})
	}
}

func TestIsGCRRegistry(t *testing.T) {
	tests := []struct {
		name     string
		registry string
		expected bool
	}{
		{
			name:     "Primary GCR domain",
			registry: "gcr.io",
			expected: true,
		},
		{
			name:     "Regional GCR domain",
			registry: "us.gcr.io",
			expected: true,
		},
		{
			name:     "European GCR domain",
			registry: "eu.gcr.io",
			expected: true,
		},
		{
			name:     "Artifact Registry",
			registry: "us-central1-docker.pkg.dev",
			expected: false,
		},
		{
			name:     "Docker Hub",
			registry: "docker.io",
			expected: false,
		},
		{
			name:     "Empty string",
			registry: "",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isGCRRegistry(tc.registry)
			assert.Equal(t, tc.expected, result)
		})
	}
}
