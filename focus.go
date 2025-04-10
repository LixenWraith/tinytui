// focus.go
package tinytui

import (
	"fmt" // For logging
	"log"
	// No tcell needed here directly
)

// SetFocus changes the currently focused widget.
// It calls Blur() on the previously focused widget and Focus() on the new one.
// It only sets focus if the target widget is Focusable and Visible.
// This method should generally only be called from within a dispatched function
// to ensure it runs in the main UI loop.
func (a *Application) SetFocus(widget Widget) {
	a.mu.Lock() // Lock for modifying focused state

	// Ensure the target widget is actually focusable and visible
	if widget != nil && (!widget.Focusable() || !widget.IsVisible()) {
		log.Printf("SetFocus: Refused for non-focusable/invisible widget %T (Focusable: %v, Visible: %v)",
			widget, widget.Focusable(), widget.IsVisible())
		a.mu.Unlock()
		return
	}

	if a.focused == widget {
		a.mu.Unlock() // Unlock if no change
		return        // No change
	}

	oldType := "nil"
	if a.focused != nil {
		oldType = fmt.Sprintf("%T", a.focused)
	}
	newType := "nil"
	if widget != nil {
		newType = fmt.Sprintf("%T", widget)
	}
	log.Printf("SetFocus: Changing focus from %s to %s", oldType, newType)

	oldFocused := a.focused
	a.focused = widget // Update internal state

	a.mu.Unlock() // Unlock before calling widget methods

	// Blur the old widget (if any)
	if oldFocused != nil {
		oldFocused.Blur()
	}

	// Focus the new widget (if any)
	if widget != nil {
		widget.Focus() // Call the widget's Focus method
	}
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
func (a *Application) findNextFocus(currentFocused Widget, searchRoot Widget, forward bool) Widget {
	if searchRoot == nil {
		log.Println("findNextFocus: searchRoot is nil.")
		return nil // Cannot search without a root
	}

	// Rebuild the list of focusable widgets within the searchRoot scope each time
	localFocusableWidgets := make([]Widget, 0)
	a.findFocusableWidgets(searchRoot, &localFocusableWidgets) // Use helper

	if len(localFocusableWidgets) == 0 {
		return nil // No focusable widgets found in this scope
	}

	numFocusable := len(localFocusableWidgets)
	currentIndex := -1 // Default if currentFocused is nil or not found in the current list

	// Find the current widget in the dynamically generated list
	if currentFocused != nil {
		for i, w := range localFocusableWidgets {
			if w == currentFocused {
				currentIndex = i
				break
			}
		}
	}

	// Determine the next index based on direction and current index
	var nextIndex int
	if currentIndex == -1 {
		// Current widget not found or nil, start from beginning/end.
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

	return localFocusableWidgets[nextIndex]
}