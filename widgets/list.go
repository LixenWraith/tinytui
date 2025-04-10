// widgets/list.go
package widgets

import (
	"sync"

	"github.com/LixenWraith/tinytui"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// List displays a scrollable list of text items.
type List struct {
	tinytui.BaseWidget
	mu            sync.RWMutex
	items         []string          // The items to display in the list
	selectedIndex int               // Index of the currently selected item (-1 if empty or no selection)
	topIndex      int               // Index of the item displayed at the top row
	style         tinytui.Style     // Style for non-selected items
	selectedStyle tinytui.Style     // Style for the selected item
	onChange      func(int, string) // Callback when the selected index changes
	onSelect      func(int, string) // Callback when an item is selected (e.g., Enter pressed)
}

// NewList creates a new List widget.
func NewList() *List {
	l := &List{
		items:         []string{},
		selectedIndex: -1,
		topIndex:      0,
		style:         tinytui.DefaultListStyle(),
		selectedStyle: tinytui.DefaultListSelectedStyle(),
	}
	l.SetVisible(true) // Explicitly set visibility
	return l
}

// SetItems replaces the current list items with a new slice of strings.
// It resets the selection and scroll position.
func (l *List) SetItems(items []string) *List {
	l.mu.Lock()
	l.items = items
	l.topIndex = 0
	if len(items) > 0 {
		l.selectedIndex = 0 // Select the first item by default
	} else {
		l.selectedIndex = -1 // No selection if empty
	}
	l.clampIndices() // Ensure indices are valid
	l.mu.Unlock()

	// Trigger initial onChange if selection is valid
	l.triggerOnChange()

	if app := l.App(); app != nil {
		app.QueueRedraw()
	}
	return l
}

// SetStyle sets the style for non-selected list items.
func (l *List) SetStyle(style tinytui.Style) *List {
	l.mu.Lock()
	l.style = style
	l.mu.Unlock()
	if app := l.App(); app != nil {
		app.QueueRedraw()
	}
	return l
}

// ApplyTheme applies the provided theme to the List widget
func (l *List) ApplyTheme(theme tinytui.Theme) {
	l.SetStyle(theme.ListStyle())
	l.SetSelectedStyle(theme.ListSelectedStyle())
}

// SetSelectedStyle sets the style for the selected list item.
func (l *List) SetSelectedStyle(style tinytui.Style) *List {
	l.mu.Lock()
	l.selectedStyle = style
	l.mu.Unlock()
	if app := l.App(); app != nil {
		app.QueueRedraw()
	}
	return l
}

// SetOnChange sets a callback function that is triggered whenever the
// selected item changes (e.g., by user navigation).
// The callback receives the new index and the item string.
func (l *List) SetOnChange(handler func(index int, item string)) *List {
	l.mu.Lock()
	l.onChange = handler
	l.mu.Unlock()
	return l
}

// SetOnSelect sets a callback function that is triggered when an item
// is actively selected (e.g., by pressing Enter).
// The callback receives the selected index and the item string.
func (l *List) SetOnSelect(handler func(index int, item string)) *List {
	l.mu.Lock()
	l.onSelect = handler
	l.mu.Unlock()
	return l
}

// SelectedIndex returns the index of the currently selected item.
// Returns -1 if the list is empty or no item is selected.
func (l *List) SelectedIndex() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.selectedIndex
}

// SelectedItem returns the string of the currently selected item.
// Returns an empty string if the list is empty or no item is selected.
func (l *List) SelectedItem() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.selectedIndex >= 0 && l.selectedIndex < len(l.items) {
		return l.items[l.selectedIndex]
	}
	return ""
}

// SetSelectedIndex programmatically sets the selected item index.
// It clamps the index to valid bounds and adjusts the scroll position.
func (l *List) SetSelectedIndex(index int) *List {
	l.mu.Lock()
	oldIndex := l.selectedIndex
	l.selectedIndex = index
	l.clampIndices() // Ensure new index is valid
	changed := l.selectedIndex != oldIndex
	l.mu.Unlock()

	if changed {
		l.triggerOnChange() // Trigger callback if index actually changed
		if app := l.App(); app != nil {
			app.QueueRedraw()
		}
	}
	return l
}

// clampIndices ensures selectedIndex and topIndex are within valid ranges.
// Must be called with l.mu held.
func (l *List) clampIndices() {
	itemCount := len(l.items)
	if itemCount == 0 {
		l.selectedIndex = -1
		l.topIndex = 0
		return
	}

	// Clamp selectedIndex
	if l.selectedIndex < 0 {
		l.selectedIndex = 0
	}
	if l.selectedIndex >= itemCount {
		l.selectedIndex = itemCount - 1
	}

	// Adjust scroll position (topIndex) based on selection and height
	_, _, _, height := l.GetRect()
	if height <= 0 {
		height = 1 // Avoid division by zero or invalid scroll logic
	}

	if l.selectedIndex < l.topIndex {
		// Selection moved above the visible area
		l.topIndex = l.selectedIndex
	} else if l.selectedIndex >= l.topIndex+height {
		// Selection moved below the visible area
		l.topIndex = l.selectedIndex - height + 1
	}

	// Clamp topIndex itself
	if l.topIndex < 0 {
		l.topIndex = 0
	}
	// Ensure topIndex doesn't scroll past the last possible full page
	maxTopIndex := itemCount - height
	if maxTopIndex < 0 {
		maxTopIndex = 0 // Handle case where items fit entirely
	}
	if l.topIndex > maxTopIndex {
		l.topIndex = maxTopIndex
	}
}

