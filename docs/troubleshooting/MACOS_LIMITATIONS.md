# macOS Limitations

[← Back to Documentation](../README.md)

## Overview

Ravact is designed for **Linux server management** and has limited functionality when run directly on macOS. This is by design, as the tool manages Linux-specific services and configurations.

## What Works on macOS

✅ **Building the application**
- You can build ravact on macOS for testing the UI
- Cross-compilation to Linux works perfectly

✅ **UI Testing**
- Main menu navigation
- Theme and layout testing
- Basic screen flows

## What Doesn't Work on macOS

❌ **User Management**
- macOS uses a different user management system (Directory Services)
- `/etc/passwd` and `/etc/group` are managed differently
- The `groups` command behaves differently
- **Solution:** Deploy to Linux VM to use this feature

❌ **Nginx Configuration**
- Nginx paths are different on macOS
- Site configuration structure is different
- **Solution:** Use on actual Linux server

❌ **Setup Scripts**
- All 12 setup scripts are designed for Ubuntu/Debian
- Package managers (apt, yum) not available
- **Solution:** Deploy to Linux VM

❌ **System Information**
- Linux-specific system calls
- Service management (systemd)
- **Solution:** Run on Linux

## Recommended Development Setup

### Option 1: Multipass VM (Easiest) ⭐

```bash
# Quick setup with our script
cd ravact-go/scripts
./setup-multipass.sh

# Develop on Mac, test on Linux VM
./quick-deploy.sh
```

### Option 2: UTM VM

```bash
# Manual setup
cd ravact-go/scripts
./setup-mac-vm.sh
```

### Option 3: Docker Testing

```bash
# Quick test in Ubuntu container
make docker-shell

# Inside container
apt update && apt install -y golang-go
go build -o ravact ./cmd/ravact
sudo ./ravact
```

## Why These Limitations Exist

1. **Different User Systems**
   - Linux: `/etc/passwd`, `/etc/group`, `useradd`, `usermod`
   - macOS: Directory Services (dscl), different user database

2. **Different Package Managers**
   - Linux: `apt`, `yum`, `dnf`
   - macOS: `brew` (Homebrew)

3. **Different Service Management**
   - Linux: `systemd`, `service` commands
   - macOS: `launchd`, `launchctl`

4. **Different File System Structure**
   - Linux: `/etc/nginx/`, `/var/www/`
   - macOS: Different paths, different conventions

## Development Workflow

### Daily Development on macOS

```bash
# 1. Edit code on Mac (any editor)
vim internal/ui/screens/main_menu.go

# 2. Test UI locally (limited features)
go run ./cmd/ravact

# 3. Build for Linux
make build-linux-arm64

# 4. Deploy to Linux VM and test fully
./scripts/quick-deploy.sh
ssh ravact-dev
cd ravact-go
sudo ./ravact
```

## What You'll See on macOS

### User Management Screen
- Shows: "⚠️ macOS Not Supported"
- Message: "This feature requires Linux (Ubuntu/Debian/RHEL)"
- Instruction: "Deploy to Linux VM to use this feature"

### Other Features
- May show errors or warnings
- Will not function as intended
- Some features might crash

## Testing on macOS

### UI Testing (Works)
```bash
# Build and run
go build -o ravact ./cmd/ravact
./ravact

# You can test:
- Navigation
- Menu layouts
- Theme rendering
- Screen transitions
```

### Unit Tests (Mostly Work)
```bash
# Run tests
make test

# Some tests are skipped on macOS
# See: *_darwin_test.go files
```

### Integration Tests (Don't Work)
```bash
# These require Linux
make test-integration

# Use Docker or VM instead
make docker-test
```

## Solutions Summary

| Issue | Solution |
|-------|----------|
| User Management hangs | ✅ **Fixed** - Now shows warning instead of hanging |
| Setup scripts don't work | Use Linux VM or server |
| Nginx config fails | Use Linux VM or server |
| Can't test full features | Use Multipass/UTM VM |
| Want quick testing | Use `make docker-shell` |

## Recent Fixes

### v0.1.1 (2026-01-23)
- ✅ Fixed user management hanging on macOS
- ✅ Added async loading with timeout
- ✅ Shows helpful warning message on macOS
- ✅ Added proper error handling
- ✅ No more infinite "Loading..." state

### How the Fix Works

1. **Async Loading**: User data loads in background (non-blocking)
2. **Timeout**: 2-second timeout on shell commands
3. **macOS Detection**: Detects Darwin and shows warning
4. **Graceful Degradation**: App doesn't crash, shows helpful message

## Get Full Functionality

To use ravact with all features:

1. **Setup a Linux VM** (5 minutes with Multipass):
   ```bash
   cd ravact-go/scripts
   ./setup-multipass.sh
   ```

2. **Deploy and test**:
   ```bash
   ./quick-deploy.sh
   ssh ravact-dev
   cd ravact-go
   sudo ./ravact
   ```

3. **All features work perfectly on Linux!** 🚀

## Questions?

- Check [DEV_VM_SETUP.md](DEV_VM_SETUP.md) for VM setup
- Check [DEVELOPMENT.md](DEVELOPMENT.md) for development guide
- Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common issues

---

**TL;DR**: Ravact is for Linux servers. Develop on Mac, test on Linux VM. Use our setup scripts for easy VM creation!
