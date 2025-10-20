# Fix CI Command

Diagnose and fix CI/CD pipeline failures in GitHub Actions.

## What This Command Does

1. Analyzes recent GitHub Actions workflow runs
2. Identifies failure patterns (tests, lint, security, build)
3. Provides root cause analysis
4. Implements fixes following CI reliability best practices
5. Validates fixes locally before pushing

## Usage

```bash
/fix-ci [workflow-name]
```

## Example

```bash
/fix-ci comprehensive-validation
```

## Diagnostic Steps

### 1. Identify Failure Type
- **Test Failures**: Check test output, flaky tests, race conditions
- **Lint Failures**: golangci-lint errors, style violations
- **Build Failures**: Compilation errors, dependency issues
- **Security Failures**: gosec vulnerabilities, dependency CVEs
- **Integration Failures**: Registry connectivity, timeout issues

### 2. Analyze Logs
- Fetch recent workflow run logs with `gh run list` and `gh run view`
- Identify error patterns and stack traces
- Check for environmental issues (timeouts, resources)
- Look for race conditions in test failures

### 3. Root Cause Analysis
- Determine if issue is:
  - Code quality (needs fixing)
  - Flaky test (needs stabilization)
  - CI infrastructure (needs retry/timeout adjustment)
  - Dependency issue (needs update/pin)

### 4. Implement Fix
- **Code Issues**: Fix the actual code problem
- **Flaky Tests**: Add proper synchronization, increase timeouts, use test fixtures
- **Infrastructure**: Update GitHub Actions configuration, add retries
- **Dependencies**: Update go.mod, verify with `go mod verify`

### 5. Validate Locally
```bash
make quality      # Run all quality checks
make test-ci      # Run CI test suite
make security     # Run security scans
make build        # Verify build succeeds
```

### 6. CI Configuration Improvements
- Add retry logic for flaky operations
- Increase timeouts for integration tests
- Add proper cleanup in test fixtures
- Implement circuit breakers for external services

## Common Fixes

### Flaky Tests
- Add proper test isolation
- Use test fixtures with cleanup
- Implement retry logic with exponential backoff
- Add timeout management

### Lint Failures
- Run `make fmt` to auto-format code
- Fix reported issues from `golangci-lint run`
- Update .golangci.yml if false positives

### Build Failures
- Run `go mod tidy` to clean dependencies
- Check for breaking changes in dependencies
- Verify Go version compatibility

### Integration Test Timeouts
- Increase TEST_TIMEOUT in Makefile
- Add health checks before running tests
- Implement proper retry mechanisms
