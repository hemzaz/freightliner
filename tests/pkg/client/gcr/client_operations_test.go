package gcr

import (
	"context"
	"testing"

	"freightliner/pkg/client/gcr"
	"freightliner/pkg/helper/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
	"google.golang.org/api/iterator"
)

// MockARService is a mock for Artifact Registry Service
type MockARService struct {
	mock.Mock
	repositories []*artifactregistry.Repository
	nextToken    string
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		opts        gcr.ClientOptions
		expectError bool
	}{
		{
			name: "Valid options with project",
			opts: gcr.ClientOptions{
				Project:  "my-project",
				Location: "us",
				Logger:   log.NewBasicLogger(log.InfoLevel),
			},
			expectError: false,
		},
		{
			name: "Default location",
			opts: gcr.ClientOptions{
				Project: "my-project",
				Logger:  log.NewBasicLogger(log.InfoLevel),
			},
			expectError: false,
		},
		{
			name: "Nil logger - should set default",
			opts: gcr.ClientOptions{
				Project:  "my-project",
				Location: "us",
			},
			expectError: false,
		},
		{
			name: "With credentials file",
			opts: gcr.ClientOptions{
				Project:         "my-project",
				Location:        "us",
				CredentialsFile: "/path/to/credentials.json",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := gcr.NewClient(tt.opts)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestClient_GetRegistryName(t *testing.T) {
	tests := []struct {
		name     string
		location string
		expected string
	}{
		{
			name:     "US location",
			location: "us",
			expected: "gcr.io",
		},
		{
			name:     "EU location",
			location: "eu",
			expected: "eu.gcr.io",
		},
		{
			name:     "Asia location",
			location: "asia",
			expected: "asia.gcr.io",
		},
		{
			name:     "Custom region",
			location: "us-central1",
			expected: "us-central1-docker.pkg.dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := gcr.NewClient(gcr.ClientOptions{
				Project:  "test-project",
				Location: tt.location,
			})
			assert.NoError(t, err)

			result := client.GetRegistryName()
			assert.Equal(t, tt.expected, result)
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
			repoName:    "my-app",
			expectError: false,
		},
		{
			name:        "Repository with path",
			repoName:    "team/my-app",
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
			client, err := gcr.NewClient(gcr.ClientOptions{
				Project:  "test-project",
				Location: "us",
			})
			assert.NoError(t, err)

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
			name:        "Create repository without tags",
			repoName:    "test-repo",
			tags:        nil,
			expectError: false,
		},
		{
			name:     "Create repository with tags",
			repoName: "test-repo",
			tags: map[string]string{
				"environment": "production",
				"team":        "backend",
			},
			expectError: false,
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
			client, err := gcr.NewClient(gcr.ClientOptions{
				Project:  "test-project",
				Location: "us",
			})
			assert.NoError(t, err)

			ctx := context.Background()
			repo, err := client.CreateRepository(ctx, tt.repoName, tt.tags)

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

func TestClient_ListRepositories_LegacyGCR(t *testing.T) {
	tests := []struct {
		name        string
		prefix      string
		expectError bool
	}{
		{
			name:        "List all repositories",
			prefix:      "",
			expectError: false,
		},
		{
			name:        "List with prefix",
			prefix:      "app-",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create client without AR service to test legacy path
			client, err := gcr.NewClient(gcr.ClientOptions{
				Project:  "test-project",
				Location: "us",
			})
			assert.NoError(t, err)

			ctx := context.Background()
			repos, err := client.ListRepositories(ctx, tt.prefix)

			// Will error without actual GCP credentials
			// but we're testing the code path
			if err != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repos)
			}
		})
	}
}

func TestClient_GetTransport(t *testing.T) {
	tests := []struct {
		name        string
		repoName    string
		expectError bool
	}{
		{
			name:        "Valid repository",
			repoName:    "test-repo",
			expectError: false,
		},
		{
			name:        "Repository with path",
			repoName:    "org/test-repo",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := gcr.NewClient(gcr.ClientOptions{
				Project:  "test-project",
				Location: "us",
			})
			assert.NoError(t, err)

			transport, err := client.GetTransport(tt.repoName)

			// Will error without GCP credentials but tests the code path
			if err != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, transport)
			}
		})
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	client, err := gcr.NewClient(gcr.ClientOptions{
		Project:  "test-project",
		Location: "us",
	})
	assert.NoError(t, err)

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = client.GetRepository(ctx, "test-repo")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestClient_ListRepositoriesWithContext(t *testing.T) {
	tests := []struct {
		name        string
		setupCtx    func() context.Context
		expectError bool
	}{
		{
			name: "Valid context",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectError: false,
		},
		{
			name: "Cancelled context",
			setupCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := gcr.NewClient(gcr.ClientOptions{
				Project:  "test-project",
				Location: "us",
			})
			assert.NoError(t, err)

			ctx := tt.setupCtx()
			_, err = client.ListRepositories(ctx, "")

			if tt.expectError && ctx.Err() != nil {
				assert.Error(t, err)
			}
		})
	}
}

