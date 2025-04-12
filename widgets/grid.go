// widgets/grid.go
package widgets

import (
	"fmt"
	"sync"

	"github.com/LixenWraith/tinytui"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// SelectionMode defines whether a Grid allows single or multiple selections
type SelectionMode int

const (
	SingleSelect SelectionMode = iota // Only one item can be selected/interacted at a time
	MultiSelect                       // Multiple items can be selected/interacted
)

// Grid displays a 2D grid of text items, allowing navigation and selection.
type Grid struct {
	tinytui.BaseWidget
	mu                     sync.RWMutex
	cells                  [][]string // The string content for each cell [row][col]
	numRows                int
	numCols                int
	selectedRow            int                             // Index of the currently selected row
	selectedCol            int                             // Index of the currently selected column
	topRow                 int                             // Index of the row displayed at the top
	leftCol                int                             // Index of the column displayed at the left
	cellWidth              int                             // Fixed width for each cell (0 for auto - not implemented yet)
	cellHeight             int                             // Fixed height for each cell (usually 1)
	padding                int                             // Padding around cell content
	style                  tinytui.Style                   // Normal style
	selectedStyle          tinytui.Style                   // Selected, not focused
	interactedStyle        tinytui.Style                   // Interacted, not focused
	focusedStyle           tinytui.Style                   // Focused normal style
	focusedSelectedStyle   tinytui.Style                   // Focused and selected
	focusedInteractedStyle tinytui.Style                   // Focused and interacted
	indicatorChar          rune                            // Character used as selection indicator
	showIndicator          bool                            // Whether to show the indicator
	onChange               func(row, col int, item string) // Callback when selection changes
	onSelect               func(row, col int, item string) // Callback when item is selected (Space)
	interactedCells        map[string]bool                 // Track interacted cells using "row:col" as key
	selectionMode          SelectionMode                   // Single or multi selection mode
}

// NewGrid creates a new, empty Grid widget.
// Default cell height is 1. Cell width needs to be set.
func NewGrid() *Grid {
	g := &Grid{
		cells:                  [][]string{},
		selectedRow:            -1, // No selection initially
		selectedCol:            -1,
		topRow:                 0,
		leftCol:                0,
		cellHeight:             tinytui.DefaultCellHeight(),
		cellWidth:              tinytui.DefaultCellWidth(),
		padding:                tinytui.DefaultPadding(), // Use theme's default padding
		style:                  tinytui.DefaultGridStyle(),
		selectedStyle:          tinytui.DefaultGridStyle().Dim(true).Underline(true),
		interactedStyle:        tinytui.DefaultGridStyle().Bold(true),
		focusedStyle:           tinytui.DefaultGridStyle(),
		focusedSelectedStyle:   tinytui.DefaultGridSelectedStyle(),
		focusedInteractedStyle: tinytui.DefaultGridSelectedStyle().Bold(true),
		indicatorChar:          '>',
		showIndicator:          true,
		interactedCells:        make(map[string]bool),
		selectionMode:          SingleSelect, // Default to single selection
	}
	g.SetVisible(true) // Explicitly set visibility
	return g
}

// SetPadding sets the padding around cell content
func (g *Grid) SetPadding(padding int) *Grid {
	g.mu.Lock()
	g.padding = padding
	g.mu.Unlock()
	if app := g.App(); app != nil {
		app.QueueRedraw()
	}
	return g
}

// SetIndicator sets the indicator character and whether to show it
func (g *Grid) SetIndicator(char rune, show bool) *Grid {
	g.mu.Lock()
	g.indicatorChar = char
	g.showIndicator = show
	g.mu.Unlock()
	if app := g.App(); app != nil {
		app.QueueRedraw()
	}
	return g
}

// SetSelectionMode sets whether the grid allows single or multiple selections
func (g *Grid) SetSelectionMode(mode SelectionMode) *Grid {
	g.mu.Lock()
	g.selectionMode = mode
	g.mu.Unlock()
	return g
}

// SetCells replaces the grid content. Input is a 2D slice [row][col].
// Resets selection and scroll position. Assumes a rectangular grid.
func (g *Grid) SetCells(cells [][]string) *Grid {
	g.mu.Lock()
	g.cells = cells
	g.numRows = len(cells)
	g.numCols = 0
	if g.numRows > 0 {
		g.numCols = len(cells[0]) // Assume rectangular
	}

	g.topRow = 0
	g.leftCol = 0
	if g.numRows > 0 && g.numCols > 0 {
		g.selectedRow = 0
		g.selectedCol = 0
	} else {
		g.selectedRow = -1
		g.selectedCol = -1
	}
	g.clampIndices()
	g.mu.Unlock()

	g.triggerOnChange() // Trigger change after initial selection is set

	if app := g.App(); app != nil {
		app.QueueRedraw()
	}
	return g
}

// SetCellSize sets the fixed width and height for each cell.
// Height is typically 1 for simple text grids. Width determines spacing.
func (g *Grid) SetCellSize(width, height int) *Grid {
	// Use built-in min function (Go 1.21+)
	width = max(1, width)
	height = max(1, height)

	g.mu.Lock()
	g.cellWidth = width
	g.cellHeight = height
	g.clampIndices() // Re-clamp needed as viewport size relative to cells changes
	g.mu.Unlock()
	if app := g.App(); app != nil {
		app.QueueRedraw()
	}
	return g
}

// SetStyle sets the style for non-selected cells.
func (g *Grid) SetStyle(style tinytui.Style) *Grid {
	g.mu.Lock()
	g.style = style
	g.mu.Unlock()
	if app := g.App(); app != nil {
		app.QueueRedraw()
	}
	return g
}

func (g *Grid) SetSelectedStyle(style tinytui.Style) *Grid {
	g.mu.Lock()
	g.selectedStyle = style
	g.mu.Unlock()
	if app := g.App(); app != nil {
		app.QueueRedraw()
	}
	return g
}

func (g *Grid) SetInteractedStyle(style tinytui.Style) *Grid {
	g.mu.Lock()
	g.interactedStyle = style
	g.mu.Unlock()
	if app := g.App(); app != nil {
		app.QueueRedraw()
	}
	return g
}

func (g *Grid) SetFocusedStyle(style tinytui.Style) *Grid {
	g.mu.Lock()
	g.focusedStyle = style
	g.mu.Unlock()
	if app := g.App(); app != nil {
		app.QueueRedraw()
	}
	return g
}

func (g *Grid) SetFocusedSelectedStyle(style tinytui.Style) *Grid {
	g.mu.Lock()
	g.focusedSelectedStyle = style
	g.mu.Unlock()
	if app := g.App(); app != nil {
		app.QueueRedraw()
	}
	return g
}

func (g *Grid) SetFocusedInteractedStyle(style tinytui.Style) *Grid {
	g.mu.Lock()
	g.focusedInteractedStyle = style
	g.mu.Unlock()
	if app := g.App(); app != nil {
		app.QueueRedraw()
	}
	return g
}

// ApplyTheme applies the provided theme to the Grid widget
func (g *Grid) ApplyTheme(theme tinytui.Theme) {
	g.SetStyle(theme.GridStyle())
	g.SetSelectedStyle(theme.GridSelectedStyle())
	g.SetInteractedStyle(theme.GridInteractedStyle())
	g.SetFocusedStyle(theme.GridFocusedStyle())
	g.SetFocusedSelectedStyle(theme.GridFocusedSelectedStyle())
	g.SetFocusedInteractedStyle(theme.GridFocusedInteractedStyle())
	g.SetPadding(theme.DefaultPadding())

	// Update the indicator color through the style
	g.SetIndicator('>', true) // Always use '>' as indicator
}

// SetOnChange sets the callback for when the selection changes via navigation.
func (g *Grid) SetOnChange(handler func(row, col int, item string)) *Grid {
	g.mu.Lock()
	g.onChange = handler
	g.mu.Unlock()
	return g
}

// SetOnSelect sets the callback for when an item is explicitly selected (e.g., Enter/Space).
func (g *Grid) SetOnSelect(handler func(row, col int, item string)) *Grid {
	g.mu.Lock()
	g.onSelect = handler
	g.mu.Unlock()
	return g
}

// SelectedIndex returns the row and column index of the selected cell.
// Returns (-1, -1) if nothing is selected or grid is empty.
func (g *Grid) SelectedIndex() (row, col int) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	// Return actual selected indices, even if grid is empty they'll be -1
	return g.selectedRow, g.selectedCol
}

