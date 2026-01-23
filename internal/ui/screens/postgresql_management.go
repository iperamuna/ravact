package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// PostgreSQLManagementModel represents the PostgreSQL management screen
type PostgreSQLManagementModel struct {
	theme   *theme.Theme
	width   int
	height  int
	manager *system.PostgreSQLManager
	config  *system.PostgreSQLConfig
	cursor  int
	actions []string
	err     error
	success string
}

// NewPostgreSQLManagementModel creates a new PostgreSQL management model
func NewPostgreSQLManagementModel() PostgreSQLManagementModel {
	manager := system.NewPostgreSQLManager()
	config, _ := manager.GetConfig()
	
	actions := []string{
		"View Current Configuration",
		"Restart PostgreSQL Service",
		"View Service Status",
		"List Databases",
		"← Back to Configurations",
	}
	
	return PostgreSQLManagementModel{
		theme:   theme.DefaultTheme(),
		manager: manager,
		config:  config,
		cursor:  0,
		actions: actions,
	}
}

func (m PostgreSQLManagementModel) Init() tea.Cmd {
	return nil
}

func (m PostgreSQLManagementModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		}
	}
	return m, nil
}

func (m PostgreSQLManagementModel) executeAction() (PostgreSQLManagementModel, tea.Cmd) {
	m.err = nil
	m.success = ""
	
	switch m.actions[m.cursor] {
	case "View Current Configuration":
		config, err := m.manager.GetConfig()
		if err != nil {
			m.err = err
		} else {
			m.config = config
			m.success = "✓ Configuration refreshed"
		}

	case "Restart PostgreSQL Service":
		err := m.manager.RestartService()
		if err != nil {
			m.err = err
		} else {
			m.success = "✓ PostgreSQL service restarted successfully"
		}

	case "View Service Status":
		_, err := m.manager.GetStatus()
		if err != nil {
			m.err = err
		} else {
			return m, func() tea.Msg {
				return ExecutionStartMsg{
					Command:     "systemctl status postgresql",
					Description: "PostgreSQL Service Status",
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

func (m PostgreSQLManagementModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.theme.Title.Render("🐘 PostgreSQL Management")

	var configInfo []string
	if m.config != nil {
		configInfo = append(configInfo, m.theme.Label.Render("Current Configuration:"))
		configInfo = append(configInfo, m.theme.MenuItem.Render(fmt.Sprintf("  Port: %d", m.config.Port)))
		configInfo = append(configInfo, m.theme.MenuItem.Render(fmt.Sprintf("  Max Connections: %d", m.config.MaxConn)))
		configInfo = append(configInfo, m.theme.MenuItem.Render(fmt.Sprintf("  Shared Buffers: %s", m.config.SharedBuf)))
		configInfo = append(configInfo, m.theme.DescriptionStyle.Render(fmt.Sprintf("  Config: %s", m.config.ConfigPath)))
	} else {
		configInfo = append(configInfo, m.theme.WarningStyle.Render("Configuration not loaded"))
	}
	
	configInfoSection := lipgloss.JoinVertical(lipgloss.Left, configInfo...)

	var actionItems []string
	for i, action := range m.actions {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
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

	var messages []string
	if m.success != "" {
		messages = append(messages, m.theme.SuccessStyle.Render(m.success))
	}
	if m.err != nil {
		messages = append(messages, m.theme.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}
	messageSection := ""
	if len(messages) > 0 {
		messageSection = lipgloss.JoinVertical(lipgloss.Left, messages...)
	}

	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Execute • Esc: Back • q: Quit")

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
	bordered := m.theme.BorderStyle.Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}
