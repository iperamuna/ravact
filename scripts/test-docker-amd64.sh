#!/bin/bash
# Test ravact in x86_64 Docker container
# Uses QEMU emulation on M1 Mac

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Ravact AMD64 Docker Testing${NC}"
echo -e "${BLUE}x86_64 via QEMU Emulation${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_DIR"

echo -e "${YELLOW}Note: This will be slower due to QEMU emulation${NC}"
echo ""

# Check if docker is available
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker not found. Install Docker Desktop for Mac${NC}"
    exit 1
fi

echo -e "${BLUE}Building ravact for AMD64...${NC}"
make build-linux
echo -e "${GREEN}✓ Build complete${NC}"
echo ""

echo -e "${BLUE}Starting x86_64 Ubuntu container...${NC}"
echo ""

docker run --rm -it \
    --platform linux/amd64 \
    -v "$PROJECT_DIR:/workspace" \
    -w /workspace \
    ubuntu:24.04 \
    bash -c '
echo "====================================="
echo "Ubuntu x86_64 Container"
echo "====================================="
echo ""
echo "Architecture: $(uname -m)"
echo "OS: $(cat /etc/os-release | grep PRETTY_NAME | cut -d= -f2 | tr -d \")"
echo ""

# Install minimal dependencies
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq sudo > /dev/null 2>&1

# Make binary executable
chmod +x dist/ravact-linux-amd64

echo ""
echo "✓ Container ready!"
echo ""
echo "You can now run:"
echo "  sudo ./dist/ravact-linux-amd64"
echo ""
echo "Or explore the x86_64 environment"
echo ""

# Start bash
exec bash
'
