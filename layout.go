// layout.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
)

// Layout organizes Panes on screen, arranging them horizontally or vertically
// according to size constraints and alignment rules.
type Layout struct {
	panes          [10]PaneInfo // Fixed-size array for panes (indices 0-9 map to user indices 1-10)
	orientation    Orientation  // Horizontal or Vertical arrangement of panes
	gap            int          // Size of the gap (in cells) between panes
	activeCount    int          // Number of active panes currently in the layout
	mainAxisAlign  Alignment    // Alignment along the layout's main axis (Start, Center, End) - Currently affects initial position
	crossAxisAlign Alignment    // Alignment along the cross axis (Start, Center, End, Stretch) - Affects pane size/position perpendicular to orientation
	rect           Rect         // The screen area allocated to this layout
	app            *Application // Reference to the parent application
	style          Style        // Background style for the layout area itself (fills gaps between panes)
}

// PaneInfo stores a reference to a Pane and its associated layout constraints (Size).
type PaneInfo struct {
	Pane   *Pane
	Size   Size // How the pane should be sized (Fixed or Proportional)
	Active bool // Is this slot in the 'panes' array currently occupied?
}

// NewLayout creates a new layout with the specified orientation.
// Initializes background style from the current theme.
func NewLayout(orientation Orientation) *Layout {
	theme := GetTheme()
	if theme == nil {
		theme = NewDefaultTheme()
	} // Fallback

	l := &Layout{
		orientation:    orientation,
		gap:            1, // Default gap of 1 cell
		activeCount:    0,
		mainAxisAlign:  AlignStart,        // Default main axis alignment (panes start at top/left)
		crossAxisAlign: AlignStretch,      // Default cross axis alignment (panes fill perpendicular space)
		style:          theme.PaneStyle(), // Use theme's pane style for layout background by default
		// panes array is zero-initialized
	}
	return l
}

// ApplyThemeRecursively applies the theme to the layout itself and propagates it to all child panes.
func (l *Layout) ApplyThemeRecursively(theme Theme) {
	if theme == nil {
		return
	}

	// Apply theme to the layout's background style
	l.style = theme.PaneStyle()

	// Apply theme to all active child panes (which will then apply to their children)
	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			l.panes[i].Pane.ApplyThemeRecursively(theme) // Pane handles its own style and recursive application
		}
	}
	// No MarkDirty needed here, theme change on children will mark them dirty.
}

// SetStyle explicitly sets the background style used for the layout's own area (filling gaps).
// Consider using themes instead for consistent styling.
func (l *Layout) SetStyle(style Style) {
	if l.style != style {
		l.style = style
		// Mark layout's children dirty? Or assume redraw will happen anyway?
		// Let's rely on redraw triggered elsewhere or component dirtiness.
	}
}

// SetApplication associates the layout with an application instance.
// Propagates app reference, sets slot indices for direct children,
// and triggers initial nav index assignment if this is the root layout.
func (l *Layout) SetApplication(app *Application) {
	if l.app == app && app != nil {
		return
	} // Avoid redundant calls
	l.app = app

	isRootLayout := app != nil && app.GetLayout() == l

	// Propagate app reference and SET SLOT INDEX for direct children
	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			pane := l.panes[i].Pane
			pane.SetApplication(app) // Propagate app reference down
			// Assign the internal SLOT index (0-9) based on its position in this layout
			pane.setSlotIndex(i)
		}
	}

	// Assign navigation indices if this is the root layout
	if isRootLayout {
		l.assignNavigationIndices()
	} else {
		// Ensure nested panes have navIndex 0 (might be redundant but safe)
		for i := range l.panes {
			if l.panes[i].Active && l.panes[i].Pane != nil {
				l.panes[i].Pane.SetNavIndex(0)
			}
		}
	}

	// Apply theme recursively AFTER setting app and indices
	// This ensures components have app context when theme is applied
	if l.app != nil {
		currentTheme := l.app.GetTheme()
		if currentTheme != nil {
			l.ApplyThemeRecursively(currentTheme)
		}
	}
}

