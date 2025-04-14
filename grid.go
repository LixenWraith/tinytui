// grid.go
package tinytui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// Grid displays a 2D grid of selectable cells.
type Grid struct {
	BaseComponent
	cells           [][]string
	selectedRow     int
	selectedCol     int
	interactedCells map[string]bool
	cellWidth       int
	cellHeight      int
	topRow          int // Scroll state
	leftCol         int // Scroll state
	padding         int // Padding within cells

	// Styles for different states
	style                  Style
	selectedStyle          Style
	interactedStyle        Style
	focusedStyle           Style
	focusedSelectedStyle   Style
	focusedInteractedStyle Style

	// Event handlers
	onChange      func(row, col int, item string)
	onSelect      func(row, col int, item string)
	selectionMode SelectionMode

	// Display options
	autoWidth     bool // Calculate width based on content
	showIndicator bool // Show selection indicator
	indicatorChar rune // Character used for selection indicator
}

// NewGrid creates a new grid component.
func NewGrid() *Grid {
	return &Grid{
		BaseComponent:          NewBaseComponent(),
		cells:                  [][]string{},
		selectedRow:            -1, // No selection initially
		selectedCol:            -1,
		interactedCells:        make(map[string]bool),
		cellWidth:              10,
		cellHeight:             1,
		padding:                1,
		autoWidth:              false,
		topRow:                 0,
		leftCol:                0,
		style:                  DefaultStyle,
		selectedStyle:          DefaultStyle.Dim(true).Underline(true),
		interactedStyle:        DefaultStyle.Bold(true),
		focusedStyle:           DefaultStyle,
		focusedSelectedStyle:   DefaultStyle.Reverse(true),
		focusedInteractedStyle: DefaultStyle.Reverse(true).Bold(true),
		selectionMode:          SingleSelect,
		showIndicator:          true,
		indicatorChar:          '>',
	}
}

// SetCells sets the grid content.
func (g *Grid) SetCells(cells [][]string) {
	// Save previous selection
	prevRow, prevCol := g.selectedRow, g.selectedCol
	hadSelection := prevRow >= 0 && prevCol >= 0

	// Set new cells
	g.cells = cells

	// Determine new dimensions
	numRows := len(cells)
	numCols := 0
	if numRows > 0 {
		numCols = len(cells[0])
	}

	// Reset scroll position
	g.topRow = 0
	g.leftCol = 0

	// Determine initial selection
	if numRows > 0 && numCols > 0 {
		if hadSelection && prevRow < numRows && prevCol < numCols {
			// Keep previous selection if valid
			g.selectedRow = prevRow
			g.selectedCol = prevCol
		} else {
			// Start with first cell selected
			g.selectedRow = 0
			g.selectedCol = 0
		}
	} else {
		// No content, clear selection
		g.selectedRow = -1
		g.selectedCol = -1
	}

	// Mark as needing redraw
	g.MarkDirty()

	// Check if selection changed
	selectionChanged := (g.selectedRow != prevRow || g.selectedCol != prevCol)

	// Save values for callback outside lock
	onChange := g.onChange
	row, col := g.selectedRow, g.selectedCol
	var item string
	if row >= 0 && row < numRows && col >= 0 && col < numCols {
		item = g.cells[row][col]
	}

	// Trigger change event if selection changed
	if selectionChanged && onChange != nil && row >= 0 && col >= 0 {
		onChange(row, col, item)
	}
}

// SetContent is an alias for SetCells to implement TextUpdater interface.
// It expects a string with newlines for row separation and tabs for column separation.
func (g *Grid) SetContent(content string) {
	// This is a simplistic implementation - in a real application,
	// you might want more sophisticated parsing of the string into cells
	rows := [][]string{}

	// Split into rows
	lines := splitString(content, '\n')
	for _, line := range lines {
		// Split row into columns
		cells := splitString(line, '\t')
		rows = append(rows, cells)
	}

	g.SetCells(rows)
}

