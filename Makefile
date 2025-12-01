# Freightliner Build System
# Optimized for CI/CD and local development

# Build configuration
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GO_VERSION ?= 1.23.4

# Build flags
LDFLAGS := -w -s \
	-X freightliner/cmd.version=$(VERSION) \
	-X freightliner/cmd.buildTime=$(BUILD_TIME) \
	-X freightliner/cmd.gitCommit=$(GIT_COMMIT)
BUILD_FLAGS := -a -installsuffix cgo -ldflags "$(LDFLAGS)"

# Test configuration  
TEST_TIMEOUT ?= 8m
PARALLEL_JOBS ?= 4
TEST_FLAGS := -v -timeout=$(TEST_TIMEOUT) -parallel=$(PARALLEL_JOBS)
COVERAGE_FILE := coverage.out

# Performance optimization
GOMAXPROCS ?= $(PARALLEL_JOBS)
BUILD_CACHE ?= true

# Docker configuration
DOCKER_IMAGE := freightliner
DOCKER_TAG ?= $(VERSION)
DOCKERFILE ?= Dockerfile.optimized

# Tools
GOLANGCI_LINT_VERSION := v1.62.2
GOSEC_VERSION := latest

.PHONY: help
help: ## Show this help message
	@echo "Freightliner Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development targets
.PHONY: dev
dev: clean deps build test ## Full development cycle

.PHONY: clean
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/ dist/ coverage.out coverage.html
	@go clean -cache -testcache -modcache

.PHONY: deps
deps: ## Download and verify dependencies
	@echo "📦 Downloading dependencies..."
	@go mod download
	@go mod verify
	@go mod tidy

# Build targets
.PHONY: build
build: ## Build the application
	@echo "🔨 Building freightliner..."
	@mkdir -p bin
	@CGO_ENABLED=0 go build $(BUILD_FLAGS) -o bin/freightliner .

.PHONY: banner
banner: ## Display ASCII banner
	@go run -ldflags "$(LDFLAGS)" . version --banner

.PHONY: build-race
build-race: ## Build with race detection
	@echo "🔨 Building freightliner with race detection..."
	@mkdir -p bin
	@CGO_ENABLED=1 go build -race $(BUILD_FLAGS) -o bin/freightliner-race .

.PHONY: build-static
build-static: ## Build static binary
	@echo "🔨 Building static freightliner binary..."
	@mkdir -p bin
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		$(BUILD_FLAGS) \
		-ldflags "$(LDFLAGS) -extldflags '-static'" \
		-tags 'netgo osusergo static_build' \
		-o bin/freightliner-static .

# Test targets
.PHONY: test
test: ## Run all tests
	@echo "🧪 Running tests..."
	@go test $(TEST_FLAGS) ./...

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "🧪 Running unit tests..."
	@go test $(TEST_FLAGS) -short ./...

.PHONY: test-integration
test-integration: ## Run integration tests only
	@echo "🧪 Running integration tests..."
	@go test $(TEST_FLAGS) -run Integration ./...

.PHONY: test-race
test-race: ## Run tests with race detection
	@echo "🧪 Running tests with race detection..."
	@CGO_ENABLED=1 go test $(TEST_FLAGS) -race ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "📊 Running tests with coverage..."
	@go test $(TEST_FLAGS) -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-ci
test-ci: ## Run CI-optimized tests
	@echo "🚀 Running CI tests..."
	@go test $(TEST_FLAGS) -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...

# Quality assurance targets
.PHONY: lint
lint: ## Run linter
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=8m; \
	else \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
		golangci-lint run --timeout=8m; \
	fi

.PHONY: vet
vet: ## Run go vet
	@echo "🔍 Running go vet..."
	@go vet ./...

.PHONY: fmt
fmt: ## Format code
	@echo "✨ Formatting code..."
	@gofmt -w .

.PHONY: fmt-check
fmt-check: ## Check code formatting
	@echo "🔍 Checking code format..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "❌ Code is not formatted. Run 'make fmt' to fix."; \
		gofmt -l .; \
		exit 1; \
	else \
		echo "✅ Code is properly formatted"; \
	fi

.PHONY: security
security: ## Run security scan
	@echo "🔒 Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION); \
		gosec ./...; \
	fi

.PHONY: quality
quality: fmt-check vet lint security ## Run all quality checks

# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@docker build \
		-f $(DOCKERFILE) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):latest \
		.

.PHONY: docker-test
docker-test: ## Test Docker image
	@echo "🧪 Testing Docker image..."
	@docker build \
		-f $(DOCKERFILE) \
		--target test \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(DOCKER_IMAGE):test \
		.

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "🚀 Running Docker container..."
	@docker run --rm -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-shell
docker-shell: ## Open shell in Docker container
	@echo "🐚 Opening shell in Docker container..."
	@docker run --rm -it --entrypoint /bin/sh $(DOCKER_IMAGE):$(DOCKER_TAG)

# Performance targets
.PHONY: bench
bench: ## Run benchmarks
	@echo "⚡ Running benchmarks..."
	@go test -bench=. -benchmem ./...

.PHONY: profile
profile: ## Generate CPU profile
	@echo "📈 Generating CPU profile..."
	@go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...
	@echo "Profiles generated: cpu.prof, mem.prof"

# Release targets
.PHONY: release-build
release-build: clean ## Build release binaries for multiple platforms
	@echo "🚀 Building release binaries..."
	@mkdir -p dist
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			if [ "$$os" = "windows" ]; then ext=".exe"; else ext=""; fi; \
			echo "Building $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build \
				$(BUILD_FLAGS) \
				-ldflags "$(LDFLAGS) -extldflags '-static'" \
				-tags 'netgo osusergo static_build' \
				-o dist/freightliner-$$os-$$arch$$ext .; \
		done; \
	done

.PHONY: install
install: build ## Install binary to GOPATH/bin
	@echo "📦 Installing freightliner..."
	@go install $(BUILD_FLAGS) .

# Utility targets
.PHONY: version
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Go Version: $(GO_VERSION)"

.PHONY: env
env: ## Show build environment
	@echo "Build Environment:"
	@echo "  GO_VERSION: $(GO_VERSION)"
	@echo "  GOOS: $$(go env GOOS)"
	@echo "  GOARCH: $$(go env GOARCH)"
	@echo "  GOROOT: $$(go env GOROOT)"
	@echo "  GOPATH: $$(go env GOPATH)"
	@echo "  CGO_ENABLED: $$(go env CGO_ENABLED)"

.PHONY: tools
tools: ## Install development tools
	@echo "🔧 Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)
	@go install honnef.co/go/tools/cmd/staticcheck@latest

# Default target
.DEFAULT_GOAL := help