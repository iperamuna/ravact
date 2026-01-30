# FrankenPHP Multi-Site Setup Guide

[← Back to Documentation](../README.md)

## Overview

FrankenPHP is a modern PHP application server built on top of the Caddy web server. It combines the power of PHP with Caddy's automatic HTTPS, making it perfect for modern PHP applications.

## Three Operation Modes

### 1. Classic Mode (Traditional)
Traditional PHP execution similar to PHP-FPM. Each request spawns a new PHP process.

**Best for:**
- Legacy applications
- Simple PHP sites
- Low-traffic sites
- Development

**Configuration:**
```caddy
example.com {
    root * /var/www/example
    php_server
    file_server
}
```

### 2. Worker Mode (High Performance)
PHP workers stay in memory, dramatically improving performance. Perfect for Laravel and Symfony applications.

**Best for:**
- Laravel applications
- Symfony applications
- High-traffic sites
- Production environments

**Performance:**
- 10-20x faster than Classic mode
- Lower memory usage per request
- Reduced CPU usage
- Keep application bootstrapped in memory

**Configuration:**
```caddy
example.com {
    root * /var/www/example/public
    
    php_server {
        workers 4              # Number of worker processes
        worker_script index.php # Bootstrap file
    }
    
    # Laravel routing
    @notStatic {
        not path *.css *.js *.png *.jpg *.gif *.svg *.ico
        file {
            try_files {path} /index.php
        }
    }
    rewrite @notStatic /index.php
    
    file_server
}
```

### 3. Mercure Mode (Real-Time)
Adds real-time push capabilities using the Mercure protocol. Perfect for applications needing live updates.

**Best for:**
- Real-time dashboards
- Chat applications
- Live notifications
- Collaborative tools
- Any app needing Server-Sent Events (SSE)

**Features:**
- Real-time updates
- Server-Sent Events (SSE)
- Pub/Sub messaging
- Authorization via JWT
- Works with any HTTP client

**Configuration:**
```caddy
example.com {
    root * /var/www/example/public
    
    mercure {
        publisher_jwt {
            key "!ChangeThisMercureHubJWTSecretKey!"
            algorithm HS256
        }
        subscriber_jwt {
            key "!ChangeThisMercureHubJWTSecretKey!"
            algorithm HS256
        }
        anonymous  # Allow anonymous subscriptions (change in production)
    }
    
    php_server {
        workers 4
        worker_script index.php
    }
    
    # Handle routing (excluding Mercure endpoint)
    @notStatic {
        not path *.css *.js *.png *.jpg *.gif *.svg *.ico
        not path /.well-known/mercure
        file {
            try_files {path} /index.php
        }
    }
    rewrite @notStatic /index.php
    
    file_server
}
```

## Multi-Site Configuration

FrankenPHP supports running multiple sites simultaneously, each with its own mode and configuration.

### Directory Structure

```
/etc/frankenphp/
├── Caddyfile                    # Main configuration
├── sites-available/             # All site configs
│   ├── site1.caddy             # Classic mode site
│   ├── site2.caddy             # Worker mode site
│   └── site3.caddy             # Mercure mode site
└── sites-enabled/               # Enabled sites (symlinks)
    ├── site1.caddy -> ../sites-available/site1.caddy
    └── site2.caddy -> ../sites-available/site2.caddy
```

### Managing Sites

**List all sites:**
```bash
frankenphp-site list
```

**Enable a site:**
```bash
frankenphp-site enable site1.caddy
systemctl reload frankenphp
```

**Disable a site:**
```bash
frankenphp-site disable site1.caddy
systemctl reload frankenphp
```