// GetSelectedCell returns the currently selected cell's content.
func (g *Grid) GetSelectedCell() (row, col int, content string) {
	row, col = g.selectedRow, g.selectedCol
	content = ""

	if row >= 0 && row < len(g.cells) && col >= 0 && col < len(g.cells[row]) {
		content = g.cells[row][col]
	}

	return row, col, content
}

// SetCellSize sets the size of each cell.
func (g *Grid) SetCellSize(width, height int) {
	if width <= 0 {
		width = 1
	}
	if height <= 0 {
		height = 1
	}

	changed := g.cellWidth != width || g.cellHeight != height
	if changed {
		g.cellWidth = width
		g.cellHeight = height
		g.autoWidth = false // Disable auto width when setting explicit size
		g.MarkDirty()
	}
}

// SetAutoWidth enables or disables automatic width calculation.
func (g *Grid) SetAutoWidth(auto bool) {
	if g.autoWidth != auto {
		g.autoWidth = auto
		g.MarkDirty()
	}
}

// SetPadding sets the padding within cells.
func (g *Grid) SetPadding(padding int) {
	if padding < 0 {
		padding = 0
	}

	if g.padding != padding {
		g.padding = padding
		g.MarkDirty()
	}
}

// SetStyle sets the style for normal cells.
func (g *Grid) SetStyle(style Style) {
	g.style = style
	g.MarkDirty()
}

// SetSelectedStyle sets the style for selected cells.
func (g *Grid) SetSelectedStyle(style Style) {
	g.selectedStyle = style
	g.MarkDirty()
}

// SetInteractedStyle sets the style for interacted cells.
func (g *Grid) SetInteractedStyle(style Style) {
	g.interactedStyle = style
	g.MarkDirty()
}

// SetFocusedStyle sets the style for when the grid has focus.
func (g *Grid) SetFocusedStyle(style Style) {
	g.focusedStyle = style
	g.MarkDirty()
}

// SetFocusedSelectedStyle sets the style for focused and selected cells.
func (g *Grid) SetFocusedSelectedStyle(style Style) {
	g.focusedSelectedStyle = style
	g.MarkDirty()
}

// SetFocusedInteractedStyle sets the style for focused and interacted cells.
func (g *Grid) SetFocusedInteractedStyle(style Style) {
	g.focusedInteractedStyle = style
	g.MarkDirty()
}

// SetIndicator sets the selection indicator.
func (g *Grid) SetIndicator(char rune, show bool) {
	g.indicatorChar = char
	g.showIndicator = show
	g.MarkDirty()
}

// SetSelectionMode sets the selection mode.
func (g *Grid) SetSelectionMode(mode SelectionMode) {
	if g.selectionMode != mode {
		g.selectionMode = mode
		g.MarkDirty()
	}
}

// SetOnChange sets the handler for selection change events.
func (g *Grid) SetOnChange(handler func(row, col int, item string)) {
	g.onChange = handler
}

// SetOnSelect sets the handler for cell selection events (Enter/Space).
func (g *Grid) SetOnSelect(handler func(row, col int, item string)) {
	g.onSelect = handler
}

// Focusable returns whether the grid can receive focus.
func (g *Grid) Focusable() bool {
	return g.IsVisible() && len(g.cells) > 0 && len(g.cells[0]) > 0
}

