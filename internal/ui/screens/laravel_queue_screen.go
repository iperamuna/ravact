package screens

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// LaravelQueueService represents a systemd queue worker service
type LaravelQueueService struct {
	Name             string // full service name
	Label            string // short label
	User             string
	Group            string
	WorkingDirectory string
	Executor         string
	QueueName        string
	Sleep            string
	Tries            string
	Timeout          string
	Status           string // active, inactive, failed
	Enabled          bool
	ServiceCount     string // Number of instances to run
}

func (s LaravelQueueService) GetServiceCount() int {
	if s.ServiceCount == "" {
		return 1
	}
	var count int
	fmt.Sscanf(s.ServiceCount, "%d", &count)
	if count < 1 {
		return 1
	}
	return count
}

// LaravelQueueState represents the screen state
type LaravelQueueState int

const (
	QueueStateList LaravelQueueState = iota
	QueueStateForm
	QueueStateActions
)

// LaravelQueueModel manages Laravel queue services
type LaravelQueueModel struct {
	theme       *theme.Theme
	width       int
	height      int
	state       LaravelQueueState
	projectPath string
	systemUser  string // Default user recommendation

	services []LaravelQueueService
	cursor   int

	// Form
	form        *huh.Form
	isEditing   bool
	editService LaravelQueueService // Holding values for form

	// Actions
	actions      []string
	actionCursor int

	// Messages
	message string
	err     error
}

func NewLaravelQueueModel(projectPath, systemUser string) LaravelQueueModel {
	m := LaravelQueueModel{
		theme:       theme.DefaultTheme(),
		projectPath: projectPath,
		systemUser:  systemUser,
		state:       QueueStateList,
	}
	m.loadServices()
	return m
}

func (m *LaravelQueueModel) loadServices() {
	// Scan systemd services
	files, _ := filepath.Glob("/etc/systemd/system/*.service")
	var services []LaravelQueueService

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		sContent := string(content)

		// Basic check if it's relevant to this project
		if !strings.Contains(sContent, fmt.Sprintf("WorkingDirectory=%s", m.projectPath)) {
			continue
		}
		// Check if it's a queue worker
		if !strings.Contains(sContent, "queue:work") {
			continue
		}

		svc := m.parseService(filepath.Base(file), sContent)
		services = append(services, svc)
	}
	m.services = services
}

func (m *LaravelQueueModel) parseService(filename, content string) LaravelQueueService {
	svc := LaravelQueueService{
		Name:             filename,
		Label:            strings.TrimSuffix(filename, ".service"),
		WorkingDirectory: m.projectPath,
		// Defaults
		Executor:  "/usr/local/bin/fpcli",
		QueueName: "default",
		Sleep:     "3",
		Tries:     "3",
		Timeout:   "90",
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "User=") {
			svc.User = strings.TrimPrefix(line, "User=")
		}
		if strings.HasPrefix(line, "Group=") {
			svc.Group = strings.TrimPrefix(line, "Group=")
		}
		if strings.HasPrefix(line, "ExecStart=") {
			// ExecStart=/usr/local/bin/fpcli /var/www/path/artisan queue:work --queue=default ...
			parts := strings.Fields(strings.TrimPrefix(line, "ExecStart="))
			if len(parts) > 0 {
				svc.Executor = parts[0]
			}
			// Parse flags simply
			if strings.Contains(line, "--queue=") {
				svc.QueueName = extractFlag(line, "--queue=")
			}
			if strings.Contains(line, "--sleep=") {
				svc.Sleep = extractFlag(line, "--sleep=")
			}
			if strings.Contains(line, "--tries=") {
				svc.Tries = extractFlag(line, "--tries=")
			}
			if strings.Contains(line, "--timeout=") {
				svc.Timeout = extractFlag(line, "--timeout=")
			}
		}
	}

	// Check status
	svc.Status = "stopped"
	cmd := exec.Command("systemctl", "is-active", svc.Name)
	if out, err := cmd.Output(); err == nil {
		svc.Status = strings.TrimSpace(string(out))
	}

	// Check enabled
	cmd = exec.Command("systemctl", "is-enabled", svc.Name)
	if out, err := cmd.Output(); err == nil {
		if strings.TrimSpace(string(out)) == "enabled" {
			svc.Enabled = true
		}
	}

	return svc
}

