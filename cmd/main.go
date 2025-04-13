package main

import (
	"fmt"
	"os"
	"time"

	"github.com/LixenWraith/tinytui"
	"github.com/LixenWraith/tinytui/widgets"
	"github.com/gdamore/tcell/v2"
)

// TodoItem represents a single todo entry with text and completion state
type TodoItem struct {
	Text      string
	Completed bool
}

// ThemeOption represents a selectable theme option
type ThemeOption struct {
	Name      string
	ThemeName tinytui.ThemeName
}

// Make the updateStats variable a function variable to allow replacement
var updateStats func(todos *[]TodoItem, statsGrid *widgets.Grid) = func(todos *[]TodoItem, statsGrid *widgets.Grid) {
	total := len(*todos)
	completed := 0

	for _, todo := range *todos {
		if todo.Completed {
			completed++
		}
	}

	remaining := total - completed
	progress := 0
	if total > 0 {
		progress = (completed * 100) / total
	}

	statsGrid.SetCells([][]string{
		{fmt.Sprintf("Total Tasks: %d", total)},
		{fmt.Sprintf("Completed: %d", completed)},
		{fmt.Sprintf("Remaining: %d", remaining)},
		{fmt.Sprintf("Progress: %d%%", progress)},
	})
}

func main() {
	// Create application
	app := tinytui.NewApplication()
	if app == nil {
		fmt.Fprintln(os.Stderr, "Error: Could not create application")
		os.Exit(1)
	}

	// Create main layout with two columns
	mainLayout := tinytui.NewFlexLayout(tinytui.Horizontal)

	// Create sidebar and content area
	sidebar := createSidebar(app)
	contentArea := createContentArea(app)

	// Add components to main layout
	mainLayout.AddChild(sidebar, 28, 0)    // Fixed width sidebar
	mainLayout.AddChild(contentArea, 0, 1) // Content takes remaining space

	// Set application root
	app.SetRoot(mainLayout, true)

	// Global key binding to quit with Escape
	app.SetKeybinding(tcell.KeyEscape, tcell.ModNone, func() bool {
		app.Stop()
		return true
	})

	// Run application
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %s\n", err)
		os.Exit(1)
	}
}

// createSidebar creates the left sidebar with theme selection and controls
func createSidebar(app *tinytui.Application) *widgets.Pane {
	// Create sidebar pane with themed border
	sidebarPane := widgets.NewPane()
	sidebarPane.SetBorder(true, tinytui.DefaultBorderType(), tinytui.DefaultPaneBorderStyle())

	// Create vertical layout for sidebar content
	sidebarLayout := tinytui.NewFlexLayout(tinytui.Vertical)
	sidebarLayout.SetGap(1)
	sidebarLayout.SetMainAxisAlignment(tinytui.AlignStart)
	sidebarLayout.SetCrossAxisAlignment(tinytui.AlignStretch)

	// Define available themes
	themeOptions := []ThemeOption{
		{Name: "Default Theme", ThemeName: tinytui.ThemeDefault},
		{Name: "Turbo Classic", ThemeName: tinytui.ThemeTurbo},
	}

	// Create title section
	titleSection := createTitleSection()

	// Create theme selector section
	themeSection := createThemeSection(app, themeOptions)

	// Create help section
	helpSection := createHelpSection()

	// Create status section
	statusSection := createStatusSection(app)

	// Add all sections to the sidebar layout
	sidebarLayout.AddChild(titleSection, 4, 0)
	sidebarLayout.AddChild(widgets.NewText(""), 1, 0) // Spacer
	sidebarLayout.AddChild(themeSection, 7, 0)
	sidebarLayout.AddChild(widgets.NewText(""), 1, 0) // Spacer
	sidebarLayout.AddChild(helpSection, 10, 0)
	sidebarLayout.AddChild(widgets.NewText(""), 1, 0) // Spacer
	sidebarLayout.AddChild(statusSection, 0, 1)       // Flexible space

	// Add footer with version
	footer := widgets.NewText("TinyTUI v1.1.0")
	footer.SetStyle(tinytui.DefaultTextStyle().Italic(true))
	sidebarLayout.AddChild(footer, 1, 0)

	// Set the layout as pane's child
	sidebarPane.SetChild(sidebarLayout)

	return sidebarPane
}

