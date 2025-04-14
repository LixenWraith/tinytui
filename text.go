// text.go
package tinytui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// Text displays static or wrapping text content. It is typically not focusable or interactive,
// serving as a label or display area. Supports basic scrolling.
type Text struct {
	BaseComponent
	content      string
	wrap         bool          // Should text wrap within component width?
	lines        []string      // Cache of processed lines (split by newline, potentially wrapped)
	scrollOffset int           // Index (0-based) of the first visible line
	style        Style         // Style applied to the text
	alignment    AlignmentText // Horizontal text alignment (Left, Center, Right)
}

// AlignmentText defines horizontal text alignment options within the component's bounds.
type AlignmentText int

const (
	AlignTextLeft   AlignmentText = iota // Align text to the left edge (default).
	AlignTextCenter                      // Center text horizontally.
	AlignTextRight                       // Align text to the right edge.
)

// NewText creates a new Text component with the specified initial content.
// Initializes style from the current theme.
func NewText(content string) *Text {
	theme := GetTheme()
	if theme == nil {
		theme = NewDefaultTheme()
	} // Fallback

	t := &Text{
		BaseComponent: NewBaseComponent(),
		content:       content,
		wrap:          false, // No wrapping by default
		scrollOffset:  0,
		style:         theme.TextStyle(), // Use theme default text style
		alignment:     AlignTextLeft,     // Default alignment
		// lines cache starts nil, calculated on first Draw or Scroll
	}
	// Apply theme initially to set the style correctly
	t.ApplyTheme(theme)
	return t
}

// ApplyTheme updates the text's style based on the provided theme.
// Implements ThemedComponent.
func (t *Text) ApplyTheme(theme Theme) {
	if theme == nil {
		return
	}
	newStyle := theme.TextStyle()
	if t.style != newStyle {
		t.style = newStyle
		t.MarkDirty() // Style change requires redraw
	}
}

// SetContent updates the text displayed by the component.
// Resets the line cache and scroll position.
func (t *Text) SetContent(content string) {
	if t.content == content {
		return
	} // No change

	t.content = content
	t.lines = nil      // Invalidate line cache, needs recalculation
	t.scrollOffset = 0 // Reset scroll offset when content changes
	t.MarkDirty()
}

// GetContent returns the raw, unprocessed text content assigned to the component.
func (t *Text) GetContent() string {
	return t.content
}

// SetWrap enables or disables text wrapping within the component's width.
// Invalidates the line cache if the setting changes.
func (t *Text) SetWrap(wrap bool) {
	if t.wrap == wrap {
		return
	} // No change

	t.wrap = wrap
	t.lines = nil // Invalidate line cache, wrapping changes line breaks
	t.MarkDirty()
}

// SetStyle explicitly sets the text style, overriding the theme default.
// Consider using themes for consistent styling.
func (t *Text) SetStyle(style Style) {
	if t.style != style {
		t.style = style
		t.MarkDirty()
	}
}

// GetStyle returns the current text style used by the component.
func (t *Text) GetStyle() Style {
	return t.style
}

// SetAlignment sets the horizontal text alignment (Left, Center, Right).
func (t *Text) SetAlignment(align AlignmentText) {
	if t.alignment != align {
		t.alignment = align
		t.MarkDirty() // Alignment change requires redraw
	}
}

// Focusable returns false, as Text components are not typically interactive or focusable.
func (t *Text) Focusable() bool {
	return false
}

