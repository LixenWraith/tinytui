// base_component.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
)

// BaseComponent provides default implementations for common Component methods.
// Concrete components embed this struct.
type BaseComponent struct {
	rect    Rect
	visible bool
	focused bool
	dirty   bool
	state   State
	app     *Application
}

// NewBaseComponent creates a new BaseComponent with sensible defaults.
func NewBaseComponent() BaseComponent {
	return BaseComponent{
		visible: true,
		focused: false,
		dirty:   true, // Start dirty to ensure initial draw
		state:   StateNormal,
	}
}

// SetRect sets the component's position and size.
func (b *BaseComponent) SetRect(x, y, width, height int) {
	newRect := Rect{X: x, Y: y, Width: width, Height: height}
	if b.rect != newRect {
		b.rect = newRect
		b.dirty = true // Geometry change requires redraw
	}
}

// GetRect returns the component's position and size.
func (b *BaseComponent) GetRect() (x, y, width, height int) {
	return b.rect.X, b.rect.Y, b.rect.Width, b.rect.Height
}

// IsVisible returns whether the component is visible.
func (b *BaseComponent) IsVisible() bool {
	return b.visible
}

// SetVisible sets the component's visibility.
func (b *BaseComponent) SetVisible(visible bool) {
	if b.visible != visible {
		b.visible = visible
		b.dirty = true

		// If hiding a focused component, it should lose focus
		if !visible && b.focused {
			b.focused = false
			app := b.app

			// Signal app to find new focus if this component was focused
			if app != nil {
				app.Dispatch(&FindNextFocusCommand{origin: b})
			}
		}
	}
}

// Focus is called when the component gains focus.
func (b *BaseComponent) Focus() {
	if !b.focused {
		b.focused = true
		b.dirty = true
	}
}

// Blur is called when the component loses focus.
func (b *BaseComponent) Blur() {
	if b.focused {
		b.focused = false
		b.dirty = true
	}
}

// IsFocused returns whether the component has focus.
func (b *BaseComponent) IsFocused() bool {
	return b.focused
}

// Focusable returns whether the component can receive focus.
// Default implementation: focusable only if visible.
// Override in concrete components as needed.
func (b *BaseComponent) Focusable() bool {
	return b.visible
}

// SetState sets the component's state.
func (b *BaseComponent) SetState(state State) {
	if b.state != state {
		b.state = state
		b.dirty = true
	}
}

// GetState returns the component's state.
func (b *BaseComponent) GetState() State {
	return b.state
}

// SetApplication sets the application the component belongs to.
func (b *BaseComponent) SetApplication(app *Application) {
	b.app = app
}

// App returns the application the component belongs to.
func (b *BaseComponent) App() *Application {
	return b.app
}

// MarkDirty marks the component as needing redraw.
func (b *BaseComponent) MarkDirty() {
	b.dirty = true
	app := b.app

	if app != nil {
		app.QueueRedraw()
	}
}

// IsDirty returns whether the component needs redraw.
func (b *BaseComponent) IsDirty() bool {
	return b.dirty
}

// ClearDirty marks the component as clean (no redraw needed).
func (b *BaseComponent) ClearDirty() {
	b.dirty = false
}

// HandleEvent handles an event for the component.
// Default implementation does nothing.
func (b *BaseComponent) HandleEvent(event tcell.Event) bool {
	return false
}

// Draw draws the component.
// Default implementation does nothing.
func (b *BaseComponent) Draw(screen tcell.Screen) {
	// Base implementation does nothing
	// Component-specific implementations will override this
}