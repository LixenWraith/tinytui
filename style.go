// style.go
package tinytui

import "github.com/gdamore/tcell/v2"

// --- Color Abstraction ---

// Color represents a TUI color. It wraps tcell.Color.
type Color = tcell.Color

// Predefined Colors (mapping to tcell)
const (
	ColorDefault       Color = tcell.ColorDefault
	ColorBlack         Color = tcell.ColorBlack
	ColorRed           Color = tcell.ColorRed
	ColorGreen         Color = tcell.ColorGreen
	ColorYellow        Color = tcell.ColorYellow
	ColorBlue          Color = tcell.ColorBlue
	ColorMagenta       Color = tcell.ColorDarkMagenta // User corrected
	ColorCyan          Color = tcell.ColorLightCyan   // User corrected
	ColorWhite         Color = tcell.ColorWhite
	ColorGray          Color = tcell.ColorGray // Note: tcell names might differ slightly
	ColorDarkGray      Color = tcell.ColorDarkGray
	ColorLightGray     Color = tcell.ColorLightGray
	ColorSilver        Color = tcell.ColorSilver
	ColorNavy          Color = tcell.ColorNavy
	ColorAqua          Color = tcell.ColorAqua
	ColorLime          Color = tcell.ColorLime
	ColorMaroon        Color = tcell.ColorMaroon
	ColorOlive         Color = tcell.ColorOlive
	ColorPurple        Color = tcell.ColorPurple
	ColorTeal          Color = tcell.ColorTeal
	ColorFuchsia       Color = tcell.ColorFuchsia
	ColorDarkRed       Color = tcell.ColorDarkRed
	ColorDarkGreen     Color = tcell.ColorDarkGreen
	ColorDarkMagenta   Color = tcell.ColorDarkMagenta
	ColorDarkCyan      Color = tcell.ColorDarkCyan
	ColorDarkBlue      Color = tcell.ColorDarkBlue
	ColorDarkGoldenrod Color = tcell.ColorDarkGoldenrod
	ColorDarkSlateGray Color = tcell.ColorDarkSlateGray
	ColorLightCyan     Color = tcell.ColorLightCyan
	ColorLightGreen    Color = tcell.ColorLightGreen
	ColorLightYellow   Color = tcell.ColorLightYellow
	// Add more named colors as needed
)

// --- Style Abstraction ---

// Style represents the display style of a cell (foreground, background, attributes).
// It currently wraps tcell.Style.
type Style struct {
	tcellStyle tcell.Style
}

// AttrMask represents text attributes (bold, blink, etc.). It wraps tcell.AttrMask.
type AttrMask = tcell.AttrMask // Type alias

// Attributes (mapping to tcell)
const (
	AttrBold      AttrMask = tcell.AttrBold
	AttrBlink     AttrMask = tcell.AttrBlink
	AttrReverse   AttrMask = tcell.AttrReverse
	AttrUnderline AttrMask = tcell.AttrUnderline
	AttrDim       AttrMask = tcell.AttrDim
	AttrItalic    AttrMask = tcell.AttrItalic        // Requires modern terminal support
	AttrStrike    AttrMask = tcell.AttrStrikeThrough // Requires modern terminal support
	AttrNone      AttrMask = 0
)

// DefaultStyle provides a baseline style.
var DefaultStyle = Style{tcellStyle: tcell.StyleDefault}

// Foreground sets the foreground color.
func (s Style) Foreground(c Color) Style {
	s.tcellStyle = s.tcellStyle.Foreground(c)
	return s
}

// Background sets the background color.
func (s Style) Background(c Color) Style {
	s.tcellStyle = s.tcellStyle.Background(c)
	return s
}

// Attributes sets the text attributes, replacing existing ones.
func (s Style) Attributes(attrs AttrMask) Style {
	s.tcellStyle = s.tcellStyle.Attributes(attrs)
	return s
}

