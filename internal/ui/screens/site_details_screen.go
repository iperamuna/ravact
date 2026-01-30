package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// SiteAction represents an action on a site
type SiteAction int

const (
	ActionToggleEnable SiteAction = iota
	ActionAddSSL
	ActionRemoveSSL
	ActionTestConfig
	ActionReloadNginx
	ActionDeleteSite
	ActionOpenEditor
	ActionBack
)

// SiteDetailsModel represents the site details screen
type SiteDetailsModel struct {
	theme        *theme.Theme
	width        int
	height       int
	nginxManager *system.NginxManager
	site         system.NginxSite
	cursor       int
	actions      []string
	err          error
	success      string
}

// NewSiteDetailsModel creates a new site details model
func NewSiteDetailsModel(site system.NginxSite) SiteDetailsModel {
	nginxManager := system.NewNginxManager()

	// Build actions list based on site's SSL status
	actions := []string{
		"Toggle Enable/Disable",
	}

	if !site.HasSSL {
		actions = append(actions, "Add SSL Certificate (Let's Encrypt)")
	} else {
		actions = append(actions, "Remove SSL Certificate")
	}

	if site.HasPHP {
		actions = append(actions, "Convert to FrankenPHP Classic Mode")
	}

	actions = append(actions,
		"Test Nginx Configuration",
		"Reload Nginx",
		"Delete Site",
		"Open in Editor",
		"← Back to Sites",
	)

	return SiteDetailsModel{
		theme:        theme.DefaultTheme(),
		nginxManager: nginxManager,
		site:         site,
		cursor:       0,
		actions:      actions,
		err:          nil,
		success:      "",
	}
}

// Init initializes the site details screen
func (m SiteDetailsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m SiteDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				return NavigateMsg{Screen: NginxConfigScreen}
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

// executeAction executes the selected action
func (m SiteDetailsModel) executeAction() (SiteDetailsModel, tea.Cmd) {
	m.err = nil
	m.success = ""

	// Map cursor position to action based on current actions list
	actionName := m.actions[m.cursor]

	switch {
	case actionName == "Toggle Enable/Disable":
		var err error
		if m.site.IsEnabled {
			err = m.nginxManager.DisableSite(m.site.Name)
			if err == nil {
				m.success = fmt.Sprintf("✓ Site '%s' disabled", m.site.Name)
				m.site.IsEnabled = false
			}
		} else {
			err = m.nginxManager.EnableSite(m.site.Name)
			if err == nil {
				m.success = fmt.Sprintf("✓ Site '%s' enabled", m.site.Name)
				m.site.IsEnabled = true
			}
		}
		if err != nil {
			m.err = err
		}

	case actionName == "Add SSL Certificate (Let's Encrypt)":
		// Navigate to SSL options screen
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: SSLOptionsScreen,
				Data: map[string]interface{}{
					"site": m.site,
				},
			}
		}

	case actionName == "Remove SSL Certificate":
		// Remove SSL configuration
		err := m.nginxManager.RemoveSSL(m.site.Name)
		if err != nil {
			m.err = fmt.Errorf("failed to remove SSL: %w", err)
		} else {
			// Test configuration
			err = m.nginxManager.TestConfig()
			if err != nil {
				m.err = fmt.Errorf("SSL removed but config test failed: %w", err)
			} else {
				// Reload nginx
				err = m.nginxManager.ReloadNginx()
				if err != nil {
					m.err = fmt.Errorf("SSL removed but reload failed: %w", err)
				} else {
					m.success = "✓ SSL certificate removed, site now uses HTTP only"
					m.site.HasSSL = false
					// Return to nginx config to refresh
					return m, func() tea.Msg {
						return NavigateMsg{Screen: NginxConfigScreen}
					}
				}
			}
		}

	case actionName == "Test Nginx Configuration":
		err := m.nginxManager.TestConfig()
		if err != nil {
			m.err = err
		} else {
			m.success = "✓ Nginx configuration is valid"
		}

	case actionName == "Reload Nginx":
		err := m.nginxManager.ReloadNginx()
		if err != nil {
			m.err = err
		} else {
			m.success = "✓ Nginx reloaded successfully"
		}

	case actionName == "Delete Site":
		// Confirm deletion
		err := m.nginxManager.DeleteSite(m.site.Name)
		if err != nil {
			m.err = err
		} else {
			m.success = fmt.Sprintf("✓ Site '%s' deleted", m.site.Name)
			// Return to nginx config screen
			return m, func() tea.Msg {
				return NavigateMsg{Screen: NginxConfigScreen}
			}
		}

	case actionName == "Convert to FrankenPHP Classic Mode":
		// Navigate to FrankenPHP classic screen with site data
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: FrankenPHPClassicScreen,
				Data: map[string]interface{}{
					"site": m.site,
				},
			}
		}

	case actionName == "Open in Editor":
		// Navigate to editor selection screen
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: EditorSelectionScreen,
				Data: map[string]interface{}{
					"site": m.site,
				},
			}
		}

	case actionName == "← Back to Sites":
		return m, func() tea.Msg {
			return NavigateMsg{Screen: NginxConfigScreen}
		}
	}

	return m, nil
}

// View renders the site details screen
func (m SiteDetailsModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	header := m.theme.Title.Render(fmt.Sprintf("Site Details: %s", m.site.Name))

	// Site information
	var info []string
	info = append(info, m.theme.Label.Render("Domain:      ")+m.theme.MenuItem.Render(m.site.Domain))
	info = append(info, m.theme.Label.Render("Root Dir:    ")+m.theme.MenuItem.Render(m.site.RootDir))
	info = append(info, m.theme.Label.Render("Config Path: ")+m.theme.DescriptionStyle.Render(m.site.ConfigPath))

	// Status
	statusText := ""
	statusStyle := m.theme.DescriptionStyle
	if m.site.IsEnabled {
		statusText = "Enabled"
		statusStyle = m.theme.SuccessStyle
	} else {
		statusText = "Disabled"
		statusStyle = m.theme.WarningStyle
	}
	info = append(info, m.theme.Label.Render("Status:      ")+statusStyle.Render(statusText))

	// SSL
	sslText := "No"
	if m.site.HasSSL {
		sslText = "Yes"
	}
	info = append(info, m.theme.Label.Render("SSL:         ")+m.theme.MenuItem.Render(sslText))

	siteInfo := lipgloss.JoinVertical(lipgloss.Left, info...)

	// Actions menu
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
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Execute • Esc: Back • q: Quit")

	// Warning for delete
	warning := ""
	if m.cursor == int(ActionDeleteSite) {
		warning = m.theme.ErrorStyle.Render("⚠ Warning: This will permanently delete the site configuration!")
	}

	// Combine all sections
	sections := []string{
		header,
		"",
		siteInfo,
		"",
		"",
		m.theme.Subtitle.Render("Actions:"),
		"",
		actionsMenu,
	}

	if messageSection != "" {
		sections = append(sections, "", messageSection)
	}

	if warning != "" {
		sections = append(sections, "", warning)
	}

	sections = append(sections, "", help)

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
