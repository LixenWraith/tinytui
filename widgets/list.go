// widgets/list.go
package widgets

import (
	"github.com/LixenWraith/tinytui"
	"github.com/gdamore/tcell/v2"
)

// List displays a scrollable list of text items.
// It's essentially a 1-column Grid with specialized rendering.
type List struct {
	grid *Grid // Use Grid as the underlying implementation

	onChange func(int, string) // Callback when the selected index changes
	onSelect func(int, string) // Callback when an item is selected (e.g., Enter pressed)
}

// NewList creates a new List widget.
func NewList() *List {
	// Create an underlying grid
	grid := NewGrid()

	l := &List{
		grid: grid,
	}

	// Set up grid to handle list's callbacks
	grid.SetOnChange(func(row, col int, item string) {
		if l.onChange != nil {
			l.onChange(row, item)
		}
	})

	grid.SetOnSelect(func(row, col int, item string) {
		if l.onSelect != nil {
			l.onSelect(row, item)
		}
	})

	return l
}

// SetItems replaces the current list items with a new slice of strings.
func (l *List) SetItems(items []string) *List {
	// Convert items to grid cells (single column)
	cells := make([][]string, len(items))
	for i, item := range items {
		cells[i] = []string{item}
	}

	l.grid.SetCells(cells)
	return l
}

// SetStyle sets the style for non-selected list items.
func (l *List) SetStyle(style tinytui.Style) *List {
	l.grid.SetStyle(style)
	return l
}

func (l *List) SetSelectedStyle(style tinytui.Style) *List {
	l.grid.SetSelectedStyle(style)
	return l
}

func (l *List) SetInteractedStyle(style tinytui.Style) *List {
	l.grid.SetInteractedStyle(style)
	return l
}

func (l *List) SetFocusedStyle(style tinytui.Style) *List {
	l.grid.SetFocusedStyle(style)
	return l
}

func (l *List) SetFocusedSelectedStyle(style tinytui.Style) *List {
	l.grid.SetFocusedSelectedStyle(style)
	return l
}

func (l *List) SetFocusedInteractedStyle(style tinytui.Style) *List {
	l.grid.SetFocusedInteractedStyle(style)
	return l
}

// ApplyTheme applies the provided theme to the List widget
func (l *List) ApplyTheme(theme tinytui.Theme) {
	l.SetStyle(theme.ListStyle())
	l.SetSelectedStyle(theme.ListSelectedStyle())
	l.SetInteractedStyle(theme.ListInteractedStyle())
	l.SetFocusedStyle(theme.ListFocusedStyle())
	l.SetFocusedSelectedStyle(theme.ListFocusedSelectedStyle())
	l.SetFocusedInteractedStyle(theme.ListFocusedInteractedStyle())
	l.grid.SetPadding(theme.DefaultPadding())
}

// SetOnChange sets a callback function that triggers when selection changes
func (l *List) SetOnChange(handler func(index int, item string)) *List {
	l.onChange = handler
	return l
}

// SetOnSelect sets a callback function that triggers on item selection
func (l *List) SetOnSelect(handler func(index int, item string)) *List {
	l.onSelect = handler
	return l
}

// SelectedIndex returns the index of the currently selected item.
func (l *List) SelectedIndex() int {
	row, _ := l.grid.SelectedIndex()
	return row
}

// SelectedItem returns the string of the currently selected item.
func (l *List) SelectedItem() string {
	return l.grid.SelectedItem()
}

// SetSelectedIndex programmatically sets the selected item index.
func (l *List) SetSelectedIndex(index int) *List {
	l.grid.SetSelectedIndex(index, 0)
	return l
}

// Draw renders the list using the underlying grid.
func (l *List) Draw(screen tcell.Screen) {
	l.grid.Draw(screen)
}

// SetRect updates the widget's dimensions.
func (l *List) SetRect(x, y, width, height int) {
	l.grid.SetRect(x, y, width, height)
}

// Focusable indicates if the list can receive focus.
func (l *List) Focusable() bool {
	return l.grid.Focusable()
}

// HandleEvent handles keyboard events.
func (l *List) HandleEvent(event tcell.Event) bool {
	return l.grid.HandleEvent(event)
}

// Focus passes focus to the underlying grid.
func (l *List) Focus() {
	l.grid.Focus()
}

// Blur removes focus from the underlying grid.
func (l *List) Blur() {
	l.grid.Blur()
}

// SetVisible sets the widget's visibility state.
func (l *List) SetVisible(visible bool) {
	l.grid.SetVisible(visible)
}

// IsVisible returns whether the widget is visible.
func (l *List) IsVisible() bool {
	return l.grid.IsVisible()
}

// SetApplication sets the application pointer.
func (l *List) SetApplication(app *tinytui.Application) {
	l.grid.SetApplication(app)
}

// App returns the application pointer.
func (l *List) App() *tinytui.Application {
	return l.grid.App()
}

// Parent returns the widget's container.
func (l *List) Parent() tinytui.Widget {
	return l.grid.Parent()
}

// SetParent sets the widget's container.
func (l *List) SetParent(parent tinytui.Widget) {
	l.grid.SetParent(parent)
}

// Children returns nil as List doesn't contain children.
func (l *List) Children() []tinytui.Widget {
	return nil
}

// GetState returns the widget's state.
func (l *List) GetState() tinytui.WidgetState {
	return l.grid.GetState()
}

// SetState updates the widget's state.
func (l *List) SetState(state tinytui.WidgetState) {
	l.grid.SetState(state)
}

// IsSelected returns whether the widget is selected.
func (l *List) IsSelected() bool {
	return l.grid.IsSelected()
}

// IsInteracted returns whether the widget is interacted.
func (l *List) IsInteracted() bool {
	return l.grid.IsInteracted()
}

// ResetState resets the widget's state to StateNormal.
func (l *List) ResetState() {
	l.grid.ResetState()
}

// GetRect returns the widget's dimensions.
func (l *List) GetRect() (x, y, width, height int) {
	return l.grid.GetRect()
}

// IsFocused returns whether the widget has focus.
func (l *List) IsFocused() bool {
	return l.grid.IsFocused()
}