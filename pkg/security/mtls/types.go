package mtls

import (
	"crypto/x509"
	"time"
)

// TLSConfig represents the configuration for TLS connections.
type TLSConfig struct {
	// CertificatePath is the file path to the certificate.
	CertificatePath string `json:"certificate_path" yaml:"certificate_path"`

	// PrivateKeyPath is the file path to the private key.
	PrivateKeyPath string `json:"private_key_path" yaml:"private_key_path"`

	// CAPath is the file path to the CA certificate bundle.
	CAPath string `json:"ca_path" yaml:"ca_path"`

	// MinVersion is the minimum TLS version to accept (e.g., "1.2", "1.3").
	MinVersion string `json:"min_version" yaml:"min_version"`

	// MaxVersion is the maximum TLS version to accept.
	MaxVersion string `json:"max_version,omitempty" yaml:"max_version,omitempty"`

	// CipherSuites is the list of cipher suites to use.
	CipherSuites []string `json:"cipher_suites,omitempty" yaml:"cipher_suites,omitempty"`

	// ClientAuth specifies the client authentication policy.
	// Values: "none", "request", "require", "verify", "require-and-verify"
	ClientAuth string `json:"client_auth" yaml:"client_auth"`

	// ServerName is the expected server name for SNI.
	ServerName string `json:"server_name,omitempty" yaml:"server_name,omitempty"`

	// InsecureSkipVerify disables certificate verification (DO NOT USE IN PRODUCTION).
	InsecureSkipVerify bool `json:"insecure_skip_verify,omitempty" yaml:"insecure_skip_verify,omitempty"`

	// RotationEnabled enables automatic certificate rotation.
	RotationEnabled bool `json:"rotation_enabled" yaml:"rotation_enabled"`

	// RotationPolicy defines the certificate rotation policy.
	RotationPolicy *RotationPolicy `json:"rotation_policy,omitempty" yaml:"rotation_policy,omitempty"`

	// UseHardwareSecurityModule enables HSM for private key operations.
	UseHardwareSecurityModule bool `json:"use_hsm,omitempty" yaml:"use_hsm,omitempty"`

	// HSMConfig contains HSM-specific configuration.
	HSMConfig *HSMConfig `json:"hsm_config,omitempty" yaml:"hsm_config,omitempty"`

	// VaultEnabled enables HashiCorp Vault integration for certificate management.
	VaultEnabled bool `json:"vault_enabled,omitempty" yaml:"vault_enabled,omitempty"`

	// VaultConfig contains Vault-specific configuration.
	VaultConfig *VaultConfig `json:"vault_config,omitempty" yaml:"vault_config,omitempty"`
}

// CertificateInfo contains information about a certificate.
type CertificateInfo struct {
	// ID is the unique identifier for this certificate.
	ID string `json:"id"`

	// Subject contains the certificate subject information.
	Subject *Subject `json:"subject"`

	// Issuer contains the certificate issuer information.
	Issuer *Subject `json:"issuer"`

	// SerialNumber is the certificate serial number.
	SerialNumber string `json:"serial_number"`

	// NotBefore is the certificate validity start time.
	NotBefore time.Time `json:"not_before"`

	// NotAfter is the certificate validity end time.
	NotAfter time.Time `json:"not_after"`

	// Fingerprint is the SHA-256 fingerprint of the certificate.
	Fingerprint string `json:"fingerprint"`

	// KeyUsage describes the key usage extensions.
	KeyUsage []string `json:"key_usage"`

	// ExtendedKeyUsage describes the extended key usage extensions.
	ExtendedKeyUsage []string `json:"extended_key_usage"`

	// DNSNames contains the Subject Alternative Names (DNS).
	DNSNames []string `json:"dns_names,omitempty"`

	// IPAddresses contains the Subject Alternative Names (IP).
	IPAddresses []string `json:"ip_addresses,omitempty"`

	// IsCA indicates if this is a CA certificate.
	IsCA bool `json:"is_ca"`

	// IsSelfSigned indicates if this certificate is self-signed.
	IsSelfSigned bool `json:"is_self_signed"`

	// Certificate is the actual x509 certificate.
	Certificate *x509.Certificate `json:"-"`

	// Status is the current certificate status.
	Status CertificateStatus `json:"status"`

	// CreatedAt is when this certificate was created/imported.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when this certificate was last updated.
	UpdatedAt time.Time `json:"updated_at"`

	// Tags are user-defined tags for certificate organization.
	Tags map[string]string `json:"tags,omitempty"`
}

// Subject represents certificate subject or issuer information.
type Subject struct {
	// CommonName is the common name (CN).
	CommonName string `json:"common_name"`

	// Organization is the organization (O).
	Organization []string `json:"organization,omitempty"`

	// OrganizationalUnit is the organizational unit (OU).
	OrganizationalUnit []string `json:"organizational_unit,omitempty"`

	// Country is the country (C).
	Country []string `json:"country,omitempty"`

	// Province is the province/state (ST).
	Province []string `json:"province,omitempty"`

	// Locality is the locality/city (L).
	Locality []string `json:"locality,omitempty"`

	// StreetAddress is the street address.
	StreetAddress []string `json:"street_address,omitempty"`

	// PostalCode is the postal code.
	PostalCode []string `json:"postal_code,omitempty"`

	// SerialNumber is the subject serial number.
	SerialNumber string `json:"serial_number,omitempty"`
}

