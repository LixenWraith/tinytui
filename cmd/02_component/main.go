// cmd/tictactoe/main.go
package main

import (
	"log"
	"os"

	"github.com/gdamore/tcell/v2"

	"github.com/LixenWraith/tinytui"
)

func main() {
	// Initialize logger for debugging
	f, err := os.Create("debug.log")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.Println("Starting TinyTUI TicTacToe test app")

	// Create a new application
	app := tinytui.NewApplication()

	// Create a vertical layout to split the screen in half
	mainLayout := tinytui.NewLayout(tinytui.Vertical)

	// Create a horizontal layout for the top half
	topLayout := tinytui.NewLayout(tinytui.Horizontal)

	// ===== FIRST PANE: Game area with title and grid =====
	gamePane := tinytui.NewPane()
	gamePane.SetTitle("Tic-Tac-Toe Game")
	gamePane.SetIndex(1) // Set as first pane (Alt+1)

	// Create a vertical layout inside the game pane to hold text and grid
	gameLayout := tinytui.NewLayout(tinytui.Vertical)

	// Create a text component for the game title
	gameTitle := tinytui.NewText("Welcome to Tic-Tac-Toe!\nPress Enter to toggle X/O on a cell.")

	// Create title pane inside the game layout
	titlePane := tinytui.NewPane()
	titlePane.SetChild(gameTitle)
	titlePane.SetBorder(tinytui.BorderNone, tinytui.DefaultStyle) // No border for title

	// Create the game grid
	grid := tinytui.NewGrid()

	// Set up a 3x3 grid for tic-tac-toe with centered content
	cells := [][]string{
		{" [ ] ", " [ ] ", " [ ] "},
		{" [ ] ", " [ ] ", " [ ] "},
		{" [ ] ", " [ ] ", " [ ] "},
	}

	grid.SetCells(cells)
	grid.SetCellSize(6, 1) // Set height to 1 per request
	grid.SetAutoWidth(true)
	grid.SetPadding(1) // Add padding around cell content

	// Set handler for cell selection (Enter key)
	grid.SetOnSelect(func(row, col int, item string) {
		log.Printf("Selected cell: %d,%d = %s\n", row, col, item)

		// Toggle cell content between [ ] and [X] or [O]
		var newValue string
		if item == " [ ] " {
			newValue = " [X] "
		} else if item == " [X] " {
			newValue = " [O] "
		} else {
			newValue = " [ ] "
		}

		// Update the cells array
		cells[row][col] = newValue
		grid.SetCells(cells)
	})

	// Create grid pane inside the game layout
	gridPane := tinytui.NewPane()
	gridPane.SetChild(grid)
	gridPane.SetBorder(tinytui.BorderNone, tinytui.DefaultStyle) // No border for inner grid

	// Add components to game layout
	gameLayout.AddPane(titlePane, tinytui.Size{Proportion: 1})
	gameLayout.AddPane(gridPane, tinytui.Size{Proportion: 4})

	// Set the layout as the pane's child
	gamePane.SetChild(gameLayout)

	// ===== SECOND PANE: Instructions =====
	instructionsPane := tinytui.NewPane()
	instructionsPane.SetTitle("Instructions")
	instructionsPane.SetIndex(2) // Set as second pane (Alt+2)

	instructions := tinytui.NewText(
		"TicTacToe Test App\n\n" +
			"Controls:\n" +
			"- Arrow keys or h,j,k,l to navigate\n" +
			"- Enter to toggle cell\n" +
			"- Alt+1/2/3 to switch panes\n" +
			"- Tab to cycle focus\n" +
			"- ESC to exit\n" +
			"- Ctrl+C to force exit")

	instructionsPane.SetChild(instructions)

	// Add panes to top layout
	topLayout.AddPane(gamePane, tinytui.Size{Proportion: 2})
	topLayout.AddPane(instructionsPane, tinytui.Size{Proportion: 1})

	// Create a pane to hold the top layout
	topPane := tinytui.NewPane()
	topPane.SetChild(topLayout)
	topPane.SetBorder(tinytui.BorderNone, tinytui.DefaultStyle) // No border for this container

	// ===== THIRD PANE: Exit button =====
	exitPane := tinytui.NewPane()
	exitPane.SetTitle("Exit Controls")
	exitPane.SetIndex(3) // Set as third pane (Alt+3)

	// Create a 1x1 grid for the exit button
	exitGrid := tinytui.NewGrid()
	exitCells := [][]string{{"EXIT"}}
	exitGrid.SetCells(exitCells)
	exitGrid.SetCellSize(8, 1)

	// Set handler for exit button
	exitGrid.SetOnSelect(func(row, col int, item string) {
		log.Println("Exit button selected, closing application")
		app.Stop()
	})

	exitPane.SetChild(exitGrid)

	// Add layouts to main layout
	mainLayout.AddPane(topPane, tinytui.Size{Proportion: 3})
	mainLayout.AddPane(exitPane, tinytui.Size{Proportion: 1})

	// Set the layout as the application's main layout
	app.SetLayout(mainLayout)

	// Register key handler for quitting with ESC
	app.RegisterKeyHandler(tcell.KeyEscape, tcell.ModNone, func() bool {
		app.Stop()
		return true
	})

	// Focus the grid component initially
	app.SetFocus(grid)

	// Run the application
	if err := app.Run(); err != nil {
		log.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}

	log.Println("Application exited normally")
}