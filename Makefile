# Makefile for Freightliner Container Registry Replication

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD)
LDFLAGS = -w -s -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)

# Docker variables
DOCKER_REGISTRY ?= ghcr.io
DOCKER_REPOSITORY ?= company/freightliner
DOCKER_TAG ?= $(VERSION)
DOCKER_IMAGE = $(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY):$(DOCKER_TAG)

# Go variables
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
CGO_ENABLED ?= 0

# Directories
BIN_DIR = bin
BUILD_DIR = build
COVERAGE_DIR = coverage

.PHONY: help
help: ## Display this help message
	@echo "Freightliner Container Registry Replication"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR) $(BUILD_DIR) $(COVERAGE_DIR)
	docker system prune -f

.PHONY: deps
deps: ## Download and verify dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod verify
	go mod tidy

.PHONY: generate
generate: ## Generate code
	@echo "Generating code..."
	go generate ./...

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

.PHONY: lint
lint: ## Run linters
	@echo "Running linters..."
	golangci-lint run ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	mkdir -p $(COVERAGE_DIR)
	go test -v -race -covermode=atomic -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html

.PHONY: test-short
test-short: ## Run short tests
	@echo "Running short tests..."
	go test -short ./...

.PHONY: bench
bench: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

.PHONY: security
security: ## Run security scanners
	@echo "Running security scanners..."
	gosec ./...
	govulncheck ./...

.PHONY: build
build: ## Build the application
	@echo "Building freightliner $(VERSION)..."
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
		-ldflags="$(LDFLAGS)" \
		-o $(BIN_DIR)/freightliner \
		.

.PHONY: build-all
build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	mkdir -p $(BIN_DIR)
	
	# Linux AMD64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags="$(LDFLAGS)" \
		-o $(BIN_DIR)/freightliner-linux-amd64 \
		.
	
	# Linux ARM64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
		-ldflags="$(LDFLAGS)" \
		-o $(BIN_DIR)/freightliner-linux-arm64 \
		.
	
	# macOS AMD64
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
		-ldflags="$(LDFLAGS)" \
		-o $(BIN_DIR)/freightliner-darwin-amd64 \
		.
	
	# macOS ARM64
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
		-ldflags="$(LDFLAGS)" \
		-o $(BIN_DIR)/freightliner-darwin-arm64 \
		.
	
	# Windows AMD64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
		-ldflags="$(LDFLAGS)" \
		-o $(BIN_DIR)/freightliner-windows-amd64.exe \
		.

.PHONY: install
install: build ## Install the application
	@echo "Installing freightliner..."
	go install -ldflags="$(LDFLAGS)" .

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image $(DOCKER_IMAGE)..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t $(DOCKER_IMAGE) \
		-t $(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY):latest \
		.

.PHONY: docker-build-dev
docker-build-dev: ## Build development Docker image
	@echo "Building development Docker image..."
	docker build -f Dockerfile.dev -t freightliner:dev .

.PHONY: docker-push
docker-push: docker-build ## Push Docker image
	@echo "Pushing Docker image $(DOCKER_IMAGE)..."
	docker push $(DOCKER_IMAGE)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY):latest

.PHONY: docker-scan
docker-scan: docker-build ## Scan Docker image for vulnerabilities
	@echo "Scanning Docker image for vulnerabilities..."
	trivy image $(DOCKER_IMAGE)

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run --rm -p 8080:8080 $(DOCKER_IMAGE)

.PHONY: compose-up
compose-up: ## Start development environment with Docker Compose
	@echo "Starting development environment..."
	docker-compose up -d

.PHONY: compose-down
compose-down: ## Stop development environment
	@echo "Stopping development environment..."
	docker-compose down

.PHONY: compose-logs
compose-logs: ## View Docker Compose logs
	@echo "Viewing logs..."
	docker-compose logs -f

.PHONY: compose-prod-up
compose-prod-up: ## Start production-like environment
	@echo "Starting production-like environment..."
	docker-compose -f docker-compose.prod.yml up -d

.PHONY: compose-prod-down
compose-prod-down: ## Stop production-like environment
	@echo "Stopping production-like environment..."
	docker-compose -f docker-compose.prod.yml down

