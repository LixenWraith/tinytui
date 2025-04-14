# TinyTUI

A lightweight Terminal User Interface (TUI) library for Go applications, built on top of [tcell](https://github.com/gdamore/tcell). TinyTUI provides a component-based architecture for building interactive terminal applications with minimal dependencies.

## Features

- **Component-Based Architecture**: Modular design with reusable UI components
- **Flexible Layout System**: Arrange components using horizontal and vertical layouts with various sizing options
- **Event Handling**: Process keyboard input, handle focus navigation, and dispatch commands
- **Theming Support**: Customize colors and styles with built-in themes (Default and Turbo)
- **Rich Component Library**:
    - Text display with optional wrapping and alignment
    - Text input fields with cursor navigation and editing
    - Grid for data display with selection and interaction
    - Custom sprite rendering for graphics
    - Panes with optional borders and titles
- **Command Pattern**: Decouple UI events from application logic
- **Focus Management**: Tab navigation and Alt+Number quick access to panes

## Installation

```bash
go get github.com/LixenWraith/tinytui
```

## Quick Start

```go
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
	button.SetCells([][]string{{" Submit "}})
	button.SetCellSize(10, 1)
	button.SetOnSelect(func(r, c int, i string) {
		// Handle button click
	})

	// Create panes and set content
	headerPane := tinytui.NewPane()
	headerPane.SetChild(header)

	inputPane := tinytui.NewPane()
	inputPane.SetTitle("Input")
	inputPane.SetChild(input)

	buttonPane := tinytui.NewPane()
	buttonPane.SetChild(button)

	// Create layout and arrange panes
	layout := tinytui.NewLayout(tinytui.Vertical)
	layout.AddPane(headerPane, tinytui.Size{FixedSize: 1})
	layout.AddPane(inputPane, tinytui.Size{Proportion: 1})
	layout.AddPane(buttonPane, tinytui.Size{FixedSize: 1})

	// Set application layout
	app.SetLayout(layout)

	// Set initial focus
	app.Dispatch(&tinytui.FocusCommand{Target: input})

	// Run application
	if err := app.Run(); err != nil {
		panic(err)
	}
}
```

## Core Concepts

### Application

The `Application` is the root object that manages the screen, event loop, and component hierarchy. It handles focus management, command dispatch, and rendering.

```go
app := tinytui.NewApplication()
app.SetScreenMode(tinytui.ScreenAlternate) // Use alternate screen buffer
app.SetTheme(tinytui.GetTheme())           // Set theme
app.SetLayout(mainLayout)                  // Set root layout
app.Run()                                  // Start event loop
```

### Components

Components are UI elements that implement the `Component` interface. TinyTUI provides several built-in components:

- **Text**: Display non-editable text content
- **TextInput**: Single-line text entry field
- **Grid**: 2D grid of selectable and potentially interactive cells
- **Sprite**: Display character-based graphics

Components share common behavior through the `BaseComponent` struct, which provides default implementations for visibility, focus, and state management.

### Panes and Layouts

Panes are containers that hold a single child (Component or Layout) and provide borders, titles, and navigation indices. Layouts arrange multiple panes in a horizontal or vertical orientation with flexible sizing options.

```go
// Create a pane with a border and title
pane := tinytui.NewPane()
pane.SetTitle("My Component")
pane.SetBorder(tinytui.BorderSingle, tinytui.DefaultPaneBorderStyle())
pane.SetChild(component)

// Create a vertical layout with multiple panes
layout := tinytui.NewLayout(tinytui.Vertical)
layout.SetGap(1) // Set gap between panes
layout.AddPane(pane1, tinytui.Size{FixedSize: 3})           // Fixed height of 3
layout.AddPane(pane2, tinytui.Size{Proportion: 1})          // Proportion of remaining space
layout.AddPane(pane3, tinytui.Size{FixedSize: 5})           // Fixed height of 5
```

### Event Handling

TinyTUI uses an event-driven architecture for handling user input:

1. **Event Dispatch**: The application dispatches tcell events to components
2. **Focus Handling**: Focused components get first opportunity to handle events
3. **Command Pattern**: Components can issue commands to be executed by the application
4. **Key Binding**: Register handlers for specific keys globally

```go
// Add component event handler
myGrid.SetOnSelect(func(row, col int, item string) {
    // Handle selection
})

// Register global key handler
app.RegisterRuneHandler('q', 0, func() bool {
    app.Stop()
    return true
})

// Dispatch a command
app.Dispatch(&tinytui.FocusCommand{Target: myInput})
```

### Styling and Theming

TinyTUI provides a theming system for consistent styling across the application:

```go
// Use built-in themes
tinytui.SetTheme(tinytui.ThemeTurbo)  // Switch to Turbo theme (blue background)
app.SetTheme(tinytui.GetTheme())      // Apply to application

// Create custom styles
style := tinytui.DefaultStyle.Foreground(tinytui.ColorRed).Bold(true)
myText.SetStyle(style)
```

## Component Reference

### Text

```go
text := tinytui.NewText("Hello, World!")
text.SetContent("New content")              // Update text
text.SetAlignment(tinytui.AlignTextCenter)  // Set text alignment
text.SetWrap(true)                          // Enable text wrapping
text.SetStyle(myStyle)                      // Set text style
```

### TextInput

```go
input := tinytui.NewTextInput()
input.SetText("Initial value")              // Set text content
input.SetMasked(true, '*')                  // Password masking
input.SetMaxLength(10)                      // Limit input length
input.SetOnChange(func(text string) {       // Text change handler
    // Handle text change
})
input.SetOnSubmit(func(text string) {       // Enter key handler
    // Handle submission
})
```

### Grid

```go
grid := tinytui.NewGrid()
grid.SetCells([][]string{                   // Set grid cell content
    {"Row 1, Col 1", "Row 1, Col 2"},
    {"Row 2, Col 1", "Row 2, Col 2"},
})
grid.SetCellSize(15, 1)                     // Set cell size
grid.SetSelectionMode(tinytui.MultiSelect)  // Enable multi-selection
grid.SetIndicator('>', true)                // Set selection indicator
grid.SetOnChange(func(row, col int, item string) {
    // Handle selection change
})
grid.SetOnSelect(func(row, col int, item string) {
    // Handle cell activation (Enter/Space key)
})
```

### Sprite

```go
sprite := tinytui.NewSprite(nil)
sprite.Resize(10, 5)                         // Set sprite dimensions
sprite.SetCellsFromStrings([]string{         // Set sprite content
    "╔════╗",
    "║ICON║",
    "╚════╝",
}, myStyle)
```

## Layout System

TinyTUI's layout system arranges panes in horizontal or vertical orientations with flexible sizing:

- **Fixed Size**: Allocate a specific number of rows or columns
- **Proportional**: Allocate a proportion of the remaining space
- **Gap**: Set spacing between panes
- **Alignment**: Control alignment along main and cross axes

```go
// Create a horizontal layout with different sizing options
layout := tinytui.NewLayout(tinytui.Horizontal)
layout.SetGap(1)
layout.SetMainAxisAlignment(tinytui.AlignCenter)
layout.SetCrossAxisAlignment(tinytui.AlignStretch)
layout.AddPane(pane1, tinytui.Size{FixedSize: 20})         // Fixed width of 20
layout.AddPane(pane2, tinytui.Size{Proportion: 2})         // 2/3 of remaining width
layout.AddPane(pane3, tinytui.Size{Proportion: 1})         // 1/3 of remaining width
```

## Advanced Usage

### Navigation Indices

Panes can be assigned navigation indices (1-10) to allow quick access with Alt+Number keys:

```go
// Navigation indices are automatically assigned to focusable panes
app.SetShowPaneIndices(true)  // Show indices in pane borders
```

### Command Pattern

Commands allow decoupling UI events from application logic:

```go
// Create a custom command
type MyCommand struct {
    Param string
}

func (c *MyCommand) Execute(app *Application) {
    // Command implementation
}

// Dispatch the command
app.Dispatch(&MyCommand{Param: "value"})
```

### Custom Components

Create custom components by implementing the `Component` interface or embedding `BaseComponent`:

```go
type MyComponent struct {
    tinytui.BaseComponent
    // Custom fields
}

func NewMyComponent() *MyComponent {
    return &MyComponent{
        BaseComponent: tinytui.NewBaseComponent(),
    }
}

// Implement Component interface methods
func (m *MyComponent) Draw(screen tcell.Screen) {
    // Drawing logic
}

func (m *MyComponent) HandleEvent(event tcell.Event) bool {
    // Event handling logic
    return false
}
```

## Example Programs

The package includes two example programs demonstrating various features:

1. `main.go`: A comprehensive demo showcasing layouts, themes, input handling, and component interactions
2. `main (1).go`: A focused example demonstrating navigation and component indexing

## Dependencies

- [github.com/gdamore/tcell/v2](https://github.com/gdamore/tcell): Terminal cell handling
- [github.com/mattn/go-runewidth](https://github.com/mattn/go-runewidth): Proper handling of wide characters

## License

BSD-3