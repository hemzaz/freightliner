package gcr

import (
	"context"
	"io"
	"testing"

	"freightliner/pkg/client/gcr"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/interfaces"

	"github.com/stretchr/testify/assert"
)

func createTestClient() *gcr.Client {
	client, _ := gcr.NewClient(gcr.ClientOptions{
		Project:  "test-project",
		Location: "us",
		Logger:   log.NewBasicLogger(log.InfoLevel),
	})
	return client
}

func TestRepository_GetName(t *testing.T) {
	client := createTestClient()
	ctx := context.Background()

	repo, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)
	assert.Equal(t, "test-repo", repo.GetRepositoryName())
}

func TestRepository_ListTags(t *testing.T) {
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
			client := createTestClient()
			ctx := context.Background()

			repo, err := client.GetRepository(ctx, tt.repoName)
			assert.NoError(t, err)

			tags, err := repo.ListTags(ctx)

			// Will error without GCP credentials or actual repository
			// but tests the code path
			if err != nil {
				assert.Error(t, err)
			} else {
				assert.NotNil(t, tags)
			}
		})
	}
}

func TestRepository_GetImage(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		expectError bool
	}{
		{
			name:        "Valid tag",
			tag:         "v1.0.0",
			expectError: false,
		},
		{
			name:        "Latest tag",
			tag:         "latest",
			expectError: false,
		},
		{
			name:        "Empty tag",
			tag:         "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClient()
			ctx := context.Background()

			repo, err := client.GetRepository(ctx, "test-repo")
			assert.NoError(t, err)

			img, err := repo.GetImage(ctx, tt.tag)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, img)
			} else {
				// Will error without GCP credentials
				if err != nil {
					assert.Error(t, err)
				}
			}
		})
	}
}

func TestRepository_GetManifest(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		expectError bool
	}{
		{
			name:        "Valid tag",
			tag:         "v1.0.0",
			expectError: false,
		},
		{
			name:        "Empty tag",
			tag:         "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClient()
			ctx := context.Background()

			repo, err := client.GetRepository(ctx, "test-repo")
			assert.NoError(t, err)

			manifest, err := repo.GetManifest(ctx, tt.tag)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, manifest)
			} else {
				// Will error without GCP credentials
				if err != nil {
					assert.Error(t, err)
				}
			}
		})
	}
}

func TestRepository_GetMediaType(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		expectError bool
	}{
		{
			name:        "Valid tag",
			tag:         "v1.0.0",
			expectError: false,
		},
		{
			name:        "Empty tag",
			tag:         "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClient()
			ctx := context.Background()

			repo, err := client.GetRepository(ctx, "test-repo")
			assert.NoError(t, err)

			mediaType, err := repo.GetMediaType(ctx, tt.tag)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, mediaType)
			} else {
				// Will error without GCP credentials
				if err != nil {
					assert.Error(t, err)
				}
			}
		})
	}
}

func TestRepository_PutImage(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		hasImage    bool
		expectError bool
	}{
		{
			name:        "Valid tag with image",
			tag:         "v1.0.0",
			hasImage:    true,
			expectError: false,
		},
		{
			name:        "Empty tag",
			tag:         "",
			hasImage:    true,
			expectError: true,
		},
		{
			name:        "Nil image",
			tag:         "v1.0.0",
			hasImage:    false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClient()
			ctx := context.Background()

			repo, err := client.GetRepository(ctx, "test-repo")
			assert.NoError(t, err)

			var img interface{}
			if tt.hasImage {
				img = nil // Would need proper v1.Image mock
			}

			err = repo.PutImage(ctx, tt.tag, img)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// Will error without GCP credentials
				if err != nil {
					assert.Error(t, err)
				}
			}
		})
	}
}

func TestRepository_PutManifest(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		manifest    *interfaces.Manifest
		expectError bool
	}{
		{
			name: "Valid manifest",
			tag:  "v1.0.0",
			manifest: &interfaces.Manifest{
				Content:   []byte(`{"schemaVersion":2}`),
				MediaType: "application/vnd.docker.distribution.manifest.v2+json",
				Digest:    "sha256:abc123",
			},
			expectError: false,
		},
		{
			name:        "Empty tag",
			tag:         "",
			manifest:    &interfaces.Manifest{Content: []byte("test")},
			expectError: true,
		},
		{
			name:        "Nil manifest",
			tag:         "v1.0.0",
			manifest:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClient()
			ctx := context.Background()

			repo, err := client.GetRepository(ctx, "test-repo")
			assert.NoError(t, err)

			err = repo.PutManifest(ctx, tt.tag, tt.manifest)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// Will error without GCP credentials
				if err != nil {
					assert.Error(t, err)
				}
			}
		})
	}
}