// createTitleSection creates the app title section
func createTitleSection() *tinytui.FlexLayout {
	titleLayout := tinytui.NewFlexLayout(tinytui.Vertical)

	// Title text with bold style
	titleText := widgets.NewText("TinyTUI Demo")
	titleText.SetStyle(tinytui.DefaultTextStyle().Bold(true))

	// Description text
	descText := widgets.NewText("A lightweight terminal UI demo")

	titleLayout.AddChild(titleText, 1, 0)
	titleLayout.AddChild(descText, 1, 0)

	return titleLayout
}

// createThemeSection creates the theme selector section
func createThemeSection(app *tinytui.Application, themeOptions []ThemeOption) *tinytui.FlexLayout {
	themeLayout := tinytui.NewFlexLayout(tinytui.Vertical)

	// Theme section header
	themeHeader := widgets.NewText("Theme Selection")
	themeHeader.SetStyle(tinytui.DefaultTextStyle().Underline(true))

	// Create theme selector grid
	themeGrid := widgets.NewGrid()
	themeGrid.SetSelectionMode(widgets.SingleSelect)
	themeGrid.SetCellSize(24, 1)
	themeGrid.SetPadding(1)

	// Populate theme grid
	themeItems := make([][]string, len(themeOptions))
	for i, theme := range themeOptions {
		themeItems[i] = []string{theme.Name}
	}
	themeGrid.SetCells(themeItems)

	// Find and set initial selection based on current theme
	currentTheme := app.GetTheme().Name()
	for i, theme := range themeOptions {
		if theme.ThemeName == currentTheme {
			themeGrid.SetSelectedIndex(i, 0)
			themeGrid.SetCellInteracted(i, 0, true)
			break
		}
	}

	// Handle theme selection
	themeGrid.SetOnSelect(func(row, col int, item string) {
		if row >= 0 && row < len(themeOptions) {
			// Apply the selected theme
			app.SetTheme(themeOptions[row].ThemeName)

			// Update grid interaction state
			themeGrid.ClearInteractions()
			themeGrid.SetCellInteracted(row, col, true)
		}
	})

	themeLayout.AddChild(themeHeader, 1, 0)
	themeLayout.AddChild(themeGrid, 0, 1)

	return themeLayout
}

// createHelpSection creates the keyboard help guide
func createHelpSection() *tinytui.FlexLayout {
	helpLayout := tinytui.NewFlexLayout(tinytui.Vertical)

	// Help section header
	helpHeader := widgets.NewText("Keyboard Controls")
	helpHeader.SetStyle(tinytui.DefaultTextStyle().Underline(true))

	// Help items
	helpItems := []string{
		"Tab: Move focus",
		"Arrow Keys: Navigate",
		"Space/Enter: Toggle selection",
		"Backspace: Clear selection",
		"Escape: Quit application",
	}

	// Add all help items
	helpLayout.AddChild(helpHeader, 1, 0)
	for _, item := range helpItems {
		helpText := widgets.NewText(item)
		helpLayout.AddChild(helpText, 1, 0)
	}

	return helpLayout
}

// createStatusSection creates the status section with clock
func createStatusSection(app *tinytui.Application) *tinytui.FlexLayout {
	statusLayout := tinytui.NewFlexLayout(tinytui.Vertical)

	// Initial clock text
	clockText := widgets.NewText(time.Now().Format("15:04:05"))

	// Create refresh button
	refreshButton := widgets.NewGrid()
	refreshButton.SetCellSize(24, 1)
	refreshButton.SetPadding(1)
	refreshButton.SetCells([][]string{{"Refresh Time"}})

	// Handle refresh button
	refreshButton.SetOnSelect(func(row, col int, item string) {
		clockText.SetContent(time.Now().Format("15:04:05"))

		// Visual feedback for button press
		refreshButton.ClearInteractions()
		refreshButton.SetCellInteracted(row, col, true)

		// Clear interaction after a short delay
		go func() {
			time.Sleep(300 * time.Millisecond)
			app.Dispatch(func(app *tinytui.Application) {
				refreshButton.ClearInteractions()
			})
		}()
	})

	statusLayout.AddChild(clockText, 1, 0)
	statusLayout.AddChild(refreshButton, 2, 0)

	return statusLayout
}

