// base_theme.go
package tinytui

// BaseTheme provides a common implementation foundation for the Theme interface,
// reducing boilerplate code in concrete theme definitions.
type BaseTheme struct {
	name              ThemeName // Unique identifier (e.g., "default", "turbo")
	textStyle         Style     // Default text style
	textSelectedStyle Style     // Style for selected text (e.g., in a future List component)

	// Grid styles for various states
	gridStyle                  Style // Normal, unfocused cell
	gridSelectedStyle          Style // Selected, unfocused cell
	gridInteractedStyle        Style // Interacted (e.g., toggled), unfocused cell
	gridFocusedStyle           Style // Normal cell when grid itself has focus
	gridFocusedSelectedStyle   Style // Selected cell when grid has focus
	gridFocusedInteractedStyle Style // Interacted cell when grid has focus

	// Pane styles
	paneStyle            Style  // Background style for the pane's content area
	paneBorderStyle      Style  // Style for the pane's border when unfocused
	paneFocusBorderStyle Style  // Style for the pane's border when focused (or child focused)
	defaultBorderType    Border // Default border type (e.g., Single, Double) for unfocused panes
	focusedBorderType    Border // Border type to use when the pane (or a child) is focused

	// Other theme attributes
	indicatorColor    Color // Color for indicators (e.g., selection cursor in Grid)
	defaultPadding    int   // Default padding within widgets like Grid cells
	defaultCellWidth  int   // Default width for Grid cells (if not auto-sized)
	defaultCellHeight int   // Default height for Grid cells
}

// Name returns the theme's identifier.
func (t *BaseTheme) Name() ThemeName {
	return t.name
}

// TextStyle returns the default style for text elements.
func (t *BaseTheme) TextStyle() Style {
	return t.textStyle
}

// TextSelectedStyle returns the style for selected text elements.
func (t *BaseTheme) TextSelectedStyle() Style {
	return t.textSelectedStyle
}

// GridStyle returns the style for normal, unfocused grid cells.
func (t *BaseTheme) GridStyle() Style {
	return t.gridStyle
}

// GridSelectedStyle returns the style for selected, unfocused grid cells.
func (t *BaseTheme) GridSelectedStyle() Style {
	return t.gridSelectedStyle
}

// GridInteractedStyle returns the style for interacted, unfocused grid cells.
func (t *BaseTheme) GridInteractedStyle() Style {
	return t.gridInteractedStyle
}

// GridFocusedStyle returns the style for normal grid cells when the grid has focus.
func (t *BaseTheme) GridFocusedStyle() Style {
	return t.gridFocusedStyle
}

// GridFocusedSelectedStyle returns the style for selected grid cells when the grid has focus.
func (t *BaseTheme) GridFocusedSelectedStyle() Style {
	return t.gridFocusedSelectedStyle
}

// GridFocusedInteractedStyle returns the style for interacted grid cells when the grid has focus.
func (t *BaseTheme) GridFocusedInteractedStyle() Style {
	return t.gridFocusedInteractedStyle
}

// PaneStyle returns the style for pane content areas (background).
func (t *BaseTheme) PaneStyle() Style {
	return t.paneStyle
}

// PaneBorderStyle returns the style for unfocused pane borders.
func (t *BaseTheme) PaneBorderStyle() Style {
	return t.paneBorderStyle
}

// PaneFocusBorderStyle returns the style for focused pane borders.
func (t *BaseTheme) PaneFocusBorderStyle() Style {
	return t.paneFocusBorderStyle
}

// DefaultCellWidth returns the theme's preferred default width for grid cells.
func (t *BaseTheme) DefaultCellWidth() int {
	return t.defaultCellWidth
}

// DefaultCellHeight returns the theme's preferred default height for grid cells.
func (t *BaseTheme) DefaultCellHeight() int {
	return t.defaultCellHeight
}

// IndicatorColor returns the theme's preferred color for selection indicators.
func (t *BaseTheme) IndicatorColor() Color {
	return t.indicatorColor
}

// DefaultPadding returns the theme's preferred default padding for widgets.
func (t *BaseTheme) DefaultPadding() int {
	return t.defaultPadding
}

// DefaultBorderType returns the theme's preferred default border type for panes.
func (t *BaseTheme) DefaultBorderType() Border {
	return t.defaultBorderType
}

// FocusedBorderType returns the theme's preferred border type for focused panes.
func (t *BaseTheme) FocusedBorderType() Border {
	return t.focusedBorderType
}

// --- Concrete Theme Definitions ---

