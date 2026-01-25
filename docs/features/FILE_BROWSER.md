# File Browser

[← Back to Documentation](../README.md)

Ravact includes a full-featured terminal file browser for navigating, previewing, and managing files without leaving the application.

## Overview

Access via: **Main Menu → Tools → File Browser**

## Features

- **Directory Navigation** - Browse the entire filesystem
- **File Preview** - View text files with line numbers
- **File Operations** - Copy, cut, paste, delete, rename
- **Search & Filter** - Find files with live filtering
- **Multi-Selection** - Select multiple files for batch operations
- **History Navigation** - Go back/forward through visited directories
- **Hidden Files Toggle** - Show/hide dotfiles
- **Sorting Options** - Sort by name, size, or date

## Keyboard Shortcuts

Press `?` while in the File Browser to see the complete help screen.

### Navigation

| Key | Action |
|-----|--------|
| `↑`/`k` | Move cursor up |
| `↓`/`j` | Move cursor down |
| `Enter`/`l`/`→` | Open directory or preview file |
| `Backspace`/`h`/`←` | Go to parent directory |
| `PgUp`/`Ctrl+U` | Page up |
| `PgDown`/`Ctrl+D` | Page down |
| `Home`/`g` | Go to first item |
| `End`/`G` | Go to last item |
| `~` | Go to home directory |
| `` ` `` | Go to root directory (/) |
| `-` | Go back in history |
| `=`/`+` | Go forward in history |

### Selection

| Key | Action |
|-----|--------|
| `Space` | Toggle selection on current item |
| `a` | Select all items |
| `A` | Clear all selections |

### File Operations

| Key | Action |
|-----|--------|
| `y` | Copy selected items to clipboard |
| `x` | Cut selected items to clipboard |
| `p` | Paste from clipboard |
| `c` | Copy file path to system clipboard |
| `n` | Create new file |
| `N` | Create new directory |
| `r` | Rename current item |
| `d` | Delete selected items (with confirmation) |
| `o` | Open with system default application |

### Search & View

| Key | Action |
|-----|--------|
| `/` | Start search (live filtering) |
| `.` | Toggle hidden files |
| `s` | Cycle sort (Name → Size → Date) |
| `S` | Reverse sort order |
| `R`/`Ctrl+R` | Refresh directory |

### Preview Mode

When viewing a file:

| Key | Action |
|-----|--------|
| `↑`/`k` | Scroll up |
| `↓`/`j` | Scroll down |
| `PgUp`/`PgDn` | Scroll page up/down |
| `Home`/`g` | Go to beginning |
| `End`/`G` | Go to end |
| `c` | Copy file content to clipboard |
| `o` | Open with external editor |
| `Esc`/`q` | Close preview |

### General

| Key | Action |
|-----|--------|
| `?` | Show/hide help screen |
| `Esc` | Go back / Cancel operation |
| `q` | Quit to main menu |
| `Ctrl+C` | Quit application |

## Display Information

The file browser displays:

- **File/Directory icon** - Visual indicator of file type
- **Name** - File or directory name
- **Size** - File size (or `<DIR>` for directories)
- **Modified Time** - Last modification time
- **Permissions** - Unix permission string (e.g., `-rw-r--r--`)

### Status Bar

Shows:
- Total number of items
- Total size of visible files
- Number of selected items
- Current sort mode
- Hidden files status

### Clipboard Indicator

When items are in clipboard, shows:
- Number of items
- Operation type (copied/cut)

## File Type Icons

The browser shows file type indicators:

| Icon | File Type |
|------|-----------|
| `Go` | Go source files |
| `JS` | JavaScript/TypeScript |
| `Py` | Python |
| `PHP` | PHP |
| `HTML` | HTML files |
| `CSS` | CSS/SCSS/SASS |
| `JSON` | JSON files |
| `YAML` | YAML files |
| `MD` | Markdown |
| `SH` | Shell scripts |
| `SQL` | SQL files |
| `IMG` | Image files |
| `PDF` | PDF documents |
| `ZIP` | Archives |

## Tips & Best Practices

1. **Use vim-style navigation**: `h`/`j`/`k`/`l` keys work just like vim

2. **Quick path copy**: Press `c` to copy the current file's full path, useful for pasting in terminals

3. **Batch operations**: Select multiple files with `Space`, then use `y`/`x`/`d` for batch copy/cut/delete

4. **Live search**: Press `/` and start typing - results filter as you type

5. **Preview before edit**: Press `Enter` on a file to preview it, then `o` to open in external editor if needed

6. **Hidden files**: Many config files are hidden. Press `.` to toggle visibility

## Limitations

- **Binary files**: Cannot preview binary files (shows "[Binary file]")
- **Large files**: Files over 1MB are not previewed (shows "[File too large to preview]")
- **System files**: Some operations require root privileges
- **Symlinks**: Displayed but operations follow the link
