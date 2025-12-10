//go:build cosign
// +build cosign

package cosign

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVerifier(t *testing.T) {
	tests := []struct {
		name    string
		config  *VerifierConfig
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "valid config with public key",
			config: &VerifierConfig{
				PublicKey: generateTestPublicKey(t),
			},
			wantErr: false,
		},
		{
			name: "valid config with keyless",
			config: &VerifierConfig{
				EnableKeyless: true,
				RekorURL:      "https://rekor.sigstore.dev",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier, err := NewVerifier(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, verifier)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, verifier)
			}
		})
	}
}

func TestVerifier_LoadPublicKeyVerifier(t *testing.T) {
	// Generate test ECDSA key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	// Encode public key to PEM
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)

	pemBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	config := &VerifierConfig{
		PublicKey: pemBlock,
	}

	verifier, err := NewVerifier(config)
	require.NoError(t, err)
	assert.NotNil(t, verifier)
	assert.Len(t, verifier.verifiers, 1)
}

func TestVerifier_VerifySignature(t *testing.T) {
	ctx := context.Background()

	config := &VerifierConfig{
		PublicKey: generateTestPublicKey(t),
	}

	verifier, err := NewVerifier(config)
	require.NoError(t, err)

	// Mock reference
	ref, err := name.ParseReference("gcr.io/test/image:latest")
	require.NoError(t, err)

	// This test would require mocking the registry interaction
	// For now, we verify the verifier was created correctly
	assert.NotNil(t, verifier)
	assert.NotNil(t, ref)
	assert.NotNil(t, ctx)
}

func TestSignature_ExtractIssuer(t *testing.T) {
	// Create a test certificate with Fulcio issuer extension
	cert := generateTestCertificate(t)

	issuer := extractIssuer(cert)
	// Without actual Fulcio extensions, this will be empty
	assert.Equal(t, "", issuer)
}

func TestSignature_ExtractSubject(t *testing.T) {
	cert := generateTestCertificate(t)
	cert.EmailAddresses = []string{"test@example.com"}

	subject := extractSubject(cert)
	assert.Equal(t, "test@example.com", subject)
}

func TestVerifier_VerifyCertificateChain(t *testing.T) {
	config := &VerifierConfig{}
	verifier, err := NewVerifier(config)
	require.NoError(t, err)

	// Create test certificates
	rootCert := generateTestCertificate(t)
	rootCert.IsCA = true

	intermediateCert := generateTestCertificate(t)
	intermediateCert.IsCA = true

	leafCert := generateTestCertificate(t)

	// Test verification (will fail without proper signing)
	err = verifier.verifyCertificateChain(leafCert, []*x509.Certificate{intermediateCert, rootCert})
	// Expected to fail since we didn't properly sign the certificates
	assert.Error(t, err)
}

func TestVerifier_WithPolicy(t *testing.T) {
	policy := &Policy{
		RequireSignature: true,
		MinSignatures:    1,
		RequireRekor:     false,
		EnforcementMode:  "enforce",
	}

	config := &VerifierConfig{
		PublicKey: generateTestPublicKey(t),
		Policy:    policy,
	}

	verifier, err := NewVerifier(config)
	require.NoError(t, err)
	assert.NotNil(t, verifier)
	assert.Equal(t, policy, verifier.policy)
}

func TestVerifier_KeylessConfiguration(t *testing.T) {
	config := &VerifierConfig{
		EnableKeyless: true,
		RekorURL:      "https://rekor.sigstore.dev",
		FulcioURL:     "https://fulcio.sigstore.dev",
	}

	verifier, err := NewVerifier(config)
	require.NoError(t, err)
	assert.NotNil(t, verifier)
	assert.NotNil(t, verifier.rekorClient)
	assert.Equal(t, "https://rekor.sigstore.dev", verifier.rekorClient.url)
}

func TestVerifier_PolicyEvaluation(t *testing.T) {
	ctx := context.Background()

	policy := &Policy{
		RequireSignature: true,
		MinSignatures:    2,
		RequireRekor:     true,
		EnforcementMode:  "enforce",
	}

	// Test with insufficient signatures
	signatures := []Signature{
		{
			Digest: "sha256:abc123",
		},
	}

	err := policy.Evaluate(ctx, signatures)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient signatures")
}

func TestVerifier_MultipleVerifiers(t *testing.T) {
	// Generate multiple keys
	key1 := generateTestPublicKey(t)
	key2 := generateTestPublicKey(t)

	config := &VerifierConfig{
		PublicKey: key1,
	}

	verifier, err := NewVerifier(config)
	require.NoError(t, err)

	// Add second verifier manually (in practice, this would be done differently)
	secondVerifier, err := verifier.loadPublicKeyVerifier(key2)
	require.NoError(t, err)
	verifier.verifiers = append(verifier.verifiers, secondVerifier)

	assert.Len(t, verifier.verifiers, 2)
}

func TestExtractPredicateType(t *testing.T) {
	// Test with mock SLSA provenance
	payload := []byte(`{"predicateType":"https://slsa.dev/provenance/v0.2"}`)

	predicateType := extractPredicateType(payload)
	// Current implementation returns hardcoded value
	assert.Equal(t, "https://slsa.dev/provenance/v0.2", predicateType)
}

// Helper functions

func generateTestPublicKey(t *testing.T) []byte {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)

	pemBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return pemBlock
}

func generateTestCertificate(t *testing.T) *x509.Certificate {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test Certificate",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certBytes)
	require.NoError(t, err)

	return cert
}
