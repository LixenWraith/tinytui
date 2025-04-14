// draw.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// Box-drawing runes
const (
	// Single line box drawing
	RuneULCorner rune = tcell.RuneULCorner // Upper left corner
	RuneURCorner rune = tcell.RuneURCorner // Upper right corner
	RuneLLCorner rune = tcell.RuneLLCorner // Lower left corner
	RuneLRCorner rune = tcell.RuneLRCorner // Lower right corner
	RuneHLine    rune = tcell.RuneHLine    // Horizontal line
	RuneVLine    rune = tcell.RuneVLine    // Vertical line

	// Double line box drawing
	RuneDoubleULCorner rune = '╔' // Upper left corner
	RuneDoubleURCorner rune = '╗' // Upper right corner
	RuneDoubleLLCorner rune = '╚' // Lower left corner
	RuneDoubleLRCorner rune = '╝' // Lower right corner
	RuneDoubleHLine    rune = '═' // Horizontal line
	RuneDoubleVLine    rune = '║' // Vertical line

	// Block elements
	RuneBlock          rune = tcell.RuneBlock // Full block
	RuneUpperHalfBlock rune = '▀'             // Upper half block
	RuneLowerHalfBlock rune = '▄'             // Lower half block
)

// Fill fills a rectangular area with the specified rune and style.
func Fill(screen tcell.Screen, x, y, width, height int, char rune, style Style) {
	if width <= 0 || height <= 0 {
		return
	}

	tcellStyle := style.ToTcell()

	// Get screen dimensions for bounds checking
	screenWidth, screenHeight := screen.Size()

	for row := y; row < y+height; row++ {
		// Skip rows outside screen bounds
		if row < 0 || row >= screenHeight {
			continue
		}

		for col := x; col < x+width; col++ {
			// Skip columns outside screen bounds
			if col < 0 || col >= screenWidth {
				continue
			}

			screen.SetContent(col, row, char, nil, tcellStyle)
		}
	}
}

// DrawBox draws a box with single-line borders.
func DrawBox(screen tcell.Screen, x, y, width, height int, style Style) {
	if width <= 1 || height <= 1 {
		return
	}

	tcellStyle := style.ToTcell()

	// Get screen dimensions for bounds checking
	screenWidth, screenHeight := screen.Size()

	// Calculate bottom position once to ensure consistency
	bottomY := y + height - 1

	// Draw the box in stages to ensure all parts are drawn correctly

	// 1. Draw the four corners first
	if y >= 0 && y < screenHeight && x >= 0 && x < screenWidth {
		screen.SetContent(x, y, RuneULCorner, nil, tcellStyle)
	}

	if y >= 0 && y < screenHeight && x+width-1 >= 0 && x+width-1 < screenWidth {
		screen.SetContent(x+width-1, y, RuneURCorner, nil, tcellStyle)
	}

	if bottomY >= 0 && bottomY < screenHeight && x >= 0 && x < screenWidth {
		screen.SetContent(x, bottomY, RuneLLCorner, nil, tcellStyle)
	}

	if bottomY >= 0 && bottomY < screenHeight && x+width-1 >= 0 && x+width-1 < screenWidth {
		screen.SetContent(x+width-1, bottomY, RuneLRCorner, nil, tcellStyle)
	}

	// 2. Draw the horizontal lines between corners
	for col := x + 1; col < x+width-1; col++ {
		// Skip columns outside screen bounds
		if col < 0 || col >= screenWidth {
			continue
		}

		// Top line
		if y >= 0 && y < screenHeight {
			screen.SetContent(col, y, RuneHLine, nil, tcellStyle)
		}

		// Bottom line - explicit drawing with careful bounds checking
		if bottomY >= 0 && bottomY < screenHeight {
			screen.SetContent(col, bottomY, RuneHLine, nil, tcellStyle)
		}
	}

	// 3. Draw the vertical lines between corners
	for row := y + 1; row < y+height-1; row++ {
		// Skip rows outside screen bounds
		if row < 0 || row >= screenHeight {
			continue
		}

		// Left line
		if x >= 0 && x < screenWidth {
			screen.SetContent(x, row, RuneVLine, nil, tcellStyle)
		}

		// Right line
		if x+width-1 >= 0 && x+width-1 < screenWidth {
			screen.SetContent(x+width-1, row, RuneVLine, nil, tcellStyle)
		}
	}
}

