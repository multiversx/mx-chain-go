package termuic

import (
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"strconv"
)

type termuiConsoleGrid struct {
	grid              *ui.Grid
	pNonce            *widgets.Paragraph
	pSynchronizeRound *widgets.Paragraph
	pCurrentRound     *widgets.Paragraph
	pIsSyncing        *widgets.Paragraph
	lLog              *widgets.List
}

func NewtermuiConsoleGrid() *termuiConsoleGrid {

	self := termuiConsoleGrid{}

	self.initWidgets()

	self.setupGrid()

	return &self
}

func (tcg *termuiConsoleGrid) initWidgets() {
	tcg.pNonce = widgets.NewParagraph()
	tcg.pSynchronizeRound = widgets.NewParagraph()
	tcg.pCurrentRound = widgets.NewParagraph()
	tcg.pIsSyncing = widgets.NewParagraph()

	tcg.lLog = widgets.NewList()
}

func (tcg *termuiConsoleGrid) setupGrid() {
	tcg.grid = ui.NewGrid()

	tcg.grid.Set(
		ui.NewRow(1.0/2,
			ui.NewCol(1.0/4, tcg.pNonce),
			ui.NewCol(1.0/4, tcg.pCurrentRound),
			ui.NewCol(1.0/4, tcg.pSynchronizeRound),
			ui.NewCol(1.0/4, tcg.pIsSyncing),
		),
		ui.NewRow(1.0/2, tcg.lLog),
	)
}

//Grid returns all console widgets that needs to be displayed
func (tcg *termuiConsoleGrid) Grid() *ui.Grid {
	return tcg.grid
}

func (tcg *termuiConsoleGrid) PrepareNonceForDisplay(Nonce int) {
	tcg.pNonce.Text = "Nonce : " + strconv.Itoa(Nonce)

	if Nonce < 50 {
		tcg.pNonce.TextStyle = ui.Style{
			Fg:       ui.ColorBlack,
			Bg:       ui.ColorRed,
			Modifier: 1,
		}
	} else {
		tcg.pNonce.TextStyle = ui.Style{
			Fg:       ui.ColorBlack,
			Bg:       ui.ColorRed,
			Modifier: 10,
		}
	}

	return
}

func (tcg *termuiConsoleGrid) PrepareSynchronizedRoundForDisplay(synchronizedRound int) {
	tcg.pSynchronizeRound.Text = "SynchronizedRound : " + strconv.Itoa(synchronizedRound)
	return
}

func (tcg *termuiConsoleGrid) PrepareIsSyncingForDisplay(isSyncing int) {
	tcg.pIsSyncing.Text = "IsSyncing : " + strconv.Itoa(isSyncing)

	tcg.pIsSyncing.TextStyle = ui.Style{
		Fg:       ui.ColorBlack,
		Bg:       ui.ColorGreen,
		Modifier: 1,
	}

	if isSyncing == 1 {
		tcg.pIsSyncing.TextStyle.Bg = ui.ColorRed
	}

	return
}

func (tcg *termuiConsoleGrid) PrepareCurrentRoundForDisplay(currentRound int) {
	tcg.pCurrentRound.Text = "CurrentRound : " + strconv.Itoa(currentRound)
	return
}

func (tcg *termuiConsoleGrid) PrepareListWithLogsForDisplay(logData []string) {

	tcg.lLog.Title = "Log info"
	tcg.lLog.Rows = logData
	tcg.lLog.TextStyle = ui.NewStyle(ui.ColorYellow)
	tcg.lLog.WrapText = false

	return
}
