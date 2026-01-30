package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/models"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// MenuCategory represents a category of menu items
type MenuCategory struct {
	Name  string
	Icon  string
	Items []MenuItem
}

// MenuItem represents a menu item
type MenuItem struct {
	Title       string
	Description string
	Screen      ScreenType
	Category    string
}

// MainMenuModel represents the main menu screen
type MainMenuModel struct {
	theme      *theme.Theme
	width      int
	height     int
	cursor     int
	categories []MenuCategory
	flatItems  []MenuItem // Flattened list for navigation
	systemInfo *models.SystemInfo
	detector   *system.Detector
	version    string
}

// NewMainMenuModel creates a new main menu model
func NewMainMenuModel(version string) MainMenuModel {
	detector := system.NewDetector()
	systemInfo, _ := detector.GetSystemInfo()
	t := theme.DefaultTheme()

	// Define menu categories following industry-standard organization
	categories := []MenuCategory{
		{
			Name: "Package Management",
			Icon: t.Symbols.Box,
			Items: []MenuItem{
				{
					Title:       "Install Software",
					Description: "Install server packages (Nginx, MySQL, PHP, Redis, etc.)",
					Screen:      SetupMenuScreen,
					Category:    "Package Management",
				},
				{
					Title:       "Installed Applications",
					Description: "View and manage installed services",
					Screen:      InstalledAppsScreen,
					Category:    "Package Management",
				},
			},
		},
		{
			Name: "Service Configuration",
			Icon: t.Symbols.Bullet,
			Items: []MenuItem{
				{
					Title:       "Service Settings",
					Description: "Configure Nginx, MySQL, PostgreSQL, Redis, PHP-FPM, etc.",
					Screen:      ConfigMenuScreen,
					Category:    "Service Configuration",
				},
			},
		},
		{
			Name: "Site Management",
			Icon: t.Symbols.ArrowRight,
			Items: []MenuItem{
				{
					Title:       "Site Commands",
					Description: "Git, Laravel, Composer, NPM, and deployment tools",
					Screen:      SiteCommandsScreen,
					Category:    "Site Management",
				},
				{
					Title:       "Developer Toolkit",
					Description: "Essential commands for Laravel & WordPress maintenance",
					Screen:      DeveloperToolkitScreen,
					Category:    "Site Management",
				},
			},
		},
		{
			Name: "System Administration",
			Icon: t.Symbols.Info,
			Items: []MenuItem{
				{
					Title:       "User Management",
					Description: "Manage users, groups, and sudo privileges",
					Screen:      UserManagementScreen,
					Category:    "System Administration",
				},
				{
					Title:       "Quick Commands",
					Description: "System diagnostics, logs, and service controls",
					Screen:      QuickCommandsScreen,
					Category:    "System Administration",
				},
			},
		},
		{
			Name: "Tools",
			Icon: t.Symbols.Box,
			Items: []MenuItem{
				{
					Title:       "File Browser",
					Description: "Full-featured file manager with preview and operations",
					Screen:      FileBrowserScreen,
					Category:    "Tools",
				},
			},
		},
	}

	// Flatten items for navigation
	var flatItems []MenuItem
	for _, cat := range categories {
		flatItems = append(flatItems, cat.Items...)
	}

	return MainMenuModel{
		theme:      t,
		categories: categories,
		flatItems:  flatItems,
		cursor:     0,
		systemInfo: systemInfo,
		detector:   detector,
		version:    version,
	}
}

// Init initializes the main menu
func (m MainMenuModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the main menu
func (m MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.flatItems)-1 {
				m.cursor++
			}

		case "enter", " ":
			selectedItem := m.flatItems[m.cursor]
			return m, func() tea.Msg {
				return NavigateMsg{Screen: selectedItem.Screen}
			}
		}
	}

	return m, nil
}

// View renders the main menu
func (m MainMenuModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header with version
	headerText := "RAVACT - Main Menu"
	if m.version != "" {
		headerText = fmt.Sprintf("RAVACT v%s - Main Menu", m.version)
	}
	header := m.theme.Title.Render(headerText)

	// System info
	sysInfo := ""
	if m.systemInfo != nil {
		infoLines := []string{}

		// Hostname with IP Address
		if m.systemInfo.Hostname != "" {
			ipAddr := system.GetPrimaryIP()
			if ipAddr != "" && ipAddr != "N/A" {
				infoLines = append(infoLines, fmt.Sprintf("Host: %s (%s)", m.systemInfo.Hostname, ipAddr))
			} else {
				infoLines = append(infoLines, fmt.Sprintf("Host: %s", m.systemInfo.Hostname))
			}
		}

		// OS info
		osInfo := fmt.Sprintf("OS: %s", m.systemInfo.OS)
		if m.systemInfo.Distribution != "" {
			osInfo = fmt.Sprintf("OS: %s %s", m.systemInfo.Distribution, m.systemInfo.Version)
		}
		infoLines = append(infoLines, osInfo)

		// Architecture
		infoLines = append(infoLines, fmt.Sprintf("Arch: %s", m.systemInfo.Arch))

		// CPU
		infoLines = append(infoLines, fmt.Sprintf("CPU: %d cores", m.systemInfo.CPUCount))

		// RAM
		if m.systemInfo.TotalRAM > 0 {
			infoLines = append(infoLines, fmt.Sprintf("RAM: %s", system.FormatBytes(m.systemInfo.TotalRAM)))
		}

		// Disk
		if m.systemInfo.TotalDisk > 0 {
			infoLines = append(infoLines, fmt.Sprintf("Disk: %s", system.FormatBytes(m.systemInfo.TotalDisk)))
		}

		// Root warning
		if !m.systemInfo.IsRoot {
			infoLines = append(infoLines, "")
			infoLines = append(infoLines, m.theme.WarningStyle.Render(m.theme.Symbols.Warning+" Not running as root"))
		}

		sysInfo = m.theme.InfoStyle.Render(lipgloss.JoinVertical(lipgloss.Left, infoLines...))
	}

	// Menu items with categories
	var menuItems []string
	itemIndex := 0

	for _, category := range m.categories {
		// Category header
		categoryHeader := m.theme.CategoryStyle.Render(fmt.Sprintf("%s %s", category.Icon, category.Name))
		menuItems = append(menuItems, categoryHeader)

		// Category items
		for _, item := range category.Items {
			cursor := "  "
			if itemIndex == m.cursor {
				cursor = m.theme.KeyStyle.Render(m.theme.Symbols.Cursor + " ")
			}

			title := item.Title
			desc := m.theme.DescriptionStyle.Render(item.Description)

			var renderedItem string
			if itemIndex == m.cursor {
				renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("  %s%s", cursor, title))
			} else {
				renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("  %s%s", cursor, title))
			}

			menuItems = append(menuItems, renderedItem)
			menuItems = append(menuItems, "      "+desc)
			itemIndex++
		}
		menuItems = append(menuItems, "")
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, menuItems...)

	// Help
	help := m.theme.Help.Render(m.theme.Symbols.ArrowUp + "/" + m.theme.Symbols.ArrowDown + ": Navigate " + m.theme.Symbols.Bullet + " Enter: Select " + m.theme.Symbols.Bullet + " q: Quit")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		sysInfo,
		"",
		menu,
		"",
		help,
	)

	// Add border and center using RenderBox for consistency	// Add border and center using RenderBox for consistency and wrapping
	bordered := m.theme.RenderBox(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}
