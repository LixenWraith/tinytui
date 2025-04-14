// grid.go
package tinytui

import (
	"fmt"
	// NOTE: Removed strconv import as Sscanf is used instead
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// Grid displays a 2D grid of selectable and potentially interactive cells.
type Grid struct {
	BaseComponent
	cells           [][]string      // The grid data [row][col]
	selectedRow     int             // Index of the currently selected row
	selectedCol     int             // Index of the currently selected column
	interactedCells map[string]bool // Tracks interacted cells (key: "row:col")
	cellWidth       int             // Fixed width of each cell
	cellHeight      int             // Fixed height of each cell (usually 1)
	topRow          int             // Index of the top-most visible row (for scrolling)
	leftCol         int             // Index of the left-most visible column (for scrolling)
	padding         int             // Padding within cells (usually left/right)

	// Styles for different states (updated by ApplyTheme)
	style                  Style
	selectedStyle          Style
	interactedStyle        Style
	focusedStyle           Style
	focusedSelectedStyle   Style
	focusedInteractedStyle Style

	// Event handlers
	onChange func(row, col int, item string) // Called when selection changes
	onSelect func(row, col int, item string) // Called when Enter/Space is pressed on a cell

	// Configuration
	selectionMode  SelectionMode // Single or Multi selection
	autoWidth      bool          // Calculate width based on content?
	showIndicator  bool          // Show indicator on the selected cell?
	indicatorChar  rune          // Character used for selection indicator
	indicatorStyle Style         // Style for the indicator (derived from theme)
}

// NewGrid creates a new grid component, initializing styles from the current theme.
func NewGrid() *Grid {
	theme := GetTheme() // Get theme at creation time
	if theme == nil {
		// Fallback if themes haven't been initialized (shouldn't normally happen)
		theme = NewDefaultTheme()
	}
	g := &Grid{
		BaseComponent:   NewBaseComponent(),
		cells:           [][]string{},
		selectedRow:     -1, // No selection initially
		selectedCol:     -1,
		interactedCells: make(map[string]bool),
		cellWidth:       theme.DefaultCellWidth(),  // Use theme default
		cellHeight:      theme.DefaultCellHeight(), // Use theme default
		padding:         theme.DefaultPadding(),    // Use theme default
		autoWidth:       false,
		topRow:          0,
		leftCol:         0,
		selectionMode:   SingleSelect,
		showIndicator:   true,
		indicatorChar:   '>',
		// Styles will be set by ApplyTheme
	}
	// Apply the initial theme
	g.ApplyTheme(theme)
	return g
}

// ApplyTheme updates the grid's styles based on the provided theme.
// Implements ThemedComponent.
func (g *Grid) ApplyTheme(theme Theme) {
	if theme == nil {
		return
	} // Safety check

	g.style = theme.GridStyle()
	g.selectedStyle = theme.GridSelectedStyle()
	g.interactedStyle = theme.GridInteractedStyle()
	g.focusedStyle = theme.GridFocusedStyle()
	g.focusedSelectedStyle = theme.GridFocusedSelectedStyle()
	g.focusedInteractedStyle = theme.GridFocusedInteractedStyle()

	// Use theme's indicator color combined with the focused selected style for the indicator
	// This ensures the indicator is visible against the selected cell background
	g.indicatorStyle = theme.GridFocusedSelectedStyle().Foreground(theme.IndicatorColor())

	// Note: We don't automatically reset explicitly set dimensions/padding on theme change.
	// The user might have customized them after creation.

	g.MarkDirty() // Mark dirty to reflect potential style changes
}

// SetCells updates the grid's content. Resets scroll and potentially selection.
// Ensures the resulting grid data is rectangular by padding shorter rows.
func (g *Grid) SetCells(cells [][]string) {
	prevRow, prevCol := g.selectedRow, g.selectedCol
	hadSelection := prevRow >= 0 && prevCol >= 0

	// Ensure grid is rectangular (pad shorter rows if necessary) for predictability
	maxCols := 0
	if len(cells) > 0 {
		for _, row := range cells {
			if len(row) > maxCols {
				maxCols = len(row)
			}
		}
		// Make a copy and pad
		paddedCells := make([][]string, len(cells))
		for i, row := range cells {
			paddedCells[i] = make([]string, maxCols)
			copy(paddedCells[i], row)
			// The rest are already empty strings from make()
		}
		g.cells = paddedCells
	} else {
		g.cells = cells // Empty grid
	}

	numRows := len(g.cells)
	numCols := maxCols // Use the calculated maxCols

	// Reset scroll position
	g.topRow = 0
	g.leftCol = 0

	// Reset selection or try to keep it
	if numRows > 0 && numCols > 0 {
		if hadSelection && prevRow < numRows && prevCol < numCols {
			// Keep previous selection if still valid
			g.selectedRow = prevRow
			g.selectedCol = prevCol
		} else {
			// Select the first cell by default
			g.selectedRow = 0
			g.selectedCol = 0
		}
	} else {
		// No content, clear selection
		g.selectedRow = -1
		g.selectedCol = -1
	}

	g.ClearInteractions()      // Clear interaction state when content changes
	g.ensureSelectionVisible() // Ensure the new selection is visible
	g.MarkDirty()

	// Check if selection actually changed and trigger onChange
	newRow, newCol := g.selectedRow, g.selectedCol
	selectionChanged := (newRow != prevRow || newCol != prevCol)
	if selectionChanged && g.onChange != nil && newRow >= 0 && newCol >= 0 {
		g.onChange(newRow, newCol, g.cells[newRow][newCol])
	} else if !hadSelection && newRow >= 0 && newCol >= 0 && g.onChange != nil {
		// Trigger onChange if selection was initially invalid but is now valid
		g.onChange(newRow, newCol, g.cells[newRow][newCol])
	}
}

// SetContent implements TextUpdater by parsing a string into cells.
// Expects newline ('\n') for row separation and tab ('\t') for column separation.
func (g *Grid) SetContent(content string) {
	rowsData := [][]string{}
	if len(content) == 0 {
		g.SetCells(rowsData)
		return
	}

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		// Handle potential trailing newline creating an empty string element, but only if it's the very last line
		if line == "" && i == len(lines)-1 && len(lines) > 1 {
			continue // Skip trailing empty line from final newline
		}
		cells := strings.Split(line, "\t")
		rowsData = append(rowsData, cells)
	}

	g.SetCells(rowsData) // SetCells handles padding to rectangular
}

