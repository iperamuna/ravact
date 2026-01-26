package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// SiteCommandItem represents a site command menu item
type SiteCommandItem struct {
	ID          string
	Name        string
	Description string
	Screen      ScreenType
}

// SiteCommandsModel represents the site commands menu screen
type SiteCommandsModel struct {
	theme  *theme.Theme
	width  int
	height int
	cursor int
	items  []SiteCommandItem
	cwd    string // Current working directory
}

// NewSiteCommandsModel creates a new site commands menu model
func NewSiteCommandsModel() SiteCommandsModel {
	items := []SiteCommandItem{
		{
			ID:          "git",
			Name:        "Git Operations",
			Description: "Manage Git repository, remotes, and connections",
			Screen:      GitManagementScreen,
		},
		{
			ID:          "frankenphp",
			Name:        "FrankenPHP Classic Mode",
			Description: "Set up FrankenPHP sites with systemd + Nginx",
			Screen:      FrankenPHPClassicScreen,
		},
		{
			ID:          "setup_php_symlink",
			Name:        "Setup PHP → FrankenPHP Symlink",
			Description: "Create php → fpcli symlink for CLI commands",
			Screen:      ExecutionScreen,
		},
		{
			ID:          "laravel",
			Name:        "Laravel Permissions",
			Description: "Set proper file permissions for Laravel projects",
			Screen:      LaravelPermissionsScreen,
		},
		{
			ID:          "npm_install",
			Name:        "NPM Install",
			Description: "Run npm install in current directory",
			Screen:      ExecutionScreen,
		},
		{
			ID:          "npm_build",
			Name:        "NPM Build",
			Description: "Run npm run build (production build)",
			Screen:      ExecutionScreen,
		},
		{
			ID:          "composer_install",
			Name:        "Composer Install",
			Description: "Run composer install using system PHP",
			Screen:      ExecutionScreen,
		},
		{
			ID:          "composer_install_fpcli",
			Name:        "Composer Install (FrankenPHP)",
			Description: "Run composer install using fpcli (FrankenPHP)",
			Screen:      ExecutionScreen,
		},
		{
			ID:          "artisan_migrate",
			Name:        "Artisan Migrate",
			Description: "Run php artisan migrate",
			Screen:      ExecutionScreen,
		},
		{
			ID:          "artisan_cache_clear",
			Name:        "Artisan Clear All Caches",
			Description: "Clear config, route, view, and application cache",
			Screen:      ExecutionScreen,
		},
		{
			ID:          "artisan_optimize",
			Name:        "Artisan Optimize",
			Description: "Run php artisan optimize for production",
			Screen:      ExecutionScreen,
		},
	}

	return SiteCommandsModel{
		theme:  theme.DefaultTheme(),
		cursor: 0,
		items:  items,
	}
}

// Init initializes the site commands screen
func (m SiteCommandsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for site commands menu
func (m SiteCommandsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc", "backspace":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: MainMenuScreen}
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		case "enter", " ":
			selectedItem := m.items[m.cursor]
			return m.executeAction(selectedItem)
		}
	}

	return m, nil
}