// Draw renders the text component onto the screen, handling wrapping, scrolling, and alignment.
func (t *Text) Draw(screen tcell.Screen) {
	if !t.IsVisible() {
		return
	}

	x, y, width, height := t.GetRect()
	if width <= 0 || height <= 0 {
		return
	} // Cannot draw in zero area

	// Ensure lines are calculated based on current width and wrap setting
	// calculateLines is memoized via t.lines being nil or not.
	t.ensureLinesCalculated(width)

	// Clear the component area with the text style's background
	Fill(screen, x, y, width, height, ' ', t.style)

	// Get the slice of lines actually visible based on scroll offset and height
	visibleLines := t.getVisibleLines(height)

	// Draw the visible lines
	for i, line := range visibleLines {
		lineScreenY := y + i // Calculate screen Y coordinate for this line

		// Truncate line if it's somehow wider than the component width (safeguard)
		// runewidth.Truncate handles wide chars correctly.
		displayLine := runewidth.Truncate(line, width, "…") // Use ellipsis for truncation

		// Calculate horizontal starting position based on alignment
		lineScreenX := x
		lineWidth := runewidth.StringWidth(displayLine) // Get visual width of the line to draw

		switch t.alignment {
		case AlignTextCenter:
			lineScreenX = x + (width-lineWidth)/2
		case AlignTextRight:
			lineScreenX = x + width - lineWidth
			// case AlignTextLeft: // Default, lineScreenX remains x
		}
		// Ensure alignment doesn't push text off-screen left (shouldn't happen with truncation)
		if lineScreenX < x {
			lineScreenX = x
		}

		// Draw the text for this line at the calculated position
		DrawText(screen, lineScreenX, lineScreenY, t.style, displayLine)
	}
}

// ensureLinesCalculated makes sure the t.lines cache is populated.
// Calls calculateLines only if the cache is nil (invalidated).
func (t *Text) ensureLinesCalculated(currentWidth int) {
	if t.lines == nil {
		t.calculateLines(currentWidth)
	}
	// TODO: Also recalculate if currentWidth != previously_calculated_width?
	// This requires storing the width used for the last calculation.
	// For now, rely on SetRect invalidating or content changes invalidating.
}

// calculateLines processes the raw content into display lines based on wrapping and width.
// The result is cached in the `t.lines` slice.
func (t *Text) calculateLines(maxWidth int) {
	if maxWidth <= 0 {
		t.lines = []string{} // No space, no lines
		return
	}

	// Split content by explicit newline characters first.
	rawLines := strings.Split(t.content, "\n")
	processedLines := make([]string, 0, len(rawLines)) // Estimate capacity

	if !t.wrap {
		// No wrapping enabled, just use the raw lines directly.
		// Truncation will happen during Draw if lines exceed maxWidth.
		processedLines = rawLines
	} else {
		// Word wrapping logic
		for _, line := range rawLines {
			// Handle empty lines resulting from consecutive newlines
			if line == "" {
				processedLines = append(processedLines, "")
				continue
			}

			// Use rune-aware processing for wrapping
			lineRunes := []rune(line)
			startIndex := 0 // Start index of the current segment being processed
			for startIndex < len(lineRunes) {
				endIndex := startIndex
				currentLineWidth := 0
				lastPotentialBreak := startIndex // Index after the last space found

				// Find the maximum number of runes that fit within maxWidth
				for endIndex < len(lineRunes) {
					r := lineRunes[endIndex]
					rWidth := runewidth.RuneWidth(r)

					if currentLineWidth+rWidth > maxWidth {
						break // This rune doesn't fit
					}
					currentLineWidth += rWidth

					// Track last space for potential word break
					if r == ' ' {
						lastPotentialBreak = endIndex + 1
					}
					endIndex++
				}

				// Determine the actual break point
				breakIndex := endIndex
				if endIndex < len(lineRunes) { // If we didn't reach the end of the line...
					// ...and we found a space to break at within the fitted segment...
					if lastPotentialBreak > startIndex {
						breakIndex = lastPotentialBreak // Break at the space
					} else {
						// No space found, and the segment exceeds width.
						// Force break at endIndex (middle of a word).
						// Ensure at least one character is included if first char is too wide.
						if breakIndex == startIndex && currentLineWidth == 0 && endIndex < len(lineRunes) {
							breakIndex = startIndex + 1
						} else if breakIndex == startIndex {
							// If the first word itself is too long, breakIndex remains endIndex
							// Example: "Superlongwordthatdoesntfit"
							// breakIndex should allow the Truncate in Draw to handle it?
							// Or should we truncate here? Let's break forcefully.
							if startIndex == 0 && runewidth.StringWidth(string(lineRunes[startIndex:endIndex])) > maxWidth {
								// Force break after maxWidth runes approx. Difficult with variable width.
								// Let Draw handle truncation in this edge case for simplicity.
								// For calculation here, take what fits.
								breakIndex = endIndex // Take the part that fits
							}
						}
					}
				}

				// Add the segment to processed lines, trimming trailing space if broken at space
				segment := lineRunes[startIndex:breakIndex]
				// Trim trailing space only if we broke at a space (lastPotentialBreak == breakIndex)
				// if lastPotentialBreak == breakIndex && len(segment) > 0 && segment[len(segment)-1] == ' ' {
				//      segment = segment[:len(segment)-1]
				// }
				// Simpler: let's not trim here, Draw handles final display width.

				processedLines = append(processedLines, string(segment))
				startIndex = breakIndex // Start next segment after the break
			}
		}
	}

	t.lines = processedLines // Cache the result
}

