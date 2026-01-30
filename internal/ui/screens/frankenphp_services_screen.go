package screens

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/stubs"
	"github.com/iperamuna/ravact/internal/system"
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
	Docroot     string
	Port        string
	User        string
	ConnType    string // "socket" or "port"
}

// FPServicesState represents the current state of the screen
type FPServicesState int

const (
	FPServicesStateList FPServicesState = iota
	FPServicesStateActions
	FPServicesStateEdit
	FPServicesStateReview
	FPServicesStateConfirmAction
	FPServicesStateConfirmDeploy
	FPServicesStateExecuting
	FPServicesStateNginxSelect
	FPServicesStateNginxView
	FPServicesStateEditFileSelect
)

// EditableFile represents a file that can be edited
type EditableFile struct {
	Name string
	Path string
}

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
	editConnType string
	editPort     string
	editUser     string
	editGroup    string
	editBinary   string // Added this

	// Detailed PHP INI fields
	editPHPMemoryLimit              string
	editPHPMaxExecutionTime         string
	editPHPOpcacheEnable            bool
	editPHPOpcacheEnableCli         bool
	editPHPOpcacheMemoryConsumption string
	editPHPOpcacheInternedStrings   string
	editPHPOpcacheMaxFiles          string
	editPHPOpcacheValidate          bool
	editPHPOpcacheRevalidateFreq    string
	editPHPOpcacheJit               bool
	editPHPOpcacheJitBufferSize     string
	editPHPRealpathCacheSize        string
	editPHPRealpathCacheTtl         string

	// Caddy settings
	editNumThreads  string
	editMaxThreads  string
	editMaxWaitTime string

	// Deployment data
	generatedFiles []GeneratedFile
	fileCursor     int
	fullCommand    string

	// Confirm action
	confirmAction string
	confirmMsg    string

	// Filtering
	filterDir string

	// Messages
	detector *system.Detector
	err      error
	message  string

	// Nginx View
	nginxForm   *huh.Form
	viewContent string
	viewTitle   string

	// File Selection for Editor
	editableFiles  []EditableFile
	editFileCursor int

	// Toggles
	showFullHelp bool
}

// NewFrankenPHPServicesModel creates a new FrankenPHP services model
func NewFrankenPHPServicesModel() FrankenPHPServicesModel {
	return NewFrankenPHPServicesModelWithFilter("")
}

// NewFrankenPHPServicesModelWithFilter creates a new FrankenPHP services model with a directory filter
func NewFrankenPHPServicesModelWithFilter(filterDir string) FrankenPHPServicesModel {
	t := theme.DefaultTheme()

	m := FrankenPHPServicesModel{
		theme:        t,
		detector:     system.NewDetector(),
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
			"View Nginx Config",
			"Delete Service",
			"← Back to List",
		},
		filterDir: filterDir,
	}

	// Load services
	m.services = m.loadFrankenPHPServices()

	// Apply filter if provided
	if filterDir != "" {
		var filtered []FrankenPHPService
		normalizedFilter := strings.TrimSuffix(filterDir, "/")
		for _, s := range m.services {
			normalizedSiteRoot := strings.TrimSuffix(s.SiteRoot, "/")
			if normalizedSiteRoot == normalizedFilter {
				filtered = append(filtered, s)
			}
		}
		m.services = filtered

		// If exactly one service found for this dir, auto-select it
		if len(m.services) == 1 {
			m.state = FPServicesStateActions
		}
	}

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
		config := m.parseServiceFileDetailed(line)
		service.SiteRoot = config.SiteRoot
		service.Docroot = config.Docroot
		service.Port = config.Port
		service.User = config.User
		service.ConnType = config.ConnType

		services = append(services, service)
	}

	return services
}

// ServiceConfig holds parsed service configuration
type ServiceConfig struct {
	SiteRoot string
	Docroot  string
	Port     string
	User     string
	Group    string
	ConnType string // "socket", "port", or "both"
}

// parseServiceFile extracts configuration from a service file
func (m *FrankenPHPServicesModel) parseServiceFile(path string) (siteRoot, port, user string) {
	config := m.parseServiceFileDetailed(path)
	return config.SiteRoot, config.Port, config.User
}

