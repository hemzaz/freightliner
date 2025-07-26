.PHONY: build test lint clean fmt imports vet test-setup test-validate test-cleanup test-full

# Build the application
build:
	go build -o bin/freightliner main.go

# Run all tests
test:
	go test -v ./...

# Run linting
lint:
	./scripts/lint.sh ./...

# Run fast linting (critical linters only)
lint-fast:
	./scripts/lint.sh --fast ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Format code with gofmt
fmt:
	go fmt ./...

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
	@go install golang.org/x/tools/cmd/goimports@latest
	@GOIMPORTS_OUTPUT=$$(goimports -l -local freightliner .); \
	if [ -n "$$GOIMPORTS_OUTPUT" ]; then \
		echo "Imports not properly organized. Run 'make imports' to fix."; \
		echo "$$GOIMPORTS_OUTPUT"; \
		exit 1; \
	fi

# Run go vet
vet:
	./scripts/vet.sh ./...

# Setup development environment
setup:
	go mod download
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest
	go install github.com/mvdan/interfacer/cmd/interfacer@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	cp scripts/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
	@echo "Setup complete. Running initial checks..."
	./scripts/lint.sh --fast
	./scripts/vet.sh
	./scripts/staticcheck.sh

# Run staticcheck
staticcheck:
	./scripts/staticcheck.sh ./...

# Run all quality checks
check: fmt imports vet lint staticcheck test

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
