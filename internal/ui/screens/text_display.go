package screens

import (
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// TextDisplayModel represents a text display screen
type TextDisplayModel struct {
	theme        *theme.Theme
	width        int
	height       int
	title        string
	content      string
	returnScreen ScreenType
	copied       bool
	copiedTimer  int
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
		case "c":
			// Copy content to clipboard
			if m.content != "" {
				clipboard.WriteAll(m.content)
				m.copied = true
				m.copiedTimer = 3
				return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
					return CopyTimerTickMsg{}
				})
			}
		}

	case CopyTimerTickMsg:
		if m.copiedTimer > 0 {
			m.copiedTimer--
			if m.copiedTimer == 0 {
				m.copied = false
			} else {
				return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
					return CopyTimerTickMsg{}
				})
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

	// Copied indicator
	copiedMsg := ""
	if m.copied {
		copiedMsg = m.theme.CopiedStyle.Render(m.theme.Symbols.Copy + " Copied to clipboard!")
	}

	help := m.theme.Help.Render("c: Copy " + m.theme.Symbols.Bullet + " Esc/Enter: Back " + m.theme.Symbols.Bullet + " q: Quit")

	sections := []string{
		header,
		"",
		content,
		"",
	}

	if copiedMsg != "" {
		sections = append(sections, copiedMsg)
	}
	sections = append(sections, help)

	contentSection := lipgloss.JoinVertical(lipgloss.Left, sections...)
	bordered := m.theme.RenderBox(contentSection)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bordered,
	)
}