// parseServiceFileDetailed extracts full configuration from a service file
func (m *FrankenPHPServicesModel) parseServiceFileDetailed(path string) ServiceConfig {
	config := ServiceConfig{}

	cmd := exec.Command("cat", path)
	output, err := cmd.Output()
	if err != nil {
		return config
	}

	content := string(output)
	lines := strings.Split(content, "\n")

	cleanPath := func(p string) string {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, "\"")
		p = strings.Trim(p, "'")
		p = strings.TrimSuffix(p, "/")
		return p
	}

	hasSocket := false
	hasPort := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			val := parts[1]

			switch key {
			case "WorkingDirectory":
				config.SiteRoot = cleanPath(val)
			case "User":
				config.User = cleanPath(val)
			case "Group":
				config.Group = cleanPath(val)
			}
		}

		// Parse ExecStart for inline arguments
		if strings.Contains(line, "ExecStart=") {
			// Extract docroot
			if strings.Contains(line, "--root") {
				parts := strings.Split(line, "--root")
				if len(parts) >= 2 {
					docPart := strings.TrimSpace(parts[1])
					docParts := strings.Fields(docPart)
					if len(docParts) > 0 {
						config.Docroot = strings.TrimSuffix(docParts[0], "\\")
					}
				}
			}

			// Extract listen/port
			if strings.Contains(line, "--listen") {
				parts := strings.Split(line, "--listen")
				if len(parts) >= 2 {
					listenPart := strings.TrimSpace(parts[1])
					listenParts := strings.Fields(listenPart)
					if len(listenParts) > 0 {
						val := listenParts[0]
						if strings.Contains(val, "unix:") || strings.Contains(val, "unix/") {
							hasSocket = true
						} else if strings.Contains(val, ":") {
							hasPort = true
							portParts := strings.Split(val, ":")
							config.Port = portParts[len(portParts)-1]
						}
					}
				}
			}
		}
	}

	// Determine connection type
	if hasSocket {
		config.ConnType = "socket"
	} else if hasPort {
		config.ConnType = "port"
	} else {
		config.ConnType = "socket" // Default
	}

	return config
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
		// Clear messages on any key (except in specific states)
		if m.state != FPServicesStateEdit && m.state != FPServicesStateReview && (m.message != "" || m.err != nil) {
			m.message = ""
			m.err = nil
		}

		switch m.state {
		case FPServicesStateList:
			return m.updateList(msg)
		case FPServicesStateActions:
			return m.updateActions(msg)
		case FPServicesStateEdit:
			// Handle escape to cancel edit
			if msg.String() == "esc" {
				m.state = FPServicesStateActions
				m.editForm = nil
				return m, nil
			}
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		case FPServicesStateReview:
			return m.updateReview(msg)
		case FPServicesStateConfirmAction:
			return m.updateConfirm(msg)
		case FPServicesStateConfirmDeploy:
			return m.updateConfirmDeploy(msg)
		case FPServicesStateNginxSelect:
			return m.updateNginxSelect(msg)
		case FPServicesStateNginxView:
			return m.updateNginxView(msg)
		case FPServicesStateEditFileSelect:
			return m.updateEditFileSelect(msg)
		}
	}

	// Update form in Nginx Select state
	if m.state == FPServicesStateNginxSelect && m.nginxForm != nil {
		form, cmd := m.nginxForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.nginxForm = f
		}
		if m.nginxForm.State == huh.StateCompleted {
			return m.generateNginxForView()
		}
		return m, cmd
	}

	// Update form if in edit state
	if m.state == FPServicesStateEdit && m.editForm != nil {
		form, cmd := m.editForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.editForm = f
		}

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
				Command:     fmt.Sprintf("sudo systemctl start %s && sudo systemctl status %s --no-pager -l", service.Name, service.Name),
				Description: fmt.Sprintf("Starting %s", service.Name),
			}
		}

	case "Stop Service":
		m.confirmAction = "stop"
		m.confirmMsg = fmt.Sprintf("Stop service %s?", service.Name)
		m.state = FPServicesStateConfirmAction
		return m, nil

	case "Restart Service":
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     fmt.Sprintf("sudo systemctl restart %s && sudo systemctl status %s --no-pager -l", service.Name, service.Name),
				Description: fmt.Sprintf("Restarting %s", service.Name),
			}
		}

	case "Enable (start on boot)":
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     fmt.Sprintf("sudo systemctl enable %s && echo '✓ Service enabled'", service.Name),
				Description: fmt.Sprintf("Enabling %s", service.Name),
			}
		}

	case "Disable (don't start on boot)":
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     fmt.Sprintf("sudo systemctl disable %s && echo '✓ Service disabled'", service.Name),
				Description: fmt.Sprintf("Disabling %s", service.Name),
			}
		}

	case "View Status":
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     fmt.Sprintf("sudo systemctl status %s --no-pager -l", service.Name),
				Description: fmt.Sprintf("Status of %s", service.Name),
			}
		}

	case "View Logs":
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     fmt.Sprintf("sudo journalctl -u %s -n 100 --no-pager", service.Name),
				Description: fmt.Sprintf("Logs for %s", service.Name),
			}
		}

	case "Edit Configuration (Form)":
		m.state = FPServicesStateEdit
		m.loadServiceForEdit(service)
		m.editForm = m.buildEditForm()
		return m, m.editForm.Init()

	case "Edit Configuration (Editor)":
		m.state = FPServicesStateEditFileSelect
		m.editFileCursor = 0
		m.editableFiles = []EditableFile{
			{Name: "Caddyfile", Path: fmt.Sprintf("/etc/frankenphp/%s/Caddyfile", service.SiteKey)},
			{Name: "Systemd Service", Path: service.ServiceFile},
			{Name: "Nginx Config", Path: fmt.Sprintf("/etc/nginx/sites-available/%s.conf", service.SiteKey)},
		}
		return m, nil

	case "View Nginx Config":
		m.state = FPServicesStateNginxSelect
		m.nginxForm = m.buildNginxSelectForm()
		return m, m.nginxForm.Init()

	case "Delete Service":
		m.confirmAction = "delete"
		m.confirmMsg = fmt.Sprintf("Delete service %s? This will stop the service and remove configuration files.", service.Name)
		m.state = FPServicesStateConfirmAction
		return m, nil

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
				Command:     fmt.Sprintf("sudo systemctl stop %s && echo '✓ Service stopped'", service.Name),
				Description: fmt.Sprintf("Stopping %s", service.Name),
			}
		}

	case "delete":
		m.state = FPServicesStateList
		return m, func() tea.Msg {
			// Construct deletion command
			var cmds []string
			cmds = append(cmds, fmt.Sprintf("sudo systemctl stop %s", service.Name))
			cmds = append(cmds, fmt.Sprintf("sudo systemctl disable %s", service.Name))
			cmds = append(cmds, fmt.Sprintf("sudo rm -f %s", service.ServiceFile))
			cmds = append(cmds, fmt.Sprintf("sudo rm -rf /etc/frankenphp/%s", service.SiteKey))
			cmds = append(cmds, fmt.Sprintf("sudo rm -f /etc/nginx/sites-available/%s.conf", service.SiteKey))
			cmds = append(cmds, fmt.Sprintf("sudo rm -f /etc/nginx/sites-enabled/%s.conf", service.SiteKey))
			cmds = append(cmds, "sudo systemctl daemon-reload")
			// Try to remove socket file if it exists, but don't fail if it doesn't
			cmds = append(cmds, fmt.Sprintf("sudo rm -f /run/frankenphp/%s.sock", service.SiteKey))

			fullCmd := strings.Join(cmds, " && ") + " && echo '✓ Service deleted'"

			return ExecutionStartMsg{
				Command:     fullCmd,
				Description: fmt.Sprintf("Deleting %s", service.Name),
			}
		}
	}

	m.state = FPServicesStateActions
	return m, nil
}

