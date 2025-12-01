# Image Signature Verification

This package provides comprehensive interfaces for container image signature verification with full support for Cosign/Sigstore integration.

## Overview

The signature verification system supports multiple signing mechanisms and trust models:

- **Cosign/Sigstore**: Keyless signing with OIDC, transparency logs (Rekor), and certificate authority (Fulcio)
- **Traditional Signing**: Key-based signing with RSA, ECDSA, or Ed25519 keys
- **Attestations**: SLSA provenance, SBOM, vulnerability scans, and custom attestations
- **Trust Policies**: Flexible policy-based verification with multiple trust roots

## Architecture

### Core Interfaces

1. **ImageSigner**: Signs container images with various signing mechanisms
2. **SignatureVerifier**: Verifies signatures and attestations
3. **KeyProvider**: Manages cryptographic keys (KMS, Vault, files)
4. **CertificateAuthority**: X.509 certificate validation and trust chains
5. **TransparencyLog**: Rekor transparency log integration
6. **PolicyEngine**: Policy-based verification decisions
7. **AttestationProvider**: Manages image attestations

### Key Types

- **SignatureMetadata**: Complete signature information including keys, certificates, timestamps
- **VerificationResult**: Comprehensive verification results with trust validation
- **SigningConfig/VerificationConfig**: Configuration for signing and verification operations
- **TrustPolicy**: Declarative trust policies for automated verification

## Cosign/Sigstore Integration

### Keyless Signing Workflow

Cosign's keyless signing uses OIDC identity for signing without managing long-lived keys:

```
┌─────────────────────────────────────────────────────────────┐
│                     Keyless Signing Flow                     │
└─────────────────────────────────────────────────────────────┘

1. User authenticates with OIDC provider (Google, GitHub, Microsoft)
   └─> Obtains OIDC identity token with email/subject

2. ImageSigner.Sign() with IdentityToken
   └─> Sends OIDC token to Fulcio CA
       └─> Fulcio issues short-lived certificate (10 min)
           └─> Certificate contains OIDC identity in SAN extension

3. Generate ephemeral key pair
   └─> Sign image digest with private key
   └─> Bundle signature with certificate

4. Upload to Rekor transparency log
   └─> Creates immutable, verifiable record
   └─> Returns SignedEntryTimestamp (proof of inclusion)

5. Attach signature to registry
   └─> Stored as OCI artifact alongside image
   └─> Contains: signature, certificate, Rekor bundle
```

### Key-Based Signing Workflow

Traditional signing with long-lived key pairs:

```
┌─────────────────────────────────────────────────────────────┐
│                     Key-Based Signing Flow                   │
└─────────────────────────────────────────────────────────────┘

1. KeyProvider.GetPrivateKey(keyID)
   └─> Load from KMS, Vault, or local file
   └─> Support for AWS KMS, GCP KMS, Azure Key Vault, HashiCorp Vault

2. ImageSigner.Sign() with KeyID
   └─> Generate signature over image digest
   └─> Include public key or key reference

3. Optional: Upload to transparency log
   └─> Provides non-repudiation
   └─> Enables public auditability

4. Attach signature to registry
   └─> Stored with key reference for verification
```

### Verification Workflow

```
┌─────────────────────────────────────────────────────────────┐
│                     Verification Flow                        │
└─────────────────────────────────────────────────────────────┘

1. SignatureVerifier.Verify(imageRef)
   └─> Fetch image digest from registry
   └─> Retrieve all signatures for digest

2. For each signature:
   a) Keyless signatures:
      └─> Verify certificate chain against Fulcio root
      └─> Check certificate validity period
      └─> Verify OIDC identity matches trust policy
      └─> Validate Rekor bundle (if required)
      └─> Verify SCT (Signed Certificate Timestamp)

   b) Key-based signatures:
      └─> Load public key from KeyProvider
      └─> Verify signature cryptographically
      └─> Check key validity and revocation status

3. Policy validation:
   └─> Check against TrustPolicy
   └─> Verify required attestations exist
   └─> Validate signature count requirements
   └─> Check allowed identities/issuers

4. Return VerificationResult
   └─> List of verified signatures
   └─> Policy compliance status
   └─> Detailed validation messages
```

