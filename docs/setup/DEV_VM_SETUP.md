# Development VM Setup Guide for M1 Mac

## Overview

This guide helps you set up an Ubuntu 24.04 ARM64 VM on your M1 MacBook Pro for testing Ravact during development. You'll be able to build on your Mac and instantly test on Linux.

## Prerequisites

- M1 MacBook Pro (ARM64 architecture)
- At least 8GB RAM (16GB+ recommended)
- 20GB+ free disk space
- Internet connection

## Option 1: UTM (Recommended for M1)

UTM is a native macOS virtualization app that works great with Apple Silicon.

### Install UTM

```bash
# Using Homebrew
brew install --cask utm

# Or download from: https://mac.getutm.app/
```

### Download Ubuntu 24.04 ARM64

```bash
# Download Ubuntu Server 24.04 ARM64
cd ~/Downloads
curl -LO https://cdimage.ubuntu.com/releases/24.04/release/ubuntu-24.04-live-server-arm64.iso
```

### Create VM in UTM

1. **Open UTM** and click "Create a New Virtual Machine"

2. **Select "Virtualize"** (not Emulate - faster on M1)

3. **Choose Linux**

4. **Configuration**:
   - Name: `Ravact-Dev-Ubuntu24`
   - Architecture: ARM64 (aarch64)
   - ISO: Browse to downloaded ubuntu-24.04-live-server-arm64.iso
   - Memory: 4096 MB (4GB) or more
   - CPU Cores: 4 cores
   - Storage: 20 GB

5. **Create and Start** the VM

### Install Ubuntu

1. Boot from ISO
2. Select "Install Ubuntu Server"
3. Follow installation:
   - Language: English
   - Keyboard: Your layout
   - Network: DHCP (automatic)
   - Storage: Use entire disk
   - Profile setup:
     - Your name: `devuser`
     - Server name: `ravact-dev`
     - Username: `devuser`
     - Password: `devuser` (or your choice)
   - **Install OpenSSH server**: YES ✓
   - Featured snaps: None needed
4. Reboot after installation
5. Remove ISO from VM settings

### Post-Install Setup

```bash
# SSH into VM (easier than UTM console)
# Get VM IP: In UTM console, run:
ip addr show | grep inet

# From Mac terminal:
ssh devuser@<VM-IP>

# Update system
sudo apt update && sudo apt upgrade -y

# Install essential tools
sudo apt install -y \
    build-essential \
    curl \
    wget \
    git \
    vim \
    htop \
    net-tools

# Install Go 1.21
cd /tmp
wget https://go.dev/dl/go1.21.0.linux-arm64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-arm64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
source ~/.bashrc

# Verify Go installation
go version
```

## Option 2: Parallels Desktop (Commercial, Best Performance)

Parallels has excellent M1 support and best performance.

### Install Parallels

```bash
# Download from: https://www.parallels.com/
# 14-day free trial available
```

### Create Ubuntu VM

1. **File → New**
2. **Install Windows or another OS from DVD or image file**
3. **Choose ubuntu-24.04-live-server-arm64.iso**
4. **Follow wizard**:
   - Name: Ravact-Dev
   - Location: Default
   - Before Installation: Customize settings
     - CPU: 4 cores
     - Memory: 4 GB
     - Hard Disk: 20 GB
5. **Continue and Install Ubuntu** (same as UTM above)

### Parallels Advantages

- Shared folders (easy file transfer)
- Better networking
- Faster graphics
- Better integration with macOS

## Option 3: Docker (Fastest for Testing)

For quick testing without full VM.

### Setup

Already configured! Use the existing Docker setup:

```bash
cd ravact-go

# Build for ARM64 Linux
make build-linux-arm64

# Test in Docker
make docker-shell

# Inside container:
cd /workspace
./dist/ravact-linux-arm64 --version
```

### Pros & Cons

**Pros:**
- ✅ Fast startup
- ✅ Already configured
- ✅ Lightweight
- ✅ Easy to reset

**Cons:**
- ❌ No persistent data
- ❌ Limited systemd support
- ❌ Can't test full installations

## Development Workflow

### Method 1: Build on Mac → Copy to VM → Test

**Setup shared folder (UTM):**

1. In UTM, click VM settings
2. Add "Shared Directory"
3. Choose your ravact-go folder
4. In VM, mount shared folder:

```bash
# Create mount point
sudo mkdir -p /mnt/ravact

# Install required tools
sudo apt install -y cifs-utils

# Mount (UTM usually auto-mounts to /media/share or similar)
# Or use scp to copy files
```

**Workflow:**

```bash
# On Mac (build for ARM64 Linux)
cd ravact-go
make build-linux-arm64

# Copy to VM (replace with your VM IP)
scp dist/ravact-linux-arm64 devuser@<VM-IP>:~/ravact

# On VM (via SSH)
ssh devuser@<VM-IP>
chmod +x ravact
sudo ./ravact
```

