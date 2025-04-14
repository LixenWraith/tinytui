// sprite.go
package tinytui

import (
	"strings" // For string processing in SetContent

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// SpriteCell defines a single 'pixel' in the sprite, containing a rune and its style.
type SpriteCell struct {
	Rune  rune
	Style Style
}

// Sprite displays a fixed grid of styled characters (SpriteCells).
// Useful for simple pixel-art style graphics or fixed character-based layouts.
// Cells with no explicit background set in their Style are treated as transparent.
type Sprite struct {
	BaseComponent
	cells [][]SpriteCell // 2D array of cells [row][col]
	style Style          // Base style applied to the background *behind* transparent sprite cells
}

// NewSprite creates a new sprite component with initial cell data.
// Initializes background style from the current theme's default text style.
func NewSprite(cells [][]SpriteCell) *Sprite {
	theme := GetTheme()
	if theme == nil {
		theme = NewDefaultTheme()
	} // Fallback

	s := &Sprite{
		BaseComponent: NewBaseComponent(),
		cells:         cells,             // Use provided cells
		style:         theme.TextStyle(), // Use theme's base text style for sprite background
	}
	// Ensure the sprite starts with the correct theme applied (for background style)
	s.ApplyTheme(theme)
	return s
}

// ApplyTheme updates the sprite's base background style from the theme.
// Individual cell styles are typically set explicitly and may not react to theme changes unless done manually.
// Implements ThemedComponent.
func (s *Sprite) ApplyTheme(theme Theme) {
	if theme == nil {
		return
	}
	newStyle := theme.TextStyle() // Update background style from theme's text style
	if s.style != newStyle {
		s.style = newStyle
		s.MarkDirty() // Mark dirty if background style changed
	}
}

// SetCells replaces the sprite's entire cell data. The input `cells` should ideally be rectangular.
func (s *Sprite) SetCells(cells [][]SpriteCell) {
	// TODO: Add validation? Ensure rectangular? Or handle jagged arrays in Draw?
	// For now, assume caller provides valid data.
	s.cells = cells
	s.MarkDirty() // Content changed, needs redraw
}

// GetCells returns a deep copy of the sprite's cell data.
// This prevents external modification of the internal state.
func (s *Sprite) GetCells() [][]SpriteCell {
	if s.cells == nil {
		return nil
	}
	// Create a deep copy
	newCells := make([][]SpriteCell, len(s.cells))
	maxCols := 0
	if len(s.cells) > 0 {
		maxCols = len(s.cells[0])
	} // Assume rectangular for copy size
	for i, row := range s.cells {
		// Ensure consistency if source was jagged
		rowLen := len(row)
		if rowLen > maxCols {
			maxCols = rowLen
		} // Find true max cols

		newCells[i] = make([]SpriteCell, rowLen) // Allocate based on actual row length
		copy(newCells[i], row)
	}
	// Optional: Pad result to be rectangular based on maxCols found? Or return as is? Return as is.
	return newCells
}

// SetStyle explicitly sets the sprite's base background style (drawn behind transparent cells).
// Consider using themes instead for consistent styling.
func (s *Sprite) SetStyle(style Style) {
	if s.style != style {
		s.style = style
		s.MarkDirty()
	}
}

// SetCell updates a specific cell (pixel) in the sprite at the given row and column.
// Coordinates are 0-based. Marks dirty if the cell exists and its value changes.
func (s *Sprite) SetCell(row, col int, cell SpriteCell) {
	// Bounds check
	if s.cells == nil || row < 0 || row >= len(s.cells) || col < 0 || col >= len(s.cells[row]) {
		return // Invalid coordinates
	}

	// Only mark dirty if the cell content actually changes
	if s.cells[row][col] != cell {
		s.cells[row][col] = cell
		s.MarkDirty()
	}
}

// GetCell retrieves the SpriteCell data at a specific coordinate.
// Returns the cell and true if coordinates are valid, otherwise an empty cell and false.
func (s *Sprite) GetCell(row, col int) (SpriteCell, bool) {
	// Bounds check
	if s.cells == nil || row < 0 || row >= len(s.cells) || col < 0 || col >= len(s.cells[row]) {
		return SpriteCell{}, false // Invalid coordinates
	}
	return s.cells[row][col], true
}

// Dimensions returns the width (max columns) and height (number of rows) of the sprite data.
func (s *Sprite) Dimensions() (width, height int) {
	if s.cells == nil {
		return 0, 0
	}
	height = len(s.cells)
	width = 0
	// Find max width in case grid is jagged
	for _, row := range s.cells {
		if len(row) > width {
			width = len(row)
		}
	}
	return width, height
}

// Focusable returns false, as Sprites are typically non-interactive display elements.
func (s *Sprite) Focusable() bool {
	return false
}

// Draw renders the sprite onto the screen within the component's allocated rectangle.
// It respects cell transparency (cells with default background).
func (s *Sprite) Draw(screen tcell.Screen) {
	if !s.IsVisible() {
		return
	}

	x, y, width, height := s.GetRect()
	if width <= 0 || height <= 0 {
		return
	} // Cannot draw in zero area
	if s.cells == nil {
		return
	} // Nothing to draw if cells are nil

	// Fill the component's background area first using the sprite's base style
	Fill(screen, x, y, width, height, ' ', s.style)

	spriteDataHeight := len(s.cells)
	if spriteDataHeight == 0 {
		return
	} // Empty sprite data

	// Get the default background color for transparency check
	_, defaultBg, _, _ := DefaultStyle.Deconstruct()

	// Determine how much of the sprite data fits within the component's bounds
	rowsToDraw := min(height, spriteDataHeight)

	// Iterate through the rows and columns of the sprite data that fit
	for row := 0; row < rowsToDraw; row++ {
		if row >= len(s.cells) {
			break
		} // Safety check for jagged arrays
		spriteRow := s.cells[row]
		spriteDataWidth := len(spriteRow)
		if spriteDataWidth == 0 {
			continue
		} // Skip empty rows

		screenX := x       // Current drawing position on screen (horizontal)
		screenY := y + row // Current drawing position on screen (vertical)

		for col := 0; col < spriteDataWidth; col++ {
			// Stop drawing this row if we exceed the component's width
			if screenX >= x+width {
				break
			}

			cell := spriteRow[col]
			runeWidth := runewidth.RuneWidth(cell.Rune)

			// A cell is considered transparent if its rune is a space AND
			// its background color is the same as the default background color.
			_, cellBg, _, _ := cell.Style.Deconstruct() // Get the cell's background
			isTransparent := cell.Rune == ' ' && cellBg == defaultBg

			if !isTransparent {
				// Cell is not transparent, draw it using its own style
				effectiveStyle := cell.Style
				// If background wasn't set, merge with base style? No, treat as overlay.
				// If cell style has no bg, it uses default term bg.

				// Draw the rune, tcell handles clipping if rune starts inside bounds
				if screenX >= x { // Only draw if start is within component rect
					screen.SetContent(screenX, screenY, cell.Rune, nil, effectiveStyle.ToTcell())
					// Clear subsequent cells if it's a wide rune and within bounds
					for i := 1; i < runeWidth; i++ {
						if screenX+i < x+width {
							screen.SetContent(screenX+i, screenY, ' ', nil, effectiveStyle.ToTcell())
						}
					}
				}
			}
			// Advance screen X position by the width of the rune we just processed/skipped
			screenX += runeWidth
		}
	}
}

// HandleEvent processes events. Sprites typically don't handle any events.
func (s *Sprite) HandleEvent(event tcell.Event) bool {
	return false // Not handled
}

// Resize changes the sprite's internal cell grid dimensions.
// Preserves existing cell data where possible. New cells are default (transparent space).
func (s *Sprite) Resize(newWidth, newHeight int) {
	if newWidth < 0 || newHeight < 0 {
		return
	} // Invalid dimensions

	oldHeight, oldWidth := s.Dimensions() // Get current dimensions

	if oldWidth == newWidth && oldHeight == newHeight {
		return
	} // No change

	// Create new cells array, initialized with default transparent cells
	newCells := make([][]SpriteCell, newHeight)
	defaultCell := SpriteCell{Rune: ' ', Style: DefaultStyle}
	for i := range newCells {
		newCells[i] = make([]SpriteCell, newWidth)
		for j := range newCells[i] {
			newCells[i][j] = defaultCell
		}
	}

	// Copy existing data within the bounds of both old and new dimensions
	copyHeight := min(newHeight, oldHeight)
	copyWidth := min(newWidth, oldWidth) // Use calculated oldWidth (max cols)
	for i := 0; i < copyHeight; i++ {
		// Ensure source row exists and copy up to copyWidth or actual row length
		if i < len(s.cells) {
			rowCopyWidth := min(copyWidth, len(s.cells[i]))
			copy(newCells[i][:rowCopyWidth], s.cells[i][:rowCopyWidth])
		}
	}

	s.cells = newCells
	s.MarkDirty()
}

// Clear sets all sprite cells to the specified cell data.
// Use a transparent cell (e.g., SpriteCell{Rune: ' ', Style: DefaultStyle})
// to effectively clear to the sprite's base background style.
func (s *Sprite) Clear(cell SpriteCell) {
	if s.cells == nil {
		return
	}
	dirty := false
	for i := range s.cells {
		for j := range s.cells[i] {
			if s.cells[i][j] != cell {
				s.cells[i][j] = cell
				dirty = true
			}
		}
	}
	if dirty {
		s.MarkDirty()
	}
}

// SetContent implements TextUpdater by converting a multi-line string into sprite cells.
// Each character becomes a cell. Non-space characters get an opaque background, spaces are transparent.
// This provides a basic way to display text as a sprite.
func (s *Sprite) SetContent(content string) {
	lines := strings.Split(content, "\n")
	// Handle potential trailing newline creating an empty string element
	if len(lines) > 0 && lines[len(lines)-1] == "" && strings.HasSuffix(content, "\n") {
		lines = lines[:len(lines)-1]
	}

	height := len(lines)
	width := 0
	for _, line := range lines {
		lineWidth := runewidth.StringWidth(line) // Calculate visual width
		if lineWidth > width {
			width = lineWidth
		}
	}

	// Create new cells array
	cells := make([][]SpriteCell, height)
	// Define styles for opaque (non-space) and transparent (space) cells
	opaqueStyle := DefaultStyle.Background(ColorBlack) // Example: Black background
	transparentStyle := DefaultStyle                   // Default background = transparent

	for i := range cells {
		cells[i] = make([]SpriteCell, width)
		lineRunes := []rune(lines[i])
		cellCol := 0 // Tracks the current column index in the cells[i] slice

		for _, r := range lineRunes { // Iterate through runes in the line
			if cellCol >= width {
				break
			} // Stop if we exceed calculated width

			rw := runewidth.RuneWidth(r)
			cellStyle := opaqueStyle
			if r == ' ' {
				cellStyle = transparentStyle // Spaces are transparent
			}

			// Set the primary cell for the rune
			cells[i][cellCol] = SpriteCell{Rune: r, Style: cellStyle}
			// Fill subsequent cells for wide runes
			for k := 1; k < rw; k++ {
				if cellCol+k < width { // Check bounds
					cells[i][cellCol+k] = SpriteCell{Rune: ' ', Style: cellStyle} // Fill with same style
				}
			}
			cellCol += rw // Advance by rune width
		}
		// Fill remaining columns in this row with transparent spaces if line was shorter
		for ; cellCol < width; cellCol++ {
			cells[i][cellCol] = SpriteCell{Rune: ' ', Style: transparentStyle}
		}
	}

	s.SetCells(cells) // Update the sprite's internal cells
}

// SetCellsFromStrings sets sprite content from a slice of strings, applying a base style.
// Each string is a row. Spaces in the strings are treated as transparent cells,
// other characters use the provided `style`. Handles wide runes.
func (s *Sprite) SetCellsFromStrings(rows []string, style Style) {
	height := len(rows)
	width := 0
	for _, row := range rows {
		rowWidth := runewidth.StringWidth(row)
		if rowWidth > width {
			width = rowWidth
		}
	}

	// Create new cells array
	cells := make([][]SpriteCell, height)
	transparentStyle := DefaultStyle // Style with no background set = transparent

	for i := range cells {
		cells[i] = make([]SpriteCell, width)
		lineRunes := []rune(rows[i])
		cellCol := 0 // Current column index in cells[i]

		for _, r := range lineRunes { // Iterate through runes
			if cellCol >= width {
				break
			} // Exceeded width

			rw := runewidth.RuneWidth(r)
			cellStyle := style // Use provided style by default
			if r == ' ' {
				cellStyle = transparentStyle // Spaces are transparent
			}

			// Set primary cell
			cells[i][cellCol] = SpriteCell{Rune: r, Style: cellStyle}
			// Fill subsequent cells for wide runes
			for k := 1; k < rw; k++ {
				if cellCol+k < width {
					cells[i][cellCol+k] = SpriteCell{Rune: ' ', Style: cellStyle}
				}
			}
			cellCol += rw // Advance column index
		}
		// Fill remaining columns with transparent spaces
		for ; cellCol < width; cellCol++ {
			cells[i][cellCol] = SpriteCell{Rune: ' ', Style: transparentStyle}
		}
	}

	s.SetCells(cells) // Update sprite data
}