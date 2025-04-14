// application.go
package tinytui

import (
	"fmt" // Import fmt for error formatting
	"github.com/gdamore/tcell/v2"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Application manages the screen, event loop, layout, focus, and drawing.
type Application struct {
	screen    tcell.Screen
	layout    *Layout
	cursorMgr *CursorManager

	// Focus management
	focusedComponent Component

	// Event management
	eventChan  chan tcell.Event
	cmdChan    chan Command
	redrawChan chan struct{} // Buffered channel (size 1) for redraw requests
	stopChan   chan struct{} // Closed to signal application stop

	// Configuration
	theme             Theme
	showPaneIndices   bool
	screenMode        ScreenMode
	clearScreenOnExit bool

	// Keybindings
	keyHandlers  map[KeyModCombo]KeyHandler   // Handlers for specific key+modifier combos
	runeHandlers []func(*tcell.EventKey) bool // Handlers specifically for rune inputs (checked in order)

	// Performance
	maxFPS     int          // Maximum redraw rate
	frameTimer *time.Ticker // Ticker for enforcing maxFPS redraw checks
}

// NewApplication creates a new application with default settings.
// Initializes the theme from the current global theme.
func NewApplication() *Application {
	// Ensure a theme is available globally
	if GetTheme() == nil {
		// This might happen if NewApplication is called before package init completes
		// or if no themes were registered. Initialize default themes here as a safeguard.
		RegisterTheme(NewDefaultTheme())
		RegisterTheme(NewTurboTheme())
		SetTheme(ThemeDefault)
	}

	app := &Application{
		eventChan:         make(chan tcell.Event, 20), // Buffer for incoming tcell events
		cmdChan:           make(chan Command, 20),     // Buffer for internal commands
		redrawChan:        make(chan struct{}, 1),     // Buffer of 1 to coalesce redraw requests
		stopChan:          make(chan struct{}),
		keyHandlers:       make(map[KeyModCombo]KeyHandler),
		runeHandlers:      make([]func(*tcell.EventKey) bool, 0),
		showPaneIndices:   true,
		screenMode:        ScreenNormal,
		clearScreenOnExit: true,
		theme:             GetTheme(), // Initialize with the globally set theme
		maxFPS:            60,         // Default FPS
	}
	return app
}

// SetTheme sets the application theme and notifies components recursively.
func (app *Application) SetTheme(theme Theme) {
	if theme == nil || app.theme == theme {
		return // Ignore nil themes or no change
	}

	app.theme = theme

	// Notify components about the theme change recursively
	app.notifyThemeChange(theme)

	// Ensure a redraw happens to reflect the new theme
	app.QueueRedraw()
}

// notifyThemeChange propagates the theme change throughout the component tree.
func (app *Application) notifyThemeChange(theme Theme) {
	if app.layout != nil {
		// Start recursive theme application from the root layout
		app.layout.ApplyThemeRecursively(theme)
	}
}

// GetTheme returns the application's current theme.
// It returns the theme specifically set on the Application instance.
func (app *Application) GetTheme() Theme {
	// Return the app's specific theme instance. It's initialized
	// from the global theme in NewApplication but can be changed per-app.
	if app.theme == nil {
		// Fallback if somehow theme becomes nil after initialization
		app.theme = GetTheme()
		if app.theme == nil {
			app.theme = NewDefaultTheme()
		}
	}
	return app.theme
}

// SetLayout sets the application's root layout.
// Associates the layout, applies theme, and relies on layout.SetApplication
// to trigger the initial navigation index assignment.
func (app *Application) SetLayout(layout *Layout) {
	app.layout = layout
	if layout != nil {
		// Associate layout with app (this calls layout.SetApplication which
		// should now trigger assignNavigationIndices if it's the root layout)
		layout.SetApplication(app)

		// Apply theme (might be redundant if SetApplication handles it fully, but safe)
		currentTheme := app.GetTheme()
		if currentTheme != nil {
			layout.ApplyThemeRecursively(currentTheme)
		}
	}
	app.QueueRedraw()
}

// GetLayout returns the application's root layout.
func (app *Application) GetLayout() *Layout {
	return app.layout
}

// SetShowPaneIndices sets whether pane indices (Alt+Number hints) should be shown in pane borders.
func (app *Application) SetShowPaneIndices(show bool) {
	if app.showPaneIndices != show {
		app.showPaneIndices = show
		// Propagate this setting to panes? Or let panes check app setting during draw?
		// Let Panes check app.showPaneIndices during Draw.
		app.QueueRedraw()
	}
}

// IsShowPaneIndicesEnabled returns whether pane indices should be shown.
// Used by Pane during drawing.
func (app *Application) IsShowPaneIndicesEnabled() bool {
	return app.showPaneIndices
}

// SetScreenMode sets the desired screen mode (Normal, Fullscreen, Alternate).
func (app *Application) SetScreenMode(mode ScreenMode) {
	if app.screenMode == mode {
		return
	}
	app.screenMode = mode

	// Apply screen mode immediately if screen is already initialized
	if app.screen != nil {
		app.applyScreenMode()
		app.QueueRedraw() // Redraw might be needed after mode change
	}
}

// SetClearScreenOnExit sets whether the screen should be cleared when the application exits.
func (app *Application) SetClearScreenOnExit(clear bool) {
	app.clearScreenOnExit = clear
}

// SetMaxFPS sets the maximum frames per second for redraws.
// Affects how often dirty component checks and redraws occur via the frame timer.
func (app *Application) SetMaxFPS(fps int) {
	if fps <= 0 {
		fps = 60 // Default to 60 FPS if given invalid value
	}
	if app.maxFPS == fps {
		return
	}

	app.maxFPS = fps

	// If running, reset the frame timer ticker
	if app.frameTimer != nil {
		app.frameTimer.Stop()
		frameDelay := time.Second / time.Duration(app.maxFPS)
		app.frameTimer.Reset(frameDelay) // Use Reset for existing ticker
	}
}

// Run initializes the screen, starts the event loop, and handles drawing and events.
// Returns an error if initialization fails.
func (app *Application) Run() error {
	var err error

	// Initialize screen if not already done
	if app.screen == nil {
		app.screen, err = tcell.NewScreen()
		if err != nil {
			return fmt.Errorf("failed to create screen: %w", err)
		}

		// Enable mouse events? Consider adding an option.
		// if err = app.screen.EnableMouse(); err != nil {
		// 	 return fmt.Errorf("failed to enable mouse: %w", err)
		// }

		if err = app.screen.Init(); err != nil {
			// Attempt cleanup before returning error
			// app.screen.Fini() // Fini might panic if Init failed partially
			return fmt.Errorf("failed to initialize screen: %w", err)
		}

		// Apply the configured screen mode
		app.applyScreenMode()
	}

	// Initialize cursor manager
	// TODO: Allow configuring blink rate via Application option
	app.cursorMgr = NewCursorManager(app, app.screen, 500*time.Millisecond)

	// Set up frame timer for max FPS control
	frameDelay := time.Second / time.Duration(app.maxFPS)
	app.frameTimer = time.NewTicker(frameDelay)
	defer app.frameTimer.Stop() // Ensure timer stops on exit

	// Start event polling in a separate goroutine
	eventPollDone := make(chan struct{})
	go func() {
		defer close(eventPollDone)
		app.pollEvents()
	}()

	// Set up signal handling for graceful shutdown (Ctrl+C, SIGTERM)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	signalHandlingDone := make(chan struct{})
	go func() {
		defer close(signalHandlingDone)
		select {
		case <-sigChan:
			app.Stop() // Request application stop on signal
		case <-app.stopChan:
			// Application is stopping normally, just exit goroutine
			return
		}
	}()

	// Initial redraw to show the UI
	app.QueueRedraw()

	// --- Main event loop ---
	defer app.shutdown() // Ensure shutdown runs even if loop exits unexpectedly

	for {
		select {
		case <-app.stopChan:
			// Stop signal received, exit loop (shutdown handled by defer)
			// Wait for signal handler goroutine to finish?
			<-signalHandlingDone
			// Wait for event polling to finish? It checks stopChan too.
			<-eventPollDone
			return nil // Normal exit

		case ev, ok := <-app.eventChan:
			if !ok {
				// Event channel closed (likely due to pollEvents stopping after screen error/finalize)
				return fmt.Errorf("event channel closed unexpectedly") // Indicate error exit
			}
			// Process the received terminal event
			app.ProcessEvent(ev)

		case cmd := <-app.cmdChan:
			// Execute command received via Dispatch
			cmd.Execute(app)

		case <-app.redrawChan:
			// Redraw request received (coalesced)
			app.draw()

		case <-app.frameTimer.C:
			// Frame tick: Check if any component marked itself as dirty
			if app.checkDirtyComponents() {
				app.draw() // Draw if components are dirty
			}
		}
	}
}

// applyScreenMode applies the current screen mode setting to the screen.
func (app *Application) applyScreenMode() {
	if app.screen == nil {
		return
	}
	switch app.screenMode {
	case ScreenFullscreen:
		// Best effort: clear and set a common background. Actual fullscreen depends on terminal.
		app.screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack))
		app.screen.Clear()
		app.screen.Sync() // Ensure changes are visible
	case ScreenAlternate:
		// tcell handles enabling the alternate screen buffer automatically on Init()
		// if the terminal supports it. No explicit action needed here after Init.
		// We might need to ensure styles are reset if switching *to* this mode.
		app.screen.SetStyle(tcell.StyleDefault) // Reset style
		app.screen.Sync()
	case ScreenNormal:
		// If switching from Alternate, tcell handles restoring on Fini().
		// Reset style to default terminal style.
		app.screen.SetStyle(tcell.StyleDefault) // Reset style
		app.screen.Sync()
	}
}