### Method 2: Git Repository (Recommended)

**Setup:**

```bash
# On Mac: Commit and push your changes
cd ravact-go
git add .
git commit -m "Update ravact"
git push

# On VM: Pull and build
ssh devuser@<VM-IP>
cd ravact-go
git pull
make build
sudo ./ravact
```

### Method 3: Rsync (Fast Sync)

**Setup:**

```bash
# Create rsync script on Mac
cat > sync-to-vm.sh << 'EOF'
#!/bin/bash
VM_IP="192.168.64.5"  # Change to your VM IP
VM_USER="devuser"

rsync -avz --exclude 'dist' --exclude '.git' \
    ~/GoApps/ravact/ravact-go/ \
    ${VM_USER}@${VM_IP}:~/ravact-go/

echo "✓ Synced to VM"
EOF

chmod +x sync-to-vm.sh

# Run whenever you make changes
./sync-to-vm.sh

# Then SSH and build on VM
ssh devuser@<VM-IP>
cd ravact-go
make build
sudo ./ravact
```

### Method 4: VS Code Remote SSH (Best Developer Experience)

**Setup:**

1. **Install VS Code extension**: "Remote - SSH"

2. **Configure SSH**:
```bash
# On Mac
nano ~/.ssh/config

# Add:
Host ravact-dev
    HostName 192.168.64.5  # Your VM IP
    User devuser
    ForwardAgent yes
```

3. **Connect in VS Code**:
   - Press F1 → "Remote-SSH: Connect to Host"
   - Select "ravact-dev"
   - Open folder: `/home/devuser/ravact-go`

4. **Develop directly on VM**:
   - Edit files in VS Code on Mac
   - Files saved directly on VM
   - Run commands in VS Code terminal
   - Build and test instantly

**Workflow:**
```bash
# In VS Code terminal (connected to VM)
make build
sudo ./ravact

# Or run tests
make test
```

## Recommended Setup: UTM + VS Code Remote SSH

**Why this combination?**
- ✅ Native ARM64 performance
- ✅ Free and open source
- ✅ Edit on Mac, run on Linux instantly
- ✅ No file syncing needed
- ✅ Full IDE experience

**Setup Steps:**

1. **Create VM with UTM** (steps above)
2. **Install Go on VM** (steps above)
3. **Clone/Copy project to VM**:
   ```bash
   ssh devuser@<VM-IP>
   git clone <your-repo> ravact-go
   cd ravact-go
   ```
4. **Configure VS Code Remote SSH** (steps above)
5. **Connect and develop!**

## Testing Workflow

### Quick Test Cycle

```bash
# 1. Make changes on Mac (or in VS Code Remote)
# Edit files in internal/, assets/, etc.

# 2. Build for Linux ARM64
make build-linux-arm64

# 3. Copy to VM (if not using VS Code Remote)
scp dist/ravact-linux-arm64 devuser@<VM-IP>:~/ravact
# OR if using shared folder: already available
# OR if using VS Code Remote: already built on VM

# 4. Test on VM
ssh devuser@<VM-IP>
sudo ./ravact

# 5. Iterate!
```

### Full Test Cycle

```bash
# On VM (via SSH or VS Code Remote)

# Run unit tests
make test

# Run integration tests
make test-integration

# Build
make build

# Test the application
sudo ./ravact

# Test specific setup script
sudo bash assets/scripts/nginx.sh

# Verify installations
systemctl status nginx
```

## Network Configuration

### Port Forwarding (Access VM from Mac)

**UTM:**
1. VM Settings → Network
2. Change to "Bridged" or "Host Only"
3. Or use port forwarding

**Access VM services from Mac:**
```bash
# If VM IP is 192.168.64.5
# Nginx running on VM
curl http://192.168.64.5

# SSH from Mac
ssh devuser@192.168.64.5
```

### Recommended VM Network Settings

**UTM:**
- Network Mode: **Shared Network** (default)
- Gives VM internet access
- VM gets IP in range like 192.168.64.x

**Parallels:**
- Network: **Shared Network**
- Enable "Connect Mac to this network"

## Snapshot Strategy

### Create Snapshots at Key Points

**UTM/Parallels:**

1. **"Fresh Install"** - Right after Ubuntu installation
2. **"Dev Ready"** - After installing Go and tools
3. **"Before Testing"** - Before running install scripts
4. **"With Nginx"** - After installing Nginx (for quick reset)

**Why?**
- Quickly restore to clean state
- Test installation scripts repeatedly
- Recover from mistakes
- Different configurations for different tests

**Create Snapshot:**
- UTM: Machine → Save Snapshot
- Parallels: Actions → Take Snapshot

