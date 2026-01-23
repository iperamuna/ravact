package screens

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// EmbeddedFS will be set by main package
var EmbeddedFS embed.FS

// ExecutionState represents the state of execution
type ExecutionState int

const (
	ExecutionRunning ExecutionState = iota
	ExecutionSuccess
	ExecutionFailed
	ExecutionCancelled
)

// ExecutionModel represents the execution screen
type ExecutionModel struct {
	theme       *theme.Theme
	width       int
	height      int
	command     string
	description string
	state       ExecutionState
	output      []string
	exitCode    int
	startTime   time.Time
	endTime     time.Time
	maxLines    int
	scrollOffset int
	autoScroll   bool
	returnScreen ScreenType
}

// ExecutionOutputMsg is sent when new output is received
type ExecutionOutputMsg struct {
	Line string
}

// NewExecutionModel creates a new execution model
func NewExecutionModel(command, description string, returnScreen ScreenType) ExecutionModel {
	return ExecutionModel{
		theme:        theme.DefaultTheme(),
		command:      command,
		description:  description,
		state:        ExecutionRunning,
		output:       []string{},
		maxLines:     1000, // Keep last 1000 lines
		autoScroll:   true,
		returnScreen: returnScreen,
	}
}

// Init initializes the execution screen
func (m ExecutionModel) Init() tea.Cmd {
	m.startTime = time.Now()
	return m.executeCommand
}

// executeCommand runs the command and streams output
func (m ExecutionModel) executeCommand() tea.Msg {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Check if this is a script path (embedded)
	var cmd *exec.Cmd
	if strings.HasSuffix(m.command, ".sh") && strings.Contains(m.command, "assets/scripts/") {
		// Check OS compatibility for setup scripts
		if runtime.GOOS != "linux" {
			errorMsg := fmt.Sprintf("⚠ Setup scripts are designed for Linux only.\n\nCurrent OS: %s\n\n", runtime.GOOS)
			errorMsg += "These scripts use Linux-specific commands:\n"
			errorMsg += "  • apt-get / yum (package managers)\n"
			errorMsg += "  • systemctl (service management)\n"
			errorMsg += "  • Linux file paths and configurations\n\n"
			errorMsg += "To use Ravact setup features:\n"
			errorMsg += "  1. Deploy to a Linux server (Ubuntu/Debian/RHEL/CentOS)\n"
			errorMsg += "  2. Use Docker: make docker-test\n"
			errorMsg += "  3. Use a Linux VM (Multipass, UTM, VirtualBox)\n\n"
			errorMsg += "See docs/MACOS_LIMITATIONS.md for details."
			
			return ExecutionCompleteMsg{
				Success: false,
				Output:  errorMsg,
				Error:   fmt.Errorf("setup scripts require Linux (current OS: %s)", runtime.GOOS),
			}
		}
		
		// Execute embedded script by reading content and piping to bash
		scriptContent, err := EmbeddedFS.ReadFile(m.command)
		if err != nil {
			return ExecutionCompleteMsg{
				Success: false,
				Output:  fmt.Sprintf("Failed to read embedded script: %v", err),
				Error:   err,
			}
		}

		// Run bash with script piped to stdin
		cmd = exec.CommandContext(ctx, "bash", "-s")
		cmd.Stdin = bytes.NewReader(scriptContent)
	} else {
		// Regular command execution
		if m.command == "" {
			return ExecutionCompleteMsg{
				Success: false,
				Output:  "No command specified",
				Error:   fmt.Errorf("empty command"),
			}
		}
		cmd = exec.CommandContext(ctx, "bash", "-c", m.command)
	}
	
	// Get stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ExecutionCompleteMsg{
			Success: false,
			Output:  fmt.Sprintf("Failed to create stdout pipe: %v", err),
			Error:   err,
		}
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return ExecutionCompleteMsg{
			Success: false,
			Output:  fmt.Sprintf("Failed to create stderr pipe: %v", err),
			Error:   err,
		}
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return ExecutionCompleteMsg{
			Success: false,
			Output:  fmt.Sprintf("Failed to start command: %v", err),
			Error:   err,
		}
	}

	// Stream output (this is a simplified version - in real TUI we'd use channels)
	outputLines := []string{}
	
	// Read stdout
	stdoutScanner := bufio.NewScanner(stdout)
	go func() {
		for stdoutScanner.Scan() {
			outputLines = append(outputLines, stdoutScanner.Text())
		}
	}()

	// Read stderr
	stderrScanner := bufio.NewScanner(stderr)
	go func() {
		for stderrScanner.Scan() {
			outputLines = append(outputLines, "[ERROR] "+stderrScanner.Text())
		}
	}()

	// Wait for command to complete
	err = cmd.Wait()
	
	// Build final output
	output := strings.Join(outputLines, "\n")
	if output == "" {
		output = "Command completed with no output"
	}

	success := err == nil
	if err != nil {
		output += fmt.Sprintf("\n\nCommand failed with error: %v", err)
	}

	return ExecutionCompleteMsg{
		Success: success,
		Output:  output,
		Error:   err,
	}
}

