#!/bin/bash
#
# Certbot Installation Script for Ravact
# Installs Certbot for Let's Encrypt SSL certificates
#

set -e  # Exit on error

echo "=========================================="
echo "  Certbot Installation"
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

# Install Certbot
echo "Installing Certbot..."
case "$OS" in
    ubuntu|debian)
        apt-get install -y certbot python3-certbot-nginx
        ;;
    centos|rhel|fedora)
        yum install -y certbot python3-certbot-nginx
        ;;
    *)
        echo "Error: Unsupported distribution"
        exit 1
        ;;
esac

# Verify installation
if command -v certbot &> /dev/null; then
    echo ""
    echo "✓ Certbot installed successfully!"
    
    # Get version
    CERTBOT_VERSION=$(certbot --version 2>&1 | grep -oP 'certbot \K[0-9.]+')
    echo "✓ Certbot version: $CERTBOT_VERSION"
    
    # Set up automatic renewal
    echo ""
    echo "Setting up automatic certificate renewal..."
    
    # Create renewal timer (systemd)
    systemctl enable certbot.timer 2>/dev/null || true
    systemctl start certbot.timer 2>/dev/null || true
    
    # Or add cron job if systemd timer not available
    if ! systemctl is-active --quiet certbot.timer; then
        CRON_CMD="0 0,12 * * * root certbot renew --quiet --deploy-hook 'systemctl reload nginx'"
        if ! grep -q "certbot renew" /etc/crontab; then
            echo "$CRON_CMD" >> /etc/crontab
            echo "✓ Added cron job for automatic renewal"
        fi
    else
        echo "✓ Systemd timer configured for automatic renewal"
    fi
    
    echo ""
    echo "=========================================="
    echo "  Installation Complete!"
    echo "=========================================="
    echo ""
    echo "Certbot installed successfully!"
    echo ""
    echo "Obtain a certificate:"
    echo "  certbot --nginx -d example.com -d www.example.com"
    echo ""
    echo "Or for manual certificate:"
    echo "  certbot certonly --standalone -d example.com"
    echo ""
    echo "Test renewal:"
    echo "  certbot renew --dry-run"
    echo ""
    echo "List certificates:"
    echo "  certbot certificates"
    echo ""
    echo "Revoke and delete:"
    echo "  certbot revoke --cert-path /etc/letsencrypt/live/example.com/cert.pem"
    echo "  certbot delete --cert-name example.com"
    echo ""
    echo "Next steps:"
    echo "  • Ensure your domain points to this server"
    echo "  • Make sure ports 80 and 443 are open"
    echo "  • Run certbot with your domain name"
    echo "  • Certificates will auto-renew every 60 days"
    echo ""
    echo "⚠ Important:"
    echo "  • Domain must be publicly accessible"
    echo "  • DNS must be configured correctly"
    echo "  • Firewall must allow ports 80 and 443"
    echo ""
else
    echo ""
    echo "✗ Error: Certbot installation failed"
    exit 1
fi

exit 0
