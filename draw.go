// draw.go
package tinytui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// Box-drawing runes (using tcell constants where available)
const (
	// Single line box drawing
	RuneULCorner rune = tcell.RuneULCorner // Upper left corner '┌'
	RuneURCorner rune = tcell.RuneURCorner // Upper right corner '┐'
	RuneLLCorner rune = tcell.RuneLLCorner // Lower left corner '└'
	RuneLRCorner rune = tcell.RuneLRCorner // Lower right corner '┘'
	RuneHLine    rune = tcell.RuneHLine    // Horizontal line '─'
	RuneVLine    rune = tcell.RuneVLine    // Vertical line '│'

	// Double line box drawing
	RuneDoubleULCorner rune = '╔' // Upper left corner
	RuneDoubleURCorner rune = '╗' // Upper right corner
	RuneDoubleLLCorner rune = '╚' // Lower left corner
	RuneDoubleLRCorner rune = '╝' // Lower right corner
	RuneDoubleHLine    rune = '═' // Horizontal line
	RuneDoubleVLine    rune = '║' // Vertical line

	// Block elements
	RuneBlock          rune = '█' // tcell.RuneBlock
	RuneUpperHalfBlock rune = '▀' // Top horizontal line
	RuneLowerHalfBlock rune = '▄' // Bottom horizontal line
)

// Fill fills a rectangular area on the screen with a given rune and style.
// Performs bounds checking against the screen dimensions.
func Fill(screen tcell.Screen, x, y, width, height int, char rune, style Style) {
	if width <= 0 || height <= 0 {
		return // Nothing to fill
	}

	tcellStyle := style.ToTcell()
	screenWidth, screenHeight := screen.Size()

	// Calculate end coordinates (exclusive)
	endX := x + width
	endY := y + height

	// Clip coordinates to screen bounds for iteration start/end
	iterX := x
	iterY := y
	if iterX < 0 {
		iterX = 0
	}
	if iterY < 0 {
		iterY = 0
	}
	if endX > screenWidth {
		endX = screenWidth
	}
	if endY > screenHeight {
		endY = screenHeight
	}

	// Draw clipped rectangle
	for row := iterY; row < endY; row++ {
		for col := iterX; col < endX; col++ {
			// Double check inside loop for safety, though iter clipping should handle it
			// if col >= 0 && col < screenWidth && row >= 0 && row < screenHeight {
			screen.SetContent(col, row, char, nil, tcellStyle)
			// }
		}
	}
}

// drawGenericBox is a helper for drawing boxes with different border runes.
// Handles bounds checking and edge cases (1xN, Nx1, 1x1).
func drawGenericBox(screen tcell.Screen, x, y, width, height int, style Style,
	ul, ur, ll, lr, h, v rune) {

	if width <= 0 || height <= 0 {
		return // Cannot draw zero or negative sized box
	}

	tcellStyle := style.ToTcell()
	screenWidth, screenHeight := screen.Size()

	// Function to safely set content within screen bounds
	safeSet := func(px, py int, r rune) {
		if px >= 0 && px < screenWidth && py >= 0 && py < screenHeight {
			screen.SetContent(px, py, r, nil, tcellStyle)
		}
	}

	// Calculate corner coordinates
	x2 := x + width - 1
	y2 := y + height - 1

	// Draw Corners (handle 1x1, 1xN, Nx1 cases)
	if width == 1 && height == 1 {
		safeSet(x, y, ul) // Draw 1x1 box as just the top-left corner rune
		return            // No lines needed
	}
	if width >= 2 && height >= 2 {
		// Standard case: 2x2 or larger
		safeSet(x, y, ul)
		safeSet(x2, y, ur)
		safeSet(x, y2, ll)
		safeSet(x2, y2, lr)
	} else if width == 1 { // width = 1, height >= 2 (vertical line)
		safeSet(x, y, ul)  // Use top-left for top
		safeSet(x, y2, ll) // Use bottom-left for bottom
	} else { // height = 1, width >= 2 (horizontal line)
		safeSet(x, y, ul)  // Use top-left for left
		safeSet(x2, y, ur) // Use top-right for right
	}

	// Draw Horizontal Lines (if width >= 2)
	if width >= 2 {
		for i := x + 1; i < x2; i++ {
			safeSet(i, y, h) // Top line
			if height >= 2 { // Only draw bottom line if height allows
				safeSet(i, y2, h) // Bottom line
			}
		}
	}

	// Draw Vertical Lines (if height >= 2)
	if height >= 2 {
		for i := y + 1; i < y2; i++ {
			safeSet(x, i, v) // Left line
			if width >= 2 {  // Only draw right line if width allows
				safeSet(x2, i, v) // Right line
			}
		}
	}
}

