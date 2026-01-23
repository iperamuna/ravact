#!/bin/bash
#
# PostgreSQL Installation Script for Ravact
# Installs PostgreSQL Server
#

set -e  # Exit on error

echo "=========================================="
echo "  PostgreSQL Installation"
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

# Install PostgreSQL
echo "Installing PostgreSQL..."
case "$OS" in
    ubuntu|debian)
        apt-get install -y postgresql postgresql-contrib
        ;;
    centos|rhel|fedora)
        yum install -y postgresql-server postgresql-contrib
        # Initialize database on RHEL-based systems
        postgresql-setup --initdb 2>/dev/null || true
        ;;
    *)
        echo "Error: Unsupported distribution"
        exit 1
        ;;
esac

# Enable and start PostgreSQL
echo "Enabling and starting PostgreSQL service..."
systemctl enable postgresql
systemctl start postgresql

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
sleep 3

# Check if PostgreSQL is running
if systemctl is-active --quiet postgresql; then
    echo ""
    echo "✓ PostgreSQL installed and running successfully!"
    
    # Get version
    PG_VERSION=$(sudo -u postgres psql --version | grep -oP '\d+\.\d+' | head -1)
    echo "✓ PostgreSQL version: $PG_VERSION"
    
    # Create default database and user if specified
    if [ ! -z "$PG_DATABASE" ] && [ ! -z "$PG_USER" ] && [ ! -z "$PG_PASSWORD" ]; then
        echo ""
        echo "Creating database and user..."
        sudo -u postgres psql <<EOF
CREATE USER $PG_USER WITH PASSWORD '$PG_PASSWORD';
CREATE DATABASE $PG_DATABASE OWNER $PG_USER;
GRANT ALL PRIVILEGES ON DATABASE $PG_DATABASE TO $PG_USER;
EOF
        echo "✓ Database '$PG_DATABASE' created"
        echo "✓ User '$PG_USER' created"
    fi
    
    # Configure firewall if available
    if command -v ufw &> /dev/null; then
        echo "Configuring firewall (UFW)..."
        echo "Note: PostgreSQL port 5432 not exposed - use SSH tunnel for remote access"
    elif command -v firewall-cmd &> /dev/null; then
        echo "Configuring firewall (firewalld)..."
        echo "Note: PostgreSQL port 5432 not exposed - use SSH tunnel for remote access"
    fi
    
    echo ""
    echo "=========================================="
    echo "  Installation Complete!"
    echo "=========================================="
    echo ""
    echo "PostgreSQL installed successfully!"
    echo ""
    echo "Connect as postgres user:"
    echo "  sudo -u postgres psql"
    echo ""
    echo "Next steps:"
    echo "  • Create application databases and users"
    echo "  • Configure postgresql.conf for optimization"
    echo "  • Set up regular backups"
    echo "  • Configure pg_hba.conf for access control"
    echo ""
else
    echo ""
    echo "✗ Error: PostgreSQL installation failed or service not running"
    exit 1
fi

exit 0
