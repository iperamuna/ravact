# Ravact - Linux Server Management Tool

**Ravact** is a modern, interactive TUI (Terminal User Interface) application for managing Linux servers. It provides an intuitive interface for common server administration tasks including software installation, user management, and service configuration.

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux-orange.svg)](https://www.linux.org/)

## ✨ Features

> **Status:** All features are fully implemented and production-ready! 🎉

### 🚀 Software Setup & Management
- **One-Click Installation** - Install popular server software with a single command
- **13 Pre-configured Packages** - Nginx, MySQL, PostgreSQL, Redis, PHP, Node.js, and more
- **Embedded Scripts** - No external dependencies, everything runs from a single binary
- **Installation Status** - View and manage installed applications
- **Automatic Detection** - Shows only installed services in configuration menu

### 🌐 Nginx Web Server Management
- **Site Management** - List, add, edit, and delete Nginx virtual hosts
- **7 Site Templates** - Static HTML, PHP, Laravel, WordPress, Symfony, Node.js, Reverse Proxy
- **SSL Certificate Management**
  - Let's Encrypt (automatic SSL with certbot)
  - Manual certificates (provide your own cert files)
  - SSL removal and configuration
- **Editor Integration** - Edit configs with nano or vi directly in Ravact

### 🔧 Service Configuration
- **Redis Cache** - Configure authentication, port, test connections
- **MySQL Database** - Change root password, configure port, create databases, list databases, service management
- **PostgreSQL Database** - Change postgres password, configure port, performance tuning (max_connections, shared_buffers), create databases, service management
- **PHP-FPM Pools** - View pools, service restart/reload, pool details
- **Supervisor** - Program management, XML-RPC configuration (IP, port, username, password), add programs with editor selection (nano/vi) and config validation

### 👥 User Management
- **Create/Delete Users** - Manage system users with home directories
- **Sudo Access** - Grant or revoke sudo privileges
- **User Details** - View user information and group memberships

### ⚡ Quick Commands
- **System Info** - Display kernel and architecture information
- **Disk Usage** - Show filesystem usage
- **Memory Info** - Show RAM usage statistics
- **Running Services** - List active systemd services
- **Network Info** - Display network interfaces and IP addresses
- **Top Processes** - Show CPU-sorted process list
- **Recent Logs** - View recent system journal entries

### 🔍 Smart Service Detection
- Configuration menu automatically detects installed services
- Uninstalled services are grayed out and not selectable
- Shows `[Not Installed]` indicator for unavailable services
- Prevents errors from attempting to configure missing services

## 📦 Installation

### Recommended (One-Command Install)

```bash
curl -sSL https://raw.githubusercontent.com/iperamuna/ravact/main/scripts/install.sh | sudo bash
```

Installs the correct binary for your OS/architecture to `/usr/local/bin/ravact`.

### Manual Download (Linux)

```bash
# Download the latest release for your architecture
curl -L https://github.com/iperamuna/ravact/releases/latest/download/ravact-linux-amd64 -o ravact-linux-amd64

# Make it executable
chmod +x ravact-linux-amd64

# Run it
sudo ./ravact-linux-amd64
```

### Available Binaries

- `ravact-linux-amd64` - Linux x86_64 (Intel/AMD)
- `ravact-linux-arm64` - Linux ARM64 (Raspberry Pi, Apple Silicon VMs)
- `ravact-darwin-arm64` - macOS Apple Silicon (UI only, setup features require Linux)
- `ravact-darwin-amd64` - macOS Intel (UI only, setup features require Linux)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/iperamuna/ravact.git
cd ravact

# Build for your current platform
make build

# Or build for all platforms (Linux & macOS, amd64 & arm64)
make build-all

# Or build for specific platforms
make build-linux        # Linux amd64
make build-linux-arm64  # Linux arm64
make build-darwin       # macOS amd64
make build-darwin-arm64 # macOS arm64 (Apple Silicon)

# Binaries will be in dist/ folder
```

## 🎯 Quick Start Guide

1. **Run Ravact** (requires root for installation features):
   ```bash
   sudo ./ravact
   ```

2. **Navigate the Menu**:
   - Use `↑`/`↓` arrow keys to navigate
   - Press `Enter` to select
   - Press `Esc` to go back
   - Press `q` to quit

3. **Install Software**:
   - Main Menu → Setup
   - Select package (e.g., Nginx)
   - Choose "Install"

4. **Configure Services**:
   - Main Menu → Configurations
   - Select service (e.g., Redis Cache)
   - Manage settings

5. **Manage Nginx Sites**:
   - Main Menu → Configurations → Nginx Web Server
   - Press `a` to add a site
   - Press `Enter` on existing site to manage

## 📚 Documentation

Comprehensive documentation is available in the [docs](docs/) directory:

### Getting Started
- [Quickstart Guide](docs/getting-started/QUICKSTART.md) - Get up and running in 5 minutes

### Features
- [Database Management](docs/features/DATABASE_MANAGEMENT.md) - MySQL & PostgreSQL configuration
- [PHP-FPM & Supervisor](docs/features/PHPFPM_SUPERVISOR_GUIDE.md) - Pool and process management
- [FrankenPHP Guide](docs/features/FRANKENPHP_GUIDE.md) - Modern PHP server setup
- [User Management](docs/features/TEST_USER_MANAGEMENT.md) - System user administration

### Testing
- [M1 Mac Testing](docs/testing/M1_MAC_MULTIPASS_TESTING.md) - Test on Apple Silicon with Multipass
- [AMD64/Intel Testing](docs/testing/AMD64_INTEL_TESTING.md) - Test on real hardware
- [Complete Test Cases](docs/testing/COMPLETE_TEST_CASES.md) - 40+ comprehensive test cases

### Setup & Scripts
- [Setup Scripts Guide](docs/setup/SETUP_SCRIPTS_GUIDE.md) - Automated setup scripts
- [Scripts Reference](docs/scripts/SCRIPTS_README.md) - All available scripts

### Development
- [Development Guide](docs/development/DEVELOPMENT.md) - Contributing and building from source
- [Release Guide](docs/releasing/RELEASE_GUIDE.md) - Release process

### Troubleshooting
- [Troubleshooting Guide](docs/troubleshooting/TROUBLESHOOTING.md) - Common issues and solutions
- [macOS Limitations](docs/troubleshooting/MACOS_LIMITATIONS.md) - Known limitations on macOS

### Project
- [Project Status](docs/project/PROJECT_STATUS.md) - Current development status
- [Changelog](docs/project/CHANGELOG.md) - Version history

## 🛠️ Technology Stack

- **Language**: Go 1.24+
- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Styling**: [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **Platform**: Linux (Ubuntu, Debian, RHEL, CentOS)

## 🔒 Security Considerations

- **Root Access**: Required for installation and system configuration
- **Password Management**: Passwords are masked in the UI
- **Configuration Files**: Direct manipulation of system configs (backups recommended)
- **SSL Certificates**: Automated Let's Encrypt or manual certificate management

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues, feature requests, or pull requests.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 👤 Author

**Indunil Peramuna**

- GitHub: [@iperamuna](https://github.com/iperamuna)

## 🙏 Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Amazing TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Beautiful terminal styling
- All contributors and testers

## 📊 Project Status

**Version**: 0.1.3

**All Features Implemented**:
- ✅ Complete software setup automation (13+ packages)
- ✅ Full Nginx site management with SSL (Let's Encrypt & manual)
- ✅ Redis configuration (password, port, connection testing)
- ✅ MySQL database management (password, port, databases, users)
- ✅ PostgreSQL database management (password, port, performance tuning)
- ✅ PHP-FPM pool management (view pools, service control)
- ✅ Supervisor configuration (programs, XML-RPC)
- ✅ Firewall management
- ✅ User management with sudo access
- ✅ Editor integration (nano/vi)
- ✅ Quick system commands

---

**⭐ Star this repository if you find it useful!**

For detailed documentation, visit the [docs](docs/) folder or check out the [full documentation index](docs/README.md).
