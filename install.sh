#!/bin/bash

set -e

# Configuration
BINARY_NAME="ai-git"
REPO="anans9/ai-git"
INSTALL_DIR="/usr/local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print functions
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Check if running as root
check_root() {
    if [[ $EUID -eq 0 ]]; then
        print_error "Don't run this script as root!"
        exit 1
    fi
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64) ARCH="x86_64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    case "$OS" in
        linux) OS="Linux" ;;
        darwin) OS="Darwin" ;;
        *)
            print_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac

    print_info "Detected platform: $OS $ARCH"
}

# Get latest release version
get_latest_version() {
    print_info "Getting latest release..."

    LATEST_VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$LATEST_VERSION" ]; then
        print_error "Failed to get latest version"
        exit 1
    fi

    print_info "Latest version: $LATEST_VERSION"
}

# Download and install
install_binary() {
    ARCHIVE_NAME="$BINARY_NAME-$LATEST_VERSION-$OS-$ARCH.tar.gz"
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_VERSION/$ARCHIVE_NAME"

    print_info "Downloading $ARCHIVE_NAME..."

    # Create temporary directory
    TMP_DIR=$(mktemp -d)
    cd "$TMP_DIR"

    # Download archive
    if ! curl -L "$DOWNLOAD_URL" -o "$ARCHIVE_NAME"; then
        print_error "Failed to download $DOWNLOAD_URL"
        exit 1
    fi

    # Extract archive
    print_info "Extracting archive..."
    tar -xzf "$ARCHIVE_NAME"

    # Find the binary (it might be in a subdirectory)
    BINARY_PATH=$(find . -name "$BINARY_NAME" -type f | head -1)

    if [ ! -f "$BINARY_PATH" ]; then
        print_error "Binary not found in archive"
        exit 1
    fi

    # Make binary executable
    chmod +x "$BINARY_PATH"

    # Install binary
    print_info "Installing to $INSTALL_DIR..."

    if [ -w "$INSTALL_DIR" ]; then
        mv "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
    else
        sudo mv "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
    fi

    # Cleanup
    cd - > /dev/null
    rm -rf "$TMP_DIR"

    print_success "Installed $BINARY_NAME to $INSTALL_DIR/$BINARY_NAME"
}

# Verify installation
verify_installation() {
    if command -v "$BINARY_NAME" > /dev/null 2>&1; then
        VERSION=$($BINARY_NAME --version 2>/dev/null || echo "unknown")
        print_success "Installation verified: $VERSION"
    else
        print_warning "Binary installed but not in PATH. You may need to restart your shell or add $INSTALL_DIR to your PATH."
    fi
}

# Main installation flow
main() {
    echo ""
    echo "🚀 AI-Git CLI Installer"
    echo "======================="
    echo ""

    check_root
    detect_platform
    get_latest_version
    install_binary
    verify_installation

    echo ""
    print_success "Installation complete!"
    echo ""
    print_info "Get started:"
    echo "  $BINARY_NAME --help"
    echo "  $BINARY_NAME init"
    echo "  $BINARY_NAME config providers set openai api_key YOUR_API_KEY"
    echo ""
    print_info "To uninstall:"
    echo "  $BINARY_NAME uninstall"
    echo ""
}

# Run main function
main "$@"
