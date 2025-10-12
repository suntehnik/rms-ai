# Production Dockerfile for Product Requirements Management System
# Multi-stage build for optimized production image

# Build stage
FROM golang:1.24.5-alpine AS builder

# Build arguments for metadata (must be after FROM)
ARG GIT_COMMIT=""
ARG GIT_BRANCH=""
ARG BUILD_DATE=""
ARG VERSION=""
ARG TARGETARCH

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Generate Swagger documentation
RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

# Set Go architecture based on target platform
RUN case ${TARGETARCH:-amd64} in \
        "amd64") GOARCH=amd64 ;; \
        "arm64") GOARCH=arm64 ;; \
        "arm") GOARCH=arm ;; \
        *) GOARCH=amd64 ;; \
    esac && \
    echo "Building for GOARCH=$GOARCH" && \
    CGO_ENABLED=0 GOOS=linux GOARCH=$GOARCH go build \
        -ldflags="-w -s -extldflags '-static' -X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildDate=${BUILD_DATE}" \
        -a -installsuffix cgo \
        -o server cmd/server/main.go

# Build migration tool
RUN case ${TARGETARCH:-amd64} in \
        "amd64") GOARCH=amd64 ;; \
        "arm64") GOARCH=arm64 ;; \
        "arm") GOARCH=arm ;; \
        *) GOARCH=amd64 ;; \
    esac && \
    CGO_ENABLED=0 GOOS=linux GOARCH=$GOARCH go build \
        -ldflags='-w -s -extldflags "-static"' \
        -a -installsuffix cgo \
        -o migrate cmd/migrate/main.go

# Build init tool
RUN case ${TARGETARCH:-amd64} in \
        "amd64") GOARCH=amd64 ;; \
        "arm64") GOARCH=arm64 ;; \
        "arm") GOARCH=arm ;; \
        *) GOARCH=amd64 ;; \
    esac && \
    CGO_ENABLED=0 GOOS=linux GOARCH=$GOARCH go build \
        -ldflags='-w -s -extldflags "-static"' \
        -a -installsuffix cgo \
        -o init cmd/init/main.go

# Production stage
FROM alpine:3.20

# Build arguments for metadata (must be redeclared in each stage)
ARG GIT_COMMIT=""
ARG GIT_BRANCH=""
ARG BUILD_DATE=""
ARG VERSION=""

# Add metadata labels
LABEL org.opencontainers.image.title="Product Requirements Management System"
LABEL org.opencontainers.image.description="Comprehensive API for managing product requirements through hierarchical structure"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.revision="${GIT_COMMIT}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.source="https://github.com/username/product-requirements-management"
LABEL org.opencontainers.image.licenses="MIT"

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl && \
    addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set timezone
ENV TZ=UTC

# Create necessary directories
RUN mkdir -p /app/migrations /app/docs /app/logs && \
    chown -R appuser:appgroup /app

# Set working directory
WORKDIR /app

# Copy binaries from builder
COPY --from=builder --chown=appuser:appgroup /app/server ./
COPY --from=builder --chown=appuser:appgroup /app/migrate ./

# Copy migrations
COPY --from=builder --chown=appuser:appgroup /app/migrations/ ./migrations/

# Copy generated documentation
COPY --from=builder --chown=appuser:appgroup /app/docs/ ./docs/

# Copy scripts
COPY --from=builder --chown=appuser:appgroup /app/scripts/ ./scripts/

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/ready || exit 1

# Default command
CMD ["./server"]