// createContentArea creates the main content area with a todo list
func createContentArea(app *tinytui.Application) *widgets.Pane {
	// Create content pane with border
	contentPane := widgets.NewPane()
	contentPane.SetBorder(true, tinytui.DefaultBorderType(), tinytui.DefaultPaneBorderStyle())

	// Create main layout for todo app
	contentLayout := tinytui.NewFlexLayout(tinytui.Vertical)
	contentLayout.SetGap(1)

	// Status text for showing actions
	statusText := widgets.NewText("Ready")
	statusText.SetStyle(tinytui.DefaultTextStyle().Italic(true))

	// Initialize todo model as a slice pointer so it can be shared
	todos := []TodoItem{
		{Text: "Learn TinyTUI basics", Completed: true},
		{Text: "Create a todo list app", Completed: false},
		{Text: "Implement theme switching", Completed: false},
		{Text: "Add keyboard navigation", Completed: false},
		{Text: "Test on different terminals", Completed: false},
	}

	// Create todo grid
	todoGrid := widgets.NewGrid()
	todoGrid.SetSelectionMode(widgets.MultiSelect)
	todoGrid.SetCellSize(40, 1)
	todoGrid.SetAutoWidth(true)
	todoGrid.SetPadding(1)

	// Create todo header
	todoHeader := widgets.NewPane()
	todoHeaderText := widgets.NewText("Todo List")
	todoHeaderText.SetStyle(tinytui.DefaultTextStyle().Bold(true))
	todoHeader.SetChild(todoHeaderText)

	// Create stats grid separately so it can be passed to update function
	statsGrid := widgets.NewGrid()
	statsGrid.SetCellSize(25, 1)
	statsGrid.SetPadding(1)
	statsGrid.SetSelectionMode(widgets.SingleSelect)

	// Create todo control section
	todoControls := createTodoControls(app, &todos, todoGrid, statusText, statsGrid)

	// Create todo stats section
	statsSection := createStatsSection(app, &todos, todoGrid, statusText, statsGrid)

	// Synchronize grid with the todo model
	updateTodoGrid := func() {
		cells := make([][]string, len(todos))
		for i, todo := range todos {
			status := "[ ]"
			if todo.Completed {
				status = "[✓]"
			}
			cells[i] = []string{fmt.Sprintf("%s %s", status, todo.Text)}
		}
		todoGrid.SetCells(cells)

		// Synchronize interaction states with completion status
		todoGrid.ClearInteractions()
		for i, todo := range todos {
			if todo.Completed {
				todoGrid.SetCellInteracted(i, 0, true)
			}
		}

		// Update stats
		updateStats(&todos, statsGrid)
	}

	// Initial grid update
	updateTodoGrid()

	// Handle todo item toggle
	todoGrid.SetOnSelect(func(row, col int, item string) {
		if row >= 0 && row < len(todos) {
			// Toggle completion status
			todos[row].Completed = !todos[row].Completed

			// Update UI to reflect the change
			updateTodoGrid()

			// Update status message
			status := "incomplete"
			if todos[row].Completed {
				status = "complete"
			}
			statusText.SetContent(fmt.Sprintf("Marked '%s' as %s", todos[row].Text, status))
		}
	})

	// Handle selection changes
	todoGrid.SetOnChange(func(row, col int, item string) {
		if row >= 0 && row < len(todos) {
			status := "incomplete"
			if todos[row].Completed {
				status = "complete"
			}
			statusText.SetContent(fmt.Sprintf("Selected: '%s' (%s)", todos[row].Text, status))
		}
	})

	// Add components to content layout
	contentLayout.AddChild(todoHeader, 2, 0)
	contentLayout.AddChild(todoGrid, 10, 0)
	contentLayout.AddChild(widgets.NewText(""), 1, 0) // Spacer
	contentLayout.AddChild(todoControls, 6, 0)
	contentLayout.AddChild(widgets.NewText(""), 1, 0) // Spacer
	contentLayout.AddChild(statsSection, 6, 0)
	contentLayout.AddChild(widgets.NewText(""), 0, 1) // Flexible spacer
	contentLayout.AddChild(statusText, 1, 0)          // Status at bottom

	// Set content layout as pane's child
	contentPane.SetChild(contentLayout)

	return contentPane
}