// DrawBox draws a box with single-line borders using the specified style.
// Requires a minimum size of 1x1. Performs bounds checking.
func DrawBox(screen tcell.Screen, x, y, width, height int, style Style) {
	drawGenericBox(screen, x, y, width, height, style,
		RuneULCorner, RuneURCorner, RuneLLCorner, RuneLRCorner, RuneHLine, RuneVLine)
}

// DrawDoubleBox draws a box with double-line borders using the specified style.
// Requires a minimum size of 1x1. Performs bounds checking.
func DrawDoubleBox(screen tcell.Screen, x, y, width, height int, style Style) {
	drawGenericBox(screen, x, y, width, height, style,
		RuneDoubleULCorner, RuneDoubleURCorner, RuneDoubleLLCorner, RuneDoubleLRCorner, RuneDoubleHLine, RuneDoubleVLine)
}

// DrawSolidBox draws a box using block elements for a solid appearance.
// Handles smaller sizes gracefully. Performs bounds checking.
func DrawSolidBox(screen tcell.Screen, x, y, width, height int, style Style) {
	if width <= 0 || height <= 0 {
		return
	}

	tcellStyle := style.ToTcell()
	screenWidth, screenHeight := screen.Size()

	// Function to safely set content within screen bounds
	safeSet := func(px, py int, r rune) {
		if px >= 0 && px < screenWidth && py >= 0 && py < screenHeight {
			screen.SetContent(px, py, r, nil, tcellStyle)
		}
	}

	// Calculate end coordinates (exclusive) and bottom Y coordinate (inclusive)
	endX := x + width
	endY := y + height
	bottomY := endY - 1 // Inclusive Y of the bottom edge

	// Draw top edge (using upper half block)
	for i := x; i < endX; i++ {
		safeSet(i, y, RuneUpperHalfBlock)
	}

	// Draw bottom edge (using lower half block, only if height > 1)
	if height > 1 {
		for i := x; i < endX; i++ {
			safeSet(i, bottomY, RuneLowerHalfBlock)
		}
	}

	// Fill the middle rows with full blocks
	// Iterate from the row below the top edge up to (but not including) the bottom edge
	for row := y + 1; row < bottomY; row++ {
		// Draw left vertical side
		safeSet(x, row, RuneBlock)
		// Fill middle area (if width > 2)
		for col := x + 1; col < endX-1; col++ {
			safeSet(col, row, RuneBlock)
		}
		// Draw right vertical side (if width > 1)
		if width > 1 {
			safeSet(endX-1, row, RuneBlock)
		}
	}
}

// DrawText draws a string at the specified position using the given style.
// Handles wide characters and clips text at the screen boundary.
func DrawText(screen tcell.Screen, x, y int, style Style, text string) {
	// Basic bounds check for the starting position
	screenWidth, screenHeight := screen.Size()
	if y < 0 || y >= screenHeight || x >= screenWidth {
		return // Start position is completely off screen
	}

	tcellStyle := style.ToTcell()
	currentX := x

	for _, r := range text {
		// Stop if we go off the right edge
		if currentX >= screenWidth {
			break
		}

		runeWidth := runewidth.RuneWidth(r)
		// Draw if the starting position is within the screen width
		if currentX >= 0 {
			screen.SetContent(currentX, y, r, nil, tcellStyle)
			// If it's a wide rune, clear the next cell(s) it occupies
			// Make sure not to clear beyond the screen width
			for i := 1; i < runeWidth; i++ {
				if currentX+i < screenWidth {
					screen.SetContent(currentX+i, y, ' ', nil, tcellStyle)
				}
			}
		} else {
			// Rune starts off-screen left, but might partially appear if wide enough.
			// tcell handles clipping based on the starting cell, so we just need to advance correctly.
		}

		currentX += runeWidth // Advance by the rune's width
	}
}