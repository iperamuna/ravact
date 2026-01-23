# Complete Test Cases for Ravact

This document provides comprehensive test cases for all Ravact features.

## Test Environment

- **OS:** Ubuntu 24.04 LTS
- **Architecture:** ARM64 or AMD64
- **Services Required:** MySQL, PostgreSQL, PHP-FPM, Supervisor, Nginx, Redis
- **Access:** Root/sudo privileges

---

## 1. Configuration Menu - Service Detection

### Test Case 1.1: Installed Services Display
**Objective:** Verify installed services are shown as available

**Steps:**
1. Navigate to Main Menu → Configurations
2. Observe the service list

**Expected Result:**
- Installed services are displayed with normal styling
- Services are selectable (can press Enter)
- Description does NOT contain "(Not Installed)"

**Services to Check:**
- ✓ Nginx Web Server
- ✓ MySQL Database
- ✓ PostgreSQL Database
- ✓ PHP-FPM Pools
- ✓ Supervisor

### Test Case 1.2: Uninstalled Services Display
**Objective:** Verify uninstalled services are grayed out

**Steps:**
1. Uninstall a service (e.g., `apt remove redis-server`)
2. Navigate to Main Menu → Configurations
3. Try to select the uninstalled service

**Expected Result:**
- Service shows `[Not Installed]` tag
- Service is grayed out
- Cannot be selected with Enter key
- Description includes "(Not Installed)"

---

## 2. MySQL Management

### Test Case 2.1: View Current Configuration
**Steps:**
1. Main Menu → Configurations → MySQL Database
2. Select "View Current Configuration"

**Expected Result:**
- Port number displayed (default: 3306)
- Bind address displayed
- Config path shown
- Data directory shown
- Success message: "✓ Configuration refreshed"

### Test Case 2.2: Change Root Password
**Steps:**
1. Main Menu → Configurations → MySQL Database
2. Select "Change Root Password"
3. Enter new password (e.g., "newpass123")
4. Press Enter

**Expected Result:**
- Password input is masked (••••••)
- Success message: "Root password changed successfully"
- Returns to management menu
- Verify: `mysql -u root -pnewpass123 -e "SELECT 1;"`

### Test Case 2.3: Change Port
**Steps:**
1. Main Menu → Configurations → MySQL Database
2. Select "Change Port"
3. Current port is displayed
4. Enter new port (e.g., 3307)
5. Press Enter

**Expected Result:**
- Current port displayed
- New port validated (must be 1024-65535)
- Service restarts automatically
- Success message: "Port changed to 3307 and service restarted"
- Verify: `mysql -P 3307 -u root -e "SELECT 1;"`

### Test Case 2.4: Invalid Port
**Steps:**
1. Select "Change Port"
2. Enter invalid port (e.g., 100 or 70000)

**Expected Result:**
- Error message: "invalid port number" or validation error
- Port is NOT changed
- Service continues running on original port

### Test Case 2.5: Restart Service
**Steps:**
1. Select "Restart MySQL Service"

**Expected Result:**
- Success message: "✓ MySQL service restarted successfully"
- Verify: `systemctl status mysql` shows "active (running)"

### Test Case 2.6: View Service Status
**Steps:**
1. Select "View Service Status"

**Expected Result:**
- Opens execution screen
- Shows `systemctl status mysql` output
- Status includes: active, running, PID, memory usage

### Test Case 2.7: List Databases
**Steps:**
1. Select "List Databases"

**Expected Result:**
- Success message shows count: "✓ Found X databases"
- System databases excluded (information_schema, mysql, performance_schema, sys)
- User databases listed

---

## 3. PostgreSQL Management

### Test Case 3.1: View Current Configuration
**Steps:**
1. Main Menu → Configurations → PostgreSQL Database
2. Select "View Current Configuration"

**Expected Result:**
- Port number (default: 5432)
- Max connections value
- Shared buffers value
- Config path
- Success message: "✓ Configuration refreshed"

### Test Case 3.2: Change Postgres Password
**Steps:**
1. Select "Change Postgres Password"
2. Enter new password
3. Press Enter

**Expected Result:**
- Password masked
- Success: "Postgres user password changed successfully"
- Verify: `sudo -u postgres psql -c "SELECT 1;"`

### Test Case 3.3: Change Port
**Steps:**
1. Select "Change Port"
2. Enter new port (e.g., 5433)

**Expected Result:**
- Port changed
- Service restarted
- Success message displayed
- Verify: `psql -U postgres -p 5433 -c "SELECT 1;"`

### Test Case 3.4: List Databases
**Steps:**
1. Select "List Databases"

**Expected Result:**
- User databases listed
- Template databases excluded
- Count displayed

---

## 4. PHP-FPM Pool Management

### Test Case 4.1: List All Pools
**Steps:**
1. Main Menu → Configurations → PHP-FPM Pools
2. Select "List All Pools"

**Expected Result:**
- Success: "✓ Found X pools"
- Each pool shows: name, PM mode, listen socket/port
- Default 'www' pool visible

