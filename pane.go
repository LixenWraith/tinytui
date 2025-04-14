// pane.go
package tinytui

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// Pane acts as a container for a single child (which can be a Component or another Layout).
// It manages the child's position relative to the pane's border and can draw the border,
// title, and user-facing index indicator.
type Pane struct {
	child            interface{}  // Holds Component or *Layout
	border           Border       // Current border type setting (might be overridden by theme focus rule)
	title            string       // Text displayed in the top border
	slotIndex        int          // Internal index (0-9) indicating the slot this pane occupies in its parent Layout. 0 if not set.
	navIndex         int          // User-facing navigation index (1-10), assigned dynamically. 0 if not navigable.
	rect             Rect         // Position and size allocated to the pane (including border area)
	style            Style        // Background style for the pane's content area
	borderStyle      Style        // Style for the border when unfocused (can be overridden by theme)
	focusBorderStyle Style        // Style for the border when focused (can be overridden by theme)
	app              *Application // Reference to the parent application
	dirty            bool         // Does the pane (border, title) or its child need redrawing?
}

// NewPane creates a new pane, initializing styles and border from the current theme.
func NewPane() *Pane {
	theme := GetTheme()
	if theme == nil {
		theme = NewDefaultTheme()
	} // Fallback

	p := &Pane{
		// Initialize visual properties from the theme
		border:           theme.DefaultBorderType(),    // Use theme default border initially
		style:            theme.PaneStyle(),            // Use theme pane background
		borderStyle:      theme.PaneBorderStyle(),      // Use theme border style
		focusBorderStyle: theme.PaneFocusBorderStyle(), // Use theme focus border style
		dirty:            true,                         // Start dirty for initial draw
		slotIndex:        0,                            // Slot index is assigned by Layout.AddPane
		navIndex:         0,                            // Navigation index is assigned dynamically
		// child and app are nil initially
	}
	return p
}

// ApplyThemeRecursively applies the theme to the pane itself and its child.
// This updates the pane's styles based on the theme and propagates the theme down.
func (p *Pane) ApplyThemeRecursively(theme Theme) {
	if theme == nil {
		return
	}

	// Apply theme styles to the pane itself
	p.border = theme.DefaultBorderType() // Update border type from theme's default
	p.style = theme.PaneStyle()
	p.borderStyle = theme.PaneBorderStyle()
	p.focusBorderStyle = theme.PaneFocusBorderStyle()
	p.dirty = true // Mark dirty as appearance might change

	// Apply theme to the child recursively
	if p.child != nil {
		if themedChild, ok := p.child.(ThemedComponent); ok {
			themedChild.ApplyTheme(theme) // Child implements custom theme handling
		} else if layoutChild, ok := p.child.(*Layout); ok {
			layoutChild.ApplyThemeRecursively(theme) // Layout handles its children
		} else if compChild, ok := p.child.(Component); ok {
			// If child is just a Component, maybe try setting its application again
			// so it can potentially re-fetch theme styles if needed? Or rely on redraw.
			if p.app != nil {
				compChild.SetApplication(p.app) // Ensure app reference is up-to-date
			}
			compChild.MarkDirty() // Mark child dirty as its container theme changed
		}
	}

	p.updateChildRect() // Re-calculate child rect in case border type changed size
}

// SetChild sets the pane's content (a Component or another Layout).
// Validates the child type and propagates application/theme settings.
func (p *Pane) SetChild(child interface{}) {
	// Validate child type
	isValid := false
	switch child.(type) {
	case Component, *Layout, nil:
		isValid = true
	}
	if !isValid {
		// Consider panic or returning an error for invalid child types
		panic("Invalid child type for Pane: must be Component, *Layout, or nil")
	}

	// Check if child is actually changing
	if p.child == child {
		return
	}
	p.child = child

	// Propagate application reference and theme to the new child
	if p.app != nil {
		app := p.app                   // Cache app reference
		currentTheme := app.GetTheme() // Get current theme

		if comp, ok := child.(Component); ok && comp != nil {
			comp.SetApplication(app)
			// Apply theme if the component supports it
			if themedComp, ok := comp.(ThemedComponent); ok {
				themedComp.ApplyTheme(currentTheme)
			} else {
				comp.MarkDirty() // Mark dirty anyway
			}
		} else if layout, ok := child.(*Layout); ok && layout != nil {
			layout.SetApplication(app)
			// ApplyThemeRecursively is handled within layout.SetApplication now
			// layout.ApplyThemeRecursively(currentTheme) // Layouts handle recursive application
		}
	}

	p.dirty = true
	p.updateChildRect() // Update child rect based on current border

	// Trigger nav index recalculation if app context exists and layout is set
	// because the focusability of the pane might have changed.
	if p.app != nil && p.app.GetLayout() != nil {
		p.app.Dispatch(&RecalculateNavIndicesCommand{})
	}
}

