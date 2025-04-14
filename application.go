// application.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Application manages the main event loop and screen
type Application struct {
	screen    tcell.Screen
	layout    *Layout
	cursorMgr *CursorManager

	// Focus management
	focusedComponent Component

	// Event management
	eventChan  chan tcell.Event
	cmdChan    chan Command
	redrawChan chan struct{}
	stopChan   chan struct{}

	// Configuration
	theme             Theme
	showPaneIndices   bool
	screenMode        ScreenMode
	clearScreenOnExit bool

	// Keybindings
	keyHandlers  map[KeyModCombo]KeyHandler
	runeHandlers []func(*tcell.EventKey) bool

	// Performance
	maxFPS     int
	frameTimer *time.Ticker
}

// NewApplication creates a new application with default settings.
func NewApplication() *Application {
	app := &Application{
		eventChan:         make(chan tcell.Event, 20),
		cmdChan:           make(chan Command, 20),
		redrawChan:        make(chan struct{}, 1), // Buffer of 1 to avoid blocking
		stopChan:          make(chan struct{}),
		keyHandlers:       make(map[KeyModCombo]KeyHandler),
		runeHandlers:      make([]func(*tcell.EventKey) bool, 0),
		showPaneIndices:   true,
		screenMode:        ScreenNormal,
		clearScreenOnExit: true,
		maxFPS:            60,
	}
	return app
}

// SetTheme sets the application theme.
func (app *Application) SetTheme(theme Theme) {
	app.theme = theme

	// Notify any themed components (minimal implementation for future expansion)
	if app.layout != nil {
		app.notifyThemeChange(theme)
	}

	app.QueueRedraw()
}

// notifyThemeChange notifies components implementing ThemedComponent about theme changes.
// This is a minimal implementation that will be expanded in future versions.
// Currently, this provides a foundation for component-specific theme handling.
func (app *Application) notifyThemeChange(theme Theme) {
	// This implementation is intentionally minimal as the ThemedComponent
	// interface is not fully utilized in the current version.
	// Future versions will have more sophisticated theme propagation.
	if app.layout == nil {
		return
	}

	// Only check the focused component as a proof of concept
	if focused := app.GetFocusedComponent(); focused != nil {
		if themed, ok := focused.(ThemedComponent); ok {
			themed.ApplyTheme(theme)
		}
	}
}

// GetTheme returns the current theme.
func (app *Application) GetTheme() Theme {
	return app.theme
}

// SetLayout sets the application layout.
func (app *Application) SetLayout(layout *Layout) {
	app.layout = layout

	if layout != nil {
		layout.SetApplication(app)
	}

	app.QueueRedraw()
}

// GetLayout returns the application's layout.
func (app *Application) GetLayout() *Layout {
	return app.layout
}

// SetShowPaneIndices sets whether pane indices should be shown.
func (app *Application) SetShowPaneIndices(show bool) {
	app.showPaneIndices = show
	app.QueueRedraw()
}

// SetScreenMode sets the screen mode.
func (app *Application) SetScreenMode(mode ScreenMode) {
	app.screenMode = mode

	// Apply screen mode if screen is initialized
	if app.screen != nil {
		switch mode {
		case ScreenFullscreen:
			app.screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack))
			app.screen.Clear()
		case ScreenAlternate:
			// tcell automatically uses alternate screen buffer
		case ScreenNormal:
			// Default behavior
		}
	}

	app.QueueRedraw()
}

// SetClearScreenOnExit sets whether the screen should be cleared on exit.
func (app *Application) SetClearScreenOnExit(clear bool) {
	app.clearScreenOnExit = clear
}

// SetMaxFPS sets the maximum frames per second for redraws.
func (app *Application) SetMaxFPS(fps int) {
	if fps <= 0 {
		fps = 60 // Default to 60 FPS if given invalid value
	}

	app.maxFPS = fps
}

