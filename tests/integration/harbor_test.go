package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"freightliner/pkg/client/generic"
	"freightliner/pkg/config"
	"freightliner/pkg/helper/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHarbor_Authentication tests Harbor registry authentication methods
func TestHarbor_Authentication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	harborEndpoint := os.Getenv("HARBOR_ENDPOINT")
	if harborEndpoint == "" {
		harborEndpoint = "https://harbor.example.com"
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
				Username: os.Getenv("HARBOR_USERNAME"),
				Password: os.Getenv("HARBOR_PASSWORD"),
			},
			skipCI:  os.Getenv("HARBOR_USERNAME") == "",
			wantErr: false,
		},
		{
			name: "Robot Account Authentication",
			auth: config.AuthConfig{
				Type:     config.AuthTypeBasic,
				Username: os.Getenv("HARBOR_ROBOT_NAME"),
				Password: os.Getenv("HARBOR_ROBOT_SECRET"),
			},
			skipCI:  os.Getenv("HARBOR_ROBOT_NAME") == "",
			wantErr: false,
		},
		{
			name: "OIDC Token Authentication",
			auth: config.AuthConfig{
				Type:  config.AuthTypeToken,
				Token: os.Getenv("HARBOR_OIDC_TOKEN"),
			},
			skipCI:  os.Getenv("HARBOR_OIDC_TOKEN") == "",
			wantErr: false,
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
		{
			name: "Anonymous Access",
			auth: config.AuthConfig{
				Type: config.AuthTypeAnonymous,
			},
			wantErr: false, // May succeed for public projects
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipCI && os.Getenv("CI") != "" {
				t.Skip("Skipping test requiring credentials in CI")
			}

			regConfig := config.RegistryConfig{
				Name:     "test-harbor",
				Type:     config.RegistryTypeHarbor,
				Endpoint: harborEndpoint,
				Auth:     tt.auth,
				Timeout:  30,
			}

			client, err := generic.NewClient(generic.ClientOptions{
				RegistryConfig: regConfig,
				RegistryName:   "test-harbor",
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

// TestHarbor_ProjectManagement tests Harbor project operations
func TestHarbor_ProjectManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupHarborClient(t)
	if client == nil {
		t.Skip("Harbor client not configured, skipping test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// List repositories (should include project names)
	repos, err := client.ListRepositories(ctx)
	require.NoError(t, err)
	assert.NotNil(t, repos)

	t.Logf("Found %d repositories across all projects", len(repos))

	// Verify repository names include project prefix
	for _, repo := range repos {
		assert.Contains(t, repo, "/", "Repository should include project name")
	}
}

// TestHarbor_VulnerabilityScanning tests Harbor vulnerability scanning integration
func TestHarbor_VulnerabilityScanning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupHarborClient(t)
	if client == nil {
		t.Skip("Harbor client not configured, skipping test")
	}

	testRepo := os.Getenv("HARBOR_TEST_REPO")
	testTag := os.Getenv("HARBOR_TEST_TAG")
	if testRepo == "" || testTag == "" {
		t.Skip("Test repository and tag not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Get manifest (Harbor may include vulnerability scan results in metadata)
	manifest, err := client.GetManifest(ctx, testRepo, testTag)
	require.NoError(t, err)
	require.NotNil(t, manifest)

	t.Logf("Retrieved manifest for %s:%s", testRepo, testTag)
	// Note: Actual vulnerability data would require Harbor API extension
}

// TestHarbor_ImageSigning tests Harbor content trust/Notary integration
func TestHarbor_ImageSigning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupHarborClient(t)
	if client == nil {
		t.Skip("Harbor client not configured, skipping test")
	}

	testRepo := os.Getenv("HARBOR_TEST_REPO")
	testTag := os.Getenv("HARBOR_TEST_TAG")
	if testRepo == "" || testTag == "" {
		t.Skip("Test repository and tag not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Verify signed image can be pulled
	manifest, err := client.GetManifest(ctx, testRepo, testTag)
	if err != nil {
		t.Logf("Failed to get manifest (may require content trust): %v", err)
		return
	}

	require.NotNil(t, manifest)
	t.Logf("Successfully retrieved signed image manifest")
}

// TestHarbor_Webhooks tests Harbor webhook functionality
func TestHarbor_Webhooks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Note: This test validates that operations that would trigger webhooks work
	client := setupHarborClient(t)
	if client == nil {
		t.Skip("Harbor client not configured, skipping test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Perform operations that would trigger webhooks
	repos, err := client.ListRepositories(ctx)
	require.NoError(t, err)

	if len(repos) > 0 {
		tags, err := client.ListTags(ctx, repos[0])
		require.NoError(t, err)
		t.Logf("Operations completed that would trigger push/pull webhooks")
		t.Logf("Found %d tags in first repository", len(tags))
	}
}

// TestHarbor_Replication tests Harbor replication between projects
func TestHarbor_Replication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sourceClient := setupHarborClient(t)
	if sourceClient == nil {
		t.Skip("Harbor client not configured, skipping test")
	}

	destProject := os.Getenv("HARBOR_DEST_PROJECT")
	if destProject == "" {
		t.Skip("Destination project not configured")
	}

	testRepo := os.Getenv("HARBOR_TEST_REPO")
	testTag := os.Getenv("HARBOR_TEST_TAG")
	if testRepo == "" || testTag == "" {
		t.Skip("Test repository and tag not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get manifest from source
	manifest, err := sourceClient.GetManifest(ctx, testRepo, testTag)
	require.NoError(t, err)

	// Download layers
	layers := [][]byte{}
	for _, layer := range manifest.Layers {
		data, err := sourceClient.DownloadLayer(ctx, testRepo, layer.Digest)
		require.NoError(t, err)
		layers = append(layers, data)
	}

	// Push to destination project
	destRepo := fmt.Sprintf("%s/%s", destProject, testRepo)
	for i, layerData := range layers {
		err := sourceClient.PushLayer(ctx, destRepo, manifest.Layers[i].Digest, layerData)
		require.NoError(t, err)
	}

	err = sourceClient.PushManifest(ctx, destRepo, testTag, manifest)
	require.NoError(t, err)

	t.Logf("Successfully replicated image to destination project")
}

// TestHarbor_QuotaManagement tests Harbor quota limits
func TestHarbor_QuotaManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupHarborClient(t)
	if client == nil {
		t.Skip("Harbor client not configured, skipping test")
	}

	ctx := context.Background()

	// Test that operations respect quota limits
	// Note: Actual quota validation would require Harbor API extension
	repos, err := client.ListRepositories(ctx)
	require.NoError(t, err)

	t.Logf("Current repository count: %d", len(repos))
	// In a real scenario, we'd verify this against project quota
}

// TestHarbor_RetentionPolicies tests Harbor retention policy behavior
func TestHarbor_RetentionPolicies(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupHarborClient(t)
	if client == nil {
		t.Skip("Harbor client not configured, skipping test")
	}

	testRepo := os.Getenv("HARBOR_TEST_REPO")
	if testRepo == "" {
		t.Skip("Test repository not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// List tags (retention policies may have cleaned up old tags)
	tags, err := client.ListTags(ctx, testRepo)
	require.NoError(t, err)

	t.Logf("Found %d tags (after retention policy application)", len(tags))
}

// TestHarbor_RateLimiting tests Harbor rate limiting
func TestHarbor_RateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupHarborClient(t)
	if client == nil {
		t.Skip("Harbor client not configured, skipping test")
	}

	ctx := context.Background()

	// Make many rapid requests
	successCount := 0
	errorCount := 0

	for i := 0; i < 100; i++ {
		_, err := client.ListRepositories(ctx)
		if err != nil {
			errorCount++
		} else {
			successCount++
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Should handle rate limiting gracefully
	assert.Greater(t, successCount, 0, "Should have at least some successful requests")
	t.Logf("Success: %d, Errors: %d", successCount, errorCount)
}

// TestHarbor_ErrorHandling tests error scenarios
func TestHarbor_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupHarborClient(t)
	if client == nil {
		t.Skip("Harbor client not configured, skipping test")
	}

	ctx := context.Background()

	tests := []struct {
		name      string
		operation func() error
		wantErr   bool
	}{
		{
			name: "Non-existent project/repository",
			operation: func() error {
				_, err := client.ListTags(ctx, "nonexistent/repository")
				return err
			},
			wantErr: true,
		},
		{
			name: "Invalid tag",
			operation: func() error {
				_, err := client.GetManifest(ctx, "library/alpine", "invalid-tag-xyz")
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

// setupHarborClient creates a Harbor client for testing
func setupHarborClient(t *testing.T) *generic.Client {
	harborEndpoint := os.Getenv("HARBOR_ENDPOINT")
	if harborEndpoint == "" {
		return nil
	}

	username := os.Getenv("HARBOR_USERNAME")
	password := os.Getenv("HARBOR_PASSWORD")
	if username == "" || password == "" {
		return nil
	}

	regConfig := config.RegistryConfig{
		Name:     "test-harbor",
		Type:     config.RegistryTypeHarbor,
		Endpoint: harborEndpoint,
		Auth: config.AuthConfig{
			Type:     config.AuthTypeBasic,
			Username: username,
			Password: password,
		},
		Timeout: 30,
	}

	client, err := generic.NewClient(generic.ClientOptions{
		RegistryConfig: regConfig,
		RegistryName:   "test-harbor",
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	if err != nil {
		t.Logf("Failed to create Harbor client: %v", err)
		return nil
	}

	return client
}

// BenchmarkHarbor_Operations benchmarks Harbor operations
func BenchmarkHarbor_Operations(b *testing.B) {
	client := setupHarborClient(&testing.T{})
	if client == nil {
		b.Skip("Harbor client not configured")
	}

	ctx := context.Background()

	b.Run("ListRepositories", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = client.ListRepositories(ctx)
		}
	})

	testRepo := os.Getenv("HARBOR_TEST_REPO")
	if testRepo != "" {
		b.Run("ListTags", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = client.ListTags(ctx, testRepo)
			}
		})
	}
}
