.PHONY: build test lint clean fmt imports vet

# Build the application
build:
	go build -o bin/freightliner main.go

# Run all tests
test:
	go test -v ./...

# Run linting
lint:
	golangci-lint run ./...

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
	go vet ./...

# Setup development environment
setup:
	go mod download
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run all quality checks
check: fmt imports vet lint test

# Install hooks
hooks:
	cp .git/hooks/pre-commit .git/hooks/pre-commit.backup 2>/dev/null || true
	cp scripts/pre-commit .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
