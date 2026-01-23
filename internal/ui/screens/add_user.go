package screens

import (
	"fmt"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// AddUserField represents a field in the add user form
type AddUserField int

const (
	UsernameField AddUserField = iota
	PasswordField
	ShellField
	GrantSudoField
	ConfirmField
)

// AddUserModel represents the add user screen
type AddUserModel struct {
	theme       *theme.Theme
	width       int
	height      int
	userManager *system.UserManager
	
	// Form fields
	username    string
	password    string
	shell       string
	grantSudo   bool
	
	// UI state
	cursor      AddUserField
	err         error
	message     string
	
	// Available shells
	shells      []string
	shellCursor int
}

// NewAddUserModel creates a new add user model
func NewAddUserModel() AddUserModel {
	return AddUserModel{
		theme:       theme.DefaultTheme(),
		userManager: system.NewUserManager(),
		username:    "",
		password:    "",
		shell:       "/bin/bash",
		grantSudo:   false,
		cursor:      UsernameField,
		shells:      []string{"/bin/bash", "/bin/sh", "/bin/zsh", "/bin/fish"},
		shellCursor: 0,
	}
}

// Init initializes the add user screen
func (m AddUserModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for add user
func (m AddUserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				// Check if it's success message
				if strings.Contains(m.message, "✓") {
					return m, func() tea.Msg {
						return NavigateMsg{Screen: UserManagementScreen}
					}
				}
				m.message = ""
				return m, nil
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

		case "esc":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: UserManagementScreen}
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j", "tab":
			if m.cursor < ConfirmField {
				m.cursor++
			}

		case "enter", " ":
			switch m.cursor {
			case ShellField:
				// Cycle through shells
				m.shellCursor = (m.shellCursor + 1) % len(m.shells)
				m.shell = m.shells[m.shellCursor]

			case GrantSudoField:
				// Toggle sudo
				m.grantSudo = !m.grantSudo

			case ConfirmField:
				// Validate and create user
				if err := m.validateAndCreateUser(); err != nil {
					m.err = err
				} else {
					m.message = fmt.Sprintf("✓ User '%s' created successfully!\n\nPress any key to return to User Management", m.username)
				}
			}

		case "backspace":
			// Handle backspace for text fields
			switch m.cursor {
			case UsernameField:
				if len(m.username) > 0 {
					m.username = m.username[:len(m.username)-1]
				}
			case PasswordField:
				if len(m.password) > 0 {
					m.password = m.password[:len(m.password)-1]
				}
			}

		default:
			// Handle text input
			if len(msg.String()) == 1 {
				char := msg.String()
				switch m.cursor {
				case UsernameField:
					// Allow lowercase letters, numbers, underscore, hyphen
					if matched, _ := regexp.MatchString(`^[a-z0-9_-]$`, char); matched {
						if len(m.username) < 32 {
							m.username += char
						}
					}
				case PasswordField:
					// Allow any printable character
					if len(m.password) < 64 {
						m.password += char
					}
				}
			}
		}
	}

	return m, nil
}

// validateAndCreateUser validates input and creates the user
func (m *AddUserModel) validateAndCreateUser() error {
	// Validate username
	if m.username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if len(m.username) < 3 {
		return fmt.Errorf("username must be at least 3 characters")
	}
	if matched, _ := regexp.MatchString(`^[a-z][a-z0-9_-]*$`, m.username); !matched {
		return fmt.Errorf("username must start with lowercase letter and contain only lowercase, numbers, _, -")
	}

	// Validate password
	if m.password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	if len(m.password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}

	// Create user
	err := m.userManager.CreateUser(m.username, m.password, m.shell)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	// Grant sudo if requested
	if m.grantSudo {
		err = m.userManager.GrantSudo(m.username)
		if err != nil {
			return fmt.Errorf("user created but failed to grant sudo: %v", err)
		}
	}

	return nil
}

