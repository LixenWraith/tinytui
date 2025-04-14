package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/LixenWraith/tinytui"
)

var (
	appInstance *tinytui.Application
	logText     *tinytui.Text
	statusText  *tinytui.Text
	spriteComp  *tinytui.Sprite
	logLines    []string
	logScroll   int
)

const maxLogLines = 100

// Helper to add log messages to both the file and the UI Text component
func appLog(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Println(msg) // Log to file

	logLines = append(logLines, msg)
	if len(logLines) > maxLogLines {
		logLines = logLines[len(logLines)-maxLogLines:] // Keep only the last N lines
	}

	if appInstance != nil && logText != nil {
		content := "--- Event Log ---\n" + strings.Join(logLines, "\n")
		appInstance.Dispatch(&tinytui.UpdateTextCommand{
			Target:  logText,
			Content: content,
		})
		// Auto-scroll log text would require enhancing the Text component or manual calculation here
		// logText.ScrollTo(len(logLines)) // Assuming ScrollTo exists and is safe to call directly (it's not)
	}
}

// updateStatus updates the status bar text
func updateStatus(msg string) {
	if appInstance != nil && statusText != nil {
		appInstance.Dispatch(&tinytui.UpdateTextCommand{
			Target:  statusText,
			Content: "Status: " + msg,
		})
	}
}

// updateSprite updates the sprite periodically
func updateSprite() {
	go func() {
		tick := time.NewTicker(500 * time.Millisecond)
		defer tick.Stop()
		state := 0
		for range tick.C {
			currentApp := appInstance
			if currentApp == nil {
				log.Println("updateSprite: App instance is nil, exiting goroutine.")
				return
			}

			currentSprite := spriteComp
			if currentSprite == nil {
				continue
			}

			w, h := currentSprite.Dimensions()
			if w == 0 || h == 0 {
				continue
			}

			cells := make([][]tinytui.SpriteCell, h) // Create new cells each time
			for r := 0; r < h; r++ {
				cells[r] = make([]tinytui.SpriteCell, w)
				for c := 0; c < w; c++ {
					style := tinytui.DefaultStyle
					runeChar := ' '
					if (r+c+state)%2 == 0 {
						style = style.Background(tinytui.ColorRed)
						runeChar = '*'
					} else {
						style = style.Background(tinytui.ColorBlue)
						runeChar = '+'
					}
					cells[r][c] = tinytui.SpriteCell{Rune: runeChar, Style: style}
				}
			}
			state++

			currentApp.Dispatch(&tinytui.UpdateSpriteCommand{
				Target:  currentSprite,
				Content: cells,
			})
		}
	}()
}

