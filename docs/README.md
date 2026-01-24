# Ravact Documentation

Welcome to the Ravact documentation! This guide will help you get started with Ravact and understand all its features.

---

## 📖 Table of Contents

### 🚀 Getting Started
New to Ravact? Start here!

- **[Quickstart Guide](getting-started/QUICKSTART.md)** - Get up and running in 5 minutes

### 🧪 Testing
Test Ravact on different platforms.

- **[M1 Mac Testing with Multipass](testing/M1_MAC_MULTIPASS_TESTING.md)** - Complete guide for testing on Apple Silicon Macs using Multipass VMs
- **[AMD64/Intel Testing](testing/AMD64_INTEL_TESTING.md)** - Comprehensive guide for testing on real AMD64/Intel hardware

### ⚙️ Setup & Installation
Installation and setup guides.

- **[Setup Scripts Guide](setup/SETUP_SCRIPTS_GUIDE.md)** - Understanding and using automated setup scripts

### ✨ Features
Learn about specific Ravact features.

- **[Database Management](features/DATABASE_MANAGEMENT.md)** - MySQL and PostgreSQL management guide
- **[PHP-FPM & Supervisor](features/PHPFPM_SUPERVISOR_GUIDE.md)** - PHP-FPM pool and Supervisor program management
- **[FrankenPHP Guide](features/FRANKENPHP_GUIDE.md)** - Using FrankenPHP with Ravact
- **[FrankenPHP Classic Setup](features/FRANKENPHP_CLASSIC_SETUP.md)** - Classic mode setup with systemd
- **[User Management](features/TEST_USER_MANAGEMENT.md)** - User and sudo management guide

### 💻 Development
Contributing to Ravact or building from source.

- **[Development Guide](development/DEVELOPMENT.md)** - Set up development environment and contribute
- **[Scripts README](scripts/SCRIPTS_README.md)** - Overview of automation scripts
- **[Release Guide](releasing/RELEASE_GUIDE.md)** - Release process and install script

### 🆘 Troubleshooting
Having issues? Check these guides.

- **[Troubleshooting Guide](troubleshooting/TROUBLESHOOTING.md)** - Common issues and solutions
- **[macOS Limitations](troubleshooting/MACOS_LIMITATIONS.md)** - Known limitations on macOS

### 📊 Project Information
Project status and version history.

- **[Project Status](project/PROJECT_STATUS.md)** - Current development status
- **[Changelog](project/CHANGELOG.md)** - Version history and changes

---

## 🎯 Quick Navigation

### For First-Time Users
1. **Installing Ravact?** → [Quickstart Guide](getting-started/QUICKSTART.md)
2. **Testing on M1 Mac?** → [M1 Mac Testing](testing/M1_MAC_MULTIPASS_TESTING.md)
3. **Need help?** → [Troubleshooting](troubleshooting/TROUBLESHOOTING.md)

### For Developers
1. **Want to contribute?** → [Development Guide](development/DEVELOPMENT.md)
2. **Testing on AMD64?** → [AMD64 Testing Guide](testing/AMD64_INTEL_TESTING.md)
3. **Using scripts?** → [Scripts README](scripts/SCRIPTS_README.md)

### For Configuration Management
Access all features via: **Main Menu → Configurations**

1. **Nginx Sites** → Configurations → Nginx Web Server
2. **Redis** → Configurations → Redis Cache
3. **MySQL** → Configurations → MySQL Database
4. **PostgreSQL** → Configurations → PostgreSQL Database
5. **PHP-FPM Pools** → Configurations → PHP-FPM Pools
6. **Supervisor** → Configurations → Supervisor
7. **SSL Certificates** → Nginx Configuration → SSL Options

---

## 📋 Feature Status

### ✅ Fully Implemented

**Core Management:**
- Setup Automation - Install and configure server components
- Nginx Management - Complete site and SSL management
- Redis Configuration - Password, port, and connection testing
- User Management - Add/remove users with sudo access

