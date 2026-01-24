package screens

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// ConfigMenuItem represents a configuration menu item
type ConfigMenuItem struct {
	ID          string
	Name        string
	Description string
	Available   bool
	Screen      ScreenType
}

// ConfigMenuModel represents the configuration menu screen
type ConfigMenuModel struct {
	theme  *theme.Theme
	width  int
	height int
	cursor int
	items  []ConfigMenuItem
}

// isServiceInstalled checks if a service is installed
func isServiceInstalled(serviceName string) bool {
	cmd := exec.Command("systemctl", "list-unit-files", serviceName+".service")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(output) > 0
}

// isFirewallInstalled checks if UFW or firewalld is installed
func isFirewallInstalled() bool {
	// Check for ufw binary
	if cmd := exec.Command("which", "ufw"); cmd.Run() == nil {
		return true
	}
	// Check for firewall-cmd binary (firewalld)
	if cmd := exec.Command("which", "firewall-cmd"); cmd.Run() == nil {
		return true
	}
	return false
}

// NewConfigMenuModel creates a new configuration menu model
func NewConfigMenuModel() ConfigMenuModel {
	// Check service installation status
	nginxInstalled := isServiceInstalled("nginx")
	redisInstalled := isServiceInstalled("redis-server") || isServiceInstalled("redis")
	mysqlInstalled := isServiceInstalled("mysql")
	postgresqlInstalled := isServiceInstalled("postgresql")
	phpfpmInstalled := isServiceInstalled("php8.3-fpm") || isServiceInstalled("php8.2-fpm") || isServiceInstalled("php8.1-fpm")
	supervisorInstalled := isServiceInstalled("supervisor")
	firewallInstalled := isFirewallInstalled()
	
	items := []ConfigMenuItem{
		{
			ID:          "nginx",
			Name:        "Nginx Web Server",
			Description: getDescription(nginxInstalled, "Manage sites, virtual hosts, and SSL certificates"),
			Available:   nginxInstalled,
			Screen:      NginxConfigScreen,
		},
		{
			ID:          "redis",
			Name:        "Redis Cache",
			Description: getDescription(redisInstalled, "Configure Redis server settings and authentication"),
			Available:   redisInstalled,
			Screen:      RedisConfigScreen,
		},
		{
			ID:          "mysql",
			Name:        "MySQL Database",
			Description: getDescription(mysqlInstalled, "Manage MySQL databases, passwords, and port configuration"),
			Available:   mysqlInstalled,
			Screen:      MySQLManagementScreen,
		},
		{
			ID:          "postgresql",
			Name:        "PostgreSQL Database",
			Description: getDescription(postgresqlInstalled, "Manage PostgreSQL databases, passwords, and performance tuning"),
			Available:   postgresqlInstalled,
			Screen:      PostgreSQLManagementScreen,
		},
		{
			ID:          "php",
			Name:        "PHP-FPM Pools",
			Description: getDescription(phpfpmInstalled, "Manage PHP-FPM pools and worker process configuration"),
			Available:   phpfpmInstalled,
			Screen:      PHPFPMManagementScreen,
		},
		{
			ID:          "supervisor",
			Name:        "Supervisor",
			Description: getDescription(supervisorInstalled, "Manage supervisor programs and XML-RPC configuration"),
			Available:   supervisorInstalled,
			Screen:      SupervisorManagementScreen,
		},
		{
			ID:          "firewall",
			Name:        "Firewall (UFW/firewalld)",
			Description: getDescription(firewallInstalled, "Manage firewall rules, ports, and security settings"),
			Available:   firewallInstalled,
			Screen:      FirewallManagementScreen,
		},
	}

	return ConfigMenuModel{
		theme:  theme.DefaultTheme(),
		items:  items,
		cursor: 0,
	}
}

// getDescription returns description with installation status
func getDescription(installed bool, desc string) string {
	if installed {
		return desc
	}
	return desc + " (Not Installed)"
}

// Init initializes the configuration menu
func (m ConfigMenuModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the configuration menu
func (m ConfigMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		case "enter", " ":
			selectedItem := m.items[m.cursor]
			if selectedItem.Available {
				return m, func() tea.Msg {
					return NavigateMsg{Screen: selectedItem.Screen}
				}
			}
		}
	}

	return m, nil
}

// View renders the configuration menu
func (m ConfigMenuModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	header := m.theme.Title.Render("Configurations")

	// Description
	desc := m.theme.DescriptionStyle.Render("Manage service configurations for installed applications")

	// Menu items
	var menuItems []string
	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		// Build item display
		itemName := item.Name
		if !item.Available {
			itemName += " " + m.theme.WarningStyle.Render("[Not Installed]")
		}

		var renderedItem string
		if i == m.cursor {
			if item.Available {
				renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, itemName))
			} else {
				// Dim style for disabled items
				renderedItem = m.theme.DescriptionStyle.Render(fmt.Sprintf("%s%s", cursor, itemName))
			}
		} else {
			if item.Available {
				renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, itemName))
			} else {
				renderedItem = m.theme.DescriptionStyle.Render(fmt.Sprintf("%s%s", cursor, itemName))
			}
		}

		itemDesc := m.theme.DescriptionStyle.Render(fmt.Sprintf("  %s", item.Description))

		menuItems = append(menuItems, renderedItem)
		menuItems = append(menuItems, itemDesc)
		menuItems = append(menuItems, "")
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, menuItems...)

	// Help
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Select • Esc: Back • q: Quit")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		desc,
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
