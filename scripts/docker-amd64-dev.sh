#!/bin/bash
# Docker AMD64 Development Container
# Keeps container running for iterative testing
# Auto-syncs code via volume mount

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CONTAINER_NAME="ravact-amd64-dev"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Ravact AMD64 Development Container${NC}"
echo -e "${BLUE}Live Code Sync via Volume Mount${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}✗ Docker is not running${NC}"
    echo ""
    echo "Please start Docker Desktop:"
    echo "  1. Open Docker from Applications"
    echo "  2. Wait for whale icon to be stable"
    echo "  3. Run this script again"
    exit 1
fi

echo -e "${GREEN}✓ Docker is running${NC}"
echo ""

# Check if container already exists
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo -e "${YELLOW}Container '${CONTAINER_NAME}' already exists${NC}"
    
    # Check if it's running
    if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        echo -e "${GREEN}✓ Container is already running${NC}"
        echo ""
        echo "Connect to it with:"
        echo "  docker exec -it ${CONTAINER_NAME} bash"
        echo ""
        read -p "Connect now? (Y/n): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Nn]$ ]]; then
            docker exec -it ${CONTAINER_NAME} bash
        fi
        exit 0
    else
        echo -e "${BLUE}Starting existing container...${NC}"
        docker start ${CONTAINER_NAME}
        docker exec -it ${CONTAINER_NAME} bash
        exit 0
    fi
fi

echo -e "${BLUE}Creating new AMD64 development container...${NC}"
echo ""

# Create and start container
docker run -it \
    --name ${CONTAINER_NAME} \
    --platform linux/amd64 \
    -v "${PROJECT_DIR}:/workspace" \
    -w /workspace \
    ubuntu:24.04 \
    bash -c '
echo "====================================="
echo "AMD64 Development Container"
echo "====================================="
echo ""
echo "Architecture: $(uname -m)"
echo "OS: $(cat /etc/os-release | grep PRETTY_NAME | cut -d= -f2 | tr -d \")"
echo ""

# Install dependencies
echo "Installing dependencies..."
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq sudo vim nano curl wget > /dev/null 2>&1

echo ""
echo "✓ Container ready!"
echo ""
echo "═══════════════════════════════════════"
echo "DEVELOPMENT WORKFLOW"
echo "═══════════════════════════════════════"
echo ""
echo "Your code is mounted at: /workspace"
echo "Changes on Mac appear instantly here!"
echo ""
echo "To test ravact:"
echo "  1. On Mac: make build-linux"
echo "  2. In container: sudo ./dist/ravact-linux-amd64"
echo "  3. Make changes on Mac"
echo "  4. Rebuild on Mac"
echo "  5. Test again (no sync needed!)"
echo ""
echo "To exit: type \"exit\" or Ctrl+D"
echo "To reconnect: docker exec -it ravact-amd64-dev bash"
echo ""
echo "═══════════════════════════════════════"
echo ""

# Start interactive shell
cd /workspace
exec bash
'
