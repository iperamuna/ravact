package theme

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

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
	Title            lipgloss.Style
	Subtitle         lipgloss.Style
	MenuItem         lipgloss.Style
	SelectedItem     lipgloss.Style
	StatusBar        lipgloss.Style
	ErrorStyle       lipgloss.Style
	SuccessStyle     lipgloss.Style
	InfoStyle        lipgloss.Style
	WarningStyle     lipgloss.Style
	BorderStyle      lipgloss.Style
	Help             lipgloss.Style
	Prompt           lipgloss.Style
	Input            lipgloss.Style
	Label            lipgloss.Style
	Value            lipgloss.Style
	KeyStyle         lipgloss.Style
	DescriptionStyle lipgloss.Style
	CopiedStyle      lipgloss.Style
	CategoryStyle    lipgloss.Style

	// Terminal capabilities and symbols
	Caps    TerminalCapabilities
	Symbols Symbols

	// Huh form theme
	HuhTheme *huh.Theme

	// Layout
	AppWidth int
}

// DefaultTheme returns the default color scheme
func DefaultTheme() *Theme {
	caps := DetectTerminalCapabilities()
	symbols := GetSymbols(caps)

	// Use ANSI 256 colors for better xterm.js compatibility
	// These are more widely supported than true color hex values
	var t *Theme

	if caps.TrueColor {
		// True color supported - use hex colors
		t = &Theme{
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
			Caps:         caps,
			Symbols:      symbols,
		}
	} else if caps.Color256 {
		// 256 color mode - use ANSI 256 color codes
		t = &Theme{
			Primary:      lipgloss.Color("208"), // Orange
			Secondary:    lipgloss.Color("24"),  // Deep blue
			Success:      lipgloss.Color("34"),  // Green
			Warning:      lipgloss.Color("220"), // Yellow
			Error:        lipgloss.Color("196"), // Red
			Info:         lipgloss.Color("33"),  // Blue
			Subtle:       lipgloss.Color("245"), // Gray
			Text:         lipgloss.Color("15"),  // White
			Background:   lipgloss.Color("234"), // Dark background
			BorderColor:  lipgloss.Color("240"), // Gray border
			Highlight:    lipgloss.Color("220"), // Gold/Yellow
			SelectedBg:   lipgloss.Color("208"), // Orange
			SelectedText: lipgloss.Color("15"),  // White
			Caps:         caps,
			Symbols:      symbols,
		}
	} else {
		// Basic 16 color mode
		t = &Theme{
			Primary:      lipgloss.Color("9"),  // Bright Red
			Secondary:    lipgloss.Color("4"),  // Blue
			Success:      lipgloss.Color("2"),  // Green
			Warning:      lipgloss.Color("3"),  // Yellow
			Error:        lipgloss.Color("1"),  // Red
			Info:         lipgloss.Color("6"),  // Cyan
			Subtle:       lipgloss.Color("8"),  // Gray
			Text:         lipgloss.Color("15"), // White
			Background:   lipgloss.Color("0"),  // Black
			BorderColor:  lipgloss.Color("8"),  // Gray
			Highlight:    lipgloss.Color("11"), // Bright Yellow
			SelectedBg:   lipgloss.Color("9"),  // Bright Red
			SelectedText: lipgloss.Color("15"), // White
			Caps:         caps,
			Symbols:      symbols,
		}
	}

	t.AppWidth = 90

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

	// Use appropriate border style based on terminal capabilities
	// Use fixed width for consistency across the app
	borderWidth := t.AppWidth + 6 // Account for borders (2) and padding (4)
	if caps.Unicode && !caps.IsBasicTerm {
		t.BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(t.BorderColor).
			Padding(1, 2).
			Width(borderWidth)
	} else {
		// ASCII-safe border for basic terminals
		t.BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(t.BorderColor).
			Padding(1, 2).
			Width(borderWidth)
	}

	t.Help = lipgloss.NewStyle().
		Foreground(t.Subtle).
		Italic(true)

	t.Prompt = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true)

	t.Input = lipgloss.NewStyle().
		Foreground(t.Text).
		Background(t.Background).
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

	t.CopiedStyle = lipgloss.NewStyle().
		Foreground(t.Success).
		Bold(true).
		Padding(0, 1)

	t.CategoryStyle = lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true).
		Padding(0, 0, 0, 0).
		MarginTop(1)

	// Create custom huh theme matching app colors
	t.HuhTheme = createHuhTheme(t)

	return t
}

// RenderBox wraps content to AppWidth and applies the BorderStyle.
// This ensures consistent width and text wrapping across all screens.
func (t *Theme) RenderBox(content string) string {
	// Wrap the content to AppWidth
	// lipgloss.NewStyle().Width(w).Render(text) handles multi-line strings correctly
	wrapped := lipgloss.NewStyle().Width(t.AppWidth).Render(content)

	// Apply the border style
	return t.BorderStyle.Render(wrapped)
}

