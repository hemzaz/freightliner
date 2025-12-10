//go:build cosign
// +build cosign

package cosign

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/sigstore/cosign/v2/pkg/cosign/bundle"
	"github.com/sigstore/cosign/v2/pkg/oci"
	ociremote "github.com/sigstore/cosign/v2/pkg/oci/remote"
	"github.com/sigstore/cosign/v2/pkg/signature"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
	"github.com/sigstore/sigstore/pkg/signature/options"
)

// Signature represents a verified signature with metadata
type Signature struct {
	Digest      string
	Certificate *x509.Certificate
	Chain       []*x509.Certificate
	Payload     []byte
	Signature   []byte
	Bundle      *bundle.RekorBundle

	// OIDC identity (for keyless)
	Issuer  string
	Subject string
}

// Attestation represents a verified attestation (e.g., SLSA provenance)
type Attestation struct {
	PredicateType string
	Payload       []byte
	Signature     *Signature
}

// VerifierConfig holds configuration for signature verification
type VerifierConfig struct {
	// Public key path (PEM format)
	PublicKeyPath string

	// Public key bytes (if already loaded)
	PublicKey []byte

	// Keyless verification settings
	EnableKeyless bool
	RekorURL      string
	FulcioURL     string

	// Policy configuration
	Policy *Policy

	// Remote options for registry access
	RemoteOpts []remote.Option
}

// Verifier handles Cosign signature and attestation verification
type Verifier struct {
	config      *VerifierConfig
	rekorClient *RekorClient
	policy      *Policy
	verifiers   []signature.Verifier
}

// NewVerifier creates a new Cosign verifier with the given configuration
func NewVerifier(config *VerifierConfig) (*Verifier, error) {
	if config == nil {
		return nil, fmt.Errorf("verifier config is required")
	}

	v := &Verifier{
		config: config,
		policy: config.Policy,
	}

	// Initialize Rekor client if transparency log verification is needed
	if config.EnableKeyless || (config.Policy != nil && config.Policy.RequireRekor) {
		rekorURL := config.RekorURL
		if rekorURL == "" {
			rekorURL = "https://rekor.sigstore.dev" // Default public instance
		}
		v.rekorClient = NewRekorClient(rekorURL)
	}

	// Load verifiers based on configuration
	if err := v.loadVerifiers(); err != nil {
		return nil, fmt.Errorf("failed to load verifiers: %w", err)
	}

	return v, nil
}

// loadVerifiers initializes signature verifiers from public keys
func (v *Verifier) loadVerifiers() error {
	// Load from public key if provided
	if v.config.PublicKeyPath != "" {
		keyBytes, err := os.ReadFile(v.config.PublicKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read public key: %w", err)
		}
		v.config.PublicKey = keyBytes
	}

	if len(v.config.PublicKey) > 0 {
		verifier, err := v.loadPublicKeyVerifier(v.config.PublicKey)
		if err != nil {
			return fmt.Errorf("failed to load public key verifier: %w", err)
		}
		v.verifiers = append(v.verifiers, verifier)
	}

	return nil
}

// loadPublicKeyVerifier creates a verifier from PEM-encoded public key
func (v *Verifier) loadPublicKeyVerifier(keyBytes []byte) (signature.Verifier, error) {
	// Decode PEM block
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Parse public key
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	// Create verifier based on key type
	switch key := pub.(type) {
	case *ecdsa.PublicKey:
		return signature.LoadECDSAVerifier(key, crypto.SHA256)
	default:
		return nil, fmt.Errorf("unsupported public key type: %T", pub)
	}
}