// NewDefaultTheme creates the default light-background theme.
func NewDefaultTheme() Theme {
	baseStyle := DefaultStyle // Assumes DefaultStyle is Reset (fg/bg default)

	// Define styles for different states
	selectedStyle := baseStyle.Bold(true)                                                        // Simple bold for unfocused selection
	interactedStyle := baseStyle.Foreground(ColorDarkGreen)                                      // Green text for unfocused interaction
	focusedStyle := baseStyle                                                                    // No change for base cell when grid focused
	focusedSelectedStyle := baseStyle.Background(ColorYellow).Foreground(ColorBlack).Bold(true)  // High contrast selection when focused
	focusedInteractedStyle := baseStyle.Background(ColorGreen).Foreground(ColorBlack).Bold(true) // High contrast interaction when focused

	return &BaseTheme{
		name:                       ThemeDefault,
		textStyle:                  baseStyle,
		textSelectedStyle:          selectedStyle.Reverse(true), // Use reverse video for selected text areas
		gridStyle:                  baseStyle,
		gridSelectedStyle:          selectedStyle,
		gridInteractedStyle:        interactedStyle,
		gridFocusedStyle:           focusedStyle, // Focused grid uses base style for normal cells
		gridFocusedSelectedStyle:   focusedSelectedStyle,
		gridFocusedInteractedStyle: focusedInteractedStyle,
		paneStyle:                  baseStyle,                                    // Pane background is default terminal bg
		paneBorderStyle:            baseStyle,                                    // Pane border uses default terminal fg/bg
		paneFocusBorderStyle:       baseStyle.Foreground(ColorYellow).Bold(true), // Focused border is yellow and bold
		defaultBorderType:          BorderSingle,
		focusedBorderType:          BorderSingle, // Focus doesn't change border type in default theme
		defaultCellWidth:           10,
		defaultCellHeight:          1,
		indicatorColor:             ColorRed, // Selection indicator is red
		defaultPadding:             1,        // 1 cell padding in grids
	}
}

// NewTurboTheme creates a theme inspired by classic Turbo Vision (blue background).
func NewTurboTheme() Theme {
	// Base colors
	bgColor := ColorDarkBlue
	fgColor := ColorWhite
	baseStyle := DefaultStyle.Background(bgColor).Foreground(fgColor)

	// Selection colors (more distinct than default theme)
	highlightBg := ColorLightCyan       // Use Cyan for focused selection BG
	highlightFg := ColorBlack           // Black text on Cyan
	unfocusedHighlightBg := bgColor     // Keep background same when unfocused
	unfocusedHighlightFg := ColorYellow // Yellow text for unfocused selection

	// Interaction colors (e.g., toggled buttons)
	interactedBg := ColorGreen               // Use Green for focused interaction BG
	interactedFg := ColorWhite               // White text on Green
	unfocusedInteractedBg := bgColor         // Keep background same when unfocused
	unfocusedInteractedFg := ColorLightGreen // Light green text for unfocused interaction

	// Border colors
	borderColor := ColorSilver      // Use Silver for normal borders
	borderFocusColor := ColorYellow // Use Yellow for focused borders

	// Define state styles based on colors
	selectedStyle := DefaultStyle.Background(unfocusedHighlightBg).Foreground(unfocusedHighlightFg).Bold(true)
	interactedStyle := DefaultStyle.Background(unfocusedInteractedBg).Foreground(unfocusedInteractedFg).Bold(true)
	focusedStyle := baseStyle // Base style when grid is focused but cell is normal
	focusedSelectedStyle := DefaultStyle.Background(highlightBg).Foreground(highlightFg).Bold(true)
	focusedInteractedStyle := DefaultStyle.Background(interactedBg).Foreground(interactedFg).Bold(true)

	return &BaseTheme{
		name:                       ThemeTurbo,
		textStyle:                  baseStyle,
		textSelectedStyle:          selectedStyle.Reverse(true), // Use reverse of the unfocused selected style for text areas
		gridStyle:                  baseStyle,
		gridSelectedStyle:          selectedStyle,
		gridInteractedStyle:        interactedStyle,
		gridFocusedStyle:           focusedStyle,
		gridFocusedSelectedStyle:   focusedSelectedStyle,
		gridFocusedInteractedStyle: focusedInteractedStyle,
		paneStyle:                  baseStyle,                                         // Pane background uses theme base
		paneBorderStyle:            baseStyle.Foreground(borderColor),                 // Use theme bg, specific border fg
		paneFocusBorderStyle:       baseStyle.Foreground(borderFocusColor).Bold(true), // Use theme bg, specific focus border fg + bold
		defaultBorderType:          BorderSingle,                                      // Default to single border
		focusedBorderType:          BorderDouble,                                      // Use double border when focused
		defaultCellWidth:           10,
		defaultCellHeight:          1,
		indicatorColor:             ColorRed, // Keep indicator red for high visibility
		defaultPadding:             1,        // Keep 1 cell padding
	}
}

// Initialize and register themes when the package loads.
// This ensures themes are available before NewApplication is called.
func init() {
	// Register default themes
	RegisterTheme(NewDefaultTheme())
	RegisterTheme(NewTurboTheme())

	// Set the default global theme (can be overridden by application via SetTheme)
	// SetTheme uses the global theme manager's mutex internally.
	SetTheme(ThemeDefault)
}