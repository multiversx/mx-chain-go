package termuic

import (
	"fmt"
	"strconv"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type termuiConsoleGrid struct {
	grid       *ui.Grid
	lLog       *widgets.List
	pPublicKey *widgets.Paragraph
	pShardId   *widgets.Paragraph

	pCountLeader         *widgets.Paragraph
	pCountConsensus      *widgets.Paragraph
	pCountAcceptedBlocks *widgets.Paragraph

	gTxPoolLoad        *widgets.Gauge
	pNumConnectedPeers *widgets.Paragraph

	tSyncInfo *widgets.Table
}

//NewtermuiConsoleGrid initialize struct termuiConsoleGrid
func NewtermuiConsoleGrid() *termuiConsoleGrid {

	self := termuiConsoleGrid{}

	self.initWidgets()

	self.setupGrid()

	return &self
}

func (tcg *termuiConsoleGrid) initWidgets() {

	tcg.pPublicKey = widgets.NewParagraph()
	tcg.pShardId = widgets.NewParagraph()

	tcg.pCountLeader = widgets.NewParagraph()
	tcg.pCountConsensus = widgets.NewParagraph()
	tcg.pCountAcceptedBlocks = widgets.NewParagraph()

	tcg.gTxPoolLoad = widgets.NewGauge()
	tcg.pNumConnectedPeers = widgets.NewParagraph()

	tcg.lLog = widgets.NewList()

	tcg.tSyncInfo = widgets.NewTable()
}

func (tcg *termuiConsoleGrid) setupGrid() {
	tcg.grid = ui.NewGrid()

	tcg.grid.Set(
		ui.NewRow(1.0/7, tcg.pPublicKey),
		ui.NewRow(1.0/10, tcg.tSyncInfo),
		ui.NewRow(1.0/3,
			ui.NewCol(1.0/2,
				ui.NewRow(1.0/4, tcg.pShardId),
				ui.NewRow(1.0/4, tcg.pCountConsensus),
				ui.NewRow(1.0/4, tcg.pCountLeader),
				ui.NewRow(1.0/4, tcg.pCountAcceptedBlocks),
			),
			ui.NewCol(1.0/2,
				ui.NewRow(1.0/2, tcg.gTxPoolLoad),
				ui.NewRow(1.0/2, tcg.pNumConnectedPeers),
			),
		),
		//ui.NewRow(1.0/3, tcg.slgNode),
		ui.NewRow(1.0/3, tcg.lLog),
	)
}

func (tcg *termuiConsoleGrid) SetupBigLogGrid() {
	tcg.grid = ui.NewGrid()

	tcg.grid.Set(
		ui.NewRow(1.0, tcg.lLog),
	)
}

//Grid returns all console widgets that needs to be displayed
func (tcg *termuiConsoleGrid) Grid() *ui.Grid {
	return tcg.grid
}

func (tcg *termuiConsoleGrid) PrepareListWithLogsForDisplay(logData []string) {

	tcg.lLog.Title = "Log info"
	tcg.lLog.Rows = logData
	tcg.lLog.TextStyle = ui.NewStyle(ui.ColorYellow)
	tcg.lLog.WrapText = false

	return
}

func (tcg *termuiConsoleGrid) PreparePublicKeyForDisplay(publicKey string) {
	tcg.pPublicKey.Title = fmt.Sprintf("PubicKey")
	tcg.pPublicKey.Text = fmt.Sprintf(publicKey)

	return
}

func (tcg *termuiConsoleGrid) PrepareShardIdForDisplay(shardId int) {
	tcg.pShardId.Text = "ShardId: " + strconv.Itoa(shardId)

	return
}

func (tcg *termuiConsoleGrid) PrepareTxPoolLoadForDisplay(txPoolLoad int) {

	tcg.gTxPoolLoad.Title = fmt.Sprintf("Tx pool load: %v", txPoolLoad)
	tcg.gTxPoolLoad.Percent = 100 * txPoolLoad / 250000
	tcg.gTxPoolLoad.Label = fmt.Sprintf("%v%% ", tcg.gTxPoolLoad.Percent)

	return
}

func (tcg *termuiConsoleGrid) PrepareNumConnectedPeersForDisplay(numConnectedPeers int) {
	tcg.pNumConnectedPeers.Text = fmt.Sprintf("Num connected peers: %v", numConnectedPeers)

	return
}

func (tcg *termuiConsoleGrid) PrepareConcensusInformationsForDisplay(countConsensus int, countLeader int, acceptedBlocks int) {
	tcg.pCountConsensus.Text = fmt.Sprintf("Count consensus group: %v", countConsensus)

	tcg.pCountLeader.Text = fmt.Sprintf("Count leader: %v", countLeader)

	tcg.pCountAcceptedBlocks.Text = fmt.Sprintf("Accepted blocks: %v", acceptedBlocks)

	return
}

func (tcg *termuiConsoleGrid) PrepareSyncInfoForDisplay(nonce, currentRound, synchronizedRound, isSyncing int) {

	isSyncingS := "Synchronized"

	tcg.tSyncInfo.TextStyle = ui.NewStyle(ui.ColorWhite)
	tcg.tSyncInfo.RowSeparator = true
	tcg.tSyncInfo.BorderStyle = ui.NewStyle(ui.ColorWhite)
	tcg.tSyncInfo.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorGreen, ui.ModifierBold)

	if isSyncing == 1 {
		isSyncingS = "IsSyncing"
		tcg.tSyncInfo.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorRed, ui.ModifierBold)

	}

	nonceS := fmt.Sprintf("Nonce: %v", nonce)
	currentRoundS := fmt.Sprintf("Current round: %v", currentRound)
	synchronizedRoundS := fmt.Sprintf("Syncronized round: %v", synchronizedRound)

	tcg.tSyncInfo.Rows = [][]string{
		{isSyncingS, nonceS, synchronizedRoundS, currentRoundS},
	}
}
