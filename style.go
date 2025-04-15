// style.go
package tinytui

import "github.com/gdamore/tcell/v2"

// Color is an alias for tcell.Color, representing a terminal color.
// Use the ColorX constants for predefined colors.
type Color = tcell.Color

// Predefined Colors (mapping directly to tcell constants for convenience and familiarity)
const (
	ColorDefault Color = tcell.ColorDefault // Default terminal foreground/background

	// Basic ANSI Colors (0-7)
	ColorBlack  Color = tcell.ColorBlack  // 0
	ColorMaroon Color = tcell.ColorMaroon // 1 (Dark Red)
	ColorGreen  Color = tcell.ColorGreen  // 2 (Dark Green)
	ColorOlive  Color = tcell.ColorOlive  // 3 (Dark Yellow / Brown)
	ColorNavy   Color = tcell.ColorNavy   // 4 (Dark Blue)
	ColorPurple Color = tcell.ColorPurple // 5 (Dark Magenta)
	ColorTeal   Color = tcell.ColorTeal   // 6 (Dark Cyan)
	ColorSilver Color = tcell.ColorSilver // 7 (Light Gray)

	// Bright ANSI Colors (8-15)
	ColorGray    Color = tcell.ColorGray    // 8 (Dark Gray)
	ColorRed     Color = tcell.ColorRed     // 9 (Bright Red, often same as DarkRed in practice)
	ColorLime    Color = tcell.ColorLime    // 10 (Bright Green)
	ColorYellow  Color = tcell.ColorYellow  // 11 (Bright Yellow)
	ColorBlue    Color = tcell.ColorBlue    // 12 (Bright Blue)
	ColorFuchsia Color = tcell.ColorFuchsia // 13 (Bright Magenta)
	ColorAqua    Color = tcell.ColorAqua    // 14 (Bright Cyan)
	ColorWhite   Color = tcell.ColorWhite   // 15 (Bright White)

	// Explicit Dark Aliases (often map to 0-7 range)
	ColorDarkRed     Color = tcell.ColorDarkRed       // Usually same as ColorMaroon
	ColorDarkGreen   Color = tcell.ColorDarkGreen     // Usually same as ColorGreen
	ColorDarkYellow  Color = tcell.ColorDarkGoldenrod // Or Olive
	ColorDarkBlue    Color = tcell.ColorDarkBlue      // Usually same as ColorNavy
	ColorDarkMagenta Color = tcell.ColorDarkMagenta   // Usually same as ColorPurple
	ColorDarkCyan    Color = tcell.ColorDarkCyan      // Usually same as ColorTeal
	ColorDarkGray    Color = tcell.ColorDarkGray      // Usually same as ColorGray
	ColorLightGray   Color = tcell.ColorLightGray     // Usually same as ColorSilver

	// Explicit Light Aliases (often map to 8-15 range)
	ColorLightRed     Color = tcell.ColorOrangeRed   // No LightRed, Usually same as ColorRed
	ColorLightGreen   Color = tcell.ColorLightGreen  // Usually same as ColorLime
	ColorLightYellow  Color = tcell.ColorLightYellow // Usually same as ColorYellow
	ColorLightBlue    Color = tcell.ColorLightBlue   // Usually same as ColorBlue
	ColorLightMagenta Color = tcell.ColorFuchsia     // No LightMagenta, Usually same as ColorFuchsia
	ColorLightCyan    Color = tcell.ColorLightCyan   // Usually same as ColorAqua

	// Other common names (check tcell definitions)
	ColorDarkGoldenrod Color = tcell.ColorDarkGoldenrod
	ColorDarkSlateGray Color = tcell.ColorDarkSlateGray
)

// Style encapsulates the visual attributes of a terminal cell:
// foreground color, background color, and text attributes (bold, italic, etc.).
// It wraps tcell.Style for compatibility but provides a fluent interface for modification.
type Style struct {
	tcellStyle tcell.Style
}

// AttrMask is an alias for tcell.AttrMask, representing a bitmask of text attributes.
type AttrMask = tcell.AttrMask

// Text Attributes (mapping directly to tcell constants)
const (
	AttrNone      AttrMask = 0                       // No attributes.
	AttrBold      AttrMask = tcell.AttrBold          // Bold text.
	AttrBlink     AttrMask = tcell.AttrBlink         // Blinking text (terminal support varies).
	AttrReverse   AttrMask = tcell.AttrReverse       // Reverse video (swap foreground/background).
	AttrUnderline AttrMask = tcell.AttrUnderline     // Underlined text.
	AttrDim       AttrMask = tcell.AttrDim           // Dim/faint text (terminal support varies).
	AttrItalic    AttrMask = tcell.AttrItalic        // Italic text (terminal support varies).
	AttrStrike    AttrMask = tcell.AttrStrikeThrough // Strikethrough text (terminal support varies).
)