## Implementation Guide

### 1. Cosign Integration

The recommended implementation uses the official Cosign libraries:

```go
import (
    "github.com/sigstore/cosign/v2/pkg/cosign"
    "github.com/sigstore/cosign/v2/pkg/oci"
    "github.com/sigstore/cosign/v2/pkg/oci/remote"
)

// CosignSigner implements ImageSigner using Cosign
type CosignSigner struct {
    opts cosign.SignOpts
}

func (s *CosignSigner) Sign(ctx context.Context, imageRef string, opts SigningOptions) (*SignatureMetadata, error) {
    // 1. Parse image reference
    ref, err := name.ParseReference(imageRef)

    // 2. Load signing configuration
    // For keyless: use OIDC token
    if opts.IdentityToken != "" {
        // Fulcio-based signing
        cert, err := fulcio.GetCert(ctx, opts.IdentityToken)
        privateKey, err := ephemeral.GenerateKey()

        // Sign with ephemeral key
        sig, err := cosign.Sign(ctx, privateKey, digest)

        // Upload to Rekor
        if opts.UploadToTLog {
            entry, err := cosign.TLogUpload(ctx, sig, cert)
        }
    } else {
        // Key-based signing
        privateKey, err := loadPrivateKey(opts.KeyID)
        sig, err := cosign.Sign(ctx, privateKey, digest)
    }

    // 3. Attach signature to registry
    err = remote.WriteSignature(ref, sig)

    return convertToSignatureMetadata(sig), nil
}
```

### 2. Verification Implementation

```go
// CosignVerifier implements SignatureVerifier
type CosignVerifier struct {
    rekorClient *client.Rekor
    fulcioRoots *x509.CertPool
}

func (v *CosignVerifier) Verify(ctx context.Context, imageRef string, opts VerificationOptions) (*VerificationResult, error) {
    // 1. Fetch signatures from registry
    ref, err := name.ParseReference(imageRef)
    sigs, err := cosign.FetchSignatures(ref)

    result := &VerificationResult{
        ImageRef: imageRef,
        VerifiedAt: time.Now(),
    }

    // 2. Verify each signature
    for _, sig := range sigs {
        // Check if keyless or key-based
        if cert, err := sig.Cert(); err == nil && cert != nil {
            // Keyless verification
            if err := v.verifyKeyless(ctx, sig, cert, opts); err != nil {
                result.FailedSignatures = append(result.FailedSignatures, sig)
                result.ValidationErrors = append(result.ValidationErrors, err)
                continue
            }
        } else {
            // Key-based verification
            pubKey, err := loadPublicKey(opts.KeyIDs)
            if err := cosign.VerifySignature(sig, pubKey); err != nil {
                result.FailedSignatures = append(result.FailedSignatures, sig)
                result.ValidationErrors = append(result.ValidationErrors, err)
                continue
            }
        }

        result.VerifiedSignatures = append(result.VerifiedSignatures, sig)
    }

    result.Verified = len(result.VerifiedSignatures) > 0
    return result, nil
}

func (v *CosignVerifier) verifyKeyless(ctx context.Context, sig oci.Signature, cert *x509.Certificate, opts VerificationOptions) error {
    // 1. Verify certificate chain against Fulcio roots
    if err := verifyCertChain(cert, v.fulcioRoots); err != nil {
        return fmt.Errorf("certificate chain verification failed: %w", err)
    }

    // 2. Check certificate validity
    now := time.Now()
    if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
        return fmt.Errorf("certificate expired or not yet valid")
    }

    // 3. Verify OIDC identity
    identity, issuer := extractIdentityFromCert(cert)
    if opts.CertificateIdentity != "" && identity != opts.CertificateIdentity {
        return fmt.Errorf("identity mismatch: expected %s, got %s", opts.CertificateIdentity, identity)
    }
    if opts.CertificateOIDCIssuer != "" && issuer != opts.CertificateOIDCIssuer {
        return fmt.Errorf("issuer mismatch: expected %s, got %s", opts.CertificateOIDCIssuer, issuer)
    }

    // 4. Verify Rekor bundle
    if opts.RequireRekorBundle {
        bundle, err := sig.Bundle()
        if err != nil || bundle == nil {
            return fmt.Errorf("rekor bundle required but not found")
        }

        if err := v.verifyRekorBundle(ctx, bundle); err != nil {
            return fmt.Errorf("rekor bundle verification failed: %w", err)
        }
    }

    // 5. Verify SCT (Signed Certificate Timestamp)
    if !opts.IgnoreSCT {
        if err := verifySCT(cert); err != nil {
            return fmt.Errorf("SCT verification failed: %w", err)
        }
    }

    return nil
}
```

