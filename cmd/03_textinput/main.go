// cmd/textinput/main.go
package main

import (
	"log"
	"os"

	"github.com/LixenWraith/tinytui"
)

func main() {
	// Initialize logger for debugging
	f, err := os.Create("textinput_debug.log")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.Println("Starting TinyTUI TextInput test app")

	// Create a new application
	app := tinytui.NewApplication()

	// Create a vertical layout to split the screen
	mainLayout := tinytui.NewLayout(tinytui.Vertical)

	// Create a horizontal layout for the top half
	topLayout := tinytui.NewLayout(tinytui.Horizontal)

	// ===== FIRST PANE: Input area =====
	inputPane := tinytui.NewPane()
	inputPane.SetTitle("Input")

	// Create a vertical layout for title and input
	inputLayout := tinytui.NewLayout(tinytui.Vertical)

	// Create title text
	inputTitle := tinytui.NewText("Enter text below and press Enter:")

	// Create a title pane
	titlePane := tinytui.NewPane()
	titlePane.SetChild(inputTitle)
	titlePane.SetBorder(tinytui.BorderNone, tinytui.DefaultStyle)

	// Create TextInput component
	textInput := tinytui.NewTextInput()

	// Create a pane for the text input
	textInputPane := tinytui.NewPane()
	textInputPane.SetChild(textInput)
	textInputPane.SetBorder(tinytui.BorderNone, tinytui.DefaultStyle)

	// Add components to input layout
	inputLayout.AddPane(titlePane, tinytui.Size{Proportion: 1})
	inputLayout.AddPane(textInputPane, tinytui.Size{Proportion: 1})

	// Set the layout as the pane's child
	inputPane.SetChild(inputLayout)

	// ===== SECOND PANE: Display area =====
	displayPane := tinytui.NewPane()
	displayPane.SetTitle("Display")

	// Create text display component
	displayText := tinytui.NewText("Input will appear here")

	// Set display text as the pane's child
	displayPane.SetChild(displayText)

	// Set up handler for text input submission
	textInput.SetOnSubmit(func(text string) {
		// Update display text with submitted input
		displayText.SetContent("You entered: " + text)

		// Keep focus on text input
		app.SetFocus(textInput)

		// Clear the input field
		textInput.SetText("")
	})

	// Add panes to top layout
	topLayout.AddPane(inputPane, tinytui.Size{Proportion: 1})
	topLayout.AddPane(displayPane, tinytui.Size{Proportion: 1})

	// Create top container pane
	topPane := tinytui.NewPane()
	topPane.SetChild(topLayout)
	topPane.SetBorder(tinytui.BorderNone, tinytui.DefaultStyle)

	// ===== THIRD PANE: Clear button =====
	clearPane := tinytui.NewPane()
	clearPane.SetTitle("Controls")

	// Create a 1x1 grid for the clear button
	clearGrid := tinytui.NewGrid()
	clearCells := [][]string{{"CLEAR"}}
	clearGrid.SetCells(clearCells)
	clearGrid.SetCellSize(8, 1)

	// Set handler for clear button
	clearGrid.SetOnSelect(func(row, col int, item string) {
		log.Println("Clear button pressed")
		displayText.SetContent("Input will appear here")

		// Set focus back to text input
		app.SetFocus(textInput)
	})

	clearPane.SetChild(clearGrid)

	// Add layouts to main layout
	mainLayout.AddPane(topPane, tinytui.Size{Proportion: 3})
	mainLayout.AddPane(clearPane, tinytui.Size{Proportion: 1})

	// Set the layout as the application's main layout
	app.SetLayout(mainLayout)

	// Focus the text input component initially
	app.SetFocus(textInput)

	// Run the application
	if err := app.Run(); err != nil {
		log.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}

	log.Println("Application exited normally")
}