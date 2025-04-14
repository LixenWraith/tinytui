// text.go
package tinytui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// Text displays static or wrapping text content. No cursor navigation.
type Text struct {
	BaseComponent
	content      string
	wrap         bool
	lines        []string // Cache for wrapped/split lines
	scrollOffset int      // For potentially scrollable text in future
	style        Style
}

// NewText creates a new Text component with the specified content.
func NewText(content string) *Text {
	t := &Text{
		BaseComponent: NewBaseComponent(),
		content:       content,
		wrap:          false,
		scrollOffset:  0,
		style:         DefaultStyle,
	}
	return t
}

// SetContent updates the text content.
func (t *Text) SetContent(content string) {
	if t.content == content {
		return
	}

	t.content = content
	t.lines = nil // Invalidate line cache
	t.MarkDirty()
}

// GetContent returns the text content.
func (t *Text) GetContent() string {
	return t.content
}

// SetWrap sets whether the text should wrap.
func (t *Text) SetWrap(wrap bool) {
	if t.wrap == wrap {
		return
	}

	t.wrap = wrap
	t.lines = nil // Invalidate line cache
	t.MarkDirty()
}

// SetStyle sets the text style.
func (t *Text) SetStyle(style Style) {
	t.style = style
	t.MarkDirty()
}

// GetStyle returns the text style.
func (t *Text) GetStyle() Style {
	return t.style
}

// Focusable returns whether the text can receive focus.
// Text components are not focusable by default.
func (t *Text) Focusable() bool {
	return false
}

// Draw draws the text component.
func (t *Text) Draw(screen tcell.Screen) {
	// Check visibility
	if !t.IsVisible() {
		return
	}

	// Get component dimensions
	x, y, width, height := t.GetRect()
	if width <= 0 || height <= 0 {
		return
	}

	// Use component's style
	style := t.style

	// Prepare text for drawing (split into lines, handle wrapping)
	if t.lines == nil {
		t.calculateLines(width)
	}

	// Clear the component area
	Fill(screen, x, y, width, height, ' ', style)

	// Get lines for safe drawing outside lock
	visibleLines := t.getVisibleLines(height)
	startY := y

	// Draw visible lines
	for i, line := range visibleLines {
		lineY := startY + i
		if lineY >= y+height {
			break
		}

		// Ensure text doesn't overflow the width
		displayLine := line
		if runewidth.StringWidth(line) > width {
			displayLine = runewidth.Truncate(line, width, "")
		}

		DrawText(screen, x, lineY, style, displayLine)
	}
}

// calculateLines splits content into lines, handling word wrapping if enabled.
// Must be called with lock held.
func (t *Text) calculateLines(width int) {
	if width <= 0 {
		t.lines = []string{}
		return
	}

	if !t.wrap {
		// Simple line splitting without wrapping
		t.lines = strings.Split(t.content, "\n")
		return
	}

	// Word wrapping
	var wrappedLines []string
	rawLines := strings.Split(t.content, "\n")

	for _, line := range rawLines {
		if line == "" {
			wrappedLines = append(wrappedLines, "")
			continue
		}

		remainingLine := line
		for len(remainingLine) > 0 {
			idx := 0
			lineWidth := 0

			// Find how much of the line fits in the available width
			for i, r := range remainingLine {
				rWidth := runewidth.RuneWidth(r)
				if lineWidth+rWidth > width {
					break
				}
				lineWidth += rWidth
				idx = i + 1
			}

			if idx == 0 && len(remainingLine) > 0 {
				// Handle case where not even one character fits
				// Force at least one character
				idx = len(string([]rune(remainingLine)[0]))
			}

			wrappedLines = append(wrappedLines, remainingLine[:idx])
			remainingLine = remainingLine[idx:]
		}
	}

	t.lines = wrappedLines
}

// getVisibleLines returns lines that should be visible based on scroll offset.
// Must be called with lock held.
func (t *Text) getVisibleLines(maxHeight int) []string {
	if len(t.lines) == 0 || maxHeight <= 0 {
		return []string{}
	}

	// Ensure scroll offset is valid
	if t.scrollOffset >= len(t.lines) {
		t.scrollOffset = len(t.lines) - 1
	}
	if t.scrollOffset < 0 {
		t.scrollOffset = 0
	}

	// Determine visible range
	startLine := t.scrollOffset
	endLine := startLine + maxHeight
	if endLine > len(t.lines) {
		endLine = len(t.lines)
	}

	return t.lines[startLine:endLine]
}

// HandleEvent handles events for the text component.
// Text components don't handle any events by default.
func (t *Text) HandleEvent(event tcell.Event) bool {
	return false
}

// ScrollTo scrolls to the specified line index.
func (t *Text) ScrollTo(lineIndex int) {
	if t.lines == nil {
		_, _, width, _ := t.GetRect()
		t.calculateLines(width)
	}

	if lineIndex < 0 {
		lineIndex = 0
	} else if len(t.lines) > 0 && lineIndex >= len(t.lines) {
		lineIndex = len(t.lines) - 1
	}

	if t.scrollOffset != lineIndex {
		t.scrollOffset = lineIndex
		t.MarkDirty()
	}
}

// ScrollDown scrolls down by the specified number of lines.
func (t *Text) ScrollDown(count int) {
	t.ScrollTo(t.scrollOffset + count)
}

// ScrollUp scrolls up by the specified number of lines.
func (t *Text) ScrollUp(count int) {
	t.ScrollTo(t.scrollOffset - count)
}