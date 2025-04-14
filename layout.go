// layout.go
package tinytui

import (
	"log" // Keep log import for potential minimal logging if needed later

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
	app            *Application // Reference to the application
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
		gap:            1, // Default gap
		activeCount:    0,
		mainAxisAlign:  AlignStart,   // Default alignment
		crossAxisAlign: AlignStretch, // Default alignment
	}
}

// SetApplication sets the application this layout belongs to.
// It also triggers index assignment for top-level panes if this layout
// is set as the application's main layout.
func (l *Layout) SetApplication(app *Application) {
	// Avoid redundant work or potential loops if called multiple times with same app
	if l.app == app && app != nil {
		// If app is already set and hasn't changed, maybe re-check main layout status?
		// This might be needed if app.layout changes elsewhere, but usually SetLayout handles it.
		// For now, assume SetLayout is the primary trigger.
		// return
	}
	l.app = app

	// Determine if this layout is the application's main layout *right now*
	isNowMainLayout := false
	if app != nil && app.layout == l {
		isNowMainLayout = true
		// log.Printf("[Layout.SetApplication] Layout %p confirmed as main layout for app %p.", l, app)
	} else {
		// log.Printf("[Layout.SetApplication] Layout %p is NOT main layout for app %p.", l, app)
	}

	// Propagate application reference to all child panes.
	// Assign user-facing indices (1-10) ONLY if this is the main layout.
	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			pane := l.panes[i].Pane

			// 1. Always propagate the app reference down
			pane.SetApplication(app)

			// 2. Assign or reset the user-facing index based on whether this is the main layout
			if isNowMainLayout {
				paneIndex := i + 1 // 1-based index for Alt+N navigation
				// log.Printf("[Layout.SetApplication] Assigning index %d to Pane (Title: '%s').", paneIndex, pane.title)
				pane.SetIndex(paneIndex) // Set the index (1-10)
			} else {
				// If this layout is NOT the main layout (or app is nil), ensure its panes have no user-facing index.
				if pane.GetIndex() != 0 { // Only reset if it currently has an index > 0
					// log.Printf("[Layout.SetApplication] Resetting index for Pane (Title: '%s') to 0.", pane.title)
					pane.SetIndex(0) // Reset index to 0
				}
			}
		}
	}
}

// SetRect sets the layout's position and size, triggering recalculation.
func (l *Layout) SetRect(x, y, width, height int) {
	newRect := Rect{X: x, Y: y, Width: width, Height: height}
	if l.rect == newRect {
		return // No change
	}
	l.rect = newRect
	l.calculateLayout() // Recalculate positions based on new size
}

// GetRect returns the layout's position and size.
func (l *Layout) GetRect() (x, y, width, height int) {
	return l.rect.X, l.rect.Y, l.rect.Width, l.rect.Height
}

// AddPane adds a pane at the first available index (0-9).
// Index assignment (1-10) is handled by SetApplication when the layout becomes the main one.
func (l *Layout) AddPane(pane *Pane, size Size) int {
	if pane == nil {
		log.Println("[Layout.AddPane] Attempted to add a nil pane.")
		return -1
	}
	// Default size if invalid
	if size.FixedSize <= 0 && size.Proportion <= 0 {
		size.Proportion = 1
	}

	// Find first available slot (0-9)
	index := -1
	for i := range l.panes {
		if !l.panes[i].Active {
			index = i
			break
		}
	}
	if index == -1 {
		log.Println("[Layout.AddPane] No available slots in layout.")
		return -1 // No slots available
	}

	// Store pane info
	l.panes[index] = PaneInfo{
		Pane:   pane,
		Size:   size,
		Active: true,
	}

	// Set pane's application reference IF the layout already has one.
	// This handles adding panes to layouts that are already part of the app structure.
	if l.app != nil {
		pane.SetApplication(l.app)
	}

	// Set the internal index to 0 initially.
	// SetApplication will assign the correct user-facing index (1-10) later
	// if this layout becomes the application's main layout.
	pane.SetIndex(0)

	l.activeCount++
	l.calculateLayout() // Recalculate needed for size/position

	// log.Printf("[Layout.AddPane] Added Pane (Title: '%s') to layout %p at internal slot %d.", pane.title, l, index)
	return index // Return the internal slot index (0-9)
}

