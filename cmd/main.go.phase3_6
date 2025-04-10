// main.go
package main

import (
	"fmt"
	"log"
	"os"
	"time" // Import time for potential delay

	"github.com/LixenWraith/tinytui"
	"github.com/LixenWraith/tinytui/widgets"
	// tcell needed only if using tcell types directly, which we avoid now
)

func main() {
	// --- Basic Logging Setup ---
	logFile, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	// Handle potential error on close
	defer func() {
		err := logFile.Close()
		if err != nil {
			log.Printf("Warning: Failed to close log file: %v", err)
		}
	}()
	log.SetOutput(logFile)
	log.Println("--- Application Starting ---")

	// --- Create Application ---
	app := tinytui.NewApplication() // Keep app instance accessible
	log.Println("main.go: Application created.")

	// --- Create Widgets ---
	log.Println("main.go: Creating widgets...")

	// 1. Title
	titleText := widgets.NewText("TinyTUI Popup Demo (Esc to Quit)").
		SetStyle(tinytui.DefaultStyle.Foreground(tinytui.ColorAqua).Bold(true))

	// 2. Grid Widget
	grid := widgets.NewGrid()
	grid.SetCellSize(18, 1) // Cell width=18

	// Create grid data...
	gridData := make([][]string, 10)
	for r := 0; r < 10; r++ {
		gridData[r] = make([]string, 5)
		for c := 0; c < 5; c++ {
			if r%2 == 0 && c%2 == 0 {
				gridData[r][c] = fmt.Sprintf("Selectable R%dC%d", r, c)
			} else {
				gridData[r][c] = fmt.Sprintf("Item R%dC%d", r, c)
			}
		}
	}
	grid.SetCells(gridData)
	log.Println("main.go: Grid created and data set.")

	// --- Popup Widgets (Initially Hidden) ---
	log.Println("main.go: Creating popup widgets...")
	// Popup Message
	popupMessage := widgets.NewText("Popup Message Here!")

	// Popup Sprite (define its data)...
	spriteData := [][]widgets.SpriteCell{
		{{Rune: '/', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorRed).Background(tinytui.ColorBlack)}, {Rune: '-', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorRed).Background(tinytui.ColorBlack)}, {Rune: '\\', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorRed).Background(tinytui.ColorBlack)}},
		{{Rune: '|', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorYellow).Background(tinytui.ColorBlack)}, {Rune: 'X', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorWhite).Background(tinytui.ColorBlack).Bold(true)}, {Rune: '|', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorYellow).Background(tinytui.ColorBlack)}},
		{{Rune: ' ', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorBlue)}, {Rune: 'O', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorBlue)}, {Rune: ' ', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorBlue)}},
		{{Rune: '\\', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorGreen).Background(tinytui.ColorBlack)}, {Rune: '_', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorGreen).Background(tinytui.ColorBlack)}, {Rune: '/', Style: tinytui.DefaultStyle.Foreground(tinytui.ColorGreen).Background(tinytui.ColorBlack)}},
	}
	popupSprite := widgets.NewSprite(spriteData)
	popupSprite.SetVisible(false) // Sprite is initially hidden

	// Popup Button to show sprite
	showSpriteButton := widgets.NewButton("Show Sprite")

	// Popup Button to close
	closePopupButton := widgets.NewButton("Close Popup")

	// Popup Content Layout (Horizontal: Message | Show Button | Close Button | Sprite Area)
	popupContentLayout := tinytui.NewFlexLayout(tinytui.Horizontal).SetGap(2).
		AddChild(popupMessage, 0, 1).                 // Message takes proportional space
		AddChild(showSpriteButton, 15, 0).            // Give buttons fixed width
		AddChild(closePopupButton, 15, 0).            // Give buttons fixed width
		AddChild(popupSprite, popupSprite.Width(), 0) // Sprite fixed width based on data

	// --- Create Pane for Popup ---
	popupPane := widgets.NewPane()
	popupPane.SetBorder(true, tinytui.BorderDouble, tinytui.DefaultStyle.Foreground(tinytui.ColorAqua))
	popupPane.SetChild(popupContentLayout) // Use SetChild to put the layout inside the pane
	popupPane.SetVisible(false)            // Pane is initially hidden
	log.Println("main.go: Popup widgets created.")

	// --- Define Interactions (Dispatch Closures) ---
	log.Println("main.go: Defining interactions...")
	// Grid selection handler
	grid.SetOnSelect(func(row, col int, item string) {
		log.Printf("Grid item selected: Row %d, Col %d, Item '%s'\n", row, col, item)
		if len(item) > 10 && item[:10] == "Selectable" {
			// --- Capture variables needed for closures ---
			msgWidget := popupMessage
			newContent := fmt.Sprintf("Selected '%s'", item)
			spriteWidget := popupSprite
			paneWidget := popupPane
			focusTarget := showSpriteButton
			// --- End Capture ---

			// Dispatch function to update text
			app.Dispatch(func(app *tinytui.Application) {
				// Direct assignment works if *widgets.Text implements TextUpdater
				var updater tinytui.TextUpdater = msgWidget
				if updater != nil {
					updater.SetContent(newContent) // Call interface method
				} else { // Should not happen if Text implements TextUpdater
					log.Printf("Error: msgWidget (*widgets.Text) does not implement TextUpdater?")
				}
			})

			// Dispatch function to hide sprite
			app.Dispatch(func(app *tinytui.Application) {
				spriteWidget.SetVisible(false)
			})

			// Dispatch function to show modal pane and set focus
			app.Dispatch(func(app *tinytui.Application) {
				paneWidget.SetVisible(true)
				app.SetModalRoot(paneWidget) // Use method to set modal root
				app.SetFocus(focusTarget)    // Set focus
			})

			log.Println("Dispatched actions to show popup")
		} else {
			log.Println("Non-selectable item chosen.")
		}
	})

	// Popup button handler (Show Sprite)
	showSpriteButton.SetOnClick(func() {
		log.Println("Show Sprite button clicked")
		spriteWidget := popupSprite
		app.Dispatch(func(app *tinytui.Application) {
			spriteWidget.SetVisible(true)
		})
	})

	// Popup button handler (Close)
	closePopupButton.SetOnClick(func() {
		log.Println("Close Popup button clicked")
		returnTarget := grid
		paneWidget := popupPane
		app.Dispatch(func(app *tinytui.Application) {
			log.Println("Action: Dispatching ClearModalRoot via Button")
			app.ClearModalRoot()
			if paneWidget != nil {
				paneWidget.SetVisible(false)
			}
			app.SetFocus(returnTarget)
		})
	})
	log.Println("main.go: Interactions defined.")

	// --- Main Layout ---
	log.Println("main.go: Creating main layout...")
	// Vertical: Title, Grid, Popup Area (Pane is initially invisible)
	layout := tinytui.NewFlexLayout(tinytui.Vertical).
		AddChild(titleText, 1, 0). // Fixed height 1
		AddChild(grid, 0, 1).      // Grid takes proportional space
		AddChild(popupPane, 7, 0)  // Add Pane, give it fixed height (adjust as needed)
	log.Println("main.go: Main layout created.")

	// --- Set Root and Run ---
	log.Println("main.go: Calling app.SetRoot()...")
	app.SetRoot(layout, true)
	log.Println("main.go: app.SetRoot() finished.")

	// --- Remove Manual Focus Test ---
	// log.Println("main.go: Manually calling SetFocus on grid before Run()")
	// app.SetFocus(grid)
	// --- End Remove Manual Focus Test ---

	// --- Add a small delay and log focus state *before* Run ---
	// This is a HACK to see if the initial focus finding in SetRoot worked,
	// but it relies on timing and isn't guaranteed to run before Run's internal dispatch.
	log.Println("main.go: Pausing briefly before Run...")
	time.Sleep(100 * time.Millisecond) // Small delay
	// We cannot directly access app.focused here as it's unexported and protected.
	// We could add a public GetFocused() method to Application for debugging,
	// but sticking to logging *within* main.go for now.
	log.Println("main.go: Pause finished. Calling app.Run()...")
	// --- End Delay ---

	if err := app.Run(); err != nil {
		log.Fatalf("Error running application: %v", err)
	}

	log.Println("--- Application Stopped ---")
}