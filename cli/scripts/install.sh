#!/bin/sh
# Envie CLI installer
# Usage: curl -fsSL https://envie.sh/install.sh | sh
#
# Environment variables:
#   ENVIE_INSTALL_DIR - Installation directory (default: /usr/local/bin or ~/.local/bin)
#   ENVIE_VERSION     - Specific version to install (default: latest)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

GITHUB_REPO="stranavad/envie"
BINARY_NAME="envie"

# Print colored message
info() {
    printf "${BLUE}==>${NC} %s\n" "$1"
}

success() {
    printf "${GREEN}==>${NC} %s\n" "$1"
}

warn() {
    printf "${YELLOW}Warning:${NC} %s\n" "$1"
}

error() {
    printf "${RED}Error:${NC} %s\n" "$1" >&2
    exit 1
}

# Detect OS
detect_os() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$OS" in
        linux*)  OS="linux" ;;
        darwin*) OS="darwin" ;;
        msys*|mingw*|cygwin*) OS="windows" ;;
        *)       error "Unsupported operating system: $OS" ;;
    esac
    echo "$OS"
}

# Detect architecture
detect_arch() {
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64|amd64)  ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        armv7l)        ARCH="arm" ;;
        i386|i686)     ARCH="386" ;;
        *)             error "Unsupported architecture: $ARCH" ;;
    esac
    echo "$ARCH"
}

# Get the latest version from GitHub
get_latest_version() {
    if command -v curl > /dev/null 2>&1; then
        curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" |
            grep '"tag_name":' |
            sed -E 's/.*"([^"]+)".*/\1/'
    elif command -v wget > /dev/null 2>&1; then
        wget -qO- "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" |
            grep '"tag_name":' |
            sed -E 's/.*"([^"]+)".*/\1/'
    else
        error "Neither curl nor wget found. Please install one of them."
    fi
}

# Determine install directory
get_install_dir() {
    if [ -n "$ENVIE_INSTALL_DIR" ]; then
        echo "$ENVIE_INSTALL_DIR"
        return
    fi

    # Try /usr/local/bin first (requires sudo on most systems)
    if [ -w "/usr/local/bin" ]; then
        echo "/usr/local/bin"
        return
    fi

    # Fall back to ~/.local/bin
    LOCAL_BIN="${HOME}/.local/bin"
    mkdir -p "$LOCAL_BIN"
    echo "$LOCAL_BIN"
}

# Check if directory is in PATH
check_path() {
    DIR="$1"
    case ":$PATH:" in
        *":$DIR:"*) return 0 ;;
        *)          return 1 ;;
    esac
}

# Download and install
install() {
    OS=$(detect_os)
    ARCH=$(detect_arch)

    info "Detected OS: $OS, Architecture: $ARCH"

    # Get version
    if [ -n "$ENVIE_VERSION" ]; then
        VERSION="$ENVIE_VERSION"
    else
        info "Fetching latest version..."
        VERSION=$(get_latest_version)
        if [ -z "$VERSION" ]; then
            error "Failed to determine latest version"
        fi
    fi

    info "Installing Envie CLI $VERSION"

    # Construct download URL
    # Binary naming: envie-{os}-{arch} or envie-{os}-{arch}.exe for Windows
    EXT=""
    if [ "$OS" = "windows" ]; then
        EXT=".exe"
    fi

    FILENAME="${BINARY_NAME}-${OS}-${ARCH}${EXT}"
    DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/cli-${FILENAME}"

    # Get install directory
    INSTALL_DIR=$(get_install_dir)
    INSTALL_PATH="${INSTALL_DIR}/${BINARY_NAME}${EXT}"

    info "Downloading from $DOWNLOAD_URL"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    TMP_FILE="${TMP_DIR}/${FILENAME}"

    # Download
    if command -v curl > /dev/null 2>&1; then
        HTTP_CODE=$(curl -fsSL -w "%{http_code}" -o "$TMP_FILE" "$DOWNLOAD_URL" 2>/dev/null) || true
        if [ "$HTTP_CODE" != "200" ]; then
            rm -rf "$TMP_DIR"
            error "Download failed (HTTP $HTTP_CODE). Check if version $VERSION exists for $OS-$ARCH"
        fi
    elif command -v wget > /dev/null 2>&1; then
        wget -q -O "$TMP_FILE" "$DOWNLOAD_URL" || {
            rm -rf "$TMP_DIR"
            error "Download failed. Check if version $VERSION exists for $OS-$ARCH"
        }
    fi

    # Verify download
    if [ ! -f "$TMP_FILE" ] || [ ! -s "$TMP_FILE" ]; then
        rm -rf "$TMP_DIR"
        error "Download failed or file is empty"
    fi

    # Make executable
    chmod +x "$TMP_FILE"

    # Install
    info "Installing to $INSTALL_PATH"

    if [ -w "$INSTALL_DIR" ]; then
        mv "$TMP_FILE" "$INSTALL_PATH"
    else
        info "Requesting sudo access to install to $INSTALL_DIR"
        sudo mv "$TMP_FILE" "$INSTALL_PATH"
    fi

    # Cleanup
    rm -rf "$TMP_DIR"

    # Verify installation
    if [ -x "$INSTALL_PATH" ]; then
        success "Envie CLI $VERSION installed successfully!"

        # Check if in PATH
        if ! check_path "$INSTALL_DIR"; then
            warn "$INSTALL_DIR is not in your PATH"
            echo ""
            echo "Add it to your shell configuration:"
            echo ""
            echo "  # For bash (~/.bashrc or ~/.bash_profile)"
            echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
            echo ""
            echo "  # For zsh (~/.zshrc)"
            echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
            echo ""
            echo "Then restart your terminal or run: source ~/.bashrc"
            echo ""
        fi

        echo ""
        echo "Get started:"
        echo "  envie --help"
        echo "  envie auth --token <your-token>"
        echo "  envie export --project <project-id>"
        echo ""
    else
        error "Installation failed"
    fi
}

# Run uninstall
uninstall() {
    INSTALL_DIR=$(get_install_dir)
    INSTALL_PATH="${INSTALL_DIR}/${BINARY_NAME}"

    if [ -f "$INSTALL_PATH" ]; then
        info "Removing $INSTALL_PATH"
        if [ -w "$INSTALL_DIR" ]; then
            rm "$INSTALL_PATH"
        else
            sudo rm "$INSTALL_PATH"
        fi
        success "Envie CLI uninstalled"
    else
        warn "Envie CLI not found at $INSTALL_PATH"
    fi
}

# Parse arguments
case "${1:-}" in
    uninstall)
        uninstall
        ;;
    *)
        install
        ;;
esac