// GetSelectedCell returns the currently selected cell's row, column, and content.
// Returns -1, -1, "" if nothing is selected or grid is empty.
func (g *Grid) GetSelectedCell() (row, col int, content string) {
	row, col = g.selectedRow, g.selectedCol
	// Validate selection against current grid dimensions
	if row >= 0 && row < len(g.cells) && col >= 0 && col < len(g.cells[row]) {
		content = g.cells[row][col]
	} else {
		row, col = -1, -1 // Ensure invalid selection returns -1
		content = ""
	}
	return
}

// SetCellSize sets the fixed size (width, height) of each cell.
// Disables autoWidth if width is set.
func (g *Grid) SetCellSize(width, height int) {
	if width <= 0 {
		width = 1
	}
	if height <= 0 {
		height = 1
	}

	changed := g.cellWidth != width || g.cellHeight != height || g.autoWidth // Change if size differs OR if switching from autoWidth
	if changed {
		g.cellWidth = width
		g.cellHeight = height
		g.autoWidth = false // Explicit size overrides auto width
		g.MarkDirty()
	}
}

// SetAutoWidth enables or disables automatic cell width calculation based on content.
func (g *Grid) SetAutoWidth(auto bool) {
	if g.autoWidth != auto {
		g.autoWidth = auto
		// Width will be recalculated on next Draw if auto is true
		g.MarkDirty()
	}
}