// createHuhTheme creates a custom huh theme matching the app's color scheme
func createHuhTheme(t *Theme) *huh.Theme {
	theme := huh.ThemeBase()

	// Form styles
	theme.Form.Base = lipgloss.NewStyle().Padding(1, 0)

	// Group styles
	theme.Group.Base = lipgloss.NewStyle().Padding(0, 0)

	// Field separator
	theme.FieldSeparator = lipgloss.NewStyle().SetString("\n")

	// Blurred (unfocused) field styles
	theme.Blurred.Base = lipgloss.NewStyle().
		PaddingLeft(1).
		BorderStyle(lipgloss.HiddenBorder()).
		BorderLeft(true)
	theme.Blurred.Title = lipgloss.NewStyle().Foreground(t.Text)
	theme.Blurred.Description = lipgloss.NewStyle().Foreground(t.Subtle)
	theme.Blurred.ErrorIndicator = lipgloss.NewStyle().Foreground(t.Error).SetString(" *")
	theme.Blurred.ErrorMessage = lipgloss.NewStyle().Foreground(t.Error)
	theme.Blurred.SelectSelector = lipgloss.NewStyle().Foreground(t.Subtle).SetString("> ")
	theme.Blurred.NextIndicator = lipgloss.NewStyle().Foreground(t.Subtle)
	theme.Blurred.PrevIndicator = lipgloss.NewStyle().Foreground(t.Subtle)
	theme.Blurred.Option = lipgloss.NewStyle().Foreground(t.Text)
	theme.Blurred.MultiSelectSelector = lipgloss.NewStyle().Foreground(t.Subtle).SetString("> ")
	theme.Blurred.SelectedOption = lipgloss.NewStyle().Foreground(t.Success)
	theme.Blurred.SelectedPrefix = lipgloss.NewStyle().Foreground(t.Success).SetString("[✓] ")
	theme.Blurred.UnselectedOption = lipgloss.NewStyle().Foreground(t.Text)
	theme.Blurred.UnselectedPrefix = lipgloss.NewStyle().Foreground(t.Subtle).SetString("[ ] ")
	theme.Blurred.FocusedButton = lipgloss.NewStyle().
		Foreground(t.Text).
		Background(t.Primary).
		Padding(0, 2).
		Bold(true)
	theme.Blurred.BlurredButton = lipgloss.NewStyle().
		Foreground(t.Subtle).
		Background(t.BorderColor).
		Padding(0, 2)
	theme.Blurred.TextInput.Cursor = lipgloss.NewStyle().Foreground(t.Primary)
	theme.Blurred.TextInput.Placeholder = lipgloss.NewStyle().Foreground(t.Subtle)
	theme.Blurred.TextInput.Prompt = lipgloss.NewStyle().Foreground(t.Subtle)
	theme.Blurred.TextInput.Text = lipgloss.NewStyle().Foreground(t.Text)
	theme.Blurred.Card = lipgloss.NewStyle().PaddingLeft(1)
	theme.Blurred.NoteTitle = lipgloss.NewStyle().Foreground(t.Info).Bold(true)

	// Focused field styles
	theme.Focused.Base = lipgloss.NewStyle().
		PaddingLeft(1).
		BorderStyle(lipgloss.ThickBorder()).
		BorderLeft(true).
		BorderForeground(t.Primary)
	theme.Focused.Title = lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
	theme.Focused.Description = lipgloss.NewStyle().Foreground(t.Subtle)
	theme.Focused.ErrorIndicator = lipgloss.NewStyle().Foreground(t.Error).SetString(" *")
	theme.Focused.ErrorMessage = lipgloss.NewStyle().Foreground(t.Error)
	theme.Focused.SelectSelector = lipgloss.NewStyle().Foreground(t.Primary).SetString("> ")
	theme.Focused.NextIndicator = lipgloss.NewStyle().Foreground(t.Primary).SetString("↓ ")
	theme.Focused.PrevIndicator = lipgloss.NewStyle().Foreground(t.Primary).SetString("↑ ")
	theme.Focused.Option = lipgloss.NewStyle().Foreground(t.Text)
	theme.Focused.MultiSelectSelector = lipgloss.NewStyle().Foreground(t.Primary).SetString("> ")
	theme.Focused.SelectedOption = lipgloss.NewStyle().Foreground(t.Success).Bold(true)
	theme.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(t.Success).SetString("[✓] ")
	theme.Focused.UnselectedOption = lipgloss.NewStyle().Foreground(t.Text)
	theme.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(t.Subtle).SetString("[ ] ")
	theme.Focused.FocusedButton = lipgloss.NewStyle().
		Foreground(t.SelectedText).
		Background(t.Primary).
		Padding(0, 2).
		Bold(true)
	theme.Focused.BlurredButton = lipgloss.NewStyle().
		Foreground(t.Text).
		Background(t.BorderColor).
		Padding(0, 2)
	theme.Focused.TextInput.Cursor = lipgloss.NewStyle().Foreground(t.Primary)
	theme.Focused.TextInput.Placeholder = lipgloss.NewStyle().Foreground(t.Subtle)
	theme.Focused.TextInput.Prompt = lipgloss.NewStyle().Foreground(t.Primary)
	theme.Focused.TextInput.Text = lipgloss.NewStyle().Foreground(t.Text)
	theme.Focused.Card = lipgloss.NewStyle().PaddingLeft(1)
	theme.Focused.NoteTitle = lipgloss.NewStyle().Foreground(t.Primary).Bold(true)

	// Help styles
	theme.Help.Ellipsis = lipgloss.NewStyle().Foreground(t.Subtle)
	theme.Help.ShortKey = lipgloss.NewStyle().Foreground(t.Highlight).Bold(true)
	theme.Help.ShortDesc = lipgloss.NewStyle().Foreground(t.Subtle)
	theme.Help.ShortSeparator = lipgloss.NewStyle().Foreground(t.Subtle)
	theme.Help.FullKey = lipgloss.NewStyle().Foreground(t.Highlight).Bold(true)
	theme.Help.FullDesc = lipgloss.NewStyle().Foreground(t.Subtle)
	theme.Help.FullSeparator = lipgloss.NewStyle().Foreground(t.Subtle)

	return theme
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
