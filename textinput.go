// textinput.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// TextInput allows single-line text entry with cursor.
type TextInput struct {
	BaseComponent
	buffer       []rune
	cursorPos    int // Rune index within buffer
	visualOffset int // Rune index of the start of the visible part of buffer
	style        Style
	maxLength    int
	onChange     func(string)
	onSubmit     func(string)
	masked       bool // For password input
	maskRune     rune // Character used for masking
}

// NewTextInput creates a new text input component.
func NewTextInput() *TextInput {
	t := &TextInput{
		BaseComponent: NewBaseComponent(),
		buffer:        []rune{},
		cursorPos:     0,
		visualOffset:  0,
		style:         DefaultStyle,
		maxLength:     0, // 0 means no limit
		masked:        false,
		maskRune:      '*',
	}
	return t
}

// SetText sets the text content.
func (t *TextInput) SetText(text string) {
	// Convert to runes
	newBuffer := []rune(text)

	// Apply maxLength if set
	if t.maxLength > 0 && len(newBuffer) > t.maxLength {
		newBuffer = newBuffer[:t.maxLength]
	}

	// Only update if content changed
	if runesEqual(t.buffer, newBuffer) {
		return
	}

	oldText := string(t.buffer)
	t.buffer = newBuffer

	// Ensure cursor position is valid
	if t.cursorPos > len(t.buffer) {
		t.cursorPos = len(t.buffer)
	}

	// Reset visual offset
	t.updateVisualOffset()

	t.MarkDirty()

	// Cache values outside lock
	onChange := t.onChange
	newText := string(t.buffer)

	// Trigger change handler if set
	if onChange != nil && oldText != newText {
		onChange(newText)
	}
}

// SetContent is an alias for SetText to implement TextUpdater interface.
func (t *TextInput) SetContent(text string) {
	t.SetText(text)
}

// GetText returns the current text content.
func (t *TextInput) GetText() string {
	return string(t.buffer)
}

// SetStyle sets the component style.
func (t *TextInput) SetStyle(style Style) {
	t.style = style
	t.MarkDirty()
}

// SetMaxLength sets the maximum text length (0 for no limit).
func (t *TextInput) SetMaxLength(max int) {
	// Don't allow negative values
	if max < 0 {
		max = 0
	}

	t.maxLength = max

	// Truncate existing content if needed
	if max > 0 && len(t.buffer) > max {
		oldText := string(t.buffer)
		t.buffer = t.buffer[:max]

		// Ensure cursor position is valid
		if t.cursorPos > max {
			t.cursorPos = max
		}

		// Reset visual offset
		t.updateVisualOffset()

		t.MarkDirty()

		// Trigger change handler if set
		onChange := t.onChange
		if onChange != nil {
			newText := string(t.buffer)
			if oldText != newText {
				onChange(newText)
			}
			return
		}
	}
}

// SetMasked sets whether the input should be masked (for passwords).
func (t *TextInput) SetMasked(masked bool, maskRune rune) {
	if t.masked == masked && (masked == false || t.maskRune == maskRune) {
		return
	}

	t.masked = masked
	if masked {
		t.maskRune = maskRune
	}

	t.MarkDirty()
}

// SetOnChange sets the handler for text change events.
func (t *TextInput) SetOnChange(handler func(string)) {
	t.onChange = handler
}

// SetOnSubmit sets the handler for submit events (Enter key).
func (t *TextInput) SetOnSubmit(handler func(string)) {
	t.onSubmit = handler
}

// Focusable returns whether the component can receive focus.
// TextInput is focusable if visible.
func (t *TextInput) Focusable() bool {
	return t.IsVisible()
}

// Draw draws the text input component.
func (t *TextInput) Draw(screen tcell.Screen) {
	// Check visibility
	if !t.IsVisible() {
		return
	}

	// Get component dimensions
	x, y, width, height := t.GetRect()
	if width <= 0 || height <= 0 {
		return
	}

	// Get current style based on focus state
	style := t.style
	if t.IsFocused() {
		// Use a style that indicates focus
		style = style.Reverse(true)
	}

	// Clear background
	Fill(screen, x, y, width, height, ' ', style)

	// Prepare text for display
	displayText := t.buffer
	if t.masked {
		// Replace with mask character
		displayText = make([]rune, len(t.buffer))
		for i := range displayText {
			displayText[i] = t.maskRune
		}
	}

	// Calculate visible portion based on cursor and width
	visibleText := t.getVisibleText(displayText, width)

	// Calculate cursor position on screen
	cursorScreenPos := x
	if t.IsFocused() {
		cursorScreenPos = x + runewidth.StringWidth(string(visibleText[:t.cursorPos-t.visualOffset]))

		// Request cursor at this position
		if app := t.App(); app != nil {
			if cm := app.GetCursorManager(); cm != nil {
				cm.Request(cursorScreenPos, y)
			}
		}
	}

	// Save displayed text for drawing outside of lock
	displayString := string(visibleText)

	// Draw the text
	DrawText(screen, x, y, style, displayString)
}

