package screens

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// SSLManualFormField represents the currently active field
type SSLManualFormField int

const (
	SSLFieldCertPath SSLManualFormField = iota
	SSLFieldKeyPath
	SSLFieldChainPath
	SSLFieldApply
)

// SSLManualModel represents the manual SSL certificate input screen
type SSLManualModel struct {
	theme        *theme.Theme
	width        int
	height       int
	site         system.NginxSite
	nginxManager *system.NginxManager
	currentField SSLManualFormField
	certPath     string
	keyPath      string
	chainPath    string
	err          error
	success      bool
}

// NewSSLManualModel creates a new manual SSL model
func NewSSLManualModel(site system.NginxSite) SSLManualModel {
	nginxManager := system.NewNginxManager()
	
	return SSLManualModel{
		theme:        theme.DefaultTheme(),
		nginxManager: nginxManager,
		site:         site,
		currentField: SSLFieldCertPath,
		certPath:     "",
		keyPath:      "",
		chainPath:    "", // Optional
		err:          nil,
		success:      false,
	}
}

// Init initializes the manual SSL screen
func (m SSLManualModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m SSLManualModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// If showing success/error, any key returns
		if m.success || m.err != nil {
			if msg.String() == "enter" || msg.String() == " " || msg.String() == "esc" {
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
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			return m, func() tea.Msg {
				return NavigateMsg{
					Screen: SSLOptionsScreen,
					Data: map[string]interface{}{
						"site": m.site,
					},
				}
			}

		case "tab", "down":
			m.currentField = SSLManualFormField((int(m.currentField) + 1) % 4)

		case "shift+tab", "up":
			m.currentField = SSLManualFormField((int(m.currentField) - 1 + 4) % 4)

		case "enter":
			if m.currentField == SSLFieldApply {
				return m.applySSLCertificate()
			}

		case "backspace":
			switch m.currentField {
			case SSLFieldCertPath:
				if len(m.certPath) > 0 {
					m.certPath = m.certPath[:len(m.certPath)-1]
				}
			case SSLFieldKeyPath:
				if len(m.keyPath) > 0 {
					m.keyPath = m.keyPath[:len(m.keyPath)-1]
				}
			case SSLFieldChainPath:
				if len(m.chainPath) > 0 {
					m.chainPath = m.chainPath[:len(m.chainPath)-1]
				}
			}

		default:
			// Type into current field
			if len(msg.String()) == 1 {
				switch m.currentField {
				case SSLFieldCertPath:
					m.certPath += msg.String()
				case SSLFieldKeyPath:
					m.keyPath += msg.String()
				case SSLFieldChainPath:
					m.chainPath += msg.String()
				}
			}
		}
	}

	return m, nil
}

// applySSLCertificate applies the manual SSL certificate
func (m SSLManualModel) applySSLCertificate() (SSLManualModel, tea.Cmd) {
	// Validate inputs
	if m.certPath == "" {
		m.err = fmt.Errorf("certificate path is required")
		return m, nil
	}
	if m.keyPath == "" {
		m.err = fmt.Errorf("private key path is required")
		return m, nil
	}

	// Check if files exist
	if _, err := os.Stat(m.certPath); os.IsNotExist(err) {
		m.err = fmt.Errorf("certificate file not found: %s", m.certPath)
		return m, nil
	}
	if _, err := os.Stat(m.keyPath); os.IsNotExist(err) {
		m.err = fmt.Errorf("private key file not found: %s", m.keyPath)
		return m, nil
	}
	if m.chainPath != "" {
		if _, err := os.Stat(m.chainPath); os.IsNotExist(err) {
			m.err = fmt.Errorf("chain file not found: %s", m.chainPath)
			return m, nil
		}
	}

	// Apply SSL to nginx config
	err := m.nginxManager.AddSSLManual(m.site.Name, m.certPath, m.keyPath, m.chainPath)
	if err != nil {
		m.err = err
		return m, nil
	}

	// Test configuration
	err = m.nginxManager.TestConfig()
	if err != nil {
		m.err = fmt.Errorf("SSL applied but config test failed: %w", err)
		return m, nil
	}

	// Reload nginx
	err = m.nginxManager.ReloadNginx()
	if err != nil {
		m.err = fmt.Errorf("SSL applied but reload failed: %w", err)
		return m, nil
	}

	m.success = true
	m.err = nil
	return m, nil
}

// View renders the manual SSL screen
func (m SSLManualModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// If success or error, show message
	if m.success {
		msg := m.theme.SuccessStyle.Render("✓ SSL certificate applied successfully!")
		help := m.theme.Help.Render("Press any key to continue...")
		content := lipgloss.JoinVertical(lipgloss.Center, "", msg, "", help)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	if m.err != nil {
		msg := m.theme.ErrorStyle.Render(fmt.Sprintf("✗ Error: %v", m.err))
		help := m.theme.Help.Render("Press any key to continue...")
		content := lipgloss.JoinVertical(lipgloss.Center, "", msg, "", help)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	// Header
	header := m.theme.Title.Render("Manual SSL Certificate")

	// Site info
	siteInfo := m.theme.DescriptionStyle.Render(fmt.Sprintf("Site: %s (%s)", m.site.Name, m.site.Domain))

	// Form fields
	var formFields []string

	// Certificate Path
	certStyle := m.theme.MenuItem
	if m.currentField == SSLFieldCertPath {
		certStyle = m.theme.SelectedItem
	}
	formFields = append(formFields, certStyle.Render(fmt.Sprintf("Certificate:   %s_", m.certPath)))

	// Private Key Path
	keyStyle := m.theme.MenuItem
	if m.currentField == SSLFieldKeyPath {
		keyStyle = m.theme.SelectedItem
	}
	formFields = append(formFields, keyStyle.Render(fmt.Sprintf("Private Key:   %s_", m.keyPath)))

	// Chain Path (Optional)
	chainStyle := m.theme.MenuItem
	if m.currentField == SSLFieldChainPath {
		chainStyle = m.theme.SelectedItem
	}
	formFields = append(formFields, chainStyle.Render(fmt.Sprintf("Chain (opt):   %s_", m.chainPath)))

	// Apply button
	applyStyle := m.theme.MenuItem
	if m.currentField == SSLFieldApply {
		applyStyle = m.theme.SelectedItem
	}
	formFields = append(formFields, "")
	formFields = append(formFields, applyStyle.Render("[ Apply SSL Certificate ]"))

	form := lipgloss.JoinVertical(lipgloss.Left, formFields...)

	// Help text
	help := m.theme.Help.Render("Tab/↑↓: Navigate • Enter: Apply • Esc: Back • q: Quit")

	// Instructions
	instructions := m.theme.DescriptionStyle.Render("Enter full paths to certificate files (e.g., /etc/ssl/certs/mydomain.crt)")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		siteInfo,
		"",
		instructions,
		"",
		form,
		"",
		help,
	)

	// Add border and center
	bordered := m.theme.BorderStyle.Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}
