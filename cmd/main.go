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
		"Tokyo Night",
		"Catppuccin Mocha",
		"Borland",
	})

	// Set a callback for when a theme is selected
	themeList.SetOnSelect(func(index int, item string) {
		var themeName tinytui.ThemeName

		switch index {
		case 0:
			themeName = tinytui.ThemeDefault
		case 1:
			themeName = tinytui.ThemeTokyoNight
		case 2:
			themeName = tinytui.ThemeCatppuccinMocha
		case 3:
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
	sidebarLayout.AddChild(themeList, 0, 1)

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
	contentLayout.AddChild(titleText, 2, 0)
	contentLayout.AddChild(gridPane, 10, 0)
	contentLayout.AddChild(textPane, 0, 1)
	contentLayout.AddChild(buttonPane, 8, 0) // Increased height for buttons

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
		// Change the title to show the selected cell
		title.SetContent(fmt.Sprintf("Grid Widget Example - Selected: [%d,%d] %s",
			row, col, item))
	})

	// Also update on navigation
	grid.SetOnChange(func(row, col int, item string) {
		// Update the title to show the highlighted cell
		title.SetContent(fmt.Sprintf("Grid Widget Example - Focus: [%d,%d]",
			row, col))
	})

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

	// Create different text widgets
	text1 := widgets.NewText("Regular text - The theme affects the default appearance.")

	text2 := widgets.NewText("Long wrapped text example. This text should wrap to multiple lines depending on the container width. The theme's text style will be applied to this widget.")
	text2.SetWrap(true)

	text3 := widgets.NewText("Custom styled text - this has custom styling.")
	text3.SetStyle(tinytui.DefaultTextStyle().Bold(true).Italic(true))

	// Create a list
	list := widgets.NewList()
	list.SetItems([]string{
		"List Item 1",
		"List Item 2",
		"List Item 3",
		"List Item 4",
	})

	// Add interactivity to the list
	list.SetOnSelect(func(index int, item string) {
		// When an item is selected, update the text3 widget to show it
		text3.SetContent(fmt.Sprintf("Selected list item: %s", item))
		text3.SetStyle(tinytui.DefaultTextStyle().Bold(true).Italic(true))
	})

	// Also show focus changes
	list.SetOnChange(func(index int, item string) {
		// When focus changes, update the text2 widget
		text2.SetContent(fmt.Sprintf("List focus changed to: %s - Use arrow keys to navigate, Enter to select", item))
		text2.SetWrap(true)
	})

	// Add widgets to layout
	layout.AddChild(title, 1, 0)
	layout.AddChild(text1, 1, 0)
	layout.AddChild(text2, 4, 0)
	layout.AddChild(text3, 1, 0)
	layout.AddChild(list, 0, 1)

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
	layout.SetMainAxisAlignment(tinytui.AlignCenter) // Center align content vertically

	// Create a title
	title := widgets.NewText("Button Examples")
	title.SetStyle(tinytui.DefaultTextStyle().Bold(true))

	// Create a horizontal layout for buttons with proper alignment
	buttonLayout := tinytui.NewFlexLayout(tinytui.Horizontal)
	buttonLayout.SetGap(3)                                  // More space between buttons
	buttonLayout.SetMainAxisAlignment(tinytui.AlignCenter)  // Center buttons horizontally
	buttonLayout.SetCrossAxisAlignment(tinytui.AlignCenter) // Center buttons vertically

	// Create buttons
	button1 := widgets.NewButton("OK")
	button2 := widgets.NewButton("Cancel")
	button3 := widgets.NewButton("Help")

	// Add actions to buttons as a demo
	button1.SetOnClick(func() {
		// Just for demonstration - we'll make it do something visible
		button1.SetLabel("Clicked!")
	})

	button2.SetOnClick(func() {
		button2.SetLabel("Canceled")
	})

	button3.SetOnClick(func() {
		button3.SetLabel("Helping!")
	})

	// Add buttons to horizontal layout with proper sizing
	// Using proportion instead of fixed size gives better flexibility
	buttonLayout.AddChild(button1, 0, 1)
	buttonLayout.AddChild(button2, 0, 1)
	buttonLayout.AddChild(button3, 0, 1)

	// Add spacing before buttons (empty text as spacer)
	spacer1 := widgets.NewText("")

	// Add widgets to main layout with better spacing
	layout.AddChild(title, 1, 0)
	layout.AddChild(spacer1, 1, 0)      // Add space before buttons
	layout.AddChild(buttonLayout, 0, 1) // Use proportion for button layout

	// Add space after buttons
	spacer2 := widgets.NewText("")
	layout.AddChild(spacer2, 1, 0)

	// Add description text
	descText := widgets.NewText("Tab between buttons to see focus styles change")
	layout.AddChild(descText, 1, 0)

	// Set layout as pane's child
	pane.SetChild(layout)

	return pane
}