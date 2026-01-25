# Developer Toolkit

[← Back to Documentation](../README.md)

The Developer Toolkit provides quick access to frequently forgotten but essential terminal commands for Laravel, WordPress, PHP, and security maintenance.

## Overview

Access via: **Main Menu → Site Management → Developer Toolkit**

The toolkit contains **34+ commands** organized into 4 categories:
- Laravel Commands
- WordPress Commands
- PHP Commands
- Security Commands

## Navigation

| Key | Action |
|-----|--------|
| `↑`/`↓` or `j`/`k` | Navigate commands |
| `Tab` or `→`/`←` | Switch categories |
| `c` | Copy command to clipboard |
| `Enter` | Execute command |
| `Esc` | Go back |

## Categories

### Laravel Commands

| Command | Description |
|---------|-------------|
| Tail Laravel Log | Watch log file in real-time (`tail -f storage/logs/laravel.log`) |
| Clear Laravel Log | Truncate log to zero bytes |
| Find Large Log Files | Find logs larger than 100MB |
| Fix Storage Permissions | Set correct permissions for storage & bootstrap/cache |
| Generate APP_KEY | Generate a new base64 APP_KEY |
| Check .env File | Display environment config (hides sensitive values) |
| List Scheduled Tasks | Show crontab entries for Laravel scheduler |
| Check Queue Workers | List running queue worker processes |
| Find Recently Modified Files | Find files modified in last 24 hours |

### WordPress Commands

| Command | Description |
|---------|-------------|
| Fix wp-content Permissions | Set correct permissions (755 dirs, 644 files) |
| Find Large Uploads | Find uploaded files larger than 10MB |
| Clear Cache Files | Remove all files from wp-content/cache |
| Generate WP Salts | Fetch fresh security salts from WordPress API |
| Check wp-config.php | Display database and debug settings |
| List Plugins | List all installed plugins |
| List Themes | List all installed themes |
| Check .htaccess | Display .htaccess contents |
| Find Modified Core Files | Find core files modified in last 7 days |

### PHP Commands

| Command | Description |
|---------|-------------|
| Check PHP Version | Display installed PHP version |
| List PHP Modules | List all installed PHP modules |
| Check PHP Memory Limit | Display memory_limit setting |
| Check PHP Upload Limits | Display upload and post size limits |
| Find php.ini Location | Show loaded configuration file path |
| Check OPcache Status | Display OPcache configuration |
| Test PHP Syntax | Check PHP files for syntax errors |
| List PHP-FPM Pools | Show PHP-FPM pool configurations |

### Security Commands

| Command | Description |
|---------|-------------|
| Scan for Malware Patterns | Search for common malware signatures |
| Find World-Writable Files | List files with 777 permissions |
| Find World-Writable Dirs | List directories with 777 permissions |
| Check for Suspicious Files | Find PHP files in upload directories |
| List Failed SSH Logins | Show recent failed authentication attempts |
| Check Open Ports | List all listening ports and services |
| Check SSL Certificate | Display SSL certificate expiry |
| Find SUID Files | List files with SUID bit set |

## Usage Tips

1. **Copy Before Execute**: Press `c` to copy a command first, then modify it for your specific needs

2. **Path-Specific Commands**: Some commands need to be run from specific directories (e.g., Laravel project root). The description indicates when this is required.

3. **Root Privileges**: Security commands often require sudo. Run Ravact with `sudo` for full functionality.

4. **Safe by Default**: Commands that modify files will ask for confirmation before executing.

## Example Workflow

### Debugging a Laravel Application

1. Navigate to Developer Toolkit
2. Go to Laravel category
3. Execute "Check .env File" to verify configuration
4. Execute "Check Queue Workers" to see if workers are running
5. Execute "Tail Laravel Log" to watch for errors
6. If logs are large, use "Clear Laravel Log" to reset

### Security Audit

1. Go to Security category
2. Run "Scan for Malware Patterns" to check for infections
3. Run "Find World-Writable Files" to identify permission issues
4. Run "Check Open Ports" to verify exposed services
5. Run "List Failed SSH Logins" to check for intrusion attempts