// Verify checks the signature of a container image
func (v *Verifier) Verify(ctx context.Context, ref name.Reference) ([]Signature, error) {
	// Get signatures from registry
	sigs, err := v.getSignatures(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to get signatures: %w", err)
	}

	if len(sigs) == 0 {
		return nil, fmt.Errorf("no signatures found for image %s", ref.String())
	}

	// Verify each signature
	var verified []Signature
	for _, sig := range sigs {
		verifiedSig, err := v.verifySignature(ctx, ref, sig)
		if err != nil {
			// Log error but continue checking other signatures
			fmt.Fprintf(os.Stderr, "Warning: signature verification failed: %v\n", err)
			continue
		}
		verified = append(verified, *verifiedSig)
	}

	if len(verified) == 0 {
		return nil, fmt.Errorf("no valid signatures found for image %s", ref.String())
	}

	// Evaluate policy if configured
	if v.policy != nil {
		if err := v.policy.Evaluate(ctx, verified); err != nil {
			return nil, fmt.Errorf("policy evaluation failed: %w", err)
		}
	}

	return verified, nil
}

// getSignatures retrieves signatures from the registry
func (v *Verifier) getSignatures(ctx context.Context, ref name.Reference) ([]oci.Signature, error) {
	opts := []ociremote.Option{
		ociremote.WithRemoteOptions(v.config.RemoteOpts...),
	}

	se, err := ociremote.SignedEntity(ref, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get signed entity: %w", err)
	}

	sigs, err := se.Signatures()
	if err != nil {
		return nil, fmt.Errorf("failed to get signatures: %w", err)
	}

	sigsList, err := sigs.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve signatures: %w", err)
	}

	return sigsList, nil
}

// verifySignature verifies a single signature
func (v *Verifier) verifySignature(ctx context.Context, ref name.Reference, sig oci.Signature) (*Signature, error) {
	// Get signature payload and base64 signature
	payload, err := sig.Payload()
	if err != nil {
		return nil, fmt.Errorf("failed to get signature payload: %w", err)
	}

	sigBytes, err := sig.Base64Signature()
	if err != nil {
		return nil, fmt.Errorf("failed to get signature bytes: %w", err)
	}

	result := &Signature{
		Payload:   payload,
		Signature: []byte(sigBytes),
	}

	// Try keyless verification first if enabled
	if v.config.EnableKeyless {
		cert, err := sig.Cert()
		if err == nil && cert != nil {
			// This is a keyless signature
			chain, err := sig.Chain()
			if err != nil {
				return nil, fmt.Errorf("failed to get certificate chain: %w", err)
			}

			result.Certificate = cert
			result.Chain = chain

			// Extract OIDC identity
			result.Issuer = extractIssuer(cert)
			result.Subject = extractSubject(cert)

			// Verify with Fulcio certificate
			if err := v.verifyKeyless(ctx, result, sig); err != nil {
				return nil, fmt.Errorf("keyless verification failed: %w", err)
			}

			return result, nil
		}
	}

	// Try public key verification
	if len(v.verifiers) > 0 {
		if err := v.verifyWithPublicKey(ctx, result); err != nil {
			return nil, fmt.Errorf("public key verification failed: %w", err)
		}
		return result, nil
	}

	return nil, fmt.Errorf("no verification method available")
}

