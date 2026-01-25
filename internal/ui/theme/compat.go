package theme

import (
	"os"
	"strings"
)

// TerminalCapabilities holds information about terminal capabilities
type TerminalCapabilities struct {
	TrueColor    bool
	Color256     bool
	Unicode      bool
	IsXtermJS    bool
	IsBasicTerm  bool
}

// DetectTerminalCapabilities detects the terminal's capabilities
func DetectTerminalCapabilities() TerminalCapabilities {
	term := os.Getenv("TERM")
	colorTerm := os.Getenv("COLORTERM")
	wtSession := os.Getenv("WT_SESSION")        // Windows Terminal
	xtermVersion := os.Getenv("XTERM_VERSION")  // Native xterm
	
	caps := TerminalCapabilities{
		TrueColor:   false,
		Color256:    false,
		Unicode:     true,
		IsXtermJS:   false,
		IsBasicTerm: false,
	}
	
	// Check for true color support
	if colorTerm == "truecolor" || colorTerm == "24bit" {
		caps.TrueColor = true
		caps.Color256 = true
	}
	
	// Windows Terminal supports true color
	if wtSession != "" {
		caps.TrueColor = true
		caps.Color256 = true
	}
	
	// Native xterm with version likely supports 256 colors
	if xtermVersion != "" {
		caps.Color256 = true
	}
	
	// Check TERM value for capabilities
	termLower := strings.ToLower(term)
	
	// 256 color terminals
	if strings.Contains(termLower, "256color") || strings.Contains(termLower, "256-color") {
		caps.Color256 = true
	}
	
	// xterm variants usually support 256 colors
	if strings.HasPrefix(termLower, "xterm") {
		caps.Color256 = true
	}
	
	// Screen/tmux usually support 256 colors
	if strings.HasPrefix(termLower, "screen") || strings.HasPrefix(termLower, "tmux") {
		caps.Color256 = true
	}
	
	// Detect xterm.js or web-based terminals
	// These often have limited capabilities or quirks
	termProgram := os.Getenv("TERM_PROGRAM")
	lcTerminal := os.Getenv("LC_TERMINAL")
	
	// Common web terminal indicators
	if strings.Contains(strings.ToLower(termProgram), "web") ||
		strings.Contains(strings.ToLower(lcTerminal), "web") ||
		os.Getenv("WETTY_HOST") != "" ||
		os.Getenv("GOTTY_TERM") != "" ||
		os.Getenv("TTYD_") != "" {
		caps.IsXtermJS = true
		caps.Color256 = true  // Most xterm.js implementations support 256 colors
		caps.TrueColor = false // But true color can be unreliable
	}
	
	// Basic/dumb terminals
	if termLower == "dumb" || termLower == "vt100" || termLower == "" {
		caps.IsBasicTerm = true
		caps.Color256 = false
		caps.TrueColor = false
		caps.Unicode = false
	}
	
	// Linux console has limited capabilities
	if termLower == "linux" {
		caps.Unicode = false
		caps.Color256 = false
	}
	
	return caps
}

// Symbols provides terminal-safe symbols based on capabilities
type Symbols struct {
	Cursor       string
	CursorEmpty  string
	CheckMark    string
	CrossMark    string
	Warning      string
	Info         string
	ArrowUp      string
	ArrowDown    string
	ArrowLeft    string
	ArrowRight   string
	Bullet       string
	Box          string
	BoxChecked   string
	Spinner      []string
	BorderH      string
	BorderV      string
	CornerTL     string
	CornerTR     string
	CornerBL     string
	CornerBR     string
	Copy         string
}

// GetSymbols returns appropriate symbols based on terminal capabilities
func GetSymbols(caps TerminalCapabilities) Symbols {
	if caps.Unicode && !caps.IsBasicTerm {
		return Symbols{
			Cursor:      "▶",
			CursorEmpty: " ",
			CheckMark:   "✓",
			CrossMark:   "✗",
			Warning:     "⚠",
			Info:        "ℹ",
			ArrowUp:     "↑",
			ArrowDown:   "↓",
			ArrowLeft:   "←",
			ArrowRight:  "→",
			Bullet:      "•",
			Box:         "☐",
			BoxChecked:  "☑",
			Spinner:     []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
			BorderH:     "─",
			BorderV:     "│",
			CornerTL:    "╭",
			CornerTR:    "╮",
			CornerBL:    "╰",
			CornerBR:    "╯",
			Copy:        "📋",
		}
	}
	
	// ASCII fallback for basic terminals
	return Symbols{
		Cursor:      ">",
		CursorEmpty: " ",
		CheckMark:   "[x]",
		CrossMark:   "[!]",
		Warning:     "[!]",
		Info:        "[i]",
		ArrowUp:     "^",
		ArrowDown:   "v",
		ArrowLeft:   "<",
		ArrowRight:  ">",
		Bullet:      "*",
		Box:         "[ ]",
		BoxChecked:  "[x]",
		Spinner:     []string{"|", "/", "-", "\\"},
		BorderH:     "-",
		BorderV:     "|",
		CornerTL:    "+",
		CornerTR:    "+",
		CornerBL:    "+",
		CornerBR:    "+",
		Copy:        "[C]",
	}
}
