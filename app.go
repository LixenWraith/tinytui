// app.go
package tinytui

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

// Application represents the main TUI application manager.
type Application struct {
	screen        tcell.Screen
	root          Widget // The top-level widget (often a layout)
	focused       Widget // The widget currently receiving keyboard events
	modalRoot     Widget // The widget defining the current modal focus scope (nil if none)
	previousFocus Widget // Widget focused before modal opened (for returning focus)

	events     chan tcell.Event        // Channel for incoming tcell events
	actionChan chan func(*Application) // Channel holds functions to execute
	stop       chan struct{}           // Channel to signal application termination
	redraw     chan struct{}           // Channel to signal a redraw is needed

	// Focus optimization
	focusableCache map[Widget][]Widget // Cache of focusable widgets by parent
	cacheValid     bool                // Whether the cache is valid

	mu sync.Mutex // Protects access to screen, root, focused, modalRoot
}

// NewApplication creates and initializes a new TUI application.
func NewApplication() *Application {
	return &Application{
		events:         make(chan tcell.Event, 10),
		actionChan:     make(chan func(*Application), 10),
		stop:           make(chan struct{}),
		redraw:         make(chan struct{}, 1),
		focusableCache: make(map[Widget][]Widget),
		cacheValid:     false,
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
	default:
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
		a.modalRoot = nil        // Ensure modal root is clear when root changes
		a.previousFocus = nil    // Clear previous focus state
		a.invalidateFocusCache() // Clear focus cache

		// Find the first focusable widget starting from the new root
		initialFocus := a.findFirstFocusable(widget)
		a.focused = initialFocus
		if initialFocus != nil {
			initialFocus.Focus()
		}
	} else {
		a.focused = nil
		a.modalRoot = nil
		a.previousFocus = nil
		a.invalidateFocusCache()
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
	actionFunc(a)
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
			// Directly call Focus() on the widget to ensure its internal state is set
			initialFocusTarget.Focus()
			// Then dispatch SetFocus to handle application-level state properly
			a.Dispatch(func(app *Application) {
				app.SetFocus(initialFocusTarget)
			})
		}

		// Queue initial redraw
		a.QueueRedraw()
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
				return
			}
			select {
			case a.events <- event:
			case <-a.stop:
				return
			}
		}
	}()

	// --- Main Event Loop ---
	for {
		select {
		case <-a.stop:
			return nil // Normal exit

		case actionFunc := <-a.actionChan:
			a.handleAction(actionFunc)

		case <-a.redraw:
			a.draw()

		case ev, ok := <-a.events:
			if !ok {
				select {
				case <-a.stop:
				default:
					close(a.stop)
				}
				continue
			}
			// Delegate event processing
			a.processEvent(ev)
		}
	}
}

// draw clears the screen and redraws the entire widget tree starting from the root.
func (a *Application) draw() {
	a.mu.Lock()
	screen := a.screen
	root := a.root
	a.mu.Unlock()

	if screen == nil {
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
		close(a.stop)
	}
}

// Screen returns the underlying tcell.Screen instance. Use with caution regarding thread safety.
func (a *Application) Screen() tcell.Screen {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.screen
}

// SetModalRoot sets the widget that defines the current modal focus scope.
// It remembers the currently focused widget to restore focus when the modal is closed.
func (a *Application) SetModalRoot(widget Widget) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.modalRoot != widget {
		// Remember currently focused widget before changing to modal
		a.previousFocus = a.focused
		a.modalRoot = widget
		a.invalidateFocusCache() // Modal changes focusable widget scope
	}
}

// ClearModalRoot removes the modal focus scope and attempts to restore
// previous focus if possible.
func (a *Application) ClearModalRoot() {
	a.mu.Lock()
	prevFocus := a.previousFocus
	a.modalRoot = nil
	a.previousFocus = nil
	a.invalidateFocusCache() // Changing modal status affects focusable widgets
	a.mu.Unlock()

	// Restore focus to previous widget if it's still valid
	if prevFocus != nil && prevFocus.IsVisible() && prevFocus.Focusable() {
		a.SetFocus(prevFocus)
	} else if a.root != nil {
		// Fall back to first focusable if previous focus is no longer valid
		a.mu.Lock()
		firstFocus := a.findFirstFocusable(a.root)
		a.mu.Unlock()
		if firstFocus != nil {
			a.SetFocus(firstFocus)
		}
	}
}