// Update handles messages for execution
func (m ExecutionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case ExecutionCompleteMsg:
		m.endTime = time.Now()
		if msg.Success {
			m.state = ExecutionSuccess
		} else {
			m.state = ExecutionFailed
		}
		
		// Add output lines
		lines := strings.Split(msg.Output, "\n")
		for _, line := range lines {
			m.output = append(m.output, line)
		}
		
		// Trim to max lines
		if len(m.output) > m.maxLines {
			m.output = m.output[len(m.output)-m.maxLines:]
		}
		
		if msg.Error != nil {
			m.exitCode = 1
		} else {
			m.exitCode = 0
		}
		
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state == ExecutionRunning {
				m.state = ExecutionCancelled
				return m, tea.Quit
			}
			return m, tea.Quit

		case "esc", "enter", " ":
			// Only allow exit if execution is complete
			if m.state != ExecutionRunning {
				return m, func() tea.Msg {
					return NavigateMsg{Screen: m.returnScreen}
				}
			}

		case "up", "k":
			if m.scrollOffset > 0 {
				m.scrollOffset--
				m.autoScroll = false
			}

		case "down", "j":
			maxScroll := len(m.output) - (m.height - 10)
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.scrollOffset < maxScroll {
				m.scrollOffset++
			}

		case "home":
			m.scrollOffset = 0
			m.autoScroll = false

		case "end":
			m.autoScroll = true
			m.scrollOffset = len(m.output) - (m.height - 10)
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}
		}
	}

	return m, nil
}

// View renders the execution screen
func (m ExecutionModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	var header string
	switch m.state {
	case ExecutionRunning:
		header = m.theme.Title.Render("⏳ Executing...")
	case ExecutionSuccess:
		header = m.theme.SuccessStyle.Render("✓ Execution Completed Successfully")
	case ExecutionFailed:
		header = m.theme.ErrorStyle.Render("✗ Execution Failed")
	case ExecutionCancelled:
		header = m.theme.WarningStyle.Render("⚠ Execution Cancelled")
	}

	// Description
	desc := m.theme.DescriptionStyle.Render(m.description)

	// Command
	cmdDisplay := m.theme.Label.Render("Command: ") + m.theme.MenuItem.Render(m.command)

	// Duration
	var duration string
	if m.state == ExecutionRunning {
		duration = fmt.Sprintf("Running for: %v", time.Since(m.startTime).Round(time.Second))
	} else {
		duration = fmt.Sprintf("Duration: %v", m.endTime.Sub(m.startTime).Round(time.Second))
	}
	durationDisplay := m.theme.InfoStyle.Render(duration)

	// Output window
	outputHeight := m.height - 12 // Reserve space for header, footer, etc.
	if outputHeight < 5 {
		outputHeight = 5
	}

	var outputLines []string
	if len(m.output) == 0 {
		if m.state == ExecutionRunning {
			outputLines = []string{
				m.theme.DescriptionStyle.Render("Waiting for output..."),
			}
		} else {
			outputLines = []string{
				m.theme.DescriptionStyle.Render("No output produced"),
			}
		}
	} else {
		// Calculate visible range
		start := m.scrollOffset
		end := m.scrollOffset + outputHeight
		if end > len(m.output) {
			end = len(m.output)
		}
		if start < 0 {
			start = 0
		}

		// Show scroll indicators
		if start > 0 {
			outputLines = append(outputLines, m.theme.DescriptionStyle.Render("  ↑ More output above..."))
		}

		// Show visible lines
		for i := start; i < end && i < len(m.output); i++ {
			line := m.output[i]
			// Color error lines
			if strings.Contains(line, "[ERROR]") || strings.Contains(line, "error:") || strings.Contains(line, "Error:") {
				outputLines = append(outputLines, m.theme.ErrorStyle.Render(line))
			} else if strings.Contains(line, "warning:") || strings.Contains(line, "Warning:") {
				outputLines = append(outputLines, m.theme.WarningStyle.Render(line))
			} else if strings.Contains(line, "✓") || strings.Contains(line, "success") {
				outputLines = append(outputLines, m.theme.SuccessStyle.Render(line))
			} else {
				outputLines = append(outputLines, line)
			}
		}

		if end < len(m.output) {
			outputLines = append(outputLines, m.theme.DescriptionStyle.Render("  ↓ More output below..."))
		}
	}

	output := lipgloss.JoinVertical(lipgloss.Left, outputLines...)
	outputBox := m.theme.BorderStyle.Render(output)

	// Progress indicator
	var progress string
	if m.state == ExecutionRunning {
		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		idx := int(time.Since(m.startTime).Milliseconds()/100) % len(spinner)
		progress = m.theme.InfoStyle.Render(fmt.Sprintf("%s Running...", spinner[idx]))
	}

	// Help text
	var help string
	if m.state == ExecutionRunning {
		help = m.theme.Help.Render("↑/↓: Scroll • Ctrl+C: Cancel • Please wait...")
	} else {
		help = m.theme.Help.Render("↑/↓: Scroll • Enter/Esc: Continue • q: Quit")
	}

	// Exit code
	var exitCodeDisplay string
	if m.state != ExecutionRunning {
		if m.exitCode == 0 {
			exitCodeDisplay = m.theme.SuccessStyle.Render(fmt.Sprintf("Exit Code: %d", m.exitCode))
		} else {
			exitCodeDisplay = m.theme.ErrorStyle.Render(fmt.Sprintf("Exit Code: %d", m.exitCode))
		}
	}

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		desc,
		cmdDisplay,
		durationDisplay,
		"",
		outputBox,
		"",
		progress,
		exitCodeDisplay,
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
