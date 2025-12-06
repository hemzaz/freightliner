# Skopeo Quick Reference for Freightliner Developers

## Transport Reference

```
docker://registry.io/repo:tag         # Docker Registry
docker://registry.io/repo@sha256:...  # By digest
oci:/path/to/layout:tag               # OCI Layout
dir:/path/to/directory                # Directory
docker-archive:/path/file.tar:tag     # Docker tar
oci-archive:/path/file.tar:tag        # OCI tar
containers-storage:image:tag          # Local storage
docker-daemon:image:tag               # Docker daemon
```

## Common Commands

### Copy Image
```bash
# Basic
skopeo copy docker://source docker://dest

# With auth
skopeo copy --src-creds user:pass --dest-creds admin:secret \
            docker://source docker://dest

# Format conversion
skopeo copy --format oci docker://source oci:dest

# Multi-arch
skopeo copy --all docker://source docker://dest

# Sign during copy
skopeo copy --sign-by GPG_KEY docker://source docker://dest
```

### Inspect Image
```bash
# Basic
skopeo inspect docker://registry.io/image:tag

# Raw manifest
skopeo inspect --raw docker://image

# Extract digest
skopeo inspect --format '{{.Digest}}' docker://image

# Get config
skopeo inspect --config docker://image
```

### List Tags
```bash
skopeo list-tags docker://registry.io/repo
```

### Delete Image
```bash
skopeo delete docker://registry.io/repo:old-tag
```

### Login/Logout
```bash
# Login
echo "$PASSWORD" | skopeo login -u user --password-stdin registry.io

# Logout
skopeo logout registry.io

# Custom auth file
skopeo login --authfile /tmp/auth.json registry.io
```

### Bulk Sync
```bash
# From YAML config
skopeo sync --src yaml --dest docker config.yaml dest-registry.io

# Single repo to dir
skopeo sync --src docker --dest dir --scoped \
            registry.io/repo /backup/images

# Dry run
skopeo sync --src yaml --dest docker --dry-run config.yaml dest
```

## Authentication Patterns

### Inline Credentials (Insecure)
```bash
--creds USERNAME:PASSWORD
--username USER --password PASS
```

### Credential File
```bash
--authfile /path/to/auth.json
```

Default: `${XDG_RUNTIME_DIR}/containers/auth.json`

### Bearer Token
```bash
--registry-token TOKEN
```

### Anonymous
```bash
--no-creds
```

## YAML Sync Configuration

```yaml
registry.example.com:
  # Explicit tags
  images:
    alpine: ["3.14", "3.15", "3.16"]
    nginx: []  # All tags
    busybox: ["sha256:abc123..."]  # By digest

  # Regex filtering
  images-by-tag-regex:
    golang: "^1\\.(19|20|21)\\."

  # Semver filtering
  images-by-semver:
    postgresql: ">=12.0 <16.0"

  # Authentication
  credentials:
    username: user
    password: pass

  # TLS
  tls-verify: true
  cert-dir: /certs

quay.io:
  images:
    prometheus/prometheus: ["v2.40.0"]
```

## Manifest Format Options

```bash
--format oci     # OCI Image Manifest
--format v2s1    # Docker v2 Schema 1 (deprecated)
--format v2s2    # Docker v2 Schema 2
```

## Multi-Architecture Handling

```bash
--all                      # Copy all images in list
--multi-arch system        # Copy only current platform (default)
--multi-arch all           # Copy all platforms
--multi-arch index-only    # Copy index only, no images
```

## Compression Options

```bash
--dest-compress                    # Compress layers
--dest-decompress                  # Decompress layers
--dest-compress-format gzip        # gzip, zstd, bzip2
--dest-compress-level 9            # 1-9
--force-compress-format            # Recompress if needed
```

## Performance Options

```bash
--image-parallel-copies 8          # Parallel layer transfers
--dest-precompute-digests         # Calculate digests for dedup
--src-shared-blob-dir /blobs      # Shared OCI blobs
--dest-shared-blob-dir /blobs
```

## Retry Configuration

```bash
--retry-times 3                    # Number of retries
--retry-delay 5s                   # Fixed delay
--retry-delay 0                    # Exponential backoff (default)
```

## TLS Configuration

```bash
--tls-verify                       # Enable verification (default)
--tls-verify=false                 # Skip verification
--cert-dir /path/to/certs          # Custom certificates
```

Certificate files expected:
- `registry.example.com.crt` - Server certificate
- `registry.example.com.cert` - Alternative
- `registry.example.com.key` - Client key
- `ca.crt` - Certificate authority

## Signing Options

### GPG Signing
```bash
--sign-by FINGERPRINT
--sign-passphrase-file /path/to/passphrase
```

### Sigstore Signing
```bash
--sign-by-sigstore /path/to/params.yaml
--sign-by-sigstore-private-key /path/to/cosign.key
```

### Sequoia-PGP
```bash
--sign-by-sq-fingerprint FINGERPRINT
```

### Remove Signatures
```bash
--remove-signatures
```

## Output Options

```bash
--digestfile /path/to/digests.txt  # Log manifest digests
--quiet, -q                         # Suppress output
--format "{{.Digest}}"             # Go template (inspect)
```

## Platform Override

```bash
--override-arch amd64              # Force architecture
--override-os linux                # Force OS
--override-variant v8              # Force variant
```

## Global Options

```bash
--debug                            # Enable debug logging
--policy /path/to/policy.json      # Trust policy
--insecure-policy                  # Skip policy checks
--command-timeout 5m               # Operation timeout
--tmpdir /tmp                      # Temporary directory
```

## Error Codes

