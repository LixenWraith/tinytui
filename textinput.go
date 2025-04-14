// textinput.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// TextInput provides a single-line text entry field with cursor navigation,
// editing capabilities (insert, delete, backspace), optional masking for passwords,
// and optional maximum length enforcement. It is focusable and interactive.
type TextInput struct {
	BaseComponent
	buffer       []rune       // Stores the text content as runes for correct indexing.
	cursorPos    int          // Cursor position as a rune index within the buffer [0, len(buffer)].
	visualOffset int          // Rune index of the start of the visible portion of the buffer (for horizontal scrolling).
	style        Style        // Base style for the input field when not focused.
	focusedStyle Style        // Style when the input field has focus.
	maxLength    int          // Maximum number of runes allowed (0 for no limit).
	onChange     func(string) // Callback function triggered when text content changes.
	onSubmit     func(string) // Callback function triggered when Enter key is pressed.
	masked       bool         // Display mask characters instead of actual text?
	maskRune     rune         // Rune to use for masking (e.g., '*').
}

// NewTextInput creates a new text input component.
// Initializes styles from the current theme.
func NewTextInput() *TextInput {
	theme := GetTheme()
	if theme == nil {
		theme = NewDefaultTheme()
	} // Fallback

	t := &TextInput{
		BaseComponent: NewBaseComponent(),
		buffer:        []rune{},
		cursorPos:     0,
		visualOffset:  0,
		style:         theme.TextStyle(),               // Base style from theme
		focusedStyle:  theme.TextStyle().Reverse(true), // Focused style: typically reverse base
		maxLength:     0,                               // No limit by default
		masked:        false,
		maskRune:      '*',
		// onChange, onSubmit are nil initially
	}
	t.ApplyTheme(theme) // Ensure initial theme application correctly sets styles
	return t
}

// ApplyTheme updates the text input's styles based on the provided theme.
// Implements ThemedComponent.
func (t *TextInput) ApplyTheme(theme Theme) {
	if theme == nil {
		return
	}
	newStyle := theme.TextStyle()
	newFocusedStyle := newStyle.Reverse(true) // Derive focused style from theme's text style

	changed := false
	if t.style != newStyle {
		t.style = newStyle
		changed = true
	}
	if t.focusedStyle != newFocusedStyle {
		t.focusedStyle = newFocusedStyle
		changed = true
	}
	if changed {
		t.MarkDirty() // Mark dirty only if styles actually changed
	}
}

// SetText replaces the current text content with the given string.
// Enforces maximum length and moves the cursor to the end.
func (t *TextInput) SetText(text string) {
	newBuffer := []rune(text)

	// Enforce maxLength if set
	if t.maxLength > 0 && len(newBuffer) > t.maxLength {
		newBuffer = newBuffer[:t.maxLength]
	}

	currentText := string(t.buffer)
	newText := string(newBuffer)

	// Only update if text actually changed
	if currentText == newText {
		// If text is same, ensure cursor is still valid (might be needed if called after external change?)
		if t.cursorPos > len(t.buffer) {
			t.cursorPos = len(t.buffer)
		}
		t.updateVisualOffset() // Still might need scroll adjustment
		return
	}

	t.buffer = newBuffer
	t.cursorPos = len(t.buffer) // Move cursor to the end
	t.visualOffset = 0          // Reset scroll
	t.updateVisualOffset()      // Adjust scroll if new end position requires it
	t.MarkDirty()

	// Trigger change handler if text content changed
	if t.onChange != nil {
		t.onChange(newText)
	}
}

// SetContent is an alias for SetText to implement the TextUpdater interface.
func (t *TextInput) SetContent(text string) {
	t.SetText(text)
}

// GetText returns the current text content as a string.
func (t *TextInput) GetText() string {
	// Return empty string if buffer is nil? Should not happen with NewTextInput.
	if t.buffer == nil {
		return ""
	}
	return string(t.buffer)
}

