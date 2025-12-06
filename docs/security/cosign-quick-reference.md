# Cosign Verification Quick Reference

## Table of Contents
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Common Commands](#common-commands)
- [Policy Templates](#policy-templates)
- [Troubleshooting](#troubleshooting)

## Installation

```bash
# Install Cosign CLI (for signing)
go install github.com/sigstore/cosign/v2/cmd/cosign@latest

# Freightliner includes verification built-in
# No additional installation needed
```

## Quick Start

### Sign an Image (Using Cosign CLI)

```bash
# Generate key pair
cosign generate-key-pair

# Sign with key
cosign sign --key cosign.key gcr.io/myproject/myimage:v1.0.0

# Sign keyless (GitHub Actions)
cosign sign --yes ghcr.io/${{ github.repository }}:${{ github.sha }}
```

### Verify During Replication (Using Freightliner)

```bash
# With public key
freightliner replicate \
  --source registry.io/image:v1 \
  --dest gcr.io/project/image:v1 \
  --verify-signature \
  --cosign-key cosign.pub

# Keyless
freightliner replicate \
  --source registry.io/image:v1 \
  --dest gcr.io/project/image:v1 \
  --verify-signature \
  --cosign-keyless

# With policy
freightliner replicate \
  --source registry.io/image:v1 \
  --dest gcr.io/project/image:v1 \
  --verify-signature \
  --cosign-policy policy.yaml
```

## Common Commands

### Cosign CLI (Signing)

```bash
# Sign with key
cosign sign --key cosign.key IMAGE

# Sign keyless
cosign sign --yes IMAGE

# Sign with attestation
cosign attest --key cosign.key --predicate provenance.json --type slsaprovenance IMAGE

# Verify (standalone)
cosign verify --key cosign.pub IMAGE
cosign verify IMAGE  # Keyless

# Generate SBOM
cosign attach sbom --sbom sbom.spdx IMAGE
```

### Freightliner CLI (Verification)

```bash
# Basic verification
freightliner replicate --verify-signature --cosign-key KEY SRC DEST

# Policy-based
freightliner replicate --verify-signature --cosign-policy POLICY SRC DEST

# Keyless
freightliner replicate --verify-signature --cosign-keyless SRC DEST

# Custom Rekor
freightliner replicate --verify-signature --cosign-rekor-url URL SRC DEST
```

## Policy Templates

### 1. Development (Permissive)

```yaml
requireSignature: true
minSignatures: 1
enforcementMode: warn  # Only warn
```

### 2. Staging (Moderate)

```yaml
requireSignature: true
minSignatures: 1
requireRekor: true
enforcementMode: enforce

allowedSigners:
  - emailRegex: ".*@example\\.com$"
```

### 3. Production (Strict)

```yaml
requireSignature: true
minSignatures: 2
requireRekor: true
enforcementMode: enforce

allowedSigners:
  - issuer: "https://token.actions.githubusercontent.com"
    subject: "https://github.com/org/repo/.github/workflows/release.yml@refs/heads/main"
  - email: "release-engineering@example.com"

requireAttestations:
  - predicateType: "https://slsa.dev/provenance/v0.2"
    minCount: 1
    requirements:
      slsaLevel: 3
```

### 4. GitHub Actions

```yaml
requireSignature: true
requireRekor: true
enforcementMode: enforce

allowedSigners:
  - issuer: "https://token.actions.githubusercontent.com"
    subjectRegex: "https://github.com/myorg/.*"

allowedIssuers:
  - "https://token.actions.githubusercontent.com"
```

### 5. Multi-Organization

```yaml
requireSignature: true
minSignatures: 1
requireRekor: true
enforcementMode: enforce

allowedSigners:
  # Org 1 - GitHub Actions
  - issuer: "https://token.actions.githubusercontent.com"
    subjectRegex: "https://github.com/org1/.*"

  # Org 2 - GitLab CI
  - issuer: "https://gitlab.com"
    uriRegex: "https://gitlab\\.com/org2/.*"

  # Manual release keys
  - email: "release@org1.com"
  - email: "release@org2.com"

deniedSigners:
  - email: "compromised@example.com"
```

### 6. Audit Mode (Monitoring)

```yaml
requireSignature: true
minSignatures: 1
enforcementMode: audit  # Log only, never fail

allowedSigners:
  - emailRegex: ".*@(example|trusted)\\.com$"
```

## Environment Variables

```bash
# Rekor URL
export COSIGN_REKOR_URL="https://rekor.sigstore.dev"

# Fulcio URL
export COSIGN_FULCIO_URL="https://fulcio.sigstore.dev"

# Disable transparency log (not recommended)
export COSIGN_EXPERIMENTAL=0
```

## GitHub Actions Example

```yaml
name: Build and Sign

on:
  push:
    branches: [main]

jobs:
  build-sign:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      id-token: write  # For keyless signing
      packages: write

    steps:
      - uses: actions/checkout@v3

      - name: Install Cosign
        uses: sigstore/cosign-installer@v3

      - name: Build image
        run: docker build -t ghcr.io/${{ github.repository }}:${{ github.sha }} .

      - name: Login to registry
        run: echo "${{ secrets.GITHUB_TOKEN }}" | docker login ghcr.io -u ${{ github.actor }} --password-stdin

      - name: Push image
        run: docker push ghcr.io/${{ github.repository }}:${{ github.sha }}

      - name: Sign image
        run: cosign sign --yes ghcr.io/${{ github.repository }}:${{ github.sha }}

      - name: Generate and attach SBOM
        run: |
          syft ghcr.io/${{ github.repository }}:${{ github.sha }} -o spdx-json > sbom.spdx
          cosign attach sbom --sbom sbom.spdx ghcr.io/${{ github.repository }}:${{ github.sha }}

      - name: Attest SLSA provenance
        run: |
          cosign attest --yes --type slsaprovenance \
            --predicate <(echo '{"buildType":"https://github.com/Attestations/GitHubActionsWorkflow@v1"}') \
            ghcr.io/${{ github.repository }}:${{ github.sha }}
```

## GitLab CI Example

```yaml
sign:
  stage: deploy
  image: gcr.io/projectsigstore/cosign:latest
  id_tokens:
    SIGSTORE_ID_TOKEN:
      aud: sigstore
  script:
    - cosign sign --yes $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  only:
    - main
```

## Troubleshooting

### Error: No signatures found

**Cause**: Image has no Cosign signatures

**Solution**:
```bash
# Sign the image first
cosign sign --key cosign.key IMAGE
```

### Error: Policy evaluation failed

**Cause**: Signature doesn't meet policy requirements

**Solution**:
```bash
# Check signature details
cosign verify --key cosign.pub IMAGE

# Review policy requirements
cat policy.yaml

# Use warn mode for testing
enforcementMode: warn
```

### Error: Rekor verification failed

**Cause**: Transparency log validation error

**Solution**:
```bash
# Check Rekor status
curl https://rekor.sigstore.dev/api/v1/log

# Sign with Rekor
cosign sign --yes IMAGE

# Disable Rekor (not recommended)
requireRekor: false
```

### Error: Certificate chain verification failed

**Cause**: Invalid Fulcio certificate

**Solution**:
```bash
# Re-sign with fresh certificate
cosign sign --yes IMAGE

# Check certificate
cosign verify IMAGE
```

### Error: Signer not allowed

**Cause**: Signature from unauthorized identity

**Solution**:
```yaml
# Add signer to policy
allowedSigners:
  - email: "authorized@example.com"
  - issuer: "https://token.actions.githubusercontent.com"
```

## Debugging Tips

### Check Signature Details

```bash
# View signature metadata
cosign verify --key cosign.pub IMAGE

# View attestations
cosign verify-attestation --key cosign.pub --type slsaprovenance IMAGE

# Check Rekor entry
cosign verify IMAGE  # Shows Rekor entry UUID
```

### Validate Policy

```go
// Test policy loading
policy, err := cosign.LoadPolicyFromFile("policy.yaml")
if err != nil {
    log.Fatal(err)
}
log.Printf("Policy: %+v", policy)
```

### Enable Verbose Logging

```bash
# Set log level
export LOG_LEVEL=debug

# Run with verbose output
freightliner replicate --verify-signature --verbose ...
```

## Common Patterns

### Pattern 1: Development Workflow

```bash
# 1. Develop locally
docker build -t myimage:dev .

# 2. Sign with local key
cosign generate-key-pair
cosign sign --key cosign.key myimage:dev

# 3. Verify
cosign verify --key cosign.pub myimage:dev

# 4. Replicate with verification
freightliner replicate \
  --verify-signature --cosign-key cosign.pub \
  localhost:5000/myimage:dev \
  gcr.io/project/myimage:dev
```

### Pattern 2: CI/CD Pipeline

```bash
# 1. Build in CI
docker build -t $IMAGE:$TAG .

# 2. Push to staging
docker push $IMAGE:$TAG

# 3. Sign keyless
cosign sign --yes $IMAGE:$TAG

# 4. Replicate to production with verification
freightliner replicate \
  --verify-signature --cosign-keyless \
  $IMAGE:$TAG \
  prod-registry/$IMAGE:$TAG
```

### Pattern 3: Multi-Environment

```bash
# Dev -> Staging (warn mode)
freightliner replicate \
  --verify-signature \
  --cosign-policy dev-policy.yaml \
  dev-registry/app:v1 \
  staging-registry/app:v1

# Staging -> Production (enforce mode)
freightliner replicate \
  --verify-signature \
  --cosign-policy prod-policy.yaml \
  staging-registry/app:v1 \
  prod-registry/app:v1
```

## Performance Tips

1. **Cache public keys**: Store keys locally to avoid repeated fetching
2. **Parallel verification**: Verify multiple signatures concurrently
3. **Policy optimization**: Use specific patterns instead of broad regex
4. **Rekor caching**: Cache Rekor entries to reduce API calls
5. **Connection pooling**: Reuse HTTP connections to Rekor

## Security Checklist

- [ ] Enable signature verification in production
- [ ] Use strict policy with signer allowlist
- [ ] Require Rekor transparency logs
- [ ] Mandate SLSA provenance attestations
- [ ] Implement multi-signature for critical images
- [ ] Regular key rotation for traditional keys
- [ ] Monitor policy violations
- [ ] Audit keyless signature OIDC identities
- [ ] Test policy in warn mode before enforcement
- [ ] Document authorized signers

## Resources

- **Sigstore**: https://www.sigstore.dev/
- **Cosign Docs**: https://docs.sigstore.dev/cosign/overview/
- **SLSA**: https://slsa.dev/
- **Rekor**: https://docs.sigstore.dev/rekor/overview/
- **Policy Examples**: `/examples/cosign-policy.yaml`
- **Full Documentation**: `/docs/security/cosign-verification.md`

## Getting Help

```bash
# Cosign help
cosign help
cosign verify --help

# Freightliner help
freightliner replicate --help

# Check Sigstore status
curl https://status.sigstore.dev/
```

---

**Quick Links:**
- [Full Documentation](./cosign-verification.md)
- [Implementation Details](../architecture/cosign-implementation.md)
- [Policy Examples](../../examples/cosign-policy.yaml)
