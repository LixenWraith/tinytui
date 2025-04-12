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
	mu                     sync.RWMutex
	label                  string
	style                  tinytui.Style // Normal style
	selectedStyle          tinytui.Style // Selected, not focused
	interactedStyle        tinytui.Style // Interacted, not focused
	focusedStyle           tinytui.Style // Focused, normal state
	focusedSelectedStyle   tinytui.Style // Focused and selected
	focusedInteractedStyle tinytui.Style // Focused and interacted
	indicator              rune          // Character used as the indicator (e.g., '>', 0 for none)
	indicatorPos           IndicatorPosition
	onClick                func() // Action to perform when activated
}

// NewButton creates a new Button widget.
func NewButton(label string) *Button {
	b := &Button{
		label:                  label,
		style:                  tinytui.DefaultButtonStyle(),
		selectedStyle:          tinytui.DefaultButtonStyle().Dim(true).Underline(true),
		interactedStyle:        tinytui.DefaultButtonStyle().Bold(true),
		focusedStyle:           tinytui.DefaultButtonFocusedStyle(),
		focusedSelectedStyle:   tinytui.DefaultButtonFocusedStyle().Dim(true),
		focusedInteractedStyle: tinytui.DefaultButtonFocusedStyle().Bold(true),
		indicator:              '>',           // Default indicator
		indicatorPos:           IndicatorLeft, // Default position
		onClick:                nil,
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
	if app := b.App(); app != nil {
		app.QueueRedraw()
	}
	return b
}

func (b *Button) SetSelectedStyle(style tinytui.Style) *Button {
	b.mu.Lock()
	b.selectedStyle = style
	b.mu.Unlock()
	if app := b.App(); app != nil {
		app.QueueRedraw()
	}
	return b
}

func (b *Button) SetInteractedStyle(style tinytui.Style) *Button {
	b.mu.Lock()
	b.interactedStyle = style
	b.mu.Unlock()
	if app := b.App(); app != nil {
		app.QueueRedraw()
	}
	return b
}

func (b *Button) SetFocusedStyle(style tinytui.Style) *Button {
	b.mu.Lock()
	b.focusedStyle = style
	b.mu.Unlock()
	if app := b.App(); app != nil {
		app.QueueRedraw()
	}
	return b
}

func (b *Button) SetFocusedSelectedStyle(style tinytui.Style) *Button {
	b.mu.Lock()
	b.focusedSelectedStyle = style
	b.mu.Unlock()
	if app := b.App(); app != nil {
		app.QueueRedraw()
	}
	return b
}

func (b *Button) SetFocusedInteractedStyle(style tinytui.Style) *Button {
	b.mu.Lock()
	b.focusedInteractedStyle = style
	b.mu.Unlock()
	if app := b.App(); app != nil {
		app.QueueRedraw()
	}
	return b
}

// ApplyTheme applies the provided theme to the Button widget
func (b *Button) ApplyTheme(theme tinytui.Theme) {
	b.SetStyle(theme.ButtonStyle())
	b.SetSelectedStyle(theme.ButtonSelectedStyle())
	b.SetInteractedStyle(theme.ButtonInteractedStyle())
	b.SetFocusedStyle(theme.ButtonFocusedStyle())
	b.SetFocusedSelectedStyle(theme.ButtonFocusedSelectedStyle())
	b.SetFocusedInteractedStyle(theme.ButtonFocusedInteractedStyle())
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

	// Determine appropriate style based on focus and state
	currentStyle := b.style
	state := b.GetState()
	isFocused := b.IsFocused()

	if isFocused {
		switch state {
		case tinytui.StateInteracted:
			currentStyle = b.focusedInteractedStyle
		case tinytui.StateSelected:
			currentStyle = b.focusedSelectedStyle
		default:
			currentStyle = b.focusedStyle
		}
	} else {
		switch state {
		case tinytui.StateInteracted:
			currentStyle = b.interactedStyle
		case tinytui.StateSelected:
			currentStyle = b.selectedStyle
		}
	}

	// Remaining properties
	showIndicator := b.indicator != 0 && b.indicatorPos != IndicatorNone && isFocused
	indicatorChar := b.indicator
	indicatorPos := b.indicatorPos
	labelText := b.label

	b.mu.RUnlock() // Release lock

	// Extract colors for background fill (without attributes)
	fg, bg, _, _ := currentStyle.Deconstruct()
	fillStyle := tinytui.DefaultStyle.Foreground(fg).Background(bg)

	// IMPORTANT: Always fill the entire button background first
	// Use fillStyle (without attributes) to avoid extending effects like underline
	tinytui.Fill(screen, x, y, width, height, ' ', fillStyle)

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

	// Add a visual button boundary to make it more visible
	if width > 2 && height > 0 {
		// Draw button outline with full style (including attributes)
		for i := 0; i < width; i++ {
			// Top border
			screen.SetContent(x+i, y, '─', nil, currentStyle.ToTcell())
			// Bottom border if height allows
			if height > 1 {
				screen.SetContent(x+i, y+height-1, '─', nil, currentStyle.ToTcell())
			}
		}
		// Side borders if width allows
		if height > 2 {
			for i := 1; i < height-1; i++ {
				screen.SetContent(x, y+i, '│', nil, currentStyle.ToTcell())
				screen.SetContent(x+width-1, y+i, '│', nil, currentStyle.ToTcell())
			}
		}

		// Corners if both width and height allow
		if width > 1 && height > 1 {
			screen.SetContent(x, y, '┌', nil, currentStyle.ToTcell())
			screen.SetContent(x+width-1, y, '┐', nil, currentStyle.ToTcell())
			screen.SetContent(x, y+height-1, '└', nil, currentStyle.ToTcell())
			screen.SetContent(x+width-1, y+height-1, '┘', nil, currentStyle.ToTcell())
		}
	}

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

		// Draw the label with full style (including attributes)
		tinytui.DrawText(screen, col, textY, currentStyle, displayText)
	}

	// Draw Indicator (if shown and space permits)
	if showIndicator && indicatorX >= x && indicatorX+indicatorWidth <= x+width {
		// Use the exported ToTcell() method here with full style
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
			// Set state to interacted
			b.SetState(tinytui.StateInteracted)

			// Trigger callback if set
			if clickHandler != nil {
				clickHandler()
			}

			// Note: We keep the interacted state after clicking
			// Optionally, we could reset it after a delay or leave it to the app logic

			return true // Enter key consumed

		} else if keyEvent.Key() == tcell.KeyRune {
			if keyEvent.Rune() == ' ' {
				// Space selects but doesn't activate
				currentState := b.GetState()
				if currentState != tinytui.StateSelected {
					b.SetState(tinytui.StateSelected)
				} else {
					// Toggle selection off if already selected
					b.SetState(tinytui.StateNormal)
				}
				return true // Space key consumed
			}
		}
	}

	return false // Event not handled
}