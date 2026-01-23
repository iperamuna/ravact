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

### Quick Start (Linux)

```bash
# Download the latest release for your architecture
wget https://github.com/iperamuna/ravact/releases/latest/download/ravact-linux-amd64

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
cd ravact/ravact-go

# Build for your platform
make build

# Or build for specific platform
make build-linux-amd64
make build-linux-arm64

# Binary will be in dist/ folder
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
- [Quickstart Guide](docs/getting-started/QUICKSTART.md)

### Testing
- [M1 Mac Testing with Multipass](docs/testing/M1_MAC_MULTIPASS_TESTING.md) - Test on Apple Silicon
- [AMD64/Intel Testing](docs/testing/AMD64_INTEL_TESTING.md) - Test on real hardware

### Features
- [Database Management Guide](docs/features/DATABASE_MANAGEMENT.md) - MySQL & PostgreSQL
- [PHP-FPM & Supervisor Guide](docs/features/PHPFPM_SUPERVISOR_GUIDE.md)
- [FrankenPHP Guide](docs/features/FRANKENPHP_GUIDE.md)
- [User Management Guide](docs/features/TEST_USER_MANAGEMENT.md)

### Setup & Installation
- [Docker Setup](docs/setup/DOCKER_SETUP.md)
- [AMD64 Setup](docs/setup/AMD64_SETUP_SUMMARY.md)
- [Dev VM Setup](docs/setup/DEV_VM_SETUP.md)
- [Setup Scripts Guide](docs/setup/SETUP_SCRIPTS_GUIDE.md)

### Features
- [FrankenPHP Guide](docs/features/FRANKENPHP_GUIDE.md)
- [User Management](docs/features/TEST_USER_MANAGEMENT.md)

### Development
- [Development Guide](docs/development/DEVELOPMENT.md)
- [Build Summary](docs/development/BUILD_SUMMARY.md)
- [Docker Workflow](docs/development/DOCKER_WORKFLOW.md)

### Testing
- [Quick Test](docs/testing/QUICK_TEST.md)
- [Test Report](docs/testing/TEST_REPORT.md)
- [VM Test Instructions](docs/testing/VM_TEST_INSTRUCTIONS.md)

### Troubleshooting
- [Troubleshooting Guide](docs/troubleshooting/TROUBLESHOOTING.md)
- [macOS Limitations](docs/troubleshooting/MACOS_LIMITATIONS.md)

### Project Information
- [Project Status](docs/project/PROJECT_STATUS.md)
- [Changelog](docs/project/CHANGELOG.md)

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

**Version**: 0.1.0

**Current Features**:
- ✅ Complete software setup automation (13 packages)
- ✅ Full Nginx site management with SSL
- ✅ Redis configuration
- ✅ User management
- ✅ Editor integration (nano/vi)

**Upcoming Features**:
- 🔄 Supervisor configuration
- 🔄 MySQL database management
- 🔄 PostgreSQL database management
- 🔄 PHP-FPM pool configuration

---

**⭐ Star this repository if you find it useful!**

For detailed documentation, visit the [docs](docs/) folder or check out the [full documentation index](docs/README.md).
