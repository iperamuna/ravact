# AMD64 VM Setup - Summary

## ✅ What's Been Created

You now have **complete multi-architecture VM support**!

### 🎯 New Scripts Created

1. **`scripts/setup-multipass-amd64.sh`** ⭐ NEW!
   - Creates x86_64/AMD64 Ubuntu VM
   - Uses QEMU emulation on M1 Mac
   - Perfect for testing x86_64 compatibility
   - Auto-installs Go 1.21 for AMD64
   - Deploys ravact AMD64 binary

2. **`scripts/sync-all-vms.sh`** ⭐ NEW!
   - Syncs to ALL VMs at once
   - Builds both ARM64 and AMD64
   - Deploys to respective VMs
   - One command for everything

3. **`scripts/vm-manager.sh`** ⭐ NEW!
   - All-in-one VM management tool
   - Control all VMs from one place
   - Start, stop, sync, run, shell access
   - Supports both ARM64 and AMD64

4. **Auto-generated sync scripts:**
   - `~/.ravact-multipass-sync.sh` (ARM64)
   - `~/.ravact-multipass-amd64-sync.sh` (AMD64)

### 📚 Documentation Created

1. **`scripts/MULTIPASS_GUIDE.md`**
   - Complete Multipass guide
   - Both architectures covered
   - Daily workflows
   - Performance comparison

2. **`scripts/README.md`**
   - Scripts directory overview
   - Quick reference
   - When to use what

---

## 🚀 Quick Start

### Setup Both VMs (Recommended)

```bash
cd ravact-go/scripts

# 1. ARM64 VM (native, fast)
./setup-multipass.sh

# 2. AMD64 VM (emulated, for testing)
./setup-multipass-amd64.sh
```

**Time:** ~10 minutes total (both VMs)

---

## 🎯 Architecture Comparison

| Feature | ARM64 (Native) | AMD64 (Emulated) |
|---------|----------------|------------------|
| **Script** | `setup-multipass.sh` | `setup-multipass-amd64.sh` |
| **VM Name** | `ravact-dev` | `ravact-dev-amd64` |
| **Performance** | ⚡⚡⚡ Fast | ⚡ Slower (QEMU) |
| **Build Target** | `make build-linux-arm64` | `make build-linux` |
| **Use For** | Daily development | x86_64 testing |
| **Boot Time** | ~10 seconds | ~30 seconds |
| **Speed** | Native M1 | Acceptable |

---

## 💻 Daily Workflow

### Option 1: VM Manager (Easiest) ⭐

```bash
# 1. Make code changes
vim internal/ui/screens/main_menu.go

# 2. Sync to all VMs
./scripts/vm-manager.sh sync all

# 3. Test on ARM64 (fast)
./scripts/vm-manager.sh run arm64

# 4. Test on AMD64 (optional)
./scripts/vm-manager.sh run amd64
```

### Option 2: Sync All Script

```bash
# 1. Make changes
vim internal/system/users.go

# 2. Build and deploy to all VMs
./scripts/sync-all-vms.sh

# 3. Test each VM
ssh ravact-dev "cd ravact-go && sudo ./ravact"
ssh ravact-dev-amd64 "cd ravact-go && sudo ./ravact"
```

### Option 3: Individual Sync

```bash
# ARM64 only
~/.ravact-multipass-sync.sh
ssh ravact-dev

# AMD64 only
~/.ravact-multipass-amd64-sync.sh
ssh ravact-dev-amd64
```

---

## 🛠️ VM Manager Commands

### List & Status
```bash
./scripts/vm-manager.sh list          # List all ravact VMs
./scripts/vm-manager.sh status        # Show VM status
./scripts/vm-manager.sh info arm64    # Detailed info
```

### Control VMs
```bash
./scripts/vm-manager.sh start all     # Start all VMs
./scripts/vm-manager.sh stop arm64    # Stop ARM64 VM
./scripts/vm-manager.sh restart amd64 # Restart AMD64 VM
```

