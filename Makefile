.PHONY: build test test-manifest test-ci test-local test-integration test-unit lint clean fmt imports vet test-setup test-validate test-cleanup test-full

# Build the application
build:
	GO111MODULE=on GOFLAGS=-mod=mod go build -o bin/freightliner main.go

# Build test manifest tool
build-test-manifest:
	@mkdir -p bin
	GO111MODULE=on GOFLAGS=-mod=mod go build -o bin/test-manifest ./cmd/test-manifest

# Run all tests (legacy - now uses manifest)
test: test-manifest

# Run tests with manifest-based filtering (auto-detect environment)
test-manifest: build-test-manifest
	@echo "Running tests with manifest-based filtering..."
	./scripts/test-with-manifest.sh

# Run tests optimized for CI environment
test-ci: build-test-manifest
	@echo "Running CI-optimized tests..."
	./scripts/test-with-manifest.sh --env ci

# Run tests for local development environment
test-local: build-test-manifest
	@echo "Running local development tests..."
	./scripts/test-with-manifest.sh --env local

# Run full integration tests
test-integration: build-test-manifest
	@echo "Running integration tests..."
	./scripts/test-with-manifest.sh --env integration

# Run only unit tests
test-unit: build-test-manifest
	@echo "Running unit tests only..."
	./scripts/test-with-manifest.sh --categories unit

# Run tests without external dependencies
test-no-deps: build-test-manifest
	@echo "Running tests without external dependencies..."
	./scripts/test-with-manifest.sh --categories unit

# Show test manifest summary
test-summary: build-test-manifest
	@echo "Test Manifest Summary:"
	./scripts/test-with-manifest.sh --summary

# Validate test manifest
test-manifest-validate: build-test-manifest
	@echo "Validating test manifest..."
	./bin/test-manifest validate

# Run legacy test command (without manifest filtering)
test-legacy:
	GO111MODULE=on GOFLAGS=-mod=mod go test -v ./...

# Run linting
lint:
	./scripts/lint.sh ./...

# Run fast linting (critical linters only)
lint-fast:
	./scripts/lint.sh --fast ./...

# Clean build artifacts
clean:
	rm -rf bin/
	GO111MODULE=on go clean

# Format code with gofmt
fmt:
	GO111MODULE=on go fmt ./...

# Check if code is formatted
fmt-check:
	@if [ "$$(gofmt -l . | wc -l)" -gt 0 ]; then \
		echo "Code is not formatted with gofmt. Run 'make fmt' to fix."; \
		gofmt -d .; \
		exit 1; \
	fi

# Organize imports
imports:
	./scripts/organize_imports.sh

# Check if imports are organized
imports-check:
	@go install golang.org/x/tools/cmd/goimports@v0.29.0
	@GOIMPORTS_OUTPUT=$$(goimports -l -local freightliner .); \
	if [ -n "$$GOIMPORTS_OUTPUT" ]; then \
		echo "Imports not properly organized. Run 'make imports' to fix."; \
		echo "$$GOIMPORTS_OUTPUT"; \
		exit 1; \
	fi

# Run go vet
vet:
	./scripts/vet.sh ./...

# Tool versions
GOIMPORTS_VERSION = v0.29.0
GOLANGCI_LINT_VERSION = v2.3.0
SHADOW_VERSION = v0.29.0
INTERFACER_VERSION = v0.0.0-20180902061238-70be1b28218b
STATICCHECK_VERSION = 2025.1.1

# Setup development environment
setup:
	GO111MODULE=on go mod download
	go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@$(SHADOW_VERSION)
	go install github.com/mvdan/interfacer/cmd/interfacer@$(INTERFACER_VERSION)
	# go install honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION)  # Now handled by golangci-lint
	cp scripts/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
	@echo "Setup complete. Running initial checks..."
	./scripts/lint.sh --fast
	./scripts/vet.sh
	# ./scripts/staticcheck.sh  # Now handled by golangci-lint

# Run staticcheck - DISABLED: now handled by golangci-lint
# staticcheck:
#	./scripts/staticcheck.sh ./...

# Run all quality checks
check: fmt imports vet lint test

# Install hooks
hooks:
	cp .git/hooks/pre-commit .git/hooks/pre-commit.backup 2>/dev/null || true
	cp scripts/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit

# Test registry setup and management
test-setup:
	@echo "Setting up test registries for Freightliner..."
	./scripts/setup-test-registries.sh

test-validate:
	@echo "Validating test registry setup..."
	./scripts/test-registry-setup.sh

test-cleanup:
	@echo "Cleaning up test registries..."
	./scripts/setup-test-registries.sh --cleanup

# Run complete test cycle with registries
test-full: test-setup
	@echo "Running tests with local registries..."
	sleep 5  # Give registries time to fully initialize
	go test ./pkg/tree/ -v -timeout=300s
	go test ./pkg/service/ -v -timeout=300s  
	go test ./pkg/copy/ -v -timeout=300s
	@echo "Tests complete. Use 'make test-cleanup' to remove registries."

# Quick test run (assumes registries are already running)
test-quick:
	@echo "Running quick tests (registries must be running)..."
	go test ./pkg/tree/ -v -timeout=300s
	go test ./pkg/service/ -v -timeout=300s
	go test ./pkg/copy/ -v -timeout=300s
