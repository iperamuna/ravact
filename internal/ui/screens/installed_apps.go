package screens

import (
	"embed"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/models"
	"github.com/iperamuna/ravact/internal/setup"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// InstalledApp represents an installed application with its info
type InstalledApp struct {
	Script models.SetupScript
	Status models.ServiceStatus
}

// InstalledAppsModel represents the installed applications screen
type InstalledAppsModel struct {
	theme           *theme.Theme
	width           int
	height          int
	cursor          int
	installedApps   []InstalledApp
	detector        *system.Detector
	executor        *setup.Executor
	loading         bool
}

// NewInstalledAppsModel creates a new installed apps model
func NewInstalledAppsModel(scriptsDir string) InstalledAppsModel {
	executor := setup.NewExecutor(scriptsDir)
	detector := system.NewDetector()
	
	// Read scripts from embedded filesystem
	var scripts []models.SetupScript
	
	// Scripts to skip from display
	skipScripts := map[string]bool{
		"php-simple": true, // Removed in favor of unified PHP management
	}

	if EmbeddedFS != (embed.FS{}) {
		// Read from embedded FS
		entries, readErr := EmbeddedFS.ReadDir(scriptsDir)
		if readErr == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sh") {
					scriptID := strings.TrimSuffix(entry.Name(), ".sh")
					// Skip excluded scripts
					if skipScripts[scriptID] {
						continue
					}
					scripts = append(scripts, models.SetupScript{
						ID:         scriptID,
						Name:       scriptID,
						ScriptPath: entry.Name(),
					})
				}
			}
		}
	} else {
		// Fallback to filesystem (for old behavior)
		scripts, _ = executor.GetAvailableScripts()
	}

	// Add descriptions and service IDs for known scripts
	for i := range scripts {
		switch scripts[i].ID {
		case "nginx":
			scripts[i].Name = "Nginx Web Server"
			scripts[i].Description = "High-performance HTTP server and reverse proxy"
			scripts[i].ServiceID = "nginx"
		case "mysql":
			scripts[i].Name = "MySQL Database"
			scripts[i].Description = "Popular open-source relational database"
			scripts[i].ServiceID = "mysql"
		case "postgresql":
			scripts[i].Name = "PostgreSQL"
			scripts[i].Description = "Advanced open-source relational database"
			scripts[i].ServiceID = "postgresql"
		case "redis":
			scripts[i].Name = "Redis Cache"
			scripts[i].Description = "In-memory data structure store and cache"
			scripts[i].ServiceID = "redis-server"
		case "dragonfly":
			scripts[i].Name = "Dragonfly"
			scripts[i].Description = "Modern Redis/Memcached replacement (faster, less memory)"
			scripts[i].ServiceID = "dragonfly"
		case "php":
			scripts[i].Name = "PHP"
			scripts[i].Description = "PHP versions and extensions management"
			scripts[i].ServiceID = "php-fpm"
		case "frankenphp":
			scripts[i].Name = "FrankenPHP"
			scripts[i].Description = "Modern PHP server with Caddy (Classic/Worker/Mercure modes)"
			scripts[i].ServiceID = "frankenphp"
		case "nodejs":
			scripts[i].Name = "Node.js"
			scripts[i].Description = "JavaScript runtime with npm, yarn, and PM2"
			scripts[i].ServiceID = "node"
		case "supervisor":
			scripts[i].Name = "Supervisor"
			scripts[i].Description = "Process control system for Unix-like systems"
			scripts[i].ServiceID = "supervisor"
		case "certbot":
			scripts[i].Name = "Certbot (Let's Encrypt)"
			scripts[i].Description = "Free SSL/TLS certificates from Let's Encrypt"
			scripts[i].ServiceID = "certbot"
		case "git":
			scripts[i].Name = "Git"
			scripts[i].Description = "Git Version control system"
			scripts[i].ServiceID = "git"
		case "firewall":
			scripts[i].Name = "Firewall (UFW/firewalld)"
			scripts[i].Description = "Configure firewall with common rules"
			scripts[i].ServiceID = "ufw"
		}
	}

	// Filter only installed apps
	var installedApps []InstalledApp
	for _, script := range scripts {
		if script.ServiceID != "" {
			status, _ := detector.GetServiceStatus(script.ServiceID)
			// Only include if installed (not StatusNotInstalled or StatusUnknown)
			if status != models.StatusNotInstalled && status != models.StatusUnknown {
				installedApps = append(installedApps, InstalledApp{
					Script: script,
					Status: status,
				})
			}
		}
	}

	return InstalledAppsModel{
		theme:         theme.DefaultTheme(),
		cursor:        0,
		installedApps: installedApps,
		detector:      detector,
		executor:      executor,
	}
}

// Init initializes the installed apps screen
func (m InstalledAppsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the installed apps screen
func (m InstalledAppsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.installedApps)-1 {
				m.cursor++
			}

		case "enter", " ":
			if len(m.installedApps) > 0 {
				selectedApp := m.installedApps[m.cursor]
				// Navigate to action screen
				return m, func() tea.Msg {
					return NavigateMsg{
						Screen: SetupActionScreen,
						Data: map[string]interface{}{
							"script": selectedApp.Script,
							"status": selectedApp.Status,
						},
					}
				}
			}

		case "r":
			// Refresh status
			if len(m.installedApps) > 0 {
				selectedApp := &m.installedApps[m.cursor]
				if selectedApp.Script.ServiceID != "" {
					status, _ := m.detector.GetServiceStatus(selectedApp.Script.ServiceID)
					selectedApp.Status = status
				}
			}
		}
	}

	return m, nil
}

// View renders the installed apps screen
func (m InstalledAppsModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	header := m.theme.Title.Render("Installed Applications")

	// Summary
	summary := m.theme.InfoStyle.Render(fmt.Sprintf("Found %d installed applications", len(m.installedApps)))

	// App items
	var appItems []string
	if len(m.installedApps) == 0 {
		noApps := m.theme.WarningStyle.Render("No applications installed yet")
		appItems = append(appItems, noApps)
		appItems = append(appItems, "")
		appItems = append(appItems, m.theme.DescriptionStyle.Render("Use Setup menu to install applications"))
	} else {
		for i, app := range m.installedApps {
			cursor := "  "
			if i == m.cursor {
				cursor = m.theme.KeyStyle.Render("▶ ")
			}

			// Status badge
			statusBadge := ""
			switch app.Status {
			case models.StatusInstalled:
				statusBadge = m.theme.InfoStyle.Render("[Installed]")
			case models.StatusRunning:
				statusBadge = m.theme.SuccessStyle.Render("[✓ Running]")
			case models.StatusStopped:
				statusBadge = m.theme.WarningStyle.Render("[⚠ Stopped]")
			case models.StatusFailed:
				statusBadge = m.theme.ErrorStyle.Render("[✗ Failed]")
			}

			title := fmt.Sprintf("%s %s", app.Script.Name, statusBadge)
			desc := ""
			if app.Script.Description != "" {
				desc = m.theme.DescriptionStyle.Render(app.Script.Description)
			}

			var renderedItem string
			if i == m.cursor {
				renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, title))
			} else {
				renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, title))
			}

			appItems = append(appItems, renderedItem)
			if desc != "" {
				appItems = append(appItems, "  "+desc)
			}
			appItems = append(appItems, "")
		}
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, appItems...)

	// Help
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Manage • r: Refresh • Esc: Back • q: Quit")

	// Info
	info := m.theme.DescriptionStyle.Render("Tip: Only shows applications managed by Ravact setup scripts")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		summary,
		"",
		"",
		menu,
		"",
		"",
		info,
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
