// Package mtls provides interfaces and types for mutual TLS authentication
// and certificate management in the Freightliner platform.
package mtls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
)

// TLSProvider defines the interface for managing TLS certificates and configurations.
// Implementations should handle certificate lifecycle, rotation, and validation.
type TLSProvider interface {
	// GetTLSConfig returns a configured tls.Config for establishing secure connections.
	// The config includes certificates, cipher suites, and protocol versions.
	GetTLSConfig(ctx context.Context) (*tls.Config, error)

	// LoadCertificate loads a certificate from the configured source (file, vault, HSM).
	// Returns the certificate info including expiry and validation details.
	LoadCertificate(ctx context.Context, certID string) (*CertificateInfo, error)

	// ValidateCertificate verifies a certificate against trust anchors and policies.
	// Checks expiry, revocation status, and chain of trust.
	ValidateCertificate(ctx context.Context, cert *x509.Certificate) error

	// GetCACertPool returns the certificate authority pool for validating peer certificates.
	GetCACertPool(ctx context.Context) (*x509.CertPool, error)

	// RefreshCertificates forces a refresh of all cached certificates.
	// Useful for handling certificate rotation events.
	RefreshCertificates(ctx context.Context) error
}

// MutualTLSAuthenticator defines the interface for mutual TLS authentication flows.
// Handles both client and server-side mTLS authentication with identity verification.
type MutualTLSAuthenticator interface {
	// AuthenticateClient verifies a client certificate and extracts identity information.
	// Returns the authenticated identity or an error if authentication fails.
	AuthenticateClient(ctx context.Context, cert *x509.Certificate) (*AuthenticatedIdentity, error)

	// AuthenticateServer verifies a server certificate during client connection.
	// Ensures the server identity matches expected values and certificate is valid.
	AuthenticateServer(ctx context.Context, cert *x509.Certificate, serverName string) error

	// VerifyPeerCertificate is a callback for custom certificate verification.
	// Can be used with tls.Config.VerifyPeerCertificate for advanced validation.
	VerifyPeerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error

	// ExtractIdentity extracts identity claims from a certificate's subject and extensions.
	// Returns structured identity information for authorization decisions.
	ExtractIdentity(cert *x509.Certificate) (*Identity, error)

	// ValidateIdentity checks if an identity meets policy requirements.
	// Verifies roles, permissions, and organizational constraints.
	ValidateIdentity(ctx context.Context, identity *Identity) error
}

// CertificateRotator defines the interface for automated certificate rotation.
// Handles proactive certificate renewal before expiry to ensure zero-downtime rotation.
type CertificateRotator interface {
	// StartRotation begins the certificate rotation process with the given policy.
	// Runs in the background and triggers rotation based on policy rules.
	StartRotation(ctx context.Context, policy *RotationPolicy) error

	// StopRotation halts the certificate rotation process gracefully.
	StopRotation(ctx context.Context) error

	// RotateNow forces an immediate certificate rotation.
	// Useful for emergency rotation or manual policy override.
	RotateNow(ctx context.Context, certID string) error

	// GetRotationStatus returns the current rotation status for a certificate.
	// Includes next rotation time, last rotation, and any pending operations.
	GetRotationStatus(ctx context.Context, certID string) (*RotationStatus, error)

	// RegisterRotationCallback registers a callback to be invoked when rotation occurs.
	// Allows applications to respond to certificate updates (e.g., reload configs).
	RegisterRotationCallback(callback RotationCallback) error

	// UnregisterRotationCallback removes a previously registered callback.
	UnregisterRotationCallback(callbackID string) error
}

// CertificateManager provides high-level certificate management operations.
// Combines TLSProvider, MutualTLSAuthenticator, and CertificateRotator functionality.
type CertificateManager interface {
	TLSProvider
	MutualTLSAuthenticator
	CertificateRotator

	// IssueCertificate requests a new certificate from the certificate authority.
	// Returns the issued certificate info including private key reference.
	IssueCertificate(ctx context.Context, req *CertificateRequest) (*CertificateInfo, error)

	// RevokeCertificate revokes a certificate before its expiry.
	// Updates CRL/OCSP and notifies dependent services.
	RevokeCertificate(ctx context.Context, certID string, reason RevocationReason) error

	// ListCertificates returns all certificates managed by this instance.
	// Useful for inventory, audit, and compliance reporting.
	ListCertificates(ctx context.Context) ([]*CertificateInfo, error)

	// ExportCertificate exports a certificate in the specified format.
	// Supports PEM, DER, PKCS12 formats for interoperability.
	ExportCertificate(ctx context.Context, certID string, format ExportFormat) ([]byte, error)
}

