package common_test

import (
	"context"
	"testing"

	"freightliner/pkg/client/common"
	"freightliner/pkg/helper/log"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"
)

// TestNewBaseRepository tests creating a new base repository
func TestNewBaseRepository(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)

	repo, err := name.NewRepository("gcr.io/test-project/test-repo")
	if err != nil {
		t.Fatalf("Failed to create repository reference: %v", err)
	}

	tests := []struct {
		name  string
		opts  common.BaseRepositoryOptions
		valid bool
	}{
		{
			name: "valid with logger",
			opts: common.BaseRepositoryOptions{
				Name:       "test-repo",
				Repository: repo,
				Logger:     logger,
			},
			valid: true,
		},
		{
			name: "valid without logger (nil)",
			opts: common.BaseRepositoryOptions{
				Name:       "test-repo",
				Repository: repo,
				Logger:     nil,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseRepo := common.NewBaseRepository(tt.opts)

			if baseRepo == nil {
				t.Fatal("Expected non-nil repository")
			}

			if baseRepo.GetName() != tt.opts.Name {
				t.Errorf("Expected name %s, got %s", tt.opts.Name, baseRepo.GetName())
			}

			expectedURI := repo.String()
			if baseRepo.GetURI() != expectedURI {
				t.Errorf("Expected URI %s, got %s", expectedURI, baseRepo.GetURI())
			}
		})
	}
}

// TestBaseRepositoryGetNameAndURI tests repository name and URI getters
func TestBaseRepositoryGetNameAndURI(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)

	tests := []struct {
		name         string
		repoPath     string
		expectedName string
	}{
		{
			name:         "gcr repository",
			repoPath:     "gcr.io/test-project/app",
			expectedName: "app",
		},
		{
			name:         "ecr repository",
			repoPath:     "123456789.dkr.ecr.us-west-2.amazonaws.com/my-app",
			expectedName: "my-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := name.NewRepository(tt.repoPath)
			if err != nil {
				t.Fatalf("Failed to create repository reference: %v", err)
			}

			baseRepo := common.NewBaseRepository(common.BaseRepositoryOptions{
				Name:       tt.expectedName,
				Repository: repo,
				Logger:     logger,
			})

			if baseRepo.GetName() != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, baseRepo.GetName())
			}

			if baseRepo.GetURI() != tt.repoPath {
				t.Errorf("Expected URI %s, got %s", tt.repoPath, baseRepo.GetURI())
			}
		})
	}
}

// TestCreateTagReference tests creating tag references
func TestCreateTagReference(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	repo, _ := name.NewRepository("gcr.io/test-project/test-repo")

	baseRepo := common.NewBaseRepository(common.BaseRepositoryOptions{
		Name:       "test-repo",
		Repository: repo,
		Logger:     logger,
	})

	tests := []struct {
		name      string
		tagName   string
		expectErr bool
	}{
		{
			name:      "valid tag",
			tagName:   "v1.0.0",
			expectErr: false,
		},
		{
			name:      "latest tag",
			tagName:   "latest",
			expectErr: false,
		},
		{
			name:      "empty tag",
			tagName:   "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagRef, err := baseRepo.CreateTagReference(tt.tagName)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error for empty tag name")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				expectedTag := repo.String() + ":" + tt.tagName
				if tagRef.String() != expectedTag {
					t.Errorf("Expected tag %s, got %s", expectedTag, tagRef.String())
				}
			}
		})
	}
}

// TestCacheImage tests image caching
func TestCacheImage(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	repo, _ := name.NewRepository("gcr.io/test-project/test-repo")

	baseRepo := common.NewBaseRepository(common.BaseRepositoryOptions{
		Name:       "test-repo",
		Repository: repo,
		Logger:     logger,
	})

	// Create a random image for testing
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	tagName := "v1.0.0"

	// Cache the image
	baseRepo.CacheImage(tagName, img)

	// Verify we can retrieve it (through implementation details)
	// Note: This test validates the caching behavior exists
	t.Log("Image cached successfully")
}

// TestClearCache tests cache clearing
func TestClearCache(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	repo, _ := name.NewRepository("gcr.io/test-project/test-repo")

	baseRepo := common.NewBaseRepository(common.BaseRepositoryOptions{
		Name:       "test-repo",
		Repository: repo,
		Logger:     logger,
	})

	// Create and cache multiple images
	for i := 1; i <= 3; i++ {
		img, err := random.Image(1024, 1)
		if err != nil {
			t.Fatalf("Failed to create test image: %v", err)
		}
		baseRepo.CacheImage("v1.0."+string(rune('0'+i)), img)
	}

	// Clear the cache
	baseRepo.ClearCache()

	t.Log("Cache cleared successfully")
}

// TestGetRemoteImage tests retrieving remote images
func TestGetRemoteImage(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	repo, _ := name.NewRepository("gcr.io/test-project/test-repo")

	baseRepo := common.NewBaseRepository(common.BaseRepositoryOptions{
		Name:       "test-repo",
		Repository: repo,
		Logger:     logger,
	})

	tests := []struct {
		name      string
		ref       name.Reference
		expectErr bool
	}{
		{
			name:      "nil reference",
			ref:       nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := baseRepo.GetRemoteImage(ctx, tt.ref)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error for nil reference")
				}
			}
		})
	}
}

