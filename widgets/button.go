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
		style:        tinytui.DefaultButtonStyle(),
		focusedStyle: tinytui.DefaultButtonFocusedStyle(),
		indicator:    '>',           // Default indicator
		indicatorPos: IndicatorLeft, // Default position
		onClick:      nil,
	}
	b.SetVisible(true) // Explicitly set visibility
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

// ApplyTheme applies the provided theme to the Button widget
func (b *Button) ApplyTheme(theme tinytui.Theme) {
	b.SetStyle(theme.ButtonStyle())
	b.SetFocusedStyle(theme.ButtonFocusedStyle())
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

func (b *Button) PreferredWidth() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	label := b.label

	// Base width: label plus some padding (2 spaces on each side)
	width := runewidth.StringWidth(label) + 4

	// Consider indicators if enabled
	if b.indicator != 0 && b.indicatorPos != IndicatorNone {
		indicatorWidth := runewidth.RuneWidth(b.indicator)
		width += indicatorWidth + 1 // Add indicator width + space
	}

	return width
}

func (b *Button) PreferredHeight() int {
	// Buttons typically are 1 line high
	return 1
}

// Draw draws the button.
func (b *Button) Draw(screen tcell.Screen) {
	b.BaseWidget.Draw(screen)

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

	// IMPORTANT: Always fill the entire button background first
	tinytui.Fill(screen, x, y, width, height, ' ', currentStyle)

	// Calculate vertical center for text alignment
	textY := y
	if height > 1 {
		// Center text vertically when button has height > 1
		textY = y + (height / 2)
	}

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
		// Truncate label if needed
		labelText = runewidth.Truncate(labelText, availableWidth, "")
		labelWidth = runewidth.StringWidth(labelText)
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
	// Draw the label (clipped) - always draw label even if very small
	if labelWidth > 0 {
		col := labelStartX
		endCol := x + width // Default right boundary for text

		// Adjust end column if right indicator is shown
		if indicatorPos == IndicatorRight && showIndicator {
			// Stop drawing before the space preceding the indicator
			endCol = indicatorX - 1
		}

		// Calculate available width for the label
		availableDisplayWidth := endCol - col
		if availableDisplayWidth <= 0 {
			availableDisplayWidth = 1 // Ensure at least 1 character can be displayed
		}

		// Truncate text to available width
		displayText := runewidth.Truncate(labelText, availableDisplayWidth, "")

		// Draw the label at vertical center when height > 1
		tinytui.DrawText(screen, col, textY, currentStyle, displayText)
	}

	// Draw Indicator (if shown and space permits)
	if showIndicator && indicatorX >= x && indicatorX+indicatorWidth <= x+width {
		// Use the exported ToTcell() method here
		screen.SetContent(indicatorX, y, indicatorChar, nil, currentStyle.ToTcell())
	}
}

// Focusable indicates that Buttons can receive focus.
func (b *Button) Focusable() bool {
	if !b.IsVisible() {
		return false
	}
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

	return false // Event not handled
}