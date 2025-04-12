package main

import (
	"fmt"
	"os"

	"github.com/LixenWraith/tinytui"
	"github.com/LixenWraith/tinytui/widgets"
	"github.com/gdamore/tcell/v2"
)

func main() {
	// Create and initialize the application
	app := tinytui.NewApplication()
	if app == nil {
		fmt.Println("Could not create application")
		os.Exit(1)
	}

	// Create the main layout (horizontal)
	mainLayout := tinytui.NewFlexLayout(tinytui.Horizontal)

	// Create sidebar for theme selection
	sidebarWidth := 25
	sidebar := createSidebar(app, sidebarWidth)

	// Create content area
	contentArea := createContentArea()

	// Add sidebar and content to main layout
	mainLayout.AddChild(sidebar, sidebarWidth, 0)
	mainLayout.AddChild(contentArea, 0, 1)

	// Set the root of the application
	app.SetRoot(mainLayout, true)

	// Set keybinding to quit the application
	mainLayout.SetKeybinding(tcell.KeyEscape, tcell.ModNone, func() bool {
		app.Stop()
		return true
	})

	// Show instructions
	statusText := widgets.NewText("ESC to quit | TAB to navigate | ENTER to select")
	statusText.SetStyle(tinytui.DefaultTextStyle().Italic(true))

	// Add a status bar at the bottom
	mainLayout.SetKeybinding(tcell.KeyF1, tcell.ModNone, func() bool {
		// Toggle theme details panel
		return true
	})

	// Start the application loop
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %s\n", err)
		os.Exit(1)
	}
}

// createSidebar creates the left sidebar with theme selection
func createSidebar(app *tinytui.Application, width int) *widgets.Pane {
	// Create a pane with a border
	sidebarPane := widgets.NewPane()
	sidebarPane.SetBorder(true, tinytui.BorderSingle, tinytui.DefaultPaneBorderStyle())

	// Create a vertical layout for the sidebar content
	sidebarLayout := tinytui.NewFlexLayout(tinytui.Vertical)
	sidebarLayout.SetGap(1)

	// Create a title text widget
	titleText := widgets.NewText("Theme Selector")
	titleText.SetStyle(tinytui.DefaultTextStyle().Bold(true))

	// Create a list of themes
	themeList := widgets.NewList()
	themeList.SetItems([]string{
		"Default Theme",
		"Borland",
	})

	// Set a callback for when a theme is selected
	themeList.SetOnSelect(func(index int, item string) {
		var themeName tinytui.ThemeName

		switch index {
		case 0:
			themeName = tinytui.ThemeDefault
		case 1:
			themeName = tinytui.ThemeBorland
		default:
			return
		}

		// Apply the selected theme
		success := app.SetTheme(themeName)

		// Add a visual confirmation that theme was changed
		titleText.SetContent(fmt.Sprintf("Theme Selector [%s %s]",
			item,
			map[bool]string{true: "✓", false: "✗"}[success]))

		// Force a redraw
		app.QueueRedraw()
	})

	// Also set an onChange handler to highlight the selection
	themeList.SetOnChange(func(index int, item string) {
		// Update the sidebar title to show which theme is being considered
		titleText.SetContent(fmt.Sprintf("Theme Selector [%s]", item))
	})

	// Add widgets to the sidebar layout
	sidebarLayout.AddChild(titleText, 2, 0)
	// sidebarLayout.AddChild(themeList, 0, 1)

	// Add instructions text at the bottom
	instructionsText := widgets.NewText("Press Enter to select theme")
	instructionsText.SetStyle(tinytui.DefaultTextStyle().Italic(true))
	sidebarLayout.AddChild(instructionsText, 1, 0)

	// Set the layout as the pane's child
	sidebarPane.SetChild(sidebarLayout)

	return sidebarPane
}

