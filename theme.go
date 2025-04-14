// theme.go
package tinytui

// ThemeName represents a predefined theme identifier.
type ThemeName string

const (
	// ThemeDefault is the fallback theme
	ThemeDefault ThemeName = "default"
	// ThemeTurbo is loosely inspired by turbo vision color scheme
	ThemeTurbo ThemeName = "turbo"
)

// Theme defines the interface for a UI theme.
type Theme interface {
	// Name returns the unique name of the theme
	Name() ThemeName

	// TextStyles
	TextStyle() Style
	TextSelectedStyle() Style

	// Grid styles based on state
	GridStyle() Style                  // Base style (equivalent to GridNormalStyle)
	GridFocusedStyle() Style           // Base focused style
	GridSelectedStyle() Style          // Selected, not focused
	GridInteractedStyle() Style        // Interacted, not focused
	GridFocusedSelectedStyle() Style   // Selected and focused
	GridFocusedInteractedStyle() Style // Interacted and focused

	// Pane styles
	PaneStyle() Style            // Style for pane content area
	PaneBorderStyle() Style      // Style for pane border (unfocused)
	PaneFocusBorderStyle() Style // Style for pane border when a child has focus

	// Default cell dimensions for grid
	DefaultCellWidth() int
	DefaultCellHeight() int

	// Indicator and padding
	IndicatorColor() Color // Color for selection indicators
	DefaultPadding() int   // Default padding for widgets

	// Border types
	DefaultBorderType() Border // Default border type for unfocused panes
	FocusedBorderType() Border // Border type for panes when a child has focus
}

// themeManager handles the current theme and theme switching.
type themeManager struct {
	current     Theme
	themes      map[ThemeName]Theme
	subscribers []func(Theme)
}

var (
	// Global theme manager instance
	globalThemeManager = &themeManager{
		themes:      make(map[ThemeName]Theme),
		subscribers: make([]func(Theme), 0),
	}
)

// RegisterTheme adds a theme to the available themes.
func RegisterTheme(theme Theme) {
	if theme == nil {
		return
	}

	name := theme.Name()
	globalThemeManager.themes[name] = theme

	// If this is the first theme registered, set it as current
	if globalThemeManager.current == nil {
		globalThemeManager.current = theme
	}
}

// SetTheme changes the current theme.
func SetTheme(name ThemeName) bool {
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

// GetTheme returns the current theme.
func GetTheme() Theme {
	return globalThemeManager.current
}

// SubscribeThemeChange registers a function to be called when the theme changes.
func SubscribeThemeChange(callback func(Theme)) {
	if callback == nil {
		return
	}

	globalThemeManager.subscribers = append(globalThemeManager.subscribers, callback)

	// Call immediately with current theme
	if globalThemeManager.current != nil {
		callback(globalThemeManager.current)
	}
}

// Default style getters
// These functions provide easy access to the current theme's styles.

// DefaultTextStyle returns the current theme's text style.
func DefaultTextStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.TextStyle()
	}
	return DefaultStyle
}

// DefaultTextSelectedStyle returns the current theme's selected text style.
func DefaultTextSelectedStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.TextSelectedStyle()
	}
	return DefaultStyle.Reverse(true)
}

// DefaultGridStyle returns the current theme's grid style.
func DefaultGridStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.GridStyle()
	}
	return DefaultStyle
}

// DefaultGridSelectedStyle returns the current theme's selected grid cell style.
func DefaultGridSelectedStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.GridSelectedStyle()
	}
	return DefaultStyle.Reverse(true)
}

// DefaultGridInteractedStyle returns the current theme's interacted grid cell style.
func DefaultGridInteractedStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.GridInteractedStyle()
	}
	return DefaultStyle.Bold(true)
}

// DefaultGridFocusedStyle returns the current theme's focused grid style.
func DefaultGridFocusedStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.GridFocusedStyle()
	}
	return DefaultStyle
}

// DefaultGridFocusedSelectedStyle returns the current theme's focused and selected grid cell style.
func DefaultGridFocusedSelectedStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.GridFocusedSelectedStyle()
	}
	return DefaultStyle.Reverse(true)
}

// DefaultGridFocusedInteractedStyle returns the current theme's focused and interacted grid cell style.
func DefaultGridFocusedInteractedStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.GridFocusedInteractedStyle()
	}
	return DefaultStyle.Reverse(true).Bold(true)
}

// DefaultPaneStyle returns the current theme's pane style.
func DefaultPaneStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.PaneStyle()
	}
	return DefaultStyle
}

// DefaultPaneBorderStyle returns the current theme's pane border style.
func DefaultPaneBorderStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.PaneBorderStyle()
	}
	return DefaultStyle
}

// DefaultPaneFocusBorderStyle returns the current theme's focused pane border style.
func DefaultPaneFocusBorderStyle() Style {
	theme := GetTheme()
	if theme != nil {
		return theme.PaneFocusBorderStyle()
	}
	return DefaultStyle.Foreground(ColorYellow).Bold(true)
}

// DefaultCellWidth returns the current theme's default cell width.
func DefaultCellWidth() int {
	theme := GetTheme()
	if theme != nil {
		return theme.DefaultCellWidth()
	}
	return 10 // Default value
}

// DefaultCellHeight returns the current theme's default cell height.
func DefaultCellHeight() int {
	theme := GetTheme()
	if theme != nil {
		return theme.DefaultCellHeight()
	}
	return 1 // Default value
}

// DefaultPadding returns the current theme's default padding.
func DefaultPadding() int {
	theme := GetTheme()
	if theme != nil {
		return theme.DefaultPadding()
	}
	return 1 // Default value
}

// DefaultIndicatorColor returns the current theme's indicator color.
func DefaultIndicatorColor() Color {
	theme := GetTheme()
	if theme != nil {
		return theme.IndicatorColor()
	}
	return ColorRed // Default value
}

// DefaultBorderType returns the current theme's default border type.
func DefaultBorderType() Border {
	theme := GetTheme()
	if theme != nil {
		return theme.DefaultBorderType()
	}
	return BorderSingle // Default fallback
}

// FocusedBorderType returns the current theme's focused border type.
func FocusedBorderType() Border {
	theme := GetTheme()
	if theme != nil {
		return theme.FocusedBorderType()
	}
	return BorderSingle // Default fallback
}

// GetGridStyle returns the appropriate style for a grid widget based on its state and focus.
func GetGridStyle(theme Theme, state State, focused bool) Style {
	if theme == nil {
		theme = GetTheme()
		if theme == nil {
			// Fall back to default styles
			switch {
			case focused && state == StateInteracted:
				return DefaultStyle.Reverse(true).Bold(true)
			case focused && state == StateSelected:
				return DefaultStyle.Reverse(true)
			case focused:
				return DefaultStyle
			case state == StateInteracted:
				return DefaultStyle.Bold(true)
			case state == StateSelected:
				return DefaultStyle.Dim(true).Underline(true)
			default:
				return DefaultStyle
			}
		}
	}

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

// ThemedComponent is an interface for components that need custom theme handling.
// Components implementing this interface can respond to theme changes with
// specialized behavior beyond the default styles.
//
// Note: This is currently a forward-looking interface that will enable more
// sophisticated theming in future versions. Its implementation is minimal in
// the current version but provides a foundation for component-specific theme
// handling.
type ThemedComponent interface {
	Component
	ApplyTheme(theme Theme)
}