// SelectedItem returns the string content of the selected cell.
// Returns "" if nothing is selected or grid is empty.
func (g *Grid) SelectedItem() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	r, c := g.selectedRow, g.selectedCol
	// Check bounds carefully
	if r >= 0 && r < g.numRows && c >= 0 && c < g.numCols {
		return g.cells[r][c]
	}
	return ""
}

// SetSelectedIndex programmatically sets the selected cell.
func (g *Grid) SetSelectedIndex(row, col int) *Grid {
	g.mu.Lock()
	oldRow, oldCol := g.selectedRow, g.selectedCol
	g.selectedRow = row
	g.selectedCol = col
	g.clampIndices() // Clamp and adjust scroll based on new selection
	// Check if selection actually changed *after* clamping
	changed := g.selectedRow != oldRow || g.selectedCol != oldCol
	g.mu.Unlock()

	if changed {
		g.triggerOnChange() // Trigger change if selection moved
		if app := g.App(); app != nil {
			app.QueueRedraw()
		}
	}
	return g
}

// IsCellInteracted checks if a specific cell is in the interacted state
func (g *Grid) IsCellInteracted(row, col int) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	cellKey := fmt.Sprintf("%d:%d", row, col)
	return g.interactedCells[cellKey]
}

// GetInteractedCells returns all cells that are in the interacted state
func (g *Grid) GetInteractedCells() []struct{ Row, Col int } {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var result []struct{ Row, Col int }

	// Extract row/col from the cellKey in interactedCells
	for cellKey := range g.interactedCells {
		var row, col int
		// Parse the "row:col" format
		if _, err := fmt.Sscanf(cellKey, "%d:%d", &row, &col); err == nil {
			result = append(result, struct{ Row, Col int }{Row: row, Col: col})
		}
	}

	return result
}

