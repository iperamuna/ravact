package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// MySQLPasswordModel represents the MySQL password change screen
type MySQLPasswordModel struct {
	theme     *theme.Theme
	width     int
	height    int
	manager   *system.MySQLManager
	textInput textinput.Model
	err       error
}

// NewMySQLPasswordModel creates a new MySQL password change model
func NewMySQLPasswordModel(manager *system.MySQLManager) MySQLPasswordModel {
	ti := textinput.New()
	ti.Placeholder = "Enter new root password"
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 40
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

	return MySQLPasswordModel{
		theme:     theme.DefaultTheme(),
		manager:   manager,
		textInput: ti,
	}
}

func (m MySQLPasswordModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m MySQLPasswordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				return NavigateMsg{Screen: MySQLManagementScreen}
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
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m MySQLPasswordModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.theme.Title.Render("🔐 Change MySQL Root Password")
	
	var content []string
	content = append(content, header)
	content = append(content, "")
	
	if m.err != nil {
		content = append(content, m.theme.ErrorStyle.Render("Error: "+m.err.Error()))
		content = append(content, "")
	}
	
	content = append(content, m.theme.Label.Render("New Root Password:"))
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
