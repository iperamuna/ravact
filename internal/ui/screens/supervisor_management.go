package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

type supervisorManagementScreen struct {
	manager      *system.SupervisorManager
	programs     []system.SupervisorProgram
	selectedItem int
	menuItems    []string
	err          error
	message      string
}

func NewSupervisorManagementScreen() *supervisorManagementScreen {
	return &supervisorManagementScreen{
		manager: system.NewSupervisorManager(),
		menuItems: []string{
			"List All Programs",
			"Add New Program",
			"Edit Program",
			"Start Program",
			"Stop Program",
			"Restart Program",
			"Delete Program",
			"Configure XML-RPC",
			"View XML-RPC Config",
			"Restart Supervisor",
			"Back to Main Menu",
		},
	}
}

func (m *supervisorManagementScreen) Init() tea.Cmd {
	// Load programs
	programs, err := m.manager.GetAllPrograms()
	if err != nil {
		m.err = err
	} else {
		m.programs = programs
	}
	return nil
}

func (m *supervisorManagementScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *supervisorManagementScreen) View() string {
	var b strings.Builder

	b.WriteString(theme.HeaderStyle.Render("⚙️  Supervisor Management"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(theme.ErrorStyle.Render("❌ Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	if m.message != "" {
		b.WriteString(theme.SuccessStyle.Render("✅ " + m.message))
		b.WriteString("\n\n")
	}

	b.WriteString(theme.SubtleStyle.Render(fmt.Sprintf("Total Programs: %d", len(m.programs))))
	b.WriteString("\n\n")

	for i, item := range m.menuItems {
		cursor := "  "
		if i == m.selectedItem {
			cursor = theme.SelectedStyle.Render("▶ ")
			item = theme.SelectedStyle.Render(item)
		}
		b.WriteString(cursor + item + "\n")
	}

	b.WriteString("\n")
	b.WriteString(theme.HelpStyle.Render("↑/↓: navigate • enter: select • esc: back • q: quit"))

	return b.String()
}

func (m *supervisorManagementScreen) handleSelection() (tea.Model, tea.Cmd) {
	m.err = nil
	m.message = ""

	switch m.selectedItem {
	case 0: // List All Programs
		programs, err := m.manager.GetAllPrograms()
		if err != nil {
			m.err = err
			return m, nil
		}
		m.programs = programs
		return NewSupervisorProgramListScreen(m.manager, programs), nil

	case 1: // Add New Program
		return NewSupervisorAddProgramScreen(m.manager), nil

	case 2: // Edit Program
		if len(m.programs) == 0 {
			m.err = fmt.Errorf("no programs available")
			return m, nil
		}
		return NewSupervisorSelectProgramScreen(m.manager, m.programs, "edit"), nil

	case 3: // Start Program
		if len(m.programs) == 0 {
			m.err = fmt.Errorf("no programs available")
			return m, nil
		}
		return NewSupervisorSelectProgramScreen(m.manager, m.programs, "start"), nil

	case 4: // Stop Program
		if len(m.programs) == 0 {
			m.err = fmt.Errorf("no programs available")
			return m, nil
		}
		return NewSupervisorSelectProgramScreen(m.manager, m.programs, "stop"), nil

	case 5: // Restart Program
		if len(m.programs) == 0 {
			m.err = fmt.Errorf("no programs available")
			return m, nil
		}
		return NewSupervisorSelectProgramScreen(m.manager, m.programs, "restart"), nil

	case 6: // Delete Program
		if len(m.programs) == 0 {
			m.err = fmt.Errorf("no programs available")
			return m, nil
		}
		return NewSupervisorSelectProgramScreen(m.manager, m.programs, "delete"), nil

	case 7: // Configure XML-RPC
		return NewSupervisorXMLRPCConfigScreen(m.manager), nil

	case 8: // View XML-RPC Config
		config, err := m.manager.GetXMLRPCConfig()
		if err != nil {
			m.err = err
			return m, nil
		}
		content := fmt.Sprintf("XML-RPC Configuration:\n\n")
		content += fmt.Sprintf("Enabled: %v\n", config.Enabled)
		content += fmt.Sprintf("IP: %s\n", config.IP)
		content += fmt.Sprintf("Port: %s\n", config.Port)
		content += fmt.Sprintf("Username: %s\n", config.Username)
		if config.Password != "" {
			content += "Password: [configured]\n"
		}
		return NewTextDisplayScreen("XML-RPC Configuration", content, m), nil

	case 9: // Restart Supervisor
		if err := m.manager.RestartSupervisor(); err != nil {
			m.err = err
		} else {
			m.message = "Supervisor restarted successfully"
		}
		return m, nil

	case 10: // Back to Main Menu
		return NewMainMenuScreen(), nil
	}

	return m, nil
}

// Supervisor Program List Screen
type supervisorProgramListScreen struct {
	manager      *system.SupervisorManager
	programs     []system.SupervisorProgram
	selectedItem int
}

func NewSupervisorProgramListScreen(manager *system.SupervisorManager, programs []system.SupervisorProgram) *supervisorProgramListScreen {
	return &supervisorProgramListScreen{
		manager:  manager,
		programs: programs,
	}
}

func (m *supervisorProgramListScreen) Init() tea.Cmd {
	return nil
}

func (m *supervisorProgramListScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewSupervisorManagementScreen(), nil
		case "up", "k":
			if m.selectedItem > 0 {
				m.selectedItem--
			}
		case "down", "j":
			if m.selectedItem < len(m.programs)-1 {
				m.selectedItem++
			}
		case "enter":
			if m.selectedItem < len(m.programs) {
				prog := m.programs[m.selectedItem]
				return NewSupervisorProgramDetailsScreen(m.manager, &prog), nil
			}
		}
	}
	return m, nil
}

func (m *supervisorProgramListScreen) View() string {
	var b strings.Builder

	b.WriteString(theme.HeaderStyle.Render("📋 Supervisor Programs"))
	b.WriteString("\n\n")

	if len(m.programs) == 0 {
		b.WriteString(theme.SubtleStyle.Render("No programs configured"))
		b.WriteString("\n\n")
	} else {
		for i, prog := range m.programs {
			cursor := "  "
			if i == m.selectedItem {
				cursor = theme.SelectedStyle.Render("▶ ")
			}
			
			stateColor := theme.SubtleStyle
			if prog.State == "RUNNING" {
				stateColor = theme.SuccessStyle
			} else if prog.State == "STOPPED" {
				stateColor = theme.ErrorStyle
			}
			
			line := fmt.Sprintf("%s [%s]", prog.Name, stateColor.Render(prog.State))
			if i == m.selectedItem {
				line = theme.SelectedStyle.Render(line)
			}
			b.WriteString(cursor + line + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(theme.HelpStyle.Render("↑/↓: navigate • enter: view details • esc: back • q: quit"))
	return b.String()
}

// Supervisor Program Details Screen
type supervisorProgramDetailsScreen struct {
	manager *system.SupervisorManager
	program *system.SupervisorProgram
}

func NewSupervisorProgramDetailsScreen(manager *system.SupervisorManager, program *system.SupervisorProgram) *supervisorProgramDetailsScreen {
	return &supervisorProgramDetailsScreen{
		manager: manager,
		program: program,
	}
}

func (m *supervisorProgramDetailsScreen) Init() tea.Cmd {
	return nil
}

func (m *supervisorProgramDetailsScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewSupervisorManagementScreen(), nil
		}
	}
	return m, nil
}

func (m *supervisorProgramDetailsScreen) View() string {
	var b strings.Builder

	b.WriteString(theme.HeaderStyle.Render(fmt.Sprintf("🔍 Program: %s", m.program.Name)))
	b.WriteString("\n\n")

	stateColor := theme.SubtleStyle
	if m.program.State == "RUNNING" {
		stateColor = theme.SuccessStyle
	} else if m.program.State == "STOPPED" {
		stateColor = theme.ErrorStyle
	}

	b.WriteString(theme.SubtleStyle.Render("Configuration:"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Name: %s\n", m.program.Name))
	b.WriteString(fmt.Sprintf("  State: %s\n", stateColor.Render(m.program.State)))
	b.WriteString(fmt.Sprintf("  Command: %s\n", m.program.Command))
	b.WriteString(fmt.Sprintf("  Directory: %s\n", m.program.Directory))
	b.WriteString(fmt.Sprintf("  User: %s\n", m.program.User))
	b.WriteString(fmt.Sprintf("  AutoStart: %v\n", m.program.AutoStart))
	b.WriteString(fmt.Sprintf("  Config Path: %s\n", m.program.ConfigPath))
	b.WriteString("\n")

	b.WriteString(theme.HelpStyle.Render("esc: back • q: quit"))
	return b.String()
}

// Supervisor Add Program Screen
type supervisorAddProgramScreen struct {
	manager    *system.SupervisorManager
	inputs     []textinput.Model
	focusIndex int
	autostart  bool
	err        error
}

func NewSupervisorAddProgramScreen(manager *system.SupervisorManager) *supervisorAddProgramScreen {
	inputs := make([]textinput.Model, 4)
	
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Program name"
	inputs[0].Focus()
	inputs[0].Width = 50
	
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Command to execute"
	inputs[1].Width = 50
	
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Working directory"
	inputs[2].Width = 50
	
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "User (default: root)"
	inputs[3].Width = 50

	return &supervisorAddProgramScreen{
		manager:    manager,
		inputs:     inputs,
		focusIndex: 0,
		autostart:  true,
	}
}

func (m *supervisorAddProgramScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m *supervisorAddProgramScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewSupervisorManagementScreen(), nil
		case "tab", "shift+tab", "up", "down":
			if msg.String() == "up" || msg.String() == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}
			if m.focusIndex > len(m.inputs)-1 {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}
			m.updateFocus()
			return m, nil
		case "enter":
			return m.createProgram()
		case "ctrl+a":
			m.autostart = !m.autostart
			return m, nil
		}
	}

	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	return m, cmd
}

func (m *supervisorAddProgramScreen) updateFocus() {
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m *supervisorAddProgramScreen) createProgram() (tea.Model, tea.Cmd) {
	name := m.inputs[0].Value()
	command := m.inputs[1].Value()
	directory := m.inputs[2].Value()
	user := m.inputs[3].Value()

	if name == "" {
		m.err = fmt.Errorf("program name is required")
		return m, nil
	}
	if command == "" {
		m.err = fmt.Errorf("command is required")
		return m, nil
	}
	if directory == "" {
		directory = "/tmp"
	}
	if user == "" {
		user = "root"
	}

	if err := m.manager.CreateProgram(name, command, directory, user, m.autostart); err != nil {
		m.err = err
		return m, nil
	}

	mgmt := NewSupervisorManagementScreen()
	mgmt.message = fmt.Sprintf("Program '%s' created successfully", name)
	return mgmt, nil
}

func (m *supervisorAddProgramScreen) View() string {
	var b strings.Builder

	b.WriteString(theme.HeaderStyle.Render("➕ Add Supervisor Program"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(theme.ErrorStyle.Render("❌ Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	b.WriteString("Program Name:\n")
	b.WriteString(m.inputs[0].View())
	b.WriteString("\n\n")

	b.WriteString("Command:\n")
	b.WriteString(m.inputs[1].View())
	b.WriteString("\n\n")

	b.WriteString("Working Directory:\n")
	b.WriteString(m.inputs[2].View())
	b.WriteString("\n\n")

	b.WriteString("User:\n")
	b.WriteString(m.inputs[3].View())
	b.WriteString("\n\n")

	autostartText := "disabled"
	if m.autostart {
		autostartText = "enabled"
	}
	b.WriteString(fmt.Sprintf("AutoStart: %s (ctrl+a to toggle)\n\n", autostartText))

	b.WriteString(theme.HelpStyle.Render("tab: next • enter: create • ctrl+a: toggle autostart • esc: cancel • q: quit"))

	return b.String()
}

// Supervisor Select Program Screen
type supervisorSelectProgramScreen struct {
	manager      *system.SupervisorManager
	programs     []system.SupervisorProgram
	selectedItem int
	action       string
	err          error
}

func NewSupervisorSelectProgramScreen(manager *system.SupervisorManager, programs []system.SupervisorProgram, action string) *supervisorSelectProgramScreen {
	return &supervisorSelectProgramScreen{
		manager:  manager,
		programs: programs,
		action:   action,
	}
}

func (m *supervisorSelectProgramScreen) Init() tea.Cmd {
	return nil
}

func (m *supervisorSelectProgramScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewSupervisorManagementScreen(), nil
		case "up", "k":
			if m.selectedItem > 0 {
				m.selectedItem--
			}
		case "down", "j":
			if m.selectedItem < len(m.programs)-1 {
				m.selectedItem++
			}
		case "enter":
			prog := m.programs[m.selectedItem]
			return m.performAction(prog)
		}
	}
	return m, nil
}

func (m *supervisorSelectProgramScreen) performAction(prog system.SupervisorProgram) (tea.Model, tea.Cmd) {
	mgmt := NewSupervisorManagementScreen()
	
	switch m.action {
	case "start":
		if err := m.manager.StartProgram(prog.Name); err != nil {
			mgmt.err = err
		} else {
			mgmt.message = fmt.Sprintf("Program '%s' started", prog.Name)
		}
	case "stop":
		if err := m.manager.StopProgram(prog.Name); err != nil {
			mgmt.err = err
		} else {
			mgmt.message = fmt.Sprintf("Program '%s' stopped", prog.Name)
		}
	case "restart":
		if err := m.manager.RestartProgram(prog.Name); err != nil {
			mgmt.err = err
		} else {
			mgmt.message = fmt.Sprintf("Program '%s' restarted", prog.Name)
		}
	case "delete":
		if err := m.manager.DeleteProgram(prog.Name); err != nil {
			mgmt.err = err
		} else {
			mgmt.message = fmt.Sprintf("Program '%s' deleted", prog.Name)
		}
	case "edit":
		return NewSupervisorEditProgramScreen(m.manager, &prog), nil
	}
	
	return mgmt, nil
}

func (m *supervisorSelectProgramScreen) View() string {
	var b strings.Builder

	actionTitle := map[string]string{
		"start":   "Start Program",
		"stop":    "Stop Program",
		"restart": "Restart Program",
		"delete":  "Delete Program",
		"edit":    "Edit Program",
	}

	b.WriteString(theme.HeaderStyle.Render(actionTitle[m.action]))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(theme.ErrorStyle.Render("❌ Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	for i, prog := range m.programs {
		cursor := "  "
		if i == m.selectedItem {
			cursor = theme.SelectedStyle.Render("▶ ")
		}
		
		line := fmt.Sprintf("%s [%s]", prog.Name, prog.State)
		if i == m.selectedItem {
			line = theme.SelectedStyle.Render(line)
		}
		b.WriteString(cursor + line + "\n")
	}

	b.WriteString("\n")
	b.WriteString(theme.HelpStyle.Render("↑/↓: navigate • enter: select • esc: back • q: quit"))

	return b.String()
}

// Supervisor Edit Program Screen
type supervisorEditProgramScreen struct {
	manager    *system.SupervisorManager
	program    *system.SupervisorProgram
	inputs     []textinput.Model
	focusIndex int
	autostart  bool
	err        error
}

func NewSupervisorEditProgramScreen(manager *system.SupervisorManager, program *system.SupervisorProgram) *supervisorEditProgramScreen {
	inputs := make([]textinput.Model, 3)
	
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Command"
	inputs[0].SetValue(program.Command)
	inputs[0].Focus()
	inputs[0].Width = 50
	
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Directory"
	inputs[1].SetValue(program.Directory)
	inputs[1].Width = 50
	
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "User"
	inputs[2].SetValue(program.User)
	inputs[2].Width = 50

	return &supervisorEditProgramScreen{
		manager:    manager,
		program:    program,
		inputs:     inputs,
		focusIndex: 0,
		autostart:  program.AutoStart,
	}
}

func (m *supervisorEditProgramScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m *supervisorEditProgramScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewSupervisorManagementScreen(), nil
		case "tab", "shift+tab", "up", "down":
			if msg.String() == "up" || msg.String() == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}
			if m.focusIndex > len(m.inputs)-1 {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}
			m.updateFocus()
			return m, nil
		case "enter":
			return m.updateProgram()
		case "ctrl+a":
			m.autostart = !m.autostart
			return m, nil
		}
	}

	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	return m, cmd
}

func (m *supervisorEditProgramScreen) updateFocus() {
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m *supervisorEditProgramScreen) updateProgram() (tea.Model, tea.Cmd) {
	command := m.inputs[0].Value()
	directory := m.inputs[1].Value()
	user := m.inputs[2].Value()

	if command == "" {
		m.err = fmt.Errorf("command is required")
		return m, nil
	}

	if err := m.manager.UpdateProgram(m.program.Name, command, directory, user, m.autostart); err != nil {
		m.err = err
		return m, nil
	}

	mgmt := NewSupervisorManagementScreen()
	mgmt.message = fmt.Sprintf("Program '%s' updated successfully", m.program.Name)
	return mgmt, nil
}

func (m *supervisorEditProgramScreen) View() string {
	var b strings.Builder

	b.WriteString(theme.HeaderStyle.Render(fmt.Sprintf("✏️  Edit Program: %s", m.program.Name)))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(theme.ErrorStyle.Render("❌ Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	b.WriteString("Command:\n")
	b.WriteString(m.inputs[0].View())
	b.WriteString("\n\n")

	b.WriteString("Directory:\n")
	b.WriteString(m.inputs[1].View())
	b.WriteString("\n\n")

	b.WriteString("User:\n")
	b.WriteString(m.inputs[2].View())
	b.WriteString("\n\n")

	autostartText := "disabled"
	if m.autostart {
		autostartText = "enabled"
	}
	b.WriteString(fmt.Sprintf("AutoStart: %s (ctrl+a to toggle)\n\n", autostartText))

	b.WriteString(theme.HelpStyle.Render("tab: next • enter: save • ctrl+a: toggle autostart • esc: cancel • q: quit"))

	return b.String()
}

// Supervisor XML-RPC Config Screen
type supervisorXMLRPCConfigScreen struct {
	manager    *system.SupervisorManager
	inputs     []textinput.Model
	focusIndex int
	err        error
}

func NewSupervisorXMLRPCConfigScreen(manager *system.SupervisorManager) *supervisorXMLRPCConfigScreen {
	inputs := make([]textinput.Model, 4)
	
	// Load current config
	config, _ := manager.GetXMLRPCConfig()
	
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "IP Address (e.g., 127.0.0.1)"
	inputs[0].Focus()
	inputs[0].Width = 40
	if config != nil {
		inputs[0].SetValue(config.IP)
	}
	
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Port (e.g., 9001)"
	inputs[1].Width = 40
	if config != nil {
		inputs[1].SetValue(config.Port)
	}
	
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Username"
	inputs[2].Width = 40
	if config != nil {
		inputs[2].SetValue(config.Username)
	}
	
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Password"
	inputs[3].Width = 40
	inputs[3].EchoMode = textinput.EchoPassword
	inputs[3].EchoCharacter = '•'

	return &supervisorXMLRPCConfigScreen{
		manager:    manager,
		inputs:     inputs,
		focusIndex: 0,
	}
}

func (m *supervisorXMLRPCConfigScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m *supervisorXMLRPCConfigScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewSupervisorManagementScreen(), nil
		case "tab", "shift+tab", "up", "down":
			if msg.String() == "up" || msg.String() == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}
			if m.focusIndex > len(m.inputs)-1 {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}
			m.updateFocus()
			return m, nil
		case "enter":
			return m.saveConfig()
		}
	}

	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	return m, cmd
}

