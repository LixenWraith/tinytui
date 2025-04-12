// themes.go
package tinytui

import "github.com/gdamore/tcell/v2"

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

// NewDefaultTheme creates the default theme
func NewDefaultTheme() Theme {
	baseStyle := DefaultStyle
	selectedStyle := baseStyle.Reverse(true)
	interactedStyle := baseStyle.Bold(true)
	focusedStyle := baseStyle.Underline(true)
	focusedSelectedStyle := baseStyle.Reverse(true).Underline(true)
	focusedInteractedStyle := baseStyle.Reverse(true).Bold(true)

	return &BaseTheme{
		name:              ThemeDefault,
		textStyle:         DefaultStyle,
		textSelectedStyle: DefaultStyle.Reverse(true),

		// Button styles
		buttonStyle:                  baseStyle,
		buttonSelectedStyle:          selectedStyle,
		buttonInteractedStyle:        interactedStyle,
		buttonFocusedStyle:           focusedStyle,
		buttonFocusedSelectedStyle:   focusedSelectedStyle,
		buttonFocusedInteractedStyle: focusedInteractedStyle,

		// List styles
		listStyle:                  baseStyle,
		listSelectedStyle:          selectedStyle,
		listInteractedStyle:        interactedStyle,
		listFocusedStyle:           focusedStyle,
		listFocusedSelectedStyle:   focusedSelectedStyle,
		listFocusedInteractedStyle: focusedInteractedStyle,

		// Grid styles
		gridStyle:                  baseStyle,
		gridSelectedStyle:          selectedStyle,
		gridInteractedStyle:        interactedStyle,
		gridFocusedStyle:           focusedStyle,
		gridFocusedSelectedStyle:   focusedSelectedStyle,
		gridFocusedInteractedStyle: focusedInteractedStyle,

		paneStyle:            DefaultStyle,
		paneBorderStyle:      DefaultStyle,
		paneFocusBorderStyle: DefaultStyle.Foreground(ColorYellow).Bold(true),
		defaultBorderType:    BorderSingle,
		defaultCellWidth:     10,
		defaultCellHeight:    1,
	}
}

// NewBorlandTheme creates a classic Borland blue theme
func NewBorlandTheme() Theme {
	// Updated Borland-style colors
	bgColor := tcell.NewRGBColor(0, 0, 170)  // Darker Blue #0000AA
	fgColor := tcell.ColorWhite              // White text
	highlightBg := tcell.ColorLightCyan      // Cyan for selection
	highlightFg := tcell.ColorBlack          // Black text on selection
	borderColor := tcell.ColorWhite          // White borders
	borderFocusColor := tcell.ColorYellow    // Yellow for focused borders
	interactedColor := tcell.ColorLightGreen // Light green for interacted items

	// Define shades of blue for different elements
	darkBlue := tcell.NewRGBColor(0, 0, 128) // #000080

	baseStyle := DefaultStyle.
		Background(bgColor).
		Foreground(fgColor)

	// Define all the state styles
	selectedStyle := DefaultStyle.Background(highlightBg).Foreground(highlightFg)
	interactedStyle := DefaultStyle.Background(bgColor).Foreground(interactedColor).Bold(true)
	focusedStyle := baseStyle.Underline(true)
	focusedSelectedStyle := selectedStyle.Underline(true)
	focusedInteractedStyle := DefaultStyle.Background(highlightBg).Foreground(interactedColor).Bold(true)

	return &BaseTheme{
		name:              ThemeBorland,
		textStyle:         baseStyle,
		textSelectedStyle: selectedStyle,

		// Button styles
		buttonStyle:                  DefaultStyle.Background(darkBlue).Foreground(fgColor),
		buttonSelectedStyle:          selectedStyle,
		buttonInteractedStyle:        interactedStyle,
		buttonFocusedStyle:           DefaultStyle.Background(darkBlue).Foreground(fgColor).Underline(true),
		buttonFocusedSelectedStyle:   focusedSelectedStyle,
		buttonFocusedInteractedStyle: focusedInteractedStyle,

		// List styles
		listStyle:                  baseStyle,
		listSelectedStyle:          selectedStyle,
		listInteractedStyle:        interactedStyle,
		listFocusedStyle:           focusedStyle,
		listFocusedSelectedStyle:   focusedSelectedStyle,
		listFocusedInteractedStyle: focusedInteractedStyle,

		// Grid styles
		gridStyle:                  baseStyle,
		gridSelectedStyle:          selectedStyle,
		gridInteractedStyle:        interactedStyle,
		gridFocusedStyle:           focusedStyle,
		gridFocusedSelectedStyle:   focusedSelectedStyle,
		gridFocusedInteractedStyle: focusedInteractedStyle,

		paneStyle:            baseStyle,
		paneBorderStyle:      baseStyle.Foreground(borderColor),
		paneFocusBorderStyle: baseStyle.Foreground(borderFocusColor).Bold(true),
		defaultBorderType:    BorderDouble,
		defaultCellWidth:     10,
		defaultCellHeight:    1,
	}
}