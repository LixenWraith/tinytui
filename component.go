// component.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
)

// Component represents any displayable element within a Pane.
type Component interface {
	// Core rendering & geometry
	Draw(screen tcell.Screen)
	SetRect(x, y, width, height int)
	GetRect() (x, y, width, height int)

	// Event handling
	HandleEvent(event tcell.Event) bool

	// Focus management
	Focus() // Called when component gains focus
	Blur()  // Called when component loses focus
	IsFocused() bool
	Focusable() bool // Can this component receive focus?

	// Visibility
	IsVisible() bool
	SetVisible(visible bool)

	// State management
	SetState(state State)
	GetState() State

	// Application linkage
	SetApplication(app *Application)
	App() *Application // Getter for convenience

	// Dirty flag for optimized rendering
	MarkDirty()
	IsDirty() bool
	ClearDirty()
}

// TextUpdater is an interface for components that can have their text content set.
type TextUpdater interface {
	Component
	SetContent(content string)
}