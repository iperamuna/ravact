# Testing Ravact on Real AMD64/Intel Hardware

This guide explains how to test Ravact on actual AMD64/Intel Linux servers and workstations.

---

## Overview

While ARM64 testing on M1 Macs via Multipass is great for development, **production testing must be done on real AMD64/Intel hardware** to ensure:

- ✅ Full compatibility with x86_64 architecture
- ✅ Performance characteristics match production
- ✅ No ARM-specific issues
- ✅ Service configurations work correctly
- ✅ Real-world systemd behavior

---

## Hardware Options

### Option 1: Bare Metal Server
- **Best for**: Production testing, performance testing
- **Pros**: True production environment, best performance
- **Cons**: Requires dedicated hardware

### Option 2: Virtual Private Server (VPS)
- **Best for**: Quick testing, CI/CD integration
- **Providers**: DigitalOcean, Linode, Vultr, Hetzner
- **Pros**: Quick setup, affordable, easy cleanup
- **Cons**: Shared resources

### Option 3: Local VM (VirtualBox/VMware)
- **Best for**: Development testing on Linux workstations
- **Pros**: Local control, no cost
- **Cons**: Requires x86_64 host machine

### Option 4: Cloud Instances
- **Best for**: Automated testing, scalability
- **Providers**: AWS EC2, Google Cloud, Azure
- **Pros**: Automation, various instance types
- **Cons**: More complex setup

---

## Recommended Testing Environment

### Minimum Specifications
- **CPU**: 2 cores (x86_64/AMD64)
- **RAM**: 4 GB
- **Disk**: 20 GB SSD
- **OS**: Ubuntu 24.04 LTS (Server)
- **Network**: Public IP (for remote testing)

### Recommended Specifications
- **CPU**: 4 cores (x86_64/AMD64)
- **RAM**: 8 GB
- **Disk**: 40 GB SSD
- **OS**: Ubuntu 24.04 LTS (Server)

---

## Quick Setup - VPS Method

### 1. Provision VPS (DigitalOcean Example)

```bash
# Using DigitalOcean CLI (doctl)
doctl compute droplet create ravact-test \
  --size s-2vcpu-4gb \
  --image ubuntu-24-04-x64 \
  --region nyc3 \
  --ssh-keys YOUR_SSH_KEY_ID

# Wait for creation
doctl compute droplet list

# Get IP address
doctl compute droplet get ravact-test --format PublicIPv4
```

### 2. Connect to Server

```bash
# SSH into server
ssh root@YOUR_SERVER_IP

# Update system
apt update && apt upgrade -y
```

### 3. Install Dependencies

```bash
# Install required services
apt install -y \
  mysql-server \
  postgresql \
  postgresql-contrib \
  php8.3-fpm \
  supervisor \
  nginx \
  git \
  build-essential

# Install Go (for building)
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### 4. Deploy Ravact

```bash
# Clone repository
git clone https://github.com/iperamuna/ravact.git
cd ravact

# Build for AMD64
go build -o ravact ./cmd/ravact

# Make executable
chmod +x ravact

# Run
sudo ./ravact
```

---

## Testing Procedure

### Complete Feature Testing

#### 1. MySQL Management Tests

```bash
# Start Ravact
sudo ./ravact

# Navigate: Main Menu → Configurations → MySQL Database

# Test 1: View Configuration
✓ Verify port is displayed (default: 3306)
✓ Check config path is correct
✓ Confirm data directory shown

# Test 2: Change Root Password
✓ Enter new password (test123!)
✓ Verify password change successful
✓ Test MySQL connection: mysql -u root -ptest123!

# Test 3: Change Port
✓ Change port to 3307
✓ Verify service restarts
✓ Test connection: mysql -u root -ptest123! -P 3307

# Test 4: Create Database
✓ Database name: testdb
✓ Username: testuser
✓ Password: testpass
✓ Verify creation: mysql -u root -ptest123! -P 3307 -e "SHOW DATABASES;"

# Test 5: List Databases
✓ Verify testdb appears in list
✓ System databases are excluded

# Test 6: Service Management
✓ Restart MySQL service
✓ Check status shows running
```

#### 2. PostgreSQL Management Tests

```bash
# Navigate: Main Menu → Configurations → PostgreSQL Database

# Test 1: View Configuration
✓ Verify port is displayed (default: 5432)
✓ Check max_connections value
✓ Verify shared_buffers setting

# Test 2: Change Postgres Password
✓ Enter new password (pgtest123!)
✓ Verify success message
✓ Test: sudo -u postgres psql -c "SELECT version();"

# Test 3: Change Port
✓ Change port to 5433
✓ Verify service restarts
✓ Test: psql -U postgres -p 5433

# Test 4: Update Max Connections
✓ Change to 200
✓ Verify service restarts
✓ Check: sudo -u postgres psql -c "SHOW max_connections;"