// selectCell selects a cell at the specified coordinates.
// Returns true if the selection changed.
func (g *Grid) selectCell(row, col int) bool {
	// Validate coordinates
	if row < 0 || col < 0 || row >= len(g.cells) || (len(g.cells) > 0 && col >= len(g.cells[0])) {
		return false
	}

	// Check if selection changed
	changed := (g.selectedRow != row || g.selectedCol != col)
	if !changed {
		return false
	}

	// Update selection
	g.selectedRow = row
	g.selectedCol = col

	// Ensure selection is visible
	g.ensureSelectionVisible()

	// Mark as needing redraw
	g.MarkDirty()

	// Save values for callback outside lock
	onChange := g.onChange
	selectedRow, selectedCol := g.selectedRow, g.selectedCol
	var selectedItem string
	if selectedRow >= 0 && selectedRow < len(g.cells) &&
		selectedCol >= 0 && selectedCol < len(g.cells[selectedRow]) {
		selectedItem = g.cells[selectedRow][selectedCol]
	}

	// Trigger change event
	if onChange != nil {
		onChange(selectedRow, selectedCol, selectedItem)
	}

	return true
}

// ensureSelectionVisible scrolls the grid to make the selection visible.
// Must be called with lock held.
func (g *Grid) ensureSelectionVisible() {
	if g.selectedRow < 0 || g.selectedCol < 0 {
		return
	}

	// Get component dimensions
	_, _, width, height := g.GetRect()

	// Calculate visible cells
	visibleRows := height / g.cellHeight

	// Calculate effective cell width
	effectiveCellWidth := g.cellWidth
	if g.autoWidth {
		effectiveCellWidth = g.calculateCellWidth()
	}

	visibleCols := width / effectiveCellWidth

	// Adjust scroll if selection is outside visible area
	if g.selectedRow < g.topRow {
		g.topRow = g.selectedRow
	} else if g.selectedRow >= g.topRow+visibleRows {
		g.topRow = g.selectedRow - visibleRows + 1
	}

	if g.selectedCol < g.leftCol {
		g.leftCol = g.selectedCol
	} else if g.selectedCol >= g.leftCol+visibleCols {
		g.leftCol = g.selectedCol - visibleCols + 1
	}

	// Ensure scroll values are valid
	if g.topRow < 0 {
		g.topRow = 0
	}

	if g.leftCol < 0 {
		g.leftCol = 0
	}

	maxTopRow := len(g.cells) - visibleRows
	if maxTopRow < 0 {
		maxTopRow = 0
	}
	if g.topRow > maxTopRow {
		g.topRow = maxTopRow
	}

	maxLeftCol := 0
	if len(g.cells) > 0 {
		maxLeftCol = len(g.cells[0]) - visibleCols
	}
	if maxLeftCol < 0 {
		maxLeftCol = 0
	}
	if g.leftCol > maxLeftCol {
		g.leftCol = maxLeftCol
	}
}

// toggleCellInteraction toggles the interaction state of the selected cell.
func (g *Grid) toggleCellInteraction() {
	// Validate selection
	if g.selectedRow < 0 || g.selectedCol < 0 ||
		g.selectedRow >= len(g.cells) || g.selectedCol >= len(g.cells[0]) {
		return
	}

	// Generate cell key
	cellKey := fmt.Sprintf("%d:%d", g.selectedRow, g.selectedCol)

	// Toggle interaction state
	interacted := g.interactedCells[cellKey]

	if g.selectionMode == SingleSelect {
		// Clear all interactions
		g.interactedCells = make(map[string]bool)
	}

	if interacted {
		delete(g.interactedCells, cellKey)
	} else {
		g.interactedCells[cellKey] = true
	}

	// Mark as needing redraw
	g.MarkDirty()

	// Save values for callback outside lock
	onSelect := g.onSelect
	row, col := g.selectedRow, g.selectedCol
	var item string
	if row >= 0 && row < len(g.cells) && col >= 0 && col < len(g.cells[row]) {
		item = g.cells[row][col]
	}

	// Trigger select event
	if onSelect != nil {
		onSelect(row, col, item)
	}
}