// SetRect sets the layout's allocated position and size on the screen.
// Triggers recalculation of child pane positions and sizes if the rectangle changes.
func (l *Layout) SetRect(x, y, width, height int) {
	newRect := Rect{X: x, Y: y, Width: width, Height: height}
	if l.rect == newRect {
		return // No change in dimensions, no recalculation needed
	}
	l.rect = newRect
	l.calculateLayout() // Recalculate child positions based on the new size
}

// GetRect returns the layout's current allocated position and size.
func (l *Layout) GetRect() (x, y, width, height int) {
	return l.rect.X, l.rect.Y, l.rect.Width, l.rect.Height
}

// AddPane adds a pane to the layout.
// Triggers layout calculation and navigation index recalculation via command.
func (l *Layout) AddPane(pane *Pane, size Size) int {
	if pane == nil {
		return -1
	}
	if size.FixedSize <= 0 && size.Proportion <= 0 {
		size.Proportion = 1
	}

	index := -1 // This is the slot index
	for i := range l.panes {
		if !l.panes[i].Active {
			index = i
			break
		}
	}
	if index == -1 {
		return -1
	} // No available slots

	l.panes[index] = PaneInfo{Pane: pane, Size: size, Active: true}
	l.activeCount++

	// Set app reference and SLOT index
	if l.app != nil {
		pane.SetApplication(l.app) // Propagates app ref
	}
	// Always set the slot index based on where it was added
	pane.setSlotIndex(index)
	// Ensure navIndex starts at 0 until recalculated
	pane.SetNavIndex(0)

	// Apply theme if app context exists
	if l.app != nil {
		currentTheme := l.app.GetTheme()
		if currentTheme != nil {
			pane.ApplyThemeRecursively(currentTheme)
		}
	}

	l.calculateLayout() // Recalculate geometry

	// Dispatch command to recalculate navigation indices for the root layout
	// Only dispatch if this layout is part of an application context.
	if l.app != nil && l.app.GetLayout() != nil {
		// Find the root layout to trigger the command on it
		// Note: This assumes GetLayout() returns the actual root.
		// A more robust way might involve traversing up if nested layouts were complex.
		rootLayout := l.app.GetLayout()
		if rootLayout != nil {
			// Dispatch command associated with the application instance
			l.app.Dispatch(&RecalculateNavIndicesCommand{})
		}
	}

	return index
}

// RemovePane removes a pane from the layout by slot index.
// Triggers layout calculation and navigation index recalculation via command.
func (l *Layout) RemovePane(index int) { // index here refers to slot index
	if index < 0 || index >= 10 || !l.panes[index].Active {
		return
	}

	// Clear indices from the pane being removed
	if pane := l.panes[index].Pane; pane != nil {
		pane.setSlotIndex(0) // Reset slot index
		pane.SetNavIndex(0)  // Ensure nav index is cleared
	}

	l.panes[index] = PaneInfo{} // Clear the slot
	l.activeCount--

	l.calculateLayout() // Recalculate geometry

	// Dispatch command to recalculate navigation indices for the root layout
	if l.app != nil && l.app.GetLayout() != nil {
		rootLayout := l.app.GetLayout()
		if rootLayout != nil {
			l.app.Dispatch(&RecalculateNavIndicesCommand{})
		}
	}
}

// SetGap sets the spacing (in cells) between panes in the layout.
func (l *Layout) SetGap(gap int) {
	if gap < 0 {
		gap = 0
	} // Ensure gap is non-negative
	if l.gap != gap {
		l.gap = gap
		l.calculateLayout() // Recalculate layout with the new gap
	}
}

