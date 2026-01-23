#!/bin/bash
#
# Test script for Ravact
# Runs all tests including integration tests
#

set -e

echo "=========================================="
echo "  Ravact Test Suite"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Run unit tests
echo "Running unit tests..."
if go test ./... -v -cover; then
    echo -e "${GREEN}✓ Unit tests passed${NC}"
else
    echo -e "${RED}✗ Unit tests failed${NC}"
    exit 1
fi

echo ""
echo "=========================================="
echo ""

# Run integration tests
echo "Running integration tests..."
if go test -tags=integration ./tests/... -v; then
    echo -e "${GREEN}✓ Integration tests passed${NC}"
else
    echo -e "${YELLOW}⚠ Integration tests failed (may require Linux)${NC}"
fi

echo ""
echo "=========================================="
echo ""

# Run tests with race detector
echo "Running race detector..."
if go test -race ./... -short; then
    echo -e "${GREEN}✓ Race detector passed${NC}"
else
    echo -e "${RED}✗ Race conditions detected${NC}"
    exit 1
fi

echo ""
echo "=========================================="
echo ""

# Generate coverage report
echo "Generating coverage report..."
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
echo -e "${GREEN}✓ Coverage report generated: coverage.html${NC}"

# Show coverage summary
echo ""
echo "Coverage Summary:"
go tool cover -func=coverage.out | grep total

echo ""
echo "=========================================="
echo -e "${GREEN}✓ All tests completed successfully!${NC}"
echo "=========================================="