// CertificateStore defines the interface for persistent certificate storage.
// Implementations can use file systems, secret managers, or HSMs.
type CertificateStore interface {
	// Store saves a certificate and its private key securely.
	Store(ctx context.Context, certID string, cert *x509.Certificate, privateKey interface{}) error

	// Retrieve loads a certificate and its private key from storage.
	Retrieve(ctx context.Context, certID string) (*x509.Certificate, interface{}, error)

	// Delete removes a certificate and its private key from storage.
	Delete(ctx context.Context, certID string) error

	// List returns all certificate IDs in the store.
	List(ctx context.Context) ([]string, error)

	// Exists checks if a certificate exists in the store.
	Exists(ctx context.Context, certID string) (bool, error)
}

// TrustManager defines the interface for managing trust anchors and certificate chains.
type TrustManager interface {
	// AddTrustAnchor adds a CA certificate to the trust store.
	AddTrustAnchor(ctx context.Context, ca *x509.Certificate) error

	// RemoveTrustAnchor removes a CA certificate from the trust store.
	RemoveTrustAnchor(ctx context.Context, caID string) error

	// GetTrustAnchors returns all trusted CA certificates.
	GetTrustAnchors(ctx context.Context) ([]*x509.Certificate, error)

	// VerifyChain verifies a certificate chain against trusted CAs.
	VerifyChain(ctx context.Context, cert *x509.Certificate, intermediates []*x509.Certificate) error

	// UpdateCRL updates the certificate revocation list.
	UpdateCRL(ctx context.Context, crl *x509.RevocationList) error

	// CheckRevocation checks if a certificate is revoked via CRL or OCSP.
	CheckRevocation(ctx context.Context, cert *x509.Certificate) (bool, error)
}

// RotationCallback is invoked when certificate rotation occurs.
type RotationCallback func(ctx context.Context, event *RotationEvent) error

// RevocationReason represents reasons for certificate revocation.
type RevocationReason int

const (
	// RevocationReasonUnspecified indicates no specific reason provided.
	RevocationReasonUnspecified RevocationReason = iota

	// RevocationReasonKeyCompromise indicates the private key was compromised.
	RevocationReasonKeyCompromise

	// RevocationReasonCACompromise indicates the CA was compromised.
	RevocationReasonCACompromise

	// RevocationReasonAffiliationChanged indicates entity affiliation changed.
	RevocationReasonAffiliationChanged

	// RevocationReasonSuperseded indicates certificate was replaced.
	RevocationReasonSuperseded

	// RevocationReasonCessationOfOperation indicates entity ceased operations.
	RevocationReasonCessationOfOperation

	// RevocationReasonCertificateHold indicates temporary revocation.
	RevocationReasonCertificateHold

	// RevocationReasonRemoveFromCRL indicates removal from CRL.
	RevocationReasonRemoveFromCRL

	// RevocationReasonPrivilegeWithdrawn indicates privileges revoked.
	RevocationReasonPrivilegeWithdrawn

	// RevocationReasonAACompromise indicates attribute authority compromised.
	RevocationReasonAACompromise
)

// ExportFormat represents certificate export formats.
type ExportFormat int

const (
	// ExportFormatPEM exports certificate in PEM format.
	ExportFormatPEM ExportFormat = iota

	// ExportFormatDER exports certificate in DER format.
	ExportFormatDER

	// ExportFormatPKCS12 exports certificate in PKCS12 format.
	ExportFormatPKCS12

	// ExportFormatJWK exports certificate as JSON Web Key.
	ExportFormatJWK
)

// String returns the string representation of RevocationReason.
func (r RevocationReason) String() string {
	reasons := []string{
		"Unspecified",
		"KeyCompromise",
		"CACompromise",
		"AffiliationChanged",
		"Superseded",
		"CessationOfOperation",
		"CertificateHold",
		"RemoveFromCRL",
		"PrivilegeWithdrawn",
		"AACompromise",
	}
	if r >= 0 && int(r) < len(reasons) {
		return reasons[r]
	}
	return "Unknown"
}

// String returns the string representation of ExportFormat.
func (f ExportFormat) String() string {
	formats := []string{"PEM", "DER", "PKCS12", "JWK"}
	if f >= 0 && int(f) < len(formats) {
		return formats[f]
	}
	return "Unknown"
}
