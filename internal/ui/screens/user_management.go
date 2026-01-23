package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// ViewMode represents which view is active
type ViewMode int

const (
	UsersView ViewMode = iota
	GroupsView
)

// UserManagementModel represents the user management screen
type UserManagementModel struct {
	theme       *theme.Theme
	width       int
	height      int
	userManager *system.UserManager
	users       []system.User
	groups      []system.Group
	cursor      int
	viewMode    ViewMode
	scrollOffset int
	maxVisible   int
	err         error
	loading     bool
}

// UsersLoadedMsg is sent when users are loaded
type UsersLoadedMsg struct {
	users  []system.User
	groups []system.Group
	err    error
}

// NewUserManagementModel creates a new user management model
func NewUserManagementModel() UserManagementModel {
	return UserManagementModel{
		theme:        theme.DefaultTheme(),
		userManager:  system.NewUserManager(),
		users:        []system.User{},
		groups:       []system.Group{},
		cursor:       0,
		viewMode:     UsersView,
		scrollOffset: 0,
		maxVisible:   10, // Show max 10 items at once
		loading:      true,
	}
}

// loadUsersCmd loads users asynchronously
func (m UserManagementModel) loadUsersCmd() tea.Msg {
	users, err := m.userManager.GetAllUsers()
	if err != nil {
		return UsersLoadedMsg{users: []system.User{}, groups: []system.Group{}, err: err}
	}
	
	groups, err := m.userManager.GetAllGroups()
	if err != nil {
		return UsersLoadedMsg{users: users, groups: []system.Group{}, err: err}
	}
	
	return UsersLoadedMsg{users: users, groups: groups, err: nil}
}

// Init initializes the user management screen
func (m UserManagementModel) Init() tea.Cmd {
	// Load users asynchronously to avoid blocking the UI
	// Wrap in a function so it runs in the background
	return func() tea.Msg {
		return m.loadUsersCmd()
	}
}

// Update handles messages for user management
func (m UserManagementModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case UsersLoadedMsg:
		m.loading = false
		m.users = msg.users
		m.groups = msg.groups
		m.err = msg.err
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// If there's an error showing, any key clears it
		if m.err != nil {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc", "backspace":
				return m, func() tea.Msg {
					return NavigateMsg{Screen: MainMenuScreen}
				}
			case "r":
				// Allow refresh even with error
				m.loading = true
				m.err = nil
				m.cursor = 0
				m.scrollOffset = 0
				return m, func() tea.Msg {
					return m.loadUsersCmd()
				}
			default:
				// Any other key clears the error
				m.err = nil
				return m, nil
			}
		}
		
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc", "backspace":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: MainMenuScreen}
			}

		case "tab":
			// Switch between users and groups view
			if m.viewMode == UsersView {
				m.viewMode = GroupsView
				m.cursor = 0
				m.scrollOffset = 0
			} else {
				m.viewMode = UsersView
				m.cursor = 0
				m.scrollOffset = 0
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// Adjust scroll offset if needed
				if m.cursor < m.scrollOffset {
					m.scrollOffset = m.cursor
				}
			}

		case "down", "j":
			maxCursor := 0
			if m.viewMode == UsersView {
				maxCursor = len(m.users) - 1
			} else {
				maxCursor = len(m.groups) - 1
			}
			if m.cursor < maxCursor {
				m.cursor++
				// Adjust scroll offset if needed
				if m.cursor >= m.scrollOffset+m.maxVisible {
					m.scrollOffset = m.cursor - m.maxVisible + 1
				}
			}

		case "r":
			// Refresh data asynchronously
			m.loading = true
			m.err = nil  // Clear any errors
			m.cursor = 0
			m.scrollOffset = 0
			return m, func() tea.Msg {
				return m.loadUsersCmd()
			}

		case "a":
			// Add user or group
			if m.viewMode == UsersView {
				return m, func() tea.Msg {
					return NavigateMsg{Screen: AddUserScreen}
				}
			} else {
				// Add group - not implemented yet
				m.err = fmt.Errorf("add group feature not implemented yet")
			}
			return m, nil

		case "enter", " ":
			// View/edit user or group details
			if m.viewMode == UsersView && len(m.users) > 0 {
				selectedUser := m.users[m.cursor]
				return m, func() tea.Msg {
					return NavigateMsg{
						Screen: UserDetailsScreen,
						Data: map[string]interface{}{
							"user": selectedUser,
						},
					}
				}
			} else if m.viewMode == GroupsView && len(m.groups) > 0 {
				selectedGroup := m.groups[m.cursor]
				// Group details not implemented yet
				m.err = fmt.Errorf("group details screen not implemented yet\nSelected: %s (GID: %d)", 
					selectedGroup.Name, selectedGroup.GID)
			}
			return m, nil
		}
	}

	return m, nil
}

