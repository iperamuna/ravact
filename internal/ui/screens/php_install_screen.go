package screens

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// PHPInstallVersion represents a PHP version for installation
type PHPInstallVersion struct {
	Version     string
	Label       string
	Description string
	Installed   bool
}

// PHPInstallAction represents an action in the PHP install screen
type PHPInstallAction struct {
	ID          string
	Name        string
	Description string
}

// PHPInstallModel represents the PHP installation screen
type PHPInstallModel struct {
	theme             *theme.Theme
	width             int
	height            int
	cursor            int
	versions          []PHPInstallVersion
	actions           []PHPInstallAction
	installedVersions []string
	mode              string // "versions" or "actions"
	selectedVersion   string
	err               error
	success           string
}

// NewPHPInstallModel creates a new PHP installation model
func NewPHPInstallModel() PHPInstallModel {
	installedVersions := detectInstalledPHPVersions()

	versions := []PHPInstallVersion{
		{Version: "7.4", Label: "PHP 7.4", Description: "Legacy - Security fixes only (EOL Nov 2022)"},
		{Version: "8.0", Label: "PHP 8.0", Description: "JIT compiler, named arguments, attributes (EOL Nov 2023)"},
		{Version: "8.1", Label: "PHP 8.1", Description: "Enums, fibers, readonly properties"},
		{Version: "8.2", Label: "PHP 8.2", Description: "Readonly classes, DNF types, null/false standalone types"},
		{Version: "8.3", Label: "PHP 8.3", Description: "Typed class constants, json_validate(), #[Override]"},
		{Version: "8.4", Label: "PHP 8.4", Description: "Property hooks, asymmetric visibility (Latest)"},
	}

	// Mark installed versions
	for i := range versions {
		versions[i].Installed = isPHPVersionInstalled(versions[i].Version, installedVersions)
	}

	actions := []PHPInstallAction{
		{ID: "extensions", Name: "Install Extensions", Description: "Install additional PHP extensions"},
		{ID: "back", Name: "← Back to Setup", Description: "Return to setup menu"},
	}

	return PHPInstallModel{
		theme:             theme.DefaultTheme(),
		cursor:            0,
		versions:          versions,
		actions:           actions,
		installedVersions: installedVersions,
		mode:              "versions",
	}
}

// detectInstalledPHPVersions finds all installed PHP versions
func detectInstalledPHPVersions() []string {
	var installed []string
	versions := []string{"7.4", "8.0", "8.1", "8.2", "8.3", "8.4"}

	for _, v := range versions {
		binary := fmt.Sprintf("php%s", v)
		cmd := exec.Command("which", binary)
		if err := cmd.Run(); err == nil {
			installed = append(installed, v)
		}
	}

	return installed
}

// isPHPVersionInstalled checks if a specific PHP version is installed
func isPHPVersionInstalled(version string, installed []string) bool {
	for _, v := range installed {
		if v == version {
			return true
		}
	}
	return false
}

// Init initializes the PHP install screen
func (m PHPInstallModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for PHP installation
func (m PHPInstallModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		totalItems := len(m.versions) + len(m.actions)

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc", "backspace":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: SetupMenuScreen}
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < totalItems-1 {
				m.cursor++
			}

		case "enter", " ":
			return m.executeAction()
		}
	}

	return m, nil
}

// executeAction handles the selected item
func (m PHPInstallModel) executeAction() (PHPInstallModel, tea.Cmd) {
	m.err = nil
	m.success = ""

	// Check if cursor is on a version or an action
	if m.cursor < len(m.versions) {
		// Version selected
		version := m.versions[m.cursor]
		
		if version.Installed {
			// Already installed - offer to remove
			return m, func() tea.Msg {
				return ExecutionStartMsg{
					Command:     fmt.Sprintf("apt-get remove -y php%s-fpm php%s-cli php%s-common && apt-get autoremove -y", version.Version, version.Version, version.Version),
					Description: fmt.Sprintf("Removing PHP %s", version.Version),
				}
			}
		} else {
			// Install this version
			installCmd := buildPHPInstallCommand(version.Version)
			return m, func() tea.Msg {
				return ExecutionStartMsg{
					Command:     installCmd,
					Description: fmt.Sprintf("Installing PHP %s with common extensions", version.Version),
				}
			}
		}
	} else {
		// Action selected
		actionIdx := m.cursor - len(m.versions)
		action := m.actions[actionIdx]

		switch action.ID {
		case "extensions":
			if len(m.installedVersions) == 0 {
				m.err = fmt.Errorf("no PHP versions installed. Install a PHP version first")
				return m, nil
			}
			return m, func() tea.Msg {
				return NavigateMsg{Screen: PHPExtensionsScreen}
			}
		case "back":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: SetupMenuScreen}
			}
		}
	}

	return m, nil
}

