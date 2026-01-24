package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/models"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// SetupAction represents an action that can be performed on a service
type SetupAction struct {
	ID          string
	Name        string
	Description string
	Command     string
}

// SetupActionModel represents the action selection screen
type SetupActionModel struct {
	theme   *theme.Theme
	width   int
	height  int
	cursor  int
	script  models.SetupScript
	status  models.ServiceStatus
	actions []SetupAction
}

// NewSetupActionModel creates a new setup action model
func NewSetupActionModel(script models.SetupScript, status models.ServiceStatus) SetupActionModel {
	// Determine available actions based on status
	actions := []SetupAction{}

	switch status {
	case models.StatusNotInstalled:
		// For PHP, navigate to PHP install screen
		if script.ID == "php" {
			actions = []SetupAction{
				{
					ID:          "install",
					Name:        "Manage PHP Versions",
					Description: "Install or remove PHP versions and extensions",
					Command:     "__php_install__",
				},
			}
		} else if script.ID == "dragonfly" {
			// For Dragonfly, use special install action that navigates to options screen
			actions = []SetupAction{
				{
					ID:          "install",
					Name:        "Install",
					Description: fmt.Sprintf("Install %s (choose installation method)", script.Name),
					Command:     "__dragonfly_install__", // Special marker for navigation
				},
			}
		} else {
			actions = []SetupAction{
				{
					ID:          "install",
					Name:        "Install",
					Description: fmt.Sprintf("Install %s", script.Name),
					Command:     fmt.Sprintf("assets/scripts/%s", script.ScriptPath),
				},
			}
		}

	case models.StatusInstalled, models.StatusStopped:
		// For PHP, always navigate to PHP install screen for management
		if script.ID == "php" {
			actions = []SetupAction{
				{
					ID:          "manage",
					Name:        "Manage PHP Versions",
					Description: "Install, remove PHP versions and extensions",
					Command:     "__php_install__",
				},
			}
		} else if script.ID == "git" || script.ID == "certbot" || script.ID == "node" {
			// Tools that don't run as services (no start/stop/restart)
			// Only show reinstall and remove for non-service tools
			actions = []SetupAction{
				{
					ID:          "reinstall",
					Name:        "Reinstall / Update",
					Description: "Reinstall or update to the latest version",
					Command:     fmt.Sprintf("assets/scripts/%s", script.ScriptPath),
				},
				{
					ID:          "remove",
					Name:        "Remove",
					Description: "Uninstall the package",
					Command:     fmt.Sprintf("apt-get remove -y %s || yum remove -y %s", script.ServiceID, script.ServiceID),
				},
			}
		} else {
			actions = []SetupAction{
				{
					ID:          "reinstall",
					Name:        "Reinstall / Update",
					Description: "Reinstall or update to the latest version",
					Command:     fmt.Sprintf("assets/scripts/%s", script.ScriptPath),
				},
				{
					ID:          "start",
					Name:        "Start Service",
					Description: "Start the service",
					Command:     fmt.Sprintf("systemctl start %s", script.ServiceID),
				},
				{
					ID:          "remove",
					Name:        "Remove",
					Description: "Uninstall and remove the service",
					Command:     fmt.Sprintf("apt-get remove -y %s || yum remove -y %s", script.ServiceID, script.ServiceID),
				},
			}
		}

	case models.StatusRunning:
		// For PHP, always navigate to PHP install screen for management
		if script.ID == "php" {
			actions = []SetupAction{
				{
					ID:          "manage",
					Name:        "Manage PHP Versions",
					Description: "Install, remove PHP versions and extensions",
					Command:     "__php_install__",
				},
			}
		} else if script.ID == "git" || script.ID == "certbot" || script.ID == "node" {
			// Tools that don't run as services (no start/stop/restart)
			// Only show reinstall and remove for non-service tools
			actions = []SetupAction{
				{
					ID:          "reinstall",
					Name:        "Reinstall / Update",
					Description: "Reinstall or update to the latest version",
					Command:     fmt.Sprintf("assets/scripts/%s", script.ScriptPath),
				},
				{
					ID:          "remove",
					Name:        "Remove",
					Description: "Uninstall the package",
					Command:     fmt.Sprintf("apt-get remove -y %s || yum remove -y %s", script.ServiceID, script.ServiceID),
				},
			}
		} else {
			actions = []SetupAction{
				{
					ID:          "restart",
					Name:        "Restart Service",
					Description: "Restart the service",
					Command:     fmt.Sprintf("systemctl restart %s", script.ServiceID),
				},
				{
					ID:          "stop",
					Name:        "Stop Service",
					Description: "Stop the service",
					Command:     fmt.Sprintf("systemctl stop %s", script.ServiceID),
				},
				{
					ID:          "reinstall",
					Name:        "Reinstall / Update",
					Description: "Reinstall or update to the latest version",
					Command:     fmt.Sprintf("assets/scripts/%s", script.ScriptPath),
				},
				{
					ID:          "remove",
					Name:        "Remove",
					Description: "Uninstall and remove the service (will stop it first)",
					Command:     fmt.Sprintf("systemctl stop %s && apt-get remove -y %s || yum remove -y %s", script.ServiceID, script.ServiceID, script.ServiceID),
				},
			}
		}

	case models.StatusFailed:
		actions = []SetupAction{
			{
				ID:          "restart",
				Name:        "Restart Service",
				Description: "Attempt to restart the failed service",
				Command:     fmt.Sprintf("systemctl restart %s", script.ServiceID),
			},
			{
				ID:          "reinstall",
				Name:        "Reinstall",
				Description: "Reinstall to fix issues",
				Command:     fmt.Sprintf("assets/scripts/%s", script.ScriptPath),
			},
			{
				ID:          "remove",
				Name:        "Remove",
				Description: "Uninstall and remove the service",
				Command:     fmt.Sprintf("systemctl stop %s; apt-get remove -y %s || yum remove -y %s", script.ServiceID, script.ServiceID, script.ServiceID),
			},
		}

	default:
		// Unknown status - show install option
		actions = []SetupAction{
			{
				ID:          "install",
				Name:        "Install",
				Description: fmt.Sprintf("Install %s", script.Name),
				Command:     fmt.Sprintf("assets/scripts/%s", script.ScriptPath),
			},
		}
	}

	return SetupActionModel{
		theme:   theme.DefaultTheme(),
		cursor:  0,
		script:  script,
		status:  status,
		actions: actions,
	}
}

