package screens

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// RedisPortModel represents the Redis port change screen
type RedisPortModel struct {
	theme        *theme.Theme
	width        int
	height       int
	redisManager *system.RedisManager
	config       *system.RedisConfig
	port         string
	currentField int // 0 = port, 1 = submit
	err          error
	success      bool
}

// NewRedisPortModel creates a new Redis port model
func NewRedisPortModel(config *system.RedisConfig) RedisPortModel {
	return RedisPortModel{
		theme:        theme.DefaultTheme(),
		redisManager: system.NewRedisManager(),
		config:       config,
		port:         config.Port,
		currentField: 0,
	}
}

// Init initializes the screen
func (m RedisPortModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m RedisPortModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// If showing success/error, any key returns
		if m.success || m.err != nil {
			if msg.String() == "enter" || msg.String() == " " || msg.String() == "esc" {
				return m, func() tea.Msg {
					return NavigateMsg{Screen: RedisConfigScreen}
				}
			}
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: RedisConfigScreen}
			}

		case "tab", "down":
			m.currentField = (m.currentField + 1) % 2

		case "shift+tab", "up":
			m.currentField = (m.currentField - 1 + 2) % 2

		case "enter":
			if m.currentField == 1 {
				return m.changePort()
			}

		case "backspace":
			if m.currentField == 0 && len(m.port) > 0 {
				m.port = m.port[:len(m.port)-1]
			}

		default:
			if len(msg.String()) == 1 && m.currentField == 0 {
				// Only allow digits
				if msg.String() >= "0" && msg.String() <= "9" {
					m.port += msg.String()
				}
			}
		}
	}

	return m, nil
}

// changePort changes the Redis port
func (m RedisPortModel) changePort() (RedisPortModel, tea.Cmd) {
	// Validate
	if m.port == "" {
		m.err = fmt.Errorf("port cannot be empty")
		return m, nil
	}
	
	portNum, err := strconv.Atoi(m.port)
	if err != nil || portNum < 1 || portNum > 65535 {
		m.err = fmt.Errorf("port must be between 1 and 65535")
		return m, nil
	}

	// Set port
	err = m.redisManager.SetPort(m.port)
	if err != nil {
		m.err = err
		return m, nil
	}

	// Restart Redis
	err = m.redisManager.RestartRedis()
	if err != nil {
		m.err = fmt.Errorf("port set but restart failed: %w", err)
		return m, nil
	}

	m.success = true
	m.err = nil
	return m, nil
}

// View renders the screen
func (m RedisPortModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// If success or error, show message
	if m.success {
		msg := m.theme.SuccessStyle.Render("✓ Redis port changed successfully!")
		help := m.theme.Help.Render("Press any key to continue...")
		content := lipgloss.JoinVertical(lipgloss.Center, "", msg, "", help)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	if m.err != nil {
		msg := m.theme.ErrorStyle.Render(fmt.Sprintf("✗ Error: %v", m.err))
		help := m.theme.Help.Render("Press any key to continue...")
		content := lipgloss.JoinVertical(lipgloss.Center, "", msg, "", help)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	// Header
	header := m.theme.Title.Render("Change Redis Port")

	// Current port info
	currentInfo := m.theme.DescriptionStyle.Render(fmt.Sprintf("Current port: %s", m.config.Port))

	// Form fields
	var formFields []string

	// Port field
	portStyle := m.theme.MenuItem
	if m.currentField == 0 {
		portStyle = m.theme.SelectedItem
	}
	formFields = append(formFields, portStyle.Render(fmt.Sprintf("New Port: %s_", m.port)))

	// Submit button
	submitStyle := m.theme.MenuItem
	if m.currentField == 1 {
		submitStyle = m.theme.SelectedItem
	}
	formFields = append(formFields, "")
	formFields = append(formFields, submitStyle.Render("[ Change Port ]"))

	form := lipgloss.JoinVertical(lipgloss.Left, formFields...)

	// Help
	help := m.theme.Help.Render("Tab/↑↓: Navigate • Enter: Submit • Esc: Cancel • q: Quit")

	// Instructions
	instructions := m.theme.DescriptionStyle.Render("Port must be between 1 and 65535. Default is 6379.")

	// Warning
	warning := m.theme.WarningStyle.Render("⚠ Changing port will require updating client connections")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		currentInfo,
		"",
		instructions,
		warning,
		"",
		form,
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
