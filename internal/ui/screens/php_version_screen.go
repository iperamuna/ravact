package screens

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// PHPVersion represents a PHP version option
type PHPVersion struct {
	Version     string
	Label       string
	Description string
	Binary      string
}

type PHPVersionModel struct {
	theme             *theme.Theme
	width             int
	height            int
	cursor            int
	versions          []PHPVersion
	commandType       string // "composer_install", "artisan_migrate", etc.
	currentVersion    string
	availableVersions []string
	systemUser        string // from git config meta.systemuser
	availableUsers    []string
	selectingUser     bool
}

// NewPHPVersionModel creates a new PHP version selection model
func NewPHPVersionModel(commandType string) PHPVersionModel {
	// Detect available PHP versions
	availableVersions := detectAvailablePHPVersions()
	currentVersion := detectPHPVersion()

	versions := []PHPVersion{
		{Version: "current", Label: "Use Current Version", Description: fmt.Sprintf("Use default php (%s)", currentVersion), Binary: "php"},
	}

	// Add available PHP versions
	phpVersions := []struct {
		version string
		label   string
		desc    string
	}{
		{"7.4", "PHP 7.4", "Legacy - Security fixes only"},
		{"8.0", "PHP 8.0", "JIT compiler, named arguments, attributes"},
		{"8.1", "PHP 8.1", "Enums, fibers, readonly properties"},
		{"8.2", "PHP 8.2", "Readonly classes, DNF types"},
		{"8.3", "PHP 8.3", "Typed class constants, json_validate()"},
		{"8.4", "PHP 8.4", "Property hooks, asymmetric visibility"},
	}

	for _, pv := range phpVersions {
		binary := fmt.Sprintf("php%s", pv.version)
		available := isVersionAvailable(pv.version, availableVersions)

		label := pv.label
		desc := pv.desc
		if !available {
			label = pv.label + " (not installed)"
			desc = "Not available - install with: sudo apt install php" + pv.version
		}

		versions = append(versions, PHPVersion{
			Version:     pv.version,
			Label:       label,
			Description: desc,
			Binary:      binary,
		})
	}

	// Get system user from git config
	systemUser := getGitSystemUser()

	// Get available users for selection
	um := system.NewUserManager()
	allUsers, _ := um.GetAllUsers()
	var availableUsers []string
	for _, user := range allUsers {
		if user.UID >= 1000 || user.Username == "www-data" {
			availableUsers = append(availableUsers, user.Username)
		}
	}

	m := PHPVersionModel{
		theme:             theme.DefaultTheme(),
		cursor:            0,
		versions:          versions,
		commandType:       commandType,
		currentVersion:    currentVersion,
		availableVersions: availableVersions,
		systemUser:        systemUser,
		availableUsers:    availableUsers,
	}

	// If system user is missing, start in selection mode
	if m.systemUser == "" {
		m.selectingUser = true
	}

	return m
}

// detectPHPVersion gets the current default PHP version
func detectPHPVersion() string {
	cmd := exec.Command("php", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "Not installed"
	}
	// Parse first line to get version
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		parts := strings.Fields(lines[0])
		if len(parts) >= 2 {
			return parts[1]
		}
	}
	return "Unknown"
}

// detectAvailablePHPVersions finds installed PHP versions
func detectAvailablePHPVersions() []string {
	var available []string

	versions := []string{"7.4", "8.0", "8.1", "8.2", "8.3", "8.4"}

	for _, v := range versions {
		binary := fmt.Sprintf("php%s", v)
		cmd := exec.Command("which", binary)
		if err := cmd.Run(); err == nil {
			available = append(available, v)
		}
	}

	return available
}

// isVersionAvailable checks if a PHP version is installed
func isVersionAvailable(version string, available []string) bool {
	for _, v := range available {
		if v == version {
			return true
		}
	}
	return false
}

// Init initializes the PHP version screen
func (m PHPVersionModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for PHP version selection
func (m PHPVersionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				return NavigateMsg{Screen: SiteCommandsScreen}
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.versions)-1 {
				m.cursor++
			}

		case "enter", " ":
			if m.selectingUser {
				m.systemUser = m.availableUsers[m.cursor]
				m.selectingUser = false
				m.cursor = 0

				// Try to save to git config if it's a git repo
				cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
				if err := cmd.Run(); err == nil {
					exec.Command("git", "config", "meta.systemuser", m.systemUser).Run()
				}
				return m, nil
			}
			return m.executeCommand()
		}
	}

	return m, nil
}

