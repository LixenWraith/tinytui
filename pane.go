// pane.go
package tinytui

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
)

// Pane is a container with optional border.
type Pane struct {
	child              interface{} // Can hold Component or *Layout
	border             Border
	title              string
	index              int // User-facing index (1-10), 0 if unset
	rect               Rect
	style              Style
	borderStyle        Style
	focusBorderStyle   Style
	showIndexIndicator bool
	app                *Application
	dirty              bool
}

// NewPane creates a new pane with default settings.
func NewPane() *Pane {
	return &Pane{
		border:             BorderSingle,
		style:              DefaultStyle,
		borderStyle:        DefaultStyle,
		focusBorderStyle:   DefaultStyle.Foreground(ColorYellow).Bold(true),
		showIndexIndicator: true,
		dirty:              true,
	}
}

// SetChild sets the pane's child (component or layout).
func (p *Pane) SetChild(child interface{}) {
	// Validate child type
	switch child.(type) {
	case Component, *Layout, nil:
		// Valid types
	default:
		// Invalid type
		return
	}

	// If child is a component, set its application
	if comp, ok := child.(Component); ok && p.app != nil {
		comp.SetApplication(p.app)
	}

	// If child is a layout, set its application
	if layout, ok := child.(*Layout); ok && p.app != nil {
		layout.SetApplication(p.app)
	}

	p.child = child
	p.dirty = true
}

// SetApplication sets the application this pane belongs to.
func (p *Pane) SetApplication(app *Application) {
	p.app = app

	// Propagate to child if it's a component
	if comp, ok := p.child.(Component); ok && comp != nil {
		comp.SetApplication(app)
	}

	// Propagate to child if it's a layout
	if layout, ok := p.child.(*Layout); ok && layout != nil {
		layout.SetApplication(app)
	}
}

// SetBorder sets the pane's border type and style.
func (p *Pane) SetBorder(border Border, style Style) {
	p.border = border
	p.borderStyle = style
	p.dirty = true
}

// SetFocusBorderStyle sets the style used for the border when a child has focus.
func (p *Pane) SetFocusBorderStyle(style Style) {
	p.focusBorderStyle = style
	p.dirty = true
}

// SetTitle sets the pane's title.
func (p *Pane) SetTitle(title string) {
	p.title = title
	p.dirty = true
}

// SetStyle sets the pane's background style.
func (p *Pane) SetStyle(style Style) {
	p.style = style
	p.dirty = true
}

// SetRect sets the pane's position and size.
func (p *Pane) SetRect(x, y, width, height int) {

	// Only process if dimensions changed
	if p.rect.X == x && p.rect.Y == y && p.rect.Width == width && p.rect.Height == height {
		return
	}

	p.rect = Rect{X: x, Y: y, Width: width, Height: height}
	p.dirty = true

	// Calculate content area (accounting for border)
	contentX, contentY, contentWidth, contentHeight := p.getContentRect()

	// Set child's dimensions
	if comp, ok := p.child.(Component); ok && comp != nil {
		comp.SetRect(contentX, contentY, contentWidth, contentHeight)
	} else if layout, ok := p.child.(*Layout); ok && layout != nil {
		layout.SetRect(contentX, contentY, contentWidth, contentHeight)
	}

}

// getContentRect calculates the inner content rectangle, accounting for borders.
func (p *Pane) getContentRect() (x, y, width, height int) {
	rect := p.rect
	border := p.border

	x, y = rect.X, rect.Y
	width, height = rect.Width, rect.Height

	// Adjust for border if present and size is sufficient
	if border != BorderNone && width > 2 && height > 2 {
		x += 1
		y += 1
		width -= 2
		height -= 2
	}

	return x, y, width, height
}

// SetIndex sets the pane's user-facing index (1-10).
// TODO: future optional autoindex feature, race condition fix with autoindex (do not make it internal)
func (p *Pane) SetIndex(index int) {
	p.index = index
	p.dirty = true
}

// GetIndex returns the pane's user-facing index (1-10).
func (p *Pane) GetIndex() int {
	return p.index
}

// SetShowIndexIndicator sets whether the pane's index should be shown.
func (p *Pane) SetShowIndexIndicator(show bool) {
	p.showIndexIndicator = show
	p.dirty = true
}

