# Skopeo Feature Catalog - Comprehensive Analysis

**Analysis Date:** 2025-12-05
**Skopeo Version:** Latest from repository
**Analyzed By:** Research Agent

---

## Executive Summary

Skopeo is a command-line utility for performing operations on container images and registries without requiring a container runtime daemon. It provides a comprehensive set of tools for image inspection, copying, synchronization, deletion, authentication, and signature management across multiple transport mechanisms.

### Key Capabilities
- **Multi-transport support**: docker://, oci://, dir://, docker-archive://, oci-archive://, containers-storage:, docker-daemon:
- **Registry operations**: Push, pull, inspect, delete, list tags
- **Authentication**: Login/logout with credential management
- **Image manipulation**: Copy, sync, convert between formats
- **Security**: GPG and Sigstore signing/verification
- **Manifest handling**: Docker v2 Schema 1/2, OCI Image Format

---

## 1. Command Inventory

### 1.1 Core Commands

#### **copy** - Image Copy and Conversion
**Purpose:** Copy images between different storage mechanisms and registries with optional transformations.

**Full Command Signature:**
```bash
skopeo copy [OPTIONS] SOURCE-IMAGE DESTINATION-IMAGE
```

**Key Options:**
- **Source/Destination Flags:**
  - `--src-authfile PATH` - Source registry auth file
  - `--src-creds USERNAME[:PASSWORD]` - Source credentials
  - `--src-username`, `--src-password` - Split credentials
  - `--src-registry-token TOKEN` - Bearer token authentication
  - `--src-cert-dir PATH` - TLS certificates directory
  - `--src-tls-verify` - Enable TLS verification
  - `--src-no-creds` - Anonymous access
  - `--src-shared-blob-dir DIR` - Shared OCI blob directory
  - `--src-daemon-host HOST` - Docker daemon connection
  - `--dest-*` - Corresponding destination flags

- **Transformation Options:**
  - `--format, -f [oci|v2s1|v2s2]` - Force manifest type conversion
  - `--all, -a` - Copy all images in a multi-arch list
  - `--multi-arch [system|all|index-only]` - Multi-architecture handling
  - `--preserve-digests` - Maintain original digests

- **Compression Control:**
  - `--dest-compress` - Compress layers (dir: transport)
  - `--dest-decompress` - Decompress layers
  - `--dest-compress-format FORMAT` - Compression algorithm
  - `--dest-compress-level LEVEL` - Compression level
  - `--force-compress-format` - Exclusive compression format
  - `--dest-oci-accept-uncompressed-layers` - Allow uncompressed layers

- **Encryption (Experimental):**
  - `--encryption-key jwe:/path/to/key.pem` - Encryption keys
  - `--encrypt-layer [0,-1,...]` - Layer indices to encrypt
  - `--decryption-key PATH` - Decryption keys

- **Signing:**
  - `--sign-by FINGERPRINT` - GPG signing
  - `--sign-by-sq-fingerprint FP` - Sequoia-PGP signing
  - `--sign-by-sigstore PATH` - Sigstore parameter file
  - `--sign-by-sigstore-private-key PATH` - Sigstore private key
  - `--sign-passphrase-file PATH` - Passphrase file
  - `--remove-signatures` - Strip existing signatures
  - `--sign-identity REF` - Override signing identity

- **Performance:**
  - `--image-parallel-copies N` - Parallel layer downloads/uploads
  - `--dest-precompute-digests` - Precompute for deduplication

- **Output:**
  - `--digestfile PATH` - Write manifest digest to file
  - `--additional-tag TAG` - Additional tags (docker-archive:)
  - `--quiet, -q` - Suppress output

- **Retry:**
  - `--retry-times N` - Number of retries
  - `--retry-delay DURATION` - Fixed retry delay (or exponential backoff)

**Examples:**
```bash
# Basic copy
skopeo copy docker://quay.io/skopeo/stable:latest docker://registry.example.com/skopeo:latest

# Copy with authentication
skopeo copy --src-creds user:pass docker://private/image \
            --dest-creds admin:secret docker://registry/image

# Multi-arch copy
skopeo copy --all docker://alpine:latest dir:/tmp/alpine-multi

# Convert format
skopeo copy --format v2s2 docker://old/image docker://new/image

# Encrypt specific layers
skopeo copy --encryption-key jwe:/keys/public.pem --encrypt-layer 0 -1 \
            docker://source oci:/dest

# Sign during copy
skopeo copy --sign-by ABCD1234 docker://unsigned docker://signed
```

---

