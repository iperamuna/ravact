#!/bin/bash
#
# Git Installation Script for Ravact
# Installs Git and sets up SSH keys for repository access
#

set -e  # Exit on error

echo "=========================================="
echo "  Git Installation & Setup"
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

# Update package list
echo "Updating package list..."
case "$OS" in
    ubuntu|debian)
        apt-get update -qq
        ;;
    centos|rhel|fedora)
        yum update -y -q
        ;;
esac

# Install Git
echo "Installing Git..."
case "$OS" in
    ubuntu|debian)
        apt-get install -y git
        ;;
    centos|rhel|fedora)
        yum install -y git
        ;;
    *)
        echo "Error: Unsupported distribution"
        exit 1
        ;;
esac

# Verify installation
if command -v git &> /dev/null; then
    echo ""
    echo "✓ Git installed successfully!"
    
    # Get version
    GIT_VERSION=$(git --version | grep -oP 'git version \K[0-9.]+')
    echo "✓ Git version: $GIT_VERSION"
    
    # Configure Git (if user info provided)
    echo ""
    echo "Configuring Git..."
    
    if [ ! -z "$GIT_USER_NAME" ]; then
        git config --global user.name "$GIT_USER_NAME"
        echo "✓ Set user.name: $GIT_USER_NAME"
    fi
    
    if [ ! -z "$GIT_USER_EMAIL" ]; then
        git config --global user.email "$GIT_USER_EMAIL"
        echo "✓ Set user.email: $GIT_USER_EMAIL"
    fi
    
    # Set default branch name
    git config --global init.defaultBranch main
    echo "✓ Set default branch: main"
    
    # Set up SSH key for deployment user
    echo ""
    echo "Setting up SSH keys..."
    
    # Create deployment user if specified
    DEPLOY_USER="${DEPLOY_USER:-www-data}"
    
    if id "$DEPLOY_USER" &>/dev/null; then
        USER_HOME=$(eval echo ~$DEPLOY_USER)
        SSH_DIR="$USER_HOME/.ssh"
        
        # Create .ssh directory
        mkdir -p "$SSH_DIR"
        chmod 700 "$SSH_DIR"
        
        # Generate SSH key if it doesn't exist
        if [ ! -f "$SSH_DIR/id_ed25519" ]; then
            ssh-keygen -t ed25519 -f "$SSH_DIR/id_ed25519" -N "" -C "$DEPLOY_USER@$(hostname)"
            echo "✓ Generated SSH key for $DEPLOY_USER"
            
            # Display public key
            echo ""
            echo "=========================================="
            echo "  SSH Public Key (Add to GitHub/GitLab)"
            echo "=========================================="
            echo ""
            cat "$SSH_DIR/id_ed25519.pub"
            echo ""
            echo "=========================================="
            echo ""
        else
            echo "✓ SSH key already exists for $DEPLOY_USER"
        fi
        
        # Set correct permissions
        chown -R $DEPLOY_USER:$DEPLOY_USER "$SSH_DIR"
        chmod 600 "$SSH_DIR/id_ed25519" 2>/dev/null || true
        chmod 644 "$SSH_DIR/id_ed25519.pub" 2>/dev/null || true
        
        # Create known_hosts with common Git hosts
        KNOWN_HOSTS="$SSH_DIR/known_hosts"
        touch "$KNOWN_HOSTS"
        
        echo "Adding common Git hosts to known_hosts..."
        ssh-keyscan github.com >> "$KNOWN_HOSTS" 2>/dev/null || true
        ssh-keyscan gitlab.com >> "$KNOWN_HOSTS" 2>/dev/null || true
        ssh-keyscan bitbucket.org >> "$KNOWN_HOSTS" 2>/dev/null || true
        
        chown $DEPLOY_USER:$DEPLOY_USER "$KNOWN_HOSTS"
        chmod 644 "$KNOWN_HOSTS"
        
        echo "✓ Added GitHub, GitLab, and Bitbucket to known_hosts"
    fi
    
    echo ""
    echo "=========================================="
    echo "  Installation Complete!"
    echo "=========================================="
    echo ""
    echo "Git installed successfully!"
    echo ""
    echo "Verify installation:"
    echo "  git --version"
    echo "  git config --list"
    echo ""
    if [ ! -z "$DEPLOY_USER" ] && [ -f "$SSH_DIR/id_ed25519.pub" ]; then
        echo "SSH key location:"
        echo "  Private: $SSH_DIR/id_ed25519"
        echo "  Public: $SSH_DIR/id_ed25519.pub"
        echo ""
        echo "View public key:"
        echo "  cat $SSH_DIR/id_ed25519.pub"
        echo ""
        echo "Add the public key to:"
        echo "  • GitHub: Settings → SSH and GPG keys"
        echo "  • GitLab: Preferences → SSH Keys"
        echo "  • Bitbucket: Personal settings → SSH keys"
        echo ""
    fi
    echo "Clone a repository:"
    echo "  git clone git@github.com:user/repo.git"
    echo ""
    echo "Next steps:"
    echo "  • Add SSH public key to your Git provider"
    echo "  • Test connection: ssh -T git@github.com"
    echo "  • Clone your repositories"
    echo "  • Set up deployment workflows"
    echo ""
else
    echo ""
    echo "✗ Error: Git installation failed"
    exit 1
fi

exit 0
