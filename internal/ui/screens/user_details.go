package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// UserDetailsModel represents the user details screen
type UserDetailsModel struct {
	theme       *theme.Theme
	width       int
	height      int
	user        system.User
	userManager *system.UserManager
	cursor      int
	actions     []string
	err         error
	message     string
	confirmAction string // Action waiting for confirmation
}

// NewUserDetailsModel creates a new user details model
func NewUserDetailsModel(user system.User) UserDetailsModel {
	um := system.NewUserManager()
	
	// Build dynamic actions based on current state
	actions := buildUserActions(user, um)

	return UserDetailsModel{
		theme:       theme.DefaultTheme(),
		user:        user,
		userManager: um,
		cursor:      0,
		actions:     actions,
	}
}

// buildUserActions builds the action list based on current state
func buildUserActions(user system.User, um *system.UserManager) []string {
	actions := []string{
		"SSH Key Management",
		"Toggle Sudo Access",
		"Change Shell",
	}

	// SSH Key Login toggle
	if um.IsSSHKeyLoginDisabled(user.Username) {
		actions = append(actions, "Enable SSH Key Login")
	} else {
		actions = append(actions, "Disable SSH Key Login")
	}

	// SSH Password Login toggle (global setting)
	if um.IsPasswordSSHLoginDisabled() {
		actions = append(actions, "Enable SSH Password Login (Global)")
	} else {
		actions = append(actions, "Disable SSH Password Login (Global)")
	}

	actions = append(actions, "Delete User")

	return actions
}

// Init initializes the user details screen
func (m UserDetailsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for user details
func (m UserDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle confirmation dialogs
		if m.confirmAction != "" {
			return m.handleConfirmation(msg)
		}

		// If there's a message, any key clears it
		if m.message != "" {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc", "backspace":
				return m, func() tea.Msg {
					return NavigateMsg{Screen: UserManagementScreen}
				}
			default:
				m.message = ""
				// Rebuild actions in case state changed
				m.actions = buildUserActions(m.user, m.userManager)
				return m, nil
			}
		}

		// If there's an error, clear it first
		if m.err != nil {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc", "backspace":
				return m, func() tea.Msg {
					return NavigateMsg{Screen: UserManagementScreen}
				}
			default:
				m.err = nil
				return m, nil
			}
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc", "backspace":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: UserManagementScreen}
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
			return m.executeAction(m.actions[m.cursor])
		}
	}

	return m, nil
}

// handleConfirmation handles confirmation dialog responses
func (m UserDetailsModel) handleConfirmation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc", "backspace", "n", "N":
		m.confirmAction = ""
		return m, nil
	case "y", "Y":
		action := m.confirmAction
		m.confirmAction = ""
		return m.confirmExecuteAction(action)
	}
	return m, nil
}

// executeAction executes the selected action
func (m UserDetailsModel) executeAction(action string) (tea.Model, tea.Cmd) {
	switch action {
	case "SSH Key Management":
		return m, func() tea.Msg {
			return NavigateMsg{Screen: SSHKeyManagementScreen, Data: m.user.Username}
		}

	case "Toggle Sudo Access":
		actionDesc := "grant"
		if m.user.HasSudo {
			actionDesc = "revoke"
		}
		err := m.userManager.ToggleSudo(m.user.Username)
		if err != nil {
			m.err = fmt.Errorf("failed to %s sudo: %v", actionDesc, err)
		} else {
			m.user.HasSudo = !m.user.HasSudo
			if m.user.HasSudo {
				m.message = fmt.Sprintf("✓ Granted sudo access to %s", m.user.Username)
			} else {
				m.message = fmt.Sprintf("✓ Revoked sudo access from %s", m.user.Username)
			}
		}

	case "Change Shell":
		m.message = "Feature coming soon: Shell selection menu"

	case "Disable SSH Key Login":
		m.confirmAction = action
		m.message = fmt.Sprintf("⚠ Disable SSH key login for '%s'?\n\nThis will rename authorized_keys to authorized_keys.disabled.\nThe user will not be able to login using SSH keys.\n\nPress 'y' to confirm, 'n' or Esc to cancel", m.user.Username)

	case "Enable SSH Key Login":
		err := m.userManager.EnableSSHKeyLogin(m.user.Username)
		if err != nil {
			m.err = fmt.Errorf("failed to enable SSH key login: %v", err)
		} else {
			m.message = fmt.Sprintf("✓ SSH key login enabled for %s", m.user.Username)
			m.actions = buildUserActions(m.user, m.userManager)
		}

	case "Disable SSH Password Login (Global)":
		m.confirmAction = action
		m.message = "⚠ Disable SSH password login globally?\n\nThis will set 'PasswordAuthentication no' in /etc/ssh/sshd_config.\nAll users will be unable to login using passwords via SSH.\nMake sure you have SSH key access configured!\n\nPress 'y' to confirm, 'n' or Esc to cancel"

	case "Enable SSH Password Login (Global)":
		err := m.userManager.EnablePasswordSSHLogin()
		if err != nil {
			m.err = fmt.Errorf("failed to enable SSH password login: %v", err)
		} else {
			m.message = "✓ SSH password login enabled globally"
			m.actions = buildUserActions(m.user, m.userManager)
		}

	case "Delete User":
		if m.user.Username == "root" {
			m.err = fmt.Errorf("cannot delete root user")
		} else {
			m.confirmAction = action
			m.message = fmt.Sprintf("⚠ Delete user '%s'?\n\nThis will remove the user account.\nPress 'y' to confirm, 'n' or Esc to cancel", m.user.Username)
		}
	}

	return m, nil
}