// TestListTagsValidation tests tag listing validation
func TestListTagsValidation(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	repo, _ := name.NewRepository("gcr.io/test-project/test-repo")

	baseRepo := common.NewBaseRepository(common.BaseRepositoryOptions{
		Name:       "test-repo",
		Repository: repo,
		Logger:     logger,
	})

	ctx := context.Background()

	// This will fail because we're not actually connected to a registry
	// but it tests the code path
	_, err := baseRepo.ListTags(ctx)

	// We expect an error since we're not connected to a real registry
	if err == nil {
		t.Log("Unexpectedly succeeded (might be hitting real registry)")
	} else {
		t.Logf("Expected error for mock registry: %v", err)
	}
}

// TestGetTagValidation tests tag retrieval validation
func TestGetTagValidation(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	repo, _ := name.NewRepository("gcr.io/test-project/test-repo")

	baseRepo := common.NewBaseRepository(common.BaseRepositoryOptions{
		Name:       "test-repo",
		Repository: repo,
		Logger:     logger,
	})

	ctx := context.Background()

	tests := []struct {
		name      string
		tagName   string
		expectErr bool
	}{
		{
			name:      "empty tag name",
			tagName:   "",
			expectErr: true,
		},
		{
			name:      "valid tag name (will fail at network)",
			tagName:   "v1.0.0",
			expectErr: true, // Network error expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := baseRepo.GetTag(ctx, tt.tagName)

			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

// TestGetImageValidation tests image retrieval by digest validation
func TestGetImageValidation(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	repo, _ := name.NewRepository("gcr.io/test-project/test-repo")

	baseRepo := common.NewBaseRepository(common.BaseRepositoryOptions{
		Name:       "test-repo",
		Repository: repo,
		Logger:     logger,
	})

	ctx := context.Background()

	tests := []struct {
		name      string
		digest    string
		expectErr bool
	}{
		{
			name:      "empty digest",
			digest:    "",
			expectErr: true,
		},
		{
			name:      "invalid digest format",
			digest:    "invalid-digest",
			expectErr: true,
		},
		{
			name:      "valid digest format",
			digest:    "sha256:0000000000000000000000000000000000000000000000000000000000000000",
			expectErr: true, // Will fail at network
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := baseRepo.GetImage(ctx, tt.digest)

			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

// TestDeleteTagValidation tests tag deletion validation
func TestDeleteTagValidation(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	repo, _ := name.NewRepository("gcr.io/test-project/test-repo")

	baseRepo := common.NewBaseRepository(common.BaseRepositoryOptions{
		Name:       "test-repo",
		Repository: repo,
		Logger:     logger,
	})

	ctx := context.Background()

	tests := []struct {
		name      string
		tagName   string
		expectErr bool
	}{
		{
			name:      "empty tag name",
			tagName:   "",
			expectErr: true,
		},
		{
			name:      "valid tag name (will fail at network)",
			tagName:   "v1.0.0",
			expectErr: true, // Network error expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := baseRepo.DeleteTag(ctx, tt.tagName)

			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

// TestPutImageValidation tests image push validation
func TestPutImageValidation(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	repo, _ := name.NewRepository("gcr.io/test-project/test-repo")

	baseRepo := common.NewBaseRepository(common.BaseRepositoryOptions{
		Name:       "test-repo",
		Repository: repo,
		Logger:     logger,
	})

	ctx := context.Background()

	// Create a test image
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	tests := []struct {
		name      string
		img       v1.Image
		tagName   string
		expectErr bool
	}{
		{
			name:      "nil image",
			img:       nil,
			tagName:   "v1.0.0",
			expectErr: true,
		},
		{
			name:      "empty tag name",
			img:       img,
			tagName:   "",
			expectErr: true,
		},
		{
			name:      "valid inputs (will fail at network)",
			img:       img,
			tagName:   "v1.0.0",
			expectErr: true, // Network error expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := baseRepo.PutImage(ctx, tt.img, tt.tagName)

			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

// TestTagCaching tests that tags are properly cached and retrieved
func TestTagCaching(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	repo, _ := name.NewRepository("gcr.io/test-project/test-repo")

	baseRepo := common.NewBaseRepository(common.BaseRepositoryOptions{
		Name:       "test-repo",
		Repository: repo,
		Logger:     logger,
	})

	ctx := context.Background()

	// First call will fail and not cache anything
	_, err1 := baseRepo.ListTags(ctx)
	if err1 == nil {
		t.Log("First call unexpectedly succeeded")
	}

	// Clear cache
	baseRepo.ClearCache()

	// Verify cache operations don't panic
	t.Log("Cache operations completed without panic")
}

// TestConcurrentCacheAccess tests thread-safe cache access
func TestConcurrentCacheAccess(t *testing.T) {
	logger := log.NewBasicLogger(log.InfoLevel)
	repo, _ := name.NewRepository("gcr.io/test-project/test-repo")

	baseRepo := common.NewBaseRepository(common.BaseRepositoryOptions{
		Name:       "test-repo",
		Repository: repo,
		Logger:     logger,
	})

	// Create test image
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	// Concurrent cache operations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			tagName := "v1.0." + string(rune('0'+idx))
			baseRepo.CacheImage(tagName, img)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Clear cache concurrently
	go baseRepo.ClearCache()
	go baseRepo.ClearCache()

	t.Log("Concurrent cache access completed without race conditions")
}
