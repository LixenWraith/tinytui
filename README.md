# TinyTUI

[![Go Reference](https://pkg.go.dev/badge/github.com/LixenWraith/tinytui.svg)](https://pkg.go.dev/github.com/LixenWraith/tinytui)

TinyTUI is a lightweight terminal UI library for Go applications that focuses on simplicity while providing powerful layout and component capabilities. It provides a component-based architecture for creating interactive terminal user interfaces with minimal dependencies.

## Features

- **Lightweight**: Minimal dependencies, focusing on core UI functionality
- **Flexible Layout System**: Support for horizontal and vertical layouts
- **Component Architecture**: Text, TextInput, Grid, and other components
- **Theming Support**: Customizable styling with color and text attributes
- **Keyboard Navigation**: Built-in support for tab navigation and Alt+number pane selection
- **Automatic Focus Management**: Intelligent focus handling and navigation

## Installation

```bash
go get github.com/LixenWraith/tinytui
```

TinyTUI requires Go 1.23 or newer.

## Dependencies

TinyTUI has minimal external dependencies:

- [github.com/gdamore/tcell/v2](https://github.com/gdamore/tcell): Terminal handling, event processing, and rendering
- [github.com/mattn/go-runewidth](https://github.com/mattn/go-runewidth): Unicode width calculation for proper text alignment

## Quick Start

Here's a minimal application with text input and display:

```go
package main

import (
    "github.com/gdamore/tcell/v2"
    "github.com/LixenWraith/tinytui"
)

func main() {
    // Create a new application
    app := tinytui.NewApplication()

    // Create a vertical layout
    mainLayout := tinytui.NewLayout(tinytui.Vertical)

    // Create panes for input and display
    inputPane := tinytui.NewPane()
    inputPane.SetTitle("Input")

    displayPane := tinytui.NewPane()
    displayPane.SetTitle("Display")

    // Create a text input component
    textInput := tinytui.NewTextInput()

    // Create text display component
    displayText := tinytui.NewText("Input will appear here")

    // Set components as pane children
    inputPane.SetChild(textInput)
    displayPane.SetChild(displayText)

    // Set up handler for text input submission
    textInput.SetOnSubmit(func(text string) {
        // Update display text with submitted input
        displayText.SetContent("You entered: " + text)
        
        // Clear the input field
        textInput.SetText("")
    })

    // Add panes to layout
    mainLayout.AddPane(inputPane, tinytui.Size{Proportion: 1})
    mainLayout.AddPane(displayPane, tinytui.Size{Proportion: 1})

    // Set the layout as the application's main layout
    app.SetLayout(mainLayout)

    // Focus the text input component initially
    app.SetFocus(textInput)

    // Register key handler for quitting with ESC
    app.RegisterKeyHandler(tcell.KeyEscape, tcell.ModNone, func() bool {
        app.Stop()
        return true
    })

    // Run the application
    if err := app.Run(); err != nil {
        panic(err)
    }
}
```

## Core Concepts

### Application

The `Application` is the main controller that:
- Manages the screen
- Processes events
- Controls the component hierarchy
- Handles focus navigation

```go
app := tinytui.NewApplication()
app.SetLayout(mainLayout)
app.Run()
```

### Layouts

Layouts organize components on the screen. TinyTUI supports:

- **Horizontal Layout**: Components arranged side by side
- **Vertical Layout**: Components arranged top to bottom

```go
// Create a new layout
layout := tinytui.NewLayout(tinytui.Vertical)

// Add panes to the layout
layout.AddPane(firstPane, tinytui.Size{Proportion: 2})  // Takes 2/3 of space
layout.AddPane(secondPane, tinytui.Size{Proportion: 1}) // Takes 1/3 of space

// Or use fixed size
layout.AddPane(headerPane, tinytui.Size{FixedSize: 3})  // 3 rows tall
```

### Panes

Panes are containers that can hold components or other layouts:

```go
// Create a pane with a border and title
pane := tinytui.NewPane()
pane.SetTitle("My Pane")
pane.SetBorder(tinytui.BorderSingle, tinytui.DefaultStyle)

// Add a component to the pane
pane.SetChild(textComponent)

// Or add a nested layout
pane.SetChild(nestedLayout)
```

### Components

Components are the interactive elements in your UI:

- **Text**: Display static or wrapping text
- **TextInput**: Single-line text entry with cursor
- **Grid**: 2D grid of selectable cells with navigation

```go
// Text component
text := tinytui.NewText("Hello, World!")
text.SetStyle(tinytui.DefaultStyle.Foreground(tinytui.ColorGreen))

// Text input
input := tinytui.NewTextInput()
input.SetMaxLength(100)
input.SetOnSubmit(func(text string) {
    // Handle text submission
})

// Grid
grid := tinytui.NewGrid()
grid.SetCells([][]string{
    {"Option 1", "Value 1"},
    {"Option 2", "Value 2"},
})
grid.SetOnSelect(func(row, col int, item string) {
    // Handle selection
})
```

### Event Handling

Events can be handled at the application level or by individual components:

```go
// Application-level key handler
app.RegisterKeyHandler(tcell.KeyCtrlS, tcell.ModNone, func() bool {
    // Handle Ctrl+S
    return true
})

// Component event handler
textInput.SetOnSubmit(func(text string) {
    // Handle text submission
})

grid.SetOnChange(func(row, col int, item string) {
    // Handle grid selection change
})
```

### Focus Management

TinyTUI manages focus automatically:

- Tab/Shift+Tab to navigate between focusable components
- Alt+Number to jump to specific panes
- Application provides explicit focus control

```go
// Set focus to a component
app.SetFocus(myTextInput)

// Check if component has focus
if myComponent.IsFocused() {
    // Component has focus
}
```

### Styling

TinyTUI provides comprehensive styling options:

```go
// Create a style
style := tinytui.DefaultStyle.
    Foreground(tinytui.ColorWhite).
    Background(tinytui.ColorBlue).
    Bold(true)

// Apply to a component
text.SetStyle(style)

// Apply to a pane border
pane.SetBorder(tinytui.BorderDouble, style)
```

Style options include:
- Foreground and background colors
- Text attributes (bold, italic, underline, reverse, etc.)
- Border styles (none, single, double, solid)

## Concurrency & Thread Safety

TinyTUI's event-driven architecture ensures thread safety. Use the command pattern to update the UI from goroutines:

```go
go func() {
    // Do background work...

    // Update UI safely
    app.Dispatch(&tinytui.UpdateTextCommand{
        Target:  statusText,
        Content: "Task completed!",
    })
}()
```

## Advanced Features

### Working with Sprites

Sprites allow for custom character-based graphics:

```go
// Create a sprite
sprite := tinytui.NewSprite([][]tinytui.SpriteCell{
    {
        {Rune: '┌', Style: borderStyle},
        {Rune: '─', Style: borderStyle},
        {Rune: '┐', Style: borderStyle},
    },
    {
        {Rune: '│', Style: borderStyle},
        {Rune: ' ', Style: contentStyle},
        {Rune: '│', Style: borderStyle},
    },
    {
        {Rune: '└', Style: borderStyle},
        {Rune: '─', Style: borderStyle},
        {Rune: '┘', Style: borderStyle},
    },
})
```

### Custom Grid Rendering

Create interactive grids with custom styling:

```go
grid := tinytui.NewGrid()
grid.SetCells([][]string{
    {"Name", "Value", "Type"},
    {"config.json", "42.5 KB", "File"},
    {"docs", "", "Directory"},
})
grid.SetOnChange(func(row, col int, item string) {
    // Handle selection change
})
grid.SetAutoWidth(true)
grid.SetCellSize(0, 1) // Auto width, 1 line height
```

## Best Practices

1. **Layout Structure**: Design your layout hierarchy before implementing
2. **Focus Management**: Let the application handle focus navigation
3. **Event Handling**: Use component-specific events when possible
4. **Styling**: Use consistent styling across your application
5. **Error Handling**: Always check errors returned by `app.Run()`

## License

BSD-3

## Dependencies

- [tcell](https://github.com/gdamore/tcell) - Terminal handling foundation
- [go-runewidth](https://github.com/mattn/go-runewidth) - Unicode width calculation