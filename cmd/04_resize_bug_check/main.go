// cmd/04_resize_bug_check/main.go
package main

import (
	"fmt"
	"log"
	"os"
	// "time" // Not needed for this minimal example

	"github.com/LixenWraith/tinytui"
)

func main() {
	// Setup logging
	logFile, err := os.OpenFile("tinytui_minimal_test.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.Println("--- Starting Minimal TinyTUI Test ---")
	log.Println("Instructions: Run and resize. Check log for SIZE MISMATCH errors.")
	log.Println("Press Esc or Ctrl+C to quit.")

	// --- Application Setup ---
	app := tinytui.NewApplication()
	app.SetShowPaneIndices(true)

	// --- Create Components ---
	topText := tinytui.NewText("Top Pane (Fixed: 5). Resize window.")
	middleInput := tinytui.NewTextInput()
	middleInput.SetText("Middle Pane (Proportional). Focusable.")
	bottomText := tinytui.NewText("Bottom Pane (Fixed: 5)")

	// --- Create Panes ---
	topPane := tinytui.NewPane()
	topPane.SetTitle("Top [1]")
	topPane.SetChild(topText) // Not focusable -> ---

	middlePane := tinytui.NewPane()
	middlePane.SetTitle("Middle [2]")
	middlePane.SetChild(middleInput) // Focusable -> [2] or [#]

	bottomPane := tinytui.NewPane()
	bottomPane.SetTitle("Bottom [3]")
	bottomPane.SetChild(bottomText) // Not focusable -> ---

	// --- Setup Layout ---
	mainLayout := tinytui.NewLayout(tinytui.Vertical)
	mainLayout.SetGap(1) // Gap makes border issues obvious

	mainLayout.AddPane(topPane, tinytui.Size{FixedSize: 5})
	mainLayout.AddPane(middlePane, tinytui.Size{Proportion: 1})
	mainLayout.AddPane(bottomPane, tinytui.Size{FixedSize: 5})

	// --- Set Application Layout ---
	app.SetLayout(mainLayout) // Triggers index assignment

	// --- Set Initial Focus ---
	app.Dispatch(&tinytui.FocusCommand{Target: middleInput})

	// --- Run Application ---
	log.Println("Running application...")
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		log.Fatalf("Error running application: %v", err)
		os.Exit(1)
	}
	log.Println("Application exited normally.")
}