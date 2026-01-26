package theme

import (
	"os"
	"testing"
)

func TestDetectTerminalCapabilities_TrueColor(t *testing.T) {
	// Save original env vars
	origColorTerm := os.Getenv("COLORTERM")
	origTerm := os.Getenv("TERM")
	defer func() {
		os.Setenv("COLORTERM", origColorTerm)
		os.Setenv("TERM", origTerm)
	}()

	// Test truecolor detection
	os.Setenv("COLORTERM", "truecolor")
	os.Setenv("TERM", "xterm-256color")

	caps := DetectTerminalCapabilities()

	if !caps.TrueColor {
		t.Error("expected TrueColor to be true when COLORTERM=truecolor")
	}
	if !caps.Color256 {
		t.Error("expected Color256 to be true when COLORTERM=truecolor")
	}
}

func TestDetectTerminalCapabilities_24bit(t *testing.T) {
	origColorTerm := os.Getenv("COLORTERM")
	defer os.Setenv("COLORTERM", origColorTerm)

	os.Setenv("COLORTERM", "24bit")

	caps := DetectTerminalCapabilities()

	if !caps.TrueColor {
		t.Error("expected TrueColor to be true when COLORTERM=24bit")
	}
}

func TestDetectTerminalCapabilities_WindowsTerminal(t *testing.T) {
	origWTSession := os.Getenv("WT_SESSION")
	defer os.Setenv("WT_SESSION", origWTSession)

	os.Setenv("WT_SESSION", "some-session-id")

	caps := DetectTerminalCapabilities()

	if !caps.TrueColor {
		t.Error("expected TrueColor to be true in Windows Terminal")
	}
	if !caps.Color256 {
		t.Error("expected Color256 to be true in Windows Terminal")
	}
}

func TestDetectTerminalCapabilities_256Color(t *testing.T) {
	origTerm := os.Getenv("TERM")
	origColorTerm := os.Getenv("COLORTERM")
	defer func() {
		os.Setenv("TERM", origTerm)
		os.Setenv("COLORTERM", origColorTerm)
	}()

	os.Setenv("COLORTERM", "")

	tests := []struct {
		name     string
		term     string
		expected bool
	}{
		{"xterm-256color", "xterm-256color", true},
		{"screen-256color", "screen-256color", true},
		{"tmux-256color", "tmux-256color", true},
		{"xterm", "xterm", true},
		{"screen", "screen", true},
		{"tmux", "tmux", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TERM", tt.term)
			caps := DetectTerminalCapabilities()
			if caps.Color256 != tt.expected {
				t.Errorf("expected Color256=%v for TERM=%s, got %v", tt.expected, tt.term, caps.Color256)
			}
		})
	}
}

func TestDetectTerminalCapabilities_BasicTerminal(t *testing.T) {
	origTerm := os.Getenv("TERM")
	origColorTerm := os.Getenv("COLORTERM")
	defer func() {
		os.Setenv("TERM", origTerm)
		os.Setenv("COLORTERM", origColorTerm)
	}()

	os.Setenv("COLORTERM", "")

	tests := []struct {
		name        string
		term        string
		isBasicTerm bool
		unicode     bool
	}{
		{"dumb terminal", "dumb", true, false},
		{"vt100 terminal", "vt100", true, false},
		{"empty TERM", "", true, false},
		{"linux console", "linux", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TERM", tt.term)
			caps := DetectTerminalCapabilities()
			if caps.IsBasicTerm != tt.isBasicTerm {
				t.Errorf("expected IsBasicTerm=%v for TERM=%s, got %v", tt.isBasicTerm, tt.term, caps.IsBasicTerm)
			}
			if caps.Unicode != tt.unicode {
				t.Errorf("expected Unicode=%v for TERM=%s, got %v", tt.unicode, tt.term, caps.Unicode)
			}
		})
	}
}

func TestDetectTerminalCapabilities_WebTerminals(t *testing.T) {
	// Save and restore all env vars
	envVars := []string{"TERM_PROGRAM", "LC_TERMINAL", "WETTY_HOST", "GOTTY_TERM", "COLORTERM"}
	origValues := make(map[string]string)
	for _, v := range envVars {
		origValues[v] = os.Getenv(v)
	}
	defer func() {
		for k, v := range origValues {
			os.Setenv(k, v)
		}
	}()

	// Clear all
	for _, v := range envVars {
		os.Setenv(v, "")
	}

	tests := []struct {
		name    string
		envVar  string
		value   string
		isXterm bool
	}{
		{"WETTY_HOST set", "WETTY_HOST", "localhost", true},
		{"GOTTY_TERM set", "GOTTY_TERM", "xterm", true},
		{"web in TERM_PROGRAM", "TERM_PROGRAM", "web-terminal", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset
			for _, v := range envVars {
				os.Setenv(v, "")
			}
			os.Setenv(tt.envVar, tt.value)

			caps := DetectTerminalCapabilities()
			if caps.IsXtermJS != tt.isXterm {
				t.Errorf("expected IsXtermJS=%v when %s=%s, got %v", tt.isXterm, tt.envVar, tt.value, caps.IsXtermJS)
			}
		})
	}
}

