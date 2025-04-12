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

	// TextStyles
	TextStyle() Style
	TextSelectedStyle() Style

	// Button styles based on state
	ButtonStyle() Style                  // Base style (equivalent to ButtonNormalStyle)
	ButtonFocusedStyle() Style           // Base focused style
	ButtonSelectedStyle() Style          // Selected, not focused
	ButtonInteractedStyle() Style        // Interacted, not focused
	ButtonFocusedSelectedStyle() Style   // Selected and focused
	ButtonFocusedInteractedStyle() Style // Interacted and focused

	// List styles based on state
	ListStyle() Style                  // Base style (equivalent to ListNormalStyle)
	ListFocusedStyle() Style           // Base focused style
	ListSelectedStyle() Style          // Selected, not focused
	ListInteractedStyle() Style        // Interacted, not focused
	ListFocusedSelectedStyle() Style   // Selected and focused
	ListFocusedInteractedStyle() Style // Interacted and focused

	// Grid styles based on state
	GridStyle() Style                  // Base style (equivalent to GridNormalStyle)
	GridFocusedStyle() Style           // Base focused style
	GridSelectedStyle() Style          // Selected, not focused
	GridInteractedStyle() Style        // Interacted, not focused
	GridFocusedSelectedStyle() Style   // Selected and focused
	GridFocusedInteractedStyle() Style // Interacted and focused

	// Pane styles
	PaneStyle() Style
	PaneBorderStyle() Style
	PaneFocusBorderStyle() Style

	// Default border type for panes
	DefaultBorderType() BorderType

	// Default cell dimensions for grid
	DefaultCellWidth() int
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

// Convenience methods to get the appropriate style for a widget based on its type, state, and focus
func GetButtonStyle(theme Theme, state WidgetState, focused bool) Style {
	switch {
	case focused && state == StateInteracted:
		return theme.ButtonFocusedInteractedStyle()
	case focused && state == StateSelected:
		return theme.ButtonFocusedSelectedStyle()
	case focused:
		return theme.ButtonFocusedStyle()
	case state == StateInteracted:
		return theme.ButtonInteractedStyle()
	case state == StateSelected:
		return theme.ButtonSelectedStyle()
	default:
		return theme.ButtonStyle()
	}
}

func GetListStyle(theme Theme, state WidgetState, focused bool) Style {
	switch {
	case focused && state == StateInteracted:
		return theme.ListFocusedInteractedStyle()
	case focused && state == StateSelected:
		return theme.ListFocusedSelectedStyle()
	case focused:
		return theme.ListFocusedStyle()
	case state == StateInteracted:
		return theme.ListInteractedStyle()
	case state == StateSelected:
		return theme.ListSelectedStyle()
	default:
		return theme.ListStyle()
	}
}

func GetGridStyle(theme Theme, state WidgetState, focused bool) Style {
	switch {
	case focused && state == StateInteracted:
		return theme.GridFocusedInteractedStyle()
	case focused && state == StateSelected:
		return theme.GridFocusedSelectedStyle()
	case focused:
		return theme.GridFocusedStyle()
	case state == StateInteracted:
		return theme.GridInteractedStyle()
	case state == StateSelected:
		return theme.GridSelectedStyle()
	default:
		return theme.GridStyle()
	}
}