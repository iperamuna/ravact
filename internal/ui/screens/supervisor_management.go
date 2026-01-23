package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// SupervisorManagementModel represents the Supervisor management screen
type SupervisorManagementModel struct {
	theme    *theme.Theme
	width    int
	height   int
	manager  *system.SupervisorManager
	programs []system.SupervisorProgram
	cursor   int
	actions  []string
	err      error
	success  string
}

// NewSupervisorManagementModel creates a new Supervisor management model
func NewSupervisorManagementModel() SupervisorManagementModel {
	manager := system.NewSupervisorManager()
	programs, _ := manager.GetAllPrograms()
	
	actions := []string{
		"List All Programs",
		"Restart Supervisor",
		"View XML-RPC Config",
		"← Back to Configurations",
	}
	
	return SupervisorManagementModel{
		theme:    theme.DefaultTheme(),
		manager:  manager,
		programs: programs,
		cursor:   0,
		actions:  actions,
	}
}

func (m SupervisorManagementModel) Init() tea.Cmd {
	return nil
}

func (m SupervisorManagementModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc", "backspace":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: ConfigMenuScreen}
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.actions)-1 {
				m.cursor++
			}
		case "enter", " ":
			return m.executeAction()
		}
	}
	return m, nil
}

func (m SupervisorManagementModel) executeAction() (SupervisorManagementModel, tea.Cmd) {
	m.err = nil
	m.success = ""
	
	switch m.actions[m.cursor] {
	case "List All Programs":
		programs, err := m.manager.GetAllPrograms()
		if err != nil {
			m.err = err
		} else {
			m.programs = programs
			m.success = fmt.Sprintf("✓ Found %d programs", len(programs))
		}

	case "Restart Supervisor":
		err := m.manager.RestartSupervisor()
		if err != nil {
			m.err = err
		} else {
			m.success = "✓ Supervisor restarted successfully"
		}

	case "View XML-RPC Config":
		config, err := m.manager.GetXMLRPCConfig()
		if err != nil {
			m.err = err
		} else {
			m.success = fmt.Sprintf("✓ XML-RPC: Enabled=%v, %s:%s", config.Enabled, config.IP, config.Port)
		}

	case "← Back to Configurations":
		return m, func() tea.Msg {
			return NavigateMsg{Screen: ConfigMenuScreen}
		}
	}

	return m, nil
}

func (m SupervisorManagementModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.theme.Title.Render("⚙️  Supervisor Management")

	var progInfo []string
	progInfo = append(progInfo, m.theme.Label.Render(fmt.Sprintf("Total Programs: %d", len(m.programs))))
	if len(m.programs) > 0 {
		for _, prog := range m.programs {
			stateStyle := m.theme.MenuItem
			if prog.State == "RUNNING" {
				stateStyle = m.theme.SuccessStyle
			} else if prog.State == "STOPPED" {
				stateStyle = m.theme.ErrorStyle
			}
			progInfo = append(progInfo, m.theme.MenuItem.Render(fmt.Sprintf("  • %s ", prog.Name))+stateStyle.Render(fmt.Sprintf("[%s]", prog.State)))
		}
	} else {
		progInfo = append(progInfo, m.theme.WarningStyle.Render("  No programs configured"))
	}
	
	progInfoSection := lipgloss.JoinVertical(lipgloss.Left, progInfo...)

	var actionItems []string
	for i, action := range m.actions {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}
		var renderedItem string
		if i == m.cursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, action))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, action))
		}
		actionItems = append(actionItems, renderedItem)
	}

	actionsMenu := lipgloss.JoinVertical(lipgloss.Left, actionItems...)

	var messages []string
	if m.success != "" {
		messages = append(messages, m.theme.SuccessStyle.Render(m.success))
	}
	if m.err != nil {
		messages = append(messages, m.theme.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}
	messageSection := ""
	if len(messages) > 0 {
		messageSection = lipgloss.JoinVertical(lipgloss.Left, messages...)
	}

	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Execute • Esc: Back • q: Quit")

	sections := []string{
		header,
		"",
		progInfoSection,
		"",
		"",
		m.theme.Subtitle.Render("Actions:"),
		"",
		actionsMenu,
	}

	if messageSection != "" {
		sections = append(sections, "", messageSection)
	}

	sections = append(sections, "", help)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	bordered := m.theme.BorderStyle.Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}