// buildPHPInstallCommand creates the installation command for a PHP version
func buildPHPInstallCommand(version string) string {
	// Common extensions to install by default
	commonExtensions := []string{
		"cli", "fpm", "common", "mysql", "pgsql", "sqlite3",
		"curl", "gd", "mbstring", "xml", "zip", "bcmath",
		"intl", "soap", "opcache", "readline",
	}

	var extPkgs []string
	for _, ext := range commonExtensions {
		extPkgs = append(extPkgs, fmt.Sprintf("php%s-%s", version, ext))
	}

	return fmt.Sprintf(`# Add PHP repository if needed
if ! grep -q "ondrej/php" /etc/apt/sources.list.d/*.list 2>/dev/null; then
    apt-get update
    apt-get install -y software-properties-common
    add-apt-repository -y ppa:ondrej/php
fi
apt-get update
apt-get install -y %s
# Enable and start PHP-FPM
systemctl enable php%s-fpm
systemctl start php%s-fpm
echo "PHP %s installed successfully!"
php%s --version`, strings.Join(extPkgs, " "), version, version, version, version)
}

// View renders the PHP install screen
func (m PHPInstallModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	header := m.theme.Title.Render("PHP Installation")

	// Status info
	var statusLines []string
	if len(m.installedVersions) > 0 {
		statusLines = append(statusLines, m.theme.Label.Render("Installed Versions: ")+m.theme.SuccessStyle.Render(strings.Join(m.installedVersions, ", ")))
	} else {
		statusLines = append(statusLines, m.theme.WarningStyle.Render("No PHP versions installed"))
	}
	statusSection := lipgloss.JoinVertical(lipgloss.Left, statusLines...)

	// Instructions
	instructions := m.theme.DescriptionStyle.Render("Select a version to install, or select an installed version to remove it.")

	// Version list
	var items []string
	items = append(items, "")
	items = append(items, m.theme.Subtitle.Render("PHP Versions:"))
	items = append(items, "")

	for i, version := range m.versions {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		statusIcon := ""
		if version.Installed {
			statusIcon = m.theme.SuccessStyle.Render(" [Installed]")
		}

		var renderedItem string
		label := fmt.Sprintf("%s%s%s", cursor, version.Label, statusIcon)
		
		if i == m.cursor {
			if version.Installed {
				renderedItem = m.theme.WarningStyle.Render(label)
			} else {
				renderedItem = m.theme.SelectedItem.Render(label)
			}
		} else {
			if version.Installed {
				renderedItem = m.theme.SuccessStyle.Render(label)
			} else {
				renderedItem = m.theme.MenuItem.Render(label)
			}
		}

		items = append(items, renderedItem)

		// Show description for selected item
		if i == m.cursor {
			actionHint := "Press Enter to install"
			if version.Installed {
				actionHint = "Press Enter to remove"
			}
			items = append(items, "    "+m.theme.DescriptionStyle.Render(version.Description))
			items = append(items, "    "+m.theme.InfoStyle.Render(actionHint))
		}
	}

	// Actions section
	items = append(items, "")
	items = append(items, m.theme.Subtitle.Render("Actions:"))
	items = append(items, "")

	for i, action := range m.actions {
		actualIdx := len(m.versions) + i
		cursor := "  "
		if actualIdx == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		var renderedItem string
		if actualIdx == m.cursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, action.Name))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, action.Name))
		}

		items = append(items, renderedItem)
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, items...)

	// Messages
	var messages []string
	if m.success != "" {
		messages = append(messages, m.theme.SuccessStyle.Render(m.success))
	}
	if m.err != nil {
		messages = append(messages, m.theme.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}
	messageSection := ""
	if len(messages) > 0 {
		messageSection = lipgloss.JoinVertical(lipgloss.Left, messages...)
	}

	// Help
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Install/Remove • Esc: Back • q: Quit")

	// Combine all sections
	sections := []string{
		header,
		"",
		statusSection,
		"",
		instructions,
		menu,
	}

	if messageSection != "" {
		sections = append(sections, "", messageSection)
	}

	sections = append(sections, "", help)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

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
