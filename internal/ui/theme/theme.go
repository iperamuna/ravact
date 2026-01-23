package theme

import "github.com/charmbracelet/lipgloss"

// Theme defines the color scheme and styles for the application
type Theme struct {
	// Colors
	Primary      lipgloss.Color
	Secondary    lipgloss.Color
	Success      lipgloss.Color
	Warning      lipgloss.Color
	Error        lipgloss.Color
	Info         lipgloss.Color
	Subtle       lipgloss.Color
	Text         lipgloss.Color
	Background   lipgloss.Color
	BorderColor  lipgloss.Color
	Highlight    lipgloss.Color
	SelectedBg   lipgloss.Color
	SelectedText lipgloss.Color

	// Styles
	Title           lipgloss.Style
	Subtitle        lipgloss.Style
	MenuItem        lipgloss.Style
	SelectedItem    lipgloss.Style
	StatusBar       lipgloss.Style
	ErrorStyle      lipgloss.Style
	SuccessStyle    lipgloss.Style
	InfoStyle       lipgloss.Style
	WarningStyle    lipgloss.Style
	BorderStyle     lipgloss.Style
	Help            lipgloss.Style
	Prompt          lipgloss.Style
	Input           lipgloss.Style
	Label           lipgloss.Style
	Value           lipgloss.Style
	KeyStyle        lipgloss.Style
	DescriptionStyle lipgloss.Style
}

// DefaultTheme returns the default color scheme
func DefaultTheme() *Theme {
	t := &Theme{
		Primary:      lipgloss.Color("#FF6B35"), // Orange/Red (Ravana inspired)
		Secondary:    lipgloss.Color("#004E89"), // Deep blue
		Success:      lipgloss.Color("#2ECC71"), // Green
		Warning:      lipgloss.Color("#F39C12"), // Yellow
		Error:        lipgloss.Color("#E74C3C"), // Red
		Info:         lipgloss.Color("#3498DB"), // Blue
		Subtle:       lipgloss.Color("#7F8C8D"), // Gray
		Text:         lipgloss.Color("#FFFFFF"), // White
		Background:   lipgloss.Color("#1A1A1A"), // Dark background
		BorderColor:  lipgloss.Color("#404040"), // Gray border
		Highlight:    lipgloss.Color("#FFD700"), // Gold
		SelectedBg:   lipgloss.Color("#FF6B35"), // Orange
		SelectedText: lipgloss.Color("#FFFFFF"), // White
	}

	// Define styles
	t.Title = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true).
		Padding(1, 2)

	t.Subtitle = lipgloss.NewStyle().
		Foreground(t.Secondary).
		Italic(true).
		Padding(0, 2)

	t.MenuItem = lipgloss.NewStyle().
		Foreground(t.Text).
		Padding(0, 2)

	t.SelectedItem = lipgloss.NewStyle().
		Foreground(t.SelectedText).
		Background(t.SelectedBg).
		Bold(true).
		Padding(0, 2)

	t.StatusBar = lipgloss.NewStyle().
		Foreground(t.Subtle).
		Background(lipgloss.Color("#2A2A2A")).
		Padding(0, 1)

	t.ErrorStyle = lipgloss.NewStyle().
		Foreground(t.Error).
		Bold(true).
		Padding(0, 1)

	t.SuccessStyle = lipgloss.NewStyle().
		Foreground(t.Success).
		Bold(true).
		Padding(0, 1)

	t.InfoStyle = lipgloss.NewStyle().
		Foreground(t.Info).
		Padding(0, 1)

	t.WarningStyle = lipgloss.NewStyle().
		Foreground(t.Warning).
		Bold(true).
		Padding(0, 1)

	t.BorderStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderColor).
		Padding(1, 2)

	t.Help = lipgloss.NewStyle().
		Foreground(t.Subtle).
		Italic(true)

	t.Prompt = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true)

	t.Input = lipgloss.NewStyle().
		Foreground(t.Text).
		Background(lipgloss.Color("#2A2A2A")).
		Padding(0, 1)

	t.Label = lipgloss.NewStyle().
		Foreground(t.Secondary).
		Bold(true)

	t.Value = lipgloss.NewStyle().
		Foreground(t.Text)

	t.KeyStyle = lipgloss.NewStyle().
		Foreground(t.Highlight).
		Bold(true)

	t.DescriptionStyle = lipgloss.NewStyle().
		Foreground(t.Subtle).
		Italic(true)

	return t
}

// ASCII art for splash screen
func SplashASCII() string {
	return `
╦═╗╔═╗╦  ╦╔═╗╔═╗╔╦╗
╠╦╝╠═╣╚╗╔╝╠═╣║   ║ 
╩╚═╩ ╩ ╚╝ ╩ ╩╚═╝ ╩ 
`
}

// Alternative larger ASCII art
func SplashASCIILarge() string {
	return `
██████╗  ██████╗ ██╗   ██╗ █████╗  ██████╗████████╗
██╔══██╗██╔═══██╗██║   ██║██╔══██╗██╔════╝╚══██╔══╝
██████╔╝███████║╚██╗ ██╔╝███████║██║        ██║   
██╔══██╗██╔══██║ ╚████╔╝ ██╔══██║██║        ██║   
██║  ██║██║  ██║  ╚██╔╝  ██║  ██║╚██████╗   ██║   
╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝ ╚═════╝   ╚═╝   
`
}