// SetStyle explicitly sets the base (unfocused) style, overriding the theme.
// Consider using themes for consistent styling.
func (t *TextInput) SetStyle(style Style) {
	if t.style != style {
		t.style = style
		// Should this also update focusedStyle based on the new base style?
		// t.focusedStyle = style.Reverse(true)? Let's require explicit SetFocusedStyle.
		t.MarkDirty()
	}
}

// SetFocusedStyle explicitly sets the focused style, overriding the theme-derived default.
func (t *TextInput) SetFocusedStyle(style Style) {
	if t.focusedStyle != style {
		t.focusedStyle = style
		t.MarkDirty()
	}
}

// SetMaxLength sets the maximum number of runes allowed in the input.
// Truncates existing text and adjusts cursor if the new limit is smaller.
// Setting max to 0 disables the length limit.
func (t *TextInput) SetMaxLength(max int) {
	if max < 0 {
		max = 0
	} // Ensure non-negative limit

	if t.maxLength == max {
		return
	} // No change

	t.maxLength = max
	truncated := false

	// If current text exceeds new limit, truncate it
	if max > 0 && len(t.buffer) > max {
		t.buffer = t.buffer[:max]
		truncated = true
		// Adjust cursor if it was beyond the new max length
		if t.cursorPos > max {
			t.cursorPos = max
		}
		// Truncation might require scroll adjustment
		t.updateVisualOffset()
		t.MarkDirty()
	}

	// Trigger change handler if text was actually truncated
	if truncated && t.onChange != nil {
		t.onChange(string(t.buffer))
	}
}

// SetMasked enables or disables password-style masking using the specified rune.
func (t *TextInput) SetMasked(masked bool, maskRune rune) {
	// Use default mask rune '*' if an invalid one (like 0) is provided
	if maskRune == 0 {
		maskRune = '*'
	}

	// Check if state is actually changing
	if t.masked == masked && (!masked || t.maskRune == maskRune) {
		return // No change
	}

	t.masked = masked
	if masked { // Only update maskRune if masking is enabled
		t.maskRune = maskRune
	}

	t.MarkDirty() // Appearance changes, needs redraw
}

// SetOnChange sets the callback function triggered whenever the text content changes due to user input.
func (t *TextInput) SetOnChange(handler func(string)) {
	t.onChange = handler
}

// SetOnSubmit sets the callback function triggered when the Enter key is pressed within the input field.
func (t *TextInput) SetOnSubmit(handler func(string)) {
	t.onSubmit = handler
}

// Focusable returns true if the component is visible, indicating it can receive input focus.
func (t *TextInput) Focusable() bool {
	return t.IsVisible()
}

// Draw renders the text input component, including text (masked or not), and requests cursor position.
func (t *TextInput) Draw(screen tcell.Screen) {
	if !t.IsVisible() {
		return
	}

	x, y, width, height := t.GetRect()
	// TextInput only uses a single line, height is ignored beyond clearing.
	if width <= 0 || height <= 0 {
		return
	}

	// Select style based on focus state
	currentStyle := t.style
	if t.IsFocused() {
		currentStyle = t.focusedStyle
	}

	// Clear the component area (typically just one line high)
	Fill(screen, x, y, width, height, ' ', currentStyle)

	// Determine text runes to display (apply masking if enabled)
	displayRunes := t.buffer
	if t.masked {
		displayRunes = make([]rune, len(t.buffer))
		for i := range displayRunes {
			displayRunes[i] = t.maskRune
		}
	}

	// Ensure visual offset keeps cursor visible before getting visible text
	t.updateVisualOffset()

	// Get the portion of text runes that fits within the component width
	visibleRunes := t.getVisibleRunes(displayRunes, width)
	visibleText := string(visibleRunes)

	// Draw the visible text onto the screen
	DrawText(screen, x, y, currentStyle, visibleText)

	// If focused, calculate and request the cursor position
	if t.IsFocused() {
		// Calculate cursor screen position (X coordinate) based on the width of runes
		// *before* the cursor *within the visible portion*.
		cursorScreenX := x
		// Find the cursor's index relative to the start of the visible runes
		cursorIndexInVisible := t.cursorPos - t.visualOffset
		// Ensure the relative index is within the bounds of the visible runes slice
		if cursorIndexInVisible >= 0 && cursorIndexInVisible <= len(visibleRunes) {
			// Calculate width of runes from start of visible portion up to the cursor index
			cursorScreenX = x + runewidth.StringWidth(string(visibleRunes[:cursorIndexInVisible]))
		} else if cursorIndexInVisible < 0 {
			// Cursor is before the visible part (shouldn't happen after updateVisualOffset)
			cursorScreenX = x // Place at start
		} else { // cursorIndexInVisible > len(visibleRunes)
			// Cursor is after the visible part (shouldn't happen)
			cursorScreenX = x + runewidth.StringWidth(visibleText) // Place at end
		}

		// Ensure cursor position doesn't exceed component width
		if cursorScreenX >= x+width {
			cursorScreenX = x + width - 1
		}
		if cursorScreenX < x {
			cursorScreenX = x
		}

		// Request cursor manager to show cursor at calculated position
		if app := t.App(); app != nil {
			if cm := app.GetCursorManager(); cm != nil {
				cm.Request(cursorScreenX, y)
			}
		}
	}
}

