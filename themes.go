// themes.go
package tinytui

import "github.com/gdamore/tcell/v2"

func NewDefaultTheme() Theme {
	baseStyle := DefaultStyle

	// Enhance selection visibility
	selectedStyle := baseStyle.Bold(true)
	interactedStyle := baseStyle.Foreground(tcell.ColorDarkGreen)
	focusedStyle := baseStyle
	focusedSelectedStyle := baseStyle.Background(tcell.ColorYellow).Foreground(tcell.ColorBlack).Bold(true)
	focusedInteractedStyle := baseStyle.Background(tcell.ColorGreen).Foreground(tcell.ColorBlack).Bold(true)

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

// NewTurboTheme creates a classic blue theme
func NewTurboTheme() Theme {
	// High contrast colors
	bgColor := tcell.ColorDarkBlue
	fgColor := tcell.ColorWhite

	// Selection colors - highlighting
	highlightBg := tcell.ColorYellow
	highlightFg := tcell.ColorBlack
	unfocusedHighlightBg := tcell.ColorDarkBlue
	unfocusedHighlightFg := tcell.ColorYellow

	// Interaction colors - buttons/toggled items
	interactedBg := tcell.ColorGreen
	interactedFg := tcell.ColorBlack
	unfocusedInteractedBg := tcell.ColorDarkBlue
	unfocusedInteractedFg := tcell.ColorGreen

	// Border and indicator colors
	borderColor := tcell.ColorWhite
	borderFocusColor := tcell.ColorYellow
	indicatorColor := tcell.ColorRed

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