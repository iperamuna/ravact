# Real AMD64/x86_64 Testing on M1 Mac

## ❌ Issue: Multipass Can't Create x86_64 VMs

**Problem discovered:** Multipass on Apple Silicon (M1/M2) **cannot create x86_64/AMD64 VMs**. It only creates ARM64 VMs natively.

**What happened:**
- Script tried to install AMD64 Go on ARM64 VM
- Result: "cannot execute binary file: Exec format error"
- VM architecture was ARM64, not AMD64

## ✅ Real Solutions for x86_64 Testing

### **Solution 1: Docker with Platform Emulation** ⭐ (Easiest)

Docker Desktop on M1 can run x86_64 containers via QEMU emulation.

#### **Quick Test:**
```bash
cd ravact-go/scripts
./test-docker-amd64.sh
```

**What it does:**
1. Builds ravact for AMD64
2. Starts x86_64 Ubuntu container (via QEMU)
3. Drops you into container
4. Run: `sudo ./dist/ravact-linux-amd64`

**Pros:**
- ✅ Real x86_64 environment
- ✅ Quick to start/stop
- ✅ No VM overhead
- ✅ Easy cleanup

**Cons:**
- ⚠️ Slower (QEMU emulation)
- ⚠️ Requires Docker Desktop

---

### **Solution 2: Cloud x86_64 Server** (Best for Production Testing)

Deploy to a real x86_64 server for accurate testing.

#### **Options:**

**AWS EC2 (Free Tier):**
```bash
# 1. Launch t2.micro Ubuntu instance (x86_64)
# 2. SSH and deploy
scp dist/ravact-linux-amd64 ec2-user@<server>:~/ravact
ssh ec2-user@<server>
sudo ./ravact
```

**DigitalOcean Droplet:**
```bash
# $4/month cheapest x86_64 droplet
# Use quick-deploy.sh
./scripts/quick-deploy.sh <droplet-ip> root
```

**Hetzner Cloud:**
```bash
# Cheapest: €3.29/month x86_64
# Fast deployment
```

**Pros:**
- ✅ Real x86_64 hardware
- ✅ Production-like environment
- ✅ Native performance
- ✅ Can test networking, firewall, etc.

**Cons:**
- 💰 Costs money (but cheap)
- 🌐 Requires internet

---

### **Solution 3: UTM with x86_64 Emulation** (Slowest)

UTM on M1 can emulate x86_64, but it's very slow.

#### **Setup:**
1. Download UTM: https://mac.getutm.app/
2. Create VM:
   - Type: **Emulate** (not Virtualize!)
   - Architecture: x86_64
   - ISO: Ubuntu 24.04 AMD64
3. Install Ubuntu (will be slow)
4. Deploy ravact

**Pros:**
- ✅ Real x86_64 VM
- ✅ Full system control

**Cons:**
- ❌ Very slow (full emulation)
- ❌ Manual setup
- ❌ High resource usage

---

### **Solution 4: GitHub Actions** (Free CI/CD)

Test on real x86_64 runners automatically.

#### **Create `.github/workflows/test-amd64.yml`:**
```yaml
name: Test AMD64

on: [push, pull_request]

jobs:
  test-amd64:
    runs-on: ubuntu-latest  # This is x86_64!
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build
        run: make build-linux
      
      - name: Test
        run: |
          ./dist/ravact-linux-amd64 --version
          # Add more tests
```

**Pros:**
- ✅ Free
- ✅ Real x86_64
- ✅ Automated
- ✅ No local resources

**Cons:**
- ⚠️ Can't interactively test UI

---

## 📊 Comparison

| Method | Speed | Cost | Setup | Real x86_64 | Interactive |
|--------|-------|------|-------|-------------|-------------|
| **Docker** ⭐ | ⚡ OK | Free | Easy | ✅ Yes | ✅ Yes |
| **Cloud Server** | ⚡⚡⚡ Fast | $4/mo | Easy | ✅ Yes | ✅ Yes |
| **UTM Emulation** | 🐌 Slow | Free | Hard | ✅ Yes | ✅ Yes |
| **GitHub Actions** | ⚡⚡ Fast | Free | Easy | ✅ Yes | ❌ No |
| **Multipass** | ❌ N/A | N/A | N/A | ❌ No | ❌ No |

---

## 🚀 Recommended Approach

### **For Daily Testing:**
Use **Docker** - Quick and easy!
```bash
./scripts/test-docker-amd64.sh
```

### **Before Release:**
Use **Cloud Server** - Real production environment
```bash
# Deploy to cheap VPS
./scripts/quick-deploy.sh <server-ip> root
```

### **Continuous Testing:**
Use **GitHub Actions** - Automated testing on every commit

---

## 🔧 Docker Solution Details

### **What the Docker script does:**

