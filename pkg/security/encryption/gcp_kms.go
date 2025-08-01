package encryption

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"strings"

	"freightliner/pkg/helper/errors"
	"freightliner/pkg/helper/log"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"google.golang.org/api/option"
)

// GCPKMS implements the Provider interface using Google Cloud KMS
type GCPKMS struct {
	client   *kms.KeyManagementClient
	keyName  string
	keyRing  string
	project  string
	location string
	logger   log.Logger
}

// GCPOpts contains options for the GCP KMS provider
type GCPOpts struct {
	Project         string
	Location        string
	KeyRing         string
	Key             string
	CredentialsFile string
	Logger          log.Logger
}

// NewGCPKMS creates a new Google Cloud KMS provider
func NewGCPKMS(ctx context.Context, opts GCPOpts) (*GCPKMS, error) {
	var clientOpts []option.ClientOption

	if opts.CredentialsFile != "" {
		clientOpts = append(clientOpts, option.WithCredentialsFile(opts.CredentialsFile))
	}

	client, err := kms.NewKeyManagementClient(ctx, clientOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create KMS client")
	}

	// Validate required fields
	if opts.Project == "" {
		return nil, errors.InvalidInputf("GCP project is required")
	}

	if opts.Location == "" {
		return nil, errors.InvalidInputf("GCP location is required")
	}

	if opts.KeyRing == "" {
		return nil, errors.InvalidInputf("GCP KMS key ring is required")
	}

	if opts.Key == "" {
		return nil, errors.InvalidInputf("GCP KMS key name is required")
	}

	if opts.Logger == nil {
		opts.Logger = log.NewBasicLogger(log.InfoLevel)
	}

	// Construction of key name for GCP KMS follows the format:
	// projects/{PROJECT_ID}/locations/{LOCATION}/keyRings/{KEY_RING}/cryptoKeys/{KEY_NAME}
	keyName := fmt.Sprintf(
		"projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		opts.Project, opts.Location, opts.KeyRing, opts.Key,
	)

	return &GCPKMS{
		client:   client,
		keyName:  keyName,
		keyRing:  opts.KeyRing,
		project:  opts.Project,
		location: opts.Location,
		logger:   opts.Logger,
	}, nil
}

// Encrypt encrypts the plaintext using the GCP KMS key
func (g *GCPKMS) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, errors.InvalidInputf("plaintext cannot be empty")
	}

	// Create the encrypt request
	req := &kmspb.EncryptRequest{
		Name:      g.keyName,
		Plaintext: plaintext,
	}

	// Call GCP KMS to encrypt the data
	resp, err := g.client.Encrypt(ctx, req)
	if err != nil {
		g.logger.WithFields(map[string]interface{}{
			"key_name": g.keyName,
		}).Error("Failed to encrypt data with GCP KMS", err)
		return nil, errors.Wrap(err, "failed to encrypt data with GCP KMS")
	}

	return resp.Ciphertext, nil
}

// Decrypt decrypts the ciphertext using the GCP KMS key
func (g *GCPKMS) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, errors.InvalidInputf("ciphertext cannot be empty")
	}

	// Create the decrypt request
	req := &kmspb.DecryptRequest{
		Name:       g.keyName,
		Ciphertext: ciphertext,
	}

	// Call GCP KMS to decrypt the data
	resp, err := g.client.Decrypt(ctx, req)
	if err != nil {
		g.logger.WithFields(map[string]interface{}{
			"key_name": g.keyName,
		}).Error("Failed to decrypt data with GCP KMS", err)
		return nil, errors.Wrap(err, "failed to decrypt data with GCP KMS")
	}

	return resp.Plaintext, nil
}

// GenerateDataKey generates a random data key and encrypts it with the GCP KMS key
func (g *GCPKMS) GenerateDataKey(ctx context.Context, keySize int) ([]byte, []byte, error) {
	if keySize <= 0 {
		return nil, nil, errors.InvalidInputf("key size must be positive, got %d", keySize)
	}

	// Generate a random data key
	plaintext := make([]byte, keySize)
	_, err := io.ReadFull(rand.Reader, plaintext)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate random data key")
	}

	// Encrypt the data key
	ciphertext, err := g.Encrypt(ctx, plaintext)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to encrypt data key")
	}

	return plaintext, ciphertext, nil
}

// Name returns the provider name
func (g *GCPKMS) Name() string {
	return "gcp-kms"
}

// GetKeyName returns the full GCP KMS key name
func (g *GCPKMS) GetKeyName() string {
	return g.keyName
}

// GetKeyInfo returns information about the KMS key
func (g *GCPKMS) GetKeyInfo() map[string]string {
	return map[string]string{
		"provider": "gcp-kms",
		"project":  g.project,
		"location": g.location,
		"keyRing":  g.keyRing,
		"keyName":  g.keyName,
	}
}

// Close closes the GCP KMS client
func (g *GCPKMS) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}

// ValidateGCPKMSKeyName validates a GCP KMS key name
func ValidateGCPKMSKeyName(keyName string) error {
	// Check if the key name is properly formatted
	// Expected format: projects/{PROJECT_ID}/locations/{LOCATION}/keyRings/{KEY_RING}/cryptoKeys/{KEY_NAME}
	parts := strings.Split(keyName, "/")
	if len(parts) != 8 {
		return errors.InvalidInputf("invalid GCP KMS key name format: %s", keyName)
	}

	if parts[0] != "projects" {
		return errors.InvalidInputf("GCP KMS key name must start with 'projects/': %s", keyName)
	}

	if parts[2] != "locations" {
		return errors.InvalidInputf("GCP KMS key name missing 'locations/': %s", keyName)
	}

	if parts[4] != "keyRings" {
		return errors.InvalidInputf("GCP KMS key name missing 'keyRings/': %s", keyName)
	}

	if parts[6] != "cryptoKeys" {
		return errors.InvalidInputf("GCP KMS key name missing 'cryptoKeys/': %s", keyName)
	}

	// Check that we have non-empty values for project, location, key ring, and key
	if parts[1] == "" {
		return errors.InvalidInputf("GCP KMS key name has empty project ID")
	}

	if parts[3] == "" {
		return errors.InvalidInputf("GCP KMS key name has empty location")
	}

	if parts[5] == "" {
		return errors.InvalidInputf("GCP KMS key name has empty key ring")
	}

	if parts[7] == "" {
		return errors.InvalidInputf("GCP KMS key name has empty key name")
	}

	return nil
}