// View renders the user management screen
func (m UserManagementModel) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	// Show loading state
	if m.loading {
		loadingMsg := m.theme.Title.Render("User Management") + "\n\n" +
			m.theme.InfoStyle.Render("Loading users and groups...") + "\n\n" +
			m.theme.Help.Render("Please wait...")
		
		bordered := m.theme.BorderStyle.Render(loadingMsg)
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			bordered,
		)
	}

	// Show error if there's an error
	if m.err != nil {
		errorTitle := "Error"
		helpText := "Press r to retry • Esc to go back • q to quit"
		
		// Check if it's a feature not implemented error
		errStr := m.err.Error()
		if strings.Contains(errStr, "not implemented") {
			errorTitle = "Feature Not Available"
			helpText = "Press any key to continue • Esc to go back"
		}
		
		errorMsg := m.theme.Title.Render("User Management") + "\n\n" +
			m.theme.WarningStyle.Render(errorTitle + ":") + "\n" +
			m.theme.DescriptionStyle.Render(m.err.Error()) + "\n\n" +
			m.theme.Help.Render(helpText)
		
		bordered := m.theme.BorderStyle.Render(errorMsg)
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			bordered,
		)
	}

	// Header
	header := m.theme.Title.Render("User Management")

	// Tab selection
	tabUsers := "Users"
	tabGroups := "Groups"
	
	if m.viewMode == UsersView {
		tabUsers = m.theme.SelectedItem.Render("[ Users ]")
		tabGroups = m.theme.MenuItem.Render("  Groups  ")
	} else {
		tabUsers = m.theme.MenuItem.Render("  Users  ")
		tabGroups = m.theme.SelectedItem.Render("[ Groups ]")
	}
	
	tabs := lipgloss.JoinHorizontal(lipgloss.Left, tabUsers, "  ", tabGroups)

	var content string
	if m.viewMode == UsersView {
		content = m.renderUsersView()
	} else {
		content = m.renderGroupsView()
	}

	// Help text
	help := ""
	if m.viewMode == UsersView {
		help = m.theme.Help.Render("↑/↓: Navigate • Enter: Details • a: Add User • r: Refresh • Tab: Switch View • Esc: Back • q: Quit")
	} else {
		help = m.theme.Help.Render("↑/↓: Navigate • Enter: Details • a: Add Group • r: Refresh • Tab: Switch View • Esc: Back • q: Quit")
	}

	// Combine all sections
	fullContent := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		tabs,
		"",
		content,
		"",
		help,
	)

	// Add border and center
	bordered := m.theme.BorderStyle.Render(fullContent)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}

// renderUsersView renders the users table
func (m UserManagementModel) renderUsersView() string {
	if len(m.users) == 0 {
		return m.theme.WarningStyle.Render("No users found")
	}

	// Summary
	totalUsers := len(m.users)
	sudoUsers := 0
	for _, user := range m.users {
		if user.HasSudo {
			sudoUsers++
		}
	}
	summary := m.theme.InfoStyle.Render(fmt.Sprintf("Total Users: %d | Sudo Users: %d", totalUsers, sudoUsers))

	// Table header
	headerStyle := m.theme.Label
	headers := []string{
		headerStyle.Render("Username"),
		headerStyle.Render("UID"),
		headerStyle.Render("Sudo"),
		headerStyle.Render("Groups"),
		headerStyle.Render("Home"),
	}
	headerRow := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Width(20).Render(headers[0]),
		lipgloss.NewStyle().Width(8).Render(headers[1]),
		lipgloss.NewStyle().Width(8).Render(headers[2]),
		lipgloss.NewStyle().Width(30).Render(headers[3]),
		lipgloss.NewStyle().Width(30).Render(headers[4]),
	)

	// Table rows (with pagination)
	var rows []string
	rows = append(rows, headerRow)
	rows = append(rows, strings.Repeat("─", 96))

	// Calculate visible range
	startIdx := m.scrollOffset
	endIdx := m.scrollOffset + m.maxVisible
	if endIdx > len(m.users) {
		endIdx = len(m.users)
	}

	// Show scroll indicators
	if m.scrollOffset > 0 {
		rows = append(rows, m.theme.DescriptionStyle.Render("  ↑ More items above..."))
	}

	for idx := startIdx; idx < endIdx; idx++ {
		user := m.users[idx]
		i := idx
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		// Username
		username := lipgloss.NewStyle().Width(18).Render(user.Username)

		// UID
		uid := lipgloss.NewStyle().Width(8).Render(fmt.Sprintf("%d", user.UID))

		// Sudo badge
		sudoBadge := ""
		if user.HasSudo {
			sudoBadge = m.theme.SuccessStyle.Render("Yes")
		} else {
			sudoBadge = m.theme.DescriptionStyle.Render("No")
		}
		sudoCol := lipgloss.NewStyle().Width(8).Render(sudoBadge)

		// Groups (first 3)
		groupsStr := ""
		if len(user.Groups) > 0 {
			if len(user.Groups) > 3 {
				groupsStr = fmt.Sprintf("%s, %s, %s +%d", user.Groups[0], user.Groups[1], user.Groups[2], len(user.Groups)-3)
			} else {
				groupsStr = strings.Join(user.Groups, ", ")
			}
		}
		if len(groupsStr) > 28 {
			groupsStr = groupsStr[:25] + "..."
		}
		groups := lipgloss.NewStyle().Width(30).Render(groupsStr)

		// Home directory
		homeDir := user.HomeDir
		if len(homeDir) > 28 {
			homeDir = "..." + homeDir[len(homeDir)-25:]
		}
		home := lipgloss.NewStyle().Width(30).Render(homeDir)

		row := lipgloss.JoinHorizontal(
			lipgloss.Left,
			cursor,
			username,
			uid,
			sudoCol,
			groups,
			home,
		)

		if i == m.cursor {
			row = m.theme.SelectedItem.Render(row)
		} else {
			row = m.theme.MenuItem.Render(row)
		}

		rows = append(rows, row)
	}

	// Show scroll indicator at bottom
	if endIdx < len(m.users) {
		rows = append(rows, m.theme.DescriptionStyle.Render("  ↓ More items below..."))
	}

	table := lipgloss.JoinVertical(lipgloss.Left, rows...)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		summary,
		"",
		table,
	)
}