// SetMainAxisAlignment sets the alignment of panes along the main axis (Vertical/Horizontal).
// Affects where panes start if there's extra space along the main axis.
func (l *Layout) SetMainAxisAlignment(align Alignment) {
	if l.mainAxisAlign != align {
		l.mainAxisAlign = align
		l.calculateLayout()
	}
}

// SetCrossAxisAlignment sets the alignment of panes along the axis perpendicular to the orientation.
// Affects pane size and position along the cross axis (e.g., Stretch, Start, Center).
func (l *Layout) SetCrossAxisAlignment(align Alignment) {
	if l.crossAxisAlign != align {
		l.crossAxisAlign = align
		l.calculateLayout()
	}
}

// calculateLayout recalculates the position and size of all active child panes
// based on the layout's orientation, size constraints, gap, and alignment settings.
func (l *Layout) calculateLayout() {
	if l.activeCount == 0 || l.rect.Width <= 0 || l.rect.Height <= 0 {
		// Hide inactive panes explicitly? Or just don't draw them?
		// Current approach: Don't calculate/set rect, Draw loop skips inactive.
		return
	}

	// --- 1. Determine Axis Sizes and Available Space ---
	mainAxisSize := 0  // Size along the layout direction (Width for Horizontal, Height for Vertical)
	crossAxisSize := 0 // Size perpendicular to layout direction
	isVertical := l.orientation == Vertical

	if isVertical {
		mainAxisSize = l.rect.Height
		crossAxisSize = l.rect.Width
	} else { // Horizontal
		mainAxisSize = l.rect.Width
		crossAxisSize = l.rect.Height
	}

	// Calculate total space needed for gaps
	totalGapSize := 0
	if l.activeCount > 1 {
		totalGapSize = l.gap * (l.activeCount - 1)
	}

	// Calculate space available for the panes themselves
	totalAvailablePaneSpace := mainAxisSize - totalGapSize
	if totalAvailablePaneSpace < 0 {
		totalAvailablePaneSpace = 0 // Cannot have negative space
	}

	// --- 2. Identify Pane Types and Calculate Requested Sizes ---
	totalFixedRequested := 0            // Sum of fixed sizes requested
	totalProportionSum := 0.0           // Sum of proportions for proportional panes
	fixedPaneIndices := []int{}         // Indices of panes with FixedSize > 0
	proportionalPaneIndices := []int{}  // Indices of panes with Proportion > 0
	activePaneIndicesInOrder := []int{} // All active indices in their slot order

	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			activePaneIndicesInOrder = append(activePaneIndicesInOrder, i)
			size := l.panes[i].Size
			if size.FixedSize > 0 {
				totalFixedRequested += size.FixedSize
				fixedPaneIndices = append(fixedPaneIndices, i)
			} else if size.Proportion > 0 {
				totalProportionSum += float64(size.Proportion)
				proportionalPaneIndices = append(proportionalPaneIndices, i)
			}
			// Panes with size {0, 0} are ignored in size calculation unless explicitly handled.
		}
	}

	// --- 3. Calculate Final Sizes, Handling Insufficient Space ---
	paneSizes := make(map[int]int) // Stores the final calculated main axis size for each pane index
	totalAllocatedFixed := 0       // Track allocated fixed space for main axis alignment calc

	// Allocate space for fixed-size panes first
	if totalFixedRequested <= totalAvailablePaneSpace {
		// Case A: Fixed panes fit (or there's extra space)
		for _, idx := range fixedPaneIndices {
			paneSizes[idx] = l.panes[idx].Size.FixedSize // Assign requested fixed size
			totalAllocatedFixed += paneSizes[idx]
		}
	} else {
		// Case B: Fixed panes DO NOT fit. Distribute available space proportionally among them.
		if totalFixedRequested > 0 { // Avoid division by zero
			// Distribute totalAvailablePaneSpace based on the *ratio* of requested fixed sizes
			allocatedSpace := 0
			for _, idx := range fixedPaneIndices {
				proportion := float64(l.panes[idx].Size.FixedSize) / float64(totalFixedRequested)
				alloc := int(float64(totalAvailablePaneSpace) * proportion) // Floor
				paneSizes[idx] = alloc
				allocatedSpace += alloc
			}
			// Distribute any remaining pixels due to flooring/rounding fairly
			remainder := totalAvailablePaneSpace - allocatedSpace
			fixedCount := len(fixedPaneIndices)
			for i := 0; i < remainder; i++ {
				idx := fixedPaneIndices[i%fixedCount] // Cycle through fixed panes
				paneSizes[idx]++
			}
			totalAllocatedFixed = totalAvailablePaneSpace // All available space used by fixed
		} else {
			totalAllocatedFixed = 0 // No fixed panes requested space
		}
	}

	spaceLeftForProportionals := totalAvailablePaneSpace - totalAllocatedFixed
	if spaceLeftForProportionals < 0 {
		spaceLeftForProportionals = 0
	} // Safety check

	totalAllocatedProportional := 0
	// Allocate remaining space for proportional panes (if any space and panes exist)
	if totalProportionSum > 0 && spaceLeftForProportionals > 0 {
		allocatedSpace := 0
		// Distribute spaceLeftForProportionals based on proportions
		for _, idx := range proportionalPaneIndices {
			proportion := float64(l.panes[idx].Size.Proportion) / totalProportionSum
			size := int(float64(spaceLeftForProportionals) * proportion) // Floor
			paneSizes[idx] = size
			allocatedSpace += size
		}
		// Distribute remainder pixels fairly
		remainder := spaceLeftForProportionals - allocatedSpace
		propCount := len(proportionalPaneIndices)
		for i := 0; i < remainder; i++ {
			idx := proportionalPaneIndices[i%propCount] // Cycle through proportional panes
			paneSizes[idx]++
		}
		totalAllocatedProportional = spaceLeftForProportionals // All remaining space used
	} else {
		// No space left or no proportional panes, ensure they get size 0
		for _, idx := range proportionalPaneIndices {
			paneSizes[idx] = 0
		}
		totalAllocatedProportional = 0
	}

	// --- 4. Calculate and Set Final Rects based on calculated sizes and alignment ---
	totalAllocatedMainSize := totalAllocatedFixed + totalAllocatedProportional
	extraMainSpace := totalAvailablePaneSpace - totalAllocatedMainSize // Usually 0, but > 0 if only fixed panes requested less than available
	if extraMainSpace < 0 {
		extraMainSpace = 0
	}

	// Determine starting position based on main axis alignment
	currentMainPos := 0 // Default for AlignStart
	if l.mainAxisAlign == AlignCenter {
		currentMainPos = extraMainSpace / 2
	} else if l.mainAxisAlign == AlignEnd {
		currentMainPos = extraMainSpace
	}
	// Add layout's own origin offset
	baseX, baseY := l.rect.X, l.rect.Y
	currentMainPos += 0 // Relative position within layout rect

	for _, paneArrIndex := range activePaneIndicesInOrder {
		paneInfo := l.panes[paneArrIndex]
		pane := paneInfo.Pane
		paneMainSize := paneSizes[paneArrIndex] // Size along layout orientation

		if paneMainSize < 0 {
			paneMainSize = 0
		} // Ensure non-negative size

		// Calculate cross-axis size and position based on alignment
		paneCrossSize := 0
		crossPos := 0 // Position offset along the cross axis, relative to layout rect edge

		switch l.crossAxisAlign {
		case AlignStretch:
			paneCrossSize = crossAxisSize // Stretch to fill cross axis
			crossPos = 0
		case AlignStart:
			// Requires knowing preferred size. For now, assume minimal? Or just position at start?
			// Let's assume it means position at start, but still give full cross size.
			// A better implementation might query the pane/component.
			paneCrossSize = crossAxisSize // Give full size for now
			crossPos = 0
		case AlignCenter:
			// Assume full size, centered position (effectively same as stretch if pane fills)
			paneCrossSize = crossAxisSize
			// crossPos = (crossAxisSize - paneCrossSize) / 2 // Centering needs actual size if not stretching
			crossPos = 0 // Treat as stretch for now
		case AlignEnd:
			// Assume full size, position at end
			paneCrossSize = crossAxisSize
			// crossPos = crossAxisSize - paneCrossSize // Needs actual size
			crossPos = 0 // Treat as stretch for now
		default:
			paneCrossSize = crossAxisSize // Default to stretch
			crossPos = 0
		}
		if paneCrossSize < 0 {
			paneCrossSize = 0
		}

		// Determine final X, Y, Width, Height based on orientation and calculated values
		var paneX, paneY, paneWidth, paneHeight int
		if isVertical {
			paneX = baseX + crossPos       // X determined by cross axis position
			paneY = baseY + currentMainPos // Y determined by main axis position
			paneWidth = paneCrossSize      // Width is cross axis size
			paneHeight = paneMainSize      // Height is main axis size
		} else { // Horizontal
			paneX = baseX + currentMainPos // X determined by main axis position
			paneY = baseY + crossPos       // Y determined by cross axis position
			paneWidth = paneMainSize       // Width is main axis size
			paneHeight = paneCrossSize     // Height is cross axis size
		}

		// Set the calculated rectangle for the child pane
		pane.SetRect(paneX, paneY, paneWidth, paneHeight)

		// Advance position for the next pane, including the gap (only if size > 0)
		if paneMainSize > 0 {
			currentMainPos += paneMainSize + l.gap
		}
	}
}