### 3. Key Management Integration

```go
// KMSKeyProvider implements KeyProvider for cloud KMS services
type KMSKeyProvider struct {
    kmsClient interface{} // AWS KMS, GCP KMS, Azure Key Vault
}

func (p *KMSKeyProvider) GetPublicKey(ctx context.Context, keyID string) (crypto.PublicKey, error) {
    // Parse key ID to determine KMS provider
    // Format: "awskms:///<key-id>" or "gcpkms:///<key-id>"
    provider, id := parseKeyID(keyID)

    switch provider {
    case "awskms":
        return p.getAWSPublicKey(ctx, id)
    case "gcpkms":
        return p.getGCPPublicKey(ctx, id)
    case "azurekms":
        return p.getAzurePublicKey(ctx, id)
    case "hashivault":
        return p.getVaultPublicKey(ctx, id)
    default:
        return nil, fmt.Errorf("unsupported KMS provider: %s", provider)
    }
}

func (p *KMSKeyProvider) GetPrivateKey(ctx context.Context, keyID string) (crypto.PrivateKey, error) {
    // Most KMS services don't export private keys
    // Return a signer interface instead
    return p.getKMSSigner(ctx, keyID)
}
```

### 4. Attestation Support

```go
// AttestationProviderImpl implements AttestationProvider
type AttestationProviderImpl struct {
    signer ImageSigner
}

func (p *AttestationProviderImpl) CreateAttestation(ctx context.Context, attestation *Attestation) (*AttestationMetadata, error) {
    // 1. Create in-toto statement
    stmt := intoto.Statement{
        StatementHeader: intoto.StatementHeader{
            Type:          intoto.StatementInTotoV01,
            PredicateType: attestation.PredicateType,
            Subject:       convertSubjects(attestation.Subject),
        },
        Predicate: attestation.Predicate,
    }

    // 2. Sign attestation
    payload, err := json.Marshal(stmt)
    sig, err := p.signer.Sign(ctx, payload, SigningOptions{})

    // 3. Store attestation
    return &AttestationMetadata{
        ID:        generateID(),
        Type:      attestation.Type,
        CreatedAt: time.Now(),
    }, nil
}

func (p *AttestationProviderImpl) VerifyAttestation(ctx context.Context, attestation *Attestation, opts VerificationOptions) (*VerificationResult, error) {
    // Verify attestation signature using standard verification
    return verifyAttestationSignature(attestation, opts)
}
```

## Configuration Examples

### Signing Configuration

```yaml
# config/signing.yaml
signing:
  enabled: true

  # Keyless signing with OIDC
  keyless:
    enabled: true
    fulcioURL: https://fulcio.sigstore.dev
    rekorURL: https://rekor.sigstore.dev
    allowedIssuers:
      - https://accounts.google.com
      - https://token.actions.githubusercontent.com

  # Key-based signing
  keyBased:
    enabled: true
    defaultKeyID: "awskms:///arn:aws:kms:us-east-1:123456789:key/abc-123"
    keyProvider: "kms"
    algorithm: "ECDSA_P256_SHA256"

  # Transparency log
  transparencyLog:
    enabled: true
    uploadToTLog: true
    rekorURL: https://rekor.sigstore.dev

  # Default annotations
  annotations:
    builder: "freightliner"
    buildTimestamp: "{{ .Timestamp }}"
    gitCommit: "{{ .GitCommit }}"
```

