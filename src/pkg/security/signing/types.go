package signing

import (
	"context"
	"io"
)

// SignaturePayload represents a signature payload
type SignaturePayload struct {
	// ManifestDigest is the digest of the manifest being signed
	ManifestDigest string

	// Repository is the repository name
	Repository string

	// Tag is the tag being signed
	Tag string

	// AdditionalData is optional additional data to include in the signature
	AdditionalData map[string]string
}

// Signature represents a container image signature
type Signature struct {
	// Payload is the signed payload data
	Payload []byte

	// Signature is the actual signature bytes
	Signature []byte

	// KeyID is the identifier of the key used to create the signature
	KeyID string

	// Metadata is additional metadata about the signature
	Metadata map[string]string
}

// Signer defines the interface for signing container images
type Signer interface {
	// Name returns the name of the signing provider
	Name() string

	// Sign signs a container image digest
	Sign(ctx context.Context, payload *SignaturePayload) (*Signature, error)

	// Verify verifies a signature against a digest
	Verify(ctx context.Context, payload *SignaturePayload, signature *Signature) (bool, error)

	// GetPublicKey returns the public key used for verification
	GetPublicKey(ctx context.Context) ([]byte, error)
}

// SignOptions defines options for signing operations
type SignOptions struct {
	// KeyID is the ID of the key to use for signing
	KeyID string

	// KeyPath is the path to the private key file (for file-based signers)
	KeyPath string

	// PassphraseReader provides a passphrase if the key is encrypted
	PassphraseReader io.Reader
}

// VerifyOptions defines options for verification operations
type VerifyOptions struct {
	// KeyID is the ID of the key to use for verification
	KeyID string

	// KeyPath is the path to the public key file (for file-based verifiers)
	KeyPath string

	// AllowedSigners is a list of allowed signer identities
	AllowedSigners []string

	// RequireSignatures is the minimum number of valid signatures required
	RequireSignatures int

	// TrustStore is the path to a trust store for verification
	TrustStore string
}