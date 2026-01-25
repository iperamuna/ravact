#!/bin/bash
# Setup script for Multipass - Easier alternative to UTM
# For M1 MacBook Pro

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
VM_NAME="ravact-dev"
VM_USER="ubuntu"
VM_CPUS="4"
VM_MEM="4G"
VM_DISK="20G"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Ravact VM Setup with Multipass${NC}"
echo -e "${BLUE}M1 Mac - Quick & Easy${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

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
print_info "Checking system..."
ARCH=$(uname -m)
if [[ "$ARCH" != "arm64" ]]; then
    print_warning "Expected Apple Silicon (M1/M2), got: $ARCH"
    print_info "Script will continue but may not work optimally"
fi
print_success "Running on macOS $ARCH"
echo ""

# Check if Multipass is installed
print_info "Checking for Multipass..."

# Check common Multipass binary locations
MULTIPASS_BIN=""
if [[ -f "/usr/local/bin/multipass" ]]; then
    MULTIPASS_BIN="/usr/local/bin/multipass"
elif [[ -f "//Library/Application Support/com.canonical.multipass/bin/multipass" ]]; then
    MULTIPASS_BIN="//Library/Application Support/com.canonical.multipass/bin/multipass"
fi

# Add to PATH if found
if [[ -n "$MULTIPASS_BIN" ]]; then
    export PATH="$(dirname "$MULTIPASS_BIN"):$PATH"
fi

if ! command -v multipass &> /dev/null; then
    # Check if app is installed but binary not in PATH
    if [[ -d "/Applications/Multipass.app" ]]; then
        print_warning "Multipass app is installed but binary not accessible"
        print_info "Creating symlink..."
        
        # Try to create symlink (will prompt for password)
        if sudo ln -sf "//Library/Application Support/com.canonical.multipass/bin/multipass" /usr/local/bin/multipass 2>/dev/null; then
            export PATH="/usr/local/bin:$PATH"
            print_success "Symlink created"
        else
            print_warning "Could not create symlink automatically"
            print_info "Please run manually:"
            print_info "  sudo ln -sf '//Library/Application Support/com.canonical.multipass/bin/multipass' /usr/local/bin/multipass"
            print_info "Then run this script again"
            exit 1
        fi
    else
        print_warning "Multipass not found. Installing..."
        if command -v brew &> /dev/null; then
            brew install --cask multipass
            
            # Wait for installation
            sleep 3
            
            # Create symlink
            print_info "Setting up Multipass..."
            sudo ln -sf "//Library/Application Support/com.canonical.multipass/bin/multipass" /usr/local/bin/multipass 2>/dev/null || true
            export PATH="/usr/local/bin:$PATH"
            
            if ! command -v multipass &> /dev/null; then
                print_error "Multipass installed but not accessible"
                print_info "Please restart your terminal or run:"
                print_info "  sudo ln -sf '//Library/Application Support/com.canonical.multipass/bin/multipass' /usr/local/bin/multipass"
                print_info "  export PATH=\"/usr/local/bin:\$PATH\""
                exit 1
            fi
            
            print_success "Multipass installed"
        else
            print_error "Homebrew not found. Please install Multipass manually:"
            print_info "Visit: https://multipass.run/install"
            exit 1
        fi
    fi
else
    print_success "Multipass is installed"
    multipass version
fi
echo ""

# Check if VM already exists
print_info "Checking for existing VM..."
if multipass list | grep -q "$VM_NAME"; then
    print_warning "VM '$VM_NAME' already exists"
    echo ""
    read -p "Delete and recreate? (y/N): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Deleting existing VM..."
        multipass delete $VM_NAME
        multipass purge
        print_success "VM deleted"
    else
        print_info "Using existing VM"
    fi
else
    print_success "No existing VM found"
fi
echo ""

# Create VM if it doesn't exist
if ! multipass list | grep -q "$VM_NAME"; then
    print_info "Creating Ubuntu VM..."
    print_info "Configuration: ${VM_CPUS} CPUs, ${VM_MEM} RAM, ${VM_DISK} disk"
    
    multipass launch \
        --name $VM_NAME \
        --cpus $VM_CPUS \
        --memory $VM_MEM \
        --disk $VM_DISK \
        24.04
    
    print_success "VM created successfully"
else
    print_success "VM already running"
fi
echo ""

# Get VM IP
print_info "Getting VM information..."
VM_IP=$(multipass info $VM_NAME | grep IPv4 | awk '{print $2}')
print_success "VM IP: $VM_IP"
echo ""

# Display VM info
echo -e "${GREEN}VM Details:${NC}"
multipass info $VM_NAME
echo ""

# Wait for VM to be ready
print_info "Waiting for VM to be fully ready..."
sleep 5
multipass exec $VM_NAME -- cloud-init status --wait
print_success "VM is ready"
echo ""

# Update SSH config
print_info "Configuring SSH access..."
SSH_CONFIG="$HOME/.ssh/config"
mkdir -p "$HOME/.ssh"

if ! grep -q "Host $VM_NAME" "$SSH_CONFIG" 2>/dev/null; then
    cat >> "$SSH_CONFIG" << EOF

# Ravact Development VM (Multipass)
Host ${VM_NAME}
    HostName ${VM_IP}
    User ${VM_USER}
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
EOF
    print_success "SSH config updated"
else
    print_info "SSH config already exists for $VM_NAME"
fi
echo ""

# Copy and run setup script on VM
print_info "Setting up VM environment..."

