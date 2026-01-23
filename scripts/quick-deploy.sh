#!/bin/bash
# Quick deployment script - Build on Mac and deploy to VM
# Usage: ./quick-deploy.sh [VM_IP] [VM_USER]

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

# Default values
VM_IP="${1:-}"
VM_USER="${2:-devuser}"
PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Quick Deploy to VM${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if VM IP is provided
if [[ -z "$VM_IP" ]]; then
    # Try to read from SSH config
    if [[ -f "$HOME/.ssh/config" ]] && grep -q "Host ravact-dev" "$HOME/.ssh/config"; then
        VM_IP=$(grep -A 5 "Host ravact-dev" "$HOME/.ssh/config" | grep "HostName" | awk '{print $2}')
        echo -e "${GREEN}✓ Found VM IP in SSH config: $VM_IP${NC}"
    else
        echo -e "${RED}✗ No VM IP provided${NC}"
        echo ""
        echo "Usage: $0 <VM_IP> [VM_USER]"
        echo ""
        echo "Example:"
        echo "  $0 192.168.64.5 devuser"
        echo ""
        exit 1
    fi
fi

echo "VM IP: $VM_IP"
echo "VM User: $VM_USER"
echo "Project: $PROJECT_DIR"
echo ""

# Test SSH connection
echo -e "${BLUE}Testing SSH connection...${NC}"
if ! ssh -o ConnectTimeout=5 ${VM_USER}@${VM_IP} "echo 'Connected'" &> /dev/null; then
    echo -e "${RED}✗ Cannot connect to VM${NC}"
    echo "Please check:"
    echo "  - VM is running"
    echo "  - IP address is correct: $VM_IP"
    echo "  - SSH is working: ssh ${VM_USER}@${VM_IP}"
    exit 1
fi
echo -e "${GREEN}✓ SSH connection successful${NC}"
echo ""

# Build ravact
echo -e "${BLUE}Building ravact for Linux ARM64...${NC}"
cd "$PROJECT_DIR"
make build-linux-arm64
echo -e "${GREEN}✓ Build complete${NC}"
echo ""

# Create directories on VM if they don't exist
echo -e "${BLUE}Preparing VM directories...${NC}"
ssh ${VM_USER}@${VM_IP} "mkdir -p ~/ravact-go/assets/scripts ~/ravact-go/assets/configs"
echo -e "${GREEN}✓ Directories ready${NC}"
echo ""

# Deploy ravact binary
echo -e "${BLUE}Deploying ravact binary...${NC}"
scp dist/ravact-linux-arm64 ${VM_USER}@${VM_IP}:~/ravact-go/ravact
ssh ${VM_USER}@${VM_IP} "chmod +x ~/ravact-go/ravact"
echo -e "${GREEN}✓ Binary deployed${NC}"
echo ""

# Deploy assets
echo -e "${BLUE}Deploying assets...${NC}"
scp -r assets/scripts/* ${VM_USER}@${VM_IP}:~/ravact-go/assets/scripts/ 2>/dev/null || true
scp -r assets/configs/* ${VM_USER}@${VM_IP}:~/ravact-go/assets/configs/ 2>/dev/null || true
echo -e "${GREEN}✓ Assets deployed${NC}"
echo ""

# Test ravact
echo -e "${BLUE}Testing ravact on VM...${NC}"
ssh ${VM_USER}@${VM_IP} "cd ~/ravact-go && ./ravact --version 2>/dev/null || echo 'Ravact ready (needs sudo to run)'"
echo ""

# Summary
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Deployment Complete! 🚀${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "To run ravact on VM:"
echo "  ssh ${VM_USER}@${VM_IP}"
echo "  cd ~/ravact-go"
echo "  sudo ./ravact"
echo ""
echo "Or run directly:"
echo "  ssh ${VM_USER}@${VM_IP} 'cd ravact-go && sudo ./ravact'"
echo ""
