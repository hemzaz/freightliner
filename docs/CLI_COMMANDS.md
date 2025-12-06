# Advanced CLI Commands

This document describes the advanced CLI commands implemented in Freightliner to provide Skopeo-like functionality for container image operations.

## Commands Overview

Freightliner now includes four advanced commands for container image management:

1. **inspect** - Inspect image manifest and metadata without pulling
2. **list-tags** - List all tags in a repository
3. **delete** - Delete images from registries
4. **sync** - Bulk synchronization using YAML configuration

## Command Details

### 1. Inspect Command

Inspect container images without downloading them. Displays manifest, configuration, layers, and metadata.

**Usage:**
```bash
freightliner inspect [flags] SOURCE
```

**Supported Transports:**
- `docker://` - Docker registry (default)
- `docker-daemon:` - Local Docker daemon (planned)
- `oci:` - OCI layout directory (planned)

**Flags:**
- `--config` - Show container configuration JSON
- `--raw` - Show raw manifest JSON
- `--format` - Output format: table (default), json, yaml

**Examples:**
```bash
# Inspect from Docker Hub
freightliner inspect docker://nginx:latest

# Show detailed config in JSON
freightliner inspect --config --format json docker://nginx:latest

# Show raw manifest
freightliner inspect --raw docker://myregistry.io/app:v1.0

# Inspect with authentication (from config)
freightliner inspect --config registries.yaml docker://private.registry.io/app:latest
```

**Output Information:**
- Image digest and size
- Architecture and OS
- Created timestamp
- Layers (digest, size, media type)
- Environment variables
- Labels
- Configuration (when --config flag used)
- Multi-platform information (for manifest indexes)

**File:** `/Users/elad/PROJ/freightliner/cmd/inspect.go` (295 lines)

---

### 2. List Tags Command

List all available tags in a repository.

**Usage:**
```bash
freightliner list-tags [flags] REPOSITORY
```

**Flags:**
- `--limit N` - Limit number of tags to display (0 = no limit)
- `--format` - Output format: table (default), json, yaml, simple
- `--sort` - Sort order: alpha, alpha-desc, recent

**Examples:**
```bash
# List all nginx tags
freightliner list-tags docker://nginx

# List with limit
freightliner list-tags --limit 10 docker://nginx

# Simple list (one tag per line)
freightliner list-tags --format simple docker://nginx

# Sort alphabetically
freightliner list-tags --sort alpha docker://myregistry.io/myapp

# JSON output
freightliner list-tags --format json docker://redis
```

**File:** `/Users/elad/PROJ/freightliner/cmd/list_tags.go` (233 lines)

---

### 3. Delete Command

Delete images from registries by tag or digest. **Use with caution - this operation is irreversible!**

**Usage:**
```bash
freightliner delete [flags] IMAGE
```

**Flags:**
- `--force` - Skip confirmation prompt
- `--all` - Delete all tags in repository (requires --force)
- `--dry-run` - Show what would be deleted without actually deleting

**Safety Features:**
- Confirmation prompt (unless --force used)
- Dry-run mode to preview deletions
- --all requires --force to prevent accidents

**Examples:**
```bash
# Delete single tag (with confirmation)
freightliner delete docker://myregistry.io/app:old-version

# Force delete without confirmation
freightliner delete --force docker://myregistry.io/app:v1.0

# Delete by digest
freightliner delete docker://myregistry.io/app@sha256:abc123...

# Dry run to see what would be deleted
freightliner delete --dry-run docker://myregistry.io/app:test

# Delete all tags (DANGEROUS!)
freightliner delete --all --force docker://myregistry.io/temp-repo
```

**Notes:**
- Requires authentication for most registries
- Some registries require special permissions for deletion
- Docker Hub does not support tag deletion via API
- Returns detailed summary of successful and failed deletions

**File:** `/Users/elad/PROJ/freightliner/cmd/delete.go` (274 lines)

---

### 4. Sync Command

Bulk image synchronization using YAML-driven configuration. Powerful tool for mirroring repositories, selective replication, and multi-image operations.

**Usage:**
```bash
freightliner sync --config FILE [flags]
```

**Flags:**
- `--config FILE` - Path to sync configuration file (required)
- `--dry-run` - Show what would be synced without syncing
- `--parallel N` - Override parallel workers from config