// loadServiceForEdit loads service config into edit form fields
func (m *FrankenPHPServicesModel) loadServiceForEdit(service FrankenPHPService) {
	// Re-parse the service file to ensure we have fresh data
	config := m.parseServiceFileDetailed(service.ServiceFile)

	// Set foundational fields
	m.editSiteRoot = config.SiteRoot
	m.editUser = config.User
	m.editGroup = config.Group
	if m.editGroup == "" {
		m.editGroup = m.editUser
	}

	// These will act as fallbacks if Caddyfile is missing or incomplete
	m.editDocroot = config.Docroot
	m.editPort = config.Port
	m.editConnType = config.ConnType

	// Load Caddyfile settings (will fill Docroot, Port, ConnType, PHP settings)
	caddyfilePath := fmt.Sprintf("/etc/frankenphp/%s/Caddyfile", service.SiteKey)
	m.loadCaddyfileForEdit(caddyfilePath)

	// Final Docroot cleanup
	if m.editDocroot == "" {
		publicPath := filepath.Join(m.editSiteRoot, "public")
		if _, err := exec.Command("test", "-d", publicPath).Output(); err == nil {
			m.editDocroot = "public"
		}
	}

	// Final Port/ConnType defaults
	if m.editPort == "" && m.editConnType == "port" {
		m.editPort = "8000"
	}
	if m.editConnType == "" {
		m.editConnType = "socket"
	}

	// Reset domains
	m.editDomains = ""

	// Try to read nginx config for domains
	nginxConfPath := fmt.Sprintf("/etc/nginx/sites-available/%s.conf", service.SiteKey)
	cmd := exec.Command("bash", "-c", fmt.Sprintf("grep -oP 'server_name \\K[^;]+' %s 2>/dev/null || true", nginxConfPath))
	output, _ := cmd.Output()
	m.editDomains = strings.TrimSpace(string(output))
}