// SetPadding sets the internal padding (space on left/right) within cells.
func (g *Grid) SetPadding(padding int) {
	if padding < 0 {
		padding = 0
	}
	if g.padding != padding {
		g.padding = padding
		g.MarkDirty()
	}
}

// SetIndicator configures the selection indicator character and visibility.
func (g *Grid) SetIndicator(char rune, show bool) {
	changed := g.indicatorChar != char || g.showIndicator != show
	if changed {
		g.indicatorChar = char
		g.showIndicator = show
		g.MarkDirty()
	}
}

// SetSelectionMode sets whether single or multiple cells can be interacted with.
func (g *Grid) SetSelectionMode(mode SelectionMode) {
	if g.selectionMode != mode {
		g.selectionMode = mode
		if mode == SingleSelect && len(g.interactedCells) > 1 {
			// If switching to single select, clear all interactions.
			g.ClearInteractions()
		}
		g.MarkDirty() // Style might change if interactions are cleared
	}
}

// SetOnChange sets the callback function triggered when the selected cell changes.
func (g *Grid) SetOnChange(handler func(row, col int, item string)) {
	g.onChange = handler
}

// SetOnSelect sets the callback function triggered when a cell is "activated" (e.g., Enter/Space).
func (g *Grid) SetOnSelect(handler func(row, col int, item string)) {
	g.onSelect = handler
}

// Focusable returns true if the grid is visible and contains selectable cells.
func (g *Grid) Focusable() bool {
	// Check if visible and has at least one cell
	return g.IsVisible() && len(g.cells) > 0 && len(g.cells[0]) > 0 // Assumes rectangular
}

// selectCell moves the selection to the specified row and col.
// Returns true if the selection actually changed. Handles initial selection.
func (g *Grid) selectCell(row, col int) bool {
	numRows := len(g.cells)
	if numRows == 0 {
		return false
	} // Cannot select in empty grid
	numCols := len(g.cells[0]) // Assumes rectangular grid
	if numCols == 0 {
		return false
	} // Cannot select if no columns

	initialSelection := false
	// Handle initial selection if none exists
	if g.selectedRow < 0 || g.selectedCol < 0 {
		initialSelection = true
		// Start selection clamped to valid range
		if row < 0 {
			row = 0
		}
		if row >= numRows {
			row = numRows - 1
		}
		if col < 0 {
			col = 0
		}
		if col >= numCols {
			col = numCols - 1
		}
	} else {
		// Clamp target coordinates to valid range for existing selection movement
		if row < 0 {
			row = 0
		}
		if row >= numRows {
			row = numRows - 1
		}
		if col < 0 {
			col = 0
		}
		if col >= numCols {
			col = numCols - 1
		}
	}

	// Check if selection actually changed from the previous valid state
	prevRow, prevCol := g.selectedRow, g.selectedCol
	if !initialSelection && prevRow == row && prevCol == col {
		return false // No change needed
	}

	g.selectedRow = row
	g.selectedCol = col

	// Ensure the new selection is visible
	g.ensureSelectionVisible()
	g.MarkDirty()

	// Trigger change event if selection coords actually changed OR if it was the initial selection
	if g.onChange != nil {
		if initialSelection || prevRow != row || prevCol != col {
			g.onChange(row, col, g.cells[row][col])
		}
	}

	return true // Selection was made or changed
}