// countTopLevelFocusablePanes counts the number of direct child panes of this layout
// that contain at least one focusable component. It also returns the index of the
// pane if exactly one such pane is found.
func (l *Layout) countTopLevelFocusablePanes() (count int, singlePaneIndex int) {
	count = 0
	singlePaneIndex = -1 // Initialize to invalid index

	// Only consider panes directly managed by *this* layout
	isRootLayout := l.app != nil && l.app.GetLayout() == l
	if !isRootLayout {
		// If this isn't the root layout, the single-pane rule doesn't apply here.
		return 0, -1
	}

	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			if l.panes[i].Pane.HasFocusableChild() {
				count++
				singlePaneIndex = i + 1 // Store the user-facing index (1-based)
			}
		}
	}

	// If count is not exactly 1, reset singlePaneIndex to invalid
	if count != 1 {
		singlePaneIndex = -1
	}

	return count, singlePaneIndex
}

// Draw draws the layout background and its active panes.
func (l *Layout) Draw(screen tcell.Screen) {
	if l.rect.Width <= 0 || l.rect.Height <= 0 {
		return
	}
	Fill(screen, l.rect.X, l.rect.Y, l.rect.Width, l.rect.Height, ' ', l.style)

	focusedComp := l.app.GetFocusedComponent() // Okay if app is nil

	// Draw each active pane
	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			pane := l.panes[i].Pane
			isChildFocused := false
			if focusedComp != nil {
				isChildFocused = pane.ContainsFocus(focusedComp)
			}
			// Pass only focus info to pane's Draw (no more single pane rule)
			pane.Draw(screen, isChildFocused)
		}
	}
}

