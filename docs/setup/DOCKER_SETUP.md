# Docker Setup for x86_64 Testing

## 🚀 Quick Install

### **Option 1: Homebrew (Running in Background)**
```bash
# The installation is currently running
# Wait for it to complete, then:
open -a Docker
```

### **Option 2: Manual Download (Faster)**
1. Download: https://desktop.docker.com/mac/main/arm64/Docker.dmg
2. Open the DMG file
3. Drag Docker to Applications
4. Open Docker from Applications
5. Wait for Docker to start (whale icon in menu bar)

---

## ✅ Verify Docker is Running

```bash
# Check if Docker is installed
docker --version

# Should show something like:
# Docker version 24.x.x, build xxxxx

# Check if Docker daemon is running
docker ps

# Should show empty table if working:
# CONTAINER ID   IMAGE   COMMAND   CREATED   STATUS   PORTS   NAMES
```

---

## 🧪 Test x86_64 Container

Once Docker is running:

```bash
cd ravact-go/scripts
./test-docker-amd64.sh
```

**Or test manually:**
```bash
# Quick test
docker run --rm --platform linux/amd64 ubuntu:24.04 uname -m
# Should print: x86_64
```

---

## ⏱️ Installation Status

### **Current Status:**
- Homebrew installation is running in background
- This can take 5-10 minutes (Docker Desktop is ~600MB)
- You'll see progress in your terminal

### **What's Happening:**
1. ⏳ Downloading Docker Desktop (~600MB)
2. ⏳ Installing to /Applications
3. ⏳ Setting up Docker daemon
4. ✅ Ready to use!

---

## 🔄 While Waiting

You can continue with other tasks:

### **Option A: Test on ARM64 VM (Works Now)**
```bash
# ARM64 testing works perfectly
cd ravact-go/scripts
./setup-multipass.sh

# Or if already set up:
./vm-manager.sh run arm64
```

### **Option B: Build and Prepare**
```bash
# Build AMD64 binary (ready for Docker)
cd ravact-go
make build-linux

# Binary will be ready at: dist/ravact-linux-amd64
```

### **Option C: Read Documentation**
```bash
cat REAL_AMD64_TESTING.md
```

---

## 🆘 Troubleshooting

### **Installation Takes Too Long**
```bash
# Cancel and use manual download instead
# Download from: https://desktop.docker.com/mac/main/arm64/Docker.dmg
```

### **Check Installation Progress**
```bash
# See if Docker is installed
ls -la /Applications/Docker.app

# Check brew installation status
brew list --cask | grep docker
```

### **Docker Won't Start**
1. Open Docker Desktop from Applications
2. Grant permissions if asked
3. Wait for whale icon in menu bar to be stable
4. Check: System Preferences → Privacy → Full Disk Access → Docker

---

## ✅ After Installation

### **Step 1: Start Docker Desktop**
```bash
open -a Docker
```

Wait for the whale icon in the menu bar to stop animating (means Docker is ready).

### **Step 2: Verify It Works**
```bash
docker --version
docker ps
```

### **Step 3: Test x86_64**
```bash
cd ravact-go/scripts
./test-docker-amd64.sh
```

You should see:
```
========================================
Ravact AMD64 Docker Testing
x86_64 via QEMU Emulation
========================================

Building ravact for AMD64...
✓ Build complete

Starting x86_64 Ubuntu container...

=====================================
Ubuntu x86_64 Container
=====================================

Architecture: x86_64
OS: Ubuntu 24.04.3 LTS

✓ Container ready!

You can now run:
  sudo ./dist/ravact-linux-amd64

root@containerid:/workspace#
```

### **Step 4: Test ravact**
```bash
# Inside the Docker container
sudo ./dist/ravact-linux-amd64

# Navigate to User Management (press 2)
# Test all features!

# Exit when done
exit
```

---

## 🎯 Quick Commands

```bash
# Start Docker Desktop
open -a Docker

# Test x86_64
cd ravact-go/scripts
./test-docker-amd64.sh

# Inside container
sudo ./dist/ravact-linux-amd64

# Exit container
exit
```

---

## 📊 Docker vs Other Methods

| Method | Speed | Startup | Size | Persistent |
|--------|-------|---------|------|------------|
| Docker | ⚡ OK | ~2 sec | 600MB | ❌ Temporary |
| Multipass | ⚡⚡⚡ Fast | ~10 sec | 500MB | ✅ Persistent |
| UTM | 🐌 Slow | ~30 sec | 2GB | ✅ Persistent |
| VPS | ⚡⚡⚡ Fast | Always on | N/A | ✅ Persistent |

**Docker is perfect for:**
- Quick compatibility testing ✅
- CI/CD pipelines ✅
- One-off tests ✅
- No VM overhead ✅

**Not ideal for:**
- Long-running services ⚠️
- Persistent testing ⚠️

---

## 💡 Pro Tips

### **Faster Workflow**
```bash
# Make changes
vim internal/ui/screens/main_menu.go

# Build
make build-linux

# Test (container auto-mounts code)
./scripts/test-docker-amd64.sh
# Changes are immediately available!
```

### **Keep Container Running**
```bash
# In one terminal, keep container alive
docker run -it --rm --platform linux/amd64 \
  -v $(pwd):/workspace -w /workspace \
  ubuntu:24.04 bash

# In another terminal, rebuild and test
make build-linux
# Switch to container terminal and test
```

### **Pre-pull Image**
```bash
# Download Ubuntu image ahead of time
docker pull --platform linux/amd64 ubuntu:24.04
```

---

## 🔄 Alternative: Manual Testing

If Docker takes too long, you can also:

### **Option 1: Use ARM64 Only**
Most deployments are ARM64 anyway!
```bash
./scripts/setup-multipass.sh
./scripts/vm-manager.sh run arm64
```

### **Option 2: GitHub Actions (Free)**
Add to `.github/workflows/test.yml`:
```yaml
jobs:
  test-amd64:
    runs-on: ubuntu-latest  # x86_64
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: make build-linux
      - run: ./dist/ravact-linux-amd64 --version
```

### **Option 3: Cloud VPS**
Deploy to $4/month x86_64 VPS:
```bash
./scripts/quick-deploy.sh <vps-ip> root
```

---

## ✅ Summary

**Current Status:**
- Docker installation running in background
- Will take 5-10 minutes
- Can use ARM64 VM in the meantime

**Once Docker is ready:**
```bash
open -a Docker
cd ravact-go/scripts
./test-docker-amd64.sh
```

**Alternative while waiting:**
```bash
./scripts/setup-multipass.sh  # ARM64 VM
```

---

**Check back in a few minutes for Docker to finish installing!** ⏳
