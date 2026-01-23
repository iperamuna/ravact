package screens

import (
	"fmt"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iperamuna/ravact/internal/ui/theme"
)

// SplashModel represents the splash screen
type SplashModel struct {
	theme   *theme.Theme
	width   int
	height  int
	counter int
}

// NewSplashModel creates a new splash screen model
func NewSplashModel() SplashModel {
	return SplashModel{
		theme:   theme.DefaultTheme(),
		counter: 0,
	}
}

// Init initializes the splash screen
func (m SplashModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the splash screen
func (m SplashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Any key continues to main menu
		return m, func() tea.Msg {
			return NavigateMsg{Screen: MainMenuScreen}
		}
	}

	return m, nil
}

// View renders the splash screen
func (m SplashModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// ASCII art
	ascii := theme.SplashASCIILarge()
	styledASCII := m.theme.Title.Render(ascii)

	// Subtitle
	subtitle := m.theme.Subtitle.Render("Linux Server Management TUI")

	// Version info with architecture
	versionText := fmt.Sprintf("Version 0.1.0 (%s/%s)", runtime.GOOS, runtime.GOARCH)
	version := m.theme.InfoStyle.Render(versionText)

	// Tagline
	tagline := m.theme.DescriptionStyle.Render("Power and Control for Your Server Infrastructure")

	// Help text
	help := m.theme.Help.Render("Press any key to continue...")

	// Center everything
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styledASCII,
		"",
		subtitle,
		"",
		tagline,
		"",
		version,
		"",
		"",
		help,
	)

	// Center on screen
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// Width returns the current width
func (m SplashModel) Width() int {
	return m.width
}

// Height returns the current height
func (m SplashModel) Height() int {
	return m.height
}

// SetSize sets the window size
func (m *SplashModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