// RotationPolicy defines when and how certificates should be rotated.
type RotationPolicy struct {
	// Enabled indicates if rotation is enabled.
	Enabled bool `json:"enabled" yaml:"enabled"`

	// RotateBeforeExpiry is the duration before expiry to trigger rotation.
	// Example: "720h" (30 days before expiry)
	RotateBeforeExpiry time.Duration `json:"rotate_before_expiry" yaml:"rotate_before_expiry"`

	// CheckInterval is how often to check for rotation needs.
	// Example: "1h"
	CheckInterval time.Duration `json:"check_interval" yaml:"check_interval"`

	// AutoRenew enables automatic certificate renewal with CA.
	AutoRenew bool `json:"auto_renew" yaml:"auto_renew"`

	// RenewalMethod specifies how to renew certificates.
	// Values: "acme", "vault", "ca_api", "manual"
	RenewalMethod string `json:"renewal_method" yaml:"renewal_method"`

	// NotifyBeforeExpiry sends notifications before expiry.
	NotifyBeforeExpiry time.Duration `json:"notify_before_expiry,omitempty" yaml:"notify_before_expiry,omitempty"`

	// NotificationChannels defines where to send notifications.
	NotificationChannels []string `json:"notification_channels,omitempty" yaml:"notification_channels,omitempty"`

	// MaxRetries is the maximum number of rotation retry attempts.
	MaxRetries int `json:"max_retries,omitempty" yaml:"max_retries,omitempty"`

	// RetryInterval is the delay between rotation retry attempts.
	RetryInterval time.Duration `json:"retry_interval,omitempty" yaml:"retry_interval,omitempty"`

	// GracePeriod is the overlap time where both old and new certs are valid.
	GracePeriod time.Duration `json:"grace_period,omitempty" yaml:"grace_period,omitempty"`
}

// RotationStatus represents the current rotation status of a certificate.
type RotationStatus struct {
	// CertificateID is the ID of the certificate.
	CertificateID string `json:"certificate_id"`

	// LastRotation is when the certificate was last rotated.
	LastRotation *time.Time `json:"last_rotation,omitempty"`

	// NextRotation is when the next rotation is scheduled.
	NextRotation *time.Time `json:"next_rotation,omitempty"`

	// Status is the current rotation state.
	Status string `json:"status"` // "scheduled", "in_progress", "completed", "failed"

	// Error contains any rotation error message.
	Error string `json:"error,omitempty"`

	// Attempts is the number of rotation attempts made.
	Attempts int `json:"attempts"`

	// ExpiresAt is when the current certificate expires.
	ExpiresAt time.Time `json:"expires_at"`

	// DaysUntilExpiry is the number of days until expiry.
	DaysUntilExpiry int `json:"days_until_expiry"`

	// RotationRequired indicates if rotation is needed now.
	RotationRequired bool `json:"rotation_required"`
}

