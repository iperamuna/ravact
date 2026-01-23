# Multipass VM Management Guide

Complete guide for managing ARM64 and AMD64 VMs for ravact testing.

## 🚀 Quick Start

### Setup VMs

```bash
# ARM64 VM (Native - Fast)
./scripts/setup-multipass.sh

# AMD64 VM (Emulated - For x86_64 testing)
./scripts/setup-multipass-amd64.sh

# Both VMs can run simultaneously!
```

---

## 📦 What's Included

### Setup Scripts
1. **`setup-multipass.sh`** - ARM64 native VM (recommended)
2. **`setup-multipass-amd64.sh`** - AMD64 emulated VM (for testing)

### Management Scripts
3. **`sync-all-vms.sh`** - Sync code to all VMs at once
4. **`vm-manager.sh`** - All-in-one VM management tool

### Auto-Generated Sync Scripts
- `~/.ravact-multipass-sync.sh` - ARM64 quick sync
- `~/.ravact-multipass-amd64-sync.sh` - AMD64 quick sync

---

## 🎯 VM Manager - One Tool to Rule Them All

The `vm-manager.sh` script is your main tool for managing VMs.

### Basic Commands

```bash
# List all ravact VMs
./scripts/vm-manager.sh list

# Show VM status
./scripts/vm-manager.sh status

# Get detailed info
./scripts/vm-manager.sh info [vm]
```

### VM Control

```bash
# Start VMs
./scripts/vm-manager.sh start all        # All VMs
./scripts/vm-manager.sh start arm64      # ARM64 only
./scripts/vm-manager.sh start amd64      # AMD64 only

# Stop VMs
./scripts/vm-manager.sh stop all
./scripts/vm-manager.sh stop arm64
./scripts/vm-manager.sh stop amd64

# Restart VMs
./scripts/vm-manager.sh restart arm64
```

### Access VMs

```bash
# Open shell
./scripts/vm-manager.sh shell arm64
./scripts/vm-manager.sh shell amd64

# SSH access
./scripts/vm-manager.sh ssh arm64

# Run ravact directly
./scripts/vm-manager.sh run arm64
./scripts/vm-manager.sh run amd64
```

### Sync Code

```bash
# Sync to all VMs
./scripts/vm-manager.sh sync all

# Sync to specific VM
./scripts/vm-manager.sh sync arm64
./scripts/vm-manager.sh sync amd64
```

### Cleanup

```bash
# Delete all ravact VMs
./scripts/vm-manager.sh clean
```

---

## 🏗️ Architecture Comparison

### ARM64 VM (Native)

**Create:**
```bash
./scripts/setup-multipass.sh
```

**Specs:**
- Architecture: ARM64 (aarch64)
- Performance: ⚡⚡⚡ Native speed
- VM Name: `ravact-dev`
- Use Case: Primary development

**Sync:**
```bash
~/.ravact-multipass-sync.sh
# Or
./scripts/vm-manager.sh sync arm64
```

**Access:**
```bash
multipass shell ravact-dev
# Or
ssh ravact-dev
# Or
./scripts/vm-manager.sh shell arm64
```

---

### AMD64 VM (Emulated)

**Create:**
```bash
./scripts/setup-multipass-amd64.sh
```

**Specs:**
- Architecture: x86_64 (AMD64)
- Performance: ⚡ Slower (QEMU emulation)
- VM Name: `ravact-dev-amd64`
- Use Case: x86_64 compatibility testing

**Sync:**
```bash
~/.ravact-multipass-amd64-sync.sh
# Or
./scripts/vm-manager.sh sync amd64
```

**Access:**
```bash
multipass shell ravact-dev-amd64
# Or
ssh ravact-dev-amd64
# Or
./scripts/vm-manager.sh shell amd64
```

---

## 🔄 Daily Workflow

### Option 1: VM Manager (Recommended)

```bash
# 1. Make code changes
vim internal/ui/screens/main_menu.go

# 2. Sync to all VMs
./scripts/vm-manager.sh sync all

# 3. Test on ARM64
./scripts/vm-manager.sh run arm64

# 4. Test on AMD64 (optional)
./scripts/vm-manager.sh run amd64
```

