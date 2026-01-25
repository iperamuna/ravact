package screens

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// AddUserModel represents the add user screen
type AddUserModel struct {
	theme       *theme.Theme
	width       int
	height      int
	userManager *system.UserManager

	// Form
	form *huh.Form

	// Form fields
	username  string
	password  string
	shell     string
	grantSudo bool

	// UI state
	err       error
	message   string
	submitted bool
}

// NewAddUserModel creates a new add user model
func NewAddUserModel() AddUserModel {
	t := theme.DefaultTheme()

	m := AddUserModel{
		theme:       t,
		userManager: system.NewUserManager(),
		username:    "",
		password:    "",
		shell:       "/bin/bash",
		grantSudo:   false,
	}

	// Create the huh form
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Username").
				Description("Must be 3+ chars, start with letter, lowercase/numbers/_/-").
				Placeholder("Enter username...").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("username cannot be empty")
					}
					if len(s) < 3 {
						return fmt.Errorf("username must be at least 3 characters")
					}
					if matched, _ := regexp.MatchString(`^[a-z][a-z0-9_-]*$`, s); !matched {
						return fmt.Errorf("must start with letter, use lowercase/numbers/_/-")
					}
					return nil
				}).
				Value(&m.username),

			huh.NewInput().
				Title("Password").
				Description("Must be at least 6 characters").
				Placeholder("Enter password...").
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("password cannot be empty")
					}
					if len(s) < 6 {
						return fmt.Errorf("password must be at least 6 characters")
					}
					return nil
				}).
				Value(&m.password),

			huh.NewSelect[string]().
				Title("Shell").
				Description("Default shell for the user").
				Options(
					huh.NewOption("/bin/bash", "/bin/bash"),
					huh.NewOption("/bin/sh", "/bin/sh"),
					huh.NewOption("/bin/zsh", "/bin/zsh"),
					huh.NewOption("/bin/fish", "/bin/fish"),
				).
				Value(&m.shell),

			huh.NewConfirm().
				Title("Grant Sudo Privileges").
				Description("Allow user to run commands as root").
				Affirmative("Yes").
				Negative("No").
				Value(&m.grantSudo),
		),
	).WithTheme(t.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)

	return m
}

// Init initializes the add user screen
func (m AddUserModel) Init() tea.Cmd {
	return m.form.Init()
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
			case "esc", "backspace", "enter", " ":
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

		// Global keys
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.form.State == huh.StateNormal {
				return m, func() tea.Msg {
					return NavigateMsg{Screen: UserManagementScreen}
				}
			}
		}
	}

	// Update the form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	// Check if form is completed
	if m.form.State == huh.StateCompleted {
		if err := m.createUser(); err != nil {
			m.err = err
			// Reset form state to allow retry
			m.form = m.rebuildForm()
		} else {
			m.message = fmt.Sprintf("✓ User '%s' created successfully!\n\nPress any key to return to User Management", m.username)
		}
		return m, nil
	}

	return m, cmd
}

// rebuildForm creates a fresh form instance
func (m *AddUserModel) rebuildForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Username").
				Description("Must be 3+ chars, start with letter, lowercase/numbers/_/-").
				Placeholder("Enter username...").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("username cannot be empty")
					}
					if len(s) < 3 {
						return fmt.Errorf("username must be at least 3 characters")
					}
					if matched, _ := regexp.MatchString(`^[a-z][a-z0-9_-]*$`, s); !matched {
						return fmt.Errorf("must start with letter, use lowercase/numbers/_/-")
					}
					return nil
				}).
				Value(&m.username),

			huh.NewInput().
				Title("Password").
				Description("Must be at least 6 characters").
				Placeholder("Enter password...").
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("password cannot be empty")
					}
					if len(s) < 6 {
						return fmt.Errorf("password must be at least 6 characters")
					}
					return nil
				}).
				Value(&m.password),

			huh.NewSelect[string]().
				Title("Shell").
				Description("Default shell for the user").
				Options(
					huh.NewOption("/bin/bash", "/bin/bash"),
					huh.NewOption("/bin/sh", "/bin/sh"),
					huh.NewOption("/bin/zsh", "/bin/zsh"),
					huh.NewOption("/bin/fish", "/bin/fish"),
				).
				Value(&m.shell),

			huh.NewConfirm().
				Title("Grant Sudo Privileges").
				Description("Allow user to run commands as root").
				Affirmative("Yes").
				Negative("No").
				Value(&m.grantSudo),
		),
	).WithTheme(m.theme.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)
}

// createUser creates the user with the form values
func (m *AddUserModel) createUser() error {
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

	// Warning
	warning := m.theme.WarningStyle.Render(m.theme.Symbols.Warning + " Requires root privileges to create users")

	// Render the huh form
	formView := m.form.View()

	// Help
	help := m.theme.Help.Render("Tab/Shift+Tab: Navigate " + m.theme.Symbols.Bullet + " Enter: Select/Submit " + m.theme.Symbols.Bullet + " Esc: Cancel")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		warning,
		"",
		formView,
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
