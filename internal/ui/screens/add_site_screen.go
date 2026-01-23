package screens

import (
	"embed"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// FormField represents the currently active field
type FormField int

const (
	FieldSiteName FormField = iota
	FieldDomain
	FieldRootDir
	FieldTemplate
	FieldSSLOption
	FieldEmail
	FieldSubmit
)

// SSLOption represents SSL configuration choice
type SSLOption int

const (
	SSLNone SSLOption = iota
	SSLLetsEncrypt
)

// AddSiteModel represents the add site screen
type AddSiteModel struct {
	theme          *theme.Theme
	width          int
	height         int
	nginxManager   *system.NginxManager
	templates      []system.NginxTemplate
	currentField   FormField
	siteName       string
	domain         string
	rootDir        string
	selectedTemplate int
	sslOption      SSLOption
	email          string
	err            error
	success        bool
}

// NewAddSiteModel creates a new add site model
func NewAddSiteModel() AddSiteModel {
	nginxManager := system.NewNginxManager()
	
	// Set embedded FS if available
	if EmbeddedFS != (embed.FS{}) {
		nginxManager.SetEmbeddedFS(&EmbeddedFS)
	}
	
	templates := nginxManager.GetTemplates()
	
	return AddSiteModel{
		theme:          theme.DefaultTheme(),
		nginxManager:   nginxManager,
		templates:      templates,
		currentField:   FieldSiteName,
		siteName:       "",
		domain:         "",
		rootDir:        "/var/www/html",
		selectedTemplate: 0,
		sslOption:      SSLNone,
		email:          "",
		err:            nil,
		success:        false,
	}
}

// Init initializes the add site screen
func (m AddSiteModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m AddSiteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					return NavigateMsg{Screen: NginxConfigScreen}
				}
			}
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: NginxConfigScreen}
			}

		case "tab", "down":
			m.currentField = FormField((int(m.currentField) + 1) % 7)

		case "shift+tab", "up":
			m.currentField = FormField((int(m.currentField) - 1 + 7) % 7)

		case "enter":
			if m.currentField == FieldSubmit {
				return m.createSite()
			}

		case "left":
			switch m.currentField {
			case FieldTemplate:
				if m.selectedTemplate > 0 {
					m.selectedTemplate--
				}
			case FieldSSLOption:
				if m.sslOption > 0 {
					m.sslOption--
				}
			}

		case "right":
			switch m.currentField {
			case FieldTemplate:
				if m.selectedTemplate < len(m.templates)-1 {
					m.selectedTemplate++
				}
			case FieldSSLOption:
				if m.sslOption < SSLLetsEncrypt {
					m.sslOption++
				}
			}

		case "backspace":
			switch m.currentField {
			case FieldSiteName:
				if len(m.siteName) > 0 {
					m.siteName = m.siteName[:len(m.siteName)-1]
				}
			case FieldDomain:
				if len(m.domain) > 0 {
					m.domain = m.domain[:len(m.domain)-1]
				}
			case FieldRootDir:
				if len(m.rootDir) > 0 {
					m.rootDir = m.rootDir[:len(m.rootDir)-1]
				}
			case FieldEmail:
				if len(m.email) > 0 {
					m.email = m.email[:len(m.email)-1]
				}
			}

		default:
			// Type into current field
			if len(msg.String()) == 1 {
				switch m.currentField {
				case FieldSiteName:
					m.siteName += msg.String()
				case FieldDomain:
					m.domain += msg.String()
				case FieldRootDir:
					m.rootDir += msg.String()
				case FieldEmail:
					m.email += msg.String()
				}
			}
		}
	}

	return m, nil
}