// createContentArea creates the main content area with multiple widget examples
func createContentArea() *widgets.Pane {
	// Create main content pane
	contentPane := widgets.NewPane()
	contentPane.SetBorder(true, tinytui.BorderSingle, tinytui.DefaultPaneBorderStyle())

	// Create layout for content
	contentLayout := tinytui.NewFlexLayout(tinytui.Vertical)
	contentLayout.SetGap(1) // Add space between elements

	// Create title for the content area
	titleText := widgets.NewText("Theme Demonstration")
	titleText.SetStyle(tinytui.DefaultTextStyle().Bold(true))

	// Grid demonstration pane
	gridPane := createGridDemo()

	// Text demonstration pane
	textPane := createTextDemo()

	// Button demonstration pane
	buttonPane := createButtonDemo()

	// Add all components to the layout with proper spacing
	// Ensure fixed heights have adequate space for content including borders
	contentLayout.AddChild(titleText, 2, 0)
	contentLayout.AddChild(gridPane, 10, 0)

	// Give text pane a reasonable fixed height instead of proportional sizing
	// This prevents it from being squeezed too small
	contentLayout.AddChild(textPane, 12, 0)

	// Ensure button pane has enough height for its content
	contentLayout.AddChild(buttonPane, 10, 0)

	// Set the content layout as the child of the content pane
	contentPane.SetChild(contentLayout)

	return contentPane
}

// createGridDemo creates a pane with a grid widget
func createGridDemo() *widgets.Pane {
	pane := widgets.NewPane()
	pane.SetBorder(true, tinytui.BorderSingle, tinytui.DefaultPaneBorderStyle())

	// Create a vertical layout
	layout := tinytui.NewFlexLayout(tinytui.Vertical)
	layout.SetGap(1)

	// Create a title
	title := widgets.NewText("Grid Widget Example")
	title.SetStyle(tinytui.DefaultTextStyle().Bold(true))

	// Create a grid
	grid := widgets.NewGrid()
	grid.SetCellSize(10, 1) // Set cell size a bit larger for better readability
	grid.SetPadding(1)      // Add padding for better visibility

	// Create sample data
	data := [][]string{
		{"Column 1", "Column 2", "Column 3", "Column 4"},
		{"Data 1,1", "Data 1,2", "Data 1,3", "Data 1,4"},
		{"Data 2,1", "Data 2,2", "Data 2,3", "Data 2,4"},
		{"Data 3,1", "Data 3,2", "Data 3,3", "Data 3,4"},
	}
	grid.SetCells(data)

	// Add interactivity to the grid
	grid.SetOnSelect(func(row, col int, item string) {
		title.SetContent(fmt.Sprintf("Grid Widget Example - Selected: [%d,%d] %s - Now in interacted state",
			row, col, item))
	})

	// Also update on navigation
	grid.SetOnChange(func(row, col int, item string) {
		title.SetContent(fmt.Sprintf("Grid Widget Example - Focus: [%d,%d] - State persists across widgets",
			row, col))
	})

	stateDesc := widgets.NewText("Selection state persists when focus moves to other widgets")
	layout.AddChild(stateDesc, 1, 0)

	// Add widgets to layout
	layout.AddChild(title, 1, 0)
	layout.AddChild(grid, 0, 1)

	// Set layout as pane's child
	pane.SetChild(layout)

	return pane
}

// createTextDemo creates a pane with text widgets
func createTextDemo() *widgets.Pane {
	pane := widgets.NewPane()
	pane.SetBorder(true, tinytui.BorderSingle, tinytui.DefaultPaneBorderStyle())

	// Create a vertical layout
	layout := tinytui.NewFlexLayout(tinytui.Vertical)
	layout.SetGap(1)

	// Create a title
	title := widgets.NewText("Text Widget Examples")
	title.SetStyle(tinytui.DefaultTextStyle().Bold(true))

	// Create different text widgets with reasonable, fixed heights
	text1 := widgets.NewText("Regular text - The theme affects the default appearance.")

	text2 := widgets.NewText("Long wrapped text example. This text should wrap to multiple lines depending on the container width. The theme's text style will be applied to this widget.")
	text2.SetWrap(true)

	text3 := widgets.NewText("Custom styled text - this has custom styling.")
	text3.SetStyle(tinytui.DefaultTextStyle().Bold(true).Italic(true))

	// Create a list with proper styling
	list := widgets.NewList()
	list.SetItems([]string{
		"List Item 1",
		"List Item 2",
		"List Item 3",
		"List Item 4",
	})

	// State persistence
	stateText := widgets.NewText("Selection state persists when focus changes between widgets")
	stateText.SetStyle(tinytui.DefaultTextStyle().Italic(true))

	// Add interactivity to the list
	list.SetOnSelect(func(index int, item string) {
		// When an item is selected, update the text3 widget to show it
		text3.SetContent(fmt.Sprintf("Selected list item: %s", item))
		text3.SetStyle(tinytui.DefaultTextStyle().Bold(true).Italic(true))
	})

	// Also show focus changes
	list.SetOnChange(func(index int, item string) {
		text2.SetContent(fmt.Sprintf("List focus on: %s - Selected state is maintained when focus moves to other widgets", item))
		text2.SetWrap(true)
	})

	// Add widgets to layout with fixed heights to ensure proper space allocation
	layout.AddChild(title, 1, 0) // Fixed 1 line for title
	layout.AddChild(text1, 1, 0) // Fixed 1 line for regular text
	layout.AddChild(text2, 2, 0) // Fixed 2 lines for wrapped text
	layout.AddChild(text3, 1, 0) // Fixed 1 line for custom text

	// Give the list a fixed height of at least 4 rows to show all items
	// This ensures the list is always visible
	// layout.AddChild(list, 5, 0) // Fixed 5 lines for list (including some margin)

	layout.AddChild(stateText, 1, 0)

	// Set layout as pane's child
	pane.SetChild(layout)

	return pane
}