# Test 5: Update Shared Buffers
✓ Change to 256MB
✓ Verify service restarts
✓ Check: sudo -u postgres psql -c "SHOW shared_buffers;"

# Test 6: Create Database
✓ Database: testpgdb
✓ User: pguser
✓ Password: pgpass
✓ Verify: sudo -u postgres psql -l

# Test 7: List Databases
✓ Verify testpgdb in list
✓ Template databases excluded
```

#### 3. PHP-FPM Pool Tests

```bash
# Navigate: Main Menu → Configurations → PHP-FPM Pools

# Test 1: List Pools
✓ Default 'www' pool appears
✓ Pool details displayed correctly

# Test 2: Create New Pool
✓ Pool name: testpool
✓ User: www-data
✓ Group: www-data
✓ Listen: /run/php/php8.3-testpool-fpm.sock
✓ Max children: 10
✓ Verify creation: ls /etc/php/8.3/fpm/pool.d/testpool.conf

# Test 3: View Pool Details
✓ All configuration displayed
✓ PM mode shown correctly
✓ Worker settings accurate

# Test 4: Edit Pool
✓ Change max_children to 15
✓ Update user (if needed)
✓ Verify: cat /etc/php/8.3/fpm/pool.d/testpool.conf

# Test 5: Service Management
✓ Reload PHP-FPM (graceful)
✓ Restart PHP-FPM (full restart)
✓ Check status shows running

# Test 6: Delete Pool
✓ Select testpool for deletion
✓ Confirm 'www' pool cannot be deleted
✓ Verify deletion: ls /etc/php/8.3/fpm/pool.d/

# Test 7: Service Status
✓ View detailed status
✓ Confirm all pools running
```

#### 4. Supervisor Management Tests

```bash
# Navigate: Main Menu → Configurations → Supervisor

# Test 1: List Programs
✓ Shows all configured programs
✓ States displayed correctly (RUNNING/STOPPED)

# Test 2: Add New Program
✓ Program name: testapp
✓ Command: /usr/bin/python3 -m http.server 8080
✓ Directory: /tmp
✓ User: www-data
✓ AutoStart: enabled
✓ Verify: cat /etc/supervisor/conf.d/testapp.conf

# Test 3: View Program Details
✓ All configuration displayed
✓ State shows correctly
✓ Config path visible

# Test 4: Start Program
✓ Select testapp
✓ Choose "Start Program"
✓ Verify: sudo supervisorctl status testapp
✓ Check: curl http://localhost:8080

# Test 5: Stop Program
✓ Select testapp
✓ Choose "Stop Program"
✓ Verify state changes to STOPPED

# Test 6: Restart Program
✓ Select testapp
✓ Choose "Restart Program"
✓ Verify comes back to RUNNING

# Test 7: Edit Program
✓ Change command or directory
✓ Update user if needed
✓ Toggle AutoStart
✓ Verify changes saved

# Test 8: Configure XML-RPC
✓ IP: 127.0.0.1 (or 0.0.0.0 for external)
✓ Port: 9001
✓ Username: admin
✓ Password: admin123
✓ Verify: cat /etc/supervisor/supervisord.conf | grep inet_http_server -A 4

# Test 9: View XML-RPC Config
✓ Enabled status shown
✓ IP and port displayed
✓ Username shown (password hidden)

# Test 10: Access XML-RPC Interface
✓ Open browser: http://SERVER_IP:9001
✓ Login with credentials
✓ Verify programs listed
✓ Test start/stop from web interface

# Test 11: Delete Program
✓ Select testapp
✓ Choose "Delete Program"
✓ Verify: ls /etc/supervisor/conf.d/
✓ Confirm removed from supervisorctl status
```

---

## Automated Testing Script

Create a comprehensive test script:

```bash
#!/bin/bash
# comprehensive-test.sh

set -e

echo "=== Ravact AMD64 Comprehensive Test ==="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

test_service() {
    local service=$1
    echo -n "Testing $service... "
    if systemctl is-active --quiet $service; then
        echo -e "${GREEN}✓ PASSED${NC}"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAILED${NC}"
        ((TESTS_FAILED++))
        return 1
    fi
}

test_command() {
    local desc=$1
    local cmd=$2
    echo -n "Testing $desc... "
    if eval $cmd > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PASSED${NC}"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAILED${NC}"
        ((TESTS_FAILED++))
        return 1
    fi
}

echo "1. Service Status Tests"
echo "======================="
test_service "mysql"
test_service "postgresql"
test_service "php8.3-fpm"
test_service "supervisor"
test_service "nginx"
echo ""

echo "2. MySQL Tests"
echo "=============="
test_command "MySQL connection" "mysql -u root -e 'SELECT 1;'"
test_command "MySQL config exists" "test -f /etc/mysql/mysql.conf.d/mysqld.cnf"
test_command "MySQL socket exists" "test -S /var/run/mysqld/mysqld.sock"
echo ""

