package screens

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// SupervisorAddProgramModel represents the add program flow
type SupervisorAddProgramModel struct {
	theme       *theme.Theme
	width       int
	height      int
	manager     *system.SupervisorManager
	step        int // 0=form, 1=editing, 2=result
	programName string
	editor      string
	form        *huh.Form
	err         error
	message     string
}

// NewSupervisorAddProgramModel creates a new add program model
func NewSupervisorAddProgramModel(manager *system.SupervisorManager) SupervisorAddProgramModel {
	t := theme.DefaultTheme()

	m := SupervisorAddProgramModel{
		theme:       t,
		manager:     manager,
		step:        0,
		programName: "",
		editor:      "nano",
	}

	m.form = m.buildForm()
	return m
}

func (m *SupervisorAddProgramModel) buildForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Program Name").
				Description("Unique identifier for the supervisor program").
				Placeholder("myprogram").
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("program name cannot be empty")
					}
					if strings.Contains(s, " ") {
						return fmt.Errorf("program name cannot contain spaces")
					}
					return nil
				}).
				Value(&m.programName),

			huh.NewSelect[string]().
				Title("Editor").
				Description("Choose editor to configure the program").
				Options(
					huh.NewOption("Nano (recommended for beginners)", "nano"),
					huh.NewOption("Vi/Vim", "vi"),
				).
				Value(&m.editor),
		),
	).WithTheme(m.theme.HuhTheme).
		WithShowHelp(true).
		WithShowErrors(true)
}

func (m SupervisorAddProgramModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m SupervisorAddProgramModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case ExecutionCompleteMsg:
		// Handle result from editor
		m.step = 2
		if msg.Success {
			m.message = msg.Output
			m.err = nil
		} else {
			m.err = msg.Error
		}
		return m, nil

	case tea.KeyMsg:
		// Handle result state
		if m.step == 2 {
			switch msg.String() {
			case "enter", " ", "esc":
				return m, func() tea.Msg {
					return NavigateMsg{Screen: SupervisorManagementScreen}
				}
			}
			return m, nil
		}

		// Handle editing state
		if m.step == 1 {
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.step == 0 && m.form.State == huh.StateNormal {
				return m, func() tea.Msg {
					return NavigateMsg{Screen: SupervisorManagementScreen}
				}
			}
		}
	}

	// Handle form in step 0
	if m.step == 0 {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}

		// Check if form is completed
		if m.form.State == huh.StateCompleted {
			m.step = 1
			return m, m.openEditor()
		}

		return m, cmd
	}

	return m, nil
}

func (m SupervisorAddProgramModel) openEditor() tea.Cmd {
	return func() tea.Msg {
		// Create temp config file
		configPath := fmt.Sprintf("/etc/supervisor/conf.d/%s.conf", m.programName)
		
		// Create initial template
		template := fmt.Sprintf(`[program:%s]
command=/path/to/your/command
directory=/path/to/working/directory
user=www-data
autostart=true
autorestart=true
redirect_stderr=true
stdout_logfile=/var/log/supervisor/%s.log
stdout_logfile_maxbytes=10MB
`, m.programName, m.programName)

		// Write template
		if err := os.WriteFile(configPath, []byte(template), 0644); err != nil {
			return ExecutionCompleteMsg{
				Success: false,
				Error:   err,
			}
		}

		// Open editor
		cmd := exec.Command(m.editor, configPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Run editor
		if err := cmd.Run(); err != nil {
			return ExecutionCompleteMsg{
				Success: false,
				Error:   fmt.Errorf("editor failed: %w", err),
			}
		}

		// Validate config
		validateCmd := exec.Command("supervisorctl", "reread")
		output, err := validateCmd.CombinedOutput()
		
		if err != nil || strings.Contains(string(output), "ERROR") {
			// Config is invalid - remove it
			os.Remove(configPath)
			return ExecutionCompleteMsg{
				Success: false,
				Error:   fmt.Errorf("configuration validation failed: %s", string(output)),
			}
		}

		// Config is valid - update supervisor
		updateCmd := exec.Command("supervisorctl", "update")
		updateCmd.Run()

		return ExecutionCompleteMsg{
			Success: true,
			Output:  fmt.Sprintf("Program '%s' added successfully and configuration is valid", m.programName),
		}
	}
}

func (m SupervisorAddProgramModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content []string

	switch m.step {
	case 0: // Form input
		header := m.theme.Title.Render("Add Supervisor Program")
		content = append(content, header)
		content = append(content, "")

		if m.err != nil {
			content = append(content, m.theme.ErrorStyle.Render(m.theme.Symbols.CrossMark+" Error: "+m.err.Error()))
			content = append(content, "")
		}

		content = append(content, m.form.View())
		content = append(content, "")
		content = append(content, m.theme.Help.Render("Tab: Navigate "+m.theme.Symbols.Bullet+" Enter: Submit "+m.theme.Symbols.Bullet+" Esc: Cancel"))

	case 1: // Editing in progress
		header := m.theme.Title.Render("Add Supervisor Program - Editing")
		content = append(content, header)
		content = append(content, "")
		content = append(content, m.theme.InfoStyle.Render("Editor is open in the terminal..."))
		content = append(content, "")
		content = append(content, m.theme.DescriptionStyle.Render("Editing configuration file with "+m.editor))
		content = append(content, m.theme.DescriptionStyle.Render("Save and exit when done"))

	case 2: // Result
		header := m.theme.Title.Render("Add Supervisor Program - Result")
		content = append(content, header)
		content = append(content, "")

		if m.err != nil {
			content = append(content, m.theme.ErrorStyle.Render(m.theme.Symbols.CrossMark+" Configuration Error"))
			content = append(content, "")
			content = append(content, m.theme.ErrorStyle.Render(m.err.Error()))
			content = append(content, "")
			content = append(content, m.theme.WarningStyle.Render(m.theme.Symbols.Warning+" The configuration is invalid and was not applied."))
		} else {
			content = append(content, m.theme.SuccessStyle.Render(m.theme.Symbols.CheckMark+" "+m.message))
			content = append(content, "")
			content = append(content, m.theme.DescriptionStyle.Render("Configuration validated successfully"))
		}

		content = append(content, "")
		content = append(content, m.theme.Help.Render("Enter: Return to Menu"))
	}

	body := lipgloss.JoinVertical(lipgloss.Left, content...)
	bordered := m.theme.BorderStyle.Render(body)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}
