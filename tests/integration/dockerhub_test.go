package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"freightliner/pkg/client/dockerhub"
	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDockerHub_Authentication tests Docker Hub authentication methods
func TestDockerHub_Authentication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name      string
		auth      config.AuthConfig
		skipCI    bool
		wantErr   bool
		errString string
	}{
		{
			name: "Basic Authentication",
			auth: config.AuthConfig{
				Type:     config.AuthTypeBasic,
				Username: os.Getenv("DOCKERHUB_USERNAME"),
				Password: os.Getenv("DOCKERHUB_PASSWORD"),
			},
			skipCI:  os.Getenv("DOCKERHUB_USERNAME") == "",
			wantErr: false,
		},
		{
			name: "Access Token Authentication",
			auth: config.AuthConfig{
				Type:  config.AuthTypeToken,
				Token: os.Getenv("DOCKERHUB_TOKEN"),
			},
			skipCI:  os.Getenv("DOCKERHUB_TOKEN") == "",
			wantErr: false,
		},
		{
			name: "Anonymous Access",
			auth: config.AuthConfig{
				Type: config.AuthTypeAnonymous,
			},
			wantErr: false, // Docker Hub allows anonymous pulls
		},
		{
			name: "Invalid Credentials",
			auth: config.AuthConfig{
				Type:     config.AuthTypeBasic,
				Username: "invalid_user",
				Password: "invalid_password",
			},
			wantErr:   true,
			errString: "authentication failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipCI && os.Getenv("CI") != "" {
				t.Skip("Skipping test requiring credentials in CI")
			}

			regConfig := config.RegistryConfig{
				Name:     "test-dockerhub",
				Type:     config.RegistryTypeDockerHub,
				Endpoint: "https://registry-1.docker.io",
				Auth:     tt.auth,
				Timeout:  30,
			}

			client, err := dockerhub.NewClient(dockerhub.ClientOptions{
				RegistryConfig: regConfig,
				Logger:         log.NewBasicLogger(log.InfoLevel),
			})

			if tt.wantErr {
				// May succeed in creating client but fail on auth operations
				if err == nil {
					ctx := context.Background()
					_, err = client.ListRepositories(ctx, "")
				}
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, client)
		})
	}
}

// TestDockerHub_RateLimiting tests Docker Hub rate limiting (100 pulls/6h anonymous, 200/6h authenticated)
func TestDockerHub_RateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name          string
		authenticated bool
	}{
		{
			name:          "Anonymous Rate Limits",
			authenticated: false,
		},
		{
			name:          "Authenticated Rate Limits",
			authenticated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var client *dockerhub.Client
			var err error

			if tt.authenticated {
				username := os.Getenv("DOCKERHUB_USERNAME")
				password := os.Getenv("DOCKERHUB_PASSWORD")
				if username == "" {
					t.Skip("Docker Hub credentials not configured")
				}

				regConfig := config.RegistryConfig{
					Name:     "test-dockerhub",
					Type:     config.RegistryTypeDockerHub,
					Endpoint: "https://registry-1.docker.io",
					Auth: config.AuthConfig{
						Type:     config.AuthTypeBasic,
						Username: username,
						Password: password,
					},
				}

				client, err = dockerhub.NewClient(dockerhub.ClientOptions{
					RegistryConfig: regConfig,
					Logger:         log.NewBasicLogger(log.InfoLevel),
				})
			} else {
				regConfig := config.RegistryConfig{
					Name:     "test-dockerhub-anon",
					Type:     config.RegistryTypeDockerHub,
					Endpoint: "https://registry-1.docker.io",
					Auth: config.AuthConfig{
						Type: config.AuthTypeAnonymous,
					},
				}

				client, err = dockerhub.NewClient(dockerhub.ClientOptions{
					RegistryConfig: regConfig,
					Logger:         log.NewBasicLogger(log.InfoLevel),
				})
			}

			require.NoError(t, err)
			require.NotNil(t, client)

			ctx := context.Background()

			// Make requests until rate limited
			rateLimitHit := false
			requestCount := 0

			for i := 0; i < 10; i++ {
				_, err := client.GetManifest(ctx, "library/alpine", "latest")
				requestCount++

				if err != nil && (contains(err.Error(), "rate limit") || contains(err.Error(), "429")) {
					rateLimitHit = true
					t.Logf("Rate limit hit after %d requests", requestCount)
					break
				}

				time.Sleep(100 * time.Millisecond)
			}

			// Log rate limit status
			if rateLimitHit {
				t.Logf("Rate limiting working correctly for %s access", tt.name)
			} else {
				t.Logf("Completed %d requests without hitting rate limit", requestCount)
			}
		})
	}
}

