package screens

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// LaravelPermAction represents a Laravel permission action
type LaravelPermAction struct {
	ID          string
	Name        string
	Description string
	Command     string
}

// LaravelPermissionsModel represents the Laravel permissions screen
type LaravelPermissionsModel struct {
	theme       *theme.Theme
	width       int
	height      int
	cursor      int
	actions     []LaravelPermAction
	isLaravel   bool
	projectPath string
	webUser     string
	err         error
	success     string
}

// NewLaravelPermissionsModel creates a new Laravel permissions model
func NewLaravelPermissionsModel() LaravelPermissionsModel {
	// Check if current directory is a Laravel project
	cwd, _ := os.Getwd()
	isLaravel := isLaravelProject(cwd)
	
	// Detect web server user
	webUser := detectWebUser()

	actions := []LaravelPermAction{
		{
			ID:          "standard",
			Name:        "Set Standard Permissions",
			Description: "Set 755 for directories, 644 for files (recommended)",
			Command:     fmt.Sprintf("find . -type d -exec chmod 755 {} \\; && find . -type f -exec chmod 644 {} \\;"),
		},
		{
			ID:          "storage_writable",
			Name:        "Make Storage Writable",
			Description: "Set storage & bootstrap/cache writable by web server",
			Command:     fmt.Sprintf("chmod -R 775 storage bootstrap/cache && chown -R $USER:%s storage bootstrap/cache", webUser),
		},
		{
			ID:          "full_reset",
			Name:        "Full Permission Reset",
			Description: "Reset all permissions and set proper ownership",
			Command:     fmt.Sprintf("find . -type d -exec chmod 755 {} \\; && find . -type f -exec chmod 644 {} \\; && chmod -R 775 storage bootstrap/cache && chown -R $USER:%s .", webUser),
		},
		{
			ID:          "storage_777",
			Name:        "Storage 777 (Development Only)",
			Description: "⚠ Set storage to 777 - use only for development",
			Command:     "chmod -R 777 storage bootstrap/cache",
		},
		{
			ID:          "fix_vendor",
			Name:        "Fix Vendor Permissions",
			Description: "Make vendor directory readable",
			Command:     "chmod -R 755 vendor",
		},
		{
			ID:          "secure_env",
			Name:        "Secure .env File",
			Description: "Set .env to 600 (owner read/write only)",
			Command:     "chmod 600 .env",
		},
		{
			ID:          "artisan_executable",
			Name:        "Make Artisan Executable",
			Description: "Set execute permission on artisan",
			Command:     "chmod +x artisan",
		},
		{
			ID:          "clear_cache_files",
			Name:        "Clear Cache Files",
			Description: "Remove compiled views and cache files",
			Command:     "rm -rf storage/framework/cache/data/* storage/framework/views/* storage/framework/sessions/* bootstrap/cache/*.php",
		},
		{
			ID:          "show_permissions",
			Name:        "Show Current Permissions",
			Description: "Display permissions for key directories",
			Command:     "echo '=== Storage ===' && ls -la storage/ && echo '' && echo '=== Bootstrap/Cache ===' && ls -la bootstrap/cache/ && echo '' && echo '=== .env ===' && ls -la .env 2>/dev/null || echo '.env not found'",
		},
		{
			ID:          "back",
			Name:        "← Back to Site Commands",
			Description: "Return to site commands menu",
			Command:     "",
		},
	}

	return LaravelPermissionsModel{
		theme:       theme.DefaultTheme(),
		cursor:      0,
		actions:     actions,
		isLaravel:   isLaravel,
		projectPath: cwd,
		webUser:     webUser,
	}
}

// isLaravelProject checks if the directory contains a Laravel project
func isLaravelProject(path string) bool {
	// Check for artisan file
	if _, err := os.Stat(filepath.Join(path, "artisan")); err == nil {
		// Check for Laravel-specific directories
		storagePath := filepath.Join(path, "storage")
		bootstrapPath := filepath.Join(path, "bootstrap", "cache")
		
		if _, err := os.Stat(storagePath); err == nil {
			if _, err := os.Stat(bootstrapPath); err == nil {
				return true
			}
		}
	}
	return false
}

