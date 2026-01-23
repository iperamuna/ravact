#!/bin/bash
# Setup script for Mac - Creates Ubuntu VM and deploys ravact
# For M1 MacBook Pro with UTM

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
VM_NAME="ravact-dev"
VM_USER="devuser"
VM_PASS="devuser"
VM_IP=""

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Ravact VM Setup for M1 Mac${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to print colored messages
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Check if running on M1 Mac
print_info "Checking system architecture..."
ARCH=$(uname -m)
if [[ "$ARCH" != "arm64" ]]; then
    print_error "This script is designed for Apple Silicon (M1/M2) Macs"
    print_info "Your architecture: $ARCH"
    exit 1
fi
print_success "Apple Silicon detected: $ARCH"

# Check if UTM is installed
print_info "Checking for UTM..."
if ! command -v utm &> /dev/null; then
    print_warning "UTM not found. Installing..."
    if command -v brew &> /dev/null; then
        brew install --cask utm
        print_success "UTM installed"
    else
        print_error "Homebrew not found. Please install UTM manually from https://mac.getutm.app/"
        exit 1
    fi
else
    print_success "UTM is installed"
fi

# Check if Ubuntu ISO exists
ISO_PATH="$HOME/Downloads/ubuntu-24.04-live-server-arm64.iso"
print_info "Checking for Ubuntu 24.04 ARM64 ISO..."
if [[ ! -f "$ISO_PATH" ]]; then
    print_warning "Ubuntu ISO not found. Downloading..."
    echo ""
    print_info "This will download ~2.5GB. Please wait..."
    cd "$HOME/Downloads"
    curl -# -LO https://cdimage.ubuntu.com/releases/24.04/release/ubuntu-24.04-live-server-arm64.iso
    print_success "Ubuntu ISO downloaded"
else
    print_success "Ubuntu ISO found: $ISO_PATH"
fi

# Instructions for VM creation (UTM doesn't have good CLI support)
echo ""
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}MANUAL STEP: Create VM in UTM${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""
echo "Please create a VM in UTM with these settings:"
echo ""
echo "  1. Open UTM → Create a New Virtual Machine"
echo "  2. Select: Virtualize (not Emulate)"
echo "  3. Operating System: Linux"
echo "  4. Settings:"
echo "     - Name: ${VM_NAME}"
echo "     - Architecture: ARM64 (aarch64)"
echo "     - ISO: $ISO_PATH"
echo "     - Memory: 4096 MB (or 8192 MB if you have 16GB+ RAM)"
echo "     - CPU Cores: 4"
echo "     - Storage: 20 GB"
echo ""
echo "  5. Install Ubuntu Server:"
echo "     - Username: ${VM_USER}"
echo "     - Password: ${VM_PASS}"
echo "     - Server name: ${VM_NAME}"
echo "     - ✓ Install OpenSSH server (IMPORTANT!)"
echo ""
echo "  6. After installation, remove the ISO and reboot"
echo ""
echo "  7. Get the VM IP address:"
echo "     - In VM console, run: ip addr show"
echo "     - Look for IP like: 192.168.64.x"
echo ""

read -p "Press Enter when VM is created and you have the IP address..."

# Get VM IP
echo ""
read -p "Enter VM IP address: " VM_IP

if [[ -z "$VM_IP" ]]; then
    print_error "No IP address provided"
    exit 1
fi

# Test SSH connection
print_info "Testing SSH connection to $VM_IP..."
if ssh -o ConnectTimeout=5 -o StrictHostKeyChecking=no ${VM_USER}@${VM_IP} "echo 'SSH connection successful'" &> /dev/null; then
    print_success "SSH connection successful"
else
    print_error "Cannot connect to VM via SSH"
    print_info "Please ensure:"
    print_info "  1. VM is running"
    print_info "  2. OpenSSH server is installed"
    print_info "  3. IP address is correct: $VM_IP"
    exit 1
fi

# Setup SSH config for easy access
print_info "Configuring SSH..."
SSH_CONFIG="$HOME/.ssh/config"
if ! grep -q "Host ${VM_NAME}" "$SSH_CONFIG" 2>/dev/null; then
    mkdir -p "$HOME/.ssh"
    cat >> "$SSH_CONFIG" << EOF

# Ravact Development VM
Host ${VM_NAME}
    HostName ${VM_IP}
    User ${VM_USER}
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
EOF
    print_success "SSH config updated"
else
    print_info "SSH config already exists for ${VM_NAME}"
fi

# Create VM setup script
print_info "Creating VM setup script..."
VM_SETUP_SCRIPT=$(cat << 'EOFSCRIPT'
#!/bin/bash
# This script runs on the Ubuntu VM to set it up for ravact development

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Setting up Ubuntu VM for Ravact${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Update system
echo -e "${BLUE}Updating system packages...${NC}"
sudo apt update && sudo apt upgrade -y

# Install essential tools
echo -e "${BLUE}Installing essential tools...${NC}"
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
    unzip

# Install Go 1.21 for ARM64
echo -e "${BLUE}Installing Go 1.21...${NC}"
GO_VERSION="1.21.6"
cd /tmp
wget -q https://go.dev/dl/go${GO_VERSION}.linux-arm64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-arm64.tar.gz

# Setup Go environment
if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo 'export GOPATH=$HOME/go' >> ~/.bashrc
    echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
fi
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go

# Verify Go installation
echo ""
echo -e "${GREEN}✓ Go installed:${NC}"
/usr/local/go/bin/go version

# Create project directory
mkdir -p ~/ravact-go

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}VM Setup Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "VM is ready for ravact development"
echo ""
EOFSCRIPT
)

# Copy and run setup script on VM
print_info "Setting up VM (this will take a few minutes)..."
echo "$VM_SETUP_SCRIPT" | ssh ${VM_USER}@${VM_IP} "cat > /tmp/setup-vm.sh && chmod +x /tmp/setup-vm.sh && bash /tmp/setup-vm.sh"

# Build ravact for Linux ARM64
print_info "Building ravact for Linux ARM64..."
cd "$(dirname "$0")/.."
make build-linux-arm64
print_success "Build complete"

# Copy ravact to VM
print_info "Deploying ravact to VM..."
scp dist/ravact-linux-arm64 ${VM_USER}@${VM_IP}:~/ravact-go/ravact
print_success "Ravact deployed"

# Copy assets to VM
print_info "Copying assets..."
ssh ${VM_USER}@${VM_IP} "mkdir -p ~/ravact-go/assets/scripts ~/ravact-go/assets/configs"
scp -r assets/scripts/* ${VM_USER}@${VM_IP}:~/ravact-go/assets/scripts/
scp -r assets/configs/* ${VM_USER}@${VM_IP}:~/ravact-go/assets/configs/
print_success "Assets copied"

# Make ravact executable
ssh ${VM_USER}@${VM_IP} "chmod +x ~/ravact-go/ravact"

# Create helper script on VM
print_info "Creating helper script on VM..."
ssh ${VM_USER}@${VM_IP} "cat > ~/ravact-go/test-ravact.sh << 'EOF'
#!/bin/bash
# Quick test script for ravact
cd ~/ravact-go
echo 'Testing ravact...'
./ravact --version 2>/dev/null || echo 'Run with: sudo ./ravact'
echo ''
echo 'To run ravact:'
echo '  cd ~/ravact-go'
echo '  sudo ./ravact'
EOF
chmod +x ~/ravact-go/test-ravact.sh"

# Create sync script on Mac
print_info "Creating sync script for future updates..."
SYNC_SCRIPT="$HOME/.ravact-sync.sh"
cat > "$SYNC_SCRIPT" << EOF
#!/bin/bash
# Quick sync script to update ravact on VM after building

set -e

VM_USER="${VM_USER}"
VM_IP="${VM_IP}"
PROJECT_DIR="$(pwd)"

echo "Building ravact for Linux ARM64..."
cd "$PROJECT_DIR"
make build-linux-arm64

echo "Deploying to VM..."
scp dist/ravact-linux-arm64 \${VM_USER}@\${VM_IP}:~/ravact-go/ravact
scp -r assets/scripts/* \${VM_USER}@\${VM_IP}:~/ravact-go/assets/scripts/
scp -r assets/configs/* \${VM_USER}@\${VM_IP}:~/ravact-go/assets/configs/

echo "✓ Ravact updated on VM"
echo ""
echo "To test:"
echo "  ssh ${VM_NAME}"
echo "  cd ravact-go"
echo "  sudo ./ravact"
EOF
chmod +x "$SYNC_SCRIPT"
print_success "Sync script created: $SYNC_SCRIPT"

# Summary
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Setup Complete! 🎉${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${GREEN}VM Details:${NC}"
echo "  Name: ${VM_NAME}"
echo "  IP: ${VM_IP}"
echo "  User: ${VM_USER}"
echo "  Pass: ${VM_PASS}"
echo ""
echo -e "${GREEN}Quick Access:${NC}"
echo "  SSH to VM:    ssh ${VM_NAME}"
echo "  Run ravact:   ssh ${VM_NAME} 'cd ravact-go && sudo ./ravact'"
echo ""
echo -e "${GREEN}Development Workflow:${NC}"
echo "  1. Make code changes on your Mac"
echo "  2. Run: $SYNC_SCRIPT"
echo "  3. SSH to VM and test: ssh ${VM_NAME}"
echo ""
echo -e "${GREEN}VS Code Remote SSH Setup:${NC}"
echo "  1. Install 'Remote - SSH' extension"
echo "  2. Press F1 → 'Remote-SSH: Connect to Host'"
echo "  3. Select '${VM_NAME}'"
echo "  4. Open folder: /home/${VM_USER}/ravact-go"
echo ""
echo -e "${BLUE}Testing ravact now...${NC}"
ssh ${VM_NAME} "cd ravact-go && ./ravact --version 2>/dev/null || echo 'Ravact is ready! Run with: sudo ./ravact'"
echo ""
echo -e "${GREEN}Ready to develop!${NC}"
echo ""
