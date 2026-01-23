# VM Setup Scripts for M1 Mac

Quick automation scripts to set up Ubuntu VM and deploy ravact for testing.

## Scripts Overview

### 1. `setup-mac-vm.sh` - Complete Automated Setup (Mac Side)
**Use this for full automation from your Mac**

- ✅ Checks/installs UTM
- ✅ Downloads Ubuntu 24.04 ARM64 ISO
- ✅ Guides you through VM creation
- ✅ Configures SSH access
- ✅ Sets up the VM automatically
- ✅ Builds and deploys ravact
- ✅ Creates sync script for future updates

**Usage:**
```bash
cd ravact-go/scripts
./setup-mac-vm.sh
```

**What it does:**
1. Verifies you're on M1 Mac
2. Installs UTM (if needed)
3. Downloads Ubuntu ISO (if needed)
4. Guides VM creation in UTM
5. Connects via SSH and sets up VM
6. Deploys ravact with all assets
7. Creates helper scripts for future updates

**Result:**
- VM fully configured with Go 1.21
- Ravact installed and ready to run
- SSH configured (`ssh ravact-dev`)
- Sync script created for quick updates

---

### 2. `setup-vm-only.sh` - VM Setup Only (Runs Inside VM)
**Use this if you manually created the VM and just want to set it up**

- ✅ Updates Ubuntu packages
- ✅ Installs development tools
- ✅ Installs Go 1.21
- ✅ Creates project directories
- ✅ Creates helper scripts

**Usage:**

Option A - Copy and run:
```bash
# From your Mac
scp scripts/setup-vm-only.sh devuser@<VM-IP>:~/

# SSH into VM
ssh devuser@<VM-IP>

# Run setup
bash setup-vm-only.sh
```

Option B - Direct download (if you have this on GitHub):
```bash
# Inside VM
bash <(curl -s https://raw.githubusercontent.com/your-repo/main/scripts/setup-vm-only.sh)
```

**Result:**
- VM ready for development
- Go installed and configured
- Project directories created
- Helper scripts available

---

### 3. `quick-deploy.sh` - Fast Deploy After Code Changes
**Use this for day-to-day development**

After you make code changes, quickly build and deploy to VM.

**Usage:**
```bash
# From Mac (in ravact-go directory)
./scripts/quick-deploy.sh 192.168.64.5 devuser

# Or if SSH config is set up:
./scripts/quick-deploy.sh
```

**What it does:**
1. Builds ravact for ARM64 Linux
2. Copies binary to VM
3. Copies assets (scripts/configs)
4. Tests the deployment

**Result:**
- Latest code running on VM in seconds
- No manual copy/paste needed

---

## Quick Start Guide

### First Time Setup

1. **Run the main setup script:**
   ```bash
   cd ravact-go/scripts
   ./setup-mac-vm.sh
   ```

2. **Follow the prompts to:**
   - Create VM in UTM
   - Install Ubuntu
   - Get VM IP address

3. **Script does the rest automatically!**

### Daily Development Workflow

**Option A - Quick Deploy (Recommended):**
```bash
# 1. Make code changes on Mac
# 2. Deploy to VM
./scripts/quick-deploy.sh

# 3. Test on VM
ssh ravact-dev
cd ravact-go
sudo ./ravact
```

**Option B - VS Code Remote SSH (Best Experience):**
```bash
# 1. Connect VS Code to VM (F1 → Remote-SSH)
# 2. Edit files directly on VM
# 3. Build in terminal: make build
# 4. Test: sudo ./ravact
```

---

## VM Details After Setup

**SSH Access:**
```bash
ssh ravact-dev              # Quick access
ssh devuser@<VM-IP>         # Direct IP
```

**VM Credentials:**
- Username: `devuser`
- Password: `devuser`
- Hostname: `ravact-dev`

**Project Location on VM:**
```
~/ravact-go/
├── ravact              # Binary
├── assets/
│   ├── scripts/        # Setup scripts
│   └── configs/        # Config files
├── test-ravact.sh      # Test helper
└── build-ravact.sh     # Build helper (if source available)
```

**Helper Commands on VM:**
```bash
cd ~/ravact-go

# Test ravact
./test-ravact.sh

# Run ravact
sudo ./ravact

# If you have source code
./build-ravact.sh
```

---

## Troubleshooting

### Cannot connect to VM via SSH

**Solution:**
```bash
# In VM console, check IP
ip addr show

# Verify SSH is running
sudo systemctl status ssh

# Test from Mac
ping <VM-IP>
ssh devuser@<VM-IP>
```

### Build fails on Mac

**Solution:**
```bash
# Ensure you're in ravact-go directory
cd ravact-go

# Try manual build
make build-linux-arm64

# Check output
ls -lh dist/
```

### Script permission denied

**Solution:**
```bash
chmod +x scripts/*.sh
```

### VM is slow

**In UTM, increase resources:**
- CPU: 4 cores (recommended)
- RAM: 8 GB (if you have 16GB+ Mac)
- Check: Activity Monitor on Mac

---

## Advanced Usage

### Create Snapshot Before Testing

**UTM:**
1. VM → Machine → Save Snapshot
2. Name: "Clean State" or "Pre-Test"

**Restore:**
- Machine → Restore Snapshot

**Use Case:**
- Test installation scripts repeatedly
- Quickly reset to clean state

### Custom Sync Script

The main setup creates `~/.ravact-sync.sh`:
```bash
# Quick sync for future updates
~/.ravact-sync.sh
```

### Build and Deploy in One Command

```bash
cd ravact-go
make build-linux-arm64 && ./scripts/quick-deploy.sh
```

---

## What Gets Installed on VM

### System Packages
- build-essential (gcc, make, etc.)
- curl, wget, git
- vim, htop, tree
- net-tools, jq, unzip
- tmux (terminal multiplexer)

### Go Environment
- Go 1.21.6 for ARM64
- GOPATH configured
- PATH updated

### Project Structure
- ~/ravact-go/ - Main directory
- Helper scripts for testing
- Assets (scripts/configs)

---

## Quick Reference

| Task | Command |
|------|---------|
| First setup | `./scripts/setup-mac-vm.sh` |
| Deploy updates | `./scripts/quick-deploy.sh` |
| SSH to VM | `ssh ravact-dev` |
| Run ravact | `ssh ravact-dev 'cd ravact-go && sudo ./ravact'` |
| Setup VM only | `bash setup-vm-only.sh` (inside VM) |

---

## Need Help?

**Common Issues:**
1. **UTM not found**: Install with `brew install --cask utm`
2. **ISO download slow**: ~2.5GB file, be patient or download manually
3. **SSH fails**: Ensure OpenSSH was installed during Ubuntu setup
4. **Build fails**: Run `make deps` first

**Manual Steps (if automation fails):**
1. Create VM in UTM manually (see main docs)
2. Copy `setup-vm-only.sh` to VM and run it
3. Use `quick-deploy.sh` to deploy ravact

---

## Scripts Location

```
ravact-go/scripts/
├── setup-mac-vm.sh      # Main setup (Mac side)
├── setup-vm-only.sh     # VM-only setup
├── quick-deploy.sh      # Fast deploy
└── VM_SETUP_README.md   # This file
```

**Related Documentation:**
- [DEV_VM_SETUP.md](../docs/DEV_VM_SETUP.md) - Full manual guide
- [DEVELOPMENT.md](../docs/DEVELOPMENT.md) - Development workflows

---

**Happy Testing! 🚀**