func (m *FrankenPHPServicesModel) loadCaddyfileForEdit(path string) {
	// Defaults for Caddy
	m.editNumThreads = "8"
	m.editMaxThreads = "auto"
	m.editMaxWaitTime = "15"

	// Defaults for PHP
	m.editPHPMemoryLimit = "256M"
	m.editPHPMaxExecutionTime = "30"
	m.editPHPOpcacheEnable = true
	m.editPHPOpcacheEnableCli = true
	m.editPHPOpcacheMemoryConsumption = "512"
	m.editPHPOpcacheInternedStrings = "32"
	m.editPHPOpcacheMaxFiles = "100000"
	m.editPHPOpcacheValidate = false
	m.editPHPOpcacheRevalidateFreq = "0"
	m.editPHPOpcacheJit = false
	m.editPHPOpcacheJitBufferSize = "0"
	m.editPHPRealpathCacheSize = "4096K"
	m.editPHPRealpathCacheTtl = "600"

	cmd := exec.Command("cat", path)
	output, err := cmd.Output()
	if err != nil {
		return
	}

	content := string(output)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "num_threads") {
			m.editNumThreads = strings.TrimSpace(strings.TrimPrefix(line, "num_threads"))
		} else if strings.HasPrefix(line, "max_threads") {
			m.editMaxThreads = strings.TrimSpace(strings.TrimPrefix(line, "max_threads"))
		} else if strings.HasPrefix(line, "max_wait_time") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "max_wait_time"))
			m.editMaxWaitTime = strings.TrimSuffix(val, "s")
		} else if strings.HasPrefix(line, "root *") {
			rootVal := strings.TrimSpace(strings.TrimPrefix(line, "root *"))
			if rootVal != "" {
				if strings.HasPrefix(rootVal, m.editSiteRoot) {
					relPath := strings.TrimPrefix(rootVal, m.editSiteRoot)
					m.editDocroot = strings.TrimLeft(relPath, "/")
				} else {
					m.editDocroot = rootVal
				}
			}
		} else if strings.HasPrefix(line, "bind ") {
			// Format: bind unix//run/frankenphp/name.sock
			// Or: bind 127.0.0.1:8000
			val := strings.TrimSpace(strings.TrimPrefix(line, "bind "))
			if strings.Contains(val, "unix://") || strings.Contains(val, "unix/") {
				m.editConnType = "socket"
			} else if strings.Contains(val, ":") {
				m.editConnType = "port"
				parts := strings.Split(val, ":")
				if len(parts) >= 2 {
					m.editPort = parts[len(parts)-1]
				}
			}
		} else if strings.HasPrefix(line, "php_ini") {
			// Format: php_ini key value
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				key := parts[1]
				val := parts[2]
				switch key {
				case "memory_limit":
					m.editPHPMemoryLimit = val
				case "max_execution_time":
					m.editPHPMaxExecutionTime = val
				case "opcache.enable":
					m.editPHPOpcacheEnable = val == "1"
				case "opcache.enable_cli":
					m.editPHPOpcacheEnableCli = val == "1"
				case "opcache.memory_consumption":
					m.editPHPOpcacheMemoryConsumption = val
				case "opcache.interned_strings_buffer":
					m.editPHPOpcacheInternedStrings = val
				case "opcache.max_accelerated_files":
					m.editPHPOpcacheMaxFiles = val
				case "opcache.validate_timestamps":
					m.editPHPOpcacheValidate = val == "1"
				case "opcache.revalidate_freq":
					m.editPHPOpcacheRevalidateFreq = val
				case "opcache.jit":
					m.editPHPOpcacheJit = val != "0" && val != "off"
				case "opcache.jit_buffer_size":
					m.editPHPOpcacheJitBufferSize = val
				case "realpath_cache_size":
					m.editPHPRealpathCacheSize = val
				case "realpath_cache_ttl":
					m.editPHPRealpathCacheTtl = val
				}
			}
		}
	}
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
				Title("Web Accessible Directory").
				Description("Document root (relative to site root, e.g. 'public')").
				Placeholder("public").
				Value(&m.editDocroot),

			huh.NewInput().
				Key("domains").
				Title("Domain Names").
				Description("Space-separated domain names").
				Placeholder("example.com www.example.com").
				Value(&m.editDomains),

			huh.NewSelect[string]().
				Key("connType").
				Title("How Nginx Connects to FrankenPHP").
				Description("Connection method between Nginx and FrankenPHP").
				Options(
					huh.NewOption("Unix Socket (recommended)", "socket"),
					huh.NewOption("TCP Port", "port"),
				).
				Value(&m.editConnType),

			huh.NewInput().
				Key("port").
				Title("Port").
				Description("TCP port for FrankenPHP (used when connection type is Port)").
				Placeholder("8000").
				Validate(func(s string) error {
					if s == "" {
						return nil
					}
					port, err := strconv.Atoi(s)
					if err != nil {
						return fmt.Errorf("port must be a number")
					}
					if port < 1 || port > 65535 {
						return fmt.Errorf("port must be between 1 and 65535")
					}
					return nil
				}).
				Value(&m.editPort),

			huh.NewInput().
				Key("user").
				Title("Run as User").
				Description("System user to run the FrankenPHP service").
				Placeholder("www-data").
				Value(&m.editUser),

			huh.NewInput().
				Key("group").
				Title("Run as Group").
				Description("System group to run the FrankenPHP service").
				Placeholder("www-data").
				Value(&m.editGroup),

			huh.NewInput().
				Key("binary").
				Title("FrankenPHP Binary Path").
				Description("Full path to the frankenphp binary").
				Placeholder("/usr/local/bin/frankenphp").
				Value(&m.editBinary),
		),

		huh.NewGroup(
			huh.NewInput().
				Key("numThreads").
				Title("Number of Threads").
				Description("Suggestion: System Threads * 2").
				Placeholder("8").
				Validate(func(s string) error {
					if _, err := strconv.Atoi(s); err != nil {
						return fmt.Errorf("must be a number")
					}
					return nil
				}).
				Value(&m.editNumThreads),

			huh.NewInput().
				Key("maxThreads").
				Title("Max Threads").
				Description("Accepts positive integer > Number of Threads, or 'auto'").
				Placeholder("auto").
				Validate(func(s string) error {
					if s == "auto" {
						return nil
					}
					v, err := strconv.Atoi(s)
					if err != nil {
						return fmt.Errorf("must be a number or 'auto'")
					}
					num, _ := strconv.Atoi(m.editNumThreads)
					if v <= num {
						return fmt.Errorf("must be greater than Number of Threads (%d)", num)
					}
					return nil
				}).
				Value(&m.editMaxThreads),

			huh.NewInput().
				Key("maxWaitTime").
				Title("Max Wait Time").
				Description("Max time to wait for a thread (in seconds)").
				Placeholder("15").
				Validate(func(s string) error {
					if _, err := strconv.Atoi(s); err != nil {
						return fmt.Errorf("must be a number")
					}
					return nil
				}).
				Value(&m.editMaxWaitTime),
		).Title("Performance Tuning"),

		huh.NewGroup(
			huh.NewInput().
				Key("memoryLimit").
				Title("PHP memory_limit").
				Placeholder("256M").
				Value(&m.editPHPMemoryLimit),

			huh.NewInput().
				Key("maxExecTime").
				Title("PHP max_execution_time").
				Placeholder("30").
				Value(&m.editPHPMaxExecutionTime),

			huh.NewConfirm().
				Key("opcacheEnable").
				Title("Enable OPcache").
				Value(&m.editPHPOpcacheEnable),

			huh.NewConfirm().
				Key("opcacheCli").
				Title("Enable OPcache CLI").
				Value(&m.editPHPOpcacheEnableCli),

			huh.NewInput().
				Key("opcacheMemory").
				Title("OPcache Memory Consumption (MB)").
				Placeholder("512").
				Value(&m.editPHPOpcacheMemoryConsumption),

			huh.NewInput().
				Key("opcacheStrings").
				Title("OPcache Interned Strings Buffer").
				Placeholder("32").
				Value(&m.editPHPOpcacheInternedStrings),

			huh.NewInput().
				Key("opcacheMaxFiles").
				Title("OPcache Max Accelerated Files").
				Placeholder("100000").
				Value(&m.editPHPOpcacheMaxFiles),

			huh.NewConfirm().
				Key("opcacheValidate").
				Title("OPcache Validate Timestamps").
				Description("Set to false for production optimization").
				Value(&m.editPHPOpcacheValidate),

			huh.NewInput().
				Key("opcacheFreq").
				Title("OPcache Revalidate Frequency").
				Placeholder("0").
				Value(&m.editPHPOpcacheRevalidateFreq),

			huh.NewConfirm().
				Key("jit").
				Title("Enable JIT").
				Value(&m.editPHPOpcacheJit),

			huh.NewInput().
				Key("jitBuffer").
				Title("JIT Buffer Size").
				Placeholder("0").
				Value(&m.editPHPOpcacheJitBufferSize),

			huh.NewInput().
				Key("realpathSize").
				Title("Realpath Cache Size").
				Placeholder("4096K").
				Value(&m.editPHPRealpathCacheSize),

			huh.NewInput().
				Key("realpathTtl").
				Title("Realpath Cache TTL").
				Placeholder("600").
				Value(&m.editPHPRealpathCacheTtl),
		).Title("PHP INIT - Core & Opcashe & Realpath"),
	).WithTheme(m.theme.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)
}