#### **inspect** - Image Metadata Inspection
**Purpose:** Retrieve detailed information about an image without downloading it.

**Full Command Signature:**
```bash
skopeo inspect [OPTIONS] IMAGE-NAME
```

**Key Options:**
- `--raw` - Output raw manifest JSON
- `--config` - Output configuration blob
- `--format, -f TEMPLATE` - Go template formatting
- `--no-tags, -n` - Skip tag enumeration
- `--manifest-digest [sha256|sha512]` - Digest algorithm
- All authentication flags (--creds, --cert-dir, --tls-verify, etc.)

**Output Fields:**
- `Name` - Full image reference
- `Digest` - Manifest digest
- `RepoTags` - Available tags in repository
- `Created` - Creation timestamp
- `DockerVersion` - Docker version used to build
- `Labels` - Image labels
- `Architecture` - CPU architecture
- `Os` - Operating system
- `Layers` - Layer digest array
- `LayersData` - Detailed layer information
- `Env` - Environment variables

**Examples:**
```bash
# Basic inspect
skopeo inspect docker://registry.fedoraproject.org/fedora

# Get raw manifest
skopeo inspect --raw docker://alpine:latest

# Extract specific fields
skopeo inspect --format "{{.Digest}}" docker://nginx:latest

# Inspect with custom digest algorithm
skopeo inspect --manifest-digest sha512 docker://image
```

---

#### **sync** - Bulk Image Synchronization
**Purpose:** Synchronize multiple images from source to destination with filtering capabilities.

**Full Command Signature:**
```bash
skopeo sync [OPTIONS] --src TRANSPORT --dest TRANSPORT SOURCE DESTINATION
```

**Supported Source Transports:**
- `docker` - Docker registry (single repo or image)
- `dir` - Local directory structure
- `yaml` - YAML configuration file

**Supported Destination Transports:**
- `docker` - Docker registry
- `dir` - Local directory structure

**Key Options:**
- `--src, -s TRANSPORT` - Source transport type (required)
- `--dest, -d TRANSPORT` - Destination transport type (required)
- `--scoped` - Preserve full source path at destination
- `--append-suffix SUFFIX` - Add suffix to destination tags
- `--all, -a` - Copy all images in lists
- `--dry-run` - Preview operations without executing
- `--keep-going` - Continue on errors
- `--digestfile PATH` - Log digests and references
- All authentication and retry flags

**YAML Configuration Format:**
```yaml
registry.example.com:
  images:
    busybox: []                    # All tags
    alpine: ["3.14", "3.15"]       # Specific tags
    nginx: ["sha256:abc123..."]    # By digest

  images-by-tag-regex:
    golang: "^1\\.(19|20)\\."      # Regex matching

  images-by-semver:
    postgresql: ">=12.0 <15.0"     # Semantic versioning

  credentials:
    username: myuser
    password: mypass

  tls-verify: true
  cert-dir: /path/to/certs

quay.io:
  images:
    prometheus/prometheus: ["v2.30.0", "v2.31.0"]
  tls-verify: false
```

**Filtering Mechanisms:**
1. **Explicit Tags/Digests:** List specific references
2. **Regex Matching:** Pattern-based tag selection
3. **Semver Constraints:** Version range filtering

**Examples:**
```bash
# Sync single repository
skopeo sync --src docker --dest dir --scoped \
            registry.example.com/busybox /media/usb

# Sync from YAML config
skopeo sync --src yaml --dest docker \
            sync-config.yaml registry.backup.com

# Dry run with digest logging
skopeo sync --src docker --dest dir --dry-run \
            --digestfile sync.log quay.io/app /backup
```

---

#### **delete** - Image Deletion
**Purpose:** Remove images from registries or local storage.

**Full Command Signature:**
```bash
skopeo delete [OPTIONS] IMAGE-NAME
```

**Key Options:**
- All authentication flags
- Retry flags

**Supported Transports:**
- `docker://` - Docker registry (v2 API)
- `containers-storage:` - Local container storage
- `oci:` - OCI layout directory

**Examples:**
```bash
# Delete from registry
skopeo delete docker://registry.example.com/old-image:v1.0

# Delete with authentication
skopeo delete --creds admin:pass docker://private/repo:tag
```

---

#### **login** - Registry Authentication
**Purpose:** Authenticate to container registries and store credentials.

**Full Command Signature:**
```bash
skopeo login [OPTIONS] REGISTRY
```

