package signing

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"time"
)

// CosignSigner implements the Signer interface using Cosign signatures
type CosignSigner struct {
	options SignOptions
	privKey crypto.PrivateKey
	pubKey  crypto.PublicKey
}

// CosignSignature represents a Cosign signature format
type CosignSignature struct {
	// Critical is the critical section of the payload (always signed)
	Critical struct {
		Identity struct {
			DockerReference string `json:"docker-reference"`
		} `json:"identity"`
		Image struct {
			DockerManifestDigest string `json:"docker-manifest-digest"`
		} `json:"image"`
		Type string `json:"type"`
	} `json:"critical"`

	// Optional is the optional section of the payload (not required to be signed)
	Optional map[string]interface{} `json:"optional,omitempty"`
}

// NewCosignSigner creates a new Cosign signer
func NewCosignSigner(options SignOptions) (*CosignSigner, error) {
	signer := &CosignSigner{
		options: options,
	}

	if options.KeyPath == "" {
		return nil, fmt.Errorf("key path is required for Cosign signer")
	}

	// Read the private key file
	keyData, err := os.ReadFile(options.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	// Parse the PEM encoded private key
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the key")
	}

	// Handle encrypted keys if passphrase is provided
	if block.Type == "ENCRYPTED PRIVATE KEY" && options.PassphraseReader != nil {
		passphrase, err := io.ReadAll(options.PassphraseReader)
		if err != nil {
			return nil, fmt.Errorf("failed to read passphrase: %w", err)
		}

		// In a real implementation, we would decrypt the key here
		// For this example, we'll just continue with a placeholder
		fmt.Println("Using passphrase to decrypt key")
	}

	// Parse the private key based on the type
	var privKey crypto.PrivateKey
	var pubKey crypto.PublicKey

	switch block.Type {
	case "PRIVATE KEY", "ECDSA PRIVATE KEY":
		// Parse ECDSA private key
		ecdsaKey, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ECDSA private key: %w", err)
		}
		privKey = ecdsaKey
		pubKey = &ecdsaKey.PublicKey

	case "ED25519 PRIVATE KEY":
		// Parse Ed25519 private key
		if len(block.Bytes) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf("invalid Ed25519 private key size")
		}
		privKey = ed25519.PrivateKey(block.Bytes)
		pubKey = privKey.(ed25519.PrivateKey).Public()

	default:
		return nil, fmt.Errorf("unsupported key type: %s", block.Type)
	}

	signer.privKey = privKey
	signer.pubKey = pubKey

	return signer, nil
}

// Name returns the name of the signing provider
func (c *CosignSigner) Name() string {
	return "cosign"
}

// Sign signs a container image digest using Cosign format
func (c *CosignSigner) Sign(ctx context.Context, payload *SignaturePayload) (*Signature, error) {
	// Create a Cosign signature payload
	cosignPayload := CosignSignature{}
	cosignPayload.Critical.Identity.DockerReference = payload.Repository + ":" + payload.Tag
	cosignPayload.Critical.Image.DockerManifestDigest = payload.ManifestDigest
	cosignPayload.Critical.Type = "cosign container image signature"

	// Add additional data to optional section
	if payload.AdditionalData != nil && len(payload.AdditionalData) > 0 {
		cosignPayload.Optional = make(map[string]interface{})
		for k, v := range payload.AdditionalData {
			cosignPayload.Optional[k] = v
		}
	}

	// Add standard fields to optional section
	timestamp := time.Now().UTC().Unix()
	if v := ctx.Value("timestamp"); v != nil {
		if ts, ok := v.(int64); ok {
			timestamp = ts
		}
	}

	if cosignPayload.Optional == nil {
		cosignPayload.Optional = make(map[string]interface{})
	}
	cosignPayload.Optional["created"] = timestamp

	// Marshal the payload to JSON
	payloadBytes, err := json.Marshal(cosignPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal signature payload: %w", err)
	}

	// Calculate the hash of the payload
	payloadHash := sha256.Sum256(payloadBytes)

	// Sign the hash based on the key type
	var signatureBytes []byte
	switch key := c.privKey.(type) {
	case *ecdsa.PrivateKey:
		// Sign with ECDSA
		signatureBytes, err = ecdsa.SignASN1(rand.Reader, key, payloadHash[:])
		if err != nil {
			return nil, fmt.Errorf("failed to sign with ECDSA: %w", err)
		}

	case ed25519.PrivateKey:
		// Sign with Ed25519
		signatureBytes = ed25519.Sign(key, payloadBytes)

	default:
		return nil, fmt.Errorf("unsupported private key type: %T", c.privKey)
	}

	// Create the signature object
	signature := &Signature{
		Payload:   payloadBytes,
		Signature: signatureBytes,
		KeyID:     c.options.KeyID,
		Metadata: map[string]string{
			"format":  "cosign",
			"type":    "container-image",
			"digest":  payload.ManifestDigest,
			"created": fmt.Sprintf("%d", timestamp),
		},
	}

	return signature, nil
}

