// layout.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
)

// Orientation defines the direction children are laid out in a FlexLayout.
type Orientation int

const (
	// Horizontal lays out children side-by-side.
	Horizontal Orientation = iota
	// Vertical lays out children one above the other.
	Vertical
)

// ChildInfo holds a widget and its layout constraints within a FlexLayout.
type ChildInfo struct {
	Widget     Widget // The child widget itself
	FixedSize  int    // Fixed size (width for Horizontal, height for Vertical). 0 means use proportion.
	Proportion int    // Proportion of the remaining flexible space to allocate. Minimum 1 if FixedSize is 0.
}

// FlexLayout arranges child widgets horizontally or vertically.
// It distributes space based on fixed sizes and proportions.
type FlexLayout struct {
	BaseWidget               // Embed BaseWidget for common widget functionality
	orientation Orientation  // How to arrange children (Horizontal or Vertical)
	children    []*ChildInfo // List of child widgets and their layout info
	gap         int          // Gap between children
}

// NewFlexLayout creates a new layout container.
func NewFlexLayout(orientation Orientation) *FlexLayout {
	l := &FlexLayout{
		orientation: orientation,
		children:    make([]*ChildInfo, 0),
		gap:         0, // Initialize gap to 0
	}
	l.SetVisible(true)
	// Although FlexLayout itself isn't focusable by default,
	// embedding BaseWidget means it inherits the SetApplication method,
	// which is crucial for propagating the app pointer to children.
	return l
}

// AddChild adds a widget to the layout.
//   - widget: The Widget to add.
//   - fixedSize: The fixed width (if Horizontal) or height (if Vertical).
//     Set to 0 to use proportional sizing.
//   - proportion: The proportion of flexible space to assign (used if fixedSize is 0).
//     Must be at least 1 if fixedSize is 0.
func (l *FlexLayout) AddChild(widget Widget, fixedSize int, proportion int) *FlexLayout {
	if widget == nil {
		return l // Ignore nil widgets
	}

	// Ensure proportion is valid if fixedSize is 0
	prop := proportion
	if fixedSize == 0 && prop < 1 {
		prop = 1
	}

	info := &ChildInfo{
		Widget:     widget,
		FixedSize:  fixedSize,
		Proportion: prop,
	}
	l.children = append(l.children, info)

	// --- Link Child to Parent and Application ---
	widget.SetParent(l) // <-- Set the layout as the parent
	if app := l.App(); app != nil {
		widget.SetApplication(app) // Propagate app pointer if layout already has one
	}
	// --- End Link ---

	// Request redraw if the layout is already part of an application
	if app := l.App(); app != nil {
		app.QueueRedraw()
	}
	return l
}

// SetApplication sets the application pointer for the layout and its children.
// Overrides BaseWidget.SetApplication to propagate to existing children.
// Also ensures parent links are set correctly if children were added before SetApplication.
func (l *FlexLayout) SetApplication(app *Application) {
	l.BaseWidget.SetApplication(app) // Call embedded method first
	// Propagate to any children added *before* the layout was added to the app
	for _, info := range l.children {
		info.Widget.SetApplication(app)
		// Ensure parent is set here too, in case AddChild happened before SetApplication
		// Although AddChild should handle it, this adds robustness.
		if info.Widget.Parent() == nil {
			info.Widget.SetParent(l)
		}
	}
}

