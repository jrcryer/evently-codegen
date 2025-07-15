#!/bin/bash
# AsyncAPI Go Code Generator - Installation Script
# This script downloads and installs the latest release of evently-codegen

set -e

# Configuration
REPO="asyncapi-go-codegen"
BINARY_NAME="evently-codegen"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
GITHUB_REPO="${GITHUB_REPO:-jrcryer/evently-codegen}"

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

# Detect OS and architecture
detect_platform() {
    local os arch
    
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)          log_error "Unsupported operating system: $(uname -s)"; exit 1 ;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        arm64|aarch64)  arch="arm64" ;;
        armv7l)         arch="arm" ;;
        *)              log_error "Unsupported architecture: $(uname -m)"; exit 1 ;;
    esac
    
    echo "${os}/${arch}"
}

# Get latest release version from GitHub
get_latest_version() {
    local version
    version=$(curl -s "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$version" ]; then
        log_error "Failed to get latest version from GitHub"
        exit 1
    fi
    
    echo "$version"
}

# Download and install binary
install_binary() {
    local version="$1"
    local platform="$2"
    local os arch extension
    
    IFS='/' read -r os arch <<< "$platform"
    
    # Determine file extension
    if [ "$os" = "windows" ]; then
        extension=".exe"
    else
        extension=""
    fi
    
    local binary_name="${BINARY_NAME}-${os}-${arch}${extension}"
    local archive_name
    
    # Determine archive format
    if [ "$os" = "windows" ]; then
        archive_name="${binary_name}.zip"
    else
        archive_name="${binary_name}.tar.gz"
    fi
    
    local download_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${archive_name}"
    local temp_dir
    temp_dir=$(mktemp -d)
    
    log_info "Downloading ${BINARY_NAME} ${version} for ${platform}..."
    log_info "Download URL: ${download_url}"
    
    # Download archive
    if ! curl -L -o "${temp_dir}/${archive_name}" "$download_url"; then
        log_error "Failed to download ${archive_name}"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    # Extract archive
    log_info "Extracting archive..."
    cd "$temp_dir"
    
    if [ "$os" = "windows" ]; then
        if command -v unzip >/dev/null 2>&1; then
            unzip -q "$archive_name"
        else
            log_error "unzip command not found. Please install unzip or download manually."
            rm -rf "$temp_dir"
            exit 1
        fi
    else
        tar -xzf "$archive_name"
    fi
    
    # Make binary executable
    chmod +x "$binary_name"
    
    # Create install directory if it doesn't exist
    if [ ! -d "$INSTALL_DIR" ]; then
        log_info "Creating install directory: $INSTALL_DIR"
        sudo mkdir -p "$INSTALL_DIR"
    fi
    
    # Install binary
    log_info "Installing ${BINARY_NAME} to ${INSTALL_DIR}..."
    if [ -w "$INSTALL_DIR" ]; then
        mv "$binary_name" "${INSTALL_DIR}/${BINARY_NAME}${extension}"
    else
        sudo mv "$binary_name" "${INSTALL_DIR}/${BINARY_NAME}${extension}"
    fi
    
    # Cleanup
    cd - >/dev/null
    rm -rf "$temp_dir"
    
    log_success "${BINARY_NAME} ${version} installed successfully!"
}

# Verify installation
verify_installation() {
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local installed_version
        installed_version=$($BINARY_NAME --version 2>/dev/null | head -n1 || echo "unknown")
        log_success "Installation verified: $installed_version"
        log_info "Run '${BINARY_NAME} --help' to get started"
    else
        log_warning "Binary installed but not found in PATH"
        log_info "Make sure ${INSTALL_DIR} is in your PATH"
        log_info "You can add it by running: export PATH=\"${INSTALL_DIR}:\$PATH\""
    fi
}

# Show usage information
show_usage() {
    cat << EOF
AsyncAPI Go Code Generator - Installation Script

Usage: $0 [OPTIONS]

Options:
    -v, --version VERSION    Install specific version (default: latest)
    -d, --dir DIRECTORY      Install directory (default: /usr/local/bin)
    -h, --help              Show this help message

Environment Variables:
    INSTALL_DIR             Installation directory
    GITHUB_REPO             GitHub repository (default: jrcryer/evently-codegen)

Examples:
    # Install latest version
    $0

    # Install specific version
    $0 --version v1.2.3

    # Install to custom directory
    $0 --dir ~/.local/bin

    # Install with environment variables
    INSTALL_DIR=~/.local/bin $0
EOF
}

# Main installation function
main() {
    local version=""
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                version="$2"
                shift 2
                ;;
            -d|--dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    log_info "Starting AsyncAPI Go Code Generator installation..."
    
    # Check dependencies
    if ! command -v curl >/dev/null 2>&1; then
        log_error "curl is required but not installed"
        exit 1
    fi
    
    if ! command -v tar >/dev/null 2>&1; then
        log_error "tar is required but not installed"
        exit 1
    fi
    
    # Detect platform
    local platform
    platform=$(detect_platform)
    log_info "Detected platform: $platform"
    
    # Get version
    if [ -z "$version" ]; then
        version=$(get_latest_version)
        log_info "Latest version: $version"
    else
        log_info "Installing version: $version"
    fi
    
    # Install binary
    install_binary "$version" "$platform"
    
    # Verify installation
    verify_installation
    
    log_success "Installation completed!"
}

# Run main function
main "$@"