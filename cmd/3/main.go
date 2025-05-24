package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/LixenWraith/tinytui" // Assuming tinytui types and NewWrapperPane are in this package
	"github.com/gdamore/tcell/v2"    // For KeyEscape
)

const (
	numLines              = 5
	tokensPerLine         = 8
	tokenLengthMin        = 1
	tokenLengthMax        = 8
	differencePercentage  = 0.20
	gameUpdateInterval    = time.Second
	buttonResetWidth      = 9  // " Reset "
	buttonRegenerateWidth = 12 // "Regenerate"
	buttonExitWidth       = 8  // " Exit "
	// Define fixed height for the bottom section to ensure enough space
	bottomSectionHeight = 9 // 1 (title) + 1 (status) + 1 (timer) + 3 (editor) + 1 (buttons) + 2 (borders/gaps) ~ approx
)

var (
	appInstance *tinytui.Application

	grid1            *tinytui.Grid
	grid2            *tinytui.Grid
	statusText       *tinytui.Text
	timerText        *tinytui.Text
	tokenEditInput   *tinytui.TextInput
	resetButton      *tinytui.Grid
	regenerateButton *tinytui.Grid
	exitButton       *tinytui.Grid
	controlsHostPane *tinytui.Pane

	text1Data [][]string
	text2Data [][]string

	editingToken            bool = false
	editingGrid             *tinytui.Grid
	editingLineIndex        int
	editingTokenIndex       int
	originalFocusedGridCell tinytui.Component

	startTime   time.Time
	differences int
	gameOver    bool = false
	randSource  *rand.Rand
	stateMutex  sync.RWMutex

	colorSame         tinytui.Style
	colorPartialDiff  tinytui.Style
	colorFullDiffWord tinytui.Style
	colorFullDiffNum  tinytui.Style
	colorSymbolDiff   tinytui.Style
)

func generateRandomToken() string {
	length := randSource.Intn(tokenLengthMax-tokenLengthMin+1) + tokenLengthMin
	var builder strings.Builder
	tokenType := randSource.Intn(100)
	if tokenType < 60 {
		for i := 0; i < length; i++ {
			builder.WriteByte(byte(randSource.Intn(26) + 'a'))
		}
	} else if tokenType < 90 {
		for i := 0; i < length; i++ {
			builder.WriteByte(byte(randSource.Intn(10) + '0'))
		}
	} else {
		symbols := "!@#$%^&*()-=_+[]{};':\",./<>?"
		builder.WriteByte(symbols[randSource.Intn(len(symbols))])
	}
	return builder.String()
}

func generateTextData() [][]string {
	data := make([][]string, numLines)
	for i := 0; i < numLines; i++ {
		lineTokens := randSource.Intn(tokensPerLine/2+1) + tokensPerLine/2
		data[i] = make([]string, lineTokens)
		for j := 0; j < lineTokens; j++ {
			data[i][j] = generateRandomToken()
		}
	}
	return data
}

func introduceDifferences(original [][]string) [][]string {
	mutated := make([][]string, len(original))
	for i, line := range original {
		mutated[i] = make([]string, len(line))
		for j, token := range line {
			if randSource.Float64() < differencePercentage {
				if randSource.Float64() < 0.5 && len(token) > 1 {
					runes := []rune(token)
					idxToChange := randSource.Intn(len(runes))
					changeType := randSource.Intn(3)
					if changeType == 0 && runes[idxToChange] >= 'a' && runes[idxToChange] <= 'z' {
						runes[idxToChange] = rune(randSource.Intn(26) + 'a')
					} else if changeType == 1 && runes[idxToChange] >= '0' && runes[idxToChange] <= '9' {
						runes[idxToChange] = rune(randSource.Intn(10) + '0')
					} else {
						runes[idxToChange] = []rune(generateRandomToken())[0]
					}
					mutated[i][j] = string(runes)
				} else {
					mutated[i][j] = generateRandomToken()
				}
			} else {
				mutated[i][j] = token
			}
		}
	}
	return mutated
}