```

## Managing Services with Ravact

Ravact provides a comprehensive TUI for managing your FrankenPHP services.

### 1. Service List
Navigate to **FrankenPHP Manager** -> **Manage Services**. You will see a list of all detected FrankenPHP services (detected by `frankenphp-*.service` files).

- **Status Indicators**: 
  - `●` Running
  - `○` Stopped
  - `✗` Failed

### 2. Service Actions
Select a service and press **Enter** to see available actions:

- **Start / Stop / Restart**: Control the service lifecycle.
- **Enable / Disable**: Configure auto-start on boot.
- **View Status**: Shows the full systemd status output.
- **View Logs**: Tails the last 100 lines of the service journal (`journalctl`).
- **Edit Configuration**:
  - **Form Mode**: Edit common settings (Port, User, PHP.ini) in a guided form.
  - **Editor Mode**: Directly edit the `Caddyfile`, `Systemd Service`, or `Nginx Config` in your preferred editor (nano/vi).
- **View Nginx Config**: Generates and displays the correct Nginx configuration block to proxy traffic to this service (supports both Socket and Port modes).
- **Delete Service**: Completely removes the service, configuration files, and data directories (with confirmation).

### 3. Nginx Reverse Proxy Integration
Ravact makes it easy to put FrankenPHP behind Nginx. 
1. Go to **View Nginx Config** in the service actions.
2. The correct `upstream` and `server` block will be generated based on your service's connection type (Unix Socket or TCP Port).
3. Press **c** to copy the config to your clipboard.
4. Paste it into your Nginx site configuration (e.g., via **Nginx Manager** in Ravact).

## Setting Up a New Site

### Step 1: Create Site Directory

```bash
# For Laravel/modern frameworks
mkdir -p /var/www/mysite/public

# For classic PHP
mkdir -p /var/www/mysite

# Set permissions
chown -R www-data:www-data /var/www/mysite
```

### Step 2: Create Site Configuration

**Choose your mode and copy example:**

```bash
# For Laravel (Worker Mode)
cp /etc/frankenphp/sites-available/example-worker.caddy \
   /etc/frankenphp/sites-available/mysite.caddy

# For Classic PHP
cp /etc/frankenphp/sites-available/example-classic.caddy \
   /etc/frankenphp/sites-available/mysite.caddy

# For Real-time app (Mercure)
cp /etc/frankenphp/sites-available/example-mercure.caddy \
   /etc/frankenphp/sites-available/mysite.caddy
```

### Step 3: Edit Configuration

```bash
nano /etc/frankenphp/sites-available/mysite.caddy
```

**Update:**
- Domain name (e.g., `mysite.example.com`)
- Root path (e.g., `/var/www/mysite/public`)
- Number of workers (2-8 depending on CPU cores)
- Mercure JWT secret (if using Mercure)
- Log file path

### Step 4: Enable Site

```bash
frankenphp-site enable mysite.caddy
systemctl reload frankenphp
```

### Step 5: Verify

```bash
# Check FrankenPHP status
systemctl status frankenphp

# Test with curl
curl -H 'Host: mysite.example.com' http://localhost

# Check logs
tail -f /var/log/frankenphp/mysite-access.log
```

## Laravel Specific Configuration

### Worker Mode Setup

```caddy
laravel-app.com {
    root * /var/www/laravel-app/public
    
    # Encode responses
    encode gzip
    
    # PHP workers
    php_server {
        workers 4
        worker_script index.php
    }
    
    # Laravel routing
    @notStatic {
        not path *.css *.js *.png *.jpg *.gif *.svg *.ico *.woff *.woff2 *.ttf
        file {
            try_files {path} /index.php
        }
    }
    
    rewrite @notStatic /index.php
    
    # Static file handling
    file_server
    
    # Security headers
    header {
        X-Frame-Options "SAMEORIGIN"
        X-Content-Type-Options "nosniff"
        X-XSS-Protection "1; mode=block"
        Referrer-Policy "strict-origin-when-cross-origin"
    }
    
    # Logs
    log {
        output file /var/log/frankenphp/laravel-app-access.log
        format json
    }
}
```

### Laravel Octane Alternative

FrankenPHP Worker mode is similar to Laravel Octane but doesn't require Octane package:

**Benefits over Octane:**
- No additional package needed
- Built-in web server (no Nginx needed)
- Automatic HTTPS with Caddy
- Early Hints support
- Easier deployment

**Limitations:**
- Must handle state management (like Octane)
- Need to clear singletons between requests
- Watch for memory leaks

## Performance Tuning

### Worker Count

**Formula:**
```
Workers = (CPU Cores × 2) to (CPU Cores × 4)
```

**Examples:**
- 2 CPU cores: 4-8 workers
- 4 CPU cores: 8-16 workers
- 8 CPU cores: 16-32 workers

**Monitoring:**
```bash
# Watch CPU and memory
htop

