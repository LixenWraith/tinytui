// types.go
package tinytui

// Rect defines a rectangular area on the screen using top-left coordinates (X, Y)
// and dimensions (Width, Height). Standard struct for component geometry.
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

// Size defines constraints for how a component should be sized within a Layout.
// Use either FixedSize (absolute cell count) or Proportion (relative share of remaining space).
// If both are zero or negative, Layout typically assumes Proportion=1.
type Size struct {
	FixedSize  int // Fixed size in cells (takes precedence over Proportion). Set to > 0 to use.
	Proportion int // Relative proportion of available space (used if FixedSize <= 0). Set to > 0 to use.
}

// State represents the interaction state of a component, primarily used for visual feedback
// in interactive elements like Grid cells or potentially Buttons/Checkboxes in the future.
type State int

const (
	// StateNormal is the default, non-selected, non-interacted state.
	StateNormal State = iota
	// StateSelected indicates the component/cell is currently selected (e.g., highlighted by a cursor).
	StateSelected
	// StateInteracted indicates the component/cell has been activated or toggled (e.g., Enter pressed on it).
	StateInteracted
)

// Orientation specifies the direction children are arranged within a Layout.
type Orientation int

const (
	// Horizontal arranges child panes side-by-side, left-to-right.
	Horizontal Orientation = iota
	// Vertical arranges child panes one above the other, top-to-bottom.
	Vertical
)

// Alignment defines how items are positioned within a container or along a layout axis.
// Used primarily for Layout's CrossAxisAlignment, potentially MainAxisAlignment in future.
type Alignment int

const (
	// AlignStart aligns items to the beginning of the axis (Top for Vertical, Left for Horizontal).
	AlignStart Alignment = iota
	// AlignCenter centers items within the available space on the axis. (Layout support may be partial).
	AlignCenter
	// AlignEnd aligns items to the end of the axis (Bottom for Vertical, Right for Horizontal). (Layout support may be partial).
	AlignEnd
	// AlignStretch expands items to fill the available space on the relevant axis (default for Layout's cross axis).
	AlignStretch
)

// Border defines the visual style of a Pane's border line/characters.
type Border int

const (
	// BorderNone indicates no border should be drawn around the pane. Content fills the entire pane rectangle.
	BorderNone Border = iota
	// BorderSingle draws a border using single-line box drawing characters ('┌', '─', '┐', etc.).
	BorderSingle
	// BorderDouble draws a border using double-line box drawing characters ('╔', '═', '╗', etc.).
	BorderDouble
	// BorderSolid draws a border using solid block characters ('▀', '█', '▄', etc.).
	BorderSolid
)

// ScreenMode controls how the application interacts with the terminal screen buffer upon start.
type ScreenMode int

const (
	// ScreenNormal operates within the terminal's main buffer, using its current size and content.
	ScreenNormal ScreenMode = iota
	// ScreenFullscreen attempts to clear and use the entire terminal window (best effort, depends on terminal).
	ScreenFullscreen
	// ScreenAlternate switches to the terminal's alternate screen buffer (if available). This typically
	// provides a clean slate and restores the previous buffer content when the application exits via Fini().
	ScreenAlternate
)

// SelectionMode defines how selection and interaction behave within a Grid component.
type SelectionMode int

const (
	// SingleSelect allows only one cell to be in the 'interacted' state at a time.
	// Interacting with a cell (e.g., pressing Enter) sets it as interacted and clears any previously interacted cell.
	SingleSelect SelectionMode = iota
	// MultiSelect allows multiple cells to be independently toggled into/out of the 'interacted' state.
	MultiSelect
)