// pollEvents polls for terminal events and sends them to the event channel.
// Runs in its own goroutine, exits when stopChan is closed.
func (app *Application) pollEvents() {
	for {
		// Check stop signal *before* polling to allow faster exit
		select {
		case <-app.stopChan:
			return // Application stopping
		default:
			// Continue polling
		}

		ev := app.screen.PollEvent()
		if ev == nil {
			// Screen was finalized or polling failed critically.
			// Signal the app to stop, if not already stopping.
			app.Stop()
			return
		}

		// Send event to main loop, or return if stopping
		select {
		case app.eventChan <- ev:
			// Event successfully sent
		case <-app.stopChan:
			// Application is stopping, terminate event polling
			return
		}
	}
}

// checkDirtyComponents checks if any component within the layout needs redrawing.
func (app *Application) checkDirtyComponents() bool {
	if app.layout == nil {
		return false // Nothing to check
	}
	// Delegate check to the layout, which checks recursively
	return app.layout.HasDirtyComponents()
}

// draw renders the current UI state to the screen.
func (app *Application) draw() {
	if app.screen == nil || app.layout == nil {
		return // Cannot draw without screen or layout
	}

	// Reset cursor request state for this frame
	if app.cursorMgr != nil {
		app.cursorMgr.ResetForFrame()
	}

	// Hide cursor temporarily during draw operations to avoid flicker
	app.screen.HideCursor()

	// Get current screen dimensions
	width, height := app.screen.Size()

	// Update layout dimensions (triggers recalculation if size changed)
	app.layout.SetRect(0, 0, width, height)

	// Draw the layout (which recursively draws panes and components)
	app.layout.Draw(app.screen)

	// Draw the cursor if requested by a component (e.g., TextInput) after components
	if app.cursorMgr != nil {
		app.cursorMgr.Draw() // This will call ShowCursor or HideCursor appropriately
	}

	// Show the updated screen buffer
	app.screen.Show()

	// Clear dirty flags recursively after a successful draw
	// Do this *after* screen.Show() to ensure flags are only cleared on success.
	app.layout.ClearAllDirtyFlags()
}

