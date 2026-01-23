# Scripts Directory

Automation scripts for ravact development and testing.

## 🚀 Quick Start

### **Option 1: ARM64 VM (Recommended)** ⭐
Native speed, most Linux servers use ARM64 now.

```bash
cd ravact-go/scripts
./setup-multipass.sh

# Daily workflow
./vm-manager.sh sync arm64
./vm-manager.sh run arm64
```

### **Option 2: AMD64 Docker (For x86_64 Testing)**
Quick compatibility testing via Docker.

```bash
cd ravact-go/scripts
./docker-build-and-test.sh
# Builds and tests in x86_64 container automatically!
```

---

## 📁 Scripts Overview

### **ARM64 VM Scripts** (Multipass)

| Script | Purpose |
|--------|---------|
| **setup-multipass.sh** | ⭐ Setup ARM64 Ubuntu VM |
| **vm-manager.sh** | Manage VM (start/stop/sync/run) |
| **quick-deploy.sh** | Quick deploy to specific VM |
| **setup-vm-only.sh** | Configure existing VM |

### **AMD64 Docker Scripts** (x86_64 Testing)

| Script | Purpose |
|--------|---------|
| **test-docker-amd64.sh** | Quick one-off x86_64 test |
| **docker-amd64-dev.sh** | Persistent dev container |
| **docker-build-and-test.sh** | ⭐ Automated build & test |
| **docker-manager.sh** | Manage Docker container |

### **Legacy/Alternative**

| Script | Purpose |
|--------|---------|
| **setup-mac-vm.sh** | UTM VM setup (if Multipass doesn't work) |

---

## 🎯 Common Commands

### **ARM64 VM (Multipass)**

```bash
# Setup (one-time)
./setup-multipass.sh

# Manage VM
./vm-manager.sh start arm64      # Start VM
./vm-manager.sh stop arm64       # Stop VM
./vm-manager.sh shell arm64      # Open shell
./vm-manager.sh sync arm64       # Deploy code
./vm-manager.sh run arm64        # Run ravact

# Quick deploy
./quick-deploy.sh                # Fastest update
```

### **AMD64 Docker (x86_64)**

```bash
# Quick test (one command!)
./docker-build-and-test.sh       # Builds & tests automatically

# Or manual workflow
./docker-amd64-dev.sh            # Start dev container
./docker-manager.sh shell        # Connect to container
./docker-manager.sh stop         # Stop container
```

---

## 📖 Documentation

| File | Description |
|------|-------------|
| **VM_SETUP_README.md** | Complete VM setup guide |
| **MULTIPASS_GUIDE.md** | Multipass usage guide |
| **UTM_TROUBLESHOOTING.md** | Fix UTM boot issues |

---

## 🏗️ Architecture Testing

### **ARM64 (Native)** - Primary Development ⭐
```bash
./setup-multipass.sh
./vm-manager.sh run arm64
```
- ⚡⚡⚡ Native M1/M2 speed
- Most cloud servers are ARM64
- Full features

### **AMD64 (Emulated)** - Compatibility Testing
```bash
./docker-build-and-test.sh
```
- ⚡ Slower (QEMU emulation via Docker)
- Tests x86_64 compatibility
- Quick and easy

---

## 🔄 Daily Workflows

### **ARM64 Development** (Recommended)
```bash
# 1. Edit code on Mac
vim ../internal/ui/screens/main_menu.go

# 2. Deploy to VM
./vm-manager.sh sync arm64
# Or: ./quick-deploy.sh (faster)

# 3. Test on VM
./vm-manager.sh run arm64
```

### **AMD64 Compatibility Testing**
```bash
# 1. Edit code on Mac
vim ../internal/system/users.go

# 2. Build & test x86_64 (one command!)
./docker-build-and-test.sh
```

### **Test Both Architectures**
```bash
# Terminal 1: ARM64
./vm-manager.sh run arm64

# Terminal 2: AMD64
./docker-build-and-test.sh

# Compare results!
```

---

## 🎨 VS Code Integration

### **ARM64 VM** (Remote SSH)
```bash
# After setup, connect via Remote-SSH extension
# Host: ravact-dev
# Folder: /home/ubuntu/ravact-go
```

### **Docker** (Dev Containers)
```bash
# Install "Dev Containers" extension
# Connect to running container
# Or edit on Mac (changes sync automatically via volume mount)
```

---

## 🔧 Troubleshooting

### **Multipass Issues**
```bash
# Multipass command not found
sudo ln -sf '//Library/Application Support/com.canonical.multipass/bin/multipass' /usr/local/bin/multipass
export PATH="/usr/local/bin:$PATH"

# VM won't start
multipass list
multipass logs ravact-dev

# Recreate VM
multipass delete ravact-dev && multipass purge
./setup-multipass.sh
```

### **Docker Issues**
```bash
# Docker not running
open -a Docker

# Container issues
./docker-manager.sh status
./docker-manager.sh recreate

# Volume mount not working
docker volume ls
./docker-manager.sh recreate
```

### **General Issues**
```bash
# Script permission denied
chmod +x *.sh

# Start fresh
./vm-manager.sh clean        # Remove VMs
./docker-manager.sh delete   # Remove containers
```

---

## ✅ Quick Reference

### **ARM64 VM (Primary)**
```bash
# Setup
./setup-multipass.sh

# Deploy & Test
./quick-deploy.sh                # Fastest!
./vm-manager.sh sync arm64       # Or use this
./vm-manager.sh run arm64        # Run ravact

# Control
./vm-manager.sh start arm64
./vm-manager.sh stop arm64
./vm-manager.sh shell arm64

# Direct access
ssh ravact-dev
multipass shell ravact-dev
```

### **AMD64 Docker (x86_64 Testing)**
```bash
# Build & Test (one command!)
./docker-build-and-test.sh      # ⭐ Easiest!

# Or manual
./docker-amd64-dev.sh            # Start container
./docker-manager.sh shell        # Connect
./docker-manager.sh stop         # Stop
```

---

## 📊 Script Summary

| Script | Purpose | Use When |
|--------|---------|----------|
| **setup-multipass.sh** | Setup ARM64 VM | First time |
| **vm-manager.sh** | Manage VMs | Daily control |
| **quick-deploy.sh** | Fast deploy | ⭐ Daily testing |
| **docker-build-and-test.sh** | x86_64 test | ⭐ Compatibility |
| **docker-amd64-dev.sh** | Dev container | Extended testing |

---

## 💡 Best Practices

1. **Use ARM64 for daily work** - Native speed, modern standard
2. **Test AMD64 before release** - Ensure x86_64 compatibility  
3. **Use quick-deploy.sh** - Fastest ARM64 iteration
4. **Use docker-build-and-test.sh** - Easiest AMD64 testing
5. **Keep containers running** - Reconnect without rebuild

---

**Happy developing! 🚀**
