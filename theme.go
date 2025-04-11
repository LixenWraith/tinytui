// theme.go
package tinytui

import (
	"sync"
)

// ThemeName represents a predefined theme identifier
type ThemeName string

const (
	// ThemeDefault is the fallback theme
	ThemeDefault ThemeName = "default"
	// ThemeTokyoNight is inspired by the Tokyo Night VSCode theme
	ThemeTokyoNight ThemeName = "tokyo-night"
	// ThemeCatppuccinMocha is inspired by the Catppuccin Mocha color scheme
	ThemeCatppuccinMocha ThemeName = "catppuccin-mocha"
	// ThemeBorland is inspired by classic Borland DOS applications
	ThemeBorland ThemeName = "borland"
)

// Theme defines the interface for a UI theme
type Theme interface {
	// Name returns the unique name of the theme
	Name() ThemeName

	// TextStyle returns the default style for text widgets
	TextStyle() Style
	// TextSelectedStyle returns the style for selected text
	TextSelectedStyle() Style

	// ButtonStyle returns the style for buttons
	ButtonStyle() Style
	// ButtonFocusedStyle returns the style for focused buttons
	ButtonFocusedStyle() Style

	// ListStyle returns the style for list items
	ListStyle() Style
	// ListSelectedStyle returns the style for selected list items
	ListSelectedStyle() Style

	// GridStyle returns the style for grid cells
	GridStyle() Style
	// GridSelectedStyle returns the style for selected grid cells
	GridSelectedStyle() Style

	// PaneStyle returns the style for pane content areas
	PaneStyle() Style
	// PaneBorderStyle returns the style for pane borders
	PaneBorderStyle() Style
	// PaneFocusBorderStyle returns the style for focused pane borders
	PaneFocusBorderStyle() Style

	// DefaultBorderType returns the default border type for panes
	DefaultBorderType() BorderType

	// DefaultCellWidth returns the default width for grid cells
	DefaultCellWidth() int
	// DefaultCellHeight returns the default height for grid cells
	DefaultCellHeight() int
}

// themeManager handles the current theme and theme switching
type themeManager struct {
	current     Theme
	themes      map[ThemeName]Theme
	mu          sync.RWMutex
	subscribers []func(Theme)
}

var (
	// Global theme manager instance
	globalThemeManager = &themeManager{
		themes:      make(map[ThemeName]Theme),
		subscribers: make([]func(Theme), 0),
	}
)

// RegisterTheme adds a theme to the available themes
func RegisterTheme(theme Theme) {
	if theme == nil {
		return
	}

	globalThemeManager.mu.Lock()
	defer globalThemeManager.mu.Unlock()

	name := theme.Name()
	globalThemeManager.themes[name] = theme

	// If this is the first theme registered, set it as current
	if globalThemeManager.current == nil {
		globalThemeManager.current = theme
	}
}

// SetTheme changes the current theme
func SetTheme(name ThemeName) bool {
	globalThemeManager.mu.Lock()
	defer globalThemeManager.mu.Unlock()

	theme, ok := globalThemeManager.themes[name]
	if !ok {
		return false
	}

	globalThemeManager.current = theme

	// Notify subscribers
	for _, subscriber := range globalThemeManager.subscribers {
		subscriber(theme)
	}

	return true
}

// GetTheme returns the current theme
func GetTheme() Theme {
	globalThemeManager.mu.RLock()
	defer globalThemeManager.mu.RUnlock()

	return globalThemeManager.current
}

// SubscribeThemeChange registers a function to be called when the theme changes
func SubscribeThemeChange(callback func(Theme)) {
	if callback == nil {
		return
	}

	globalThemeManager.mu.Lock()
	defer globalThemeManager.mu.Unlock()

	globalThemeManager.subscribers = append(globalThemeManager.subscribers, callback)

	// Call immediately with current theme
	if globalThemeManager.current != nil {
		callback(globalThemeManager.current)
	}
}

// Initialize default themes
func init() {
	// Register all predefined themes
	RegisterTheme(NewDefaultTheme())
	RegisterTheme(NewTokyoNightTheme())
	RegisterTheme(NewCatppuccinMochaTheme())
	RegisterTheme(NewBorlandTheme())

	// Set default theme
	SetTheme(ThemeDefault)
}

// DefaultTextStyle returns the current theme's text style
func DefaultTextStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.TextStyle()
	}
	return DefaultStyle
}

// DefaultTextSelectedStyle returns the current theme's selected text style
func DefaultTextSelectedStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.TextSelectedStyle()
	}
	return DefaultStyle.Reverse(true)
}

// DefaultButtonStyle returns the current theme's button style
func DefaultButtonStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.ButtonStyle()
	}
	return DefaultStyle
}

// DefaultButtonFocusedStyle returns the current theme's focused button style
func DefaultButtonFocusedStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.ButtonFocusedStyle()
	}
	return DefaultStyle.Reverse(true)
}

// DefaultListStyle returns the current theme's list style
func DefaultListStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.ListStyle()
	}
	return DefaultStyle
}

// DefaultListSelectedStyle returns the current theme's selected list item style
func DefaultListSelectedStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.ListSelectedStyle()
	}
	return DefaultStyle.Reverse(true)
}

// DefaultGridStyle returns the current theme's grid style
func DefaultGridStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.GridStyle()
	}
	return DefaultStyle
}

// DefaultGridSelectedStyle returns the current theme's selected grid cell style
func DefaultGridSelectedStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.GridSelectedStyle()
	}
	return DefaultStyle.Reverse(true)
}

// DefaultPaneStyle returns the current theme's pane style
func DefaultPaneStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.PaneStyle()
	}
	return DefaultStyle
}

// DefaultPaneBorderStyle returns the current theme's pane border style
func DefaultPaneBorderStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.PaneBorderStyle()
	}
	return DefaultStyle
}

// DefaultPaneFocusBorderStyle returns the current theme's focused pane border style
func DefaultPaneFocusBorderStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.PaneFocusBorderStyle()
	}
	return DefaultStyle.Foreground(ColorYellow).Bold(true)
}

// DefaultBorderType returns the current theme's default border type
func DefaultBorderType() BorderType {
	theme := GetTheme()
	if theme != nil {
		return theme.DefaultBorderType()
	}
	return BorderSingle
}

// DefaultCellWidth returns the current theme's default cell width
func DefaultCellWidth() int {
	theme := GetTheme()
	if theme != nil {
		return theme.DefaultCellWidth()
	}
	return 10 // Default value
}

// DefaultCellHeight returns the current theme's default cell height
func DefaultCellHeight() int {
	theme := GetTheme()
	if theme != nil {
		return theme.DefaultCellHeight()
	}
	return 1 // Default value
}