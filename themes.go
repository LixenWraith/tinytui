// themes.go
package tinytui

import "github.com/gdamore/tcell/v2"

func NewDefaultTheme() Theme {
	baseStyle := DefaultStyle
	selectedStyle := baseStyle.Reverse(true)
	interactedStyle := baseStyle.Bold(true)
	focusedStyle := baseStyle
	focusedSelectedStyle := baseStyle.Reverse(true).Underline(true)
	focusedInteractedStyle := baseStyle.Reverse(true).Bold(true)

	return &BaseTheme{
		name:              ThemeDefault,
		textStyle:         DefaultStyle,
		textSelectedStyle: DefaultStyle.Reverse(true),

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

		// Default dimensions
		defaultCellWidth:  10,
		defaultCellHeight: 1,

		// Add new theme properties
		indicatorColor: ColorRed, // Red indicators for default theme
		defaultPadding: 0,        // No default padding
	}
}

// NewBorlandTheme creates a classic Borland blue theme
func NewBorlandTheme() Theme {
	// Updated Borland-style colors
	bgColor := tcell.ColorDarkBlue           // Darker Blue #0000AA
	fgColor := tcell.ColorWhite              // White text
	highlightBg := tcell.ColorLightCyan      // Cyan for selection
	highlightFg := tcell.ColorBlack          // Black text on selection
	borderColor := tcell.ColorWhite          // White borders
	borderFocusColor := tcell.ColorYellow    // Yellow for focused borders
	indicatorColor := tcell.ColorRed         // Red for focus indicators
	interactedColor := tcell.ColorLightGreen // Light green for interacted items

	baseStyle := DefaultStyle.
		Background(bgColor).
		Foreground(fgColor)

	// Define all the state styles
	selectedStyle := DefaultStyle.Background(highlightBg).Foreground(highlightFg)
	interactedStyle := DefaultStyle.Background(bgColor).Foreground(interactedColor).Bold(true)
	focusedStyle := baseStyle
	focusedSelectedStyle := selectedStyle.Underline(true)
	focusedInteractedStyle := DefaultStyle.Background(highlightBg).Foreground(interactedColor).Bold(true)

	return &BaseTheme{
		name:              ThemeBorland,
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

		// Default dimensions
		defaultCellWidth:  10,
		defaultCellHeight: 1,

		// Add indicator color to the theme
		indicatorColor: indicatorColor,
		defaultPadding: 0, // Default padding should be 0
	}
}