**Restore:**
- UTM: Machine → Restore Snapshot
- Parallels: Actions → Manage Snapshots → Restore

## Performance Optimization

### VM Settings

**CPU:**
- Minimum: 2 cores
- Recommended: 4 cores
- Maximum: Half of your Mac's cores

**Memory:**
- Minimum: 2 GB
- Recommended: 4 GB
- Optimal: 8 GB (if you have 16GB+ Mac)

**Disk:**
- Minimum: 10 GB
- Recommended: 20 GB
- For multiple snapshots: 40 GB

### Mac Host

```bash
# Check available resources
sysctl hw.ncpu        # Total CPU cores
sysctl hw.memsize     # Total RAM
df -h                 # Disk space
```

## SSH Key Setup (Easier Access)

### Generate SSH Key

```bash
# On Mac
ssh-keygen -t ed25519 -f ~/.ssh/ravact-vm -N ""

# Copy to VM
ssh-copy-id -i ~/.ssh/ravact-vm devuser@<VM-IP>

# Add to SSH config
cat >> ~/.ssh/config << EOF
Host ravact-vm
    HostName <VM-IP>
    User devuser
    IdentityFile ~/.ssh/ravact-vm
EOF

# Now connect easily
ssh ravact-vm
```

## Recommended Development Setup

### My Suggested Setup

**For your M1 Mac:**

1. **VM**: UTM with Ubuntu 24.04 ARM64
   - 4 CPU cores
   - 4-8 GB RAM
   - 20 GB disk
   - Shared network

2. **IDE**: VS Code with Remote-SSH
   - Develop directly on VM
   - No file syncing issues
   - Terminal integrated

3. **Workflow**:
   ```
   Edit in VS Code (connected to VM)
     ↓
   Save (instantly on VM)
     ↓
   Run in terminal: make build
     ↓
   Test: sudo ./ravact
     ↓
   Iterate!
   ```

4. **Snapshots**:
   - "Clean Ubuntu" - Fresh start
   - "With Dependencies" - Go + tools installed
   - "Pre-Test" - Before installation tests

## Quick Setup Script

Save this as `setup-vm.sh` and run on Ubuntu VM:

```bash
#!/bin/bash
# Quick setup script for Ubuntu VM

set -e

echo "Setting up Ubuntu VM for Ravact development..."

# Update system
sudo apt update && sudo apt upgrade -y

# Install essential tools
sudo apt install -y \
    build-essential \
    curl \
    wget \
    git \
    vim \
    htop \
    net-tools \
    tree

# Install Go 1.21
cd /tmp
wget -q https://go.dev/dl/go1.21.0.linux-arm64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-arm64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
source ~/.bashrc

# Verify
go version

# Clone or create project directory
mkdir -p ~/ravact-go

echo ""
echo "✓ VM setup complete!"
echo ""
echo "Next steps:"
echo "  1. Copy your project files: scp -r ... or git clone"
echo "  2. cd ravact-go"
echo "  3. make build"
echo "  4. sudo ./ravact"
```

## Testing Setup Scripts

### Snapshot-Based Testing

**Strategy:**

1. **Create "Pre-Test" snapshot** - Clean Ubuntu with Go installed

2. **Test installation script**:
   ```bash
   sudo bash assets/scripts/nginx.sh
   systemctl status nginx
   curl http://localhost
   ```

3. **If works**: Document and commit

4. **Restore snapshot** for next test

5. **Test next script**

**Benefits:**
- Test on clean system every time
- No conflicts between installations
- Repeatable testing
- Quick iterations

### Example Test Session

```bash
# Restore to clean snapshot
# (Do this in UTM/Parallels GUI)

# SSH to VM
ssh ravact-vm
cd ravact-go

# Test Nginx installation
sudo bash assets/scripts/nginx.sh

# Verify
systemctl status nginx
curl http://localhost
nginx -v

# Test Ravact TUI
sudo ./ravact
# Navigate to Installed Applications
# Should show Nginx as installed

# Test another script
sudo bash assets/scripts/mysql.sh

# Verify
systemctl status mysql
mysql --version

# All good? Commit changes
# Then restore snapshot for next test
```

## Troubleshooting

### VM Won't Boot

**Check:**
- Architecture: Must be ARM64 (not x86_64)
- ISO: Use ARM64 version
- UTM: Use "Virtualize" not "Emulate"

### Slow Performance

**Fix:**
- Increase CPU cores (4 recommended)
- Increase RAM (4 GB minimum, 8 GB better)
- Use SSD for VM storage
- Close other apps on Mac

### Can't SSH to VM

**Check IP:**
```bash
# In VM console
ip addr show
# Look for IP address (usually 192.168.64.x)
```

**Test connection:**
```bash
# From Mac
ping <VM-IP>
ssh devuser@<VM-IP>
```

