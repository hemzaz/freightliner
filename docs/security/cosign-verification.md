# Cosign Signature Verification

Comprehensive Cosign signature verification implementation for container images, supporting both traditional key-based and keyless (Sigstore) verification workflows.

## Overview

This implementation provides:

- **Public Key Verification**: Verify signatures using PEM-encoded public keys
- **Keyless Verification**: Verify signatures using Fulcio certificates and Rekor transparency logs
- **Policy Engine**: Flexible policy-based signature validation
- **Attestation Support**: Verify SLSA provenance and SBOM attestations
- **Multi-Signature**: Support for multiple signatures per image
- **Rekor Integration**: Transparency log verification for supply chain security

## Architecture

### Components

```
pkg/security/cosign/
├── verifier.go    - Core signature verification (350+ lines)
├── policy.go      - Policy engine and evaluation (250+ lines)
└── rekor.go       - Rekor transparency log client (350+ lines)
```

### Key Types

```go
// Verifier - Main verification orchestrator
type Verifier struct {
    config      *VerifierConfig
    rekorClient *RekorClient
    policy      *Policy
    verifiers   []signature.Verifier
}

// Policy - Signature verification policy
type Policy struct {
    RequireSignature bool
    MinSignatures    int
    AllowedSigners   []SignerIdentity
    RequireRekor     bool
    EnforcementMode  string
}

// Signature - Verified signature metadata
type Signature struct {
    Digest      string
    Certificate *x509.Certificate
    Payload     []byte
    Bundle      *bundle.RekorBundle
    Issuer      string  // OIDC issuer
    Subject     string  // OIDC subject
}
```

## Usage

### Basic Verification with Public Key

```go
import (
    "context"
    "github.com/google/go-containerregistry/pkg/name"
    "freightliner/pkg/security/cosign"
)

// Create verifier with public key
config := &cosign.VerifierConfig{
    PublicKeyPath: "/path/to/cosign.pub",
}

verifier, err := cosign.NewVerifier(config)
if err != nil {
    log.Fatal(err)
}

// Verify image signature
ref, _ := name.ParseReference("gcr.io/myorg/myimage:v1.0.0")
signatures, err := verifier.Verify(context.Background(), ref)
if err != nil {
    log.Fatalf("Verification failed: %v", err)
}

log.Printf("Found %d valid signatures", len(signatures))
```

### Keyless Verification (Sigstore)

```go
// Create verifier for keyless verification
config := &cosign.VerifierConfig{
    EnableKeyless: true,
    RekorURL:      "https://rekor.sigstore.dev",
    FulcioURL:     "https://fulcio.sigstore.dev",
}

verifier, err := cosign.NewVerifier(config)
if err != nil {
    log.Fatal(err)
}

signatures, err := verifier.Verify(context.Background(), ref)
if err != nil {
    log.Fatalf("Keyless verification failed: %v", err)
}

// Check OIDC identity
for _, sig := range signatures {
    log.Printf("Signed by: %s (issuer: %s)", sig.Subject, sig.Issuer)
}
```

### Policy-Based Verification

```go
// Load policy from file
policy, err := cosign.LoadPolicyFromFile("cosign-policy.yaml")
if err != nil {
    log.Fatal(err)
}

// Create verifier with policy
config := &cosign.VerifierConfig{
    EnableKeyless: true,
    Policy:        policy,
}

verifier, err := cosign.NewVerifier(config)
if err != nil {
    log.Fatal(err)
}

// Verification automatically applies policy
signatures, err := verifier.Verify(context.Background(), ref)
// Fails if policy requirements are not met
```

### Attestation Verification

```go
// Verify SLSA provenance attestations
attestations, err := verifier.VerifyAttestation(context.Background(), ref)
if err != nil {
    log.Fatalf("Attestation verification failed: %v", err)
}

for _, att := range attestations {
    log.Printf("Attestation type: %s", att.PredicateType)
    // Parse attestation payload for provenance details
}
```

## Policy Configuration

### Policy File Format (YAML)

