# Ravact Documentation

Welcome to the Ravact documentation! This guide will help you get started with Ravact and understand all its features.

## 📖 Table of Contents

### 🚀 Getting Started
New to Ravact? Start here!

- **[Quickstart Guide](getting-started/QUICKSTART.md)** - Get up and running in 5 minutes
- **[Quick Summary](getting-started/QUICK_SUMMARY.md)** - Overview of features and capabilities
- **[Docker Quick Start](getting-started/QUICK_DOCKER_START.md)** - Run Ravact in Docker for testing

### ⚙️ Setup & Installation
Detailed installation guides for different environments.

- **[Docker Setup](setup/DOCKER_SETUP.md)** - Complete Docker installation guide
- **[AMD64 Setup Summary](setup/AMD64_SETUP_SUMMARY.md)** - Intel/AMD 64-bit Linux setup
- **[Development VM Setup](setup/DEV_VM_SETUP.md)** - Set up a development VM
- **[Setup Scripts Guide](setup/SETUP_SCRIPTS_GUIDE.md)** - Understanding the setup scripts

### ✨ Features
Learn about specific Ravact features.

- **[FrankenPHP Guide](features/FRANKENPHP_GUIDE.md)** - Using FrankenPHP with Ravact
- **[User Management](features/TEST_USER_MANAGEMENT.md)** - User and sudo management guide

### 💻 Development
Contributing to Ravact or building from source.

- **[Development Guide](development/DEVELOPMENT.md)** - Set up development environment
- **[Build Summary](development/BUILD_SUMMARY.md)** - Build process and compilation
- **[Docker Workflow](development/DOCKER_WORKFLOW.md)** - Development with Docker
- **[Fixes Applied](development/FIXES_APPLIED.md)** - History of bug fixes and improvements

### 🧪 Testing
Testing guides and reports.

- **[Quick Test](testing/QUICK_TEST.md)** - Quick testing checklist
- **[Test Report](testing/TEST_REPORT.md)** - Comprehensive test results
- **[Real AMD64 Testing](testing/REAL_AMD64_TESTING.md)** - Testing on real AMD64 hardware
- **[VM Test Instructions](testing/VM_TEST_INSTRUCTIONS.md)** - Testing in virtual machines

### 🆘 Troubleshooting
Having issues? Check these guides.

- **[Troubleshooting Guide](troubleshooting/TROUBLESHOOTING.md)** - Common issues and solutions
- **[macOS Limitations](troubleshooting/MACOS_LIMITATIONS.md)** - Known limitations on macOS

### 🔧 Scripts & Utilities
Documentation for helper scripts.

- **[Scripts README](scripts/SCRIPTS_README.md)** - Overview of utility scripts
- **[Multipass Guide](scripts/MULTIPASS_GUIDE.md)** - Using Multipass for VMs
- **[UTM Troubleshooting](scripts/UTM_TROUBLESHOOTING.md)** - UTM VM issues
- **[VM Setup README](scripts/VM_SETUP_README.md)** - Virtual machine setup guide

### 📊 Project Information
Project status, roadmap, and changes.

- **[Project Status](project/PROJECT_STATUS.md)** - Current development status
- **[Changelog](project/CHANGELOG.md)** - Version history and changes

---

## 🎯 Quick Navigation

### For Users
1. **First time?** → [Quickstart Guide](getting-started/QUICKSTART.md)
2. **Installing on server?** → [Setup Guides](setup/)
3. **Need help?** → [Troubleshooting](troubleshooting/TROUBLESHOOTING.md)

### For Developers
1. **Want to contribute?** → [Development Guide](development/DEVELOPMENT.md)
2. **Building from source?** → [Build Summary](development/BUILD_SUMMARY.md)
3. **Testing changes?** → [Testing Guides](testing/)

### For Specific Features
1. **Nginx Sites** → Built-in, see Main Menu → Configurations → Nginx
2. **Redis Config** → Built-in, see Main Menu → Configurations → Redis
3. **User Management** → Built-in, see Main Menu → User Management
4. **SSL Certificates** → Nginx Configuration → Add/Manage SSL

---

## 📋 Feature Status

### ✅ Fully Implemented
- **Setup Automation** - Install 13 software packages
- **Nginx Management** - Complete site and SSL management
- **Redis Configuration** - Password, port, connection testing
- **User Management** - Add/remove users, sudo access
- **Editor Integration** - nano and vi support

### 🔄 In Progress
- **Supervisor Configuration** - Process management (manager created)
- **MySQL Management** - Database and user management
- **PostgreSQL Management** - Database and role management
- **PHP Configuration** - PHP-FPM pool management

### 🎯 Planned
- Apache configuration support
- Firewall management UI
- Backup and restore tools
- System monitoring dashboard

---

## 🏗️ Architecture

Ravact is built with a clean, modular architecture:

```
ravact/
├── cmd/ravact/          # Main application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── executor/        # Script execution
│   ├── models/          # Data models
│   ├── setup/           # Setup logic
│   ├── system/          # System managers (nginx, redis, etc.)
│   └── ui/              # Terminal UI components
│       ├── components/  # Reusable UI components
│       ├── screens/     # Application screens
│       └── theme/       # Visual theming
├── assets/
│   ├── configs/         # Configuration templates
│   └── scripts/         # Setup scripts (embedded)
└── docs/                # Documentation (you are here!)
```