1. **Builds AMD64 binary:**
   ```bash
   make build-linux  # Creates dist/ravact-linux-amd64
   ```

2. **Starts x86_64 container:**
   ```bash
   docker run --platform linux/amd64 ubuntu:24.04
   ```

3. **Mounts your code:**
   ```bash
   -v "$PROJECT_DIR:/workspace"
   ```

4. **Installs minimal dependencies:**
   ```bash
   apt-get install sudo
   ```

5. **Runs ravact:**
   ```bash
   sudo ./dist/ravact-linux-amd64
   ```

### **Usage:**

```bash
# Start container
./scripts/test-docker-amd64.sh

# Inside container
root@container:/workspace# uname -m
x86_64

root@container:/workspace# sudo ./dist/ravact-linux-amd64
# Test all features!

# Exit when done
root@container:/workspace# exit
```

### **Rebuild after changes:**

```bash
# On Mac: Make changes
vim internal/ui/screens/main_menu.go

# Build for AMD64
make build-linux

# Test in Docker
./scripts/test-docker-amd64.sh
# Binary automatically updated in container
```

---

## 🎯 Quick Start Guide

### **Step 1: Install Docker Desktop**
```bash
brew install --cask docker
# Start Docker Desktop app
```

### **Step 2: Test AMD64 ravact**
```bash
cd ravact-go/scripts
./test-docker-amd64.sh
```

### **Step 3: Inside Container**
```bash
# You're now in x86_64 Ubuntu!
uname -m  # Shows: x86_64

# Run ravact
sudo ./dist/ravact-linux-amd64

# Test User Management (press 2)
# Should work perfectly!

# Exit when done
exit
```

---

## 📝 Why Multipass Doesn't Work

**Technical reason:**
- Multipass on Apple Silicon uses **native virtualization** (Hypervisor.framework)
- This only supports ARM64 guests on ARM64 hosts
- x86_64 would require **full emulation** (QEMU TCG mode)
- Multipass doesn't implement QEMU emulation mode

**Alternatives that DO work:**
- **Docker**: Uses QEMU for cross-platform (built-in)
- **UTM**: Has full QEMU emulation support
- **Parallels**: Supports x86_64 emulation (paid)
- **VMware Fusion**: Supports x86_64 emulation (paid)

---

## 🔄 Updated Scripts

### **Fixed: `setup-multipass-amd64.sh`**
- Now creates second ARM64 VM (not AMD64)
- Shows clear warning about limitation
- Suggests Docker for real x86_64 testing
- Use only if you want multiple ARM64 VMs

### **New: `test-docker-amd64.sh`** ⭐
- Real x86_64 testing via Docker
- Quick and easy
- No VM overhead
- **Use this for AMD64 testing!**

### **Updated: `sync-all-vms.sh`**
- Still works for multiple ARM64 VMs
- Add Docker testing separately

---

## ✅ Summary

**What you need to know:**

1. ❌ **Multipass can't create x86_64 VMs on M1**
   - It's a technical limitation
   - Not a bug, just how it works

2. ✅ **Use Docker for x86_64 testing**
   - Quick, easy, free
   - Real x86_64 environment
   - `./scripts/test-docker-amd64.sh`

3. ✅ **Use ARM64 VM for daily work**
   - Native performance
   - Everything works
   - `./scripts/setup-multipass.sh`

4. ✅ **Use cloud server for production testing**
   - Real hardware
   - Production environment
   - Cheap ($4/month)

---

## 🎯 Action Items

### **Immediate: Use Docker**
```bash
# Install Docker Desktop
brew install --cask docker

# Test AMD64
cd ravact-go/scripts
./test-docker-amd64.sh
```

### **Optional: Clean up failed VM**
```bash
# Remove the ARM64 VM named "amd64"
multipass delete ravact-dev-amd64
multipass purge
```

### **Long-term: Set up CI/CD**
- Add GitHub Actions for automated x86_64 testing
- Deploy to cheap VPS for manual testing
- Keep ARM64 VM for daily development

---

## 💡 Best Practice Workflow

```bash
# 1. Daily development on ARM64
./scripts/setup-multipass.sh
./scripts/vm-manager.sh sync arm64
./scripts/vm-manager.sh run arm64

# 2. Quick x86_64 compatibility check
./scripts/test-docker-amd64.sh

# 3. Before release: Deploy to real x86_64 server
./scripts/quick-deploy.sh <vps-ip> root
ssh <vps-ip>
cd ravact-go && sudo ./ravact

# 4. Automated testing via GitHub Actions
git push  # Triggers x86_64 tests
```

---

**Bottom line:** Use Docker for x86_64 testing! It's the easiest and most practical solution.

```bash
./scripts/test-docker-amd64.sh
```

🚀 **That's it!**
