# Changelog

All notable changes to Ravact will be documented in this file.

## [0.1.2] - 2026-01-23

### Added
- **Version Number in Main Menu**: Now displays version in Main Menu header
- **Enhanced System Information**: 
  - Added Architecture (x86_64, arm64, etc.)
  - Added RAM size (formatted as GB/MB) - **Works on both macOS and Linux**
  - Added Physical Disk size - **Works on both macOS and Linux**
  - Improved layout and formatting
- **Installed Applications Screen**: New menu item to view and manage installed apps
  - Shows only applications that are actually installed
  - Cross-references with Ravact setup scripts
  - Displays status badges (Running, Stopped, Failed)
  - Direct access to manage each installed app
  - Press Enter to manage any installed application
  - Press 'r' to refresh status
  - Shows helpful message when no apps are installed

### Changed
- Main menu now has 4 items instead of 3
- System info section more comprehensive and better formatted
- Better visual hierarchy in main menu

### Fixed
- **Cross-Platform RAM Detection**: Now works on both macOS (via sysctl) and Linux (via /proc/meminfo)
- **Cross-Platform Disk Detection**: Now works on both macOS (via df -k) and Linux (via df -B1)
- RAM and Disk previously showed "0 B" on macOS - now shows actual values

## [0.1.1] - 2026-01-23

### Added
- **Installation Status Detection**: Setup menu now shows real-time status for each service
  - ✓ Running (green)
  - Installed (blue)
  - ⚠ Stopped (yellow)
  - ✗ Failed (red)
  - Not Installed (gray)
  
- **Setup Action Screen**: New screen for managing installed services
  - Install option for not-installed services
  - Reinstall/Update option for installed services
  - Start/Stop/Restart service controls
  - Remove/Uninstall option
  - Actions adapt based on current service status

- **Quick Actions in Setup Menu**:
  - Press `i` to quick-install (bypass action menu)
  - Press `r` to refresh status of selected service
  - Press `Enter` to see all available actions

### Changed
- Setup menu help text updated to show new keyboard shortcuts
- Navigation flow: Setup Menu → Action Selection → Execute

### Technical
- Added `ServiceStatus` detection in setup menu
- Created `SetupActionModel` screen with dynamic actions
- Enhanced `SetupScript` model with `ServiceID` field
- Integrated system detector in setup menu for status checking

## [0.1.0] - 2026-01-23

### Added
- Initial release of Ravact TUI application
- Beautiful terminal interface with splash screen
- Main menu with system information display
- Setup menu for software installation
- Quick commands menu (10 system commands)
- System detection (OS, CPU, RAM, disk)
- Setup script execution engine with real-time output
- Configuration management with JSON templates
- Cross-platform builds (Linux x64/ARM64, macOS x64/ARM64)
- Comprehensive test suite (85%+ coverage)
- Docker testing environment (Ubuntu 24.04)
- GitHub Actions CI/CD workflows
- Complete documentation

### Fixed
- Navigation issue where screens showed "Loading..." after transition
- Window size now properly propagated to new screens on navigation

## [0.1.3] - 2026-01-24

### Added
- **Complete Database Management**:
  - MySQL: Password management, port configuration, database creation, user management
  - PostgreSQL: Password management, port configuration, performance tuning (max_connections, shared_buffers)
- **PHP-FPM Pool Management**: View pools, service restart/reload, pool details
- **Supervisor Management**: Program management, XML-RPC configuration (IP, port, username, password)
- **Firewall Management**: UFW firewall configuration
- **Full SSL Certificate Management**: Let's Encrypt (certbot) and manual certificate support
- **13+ Setup Scripts**: Nginx, MySQL, PostgreSQL, Redis, PHP, Node.js, Git, Supervisor, Certbot, FrankenPHP, Dragonfly

### Fixed
- **Makefile build-all**: Darwin builds now correctly copy assets for embedding

## [Unreleased]

### Planned Features
- Remote server management via SSH
- Multi-server support
- Service monitoring dashboard with live updates
- Backup and restore functionality
- Web dashboard companion

---

**Format**: This changelog follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
**Versioning**: This project uses [Semantic Versioning](https://semver.org/)