// ensureSelectionVisible adjusts the scroll offsets (topRow, leftCol)
// so that the currently selected cell is within the visible area.
func (g *Grid) ensureSelectionVisible() {
	if g.selectedRow < 0 || g.selectedCol < 0 {
		return
	} // No selection

	_, _, width, height := g.GetRect()
	if width <= 0 || height <= 0 {
		return
	} // Component not sized

	// Calculate effective cell dimensions for visibility check
	effectiveCellWidth := g.cellWidth
	if g.autoWidth {
		effectiveCellWidth = g.calculateCellWidth()
		if effectiveCellWidth <= 0 {
			effectiveCellWidth = 1
		} // Avoid division by zero if calculation fails
	}
	effectiveCellHeight := g.cellHeight
	if effectiveCellHeight <= 0 {
		effectiveCellHeight = 1
	} // Avoid division by zero

	// Calculate number of visible rows/cols based on component size and cell size
	visibleRows := height / effectiveCellHeight
	visibleCols := width / effectiveCellWidth
	if visibleRows <= 0 {
		visibleRows = 1
	} // Ensure at least one row is considered visible
	if visibleCols <= 0 {
		visibleCols = 1
	} // Ensure at least one col is considered visible

	// Adjust vertical scroll (topRow)
	if g.selectedRow < g.topRow {
		g.topRow = g.selectedRow // Scroll up: Make selected row the top row
	} else if g.selectedRow >= g.topRow+visibleRows {
		g.topRow = g.selectedRow - visibleRows + 1 // Scroll down: Make selected row the bottom row
	}

	// Adjust horizontal scroll (leftCol)
	if g.selectedCol < g.leftCol {
		g.leftCol = g.selectedCol // Scroll left: Make selected col the left col
	} else if g.selectedCol >= g.leftCol+visibleCols {
		g.leftCol = g.selectedCol - visibleCols + 1 // Scroll right: Make selected col the right col
	}

	// --- Clamp scroll values to valid ranges ---
	numRows := len(g.cells)
	numCols := 0
	if numRows > 0 {
		numCols = len(g.cells[0])
	} // Assumes rectangular

	// Clamp topRow
	if g.topRow < 0 {
		g.topRow = 0
	}
	maxTopRow := numRows - visibleRows // Max topRow is numRows minus the number that fit
	if maxTopRow < 0 {
		maxTopRow = 0
	} // Handle case where visibleRows > numRows
	if g.topRow > maxTopRow {
		g.topRow = maxTopRow
	}

	// Clamp leftCol
	if g.leftCol < 0 {
		g.leftCol = 0
	}
	maxLeftCol := numCols - visibleCols // Max leftCol is numCols minus the number that fit
	if maxLeftCol < 0 {
		maxLeftCol = 0
	} // Handle case where visibleCols > numCols
	if g.leftCol > maxLeftCol {
		g.leftCol = maxLeftCol
	}
	// No need to MarkDirty here, as this is called before drawing or after selection change which already marks dirty.
}

// toggleCellInteraction toggles the interaction state of the currently selected cell
// based on the SelectionMode and triggers the onSelect callback.
func (g *Grid) toggleCellInteraction() {
	// Ensure a valid cell is selected
	row, col := g.selectedRow, g.selectedCol
	if row < 0 || row >= len(g.cells) || col < 0 || col >= len(g.cells[row]) {
		return // Cannot interact with invalid selection
	}

	cellKey := fmt.Sprintf("%d:%d", row, col)
	currentlyInteracted := g.interactedCells[cellKey]
	stateChanged := false

	if g.selectionMode == SingleSelect {
		// If single select, behavior depends on current state
		if currentlyInteracted {
			// If it was interacted, toggle means turn it off
			delete(g.interactedCells, cellKey)
			stateChanged = true
		} else {
			// If it wasn't interacted, turn it on and clear others
			if len(g.interactedCells) > 0 {
				g.interactedCells = make(map[string]bool) // Clear previous
			}
			g.interactedCells[cellKey] = true
			stateChanged = true
		}
	} else { // MultiSelect
		// Just toggle the state of the current cell
		if currentlyInteracted {
			delete(g.interactedCells, cellKey)
		} else {
			g.interactedCells[cellKey] = true
		}
		stateChanged = true // Always consider state changed in multi-select toggle
	}

	if stateChanged {
		g.MarkDirty() // Interaction state changed, need redraw
	}

	// Trigger the select event callback regardless of state change (activation event)
	if g.onSelect != nil {
		g.onSelect(row, col, g.cells[row][col])
	}
}