// RemovePane removes a pane by its array index (0-9).
func (l *Layout) RemovePane(index int) {
	if index < 0 || index >= 10 || !l.panes[index].Active {
		return // Invalid index or slot already inactive
	}

	// Clear pane's index and potentially other references if needed
	if l.panes[index].Pane != nil {
		l.panes[index].Pane.SetIndex(0) // Ensure index is cleared
		// Consider detaching app reference? pane.SetApplication(nil)? Depends on desired lifecycle.
	}

	// Mark slot as inactive
	l.panes[index] = PaneInfo{} // Zero out the struct
	l.activeCount--

	l.calculateLayout() // Recalculate layout
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
	if l.activeCount == 0 || l.rect.Width <= 0 || l.rect.Height <= 0 {
		return // Nothing to lay out or no space
	}

	// --- Calculate total fixed size and total proportion ---
	totalFixedSize := 0
	totalProportion := 0.0    // Use float for proportion calculation
	activePanesForLayout := 0 // Count panes participating in layout sizing

	for i := range l.panes {
		if !l.panes[i].Active || l.panes[i].Pane == nil {
			continue
		}
		activePanesForLayout++
		size := l.panes[i].Size
		if size.FixedSize > 0 {
			totalFixedSize += size.FixedSize
		} else if size.Proportion > 0 { // Ensure proportion is positive
			totalProportion += float64(size.Proportion)
		} else {
			// If neither fixed nor proportion > 0, treat as proportion 1 (common default)
			// This prevents division by zero if only fixed-size panes exist.
			// However, if totalProportion remains 0, proportional panes get no space.
			// A better default might be needed if mixing fixed and zero-proportion panes.
			// For now, assume valid proportions or fixed sizes are given.
			// If a pane has Size{}, it defaults to Proportion=1 in AddPane.
		}
	}

	// --- Calculate total gap size ---
	totalGapSize := 0
	if activePanesForLayout > 1 {
		totalGapSize = l.gap * (activePanesForLayout - 1)
	}

	// --- Calculate available space for proportional panes ---
	availableProportionalSpace := 0
	mainAxisSize := 0
	crossAxisSize := 0

	if l.orientation == Horizontal {
		mainAxisSize = l.rect.Width
		crossAxisSize = l.rect.Height
	} else { // Vertical
		mainAxisSize = l.rect.Height
		crossAxisSize = l.rect.Width
	}
	availableProportionalSpace = mainAxisSize - totalFixedSize - totalGapSize
	if availableProportionalSpace < 0 {
		availableProportionalSpace = 0 // Cannot have negative space
	}

	// --- Distribute space and set pane rects ---
	currentPos := 0 // Start position along the main axis within the layout rect
	allocatedProportionalSpace := 0

	// Use a slice to iterate only over active panes in order
	activePaneIndices := []int{}
	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			activePaneIndices = append(activePaneIndices, i)
		}
	}

	for idx, paneArrIndex := range activePaneIndices {
		paneInfo := l.panes[paneArrIndex]
		pane := paneInfo.Pane
		size := paneInfo.Size

		// Calculate size along main axis
		paneMainSize := 0
		if size.FixedSize > 0 {
			paneMainSize = size.FixedSize
		} else if totalProportion > 0 && size.Proportion > 0 {
			proportion := float64(size.Proportion) / totalProportion
			paneMainSize = int(float64(availableProportionalSpace) * proportion)
			allocatedProportionalSpace += paneMainSize
		} else {
			// Should not happen if AddPane defaults correctly, but handle defensively
			paneMainSize = 0
		}

		// --- Handle rounding errors / remaining space for the *last* proportional pane ---
		// If this is the last pane in the loop AND there was proportional space to distribute
		isLastPane := (idx == len(activePaneIndices)-1)
		if isLastPane && totalProportion > 0 {
			remainder := availableProportionalSpace - allocatedProportionalSpace
			paneMainSize += remainder // Give remainder to the last proportional pane
		}

		// Ensure minimum size (e.g., for borders) - adjust as needed
		minSize := 1
		if pane.border != BorderNone {
			minSize = 3 // Need space for corners + content
		}
		if paneMainSize < minSize {
			// This might steal space needed by others if not careful.
			// A more complex priority system might be needed for robust minimums.
			// For now, just ensure it's at least minSize if possible.
			// paneMainSize = minSize // Be cautious with enforcing minimums this way
		}
		if paneMainSize < 0 {
			paneMainSize = 0 // Cannot be negative
		}

		// Calculate cross-axis size (usually stretch)
		paneCrossSize := 0
		switch l.crossAxisAlign {
		case AlignStretch:
			paneCrossSize = crossAxisSize
		// TODO: Implement AlignStart, AlignCenter, AlignEnd for cross-axis if needed
		// These would require calculating the component's preferred size.
		default:
			paneCrossSize = crossAxisSize // Default to stretch
		}
		if paneCrossSize < 0 {
			paneCrossSize = 0
		}

		// Determine X, Y, Width, Height based on orientation
		var paneX, paneY, paneWidth, paneHeight int
		if l.orientation == Horizontal {
			paneX = l.rect.X + currentPos
			paneY = l.rect.Y // Cross-axis starts at layout top
			paneWidth = paneMainSize
			paneHeight = paneCrossSize

			// Apply cross-axis alignment (if not stretch)
			// if l.crossAxisAlign == AlignCenter { paneY += (crossAxisSize - paneHeight) / 2 }
			// if l.crossAxisAlign == AlignEnd { paneY += crossAxisSize - paneHeight }

		} else { // Vertical
			paneX = l.rect.X // Cross-axis starts at layout left
			paneY = l.rect.Y + currentPos
			paneWidth = paneCrossSize
			paneHeight = paneMainSize

			// Apply cross-axis alignment (if not stretch)
			// if l.crossAxisAlign == AlignCenter { paneX += (crossAxisSize - paneWidth) / 2 }
			// if l.crossAxisAlign == AlignEnd { paneX += crossAxisSize - paneWidth }
		}

		// Clamp dimensions to layout bounds (safety check)
		if paneX < l.rect.X {
			paneX = l.rect.X
		}
		if paneY < l.rect.Y {
			paneY = l.rect.Y
		}
		if paneX+paneWidth > l.rect.X+l.rect.Width {
			paneWidth = l.rect.X + l.rect.Width - paneX
		}
		if paneY+paneHeight > l.rect.Y+l.rect.Height {
			paneHeight = l.rect.Y + l.rect.Height - paneY
		}
		if paneWidth < 0 {
			paneWidth = 0
		}
		if paneHeight < 0 {
			paneHeight = 0
		}

		// Set the calculated rectangle on the pane
		pane.SetRect(paneX, paneY, paneWidth, paneHeight)

		// Advance position for the next pane
		currentPos += paneMainSize + l.gap
	}
}

