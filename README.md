# Ravact - Linux Server Management Tool

**Ravact** is a modern, interactive TUI (Terminal User Interface) application for managing Linux servers. It provides an intuitive interface for common server administration tasks including software installation, user management, and service configuration.

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux-orange.svg)](https://www.linux.org/)

![Ravact Main Menu](docs/assets/screenshot.png)

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
- **Passwordless Users** - Create users with SSH key-only authentication (no password)
- **Passwordless Sudo (NOPASSWD)** - Grant sudo without password prompts
- **Passwordless su** - Switch to user without password
- **User Details** - View user information and group memberships

### ⚡ Quick Commands
- **System Info** - Display kernel and architecture information
- **Disk Usage** - Show filesystem usage
- **Memory Info** - Show RAM usage statistics
- **Running Services** - List active systemd services
- **Network Info** - Display network interfaces and IP addresses
- **Top Processes** - Show CPU-sorted process list
- **Recent Logs** - View recent system journal entries

### 🛠️ Developer Toolkit (NEW)
- **34+ Essential Commands** - Frequently forgotten terminal commands at your fingertips
- **Laravel Commands** - Tail logs, fix permissions, generate APP_KEY, check queue workers
- **WordPress Commands** - Fix permissions, clear cache, generate salts, find malware patterns
- **PHP Commands** - Check version, list modules, find php.ini, check OPcache
- **Security Commands** - Scan for malware, find world-writable files, check open ports
- **Copy to Clipboard** - Press `c` to copy any command instantly

### 📁 File Browser (NEW)
- **Full-Featured File Manager** - Navigate, preview, and manage files without leaving the app
- **File Operations** - Copy, cut, paste, delete, rename, create files/directories
- **File Preview** - View text files with syntax highlighting and line numbers
- **Search & Filter** - Find files quickly with live filtering
- **Keyboard-Driven** - Vim-like navigation (h/j/k/l) plus standard arrows
- **Help Screen** - Press `?` for complete keyboard shortcuts reference

### 🎨 Modern UI/UX (NEW)
- **Categorized Menus** - Logically organized menu items (Package Management, Service Configuration, Site Management, System Administration, Tools)
- **Beautiful Forms** - Powered by [huh](https://github.com/charmbracelet/huh) with custom theme
- **xterm.js Compatible** - Works perfectly in web-based terminals
- **Copy Support** - Press `c` on most screens to copy content to clipboard
- **Terminal-Aware** - Auto-detects terminal capabilities and adjusts colors/symbols

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
   - Main Menu → Package Management → Install Software
   - Select package (e.g., Nginx)
   - Choose "Install"

4. **Configure Services**:
   - Main Menu → Service Configuration → Service Settings
   - Select service (e.g., Redis Cache)
   - Manage settings

5. **Use Developer Toolkit**:
   - Main Menu → Site Management → Developer Toolkit
   - Navigate categories with `Tab` or arrow keys
   - Press `c` to copy command, `Enter` to execute

6. **Browse Files**:
   - Main Menu → Tools → File Browser
   - Press `?` for full keyboard shortcuts help

## ⌨️ Keyboard Shortcuts

### Global
| Key | Action |
|-----|--------|
| `↑`/`↓` or `j`/`k` | Navigate up/down |
| `Enter` | Select/Confirm |
| `Esc` | Go back |
| `q` | Quit |
| `c` | Copy to clipboard (where supported) |

### File Browser
| Key | Action |
|-----|--------|
| `?` | Show help screen |
| `Space` | Toggle selection |
| `y`/`x`/`p` | Copy/Cut/Paste |
| `n`/`N` | New file/directory |
| `d` | Delete |
| `r` | Rename |
| `/` | Search |
| `.` | Toggle hidden files |

### Forms (huh)
| Key | Action |
|-----|--------|
| `Tab`/`Shift+Tab` | Navigate fields |
| `Enter` | Submit/Select |
| `↑`/`↓` | Change option (in selects) |

## 📚 Documentation

Comprehensive documentation is available in the [docs](docs/) directory.

👉 **[View Full Documentation](docs/README.md)**

### Quick Links

| Category | Guides |
|----------|--------|
| **Getting Started** | [Quick Start](docs/getting-started/QUICKSTART.md) • [Testing Guide](docs/testing/TESTING_GUIDE.md) |
| **Features** | [Developer Toolkit](docs/features/DEVELOPER_TOOLKIT.md) • [File Browser](docs/features/FILE_BROWSER.md) • [Database Management](docs/features/DATABASE_MANAGEMENT.md) |
| **UI** | [UI Guide](docs/ui/UI_GUIDE.md) • [Keyboard Shortcuts](docs/ui/KEYBOARD_SHORTCUTS.md) |
| **Development** | [Development Guide](docs/development/DEVELOPMENT.md) • [Setup Scripts](docs/setup/SETUP_SCRIPTS_GUIDE.md) |
| **Help** | [Troubleshooting](docs/troubleshooting/TROUBLESHOOTING.md) • [macOS Limitations](docs/troubleshooting/MACOS_LIMITATIONS.md) |
| **Project** | [Changelog](docs/project/CHANGELOG.md) • [Project Status](docs/project/PROJECT_STATUS.md) • [TODO/Roadmap](docs/project/TODO.md) |

## 🛠️ Technology Stack

- **Language**: Go 1.24+
- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Styling**: [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **Forms**: [huh](https://github.com/charmbracelet/huh) - Beautiful, customizable forms
- **Clipboard**: [clipboard](https://github.com/atotto/clipboard) - Cross-platform clipboard support
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

**Version**: 0.3.1

**All Features Implemented**:
- ✅ Complete software setup automation (13+ packages)
- ✅ Full Nginx site management with SSL (Let's Encrypt & manual)
- ✅ Redis configuration (password, port, connection testing)
- ✅ MySQL database management (password, port, databases, users)
- ✅ PostgreSQL database management (password, port, performance tuning)
- ✅ PHP-FPM pool management (view pools, service control)
- ✅ Supervisor configuration (programs, XML-RPC)
- ✅ Firewall management
- ✅ Passwordless user management with NOPASSWD sudo
- ✅ Editor integration (nano/vi)
- ✅ Quick system commands

**New in v0.3.0**:
- ✅ Git System User (`meta.systemuser`) - automatic user tracking per repository
- ✅ Laravel App menu with .env creation, scheduler setup, artisan commands
- ✅ Passwordless user creation (SSH key-only authentication)
- ✅ Passwordless sudo (NOPASSWD) and su support
- ✅ FrankenPHP Socket/Port selection for service creation
- ✅ NPM build runs `npm install && npm run build` as system user
- ✅ Removed FrankenPHP/Node.js from Setup Install (available via Site Commands)

**New in v0.2.0**:
- ✅ Developer Toolkit with 34+ essential Laravel/WordPress/PHP/Security commands
- ✅ Full-featured File Browser with preview, search, and file operations
- ✅ Beautiful forms powered by huh with custom theme
- ✅ Categorized menu organization following industry standards
- ✅ Copy to clipboard support across all screens
- ✅ xterm.js compatibility for web-based terminals
- ✅ Terminal capability detection with graceful fallbacks
- ✅ Keyboard shortcuts help screen (press `?` in File Browser)

---

**⭐ Star this repository if you find it useful!**

For detailed documentation, visit the [docs](docs/) folder or check out the [full documentation index](docs/README.md).
