# Go Build - Build the Application

Build the Freightliner application with various configurations.

## Usage

```bash
# Standard build
make build

# Build with race detection (for debugging)
make build-race

# Build static binary for Linux
make build-static

# Build release binaries (multi-platform)
make release-build

# Quick manual build
go build -o bin/freightliner .

# Build with specific flags
CGO_ENABLED=0 go build -ldflags "-w -s" -o bin/freightliner .
```

## Build Targets

**make build**
- Builds standard binary to `bin/freightliner`
- CGO disabled, static linking
- Includes version, build time, git commit

**make build-race**
- Builds with race detector enabled
- Requires CGO_ENABLED=1
- Output: `bin/freightliner-race`
- Use for debugging concurrency issues

**make build-static**
- Builds static binary for Linux (amd64)
- Fully static, no external dependencies
- Ideal for container deployment

**make release-build**
- Builds for multiple platforms: linux, darwin, windows
- Architectures: amd64, arm64
- Output: `dist/freightliner-<os>-<arch>[.exe]`

## Build Flags

The Makefile uses these LDFLAGS:
```bash
-w -s                           # Strip debug info (smaller binary)
-X main.version=$(VERSION)      # Inject version
-X main.buildTime=$(BUILD_TIME) # Inject build timestamp
-X main.gitCommit=$(GIT_COMMIT) # Inject git commit
```

## Version Information

```bash
# Check version info
make version

# Run binary to see version
./bin/freightliner version
```

## Troubleshooting

**Build fails with module errors:**
```bash
make deps  # Download and verify dependencies
```

**Build is slow:**
- Check build cache: `go env GOCACHE`
- Clean cache: `go clean -cache`

**Need to debug build:**
```bash
go build -v -x -o bin/freightliner .  # Verbose + show commands
```