### Test Case 4.2: View Pool Details
**Steps:**
1. After listing pools, navigate to pool list
2. Select a pool

**Expected Result:**
- Pool configuration displayed:
  - Name, User, Group
  - Listen address
  - PM mode and worker settings
  - Config file path

### Test Case 4.3: Restart Service
**Steps:**
1. Select "Restart PHP-FPM Service"

**Expected Result:**
- Success: "✓ PHP-FPM service restarted successfully"
- Verify: `systemctl status php*-fpm`

### Test Case 4.4: Reload Service
**Steps:**
1. Select "Reload PHP-FPM Service"

**Expected Result:**
- Success: "✓ PHP-FPM service reloaded successfully"
- Graceful reload (no downtime)

---

## 5. Supervisor Management

### Test Case 5.1: List All Programs
**Steps:**
1. Main Menu → Configurations → Supervisor
2. Select "List All Programs"

**Expected Result:**
- Programs listed with states
- Color coding: RUNNING (green), STOPPED (red)
- Count displayed

### Test Case 5.2: Configure XML-RPC
**Steps:**
1. Select "Configure XML-RPC"
2. Enter IP: 127.0.0.1
3. Enter Port: 9001
4. Enter Username: admin
5. Enter Password: secret123
6. Press Enter

**Expected Result:**
- All fields accept input
- Password is masked
- Success: "XML-RPC configured successfully. Supervisor will restart."
- Verify: Check `http://127.0.0.1:9001` (requires auth)
- Config in `/etc/supervisor/supervisord.conf`

### Test Case 5.3: View XML-RPC Config
**Steps:**
1. Select "View XML-RPC Config"

**Expected Result:**
- Shows: Enabled status, IP, Port, Username
- Password shown as "[configured]"

### Test Case 5.4: Add New Program - Valid Config
**Steps:**
1. Select "Add New Program"
2. Enter program name: "testapp"
3. Choose editor: nano
4. Edit config (keep valid template or modify correctly)
5. Save and exit editor

**Expected Result:**
- Step 1: Name input accepted
- Step 2: Editor selection shows nano/vi
- Step 3: Editor opens with template
- Validation runs: `supervisorctl reread`
- Success: "Program 'testapp' added successfully and configuration is valid"
- Verify: `supervisorctl status testapp`

### Test Case 5.5: Add New Program - Invalid Config
**Steps:**
1. Select "Add New Program"
2. Enter name: "badapp"
3. Choose editor
4. Break the config (remove required field like `command=`)
5. Save and exit

**Expected Result:**
- Validation fails
- Error: "Configuration validation failed: [error details]"
- Warning: "The configuration is invalid and was not applied"
- Config file removed
- Verify: `ls /etc/supervisor/conf.d/badapp.conf` (should not exist)

### Test Case 5.6: Restart Supervisor
**Steps:**
1. Select "Restart Supervisor"

**Expected Result:**
- Success: "✓ Supervisor restarted successfully"
- All programs reload
- Verify: `systemctl status supervisor`

---

## 6. Quick Commands

### Test Case 6.1: System Info
**Steps:**
1. Main Menu → Quick Commands
2. Select "System Info"

**Expected Result:**
- Shows output of `uname -a`
- Displays: OS name, kernel version, architecture
- Example: "Linux ravact-dev 6.8.0-51-generic ... aarch64 GNU/Linux"

### Test Case 6.2: Disk Usage
**Steps:**
1. Select "Disk Usage"

**Expected Result:**
- Shows `df -h` output
- Table with: Filesystem, Size, Used, Avail, Use%, Mounted on
- Human-readable sizes (GB, MB)

### Test Case 6.3: Memory Info
**Steps:**
1. Select "Memory Info"

**Expected Result:**
- Shows `free -h` output
- Displays: total, used, free, shared, buff/cache, available
- Human-readable format

### Test Case 6.4: Running Services
**Steps:**
1. Select "Running Services"

**Expected Result:**
- Shows active systemd services
- List includes: mysql, postgresql, nginx, supervisor, etc.
- Service names with .service extension

### Test Case 6.5: Network Info
**Steps:**
1. Select "Network Info"

**Expected Result:**
- Shows `ip addr show` output
- Lists network interfaces: lo, eth0, etc.
- IP addresses displayed
- MAC addresses shown

### Test Case 6.6: Top Processes
**Steps:**
1. Select "Top Processes"

**Expected Result:**
- Shows `ps aux --sort=-pcpu | head -20`
- 20 processes sorted by CPU usage
- Columns: USER, PID, %CPU, %MEM, COMMAND

### Test Case 6.7: Recent Logs
**Steps:**
1. Select "Recent Logs"

**Expected Result:**
- Shows `journalctl -n 50` output
- Last 50 system journal entries
- Timestamps and service names visible

---

## 7. User Management

### Test Case 7.1: List Users
**Steps:**
1. Main Menu → User Management

**Expected Result:**
- System users listed
- Shows user details
- Indicates sudo access

### Test Case 7.2: Add User
**Steps:**
1. Select "Add User"
2. Enter username
3. Enter password
4. Optionally grant sudo

