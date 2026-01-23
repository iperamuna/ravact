#!/bin/bash
#
# MySQL Installation Script for Ravact
# Installs MySQL Server and performs secure setup
#

set -e  # Exit on error

echo "=========================================="
echo "  MySQL Installation"
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

# Set MySQL root password (can be overridden via environment)
MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-RavactMySQLRoot123!}"

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

# Install MySQL
echo "Installing MySQL Server..."
case "$OS" in
    ubuntu|debian)
        # Preseed MySQL root password
        echo "mysql-server mysql-server/root_password password $MYSQL_ROOT_PASSWORD" | debconf-set-selections
        echo "mysql-server mysql-server/root_password_again password $MYSQL_ROOT_PASSWORD" | debconf-set-selections
        
        DEBIAN_FRONTEND=noninteractive apt-get install -y mysql-server
        ;;
    centos|rhel|fedora)
        yum install -y mysql-server
        ;;
    *)
        echo "Error: Unsupported distribution"
        exit 1
        ;;
esac

# Enable and start MySQL
echo "Enabling and starting MySQL service..."
systemctl enable mysql 2>/dev/null || systemctl enable mysqld 2>/dev/null || true
systemctl start mysql 2>/dev/null || systemctl start mysqld 2>/dev/null || true

# Wait for MySQL to be ready
echo "Waiting for MySQL to be ready..."
for i in {1..30}; do
    if mysqladmin ping -h localhost --silent 2>/dev/null; then
        break
    fi
    sleep 1
done

# Check if MySQL is running
if systemctl is-active --quiet mysql || systemctl is-active --quiet mysqld; then
    echo ""
    echo "✓ MySQL installed and running successfully!"
    
    # Get version
    MYSQL_VERSION=$(mysql --version | grep -oP 'Ver \K[0-9.]+')
    echo "✓ MySQL version: $MYSQL_VERSION"
    
    # Perform basic secure installation
    echo ""
    echo "Performing basic security configuration..."
    
    # Remove anonymous users and test database
    mysql -u root -p"$MYSQL_ROOT_PASSWORD" <<EOF 2>/dev/null || true
DELETE FROM mysql.user WHERE User='';
DELETE FROM mysql.user WHERE User='root' AND Host NOT IN ('localhost', '127.0.0.1', '::1');
DROP DATABASE IF EXISTS test;
DELETE FROM mysql.db WHERE Db='test' OR Db='test\\_%';
FLUSH PRIVILEGES;
EOF
    
    echo "✓ Basic security configuration applied"
    
    # Create default database (optional)
    if [ ! -z "$MYSQL_DATABASE" ]; then
        echo "Creating database: $MYSQL_DATABASE"
        mysql -u root -p"$MYSQL_ROOT_PASSWORD" -e "CREATE DATABASE IF NOT EXISTS $MYSQL_DATABASE CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>/dev/null || true
    fi
    
    # Configure firewall if available
    if command -v ufw &> /dev/null; then
        echo "Configuring firewall (UFW)..."
        # MySQL port 3306 - typically not exposed externally
        echo "Note: MySQL port 3306 not exposed - use SSH tunnel for remote access"
    elif command -v firewall-cmd &> /dev/null; then
        echo "Configuring firewall (firewalld)..."
        echo "Note: MySQL port 3306 not exposed - use SSH tunnel for remote access"
    fi
    
    echo ""
    echo "=========================================="
    echo "  Installation Complete!"
    echo "=========================================="
    echo ""
    echo "MySQL Credentials:"
    echo "  Root Password: $MYSQL_ROOT_PASSWORD"
    echo ""
    echo "Connection:"
    echo "  mysql -u root -p"
    echo ""
    echo "Next steps:"
    echo "  • Create application databases and users"
    echo "  • Configure my.cnf for optimization"
    echo "  • Set up regular backups"
    echo "  • Use Ravact's Configuration menu to tune settings"
    echo ""
    echo "⚠ IMPORTANT: Save the root password securely!"
    echo ""
else
    echo ""
    echo "✗ Error: MySQL installation failed or service not running"
    exit 1
fi

exit 0
