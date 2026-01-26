package screens

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// FrankenPHPService represents a FrankenPHP systemd service
type FrankenPHPService struct {
	Name        string
	ServiceFile string
	SiteKey     string
	Status      string // running, stopped, failed
	Enabled     bool
	SiteRoot    string
	Port        string
	User        string
}

// FPServicesState represents the current state of the screen
type FPServicesState int

const (
	FPServicesStateList FPServicesState = iota
	FPServicesStateActions
	FPServicesStateEdit
	FPServicesStateConfirm
)

// FrankenPHPServicesModel represents the FrankenPHP services management screen
type FrankenPHPServicesModel struct {
	theme    *theme.Theme
	width    int
	height   int
	state    FPServicesState
	services []FrankenPHPService
	cursor   int

	// Action menu
	actionCursor int
	actions      []string

	// Edit form
	editForm     *huh.Form
	editSiteRoot string
	editDocroot  string
	editDomains  string
	editPort     string
	editUser     string
	editGroup    string

	// Confirm action
	confirmAction string
	confirmMsg    string

	// Messages
	err     error
	message string
}

// NewFrankenPHPServicesModel creates a new FrankenPHP services model
func NewFrankenPHPServicesModel() FrankenPHPServicesModel {
	t := theme.DefaultTheme()

	m := FrankenPHPServicesModel{
		theme:        t,
		state:        FPServicesStateList,
		cursor:       0,
		actionCursor: 0,
		actions: []string{
			"Start Service",
			"Stop Service",
			"Restart Service",
			"Enable (start on boot)",
			"Disable (don't start on boot)",
			"View Status",
			"View Logs",
			"Edit Configuration (Form)",
			"Edit Configuration (Editor)",
			"← Back to List",
		},
	}

	// Load services
	m.services = m.loadFrankenPHPServices()

	return m
}

// loadFrankenPHPServices discovers FrankenPHP systemd services
func (m *FrankenPHPServicesModel) loadFrankenPHPServices() []FrankenPHPService {
	var services []FrankenPHPService

	// Find all frankenphp-*.service files
	cmd := exec.Command("bash", "-c", `ls /etc/systemd/system/frankenphp-*.service 2>/dev/null || true`)
	output, _ := cmd.Output()

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Extract site key from filename
		// /etc/systemd/system/frankenphp-mysite.service -> mysite
		filename := strings.TrimPrefix(line, "/etc/systemd/system/frankenphp-")
		siteKey := strings.TrimSuffix(filename, ".service")

		if siteKey == "" {
			continue
		}

		service := FrankenPHPService{
			Name:        fmt.Sprintf("frankenphp-%s", siteKey),
			ServiceFile: line,
			SiteKey:     siteKey,
		}

		// Get service status
		statusCmd := exec.Command("systemctl", "is-active", service.Name)
		statusOutput, _ := statusCmd.Output()
		service.Status = strings.TrimSpace(string(statusOutput))
		if service.Status == "" {
			service.Status = "unknown"
		}

		// Check if enabled
		enabledCmd := exec.Command("systemctl", "is-enabled", service.Name)
		enabledOutput, _ := enabledCmd.Output()
		service.Enabled = strings.TrimSpace(string(enabledOutput)) == "enabled"

		// Parse service file for details
		service.SiteRoot, service.Port, service.User = m.parseServiceFile(line)

		services = append(services, service)
	}

	return services
}

// parseServiceFile extracts configuration from a service file
func (m *FrankenPHPServicesModel) parseServiceFile(path string) (siteRoot, port, user string) {
	cmd := exec.Command("cat", path)
	output, err := cmd.Output()
	if err != nil {
		return "", "", ""
	}

	content := string(output)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "WorkingDirectory=") {
			siteRoot = strings.TrimPrefix(line, "WorkingDirectory=")
		} else if strings.HasPrefix(line, "User=") {
			user = strings.TrimPrefix(line, "User=")
		} else if strings.Contains(line, "--listen") && strings.Contains(line, ":") {
			// Extract port from --listen 127.0.0.1:8000
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				// Get last part which should contain port
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if len(p) > 0 && len(p) <= 5 {
						// Check if it looks like a port number
						isPort := true
						for _, c := range p {
							if c < '0' || c > '9' {
								isPort = false
								break
							}
						}
						if isPort && p != "" {
							port = strings.Split(p, " ")[0]
							port = strings.TrimSuffix(port, "\\")
						}
					}
				}
			}
		}
	}

	return siteRoot, port, user
}

