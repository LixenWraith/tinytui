// themes.go
package tinytui

import "github.com/gdamore/tcell/v2"

// BaseTheme provides a common implementation of the Theme interface
// to minimize duplication in concrete themes
type BaseTheme struct {
	name                 ThemeName
	textStyle            Style
	textSelectedStyle    Style
	buttonStyle          Style
	buttonFocusedStyle   Style
	listStyle            Style
	listSelectedStyle    Style
	gridStyle            Style
	gridSelectedStyle    Style
	paneStyle            Style
	paneBorderStyle      Style
	paneFocusBorderStyle Style
	defaultBorderType    BorderType
	defaultCellWidth     int
	defaultCellHeight    int
}

// Name returns the theme's identifier
func (t *BaseTheme) Name() ThemeName {
	return t.name
}

// TextStyle returns the default style for text widgets
func (t *BaseTheme) TextStyle() Style {
	return t.textStyle
}

// TextSelectedStyle returns the style for selected text
func (t *BaseTheme) TextSelectedStyle() Style {
	return t.textSelectedStyle
}

// ButtonStyle returns the style for buttons
func (t *BaseTheme) ButtonStyle() Style {
	return t.buttonStyle
}

// ButtonFocusedStyle returns the style for focused buttons
func (t *BaseTheme) ButtonFocusedStyle() Style {
	return t.buttonFocusedStyle
}

// ListStyle returns the style for list items
func (t *BaseTheme) ListStyle() Style {
	return t.listStyle
}

// ListSelectedStyle returns the style for selected list items
func (t *BaseTheme) ListSelectedStyle() Style {
	return t.listSelectedStyle
}

// GridStyle returns the style for grid cells
func (t *BaseTheme) GridStyle() Style {
	return t.gridStyle
}

// GridSelectedStyle returns the style for selected grid cells
func (t *BaseTheme) GridSelectedStyle() Style {
	return t.gridSelectedStyle
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
	return &BaseTheme{
		name:                 ThemeDefault,
		textStyle:            DefaultStyle,
		textSelectedStyle:    DefaultStyle.Reverse(true),
		buttonStyle:          DefaultStyle,
		buttonFocusedStyle:   DefaultStyle.Reverse(true),
		listStyle:            DefaultStyle,
		listSelectedStyle:    DefaultStyle.Reverse(true),
		gridStyle:            DefaultStyle,
		gridSelectedStyle:    DefaultStyle.Reverse(true),
		paneStyle:            DefaultStyle,
		paneBorderStyle:      DefaultStyle,
		paneFocusBorderStyle: DefaultStyle.Foreground(ColorYellow).Bold(true),
		defaultBorderType:    BorderSingle,
		defaultCellWidth:     10,
		defaultCellHeight:    1,
	}
}

// NewTokyoNightTheme creates a theme inspired by Tokyo Night VSCode theme
func NewTokyoNightTheme() Theme {
	// Tokyo Night palette
	bgColor := tcell.NewRGBColor(26, 27, 38)         // #1a1b26
	bgAltColor := tcell.NewRGBColor(36, 40, 59)      // #24283b
	fgColor := tcell.NewRGBColor(169, 177, 214)      // #a9b1d6
	accentColor := tcell.NewRGBColor(125, 207, 255)  // #7dcfff
	selectionColor := tcell.NewRGBColor(73, 82, 115) // #495273
	// highlightColor := tcell.NewRGBColor(187, 154, 247) // #bb9af7
	// errorColor := tcell.NewRGBColor(247, 118, 142) // #f7768e

	baseStyle := DefaultStyle.
		Background(bgColor).
		Foreground(fgColor)

	return &BaseTheme{
		name:                 ThemeTokyoNight,
		textStyle:            baseStyle,
		textSelectedStyle:    baseStyle.Background(selectionColor),
		buttonStyle:          baseStyle.Background(bgAltColor),
		buttonFocusedStyle:   baseStyle.Background(selectionColor).Foreground(accentColor),
		listStyle:            baseStyle,
		listSelectedStyle:    baseStyle.Background(selectionColor),
		gridStyle:            baseStyle,
		gridSelectedStyle:    baseStyle.Background(selectionColor),
		paneStyle:            baseStyle,
		paneBorderStyle:      baseStyle.Foreground(fgColor),
		paneFocusBorderStyle: baseStyle.Foreground(accentColor).Bold(true),
		defaultBorderType:    BorderSingle,
		defaultCellWidth:     10,
		defaultCellHeight:    1,
	}
}