func extractFlag(line, flag string) string {
	idx := strings.Index(line, flag)
	if idx == -1 {
		return ""
	}
	rest := line[idx+len(flag):]
	end := strings.Index(rest, " ")
	if end == -1 {
		return rest
	}
	return rest[:end]
}

func (m LaravelQueueModel) Init() tea.Cmd {
	return nil
}

func (m LaravelQueueModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch m.state {
		case QueueStateList:
			return m.updateList(msg)
		case QueueStateActions:
			return m.updateActions(msg)
		}
	}

	if m.state == QueueStateForm && m.form != nil {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}
		if m.form.State == huh.StateCompleted {
			return m.saveService()
		}
		return m, cmd
	}

	return m, nil
}

func (m LaravelQueueModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
		return m, tea.Quit
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
		// Back to previous screen (LaravelPermissionsModel will likely re-init)
		return NewLaravelPermissionsModel(), nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if m.cursor < len(m.services) { // +1 for "Add New"
			m.cursor++
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		if m.cursor == len(m.services) {
			// Add New
			m.startAdd()
			return m, m.form.Init()
		} else {
			// Select existing
			m.state = QueueStateActions
			m.actionCursor = 0
			m.actions = []string{
				"Start", "Stop", "Restart",
				"Enable", "Disable",
				"View Status", "View Logs",
				"Edit Configuration (Form)",
				"Edit Configuration (Editor)",
				"Delete Service",
				"← Back",
			}
			return m, nil
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
		m.loadServices()
	}
	return m, nil
}

func (m LaravelQueueModel) updateActions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q"))):
		m.state = QueueStateList
		return m, nil
	case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
		if m.actionCursor > 0 {
			m.actionCursor--
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
		if m.actionCursor < len(m.actions)-1 {
			m.actionCursor++
		}
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
		action := m.actions[m.actionCursor]
		svc := m.services[m.cursor]
		return m.executeAction(action, svc)
	}
	return m, nil
}

func (m *LaravelQueueModel) startAdd() {
	m.isEditing = false

	// Default label suggestion based on folder name
	folder := filepath.Base(m.projectPath)
	defaultLabel := fmt.Sprintf("laravel-queue-%s", folder)

	m.editService = LaravelQueueService{
		Label:            defaultLabel,
		User:             m.systemUser,
		Group:            m.systemUser,
		WorkingDirectory: m.projectPath,
		Executor:         "/usr/local/bin/fpcli",
		QueueName:        "default",
		Sleep:            "3",
		Tries:            "3",
		Timeout:          "90",
		ServiceCount:     "1",
	}
	if m.editService.User == "" {
		m.editService.User = "www-data"
		m.editService.Group = "www-data"
	}

	m.buildForm()
	m.state = QueueStateForm
}

func (m *LaravelQueueModel) startEdit(svc LaravelQueueService) {
	m.isEditing = true
	m.editService = svc
	m.buildForm()
	m.state = QueueStateForm
}

