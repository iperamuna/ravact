package screens

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// GitInfo holds information about the current git repository
type GitInfo struct {
	IsRepo       bool
	RemoteURL    string
	RemoteName   string
	Branch       string
	LastCommit   string
	CommitMsg    string
	HasChanges   bool
	Ahead        int
	Behind       int
}

// GitAction represents a git action menu item
type GitAction struct {
	ID          string
	Name        string
	Description string
}

// GitManagementModel represents the git management screen
type GitManagementModel struct {
	theme       *theme.Theme
	width       int
	height      int
	cursor      int
	actions     []GitAction
	gitInfo     GitInfo
	err         error
	success     string
	inputMode   bool
	inputField  string
	inputValue  string
	inputPrompt string
	remoteType  string // "https" or "ssh"
}

// NewGitManagementModel creates a new git management model
func NewGitManagementModel() GitManagementModel {
	gitInfo := getGitInfo()

	actions := []GitAction{
		{ID: "refresh", Name: "Refresh Git Info", Description: "Refresh repository information"},
		{ID: "test_connection", Name: "Test Git Connection", Description: "Test connection to remote repository"},
		{ID: "add_remote", Name: "Add Remote", Description: "Add a new git remote URL"},
		{ID: "change_remote", Name: "Change Remote URL", Description: "Update the remote URL"},
		{ID: "remove_remote", Name: "Remove Remote", Description: "Remove the git remote"},
		{ID: "git_pull", Name: "Git Pull", Description: "Pull latest changes from remote"},
		{ID: "git_fetch", Name: "Git Fetch", Description: "Fetch changes from remote without merging"},
		{ID: "git_status", Name: "Git Status", Description: "Show detailed git status"},
		{ID: "back", Name: "← Back to Site Commands", Description: "Return to site commands menu"},
	}

	return GitManagementModel{
		theme:   theme.DefaultTheme(),
		cursor:  0,
		actions: actions,
		gitInfo: gitInfo,
	}
}

// getGitInfo retrieves git repository information
func getGitInfo() GitInfo {
	info := GitInfo{}

	// Check if we're in a git repo
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	if err := cmd.Run(); err != nil {
		info.IsRepo = false
		return info
	}
	info.IsRepo = true

	// Get current branch
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	if output, err := cmd.Output(); err == nil {
		info.Branch = strings.TrimSpace(string(output))
	}

	// Get remote name and URL
	cmd = exec.Command("git", "remote")
	if output, err := cmd.Output(); err == nil {
		remotes := strings.Fields(string(output))
		if len(remotes) > 0 {
			info.RemoteName = remotes[0]

			// Get remote URL
			cmd = exec.Command("git", "remote", "get-url", info.RemoteName)
			if urlOutput, err := cmd.Output(); err == nil {
				info.RemoteURL = strings.TrimSpace(string(urlOutput))
			}
		}
	}

	// Get last commit hash (short)
	cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
	if output, err := cmd.Output(); err == nil {
		info.LastCommit = strings.TrimSpace(string(output))
	}

	// Get last commit message
	cmd = exec.Command("git", "log", "-1", "--pretty=%s")
	if output, err := cmd.Output(); err == nil {
		info.CommitMsg = strings.TrimSpace(string(output))
		// Truncate if too long
		if len(info.CommitMsg) > 60 {
			info.CommitMsg = info.CommitMsg[:57] + "..."
		}
	}

	// Check for uncommitted changes
	cmd = exec.Command("git", "status", "--porcelain")
	if output, err := cmd.Output(); err == nil {
		info.HasChanges = len(strings.TrimSpace(string(output))) > 0
	}

	// Get ahead/behind info
	if info.RemoteName != "" && info.Branch != "" {
		cmd = exec.Command("git", "rev-list", "--left-right", "--count", fmt.Sprintf("%s/%s...HEAD", info.RemoteName, info.Branch))
		if output, err := cmd.Output(); err == nil {
			parts := strings.Fields(string(output))
			if len(parts) == 2 {
				fmt.Sscanf(parts[0], "%d", &info.Behind)
				fmt.Sscanf(parts[1], "%d", &info.Ahead)
			}
		}
	}

	return info
}

