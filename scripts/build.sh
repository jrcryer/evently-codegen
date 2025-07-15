#!/bin/bash
# AsyncAPI Go Code Generator - Build Script
# Comprehensive build script for development and release builds

set -e

# Configuration
PROJECT_NAME="evently-codegen"
MODULE_NAME="github.com/jrcryer/evently-codegen"
BUILD_DIR="bin"
DIST_DIR="dist"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Get build information
get_version() {
    git describe --tags --always --dirty 2>/dev/null || echo "dev"
}

get_commit() {
    git rev-parse --short HEAD 2>/dev/null || echo "unknown"
}

get_branch() {
    git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown"
}

get_build_time() {
    date -u '+%Y-%m-%d_%H:%M:%S'
}

# Build ldflags
build_ldflags() {
    local version="$1"
    local build_time="$2"
    local git_commit="$3"
    local git_branch="$4"
    
    echo "-ldflags \"-s -w \
        -X '${MODULE_NAME}/internal/version.Version=${version}' \
        -X '${MODULE_NAME}/internal/version.BuildTime=${build_time}' \
        -X '${MODULE_NAME}/internal/version.GitCommit=${git_commit}' \
        -X '${MODULE_NAME}/internal/version.GitBranch=${git_branch}'\""
}

# Clean build artifacts
clean() {
    log_info "Cleaning build artifacts..."
    rm -rf "$BUILD_DIR" "$DIST_DIR"
    go clean -cache -testcache
    log_success "Clean completed"
}

# Build for current platform
build_local() {
    local version build_time git_commit git_branch ldflags
    
    version=$(get_version)
    build_time=$(get_build_time)
    git_commit=$(get_commit)
    git_branch=$(get_branch)
    ldflags=$(build_ldflags "$version" "$build_time" "$git_commit" "$git_branch")
    
    log_info "Building ${PROJECT_NAME} for local platform..."
    log_info "Version: $version"
    log_info "Commit: $git_commit"
    log_info "Branch: $git_branch"
    
    mkdir -p "$BUILD_DIR"
    
    eval "CGO_ENABLED=0 go build $ldflags -o ${BUILD_DIR}/${PROJECT_NAME} ./cmd/${PROJECT_NAME}"
    
    log_success "Build completed: ${BUILD_DIR}/${PROJECT_NAME}"
}

# Build debug version
build_debug() {
    local version build_time git_commit git_branch
    
    version=$(get_version)
    build_time=$(get_build_time)
    git_commit=$(get_commit)
    git_branch=$(get_branch)
    
    log_info "Building debug version..."
    
    mkdir -p "$BUILD_DIR"
    
    CGO_ENABLED=0 go build \
        -gcflags="all=-N -l" \
        -ldflags "-X '${MODULE_NAME}/internal/version.Version=${version}' \
                  -X '${MODULE_NAME}/internal/version.BuildTime=${build_time}' \
                  -X '${MODULE_NAME}/internal/version.GitCommit=${git_commit}' \
                  -X '${MODULE_NAME}/internal/version.GitBranch=${git_branch}'" \
        -o "${BUILD_DIR}/${PROJECT_NAME}-debug" \
        "./cmd/${PROJECT_NAME}"
    
    log_success "Debug build completed: ${BUILD_DIR}/${PROJECT_NAME}-debug"
}

# Cross-compile for all platforms
build_all() {
    local version build_time git_commit git_branch
    local platforms=(
        "linux/amd64"
        "linux/arm64"
        "darwin/amd64"
        "darwin/arm64"
        "windows/amd64"
        "windows/arm64"
    )
    
    version=$(get_version)
    build_time=$(get_build_time)
    git_commit=$(get_commit)
    git_branch=$(get_branch)
    
    log_info "Cross-compiling for all platforms..."
    log_info "Version: $version"
    
    mkdir -p "$DIST_DIR"
    
    for platform in "${platforms[@]}"; do
        IFS='/' read -r os arch <<< "$platform"
        
        local output_name="${PROJECT_NAME}-${os}-${arch}"
        if [ "$os" = "windows" ]; then
            output_name="${output_name}.exe"
        fi
        
        log_info "Building for ${os}/${arch}..."
        
        GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build \
            -ldflags "-s -w \
                      -X '${MODULE_NAME}/internal/version.Version=${version}' \
                      -X '${MODULE_NAME}/internal/version.BuildTime=${build_time}' \
                      -X '${MODULE_NAME}/internal/version.GitCommit=${git_commit}' \
                      -X '${MODULE_NAME}/internal/version.GitBranch=${git_branch}'" \
            -o "${DIST_DIR}/${output_name}" \
            "./cmd/${PROJECT_NAME}"
    done
    
    log_success "Cross-compilation completed"
}

