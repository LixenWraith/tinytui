// widgets/sprite.go
package widgets

import (
	"sync"

	"github.com/LixenWraith/tinytui"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// SpriteCell defines the content and style of a single cell within a Sprite.
type SpriteCell struct {
	Rune  rune
	Style tinytui.Style
}

// Sprite is a widget that displays a fixed 2D grid of styled characters.
// Cells where the Style has no explicitly set background color are treated as transparent.
type Sprite struct {
	tinytui.BaseWidget
	mu    sync.RWMutex
	cells [][]SpriteCell // The grid data [row][col]
}

// NewSprite creates a new Sprite widget with the given initial data.
// Data is expected as a slice of rows, where each row is a slice of SpriteCells.
// The sprite's dimensions are determined by the provided data.
func NewSprite(data [][]SpriteCell) *Sprite {
	s := &Sprite{
		cells: data,
	}
	s.SetVisible(true) // Explicitly set visibility
	// Initial SetRect will be called by the layout later
	return s
}

// SetData updates the data displayed by the sprite.
func (s *Sprite) SetData(data [][]SpriteCell) *Sprite {
	s.mu.Lock()
	s.cells = data
	s.mu.Unlock()
	if app := s.App(); app != nil {
		app.QueueRedraw()
	}
	return s
}

// GetData returns the current sprite data.
// Returns a copy to prevent modification issues? Or rely on caller politeness?
// Let's return the internal slice for now for efficiency, but document it.
// Note: Modifying the returned slice directly is not recommended; use SetData.
func (s *Sprite) GetData() [][]SpriteCell {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Consider returning a deep copy if external modification is a concern
	return s.cells
}

// Draw renders the sprite onto the screen.
// It respects the widget's bounds and treats cells with no explicit
// background style as transparent.
func (s *Sprite) Draw(screen tcell.Screen) {
	s.BaseWidget.Draw(screen)

	x, y, width, height := s.GetRect()
	if width <= 0 || height <= 0 {
		return
	}

	s.mu.RLock()
	cellsData := s.cells
	s.mu.RUnlock()

	spriteHeight := len(cellsData)
	if spriteHeight == 0 {
		return
	}
	spriteWidth := 0
	if spriteHeight > 0 {
		spriteWidth = len(cellsData[0]) // Assume rectangular
	}
	if spriteWidth == 0 {
		return
	}

	// Iterate over the widget's area
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			// Calculate corresponding cell index in the sprite data
			spriteRow := row
			spriteCol := col

			// Check if the cell index is within the sprite data bounds
			if spriteRow >= 0 && spriteRow < spriteHeight && spriteCol >= 0 && spriteCol < spriteWidth {
				cell := cellsData[spriteRow][spriteCol]

				// Check for transparency: Is the background color explicitly set?
				_, _, _, bgSet := cell.Style.Deconstruct()

				if bgSet {
					// Not transparent: Draw the cell
					// Ensure we don't draw outside the widget's bounds (redundant check?)
					drawX, drawY := x+col, y+row
					if drawX >= x && drawX < x+width && drawY >= y && drawY < y+height {
						// Handle rune width (though sprite cells are usually 1 wide)
						// For simplicity here, assume single-width runes or clipping handled by SetContent
						rw := runewidth.RuneWidth(cell.Rune)
						if drawX+rw <= x+width { // Check if rune fits
							screen.SetContent(drawX, drawY, cell.Rune, nil, cell.Style.ToTcell())
							// If rune width > 1, clear subsequent cells it covers
							for i := 1; i < rw; i++ {
								if drawX+i < x+width {
									// Setting content with a zero rune might clear it,
									// or we might need to use ' ' with the same style background.
									// Let's use ' ' for clarity.
									screen.SetContent(drawX+i, drawY, ' ', nil, cell.Style.ToTcell())
								}
							}
							col += rw - 1 // Advance col index past the wide rune
						} else {
							// Rune doesn't fit, draw nothing for this cell position
						}
					}
				} else {
					// Transparent: Do nothing, leave the underlying content visible
				}
			} else {
				// This part of the widget rect is outside the sprite data bounds.
				// Should we clear it or leave it? Let's leave it for now,
				// assuming the layout/parent handles clearing.
			}
		}
	}
}

// Focusable returns false, Sprites are not focusable by default.
func (s *Sprite) Focusable() bool {
	if !s.IsVisible() {
		return false
	}
	return false
}

// HandleEvent delegates to BaseWidget. Sprites don't handle events by default.
func (s *Sprite) HandleEvent(event tcell.Event) bool {
	return s.BaseWidget.HandleEvent(event)
}

// Width returns the width of the sprite data (number of columns).
func (s *Sprite) Width() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.cells) > 0 {
		return len(s.cells[0]) // Assume rectangular
	}
	return 0
}

// Height returns the height of the sprite data (number of rows).
func (s *Sprite) Height() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.cells)
}