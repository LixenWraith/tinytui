// cursor.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
	"time"
)

// CursorManager handles the visibility, position, and blinking of the terminal cursor,
// typically controlled by input components like TextInput. It ensures only one
// cursor is active and manages its blinking cycle independently.
type CursorManager struct {
	screen tcell.Screen // The application screen to draw the cursor on
	app    *Application // Reference to the application for queuing redraws

	// Cursor state for the current frame
	requestedX   int  // Requested X position (column) for this frame
	requestedY   int  // Requested Y position (row) for this frame
	requestValid bool // Was Request() called during the current draw cycle?

	// Blinking behavior
	blinkRate  time.Duration // Duration between blink state changes
	blinkState bool          // Is the cursor currently visible in the blink cycle?
	blinkTimer *time.Ticker  // Timer driving the blink cycle
	stopBlink  chan struct{} // Channel to signal the blink loop to stop gracefully
}

// NewCursorManager creates and starts a cursor manager associated with an application and screen.
func NewCursorManager(app *Application, screen tcell.Screen, rate time.Duration) *CursorManager {
	if rate <= 0 {
		rate = 500 * time.Millisecond // Default blink rate if invalid
	}

	cm := &CursorManager{
		app:        app,
		screen:     screen,
		blinkRate:  rate,
		blinkState: true, // Start with cursor visible
		stopBlink:  make(chan struct{}),
	}

	// Start the blink timer and the goroutine that manages blinking
	cm.blinkTimer = time.NewTicker(rate)
	go cm.blinkLoop()

	return cm
}

// Request sets the desired cursor position for the *current* draw frame.
// This should be called only once per frame, typically by the focused input component
// during its Draw() method. If called multiple times, the last call wins.
func (cm *CursorManager) Request(x, y int) {
	// Store the requested position and mark that a request was made for this frame.
	cm.requestedX = x
	cm.requestedY = y
	cm.requestValid = true
}

// ResetForFrame clears the cursor request state at the beginning of a draw cycle.
// This is called by the Application before drawing components to ensure no stale request persists.
func (cm *CursorManager) ResetForFrame() {
	cm.requestValid = false
}

// Draw renders the cursor on the screen based on the current frame's request and blink state.
// This is called by the Application *after* all components have been drawn.
func (cm *CursorManager) Draw() {
	// Determine if the cursor should be shown: request must be valid and blink state must be visible.
	shouldShow := cm.requestValid && cm.blinkState

	if shouldShow {
		// Show the terminal cursor at the requested position.
		cm.screen.ShowCursor(cm.requestedX, cm.requestedY)
	} else {
		// Hide the terminal cursor if not requested or if blinked off.
		cm.screen.HideCursor()
	}
}

// Stop halts the blinking timer goroutine and cleans up associated resources.
// Should be called when the application shuts down.
func (cm *CursorManager) Stop() {
	// Stop the timer first to prevent further ticks
	if cm.blinkTimer != nil {
		cm.blinkTimer.Stop()
	}
	// Signal the blinkLoop goroutine to exit using the stop channel
	// Use a non-blocking send or check if already closed for safety if Stop can be called multiple times.
	select {
	case <-cm.stopBlink:
		// Already closed
	default:
		close(cm.stopBlink)
	}
	cm.blinkTimer = nil // Release timer reference
}

// blinkLoop is the goroutine that toggles the cursor visibility state periodically
// and queues application redraws to reflect the change.
func (cm *CursorManager) blinkLoop() {
	defer func() {
		// Ensure cursor is hidden when blink loop stops (e.g., during shutdown)
		// Check screen still exists as Fini might have been called.
		if cm.screen != nil {
			// Use screen directly, Draw might not be called again
			cm.screen.HideCursor()
			// We might need a final Show() to ensure the hidden cursor state is flushed,
			// but shutdown() usually handles screen state restoration.
			// cm.screen.Show()
		}
	}()

	for {
		select {
		case <-cm.blinkTimer.C:
			// Timer ticked, toggle blink state
			cm.blinkState = !cm.blinkState

			// If a cursor was requested in the *previous* frame (and thus potentially visible),
			// queue a redraw now to reflect the blink change.
			// This check prevents unnecessary redraws if the cursor wasn't shown anyway.
			if cm.app != nil && cm.requestValid {
				cm.app.QueueRedraw()
			}
		case <-cm.stopBlink:
			// Stop signal received, exit the goroutine
			return
		}
	}
}

// SetBlinkRate changes the cursor blink rate dynamically.
// Note: Dynamically changing the rate while running requires careful handling
// of the timer and goroutine restart. This implementation assumes it's called
// infrequently or when the application is stable.
func (cm *CursorManager) SetBlinkRate(rate time.Duration) {
	if rate <= 0 || cm.blinkRate == rate {
		return // Ignore invalid rate or no change
	}

	cm.blinkRate = rate

	// Reset the timer with the new rate. The existing blinkLoop goroutine
	// will automatically pick up the new rate on the next tick it receives.
	if cm.blinkTimer != nil {
		cm.blinkTimer.Reset(rate)
		cm.blinkState = true // Reset to visible state when rate changes? Optional.
	}
}

// IsCursorRequested returns whether a cursor position was requested in the current frame.
// (Used internally or for debugging).
func (cm *CursorManager) IsCursorRequested() bool {
	return cm.requestValid
}