// theme_tokyosweet.go
package tinytui

// NewTokyoSweetTheme creates a modern theme inspired by Tokyo Night and KDE Sweet
func NewTokyoSweetTheme() Theme {
	// Tokyo Night palette with KDE Sweet accents
	const (
		// Background shades (dark to light)
		bg0 = Color(0x1a1b26) // Darkest background (e.g., main content areas, base component background)
		// bg1 = Color(0x24283b) // Slightly lighter dark background (can be used for subtle differentiation if needed)
		bg2 = Color(0x414868) // Medium dark background (e.g., unfocused selected item background)

		// Foreground shades
		fg0 = Color(0xc0caf5) // Primary text (brightest, for main content)
		fg1 = Color(0xa9b1d6) // Secondary text (e.g. for less important info, or placeholder-like text)
		fg2 = Color(0x9aa5ce) // Tertiary text / unfocused brighter elements (e.g. unfocused pane border)

		// Sweet accent colors - these are intended to be vibrant and high-contrast
		sweetPurple = Color(0xbb9af7) // Primary accent (e.g., focused selected item background)
		sweetPink   = Color(0xff007c) // KDE Sweet signature pink (e.g., focused pane border, important actions)
		sweetBlue   = Color(0x7aa2f7) // Soft blue (e.g., unfocused selected item foreground, secondary actions)
		sweetCyan   = Color(0x7dcfff) // Bright cyan (e.g., special highlights, information states)
		sweetGreen  = Color(0x9ece6a) // Success green (e.g., interacted items, positive status)
		sweetOrange = Color(0xff9e64) // Warning orange (e.g., partially different text highlight)
		sweetRed    = Color(0xf7768e) // Error/indicator red (e.g., fully different text highlight, critical status)
		trueWhite   = Color(0xffffff) // For maximum contrast text on deeply colored backgrounds
	)

	// Base style for most components: Darkest background with bright primary text.
	// This increases overall contrast and allows "Sweet" accents to pop more.
	baseStyle := DefaultStyle.Background(bg0).Foreground(fg0)

	// Text styles
	textStyle := baseStyle // General text uses the new baseStyle
	// Selected text (e.g., in a conceptual List component or for text selection highlight)
	textSelectedStyle := DefaultStyle.Background(sweetPurple).Foreground(trueWhite).Bold(true)

	// Grid styles
	gridStyle := baseStyle // Normal, unfocused cell uses the new baseStyle

	// Selected, unfocused grid cell: Subtle selection, but clear.
	gridSelectedStyle := DefaultStyle.Background(bg2).Foreground(sweetBlue).Bold(true)
	// Interacted, unfocused grid cell: Clear indication of interaction.
	gridInteractedStyle := DefaultStyle.Background(bg2).Foreground(sweetGreen).Bold(true)

	// Normal cell when grid itself has focus (consistent with unfocused normal cell).
	gridFocusedStyle := baseStyle
	// Selected cell when grid has focus: Vibrant and clear, using primary accent.
	gridFocusedSelectedStyle := DefaultStyle.Background(sweetPurple).Foreground(trueWhite).Bold(true)
	// Interacted cell when grid has focus: Strong indication, e.g., green on accent.
	gridFocusedInteractedStyle := DefaultStyle.Background(sweetGreen).Foreground(bg0).Bold(true) // Text on green

	// Pane styles
	// Pane content area: Uses the new baseStyle (darkest background for content).
	paneStyle := baseStyle
	// Unfocused pane border: Visible but not distracting, on the same dark background.
	paneBorderStyle := DefaultStyle.Background(bg0).Foreground(fg2) // fg2 is a medium-bright grey/purple
	// Focused pane border: Very vibrant and clear focus indication using signature Sweet color.
	paneFocusBorderStyle := DefaultStyle.Background(bg0).Foreground(sweetPink).Bold(true)

	// TextInput focused style is derived by reversing textStyle (fg0 becomes BG, bg0 becomes FG).
	// This will make focused TextInputs have a light purplish-blue background (fg0)
	// and very dark text (bg0), providing strong visual feedback.

	return &BaseTheme{
		name:              ThemeTokyoSweet,
		textStyle:         textStyle,
		textSelectedStyle: textSelectedStyle,

		gridStyle:                  gridStyle,
		gridSelectedStyle:          gridSelectedStyle,
		gridInteractedStyle:        gridInteractedStyle,
		gridFocusedStyle:           gridFocusedStyle,
		gridFocusedSelectedStyle:   gridFocusedSelectedStyle,
		gridFocusedInteractedStyle: gridFocusedInteractedStyle,

		paneStyle:            paneStyle,
		paneBorderStyle:      paneBorderStyle,
		paneFocusBorderStyle: paneFocusBorderStyle,

		defaultBorderType: BorderSingle, // Standard single line border
		focusedBorderType: BorderDouble, // Double line for focused panes for more emphasis

		defaultCellWidth:  12,
		defaultCellHeight: 1,
		// IndicatorColor for grids (e.g., selection cursor, or error highlight)
		// sweetRed is good for an error/difference indicator.
		// For a general selection indicator in a grid, sweetCyan or sweetOrange might also work.
		indicatorColor: sweetRed,
		defaultPadding: 1,
	}
}