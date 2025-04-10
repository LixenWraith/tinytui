// widgets/text.go
package widgets

import (
	"strings"
	"sync"

	"github.com/LixenWraith/tinytui"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// Text is a widget for displaying static or wrapping text.
type Text struct {
	tinytui.BaseWidget
	mu      sync.RWMutex  // Protects access to content and lines
	content string        // The raw text content
	style   tinytui.Style // Style for the text
	wrap    bool          // Whether to wrap text
	lines   []string      // Cached wrapped lines
}

// NewText creates a new Text widget.
func NewText(content string) *Text {
	t := &Text{
		content: content,
		style:   tinytui.DefaultStyle,
		wrap:    false, // Default to no wrapping
		lines:   nil,   // Initially nil, calculated on demand or SetRect
	}
	t.SetVisible(true)
	// Text is not focusable by default
	return t
}

// SetContent updates the text content displayed by the widget.
// NOTE: Return type changed from *Text to void to satisfy tinytui.TextUpdater interface.
func (t *Text) SetContent(content string) {
	t.mu.Lock()
	if t.content == content {
		t.mu.Unlock()
		return // No change
	}
	t.content = content
	t.lines = nil // Invalidate cached lines
	t.mu.Unlock()

	if app := t.App(); app != nil {
		app.QueueRedraw()
	}
	// No longer returns t
}

// SetStyle sets the style used to draw the text.
func (t *Text) SetStyle(style tinytui.Style) *Text {
	t.mu.Lock()
	t.style = style
	t.mu.Unlock()

	if app := t.App(); app != nil {
		app.QueueRedraw()
	}
	return t
}

// SetWrap enables or disables word wrapping.
func (t *Text) SetWrap(wrap bool) *Text {
	t.mu.Lock()
	if t.wrap == wrap {
		t.mu.Unlock()
		return t
	}
	t.wrap = wrap
	t.lines = nil // Invalidate cached lines, needs recalculation
	t.mu.Unlock()

	if app := t.App(); app != nil {
		app.QueueRedraw()
	}
	return t
}

// recalculateLines updates the internal 'lines' cache based on content,
// wrap setting, and current widget width.
// Must be called with t.mu held or when mutex is not needed (e.g., init).
func (t *Text) recalculateLines() {
	_, _, width, _ := t.GetRect() // Get current width

	// If width is zero or negative, can't calculate lines
	if width <= 0 {
		t.lines = []string{} // Set to empty, not nil
		return
	}

	if !t.wrap {
		// No wrapping, just split by explicit newlines
		t.lines = strings.Split(t.content, "\n")
		return
	}

	// --- Word wrapping logic (adopted from context version) ---
	var calculatedLines []string
	paragraphs := strings.Split(t.content, "\n") // Handle explicit newlines first

	for _, paragraph := range paragraphs {
		if paragraph == "" { // Handle empty lines from double newlines
			calculatedLines = append(calculatedLines, "")
			continue
		}

		wordsInParagraph := strings.Fields(paragraph) // Split paragraph by spaces
		if len(wordsInParagraph) == 0 {               // Handle lines with only spaces
			calculatedLines = append(calculatedLines, "") // Treat as empty line
			continue
		}

		currentLine := ""
		currentLineWidth := 0

		for _, word := range wordsInParagraph {
			wordWidth := runewidth.StringWidth(word)

			// If the word itself is wider than the available width, it needs hard break
			if wordWidth > width {
				// Break the long word
				if currentLineWidth > 0 { // Add the current line before breaking word
					calculatedLines = append(calculatedLines, currentLine)
					currentLine = ""
					currentLineWidth = 0
				}
				// Hard break the word character by character
				brokenWordPart := ""
				brokenWordWidth := 0
				for _, r := range word {
					rw := runewidth.RuneWidth(r)
					if brokenWordWidth+rw > width {
						calculatedLines = append(calculatedLines, brokenWordPart)
						brokenWordPart = string(r)
						brokenWordWidth = rw
					} else {
						brokenWordPart += string(r)
						brokenWordWidth += rw
					}
				}
				// The last part of the broken word becomes the start of the next potential line
				currentLine = brokenWordPart
				currentLineWidth = brokenWordWidth
				// Don't immediately add this part; it might fit with the next word
				continue // Move to the next word
			}

			// Check if the word fits on the current line
			separatorWidth := 0
			if currentLineWidth > 0 {
				separatorWidth = 1 // Space separator
			}

			if currentLineWidth+separatorWidth+wordWidth <= width {
				// Word fits
				if currentLineWidth > 0 {
					currentLine += " "
				}
				currentLine += word
				currentLineWidth += separatorWidth + wordWidth
			} else {
				// Word doesn't fit, start a new line
				calculatedLines = append(calculatedLines, currentLine)
				currentLine = word
				currentLineWidth = wordWidth
			}
		}
		// Add the last line of the paragraph
		if currentLine != "" {
			calculatedLines = append(calculatedLines, currentLine)
		}
	}

	t.lines = calculatedLines
	// --- End word wrapping logic ---
}

// Draw draws the text content within the widget's bounds.
func (t *Text) Draw(screen tcell.Screen) {
	if !t.IsVisible() {
		return
	}
	x, y, width, height := t.GetRect()
	if width <= 0 || height <= 0 {
		return // Nothing to draw
	}

	t.mu.RLock() // Use RLock for reading content/lines
	// Ensure lines are calculated if needed
	linesNeedRecalc := t.lines == nil // Check if lines are nil under RLock

	if linesNeedRecalc {
		// Need to release RLock and acquire Lock to modify t.lines
		t.mu.RUnlock()
		t.mu.Lock()
		// Double check after acquiring write lock in case another goroutine calculated it
		if t.lines == nil {
			t.recalculateLines() // This uses GetRect, safe as we hold the lock
		}
		t.mu.Unlock()
		t.mu.RLock() // Re-acquire read lock for drawing
	}

	// If lines is *still* nil after trying to recalculate (e.g., width was 0), return
	if t.lines == nil {
		t.mu.RUnlock()
		return
	}

	currentStyle := t.style
	linesToDraw := t.lines
	t.mu.RUnlock() // Unlock after accessing shared data

	// Fill background first (optional, depends on desired behavior)
	// tinytui.Fill(screen, x, y, width, height, ' ', currentStyle)

	// Draw the lines
	for i, line := range linesToDraw {
		if i >= height {
			break // Don't draw more lines than the widget's height
		}
		// DrawText handles clipping within the line's width
		tinytui.DrawText(screen, x, y+i, currentStyle, line)
	}

	// Clear remaining lines if content is shorter than height
	for i := len(linesToDraw); i < height; i++ {
		tinytui.Fill(screen, x, y+i, width, 1, ' ', currentStyle)
	}
}

// SetRect updates the widget's dimensions and recalculates wrapped lines if needed.
func (t *Text) SetRect(x, y, width, height int) {
	t.mu.Lock()
	// Check if width actually changed, matters for wrapping
	_, _, oldWidth, _ := t.GetRect() // Get old dimensions before setting new ones
	needsRecalc := t.wrap && (width != oldWidth || t.lines == nil)

	t.BaseWidget.SetRect(x, y, width, height) // Call embedded method to update rect

	if needsRecalc {
		t.recalculateLines() // Recalculate lines based on new width
	}
	t.mu.Unlock()
	// No redraw queued here, SetRect is usually called during a redraw cycle
}

// Focusable returns false, Text widgets are not focusable by default.
func (t *Text) Focusable() bool {
	return false
}

// HandleEvent handles events for the Text widget.
// By default, it only delegates to BaseWidget for potential keybindings
// set directly on the Text widget itself (uncommon for static text).
func (t *Text) HandleEvent(event tcell.Event) bool {
	// Let BaseWidget handle its own keybindings, if any were set.
	return t.BaseWidget.HandleEvent(event)
}