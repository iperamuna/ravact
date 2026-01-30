package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// FirewallManagementModel represents the firewall management screen
type FirewallManagementModel struct {
	theme           *theme.Theme
	width           int
	height          int
	firewallManager *system.FirewallManager
	cursor          int
	actions         []string
	rules           []system.FirewallRule
	status          string
	err             error
	success         string
	inputMode       bool
	inputField      string
	inputValue      string
	inputPrompt     string
}

// NewFirewallManagementModel creates a new firewall management model
func NewFirewallManagementModel() FirewallManagementModel {
	firewallManager := system.NewFirewallManager()
	status, _ := firewallManager.GetStatus()
	rules, _ := firewallManager.GetRules()

	actions := []string{
		"View Current Rules",
		"Allow Port",
		"Deny Port",
		"Delete Rule",
		"Enable Firewall",
		"Disable Firewall",
		"Reload Firewall",
		"← Back to Configurations",
	}

	return FirewallManagementModel{
		theme:           theme.DefaultTheme(),
		firewallManager: firewallManager,
		cursor:          0,
		actions:         actions,
		rules:           rules,
		status:          status,
	}
}

// Init initializes the firewall management screen
func (m FirewallManagementModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m FirewallManagementModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle input mode
		if m.inputMode {
			switch msg.String() {
			case "enter":
				return m.processInput()
			case "esc":
				m.inputMode = false
				m.inputValue = ""
				m.inputField = ""
				m.inputPrompt = ""
				return m, nil
			case "backspace":
				if len(m.inputValue) > 0 {
					m.inputValue = m.inputValue[:len(m.inputValue)-1]
				}
			default:
				// Add character to input (filter valid chars for port)
				char := msg.String()
				if len(char) == 1 && (char[0] >= '0' && char[0] <= '9' || char[0] == '/' || char[0] >= 'a' && char[0] <= 'z') {
					m.inputValue += char
				}
			}
			return m, nil
		}

		// Normal mode
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

// processInput processes the user input
func (m FirewallManagementModel) processInput() (FirewallManagementModel, tea.Cmd) {
	m.err = nil
	m.success = ""

	port := m.inputValue
	protocol := "tcp"

	// Parse port/protocol format (e.g., "8080/tcp" or just "8080")
	if strings.Contains(port, "/") {
		parts := strings.Split(port, "/")
		port = parts[0]
		if len(parts) > 1 {
			protocol = parts[1]
		}
	}

	if port == "" {
		m.err = fmt.Errorf("port cannot be empty")
		m.inputMode = false
		m.inputValue = ""
		return m, nil
	}

	switch m.inputField {
	case "allow":
		if err := m.firewallManager.AllowPort(port, protocol); err != nil {
			m.err = err
		} else {
			m.success = fmt.Sprintf("✓ Port %s/%s allowed", port, protocol)
			m.rules, _ = m.firewallManager.GetRules()
		}

	case "deny":
		if err := m.firewallManager.DenyPort(port, protocol); err != nil {
			m.err = err
		} else {
			m.success = fmt.Sprintf("✓ Port %s/%s denied", port, protocol)
			m.rules, _ = m.firewallManager.GetRules()
		}

	case "delete":
		if err := m.firewallManager.DeleteRule(port, protocol); err != nil {
			m.err = err
		} else {
			m.success = fmt.Sprintf("✓ Rule for port %s/%s deleted", port, protocol)
			m.rules, _ = m.firewallManager.GetRules()
		}
	}

	m.inputMode = false
	m.inputValue = ""
	m.inputField = ""
	m.inputPrompt = ""

	return m, nil
}

// executeAction executes the selected action
func (m FirewallManagementModel) executeAction() (FirewallManagementModel, tea.Cmd) {
	m.err = nil
	m.success = ""

	actionName := m.actions[m.cursor]

	switch actionName {
	case "View Current Rules":
		// Refresh rules
		rules, err := m.firewallManager.GetRules()
		if err != nil {
			m.err = err
		} else {
			m.rules = rules
			m.success = fmt.Sprintf("✓ Found %d rules", len(rules))
		}

	case "Allow Port":
		m.inputMode = true
		m.inputField = "allow"
		m.inputPrompt = "Enter port to allow (e.g., 8080 or 8080/tcp):"
		m.inputValue = ""

	case "Deny Port":
		m.inputMode = true
		m.inputField = "deny"
		m.inputPrompt = "Enter port to deny (e.g., 8080 or 8080/tcp):"
		m.inputValue = ""

	case "Delete Rule":
		m.inputMode = true
		m.inputField = "delete"
		m.inputPrompt = "Enter port to delete rule for (e.g., 8080 or 8080/tcp):"
		m.inputValue = ""

	case "Enable Firewall":
		if err := m.firewallManager.EnableFirewall(); err != nil {
			m.err = err
		} else {
			m.success = "✓ Firewall enabled"
			m.status, _ = m.firewallManager.GetStatus()
		}

	case "Disable Firewall":
		if err := m.firewallManager.DisableFirewall(); err != nil {
			m.err = err
		} else {
			m.success = "✓ Firewall disabled"
			m.status, _ = m.firewallManager.GetStatus()
		}

	case "Reload Firewall":
		if err := m.firewallManager.ReloadFirewall(); err != nil {
			m.err = err
		} else {
			m.success = "✓ Firewall reloaded"
			m.rules, _ = m.firewallManager.GetRules()
		}

	case "← Back to Configurations":
		return m, func() tea.Msg {
			return NavigateMsg{Screen: ConfigMenuScreen}
		}
	}

	return m, nil
}

// View renders the firewall management screen
func (m FirewallManagementModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	firewallType := string(m.firewallManager.GetFirewallType())
	header := m.theme.Title.Render(fmt.Sprintf("Firewall Configuration (%s)", strings.ToUpper(firewallType)))

	// Status
	statusStyle := m.theme.DescriptionStyle
	statusText := m.status
	if m.status == "active" {
		statusStyle = m.theme.SuccessStyle
		statusText = "Active (Enabled)"
	} else if m.status == "inactive" {
		statusStyle = m.theme.WarningStyle
		statusText = "Inactive (Disabled)"
	}
	statusLine := m.theme.Label.Render("Status: ") + statusStyle.Render(statusText)

	// Current rules summary
	var rulesInfo []string
	rulesInfo = append(rulesInfo, m.theme.Label.Render(fmt.Sprintf("Current Rules (%d):", len(m.rules))))
	
	if len(m.rules) == 0 {
		rulesInfo = append(rulesInfo, m.theme.DescriptionStyle.Render("  No rules configured"))
	} else {
		// Show up to 5 rules
		maxRules := 5
		if len(m.rules) < maxRules {
			maxRules = len(m.rules)
		}
		for i := 0; i < maxRules; i++ {
			rule := m.rules[i]
			ruleText := fmt.Sprintf("  • %s/%s %s from %s", rule.Port, rule.Protocol, strings.ToUpper(rule.Action), rule.From)
			if rule.Action == "allow" {
				rulesInfo = append(rulesInfo, m.theme.SuccessStyle.Render(ruleText))
			} else {
				rulesInfo = append(rulesInfo, m.theme.ErrorStyle.Render(ruleText))
			}
		}
		if len(m.rules) > maxRules {
			rulesInfo = append(rulesInfo, m.theme.DescriptionStyle.Render(fmt.Sprintf("  ... and %d more rules", len(m.rules)-maxRules)))
		}
	}
	rulesSection := lipgloss.JoinVertical(lipgloss.Left, rulesInfo...)

	// Input mode display
	var inputSection string
	if m.inputMode {
		inputSection = lipgloss.JoinVertical(lipgloss.Left,
			"",
			m.theme.Label.Render(m.inputPrompt),
			m.theme.SelectedItem.Render(fmt.Sprintf("> %s_", m.inputValue)),
			m.theme.DescriptionStyle.Render("Press Enter to confirm, Esc to cancel"),
		)
	}

	// Actions menu
	var actionItems []string
	actionItems = append(actionItems, m.theme.Subtitle.Render("Actions:"))
	actionItems = append(actionItems, "")
	
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

	// Messages
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

	// Help
	var help string
	if m.inputMode {
		help = m.theme.Help.Render("Enter: Confirm • Esc: Cancel")
	} else {
		help = m.theme.Help.Render("↑/↓: Navigate • Enter: Execute • Esc: Back • q: Quit")
	}

	// Warning
	warning := m.theme.WarningStyle.Render("⚠ Be careful! Incorrect firewall rules may lock you out of SSH.")

	// Combine all sections
	sections := []string{
		header,
		"",
		statusLine,
		"",
		rulesSection,
	}

	if inputSection != "" {
		sections = append(sections, inputSection)
	}

	sections = append(sections, "", actionsMenu)

	if messageSection != "" {
		sections = append(sections, "", messageSection)
	}

	sections = append(sections, "", warning, "", help)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

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
