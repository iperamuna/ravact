package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

type mysqlManagementScreen struct {
	manager      *system.MySQLManager
	config       *system.MySQLConfig
	selectedItem int
	menuItems    []string
	err          error
	message      string
}

func NewMySQLManagementScreen() *mysqlManagementScreen {
	return &mysqlManagementScreen{
		manager: system.NewMySQLManager(),
		menuItems: []string{
			"View Current Configuration",
			"Change Root Password",
			"Change Port",
			"Restart MySQL Service",
			"View Service Status",
			"Create Database",
			"List Databases",
			"Back to Main Menu",
		},
	}
}

func (m *mysqlManagementScreen) Init() tea.Cmd {
	// Load current config
	config, err := m.manager.GetConfig()
	if err != nil {
		m.err = err
	} else {
		m.config = config
	}
	return nil
}

func (m *mysqlManagementScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.selectedItem > 0 {
				m.selectedItem--
			}

		case "down", "j":
			if m.selectedItem < len(m.menuItems)-1 {
				m.selectedItem++
			}

		case "enter":
			return m.handleSelection()

		case "esc":
			return NewMainMenuScreen(), nil
		}
	}

	return m, nil
}

func (m *mysqlManagementScreen) View() string {
	var b strings.Builder

	// Header
	b.WriteString(theme.HeaderStyle.Render("🗄️  MySQL Management"))
	b.WriteString("\n\n")

	// Show error if any
	if m.err != nil {
		b.WriteString(theme.ErrorStyle.Render("❌ Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	// Show message if any
	if m.message != "" {
		b.WriteString(theme.SuccessStyle.Render("✅ " + m.message))
		b.WriteString("\n\n")
	}

	// Show current config if loaded
	if m.config != nil {
		b.WriteString(theme.SubtleStyle.Render("Current Configuration:"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  Port: %d\n", m.config.Port))
		b.WriteString(fmt.Sprintf("  Bind Address: %s\n", m.config.BindAddress))
		b.WriteString(fmt.Sprintf("  Config Path: %s\n", m.config.ConfigPath))
		b.WriteString(fmt.Sprintf("  Data Dir: %s\n", m.config.DataDir))
		b.WriteString("\n")
	}

	// Menu items
	for i, item := range m.menuItems {
		cursor := "  "
		if i == m.selectedItem {
			cursor = theme.SelectedStyle.Render("▶ ")
			item = theme.SelectedStyle.Render(item)
		}
		b.WriteString(cursor + item + "\n")
	}

	// Help
	b.WriteString("\n")
	b.WriteString(theme.HelpStyle.Render("↑/↓: navigate • enter: select • esc: back • q: quit"))

	return b.String()
}

func (m *mysqlManagementScreen) handleSelection() (tea.Model, tea.Cmd) {
	m.err = nil
	m.message = ""

	switch m.selectedItem {
	case 0: // View Current Configuration
		config, err := m.manager.GetConfig()
		if err != nil {
			m.err = err
		} else {
			m.config = config
			m.message = "Configuration refreshed"
		}
		return m, nil

	case 1: // Change Root Password
		return NewMySQLPasswordScreen(m.manager), nil

	case 2: // Change Port
		return NewMySQLPortScreen(m.manager, m.config), nil

	case 3: // Restart MySQL Service
		if err := m.manager.RestartService(); err != nil {
			m.err = err
		} else {
			m.message = "MySQL service restarted successfully"
		}
		return m, nil

	case 4: // View Service Status
		status, err := m.manager.GetStatus()
		if err != nil {
			m.err = err
		} else {
			return NewTextDisplayScreen("MySQL Service Status", status, m), nil
		}
		return m, nil

	case 5: // Create Database
		return NewMySQLCreateDBScreen(m.manager), nil

	case 6: // List Databases
		databases, err := m.manager.ListDatabases()
		if err != nil {
			m.err = err
			return m, nil
		}
		content := "Databases:\n\n"
		for _, db := range databases {
			content += "  • " + db + "\n"
		}
		return NewTextDisplayScreen("MySQL Databases", content, m), nil

	case 7: // Back to Main Menu
		return NewMainMenuScreen(), nil
	}

	return m, nil
}

// MySQL Password Change Screen
type mysqlPasswordScreen struct {
	manager   *system.MySQLManager
	textInput textinput.Model
	err       error
}

func NewMySQLPasswordScreen(manager *system.MySQLManager) *mysqlPasswordScreen {
	ti := textinput.New()
	ti.Placeholder = "Enter new root password"
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 50
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

	return &mysqlPasswordScreen{
		manager:   manager,
		textInput: ti,
	}
}

func (m *mysqlPasswordScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m *mysqlPasswordScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewMySQLManagementScreen(), nil
		case "enter":
			password := m.textInput.Value()
			if password == "" {
				m.err = fmt.Errorf("password cannot be empty")
				return m, nil
			}

			if err := m.manager.ChangeRootPassword(password); err != nil {
				m.err = err
				return m, nil
			}

			// Success - return to management screen
			mgmt := NewMySQLManagementScreen()
			mgmt.message = "Root password changed successfully"
			return mgmt, nil
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *mysqlPasswordScreen) View() string {
	var b strings.Builder

	b.WriteString(theme.HeaderStyle.Render("🔐 Change MySQL Root Password"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(theme.ErrorStyle.Render("❌ Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	b.WriteString("Enter the new root password:\n\n")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")
	b.WriteString(theme.HelpStyle.Render("enter: save • esc: cancel • q: quit"))

	return b.String()
}

// MySQL Port Change Screen
type mysqlPortScreen struct {
	manager   *system.MySQLManager
	config    *system.MySQLConfig
	textInput textinput.Model
	err       error
}

func NewMySQLPortScreen(manager *system.MySQLManager, config *system.MySQLConfig) *mysqlPortScreen {
	ti := textinput.New()
	ti.Placeholder = "Enter new port (1024-65535)"
	ti.Focus()
	ti.CharLimit = 5
	ti.Width = 30

	if config != nil {
		ti.SetValue(fmt.Sprintf("%d", config.Port))
	}

	return &mysqlPortScreen{
		manager:   manager,
		config:    config,
		textInput: ti,
	}
}

func (m *mysqlPortScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m *mysqlPortScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewMySQLManagementScreen(), nil
		case "enter":
			portStr := m.textInput.Value()
			var port int
			if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
				m.err = fmt.Errorf("invalid port number")
				return m, nil
			}

			if err := m.manager.ChangePort(port); err != nil {
				m.err = err
				return m, nil
			}

			// Restart service
			if err := m.manager.RestartService(); err != nil {
				m.err = fmt.Errorf("port changed but failed to restart service: %w", err)
				return m, nil
			}

			// Success - return to management screen
			mgmt := NewMySQLManagementScreen()
			mgmt.message = fmt.Sprintf("Port changed to %d and service restarted", port)
			return mgmt, nil
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *mysqlPortScreen) View() string {
	var b strings.Builder

	b.WriteString(theme.HeaderStyle.Render("🔌 Change MySQL Port"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(theme.ErrorStyle.Render("❌ Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	if m.config != nil {
		b.WriteString(fmt.Sprintf("Current port: %d\n\n", m.config.Port))
	}

	b.WriteString("Enter the new port:\n\n")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")
	b.WriteString(theme.SubtleStyle.Render("Note: Service will be restarted after changing port"))
	b.WriteString("\n\n")
	b.WriteString(theme.HelpStyle.Render("enter: save • esc: cancel • q: quit"))

	return b.String()
}

// MySQL Create Database Screen
type mysqlCreateDBScreen struct {
	manager      *system.MySQLManager
	dbNameInput  textinput.Model
	userInput    textinput.Model
	passInput    textinput.Model
	focusIndex   int
	err          error
}

func NewMySQLCreateDBScreen(manager *system.MySQLManager) *mysqlCreateDBScreen {
	dbInput := textinput.New()
	dbInput.Placeholder = "Database name"
	dbInput.Focus()
	dbInput.CharLimit = 64
	dbInput.Width = 40

	userInput := textinput.New()
	userInput.Placeholder = "Username (optional)"
	userInput.CharLimit = 32
	userInput.Width = 40

	passInput := textinput.New()
	passInput.Placeholder = "Password (optional)"
	passInput.CharLimit = 64
	passInput.Width = 40
	passInput.EchoMode = textinput.EchoPassword
	passInput.EchoCharacter = '•'

	return &mysqlCreateDBScreen{
		manager:     manager,
		dbNameInput: dbInput,
		userInput:   userInput,
		passInput:   passInput,
		focusIndex:  0,
	}
}

func (m *mysqlCreateDBScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m *mysqlCreateDBScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewMySQLManagementScreen(), nil
		case "tab", "shift+tab", "up", "down":
			if msg.String() == "up" || msg.String() == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > 2 {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = 2
			}

			m.updateFocus()
			return m, nil

		case "enter":
			dbName := m.dbNameInput.Value()
			if dbName == "" {
				m.err = fmt.Errorf("database name is required")
				return m, nil
			}

			username := m.userInput.Value()
			password := m.passInput.Value()

			if err := m.manager.CreateDatabase(dbName, username, password); err != nil {
				m.err = err
				return m, nil
			}

			// Success
			mgmt := NewMySQLManagementScreen()
			mgmt.message = fmt.Sprintf("Database '%s' created successfully", dbName)
			return mgmt, nil
		}
	}

	// Update focused input
	switch m.focusIndex {
	case 0:
		m.dbNameInput, cmd = m.dbNameInput.Update(msg)
	case 1:
		m.userInput, cmd = m.userInput.Update(msg)
	case 2:
		m.passInput, cmd = m.passInput.Update(msg)
	}

	return m, cmd
}

func (m *mysqlCreateDBScreen) updateFocus() {
	m.dbNameInput.Blur()
	m.userInput.Blur()
	m.passInput.Blur()

	switch m.focusIndex {
	case 0:
		m.dbNameInput.Focus()
	case 1:
		m.userInput.Focus()
	case 2:
		m.passInput.Focus()
	}
}

func (m *mysqlCreateDBScreen) View() string {
	var b strings.Builder

	b.WriteString(theme.HeaderStyle.Render("➕ Create MySQL Database"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(theme.ErrorStyle.Render("❌ Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	b.WriteString("Database Name:\n")
	b.WriteString(m.dbNameInput.View())
	b.WriteString("\n\n")

	b.WriteString("User (optional):\n")
	b.WriteString(m.userInput.View())
	b.WriteString("\n\n")

	b.WriteString("Password (optional):\n")
	b.WriteString(m.passInput.View())
	b.WriteString("\n\n")

	b.WriteString(theme.SubtleStyle.Render("If user is specified, they will be granted full access to the database"))
	b.WriteString("\n\n")
	b.WriteString(theme.HelpStyle.Render("tab: next field • enter: create • esc: cancel • q: quit"))

	return b.String()
}
