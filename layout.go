// layout.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
	"sync"
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
	Widget     Widget    // The child widget itself
	FixedSize  int       // Fixed size (width for Horizontal, height for Vertical). 0 means use proportion.
	Proportion int       // Proportion of the remaining flexible space to allocate. Minimum 1 if FixedSize is 0.
	Alignment  Alignment // How this child aligns on the cross axis (overrides parent's crossAxisAlign)
}

// Alignment defines how items are aligned along a layout axis.
type Alignment int

const (
	// AlignStart aligns items at the start (top/left)
	AlignStart Alignment = iota
	// AlignCenter centers items within the available space
	AlignCenter
	// AlignEnd aligns items at the end (bottom/right)
	AlignEnd
	// AlignStretch stretches items to fill the container (default)
	AlignStretch
)

// FlexLayout arranges child widgets horizontally or vertically.
// It distributes space based on fixed sizes and proportions.
type FlexLayout struct {
	BaseWidget               // Embed BaseWidget for common widget functionality
	orientation Orientation  // How to arrange children (Horizontal or Vertical)
	children    []*ChildInfo // List of child widgets and their layout info
	gap         int          // Gap between children

	// Alignment properties
	mainAxisAlign  Alignment // Alignment along the main axis (horizontal for Horizontal, vertical for Vertical)
	crossAxisAlign Alignment // Alignment along the cross axis (vertical for Horizontal, horizontal for Vertical)

	mu sync.RWMutex // Protect concurrent access to properties
}

// NewFlexLayout creates a new layout container with default alignments.
func NewFlexLayout(orientation Orientation) *FlexLayout {
	l := &FlexLayout{
		orientation:    orientation,
		children:       make([]*ChildInfo, 0),
		gap:            0,
		mainAxisAlign:  AlignStart,   // Default: items start at beginning
		crossAxisAlign: AlignStretch, // Default: stretch items in cross-axis
	}
	l.SetVisible(true) // Explicitly set visibility
	return l
}

