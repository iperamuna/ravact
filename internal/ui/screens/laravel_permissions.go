package screens

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	systemUser  string // from git config meta.systemuser
	err         error
	success     string
	
	// .env creation state
	envState       string // "", "select_env", "confirm_key"
	envType        string // "local", "staging", "production"
	envCursor      int
}

// NewLaravelPermissionsModel creates a new Laravel permissions model
func NewLaravelPermissionsModel() LaravelPermissionsModel {
	// Check if current directory is a Laravel project
	cwd, _ := os.Getwd()
	isLaravel := isLaravelProject(cwd)
	
	// Detect web server user
	webUser := detectWebUser()
	
	// Get system user from git config
	systemUser := getGitSystemUser()

	// Use system user if available, otherwise fall back to $USER
	ownerUser := "$USER"
	if systemUser != "" {
		ownerUser = systemUser
	}

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
			Command:     fmt.Sprintf("chmod -R 775 storage bootstrap/cache && chown -R %s:%s storage bootstrap/cache", ownerUser, webUser),
		},
		{
			ID:          "full_reset",
			Name:        "Full Permission Reset",
			Description: "Reset all permissions and set proper ownership",
			Command:     fmt.Sprintf("find . -type d -exec chmod 755 {} \\; && find . -type f -exec chmod 644 {} \\; && chmod -R 775 storage bootstrap/cache && chown -R %s:%s .", ownerUser, webUser),
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
			ID:          "create_env",
			Name:        "Create .env from .env.example",
			Description: "Copy .env.example to .env and optionally generate APP_KEY",
			Command:     "", // Special handling
		},
		{
			ID:          "artisan_migrate",
			Name:        "Artisan Migrate",
			Description: "Run php artisan migrate",
			Command:     "php artisan migrate",
		},
		{
			ID:          "artisan_cache_clear",
			Name:        "Artisan Clear All Caches",
			Description: "Clear config, route, view, and application cache",
			Command:     "php artisan config:clear && php artisan route:clear && php artisan view:clear && php artisan cache:clear && echo '✓ All caches cleared'",
		},
		{
			ID:          "artisan_optimize",
			Name:        "Artisan Optimize",
			Description: "Run php artisan optimize for production",
			Command:     "php artisan optimize",
		},
		{
			ID:          "artisan_key_generate",
			Name:        "Artisan Key Generate",
			Description: "Generate new APP_KEY",
			Command:     "php artisan key:generate",
		},
		{
			ID:          "setup_scheduler",
			Name:        "Setup Laravel Scheduler",
			Description: "Add scheduler cron job for www-data user",
			Command:     "", // Special handling
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
		systemUser:  systemUser,
	}
}

// getGitSystemUser retrieves the meta.systemuser from git config
func getGitSystemUser() string {
	cmd := exec.Command("git", "config", "--get", "meta.systemuser")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
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
		// Handle env selection state
		if m.envState == "select_env" {
			return m.updateEnvSelection(msg)
		}
		if m.envState == "confirm_key" {
			return m.updateKeyConfirm(msg)
		}

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

// updateEnvSelection handles environment type selection
func (m LaravelPermissionsModel) updateEnvSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	envOptions := []string{"local", "staging", "production"}
	
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc", "backspace":
		m.envState = ""
		return m, nil
	case "up", "k":
		if m.envCursor > 0 {
			m.envCursor--
		}
	case "down", "j":
		if m.envCursor < len(envOptions)-1 {
			m.envCursor++
		}
	case "enter", " ":
		m.envType = envOptions[m.envCursor]
		m.envState = "confirm_key"
		m.envCursor = 0
		return m, nil
	}
	return m, nil
}

// updateKeyConfirm handles key generation confirmation
func (m LaravelPermissionsModel) updateKeyConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc", "backspace":
		m.envState = "select_env"
		return m, nil
	case "up", "k":
		if m.envCursor > 0 {
			m.envCursor--
		}
	case "down", "j":
		if m.envCursor < 1 {
			m.envCursor++
		}
	case "enter", " ":
		return m.executeEnvCreation()
	}
	return m, nil
}

// executeEnvCreation creates the .env file
func (m LaravelPermissionsModel) executeEnvCreation() (tea.Model, tea.Cmd) {
	generateKey := m.envCursor == 0 // Yes is first option
	
	var command string
	if generateKey {
		command = fmt.Sprintf(`cp .env.example .env && sed -i 's/APP_ENV=.*/APP_ENV=%s/' .env && echo '✓ Created .env with APP_ENV=%s' && php artisan key:generate`, m.envType, m.envType)
	} else {
		command = fmt.Sprintf(`cp .env.example .env && sed -i 's/APP_ENV=.*/APP_ENV=%s/' .env && echo '✓ Created .env with APP_ENV=%s'`, m.envType, m.envType)
	}
	
	m.envState = ""
	
	return m, func() tea.Msg {
		return ExecutionStartMsg{
			Command:     command,
			Description: fmt.Sprintf("Creating .env (%s)", m.envType),
		}
	}
}

