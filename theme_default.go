// theme_default.go
package tinytui

// NewDefaultTheme creates the default theme.
func NewDefaultTheme() Theme {
	const (
		defaultBg     = ColorBlack
		defaultBorder = ColorGray
		defaultFg     = ColorWhite
		defaultAccent = ColorBlue
		defaultSelect = ColorLightCyan
		defaultYes    = ColorGreen
		defaultNo     = ColorRed
	)

	baseStyle := DefaultStyle.Background(defaultBg).Foreground(defaultFg)

	return &BaseTheme{
		name:                       ThemeDefault,
		textStyle:                  baseStyle,
		textSelectedStyle:          baseStyle.Background(defaultSelect).Foreground(defaultNo).Bold(true),
		gridStyle:                  baseStyle,
		gridSelectedStyle:          baseStyle.Foreground(defaultAccent).Bold(true),
		gridInteractedStyle:        baseStyle.Foreground(defaultYes).Bold(true),
		gridFocusedStyle:           baseStyle,
		gridFocusedSelectedStyle:   baseStyle.Background(defaultSelect).Foreground(defaultBorder).Bold(true),
		gridFocusedInteractedStyle: baseStyle.Background(defaultYes).Foreground(defaultBg).Bold(true),
		paneStyle:                  baseStyle,
		paneBorderStyle:            baseStyle.Foreground(defaultBorder),
		paneFocusBorderStyle:       baseStyle.Foreground(defaultAccent).Bold(true),
		defaultBorderType:          BorderSingle,
		focusedBorderType:          BorderSingle,
		defaultCellWidth:           10,
		defaultCellHeight:          1,
		indicatorColor:             defaultNo,
		defaultPadding:             1,
	}
}