// triggerOnChange safely calls the onChange callback.
func (l *List) triggerOnChange() {
	l.mu.RLock()
	handler := l.onChange
	idx := l.selectedIndex
	item := ""
	if idx >= 0 && idx < len(l.items) {
		item = l.items[idx]
	}
	l.mu.RUnlock()

	if handler != nil && idx != -1 {
		handler(idx, item)
	}
}

// triggerOnSelect safely calls the onSelect callback.
func (l *List) triggerOnSelect() {
	l.mu.RLock()
	handler := l.onSelect
	idx := l.selectedIndex
	item := ""
	if idx >= 0 && idx < len(l.items) {
		item = l.items[idx]
	}
	l.mu.RUnlock()

	if handler != nil && idx != -1 {
		handler(idx, item)
	}
}

// Draw draws the list items within the widget's bounds.
func (l *List) Draw(screen tcell.Screen) {
	l.BaseWidget.Draw(screen)

	x, y, width, height := l.GetRect()
	if width <= 0 || height <= 0 {
		return
	}

	l.mu.RLock() // Read lock for accessing items and indices
	itemsToDraw := l.items
	selIdx := l.selectedIndex
	topIdx := l.topIndex
	baseStyle := l.style
	selectedStyle := l.selectedStyle
	l.mu.RUnlock() // Release lock after getting needed data

	// Fill the entire background with the base style
	// This ensures consistent appearance even with partial item lists
	tinytui.Fill(screen, x, y, width, height, ' ', baseStyle)

	// Add some padding for better readability
	padding := 1
	effectiveWidth := width - (padding * 2)
	if effectiveWidth < 1 {
		effectiveWidth = 1
	}

	// Draw visible items
	for i := 0; i < height; i++ {
		itemIndex := topIdx + i
		drawY := y + i

		if itemIndex >= 0 && itemIndex < len(itemsToDraw) {
			item := itemsToDraw[itemIndex]
			style := baseStyle
			if itemIndex == selIdx {
				style = selectedStyle
			}

			// Clear the line with the style's background
			tinytui.Fill(screen, x, drawY, width, 1, ' ', style)

			// Item indicator for selected items (shows focus clearly)
			if itemIndex == selIdx {
				// Draw a focus indicator
				screen.SetContent(x, drawY, '>', nil, style.ToTcell())
				padding = 2 // More padding when showing indicator
			}

			// Truncate text if needed and draw with padding
			displayText := item
			if runewidth.StringWidth(item) > effectiveWidth {
				displayText = runewidth.Truncate(item, effectiveWidth, "")
			}

			// Draw the item text with padding
			tinytui.DrawText(screen, x+padding, drawY, style, displayText)
		}
	}
}

// SetRect updates the widget's dimensions and potentially adjusts scroll.
func (l *List) SetRect(x, y, width, height int) {
	l.mu.Lock()
	l.BaseWidget.SetRect(x, y, width, height) // Call embedded method
	l.clampIndices()                          // Re-clamp indices based on new height
	l.mu.Unlock()
	// No redraw queued here, SetRect is usually called during a redraw cycle
}

// Focusable indicates that Lists can receive focus.
func (l *List) Focusable() bool {
	if !l.IsVisible() {
		return false
	}
	return true
}

// HandleEvent handles keyboard navigation and selection within the list.
func (l *List) HandleEvent(event tcell.Event) bool {
	// Check base widget bindings first
	if l.BaseWidget.HandleEvent(event) {
		return true
	}

	if !l.IsFocused() {
		return false // Only handle keys when focused
	}

	keyEvent, ok := event.(*tcell.EventKey)
	if !ok {
		return false // Not a key event
	}

	l.mu.Lock() // Lock for modifying indices
	currentIndex := l.selectedIndex
	itemCount := len(l.items)
	_, _, _, height := l.GetRect()
	if height <= 0 {
		height = 1
	}
	needsRedraw := false
	indexChanged := false

	if itemCount == 0 {
		l.mu.Unlock()
		return false // Nothing to do in an empty list
	}

	newIndex := currentIndex

	switch keyEvent.Key() {
	case tcell.KeyUp:
		newIndex--
		needsRedraw = true
	case tcell.KeyDown:
		newIndex++
		needsRedraw = true
	case tcell.KeyHome:
		newIndex = 0
		needsRedraw = true
	case tcell.KeyEnd:
		newIndex = itemCount - 1
		needsRedraw = true
	case tcell.KeyPgUp:
		newIndex -= height
		if newIndex < 0 {
			newIndex = 0
		}
		needsRedraw = true
	case tcell.KeyPgDn:
		newIndex += height
		if newIndex >= itemCount {
			newIndex = itemCount - 1
		}
		needsRedraw = true
	case tcell.KeyEnter:
		l.mu.Unlock()       // Unlock before calling callback
		l.triggerOnSelect() // Trigger select action
		return true         // Event handled

	default:
		l.mu.Unlock()
		return false // Key not handled by list navigation
	}

	// Apply changes if needed
	if needsRedraw {
		if newIndex != currentIndex {
			l.selectedIndex = newIndex
			l.clampIndices()                                 // Clamp and adjust scroll
			indexChanged = (l.selectedIndex != currentIndex) // Check if index actually changed after clamping
		}
		l.mu.Unlock() // Unlock before potentially calling callbacks or queuing redraw

		if indexChanged {
			l.triggerOnChange() // Trigger change callback
		}
		if app := l.App(); app != nil {
			app.QueueRedraw()
		}
		return true // Event handled
	}

	l.mu.Unlock()
	return false // Event not handled
}