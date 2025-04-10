// widget.go
package tinytui

import "github.com/gdamore/tcell/v2" // Keep tcell for Widget interface methods for now

// Rect defines a rectangular area on the screen.
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

// --- Add TextUpdater Interface ---

// TextUpdater defines an interface for widgets that can have their text content set.
type TextUpdater interface {
	SetContent(content string)
}

// --- End TextUpdater Interface ---

// Widget is the core interface for all drawable and interactive elements.
type Widget interface {
	// Draw renders the widget onto the provided screen within its bounds.
	// Note: screen is still tcell.Screen for now.
	Draw(screen tcell.Screen)

	// SetRect informs the widget of its allocated space. Widgets must respect these bounds.
	SetRect(x, y, width, height int)

	// GetRect returns the current bounds of the widget.
	GetRect() (x, y, width, height int)

	// HandleEvent processes a tcell event. It returns true if the event was
	// consumed and should stop propagation, false otherwise.
	// Note: event is still tcell.Event for now.
	HandleEvent(event tcell.Event) bool

	// Focusable returns true if the widget can receive keyboard focus.
	Focusable() bool

	// Focus is called when the widget gains focus.
	Focus()

	// Blur is called when the widget loses focus.
	Blur()

	// SetApplication links the widget back to the main application, primarily
	// for queuing redraws. This is typically called by the parent (layout or app).
	SetApplication(app *Application)

	// App returns the application pointer associated with the widget.
	// Returns nil if SetApplication has not been called.
	App() *Application

	// IsFocused returns whether the widget currently has focus.
	IsFocused() bool

	// Children returns a slice of the widget's immediate children.
	// This is used by the application for navigating the widget tree (e.g., focus).
	// Widgets that cannot contain children should return nil or an empty slice.
	Children() []Widget

	// Parent returns the widget's container (parent) in the hierarchy.
	// Returns nil if the widget is the root or has no parent set.
	Parent() Widget

	// SetParent establishes the link to the widget's container.
	// This is typically called by container widgets when adding a child.
	SetParent(parent Widget)

	// IsVisible returns the effective visibility state of the widget,
	// considering its own state and its parents' visibility.
	IsVisible() bool

	// SetVisible sets the local visibility state of the widget.
	SetVisible(visible bool)
}

// Note: Assumes Application is defined in app.go within the same package.