// saveServiceConfig prepares for review
func (m FrankenPHPServicesModel) saveServiceConfig() (tea.Model, tea.Cmd) {
	if len(m.services) == 0 || m.cursor >= len(m.services) {
		m.state = FPServicesStateList
		m.editForm = nil
		m.err = fmt.Errorf("no service selected")
		return m, nil
	}

	m.state = FPServicesStateReview
	m.fileCursor = 0
	return m.generateConfigFiles(), nil
}

// generateConfigFiles generates the content for all relevant config files
func (m FrankenPHPServicesModel) generateConfigFiles() FrankenPHPServicesModel {
	m.generatedFiles = []GeneratedFile{}
	service := m.services[m.cursor]
	id := service.SiteKey

	// 1. Caddyfile
	caddyTemplate := m.generateCaddyfileContent()
	m.generatedFiles = append(m.generatedFiles, GeneratedFile{
		Name:    "Caddyfile",
		Path:    fmt.Sprintf("/etc/frankenphp/%s/Caddyfile", id),
		Content: caddyTemplate,
	})

	// 2. Systemd Service
	serviceTemplate := m.generateServiceFileContent()
	m.generatedFiles = append(m.generatedFiles, GeneratedFile{
		Name:    "Systemd Service",
		Path:    fmt.Sprintf("/etc/systemd/system/frankenphp-%s.service", id),
		Content: serviceTemplate,
	})

	// 3. Nginx Config
	nginxTemplate := m.generateNginxContent()
	m.generatedFiles = append(m.generatedFiles, GeneratedFile{
		Name:    "Nginx Config",
		Path:    fmt.Sprintf("/etc/nginx/sites-available/%s.conf", id),
		Content: nginxTemplate,
	})

	// 4. fpcli Wrapper
	fpcliTemplate := m.generateFpcliContent()
	m.generatedFiles = append(m.generatedFiles, GeneratedFile{
		Name:    "fpcli Wrapper",
		Path:    "/usr/local/bin/fpcli",
		Content: fpcliTemplate,
	})

	return m
}

func (m FrankenPHPServicesModel) generateCaddyfileContent() string {
	id := m.services[m.cursor].SiteKey
	docroot := m.getFullDocroot()
	port := m.editPort
	if port == "" {
		port = "8000"
	}

	var bindLine string
	if m.editConnType == "socket" {
		bindLine = fmt.Sprintf("bind unix//run/frankenphp/%s.sock", id)
	} else {
		bindLine = fmt.Sprintf("bind 127.0.0.1:%s", port)
	}

	// Build PHP directives
	var phpDirectives strings.Builder
	settings := map[string]string{
		"memory_limit":                    m.editPHPMemoryLimit,
		"max_execution_time":              m.editPHPMaxExecutionTime,
		"opcache.enable":                  "0",
		"opcache.enable_cli":              "0",
		"opcache.memory_consumption":      m.editPHPOpcacheMemoryConsumption,
		"opcache.interned_strings_buffer": m.editPHPOpcacheInternedStrings,
		"opcache.max_accelerated_files":   m.editPHPOpcacheMaxFiles,
		"opcache.validate_timestamps":     "0",
		"opcache.revalidate_freq":         m.editPHPOpcacheRevalidateFreq,
		"opcache.jit":                     "0",
		"opcache.jit_buffer_size":         m.editPHPOpcacheJitBufferSize,
		"realpath_cache_size":             m.editPHPRealpathCacheSize,
		"realpath_cache_ttl":              m.editPHPRealpathCacheTtl,
	}

	if m.editPHPOpcacheEnable {
		settings["opcache.enable"] = "1"
	}
	if m.editPHPOpcacheEnableCli {
		settings["opcache.enable_cli"] = "1"
	}
	if m.editPHPOpcacheValidate {
		settings["opcache.validate_timestamps"] = "1"
	}
	if m.editPHPOpcacheJit {
		settings["opcache.jit"] = "1255"
	}

	keys := []string{
		"memory_limit", "max_execution_time", "opcache.enable", "opcache.enable_cli",
		"opcache.memory_consumption", "opcache.interned_strings_buffer", "opcache.max_accelerated_files",
		"opcache.validate_timestamps", "opcache.revalidate_freq", "opcache.jit",
		"opcache.jit_buffer_size", "realpath_cache_size", "realpath_cache_ttl",
	}

	for _, k := range keys {
		if v, ok := settings[k]; ok && v != "" {
			phpDirectives.WriteString(fmt.Sprintf("\t\tphp_ini %s %s\n", k, v))
		}
	}

	content, _ := stubs.LoadAndReplace("caddyfile", map[string]string{
		"SITE_KEY":       id,
		"NUM_THREADS":    m.editNumThreads,
		"MAX_THREADS":    m.editMaxThreads,
		"MAX_WAIT_TIME":  m.editMaxWaitTime,
		"PORT":           port,
		"BIND_LINE":      bindLine,
		"DOCROOT":        docroot,
		"PHP_DIRECTIVES": strings.TrimSpace(phpDirectives.String()),
	})

	return content
}