// invalidateFocusCache clears the cached focusable widgets
func (a *Application) invalidateFocusCache() {
	a.focusableCache = make(map[Widget][]Widget)
	a.cacheValid = false
}

// findFocusableWidgetsCached performs a DFS to find all visible and focusable widgets,
// using a cache for improved performance with large widget trees.
func (a *Application) findFocusableWidgetsCached(searchRoot Widget) []Widget {
	if searchRoot == nil {
		return nil
	}

	// Check cache first
	if a.cacheValid {
		if widgets, ok := a.focusableCache[searchRoot]; ok {
			return widgets
		}
	}

	// Not in cache, rebuild list
	focusableWidgets := make([]Widget, 0)
	a.findFocusableWidgets(searchRoot, &focusableWidgets)

	// Cache the result for future use
	a.focusableCache[searchRoot] = focusableWidgets
	a.cacheValid = true

	return focusableWidgets
}

// findFocusableWidgets performs a DFS to find all *visible* and *focusable* widgets
// starting from the given node.
func (a *Application) findFocusableWidgets(startNode Widget, focusable *[]Widget) {
	if startNode == nil || !startNode.IsVisible() { // Check visibility first
		return // Don't traverse invisible widgets or their children
	}

	// Check focusable *after* visibility
	if startNode.Focusable() {
		*focusable = append(*focusable, startNode)
	}

	// Recursively check children
	if children := startNode.Children(); children != nil {
		for _, child := range children {
			a.findFocusableWidgets(child, focusable)
		}
	}
}

// findFirstFocusable finds the first *visible* and *focusable* widget in a DFS traversal.
func (a *Application) findFirstFocusable(start Widget) Widget {
	if start == nil || !start.IsVisible() {
		return nil
	}
	if start.Focusable() { // Focusable check implies visible here
		return start
	}
	if children := start.Children(); children != nil {
		for _, child := range children {
			found := a.findFirstFocusable(child)
			if found != nil {
				return found
			}
		}
	}
	return nil
}

// findNextFocus finds the next (or previous) focusable widget within the scope of searchRoot.
// This version uses the cache for better performance.
func (a *Application) findNextFocus(currentFocused Widget, searchRoot Widget, forward bool) Widget {
	if searchRoot == nil {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	focusableWidgets := a.findFocusableWidgetsCached(searchRoot)

	if len(focusableWidgets) == 0 {
		return nil
	}

	numFocusable := len(focusableWidgets)
	currentIndex := -1

	// Find the current widget in the list
	if currentFocused != nil {
		for i, w := range focusableWidgets {
			if w == currentFocused {
				currentIndex = i
				break
			}
		}
	}

	// Determine the next index based on direction and current index
	var nextIndex int
	if currentIndex == -1 {
		// Current widget not found or nil, start from beginning/end
		if forward {
			nextIndex = 0
		} else {
			nextIndex = numFocusable - 1
		}
	} else {
		// Move from the found index, wrapping around
		if forward {
			nextIndex = (currentIndex + 1) % numFocusable
		} else {
			nextIndex = (currentIndex - 1 + numFocusable) % numFocusable
		}
	}

	return focusableWidgets[nextIndex]
}

// SetFocus changes the currently focused widget with improved safety.
// It calls Blur() on the previously focused widget and Focus() on the new one.
// It only sets focus if the target widget is Focusable and Visible.
func (a *Application) SetFocus(widget Widget) {
	// Check if widget is focusable and visible before acquiring the lock
	if widget != nil && (!widget.Focusable() || !widget.IsVisible()) {
		return
	}

	a.mu.Lock()
	// No change needed?
	if a.focused == widget {
		a.mu.Unlock()
		return
	}

	// Capture values safely under lock
	oldWidget := a.focused
	a.focused = widget
	a.mu.Unlock()

	// Call methods with captured references after releasing the lock
	if oldWidget != nil {
		oldWidget.Blur()
	}

	if widget != nil {
		widget.Focus()
	}

	// Queue a redraw after changing focus
	a.QueueRedraw()
}