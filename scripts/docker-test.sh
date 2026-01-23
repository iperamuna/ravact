#!/bin/bash
#
# Docker Test Script for Ravact
# Tests the application in Ubuntu 24.04 container
#

set -e

echo "=========================================="
echo "  Ravact Docker Test"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running"
    exit 1
fi

echo "Building test Docker image..."
docker build -t ravact-test -f Dockerfile.test .

echo ""
echo "Running tests in Ubuntu 24.04 container..."
docker run --rm ravact-test

echo ""
echo -e "${GREEN}✓ Docker tests completed successfully!${NC}"
echo ""

# Optional: Run the binary in Docker to test it works
echo "Testing compiled binary in Ubuntu 24.04..."
docker run --rm -v $(pwd):/workspace -w /workspace ubuntu:24.04 bash -c "
    apt-get update -qq > /dev/null 2>&1 && \
    ./dist/ravact-linux-amd64 --version
"

echo ""
echo -e "${GREEN}✓ Binary works in Ubuntu 24.04!${NC}"
echo ""
echo "=========================================="
echo "You can also test interactively with:"
echo "  make docker-shell"
echo "=========================================="
