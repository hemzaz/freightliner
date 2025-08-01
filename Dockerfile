# Production-ready multi-stage Dockerfile for Freightliner
# Optimized for security, performance, and minimal attack surface

# Build stage
FROM golang:1.21-alpine AS builder

# Install security updates and build dependencies
RUN apk update && apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

# Create non-root user for build
RUN adduser -D -s /bin/sh -u 1001 appuser

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with security flags
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}" \
    -o freightliner \
    .

# Security scan stage (optional - can be used in CI)
FROM aquasec/trivy:latest AS security-scan
COPY --from=builder /build/freightliner /tmp/freightliner
RUN trivy filesystem --exit-code 0 --no-progress --format table /tmp/

# Final production stage
FROM scratch

# Import CA certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Import timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Import user/group files for non-root user
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy the binary
COPY --from=builder /build/freightliner /app/freightliner

# Use non-root user
USER 1001:1001

# Set working directory
WORKDIR /app

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/app/freightliner", "health-check"]

# Set entrypoint
ENTRYPOINT ["/app/freightliner"]
CMD ["serve"]

# Labels for metadata
LABEL maintainer="Platform Team <platform@company.com>"
LABEL description="Freightliner Container Registry Replication"
LABEL version="${VERSION}"
LABEL org.opencontainers.image.title="freightliner"
LABEL org.opencontainers.image.description="Container registry replication between AWS ECR and GCP Artifact Registry"
LABEL org.opencontainers.image.vendor="Company Platform Team"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.source="https://github.com/company/freightliner"