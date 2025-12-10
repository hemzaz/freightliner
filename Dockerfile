# syntax=docker/dockerfile:1.7
# Optimized multi-stage Dockerfile for CI/CD and production

# Build stage
FROM golang:1.24.5-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# Set working directory
WORKDIR /app

# Configure Go environment  
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

# Copy go module files for better caching
COPY go.mod go.sum ./

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

# Copy source code
COPY . .

# Test stage - can be targeted for CI testing
FROM builder AS test
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go test -v -race -coverprofile=coverage.out ./...

# Build binary with optimizations
FROM builder AS build
ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-w -s" -trimpath -o freightliner .

# Production stage - minimal runtime
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -s /bin/sh -u 1001 appuser

# Set working directory and user
WORKDIR /app
USER 1001:1001

# Copy binary
COPY --from=build --chown=1001:1001 /app/freightliner .

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ./freightliner --version || exit 1

# Default command
ENTRYPOINT ["./freightliner"]
CMD ["--help"]

# Metadata
LABEL org.opencontainers.image.title="freightliner" \
      org.opencontainers.image.description="Container registry replication tool" \
      org.opencontainers.image.source="https://github.com/company/freightliner"