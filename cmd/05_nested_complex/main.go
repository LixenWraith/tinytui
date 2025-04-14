// cmd/05_nested_complex/main.go
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/LixenWraith/tinytui"
)

var (
	todoTasks      = []string{"Buy milk", "Learn Go TUI", "Fix bugs"}
	completedTasks = []string{}
	taskListGrid   *tinytui.Grid // Global refs for easy access in handlers
	completedText  *tinytui.Text
	statsText      *tinytui.Text
	taskInput      *tinytui.TextInput
	appInstance    *tinytui.Application // Ref to dispatch commands
)

func updateTaskListGrid() {
	// Convert tasks slice to [][]string for the grid
	gridData := make([][]string, len(todoTasks))
	for i, task := range todoTasks {
		gridData[i] = []string{task} // Grid with one column
	}
	// Use Dispatch to update grid safely from potentially different goroutine (event handler)
	if appInstance != nil {
		appInstance.Dispatch(&tinytui.UpdateGridCommand{
			Target:  taskListGrid,
			Content: gridData,
		})
	}
}

func updateCompletedText() {
	content := "Completed Tasks:\n" + strings.Join(completedTasks, "\n")
	if appInstance != nil {
		appInstance.Dispatch(&tinytui.UpdateTextCommand{
			Target:  completedText,
			Content: content,
		})
	}
}

func updateStatsText() {
	content := fmt.Sprintf(" Stats\n-------\nTodo: %d\nDone: %d", len(todoTasks), len(completedTasks))
	if appInstance != nil {
		appInstance.Dispatch(&tinytui.UpdateTextCommand{
			Target:  statsText,
			Content: content,
		})
	}
}