// DefaultStyle represents the base style with default terminal colors and no attributes.
// It serves as a starting point for creating custom styles.
var DefaultStyle = Style{tcellStyle: tcell.StyleDefault}

// Foreground returns a new Style with the specified foreground color set.
// Does not modify the original Style.
func (s Style) Foreground(c Color) Style {
	s.tcellStyle = s.tcellStyle.Foreground(c)
	return s
}

// Background returns a new Style with the specified background color set.
// Does not modify the original Style.
func (s Style) Background(c Color) Style {
	s.tcellStyle = s.tcellStyle.Background(c)
	return s
}

// Attributes returns a new Style with the specified text attributes mask set,
// *replacing* any previously set attributes. Use the specific attribute methods
// (e.g., Bold(true)) or bitwise OR operations to add attributes cumulatively.
// Does not modify the original Style.
func (s Style) Attributes(attrs AttrMask) Style {
	s.tcellStyle = s.tcellStyle.Attributes(attrs)
	return s
}

// Bold returns a new Style with the bold attribute set (if enable is true) or cleared (if enable is false).
// Does not modify the original Style.
func (s Style) Bold(enable bool) Style {
	s.tcellStyle = s.tcellStyle.Bold(enable)
	return s
}

// Italic returns a new Style with the italic attribute set or cleared.
// Does not modify the original Style.
func (s Style) Italic(enable bool) Style {
	s.tcellStyle = s.tcellStyle.Italic(enable)
	return s
}

// Underline returns a new Style with the underline attribute set or cleared.
// Does not modify the original Style.
func (s Style) Underline(enable bool) Style {
	s.tcellStyle = s.tcellStyle.Underline(enable)
	return s
}

// Reverse returns a new Style with the reverse video attribute set or cleared.
// Does not modify the original Style.
func (s Style) Reverse(enable bool) Style {
	s.tcellStyle = s.tcellStyle.Reverse(enable)
	return s
}

// Blink returns a new Style with the blink attribute set or cleared.
// Does not modify the original Style.
func (s Style) Blink(enable bool) Style {
	s.tcellStyle = s.tcellStyle.Blink(enable)
	return s
}

// Dim returns a new Style with the dim attribute set or cleared.
// Does not modify the original Style.
func (s Style) Dim(enable bool) Style {
	s.tcellStyle = s.tcellStyle.Dim(enable)
	return s
}

// StrikeThrough returns a new Style with the strikethrough attribute set or cleared.
// Does not modify the original Style.
func (s Style) StrikeThrough(enable bool) Style {
	s.tcellStyle = s.tcellStyle.StrikeThrough(enable)
	return s
}

// Deconstruct breaks down the style into its component parts: foreground color,
// background color, and attributes mask. It also returns a boolean `bgSet` which
// is true if the background color is *not* the default terminal background color.
// This helps determine if a style intends to be opaque or transparent.
func (s Style) Deconstruct() (fg Color, bg Color, attrs AttrMask, bgSet bool) {
	fg, bg, attrs = s.tcellStyle.Decompose()
	// Check if the background color is different from the *global* default background.
	_, defaultBg, _ := tcell.StyleDefault.Decompose()
	// Note: This comparison might be tricky if the default background isn't truly "default".
	// Tcell uses special values; comparing directly should work.
	bgSet = (bg != defaultBg)
	return fg, bg, attrs, bgSet
}

// MergeWith creates a new style by overlaying the properties of 'other' onto 's'.
// - Foreground: Uses 'other' foreground if it's not ColorDefault, otherwise uses 's' foreground.
// - Background: Uses 'other' background if it's explicitly set (`bgSet` is true for 'other'), otherwise uses 's' background.
// - Attributes: Combines attributes from both styles using bitwise OR.
func (s Style) MergeWith(other Style) Style {
	fg1, bg1, attrs1, bgSet1 := s.Deconstruct()
	fg2, bg2, attrs2, bgSet2 := other.Deconstruct()

	finalFg := fg1
	finalBg := bg1
	finalAttrs := attrs1
	finalBgSet := bgSet1 // Track if the final background is explicitly set

	// Apply foreground from 'other' if it's not the default color
	if fg2 != ColorDefault {
		finalFg = fg2
	}

	// Apply background from 'other' only if it was explicitly set
	if bgSet2 {
		finalBg = bg2
		finalBgSet = true // Mark that the background is now explicitly set
	}

	// Combine attributes using bitwise OR
	finalAttrs |= attrs2

	// Reconstruct the final style carefully
	result := DefaultStyle // Start from default
	if finalFg != ColorDefault {
		result = result.Foreground(finalFg)
	}
	// Apply background *only if* the final determination was that it should be set
	if finalBgSet {
		result = result.Background(finalBg)
	}
	result = result.Attributes(finalAttrs) // Apply combined attributes

	return result
}

// ToTcell converts this tinytui Style back into the underlying tcell.Style
// required by tcell screen drawing methods.
func (s Style) ToTcell() tcell.Style {
	return s.tcellStyle
}