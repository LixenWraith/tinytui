// base_widget.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
	"sync"
)

// keyModCombo is used as a map key for keybindings.
type keyModCombo struct {
	Key tcell.Key
	Mod tcell.ModMask
}

// BaseWidget provides a default implementation for the Widget interface.
// Concrete widgets can embed this type to inherit default behavior.
type BaseWidget struct {
	rect        Rect
	focused     bool
	visible     bool                        // Visibility flag (defaults to false, initialize in constructors or SetVisible)
	app         *Application                // Pointer back to the app for queuing redraws
	parent      Widget                      // Pointer to the container widget
	keyBindings map[keyModCombo]func() bool // Map for keybindings: Key+Mod -> handler
	mu          sync.RWMutex
}

// Draw checks visibility before proceeding. Concrete widgets should override this.
func (b *BaseWidget) Draw(screen tcell.Screen) {
	// Visibility check is implicitly handled by IsVisible() called by callers or within overrides.
	// No explicit check needed here unless BaseWidget itself drew something.
	if !b.IsVisible() {
		// If a widget is invisible, it should not draw anything,
		// including its background or border.
		return
	}

	// Default: do nothing further (concrete widgets override)
}

// SetRect stores the allocated rectangle for the widget.
func (b *BaseWidget) SetRect(x, y, width, height int) {
	b.mu.Lock()
	b.rect = Rect{X: x, Y: y, Width: width, Height: height}
	b.mu.Unlock()
}

// GetRect returns the widget's current rectangle.
func (b *BaseWidget) GetRect() (x, y, width, height int) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.rect.X, b.rect.Y, b.rect.Width, b.rect.Height
}

// HandleEvent checks registered keybindings for *tcell.EventKey.
// If a binding matches (Key + ModMask) and its handler returns true,
// HandleEvent returns true. Otherwise, it returns false.
// Note: For KeyRune, this doesn't distinguish between different runes by default.
//
//	The registered handler function should check event.Rune() if needed.
func (b *BaseWidget) HandleEvent(event tcell.Event) bool {
	// Invisible widgets should generally not handle events either.
	if !b.IsVisible() {
		return false
	}

	keyEvent, ok := event.(*tcell.EventKey)
	if !ok {
		return false // Not a key event
	}

	b.mu.RLock()
	bindings := b.keyBindings
	b.mu.RUnlock()

	if bindings == nil {
		return false // No bindings registered
	}

	combo := keyModCombo{
		Key: keyEvent.Key(),
		Mod: keyEvent.Modifiers(),
	}

	// RLock again briefly to check the map
	b.mu.RLock()
	handler, found := bindings[combo]
	b.mu.RUnlock()

	if found {
		// Execute the handler. The handler itself might need to check
		// keyEvent.Rune() if the binding was for tcell.KeyRune.
		return handler() // Return handler's result (true if consumed)
	}

	return false // No matching binding found
}

// Focusable returns false by default. Widgets that can be focused should override this.
// Overrides should also check visibility.
func (b *BaseWidget) Focusable() bool {
	// Base implementation doesn't need visibility check as it always returns false.
	// Concrete implementations MUST check IsVisible().
	return false
}

// Focus sets the focused state to true and queues a redraw if the state changed.
func (b *BaseWidget) Focus() {
	// Cannot focus an invisible widget (checked using effective visibility)
	if !b.IsVisible() {
		// Ensure focus flag is false if becoming invisible prevents focus
		b.mu.Lock()
		if b.focused {
			b.focused = false
			app := b.app
			b.mu.Unlock()
			if app != nil {
				app.QueueRedraw() // Redraw if focus was lost due to invisibility
			}
		} else {
			b.mu.Unlock()
		}
		return
	}

	b.mu.Lock()
	changed := !b.focused
	b.focused = true
	app := b.app
	b.mu.Unlock()

	if changed && app != nil {
		app.QueueRedraw() // Redraw to potentially show focus indicator
	}
}

// Blur sets the focused state to false and queues a redraw if the state changes.
func (b *BaseWidget) Blur() {
	b.mu.Lock()
	changed := b.focused
	b.focused = false
	app := b.app
	b.mu.Unlock()

	if changed && app != nil {
		app.QueueRedraw() // Redraw to potentially remove focus indicator
	}
}

// IsFocused returns whether the widget currently has focus (considering visibility).
func (b *BaseWidget) IsFocused() bool {
	b.mu.RLock()
	isLocallyFocused := b.focused
	b.mu.RUnlock()
	// Check effective visibility without holding lock to avoid deadlock during parent check
	return isLocallyFocused && b.IsVisible()
}

// App returns the application pointer associated with the widget.
func (b *BaseWidget) App() *Application {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.app
}

// SetApplication stores the application pointer.
func (b *BaseWidget) SetApplication(app *Application) {
	b.mu.Lock()
	b.app = app
	b.mu.Unlock()
}

// Children returns nil for BaseWidget as it cannot contain children by default.
func (b *BaseWidget) Children() []Widget {
	return nil // Base widgets don't have children
}

// Parent returns the widget's container (parent).
func (b *BaseWidget) Parent() Widget {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.parent
}

// SetParent sets the widget's container (parent).
func (b *BaseWidget) SetParent(parent Widget) {
	b.mu.Lock()
	b.parent = parent
	b.mu.Unlock()
}

// SetKeybinding registers a handler function for a specific key combination.
func (b *BaseWidget) SetKeybinding(key tcell.Key, mod tcell.ModMask, handler func() bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.keyBindings == nil {
		b.keyBindings = make(map[keyModCombo]func() bool)
	}

	combo := keyModCombo{
		Key: key,
		Mod: mod,
	}
	b.keyBindings[combo] = handler
}

// IsVisible returns true if the widget's local visible flag is true
// AND its parent (if any) is also visible.
func (b *BaseWidget) IsVisible() bool {
	b.mu.RLock()
	isVisibleLocally := b.visible
	parentWidget := b.parent
	b.mu.RUnlock()

	if !isVisibleLocally {
		return false
	}
	// Check parent recursively (without holding lock on self)
	if parentWidget != nil {
		return parentWidget.IsVisible()
	}
	return true // No parent or parent is visible
}

// SetVisible sets the *local* visibility state of the widget.
// If visibility changes, it queues a redraw and handles focus loss if hiding.
func (b *BaseWidget) SetVisible(visible bool) {
	b.mu.Lock()
	changed := b.visible != visible
	b.visible = visible
	isCurrentlyFocused := b.focused
	app := b.app
	b.mu.Unlock()

	if changed {
		// If hiding a focused widget, blur it.
		// The application focus logic should handle moving focus away later.
		if !visible && isCurrentlyFocused {
			b.Blur()
		}
		if app != nil {
			app.QueueRedraw() // Redraw needed to show/hide
		}
	}
}

// PreferredWidth fallback if not implemented by concrete widget
func (b *BaseWidget) PreferredWidth() int {
	return 10 // Default fallback // TODO: Implement as constant
}

// PreferredHeight fallback if not implemented by concrete widget
func (b *BaseWidget) PreferredHeight() int {
	return 1 // Default fallback // TODO: Implement as constant
}