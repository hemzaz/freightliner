package signatures

import (
	"crypto"
	"time"
)

// SignatureMetadata contains metadata about a signature
type SignatureMetadata struct {
	// ImageRef is the reference to the signed image
	ImageRef string `json:"imageRef"`

	// Digest is the image digest that was signed
	Digest string `json:"digest"`

	// SignatureDigest is the digest of the signature itself
	SignatureDigest string `json:"signatureDigest"`

	// Algorithm is the signature algorithm used
	Algorithm string `json:"algorithm"`

	// KeyID identifies the key used for signing
	KeyID string `json:"keyId"`

	// PublicKey is the public key used (optional, may be in certificate)
	PublicKey crypto.PublicKey `json:"-"`

	// Certificate is the X.509 certificate (for keyless signing)
	Certificate []byte `json:"certificate,omitempty"`

	// CertificateChain is the full certificate chain
	CertificateChain [][]byte `json:"certificateChain,omitempty"`

	// SignedAt is when the signature was created
	SignedAt time.Time `json:"signedAt"`

	// ExpiresAt is when the signature expires (optional)
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`

	// Issuer is the OIDC issuer for keyless signatures
	Issuer string `json:"issuer,omitempty"`

	// Subject is the certificate subject or OIDC subject
	Subject string `json:"subject,omitempty"`

	// Annotations are additional metadata
	Annotations map[string]string `json:"annotations,omitempty"`

	// RekorBundle contains the transparency log entry
	RekorBundle *RekorBundle `json:"rekorBundle,omitempty"`

	// SignaturePayload is the actual signature bytes
	SignaturePayload []byte `json:"signaturePayload"`

	// PayloadFormat describes the signature payload format (e.g., "cosign", "notary")
	PayloadFormat string `json:"payloadFormat"`
}

// VerificationResult contains the result of signature verification
type VerificationResult struct {
	// Verified indicates if verification was successful
	Verified bool `json:"verified"`

	// ImageRef is the image that was verified
	ImageRef string `json:"imageRef"`

	// Digest is the verified image digest
	Digest string `json:"digest"`

	// Signatures are all signatures that were verified
	Signatures []*SignatureMetadata `json:"signatures"`

	// VerifiedSignatures are signatures that passed verification
	VerifiedSignatures []*SignatureMetadata `json:"verifiedSignatures"`

	// FailedSignatures are signatures that failed verification
	FailedSignatures []*SignatureMetadata `json:"failedSignatures"`

	// TrustRoot describes the trust root used for verification
	TrustRoot string `json:"trustRoot"`

	// VerifiedAt is when verification was performed
	VerifiedAt time.Time `json:"verifiedAt"`

	// ValidationErrors contains any validation errors
	ValidationErrors []error `json:"-"`

	// ValidationMessages contains human-readable validation messages
	ValidationMessages []string `json:"validationMessages,omitempty"`

	// CertificateIdentity is the verified certificate identity (keyless)
	CertificateIdentity string `json:"certificateIdentity,omitempty"`

	// CertificateOIDCIssuer is the verified OIDC issuer (keyless)
	CertificateOIDCIssuer string `json:"certificateOIDCIssuer,omitempty"`

	// TransparencyLogVerified indicates if transparency log was verified
	TransparencyLogVerified bool `json:"transparencyLogVerified"`

	// Metadata contains additional verification metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SigningConfig contains configuration for signing operations
type SigningConfig struct {
	// Enabled indicates if signing is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// DefaultKeyID is the default key to use for signing
	DefaultKeyID string `json:"defaultKeyId" yaml:"defaultKeyId"`

	// Algorithm is the default signature algorithm
	Algorithm string `json:"algorithm" yaml:"algorithm"`

	// KeyProvider specifies the key provider type (kms, vault, file)
	KeyProvider string `json:"keyProvider" yaml:"keyProvider"`

	// KeyProviderConfig contains provider-specific configuration
	KeyProviderConfig map[string]interface{} `json:"keyProviderConfig,omitempty" yaml:"keyProviderConfig,omitempty"`

	// RekorURL is the Rekor transparency log URL
	RekorURL string `json:"rekorUrl,omitempty" yaml:"rekorUrl,omitempty"`

	// FulcioURL is the Fulcio certificate authority URL
	FulcioURL string `json:"fulcioUrl,omitempty" yaml:"fulcioUrl,omitempty"`

	// AllowInsecure allows insecure registries
	AllowInsecure bool `json:"allowInsecure,omitempty" yaml:"allowInsecure,omitempty"`

	// UploadToTLog automatically uploads signatures to transparency log
	UploadToTLog bool `json:"uploadToTLog" yaml:"uploadToTLog"`

	// Timeout for signing operations
	Timeout time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`

	// Annotations to add to all signatures
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

// VerificationConfig contains configuration for verification operations
type VerificationConfig struct {
	// Enabled indicates if verification is enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// EnforceVerification fails operations if verification fails
	EnforceVerification bool `json:"enforceVerification" yaml:"enforceVerification"`

	// PublicKeys are trusted public keys (PEM format)
	PublicKeys []string `json:"publicKeys,omitempty" yaml:"publicKeys,omitempty"`

	// TrustRootCertificates are trusted root certificates
	TrustRootCertificates []string `json:"trustRootCertificates,omitempty" yaml:"trustRootCertificates,omitempty"`

	// CertificateIdentities are trusted certificate identities (for keyless)
	CertificateIdentities []string `json:"certificateIdentities,omitempty" yaml:"certificateIdentities,omitempty"`

	// CertificateOIDCIssuers are trusted OIDC issuers (for keyless)
	CertificateOIDCIssuers []string `json:"certificateOIDCIssuers,omitempty" yaml:"certificateOIDCIssuers,omitempty"`

	// RekorURL is the Rekor transparency log URL
	RekorURL string `json:"rekorUrl,omitempty" yaml:"rekorUrl,omitempty"`

	// RequireRekorBundle requires Rekor bundle for verification
	RequireRekorBundle bool `json:"requireRekorBundle" yaml:"requireRekorBundle"`

	// MaxSignatureAge is the maximum age of signatures to accept
	MaxSignatureAge time.Duration `json:"maxSignatureAge,omitempty" yaml:"maxSignatureAge,omitempty"`

	// IgnoreSCT ignores SignedCertificateTimestamp validation
	IgnoreSCT bool `json:"ignoreSCT,omitempty" yaml:"ignoreSCT,omitempty"`

	// IgnoreTLog ignores transparency log verification
	IgnoreTLog bool `json:"ignoreTLog,omitempty" yaml:"ignoreTLog,omitempty"`

	// AllowInsecure allows insecure registries
	AllowInsecure bool `json:"allowInsecure,omitempty" yaml:"allowInsecure,omitempty"`

	// Timeout for verification operations
	Timeout time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`

	// TrustPolicies define trust policies for verification
	TrustPolicies []*TrustPolicy `json:"trustPolicies,omitempty" yaml:"trustPolicies,omitempty"`
}

// KeyInfo contains information about a cryptographic key
type KeyInfo struct {
	// ID is the unique key identifier
	ID string `json:"id"`

	// Type is the key type (RSA, ECDSA, Ed25519)
	Type KeyType `json:"type"`

	// Algorithm is the specific algorithm (e.g., "ECDSA_P256")
	Algorithm string `json:"algorithm"`

	// PublicKey is the public key
	PublicKey crypto.PublicKey `json:"-"`

	// PublicKeyPEM is the PEM-encoded public key
	PublicKeyPEM string `json:"publicKeyPem"`

	// CreatedAt is when the key was created
	CreatedAt time.Time `json:"createdAt"`

	// ExpiresAt is when the key expires (optional)
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`

	// Revoked indicates if the key is revoked
	Revoked bool `json:"revoked"`

	// RevokedAt is when the key was revoked (if applicable)
	RevokedAt *time.Time `json:"revokedAt,omitempty"`

	// RevocationReason is the reason for revocation
	RevocationReason string `json:"revocationReason,omitempty"`

	// Metadata contains additional key metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// KeyType represents the type of cryptographic key
type KeyType string

const (
	KeyTypeRSA     KeyType = "RSA"
	KeyTypeECDSA   KeyType = "ECDSA"
	KeyTypeEd25519 KeyType = "Ed25519"
)

// AttestationType represents the type of attestation
type AttestationType string

const (
	AttestationTypeSLSAProvenance    AttestationType = "slsaprovenance"
	AttestationTypeSPDX              AttestationType = "spdx"
	AttestationTypeCycloneDX         AttestationType = "cyclonedx"
	AttestationTypeVulnerabilityScan AttestationType = "vulnerabilityscan"
	AttestationTypeCustom            AttestationType = "custom"
)

// AttestationResult contains the result of attestation verification
type AttestationResult struct {
	// Verified indicates if attestation was verified
	Verified bool `json:"verified"`

	// AttestationType is the type of attestation
	AttestationType AttestationType `json:"attestationType"`

	// Attestation is the verified attestation
	Attestation *Attestation `json:"attestation"`

	// VerifiedAt is when verification was performed
	VerifiedAt time.Time `json:"verifiedAt"`

	// ValidationErrors contains any validation errors
	ValidationErrors []error `json:"-"`

	// ValidationMessages contains human-readable validation messages
	ValidationMessages []string `json:"validationMessages,omitempty"`
}

// Attestation represents an attestation document
type Attestation struct {
	// Type is the attestation type
	Type AttestationType `json:"type"`

	// PredicateType is the in-toto predicate type
	PredicateType string `json:"predicateType"`

	// Subject identifies what the attestation is about
	Subject []Subject `json:"subject"`

	// Predicate contains the attestation payload
	Predicate map[string]interface{} `json:"predicate"`

	// Signature is the attestation signature
	Signature []byte `json:"signature,omitempty"`

	// SignatureMetadata contains signature metadata
	SignatureMetadata *SignatureMetadata `json:"signatureMetadata,omitempty"`

	// CreatedAt is when the attestation was created
	CreatedAt time.Time `json:"createdAt"`
}

// AttestationMetadata contains metadata about an attestation
type AttestationMetadata struct {
	// ID is the attestation identifier
	ID string `json:"id"`

	// Type is the attestation type
	Type AttestationType `json:"type"`

	// ImageRef is the image the attestation is for
	ImageRef string `json:"imageRef"`

	// Digest is the image digest
	Digest string `json:"digest"`

	// CreatedAt is when the attestation was created
	CreatedAt time.Time `json:"createdAt"`

	// SignedBy identifies who signed the attestation
	SignedBy string `json:"signedBy,omitempty"`
}

// Subject represents an attestation subject
type Subject struct {
	// Name is the subject name (e.g., image reference)
	Name string `json:"name"`

	// Digest is the subject digest
	Digest map[string]string `json:"digest"`
}

// TrustPolicy defines a trust policy for image verification
type TrustPolicy struct {
	// Name is the policy name
	Name string `json:"name" yaml:"name"`

	// ImagePattern is the image pattern to match (glob or regex)
	ImagePattern string `json:"imagePattern" yaml:"imagePattern"`

	// TrustRoots are trusted public keys or certificates
	TrustRoots []string `json:"trustRoots" yaml:"trustRoots"`

	// RequireSignatureCount is the minimum number of signatures required
	RequireSignatureCount int `json:"requireSignatureCount" yaml:"requireSignatureCount"`

	// RequireAttestations specifies required attestations
	RequireAttestations []AttestationType `json:"requireAttestations,omitempty" yaml:"requireAttestations,omitempty"`

	// AllowedIssuers are allowed OIDC issuers (for keyless)
	AllowedIssuers []string `json:"allowedIssuers,omitempty" yaml:"allowedIssuers,omitempty"`

	// AllowedIdentities are allowed certificate identities (for keyless)
	AllowedIdentities []string `json:"allowedIdentities,omitempty" yaml:"allowedIdentities,omitempty"`

	// EnforceTransparencyLog requires transparency log verification
	EnforceTransparencyLog bool `json:"enforceTransparencyLog" yaml:"enforceTransparencyLog"`
}

// PolicyValidationResult contains the result of policy validation
type PolicyValidationResult struct {
	// Valid indicates if the image meets the policy
	Valid bool `json:"valid"`

	// Policy is the policy that was evaluated
	Policy *TrustPolicy `json:"policy"`

	// ImageRef is the image that was validated
	ImageRef string `json:"imageRef"`

	// Violations are policy violations
	Violations []string `json:"violations,omitempty"`

	// ValidationMessages contains detailed validation messages
	ValidationMessages []string `json:"validationMessages,omitempty"`

	// ValidatedAt is when validation was performed
	ValidatedAt time.Time `json:"validatedAt"`
}

// RekorBundle contains a Rekor transparency log bundle
type RekorBundle struct {
	// SignedEntryTimestamp is the signed timestamp from Rekor
	SignedEntryTimestamp []byte `json:"signedEntryTimestamp"`

	// Payload is the Rekor entry payload
	Payload RekorPayload `json:"payload"`
}

// RekorPayload contains the Rekor entry payload
type RekorPayload struct {
	// Body is the entry body
	Body string `json:"body"`

	// IntegratedTime is the log entry timestamp
	IntegratedTime int64 `json:"integratedTime"`

	// LogIndex is the log entry index
	LogIndex int64 `json:"logIndex"`

	// LogID is the log identifier
	LogID string `json:"logID"`
}

// LogEntry represents a transparency log entry
type LogEntry struct {
	// Kind is the entry kind (e.g., "hashedrekord")
	Kind string `json:"kind"`

	// APIVersion is the entry schema version
	APIVersion string `json:"apiVersion"`

	// Spec contains the entry specification
	Spec map[string]interface{} `json:"spec"`
}

// LogEntryResult represents a transparency log entry result
type LogEntryResult struct {
	// UUID is the entry unique identifier
	UUID string `json:"uuid"`

	// Body is the entry body
	Body *LogEntry `json:"body"`

	// IntegratedTime is when the entry was integrated
	IntegratedTime int64 `json:"integratedTime"`

	// LogIndex is the log index
	LogIndex int64 `json:"logIndex"`

	// LogID is the log identifier
	LogID string `json:"logID"`

	// Verification contains verification information
	Verification *LogEntryVerification `json:"verification,omitempty"`
}

// LogEntryVerification contains log entry verification data
type LogEntryVerification struct {
	// SignedEntryTimestamp is the signed timestamp
	SignedEntryTimestamp []byte `json:"signedEntryTimestamp"`

	// InclusionProof is the inclusion proof
	InclusionProof *InclusionProof `json:"inclusionProof,omitempty"`
}

// InclusionProof represents a Merkle tree inclusion proof
type InclusionProof struct {
	// TreeSize is the tree size
	TreeSize int64 `json:"treeSize"`

	// RootHash is the tree root hash
	RootHash []byte `json:"rootHash"`

	// LogIndex is the log index
	LogIndex int64 `json:"logIndex"`

	// Hashes are the proof hashes
	Hashes [][]byte `json:"hashes"`
}

// SearchQuery represents a transparency log search query
type SearchQuery struct {
	// Email searches by email
	Email string `json:"email,omitempty"`

	// Hash searches by artifact hash
	Hash string `json:"hash,omitempty"`

	// PublicKey searches by public key
	PublicKey crypto.PublicKey `json:"-"`

	// LogIndex searches by log index
	LogIndex *int64 `json:"logIndex,omitempty"`
}

// Policy represents a verification policy
type Policy struct {
	// Name is the policy name
	Name string `json:"name"`

	// Version is the policy version
	Version string `json:"version"`

	// Type is the policy type (rego, cel, custom)
	Type string `json:"type"`

	// Content is the policy content
	Content string `json:"content"`

	// Metadata contains policy metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// PolicySubject represents the subject of policy evaluation
type PolicySubject struct {
	// ImageRef is the image reference
	ImageRef string `json:"imageRef"`

	// Digest is the image digest
	Digest string `json:"digest"`

	// Signatures are the image signatures
	Signatures []*SignatureMetadata `json:"signatures"`

	// Attestations are the image attestations
	Attestations []*Attestation `json:"attestations"`

	// Metadata contains additional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PolicyResult represents the result of policy evaluation
type PolicyResult struct {
	// Allowed indicates if the policy allows the action
	Allowed bool `json:"allowed"`

	// Violations are policy violations
	Violations []string `json:"violations,omitempty"`

	// Warnings are policy warnings
	Warnings []string `json:"warnings,omitempty"`

	// Metadata contains additional result metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PolicySource represents a policy source
type PolicySource struct {
	// Type is the source type (file, url, inline)
	Type string `json:"type"`

	// Location is the source location
	Location string `json:"location"`

	// Content is the inline content (if Type is "inline")
	Content string `json:"content,omitempty"`
}

// PolicyInfo contains information about a policy
type PolicyInfo struct {
	// Name is the policy name
	Name string `json:"name"`

	// Version is the policy version
	Version string `json:"version"`

	// Type is the policy type
	Type string `json:"type"`

	// LoadedAt is when the policy was loaded
	LoadedAt time.Time `json:"loadedAt"`

	// Source is the policy source
	Source PolicySource `json:"source"`
}

// PolicyTestCase represents a policy test case
type PolicyTestCase struct {
	// Name is the test case name
	Name string `json:"name"`

	// Subject is the test subject
	Subject *PolicySubject `json:"subject"`

	// ExpectedResult is the expected result
	ExpectedResult *PolicyResult `json:"expectedResult"`
}

// PolicyTestResult represents a policy test result
type PolicyTestResult struct {
	// TestCase is the test case
	TestCase *PolicyTestCase `json:"testCase"`

	// Passed indicates if the test passed
	Passed bool `json:"passed"`

	// ActualResult is the actual result
	ActualResult *PolicyResult `json:"actualResult"`

	// Differences are differences between expected and actual
	Differences []string `json:"differences,omitempty"`
}