echo "3. PostgreSQL Tests"
echo "==================="
test_command "PostgreSQL connection" "sudo -u postgres psql -c 'SELECT 1;'"
test_command "PostgreSQL config exists" "ls /etc/postgresql/*/main/postgresql.conf"
test_command "PostgreSQL running" "pgrep -f postgres"
echo ""

echo "4. PHP-FPM Tests"
echo "================"
test_command "PHP-FPM config test" "php-fpm8.3 -t"
test_command "PHP-FPM pool.d exists" "test -d /etc/php/8.3/fpm/pool.d"
test_command "Default www pool" "test -f /etc/php/8.3/fpm/pool.d/www.conf"
echo ""

echo "5. Supervisor Tests"
echo "==================="
test_command "Supervisor config" "test -f /etc/supervisor/supervisord.conf"
test_command "Supervisor programs dir" "test -d /etc/supervisor/conf.d"
test_command "Supervisorctl" "supervisorctl version"
echo ""

echo "6. System Tests"
echo "==============="
test_command "Systemd available" "systemctl --version"
test_command "Root privileges" "test $(id -u) -eq 0"
test_command "Network connectivity" "ping -c 1 8.8.8.8"
echo ""

echo "=== Test Summary ==="
echo -e "Passed: ${GREEN}${TESTS_PASSED}${NC}"
echo -e "Failed: ${RED}${TESTS_FAILED}${NC}"
echo -e "Total:  $((TESTS_PASSED + TESTS_FAILED))"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed! ✓${NC}"
    exit 0
else
    echo -e "${YELLOW}Some tests failed. Review above.${NC}"
    exit 1
fi
```

Run the test:
```bash
chmod +x comprehensive-test.sh
sudo ./comprehensive-test.sh
```

---

## Performance Testing

### Load Testing MySQL

```bash
# Install sysbench
apt install sysbench -y

# Prepare test
sysbench /usr/share/sysbench/oltp_read_write.lua \
  --mysql-user=root \
  --mysql-password=yourpassword \
  --mysql-db=testdb \
  --tables=10 \
  --table-size=10000 \
  prepare

# Run benchmark
sysbench /usr/share/sysbench/oltp_read_write.lua \
  --mysql-user=root \
  --mysql-password=yourpassword \
  --mysql-db=testdb \
  --tables=10 \
  --table-size=10000 \
  --threads=10 \
  --time=60 \
  run

# Cleanup
sysbench /usr/share/sysbench/oltp_read_write.lua \
  --mysql-user=root \
  --mysql-password=yourpassword \
  --mysql-db=testdb \
  cleanup
```

### Load Testing PHP-FPM

```bash
# Install Apache Bench
apt install apache2-utils -y

# Create test PHP file
cat > /var/www/html/test.php << 'EOF'
<?php
echo "Hello from PHP-FPM!";
phpinfo();
EOF

# Configure nginx
cat > /etc/nginx/sites-available/default << 'EOF'
server {
    listen 80;
    root /var/www/html;
    index index.php;
    
    location ~ \.php$ {
        fastcgi_pass unix:/run/php/php8.3-fpm.sock;
        fastcgi_index index.php;
        include fastcgi_params;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
    }
}
EOF

systemctl reload nginx

# Run load test
ab -n 10000 -c 100 http://localhost/test.php
```

---

## Validation Checklist

### Pre-Testing
- [ ] Server is AMD64/Intel architecture
- [ ] Ubuntu 24.04 LTS installed
- [ ] All services installed and running
- [ ] Root/sudo access available
- [ ] Network connectivity working

### MySQL Testing
- [ ] Can view configuration
- [ ] Root password changes successfully
- [ ] Port changes and service restarts
- [ ] Database creation works
- [ ] Database listing shows correct data
- [ ] User privileges granted correctly
- [ ] Service status displays properly

### PostgreSQL Testing
- [ ] Can view configuration
- [ ] Postgres password changes successfully
- [ ] Port changes and service restarts
- [ ] Max connections updates correctly
- [ ] Shared buffers updates correctly
- [ ] Database creation works
- [ ] User privileges granted correctly
- [ ] Service status displays properly

### PHP-FPM Testing
- [ ] Pools list correctly
- [ ] New pool creation succeeds
- [ ] Pool configuration is valid
- [ ] Pool editing works
- [ ] Service reload is graceful
- [ ] Service restart works
- [ ] Cannot delete default 'www' pool
- [ ] Custom pools can be deleted
- [ ] Service status shows all pools

### Supervisor Testing
- [ ] Programs list with correct states
- [ ] New program creation works
- [ ] Program starts successfully
- [ ] Program stops successfully
- [ ] Program restarts successfully
- [ ] Program editing saves changes
- [ ] AutoStart toggle works
- [ ] XML-RPC configuration saves
- [ ] XML-RPC web interface accessible
- [ ] Programs can be controlled via web
- [ ] Program deletion works
- [ ] Logs are created properly

---

## Troubleshooting

### Build Issues

```bash
# Verify Go installation
go version

