package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/models"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// MenuItem represents a menu item
type MenuItem struct {
	Title       string
	Description string
	Screen      ScreenType
}

// MainMenuModel represents the main menu screen
type MainMenuModel struct {
	theme      *theme.Theme
	width      int
	height     int
	cursor     int
	items      []MenuItem
	systemInfo *models.SystemInfo
	detector   *system.Detector
	version    string
}

// NewMainMenuModel creates a new main menu model
func NewMainMenuModel(version string) MainMenuModel {
	detector := system.NewDetector()
	systemInfo, _ := detector.GetSystemInfo()

	items := []MenuItem{
		{
			Title:       "Setup",
			Description: "Install server software packages",
			Screen:      SetupMenuScreen,
		},
		{
			Title:       "Installed Applications",
			Description: "View and manage installed services",
			Screen:      InstalledAppsScreen,
		},
		{
			Title:       "Configurations",
			Description: "Manage service configurations (Nginx, Redis, MySQL, etc.)",
			Screen:      ConfigMenuScreen,
		},
		{
			Title:       "Quick Commands",
			Description: "Execute common administrative tasks",
			Screen:      QuickCommandsScreen,
		},
		{
			Title:       "User Management",
			Description: "Manage users, groups, and sudo privileges",
			Screen:      UserManagementScreen,
		},
	}

	return MainMenuModel{
		theme:      theme.DefaultTheme(),
		items:      items,
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
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		case "enter", " ":
			selectedItem := m.items[m.cursor]
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
			infoLines = append(infoLines, m.theme.WarningStyle.Render("⚠ Not running as root"))
		}
		
		sysInfo = m.theme.InfoStyle.Render(lipgloss.JoinVertical(lipgloss.Left, infoLines...))
	}

	// Menu items
	var menuItems []string
	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		title := item.Title
		desc := m.theme.DescriptionStyle.Render(item.Description)

		var renderedItem string
		if i == m.cursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, title))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, title))
		}

		menuItems = append(menuItems, renderedItem)
		menuItems = append(menuItems, "  "+desc)
		menuItems = append(menuItems, "")
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, menuItems...)

	// Help
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Select • q: Quit")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		sysInfo,
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
