// Package signatures provides interfaces for container image signature verification
// with support for Cosign/Sigstore integration.
package signatures

import (
	"context"
	"crypto"
	"io"
	"time"
)

// ImageSigner defines the interface for signing container images.
// Implementations should support various signing mechanisms including:
// - Cosign (sigstore)
// - Notary v2
// - GPG-based signing
type ImageSigner interface {
	// Sign signs a container image and returns signature metadata
	Sign(ctx context.Context, imageRef string, opts SigningOptions) (*SignatureMetadata, error)

	// SignDigest signs a specific image digest
	SignDigest(ctx context.Context, digest string, opts SigningOptions) (*SignatureMetadata, error)

	// SignManifest signs an OCI manifest directly
	SignManifest(ctx context.Context, manifest []byte, opts SigningOptions) (*SignatureMetadata, error)

	// GetSigningKey retrieves the public key used for signing
	GetSigningKey(ctx context.Context) (crypto.PublicKey, error)

	// SupportedAlgorithms returns the list of supported signature algorithms
	SupportedAlgorithms() []string
}

// SignatureVerifier defines the interface for verifying container image signatures.
// Supports multiple verification strategies and trust policies.
type SignatureVerifier interface {
	// Verify verifies all signatures for a given image reference
	Verify(ctx context.Context, imageRef string, opts VerificationOptions) (*VerificationResult, error)

	// VerifyDigest verifies signatures for a specific image digest
	VerifyDigest(ctx context.Context, digest string, opts VerificationOptions) (*VerificationResult, error)

	// VerifyWithKey verifies a signature using a specific public key
	VerifyWithKey(ctx context.Context, imageRef string, publicKey crypto.PublicKey, opts VerificationOptions) (*VerificationResult, error)

	// VerifyAttestation verifies image attestations (SLSA provenance, SBOM, etc.)
	VerifyAttestation(ctx context.Context, imageRef string, attestationType AttestationType, opts VerificationOptions) (*AttestationResult, error)

	// ListSignatures retrieves all signatures associated with an image
	ListSignatures(ctx context.Context, imageRef string) ([]*SignatureMetadata, error)

	// ValidateTrustPolicy checks if image meets trust policy requirements
	ValidateTrustPolicy(ctx context.Context, imageRef string, policy *TrustPolicy) (*PolicyValidationResult, error)
}

// KeyProvider defines the interface for managing cryptographic keys.
// Supports various key storage backends including KMS, Vault, and local files.
type KeyProvider interface {
	// GetPublicKey retrieves a public key by identifier
	GetPublicKey(ctx context.Context, keyID string) (crypto.PublicKey, error)

	// GetPrivateKey retrieves a private key by identifier (requires authorization)
	GetPrivateKey(ctx context.Context, keyID string) (crypto.PrivateKey, error)

	// ListKeys returns all available key identifiers
	ListKeys(ctx context.Context) ([]KeyInfo, error)

	// CreateKey generates a new key pair
	CreateKey(ctx context.Context, keyType KeyType, opts KeyGenerationOptions) (*KeyInfo, error)

	// RotateKey rotates an existing key pair
	RotateKey(ctx context.Context, keyID string) (*KeyInfo, error)

	// RevokeKey marks a key as revoked
	RevokeKey(ctx context.Context, keyID string, reason string) error

	// ValidateKey checks if a key is valid and not revoked
	ValidateKey(ctx context.Context, keyID string) (bool, error)

	// ExportPublicKey exports a public key in PEM format
	ExportPublicKey(ctx context.Context, keyID string, w io.Writer) error

	// ImportPublicKey imports a public key from PEM format
	ImportPublicKey(ctx context.Context, keyID string, r io.Reader) error
}

// CertificateAuthority defines the interface for certificate-based trust.
// Supports X.509 certificate validation and trust chain verification.
type CertificateAuthority interface {
	// VerifyCertificate verifies a certificate against the CA
	VerifyCertificate(ctx context.Context, cert []byte) error

	// GetTrustBundle returns the CA trust bundle
	GetTrustBundle(ctx context.Context) ([]byte, error)

	// ValidateChain validates a certificate chain
	ValidateChain(ctx context.Context, chain [][]byte) error

	// IssueCertificate issues a new certificate (if CA has signing capability)
	IssueCertificate(ctx context.Context, csr []byte, opts CertificateOptions) ([]byte, error)

	// RevokeCertificate revokes a certificate
	RevokeCertificate(ctx context.Context, serial string, reason RevocationReason) error

	// CheckRevocation checks if a certificate is revoked
	CheckRevocation(ctx context.Context, serial string) (bool, error)
}

// TransparencyLog defines the interface for interacting with transparency logs.
// Primarily designed for Rekor (Sigstore) integration but extensible to other systems.
type TransparencyLog interface {
	// Upload uploads an entry to the transparency log
	Upload(ctx context.Context, entry *LogEntry) (*LogEntryResult, error)

	// Verify verifies an entry in the transparency log
	Verify(ctx context.Context, entryID string) (*LogEntryResult, error)

	// Search searches the transparency log for entries
	Search(ctx context.Context, query SearchQuery) ([]*LogEntryResult, error)

	// GetEntry retrieves a specific log entry
	GetEntry(ctx context.Context, entryID string) (*LogEntry, error)

	// GetProof gets inclusion proof for a log entry
	GetProof(ctx context.Context, entryID string) (*InclusionProof, error)

	// VerifyProof verifies an inclusion proof
	VerifyProof(ctx context.Context, proof *InclusionProof) (bool, error)
}

