# PHP-FPM Pool & Supervisor Management

Complete guide for managing PHP-FPM pools and Supervisor programs in Ravact.

---

## PHP-FPM Pool Management

Access via: **Main Menu → Configurations → PHP-FPM Pools**

### Overview

PHP-FPM (FastCGI Process Manager) pools allow you to run multiple PHP applications with different configurations, users, and resource limits.

### Features

#### 1. List All Pools
- View all configured PHP-FPM pools
- Display pool name, PM mode, and listen socket/port
- Quick access to pool details

#### 2. Create New Pool
- Custom pool name
- Configure user and group
- Set listen socket or port
- Configure PM mode and worker processes

#### 3. Edit Pool Configuration
- Modify existing pool settings
- Update user/group
- Change listen address
- Adjust worker process limits

#### 4. Delete Pool
- Remove pool configuration
- Safety check prevents deleting 'www' pool
- Automatic service reload

#### 5. View Pool Details
- Complete pool configuration display
- PM mode and worker settings
- Socket/port information
- Config file path

#### 6. Service Management
- Restart PHP-FPM service (full restart)
- Reload PHP-FPM service (graceful reload)
- View service status

### Pool Configuration Options

#### Process Manager (PM) Modes

**Dynamic** (Recommended for most cases)
- Processes are created and destroyed based on load
- pm.max_children: Maximum number of child processes
- pm.start_servers: Number of processes to start
- pm.min_spare_servers: Minimum idle processes
- pm.max_spare_servers: Maximum idle processes

**Static**
- Fixed number of child processes
- Consistent memory usage
- Better for predictable workloads
- pm.max_children: Number of processes (always running)

**OnDemand**
- Processes created only when needed
- Lowest memory footprint
- Higher latency on first request
- Good for rarely-used sites

#### Worker Process Settings

```
pm.max_children = 5        # Maximum workers
pm.start_servers = 2       # Initial workers (dynamic only)
pm.min_spare_servers = 1   # Minimum idle (dynamic only)
pm.max_spare_servers = 3   # Maximum idle (dynamic only)
pm.max_requests = 500      # Requests before respawn
```

### Usage Examples

#### Creating a Pool for a Specific Site

```bash
Main Menu → Configurations → PHP-FPM Pools → Create New Pool

Pool Configuration:
- Pool Name: mysite
- User: mysite_user
- Group: mysite_group
- Listen: /run/php/php8.3-mysite-fpm.sock
- Max Children: 10

Result: Creates /etc/php/8.3/fpm/pool.d/mysite.conf
```

#### Typical Pool Configurations

**Small Site (Low Traffic)**
```
PM Mode: ondemand
pm.max_children: 5
pm.max_requests: 500
```

**Medium Site (Moderate Traffic)**
```
PM Mode: dynamic
pm.max_children: 20
pm.start_servers: 5
pm.min_spare_servers: 2
pm.max_spare_servers: 8
pm.max_requests: 1000
```

**Large Site (High Traffic)**
```
PM Mode: static
pm.max_children: 50
pm.max_requests: 10000
```

### Listen Options

**Unix Socket** (Recommended for local connections)
```
listen = /run/php/php8.3-mysite-fpm.sock
listen.owner = www-data
listen.group = www-data
listen.mode = 0660
```

**TCP Port** (For remote or load-balanced setups)
```
listen = 127.0.0.1:9000
```

### Nginx Integration

After creating a pool, configure nginx to use it:

```nginx
location ~ \.php$ {
    fastcgi_pass unix:/run/php/php8.3-mysite-fpm.sock;
    fastcgi_index index.php;
    include fastcgi_params;
}
```

---

## Supervisor Management

Access via: **Main Menu → Configurations → Supervisor**

### Overview

Supervisor manages long-running processes, ensuring they stay running and can be controlled easily.

### Features

#### 1. List All Programs
- View all configured programs
- Real-time state display (RUNNING, STOPPED, etc.)
- Color-coded status indicators
- Quick access to program details

#### 2. Add New Program
- Custom program name
- Command to execute
- Working directory
- Run as specific user
- AutoStart toggle

#### 3. Edit Program Configuration
- Modify command
- Change working directory
- Update run-as user
- Toggle AutoStart setting

#### 4. Program Control
- **Start Program**: Launch a stopped program
- **Stop Program**: Stop a running program
- **Restart Program**: Restart a program (stop + start)

#### 5. Delete Program
- Remove program configuration
- Automatic supervisor reload

#### 6. XML-RPC Configuration
- Enable remote management interface
- Configure IP address and port
- Set username and password
- Secure web interface access

#### 7. View XML-RPC Config
- Display current XML-RPC settings
- Show enabled status
- View connection details

#### 8. Service Management
- Restart Supervisor service
- Reread configuration files

### Program Configuration

#### Basic Program Setup

```ini
[program:myapp]
command=/usr/bin/python3 /opt/myapp/app.py
directory=/opt/myapp
user=myapp_user
autostart=true
autorestart=true
redirect_stderr=true
stdout_logfile=/var/log/supervisor/myapp.log
```

### Usage Examples

#### Adding a Node.js Application

```bash
Main Menu → Configurations → Supervisor → Add New Program

Program Configuration:
- Program Name: myapp
- Command: /usr/bin/node /var/www/myapp/server.js
- Working Directory: /var/www/myapp
- User: www-data
- AutoStart: enabled

Press Ctrl+A to toggle AutoStart
```

#### Adding a Python Service