func (m *LaravelQueueModel) buildForm() {
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("label").
				Title("Service Label").
				Description("Unique name for the service").
				Value(&m.editService.Label).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("required")
					}
					return nil
				}),
			huh.NewInput().
				Key("user").
				Title("Run As User").
				Value(&m.editService.User),
			huh.NewInput().
				Key("executor").
				Title("Executor Path").
				Description("Path to fpcli or php").
				Value(&m.editService.Executor),
			huh.NewInput().
				Key("queue").
				Title("Queue Name").
				Value(&m.editService.QueueName),
			huh.NewInput().
				Key("sleep").
				Title("Sleep (seconds)").
				Value(&m.editService.Sleep),
			huh.NewInput().
				Key("tries").
				Title("Max Tries").
				Value(&m.editService.Tries),
			huh.NewInput().
				Key("timeout").
				Title("Timeout (seconds)").
				Value(&m.editService.Timeout),
			huh.NewInput().
				Key("count").
				Title("Service Count").
				Description("Number of concurrent workers").
				Value(&m.editService.ServiceCount).
				Validate(func(s string) error {
					var i int
					_, err := fmt.Sscanf(s, "%d", &i)
					if err != nil || i < 1 {
						return fmt.Errorf("must be number >= 1")
					}
					return nil
				}),
		),
	).WithTheme(m.theme.HuhTheme)
}

func (m LaravelQueueModel) saveService() (tea.Model, tea.Cmd) {
	// Need to check pointers/values references after form?
	// Huh binds to pointers passed, so m.editService fields should be updated?
	// Actually no, I used Value(&m.editService.Field).

	// Check if this works in Update loop:
	// The form updates the values in place because we passed pointers.

	svc := m.editService

	// Generate content
	// [Unit] ...

	serviceName := svc.Label
	// Ensure template naming
	if !strings.HasSuffix(serviceName, "@") {
		serviceName = strings.TrimSuffix(serviceName, ".service")
		if !strings.HasSuffix(serviceName, "@") {
			serviceName += "@"
		}
	}

	// Full filename
	serviceFileName := serviceName + ".service"

	// Ensure User matches Group if not specified?
	group := svc.User

	content := fmt.Sprintf(`[Unit]
Description=Laravel Queue Worker (%%i)
After=network.target

[Service]
User=%s
Group=%s
WorkingDirectory=%s

# ExecStart command
ExecStart=%s %s/artisan queue:work --queue=%s --sleep=%s --tries=%s --timeout=%s

Restart=always
RestartSec=5s
LimitNOFILE=65535

NoNewPrivileges=true
PrivateTmp=true

StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`, svc.User, group, m.projectPath, svc.Executor, m.projectPath, svc.QueueName, svc.Sleep, svc.Tries, svc.Timeout)

	path := filepath.Join("/etc/systemd/system", serviceFileName)

	count := svc.GetServiceCount()

	// Script to write file, enable N instances, start N instances
	// Logic:
	// 1. Stop/Disable ALL potential instances of this template to be safe?
	//    Or just up to new count?
	//    Safest is to stop all matching this pattern, then start new ones.
	//    But we only know the prefix.

	setupScript := fmt.Sprintf(`
# Write service file
cat <<EOF > %s
%s
EOF

systemctl daemon-reload

SERVICE_BASE="%s"
COUNT=%d

# Stop and disable all existing instances of this service pattern
# We rely on systemctl list-units to find them
for unit in $(systemctl list-units --full --all --no-legend "${SERVICE_BASE}*" | awk '{print $1}'); do
  echo "Stopping old instance: $unit"
  systemctl stop "$unit" || true
  systemctl disable "$unit" || true
done

# Enable and start new instances
for i in $(seq 1 $COUNT); do
  INSTANCE="${SERVICE_BASE}${i}"
  echo "Enabling and starting $INSTANCE"
  systemctl enable "$INSTANCE"
  systemctl start "$INSTANCE"
done
`, path, content, serviceName, count)

	m.state = QueueStateList

	return m, func() tea.Msg {
		return ExecutionStartMsg{
			Command:     fmt.Sprintf("sudo bash -c '%s'", setupScript),
			Description: fmt.Sprintf("Setting up %s (%d workers)", serviceName, count),
		}
	}
}

