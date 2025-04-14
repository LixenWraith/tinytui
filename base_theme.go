// base_theme.go
package tinytui

// BaseTheme provides a common implementation of the Theme interface
// to minimize duplication in concrete themes.
type BaseTheme struct {
	name              ThemeName
	textStyle         Style
	textSelectedStyle Style

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
	defaultBorderType    Border
	focusedBorderType    Border

	// Added fields for new functionality
	indicatorColor Color // Color for selection indicators
	defaultPadding int   // Default padding for widgets

	defaultCellWidth  int
	defaultCellHeight int
}

// Name returns the theme's identifier.
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

// PaneStyle returns the style for pane content areas.
func (t *BaseTheme) PaneStyle() Style {
	return t.paneStyle
}

// PaneBorderStyle returns the style for pane borders.
func (t *BaseTheme) PaneBorderStyle() Style {
	return t.paneBorderStyle
}

// PaneFocusBorderStyle returns the style for focused pane borders.
func (t *BaseTheme) PaneFocusBorderStyle() Style {
	return t.paneFocusBorderStyle
}

// DefaultCellWidth returns the default width for grid cells.
func (t *BaseTheme) DefaultCellWidth() int {
	return t.defaultCellWidth
}

// DefaultCellHeight returns the default height for grid cells.
func (t *BaseTheme) DefaultCellHeight() int {
	return t.defaultCellHeight
}

// IndicatorColor returns the color to use for selection indicators.
func (t *BaseTheme) IndicatorColor() Color {
	return t.indicatorColor
}

// DefaultPadding returns the default padding for widgets.
func (t *BaseTheme) DefaultPadding() int {
	return t.defaultPadding
}

// DefaultBorderType returns the default border type for panes.
func (t *BaseTheme) DefaultBorderType() Border {
	return t.defaultBorderType
}

// FocusedBorderType returns the border type for panes when a child has focus.
func (t *BaseTheme) FocusedBorderType() Border {
	return t.focusedBorderType
}

// NewDefaultTheme creates the default theme.
func NewDefaultTheme() Theme {
	baseStyle := DefaultStyle

	// Enhance selection visibility
	selectedStyle := baseStyle.Bold(true)
	interactedStyle := baseStyle.Foreground(ColorDarkGreen)
	focusedStyle := baseStyle
	focusedSelectedStyle := baseStyle.Background(ColorYellow).Foreground(ColorBlack).Bold(true)
	focusedInteractedStyle := baseStyle.Background(ColorGreen).Foreground(ColorBlack).Bold(true)

	return &BaseTheme{
		name:              ThemeDefault,
		textStyle:         DefaultStyle,
		textSelectedStyle: selectedStyle,

		// Grid styles
		gridStyle:                  baseStyle,
		gridSelectedStyle:          selectedStyle,
		gridInteractedStyle:        interactedStyle,
		gridFocusedStyle:           focusedStyle,
		gridFocusedSelectedStyle:   focusedSelectedStyle,
		gridFocusedInteractedStyle: focusedInteractedStyle,

		// Pane styles - only control colors, not border type
		paneStyle:            DefaultStyle,
		paneBorderStyle:      DefaultStyle,
		paneFocusBorderStyle: DefaultStyle.Foreground(ColorYellow).Bold(true),
		defaultBorderType:    BorderSingle,
		focusedBorderType:    BorderSingle,

		// Default dimensions
		defaultCellWidth:  10,
		defaultCellHeight: 1,

		// Add new theme properties
		indicatorColor: ColorRed, // Red indicators for default theme
		defaultPadding: 1,        // Add default padding for better readability
	}
}

// NewTurboTheme creates a classic blue theme.
func NewTurboTheme() Theme {
	// High contrast colors
	bgColor := ColorDarkBlue
	fgColor := ColorWhite

	// Selection colors - highlighting
	highlightBg := ColorYellow
	highlightFg := ColorBlack
	unfocusedHighlightBg := ColorDarkBlue
	unfocusedHighlightFg := ColorYellow

	// Interaction colors - buttons/toggled items
	interactedBg := ColorGreen
	interactedFg := ColorBlack
	unfocusedInteractedBg := ColorDarkBlue
	unfocusedInteractedFg := ColorGreen

	// Border and indicator colors
	borderColor := ColorWhite
	borderFocusColor := ColorYellow
	indicatorColor := ColorRed

	baseStyle := DefaultStyle.
		Background(bgColor).
		Foreground(fgColor)

	// Define all the state styles with enhanced contrast and visibility
	selectedStyle := DefaultStyle.Background(unfocusedHighlightBg).Foreground(unfocusedHighlightFg).Bold(true)
	interactedStyle := DefaultStyle.Background(unfocusedInteractedBg).Foreground(unfocusedInteractedFg).Bold(true)
	focusedStyle := baseStyle
	focusedSelectedStyle := DefaultStyle.Background(highlightBg).Foreground(highlightFg)
	focusedInteractedStyle := DefaultStyle.Background(interactedBg).Foreground(interactedFg)

	return &BaseTheme{
		name:              ThemeTurbo,
		textStyle:         baseStyle,
		textSelectedStyle: selectedStyle,

		// Grid styles
		gridStyle:                  baseStyle,
		gridSelectedStyle:          selectedStyle,
		gridInteractedStyle:        interactedStyle,
		gridFocusedStyle:           focusedStyle,
		gridFocusedSelectedStyle:   focusedSelectedStyle,
		gridFocusedInteractedStyle: focusedInteractedStyle,

		// Pane styles - only control colors, not border type
		paneStyle:            baseStyle,
		paneBorderStyle:      baseStyle.Foreground(borderColor),
		paneFocusBorderStyle: baseStyle.Foreground(borderFocusColor).Bold(true),
		defaultBorderType:    BorderSingle,
		focusedBorderType:    BorderDouble,

		// Default dimensions
		defaultCellWidth:  10,
		defaultCellHeight: 1,

		// Add indicator color to the theme
		indicatorColor: indicatorColor,
		defaultPadding: 1, // Increase default padding for better readability
	}
}

// Initialize themes in init function
func init() {
	// Register default themes
	RegisterTheme(NewDefaultTheme())
	RegisterTheme(NewTurboTheme())

	// Set default theme
	SetTheme(ThemeDefault)
}
