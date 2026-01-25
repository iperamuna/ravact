# UI Guide

[← Back to Documentation](../README.md)

Ravact uses a modern, consistent UI design throughout the application. This guide covers the menu structure, form system, theming, and terminal compatibility.

## Menu Structure

The main menu is organized into logical categories following industry standards:

### 📦 Package Management
- **Install Software** - Install server packages (Nginx, MySQL, PHP, Redis, etc.)
- **Installed Applications** - View and manage installed services

### ⚙️ Service Configuration
- **Service Settings** - Configure Nginx, MySQL, PostgreSQL, Redis, PHP-FPM, etc.

### 🌐 Site Management
- **Site Commands** - Git, Laravel, Composer, NPM, and deployment tools
- **Developer Toolkit** - Essential commands for Laravel & WordPress maintenance

### 👥 System Administration
- **User Management** - Manage users, groups, and sudo privileges
- **Quick Commands** - System diagnostics, logs, and service controls

### 🔧 Tools
- **File Browser** - Full-featured file manager with preview and operations

## Form System

Ravact uses the [huh](https://github.com/charmbracelet/huh) library for beautiful, interactive forms.

### Form Types

1. **Text Input** - For entering text (usernames, paths, etc.)
2. **Password Input** - Masked input for sensitive data
3. **Select** - Choose from a list of options
4. **Confirm** - Yes/No toggle

### Form Navigation

| Key | Action |
|-----|--------|
| `Tab` | Move to next field |
| `Shift+Tab` | Move to previous field |
| `Enter` | Submit form / Select option |
| `↑`/`↓` | Change option in selects |
| `Space` | Toggle confirms |
| `Esc` | Cancel form |

### Form Validation

- Fields show validation errors in real-time
- Error messages appear below the field in red
- Required fields show `*` indicator when invalid
- Form cannot be submitted until all validations pass

## Theming

Ravact uses a custom color theme designed for readability and accessibility.

### Color Palette

| Color | Usage | Hex (True Color) | ANSI 256 |
|-------|-------|------------------|----------|
| Primary (Orange) | Highlights, focused elements | `#FF6B35` | `208` |
| Secondary (Blue) | Subtitles, labels | `#004E89` | `24` |
| Success (Green) | Success messages, checkmarks | `#2ECC71` | `34` |
| Warning (Yellow) | Warnings, cautions | `#F39C12` | `220` |
| Error (Red) | Errors, failures | `#E74C3C` | `196` |
| Info (Blue) | Information, links | `#3498DB` | `33` |
| Subtle (Gray) | Descriptions, disabled | `#7F8C8D` | `245` |
| Text (White) | Main text | `#FFFFFF` | `15` |
| Highlight (Gold) | Keyboard shortcuts | `#FFD700` | `220` |

### Style Elements

- **Title** - Bold, primary color
- **Subtitle** - Italic, secondary color
- **Menu Item** - Normal text color
- **Selected Item** - Bold, white on primary background
- **Help Text** - Italic, subtle gray
- **Borders** - Rounded corners (Unicode) or normal (ASCII)

## Terminal Compatibility

Ravact automatically detects terminal capabilities and adjusts its display.

### Detection

The app checks for:
- True color support (`COLORTERM=truecolor`)
- 256-color support (`TERM=*256color*`)
- Unicode support
- xterm.js/web terminal environments

### Fallbacks

| Feature | Full Terminal | Basic Terminal |
|---------|--------------|----------------|
| Colors | True color (hex) | ANSI 256 or 16 colors |
| Borders | Rounded (`╭╮╰╯`) | ASCII (`+-\|`) |
| Cursor | `▶` | `>` |
| Checkmark | `✓` | `[x]` |
| Cross | `✗` | `[!]` |
| Bullet | `•` | `*` |
| Arrows | `↑↓←→` | `^v<>` |

### xterm.js Support

Ravact works in web-based terminals (ttyd, wetty, gotty, etc.):
- ANSI 256 colors are used by default
- Unicode symbols work in most modern browsers
- Copy functionality works via clipboard API

## Copy Functionality

Most screens support copying content to the system clipboard.

### How to Copy

1. Navigate to content you want to copy
2. Press `c`
3. See "Copied to clipboard!" confirmation
4. Paste anywhere with `Ctrl+V` or `Cmd+V`

### Screens with Copy Support

- **Execution Output** - Copy command output
- **Text Display** - Copy displayed text
- **Developer Toolkit** - Copy commands
- **File Browser** - Copy file paths or file content
- **MySQL/PostgreSQL/Redis Config** - Copy configuration details

## Accessibility

- **Keyboard-first** - All features accessible via keyboard
- **High contrast** - Colors chosen for readability
- **Clear feedback** - Visual confirmation for all actions
- **Consistent navigation** - Same keys work across all screens

## Tips

1. **Use `q` to quit** - Works on most screens
2. **`Esc` goes back** - Navigate up the menu hierarchy
3. **`c` copies** - When in doubt, try pressing `c`
4. **`?` for help** - In File Browser, shows all shortcuts
5. **Tab through forms** - Don't use arrow keys in forms
