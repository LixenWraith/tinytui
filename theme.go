// theme.go
package tinytui

import (
	"sync" // Use sync for thread-safe access to global theme manager
)

// ThemeName identifies a predefined theme (e.g., "default", "turbo").
// Used for registering and setting themes.
type ThemeName string

const (
	// ThemeDefault is the standard light-background fallback theme.
	ThemeDefault ThemeName = "default"
	// ThemeTurbo is a theme inspired by Turbo Vision's classic blue-background look.
	ThemeTurbo ThemeName = "turbo"
)

// Theme defines the interface for providing styles and properties for UI elements.
// Implementations of this interface determine the visual appearance of the application.
type Theme interface {
	// Name returns the unique identifier of the theme (e.g., "default", "turbo").
	Name() ThemeName

	// --- Style Getters ---

	// TextStyle returns the default style for standard text elements like Text components.
	TextStyle() Style
	// TextSelectedStyle returns the style for selected text elements (e.g., in a future List component).
	TextSelectedStyle() Style

	// GridStyle returns the style for normal, unfocused grid cells.
	GridStyle() Style
	// GridSelectedStyle returns the style for selected grid cells when the grid is not focused.
	GridSelectedStyle() Style
	// GridInteractedStyle returns the style for interacted (e.g., toggled) grid cells when the grid is not focused.
	GridInteractedStyle() Style
	// GridFocusedStyle returns the style for normal grid cells when the grid itself has input focus.
	GridFocusedStyle() Style
	// GridFocusedSelectedStyle returns the style for selected grid cells when the grid has input focus.
	GridFocusedSelectedStyle() Style
	// GridFocusedInteractedStyle returns the style for interacted grid cells when the grid has input focus.
	GridFocusedInteractedStyle() Style

	// PaneStyle returns the background style for the content area within panes (inside the border).
	PaneStyle() Style
	// PaneBorderStyle returns the style for pane borders when the pane (or its children) are not focused.
	PaneBorderStyle() Style
	// PaneFocusBorderStyle returns the style for pane borders when the pane (or its children) has input focus.
	PaneFocusBorderStyle() Style

	// --- Property Getters ---

	// DefaultCellWidth returns the theme's preferred default width for grid cells (used if Grid.autoWidth is false).
	DefaultCellWidth() int
	// DefaultCellHeight returns the theme's preferred default height for grid cells (usually 1).
	DefaultCellHeight() int
	// DefaultPadding returns the theme's preferred default internal padding within widgets like Grid cells.
	DefaultPadding() int

	// IndicatorColor returns the theme's preferred color for selection indicators (e.g., the cursor in a Grid).
	IndicatorColor() Color

	// DefaultBorderType returns the theme's preferred default border type for panes (e.g., BorderSingle, BorderDouble).
	DefaultBorderType() Border
	// FocusedBorderType returns the theme's preferred border type for panes when they (or their children) have focus.
	FocusedBorderType() Border
}

// themeManager manages the set of available themes and the currently active global theme.
// Access to the manager's state is protected by a RWMutex for thread safety.
type themeManager struct {
	current     Theme
	themes      map[ThemeName]Theme
	mu          sync.RWMutex  // Read/Write mutex for thread-safe access
	subscribers []func(Theme) // Slice of functions to call on theme change
}

var (
	// globalThemeManager holds the single, shared instance managing themes for the application process.
	globalThemeManager = &themeManager{
		themes:      make(map[ThemeName]Theme),
		subscribers: make([]func(Theme), 0),
		// current theme is set during package initialization (see base_theme.go)
	}
)

// RegisterTheme adds a new theme implementation to the manager.
// If it's the first theme registered, it automatically becomes the current global theme.
func RegisterTheme(theme Theme) {
	if theme == nil {
		return // Ignore attempts to register nil themes
	}

	globalThemeManager.mu.Lock() // Acquire write lock to modify themes map and potentially current
	defer globalThemeManager.mu.Unlock()

	name := theme.Name()
	if name == "" {
		return
	} // Ignore themes with empty names
	globalThemeManager.themes[name] = theme

	// Set as current global theme if no theme is currently set
	if globalThemeManager.current == nil {
		globalThemeManager.current = theme
	}
}

// SetTheme changes the globally active theme to the one identified by `name`.
// Returns true if the theme was found and successfully set, false otherwise.
// Notifies all registered subscribers about the theme change.
// Note: This changes the *global* theme; individual Application instances might use app.SetTheme later.
func SetTheme(name ThemeName) bool {
	globalThemeManager.mu.Lock() // Acquire write lock for changing current theme and notifying subscribers

	theme, ok := globalThemeManager.themes[name]
	if !ok {
		globalThemeManager.mu.Unlock() // Release lock before returning
		return false                   // Theme name not registered
	}

	// Only proceed if the theme is actually changing
	if globalThemeManager.current == theme {
		globalThemeManager.mu.Unlock() // Release lock before returning
		return true                    // No change needed
	}

	globalThemeManager.current = theme

	// Create a copy of the subscribers slice to safely iterate outside the lock
	subs := make([]func(Theme), len(globalThemeManager.subscribers))
	copy(subs, globalThemeManager.subscribers)

	globalThemeManager.mu.Unlock() // *** Release lock before calling subscribers ***

	// Notify subscribers about the new theme
	for _, subscriber := range subs {
		// Consider error handling or timeouts for subscriber calls? For now, call directly.
		subscriber(theme)
	}

	// No need to re-acquire lock here

	return true
}

