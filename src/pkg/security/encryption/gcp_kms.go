package encryption

import (
	"context"
	"fmt"
	"strings"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GCPKMS implements the Provider interface for Google Cloud KMS
type GCPKMS struct {
	client   *kms.KeyManagementClient
	location string
	keyID    string
	project  string
}

// GCPOpts defines GCP KMS-specific options
type GCPOpts struct {
	// Project is the GCP project ID
	Project string

	// Location is the GCP location where the KMS key is stored
	Location string

	// KeyRing is the KMS key ring name
	KeyRing string

	// Key is the KMS key name
	Key string

	// KeyVersion is the KMS key version (optional, defaults to latest)
	KeyVersion string

	// CredentialsFile is the path to a service account key file
	CredentialsFile string

	// CredentialsJSON is the raw JSON credentials content
	CredentialsJSON string
}

// NewGCPKMS creates a new GCP KMS encryption provider
func NewGCPKMS(ctx context.Context, opts GCPOpts) (*GCPKMS, error) {
	var clientOpts []option.ClientOption

	// Use credentials file if provided
	if opts.CredentialsFile != "" {
		clientOpts = append(clientOpts, option.WithCredentialsFile(opts.CredentialsFile))
	}

	// Use credentials JSON if provided (takes precedence over file)
	if opts.CredentialsJSON != "" {
		clientOpts = append(clientOpts, option.WithCredentialsJSON([]byte(opts.CredentialsJSON)))
	}

	// Create the KMS client
	client, err := kms.NewKeyManagementClient(ctx, clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP KMS client: %w", err)
	}

	// Build full KMS key path
	keyVersion := "1"
	if opts.KeyVersion != "" {
		keyVersion = opts.KeyVersion
	}

	// Construct the full key path
	keyPath := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s/cryptoKeyVersions/%s",
		opts.Project, opts.Location, opts.KeyRing, opts.Key, keyVersion)

	return &GCPKMS{
		client:   client,
		location: opts.Location,
		keyID:    keyPath,
		project:  opts.Project,
	}, nil
}

// Name returns the provider name
func (g *GCPKMS) Name() string {
	return "gcp-kms"
}

// Encrypt encrypts plaintext using GCP KMS
func (g *GCPKMS) Encrypt(ctx context.Context, plaintext []byte, keyID string) ([]byte, error) {
	// Use default key if none specified
	if keyID == "" {
		keyID = g.keyID
	}

	if keyID == "" {
		return nil, fmt.Errorf("no KMS key ID specified for encryption")
	}

	req := &kmspb.EncryptRequest{
		Name:      keyID,
		Plaintext: plaintext,
	}

	resp, err := g.client.Encrypt(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt with GCP KMS: %w", err)
	}

	return resp.Ciphertext, nil
}

// Decrypt decrypts ciphertext using GCP KMS
func (g *GCPKMS) Decrypt(ctx context.Context, ciphertext []byte, keyID string) ([]byte, error) {
	// Use default key if none specified
	if keyID == "" {
		keyID = g.keyID
	}

	if keyID == "" {
		return nil, fmt.Errorf("no KMS key ID specified for decryption")
	}

	// GCP KMS requires the crypto key, not the version
	keyPath := keyID
	// If the key includes a version, remove it
	if strings.Contains(keyID, "cryptoKeyVersions") {
		parts := strings.Split(keyID, "/cryptoKeyVersions")
		keyPath = parts[0]
	}

	req := &kmspb.DecryptRequest{
		Name:       keyPath,
		Ciphertext: ciphertext,
	}

	resp, err := g.client.Decrypt(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with GCP KMS: %w", err)
	}

	return resp.Plaintext, nil
}

// GenerateDataKey generates a data key for envelope encryption
func (g *GCPKMS) GenerateDataKey(ctx context.Context, keyID string, keyLength int) (*DataKey, error) {
	// Use default key if none specified
	if keyID == "" {
		keyID = g.keyID
	}

	if keyID == "" {
		return nil, fmt.Errorf("no KMS key ID specified for data key generation")
	}

	// Default key length to AES-256 if not specified
	if keyLength <= 0 {
		keyLength = 32 // 256 bits
	}

	// Generate random data key using crypto/rand
	// Note: GCP KMS doesn't have a direct equivalent to AWS KMS GenerateDataKey,
	// so we generate a random key and encrypt it with KMS
	plaintext := make([]byte, keyLength)
	_, err := getRandomBytes(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random data: %w", err)
	}

	// Encrypt the data key with KMS
	ciphertext, err := g.Encrypt(ctx, plaintext, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data key: %w", err)
	}

	return &DataKey{
		Plaintext:  plaintext,
		Ciphertext: ciphertext,
	}, nil
}

