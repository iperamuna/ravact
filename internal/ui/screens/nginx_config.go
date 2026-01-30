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

// NginxViewMode represents which view is active
type NginxViewMode int

const (
	SitesListView NginxViewMode = iota
	GlobalConfigView
)

// NginxConfigModel represents the Nginx configuration screen
type NginxConfigModel struct {
	theme        *theme.Theme
	width        int
	height       int
	nginxManager *system.NginxManager
	sites        []system.NginxSite
	cursor       int
	viewMode     NginxViewMode
	scrollOffset int
	maxVisible   int
	err          error
}

// NewNginxConfigModel creates a new Nginx config model
func NewNginxConfigModel() NginxConfigModel {
	nginxManager := system.NewNginxManager()
	
	// Set embedded FS if available
	if EmbeddedFS != (embed.FS{}) {
		nginxManager.SetEmbeddedFS(&EmbeddedFS)
	}
	
	sites, _ := nginxManager.GetAllSites()

	return NginxConfigModel{
		theme:        theme.DefaultTheme(),
		nginxManager: nginxManager,
		sites:        sites,
		cursor:       0,
		viewMode:     SitesListView,
		scrollOffset: 0,
		maxVisible:   10,
	}
}

// Init initializes the Nginx config screen
func (m NginxConfigModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for Nginx config
func (m NginxConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				return NavigateMsg{Screen: MainMenuScreen}
			}

		case "tab":
			// Switch between sites and global config
			if m.viewMode == SitesListView {
				m.viewMode = GlobalConfigView
			} else {
				m.viewMode = SitesListView
				m.cursor = 0
			}

		case "up", "k":
			if m.viewMode == SitesListView && m.cursor > 0 {
				m.cursor--
				if m.cursor < m.scrollOffset {
					m.scrollOffset = m.cursor
				}
			}

		case "down", "j":
			if m.viewMode == SitesListView && m.cursor < len(m.sites)-1 {
				m.cursor++
				if m.cursor >= m.scrollOffset+m.maxVisible {
					m.scrollOffset = m.cursor - m.maxVisible + 1
				}
			}

		case "r":
			// Refresh sites list
			m.sites, _ = m.nginxManager.GetAllSites()
			m.cursor = 0
			m.scrollOffset = 0

		case "a":
			// Add new site
			if m.viewMode == SitesListView {
				return m, func() tea.Msg {
					return NavigateMsg{
						Screen: ConfigEditorScreen,
						Data: map[string]interface{}{
							"action": "add_nginx_site",
						},
					}
				}
			}

		case "e":
			// Enable/Disable site
			if m.viewMode == SitesListView && len(m.sites) > 0 {
				site := m.sites[m.cursor]
				var err error
				if site.IsEnabled {
					err = m.nginxManager.DisableSite(site.Name)
				} else {
					err = m.nginxManager.EnableSite(site.Name)
				}
				
				if err == nil {
					// Test config
					if testErr := m.nginxManager.TestConfig(); testErr == nil {
						m.nginxManager.ReloadNginx()
						m.sites, _ = m.nginxManager.GetAllSites()
					}
				}
			}

		case "t":
			// Test nginx config
			if err := m.nginxManager.TestConfig(); err != nil {
				m.err = err
			} else {
				m.err = nil
			}

		case "enter", " ":
			// View/edit site details
			if m.viewMode == SitesListView && len(m.sites) > 0 {
				selectedSite := m.sites[m.cursor]
				return m, func() tea.Msg {
					return NavigateMsg{
						Screen: ConfigEditorScreen,
						Data: map[string]interface{}{
							"action": "edit_nginx_site",
							"site":   selectedSite,
						},
					}
				}
			}
		}
	}

	return m, nil
}

// View renders the Nginx config screen
func (m NginxConfigModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	// Header with host info
	hostInfo := system.GetHostInfo()
	headerText := "Nginx Configuration"
	if hostInfo != "" {
		headerText = fmt.Sprintf("Nginx Configuration  %s", m.theme.DescriptionStyle.Render(hostInfo))
	}
	header := m.theme.Title.Render(headerText)

	// Tab selection
	tabSites := "Sites"
	tabGlobal := "Global Config"
	
	if m.viewMode == SitesListView {
		tabSites = m.theme.SelectedItem.Render("[ Sites ]")
		tabGlobal = m.theme.MenuItem.Render("  Global Config  ")
	} else {
		tabSites = m.theme.MenuItem.Render("  Sites  ")
		tabGlobal = m.theme.SelectedItem.Render("[ Global Config ]")
	}
	
	tabs := lipgloss.JoinHorizontal(lipgloss.Left, tabSites, "  ", tabGlobal)

	var content string
	if m.viewMode == SitesListView {
		content = m.renderSitesView()
	} else {
		content = m.renderGlobalConfigView()
	}

	// Error message if any
	errorMsg := ""
	if m.err != nil {
		errorMsg = m.theme.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	// Help text
	help := ""
	if m.viewMode == SitesListView {
		help = m.theme.Help.Render(m.theme.Symbols.ArrowUp + "/" + m.theme.Symbols.ArrowDown + ": Navigate " + m.theme.Symbols.Bullet + " Enter: Edit " + m.theme.Symbols.Bullet + " a: Add " + m.theme.Symbols.Bullet + " e: Enable/Disable " + m.theme.Symbols.Bullet + " t: Test " + m.theme.Symbols.Bullet + " r: Refresh " + m.theme.Symbols.Bullet + " Esc: Back")
	} else {
		help = m.theme.Help.Render("Tab: Switch to Sites " + m.theme.Symbols.Bullet + " Esc: Back " + m.theme.Symbols.Bullet + " q: Quit")
	}

	// Combine all sections
	fullContent := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		tabs,
		"",
	)

	if errorMsg != "" {
		fullContent = lipgloss.JoinVertical(lipgloss.Left, fullContent, errorMsg, "")
	}

	fullContent = lipgloss.JoinVertical(
		lipgloss.Left,
		fullContent,
		content,
		"",
		help,
	)

	// Add border and center
	bordered := m.theme.RenderBox(fullContent)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}