// getVisibleRunes calculates the slice of runes that should be visible
// based on the current visualOffset and available component width.
func (t *TextInput) getVisibleRunes(runes []rune, maxWidth int) []rune {
	totalRunes := len(runes)
	if totalRunes == 0 || maxWidth <= 0 || t.visualOffset >= totalRunes {
		return []rune{} // Nothing to display
	}

	availableWidth := maxWidth
	startIndex := t.visualOffset
	endIndex := startIndex // Exclusive end index

	// Iterate from start index, accumulating width until maxWidth is reached or runes end
	for endIndex < totalRunes {
		runeWidth := runewidth.RuneWidth(runes[endIndex])
		if availableWidth < runeWidth {
			break // Next rune doesn't fit
		}
		availableWidth -= runeWidth
		endIndex++
	}

	// Return the slice from startIndex up to (but not including) endIndex
	return runes[startIndex:endIndex]
}

// updateVisualOffset adjusts the visualOffset (horizontal scroll position)
// to ensure the cursor is always visible within the component's width.
func (t *TextInput) updateVisualOffset() {
	// Ensure cursor position is valid first
	if t.cursorPos < 0 {
		t.cursorPos = 0
	}
	if t.cursorPos > len(t.buffer) {
		t.cursorPos = len(t.buffer)
	}

	width := t.rect.Width // Get current component width
	if width <= 0 {
		t.visualOffset = 0 // Cannot determine visibility if width is unknown
		return
	}

	// --- Check if cursor is outside the current view [visualOffset, visualOffset + width) ---

	// Case 1: Cursor is to the left of the visible area (cursorPos < visualOffset)
	if t.cursorPos < t.visualOffset {
		t.visualOffset = t.cursorPos // Scroll left so cursor is the first visible character
		return
	}

	// Case 2: Cursor is potentially to the right of the visible area
	// Calculate the visual width required to display runes from visualOffset up to cursorPos
	widthToCursor := 0
	if t.visualOffset <= t.cursorPos && t.visualOffset < len(t.buffer) {
		// Iterate runes from visualOffset up to (but not including) cursorPos
		for i := t.visualOffset; i < t.cursorPos; i++ {
			if i < len(t.buffer) { // Check buffer bounds
				widthToCursor += runewidth.RuneWidth(t.buffer[i])
			} else {
				break
			} // Should not happen if cursorPos is valid
		}
	}

	// If width needed >= component width, cursor is at or past the right edge, need to scroll right.
	// We want the cursor to be the *last* fully visible character, or just inside the right edge.
	if widthToCursor >= width {
		// Start potential new offset at the cursor position and move leftwards,
		// accumulating width until we have just enough runes to fill the width.
		newOffset := t.cursorPos
		accumulatedWidth := 0
		for newOffset > 0 {
			prevRuneIndex := newOffset - 1
			runeW := runewidth.RuneWidth(t.buffer[prevRuneIndex])
			// If adding this rune makes it too wide, the current newOffset is correct.
			if accumulatedWidth+runeW >= width {
				break
			}
			accumulatedWidth += runeW
			newOffset-- // Move potential start position left
		}

		// Ensure offset is not negative
		if newOffset < 0 {
			newOffset = 0
		}

		t.visualOffset = newOffset
	}
	// Case 3: Cursor is already within the visible area [visualOffset, visualOffset + width)
	// No change needed in visualOffset.
}