func populateGridFromData(grid *tinytui.Grid, data [][]string, targetData [][]string) {
	stateMutex.RLock()
	defer stateMutex.RUnlock()
	if grid == nil || data == nil {
		return
	}
	gridCellData := make([][]string, len(data))
	maxCols := 0
	for i, line := range data {
		gridCellData[i] = make([]string, len(line))
		if len(line) > maxCols {
			maxCols = len(line)
		}
		for j, token := range line {
			gridCellData[i][j] = token
		}
	}
	for i := range gridCellData {
		if len(gridCellData[i]) < maxCols {
			paddedLine := make([]string, maxCols)
			copy(paddedLine, gridCellData[i])
			for j := len(gridCellData[i]); j < maxCols; j++ {
				paddedLine[j] = ""
			}
			gridCellData[i] = paddedLine
		}
	}
	if len(gridCellData) == 0 {
		grid.SetCells([][]string{{""}})
		return
	}
	grid.SetCells(gridCellData)
	applyHighlightingLogic(data, targetData)
	appInstance.QueueRedraw()
}

func isSymbol(token string) bool {
	if len(token) != 1 {
		return false
	}
	r := rune(token[0])
	return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'))
}
func isNumeric(token string) bool {
	if token == "" {
		return false
	}
	for _, r := range token {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func applyHighlightingLogic(data [][]string, targetData [][]string) {
	currentDiffs := 0
	// PLACEHOLDER_GRID_STYLING: The 'styleForLogic' variable is for difference counting.
	// tinytui.Grid doesn't support per-cell text styling for dynamic coloring as requested.
	for r, dataLine := range data {
		for c, dataToken := range dataLine {
			styleForLogic := colorSame // Not visually applied to grid cells

			if r < len(targetData) && c < len(targetData[r]) {
				targetToken := targetData[r][c]
				if dataToken != targetToken {
					currentDiffs++
					if isSymbol(dataToken) || isSymbol(targetToken) {
						styleForLogic = colorSymbolDiff
					} else if isNumeric(dataToken) || isNumeric(targetToken) {
						styleForLogic = colorFullDiffNum
					} else { // Word-like
						commonPrefix := 0
						for k := 0; k < len(dataToken) && k < len(targetToken); k++ {
							if dataToken[k] == targetToken[k] {
								commonPrefix++
							} else {
								break
							}
						}
						maxLen := len(dataToken)
						if len(targetToken) > maxLen {
							maxLen = len(targetToken)
						}
						if commonPrefix > 0 && commonPrefix < maxLen && maxLen > 2 {
							styleForLogic = colorPartialDiff
						} else {
							styleForLogic = colorFullDiffWord
						}
					}
				}
			} else { // Token exists in data but not in target
				currentDiffs++
				if isSymbol(dataToken) {
					styleForLogic = colorSymbolDiff
				} else {
					styleForLogic = colorFullDiffWord
				}
			}
			_ = styleForLogic // Use the styleForLogic to avoid unused error until Grid can use it
		}
	}

	// Also count tokens in targetData that are not in data
	for r, targetLine := range targetData {
		for c := range targetLine {
			if r >= len(data) || c >= len(data[r]) {
				currentDiffs++ // This token is in target but not in source
			}
		}
	}
	differences = currentDiffs // Update global differences count
}

func updateGameStatus() {
	stateMutex.RLock()
	defer stateMutex.RUnlock()
	if gameOver {
		return
	}
	elapsed := time.Since(startTime).Round(time.Second)
	if statusText != nil {
		statusText.SetContent(fmt.Sprintf("Differences: %d", differences))
	}
	if timerText != nil {
		timerText.SetContent(fmt.Sprintf("Time: %s", elapsed))
	}
}

func checkWinCondition() {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	if gameOver {
		return
	}
	if differences == 0 {
		gameOver = true
		appInstance.Dispatch(&tinytui.SimpleCommand{Func: func(app *tinytui.Application) {
			winMessageText := tinytui.NewText("Congratulations! Texts Matched!")
			winMessageText.SetAlignment(tinytui.AlignTextCenter)
			winMessageText.SetStyle(app.GetTheme().TextStyle().Foreground(tinytui.ColorLime).Bold(true))
			if controlsHostPane != nil {
				controlsHostPane.SetChild(winMessageText)
			}
			if statusText != nil {
				statusText.SetContent(fmt.Sprintf("YOU WON! Matched in %s.", time.Since(startTime).Round(time.Second)))
			}
			if timerText != nil {
				timerText.SetContent("--- Game Over ---")
			}
			if editingToken {
				tokenEditInput.SetText("")
				tokenEditInput.SetVisible(false)
				editingToken = false
				if originalFocusedGridCell != nil {
					app.Dispatch(&tinytui.FocusCommand{Target: originalFocusedGridCell})
				}
			}
			app.QueueRedraw()
		}})
	}
}

func buildButtonsLayout() *tinytui.Layout {
	buttonsLayout := tinytui.NewLayout(tinytui.Horizontal)
	buttonsLayout.SetGap(1)
	buttonsLayout.SetMainAxisAlignment(tinytui.AlignCenter)
	paneR := tinytui.NewWrapperPane(resetButton)
	paneG := tinytui.NewWrapperPane(regenerateButton)
	paneE := tinytui.NewWrapperPane(exitButton)
	buttonsLayout.AddPane(paneR, tinytui.Size{FixedSize: buttonResetWidth})
	buttonsLayout.AddPane(paneG, tinytui.Size{FixedSize: buttonRegenerateWidth})
	buttonsLayout.AddPane(paneE, tinytui.Size{FixedSize: buttonExitWidth})
	return buttonsLayout
}

func resetControlsToButtons() {
	if controlsHostPane != nil {
		buttonsL := buildButtonsLayout()
		paneHoldingButtonsLayout := tinytui.NewPane()
		paneHoldingButtonsLayout.SetBorder(tinytui.BorderNone, tinytui.DefaultStyle)
		paneHoldingButtonsLayout.SetChild(buttonsL)
		controlsContainerLayout := tinytui.NewLayout(tinytui.Vertical)
		controlsContainerLayout.AddPane(paneHoldingButtonsLayout, tinytui.Size{FixedSize: 1})
		controlsHostPane.SetChild(controlsContainerLayout)
		appInstance.QueueRedraw()
	}
}

func resetGame() {
	stateMutex.Lock()
	startTime = time.Now()
	gameOver = false
	text1Data = introduceDifferences(text2Data)
	stateMutex.Unlock()
	appInstance.Dispatch(&tinytui.SimpleCommand{Func: func(app *tinytui.Application) { resetControlsToButtons() }})
	populateGridFromData(grid1, text1Data, text2Data)
	checkWinCondition()
	updateGameStatus()
	appInstance.Dispatch(&tinytui.FocusCommand{Target: grid1})
}

func regenerateTexts() {
	stateMutex.Lock()
	startTime = time.Now()
	gameOver = false
	text2Data = generateTextData()
	text1Data = introduceDifferences(text2Data)
	stateMutex.Unlock()
	appInstance.Dispatch(&tinytui.SimpleCommand{Func: func(app *tinytui.Application) { resetControlsToButtons() }})
	populateGridFromData(grid1, text1Data, text2Data)
	populateGridFromData(grid2, text2Data, text1Data)
	checkWinCondition()
	updateGameStatus()
	appInstance.Dispatch(&tinytui.FocusCommand{Target: grid1})
}

func startTokenEdit(grid *tinytui.Grid, r, c int) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	if gameOver || r < 0 || r >= len(text1Data) || c < 0 || c >= len(text1Data[r]) {
		return
	}
	editingToken = true
	editingGrid = grid
	editingLineIndex = r
	editingTokenIndex = c
	originalFocusedGridCell = appInstance.GetFocusedComponent()
	tokenEditInput.SetText(text1Data[r][c])
	tokenEditInput.SetVisible(true)
	appInstance.Dispatch(&tinytui.FocusCommand{Target: tokenEditInput})
	if statusText != nil {
		statusText.SetContent("Editing token...")
	}
}

func finishTokenEdit(submit bool) {
	stateMutex.Lock()
	if !editingToken {
		stateMutex.Unlock()
		return
	}
	newText := tokenEditInput.GetText()
	r, tokIdx := editingLineIndex, editingTokenIndex
	if submit {
		if r >= 0 && r < len(text1Data) && tokIdx >= 0 && tokIdx < len(text1Data[r]) {
			text1Data[r][tokIdx] = newText
		}
	}
	stateMutex.Unlock()
	tokenEditInput.SetText("")
	tokenEditInput.SetVisible(false)
	editingToken = false
	populateGridFromData(grid1, text1Data, text2Data)
	checkWinCondition()
	if !gameOver {
		updateGameStatus()
	}
	if originalFocusedGridCell != nil {
		appInstance.Dispatch(&tinytui.FocusCommand{Target: originalFocusedGridCell})
	} else {
		appInstance.Dispatch(&tinytui.FocusCommand{Target: editingGrid})
	}
	originalFocusedGridCell = nil
}

func main() {
	logFile, _ := os.OpenFile("text_diff_game.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	defer logFile.Close()
	log.SetOutput(logFile)
	log.Println("Starting Text Diff Game")

	randSource = rand.New(rand.NewSource(time.Now().UnixNano()))
	app := tinytui.NewApplication()
	appInstance = app
	if !tinytui.SetTheme(tinytui.ThemeTokyoSweet) {
		fmt.Printf("Warning: Theme '%s' not found.\n", tinytui.ThemeTokyoSweet)
		time.Sleep(2 * time.Second)
	}
	app.SetTheme(tinytui.GetTheme())
	theme := app.GetTheme()
	colorSame = theme.TextStyle()
	colorPartialDiff = theme.TextStyle().Foreground(tinytui.Color(0xff9e64)) // sweetOrange
	colorFullDiffWord = theme.TextStyle().Foreground(tinytui.Color(0xf7768e))
	colorFullDiffNum = theme.TextStyle().Foreground(tinytui.Color(0xf7768e)) // sweetRed
	colorSymbolDiff = theme.TextStyle().Foreground(tinytui.Color(0x9ece6a))  // sweetGreen

	text2Data = generateTextData()
	text1Data = introduceDifferences(text2Data)
	startTime = time.Now()

	grid1 = tinytui.NewGrid()
	grid1.SetSelectionMode(tinytui.SingleSelect)
	grid1.SetOnSelect(func(r, c int, i string) {
		if !editingToken {
			startTokenEdit(grid1, r, c)
		}
	})
	grid2 = tinytui.NewGrid()
	grid2.SetSelectionMode(tinytui.SingleSelect) // Read-only reference

	statusText = tinytui.NewText("")
	statusText.SetAlignment(tinytui.AlignTextCenter)
	timerText = tinytui.NewText("")
	timerText.SetAlignment(tinytui.AlignTextCenter)

	tokenEditInput = tinytui.NewTextInput()
	tokenEditInput.SetVisible(false)
	tokenEditInput.SetOnSubmit(func(text string) { finishTokenEdit(true) })
	app.RegisterKeyHandler(tcell.KeyEscape, 0, func() bool {
		if editingToken && app.GetFocusedComponent() == tokenEditInput {
			finishTokenEdit(false)
			return true
		}
		return false
	})

	resetButton = tinytui.NewGrid()
	resetButton.SetCells([][]string{{" Reset "}})
	resetButton.SetCellSize(buttonResetWidth, 1)
	resetButton.SetOnSelect(func(r, c int, i string) {
		if !gameOver {
			resetGame()
			resetButton.SetCellInteracted(r, c, false)
		}
	})
	regenerateButton = tinytui.NewGrid()
	regenerateButton.SetCells([][]string{{"Regenerate"}})
	regenerateButton.SetCellSize(buttonRegenerateWidth, 1)
	regenerateButton.SetOnSelect(func(r, c int, i string) {
		if !gameOver {
			regenerateTexts()
			regenerateButton.SetCellInteracted(r, c, false)
		}
	})
	exitButton = tinytui.NewGrid()
	exitButton.SetCells([][]string{{" Exit "}})
	exitButton.SetCellSize(buttonExitWidth, 1)
	exitButton.SetOnSelect(func(r, c int, i string) { app.Stop() })

	populateGridFromData(grid1, text1Data, text2Data)
	populateGridFromData(grid2, text2Data, text1Data)

	pane1 := tinytui.NewPane()
	pane1.SetTitle("Your Text (Select cell & Enter to Edit)")
	pane1.SetChild(grid1)
	pane2 := tinytui.NewPane()
	pane2.SetTitle("Target Text (Reference)")
	pane2.SetChild(grid2)

	topLayout := tinytui.NewLayout(tinytui.Horizontal) // This is a *tinytui.Layout
	topLayout.SetGap(1)
	topLayout.AddPane(pane1, tinytui.Size{Proportion: 1})
	topLayout.AddPane(pane2, tinytui.Size{Proportion: 1})

	topWrapperPane := tinytui.NewPane()
	topWrapperPane.SetBorder(tinytui.BorderNone, tinytui.DefaultStyle)
	topWrapperPane.SetChild(topLayout)

	pStatus := tinytui.NewWrapperPane(statusText)
	pTimer := tinytui.NewWrapperPane(timerText)

	tokenEditPane := tinytui.NewPane()
	tokenEditPane.SetBorder(tinytui.BorderSingle, theme.PaneFocusBorderStyle())
	tokenEditPane.SetTitle("Edit Token (Enter: Save, Esc: Cancel)")
	tokenEditPane.SetChild(tokenEditInput)

	controlsHostPane = tinytui.NewPane()
	controlsHostPane.SetBorder(tinytui.BorderNone, tinytui.DefaultStyle)
	resetControlsToButtons() // Initialize controlsHostPane with buttons

	bottomContentLayout := tinytui.NewLayout(tinytui.Vertical)
	bottomContentLayout.SetGap(0) // No gap for tighter packing
	bottomContentLayout.AddPane(pStatus, tinytui.Size{FixedSize: 1})
	bottomContentLayout.AddPane(pTimer, tinytui.Size{FixedSize: 1})
	bottomContentLayout.AddPane(tokenEditPane, tinytui.Size{FixedSize: 3})    // Height for input, title, borders
	bottomContentLayout.AddPane(controlsHostPane, tinytui.Size{FixedSize: 1}) // Height for buttons/win message

	bottomPane := tinytui.NewPane()
	bottomPane.SetTitle("Status & Controls")
	bottomPane.SetChild(bottomContentLayout)

	splitLayout := tinytui.NewLayout(tinytui.Vertical)
	splitLayout.SetGap(1)
	splitLayout.AddPane(topWrapperPane, tinytui.Size{Proportion: 1})              // Top takes remaining space
	splitLayout.AddPane(bottomPane, tinytui.Size{FixedSize: bottomSectionHeight}) // Bottom has fixed height

	app.SetLayout(splitLayout)

	// --- Game Loop in Goroutine ---
	go func() {
		ticker := time.NewTicker(gameUpdateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if appInstance == nil {
					return // App instance gone, exit goroutine
				}

				// Non-blocking check if the application is stopping
				select {
				case <-appInstance.StopChan():
					log.Println("Game loop (ticker) detected app stop.")
					return
				default:
					// Continue if not stopping
				}

				stateMutex.RLock()
				isGameOver := gameOver
				isEditing := editingToken
				stateMutex.RUnlock()

				if !isGameOver && !isEditing {
					updateGameStatus()
					appInstance.QueueRedraw() // For timer updates
				}
			case <-appInstance.StopChan():
				log.Println("Game loop detected app stop directly.")
				return
			}
		}
	}()

	checkWinCondition()
	updateGameStatus()
	app.Dispatch(&tinytui.FocusCommand{Target: grid1})
	log.Println("Running application...")
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	log.Println("Application exited.")
}