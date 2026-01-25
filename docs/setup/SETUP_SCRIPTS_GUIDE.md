# Setup Scripts Guide

[← Back to Documentation](../README.md)

## Overview

Ravact includes 10 professional installation scripts with advanced features like interactive configuration, version selection, and extension management.

## Available Scripts

### Web Servers
- **nginx.sh** - Nginx with automatic firewall configuration

### Databases
- **mysql.sh** - MySQL with secure setup and password configuration
- **postgresql.sh** - PostgreSQL with database/user creation
- **redis.sh** - Redis cache with optimized settings

### Runtimes
- **php.sh** (Interactive) - PHP with version and extension selection
- **nodejs.sh** (Interactive) - Node.js with NVM option

### Process Management
- **supervisor.sh** - Supervisor with example configurations

### SSL/Security
- **certbot.sh** - Let's Encrypt with auto-renewal

### Development
- **git.sh** - Git with SSH key generation

### System
- **firewall.sh** - UFW/firewalld configuration

## Interactive Features

### PHP Installation (php.sh)

#### Version Selection
Choose from PHP versions 7.4 to 8.3:
1. PHP 7.4 (Legacy - Security updates only)
2. PHP 8.0 (Legacy)
3. PHP 8.1 (Active support)
4. PHP 8.2 (Active support) **[RECOMMENDED]**
5. PHP 8.3 (Latest)

#### Extension Selection

**Option 1: Laravel Defaults (Recommended)**
Installs essential extensions for Laravel applications:
- CLI, FPM, Common
- MySQL, PostgreSQL, SQLite3
- Redis
- cURL, GD, Mbstring, XML, Zip
- BCMath, Intl, OPcache

**Option 2: Advanced Selection**
Choose from categorized extensions:

**Database Drivers:**
- mysql, mysqli, pgsql, sqlite3, pdo
- mongodb, redis, memcached

**Web Development:**
- curl, gd, imagick, fileinfo, exif, gettext

**Data Processing:**
- json, xml, simplexml, dom, xmlreader, xmlwriter
- yaml, igbinary, msgpack

**Security & Encryption:**
- openssl, sodium, hash

**Performance:**
- opcache, apcu, xdebug

**Utilities:**
- zip, bz2, iconv, mbstring, intl, bcmath, gmp

**Advanced:**
- swoole, grpc, rdkafka, amqp, ssh2
- ldap, imap, soap

#### Usage Example

```bash
# Interactive installation (prompts for version and extensions)
sudo ./assets/scripts/php.sh

# Pre-configured installation
sudo PHP_VERSION=8.2 ./assets/scripts/php.sh
```

### Node.js Installation (nodejs.sh)

#### Installation Methods

**Option 1: NVM (Recommended)**
- Install Node Version Manager
- Allows multiple Node.js versions
- Easy version switching
- Per-user installation

**Option 2: Direct Installation**
- System-wide installation via NodeSource
- Single version
- Managed by package manager

#### NVM Features

```bash
# Install specific versions
nvm install 18
nvm install 20

# Switch versions
nvm use 20
nvm use 18

# Set default version
nvm alias default 20

# List versions
nvm ls                # Installed versions
nvm ls-remote         # Available versions
```

#### Usage Example

```bash
# Interactive installation (prompts for method)
sudo ./assets/scripts/nodejs.sh

# Pre-configured NVM installation
sudo NODE_VERSION=20 ./assets/scripts/nodejs.sh

# Pre-configured direct installation (via environment)
sudo INSTALL_METHOD=2 NODE_VERSION=20 ./assets/scripts/nodejs.sh
```

## PHP Extensions Configuration

The PHP extensions are defined in `assets/configs/php-extensions.json`:

```json
{
  "php_versions": ["7.4", "8.0", "8.1", "8.2", "8.3"],
  "default_version": "8.2",
  "extension_categories": {
    "laravel_default": { ... },
    "database": { ... },
    "web": { ... },
    ...
  },
  "extension_info": {
    "mysql": {
      "description": "MySQL database driver",
      "laravel": true
    },
    ...
  }
}
```

### Modifying Available Extensions

To add or remove extensions:

1. Edit `assets/configs/php-extensions.json`
2. Add extension to appropriate category
3. Add extension info with description
4. Rebuild if using TUI integration

Example - Adding a new extension:

```json
{
  "extension_categories": {
    "advanced": {
      "extensions": [
        ...,
        "my-new-extension"
      ]
    }
  },
  "extension_info": {
    "my-new-extension": {
      "description": "My custom extension",
      "laravel": false
    }
  }
}
```