# Check worker performance
journalctl -u frankenphp -f
```

### Memory Management

**Per-worker memory:**
- Classic app: 30-50 MB
- Laravel app: 50-100 MB
- Complex app: 100-200 MB

**Calculate max workers:**
```
Max Workers = (Available RAM - System RAM) / Worker Memory
```

**Example:**
- 4 GB RAM server
- 1 GB for system
- 100 MB per worker
- Max workers: (4096 - 1024) / 100 = ~30 workers

## Mercure Real-Time Setup

### Generating JWT Secrets

```bash
# Generate secure secret
openssl rand -base64 32
```

### Laravel + Mercure

**Install package:**
```bash
composer require symfony/mercure-bundle
```

**Configure (.env):**
```env
MERCURE_URL=http://localhost/.well-known/mercure
MERCURE_PUBLIC_URL=https://your-domain.com/.well-known/mercure
MERCURE_JWT_SECRET=!ChangeThisMercureHubJWTSecretKey!
```

**Publishing updates:**
```php
use Symfony\Component\Mercure\Update;
use Symfony\Component\Mercure\HubInterface;

class NotificationController
{
    public function notify(HubInterface $hub)
    {
        $update = new Update(
            'https://your-domain.com/notifications/1',
            json_encode(['message' => 'Hello from Mercure!'])
        );
        
        $hub->publish($update);
        
        return response()->json(['status' => 'published']);
    }
}
```

**Subscribing (JavaScript):**
```javascript
const eventSource = new EventSource(
    'https://your-domain.com/.well-known/mercure?topic=https://your-domain.com/notifications/1'
);

eventSource.onmessage = event => {
    console.log(JSON.parse(event.data));
};
```

## SSL/HTTPS Configuration

Caddy automatically handles HTTPS with Let's Encrypt:

```caddy
# Automatic HTTPS
mysite.com {
    root * /var/www/mysite/public
    php_server { workers 4 }
    file_server
}
```

**Requirements:**
- Domain points to your server
- Ports 80 and 443 open
- Valid DNS records

## Troubleshooting

### Site Not Loading

**Check service:**
```bash
systemctl status frankenphp
journalctl -u frankenphp -n 50
```

**Check configuration:**
```bash
frankenphp validate --config /etc/frankenphp/Caddyfile
```

**Check site is enabled:**
```bash
frankenphp-site list
```

### Worker Mode Issues

**Memory leaks:**
- Clear Laravel cache between requests
- Avoid global state
- Use `ResetInterface` in Laravel 11+

**Performance issues:**
- Reduce worker count
- Check for slow database queries
- Monitor memory usage

### Mercure Not Working

**Check Mercure endpoint:**
```bash
curl http://localhost/.well-known/mercure
```

**Verify JWT secret:**
- Must be same in config and application
- Must be URL-safe (no special chars)

**CORS issues:**
- Add `cors_origins *` for development
- Restrict in production

## Migration from Nginx + PHP-FPM

### Steps

1. **Keep Nginx running** (don't stop yet)

2. **Install FrankenPHP** on different port:
```caddy
:8080
```

3. **Test thoroughly**

4. **Update DNS/proxy** to FrankenPHP

5. **Stop Nginx** when confirmed working

### Comparison

| Feature | Nginx + PHP-FPM | FrankenPHP |
|---------|----------------|------------|
| Setup | Complex | Simple |
| Performance | Good | Excellent (Worker mode) |
| HTTPS | Manual (Certbot) | Automatic |
| Configuration | Two configs | One config |
| Real-time | Requires additional setup | Built-in (Mercure) |
| Early Hints | No | Yes |

## Best Practices

1. **Use Worker Mode** for production Laravel apps
2. **Set appropriate worker count** based on CPU/RAM
3. **Monitor memory usage** to prevent leaks
4. **Use Mercure** for real-time features
5. **Enable logging** for debugging
6. **Set security headers** in all sites
7. **Use HTTPS** in production (automatic with Caddy)
8. **Test locally** before production deployment
9. **Keep FrankenPHP updated** for security and features
10. **Backup configurations** before changes

## Resources

- **FrankenPHP Documentation**: https://frankenphp.dev
- **Caddy Documentation**: https://caddyserver.com/docs
- **Mercure Protocol**: https://mercure.rocks
- **Laravel Octane Alternative**: https://frankenphp.dev/docs/laravel/

---

**Created by Ravact** - Linux Server Management TUI
