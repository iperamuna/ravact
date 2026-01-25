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

// RedisConfigAction represents an action on Redis config
type RedisConfigAction int

const (
	RedisActionChangePassword RedisConfigAction = iota
	RedisActionChangePort
	RedisActionTestConnection
	RedisActionRestart
	RedisActionViewConfig
	RedisActionBack
)

// RedisConfigModel represents the Redis configuration screen
type RedisConfigModel struct {
	theme        *theme.Theme
	width        int
	height       int
	redisManager *system.RedisManager
	config       *system.RedisConfig
	cursor       int
	actions      []string
	status       string
	err          error
	success      string
	copied       bool
	copiedTimer  int
}

// NewRedisConfigModel creates a new Redis config model
func NewRedisConfigModel() RedisConfigModel {
	redisManager := system.NewRedisManager()
	config, _ := redisManager.GetConfig()
	status, _ := redisManager.GetStatus()
	
	actions := []string{
		"Change Password",
		"Change Port",
		"Test Connection",
		"Restart Redis",
		"View Configuration File",
		"← Back to Configurations",
	}
	
	return RedisConfigModel{
		theme:        theme.DefaultTheme(),
		redisManager: redisManager,
		config:       config,
		cursor:       0,
		actions:      actions,
		status:       status,
	}
}

// Init initializes the Redis config screen
func (m RedisConfigModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m RedisConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				passStatus := "Not Set"
				if m.config.RequirePass != "" {
					passStatus = "Set"
				}
				content := fmt.Sprintf("Redis Configuration\nPort: %s\nPassword: %s\nStatus: %s\nConfig Path: %s",
					m.config.Port, passStatus, m.status, m.config.ConfigPath)
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
func (m RedisConfigModel) executeAction() (RedisConfigModel, tea.Cmd) {
	m.err = nil
	m.success = ""
	
	actionName := m.actions[m.cursor]

	switch actionName {
	case "Change Password":
		// Navigate to password change screen
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: RedisPasswordScreen,
				Data: map[string]interface{}{
					"config": m.config,
				},
			}
		}

	case "Change Port":
		// Navigate to port change screen
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: RedisPortScreen,
				Data: map[string]interface{}{
					"config": m.config,
				},
			}
		}

	case "Test Connection":
		err := m.redisManager.TestConnection()
		if err != nil {
			m.err = err
		} else {
			m.success = "✓ Redis connection successful!"
		}

	case "Restart Redis":
		err := m.redisManager.RestartRedis()
		if err != nil {
			m.err = err
		} else {
			m.success = "✓ Redis restarted successfully"
			// Refresh status
			m.status, _ = m.redisManager.GetStatus()
		}

	case "View Configuration File":
		// Open config file in editor selection
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     fmt.Sprintf("cat %s | less", m.config.ConfigPath),
				Description: fmt.Sprintf("Viewing %s", m.config.ConfigPath),
			}
		}

	case "← Back to Configurations":
		return m, func() tea.Msg {
			return NavigateMsg{Screen: ConfigMenuScreen}
		}
	}

	return m, nil
}

// View renders the Redis config screen
func (m RedisConfigModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	header := m.theme.Title.Render("Redis Configuration")

	// Current config info
	var configInfo []string
	if m.config != nil {
		configInfo = append(configInfo, m.theme.Label.Render("Current Configuration:"))
		configInfo = append(configInfo, m.theme.MenuItem.Render(fmt.Sprintf("  Port: %s", m.config.Port)))

		if m.config.RequirePass != "" {
			configInfo = append(configInfo, m.theme.MenuItem.Render(fmt.Sprintf("  Password: %s", "********")))
		} else {
			configInfo = append(configInfo, m.theme.WarningStyle.Render("  Password: Not Set (Insecure!)"))
		}

		configInfo = append(configInfo, m.theme.DescriptionStyle.Render(fmt.Sprintf("  Config: %s", m.config.ConfigPath)))
	}

	// Status
	statusStyle := m.theme.DescriptionStyle
	statusText := m.status
	if m.status == "active" {
		statusStyle = m.theme.SuccessStyle
		statusText = "Running"
	} else if m.status == "inactive" {
		statusStyle = m.theme.ErrorStyle
		statusText = "Stopped"
	}
	configInfo = append(configInfo, m.theme.Label.Render("Status: ")+statusStyle.Render(statusText))

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
