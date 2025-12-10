package common

import (
	"context"
	"testing"

	"freightliner/pkg/helper/log"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseClient(t *testing.T) {
	tests := []struct {
		name         string
		registryName string
		logger       log.Logger
	}{
		{
			name:         "With logger",
			registryName: "example.com",
			logger:       log.NewBasicLogger(log.InfoLevel),
		},
		{
			name:         "Without logger (should create default)",
			registryName: "example.com",
			logger:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewBaseClient(BaseClientOptions{
				RegistryName: tt.registryName,
				Logger:       tt.logger,
			})

			assert.NotNil(t, client)
			assert.Equal(t, tt.registryName, client.registryName)
			assert.NotNil(t, client.logger) // Should always have a logger
			assert.NotNil(t, client.util)
			assert.NotNil(t, client.repositories)
		})
	}
}

func TestBaseClient_GetRegistryName(t *testing.T) {
	registryName := "registry.example.com"
	client := NewBaseClient(BaseClientOptions{
		RegistryName: registryName,
		Logger:       log.NewBasicLogger(log.InfoLevel),
	})

	assert.Equal(t, registryName, client.GetRegistryName())
}

func TestBaseClient_GetRepository(t *testing.T) {
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
			name:        "Empty repository name",
			repoName:    "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewBaseClient(BaseClientOptions{
				RegistryName: "example.com",
				Logger:       log.NewBasicLogger(log.InfoLevel),
			})

			ctx := context.Background()
			repo, err := client.GetRepository(ctx, tt.repoName)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, repo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)

				// Second call should return cached repository
				repo2, err := client.GetRepository(ctx, tt.repoName)
				assert.NoError(t, err)
				assert.Equal(t, repo, repo2)
			}
		})
	}
}

func TestBaseClient_GetCachedRepository(t *testing.T) {
	client := NewBaseClient(BaseClientOptions{
		RegistryName: "example.com",
		Logger:       log.NewBasicLogger(log.InfoLevel),
	})

	ctx := context.Background()
	repoName := "test-repo"

	// Factory function to create mock repository
	factory := func(ref name.Repository) interface{} {
		return &BaseRepository{
			name:       repoName,
			repository: ref,
			logger:     client.logger,
		}
	}

	// First call - should create new repository
	repo1, err := client.GetCachedRepository(ctx, repoName, factory)
	assert.NoError(t, err)
	assert.NotNil(t, repo1)

	// Second call - should return cached repository
	repo2, err := client.GetCachedRepository(ctx, repoName, factory)
	assert.NoError(t, err)
	assert.Equal(t, repo1, repo2)

	// Test with empty repo name
	_, err = client.GetCachedRepository(ctx, "", factory)
	assert.Error(t, err)
}

func TestBaseClient_GetRemoteOptions(t *testing.T) {
	client := NewBaseClient(BaseClientOptions{
		RegistryName: "example.com",
		Logger:       log.NewBasicLogger(log.InfoLevel),
	})

	// Test with nil transport
	opts := client.GetRemoteOptions(nil)
	assert.Len(t, opts, 0)

	// Test with mock transport
	mockTransport := &mockRoundTripper{}
	opts = client.GetRemoteOptions(mockTransport)
	assert.Len(t, opts, 1)
}

func TestBaseClient_ValidateRepositoryName(t *testing.T) {
	client := NewBaseClient(BaseClientOptions{
		RegistryName: "example.com",
		Logger:       log.NewBasicLogger(log.InfoLevel),
	})

	tests := []struct {
		name        string
		repoName    string
		shouldError bool
	}{
		{
			name:        "Valid repo name",
			repoName:    "my-repo",
			shouldError: false,
		},
		{
			name:        "Valid repo with path",
			repoName:    "org/my-repo",
			shouldError: false,
		},
		{
			name:        "Empty repo name",
			repoName:    "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.ValidateRepositoryName(tt.repoName)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBaseClient_LogOperation(t *testing.T) {
	client := NewBaseClient(BaseClientOptions{
		RegistryName: "example.com",
		Logger:       log.NewBasicLogger(log.InfoLevel),
	})

	ctx := context.Background()

	// Should not panic
	assert.NotPanics(t, func() {
		client.LogOperation(ctx, "pull", "my-repo", map[string]interface{}{
			"tag": "latest",
		})
	})

	// Test with nil extra fields
	assert.NotPanics(t, func() {
		client.LogOperation(ctx, "push", "another-repo", nil)
	})
}

func TestBaseClient_ConcurrentAccess(t *testing.T) {
	client := NewBaseClient(BaseClientOptions{
		RegistryName: "example.com",
		Logger:       log.NewBasicLogger(log.InfoLevel),
	})

	ctx := context.Background()
	repoNames := []string{"repo1", "repo2", "repo3", "repo4", "repo5"}

	// Test concurrent repository creation
	done := make(chan bool, len(repoNames))

	for _, name := range repoNames {
		go func(n string) {
			_, err := client.GetRepository(ctx, n)
			require.NoError(t, err)
			done <- true
		}(name)
	}

	// Wait for all goroutines to complete
	for i := 0; i < len(repoNames); i++ {
		<-done
	}

	// Verify all repositories are cached
	for _, name := range repoNames {
		repo, err := client.GetRepository(ctx, name)
		assert.NoError(t, err)
		assert.NotNil(t, repo)
	}
}