// Draw draws all active panes.
func (l *Layout) Draw(screen tcell.Screen) {
	// Create a list of panes to draw to avoid issues if panes are modified concurrently
	panesToDraw := make([]*Pane, 0, l.activeCount)
	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			panesToDraw = append(panesToDraw, l.panes[i].Pane)
		}
	}

	focusedComp := l.app.GetFocusedComponent() // Get focused component once

	for _, pane := range panesToDraw {
		// Check if this pane (or one of its children) has focus
		isChildFocused := false
		if focusedComp != nil {
			// Check if the focused component is this pane's direct child
			if childComp, ok := pane.child.(Component); ok && childComp == focusedComp {
				isChildFocused = true
			} else if childLayout, ok := pane.child.(*Layout); ok {
				// If child is a layout, recursively check if it contains the focused component
				isChildFocused = childLayout.ContainsFocus(focusedComp)
			}
		}
		pane.Draw(screen, isChildFocused)
	}
}

// ContainsFocus checks recursively if this layout or any of its children contain the focused component.
func (l *Layout) ContainsFocus(focused Component) bool {
	if focused == nil {
		return false
	}
	for i := range l.panes {
		if !l.panes[i].Active || l.panes[i].Pane == nil {
			continue
		}
		pane := l.panes[i].Pane
		// Check direct component child
		if childComp, ok := pane.child.(Component); ok && childComp == focused {
			return true
		}
		// Check child layout recursively
		if childLayout, ok := pane.child.(*Layout); ok {
			if childLayout.ContainsFocus(focused) {
				return true
			}
		}
	}
	return false
}

// GetPaneByIndex returns the pane with the given user-facing index (1-10).
// Returns nil if no pane has that index or index is out of range.
func (l *Layout) GetPaneByIndex(userIndex int) *Pane {
	if userIndex < 1 || userIndex > 10 {
		return nil // Invalid user-facing index
	}
	// Convert to internal array index (0-9)
	index := userIndex - 1

	// Check if the slot is active and the pane exists
	if !l.panes[index].Active || l.panes[index].Pane == nil {
		return nil
	}
	// Verify the pane actually has this index assigned (safety check)
	if l.panes[index].Pane.GetIndex() != userIndex {
		// This case should theoretically not happen if SetApplication works correctly
		log.Printf("[Layout.GetPaneByIndex] WARNING: Pane at slot %d has index %d, expected %d.", index, l.panes[index].Pane.GetIndex(), userIndex)
		return nil // Or return the pane anyway? Returning nil is safer.
	}

	return l.panes[index].Pane
}

// GetAllFocusableComponents returns all focusable components in all panes recursively.
func (l *Layout) GetAllFocusableComponents() []Component {
	var focusables []Component
	for i := range l.panes {
		if !l.panes[i].Active || l.panes[i].Pane == nil {
			continue
		}
		// Delegate to pane's recursive search
		focusables = append(focusables, l.panes[i].Pane.GetFocusableComponents()...)
	}
	return focusables
}

// HasDirtyComponents checks if any pane or component within this layout needs redrawing.
func (l *Layout) HasDirtyComponents() bool {
	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			if l.panes[i].Pane.IsDirty() { // IsDirty checks recursively
				return true
			}
		}
	}
	return false
}

// ClearAllDirtyFlags clears dirty flags for all panes and components recursively.
func (l *Layout) ClearAllDirtyFlags() {
	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			l.panes[i].Pane.ClearDirtyFlags() // ClearDirtyFlags clears recursively
		}
	}
}