// createTodoControls creates the add/remove todo controls
// Improved to directly connect selection to action
func createTodoControls(app *tinytui.Application, todos *[]TodoItem, todoGrid *widgets.Grid, statusText *widgets.Text, statsGrid *widgets.Grid) *tinytui.FlexLayout {
	controlsLayout := tinytui.NewFlexLayout(tinytui.Vertical)

	// Add new todo section header
	addHeader := widgets.NewText("Add New Todo")
	addHeader.SetStyle(tinytui.DefaultTextStyle().Underline(true))

	// Explanation text
	addText := widgets.NewText("Select an option to add a new todo:")

	// Predefined options
	options := []string{
		"Buy groceries",
		"Write documentation",
		"Fix reported bugs",
		"Review pull requests",
	}

	// Create option buttons - each option is directly clickable
	optionsLayout := tinytui.NewFlexLayout(tinytui.Vertical)

	for _, option := range options {
		// Create a button-like grid for each option
		optionGrid := widgets.NewGrid()
		optionGrid.SetCellSize(40, 1)
		optionGrid.SetPadding(1)
		optionGrid.SetSelectionMode(widgets.SingleSelect)
		optionGrid.SetCells([][]string{{option}})

		// Set direct action on selection
		todoText := option // Capture the option text outside the closure
		optionGrid.SetOnSelect(func(row, col int, item string) {
			// Add new todo item using the option text

			// Add new todo item
			*todos = append(*todos, TodoItem{Text: todoText, Completed: false})

			// Update UI
			updateTodoGridFromOptions(*todos, todoGrid, statsGrid)

			// Update status
			statusText.SetContent(fmt.Sprintf("Added new todo: '%s'", todoText))

			// Provide visual feedback
			optionGrid.ClearInteractions()
			optionGrid.SetCellInteracted(row, col, true)

			// Clear button interaction after a short delay
			go func() {
				time.Sleep(300 * time.Millisecond)
				app.Dispatch(func(app *tinytui.Application) {
					optionGrid.ClearInteractions()
				})
			}()
		})

		optionsLayout.AddChild(optionGrid, 1, 0)
	}

	// Add all components to layout
	controlsLayout.AddChild(addHeader, 1, 0)
	controlsLayout.AddChild(addText, 1, 0)
	controlsLayout.AddChild(optionsLayout, 0, 1)

	return controlsLayout
}

// Helper function to update todo grid from options
func updateTodoGridFromOptions(todos []TodoItem, todoGrid *widgets.Grid, statsGrid *widgets.Grid) {
	// Update grid cells
	cells := make([][]string, len(todos))
	for i, todo := range todos {
		status := "[ ]"
		if todo.Completed {
			status = "[✓]"
		}
		cells[i] = []string{fmt.Sprintf("%s %s", status, todo.Text)}
	}
	todoGrid.SetCells(cells)

	// Sync interaction states
	todoGrid.ClearInteractions()
	for i, todo := range todos {
		if todo.Completed {
			todoGrid.SetCellInteracted(i, 0, true)
		}
	}

	// Update stats
	updateStats(&todos, statsGrid)
}

