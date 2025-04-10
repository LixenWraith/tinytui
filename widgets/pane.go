// widgets/pane.go
package widgets

import (
	"sync" // Added sync back if needed by BaseWidget or other parts

	"github.com/LixenWraith/tinytui"
	"github.com/gdamore/tcell/v2"
)

// Pane is a simple container widget with optional borders and styling.
// It can hold a single child widget.
type Pane struct {
	tinytui.BaseWidget
	style       tinytui.Style      // Style for the content area
	border      bool               // Whether to draw a border
	borderType  tinytui.BorderType // Type of border lines
	borderStyle tinytui.Style      // Style for the border itself

	// Store the original border style to restore on Blur
	originalBorderStyle tinytui.Style
	focusBorderStyle    tinytui.Style  // Style for the border when focused
	child               tinytui.Widget // The single child widget

	// Added mutex for child access, although most access is now through methods
	// Consider if BaseWidget's mutex is sufficient or if child needs separate protection
	mu sync.RWMutex
}

// NewPane creates a new Pane widget.
func NewPane() *Pane {
	p := &Pane{
		style:      tinytui.DefaultStyle,
		border:     false,
		borderType: tinytui.BorderNone,
		// Default border style is the same as content style initially
		borderStyle: tinytui.DefaultStyle,
	}
	// Set a default focus border style (e.g., bright yellow foreground)
	p.focusBorderStyle = tinytui.DefaultStyle.Foreground(tinytui.ColorYellow).Bold(true)
	p.originalBorderStyle = p.borderStyle // Initialize original style
	p.SetVisible(true)                    // Panes are visible by default
	return p
}

// SetStyle sets the background and foreground style of the pane's content area.
func (p *Pane) SetStyle(style tinytui.Style) *Pane {
	p.mu.Lock() // Lock for style changes
	p.style = style
	// If border style hasn't been explicitly set, keep it matching the content style
	if p.borderStyle == p.originalBorderStyle { // Check if it's still the default/previous content style
		p.borderStyle = style
		p.originalBorderStyle = style // Update original to match new content style

		// Update focus style background to match new content style, keep focus foreground
		_, bg, _, _ := style.Deconstruct()
		p.focusBorderStyle = p.focusBorderStyle.Background(bg)
	}
	p.mu.Unlock() // Unlock after style changes

	if app := p.App(); app != nil {
		app.QueueRedraw()
	}
	return p
}

// SetBorder configures the pane's border.
// - enabled: Show or hide the border.
// - borderType: The style of the border lines (Single, Double, etc.).
// - style: The style (colors, attributes) for the border lines.
func (p *Pane) SetBorder(enabled bool, borderType tinytui.BorderType, style tinytui.Style) *Pane {
	p.mu.Lock() // Lock for border changes
	p.border = enabled
	p.borderType = borderType
	p.borderStyle = style
	p.originalBorderStyle = style // Store the explicitly set border style

	// Update focus style background to match new border style bg, keep focus foreground
	_, bg, _, _ := style.Deconstruct()
	p.focusBorderStyle = p.focusBorderStyle.Background(bg)
	p.mu.Unlock() // Unlock after border changes

	if app := p.App(); app != nil {
		app.QueueRedraw()
	}
	return p
}

// SetFocusBorderStyle allows customizing the border appearance when the pane is focused.
func (p *Pane) SetFocusBorderStyle(style tinytui.Style) *Pane {
	p.mu.Lock() // Lock for style changes
	p.focusBorderStyle = style
	// Ensure the background matches the original border style's background
	// unless explicitly set otherwise in the provided style.
	_, _, _, bgSet := style.Deconstruct()
	if !bgSet {
		_, origBg, _, _ := p.originalBorderStyle.Deconstruct()
		p.focusBorderStyle = p.focusBorderStyle.Background(origBg)
	}
	isFocused := p.IsFocused() // Read focus state under lock
	p.mu.Unlock()              // Unlock after style changes

	if isFocused && p.App() != nil {
		p.App().QueueRedraw() // Redraw if currently focused
	}
	return p
}

// Draw draws the pane, including its border, background, and child widget.
func (p *Pane) Draw(screen tcell.Screen) {
	// Use BaseWidget's visibility check first
	if !p.IsVisible() {
		return
	}

	x, y, width, height := p.GetRect()
	if width <= 0 || height <= 0 {
		return
	}

	p.mu.RLock() // RLock for reading styles and border settings
	currentBorderStyle := p.borderStyle
	borderEnabled := p.border
	bType := p.borderType
	contentStyle := p.style
	childWidget := p.child     // Read child under lock
	isFocused := p.IsFocused() // Read focus state under lock

	if isFocused && borderEnabled { // Use focus style only if border is enabled
		currentBorderStyle = p.focusBorderStyle
		// Ensure focus background matches original if not explicitly set in focus style
		_, _, _, bgSet := currentBorderStyle.Deconstruct()
		if !bgSet {
			_, origBg, _, _ := p.originalBorderStyle.Deconstruct()
			currentBorderStyle = currentBorderStyle.Background(origBg)
		}
	}
	p.mu.RUnlock() // Unlock after reading state

	// Calculate content area first
	contentX, contentY, contentWidth, contentHeight := x, y, width, height
	if borderEnabled && bType != tinytui.BorderNone && width > 1 && height > 1 {
		contentX++
		contentY++
		contentWidth -= 2
		contentHeight -= 2
	}

	// Draw border if enabled (uses original dimensions)
	if borderEnabled && bType != tinytui.BorderNone && width > 0 && height > 0 {
		switch bType {
		case tinytui.BorderSingle:
			tinytui.DrawBox(screen, x, y, width, height, currentBorderStyle)
		case tinytui.BorderDouble:
			tinytui.DrawDoubleBox(screen, x, y, width, height, currentBorderStyle)
		case tinytui.BorderSolid:
			tinytui.DrawSolidBox(screen, x, y, width, height, currentBorderStyle)
			// Add other cases if more border types are defined
		}
	}

	// Fill content area if dimensions are valid
	if contentWidth > 0 && contentHeight > 0 {
		tinytui.Fill(screen, contentX, contentY, contentWidth, contentHeight, ' ', contentStyle)
	}

	// Draw Child (within content area)
	if childWidget != nil {
		// Child's SetRect was called in Pane.SetRect
		childWidget.Draw(screen) // Draw the child read under lock
	}
}