// Bold sets or clears the bold attribute.
func (s Style) Bold(enable bool) Style {
	s.tcellStyle = s.tcellStyle.Bold(enable)
	return s
}

// Italic sets or clears the italic attribute.
func (s Style) Italic(enable bool) Style {
	s.tcellStyle = s.tcellStyle.Italic(enable)
	return s
}

// Underline sets or clears the underline attribute.
func (s Style) Underline(enable bool) Style {
	s.tcellStyle = s.tcellStyle.Underline(enable)
	return s
}

// Reverse sets or clears the reverse attribute.
func (s Style) Reverse(enable bool) Style {
	s.tcellStyle = s.tcellStyle.Reverse(enable)
	return s
}

// Blink sets or clears the blink attribute.
func (s Style) Blink(enable bool) Style {
	s.tcellStyle = s.tcellStyle.Blink(enable)
	return s
}

// Dim sets or clears the dim attribute.
func (s Style) Dim(enable bool) Style {
	s.tcellStyle = s.tcellStyle.Dim(enable)
	return s
}

// StrikeThrough sets or clears the strikethrough attribute.
func (s Style) StrikeThrough(enable bool) Style {
	s.tcellStyle = s.tcellStyle.StrikeThrough(enable)
	return s
}

// Deconstruct breaks down the style into its component parts.
// It returns the foreground color, background color, and attributes.
// It also returns a boolean indicating if the background color was explicitly set.
func (s Style) Deconstruct() (fg Color, bg Color, attrs AttrMask, bgSet bool) {
	// TODO: Modify Decompose in the future since it's deprecated (requires more complex implementation with tcell)
	fg, bg, attrs = s.tcellStyle.Decompose()
	// Check if the background color is different from the default background
	// This is an approximation for whether it was explicitly set.
	_, defaultBg, _ := tcell.StyleDefault.Decompose()
	bgSet = (bg != defaultBg)
	return fg, bg, attrs, bgSet
}

// GetStateStyle returns the appropriate style for a given widget state
func (s Style) GetStateStyle(state WidgetState, focused bool) Style {
	switch {
	case focused && state == StateInteracted:
		// Return focused+interacted style - using reverse and bold attributes by default
		return s.Reverse(true).Bold(true)
	case focused && state == StateSelected:
		// Return focused+selected style - using reverse attribute by default
		return s.Reverse(true)
	case focused:
		// Return focused style - using dim and underline attributes by default
		return s.Dim(false).Underline(true)
	case state == StateInteracted:
		// Return interacted style - using bold attribute by default
		return s.Bold(true)
	case state == StateSelected:
		// Return selected style - using dim and underline attributes by default
		return s.Dim(true).Underline(true)
	default:
		// Return normal style
		return s
	}
}

// MergeWith creates a new style by merging attributes from another style
// The other style's attributes will override this style's attributes if set
func (s Style) MergeWith(other Style) Style {
	_, _, attrs1, _ := s.Deconstruct()
	fg2, bg2, attrs2, bgSet2 := other.Deconstruct()

	result := s

	// Apply foreground if not default
	if fg2 != ColorDefault {
		result = result.Foreground(fg2)
	}

	// Apply background if explicitly set
	if bgSet2 {
		result = result.Background(bg2)
	}

	// Merge attributes
	if attrs2 != AttrNone {
		result = result.Attributes(attrs1 | attrs2)
	}

	return result
}

// --- Border Types ---

// BorderType defines the style of border to draw.
type BorderType int

const (
	BorderNone   BorderType = iota // No border
	BorderSingle                   // Single line border (uses tcell.Rune HLine, VLine, etc.)
	BorderDouble                   // Double line border (uses tcell.Rune DoubleHLine, DoubleVLine, etc.)
	BorderSolid                    // Solid block border (uses block elements like █, ▀, ▄)
)

// --- Helper Functions ---

// ToTcell converts tinytui.Style to tcell.Style for internal use
// or when direct screen manipulation is needed.
func (s Style) ToTcell() tcell.Style {
	return s.tcellStyle
}