// SetApplication associates the pane with an application instance and propagates it to the child.
func (p *Pane) SetApplication(app *Application) {
	if p.app == app {
		return
	} // No change
	p.app = app

	// Propagate to existing child
	if p.child != nil {
		if comp, ok := p.child.(Component); ok && comp != nil {
			comp.SetApplication(app)
		} else if layout, ok := p.child.(*Layout); ok && layout != nil {
			layout.SetApplication(app) // Layout handles its own children
		}
	}
}

// SetBorder allows explicitly setting the pane's default (unfocused) border type and style.
// Note: This overrides the theme's DefaultBorderType and PaneBorderStyle for this pane.
// The theme's *focused* border type/style might still apply when focused.
func (p *Pane) SetBorder(border Border, style Style) {
	if p.border != border || p.borderStyle != style {
		p.border = border
		p.borderStyle = style
		p.dirty = true
		p.updateChildRect() // Border change affects content area size
	}
}

// SetFocusBorderStyle allows explicitly setting the focused border style.
// Note: This overrides the theme's PaneFocusBorderStyle for this pane.
func (p *Pane) SetFocusBorderStyle(style Style) {
	if p.focusBorderStyle != style {
		p.focusBorderStyle = style
		// Only mark dirty if focused? Or always? Always for simplicity.
		p.dirty = true
	}
}

// SetTitle sets the text displayed in the top border of the pane.
func (p *Pane) SetTitle(title string) {
	if p.title != title {
		p.title = title
		p.dirty = true // Border appearance changes
	}
}

// SetStyle sets the background style for the pane's content area (inside the border).
// Note: This overrides the theme's PaneStyle for this specific pane.
func (p *Pane) SetStyle(style Style) {
	if p.style != style {
		p.style = style
		p.dirty = true // Background appearance changes
	}
}

// SetRect sets the pane's outer position and size (including any border area).
// It recalculates and sets the inner rectangle for the child component/layout.
func (p *Pane) SetRect(x, y, width, height int) {
	newRect := Rect{X: x, Y: y, Width: width, Height: height}
	if p.rect == newRect {
		return
	} // No change
	p.rect = newRect
	p.dirty = true
	p.updateChildRect() // Update child dimensions based on new pane size and current border
}

// updateChildRect calculates content area and sets child's rect.
// Calls the corrected getContentRectForBorder helper.
func (p *Pane) updateChildRect() {
	// Determine the border type to use for geometry calculation (usually the default)
	borderForGeometry := p.border
	if borderForGeometry != BorderNone && (p.rect.Width < 2 || p.rect.Height < 2) {
		borderForGeometry = BorderNone
	}

	// Calculate content area using the helper
	contentX, contentY, contentWidth, contentHeight := p.getContentRectForBorder(borderForGeometry)

	// Set the calculated rectangle for the child
	if p.child != nil {
		if comp, ok := p.child.(Component); ok && comp != nil {
			comp.SetRect(contentX, contentY, contentWidth, contentHeight)
		} else if layout, ok := p.child.(*Layout); ok && layout != nil {
			layout.SetRect(contentX, contentY, contentWidth, contentHeight)
		}
	}
}