**Key Options:**
- `--username, -u USERNAME` - Registry username
- `--password, -p PASSWORD` - Registry password (insecure)
- `--password-stdin` - Read password from stdin
- `--authfile PATH` - Custom auth file location
- `--tls-verify` - Enable TLS verification
- `--cert-dir PATH` - Certificate directory
- `--get-login` - Retrieve current credentials

**Credential Storage:**
- Default: `${XDG_RUNTIME_DIR}/containers/auth.json`
- Format: Docker-compatible JSON
- Supports repository-scoped authentication

**Examples:**
```bash
# Interactive login
skopeo login quay.io

# Non-interactive with credentials
echo "$PASSWORD" | skopeo login -u user --password-stdin registry.io

# Custom auth file
skopeo login --authfile /tmp/auth.json private-registry.com
```

---

#### **logout** - Registry Logout
**Purpose:** Remove stored registry credentials.

**Full Command Signature:**
```bash
skopeo logout [OPTIONS] REGISTRY
```

**Key Options:**
- `--authfile PATH` - Custom auth file
- `--all, -a` - Remove all credentials

**Examples:**
```bash
# Logout from specific registry
skopeo logout quay.io

# Clear all credentials
skopeo logout --all
```

---

#### **list-tags** - Tag Enumeration
**Purpose:** List all tags for a repository.

**Full Command Signature:**
```bash
skopeo list-tags [OPTIONS] SOURCE-IMAGE
```

**Supported Transports:**
- `docker://` - Docker registry
- `docker-archive:` - Docker tar archives

**Output Format:**
```json
{
    "Repository": "docker.io/library/alpine",
    "Tags": [
        "3.14",
        "3.15",
        "latest"
    ]
}
```

**Examples:**
```bash
# List tags from registry
skopeo list-tags docker://docker.io/fedora

# List from archive
skopeo list-tags docker-archive:/path/to/image.tar
```

---

### 1.2 Signature Management Commands

#### **standalone-sign** - Offline Signing
**Purpose:** Create detached signatures for manifests using local files.

**Command:**
```bash
skopeo standalone-sign MANIFEST DOCKER-REFERENCE KEY-FINGERPRINT \
                       --output SIGNATURE [--passphrase-file PATH]
```

**Process:**
1. Read manifest from local file
2. Sign using GPG key
3. Write detached signature file

---

#### **standalone-verify** - Offline Verification
**Purpose:** Verify detached signatures using local files.

**Command:**
```bash
skopeo standalone-verify MANIFEST DOCKER-REFERENCE KEY-FINGERPRINTS SIGNATURE \
                         [--public-key-file PATH]
```

**Verification:**
- Supports comma-separated fingerprint lists
- `any` keyword to accept any key from public key file
- Uses local GPG keyring if no public key file specified

---

#### **generate-sigstore-key** - Sigstore Key Generation
**Purpose:** Generate Sigstore-compatible signing keys.

**Command:**
```bash
skopeo generate-sigstore-key [OPTIONS]
```

---

#### **manifest-digest** - Digest Calculation
**Purpose:** Compute manifest digests.

**Command:**
```bash
skopeo manifest-digest [MANIFEST-FILE]
```

---

### 1.3 Experimental Commands

#### **proxy** (Experimental)
**Purpose:** Act as an image proxy for programmatic access.

**Features:**
- Socket-based IPC protocol
- JSON-RPC style API
- Blob streaming over pipes
- Image caching and lifecycle management

**Protocol Version:** 0.2.8

**Methods:**
- `Initialize` - Setup proxy context
- `OpenImage` - Open image reference
- `CloseImage` - Release image handle
- `GetManifest` - Retrieve manifest
- `GetFullConfig` - Get configuration blob
- `GetLayerInfo` - Layer metadata
- `GetRawBlob` - Stream blob data
- `FinishPipe` - Complete pipe operation

---

## 2. Transport Mechanisms

### 2.1 Supported Transports

| Transport | Read | Write | Use Case |
|-----------|------|-------|----------|
| **docker://** | ✅ | ✅ | Remote Docker Registry v2 API |
| **oci:** | ✅ | ✅ | OCI Image Layout directory |
| **dir:** | ✅ | ✅ | Directory with manifest and layers |
| **docker-archive:** | ✅ | ✅ | Docker save/load tar format |
| **oci-archive:** | ✅ | ✅ | OCI tar archive |
| **containers-storage:** | ✅ | ✅ | Local container storage (Podman/Buildah) |
| **docker-daemon:** | ✅ | ⚠️ | Docker daemon via socket |

### 2.2 Transport-Specific Features