// RotationEvent represents a certificate rotation event.
type RotationEvent struct {
	// CertificateID is the ID of the rotated certificate.
	CertificateID string `json:"certificate_id"`

	// EventType is the type of rotation event.
	// Values: "scheduled", "started", "completed", "failed", "revoked"
	EventType string `json:"event_type"`

	// OldCertificate is the previous certificate info.
	OldCertificate *CertificateInfo `json:"old_certificate,omitempty"`

	// NewCertificate is the new certificate info.
	NewCertificate *CertificateInfo `json:"new_certificate,omitempty"`

	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`

	// Message contains additional event information.
	Message string `json:"message,omitempty"`

	// Error contains any error information.
	Error string `json:"error,omitempty"`
}

// CertificateRequest represents a request to issue a new certificate.
type CertificateRequest struct {
	// CommonName is the certificate CN.
	CommonName string `json:"common_name"`

	// Organization is the organization name.
	Organization string `json:"organization,omitempty"`

	// OrganizationalUnit is the organizational unit.
	OrganizationalUnit string `json:"organizational_unit,omitempty"`

	// Country is the country code.
	Country string `json:"country,omitempty"`

	// Province is the province/state.
	Province string `json:"province,omitempty"`

	// Locality is the locality/city.
	Locality string `json:"locality,omitempty"`

	// DNSNames are the Subject Alternative Names (DNS).
	DNSNames []string `json:"dns_names,omitempty"`

	// IPAddresses are the Subject Alternative Names (IP).
	IPAddresses []string `json:"ip_addresses,omitempty"`

	// KeyType is the key type ("rsa", "ecdsa", "ed25519").
	KeyType string `json:"key_type"`

	// KeySize is the key size in bits (for RSA).
	KeySize int `json:"key_size,omitempty"`

	// Curve is the elliptic curve name (for ECDSA).
	Curve string `json:"curve,omitempty"`

	// ValidityDuration is how long the certificate should be valid.
	ValidityDuration time.Duration `json:"validity_duration"`

	// IsCA indicates if this should be a CA certificate.
	IsCA bool `json:"is_ca,omitempty"`

	// KeyUsage specifies the key usage extensions.
	KeyUsage []string `json:"key_usage,omitempty"`

	// ExtendedKeyUsage specifies the extended key usage extensions.
	ExtendedKeyUsage []string `json:"extended_key_usage,omitempty"`
}

// AuthenticatedIdentity represents an authenticated identity from a certificate.
type AuthenticatedIdentity struct {
	// ID is the unique identity identifier.
	ID string `json:"id"`

	// CommonName is the certificate CN.
	CommonName string `json:"common_name"`

	// Organization is the organization.
	Organization string `json:"organization,omitempty"`

	// OrganizationalUnit is the organizational unit.
	OrganizationalUnit string `json:"organizational_unit,omitempty"`

	// Roles are the roles assigned to this identity.
	Roles []string `json:"roles,omitempty"`

	// Permissions are the permissions granted to this identity.
	Permissions []string `json:"permissions,omitempty"`

	// Attributes are custom identity attributes.
	Attributes map[string]string `json:"attributes,omitempty"`

	// CertificateFingerprint is the certificate fingerprint.
	CertificateFingerprint string `json:"certificate_fingerprint"`

	// AuthenticatedAt is when the authentication occurred.
	AuthenticatedAt time.Time `json:"authenticated_at"`

	// ExpiresAt is when the authentication expires.
	ExpiresAt time.Time `json:"expires_at"`
}

// Identity represents an identity extracted from a certificate.
type Identity struct {
	// Subject contains the certificate subject.
	Subject *Subject `json:"subject"`

	// SerialNumber is the certificate serial number.
	SerialNumber string `json:"serial_number"`

	// Fingerprint is the certificate fingerprint.
	Fingerprint string `json:"fingerprint"`

	// Claims are custom claims extracted from certificate extensions.
	Claims map[string]interface{} `json:"claims,omitempty"`

	// Verified indicates if the identity has been verified.
	Verified bool `json:"verified"`
}

// CertificateStatus represents the status of a certificate.
type CertificateStatus string

const (
	// CertificateStatusActive indicates the certificate is active and valid.
	CertificateStatusActive CertificateStatus = "active"

	// CertificateStatusExpired indicates the certificate has expired.
	CertificateStatusExpired CertificateStatus = "expired"

	// CertificateStatusRevoked indicates the certificate has been revoked.
	CertificateStatusRevoked CertificateStatus = "revoked"

	// CertificateStatusPending indicates the certificate is pending issuance.
	CertificateStatusPending CertificateStatus = "pending"

	// CertificateStatusRotating indicates the certificate is being rotated.
	CertificateStatusRotating CertificateStatus = "rotating"
)

// HSMConfig contains Hardware Security Module configuration.
type HSMConfig struct {
	// Provider is the HSM provider ("pkcs11", "awskms", "azurekeyvault", "gcpkms").
	Provider string `json:"provider" yaml:"provider"`

	// LibraryPath is the path to the PKCS11 library.
	LibraryPath string `json:"library_path,omitempty" yaml:"library_path,omitempty"`

	// Pin is the HSM PIN/password.
	Pin string `json:"pin,omitempty" yaml:"pin,omitempty"`

	// SlotID is the HSM slot ID.
	SlotID int `json:"slot_id,omitempty" yaml:"slot_id,omitempty"`

	// KeyID is the key identifier in the HSM.
	KeyID string `json:"key_id,omitempty" yaml:"key_id,omitempty"`

	// Region is the cloud provider region (for cloud HSMs).
	Region string `json:"region,omitempty" yaml:"region,omitempty"`

	// Endpoint is the HSM endpoint URL.
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
}

// VaultConfig contains HashiCorp Vault configuration.
type VaultConfig struct {
	// Address is the Vault server address.
	Address string `json:"address" yaml:"address"`

	// Token is the Vault authentication token.
	Token string `json:"token,omitempty" yaml:"token,omitempty"`

	// RoleID is the AppRole role ID.
	RoleID string `json:"role_id,omitempty" yaml:"role_id,omitempty"`

	// SecretID is the AppRole secret ID.
	SecretID string `json:"secret_id,omitempty" yaml:"secret_id,omitempty"`

	// PKIMount is the Vault PKI mount path.
	PKIMount string `json:"pki_mount" yaml:"pki_mount"`

	// Role is the Vault PKI role to use.
	Role string `json:"role" yaml:"role"`

	// Namespace is the Vault namespace (Vault Enterprise).
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`

	// CAPath is the path to Vault CA certificate.
	CAPath string `json:"ca_path,omitempty" yaml:"ca_path,omitempty"`

	// TLSSkipVerify disables TLS verification (DO NOT USE IN PRODUCTION).
	TLSSkipVerify bool `json:"tls_skip_verify,omitempty" yaml:"tls_skip_verify,omitempty"`
}