// getVisibleLines returns the slice of processed lines that should be visible
// based on the current scrollOffset and available component height.
func (t *Text) getVisibleLines(maxHeight int) []string {
	// Ensure lines are calculated first
	if t.lines == nil {
		// This should ideally not happen if ensureLinesCalculated was called in Draw
		// But as a fallback, try calculating now (might use outdated width if rect changed).
		t.ensureLinesCalculated(t.rect.Width)
	}

	if len(t.lines) == 0 || maxHeight <= 0 {
		return []string{}
	}

	// Clamp scroll offset to valid range [0, len(lines)-1]
	lastPossibleOffset := len(t.lines) - 1
	if lastPossibleOffset < 0 {
		lastPossibleOffset = 0
	} // Handle empty lines case result
	if t.scrollOffset < 0 {
		t.scrollOffset = 0
	}
	if t.scrollOffset > lastPossibleOffset {
		t.scrollOffset = lastPossibleOffset
	}

	// Determine the range of lines to display [start, end)
	startLine := t.scrollOffset
	endLine := startLine + maxHeight // Calculate potential end index (exclusive)
	if endLine > len(t.lines) {
		endLine = len(t.lines) // Clamp end index to actual number of lines
	}

	// Return the visible slice, handle invalid range possibility
	if startLine >= endLine || startLine < 0 {
		return []string{}
	}
	return t.lines[startLine:endLine]
}

// HandleEvent processes events. Text components typically don't handle events by default.
// Scrolling could potentially be added here if the component were made focusable.
func (t *Text) HandleEvent(event tcell.Event) bool {
	// Example: Make Text scrollable if focusable
	// if t.Focusable() && t.IsFocused() {
	// 	if keyEvent, ok := event.(*tcell.EventKey); ok {
	// 		switch keyEvent.Key() {
	// 		case tcell.KeyDown:
	// 			t.ScrollDown(1)
	// 			return true
	// 		case tcell.KeyUp:
	// 			t.ScrollUp(1)
	// 			return true
	// 		case tcell.KeyPgDn:
	// 			_, _, _, h := t.GetRect()
	// 			t.ScrollDown(max(1, h)) // Scroll approx one page
	// 			return true
	// 		case tcell.KeyPgUp:
	// 			_, _, _, h := t.GetRect()
	// 			t.ScrollUp(max(1, h)) // Scroll approx one page
	// 			return true
	// 		}
	// 	}
	// }
	return false // Event not handled
}

// ScrollTo attempts to scroll the text so that the specified line index is at the top.
// Line index is 0-based. Clamps to valid range. Recalculates lines if needed.
func (t *Text) ScrollTo(lineIndex int) {
	// Ensure lines are calculated based on current width before scrolling
	t.ensureLinesCalculated(t.rect.Width)

	numLines := len(t.lines)
	targetOffset := lineIndex

	// Clamp target offset to valid range [0, numLines-1]
	if numLines == 0 {
		targetOffset = 0
	} else {
		if targetOffset < 0 {
			targetOffset = 0
		}
		lastLineIdx := numLines - 1
		if targetOffset > lastLineIdx {
			targetOffset = lastLineIdx
		}
	}

	// Only update and mark dirty if the offset actually changes
	if t.scrollOffset != targetOffset {
		t.scrollOffset = targetOffset
		t.MarkDirty()
	}
}

// ScrollDown scrolls down by the specified number of lines. Does nothing if count <= 0.
func (t *Text) ScrollDown(count int) {
	if count <= 0 {
		return
	}
	t.ScrollTo(t.scrollOffset + count)
}

// ScrollUp scrolls up by the specified number of lines. Does nothing if count <= 0.
func (t *Text) ScrollUp(count int) {
	if count <= 0 {
		return
	}
	t.ScrollTo(t.scrollOffset - count)
}