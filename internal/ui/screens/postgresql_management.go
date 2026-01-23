package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

type postgresqlManagementScreen struct {
	manager      *system.PostgreSQLManager
	config       *system.PostgreSQLConfig
	selectedItem int
	menuItems    []string
	err          error
	message      string
}

func NewPostgreSQLManagementScreen() *postgresqlManagementScreen {
	return &postgresqlManagementScreen{
		manager: system.NewPostgreSQLManager(),
		menuItems: []string{
			"View Current Configuration",
			"Change Postgres Password",
			"Change Port",
			"Update Max Connections",
			"Update Shared Buffers",
			"Restart PostgreSQL Service",
			"View Service Status",
			"Create Database",
			"List Databases",
			"Back to Main Menu",
		},
	}
}

func (m *postgresqlManagementScreen) Init() tea.Cmd {
	// Load current config
	config, err := m.manager.GetConfig()
	if err != nil {
		m.err = err
	} else {
		m.config = config
	}
	return nil
}

func (m *postgresqlManagementScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *postgresqlManagementScreen) View() string {
	var b strings.Builder

	// Header
	b.WriteString(theme.HeaderStyle.Render("🐘 PostgreSQL Management"))
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
		b.WriteString(fmt.Sprintf("  Max Connections: %d\n", m.config.MaxConn))
		b.WriteString(fmt.Sprintf("  Shared Buffers: %s\n", m.config.SharedBuf))
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

func (m *postgresqlManagementScreen) handleSelection() (tea.Model, tea.Cmd) {
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

	case 1: // Change Postgres Password
		return NewPostgreSQLPasswordScreen(m.manager), nil

	case 2: // Change Port
		return NewPostgreSQLPortScreen(m.manager, m.config), nil

	case 3: // Update Max Connections
		return NewPostgreSQLMaxConnScreen(m.manager, m.config), nil

	case 4: // Update Shared Buffers
		return NewPostgreSQLSharedBufScreen(m.manager, m.config), nil

	case 5: // Restart PostgreSQL Service
		if err := m.manager.RestartService(); err != nil {
			m.err = err
		} else {
			m.message = "PostgreSQL service restarted successfully"
		}
		return m, nil

	case 6: // View Service Status
		status, err := m.manager.GetStatus()
		if err != nil {
			m.err = err
		} else {
			return NewTextDisplayScreen("PostgreSQL Service Status", status, m), nil
		}
		return m, nil

	case 7: // Create Database
		return NewPostgreSQLCreateDBScreen(m.manager), nil

	case 8: // List Databases
		databases, err := m.manager.ListDatabases()
		if err != nil {
			m.err = err
			return m, nil
		}
		content := "Databases:\n\n"
		for _, db := range databases {
			content += "  • " + db + "\n"
		}
		return NewTextDisplayScreen("PostgreSQL Databases", content, m), nil

	case 9: // Back to Main Menu
		return NewMainMenuScreen(), nil
	}

	return m, nil
}

// PostgreSQL Password Change Screen
type postgresqlPasswordScreen struct {
	manager   *system.PostgreSQLManager
	textInput textinput.Model
	err       error
}

func NewPostgreSQLPasswordScreen(manager *system.PostgreSQLManager) *postgresqlPasswordScreen {
	ti := textinput.New()
	ti.Placeholder = "Enter new postgres user password"
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 50
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

	return &postgresqlPasswordScreen{
		manager:   manager,
		textInput: ti,
	}
}

