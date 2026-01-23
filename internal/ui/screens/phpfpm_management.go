package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

type phpfpmManagementScreen struct {
	manager      *system.PHPFPMManager
	pools        []system.PHPFPMPool
	selectedItem int
	menuItems    []string
	err          error
	message      string
}

func NewPHPFPMManagementScreen() *phpfpmManagementScreen {
	manager := system.NewPHPFPMManager("")
	// Try to detect PHP version
	manager.DetectPHPVersion()
	
	return &phpfpmManagementScreen{
		manager: manager,
		menuItems: []string{
			"List All Pools",
			"Create New Pool",
			"Edit Pool",
			"Delete Pool",
			"Restart PHP-FPM Service",
			"Reload PHP-FPM Service",
			"View Service Status",
			"Back to Main Menu",
		},
	}
}

func (m *phpfpmManagementScreen) Init() tea.Cmd {
	// Load pools
	pools, err := m.manager.ListPools()
	if err != nil {
		m.err = err
	} else {
		m.pools = pools
	}
	return nil
}

func (m *phpfpmManagementScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *phpfpmManagementScreen) View() string {
	var b strings.Builder

	// Header
	b.WriteString(theme.HeaderStyle.Render("🐘 PHP-FPM Pool Management"))
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

	// Show pools count
	b.WriteString(theme.SubtleStyle.Render(fmt.Sprintf("Total Pools: %d", len(m.pools))))
	b.WriteString("\n\n")

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

func (m *phpfpmManagementScreen) handleSelection() (tea.Model, tea.Cmd) {
	m.err = nil
	m.message = ""

	switch m.selectedItem {
	case 0: // List All Pools
		pools, err := m.manager.ListPools()
		if err != nil {
			m.err = err
			return m, nil
		}
		m.pools = pools
		return NewPHPFPMPoolListScreen(m.manager, pools), nil

	case 1: // Create New Pool
		return NewPHPFPMCreatePoolScreen(m.manager), nil

	case 2: // Edit Pool
		if len(m.pools) == 0 {
			m.err = fmt.Errorf("no pools available to edit")
			return m, nil
		}
		return NewPHPFPMSelectPoolScreen(m.manager, m.pools, "edit"), nil

	case 3: // Delete Pool
		if len(m.pools) == 0 {
			m.err = fmt.Errorf("no pools available to delete")
			return m, nil
		}
		return NewPHPFPMSelectPoolScreen(m.manager, m.pools, "delete"), nil

	case 4: // Restart PHP-FPM Service
		if err := m.manager.RestartService(); err != nil {
			m.err = err
		} else {
			m.message = "PHP-FPM service restarted successfully"
		}
		return m, nil

	case 5: // Reload PHP-FPM Service
		if err := m.manager.ReloadService(); err != nil {
			m.err = err
		} else {
			m.message = "PHP-FPM service reloaded successfully"
		}
		return m, nil

	case 6: // View Service Status
		status, err := m.manager.GetStatus()
		if err != nil {
			m.err = err
		} else {
			return NewTextDisplayScreen("PHP-FPM Service Status", status, m), nil
		}
		return m, nil

	case 7: // Back to Main Menu
		return NewMainMenuScreen(), nil
	}

	return m, nil
}

// PHP-FPM Pool List Screen
type phpfpmPoolListScreen struct {
	manager      *system.PHPFPMManager
	pools        []system.PHPFPMPool
	selectedItem int
}

func NewPHPFPMPoolListScreen(manager *system.PHPFPMManager, pools []system.PHPFPMPool) *phpfpmPoolListScreen {
	return &phpfpmPoolListScreen{
		manager: manager,
		pools:   pools,
	}
}

func (m *phpfpmPoolListScreen) Init() tea.Cmd {
	return nil
}

func (m *phpfpmPoolListScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewPHPFPMManagementScreen(), nil
		case "up", "k":
			if m.selectedItem > 0 {
				m.selectedItem--
			}
		case "down", "j":
			if m.selectedItem < len(m.pools)-1 {
				m.selectedItem++
			}
		case "enter":
			if m.selectedItem < len(m.pools) {
				pool := m.pools[m.selectedItem]
				return NewPHPFPMPoolDetailsScreen(m.manager, &pool), nil
			}
		}
	}

	return m, nil
}