func main() {
	// --- Logging Setup ---
	logFile, err := os.OpenFile("tinytui_demo.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.Println("--- Starting TinyTUI Demo ---")

	// --- Application Setup ---
	app := tinytui.NewApplication()
	appInstance = app // Store global ref for logging and commands
	app.SetMaxFPS(30) // Limit FPS

	// --- Create Components ---
	header := tinytui.NewText("TinyTUI Demo Application")
	header.SetAlignment(tinytui.AlignTextCenter)
	footer := tinytui.NewText(" [Tab] Cycle Focus | [Alt+Num] Focus Pane | [T] Theme | [Esc] Quit ")
	statusText = tinytui.NewText("Status: Initializing...")

	nameInput := tinytui.NewTextInput()
	nameInput.SetText("Enter Name")
	submitButton := tinytui.NewGrid()
	submitButton.SetCells([][]string{{" Submit "}})
	submitButton.SetCellSize(10, 1)
	submitButton.SetSelectionMode(tinytui.SingleSelect)
	themeButton := tinytui.NewGrid()
	themeButton.SetCells([][]string{{" Theme "}})
	themeButton.SetCellSize(9, 1)
	themeButton.SetSelectionMode(tinytui.SingleSelect)

	logText = tinytui.NewText("--- Event Log ---")
	logText.SetWrap(true)

	selectableGrid := tinytui.NewGrid()
	gridData := [][]string{
		{"Option 1", "Value A"}, {"Option 2", "Value B"}, {"Option 3", "Value C"},
		{"Option 4", "Value D"}, {"Long Option 5", "Value E"},
	}
	selectableGrid.SetCells(gridData)
	selectableGrid.SetSelectionMode(tinytui.MultiSelect)
	selectableGrid.SetIndicator('*', true)

	spriteComp = tinytui.NewSprite(make([][]tinytui.SpriteCell, 5))
	spriteComp.Resize(10, 5)

	// --- Create Panes ---
	headerPane := tinytui.NewPane()
	headerPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle())
	headerPane.SetChild(header)

	footerPane := tinytui.NewPane()
	footerPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle())
	footerPane.SetChild(footer)

	statusPane := tinytui.NewPane()
	statusPane.SetBorder(tinytui.BorderSingle, tinytui.DefaultPaneBorderStyle())
	statusPane.SetChild(statusText)

	inputPane := tinytui.NewPane()
	inputPane.SetTitle("Input")
	inputPane.SetChild(nameInput)

	buttonPane := tinytui.NewPane()
	buttonPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle())

	logPane := tinytui.NewPane()
	logPane.SetTitle("Log Output")
	logPane.SetChild(logText)

	gridPane := tinytui.NewPane()
	gridPane.SetTitle("Options Grid")
	gridPane.SetChild(selectableGrid)

	spritePane := tinytui.NewPane()
	spritePane.SetTitle("Sprite Animation")
	spritePane.SetChild(spriteComp)

	// --- Setup Layouts ---
	// Button Layout (Horizontal) inside Button Pane
	buttonLayout := tinytui.NewLayout(tinytui.Horizontal)
	buttonLayout.SetGap(1)
	submitButtonPane := tinytui.NewPane()
	submitButtonPane.SetChild(submitButton)
	submitButtonPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle())
	themeButtonPane := tinytui.NewPane()
	themeButtonPane.SetChild(themeButton)
	themeButtonPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle())
	buttonLayout.AddPane(submitButtonPane, tinytui.Size{FixedSize: 10})
	buttonLayout.AddPane(themeButtonPane, tinytui.Size{FixedSize: 9})
	buttonPane.SetChild(buttonLayout)

	// Left Column Layout (Vertical)
	leftColLayout := tinytui.NewLayout(tinytui.Vertical)
	leftColLayout.SetGap(1)
	leftColLayout.AddPane(inputPane, tinytui.Size{FixedSize: 3})
	leftColLayout.AddPane(buttonPane, tinytui.Size{FixedSize: 1})

	// Right Column Layout (Vertical)
	rightColLayout := tinytui.NewLayout(tinytui.Vertical)
	rightColLayout.SetGap(1)
	rightColLayout.AddPane(gridPane, tinytui.Size{Proportion: 1})
	rightColLayout.AddPane(spritePane, tinytui.Size{FixedSize: 7})

	// Middle Area Layout (Horizontal: Left Col | Log | Right Col)
	middleLayout := tinytui.NewLayout(tinytui.Horizontal)
	middleLayout.SetGap(1)
	leftColWrapperPane := tinytui.NewPane()
	leftColWrapperPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle())
	leftColWrapperPane.SetChild(leftColLayout)
	rightColWrapperPane := tinytui.NewPane()
	rightColWrapperPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle())
	rightColWrapperPane.SetChild(rightColLayout)
	middleLayout.AddPane(leftColWrapperPane, tinytui.Size{FixedSize: 25})
	middleLayout.AddPane(logPane, tinytui.Size{Proportion: 1})
	middleLayout.AddPane(rightColWrapperPane, tinytui.Size{FixedSize: 30})

	// Main Layout (Vertical: Header | Middle | Status | Footer)
	mainLayout := tinytui.NewLayout(tinytui.Vertical)
	mainLayout.SetGap(0)
	middleWrapperPane := tinytui.NewPane()
	middleWrapperPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle())
	middleWrapperPane.SetChild(middleLayout)
	mainLayout.AddPane(headerPane, tinytui.Size{FixedSize: 1})
	mainLayout.AddPane(middleWrapperPane, tinytui.Size{Proportion: 1})
	mainLayout.AddPane(statusPane, tinytui.Size{FixedSize: 1})
	mainLayout.AddPane(footerPane, tinytui.Size{FixedSize: 1})

	// --- Set Application Layout ---
	app.SetLayout(mainLayout)

	// --- Event Handlers ---
	nameInput.SetOnChange(func(text string) {
		appLog("Input changed: %s", text)
		updateStatus("Typing...")
	})
	nameInput.SetOnSubmit(func(text string) {
		appLog("Input submitted: %s", text)
		updateStatus("Submitted: " + text)
		app.Dispatch(&tinytui.FocusCommand{Target: submitButton})
	})

	submitButton.SetOnSelect(func(r, c int, i string) {
		name := nameInput.GetText()
		appLog("Submit button pressed! Name: %s", name)
		updateStatus("Submitted: " + name)
		submitButton.SetCellInteracted(r, c, false)
		app.Dispatch(&tinytui.FocusCommand{Target: nameInput})
	})

	// **** START: Definition of switchThemeFunc ****
	// Define the theme switching logic in a local variable within main
	// This allows reusing the logic for both the button and the key handler
	switchThemeFunc := func() {
		appLog("Theme switch triggered!")
		currentThemeName := app.GetTheme().Name() // Get current theme NAME from the app instance
		var targetThemeName tinytui.ThemeName
		var statusMsg string

		if currentThemeName == tinytui.ThemeDefault {
			targetThemeName = tinytui.ThemeTurbo
			statusMsg = "Theme changed to Turbo"
		} else {
			targetThemeName = tinytui.ThemeDefault
			statusMsg = "Theme changed to Default"
		}

		// Set the GLOBAL theme first
		success := tinytui.SetTheme(targetThemeName)
		if success {
			// If global theme set successfully, update the specific APP instance's theme
			// This triggers the recursive ApplyTheme calls via app.SetTheme
			app.SetTheme(tinytui.GetTheme()) // Get the globally set theme instance
			updateStatus(statusMsg)
		} else {
			appLog("Failed to set theme: %s", targetThemeName)
			updateStatus("Failed to change theme.")
		}
		appLog("Global theme set success: %v, App theme is now: %s", success, app.GetTheme().Name())
	}
	// **** END: Definition of switchThemeFunc ****

	// Handler for the Theme Button Grid uses the function defined above
	themeButton.SetOnSelect(func(r, c int, i string) {
		switchThemeFunc()                          // Call the common logic
		themeButton.SetCellInteracted(r, c, false) // Deselect button visually
	})

	selectableGrid.SetOnChange(func(row, col int, item string) {
		appLog("Grid selection changed: Row %d, Col %d, Item '%s'", row, col, item)
		updateStatus(fmt.Sprintf("Grid selected: R%d C%d", row, col))
	})

	selectableGrid.SetOnSelect(func(row, col int, item string) {
		interacted := selectableGrid.IsCellInteracted(row, col) // Check state *after* internal toggle
		appLog("Grid item interacted: Row %d, Col %d ('%s'), State: %v", row, col, item, interacted)
		updateStatus(fmt.Sprintf("Grid interacted: R%d C%d (%v)", row, col, interacted))
		selected := selectableGrid.GetInteractedCells()
		appLog("Currently interacted cells: %v", selected)
	})

	// Register global key handler for 't'/'T' to toggle theme
	// This handler ALSO uses the switchThemeFunc defined above
	toggleThemeHandler := func() bool {
		appLog("Global key 'T'/'t' pressed")
		themeButton.SetCellInteracted(0, 0, true) // Visually press button

		time.AfterFunc(150*time.Millisecond, func() {
			currentApp := appInstance
			if currentApp != nil {
				// Use the SimpleCommand defined in tinytui/event.go
				currentApp.Dispatch(&tinytui.SimpleCommand{
					Func: func(a *tinytui.Application) { themeButton.SetCellInteracted(0, 0, false) },
				})
			}
		})

		switchThemeFunc() // Execute the shared theme switching logic

		return true // Mark event as handled
	}
	app.RegisterRuneHandler('t', 0, toggleThemeHandler) // For lowercase 't'
	app.RegisterRuneHandler('T', 0, toggleThemeHandler) // For uppercase 'T'

	// --- Initial State & Focus ---
	appLog("Application initialized.")
	updateStatus("Ready.")
	app.Dispatch(&tinytui.FocusCommand{Target: nameInput}) // Start focus in the input field
	updateSprite()                                         // Start sprite animation goroutine

	// --- Run Application ---
	log.Println("Running application event loop...")
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		log.Fatalf("Error running application: %v", err)
		os.Exit(1)
	}
	log.Println("Application exited normally.")
}