// ReEncrypt re-encrypts already encrypted data with a different key
func (g *GCPKMS) ReEncrypt(ctx context.Context, ciphertext []byte, sourceKeyID, destinationKeyID string) ([]byte, error) {
	// GCP KMS doesn't have a direct ReEncrypt API like AWS KMS
	// We need to decrypt and then encrypt again

	// Use default key if none specified for destination
	if destinationKeyID == "" {
		destinationKeyID = g.keyID
	}

	if destinationKeyID == "" {
		return nil, fmt.Errorf("no destination KMS key ID specified for re-encryption")
	}

	// Decrypt with source key
	plaintext, err := g.Decrypt(ctx, ciphertext, sourceKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with source key during re-encryption: %w", err)
	}

	// Encrypt with destination key
	newCiphertext, err := g.Encrypt(ctx, plaintext, destinationKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt with destination key during re-encryption: %w", err)
	}

	return newCiphertext, nil
}

// GetKeyInfo retrieves information about a KMS key
func (g *GCPKMS) GetKeyInfo(ctx context.Context, keyID string) (*KeyInfo, error) {
	// Use default key if none specified
	if keyID == "" {
		keyID = g.keyID
	}

	if keyID == "" {
		return nil, fmt.Errorf("no KMS key ID specified for key info")
	}

	// Extract the crypto key path (without version)
	keyPath := keyID
	if strings.Contains(keyID, "/cryptoKeyVersions/") {
		parts := strings.Split(keyID, "/cryptoKeyVersions/")
		keyPath = parts[0]
	}

	// Get key information
	req := &kmspb.GetCryptoKeyRequest{
		Name: keyPath,
	}

	_, err := g.client.GetCryptoKey(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get key info from GCP KMS: %w", err)
	}

	// Get the primary version for more details
	versionReq := &kmspb.GetCryptoKeyVersionRequest{
		Name: fmt.Sprintf("%s/cryptoKeyVersions/1", keyPath),
	}

	keyVersion, err := g.client.GetCryptoKeyVersion(ctx, versionReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get key version info from GCP KMS: %w", err)
	}

	// Create the key info
	keyInfo := &KeyInfo{
		ID:         keyPath,
		ARN:        keyPath, // GCP doesn't have ARNs, using full path as equivalent
		Algorithm:  keyVersion.Algorithm.String(),
		State:      keyVersion.State.String(),
		Enabled:    keyVersion.State == kmspb.CryptoKeyVersion_ENABLED,
		Provider:   "gcp-kms",
		Region:     g.location,
		CreateTime: keyVersion.CreateTime.AsTime(),
	}

	// GCP doesn't have a direct "customer managed" flag, but we can infer it
	// All keys created through the API are customer managed
	keyInfo.CustomerManaged = true

	return keyInfo, nil
}

// ListKeyRings lists all key rings in the specified project and location
func (g *GCPKMS) ListKeyRings(ctx context.Context) ([]string, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", g.project, g.location)

	req := &kmspb.ListKeyRingsRequest{
		Parent: parent,
	}

	it := g.client.ListKeyRings(ctx, req)

	var keyRings []string
	for {
		keyRing, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list key rings: %w", err)
		}

		keyRings = append(keyRings, keyRing.Name)
	}

	return keyRings, nil
}

// ListKeys lists all keys in the specified key ring
func (g *GCPKMS) ListKeys(ctx context.Context, keyRingPath string) ([]string, error) {
	req := &kmspb.ListCryptoKeysRequest{
		Parent: keyRingPath,
	}

	it := g.client.ListCryptoKeys(ctx, req)

	var keys []string
	for {
		key, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list keys: %w", err)
		}

		keys = append(keys, key.Name)
	}

	return keys, nil
}