// executeCommand runs the PHP command with selected version
func (m PHPVersionModel) executeCommand() (PHPVersionModel, tea.Cmd) {
	selectedVersion := m.versions[m.cursor]

	// Check if selected version is available (skip for "current")
	if selectedVersion.Version != "current" && !isVersionAvailable(selectedVersion.Version, m.availableVersions) {
		// Version not installed, don't execute
		return m, nil
	}

	var command string
	var description string

	phpBinary := selectedVersion.Binary

	switch m.commandType {
	case "composer_install":
		command = fmt.Sprintf("%s $(which composer) install --no-interaction", phpBinary)
		description = fmt.Sprintf("Running composer install with %s", selectedVersion.Label)

	case "artisan_migrate":
		command = fmt.Sprintf("%s artisan migrate --force", phpBinary)
		description = fmt.Sprintf("Running migrations with %s", selectedVersion.Label)

	case "artisan_cache_clear":
		command = fmt.Sprintf("%s artisan config:clear && %s artisan route:clear && %s artisan view:clear && %s artisan cache:clear", phpBinary, phpBinary, phpBinary, phpBinary)
		description = fmt.Sprintf("Clearing caches with %s", selectedVersion.Label)

	case "artisan_optimize":
		command = fmt.Sprintf("%s artisan optimize", phpBinary)
		description = fmt.Sprintf("Optimizing with %s", selectedVersion.Label)

	default:
		command = fmt.Sprintf("%s artisan", phpBinary)
		description = "Running artisan"
	}

	// If system user is configured, wrap in sudo -i -u
	if m.systemUser != "" {
		cwd, _ := os.Getwd()
		command = fmt.Sprintf(`sudo -i -u %s bash << 'EOF'
cd "%s"
%s
EOF
`, m.systemUser, cwd, command)
		description = fmt.Sprintf("%s (as %s)", description, m.systemUser)
	}

	return m, func() tea.Msg {
		return ExecutionStartMsg{
			Command:     command,
			Description: description,
		}
	}
}

// View renders the PHP version selection screen
func (m PHPVersionModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Handle user selection state
	if m.selectingUser {
		return m.viewUserSelection()
	}

	// Header
	title := "Select PHP Version"
	switch m.commandType {
	case "composer_install":
		title = "Composer Install - Select PHP Version"
	case "artisan_migrate":
		title = "Artisan Migrate - Select PHP Version"
	case "artisan_cache_clear":
		title = "Clear Caches - Select PHP Version"
	case "artisan_optimize":
		title = "Artisan Optimize - Select PHP Version"
	}
	header := m.theme.Title.Render(title)

	// Current status
	var statusLines []string
	statusLines = append(statusLines, m.theme.Label.Render("Default PHP: ")+m.theme.InfoStyle.Render(m.currentVersion))

	if len(m.availableVersions) > 0 {
		statusLines = append(statusLines, m.theme.Label.Render("Installed: ")+m.theme.SuccessStyle.Render(strings.Join(m.availableVersions, ", ")))
	} else {
		statusLines = append(statusLines, m.theme.WarningStyle.Render("⚠ No additional PHP versions detected"))
	}

	// Show system user if configured
	if m.systemUser != "" {
		statusLines = append(statusLines, m.theme.Label.Render("Run as: ")+m.theme.SuccessStyle.Render(m.systemUser)+" (from git config)")
	}

	statusSection := lipgloss.JoinVertical(lipgloss.Left, statusLines...)

	// Version options
	var versionItems []string
	versionItems = append(versionItems, "")
	versionItems = append(versionItems, m.theme.Subtitle.Render("Select PHP Version:"))
	versionItems = append(versionItems, "")

	for i, version := range m.versions {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		// Check if version is available
		available := version.Version == "current" || isVersionAvailable(version.Version, m.availableVersions)

		var renderedItem string
		if i == m.cursor {
			if available {
				renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, version.Label))
			} else {
				renderedItem = m.theme.WarningStyle.Render(fmt.Sprintf("%s%s", cursor, version.Label))
			}
		} else {
			if available {
				renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, version.Label))
			} else {
				renderedItem = m.theme.DescriptionStyle.Render(fmt.Sprintf("%s%s", cursor, version.Label))
			}
		}

		versionItems = append(versionItems, renderedItem)

		// Show description for selected item
		if i == m.cursor {
			versionItems = append(versionItems, "    "+m.theme.DescriptionStyle.Render(version.Description))
		}
	}

	versionsMenu := lipgloss.JoinVertical(lipgloss.Left, versionItems...)

	// Help
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Run • Esc: Back • q: Quit")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		statusSection,
		versionsMenu,
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

func (m PHPVersionModel) viewUserSelection() string {
	header := m.theme.Title.Render("Select System User")

	description := m.theme.DescriptionStyle.Render("Select a user to run PHP/Composer commands as.")

	var items []string
	items = append(items, "")
	for i, user := range m.availableUsers {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		var renderedItem string
		if i == m.cursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, user))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, user))
		}
		items = append(items, renderedItem)
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, items...)
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Select • Esc: Back")

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", description, menu, "", help)
	bordered := m.theme.RenderBox(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}