func (m LaravelQueueModel) executeAction(action string, svc LaravelQueueService) (tea.Model, tea.Cmd) {
	// svc.Name is likely "laravel-queue@.service" now?
	// Or we need the base name "laravel-queue@"

	serviceBase := svc.Label
	if !strings.HasSuffix(serviceBase, "@") {
		// If it's a legacy non-template service:
		if !strings.Contains(svc.Name, "@") {
			// Fallback to old behavior for old services
			return m.executeLegacyAction(action, svc)
		}
		serviceBase += "@"
	} else {
		// If label has @, ensure it ends with @
	}

	count := svc.GetServiceCount()

	switch action {
	case "Start", "Stop", "Restart", "Enable", "Disable":
		actLower := strings.ToLower(action)
		script := fmt.Sprintf(`
BASE="%s"
COUNT=%d
ACTION="%s"

for i in $(seq 1 $COUNT); do
  INSTANCE="${BASE}${i}"
  echo "${ACTION}ing $INSTANCE"
  systemctl $ACTION "$INSTANCE"
done
`, serviceBase, count, actLower)

		return m, func() tea.Msg {
			return ExecutionStartMsg{
				Command:     fmt.Sprintf("sudo bash -c '%s'", script),
				Description: fmt.Sprintf("%s %s (1..%d)", action, serviceBase, count),
			}
		}

	case "View Status":
		// Just show status of first one? Or all?
		// Systemctl status foo@1 foo@2 ...
		var instances []string
		for i := 1; i <= count; i++ {
			instances = append(instances, fmt.Sprintf("%s%d", serviceBase, i))
		}
		cmd := "systemctl status " + strings.Join(instances, " ") + " --no-pager -l"
		return m, func() tea.Msg { return ExecutionStartMsg{Command: cmd, Description: "Status " + serviceBase + "*"} }

	case "View Logs":
		// Logs for unit pattern
		// journalctl -u "name@*"
		cmd := fmt.Sprintf("journalctl -u '%s*' -f -n 20", serviceBase)
		return m, func() tea.Msg { return ExecutionStartMsg{Command: cmd, Description: "Logs " + serviceBase + "*"} }

	case "Delete Service":
		return m, func() tea.Msg {
			// Stop, disable all instances, remove template file
			script := fmt.Sprintf(`
BASE="%s"
FILE="/etc/systemd/system/%s.service"

for unit in $(systemctl list-units --full --all --no-legend "${BASE}*" | awk '{print $1}'); do
  systemctl stop "$unit" || true
  systemctl disable "$unit" || true
done
rm "$FILE" || true
systemctl daemon-reload
`, serviceBase, serviceBase) // Warning: this assumes Label == Filename base. Usually true.

			return ExecutionStartMsg{Command: "sudo bash -c '" + script + "'", Description: "Deleting " + serviceBase}
		}

	case "Edit Configuration (Form)":
		m.startEdit(svc)
		return m, m.form.Init()

	case "Edit Configuration (Editor)":
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: EditorSelectionScreen,
				Data: map[string]interface{}{
					"file":        filepath.Join("/etc/systemd/system", svc.Name),
					"description": "Edit " + svc.Name,
				},
			}
		}
	case "← Back":
		m.state = QueueStateList
		return m, nil
	}
	return m, nil
}

