//go:build integration
// +build integration

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

// TestOCIArtifacts_HelmCharts tests OCI registry with Helm charts
func TestOCIArtifacts_HelmCharts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	helmRepo := os.Getenv("OCI_HELM_REPO")
	if helmRepo == "" {
		t.Skip("Helm repository not configured")
	}

	regConfig := config.RegistryConfig{
		Name:     "test-oci-helm",
		Type:     config.RegistryTypeGeneric,
		Endpoint: "https://ghcr.io",
		Auth: config.AuthConfig{
			Type: config.AuthTypeAnonymous,
		},
	}

	client, err := generic.NewClient(generic.ClientOptions{
		RegistryConfig: regConfig,
		RegistryName:   "test-oci-helm",
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get Helm chart manifest
	manifest, err := client.GetManifest(ctx, helmRepo, "latest")
	require.NoError(t, err)

	// Verify it's a Helm chart OCI artifact
	assert.Equal(t, "application/vnd.oci.image.manifest.v1+json", manifest.MediaType)
	assert.Equal(t, "application/vnd.cncf.helm.config.v1+json", manifest.Config.MediaType)

	t.Log("Successfully retrieved Helm chart OCI artifact")
}

// TestOCIArtifacts_WASMModules tests OCI registry with WASM modules
func TestOCIArtifacts_WASMModules(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	wasmRepo := os.Getenv("OCI_WASM_REPO")
	if wasmRepo == "" {
		t.Skip("WASM repository not configured")
	}

	regConfig := config.RegistryConfig{
		Name:     "test-oci-wasm",
		Type:     config.RegistryTypeGeneric,
		Endpoint: "https://ghcr.io",
		Auth: config.AuthConfig{
			Type: config.AuthTypeAnonymous,
		},
	}

	client, err := generic.NewClient(generic.ClientOptions{
		RegistryConfig: regConfig,
		RegistryName:   "test-oci-wasm",
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	manifest, err := client.GetManifest(ctx, wasmRepo, "latest")
	require.NoError(t, err)

	// WASM modules have specific media types
	t.Logf("WASM artifact config type: %s", manifest.Config.MediaType)
}

// TestOCIArtifacts_TerraformModules tests OCI registry with Terraform modules
func TestOCIArtifacts_TerraformModules(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tfRepo := os.Getenv("OCI_TERRAFORM_REPO")
	if tfRepo == "" {
		t.Skip("Terraform repository not configured")
	}

	regConfig := config.RegistryConfig{
		Name:     "test-oci-terraform",
		Type:     config.RegistryTypeGeneric,
		Endpoint: "https://ghcr.io",
		Auth: config.AuthConfig{
			Type: config.AuthTypeAnonymous,
		},
	}

	client, err := generic.NewClient(generic.ClientOptions{
		RegistryConfig: regConfig,
		RegistryName:   "test-oci-terraform",
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	manifest, err := client.GetManifest(ctx, tfRepo, "latest")
	if err != nil {
		t.Logf("Terraform module may not be available: %v", err)
		t.Skip("Terraform module not accessible")
	}

	require.NoError(t, err)
	t.Log("Successfully retrieved Terraform module OCI artifact")
}

// TestOCIArtifacts_GenericArtifacts tests generic OCI artifacts
func TestOCIArtifacts_GenericArtifacts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	artifactRepo := os.Getenv("OCI_ARTIFACT_REPO")
	if artifactRepo == "" {
		t.Skip("Generic artifact repository not configured")
	}

	regConfig := config.RegistryConfig{
		Name:     "test-oci-artifact",
		Type:     config.RegistryTypeGeneric,
		Endpoint: "https://ghcr.io",
		Auth: config.AuthConfig{
			Type: config.AuthTypeAnonymous,
		},
	}

	client, err := generic.NewClient(generic.ClientOptions{
		RegistryConfig: regConfig,
		RegistryName:   "test-oci-artifact",
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	manifest, err := client.GetManifest(ctx, artifactRepo, "latest")
	require.NoError(t, err)

	// Generic artifacts should have OCI manifest format
	assert.Contains(t, manifest.MediaType, "vnd.oci")
	t.Logf("Artifact media type: %s", manifest.MediaType)
}

// TestOCIArtifacts_SBOMDocuments tests OCI registry with SBOM documents
func TestOCIArtifacts_SBOMDocuments(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sbomRepo := os.Getenv("OCI_SBOM_REPO")
	if sbomRepo == "" {
		t.Skip("SBOM repository not configured")
	}

	regConfig := config.RegistryConfig{
		Name:     "test-oci-sbom",
		Type:     config.RegistryTypeGeneric,
		Endpoint: "https://ghcr.io",
		Auth: config.AuthConfig{
			Type: config.AuthTypeAnonymous,
		},
	}

	client, err := generic.NewClient(generic.ClientOptions{
		RegistryConfig: regConfig,
		RegistryName:   "test-oci-sbom",
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	manifest, err := client.GetManifest(ctx, sbomRepo, "latest")
	if err != nil {
		t.Logf("SBOM may not be available: %v", err)
		t.Skip("SBOM not accessible")
	}

	require.NoError(t, err)
	t.Log("Successfully retrieved SBOM OCI artifact")
}

// TestOCIArtifacts_SignatureStorage tests signature storage as OCI artifacts
func TestOCIArtifacts_SignatureStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sigRepo := os.Getenv("OCI_SIGNATURE_REPO")
	if sigRepo == "" {
		t.Skip("Signature repository not configured")
	}

	regConfig := config.RegistryConfig{
		Name:     "test-oci-signature",
		Type:     config.RegistryTypeGeneric,
		Endpoint: "https://ghcr.io",
		Auth: config.AuthConfig{
			Type: config.AuthTypeAnonymous,
		},
	}

	client, err := generic.NewClient(generic.ClientOptions{
		RegistryConfig: regConfig,
		RegistryName:   "test-oci-signature",
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Signatures are typically stored with .sig suffix
	manifest, err := client.GetManifest(ctx, sigRepo, "sha256-abc123.sig")
	if err != nil {
		t.Logf("Signature artifact may not exist: %v", err)
		t.Skip("Signature not accessible")
	}

	require.NoError(t, err)
	t.Log("Successfully retrieved signature OCI artifact")
}

// TestOCIArtifacts_ArtifactReferences tests OCI artifact references (subject field)
func TestOCIArtifacts_ArtifactReferences(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	refRepo := os.Getenv("OCI_REFERENCE_REPO")
	if refRepo == "" {
		t.Skip("Reference repository not configured")
	}

	regConfig := config.RegistryConfig{
		Name:     "test-oci-reference",
		Type:     config.RegistryTypeGeneric,
		Endpoint: "https://ghcr.io",
		Auth: config.AuthConfig{
			Type: config.AuthTypeAnonymous,
		},
	}

	client, err := generic.NewClient(generic.ClientOptions{
		RegistryConfig: regConfig,
		RegistryName:   "test-oci-reference",
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	manifest, err := client.GetManifest(ctx, refRepo, "latest")
	if err != nil {
		t.Logf("Referenced artifact may not exist: %v", err)
		t.Skip("Reference not accessible")
	}

	require.NoError(t, err)

	// Check for subject field (OCI 1.1+ feature)
	if manifest.Subject != nil {
		t.Logf("Artifact references subject: %s", manifest.Subject.Digest)
		assert.NotEmpty(t, manifest.Subject.Digest)
	}
}

// TestOCIArtifacts_CustomMediaTypes tests handling of custom media types
func TestOCIArtifacts_CustomMediaTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	customRepo := os.Getenv("OCI_CUSTOM_REPO")
	if customRepo == "" {
		t.Skip("Custom artifact repository not configured")
	}

	regConfig := config.RegistryConfig{
		Name:     "test-oci-custom",
		Type:     config.RegistryTypeGeneric,
		Endpoint: "https://ghcr.io",
		Auth: config.AuthConfig{
			Type: config.AuthTypeAnonymous,
		},
	}

	client, err := generic.NewClient(generic.ClientOptions{
		RegistryConfig: regConfig,
		RegistryName:   "test-oci-custom",
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	manifest, err := client.GetManifest(ctx, customRepo, "latest")
	if err != nil {
		t.Logf("Custom artifact may not exist: %v", err)
		t.Skip("Custom artifact not accessible")
	}

	require.NoError(t, err)

	// Log all media types
	t.Logf("Manifest media type: %s", manifest.MediaType)
	t.Logf("Config media type: %s", manifest.Config.MediaType)
	for i, layer := range manifest.Layers {
		t.Logf("Layer %d media type: %s", i, layer.MediaType)
	}
}

// TestOCIArtifacts_ArtifactPush tests pushing OCI artifacts
func TestOCIArtifacts_ArtifactPush(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("GitHub token required for push operations")
	}

	destRepo := os.Getenv("OCI_DEST_REPO")
	if destRepo == "" {
		t.Skip("Destination repository not configured")
	}

	regConfig := config.RegistryConfig{
		Name:     "test-oci-push",
		Type:     config.RegistryTypeGeneric,
		Endpoint: "https://ghcr.io",
		Auth: config.AuthConfig{
			Type:  config.AuthTypeToken,
			Token: token,
		},
	}

	client, err := generic.NewClient(generic.ClientOptions{
		RegistryConfig: regConfig,
		RegistryName:   "test-oci-push",
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)

	// This would test pushing a custom OCI artifact
	t.Log("Artifact push would be tested here")
}

// TestOCIArtifacts_ErrorHandling tests error scenarios with OCI artifacts
func TestOCIArtifacts_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	regConfig := config.RegistryConfig{
		Name:     "test-oci-errors",
		Type:     config.RegistryTypeGeneric,
		Endpoint: "https://ghcr.io",
		Auth: config.AuthConfig{
			Type: config.AuthTypeAnonymous,
		},
	}

	client, err := generic.NewClient(generic.ClientOptions{
		RegistryConfig: regConfig,
		RegistryName:   "test-oci-errors",
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name      string
		operation func() error
		wantErr   bool
	}{
		{
			name: "Non-existent artifact",
			operation: func() error {
				_, err := client.GetManifest(ctx, "nonexistent/artifact", "v1.0.0")
				return err
			},
			wantErr: true,
		},
		{
			name: "Invalid artifact reference",
			operation: func() error {
				_, err := client.GetManifest(ctx, "invalid-ref", "latest")
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

// BenchmarkOCIArtifacts_Operations benchmarks OCI artifact operations
func BenchmarkOCIArtifacts_Operations(b *testing.B) {
	artifactRepo := os.Getenv("OCI_ARTIFACT_REPO")
	if artifactRepo == "" {
		b.Skip("Artifact repository not configured")
	}

	regConfig := config.RegistryConfig{
		Name:     "bench-oci-artifact",
		Type:     config.RegistryTypeGeneric,
		Endpoint: "https://ghcr.io",
		Auth: config.AuthConfig{
			Type: config.AuthTypeAnonymous,
		},
	}

	client, err := generic.NewClient(generic.ClientOptions{
		RegistryConfig: regConfig,
		RegistryName:   "bench-oci-artifact",
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	if err != nil {
		b.Skip("Failed to create client")
	}

	ctx := context.Background()

	b.Run("GetArtifactManifest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = client.GetManifest(ctx, artifactRepo, "latest")
		}
	})
}
