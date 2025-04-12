package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/LixenWraith/tinytui"
	"github.com/LixenWraith/tinytui/widgets"
	"github.com/gdamore/tcell/v2"
)

type TodoItem struct {
	Text      string
	Completed bool
}

func main() {
	// Create application
	app := tinytui.NewApplication()
	if app == nil {
		fmt.Println("Error: Could not create application")
		os.Exit(1)
	}

	// Create main layout with two columns
	mainLayout := tinytui.NewFlexLayout(tinytui.Horizontal)

	// Create sidebar for theme and controls
	sidebar := createSidebar(app)

	// Create content area with todo list example
	contentArea := createContentArea(app)

	// Add components to main layout
	mainLayout.AddChild(sidebar, 28, 0)    // Fixed width sidebar
	mainLayout.AddChild(contentArea, 0, 1) // Content takes remaining space

	// Set application root
	app.SetRoot(mainLayout, true)

	// Global key binding to quit with Escape
	mainLayout.SetKeybinding(tcell.KeyEscape, tcell.ModNone, func() bool {
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
	// Create sidebar pane with border
	sidebarPane := widgets.NewPane()
	sidebarPane.SetBorder(true, tinytui.BorderSingle, tinytui.DefaultPaneBorderStyle())

	// Create vertical layout for sidebar content
	sidebarLayout := tinytui.NewFlexLayout(tinytui.Vertical)
	sidebarLayout.SetGap(1)

	// Add title
	titleText := widgets.NewText("TinyTUI Demo")
	titleText.SetStyle(tinytui.DefaultTextStyle().Bold(true))

	// Add clock display that updates
	clockText := widgets.NewText(time.Now().Format("15:04:05"))

	// Create theme selector section with header
	themeHeader := widgets.NewText("Theme Selection")
	themeHeader.SetStyle(tinytui.DefaultTextStyle().Underline(true))

	// Theme selector grid
	themeGrid := widgets.NewGrid()
	themeGrid.SetCellSize(24, 1)
	themeGrid.SetPadding(1)
	themeGrid.SetCells([][]string{
		{"Default Theme"},
		{"Borland Classic"},
	})

	// Handle theme selection
	themeGrid.SetOnSelect(func(row, col int, item string) {
		var themeName tinytui.ThemeName
		switch row {
		case 0:
			themeName = tinytui.ThemeDefault
		case 1:
			themeName = tinytui.ThemeBorland
		}

		// Apply selected theme
		if app.SetTheme(themeName) {
			clockText.SetContent(fmt.Sprintf("%s - Theme Changed", time.Now().Format("15:04:05")))
		}
	})

	// Create help section
	helpHeader := widgets.NewText("Keyboard Controls")
	helpHeader.SetStyle(tinytui.DefaultTextStyle().Underline(true))

	helpItems := []string{
		"Tab: Move focus",
		"Arrow Keys: Navigate",
		"Space: Interact",
		"Enter: Toggle interaction",
		"Backspace: Cancel interaction",
		"Escape: Quit application",
	}

	// Create button-like grid for refreshing
	refreshGrid := widgets.NewGrid()
	refreshGrid.SetCellSize(24, 1)
	refreshGrid.SetPadding(1)
	refreshGrid.SetCells([][]string{{"Refresh Time"}})

	// Handle refresh button
	refreshGrid.SetOnSelect(func(row, col int, item string) {
		clockText.SetContent(time.Now().Format("15:04:05"))
	})

	// Add all components to layout
	sidebarLayout.AddChild(titleText, 2, 0)
	sidebarLayout.AddChild(clockText, 2, 0)
	sidebarLayout.AddChild(widgets.NewText(""), 1, 0) // Spacer

	sidebarLayout.AddChild(themeHeader, 1, 0)
	sidebarLayout.AddChild(themeGrid, 4, 0)
	sidebarLayout.AddChild(widgets.NewText(""), 1, 0) // Spacer

	sidebarLayout.AddChild(helpHeader, 1, 0)
	for _, item := range helpItems {
		helpText := widgets.NewText(item)
		sidebarLayout.AddChild(helpText, 1, 0)
	}

	sidebarLayout.AddChild(widgets.NewText(""), 1, 0) // Spacer
	sidebarLayout.AddChild(refreshGrid, 3, 0)

	// Add footer with version
	footer := widgets.NewText("TinyTUI v1.0.0")
	footer.SetStyle(tinytui.DefaultTextStyle().Italic(true))
	sidebarLayout.AddChild(footer, 0, 1) // Take remaining space

	// Set the layout as pane's child
	sidebarPane.SetChild(sidebarLayout)

	return sidebarPane
}

// createContentArea creates the main content area with a todo list
func createContentArea(app *tinytui.Application) *widgets.Pane {
	// Create main content pane with border
	contentPane := widgets.NewPane()
	contentPane.SetBorder(true, tinytui.BorderSingle, tinytui.DefaultPaneBorderStyle())

	// Create layout for content
	contentLayout := tinytui.NewFlexLayout(tinytui.Vertical)
	contentLayout.SetGap(1)

	// Add title
	titleText := widgets.NewText("Todo List Example")
	titleText.SetStyle(tinytui.DefaultTextStyle().Bold(true))

	// Create a description
	descText := widgets.NewText("Use Space to toggle item completion. Enter to edit status.")
	descText.SetWrap(true)

	// Status text for showing actions
	statusText := widgets.NewText("Ready")
	statusText.SetStyle(tinytui.DefaultTextStyle().Italic(true))

	// Initial todo items
	todos := []TodoItem{
		{Text: "Learn TinyTUI basics", Completed: true},
		{Text: "Create a todo list app", Completed: false},
		{Text: "Implement theme switching", Completed: false},
		{Text: "Add keyboard navigation", Completed: false},
		{Text: "Test on different terminals", Completed: false},
		{Text: "Document the code", Completed: false},
		{Text: "Share with others", Completed: false},
	}

	// Create grid for todo list
	todoGrid := widgets.NewGrid()
	todoGrid.SetCellSize(0, 1) // Auto-width
	todoGrid.SetPadding(1)

	// Function to update grid cells from todos
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
	}

	// Initial grid update
	updateTodoGrid()

	// Set interactions for todo grid
	todoGrid.SetOnSelect(func(row, col int, item string) {
		// Toggle completion status
		if row >= 0 && row < len(todos) {
			todos[row].Completed = !todos[row].Completed
			updateTodoGrid()

			item := todos[row]
			status := "incomplete"
			if item.Completed {
				status = "complete"
			}
			statusText.SetContent(fmt.Sprintf("Marked '%s' as %s", item.Text, status))
		}
	})

	todoGrid.SetOnChange(func(row, col int, item string) {
		if row >= 0 && row < len(todos) {
			todo := todos[row]
			status := "incomplete"
			if todo.Completed {
				status = "complete"
			}
			statusText.SetContent(fmt.Sprintf("Selected: '%s' (%s)", todo.Text, status))
		}
	})

	// Create a section for adding new todos
	addTodoHeader := widgets.NewText("Add New Todo")
	addTodoHeader.SetStyle(tinytui.DefaultTextStyle().Underline(true))

	// Text explaining how to add (since we don't have real input)
	addTodoText := widgets.NewText("Select an action to simulate adding a todo")

	// Grid with add options
	addOptionsGrid := widgets.NewGrid()
	addOptionsGrid.SetCellSize(30, 1)
	addOptionsGrid.SetPadding(1)
	addOptionsGrid.SetCells([][]string{
		{"Add: Buy groceries"},
		{"Add: Call mom"},
		{"Add: Fix the bug"},
		{"Add: Clean up code"},
	})

	// Handle adding new todos
	addOptionsGrid.SetOnSelect(func(row, col int, item string) {
		// Extract todo text from the selection
		text := strings.TrimPrefix(item, "Add: ")

		// Add the new todo
		todos = append(todos, TodoItem{Text: text, Completed: false})
		updateTodoGrid()

		statusText.SetContent(fmt.Sprintf("Added new todo: '%s'", text))
	})

	// Create a section for data summary
	statsHeader := widgets.NewText("Summary")
	statsHeader.SetStyle(tinytui.DefaultTextStyle().Underline(true))

	// Create grid for statistics
	statsGrid := widgets.NewGrid()
	statsGrid.SetCellSize(20, 1)
	statsGrid.SetPadding(1)

	// Function to update statistics
	updateStats := func() {
		total := len(todos)
		completed := 0
		for _, todo := range todos {
			if todo.Completed {
				completed++
			}
		}
		remaining := total - completed

		statsGrid.SetCells([][]string{
			{fmt.Sprintf("Total Tasks: %d", total)},
			{fmt.Sprintf("Completed: %d", completed)},
			{fmt.Sprintf("Remaining: %d", remaining)},
			{fmt.Sprintf("Progress: %d%%", (completed*100)/max(total, 1))},
		})
	}

	// Initial stats update
	updateStats()

	// Add interaction to refresh stats
	statsGrid.SetOnSelect(func(row, col int, item string) {
		updateStats()
		statusText.SetContent("Statistics updated")
	})

	// Add clear completed button
	clearGrid := widgets.NewGrid()
	clearGrid.SetCellSize(30, 1)
	clearGrid.SetPadding(1)
	clearGrid.SetCells([][]string{{"Clear Completed Tasks"}})

	clearGrid.SetOnSelect(func(row, col int, item string) {
		// Filter out completed tasks
		newTodos := []TodoItem{}
		removedCount := 0

		for _, todo := range todos {
			if !todo.Completed {
				newTodos = append(newTodos, todo)
			} else {
				removedCount++
			}
		}

		todos = newTodos
		updateTodoGrid()
		updateStats()

		statusText.SetContent(fmt.Sprintf("Removed %d completed tasks", removedCount))
	})

	// Add all components to layout
	contentLayout.AddChild(titleText, 1, 0)
	contentLayout.AddChild(descText, 2, 0)
	contentLayout.AddChild(todoGrid, 10, 0)

	contentLayout.AddChild(widgets.NewText(""), 1, 0) // Spacer

	// Add new todo section
	contentLayout.AddChild(addTodoHeader, 1, 0)
	contentLayout.AddChild(addTodoText, 1, 0)
	contentLayout.AddChild(addOptionsGrid, 6, 0)

	contentLayout.AddChild(widgets.NewText(""), 1, 0) // Spacer

	// Stats section
	contentLayout.AddChild(statsHeader, 1, 0)
	contentLayout.AddChild(statsGrid, 6, 0)
	contentLayout.AddChild(clearGrid, 3, 0)

	contentLayout.AddChild(widgets.NewText(""), 1, 0) // Spacer

	// Status bar at bottom
	contentLayout.AddChild(statusText, 0, 1) // Fill remaining space

	// Set the layout as pane's child
	contentPane.SetChild(contentLayout)

	return contentPane
}

// max returns the larger of x or y
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}