func main() {
	// --- Logging Setup ---
	logFile, err := os.OpenFile("tinytui_todo_test.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.Println("--- Starting Todo App Test ---")

	// --- Application Setup ---
	app := tinytui.NewApplication()
	appInstance = app // Store global ref
	// app.SetTheme(tinytui.ThemeTurbo)
	app.SetShowPaneIndices(true)

	// --- Create Components ---
	// Header/Footer
	headerText := tinytui.NewText(" Tiny TUI Todo App ") // Add spaces for visual padding
	footerText := tinytui.NewText(" [Tab] Navigate | [Enter] Add/Select/Delete | [Esc] Quit ")

	// Left Pane Components
	taskInput = tinytui.NewTextInput()
	taskInput.SetText("") // Start empty

	addButtonGrid := tinytui.NewGrid()
	addButtonGrid.SetCells([][]string{{" Add "}}) // Spaces for padding
	addButtonGrid.SetCellSize(7, 1)
	addButtonGrid.SetSelectionMode(tinytui.SingleSelect)

	clearButtonGrid := tinytui.NewGrid()
	clearButtonGrid.SetCells([][]string{{" Clear "}})
	clearButtonGrid.SetCellSize(9, 1)
	clearButtonGrid.SetSelectionMode(tinytui.SingleSelect)

	// Center Pane Components
	taskListGrid = tinytui.NewGrid()
	taskListGrid.SetSelectionMode(tinytui.MultiSelect) // Allow selecting multiple tasks
	taskListGrid.SetAutoWidth(true)                    // Adjust width to content
	taskListGrid.SetIndicator('*', true)               // Use '*' for selection/interaction marker

	deleteButtonGrid := tinytui.NewGrid()
	deleteButtonGrid.SetCells([][]string{{"Delete Selected"}})
	deleteButtonGrid.SetCellSize(17, 1)
	deleteButtonGrid.SetSelectionMode(tinytui.SingleSelect)

	// Right Pane Components
	statsText = tinytui.NewText("")     // Content updated dynamically
	completedText = tinytui.NewText("") // Content updated dynamically
	completedText.SetWrap(true)

	// --- Create Panes ---
	headerPane := tinytui.NewPane()
	headerPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle())
	headerPane.SetChild(headerText)

	footerPane := tinytui.NewPane()
	footerPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle())
	footerPane.SetChild(footerText)

	// Left Column Panes
	inputFieldPane := tinytui.NewPane()
	inputFieldPane.SetTitle("New Task")
	inputFieldPane.SetChild(taskInput)

	addButtonPane := tinytui.NewPane()
	addButtonPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle()) // Borderless button area
	addButtonPane.SetChild(addButtonGrid)

	clearButtonPane := tinytui.NewPane()
	clearButtonPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle()) // Borderless button area
	clearButtonPane.SetChild(clearButtonGrid)

	// Center Column Panes
	taskListPane := tinytui.NewPane()
	taskListPane.SetTitle("Todo List (Select with Enter)")
	taskListPane.SetChild(taskListGrid)

	deleteButtonPane := tinytui.NewPane()
	deleteButtonPane.SetTitle("Actions")
	deleteButtonPane.SetChild(deleteButtonGrid)

	// Right Column Panes
	statsPane := tinytui.NewPane()
	// statsPane.SetTitle("Stats") // Title embedded in text
	statsPane.SetChild(statsText)

	completedPane := tinytui.NewPane()
	completedPane.SetTitle("Completed")
	completedPane.SetChild(completedText)

	// --- Setup Layouts ---
	// Top level: Vertical (Header, Middle Horizontal, Footer)
	mainLayout := tinytui.NewLayout(tinytui.Vertical)
	mainLayout.SetGap(0) // No gap for header/footer

	// Middle level: Horizontal (Left Column, Center Column, Right Column)
	middleLayout := tinytui.NewLayout(tinytui.Horizontal)
	middleLayout.SetGap(1)

	// Left Column: Vertical (Input Field, Buttons Horizontal)
	leftColLayout := tinytui.NewLayout(tinytui.Vertical)
	leftColLayout.SetGap(1)

	// Buttons Area: Horizontal (Add, Clear)
	buttonsLayout := tinytui.NewLayout(tinytui.Horizontal)
	buttonsLayout.SetGap(1)

	// Center Column: Vertical (Task List, Delete Button)
	centerColLayout := tinytui.NewLayout(tinytui.Vertical)
	centerColLayout.SetGap(1)

	// Right Column: Vertical (Stats, Completed List)
	rightColLayout := tinytui.NewLayout(tinytui.Vertical)
	rightColLayout.SetGap(1)

	// --- Assemble Layouts ---
	// Buttons H -> Buttons Pane (wrapper) -> Left Column V
	buttonsPane := tinytui.NewPane()
	buttonsPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle())
	buttonsLayout.AddPane(addButtonPane, tinytui.Size{FixedSize: 7})   // Width 7
	buttonsLayout.AddPane(clearButtonPane, tinytui.Size{FixedSize: 9}) // Width 9
	buttonsPane.SetChild(buttonsLayout)

	// Left Column V -> Left Pane (wrapper) -> Middle H
	leftPane := tinytui.NewPane()
	leftPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle()) // Wrapper pane borderless
	leftColLayout.AddPane(inputFieldPane, tinytui.Size{FixedSize: 3})        // Height 3
	leftColLayout.AddPane(buttonsPane, tinytui.Size{FixedSize: 1})           // Height 1 (buttons grid is 1 high)
	leftPane.SetChild(leftColLayout)

	// Center Column V -> Center Pane (wrapper) -> Middle H
	centerPane := tinytui.NewPane()
	centerPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle())
	centerColLayout.AddPane(taskListPane, tinytui.Size{Proportion: 1})    // Proportional height
	centerColLayout.AddPane(deleteButtonPane, tinytui.Size{FixedSize: 3}) // Height 3
	centerPane.SetChild(centerColLayout)

	// Right Column V -> Right Pane (wrapper) -> Middle H
	rightPane := tinytui.NewPane()
	rightPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle())
	rightColLayout.AddPane(statsPane, tinytui.Size{FixedSize: 5})      // Fixed height for stats
	rightColLayout.AddPane(completedPane, tinytui.Size{Proportion: 1}) // Proportional height
	rightPane.SetChild(rightColLayout)

	// Middle H -> Middle Wrapper Pane -> Main V
	middleWrapperPane := tinytui.NewPane()
	middleWrapperPane.SetBorder(tinytui.BorderNone, tinytui.DefaultPaneBorderStyle())
	middleLayout.AddPane(leftPane, tinytui.Size{FixedSize: 30})   // Fixed width left
	middleLayout.AddPane(centerPane, tinytui.Size{Proportion: 1}) // Proportional width center
	middleLayout.AddPane(rightPane, tinytui.Size{FixedSize: 30})  // Fixed width right
	middleWrapperPane.SetChild(middleLayout)

	// Main V
	mainLayout.AddPane(headerPane, tinytui.Size{FixedSize: 1})         // Height 1 header
	mainLayout.AddPane(middleWrapperPane, tinytui.Size{Proportion: 1}) // Proportional height middle section
	mainLayout.AddPane(footerPane, tinytui.Size{FixedSize: 1})         // Height 1 footer

	// --- Set Application Layout ---
	app.SetLayout(mainLayout) // Triggers index assignment for top-level panes (Header, MiddleWrapper, Footer)

	// --- Event Handlers ---
	handleAddTask := func() {
		newTask := strings.TrimSpace(taskInput.GetText())
		if newTask != "" {
			log.Printf("Adding task: %s", newTask)
			todoTasks = append(todoTasks, newTask)
			updateTaskListGrid()
			updateStatsText()
			// Clear input via command
			app.Dispatch(&tinytui.UpdateTextCommand{Target: taskInput, Content: ""})
		}
		// Always focus input after trying to add
		app.Dispatch(&tinytui.FocusCommand{Target: taskInput})
	}

	taskInput.SetOnSubmit(func(text string) {
		handleAddTask()
	})

	addButtonGrid.SetOnSelect(func(r, c int, i string) {
		log.Println("Add button selected")
		handleAddTask()
		addButtonGrid.SetCellInteracted(r, c, false) // Deselect button visually
	})

	clearButtonGrid.SetOnSelect(func(r, c int, i string) {
		log.Println("Clear button selected")
		app.Dispatch(&tinytui.UpdateTextCommand{Target: taskInput, Content: ""})
		app.Dispatch(&tinytui.FocusCommand{Target: taskInput}) // Focus input after clear
		clearButtonGrid.SetCellInteracted(r, c, false)         // Deselect button visually
	})

	// Task list selection toggles interaction state (for deletion)
	taskListGrid.SetOnSelect(func(row, col int, item string) {
		log.Printf("Task list item selected/deselected: Row %d (%s)", row, item)
		// State is toggled internally by the grid's toggleCellInteraction
		// No further action needed here, just rely on GetInteractedCells later
	})

	deleteButtonGrid.SetOnSelect(func(r, c int, i string) {
		log.Println("Delete button selected")
		interacted := taskListGrid.GetInteractedCells() // Get [row, col] pairs
		if len(interacted) == 0 {
			log.Println("No tasks selected for deletion.")
			app.Dispatch(&tinytui.FocusCommand{Target: taskListGrid}) // Focus list if nothing to delete
			deleteButtonGrid.SetCellInteracted(r, c, false)           // Deselect button
			return
		}

		newTodoTasks := []string{}
		indicesToDelete := make(map[int]bool)
		for _, cell := range interacted {
			indicesToDelete[cell[0]] = true // Mark row index for deletion
		}

		for idx, task := range todoTasks {
			if indicesToDelete[idx] {
				log.Printf("Completing task: %s", task)
				completedTasks = append(completedTasks, task) // Add to completed
			} else {
				newTodoTasks = append(newTodoTasks, task) // Keep in todo list
			}
		}
		todoTasks = newTodoTasks // Update main task list

		taskListGrid.ClearInteractions() // Clear visual selection marks
		updateTaskListGrid()
		updateCompletedText()
		updateStatsText()

		app.Dispatch(&tinytui.FocusCommand{Target: taskListGrid}) // Focus list after deleting
		deleteButtonGrid.SetCellInteracted(r, c, false)           // Deselect button
	})

	// --- Initial State Update ---
	updateTaskListGrid()
	updateCompletedText()
	updateStatsText()

	// --- Set Initial Focus ---
	app.Dispatch(&tinytui.FocusCommand{Target: taskInput})

	// --- Run Application ---
	log.Println("Running Todo application...")
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		log.Fatalf("Error running application: %v", err)
		os.Exit(1)
	}
	log.Println("Todo application exited normally.")
}