// draw.go
package tinytui // Package tinytui provides basic TUI building blocks.

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// Define necessary box-drawing runes not available as tcell constants
const (
	// Double Line Box Drawing Runes
	runeDoubleULCorner rune = '╔' // U+2554
	runeDoubleURCorner rune = '╗' // U+2557
	runeDoubleLLCorner rune = '╚' // U+255A
	runeDoubleLRCorner rune = '╝' // U+255D
	runeDoubleHLine    rune = '═' // U+2550
	runeDoubleVLine    rune = '║' // U+2551

	// Block Element Runes (Half Blocks)
	runeUpperHalfBlock rune = '▀' // U+2580
	runeLowerHalfBlock rune = '▄' // U+2584
	// tcell.RuneBlock ('█' U+2588) is already available
)

// Fill fills a rectangular area with the specified rune and style.
// It accepts tinytui.Style.
func Fill(screen tcell.Screen, x, y, width, height int, char rune, style Style) {
	if width <= 0 || height <= 0 {
		return
	}
	tcellStyle := style.ToTcell() // Convert to underlying tcell type
	for row := y; row < y+height; row++ {
		for col := x; col < x+width; col++ {
			// Basic bounds check for safety, although tcell might handle it
			sw, sh := screen.Size()
			if col >= 0 && col < sw && row >= 0 && row < sh {
				screen.SetContent(col, row, char, nil, tcellStyle)
			}
		}
	}
}

// DrawBox draws a box with single lines using the specified style.
// It accepts tinytui.Style. Requires width and height >= 2.
func DrawBox(screen tcell.Screen, x, y, width, height int, style Style) {
	if width < 2 || height < 2 {
		return
	}
	tcellStyle := style.ToTcell() // Convert to underlying tcell type

	// Draw corners
	screen.SetContent(x, y, tcell.RuneULCorner, nil, tcellStyle)
	screen.SetContent(x+width-1, y, tcell.RuneURCorner, nil, tcellStyle)
	screen.SetContent(x, y+height-1, tcell.RuneLLCorner, nil, tcellStyle)
	screen.SetContent(x+width-1, y+height-1, tcell.RuneLRCorner, nil, tcellStyle)

	// Draw horizontal lines
	for col := x + 1; col < x+width-1; col++ {
		screen.SetContent(col, y, tcell.RuneHLine, nil, tcellStyle)
		screen.SetContent(col, y+height-1, tcell.RuneHLine, nil, tcellStyle)
	}

	// Draw vertical lines
	for row := y + 1; row < y+height-1; row++ {
		screen.SetContent(x, row, tcell.RuneVLine, nil, tcellStyle)
		screen.SetContent(x+width-1, row, tcell.RuneVLine, nil, tcellStyle)
	}
}

// DrawDoubleBox draws a box with double lines using the specified style.
// It requires width and height to be at least 2.
func DrawDoubleBox(screen tcell.Screen, x, y, width, height int, style Style) {
	if width < 2 || height < 2 {
		return // Not enough space to draw a box
	}
	tcellStyle := style.ToTcell() // Convert to underlying tcell type

	// Draw corners using defined runes
	screen.SetContent(x, y, runeDoubleULCorner, nil, tcellStyle)
	screen.SetContent(x+width-1, y, runeDoubleURCorner, nil, tcellStyle)
	screen.SetContent(x, y+height-1, runeDoubleLLCorner, nil, tcellStyle)
	screen.SetContent(x+width-1, y+height-1, runeDoubleLRCorner, nil, tcellStyle)

	// Draw horizontal lines using defined runes
	for col := x + 1; col < x+width-1; col++ {
		screen.SetContent(col, y, runeDoubleHLine, nil, tcellStyle)
		screen.SetContent(col, y+height-1, runeDoubleHLine, nil, tcellStyle)
	}

	// Draw vertical lines using defined runes
	for row := y + 1; row < y+height-1; row++ {
		screen.SetContent(x, row, runeDoubleVLine, nil, tcellStyle)
		screen.SetContent(x+width-1, row, runeDoubleVLine, nil, tcellStyle)
	}
}

// DrawSolidBox draws a box using block elements for sides and half-blocks for top/bottom.
// This aims for visually thinner horizontal lines compared to vertical ones.
// Requires width and height >= 2 for proper rendering.
func DrawSolidBox(screen tcell.Screen, x, y, width, height int, style Style) {
	// Fallback to full block fill if dimensions are too small for the half-block style
	if width < 2 || height < 2 {
		if width > 0 && height > 0 {
			Fill(screen, x, y, width, height, tcell.RuneBlock, style)
		}
		return
	}

	tcellStyle := style.ToTcell() // Convert to underlying tcell type

	// Draw top and bottom lines using defined runes
	// Skipping side columns to be filled with full blocks
	for col := x + 1; col < x+width-1; col++ {
		// Top line
		if height == 1 {
			// If height is 1, use full block for the single line
			screen.SetContent(col, y, tcell.RuneBlock, nil, tcellStyle)
		} else {
			screen.SetContent(col, y, runeUpperHalfBlock, nil, tcellStyle)
		}

		// Bottom line (only if height > 1)
		if height > 1 {
			screen.SetContent(col, y+height-1, runeLowerHalfBlock, nil, tcellStyle)
		}
	}

	// Draw side lines always to ensure full width and full block corners
	if height > 1 {
		for row := y; row < y+height; row++ {
			// Left line
			if width > 0 {
				screen.SetContent(x, row, tcell.RuneBlock, nil, tcellStyle)
			}
			// Right line (only if width > 1)
			if width > 1 {
				screen.SetContent(x+width-1, row, tcell.RuneBlock, nil, tcellStyle)
			}
		}
	}
}

// DrawText draws a string at the specified position using the given style, clipping at screen bounds.
// It accepts tinytui.Style. Handles rune width.
func DrawText(screen tcell.Screen, x, y int, style Style, text string) {
	col := x
	tcellStyle := style.ToTcell() // Convert to underlying tcell type
	sw, sh := screen.Size()

	if y < 0 || y >= sh { // Don't draw if row is outside screen
		return
	}

	for _, r := range text {
		rw := runewidth.RuneWidth(r)
		if col+rw <= 0 { // Skip characters that are entirely to the left
			col += rw
			continue
		}
		if col >= sw { // Stop if we are entirely past the right edge
			break
		}
		// Clip character if it starts before x=0
		if col < 0 {
			// This is complex to handle correctly for wide runes, skip partial drawing for now
			col += rw
			continue
		}
		screen.SetContent(col, y, r, nil, tcellStyle)
		col += rw
	}
}