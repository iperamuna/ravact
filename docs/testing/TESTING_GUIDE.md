# Testing Guide

[← Back to Documentation](../README.md)

This guide covers how to test Ravact on different platforms.

## Quick Start Testing

### On Ubuntu/Debian (Native)

```bash
# Build
go build -o ravact ./cmd/ravact

# Run (requires root for full functionality)
sudo ./ravact
```

### On macOS with Multipass (Recommended for Mac users)

```bash
# Install Multipass
brew install --cask multipass

# Create Ubuntu VM
multipass launch 24.04 --name ravact-test --memory 4G --cpus 2 --disk 20G

# Build for Linux ARM64 (Apple Silicon) or AMD64 (Intel Mac)
GOOS=linux GOARCH=arm64 go build -o ravact-linux ./cmd/ravact
# OR for Intel Mac:
# GOOS=linux GOARCH=amd64 go build -o ravact-linux ./cmd/ravact

# Transfer and run
multipass transfer ravact-linux ravact-test:/home/ubuntu/ravact
multipass shell ravact-test
sudo ./ravact
```

### On Cloud/VPS (AMD64)

```bash
# On your local machine, build for Linux
GOOS=linux GOARCH=amd64 go build -o ravact-linux ./cmd/ravact

# Transfer to server
scp ravact-linux user@server:/home/user/

# SSH and run
ssh user@server
sudo ./ravact
```

## Test Checklist

### Core Features

| Feature | Test Steps | Expected Result |
|---------|------------|-----------------|
| **Main Menu** | Launch app, navigate with arrows | Menu displays, navigation works |
| **Install Software** | Package Management → Install Software | Shows available packages |
| **Service Config** | Service Configuration → Service Settings | Shows installed services |
| **Developer Toolkit** | Site Management → Developer Toolkit | Shows command categories |
| **File Browser** | Tools → File Browser | Shows current directory |
| **User Management** | System Administration → User Management | Shows users list |
| **Quick Commands** | System Administration → Quick Commands | Shows system commands |

### UI Features

| Feature | Test Steps | Expected Result |
|---------|------------|-----------------|
| **Navigation** | Use ↑/↓/Enter/Esc | Smooth navigation |
| **Copy** | Press `c` on copyable screen | "Copied!" message appears |
| **Forms** | Add User → Fill form | Tab navigation, validation works |
| **File Browser Help** | File Browser → Press `?` | Help screen displays |
| **Search** | File Browser → Press `/` | Search mode activates |

### Service Tests (Requires Services Installed)

```bash
# Install test services first
sudo apt update
sudo apt install -y nginx mysql-server postgresql redis-server php8.3-fpm supervisor
```

| Service | Test |
|---------|------|
| **Nginx** | Add site, enable/disable, test config |
| **MySQL** | Change password, change port |
| **PostgreSQL** | Change password, change port |
| **Redis** | Change password, change port, test connection |
| **PHP-FPM** | View pools, restart service |
| **Supervisor** | Add program, view status |

## Multipass VM Management

```bash
# List VMs
multipass list

# Start/Stop VM
multipass start ravact-test
multipass stop ravact-test

# Shell into VM
multipass shell ravact-test

# Delete VM when done
multipass delete ravact-test
multipass purge
```

## Troubleshooting

### "Not running as root" warning
Run with sudo: `sudo ./ravact`

### Services not showing in config menu
Install the required services first, then refresh.

### Terminal colors look wrong
- Check `TERM` environment variable
- Try: `export TERM=xterm-256color`

### Copy not working
- Clipboard requires X11 or Wayland on Linux
- In SSH, use terminal's copy mode instead

## Platform Notes

### macOS
- UI works but setup features require Linux
- Use Multipass for full testing

### WSL2
- Works with some limitations
- Systemd required for service management

### Docker
- Not recommended (no systemd)
- Use Multipass or real VMs instead

---

[← Back to Documentation](../README.md)
