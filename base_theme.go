// themes.go
package tinytui

// BaseTheme provides a common implementation of the Theme interface
// to minimize duplication in concrete themes
type BaseTheme struct {
	name              ThemeName
	textStyle         Style
	textSelectedStyle Style

	// Button styles
	buttonStyle                  Style
	buttonSelectedStyle          Style
	buttonInteractedStyle        Style
	buttonFocusedStyle           Style
	buttonFocusedSelectedStyle   Style
	buttonFocusedInteractedStyle Style

	// List styles
	listStyle                  Style
	listSelectedStyle          Style
	listInteractedStyle        Style
	listFocusedStyle           Style
	listFocusedSelectedStyle   Style
	listFocusedInteractedStyle Style

	// Grid styles
	gridStyle                  Style
	gridSelectedStyle          Style
	gridInteractedStyle        Style
	gridFocusedStyle           Style
	gridFocusedSelectedStyle   Style
	gridFocusedInteractedStyle Style

	// Pane styles
	paneStyle            Style
	paneBorderStyle      Style
	paneFocusBorderStyle Style

	// Added fields for new functionality
	indicatorColor Color // Color for selection indicators
	defaultPadding int   // Default padding for widgets

	defaultBorderType BorderType
	defaultCellWidth  int
	defaultCellHeight int
}

// Name returns the theme's identifier
func (t *BaseTheme) Name() ThemeName {
	return t.name
}

// Text style methods
func (t *BaseTheme) TextStyle() Style {
	return t.textStyle
}

func (t *BaseTheme) TextSelectedStyle() Style {
	return t.textSelectedStyle
}

// Button style methods
func (t *BaseTheme) ButtonStyle() Style {
	return t.buttonStyle
}

func (t *BaseTheme) ButtonSelectedStyle() Style {
	return t.buttonSelectedStyle
}

func (t *BaseTheme) ButtonInteractedStyle() Style {
	return t.buttonInteractedStyle
}

func (t *BaseTheme) ButtonFocusedStyle() Style {
	return t.buttonFocusedStyle
}

func (t *BaseTheme) ButtonFocusedSelectedStyle() Style {
	return t.buttonFocusedSelectedStyle
}

func (t *BaseTheme) ButtonFocusedInteractedStyle() Style {
	return t.buttonFocusedInteractedStyle
}

// List style methods
func (t *BaseTheme) ListStyle() Style {
	return t.listStyle
}

func (t *BaseTheme) ListSelectedStyle() Style {
	return t.listSelectedStyle
}

func (t *BaseTheme) ListInteractedStyle() Style {
	return t.listInteractedStyle
}

func (t *BaseTheme) ListFocusedStyle() Style {
	return t.listFocusedStyle
}

func (t *BaseTheme) ListFocusedSelectedStyle() Style {
	return t.listFocusedSelectedStyle
}

func (t *BaseTheme) ListFocusedInteractedStyle() Style {
	return t.listFocusedInteractedStyle
}

// Grid style methods
func (t *BaseTheme) GridStyle() Style {
	return t.gridStyle
}

func (t *BaseTheme) GridSelectedStyle() Style {
	return t.gridSelectedStyle
}

func (t *BaseTheme) GridInteractedStyle() Style {
	return t.gridInteractedStyle
}

func (t *BaseTheme) GridFocusedStyle() Style {
	return t.gridFocusedStyle
}

func (t *BaseTheme) GridFocusedSelectedStyle() Style {
	return t.gridFocusedSelectedStyle
}

func (t *BaseTheme) GridFocusedInteractedStyle() Style {
	return t.gridFocusedInteractedStyle
}

// PaneStyle returns the style for pane content areas
func (t *BaseTheme) PaneStyle() Style {
	return t.paneStyle
}

// PaneBorderStyle returns the style for pane borders
func (t *BaseTheme) PaneBorderStyle() Style {
	return t.paneBorderStyle
}

// PaneFocusBorderStyle returns the style for focused pane borders
func (t *BaseTheme) PaneFocusBorderStyle() Style {
	return t.paneFocusBorderStyle
}

// DefaultBorderType returns the default border type for panes
func (t *BaseTheme) DefaultBorderType() BorderType {
	return t.defaultBorderType
}

// DefaultCellWidth returns the default width for grid cells
func (t *BaseTheme) DefaultCellWidth() int {
	return t.defaultCellWidth
}

// DefaultCellHeight returns the default height for grid cells
func (t *BaseTheme) DefaultCellHeight() int {
	return t.defaultCellHeight
}

// IndicatorColor returns the color to use for selection indicators
func (t *BaseTheme) IndicatorColor() Color {
	return t.indicatorColor
}

// DefaultPadding returns the default padding for widgets
func (t *BaseTheme) DefaultPadding() int {
	return t.defaultPadding
}