### Option 2: Individual Sync Scripts

```bash
# 1. Make code changes
vim internal/ui/screens/main_menu.go

# 2a. Sync to ARM64
~/.ravact-multipass-sync.sh

# 2b. Sync to AMD64
~/.ravact-multipass-amd64-sync.sh

# 3. Test
ssh ravact-dev
cd ravact-go && sudo ./ravact
```

### Option 3: Sync All Script

```bash
# 1. Make changes
vim internal/system/users.go

# 2. Sync to all VMs at once
./scripts/sync-all-vms.sh

# This builds for both architectures and deploys
```

---

## 📊 Command Reference

### VM Manager Quick Reference

| Command | Description | Example |
|---------|-------------|---------|
| `list` | List all VMs | `./scripts/vm-manager.sh list` |
| `status` | Show VM status | `./scripts/vm-manager.sh status` |
| `info [vm]` | Detailed VM info | `./scripts/vm-manager.sh info arm64` |
| `start [vm]` | Start VM(s) | `./scripts/vm-manager.sh start all` |
| `stop [vm]` | Stop VM(s) | `./scripts/vm-manager.sh stop arm64` |
| `restart [vm]` | Restart VM(s) | `./scripts/vm-manager.sh restart amd64` |
| `shell [vm]` | Open shell | `./scripts/vm-manager.sh shell arm64` |
| `ssh [vm]` | SSH to VM | `./scripts/vm-manager.sh ssh amd64` |
| `run [vm]` | Run ravact | `./scripts/vm-manager.sh run arm64` |
| `sync [vm]` | Sync code | `./scripts/vm-manager.sh sync all` |
| `clean` | Delete all VMs | `./scripts/vm-manager.sh clean` |

### VM Identifiers

You can use these aliases:

| Alias | VM |
|-------|-----|
| `arm64`, `arm` | ravact-dev (ARM64) |
| `amd64`, `amd`, `x86` | ravact-dev-amd64 (AMD64) |
| `all` | Both VMs |

---

## 🛠️ Advanced Usage

### Run Both VMs Side-by-Side

```bash
# Terminal 1: ARM64 VM
./scripts/vm-manager.sh shell arm64
cd ravact-go && sudo ./ravact

# Terminal 2: AMD64 VM
./scripts/vm-manager.sh shell amd64
cd ravact-go && sudo ./ravact

# Compare performance and behavior!
```

### Performance Testing

```bash
# Test on ARM64 (native)
time ./scripts/vm-manager.sh run arm64

# Test on AMD64 (emulated)
time ./scripts/vm-manager.sh run amd64

# ARM64 should be significantly faster
```

### Snapshot Management

```bash
# Create snapshot
multipass snapshot ravact-dev --name clean-state
multipass snapshot ravact-dev-amd64 --name clean-state

# Restore snapshot
multipass restore ravact-dev --snapshot clean-state

# List snapshots
multipass snapshot list ravact-dev
```

### Resource Adjustment

```bash
# Stop VM first
./scripts/vm-manager.sh stop arm64

# Edit resources (requires VM recreation)
multipass delete ravact-dev
multipass purge

# Recreate with more resources
multipass launch -n ravact-dev -c 8 -m 8G -d 40G 24.04
```

---

## 🎨 VS Code Integration

### Remote SSH to Both VMs

**Setup:**
1. Install "Remote - SSH" extension
2. Both VMs are already in `~/.ssh/config`

**Connect:**
- ARM64: Connect to `ravact-dev`
- AMD64: Connect to `ravact-dev-amd64`

**Workflow:**
1. Open VS Code
2. F1 → "Remote-SSH: Connect to Host"
3. Select `ravact-dev` or `ravact-dev-amd64`
4. Open folder: `/home/ubuntu/ravact-go`
5. Edit and test directly on VM!

---

## 🔧 Troubleshooting

### VM Won't Start