// shutdown cleans up resources and restores the terminal. Called on normal exit.
func (app *Application) shutdown() error {
	// Stop timers and managers first
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
		// Ensure cursor is hidden (though cursorMgr.Stop might do this)
		app.screen.HideCursor()

		// Disable mouse?
		// app.screen.DisableMouse()

		if app.clearScreenOnExit {
			app.screen.Clear()
			app.screen.Sync() // Ensure clear is displayed
		}

		// Restore terminal state (important!)
		app.screen.Fini()
		app.screen = nil // Prevent further use
	}

	// Channels will be implicitly handled as goroutines exit due to stopChan closing

	return nil // Indicate successful shutdown
}

// Stop signals the application to gracefully terminate the main loop. Idempotent.
func (app *Application) Stop() {
	select {
	case <-app.stopChan:
		// Already stopping or stopped
		return
	default:
		// Close stopChan to signal all loops and goroutines
		close(app.stopChan)
	}
}

// QueueRedraw requests a redraw on the next cycle of the event loop.
// It's buffered (size 1), so multiple calls between draw cycles result in only one redraw.
func (app *Application) QueueRedraw() {
	// Non-blocking send to redraw channel
	select {
	case app.redrawChan <- struct{}{}:
		// Redraw request successfully queued
	default:
		// Redraw already queued, do nothing (avoids blocking)
	}
}