// ClearInteractions removes all interactions from the grid
func (g *Grid) ClearInteractions() *Grid {
	g.mu.Lock()
	g.interactedCells = make(map[string]bool)
	g.mu.Unlock()

	if app := g.App(); app != nil {
		app.QueueRedraw()
	}
	return g
}

// clampIndices ensures selection and scroll indices are valid.
// Must be called with g.mu held.
func (g *Grid) clampIndices() {
	// Clamp selection first
	if g.numRows <= 0 || g.numCols <= 0 {
		g.selectedRow, g.selectedCol = -1, -1
	} else {
		// Use built-in min/max (Go 1.21+)
		g.selectedRow = max(0, g.selectedRow)
		g.selectedRow = min(g.numRows-1, g.selectedRow)
		g.selectedCol = max(0, g.selectedCol)
		g.selectedCol = min(g.numCols-1, g.selectedCol)
	}

	// Adjust scroll based on selection and viewport
	_, _, width, height := g.GetRect() // Use BaseWidget's GetRect
	if width <= 0 || height <= 0 || g.cellWidth <= 0 || g.cellHeight <= 0 {
		// Cannot calculate viewport, ensure scroll is at least 0
		g.topRow = max(0, g.topRow)
		g.leftCol = max(0, g.leftCol)
		return
	}

	visibleRows := max(1, height/g.cellHeight)
	visibleCols := max(1, width/g.cellWidth)

	// Adjust scroll only if there's a valid selection
	if g.selectedRow != -1 { // Check if selection is valid
		// Vertical scroll adjustment
		if g.selectedRow < g.topRow {
			g.topRow = g.selectedRow
		} else if g.selectedRow >= g.topRow+visibleRows {
			g.topRow = g.selectedRow - visibleRows + 1
		}

		// Horizontal scroll adjustment
		if g.selectedCol < g.leftCol {
			g.leftCol = g.selectedCol
		} else if g.selectedCol >= g.leftCol+visibleCols {
			g.leftCol = g.selectedCol - visibleCols + 1
		}
	}

	// Clamp scroll indices based on grid size and viewport
	g.topRow = max(0, g.topRow)
	maxTopRow := max(0, g.numRows-visibleRows) // Ensure maxTopRow is not negative
	g.topRow = min(maxTopRow, g.topRow)

	g.leftCol = max(0, g.leftCol)
	maxLeftCol := max(0, g.numCols-visibleCols) // Ensure maxLeftCol is not negative
	g.leftCol = min(maxLeftCol, g.leftCol)
}