// ContainsFocus checks recursively if this layout or any of its descendant panes/layouts
// contain the specified focused component.
func (l *Layout) ContainsFocus(focused Component) bool {
	if focused == nil {
		return false
	}

	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			// Delegate the check to the pane, which handles its own child type
			if l.panes[i].Pane.ContainsFocus(focused) {
				return true
			}
		}
	}
	return false // Focus not found in any child pane
}

// GetPaneBySlotIndex returns the pane at the specified internal slot index (0-9).
func (l *Layout) GetPaneBySlotIndex(slotIndex int) *Pane {
	if slotIndex < 0 || slotIndex >= 10 || !l.panes[slotIndex].Active || l.panes[slotIndex].Pane == nil {
		return nil
	}
	return l.panes[slotIndex].Pane
}

// GetPaneByNavIndex returns the first pane matching the user navigation index (1-10).
// Iterates in slot order to ensure Alt+1 targets the *first* eligible pane.
func (l *Layout) GetPaneByNavIndex(navIndex int) *Pane {
	if navIndex < 1 || navIndex > 10 {
		return nil
	} // Validate nav index range
	for i := range l.panes { // Check in slot order (0-9)
		if l.panes[i].Active && l.panes[i].Pane != nil {
			if l.panes[i].Pane.GetNavIndex() == navIndex {
				return l.panes[i].Pane // Found the pane with the matching navIndex
			}
		}
	}
	return nil // Not found
}

