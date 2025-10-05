# Go Lint - Run Code Quality Checks

Run linting and static analysis for the Freightliner project.

## Usage

```bash
# Run all quality checks
make quality

# Run only linter
make lint

# Run go vet
make vet

# Format code
make fmt

# Check formatting without modifying
make fmt-check

# Run security scan
make security
```

## Linters Configured

The project uses **golangci-lint v1.62.2** with focused checks:

- **errcheck** - Detect unchecked errors (critical for reliability)
- **govet** - Standard Go vet checks (real bug detection)
- **ineffassign** - Ineffectual assignment detection
- **misspell** - Spelling mistake detection

## Individual Tools

**golangci-lint**
```bash
golangci-lint run --timeout=8m
golangci-lint run ./pkg/client/...  # Specific package
golangci-lint run --fix             # Auto-fix issues
```

**go vet**
```bash
go vet ./...
go vet ./pkg/replication/
```

**gofmt**
```bash
gofmt -w .          # Format all files
gofmt -l .          # List unformatted files
gofmt -d file.go    # Show diff
```

**gosec** (security scanner)
```bash
gosec ./...
gosec -fmt=json -out=results.json ./...
```

## Configuration Files

- `.golangci.yml` - golangci-lint configuration
- `.gitleaks.toml` - Secret scanning config
- `.goimportsignore` - Files to ignore for import formatting

## Quality Standards

The project maintains:
- ✅ 0 linting issues in CI
- ✅ All code properly formatted
- ✅ No unchecked errors
- ✅ No security vulnerabilities

## Troubleshooting

**Linter takes too long:**
- Timeout configured to 8 minutes
- Use `--timeout=15m` if needed

**False positives:**
- Use `//nolint:lintername` comment
- Document why lint is disabled

**Linter not installed:**
```bash
make tools  # Install all dev tools
```
