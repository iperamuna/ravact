# Docker AMD64 Workflow - No Sync Needed!

## 🎯 Key Concept: Volume Mounts vs Sync

**Docker containers don't need "syncing"** - they use **volume mounts**!

- ✅ Your Mac code folder is **mounted live** into the container
- ✅ Changes on Mac appear **instantly** in container
- ✅ No copy, no sync, no delay!
- ✅ Build on Mac, test in container immediately

---

## 🚀 Three Ways to Work with Docker AMD64

### **Method 1: Quick Test (One-Off)** ⭐

For quick compatibility checks:

```bash
# Run this script
cd ravact-go/scripts
./test-docker-amd64.sh

# Inside container, code is already there!
sudo ./dist/ravact-linux-amd64

# Exit when done
exit
```

**Use case:** Quick one-time test

---

### **Method 2: Development Container (Persistent)** ⭐⭐

Keep container running for iterative testing:

```bash
# Start persistent dev container
cd ravact-go/scripts
./docker-amd64-dev.sh

# Container shell opens
# Code is live-mounted at /workspace
```

**Then on your Mac (in another terminal):**
```bash
# Make changes
vim internal/ui/screens/main_menu.go

# Build
make build-linux

# Switch to container terminal
# Changes are ALREADY there!
sudo ./dist/ravact-linux-amd64
```

**Use case:** Iterative development and testing

---

### **Method 3: Build & Test Script (Easiest)** ⭐⭐⭐

Automates the whole workflow:

```bash
# One command: builds on Mac, tests in container
cd ravact-go/scripts
./docker-build-and-test.sh
```

**What it does:**
1. Builds ravact for AMD64 on Mac
2. Starts container (if needed)
3. Runs ravact in container
4. You see the app immediately!

**Use case:** Fastest workflow for daily testing

---

## 📊 Workflow Comparison

| Method | Setup Time | Best For | Persistence |
|--------|-----------|----------|-------------|
| **test-docker-amd64.sh** | 1 sec | Quick tests | ❌ One-off |
| **docker-amd64-dev.sh** | 5 sec | Development | ✅ Persistent |
| **docker-build-and-test.sh** | 1 sec | Daily work | ✅ Persistent |

---

## 🎯 Recommended Workflow

### **Daily Development:**

**Terminal 1 (Mac):**
```bash
cd ravact-go

# Make changes
vim internal/ui/screens/user_management.go

# Build for AMD64
make build-linux
```

**Terminal 2 (Docker Container):**
```bash
# Start persistent container (once)
cd ravact-go/scripts
./docker-amd64-dev.sh

# Inside container, code updates automatically!
# Just run after Mac build:
sudo ./dist/ravact-linux-amd64

# Test, then rebuild on Mac, test again!
```

**Or use the automation:**
```bash
# One command does everything!
./scripts/docker-build-and-test.sh
```

---

## 🛠️ Docker Manager Commands

I created a manager script for easy container control:

```bash
cd ravact-go/scripts

# Container lifecycle
./docker-manager.sh start         # Start container
./docker-manager.sh stop          # Stop container
./docker-manager.sh restart       # Restart container

# Access container
./docker-manager.sh shell         # Open shell
./docker-manager.sh run           # Build & test

# Information
./docker-manager.sh status        # Check status
./docker-manager.sh logs          # View logs

# Cleanup
./docker-manager.sh delete        # Remove container
./docker-manager.sh recreate      # Delete & recreate
```

---

## 💡 How Volume Mounting Works

```
Mac (Host)                  Docker Container (Guest)
─────────────────────────── ───────────────────────────
/Users/you/ravact-go   →    /workspace
    ├── cmd/                    ├── cmd/           ← Same files!
    ├── internal/               ├── internal/      ← Live updates!
    ├── dist/                   ├── dist/          ← Binary here!
    └── assets/                 └── assets/        ← All synced!
```

**When you build on Mac:**
- Binary created at `dist/ravact-linux-amd64`
- Container sees it **instantly** at `/workspace/dist/ravact-linux-amd64`
- No copy needed!

---

## 🔄 Complete Example Workflow

### **Scenario: Fix User Management Bug**