// Draw renders the grid component onto the screen.
func (g *Grid) Draw(screen tcell.Screen) {
	if !g.IsVisible() {
		return
	}

	x, y, width, height := g.GetRect()
	if width <= 0 || height <= 0 {
		return
	}

	// Ensure scroll/selection is valid before drawing
	g.ensureSelectionVisible()

	// Calculate effective cell width (considering autoWidth)
	effectiveCellWidth := g.cellWidth
	if g.autoWidth {
		effectiveCellWidth = g.calculateCellWidth()
		if effectiveCellWidth <= 0 {
			effectiveCellWidth = 1
		} // Safety
	}
	effectiveCellHeight := g.cellHeight
	if effectiveCellHeight <= 0 {
		effectiveCellHeight = 1
	} // Safety

	// Calculate how many cells fit
	visibleRows := 0
	if effectiveCellHeight > 0 {
		visibleRows = height / effectiveCellHeight
	} else {
		return
	} // Can't draw if cell height is invalid
	visibleCols := 0
	if effectiveCellWidth > 0 {
		visibleCols = width / effectiveCellWidth
	} else {
		return
	} // Can't draw if cell width is invalid

	// Get necessary state for drawing
	isFocused := g.IsFocused()
	currentTopRow := g.topRow
	currentLeftCol := g.leftCol
	selectedRow := g.selectedRow
	selectedCol := g.selectedCol
	// Create a copy of interacted cells for safe iteration during drawing
	interacted := make(map[string]bool, len(g.interactedCells))
	for k, v := range g.interactedCells {
		interacted[k] = v
	}

	// Fill background of the entire grid area using the grid's base style
	Fill(screen, x, y, width, height, ' ', g.style)

	// Draw visible cells
	for r := 0; r < visibleRows; r++ {
		gridRow := currentTopRow + r
		if gridRow >= len(g.cells) {
			break
		} // Stop if we run out of rows

		for c := 0; c < visibleCols; c++ {
			gridCol := currentLeftCol + c
			// Assumes rectangular grid, check column bounds based on the first row? Safest is to check each row.
			if gridCol >= len(g.cells[gridRow]) {
				break
			} // Stop if we run out of columns for *this* row

			// Calculate screen coordinates for this cell
			cellX := x + c*effectiveCellWidth
			cellY := y + r*effectiveCellHeight

			// Determine cell state
			isSelected := (gridRow == selectedRow && gridCol == selectedCol)
			cellKey := fmt.Sprintf("%d:%d", gridRow, gridCol)
			isInteracted := interacted[cellKey]

			// Determine cell style based on state and focus using the theme helper
			cellStyle := GetGridStyle(nil, // Use global theme
				func() State { // Determine state
					if isInteracted {
						return StateInteracted
					}
					if isSelected {
						return StateSelected
					}
					return StateNormal
				}(),
				isFocused, // Pass focus state
			)

			// Draw cell background using the determined style
			Fill(screen, cellX, cellY, effectiveCellWidth, effectiveCellHeight, ' ', cellStyle)

			// Draw selection indicator (if applicable)
			indicatorWidth := 0
			if g.showIndicator && isSelected && isFocused {
				// Draw indicator at the beginning of the cell
				indicatorX := cellX
				// Position indicator vertically in the middle if cellHeight > 1? For now, top.
				indicatorY := cellY + (effectiveCellHeight / 2)
				if effectiveCellHeight == 1 {
					indicatorY = cellY
				}

				// Use the dedicated indicator style
				DrawText(screen, indicatorX, indicatorY, g.indicatorStyle, string(g.indicatorChar))
				indicatorWidth = runewidth.RuneWidth(g.indicatorChar)
			}

			// Draw cell content with padding and truncation
			// Content starts after indicator (if shown) and left padding
			contentStartX := cellX + indicatorWidth + g.padding
			// Available width is cell width minus left padding, right padding, and indicator width
			contentMaxWidth := effectiveCellWidth - g.padding - g.padding - indicatorWidth
			// Position content vertically in the middle if cellHeight > 1? For now, top.
			contentY := cellY + (effectiveCellHeight / 2)
			if effectiveCellHeight == 1 {
				contentY = cellY
			}

			if contentMaxWidth > 0 && contentY < y+height { // Check content fits and Y is valid
				content := g.cells[gridRow][gridCol]
				// Truncate content if it's wider than available space
				displayText := runewidth.Truncate(content, contentMaxWidth, "…") // Use ellipsis for truncation
				DrawText(screen, contentStartX, contentY, cellStyle, displayText)
			}
		}
	}
}

