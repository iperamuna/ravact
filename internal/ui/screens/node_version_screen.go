package screens

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// NodeVersion represents a Node.js version option
type NodeVersion struct {
	Version     string
	Label       string
	Description string
}

// NodeVersionModel represents the node version selection screen
type NodeVersionModel struct {
	theme          *theme.Theme
	width          int
	height         int
	cursor         int
	versions       []NodeVersion
	commandType    string // "npm_install" or "npm_build"
	currentVersion string
	nvmInstalled   bool
	systemUser     string // from git config meta.systemuser
}

// NewNodeVersionModel creates a new node version selection model
func NewNodeVersionModel(commandType string) NodeVersionModel {
	versions := []NodeVersion{
		{Version: "current", Label: "Use Current Version", Description: "Use the currently active Node.js version"},
		{Version: "16", Label: "Node.js 16 (LTS)", Description: "Maintenance LTS - Legacy support"},
		{Version: "18", Label: "Node.js 18 (LTS)", Description: "Active LTS - Recommended for most projects"},
		{Version: "20", Label: "Node.js 20 (LTS)", Description: "Active LTS - Latest features with stability"},
		{Version: "21", Label: "Node.js 21", Description: "Current - Latest features"},
		{Version: "22", Label: "Node.js 22 (LTS)", Description: "Active LTS - Newest LTS version"},
	}

	// Detect current Node version
	currentVersion := detectNodeVersion()
	nvmInstalled := isNvmInstalled()
	
	// Get system user from git config
	systemUser := getNodeSystemUser()

	return NodeVersionModel{
		theme:          theme.DefaultTheme(),
		cursor:         0,
		versions:       versions,
		commandType:    commandType,
		currentVersion: currentVersion,
		nvmInstalled:   nvmInstalled,
		systemUser:     systemUser,
	}
}

// getNodeSystemUser retrieves the meta.systemuser from git config
func getNodeSystemUser() string {
	cmd := exec.Command("git", "config", "--get", "meta.systemuser")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// detectNodeVersion gets the current Node.js version
func detectNodeVersion() string {
	cmd := exec.Command("node", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "Not installed"
	}
	return strings.TrimSpace(string(output))
}

// isNvmInstalled checks if nvm is available
func isNvmInstalled() bool {
	// Check for nvm by looking for the directory
	cmd := exec.Command("bash", "-c", "[ -d \"$HOME/.nvm\" ] && echo yes || echo no")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "yes"
}

// Init initializes the node version screen
func (m NodeVersionModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for node version selection
func (m NodeVersionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			return m.executeCommand()
		}
	}

	return m, nil
}

// executeCommand runs the npm command with selected node version
func (m NodeVersionModel) executeCommand() (NodeVersionModel, tea.Cmd) {
	selectedVersion := m.versions[m.cursor]
	
	var command string
	var description string

	npmCmd := "npm install"
	if m.commandType == "npm_build" {
		npmCmd = "npm install && npm run build"
	}

	// Build the base command
	var baseCmd string
	if selectedVersion.Version == "current" {
		// Use current version directly
		baseCmd = npmCmd
		description = fmt.Sprintf("Running %s (Node %s)", npmCmd, m.currentVersion)
	} else if m.nvmInstalled {
		// Use nvm to switch version
		baseCmd = fmt.Sprintf("source $HOME/.nvm/nvm.sh && nvm use %s && %s", selectedVersion.Version, npmCmd)
		description = fmt.Sprintf("Running %s with Node.js %s", npmCmd, selectedVersion.Version)
	} else {
		// No nvm, but user selected a specific version - warn them
		baseCmd = fmt.Sprintf("echo 'Node.js %s selected but nvm is not installed.' && echo 'Install nvm first: curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash' && echo '' && echo 'Running with current version instead...' && %s", selectedVersion.Version, npmCmd)
		description = fmt.Sprintf("Running %s (nvm not installed, using current)", npmCmd)
	}

	// If system user is configured, run as that user
	if m.systemUser != "" {
		cwd, _ := exec.Command("pwd").Output()
		cwdStr := strings.TrimSpace(string(cwd))
		command = fmt.Sprintf(`su - %s -c 'cd "%s" && %s'`, m.systemUser, cwdStr, baseCmd)
		description = fmt.Sprintf("%s (as %s)", description, m.systemUser)
	} else {
		command = baseCmd
	}

	return m, func() tea.Msg {
		return ExecutionStartMsg{
			Command:     command,
			Description: description,
		}
	}
}

// View renders the node version selection screen
func (m NodeVersionModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	title := "NPM Install"
	if m.commandType == "npm_build" {
		title = "NPM Build"
	}
	header := m.theme.Title.Render(title + " - Select Node Version")

	// Current status
	var statusLines []string
	statusLines = append(statusLines, m.theme.Label.Render("Current Node.js: ")+m.theme.InfoStyle.Render(m.currentVersion))
	
	if m.nvmInstalled {
		statusLines = append(statusLines, m.theme.SuccessStyle.Render("✓ nvm detected - version switching available"))
	} else {
		statusLines = append(statusLines, m.theme.WarningStyle.Render("⚠ nvm not installed - using current version only"))
	}
	
	// Show system user if configured
	if m.systemUser != "" {
		statusLines = append(statusLines, m.theme.Label.Render("Run as: ")+m.theme.SuccessStyle.Render(m.systemUser)+" (from git config)")
	}
	
	statusSection := lipgloss.JoinVertical(lipgloss.Left, statusLines...)

	// Version options
	var versionItems []string
	versionItems = append(versionItems, "")
	versionItems = append(versionItems, m.theme.Subtitle.Render("Select Node.js Version:"))
	versionItems = append(versionItems, "")

	for i, version := range m.versions {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		var renderedItem string
		if i == m.cursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, version.Label))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, version.Label))
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
	bordered := m.theme.BorderStyle.Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}
