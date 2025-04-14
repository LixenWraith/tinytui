// sprite.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// SpriteCell defines a single 'pixel' in the sprite.
type SpriteCell struct {
	Rune  rune
	Style Style
}

// Sprite displays a fixed grid of styled characters (pixels).
type Sprite struct {
	BaseComponent
	cells [][]SpriteCell
	style Style // Base style for background
}

// NewSprite creates a new sprite component.
func NewSprite(cells [][]SpriteCell) *Sprite {
	s := &Sprite{
		BaseComponent: NewBaseComponent(),
		cells:         cells,
		style:         DefaultStyle,
	}
	return s
}

// SetCells updates the sprite's character data.
func (s *Sprite) SetCells(cells [][]SpriteCell) {
	s.cells = cells
	s.MarkDirty()
}

// GetCells returns the sprite's character data.
func (s *Sprite) GetCells() [][]SpriteCell {
	// Return a deep copy to prevent external modification
	cells := make([][]SpriteCell, len(s.cells))
	for i, row := range s.cells {
		cells[i] = make([]SpriteCell, len(row))
		copy(cells[i], row)
	}

	return cells
}

// SetStyle sets the sprite's background style.
func (s *Sprite) SetStyle(style Style) {
	s.style = style
	s.MarkDirty()
}

// SetCell sets a specific cell in the sprite.
func (s *Sprite) SetCell(row, col int, cell SpriteCell) {
	// Ensure the cell is within bounds
	if row < 0 || row >= len(s.cells) || col < 0 || col >= len(s.cells[row]) {
		return
	}

	s.cells[row][col] = cell
	s.MarkDirty()
}

// GetCell gets a specific cell from the sprite.
func (s *Sprite) GetCell(row, col int) (SpriteCell, bool) {
	// Check if cell is within bounds
	if row < 0 || row >= len(s.cells) || col < 0 || col >= len(s.cells[row]) {
		return SpriteCell{}, false
	}

	return s.cells[row][col], true
}

// Dimensions returns the width and height of the sprite in cells.
func (s *Sprite) Dimensions() (width, height int) {
	height = len(s.cells)
	if height > 0 {
		width = len(s.cells[0])
	}

	return width, height
}

// Focusable returns whether the sprite can receive focus.
// Sprites are not focusable by default.
func (s *Sprite) Focusable() bool {
	return false
}

// Draw draws the sprite.
func (s *Sprite) Draw(screen tcell.Screen) {
	// Check visibility
	if !s.IsVisible() {
		return
	}

	// Get component dimensions
	x, y, width, height := s.GetRect()
	if width <= 0 || height <= 0 {
		return
	}

	// Fill background with base style
	Fill(screen, x, y, width, height, ' ', s.style)

	// Get sprite dimensions
	spriteHeight := len(s.cells)
	spriteWidth := 0
	if spriteHeight > 0 {
		spriteWidth = len(s.cells[0])
	}

	// Nothing to draw if sprite is empty
	if spriteHeight == 0 || spriteWidth == 0 {
		return
	}

	// Determine how much of the sprite to draw
	drawHeight := min(height, spriteHeight)
	drawWidth := min(width, spriteWidth)

	// Make a copy of cells for drawing outside lock
	cells := make([][]SpriteCell, drawHeight)
	for i := 0; i < drawHeight; i++ {
		cells[i] = make([]SpriteCell, drawWidth)
		copy(cells[i], s.cells[i][:drawWidth])
	}

	// Draw cells
	for row := 0; row < drawHeight; row++ {
		for col := 0; col < drawWidth; col++ {
			cell := cells[row][col]

			// Check if the cell has an explicit background color
			_, _, _, bgSet := cell.Style.Deconstruct()

			// If background color is set, draw the cell
			if bgSet {
				drawX := x + col
				drawY := y + row

				// Handle wide runes
				runeWidth := runewidth.RuneWidth(cell.Rune)

				// Draw only if it fits within bounds
				if drawX+runeWidth <= x+width {
					screen.SetContent(drawX, drawY, cell.Rune, nil, cell.Style.ToTcell())

					// Clear cells taken by wide rune
					for i := 1; i < runeWidth; i++ {
						if drawX+i < x+width {
							screen.SetContent(drawX+i, drawY, ' ', nil, cell.Style.ToTcell())
						}
					}
				}
			}
			// If no background color, treat as transparent
		}
	}
}

