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

// Initialize and register themes when the package loads.
// This ensures themes are available before NewApplication is called.
func init() {
	// Register bundled themes
	RegisterTheme(NewDefaultTheme())
	RegisterTheme(NewTurboTheme())
	RegisterTheme(NewTokyoSweetTheme())

	// Set the default global theme (can be overridden by application via SetTheme)
	// SetTheme uses the global theme manager's mutex internally.
	SetTheme(ThemeDefault)
}