### Verification Configuration

```yaml
# config/verification.yaml
verification:
  enabled: true
  enforceVerification: true

  # Trust configuration
  trust:
    # Trusted public keys (for key-based signing)
    publicKeys:
      - |
        -----BEGIN PUBLIC KEY-----
        MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE...
        -----END PUBLIC KEY-----

    # Trusted root certificates (for keyless signing)
    trustRootCertificates:
      - |
        -----BEGIN CERTIFICATE-----
        MIICGjCCAaGgAwIBAgIUALnViVfnU0brJasmRk...
        -----END CERTIFICATE-----

    # Allowed OIDC issuers
    certificateOIDCIssuers:
      - "https://accounts.google.com"
      - "https://token.actions.githubusercontent.com"

    # Allowed identities (email addresses)
    certificateIdentities:
      - "builder@example.com"
      - "*@mycompany.com"  # Wildcard support

  # Transparency log verification
  transparencyLog:
    enabled: true
    requireRekorBundle: true
    rekorURL: https://rekor.sigstore.dev

  # Verification policies
  maxSignatureAge: 720h  # 30 days
  ignoreSCT: false
  ignoreTLog: false

  # Trust policies
  trustPolicies:
    - name: "production-images"
      imagePattern: "registry.example.com/prod/*"
      requireSignatureCount: 2
      requireAttestations:
        - slsaprovenance
        - vulnerabilityscan
      enforceTransparencyLog: true

    - name: "internal-images"
      imagePattern: "registry.example.com/internal/*"
      requireSignatureCount: 1
      trustRoots:
        - "awskms:///arn:aws:kms:us-east-1:123456789:key/abc-123"
```

## Key Management Strategy

### Key Storage Options

1. **Cloud KMS**:
   - AWS KMS: `awskms:///arn:aws:kms:region:account:key/key-id`
   - GCP KMS: `gcpkms://projects/PROJECT/locations/LOCATION/keyRings/RING/cryptoKeys/KEY`
   - Azure Key Vault: `azurekms://VAULT_NAME.vault.azure.net/keys/KEY_NAME`

2. **HashiCorp Vault**:
   - `hashivault://vault.example.com/transit/keys/signing-key`

3. **Local Files** (not recommended for production):
   - `file:///path/to/private-key.pem`

### Key Rotation Strategy

```yaml
# Automated key rotation
keyRotation:
  enabled: true
  rotationInterval: 90d

  # Retain old keys for verification
  retentionPolicy:
    keepPreviousKeys: 3
    gracePeriod: 30d

  # Notification on rotation
  notifications:
    - type: "email"
      recipients: ["security@example.com"]
    - type: "slack"
      webhook: "https://hooks.slack.com/..."
```

### Key Revocation

```yaml
# Key revocation configuration
revocation:
  # Check revocation status
  checkOnVerification: true

  # Revocation list sources
  crl:
    - url: "https://pki.example.com/crl.pem"
      checkInterval: 24h

  ocsp:
    enabled: true
    responderURL: "https://ocsp.example.com"

  # Custom revocation list
  customRevocationList:
    - keyID: "key-123"
      revokedAt: "2024-01-15T10:00:00Z"
      reason: "keyCompromise"
```

## Attestation Integration

### SLSA Provenance