// DrawDoubleBox draws a box with double-line borders.
func DrawDoubleBox(screen tcell.Screen, x, y, width, height int, style Style) {
	if width <= 1 || height <= 1 {
		return
	}

	tcellStyle := style.ToTcell()

	// Get screen dimensions for bounds checking
	screenWidth, screenHeight := screen.Size()

	// Calculate bottom position once to ensure consistency
	bottomY := y + height - 1

	// Draw the box in stages to ensure all parts are drawn correctly

	// 1. Draw the four corners first
	if y >= 0 && y < screenHeight && x >= 0 && x < screenWidth {
		screen.SetContent(x, y, RuneDoubleULCorner, nil, tcellStyle)
	}

	if y >= 0 && y < screenHeight && x+width-1 >= 0 && x+width-1 < screenWidth {
		screen.SetContent(x+width-1, y, RuneDoubleURCorner, nil, tcellStyle)
	}

	if bottomY >= 0 && bottomY < screenHeight && x >= 0 && x < screenWidth {
		screen.SetContent(x, bottomY, RuneDoubleLLCorner, nil, tcellStyle)
	}

	if bottomY >= 0 && bottomY < screenHeight && x+width-1 >= 0 && x+width-1 < screenWidth {
		screen.SetContent(x+width-1, bottomY, RuneDoubleLRCorner, nil, tcellStyle)
	}

	// 2. Draw the horizontal lines between corners
	for col := x + 1; col < x+width-1; col++ {
		// Skip columns outside screen bounds
		if col < 0 || col >= screenWidth {
			continue
		}

		// Top line
		if y >= 0 && y < screenHeight {
			screen.SetContent(col, y, RuneDoubleHLine, nil, tcellStyle)
		}

		// Bottom line - explicit drawing with careful bounds checking
		if bottomY >= 0 && bottomY < screenHeight {
			screen.SetContent(col, bottomY, RuneDoubleHLine, nil, tcellStyle)
		}
	}

	// 3. Draw the vertical lines between corners
	for row := y + 1; row < y+height-1; row++ {
		// Skip rows outside screen bounds
		if row < 0 || row >= screenHeight {
			continue
		}

		// Left line
		if x >= 0 && x < screenWidth {
			screen.SetContent(x, row, RuneDoubleVLine, nil, tcellStyle)
		}

		// Right line
		if x+width-1 >= 0 && x+width-1 < screenWidth {
			screen.SetContent(x+width-1, row, RuneDoubleVLine, nil, tcellStyle)
		}
	}
}

// DrawSolidBox draws a box using block elements.
func DrawSolidBox(screen tcell.Screen, x, y, width, height int, style Style) {
	if width <= 0 || height <= 0 {
		return
	}

	tcellStyle := style.ToTcell()

	// Get screen dimensions for bounds checking
	screenWidth, screenHeight := screen.Size()

	// Calculate bottom position once to ensure consistency
	bottomY := y + height - 1

	// If dimensions are too small, draw a solid block
	if width == 1 && height == 1 {
		if x >= 0 && x < screenWidth && y >= 0 && y < screenHeight {
			screen.SetContent(x, y, RuneBlock, nil, tcellStyle)
		}
		return
	}

	// Draw the box in stages to ensure all parts are drawn correctly

	// 1. Draw top and bottom lines
	for col := x; col < x+width; col++ {
		// Skip columns outside screen bounds
		if col < 0 || col >= screenWidth {
			continue
		}

		// Top line
		if y >= 0 && y < screenHeight {
			screen.SetContent(col, y, RuneUpperHalfBlock, nil, tcellStyle)
		}

		// Bottom line - explicit drawing with careful bounds checking
		if bottomY >= 0 && bottomY < screenHeight && height > 1 {
			screen.SetContent(col, bottomY, RuneLowerHalfBlock, nil, tcellStyle)
		}
	}

	// 2. Draw the vertical sides
	for row := y + 1; row < bottomY; row++ {
		// Skip rows outside screen bounds
		if row < 0 || row >= screenHeight {
			continue
		}

		// Left line
		if x >= 0 && x < screenWidth {
			screen.SetContent(x, row, RuneBlock, nil, tcellStyle)
		}

		// Right line
		if x+width-1 >= 0 && x+width-1 < screenWidth && width > 1 {
			screen.SetContent(x+width-1, row, RuneBlock, nil, tcellStyle)
		}
	}
}

// DrawText draws text at the specified position with the given style.
func DrawText(screen tcell.Screen, x, y int, style Style, text string) {
	if y < 0 {
		return
	}

	// Get screen dimensions for bounds checking
	screenWidth, screenHeight := screen.Size()

	if y >= screenHeight {
		return
	}

	tcellStyle := style.ToTcell()

	// Draw each rune, accounting for wide characters
	startX := x

	for _, r := range text {
		width := runewidth.RuneWidth(r)

		// Skip characters that would be drawn before the screen boundary
		if startX+width <= 0 {
			startX += width
			continue
		}

		// Stop if we've reached the right edge of the screen
		if startX >= screenWidth {
			break
		}

		// Only draw characters that are at least partially visible
		if startX >= 0 {
			screen.SetContent(startX, y, r, nil, tcellStyle)
		}

		startX += width
	}
}