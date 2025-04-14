// cmd/01_basic/main.go
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
	log.Println("Starting TinyTUI test app")

	// Create a new application
	app := tinytui.NewApplication()

	// Set theme
	app.SetTheme(tinytui.NewTurboTheme())

	// Create a horizontal layout
	layout := tinytui.NewLayout(tinytui.Horizontal)

	// Add a pane with a text component
	textPane := tinytui.NewPane()
	textPane.SetTitle("Text Test")

	text := tinytui.NewText("Hello, TinyTUI!")
	textPane.SetChild(text)

	layout.AddPane(textPane, tinytui.Size{Proportion: 1})

	// Set the layout as the application's main layout
	app.SetLayout(layout)

	// Register key handler for quitting with 'q'
	app.RegisterRuneHandler('q', tcell.ModNone, func() bool {
		app.Stop()
		return true
	})

	// Run the application
	if err := app.Run(); err != nil {
		log.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}

	log.Println("Application exited normally")
}