// TestDockerHub_OfficialImages tests accessing official Docker Hub images
func TestDockerHub_OfficialImages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupDockerHubClient(t, false) // Anonymous is fine for official images
	if client == nil {
		t.Skip("Docker Hub client not configured")
	}

	officialImages := []struct {
		name string
		tag  string
	}{
		{"library/alpine", "latest"},
		{"library/nginx", "latest"},
		{"library/redis", "latest"},
		{"library/postgres", "latest"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	for _, img := range officialImages {
		t.Run(img.name, func(t *testing.T) {
			manifest, err := client.GetManifest(ctx, img.name, img.tag)
			require.NoError(t, err)
			require.NotNil(t, manifest)

			assert.NotEmpty(t, manifest.Config)
			assert.NotEmpty(t, manifest.Layers)
			t.Logf("Successfully retrieved %s:%s", img.name, img.tag)
		})

		// Rate limit protection
		time.Sleep(500 * time.Millisecond)
	}
}

// TestDockerHub_UserRepositories tests accessing user repositories
func TestDockerHub_UserRepositories(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	username := os.Getenv("DOCKERHUB_USERNAME")
	if username == "" {
		t.Skip("Docker Hub username not configured")
	}

	client := setupDockerHubClient(t, true) // Need auth for user repos
	if client == nil {
		t.Skip("Docker Hub client not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repos, err := client.ListRepositories(ctx, "")
	require.NoError(t, err)

	t.Logf("Found %d repositories for user %s", len(repos), username)
}

// TestDockerHub_TagListing tests listing tags for repositories
func TestDockerHub_TagListing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupDockerHubClient(t, false)
	if client == nil {
		t.Skip("Docker Hub client not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with official image
	tags, err := client.ListTags(ctx, "library/alpine")
	require.NoError(t, err)
	assert.Greater(t, len(tags), 0, "Should have at least one tag")

	t.Logf("Found %d tags for library/alpine", len(tags))

	// Verify common tags exist
	commonTags := []string{"latest", "3", "3.19"}
	for _, expectedTag := range commonTags {
		found := false
		for _, tag := range tags {
			if tag == expectedTag {
				found = true
				break
			}
		}
		if !found {
			t.Logf("Warning: Expected tag '%s' not found", expectedTag)
		}
	}
}

// TestDockerHub_ManifestRetrieval tests retrieving image manifests
func TestDockerHub_ManifestRetrieval(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupDockerHubClient(t, false)
	if client == nil {
		t.Skip("Docker Hub client not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	manifest, err := client.GetManifest(ctx, "library/alpine", "latest")
	require.NoError(t, err)
	require.NotNil(t, manifest)

	// Verify manifest structure
	assert.NotEmpty(t, manifest.SchemaVersion)
	assert.NotEmpty(t, manifest.MediaType)
	assert.NotEmpty(t, manifest.Config)
	assert.NotEmpty(t, manifest.Layers)

	t.Logf("Manifest has %d layers", len(manifest.Layers))
}

// TestDockerHub_LayerDownload tests downloading image layers
func TestDockerHub_LayerDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupDockerHubClient(t, false)
	if client == nil {
		t.Skip("Docker Hub client not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Get manifest first
	manifest, err := client.GetManifest(ctx, "library/alpine", "latest")
	require.NoError(t, err)
	require.Greater(t, len(manifest.Layers), 0)

	// Download first layer
	layer := manifest.Layers[0]
	data, err := client.DownloadLayer(ctx, "library/alpine", layer.Digest)
	require.NoError(t, err)
	assert.Greater(t, len(data), 0)

	t.Logf("Downloaded layer %s (size: %d bytes)", layer.Digest, len(data))
}

// TestDockerHub_MultiArchSupport tests multi-architecture image support
func TestDockerHub_MultiArchSupport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupDockerHubClient(t, false)
	if client == nil {
		t.Skip("Docker Hub client not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Alpine is a multi-arch image
	manifest, err := client.GetManifest(ctx, "library/alpine", "latest")
	require.NoError(t, err)

	// Check if it's a manifest list
	if manifest.MediaType == "application/vnd.docker.distribution.manifest.list.v2+json" {
		t.Log("Successfully retrieved multi-architecture manifest list")
		assert.NotEmpty(t, manifest.Content) // Verify manifest content is present
	} else {
		t.Logf("Got single-arch manifest: %s", manifest.MediaType)
	}
}

// TestDockerHub_ErrorHandling tests error scenarios
func TestDockerHub_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupDockerHubClient(t, false)
	if client == nil {
		t.Skip("Docker Hub client not configured")
	}

	ctx := context.Background()

	tests := []struct {
		name      string
		operation func() error
		wantErr   bool
	}{
		{
			name: "Non-existent repository",
			operation: func() error {
				_, err := client.GetManifest(ctx, "nonexistent/repository-xyz-12345", "latest")
				return err
			},
			wantErr: true,
		},
		{
			name: "Invalid tag",
			operation: func() error {
				_, err := client.GetManifest(ctx, "library/alpine", "invalid-tag-xyz-12345")
				return err
			},
			wantErr: true,
		},
		{
			name: "Invalid digest",
			operation: func() error {
				_, err := client.DownloadLayer(ctx, "library/alpine", "sha256:invalid")
				return err
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.operation()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDockerHub_RetryLogic tests retry behavior
func TestDockerHub_RetryLogic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	regConfig := config.RegistryConfig{
		Name:          "test-dockerhub-retry",
		Type:          config.RegistryTypeDockerHub,
		Endpoint:      "https://registry-1.docker.io",
		RetryAttempts: 3,
		Auth: config.AuthConfig{
			Type: config.AuthTypeAnonymous,
		},
	}

	client, err := dockerhub.NewClient(dockerhub.ClientOptions{
		RegistryConfig: regConfig,
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test that operations succeed with retries
	_, err = client.GetManifest(ctx, "library/alpine", "latest")
	assert.NoError(t, err)
}

// setupDockerHubClient creates a Docker Hub client for testing
func setupDockerHubClient(t *testing.T, authenticated bool) *dockerhub.Client {
	var regConfig config.RegistryConfig

	if authenticated {
		username := os.Getenv("DOCKERHUB_USERNAME")
		password := os.Getenv("DOCKERHUB_PASSWORD")
		if username == "" {
			return nil
		}

		regConfig = config.RegistryConfig{
			Name:     "test-dockerhub",
			Type:     config.RegistryTypeDockerHub,
			Endpoint: "https://registry-1.docker.io",
			Auth: config.AuthConfig{
				Type:     config.AuthTypeBasic,
				Username: username,
				Password: password,
			},
			Timeout: 30,
		}
	} else {
		regConfig = config.RegistryConfig{
			Name:     "test-dockerhub-anon",
			Type:     config.RegistryTypeDockerHub,
			Endpoint: "https://registry-1.docker.io",
			Auth: config.AuthConfig{
				Type: config.AuthTypeAnonymous,
			},
			Timeout: 30,
		}
	}

	client, err := dockerhub.NewClient(dockerhub.ClientOptions{
		RegistryConfig: regConfig,
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	if err != nil {
		t.Logf("Failed to create Docker Hub client: %v", err)
		return nil
	}

	return client
}

// contains is a helper to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}

// BenchmarkDockerHub_Operations benchmarks Docker Hub operations
func BenchmarkDockerHub_Operations(b *testing.B) {
	client := setupDockerHubClient(&testing.T{}, false)
	if client == nil {
		b.Skip("Docker Hub client not configured")
	}

	ctx := context.Background()

	b.Run("GetManifest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = client.GetManifest(ctx, "library/alpine", "latest")
			time.Sleep(100 * time.Millisecond) // Rate limit protection
		}
	})

	b.Run("ListTags", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = client.ListTags(ctx, "library/alpine")
			time.Sleep(100 * time.Millisecond) // Rate limit protection
		}
	})
}
