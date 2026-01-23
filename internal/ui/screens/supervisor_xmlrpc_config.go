package screens

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// SupervisorXMLRPCConfigModel represents the XML-RPC configuration screen
type SupervisorXMLRPCConfigModel struct {
	theme      *theme.Theme
	width      int
	height     int
	manager    *system.SupervisorManager
	inputs     []textinput.Model
	focusIndex int
	err        error
}

// NewSupervisorXMLRPCConfigModel creates a new XML-RPC config model
func NewSupervisorXMLRPCConfigModel(manager *system.SupervisorManager) SupervisorXMLRPCConfigModel {
	// Load current config
	config, _ := manager.GetXMLRPCConfig()
	
	inputs := make([]textinput.Model, 4)
	
	// IP Address
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "IP Address (e.g., 127.0.0.1)"
	inputs[0].Focus()
	inputs[0].Width = 40
	if config != nil && config.IP != "" {
		inputs[0].SetValue(config.IP)
	} else {
		inputs[0].SetValue("127.0.0.1")
	}
	
	// Port
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Port (e.g., 9001)"
	inputs[1].Width = 40
	if config != nil && config.Port != "" {
		inputs[1].SetValue(config.Port)
	} else {
		inputs[1].SetValue("9001")
	}
	
	// Username
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Username"
	inputs[2].Width = 40
	if config != nil {
		inputs[2].SetValue(config.Username)
	}
	
	// Password
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Password"
	inputs[3].Width = 40
	inputs[3].EchoMode = textinput.EchoPassword
	inputs[3].EchoCharacter = '•'

	return SupervisorXMLRPCConfigModel{
		theme:      theme.DefaultTheme(),
		manager:    manager,
		inputs:     inputs,
		focusIndex: 0,
	}
}

func (m SupervisorXMLRPCConfigModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SupervisorXMLRPCConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				return NavigateMsg{Screen: SupervisorManagementScreen}
			}
		case "tab", "shift+tab", "up", "down":
			if msg.String() == "up" || msg.String() == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}
			if m.focusIndex > 3 {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = 3
			}
			m.updateFocus()
			return m, nil
		case "enter":
			return m.saveConfig()
		}
	}

	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	return m, cmd
}

func (m SupervisorXMLRPCConfigModel) updateFocus() {
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m SupervisorXMLRPCConfigModel) saveConfig() (SupervisorXMLRPCConfigModel, tea.Cmd) {
	ip := m.inputs[0].Value()
	port := m.inputs[1].Value()
	username := m.inputs[2].Value()
	password := m.inputs[3].Value()

	if ip == "" {
		ip = "127.0.0.1"
	}
	if port == "" {
		port = "9001"
	}

	err := m.manager.SetXMLRPCConfig(ip, port, username, password)
	if err != nil {
		m.err = err
		return m, nil
	}

	// Success
	return m, func() tea.Msg {
		return NavigateMsg{
			Screen: SupervisorManagementScreen,
			Data: map[string]interface{}{
				"success": "XML-RPC configured successfully. Supervisor will restart.",
			},
		}
	}
}

func (m SupervisorXMLRPCConfigModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.theme.Title.Render("🔧 Configure XML-RPC Server")
	
	var content []string
	content = append(content, header)
	content = append(content, "")
	content = append(content, m.theme.DescriptionStyle.Render("Configure Supervisor XML-RPC interface for remote management"))
	content = append(content, "")
	
	if m.err != nil {
		content = append(content, m.theme.ErrorStyle.Render("Error: "+m.err.Error()))
		content = append(content, "")
	}
	
	content = append(content, m.theme.Label.Render("IP Address:"))
	content = append(content, m.inputs[0].View())
	content = append(content, "")
	
	content = append(content, m.theme.Label.Render("Port:"))
	content = append(content, m.inputs[1].View())
	content = append(content, "")
	
	content = append(content, m.theme.Label.Render("Username:"))
	content = append(content, m.inputs[2].View())
	content = append(content, "")
	
	content = append(content, m.theme.Label.Render("Password:"))
	content = append(content, m.inputs[3].View())
	content = append(content, "")
	
	content = append(content, m.theme.DescriptionStyle.Render("Note: Supervisor will be restarted after saving"))
	content = append(content, "")
	content = append(content, m.theme.Help.Render("Tab: Next Field • Enter: Save • Esc: Cancel • q: Quit"))

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
