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
	style          Style
}

// PaneInfo holds a pane and its layout constraints.
type PaneInfo struct {
	Pane   *Pane
	Size   Size
	Active bool // Is this slot used?
}

// NewLayout creates a new layout with the specified orientation.
func NewLayout(orientation Orientation) *Layout {
	layoutStyle := DefaultStyle
	theme := GetTheme()
	if theme != nil {
		layoutStyle = theme.PaneStyle() // Use PaneStyle for layout background
	}

	return &Layout{
		orientation:    orientation,
		gap:            1,
		activeCount:    0,
		mainAxisAlign:  AlignStart,
		crossAxisAlign: AlignStretch,
		style:          layoutStyle,
	}
}

// SetStyle sets the background style used for clearing the layout's area.
func (l *Layout) SetStyle(style Style) {
	l.style = style
}

// SetApplication sets the application this layout belongs to and assigns indices.
func (l *Layout) SetApplication(app *Application) {
	if l.app == app && app != nil {
		return
	}
	l.app = app

	isNowMainLayout := app != nil && app.layout == l

	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			pane := l.panes[i].Pane
			pane.SetApplication(app) // Propagate app reference

			if isNowMainLayout {
				pane.SetIndex(i + 1)
			} else {
				if pane.GetIndex() != 0 {
					pane.SetIndex(0) // Reset index if not main layout
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
func (l *Layout) AddPane(pane *Pane, size Size) int {
	if pane == nil {
		return -1
	}
	if size.FixedSize <= 0 && size.Proportion <= 0 {
		size.Proportion = 1 // Default size
	}

	index := -1
	for i := range l.panes {
		if !l.panes[i].Active {
			index = i
			break
		}
	}
	if index == -1 {
		return -1
	}

	l.panes[index] = PaneInfo{Pane: pane, Size: size, Active: true}

	if l.app != nil {
		pane.SetApplication(l.app)
		if l.app.layout == l {
			pane.SetIndex(index + 1)
		} else {
			pane.SetIndex(0)
		}
	} else {
		pane.SetIndex(0)
	}

	l.activeCount++
	l.calculateLayout()
	return index
}

// RemovePane removes a pane by its array index (0-9).
func (l *Layout) RemovePane(index int) {
	if index < 0 || index >= 10 || !l.panes[index].Active {
		return
	}
	if l.panes[index].Pane != nil {
		l.panes[index].Pane.SetIndex(0)
	}
	l.panes[index] = PaneInfo{} // Zero out
	l.activeCount--
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

// SetMainAxisAlignment sets alignment along the main axis.
func (l *Layout) SetMainAxisAlignment(align Alignment) {
	if l.mainAxisAlign != align {
		l.mainAxisAlign = align
		l.calculateLayout()
	}
}

// SetCrossAxisAlignment sets alignment along the cross axis.
func (l *Layout) SetCrossAxisAlignment(align Alignment) {
	if l.crossAxisAlign != align {
		l.crossAxisAlign = align
		l.calculateLayout()
	}
}

// calculateLayout determines pane positions, handling insufficient space.
func (l *Layout) calculateLayout() {
	if l.activeCount == 0 || l.rect.Width <= 0 || l.rect.Height <= 0 {
		return
	}

	// --- 1. Determine Axis Sizes and Available Space ---
	mainAxisSize := 0
	crossAxisSize := 0
	isVertical := l.orientation == Vertical
	if isVertical {
		mainAxisSize = l.rect.Height
		crossAxisSize = l.rect.Width
	} else {
		mainAxisSize = l.rect.Width
		crossAxisSize = l.rect.Height
	}

	totalGapSize := 0
	if l.activeCount > 1 {
		totalGapSize = l.gap * (l.activeCount - 1)
	}

	totalAvailablePaneSpace := mainAxisSize - totalGapSize
	if totalAvailablePaneSpace < 0 {
		totalAvailablePaneSpace = 0
	}

	// --- 2. Identify Pane Types and Calculate Requested Sizes ---
	totalFixedRequested := 0
	totalProportionSum := 0.0
	fixedPaneIndices := []int{}
	proportionalPaneIndices := []int{}
	activePaneIndicesInOrder := []int{}

	for i := range l.panes {
		if !l.panes[i].Active || l.panes[i].Pane == nil {
			continue
		}
		activePaneIndicesInOrder = append(activePaneIndicesInOrder, i)
		size := l.panes[i].Size
		if size.FixedSize > 0 {
			totalFixedRequested += size.FixedSize
			fixedPaneIndices = append(fixedPaneIndices, i)
		} else if size.Proportion > 0 {
			totalProportionSum += float64(size.Proportion)
			proportionalPaneIndices = append(proportionalPaneIndices, i)
		}
	}

	// --- 3. Calculate Final Sizes, Handling Insufficient Space ---
	paneSizes := make(map[int]int) // Final main size for each pane index
	spaceLeftForProportionals := 0

	if totalFixedRequested <= totalAvailablePaneSpace {
		// Case A: Fixed panes fit perfectly (or have space left over)
		for _, idx := range fixedPaneIndices {
			paneSizes[idx] = l.panes[idx].Size.FixedSize // Assign requested fixed size
		}
		spaceLeftForProportionals = totalAvailablePaneSpace - totalFixedRequested
	} else {
		// Case B: Fixed panes DO NOT fit. Distribute available space proportionally among them.
		allocatedFixedSpace := 0
		remainder := 0
		if totalFixedRequested > 0 { // Avoid division by zero if only proportional panes requested > 0 space
			// Calculate and distribute base allocation
			for _, idx := range fixedPaneIndices {
				proportion := float64(l.panes[idx].Size.FixedSize) / float64(totalFixedRequested)
				alloc := int(float64(totalAvailablePaneSpace) * proportion) // Floor
				paneSizes[idx] = alloc
				allocatedFixedSpace += alloc
			}
			// Distribute remainder from flooring
			remainder = totalAvailablePaneSpace - allocatedFixedSpace
			if remainder > 0 && len(fixedPaneIndices) > 0 {
				pixelsPerPane := remainder / len(fixedPaneIndices)
				extraPixels := remainder % len(fixedPaneIndices)
				for i, idx := range fixedPaneIndices {
					paneSizes[idx] += pixelsPerPane
					if i < extraPixels {
						paneSizes[idx]++
					}
				}
			}
		}
		// Proportional panes get zero space in this case
		spaceLeftForProportionals = 0
	}

	// Calculate proportional sizes based on remaining space (if any)
	if totalProportionSum > 0 && spaceLeftForProportionals > 0 {
		allocatedProportionalSpace := 0
		remainder := 0
		// Calculate base proportional sizes
		for _, idx := range proportionalPaneIndices {
			proportion := float64(l.panes[idx].Size.Proportion) / totalProportionSum
			size := int(float64(spaceLeftForProportionals) * proportion) // Floor
			paneSizes[idx] = size
			allocatedProportionalSpace += size
		}
		// Distribute remainder
		remainder = spaceLeftForProportionals - allocatedProportionalSpace
		if remainder > 0 && len(proportionalPaneIndices) > 0 {
			pixelsPerPane := remainder / len(proportionalPaneIndices)
			extraPixels := remainder % len(proportionalPaneIndices)
			for i, idx := range proportionalPaneIndices {
				paneSizes[idx] += pixelsPerPane
				if i < extraPixels {
					paneSizes[idx]++
				}
			}
		}
	} else {
		// No space or no proportional panes, ensure they get size 0
		for _, idx := range proportionalPaneIndices {
			paneSizes[idx] = 0
		}
	}

	// --- 4. Set Final Rects ---
	currentPos := 0
	for _, paneArrIndex := range activePaneIndicesInOrder {
		paneInfo := l.panes[paneArrIndex]
		pane := paneInfo.Pane
		paneMainSize := paneSizes[paneArrIndex]

		if paneMainSize < 0 {
			paneMainSize = 0
		} // Should not happen, but safety first

		paneCrossSize := 0
		switch l.crossAxisAlign {
		case AlignStretch:
			fallthrough
		default:
			paneCrossSize = crossAxisSize
		}
		if paneCrossSize < 0 {
			paneCrossSize = 0
		}

		var paneX, paneY, paneWidth, paneHeight int
		if isVertical {
			paneX = l.rect.X
			paneY = l.rect.Y + currentPos
			paneWidth = paneCrossSize
			paneHeight = paneMainSize
		} else {
			paneX = l.rect.X + currentPos
			paneY = l.rect.Y
			paneWidth = paneMainSize
			paneHeight = paneCrossSize
		}

		// Clamp final dimensions to layout bounds
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

		pane.SetRect(paneX, paneY, paneWidth, paneHeight)
		currentPos += paneMainSize + l.gap // Advance position for the next pane
	}
}

// Draw draws the layout background and all active panes.
func (l *Layout) Draw(screen tcell.Screen) {
	// *** FIX for Border Artifacts: Clear layout's own area first ***
	Fill(screen, l.rect.X, l.rect.Y, l.rect.Width, l.rect.Height, ' ', l.style)

	// Create a list of panes to draw
	panesToDraw := make([]*Pane, 0, l.activeCount)
	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			panesToDraw = append(panesToDraw, l.panes[i].Pane)
		}
	}

	// Determine focus once
	focusedComp := l.app.GetFocusedComponent()

	// Draw each pane
	for _, pane := range panesToDraw {
		isChildFocused := false
		if focusedComp != nil {
			// Check if pane or its descendants contain the focus
			if childComp, ok := pane.child.(Component); ok && childComp == focusedComp {
				isChildFocused = true
			} else if childLayout, ok := pane.child.(*Layout); ok {
				isChildFocused = childLayout.ContainsFocus(focusedComp)
			}
		}
		pane.Draw(screen, isChildFocused) // Pane draws itself and its child
	}
}

// ContainsFocus checks recursively if this layout or any children contain the focused component.
func (l *Layout) ContainsFocus(focused Component) bool {
	if focused == nil {
		return false
	}
	for i := range l.panes {
		if !l.panes[i].Active || l.panes[i].Pane == nil {
			continue
		}
		pane := l.panes[i].Pane
		if childComp, ok := pane.child.(Component); ok && childComp == focused {
			return true
		}
		if childLayout, ok := pane.child.(*Layout); ok {
			if childLayout.ContainsFocus(focused) {
				return true
			}
		}
	}
	return false
}

// GetPaneByIndex returns the pane with the given user-facing index (1-10).
func (l *Layout) GetPaneByIndex(userIndex int) *Pane {
	if userIndex < 1 || userIndex > 10 {
		return nil
	}
	index := userIndex - 1
	if index >= len(l.panes) || !l.panes[index].Active || l.panes[index].Pane == nil {
		return nil
	}
	return l.panes[index].Pane
}

// GetAllFocusableComponents returns all focusable components recursively.
func (l *Layout) GetAllFocusableComponents() []Component {
	var focusables []Component
	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			focusables = append(focusables, l.panes[i].Pane.GetFocusableComponents()...)
		}
	}
	return focusables
}

// HasDirtyComponents checks if any pane or component needs redrawing.
func (l *Layout) HasDirtyComponents() bool {
	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			if l.panes[i].Pane.IsDirty() {
				return true
			}
		}
	}
	return false
}

// ClearAllDirtyFlags clears dirty flags recursively.
func (l *Layout) ClearAllDirtyFlags() {
	for i := range l.panes {
		if l.panes[i].Active && l.panes[i].Pane != nil {
			l.panes[i].Pane.ClearDirtyFlags()
		}
	}
}