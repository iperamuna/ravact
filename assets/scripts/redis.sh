#!/bin/bash
#
# Redis Installation Script for Ravact
# Installs Redis Server for caching and message queues
#

set -e  # Exit on error

echo "=========================================="
echo "  Redis Installation"
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

# Install Redis
echo "Installing Redis Server..."
case "$OS" in
    ubuntu|debian)
        apt-get install -y redis-server
        ;;
    centos|rhel|fedora)
        yum install -y redis
        ;;
    *)
        echo "Error: Unsupported distribution"
        exit 1
        ;;
esac

# Enable and start Redis
echo "Enabling and starting Redis service..."
systemctl enable redis-server 2>/dev/null || systemctl enable redis 2>/dev/null || true
systemctl start redis-server 2>/dev/null || systemctl start redis 2>/dev/null || true

# Wait for Redis to be ready
echo "Waiting for Redis to be ready..."
sleep 2

# Check if Redis is running
if systemctl is-active --quiet redis-server || systemctl is-active --quiet redis; then
    echo ""
    echo "✓ Redis installed and running successfully!"
    
    # Get version
    REDIS_VERSION=$(redis-cli --version | grep -oP 'redis-cli \K[0-9.]+')
    echo "✓ Redis version: $REDIS_VERSION"
    
    # Test connection
    if redis-cli ping | grep -q PONG; then
        echo "✓ Redis is responding to commands"
    fi
    
    # Basic configuration
    echo ""
    echo "Applying basic configuration..."
    
    # Set maxmemory policy if not set
    REDIS_CONF="/etc/redis/redis.conf"
    [ -f "/etc/redis.conf" ] && REDIS_CONF="/etc/redis.conf"
    
    if [ -f "$REDIS_CONF" ]; then
        # Backup original config
        cp "$REDIS_CONF" "$REDIS_CONF.backup"
        
        # Set maxmemory policy to allkeys-lru if not set
        if ! grep -q "^maxmemory-policy" "$REDIS_CONF"; then
            echo "maxmemory-policy allkeys-lru" >> "$REDIS_CONF"
            echo "✓ Set maxmemory-policy to allkeys-lru"
        fi
        
        # Set maxmemory to 256MB if not set
        if ! grep -q "^maxmemory" "$REDIS_CONF"; then
            echo "maxmemory 268435456" >> "$REDIS_CONF"
            echo "✓ Set maxmemory to 256MB"
        fi
        
        # Restart to apply changes
        systemctl restart redis-server 2>/dev/null || systemctl restart redis 2>/dev/null || true
        sleep 1
    fi
    
    echo ""
    echo "=========================================="
    echo "  Installation Complete!"
    echo "=========================================="
    echo ""
    echo "Redis installed successfully!"
    echo ""
    echo "Connect to Redis:"
    echo "  redis-cli"
    echo ""
    echo "Test connection:"
    echo "  redis-cli ping"
    echo ""
    echo "Configuration:"
    echo "  Config file: $REDIS_CONF"
    echo "  Default port: 6379"
    echo "  Listening on: localhost (127.0.0.1)"
    echo ""
    echo "Next steps:"
    echo "  • Configure password authentication"
    echo "  • Adjust maxmemory settings for your needs"
    echo "  • Set up persistence if required"
    echo "  • Use Ravact's Configuration menu to tune settings"
    echo ""
else
    echo ""
    echo "✗ Error: Redis installation failed or service not running"
    exit 1
fi

exit 0
