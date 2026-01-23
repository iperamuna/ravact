# Testing Ravact on M1/M2 Macs with Multipass

This guide explains how to test Ravact on Apple Silicon (M1/M2/M3) Macs using Multipass VMs running Ubuntu ARM64.

---

## Why Multipass for M1 Macs?

**Ravact** is designed for **Linux systems only** - it directly manages system services like nginx, MySQL, PHP-FPM, and Supervisor which are Linux-specific. macOS doesn't use systemd, has different package managers, and different service management.

**Solution**: Use Multipass to create Ubuntu ARM64 VMs on your M1 Mac.

### Benefits:
- ✅ Native ARM64 Ubuntu VMs on Apple Silicon
- ✅ Fast and lightweight
- ✅ Easy VM management
- ✅ Share folders between Mac and VM
- ✅ Test real Linux environment

---

## Prerequisites

- macOS 11.0 or later (Big Sur+)
- Apple Silicon Mac (M1, M2, M3)
- At least 8GB RAM (16GB recommended)
- 20GB free disk space
- Internet connection

---

## Installation

### 1. Install Multipass

**Option A: Using Homebrew (Recommended)**
```bash
brew install --cask multipass
```

**Option B: Download Installer**
- Visit: https://multipass.run/
- Download macOS installer
- Install the .pkg file

### 2. Verify Installation
```bash
multipass version
```

---

## Setup Ubuntu VM for Testing

### Quick Start Script

Use the provided setup script:

```bash
# From Ravact repository root
cd scripts
chmod +x setup-multipass.sh
./setup-multipass.sh
```

This creates a VM named `ravact-test` with:
- Ubuntu 24.04 LTS (ARM64)
- 4 GB RAM
- 2 CPU cores
- 20 GB disk

### Manual Setup

```bash
# Create VM
multipass launch 24.04 \
  --name ravact-test \
  --memory 4G \
  --cpus 2 \
  --disk 20G

# Get VM info
multipass info ravact-test

# Access VM shell
multipass shell ravact-test
```

---

## Building Ravact for ARM64

### Option 1: Build on Mac, Run in VM

```bash
# On your Mac
cd /path/to/ravact

# Build for Linux ARM64
GOOS=linux GOARCH=arm64 go build -o ravact-linux-arm64 ./cmd/ravact

# Transfer to VM
multipass transfer ravact-linux-arm64 ravact-test:/home/ubuntu/ravact

# Enter VM and run
multipass shell ravact-test
chmod +x /home/ubuntu/ravact
sudo /home/ubuntu/ravact
```

### Option 2: Build Inside VM

```bash
# Enter VM
multipass shell ravact-test

# Install Go
sudo snap install go --classic

# Clone repository
git clone https://github.com/iperamuna/ravact.git
cd ravact

# Build
go build -o ravact ./cmd/ravact

# Run
sudo ./ravact
```

---

## Testing New Features

### 1. MySQL Management Testing

```bash
# Inside VM
multipass shell ravact-test

# Install MySQL
sudo apt update
sudo apt install mysql-server -y

# Run Ravact
sudo /home/ubuntu/ravact

# Navigate to:
# Main Menu → Configurations → MySQL Database

# Test features:
# ✓ View Current Configuration
# ✓ Change Root Password
# ✓ Change Port (e.g., 3307)
# ✓ Create Database
# ✓ List Databases
```

### 2. PostgreSQL Management Testing

```bash
# Install PostgreSQL
sudo apt install postgresql postgresql-contrib -y

# Run Ravact
sudo /home/ubuntu/ravact

# Navigate to:
# Main Menu → Configurations → PostgreSQL Database

# Test features:
# ✓ View Current Configuration
# ✓ Change Postgres Password
# ✓ Change Port (e.g., 5433)
# ✓ Update Max Connections
# ✓ Update Shared Buffers
# ✓ Create Database
```

### 3. PHP-FPM Pool Testing