// PolicyEngine defines the interface for evaluating signature policies.
// Supports OPA, Rego, and custom policy languages.
type PolicyEngine interface {
	// EvaluatePolicy evaluates a policy against image metadata
	EvaluatePolicy(ctx context.Context, policy *Policy, subject *PolicySubject) (*PolicyResult, error)

	// ValidatePolicy validates policy syntax and structure
	ValidatePolicy(ctx context.Context, policy *Policy) error

	// LoadPolicy loads a policy from a source
	LoadPolicy(ctx context.Context, source PolicySource) (*Policy, error)

	// ListPolicies returns all loaded policies
	ListPolicies(ctx context.Context) ([]*PolicyInfo, error)

	// TestPolicy tests a policy with sample data
	TestPolicy(ctx context.Context, policy *Policy, testCases []PolicyTestCase) ([]*PolicyTestResult, error)
}

// AttestationProvider defines the interface for managing attestations.
// Supports SLSA provenance, SBOMs, vulnerability scans, and custom attestations.
type AttestationProvider interface {
	// CreateAttestation creates a new attestation
	CreateAttestation(ctx context.Context, attestation *Attestation) (*AttestationMetadata, error)

	// GetAttestation retrieves an attestation
	GetAttestation(ctx context.Context, imageRef string, attestationType AttestationType) (*Attestation, error)

	// ListAttestations lists all attestations for an image
	ListAttestations(ctx context.Context, imageRef string) ([]*AttestationMetadata, error)

	// VerifyAttestation verifies an attestation signature
	VerifyAttestation(ctx context.Context, attestation *Attestation, opts VerificationOptions) (*VerificationResult, error)

	// AttachAttestation attaches an attestation to an image
	AttachAttestation(ctx context.Context, imageRef string, attestation *Attestation) error

	// ValidateAttestation validates attestation content against schema
	ValidateAttestation(ctx context.Context, attestation *Attestation) error
}

// SigningOptions defines options for signing operations
type SigningOptions struct {
	// KeyID specifies which key to use for signing
	KeyID string

	// Algorithm specifies the signature algorithm (e.g., "ECDSA_P256_SHA256")
	Algorithm string

	// Annotations are additional metadata to include in signature
	Annotations map[string]string

	// Recursive indicates whether to sign recursively (for multi-arch images)
	Recursive bool

	// UploadToTLog indicates whether to upload to transparency log
	UploadToTLog bool

	// TLogURL specifies the transparency log URL (defaults to Rekor)
	TLogURL string

	// IdentityToken is an OIDC identity token for keyless signing
	IdentityToken string

	// AllowInsecure allows insecure registries
	AllowInsecure bool

	// Timeout for the signing operation
	Timeout time.Duration
}

// VerificationOptions defines options for verification operations
type VerificationOptions struct {
	// PublicKeys are explicit public keys to verify against
	PublicKeys []crypto.PublicKey

	// KeyIDs are key identifiers to load from KeyProvider
	KeyIDs []string

	// CertificateIdentity for verifying keyless signatures
	CertificateIdentity string

	// CertificateOIDCIssuer for verifying keyless signatures
	CertificateOIDCIssuer string

	// CheckClaims validates specific claims in the signature
	CheckClaims map[string]string

	// RequireRekorBundle requires a Rekor bundle for verification
	RequireRekorBundle bool

	// AllowInsecure allows insecure registries
	AllowInsecure bool

	// IgnoreSCT ignores SignedCertificateTimestamp validation
	IgnoreSCT bool

	// IgnoreTLog ignores transparency log verification
	IgnoreTLog bool

	// MaxSignatureAge is the maximum age of signatures to accept
	MaxSignatureAge time.Duration

	// Timeout for the verification operation
	Timeout time.Duration
}

// KeyGenerationOptions defines options for key generation
type KeyGenerationOptions struct {
	// Algorithm specifies the key algorithm (RSA, ECDSA, Ed25519)
	Algorithm string

	// KeySize specifies the key size in bits (for RSA)
	KeySize int

	// Curve specifies the elliptic curve (for ECDSA)
	Curve string

	// Password for key encryption (optional)
	Password string

	// Metadata for the key
	Metadata map[string]string

	// ExpiresAt specifies when the key expires
	ExpiresAt *time.Time
}

// CertificateOptions defines options for certificate issuance
type CertificateOptions struct {
	// Subject is the certificate subject
	Subject string

	// DNSNames are DNS SANs
	DNSNames []string

	// EmailAddresses are email SANs
	EmailAddresses []string

	// ValidFor is the certificate validity duration
	ValidFor time.Duration

	// KeyUsage specifies key usage extensions
	KeyUsage []string

	// ExtendedKeyUsage specifies extended key usage
	ExtendedKeyUsage []string
}

// RevocationReason defines reasons for key/certificate revocation
type RevocationReason string

const (
	RevocationReasonUnspecified          RevocationReason = "unspecified"
	RevocationReasonKeyCompromise        RevocationReason = "keyCompromise"
	RevocationReasonCACompromise         RevocationReason = "caCompromise"
	RevocationReasonAffiliationChanged   RevocationReason = "affiliationChanged"
	RevocationReasonSuperseded           RevocationReason = "superseded"
	RevocationReasonCessationOfOperation RevocationReason = "cessationOfOperation"
	RevocationReasonPrivilegeWithdrawn   RevocationReason = "privilegeWithdrawn"
)