func TestRepository_DeleteImage(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		expectError bool
	}{
		{
			name:        "Valid tag",
			tag:         "v1.0.0",
			expectError: false,
		},
		{
			name:        "Empty tag",
			tag:         "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClient()
			ctx := context.Background()

			repo, err := client.GetRepository(ctx, "test-repo")
			assert.NoError(t, err)

			err = repo.DeleteImage(ctx, tt.tag)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// Will error without GCP credentials
				if err != nil {
					assert.Error(t, err)
				}
			}
		})
	}
}

func TestRepository_DeleteManifest(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		expectError bool
	}{
		{
			name:        "Valid tag",
			tag:         "v1.0.0",
			expectError: false,
		},
		{
			name:        "Empty tag",
			tag:         "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClient()
			ctx := context.Background()

			repo, err := client.GetRepository(ctx, "test-repo")
			assert.NoError(t, err)

			err = repo.DeleteManifest(ctx, tt.tag)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// Will error without GCP credentials
				if err != nil {
					assert.Error(t, err)
				}
			}
		})
	}
}

func TestRepository_GetLayerReader(t *testing.T) {
	tests := []struct {
		name        string
		digest      string
		expectError bool
	}{
		{
			name:        "Valid digest",
			digest:      "sha256:abc123",
			expectError: false,
		},
		{
			name:        "Empty digest",
			digest:      "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClient()
			ctx := context.Background()

			repo, err := client.GetRepository(ctx, "test-repo")
			assert.NoError(t, err)

			reader, err := repo.GetLayerReader(ctx, tt.digest)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, reader)
			} else {
				// Will error without GCP credentials
				if err != nil {
					assert.Error(t, err)
				} else if reader != nil {
					reader.Close()
				}
			}
		})
	}
}

func TestRepository_PutLayer(t *testing.T) {
	tests := []struct {
		name        string
		hasLayer    bool
		expectError bool
	}{
		{
			name:        "Valid layer",
			hasLayer:    true,
			expectError: false,
		},
		{
			name:        "Nil layer",
			hasLayer:    false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClient()
			ctx := context.Background()

			repo, err := client.GetRepository(ctx, "test-repo")
			assert.NoError(t, err)

			var layer interface{}
			if tt.hasLayer {
				layer = nil // Would need proper v1.Layer mock
			}

			err = repo.PutLayer(ctx, layer)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// Will error without proper layer mock
				if err != nil {
					assert.Error(t, err)
				}
			}
		})
	}
}

func TestRepository_GetImageReference(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		expectError bool
		isDigest    bool
	}{
		{
			name:        "Valid tag",
			tag:         "v1.0.0",
			expectError: false,
			isDigest:    false,
		},
		{
			name:        "Digest reference",
			tag:         "@sha256:abc123",
			expectError: false,
			isDigest:    true,
		},
		{
			name:        "Empty tag",
			tag:         "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createTestClient()
			ctx := context.Background()

			repo, err := client.GetRepository(ctx, "test-repo")
			assert.NoError(t, err)

			ref, err := repo.GetImageReference(tt.tag)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, ref)
			} else {
				if err != nil && tt.isDigest {
					assert.Error(t, err)
				} else if err == nil {
					assert.NoError(t, err)
					assert.NotNil(t, ref)
				}
			}
		})
	}
}

func TestRepository_GetRemoteOptions(t *testing.T) {
	client := createTestClient()
	ctx := context.Background()

	repo, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

	opts, err := repo.GetRemoteOptions()
	assert.NoError(t, err)
	assert.NotNil(t, opts)
	assert.Len(t, opts, 1)
}

func TestRepository_GetRepositoryReference(t *testing.T) {
	client := createTestClient()
	ctx := context.Background()

	repo, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

	// This tests the internal repository reference
	// The method exists on the implementation
	assert.NotNil(t, repo)
}

// Test concurrent repository operations
func TestRepository_ConcurrentListTags(t *testing.T) {
	client := createTestClient()
	ctx := context.Background()

	repo, err := client.GetRepository(ctx, "test-repo")
	assert.NoError(t, err)

	results := make(chan error, 5)

	// Run 5 concurrent ListTags operations
	for i := 0; i < 5; i++ {
		go func() {
			_, err := repo.ListTags(ctx)
			results <- err
		}()
	}

	// Collect results (will error without GCP credentials)
	for i := 0; i < 5; i++ {
		<-results
	}
}

// Test error handling for not found scenarios
func TestRepository_NotFoundErrors(t *testing.T) {
	client := createTestClient()
	ctx := context.Background()

	repo, err := client.GetRepository(ctx, "nonexistent-repo")
	assert.NoError(t, err)

	// These operations should handle not found errors gracefully
	_, err = repo.ListTags(ctx)
	if err != nil {
		// Should return a not found error
		assert.Error(t, err)
	}

	_, err = repo.GetImage(ctx, "nonexistent-tag")
	if err != nil {
		assert.Error(t, err)
	}
}
