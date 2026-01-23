package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// RedisPasswordModel represents the Redis password change screen
type RedisPasswordModel struct {
	theme        *theme.Theme
	width        int
	height       int
	redisManager *system.RedisManager
	config       *system.RedisConfig
	password     string
	confirm      string
	currentField int // 0 = password, 1 = confirm, 2 = submit
	err          error
	success      bool
}

// NewRedisPasswordModel creates a new Redis password model
func NewRedisPasswordModel(config *system.RedisConfig) RedisPasswordModel {
	return RedisPasswordModel{
		theme:        theme.DefaultTheme(),
		redisManager: system.NewRedisManager(),
		config:       config,
		currentField: 0,
	}
}

// Init initializes the screen
func (m RedisPasswordModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m RedisPasswordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.currentField = (m.currentField + 1) % 3

		case "shift+tab", "up":
			m.currentField = (m.currentField - 1 + 3) % 3

		case "enter":
			if m.currentField == 2 {
				return m.changePassword()
			}

		case "backspace":
			if m.currentField == 0 && len(m.password) > 0 {
				m.password = m.password[:len(m.password)-1]
			} else if m.currentField == 1 && len(m.confirm) > 0 {
				m.confirm = m.confirm[:len(m.confirm)-1]
			}

		default:
			if len(msg.String()) == 1 {
				if m.currentField == 0 {
					m.password += msg.String()
				} else if m.currentField == 1 {
					m.confirm += msg.String()
				}
			}
		}
	}

	return m, nil
}

// changePassword changes the Redis password
func (m RedisPasswordModel) changePassword() (RedisPasswordModel, tea.Cmd) {
	// Validate
	if m.password == "" {
		m.err = fmt.Errorf("password cannot be empty")
		return m, nil
	}
	if m.password != m.confirm {
		m.err = fmt.Errorf("passwords do not match")
		return m, nil
	}
	if len(m.password) < 8 {
		m.err = fmt.Errorf("password must be at least 8 characters")
		return m, nil
	}

	// Set password
	err := m.redisManager.SetPassword(m.password)
	if err != nil {
		m.err = err
		return m, nil
	}

	// Restart Redis
	err = m.redisManager.RestartRedis()
	if err != nil {
		m.err = fmt.Errorf("password set but restart failed: %w", err)
		return m, nil
	}

	m.success = true
	m.err = nil
	return m, nil
}

// View renders the screen
func (m RedisPasswordModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// If success or error, show message
	if m.success {
		msg := m.theme.SuccessStyle.Render("✓ Redis password changed successfully!")
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
	header := m.theme.Title.Render("Change Redis Password")

	// Current password info
	currentInfo := ""
	if m.config != nil && m.config.RequirePass != "" {
		currentInfo = m.theme.DescriptionStyle.Render("Current password: ********")
	} else {
		currentInfo = m.theme.WarningStyle.Render("No password currently set")
	}

	// Form fields
	var formFields []string

	// Password field
	passwordStyle := m.theme.MenuItem
	if m.currentField == 0 {
		passwordStyle = m.theme.SelectedItem
	}
	passwordDisplay := ""
	for i := 0; i < len(m.password); i++ {
		passwordDisplay += "*"
	}
	formFields = append(formFields, passwordStyle.Render(fmt.Sprintf("New Password:     %s_", passwordDisplay)))

	// Confirm field
	confirmStyle := m.theme.MenuItem
	if m.currentField == 1 {
		confirmStyle = m.theme.SelectedItem
	}
	confirmDisplay := ""
	for i := 0; i < len(m.confirm); i++ {
		confirmDisplay += "*"
	}
	formFields = append(formFields, confirmStyle.Render(fmt.Sprintf("Confirm Password: %s_", confirmDisplay)))

	// Submit button
	submitStyle := m.theme.MenuItem
	if m.currentField == 2 {
		submitStyle = m.theme.SelectedItem
	}
	formFields = append(formFields, "")
	formFields = append(formFields, submitStyle.Render("[ Change Password ]"))

	form := lipgloss.JoinVertical(lipgloss.Left, formFields...)

	// Help
	help := m.theme.Help.Render("Tab/↑↓: Navigate • Enter: Submit • Esc: Cancel • q: Quit")

	// Instructions
	instructions := m.theme.DescriptionStyle.Render("Password must be at least 8 characters")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		currentInfo,
		"",
		instructions,
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