// ApplyTheme applies the provided theme to the FlexLayout widget
// FlexLayout doesn't have visual styles of its own, so it just passes the theme to children
func (l *FlexLayout) ApplyTheme(theme Theme) {
	// FlexLayout doesn't have its own style to update
	// Children will be handled by the recursive application logic
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

// Modified SetRect to handle alignment
func (l *FlexLayout) SetRect(x, y, width, height int) {
	l.BaseWidget.SetRect(x, y, width, height) // Store our own rect

	l.mu.RLock()
	defer l.mu.RUnlock()

	// Filter visible children
	visibleChildren := make([]*ChildInfo, 0, len(l.children))
	for _, info := range l.children {
		if info.Widget != nil && info.Widget.IsVisible() {
			visibleChildren = append(visibleChildren, info)
		} else if info.Widget != nil {
			info.Widget.SetRect(x, y, 0, 0)
		}
	}

	numVisibleChildren := len(visibleChildren)
	if numVisibleChildren == 0 {
		return
	}

	// Calculate sizes for each child first
	sizes := make([]int, numVisibleChildren)

	totalFixedSize := 0
	totalProportion := 0
	flexibleChildrenCount := 0

	// Calculate total fixed size and proportion
	for i, info := range visibleChildren {
		if info.FixedSize > 0 {
			sizes[i] = info.FixedSize
			totalFixedSize += info.FixedSize
		} else {
			totalProportion += info.Proportion
			flexibleChildrenCount++
		}
	}

	// Calculate total gap space
	totalGap := 0
	if numVisibleChildren > 1 {
		totalGap = l.gap * (numVisibleChildren - 1)
	}

	// Calculate available space for proportional items
	var availableSpace int
	if l.orientation == Horizontal {
		availableSpace = width - totalFixedSize - totalGap
	} else {
		availableSpace = height - totalFixedSize - totalGap
	}

	if availableSpace < 0 {
		availableSpace = 0
	}

	// Calculate sizes for proportional children
	if totalProportion > 0 && availableSpace > 0 {
		spacePerProportion := availableSpace / totalProportion
		remainder := availableSpace % totalProportion

		spaceAllocatedToFlex := 0

		for i, info := range visibleChildren {
			if info.FixedSize <= 0 {
				size := info.Proportion * spacePerProportion
				if remainder > 0 {
					size++
					remainder--
				}

				spaceAllocatedToFlex += size
				if spaceAllocatedToFlex > availableSpace {
					diff := spaceAllocatedToFlex - availableSpace
					size -= diff
					spaceAllocatedToFlex -= diff
				}

				sizes[i] = max(0, size)
			}
		}
	}

	// Calculate offsets for main axis alignment
	mainAxisOffset := 0
	if l.mainAxisAlign != AlignStart {
		totalContentSize := totalFixedSize + availableSpace + totalGap

		mainAxisSpace := 0
		if l.orientation == Horizontal {
			mainAxisSpace = width - totalContentSize
		} else {
			mainAxisSpace = height - totalContentSize
		}

		if mainAxisSpace > 0 {
			if l.mainAxisAlign == AlignCenter {
				mainAxisOffset = mainAxisSpace / 2
			} else if l.mainAxisAlign == AlignEnd {
				mainAxisOffset = mainAxisSpace
			}
		}
	}

	// Calculate starting position
	currentPos := mainAxisOffset
	if l.orientation == Horizontal {
		currentPos += x
	} else {
		currentPos += y
	}

	// Position each child based on alignments, ensuring they don't exceed container bounds
	for i, info := range visibleChildren {
		childWidth := width
		childHeight := height
		childX := x
		childY := y

		if l.orientation == Horizontal {
			// In horizontal layout, width is the child's size
			childWidth = sizes[i]

			// Set horizontal position
			childX = currentPos

			// Calculate vertical position based on cross-axis alignment
			childAlign := l.crossAxisAlign
			if info.Alignment != AlignStretch { // Use child's own alignment if specified
				childAlign = info.Alignment
			}

			if childAlign == AlignStart {
				childY = y
			} else if childAlign == AlignCenter {
				// Only center if not stretching
				childHeight = min(height, getPreferredHeight(info.Widget))
				childY = y + (height-childHeight)/2
			} else if childAlign == AlignEnd {
				// Only align to end if not stretching
				childHeight = min(height, getPreferredHeight(info.Widget))
				childY = y + height - childHeight
			}
			// AlignStretch uses full height

			// Move to next position
			currentPos += childWidth + l.gap
		} else {
			// In vertical layout, height is the child's size
			childHeight = sizes[i]

			// Set vertical position
			childY = currentPos

			// Calculate horizontal position based on cross-axis alignment
			childAlign := l.crossAxisAlign
			if info.Alignment != AlignStretch { // Use child's own alignment if specified
				childAlign = info.Alignment
			}

			if childAlign == AlignStart {
				childX = x
			} else if childAlign == AlignCenter {
				// Only center if not stretching
				childWidth = min(width, getPreferredWidth(info.Widget))
				childX = x + (width-childWidth)/2
			} else if childAlign == AlignEnd {
				// Only align to end if not stretching
				childWidth = min(width, getPreferredWidth(info.Widget))
				childX = x + width - childWidth
			}
			// AlignStretch uses full width

			// Move to next position
			currentPos += childHeight + l.gap
		}

		// Ensure we don't allocate more space than the container has
		if childX+childWidth > x+width {
			childWidth = (x + width) - childX
		}
		if childY+childHeight > y+height {
			childHeight = (y + height) - childY
		}

		// Apply the calculated position and size
		if childWidth > 0 && childHeight > 0 {
			info.Widget.SetRect(childX, childY, childWidth, childHeight)
		} else {
			// If dimensions are invalid, set to zero size to prevent drawing
			info.Widget.SetRect(childX, childY, 0, 0)
		}
	}
}

// Helper function for preferred sizing
func getPreferredWidth(w Widget) int {
	if widget, ok := w.(interface{ PreferredWidth() int }); ok {
		return widget.PreferredWidth()
	}
	return 10 // Default fallback
}

func getPreferredHeight(w Widget) int {
	if widget, ok := w.(interface{ PreferredHeight() int }); ok {
		return widget.PreferredHeight()
	}
	return 1 // Default fallback
}

// SetMainAxisAlignment sets how items are aligned along the main axis.
func (l *FlexLayout) SetMainAxisAlignment(align Alignment) *FlexLayout {
	l.mu.Lock()
	l.mainAxisAlign = align
	l.mu.Unlock()

	if app := l.App(); app != nil {
		app.QueueRedraw()
	}
	return l
}

// SetCrossAxisAlignment sets the default alignment of items along the cross axis.
func (l *FlexLayout) SetCrossAxisAlignment(align Alignment) *FlexLayout {
	l.mu.Lock()
	l.crossAxisAlign = align
	l.mu.Unlock()

	if app := l.App(); app != nil {
		app.QueueRedraw()
	}
	return l
}

// AddChildWithAlign adds a widget to the layout with specific alignment.
func (l *FlexLayout) AddChildWithAlign(widget Widget, fixedSize int, proportion int, align Alignment) *FlexLayout {
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
		Alignment:  align,
	}

	l.mu.Lock()
	l.children = append(l.children, info)
	l.mu.Unlock()

	// Link Child to Parent and Application
	widget.SetParent(l)
	if app := l.App(); app != nil {
		widget.SetApplication(app)
	}

	// Request redraw if the layout is already part of an application
	if app := l.App(); app != nil {
		app.QueueRedraw()
	}
	return l
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
	l.BaseWidget.Draw(screen)

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