---

## 🎨 Screenshots

### Main Menu
```
┌─────────────────────────────────────────────┐
│      RAVACT v0.1.0 - Main Menu              │
│                                             │
│  ▶ Setup                                    │
│    Install server software packages         │
│                                             │
│    Installed Applications                   │
│    View and manage installed services       │
│                                             │
│    Configurations                           │
│    Manage service configurations            │
│                                             │
│    Quick Commands                           │
│    Execute common administrative tasks      │
│                                             │
│    User Management                          │
│    Manage users, groups, and sudo           │
└─────────────────────────────────────────────┘
```

### Nginx Site Management
- Interactive site creation with templates
- SSL certificate management
- Enable/disable sites with one click
- Edit configurations with nano or vi

### Redis Configuration
- Secure password configuration
- Port management
- Connection testing
- Service status monitoring

---

## 🔐 Security

- **Root Access Required** - For system-level operations
- **Password Security** - Passwords masked in UI, no plaintext logging
- **SSL Support** - Automated Let's Encrypt or manual certificates
- **Sudo Management** - Control which users have elevated privileges

---

## 🌟 Why Ravact?

### Traditional Approach
```bash
# Manual installation
sudo apt-get install nginx
sudo nano /etc/nginx/sites-available/mysite
sudo ln -s /etc/nginx/sites-available/mysite /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
sudo certbot --nginx -d mydomain.com
# ... many more commands
```

### With Ravact
```bash
sudo ./ravact
# Use arrow keys and Enter
# Visual interface guides you through everything
# No need to remember commands or file paths
```

**Benefits:**
- ✅ **Faster** - Complete tasks in seconds, not minutes
- ✅ **Safer** - Visual confirmation before executing
- ✅ **Easier** - No need to remember complex commands
- ✅ **Professional** - Clean, modern interface
- ✅ **Portable** - Single binary, no dependencies

---

## 🖥️ System Requirements

### Minimum Requirements
- **OS**: Linux (Ubuntu 20.04+, Debian 10+, RHEL 8+, CentOS 8+)
- **Architecture**: x86_64 (amd64) or ARM64 (aarch64)
- **RAM**: 512 MB minimum (1 GB recommended)
- **Disk**: 10 MB for binary, varies by installed software
- **Privileges**: Root access (sudo) required for installations

### Supported Distributions
- ✅ Ubuntu 20.04, 22.04, 24.04
- ✅ Debian 10, 11, 12
- ✅ RHEL/Rocky Linux 8, 9
- ✅ CentOS 8+
- ✅ Other systemd-based distributions

### macOS Support
- **UI**: ✅ Works for testing and development
- **Setup Features**: ❌ Requires Linux (use Docker or VM)
- See [macOS Limitations](troubleshooting/MACOS_LIMITATIONS.md)

---

## 🐛 Troubleshooting

Common issues and solutions:

### "Command not found" or "Permission denied"
```bash
# Make sure the binary is executable
chmod +x ravact

# Run with sudo for installation features
sudo ./ravact
```

### Setup scripts fail on macOS
Ravact setup scripts are designed for Linux. On macOS:
- Use Docker: `make docker-test`
- Use a Linux VM: See [VM Setup Guide](docs/scripts/VM_SETUP_README.md)
- Deploy to a Linux server

### Nginx configuration errors
- Test config: `sudo nginx -t`
- Check logs: `sudo tail -f /var/log/nginx/error.log`
- Use Ravact's built-in test feature

For more help, see the [Troubleshooting Guide](docs/troubleshooting/TROUBLESHOOTING.md).

---

## 🚀 Roadmap

### Version 0.2.0 (Next Release)
- [ ] Supervisor configuration UI
- [ ] MySQL database management
- [ ] PostgreSQL database management
- [ ] PHP-FPM pool management
- [ ] Enhanced monitoring dashboard

### Version 0.3.0
- [ ] Apache web server support
- [ ] Firewall management (UFW/iptables)
- [ ] Backup and restore tools
- [ ] Docker container management

### Future Versions
- [ ] Multi-server management
- [ ] Automated backup scheduling
- [ ] Performance monitoring
- [ ] Log viewer and analysis

---

## 💬 Support

- **Issues**: [GitHub Issues](https://github.com/iperamuna/ravact/issues)
- **Discussions**: [GitHub Discussions](https://github.com/iperamuna/ravact/discussions)
- **Documentation**: [docs/](docs/)

---

## 📈 Statistics

- **Lines of Code**: ~15,000+
- **Screens**: 25+ interactive screens
- **Embedded Assets**: 13 scripts + 7 templates
- **Binary Size**: ~4 MB (all-in-one)
- **Supported Packages**: 13 pre-configured

---

## ⚡ Performance

- **Fast Startup**: < 100ms
- **Low Memory**: ~20 MB RAM usage
- **Efficient**: Single binary, no runtime dependencies
- **Responsive**: Smooth TUI interactions

---

**Made with ❤️ for Linux system administrators**

For questions, issues, or contributions, please visit the [GitHub repository](https://github.com/iperamuna/ravact).
