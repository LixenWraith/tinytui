// event.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
)

// Command pattern for async operations
type Command interface {
	Execute(app *Application)
}

// RedrawCommand signals that a redraw is needed
type RedrawCommand struct{}

func (c *RedrawCommand) Execute(app *Application) {
	app.queueRedraw()
}

// FocusCommand changes focus to the target component
type FocusCommand struct {
	Target Component
}

func (c *FocusCommand) Execute(app *Application) {
	app.SetFocus(c.Target)
}

// UpdateTextCommand updates text content of a TextUpdater component
type UpdateTextCommand struct {
	Target  TextUpdater
	Content string
}

func (c *UpdateTextCommand) Execute(app *Application) {
	if c.Target != nil {
		c.Target.SetContent(c.Content)
	}
}

// UpdateGridCommand updates grid content
type UpdateGridCommand struct {
	Target  *Grid
	Content [][]string
}

func (c *UpdateGridCommand) Execute(app *Application) {
	if c.Target != nil {
		c.Target.SetCells(c.Content)
	}
}

// UpdateSpriteCommand updates sprite content
type UpdateSpriteCommand struct {
	Target  *Sprite
	Content [][]SpriteCell
}

func (c *UpdateSpriteCommand) Execute(app *Application) {
	if c.Target != nil {
		c.Target.SetCells(c.Content)
	}
}

// AddPaneCommand adds a pane to the layout
type AddPaneCommand struct {
	Pane *Pane
	Size Size
}

func (c *AddPaneCommand) Execute(app *Application) {
	if app.layout != nil {
		app.layout.AddPane(c.Pane, c.Size)
	}
}

// RemovePaneCommand removes a pane from the layout
type RemovePaneCommand struct {
	Index int // 0-9 array index
}

func (c *RemovePaneCommand) Execute(app *Application) {
	if app.layout != nil {
		app.layout.RemovePane(c.Index)
	}
}

// FindNextFocusCommand attempts to find the next focusable component
// after the specified origin loses focus
type FindNextFocusCommand struct {
	origin Component // Component that lost focus
}

func (c *FindNextFocusCommand) Execute(app *Application) {
	// Skip if the origin is nil
	if c.origin == nil {
		return
	}

	// If we have a focused component, and it's not the origin,
	// no need to find a new focus
	if app.focusedComponent != nil && app.focusedComponent != c.origin {
		return
	}

	// Clear current focus if it's the origin
	if app.focusedComponent == c.origin {
		app.focusedComponent = nil
	}

	// Find the next focusable component
	if app.layout != nil {
		focusables := app.layout.GetAllFocusableComponents()
		if len(focusables) > 0 {
			app.SetFocus(focusables[0])
		}
	}
}

// KeyModCombo represents a key+modifier combination for keybindings.
type KeyModCombo struct {
	Key tcell.Key
	Mod tcell.ModMask
}

// KeyHandler handles key events.
type KeyHandler func() bool

// ProcessEvent is called by the application to handle a tcell event.
func (app *Application) ProcessEvent(ev tcell.Event) {
	// Key event handling
	if keyEvent, ok := ev.(*tcell.EventKey); ok {
		// Escape key or Ctrl+C to quit
		if keyEvent.Key() == tcell.KeyEscape || keyEvent.Key() == tcell.KeyCtrlC {
			app.Stop()
			return
		}

		// Check Alt+number for pane navigation
		if keyEvent.Modifiers()&tcell.ModAlt != 0 {
			if r := keyEvent.Rune(); r >= '1' && r <= '9' {
				index := int(r - '1') // Convert 1-9 to 0-8
				app.handleAltNumberNavigation(index)
				return
			} else if r == '0' {
				// Handle Alt+0 as 10th pane (array index 9)
				app.handleAltNumberNavigation(9)
				return
			}
		}
		// Safely get reference to focused component
		var focusedComp Component
		var runeHandlersCopy []func(*tcell.EventKey) bool
		var keyHandler KeyHandler
		var hasHandler bool

		focusedComp = app.focusedComponent

		// Make a copy of rune handlers if needed
		if keyEvent.Key() == tcell.KeyRune {
			runeHandlersCopy = make([]func(*tcell.EventKey) bool, len(app.runeHandlers))
			copy(runeHandlersCopy, app.runeHandlers)
		}

		// Check if the key combo has a global handler
		if keyEvent.Key() != tcell.KeyRune {
			combo := KeyModCombo{
				Key: keyEvent.Key(),
				Mod: keyEvent.Modifiers(),
			}
			keyHandler, hasHandler = app.keyHandlers[combo]
		}

		// Check for rune handlers outside of lock
		if keyEvent.Key() == tcell.KeyRune {
			for _, handler := range runeHandlersCopy {
				if handler(keyEvent) {
					return // Event handled
				}
			}
		}

		// Execute key handler outside of lock
		if hasHandler {
			if keyHandler() {
				return // Event handled
			}
		}

		// Let focused component handle the event outside of lock
		if focusedComp != nil {
			if focusedComp.HandleEvent(ev) {
				return // Event handled by component
			}
		}

		// Tab/Shift+Tab for focus navigation
		if keyEvent.Key() == tcell.KeyTab {
			app.cycleFocus(true) // Forward
			return
		} else if keyEvent.Key() == tcell.KeyBacktab {
			app.cycleFocus(false) // Backward
			return
		}
	}

	// Window resize event
	if resizeEvent, ok := ev.(*tcell.EventResize); ok {
		app.handleResize(resizeEvent)
		return
	}
}