func (m FrankenPHPServicesModel) generateServiceFileContent() string {
	id := m.services[m.cursor].SiteKey
	siteRoot := m.editSiteRoot
	user := m.editUser
	group := m.editGroup
	binary := m.editBinary
	if binary == "" {
		binary = "/usr/local/bin/frankenphp"
	}

	var preStart string
	var postStart string
	if m.editConnType == "socket" {
		preStart = fmt.Sprintf("ExecStartPre=/usr/bin/rm -f /run/frankenphp/%s.sock\n", id)
		postStart = fmt.Sprintf("ExecStartPost=/bin/sh -c 'for i in $(seq 1 50); do [ -S /run/frankenphp/%s.sock ] && chmod 0660 /run/frankenphp/%s.sock && exit 0; sleep 0.1; done; echo \"Socket not created: /run/frankenphp/%s.sock\" >&2; exit 1'\n", id, id, id)
	}

	caddyfile := fmt.Sprintf("/etc/frankenphp/%s/Caddyfile", id)

	content, _ := stubs.LoadAndReplace("service", map[string]string{
		"ID":                id,
		"USER":              user,
		"GROUP":             group,
		"WORKING_DIRECTORY": siteRoot,
		"APP_BASE_PATH":     siteRoot,
		"PRE_START":         preStart,
		"BINARY":            binary,
		"CADDYFILE":         caddyfile,
		"POST_START":        postStart,
	})

	return content
}

func (m FrankenPHPServicesModel) generateNginxContent() string {
	id := m.services[m.cursor].SiteKey
	domains := m.editDomains
	port := m.editPort
	if port == "" {
		port = "8000"
	}

	var upstream string
	if m.editConnType == "socket" {
		upstream = fmt.Sprintf("unix:/run/frankenphp/%s.sock", id)
	} else {
		upstream = fmt.Sprintf("127.0.0.1:%s", port)
	}

	content, _ := stubs.LoadAndReplace("nginx", map[string]string{
		"DOMAINS":  domains,
		"UPSTREAM": upstream,
		"SITE_KEY": id,
	})

	return content
}

func (m FrankenPHPServicesModel) getFullDocroot() string {
	if m.editDocroot == "" {
		return m.editSiteRoot
	}
	if strings.HasPrefix(m.editDocroot, "/") {
		return m.editDocroot
	}
	return filepath.Join(m.editSiteRoot, m.editDocroot)
}

func (m FrankenPHPServicesModel) buildDeployCommand() string {
	service := m.services[m.cursor]
	siteKey := service.SiteKey
	user := m.editUser
	group := m.editGroup

	var script strings.Builder
	script.WriteString("#!/bin/bash\nset -e\n\n")
	script.WriteString(fmt.Sprintf("echo \"Updating FrankenPHP Service: %s\"\n", service.Name))

	// Create storage directory
	// Create storage directory structure
	script.WriteString(fmt.Sprintf("\nsudo mkdir -p /var/lib/caddy/%s/config\n", siteKey))
	script.WriteString(fmt.Sprintf("sudo mkdir -p /var/lib/caddy/%s/data\n", siteKey))
	script.WriteString(fmt.Sprintf("sudo mkdir -p /var/lib/caddy/%s/tls\n", siteKey))

	// Set permissions
	script.WriteString(fmt.Sprintf("sudo chown -R %s:%s /var/lib/caddy/%s\n", user, user, siteKey))
	script.WriteString(fmt.Sprintf("sudo chmod -R 750 /var/lib/caddy/%s\n", siteKey))

	// Write generated files
	for _, file := range m.generatedFiles {
		script.WriteString(fmt.Sprintf("\nif [ -f \"%s\" ]; then\n", file.Path))
		script.WriteString(fmt.Sprintf("    cp \"%s\" \"%s.bak\"\n", file.Path, file.Path))
		script.WriteString("fi\n")
		script.WriteString(fmt.Sprintf("cat > \"%s\" <<'EOF'\n", file.Path))
		script.WriteString(file.Content)
		script.WriteString("\nEOF\n")
	}

	// Fix permissions and restart
	binary := m.editBinary
	if binary == "" {
		binary = "/usr/local/bin/frankenphp"
	}
	caddyfilePath := fmt.Sprintf("/etc/frankenphp/%s/Caddyfile", siteKey)
	script.WriteString(fmt.Sprintf("\n%s fmt --overwrite %s\n", binary, caddyfilePath))

	// Fix permissions on config directory before restart
	script.WriteString(fmt.Sprintf("sudo chown -R %s:%s /etc/frankenphp/%s\n", user, group, siteKey))

	script.WriteString("\nsudo systemctl daemon-reload\n")
	script.WriteString(fmt.Sprintf("sudo systemctl restart %s\n", service.Name))

	// Verification
	script.WriteString("\nset +e\n")
	script.WriteString(fmt.Sprintf("if sudo systemctl is-active --quiet %s; then\n", service.Name))
	script.WriteString(fmt.Sprintf("    echo \"✓ Service %s restarted successfully\"\n", service.Name))
	script.WriteString("else\n")
	script.WriteString(fmt.Sprintf("    echo \"✗ Service %s failed to restart!\"\n", service.Name))
	script.WriteString(fmt.Sprintf("    sudo systemctl status %s --no-pager -l\n", service.Name))
	script.WriteString("fi\n")

	return script.String()
}