func (m *postgresqlPasswordScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m *postgresqlPasswordScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewPostgreSQLManagementScreen(), nil
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
			mgmt := NewPostgreSQLManagementScreen()
			mgmt.message = "Postgres user password changed successfully"
			return mgmt, nil
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *postgresqlPasswordScreen) View() string {
	var b strings.Builder

	b.WriteString(theme.HeaderStyle.Render("🔐 Change PostgreSQL Password"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(theme.ErrorStyle.Render("❌ Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	b.WriteString("Enter the new postgres user password:\n\n")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")
	b.WriteString(theme.HelpStyle.Render("enter: save • esc: cancel • q: quit"))

	return b.String()
}

// PostgreSQL Port Change Screen
type postgresqlPortScreen struct {
	manager   *system.PostgreSQLManager
	config    *system.PostgreSQLConfig
	textInput textinput.Model
	err       error
}

func NewPostgreSQLPortScreen(manager *system.PostgreSQLManager, config *system.PostgreSQLConfig) *postgresqlPortScreen {
	ti := textinput.New()
	ti.Placeholder = "Enter new port (1024-65535)"
	ti.Focus()
	ti.CharLimit = 5
	ti.Width = 30

	if config != nil {
		ti.SetValue(fmt.Sprintf("%d", config.Port))
	}

	return &postgresqlPortScreen{
		manager:   manager,
		config:    config,
		textInput: ti,
	}
}

func (m *postgresqlPortScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m *postgresqlPortScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewPostgreSQLManagementScreen(), nil
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
			mgmt := NewPostgreSQLManagementScreen()
			mgmt.message = fmt.Sprintf("Port changed to %d and service restarted", port)
			return mgmt, nil
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *postgresqlPortScreen) View() string {
	var b strings.Builder

	b.WriteString(theme.HeaderStyle.Render("🔌 Change PostgreSQL Port"))
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

// PostgreSQL Max Connections Screen
type postgresqlMaxConnScreen struct {
	manager   *system.PostgreSQLManager
	config    *system.PostgreSQLConfig
	textInput textinput.Model
	err       error
}

func NewPostgreSQLMaxConnScreen(manager *system.PostgreSQLManager, config *system.PostgreSQLConfig) *postgresqlMaxConnScreen {
	ti := textinput.New()
	ti.Placeholder = "Enter max connections (10-10000)"
	ti.Focus()
	ti.CharLimit = 5
	ti.Width = 30

	if config != nil {
		ti.SetValue(fmt.Sprintf("%d", config.MaxConn))
	}

	return &postgresqlMaxConnScreen{
		manager:   manager,
		config:    config,
		textInput: ti,
	}
}

func (m *postgresqlMaxConnScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m *postgresqlMaxConnScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewPostgreSQLManagementScreen(), nil
		case "enter":
			connStr := m.textInput.Value()
			var maxConn int
			if _, err := fmt.Sscanf(connStr, "%d", &maxConn); err != nil {
				m.err = fmt.Errorf("invalid number")
				return m, nil
			}

			if err := m.manager.UpdateMaxConnections(maxConn); err != nil {
				m.err = err
				return m, nil
			}

			// Restart service
			if err := m.manager.RestartService(); err != nil {
				m.err = fmt.Errorf("max_connections changed but failed to restart service: %w", err)
				return m, nil
			}

			// Success
			mgmt := NewPostgreSQLManagementScreen()
			mgmt.message = fmt.Sprintf("Max connections changed to %d and service restarted", maxConn)
			return mgmt, nil
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *postgresqlMaxConnScreen) View() string {
	var b strings.Builder

	b.WriteString(theme.HeaderStyle.Render("🔗 Update Max Connections"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(theme.ErrorStyle.Render("❌ Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	if m.config != nil {
		b.WriteString(fmt.Sprintf("Current max_connections: %d\n\n", m.config.MaxConn))
	}

	b.WriteString("Enter new max_connections value:\n\n")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")
	b.WriteString(theme.SubtleStyle.Render("Note: Service will be restarted after change"))
	b.WriteString("\n\n")
	b.WriteString(theme.HelpStyle.Render("enter: save • esc: cancel • q: quit"))

	return b.String()
}

// PostgreSQL Shared Buffers Screen
type postgresqlSharedBufScreen struct {
	manager   *system.PostgreSQLManager
	config    *system.PostgreSQLConfig
	textInput textinput.Model
	err       error
}

func NewPostgreSQLSharedBufScreen(manager *system.PostgreSQLManager, config *system.PostgreSQLConfig) *postgresqlSharedBufScreen {
	ti := textinput.New()
	ti.Placeholder = "e.g., 256MB, 1GB"
	ti.Focus()
	ti.CharLimit = 10
	ti.Width = 30

	if config != nil {
		ti.SetValue(config.SharedBuf)
	}

	return &postgresqlSharedBufScreen{
		manager:   manager,
		config:    config,
		textInput: ti,
	}
}

func (m *postgresqlSharedBufScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m *postgresqlSharedBufScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewPostgreSQLManagementScreen(), nil
		case "enter":
			sharedBuf := m.textInput.Value()
			if sharedBuf == "" {
				m.err = fmt.Errorf("value cannot be empty")
				return m, nil
			}

			if err := m.manager.UpdateSharedBuffers(sharedBuf); err != nil {
				m.err = err
				return m, nil
			}

			// Restart service
			if err := m.manager.RestartService(); err != nil {
				m.err = fmt.Errorf("shared_buffers changed but failed to restart service: %w", err)
				return m, nil
			}

			// Success
			mgmt := NewPostgreSQLManagementScreen()
			mgmt.message = fmt.Sprintf("Shared buffers changed to %s and service restarted", sharedBuf)
			return mgmt, nil
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *postgresqlSharedBufScreen) View() string {
	var b strings.Builder

	b.WriteString(theme.HeaderStyle.Render("💾 Update Shared Buffers"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(theme.ErrorStyle.Render("❌ Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	if m.config != nil {
		b.WriteString(fmt.Sprintf("Current shared_buffers: %s\n\n", m.config.SharedBuf))
	}

	b.WriteString("Enter new shared_buffers value (e.g., 256MB, 1GB):\n\n")
	b.WriteString(m.textInput.View())
	b.WriteString("\n\n")
	b.WriteString(theme.SubtleStyle.Render("Note: Service will be restarted after change"))
	b.WriteString("\n\n")
	b.WriteString(theme.HelpStyle.Render("enter: save • esc: cancel • q: quit"))

	return b.String()
}

// PostgreSQL Create Database Screen
type postgresqlCreateDBScreen struct {
	manager     *system.PostgreSQLManager
	dbNameInput textinput.Model
	userInput   textinput.Model
	passInput   textinput.Model
	focusIndex  int
	err         error
}

func NewPostgreSQLCreateDBScreen(manager *system.PostgreSQLManager) *postgresqlCreateDBScreen {
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

	return &postgresqlCreateDBScreen{
		manager:     manager,
		dbNameInput: dbInput,
		userInput:   userInput,
		passInput:   passInput,
		focusIndex:  0,
	}
}

func (m *postgresqlCreateDBScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m *postgresqlCreateDBScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewPostgreSQLManagementScreen(), nil
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
			mgmt := NewPostgreSQLManagementScreen()
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

func (m *postgresqlCreateDBScreen) updateFocus() {
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

func (m *postgresqlCreateDBScreen) View() string {
	var b strings.Builder

	b.WriteString(theme.HeaderStyle.Render("➕ Create PostgreSQL Database"))
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
