#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

REPO="PaleBlueDot-AI-Open/pbd-cli"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="pbd-cli"
DEV_MODE=false

# Print colored message
info() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --dev)
            DEV_MODE=true
            BINARY_NAME="pbd-cli-dev"
            shift
            ;;
        *)
            warn "Unknown option: $1"
            shift
            ;;
    esac
done

# Check for curl
if ! command -v curl &> /dev/null; then
    error "curl is required but not installed. Please install curl first."
fi

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case $OS in
    linux|darwin) ;;
    mingw*|msys*|cygwin*)
        error "Windows detected. Please use PowerShell or download from: https://github.com/$REPO/releases"
        ;;
    *) error "Unsupported OS: $OS" ;;
esac

# Detect ARCH
ARCH=$(uname -m)
case $ARCH in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) error "Unsupported architecture: $ARCH" ;;
esac

# Get latest version
info "Fetching latest release..."
if [ "$DEV_MODE" = true ]; then
    # Get latest dev release (tags ending with -dev)
    LATEST=$(curl -s "https://api.github.com/repos/$REPO/releases?per_page=100" | \
             grep -E '"tag_name": "v.*-dev"' | \
             head -1 | \
             sed -E 's/.*"([^"]+)".*/\1/')
else
    # Get latest prod release
    LATEST=$(curl -s https://api.github.com/repos/$REPO/releases/latest | \
             grep '"tag_name":' | \
             sed -E 's/.*"([^"]+)".*/\1/')
fi

if [ -z "$LATEST" ]; then
    error "Failed to get latest version. Check your internet connection."
fi

info "Latest version: $LATEST"

# Build download URL based on mode
if [ "$DEV_MODE" = true ]; then
    # Dev release files: pbd-cli_1.0.0-dev_${OS}_${ARCH}.tar.gz
    # Tag format: v1.0.0-dev
    VERSION="${LATEST#v}"
    URL="https://github.com/$REPO/releases/download/$LATEST/pbd-cli_${VERSION}_${OS}_${ARCH}.tar.gz"
    info "Installing DEV version (test environment)"
else
    # Prod release files: pbd-cli_1.0.0_${OS}_${ARCH}.tar.gz
    # Tag format: v1.0.0
    VERSION="${LATEST#v}"
    URL="https://github.com/$REPO/releases/download/$LATEST/pbd-cli_${VERSION}_${OS}_${ARCH}.tar.gz"
    info "Installing PROD version (production environment)"
fi

info "Downloading from: $URL"

# Download and extract
if ! curl -sL --fail "$URL" -o /tmp/pbd-cli.tar.gz; then
    error "Download failed. The release for your platform may not exist yet."
fi

info "Extracting..."
tar xzf /tmp/pbd-cli.tar.gz -C /tmp

# Install
info "Installing to $INSTALL_DIR..."

# The binary inside the archive is always named 'pbd-cli'
# We rename it to the target binary name during install
if [ -w "$INSTALL_DIR" ]; then
    mv /tmp/pbd-cli "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    info "Requires sudo to install to $INSTALL_DIR"
    sudo mv /tmp/pbd-cli "$INSTALL_DIR/$BINARY_NAME"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

# Cleanup
rm -f /tmp/pbd-cli.tar.gz

# Verify
if command -v $BINARY_NAME &> /dev/null; then
    info "Successfully installed $BINARY_NAME $LATEST"
    info "Run '$BINARY_NAME --help' to get started"
else
    warn "Installed but '$BINARY_NAME' not in PATH. Add $INSTALL_DIR to your PATH."
fi