// SetRect calculates and sets the bounds for all child widgets based on
// the layout's orientation, children's size constraints, and the gap.
func (l *FlexLayout) SetRect(x, y, width, height int) {
	l.BaseWidget.SetRect(x, y, width, height) // Store our own rect

	// --- Filter Visible Children ---
	visibleChildren := make([]*ChildInfo, 0, len(l.children))
	for _, info := range l.children {
		if info.Widget != nil && info.Widget.IsVisible() { // Check visibility
			visibleChildren = append(visibleChildren, info)
		} else if info.Widget != nil {
			// Ensure invisible widgets have zero size rect
			info.Widget.SetRect(x, y, 0, 0)
		}
	}
	// --- End Filter ---

	numVisibleChildren := len(visibleChildren) // Use count of visible children
	if numVisibleChildren == 0 {
		return // Nothing visible to lay out
	}

	totalFixedSize := 0
	totalProportion := 0
	flexibleChildrenCount := 0

	// First pass: Calculate total fixed size and total proportion for VISIBLE children
	for _, info := range visibleChildren { // Iterate over visible children
		if info.FixedSize > 0 {
			totalFixedSize += info.FixedSize
		} else {
			totalProportion += info.Proportion
			flexibleChildrenCount++
		}
	}

	// Calculate total gap space needed for VISIBLE children
	totalGap := 0
	if numVisibleChildren > 1 { // Use visible count
		totalGap = l.gap * (numVisibleChildren - 1)
	}

	var availableSpace int
	if l.orientation == Horizontal {
		availableSpace = width - totalFixedSize - totalGap
	} else {
		availableSpace = height - totalFixedSize - totalGap
	}

	if availableSpace < 0 {
		availableSpace = 0
	}

	spacePerProportion := 0
	if totalProportion > 0 && availableSpace > 0 {
		spacePerProportion = availableSpace / totalProportion
	}
	remainder := 0
	if totalProportion > 0 && availableSpace > 0 {
		remainder = availableSpace % totalProportion
	}

	currentX, currentY := x, y
	spaceAllocatedToFlex := 0

	// Second pass: Assign rectangles to VISIBLE children
	for i, info := range visibleChildren { // Iterate over visible children
		childWidth := 0
		childHeight := 0
		size := 0

		if info.FixedSize > 0 {
			size = info.FixedSize
		} else if totalProportion > 0 {
			size = info.Proportion * spacePerProportion
			if remainder > 0 {
				size++
				remainder--
			}
			spaceAllocatedToFlex += size
		}

		if info.FixedSize == 0 && spaceAllocatedToFlex > availableSpace {
			diff := spaceAllocatedToFlex - availableSpace
			size -= diff
			spaceAllocatedToFlex -= diff
		}

		if size < 0 {
			size = 0
		}

		if l.orientation == Horizontal {
			childWidth = size
			childHeight = height
			info.Widget.SetRect(currentX, currentY, childWidth, childHeight)
			currentX += childWidth
			if i < numVisibleChildren-1 { // Use visible count for gap check
				currentX += l.gap
			}
		} else { // Vertical
			childWidth = width
			childHeight = size
			info.Widget.SetRect(currentX, currentY, childWidth, childHeight)
			currentY += childHeight
			if i < numVisibleChildren-1 { // Use visible count for gap check
				currentY += l.gap
			}
		}
	}
}

// SetGap sets the gap size between children.
func (l *FlexLayout) SetGap(gap int) *FlexLayout {
	if gap < 0 {
		gap = 0 // Ensure gap is not negative
	}
	if l.gap == gap {
		return l // No change
	}
	l.gap = gap
	if app := l.App(); app != nil {
		app.QueueRedraw() // Changing gap requires redraw
	}
	return l
}

// Draw iterates through the children and calls their Draw methods.
// It relies on SetRect having already positioned the children correctly.
func (l *FlexLayout) Draw(screen tcell.Screen) {
	// --- Visibility Check for Layout Itself ---
	if !l.IsVisible() {
		return
	}
	// --- End Check ---

	// Draw children (BaseWidget.Draw handles visibility check for each child)
	for _, info := range l.children {
		// No need for explicit visibility check here, child.Draw handles it.
		_, _, cw, ch := info.Widget.GetRect()
		if cw > 0 && ch > 0 { // Still draw only if it has dimensions
			info.Widget.Draw(screen)
		}
	}
}

// HandleEvent currently does nothing for the layout itself.
// It relies on the event bubbling mechanism starting from the focused child.
// If the layout itself needed keybindings, they would be set via SetKeybinding.
func (l *FlexLayout) HandleEvent(event tcell.Event) bool {
	// Let BaseWidget handle its own keybindings, if any were set on the layout itself.
	return l.BaseWidget.HandleEvent(event)
}

// Focusable returns false, as the layout container itself is not focusable by default.
func (l *FlexLayout) Focusable() bool {
	return false
}

// Focus does nothing for the layout container.
func (l *FlexLayout) Focus() {
	// No-op
}

// Blur does nothing for the layout container.
func (l *FlexLayout) Blur() {
	// No-op
}

// Children returns the widgets managed by this layout.
// Required by the Widget interface for focus traversal and event bubbling checks.
// --- No change needed here, returns all children regardless of visibility ---
func (l *FlexLayout) Children() []Widget {
	widgets := make([]Widget, len(l.children))
	for i, info := range l.children {
		widgets[i] = info.Widget
	}
	return widgets
}

// Parent returns the layout's parent (set via SetParent).
// Relies on embedded BaseWidget's implementation.

// SetParent sets the layout's parent.
// Relies on embedded BaseWidget's implementation.