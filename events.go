// events.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
)

// processEvent is called by the main loop to handle a single tcell event.
func (a *Application) processEvent(ev tcell.Event) {
	// Handle key events with global keybindings first
	if keyEvent, ok := ev.(*tcell.EventKey); ok {
		a.mu.Lock()
		combo := keyModCombo{
			Key: keyEvent.Key(),
			Mod: keyEvent.Modifiers(),
		}

		handler, found := a.globalKeyBindings[combo]
		a.mu.Unlock()

		if found && handler() {
			return // Global keybinding handled the event
		}
	}

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

		// 1. Global/Modal key bindings
		key := event.Key()
		consumed = a.handleGlobalKeys(key, currentRoot, currentFocused, currentModalRoot)

		// 2. Pass to focused widget (if not consumed)
		if !consumed && currentFocused != nil {
			consumed = currentFocused.HandleEvent(event)
		} else if !consumed && currentFocused == nil {
		}

		// 3. Bubbling (if not consumed and focus exists)
		if !consumed && currentFocused != nil {
			bubbleTarget := currentFocused.Parent()
			for bubbleTarget != nil {
				// Stop bubbling if we hit the parent of the modal root
				if currentModalRoot != nil && bubbleTarget == currentModalRoot.Parent() {
					break
				}
				consumed = bubbleTarget.HandleEvent(event)
				if consumed {
					break
				}
				bubbleTarget = bubbleTarget.Parent()
			}
		}

	// No mouse handling at this time, pass through for terminal support
	// case *tcell.EventMouse:
	// 	// Basic mouse handling: Pass to focused widget first.
	// 	if currentFocused != nil {
	// 		consumed = currentFocused.HandleEvent(event)
	// 	}
	// NOTE: consumed is unused, keeping for potential future use or debugging

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
	// Don't handle Escape here since it's now handled by global keybindings
	switch key {
	case tcell.KeyCtrlC: // Ctrl+C always quits
		a.Stop()
		return true

	case tcell.KeyTab: // --- Focus Forward ---
		searchRoot := currentRoot
		if currentModalRoot != nil {
			searchRoot = currentModalRoot
		}
		if searchRoot != nil {
			next := a.findNextFocus(currentFocused, searchRoot, true)
			if next != nil && next != currentFocused {
				a.Dispatch(func(app *Application) { app.SetFocus(next) })
			}
		}
		return true // Consume Tab

	case tcell.KeyBacktab: // --- Focus Backward ---
		searchRoot := currentRoot
		if currentModalRoot != nil {
			searchRoot = currentModalRoot
		}
		if searchRoot != nil {
			prev := a.findNextFocus(currentFocused, searchRoot, false)
			if prev != nil && prev != currentFocused {
				a.Dispatch(func(app *Application) { app.SetFocus(prev) })
			}
		}
		return true // Consume Shift+Tab
	}
	return false // Key not handled globally
}