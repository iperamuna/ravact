package screens

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
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
	step        int // 0=get name, 1=choose editor, 2=editing, 3=validating
	programName string
	editor      string
	textInput   textinput.Model
	editorIndex int
	editors     []string
	err         error
	message     string
}

// NewSupervisorAddProgramModel creates a new add program model
func NewSupervisorAddProgramModel(manager *system.SupervisorManager) SupervisorAddProgramModel {
	ti := textinput.New()
	ti.Placeholder = "Enter program name"
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 40

	return SupervisorAddProgramModel{
		theme:       theme.DefaultTheme(),
		manager:     manager,
		step:        0,
		textInput:   ti,
		editors:     []string{"nano", "vi"},
		editorIndex: 0,
	}
}

func (m SupervisorAddProgramModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SupervisorAddProgramModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.step == 2 {
				// Don't quit while editing
				return m, nil
			}
			return m, tea.Quit
		case "esc":
			if m.step == 2 {
				// Don't allow escape while editing
				return m, nil
			}
			return m, func() tea.Msg {
				return NavigateMsg{Screen: SupervisorManagementScreen}
			}
		case "enter":
			return m.handleEnter()
		case "up", "k":
			if m.step == 1 && m.editorIndex > 0 {
				m.editorIndex--
			}
		case "down", "j":
			if m.step == 1 && m.editorIndex < len(m.editors)-1 {
				m.editorIndex++
			}
		}
	}

	if m.step == 0 {
		m.textInput, cmd = m.textInput.Update(msg)
	}
	return m, cmd
}

func (m SupervisorAddProgramModel) handleEnter() (SupervisorAddProgramModel, tea.Cmd) {
	switch m.step {
	case 0: // Get program name
		name := m.textInput.Value()
		if name == "" {
			m.err = fmt.Errorf("program name cannot be empty")
			return m, nil
		}
		m.programName = name
		m.step = 1
		m.err = nil
		return m, nil

	case 1: // Choose editor
		m.editor = m.editors[m.editorIndex]
		return m, m.openEditor()

	case 3: // After validation, return to management
		return m, func() tea.Msg {
			return NavigateMsg{
				Screen: SupervisorManagementScreen,
				Data: map[string]interface{}{
					"success": m.message,
				},
			}
		}
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
	case 0: // Get program name
		header := m.theme.Title.Render("➕ Add Supervisor Program - Step 1")
		content = append(content, header)
		content = append(content, "")
		
		if m.err != nil {
			content = append(content, m.theme.ErrorStyle.Render("Error: "+m.err.Error()))
			content = append(content, "")
		}
		
		content = append(content, m.theme.Label.Render("Program Name:"))
		content = append(content, m.textInput.View())
		content = append(content, "")
		content = append(content, m.theme.Help.Render("Enter: Next • Esc: Cancel • q: Quit"))

	case 1: // Choose editor
		header := m.theme.Title.Render("➕ Add Supervisor Program - Step 2")
		content = append(content, header)
		content = append(content, "")
		content = append(content, m.theme.Label.Render(fmt.Sprintf("Program: %s", m.programName)))
		content = append(content, "")
		content = append(content, m.theme.Label.Render("Choose your editor:"))
		content = append(content, "")
		
		for i, editor := range m.editors {
			cursor := "  "
			if i == m.editorIndex {
				cursor = m.theme.KeyStyle.Render("▶ ")
			}
			
			var line string
			if i == m.editorIndex {
				line = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, editor))
			} else {
				line = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, editor))
			}
			content = append(content, line)
		}
		
		content = append(content, "")
		content = append(content, m.theme.Help.Render("↑/↓: Select • Enter: Open Editor • Esc: Cancel"))

	case 2: // Editing in progress
		header := m.theme.Title.Render("➕ Add Supervisor Program - Editing")
		content = append(content, header)
		content = append(content, "")
		content = append(content, m.theme.Label.Render("Editor is open in the terminal..."))
		content = append(content, "")
		content = append(content, m.theme.DescriptionStyle.Render("Editing configuration file with "+m.editor))
		content = append(content, m.theme.DescriptionStyle.Render("Save and exit when done"))

	case 3: // Validation result
		header := m.theme.Title.Render("➕ Add Supervisor Program - Result")
		content = append(content, header)
		content = append(content, "")
		
		if m.err != nil {
			content = append(content, m.theme.ErrorStyle.Render("❌ Configuration Error"))
			content = append(content, "")
			content = append(content, m.theme.ErrorStyle.Render(m.err.Error()))
			content = append(content, "")
			content = append(content, m.theme.WarningStyle.Render("The configuration is invalid and was not applied."))
			content = append(content, m.theme.DescriptionStyle.Render("Program was not added to Supervisor."))
		} else {
			content = append(content, m.theme.SuccessStyle.Render("✓ "+m.message))
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