// Init initializes the git management screen
func (m GitManagementModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for git management
func (m GitManagementModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle input mode
		if m.inputMode {
			switch msg.String() {
			case "enter":
				return m.processInput()
			case "esc":
				m.inputMode = false
				m.inputValue = ""
				m.inputField = ""
				m.inputPrompt = ""
				m.remoteType = ""
				return m, nil
			case "backspace":
				if len(m.inputValue) > 0 {
					m.inputValue = m.inputValue[:len(m.inputValue)-1]
				}
			case "1":
				if m.inputField == "remote_type" {
					m.remoteType = "https"
					m.inputField = "remote_url"
					m.inputPrompt = "Enter HTTPS remote URL (e.g., https://github.com/user/repo.git):"
					m.inputValue = ""
				} else {
					m.inputValue += msg.String()
				}
			case "2":
				if m.inputField == "remote_type" {
					m.remoteType = "ssh"
					m.inputField = "remote_url"
					m.inputPrompt = "Enter SSH remote URL (e.g., git@github.com:user/repo.git):"
					m.inputValue = ""
				} else {
					m.inputValue += msg.String()
				}
			default:
				if m.inputField != "remote_type" {
					char := msg.String()
					if len(char) == 1 {
						m.inputValue += char
					}
				}
			}
			return m, nil
		}

		// Normal mode
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

// processInput processes the user input for git operations
func (m GitManagementModel) processInput() (GitManagementModel, tea.Cmd) {
	m.err = nil
	m.success = ""

	url := strings.TrimSpace(m.inputValue)

	if m.inputField == "remote_url" && url != "" {
		remoteName := "origin"
		if m.gitInfo.RemoteName != "" {
			remoteName = m.gitInfo.RemoteName
		}

		var cmd *exec.Cmd
		if m.gitInfo.RemoteURL == "" {
			// Add new remote
			cmd = exec.Command("git", "remote", "add", remoteName, url)
		} else {
			// Change existing remote
			cmd = exec.Command("git", "remote", "set-url", remoteName, url)
		}

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			m.err = fmt.Errorf("%s", stderr.String())
		} else {
			if m.gitInfo.RemoteURL == "" {
				m.success = fmt.Sprintf("✓ Remote '%s' added successfully", remoteName)
			} else {
				m.success = fmt.Sprintf("✓ Remote '%s' URL updated successfully", remoteName)
			}
			m.gitInfo = getGitInfo()
		}
	}

	m.inputMode = false
	m.inputValue = ""
	m.inputField = ""
	m.inputPrompt = ""
	m.remoteType = ""

	return m, nil
}

// executeAction executes the selected git action
func (m GitManagementModel) executeAction() (GitManagementModel, tea.Cmd) {
	m.err = nil
	m.success = ""

	action := m.actions[m.cursor]

	switch action.ID {
	case "refresh":
		m.gitInfo = getGitInfo()
		m.success = "✓ Git info refreshed"

	case "test_connection":
		if m.gitInfo.RemoteURL == "" {
			m.err = fmt.Errorf("no remote configured")
			return m, nil
		}

		// Test git connection
		cmd := exec.Command("git", "ls-remote", "--exit-code", "-h", m.gitInfo.RemoteURL)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			errMsg := strings.TrimSpace(stderr.String())
			if strings.Contains(errMsg, "Permission denied") || strings.Contains(errMsg, "publickey") {
				m.err = fmt.Errorf("SSH connection failed.\n\nTroubleshooting:\n• Check SSH key is added: ssh-add -l\n• Verify key in GitHub/GitLab settings\n• Test SSH: ssh -T git@github.com\n• Check ~/.ssh/config file")
			} else if strings.Contains(errMsg, "Could not resolve host") {
				m.err = fmt.Errorf("could not resolve host. Check your internet connection")
			} else if errMsg != "" {
				m.err = fmt.Errorf("%s", errMsg)
			} else {
				m.err = fmt.Errorf("connection failed: %v", err)
			}
		} else {
			m.success = "✓ Git connection successful! Repository is accessible."
		}

	case "add_remote", "change_remote":
		if action.ID == "change_remote" && m.gitInfo.RemoteURL == "" {
			m.err = fmt.Errorf("no remote to change. Use 'Add Remote' first")
			return m, nil
		}
		m.inputMode = true
		m.inputField = "remote_type"
		m.inputPrompt = "Select remote type:\n\n  1. HTTPS (username/password or token)\n  2. SSH (key-based authentication)\n\nPress 1 or 2:"
		m.inputValue = ""

	case "remove_remote":
		if m.gitInfo.RemoteName == "" {
			m.err = fmt.Errorf("no remote configured")
			return m, nil
		}

		cmd := exec.Command("git", "remote", "remove", m.gitInfo.RemoteName)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			m.err = fmt.Errorf("%s", stderr.String())
		} else {
			m.success = fmt.Sprintf("✓ Remote '%s' removed", m.gitInfo.RemoteName)
			m.gitInfo = getGitInfo()
		}

	case "git_pull":
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     "git pull",
				Description: "Pulling latest changes",
			}
		}

	case "git_fetch":
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     "git fetch --all",
				Description: "Fetching from all remotes",
			}
		}

	case "git_status":
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     "git status",
				Description: "Git Status",
			}
		}

	case "back":
		return m, func() tea.Msg {
			return NavigateMsg{Screen: SiteCommandsScreen}
		}
	}

	return m, nil
}