func (m *phpfpmPoolListScreen) View() string {
	var b strings.Builder

	b.WriteString(theme.HeaderStyle.Render("📋 PHP-FPM Pools"))
	b.WriteString("\n\n")

	if len(m.pools) == 0 {
		b.WriteString(theme.SubtleStyle.Render("No pools configured"))
		b.WriteString("\n\n")
	} else {
		for i, pool := range m.pools {
			cursor := "  "
			if i == m.selectedItem {
				cursor = theme.SelectedStyle.Render("▶ ")
			}
			
			line := fmt.Sprintf("%s [%s] - %s", pool.Name, pool.PM, pool.Listen)
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

// PHP-FPM Pool Details Screen
type phpfpmPoolDetailsScreen struct {
	manager *system.PHPFPMManager
	pool    *system.PHPFPMPool
}

func NewPHPFPMPoolDetailsScreen(manager *system.PHPFPMManager, pool *system.PHPFPMPool) *phpfpmPoolDetailsScreen {
	return &phpfpmPoolDetailsScreen{
		manager: manager,
		pool:    pool,
	}
}

func (m *phpfpmPoolDetailsScreen) Init() tea.Cmd {
	return nil
}

func (m *phpfpmPoolDetailsScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewPHPFPMManagementScreen(), nil
		}
	}

	return m, nil
}

func (m *phpfpmPoolDetailsScreen) View() string {
	var b strings.Builder

	b.WriteString(theme.HeaderStyle.Render(fmt.Sprintf("🔍 Pool: %s", m.pool.Name)))
	b.WriteString("\n\n")

	b.WriteString(theme.SubtleStyle.Render("Configuration:"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Name: %s\n", m.pool.Name))
	b.WriteString(fmt.Sprintf("  User: %s\n", m.pool.User))
	b.WriteString(fmt.Sprintf("  Group: %s\n", m.pool.Group))
	b.WriteString(fmt.Sprintf("  Listen: %s\n", m.pool.Listen))
	b.WriteString(fmt.Sprintf("  Listen Owner: %s\n", m.pool.ListenOwner))
	b.WriteString(fmt.Sprintf("  Listen Group: %s\n", m.pool.ListenGroup))
	b.WriteString(fmt.Sprintf("  Listen Mode: %s\n", m.pool.ListenMode))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  PM Mode: %s\n", m.pool.PM))
	b.WriteString(fmt.Sprintf("  PM Max Children: %d\n", m.pool.PMMaxChildren))
	if m.pool.PM == "dynamic" {
		b.WriteString(fmt.Sprintf("  PM Start Servers: %d\n", m.pool.PMStartServers))
		b.WriteString(fmt.Sprintf("  PM Min Spare: %d\n", m.pool.PMMinSpareServers))
		b.WriteString(fmt.Sprintf("  PM Max Spare: %d\n", m.pool.PMMaxSpareServers))
	}
	b.WriteString(fmt.Sprintf("  PM Max Requests: %d\n", m.pool.PMMaxRequests))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Config Path: %s\n", m.pool.ConfigPath))
	b.WriteString("\n")

	b.WriteString(theme.HelpStyle.Render("esc: back • q: quit"))

	return b.String()
}

// PHP-FPM Create Pool Screen
type phpfpmCreatePoolScreen struct {
	manager    *system.PHPFPMManager
	inputs     []textinput.Model
	focusIndex int
	err        error
	pmMode     string
}

func NewPHPFPMCreatePoolScreen(manager *system.PHPFPMManager) *phpfpmCreatePoolScreen {
	inputs := make([]textinput.Model, 5)
	
	// Pool name
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Pool name"
	inputs[0].Focus()
	inputs[0].Width = 40
	
	// User
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "User (default: www-data)"
	inputs[1].Width = 40
	
	// Group
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Group (default: www-data)"
	inputs[2].Width = 40
	
	// Listen socket/port
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Listen (default: auto-generated socket)"
	inputs[3].Width = 40
	
	// Max children
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "Max children (default: 5)"
	inputs[4].Width = 40

	return &phpfpmCreatePoolScreen{
		manager:    manager,
		inputs:     inputs,
		focusIndex: 0,
		pmMode:     "dynamic",
	}
}

func (m *phpfpmCreatePoolScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m *phpfpmCreatePoolScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewPHPFPMManagementScreen(), nil
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
			return m.createPool()
		}
	}

	// Update focused input
	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	return m, cmd
}

func (m *phpfpmCreatePoolScreen) updateFocus() {
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m *phpfpmCreatePoolScreen) createPool() (tea.Model, tea.Cmd) {
	poolName := m.inputs[0].Value()
	if poolName == "" {
		m.err = fmt.Errorf("pool name is required")
		return m, nil
	}

	pool := &system.PHPFPMPool{
		Name:   poolName,
		User:   m.inputs[1].Value(),
		Group:  m.inputs[2].Value(),
		Listen: m.inputs[3].Value(),
		PM:     m.pmMode,
	}

	// Parse max children if provided
	if m.inputs[4].Value() != "" {
		fmt.Sscanf(m.inputs[4].Value(), "%d", &pool.PMMaxChildren)
	}

	if err := m.manager.CreatePool(pool); err != nil {
		m.err = err
		return m, nil
	}

	// Reload service
	if err := m.manager.ReloadService(); err != nil {
		m.err = fmt.Errorf("pool created but failed to reload service: %w", err)
		return m, nil
	}

	// Success
	mgmt := NewPHPFPMManagementScreen()
	mgmt.message = fmt.Sprintf("Pool '%s' created successfully", poolName)
	return mgmt, nil
}

func (m *phpfpmCreatePoolScreen) View() string {
	var b strings.Builder

	b.WriteString(theme.HeaderStyle.Render("➕ Create PHP-FPM Pool"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(theme.ErrorStyle.Render("❌ Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	b.WriteString("Pool Name:\n")
	b.WriteString(m.inputs[0].View())
	b.WriteString("\n\n")

	b.WriteString("User:\n")
	b.WriteString(m.inputs[1].View())
	b.WriteString("\n\n")

	b.WriteString("Group:\n")
	b.WriteString(m.inputs[2].View())
	b.WriteString("\n\n")

	b.WriteString("Listen:\n")
	b.WriteString(m.inputs[3].View())
	b.WriteString("\n\n")

	b.WriteString("Max Children:\n")
	b.WriteString(m.inputs[4].View())
	b.WriteString("\n\n")

	b.WriteString(theme.HelpStyle.Render("tab: next field • enter: create • esc: cancel • q: quit"))

	return b.String()
}

// PHP-FPM Select Pool Screen (for edit/delete)
type phpfpmSelectPoolScreen struct {
	manager      *system.PHPFPMManager
	pools        []system.PHPFPMPool
	selectedItem int
	action       string // "edit" or "delete"
}

func NewPHPFPMSelectPoolScreen(manager *system.PHPFPMManager, pools []system.PHPFPMPool, action string) *phpfpmSelectPoolScreen {
	return &phpfpmSelectPoolScreen{
		manager: manager,
		pools:   pools,
		action:  action,
	}
}

func (m *phpfpmSelectPoolScreen) Init() tea.Cmd {
	return nil
}

func (m *phpfpmSelectPoolScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewPHPFPMManagementScreen(), nil
		case "up", "k":
			if m.selectedItem > 0 {
				m.selectedItem--
			}
		case "down", "j":
			if m.selectedItem < len(m.pools)-1 {
				m.selectedItem++
			}
		case "enter":
			pool := m.pools[m.selectedItem]
			if m.action == "edit" {
				return NewPHPFPMEditPoolScreen(m.manager, &pool), nil
			} else if m.action == "delete" {
				return m.deletePool(pool.Name)
			}
		}
	}

	return m, nil
}

func (m *phpfpmSelectPoolScreen) deletePool(poolName string) (tea.Model, tea.Cmd) {
	if err := m.manager.DeletePool(poolName); err != nil {
		mgmt := NewPHPFPMManagementScreen()
		mgmt.err = err
		return mgmt, nil
	}

	// Reload service
	m.manager.ReloadService()

	mgmt := NewPHPFPMManagementScreen()
	mgmt.message = fmt.Sprintf("Pool '%s' deleted successfully", poolName)
	return mgmt, nil
}

func (m *phpfpmSelectPoolScreen) View() string {
	var b strings.Builder

	title := "Select Pool to Edit"
	if m.action == "delete" {
		title = "Select Pool to Delete"
	}
	b.WriteString(theme.HeaderStyle.Render(title))
	b.WriteString("\n\n")

	for i, pool := range m.pools {
		cursor := "  "
		if i == m.selectedItem {
			cursor = theme.SelectedStyle.Render("▶ ")
		}
		
		line := fmt.Sprintf("%s [%s]", pool.Name, pool.Listen)
		if i == m.selectedItem {
			line = theme.SelectedStyle.Render(line)
		}
		b.WriteString(cursor + line + "\n")
	}

	b.WriteString("\n")
	b.WriteString(theme.HelpStyle.Render("↑/↓: navigate • enter: select • esc: back • q: quit"))

	return b.String()
}

// PHP-FPM Edit Pool Screen
type phpfpmEditPoolScreen struct {
	manager    *system.PHPFPMManager
	pool       *system.PHPFPMPool
	inputs     []textinput.Model
	focusIndex int
	err        error
}

func NewPHPFPMEditPoolScreen(manager *system.PHPFPMManager, pool *system.PHPFPMPool) *phpfpmEditPoolScreen {
	inputs := make([]textinput.Model, 4)
	
	// User
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "User"
	inputs[0].SetValue(pool.User)
	inputs[0].Focus()
	inputs[0].Width = 40
	
	// Group
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Group"
	inputs[1].SetValue(pool.Group)
	inputs[1].Width = 40
	
	// Listen
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Listen"
	inputs[2].SetValue(pool.Listen)
	inputs[2].Width = 40
	
	// Max children
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "Max children"
	inputs[3].SetValue(fmt.Sprintf("%d", pool.PMMaxChildren))
	inputs[3].Width = 40

	return &phpfpmEditPoolScreen{
		manager:    manager,
		pool:       pool,
		inputs:     inputs,
		focusIndex: 0,
	}
}

func (m *phpfpmEditPoolScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (m *phpfpmEditPoolScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			return NewPHPFPMManagementScreen(), nil
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
			return m.updatePool()
		}
	}

	// Update focused input
	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
	return m, cmd
}

func (m *phpfpmEditPoolScreen) updateFocus() {
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m *phpfpmEditPoolScreen) updatePool() (tea.Model, tea.Cmd) {
	m.pool.User = m.inputs[0].Value()
	m.pool.Group = m.inputs[1].Value()
	m.pool.Listen = m.inputs[2].Value()
	
	if m.inputs[3].Value() != "" {
		fmt.Sscanf(m.inputs[3].Value(), "%d", &m.pool.PMMaxChildren)
	}

	if err := m.manager.UpdatePool(m.pool); err != nil {
		m.err = err
		return m, nil
	}

	// Reload service
	if err := m.manager.ReloadService(); err != nil {
		m.err = fmt.Errorf("pool updated but failed to reload service: %w", err)
		return m, nil
	}

	// Success
	mgmt := NewPHPFPMManagementScreen()
	mgmt.message = fmt.Sprintf("Pool '%s' updated successfully", m.pool.Name)
	return mgmt, nil
}

func (m *phpfpmEditPoolScreen) View() string {
	var b strings.Builder

	b.WriteString(theme.HeaderStyle.Render(fmt.Sprintf("✏️  Edit Pool: %s", m.pool.Name)))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(theme.ErrorStyle.Render("❌ Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	b.WriteString("User:\n")
	b.WriteString(m.inputs[0].View())
	b.WriteString("\n\n")

	b.WriteString("Group:\n")
	b.WriteString(m.inputs[1].View())
	b.WriteString("\n\n")

	b.WriteString("Listen:\n")
	b.WriteString(m.inputs[2].View())
	b.WriteString("\n\n")

	b.WriteString("Max Children:\n")
	b.WriteString(m.inputs[3].View())
	b.WriteString("\n\n")

	b.WriteString(theme.HelpStyle.Render("tab: next field • enter: save • esc: cancel • q: quit"))

	return b.String()
}
