// types.go
package tinytui

// Rect defines a rectangular area on the screen.
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

// Size defines how components should be sized in layouts.
// Either FixedSize or Proportion should be set, not both.
type Size struct {
	FixedSize  int // Fixed size in cells
	Proportion int // Relative proportion of available space (if FixedSize == 0)
}

// State represents the current state of a component.
type State int

const (
	// StateNormal is the default state
	StateNormal State = iota
	// StateSelected indicates the component is selected but not interacted with
	StateSelected
	// StateInteracted indicates the component is actively interacted with
	StateInteracted
)

// Orientation defines the direction children are laid out in a layout.
type Orientation int

const (
	// Horizontal lays out children side-by-side.
	Horizontal Orientation = iota
	// Vertical lays out children one above the other.
	Vertical
)

// Alignment defines how items are aligned along a layout axis.
type Alignment int

const (
	// AlignStart aligns items at the start (top/left)
	AlignStart Alignment = iota
	// AlignCenter centers items within the available space
	AlignCenter
	// AlignEnd aligns items at the end (bottom/right)
	AlignEnd
	// AlignStretch stretches items to fill the container (default)
	AlignStretch
)

// Border defines the style of border to draw.
type Border int

const (
	// BorderNone indicates no border should be drawn
	BorderNone Border = iota
	// BorderSingle indicates a single-line border
	BorderSingle
	// BorderDouble indicates a double-line border
	BorderDouble
	// BorderSolid indicates a solid block border
	BorderSolid
)

// ScreenMode defines how the terminal screen is handled.
type ScreenMode int

const (
	// ScreenNormal uses the terminal's current size
	ScreenNormal ScreenMode = iota
	// ScreenFullscreen attempts to use the entire terminal
	ScreenFullscreen
	// ScreenAlternate uses the terminal's alternate screen buffer
	ScreenAlternate
)

// SelectionMode defines whether a Grid allows single or multiple selections
type SelectionMode int

const (
	// SingleSelect allows only one item to be selected/interacted at a time
	SingleSelect SelectionMode = iota
	// MultiSelect allows multiple items to be selected/interacted
	MultiSelect
)