- **retryable** - Transient failure, retry automatically
- **EPIPE** - Broken pipe (client disconnected)
- **other** - Non-retryable errors

## Useful Patterns

### Check if Image Exists
```bash
if skopeo inspect docker://registry/image:tag &>/dev/null; then
  echo "Image exists"
fi
```

### Get Digest Programmatically
```bash
DIGEST=$(skopeo inspect --format '{{.Digest}}' docker://image)
```

### Copy to Multiple Registries
```bash
for registry in reg1 reg2 reg3; do
  skopeo copy docker://source docker://${registry}/dest &
done
wait
```

### Mirror with Different Tag
```bash
skopeo copy docker://source/app:v1.0 docker://dest/app:latest
```

### Extract Manifest
```bash
skopeo inspect --raw docker://image > manifest.json
```

### Offline Signing
```bash
# Export manifest
skopeo inspect --raw docker://image > manifest.json

# Sign offline
skopeo standalone-sign manifest.json \
      docker://registry/image:tag \
      GPG_FINGERPRINT \
      -o signature.sig

# Verify offline
skopeo standalone-verify manifest.json \
      docker://registry/image:tag \
      GPG_FINGERPRINT \
      signature.sig
```

## Common Use Cases

### CI/CD Pipeline
```bash
#!/bin/bash
IMAGE="myapp"
TAG="${CI_COMMIT_TAG}"

# Push to registry
skopeo copy \
  --dest-creds "${CI_REGISTRY_USER}:${CI_REGISTRY_PASSWORD}" \
  containers-storage:${IMAGE}:${TAG} \
  docker://${CI_REGISTRY}/${IMAGE}:${TAG}

# Write digest
skopeo inspect --format '{{.Digest}}' \
  docker://${CI_REGISTRY}/${IMAGE}:${TAG} > digest.txt
```

### Air-Gap Transfer
```bash
# Export
skopeo sync --src docker --dest dir --scoped \
            docker.io/library/alpine \
            /usb/images

# Import
skopeo sync --src dir --dest docker \
            /usb/images \
            private-registry.local
```

### Registry Migration
```bash
#!/bin/bash
SOURCE="old-registry.io"
DEST="new-registry.io"
REPOS="app1 app2 app3"

for repo in ${REPOS}; do
  echo "Migrating ${repo}"

  # Get all tags
  TAGS=$(skopeo list-tags docker://${SOURCE}/${repo} | jq -r '.Tags[]')

  for tag in ${TAGS}; do
    echo "  Copying ${repo}:${tag}"
    skopeo copy --retry-times 3 \
                docker://${SOURCE}/${repo}:${tag} \
                docker://${DEST}/${repo}:${tag}
  done
done
```

### Cleanup Old Tags
```bash
#!/bin/bash
REPO="registry.io/myapp"
KEEP_LATEST=10

# Get all tags sorted by creation date (requires API or inspection)
ALL_TAGS=$(skopeo list-tags docker://${REPO} | jq -r '.Tags[]')

# Keep only latest N tags (simple approach - delete rest)
echo "${ALL_TAGS}" | tail -n +$((KEEP_LATEST + 1)) | while read tag; do
  skopeo delete docker://${REPO}:${tag}
done
```

## Freightliner-Specific Patterns

### Multi-Registry Push
```bash
# Push to primary, secondary, and backup registries
REGISTRIES="primary.io secondary.io backup.io"
IMAGE="myapp:v1.0"

for registry in ${REGISTRIES}; do
  skopeo copy --retry-times 3 \
              docker://source/${IMAGE} \
              docker://${registry}/${IMAGE} &
done
wait

# Verify all succeeded
for registry in ${REGISTRIES}; do
  if ! skopeo inspect docker://${registry}/${IMAGE} &>/dev/null; then
    echo "ERROR: Failed to push to ${registry}"
    exit 1
  fi
done
```

### Registry Health Check
```bash
check_registry() {
  local registry=$1
  if skopeo inspect docker://${registry}/hello-world:latest &>/dev/null; then
    return 0
  else
    return 1
  fi
}

if check_registry "primary-registry.io"; then
  echo "Primary registry is healthy"
else
  echo "WARNING: Primary registry is down, using fallback"
fi
```

---

## Architecture Patterns for Implementation

### Transport Interface
```go
type Transport interface {
    Name() string
    ParseReference(string) (Reference, error)
}

type Reference interface {
    Transport() Transport
    NewImageSource(context.Context, *SystemContext) (ImageSource, error)
    NewImageDestination(context.Context, *SystemContext) (ImageDestination, error)
    DeleteImage(context.Context, *SystemContext) error
}
```

### System Context
```go
type SystemContext struct {
    // Auth
    AuthFilePath     string
    DockerAuthConfig *DockerAuthConfig

    // TLS
    DockerCertPath              string
    DockerInsecureSkipTLSVerify OptionalBool

    // Platform
    ArchitectureChoice string
    OSChoice           string

    // Storage
    OCISharedBlobDirPath string
}
```

### Retry Wrapper
```go
func RetryIfNecessary(ctx context.Context, fn func() error, opts *RetryOptions) error {
    for i := 0; i <= opts.MaxRetry; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        if !isRetryable(err) {
            return err
        }
        if i < opts.MaxRetry {
            time.Sleep(calculateBackoff(i, opts.Delay))
        }
    }
    return errors.New("max retries exceeded")
}
```

---

**Quick Reference Complete** ✅

For full details, see:
- `skopeo-features.md` - Complete feature catalog
- `skopeo-analysis-summary.md` - Executive summary
- `skopeo-key-findings.json` - Structured data
