// widgets/grid.go
package widgets

import (
	"sync"

	"github.com/LixenWraith/tinytui"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
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
	style                  tinytui.Style                   // Normal style
	selectedStyle          tinytui.Style                   // Selected, not focused
	interactedStyle        tinytui.Style                   // Interacted, not focused
	focusedStyle           tinytui.Style                   // Focused normal style
	focusedSelectedStyle   tinytui.Style                   // Focused and selected
	focusedInteractedStyle tinytui.Style                   // Focused and interacted
	onChange               func(row, col int, item string) // Callback when selection changes
	onSelect               func(row, col int, item string) // Callback when item is selected (Space)
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
		style:                  tinytui.DefaultGridStyle(),
		selectedStyle:          tinytui.DefaultGridStyle().Dim(true).Underline(true),
		interactedStyle:        tinytui.DefaultGridStyle().Bold(true),
		focusedStyle:           tinytui.DefaultGridStyle(),
		focusedSelectedStyle:   tinytui.DefaultGridSelectedStyle(),
		focusedInteractedStyle: tinytui.DefaultGridSelectedStyle().Bold(true),
	}
	g.SetVisible(true) // Explicitly set visibility
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
	state := g.GetState()
	isFocused := g.IsFocused() // Check focus state for drawing

	// Base style
	baseStyle := g.style
	if isFocused {
		baseStyle = g.focusedStyle
	}

	cells := g.cells
	rows, cols := g.numRows, g.numCols
	g.mu.RUnlock()

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

			// Check if this is the currently selected cell
			isCurrentCell := (gridRow == selRow && gridCol == selCol)

			if isCurrentCell {
				if isFocused {
					// Focused and selected cell
					if state == tinytui.StateInteracted {
						cellStyle = g.focusedInteractedStyle
					} else {
						cellStyle = g.focusedSelectedStyle
					}
				} else {
					// Not focused but selected cell
					if state == tinytui.StateInteracted {
						cellStyle = g.interactedStyle
					} else {
						cellStyle = g.selectedStyle
					}
				}
			}

			// Extract just colors for background fill
			cellFg, cellBg, _, _ := cellStyle.Deconstruct()
			cellFillStyle := tinytui.DefaultStyle.Foreground(cellFg).Background(cellBg)

			// Clear cell background with colors only (no attributes)
			tinytui.Fill(screen, cellX, cellY, drawWidth, drawHeight, ' ', cellFillStyle)

			// Draw content with full style including attributes
			item := cells[gridRow][gridCol]
			// Simple truncation for drawing within the cell
			displayText := runewidth.Truncate(item, drawWidth, "") // Use runewidth for truncation

			// Draw only on the first line of the cell area for now
			if cellY >= y && cellY < y+height { // Ensure Y is within bounds
				tinytui.DrawText(screen, cellX, cellY, cellStyle, displayText)
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

	// Enter for Selection
	case tcell.KeyEnter:
		// Set state to interacted
		g.SetState(tinytui.StateInteracted)
		g.mu.Unlock()       // Unlock before calling callback
		g.triggerOnSelect() // Trigger the selection callback
		return true         // Enter consumed

	// Vim Keys (h,j,k,l)
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
		case ' ': // Space
			currentState := g.GetState()
			if currentState != tinytui.StateSelected {
				g.SetState(tinytui.StateSelected)
			} else {
				g.SetState(tinytui.StateNormal)
			}
			g.mu.Unlock()
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
			// When the selection changes, set state to selected
			g.SetState(tinytui.StateSelected)
			g.triggerOnChange() // Selection moved
		}
		if app := g.App(); app != nil {
			app.QueueRedraw() // Request redraw to show new selection/scroll
		}
		return true // Navigation key consumed
	}

	// Should not be reached if needsRedraw was false, but unlock just in case
	g.mu.Unlock()
	return false
}