**Database Management:**
- MySQL - Password, port, database creation, user management
- PostgreSQL - Password, port, performance tuning, database creation

**Application Management:**
- PHP-FPM - Pool management, worker tuning, PM modes
- Supervisor - Program management, XML-RPC configuration

### 🎨 User Interface
- Terminal UI (TUI) with Bubble Tea framework
- Intuitive navigation with keyboard shortcuts
- Real-time status updates
- Secure password input (masked)
- Error handling and validation

### 🔧 Supported Software
- **Web Servers:** Nginx
- **Databases:** MySQL, PostgreSQL
- **PHP:** PHP-FPM (multiple versions), FrankenPHP
- **Caching:** Redis, Dragonfly
- **Process Management:** Supervisor
- **JavaScript:** Node.js
- **Version Control:** Git
- **SSL:** Certbot (Let's Encrypt)
- **Editors:** nano, vi/vim

---

## 🏗️ Architecture

### Target Platform
- **Operating System:** Linux only (Ubuntu 24.04 LTS recommended)
- **Architectures:** AMD64 (x86_64), ARM64 (aarch64)
- **System Requirements:** systemd, sudo, standard Linux utilities

### Why Linux Only?
Ravact directly manages Linux system services using:
- `systemd` for service management
- Linux-specific configuration paths
- Native Linux package managers (apt)
- Linux filesystem structure

**For macOS users:** Use Multipass to run Ubuntu VMs (see [M1 Mac Testing Guide](testing/M1_MAC_MULTIPASS_TESTING.md))

---

## 🚦 Testing Workflow

### For Apple Silicon (M1/M2/M3) Mac Users

1. **Install Multipass:**
   ```bash
   brew install --cask multipass
   ```

2. **Create Ubuntu VM:**
   ```bash
   multipass launch 24.04 --name ravact-test --memory 4G --cpus 2
   ```

3. **Build and Test:**
   ```bash
   GOOS=linux GOARCH=arm64 go build -o ravact-linux-arm64 ./cmd/ravact
   multipass transfer ravact-linux-arm64 ravact-test:~/ravact
   multipass shell ravact-test
   sudo ./ravact
   ```

**See:** [Complete M1 Testing Guide](testing/M1_MAC_MULTIPASS_TESTING.md)

### For AMD64/Intel Hardware

1. **Provision Ubuntu Server:**
   - VPS (DigitalOcean, Linode, Vultr)
   - Cloud (AWS, Google Cloud, Azure)
   - Bare metal server
   - Local VM (VirtualBox, VMware)

2. **Install Dependencies:**
   ```bash
   apt install -y mysql-server postgresql php8.3-fpm supervisor nginx
   ```

3. **Build and Run:**
   ```bash
   go build -o ravact ./cmd/ravact
   sudo ./ravact
   ```

**See:** [Complete AMD64 Testing Guide](testing/AMD64_INTEL_TESTING.md)

---

## 📚 Additional Resources

### Documentation
- **Main README:** [../README.md](../README.md) - Project overview and quick start
- **Feature Guides:** [features/](features/) - Detailed feature documentation
- **Testing Guides:** [testing/](testing/) - Platform-specific testing instructions

### External Links
- **Repository:** https://github.com/iperamuna/ravact
- **Go Language:** https://go.dev/
- **Bubble Tea:** https://github.com/charmbracelet/bubbletea
- **Multipass:** https://multipass.run/

---

## 🤝 Contributing

We welcome contributions! See the [Development Guide](development/DEVELOPMENT.md) for:
- Setting up your development environment
- Code style and conventions
- Testing requirements
- Pull request process

---

## 📄 License

See the main repository for license information.

---

## 💬 Support

- **Issues:** Report bugs or request features on GitHub
- **Documentation:** Check this documentation for guides and troubleshooting
- **Community:** Contribute to discussions and share your experience

---

**Last Updated:** January 2026
