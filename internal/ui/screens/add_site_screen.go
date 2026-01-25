package screens

import (
	"embed"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// AddSiteModel represents the add site screen
type AddSiteModel struct {
	theme        *theme.Theme
	width        int
	height       int
	nginxManager *system.NginxManager
	templates    []system.NginxTemplate

	// Form
	form *huh.Form

	// Form fields
	siteName         string
	domain           string
	rootDir          string
	selectedTemplate string
	sslOption        string
	email            string

	// State
	err     error
	success bool
}

// NewAddSiteModel creates a new add site model
func NewAddSiteModel() AddSiteModel {
	nginxManager := system.NewNginxManager()

	// Set embedded FS if available
	if EmbeddedFS != (embed.FS{}) {
		nginxManager.SetEmbeddedFS(&EmbeddedFS)
	}

	templates := nginxManager.GetTemplates()
	t := theme.DefaultTheme()

	m := AddSiteModel{
		theme:            t,
		nginxManager:     nginxManager,
		templates:        templates,
		siteName:         "",
		domain:           "",
		rootDir:          "/var/www/html",
		selectedTemplate: "static",
		sslOption:        "none",
		email:            "",
		err:              nil,
		success:          false,
	}

	// Build template options
	templateOptions := []huh.Option[string]{}
	for _, tpl := range templates {
		templateOptions = append(templateOptions, huh.NewOption(tpl.Name, tpl.ID))
	}
	if len(templateOptions) == 0 {
		templateOptions = append(templateOptions, huh.NewOption("Static HTML", "static"))
	}

	// Create the huh form
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Site Name").
				Description("Unique identifier for the site configuration").
				Placeholder("mysite").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("site name is required")
					}
					if strings.Contains(s, " ") {
						return fmt.Errorf("site name cannot contain spaces")
					}
					return nil
				}).
				Value(&m.siteName),

			huh.NewInput().
				Title("Domain").
				Description("Domain name for the site (e.g., example.com)").
				Placeholder("example.com").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("domain is required")
					}
					return nil
				}).
				Value(&m.domain),

			huh.NewInput().
				Title("Root Directory").
				Description("Document root path for web files").
				Placeholder("/var/www/html").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("root directory is required")
					}
					if !strings.HasPrefix(s, "/") {
						return fmt.Errorf("must be an absolute path")
					}
					return nil
				}).
				Value(&m.rootDir),

			huh.NewSelect[string]().
				Title("Template").
				Description("Nginx configuration template").
				Options(templateOptions...).
				Value(&m.selectedTemplate),

			huh.NewSelect[string]().
				Title("SSL Certificate").
				Description("SSL/HTTPS configuration").
				Options(
					huh.NewOption("None (HTTP only)", "none"),
					huh.NewOption("Let's Encrypt (Free SSL)", "letsencrypt"),
				).
				Value(&m.sslOption),

			huh.NewInput().
				Title("Email (for Let's Encrypt)").
				Description("Only required if using Let's Encrypt SSL").
				Placeholder("admin@example.com").
				Value(&m.email),
		),
	).WithTheme(t.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)

	return m
}

// Init initializes the add site screen
func (m AddSiteModel) Init() tea.Cmd {
	return m.form.Init()
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
			return m, nil
		}

		// Global keys
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.form.State == huh.StateNormal {
				return m, func() tea.Msg {
					return NavigateMsg{Screen: NginxConfigScreen}
				}
			}
		}
	}

	// Update the form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	// Check if form is completed
	if m.form.State == huh.StateCompleted {
		return m.createSite()
	}

	return m, cmd
}

// createSite creates the nginx site configuration
func (m AddSiteModel) createSite() (AddSiteModel, tea.Cmd) {
	// Validate email for Let's Encrypt
	if m.sslOption == "letsencrypt" && m.email == "" {
		m.err = fmt.Errorf("email is required for Let's Encrypt")
		return m, nil
	}

	// Determine SSL settings
	useSSL := m.sslOption != "none"
	useCertbot := m.sslOption == "letsencrypt"

	// Create the site
	err := m.nginxManager.CreateSite(m.siteName, m.domain, m.rootDir, m.selectedTemplate, useSSL, useCertbot)
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

	// If success, show message
	if m.success {
		msg := m.theme.SuccessStyle.Render(m.theme.Symbols.CheckMark + " Site created successfully!")
		help := m.theme.Help.Render("Press any key to continue...")
		content := lipgloss.JoinVertical(lipgloss.Center, "", msg, "", help)
		bordered := m.theme.BorderStyle.Render(content)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
	}

	// If error, show message
	if m.err != nil {
		msg := m.theme.ErrorStyle.Render(m.theme.Symbols.CrossMark + " Error: " + m.err.Error())
		help := m.theme.Help.Render("Press any key to continue...")
		content := lipgloss.JoinVertical(lipgloss.Center, "", msg, "", help)
		bordered := m.theme.BorderStyle.Render(content)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
	}

	// Header
	header := m.theme.Title.Render("Add Nginx Site")

	// Render the huh form
	formView := m.form.View()

	// Help text
	help := m.theme.Help.Render("Tab/Shift+Tab: Navigate " + m.theme.Symbols.Bullet + " Enter: Select/Submit " + m.theme.Symbols.Bullet + " Esc: Cancel")

	// Template description
	templateDesc := ""
	for _, tpl := range m.templates {
		if tpl.ID == m.selectedTemplate {
			templateDesc = m.theme.DescriptionStyle.Render("Template: " + tpl.Description)
			if len(tpl.RecommendedFor) > 0 {
				templateDesc += "\n" + m.theme.DescriptionStyle.Render("Recommended for: " + strings.Join(tpl.RecommendedFor, ", "))
			}
			break
		}
	}

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		formView,
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
