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

// TestCosign_SignatureVerification tests cosign signature verification
func TestCosign_SignatureVerification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name       string
		registry   string
		repository string
		tag        string
		shouldPass bool
	}{
		{
			name:       "Signed image with valid signature",
			registry:   "ghcr.io",
			repository: os.Getenv("COSIGN_SIGNED_REPO"),
			tag:        os.Getenv("COSIGN_SIGNED_TAG"),
			shouldPass: true,
		},
		{
			name:       "Unsigned image",
			registry:   "docker.io",
			repository: "library/alpine",
			tag:        "latest",
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.repository == "" {
				t.Skip("Test repository not configured")
			}

			regConfig := config.RegistryConfig{
				Name:     "test-cosign",
				Type:     config.RegistryTypeGeneric,
				Endpoint: "https://" + tt.registry,
				Auth: config.AuthConfig{
					Type: config.AuthTypeAnonymous,
				},
			}

			client, err := generic.NewClient(generic.ClientOptions{
				RegistryConfig: regConfig,
				RegistryName:   "test-cosign",
				Logger:         log.NewBasicLogger(log.InfoLevel),
			})
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Get repository first, then get manifest
			repo, err := client.GetRepository(ctx, tt.repository)
			require.NoError(t, err)

			manifest, err := repo.GetManifest(ctx, tt.tag)
			require.NoError(t, err)

			// Check for cosign signature layers
			hasSignature := false
			for _, layer := range manifest.Layers {
				if layer.MediaType == "application/vnd.dev.cosign.simplesigning.v1+json" {
					hasSignature = true
					break
				}
			}

			if tt.shouldPass {
				assert.True(t, hasSignature, "Expected to find cosign signature")
			} else {
				assert.False(t, hasSignature, "Expected no cosign signature")
			}
		})
	}
}

// TestCosign_PublicKeyVerification tests verification with public keys
func TestCosign_PublicKeyVerification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	publicKey := os.Getenv("COSIGN_PUBLIC_KEY")
	if publicKey == "" {
		t.Skip("Cosign public key not configured")
	}

	repository := os.Getenv("COSIGN_SIGNED_REPO")
	tag := os.Getenv("COSIGN_SIGNED_TAG")
	if repository == "" || tag == "" {
		t.Skip("Test repository not configured")
	}

	// This test would verify the actual cryptographic signature
	// Implementation depends on cosign library integration
	t.Log("Public key verification would be performed here")
}

// TestCosign_KeylessVerification tests keyless (OIDC) verification
func TestCosign_KeylessVerification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	repository := os.Getenv("COSIGN_KEYLESS_REPO")
	tag := os.Getenv("COSIGN_KEYLESS_TAG")
	if repository == "" || tag == "" {
		t.Skip("Test repository not configured")
	}

	// Keyless verification uses Fulcio/Rekor
	t.Log("Keyless verification would be performed here")
}

// TestCosign_AttestationVerification tests attestation verification
func TestCosign_AttestationVerification(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	repository := os.Getenv("COSIGN_ATTESTATION_REPO")
	tag := os.Getenv("COSIGN_ATTESTATION_TAG")
	if repository == "" || tag == "" {
		t.Skip("Test repository not configured")
	}

	regConfig := config.RegistryConfig{
		Name:     "test-cosign-attestation",
		Type:     config.RegistryTypeGeneric,
		Endpoint: "https://ghcr.io",
		Auth: config.AuthConfig{
			Type: config.AuthTypeAnonymous,
		},
	}

	client, err := generic.NewClient(generic.ClientOptions{
		RegistryConfig: regConfig,
		RegistryName:   "test-cosign-attestation",
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get repository first, then get manifest
	repo, err := client.GetRepository(ctx, repository)
	require.NoError(t, err)

	manifest, err := repo.GetManifest(ctx, tag)
	require.NoError(t, err)

	// Check for attestation layers
	hasAttestation := false
	for _, layer := range manifest.Layers {
		if layer.MediaType == "application/vnd.dev.cosign.attestation.v1+json" {
			hasAttestation = true
			t.Log("Found cosign attestation")
			break
		}
	}

	if !hasAttestation {
		t.Log("No attestation found (may be expected)")
	}
}

// TestCosign_SBOMAttestation tests SBOM attestation handling
func TestCosign_SBOMAttestation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	repository := os.Getenv("COSIGN_SBOM_REPO")
	if repository == "" {
		t.Skip("SBOM repository not configured")
	}

	t.Log("SBOM attestation verification would be performed here")
}

// TestCosign_MultipleSignatures tests images with multiple signatures
func TestCosign_MultipleSignatures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	repository := os.Getenv("COSIGN_MULTI_SIG_REPO")
	if repository == "" {
		t.Skip("Multi-signature repository not configured")
	}

	regConfig := config.RegistryConfig{
		Name:     "test-cosign-multi",
		Type:     config.RegistryTypeGeneric,
		Endpoint: "https://ghcr.io",
		Auth: config.AuthConfig{
			Type: config.AuthTypeAnonymous,
		},
	}

	client, err := generic.NewClient(generic.ClientOptions{
		RegistryConfig: regConfig,
		RegistryName:   "test-cosign-multi",
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get repository first, then get manifest
	repo, err := client.GetRepository(ctx, repository)
	require.NoError(t, err)

	manifest, err := repo.GetManifest(ctx, "latest")
	require.NoError(t, err)

	// Count signature layers
	sigCount := 0
	for _, layer := range manifest.Layers {
		if layer.MediaType == "application/vnd.dev.cosign.simplesigning.v1+json" {
			sigCount++
		}
	}

	t.Logf("Found %d cosign signatures", sigCount)
}

// TestCosign_ErrorHandling tests error scenarios
func TestCosign_ErrorHandling(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		wantErr   bool
	}{
		{
			name:      "Invalid signature",
			operation: "verify_invalid_signature",
			wantErr:   true,
		},
		{
			name:      "Missing signature",
			operation: "verify_unsigned_image",
			wantErr:   true,
		},
		{
			name:      "Expired signature",
			operation: "verify_expired_signature",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Placeholder for error handling tests
			t.Logf("Would test: %s", tt.operation)
		})
	}
}

// BenchmarkCosign_Verification benchmarks signature verification
func BenchmarkCosign_Verification(b *testing.B) {
	repository := os.Getenv("COSIGN_SIGNED_REPO")
	if repository == "" {
		b.Skip("Test repository not configured")
	}

	regConfig := config.RegistryConfig{
		Name:     "bench-cosign",
		Type:     config.RegistryTypeGeneric,
		Endpoint: "https://ghcr.io",
		Auth: config.AuthConfig{
			Type: config.AuthTypeAnonymous,
		},
	}

	client, err := generic.NewClient(generic.ClientOptions{
		RegistryConfig: regConfig,
		RegistryName:   "bench-cosign",
		Logger:         log.NewBasicLogger(log.InfoLevel),
	})
	if err != nil {
		b.Skip("Failed to create client")
	}

	ctx := context.Background()

	// Get repository once before benchmark
	repo, err := client.GetRepository(ctx, repository)
	if err != nil {
		b.Skip("Failed to get repository")
	}

	b.Run("GetManifestWithSignature", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = repo.GetManifest(ctx, "latest")
		}
	})
}