# Clean build cache
go clean -cache -modcache -i -r

# Rebuild
go build -v -o ravact ./cmd/ravact

# Check binary
file ravact
# Should show: ELF 64-bit LSB executable, x86-64
```

### Service Issues

```bash
# Check service status
systemctl status mysql
systemctl status postgresql
systemctl status php8.3-fpm
systemctl status supervisor

# View logs
journalctl -u mysql -n 50
journalctl -u postgresql -n 50
journalctl -u php8.3-fpm -n 50
journalctl -u supervisor -n 50

# Restart services
systemctl restart mysql
systemctl restart postgresql
systemctl restart php8.3-fpm
systemctl restart supervisor
```

### Permission Issues

```bash
# Ensure running as root
sudo ./ravact

# Check file ownership
ls -la /etc/mysql/
ls -la /etc/postgresql/
ls -la /etc/php/8.3/
ls -la /etc/supervisor/

# Fix permissions if needed
chmod 644 /etc/mysql/mysql.conf.d/mysqld.cnf
chmod 644 /etc/postgresql/*/main/postgresql.conf
```

---

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Test on AMD64

on: [push, pull_request]

jobs:
  test-amd64:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      
      - name: Install services
        run: |
          sudo apt update
          sudo apt install -y mysql-server postgresql php8.3-fpm supervisor
      
      - name: Build Ravact
        run: go build -v -o ravact ./cmd/ravact
      
      - name: Run tests
        run: |
          chmod +x comprehensive-test.sh
          sudo ./comprehensive-test.sh
```

---

## Cleanup

### Remove Test Data

```bash
# Drop test databases
mysql -u root -pYOURPASS -e "DROP DATABASE IF EXISTS testdb;"
sudo -u postgres psql -c "DROP DATABASE IF EXISTS testpgdb;"

# Remove test PHP-FPM pool
rm /etc/php/8.3/fpm/pool.d/testpool.conf
systemctl reload php8.3-fpm

# Remove test Supervisor program
rm /etc/supervisor/conf.d/testapp.conf
supervisorctl reread
supervisorctl update

# Reset configurations to defaults
```

### Destroy VPS

```bash
# DigitalOcean example
doctl compute droplet delete ravact-test

# Or via web console
```

---

## Best Practices

1. **Test on Clean System**: Start with fresh Ubuntu installation
2. **Document Issues**: Note any problems encountered
3. **Performance Test**: Run benchmarks to verify performance
4. **Security Test**: Verify secure password handling
5. **Error Handling**: Test with invalid inputs
6. **Service Recovery**: Test service restart scenarios
7. **Concurrent Access**: Test multiple simultaneous operations
8. **Resource Limits**: Test with limited resources

---

## Reporting Test Results

Create a test report:

```markdown
# Ravact AMD64 Test Report

**Date**: 2026-01-23
**Tester**: Your Name
**Hardware**: DigitalOcean Droplet (4 CPU, 8GB RAM)
**OS**: Ubuntu 24.04 LTS (AMD64)

## Test Results

### MySQL Management: PASS ✓
- View Configuration: PASS
- Change Root Password: PASS
- Change Port: PASS
- Create Database: PASS
- List Databases: PASS

### PostgreSQL Management: PASS ✓
- View Configuration: PASS
- Change Password: PASS
- Change Port: PASS
- Update Max Connections: PASS
- Update Shared Buffers: PASS
- Create Database: PASS

### PHP-FPM Management: PASS ✓
- List Pools: PASS
- Create Pool: PASS
- Edit Pool: PASS
- Delete Pool: PASS
- Service Control: PASS

### Supervisor Management: PASS ✓
- List Programs: PASS
- Add Program: PASS
- Start/Stop/Restart: PASS
- Edit Program: PASS
- Configure XML-RPC: PASS
- Delete Program: PASS

## Issues Found
None

## Recommendations
All features working as expected on AMD64.
```

---

## Next Steps

After successful AMD64 testing:
1. ✅ Compare results with ARM64 testing
2. ✅ Document any architecture-specific issues
3. ✅ Create release builds for both architectures
4. ✅ Update documentation with findings
5. ✅ Push commits to repository

---

## Quick Reference

```bash
# Build for AMD64
go build -o ravact ./cmd/ravact

# Run comprehensive tests
sudo ./comprehensive-test.sh

# Check architecture
uname -m  # Should show: x86_64

# Verify binary
file ravact  # Should show: x86-64
```
