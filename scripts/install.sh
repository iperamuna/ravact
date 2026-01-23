#!/bin/bash
# Ravact Installation Script
# Downloads the correct binary for your system and installs to /usr/local/bin

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔═══════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                   Ravact Installation Script                      ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if running with sudo
if [ "$EUID" -ne 0 ]; then 
    echo -e "${YELLOW}⚠️  This script should be run with sudo for system-wide installation${NC}"
    echo ""
    echo "Usage: sudo bash install.sh [--local /path/to/binary]"
    echo ""
    exit 1
fi

# Parse arguments
LOCAL_BINARY=""
if [ "$1" = "--local" ] && [ -n "$2" ]; then
    LOCAL_BINARY="$2"
    echo -e "${YELLOW}⚠️  Local install mode enabled${NC}"
    echo "Using local binary: $LOCAL_BINARY"
    echo ""
fi

# Detect system architecture
echo -e "${BLUE}Detecting system information...${NC}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

echo "Operating System: $OS"
echo "Architecture: $ARCH"
echo ""

# Map architecture to binary name
case "$ARCH" in
    x86_64)
        BINARY_ARCH="amd64"
        ;;
    aarch64)
        BINARY_ARCH="arm64"
        ;;
    arm64)
        BINARY_ARCH="arm64"
        ;;
    *)
        echo -e "${RED}❌ Unsupported architecture: $ARCH${NC}"
        echo "Supported architectures: x86_64 (AMD64), aarch64 (ARM64)"
        exit 1
        ;;
esac

# Map OS to binary name
case "$OS" in
    linux)
        BINARY_OS="linux"
        ;;
    darwin)
        BINARY_OS="darwin"
        ;;
    *)
        echo -e "${RED}❌ Unsupported OS: $OS${NC}"
        echo "Supported systems: Linux, macOS"
        exit 1
        ;;
esac

BINARY_NAME="ravact-${BINARY_OS}-${BINARY_ARCH}"

echo -e "${GREEN}✓ Will download: ${BINARY_NAME}${NC}"
echo ""

# Get latest version (skip if using local binary)
if [ -n "$LOCAL_BINARY" ]; then
    VERSION="local"
else
    echo -e "${BLUE}Fetching latest release information...${NC}"

    LATEST_VERSION=$(curl -s https://api.github.com/repos/iperamuna/ravact/releases/latest | grep -o '"tag_name": "[^"]*"' | cut -d'"' -f4)

    if [ -z "$LATEST_VERSION" ]; then
        echo -e "${YELLOW}⚠️  Failed to fetch latest version from GitHub${NC}"
        echo "This may indicate no internet access or GitHub API is blocked."
        echo ""
        read -p "Enter version to install (e.g., v1.0.0): " MANUAL_VERSION
        if [ -z "$MANUAL_VERSION" ]; then
            echo -e "${RED}❌ No version provided. Exiting.${NC}"
            exit 1
        fi
        VERSION="$MANUAL_VERSION"
        echo -e "${GREEN}✓ Using manually specified version: ${VERSION}${NC}"
    else
        echo -e "${GREEN}✓ Latest version: ${LATEST_VERSION}${NC}"
        echo ""

        # Allow custom version
        read -p "Install version [${LATEST_VERSION}]: " CUSTOM_VERSION
        VERSION=${CUSTOM_VERSION:-$LATEST_VERSION}
    fi
fi

echo ""
echo -e "${BLUE}Downloading Ravact ${VERSION}...${NC}"

DOWNLOAD_URL="https://github.com/iperamuna/ravact/releases/download/${VERSION}/${BINARY_NAME}"

# Create temporary directory
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

BINARY_PATH="$TEMP_DIR/ravact"

if [ -n "$LOCAL_BINARY" ]; then
    if [ ! -f "$LOCAL_BINARY" ]; then
        echo -e "${RED}❌ Local binary not found: $LOCAL_BINARY${NC}"
        exit 1
    fi
    echo -e "${YELLOW}⚠️  Using local binary instead of download${NC}"
    cp "$LOCAL_BINARY" "$BINARY_PATH"
else
    echo "URL: $DOWNLOAD_URL"
    echo ""
fi

# Download binary
if [ -z "$LOCAL_BINARY" ]; then
    if ! curl -L --progress-bar "$DOWNLOAD_URL" -o "$BINARY_PATH" 2>/dev/null; then
        echo -e "${RED}❌ Failed to download binary${NC}"
        echo "Please check:"
        echo "  • Your internet connection"
        echo "  • The version exists: ${VERSION}"
        echo "  • Your system is supported: ${BINARY_OS}/${BINARY_ARCH}"
        echo "  • Or use --local /path/to/binary"
        exit 1
    fi

    echo ""
    echo -e "${GREEN}✓ Download complete${NC}"
    echo ""
fi

# Verify binary
echo -e "${BLUE}Verifying binary...${NC}"

if [ ! -f "$BINARY_PATH" ]; then
    echo -e "${RED}❌ Binary file not found${NC}"
    exit 1
fi

# Check if executable
if ! file "$BINARY_PATH" | grep -q "executable"; then
    echo -e "${YELLOW}⚠️  Making binary executable${NC}"
fi

chmod +x "$BINARY_PATH"

echo -e "${GREEN}✓ Binary verified and executable${NC}"
echo ""

# Install
echo -e "${BLUE}Installing to /usr/local/bin/ravact...${NC}"

# Backup existing installation if present
if [ -f "/usr/local/bin/ravact" ]; then
    echo -e "${YELLOW}⚠️  Existing installation found, creating backup${NC}"
    cp /usr/local/bin/ravact /usr/local/bin/ravact.bak
    echo "Backup saved to: /usr/local/bin/ravact.bak"
fi

# Copy to /usr/local/bin
cp "$BINARY_PATH" /usr/local/bin/ravact

# Make sure it's executable
chmod +x /usr/local/bin/ravact

echo -e "${GREEN}✓ Installation complete${NC}"
echo ""

# Verify installation
echo -e "${BLUE}Verifying installation...${NC}"

if ! command -v ravact &> /dev/null; then
    echo -e "${RED}❌ Installation verification failed${NC}"
    exit 1
fi

INSTALLED_VERSION=$(/usr/local/bin/ravact --version 2>/dev/null || echo "unknown")
echo -e "${GREEN}✓ Ravact installed successfully${NC}"
echo ""

# Display summary
echo -e "${BLUE}╔═══════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                     Installation Summary                          ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}Location:${NC} /usr/local/bin/ravact"
echo -e "${GREEN}Version:${NC} ${VERSION}"
echo -e "${GREEN}System:${NC} ${BINARY_OS}/${BINARY_ARCH}"
echo ""

# Usage instructions
echo -e "${YELLOW}Usage:${NC}"
echo ""
echo "  Start Ravact:"
echo -e "    ${BLUE}sudo ravact${NC}"
echo ""
echo "  For help:"
echo -e "    ${BLUE}ravact --help${NC}"
echo ""

# Check for updates
echo -e "${YELLOW}Stay updated:${NC}"
echo ""
echo "  To reinstall or update to a newer version, run this script again:"
echo -e "    ${BLUE}curl -sSL https://raw.githubusercontent.com/iperamuna/ravact/main/scripts/install.sh | sudo bash${NC}"
echo ""

echo -e "${GREEN}✅ Installation successful!${NC}"
echo ""
echo -e "You can now run: ${BLUE}sudo ravact${NC}"
echo ""
