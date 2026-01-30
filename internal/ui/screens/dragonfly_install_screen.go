package screens

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// DragonflyInstallOption represents an installation option
type DragonflyInstallOption struct {
	ID          string
	Name        string
	Description string
	EnvValue    string
}

// DragonflyInstallModel represents the Dragonfly installation options screen
type DragonflyInstallModel struct {
	theme   *theme.Theme
	width   int
	height  int
	cursor  int
	options []DragonflyInstallOption
}

// NewDragonflyInstallModel creates a new Dragonfly installation options model
func NewDragonflyInstallModel() DragonflyInstallModel {
	options := []DragonflyInstallOption{
		{
			ID:          "binary",
			Name:        "Binary Installation (Recommended)",
			Description: "Download and install Dragonfly binary directly. Best for production use.",
			EnvValue:    "1",
		},
		{
			ID:          "docker",
			Name:        "Docker Installation",
			Description: "Run Dragonfly in a Docker container. Requires Docker to be installed.",
			EnvValue:    "2",
		},
	}

	return DragonflyInstallModel{
		theme:   theme.DefaultTheme(),
		cursor:  0,
		options: options,
	}
}

// Init initializes the Dragonfly install screen
func (m DragonflyInstallModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the Dragonfly install screen
func (m DragonflyInstallModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc", "backspace":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: SetupMenuScreen}
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}

		case "enter", " ":
			if len(m.options) > 0 {
				selectedOption := m.options[m.cursor]
				// Set environment variable and run the script
				command := fmt.Sprintf("DRAGONFLY_INSTALL_METHOD=%s assets/scripts/dragonfly.sh", selectedOption.EnvValue)
				return m, func() tea.Msg {
					return ExecutionStartMsg{
						Command:     command,
						Description: fmt.Sprintf("Installing Dragonfly (%s)", selectedOption.Name),
					}
				}
			}
		}
	}

	return m, nil
}

// View renders the Dragonfly install screen
func (m DragonflyInstallModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	header := m.theme.Title.Render("Dragonfly Installation")

	// Description
	desc := m.theme.DescriptionStyle.Render("Dragonfly is a modern Redis/Memcached replacement with better performance and lower memory usage.")

	// Subtitle
	subtitle := m.theme.Subtitle.Render("Select Installation Method:")

	// Option items
	var optionItems []string
	for i, option := range m.options {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		var renderedItem string
		if i == m.cursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, option.Name))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, option.Name))
		}

		optionDesc := m.theme.DescriptionStyle.Render("  " + option.Description)

		optionItems = append(optionItems, renderedItem)
		optionItems = append(optionItems, optionDesc)
		optionItems = append(optionItems, "")
	}

	menu := lipgloss.JoinVertical(lipgloss.Left, optionItems...)

	// Info box
	infoLines := []string{
		m.theme.Label.Render("Binary Installation:"),
		m.theme.DescriptionStyle.Render("  • Downloads from GitHub releases"),
		m.theme.DescriptionStyle.Render("  • Creates systemd service"),
		m.theme.DescriptionStyle.Render("  • Configures automatic startup"),
		"",
		m.theme.Label.Render("Docker Installation:"),
		m.theme.DescriptionStyle.Render("  • Pulls official Docker image"),
		m.theme.DescriptionStyle.Render("  • Creates persistent container"),
		m.theme.DescriptionStyle.Render("  • Requires Docker to be installed"),
	}
	infoSection := lipgloss.JoinVertical(lipgloss.Left, infoLines...)

	// Help
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Install • Esc: Back • q: Quit")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		desc,
		"",
		"",
		subtitle,
		"",
		menu,
		"",
		infoSection,
		"",
		"",
		help,
	)

	// Add border and center
	bordered := m.theme.RenderBox(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}