# Create release packages
package() {
    log_info "Creating release packages..."
    
    if [ ! -d "$DIST_DIR" ]; then
        log_error "Distribution directory not found. Run build-all first."
        exit 1
    fi
    
    cd "$DIST_DIR"
    
    for binary in ${PROJECT_NAME}-*; do
        if [[ "$binary" == *".exe" ]]; then
            # Windows binary - create zip
            zip "${binary}.zip" "$binary"
            log_info "Created ${binary}.zip"
        else
            # Unix binary - create tar.gz
            tar -czf "${binary}.tar.gz" "$binary"
            log_info "Created ${binary}.tar.gz"
        fi
    done
    
    cd - >/dev/null
    log_success "Release packages created"
}

# Generate checksums
checksums() {
    log_info "Generating checksums..."
    
    if [ ! -d "$DIST_DIR" ]; then
        log_error "Distribution directory not found. Run package first."
        exit 1
    fi
    
    cd "$DIST_DIR"
    sha256sum *.tar.gz *.zip > checksums.txt 2>/dev/null || true
    cd - >/dev/null
    
    log_success "Checksums generated: ${DIST_DIR}/checksums.txt"
}

# Run tests
test() {
    log_info "Running tests..."
    go test -v -race ./...
    log_success "Tests completed"
}

# Run tests with coverage
test_coverage() {
    log_info "Running tests with coverage..."
    go test -v -race -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    go tool cover -func=coverage.out | tail -1
    log_success "Coverage report generated: coverage.html"
}

# Lint code
lint() {
    log_info "Running linter..."
    if command -v golangci-lint >/dev/null 2>&1; then
        golangci-lint run ./...
        log_success "Linting completed"
    else
        log_warning "golangci-lint not found. Install it with:"
        log_warning "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    fi
}

# Security check
security() {
    log_info "Running security checks..."
    
    # Run gosec if available
    if command -v gosec >/dev/null 2>&1; then
        gosec ./...
    else
        log_warning "gosec not found. Install it with:"
        log_warning "go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"
    fi
    
    # Run govulncheck if available
    if command -v govulncheck >/dev/null 2>&1; then
        govulncheck ./...
    else
        log_warning "govulncheck not found. Install it with:"
        log_warning "go install golang.org/x/vuln/cmd/govulncheck@latest"
    fi
    
    log_success "Security checks completed"
}

# Full release build
release() {
    log_info "Starting full release build..."
    
    clean
    test
    lint
    security
    build_all
    package
    checksums
    
    log_success "Release build completed!"
    log_info "Release artifacts:"
    ls -la "$DIST_DIR"
}

# Development build
dev() {
    log_info "Starting development build..."
    
    go mod tidy
    go fmt ./...
    go vet ./...
    test
    build_local
    
    log_success "Development build completed!"
}

# Show usage
usage() {
    cat << EOF
AsyncAPI Go Code Generator - Build Script

Usage: $0 [COMMAND]

Commands:
    clean           Clean build artifacts
    build           Build for current platform
    build-debug     Build debug version
    build-all       Cross-compile for all platforms
    package         Create release packages
    checksums       Generate checksums
    test            Run tests
    test-coverage   Run tests with coverage
    lint            Run linter
    security        Run security checks
    release         Full release build (clean, test, lint, security, build-all, package, checksums)
    dev             Development build (tidy, fmt, vet, test, build)
    help            Show this help message

Examples:
    $0 build        # Build for current platform
    $0 release      # Create full release
    $0 dev          # Quick development build
EOF
}

# Main function
main() {
    case "${1:-build}" in
        clean)
            clean
            ;;
        build)
            build_local
            ;;
        build-debug)
            build_debug
            ;;
        build-all)
            build_all
            ;;
        package)
            package
            ;;
        checksums)
            checksums
            ;;
        test)
            test
            ;;
        test-coverage)
            test_coverage
            ;;
        lint)
            lint
            ;;
        security)
            security
            ;;
        release)
            release
            ;;
        dev)
            dev
            ;;
        help|--help|-h)
            usage
            ;;
        *)
            log_error "Unknown command: $1"
            usage
            exit 1
            ;;
    esac
}

# Run main function
main "$@"