// queueRedraw is an internal helper used by RedrawCommand.
func (app *Application) queueRedraw() {
	app.QueueRedraw()
}

// Dispatch sends a command to be executed asynchronously by the application's main loop.
// This is the safe way for components or other goroutines to modify application state or trigger actions.
func (app *Application) Dispatch(cmd Command) {
	if cmd == nil {
		return // Ignore nil commands
	}

	// Non-blocking send to command channel
	select {
	case app.cmdChan <- cmd:
		// Command successfully queued
	case <-app.stopChan:
		// Application is stopping, discard command
		// Log discarded command? fmt.Printf("Warning: Command %T discarded during shutdown\n", cmd)
	}
}

// SetFocus changes the focused component, handling blur/focus events.
func (app *Application) SetFocus(component Component) {
	// Don't focus nil, non-focusable, or invisible components
	if component != nil && (!component.Focusable() || !component.IsVisible()) {
		component = nil // Treat request to focus invalid component as request to remove focus
	}

	currentFocus := app.focusedComponent
	// Don't change focus if it's already the target
	if currentFocus == component {
		return
	}

	// Blur the previously focused component
	if currentFocus != nil {
		currentFocus.Blur()
	}

	// Set the new focused component (can be nil)
	app.focusedComponent = component

	// Focus the new component (if not nil)
	if component != nil {
		component.Focus()
	}

	// Queue a redraw to reflect focus changes (e.g., style, cursor)
	app.QueueRedraw()
}

// GetFocusedComponent returns the currently focused component, or nil if none.
func (app *Application) GetFocusedComponent() Component {
	return app.focusedComponent
}

// cycleFocus moves focus to the next or previous focusable component in the layout tree.
func (app *Application) cycleFocus(forward bool) {
	if app.layout == nil {
		return
	}

	// Get all currently focusable components in the layout
	focusables := app.layout.GetAllFocusableComponents()
	count := len(focusables)
	if count <= 1 {
		// If only one focusable item, ensure it's focused
		if count == 1 && app.focusedComponent != focusables[0] {
			app.SetFocus(focusables[0])
		}
		return // No focus change possible if 0 or 1 focusable items
	}

	currentFocused := app.focusedComponent
	currentIndex := -1
	if currentFocused != nil {
		// Find the index of the currently focused component in the list
		for i, comp := range focusables {
			if comp == currentFocused {
				currentIndex = i
				break
			}
		}
	}

	// Calculate the next index based on direction and current index
	nextIndex := 0
	if currentIndex != -1 { // If something was focused
		if forward {
			nextIndex = (currentIndex + 1) % count
		} else {
			nextIndex = (currentIndex - 1 + count) % count // Modulo arithmetic for wrapping backward
		}
	} else if !forward { // Nothing focused, cycling backward
		nextIndex = count - 1 // Start from the last item
	}
	// If nothing focused and cycling forward, nextIndex remains 0 (the first item)

	// Set focus to the next component in the cycle
	app.SetFocus(focusables[nextIndex])
}

// handleResize handles terminal resize events.
func (app *Application) handleResize(ev *tcell.EventResize) {
	// Sync the screen size with tcell's internal state
	if app.screen != nil {
		app.screen.Sync()
	}
	// Queue a redraw to re-layout and redraw everything for the new size
	app.QueueRedraw()
}

// RegisterKeyHandler registers a handler function for a specific key (non-rune) and modifier combination.
// The handler function should return true if the event was handled, false otherwise.
func (app *Application) RegisterKeyHandler(key tcell.Key, mod tcell.ModMask, handler func() bool) {
	// We specifically don't handle tcell.KeyRune here; use RegisterRuneHandler for that.
	if key == tcell.KeyRune {
		// Log a warning? This function isn't intended for rune keys.
		// fmt.Printf("Warning: RegisterKeyHandler called with tcell.KeyRune for key %v\n", key)
		return
	}
	combo := KeyModCombo{
		Key: key,
		Mod: mod,
	}
	// TODO: Add locking if handlers can be registered/deregistered concurrently with event loop?
	// For now, assume registration happens before Run() or via Dispatch command.
	app.keyHandlers[combo] = handler
}