// Draw draws the grid.
func (g *Grid) Draw(screen tcell.Screen) {
	// Check visibility
	if !g.IsVisible() {
		return
	}

	// Get component dimensions
	x, y, width, height := g.GetRect()
	if width <= 0 || height <= 0 {
		return
	}

	// Calculate effective cell width
	effectiveCellWidth := g.cellWidth
	if g.autoWidth {
		effectiveCellWidth = g.calculateCellWidth()
	}

	// Save state for drawing
	isFocused := g.IsFocused()
	showIndicator := g.showIndicator
	indicatorChar := g.indicatorChar
	padding := g.padding
	cells := g.cells
	selectedRow := g.selectedRow
	selectedCol := g.selectedCol
	topRow := g.topRow
	leftCol := g.leftCol
	cellHeight := g.cellHeight
	interactedCells := make(map[string]bool)
	for k, v := range g.interactedCells {
		interactedCells[k] = v
	}

	style := g.style
	selectedStyle := g.selectedStyle
	interactedStyle := g.interactedStyle
	focusedStyle := g.focusedStyle
	focusedSelectedStyle := g.focusedSelectedStyle
	focusedInteractedStyle := g.focusedInteractedStyle

	// Calculate visible cells
	visibleRows := height / cellHeight
	visibleCols := width / effectiveCellWidth

	// Fill background
	Fill(screen, x, y, width, height, ' ', style)

	// Draw visible cells
	for row := 0; row < visibleRows; row++ {
		gridRow := topRow + row

		// Skip if out of bounds
		if gridRow >= len(cells) {
			break
		}

		for col := 0; col < visibleCols; col++ {
			gridCol := leftCol + col

			// Skip if out of bounds
			if gridCol >= len(cells[gridRow]) {
				break
			}

			// Calculate cell position
			cellX := x + col*effectiveCellWidth
			cellY := y + row*cellHeight

			// Determine cell style based on selection and interaction state
			isSelected := gridRow == selectedRow && gridCol == selectedCol
			cellKey := fmt.Sprintf("%d:%d", gridRow, gridCol)
			isInteracted := interactedCells[cellKey]

			cellStyle := style
			if isFocused {
				if isSelected && isInteracted {
					cellStyle = focusedInteractedStyle
				} else if isSelected {
					cellStyle = focusedSelectedStyle
				} else if isInteracted {
					cellStyle = interactedStyle
				} else {
					cellStyle = focusedStyle
				}
			} else {
				if isSelected && isInteracted {
					cellStyle = interactedStyle
				} else if isSelected {
					cellStyle = selectedStyle
				} else if isInteracted {
					cellStyle = interactedStyle
				}
			}

			// Clear cell background
			Fill(screen, cellX, cellY, effectiveCellWidth, cellHeight, ' ', cellStyle)

			// Draw selection indicator if this is the selected cell
			// indicatorWidth := 0
			if showIndicator && isSelected && isFocused {
				screen.SetContent(cellX, cellY, indicatorChar, nil, cellStyle.ToTcell())
				// indicatorWidth = 1
			}

			// Draw cell content
			contentX := cellX + padding + 1 // 1 for indicator
			contentMaxWidth := effectiveCellWidth - padding*2 - 1

			if contentMaxWidth > 0 {
				content := cells[gridRow][gridCol]
				if len(content) > 0 {
					// Truncate content if necessary
					displayText := content
					if runewidth.StringWidth(content) > contentMaxWidth {
						displayText = runewidth.Truncate(content, contentMaxWidth, "")
					}

					DrawText(screen, contentX, cellY, cellStyle, displayText)
				}
			}
		}
	}
}

// calculateCellWidth determines the width needed for cells based on content.
// Must be called with lock held.
func (g *Grid) calculateCellWidth() int {
	// Start with padding and optional indicator
	minWidth := g.padding * 2
	if g.showIndicator {
		minWidth++
	}

	// Find the widest content
	maxContentWidth := 0
	for _, row := range g.cells {
		for _, cell := range row {
			width := runewidth.StringWidth(cell)
			if width > maxContentWidth {
				maxContentWidth = width
			}
		}
	}

	totalWidth := minWidth + maxContentWidth

	// Ensure a minimum width
	if totalWidth < 5 {
		return 5
	}

	return totalWidth
}

