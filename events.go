// events.go
package tinytui

import (
	"log"

	"github.com/gdamore/tcell/v2"
)

// processEvent is called by the main loop to handle a single tcell event.
func (a *Application) processEvent(ev tcell.Event) {
	// Get current state needed for event routing
	a.mu.Lock()
	currentFocused := a.focused
	currentRoot := a.root
	currentModalRoot := a.modalRoot
	screen := a.screen // Needed for Resize
	a.mu.Unlock()

	consumed := false

	switch event := ev.(type) {
	case *tcell.EventResize:
		if screen != nil {
			screen.Sync()
		}
		// Recalculate layout in the next draw cycle
		a.QueueRedraw()
		consumed = true

	case *tcell.EventKey:
		log.Printf("Event Loop: Key Event: %v, Mod: %v, Rune: %q. Focused: %T", event.Key(), event.Modifiers(), event.Rune(), currentFocused)

		// 1. Global/Modal key bindings
		key := event.Key()
		consumed = a.handleGlobalKeys(key, currentRoot, currentFocused, currentModalRoot)

		// 2. Pass to focused widget (if not consumed)
		if !consumed && currentFocused != nil {
			log.Printf("Event Loop: Passing key event to focused widget %T", currentFocused)
			consumed = currentFocused.HandleEvent(event)
			log.Printf("Event Loop: Focused widget %T consumed event: %v", currentFocused, consumed)
		} else if !consumed && currentFocused == nil {
			log.Println("Event Loop: Key event not consumed globally, but no widget is focused.")
		}

		// 3. Bubbling (if not consumed and focus exists)
		if !consumed && currentFocused != nil {
			bubbleTarget := currentFocused.Parent()
			for bubbleTarget != nil {
				// Stop bubbling if we hit the parent of the modal root
				if currentModalRoot != nil && bubbleTarget == currentModalRoot.Parent() {
					break
				}
				log.Printf("Event Loop: Bubbling key event to parent %T", bubbleTarget)
				consumed = bubbleTarget.HandleEvent(event)
				log.Printf("Event Loop: Bubbled widget %T consumed event: %v", bubbleTarget, consumed)
				if consumed {
					break
				}
				bubbleTarget = bubbleTarget.Parent()
			}
		}

	case *tcell.EventMouse:
		// Basic mouse handling: Pass to focused widget first.
		// TODO: More advanced: Check widget under cursor
		if currentFocused != nil {
			consumed = currentFocused.HandleEvent(event)
		}

	default:
		// Pass other unhandled event types to focused widget
		if currentFocused != nil {
			consumed = currentFocused.HandleEvent(event)
		}
	}
}

// handleGlobalKeys processes key events that have application-wide or modal-specific behavior.
// Returns true if the key was consumed.
func (a *Application) handleGlobalKeys(key tcell.Key, currentRoot, currentFocused, currentModalRoot Widget) bool {
	switch key {
	case tcell.KeyCtrlC: // Ctrl+C always quits
		log.Println("handleGlobalKeys: Ctrl+C detected, stopping.")
		a.Stop()
		return true

	case tcell.KeyEscape: // Contextual Escape
		if currentModalRoot != nil {
			log.Println("handleGlobalKeys: Escape detected with modal, dispatching close.")
			// Dispatch function to close modal
			a.Dispatch(func(app *Application) {
				log.Printf("Action: Closing modal root %T via Escape\n", currentModalRoot)
				if currentModalRoot != nil { // Use the captured currentModalRoot
					currentModalRoot.SetVisible(false) // Hide the widget that was modal
					app.ClearModalRoot()               // Use method to clear internal state

					// Find focus target *outside* the now-hidden modal
					var returnFocus Widget
					if app.root != nil {
						returnFocus = app.findFirstFocusable(app.root)
					}
					app.SetFocus(returnFocus)
				}
			})
			return true
		} else {
			log.Println("handleGlobalKeys: Escape detected without modal, stopping.")
			a.Stop()
			return true
		}

	case tcell.KeyTab: // --- Focus Forward ---
		log.Println("handleGlobalKeys: Tab detected.")
		searchRoot := currentRoot
		if currentModalRoot != nil {
			searchRoot = currentModalRoot
			log.Println("handleGlobalKeys: Tab - Searching within modal root.")
		} else {
			log.Println("handleGlobalKeys: Tab - Searching within main root.")
		}
		if searchRoot != nil {
			next := a.findNextFocus(currentFocused, searchRoot, true)
			if next != nil && next != currentFocused {
				log.Printf("handleGlobalKeys: Tab - Found next focus: %T. Dispatching SetFocus.", next)
				a.Dispatch(func(app *Application) { app.SetFocus(next) })
			} else if next == nil {
				log.Println("handleGlobalKeys: Tab - No next focus found within scope.")
			} else {
				log.Println("handleGlobalKeys: Tab - Next focus is same as current.")
			}
		}
		return true // Consume Tab

	case tcell.KeyBacktab: // --- Focus Backward ---
		log.Println("handleGlobalKeys: Backtab detected.")
		searchRoot := currentRoot
		if currentModalRoot != nil {
			searchRoot = currentModalRoot
			log.Println("handleGlobalKeys: Backtab - Searching within modal root.")
		} else {
			log.Println("handleGlobalKeys: Backtab - Searching within main root.")
		}
		if searchRoot != nil {
			prev := a.findNextFocus(currentFocused, searchRoot, false)
			if prev != nil && prev != currentFocused {
				log.Printf("handleGlobalKeys: Backtab - Found previous focus: %T. Dispatching SetFocus.", prev)
				a.Dispatch(func(app *Application) { app.SetFocus(prev) })
			} else if prev == nil {
				log.Println("handleGlobalKeys: Backtab - No previous focus found within scope.")
			} else {
				log.Println("handleGlobalKeys: Backtab - Previous focus is same as current.")
			}
		}
		return true // Consume Shift+Tab
	}
	return false // Key not handled globally
}