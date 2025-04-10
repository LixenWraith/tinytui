// main.go - Enhanced demo application
package main

import (
	"fmt"
	"os"

	"github.com/LixenWraith/tinytui"
	"github.com/LixenWraith/tinytui/widgets"
)

func main() {
	// --- Create Application ---
	app := tinytui.NewApplication()

	// --- Create Header ---
	headerText := widgets.NewText("✨ TinyTUI Interactive Demo ✨").
		SetStyle(tinytui.DefaultStyle.
			Foreground(tinytui.ColorAqua).
			Background(tinytui.ColorNavy).
			Bold(true))

	// --- Create Footer ---
	footerText := widgets.NewText("Press Tab to navigate | Space/Enter to select | Esc to exit").
		SetStyle(tinytui.DefaultStyle.
			Foreground(tinytui.ColorSilver).
			Background(tinytui.ColorDarkBlue))

	// --- Create Grid Widget with prettier, smaller grid ---
	grid := widgets.NewGrid()
	grid.SetCellSize(16, 1) // Smaller cells

	// // Color palette for grid items
	// colorPalette := []tinytui.Color{
	// 	tinytui.ColorLightGray,
	// 	tinytui.ColorLightCyan,
	// 	tinytui.ColorLightGreen,
	// 	tinytui.ColorLightYellow,
	// }

	// Create grid data with more interesting names and fewer items
	gridData := make([][]string, 5) // Smaller grid
	for r := 0; r < 5; r++ {
		gridData[r] = make([]string, 4)
		for c := 0; c < 4; c++ {
			if r%2 == 0 && c%2 == 0 {
				// Make selectable items more distinct
				gridData[r][c] = fmt.Sprintf("🔹 Option %d-%d", r+1, c+1)
			} else {
				gridData[r][c] = fmt.Sprintf("⚪ Item %d-%d", r+1, c+1)
			}
		}
	}
	grid.SetCells(gridData)

	// Customize grid styles
	grid.SetStyle(tinytui.DefaultStyle.Background(tinytui.ColorDarkBlue))
	grid.SetSelectedStyle(tinytui.DefaultStyle.
		Background(tinytui.ColorNavy).
		Foreground(tinytui.ColorWhite).
		Bold(true))

	// --- Create Info Panel (right side) ---
	infoTitle := widgets.NewText("Item Details").
		SetStyle(tinytui.DefaultStyle.
			Foreground(tinytui.ColorAqua).
			Bold(true))

	infoContent := widgets.NewText("Select an item from the grid to view details").
		SetWrap(true).
		SetStyle(tinytui.DefaultStyle.
			Foreground(tinytui.ColorSilver))

	// Container for info content
	infoPane := widgets.NewPane()
	infoPane.SetBorder(true, tinytui.BorderSingle,
		tinytui.DefaultStyle.Foreground(tinytui.ColorGray))

	// Stack title and content
	infoLayout := tinytui.NewFlexLayout(tinytui.Vertical).
		SetGap(1).
		AddChild(infoTitle, 1, 0).
		AddChild(infoContent, 0, 1)

	infoPane.SetChild(infoLayout)

	// --- Create Popup Content ---
	// Popup title with emoji and styled text
	popupTitle := widgets.NewText("🌟 Interactive Options 🌟").
		SetStyle(tinytui.DefaultStyle.
			Foreground(tinytui.ColorYellow).
			Bold(true))

	// Popup message with better styling
	popupMessage := widgets.NewText("Select an action:").
		SetStyle(tinytui.DefaultStyle.
			Foreground(tinytui.ColorSilver))

	// Sprite for visual interest
	spriteData := [][]widgets.SpriteCell{
		{
			{Rune: '╔', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorAqua)},
			{Rune: '═', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorAqua)},
			{Rune: '╗', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorAqua)},
		},
		{
			{Rune: '║', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorAqua)},
			{Rune: '◉', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorRed).Bold(true)},
			{Rune: '║', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorAqua)},
		},
		{
			{Rune: '╚', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorAqua)},
			{Rune: '═', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorAqua)},
			{Rune: '╝', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorAqua)},
		},
	}

	popupSprite := widgets.NewSprite(spriteData)
	popupSprite.SetVisible(false) // Initially hidden

	// Create buttons with improved styling
	showSpriteButton := widgets.NewButton("📊 Show Icon")
	showSpriteButton.SetStyle(tinytui.DefaultStyle.
		Foreground(tinytui.ColorWhite).
		Background(tinytui.ColorNavy))
	showSpriteButton.SetFocusedStyle(tinytui.DefaultStyle.
		Foreground(tinytui.ColorBlack).
		Background(tinytui.ColorAqua).
		Bold(true))

	closePopupButton := widgets.NewButton("✖ Close")
	closePopupButton.SetStyle(tinytui.DefaultStyle.
		Foreground(tinytui.ColorWhite).
		Background(tinytui.ColorMaroon))
	closePopupButton.SetFocusedStyle(tinytui.DefaultStyle.
		Foreground(tinytui.ColorWhite).
		Background(tinytui.ColorRed).
		Bold(true))

	// Create a button layout with better spacing and alignment
	buttonsLayout := tinytui.NewFlexLayout(tinytui.Horizontal).
		SetGap(2).
		SetMainAxisAlignment(tinytui.AlignCenter).
		SetCrossAxisAlignment(tinytui.AlignCenter).
		AddChildWithAlign(showSpriteButton, 0, 0, tinytui.AlignCenter).
		AddChildWithAlign(closePopupButton, 0, 0, tinytui.AlignCenter)

	// Arrange sprite in its own centered container
	spriteLayout := tinytui.NewFlexLayout(tinytui.Vertical).
		SetMainAxisAlignment(tinytui.AlignCenter).
		SetCrossAxisAlignment(tinytui.AlignCenter).
		AddChildWithAlign(popupSprite, 0, 0, tinytui.AlignCenter)

	// Main popup content layout
	popupContentLayout := tinytui.NewFlexLayout(tinytui.Vertical).
		SetGap(1).
		AddChild(popupTitle, 1, 0).
		AddChild(popupMessage, 1, 0).
		AddChild(buttonsLayout, 2, 0).
		AddChild(spriteLayout, 3, 0)

	// Create nicer popup pane
	popupPane := widgets.NewPane()
	popupPane.SetStyle(tinytui.DefaultStyle.Background(tinytui.ColorDarkBlue))
	popupPane.SetBorder(true, tinytui.BorderDouble,
		tinytui.DefaultStyle.Foreground(tinytui.ColorAqua).Bold(true))
	popupPane.SetChild(popupContentLayout)
	popupPane.SetVisible(false) // Initially hidden

	// --- Define Interactions ---

	// Grid selection handler
	grid.SetOnSelect(func(row, col int, item string) {
		// Update info panel
		infoContent.SetContent(fmt.Sprintf(
			"Selected: %s\n\nPosition: Row %d, Column %d\n\nThis item %s be used with the interactive options.",
			item, row+1, col+1,
			map[bool]string{true: "can", false: "cannot"}[row%2 == 0 && col%2 == 0],
		))

		// Only show popup for "Option" items
		if row%2 == 0 && col%2 == 0 {
			// Update popup message
			popupMessage.SetContent(fmt.Sprintf("You selected: %s", item))

			// Hide sprite
			popupSprite.SetVisible(false)

			// Show popup and set focus
			app.Dispatch(func(app *tinytui.Application) {
				popupPane.SetVisible(true)
				app.SetModalRoot(popupPane)
				app.SetFocus(showSpriteButton)
			})
		}
	})

	// Popup button handlers
	showSpriteButton.SetOnClick(func() {
		app.Dispatch(func(app *tinytui.Application) {
			popupSprite.SetVisible(true)
		})
	})

	closePopupButton.SetOnClick(func() {
		app.Dispatch(func(app *tinytui.Application) {
			app.ClearModalRoot()
			popupPane.SetVisible(false)
		})
	})

	// --- Create Main Layout ---

	// Create split main content area (grid + info panel side by side)
	contentLayout := tinytui.NewFlexLayout(tinytui.Horizontal).
		SetGap(1).
		AddChild(grid, 0, 3).    // Grid takes 3/4 of space
		AddChild(infoPane, 0, 1) // Info takes 1/4 of space

	// Main vertical layout
	mainLayout := tinytui.NewFlexLayout(tinytui.Vertical).
		AddChild(headerText, 1, 0).    // Fixed height header
		AddChild(contentLayout, 0, 1). // Content takes all available space
		AddChild(footerText, 1, 0).    // Fixed height footer
		AddChild(popupPane, 10, 0)     // Fixed height for popup

	// --- Set Root and Run ---
	app.SetRoot(mainLayout, true)

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}