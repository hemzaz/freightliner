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

// TestGeneric_RegistryCompatibility tests generic client with various registry types
func TestGeneric_RegistryCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	registries := []struct {
		name     string
		endpoint string
		authType config.AuthType
		skipCI   bool
	}{
		{
			name:     "Docker Hub via Generic",
			endpoint: "https://registry-1.docker.io",
			authType: config.AuthTypeAnonymous,
			skipCI:   false,
		},
		{
			name:     "Quay.io via Generic",
			endpoint: "https://quay.io",
			authType: config.AuthTypeAnonymous,
			skipCI:   false,
		},
		{
			name:     "GHCR via Generic",
			endpoint: "https://ghcr.io",
			authType: config.AuthTypeAnonymous,
			skipCI:   false,
		},
		{
			name:     "GitLab Registry via Generic",
			endpoint: "https://registry.gitlab.com",
			authType: config.AuthTypeAnonymous,
			skipCI:   false,
		},
	}

	for _, reg := range registries {
		t.Run(reg.name, func(t *testing.T) {
			if reg.skipCI && os.Getenv("CI") != "" {
				t.Skip("Skipping test in CI")
			}

			regConfig := config.RegistryConfig{
				Name:     reg.name,
				Type:     config.RegistryTypeGeneric,
				Endpoint: reg.endpoint,
				Auth: config.AuthConfig{
					Type: reg.authType,
				},
				Timeout: 30,
			}

			client, err := generic.NewClient(generic.ClientOptions{
				RegistryConfig: regConfig,
				RegistryName:   reg.name,
				Logger:         log.NewBasicLogger(log.InfoLevel),
			})

			require.NoError(t, err)
			require.NotNil(t, client)

			// Verify basic operations work
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Try to list repositories (may not work for all registries without auth)
			_, err = client.ListRepositories(ctx)
			if err != nil {
				t.Logf("ListRepositories not available or requires auth for %s: %v", reg.name, err)
			}
		})
	}
}

// TestGeneric_AllAuthTypes tests all authentication types
func TestGeneric_AllAuthTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	endpoint := os.Getenv("GENERIC_REGISTRY_ENDPOINT")
	if endpoint == "" {
		t.Skip("Generic registry endpoint not configured")
	}

	tests := []struct {
		name    string
		auth    config.AuthConfig
		skipCI  bool
		wantErr bool
	}{
		{
			name: "Basic Authentication",
			auth: config.AuthConfig{
				Type:     config.AuthTypeBasic,
				Username: os.Getenv("REGISTRY_USERNAME"),
				Password: os.Getenv("REGISTRY_PASSWORD"),
			},
			skipCI: os.Getenv("REGISTRY_USERNAME") == "",
		},
		{
			name: "Token Authentication",
			auth: config.AuthConfig{
				Type:  config.AuthTypeToken,
				Token: os.Getenv("REGISTRY_TOKEN"),
			},
			skipCI: os.Getenv("REGISTRY_TOKEN") == "",
		},
		{
			name: "Anonymous Access",
			auth: config.AuthConfig{
				Type: config.AuthTypeAnonymous,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipCI && os.Getenv("CI") != "" {
				t.Skip("Skipping test requiring credentials in CI")
			}

			regConfig := config.RegistryConfig{
				Name:     "test-generic",
				Type:     config.RegistryTypeGeneric,
				Endpoint: endpoint,
				Auth:     tt.auth,
				Timeout:  30,
			}

			client, err := generic.NewClient(generic.ClientOptions{
				RegistryConfig: regConfig,
				RegistryName:   "test-generic",
				Logger:         log.NewBasicLogger(log.InfoLevel),
			})

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, client)
		})
	}
}