// Init initializes the screen
func (m FrankenPHPServicesModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m FrankenPHPServicesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Clear messages on any key
		if m.message != "" || m.err != nil {
			m.message = ""
			m.err = nil
		}

		switch m.state {
		case FPServicesStateList:
			return m.updateList(msg)
		case FPServicesStateActions:
			return m.updateActions(msg)
		case FPServicesStateEdit:
			return m.updateEdit(msg)
		case FPServicesStateConfirm:
			return m.updateConfirm(msg)
		}
	}

	// Update form if in edit state
	if m.state == FPServicesStateEdit && m.editForm != nil {
		form, cmd := m.editForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.editForm = f
		}

		// Check if form is completed after update
		if m.editForm.State == huh.StateCompleted {
			return m.saveServiceConfig()
		}

		return m, cmd
	}

	return m, nil
}

// updateList handles list view navigation
func (m FrankenPHPServicesModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c", "q"))):
		return m, tea.Quit
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		return m, func() tea.Msg { return BackMsg{} }
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if m.cursor < len(m.services)-1 {
			m.cursor++
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter", " "))):
		if len(m.services) > 0 {
			m.state = FPServicesStateActions
			m.actionCursor = 0
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
		// Refresh services list
		m.services = m.loadFrankenPHPServices()
		m.message = "Services refreshed"
	}
	return m, nil
}

// updateActions handles action menu navigation
func (m FrankenPHPServicesModel) updateActions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c", "q"))):
		return m, tea.Quit
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		m.state = FPServicesStateList
		return m, nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if m.actionCursor > 0 {
			m.actionCursor--
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if m.actionCursor < len(m.actions)-1 {
			m.actionCursor++
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter", " "))):
		return m.executeAction()
	}
	return m, nil
}

// updateEdit handles edit form
func (m FrankenPHPServicesModel) updateEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editForm == nil {
		// Form not initialized, go back to actions
		m.state = FPServicesStateActions
		return m, nil
	}

	// Handle escape before form update
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.state = FPServicesStateActions
		m.editForm = nil
		return m, nil
	}

	return m, nil
}

// updateConfirm handles confirmation dialog
func (m FrankenPHPServicesModel) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc", "n", "N":
		m.state = FPServicesStateActions
		return m, nil
	case "y", "Y", "enter":
		return m.doConfirmedAction()
	}
	return m, nil
}

// executeAction executes the selected action
func (m FrankenPHPServicesModel) executeAction() (tea.Model, tea.Cmd) {
	if len(m.services) == 0 || m.cursor >= len(m.services) {
		return m, nil
	}

	service := m.services[m.cursor]
	action := m.actions[m.actionCursor]

	switch action {
	case "Start Service":
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     fmt.Sprintf("systemctl start %s && systemctl status %s", service.Name, service.Name),
				Description: fmt.Sprintf("Starting %s", service.Name),
			}
		}

	case "Stop Service":
		m.confirmAction = "stop"
		m.confirmMsg = fmt.Sprintf("Stop service %s?", service.Name)
		m.state = FPServicesStateConfirm
		return m, nil

	case "Restart Service":
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     fmt.Sprintf("systemctl restart %s && systemctl status %s", service.Name, service.Name),
				Description: fmt.Sprintf("Restarting %s", service.Name),
			}
		}

	case "Enable (start on boot)":
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     fmt.Sprintf("systemctl enable %s && echo '✓ Service enabled'", service.Name),
				Description: fmt.Sprintf("Enabling %s", service.Name),
			}
		}

	case "Disable (don't start on boot)":
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     fmt.Sprintf("systemctl disable %s && echo '✓ Service disabled'", service.Name),
				Description: fmt.Sprintf("Disabling %s", service.Name),
			}
		}

	case "View Status":
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     fmt.Sprintf("systemctl status %s", service.Name),
				Description: fmt.Sprintf("Status of %s", service.Name),
			}
		}

	case "View Logs":
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     fmt.Sprintf("journalctl -u %s -n 100 --no-pager", service.Name),
				Description: fmt.Sprintf("Logs for %s", service.Name),
			}
		}

	case "Edit Configuration (Form)":
		m.state = FPServicesStateEdit
		m.loadServiceForEdit(service)
		m.editForm = m.buildEditForm()
		return m, m.editForm.Init()

	case "Edit Configuration (Editor)":
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: EditorSelectionScreen,
				Data: map[string]interface{}{
					"file":        service.ServiceFile,
					"description": fmt.Sprintf("Edit %s service file", service.Name),
				},
			}
		}

	case "← Back to List":
		m.state = FPServicesStateList
		return m, nil
	}

	return m, nil
}

