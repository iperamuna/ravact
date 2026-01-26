# Ravact - Project Context & Development Guide

## Project Overview

**Ravact** is a Linux server management TUI (Terminal User Interface) tool built in Go. It provides a modern, keyboard-driven interface for managing web servers, databases, and system services - primarily targeting Ubuntu/Debian systems.

**Current Version:** 0.2.1 (Feature Complete)
**Go Version:** 1.24.0
**Target Platform:** Linux (Ubuntu/Debian primarily, with CentOS/RHEL/Fedora support)

## Architecture

### Directory Structure

```
ravact/
├── cmd/ravact/           # Main entry point (embeds assets at build time)
├── internal/
│   ├── config/           # Configuration management (manager.go)
│   ├── executor/         # Script execution (script_runner.go)
│   ├── models/           # Data models (Service, SetupScript, ConfigTemplate, etc.)
│   ├── setup/            # Setup execution logic
│   ├── system/           # System operations (users, services, detection)
│   └── ui/
│       ├── screens/      # All UI screens (40+ screens)
│       └── theme/        # Theming system with terminal compatibility
├── assets/
│   ├── configs/          # JSON configs (nginx templates, php extensions)
│   └── scripts/          # Shell installation scripts (nginx.sh, php.sh, etc.)
├── scripts/              # Development/deployment helper scripts
├── docs/                 # Comprehensive documentation
└── tests/                # Integration tests
```

### Key Technologies

- **UI Framework:** Charm libraries (bubbletea, bubbles, lipgloss, huh)
- **Pattern:** Model-View-Update (MVU) architecture via bubbletea
- **Build:** Assets embedded at compile time via `//go:embed`

## Code Patterns & Standards

### Screen Implementation Pattern

All screens follow the bubbletea Model pattern:

```go
type SomeScreenModel struct {
    theme      *theme.Theme
    width      int
    height     int
    cursor     int
    // ... screen-specific fields
}

func NewSomeScreenModel() SomeScreenModel { ... }
func (m SomeScreenModel) Init() tea.Cmd { ... }
func (m SomeScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { ... }
func (m SomeScreenModel) View() string { ... }
```

### Styling Rules

- **Always use:** `m.theme.*` for all styling
- **Symbols:** Use `m.theme.Symbols.*` for terminal-safe icons
- **Forms:** Use `m.theme.HuhTheme` (not `huh.ThemeDracula()`)
- **Navigation:** Use `NavigateMsg`, not direct function calls

### Navigation System

```go
// Navigate to a screen
return m, tea.Cmd(func() tea.Msg {
    return NavigateMsg{Screen: TargetScreen}
})

// Go back
return m, tea.Cmd(func() tea.Msg {
    return BackMsg{}
})
```

### Screen Types (defined in navigation.go)

Key screens: `MainMenuScreen`, `SetupMenuScreen`, `ConfigMenuScreen`, `UserManagementScreen`, `FileBrowserScreen`, `DeveloperToolkitScreen`, `QuickCommandsScreen`, etc.

## Features

### Core Functionality

1. **Package Management** - Install Nginx, MySQL, PostgreSQL, Redis, PHP, Node.js, FrankenPHP, Supervisor
2. **Service Configuration** - Configure all installed services via TUI
3. **Site Management** - Nginx site management, SSL certificates, Laravel/WordPress tools
4. **User Management** - System users, groups, sudo privileges
5. **Developer Toolkit** - 34+ essential commands for Laravel/WordPress/PHP/Security
6. **File Browser** - Full vim-style file manager with preview, search, multi-select

### Keyboard Conventions

- `↑`/`↓` or `j`/`k` - Navigate
- `Enter` - Select/Execute
- `Esc` - Go back
- `q` - Quit
- `c` - Copy to clipboard
- `?` - Help (in File Browser)
- `Tab` - Switch categories/fields

## Building & Testing

### Build Commands

```bash
make build              # Current platform
make build-linux        # Linux amd64
make build-linux-arm64  # Linux arm64
make build-darwin-arm64 # macOS Apple Silicon
make build-all          # All platforms
```

### Testing

```bash
make test               # Run all tests
make test-coverage      # With coverage report
make docker-test        # Test in Docker container

# Manual testing (requires Linux or Multipass VM)
sudo ./ravact
```

### Testing on macOS

UI works but setup features require Linux. Use Multipass:

```bash
multipass launch 24.04 --name ravact-test --memory 4G --cpus 2 --disk 20G
GOOS=linux GOARCH=arm64 go build -o ravact-linux ./cmd/ravact
multipass transfer ravact-linux ravact-test:/home/ubuntu/ravact
multipass shell ravact-test
sudo ./ravact
```

## Setup Scripts

Located in `assets/scripts/`. Each script:
- Checks for root privileges
- Detects OS distribution
- Installs and configures the service
- Enables systemd service
- Configures firewall if available

Scripts accept environment variables for customization (e.g., `MYSQL_ROOT_PASSWORD`, `PHP_VERSION`).

## Terminal Compatibility

The theme system auto-detects terminal capabilities:
- True color support → Full hex colors
- 256-color → ANSI 256 fallback
- Basic terminal → ASCII symbols instead of Unicode

Works in: Standard terminals, SSH, xterm.js, web terminals (ttyd, wetty)

## Adding New Features

### New Screen

1. Create file in `internal/ui/screens/`
2. Implement Model interface (Init, Update, View)
3. Add ScreenType constant to `navigation.go`
4. Add navigation case in parent screen
5. Use existing screen as template (e.g., `redis_config_screen.go`)

### New Setup Script

1. Create script in `assets/scripts/`
2. Follow existing script patterns (root check, OS detection, systemd)
3. Add entry in setup menu screen

## Documentation

All documentation is in `docs/`:
- `getting-started/QUICKSTART.md` - Installation and first run
- `features/` - Feature-specific guides
- `development/DEVELOPMENT.md` - Contributing guide
- `testing/TESTING_GUIDE.md` - Testing procedures
- `project/` - Status, changelog, roadmap

## Important Files Reference

| File | Purpose |
|------|---------|
| `internal/models/models.go` | All data structures |
| `internal/system/system.go` | System detection, service status |
| `internal/ui/screens/main_menu.go` | Entry point, menu structure |
| `internal/ui/screens/navigation.go` | Screen types, navigation messages |
| `internal/ui/theme/theme.go` | Colors, styles, symbols |
| `assets/configs/nginx-templates.json` | Nginx site templates |
| `assets/configs/php-extensions.json` | PHP extension definitions |
