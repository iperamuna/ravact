package screens

import (
	"fmt"

	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// MySQLPasswordModel represents the MySQL password change screen
type MySQLPasswordModel struct {
	theme    *theme.Theme
	width    int
	height   int
	manager  *system.MySQLManager
	form     *huh.Form
	password string
	err      error
}

// NewMySQLPasswordModel creates a new MySQL password change model
func NewMySQLPasswordModel(manager *system.MySQLManager) MySQLPasswordModel {
	t := theme.DefaultTheme()

	m := MySQLPasswordModel{
		theme:    t,
		manager:  manager,
		password: "",
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("New Root Password").
				Description("Enter a strong password for the MySQL root user").
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
	).WithTheme(t.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)

	return m
}

func (m MySQLPasswordModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m MySQLPasswordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					return NavigateMsg{Screen: MySQLManagementScreen}
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
			// Rebuild form to allow retry
			m.form = huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("New Root Password").
						Description("Enter a strong password for the MySQL root user").
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
			return m, nil
		}

		// Success - navigate back
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: MySQLManagementScreen,
				Data: map[string]interface{}{
					"success": "Root password changed successfully",
				},
			}
		}
	}

	return m, cmd
}

func (m MySQLPasswordModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.theme.Title.Render("Change MySQL Root Password")

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
	bordered := m.theme.RenderBox(body)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}