## Environment Variables

All scripts support environment variables for automation:

### MySQL
```bash
MYSQL_ROOT_PASSWORD="SecurePass123"
MYSQL_DATABASE="myapp_db"
```

### PostgreSQL
```bash
PG_DATABASE="myapp_db"
PG_USER="myapp_user"
PG_PASSWORD="SecurePass123"
```

### PHP
```bash
PHP_VERSION="8.2"
```

### Node.js
```bash
NODE_VERSION="20"
INSTALL_METHOD="1"  # 1=NVM, 2=Direct
```

### Git
```bash
GIT_USER_NAME="John Doe"
GIT_USER_EMAIL="john@example.com"
DEPLOY_USER="www-data"
```

### Firewall
```bash
SSH_PORT="22"
EXTRA_PORTS="8080 8443"
RESET_FIREWALL="yes"
```

## Installation Flow

### Via Ravact TUI

1. Run Ravact: `./ravact`
2. Navigate to **Setup** menu
3. Select application
4. Press **Enter** to view actions
5. Choose **Install**
6. Follow interactive prompts (if applicable)

### Manual Installation

```bash
# Basic installation
sudo bash assets/scripts/nginx.sh

# With environment variables
sudo MYSQL_ROOT_PASSWORD="test123" bash assets/scripts/mysql.sh

# Interactive scripts
sudo bash assets/scripts/php.sh
# (Follow prompts)
```

## Testing Scripts

### In Docker (Ubuntu 24.04)

```bash
# Enter Docker shell
make docker-shell

# Test a script
cd /workspace
bash assets/scripts/nginx.sh

# Verify service
systemctl status nginx
```

### On VM/Server

```bash
# Copy script to server
scp assets/scripts/nginx.sh user@server:/tmp/

# Run on server
ssh user@server
sudo bash /tmp/nginx.sh
```

## Script Features

All scripts include:
- ✅ Distribution detection (Ubuntu, Debian, CentOS, RHEL, Fedora)
- ✅ Root privilege check
- ✅ Error handling (`set -e`)
- ✅ Service verification
- ✅ Version detection
- ✅ Automatic service startup
- ✅ Basic configuration
- ✅ Environment variable support
- ✅ Clear output and next steps

## Troubleshooting

### Script Fails to Run

**Check permissions:**
```bash
chmod +x assets/scripts/*.sh
```

**Check shebang:**
All scripts should start with `#!/bin/bash`

### Interactive Input Not Working

**Run in interactive shell:**
```bash
bash -i assets/scripts/php.sh
```

**Pipe input:**
```bash
echo -e "4\n1\ny" | bash assets/scripts/php.sh
# (Version 4 = PHP 8.2, Mode 1 = Laravel defaults, Confirm = Yes)
```

### Service Not Starting

**Check logs:**
```bash
journalctl -u nginx -n 50
journalctl -u php8.2-fpm -n 50
```

**Manually start:**
```bash
systemctl start nginx
systemctl status nginx
```

### Extension Not Found (PHP)

Some extensions may not be available for all PHP versions or distributions.

**Check availability:**
```bash
apt-cache search php8.2-
```

**Alternative package name:**
Some distributions use different names (e.g., `php-mysqlnd` instead of `php-mysql`)

## Best Practices

1. **Test First**: Always test scripts in a safe environment (Docker, VM) before production
2. **Use Environment Variables**: Automate installations with environment variables
3. **Check Prerequisites**: Ensure system requirements are met
4. **Backup Data**: Backup existing configurations before running scripts
5. **Review Output**: Check script output for errors or warnings
6. **Verify Services**: Confirm services are running after installation
7. **Security**: Change default passwords immediately
8. **Updates**: Keep scripts updated with latest versions and security patches

## Future Enhancements

Planned improvements:
- [ ] Dragonfly database installation
- [ ] FrankenPHP installation (Classic, Worker, Caddy modes)
- [ ] Configuration templates for all services
- [ ] Health check and monitoring integration
- [ ] Multi-service setup profiles
- [ ] Rollback capability
- [ ] Dry-run mode
- [ ] Installation logs

## Contributing

To add a new setup script:

1. Create script in `assets/scripts/`
2. Follow naming convention: `servicename.sh`
3. Include standard features (error handling, checks, etc.)
4. Add to `setup_menu.go` service list
5. Add to `installed_apps.go` service list
6. Document in this guide
7. Test on supported distributions

---

**Last Updated**: January 23, 2026