// getContentRectForBorder calculates the inner content rectangle based on a given
// border type and the pane's outer rectangle.
func (p *Pane) getContentRectForBorder(border Border) (x, y, width, height int) {
	// Use the pane's current rectangle (p.rect)
	rect := p.rect
	x, y = rect.X, rect.Y
	width, height = rect.Width, rect.Height

	// Adjust ONLY if border is present AND there's enough space for it
	if border != BorderNone && width >= 2 && height >= 2 {
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

// Draw renders the pane. Signature changed back.
func (p *Pane) Draw(screen tcell.Screen, hasFocus bool) {
	rect := p.rect
	if rect.Width <= 0 || rect.Height <= 0 {
		return
	}
	p.dirty = false

	// --- Determine Effective Border and Style --- (Logic mostly unchanged)
	effectiveBorder := p.border
	currentBorderStyle := p.borderStyle
	theme := GetTheme()
	if theme == nil {
		theme = NewDefaultTheme()
	}
	if hasFocus {
		if p.focusBorderStyle != NewPane().focusBorderStyle {
			currentBorderStyle = p.focusBorderStyle
		} else {
			currentBorderStyle = theme.PaneFocusBorderStyle()
		}
		focusedThemeBorder := theme.FocusedBorderType()
		if focusedThemeBorder != BorderNone && focusedThemeBorder != p.border {
			effectiveBorder = focusedThemeBorder
		} else {
			effectiveBorder = p.border
		}
	} else {
		if p.borderStyle == NewPane().borderStyle {
			currentBorderStyle = theme.PaneBorderStyle()
		}
		if p.border == NewPane().border {
			effectiveBorder = theme.DefaultBorderType()
		}
	}
	if effectiveBorder != BorderNone && (rect.Width < 2 || rect.Height < 2) {
		effectiveBorder = BorderNone
	}

	// --- Draw Background ---
	Fill(screen, rect.X, rect.Y, rect.Width, rect.Height, ' ', p.style)

	// --- Draw Border, Title, Index ---
	if effectiveBorder != BorderNone {
		drawBorderByType(screen, rect.X, rect.Y, rect.Width, rect.Height, currentBorderStyle, effectiveBorder)
		titleAreaX := rect.X + 1
		titleAreaY := rect.Y
		titleAreaWidth := rect.Width - 2
		if titleAreaWidth < 0 {
			titleAreaWidth = 0
		}

		// --- REVISED INDEX LOGIC ---
		indexDisplayString := ""
		// Display indicator ONLY if navIndex is set (>0) and app setting is enabled
		shouldDisplayIndexIndicator := p.app != nil && p.app.IsShowPaneIndicesEnabled() && p.navIndex > 0

		indexDisplayLen := 0
		if shouldDisplayIndexIndicator {
			if hasFocus {
				indexDisplayString = "[#]" // Focused indicator
			} else {
				indexNumStr := strconv.Itoa(p.navIndex % 10) // Unfocused indicator (0 for 10)
				indexDisplayString = "[" + indexNumStr + "]"
			}
			// Draw if it fits
			if titleAreaWidth >= runewidth.StringWidth(indexDisplayString) {
				DrawText(screen, titleAreaX, titleAreaY, currentBorderStyle, indexDisplayString)
				indexDisplayLen = runewidth.StringWidth(indexDisplayString)
			}
		} // If navIndex is 0 or setting disabled, indicator is never drawn.
		// --- Removed single-pane logic and [ ] placeholder logic ---

		// --- Title Drawing (Adjusted) ---
		if p.title != "" && titleAreaWidth > 0 {
			titleStartX := titleAreaX
			availableTitleWidth := titleAreaWidth
			padding := 1
			if indexDisplayLen > 0 { // If index *was* drawn
				titleStartX += indexDisplayLen + padding
				availableTitleWidth -= (indexDisplayLen + padding)
			} else { // If index was *not* drawn
				// Add padding from the left edge only if title exists
				titleStartX += padding
				availableTitleWidth -= padding
			}
			if availableTitleWidth > 0 {
				truncatedTitle := runewidth.Truncate(p.title, availableTitleWidth, "…")
				DrawText(screen, titleStartX, titleAreaY, currentBorderStyle, truncatedTitle)
			}
		}
	} // --- End Border and Index/Title Drawing ---

	// --- Draw Child --- (Logic unchanged)
	_, _, contentWidth, contentHeight := p.getContentRectForBorder(effectiveBorder)
	if p.child != nil && contentWidth > 0 && contentHeight > 0 {
		if comp, ok := p.child.(Component); ok && comp != nil {
			comp.Draw(screen)
		} else if layout, ok := p.child.(*Layout); ok && layout != nil {
			layout.Draw(screen) // Layout draw doesn't need focus info passed down directly here
		}
	}
}

// ContainsFocus checks recursively if this pane or its child contains the specified focused component.
func (p *Pane) ContainsFocus(focused Component) bool {
	if focused == nil {
		return false
	}
	if p.child == nil {
		return false
	}

	// Check if the direct child is the focused component
	if childComp, ok := p.child.(Component); ok && childComp == focused {
		return true
	}
	// Check if the child is a layout and recursively check if it contains the focus
	if childLayout, ok := p.child.(*Layout); ok {
		return childLayout.ContainsFocus(focused)
	}
	// Otherwise, focus is not within this pane's hierarchy
	return false
}

// HasFocusableChild checks if the pane's child (recursively) contains any focusable component.
// Used by Draw to determine if the index indicator should potentially be shown.
func (p *Pane) HasFocusableChild() bool {
	// Use GetFirstFocusableComponent which performs the recursive check efficiently.
	return p.GetFirstFocusableComponent() != nil
}

// IsDirty returns true if the pane itself (border, title, style) or its child (recursively) needs redrawing.
func (p *Pane) IsDirty() bool {
	if p.dirty {
		return true
	} // Pane properties changed

	// Check if child is dirty (recursively)
	if p.child != nil {
		if comp, ok := p.child.(Component); ok && comp != nil {
			return comp.IsDirty()
		}
		if layout, ok := p.child.(*Layout); ok && layout != nil {
			return layout.HasDirtyComponents() // Layout checks its children
		}
	}
	return false // Not dirty and no dirty children
}

// ClearDirtyFlags clears the dirty flag for this pane and its child (recursively).
// Called by the layout/application after drawing.
func (p *Pane) ClearDirtyFlags() {
	p.dirty = false
	// Clear child's dirty flag recursively
	if p.child != nil {
		if comp, ok := p.child.(Component); ok && comp != nil {
			comp.ClearDirty()
		}
		if layout, ok := p.child.(*Layout); ok && layout != nil {
			layout.ClearAllDirtyFlags() // Layout handles its children
		}
	}
}

// GetChildComponent returns the pane's child if it's a Component, otherwise nil.
func (p *Pane) GetChildComponent() Component {
	if comp, ok := p.child.(Component); ok {
		return comp
	}
	return nil
}

// GetChildLayout returns the pane's child if it's a Layout, otherwise nil.
func (p *Pane) GetChildLayout() *Layout {
	if layout, ok := p.child.(*Layout); ok {
		return layout
	}
	return nil
}

// GetFocusableComponents returns a slice of all focusable components within this pane's child hierarchy.
// The order depends on the child type (single component or layout's traversal order).
func (p *Pane) GetFocusableComponents() []Component {
	var focusables []Component
	if p.child != nil {
		if comp, ok := p.child.(Component); ok && comp != nil {
			// If the child is a component, check if it's focusable itself
			if comp.Focusable() {
				focusables = append(focusables, comp)
			}
		} else if layout, ok := p.child.(*Layout); ok && layout != nil {
			// If the child is a layout, delegate to get all its focusable components
			focusables = append(focusables, layout.GetAllFocusableComponents()...)
		}
	}
	return focusables
}

// GetFirstFocusableComponent finds and returns the first focusable component
// encountered within this pane's child hierarchy (depth-first). Returns nil if none found.
func (p *Pane) GetFirstFocusableComponent() Component {
	// Use the slice returned by GetFocusableComponents for simplicity
	focusables := p.GetFocusableComponents()
	if len(focusables) == 0 {
		return nil
	}
	return focusables[0] // Return the first one found
}

func drawBorderByType(screen tcell.Screen, x, y, width, height int, style Style, borderType Border) {
	// Let the specific Draw functions handle edge cases like 1x1
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

// setSlotIndex sets the pane's internal slot index (0-9). Called by Layout.
func (p *Pane) setSlotIndex(index int) {
	// No clamping needed here, Layout manages valid indices 0-9
	if p.slotIndex != index {
		p.slotIndex = index
		// No need to mark dirty just for slot index change unless drawing depends on it.
		// p.dirty = true
	}
}

// GetSlotIndex returns the pane's internal slot index (0-9), or 0 if not set.
func (p *Pane) GetSlotIndex() int {
	return p.slotIndex
}

// SetNavIndex sets the pane's user-facing navigation index (1-10), or 0 if not navigable.
// Called dynamically by Layout.assignNavigationIndices.
func (p *Pane) SetNavIndex(ni int) {
	newNavIndex := 0
	if ni > 0 && ni <= 10 {
		newNavIndex = ni
	}
	if p.navIndex != newNavIndex {
		p.navIndex = newNavIndex
		p.dirty = true // Mark dirty as border appearance (index indicator) might change
	}
}

// GetNavIndex returns the pane's user-facing navigation index (1-10), or 0 if none.
func (p *Pane) GetNavIndex() int {
	return p.navIndex
}