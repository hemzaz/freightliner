# GitHub Actions & CI/CD Skill

Expert skill for creating production-ready GitHub Actions workflows and CI/CD pipelines.

## Core Concepts

### Workflow Triggers
```yaml
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]
  workflow_dispatch:  # Manual trigger
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM UTC
```

### Job Strategy
```yaml
jobs:
  test:
    strategy:
      fail-fast: false  # Don't stop other jobs on failure
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go-version: ['1.23', '1.24']
```

### Caching
```yaml
- uses: actions/cache@v4
  with:
    path: |
      ~/go/pkg/mod
      ~/.cache/go-build
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-
```

## Go Project CI/CD Patterns

### Build and Test
```yaml
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.24'
    cache: true

- name: Download dependencies
  run: |
    go mod download
    go mod verify

- name: Build
  run: go build -v ./...

- name: Test
  run: go test -v -race -coverprofile=coverage.out ./...

- name: Upload coverage
  uses: codecov/codecov-action@v4
  with:
    file: ./coverage.out
```

### Linting
```yaml
- name: golangci-lint
  uses: golangci/golangci-lint-action@v4
  with:
    version: v1.62.2
    args: --timeout=10m
```

### Security Scanning
```yaml
- name: Run gosec
  run: |
    go install github.com/securego/gosec/v2/cmd/gosec@latest
    gosec -fmt sarif -out gosec.sarif ./...

- name: Upload SARIF
  uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: gosec.sarif
```

### Docker Build
```yaml
- name: Set up Docker Buildx
  uses: docker/setup-buildx-action@v3

- name: Build Docker image
  uses: docker/build-push-action@v5
  with:
    context: .
    push: false
    tags: myapp:${{ github.sha }}
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

## Best Practices

### 1. Use Concurrency Control
```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
```

### 2. Set Timeouts
```yaml
jobs:
  test:
    timeout-minutes: 30
    steps:
      - name: Long running step
        timeout-minutes: 10
        run: ...
```

### 3. Conditional Execution
```yaml
- name: Deploy
  if: github.ref == 'refs/heads/main' && github.event_name == 'push'
  run: ./deploy.sh
```

### 4. Secrets Management
```yaml
env:
  API_KEY: ${{ secrets.API_KEY }}

- name: Use secret
  run: |
    echo "::add-mask::$API_KEY"  # Mask in logs
    ./script.sh
```

### 5. Artifact Management
```yaml
- name: Upload artifacts
  uses: actions/upload-artifact@v4
  with:
    name: test-results
    path: |
      *.log
      coverage.out
    retention-days: 30
```

## Complete CI Workflow Template

```yaml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  GO_VERSION: '1.24'

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Test
        run: make test-ci

  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - uses: golangci/golangci-lint-action@v4
        with:
          version: latest

  security:
    name: Security
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true

      - name: Run gosec
        run: make security
```

## Advanced Patterns

### Matrix Testing
```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest]
    go: ['1.23', '1.24']
    include:
      - os: ubuntu-latest
        go: '1.24'
        is-latest: true
```

### Reusable Workflows
```yaml
# .github/workflows/reusable-test.yml
on:
  workflow_call:
    inputs:
      go-version:
        required: true
        type: string

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ inputs.go-version }}
      - run: go test ./...
```

### Composite Actions
```yaml
# .github/actions/setup-go/action.yml
name: 'Setup Go Environment'
description: 'Sets up Go with caching'
inputs:
  go-version:
    description: 'Go version'
    required: true
runs:
  using: composite
  steps:
    - uses: actions/setup-go@v5
      with:
        go-version: ${{ inputs.go-version }}
        cache: true
    - run: go mod download
      shell: bash
```

## Troubleshooting

### Common Issues

#### 1. Cache Not Working
```yaml
# Ensure cache key includes go.sum hash
key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
```

#### 2. Permission Denied
```yaml
# Add permissions
permissions:
  contents: read
  packages: write
```

#### 3. Timeout Issues
```yaml
# Increase timeout and add retry
timeout-minutes: 60
- uses: nick-fields/retry-action@v2
  with:
    timeout_minutes: 10
    max_attempts: 3
    command: go test ./...
```

## Performance Optimization

### 1. Parallel Jobs
```yaml
jobs:
  test:
    # Runs in parallel with lint
  lint:
    # Runs in parallel with test
  deploy:
    needs: [test, lint]  # Waits for both
```

### 2. Conditional Jobs
```yaml
if: |
  github.event_name == 'push' ||
  contains(github.event.pull_request.labels.*.name, 'run-ci')
```

### 3. Sparse Checkout
```yaml
- uses: actions/checkout@v4
  with:
    sparse-checkout: |
      src/
      go.mod
      go.sum
```

## Security Best Practices

1. **Pin Actions to SHA**: `uses: actions/checkout@8ade135a41bc03ea155e62e844d188df1ea18608`
2. **Minimal Permissions**: Grant only necessary permissions
3. **Secrets Scanning**: Enable GitHub secret scanning
4. **SARIF Upload**: Upload security scan results
5. **Dependency Review**: Use `actions/dependency-review-action@v3`

## Freightliner-Specific Patterns

### Skip Cloud Integration Tests
```yaml
- name: Run tests (skip cloud)
  run: go test -v -short ./...
  env:
    SKIP_INTEGRATION: "true"
```

### Build Without Cloud SDKs
```yaml
- name: Build (minimal)
  run: go build -tags=nocloud -o freightliner .
```

### Docker Build Optimization
```yaml
- uses: docker/build-push-action@v5
  with:
    context: .
    target: build  # Stop at build stage
    cache-from: type=gha
    cache-to: type=gha,mode=max
```