// View renders the git management screen
func (m GitManagementModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	header := m.theme.Title.Render("Git Operations")

	// Git repository info
	var infoLines []string

	if !m.gitInfo.IsRepo {
		infoLines = append(infoLines, m.theme.WarningStyle.Render("⚠ Not a Git repository"))
		infoLines = append(infoLines, m.theme.DescriptionStyle.Render("  Navigate to a directory with a Git repository"))
	} else {
		// Branch
		branchLabel := m.theme.Label.Render("Branch: ")
		branchValue := m.theme.SuccessStyle.Render(m.gitInfo.Branch)
		infoLines = append(infoLines, branchLabel+branchValue)

		// Remote
		remoteLabel := m.theme.Label.Render("Remote: ")
		if m.gitInfo.RemoteURL != "" {
			remoteValue := m.theme.InfoStyle.Render(fmt.Sprintf("%s (%s)", m.gitInfo.RemoteURL, m.gitInfo.RemoteName))
			infoLines = append(infoLines, remoteLabel+remoteValue)
		} else {
			infoLines = append(infoLines, remoteLabel+m.theme.WarningStyle.Render("No remote configured"))
		}

		// Last commit
		if m.gitInfo.LastCommit != "" {
			commitLabel := m.theme.Label.Render("Last Commit: ")
			commitValue := m.theme.KeyStyle.Render(m.gitInfo.LastCommit) + " " + m.theme.DescriptionStyle.Render(m.gitInfo.CommitMsg)
			infoLines = append(infoLines, commitLabel+commitValue)
		}

		// Status indicators
		var statusParts []string
		if m.gitInfo.HasChanges {
			statusParts = append(statusParts, m.theme.WarningStyle.Render("● Uncommitted changes"))
		}
		if m.gitInfo.Ahead > 0 {
			statusParts = append(statusParts, m.theme.SuccessStyle.Render(fmt.Sprintf("↑ %d ahead", m.gitInfo.Ahead)))
		}
		if m.gitInfo.Behind > 0 {
			statusParts = append(statusParts, m.theme.ErrorStyle.Render(fmt.Sprintf("↓ %d behind", m.gitInfo.Behind)))
		}
		if len(statusParts) > 0 {
			infoLines = append(infoLines, m.theme.Label.Render("Status: ")+strings.Join(statusParts, " • "))
		} else if !m.gitInfo.HasChanges {
			infoLines = append(infoLines, m.theme.Label.Render("Status: ")+m.theme.SuccessStyle.Render("✓ Clean working tree"))
		}
	}

	infoSection := lipgloss.JoinVertical(lipgloss.Left, infoLines...)

	// Input mode display
	var inputSection string
	if m.inputMode {
		inputSection = lipgloss.JoinVertical(lipgloss.Left,
			"",
			m.theme.Label.Render(m.inputPrompt),
		)
		if m.inputField == "remote_url" {
			inputSection = lipgloss.JoinVertical(lipgloss.Left,
				"",
				m.theme.Label.Render(m.inputPrompt),
				m.theme.SelectedItem.Render(fmt.Sprintf("> %s_", m.inputValue)),
				m.theme.DescriptionStyle.Render("Press Enter to confirm, Esc to cancel"),
			)
		}
	}

	// Actions menu
	var actionItems []string
	actionItems = append(actionItems, m.theme.Subtitle.Render("Actions:"))
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
	var help string
	if m.inputMode {
		if m.inputField == "remote_type" {
			help = m.theme.Help.Render("1: HTTPS • 2: SSH • Esc: Cancel")
		} else {
			help = m.theme.Help.Render("Enter: Confirm • Esc: Cancel")
		}
	} else {
		help = m.theme.Help.Render("↑/↓: Navigate • Enter: Execute • Esc: Back • q: Quit")
	}

	// Combine all sections
	sections := []string{
		header,
		"",
		infoSection,
	}

	if inputSection != "" {
		sections = append(sections, inputSection)
	}

	sections = append(sections, "", actionsMenu)

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