```bash
# Check status
multipass list

# Try restart
./scripts/vm-manager.sh restart arm64

# Check logs
multipass logs ravact-dev
```

### Sync Fails

```bash
# Ensure VM is running
./scripts/vm-manager.sh start arm64

# Test connection
ssh ravact-dev echo "OK"

# Rebuild manually
make build-linux-arm64
multipass transfer dist/ravact-linux-arm64 ravact-dev:/home/ubuntu/ravact-go/ravact
```

### AMD64 VM Too Slow

**Expected:** AMD64 runs under QEMU emulation on M1, so it will be slower.

**Solutions:**
1. Use for compatibility testing only
2. Use ARM64 for daily development
3. Or deploy to actual x86_64 server for real testing

### Out of Disk Space

```bash
# Check disk usage
./scripts/vm-manager.sh shell arm64
df -h

# Clean up
sudo apt clean
sudo apt autoremove
```

---

## 📈 Performance Comparison

| Metric | ARM64 (Native) | AMD64 (Emulated) |
|--------|----------------|------------------|
| Boot Time | ~10 seconds | ~30 seconds |
| App Startup | ~1 second | ~5 seconds |
| Build Speed | ⚡⚡⚡ Fast | ⚡ Slower |
| Runtime | ⚡⚡⚡ Native | ⚡ Acceptable |
| Use Case | Daily development | Compatibility testing |

---

## 🎯 Best Practices

### 1. Use ARM64 for Development
```bash
# Primary development
./scripts/setup-multipass.sh
./scripts/vm-manager.sh sync arm64
```

### 2. Test on AMD64 Before Release
```bash
# Before releasing
./scripts/setup-multipass-amd64.sh
./scripts/sync-all-vms.sh

# Test both
./scripts/vm-manager.sh run arm64
./scripts/vm-manager.sh run amd64
```

### 3. Keep VMs Updated
```bash
# Update system packages regularly
./scripts/vm-manager.sh shell arm64
sudo apt update && sudo apt upgrade -y
```

### 4. Use Snapshots
```bash
# Before major changes
multipass snapshot ravact-dev --name before-feature-x

# If something breaks
multipass restore ravact-dev --snapshot before-feature-x
```

---

## 📚 Related Documentation

- **VM Setup:** `scripts/VM_SETUP_README.md`
- **UTM Alternative:** `scripts/UTM_TROUBLESHOOTING.md`
- **macOS Limitations:** `docs/MACOS_LIMITATIONS.md`
- **Development Guide:** `docs/DEVELOPMENT.md`

---

## 🚀 Quick Commands Cheat Sheet

```bash
# Setup
./scripts/setup-multipass.sh              # ARM64
./scripts/setup-multipass-amd64.sh        # AMD64

# List & Status
./scripts/vm-manager.sh list              # List VMs
./scripts/vm-manager.sh status            # Status

# Access
./scripts/vm-manager.sh shell arm64       # Shell
./scripts/vm-manager.sh run arm64         # Run ravact

# Sync
./scripts/vm-manager.sh sync all          # Sync to all
./scripts/sync-all-vms.sh                 # Alternative

# Control
./scripts/vm-manager.sh start all         # Start
./scripts/vm-manager.sh stop all          # Stop
./scripts/vm-manager.sh restart arm64     # Restart

# Direct Multipass
multipass shell ravact-dev                # ARM64 shell
multipass shell ravact-dev-amd64          # AMD64 shell
multipass list                            # List all VMs

# Direct SSH
ssh ravact-dev                            # ARM64
ssh ravact-dev-amd64                      # AMD64
```

---

## ✅ Summary

You now have:
- ✅ ARM64 VM for fast native development
- ✅ AMD64 VM for x86_64 compatibility testing
- ✅ Unified VM manager for all operations
- ✅ Sync scripts for quick deployments
- ✅ Complete documentation

**Start developing:**
```bash
./scripts/setup-multipass.sh
./scripts/vm-manager.sh sync arm64
./scripts/vm-manager.sh run arm64
```

**Happy testing on both architectures! 🎉**