// TestGeneric_TLSConfiguration tests TLS configuration options
func TestGeneric_TLSConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name     string
		tls      config.TLSConfig
		insecure bool
	}{
		{
			name:     "Secure TLS",
			tls:      config.TLSConfig{},
			insecure: false,
		},
		{
			name:     "Insecure TLS (skip verification)",
			tls:      config.TLSConfig{InsecureSkipVerify: true},
			insecure: true,
		},
		{
			name: "Custom CA Certificate",
			tls: config.TLSConfig{
				CAFile: os.Getenv("REGISTRY_CA_FILE"),
			},
			insecure: false,
		},
		{
			name: "Client Certificate Auth",
			tls: config.TLSConfig{
				CertFile: os.Getenv("REGISTRY_CERT_FILE"),
				KeyFile:  os.Getenv("REGISTRY_KEY_FILE"),
			},
			insecure: false,
		},
	}

	endpoint := os.Getenv("GENERIC_REGISTRY_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://registry-1.docker.io"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regConfig := config.RegistryConfig{
				Name:     "test-generic-tls",
				Type:     config.RegistryTypeGeneric,
				Endpoint: endpoint,
				TLS:      tt.tls,
				Insecure: tt.insecure,
				Auth: config.AuthConfig{
					Type: config.AuthTypeAnonymous,
				},
			}

			client, err := generic.NewClient(generic.ClientOptions{
				RegistryConfig: regConfig,
				RegistryName:   "test-generic-tls",
				Logger:         log.NewBasicLogger(log.InfoLevel),
			})

			require.NoError(t, err)
			require.NotNil(t, client)
		})
	}
}

// TestGeneric_TimeoutConfiguration tests timeout handling
func TestGeneric_TimeoutConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name    string
		timeout int
	}{
		{"Short timeout (5s)", 5},
		{"Medium timeout (30s)", 30},
		{"Long timeout (120s)", 120},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regConfig := config.RegistryConfig{
				Name:     "test-generic-timeout",
				Type:     config.RegistryTypeGeneric,
				Endpoint: "https://registry-1.docker.io",
				Timeout:  tt.timeout,
				Auth: config.AuthConfig{
					Type: config.AuthTypeAnonymous,
				},
			}

			client, err := generic.NewClient(generic.ClientOptions{
				RegistryConfig: regConfig,
				RegistryName:   "test-generic-timeout",
				Logger:         log.NewBasicLogger(log.InfoLevel),
			})

			require.NoError(t, err)
			require.NotNil(t, client)

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(tt.timeout)*time.Second)
			defer cancel()

			// Test operation completes within timeout
			_, err = client.GetManifest(ctx, "library/alpine", "latest")
			if err != nil {
				t.Logf("Operation failed (may be expected for short timeouts): %v", err)
			}
		})
	}
}

// TestGeneric_RetryConfiguration tests retry logic
func TestGeneric_RetryConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name          string
		retryAttempts int
	}{
		{"No retries", 0},
		{"Few retries (3)", 3},
		{"Many retries (10)", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regConfig := config.RegistryConfig{
				Name:          "test-generic-retry",
				Type:          config.RegistryTypeGeneric,
				Endpoint:      "https://registry-1.docker.io",
				RetryAttempts: tt.retryAttempts,
				Auth: config.AuthConfig{
					Type: config.AuthTypeAnonymous,
				},
			}

			client, err := generic.NewClient(generic.ClientOptions{
				RegistryConfig: regConfig,
				RegistryName:   "test-generic-retry",
				Logger:         log.NewBasicLogger(log.InfoLevel),
			})

			require.NoError(t, err)
			require.NotNil(t, client)

			ctx := context.Background()
			_, err = client.GetManifest(ctx, "library/alpine", "latest")
			assert.NoError(t, err)
		})
	}
}

// TestGeneric_ManifestFormats tests handling different manifest formats
func TestGeneric_ManifestFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupGenericClient(t, "https://registry-1.docker.io")
	if client == nil {
		t.Skip("Generic client not configured")
	}

	tests := []struct {
		name string
		repo string
		tag  string
	}{
		{
			name: "Docker Schema 2",
			repo: "library/alpine",
			tag:  "latest",
		},
		{
			name: "OCI Image",
			repo: "library/nginx",
			tag:  "latest",
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := client.GetManifest(ctx, tt.repo, tt.tag)
			require.NoError(t, err)
			require.NotNil(t, manifest)

			assert.NotEmpty(t, manifest.SchemaVersion)
			assert.NotEmpty(t, manifest.MediaType)
			t.Logf("Manifest format: %s", manifest.MediaType)
		})

		time.Sleep(500 * time.Millisecond) // Rate limit protection
	}
}