// Draw draws the pane and its child.
func (p *Pane) Draw(screen tcell.Screen, isFocused bool) {
	// Skip if outside screen bounds or too small
	if p.rect.Width <= 0 || p.rect.Height <= 0 {
		return
	}

	// Copy all needed data for drawing, avoid holding lock while drawing
	rect := p.rect
	border := p.border
	title := p.title
	index := p.index
	showIndexIndicator := p.showIndexIndicator
	style := p.style
	borderStyle := p.borderStyle
	focusBorderStyle := p.focusBorderStyle
	child := p.child

	// Clear dirty flag before drawing child
	p.dirty = false

	// Determine border style based on focus
	currentBorderStyle := borderStyle
	if isFocused {
		currentBorderStyle = focusBorderStyle
	}

	// Ensure pane has minimum size for border
	if border != BorderNone && (rect.Width < 3 || rect.Height < 3) {
		border = BorderNone
	}

	// Draw background
	Fill(screen, rect.X, rect.Y, rect.Width, rect.Height, ' ', style)

	// Draw border if present and size is sufficient
	if border != BorderNone && rect.Width > 2 && rect.Height > 2 {
		switch border {
		case BorderSingle:
			DrawBox(screen, rect.X, rect.Y, rect.Width, rect.Height, currentBorderStyle)
		case BorderDouble:
			DrawDoubleBox(screen, rect.X, rect.Y, rect.Width, rect.Height, currentBorderStyle)
		case BorderSolid:
			DrawSolidBox(screen, rect.X, rect.Y, rect.Width, rect.Height, currentBorderStyle)
		}

		// Draw title if present
		if title != "" {
			titleX := rect.X + 2
			titleMaxWidth := rect.Width - 4

			if titleMaxWidth > 0 {
				titleText := title
				if len(titleText) > titleMaxWidth {
					titleText = titleText[:titleMaxWidth]
				}

				DrawText(screen, titleX, rect.Y, currentBorderStyle, titleText)
			}
		}

		// Draw index indicator if present and enabled
		if index > 0 && showIndexIndicator && p.GetFirstFocusableComponent() != nil {
			// Format index as string
			indexStr := strconv.Itoa(index)
			if index == 10 {
				indexStr = "0" // Use 0 to represent 10 (for Alt+0)
			}

			// Position at top-right inside border
			indexX := rect.X + rect.Width - 2
			indexY := rect.Y

			// Draw index only if there's enough space
			if indexX > rect.X+1 && indexX < rect.X+rect.Width-1 {
				DrawText(screen, indexX, indexY, currentBorderStyle, indexStr)
			}
		}
	}

	// Calculate content area (accounting for border)
	contentX, contentY, contentWidth, contentHeight := p.getContentRect()

	// Draw child if present - outside of lock
	if child != nil {
		if comp, ok := child.(Component); ok && comp != nil {
			// Update component position before drawing
			comp.SetRect(contentX, contentY, contentWidth, contentHeight)
			comp.Draw(screen)
		} else if layout, ok := child.(*Layout); ok && layout != nil {
			// Update layout position before drawing
			layout.SetRect(contentX, contentY, contentWidth, contentHeight)
			layout.Draw(screen)
		}
	}
}

// drawIndexIndicator draws the pane index in the top-right corner.
func (p *Pane) drawIndexIndicator(screen tcell.Screen, isFocused bool) {
	// Only draw if index is valid and border is present
	if p.index <= 0 || p.index > 10 || p.border == BorderNone || p.rect.Width < 4 {
		return
	}

	// Format index as string
	indexStr := strconv.Itoa(p.index)
	if p.index == 10 {
		indexStr = "0" // Use 0 to represent 10 (for Alt+0)
	}

	// Position at top-right inside border
	indexX := p.rect.X + p.rect.Width - 2
	indexY := p.rect.Y

	// Use border style with possible focus enhancement
	style := p.borderStyle
	if isFocused {
		style = p.focusBorderStyle
	}

	// Draw index
	DrawText(screen, indexX, indexY, style, indexStr)
}

// IsDirty returns whether the pane or its child needs redrawing.
func (p *Pane) IsDirty() bool {
	// Check if pane itself is dirty
	if p.dirty {
		return true
	}

	// Get the child reference without holding lock during child checks
	child := p.child

	// Check if child is dirty - outside of lock
	if child != nil {
		if comp, ok := child.(Component); ok && comp != nil {
			return comp.IsDirty()
		} else if layout, ok := child.(*Layout); ok && layout != nil {
			return layout.HasDirtyComponents()
		}
	}

	return false
}

// ClearDirtyFlags clears dirty flags for this pane and its children recursively.
func (p *Pane) ClearDirtyFlags() {
	// Clear own dirty flag
	p.dirty = false
	child := p.child // Get child reference

	// Clear child's dirty flag - outside of lock
	if child != nil {
		if comp, ok := child.(Component); ok && comp != nil {
			comp.ClearDirty()
		} else if layout, ok := child.(*Layout); ok && layout != nil {
			layout.ClearAllDirtyFlags()
		}
	}
}

// GetChildComponent returns the pane's child component, or nil if it's not a component.
func (p *Pane) GetChildComponent() Component {
	if comp, ok := p.child.(Component); ok {
		return comp
	}

	return nil
}

// GetChildLayout returns the pane's child layout, or nil if it's not a layout.
func (p *Pane) GetChildLayout() *Layout {
	if layout, ok := p.child.(*Layout); ok {
		return layout
	}

	return nil
}

// GetAllComponents returns all components within this pane (recursively).
func (p *Pane) GetAllComponents() []Component {
	child := p.child // Get child reference

	var components []Component

	// Process child outside of lock
	if child != nil {
		if comp, ok := child.(Component); ok && comp != nil {
			components = append(components, comp)
		} else if layout, ok := child.(*Layout); ok && layout != nil {
			// Process layout's components
			for _, comp := range layout.GetAllFocusableComponents() {
				components = append(components, comp)
			}
		}
	}

	return components
}

// GetFocusableComponents returns all focusable components within this pane.
func (p *Pane) GetFocusableComponents() []Component {
	child := p.child // Get child reference

	var focusables []Component

	// Process child outside of lock
	if child != nil {
		if comp, ok := child.(Component); ok && comp != nil && comp.Focusable() {
			focusables = append(focusables, comp)
		} else if layout, ok := child.(*Layout); ok && layout != nil {
			focusables = append(focusables, layout.GetAllFocusableComponents()...)
		}
	}

	return focusables
}

// GetFirstFocusableComponent returns the first focusable component within this pane.
func (p *Pane) GetFirstFocusableComponent() Component {
	focusables := p.GetFocusableComponents()
	if len(focusables) == 0 {
		return nil
	}
	return focusables[0]
}