**If fails:**
- Check VM network settings (Shared Network)
- Verify SSH is running: `sudo systemctl status ssh`
- Check firewall: `sudo ufw status`

### File Transfer Issues

**Use SCP:**
```bash
# Copy file to VM
scp file.txt devuser@<VM-IP>:~/

# Copy directory to VM
scp -r directory/ devuser@<VM-IP>:~/

# Copy from VM to Mac
scp devuser@<VM-IP>:~/file.txt ./
```

**Or use rsync:**
```bash
# Sync entire project
rsync -avz --exclude 'dist' --exclude '.git' \
    ~/GoApps/ravact/ravact-go/ \
    devuser@<VM-IP>:~/ravact-go/
```

## Alternative: Multipass (Simple but Limited)

Multipass is Canonical's VM tool, very simple:

```bash
# Install
brew install multipass

# Create Ubuntu VM
multipass launch --name ravact-dev --cpus 4 --memory 4G --disk 20G

# Shell into VM
multipass shell ravact-dev

# Install Go and tools (same as above)

# Transfer files
multipass transfer file.txt ravact-dev:/home/ubuntu/
```

**Limitations:**
- Basic networking
- No GUI
- Limited customization
- But very simple to use!

## Recommended Final Setup

### For Best Experience

```
┌─────────────────────────────────────────────────────────────────┐
│  M1 MacBook Pro                                                 │
│                                                                 │
│  ┌────────────────────────────────────────────────────────┐    │
│  │  VS Code (with Remote-SSH)                             │    │
│  │  - Edit files on VM                                    │    │
│  │  - Instant saves                                       │    │
│  │  - Integrated terminal                                 │    │
│  └────────────────────────────────────────────────────────┘    │
│                              │                                  │
│                              │ SSH                              │
│                              ▼                                  │
│  ┌────────────────────────────────────────────────────────┐    │
│  │  UTM VM - Ubuntu 24.04 ARM64                           │    │
│  │  - 4 CPU cores                                         │    │
│  │  - 8 GB RAM                                            │    │
│  │  - 20 GB disk                                          │    │
│  │  - Go 1.21 installed                                   │    │
│  │  - ravact-go project                                   │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                 │
│  Snapshots:                                                     │
│  1. Fresh Ubuntu                                                │
│  2. Dev Ready (Go + tools)                                      │
│  3. Pre-Test (before installations)                             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Daily Workflow

```bash
# Morning: Start VM
# Open UTM → Start ravact-dev

# Open VS Code
code .

# Connect to VM
# F1 → "Remote-SSH: Connect to Host" → ravact-vm

# Open project folder on VM
# Open ~/ravact-go

# Edit, save, build, test - all in VS Code!
# Terminal in VS Code:
make build
sudo ./ravact

# Test specific features
sudo bash assets/scripts/nginx.sh
curl http://localhost

# Evening: Create snapshot if major milestone
# UTM → Machine → Save Snapshot → "Feature X complete"
```

## VM Maintenance

### Keep VM Updated

```bash
# Weekly
ssh ravact-vm
sudo apt update && sudo apt upgrade -y
sudo apt autoremove -y
```

### Clean Up Disk Space

```bash
# Remove old packages
sudo apt autoremove -y
sudo apt clean

# Clear logs
sudo journalctl --vacuum-time=7d

# Check space
df -h
```

### Backup Important Snapshots

- Export snapshots before major VM changes
- Keep "Fresh Install" snapshot always
- Can recreate VM from snapshot if needed

## Performance Tips

### Mac Settings

```bash
# Give VM priority when running
# Close heavy apps (Chrome, etc.)
# Use Activity Monitor to check resources
```

### VM Settings

```bash
# Install VM tools (if using Parallels)
# Keep VM updated
# Don't over-allocate resources (leave some for Mac)
```

### Build Optimization

```bash
# On VM: Use all cores for building
export GOMAXPROCS=4

# Faster builds
go build -o ravact ./cmd/ravact

# Production builds
make build-linux-arm64
```

## Summary

**Recommended Setup for M1 Mac:**
- ✅ UTM with Ubuntu 24.04 ARM64
- ✅ VS Code with Remote-SSH
- ✅ 4 CPU cores, 8 GB RAM, 20 GB disk
- ✅ Git-based workflow or direct editing
- ✅ Snapshots for quick resets

**Development Cycle:**
1. Edit in VS Code (on VM via Remote-SSH)
2. Save → Build → Test (instant)
3. Create snapshots at milestones
4. Restore snapshot to test installations

**Benefits:**
- 🚀 Fast ARM64 native performance
- 💻 Best IDE experience
- 🔄 Quick iterations
- 📸 Easy snapshots for testing
- 🎯 True Linux environment

---

**Ready to set up your VM?** Follow the UTM + VS Code Remote SSH method for the best experience!
