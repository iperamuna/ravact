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
}

// NewUserDetailsModel creates a new user details model
func NewUserDetailsModel(user system.User) UserDetailsModel {
	actions := []string{
		"Toggle Sudo Access",
		"Change Shell",
		"Delete User",
	}

	return UserDetailsModel{
		theme:       theme.DefaultTheme(),
		user:        user,
		userManager: system.NewUserManager(),
		cursor:      0,
		actions:     actions,
	}
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
			switch m.cursor {
			case 0: // Toggle Sudo
				action := "grant"
				if m.user.HasSudo {
					action = "revoke"
				}
				err := m.userManager.ToggleSudo(m.user.Username)
				if err != nil {
					m.err = fmt.Errorf("failed to %s sudo: %v", action, err)
				} else {
					m.user.HasSudo = !m.user.HasSudo
					if m.user.HasSudo {
						m.message = fmt.Sprintf("✓ Granted sudo access to %s", m.user.Username)
					} else {
						m.message = fmt.Sprintf("✓ Revoked sudo access from %s", m.user.Username)
					}
				}

			case 1: // Change Shell
				m.message = "Feature coming soon: Shell selection menu"

			case 2: // Delete User
				if m.user.Username == "root" {
					m.err = fmt.Errorf("cannot delete root user")
				} else {
					m.message = fmt.Sprintf("⚠ Delete user '%s'?\nPress 'y' to confirm, any other key to cancel", m.user.Username)
				}
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
