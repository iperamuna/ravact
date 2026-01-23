#!/bin/bash
#
# Firewall Configuration Script for Ravact
# Configures UFW/firewalld with common server rules
#

set -e  # Exit on error

echo "=========================================="
echo "  Firewall Configuration"
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

# Determine which firewall to use
FIREWALL=""
if command -v ufw &> /dev/null; then
    FIREWALL="ufw"
elif command -v firewall-cmd &> /dev/null; then
    FIREWALL="firewalld"
else
    echo "Installing firewall..."
    case "$OS" in
        ubuntu|debian)
            apt-get update -qq
            apt-get install -y ufw
            FIREWALL="ufw"
            ;;
        centos|rhel|fedora)
            yum update -y -q
            yum install -y firewalld
            systemctl enable firewalld
            systemctl start firewalld
            FIREWALL="firewalld"
            ;;
        *)
            echo "Error: Cannot determine firewall for this distribution"
            exit 1
            ;;
    esac
fi

echo "Using firewall: $FIREWALL"
echo ""

# Configure firewall based on type
if [ "$FIREWALL" = "ufw" ]; then
    echo "Configuring UFW firewall..."
    
    # Reset to default if needed
    if [ "${RESET_FIREWALL}" = "yes" ]; then
        ufw --force reset
        echo "✓ Reset firewall to defaults"
    fi
    
    # Set default policies
    ufw default deny incoming
    ufw default allow outgoing
    echo "✓ Set default policies"
    
    # Allow SSH (IMPORTANT - do this first!)
    SSH_PORT="${SSH_PORT:-22}"
    ufw allow $SSH_PORT/tcp comment 'SSH'
    echo "✓ Allowed SSH on port $SSH_PORT"
    
    # Allow HTTP and HTTPS
    ufw allow 80/tcp comment 'HTTP'
    ufw allow 443/tcp comment 'HTTPS'
    echo "✓ Allowed HTTP (80) and HTTPS (443)"
    
    # Allow additional ports if specified
    if [ ! -z "$EXTRA_PORTS" ]; then
        for PORT in $EXTRA_PORTS; do
            ufw allow $PORT comment 'Custom port'
            echo "✓ Allowed custom port: $PORT"
        done
    fi
    
    # Enable firewall
    ufw --force enable
    
    echo ""
    echo "✓ UFW firewall configured and enabled!"
    echo ""
    echo "Current rules:"
    ufw status numbered
    
elif [ "$FIREWALL" = "firewalld" ]; then
    echo "Configuring firewalld..."
    
    # Ensure firewalld is running
    systemctl start firewalld
    
    # Set default zone
    firewall-cmd --set-default-zone=public
    
    # Allow SSH (IMPORTANT!)
    firewall-cmd --permanent --add-service=ssh
    echo "✓ Allowed SSH"
    
    # Allow HTTP and HTTPS
    firewall-cmd --permanent --add-service=http
    firewall-cmd --permanent --add-service=https
    echo "✓ Allowed HTTP and HTTPS"
    
    # Allow additional ports if specified
    if [ ! -z "$EXTRA_PORTS" ]; then
        for PORT in $EXTRA_PORTS; do
            firewall-cmd --permanent --add-port=$PORT/tcp
            echo "✓ Allowed custom port: $PORT"
        done
    fi
    
    # Reload to apply changes
    firewall-cmd --reload
    
    echo ""
    echo "✓ firewalld configured!"
    echo ""
    echo "Current rules:"
    firewall-cmd --list-all
fi

echo ""
echo "=========================================="
echo "  Configuration Complete!"
echo "=========================================="
echo ""
echo "Firewall is now active and configured!"
echo ""
echo "Common commands:"
if [ "$FIREWALL" = "ufw" ]; then
    echo "  ufw status              # Check status"
    echo "  ufw allow 8080/tcp      # Allow port"
    echo "  ufw deny 8080/tcp       # Deny port"
    echo "  ufw delete allow 8080   # Remove rule"
    echo "  ufw disable             # Disable firewall"
    echo "  ufw enable              # Enable firewall"
else
    echo "  firewall-cmd --list-all                    # Check rules"
    echo "  firewall-cmd --add-port=8080/tcp          # Allow port (temporary)"
    echo "  firewall-cmd --permanent --add-port=8080/tcp  # Allow port (permanent)"
    echo "  firewall-cmd --remove-port=8080/tcp       # Remove port"
    echo "  firewall-cmd --reload                     # Apply changes"
fi
echo ""
echo "⚠ IMPORTANT:"
echo "  • SSH port ($SSH_PORT) is allowed - DO NOT lock yourself out!"
echo "  • Always test rules before logging out"
echo "  • Keep a backup console access method"
echo ""
echo "Allowed services:"
echo "  • SSH (port $SSH_PORT)"
echo "  • HTTP (port 80)"
echo "  • HTTPS (port 443)"
if [ ! -z "$EXTRA_PORTS" ]; then
    echo "  • Custom ports: $EXTRA_PORTS"
fi
echo ""

exit 0