// detectWebUser tries to detect the web server user
func detectWebUser() string {
	// Common web server users in order of likelihood
	users := []string{"www-data", "nginx", "apache", "http", "nobody"}
	
	// Check if /etc/passwd exists (Linux system)
	if _, err := os.Stat("/etc/passwd"); err == nil {
		// Return first common user (www-data for Debian/Ubuntu)
		return users[0]
	}
	return "www-data"
}

// Init initializes the Laravel permissions screen
func (m LaravelPermissionsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for Laravel permissions
func (m LaravelPermissionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.actions)-1 {
				m.cursor++
			}

		case "enter", " ":
			return m.executeAction()
		}
	}

	return m, nil
}

// executeAction executes the selected permission action
func (m LaravelPermissionsModel) executeAction() (LaravelPermissionsModel, tea.Cmd) {
	m.err = nil
	m.success = ""

	action := m.actions[m.cursor]

	if action.ID == "back" {
		return m, func() tea.Msg {
			return NavigateMsg{Screen: SiteCommandsScreen}
		}
	}

	if action.Command == "" {
		return m, nil
	}

	return m, func() tea.Msg {
		return ExecutionStartMsg{
			Command:     action.Command,
			Description: action.Name,
		}
	}
}

// View renders the Laravel permissions screen
func (m LaravelPermissionsModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	header := m.theme.Title.Render("Laravel Permissions")

	// Project info
	var infoLines []string

	if !m.isLaravel {
		infoLines = append(infoLines, m.theme.WarningStyle.Render("⚠ This doesn't appear to be a Laravel project"))
		infoLines = append(infoLines, m.theme.DescriptionStyle.Render("  Navigate to a Laravel project directory"))
		infoLines = append(infoLines, "")
		infoLines = append(infoLines, m.theme.DescriptionStyle.Render("  Commands can still be run but may not work as expected."))
	} else {
		infoLines = append(infoLines, m.theme.SuccessStyle.Render("✓ Laravel project detected"))
	}
	
	infoLines = append(infoLines, "")
	infoLines = append(infoLines, m.theme.Label.Render("Web User: ")+m.theme.InfoStyle.Render(m.webUser))
	infoLines = append(infoLines, m.theme.Label.Render("Path: ")+m.theme.DescriptionStyle.Render(m.projectPath))

	infoSection := lipgloss.JoinVertical(lipgloss.Left, infoLines...)

	// Info box about permissions
	permInfo := lipgloss.JoinVertical(lipgloss.Left,
		m.theme.Subtitle.Render("Laravel Permission Requirements:"),
		m.theme.DescriptionStyle.Render("  • storage/ - Must be writable by web server"),
		m.theme.DescriptionStyle.Render("  • bootstrap/cache/ - Must be writable by web server"),
		m.theme.DescriptionStyle.Render("  • .env - Should be readable only by owner"),
	)

	// Actions menu
	var actionItems []string
	actionItems = append(actionItems, "")
	actionItems = append(actionItems, m.theme.Subtitle.Render("Permission Actions:"))
	actionItems = append(actionItems, "")

	for i, action := range m.actions {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		var renderedItem string
		if i == m.cursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, action.Name))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, action.Name))
		}

		actionItems = append(actionItems, renderedItem)
		
		// Show description for selected item
		if i == m.cursor {
			actionItems = append(actionItems, "    "+m.theme.DescriptionStyle.Render(action.Description))
		}
	}

	actionsMenu := lipgloss.JoinVertical(lipgloss.Left, actionItems...)

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
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Execute • Esc: Back • q: Quit")

	// Combine all sections
	sections := []string{
		header,
		"",
		infoSection,
		"",
		permInfo,
		actionsMenu,
	}

	if messageSection != "" {
		sections = append(sections, "", messageSection)
	}

	sections = append(sections, "", help)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

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