// HandleEvent handles events for the sprite.
// Sprites don't handle any events by default.
func (s *Sprite) HandleEvent(event tcell.Event) bool {
	return false
}

// Resize changes the sprite's dimensions, preserving existing cell data.
func (s *Sprite) Resize(width, height int) {
	if width < 0 || height < 0 {
		return
	}

	// Create new cells array
	newCells := make([][]SpriteCell, height)
	for i := range newCells {
		newCells[i] = make([]SpriteCell, width)
	}

	// Copy existing data
	copyHeight := min(height, len(s.cells))
	for i := 0; i < copyHeight; i++ {
		copyWidth := min(width, len(s.cells[i]))
		for j := 0; j < copyWidth; j++ {
			newCells[i][j] = s.cells[i][j]
		}
	}

	// Update cells
	s.cells = newCells
	s.MarkDirty()
}

// Clear sets all sprite cells to the specified cell.
// Use a transparent style (no background set) to clear to transparency.
func (s *Sprite) Clear(cell SpriteCell) {
	for i := range s.cells {
		for j := range s.cells[i] {
			s.cells[i][j] = cell
		}
	}

	s.MarkDirty()
}

// SetContent is an alias for SetCells that expects a string representation.
// This is a simplistic implementation to satisfy the TextUpdater interface.
// Each character in the string becomes a cell with the default style.
func (s *Sprite) SetContent(content string) {
	// Split the content by lines
	lines := splitString(content, '\n')

	// Determine dimensions
	height := len(lines)
	width := 0
	for _, line := range lines {
		lineWidth := 0
		for _, r := range line {
			lineWidth += runewidth.RuneWidth(r)
		}

		if lineWidth > width {
			width = lineWidth
		}
	}

	// Create cells
	cells := make([][]SpriteCell, height)
	for i := range cells {
		cells[i] = make([]SpriteCell, width)

		// Fill with spaces and transparent style
		for j := range cells[i] {
			cells[i][j] = SpriteCell{
				Rune:  ' ',
				Style: DefaultStyle,
			}
		}

		// Set characters from the line
		col := 0
		for _, r := range lines[i] {
			rw := runewidth.RuneWidth(r)

			if col < width {
				cells[i][col] = SpriteCell{
					Rune:  r,
					Style: DefaultStyle.Background(ColorBlack),
				}
			}

			col += rw
		}
	}

	s.SetCells(cells)
}

// SetCellsFromStrings creates a sprite from string representation.
// Each string represents a row, and each character becomes a cell.
// The style is applied to all non-space characters.
func (s *Sprite) SetCellsFromStrings(rows []string, style Style) {
	// Determine dimensions
	height := len(rows)
	width := 0
	for _, row := range rows {
		rowWidth := 0
		for _, r := range row {
			rowWidth += runewidth.RuneWidth(r)
		}

		if rowWidth > width {
			width = rowWidth
		}
	}

	// Create cells
	cells := make([][]SpriteCell, height)
	for i := range cells {
		cells[i] = make([]SpriteCell, width)

		// Fill with spaces and transparent style
		for j := range cells[i] {
			cells[i][j] = SpriteCell{
				Rune:  ' ',
				Style: DefaultStyle, // No background set = transparent
			}
		}

		// Set characters from the row
		col := 0
		for _, r := range rows[i] {
			rw := runewidth.RuneWidth(r)

			if col < width {
				cellStyle := style
				if r == ' ' {
					// Make spaces transparent
					cellStyle = DefaultStyle
				}

				cells[i][col] = SpriteCell{
					Rune:  r,
					Style: cellStyle,
				}
			}

			col += rw
		}
	}

	s.SetCells(cells)
}