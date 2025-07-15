# AsyncAPI Go Code Generator - Multi-stage Docker build
# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the binary with version information
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT
ARG GIT_BRANCH

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w \
    -X 'github.com/jrcryer/evently-codegen/internal/version.Version=${VERSION}' \
    -X 'github.com/jrcryer/evently-codegen/internal/version.BuildTime=${BUILD_TIME}' \
    -X 'github.com/jrcryer/evently-codegen/internal/version.GitCommit=${GIT_COMMIT}' \
    -X 'github.com/jrcryer/evently-codegen/internal/version.GitBranch=${GIT_BRANCH}'" \
    -o evently-codegen ./cmd/evently-codegen

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S asyncapi && \
    adduser -u 1001 -S asyncapi -G asyncapi

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/evently-codegen /usr/local/bin/evently-codegen

# Copy example files for testing
COPY --from=builder /app/testdata /app/testdata

# Change ownership
RUN chown -R asyncapi:asyncapi /app

# Switch to non-root user
USER asyncapi

# Set entrypoint
ENTRYPOINT ["evently-codegen"]

# Default command shows help
CMD ["--help"]

# Metadata
LABEL org.opencontainers.image.title="AsyncAPI Go Code Generator"
LABEL org.opencontainers.image.description="Generate Go type definitions from AsyncAPI specifications"
LABEL org.opencontainers.image.source="https://github.com/jrcryer/evently-codegen"
LABEL org.opencontainers.image.licenses="MIT"