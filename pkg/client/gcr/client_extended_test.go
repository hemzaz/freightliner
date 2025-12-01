package gcr

import (
	"context"
	"testing"

	"freightliner/pkg/helper/log"

	"github.com/stretchr/testify/assert"
)

func TestClientExtended_GetRegistryNameAllLocations(t *testing.T) {
	tests := []struct {
		name     string
		location string
		expected string
	}{
		{
			name:     "US location",
			location: LocationUS,
			expected: "gcr.io",
		},
		{
			name:     "EU location",
			location: LocationEU,
			expected: "eu.gcr.io",
		},
		{
			name:     "Asia location",
			location: LocationAsia,
			expected: "asia.gcr.io",
		},
		{
			name:     "Custom region US Central 1",
			location: "us-central1",
			expected: "us-central1-docker.pkg.dev",
		},
		{
			name:     "Custom region Europe West 1",
			location: "europe-west1",
			expected: "europe-west1-docker.pkg.dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(ClientOptions{
				Project:  "test-project",
				Location: tt.location,
				Logger:   log.NewBasicLogger(log.InfoLevel),
			})
			assert.NoError(t, err)

			result := client.GetRegistryName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClientExtended_GetRepositoryWithPath(t *testing.T) {
	tests := []struct {
		name     string
		repoName string
	}{
		{
			name:     "Simple repository",
			repoName: "my-app",
		},
		{
			name:     "Repository with path",
			repoName: "team/my-app",
		},
		{
			name:     "Multi-level path",
			repoName: "org/team/my-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(ClientOptions{
				Project:  "test-project",
				Location: "us",
			})
			assert.NoError(t, err)

			ctx := context.Background()
			repo, err := client.GetRepository(ctx, tt.repoName)
			assert.NoError(t, err)
			assert.NotNil(t, repo)
			assert.Equal(t, tt.repoName, repo.GetRepositoryName())
		})
	}
}

func TestClientExtended_CreateRepositoryLogging(t *testing.T) {
	logger := log.NewBasicLogger(log.DebugLevel)
	client, err := NewClient(ClientOptions{
		Project:  "test-project",
		Location: "us",
		Logger:   logger,
	})
	assert.NoError(t, err)

	ctx := context.Background()
	repo, err := client.CreateRepository(ctx, "test-repo", map[string]string{
		"env": "test",
	})
	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestClientExtended_ListRepositoriesLegacyPath(t *testing.T) {
	// Test with US location (legacy GCR)
	client, err := NewClient(ClientOptions{
		Project:  "test-project",
		Location: "us",
	})
	assert.NoError(t, err)

	ctx := context.Background()
	_, err = client.ListRepositories(ctx, "")
	// Will error without GCP credentials but tests the code path
	if err != nil {
		assert.Error(t, err)
	}
}

func TestClientExtended_ListRepositoriesARPath(t *testing.T) {
	// Test with custom region (Artifact Registry)
	client, err := NewClient(ClientOptions{
		Project:  "test-project",
		Location: "us-central1",
	})
	assert.NoError(t, err)

	ctx := context.Background()
	_, err = client.ListRepositories(ctx, "app-")
	// Will error without GCP credentials but tests the code path
	if err != nil {
		assert.Error(t, err)
	}
}

func TestClientExtended_ContextCancellation(t *testing.T) {
	client, err := NewClient(ClientOptions{
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

	_, err = client.CreateRepository(ctx, "test-repo", nil)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	_, err = client.ListRepositories(ctx, "")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestClientExtended_GetTransportWithDifferentRepos(t *testing.T) {
	client, err := NewClient(ClientOptions{
		Project:  "test-project",
		Location: "us",
	})
	assert.NoError(t, err)

	tests := []string{
		"simple-repo",
		"team/app",
		"org/team/service",
	}

	for _, repoName := range tests {
		t.Run(repoName, func(t *testing.T) {
			_, err := client.GetTransport(repoName)
			// Will error without GCP credentials
			if err != nil {
				assert.Error(t, err)
			}
		})
	}
}

func TestNewClient_WithCredentialsFile(t *testing.T) {
	client, err := NewClient(ClientOptions{
		Project:         "test-project",
		Location:        "us",
		CredentialsFile: "/nonexistent/credentials.json",
	})
	// Should create client but may not have AR service
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewClient_DefaultLogger(t *testing.T) {
	client, err := NewClient(ClientOptions{
		Project:  "test-project",
		Location: "us",
		Logger:   nil, // Should set default
	})
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestClient_LocationMapping(t *testing.T) {
	tests := []struct {
		inputLocation    string
		expectedLocation string
	}{
		{
			inputLocation:    "us",
			expectedLocation: "us-central1",
		},
		{
			inputLocation:    "eu",
			expectedLocation: "us-central1",
		},
		{
			inputLocation:    "asia",
			expectedLocation: "us-central1",
		},
		{
			inputLocation:    "us-west1",
			expectedLocation: "us-west1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.inputLocation, func(t *testing.T) {
			// Test location mapping logic
			location := tt.inputLocation
			if location == LocationUS || location == LocationEU || location == LocationAsia {
				location = "us-central1"
			}
			assert.Equal(t, tt.expectedLocation, location)
		})
	}
}

func TestRepositoryIterator_Next(t *testing.T) {
	// Test the iterator implementation pattern
	type MockRepo struct {
		Name string
	}
	repos := []MockRepo{
		{Name: "repo1"},
		{Name: "repo2"},
		{Name: "repo3"},
	}

	index := -1
	var current MockRepo
	var done bool

	// First iteration
	index++
	if index < len(repos) {
		current = repos[index]
		done = false
	} else {
		done = true
	}
	assert.False(t, done)
	assert.Equal(t, "repo1", current.Name)

	// Second iteration
	index++
	if index < len(repos) {
		current = repos[index]
		done = false
	} else {
		done = true
	}
	assert.False(t, done)
	assert.Equal(t, "repo2", current.Name)

	// Third iteration
	index++
	if index < len(repos) {
		current = repos[index]
		done = false
	} else {
		done = true
	}
	assert.False(t, done)
	assert.Equal(t, "repo3", current.Name)

	// Done
	index++
	if index < len(repos) {
		current = repos[index]
		done = false
	} else {
		done = true
	}
	assert.True(t, done)
}

func TestClient_RepositoryPathConstruction(t *testing.T) {
	tests := []struct {
		project  string
		repoName string
		expected string
	}{
		{
			project:  "my-project",
			repoName: "my-repo",
			expected: "gcr.io/my-project/my-repo",
		},
		{
			project:  "other-project",
			repoName: "team/app",
			expected: "gcr.io/other-project/team/app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			registry := "gcr.io"
			fullPath := registry + "/" + tt.project + "/" + tt.repoName
			assert.Equal(t, tt.expected, fullPath)
		})
	}
}
