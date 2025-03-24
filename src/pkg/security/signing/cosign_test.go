package signing

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func generateTestECDSAKey(t *testing.T) (string, func()) {
	// Generate a temporary directory for the key
	tempDir, err := os.MkdirTemp("", "cosign-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	
	// Generate a private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to generate key: %v", err)
	}
	
	// Marshal the private key to DER format
	derBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to marshal private key: %v", err)
	}
	
	// Create PEM block
	block := &pem.Block{
		Type:  "ECDSA PRIVATE KEY",
		Bytes: derBytes,
	}
	
	// Write the key to a file
	keyPath := filepath.Join(tempDir, "test-key.pem")
	err = os.WriteFile(keyPath, pem.EncodeToMemory(block), 0600)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to write key file: %v", err)
	}
	
	cleanup := func() {
		os.RemoveAll(tempDir)
	}
	
	return keyPath, cleanup
}

func TestCosignSigner(t *testing.T) {
	// Generate a test key
	keyPath, cleanup := generateTestECDSAKey(t)
	defer cleanup()
	
	// Create a signer
	signer, err := NewCosignSigner(SignOptions{
		KeyID:   "test-key",
		KeyPath: keyPath,
	})
	if err != nil {
		t.Fatalf("NewCosignSigner failed: %v", err)
	}
	
	// Check the signer properties
	if signer.Name() != "cosign" {
		t.Errorf("Expected Name()=cosign, got %s", signer.Name())
	}
	
	// Create a payload to sign
	payload := &SignaturePayload{
		ManifestDigest: "sha256:1234567890abcdef",
		Repository:     "example/image",
		Tag:            "v1.0",
		AdditionalData: map[string]string{
			"creator": "test",
		},
	}
	
	// Sign the payload
	ctx := context.WithValue(context.Background(), "timestamp", int64(1617979891))
	signature, err := signer.Sign(ctx, payload)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}
	
	// Check the signature
	if signature.KeyID != "test-key" {
		t.Errorf("Expected KeyID=test-key, got %s", signature.KeyID)
	}
	
	if len(signature.Signature) == 0 {
		t.Errorf("Expected non-empty signature")
	}
	
	if len(signature.Payload) == 0 {
		t.Errorf("Expected non-empty payload")
	}
	
	// Verify the signature
	valid, err := signer.Verify(ctx, payload, signature)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	
	if !valid {
		t.Errorf("Expected signature to verify as valid")
	}
	
	// Modify the payload and verify it should fail
	modifiedPayload := &SignaturePayload{
		ManifestDigest: "sha256:modified",
		Repository:     payload.Repository,
		Tag:            payload.Tag,
	}
	
	valid, err = signer.Verify(ctx, modifiedPayload, signature)
	if err == nil || valid {
		t.Errorf("Expected signature verification to fail with modified payload")
	}
	
	// Get the public key
	publicKey, err := signer.GetPublicKey(ctx)
	if err != nil {
		t.Fatalf("GetPublicKey failed: %v", err)
	}
	
	if len(publicKey) == 0 {
		t.Errorf("Expected non-empty public key")
	}
	
	// Check that the public key is in PEM format
	block, _ := pem.Decode(publicKey)
	if block == nil {
		t.Errorf("Expected PEM formatted public key")
	}
	
	if block.Type != "ECDSA PUBLIC KEY" {
		t.Errorf("Expected ECDSA PUBLIC KEY type, got %s", block.Type)
	}
}