// HandleEvent processes key events for text input manipulation (insert, delete, backspace),
// cursor movement (arrows, home, end), and submission (Enter).
func (t *TextInput) HandleEvent(event tcell.Event) bool {
	keyEvent, ok := event.(*tcell.EventKey)
	if !ok {
		return false // Not a key event
	}

	textBefore := string(t.buffer) // Store state before modification for onChange check
	contentChanged := false
	cursorMoved := false

	switch keyEvent.Key() {
	// --- Character Input ---
	case tcell.KeyRune:
		// Check max length before inserting rune
		if t.maxLength > 0 && len(t.buffer) >= t.maxLength {
			return true // Max length reached, consume event but do nothing
		}
		r := keyEvent.Rune()
		// Insert rune at cursor position using slice manipulation
		t.buffer = append(t.buffer[:t.cursorPos], append([]rune{r}, t.buffer[t.cursorPos:]...)...)
		t.cursorPos++ // Move cursor after inserted rune
		contentChanged = true

	// --- Deletion ---
	case tcell.KeyDelete: // Delete character *after* cursor (at cursor index)
		if t.cursorPos < len(t.buffer) { // Only if cursor is not at the very end
			t.buffer = append(t.buffer[:t.cursorPos], t.buffer[t.cursorPos+1:]...)
			contentChanged = true
			// Cursor position does not change relative to remaining text before it
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2: // Delete character *before* cursor
		if t.cursorPos > 0 { // Only if cursor is not at the very beginning
			t.buffer = append(t.buffer[:t.cursorPos-1], t.buffer[t.cursorPos:]...)
			t.cursorPos-- // Move cursor back
			contentChanged = true
		}

	// --- Cursor Movement ---
	case tcell.KeyLeft:
		if t.cursorPos > 0 {
			t.cursorPos--
			cursorMoved = true
		}
	case tcell.KeyRight:
		if t.cursorPos < len(t.buffer) {
			t.cursorPos++
			cursorMoved = true
		}
	case tcell.KeyHome, tcell.KeyCtrlA: // Treat Ctrl+A like Home
		if t.cursorPos != 0 {
			t.cursorPos = 0
			cursorMoved = true
		}
	case tcell.KeyEnd, tcell.KeyCtrlE: // Treat Ctrl+E like End
		if t.cursorPos != len(t.buffer) {
			t.cursorPos = len(t.buffer)
			cursorMoved = true
		}
	// TODO: Add Ctrl+Left/Right for word navigation? Requires word boundary detection.
	// TODO: Add Ctrl+U to delete line before cursor? Ctrl+K delete after?

	// --- Submission ---
	case tcell.KeyEnter:
		// Trigger the onSubmit callback if it's set
		if t.onSubmit != nil {
			t.onSubmit(string(t.buffer))
		}
		return true // Event handled (submission)

	// --- Unhandled Keys ---
	default:
		// Ignore other keys (like Shift, Alt, Ctrl by themselves, function keys, etc.)
		return false // Indicate key was not handled by text input
	}

	// --- Post-Action Updates (if event was handled) ---
	if contentChanged || cursorMoved {
		// Ensure cursor visibility after any change
		t.updateVisualOffset()
		// Mark dirty to redraw the text and potentially the cursor position
		t.MarkDirty()
	}

	// Trigger onChange callback if content actually changed
	if contentChanged && t.onChange != nil {
		newText := string(t.buffer)
		// Sanity check: ensure text actually differs from before the event
		if textBefore != newText {
			t.onChange(newText)
		}
	}

	// If we reached here, the key event was processed (input, deletion, movement)
	return true
}