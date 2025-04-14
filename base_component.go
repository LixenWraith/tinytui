// base_component.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
)

// BaseComponent provides default implementations for common Component methods.
// Concrete components should embed this struct to inherit baseline behavior.
type BaseComponent struct {
	rect    Rect         // Component position and size
	visible bool         // Is the component currently visible?
	focused bool         // Does the component have input focus?
	dirty   bool         // Does the component need to be redrawn?
	state   State        // Current interaction state (Normal, Selected, Interacted)
	app     *Application // Reference to the parent application
}

// NewBaseComponent creates a new BaseComponent with sensible defaults.
func NewBaseComponent() BaseComponent {
	return BaseComponent{
		visible: true,
		focused: false,
		dirty:   true, // Start dirty to ensure initial draw
		state:   StateNormal,
		// app is initially nil
	}
}

// SetRect sets the component's position and size.
// Marks the component as dirty if the rectangle changes.
func (b *BaseComponent) SetRect(x, y, width, height int) {
	newRect := Rect{X: x, Y: y, Width: width, Height: height}
	if b.rect != newRect {
		b.rect = newRect
		b.MarkDirty() // Geometry change requires redraw
	}
}

// GetRect returns the component's current position and size.
func (b *BaseComponent) GetRect() (x, y, width, height int) {
	return b.rect.X, b.rect.Y, b.rect.Width, b.rect.Height
}

// IsVisible returns whether the component is currently set to be visible.
func (b *BaseComponent) IsVisible() bool {
	return b.visible
}

// SetVisible sets the component's visibility state.
// If hiding a focused component, it dispatches a command to find a new focus target.
func (b *BaseComponent) SetVisible(visible bool) {
	if b.visible != visible {
		b.visible = visible
		b.MarkDirty() // Visibility change requires redraw

		// If hiding the currently focused component, it should lose focus
		// and the application should try to focus something else.
		if !visible && b.focused {
			b.focused = false // Lose focus state internally
			if b.app != nil {
				// Tell the app to find a new focus target, indicating this component lost focus due to hiding
				b.app.Dispatch(&FindNextFocusCommand{origin: b})
			}
		}
		// If making visible, it might become focusable, but focus doesn't automatically move to it.
	}
}

// Focus is called by the application when the component gains input focus. Marks the component dirty.
func (b *BaseComponent) Focus() {
	if !b.focused {
		b.focused = true
		b.MarkDirty() // Need redraw to reflect focused state (style, cursor)
	}
}

// Blur is called by the application when the component loses input focus. Marks the component dirty.
func (b *BaseComponent) Blur() {
	if b.focused {
		b.focused = false
		b.MarkDirty() // Need redraw to reflect unfocused state
	}
}

// IsFocused returns whether the component currently has input focus.
func (b *BaseComponent) IsFocused() bool {
	return b.focused
}

// Focusable returns whether the component can receive input focus.
// Default implementation: focusable only if visible.
// Concrete components (like TextInput, Grid) override this with more specific logic.
func (b *BaseComponent) Focusable() bool {
	return b.visible
}

// SetState sets the component's interaction state (Normal, Selected, Interacted).
// Marks the component dirty if the state changes, as appearance might depend on state.
func (b *BaseComponent) SetState(state State) {
	if b.state != state {
		b.state = state
		b.MarkDirty()
	}
}

// GetState returns the component's current interaction state.
func (b *BaseComponent) GetState() State {
	return b.state
}

// SetApplication sets the application instance the component belongs to.
// This is typically called by the parent container (Pane or Layout) during setup.
func (b *BaseComponent) SetApplication(app *Application) {
	b.app = app
}

// App returns the application instance the component belongs to, or nil if not set.
func (b *BaseComponent) App() *Application {
	return b.app
}

// MarkDirty flags the component as needing a redraw in the next draw cycle.
// It also queues a redraw request with the application if the component is part of one.
func (b *BaseComponent) MarkDirty() {
	b.dirty = true
	// If component is linked to an application, signal the app that *something* needs redrawing.
	if b.app != nil {
		b.app.QueueRedraw()
	}
}

// IsDirty returns whether the component is flagged as needing a redraw.
// This checks the component's own flag, not its children. Containers override this.
func (b *BaseComponent) IsDirty() bool {
	return b.dirty
}

// ClearDirty marks the component as clean (no redraw needed for itself).
// This is typically called by the application's drawing logic after the component has been drawn.
func (b *BaseComponent) ClearDirty() {
	b.dirty = false
}

// HandleEvent provides a default event handler implementation.
// Base implementation does nothing and indicates the event was not handled.
// Concrete components override this to process specific events (e.g., key presses).
func (b *BaseComponent) HandleEvent(event tcell.Event) bool {
	return false // Event not handled by default
}

// Draw provides a default drawing implementation.
// Base implementation does nothing, as base components have no visual representation.
// Concrete components override this to draw their content onto the screen.
func (b *BaseComponent) Draw(screen tcell.Screen) {
	// Base component doesn't draw anything itself.
}