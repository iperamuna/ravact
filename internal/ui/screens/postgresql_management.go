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

// PostgreSQLManagementModel represents the PostgreSQL management screen
type PostgreSQLManagementModel struct {
	theme       *theme.Theme
	width       int
	height      int
	manager     *system.PostgreSQLManager
	config      *system.PostgreSQLConfig
	cursor      int
	actions     []string
	err         error
	success     string
	copied      bool
	copiedTimer int
}

// NewPostgreSQLManagementModel creates a new PostgreSQL management model
func NewPostgreSQLManagementModel() PostgreSQLManagementModel {
	manager := system.NewPostgreSQLManager()
	config, _ := manager.GetConfig()
	
	actions := []string{
		"View Current Configuration",
		"Change Postgres Password",
		"Change Port",
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
		case "c":
			// Copy configuration to clipboard
			if m.config != nil {
				content := fmt.Sprintf("PostgreSQL Configuration\nPort: %d\nMax Connections: %d\nShared Buffers: %s\nConfig Path: %s",
					m.config.Port, m.config.MaxConn, m.config.SharedBuf, m.config.ConfigPath)
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

	case "Change Postgres Password":
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: PostgreSQLPasswordScreen,
				Data: map[string]interface{}{
					"manager": m.manager,
				},
			}
		}

	case "Change Port":
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: PostgreSQLPortScreen,
				Data: map[string]interface{}{
					"manager": m.manager,
					"config":  m.config,
				},
			}
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

	header := m.theme.Title.Render("PostgreSQL Management")

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

	help := m.theme.Help.Render(m.theme.Symbols.ArrowUp + "/" + m.theme.Symbols.ArrowDown + ": Navigate " + m.theme.Symbols.Bullet + " Enter: Execute " + m.theme.Symbols.Bullet + " c: Copy " + m.theme.Symbols.Bullet + " Esc: Back " + m.theme.Symbols.Bullet + " q: Quit")

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
	bordered := m.theme.RenderBox(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}

// SetSuccess sets a success message (called when returning from sub-screens)
func (m *PostgreSQLManagementModel) SetSuccess(msg string) {
	m.success = msg
}