```bash
# Install PHP-FPM
sudo apt install php8.3-fpm -y

# Run Ravact
sudo /home/ubuntu/ravact

# Navigate to:
# Main Menu → Configurations → PHP-FPM Pools

# Test features:
# ✓ List All Pools
# ✓ Create New Pool (name: testpool)
# ✓ View Pool Details
# ✓ Edit Pool
# ✓ Restart PHP-FPM Service
```

### 4. Supervisor Testing

```bash
# Install Supervisor
sudo apt install supervisor -y

# Run Ravact
sudo /home/ubuntu/ravact

# Navigate to:
# Main Menu → Configurations → Supervisor

# Test features:
# ✓ List All Programs
# ✓ Add New Program
# ✓ Start/Stop/Restart Program
# ✓ Configure XML-RPC
# ✓ Edit Program
```

---

## VM Management Commands

### Basic Operations

```bash
# List all VMs
multipass list

# Start VM
multipass start ravact-test

# Stop VM
multipass stop ravact-test

# Restart VM
multipass restart ravact-test

# Delete VM
multipass delete ravact-test
multipass purge

# Get VM IP
multipass info ravact-test | grep IPv4
```

### File Transfer

```bash
# Copy file to VM
multipass transfer myfile.txt ravact-test:/home/ubuntu/

# Copy from VM to Mac
multipass transfer ravact-test:/home/ubuntu/myfile.txt .

# Mount Mac folder in VM
multipass mount /path/on/mac ravact-test:/home/ubuntu/shared
```

### VM Shell Access

```bash
# Interactive shell
multipass shell ravact-test

# Run single command
multipass exec ravact-test -- ls -la

# Run with sudo
multipass exec ravact-test -- sudo systemctl status mysql
```

---

## Complete Test Workflow

### Step-by-Step Testing

```bash
# 1. Create and setup VM
multipass launch 24.04 --name ravact-test --memory 4G --cpus 2 --disk 20G

# 2. Update system
multipass exec ravact-test -- sudo apt update
multipass exec ravact-test -- sudo apt upgrade -y

# 3. Install test dependencies
multipass exec ravact-test -- sudo apt install -y \
  mysql-server \
  postgresql \
  postgresql-contrib \
  php8.3-fpm \
  supervisor \
  nginx

# 4. Build and transfer Ravact
GOOS=linux GOARCH=arm64 go build -o ravact-linux-arm64 ./cmd/ravact
multipass transfer ravact-linux-arm64 ravact-test:/home/ubuntu/ravact
multipass exec ravact-test -- chmod +x /home/ubuntu/ravact

# 5. Run Ravact
multipass shell ravact-test
sudo /home/ubuntu/ravact
```

### Verification Tests

```bash
# Inside VM - verify services
sudo systemctl status mysql
sudo systemctl status postgresql
sudo systemctl status php8.3-fpm
sudo systemctl status supervisor

# Check MySQL connection
mysql -u root -e "SHOW DATABASES;"

# Check PostgreSQL
sudo -u postgres psql -c "\l"

# Check PHP-FPM pools
sudo php-fpm8.3 -t
ls -la /etc/php/8.3/fpm/pool.d/

# Check Supervisor programs
sudo supervisorctl status
```

---

## Automated Testing Script

Create a test script inside the VM:

```bash
#!/bin/bash
# test-ravact-features.sh

echo "Testing Ravact Features..."

# Test MySQL
echo "1. Testing MySQL..."
sudo systemctl status mysql >/dev/null 2>&1 && echo "✓ MySQL running" || echo "✗ MySQL not running"

# Test PostgreSQL
echo "2. Testing PostgreSQL..."
sudo systemctl status postgresql >/dev/null 2>&1 && echo "✓ PostgreSQL running" || echo "✗ PostgreSQL not running"

# Test PHP-FPM
echo "3. Testing PHP-FPM..."
sudo systemctl status php8.3-fpm >/dev/null 2>&1 && echo "✓ PHP-FPM running" || echo "✗ PHP-FPM not running"

# Test Supervisor
echo "4. Testing Supervisor..."
sudo systemctl status supervisor >/dev/null 2>&1 && echo "✓ Supervisor running" || echo "✗ Supervisor not running"

# Test configurations
echo "5. Testing configurations..."
[ -f /etc/mysql/mysql.conf.d/mysqld.cnf ] && echo "✓ MySQL config exists"
[ -f /etc/postgresql/*/main/postgresql.conf ] && echo "✓ PostgreSQL config exists"
[ -d /etc/php/8.3/fpm/pool.d ] && echo "✓ PHP-FPM pool.d exists"
[ -f /etc/supervisor/supervisord.conf ] && echo "✓ Supervisor config exists"

echo "Testing complete!"
```

Run it:
```bash
chmod +x test-ravact-features.sh
./test-ravact-features.sh
```

---

## Performance Notes

### ARM64 on M1 Macs
- **Excellent performance**: Native ARM64 execution
- **Fast compilation**: Go builds quickly on ARM
- **Responsive VMs**: Multipass VMs run smoothly
- **Low overhead**: Better than x86 emulation

### Resource Recommendations
- **Minimal**: 2GB RAM, 1 CPU, 10GB disk
- **Recommended**: 4GB RAM, 2 CPUs, 20GB disk
- **Heavy testing**: 8GB RAM, 4 CPUs, 40GB disk

---

## Troubleshooting

### Issue: VM won't start
```bash
# Check Multipass status
multipass version

# Restart Multipass
sudo launchctl stop com.canonical.multipassd
sudo launchctl start com.canonical.multipassd

# Check logs
tail -f ~/Library/Logs/Multipass/multipassd.log
```

### Issue: Can't connect to VM
```bash
# Get VM status
multipass info ravact-test

# Restart VM
multipass restart ravact-test

# Check network
multipass exec ravact-test -- ip addr
```

### Issue: Build fails
```bash
# Ensure correct GOOS/GOARCH
echo $GOOS $GOARCH

# Clean and rebuild
go clean
GOOS=linux GOARCH=arm64 go build -v -o ravact-linux-arm64 ./cmd/ravact
```

### Issue: Permission denied in VM
```bash
# Always use sudo for Ravact
sudo /home/ubuntu/ravact

# Check file permissions
ls -la /home/ubuntu/ravact
chmod +x /home/ubuntu/ravact
```

---

## Cleanup

### Remove test VM
```bash
# Stop and delete
multipass stop ravact-test
multipass delete ravact-test
multipass purge

# Verify removal
multipass list
```

### Uninstall Multipass
```bash
# Using Homebrew
brew uninstall --cask multipass

# Manual: Delete app from Applications folder
```

---

## Next Steps

After testing on Multipass:
1. ✅ Verify all features work on ARM64
2. ✅ Test on real AMD64/Intel hardware (see AMD64_TESTING.md)
3. ✅ Run integration tests
4. ✅ Create release builds for both architectures

---

## Additional Resources

- **Multipass Documentation**: https://multipass.run/docs
- **Ubuntu Cloud Images**: https://cloud-images.ubuntu.com/
- **Ravact Repository**: https://github.com/iperamuna/ravact
- **Go Cross-Compilation**: https://go.dev/doc/install/source#environment

---

## Quick Reference

```bash
# Essential Commands
multipass launch 24.04 --name ravact-test --memory 4G --cpus 2 --disk 20G
multipass shell ravact-test
multipass transfer file.txt ravact-test:/home/ubuntu/
multipass exec ravact-test -- command
multipass stop ravact-test
multipass delete ravact-test && multipass purge

# Build for ARM64 Linux
GOOS=linux GOARCH=arm64 go build -o ravact-linux-arm64 ./cmd/ravact
```
