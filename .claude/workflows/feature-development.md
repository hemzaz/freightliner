# Feature Development Workflow

Complete workflow for developing a new feature in Freightliner.

## Workflow Steps

### 1. Feature Planning (golang-pro + architect-reviewer)
- Create feature specification in `.claude/specs/features/`
- Design interfaces and package structure
- Document architecture decisions
- Review with architect-reviewer agent

### 2. Implementation (golang-pro)
- Create package structure following conventions
- Implement interfaces with proper error handling
- Add structured logging and metrics
- Follow Go best practices and project patterns

### 3. Testing (test-automator)
- Write table-driven unit tests
- Create integration tests if needed
- Achieve 80%+ code coverage
- Test error conditions and edge cases

### 4. Security Review (security-auditor)
- Check for credential exposure
- Review authentication/authorization
- Validate input sanitization
- Check for common vulnerabilities

### 5. Code Review (code-reviewer)
- Review code quality and style
- Check for Go anti-patterns
- Verify error handling
- Ensure proper resource cleanup

### 6. Performance Review (performance-engineer)
- Identify potential bottlenecks
- Review resource usage
- Suggest optimizations
- Run benchmarks if needed

### 7. Documentation (api-documenter)
- Add GoDoc comments
- Update README if needed
- Create usage examples
- Update architecture docs

### 8. CI/CD Integration (deployment-engineer)
- Ensure all tests pass in CI
- Verify build succeeds
- Check security scans pass
- Validate Docker build

## Parallel Agent Execution

You can run multiple agents in parallel for efficiency:

```bash
# Run design review and security audit in parallel
/feature-develop "Add ACR support" --agents="golang-pro,security-auditor,architect-reviewer"
```

## Quality Gates

All must pass before merging:
- [ ] Code compiles without errors
- [ ] All tests pass (make test-ci)
- [ ] Code coverage >= 80%
- [ ] Linting passes (make lint)
- [ ] Security scan clean (make security)
- [ ] Code review approved
- [ ] Documentation updated
- [ ] CI/CD pipeline green

## Example Usage

```bash
# Start feature development
/add-feature "Add Azure Container Registry (ACR) support"

# This will:
# 1. Create feature spec with golang-pro
# 2. Design interfaces with architect-reviewer
# 3. Implement with golang-pro
# 4. Add tests with test-automator
# 5. Security review with security-auditor
# 6. Code review with code-reviewer
# 7. Performance review with performance-engineer
# 8. Document with api-documenter
# 9. CI integration with deployment-engineer
```

## Agent Coordination

The workflow automatically coordinates multiple agents:
- **context-manager**: Preserves context across agents
- **golang-pro**: Primary implementation
- **test-automator**: Testing
- **security-auditor**: Security review
- **code-reviewer**: Quality review
- **architect-reviewer**: Architecture alignment
- **api-documenter**: Documentation
