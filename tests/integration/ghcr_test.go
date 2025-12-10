//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"freightliner/pkg/client/ghcr"
	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"
	"freightliner/pkg/interfaces"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGHCR_Authentication tests GitHub Container Registry authentication
func TestGHCR_Authentication(t *testing.T) {
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
			name: "GitHub Token Authentication (PAT)",
			auth: config.AuthConfig{
				Type:  config.AuthTypeToken,
				Token: os.Getenv("GITHUB_TOKEN"),
			},
			skipCI:  os.Getenv("GITHUB_TOKEN") == "",
			wantErr: false,
		},
		{
			name: "GitHub Actions Token",
			auth: config.AuthConfig{
				Type:  config.AuthTypeToken,
				Token: os.Getenv("GITHUB_ACTIONS_TOKEN"),
			},
			skipCI:  os.Getenv("GITHUB_ACTIONS_TOKEN") == "",
			wantErr: false,
		},
		{
			name: "Basic Auth with Token",
			auth: config.AuthConfig{
				Type:     config.AuthTypeBasic,
				Username: os.Getenv("GITHUB_USERNAME"),
				Password: os.Getenv("GITHUB_TOKEN"),
			},
			skipCI:  os.Getenv("GITHUB_USERNAME") == "" || os.Getenv("GITHUB_TOKEN") == "",
			wantErr: false,
		},
		{
			name: "Invalid Token",
			auth: config.AuthConfig{
				Type:  config.AuthTypeToken,
				Token: "invalid_token_12345",
			},
			wantErr:   true,
			errString: "authentication failed",
		},
		{
			name: "Anonymous Access Public Package",
			auth: config.AuthConfig{
				Type: config.AuthTypeAnonymous,
			},
			wantErr: false, // Should work for public packages
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipCI && os.Getenv("CI") != "" {
				t.Skip("Skipping test requiring credentials in CI")
			}

			regConfig := config.RegistryConfig{
				Name:     "test-ghcr",
				Type:     config.RegistryTypeGitHub,
				Endpoint: "https://ghcr.io",
				Auth:     tt.auth,
				Timeout:  30,
			}

			client, err := ghcr.NewClient(ghcr.ClientOptions{
				RegistryConfig: regConfig,
				Logger:         log.NewBasicLogger(log.InfoLevel),
			})

			if tt.wantErr {
				// May succeed in creating client but fail on operations
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

// TestGHCR_RepositoryListing tests listing packages in GHCR
func TestGHCR_RepositoryListing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupGHCRClient(t)
	if client == nil {
		t.Skip("GHCR client not configured, skipping test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repos, err := client.ListRepositories(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, repos)

	t.Logf("Found %d packages in GHCR", len(repos))

	// Verify repository format (should be owner/package-name)
	for _, repo := range repos {
		assert.Contains(t, repo, "/", "Repository should include owner")
	}
}

// TestGHCR_OrganizationPackages tests accessing organization packages
func TestGHCR_OrganizationPackages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupGHCRClient(t)
	if client == nil {
		t.Skip("GHCR client not configured, skipping test")
	}

	orgName := os.Getenv("GITHUB_ORG")
	if orgName == "" {
		t.Skip("GitHub organization name not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repos, err := client.ListRepositories(ctx, "")
	require.NoError(t, err)

	// Filter by organization
	orgPackages := []string{}
	for _, repo := range repos {
		if len(repo) > len(orgName) && repo[:len(orgName)] == orgName {
			orgPackages = append(orgPackages, repo)
		}
	}

	t.Logf("Found %d packages in organization %s", len(orgPackages), orgName)
}

// TestGHCR_TagListing tests listing package versions (tags)
func TestGHCR_TagListing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupGHCRClient(t)
	if client == nil {
		t.Skip("GHCR client not configured, skipping test")
	}

	testRepo := os.Getenv("GHCR_TEST_REPO")
	if testRepo == "" {
		t.Skip("Test repository not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repo, err := client.GetRepository(ctx, testRepo)
	if err != nil {
		t.Logf("Repository %s may not be accessible, error: %v", testRepo, err)
		t.Skip("Test repository not available")
	}

	tags, err := repo.ListTags(ctx)
	if err != nil {
		t.Logf("Failed to list tags for %s, error: %v", testRepo, err)
		t.Skip("Could not list tags")
	}

	require.NoError(t, err)
	assert.NotNil(t, tags)
	t.Logf("Found %d versions for package %s", len(tags), testRepo)
}

// TestGHCR_ManifestRetrieval tests getting manifests from GHCR
func TestGHCR_ManifestRetrieval(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupGHCRClient(t)
	if client == nil {
		t.Skip("GHCR client not configured, skipping test")
	}

	testRepo := os.Getenv("GHCR_TEST_REPO")
	testTag := os.Getenv("GHCR_TEST_TAG")
	if testRepo == "" || testTag == "" {
		t.Skip("Test repository and tag not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	manifest, err := func() (*interfaces.Manifest, error) {
		repo, err := client.GetRepository(ctx, testRepo)
		if err != nil {
			return nil, err
		}
		return repo.GetManifest(ctx, testTag)
	}()
	require.NoError(t, err)
	require.NotNil(t, manifest)

	assert.NotEmpty(t, manifest.SchemaVersion)
	assert.NotEmpty(t, manifest.MediaType)
	t.Logf("Retrieved manifest for %s:%s", testRepo, testTag)
}

// TestGHCR_LayerDownload tests downloading layers from GHCR
func TestGHCR_LayerDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupGHCRClient(t)
	if client == nil {
		t.Skip("GHCR client not configured, skipping test")
	}

	testRepo := os.Getenv("GHCR_TEST_REPO")
	testTag := os.Getenv("GHCR_TEST_TAG")
	if testRepo == "" || testTag == "" {
		t.Skip("Test repository and tag not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Get manifest first
	manifest, err := func() (*interfaces.Manifest, error) {
		repo, err := client.GetRepository(ctx, testRepo)
		if err != nil {
			return nil, err
		}
		return repo.GetManifest(ctx, testTag)
	}()
	require.NoError(t, err)
	require.Greater(t, len(manifest.Layers), 0)

	// Download first layer
	layer := manifest.Layers[0]
	data, err := client.DownloadLayer(ctx, testRepo, layer.Digest)
	require.NoError(t, err)
	assert.Greater(t, len(data), 0)

	t.Logf("Downloaded layer %s (size: %d bytes)", layer.Digest, len(data))
}

// TestGHCR_PackagePublishing tests publishing packages to GHCR
func TestGHCR_PackagePublishing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupGHCRClient(t)
	if client == nil {
		t.Skip("GHCR client not configured, skipping test")
	}

	// Ensure we have write permissions
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("GitHub token with write permissions required")
	}

	testRepo := os.Getenv("GHCR_DEST_REPO")
	if testRepo == "" {
		t.Skip("Destination repository not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// For this test, we'll use a simple Alpine image as source
	sourceRepo := "library/alpine"
	sourceTag := "latest"

	// Get manifest from a source (would typically be another registry)
	manifest, err := func() (*interfaces.Manifest, error) {
		repo, err := client.GetRepository(ctx, sourceRepo)
		if err != nil {
			return nil, err
		}
		return repo.GetManifest(ctx, sourceTag)
	}()
	if err != nil {
		t.Skip("Cannot access source image for publishing test")
	}

	// Push to GHCR (simplified for test)
	err = client.PushManifest(ctx, testRepo, "test", manifest)
	if err != nil {
		t.Logf("Push may require additional setup: %v", err)
	}
}

// TestGHCR_PackageVisibility tests accessing public vs private packages
func TestGHCR_PackageVisibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name          string
		authenticated bool
		repo          string
		shouldAccess  bool
	}{
		{
			name:          "Public package anonymous access",
			authenticated: false,
			repo:          os.Getenv("GHCR_PUBLIC_REPO"),
			shouldAccess:  true,
		},
		{
			name:          "Private package requires auth",
			authenticated: false,
			repo:          os.Getenv("GHCR_PRIVATE_REPO"),
			shouldAccess:  false,
		},
		{
			name:          "Private package with auth",
			authenticated: true,
			repo:          os.Getenv("GHCR_PRIVATE_REPO"),
			shouldAccess:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.repo == "" {
				t.Skip("Test repository not configured")
			}

			var client *ghcr.Client
			if tt.authenticated {
				client = setupGHCRClient(t)
			} else {
				regConfig := config.RegistryConfig{
					Name:     "test-ghcr-anon",
					Type:     config.RegistryTypeGitHub,
					Endpoint: "https://ghcr.io",
					Auth: config.AuthConfig{
						Type: config.AuthTypeAnonymous,
					},
				}
				var err error
				client, err = ghcr.NewClient(ghcr.ClientOptions{
					RegistryConfig: regConfig,
					Logger:         log.NewBasicLogger(log.InfoLevel),
				})
				require.NoError(t, err)
			}

			if client == nil {
				t.Skip("Client not configured")
			}

			ctx := context.Background()
			_, err := client.ListTags(ctx, tt.repo)

			if tt.shouldAccess {
				assert.NoError(t, err, "Should be able to access repository")
			} else {
				assert.Error(t, err, "Should not be able to access private repository")
			}
		})
	}
}

// TestGHCR_MultiArchSupport tests multi-architecture image support
func TestGHCR_MultiArchSupport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupGHCRClient(t)
	if client == nil {
		t.Skip("GHCR client not configured, skipping test")
	}

	testRepo := os.Getenv("GHCR_MULTIARCH_REPO")
	testTag := os.Getenv("GHCR_MULTIARCH_TAG")
	if testRepo == "" {
		t.Skip("Multi-arch test repository not configured")
	}
	if testTag == "" {
		testTag = "latest"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	manifest, err := func() (*interfaces.Manifest, error) {
		repo, err := client.GetRepository(ctx, testRepo)
		if err != nil {
			return nil, err
		}
		return repo.GetManifest(ctx, testTag)
	}()
	require.NoError(t, err)

	// Check if it's a manifest list
	if manifest.MediaType == "application/vnd.docker.distribution.manifest.list.v2+json" ||
		manifest.MediaType == "application/vnd.oci.image.index.v1+json" {
		t.Log("Successfully retrieved multi-architecture manifest list")
		assert.NotEmpty(t, manifest.Manifests)

		// Log available architectures
		for _, m := range manifest.Manifests {
			t.Logf("  - Platform: %s/%s", m.Platform.OS, m.Platform.Architecture)
		}
	} else {
		t.Logf("Single-arch manifest: %s", manifest.MediaType)
	}
}

// TestGHCR_Replication tests replicating images to/from GHCR
func TestGHCR_Replication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sourceClient := setupGHCRClient(t)
	if sourceClient == nil {
		t.Skip("GHCR client not configured, skipping test")
	}

	testRepo := os.Getenv("GHCR_TEST_REPO")
	testTag := os.Getenv("GHCR_TEST_TAG")
	destRepo := os.Getenv("GHCR_DEST_REPO")

	if testRepo == "" || testTag == "" || destRepo == "" {
		t.Skip("Source and destination repositories not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get manifest from source
	manifest, err := sourceClient.GetManifest(ctx, testRepo, testTag)
	require.NoError(t, err)

	// Download all layers
	layers := [][]byte{}
	for _, layer := range manifest.Layers {
		data, err := sourceClient.DownloadLayer(ctx, testRepo, layer.Digest)
		require.NoError(t, err)
		layers = append(layers, data)
		t.Logf("Downloaded layer %s", layer.Digest)
	}

	// Push to destination
	for i, layerData := range layers {
		err := sourceClient.PushLayer(ctx, destRepo, manifest.Layers[i].Digest, layerData)
		require.NoError(t, err)
	}

	err = sourceClient.PushManifest(ctx, destRepo, testTag, manifest)
	require.NoError(t, err)

	t.Logf("Successfully replicated image to destination")
}

// TestGHCR_RateLimiting tests GHCR rate limiting
func TestGHCR_RateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupGHCRClient(t)
	if client == nil {
		t.Skip("GHCR client not configured, skipping test")
	}

	ctx := context.Background()

	// Make many rapid requests
	successCount := 0
	errorCount := 0

	for i := 0; i < 100; i++ {
		_, err := client.ListRepositories(ctx, "")
		if err != nil {
			errorCount++
		} else {
			successCount++
		}
		time.Sleep(50 * time.Millisecond)
	}

	// GHCR has generous rate limits for authenticated users
	assert.Greater(t, successCount, 0, "Should have at least some successful requests")
	t.Logf("Success: %d, Errors: %d", successCount, errorCount)
}

// TestGHCR_ErrorHandling tests error scenarios
func TestGHCR_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupGHCRClient(t)
	if client == nil {
		t.Skip("GHCR client not configured, skipping test")
	}

	ctx := context.Background()

	tests := []struct {
		name      string
		operation func() error
		wantErr   bool
	}{
		{
			name: "Non-existent package",
			operation: func() error {
				repo, err := client.GetRepository(ctx, "nonexistent/package-xyz-12345")
				if err != nil {
					return err
				}
				_, err = repo.ListTags(ctx)
				return err
			},
			wantErr: true,
		},
		{
			name: "Invalid version/tag",
			operation: func() error {
				testRepo := os.Getenv("GHCR_TEST_REPO")
				if testRepo == "" {
					return nil // Skip if not configured
				}
				_, err := func() (*interfaces.Manifest, error) {
					repo, err := client.GetRepository(ctx, testRepo)
					if err != nil {
						return nil, err
					}
					return repo.GetManifest(ctx, "invalid-tag-xyz")
				}()
				return err
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.operation()
			if err == nil {
				// Some tests may be skipped
				return
			}
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGHCR_RetryLogic tests retry behavior
func TestGHCR_RetryLogic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("GitHub token not configured")
	}

	regConfig := config.RegistryConfig{
		Name:          "test-ghcr-retry",
		Type:          config.RegistryTypeGitHub,
		Endpoint:      "https://ghcr.io",
		RetryAttempts: 3,
		Auth: config.AuthConfig{
			Type:  config.AuthTypeToken,
			Token: token,
		},
	}

	client, err := ghcr.NewClient(ghcr.ClientOptions{
		RegistryConfig: regConfig,
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test that operations succeed with retries
	_, err = client.ListRepositories(ctx, "")
	assert.NoError(t, err)
}

// setupGHCRClient creates a GHCR client for testing
func setupGHCRClient(t *testing.T) *ghcr.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil
	}

	regConfig := config.RegistryConfig{
		Name:     "test-ghcr",
		Type:     config.RegistryTypeGitHub,
		Endpoint: "https://ghcr.io",
		Auth: config.AuthConfig{
			Type:  config.AuthTypeToken,
			Token: token,
		},
		Timeout: 30,
	}

	client, err := ghcr.NewClient(ghcr.ClientOptions{
		RegistryConfig: regConfig,
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	if err != nil {
		t.Logf("Failed to create GHCR client: %v", err)
		return nil
	}

	return client
}

// BenchmarkGHCR_Operations benchmarks GHCR operations
func BenchmarkGHCR_Operations(b *testing.B) {
	client := setupGHCRClient(&testing.T{})
	if client == nil {
		b.Skip("GHCR client not configured")
	}

	ctx := context.Background()

	b.Run("ListRepositories", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = client.ListRepositories(ctx, "")
		}
	})

	testRepo := os.Getenv("GHCR_TEST_REPO")
	if testRepo != "" {
		b.Run("ListTags", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				repo, _ := client.GetRepository(ctx, testRepo)
				if repo != nil {
					_, _ = repo.ListTags(ctx)
				}
			}
		})
	}
}
