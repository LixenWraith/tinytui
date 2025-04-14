// component.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
)

// Component is the fundamental interface for all visual elements within a Pane.
// It defines methods for drawing, geometry management, event handling, focus, visibility, and state.
type Component interface {
	// Draw renders the component onto the screen within its allocated rectangle.
	// Implementations should respect the component's visibility and bounds.
	Draw(screen tcell.Screen)

	// SetRect informs the component of its allocated position and size (x, y, width, height).
	// Components should mark themselves dirty if the rectangle changes.
	SetRect(x, y, width, height int)

	// GetRect returns the component's current position and size.
	GetRect() (x, y, width, height int)

	// HandleEvent processes a terminal event (e.g., key press, mouse event).
	// Returns true if the event was handled by this component, false otherwise,
	// allowing the event to potentially bubble up or be handled globally.
	HandleEvent(event tcell.Event) bool

	// Focus is called by the application when the component gains input focus.
	// Implementations should update internal state and mark dirty if appearance changes.
	Focus()

	// Blur is called by the application when the component loses input focus.
	// Implementations should update internal state and mark dirty if appearance changes.
	Blur()

	// IsFocused returns true if the component currently has input focus.
	IsFocused() bool

	// Focusable returns true if the component is capable of receiving input focus
	// (e.g., interactive elements like TextInput, Grid). Non-focusable components
	// are skipped during focus cycling (Tab/Shift+Tab).
	Focusable() bool

	// IsVisible returns true if the component should be drawn and considered for layout/focus.
	IsVisible() bool

	// SetVisible sets the visibility state of the component.
	// Hidden components are not drawn and cannot be focused. Hiding a focused
	// component should trigger focus loss handling in the application.
	SetVisible(visible bool)

	// SetState sets the interaction state of the component (Normal, Selected, Interacted).
	// Used for visual feedback (e.g., highlighting selected grid cells).
	SetState(state State)

	// GetState returns the current interaction state of the component.
	GetState() State

	// SetApplication links the component to its parent application instance.
	// This allows components to dispatch commands or access application-level resources
	// like the theme or cursor manager. Usually called by the parent container.
	SetApplication(app *Application)

	// App returns the parent application instance, or nil if not set.
	App() *Application

	// MarkDirty flags the component as needing a redraw in the next draw cycle.
	// Implementations should ideally also notify the application via App().QueueRedraw()
	// to ensure the draw cycle runs.
	MarkDirty()

	// IsDirty returns true if the component has been flagged as needing a redraw.
	// Containers should override this to check their children recursively.
	IsDirty() bool

	// ClearDirty resets the dirty flag. Called by the application after drawing.
	// Containers should override this to clear flags recursively.
	ClearDirty()
}

// TextUpdater is an optional interface for components whose primary content
// can be updated programmatically via a string, often used with UpdateTextCommand.
type TextUpdater interface {
	Component
	// SetContent updates the main text content of the component.
	SetContent(content string)
}

// ThemedComponent is an optional interface for components that require custom logic
// to update their appearance when the application's theme changes. Components
// implementing this will have their ApplyTheme method called automatically when
// app.SetTheme() is used or when added to a layout within an application.
type ThemedComponent interface {
	Component
	// ApplyTheme updates the component's appearance (e.g., internal styles)
	// based on the properties of the provided theme.
	ApplyTheme(theme Theme)
}