**Expected Result:**
- User created
- Home directory created
- Sudo access granted if selected
- Verify: `id username`

---

## 8. Nginx Management

### Test Case 8.1: List Sites
**Steps:**
1. Main Menu → Configurations → Nginx Web Server

**Expected Result:**
- Shows configured sites
- Enabled/disabled status

### Test Case 8.2: Add Site
**Steps:**
1. Press 'a' to add site
2. Enter domain
3. Choose template
4. Configure options

**Expected Result:**
- Site configuration created
- Nginx config valid
- Site listed

---

## 9. Redis Configuration

### Test Case 9.1: Configure Password
**Steps:**
1. Main Menu → Configurations → Redis Cache
2. Configure password

**Expected Result:**
- Password set in redis.conf
- Service restarted
- Connection test works

---

## 10. Error Handling

### Test Case 10.1: Service Not Running
**Steps:**
1. Stop a service: `systemctl stop mysql`
2. Try to configure it in Ravact

**Expected Result:**
- Appropriate error message
- Suggestion to start service

### Test Case 10.2: Permission Denied
**Steps:**
1. Run Ravact without sudo
2. Try to configure services

**Expected Result:**
- Permission error displayed
- Suggestion to run with sudo

### Test Case 10.3: Invalid Input
**Steps:**
1. Enter invalid data (empty string, special chars)

**Expected Result:**
- Validation error displayed
- No changes applied
- Clear error message

---

## 11. Navigation Testing

### Test Case 11.1: Back Navigation
**Steps:**
1. Navigate deep into menus
2. Press Esc at each level

**Expected Result:**
- Returns to previous screen
- No data loss
- Smooth transitions

### Test Case 11.2: Quit Application
**Steps:**
1. Press 'q' from various screens

**Expected Result:**
- Application exits cleanly
- No hanging processes
- Terminal restored

---

## 12. Integration Testing

### Test Case 12.1: Full Workflow - MySQL
**Steps:**
1. View MySQL config
2. Change password
3. Change port
4. Create database
5. List databases
6. Restart service

**Expected Result:**
- All operations succeed
- Changes persist
- No conflicts

### Test Case 12.2: Full Workflow - Supervisor
**Steps:**
1. Configure XML-RPC
2. Add program with valid config
3. Start program
4. View status
5. Stop program

**Expected Result:**
- All operations work
- Program lifecycle managed correctly
- XML-RPC accessible

---

## Test Execution Checklist

### Pre-Test Setup
- [ ] Fresh Ubuntu 24.04 installation
- [ ] All services installed (MySQL, PostgreSQL, PHP-FPM, Supervisor, Nginx, Redis)
- [ ] Root/sudo access available
- [ ] Ravact binary built and executable

### Core Features
- [ ] Service detection works
- [ ] MySQL management (all 7 features)
- [ ] PostgreSQL management (all 7 features)
- [ ] PHP-FPM management (all 4 features)
- [ ] Supervisor management (all 6 features)
- [ ] Quick Commands (all 7 commands)
- [ ] User management
- [ ] Nginx management
- [ ] Redis configuration

### Edge Cases
- [ ] Invalid inputs handled
- [ ] Service failures handled
- [ ] Permission errors handled
- [ ] Config validation works
- [ ] Rollback on errors

### Performance
- [ ] UI responsive
- [ ] Commands execute quickly
- [ ] No memory leaks
- [ ] Clean exit

---

## Bug Reporting Template

```markdown
**Test Case:** [Test case number and name]
**Expected:** [Expected result]
**Actual:** [Actual result]
**Steps to Reproduce:**
1. 
2. 
3. 

**Environment:**
- OS: Ubuntu 24.04
- Architecture: ARM64/AMD64
- Ravact Version: [version]
- Service Version: [if applicable]

**Logs/Output:**
```
[paste relevant output]
```
```

---

## Automated Test Script

```bash
#!/bin/bash
# automated-test.sh

echo "=== Ravact Automated Tests ==="

# Check all services
services=(mysql postgresql php8.3-fpm supervisor nginx redis-server)
for svc in "${services[@]}"; do
    if systemctl is-active --quiet $svc 2>/dev/null; then
        echo "✓ $svc is running"
    else
        echo "✗ $svc is NOT running"
    fi
done

# Check configurations exist
configs=(
    "/etc/mysql/mysql.conf.d/mysqld.cnf"
    "/etc/postgresql/*/main/postgresql.conf"
    "/etc/php/*/fpm/pool.d/www.conf"
    "/etc/supervisor/supervisord.conf"
    "/etc/nginx/nginx.conf"
    "/etc/redis/redis.conf"
)

for cfg in "${configs[@]}"; do
    if ls $cfg 1>/dev/null 2>&1; then
        echo "✓ Config exists: $cfg"
    else
        echo "✗ Config missing: $cfg"
    fi
done

echo ""
echo "Manual testing required for Ravact UI features"
echo "Run: sudo ./ravact"
```

---

**Test Status:** All features implemented and ready for testing
**Last Updated:** January 2026
**Total Test Cases:** 40+
