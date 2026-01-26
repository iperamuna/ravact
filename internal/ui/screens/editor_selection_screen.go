package screens

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/system"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// EditorSelectionModel represents the editor selection screen
type EditorSelectionModel struct {
	theme       *theme.Theme
	width       int
	height      int
	site        system.NginxSite
	cursor      int
	editors     []string
	filePath    string
	description string
	returnScreen ScreenType
}

// NewEditorSelectionModel creates a new editor selection model for nginx sites
func NewEditorSelectionModel(site system.NginxSite) EditorSelectionModel {
	editors := []string{
		"nano - User-friendly editor (recommended)",
		"vi - Classic Unix editor (advanced)",
		"← Cancel",
	}
	
	return EditorSelectionModel{
		theme:        theme.DefaultTheme(),
		site:         site,
		cursor:       0,
		editors:      editors,
		filePath:     site.ConfigPath,
		description:  site.Name,
		returnScreen: ConfigEditorScreen,
	}
}

// NewEditorSelectionModelForFile creates a new editor selection model for any file
func NewEditorSelectionModelForFile(filePath, description string, returnScreen ScreenType) EditorSelectionModel {
	editors := []string{
		"nano - User-friendly editor (recommended)",
		"vi - Classic Unix editor (advanced)",
		"← Cancel",
	}
	
	return EditorSelectionModel{
		theme:        theme.DefaultTheme(),
		cursor:       0,
		editors:      editors,
		filePath:     filePath,
		description:  description,
		returnScreen: returnScreen,
	}
}

// Init initializes the editor selection screen
func (m EditorSelectionModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m EditorSelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			return m, func() tea.Msg {
				return BackMsg{}
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.editors)-1 {
				m.cursor++
			}

		case "enter", " ":
			return m.executeSelection()
		}
	}

	return m, nil
}

// executeSelection executes the selected editor
func (m EditorSelectionModel) executeSelection() (EditorSelectionModel, tea.Cmd) {
	switch m.cursor {
	case 0: // nano
		return m, tea.ExecProcess(exec.Command("nano", m.filePath), func(err error) tea.Msg {
			if err != nil {
				return EditorCompleteMsg{
					Error: fmt.Sprintf("Failed to run nano: %v", err),
				}
			}
			return EditorCompleteMsg{
				Success: "Config file edited with nano",
			}
		})

	case 1: // vi
		return m, tea.ExecProcess(exec.Command("vi", m.filePath), func(err error) tea.Msg {
			if err != nil {
				return EditorCompleteMsg{
					Error: fmt.Sprintf("Failed to run vi: %v", err),
				}
			}
			return EditorCompleteMsg{
				Success: "Config file edited with vi",
			}
		})

	case 2: // Cancel
		return m, func() tea.Msg {
			return BackMsg{}
		}
	}

	return m, nil
}

// EditorCompleteMsg is sent when editor finishes
type EditorCompleteMsg struct {
	Success string
	Error   string
}

// View renders the editor selection screen
func (m EditorSelectionModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Header
	header := m.theme.Title.Render("Choose Editor")

	// File info
	siteInfo := m.theme.DescriptionStyle.Render(fmt.Sprintf("Editing: %s", m.description))
	filePath := m.theme.DescriptionStyle.Render(fmt.Sprintf("File: %s", m.filePath))

	// Instructions
	instructions := lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		m.theme.Label.Render("Select your preferred text editor:"),
		"",
		m.theme.DescriptionStyle.Render("nano - Easy to use, shows keyboard shortcuts at bottom"),
		m.theme.DescriptionStyle.Render("       Press Ctrl+O to save, Ctrl+X to exit"),
		"",
		m.theme.DescriptionStyle.Render("vi   - Powerful but requires learning commands"),
		m.theme.DescriptionStyle.Render("       Press 'i' to insert, ESC then ':wq' to save & exit"),
		"",
	)

	// Editor menu
	var editorItems []string
	for i, editor := range m.editors {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.KeyStyle.Render("▶ ")
		}

		var renderedItem string
		if i == m.cursor {
			renderedItem = m.theme.SelectedItem.Render(fmt.Sprintf("%s%s", cursor, editor))
		} else {
			renderedItem = m.theme.MenuItem.Render(fmt.Sprintf("%s%s", cursor, editor))
		}

		editorItems = append(editorItems, renderedItem)
	}

	editorsMenu := lipgloss.JoinVertical(lipgloss.Left, editorItems...)

	// Help
	help := m.theme.Help.Render("↑/↓: Navigate • Enter: Open Editor • Esc: Back • q: Quit")

	// Combine all sections
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		siteInfo,
		filePath,
		instructions,
		editorsMenu,
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