// doConfirmedAction executes the confirmed action
func (m FrankenPHPServicesModel) doConfirmedAction() (tea.Model, tea.Cmd) {
	if len(m.services) == 0 || m.cursor >= len(m.services) {
		return m, nil
	}

	service := m.services[m.cursor]

	switch m.confirmAction {
	case "stop":
		m.state = FPServicesStateList
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     fmt.Sprintf("systemctl stop %s && echo '✓ Service stopped'", service.Name),
				Description: fmt.Sprintf("Stopping %s", service.Name),
			}
		}
	}

	m.state = FPServicesStateActions
	return m, nil
}

// loadServiceForEdit loads service config into edit form fields
func (m *FrankenPHPServicesModel) loadServiceForEdit(service FrankenPHPService) {
	m.editSiteRoot = service.SiteRoot
	m.editPort = service.Port
	m.editUser = service.User
	m.editGroup = service.User // Often same as user
	m.editDocroot = ""
	m.editDomains = ""

	// Try to read nginx config for domains
	nginxConf := fmt.Sprintf("/etc/nginx/sites-available/%s.conf", service.SiteKey)
	cmd := exec.Command("bash", "-c", fmt.Sprintf("grep -oP 'server_name \\K[^;]+' %s 2>/dev/null || true", nginxConf))
	output, _ := cmd.Output()
	m.editDomains = strings.TrimSpace(string(output))
}

// buildEditForm creates the edit form
func (m *FrankenPHPServicesModel) buildEditForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("siteRoot").
				Title("Site Root").
				Description("Full path to your application root").
				Value(&m.editSiteRoot),

			huh.NewInput().
				Key("docroot").
				Title("Document Root (optional)").
				Description("Web-accessible directory (e.g., /public for Laravel)").
				Value(&m.editDocroot),

			huh.NewInput().
				Key("domains").
				Title("Domain Names").
				Description("Space-separated domain names").
				Value(&m.editDomains),

			huh.NewInput().
				Key("port").
				Title("Port").
				Description("TCP port for FrankenPHP").
				Value(&m.editPort),

			huh.NewInput().
				Key("user").
				Title("Run as User").
				Value(&m.editUser),

			huh.NewInput().
				Key("group").
				Title("Run as Group").
				Value(&m.editGroup),
		),
	).WithTheme(m.theme.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)
}

