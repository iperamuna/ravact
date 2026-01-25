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

// PostgreSQLPortModel represents the PostgreSQL port change screen
type PostgreSQLPortModel struct {
	theme   *theme.Theme
	width   int
	height  int
	manager *system.PostgreSQLManager
	config  *system.PostgreSQLConfig
	form    *huh.Form
	port    string
	err     error
}

// NewPostgreSQLPortModel creates a new PostgreSQL port change model
func NewPostgreSQLPortModel(manager *system.PostgreSQLManager, config *system.PostgreSQLConfig) PostgreSQLPortModel {
	t := theme.DefaultTheme()

	m := PostgreSQLPortModel{
		theme:   t,
		manager: manager,
		config:  config,
		port:    "5432",
	}

	if config != nil {
		m.port = fmt.Sprintf("%d", config.Port)
	}

	m.form = m.buildForm()
	return m
}

func (m *PostgreSQLPortModel) buildForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("New Port").
				Description("Enter port number (1024-65535). Service will be restarted.").
				Placeholder("5432").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("port cannot be empty")
					}
					port, err := strconv.Atoi(s)
					if err != nil {
						return fmt.Errorf("invalid port number")
					}
					if port < 1024 || port > 65535 {
						return fmt.Errorf("port must be between 1024-65535")
					}
					return nil
				}).
				Value(&m.port),
		),
	).WithTheme(m.theme.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)
}

func (m PostgreSQLPortModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m PostgreSQLPortModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.form.State == huh.StateNormal {
				return m, func() tea.Msg {
					return NavigateMsg{Screen: PostgreSQLManagementScreen}
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
		port, _ := strconv.Atoi(m.port)

		if err := m.manager.ChangePort(port); err != nil {
			m.err = err
			m.form = m.buildForm()
			return m, nil
		}

		// Restart service
		if err := m.manager.RestartService(); err != nil {
			m.err = fmt.Errorf("port changed but failed to restart: %w", err)
			return m, nil
		}

		// Success
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: PostgreSQLManagementScreen,
				Data: map[string]interface{}{
					"success": fmt.Sprintf("Port changed to %d and service restarted", port),
				},
			}
		}
	}

	return m, cmd
}

func (m PostgreSQLPortModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.theme.Title.Render("Change PostgreSQL Port")

	var content []string
	content = append(content, header)
	content = append(content, "")

	if m.config != nil {
		content = append(content, m.theme.Label.Render(fmt.Sprintf("Current Port: %d", m.config.Port)))
		content = append(content, "")
	}

	if m.err != nil {
		content = append(content, m.theme.ErrorStyle.Render(m.theme.Symbols.CrossMark+" Error: "+m.err.Error()))
		content = append(content, "")
	}

	content = append(content, m.form.View())
	content = append(content, "")
	content = append(content, m.theme.WarningStyle.Render(m.theme.Symbols.Warning+" Service will be restarted"))
	content = append(content, "")
	content = append(content, m.theme.Help.Render("Enter: Submit "+m.theme.Symbols.Bullet+" Esc: Cancel"))

	body := lipgloss.JoinVertical(lipgloss.Left, content...)
	bordered := m.theme.BorderStyle.Render(body)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}
