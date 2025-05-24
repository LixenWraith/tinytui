package main

import (
	"github.com/LixenWraith/tinytui"
)

func main() {
	// Create application
	app := tinytui.NewApplication()

	// Create components
	header := tinytui.NewText("TinyTUI Example")
	header.SetAlignment(tinytui.AlignTextCenter)

	input := tinytui.NewTextInput()
	input.SetText("Enter text")

	button := tinytui.NewGrid()
	button.SetCells([][]string{{" Submit "}}) // Using a Grid as a simple button
	button.SetCellSize(10, 1)                 // Define the button's size
	button.SetOnSelect(func(r, c int, i string) {
		// Handle button click/selection
		// For a real app, you might close a dialog, submit data, etc.
		// For this example, we can just stop the app or print something.
		app.Stop() // Example action: stop the app when button is "clicked"
	})

	// Create panes and set content
	headerPane := tinytui.NewPane()
	headerPane.SetChild(header)

	inputPane := tinytui.NewPane()
	inputPane.SetTitle("Input") // Give the pane a title
	inputPane.SetChild(input)

	buttonPane := tinytui.NewPane()
	buttonPane.SetChild(button)

	// Create layout and arrange panes
	// This creates a vertical layout: header, then input, then button
	layout := tinytui.NewLayout(tinytui.Vertical)
	layout.AddPane(headerPane, tinytui.Size{FixedSize: 1}) // Header takes 1 row
	layout.AddPane(inputPane, tinytui.Size{Proportion: 1}) // Input takes remaining proportional space
	layout.AddPane(buttonPane, tinytui.Size{FixedSize: 1}) // Button takes 1 row

	// Set application layout
	app.SetLayout(layout)

	// Set initial focus to the input field
	app.Dispatch(&tinytui.FocusCommand{Target: input})

	// Run application
	if err := app.Run(); err != nil {
		panic(err) // Or handle error more gracefully
	}
}