// GetTheme returns the currently active global theme.
// It's safe for concurrent reading due to the RWMutex.
func GetTheme() Theme {
	globalThemeManager.mu.RLock() // Acquire read lock
	defer globalThemeManager.mu.RUnlock()

	// The init() function in base_theme.go ensures `current` is non-nil after package load.
	// If this were somehow called before init, it might return nil.
	return globalThemeManager.current
}

// SubscribeThemeChange registers a callback function to be executed whenever the global theme changes via SetTheme.
// The callback is also executed immediately with the current theme upon successful registration.
func SubscribeThemeChange(callback func(Theme)) {
	if callback == nil {
		return // Ignore nil callbacks
	}

	globalThemeManager.mu.Lock() // Acquire write lock to modify subscribers slice
	defer globalThemeManager.mu.Unlock()

	globalThemeManager.subscribers = append(globalThemeManager.subscribers, callback)

	// Call immediately with the current theme if one exists
	if globalThemeManager.current != nil {
		currentTheme := globalThemeManager.current
		// Temporarily release lock for the immediate callback to prevent deadlocks
		globalThemeManager.mu.Unlock()
		callback(currentTheme)
		globalThemeManager.mu.Lock() // Re-acquire lock before returning
	}
}

// --- Theme Getters (Convenience Functions) ---
// These provide easy global access to the properties of the *currently active global* theme.
// They handle the case where GetTheme() might theoretically return nil (though unlikely after init).

func DefaultTextStyle() Style {
	t := GetTheme()
	if t == nil {
		return DefaultStyle
	}
	return t.TextStyle()
}
func DefaultTextSelectedStyle() Style {
	t := GetTheme()
	if t == nil {
		return DefaultStyle.Reverse(true)
	}
	return t.TextSelectedStyle()
}
func DefaultGridStyle() Style {
	t := GetTheme()
	if t == nil {
		return DefaultStyle
	}
	return t.GridStyle()
}
func DefaultGridSelectedStyle() Style {
	t := GetTheme()
	if t == nil {
		return DefaultStyle.Reverse(true)
	}
	return t.GridSelectedStyle()
}
func DefaultGridInteractedStyle() Style {
	t := GetTheme()
	if t == nil {
		return DefaultStyle.Bold(true)
	}
	return t.GridInteractedStyle()
}
func DefaultGridFocusedStyle() Style {
	t := GetTheme()
	if t == nil {
		return DefaultStyle
	}
	return t.GridFocusedStyle()
}
func DefaultGridFocusedSelectedStyle() Style {
	t := GetTheme()
	if t == nil {
		return DefaultStyle.Reverse(true)
	}
	return t.GridFocusedSelectedStyle()
}
func DefaultGridFocusedInteractedStyle() Style {
	t := GetTheme()
	if t == nil {
		return DefaultStyle.Reverse(true).Bold(true)
	}
	return t.GridFocusedInteractedStyle()
}
func DefaultPaneStyle() Style {
	t := GetTheme()
	if t == nil {
		return DefaultStyle
	}
	return t.PaneStyle()
}
func DefaultPaneBorderStyle() Style {
	t := GetTheme()
	if t == nil {
		return DefaultStyle
	}
	return t.PaneBorderStyle()
}
func DefaultPaneFocusBorderStyle() Style {
	t := GetTheme()
	if t == nil {
		return DefaultStyle.Foreground(ColorYellow).Bold(true)
	}
	return t.PaneFocusBorderStyle()
}
func DefaultCellWidth() int {
	t := GetTheme()
	if t == nil {
		return 10
	}
	return t.DefaultCellWidth()
}
func DefaultCellHeight() int {
	t := GetTheme()
	if t == nil {
		return 1
	}
	return t.DefaultCellHeight()
}
func DefaultPadding() int {
	t := GetTheme()
	if t == nil {
		return 1
	}
	return t.DefaultPadding()
}
func DefaultIndicatorColor() Color {
	t := GetTheme()
	if t == nil {
		return ColorRed
	}
	return t.IndicatorColor()
}
func DefaultBorderType() Border {
	t := GetTheme()
	if t == nil {
		return BorderSingle
	}
	return t.DefaultBorderType()
}
func FocusedBorderType() Border {
	t := GetTheme()
	if t == nil {
		return BorderSingle
	}
	return t.FocusedBorderType()
}

// GetGridStyle is a helper function to retrieve the appropriate style for a grid cell
// based on its state (Normal, Selected, Interacted), whether the grid itself has focus,
// and the provided theme. If `theme` is nil, it uses the current global theme.
func GetGridStyle(theme Theme, state State, focused bool) Style {
	// Use global theme if a specific theme is not provided
	activeTheme := theme
	if activeTheme == nil {
		activeTheme = GetTheme()
		if activeTheme == nil { // Absolute fallback if global theme failed
			activeTheme = NewDefaultTheme()
		}
	}

	// Determine the correct style based on focus and state, using theme methods
	switch {
	case focused && state == StateInteracted:
		return activeTheme.GridFocusedInteractedStyle()
	case focused && state == StateSelected:
		return activeTheme.GridFocusedSelectedStyle()
	case focused: // Focused, normal state
		return activeTheme.GridFocusedStyle()
	case state == StateInteracted: // Unfocused, interacted state
		return activeTheme.GridInteractedStyle()
	case state == StateSelected: // Unfocused, selected state
		return activeTheme.GridSelectedStyle()
	default: // Unfocused, normal state
		return activeTheme.GridStyle()
	}
}