// Init initializes the setup action screen
func (m SetupActionModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the setup action screen
func (m SetupActionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				return NavigateMsg{Screen: SetupMenuScreen}
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.actions)-1 {
				m.cursor++
			}

		case "enter", " ":
			if len(m.actions) > 0 {
				selectedAction := m.actions[m.cursor]
				
				// Handle special navigation for PHP install
				if selectedAction.Command == "__php_install__" {
					return m, func() tea.Msg {
						return NavigateMsg{Screen: PHPInstallScreen}
					}
				}
				
				// Handle special navigation for Dragonfly install
				if selectedAction.Command == "__dragonfly_install__" {
					return m, func() tea.Msg {
						return NavigateMsg{Screen: DragonflyInstallScreen}
					}
				}
				
				return m, func() tea.Msg {
					return ExecutionStartMsg{
						Command:     selectedAction.Command,
						Description: selectedAction.Description,
					}
				}
			}
		}
	}

	return m, nil
}

// View renders the setup action screen
func (m SetupActionModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	header := m.theme.Title.Render(fmt.Sprintf("%s - Actions", m.script.Name))

	// Status info
	statusText := ""
	statusColor := m.theme.InfoStyle
	switch m.status {
	case models.StatusNotInstalled:
		statusText = "Not Installed"
		statusColor = m.theme.DescriptionStyle
	case models.StatusInstalled:
		statusText = "Installed"
		statusColor = m.theme.InfoStyle
	case models.StatusRunning:
		statusText = "✓ Running"
		statusColor = m.theme.SuccessStyle
	case models.StatusStopped:
		statusText = "⚠ Stopped"
		statusColor = m.theme.WarningStyle
	case models.StatusFailed:
		statusText = "✗ Failed"
		statusColor = m.theme.ErrorStyle
	default:
		statusText = "Unknown"
		statusColor = m.theme.DescriptionStyle
	}
	statusInfo := statusColor.Render(fmt.Sprintf("Current Status: %s", statusText))

	// Description
	desc := m.theme.DescriptionStyle.Render(m.script.Description)

	// Action items
	var actionItems []string
	for i, action := range m.actions {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		title := action.Name
		actionDesc := m.theme.DescriptionStyle.Render(action.Description)

		var renderedItem string
		if i == m.cursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, title))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, title))
		}

		actionItems = append(actionItems, renderedItem)
		actionItems = append(actionItems, "  "+actionDesc)
		actionItems = append(actionItems, "")
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, actionItems...)

	// Help
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Execute • Esc: Back • q: Quit")

	// Warning
	warning := m.theme.WarningStyle.Render("⚠ Actions may require root privileges")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		statusInfo,
		"",
		desc,
		"",
		"",
		menu,
		"",
		"",
		warning,
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
