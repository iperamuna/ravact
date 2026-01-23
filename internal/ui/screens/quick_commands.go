package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/models"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// QuickCommandsModel represents the quick commands screen
type QuickCommandsModel struct {
	theme    *theme.Theme
	width    int
	height   int
	cursor   int
	commands []models.QuickCommand
}

// NewQuickCommandsModel creates a new quick commands model
func NewQuickCommandsModel() QuickCommandsModel {
	// Define common quick commands
	commands := []models.QuickCommand{
		{
			ID:          "restart-nginx",
			Name:        "Restart Nginx",
			Description: "Restart the Nginx web server",
			Command:     "systemctl",
			Args:        []string{"restart", "nginx"},
			RequireRoot: true,
			Confirm:     true,
		},
		{
			ID:          "reload-nginx",
			Name:        "Reload Nginx Configuration",
			Description: "Reload Nginx configuration without dropping connections",
			Command:     "systemctl",
			Args:        []string{"reload", "nginx"},
			RequireRoot: true,
			Confirm:     false,
		},
		{
			ID:          "test-nginx",
			Name:        "Test Nginx Configuration",
			Description: "Test Nginx configuration for syntax errors",
			Command:     "nginx",
			Args:        []string{"-t"},
			RequireRoot: true,
			Confirm:     false,
		},
		{
			ID:          "view-nginx-status",
			Name:        "View Nginx Status",
			Description: "Show Nginx service status",
			Command:     "systemctl",
			Args:        []string{"status", "nginx"},
			RequireRoot: false,
			Confirm:     false,
		},
		{
			ID:          "view-error-log",
			Name:        "View Nginx Error Log",
			Description: "Display last 50 lines of Nginx error log",
			Command:     "tail",
			Args:        []string{"-n", "50", "/var/log/nginx/error.log"},
			RequireRoot: true,
			Confirm:     false,
		},
		{
			ID:          "view-access-log",
			Name:        "View Nginx Access Log",
			Description: "Display last 50 lines of Nginx access log",
			Command:     "tail",
			Args:        []string{"-n", "50", "/var/log/nginx/access.log"},
			RequireRoot: true,
			Confirm:     false,
		},
		{
			ID:          "disk-usage",
			Name:        "Check Disk Usage",
			Description: "Show disk space usage",
			Command:     "df",
			Args:        []string{"-h"},
			RequireRoot: false,
			Confirm:     false,
		},
		{
			ID:          "memory-usage",
			Name:        "Check Memory Usage",
			Description: "Show memory usage statistics",
			Command:     "free",
			Args:        []string{"-h"},
			RequireRoot: false,
			Confirm:     false,
		},
		{
			ID:          "top-processes",
			Name:        "View Top Processes",
			Description: "Show top CPU-consuming processes",
			Command:     "ps",
			Args:        []string{"aux", "--sort=-pcpu"},
			RequireRoot: false,
			Confirm:     false,
		},
		{
			ID:          "system-info",
			Name:        "System Information",
			Description: "Display system information",
			Command:     "uname",
			Args:        []string{"-a"},
			RequireRoot: false,
			Confirm:     false,
		},
	}

	return QuickCommandsModel{
		theme:    theme.DefaultTheme(),
		cursor:   0,
		commands: commands,
	}
}

// Init initializes the quick commands screen
func (m QuickCommandsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for quick commands
func (m QuickCommandsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.commands)-1 {
				m.cursor++
			}

		case "enter", " ":
			if len(m.commands) > 0 {
				selectedCmd := m.commands[m.cursor]
				return m, func() tea.Msg {
					return ExecutionStartMsg{
						Command:     fmt.Sprintf("%s %v", selectedCmd.Command, selectedCmd.Args),
						Description: selectedCmd.Description,
					}
				}
			}
		}
	}

	return m, nil
}

// View renders the quick commands screen
func (m QuickCommandsModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	header := m.theme.Title.Render("Quick Commands")

	// Group commands by category
	var menuItems []string

	// Add category: Nginx Commands
	categoryHeader := m.theme.Label.Render("▼ Nginx Commands")
	menuItems = append(menuItems, categoryHeader, "")

	for i := 0; i < 6; i++ {
		if i >= len(m.commands) {
			break
		}
		menuItems = append(menuItems, m.renderCommand(i, m.commands[i]))
	}

	menuItems = append(menuItems, "")

	// Add category: System Commands
	categoryHeader = m.theme.Label.Render("▼ System Commands")
	menuItems = append(menuItems, categoryHeader, "")

	for i := 6; i < len(m.commands); i++ {
		menuItems = append(menuItems, m.renderCommand(i, m.commands[i]))
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, menuItems...)

	// Help
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Execute • Esc: Back • q: Quit")

	// Warning about root
	var warning string
	hasRootCommands := false
	for _, cmd := range m.commands {
		if cmd.RequireRoot {
			hasRootCommands = true
			break
		}
	}
	if hasRootCommands {
		warning = m.theme.WarningStyle.Render("Note: Some commands require root privileges")
	}

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
	)

	if warning != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content, warning, "")
	}

	content = lipgloss.JoinVertical(
		lipgloss.Left,
		content,
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

// renderCommand renders a single command item
func (m QuickCommandsModel) renderCommand(index int, cmd models.QuickCommand) string {
	cursor := "  "
	if index == m.cursor {
		cursor = m.theme.KeyStyle.Render("▶ ")
	}

	// Add indicators
	indicators := ""
	if cmd.RequireRoot {
		indicators += m.theme.ErrorStyle.Render("[root]") + " "
	}
	if cmd.Confirm {
		indicators += m.theme.WarningStyle.Render("[confirm]") + " "
	}

	title := cmd.Name
	if indicators != "" {
		title += " " + indicators
	}

	desc := m.theme.DescriptionStyle.Render(cmd.Description)

	var renderedItem string
	if index == m.cursor {
		renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, title))
	} else {
		renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, title))
	}

	return renderedItem + "\n  " + desc + "\n"
}