VM_SETUP_SCRIPT=$(cat << 'EOFSCRIPT'
#!/bin/bash
set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Setting up Ubuntu VM for Ravact...${NC}"
echo ""

# Update system
echo -e "${BLUE}Updating packages...${NC}"
sudo apt update
sudo DEBIAN_FRONTEND=noninteractive apt upgrade -y

# Install essentials
echo -e "${BLUE}Installing development tools...${NC}"
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
    tmux

# Install Go
echo -e "${BLUE}Installing Go 1.24...${NC}"
GO_VERSION="1.24.0"
cd /tmp
wget -q https://go.dev/dl/go${GO_VERSION}.linux-arm64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-arm64.tar.gz

# Setup Go environment
if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
    cat >> ~/.bashrc << 'EOF'

# Go environment
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
EOF
fi

export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go

# Create directories
mkdir -p ~/ravact-go/assets/scripts
mkdir -p ~/ravact-go/assets/configs

# Create helper scripts
cat > ~/ravact-go/test-ravact.sh << 'EOF'
#!/bin/bash
cd ~/ravact-go
if [[ ! -f "./ravact" ]]; then
    echo "Error: ravact binary not found"
    exit 1
fi
echo "Testing ravact..."
./ravact --version 2>/dev/null || echo "Run with: sudo ./ravact"
EOF
chmod +x ~/ravact-go/test-ravact.sh

echo ""
echo -e "${GREEN}✓ VM setup complete${NC}"
echo ""
/usr/local/go/bin/go version
EOFSCRIPT
)

# Write and execute setup script
echo "$VM_SETUP_SCRIPT" | multipass exec $VM_NAME -- bash

print_success "VM environment configured"
echo ""

# Build ravact
print_info "Building ravact for Linux ARM64..."
cd "$(dirname "$0")/.."
make build-linux-arm64
print_success "Build complete"
echo ""

# Deploy ravact
print_info "Deploying ravact to VM..."
multipass transfer dist/ravact-linux-arm64 ${VM_NAME}:/home/ubuntu/ravact-go/ravact
multipass exec $VM_NAME -- chmod +x /home/ubuntu/ravact-go/ravact
print_success "Binary deployed"
echo ""

# Deploy assets
print_info "Copying assets..."
if [[ -d "assets/scripts" ]]; then
    for script in assets/scripts/*; do
        multipass transfer "$script" ${VM_NAME}:/home/ubuntu/ravact-go/assets/scripts/
    done
fi
if [[ -d "assets/configs" ]]; then
    for config in assets/configs/*; do
        multipass transfer "$config" ${VM_NAME}:/home/ubuntu/ravact-go/assets/configs/
    done
fi
print_success "Assets deployed"
echo ""

# Create sync script
print_info "Creating sync script..."
SYNC_SCRIPT="$HOME/.ravact-multipass-sync.sh"
cat > "$SYNC_SCRIPT" << EOF
#!/bin/bash
# Quick sync for Multipass VM

set -e

PROJECT_DIR="$(pwd)"

echo "Building ravact..."
cd "\$PROJECT_DIR"
make build-linux-arm64

echo "Deploying to VM..."
multipass transfer dist/ravact-linux-arm64 ${VM_NAME}:/home/ubuntu/ravact-go/ravact
multipass exec ${VM_NAME} -- chmod +x /home/ubuntu/ravact-go/ravact

# Copy assets
for script in assets/scripts/*; do
    [[ -f "\$script" ]] && multipass transfer "\$script" ${VM_NAME}:/home/ubuntu/ravact-go/assets/scripts/
done

for config in assets/configs/*; do
    [[ -f "\$config" ]] && multipass transfer "\$config" ${VM_NAME}:/home/ubuntu/ravact-go/assets/configs/
done

echo "✓ Ravact updated on VM"
echo ""
echo "To test:"
echo "  multipass shell ${VM_NAME}"
echo "  cd ravact-go && sudo ./ravact"
EOF
chmod +x "$SYNC_SCRIPT"
print_success "Sync script created: $SYNC_SCRIPT"
echo ""

# Test ravact
print_info "Testing ravact on VM..."
multipass exec $VM_NAME -- /home/ubuntu/ravact-go/ravact --version 2>/dev/null || \
    echo "Ravact installed (needs sudo to run)"
echo ""

# Summary
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Setup Complete! 🎉${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${GREEN}VM Details:${NC}"
echo "  Name: $VM_NAME"
echo "  IP: $VM_IP"
echo "  User: $VM_USER"
echo "  CPUs: $VM_CPUS"
echo "  RAM: $VM_MEM"
echo "  Disk: $VM_DISK"
echo ""
echo -e "${GREEN}Quick Commands:${NC}"
echo ""
echo "  Shell access:"
echo "    multipass shell $VM_NAME"
echo "    ssh $VM_NAME"
echo ""
echo "  Run ravact:"
echo "    multipass exec $VM_NAME -- sudo /home/ubuntu/ravact-go/ravact"
echo ""
echo "  VM management:"
echo "    multipass stop $VM_NAME      # Stop VM"
echo "    multipass start $VM_NAME     # Start VM"
echo "    multipass restart $VM_NAME   # Restart VM"
echo "    multipass delete $VM_NAME    # Delete VM"
echo ""
echo "  Deploy updates:"
echo "    $SYNC_SCRIPT"
echo ""
echo -e "${GREEN}Development Workflow:${NC}"
echo "  1. Make code changes on Mac"
echo "  2. Run: $SYNC_SCRIPT"
echo "  3. Test: multipass shell $VM_NAME"
echo "     cd ravact-go && sudo ./ravact"
echo ""
echo -e "${GREEN}VS Code Remote SSH:${NC}"
echo "  1. Install 'Remote - SSH' extension"
echo "  2. Connect to: $VM_NAME"
echo "  3. Open: /home/ubuntu/ravact-go"
echo ""
echo -e "${BLUE}Starting shell in VM...${NC}"
echo ""
multipass shell $VM_NAME