// Verify verifies a Cosign signature
func (c *CosignSigner) Verify(ctx context.Context, payload *SignaturePayload, signature *Signature) (bool, error) {
	// Verify the payload matches what we expect
	var cosignPayload CosignSignature
	if err := json.Unmarshal(signature.Payload, &cosignPayload); err != nil {
		return false, fmt.Errorf("failed to unmarshal signature payload: %w", err)
	}

	// Verify the manifest digest matches
	if cosignPayload.Critical.Image.DockerManifestDigest != payload.ManifestDigest {
		return false, fmt.Errorf("manifest digest mismatch")
	}

	// Verify the repository and tag
	expectedRef := payload.Repository + ":" + payload.Tag
	if cosignPayload.Critical.Identity.DockerReference != expectedRef {
		return false, fmt.Errorf("docker reference mismatch")
	}

	// Calculate the hash of the payload
	payloadHash := sha256.Sum256(signature.Payload)

	// Verify the signature based on the key type
	var valid bool
	switch key := c.pubKey.(type) {
	case *ecdsa.PublicKey:
		// Verify with ECDSA
		valid = ecdsa.VerifyASN1(key, payloadHash[:], signature.Signature)

	case ed25519.PublicKey:
		// Verify with Ed25519
		valid = ed25519.Verify(key, signature.Payload, signature.Signature)

	default:
		return false, fmt.Errorf("unsupported public key type: %T", c.pubKey)
	}

	return valid, nil
}

// GetPublicKey returns the public key used for verification
func (c *CosignSigner) GetPublicKey(ctx context.Context) ([]byte, error) {
	// Extract the public key based on its type
	var pubKeyBytes []byte
	var pubKeyType string

	switch key := c.pubKey.(type) {
	case *ecdsa.PublicKey:
		var err error
		pubKeyBytes, err = x509.MarshalPKIXPublicKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal ECDSA public key: %w", err)
		}
		pubKeyType = "ECDSA PUBLIC KEY"

	case ed25519.PublicKey:
		pubKeyBytes = []byte(key)
		pubKeyType = "ED25519 PUBLIC KEY"

	default:
		return nil, fmt.Errorf("unsupported public key type: %T", c.pubKey)
	}

	// Encode the public key in PEM format
	pemBlock := &pem.Block{
		Type:  pubKeyType,
		Bytes: pubKeyBytes,
	}

	return pem.EncodeToMemory(pemBlock), nil
}

// StoreSignature stores a signature for a container image
func (c *CosignSigner) StoreSignature(ctx context.Context, signature *Signature, path string) error {
	// Encode the signature to store - we'll use a simple JSON format
	data := map[string]interface{}{
		"payload":   base64.StdEncoding.EncodeToString(signature.Payload),
		"signature": base64.StdEncoding.EncodeToString(signature.Signature),
		"keyId":     signature.KeyID,
		"metadata":  signature.Metadata,
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal signature data: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write signature to file: %w", err)
	}

	return nil
}

// LoadSignature loads a signature from disk
func (c *CosignSigner) LoadSignature(ctx context.Context, path string) (*Signature, error) {
	// Read the signature file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read signature file: %w", err)
	}

	// Parse the JSON
	var sigData map[string]interface{}
	if err := json.Unmarshal(data, &sigData); err != nil {
		return nil, fmt.Errorf("failed to parse signature data: %w", err)
	}

	// Extract and decode the fields
	payloadStr, ok := sigData["payload"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid signature format: missing payload")
	}

	sigStr, ok := sigData["signature"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid signature format: missing signature")
	}

	keyID, _ := sigData["keyId"].(string)

	// Convert metadata map
	metadata := make(map[string]string)
	if metaMap, ok := sigData["metadata"].(map[string]interface{}); ok {
		for k, v := range metaMap {
			if strVal, ok := v.(string); ok {
				metadata[k] = strVal
			}
		}
	}

	// Decode from base64
	payload, err := base64.StdEncoding.DecodeString(payloadStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	sig, err := base64.StdEncoding.DecodeString(sigStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	// Create and return the signature
	return &Signature{
		Payload:   payload,
		Signature: sig,
		KeyID:     keyID,
		Metadata:  metadata,
	}, nil
}