#!/bin/bash
#
# Supervisor Installation Script for Ravact
# Installs Supervisor process control system
#

set -e  # Exit on error

echo "=========================================="
echo "  Supervisor Installation"
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

# Install Supervisor
echo "Installing Supervisor..."
case "$OS" in
    ubuntu|debian)
        apt-get install -y supervisor
        ;;
    centos|rhel|fedora)
        yum install -y supervisor
        ;;
    *)
        echo "Error: Unsupported distribution"
        exit 1
        ;;
esac

# Enable and start Supervisor
echo "Enabling and starting Supervisor service..."
systemctl enable supervisor 2>/dev/null || systemctl enable supervisord 2>/dev/null || true
systemctl start supervisor 2>/dev/null || systemctl start supervisord 2>/dev/null || true

# Wait for Supervisor to be ready
echo "Waiting for Supervisor to be ready..."
sleep 2

# Check if Supervisor is running
if systemctl is-active --quiet supervisor || systemctl is-active --quiet supervisord; then
    echo ""
    echo "✓ Supervisor installed and running successfully!"
    
    # Get version
    SUPERVISOR_VERSION=$(supervisord --version)
    echo "✓ Supervisor version: $SUPERVISOR_VERSION"
    
    # Create example configuration
    echo ""
    echo "Creating example configuration..."
    
    CONF_DIR="/etc/supervisor/conf.d"
    mkdir -p "$CONF_DIR"
    
    cat > "$CONF_DIR/example-laravel-queue.conf" << 'EOF'
; Example: Laravel Queue Worker
; Copy and modify this file for your application
; File: /etc/supervisor/conf.d/laravel-queue.conf

[program:laravel-queue]
process_name=%(program_name)s_%(process_num)02d
command=php /var/www/html/artisan queue:work --sleep=3 --tries=3 --max-time=3600
autostart=true
autorestart=true
stopasgroup=true
killasgroup=true
user=www-data
numprocs=2
redirect_stderr=true
stdout_logfile=/var/www/html/storage/logs/queue.log
stopwaitsecs=3600
EOF
    
    echo "✓ Created example configuration: $CONF_DIR/example-laravel-queue.conf"
    
    # Reload configuration
    supervisorctl reread 2>/dev/null || true
    supervisorctl update 2>/dev/null || true
    
    echo ""
    echo "=========================================="
    echo "  Installation Complete!"
    echo "=========================================="
    echo ""
    echo "Supervisor installed successfully!"
    echo ""
    echo "Configuration directory:"
    echo "  $CONF_DIR"
    echo ""
    echo "Supervisor commands:"
    echo "  supervisorctl status           # View all programs"
    echo "  supervisorctl start <program>  # Start a program"
    echo "  supervisorctl stop <program>   # Stop a program"
    echo "  supervisorctl restart <program># Restart a program"
    echo "  supervisorctl reread           # Reload config files"
    echo "  supervisorctl update           # Apply changes"
    echo ""
    echo "Next steps:"
    echo "  • Create program config files in $CONF_DIR"
    echo "  • Run 'supervisorctl reread && supervisorctl update'"
    echo "  • Check status with 'supervisorctl status'"
    echo ""
    echo "Example program config created:"
    echo "  $CONF_DIR/example-laravel-queue.conf"
    echo ""
else
    echo ""
    echo "✗ Error: Supervisor installation failed or service not running"
    exit 1
fi

exit 0
