#!/bin/bash
#
# Interactive PHP Installation Script for Ravact
# Allows version and extension selection with Laravel defaults
#

set -e  # Exit on error

echo "=========================================="
echo "  Interactive PHP Installation"
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

# Function to show menu
show_menu() {
    local title="$1"
    shift
    local options=("$@")
    
    echo "$title"
    for i in "${!options[@]}"; do
        echo "  $((i+1)). ${options[$i]}"
    done
    echo ""
}

# Select PHP version
echo "Select PHP version to install:"
echo ""
echo "  1. PHP 7.4 (Legacy - Security updates only)"
echo "  2. PHP 8.0 (Legacy)"
echo "  3. PHP 8.1 (Active support)"
echo "  4. PHP 8.2 (Active support) [RECOMMENDED]"
echo "  5. PHP 8.3 (Latest)"
echo ""
read -p "Enter choice [1-5] (default: 4): " version_choice
version_choice=${version_choice:-4}

case $version_choice in
    1) PHP_VERSION="7.4" ;;
    2) PHP_VERSION="8.0" ;;
    3) PHP_VERSION="8.1" ;;
    4) PHP_VERSION="8.2" ;;
    5) PHP_VERSION="8.3" ;;
    *) echo "Invalid choice. Using 8.2"; PHP_VERSION="8.2" ;;
esac

echo ""
echo "Selected: PHP $PHP_VERSION"
echo ""

# Extension selection mode
echo "Select extension configuration:"
echo ""
echo "  1. Laravel Defaults (Recommended for Laravel apps)"
echo "  2. Advanced Selection (Choose individual extensions)"
echo ""
read -p "Enter choice [1-2] (default: 1): " ext_mode
ext_mode=${ext_mode:-1}

# Define Laravel default extensions
LARAVEL_EXTENSIONS=(
    "cli"
    "fpm"
    "common"
    "mysql"
    "pgsql"
    "sqlite3"
    "redis"
    "curl"
    "gd"
    "mbstring"
    "xml"
    "zip"
    "bcmath"
    "intl"
    "opcache"
)

# All available extensions organized by category
declare -A ALL_EXTENSIONS
ALL_EXTENSIONS=(
    ["Database Drivers"]="mysql mysqli pgsql sqlite3 pdo mongodb redis memcached"
    ["Web Development"]="curl gd imagick fileinfo exif gettext"
    ["Data Processing"]="json xml simplexml dom xmlreader xmlwriter yaml igbinary msgpack"
    ["Security"]="openssl sodium hash"
    ["Performance"]="opcache apcu xdebug"
    ["Utilities"]="zip bz2 iconv mbstring intl bcmath gmp"
    ["Advanced"]="swoole grpc rdkafka amqp ssh2 ldap imap soap"
)

SELECTED_EXTENSIONS=()

if [ "$ext_mode" = "1" ]; then
    echo ""
    echo "Using Laravel default extensions:"
    for ext in "${LARAVEL_EXTENSIONS[@]}"; do
        echo "  ✓ $ext"
        SELECTED_EXTENSIONS+=("$ext")
    done
else
    echo ""
    echo "=========================================="
    echo "  Advanced Extension Selection"
    echo "=========================================="
    echo ""
    echo "Press ENTER to toggle selection, type 'done' when finished"
    echo "Type 'search <term>' to filter extensions"
    echo ""
    
    # Start with Laravel defaults selected
    for ext in "${LARAVEL_EXTENSIONS[@]}"; do
        SELECTED_EXTENSIONS+=("$ext")
    done
    
    # Show categories
    for category in "Database Drivers" "Web Development" "Data Processing" "Security" "Performance" "Utilities" "Advanced"; do
        echo ""
        echo "=== $category ==="
        
        extensions=(${ALL_EXTENSIONS[$category]})
        for ext in "${extensions[@]}"; do
            # Check if selected
            selected="  "
            for sel in "${SELECTED_EXTENSIONS[@]}"; do
                if [ "$sel" = "$ext" ]; then
                    selected="✓ "
                    break
                fi
            done
            
            echo "${selected}$ext"
        done
    done
    
    echo ""
    echo "Default Laravel extensions are pre-selected."
    read -p "Proceed with these extensions? [Y/n]: " confirm
    confirm=${confirm:-y}
    
    if [[ ! $confirm =~ ^[Yy]$ ]]; then
        echo "Extension selection cancelled. Using defaults."
    fi
