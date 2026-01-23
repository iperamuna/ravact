package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// PostgreSQLPortModel represents the PostgreSQL port change screen
type PostgreSQLPortModel struct {
	theme       *theme.Theme
	width       int
	height      int
	manager     *system.PostgreSQLManager
	config      *system.PostgreSQLConfig
	textInput   textinput.Model
	err         error
}

// NewPostgreSQLPortModel creates a new PostgreSQL port change model
func NewPostgreSQLPortModel(manager *system.PostgreSQLManager, config *system.PostgreSQLConfig) PostgreSQLPortModel {
	ti := textinput.New()
	ti.Placeholder = "Enter new port (1024-65535)"
	ti.Focus()
	ti.CharLimit = 5
	ti.Width = 30

	if config != nil {
		ti.SetValue(fmt.Sprintf("%d", config.Port))
	}

	return PostgreSQLPortModel{
		theme:     theme.DefaultTheme(),
		manager:   manager,
		config:    config,
		textInput: ti,
	}
}

func (m PostgreSQLPortModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m PostgreSQLPortModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			portStr := m.textInput.Value()
			var port int
			if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
				m.err = fmt.Errorf("invalid port number")
				return m, nil
			}

			if err := m.manager.ChangePort(port); err != nil {
				m.err = err
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
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m PostgreSQLPortModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.theme.Title.Render("🔌 Change PostgreSQL Port")
	
	var content []string
	content = append(content, header)
	content = append(content, "")
	
	if m.config != nil {
		content = append(content, m.theme.Label.Render(fmt.Sprintf("Current Port: %d", m.config.Port)))
		content = append(content, "")
	}
	
	if m.err != nil {
		content = append(content, m.theme.ErrorStyle.Render("Error: "+m.err.Error()))
		content = append(content, "")
	}
	
	content = append(content, m.theme.Label.Render("New Port:"))
	content = append(content, m.textInput.View())
	content = append(content, "")
	content = append(content, m.theme.DescriptionStyle.Render("Note: Service will be restarted"))
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