// createStatsSection creates the statistics and clear completed section
// Improved with fixes for button visibility
func createStatsSection(app *tinytui.Application, todos *[]TodoItem, todoGrid *widgets.Grid, statusText *widgets.Text, statsGrid *widgets.Grid) *tinytui.FlexLayout {
	statsLayout := tinytui.NewFlexLayout(tinytui.Vertical)

	// Stats header
	statsHeader := widgets.NewText("Summary")
	statsHeader.SetStyle(tinytui.DefaultTextStyle().Underline(true))

	// Create non-focusable stats display using Text widgets
	statsDisplay := tinytui.NewFlexLayout(tinytui.Vertical)

	// Create individual text fields for each stat
	totalText := widgets.NewText("")
	completedText := widgets.NewText("")
	remainingText := widgets.NewText("")
	progressText := widgets.NewText("")

	// Style the stats text - make them stand out more
	totalStyle := tinytui.DefaultTextStyle().Bold(true)
	totalText.SetStyle(totalStyle)

	// Add them to display layout
	statsDisplay.AddChild(totalText, 1, 0)
	statsDisplay.AddChild(completedText, 1, 0)
	statsDisplay.AddChild(remainingText, 1, 0)
	statsDisplay.AddChild(progressText, 1, 0)

	// Create custom updateStats function that works with Text widgets
	updateStatsDisplay := func(todos *[]TodoItem) {
		total := len(*todos)
		completed := 0

		for _, todo := range *todos {
			if todo.Completed {
				completed++
			}
		}

		remaining := total - completed
		progress := 0
		if total > 0 {
			progress = (completed * 100) / total
		}

		// Update text widgets with stats
		totalText.SetContent(fmt.Sprintf("Total Tasks: %d", total))
		completedText.SetContent(fmt.Sprintf("Completed: %d", completed))
		remainingText.SetContent(fmt.Sprintf("Remaining: %d", remaining))
		progressText.SetContent(fmt.Sprintf("Progress: %d%%", progress))

		// Also update the original statsGrid for compatibility with other functions
		statsGrid.SetCells([][]string{
			{fmt.Sprintf("Total Tasks: %d", total)},
			{fmt.Sprintf("Completed: %d", completed)},
			{fmt.Sprintf("Remaining: %d", remaining)},
			{fmt.Sprintf("Progress: %d%%", progress)},
		})
	}

	// Initial stats update
	updateStatsDisplay(todos)

	// Replace updateStats globally
	updateStats = func(todos *[]TodoItem, statsGrid *widgets.Grid) {
		updateStatsDisplay(todos)
	}

	// Add a visible spacer before clear button
	spacer := widgets.NewText("")

	// Clear completed button - fixed visibility issue
	clearButton := widgets.NewGrid()
	clearButton.SetCellSize(25, 1)
	clearButton.SetPadding(1)
	clearButton.SetSelectionMode(widgets.SingleSelect) // Ensure it's selectable
	clearButton.SetCells([][]string{{"Clear Completed Tasks"}})

	// Ensure the grid has proper styling
	clearButton.SetStyle(tinytui.DefaultGridStyle())                   // Set standard style
	clearButton.SetSelectedStyle(tinytui.DefaultGridSelectedStyle())   // Set highlighted style
	clearButton.SetFocusedStyle(tinytui.DefaultGridStyle().Bold(true)) // Set focused style

	// Make sure it's visible
	clearButton.SetVisible(true)

	// Handle clear button
	clearButton.SetOnSelect(func(row, col int, item string) {
		// Filter out completed tasks
		newTodos := []TodoItem{}
		removedCount := 0

		for _, todo := range *todos {
			if !todo.Completed {
				newTodos = append(newTodos, todo)
			} else {
				removedCount++
			}
		}

		// Update todos slice
		*todos = newTodos

		// Update UI
		updateTodoGridFromOptions(*todos, todoGrid, statsGrid)

		// Update stats display
		updateStatsDisplay(todos)

		// Update status
		statusText.SetContent(fmt.Sprintf("Removed %d completed tasks", removedCount))

		// Provide visual feedback
		clearButton.ClearInteractions()
		clearButton.SetCellInteracted(row, col, true)

		// Clear button interaction after a short delay
		go func() {
			time.Sleep(300 * time.Millisecond)
			app.Dispatch(func(app *tinytui.Application) {
				clearButton.ClearInteractions()
			})
		}()
	})

	// Add components to layout with explicit sizing
	statsLayout.AddChild(statsHeader, 1, 0)
	statsLayout.AddChild(statsDisplay, 4, 0)
	statsLayout.AddChild(spacer, 1, 0)      // Visible spacer
	statsLayout.AddChild(clearButton, 2, 0) // Give it 2 rows height for visibility

	return statsLayout
}