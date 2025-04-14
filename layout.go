// layout.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
)

// Layout organizes panes on screen.
type Layout struct {
	panes          [10]PaneInfo // Fixed size array [0..9]
	orientation    Orientation
	gap            int
	activeCount    int
	mainAxisAlign  Alignment
	crossAxisAlign Alignment
	rect           Rect
	app            *Application
}

// PaneInfo holds a pane and its layout constraints.
type PaneInfo struct {
	Pane   *Pane
	Size   Size
	Active bool // Is this slot used?
}

// NewLayout creates a new layout with the specified orientation.
func NewLayout(orientation Orientation) *Layout {
	return &Layout{
		orientation:    orientation,
		gap:            1,
		activeCount:    0,
		mainAxisAlign:  AlignStart,
		crossAxisAlign: AlignStretch,
	}
}

// SetApplication sets the application this layout belongs to.
func (l *Layout) SetApplication(app *Application) {
	l.app = app

	// Propagate application to all panes
	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			l.panes[i].Pane.SetApplication(app)
		}
	}
}

// SetRect sets the layout's position and size.
func (l *Layout) SetRect(x, y, width, height int) {
	// Only process if dimensions actually changed
	if l.rect.X == x && l.rect.Y == y && l.rect.Width == width && l.rect.Height == height {
		return
	}

	l.rect = Rect{X: x, Y: y, Width: width, Height: height}

	// Calculate layout for all panes
	l.calculateLayout()
}

// GetRect returns the layout's position and size.
func (l *Layout) GetRect() (x, y, width, height int) {
	return l.rect.X, l.rect.Y, l.rect.Width, l.rect.Height
}

// AddPane adds a pane at the first available index.
// Returns the index (0-9) or -1 if no slots are available.
func (l *Layout) AddPane(pane *Pane, size Size) int {
	if pane == nil {
		return -1
	}

	// Validate size
	if size.FixedSize <= 0 && size.Proportion <= 0 {
		size.Proportion = 1 // Default to proportion 1
	}

	// Find first available slot
	index := -1
	for i := range l.panes {
		if !l.panes[i].Active {
			index = i
			break
		}
	}

	if index == -1 {
		return -1 // No slots available
	}

	// Store pane in the slot
	l.panes[index] = PaneInfo{
		Pane:   pane,
		Size:   size,
		Active: true,
	}

	// Set pane's index (1-based for user-facing index)
	pane.SetIndex(index + 1)

	// Set pane's application
	if l.app != nil {
		pane.SetApplication(l.app)
	}

	l.activeCount++

	// Recalculate layout
	l.calculateLayout()

	return index
}

// RemovePane removes a pane by its array index (0-9).
func (l *Layout) RemovePane(index int) {
	if index < 0 || index >= 10 {
		return
	}

	if !l.panes[index].Active {
		return // Slot is already inactive
	}

	// Clear pane's index
	if l.panes[index].Pane != nil {
		l.panes[index].Pane.SetIndex(-1)
	}

	// Mark slot as inactive
	l.panes[index].Active = false
	l.panes[index].Pane = nil
	l.activeCount--

	// Recalculate layout
	l.calculateLayout()
}

// SetGap sets the gap between panes.
func (l *Layout) SetGap(gap int) {
	if gap < 0 {
		gap = 0
	}

	if l.gap != gap {
		l.gap = gap
		l.calculateLayout()
	}
}

// SetMainAxisAlignment sets how panes are aligned along the main axis.
func (l *Layout) SetMainAxisAlignment(align Alignment) {
	if l.mainAxisAlign != align {
		l.mainAxisAlign = align
		l.calculateLayout()
	}
}

// SetCrossAxisAlignment sets how panes are aligned along the cross axis.
func (l *Layout) SetCrossAxisAlignment(align Alignment) {
	if l.crossAxisAlign != align {
		l.crossAxisAlign = align
		l.calculateLayout()
	}
}