// createSite creates the nginx site configuration
func (m AddSiteModel) createSite() (AddSiteModel, tea.Cmd) {
	// Validate inputs
	if m.siteName == "" {
		m.err = fmt.Errorf("site name is required")
		return m, nil
	}
	if m.domain == "" {
		m.err = fmt.Errorf("domain is required")
		return m, nil
	}
	if m.rootDir == "" {
		m.err = fmt.Errorf("root directory is required")
		return m, nil
	}
	if m.sslOption == SSLLetsEncrypt && m.email == "" {
		m.err = fmt.Errorf("email is required for Let's Encrypt")
		return m, nil
	}

	// Get template ID
	templateID := "static"
	if len(m.templates) > 0 && m.selectedTemplate < len(m.templates) {
		templateID = m.templates[m.selectedTemplate].ID
	}

	// Determine SSL settings
	useSSL := m.sslOption != SSLNone
	useCertbot := m.sslOption == SSLLetsEncrypt

	// Create the site
	err := m.nginxManager.CreateSite(m.siteName, m.domain, m.rootDir, templateID, useSSL, useCertbot)
	if err != nil {
		m.err = err
		return m, nil
	}

	// Enable the site
	err = m.nginxManager.EnableSite(m.siteName)
	if err != nil {
		m.err = fmt.Errorf("site created but failed to enable: %w", err)
		return m, nil
	}

	// Test configuration
	err = m.nginxManager.TestConfig()
	if err != nil {
		m.err = fmt.Errorf("site created but config test failed: %w", err)
		return m, nil
	}

	// Reload nginx
	err = m.nginxManager.ReloadNginx()
	if err != nil {
		m.err = fmt.Errorf("site created but reload failed: %w", err)
		return m, nil
	}

	// If using certbot, obtain certificate
	if useCertbot {
		err = m.nginxManager.ObtainSSLCertificate(m.domain)
		if err != nil {
			m.err = fmt.Errorf("site created but certbot failed: %w", err)
			return m, nil
		}
	}

	m.success = true
	m.err = nil
	return m, nil
}

// View renders the add site screen
func (m AddSiteModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// If success or error, show message
	if m.success {
		msg := m.theme.SuccessStyle.Render("✓ Site created successfully!")
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
	header := m.theme.Title.Render("Add Nginx Site")

	// Form fields
	var formFields []string

	// Site Name
	siteNameStyle := m.theme.MenuItem
	if m.currentField == FieldSiteName {
		siteNameStyle = m.theme.SelectedItem
	}
	formFields = append(formFields, siteNameStyle.Render(fmt.Sprintf("Site Name:     %s_", m.siteName)))

	// Domain
	domainStyle := m.theme.MenuItem
	if m.currentField == FieldDomain {
		domainStyle = m.theme.SelectedItem
	}
	formFields = append(formFields, domainStyle.Render(fmt.Sprintf("Domain:        %s_", m.domain)))

	// Root Directory
	rootDirStyle := m.theme.MenuItem
	if m.currentField == FieldRootDir {
		rootDirStyle = m.theme.SelectedItem
	}
	formFields = append(formFields, rootDirStyle.Render(fmt.Sprintf("Root Dir:      %s_", m.rootDir)))

	// Template Selection
	templateStyle := m.theme.MenuItem
	if m.currentField == FieldTemplate {
		templateStyle = m.theme.SelectedItem
	}
	templateName := "static"
	if len(m.templates) > 0 && m.selectedTemplate < len(m.templates) {
		templateName = m.templates[m.selectedTemplate].Name
	}
	formFields = append(formFields, templateStyle.Render(fmt.Sprintf("Template:      ◀ %s ▶", templateName)))

	// SSL Option
	sslStyle := m.theme.MenuItem
	if m.currentField == FieldSSLOption {
		sslStyle = m.theme.SelectedItem
	}
	sslOptions := []string{"None", "Let's Encrypt"}
	formFields = append(formFields, sslStyle.Render(fmt.Sprintf("SSL:           ◀ %s ▶", sslOptions[m.sslOption])))

	// Email (only show if Let's Encrypt selected)
	if m.sslOption == SSLLetsEncrypt {
		emailStyle := m.theme.MenuItem
		if m.currentField == FieldEmail {
			emailStyle = m.theme.SelectedItem
		}
		formFields = append(formFields, emailStyle.Render(fmt.Sprintf("Email:         %s_", m.email)))
	}

	// Submit button
	submitStyle := m.theme.MenuItem
	if m.currentField == FieldSubmit {
		submitStyle = m.theme.SelectedItem
	}
	formFields = append(formFields, "")
	formFields = append(formFields, submitStyle.Render("[ Create Site ]"))

	form := lipgloss.JoinVertical(lipgloss.Left, formFields...)

	// Help text
	help := m.theme.Help.Render("Tab/↑↓: Navigate • ←→: Change option • Enter: Submit • Esc: Cancel")

	// Template description
	templateDesc := ""
	if len(m.templates) > 0 && m.selectedTemplate < len(m.templates) {
		tpl := m.templates[m.selectedTemplate]
		templateDesc = m.theme.DescriptionStyle.Render(fmt.Sprintf("  %s", tpl.Description))
		if len(tpl.RecommendedFor) > 0 {
			templateDesc += "\n" + m.theme.DescriptionStyle.Render(fmt.Sprintf("  For: %s", strings.Join(tpl.RecommendedFor, ", ")))
		}
	}

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		form,
		"",
		templateDesc,
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
