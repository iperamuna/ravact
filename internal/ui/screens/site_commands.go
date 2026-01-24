package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
			Description: "Run composer install in current directory",
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
	header := m.theme.Title.Render("Site Commands")

	// Description
	desc := m.theme.DescriptionStyle.Render("Commands run from current working directory. Navigate to your project folder before running.")

	// Menu items
	var menuItems []string
	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
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
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Select • Esc: Back • q: Quit")

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
