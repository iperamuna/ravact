#!/bin/bash
#
# PHP Installation Script for Ravact
# Installs PHP-FPM with common extensions
#

set -e  # Exit on error

echo "=========================================="
echo "  PHP Installation"
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

# Set PHP version (can be overridden via environment)
PHP_VERSION="${PHP_VERSION:-8.2}"

# Update package list
echo "Updating package list..."
case "$OS" in
    ubuntu|debian)
        apt-get update -qq
        
        # Add Ondrej PHP repository for latest versions
        if [ "$OS" = "ubuntu" ]; then
            apt-get install -y software-properties-common
            add-apt-repository -y ppa:ondrej/php
            apt-get update -qq
        fi
        ;;
    centos|rhel|fedora)
        yum update -y -q
        ;;
esac

# Install PHP
echo "Installing PHP $PHP_VERSION with common extensions..."
case "$OS" in
    ubuntu|debian)
        apt-get install -y \
            php${PHP_VERSION}-fpm \
            php${PHP_VERSION}-cli \
            php${PHP_VERSION}-common \
            php${PHP_VERSION}-mysql \
            php${PHP_VERSION}-pgsql \
            php${PHP_VERSION}-sqlite3 \
            php${PHP_VERSION}-redis \
            php${PHP_VERSION}-curl \
            php${PHP_VERSION}-gd \
            php${PHP_VERSION}-mbstring \
            php${PHP_VERSION}-xml \
            php${PHP_VERSION}-zip \
            php${PHP_VERSION}-bcmath \
            php${PHP_VERSION}-intl \
            php${PHP_VERSION}-opcache
        ;;
    centos|rhel|fedora)
        yum install -y \
            php-fpm \
            php-cli \
            php-mysqlnd \
            php-pgsql \
            php-redis \
            php-gd \
            php-mbstring \
            php-xml \
            php-zip \
            php-bcmath \
            php-intl \
            php-opcache
        ;;
    *)
        echo "Error: Unsupported distribution"
        exit 1
        ;;
esac

# Enable and start PHP-FPM
echo "Enabling and starting PHP-FPM service..."
systemctl enable php${PHP_VERSION}-fpm 2>/dev/null || systemctl enable php-fpm 2>/dev/null || true
systemctl start php${PHP_VERSION}-fpm 2>/dev/null || systemctl start php-fpm 2>/dev/null || true

# Wait for PHP-FPM to be ready
echo "Waiting for PHP-FPM to be ready..."
sleep 2

# Check if PHP-FPM is running
if systemctl is-active --quiet php${PHP_VERSION}-fpm || systemctl is-active --quiet php-fpm; then
    echo ""
    echo "✓ PHP installed and running successfully!"
    
    # Get version
    PHP_VERSION_FULL=$(php -v | head -1 | grep -oP 'PHP \K[0-9.]+')
    echo "✓ PHP version: $PHP_VERSION_FULL"
    
    # List installed extensions
    echo ""
    echo "Installed extensions:"
    php -m | grep -E "(mysql|pgsql|redis|curl|gd|mbstring|xml|zip|bcmath|intl|opcache)" | sed 's/^/  • /'
    
    # Configure PHP
    echo ""
    echo "Applying recommended PHP configuration..."
    
    PHP_INI="/etc/php/${PHP_VERSION}/fpm/php.ini"
    [ -f "/etc/php.ini" ] && PHP_INI="/etc/php.ini"
    
    if [ -f "$PHP_INI" ]; then
        # Backup original config
        cp "$PHP_INI" "$PHP_INI.backup"
        
        # Apply recommended settings
        sed -i 's/^memory_limit.*/memory_limit = 256M/' "$PHP_INI"
        sed -i 's/^upload_max_filesize.*/upload_max_filesize = 20M/' "$PHP_INI"
        sed -i 's/^post_max_size.*/post_max_size = 25M/' "$PHP_INI"
        sed -i 's/^max_execution_time.*/max_execution_time = 60/' "$PHP_INI"
        
        echo "✓ Applied recommended PHP settings"
        echo "  • memory_limit = 256M"
        echo "  • upload_max_filesize = 20M"
        echo "  • post_max_size = 25M"
        echo "  • max_execution_time = 60s"
        
        # Restart to apply changes
        systemctl restart php${PHP_VERSION}-fpm 2>/dev/null || systemctl restart php-fpm 2>/dev/null || true
    fi
    
    # Install Composer
    echo ""
    echo "Installing Composer..."
    if ! command -v composer &> /dev/null; then
        EXPECTED_CHECKSUM="$(php -r 'copy("https://composer.github.io/installer.sig", "php://stdout");')"
        php -r "copy('https://getcomposer.org/installer', 'composer-setup.php');"
        ACTUAL_CHECKSUM="$(php -r "echo hash_file('sha384', 'composer-setup.php');")"
        
        if [ "$EXPECTED_CHECKSUM" != "$ACTUAL_CHECKSUM" ]; then
            echo "✗ Composer installer corrupt"
            rm composer-setup.php
        else
            php composer-setup.php --quiet --install-dir=/usr/local/bin --filename=composer
            rm composer-setup.php
            echo "✓ Composer installed successfully"
            composer --version
        fi
    else
        echo "✓ Composer already installed"
        composer --version
    fi
    
    echo ""
    echo "=========================================="
    echo "  Installation Complete!"
    echo "=========================================="
    echo ""
    echo "PHP-FPM socket:"
    echo "  /run/php/php${PHP_VERSION}-fpm.sock"
    echo ""
    echo "Configuration files:"
    echo "  PHP INI: $PHP_INI"
    echo "  FPM Pool: /etc/php/${PHP_VERSION}/fpm/pool.d/www.conf"
    echo ""
    echo "Check PHP info:"
    echo "  php -i"
    echo ""
    echo "Next steps:"
    echo "  • Configure Nginx/Apache to use PHP-FPM"
    echo "  • Adjust PHP settings for your application"
    echo "  • Install additional extensions if needed"
    echo "  • Set up OPcache for production"
    echo ""
else
    echo ""
    echo "✗ Error: PHP installation failed or service not running"
    exit 1
fi

exit 0