// Run initializes the screen and starts the main event loop.
func (app *Application) Run() error {
	var err error

	// Initialize screen
	if app.screen == nil {
		app.screen, err = tcell.NewScreen()
		if err != nil {
			return err
		}

		if err = app.screen.Init(); err != nil {
			return err
		}

		// Apply screen mode
		app.applyScreenMode()
	}

	// Initialize cursor manager
	app.cursorMgr = NewCursorManager(app, app.screen, 500*time.Millisecond)

	// Set up frame timer
	frameDelay := time.Second / time.Duration(app.maxFPS)
	app.frameTimer = time.NewTicker(frameDelay)

	// Start event polling
	go app.pollEvents()

	// Set up signal handling for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-sigChan:
			// Force application to stop when Ctrl+C is pressed
			app.Stop()
		case <-app.stopChan:
			// Stop channel was closed elsewhere, just return
			return
		}
	}()

	// Initial redraw
	app.QueueRedraw()

	// Main event loop
	for {
		select {
		case <-app.stopChan:
			return app.shutdown()

		case ev, ok := <-app.eventChan:
			if !ok {
				// Channel closed, exit
				return app.shutdown()
			}
			app.ProcessEvent(ev)

		case cmd := <-app.cmdChan:
			cmd.Execute(app)

		case <-app.redrawChan:
			app.draw()

		case <-app.frameTimer.C:
			// Check if any components need redrawing
			needsRedraw := app.checkDirtyComponents()
			if needsRedraw {
				app.draw()
			}
		}
	}
}

// applyScreenMode applies the current screen mode to the screen.
func (app *Application) applyScreenMode() {
	switch app.screenMode {
	case ScreenFullscreen:
		app.screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack))
		app.screen.Clear()
	case ScreenAlternate:
		// tcell automatically uses alternate screen buffer
	case ScreenNormal:
		// Default behavior
	}
}

// pollEvents goroutine polls for terminal events and sends them to the event channel.
func (app *Application) pollEvents() {
	// Defer recover to prevent panic if the channel is closed unexpectedly
	// TODO: Remove after debugging
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		// Just log the panic or handle it gracefully
	// 		// This prevents the "close of closed channel" panic from propagating
	// 	}
	// }()
	//
	// Main event loop
	for {
		ev := app.screen.PollEvent()
		if ev == nil {
			// Screen was finalized, stop the application
			return
		}

		select {
		case app.eventChan <- ev:
			// Event sent successfully
		case <-app.stopChan:
			// Application is stopping
			return
		}
	}
}

// checkDirtyComponents checks if any components need redrawing.
func (app *Application) checkDirtyComponents() bool {
	layout := app.layout

	if layout == nil {
		return false
	}

	return app.layout.HasDirtyComponents()
}

// draw renders the current state to the screen.
func (app *Application) draw() {
	if app.screen == nil || app.layout == nil {
		return
	}

	// First, safely check if we can draw
	screen := app.screen
	layout := app.layout
	cursorMgr := app.cursorMgr

	// Reset cursor request for this frame - lock-free operation
	if cursorMgr != nil {
		cursorMgr.ResetForFrame()
	}

	// Get screen dimensions - lock-free operation
	width, height := screen.Size()

	// Set layout size - this might acquire layout lock
	layout.SetRect(0, 0, width, height)

	// Draw layout (which will draw all panes and components)
	layout.Draw(screen)

	// Draw cursor (if requested by a TextInput component)
	if cursorMgr != nil {
		cursorMgr.Draw()
	}

	// Show the screen - lock-free operation
	screen.Show()

	// Clear dirty flags after successful draw
	layout.ClearAllDirtyFlags()
}

// shutdown cleans up resources and returns the terminal to its original state.
func (app *Application) shutdown() error {
	// Stop timers
	if app.frameTimer != nil {
		app.frameTimer.Stop()
		app.frameTimer = nil
	}

	if app.cursorMgr != nil {
		app.cursorMgr.Stop()
		app.cursorMgr = nil
	}

	// Clean up screen
	if app.screen != nil {
		// Always hide cursor when exiting
		app.screen.HideCursor()

		if app.clearScreenOnExit {
			app.screen.Clear()
			app.screen.Sync()
		}

		// Ensure the terminal is properly restored
		app.screen.Fini()
	}
	app.screen = nil

	return nil
}

// Stop signals the application to stop.
func (app *Application) Stop() {
	select {
	case <-app.stopChan:
		// Already stopped
		return
	default:
		close(app.stopChan)
	}
}