// RegisterRuneHandler registers a handler function for a specific rune and modifier combination.
// The handler function should return true if the event was handled, false otherwise.
// Handlers are checked in the order they are registered.
func (app *Application) RegisterRuneHandler(r rune, mod tcell.ModMask, handler func() bool) {
	// TODO: Add locking if handlers can be registered/deregistered concurrently?
	app.runeHandlers = append(app.runeHandlers, func(ev *tcell.EventKey) bool {
		// Check if the event matches the specific rune and modifiers
		if ev.Key() == tcell.KeyRune && ev.Rune() == r && ev.Modifiers() == mod {
			return handler() // Execute the handler
		}
		return false // Event doesn't match this handler
	})
}

// GetCursorManager returns the application's cursor manager instance.
// Used by input components to request cursor visibility and position.
func (app *Application) GetCursorManager() *CursorManager {
	return app.cursorMgr
}

// application.go

// ProcessEvent handles incoming tcell events. Updated Alt+Num logic.
func (app *Application) ProcessEvent(ev tcell.Event) {
	focusedComp := app.GetFocusedComponent()

	switch ev := ev.(type) {
	case *tcell.EventKey:
		key := ev.Key()
		mod := ev.Modifiers()
		r := ev.Rune()

		// --- 1. Critical Global Keys ---
		if key == tcell.KeyCtrlC {
			app.Stop()
			return
		}

		// --- 2. Focused Component Handling ---
		if focusedComp != nil && focusedComp.HandleEvent(ev) {
			return
		}

		// --- 3. Global Escape Key ---
		if key == tcell.KeyEscape {
			app.Stop()
			return
		}

		// --- 4. Alt+Number Pane Navigation (REVISED) ---
		if mod&tcell.ModAlt != 0 {
			navIndex := 0
			if r >= '1' && r <= '9' {
				navIndex = int(r - '0') // Direct conversion '1'->1, '9'->9
			} else if r == '0' {
				navIndex = 10 // Alt+0 maps to navigation index 10
			}
			// If a valid Alt+Number combo was pressed (resulting in navIndex 1-10)
			if navIndex > 0 {
				app.handleAltNumberNavigation(navIndex) // Call handler with 1-10 index
				return                                  // Event handled
			}
		}
		// --- End Alt+Number ---

		// --- 5. Registered Global Handlers ---
		keyHandled := false
		if key == tcell.KeyRune {
			handlers := make([]func(*tcell.EventKey) bool, len(app.runeHandlers))
			copy(handlers, app.runeHandlers)
			for _, handler := range handlers {
				if handler(ev) {
					keyHandled = true
					break
				}
			}
		} else {
			combo := KeyModCombo{Key: key, Mod: mod}
			if handler, ok := app.keyHandlers[combo]; ok {
				if handler() {
					keyHandled = true
				}
			}
		}
		if keyHandled {
			return
		} // Event handled by registered handler

		// --- 6. Global Focus Navigation (Tab / Shift+Tab) ---
		if key == tcell.KeyTab {
			app.cycleFocus(true)
			return
		}
		if key == tcell.KeyBacktab {
			app.cycleFocus(false)
			return
		}

		// --- Event Ignored ---

	case *tcell.EventResize:
		// Handle terminal resize events
		app.handleResize(ev)
		return

	case *tcell.EventMouse:
		// TODO: Implement Mouse Event Handling if needed
		return // Ignore mouse for now

		// Handle other event types if necessary
	}
}

// StopChan returns the channel that is closed when the application stops.
// Can be used in select statements by goroutines to react to application shutdown.
func (app *Application) StopChan() <-chan struct{} {
	return app.stopChan
}

// handleAltNumberNavigation handles Alt+[1-9, 0] key presses for pane navigation using NavIndex.
func (app *Application) handleAltNumberNavigation(targetNavIndex int) { // Now takes 1-10
	if app.layout == nil {
		return
	}

	// Use the layout method to find the pane by navigation index (1-10)
	pane := app.layout.GetPaneByNavIndex(targetNavIndex)
	if pane == nil {
		// Optional: Add a beep or status message?
		// appLog("No navigable pane found for Alt+%d", targetNavIndex % 10)
		return // No pane found for this navigation index
	}

	// Find the first focusable component within that pane
	comp := pane.GetFirstFocusableComponent()
	if comp != nil {
		app.SetFocus(comp) // Focus the found component
	} else {
		// This case should technically not happen if GetPaneByNavIndex only returns
		// panes that have focusable children, but added as safety.
		// appLog("Pane %d found but has no focusable component?", targetNavIndex)
	}
}