// GetAllFocusableComponents returns a slice of all focusable components
// found recursively within this layout's active panes, in the order they appear.
func (l *Layout) GetAllFocusableComponents() []Component {
	// Estimate capacity based on active count? Might be inaccurate.
	var focusables []Component
	for i := range l.panes { // Iterate in slot order
		if l.panes[i].Active && l.panes[i].Pane != nil {
			// Append focusable components found within each active pane
			focusables = append(focusables, l.panes[i].Pane.GetFocusableComponents()...)
		}
	}
	return focusables
}

// HasDirtyComponents checks if the layout itself or any of its descendant panes
// or components are marked as dirty (need redrawing).
func (l *Layout) HasDirtyComponents() bool {
	// Note: Layout itself doesn't have its own dirty flag, it depends on children.
	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			// Delegate check to the pane (which checks its child recursively)
			if l.panes[i].Pane.IsDirty() {
				return true // Found a dirty descendant
			}
		}
	}
	return false // No dirty components found
}

// ClearAllDirtyFlags recursively clears the dirty flag for all descendant panes and components.
// Called by the application after a successful draw cycle.
func (l *Layout) ClearAllDirtyFlags() {
	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			// Delegate clearing to the pane
			l.panes[i].Pane.ClearDirtyFlags()
		}
	}
}

// assignNavigationIndices scans through the direct children (panes) of this layout
// and assigns sequential navigation indices (1-10) only to those that contain
// focusable components.
// This should ONLY be called on the application's root layout.
func (l *Layout) assignNavigationIndices() {
	// Ensure this is only run in the context of an application and its root layout
	if l.app == nil || l.app.GetLayout() != l {
		// If called on a nested layout (shouldn't happen via command), ensure its direct children have navIndex 0
		// This might be redundant if SetApplication handles it, but provides safety.
		for i := range l.panes {
			if l.panes[i].Active && l.panes[i].Pane != nil {
				l.panes[i].Pane.SetNavIndex(0)
			}
		}
		return
	}

	currentNavIndex := 1 // Start assigning from 1
	// Iterate through panes in their slot order (0-9)
	for i := range l.panes {
		// Reset navIndex first before potentially assigning a new one
		// Important if a previously navigable pane becomes non-navigable
		if l.panes[i].Active && l.panes[i].Pane != nil {
			pane := l.panes[i].Pane
			assignedIndex := 0 // Default to 0 (not navigable)

			// Check if the pane is eligible: contains focusable children and we haven't assigned 10 indices yet
			if pane.HasFocusableChild() && currentNavIndex <= 10 {
				assignedIndex = currentNavIndex
				currentNavIndex++ // Increment for the next eligible pane
			}
			pane.SetNavIndex(assignedIndex) // Set the calculated index (0 or 1-10)
		} else if l.panes[i].Pane != nil {
			// Ensure inactive panes also have navIndex cleared
			l.panes[i].Pane.SetNavIndex(0)
		}
	}
	// Panes that were inactive, nil, or not focusable will have navIndex 0.
	// Panes beyond the 10th focusable one will also have navIndex 0.
}