// calculateCellWidth determines the required width for cells when autoWidth is enabled.
// It finds the widest cell content and adds padding/indicator space.
func (g *Grid) calculateCellWidth() int {
	if !g.autoWidth {
		return g.cellWidth
	} // Return fixed width if not auto

	// Calculate base width needed for padding and potential indicator
	// Indicator space is only added if shown, assume max 1 cell width
	indicatorSpace := 0
	if g.showIndicator {
		indicatorSpace = runewidth.RuneWidth(g.indicatorChar)
	}
	baseWidth := g.padding + g.padding + indicatorSpace // Left pad + Right pad + Indicator

	// Find the maximum width of cell content
	maxContentWidth := 0
	for _, row := range g.cells {
		for _, cell := range row {
			width := runewidth.StringWidth(cell)
			if width > maxContentWidth {
				maxContentWidth = width
			}
		}
	}

	// Total width is base + max content
	totalWidth := baseWidth + maxContentWidth

	// Ensure a minimum reasonable width (e.g., padding + 1 char + indicator)
	minWidth := baseWidth + 1
	if totalWidth < minWidth {
		return minWidth
	}

	return totalWidth
}

// HandleEvent processes keyboard events for grid navigation and interaction.
func (g *Grid) HandleEvent(event tcell.Event) bool {
	keyEvent, ok := event.(*tcell.EventKey)
	if !ok {
		return false // Not a key event
	}

	// Ensure grid has content to navigate/interact with
	numRows := len(g.cells)
	numCols := 0
	if numRows > 0 {
		numCols = len(g.cells[0])
	} // Assumes rectangular
	hasContent := numRows > 0 && numCols > 0

	if !hasContent {
		return false // Cannot navigate/interact with empty grid
	}

	// --- Navigation ---
	currentRow, currentCol := g.selectedRow, g.selectedCol
	// If no selection yet, start at 0,0 for navigation calculations
	if currentRow < 0 {
		currentRow = 0
	}
	if currentCol < 0 {
		currentCol = 0
	}

	newRow, newCol := currentRow, currentCol

	switch keyEvent.Key() {
	case tcell.KeyUp:
		newRow--
	case tcell.KeyDown:
		newRow++
	case tcell.KeyLeft:
		newCol--
	case tcell.KeyRight:
		newCol++
	case tcell.KeyHome:
		newCol = 0
	case tcell.KeyEnd:
		newCol = numCols - 1
	case tcell.KeyPgUp:
		_, _, _, height := g.GetRect()
		if height <= 0 {
			height = 1
		} // Avoid division by zero
		cellH := g.cellHeight
		if cellH <= 0 {
			cellH = 1
		}
		pageSize := height / cellH
		if pageSize <= 0 {
			pageSize = 1
		}
		newRow -= pageSize
	case tcell.KeyPgDn:
		_, _, _, height := g.GetRect()
		if height <= 0 {
			height = 1
		}
		cellH := g.cellHeight
		if cellH <= 0 {
			cellH = 1
		}
		pageSize := height / cellH
		if pageSize <= 0 {
			pageSize = 1
		}
		newRow += pageSize
	case tcell.KeyEnter, tcell.KeyRune: // Check Enter or specific runes
		if keyEvent.Key() == tcell.KeyEnter || keyEvent.Rune() == ' ' { // Enter or Space for interaction
			g.toggleCellInteraction()
			return true // Event handled (interaction)
		}
		// Check vim-style navigation runes
		if keyEvent.Key() == tcell.KeyRune {
			switch keyEvent.Rune() {
			case 'k':
				newRow-- // Up
			case 'j':
				newRow++ // Down
			case 'h':
				newCol-- // Left
			case 'l':
				newCol++ // Right
			default:
				return false // Unhandled rune
			}
			// Navigation rune handled, proceed to selectCell
		} else {
			return false // Unhandled non-rune key
		}

	default:
		return false // Unhandled key
	}

	// If navigation keys were pressed, attempt to select the new cell
	// selectCell handles bounds checking and returns true if selection changed
	return g.selectCell(newRow, newCol)
}