```go
// Generate SLSA provenance attestation
func GenerateSLSAProvenance(imageRef string, buildInfo BuildInfo) (*Attestation, error) {
    provenance := slsa.ProvenancePredicate{
        Builder: slsa.Builder{
            ID: "https://builder.example.com",
        },
        BuildType: "https://example.com/BuildType/v1",
        Invocation: slsa.Invocation{
            ConfigSource: slsa.ConfigSource{
                URI:        buildInfo.RepoURI,
                Digest:     map[string]string{"sha1": buildInfo.CommitSHA},
                EntryPoint: buildInfo.Workflow,
            },
        },
        BuildConfig: buildInfo.Config,
        Materials: []slsa.Material{
            {
                URI:    buildInfo.SourceURI,
                Digest: map[string]string{"sha256": buildInfo.SourceDigest},
            },
        },
    }

    return &Attestation{
        Type:          AttestationTypeSLSAProvenance,
        PredicateType: slsa.PredicateSLSAProvenance,
        Subject: []Subject{{
            Name:   imageRef,
            Digest: map[string]string{"sha256": buildInfo.ImageDigest},
        }},
        Predicate: toMap(provenance),
        CreatedAt: time.Now(),
    }, nil
}
```

### SBOM Attestation

```go
// Generate SBOM attestation
func GenerateSBOMAttestation(imageRef string, sbom *SBOM) (*Attestation, error) {
    return &Attestation{
        Type:          AttestationTypeSPDX,
        PredicateType: "https://spdx.dev/Document",
        Subject: []Subject{{
            Name:   imageRef,
            Digest: map[string]string{"sha256": sbom.ImageDigest},
        }},
        Predicate: toMap(sbom),
        CreatedAt: time.Now(),
    }, nil
}
```

## Security Best Practices

1. **Use Keyless Signing for CI/CD**:
   - No long-lived credentials to manage
   - OIDC identity provides strong authentication
   - Transparency log provides audit trail

2. **Require Multiple Signatures for Production**:
   - Builder signature (automated)
   - Security team signature (manual approval)
   - Release manager signature (release approval)

3. **Enforce Attestations**:
   - SLSA provenance for build integrity
   - SBOM for supply chain transparency
   - Vulnerability scan results for security posture

4. **Use Transparency Logs**:
   - Non-repudiation of signatures
   - Public auditability
   - Tamper detection

5. **Implement Strong Trust Policies**:
   - Restrict allowed OIDC issuers
   - Validate certificate identities
   - Require minimum signature count
   - Verify attestations

6. **Key Management**:
   - Use cloud KMS for production keys
   - Implement automated key rotation
   - Monitor key usage and access
   - Maintain key revocation list

7. **Monitor and Alert**:
   - Log all signature operations
   - Alert on verification failures
   - Track signature coverage metrics
   - Audit transparency log entries

## Testing Strategy

```go
// Test keyless signing and verification
func TestKeylessSigning(t *testing.T) {
    // Mock OIDC token
    token := generateTestOIDCToken()

    signer := NewCosignSigner()
    sig, err := signer.Sign(ctx, "test-image:latest", SigningOptions{
        IdentityToken: token,
        UploadToTLog:  true,
    })
    require.NoError(t, err)

    verifier := NewCosignVerifier()
    result, err := verifier.Verify(ctx, "test-image:latest", VerificationOptions{
        CertificateIdentity:   "test@example.com",
        CertificateOIDCIssuer: "https://accounts.google.com",
        RequireRekorBundle:    true,
    })
    require.NoError(t, err)
    assert.True(t, result.Verified)
}
```

## Future Enhancements

1. **Policy as Code**:
   - OPA/Rego policy engine integration
   - CEL (Common Expression Language) support
   - Custom policy DSL

2. **Advanced Attestations**:
   - Vulnerability remediation attestations
   - Runtime behavior attestations
   - Compliance attestations

3. **Federation**:
   - Cross-organization trust
   - Federated transparency logs
   - Multi-party signing

4. **Hardware Security**:
   - TPM-based signing
   - HSM integration
   - Secure enclave support

## Resources

- [Cosign Documentation](https://docs.sigstore.dev/cosign/overview/)
- [Sigstore Architecture](https://docs.sigstore.dev/system_config/architecture/)
- [SLSA Provenance](https://slsa.dev/provenance/)
- [In-Toto Attestations](https://github.com/in-toto/attestation)
- [OCI Image Spec](https://github.com/opencontainers/image-spec)
