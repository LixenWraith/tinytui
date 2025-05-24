package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/LixenWraith/tinytui"
	// IMPORTANT: Ensure the tokyosweet theme is registered.
	// This typically happens in an init() function within its own theme file
	// (e.g., theme_tokyosweet.go) by calling tinytui.RegisterTheme(NewTokyoSweetTheme()).
	// If your project structure is different, you might need to explicitly call
	// a registration function from your main package or ensure the theme package
	// is imported for its side effects (init function).
	// Example:
	// _ "path/to/your/tokyosweet/theme/package"
)

const (
	// This must match the name used when registering the tokyosweet theme.
	tokyoSweetThemeName  tinytui.ThemeName = "tokyosweet"
	baseTextLength                         = 40
	differencePercentage                   = 0.25 // 25% difference
)

// generateSimilarText creates two strings: a base random string,
// and a second string that is similar to the first but with some characters changed.
func generateSimilarText(length int, diffPercentage float64) (string, string) {
	// Seed random number generator
	//nolint:gosec // For this example, weak random is fine.
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var baseBuilder strings.Builder
	for i := 0; i < length; i++ {
		// Printable ASCII characters from space to ~
		baseBuilder.WriteByte(byte(r.Intn('~'-' '+1) + ' '))
	}
	baseStr := baseBuilder.String()

	var similarBuilder strings.Builder
	runesBase := []rune(baseStr) // Work with runes for broader character support
	for _, char := range runesBase {
		if r.Float64() < diffPercentage {
			similarBuilder.WriteRune(rune(r.Intn('~'-' '+1) + ' '))
		} else {
			similarBuilder.WriteRune(char)
		}
	}
	return baseStr, similarBuilder.String()
}

func main() {
	app := tinytui.NewApplication()

	// Attempt to set the TokyoSweet theme globally.
	// Components created after this will try to use it.
	if !tinytui.SetTheme(tokyoSweetThemeName) {
		// This message will print to stdout before the TUI starts if the theme isn't found.
		fmt.Printf("Warning: Theme '%s' not found or not registered. Using current theme: %s\n",
			tokyoSweetThemeName, tinytui.GetTheme().Name())
		// Allow a moment for the user to see the warning if running directly
		time.Sleep(2 * time.Second)
	}
	// Ensure the application instance itself is explicitly synced if SetTheme was called after NewApplication
	// For components created *after* global SetTheme, this is often implicit.
	// But to be certain the app instance uses the globally set theme:
	app.SetTheme(tinytui.GetTheme())

	initialText1, initialText2 := generateSimilarText(baseTextLength, differencePercentage)

	// --- Components ---
	input1 := tinytui.NewTextInput()
	input1.SetText(initialText1)
	// TextInput itself doesn't have a title; its containing Pane will.

	input2 := tinytui.NewTextInput()
	input2.SetText(initialText2)

	submitButton := tinytui.NewGrid()
	submitButton.SetCells([][]string{{" Compare Texts "}})
	submitButton.SetCellSize(19, 1) // Adjusted for text length
	submitButton.SetSelectionMode(tinytui.SingleSelect)

	resultText := tinytui.NewText("Press 'Compare Texts' to check status.")
	resultText.SetAlignment(tinytui.AlignTextCenter)

	// --- Panes ---
	inputPane1 := tinytui.NewPane()
	inputPane1.SetTitle("Text Area 1 (Goal: Match Text Area 2)")
	inputPane1.SetChild(input1)

	inputPane2 := tinytui.NewPane()
	inputPane2.SetTitle("Text Area 2 (Goal: Match Text Area 1)")
	inputPane2.SetChild(input2)

	// Pane for the bottom section (result message and button)
	bottomSectionPane := tinytui.NewPane()
	bottomSectionPane.SetTitle("Controls & Status") // Give the whole bottom section a title

	// Panes within the bottom section (to control layout of result and button)
	resultDisplayPane := tinytui.NewPane()
	resultDisplayPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle()) // No border for internal layout elements
	resultDisplayPane.SetChild(resultText)

	buttonWrapperPane := tinytui.NewPane()
	buttonWrapperPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle())
	buttonWrapperPane.SetChild(submitButton)

	// --- Layouts ---
	// Top layout for the two input panes, side-by-side
	topEditLayout := tinytui.NewLayout(tinytui.Horizontal)
	topEditLayout.SetGap(1) // Add a 1-cell gap between the input panes
	topEditLayout.AddPane(inputPane1, tinytui.Size{Proportion: 1})
	topEditLayout.AddPane(inputPane2, tinytui.Size{Proportion: 1})

	// Create a wrapper pane for the topEditLayout
	topWrapperPane := tinytui.NewPane()
	topWrapperPane.SetChild(topEditLayout)
	// Optionally, give the topWrapperPane a title or specific border
	// topWrapperPane.SetTitle("Editing Area")
	// topWrapperPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle()) // If you don't want a border around the whole top section

	// Internal layout for the bottom section (result text above button)
	bottomInternalLayout := tinytui.NewLayout(tinytui.Vertical)
	bottomInternalLayout.SetGap(0)                                              // No gap between result text and button internally
	bottomInternalLayout.AddPane(resultDisplayPane, tinytui.Size{FixedSize: 1}) // Result text takes 1 row
	bottomInternalLayout.AddPane(buttonWrapperPane, tinytui.Size{FixedSize: 1}) // Button takes 1 row
	bottomSectionPane.SetChild(bottomInternalLayout)                            // Assign this internal layout to the main bottom pane

	// Main application layout: top editing area over the bottom controls/status area
	mainLayout := tinytui.NewLayout(tinytui.Vertical)
	mainLayout.SetGap(1)                                              // Add a 1-cell gap between top and bottom sections
	mainLayout.AddPane(topWrapperPane, tinytui.Size{Proportion: 1})   // Add the wrapper pane
	mainLayout.AddPane(bottomSectionPane, tinytui.Size{FixedSize: 4}) // Bottom area: 1 for title, 1 for result, 1 for button, 1 for bottom border = 4

	app.SetLayout(mainLayout)

	// --- Event Handlers ---
	submitButton.SetOnSelect(func(r, c int, item string) {
		text1 := input1.GetText()
		text2 := input2.GetText()

		var message string
		if text1 == text2 {
			message = "OK: Texts are identical!"
		} else {
			message = "NOT OK: Texts differ."
		}

		// Update the resultText component via a command dispatched to the application
		app.Dispatch(&tinytui.UpdateTextCommand{Target: resultText, Content: message})

		// Visually "deselect" the button after click by clearing its interacted state.
		// This assumes SetOnSelect might set it to interacted.
		submitButton.SetCellInteracted(r, c, false)
		app.QueueRedraw() // Ensure the UI redraws to reflect the change
	})

	// --- Initial Focus ---
	// Start with focus on the first text input area
	app.Dispatch(&tinytui.FocusCommand{Target: input1})

	// --- Run Application ---
	if err := app.Run(); err != nil {
		// Use fmt.Fprintf for stderr, as 'log' might be configured elsewhere or not visible
		// if the TUI fails very early.
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}