// createButtonDemo creates a pane with buttons
func createButtonDemo() *widgets.Pane {
	pane := widgets.NewPane()
	pane.SetBorder(true, tinytui.BorderSingle, tinytui.DefaultPaneBorderStyle())

	// Create a vertical layout
	layout := tinytui.NewFlexLayout(tinytui.Vertical)
	layout.SetGap(1)
	// Use AlignStart instead of AlignCenter to avoid pushing buttons to the border
	layout.SetMainAxisAlignment(tinytui.AlignStart)

	// Create a title
	title := widgets.NewText("Button Examples")
	title.SetStyle(tinytui.DefaultTextStyle().Bold(true))

	// Create a horizontal layout for buttons with proper alignment
	buttonLayout := tinytui.NewFlexLayout(tinytui.Horizontal)
	buttonLayout.SetGap(3)                                  // Space between buttons
	buttonLayout.SetMainAxisAlignment(tinytui.AlignCenter)  // Center buttons horizontally
	buttonLayout.SetCrossAxisAlignment(tinytui.AlignCenter) // Center buttons vertically

	// Create buttons with indicators
	button1 := widgets.NewButton("OK")
	button1.SetIndicator('>', widgets.IndicatorLeft) // Show indicator on left

	button2 := widgets.NewButton("Cancel")
	button2.SetIndicator('>', widgets.IndicatorLeft)

	button3 := widgets.NewButton("Help")
	button3.SetIndicator('>', widgets.IndicatorLeft)

	// Demo action and state changes
	button1.SetOnClick(func() {
		button1.SetLabel("Clicked!")
		// State is now StateInteracted from HandleEvent
	})

	button2.SetOnClick(func() {
		button2.SetLabel("Canceled")
		// State is now StateInteracted, no need to reset it manually
	})

	button3.SetOnClick(func() {
		button3.SetLabel("Helping!")
		// State is now StateInteracted from HandleEvent
	})

	// Add description text explaining state management
	stateDesc := widgets.NewText("Tab to another widget - selection state is preserved")

	// Add widgets to layout with proper spacing
	layout.AddChild(title, 1, 0)
	layout.AddChild(stateDesc, 1, 0)

	// Add a fixed height (1) empty text for spacing
	spacer1 := widgets.NewText("")
	layout.AddChild(spacer1, 1, 0)

	// Add buttons to horizontal layout with FIXED WIDTH to ensure visibility
	// Increase widths to ensure buttons are properly visible
	buttonLayout.AddChild(button1, 10, 0) // Increased from 8 to 10
	buttonLayout.AddChild(button2, 12, 0) // Increased from 10 to 12
	buttonLayout.AddChild(button3, 10, 0) // Increased from 8 to 10

	// Increase button layout height for better visibility
	layout.AddChild(buttonLayout, 3, 0) // Fixed height 3

	// Add description text with proper spacing
	spacer2 := widgets.NewText("")
	layout.AddChild(spacer2, 1, 0)

	descText := widgets.NewText("Tab between buttons to see focus styles change")
	layout.AddChild(descText, 1, 0)

	// Set layout as pane's child
	pane.SetChild(layout)

	return pane
}