// getVisibleText returns the portion of text that should be visible.
// Must be called with lock held.
func (t *TextInput) getVisibleText(text []rune, width int) []rune {
	if len(text) == 0 {
		return []rune{}
	}

	// Always ensure cursor is visible by adjusting visualOffset if needed
	t.updateVisualOffset()

	// Start from visualOffset
	if t.visualOffset >= len(text) {
		return []rune{}
	}

	// Determine how much text fits in the available width
	availWidth := width
	visibleEnd := t.visualOffset

	for visibleEnd < len(text) {
		charWidth := runewidth.RuneWidth(text[visibleEnd])
		if charWidth > availWidth {
			break
		}
		availWidth -= charWidth
		visibleEnd++
	}

	return text[t.visualOffset:visibleEnd]
}

// updateVisualOffset ensures cursor is visible by adjusting visualOffset.
// Must be called with lock held.
func (t *TextInput) updateVisualOffset() {
	if t.cursorPos < t.visualOffset {
		// Cursor is before the visible area, adjust left
		t.visualOffset = t.cursorPos
	} else {
		// Check if cursor would be visible given current visualOffset
		// Access rect directly since lock is already held
		width := t.rect.Width

		// Skip calculation if width is not set yet (during initialization)
		if width <= 0 {
			t.visualOffset = 0
			return
		}

		// Calculate width of text from visualOffset to cursor
		textWidth := 0
		for i := t.visualOffset; i < t.cursorPos; i++ {
			if i >= len(t.buffer) {
				break
			}
			textWidth += runewidth.RuneWidth(t.buffer[i])
		}

		// If cursor is off-screen, adjust visualOffset
		if textWidth >= width {
			// Find new visualOffset where cursor would be visible
			newOffset := t.visualOffset
			visibleWidth := width

			for newOffset < t.cursorPos {
				if newOffset >= len(t.buffer) {
					break
				}

				charWidth := runewidth.RuneWidth(t.buffer[newOffset])
				if visibleWidth <= charWidth {
					break
				}

				visibleWidth -= charWidth
				newOffset++
			}

			t.visualOffset = newOffset
		}
	}
}

// HandleEvent handles input events.
func (t *TextInput) HandleEvent(event tcell.Event) bool {
	// Only handle key events
	keyEvent, ok := event.(*tcell.EventKey)
	if !ok {
		return false
	}

	// Track if content changed
	contentChanged := false

	switch keyEvent.Key() {
	case tcell.KeyRune:
		// Insert character at cursor position
		if t.maxLength > 0 && len(t.buffer) >= t.maxLength {
			// Max length reached
			return true
		}

		// Insert rune at cursor position
		r := keyEvent.Rune()
		if t.cursorPos == len(t.buffer) {
			t.buffer = append(t.buffer, r)
		} else {
			t.buffer = append(t.buffer[:t.cursorPos], append([]rune{r}, t.buffer[t.cursorPos:]...)...)
		}
		t.cursorPos++
		contentChanged = true

	case tcell.KeyDelete:
		// Delete character after cursor
		if t.cursorPos < len(t.buffer) {
			t.buffer = append(t.buffer[:t.cursorPos], t.buffer[t.cursorPos+1:]...)
			contentChanged = true
		}

	case tcell.KeyBackspace, tcell.KeyBackspace2:
		// Delete character before cursor
		if t.cursorPos > 0 {
			t.buffer = append(t.buffer[:t.cursorPos-1], t.buffer[t.cursorPos:]...)
			t.cursorPos--
			contentChanged = true
		}

	case tcell.KeyLeft:
		// Move cursor left
		if t.cursorPos > 0 {
			t.cursorPos--
			t.MarkDirty() // Need redraw for cursor movement
		}

	case tcell.KeyRight:
		// Move cursor right
		if t.cursorPos < len(t.buffer) {
			t.cursorPos++
			t.MarkDirty() // Need redraw for cursor movement
		}

	case tcell.KeyHome:
		// Move cursor to beginning
		if t.cursorPos != 0 {
			t.cursorPos = 0
			t.MarkDirty() // Need redraw for cursor movement
		}

	case tcell.KeyEnd:
		// Move cursor to end
		if t.cursorPos != len(t.buffer) {
			t.cursorPos = len(t.buffer)
			t.MarkDirty() // Need redraw for cursor movement
		}

	case tcell.KeyEnter:
		// Submit the content
		text := string(t.buffer)
		onSubmit := t.onSubmit

		if onSubmit != nil {
			onSubmit(text)
		}
		return true

	default:
		// Unhandled key
		return false
	}

	if contentChanged {
		t.updateVisualOffset()
		t.MarkDirty()

		// Trigger onChange handler
		text := string(t.buffer)
		onChange := t.onChange

		if onChange != nil {
			onChange(text)
		}
	}

	return true
}

// runesEqual compares two rune slices for equality.
func runesEqual(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}