// --- Interaction State Methods ---

// IsCellInteracted checks if a specific cell is marked as interacted.
func (g *Grid) IsCellInteracted(row, col int) bool {
	// Validate coords against grid bounds
	if row < 0 || row >= len(g.cells) || col < 0 || col >= len(g.cells[row]) {
		return false
	}
	cellKey := fmt.Sprintf("%d:%d", row, col)
	return g.interactedCells[cellKey] // Returns false if key doesn't exist
}

// SetCellInteracted explicitly sets the interaction state of a cell.
// Respects the SelectionMode (clears others if SingleSelect).
func (g *Grid) SetCellInteracted(row, col int, interacted bool) {
	// Validate coordinates
	if row < 0 || row >= len(g.cells) || col < 0 || col >= len(g.cells[row]) {
		return // Cannot set state for invalid cell
	}

	cellKey := fmt.Sprintf("%d:%d", row, col)
	currentState := g.interactedCells[cellKey]

	// Only proceed if state needs changing
	if currentState == interacted {
		return
	}

	stateChanged := false
	if interacted { // Trying to set to true
		if g.selectionMode == SingleSelect {
			// Clear all existing interactions before setting the new one
			if len(g.interactedCells) > 0 {
				g.interactedCells = make(map[string]bool)
				stateChanged = true // State changed if we cleared anything
			}
		}
		g.interactedCells[cellKey] = true
		stateChanged = true // State changed (either added or cleared then added)
	} else { // Trying to set to false
		delete(g.interactedCells, cellKey)
		stateChanged = true // State changed (removed)
	}

	if stateChanged {
		g.MarkDirty() // State changed
	}
}

// GetInteractedCells returns a slice of [row, col] pairs for all interacted cells.
// Returns an empty slice if no cells are interacted.
func (g *Grid) GetInteractedCells() [][2]int {
	// Pre-allocate slice with exact capacity
	result := make([][2]int, 0, len(g.interactedCells))

	for key := range g.interactedCells {
		var r, c int
		// Use Sscanf to parse the row:col key back into integers
		_, err := fmt.Sscanf(key, "%d:%d", &r, &c)
		if err == nil { // Only add if parsing was successful
			result = append(result, [2]int{r, c})
		}
		// Consider logging parse errors?
	}
	// TODO: Consider sorting the result by row then column for predictable order?
	// sort.Slice(result, func(i, j int) bool {
	//     if result[i][0] != result[j][0] {
	//         return result[i][0] < result[j][0]
	//     }
	//     return result[i][1] < result[j][1]
	// })
	return result
}

// ClearInteractions resets the interaction state for all cells.
func (g *Grid) ClearInteractions() {
	if len(g.interactedCells) > 0 {
		g.interactedCells = make(map[string]bool) // Reset the map
		g.MarkDirty()                             // Need redraw if interactions cleared
	}
}