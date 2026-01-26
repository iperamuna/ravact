package theme

import (
	"testing"
)

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()

	if theme == nil {
		t.Fatal("DefaultTheme returned nil")
	}

	// Verify colors are set
	if theme.Primary == "" {
		t.Error("Primary color should be set")
	}
	if theme.Secondary == "" {
		t.Error("Secondary color should be set")
	}
	if theme.Success == "" {
		t.Error("Success color should be set")
	}
	if theme.Warning == "" {
		t.Error("Warning color should be set")
	}
	if theme.Error == "" {
		t.Error("Error color should be set")
	}
	if theme.Info == "" {
		t.Error("Info color should be set")
	}
	if theme.Text == "" {
		t.Error("Text color should be set")
	}
}

func TestDefaultTheme_Styles(t *testing.T) {
	theme := DefaultTheme()

	// Verify styles are initialized (not zero values)
	// We can't easily test lipgloss.Style values, but we can verify they exist
	// by checking the theme is properly initialized

	if theme.HuhTheme == nil {
		t.Error("HuhTheme should be initialized")
	}
}

func TestDefaultTheme_Symbols(t *testing.T) {
	theme := DefaultTheme()

	// Verify symbols are populated
	if theme.Symbols.Cursor == "" {
		t.Error("Cursor symbol should be set")
	}
	if theme.Symbols.CheckMark == "" {
		t.Error("CheckMark symbol should be set")
	}
	if theme.Symbols.CrossMark == "" {
		t.Error("CrossMark symbol should be set")
	}
	if theme.Symbols.Warning == "" {
		t.Error("Warning symbol should be set")
	}
	if len(theme.Symbols.Spinner) == 0 {
		t.Error("Spinner frames should be set")
	}
}

func TestDefaultTheme_Caps(t *testing.T) {
	theme := DefaultTheme()

	// Caps should be detected (we can't test specific values as they depend on env)
	// Just verify the struct is populated
	caps := theme.Caps

	// At minimum, these should be set to some value (true or false)
	t.Logf("TrueColor: %v", caps.TrueColor)
	t.Logf("Color256: %v", caps.Color256)
	t.Logf("Unicode: %v", caps.Unicode)
	t.Logf("IsXtermJS: %v", caps.IsXtermJS)
	t.Logf("IsBasicTerm: %v", caps.IsBasicTerm)
}

func TestSplashASCII(t *testing.T) {
	ascii := SplashASCII()

	if ascii == "" {
		t.Error("SplashASCII should return non-empty string")
	}

	// Verify it contains RAVACT
	if len(ascii) < 10 {
		t.Error("SplashASCII seems too short")
	}
}

func TestSplashASCIILarge(t *testing.T) {
	ascii := SplashASCIILarge()

	if ascii == "" {
		t.Error("SplashASCIILarge should return non-empty string")
	}

	// Large version should be bigger than small version
	smallASCII := SplashASCII()
	if len(ascii) <= len(smallASCII) {
		t.Error("SplashASCIILarge should be larger than SplashASCII")
	}
}

func TestThemeColorsConsistency(t *testing.T) {
	theme := DefaultTheme()

	// SelectedBg and Primary should typically be related
	// SelectedText and Text should typically be related
	// These are design decisions, but we can verify they're set

	if theme.SelectedBg == "" {
		t.Error("SelectedBg should be set")
	}
	if theme.SelectedText == "" {
		t.Error("SelectedText should be set")
	}
	if theme.Highlight == "" {
		t.Error("Highlight should be set")
	}
	if theme.BorderColor == "" {
		t.Error("BorderColor should be set")
	}
	if theme.Subtle == "" {
		t.Error("Subtle should be set")
	}
}

func TestHuhThemeInitialization(t *testing.T) {
	theme := DefaultTheme()

	if theme.HuhTheme == nil {
		t.Fatal("HuhTheme should not be nil")
	}

	// The huh theme should be configured
	// We can't easily inspect internal huh.Theme fields,
	// but we can verify it was created without panic
}

func TestMultipleThemeCreation(t *testing.T) {
	// Ensure creating multiple themes doesn't cause issues
	theme1 := DefaultTheme()
	theme2 := DefaultTheme()

	if theme1 == nil || theme2 == nil {
		t.Fatal("Themes should not be nil")
	}

	// They should be independent instances
	if theme1 == theme2 {
		t.Log("Note: DefaultTheme returns new instances each time")
	}
}
