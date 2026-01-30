package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// PHPFPMManagementModel represents the PHP-FPM management screen
type PHPFPMManagementModel struct {
	theme   *theme.Theme
	width   int
	height  int
	manager *system.PHPFPMManager
	pools   []system.PHPFPMPool
	cursor  int
	actions []string
	err     error
	success string
}

// NewPHPFPMManagementModel creates a new PHP-FPM management model
func NewPHPFPMManagementModel() PHPFPMManagementModel {
	manager := system.NewPHPFPMManager("")
	manager.DetectPHPVersion()
	pools, _ := manager.ListPools()
	
	actions := []string{
		"List All Pools",
		"Restart PHP-FPM Service",
		"Reload PHP-FPM Service",
		"View Service Status",
		"← Back to Configurations",
	}
	
	return PHPFPMManagementModel{
		theme:   theme.DefaultTheme(),
		manager: manager,
		pools:   pools,
		cursor:  0,
		actions: actions,
	}
}

func (m PHPFPMManagementModel) Init() tea.Cmd {
	return nil
}

func (m PHPFPMManagementModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m PHPFPMManagementModel) executeAction() (PHPFPMManagementModel, tea.Cmd) {
	m.err = nil
	m.success = ""
	
	switch m.actions[m.cursor] {
	case "List All Pools":
		pools, err := m.manager.ListPools()
		if err != nil {
			m.err = err
		} else {
			m.pools = pools
			m.success = fmt.Sprintf("✓ Found %d pools", len(pools))
		}

	case "Restart PHP-FPM Service":
		err := m.manager.RestartService()
		if err != nil {
			m.err = err
		} else {
			m.success = "✓ PHP-FPM service restarted successfully"
		}

	case "Reload PHP-FPM Service":
		err := m.manager.ReloadService()
		if err != nil {
			m.err = err
		} else {
			m.success = "✓ PHP-FPM service reloaded successfully"
		}

	case "View Service Status":
		_, err := m.manager.GetStatus()
		if err != nil {
			m.err = err
		} else {
			return m, func() tea.Msg {
				return ExecutionStartMsg{
					Command:     "systemctl status php*-fpm",
					Description: "PHP-FPM Service Status",
				}
			}
		}

	case "← Back to Configurations":
		return m, func() tea.Msg {
			return NavigateMsg{Screen: ConfigMenuScreen}
		}
	}

	return m, nil
}

func (m PHPFPMManagementModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.theme.Title.Render("🐘 PHP-FPM Pool Management")

	var poolInfo []string
	poolInfo = append(poolInfo, m.theme.Label.Render(fmt.Sprintf("Total Pools: %d", len(m.pools))))
	if len(m.pools) > 0 {
		for _, pool := range m.pools {
			poolInfo = append(poolInfo, m.theme.MenuItem.Render(fmt.Sprintf("  • %s [%s] - %s", pool.Name, pool.PM, pool.Listen)))
		}
	} else {
		poolInfo = append(poolInfo, m.theme.WarningStyle.Render("  No pools configured"))
	}
	
	poolInfoSection := lipgloss.JoinVertical(lipgloss.Left, poolInfo...)

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
		poolInfoSection,
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
	bordered := m.theme.RenderBox(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}
