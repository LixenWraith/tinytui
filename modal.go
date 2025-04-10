// modal.go
package tinytui

// SetModalRoot sets the widget that defines the current modal focus scope.
// Should only be called from within a dispatched function.
func (a *Application) SetModalRoot(widget Widget) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.modalRoot != widget {
		a.modalRoot = widget
	}
}

// ClearModalRoot removes the modal focus scope.
// Should only be called from within a dispatched function.
func (a *Application) ClearModalRoot() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.modalRoot != nil { // Only log if it was actually set
		a.modalRoot = nil
	}
}