func (m *supervisorXMLRPCConfigScreen) updateFocus() {
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m *supervisorXMLRPCConfigScreen) saveConfig() (tea.Model, tea.Cmd) {
	ip := m.inputs[0].Value()
	port := m.inputs[1].Value()
	username := m.inputs[2].Value()
	password := m.inputs[3].Value()

	if ip == "" {
		ip = "127.0.0.1"
	}
	if port == "" {
		port = "9001"
	}
	if username == "" {
		m.err = fmt.Errorf("username is required")
		return m, nil
	}
	if password == "" {
		m.err = fmt.Errorf("password is required")
		return m, nil
	}

	if err := m.manager.SetXMLRPCConfig(ip, port, username, password); err != nil {
		m.err = err
		return m, nil
	}

	mgmt := NewSupervisorManagementScreen()
	mgmt.message = "XML-RPC configured successfully. Supervisor will restart."
	return mgmt, nil
}

func (m *supervisorXMLRPCConfigScreen) View() string {
	var b strings.Builder

	b.WriteString(theme.HeaderStyle.Render("🔧 Configure XML-RPC Server"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(theme.ErrorStyle.Render("❌ Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	b.WriteString(theme.SubtleStyle.Render("Configure the Supervisor XML-RPC interface for remote management"))
	b.WriteString("\n\n")

	b.WriteString("IP Address:\n")
	b.WriteString(m.inputs[0].View())
	b.WriteString("\n\n")

	b.WriteString("Port:\n")
	b.WriteString(m.inputs[1].View())
	b.WriteString("\n\n")

	b.WriteString("Username:\n")
	b.WriteString(m.inputs[2].View())
	b.WriteString("\n\n")

	b.WriteString("Password:\n")
	b.WriteString(m.inputs[3].View())
	b.WriteString("\n\n")

	b.WriteString(theme.SubtleStyle.Render("Note: Supervisor will be restarted after saving"))
	b.WriteString("\n\n")

	b.WriteString(theme.HelpStyle.Render("tab: next field • enter: save • esc: cancel • q: quit"))

	return b.String()
}