// executeAction handles the selected menu item
func (m SiteCommandsModel) executeAction(item SiteCommandItem) (SiteCommandsModel, tea.Cmd) {
	switch item.ID {
	case "git":
		return m, func() tea.Msg {
			return NavigateMsg{Screen: GitManagementScreen}
		}

	case "frankenphp":
		return m, func() tea.Msg {
			return NavigateMsg{Screen: FrankenPHPClassicScreen}
		}

	case "setup_php_symlink":
		// Create php → fpcli symlink
		script := `
echo ""
echo "========================================="
echo "Setting up PHP → FrankenPHP Symlink"
echo "========================================="
echo ""

# Check if fpcli exists
if [ ! -f /usr/local/bin/fpcli ]; then
    echo "Error: /usr/local/bin/fpcli not found!"
    echo ""
    echo "Please set up FrankenPHP Classic Mode first to create the fpcli wrapper."
    exit 1
fi

# Backup existing php if it exists and is not a symlink
if [ -f /usr/local/bin/php ] && [ ! -L /usr/local/bin/php ]; then
    echo "Backing up existing /usr/local/bin/php to /usr/local/bin/php.bak"
    mv /usr/local/bin/php /usr/local/bin/php.bak
fi

# Create symlink
ln -sf /usr/local/bin/fpcli /usr/local/bin/php
hash -r 2>/dev/null || true

echo "✓ Created symlink: /usr/local/bin/php -> /usr/local/bin/fpcli"
echo ""
echo "Verification:"
echo "  Location: $(which php)"
echo "  Version:"
php -v
echo ""
echo "✓ 'php' command now uses FrankenPHP!"
echo ""
echo "Note: System PHP (if installed) is still available at /usr/bin/php"
`
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     script,
				Description: "Setting up PHP → FrankenPHP symlink",
			}
		}

	case "laravel":
		return m, func() tea.Msg {
			return NavigateMsg{Screen: LaravelPermissionsScreen}
		}

	case "npm_install":
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: NodeVersionScreen,
				Data:   map[string]interface{}{"commandType": "npm_install"},
			}
		}

	case "npm_build":
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: NodeVersionScreen,
				Data:   map[string]interface{}{"commandType": "npm_build"},
			}
		}

	case "composer_install":
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: PHPVersionScreen,
				Data:   map[string]interface{}{"commandType": "composer_install"},
			}
		}

	case "composer_install_fpcli":
		// Run composer install using fpcli (FrankenPHP)
		script := `
echo ""
echo "========================================="
echo "Composer Install (FrankenPHP)"
echo "========================================="
echo ""

# Check if fpcli exists
if [ ! -f /usr/local/bin/fpcli ]; then
    echo "Error: /usr/local/bin/fpcli not found!"
    echo ""
    echo "Please set up FrankenPHP Classic Mode first to create the fpcli wrapper."
    exit 1
fi

# Check if composer.phar exists, if not check for composer command
COMPOSER_CMD=""
if [ -f /usr/local/bin/composer.phar ]; then
    COMPOSER_CMD="/usr/local/bin/fpcli /usr/local/bin/composer.phar"
    echo "Using: fpcli + composer.phar"
elif [ -f /usr/local/bin/composer ]; then
    # Check if it's already a wrapper
    if head -1 /usr/local/bin/composer 2>/dev/null | grep -q "bash\|sh"; then
        COMPOSER_CMD="/usr/local/bin/composer"
        echo "Using: composer wrapper"
    else
        COMPOSER_CMD="/usr/local/bin/fpcli /usr/local/bin/composer"
        echo "Using: fpcli + composer"
    fi
elif command -v composer &> /dev/null; then
    COMPOSER_PATH=$(which composer)
    COMPOSER_CMD="/usr/local/bin/fpcli $COMPOSER_PATH"
    echo "Using: fpcli + $COMPOSER_PATH"
else
    echo "Error: Composer not found!"
    echo ""
    echo "Please install Composer first via FrankenPHP Classic Mode setup"
    echo "or run: curl -sS https://getcomposer.org/installer | fpcli"
    exit 1
fi

echo ""
echo "Running: $COMPOSER_CMD install"
echo ""

$COMPOSER_CMD install

echo ""
echo "✓ Composer install completed!"
`
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     script,
				Description: "Running composer install with FrankenPHP",
			}
		}

	case "artisan_migrate":
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: PHPVersionScreen,
				Data:   map[string]interface{}{"commandType": "artisan_migrate"},
			}
		}

	case "artisan_cache_clear":
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: PHPVersionScreen,
				Data:   map[string]interface{}{"commandType": "artisan_cache_clear"},
			}
		}

	case "artisan_optimize":
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: PHPVersionScreen,
				Data:   map[string]interface{}{"commandType": "artisan_optimize"},
			}
		}
	}

	return m, nil
}

// View renders the site commands menu
func (m SiteCommandsModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	// Header with host info
	hostInfo := system.GetHostInfo()
	headerText := "Site Commands"
	if hostInfo != "" {
		headerText = fmt.Sprintf("Site Commands  %s", m.theme.DescriptionStyle.Render(hostInfo))
	}
	header := m.theme.Title.Render(headerText)

	// Description
	desc := m.theme.DescriptionStyle.Render("Commands run from current working directory. Navigate to your project folder before running.")

	// Menu items
	var menuItems []string
	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render(m.theme.Symbols.Cursor + " ")
		}

		var renderedItem string
		if i == m.cursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, item.Name))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, item.Name))
		}

		itemDesc := m.theme.DescriptionStyle.Render("  " + item.Description)

		menuItems = append(menuItems, renderedItem)
		menuItems = append(menuItems, itemDesc)
		menuItems = append(menuItems, "")
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, menuItems...)

	// Help
	help := m.theme.Help.Render(m.theme.Symbols.ArrowUp + "/" + m.theme.Symbols.ArrowDown + ": Navigate " + m.theme.Symbols.Bullet + " Enter: Select " + m.theme.Symbols.Bullet + " Esc: Back " + m.theme.Symbols.Bullet + " q: Quit")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		desc,
		"",
		"",
		menu,
		"",
		help,
	)

	// Add border and center
	bordered := m.theme.BorderStyle.Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}