#### **docker:// Transport**
- Full Docker Registry HTTP API v2 support
- Authentication: Basic, Bearer token, OAuth
- Manifest formats: Schema 1, Schema 2, OCI
- Multi-architecture manifest lists
- Blob mounting for efficient cross-repository copying
- Resumable uploads
- Registry catalog operations
- Signature storage integration

**Registry Support Matrix:**
| Registry | Authentication | TLS | Signatures |
|----------|---------------|-----|------------|
| Docker Hub | ✅ | ✅ | ✅ |
| Quay.io | ✅ | ✅ | ✅ |
| GCR | ✅ Bearer | ✅ | ✅ |
| ECR | ✅ Bearer | ✅ | ⚠️ Limited |
| ACR | ✅ | ✅ | ✅ |
| Harbor | ✅ | ✅ | ✅ |
| Artifactory | ✅ | ✅ | ✅ |
| GitLab CR | ✅ | ✅ | ✅ |

#### **oci: Transport**
- OCI Image Layout Specification v1.0+
- `index.json` as entry point
- `blobs/sha256/` content-addressed storage
- Supports image and index manifests

#### **dir: Transport**
- Simple directory structure
- `manifest.json` + layer tarballs
- Optional compression control
- Useful for offline transfers

#### **docker-archive: Transport**
- Compatible with `docker save`
- Multiple images in single tarball
- Preserves repository:tag information

#### **containers-storage: Transport**
- Direct access to containers/storage
- Shared with Podman, Buildah, CRI-O
- Graph driver backends (overlay, vfs, etc.)

---

## 3. Authentication and Credential Handling

### 3.1 Authentication Mechanisms

#### **Username/Password (Basic Auth)**
```bash
--creds USERNAME:PASSWORD
--username USERNAME --password PASSWORD
```

#### **Bearer Token**
```bash
--registry-token TOKEN
```

#### **Anonymous Access**
```bash
--no-creds
```

#### **Credential Files**
- **Path:** `${XDG_RUNTIME_DIR}/containers/auth.json`
- **Format:** Docker-compatible JSON
- **Structure:**
```json
{
  "auths": {
    "registry.example.com": {
      "auth": "base64(username:password)"
    },
    "quay.io": {
      "auth": "base64(username:password)"
    }
  }
}
```

#### **Per-Registry Configuration**
YAML sync configuration supports per-registry credentials:
```yaml
registry.example.com:
  credentials:
    username: user
    password: pass
```

### 3.2 TLS/Certificate Management

#### **TLS Verification**
```bash
--tls-verify=true|false
```

#### **Custom Certificates**
```bash
--cert-dir /path/to/certs
```

Expected files:
- `registry.example.com.crt` - Server certificate
- `registry.example.com.cert` - Alternative name
- `registry.example.com.key` - Client key
- `ca.crt` - Certificate authority

#### **Insecure Registries**
YAML configuration:
```yaml
registry.example.com:
  tls-verify: false
```

### 3.3 Credential Security

**Best Practices:**
- Use `--password-stdin` instead of `--password`
- Store credentials in protected directories
- Use `--authfile` for isolated environments
- Leverage credential helpers (future consideration)

---

## 4. Manifest and Layer Handling

### 4.1 Supported Manifest Formats

#### **Docker Image Manifest V2, Schema 1**
- **Media Type:** `application/vnd.docker.distribution.manifest.v1+json`
- **Status:** Deprecated, but supported for compatibility
- **Features:**
  - Signed JSON structure
  - Inline configuration
  - History embedded in manifest

#### **Docker Image Manifest V2, Schema 2**
- **Media Type:** `application/vnd.docker.distribution.manifest.v2+json`
- **Features:**
  - Separate config blob
  - Content-addressable layers
  - Efficient layer deduplication

#### **OCI Image Manifest v1**
- **Media Type:** `application/vnd.oci.image.manifest.v1+json`
- **Features:**
  - OCI Image Format Specification
  - Enhanced annotations
  - Artifact support

#### **Manifest Lists (Multi-Architecture)**
- **Docker:** `application/vnd.docker.distribution.manifest.list.v2+json`
- **OCI:** `application/vnd.oci.image.index.v1+json`
- **Capabilities:**
  - Multiple platform-specific manifests
  - Automatic platform selection
  - `--all` flag to copy entire list

### 4.2 Layer Operations

#### **Layer Types**
- Standard filesystem layers (tar or tar.gz)
- Non-distributable layers (Windows base layers)
- Foreign layers (external URLs)

#### **Compression**
Supported formats:
- **gzip** (default)
- **zstd** (Zstandard)
- **bzip2**
- Uncompressed