func (m FrankenPHPServicesModel) updateReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.fileCursor > 0 {
			m.fileCursor--
		}
	case "down", "j":
		if m.fileCursor < len(m.generatedFiles)-1 {
			m.fileCursor++
		}
	case "enter":
		m.state = FPServicesStateConfirmDeploy
	case "esc":
		m.state = FPServicesStateEdit
	case "v":
		// Navigate to editor for the selected file
		if len(m.generatedFiles) > 0 {
			file := m.generatedFiles[m.fileCursor]
			return m, func() tea.Msg {
				return NavigateMsg{
					Screen: EditorSelectionScreen,
					Data: map[string]interface{}{
						"file":        file.Path,
						"description": fmt.Sprintf("Edit %s", file.Name),
					},
				}
			}
		}
	}
	return m, nil
}

func (m FrankenPHPServicesModel) updateConfirmDeploy(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		m.state = FPServicesStateExecuting
		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     m.buildDeployCommand(),
				Description: "Deploying FrankenPHP Configuration Changes",
			}
		}
	case "n", "N", "esc":
		m.state = FPServicesStateReview
	}
	return m, nil
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
	case FPServicesStateReview:
		return m.viewReview()
	case FPServicesStateConfirmAction:
		return m.viewConfirm()
	case FPServicesStateConfirmDeploy:
		return m.viewConfirmDeploy()
	case FPServicesStateExecuting:
		return "Deploying Changes..." // Execution screen will take over
	case FPServicesStateNginxSelect:
		return m.viewNginxSelection()
	case FPServicesStateNginxView:
		return m.viewNginxContent()
	case FPServicesStateEditFileSelect:
		return m.viewEditFileSelect()
	}

	return "Unknown state"
}

