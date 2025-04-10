// app.go
package tinytui

import (
	"log"
	"sync"

	"github.com/gdamore/tcell/v2"
)

// Application represents the main TUI application manager.
type Application struct {
	screen    tcell.Screen
	root      Widget // The top-level widget (often a layout)
	focused   Widget // The widget currently receiving keyboard events
	modalRoot Widget // The widget defining the current modal focus scope (nil if none)

	events     chan tcell.Event        // Channel for incoming tcell events
	actionChan chan func(*Application) // Channel holds functions to execute
	stop       chan struct{}           // Channel to signal application termination
	redraw     chan struct{}           // Channel to signal a redraw is needed

	mu sync.Mutex // Protects access to screen, root, focused, modalRoot
}

// NewApplication creates and initializes a new TUI application.
func NewApplication() *Application {
	return &Application{
		events:     make(chan tcell.Event, 10),        // Buffered channel for events
		actionChan: make(chan func(*Application), 10), // Buffered channel for functions
		stop:       make(chan struct{}),
		redraw:     make(chan struct{}, 1), // Buffered channel, 1 is enough
	}
}

// Dispatch sends a function to be executed safely within the main application loop.
// This is the primary way UI elements should request state changes.
func (a *Application) Dispatch(actionFunc func(*Application)) {
	if actionFunc == nil {
		return
	}
	select {
	case a.actionChan <- actionFunc:
	case <-a.stop:
		log.Println("Warning: Dispatch ignored, application stopping.")
	default:
		log.Println("Warning: Action channel full. Dropping dispatched function.")
	}
}

// SetRoot sets the root widget for the application.
func (a *Application) SetRoot(widget Widget, fullscreen bool) *Application {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.root = widget
	if widget != nil {
		widget.SetApplication(a) // Link widget back to the app

		// Clear previous focus before determining the new one
		a.focused = nil
		a.modalRoot = nil // Ensure modal root is clear when root changes

		// Find the first focusable widget starting from the new root
		initialFocus := a.findFirstFocusable(widget) // Uses function from focus.go
		// Store the target, Run() will dispatch the focus action
		a.focused = initialFocus
		if initialFocus != nil {
			log.Printf("SetRoot: Found initial focus target: %T", initialFocus)
		} else {
			log.Println("SetRoot: Warning - Did not find any initial focus target.")
		}

	} else {
		a.focused = nil
		a.modalRoot = nil
	}
	a.QueueRedraw() // Queue redraw when root changes
	return a
}

// handleAction executes the dispatched function.
// This should be called ONLY from the main application goroutine.
func (a *Application) handleAction(actionFunc func(*Application)) {
	if actionFunc == nil {
		return
	}
	actionFunc(a) // Execute the closure
}

// Run starts the application's main event loop.
func (a *Application) Run() error {
	var err error
	a.mu.Lock()
	if a.screen == nil {
		a.screen, err = tcell.NewScreen()
		if err != nil {
			a.mu.Unlock()
			return err
		}
		if err = a.screen.Init(); err != nil {
			a.mu.Unlock()
			return err
		}
	}
	screen := a.screen
	root := a.root
	initialFocusTarget := a.focused // Get the target determined by SetRoot
	a.mu.Unlock()

	defer func() {
		if screen != nil {
			screen.Fini()
		}
	}()

	// --- Initial Setup ---
	if root != nil {
		w, h := screen.Size()
		root.SetRect(0, 0, w, h) // Initial layout calculation
		if initialFocusTarget != nil {
			// Dispatch initial focus setting
			targetForLog := initialFocusTarget
			a.Dispatch(func(app *Application) {
				log.Printf("Run: Dispatching initial SetFocus for %T", targetForLog)
				app.SetFocus(targetForLog) // Uses function from focus.go
			})
		} else {
			log.Println("Run: No initial focus target to dispatch SetFocus for.")
		}
	}

	// Start event polling goroutine
	go func() {
		for {
			event := screen.PollEvent()
			if event == nil {
				select {
				case <-a.stop:
				default:
					close(a.stop)
				}
				log.Println("Event polling goroutine: Screen closed, exiting.")
				return
			}
			select {
			case a.events <- event:
			case <-a.stop:
				log.Println("Event polling goroutine: Application stopping, exiting.")
				return
			}
		}
	}()

	a.QueueRedraw() // Queue initial draw

	// --- Main Event Loop ---
	for {
		select {
		case <-a.stop:
			log.Println("Main loop: Stop signal received, exiting.")
			return nil // Normal exit

		case actionFunc := <-a.actionChan:
			a.handleAction(actionFunc)

		case <-a.redraw:
			a.draw() // Uses function from draw.go (if we create one, or keep it here)

		case ev, ok := <-a.events:
			if !ok {
				log.Println("Main loop: Event channel closed, stopping.")
				select {
				case <-a.stop:
				default:
					close(a.stop)
				}
				continue
			}
			// Delegate event processing
			a.processEvent(ev) // Uses function from events.go
		}
	}
}

// draw clears the screen and redraws the entire widget tree starting from the root.
// (Keeping draw here for now as it's tightly coupled with screen/root access)
func (a *Application) draw() {
	a.mu.Lock()
	screen := a.screen
	root := a.root
	a.mu.Unlock()

	if screen == nil {
		log.Println("Draw: Screen not initialized.")
		return
	}

	screen.HideCursor()

	// Recalculate layout before drawing
	sw, sh := screen.Size()
	if root != nil {
		root.SetRect(0, 0, sw, sh)
	}

	screen.Clear()
	if root != nil {
		root.Draw(screen)
	}
	screen.Show()
}

// QueueRedraw requests a redraw of the application screen. It's non-blocking.
func (a *Application) QueueRedraw() {
	select {
	case a.redraw <- struct{}{}:
	default: // Avoid blocking if a redraw is already pending
	}
}

// Stop signals the application to terminate its event loop and clean up.
func (a *Application) Stop() {
	select {
	case <-a.stop: // Already stopping
		return
	default:
		log.Println("Stop: Closing stop channel.")
		close(a.stop)
	}
}

// Screen returns the underlying tcell.Screen instance. Use with caution regarding thread safety.
func (a *Application) Screen() tcell.Screen {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.screen
}