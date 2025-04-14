// pane.go
package tinytui

import (
	"log" // Keep for critical errors or temporary debug prints
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// Pane is a container with optional border.
type Pane struct {
	child              interface{} // Can hold Component or *Layout
	border             Border
	title              string
	index              int // User-facing index (1-10), 0 if unset/nested
	rect               Rect
	style              Style
	borderStyle        Style
	focusBorderStyle   Style
	showIndexIndicator bool
	app                *Application
	dirty              bool
}

// NewPane creates a new pane with default settings derived from the current theme.
func NewPane() *Pane {
	theme := GetTheme() // Get theme at creation time
	defaultBorder := BorderSingle
	defaultBorderStyle := DefaultStyle
	defaultFocusBorderStyle := DefaultStyle.Foreground(ColorYellow).Bold(true)
	defaultPaneStyle := DefaultStyle

	if theme != nil {
		defaultBorder = theme.DefaultBorderType()
		defaultBorderStyle = theme.PaneBorderStyle()
		defaultFocusBorderStyle = theme.PaneFocusBorderStyle()
		defaultPaneStyle = theme.PaneStyle()
	}

	return &Pane{
		border:             defaultBorder,
		style:              defaultPaneStyle,
		borderStyle:        defaultBorderStyle,
		focusBorderStyle:   defaultFocusBorderStyle,
		showIndexIndicator: true,
		dirty:              true,
		index:              0, // Initialize index to 0 (meaning no user-facing index)
	}
}

// SetChild sets the pane's child (component or layout).
func (p *Pane) SetChild(child interface{}) {
	switch child.(type) {
	case Component, *Layout, nil:
		// Valid
	default:
		log.Printf("[Pane.SetChild] ERROR: Invalid child type provided to Pane (Title: '%s'): %T", p.title, child)
		return
	}

	// Propagate app reference if pane already has one
	if p.app != nil {
		if comp, ok := child.(Component); ok && comp != nil {
			comp.SetApplication(p.app)
		} else if layout, ok := child.(*Layout); ok && layout != nil {
			layout.SetApplication(p.app)
		}
	}

	p.child = child
	p.dirty = true
}

// SetApplication sets the application this pane belongs to and propagates it.
func (p *Pane) SetApplication(app *Application) {
	if p.app == app {
		return // No change
	}
	p.app = app

	// Propagate to existing child
	if p.child != nil {
		if comp, ok := p.child.(Component); ok && comp != nil {
			comp.SetApplication(app)
		} else if layout, ok := p.child.(*Layout); ok && layout != nil {
			layout.SetApplication(app)
		}
	}
}

// SetBorder sets the pane's border type and style.
func (p *Pane) SetBorder(border Border, style Style) {
	if p.border != border || p.borderStyle != style {
		p.border = border
		p.borderStyle = style
		p.dirty = true
	}
}

// SetFocusBorderStyle sets the style used for the border when a child has focus.
func (p *Pane) SetFocusBorderStyle(style Style) {
	if p.focusBorderStyle != style {
		p.focusBorderStyle = style
		p.dirty = true
	}
}

// SetTitle sets the pane's title.
func (p *Pane) SetTitle(title string) {
	if p.title != title {
		p.title = title
		p.dirty = true
	}
}

// SetStyle sets the pane's background style.
func (p *Pane) SetStyle(style Style) {
	if p.style != style {
		p.style = style
		p.dirty = true
	}
}

// SetRect sets the pane's position and size.
func (p *Pane) SetRect(x, y, width, height int) {
	newRect := Rect{X: x, Y: y, Width: width, Height: height}
	if p.rect == newRect {
		return // No change
	}
	p.rect = newRect
	p.dirty = true

	// Update child dimensions immediately
	p.updateChildRect()
}

// updateChildRect calculates content area and sets child's rect.
func (p *Pane) updateChildRect() {
	borderType := p.border
	// Determine effective border for content calculation
	if borderType != BorderNone && (p.rect.Width < 3 || p.rect.Height < 3) {
		borderType = BorderNone
	}
	contentX, contentY, contentWidth, contentHeight := p.getContentRectForBorder(borderType)

	if p.child != nil {
		if comp, ok := p.child.(Component); ok && comp != nil {
			comp.SetRect(contentX, contentY, contentWidth, contentHeight)
		} else if layout, ok := p.child.(*Layout); ok && layout != nil {
			layout.SetRect(contentX, contentY, contentWidth, contentHeight)
		}
	}
}

// getContentRectForBorder calculates the inner content rectangle based on a given border type.
func (p *Pane) getContentRectForBorder(border Border) (x, y, width, height int) {
	rect := p.rect
	x, y = rect.X, rect.Y
	width, height = rect.Width, rect.Height

	if border != BorderNone {
		x += 1
		y += 1
		width -= 2
		height -= 2
		if width < 0 {
			width = 0
		}
		if height < 0 {
			height = 0
		}
	}
	return x, y, width, height
}

// SetIndex sets the pane's user-facing index (1-10).
// Called by Layout.SetApplication for the main layout.
// Use 0 internally for non-indexed panes.
func (p *Pane) SetIndex(index int) {
	newIndex := 0
	if index > 0 && index <= 10 {
		newIndex = index
	}
	if p.index != newIndex {
		p.index = newIndex
		p.dirty = true // Mark dirty as appearance might change
	}
}

// GetIndex returns the pane's user-facing index (1-10), or 0 if none.
func (p *Pane) GetIndex() int {
	return p.index
}

// SetShowIndexIndicator sets whether the pane's index should be shown.
func (p *Pane) SetShowIndexIndicator(show bool) {
	if p.showIndexIndicator != show {
		p.showIndexIndicator = show
		p.dirty = true
	}
}

// Draw draws the pane and its child.
func (p *Pane) Draw(screen tcell.Screen, isFocused bool) {
	if p.rect.Width <= 0 || p.rect.Height <= 0 {
		return
	}

	// Copy state needed for drawing
	rect := p.rect
	border := p.border // Base border type
	title := p.title
	index := p.index // Current index (0 if none)
	showIndexIndicator := p.showIndexIndicator
	style := p.style
	borderStyle := p.borderStyle
	focusBorderStyle := p.focusBorderStyle
	child := p.child

	p.dirty = false // Clear dirty flag for this pane itself

	// Determine effective border type (might be None if too small, might change on focus)
	effectiveBorder := border
	if effectiveBorder != BorderNone && (rect.Width < 3 || rect.Height < 3) {
		effectiveBorder = BorderNone
	}

	// Determine border style and potentially update effectiveBorder based on focus/theme
	currentBorderStyle := borderStyle
	theme := GetTheme() // Get current theme
	if isFocused {
		if theme != nil {
			focusedThemeBorder := theme.FocusedBorderType()
			// Only change border type if theme specifies a *different* one for focus
			if focusedThemeBorder != border && focusedThemeBorder != BorderNone {
				effectiveBorder = focusedThemeBorder
			}
			currentBorderStyle = theme.PaneFocusBorderStyle() // Use theme's focus style
		} else {
			currentBorderStyle = focusBorderStyle // Fallback focus style
		}
	} else {
		if theme != nil {
			// Ensure we use the theme's default border type if not focused
			effectiveBorder = theme.DefaultBorderType()
			currentBorderStyle = theme.PaneBorderStyle()
		}
		// If no theme, border/borderStyle remain as set on the pane
	}
	// Re-check size constraint with potentially changed effectiveBorder
	if effectiveBorder != BorderNone && (rect.Width < 3 || rect.Height < 3) {
		effectiveBorder = BorderNone
	}

	// Draw background
	Fill(screen, rect.X, rect.Y, rect.Width, rect.Height, ' ', style)

	// --- Start Border and Index/Title Drawing ---
	indexDisplayString := "" // Calculated below

	if effectiveBorder != BorderNone {
		// Draw the actual border lines/corners using the determined effective type and style
		drawBorderByType(screen, rect.X, rect.Y, rect.Width, rect.Height, currentBorderStyle, effectiveBorder)

		// --- Calculate Index Display String ---
		startX := rect.X + 1 // Position right after the left border column

		// Conditions for showing the actual index number
		isPotentiallyNavigable := index > 0                        // Has a valid index assigned (1-10)
		containsFocusable := p.GetFirstFocusableComponent() != nil // Check if navigable
		shouldDisplayIndexNumber := isPotentiallyNavigable && showIndexIndicator && containsFocusable

		if shouldDisplayIndexNumber {
			indexNumStr := strconv.Itoa(index)
			if index == 10 {
				indexNumStr = "0"
			} // Alt+0 convention
			displayChar := indexNumStr
			if isFocused {
				displayChar = "#"
			}
			prefix, suffix := "[", "]"
			indexDisplayString = prefix + displayChar + suffix
		} else {
			// Show placeholder if a border exists but index shouldn't be shown
			var horizontalChar rune
			switch effectiveBorder {
			case BorderSingle:
				horizontalChar = RuneHLine
			case BorderDouble:
				horizontalChar = RuneDoubleHLine
			case BorderSolid:
				horizontalChar = RuneBlock
			default:
				horizontalChar = '-'
			}
			indexDisplayString = strings.Repeat(string(horizontalChar), 3)
		}

		// --- Draw Index String (or placeholder) ---
		indexDisplayLen := runewidth.StringWidth(indexDisplayString)
		canDrawIndex := startX+indexDisplayLen <= rect.X+rect.Width-1 && rect.Width > 4

		if canDrawIndex {
			DrawText(screen, startX, rect.Y, currentBorderStyle, indexDisplayString)
		}

		// --- Draw Title ---
		if title != "" {
			titleX := startX + indexDisplayLen + 1              // Start after index string + 1 space
			titleMaxWidth := rect.Width - (titleX - rect.X) - 1 // Max width = Total width - start offset - right border(1)

			if titleMaxWidth > 0 {
				titleText := runewidth.Truncate(title, titleMaxWidth, "…") // Use ellipsis
				DrawText(screen, titleX, rect.Y, currentBorderStyle, titleText)
			}
		}
	}
	// --- End Border and Index/Title Drawing ---

	// Calculate content area based on the final effectiveBorder used for drawing
	contentX, contentY, contentWidth, contentHeight := p.getContentRectForBorder(effectiveBorder)

	// Draw child if present
	if child != nil {
		// Ensure child rect is updated *before* drawing, in case it wasn't set by SetRect yet
		// (e.g., if only style changed, SetRect might not have run)
		if comp, ok := child.(Component); ok && comp != nil {
			comp.SetRect(contentX, contentY, contentWidth, contentHeight) // Ensure rect is current
			comp.Draw(screen)
		} else if layout, ok := child.(*Layout); ok && layout != nil {
			layout.SetRect(contentX, contentY, contentWidth, contentHeight) // Ensure rect is current
			layout.Draw(screen)
		}
	}
}

// IsDirty returns whether the pane or its child needs redrawing.
func (p *Pane) IsDirty() bool {
	if p.dirty {
		return true
	}
	// Check child recursively
	child := p.child
	if child != nil {
		if comp, ok := child.(Component); ok && comp != nil {
			return comp.IsDirty()
		} else if layout, ok := child.(*Layout); ok && layout != nil {
			return layout.HasDirtyComponents() // Recursive check
		}
	}
	return false
}

// ClearDirtyFlags clears dirty flags for this pane and its children recursively.
func (p *Pane) ClearDirtyFlags() {
	p.dirty = false
	child := p.child
	if child != nil {
		if comp, ok := child.(Component); ok && comp != nil {
			comp.ClearDirty()
		} else if layout, ok := child.(*Layout); ok && layout != nil {
			layout.ClearAllDirtyFlags() // Recursive clear
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
// Note: Needs a corresponding GetAllComponents in Layout for full recursion.
func (p *Pane) GetAllComponents() []Component {
	child := p.child
	var components []Component
	if child != nil {
		if comp, ok := child.(Component); ok && comp != nil {
			components = append(components, comp)
		} else if layout, ok := child.(*Layout); ok && layout != nil {
			// Assuming Layout needs a GetAllComponents method similar to GetAllFocusableComponents
			// components = append(components, layout.GetAllComponents()...)
			// Using focusable as fallback for now:
			components = append(components, layout.GetAllFocusableComponents()...)
		}
	}
	return components
}

// GetFocusableComponents returns all focusable components within this pane recursively.
func (p *Pane) GetFocusableComponents() []Component {
	child := p.child
	var focusables []Component
	if child != nil {
		if comp, ok := child.(Component); ok && comp != nil {
			if comp.Focusable() {
				focusables = append(focusables, comp)
			}
		} else if layout, ok := child.(*Layout); ok && layout != nil {
			// Delegate recursive search to layout
			focusables = append(focusables, layout.GetAllFocusableComponents()...)
		}
	}
	return focusables
}

// GetFirstFocusableComponent returns the first focusable component found recursively.
func (p *Pane) GetFirstFocusableComponent() Component {
	focusables := p.GetFocusableComponents() // Performs the recursive search
	if len(focusables) == 0 {
		return nil
	}
	return focusables[0]
}

// Helper function moved from previous response for clarity
func drawBorderByType(screen tcell.Screen, x, y, width, height int, style Style, borderType Border) {
	if width < 2 || height < 2 { // Need at least 2x2 to draw corners
		return
	}
	switch borderType {
	case BorderSingle:
		DrawBox(screen, x, y, width, height, style)
	case BorderDouble:
		DrawDoubleBox(screen, x, y, width, height, style)
	case BorderSolid:
		DrawSolidBox(screen, x, y, width, height, style)
	case BorderNone:
		// Do nothing
	}
}