Control:
```bash
--dest-compress-format zstd
--dest-compress-level 9
--force-compress-format  # Recompress if needed
```

#### **Layer Deduplication**
- Content-addressed storage
- `--dest-precompute-digests` for registry dedup
- Blob mounting when possible

#### **Encryption (Experimental)**
```bash
--encryption-key jwe:/path/to/public.pem
--encrypt-layer 0,-1  # First and last layer
--decryption-key /path/to/private.pem
```

Uses **ocicrypt** library with:
- JWE (JSON Web Encryption)
- PGP
- PKCS7

### 4.3 Manifest Digest Calculation

**Algorithms:**
- sha256 (default)
- sha512

**Usage:**
```bash
--manifest-digest sha512
```

**Output to File:**
```bash
--digestfile /path/to/digests.txt
```

Format per line:
```
sha256:abc123... docker://registry.io/repo:tag
```

---

## 5. Security Features

### 5.1 Signature Support

#### **GPG Signatures (Simple Signing)**
**Signing:**
```bash
--sign-by GPG_FINGERPRINT
--sign-passphrase-file /path/to/passphrase
```

**Verification:**
- Policy-based via `/etc/containers/policy.json`
- `signedBy` requirements
- Signature storage: registries.d configuration

**Storage Locations:**
- Local: `file:///var/lib/containers/sigstore`
- Remote: `https://registry.io/signatures`

#### **Sigstore Signatures**
**Signing:**
```bash
--sign-by-sigstore /path/to/params.yaml
--sign-by-sigstore-private-key /path/to/cosign.key
```

**Features:**
- Fulcio CA integration
- Rekor transparency log
- Keyless signing
- OIDC authentication

#### **Sequoia-PGP Signatures**
```bash
--sign-by-sq-fingerprint FINGERPRINT
```

### 5.2 Image Verification

#### **Trust Policy**
**Global:**
```bash
--policy /etc/containers/policy.json
--insecure-policy  # Accept anything (testing only)
```

**Policy Structure:**
```json
{
  "default": [{"type": "reject"}],
  "transports": {
    "docker": {
      "registry.example.com": [
        {
          "type": "signedBy",
          "keyType": "GPGKeys",
          "keyPath": "/path/to/pubkey.gpg"
        }
      ]
    }
  }
}
```

#### **Signature Verification Flow**
1. Download manifest
2. Check policy requirements
3. Fetch signatures from configured locations
4. Verify signatures against trusted keys
5. Accept/reject based on policy

### 5.3 Security Best Practices

**Implemented in Skopeo:**
- TLS enforcement by default
- Certificate validation
- Signature verification framework
- Credential encryption in auth files
- No setuid/setgid requirements
- Minimal privilege operation

**Recommendations:**
- Always use `--tls-verify` in production
- Implement signature verification policies
- Rotate signing keys regularly
- Use credential files with restricted permissions
- Audit registry access logs

---

## 6. Performance and Concurrency

### 6.1 Parallel Operations

#### **Image-Level Parallelism**
```bash
--image-parallel-copies N
```
- Parallel layer downloads/uploads
- Default: Sequential (safe)
- Recommended: 4-8 for high-bandwidth connections

#### **Sync-Level Parallelism**
- Sequential by design for reliability
- Use `--keep-going` to continue on errors

### 6.2 Optimization Features

#### **Shared Blob Directory**
```bash
--src-shared-blob-dir /path/to/blobs
--dest-shared-blob-dir /path/to/blobs
```
- Share blobs across OCI repositories
- Reduces storage duplication
- Faster local copies

#### **Registry Optimizations**
```bash
--dest-precompute-digests
```
- Calculate digests before upload
- Skip already-present layers
- Faster push to registries

#### **Blob Mounting**
- Automatic cross-repository mounting
- Docker Registry v2 API feature
- Zero-copy layer transfers within registry

### 6.3 Retry Mechanism

```bash
--retry-times 3
--retry-delay 5s  # Fixed delay
--retry-delay 0   # Exponential backoff (default)
```

**Retry Logic:**
- Network errors
- Transient registry failures
- Rate limiting (429 responses)
- Not retried: Authentication failures, not found errors

---

## 7. Registry Type Support Matrix

| Feature | Docker Hub | Quay | GCR | ECR | ACR | Harbor | Artifactory |
|---------|------------|------|-----|-----|-----|--------|-------------|
| **Push** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Pull** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Delete** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **List Tags** | ✅ | ✅ | ✅ | ⚠️ Limited | ✅ | ✅ | ✅ |
| **Signatures** | ✅ | ✅ | ✅ | ⚠️ Limited | ✅ | ✅ | ✅ |
| **Multi-Arch** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **OCI Artifacts** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

