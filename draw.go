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

	// Draw corners
	if y >= 0 && y < screenHeight && x >= 0 && x < screenWidth {
		screen.SetContent(x, y, RuneULCorner, nil, tcellStyle)
	}
	if y >= 0 && y < screenHeight && x+width-1 >= 0 && x+width-1 < screenWidth {
		screen.SetContent(x+width-1, y, RuneURCorner, nil, tcellStyle)
	}
	if y+height-1 >= 0 && y+height-1 < screenHeight && x >= 0 && x < screenWidth {
		screen.SetContent(x, y+height-1, RuneLLCorner, nil, tcellStyle)
	}
	if y+height-1 >= 0 && y+height-1 < screenHeight && x+width-1 >= 0 && x+width-1 < screenWidth {
		screen.SetContent(x+width-1, y+height-1, RuneLRCorner, nil, tcellStyle)
	}

	// Draw horizontal lines
	for col := x + 1; col < x+width-1; col++ {
		// Skip columns outside screen bounds
		if col < 0 || col >= screenWidth {
			continue
		}

		// Top line
		if y >= 0 && y < screenHeight {
			screen.SetContent(col, y, RuneHLine, nil, tcellStyle)
		}

		// Bottom line
		if y+height-1 >= 0 && y+height-1 < screenHeight {
			screen.SetContent(col, y+height-1, RuneHLine, nil, tcellStyle)
		}
	}

	// Draw vertical lines
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

	// Draw corners
	if y >= 0 && y < screenHeight && x >= 0 && x < screenWidth {
		screen.SetContent(x, y, RuneDoubleULCorner, nil, tcellStyle)
	}
	if y >= 0 && y < screenHeight && x+width-1 >= 0 && x+width-1 < screenWidth {
		screen.SetContent(x+width-1, y, RuneDoubleURCorner, nil, tcellStyle)
	}
	if y+height-1 >= 0 && y+height-1 < screenHeight && x >= 0 && x < screenWidth {
		screen.SetContent(x, y+height-1, RuneDoubleLLCorner, nil, tcellStyle)
	}
	if y+height-1 >= 0 && y+height-1 < screenHeight && x+width-1 >= 0 && x+width-1 < screenWidth {
		screen.SetContent(x+width-1, y+height-1, RuneDoubleLRCorner, nil, tcellStyle)
	}

	// Draw horizontal lines
	for col := x + 1; col < x+width-1; col++ {
		// Skip columns outside screen bounds
		if col < 0 || col >= screenWidth {
			continue
		}

		// Top line
		if y >= 0 && y < screenHeight {
			screen.SetContent(col, y, RuneDoubleHLine, nil, tcellStyle)
		}

		// Bottom line
		if y+height-1 >= 0 && y+height-1 < screenHeight {
			screen.SetContent(col, y+height-1, RuneDoubleHLine, nil, tcellStyle)
		}
	}

	// Draw vertical lines
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

	// If dimensions are too small, draw a solid block
	if width == 1 && height == 1 {
		if x >= 0 && x < screenWidth && y >= 0 && y < screenHeight {
			screen.SetContent(x, y, RuneBlock, nil, tcellStyle)
		}
		return
	}

	// Draw top and bottom lines
	for col := x; col < x+width; col++ {
		// Skip columns outside screen bounds
		if col < 0 || col >= screenWidth {
			continue
		}

		// Top line
		if y >= 0 && y < screenHeight {
			screen.SetContent(col, y, RuneUpperHalfBlock, nil, tcellStyle)
		}

		// Bottom line
		if y+height-1 >= 0 && y+height-1 < screenHeight && height > 1 {
			screen.SetContent(col, y+height-1, RuneLowerHalfBlock, nil, tcellStyle)
		}
	}

	// Draw left and right lines
	for row := y + 1; row < y+height-1; row++ {
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

// DrawTextCentered draws text centered horizontally within the specified width.
func DrawTextCentered(screen tcell.Screen, x, y, width int, style Style, text string) {
	textWidth := runewidth.StringWidth(text)

	// Calculate start position for centering
	startX := x + (width-textWidth)/2
	if startX < x {
		startX = x // Don't start before the specified x position
	}

	DrawText(screen, startX, y, style, text)
}

// DrawTextRight draws text aligned to the right within the specified width.
func DrawTextRight(screen tcell.Screen, x, y, width int, style Style, text string) {
	textWidth := runewidth.StringWidth(text)

	// Calculate start position for right alignment
	startX := x + width - textWidth
	if startX < x {
		startX = x // Don't start before the specified x position
	}

	DrawText(screen, startX, y, style, text)
}