// View renders the add user screen
func (m AddUserModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Show error if there's one
	if m.err != nil {
		errorMsg := m.theme.Title.Render("Add User") + "\n\n" +
			m.theme.ErrorStyle.Render("Error:") + "\n" +
			m.theme.DescriptionStyle.Render(m.err.Error()) + "\n\n" +
			m.theme.Help.Render("Press any key to continue • Esc to cancel")

		bordered := m.theme.BorderStyle.Render(errorMsg)
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			bordered,
		)
	}

	// Show success message if there's one
	if m.message != "" {
		msgStyle := m.theme.SuccessStyle
		if strings.Contains(m.message, "⚠") {
			msgStyle = m.theme.WarningStyle
		}

		messageDisplay := m.theme.Title.Render("Add User") + "\n\n" +
			msgStyle.Render(m.message)

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
	header := m.theme.Title.Render("Add New User")
	
	// Instructions
	instructions := m.theme.DescriptionStyle.Render("Fill in the details to create a new user account")

	// Form fields
	var formItems []string

	// Username field
	usernameLabel := "Username:"
	usernameValue := m.username
	if usernameValue == "" {
		usernameValue = "(type username...)"
	}
	if m.cursor == UsernameField {
		usernameLabel = m.theme.KeyStyle.Render("▶ " + usernameLabel)
		usernameValue = m.theme.SelectedItem.Render(usernameValue + "_")
	} else {
		usernameLabel = "  " + usernameLabel
		usernameValue = m.theme.MenuItem.Render(usernameValue)
	}
	formItems = append(formItems, usernameLabel+" "+usernameValue)
	formItems = append(formItems, m.theme.Help.Render("  Must be 3+ chars, start with letter, use lowercase/numbers/_/-"))
	formItems = append(formItems, "")

	// Password field
	passwordLabel := "Password:"
	passwordValue := strings.Repeat("*", len(m.password))
	if passwordValue == "" {
		passwordValue = "(type password...)"
	}
	if m.cursor == PasswordField {
		passwordLabel = m.theme.KeyStyle.Render("▶ " + passwordLabel)
		passwordValue = m.theme.SelectedItem.Render(passwordValue + "_")
	} else {
		passwordLabel = "  " + passwordLabel
		passwordValue = m.theme.MenuItem.Render(passwordValue)
	}
	formItems = append(formItems, passwordLabel+" "+passwordValue)
	formItems = append(formItems, m.theme.Help.Render("  Must be 6+ characters"))
	formItems = append(formItems, "")

	// Shell field
	shellLabel := "Shell:"
	shellValue := m.shell
	if m.cursor == ShellField {
		shellLabel = m.theme.KeyStyle.Render("▶ " + shellLabel)
		shellValue = m.theme.SelectedItem.Render(shellValue + " (press Enter to change)")
	} else {
		shellLabel = "  " + shellLabel
		shellValue = m.theme.MenuItem.Render(shellValue)
	}
	formItems = append(formItems, shellLabel+" "+shellValue)
	formItems = append(formItems, "")

	// Sudo field
	sudoLabel := "Grant Sudo:"
	sudoValue := "No"
	if m.grantSudo {
		sudoValue = m.theme.SuccessStyle.Render("✓ Yes")
	} else {
		sudoValue = m.theme.MenuItem.Render("✗ No")
	}
	if m.cursor == GrantSudoField {
		sudoLabel = m.theme.KeyStyle.Render("▶ " + sudoLabel)
		sudoValue = m.theme.SelectedItem.Render(sudoValue + " (press Enter to toggle)")
	} else {
		sudoLabel = "  " + sudoLabel
	}
	formItems = append(formItems, sudoLabel+" "+sudoValue)
	formItems = append(formItems, "")

	// Confirm button
	confirmButton := "[ Create User ]"
	if m.cursor == ConfirmField {
		confirmButton = m.theme.SelectedItem.Render("▶ " + confirmButton)
	} else {
		confirmButton = m.theme.MenuItem.Render("  " + confirmButton)
	}
	formItems = append(formItems, "")
	formItems = append(formItems, confirmButton)

	form := lipgloss.JoinVertical(lipgloss.Left, formItems...)

	// Help
	help := m.theme.Help.Render("↑/↓/Tab: Navigate • Type: Enter text • Enter: Select/Toggle • Esc: Cancel • q: Quit")

	// Warning
	warning := m.theme.WarningStyle.Render("⚠ Requires root privileges to create users")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		instructions,
		"",
		"",
		form,
		"",
		"",
		warning,
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
