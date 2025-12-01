package common

import (
	"context"
	"testing"

	"freightliner/pkg/helper/log"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseRepository(t *testing.T) {
	repoRef, err := name.NewRepository("example.com/test-repo")
	require.NoError(t, err)

	tests := []struct {
		name       string
		repoName   string
		logger     log.Logger
		repository name.Repository
	}{
		{
			name:       "With logger",
			repoName:   "test-repo",
			logger:     log.NewBasicLogger(log.InfoLevel),
			repository: repoRef,
		},
		{
			name:       "Without logger (should create default)",
			repoName:   "test-repo",
			logger:     nil,
			repository: repoRef,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewBaseRepository(BaseRepositoryOptions{
				Name:       tt.repoName,
				Repository: tt.repository,
				Logger:     tt.logger,
			})

			assert.NotNil(t, repo)
			assert.Equal(t, tt.repoName, repo.name)
			assert.NotNil(t, repo.logger)
			assert.NotNil(t, repo.images)
		})
	}
}

func TestBaseRepository_GetName(t *testing.T) {
	repoRef, err := name.NewRepository("example.com/test-repo")
	require.NoError(t, err)

	repo := NewBaseRepository(BaseRepositoryOptions{
		Name:       "test-repo",
		Repository: repoRef,
		Logger:     log.NewBasicLogger(log.InfoLevel),
	})

	assert.Equal(t, "test-repo", repo.GetName())
}

func TestBaseRepository_GetURI(t *testing.T) {
	repoRef, err := name.NewRepository("example.com/test-repo")
	require.NoError(t, err)

	repo := NewBaseRepository(BaseRepositoryOptions{
		Name:       "test-repo",
		Repository: repoRef,
		Logger:     log.NewBasicLogger(log.InfoLevel),
	})

	uri := repo.GetURI()
	assert.Contains(t, uri, "example.com")
	assert.Contains(t, uri, "test-repo")
}

func TestBaseRepository_CreateTagReference(t *testing.T) {
	repoRef, err := name.NewRepository("example.com/test-repo")
	require.NoError(t, err)

	repo := NewBaseRepository(BaseRepositoryOptions{
		Name:       "test-repo",
		Repository: repoRef,
		Logger:     log.NewBasicLogger(log.InfoLevel),
	})

	tests := []struct {
		name        string
		tagName     string
		shouldError bool
	}{
		{
			name:        "Valid tag",
			tagName:     "latest",
			shouldError: false,
		},
		{
			name:        "Valid version tag",
			tagName:     "v1.0.0",
			shouldError: false,
		},
		{
			name:        "Empty tag",
			tagName:     "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagRef, err := repo.CreateTagReference(tt.tagName)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, tagRef.String(), tt.tagName)
			}
		})
	}
}

func TestBaseRepository_CacheImage(t *testing.T) {
	repoRef, err := name.NewRepository("example.com/test-repo")
	require.NoError(t, err)

	repo := NewBaseRepository(BaseRepositoryOptions{
		Name:       "test-repo",
		Repository: repoRef,
		Logger:     log.NewBasicLogger(log.InfoLevel),
	})

	// Cache should be empty
	assert.Len(t, repo.images, 0)

	// Cache a mock image (using nil as placeholder since we can't create real images in unit tests)
	repo.CacheImage("latest", nil)

	// Verify it's cached
	assert.Len(t, repo.images, 1)
	assert.Contains(t, repo.images, "latest")
}

func TestBaseRepository_ClearCache(t *testing.T) {
	repoRef, err := name.NewRepository("example.com/test-repo")
	require.NoError(t, err)

	repo := NewBaseRepository(BaseRepositoryOptions{
		Name:       "test-repo",
		Repository: repoRef,
		Logger:     log.NewBasicLogger(log.InfoLevel),
	})

	// Add some mock cached data
	repo.CacheImage("latest", nil)
	repo.CacheImage("v1.0.0", nil)
	repo.tags = []string{"latest", "v1.0.0"}

	// Verify cache is populated
	assert.Len(t, repo.images, 2)
	assert.Len(t, repo.tags, 2)

	// Clear cache
	repo.ClearCache()

	// Verify cache is empty
	assert.Len(t, repo.images, 0)
	assert.Nil(t, repo.tags)
}

func TestBaseRepository_ConcurrentCacheAccess(t *testing.T) {
	repoRef, err := name.NewRepository("example.com/test-repo")
	require.NoError(t, err)

	repo := NewBaseRepository(BaseRepositoryOptions{
		Name:       "test-repo",
		Repository: repoRef,
		Logger:     log.NewBasicLogger(log.InfoLevel),
	})

	// Test concurrent cache operations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			tagName := "tag-" + string(rune('0'+idx))
			repo.CacheImage(tagName, nil)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have cached all images
	assert.Len(t, repo.images, 10)
}

func TestBaseRepository_GetRemoteImage(t *testing.T) {
	repoRef, err := name.NewRepository("example.com/test-repo")
	require.NoError(t, err)

	repo := NewBaseRepository(BaseRepositoryOptions{
		Name:       "test-repo",
		Repository: repoRef,
		Logger:     log.NewBasicLogger(log.InfoLevel),
	})

	ctx := context.Background()

	// Test with nil reference
	_, err = repo.GetRemoteImage(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reference cannot be nil")
}