**Notes:**
- **ECR Limitations:**
  - Tag listing blocked without `ecr:ListImages` IAM permission
  - Public ECR (`public.ecr.aws`) doesn't implement tag listing endpoint
- All registries support Docker v2 API specification

---

## 8. Advanced Configuration

### 8.1 Global Options

```bash
--debug                     # Enable debug logging
--policy PATH               # Trust policy file
--insecure-policy          # Disable policy checks
--registries.d DIR         # Signature storage config
--override-arch ARCH       # Force architecture (amd64, arm64, etc.)
--override-os OS           # Force OS (linux, windows, etc.)
--override-variant VAR     # Force variant (v7, v8, etc.)
--command-timeout DUR      # Operation timeout
--tmpdir DIR               # Temporary file directory
--user-agent-prefix STR    # Custom User-Agent prefix
```

### 8.2 SystemContext Configuration

Internal structure that combines CLI flags into unified context:

```go
type SystemContext struct {
    // Authentication
    AuthFilePath              string
    DockerAuthConfig          *DockerAuthConfig
    DockerBearerRegistryToken string

    // TLS
    DockerCertPath                   string
    DockerInsecureSkipTLSVerify      OptionalBool
    DockerDaemonInsecureSkipTLSVerify bool

    // Registry
    RegistriesDirPath             string
    SystemRegistriesConfPath      string
    DockerRegistryUserAgent       string

    // Platform
    ArchitectureChoice string
    OSChoice           string
    VariantChoice      string

    // Storage
    OCISharedBlobDirPath string
    BigFilesTemporaryDir string

    // Compression
    CompressionFormat *compression.Algorithm
    CompressionLevel  *int
}
```

### 8.3 Registries Configuration

**File:** `/etc/containers/registries.conf`

**Purpose:**
- Registry mirrors
- Blocked registries
- Insecure registry list
- Short name aliases

**Example:**
```toml
[[registry]]
prefix = "docker.io"
location = "docker.io"
insecure = false

[[registry.mirror]]
location = "mirror.example.com:5000"
insecure = false

[[registry]]
location = "localhost:5000"
insecure = true
```

---

## 9. Use Cases and Patterns

### 9.1 CI/CD Integration

**Build and Push:**
```bash
# Build with buildah/docker, push with skopeo
skopeo copy --dest-creds "${CI_USER}:${CI_TOKEN}" \
            containers-storage:myapp:latest \
            docker://registry.example.com/myapp:${CI_COMMIT_TAG}
```

**Multi-Registry Mirroring:**
```bash
# Mirror to multiple registries
for reg in prod-registry1 prod-registry2 prod-registry3; do
  skopeo copy --retry-times 3 \
              docker://build-registry/app:v1.0 \
              docker://${reg}/app:v1.0
done
```

### 9.2 Air-Gapped Environments

**Export Images:**
```bash
# Create directory structure
skopeo sync --src docker --dest dir --scoped \
            registry.access.redhat.com/ubi8 \
            /media/usb/images

# Or create archives
skopeo copy docker://registry.io/app:v1 \
            oci-archive:/media/usb/app-v1.tar:v1
```

**Import Images:**
```bash
# Import to private registry
skopeo sync --src dir --dest docker \
            /media/usb/images \
            airgap-registry.local
```

### 9.3 Image Promotion

**Dev → Stage → Prod:**
```bash
#!/bin/bash
IMAGE="myapp"
VERSION="1.2.3"

# Promote from dev to stage
skopeo copy \
  docker://dev-registry/${IMAGE}:${VERSION} \
  docker://stage-registry/${IMAGE}:${VERSION}

# Sign for stage
skopeo copy --sign-by STAGE_KEY \
  docker://stage-registry/${IMAGE}:${VERSION} \
  docker://stage-registry/${IMAGE}:${VERSION}

# After testing, promote to prod
skopeo copy --sign-by PROD_KEY \
  docker://stage-registry/${IMAGE}:${VERSION} \
  docker://prod-registry/${IMAGE}:${VERSION}
```

### 9.4 Format Conversion

**Docker → OCI:**
```bash
skopeo copy --format oci \
            docker-archive:legacy-image.tar \
            oci:modern-image:latest
```

**Normalize Old Images:**
```bash
skopeo copy --format v2s2 \
            docker://old-registry/legacy:v1 \
            docker://new-registry/updated:v1
```

### 9.5 Bulk Synchronization

