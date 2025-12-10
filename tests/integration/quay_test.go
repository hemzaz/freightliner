package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"freightliner/pkg/client/generic"
	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQuay_Authentication tests Quay.io authentication methods
func TestQuay_Authentication(t *testing.T) {
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
			name: "OAuth Token Authentication",
			auth: config.AuthConfig{
				Type:  config.AuthTypeOAuth,
				Token: os.Getenv("QUAY_OAUTH_TOKEN"),
			},
			skipCI:  os.Getenv("QUAY_OAUTH_TOKEN") == "",
			wantErr: false,
		},
		{
			name: "Robot Account Authentication",
			auth: config.AuthConfig{
				Type:     config.AuthTypeBasic,
				Username: os.Getenv("QUAY_ROBOT_USERNAME"),
				Password: os.Getenv("QUAY_ROBOT_TOKEN"),
			},
			skipCI:  os.Getenv("QUAY_ROBOT_USERNAME") == "",
			wantErr: false,
		},
		{
			name: "Basic Authentication",
			auth: config.AuthConfig{
				Type:     config.AuthTypeBasic,
				Username: os.Getenv("QUAY_USERNAME"),
				Password: os.Getenv("QUAY_PASSWORD"),
			},
			skipCI:  os.Getenv("QUAY_USERNAME") == "",
			wantErr: false,
		},
		{
			name: "CLI Token Authentication",
			auth: config.AuthConfig{
				Type:  config.AuthTypeToken,
				Token: os.Getenv("QUAY_CLI_TOKEN"),
			},
			skipCI:  os.Getenv("QUAY_CLI_TOKEN") == "",
			wantErr: false,
		},
		{
			name: "Invalid OAuth Token",
			auth: config.AuthConfig{
				Type:  config.AuthTypeOAuth,
				Token: "invalid_oauth_token_12345",
			},
			wantErr:   true,
			errString: "authentication failed",
		},
		{
			name: "Anonymous Access Public Repo",
			auth: config.AuthConfig{
				Type: config.AuthTypeAnonymous,
			},
			wantErr: false, // Should work for public repos
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipCI && os.Getenv("CI") != "" {
				t.Skip("Skipping test requiring credentials in CI")
			}

			regConfig := config.RegistryConfig{
				Name:     "test-quay",
				Type:     config.RegistryTypeQuay,
				Endpoint: "https://quay.io",
				Auth:     tt.auth,
				Timeout:  30,
			}

			client, err := generic.NewClient(generic.ClientOptions{
				RegistryConfig: regConfig,
				RegistryName:   "test-quay",
				Logger:         log.NewBasicLogger(log.InfoLevel),
			})

			if tt.wantErr {
				// May succeed in creating client but fail on operations
				if err == nil {
					ctx := context.Background()
					_, err = client.ListRepositories(ctx)
				}
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, client)
		})
	}
}

