package screens

import (
	"embed"
	"fmt"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/models"
	"github.com/iperamuna/ravact/internal/setup"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// SetupMenuModel represents the setup menu screen
type SetupMenuModel struct {
	theme           *theme.Theme
	width           int
	height          int
	cursor          int
	scripts         []models.SetupScript
	executor        *setup.Executor
	detector        *system.Detector
	serviceStatuses map[string]models.ServiceStatus
	loading         bool
	err             error
}

// NewSetupMenuModel creates a new setup menu model
func NewSetupMenuModel(scriptsDir string) SetupMenuModel {
	executor := setup.NewExecutor(scriptsDir)
	detector := system.NewDetector()
	
	// Read scripts from embedded filesystem
	var scripts []models.SetupScript
	var err error
	
	// Scripts to skip from display
	skipScripts := map[string]bool{
		"php-simple":  true, // Removed in favor of unified PHP management
		"frankenphp":  true, // Available via Site Commands → FrankenPHP Classic Mode
		"nodejs":      true, // Available via Site Commands → NPM operations
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
		} else {
			err = readErr
		}
	} else {
		// Fallback to filesystem (for old behavior)
		scripts, err = executor.GetAvailableScripts()
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

	// Check installation status for all scripts
	serviceStatuses := make(map[string]models.ServiceStatus)
	for _, script := range scripts {
		if script.ServiceID != "" {
			status, _ := detector.GetServiceStatus(script.ServiceID)
			serviceStatuses[script.ID] = status
		}
	}

	return SetupMenuModel{
		theme:           theme.DefaultTheme(),
		cursor:          0,
		scripts:         scripts,
		executor:        executor,
		detector:        detector,
		serviceStatuses: serviceStatuses,
		err:             err,
	}
}

// Init initializes the setup menu
func (m SetupMenuModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the setup menu
func (m SetupMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.scripts)-1 {
				m.cursor++
			}

		case "enter", " ":
			if len(m.scripts) > 0 {
				// Check if running on macOS and warn user
				if runtime.GOOS == "darwin" {
					// Still allow navigation but the script will fail with clear error
					// User should see the macOS warning in the menu
				}
				
				selectedScript := m.scripts[m.cursor]
				status := m.serviceStatuses[selectedScript.ID]
				
				// Navigate to action selection screen
				return m, func() tea.Msg {
					return NavigateMsg{
						Screen: SetupActionScreen,
						Data: map[string]interface{}{
							"script": selectedScript,
							"status": status,
						},
					}
				}
			}

		case "i":
			// Quick install (bypass action menu)
			if len(m.scripts) > 0 {
				// Prevent installation on non-Linux systems
				if runtime.GOOS != "linux" {
					// Don't execute - user will see warning in menu
					return m, nil
				}
				
				selectedScript := m.scripts[m.cursor]
				return m, func() tea.Msg {
					return ExecutionStartMsg{
						Command:     fmt.Sprintf("assets/scripts/%s", selectedScript.ScriptPath),
						Description: fmt.Sprintf("Installing %s", selectedScript.Name),
					}
				}
			}

		case "r":
			// Quick refresh status
			if len(m.scripts) > 0 {
				selectedScript := m.scripts[m.cursor]
				if selectedScript.ServiceID != "" {
					status, _ := m.detector.GetServiceStatus(selectedScript.ServiceID)
					m.serviceStatuses[selectedScript.ID] = status
				}
			}
		}
	}

	return m, nil
}

// View renders the setup menu
func (m SetupMenuModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header with host info
	hostInfo := system.GetHostInfo()
	headerText := "Setup - Install Software"
	if hostInfo != "" {
		headerText = fmt.Sprintf("Setup - Install Software  %s", m.theme.DescriptionStyle.Render(hostInfo))
	}
	header := m.theme.Title.Render(headerText)

	// OS compatibility warning
	var osWarning string
	if runtime.GOOS != "linux" {
		osWarning = m.theme.ErrorStyle.Render(fmt.Sprintf("⚠ WARNING: Setup scripts require Linux. Current OS: %s", runtime.GOOS)) + "\n" +
			m.theme.DescriptionStyle.Render("   These scripts will not work on macOS/Windows. Use Docker or a Linux VM for testing.")
	}

	// Error message if scripts couldn't be loaded
	var errorMsg string
	if m.err != nil {
		errorMsg = m.theme.ErrorStyle.Render(fmt.Sprintf("Error loading scripts: %v", m.err))
	}

	// Menu items
	var menuItems []string
	if len(m.scripts) == 0 {
		noScripts := m.theme.WarningStyle.Render("No setup scripts available")
		menuItems = append(menuItems, noScripts)
		menuItems = append(menuItems, "")
		menuItems = append(menuItems, m.theme.DescriptionStyle.Render("Add .sh scripts to assets/scripts/ directory"))
	} else {
		for i, script := range m.scripts {
			cursor := "  "
			if i == m.cursor {
				cursor = m.theme.KeyStyle.Render(m.theme.Symbols.Cursor + " ")
			}

			// Get installation status
			status := m.serviceStatuses[script.ID]
			statusBadge := ""
			switch status {
			case models.StatusNotInstalled:
				statusBadge = m.theme.DescriptionStyle.Render("[Not Installed]")
			case models.StatusInstalled:
				statusBadge = m.theme.InfoStyle.Render("[Installed]")
			case models.StatusRunning:
				statusBadge = m.theme.SuccessStyle.Render("[✓ Running]")
			case models.StatusStopped:
				statusBadge = m.theme.WarningStyle.Render("[⚠ Stopped]")
			case models.StatusFailed:
				statusBadge = m.theme.ErrorStyle.Render("[✗ Failed]")
			}

			title := script.Name
			if statusBadge != "" {
				title = fmt.Sprintf("%s %s", title, statusBadge)
			}

			desc := ""
			if script.Description != "" {
				desc = m.theme.DescriptionStyle.Render(script.Description)
			}

			var renderedItem string
			if i == m.cursor {
				renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, title))
			} else {
				renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, title))
			}

			menuItems = append(menuItems, renderedItem)
			if desc != "" {
				menuItems = append(menuItems, "  "+desc)
			}
			menuItems = append(menuItems, "")
		}
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, menuItems...)

	// Help
	help := m.theme.Help.Render(m.theme.Symbols.ArrowUp + "/" + m.theme.Symbols.ArrowDown + ": Navigate " + m.theme.Symbols.Bullet + " Enter: Actions " + m.theme.Symbols.Bullet + " i: Install " + m.theme.Symbols.Bullet + " r: Refresh " + m.theme.Symbols.Bullet + " Esc: Back " + m.theme.Symbols.Bullet + " q: Quit")

	// Warning about root
	warning := m.theme.WarningStyle.Render("Note: Installation requires root privileges")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
	)

	if osWarning != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content, osWarning, "")
	}

	if errorMsg != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content, errorMsg, "")
	}

	content = lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		warning,
		"",
		"",
		menu,
		"",
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