```yaml
# Basic policy
requireSignature: true
minSignatures: 1
requireRekor: true
enforcementMode: enforce  # enforce, warn, or audit

# Allowed signers (whitelist)
allowedSigners:
  # GitHub Actions
  - issuer: "https://token.actions.githubusercontent.com"
    subject: "https://github.com/myorg/myrepo/.github/workflows/release.yml@refs/heads/main"

  # Email-based
  - email: "release@example.com"
  - emailRegex: ".*@example\\.com$"

  # GitLab CI
  - issuer: "https://gitlab.com"
    uriRegex: "https://gitlab\\.com/myorg/.*"

# Denied signers (blacklist)
deniedSigners:
  - email: "compromised@example.com"

# Allowed OIDC issuers
allowedIssuers:
  - "https://token.actions.githubusercontent.com"
  - "https://gitlab.com"

# Key requirements (for traditional signing)
keyRequirements:
  minKeySize: 2048
  allowedAlgorithms:
    - "ECDSA"
    - "RSA"
  requireHardwareKey: false

# Required attestations
requireAttestations:
  - predicateType: "https://slsa.dev/provenance/v0.2"
    minCount: 1
    requirements:
      slsaLevel: 3
  - predicateType: "https://spdx.dev/Document"
    minCount: 1
```

### Policy Examples

**Minimal Policy** (Basic verification):
```yaml
requireSignature: true
minSignatures: 1
enforcementMode: enforce
```

**Strict Production Policy**:
```yaml
requireSignature: true
minSignatures: 2  # Dual signatures required
enforcementMode: enforce
requireRekor: true
allowedSigners:
  - issuer: "https://token.actions.githubusercontent.com"
    subject: "https://github.com/myorg/myrepo/.github/workflows/release.yml@refs/heads/main"
requireAttestations:
  - predicateType: "https://slsa.dev/provenance/v0.2"
    minCount: 1
    requirements:
      slsaLevel: 4
```

**Warn-Only Policy** (Gradual rollout):
```yaml
requireSignature: true
enforcementMode: warn  # Only warn, don't fail
requireRekor: true
```

## Enforcement Modes

- **enforce**: Fail verification if policy requirements are not met (default)
- **warn**: Log warnings but allow verification to pass
- **audit**: Log violations for monitoring without blocking

## Integration with Freightliner

### CLI Flags

```bash
# Verify signatures during replication
freightliner replicate \
  --source registry.example.com/myimage:v1.0.0 \
  --dest gcr.io/myproject/myimage:v1.0.0 \
  --verify-signature \
  --cosign-key /path/to/cosign.pub

# With policy file
freightliner replicate \
  --source registry.example.com/myimage:v1.0.0 \
  --dest gcr.io/myproject/myimage:v1.0.0 \
  --verify-signature \
  --cosign-policy cosign-policy.yaml

# Keyless verification
freightliner replicate \
  --source registry.example.com/myimage:v1.0.0 \
  --dest gcr.io/myproject/myimage:v1.0.0 \
  --verify-signature \
  --cosign-keyless \
  --cosign-policy cosign-policy.yaml
```

## Security Features

### Public Key Verification

- PEM-encoded ECDSA/RSA public keys
- Multiple key support (key rotation)
- Hardware key requirement enforcement

### Keyless Verification (Sigstore)

- Fulcio certificate verification
- X.509 certificate chain validation
- OIDC identity verification
- Rekor transparency log verification
- Signed Entry Timestamp (SET) validation
- Merkle inclusion proof verification

### Policy Enforcement

- Minimum signature count requirements
- Signer identity allowlist/denylist
- Email and URI pattern matching
- OIDC issuer restrictions
- Attestation type requirements
- Flexible enforcement modes

### Attestation Verification

- SLSA provenance verification
- SBOM (Software Bill of Materials) verification
- Custom predicate type support
- Attestation signature verification

## Testing

### Unit Tests

```bash
# Run all security tests
go test ./tests/pkg/security/cosign/...

# Run with coverage
go test -cover ./tests/pkg/security/cosign/...

# Run specific test
go test -v -run TestVerifier_Verify ./tests/pkg/security/cosign/
```

### Test Coverage

- **verifier_test.go**: Verifier initialization, signature verification, keyless verification
- **policy_test.go**: Policy validation, evaluation, enforcement modes, signer matching
- **rekor_test.go**: Rekor client operations, bundle verification, entry retrieval

### Example Test Cases

