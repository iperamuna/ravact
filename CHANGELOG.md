# Changelog

[← Back to Documentation](README.md)

## [0.4.1] - 2026-01-31

### Added
- **FrankenPHP Caddy Metrics**: Added comprehensive support for Caddy metrics in FrankenPHP services
  - **Metrics Toggle**: Enable/Disable metrics collection via the service actions menu
  - **Port Configuration**: Configurable metrics port (default 2222), accessible via localhost only
  - **Auto-Update**: Automatically updates Caddyfile and restarts service when enabling/disabling
- **Laravel Permissions Enhancements**:
  - **Storage Link Detection**: Automatically detects missing `public/storage` link and prompts to create it during permission actions
  - **Group Sync**: "Full Permission Reset" now ensures the system user is part of the web server group (using `usermod`)
  - **Improved Reset Logic**: more robust permissions reset command with explicit `sudo` usage for all commands

### Changed
- **Nginx Config Handling**: Stopped automatic generation of Nginx config files during service creation/update (use "View Nginx Config" to generate on demand) for better manual control
- **Caddyfile Generation**: Improved default PHP settings in Caddyfile including customizable upload and post max sizes

---

## [0.4.0] - 2026-01-31

### Added
- **Laravel Queue Management** - Dedicated interface for managing systemd queue workers
  - **Service Templates** - Supports multiple worker instances (e.g., `queue@1`, `queue@2`) via systemd templates
  - **Creation Wizard** - Form-based setup for Service Label, User, Executor, Queue Name, and Process Count
  - **Bulk Actions** - Start, Stop, Restart applies to all worker instances automatically
  - **Log Viewing** - Live journalctl tail for all worker instances
  
- **Laravel Scheduler Configuration**
  - **Interactive Setup** - New form to configure Scheduler User and Executor path
  - **Smart Detection** - Detects existing Cron entries and allows editing
  - **Duplicate Prevention** - Automatically cleans up old cron entries for the project before adding new ones

- **Laravel App Menu**
  - Added "Setup Queue Services" to the menu
  - Enhanced "Setup Laravel Scheduler" with the new interactive form

### Changed
- **FrankenPHP Service Management**
  - Services now utilize systemd templates for better scalability

### Fixed
- **Laravel Permissions** - Improved validation for user selection and path handling

---

## [0.3.5] - 2026-01-31

### Added
- **FrankenPHP Service Management Enhancements**
  - **Nginx Config Generation**: New "View Nginx Config" action generates specific upstream configuration for both Socket and Port modes
  - **Auto-Detection**: Nginx config view automatically detects connection type and pre-fills socket paths or ports
  - **Editor Integration**: New "Edit Configuration (Editor)" action with file selection menu (Caddyfile, Service, Nginx)
  - **Service Deletion**: Full cleanup workflow to delete services along with their config and data directories
  - **Copy Support**: Press `c` to copy generated Nginx configs to clipboard

- **Documentation**
  - Updated FrankenPHP Guide with comprehensive "Managing Services" section
  - Documented new TUI workspaces and Nginx integration workflow

### Changed
- **FrankenPHP Directory Structure**:
  - Improved standard layout: `/var/lib/caddy/{sitekey}/` now contains `config`, `data`, and `tls` subdirectories
  - Better permission handling: Recursive ownership fixes during setup and service restarts to ensure Caddy reads configs correctly

### Fixed
- **Nginx Config Template**: Fixed syntax error in generated `upstream` blocks
- **TUI Interactions**: 
  - Fixed "View Nginx Config" form getting stuck (input blocking issue)
  - Fixed compiler errors in service screen logic

### Test
- Added unit tests for systemd service file parsing logic

---

## [0.3.1] - 2026-01-26

### Added
- **Git System User Management** - Track and manage git operations per repository
  - `git config meta.systemuser` automatically set after cloning
  - Git pull, fetch, status now use configured system user automatically
  - New "Set System User" option in Git Operations menu
  - Prevents duplicate repo setups for same user
  - System user displayed in Git Operations info panel

- **Laravel App Menu** - Reorganized Laravel tools into dedicated submenu
  - Moved from Site Commands: Laravel Permissions, Artisan commands
  - New: Create .env from .env.example with environment selection (local/staging/production)
  - New: Option to auto-generate APP_KEY after .env creation
  - New: Artisan Key Generate command
  - New: Setup Laravel Scheduler (www-data crontab)

- **Passwordless User Creation** - Industry-standard SSH key-only authentication
  - Users created without password (SSH key-only auth)
  - Passwordless su - switch users without password prompt
  - Passwordless sudo (NOPASSWD) - creates `/etc/sudoers.d/{username}` with NOPASSWD:ALL
  - New form option: "Passwordless Access (NOPASSWD)"
  - Validates sudoers file with `visudo -c` before applying

- **FrankenPHP Enhancements**
  - Socket/Port selection for service creation (Unix Socket recommended, TCP Port option)
  - Fixed web directory path - now relative to site root (e.g., "public" becomes /var/www/site/public)
  - Fixed service edit form not completing
  - Fixed service edit crash with bounds checking