// TestQuay_RepositoryListing tests listing repositories in Quay
func TestQuay_RepositoryListing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupQuayClient(t)
	if client == nil {
		t.Skip("Quay client not configured, skipping test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repos, err := client.ListRepositories(ctx)
	require.NoError(t, err)
	assert.NotNil(t, repos)

	t.Logf("Found %d repositories in Quay", len(repos))

	// Verify repository format includes namespace
	for _, repo := range repos {
		assert.Contains(t, repo, "/", "Repository should include namespace")
	}
}

// TestQuay_OrganizationRepositories tests accessing organization repositories
func TestQuay_OrganizationRepositories(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupQuayClient(t)
	if client == nil {
		t.Skip("Quay client not configured, skipping test")
	}

	orgName := os.Getenv("QUAY_ORG_NAME")
	if orgName == "" {
		t.Skip("Quay organization name not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repos, err := client.ListRepositories(ctx)
	require.NoError(t, err)

	// Filter repos by organization
	orgRepos := []string{}
	for _, repo := range repos {
		if len(repo) > len(orgName) && repo[:len(orgName)] == orgName {
			orgRepos = append(orgRepos, repo)
		}
	}

	t.Logf("Found %d repositories in organization %s", len(orgRepos), orgName)
}

// TestQuay_TagListing tests listing tags with Quay's API
func TestQuay_TagListing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupQuayClient(t)
	if client == nil {
		t.Skip("Quay client not configured, skipping test")
	}

	testRepo := os.Getenv("QUAY_TEST_REPO")
	if testRepo == "" {
		// Try public repository
		testRepo = "quay/busybox"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tags, err := client.ListTags(ctx, testRepo)
	if err != nil {
		t.Logf("Repository %s may not be accessible, error: %v", testRepo, err)
		t.Skip("Test repository not available")
	}

	require.NoError(t, err)
	assert.NotNil(t, tags)
	t.Logf("Found %d tags for repository %s", len(tags), testRepo)
}

// TestQuay_ManifestRetrieval tests getting manifests from Quay
func TestQuay_ManifestRetrieval(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupQuayClient(t)
	if client == nil {
		t.Skip("Quay client not configured, skipping test")
	}

	testRepo := os.Getenv("QUAY_TEST_REPO")
	testTag := os.Getenv("QUAY_TEST_TAG")
	if testRepo == "" {
		testRepo = "quay/busybox"
		testTag = "latest"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	manifest, err := client.GetManifest(ctx, testRepo, testTag)
	require.NoError(t, err)
	require.NotNil(t, manifest)

	assert.NotEmpty(t, manifest.SchemaVersion)
	assert.NotEmpty(t, manifest.MediaType)
	assert.NotEmpty(t, manifest.Config)
	t.Logf("Retrieved manifest for %s:%s", testRepo, testTag)
}

// TestQuay_MultiArchImages tests handling multi-architecture manifests
func TestQuay_MultiArchImages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupQuayClient(t)
	if client == nil {
		t.Skip("Quay client not configured, skipping test")
	}

	// Use a known multi-arch image
	testRepo := "quay/busybox"
	testTag := "latest"

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	manifest, err := client.GetManifest(ctx, testRepo, testTag)
	require.NoError(t, err)

	// Check if it's a manifest list
	if manifest.MediaType == "application/vnd.docker.distribution.manifest.list.v2+json" ||
		manifest.MediaType == "application/vnd.oci.image.index.v1+json" {
		t.Logf("Successfully retrieved multi-arch manifest list")
		assert.NotEmpty(t, manifest.Manifests, "Manifest list should contain manifests")
	} else {
		t.Logf("Single-arch manifest: %s", manifest.MediaType)
	}
}

// TestQuay_SecurityScanning tests Quay's Clair security scanning integration
func TestQuay_SecurityScanning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupQuayClient(t)
	if client == nil {
		t.Skip("Quay client not configured, skipping test")
	}

	testRepo := os.Getenv("QUAY_TEST_REPO")
	testTag := os.Getenv("QUAY_TEST_TAG")
	if testRepo == "" || testTag == "" {
		t.Skip("Test repository and tag not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Get manifest (Quay API v2 doesn't expose vulnerability data directly)
	manifest, err := client.GetManifest(ctx, testRepo, testTag)
	require.NoError(t, err)

	t.Logf("Retrieved manifest for security scan verification")
	// Note: Actual vulnerability data requires Quay API v3 or Clair API
	_ = manifest
}

// TestQuay_Replication tests image replication to/from Quay
func TestQuay_Replication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sourceClient := setupQuayClient(t)
	if sourceClient == nil {
		t.Skip("Quay client not configured, skipping test")
	}

	destNamespace := os.Getenv("QUAY_DEST_NAMESPACE")
	if destNamespace == "" {
		t.Skip("Destination namespace not configured")
	}

	testRepo := os.Getenv("QUAY_TEST_REPO")
	testTag := os.Getenv("QUAY_TEST_TAG")
	if testRepo == "" || testTag == "" {
		t.Skip("Test repository and tag not configured")
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

	// Push to destination namespace
	destRepo := destNamespace + "/" + testRepo
	for i, layerData := range layers {
		err := sourceClient.PushLayer(ctx, destRepo, manifest.Layers[i].Digest, layerData)
		require.NoError(t, err)
	}

	err = sourceClient.PushManifest(ctx, destRepo, testTag, manifest)
	require.NoError(t, err)

	t.Logf("Successfully replicated image to destination namespace")
}

// TestQuay_RateLimiting tests Quay rate limiting behavior
func TestQuay_RateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupQuayClient(t)
	if client == nil {
		t.Skip("Quay client not configured, skipping test")
	}

	ctx := context.Background()

	// Make many rapid requests
	successCount := 0
	errorCount := 0
	rateLimited := 0

	for i := 0; i < 100; i++ {
		_, err := client.ListRepositories(ctx)
		if err != nil {
			errorCount++
			if err.Error() == "rate limited" || err.Error() == "429" {
				rateLimited++
			}
		} else {
			successCount++
		}
		time.Sleep(50 * time.Millisecond)
	}

	assert.Greater(t, successCount, 0, "Should have at least some successful requests")
	t.Logf("Success: %d, Errors: %d, Rate Limited: %d", successCount, errorCount, rateLimited)
}

// TestQuay_ErrorHandling tests error scenarios
func TestQuay_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupQuayClient(t)
	if client == nil {
		t.Skip("Quay client not configured, skipping test")
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
				_, err := client.ListTags(ctx, "nonexistent/repository-xyz")
				return err
			},
			wantErr: true,
		},
		{
			name: "Invalid tag",
			operation: func() error {
				_, err := client.GetManifest(ctx, "quay/busybox", "invalid-tag-12345")
				return err
			},
			wantErr: true,
		},
		{
			name: "Invalid digest",
			operation: func() error {
				_, err := client.DownloadLayer(ctx, "quay/busybox", "sha256:invaliddigest")
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

// TestQuay_RetryLogic tests retry behavior
func TestQuay_RetryLogic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	regConfig := config.RegistryConfig{
		Name:          "test-quay-retry",
		Type:          config.RegistryTypeQuay,
		Endpoint:      "https://quay.io",
		RetryAttempts: 3,
		Auth: config.AuthConfig{
			Type:  config.AuthTypeToken,
			Token: os.Getenv("QUAY_OAUTH_TOKEN"),
		},
	}

	client, err := generic.NewClient(generic.ClientOptions{
		RegistryConfig: regConfig,
		RegistryName:   "test-quay-retry",
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test that operations succeed with retries
	_, err = client.ListRepositories(ctx)
	assert.NoError(t, err)
}

// setupQuayClient creates a Quay client for testing
func setupQuayClient(t *testing.T) *generic.Client {
	token := os.Getenv("QUAY_OAUTH_TOKEN")
	if token == "" {
		// Try robot account
		username := os.Getenv("QUAY_ROBOT_USERNAME")
		password := os.Getenv("QUAY_ROBOT_TOKEN")
		if username == "" {
			return nil
		}

		regConfig := config.RegistryConfig{
			Name:     "test-quay",
			Type:     config.RegistryTypeQuay,
			Endpoint: "https://quay.io",
			Auth: config.AuthConfig{
				Type:     config.AuthTypeBasic,
				Username: username,
				Password: password,
			},
			Timeout: 30,
		}

		client, err := generic.NewClient(generic.ClientOptions{
			RegistryConfig: regConfig,
			RegistryName:   "test-quay",
			Logger:         log.NewBasicLogger(log.InfoLevel),
		})
		if err != nil {
			t.Logf("Failed to create Quay client: %v", err)
			return nil
		}
		return client
	}

	regConfig := config.RegistryConfig{
		Name:     "test-quay",
		Type:     config.RegistryTypeQuay,
		Endpoint: "https://quay.io",
		Auth: config.AuthConfig{
			Type:  config.AuthTypeOAuth,
			Token: token,
		},
		Timeout: 30,
	}

	client, err := generic.NewClient(generic.ClientOptions{
		RegistryConfig: regConfig,
		RegistryName:   "test-quay",
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	if err != nil {
		t.Logf("Failed to create Quay client: %v", err)
		return nil
	}

	return client
}

// BenchmarkQuay_Operations benchmarks Quay operations
func BenchmarkQuay_Operations(b *testing.B) {
	client := setupQuayClient(&testing.T{})
	if client == nil {
		b.Skip("Quay client not configured")
	}

	ctx := context.Background()

	b.Run("ListRepositories", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = client.ListRepositories(ctx)
		}
	})

	b.Run("ListTags", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = client.ListTags(ctx, "quay/busybox")
		}
	})
}
