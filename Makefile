.PHONY: build test lint clean fmt imports vet

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

# Organize imports
imports:
	./scripts/organize_imports.sh

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
