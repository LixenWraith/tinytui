// themes.go
package tinytui

import "github.com/gdamore/tcell/v2"

// NewDefaultTheme creates the default theme
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

		// Add new theme properties
		indicatorColor: ColorRed, // Red indicators for default theme
		defaultPadding: 0,        // No default padding
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
	indicatorColor := tcell.ColorRed         // Red for focus indicators
	interactedColor := tcell.ColorLightGreen // Light green for interacted items

	// Define shades of blue for different elements
	darkBlue := tcell.NewRGBColor(0, 0, 128) // #000080

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
		defaultBorderType:    BorderDouble, // Change to double border for Borland theme
		defaultCellWidth:     10,
		defaultCellHeight:    1,

		// Add indicator color to the theme
		indicatorColor: indicatorColor,
		defaultPadding: 0, // Default padding should be 0
	}
}