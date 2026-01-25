package screens

import (
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// MySQLManagementModel represents the MySQL management screen
type MySQLManagementModel struct {
	theme       *theme.Theme
	width       int
	height      int
	manager     *system.MySQLManager
	config      *system.MySQLConfig
	cursor      int
	actions     []string
	err         error
	success     string
	copied      bool
	copiedTimer int
}

// NewMySQLManagementModel creates a new MySQL management model
func NewMySQLManagementModel() MySQLManagementModel {
	manager := system.NewMySQLManager()
	config, _ := manager.GetConfig()
	
	actions := []string{
		"View Current Configuration",
		"Change Root Password",
		"Change Port",
		"Restart MySQL Service",
		"View Service Status",
		"List Databases",
		"← Back to Configurations",
	}
	
	return MySQLManagementModel{
		theme:   theme.DefaultTheme(),
		manager: manager,
		config:  config,
		cursor:  0,
		actions: actions,
	}
}

// Init initializes the MySQL management screen
func (m MySQLManagementModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m MySQLManagementModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				return NavigateMsg{Screen: ConfigMenuScreen}
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
			return m.executeAction()

		case "c":
			// Copy configuration to clipboard
			if m.config != nil {
				content := fmt.Sprintf("MySQL Configuration\nPort: %d\nBind Address: %s\nConfig Path: %s\nData Dir: %s",
					m.config.Port, m.config.BindAddress, m.config.ConfigPath, m.config.DataDir)
				clipboard.WriteAll(content)
				m.copied = true
				m.copiedTimer = 3
				return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
					return CopyTimerTickMsg{}
				})
			}
		}

	case CopyTimerTickMsg:
		if m.copiedTimer > 0 {
			m.copiedTimer--
			if m.copiedTimer == 0 {
				m.copied = false
			} else {
				return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
					return CopyTimerTickMsg{}
				})
			}
		}
	}

	return m, nil
}

// executeAction executes the selected action
func (m MySQLManagementModel) executeAction() (MySQLManagementModel, tea.Cmd) {
	m.err = nil
	m.success = ""
	
	actionName := m.actions[m.cursor]

	switch actionName {
	case "View Current Configuration":
		config, err := m.manager.GetConfig()
		if err != nil {
			m.err = err
		} else {
			m.config = config
			m.success = "✓ Configuration refreshed"
		}

	case "Change Root Password":
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: MySQLPasswordScreen,
				Data: map[string]interface{}{
					"manager": m.manager,
				},
			}
		}

	case "Change Port":
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: MySQLPortScreen,
				Data: map[string]interface{}{
					"manager": m.manager,
					"config":  m.config,
				},
			}
		}

	case "Restart MySQL Service":
		err := m.manager.RestartService()
		if err != nil {
			m.err = err
		} else {
			m.success = "✓ MySQL service restarted successfully"
		}

	case "View Service Status":
		_, err := m.manager.GetStatus()
		if err != nil {
			m.err = err
		} else {
			return m, func() tea.Msg {
				return ExecutionStartMsg{
					Command:     "systemctl status mysql",
					Description: "MySQL Service Status",
				}
			}
		}

	case "List Databases":
		databases, err := m.manager.ListDatabases()
		if err != nil {
			m.err = err
		} else {
			if len(databases) > 0 {
				m.success = fmt.Sprintf("✓ Found %d databases: %v", len(databases), databases)
			} else {
				m.success = "No user databases found"
			}
		}

	case "← Back to Configurations":
		return m, func() tea.Msg {
			return NavigateMsg{Screen: ConfigMenuScreen}
		}
	}

	return m, nil
}

// View renders the MySQL management screen
func (m MySQLManagementModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	header := m.theme.Title.Render("MySQL Management")

	// Current config info
	var configInfo []string
	if m.config != nil {
		configInfo = append(configInfo, m.theme.Label.Render("Current Configuration:"))
		configInfo = append(configInfo, m.theme.MenuItem.Render(fmt.Sprintf("  Port: %d", m.config.Port)))
		configInfo = append(configInfo, m.theme.MenuItem.Render(fmt.Sprintf("  Bind Address: %s", m.config.BindAddress)))
		configInfo = append(configInfo, m.theme.DescriptionStyle.Render(fmt.Sprintf("  Config: %s", m.config.ConfigPath)))
		configInfo = append(configInfo, m.theme.DescriptionStyle.Render(fmt.Sprintf("  Data Dir: %s", m.config.DataDir)))
	} else {
		configInfo = append(configInfo, m.theme.WarningStyle.Render("Configuration not loaded"))
	}

	configInfoSection := lipgloss.JoinVertical(lipgloss.Left, configInfo...)

	// Actions menu
	var actionItems []string
	for i, action := range m.actions {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render(m.theme.Symbols.Cursor + " ")
		}

		var renderedItem string
		if i == m.cursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, action))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, action))
		}

		actionItems = append(actionItems, renderedItem)
	}

	actionsMenu := lipgloss.JoinVertical(lipgloss.Left, actionItems...)

	// Messages
	var messages []string
	if m.success != "" {
		messages = append(messages, m.theme.SuccessStyle.Render(m.success))
	}
	if m.err != nil {
		messages = append(messages, m.theme.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}
	if m.copied {
		messages = append(messages, m.theme.CopiedStyle.Render(m.theme.Symbols.Copy+" Copied to clipboard!"))
	}
	messageSection := ""
	if len(messages) > 0 {
		messageSection = lipgloss.JoinVertical(lipgloss.Left, messages...)
	}

	// Help
	help := m.theme.Help.Render(m.theme.Symbols.ArrowUp + "/" + m.theme.Symbols.ArrowDown + ": Navigate " + m.theme.Symbols.Bullet + " Enter: Execute " + m.theme.Symbols.Bullet + " c: Copy " + m.theme.Symbols.Bullet + " Esc: Back " + m.theme.Symbols.Bullet + " q: Quit")

	// Combine all sections
	sections := []string{
		header,
		"",
		configInfoSection,
		"",
		"",
		m.theme.Subtitle.Render("Actions:"),
		"",
		actionsMenu,
	}

	if messageSection != "" {
		sections = append(sections, "", messageSection)
	}

	sections = append(sections, "", help)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

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

// SetSuccess sets a success message (called when returning from sub-screens)
func (m *MySQLManagementModel) SetSuccess(msg string) {
	m.success = msg
}
