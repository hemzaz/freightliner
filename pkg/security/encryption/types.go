package encryption

import (
	"context"
	"crypto/rand"
	"time"
)

// Provider defines the interface for encryption/decryption providers
type Provider interface {
	// Name returns the name of the encryption provider
	Name() string

	// Encrypt encrypts the plaintext
	Encrypt(ctx context.Context, plaintext []byte) ([]byte, error)

	// Decrypt decrypts the ciphertext
	Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error)

	// GenerateDataKey generates a data key that can be used for envelope encryption
	GenerateDataKey(ctx context.Context, keyLength int) ([]byte, []byte, error)

	// GetKeyInfo returns metadata about the KMS key used by this provider
	GetKeyInfo() map[string]string
}

// DataKey represents an envelope encryption data key
type DataKey struct {
	// Plaintext is the plain data key that can be used for local encryption
	Plaintext []byte

	// Ciphertext is the encrypted data key that should be stored alongside the encrypted data
	Ciphertext []byte
}

// KeyInfo provides details about a KMS key
type KeyInfo struct {
	// ID is the key ID
	ID string

	// ARN is the AWS ARN or full resource path for GCP
	ARN string

	// Algorithm is the cryptographic algorithm of the key
	Algorithm string

	// State is the current state of the key (e.g., enabled, disabled)
	State string

	// Enabled indicates whether the key is enabled for use
	Enabled bool

	// CustomerManaged indicates whether this is a customer-managed key (vs. service-managed)
	CustomerManaged bool

	// Provider is the KMS provider name (aws-kms, gcp-kms)
	Provider string

	// Region is the region or location where the key is stored
	Region string

	// CreateTime is when the key was created
	CreateTime time.Time
}

// ProviderOptions defines common options for encryption providers
type ProviderOptions struct {
	// Region is the cloud provider region (AWS region or GCP location)
	Region string

	// KeyID is the default KMS key ID to use if none is specified
	KeyID string

	// CustomerManagedKey indicates whether to use a customer-managed key
	CustomerManagedKey bool
}

// EncryptionConfig defines the configuration for encrypting container images
type EncryptionConfig struct {
	// Provider is the name of the encryption provider (aws-kms, gcp-kms)
	Provider string

	// KeyID is the ID of the key to use for encryption
	KeyID string

	// Region is the region where the key is located
	Region string

	// CustomerManagedKey indicates whether to use a customer-managed key
	CustomerManagedKey bool

	// EnvelopeEncryption indicates whether to use envelope encryption
	// (encrypt content with a data key, then encrypt the data key with KMS)
	EnvelopeEncryption bool

	// DataKeyLength is the length of data keys for envelope encryption (in bytes)
	DataKeyLength int
}

// getRandomBytes fills the provided byte slice with cryptographically secure random bytes
func getRandomBytes(b []byte) (n int, err error) {
	return rand.Read(b)
}

// EncryptOptions contains options for encryption operations
type EncryptOptions struct {
	// Provider specifies which encryption provider to use
	Provider string
}

// DecryptOptions contains options for decryption operations
type DecryptOptions struct {
	// Provider specifies which encryption provider to use
	Provider string
}

// Envelope holds metadata for envelope encryption
type Envelope struct {
	// EncryptedKey is the encrypted data key
	EncryptedKey []byte `json:"encrypted_key"`

	// Ciphertext is the encrypted data (only used for data encryption, not for streams)
	Ciphertext []byte `json:"ciphertext,omitempty"`

	// ProviderInfo contains metadata about the provider used for encryption
	ProviderInfo map[string]string `json:"provider_info,omitempty"`

	// EnvelopeFormat specifies the format of the envelope (e.g., AES-GCM)
	EnvelopeFormat string `json:"envelope_format,omitempty"`
}
