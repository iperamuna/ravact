package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// TextDisplayScreen displays text content with a title
type textDisplayScreen struct {
	title      string
	content    string
	returnTo   tea.Model
}

func NewTextDisplayScreen(title, content string, returnTo tea.Model) *textDisplayScreen {
	return &textDisplayScreen{
		title:    title,
		content:  content,
		returnTo: returnTo,
	}
}

func (m *textDisplayScreen) Init() tea.Cmd {
	return nil
}

func (m *textDisplayScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc", "enter":
			if m.returnTo != nil {
				return m.returnTo, nil
			}
			return NewMainMenuScreen(), nil
		}
	}

	return m, nil
}

func (m *textDisplayScreen) View() string {
	var b strings.Builder

	b.WriteString(theme.HeaderStyle.Render(m.title))
	b.WriteString("\n\n")

	b.WriteString(m.content)
	b.WriteString("\n\n")

	b.WriteString(theme.HelpStyle.Render("esc/enter: back • q: quit"))

	return b.String()
}