// saveServiceConfig saves the edited service configuration
func (m FrankenPHPServicesModel) saveServiceConfig() (tea.Model, tea.Cmd) {
	if len(m.services) == 0 || m.cursor >= len(m.services) {
		m.state = FPServicesStateList
		m.editForm = nil
		m.err = fmt.Errorf("no service selected")
		return m, nil
	}

	service := m.services[m.cursor]

	// Read form values
	if m.editForm != nil {
		if v := m.editForm.GetString("siteRoot"); v != "" {
			m.editSiteRoot = v
		}
		if v := m.editForm.GetString("docroot"); v != "" {
			m.editDocroot = v
		}
		if v := m.editForm.GetString("domains"); v != "" {
			m.editDomains = v
		}
		if v := m.editForm.GetString("port"); v != "" {
			m.editPort = v
		}
		if v := m.editForm.GetString("user"); v != "" {
			m.editUser = v
		}
		if v := m.editForm.GetString("group"); v != "" {
			m.editGroup = v
		}
	}

	// Build update script
	script := fmt.Sprintf(`
echo "Updating service configuration for %s..."
echo ""

# Update systemd service file
SERVICE_FILE="%s"
SITE_ROOT="%s"
PORT="%s"
USER="%s"
GROUP="%s"

# Create backup
cp "$SERVICE_FILE" "${SERVICE_FILE}.bak"
echo "✓ Created backup: ${SERVICE_FILE}.bak"

# Update WorkingDirectory
sed -i "s|WorkingDirectory=.*|WorkingDirectory=${SITE_ROOT}|" "$SERVICE_FILE"
echo "✓ Updated WorkingDirectory"

# Update User
sed -i "s|User=.*|User=${USER}|" "$SERVICE_FILE"
echo "✓ Updated User"

# Update Group
sed -i "s|Group=.*|Group=${GROUP}|" "$SERVICE_FILE"
echo "✓ Updated Group"

# Update port in ExecStart (if present)
sed -i "s|127.0.0.1:[0-9]*|127.0.0.1:${PORT}|g" "$SERVICE_FILE"
echo "✓ Updated Port"

# Reload systemd
systemctl daemon-reload
echo "✓ Reloaded systemd"

echo ""
echo "Service configuration updated!"
echo "Restart the service to apply changes: systemctl restart %s"
`, service.Name, service.ServiceFile, m.editSiteRoot, m.editPort, m.editUser, m.editGroup, service.Name)

	m.state = FPServicesStateList
	m.editForm = nil

	return m, func() tea.Msg {
		return ExecutionStartMsg{
			Command:     script,
			Description: fmt.Sprintf("Updating %s configuration", service.Name),
		}
	}
}

// View renders the screen
func (m FrankenPHPServicesModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	switch m.state {
	case FPServicesStateList:
		return m.viewList()
	case FPServicesStateActions:
		return m.viewActions()
	case FPServicesStateEdit:
		return m.viewEdit()
	case FPServicesStateConfirm:
		return m.viewConfirm()
	}

	return "Unknown state"
}