fi

echo ""
echo "=========================================="
echo "  Installation Summary"
echo "=========================================="
echo ""
echo "PHP Version: $PHP_VERSION"
echo "Extensions to install: ${#SELECTED_EXTENSIONS[@]}"
echo ""
read -p "Proceed with installation? [Y/n]: " proceed
proceed=${proceed:-y}

if [[ ! $proceed =~ ^[Yy]$ ]]; then
    echo "Installation cancelled."
    exit 0
fi

# Start installation
echo ""
echo "Starting installation..."
echo ""

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

# Build package list
PACKAGES=()
for ext in "${SELECTED_EXTENSIONS[@]}"; do
    case "$OS" in
        ubuntu|debian)
            # Handle special cases
            if [ "$ext" = "mysqli" ]; then
                PACKAGES+=("php${PHP_VERSION}-mysql")
            elif [ "$ext" = "pdo" ]; then
                PACKAGES+=("php${PHP_VERSION}-mysql")
            else
                PACKAGES+=("php${PHP_VERSION}-${ext}")
            fi
            ;;
        centos|rhel|fedora)
            if [ "$ext" = "mysqli" ]; then
                PACKAGES+=("php-mysqlnd")
            elif [ "$ext" = "pdo" ]; then
                PACKAGES+=("php-pdo")
            else
                PACKAGES+=("php-${ext}")
            fi
            ;;
    esac
done

# Remove duplicates
PACKAGES=($(echo "${PACKAGES[@]}" | tr ' ' '\n' | sort -u | tr '\n' ' '))

# Install packages
echo "Installing PHP $PHP_VERSION with selected extensions..."
case "$OS" in
    ubuntu|debian)
        apt-get install -y "${PACKAGES[@]}"
        ;;
    centos|rhel|fedora)
        yum install -y "${PACKAGES[@]}"
        ;;
esac

# Enable and start PHP-FPM
echo "Enabling and starting PHP-FPM..."
systemctl enable php${PHP_VERSION}-fpm 2>/dev/null || systemctl enable php-fpm 2>/dev/null || true
systemctl start php${PHP_VERSION}-fpm 2>/dev/null || systemctl start php-fpm 2>/dev/null || true

# Wait for PHP-FPM to be ready
sleep 2

# Verify installation
if systemctl is-active --quiet php${PHP_VERSION}-fpm || systemctl is-active --quiet php-fpm; then
    echo ""
    echo "✓ PHP installed and running successfully!"
    
    PHP_VERSION_FULL=$(php -v | head -1 | grep -oP 'PHP \K[0-9.]+')
    echo "✓ PHP version: $PHP_VERSION_FULL"
    
    echo ""
    echo "Installed extensions:"
    php -m | grep -v "^\[" | sed 's/^/  • /'
    
    # Configure PHP
    echo ""
    echo "Applying recommended PHP configuration..."
    
    PHP_INI="/etc/php/${PHP_VERSION}/fpm/php.ini"
    [ -f "/etc/php.ini" ] && PHP_INI="/etc/php.ini"
    
    if [ -f "$PHP_INI" ]; then
        cp "$PHP_INI" "$PHP_INI.backup"
        
        sed -i 's/^memory_limit.*/memory_limit = 256M/' "$PHP_INI"
        sed -i 's/^upload_max_filesize.*/upload_max_filesize = 20M/' "$PHP_INI"
        sed -i 's/^post_max_size.*/post_max_size = 25M/' "$PHP_INI"
        sed -i 's/^max_execution_time.*/max_execution_time = 60/' "$PHP_INI"
        
        echo "✓ Applied recommended settings"
        
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
            echo "✓ Composer installed"
        fi
    else
        echo "✓ Composer already installed"
    fi
    
    echo ""
    echo "=========================================="
    echo "  Installation Complete!"
    echo "=========================================="
    echo ""
    
else
    echo ""
    echo "✗ Error: PHP installation failed"
    exit 1
fi

exit 0