// calculateLayout determines pane positions based on layout constraints.
func (l *Layout) calculateLayout() {
	// Skip if no panes or invalid dimensions
	if l.activeCount == 0 || l.rect.Width <= 0 || l.rect.Height <= 0 {
		return
	}

	// Calculate total fixed size and proportion
	totalFixedSize := 0
	totalProportion := 0

	for i := range l.panes {
		if !l.panes[i].Active || l.panes[i].Pane == nil {
			continue
		}

		size := l.panes[i].Size
		if size.FixedSize > 0 {
			totalFixedSize += size.FixedSize
		} else {
			totalProportion += size.Proportion
		}
	}

	// Calculate gaps
	totalGapSize := 0
	if l.activeCount > 1 {
		totalGapSize = l.gap * (l.activeCount - 1)
	}

	// Calculate available space for proportional panes
	availableSpace := 0
	if l.orientation == Horizontal {
		availableSpace = l.rect.Width - totalFixedSize - totalGapSize
	} else {
		availableSpace = l.rect.Height - totalFixedSize - totalGapSize
	}

	if availableSpace < 0 {
		availableSpace = 0
	}

	// Calculate offsets based on main axis alignment
	mainAxisOffset := 0
	if l.mainAxisAlign != AlignStart && availableSpace > 0 {
		switch l.mainAxisAlign {
		case AlignCenter:
			mainAxisOffset = availableSpace / 2
		case AlignEnd:
			mainAxisOffset = availableSpace
		}
	}

	// Calculate pane positions and sizes
	currentPos := mainAxisOffset
	if l.orientation == Horizontal {
		currentPos += l.rect.X
	} else {
		currentPos += l.rect.Y
	}

	for i := range l.panes {
		if !l.panes[i].Active || l.panes[i].Pane == nil {
			continue
		}

		// Calculate size along main axis
		paneMainSize := 0
		size := l.panes[i].Size

		if size.FixedSize > 0 {
			paneMainSize = size.FixedSize
		} else if totalProportion > 0 {
			// Calculate proportional size
			proportion := float64(size.Proportion) / float64(totalProportion)
			paneMainSize = int(float64(availableSpace) * proportion)

			// Ensure we don't create panes that are too small
			if paneMainSize < 3 {
				paneMainSize = 3 // Minimum size to accommodate borders
			}
		}

		// Calculate cross-axis position and size
		var paneX, paneY, paneWidth, paneHeight int

		if l.orientation == Horizontal {
			paneX = currentPos
			paneWidth = paneMainSize

			switch l.crossAxisAlign {
			case AlignStart:
				paneY = l.rect.Y
				paneHeight = l.rect.Height
			case AlignCenter:
				paneY = l.rect.Y
				paneHeight = l.rect.Height
			case AlignEnd:
				paneY = l.rect.Y
				paneHeight = l.rect.Height
			case AlignStretch:
				paneY = l.rect.Y
				paneHeight = l.rect.Height
			}

			currentPos += paneWidth + l.gap

		} else { // Vertical orientation
			paneY = currentPos
			paneHeight = paneMainSize

			switch l.crossAxisAlign {
			case AlignStart:
				paneX = l.rect.X
				paneWidth = l.rect.Width
			case AlignCenter:
				paneX = l.rect.X
				paneWidth = l.rect.Width
			case AlignEnd:
				paneX = l.rect.X
				paneWidth = l.rect.Width
			case AlignStretch:
				paneX = l.rect.X
				paneWidth = l.rect.Width
			}

			currentPos += paneHeight + l.gap
		}

		// Ensure panes don't extend beyond the layout's rect
		if paneX+paneWidth > l.rect.X+l.rect.Width {
			paneWidth = l.rect.X + l.rect.Width - paneX
		}
		if paneY+paneHeight > l.rect.Y+l.rect.Height {
			paneHeight = l.rect.Y + l.rect.Height - paneY
		}

		// Set pane's position and size
		l.panes[i].Pane.SetRect(paneX, paneY, paneWidth, paneHeight)
	}
}

// Draw draws all active panes.
func (l *Layout) Draw(screen tcell.Screen) {
	// Make a copy of active panes to avoid holding lock during drawing
	var activePanes []*Pane
	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			activePanes = append(activePanes, l.panes[i].Pane)
		}
	}
	app := l.app // Copy app reference

	for _, pane := range activePanes {
		// Check if this pane contains the focused component
		isChildFocused := false

		if app != nil {
			isChildFocused = l.isPaneFocused(pane)
		}

		pane.Draw(screen, isChildFocused)
	}
}

// isPaneFocused checks if the pane contains the focused component.
func (l *Layout) isPaneFocused(pane *Pane) bool {
	if l.app == nil || pane == nil {
		return false
	}

	// Get focused component from app
	focusedComp := l.app.GetFocusedComponent()
	if focusedComp == nil {
		return false
	}

	// Get all components in the pane - this might still acquire locks
	// but we're outside of the layout lock
	paneComps := pane.GetAllComponents()

	// Check if focused component is in this pane
	for _, comp := range paneComps {
		if comp == focusedComp {
			return true
		}
	}

	return false
}

// GetPaneByIndex returns the pane with the given user-facing index (1-10).
// Returns nil if no pane has that index.
func (l *Layout) GetPaneByIndex(userIndex int) *Pane {
	if userIndex < 1 || userIndex > 10 {
		return nil
	}

	// Convert to array index (0-9)
	index := userIndex - 1

	if !l.panes[index].Active {
		return nil
	}

	return l.panes[index].Pane
}

// GetAllFocusableComponents returns all focusable components in all panes.
func (l *Layout) GetAllFocusableComponents() []Component {
	var focusables []Component

	for i := range l.panes {
		if !l.panes[i].Active || l.panes[i].Pane == nil {
			continue
		}

		// Get focusable components from this pane
		comps := l.panes[i].Pane.GetFocusableComponents()
		focusables = append(focusables, comps...)
	}

	return focusables
}

// HasDirtyComponents checks if any pane or component needs redrawing.
func (l *Layout) HasDirtyComponents() bool {
	// Create a list of panes to check outside of lock
	var panesToCheck []*Pane
	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			panesToCheck = append(panesToCheck, l.panes[i].Pane)
		}
	}

	// Check dirty state outside of lock
	for _, pane := range panesToCheck {
		if pane.IsDirty() {
			return true
		}
	}

	return false
}

// ClearAllDirtyFlags clears dirty flags for all components recursively.
func (l *Layout) ClearAllDirtyFlags() {
	var panes []*Pane
	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			panes = append(panes, l.panes[i].Pane)
		}
	}

	// Clear dirty flags outside of lock
	for _, pane := range panes {
		pane.ClearDirtyFlags()
	}
}