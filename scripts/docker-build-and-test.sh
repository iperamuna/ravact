#!/bin/bash
# Quick build and test in AMD64 container
# Run from Mac - automatically builds and tests

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CONTAINER_NAME="ravact-amd64-dev"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Quick Build & Test (AMD64)${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}✗ Docker is not running${NC}"
    exit 1
fi

# Build on Mac
echo -e "${BLUE}Building ravact for AMD64...${NC}"
cd "$PROJECT_DIR"
make build-linux
echo -e "${GREEN}✓ Build complete${NC}"
echo ""

# Check if container exists
if ! docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo -e "${YELLOW}Container doesn't exist. Creating...${NC}"
    ./scripts/docker-amd64-dev.sh
    exit 0
fi

# Start container if not running
if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo -e "${BLUE}Starting container...${NC}"
    docker start ${CONTAINER_NAME} > /dev/null
    echo -e "${GREEN}✓ Container started${NC}"
fi

echo -e "${BLUE}Running ravact in AMD64 container...${NC}"
echo ""
echo "═══════════════════════════════════════"
echo ""

# Run ravact in container
docker exec -it ${CONTAINER_NAME} bash -c 'cd /workspace && sudo ./dist/ravact-linux-amd64'
