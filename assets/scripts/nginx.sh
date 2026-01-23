#!/bin/bash
#
# Nginx Installation Script for Ravact
# This script installs and configures Nginx web server
#

set -e  # Exit on error

echo "=========================================="
echo "  Nginx Installation"
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
    *)
        echo "Warning: Unsupported distribution. Attempting to continue..."
        ;;
esac

# Install Nginx
echo "Installing Nginx..."
case "$OS" in
    ubuntu|debian)
        apt-get install -y nginx
        ;;
    centos|rhel|fedora)
        yum install -y nginx
        ;;
    *)
        echo "Error: Cannot install Nginx on this distribution"
        exit 1
        ;;
esac

# Enable and start Nginx
echo "Enabling and starting Nginx service..."
systemctl enable nginx
systemctl start nginx

# Check if Nginx is running
if systemctl is-active --quiet nginx; then
    echo ""
    echo "✓ Nginx installed and running successfully!"
    
    # Get version
    NGINX_VERSION=$(nginx -v 2>&1 | grep -oP 'nginx/\K[0-9.]+')
    echo "✓ Nginx version: $NGINX_VERSION"
    
    # Configure firewall if available
    if command -v ufw &> /dev/null; then
        echo "Configuring firewall (UFW)..."
        ufw allow 'Nginx Full' > /dev/null 2>&1 || true
    elif command -v firewall-cmd &> /dev/null; then
        echo "Configuring firewall (firewalld)..."
        firewall-cmd --permanent --add-service=http > /dev/null 2>&1 || true
        firewall-cmd --permanent --add-service=https > /dev/null 2>&1 || true
        firewall-cmd --reload > /dev/null 2>&1 || true
    fi
    
    echo ""
    echo "=========================================="
    echo "  Installation Complete!"
    echo "=========================================="
    echo ""
    echo "Next steps:"
    echo "  • Visit http://localhost to see the default page"
    echo "  • Configure sites in /etc/nginx/sites-available/"
    echo "  • Use Ravact's Configuration menu to edit settings"
    echo ""
else
    echo ""
    echo "✗ Error: Nginx installation failed or service not running"
    exit 1
fi

exit 0