// verifyKeyless performs keyless signature verification using Fulcio and Rekor
func (v *Verifier) verifyKeyless(ctx context.Context, sig *Signature, ociSig oci.Signature) error {
	// Get Rekor bundle
	bundle, err := ociSig.Bundle()
	if err != nil {
		return fmt.Errorf("failed to get Rekor bundle: %w", err)
	}
	sig.Bundle = bundle

	// Verify certificate chain
	if err := v.verifyCertificateChain(sig.Certificate, sig.Chain); err != nil {
		return fmt.Errorf("certificate chain verification failed: %w", err)
	}

	// Verify against Rekor transparency log
	if v.rekorClient != nil && bundle != nil {
		if err := v.rekorClient.VerifyBundle(ctx, bundle, sig.Payload); err != nil {
			return fmt.Errorf("Rekor bundle verification failed: %w", err)
		}
	}

	// Verify signature using certificate's public key
	pubKey := sig.Certificate.PublicKey
	verifier, err := signature.LoadVerifier(pubKey, crypto.SHA256)
	if err != nil {
		return fmt.Errorf("failed to load verifier from certificate: %w", err)
	}

	if err := verifier.VerifySignature(bytes.NewReader(sig.Signature), bytes.NewReader(sig.Payload)); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// verifyWithPublicKey verifies signature using configured public keys
func (v *Verifier) verifyWithPublicKey(ctx context.Context, sig *Signature) error {
	var lastErr error
	for _, verifier := range v.verifiers {
		err := verifier.VerifySignature(
			bytes.NewReader(sig.Signature),
			bytes.NewReader(sig.Payload),
			options.WithContext(ctx),
		)
		if err == nil {
			return nil // Success with this verifier
		}
		lastErr = err
	}

	if lastErr != nil {
		return fmt.Errorf("all verifiers failed, last error: %w", lastErr)
	}
	return fmt.Errorf("no verifiers available")
}

// verifyCertificateChain verifies the X.509 certificate chain
func (v *Verifier) verifyCertificateChain(cert *x509.Certificate, chain []*x509.Certificate) error {
	// Build certificate pool with chain
	roots := x509.NewCertPool()
	intermediates := x509.NewCertPool()

	for _, c := range chain {
		if c.IsCA {
			roots.AddCert(c)
		} else {
			intermediates.AddCert(c)
		}
	}

	// Verify certificate
	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
	}

	if _, err := cert.Verify(opts); err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	return nil
}

// VerifyAttestation checks SLSA provenance attestations
func (v *Verifier) VerifyAttestation(ctx context.Context, ref name.Reference) ([]Attestation, error) {
	opts := []ociremote.Option{
		ociremote.WithRemoteOptions(v.config.RemoteOpts...),
	}

	se, err := ociremote.SignedEntity(ref, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get signed entity: %w", err)
	}

	atts, err := se.Attestations()
	if err != nil {
		return nil, fmt.Errorf("failed to get attestations: %w", err)
	}

	attsList, err := atts.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve attestations: %w", err)
	}

	var verified []Attestation
	for _, att := range attsList {
		verifiedAtt, err := v.verifyAttestation(ctx, ref, att)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: attestation verification failed: %v\n", err)
			continue
		}
		verified = append(verified, *verifiedAtt)
	}

	return verified, nil
}

// verifyAttestation verifies a single attestation
func (v *Verifier) verifyAttestation(ctx context.Context, ref name.Reference, att oci.Signature) (*Attestation, error) {
	// Verify attestation signature
	sig, err := v.verifySignature(ctx, ref, att)
	if err != nil {
		return nil, fmt.Errorf("attestation signature verification failed: %w", err)
	}

	// Get attestation payload
	payload, err := att.Payload()
	if err != nil {
		return nil, fmt.Errorf("failed to get attestation payload: %w", err)
	}

	result := &Attestation{
		Payload:   payload,
		Signature: sig,
	}

	// Parse predicate type from payload
	// In-toto attestations have a predicateType field
	result.PredicateType = extractPredicateType(payload)

	return result, nil
}

// Helper functions

func extractIssuer(cert *x509.Certificate) string {
	for _, ext := range cert.Extensions {
		// Fulcio OID for issuer: 1.3.6.1.4.1.57264.1.1
		if ext.Id.Equal([]int{1, 3, 6, 1, 4, 1, 57264, 1, 1}) {
			return string(ext.Value)
		}
	}
	return ""
}

func extractSubject(cert *x509.Certificate) string {
	// Subject Alternative Name contains the OIDC subject
	for _, san := range cert.EmailAddresses {
		return san
	}
	if len(cert.URIs) > 0 {
		return cert.URIs[0].String()
	}
	return cert.Subject.CommonName
}

func extractPredicateType(payload []byte) string {
	// Parse JSON payload to extract predicateType
	// Simplified - in production, use proper JSON parsing
	// Example: {"predicateType": "https://slsa.dev/provenance/v0.2"}
	// For now, return a placeholder
	return "https://slsa.dev/provenance/v0.2"
}

// Suppress unused import warnings
var (
	_ = cosign.CheckOpts{}
	_ = cryptoutils.PEMType("")
	_ io.Reader
)