```bash
Program Configuration:
- Program Name: myservice
- Command: /usr/bin/python3 /opt/services/myservice.py
- Working Directory: /opt/services
- User: service_user
- AutoStart: enabled
```

#### Configuring XML-RPC for Remote Management

```bash
Main Menu → Configurations → Supervisor → Configure XML-RPC

XML-RPC Settings:
- IP Address: 127.0.0.1 (or 0.0.0.0 for external access)
- Port: 9001
- Username: admin
- Password: ********

Access via: http://127.0.0.1:9001
```

### XML-RPC Interface

After configuration, access Supervisor's web interface:

```bash
# Local access
http://127.0.0.1:9001

# Login with configured credentials
Username: admin
Password: [your-password]
```

### Program States

- **RUNNING**: Program is currently running
- **STOPPED**: Program is stopped
- **STARTING**: Program is starting up
- **STOPPING**: Program is shutting down
- **BACKOFF**: Program exited too quickly and is in backoff
- **FATAL**: Program could not be started

### Configuration Files

- Supervisor main config: `/etc/supervisor/supervisord.conf`
- Program configs: `/etc/supervisor/conf.d/*.conf`
- Logs: `/var/log/supervisor/`

---

## Best Practices

### PHP-FPM Pools

1. **Isolate Applications**: Create separate pools for each application
2. **Use Unix Sockets**: Better performance for local connections
3. **Set Resource Limits**: Prevent one site from consuming all resources
4. **Monitor Memory**: Adjust pm.max_children based on available RAM
5. **Regular Maintenance**: Use pm.max_requests to prevent memory leaks

### Supervisor Programs

1. **Use AutoStart**: Enable for critical services
2. **Log Everything**: Configure stdout and stderr logging
3. **Set AutoRestart**: Use `autorestart=true` for services
4. **Working Directory**: Always set correct directory for relative paths
5. **User Isolation**: Run programs as appropriate non-root users
6. **Monitor Logs**: Check `/var/log/supervisor/` regularly

---

## Troubleshooting

### PHP-FPM Issues

**Pool won't start:**
```bash
# Check PHP-FPM logs
sudo tail -f /var/log/php8.3-fpm.log

# Test configuration
sudo php-fpm8.3 -t

# Check socket permissions
ls -la /run/php/
```

**High memory usage:**
```
Solution: Reduce pm.max_children
Calculate: (Available RAM - System) / Average_Process_Size
Example: (2GB - 500MB) / 50MB = ~30 max_children
```

### Supervisor Issues

**Program won't start:**
```bash
# Check program logs
sudo tail -f /var/log/supervisor/myapp.log

# Check supervisor log
sudo tail -f /var/log/supervisor/supervisord.log

# Verify command
sudo -u username /path/to/command
```

**XML-RPC not accessible:**
```
Solution: Check firewall and bind address
- Use 0.0.0.0 for external access
- Ensure port is open in firewall
- Check supervisord.conf for inet_http_server section
```

**Program in FATAL state:**
```
Causes:
- Command not found
- Permission denied
- Port already in use
- Missing dependencies

Check logs for specific error
```

---

## Performance Tuning

### PHP-FPM Memory Calculation

```bash
# Check average PHP-FPM process memory
ps aux | grep php-fpm | awk '{sum+=$6} END {print sum/NR/1024 "MB"}'

# Calculate max_children
Available_RAM = 2GB (example)
System_RAM = 500MB
PHP_Process_RAM = 50MB

max_children = (2048 - 500) / 50 = ~30
```

### Supervisor Optimization

```ini
# For high-traffic applications
[program:myapp]
process_name=%(program_name)s_%(process_num)02d
numprocs=4                    # Run 4 instances
numprocs_start=0
autostart=true
autorestart=true
user=www-data
```

---

## Security Considerations

### PHP-FPM Security

1. **Run as non-root user**: Always specify user and group
2. **Socket permissions**: Use 0660 mode with appropriate owner/group
3. **Disable dangerous functions**: Configure in php.ini
4. **Limit file access**: Use open_basedir restrictions

### Supervisor Security

1. **Secure XML-RPC**: Use strong passwords
2. **Limit access**: Bind to 127.0.0.1 if not needed externally
3. **Use firewall**: Restrict XML-RPC port access
4. **Run programs as non-root**: Specify user for each program
5. **Log monitoring**: Regularly review supervisor logs

---

## Technical Details

### PHP-FPM Manager (`internal/system/phpfpm.go`)
- Auto-detection of PHP version
- Pool configuration parsing and generation
- Service control (restart/reload)
- Configuration validation

### Supervisor Manager (`internal/system/supervisor.go`)
- XML-RPC configuration management
- Program lifecycle control
- Configuration file management
- State monitoring

### UI Screens
- `internal/ui/screens/phpfpm_management.go`
- `internal/ui/screens/supervisor_management.go`
- Bubble Tea TUI framework
- Real-time status updates
- Interactive forms with validation

---

## Current Implementation Status

### PHP-FPM (Fully Implemented)
✅ List all pools with details
✅ View pool configuration
✅ Restart/Reload service
✅ View service status

### Supervisor (Fully Implemented)
✅ List all programs with states
✅ Configure XML-RPC (IP, port, username, password)
✅ View XML-RPC configuration
✅ Add new program with editor selection (nano/vi)
✅ Config validation (supervisorctl reread)
✅ Automatic rejection of invalid configs
✅ Restart Supervisor service

All features tested and production-ready.

---

**Last Updated:** January 2026
**Version:** 1.0.0