// renderSitesView renders the sites list
func (m NginxConfigModel) renderSitesView() string {
	if len(m.sites) == 0 {
		return m.theme.WarningStyle.Render("No sites configured. Press 'a' to add a new site.")
	}

	// Summary
	totalSites := len(m.sites)
	enabledSites := 0
	sslSites := 0
	for _, site := range m.sites {
		if site.IsEnabled {
			enabledSites++
		}
		if site.HasSSL {
			sslSites++
		}
	}
	summary := m.theme.InfoStyle.Render(fmt.Sprintf("Total Sites: %d | Enabled: %d | SSL: %d", totalSites, enabledSites, sslSites))

	// Table header
	headerStyle := m.theme.Label
	headers := []string{
		headerStyle.Render("Site Name"),
		headerStyle.Render("Domain"),
		headerStyle.Render("Status"),
		headerStyle.Render("SSL"),
		headerStyle.Render("Root Directory"),
	}
	headerRow := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Width(25).Render(headers[0]),
		lipgloss.NewStyle().Width(25).Render(headers[1]),
		lipgloss.NewStyle().Width(12).Render(headers[2]),
		lipgloss.NewStyle().Width(8).Render(headers[3]),
		lipgloss.NewStyle().Width(35).Render(headers[4]),
	)

	// Table rows (with pagination)
	var rows []string
	rows = append(rows, headerRow)
	rows = append(rows, strings.Repeat("─", 105))

	// Calculate visible range
	startIdx := m.scrollOffset
	endIdx := m.scrollOffset + m.maxVisible
	if endIdx > len(m.sites) {
		endIdx = len(m.sites)
	}

	// Show scroll indicators
	if m.scrollOffset > 0 {
		rows = append(rows, m.theme.DescriptionStyle.Render("  ↑ More sites above..."))
	}

	for idx := startIdx; idx < endIdx; idx++ {
		site := m.sites[idx]
		i := idx
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render(m.theme.Symbols.Cursor + " ")
		}

		// Site name
		siteName := site.Name
		if len(siteName) > 23 {
			siteName = siteName[:20] + "..."
		}
		siteNameCol := lipgloss.NewStyle().Width(23).Render(siteName)

		// Domain
		domain := site.Domain
		if domain == "" {
			domain = m.theme.DescriptionStyle.Render("(not set)")
		}
		if len(domain) > 23 {
			domain = domain[:20] + "..."
		}
		domainCol := lipgloss.NewStyle().Width(25).Render(domain)

		// Status (Enabled/Disabled)
		statusBadge := ""
		if site.IsEnabled {
			statusBadge = m.theme.SuccessStyle.Render("✓ Live")
		} else {
			statusBadge = m.theme.DescriptionStyle.Render("○ Disabled")
		}
		statusCol := lipgloss.NewStyle().Width(12).Render(statusBadge)

		// SSL badge
		sslBadge := ""
		if site.HasSSL {
			sslBadge = m.theme.SuccessStyle.Render("Yes")
		} else {
			sslBadge = m.theme.WarningStyle.Render("No")
		}
		sslCol := lipgloss.NewStyle().Width(8).Render(sslBadge)

		// Root directory
		rootDir := site.RootDir
		if rootDir == "" {
			rootDir = m.theme.DescriptionStyle.Render("(not set)")
		}
		if len(rootDir) > 33 {
			rootDir = "..." + rootDir[len(rootDir)-30:]
		}
		rootCol := lipgloss.NewStyle().Width(35).Render(rootDir)

		row := lipgloss.JoinHorizontal(
			lipgloss.Left,
			cursor,
			siteNameCol,
			domainCol,
			statusCol,
			sslCol,
			rootCol,
		)

		if i == m.cursor {
			row = m.theme.SelectedItem.Render(row)
		} else {
			row = m.theme.MenuItem.Render(row)
		}

		rows = append(rows, row)
	}

	// Show scroll indicator at bottom
	if endIdx < len(m.sites) {
		rows = append(rows, m.theme.DescriptionStyle.Render("  ↓ More sites below..."))
	}

	table := lipgloss.JoinVertical(lipgloss.Left, rows...)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		summary,
		"",
		table,
	)
}

// renderGlobalConfigView renders the global config view
func (m NginxConfigModel) renderGlobalConfigView() string {
	content := m.theme.InfoStyle.Render("Global Nginx Configuration")
	
	info := `
Main Config: /etc/nginx/nginx.conf
Sites Available: /etc/nginx/sites-available/
Sites Enabled: /etc/nginx/sites-enabled/

Commands:
  nginx -t              Test configuration
  systemctl reload nginx Reload nginx
  systemctl restart nginx Restart nginx
  
Global configuration editing coming soon...
`

	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		"",
		m.theme.DescriptionStyle.Render(info),
	)
}
