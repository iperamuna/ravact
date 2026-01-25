# Ravact Scripts

This directory contains utility scripts for development, testing, deployment, and release management of Ravact.

---

## 📋 Quick Reference

| Script | Purpose | Platform |
|--------|---------|----------|
| [`install.sh`](#installsh) | Install Ravact on any system | Linux/macOS |
| [`release.sh`](#releasesh) | Create new releases | macOS/Linux |
| [`test.sh`](#testsh) | Run all tests | Any |
| [`quick-deploy.sh`](#quick-deploysh) | Deploy to remote VM | macOS |
| [`vm-manager.sh`](#vm-managersh) | Manage development VMs | macOS |

---

## 🚀 Installation

### install.sh

**Purpose:** Downloads and installs the correct Ravact binary for your system.

**Usage:**
```bash
# Recommended: One-command install
curl -sSL https://raw.githubusercontent.com/iperamuna/ravact/main/scripts/install.sh | sudo bash

# Or run locally
sudo ./scripts/install.sh

# Install from local binary
sudo ./scripts/install.sh --local /path/to/ravact-binary
```

**Features:**
- Auto-detects OS (Linux/macOS) and architecture (AMD64/ARM64)
- Downloads latest release from GitHub
- Installs to `/usr/local/bin/ravact`
- Creates backup of existing installation
- Supports manual version selection

---

## 🧪 Testing Scripts

### test.sh

**Purpose:** Run the complete test suite including unit tests, integration tests, and race detection.

**Usage:**
```bash
./scripts/test.sh
```

**What it does:**
1. Runs unit tests with coverage
2. Runs integration tests (requires Linux)
3. Runs race detector
4. Generates coverage report (`coverage.html`)

---

### test-ravact-features.sh

**Purpose:** Comprehensive feature testing script that validates the binary and system requirements.

**Usage:**
```bash
./scripts/test-ravact-features.sh
```

**Tests performed:**
- Binary execution and version check
- Architecture verification (x86_64)
- Assets existence (scripts and configs)
- System files (`/etc/passwd`, `/etc/group`)
- Required commands availability
- Root privileges check

---

### test-user-management.sh

**Purpose:** Specifically tests the user management feature that had known issues.

**Usage:**
```bash
./scripts/test-user-management.sh
```

**Tests:**
- `groups` command functionality
- `/etc/passwd` reading
- `/etc/group` reading
- Provides manual testing instructions

---

### manual-test-guide.sh

**Purpose:** Interactive guide for manual testing of Ravact features.

**Usage:**
```bash
./scripts/manual-test-guide.sh
```

**Provides:**
- Step-by-step testing instructions
- Test checklist for all features
- Pre-flight system checks
- Expected vs broken behavior descriptions

---

## 🐳 Docker Scripts

### docker-amd64-dev.sh

**Purpose:** Creates a persistent AMD64 development container for testing on Apple Silicon Macs.

**Usage:**
```bash
./scripts/docker-amd64-dev.sh
```

**Features:**
- Creates Ubuntu 24.04 AMD64 container via QEMU emulation
- Mounts project directory for live code sync
- Persists between sessions
- Auto-reconnects if container exists

**Workflow:**
1. Run script to create/connect to container
2. Build on Mac: `make build-linux`
3. Test in container: `sudo ./dist/ravact-linux-amd64`

---

### docker-build-and-test.sh

**Purpose:** One-command build and test in AMD64 container.

**Usage:**
```bash
./scripts/docker-build-and-test.sh
```

**What it does:**
1. Builds Ravact for AMD64 on Mac
2. Starts/creates container if needed
3. Runs Ravact in container

---

### docker-manager.sh

**Purpose:** Manage the AMD64 development container.

**Usage:**
```bash
./scripts/docker-manager.sh <command>

# Commands:
./scripts/docker-manager.sh start      # Start container
./scripts/docker-manager.sh stop       # Stop container
./scripts/docker-manager.sh restart    # Restart container
./scripts/docker-manager.sh shell      # Open shell in container
./scripts/docker-manager.sh status     # Show container status
./scripts/docker-manager.sh logs       # Show container logs
./scripts/docker-manager.sh run        # Build & test
./scripts/docker-manager.sh delete     # Delete container
./scripts/docker-manager.sh recreate   # Delete and recreate
```

---

### docker-test.sh

**Purpose:** Run tests in a Docker container (requires `Dockerfile.test`).

**Usage:**
```bash
./scripts/docker-test.sh
```

---

### test-docker-amd64.sh

**Purpose:** Quick test of Ravact in x86_64 Docker container with QEMU emulation.

**Usage:**
```bash
./scripts/test-docker-amd64.sh
```

**Note:** Slower than native due to QEMU emulation on Apple Silicon.

---

## 🖥️ VM Setup Scripts

### setup-multipass.sh

**Purpose:** Sets up an Ubuntu VM using Multipass (easiest option for Mac users).

**Usage:**
```bash
./scripts/setup-multipass.sh
```

**What it does:**
1. Installs Multipass if not present
2. Creates Ubuntu 24.04 VM (4 CPUs, 4GB RAM, 20GB disk)
3. Installs Go 1.24 and development tools
4. Deploys Ravact binary and assets
5. Creates sync script for future updates

**VM Specs:**
- Name: `ravact-dev`
- OS: Ubuntu 24.04 LTS
- Architecture: ARM64 (native on Apple Silicon)

---

### setup-mac-vm.sh

**Purpose:** Sets up Ubuntu VM using UTM (alternative to Multipass).

**Usage:**
```bash
./scripts/setup-mac-vm.sh
```

**Requirements:**
- Apple Silicon Mac (M1/M2/M3)
- UTM installed (or will install via Homebrew)
- Ubuntu 24.04 ARM64 ISO

**Features:**
- Interactive VM creation guide
- SSH configuration
- Automatic deployment after setup

---

### setup-vm-only.sh

**Purpose:** Script to run INSIDE an existing Ubuntu VM to set up the development environment.

**Usage:**
```bash
# SSH into your VM first, then:
bash setup-vm-only.sh

# Or via curl:
bash <(curl -s https://raw.githubusercontent.com/iperamuna/ravact/main/scripts/setup-vm-only.sh)
```

**Installs:**
- Go 1.24
- Build tools (gcc, make, etc.)
- Development utilities (git, vim, htop, etc.)
- Project directories

---

### vm-manager.sh

**Purpose:** Unified management of all Ravact development VMs.

**Usage:**
```bash
./scripts/vm-manager.sh <command> [vm]

# Commands:
./scripts/vm-manager.sh list              # List all VMs
./scripts/vm-manager.sh info [vm]         # Show VM info
./scripts/vm-manager.sh start [vm]        # Start VM(s)
./scripts/vm-manager.sh stop [vm]         # Stop VM(s)
./scripts/vm-manager.sh restart [vm]      # Restart VM(s)
./scripts/vm-manager.sh shell [vm]        # Open shell
./scripts/vm-manager.sh ssh [vm]          # SSH to VM
./scripts/vm-manager.sh run [vm]          # Run Ravact
./scripts/vm-manager.sh sync [vm]         # Sync code
./scripts/vm-manager.sh clean             # Delete all VMs

# VM targets:
#   arm64, arm     - ARM64 VM (native)
#   amd64, amd     - AMD64 VM (emulated)
#   all            - All VMs (default)
```

---

## 🚀 Deployment Scripts

### quick-deploy.sh

**Purpose:** Build on Mac and deploy to a remote VM via SSH.

**Usage:**
```bash
./scripts/quick-deploy.sh <VM_IP> [VM_USER]

# Examples:
./scripts/quick-deploy.sh 192.168.64.5 devuser
./scripts/quick-deploy.sh  # Uses SSH config if available
```

**What it does:**
1. Tests SSH connection
2. Builds for Linux ARM64
3. Creates directories on VM
4. Deploys binary and assets
5. Tests installation

---

## 📦 Release Scripts

### release.sh

**Purpose:** Automates the complete release process.

**Usage:**
```bash
./scripts/release.sh
```

**Process:**
1. Checks for clean git working directory
2. Prompts for version bump (patch/minor/major/custom)
3. Generates release notes from commits
4. Opens editor for release notes customization
5. Builds binaries for all platforms:
   - Linux AMD64
   - Linux ARM64
   - macOS AMD64
   - macOS ARM64
6. Generates SHA256 checksums
7. Creates git tag
8. Optionally pushes to GitHub

**Output:**
- Binaries in `dist/` directory
- `checksums.txt` for verification
- Git tag ready for release

---

## 📁 Script Categories

### Development Workflow

| Task | Script |
|------|--------|
| Set up new dev environment | `setup-multipass.sh` or `setup-mac-vm.sh` |
| Quick build & test | `docker-build-and-test.sh` |
| Manage containers | `docker-manager.sh` |
| Manage VMs | `vm-manager.sh` |
| Deploy to VM | `quick-deploy.sh` |

### Testing

| Task | Script |
|------|--------|
| Full test suite | `test.sh` |
| Feature validation | `test-ravact-features.sh` |
| User management testing | `test-user-management.sh` |
| Manual testing guide | `manual-test-guide.sh` |
| Docker AMD64 testing | `test-docker-amd64.sh` |

### Release & Deployment

| Task | Script |
|------|--------|
| Create release | `release.sh` |
| Install Ravact | `install.sh` |
| Deploy to server | `quick-deploy.sh` |

---

## 🔧 Requirements

### For Development Scripts
- **macOS**: Homebrew, Docker Desktop, or Multipass
- **Linux**: Docker (optional)

### For Testing Scripts
- Go 1.24+
- Docker (for container tests)
- Linux environment (for full integration tests)

### For Release Scripts
- Git with clean working directory
- Go 1.24+
- GitHub CLI (optional, for automated releases)

---

## 💡 Tips

### Fastest Development Cycle on Mac

1. **First time setup:**
   ```bash
   ./scripts/setup-multipass.sh
   ```

2. **Daily workflow:**
   ```bash
   # Make code changes, then:
   make build-linux-arm64
   
   # Sync and test:
   multipass transfer dist/ravact-linux-arm64 ravact-dev:/home/ubuntu/ravact-go/ravact
   multipass exec ravact-dev -- sudo /home/ubuntu/ravact-go/ravact
   ```

### Testing AMD64 on Apple Silicon

```bash
# Start persistent container
./scripts/docker-amd64-dev.sh

# In another terminal, build and deploy
make build-linux
docker exec -it ravact-amd64-dev bash -c 'sudo ./dist/ravact-linux-amd64'
```

### Quick Release

```bash
# Ensure everything is committed
git status

# Run release script
./scripts/release.sh

# Follow prompts to create release
```

---

## 📝 Notes

- All scripts use consistent color-coded output (green=success, yellow=warning, red=error, blue=info)
- Scripts are designed to be idempotent (safe to run multiple times)
- VM scripts create sync scripts in `~/.ravact-*-sync.sh` for quick updates
- Docker container name: `ravact-amd64-dev`
- Default VM name: `ravact-dev`

---

**Version:** 0.2.1 | **Last Updated:** January 2026
