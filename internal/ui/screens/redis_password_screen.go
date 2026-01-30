package screens

import (
	"fmt"

	"github.com/charmbracelet/huh"
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
	form         *huh.Form
	password     string
	confirm      string
	err          error
	success      bool
}

// NewRedisPasswordModel creates a new Redis password model
func NewRedisPasswordModel(config *system.RedisConfig) RedisPasswordModel {
	t := theme.DefaultTheme()

	m := RedisPasswordModel{
		theme:        t,
		redisManager: system.NewRedisManager(),
		config:       config,
		password:     "",
		confirm:      "",
	}

	m.form = m.buildForm()
	return m
}

func (m *RedisPasswordModel) buildForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("New Password").
				Description("Password must be at least 8 characters").
				Placeholder("Enter new password...").
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("password cannot be empty")
					}
					if len(s) < 8 {
						return fmt.Errorf("password must be at least 8 characters")
					}
					return nil
				}).
				Value(&m.password),

			huh.NewInput().
				Title("Confirm Password").
				Description("Re-enter the password to confirm").
				Placeholder("Confirm password...").
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("please confirm password")
					}
					return nil
				}).
				Value(&m.confirm),
		),
	).WithTheme(m.theme.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)
}

// Init initializes the screen
func (m RedisPasswordModel) Init() tea.Cmd {
	return m.form.Init()
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
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.form.State == huh.StateNormal {
				return m, func() tea.Msg {
					return NavigateMsg{Screen: RedisConfigScreen}
				}
			}
		}
	}

	// Update the form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	// Check if form is completed
	if m.form.State == huh.StateCompleted {
		return m.changePassword()
	}

	return m, cmd
}

// changePassword changes the Redis password
func (m RedisPasswordModel) changePassword() (RedisPasswordModel, tea.Cmd) {
	// Validate passwords match
	if m.password != m.confirm {
		m.err = fmt.Errorf("passwords do not match")
		m.form = m.buildForm()
		return m, nil
	}

	// Set password
	err := m.redisManager.SetPassword(m.password)
	if err != nil {
		m.err = err
		m.form = m.buildForm()
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

	// If success, show message
	if m.success {
		msg := m.theme.SuccessStyle.Render(m.theme.Symbols.CheckMark + " Redis password changed successfully!")
		help := m.theme.Help.Render("Press any key to continue...")
		content := lipgloss.JoinVertical(lipgloss.Center, "", msg, "", help)
		bordered := m.theme.RenderBox(content)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
	}

	// If error, show message
	if m.err != nil {
		msg := m.theme.ErrorStyle.Render(m.theme.Symbols.CrossMark + " Error: " + m.err.Error())
		help := m.theme.Help.Render("Press any key to continue...")
		content := lipgloss.JoinVertical(lipgloss.Center, "", msg, "", help)
		bordered := m.theme.RenderBox(content)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
	}

	// Header
	header := m.theme.Title.Render("Change Redis Password")

	// Current password info
	currentInfo := ""
	if m.config != nil && m.config.RequirePass != "" {
		currentInfo = m.theme.DescriptionStyle.Render("Current password: ********")
	} else {
		currentInfo = m.theme.WarningStyle.Render(m.theme.Symbols.Warning + " No password currently set")
	}

	// Help
	help := m.theme.Help.Render("Tab/Shift+Tab: Navigate " + m.theme.Symbols.Bullet + " Enter: Submit " + m.theme.Symbols.Bullet + " Esc: Cancel")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		currentInfo,
		"",
		m.form.View(),
		"",
		help,
	)

	// Add border and center
	bordered := m.theme.RenderBox(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}