**Configuration File Structure:**
```yaml
source:
  registry: "registry-1.docker.io"
  auth:
    username: "user"
    password: "pass"

destination:
  registry: "my-registry.io"
  auth:
    username: "user"
    password: "pass"

parallel: 5
skip_existing: true
continue_on_error: true
timeout: 600

images:
  - repository: "library/nginx"
    tags: ["latest", "1.21", "1.22"]

  - repository: "library/redis"
    tag_regex: "^7\\..*"
    destination_repository: "cache/redis"

  - repository: "library/postgres"
    all_tags: true
    limit: 10
```

**Image Filter Options:**
- `repository` - Source repository path
- `tags` - List of specific tags to sync
- `tag_regex` - Regex pattern for tag matching
- `all_tags` - Sync all tags in repository
- `destination_repository` - Override destination repository path
- `destination_prefix` - Add prefix to destination tags
- `limit` - Limit number of tags to sync

**Examples:**
```bash
# Basic sync
freightliner sync --config sync.yaml

# Dry run
freightliner sync --config sync.yaml --dry-run

# Override parallelism
freightliner sync --config sync.yaml --parallel 10

# With environment variables for auth
export DOCKER_USERNAME=myuser
export DOCKER_PASSWORD=mypass
freightliner sync --config sync.yaml
```

**Features:**
- Parallel execution with configurable workers
- Regex-based tag filtering
- Repository renaming and tag prefixing
- Skip existing images optimization
- Continue on error for resilience
- Comprehensive summary reporting
- Environment variable support in config

**Example Configuration:** `/Users/elad/PROJ/freightliner/examples/sync-config.yaml`

**File:** `/Users/elad/PROJ/freightliner/cmd/sync.go` (472 lines)

---

## Authentication

All commands support authentication through:

1. **Registry Configuration File** (--config flag):
   ```yaml
   registries:
     - name: "my-registry"
       type: "generic"
       endpoint: "https://registry.example.com"
       auth:
         type: "basic"
         username: "user"
         password: "pass"
   ```

2. **Default Docker Keychain**: Automatically uses credentials from `~/.docker/config.json`

3. **Anonymous Access**: For public registries

## Transport Schemes

- `docker://` - Docker Registry V2 protocol (default)
- `docker-daemon:` - Local Docker daemon (planned)
- `oci:` - OCI image layout directory (planned)

If no transport is specified, `docker://` is assumed.

## Implementation Details

### Dependencies

The commands use the following libraries:

- `github.com/google/go-containerregistry` - Container registry operations
- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML parsing

### Command Registration

All commands are registered in `/Users/elad/PROJ/freightliner/cmd/root.go`:

```go
// Add new advanced CLI commands (Skopeo-like functionality)
rootCmd.AddCommand(newInspectCmd())
rootCmd.AddCommand(newListTagsCmd())
rootCmd.AddCommand(newDeleteCmd())
rootCmd.AddCommand(newSyncCmd())
```

### Error Handling

All commands implement:
- Context cancellation for graceful shutdown
- Structured logging with configurable levels
- User-friendly error messages
- Exit codes for automation

### Testing

To test the commands:

```bash
# Build the binary
go build -o freightliner

# Test help
./freightliner inspect --help
./freightliner list-tags --help
./freightliner delete --help
./freightliner sync --help

# Test with public images
./freightliner inspect docker://nginx:latest
./freightliner list-tags --limit 10 docker://nginx
```

## Comparison with Skopeo

| Feature | Skopeo | Freightliner |
|---------|--------|--------------|
| Inspect images | ✅ | ✅ |
| List tags | ✅ | ✅ |
| Delete images | ✅ | ✅ |
| Copy images | ✅ | ✅ (sync command) |
| Bulk operations | ❌ | ✅ (YAML config) |
| Regex filtering | ❌ | ✅ |
| Parallel sync | ❌ | ✅ |
| Multi-registry | ✅ | ✅ |
| Authentication | ✅ | ✅ |

## Future Enhancements

Potential future additions:

1. **docker-daemon transport**: Inspect and copy from local Docker
2. **oci transport**: Work with OCI layout directories
3. **Signature verification**: Verify image signatures
4. **SBOM integration**: Include SBOM in inspect output
5. **Copy command**: Single image copy (alternative to sync)
6. **Tag manipulation**: Retag images without copying
7. **Progress tracking**: Real-time progress for long operations
8. **Resume capability**: Resume interrupted sync operations

## Examples

See `/Users/elad/PROJ/freightliner/examples/sync-config.yaml` for comprehensive sync configuration examples.

## Support

For issues or questions:
- Check command help: `freightliner <command> --help`
- View logs with: `--log-level debug`
- Report issues at: [project issue tracker]
