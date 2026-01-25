#!/bin/bash
# This script runs ONLY on the Ubuntu VM to set it up for ravact development
# Usage: ssh into your VM and run: bash <(curl -s URL_TO_THIS_SCRIPT)
# Or: Copy this script to VM and run: bash setup-vm-only.sh

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Ravact VM Setup Script${NC}"
echo -e "${BLUE}Ubuntu 24.04 ARM64${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if running on Linux
if [[ "$(uname)" != "Linux" ]]; then
    echo -e "${RED}✗ This script must run on Linux (inside the VM)${NC}"
    exit 1
fi

# Check architecture
ARCH=$(uname -m)
if [[ "$ARCH" != "aarch64" ]] && [[ "$ARCH" != "arm64" ]]; then
    echo -e "${YELLOW}⚠ Warning: Expected ARM64 architecture, got: $ARCH${NC}"
    echo -e "${YELLOW}⚠ Script will continue but may not work correctly${NC}"
fi

echo -e "${GREEN}✓ Running on Linux $ARCH${NC}"
echo ""

# Update system
echo -e "${BLUE}Updating system packages...${NC}"
sudo apt update
sudo apt upgrade -y
echo -e "${GREEN}✓ System updated${NC}"
echo ""

# Install essential tools
echo -e "${BLUE}Installing essential development tools...${NC}"
sudo apt install -y \
    build-essential \
    curl \
    wget \
    git \
    vim \
    htop \
    net-tools \
    tree \
    jq \
    unzip \
    tmux \
    software-properties-common
echo -e "${GREEN}✓ Essential tools installed${NC}"
echo ""

# Install Go 1.24 (latest stable)
echo -e "${BLUE}Installing Go 1.24...${NC}"
GO_VERSION="1.24.0"
GO_TAR="go${GO_VERSION}.linux-arm64.tar.gz"

# Detect architecture for Go download
if [[ "$ARCH" == "x86_64" ]] || [[ "$ARCH" == "amd64" ]]; then
    GO_TAR="go${GO_VERSION}.linux-amd64.tar.gz"
fi

cd /tmp
wget -q --show-progress https://go.dev/dl/${GO_TAR}

# Remove old Go installation
sudo rm -rf /usr/local/go

# Install new Go
sudo tar -C /usr/local -xzf ${GO_TAR}

# Setup Go environment
if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
    cat >> ~/.bashrc << 'EOF'

# Go environment
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
EOF
    echo -e "${GREEN}✓ Go environment variables added to ~/.bashrc${NC}"
fi

# Source for current session
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Verify Go installation
echo ""
echo -e "${GREEN}✓ Go installed successfully:${NC}"
go version
echo ""

# Create project directories
echo -e "${BLUE}Creating project directories...${NC}"
mkdir -p ~/ravact-go/assets/scripts
mkdir -p ~/ravact-go/assets/configs
mkdir -p ~/go/src
echo -e "${GREEN}✓ Directories created${NC}"
echo ""

# Create helper scripts
echo -e "${BLUE}Creating helper scripts...${NC}"

# Test script
cat > ~/ravact-go/test-ravact.sh << 'EOF'
#!/bin/bash
# Quick test script for ravact

cd ~/ravact-go

if [[ ! -f "./ravact" ]]; then
    echo "Error: ravact binary not found"
    echo "Please copy ravact binary to ~/ravact-go/"
    exit 1
fi

echo "Testing ravact..."
./ravact --version 2>/dev/null || ./ravact --help 2>/dev/null || echo "Run with: sudo ./ravact"

echo ""
echo "To run ravact:"
echo "  cd ~/ravact-go"
echo "  sudo ./ravact"
EOF
chmod +x ~/ravact-go/test-ravact.sh

# Build script (if source is available)
cat > ~/ravact-go/build-ravact.sh << 'EOF'
#!/bin/bash
# Build ravact from source

cd ~/ravact-go

if [[ ! -f "go.mod" ]]; then
    echo "Error: go.mod not found"
    echo "This script should be run in the ravact-go source directory"
    exit 1
fi

echo "Building ravact..."
go build -o ravact ./cmd/ravact

if [[ $? -eq 0 ]]; then
    echo "✓ Build successful"
    echo ""
    echo "Run with: sudo ./ravact"
else
    echo "✗ Build failed"
    exit 1
fi
EOF
chmod +x ~/ravact-go/build-ravact.sh

echo -e "${GREEN}✓ Helper scripts created${NC}"
echo ""

# Setup Git (optional but recommended)
echo -e "${BLUE}Configuring Git...${NC}"
if [[ -z "$(git config --global user.name 2>/dev/null)" ]]; then
    echo "Git user not configured. You can set it up later with:"
    echo "  git config --global user.name 'Your Name'"
    echo "  git config --global user.email 'you@example.com'"
else
    echo -e "${GREEN}✓ Git already configured${NC}"
fi
echo ""

# Display system info
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}System Information${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "OS: $(lsb_release -d | cut -f2)"
echo "Kernel: $(uname -r)"
echo "Architecture: $(uname -m)"
echo "CPU Cores: $(nproc)"
echo "RAM: $(free -h | awk '/^Mem:/ {print $2}')"
echo "Disk: $(df -h / | awk 'NR==2 {print $4}') free"
echo ""
echo "Go Version: $(go version)"
echo "Go Path: $(go env GOPATH)"
echo ""

# Summary
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}VM Setup Complete! 🎉${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${GREEN}What's installed:${NC}"
echo "  ✓ Ubuntu 24.04 fully updated"
echo "  ✓ Go 1.24.0 for $(uname -m)"
echo "  ✓ Build tools (gcc, make, etc.)"
echo "  ✓ Development utilities"
echo "  ✓ Project directories: ~/ravact-go"
echo ""
echo -e "${GREEN}Next steps:${NC}"
echo ""
echo "1. Copy ravact binary to VM:"
echo "   From your Mac:"
echo "   scp dist/ravact-linux-arm64 user@vm-ip:~/ravact-go/ravact"
echo ""
echo "2. Or clone and build from source:"
echo "   cd ~/ravact-go"
echo "   git clone <your-repo> ."
echo "   go build -o ravact ./cmd/ravact"
echo ""
echo "3. Run ravact:"
echo "   cd ~/ravact-go"
echo "   sudo ./ravact"
echo ""
echo -e "${GREEN}Helper scripts available:${NC}"
echo "  ~/ravact-go/test-ravact.sh   - Test ravact binary"
echo "  ~/ravact-go/build-ravact.sh  - Build from source"
echo ""
echo -e "${BLUE}To apply environment changes, run:${NC}"
echo "  source ~/.bashrc"
echo ""
echo -e "${GREEN}Happy developing! 🚀${NC}"
echo ""
