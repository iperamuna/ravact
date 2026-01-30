package screens

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
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
	form         *huh.Form
	port         string
	err          error
	success      bool
}

// NewRedisPortModel creates a new Redis port model
func NewRedisPortModel(config *system.RedisConfig) RedisPortModel {
	t := theme.DefaultTheme()

	m := RedisPortModel{
		theme:        t,
		redisManager: system.NewRedisManager(),
		config:       config,
		port:         config.Port,
	}

	m.form = m.buildForm()
	return m
}

func (m *RedisPortModel) buildForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("New Port").
				Description("Port must be between 1-65535. Default is 6379.").
				Placeholder("6379").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("port cannot be empty")
					}
					portNum, err := strconv.Atoi(s)
					if err != nil {
						return fmt.Errorf("invalid port number")
					}
					if portNum < 1 || portNum > 65535 {
						return fmt.Errorf("port must be between 1-65535")
					}
					return nil
				}).
				Value(&m.port),
		),
	).WithTheme(m.theme.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)
}

// Init initializes the screen
func (m RedisPortModel) Init() tea.Cmd {
	return m.form.Init()
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
		return m.changePort()
	}

	return m, cmd
}

// changePort changes the Redis port
func (m RedisPortModel) changePort() (RedisPortModel, tea.Cmd) {
	// Set port
	err := m.redisManager.SetPort(m.port)
	if err != nil {
		m.err = err
		m.form = m.buildForm()
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

	// If success, show message
	if m.success {
		msg := m.theme.SuccessStyle.Render(m.theme.Symbols.CheckMark + " Redis port changed successfully!")
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
	header := m.theme.Title.Render("Change Redis Port")

	// Current port info
	currentInfo := m.theme.Label.Render(fmt.Sprintf("Current port: %s", m.config.Port))

	// Warning
	warning := m.theme.WarningStyle.Render(m.theme.Symbols.Warning + " Changing port will require updating client connections")

	// Help
	help := m.theme.Help.Render("Enter: Submit " + m.theme.Symbols.Bullet + " Esc: Cancel")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		currentInfo,
		"",
		m.form.View(),
		"",
		warning,
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