- **NPM Build Improvements**
  - Now runs `npm install && npm run build` (both commands)
  - Runs as configured system user via `su - {user}`

- **Developer Toolkit Integration**
  - System user passed to toolkit commands for proper execution context

### Changed
- **Menu Reorganization**
  - Removed FrankenPHP from Setup Install menu (available via Site Commands)
  - Removed Node.js from Setup Install menu (managed via npm commands)
  - Removed FrankenPHP and Node.js from Installed Applications
  - Laravel permissions screen renamed to "Laravel App"

- **User Creation Form**
  - Removed password field (passwordless by default)
  - Added "Passwordless Access (NOPASSWD)" toggle (default: Yes)
  - Grant Sudo now defaults to Yes
  - Added form field keys for reliable value capture

### Fixed
- **User Creation Form** - Username field now properly captures input
- **FrankenPHP Service Edit** - Form completion now triggers save correctly
- **FrankenPHP Service Edit Crash** - Added nil checks and bounds validation

### Technical
- New UserManager methods: `CreateUserPasswordless`, `GrantSudoNoPassword`, `EnablePasswordlessSu`, `RevokeSudoNoPassword`
- Git operations now check `meta.systemuser` config before prompting for user
- Laravel permissions use system user from git config for chown commands

---

## [0.2.2] - 2026-01-26

### Added
- **Scripts Documentation** - Comprehensive `scripts/README.md` with detailed documentation for all 16 utility scripts
  - Quick reference table for all scripts
  - Detailed usage instructions and examples
  - Script categories (Development, Testing, Release, Deployment)
  - Tips for fastest development workflow

### Changed
- **Go Version Updated** - Updated Go version from 1.21.6 to 1.24.0 in VM setup scripts:
  - `setup-vm-only.sh`
  - `setup-multipass.sh`
  - `setup-mac-vm.sh`
- **Script Headers** - Added consistent headers with usage instructions to:
  - `test.sh`
  - `docker-test.sh`
  - `install.sh`
  - `release.sh`

### Documentation
- Added scripts/README.md with complete documentation for all utility scripts

---

## [0.2.1] - 2026-01-25

### Fixed
- **Form Auto-Focus** - Forms using huh library now automatically focus on first field when navigating to form screens (previously required pressing Enter first)
- **Email Field Validation** - Email field in Add Nginx Site form now correctly shows "Only required if using Let's Encrypt SSL" instead of always being required
- **Execution Output Auto-Scroll** - Script execution output now auto-scrolls to bottom while running, so users can see live output

### Changed
- Moved TODO.md to docs/project/ for better documentation organization

---

## [0.2.0] - 2026-01-25

### Added
- **Developer Toolkit** - 34+ essential commands for Laravel, WordPress, PHP, and Security maintenance
  - Laravel: Tail logs, fix permissions, generate APP_KEY, check queue workers, etc.
  - WordPress: Fix permissions, clear cache, generate salts, find malware patterns
  - PHP: Check version, list modules, find php.ini, check OPcache
  - Security: Scan for malware, find world-writable files, check open ports
- **File Browser** - Full-featured terminal file manager
  - Directory navigation with vim-style keys
  - File preview with line numbers
  - File operations: copy, cut, paste, delete, rename, create
  - Search and filter with live results
  - Multi-selection for batch operations
  - History navigation (back/forward)
  - Hidden files toggle and sorting options
  - Help screen (press `?`)
- **Modern Forms** - Integrated [huh](https://github.com/charmbracelet/huh) library
  - Beautiful, interactive form components
  - Custom theme matching Ravact color scheme
  - Real-time validation with error messages
  - Password masking for sensitive inputs
- **Copy to Clipboard** - Press `c` on most screens to copy content
  - Works in execution output, text displays, config screens
  - Visual feedback when content is copied
- **Categorized Menu** - Reorganized main menu structure
  - Package Management (Install Software, Installed Applications)
  - Service Configuration (Service Settings)
  - Site Management (Site Commands, Developer Toolkit)
  - System Administration (User Management, Quick Commands)
  - Tools (File Browser)
- **xterm.js Compatibility** - Works in web-based terminals
  - Terminal capability detection
  - ANSI 256 color fallback
  - ASCII symbol fallback for basic terminals
- **Standardized UI** - Consistent styling across all screens
  - Theme-aware symbols (Unicode/ASCII)
  - Consistent help text format
  - Uniform spacing and borders

### Changed
- Updated all form screens to use huh library:
  - Add User
  - Add Site
  - MySQL Password/Port
  - PostgreSQL Password/Port
  - Redis Password/Port
  - Supervisor Add Program
- Main menu now shows organized categories instead of flat list
- Help text uses theme symbols for better terminal compatibility
- Cursor indicators use theme-aware symbols

### Technical
- Added `github.com/charmbracelet/huh` dependency
- Added `internal/ui/theme/compat.go` for terminal detection
- Added custom huh theme in `theme.go`
- New screen files:
  - `developer_toolkit.go`
  - `file_browser.go`

---

## Previous Versions

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