// NewCatppuccinMochaTheme creates a theme inspired by Catppuccin Mocha
func NewCatppuccinMochaTheme() Theme {
	// Catppuccin Mocha palette
	bgColor := tcell.NewRGBColor(30, 30, 46)     // #1e1e2e
	mantle := tcell.NewRGBColor(24, 24, 37)      // #181825
	crust := tcell.NewRGBColor(17, 17, 27)       // #11111b
	text := tcell.NewRGBColor(205, 214, 244)     // #cdd6f4
	subtext0 := tcell.NewRGBColor(166, 173, 200) // #a6adc8
	lavender := tcell.NewRGBColor(180, 190, 254) // #b4befe
	blue := tcell.NewRGBColor(137, 180, 250)     // #89b4fa
	// sapphire := tcell.NewRGBColor(116, 199, 236) // #74c7ec
	// pink := tcell.NewRGBColor(245, 194, 231)     // #f5c2e7

	baseStyle := DefaultStyle.
		Background(bgColor).
		Foreground(text)

	return &BaseTheme{
		name:                 ThemeCatppuccinMocha,
		textStyle:            baseStyle,
		textSelectedStyle:    baseStyle.Background(mantle).Foreground(lavender),
		buttonStyle:          baseStyle.Background(mantle),
		buttonFocusedStyle:   baseStyle.Background(crust).Foreground(lavender),
		listStyle:            baseStyle,
		listSelectedStyle:    baseStyle.Background(mantle).Foreground(lavender),
		gridStyle:            baseStyle,
		gridSelectedStyle:    baseStyle.Background(mantle).Foreground(lavender),
		paneStyle:            baseStyle,
		paneBorderStyle:      baseStyle.Foreground(subtext0),
		paneFocusBorderStyle: baseStyle.Foreground(blue).Bold(true),
		defaultBorderType:    BorderSingle,
		defaultCellWidth:     10,
		defaultCellHeight:    1,
	}
}

// NewBorlandTheme creates a classic Borland blue theme
func NewBorlandTheme() Theme {
	// Borland-style colors
	bgColor := ColorBlue
	fgColor := ColorWhite
	highlightBg := ColorCyan
	highlightFg := ColorBlack
	borderColor := ColorWhite
	borderFocusColor := ColorYellow

	baseStyle := DefaultStyle.
		Background(bgColor).
		Foreground(fgColor)

	return &BaseTheme{
		name:                 ThemeBorland,
		textStyle:            baseStyle,
		textSelectedStyle:    DefaultStyle.Background(highlightBg).Foreground(highlightFg),
		buttonStyle:          baseStyle,
		buttonFocusedStyle:   DefaultStyle.Background(highlightBg).Foreground(highlightFg),
		listStyle:            baseStyle,
		listSelectedStyle:    DefaultStyle.Background(highlightBg).Foreground(highlightFg),
		gridStyle:            baseStyle,
		gridSelectedStyle:    DefaultStyle.Background(highlightBg).Foreground(highlightFg),
		paneStyle:            baseStyle,
		paneBorderStyle:      baseStyle.Foreground(borderColor),
		paneFocusBorderStyle: baseStyle.Foreground(borderFocusColor).Bold(true),
		defaultBorderType:    BorderDouble, // Classic DOS-style double borders
		defaultCellWidth:     10,
		defaultCellHeight:    1,
	}
}