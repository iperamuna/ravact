package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// ConfigMenuItem represents a configuration menu item
type ConfigMenuItem struct {
	ID          string
	Name        string
	Description string
	Available   bool
	Screen      ScreenType
}

// ConfigMenuModel represents the configuration menu screen
type ConfigMenuModel struct {
	theme  *theme.Theme
	width  int
	height int
	cursor int
	items  []ConfigMenuItem
}

// NewConfigMenuModel creates a new configuration menu model
func NewConfigMenuModel() ConfigMenuModel {
	items := []ConfigMenuItem{
		{
			ID:          "nginx",
			Name:        "Nginx Web Server",
			Description: "Manage sites, virtual hosts, and SSL certificates",
			Available:   true, // Always show
			Screen:      NginxConfigScreen,
		},
		{
			ID:          "redis",
			Name:        "Redis Cache",
			Description: "Configure Redis server settings and authentication",
			Available:   true,
			Screen:      RedisConfigScreen,
		},
		{
			ID:          "mysql",
			Name:        "MySQL Database",
			Description: "Manage MySQL databases, passwords, and port configuration",
			Available:   true,
			Screen:      MySQLManagementScreen,
		},
		{
			ID:          "postgresql",
			Name:        "PostgreSQL Database",
			Description: "Manage PostgreSQL databases, passwords, and performance tuning",
			Available:   true,
			Screen:      PostgreSQLManagementScreen,
		},
		{
			ID:          "php",
			Name:        "PHP-FPM Pools",
			Description: "Manage PHP-FPM pools and worker process configuration",
			Available:   true,
			Screen:      PHPFPMManagementScreen,
		},
		{
			ID:          "supervisor",
			Name:        "Supervisor",
			Description: "Manage supervisor programs and XML-RPC configuration",
			Available:   true,
			Screen:      SupervisorManagementScreen,
		},
	}

	return ConfigMenuModel{
		theme:  theme.DefaultTheme(),
		items:  items,
		cursor: 0,
	}
}

// Init initializes the configuration menu
func (m ConfigMenuModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the configuration menu
func (m ConfigMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if selectedItem.Available {
				return m, func() tea.Msg {
					return NavigateMsg{Screen: selectedItem.Screen}
				}
			}
		}
	}

	return m, nil
}

// View renders the configuration menu
func (m ConfigMenuModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	header := m.theme.Title.Render("Configurations")

	// Description
	desc := m.theme.DescriptionStyle.Render("Manage service configurations for installed applications")

	// Menu items
	var menuItems []string
	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		// Build item display
		itemName := item.Name
		if !item.Available {
			itemName += " " + m.theme.DescriptionStyle.Render("(Coming Soon)")
		}

		var renderedItem string
		if i == m.cursor {
			if item.Available {
				renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, itemName))
			} else {
				renderedItem = m.theme.DescriptionStyle.Render(fmt.Sprintf("%s%s", cursor, itemName))
			}
		} else {
			if item.Available {
				renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, itemName))
			} else {
				renderedItem = m.theme.DescriptionStyle.Render(fmt.Sprintf("%s%s", cursor, itemName))
			}
		}

		itemDesc := m.theme.DescriptionStyle.Render(fmt.Sprintf("  %s", item.Description))

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
