package screens

import (
	"fmt"

	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// PostgreSQLPasswordModel represents the PostgreSQL password change screen
type PostgreSQLPasswordModel struct {
	theme    *theme.Theme
	width    int
	height   int
	manager  *system.PostgreSQLManager
	form     *huh.Form
	password string
	err      error
}

// NewPostgreSQLPasswordModel creates a new PostgreSQL password change model
func NewPostgreSQLPasswordModel(manager *system.PostgreSQLManager) PostgreSQLPasswordModel {
	t := theme.DefaultTheme()

	m := PostgreSQLPasswordModel{
		theme:    t,
		manager:  manager,
		password: "",
	}

	m.form = m.buildForm()
	return m
}

func (m *PostgreSQLPasswordModel) buildForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("New Postgres User Password").
				Description("Enter a strong password for the postgres user").
				Placeholder("Enter password...").
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("password cannot be empty")
					}
					if len(s) < 6 {
						return fmt.Errorf("password must be at least 6 characters")
					}
					return nil
				}).
				Value(&m.password),
		),
	).WithTheme(m.theme.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)
}

func (m PostgreSQLPasswordModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m PostgreSQLPasswordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		err := m.manager.ChangeRootPassword(m.password)
		if err != nil {
			m.err = err
			m.form = m.buildForm()
			return m, nil
		}

		// Success
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: PostgreSQLManagementScreen,
				Data: map[string]interface{}{
					"success": "Postgres user password changed successfully",
				},
			}
		}
	}

	return m, cmd
}

func (m PostgreSQLPasswordModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.theme.Title.Render("Change PostgreSQL Password")

	var content []string
	content = append(content, header)
	content = append(content, "")

	if m.err != nil {
		content = append(content, m.theme.ErrorStyle.Render(m.theme.Symbols.CrossMark+" Error: "+m.err.Error()))
		content = append(content, "")
	}

	content = append(content, m.form.View())
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