// hasFocusableDescendant recursively checks if a widget or any of its children are focusable.
func hasFocusableDescendant(w tinytui.Widget) bool {
	if w == nil {
		return false
	}
	// Check if the widget itself is focusable AND visible
	// (An invisible widget, even if focusable, shouldn't count)
	if w.Focusable() && w.IsVisible() {
		return true
	}
	// Recursively check children
	if children := w.Children(); children != nil {
		for _, child := range children {
			if hasFocusableDescendant(child) {
				return true // Found a focusable descendant
			}
		}
	}
	// No focusable descendant found in this branch
	return false
}

// Focusable indicates if the Pane itself should receive focus.
// It should only be focusable if it's visible AND it does not contain
// any focusable descendants.
func (p *Pane) Focusable() bool {
	p.mu.RLock()
	childWidget := p.child
	p.mu.RUnlock()

	isVisible := p.IsVisible() // Check visibility using BaseWidget method

	// Pane is focusable only if it's visible AND has no focusable descendants
	return isVisible && !hasFocusableDescendant(childWidget)
}

// Focus is called when the pane gains focus.
func (p *Pane) Focus() {
	p.BaseWidget.Focus() // Handle internal state and redraw queue
}

// Blur is called when the pane loses focus.
func (p *Pane) Blur() {
	p.BaseWidget.Blur() // Handle internal state and redraw queue
}

// HandleEvent handles events for the Pane.
// If the Pane itself is focusable (meaning no focusable children),
// it might handle events like scrolling in the future.
// Currently, it just delegates to BaseWidget.
func (p *Pane) HandleEvent(event tcell.Event) bool {
	// Let BaseWidget check its keybindings first
	if p.BaseWidget.HandleEvent(event) {
		return true
	}
	// If the pane itself is focused (only possible if Focusable() returned true),
	// add pane-specific event handling here (e.g., scrolling).
	// For now, we don't have any pane-specific actions.
	return false
}

// --- Container Methods ---

// SetChild sets the single child widget contained within the pane.
func (p *Pane) SetChild(widget tinytui.Widget) *Pane {
	p.mu.Lock() // Lock for modifying child
	p.child = widget
	p.mu.Unlock() // Unlock after modifying child

	if widget != nil {
		widget.SetParent(p)
		appInstance := p.App() // Get app instance once
		if appInstance != nil {
			widget.SetApplication(appInstance)
		}
		// Update child's rect immediately if pane already has one
		// GetRect() is safe here as it reads the BaseWidget's rect field
		x, y, w, h := p.GetRect()
		p.SetRect(x, y, w, h) // Call SetRect to recalculate child bounds
	}

	if app := p.App(); app != nil {
		app.QueueRedraw() // Child change requires redraw
	}
	return p
}

// SetRect sets the pane's rectangle and calculates the child's rectangle.
func (p *Pane) SetRect(x, y, width, height int) {
	p.BaseWidget.SetRect(x, y, width, height) // Set own rect first

	p.mu.RLock() // RLock for reading border settings and child
	borderEnabled := p.border
	bType := p.borderType
	childWidget := p.child
	p.mu.RUnlock() // Unlock after reading

	contentX, contentY, contentWidth, contentHeight := x, y, width, height
	if borderEnabled && bType != tinytui.BorderNone && width > 1 && height > 1 {
		contentX++
		contentY++
		contentWidth -= 2
		contentHeight -= 2
	}

	// Ensure dimensions are non-negative
	if contentWidth < 0 {
		contentWidth = 0
	}
	if contentHeight < 0 {
		contentHeight = 0
	}

	if childWidget != nil {
		childWidget.SetRect(contentX, contentY, contentWidth, contentHeight)
	}
}

// Children returns the single child widget in a slice, or nil.
// This is needed for focus traversal by the Application.
func (p *Pane) Children() []tinytui.Widget {
	p.mu.RLock() // RLock for reading child
	defer p.mu.RUnlock()
	if p.child != nil {
		return []tinytui.Widget{p.child}
	}
	return nil
}

// SetApplication propagates the application instance to the child.
func (p *Pane) SetApplication(app *tinytui.Application) {
	p.BaseWidget.SetApplication(app) // Set on BaseWidget first

	p.mu.RLock() // RLock for reading child
	childWidget := p.child
	p.mu.RUnlock() // Unlock after reading

	if childWidget != nil {
		childWidget.SetApplication(app)
	}
}