// triggerOnChange safely calls the onChange callback if selection is valid.
func (g *Grid) triggerOnChange() {
	g.mu.RLock()
	handler := g.onChange
	r, c := g.selectedRow, g.selectedCol
	item := ""
	isValidSelection := r >= 0 && r < g.numRows && c >= 0 && c < g.numCols
	if isValidSelection {
		item = g.cells[r][c]
	}
	g.mu.RUnlock()

	if handler != nil && isValidSelection { // Only call if selection is valid
		handler(r, c, item)
	}
}

// triggerOnSelect safely calls the onSelect callback if selection is valid.
func (g *Grid) triggerOnSelect() {
	g.mu.RLock()
	handler := g.onSelect
	r, c := g.selectedRow, g.selectedCol
	item := ""
	isValidSelection := r >= 0 && r < g.numRows && c >= 0 && c < g.numCols
	if isValidSelection {
		item = g.cells[r][c]
	}
	g.mu.RUnlock()

	if handler != nil && isValidSelection { // Only call if selection is valid
		handler(r, c, item)
	}
}

// Draw renders the visible portion of the grid.
// Updated for consistent state display and indicators
func (g *Grid) Draw(screen tcell.Screen) {
	g.BaseWidget.Draw(screen)

	x, y, width, height := g.GetRect()
	if width <= 0 || height <= 0 || g.cellWidth <= 0 || g.cellHeight <= 0 {
		return // Cannot draw
	}

	g.mu.RLock() // Use RLock for reading content/lines
	// Read all necessary state under lock
	selRow, selCol := g.selectedRow, g.selectedCol
	topRow, leftCol := g.topRow, g.leftCol
	cWidth, cHeight := g.cellWidth, g.cellHeight
	padding := g.padding
	isFocused := g.IsFocused()
	showIndicator := g.showIndicator // Now we use this to reserve space, not just for display
	indicatorChar := g.indicatorChar

	// Base style
	baseStyle := g.style
	if isFocused {
		baseStyle = g.focusedStyle
	}

	cells := g.cells
	rows, cols := g.numRows, g.numCols

	// Copy the interacted cells map to avoid holding lock during drawing
	interactedCells := make(map[string]bool)
	for k, v := range g.interactedCells {
		interactedCells[k] = v
	}
	g.mu.RUnlock()

	// Get indicator color from theme
	indicatorStyle := baseStyle
	if app := g.App(); app != nil {
		if theme := app.GetTheme(); theme != nil {
			indicatorStyle = indicatorStyle.Foreground(theme.IndicatorColor())
		}
	}

	// Extract base colors for background fills
	baseFg, baseBg, _, _ := baseStyle.Deconstruct()
	baseFillStyle := tinytui.DefaultStyle.Foreground(baseFg).Background(baseBg)

	// Fill the entire grid background with base style (without attributes)
	tinytui.Fill(screen, x, y, width, height, ' ', baseFillStyle)

	visibleRows := height / cHeight
	visibleCols := width / cWidth

	// Draw visible cells
	for rOffset := 0; rOffset < visibleRows; rOffset++ {
		for cOffset := 0; cOffset < visibleCols; cOffset++ {
			gridRow := topRow + rOffset
			gridCol := leftCol + cOffset

			// Check if this cell is actually within the grid bounds
			if gridRow < 0 || gridRow >= rows || gridCol < 0 || gridCol >= cols {
				continue // Skip drawing if outside grid data
			}

			cellX := x + cOffset*cWidth
			cellY := y + rOffset*cHeight

			// Calculate actual cell dimensions considering widget boundaries
			drawWidth := cWidth
			drawHeight := cHeight
			if cellX+drawWidth > x+width {
				drawWidth = x + width - cellX
			}
			if cellY+drawHeight > y+height {
				drawHeight = y + height - cellY
			}

			if drawWidth <= 0 || drawHeight <= 0 {
				continue // Skip cells completely outside drawable bounds
			}

			// Determine cell style based on focus, selection state
			cellStyle := baseStyle

			// Check if this is the currently selected cell and/or interacted
			isCurrentCell := (gridRow == selRow && gridCol == selCol)
			cellKey := fmt.Sprintf("%d:%d", gridRow, gridCol)
			isInteracted := interactedCells[cellKey]

			if isCurrentCell {
				if isFocused {
					// Focused and selected cell
					if isInteracted {
						cellStyle = g.focusedInteractedStyle
					} else {
						cellStyle = g.focusedSelectedStyle
					}
				} else {
					// Not focused but selected cell
					if isInteracted {
						cellStyle = g.interactedStyle
					} else {
						cellStyle = g.selectedStyle
					}
				}
			} else if isInteracted {
				// Not selected but interacted
				if isFocused {
					cellStyle = g.interactedStyle.Bold(true) // Add emphasis for focused window
				} else {
					cellStyle = g.interactedStyle
				}
			}

			// Extract just colors for background fill
			cellFg, cellBg, _, _ := cellStyle.Deconstruct()
			cellFillStyle := tinytui.DefaultStyle.Foreground(cellFg).Background(cellBg)

			// Clear cell background with colors only (no attributes)
			tinytui.Fill(screen, cellX, cellY, drawWidth, drawHeight, ' ', cellFillStyle)

			// Draw content with full style including attributes
			item := cells[gridRow][gridCol]

			// Always reserve space for indicator if enabled, draw only when on current cell
			if showIndicator {
				if isCurrentCell && isFocused {
					// Draw indicator for current cell when focused
					if cellX >= x && cellX < x+width {
						screen.SetContent(cellX, cellY, indicatorChar, nil, indicatorStyle.ToTcell())
					}
				} else {
					// Draw empty space for indicator position to maintain alignment
					if cellX >= x && cellX < x+width {
						screen.SetContent(cellX, cellY, ' ', nil, cellFillStyle.ToTcell())
					}
				}
				// Always adjust content position by indicator width
				cellX += 1
				drawWidth -= 1
			}

			// Add padding to content position
			contentX := cellX + padding
			effectiveWidth := drawWidth - (padding * 2)
			if effectiveWidth < 1 {
				effectiveWidth = 1
			}

			// Simple truncation for drawing within the cell
			displayText := runewidth.Truncate(item, effectiveWidth, "")

			// Draw only on the first line of the cell area for now
			if cellY >= y && cellY < y+height && contentX >= x && contentX < x+width {
				tinytui.DrawText(screen, contentX, cellY, cellStyle, displayText)
			}
		}
	}
}

