# Freightliner CLI Quick Reference

## Advanced Commands Cheat Sheet

### Inspect Images
```bash
# Basic inspect
freightliner inspect docker://nginx:latest

# Show config
freightliner inspect --config docker://nginx:latest

# JSON output
freightliner inspect --format json docker://nginx:latest

# Raw manifest
freightliner inspect --raw docker://nginx:latest
```

### List Tags
```bash
# List all tags
freightliner list-tags docker://nginx

# Limit results
freightliner list-tags --limit 10 docker://nginx

# Simple output (one per line)
freightliner list-tags --format simple docker://nginx

# Sort alphabetically
freightliner list-tags --sort alpha docker://nginx
```

### Delete Images
```bash
# Delete with confirmation
freightliner delete docker://registry.io/repo:tag

# Force delete
freightliner delete --force docker://registry.io/repo:tag

# Dry run
freightliner delete --dry-run docker://registry.io/repo:tag

# Delete all tags (DANGEROUS!)
freightliner delete --all --force docker://registry.io/repo
```

### Sync Images
```bash
# Basic sync
freightliner sync --config sync.yaml

# Dry run
freightliner sync --config sync.yaml --dry-run

# Override parallelism
freightliner sync --config sync.yaml --parallel 10
```

## Common Patterns

### Mirror Repository
```yaml
# sync.yaml
source:
  registry: "docker.io"
destination:
  registry: "my-registry.io"
images:
  - repository: "library/nginx"
    all_tags: true
```

### Selective Sync with Regex
```yaml
images:
  - repository: "library/python"
    tag_regex: "^3\\.(10|11|12)-slim$"
```

### Rename Repository
```yaml
images:
  - repository: "library/redis"
    destination_repository: "cache/redis"
    tags: ["latest", "7.0"]
```

## Authentication

### Using Config File
```bash
freightliner inspect --config registries.yaml docker://private.io/app:v1
```

### Using Environment Variables
```bash
export DOCKER_USERNAME=user
export DOCKER_PASSWORD=pass
freightliner sync --config sync.yaml
```

## Output Formats

All commands support multiple output formats:
- `table` - Human-readable tables (default)
- `json` - JSON format for scripting
- `yaml` - YAML format
- `simple` - Plain text (list-tags only)

## Exit Codes

- `0` - Success
- `1` - General error
- Interrupts handled gracefully (SIGINT, SIGTERM)

## Logging

Control log level:
```bash
freightliner --log-level debug inspect docker://nginx:latest
```

Levels: debug, info, warn, error

## Tips

1. **Use dry-run first**: Always test with `--dry-run` before actual operations
2. **Authentication**: Configure registries in YAML for repeated use
3. **Parallel operations**: Adjust `--parallel` based on network and registry limits
4. **Regex testing**: Test regex patterns with dry-run before sync
5. **Error handling**: Use `continue_on_error: true` in sync for resilience