**YAML-Driven Sync:**
```yaml
# repos.yaml
docker.io:
  images:
    library/alpine: ["3.14", "3.15", "3.16"]
    library/nginx: []
  images-by-tag-regex:
    library/postgres: "^(12|13|14)\\."

quay.io:
  images:
    prometheus/prometheus: []
  images-by-semver:
    prometheus/alertmanager: ">=0.24.0"
```

```bash
skopeo sync --src yaml --dest docker \
            --keep-going --digestfile sync.log \
            repos.yaml backup-registry.io
```

### 9.6 Cleanup Operations

**Delete Old Tags:**
```bash
#!/bin/bash
REPO="registry.io/myapp"
KEEP_TAGS="latest v1.0 v2.0"

# Get all tags
ALL_TAGS=$(skopeo list-tags docker://${REPO} | jq -r '.Tags[]')

# Delete old tags
for tag in ${ALL_TAGS}; do
  if ! echo "${KEEP_TAGS}" | grep -qw "${tag}"; then
    echo "Deleting ${REPO}:${tag}"
    skopeo delete docker://${REPO}:${tag}
  fi
done
```

---

## 10. Integration Points

### 10.1 Library Usage

Skopeo is built on reusable libraries:

- **github.com/containers/image/v5** - Core image operations
- **github.com/containers/storage** - Container storage
- **github.com/containers/common** - Shared container utilities
- **github.com/opencontainers/go-digest** - Content addressing
- **github.com/opencontainers/image-spec** - OCI specifications

### 10.2 Ecosystem Compatibility

**Compatible Tools:**
- **Podman** - Shares storage backend
- **Buildah** - Uses same libraries
- **CRI-O** - Common storage and transport
- **Docker** - Compatible image formats and registries

### 10.3 API and Automation

**Proxy Mode (Experimental):**
- Socket-based IPC
- JSON request/response
- Blob streaming over pipes
- Suitable for language bindings

**Shell Integration:**
```bash
# Get digest programmatically
DIGEST=$(skopeo inspect --format '{{.Digest}}' docker://image)

# Check if image exists
if skopeo inspect docker://registry/image:tag &>/dev/null; then
  echo "Image exists"
fi
```

---

## 11. Limitations and Considerations

### 11.1 Known Limitations

1. **No Image Building:** Skopeo only operates on existing images
2. **No Container Execution:** No runtime functionality
3. **Limited Streaming:** Full manifests downloaded before processing
4. **Sync Parallelism:** Sync is sequential by design
5. **Windows Support:** Limited proxy functionality on Windows

### 11.2 Performance Considerations

1. **Network Bandwidth:** Primary bottleneck for remote operations
2. **Disk I/O:** Significant for local operations and caching
3. **Memory:** Large manifests and layer metadata held in memory
4. **TLS Overhead:** Certificate validation adds latency

### 11.3 Security Considerations

1. **Credential Exposure:** Use `--password-stdin` and secure auth files
2. **Registry Trust:** Always verify TLS certificates in production
3. **Signature Verification:** Enable policy-based verification
4. **Private Data:** Be cautious with debug logs containing credentials

---

## 12. Key Findings Summary

### 12.1 Strengths

1. **Daemonless Operation:** No privileged daemon required
2. **Multi-Transport Support:** 8 different transport types
3. **Comprehensive Authentication:** Multiple auth mechanisms
4. **Signature Support:** GPG, Sigstore, Sequoia-PGP
5. **Format Conversion:** Seamless manifest format translation
6. **Bulk Operations:** Powerful sync with filtering
7. **Production-Ready:** Mature codebase with extensive error handling
8. **Registry Agnostic:** Works with all Docker v2 compatible registries

### 12.2 Unique Features

1. **YAML-Driven Sync:** Declarative bulk synchronization
2. **Regex/Semver Filtering:** Advanced tag selection
3. **Standalone Signing:** Offline signature creation
4. **Proxy Mode:** Programmatic access protocol
5. **Layer Encryption:** Experimental encryption support
6. **Multi-Arch Handling:** Granular multi-architecture control

### 12.3 Architectural Patterns

1. **Transport Abstraction:** Clean separation of transport implementations
2. **SystemContext Pattern:** Unified configuration context
3. **Retry Logic:** Built-in retry mechanism with backoff
4. **Error Handling:** Comprehensive error types and codes
5. **Streaming:** Efficient blob streaming with pipes
6. **Verification:** Policy-based trust framework

---

## 13. Implementation Recommendations for Freightliner

### 13.1 Direct Applicability

