package termuic

import (
	"fmt"
	"strconv"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type termuiConsoleGrid struct {
	grid              *ui.Grid
	pNonce            *widgets.Paragraph
	pSynchronizeRound *widgets.Paragraph
	pCurrentRound     *widgets.Paragraph
	pIsSyncing        *widgets.Paragraph
	lLog              *widgets.List
	pPublicKey        *widgets.Paragraph
	pShardId          *widgets.Paragraph

	pCountLeader         *widgets.Paragraph
	pCountConsensus      *widgets.Paragraph
	pCountAcceptedBlocks *widgets.Paragraph

	slTxPoolLoad       *widgets.Sparkline
	slNumConnectedPeer *widgets.Sparkline

	slgNode *widgets.SparklineGroup

	slTxPoolLoadData        []float64
	slNumConnectedPeersData []float64
}

//NewtermuiConsoleGrid initialize struct termuiConsoleGrid
func NewtermuiConsoleGrid() *termuiConsoleGrid {

	self := termuiConsoleGrid{}

	self.initWidgets()

	self.setupGrid()

	self.slTxPoolLoadData = make([]float64, 0, 200)
	self.slNumConnectedPeersData = make([]float64, 0, 200)

	return &self
}

func (tcg *termuiConsoleGrid) initWidgets() {
	tcg.pNonce = widgets.NewParagraph()
	tcg.pSynchronizeRound = widgets.NewParagraph()
	tcg.pCurrentRound = widgets.NewParagraph()
	tcg.pIsSyncing = widgets.NewParagraph()
	tcg.pPublicKey = widgets.NewParagraph()
	tcg.pShardId = widgets.NewParagraph()

	tcg.pCountLeader = widgets.NewParagraph()
	tcg.pCountConsensus = widgets.NewParagraph()
	tcg.pCountAcceptedBlocks = widgets.NewParagraph()

	tcg.slTxPoolLoad = widgets.NewSparkline()
	tcg.slNumConnectedPeer = widgets.NewSparkline()
	tcg.slgNode = widgets.NewSparklineGroup(tcg.slTxPoolLoad, tcg.slNumConnectedPeer)

	tcg.lLog = widgets.NewList()

}

func (tcg *termuiConsoleGrid) setupGrid() {
	tcg.grid = ui.NewGrid()

	tcg.grid.Set(
		ui.NewRow(1.0/7, tcg.pPublicKey),
		ui.NewRow(1.0/3,
			ui.NewCol(1.0/2,
				ui.NewRow(1.0/4, tcg.pShardId),
				ui.NewRow(1.0/4, tcg.pCountConsensus),
				ui.NewRow(1.0/4, tcg.pCountLeader),
				ui.NewRow(1.0/4, tcg.pCountAcceptedBlocks),
			),
			ui.NewCol(1.0/2, tcg.slgNode),
		),
		ui.NewRow(1.0/7,
			ui.NewCol(1.0/4, tcg.pNonce),
			ui.NewCol(1.0/4, tcg.pCurrentRound),
			ui.NewCol(1.0/4, tcg.pSynchronizeRound),
			ui.NewCol(1.0/4, tcg.pIsSyncing),
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

func (tcg *termuiConsoleGrid) PrepareNonceForDisplay(Nonce int) {
	tcg.pNonce.Text = "Nonce : " + strconv.Itoa(Nonce)
	tcg.pNonce.TextStyle = ui.Style{
		Fg:       ui.ColorBlack,
		Bg:       ui.ColorGreen,
		Modifier: 1,
	}

	return
}

func (tcg *termuiConsoleGrid) PrepareSynchronizedRoundForDisplay(synchronizedRound int) {
	tcg.pSynchronizeRound.Text = "Synchronized Round : " + strconv.Itoa(synchronizedRound)
	return
}

func (tcg *termuiConsoleGrid) PrepareIsSyncingForDisplay(isSyncing int) {
	tcg.pIsSyncing.Text = "Synchronized"

	tcg.pIsSyncing.TextStyle = ui.Style{
		Fg:       ui.ColorBlack,
		Bg:       ui.ColorGreen,
		Modifier: 1,
	}

	if isSyncing == 1 {
		tcg.pIsSyncing.Text = "IsSyncing"
		tcg.pIsSyncing.TextStyle.Bg = ui.ColorRed
	}

	return
}

func (tcg *termuiConsoleGrid) PrepareCurrentRoundForDisplay(currentRound int) {
	tcg.pCurrentRound.Text = "Current Round : " + strconv.Itoa(currentRound)
	return
}

func (tcg *termuiConsoleGrid) PrepareListWithLogsForDisplay(logData []string) {

	tcg.lLog.Title = "Log info"
	tcg.lLog.Rows = logData
	tcg.lLog.TextStyle = ui.NewStyle(ui.ColorYellow)
	tcg.lLog.WrapText = false

	return
}

func (tcg *termuiConsoleGrid) PreparePublicKeyForDisplay(publicKey string) {
	tcg.pPublicKey.Text = "PublicKey : " + publicKey

	return
}

func (tcg *termuiConsoleGrid) PrepareShardIdForDisplay(shardId int) {
	tcg.pShardId.Text = "ShardId : " + strconv.Itoa(shardId)

	return
}

func (tcg *termuiConsoleGrid) PrepareSparkLineGroupForDisplay(txPoolLoad int, numConnectedPeers int) {

	if len(tcg.slTxPoolLoadData) >= 50 {
		tcg.slTxPoolLoadData = tcg.slTxPoolLoadData[1:50]
	}
	tcg.slTxPoolLoadData = append(tcg.slTxPoolLoadData, float64(txPoolLoad))
	tcg.slTxPoolLoad.Data = tcg.slTxPoolLoadData
	tcg.slTxPoolLoad.Title = fmt.Sprintf("Tx Pool Load %v", txPoolLoad)

	if len(tcg.slNumConnectedPeersData) >= 50 {
		tcg.slNumConnectedPeersData = tcg.slNumConnectedPeersData[1:50]
	}
	tcg.slNumConnectedPeersData = append(tcg.slNumConnectedPeersData, float64(numConnectedPeers))
	tcg.slNumConnectedPeer.Data = tcg.slNumConnectedPeersData
	tcg.slNumConnectedPeer.Title = fmt.Sprintf("Num Connecter Peers %v", numConnectedPeers)

	return
}

func (tcg *termuiConsoleGrid) PrepareConcensusInformationsForDisplay(countConsensus int, countLeader int, acceptedBlocks int) {
	tcg.pCountConsensus.Text = fmt.Sprintf("Count Consensus Group %v", countConsensus)

	tcg.pCountLeader.Text = fmt.Sprintf("Count Leader %v", countLeader)

	tcg.pCountAcceptedBlocks.Text = fmt.Sprintf("Accepted Blocks %v", acceptedBlocks)

	return
}
