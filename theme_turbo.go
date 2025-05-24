// theme_turbo.go
package tinytui

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