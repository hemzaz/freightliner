# Add Feature Command

Implement a new feature for Freightliner following best practices and the project's architecture patterns.

## What This Command Does

1. Creates a detailed feature design following the project's architecture
2. Implements the feature with proper interfaces and error handling
3. Adds comprehensive tests (unit + integration)
4. Updates documentation
5. Ensures all quality gates pass

## Usage

```bash
/add-feature <feature-description>
```

## Example

```bash
/add-feature Add support for Azure Container Registry (ACR) replication
```

## Process

### 1. Design Phase
- Create feature specification in `.claude/specs/features/`
- Define interfaces following `pkg/interfaces/` patterns
- Plan package structure and dependencies
- Document security and performance considerations

### 2. Implementation Phase
- Create package structure following project conventions
- Implement interfaces with proper error handling
- Use context.Context for all operations
- Follow Options pattern for configuration
- Add structured logging with logger.WithFields()
- Implement proper resource cleanup

### 3. Testing Phase
- Write table-driven tests with subtests
- Achieve minimum 80% code coverage
- Create integration tests if needed
- Test error conditions and edge cases

### 4. Documentation Phase
- Add GoDoc comments to all exported types/functions
- Update README.md if user-facing feature
- Create usage examples
- Update ARCHITECTURE.md if architectural changes

### 5. Quality Assurance
- Run `make quality` (fmt, vet, lint, security)
- Run `make test-ci` for full test suite
- Verify no regressions in existing functionality
- Check for security vulnerabilities with gosec

## Architecture Constraints

- Must use interface-driven design
- Must support context cancellation
- Must include structured logging
- Must include Prometheus metrics if applicable
- Must follow error wrapping patterns
- Must be cloud-agnostic where possible

## Code Quality Requirements

- All exported symbols must have GoDoc comments
- Error messages must include context
- No hardcoded credentials
- Proper resource cleanup with defer
- Thread-safe if used concurrently
