package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// PostgreSQLPasswordModel represents the PostgreSQL password change screen
type PostgreSQLPasswordModel struct {
	theme     *theme.Theme
	width     int
	height    int
	manager   *system.PostgreSQLManager
	textInput textinput.Model
	err       error
}

// NewPostgreSQLPasswordModel creates a new PostgreSQL password change model
func NewPostgreSQLPasswordModel(manager *system.PostgreSQLManager) PostgreSQLPasswordModel {
	ti := textinput.New()
	ti.Placeholder = "Enter new postgres user password"
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 40
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

	return PostgreSQLPasswordModel{
		theme:     theme.DefaultTheme(),
		manager:   manager,
		textInput: ti,
	}
}

func (m PostgreSQLPasswordModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m PostgreSQLPasswordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: PostgreSQLManagementScreen}
			}
		case "enter":
			password := m.textInput.Value()
			if password == "" {
				m.err = fmt.Errorf("password cannot be empty")
				return m, nil
			}

			err := m.manager.ChangeRootPassword(password)
			if err != nil {
				m.err = err
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
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m PostgreSQLPasswordModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.theme.Title.Render("🔐 Change PostgreSQL Password")
	
	var content []string
	content = append(content, header)
	content = append(content, "")
	
	if m.err != nil {
		content = append(content, m.theme.ErrorStyle.Render("Error: "+m.err.Error()))
		content = append(content, "")
	}
	
	content = append(content, m.theme.Label.Render("New Postgres User Password:"))
	content = append(content, m.textInput.View())
	content = append(content, "")
	content = append(content, m.theme.Help.Render("Enter: Save • Esc: Cancel • q: Quit"))

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
