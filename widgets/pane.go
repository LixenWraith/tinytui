// widgets/pane.go
package widgets

import (
	"sync"

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
		style:       tinytui.DefaultPaneStyle(),
		border:      false,
		borderType:  tinytui.BorderSingle, // Default to single border type
		borderStyle: tinytui.DefaultPaneBorderStyle(),
	}
	// Set a default focus border style from theme
	p.focusBorderStyle = tinytui.DefaultPaneFocusBorderStyle()
	p.originalBorderStyle = p.borderStyle // Initialize original style
	p.SetVisible(true)                    // Explicitly set visibility
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

// ApplyTheme applies the provided theme to the Pane widget
// while preserving the current border type
func (p *Pane) ApplyTheme(theme tinytui.Theme) {
	p.mu.Lock()

	// Apply content style
	p.style = theme.PaneStyle()

	// Get current border settings
	hasBorder := p.border

	if hasBorder {
		// Update only the styles, preserve the current border type unless focused
		p.borderStyle = theme.PaneBorderStyle()
		p.originalBorderStyle = theme.PaneBorderStyle()

		// Don't update borderType here - that's set by SetBorder and used conditionally during drawing
	}

	// Update focus border style
	p.focusBorderStyle = theme.PaneFocusBorderStyle()

	p.mu.Unlock()

	// Queue redraw to apply changes
	if app := p.App(); app != nil {
		app.QueueRedraw()
	}
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
	p.originalBorderStyle = style

	// Update focus style background to match new border style bg, keep focus foreground
	_, bg, _, _ := style.Deconstruct()
	p.focusBorderStyle = p.focusBorderStyle.Background(bg)
	p.mu.Unlock() // Unlock after border changes

	if app := p.App(); app != nil {
		app.QueueRedraw()
	}
	return p
}

func (p *Pane) HasBorder() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.border
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
// Updated to handle focus visuals based on child focus state
func (p *Pane) Draw(screen tcell.Screen) {
	p.BaseWidget.Draw(screen)

	x, y, width, height := p.GetRect()
	if width <= 0 || height <= 0 {
		return
	}

	p.mu.RLock() // RLock for reading styles and border settings
	currentBorderStyle := p.borderStyle
	borderEnabled := p.border
	bType := p.borderType
	contentStyle := p.style
	childWidget := p.child // Read child under lock

	// Check if any child is focused or if any descendant is focused
	childFocused := false
	if childWidget != nil {
		childFocused = childWidget.IsFocused() || hasAnyFocusedDescendant(childWidget)
	}
	p.mu.RUnlock() // Unlock after reading state

	// Use focused border style if any child is focused
	if childFocused && borderEnabled {
		currentBorderStyle = p.focusBorderStyle

		// Get focused border type from app theme
		if app := p.App(); app != nil {
			if theme := app.GetTheme(); theme != nil {
				bType = theme.FocusedBorderType()
			}
		}
	}

	// Calculate content area with proper border handling
	contentX, contentY, contentWidth, contentHeight := x, y, width, height

	// Apply border adjustments if enabled and dimensions are sufficient
	if borderEnabled && bType != tinytui.BorderNone && width > 1 && height > 1 {
		contentX++
		contentY++
		contentWidth -= 2
		contentHeight -= 2
	}

	// Ensure content dimensions are valid
	if contentWidth < 0 {
		contentWidth = 0
	}
	if contentHeight < 0 {
		contentHeight = 0
	}

	// Draw border if enabled and there's enough space
	if borderEnabled && bType != tinytui.BorderNone && width > 1 && height > 1 {
		switch bType {
		case tinytui.BorderSingle:
			tinytui.DrawBox(screen, x, y, width, height, currentBorderStyle)
		case tinytui.BorderDouble:
			tinytui.DrawDoubleBox(screen, x, y, width, height, currentBorderStyle)
		case tinytui.BorderSolid:
			tinytui.DrawSolidBox(screen, x, y, width, height, currentBorderStyle)
		}
	}

	// Fill content area if dimensions are valid
	if contentWidth > 0 && contentHeight > 0 {
		tinytui.Fill(screen, contentX, contentY, contentWidth, contentHeight, ' ', contentStyle)
	}

	// Draw Child within content area if there's space
	if childWidget != nil && contentWidth > 0 && contentHeight > 0 {
		childWidget.SetRect(contentX, contentY, contentWidth, contentHeight)
		childWidget.Draw(screen)
	}
}

// hasAnyFocusedDescendant recursively checks if a widget or any of its children have focus.
func hasAnyFocusedDescendant(w tinytui.Widget) bool {
	if w == nil {
		return false
	}
	// Check if the widget itself is focused
	if w.IsFocused() {
		return true
	}
	// Recursively check children
	if children := w.Children(); children != nil {
		for _, child := range children {
			if hasAnyFocusedDescendant(child) {
				return true // Found a focused descendant
			}
		}
	}
	// No focused descendant found in this branch
	return false
}

// Focusable indicates if the Pane itself should receive focus.
// It's only focusable if it contains no focusable descendants
func (p *Pane) Focusable() bool {
	// Panes are never directly focusable in tab navigation
	return false
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

	// Calculate content area
	contentX, contentY, contentWidth, contentHeight := x, y, width, height

	// If border is enabled, reserve space for it
	if borderEnabled && bType != tinytui.BorderNone && width > 1 && height > 1 {
		contentX++
		contentY++
		contentWidth -= 2
		contentHeight -= 2
	}

	// Ensure dimensions are non-negative and enforce minimum size
	if contentWidth < 0 {
		contentWidth = 0
	}
	if contentHeight < 0 {
		contentHeight = 0
	}

	// Set child widget dimensions, strictly enforcing the content area bounds
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