.PHONY: k8s-deploy-dev
k8s-deploy-dev: ## Deploy to development Kubernetes cluster
	@echo "Deploying to development cluster..."
	helm upgrade --install freightliner-dev ./deployments/helm/freightliner \
		--namespace freightliner-dev \
		--create-namespace \
		--set image.tag=$(VERSION) \
		--set replicaCount=1 \
		--set resources.requests.cpu=100m \
		--set resources.requests.memory=256Mi \
		--set config.logLevel=debug

.PHONY: k8s-deploy-staging
k8s-deploy-staging: ## Deploy to staging Kubernetes cluster
	@echo "Deploying to staging cluster..."
	helm upgrade --install freightliner-staging ./deployments/helm/freightliner \
		--namespace freightliner-staging \
		--create-namespace \
		--values ./deployments/helm/freightliner/values-staging.yaml \
		--set image.tag=$(VERSION)

.PHONY: k8s-deploy-prod
k8s-deploy-prod: ## Deploy to production Kubernetes cluster
	@echo "Deploying to production cluster..."
	helm upgrade --install freightliner-prod ./deployments/helm/freightliner \
		--namespace freightliner \
		--create-namespace \
		--values ./deployments/helm/freightliner/values-production.yaml \
		--set image.tag=$(VERSION)

.PHONY: terraform-init-aws
terraform-init-aws: ## Initialize Terraform for AWS
	@echo "Initializing Terraform for AWS..."
	cd deployments/terraform/aws && terraform init

.PHONY: terraform-plan-aws
terraform-plan-aws: ## Plan Terraform changes for AWS
	@echo "Planning Terraform changes for AWS..."
	cd deployments/terraform/aws && terraform plan

.PHONY: terraform-apply-aws
terraform-apply-aws: ## Apply Terraform changes for AWS
	@echo "Applying Terraform changes for AWS..."
	cd deployments/terraform/aws && terraform apply

.PHONY: terraform-init-gcp
terraform-init-gcp: ## Initialize Terraform for GCP
	@echo "Initializing Terraform for GCP..."
	cd deployments/terraform/gcp && terraform init

.PHONY: terraform-plan-gcp
terraform-plan-gcp: ## Plan Terraform changes for GCP
	@echo "Planning Terraform changes for GCP..."
	cd deployments/terraform/gcp && terraform plan

.PHONY: terraform-apply-gcp
terraform-apply-gcp: ## Apply Terraform changes for GCP
	@echo "Applying Terraform changes for GCP..."
	cd deployments/terraform/gcp && terraform apply

.PHONY: tools
tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/cosmtrek/air@latest
	go install github.com/go-delve/delve/cmd/dlv@latest

.PHONY: validate
validate: deps fmt vet lint test security ## Run all validation checks

.PHONY: release
release: validate build-all docker-build docker-push ## Create a release

.PHONY: dev
dev: ## Run development server with hot reload
	@echo "Starting development server with hot reload..."
	air

.PHONY: debug
debug: ## Run with debugger
	@echo "Starting with debugger..."
	dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient

.PHONY: check-env
check-env: ## Check environment variables
	@echo "Checking environment variables..."
	@echo "GO version: $(shell go version)"
	@echo "Docker version: $(shell docker --version)"
	@echo "Kubectl version: $(shell kubectl version --client --short 2>/dev/null || echo 'kubectl not found')"
	@echo "Helm version: $(shell helm version --short 2>/dev/null || echo 'helm not found')"
	@echo "Terraform version: $(shell terraform version 2>/dev/null || echo 'terraform not found')"

.PHONY: docs
docs: ## Generate documentation
	@echo "Generating documentation..."
	go doc -all . > docs/api.md

.PHONY: benchmark-report
benchmark-report: ## Generate benchmark report
	@echo "Generating benchmark report..."
	mkdir -p $(COVERAGE_DIR)
	go test -bench=. -benchmem -cpuprofile $(COVERAGE_DIR)/cpu.prof -memprofile $(COVERAGE_DIR)/mem.prof ./...
	go tool pprof -http=:8081 $(COVERAGE_DIR)/cpu.prof &
	@echo "CPU profile server started at http://localhost:8081"

.PHONY: all
all: validate build docker-build ## Run all checks and build everything