func (m FrankenPHPServicesModel) viewReview() string {
	header := m.theme.Title.Render("Review Changes")
	desc := m.theme.DescriptionStyle.Render("Review the generated configuration files below.")

	var items []string
	for i, file := range m.generatedFiles {
		prefix := "  "
		if i == m.fileCursor {
			prefix = m.theme.KeyStyle.Render("▶ ")
		}
		items = append(items, fmt.Sprintf("%s%s (%s)", prefix, file.Name, file.Path))
	}

	fileList := lipgloss.JoinVertical(lipgloss.Left, items...)

	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Deploy • v: View/Edit File • Esc: Back to Form")

	content := lipgloss.JoinVertical(lipgloss.Left, header, desc, "", fileList, "", help)
	bordered := m.theme.RenderBox(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

func (m FrankenPHPServicesModel) viewConfirmDeploy() string {
	header := m.theme.Title.Render("Confirm Deployment")
	warning := m.theme.WarningStyle.Render("This will overwrite existing configuration files and restart the service.")

	options := lipgloss.JoinVertical(lipgloss.Left,
		"",
		m.theme.MenuItem.Render("  Press 'y' to confirm and deploy"),
		m.theme.MenuItem.Render("  Press 'n' or Esc to cancel"),
	)

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", warning, options)
	bordered := m.theme.RenderBox(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

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
		bordered := m.theme.RenderBox(content)
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

		userStr := ""
		if svc.User != "" {
			userStr = m.theme.DescriptionStyle.Render(fmt.Sprintf(" (%s)", svc.User))
		}
		name := fmt.Sprintf("%s %s%s%s", statusIndicator, svc.Name, enabledStr, userStr)

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
	bordered := m.theme.RenderBox(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

func (m FrankenPHPServicesModel) generateFpcliContent() string {
	binary := m.editBinary
	if binary == "" {
		binary = "/usr/local/bin/frankenphp"
	}

	content, _ := stubs.LoadAndReplace("fpcli", map[string]string{
		"BINARY": binary,
	})

	return content
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
		m.theme.Label.Render("User: ")+m.theme.DescriptionStyle.Render(service.User),
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
	bordered := m.theme.RenderBox(content)
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
	bordered := m.theme.RenderBox(content)
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
	bordered := m.theme.RenderBox(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

// buildNginxSelectForm builds the Nginx Selection Form
func (m *FrankenPHPServicesModel) buildNginxSelectForm() *huh.Form {
	service := m.services[m.cursor]

	// Determine default values
	defaultType := "socket"
	defaultParam := fmt.Sprintf("/run/frankenphp/%s.sock", service.SiteKey)

	// Load Config to check actual bind/port if not in memory
	// This helps if the service struct hasn't been fully populated from buildServiceConfig
	if service.ConnType == "" {
		m.loadServiceForEdit(service)
		service.ConnType = m.editConnType
		service.Port = m.editPort
	}

	// Determine default values
	defaultType = "socket"
	defaultParam = fmt.Sprintf("/run/frankenphp/%s.sock", service.SiteKey)

	if service.ConnType == "port" {
		defaultType = "port"
		defaultParam = service.Port
		if defaultParam == "" {
			defaultParam = "8000" // Fallback
		}
	} else {
		// Even for socket, make sure we use the one from config if available
		// But usually it follows the pattern.
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("connType").
				Title("Select Connection Type").
				Options(
					huh.NewOption("Unix Socket", "socket"),
					huh.NewOption("TCP Port", "port"),
				).
				Value(&defaultType), // Use Value instead of Default to ensure binding
			huh.NewInput().
				Key("param").
				Title("Socket Path or Port Number").
				Description("Enter the socket path or port number").
				Placeholder("e.g. 8000 or mysite (will use /run/frankenphp/mysite.sock)").
				Value(&defaultParam).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("required")
					}
					return nil
				}),
		),
	).WithTheme(m.theme.HuhTheme)
}

func (m FrankenPHPServicesModel) updateNginxSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// handled in main update loop for form
	return m, nil
}

func (m FrankenPHPServicesModel) generateNginxForView() (tea.Model, tea.Cmd) {
	connType := m.nginxForm.GetString("connType")
	param := m.nginxForm.GetString("param")

	service := m.services[m.cursor]

	var upstream string
	if connType == "socket" {
		// Clean up socket path
		if !strings.Contains(param, "/") {
			upstream = fmt.Sprintf("unix:/run/frankenphp/%s.sock", param)
		} else if strings.HasPrefix(param, "unix:") {
			upstream = param
		} else {
			upstream = fmt.Sprintf("unix:%s", param)
		}
	} else {
		// Port - clean it up
		cleanParam := strings.TrimPrefix(param, ":")
		// Check if it already has IP
		if strings.Contains(cleanParam, ":") {
			upstream = cleanParam
		} else {
			upstream = fmt.Sprintf("127.0.0.1:%s", cleanParam)
		}
	}

	content, _ := stubs.LoadAndReplace("nginx", map[string]string{
		"DOMAINS":  "your-domain.com",
		"UPSTREAM": upstream,
		"SITE_KEY": service.SiteKey,
	})

	m.viewContent = content
	m.viewTitle = fmt.Sprintf("Nginx Config (%s)", connType)
	m.state = FPServicesStateNginxView
	return m, nil
}

func (m FrankenPHPServicesModel) updateNginxView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.state = FPServicesStateActions
		m.message = ""
		return m, nil
	case "c":
		// Copy to clipboard
		go func() {
			cmd := exec.Command("pbcopy")
			cmd.Stdin = strings.NewReader(m.viewContent)
			_ = cmd.Run()
		}()
		m.message = "✓ Copied to clipboard"
		return m, nil
	}
	return m, nil
}

func (m FrankenPHPServicesModel) viewNginxSelection() string {
	if m.nginxForm == nil {
		return "Loading..."
	}

	header := m.theme.Title.Render("Configure Nginx View")
	form := m.nginxForm.View()
	content := lipgloss.JoinVertical(lipgloss.Left, header, "", form)
	bordered := m.theme.RenderBox(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}

func (m FrankenPHPServicesModel) viewNginxContent() string {
	header := m.theme.Title.Render(m.viewTitle)

	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Align(lipgloss.Left) // Ensure left alignment

	content := contentStyle.Render(m.viewContent)

	helpText := "c: Copy to Clipboard • q/Esc: Back"
	if m.message != "" {
		helpText = m.theme.SuccessStyle.Render(m.message) + " • " + helpText
	}
	help := m.theme.Help.Render(helpText)

	ui := lipgloss.JoinVertical(lipgloss.Center, header, content, help)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, ui)
}

func (m FrankenPHPServicesModel) updateEditFileSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c", "q"))):
		return m, tea.Quit
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		m.state = FPServicesStateActions
		return m, nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if m.editFileCursor > 0 {
			m.editFileCursor--
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if m.editFileCursor < len(m.editableFiles)-1 {
			m.editFileCursor++
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter", " "))):
		selectedFile := m.editableFiles[m.editFileCursor]
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: EditorSelectionScreen,
				Data: map[string]interface{}{
					"file":        selectedFile.Path,
					"description": fmt.Sprintf("Edit %s", selectedFile.Name),
				},
			}
		}
	}
	return m, nil
}

func (m FrankenPHPServicesModel) viewEditFileSelect() string {
	header := m.theme.Title.Render("Select File to Edit")

	var items []string
	for i, file := range m.editableFiles {
		cursor := "  "
		if i == m.editFileCursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		item := fmt.Sprintf("%s%s (%s)", cursor, file.Name, file.Path)
		if i == m.editFileCursor {
			item = m.theme.SelectedItem.Render(item)
		} else {
			item = m.theme.MenuItem.Render(item)
		}
		items = append(items, item)
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, items...)
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Edit • Esc: Back")

	content := lipgloss.JoinVertical(lipgloss.Left, header, "", menu, "", help)
	bordered := m.theme.RenderBox(content)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
}