func (m LaravelQueueModel) executeLegacyAction(action string, svc LaravelQueueService) (tea.Model, tea.Cmd) {
	svcName := svc.Name
	switch action {
	case "Start":
		return m, func() tea.Msg {
			return ExecutionStartMsg{Command: "sudo systemctl start " + svcName, Description: "Starting " + svcName}
		}
	case "Stop":
		return m, func() tea.Msg {
			return ExecutionStartMsg{Command: "sudo systemctl stop " + svcName, Description: "Stopping " + svcName}
		}
	case "Restart":
		return m, func() tea.Msg {
			return ExecutionStartMsg{Command: "sudo systemctl restart " + svcName, Description: "Restarting " + svcName}
		}
	case "Enable":
		return m, func() tea.Msg {
			return ExecutionStartMsg{Command: "sudo systemctl enable " + svcName, Description: "Enabling " + svcName}
		}
	case "Disable":
		return m, func() tea.Msg {
			return ExecutionStartMsg{Command: "sudo systemctl disable " + svcName, Description: "Disabling " + svcName}
		}
	case "View Status":
		return m, func() tea.Msg {
			return ExecutionStartMsg{Command: "systemctl status " + svcName + " --no-pager -l", Description: "Status " + svcName}
		}
	case "View Logs":
		return m, func() tea.Msg {
			return ExecutionStartMsg{Command: "journalctl -u " + svcName + " -f -n 50", Description: "Logs " + svcName}
		}
	case "Delete Service":
		return m, func() tea.Msg {
			cmd := fmt.Sprintf("systemctl stop %s && systemctl disable %s && rm /etc/systemd/system/%s && systemctl daemon-reload", svcName, svcName, svcName)
			return ExecutionStartMsg{Command: "sudo bash -c '" + cmd + "'", Description: "Deleting " + svcName}
		}
	default:
		return m, nil
	}
}

func (m LaravelQueueModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	switch m.state {
	case QueueStateList:
		return m.viewList()
	case QueueStateForm:
		return m.viewForm()
	case QueueStateActions:
		return m.viewActions()
	}
	return ""
}

func (m LaravelQueueModel) viewList() string {
	header := m.theme.Title.Render("Laravel Queue Services")
	pathInfo := m.theme.DescriptionStyle.Render("Project: " + m.projectPath)

	var items []string

	// List existing services
	for i, svc := range m.services {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		status := " "
		if svc.Status == "active" {
			status = m.theme.SuccessStyle.Render("●")
		} else if svc.Status == "failed" {
			status = m.theme.ErrorStyle.Render("✗")
		} else {
			status = m.theme.WarningStyle.Render("○")
		}

		txt := fmt.Sprintf("%s%s %s (%s)", cursor, status, svc.Name, svc.QueueName)
		if i == m.cursor {
			txt = m.theme.SelectedItem.Render(txt)
		}
		items = append(items, txt)
	}

	// Add New option
	cursor := "  "
	if m.cursor == len(m.services) {
		cursor = m.theme.KeyStyle.Render("▶ ")
	}
	addTxt := fmt.Sprintf("%s+ Setup New Queue Service", cursor)
	if m.cursor == len(m.services) {
		addTxt = m.theme.SelectedItem.Render(addTxt)
	}
	items = append(items, "", addTxt)

	list := lipgloss.JoinVertical(lipgloss.Left, items...)
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Select • r: Refresh • Esc: Back")

	content := lipgloss.JoinVertical(lipgloss.Left, header, pathInfo, "", list, "", help)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.theme.RenderBox(content))
}

func (m LaravelQueueModel) viewForm() string {
	header := m.theme.Title.Render("Configure Queue Service")
	form := m.form.View()
	help := m.theme.Help.Render("Enter: Save • Esc: Cancel")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.theme.RenderBox(lipgloss.JoinVertical(lipgloss.Left, header, "", form, "", help)))
}

func (m LaravelQueueModel) viewActions() string {
	svc := m.services[m.cursor]
	header := m.theme.Title.Render(svc.Name)

	var items []string
	for i, act := range m.actions {
		cursor := "  "
		if i == m.actionCursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}
		txt := fmt.Sprintf("%s%s", cursor, act)
		if i == m.actionCursor {
			txt = m.theme.SelectedItem.Render(txt)
		}
		items = append(items, txt)
	}

	list := lipgloss.JoinVertical(lipgloss.Left, items...)
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Execute • Esc: Back")

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.theme.RenderBox(lipgloss.JoinVertical(lipgloss.Left, header, "", list, "", help)))
}