```go
func TestVerifier_Verify(t *testing.T) {
    config := &VerifierConfig{
        PublicKey: generateTestPublicKey(t),
    }

    verifier, err := NewVerifier(config)
    require.NoError(t, err)

    // Test verification logic
}

func TestPolicy_Evaluate(t *testing.T) {
    policy := &Policy{
        RequireSignature: true,
        MinSignatures:    2,
    }

    signatures := []Signature{
        {Digest: "sha256:abc"},
    }

    err := policy.Evaluate(context.Background(), signatures)
    assert.Error(t, err) // Should fail with insufficient signatures
}
```

## Rekor Transparency Log

### Features

- Entry search by digest
- Entry retrieval by UUID
- Bundle verification
- Signed Entry Timestamp (SET) verification
- Merkle inclusion proof validation
- Public key retrieval

### Usage

```go
rekorClient := cosign.NewRekorClient("https://rekor.sigstore.dev")

// Search for entries by image digest
entries, err := rekorClient.SearchByDigest(ctx, "sha256:abc123...")
if err != nil {
    log.Fatal(err)
}

// Verify specific entry
entry := entries[0]
err = rekorClient.VerifyEntry(ctx, entry)
if err != nil {
    log.Fatalf("Entry verification failed: %v", err)
}
```

## Best Practices

### Production Deployments

1. **Use Policy Files**: Define clear signature requirements
2. **Enable Rekor**: Ensure transparency log verification
3. **Restrict Signers**: Use allowlist for authorized identities
4. **Require Attestations**: Verify SLSA provenance for build integrity
5. **Monitor Warnings**: Track policy violations in audit/warn modes
6. **Key Rotation**: Support multiple public keys for seamless rotation

### GitHub Actions Integration

```yaml
# .github/workflows/release.yml
- name: Sign image with keyless
  run: |
    cosign sign --yes ghcr.io/${{ github.repository }}:${{ github.sha }}

# Verification policy
allowedSigners:
  - issuer: "https://token.actions.githubusercontent.com"
    subject: "https://github.com/${{ github.repository }}/.github/workflows/release.yml@refs/heads/main"
```

### GitLab CI Integration

```yaml
# .gitlab-ci.yml
sign:
  script:
    - cosign sign --yes $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA

# Verification policy
allowedSigners:
  - issuer: "https://gitlab.com"
    uriRegex: "https://gitlab\\.com/${CI_PROJECT_NAMESPACE}/.*"
```

## Error Handling

### Common Errors

- **No signatures found**: Image has no Cosign signatures
- **Policy evaluation failed**: Signatures don't meet policy requirements
- **Rekor verification failed**: Transparency log validation error
- **Certificate chain verification failed**: Invalid Fulcio certificate
- **Signer not allowed**: Signature from unauthorized identity

### Debugging

```go
// Enable verbose logging
verifier.config.Verbose = true

// Check signature details
for _, sig := range signatures {
    log.Printf("Signature: %+v", sig)
    if sig.Certificate != nil {
        log.Printf("Certificate Subject: %s", sig.Certificate.Subject)
        log.Printf("Certificate Issuer: %s", sig.Certificate.Issuer)
    }
}
```

## Performance Considerations

- **Caching**: Cache Rekor entries and public keys
- **Parallel Verification**: Verify multiple signatures concurrently
- **Registry Optimization**: Use efficient registry clients
- **Timeout Configuration**: Set appropriate HTTP timeouts

## Security Considerations

- **Key Management**: Securely store and distribute public keys
- **Policy Storage**: Protect policy files from tampering
- **Certificate Validation**: Always verify certificate chains
- **Transparency Logs**: Enable Rekor for tamper evidence
- **Identity Verification**: Strictly validate OIDC identities

## References

- [Sigstore Documentation](https://docs.sigstore.dev/)
- [Cosign GitHub](https://github.com/sigstore/cosign)
- [SLSA Framework](https://slsa.dev/)
- [Rekor Transparency Log](https://github.com/sigstore/rekor)
- [Fulcio Certificate Authority](https://github.com/sigstore/fulcio)

## Future Enhancements

- [ ] Hardware security module (HSM) support
- [ ] Advanced SLSA level validation
- [ ] Custom attestation predicate parsers
- [ ] Policy template library
- [ ] Grafana dashboard for signature metrics
- [ ] Automated key rotation
- [ ] Multi-cloud KMS integration