// SetRect updates dimensions and clamps indices.
func (g *Grid) SetRect(x, y, width, height int) {
	g.mu.Lock()
	g.BaseWidget.SetRect(x, y, width, height)
	g.clampIndices() // Re-clamp based on new viewport size
	g.mu.Unlock()
	// No redraw needed here, usually called during redraw cycle
}

// Focusable indicates Grid can receive focus.
func (g *Grid) Focusable() bool {
	if !g.IsVisible() {
		return false
	}

	g.mu.RLock()
	hasContent := g.numRows > 0 && g.numCols > 0
	g.mu.RUnlock()
	// A grid should only be focusable if it's visible and actually has cells
	return g.IsVisible() && hasContent
}

// HandleEvent handles keyboard navigation (arrows, vim keys) and selection (Enter/Space).
// Updated for consistent key handling across widgets
func (g *Grid) HandleEvent(event tcell.Event) bool {
	// Allow BaseWidget to handle its own potential keybindings first
	if g.BaseWidget.HandleEvent(event) {
		return true
	}

	keyEvent, ok := event.(*tcell.EventKey)
	if !ok {
		return false // Not a key event
	}

	g.mu.Lock() // Lock for modifying selection/scroll state

	currentRow, currentCol := g.selectedRow, g.selectedCol
	rows, cols := g.numRows, g.numCols
	selectionMode := g.selectionMode

	// If grid is empty or has no selection, cannot handle navigation/selection
	if rows <= 0 || cols <= 0 || currentRow < 0 || currentCol < 0 {
		g.mu.Unlock()
		return false
	}

	needsRedraw := false
	indexChanged := false
	newRow, newCol := currentRow, currentCol

	switch keyEvent.Key() {
	// Arrow Keys
	case tcell.KeyUp:
		newRow--
		needsRedraw = true
	case tcell.KeyDown:
		newRow++
		needsRedraw = true
	case tcell.KeyLeft:
		newCol--
		needsRedraw = true
	case tcell.KeyRight:
		newCol++
		needsRedraw = true

	// Enter toggles interaction state for the current cell
	case tcell.KeyEnter:
		cellKey := fmt.Sprintf("%d:%d", currentRow, currentCol)
		isInteracted := g.interactedCells[cellKey]

		if selectionMode == SingleSelect {
			// For single select, clear all other interactions first
			g.interactedCells = make(map[string]bool)
		}

		// Toggle the current cell's interaction state
		if isInteracted {
			delete(g.interactedCells, cellKey)
		} else {
			g.interactedCells[cellKey] = true
		}

		g.mu.Unlock()
		g.triggerOnSelect() // Trigger selection callback
		if app := g.App(); app != nil {
			app.QueueRedraw()
		}
		return true // Enter consumed

	// Backspace cancels interaction on the current cell
	case tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyDelete:
		cellKey := fmt.Sprintf("%d:%d", currentRow, currentCol)
		if g.interactedCells[cellKey] {
			delete(g.interactedCells, cellKey)
			g.mu.Unlock()
			if app := g.App(); app != nil {
				app.QueueRedraw()
			}
			return true
		}
		g.mu.Unlock()
		return false

	// Vim Keys (h,j,k,l) and Space
	case tcell.KeyRune:
		switch keyEvent.Rune() {
		case 'k': // Up
			newRow--
			needsRedraw = true
		case 'j': // Down
			newRow++
			needsRedraw = true
		case 'h': // Left
			newCol--
			needsRedraw = true
		case 'l': // Right
			newCol++
			needsRedraw = true
		case ' ': // Space creates/toggles interaction
			cellKey := fmt.Sprintf("%d:%d", currentRow, currentCol)
			isInteracted := g.interactedCells[cellKey]

			if selectionMode == SingleSelect {
				// For single select, clear all other interactions first
				g.interactedCells = make(map[string]bool)
			}

			// Toggle the current cell's interaction state
			if isInteracted {
				delete(g.interactedCells, cellKey)
			} else {
				g.interactedCells[cellKey] = true
			}

			g.mu.Unlock()
			g.triggerOnSelect() // Trigger selection callback
			if app := g.App(); app != nil {
				app.QueueRedraw()
			}
			return true // Space consumed
		default:
			g.mu.Unlock()
			return false // Rune not handled
		}

	default:
		g.mu.Unlock()
		return false // Key not handled
	}

	// Apply navigation changes if any key was processed
	if needsRedraw {
		// Check if the calculated new selection is different
		if newRow != currentRow || newCol != currentCol {
			g.selectedRow = newRow
			g.selectedCol = newCol
			// Clamp indices also handles scroll adjustment
			g.clampIndices()
			// Check if selection *actually* changed after clamping
			indexChanged = (g.selectedRow != currentRow || g.selectedCol != currentCol)
		}
		// Unlock *after* state modification and clamping
		g.mu.Unlock()

		// Trigger callbacks and redraw outside the lock
		if indexChanged {
			g.triggerOnChange() // Selection moved
		}
		if app := g.App(); app != nil {
			app.QueueRedraw() // Request redraw to show new selection/scroll
		}
		return true // Navigation key consumed
	}

	// Should not be reached if needsRedraw was true, but unlock just in case
	g.mu.Unlock()
	return false
}