// setupScheduler adds the Laravel scheduler cron job to www-data's crontab
func (m LaravelPermissionsModel) setupScheduler() (tea.Model, tea.Cmd) {
	projectPath := m.projectPath
	
	// Command to:
	// 1. Check if cron entry already exists
	// 2. If not, add it to www-data's crontab
	command := fmt.Sprintf(`#!/bin/bash
set -e

PROJECT_PATH="%s"
CRON_ENTRY="* * * * * cd ${PROJECT_PATH} && php artisan schedule:run >> /dev/null 2>&1"
CRON_USER="www-data"

echo "Setting up Laravel Scheduler for: ${PROJECT_PATH}"
echo "Cron user: ${CRON_USER}"
echo ""

# Check if the cron entry already exists
EXISTING=$(crontab -u ${CRON_USER} -l 2>/dev/null | grep -F "${PROJECT_PATH}" | grep -F "schedule:run" || true)

if [ -n "$EXISTING" ]; then
    echo "⚠ Scheduler cron job already exists for this project:"
    echo "  $EXISTING"
    echo ""
    echo "No changes made."
else
    # Add the cron entry to www-data's crontab
    (crontab -u ${CRON_USER} -l 2>/dev/null || true; echo "${CRON_ENTRY}") | crontab -u ${CRON_USER} -
    
    echo "✓ Laravel scheduler cron job added successfully!"
    echo ""
    echo "Cron entry:"
    echo "  ${CRON_ENTRY}"
    echo ""
    echo "Current www-data crontab:"
    crontab -u ${CRON_USER} -l 2>/dev/null || echo "  (empty)"
fi

echo ""
echo "To verify scheduler is working, run:"
echo "  php artisan schedule:list"
`, projectPath)

	return m, func() tea.Msg {
		return ExecutionStartMsg{
			Command:     command,
			Description: "Setup Laravel Scheduler (www-data crontab)",
		}
	}
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

	// Handle .env creation specially
	if action.ID == "create_env" {
		// Check if .env.example exists
		if _, err := os.Stat(filepath.Join(m.projectPath, ".env.example")); os.IsNotExist(err) {
			m.err = fmt.Errorf(".env.example not found in this directory")
			return m, nil
		}
		// Check if .env already exists
		if _, err := os.Stat(filepath.Join(m.projectPath, ".env")); err == nil {
			m.err = fmt.Errorf(".env already exists. Delete it first if you want to recreate it")
			return m, nil
		}
		m.envState = "select_env"
		m.envCursor = 0
		return m, nil
	}

	// Handle scheduler setup
	if action.ID == "setup_scheduler" {
		// Check if artisan exists (Laravel project)
		if _, err := os.Stat(filepath.Join(m.projectPath, "artisan")); os.IsNotExist(err) {
			m.err = fmt.Errorf("artisan not found - not a Laravel project")
			return m, nil
		}
		model, cmd := m.setupScheduler()
		return model.(LaravelPermissionsModel), cmd
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

	// Handle env selection states
	if m.envState == "select_env" {
		return m.viewEnvSelection()
	}
	if m.envState == "confirm_key" {
		return m.viewKeyConfirm()
	}

	// Header
	header := m.theme.Title.Render("Laravel App")

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
	if m.systemUser != "" {
		infoLines = append(infoLines, m.theme.Label.Render("Owner User: ")+m.theme.SuccessStyle.Render(m.systemUser)+" (from git config)")
	} else {
		infoLines = append(infoLines, m.theme.Label.Render("Owner User: ")+m.theme.WarningStyle.Render("$USER")+" (set via Git → Set System User)")
	}
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

// viewEnvSelection renders the environment selection screen
func (m LaravelPermissionsModel) viewEnvSelection() string {
	header := m.theme.Title.Render("Create .env - Select Environment")

	description := m.theme.DescriptionStyle.Render("Select the environment type for your .env file:")

	envOptions := []string{"local", "staging", "production"}
	envDescriptions := []string{
		"Development environment with debug enabled",
		"Staging/testing environment",
		"Production environment with optimizations",
	}

	var items []string
	items = append(items, "")
	for i, opt := range envOptions {
		cursor := "  "
		if i == m.envCursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		var renderedItem string
		if i == m.envCursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, opt))
			items = append(items, renderedItem)
			items = append(items, "    "+m.theme.DescriptionStyle.Render(envDescriptions[i]))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, opt))
			items = append(items, renderedItem)
		}
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, items...)

	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Select • Esc: Cancel")

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", description, menu, "", help)
	bordered := m.theme.BorderStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

// viewKeyConfirm renders the key generation confirmation screen
func (m LaravelPermissionsModel) viewKeyConfirm() string {
	header := m.theme.Title.Render("Create .env - Generate APP_KEY?")

	info := lipgloss.JoinVertical(lipgloss.Left,
		m.theme.Label.Render("Environment: ")+m.theme.InfoStyle.Render(m.envType),
		"",
		m.theme.DescriptionStyle.Render("Do you want to generate a new APP_KEY after creating .env?"),
	)

	options := []string{"Yes, generate new APP_KEY", "No, I'll set it manually"}

	var items []string
	items = append(items, "")
	for i, opt := range options {
		cursor := "  "
		if i == m.envCursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		var renderedItem string
		if i == m.envCursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, opt))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, opt))
		}
		items = append(items, renderedItem)
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, items...)

	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Confirm • Esc: Back")

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", info, menu, "", help)
	bordered := m.theme.BorderStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}
