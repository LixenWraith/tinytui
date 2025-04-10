// widgets/button.go
package widgets

import (
	"sync"

	"github.com/LixenWraith/tinytui"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// IndicatorPosition specifies where the focus/action indicator appears on a Button.
type IndicatorPosition int

const (
	// IndicatorNone means no indicator is shown.
	IndicatorNone IndicatorPosition = iota
	// IndicatorLeft places the indicator to the left of the label.
	IndicatorLeft
	// IndicatorRight places the indicator to the right of the label.
	IndicatorRight
)

// Button is a focusable widget that displays a label and triggers an action.
type Button struct {
	tinytui.BaseWidget
	mu           sync.RWMutex
	label        string
	style        tinytui.Style
	focusedStyle tinytui.Style
	indicator    rune // Character used as the indicator (e.g., '>', 0 for none).
	indicatorPos IndicatorPosition
	onClick      func() // Action to perform when activated
}

// NewButton creates a new Button widget.
func NewButton(label string) *Button {
	b := &Button{
		label:        label,
		style:        tinytui.DefaultStyle,
		focusedStyle: tinytui.DefaultStyle.Reverse(true), // Default focus: reverse video
		indicator:    '>',                                // Default indicator
		indicatorPos: IndicatorLeft,                      // Default position
		onClick:      nil,
	}
	b.SetVisible(true) // Buttons are always visible on creation
	return b
}

// SetLabel updates the text displayed on the button.
func (b *Button) SetLabel(label string) *Button {
	b.mu.Lock()
	b.label = label
	b.mu.Unlock()
	if app := b.App(); app != nil {
		app.QueueRedraw()
	}
	return b
}

// SetStyle sets the base style (color, attributes) used when the button is not focused.
func (b *Button) SetStyle(style tinytui.Style) *Button {
	b.mu.Lock()
	b.style = style
	b.mu.Unlock()
	if !b.IsFocused() && b.App() != nil {
		b.App().QueueRedraw()
	}
	return b
}

// SetFocusedStyle sets the style used when the button has focus.
func (b *Button) SetFocusedStyle(style tinytui.Style) *Button {
	b.mu.Lock()
	b.focusedStyle = style
	b.mu.Unlock()
	if b.IsFocused() && b.App() != nil {
		b.App().QueueRedraw()
	}
	return b
}

// SetIndicator configures the focus/action indicator character and its position.
// Set indicator to 0 or position to IndicatorNone to disable it.
func (b *Button) SetIndicator(indicator rune, position IndicatorPosition) *Button {
	b.mu.Lock()
	b.indicator = indicator
	b.indicatorPos = position
	b.mu.Unlock()
	if app := b.App(); app != nil {
		app.QueueRedraw()
	}
	return b
}

// SetOnClick sets the function to be called when the button is activated (e.g., by pressing Enter).
func (b *Button) SetOnClick(handler func()) *Button {
	b.mu.Lock()
	b.onClick = handler
	b.mu.Unlock()
	// No redraw needed for changing the handler
	return b
}

// Draw draws the button.
func (b *Button) Draw(screen tcell.Screen) {
	x, y, width, height := b.GetRect()
	if width <= 0 || height <= 0 {
		return
	}

	b.mu.RLock() // Read lock for accessing properties
	currentStyle := b.style
	// Show indicator only when focused and configured
	showIndicator := b.indicator != 0 && b.indicatorPos != IndicatorNone && b.IsFocused()
	indicatorChar := b.indicator
	indicatorPos := b.indicatorPos
	labelText := b.label
	if b.IsFocused() {
		currentStyle = b.focusedStyle
	}
	b.mu.RUnlock() // Release lock

	// --- Calculate Layout ---
	indicatorWidth := 0
	indicatorX := -1 // Screen X coordinate for indicator
	labelStartX := x // Screen X coordinate for label start
	availableWidth := width

	if showIndicator {
		indicatorWidth = runewidth.RuneWidth(indicatorChar)
		if indicatorPos == IndicatorLeft {
			// Ensure space for indicator AND a space after it
			if indicatorWidth+1 < availableWidth {
				indicatorX = x
				labelStartX = x + indicatorWidth + 1
				availableWidth -= indicatorWidth + 1
			} else {
				showIndicator = false // Not enough space
			}
		} else { // IndicatorRight
			// Ensure space for indicator AND a space before it
			if indicatorWidth+1 < availableWidth {
				indicatorX = x + width - indicatorWidth
				availableWidth -= indicatorWidth + 1
			} else {
				showIndicator = false // Not enough space
			}
		}
	}

	// Center the label text within the available space
	labelWidth := runewidth.StringWidth(labelText)
	if labelWidth > availableWidth {
		// TODO: Truncate label if needed, maybe add "..."
		labelWidth = availableWidth // Use the available width for centering calculation
	}

	if availableWidth > 0 {
		labelStartX += (availableWidth - labelWidth) / 2 // Center alignment
	} else {
		labelStartX = x // Fallback if no space
	}
	if labelStartX < x {
		labelStartX = x // Ensure it doesn't go left of the button start
	}
	// Ensure label doesn't overlap left indicator+space
	if indicatorPos == IndicatorLeft && showIndicator && labelStartX < x+indicatorWidth+1 {
		labelStartX = x + indicatorWidth + 1
	}

	// --- Drawing ---

	// 1. Fill background
	tinytui.Fill(screen, x, y, width, height, ' ', currentStyle)

	// 2. Draw Indicator (if shown and space permits)
	if showIndicator && indicatorX >= x && indicatorX+indicatorWidth <= x+width {
		// Use the exported ToTcell() method here
		screen.SetContent(indicatorX, y, indicatorChar, nil, currentStyle.ToTcell())
	}

	// 3. Draw Label (clipped)
	col := labelStartX
	sw, _ := screen.Size() // Screen width for boundary check
	endCol := x + width    // Default right boundary for text

	// Adjust end column if right indicator is shown
	if indicatorPos == IndicatorRight && showIndicator {
		// Stop drawing before the space preceding the indicator
		endCol = indicatorX - 1
	}

	// Draw the label character by character, respecting rune width and clipping
	tcellStyle := currentStyle.ToTcell() // Convert style once for drawing loop
	for _, r := range labelText {
		rw := runewidth.RuneWidth(r)
		// Check bounds: ensure rune fits before endCol and within screen width
		if col+rw > endCol || col >= sw {
			break
		}
		// Only draw if within button horizontal bounds (col >= x)
		if col >= x {
			screen.SetContent(col, y, r, nil, tcellStyle)
		}
		col += rw
	}
}

// Focusable indicates that Buttons can receive focus.
func (b *Button) Focusable() bool {
	return true
}

// Focus is called when the button gains focus.
// BaseWidget handles the state change and redraw request.
func (b *Button) Focus() {
	b.BaseWidget.Focus()
}

// Blur is called when the button loses focus.
// BaseWidget handles the state change and redraw request.
func (b *Button) Blur() {
	b.BaseWidget.Blur()
}

// HandleEvent handles input events for the Button.
func (b *Button) HandleEvent(event tcell.Event) bool {
	// Check base widget bindings first (allows overriding default behavior)
	if b.BaseWidget.HandleEvent(event) {
		return true
	}

	// Buttons only react when focused
	if !b.IsFocused() {
		return false
	}

	b.mu.RLock()
	clickHandler := b.onClick
	b.mu.RUnlock()

	// Handle activation keys (Enter)
	if keyEvent, ok := event.(*tcell.EventKey); ok {
		if keyEvent.Key() == tcell.KeyEnter {
			if clickHandler != nil {
				clickHandler()
			}
			return true // Enter key consumed
		}
	}

	// Handle mouse clicks (Phase 4 - Placeholder)
	// ... (mouse handling code remains commented out) ...

	return false // Event not handled
}