// confirmExecuteAction executes an action after confirmation
func (m UserDetailsModel) confirmExecuteAction(action string) (tea.Model, tea.Cmd) {
	switch action {
	case "Disable SSH Key Login":
		err := m.userManager.DisableSSHKeyLogin(m.user.Username)
		if err != nil {
			m.err = fmt.Errorf("failed to disable SSH key login: %v", err)
		} else {
			m.message = fmt.Sprintf("✓ SSH key login disabled for %s", m.user.Username)
			m.actions = buildUserActions(m.user, m.userManager)
		}

	case "Disable SSH Password Login (Global)":
		err := m.userManager.DisablePasswordSSHLogin(m.user.Username)
		if err != nil {
			m.err = fmt.Errorf("failed to disable SSH password login: %v", err)
		} else {
			m.message = "✓ SSH password login disabled globally\n\n⚠ Make sure you have SSH key access configured!"
			m.actions = buildUserActions(m.user, m.userManager)
		}

	case "Delete User":
		err := m.userManager.DeleteUser(m.user.Username, false)
		if err != nil {
			m.err = fmt.Errorf("failed to delete user: %v", err)
		} else {
			m.message = fmt.Sprintf("✓ User '%s' deleted successfully", m.user.Username)
			return m, func() tea.Msg {
				return NavigateMsg{Screen: UserManagementScreen}
			}
		}
	}

	return m, nil
}

// View renders the user details screen
func (m UserDetailsModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Show error if there's one
	if m.err != nil {
		errorMsg := m.theme.Title.Render("User Details") + "\n\n" +
			m.theme.ErrorStyle.Render("Error:") + "\n" +
			m.theme.DescriptionStyle.Render(m.err.Error()) + "\n\n" +
			m.theme.Help.Render("Press any key to continue • Esc to go back")

		bordered := m.theme.BorderStyle.Render(errorMsg)
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			bordered,
		)
	}

	// Show message if there's one
	if m.message != "" {
		msgStyle := m.theme.InfoStyle
		if strings.Contains(m.message, "✓") {
			msgStyle = m.theme.SuccessStyle
		} else if strings.Contains(m.message, "⚠") {
			msgStyle = m.theme.WarningStyle
		}

		messageDisplay := m.theme.Title.Render("User Details") + "\n\n" +
			msgStyle.Render(m.message) + "\n\n" +
			m.theme.Help.Render("Press any key to continue • Esc to go back")

		bordered := m.theme.BorderStyle.Render(messageDisplay)
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			bordered,
		)
	}

	// Header
	header := m.theme.Title.Render(fmt.Sprintf("User Details: %s", m.user.Username))

	// User information
	infoLines := []string{
		m.theme.Label.Render("Username:   ") + m.theme.MenuItem.Render(m.user.Username),
		m.theme.Label.Render("UID:        ") + m.theme.MenuItem.Render(fmt.Sprintf("%d", m.user.UID)),
		m.theme.Label.Render("GID:        ") + m.theme.MenuItem.Render(fmt.Sprintf("%d", m.user.GID)),
		m.theme.Label.Render("Home:       ") + m.theme.MenuItem.Render(m.user.HomeDir),
		m.theme.Label.Render("Shell:      ") + m.theme.MenuItem.Render(m.user.Shell),
	}

	// Sudo status
	sudoStatus := m.theme.ErrorStyle.Render("✗ No sudo access")
	if m.user.HasSudo {
		sudoStatus = m.theme.SuccessStyle.Render("✓ Has sudo access")
	}
	infoLines = append(infoLines, m.theme.Label.Render("Sudo:       ")+sudoStatus)

	// Groups
	groupsStr := "None"
	if len(m.user.Groups) > 0 {
		groupsStr = strings.Join(m.user.Groups, ", ")
	}
	infoLines = append(infoLines, m.theme.Label.Render("Groups:     ")+m.theme.MenuItem.Render(groupsStr))

	info := lipgloss.JoinVertical(lipgloss.Left, infoLines...)

	// Actions
	var actionItems []string
	actionItems = append(actionItems, m.theme.Label.Render("Actions:"))
	actionItems = append(actionItems, "")

	for i, action := range m.actions {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		var renderedItem string
		if i == m.cursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, action))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, action))
		}

		actionItems = append(actionItems, renderedItem)
	}

	actions := lipgloss.JoinVertical(lipgloss.Left, actionItems...)

	// Help
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Execute Action • Esc: Back • q: Quit")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		info,
		"",
		"",
		actions,
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
