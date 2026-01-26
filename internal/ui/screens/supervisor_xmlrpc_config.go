package screens

import (
	"fmt"
	"net"
	"strconv"

	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// SupervisorXMLRPCConfigModel represents the XML-RPC configuration screen
type SupervisorXMLRPCConfigModel struct {
	theme    *theme.Theme
	width    int
	height   int
	manager  *system.SupervisorManager
	form     *huh.Form
	ip       string
	port     string
	username string
	password string
	err      error
	success  bool
}

// NewSupervisorXMLRPCConfigModel creates a new XML-RPC config model
func NewSupervisorXMLRPCConfigModel(manager *system.SupervisorManager) SupervisorXMLRPCConfigModel {
	t := theme.DefaultTheme()

	// Load current config
	config, _ := manager.GetXMLRPCConfig()

	m := SupervisorXMLRPCConfigModel{
		theme:    t,
		manager:  manager,
		ip:       "127.0.0.1",
		port:     "9001",
		username: "",
		password: "",
	}

	// Set values from existing config if available
	if config != nil {
		if config.IP != "" {
			m.ip = config.IP
		}
		if config.Port != "" {
			m.port = config.Port
		}
		m.username = config.Username
		// Note: password is not loaded for security reasons
	}

	m.form = m.buildForm()
	return m
}

func (m *SupervisorXMLRPCConfigModel) buildForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("ip").
				Title("IP Address").
				Description("IP address to bind the XML-RPC server (use 0.0.0.0 for all interfaces)").
				Placeholder("127.0.0.1").
				Validate(func(s string) error {
					if s == "" {
						return nil // Will use default
					}
					if s != "0.0.0.0" && s != "127.0.0.1" && s != "localhost" {
						if ip := net.ParseIP(s); ip == nil {
							return fmt.Errorf("invalid IP address")
						}
					}
					return nil
				}).
				Value(&m.ip),

			huh.NewInput().
				Key("port").
				Title("Port").
				Description("Port number for XML-RPC server (default: 9001)").
				Placeholder("9001").
				Validate(func(s string) error {
					if s == "" {
						return nil // Will use default
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
				Value(&m.port),

			huh.NewInput().
				Key("username").
				Title("Username").
				Description("Username for XML-RPC authentication (optional)").
				Placeholder("admin").
				Value(&m.username),

			huh.NewInput().
				Key("password").
				Title("Password").
				Description("Password for XML-RPC authentication (optional)").
				Placeholder("Enter password...").
				EchoMode(huh.EchoModePassword).
				Value(&m.password),
		),
	).WithTheme(m.theme.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)
}

func (m SupervisorXMLRPCConfigModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m SupervisorXMLRPCConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// If showing success/error, any key returns
		if m.success || m.err != nil {
			if msg.String() == "enter" || msg.String() == " " || msg.String() == "esc" {
				return m, func() tea.Msg {
					return NavigateMsg{Screen: SupervisorManagementScreen}
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.form.State == huh.StateNormal {
				return m, func() tea.Msg {
					return NavigateMsg{Screen: SupervisorManagementScreen}
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
		return m.saveConfig()
	}

	return m, cmd
}

func (m SupervisorXMLRPCConfigModel) saveConfig() (SupervisorXMLRPCConfigModel, tea.Cmd) {
	// Apply defaults if empty
	ip := m.ip
	port := m.port

	if ip == "" {
		ip = "127.0.0.1"
	}
	if port == "" {
		port = "9001"
	}

	err := m.manager.SetXMLRPCConfig(ip, port, m.username, m.password)
	if err != nil {
		m.err = err
		m.form = m.buildForm()
		return m, nil
	}

	m.success = true
	m.err = nil
	return m, nil
}

func (m SupervisorXMLRPCConfigModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// If success, show message
	if m.success {
		msg := m.theme.SuccessStyle.Render(m.theme.Symbols.CheckMark + " XML-RPC configured successfully!")
		note := m.theme.DescriptionStyle.Render("Supervisor will be restarted to apply changes.")
		help := m.theme.Help.Render("Press any key to continue...")
		content := lipgloss.JoinVertical(lipgloss.Center, "", msg, "", note, "", help)
		bordered := m.theme.BorderStyle.Render(content)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
	}

	// If error, show message
	if m.err != nil {
		msg := m.theme.ErrorStyle.Render(m.theme.Symbols.CrossMark + " Error: " + m.err.Error())
		help := m.theme.Help.Render("Press any key to continue...")
		content := lipgloss.JoinVertical(lipgloss.Center, "", msg, "", help)
		bordered := m.theme.BorderStyle.Render(content)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, bordered)
	}

	// Header
	header := m.theme.Title.Render(m.theme.Symbols.Box + " Configure XML-RPC Server")
	subtitle := m.theme.Subtitle.Render("Configure Supervisor XML-RPC interface for remote management")

	// Note about restart
	note := m.theme.WarningStyle.Render(m.theme.Symbols.Warning + " Note: Supervisor will be restarted after saving")

	// Help
	help := m.theme.Help.Render("Tab/Shift+Tab: Navigate " + m.theme.Symbols.Bullet + " Enter: Submit " + m.theme.Symbols.Bullet + " Esc: Cancel")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		subtitle,
		"",
		m.form.View(),
		"",
		note,
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
