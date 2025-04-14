// event.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
)

// --- Command Pattern ---
// Commands provide a way to decouple event handlers/components from direct
// modification of application state or triggering complex actions. They are
// queued via app.Dispatch() and executed sequentially in the main event loop.

// Command defines the interface for actions executed by the Application.
type Command interface {
	Execute(app *Application)
}

// RedrawCommand signals that the entire UI needs a redraw.
type RedrawCommand struct{}

// Execute implements the Command interface.
func (c *RedrawCommand) Execute(app *Application) {
	app.queueRedraw() // Use the internal method for consistency
}

// FocusCommand requests focus to be set on the target component.
type FocusCommand struct {
	Target Component // The component to receive focus.
}

// Execute implements the Command interface.
func (c *FocusCommand) Execute(app *Application) {
	app.SetFocus(c.Target)
}

// UpdateTextCommand requests updating the content of a TextUpdater component.
type UpdateTextCommand struct {
	Target  TextUpdater // Component must implement TextUpdater.
	Content string      // The new text content.
}

// Execute implements the Command interface.
func (c *UpdateTextCommand) Execute(app *Application) {
	if c.Target != nil {
		c.Target.SetContent(c.Content) // Component's SetContent should handle MarkDirty.
	}
}

// UpdateGridCommand requests updating the cells of a Grid component.
type UpdateGridCommand struct {
	Target  *Grid      // The target Grid component.
	Content [][]string // The new cell data.
}

// Execute implements the Command interface.
func (c *UpdateGridCommand) Execute(app *Application) {
	if c.Target != nil {
		c.Target.SetCells(c.Content) // Component's SetCells should handle MarkDirty.
	}
}

// UpdateSpriteCommand requests updating the cells of a Sprite component.
type UpdateSpriteCommand struct {
	Target  *Sprite        // The target Sprite component.
	Content [][]SpriteCell // The new sprite cell data.
}

// Execute implements the Command interface.
func (c *UpdateSpriteCommand) Execute(app *Application) {
	if c.Target != nil {
		c.Target.SetCells(c.Content) // Component's SetCells should handle MarkDirty.
	}
}

// FindNextFocusCommand requests the application find a suitable component to focus.
// This is typically dispatched when the currently focused component is about to be
// hidden or removed, ensuring focus doesn't get lost.
type FindNextFocusCommand struct {
	origin Component // The component that triggered this request (e.g., by being hidden).
}

// Execute implements the Command interface.
func (c *FindNextFocusCommand) Execute(app *Application) {
	currentFocus := app.GetFocusedComponent()

	// Only proceed if focus is currently nil OR if the focus is still on the component
	// that triggered this command. If focus moved elsewhere already, do nothing.
	if currentFocus != nil && currentFocus != c.origin {
		return
	}

	// If the origin component *was* the focused one, explicitly set focus to nil
	// *before* searching for the next one. This prevents SetFocus(next) from
	// potentially blurring the origin component if it hasn't been removed yet.
	if currentFocus == c.origin {
		app.focusedComponent = nil // Directly set to nil, avoid calling SetFocus(nil)
	}

	// Find the first available focusable component in the entire layout.
	if app.layout != nil {
		focusables := app.layout.GetAllFocusableComponents()
		if len(focusables) > 0 {
			app.SetFocus(focusables[0]) // Set focus to the first one found
		}
		// If no focusable components are left, focus remains nil.
	}
}

// SimpleCommand allows dispatching an arbitrary Func function to be run in the main loop.
type SimpleCommand struct {
	Func func(app *Application)
}

// Execute implements the Command interface.
func (c *SimpleCommand) Execute(app *Application) {
	if c.Func != nil {
		c.Func(app)
	}
}

// AddPaneCommand requests adding a pane.
// The Layout.AddPane method itself now dispatches the RecalculateNavIndicesCommand.
type AddPaneCommand struct {
	Pane *Pane
	Size Size
}

func (c *AddPaneCommand) Execute(app *Application) {
	if app.layout != nil && c.Pane != nil {
		app.layout.AddPane(c.Pane, c.Size) // AddPane triggers recalculation command
	}
}

// RemovePaneCommand requests removing a pane by slot index.
// The Layout.RemovePane method itself now dispatches the RecalculateNavIndicesCommand.
type RemovePaneCommand struct{ Index int } // Index is Slot Index
func (c *RemovePaneCommand) Execute(app *Application) {
	if app.layout != nil {
		app.layout.RemovePane(c.Index) // RemovePane triggers recalculation command
	}
}

// RecalculateNavIndicesCommand signals the application to recalculate and assign
// navigation indices to the top-level panes in the root layout.
type RecalculateNavIndicesCommand struct{}

// Execute implements the Command interface.
func (c *RecalculateNavIndicesCommand) Execute(app *Application) {
	// Ensure layout exists and call assignNavigationIndices on the root layout
	if app.layout != nil {
		rootLayout := app.GetLayout() // Get the root layout instance
		if rootLayout != nil {
			rootLayout.assignNavigationIndices()
			app.QueueRedraw() // Redraw needed to show updated indices
		}
	}
}

// --- Key Handling Structures ---

// KeyModCombo represents a non-rune key + modifier combination used for keybindings.
type KeyModCombo struct {
	Key tcell.Key     // The specific key (e.g., tcell.KeyEnter, tcell.KeyTab).
	Mod tcell.ModMask // The modifier mask (e.g., tcell.ModAlt, tcell.ModCtrl).
}

// KeyHandler defines the function signature for handling registered key events (non-rune or specific runes).
// It should return true if the key event was handled (consumed), false otherwise.
type KeyHandler func() bool