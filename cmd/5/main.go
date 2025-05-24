// main_test_indexing.go
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/LixenWraith/tinytui"
)

var appInstance *tinytui.Application

// Simple logger for testing focus/interaction
func appLog(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Println(msg) // Log to file only for simplicity
}

// Helper to create a simple 1x1 focusable grid
func createFocusableGrid(label string) *tinytui.Grid {
	grid := tinytui.NewGrid()
	grid.SetCells([][]string{{label}})
	grid.SetCellSize(len(label)+2, 1) // Auto-size roughly
	grid.SetSelectionMode(tinytui.SingleSelect)
	grid.SetOnSelect(func(r, c int, i string) {
		appLog("Selected Grid: %s", label)
		grid.SetCellInteracted(r, c, true) // Show interaction
		time.AfterFunc(150*time.Millisecond, func() {
			currentApp := appInstance
			if currentApp != nil {
				currentApp.Dispatch(&tinytui.SimpleCommand{
					Func: func(a *tinytui.Application) { grid.SetCellInteracted(r, c, false) }, // Clear interaction
				})
			}
		})
	})
	return grid
}

func main() {
	// --- Logging Setup ---
	logFile, err := os.OpenFile("tinytui_indexing_test.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.Println("--- Starting TinyTUI Indexing Test ---")

	// --- Application Setup ---
	app := tinytui.NewApplication()
	appInstance = app
	app.SetClearScreenOnExit(true)
	// app.SetShowPaneIndices(true) // Ensure it's enabled (default is true)

	// --- Create Components ---
	// Focusable elements
	gridA := createFocusableGrid("Grid A")
	gridNested := createFocusableGrid("Nested Grid")

	// Non-focusable elements
	textB := tinytui.NewText("Pane B\n(No focus)")
	textNested := tinytui.NewText("Nested Text")
	headerText := tinytui.NewText("Indexing Test - [Alt+1/3] Focus - [Esc] Quit") // Updated help text
	headerText.SetAlignment(tinytui.AlignTextCenter)

	// --- Create Panes ---
	// Top Level Panes (Will be added directly to Root Layout)
	headerPane := tinytui.NewPane()
	headerPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle())
	headerPane.SetChild(headerText)

	paneA := tinytui.NewPane()
	paneA.SetTitle("Pane A (Focusable)")
	paneA.SetChild(gridA)

	paneB := tinytui.NewPane()
	paneB.SetTitle("Pane B (Not Focusable)")
	paneB.SetChild(textB)

	paneC := tinytui.NewPane()
	paneC.SetTitle("Pane C (Nested Focusable)")
	// Child (nestedLayout) will be set below

	// Nested Panes (Inside Pane C's Layout) - NO INDICES EXPECTED
	nestedPaneText := tinytui.NewPane()
	nestedPaneText.SetTitle("Nested Text Pane") // Title to verify nesting visually
	nestedPaneText.SetChild(textNested)

	nestedPaneGrid := tinytui.NewPane()
	nestedPaneGrid.SetTitle("Nested Grid Pane") // Title to verify nesting visually
	nestedPaneGrid.SetChild(gridNested)

	// --- Setup Layouts ---

	// Nested Layout (Inside Pane C)
	nestedLayout := tinytui.NewLayout(tinytui.Vertical)
	nestedLayout.SetGap(1)
	nestedLayout.AddPane(nestedPaneText, tinytui.Size{FixedSize: 3})
	nestedLayout.AddPane(nestedPaneGrid, tinytui.Size{Proportion: 1})
	// Now set this layout as the child of Pane C
	paneC.SetChild(nestedLayout)

	// *** REMOVED middleLayout and middleWrapperPane ***

	// Root Layout (Vertical) - Add panes A, B, C directly
	rootLayout := tinytui.NewLayout(tinytui.Vertical)
	rootLayout.SetGap(1)                                       // Add a gap between the vertical panes
	rootLayout.AddPane(headerPane, tinytui.Size{FixedSize: 1}) // Top-Level (Index 0 internally, not shown)
	rootLayout.AddPane(paneA, tinytui.Size{Proportion: 1})     // Top-Level 1
	rootLayout.AddPane(paneB, tinytui.Size{FixedSize: 5})      // Top-Level 2
	rootLayout.AddPane(paneC, tinytui.Size{Proportion: 1})     // Top-Level 3

	// --- Set Application Layout ---
	app.SetLayout(rootLayout)

	// --- Initial Focus ---
	// Focus Grid A in Pane A initially
	app.Dispatch(&tinytui.FocusCommand{Target: gridA})

	// --- Run Application ---
	log.Println("Running indexing test application...")
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		log.Fatalf("Error running application: %v", err)
		os.Exit(1)
	}
	log.Println("Indexing test application exited normally.")
}