// TestGeneric_LargeLayerDownload tests downloading large layers
func TestGeneric_LargeLayerDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupGenericClient(t, "https://registry-1.docker.io")
	if client == nil {
		t.Skip("Generic client not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Use a larger image
	manifest, err := client.GetManifest(ctx, "library/ubuntu", "latest")
	require.NoError(t, err)

	if len(manifest.Layers) == 0 {
		t.Skip("No layers found")
	}

	// Download the first layer (usually largest)
	layer := manifest.Layers[0]
	start := time.Now()
	data, err := client.DownloadLayer(ctx, "library/ubuntu", layer.Digest)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.Greater(t, len(data), 0)

	t.Logf("Downloaded layer %s (%d bytes) in %v", layer.Digest, len(data), duration)
}

// TestGeneric_ConcurrentOperations tests concurrent registry operations
func TestGeneric_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupGenericClient(t, "https://registry-1.docker.io")
	if client == nil {
		t.Skip("Generic client not configured")
	}

	ctx := context.Background()

	// Perform multiple concurrent operations
	done := make(chan bool, 10)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			_, err := client.GetManifest(ctx, "library/alpine", "latest")
			if err != nil {
				errors <- err
			}
		}(i)

		time.Sleep(100 * time.Millisecond) // Stagger requests
	}

	// Wait for all operations
	for i := 0; i < 10; i++ {
		<-done
	}
	close(errors)

	errorCount := 0
	for err := range errors {
		t.Logf("Concurrent operation error: %v", err)
		errorCount++
	}

	// Allow some failures due to rate limiting
	assert.LessOrEqual(t, errorCount, 3, "Too many failures in concurrent operations")
}

// TestGeneric_ErrorHandling tests comprehensive error handling
func TestGeneric_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupGenericClient(t, "https://registry-1.docker.io")
	if client == nil {
		t.Skip("Generic client not configured")
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
				_, err := client.GetManifest(ctx, "nonexistent/repo-xyz", "latest")
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
			name: "Malformed digest",
			operation: func() error {
				_, err := client.DownloadLayer(ctx, "library/alpine", "invalid:digest")
				return err
			},
			wantErr: true,
		},
		{
			name: "Valid operation",
			operation: func() error {
				_, err := client.GetManifest(ctx, "library/alpine", "latest")
				return err
			},
			wantErr: false,
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

		time.Sleep(200 * time.Millisecond) // Rate limit protection
	}
}

// setupGenericClient creates a generic registry client for testing
func setupGenericClient(t *testing.T, endpoint string) *generic.Client {
	regConfig := config.RegistryConfig{
		Name:     "test-generic",
		Type:     config.RegistryTypeGeneric,
		Endpoint: endpoint,
		Auth: config.AuthConfig{
			Type: config.AuthTypeAnonymous,
		},
		Timeout:       30,
		RetryAttempts: 3,
	}

	client, err := generic.NewClient(generic.ClientOptions{
		RegistryConfig: regConfig,
		RegistryName:   "test-generic",
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	if err != nil {
		t.Logf("Failed to create generic client: %v", err)
		return nil
	}

	return client
}

// BenchmarkGeneric_Operations benchmarks generic client operations
func BenchmarkGeneric_Operations(b *testing.B) {
	client := setupGenericClient(&testing.T{}, "https://registry-1.docker.io")
	if client == nil {
		b.Skip("Generic client not configured")
	}

	ctx := context.Background()

	b.Run("GetManifest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = client.GetManifest(ctx, "library/alpine", "latest")
			time.Sleep(100 * time.Millisecond) // Rate limit
		}
	})

	b.Run("ListTags", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = client.ListTags(ctx, "library/alpine")
			time.Sleep(100 * time.Millisecond) // Rate limit
		}
	})
}
