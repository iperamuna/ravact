package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// SSLOptionsModel represents the SSL configuration options screen
type SSLOptionsModel struct {
	theme        *theme.Theme
	width        int
	height       int
	site         system.NginxSite
	cursor       int
	options      []string
}

// NewSSLOptionsModel creates a new SSL options model
func NewSSLOptionsModel(site system.NginxSite) SSLOptionsModel {
	options := []string{
		"Let's Encrypt (Automatic)",
		"Manual Certificate (Provide paths)",
		"← Cancel",
	}
	
	return SSLOptionsModel{
		theme:   theme.DefaultTheme(),
		site:    site,
		cursor:  0,
		options: options,
	}
}

// Init initializes the SSL options screen
func (m SSLOptionsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m SSLOptionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				return NavigateMsg{
					Screen: ConfigEditorScreen,
					Data: map[string]interface{}{
						"action": "edit_nginx_site",
						"site":   m.site,
					},
				}
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}

		case "enter", " ":
			return m.executeOption()
		}
	}

	return m, nil
}

// executeOption executes the selected option
func (m SSLOptionsModel) executeOption() (SSLOptionsModel, tea.Cmd) {
	option := m.options[m.cursor]

	switch option {
	case "Let's Encrypt (Automatic)":
		// Navigate to execution screen to run certbot
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     fmt.Sprintf("certbot --nginx -d %s", m.site.Domain),
				Description: fmt.Sprintf("Installing SSL certificate for %s", m.site.Domain),
			}
		}

	case "Manual Certificate (Provide paths)":
		// Navigate to manual SSL certificate screen
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: SSLManualScreen,
				Data: map[string]interface{}{
					"site": m.site,
				},
			}
		}

	case "← Cancel":
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: ConfigEditorScreen,
				Data: map[string]interface{}{
					"action": "edit_nginx_site",
					"site":   m.site,
				},
			}
		}
	}

	return m, nil
}

// View renders the SSL options screen
func (m SSLOptionsModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	header := m.theme.Title.Render("Add SSL Certificate")

	// Site info
	siteInfo := m.theme.DescriptionStyle.Render(fmt.Sprintf("Site: %s (%s)", m.site.Name, m.site.Domain))

	// Instructions
	instructions := lipgloss.JoinVertical(
		lipgloss.Left,
		m.theme.Label.Render("Choose SSL certificate method:"),
		"",
		m.theme.DescriptionStyle.Render("Let's Encrypt: Free, automatic certificates"),
		m.theme.DescriptionStyle.Render("  • Requires domain to point to this server"),
		m.theme.DescriptionStyle.Render("  • Ports 80 & 443 must be accessible"),
		m.theme.DescriptionStyle.Render("  • Email required for renewal notifications"),
		"",
		m.theme.DescriptionStyle.Render("Manual: Use your own certificate files"),
		m.theme.DescriptionStyle.Render("  • Requires certificate and private key files"),
		m.theme.DescriptionStyle.Render("  • You manage renewals"),
	)

	// Options menu
	var optionItems []string
	for i, option := range m.options {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		var renderedItem string
		if i == m.cursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, option))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, option))
		}

		optionItems = append(optionItems, renderedItem)
	}

	optionsMenu := lipgloss.JoinVertical(lipgloss.Left, optionItems...)

	// Help
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Select • Esc: Back • q: Quit")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		siteInfo,
		"",
		instructions,
		"",
		"",
		optionsMenu,
		"",
		help,
	)

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
