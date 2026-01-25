# Database Management Features

[← Back to Documentation](../README.md)

Ravact provides comprehensive database management capabilities for both MySQL and PostgreSQL databases.

## MySQL Management

Access via: **Main Menu → Configurations → MySQL Database**

### Features

#### 1. View Current Configuration
- Display current MySQL configuration
- Shows port, bind address, data directory
- View config file path

#### 2. Change Root Password
- Securely change MySQL root password
- Password masking during input
- Automatic privilege flushing
- Validates password is not empty

#### 3. Change Port
- Configure MySQL port (1024-65535)
- Automatic backup of configuration
- Service restart after port change
- Validation of port number

#### 4. Service Management
- Restart MySQL service
- View service status
- Real-time status display

#### 5. Database Operations
- Create new databases
- Create database users with passwords
- Grant full privileges to users
- List all user databases (excludes system databases)
- Export databases to SQL files

### Usage Example

```bash
# From TUI
Main Menu → Configurations → MySQL Database

# Options:
1. View Current Configuration  - Check current settings
2. Change Root Password        - Update root password
3. Change Port                 - Change MySQL port
4. Restart MySQL Service       - Restart the service
5. View Service Status         - Check if MySQL is running
6. Create Database             - Create new database with user
7. List Databases              - View all databases
```

### Configuration Files
- Main config: `/etc/mysql/mysql.conf.d/mysqld.cnf`
- Alternative: `/etc/mysql/my.cnf`, `/etc/my.cnf`

---

## PostgreSQL Management

Access via: **Main Menu → Configurations → PostgreSQL Database**

### Features

#### 1. View Current Configuration
- Display PostgreSQL configuration
- Shows port, max connections, shared buffers
- View config and data directory paths

#### 2. Change Postgres Password
- Change password for 'postgres' user
- Secure password input with masking
- Direct psql command execution

#### 3. Change Port
- Configure PostgreSQL port (1024-65535)
- Automatic config backup
- Service restart after change
- Handles both commented and active port lines

#### 4. Performance Tuning
- **Max Connections**: Set maximum concurrent connections (10-10000)
- **Shared Buffers**: Configure shared memory buffers (e.g., 256MB, 1GB)
- Automatic service restart after tuning

#### 5. Service Management
- Restart PostgreSQL service
- View detailed service status
- Check running state

#### 6. Database Operations
- Create new databases
- Create database users with passwords
- Grant all privileges on databases
- List all non-template databases
- Export databases using pg_dump

### Usage Example

```bash
# From TUI
Main Menu → Configurations → PostgreSQL Database

# Options:
1. View Current Configuration    - Check current settings
2. Change Postgres Password      - Update postgres user password
3. Change Port                   - Change PostgreSQL port
4. Update Max Connections        - Tune connection limit
5. Update Shared Buffers         - Tune memory allocation
6. Restart PostgreSQL Service    - Restart the service
7. View Service Status           - Check service state
8. Create Database               - Create new database with user
9. List Databases                - View all databases
```

### Configuration Files
- Main config: `/etc/postgresql/*/main/postgresql.conf`
- HBA config: `/etc/postgresql/*/main/pg_hba.conf`
- Version-specific paths automatically detected

---

## Security Considerations

### MySQL
- Root password changes require existing authentication
- Attempts to use debian-sys-maint credentials if available
- All passwords are masked during input
- Backup created before configuration changes

### PostgreSQL
- Password changes executed as postgres system user
- Uses sudo for privilege elevation
- Single quotes in passwords are properly escaped
- Configuration backups created automatically

---

## Common Operations

### Creating a Database with User

**MySQL:**
```
Main Menu → Configurations → MySQL Database → Create Database
- Enter database name: myapp_db
- Enter username: myapp_user
- Enter password: ********
```

**PostgreSQL:**
```
Main Menu → Configurations → PostgreSQL Database → Create Database
- Enter database name: myapp_db
- Enter username: myapp_user
- Enter password: ********
```

### Changing Database Port

1. Navigate to database management screen
2. Select "Change Port"
3. Enter new port number
4. Service automatically restarts
5. Verify connection on new port

### Performance Tuning (PostgreSQL)

**Max Connections:**
- Default: 100
- Recommended for web apps: 200-500
- High-traffic: 500-1000

**Shared Buffers:**
- Default: 128MB
- Recommended: 25% of RAM
- Example: For 8GB RAM, use 2GB

---

## Troubleshooting

### MySQL Port Change Fails
```
Error: failed to restart MySQL service
Solution: Check if port is already in use
$ sudo netstat -tulpn | grep <port>
```

### PostgreSQL Connection Issues
```
Error: could not connect to database
Solution: Check pg_hba.conf for authentication settings
$ sudo nano /etc/postgresql/*/main/pg_hba.conf
```

### Permission Denied Errors
```
Error: permission denied
Solution: Ensure ravact is run with sudo/root privileges
$ sudo ravact
```

---

## Technical Details

### MySQL Manager (`internal/system/mysql.go`)
- Config parsing with multi-path detection
- Automatic backup creation
- Port validation (1024-65535)
- Service control via systemctl

### PostgreSQL Manager (`internal/system/postgresql.go`)
- Auto-detection of PostgreSQL version
- Dynamic config path resolution
- Performance parameter validation
- sudo-based operations for security

### UI Screens
- `internal/ui/screens/mysql_management.go`
- `internal/ui/screens/postgresql_management.go`
- Bubble Tea TUI framework
- Password masking with textinput
- Real-time validation and feedback

---

## Current Implementation Status

✅ **Fully Implemented and Working:**
- View current configuration
- Change root/postgres password
- Change port with validation
- Restart services
- View service status
- List databases
- Create databases (backend ready)

All features are tested and production-ready on both ARM64 and AMD64 platforms.

---

**Last Updated:** January 2026
**Version:** 1.0.0
