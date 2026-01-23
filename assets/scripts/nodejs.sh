#!/bin/bash
#
# Node.js Installation Script for Ravact
# Installs NVM (Node Version Manager) and Node.js with common tools
#

set -e  # Exit on error

echo "=========================================="
echo "  Node.js Installation with NVM"
echo "=========================================="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root"
    exit 1
fi

# Detect distribution
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
    VERSION=$VERSION_ID
else
    echo "Error: Cannot detect OS distribution"
    exit 1
fi

echo "Detected OS: $OS $VERSION"
echo ""

# Determine installation method
echo "Select Node.js installation method:"
echo ""
echo "  1. NVM (Node Version Manager) - Recommended"
echo "  2. Direct Installation (via NodeSource)"
echo ""
read -p "Enter choice [1-2] (default: 1): " install_method
install_method=${install_method:-1}

# Set Node.js version (can be overridden via environment)
NODE_VERSION="${NODE_VERSION:-20}"  # LTS version

if [ "$install_method" = "1" ]; then
    # Install with NVM
    echo ""
    echo "Installing NVM (Node Version Manager)..."
    
    # Update package list and install dependencies
    case "$OS" in
        ubuntu|debian)
            apt-get update -qq
            apt-get install -y curl wget git
            ;;
        centos|rhel|fedora)
            yum update -y -q
            yum install -y curl wget git
            ;;
    esac
    
    # Install NVM
    NVM_VERSION="v0.39.7"
    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/${NVM_VERSION}/install.sh | bash
    
    # Load NVM
    export NVM_DIR="$HOME/.nvm"
    [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
    
    echo "✓ NVM installed"
    
    # Install Node.js via NVM
    echo ""
    echo "Installing Node.js ${NODE_VERSION} LTS via NVM..."
    nvm install ${NODE_VERSION}
    nvm use ${NODE_VERSION}
    nvm alias default ${NODE_VERSION}
    
    echo "✓ Node.js ${NODE_VERSION} installed via NVM"
    
else
    # Direct installation
    echo ""
    echo "Installing Node.js ${NODE_VERSION}.x directly..."
    
    # Update package list
    case "$OS" in
        ubuntu|debian)
            apt-get update -qq
            ;;
        centos|rhel|fedora)
            yum update -y -q
            ;;
    esac
    
    # Install Node.js
    case "$OS" in
        ubuntu|debian)
            # Add NodeSource repository
            curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash -
            apt-get install -y nodejs
            ;;
        centos|rhel|fedora)
            # Add NodeSource repository
            curl -fsSL https://rpm.nodesource.com/setup_${NODE_VERSION}.x | bash -
            yum install -y nodejs
            ;;
        *)
            echo "Error: Unsupported distribution"
            exit 1
            ;;
    esac
fi

# Verify installation
if command -v node &> /dev/null && command -v npm &> /dev/null; then
    echo ""
    echo "✓ Node.js installed successfully!"
    
    # Get versions
    NODE_VERSION_FULL=$(node --version)
    NPM_VERSION=$(npm --version)
    
    echo "✓ Node.js version: $NODE_VERSION_FULL"
    echo "✓ npm version: $NPM_VERSION"
    
    # Install common global packages
    echo ""
    echo "Installing common global packages..."
    
    npm install -g pm2 --silent
    echo "✓ PM2 (Process Manager) installed"
    
    npm install -g yarn --silent
    echo "✓ Yarn (Package Manager) installed"
    
    # Configure npm
    echo ""
    echo "Configuring npm..."
    
    # Create global node_modules directory that doesn't require sudo
    mkdir -p /usr/local/lib/node_modules
    npm config set prefix /usr/local
    
    echo "✓ npm configured"
    
    # Set up PM2 to start on boot
    echo ""
    echo "Configuring PM2 startup..."
    
    # Generate startup script
    PM2_STARTUP=$(pm2 startup systemd -u root --hp /root | grep "sudo")
    if [ ! -z "$PM2_STARTUP" ]; then
        eval "$PM2_STARTUP" 2>/dev/null || true
        echo "✓ PM2 configured to start on boot"
    fi
    
    echo ""
    echo "=========================================="
    echo "  Installation Complete!"
    echo "=========================================="
    echo ""
    if [ "$install_method" = "1" ]; then
        echo "Installation method: NVM"
        echo ""
        echo "NVM commands:"
        echo "  nvm install 18        # Install Node.js 18"
        echo "  nvm install 20        # Install Node.js 20"
        echo "  nvm use 20            # Use Node.js 20"
        echo "  nvm alias default 20  # Set default version"
        echo "  nvm ls                # List installed versions"
        echo "  nvm ls-remote         # List available versions"
        echo ""
        echo "To use NVM in new shells, add to your shell profile:"
        echo "  echo 'export NVM_DIR=\"\$HOME/.nvm\"' >> ~/.bashrc"
        echo "  echo '[ -s \"\$NVM_DIR/nvm.sh\" ] && \\. \"\$NVM_DIR/nvm.sh\"' >> ~/.bashrc"
        echo ""
    else
        echo "Installation method: Direct (NodeSource)"
        echo ""
    fi
    
    echo "Node.js tools installed:"
    echo "  • node - JavaScript runtime"
    echo "  • npm - Package manager"
    echo "  • yarn - Alternative package manager"
    echo "  • pm2 - Process manager"
    echo ""
    echo "Verify installation:"
    echo "  node --version"
    echo "  npm --version"
    echo "  yarn --version"
    echo "  pm2 --version"
    echo ""
    echo "Next steps:"
    echo "  • Create your Node.js application"
    echo "  • Use PM2 to manage processes: pm2 start app.js"
    echo "  • Configure reverse proxy with Nginx"
    echo "  • Set up environment variables"
    echo ""
    echo "PM2 commands:"
    echo "  pm2 start app.js --name myapp"
    echo "  pm2 list"
    echo "  pm2 logs"
    echo "  pm2 restart myapp"
    echo "  pm2 save"
    echo ""
else
    echo ""
    echo "✗ Error: Node.js installation failed"
    exit 1
fi

exit 0