// HandleEvent handles keyboard input for grid navigation and selection.
func (g *Grid) HandleEvent(event tcell.Event) bool {
	// Only handle key events
	keyEvent, ok := event.(*tcell.EventKey)
	if !ok {
		return false
	}

	// Need valid grid with content
	hasContent := len(g.cells) > 0 && len(g.cells[0]) > 0

	if !hasContent {
		return false
	}

	switch keyEvent.Key() {
	case tcell.KeyUp:
		// Move selection up
		newRow := g.selectedRow - 1
		newCol := g.selectedCol
		return g.selectCell(newRow, newCol)

	case tcell.KeyDown:
		// Move selection down
		newRow := g.selectedRow + 1
		newCol := g.selectedCol
		return g.selectCell(newRow, newCol)

	case tcell.KeyLeft:
		// Move selection left
		newRow := g.selectedRow
		newCol := g.selectedCol - 1
		return g.selectCell(newRow, newCol)

	case tcell.KeyRight:
		// Move selection right
		newRow := g.selectedRow
		newCol := g.selectedCol + 1
		return g.selectCell(newRow, newCol)

	case tcell.KeyEnter:
		// Toggle interaction state of selected cell
		g.toggleCellInteraction()
		return true
	}

	// Check for vim-style navigation (h,j,k,l)
	if keyEvent.Key() == tcell.KeyRune {
		switch keyEvent.Rune() {
		case 'k': // Up
			newRow := g.selectedRow - 1
			newCol := g.selectedCol
			return g.selectCell(newRow, newCol)

		case 'j': // Down
			newRow := g.selectedRow + 1
			newCol := g.selectedCol
			return g.selectCell(newRow, newCol)

		case 'h': // Left
			newRow := g.selectedRow
			newCol := g.selectedCol - 1
			return g.selectCell(newRow, newCol)

		case 'l': // Right
			newRow := g.selectedRow
			newCol := g.selectedCol + 1
			return g.selectCell(newRow, newCol)
		}
	}

	return false
}

// IsCellInteracted checks if a cell is in the interacted state.
func (g *Grid) IsCellInteracted(row, col int) bool {
	cellKey := fmt.Sprintf("%d:%d", row, col)
	return g.interactedCells[cellKey]
}

// SetCellInteracted sets a cell's interaction state.
func (g *Grid) SetCellInteracted(row, col int, interacted bool) {
	// Validate coordinates
	if row < 0 || col < 0 || row >= len(g.cells) || col >= len(g.cells[0]) {
		return
	}

	cellKey := fmt.Sprintf("%d:%d", row, col)

	if g.selectionMode == SingleSelect && interacted {
		// Clear all other interactions
		g.interactedCells = make(map[string]bool)
	}

	// Set interaction state
	if interacted {
		g.interactedCells[cellKey] = true
	} else {
		delete(g.interactedCells, cellKey)
	}

	g.MarkDirty()
}

// GetInteractedCells returns all interacted cells.
func (g *Grid) GetInteractedCells() [][2]int {
	result := make([][2]int, 0, len(g.interactedCells))

	for key := range g.interactedCells {
		var row, col int
		fmt.Sscanf(key, "%d:%d", &row, &col)
		result = append(result, [2]int{row, col})
	}

	return result
}

// ClearInteractions clears all cell interactions.
func (g *Grid) ClearInteractions() {
	if len(g.interactedCells) > 0 {
		g.interactedCells = make(map[string]bool)
		g.MarkDirty()
	}
}

// splitString splits a string by a delimiter.
func splitString(s string, delim rune) []string {
	var result []string
	current := ""

	for _, r := range s {
		if r == delim {
			result = append(result, current)
			current = ""
		} else {
			current += string(r)
		}
	}

	// Add the last part
	result = append(result, current)

	return result
}