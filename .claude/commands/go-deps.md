# Go Dependencies - Manage Project Dependencies

Manage Go module dependencies for the Freightliner project.

## Usage

```bash
# Download and verify all dependencies
make deps

# Update dependencies
go get -u && go mod tidy

# Update specific dependency
go get -u github.com/spf13/cobra@latest

# Add new dependency
go get github.com/new/package@v1.2.3

# Remove unused dependencies
go mod tidy

# Verify dependencies
go mod verify

# View dependency graph
go mod graph

# List all modules
go list -m all

# Check for available updates
go list -u -m all
```

## Dependency Files

- `go.mod` - Module definition and requirements
- `go.sum` - Checksums for dependency verification

## Key Dependencies

**Core Libraries:**
- `github.com/spf13/cobra` v1.9.1 - CLI framework
- `github.com/google/go-containerregistry` v0.20.3 - OCI registry
- `github.com/stretchr/testify` v1.10.0 - Testing

**AWS Integration:**
- `github.com/aws/aws-sdk-go-v2` v1.36.3
- `github.com/aws/aws-sdk-go-v2/service/ecr` v1.43.0
- `github.com/aws/aws-sdk-go-v2/service/kms` v1.38.1

**GCP Integration:**
- `cloud.google.com/go/secretmanager` v1.14.6
- `cloud.google.com/go/kms` v1.21.1
- `google.golang.org/api` v0.228.0

**Observability:**
- `github.com/prometheus/client_golang` v1.21.1

## Maintenance Tasks

**Security Updates:**
```bash
# Check for vulnerabilities
go list -json -m all | go run golang.org/x/vuln/cmd/govulncheck@latest -json -

# Update vulnerable dependencies
go get -u <vulnerable-package>@<fixed-version>
go mod tidy
```

**Dependency Audit:**
```bash
# List direct dependencies only
go list -m -f '{{if not .Indirect}}{{.Path}} {{.Version}}{{end}}' all

# Find why a package is needed
go mod why github.com/some/package

# Show full dependency tree
go mod graph | grep package-name
```

**Clean Module Cache:**
```bash
go clean -modcache
```

## Troubleshooting

**Module not found errors:**
```bash
go mod download
go mod verify
```

**Incompatible versions:**
```bash
go get package@version  # Pin specific version
go mod tidy             # Clean up
```

**Dependency conflicts:**
```bash
go get -u ./...         # Update all
go mod tidy             # Resolve conflicts
```

**Build cache issues:**
```bash
go clean -cache -modcache -testcache
make deps
```