// QueueRedraw requests a redraw on the next frame.
func (app *Application) QueueRedraw() {
	select {
	case app.redrawChan <- struct{}{}:
		// Redraw request sent
	default:
		// Channel full, a redraw is already queued
	}
}

// queueRedraw is an internal method for the RedrawCommand to use.
func (app *Application) queueRedraw() {
	app.QueueRedraw()
}

// Dispatch sends a command to be executed by the application.
func (app *Application) Dispatch(cmd Command) {
	if cmd == nil {
		return
	}

	select {
	case app.cmdChan <- cmd:
		// Command sent
	case <-app.stopChan:
		// Application is stopping
	}
}

// SetFocus changes the focused component.
func (app *Application) SetFocus(component Component) {
	// Don't focus nil components or non-focusable/invisible components
	if component == nil || !component.Focusable() || !component.IsVisible() {
		return
	}

	// Don't change focus if already focused
	if app.focusedComponent == component {
		return
	}

	// Get previous focused component
	prevFocused := app.focusedComponent

	// Set the new focused component
	app.focusedComponent = component

	// Blur the previously focused component
	if prevFocused != nil {
		prevFocused.Blur()
	}

	// Focus the new component
	component.Focus()

	// Queue a redraw
	app.QueueRedraw()
}

func (app *Application) GetFocusedComponent() Component {
	return app.focusedComponent
}

// handleAltNumberNavigation handles Alt+Number key presses for pane navigation.
func (app *Application) handleAltNumberNavigation(index int) {
	// Get layout reference safely
	layout := app.layout

	if layout == nil {
		return
	}

	// Convert 0-9 array index to 1-10 pane index
	paneIndex := index + 1

	// Get the pane at the specified index
	pane := layout.GetPaneByIndex(paneIndex)
	if pane == nil {
		return
	}

	// Find the first focusable component in the pane
	comp := pane.GetFirstFocusableComponent()
	if comp != nil {
		app.SetFocus(comp)
	}
}

// cycleFocus cycles focus through all focusable components.
func (app *Application) cycleFocus(forward bool) {
	// Get layout reference safely
	layout := app.layout
	currentFocused := app.focusedComponent

	if layout == nil {
		return
	}

	// Get all focusable components
	focusables := layout.GetAllFocusableComponents()
	if len(focusables) == 0 {
		return
	}

	// Find index of currently focused component
	currentIndex := -1
	for i, comp := range focusables {
		if comp == currentFocused {
			currentIndex = i
			break
		}
	}

	// Calculate next index
	nextIndex := 0
	if currentIndex >= 0 {
		if forward {
			nextIndex = (currentIndex + 1) % len(focusables)
		} else {
			nextIndex = (currentIndex - 1 + len(focusables)) % len(focusables)
		}
	}

	// Focus the next component
	app.SetFocus(focusables[nextIndex])
}

// handleResize handles window resize events.
func (app *Application) handleResize(ev *tcell.EventResize) {
	// Sync the screen
	app.screen.Sync()

	// Queue a redraw
	app.QueueRedraw()
}

// ASSIST: Logic (rewrite keyhandler functions)
// RegisterKeyHandler registers a handler for a specific key combination.
func (app *Application) RegisterKeyHandler(key tcell.Key, mod tcell.ModMask, handler func() bool) {
	// For rune keys, we still need the rune value
	if key == tcell.KeyRune {
		// This case requires a different approach since we can't determine
		// which rune this binding is for without additional info
		// We'll keep this as a special case
		app.runeHandlers = append(app.runeHandlers, func(ev *tcell.EventKey) bool {
			if ev.Modifiers() == mod {
				return handler()
			}
			return false
		})
	} else {
		// Normal key combo registration
		combo := KeyModCombo{
			Key: key,
			Mod: mod,
		}

		app.keyHandlers[combo] = handler
	}
}

// RegisterRuneHandler registers a handler for a specific rune key with modifiers.
func (app *Application) RegisterRuneHandler(r rune, mod tcell.ModMask, handler func() bool) {
	app.runeHandlers = append(app.runeHandlers, func(ev *tcell.EventKey) bool {
		if ev.Rune() == r && ev.Modifiers() == mod {
			return handler()
		}
		return false
	})
}

// GetCursorManager returns the application's cursor manager.
func (app *Application) GetCursorManager() *CursorManager {
	return app.cursorMgr
}