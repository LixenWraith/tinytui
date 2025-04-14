// cursor.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
	"time"
)

// CursorManager handles the visibility and drawing of the single cursor
// used by TextInput components.
type CursorManager struct {
	screen       tcell.Screen
	app          *Application
	requestedX   int
	requestedY   int
	requestValid bool // Was Request() called this frame?
	blinkRate    time.Duration
	blinkState   bool // Is the cursor currently visible due to blinking?
	blinkTimer   *time.Ticker
}

// NewCursorManager creates a cursor manager.
func NewCursorManager(app *Application, screen tcell.Screen, rate time.Duration) *CursorManager {
	if rate <= 0 {
		rate = 500 * time.Millisecond // Default blink rate
	}

	cm := &CursorManager{
		app:        app,
		screen:     screen,
		blinkRate:  rate,
		blinkState: true, // Start with cursor visible
	}

	// Start blink timer
	cm.blinkTimer = time.NewTicker(rate)
	go cm.blinkLoop()

	return cm
}

// Request sets the desired cursor position for the current frame.
// Called by the focused TextInput during its Draw phase.
func (cm *CursorManager) Request(x, y int) {
	cm.requestedX = x
	cm.requestedY = y
	cm.requestValid = true
}

// ResetForFrame clears the cursor request before drawing starts.
// Called by Application at the beginning of its draw cycle.
func (cm *CursorManager) ResetForFrame() {
	cm.requestValid = false
}

// Draw renders the cursor on the screen if requested and blink state allows.
// Called by Application after all components are drawn.
func (cm *CursorManager) Draw() {
	shouldShow := cm.requestValid && cm.blinkState
	x, y := cm.requestedX, cm.requestedY

	if shouldShow {
		cm.screen.ShowCursor(x, y)
	} else {
		cm.screen.HideCursor()
	}
}

// Stop halts the blinking timer goroutine.
func (cm *CursorManager) Stop() {
	if cm.blinkTimer != nil {
		cm.blinkTimer.Stop()
	}
}

// blinkLoop toggles the cursor visibility state at the blink rate.
func (cm *CursorManager) blinkLoop() {
	for range cm.blinkTimer.C {
		cm.blinkState = !cm.blinkState

		// Redraw on blink toggle
		if cm.app != nil {
			cm.app.QueueRedraw()
		}
	}
}

// SetBlinkRate changes the cursor blink rate.
func (cm *CursorManager) SetBlinkRate(rate time.Duration) {
	if rate <= 0 {
		return
	}

	if cm.blinkRate == rate {
		return
	}

	cm.blinkRate = rate

	// Restart the timer with the new rate
	if cm.blinkTimer != nil {
		cm.blinkTimer.Stop()
	}
	cm.blinkTimer = time.NewTicker(rate)

	// Restart the blink loop
	go cm.blinkLoop()
}

// IsCursorRequested returns whether a cursor has been requested for the current frame.
func (cm *CursorManager) IsCursorRequested() bool {
	return cm.requestValid
}