func TestGetSymbols_Unicode(t *testing.T) {
	caps := TerminalCapabilities{
		Unicode:     true,
		IsBasicTerm: false,
	}

	symbols := GetSymbols(caps)

	// Check Unicode symbols are returned
	if symbols.Cursor != "▶" {
		t.Errorf("expected Unicode cursor '▶', got '%s'", symbols.Cursor)
	}
	if symbols.CheckMark != "✓" {
		t.Errorf("expected Unicode checkmark '✓', got '%s'", symbols.CheckMark)
	}
	if symbols.CrossMark != "✗" {
		t.Errorf("expected Unicode crossmark '✗', got '%s'", symbols.CrossMark)
	}
	if symbols.Warning != "⚠" {
		t.Errorf("expected Unicode warning '⚠', got '%s'", symbols.Warning)
	}
	if symbols.ArrowUp != "↑" {
		t.Errorf("expected Unicode arrow '↑', got '%s'", symbols.ArrowUp)
	}
	if symbols.Bullet != "•" {
		t.Errorf("expected Unicode bullet '•', got '%s'", symbols.Bullet)
	}
	if len(symbols.Spinner) != 10 {
		t.Errorf("expected 10 spinner frames, got %d", len(symbols.Spinner))
	}
}

func TestGetSymbols_ASCII(t *testing.T) {
	caps := TerminalCapabilities{
		Unicode:     false,
		IsBasicTerm: true,
	}

	symbols := GetSymbols(caps)

	// Check ASCII fallback symbols
	if symbols.Cursor != ">" {
		t.Errorf("expected ASCII cursor '>', got '%s'", symbols.Cursor)
	}
	if symbols.CheckMark != "[x]" {
		t.Errorf("expected ASCII checkmark '[x]', got '%s'", symbols.CheckMark)
	}
	if symbols.CrossMark != "[!]" {
		t.Errorf("expected ASCII crossmark '[!]', got '%s'", symbols.CrossMark)
	}
	if symbols.Warning != "[!]" {
		t.Errorf("expected ASCII warning '[!]', got '%s'", symbols.Warning)
	}
	if symbols.ArrowUp != "^" {
		t.Errorf("expected ASCII arrow '^', got '%s'", symbols.ArrowUp)
	}
	if symbols.Bullet != "*" {
		t.Errorf("expected ASCII bullet '*', got '%s'", symbols.Bullet)
	}
	if len(symbols.Spinner) != 4 {
		t.Errorf("expected 4 ASCII spinner frames, got %d", len(symbols.Spinner))
	}
}

func TestGetSymbols_BasicTermOverride(t *testing.T) {
	// Even with Unicode true, IsBasicTerm should force ASCII
	caps := TerminalCapabilities{
		Unicode:     true,
		IsBasicTerm: true,
	}

	symbols := GetSymbols(caps)

	// Should use ASCII fallback
	if symbols.Cursor != ">" {
		t.Errorf("expected ASCII cursor '>' when IsBasicTerm=true, got '%s'", symbols.Cursor)
	}
}

func TestSymbolsAllFieldsPopulated(t *testing.T) {
	// Test that all symbol fields are populated for both modes
	testCases := []struct {
		name string
		caps TerminalCapabilities
	}{
		{"Unicode", TerminalCapabilities{Unicode: true, IsBasicTerm: false}},
		{"ASCII", TerminalCapabilities{Unicode: false, IsBasicTerm: true}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			symbols := GetSymbols(tc.caps)

			if symbols.Cursor == "" {
				t.Error("Cursor should not be empty")
			}
			if symbols.CheckMark == "" {
				t.Error("CheckMark should not be empty")
			}
			if symbols.CrossMark == "" {
				t.Error("CrossMark should not be empty")
			}
			if symbols.Warning == "" {
				t.Error("Warning should not be empty")
			}
			if symbols.Info == "" {
				t.Error("Info should not be empty")
			}
			if symbols.ArrowUp == "" {
				t.Error("ArrowUp should not be empty")
			}
			if symbols.ArrowDown == "" {
				t.Error("ArrowDown should not be empty")
			}
			if symbols.ArrowLeft == "" {
				t.Error("ArrowLeft should not be empty")
			}
			if symbols.ArrowRight == "" {
				t.Error("ArrowRight should not be empty")
			}
			if symbols.Bullet == "" {
				t.Error("Bullet should not be empty")
			}
			if symbols.Box == "" {
				t.Error("Box should not be empty")
			}
			if symbols.BoxChecked == "" {
				t.Error("BoxChecked should not be empty")
			}
			if len(symbols.Spinner) == 0 {
				t.Error("Spinner should not be empty")
			}
			if symbols.BorderH == "" {
				t.Error("BorderH should not be empty")
			}
			if symbols.BorderV == "" {
				t.Error("BorderV should not be empty")
			}
			if symbols.CornerTL == "" {
				t.Error("CornerTL should not be empty")
			}
			if symbols.CornerTR == "" {
				t.Error("CornerTR should not be empty")
			}
			if symbols.CornerBL == "" {
				t.Error("CornerBL should not be empty")
			}
			if symbols.CornerBR == "" {
				t.Error("CornerBR should not be empty")
			}
			if symbols.Copy == "" {
				t.Error("Copy should not be empty")
			}
		})
	}
}