```bash
# ─────────────────────────────────────────
# TERMINAL 1: Mac Development
# ─────────────────────────────────────────

cd ravact-go

# Edit code
vim internal/system/users.go

# Build for AMD64
make build-linux
# ✓ Binary created: dist/ravact-linux-amd64


# ─────────────────────────────────────────
# TERMINAL 2: Docker Container
# ─────────────────────────────────────────

cd ravact-go/scripts
./docker-amd64-dev.sh

# Inside container:
sudo ./dist/ravact-linux-amd64
# Test the fix!

# Exit and rebuild on Mac
exit

# Reconnect
./docker-manager.sh shell

# Binary already updated! Test again
sudo ./dist/ravact-linux-amd64
```

---

## 🚀 Quick Start (Step by Step)

### **Step 1: Wait for Docker to Finish Installing**
Check if ready:
```bash
docker --version
docker ps
```

### **Step 2: Start Persistent Dev Container**
```bash
cd ravact-go/scripts
./docker-amd64-dev.sh
```

You'll see:
```
═════════════════════════════════════════
AMD64 Development Container
═════════════════════════════════════════

Architecture: x86_64
OS: Ubuntu 24.04.3 LTS

✓ Container ready!

Your code is mounted at: /workspace
Changes on Mac appear instantly here!
```

### **Step 3: Test Existing Build**
```bash
# Inside container
cd /workspace
sudo ./dist/ravact-linux-amd64
```

### **Step 4: Make Changes & Rebuild**
```bash
# On Mac (new terminal)
cd ravact-go
vim internal/ui/screens/main_menu.go
make build-linux

# Back in container
# Binary updated automatically!
sudo ./dist/ravact-linux-amd64
```

---

## 📝 Key Scripts Summary

| Script | Purpose | When to Use |
|--------|---------|-------------|
| **test-docker-amd64.sh** | One-off test | Quick compatibility check |
| **docker-amd64-dev.sh** | Dev container | Iterative development |
| **docker-build-and-test.sh** | Build & test | Automated workflow |
| **docker-manager.sh** | Container control | Start/stop/manage |

---

## 🆚 Docker vs Multipass VMs

| Feature | Docker | Multipass VM |
|---------|--------|--------------|
| **Sync Method** | Volume mount (instant) | scp/rsync (manual) |
| **Startup** | ~2 seconds | ~10 seconds |
| **Resources** | Lightweight | Heavier |
| **Persistence** | Optional | Always |
| **Use Case** | Testing, CI/CD | Development, persistent |
| **Architecture** | x86_64 (emulated) | ARM64 (native) |

---

## 💻 VS Code Integration

You can also edit directly in the container!

### **Method 1: Remote Containers Extension**
1. Install "Dev Containers" extension
2. Open container in VS Code
3. Edit directly in container

### **Method 2: Edit on Mac (Easier)**
1. Edit files on Mac with VS Code
2. Build on Mac: `make build-linux`
3. Test in container: auto-updated!

---

## 🎯 Best Practices

### **1. Keep Container Running During Dev Session**
```bash
# Morning: Start container
./scripts/docker-amd64-dev.sh

# Work throughout the day
# Make changes on Mac, test in container

# Evening: Stop container
./scripts/docker-manager.sh stop
```

### **2. Use Build & Test Script**
```bash
# Fastest iteration
./scripts/docker-build-and-test.sh
# Builds + tests in one command!
```

### **3. Clean Up When Done**
```bash
# Remove container
./scripts/docker-manager.sh delete

# Or keep it for tomorrow
./scripts/docker-manager.sh stop
```

---

## 🔧 Troubleshooting

### **"Docker is not running"**
```bash
# Start Docker Desktop
open -a Docker

# Wait for whale icon to be stable
# Then try again
```

### **"Changes not appearing in container"**
Volume mounts should work automatically, but if not:
```bash
# Recreate container
./scripts/docker-manager.sh recreate
```

### **"Binary not found"**
```bash
# Make sure to build for correct arch
make build-linux

# Check it was created
ls -lh dist/ravact-linux-amd64
```

### **"Permission denied"**
```bash
# Inside container, use sudo
sudo ./dist/ravact-linux-amd64
```

---

## ✅ Summary

**No syncing needed!** Docker uses live volume mounts:

1. ✅ **Code on Mac = Code in container** (same files!)
2. ✅ **Build on Mac** → Binary instantly in container
3. ✅ **No copy, no sync, no delay**
4. ✅ **Fastest workflow possible**

**Quick commands:**
```bash
# Start dev container
./scripts/docker-amd64-dev.sh

# Or use automated workflow
./scripts/docker-build-and-test.sh

# Manage container
./scripts/docker-manager.sh help
```

---

**Docker makes x86_64 testing easy!** 🎉

Once Docker finishes installing, you're ready to go! 🚀
