# Go Test - Run Project Tests

Run Go tests with various options for the Freightliner project.

## Usage

```bash
# Run all tests
make test

# Run unit tests only (fast)
make test-unit

# Run with race detection
make test-race

# Run with coverage report
make test-coverage

# Run CI tests (race + coverage)
make test-ci

# Run specific package tests
go test -v ./pkg/client/ecr/...

# Run specific test function
go test -v ./pkg/replication -run TestWorkerPool

# Run with verbose output and timeout
go test -v -timeout=8m ./...
```

## Common Test Flags

- `-v` - Verbose output
- `-race` - Enable race detector (requires CGO_ENABLED=1)
- `-cover` - Show coverage statistics
- `-coverprofile=coverage.out` - Generate coverage profile
- `-run <pattern>` - Run tests matching pattern
- `-short` - Run only short tests (unit tests)
- `-timeout <duration>` - Set test timeout (default: 8m)
- `-parallel <n>` - Run tests in parallel (default: 4)

## Test Organization

- Unit tests: Fast, no external dependencies, use `-short` flag
- Integration tests: Named `Test*Integration`, require services
- Table-driven tests: Preferred pattern with subtests
- Mocks: Use gomock for interface mocking

## Troubleshooting

**Race detector fails in Docker:**
- Race detector requires CGO_ENABLED=1
- Set CGO_ENABLED=1 in environment

**Test timeouts:**
- Increase timeout: `go test -timeout=15m ./...`
- Default project timeout: 8 minutes

**Integration tests failing:**
- Check test registry setup: `./scripts/setup-test-registries.sh`
- Ensure AWS/GCP credentials configured