// viewList renders the services list
func (m FrankenPHPServicesModel) viewList() string {
	header := m.theme.Title.Render("FrankenPHP Services")

	if len(m.services) == 0 {
		noServices := lipgloss.JoinVertical(lipgloss.Left,
			"",
			m.theme.WarningStyle.Render("No FrankenPHP services found."),
			"",
			m.theme.DescriptionStyle.Render("Set up a site using 'FrankenPHP Classic Mode' first."),
		)
		help := m.theme.Help.Render("Esc: Back • r: Refresh")
		content := lipgloss.JoinVertical(lipgloss.Left, header, noServices, "", help)
		bordered := m.theme.BorderStyle.Render(content)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
	}

	// Services list
	var items []string
	items = append(items, "")

	for i, svc := range m.services {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		// Status indicator
		var statusIndicator string
		switch svc.Status {
		case "active":
			statusIndicator = m.theme.SuccessStyle.Render("●")
		case "inactive":
			statusIndicator = m.theme.WarningStyle.Render("○")
		case "failed":
			statusIndicator = m.theme.ErrorStyle.Render("✗")
		default:
			statusIndicator = m.theme.DescriptionStyle.Render("?")
		}

		// Enabled indicator
		enabledStr := ""
		if svc.Enabled {
			enabledStr = m.theme.DescriptionStyle.Render(" [enabled]")
		}

		name := fmt.Sprintf("%s %s%s", statusIndicator, svc.Name, enabledStr)

		var renderedItem string
		if i == m.cursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, name))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, name))
		}
		items = append(items, renderedItem)

		// Show details for selected service
		if i == m.cursor {
			details := []string{}
			if svc.SiteRoot != "" {
				details = append(details, fmt.Sprintf("    Root: %s", svc.SiteRoot))
			}
			if svc.Port != "" {
				details = append(details, fmt.Sprintf("    Port: %s", svc.Port))
			}
			if svc.User != "" {
				details = append(details, fmt.Sprintf("    User: %s", svc.User))
			}
			for _, d := range details {
				items = append(items, m.theme.DescriptionStyle.Render(d))
			}
		}
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, items...)

	// Legend
	legend := lipgloss.JoinVertical(lipgloss.Left,
		"",
		m.theme.DescriptionStyle.Render("Legend: "+m.theme.SuccessStyle.Render("●")+" running  "+m.theme.WarningStyle.Render("○")+" stopped  "+m.theme.ErrorStyle.Render("✗")+" failed"),
	)

	// Message
	messageSection := ""
	if m.message != "" {
		messageSection = m.theme.SuccessStyle.Render(m.message)
	}
	if m.err != nil {
		messageSection = m.theme.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Actions • r: Refresh • Esc: Back")

	sections := []string{header, menu, legend}
	if messageSection != "" {
		sections = append(sections, "", messageSection)
	}
	sections = append(sections, "", help)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	bordered := m.theme.BorderStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

// viewActions renders the actions menu
func (m FrankenPHPServicesModel) viewActions() string {
	if len(m.services) == 0 || m.cursor >= len(m.services) {
		return m.viewList()
	}

	service := m.services[m.cursor]
	header := m.theme.Title.Render(fmt.Sprintf("Actions: %s", service.Name))

	// Service info
	info := lipgloss.JoinVertical(lipgloss.Left,
		m.theme.Label.Render("Status: ")+m.getStatusStyle(service.Status),
		m.theme.Label.Render("Enabled: ")+m.theme.InfoStyle.Render(fmt.Sprintf("%v", service.Enabled)),
	)

	// Actions menu
	var items []string
	items = append(items, "")

	for i, action := range m.actions {
		cursor := "  "
		if i == m.actionCursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		var renderedItem string
		if i == m.actionCursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, action))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, action))
		}
		items = append(items, renderedItem)
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, items...)

	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Select • Esc: Back")

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", info, menu, "", help)
	bordered := m.theme.BorderStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

// viewEdit renders the edit form
func (m FrankenPHPServicesModel) viewEdit() string {
	if len(m.services) == 0 || m.cursor >= len(m.services) {
		return m.viewList()
	}

	service := m.services[m.cursor]
	header := m.theme.Title.Render(fmt.Sprintf("Edit: %s", service.Name))

	formView := ""
	if m.editForm != nil {
		formView = m.editForm.View()
	}

	help := m.theme.Help.Render("Tab: Next field • Shift+Tab: Previous • Enter: Save • Esc: Cancel")

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", formView, "", help)
	bordered := m.theme.BorderStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

// viewConfirm renders the confirmation dialog
func (m FrankenPHPServicesModel) viewConfirm() string {
	header := m.theme.Title.Render("Confirm Action")

	message := m.theme.WarningStyle.Render(m.confirmMsg)

	options := lipgloss.JoinVertical(lipgloss.Left,
		"",
		m.theme.MenuItem.Render("  Press 'y' to confirm"),
		m.theme.MenuItem.Render("  Press 'n' or Esc to cancel"),
	)

	help := m.theme.Help.Render("y: Yes • n/Esc: No")

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", message, options, "", help)
	bordered := m.theme.BorderStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

// getStatusStyle returns styled status text
func (m FrankenPHPServicesModel) getStatusStyle(status string) string {
	switch status {
	case "active":
		return m.theme.SuccessStyle.Render("running")
	case "inactive":
		return m.theme.WarningStyle.Render("stopped")
	case "failed":
		return m.theme.ErrorStyle.Render("failed")
	default:
		return m.theme.DescriptionStyle.Render(status)
	}
}