// renderGroupsView renders the groups table
func (m UserManagementModel) renderGroupsView() string {
	if len(m.groups) == 0 {
		return m.theme.WarningStyle.Render("No groups found")
	}

	// Summary
	totalGroups := len(m.groups)
	emptyGroups := 0
	for _, group := range m.groups {
		if len(group.Members) == 0 {
			emptyGroups++
		}
	}
	summary := m.theme.InfoStyle.Render(fmt.Sprintf("Total Groups: %d | Empty Groups: %d", totalGroups, emptyGroups))

	// Table header
	headerStyle := m.theme.Label
	headers := []string{
		headerStyle.Render("Group Name"),
		headerStyle.Render("GID"),
		headerStyle.Render("Members"),
		headerStyle.Render("Member List"),
	}
	headerRow := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Width(25).Render(headers[0]),
		lipgloss.NewStyle().Width(10).Render(headers[1]),
		lipgloss.NewStyle().Width(10).Render(headers[2]),
		lipgloss.NewStyle().Width(50).Render(headers[3]),
	)

	// Table rows (with pagination)
	var rows []string
	rows = append(rows, headerRow)
	rows = append(rows, strings.Repeat("─", 95))

	// Calculate visible range
	startIdx := m.scrollOffset
	endIdx := m.scrollOffset + m.maxVisible
	if endIdx > len(m.groups) {
		endIdx = len(m.groups)
	}

	// Show scroll indicators
	if m.scrollOffset > 0 {
		rows = append(rows, m.theme.DescriptionStyle.Render("  ↑ More items above..."))
	}

	for idx := startIdx; idx < endIdx; idx++ {
		group := m.groups[idx]
		i := idx
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		// Group name
		groupName := lipgloss.NewStyle().Width(23).Render(group.Name)

		// GID
		gid := lipgloss.NewStyle().Width(10).Render(fmt.Sprintf("%d", group.GID))

		// Member count
		memberCount := fmt.Sprintf("%d", len(group.Members))
		if len(group.Members) == 0 {
			memberCount = m.theme.DescriptionStyle.Render("0 (empty)")
		}
		memberCountCol := lipgloss.NewStyle().Width(10).Render(memberCount)

		// Member list (first 5)
		membersStr := ""
		if len(group.Members) > 0 {
			if len(group.Members) > 5 {
				membersStr = fmt.Sprintf("%s +%d more", strings.Join(group.Members[:5], ", "), len(group.Members)-5)
			} else {
				membersStr = strings.Join(group.Members, ", ")
			}
		} else {
			membersStr = m.theme.DescriptionStyle.Render("(no members)")
		}
		if len(membersStr) > 48 {
			membersStr = membersStr[:45] + "..."
		}
		members := lipgloss.NewStyle().Width(50).Render(membersStr)

		row := lipgloss.JoinHorizontal(
			lipgloss.Left,
			cursor,
			groupName,
			gid,
			memberCountCol,
			members,
		)

		if i == m.cursor {
			row = m.theme.SelectedItem.Render(row)
		} else {
			row = m.theme.MenuItem.Render(row)
		}

		rows = append(rows, row)
	}

	// Show scroll indicator at bottom
	if endIdx < len(m.groups) {
		rows = append(rows, m.theme.DescriptionStyle.Render("  ↓ More items below..."))
	}

	table := lipgloss.JoinVertical(lipgloss.Left, rows...)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		summary,
		"",
		table,
	)
}
