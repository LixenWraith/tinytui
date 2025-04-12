# TinyTUI

[![Go Reference](https://pkg.go.dev/badge/github.com/LixenWraith/tinytui.svg)](https://pkg.go.dev/github.com/LixenWraith/tinytui)

TinyTUI is a lightweight, flexible terminal UI library for Go applications that focuses on simplicity while providing powerful layout and widget capabilities. It provides a component-based architecture for creating interactive terminal user interfaces with minimal dependencies.

![TinyTUI Demo Screenshot](https://example.com/screenshot.png)

## Features

- **Lightweight**: Minimal dependencies, focusing on core UI functionality
- **Flexible Layout System**: CSS-like flexbox-inspired layouts with alignment controls
- **Rich Widget Set**: Buttons, text areas, grids, lists, panes, and more
- **Comprehensive Theming**: Built-in theme system with prefab themes and custom theme support
- **Styling & Theming**: Comprehensive styling options with colors and text attributes
- **Event Handling**: Keyboard and mouse event handling with focus management
- **Modal Dialogs**: Support for modal interfaces with proper focus handling
- **Multiple Selection Modes**: Single-select (radio button) and multi-select (checkbox) interfaces

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

Here's a minimal "Hello World" application:

```go
package main

import (
	"github.com/LixenWraith/tinytui"
	"github.com/LixenWraith/tinytui/widgets"
)

func main() {
	// Create the application
	app := tinytui.NewApplication()

	// Create a text widget
	text := widgets.NewText("Hello, TinyTUI World!").
		SetStyle(tinytui.DefaultTextStyle().Foreground(tinytui.ColorAqua).Bold(true))

	// Create a button widget
	button := widgets.NewButton("Quit")
	button.SetOnClick(func() {
		app.Stop() // Exit the application when clicked
	})

	// Create a vertical layout to arrange widgets
	layout := tinytui.NewFlexLayout(tinytui.Vertical).
		SetGap(1).
		AddChild(text, 0, 1).
		AddChild(button, 1, 0)

	// Set the layout as the application's root widget
	app.SetRoot(layout, true)

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
- Controls the widget hierarchy
- Handles focus navigation
- Manages themes

```go
app := tinytui.NewApplication()
app.SetRoot(rootWidget, true)
app.SetTheme(tinytui.ThemeTurbo) // Apply a theme
app.Run()
```

### Widgets

Widgets are the building blocks for your UI. All widgets implement the `Widget` interface:

```go
type Widget interface {
    Draw(screen tcell.Screen)
    SetRect(x, y, width, height int)
    GetRect() (x, y, width, height int)
    HandleEvent(event tcell.Event) bool
    Focusable() bool
    Focus()
    Blur()
    SetApplication(app *Application)
    App() *Application
    IsFocused() bool
    Children() []Widget
    Parent() Widget
    SetParent(parent Widget)
    IsVisible() bool
    SetVisible(visible bool)
    PreferredWidth() int
    PreferredHeight() int
    SetState(state WidgetState)
    GetState() WidgetState
    ApplyTheme(theme Theme)
}
```

Built-in widgets include:

- **Text**: Display static or wrapping text
- **Button**: User-clickable button with custom handlers
- **List**: Scrollable list of selectable items
- **Grid**: 2D grid of selectable cells, supporting both single and multi-selection
- **Pane**: Container with optional border and background
- **Sprite**: Fixed graphical element with styled characters

### Layouts

Layouts are special widgets that arrange their children:

```go
// Horizontal layout with spacing between children
hLayout := tinytui.NewFlexLayout(tinytui.Horizontal).
    SetGap(1).
    SetMainAxisAlignment(tinytui.AlignCenter).
    SetCrossAxisAlignment(tinytui.AlignCenter).
    AddChild(button1, 10, 0).           // Fixed width 10
    AddChild(button2, 0, 1).            // Flexible width, 1/3 of remaining space
    AddChildWithAlign(button3, 0, 2, tinytui.AlignEnd)  // 2/3 of remaining space, aligned to bottom
```

Layout options include:
- **Fixed vs. Proportional Sizing**: Control space allocation
- **Gap**: Control spacing between elements
- **Main Axis Alignment**: How items are positioned in the layout direction
- **Cross Axis Alignment**: How items are positioned perpendicular to layout direction

### Widget State Management

TinyTUI widgets have a standardized state system:

```go
// WidgetState enumeration
const (
    StateNormal     WidgetState = iota // Normal, unselected state
    StateSelected                      // Selected but not interacted
    StateInteracted                    // Selected and interacted with
)
```

State can be managed with:

```go
// Get a widget's current state
state := myWidget.GetState()

// Set a widget's state (will trigger redraw if changed)
myWidget.SetState(tinytui.StateSelected)

// Helper methods for state checking
if myWidget.IsSelected() {
    // Widget is either Selected or Interacted
}

if myWidget.IsInteracted() {
    // Widget is specifically in Interacted state
}

// Reset to normal state
myWidget.ResetState()
```

### Grid Selection Modes

Grids can operate in two selection modes:

```go
// Selection mode types
const (
    SingleSelect SelectionMode = iota // Only one item can be selected/interacted at a time
    MultiSelect                       // Multiple items can be selected/interacted
)

// Create a single-select grid (like radio buttons)
themeSelector := widgets.NewGrid()
themeSelector.SetSelectionMode(widgets.SingleSelect)
themeSelector.SetCells([][]string{
    {"Light Theme"},
    {"Dark Theme"},
    {"Turbo"},
})

// Create a multi-select grid (like checkboxes)
optionsGrid := widgets.NewGrid()
optionsGrid.SetSelectionMode(widgets.MultiSelect)
optionsGrid.SetCells([][]string{
    {"Enable Logging"},
    {"Show Line Numbers"},
    {"Auto Save"},
})

// Get all interacted cells from a grid
interactedCells := optionsGrid.GetInteractedCells()
```

### Event Handling

Events are processed through the widget hierarchy. Focused widgets get priority:

```go
button := widgets.NewButton("Click Me")
button.SetOnClick(func() {
    // Handler when the button is activated (e.g., Enter key)
})

// For custom key handling:
button.SetKeybinding(tcell.KeyRune, tcell.ModNone, func() bool {
    // Handle a specific key
    return true // Return true if handled
})
```

## Theme System

TinyTUI includes a comprehensive theme system that provides consistent styling across your application:

### Built-in Themes

TinyTUI comes with several built-in themes:

- **Default**: Basic theme using terminal defaults (transparent background)
- **Turbo**: Classic high-contrast blue background with white text reminiscent of 1990s DOS applications (non-transparent background)

### Using Themes

```go
// At application startup
app := tinytui.NewApplication()
app.SetTheme(tinytui.ThemeTurbo)

// Or change the theme at runtime
app.SetTheme(tinytui.ThemeDefault)
```

### Creating Custom Themes

Create your own theme by implementing the Theme interface:

```go
type CustomTheme struct {
    tinytui.BaseTheme
}

// Override any methods needed...
func (t *CustomTheme) Name() tinytui.ThemeName {
    return "custom-theme"
}

func (t *CustomTheme) TextStyle() tinytui.Style {
    return tinytui.DefaultStyle.
        Foreground(tinytui.ColorWhite).
        Background(tinytui.ColorNavy)
}

// Register the theme
tinytui.RegisterTheme(&CustomTheme{})
```

### Theme Interface

The Theme interface provides methods for retrieving styles for different widget states:

```go
type Theme interface {
    Name() ThemeName
    
    // Text styles
    TextStyle() Style
    TextSelectedStyle() Style
    
    // Grid styles based on state
    GridStyle() Style
    GridFocusedStyle() Style
    GridSelectedStyle() Style
    GridInteractedStyle() Style
    GridFocusedSelectedStyle() Style
    GridFocusedInteractedStyle() Style
    
    // Pane styles
    PaneStyle() Style
    PaneBorderStyle() Style
    PaneFocusBorderStyle() Style
    
    // Default dimensions
    DefaultCellWidth() int
    DefaultCellHeight() int
    
    // Style indicators and padding
    IndicatorColor() Color
    DefaultPadding() int
}
```

### Widget Theme Integration

Widgets automatically use the current theme's styles. You can also access theme styles directly:

```go
// Create a widget with themed style
text := widgets.NewText("Themed Text")

// Override with a custom style based on theme
myGrid := widgets.NewGrid().
    SetStyle(tinytui.GetTheme().GridStyle().Bold(true))
```

## Layout System

TinyTUI's layout system is inspired by CSS Flexbox:

### Orientation

- `Horizontal`: Children arranged in a row (left to right)
- `Vertical`: Children arranged in a column (top to bottom)

### Sizing

- **Fixed Size**: Exact width/height in characters
- **Proportional Size**: Fraction of remaining space

### Alignment

- `AlignStart`: Items at the beginning of the axis
- `AlignCenter`: Items centered along the axis
- `AlignEnd`: Items at the end of the axis
- `AlignStretch`: Items stretch to fill the axis (default)

### Example

```go
// Create a centered, padded dialog box
dialogContent := tinytui.NewFlexLayout(tinytui.Vertical).
    SetGap(1).
    SetMainAxisAlignment(tinytui.AlignCenter).
    AddChild(titleText, 1, 0).
    AddChild(messageText, 0, 1).
    AddChildWithAlign(buttonsLayout, 3, 0, tinytui.AlignCenter)

dialogPane := widgets.NewPane().
    SetBorder(true, tinytui.BorderDouble, tinytui.DefaultPaneBorderStyle()).
    SetChild(dialogContent)
```

## Styling

TinyTUI provides comprehensive styling options:

```go
// Create styled text
titleText := widgets.NewText("Application Title").
    SetStyle(tinytui.DefaultTextStyle().
        Foreground(tinytui.ColorWhite).
        Background(tinytui.ColorBlue).
        Bold(true))

// Style a grid
grid := widgets.NewGrid().
    SetStyle(tinytui.DefaultGridStyle().
        Foreground(tinytui.ColorWhite).
        Background(tinytui.ColorBlue)).
    SetSelectedStyle(tinytui.DefaultGridSelectedStyle().
        Foreground(tinytui.ColorBlue).
        Background(tinytui.ColorWhite).
        Bold(true)).
    SetInteractedStyle(tinytui.DefaultGridInteractedStyle().
        Bold(true))
```

Style options include:
- Foreground and background colors
- Text attributes (bold, italic, underline, reverse, etc.)
- Border styles (single, double, solid)

## Modal Dialogs

Modal dialogs temporarily capture focus:

```go
// Create a modal dialog
confirmDialog := createConfirmDialog() // Returns a Widget
confirmDialog.SetVisible(false) // Initially hidden

// Show modal when needed
app.Dispatch(func(app *tinytui.Application) {
    confirmDialog.SetVisible(true)
    app.SetModalRoot(confirmDialog)
    app.SetFocus(findFirstFocusableIn(confirmDialog))
})

// Close modal
app.Dispatch(func(app *tinytui.Application) {
    app.ClearModalRoot()
    confirmDialog.SetVisible(false)
})
```

## Concurrency & Thread Safety

TinyTUI's event-driven architecture ensures thread safety. Use `Dispatch` to interact with the UI from goroutines:

```go
go func() {
    // Do background work...

    // Update UI safely
    app.Dispatch(func(app *tinytui.Application) {
        statusText.SetContent("Task completed!")
        app.QueueRedraw()
    })
}()
```

## Best Practices

1. **Dispatch for UI Updates**: Always use `app.Dispatch()` for UI changes from outside the main event loop
2. **Preserve Focus State**: Save and restore focus when showing/hiding widgets
3. **Visibility Before Focus**: Only focus visible and focusable widgets
4. **Widget Reuse**: Structure your code to reuse widget trees for common patterns
5. **Error Handling**: Always check errors returned by `app.Run()`
6. **Consistent Theming**: Use the theme system for consistent styling across your application
7. **Layout Spacing**: Use spacer widgets and proper gaps to ensure good UI spacing
8. **Selection Modes**: Use single-select for exclusive choices, multi-select for independent options

## Common Patterns

### Data Binding

```go
type Model struct {
    Name string
    Email string
    // other fields...
}

func bindModelToForm(model *Model, nameText, emailText *widgets.Text) {
    // Update UI from model
    nameText.SetContent(model.Name)
    emailText.SetContent(model.Email)

    // Update model from UI handled in event handlers
}
```

### Grid Selections and Interactions

```go
// Create a multi-select grid for todo items
todoGrid := widgets.NewGrid()
todoGrid.SetSelectionMode(widgets.MultiSelect)

// Handle toggling item state
todoGrid.SetOnSelect(func(row, col int, item string) {
    // Toggle completion when the user selects an item
    if todoGrid.IsCellInteracted(row, col) {
        statusText.SetContent(fmt.Sprintf("Marked '%s' as completed", item))
    } else {
        statusText.SetContent(fmt.Sprintf("Marked '%s' as incomplete", item))
    }
    
    // Update visualization
    updateTodoDisplay()
})

// Get all completed items
func getCompletedItems() []string {
    var completed []string
    for _, cell := range todoGrid.GetInteractedCells() {
        item := todoItems[cell.Row]
        completed = append(completed, item)
    }
    return completed
}
```

### Custom Widgets

Create custom widgets by embedding `BaseWidget` and implementing `ThemedWidget`:

```go
type ColorPicker struct {
tinytui.BaseWidget
colors []tinytui.Color
selected int
onChange func(tinytui.Color)
style tinytui.Style
}

func NewColorPicker(colors []tinytui.Color) *ColorPicker {
cp := &ColorPicker{
colors: colors,
selected: 0,
style: tinytui.DefaultStyle,
}
cp.SetVisible(true)
return cp
}

// Implement required methods: Draw, HandleEvent, etc.

// Implement ApplyTheme for theme support
func (cp *ColorPicker) ApplyTheme(theme tinytui.Theme) {
cp.style = theme.TextStyle()
if app := cp.App(); app != nil {
app.QueueRedraw()
}
}
```

## License

BSD-3

## Acknowledgments

- [tcell](https://github.com/gdamore/tcell) - Terminal handling foundation
- [go-runewidth](https://github.com/mattn/go-runewidth) - Unicode width calculation
- This project was developed by extensive use of LLM (Gemini Code Assist, Claude Sonnet 3.7) with human iteration.