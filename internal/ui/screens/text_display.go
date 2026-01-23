package screens

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// TextDisplayModel represents a text display screen
type TextDisplayModel struct {
	theme      *theme.Theme
	width      int
	height     int
	title      string
	content    string
	returnScreen ScreenType
}

// NewTextDisplayModel creates a new text display model
func NewTextDisplayModel(title, content string, returnScreen ScreenType) TextDisplayModel {
	return TextDisplayModel{
		theme:        theme.DefaultTheme(),
		title:        title,
		content:      content,
		returnScreen: returnScreen,
	}
}

func (m TextDisplayModel) Init() tea.Cmd {
	return nil
}

func (m TextDisplayModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc", "enter", "backspace":
			return m, func() tea.Msg {
				return NavigateMsg{Screen: m.returnScreen}
			}
		}
	}

	return m, nil
}

func (m TextDisplayModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.theme.Title.Render(m.title)
	content := m.theme.MenuItem.Render(m.content)
	help := m.theme.Help.Render("esc/enter: back • q: quit")

	sections := []string{
		header,
		"",
		content,
		"",
		help,
	}

	contentSection := lipgloss.JoinVertical(lipgloss.Left, sections...)
	bordered := m.theme.BorderStyle.Render(contentSection)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}