**High Priority:**
1. ✅ Multi-transport architecture
2. ✅ Authentication credential management patterns
3. ✅ TLS certificate handling
4. ✅ Retry logic with exponential backoff
5. ✅ Manifest format detection and conversion

**Medium Priority:**
1. ⚠️ Signature verification framework (if signing required)
2. ⚠️ Sync filtering mechanisms (regex/semver)
3. ⚠️ Blob mounting optimization

**Low Priority:**
1. ⏸️ Proxy mode IPC (complex, experimental)
2. ⏸️ Layer encryption (experimental)
3. ⏸️ Deprecated layers command

### 13.2 Code Patterns to Adopt

```go
// SystemContext pattern for unified configuration
type RegistryContext struct {
    AuthConfig      *AuthConfig
    TLSVerify       bool
    CertPath        string
    UserAgent       string
    Timeout         time.Duration
}

// Retry wrapper
func retryOperation(ctx context.Context, opts *retry.Options, fn func() error) error {
    return retry.IfNecessary(ctx, fn, opts)
}

// Transport abstraction
type Transport interface {
    Name() string
    ParseReference(string) (Reference, error)
}

type Reference interface {
    Transport() Transport
    NewImageSource(context.Context, *SystemContext) (ImageSource, error)
    NewImageDestination(context.Context, *SystemContext) (ImageDestination, error)
}
```

### 13.3 Features to Implement

1. **Registry Configuration:**
   - Support registries.yaml similar to Skopeo's sync
   - Per-registry TLS and auth settings
   - Mirror and fallback support

2. **Authentication:**
   - Credential file support (Docker-compatible)
   - Multiple auth mechanisms (basic, bearer, token)
   - Secure credential storage

3. **Multi-Registry Operations:**
   - Parallel copying to multiple registries
   - Health checks before operations
   - Digest verification

4. **Error Handling:**
   - Distinguish retryable vs non-retryable errors
   - Structured error types
   - Comprehensive logging

---

## Appendix A: Transport Reference Quick Guide

```
# Docker Registry
docker://registry.io/repo:tag
docker://registry.io/repo@sha256:abc123

# OCI Layout
oci:/path/to/layout:tag
oci:/path/to/layout:tag@sha256:abc123

# Directory
dir:/path/to/directory

# Docker Archive
docker-archive:/path/to/file.tar:repo:tag

# OCI Archive
oci-archive:/path/to/file.tar:tag

# Container Storage
containers-storage:image:tag
containers-storage:[backend@]graph-root+run-root:image:tag

# Docker Daemon
docker-daemon:image:tag
```

---

## Appendix B: Authentication Examples

```bash
# Method 1: Inline credentials (insecure)
skopeo copy --src-creds user:pass docker://private/image oci:output

# Method 2: Split credentials
skopeo copy --src-username user --src-password pass docker://image oci:out

# Method 3: Auth file
skopeo copy --authfile /secure/auth.json docker://image oci:out

# Method 4: Bearer token
skopeo copy --registry-token "$TOKEN" docker://image oci:out

# Method 5: Password from stdin
echo "$PASSWORD" | skopeo login -u user --password-stdin registry.io
skopeo copy docker://registry.io/private/image oci:output

# Method 6: Anonymous
skopeo copy --src-no-creds docker://public/image oci:output
```

---

## Appendix C: Manifest Format Details

### Docker v2 Schema 1 (Deprecated)
```json
{
   "schemaVersion": 1,
   "name": "library/alpine",
   "tag": "latest",
   "architecture": "amd64",
   "fsLayers": [
      {"blobSum": "sha256:..."}
   ],
   "history": [...]
}
```

### Docker v2 Schema 2
```json
{
   "schemaVersion": 2,
   "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
   "config": {
      "mediaType": "application/vnd.docker.container.image.v1+json",
      "size": 1234,
      "digest": "sha256:..."
   },
   "layers": [
      {
         "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
         "size": 5678,
         "digest": "sha256:..."
      }
   ]
}
```

### OCI Image Manifest
```json
{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "config": {
    "mediaType": "application/vnd.oci.image.config.v1+json",
    "size": 1234,
    "digest": "sha256:..."
  },
  "layers": [
    {
      "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      "size": 5678,
      "digest": "sha256:..."
    }
  ],
  "annotations": {
    "org.opencontainers.image.created": "2023-01-01T00:00:00Z"
  }
}
```

---

**End of Analysis**

This comprehensive catalog covers all major features, options, and capabilities of Skopeo discovered through source code analysis. The tool demonstrates mature patterns for registry operations that can inform the Freightliner implementation.