### Access VMs
```bash
./scripts/vm-manager.sh shell arm64   # Open shell in ARM64
./scripts/vm-manager.sh shell amd64   # Open shell in AMD64
./scripts/vm-manager.sh ssh arm64     # SSH to VM
```

### Deploy & Test
```bash
./scripts/vm-manager.sh sync all      # Sync to all VMs
./scripts/vm-manager.sh sync arm64    # Sync to ARM64 only
./scripts/vm-manager.sh sync amd64    # Sync to AMD64 only
./scripts/vm-manager.sh run arm64     # Run ravact on ARM64
```

### Cleanup
```bash
./scripts/vm-manager.sh clean         # Delete all VMs
```

---

## 📊 What Each Script Does

### setup-multipass-amd64.sh

Creates AMD64 VM with:
1. ✅ Ubuntu 24.04 x86_64
2. ✅ QEMU emulation (automatic)
3. ✅ Go 1.21 for AMD64
4. ✅ Build tools
5. ✅ Ravact AMD64 binary
6. ✅ All assets
7. ✅ SSH configuration
8. ✅ Auto-generated sync script

**Result:** Ready-to-use AMD64 VM!

---

### sync-all-vms.sh

Syncs to multiple VMs:
1. ✅ Detects which VMs exist
2. ✅ Builds ARM64 binary (if ARM64 VM exists)
3. ✅ Builds AMD64 binary (if AMD64 VM exists)
4. ✅ Deploys to respective VMs
5. ✅ Copies all assets
6. ✅ Shows summary

**Result:** All VMs updated with latest code!

---

### vm-manager.sh

One tool for everything:
1. ✅ List VMs
2. ✅ Start/stop/restart VMs
3. ✅ Shell/SSH access
4. ✅ Sync code
5. ✅ Run ravact
6. ✅ Get VM info
7. ✅ Clean up VMs

**Result:** Single interface for all VM operations!

---

## 🎯 Use Cases

### Use Case 1: Daily Development (ARM64 only)

```bash
# Setup (one-time)
./scripts/setup-multipass.sh

# Daily work
vim internal/ui/screens/user_management.go
./scripts/vm-manager.sh sync arm64
./scripts/vm-manager.sh run arm64
```

**Why:** Fastest performance, native speed

---

### Use Case 2: Pre-Release Testing (Both)

```bash
# Setup (one-time)
./scripts/setup-multipass.sh
./scripts/setup-multipass-amd64.sh

# Before release
./scripts/sync-all-vms.sh

# Test both architectures
./scripts/vm-manager.sh run arm64
./scripts/vm-manager.sh run amd64

# Verify compatibility
```

**Why:** Ensure compatibility across architectures

---

### Use Case 3: Performance Comparison

```bash
# Run on both VMs
time ./scripts/vm-manager.sh run arm64
time ./scripts/vm-manager.sh run amd64

# Compare results
```

**Why:** See performance difference, optimize code

---

### Use Case 4: Bug Reproduction

```bash
# If bug only happens on x86_64
./scripts/setup-multipass-amd64.sh
./scripts/vm-manager.sh shell amd64
cd ravact-go && sudo ./ravact

# Debug on AMD64
```

**Why:** Test architecture-specific issues

---

## 🔄 Complete Example Workflow

### Scenario: Add new feature and test on both architectures

```bash
# 1. Create feature branch
git checkout -b feature-new-menu

# 2. Write code
vim internal/ui/screens/new_feature.go

# 3. Deploy to all VMs
./scripts/sync-all-vms.sh

# 4. Test on ARM64 (primary)
./scripts/vm-manager.sh shell arm64
cd ravact-go && sudo ./ravact
# Test feature...
exit

# 5. Test on AMD64 (compatibility)
./scripts/vm-manager.sh shell amd64
cd ravact-go && sudo ./ravact
# Test feature...
exit

# 6. Both work? Commit!
git add .
git commit -m "Add new menu feature"

# 7. Create snapshots (optional)
multipass snapshot ravact-dev --name feature-working
multipass snapshot ravact-dev-amd64 --name feature-working
```