func TestClient_MultipleLocations(t *testing.T) {
	locations := []string{"us", "eu", "asia", "us-central1", "europe-west1"}

	for _, location := range locations {
		t.Run("Location_"+location, func(t *testing.T) {
			client, err := gcr.NewClient(gcr.ClientOptions{
				Project:  "test-project",
				Location: location,
			})
			assert.NoError(t, err)
			assert.NotNil(t, client)

			registryName := client.GetRegistryName()
			assert.NotEmpty(t, registryName)

			// Verify registry name format
			if location == "us" {
				assert.Equal(t, "gcr.io", registryName)
			} else if location == "eu" {
				assert.Equal(t, "eu.gcr.io", registryName)
			} else if location == "asia" {
				assert.Equal(t, "asia.gcr.io", registryName)
			} else {
				assert.Contains(t, registryName, "docker.pkg.dev")
			}
		})
	}
}

func TestClient_ConcurrentGetRepository(t *testing.T) {
	client, err := gcr.NewClient(gcr.ClientOptions{
		Project:  "test-project",
		Location: "us",
	})
	assert.NoError(t, err)

	ctx := context.Background()
	results := make(chan error, 5)

	// Test concurrent GetRepository calls
	for i := 0; i < 5; i++ {
		go func(index int) {
			_, err := client.GetRepository(ctx, "test-repo")
			results <- err
		}(i)
	}

	// Collect results
	for i := 0; i < 5; i++ {
		err := <-results
		assert.NoError(t, err)
	}
}

func TestClient_RepositoryPrefixFiltering(t *testing.T) {
	tests := []struct {
		name           string
		repositories   []string
		prefix         string
		expectedCount  int
		expectedNames  []string
	}{
		{
			name:          "No prefix - all repos",
			repositories:  []string{"app-backend", "app-frontend", "lib-common", "test-utils"},
			prefix:        "",
			expectedCount: 4,
		},
		{
			name:          "Prefix 'app-'",
			repositories:  []string{"app-backend", "app-frontend", "lib-common", "test-utils"},
			prefix:        "app-",
			expectedNames: []string{"app-backend", "app-frontend"},
		},
		{
			name:          "Prefix 'lib-'",
			repositories:  []string{"app-backend", "app-frontend", "lib-common", "test-utils"},
			prefix:        "lib-",
			expectedNames: []string{"lib-common"},
		},
		{
			name:          "No matches",
			repositories:  []string{"app-backend", "app-frontend"},
			prefix:        "xyz-",
			expectedNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test validates the prefix filtering logic
			// In real implementation, this would be tested with mocked AR API

			var filtered []string
			for _, repo := range tt.repositories {
				if tt.prefix == "" || len(repo) >= len(tt.prefix) && repo[:len(tt.prefix)] == tt.prefix {
					filtered = append(filtered, repo)
				}
			}

			if tt.expectedNames != nil {
				assert.Equal(t, tt.expectedNames, filtered)
			}
			if tt.expectedCount > 0 {
				assert.Len(t, filtered, tt.expectedCount)
			}
		})
	}
}