---

## 📈 Performance Expectations

### ARM64 VM (Native)
- **Boot:** ~10 seconds
- **App startup:** ~1 second
- **Build time:** Fast
- **Runtime:** Native speed
- **Recommended for:** Daily development

### AMD64 VM (Emulated)
- **Boot:** ~30 seconds
- **App startup:** ~3-5 seconds
- **Build time:** Slower (2-3x)
- **Runtime:** Acceptable
- **Recommended for:** Testing before release

---

## 🎨 VS Code Integration

Both VMs work with VS Code Remote-SSH:

```bash
# After setup, in VS Code:
# 1. Install "Remote - SSH" extension
# 2. F1 → Remote-SSH: Connect to Host
# 3. Choose:
#    - ravact-dev (ARM64)
#    - ravact-dev-amd64 (AMD64)
# 4. Open folder: /home/ubuntu/ravact-go
# 5. Edit directly on VM!
```

**Benefits:**
- Edit on VM, no syncing needed
- Full Go IntelliSense
- Debug on actual target
- Terminal access built-in

---

## 🔧 Troubleshooting

### AMD64 VM is slow
**Expected!** It's running under QEMU emulation.

**Solutions:**
- Use ARM64 for daily work
- Use AMD64 only for testing
- Or deploy to real x86_64 server

### "VM not found"
```bash
# List all VMs
multipass list

# Check specific VM
./scripts/vm-manager.sh list
```

### Sync fails
```bash
# Ensure VM is running
./scripts/vm-manager.sh start amd64

# Test connection
ssh ravact-dev-amd64 echo "OK"

# Try manual sync
~/.ravact-multipass-amd64-sync.sh
```

### Need fresh start
```bash
# Delete all and recreate
./scripts/vm-manager.sh clean
./scripts/setup-multipass.sh
./scripts/setup-multipass-amd64.sh
```

---

## 📚 Full Documentation

- **`scripts/MULTIPASS_GUIDE.md`** - Complete Multipass guide
- **`scripts/README.md`** - Scripts overview
- **`scripts/VM_SETUP_README.md`** - Detailed setup
- **`docs/DEVELOPMENT.md`** - Development guide

---

## ✅ Summary

You now have:

1. ✅ **ARM64 VM** - Native speed, daily development
2. ✅ **AMD64 VM** - x86_64 testing, compatibility
3. ✅ **VM Manager** - One tool for all operations
4. ✅ **Sync Scripts** - Quick deployment
5. ✅ **Complete Docs** - Guides and references

### Quick Commands

```bash
# Setup (one-time)
./scripts/setup-multipass.sh          # ARM64
./scripts/setup-multipass-amd64.sh    # AMD64

# Daily use
./scripts/vm-manager.sh sync all      # Deploy
./scripts/vm-manager.sh run arm64     # Test

# Management
./scripts/vm-manager.sh list          # List VMs
./scripts/vm-manager.sh help          # Get help
```

---

## 🚀 Ready to Start!

**Quick setup:**
```bash
cd ravact-go/scripts
./setup-multipass-amd64.sh
```

**Then test:**
```bash
./vm-manager.sh run amd64
```

**That's it!** You can now test on both ARM64 and AMD64! 🎉

---

## 💡 Tips

1. **Use ARM64 for speed** - Daily development
2. **Test on AMD64 before release** - Ensure compatibility
3. **Use VM Manager** - Simplifies everything
4. **Create snapshots** - Before big changes
5. **VS Code Remote** - Best editing experience

---

**Need help?**
```bash
./scripts/vm-manager.sh help
cat scripts/MULTIPASS_GUIDE.md
```

**Happy multi-architecture testing! 🚀**
