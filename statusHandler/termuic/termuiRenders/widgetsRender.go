package termuiRenders

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/sharding"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const completeRow = 1.0
const completeCol = 1.0

const numLogLinesSmallLog = 10

//WidgetsRender will define termui widgets that need to display a termui console
type WidgetsRender struct {
	grid       *ui.Grid
	lLog       *widgets.List
	pPublicKey *widgets.Paragraph
	pShardId   *widgets.Paragraph

	pCountLeader         *widgets.Paragraph
	pCountConsensus      *widgets.Paragraph
	pCountAcceptedBlocks *widgets.Paragraph

	pTxPoolLoad        *widgets.Paragraph
	pNumConnectedPeers *widgets.Paragraph

	tSyncInfo *widgets.Table

	lInstanceInfo        *widgets.List
	lChainInfo           *widgets.List
	termuiConsoleMetrics *sync.Map
}

//NewWidgetsRender method will create new WidgetsRender that display termui console
func NewWidgetsRender(metricData *sync.Map, grid *ui.Grid) *WidgetsRender {
	self := &WidgetsRender{
		termuiConsoleMetrics: metricData,
		grid:                 grid,
	}
	self.initWidgets()
	self.setGrid()

	return self
}

func (wr *WidgetsRender) initWidgets() {
	wr.pPublicKey = widgets.NewParagraph()
	wr.pShardId = widgets.NewParagraph()

	wr.pCountLeader = widgets.NewParagraph()
	wr.pCountConsensus = widgets.NewParagraph()
	wr.pCountAcceptedBlocks = widgets.NewParagraph()

	wr.pTxPoolLoad = widgets.NewParagraph()
	wr.pNumConnectedPeers = widgets.NewParagraph()

	wr.lLog = widgets.NewList()

	wr.tSyncInfo = widgets.NewTable()
	wr.tSyncInfo.Rows = [][]string{
		{"", "", "", ""},
	}

	wr.lInstanceInfo = widgets.NewList()
	wr.lChainInfo = widgets.NewList()
}

func (wr *WidgetsRender) setGrid() {

	publicKeyHeight := completeRow / 7
	syncInfoHeight := completeRow / 10
	consensusInfoHeight := completeRow / 3
	consensusInfoWidth := completeCol / 2
	consensusRowInfoWidth := completeRow / 4
	logHeight := completeRow / 3
	txPoolLoadColWidth := completeCol / 2
	txPoolLoadRowHeight := completeRow / 2
	numConnectedPeersRowHeight := completeRow / 2

	wr.grid.Set(
		ui.NewRow(publicKeyHeight, wr.pPublicKey),
		ui.NewRow(syncInfoHeight, wr.tSyncInfo),
		ui.NewRow(consensusInfoHeight,
			ui.NewCol(consensusInfoWidth,
				ui.NewRow(consensusRowInfoWidth, wr.pShardId),
				ui.NewRow(consensusRowInfoWidth, wr.pCountConsensus),
				ui.NewRow(consensusRowInfoWidth, wr.pCountLeader),
				ui.NewRow(consensusRowInfoWidth, wr.pCountAcceptedBlocks),
			),
			ui.NewCol(txPoolLoadColWidth,
				ui.NewRow(txPoolLoadRowHeight, wr.pTxPoolLoad),
				ui.NewRow(numConnectedPeersRowHeight, wr.pNumConnectedPeers),
			),
		),
		ui.NewRow(logHeight, wr.lLog),
	)
}

//RefreshData method is used to prepare data that are displayed on grid
func (wr *WidgetsRender) RefreshData(logLines []string) {
	wr.prepareSyncData()
	wr.preparePkData()
	wr.prepareShardIdData()
	wr.prepareNumConnectedPeersData()
	wr.prepareTxPoolLoadData()
	wr.prepareConsensusDataForDisplay()
	wr.prepareListWithLogsForDisplay(logLines)

	return
}

func (wr *WidgetsRender) prepareSyncData() {
	nonceI, _ := wr.termuiConsoleMetrics.Load(core.MetricNonce)
	nonce := nonceI.(uint64)

	synchronizedRoundI, _ := wr.termuiConsoleMetrics.Load(core.MetricSynchronizedRound)
	synchronizedRound := synchronizedRoundI.(uint64)

	currentRoundI, _ := wr.termuiConsoleMetrics.Load(core.MetricCurrentRound)
	currentRound := currentRoundI.(int64)

	isSyncingI, _ := wr.termuiConsoleMetrics.Load(core.MetricIsSyncing)
	isSyncing := isSyncingI.(uint64)

	wr.prepareSyncInfoForDisplay(nonce, currentRound, synchronizedRound, isSyncing)
}

func (wr *WidgetsRender) preparePkData() {
	publicKeyI, _ := wr.termuiConsoleMetrics.Load(core.MetricPublicKey)
	publicKey := publicKeyI.(string)
	wr.preparePublicKeyForDisplay(publicKey)
}

func (wr *WidgetsRender) prepareShardIdData() {
	shardIdI, _ := wr.termuiConsoleMetrics.Load(core.MetricShardId)
	shardId := shardIdI.(uint64)
	wr.prepareShardIdForDisplay(shardId)
}

func (wr *WidgetsRender) prepareNumConnectedPeersData() {
	numConnectedPeersI, _ := wr.termuiConsoleMetrics.Load(core.MetricNumConnectedPeers)
	numConnectedPeers := numConnectedPeersI.(int64)
	wr.prepareNumConnectedPeersForDisplay(numConnectedPeers)
}

func (wr *WidgetsRender) prepareTxPoolLoadData() {
	txPoolLoadI, _ := wr.termuiConsoleMetrics.Load(core.MetricTxPoolLoad)
	txPoolLoad := txPoolLoadI.(int64)
	wr.prepareTxPoolLoadForDisplay(txPoolLoad)
}

func (wr *WidgetsRender) prepareListWithLogsForDisplay(logData []string) {
	wr.lLog.Title = "Log info"
	wr.lLog.Rows = wr.prepareLogLines(logData)
	wr.lLog.TextStyle = ui.NewStyle(ui.ColorYellow)
	wr.lLog.WrapText = false

	return
}

func (wr *WidgetsRender) prepareConsensusDataForDisplay() {
	countConsensusI, _ := wr.termuiConsoleMetrics.Load(core.MetricCountConsensus)
	countConsensus := countConsensusI.(uint64)

	countLeaderI, _ := wr.termuiConsoleMetrics.Load(core.MetricCountLeader)
	countLeader := countLeaderI.(uint64)

	countAcceptedBlocksI, _ := wr.termuiConsoleMetrics.Load(core.MetricCountAcceptedBlocks)
	countAcceptedBlocks := countAcceptedBlocksI.(uint64)

	wr.prepareConsensusInformationsForDisplay(countConsensus, countLeader, countAcceptedBlocks)
}

func (wr *WidgetsRender) preparePublicKeyForDisplay(publicKey string) {
	wr.pPublicKey.Title = fmt.Sprintf("PublicKey")
	wr.pPublicKey.Text = fmt.Sprintf(publicKey)

	return
}

func (wr *WidgetsRender) prepareShardIdForDisplay(shardId uint64) {
	//Check if node is in meta chain
	//Node is in meta chain if shardId is equals with max uint32 value
	if shardId == uint64(sharding.MetachainShardId) {
		wr.pShardId.Text = "ShardId: meta"
	} else {
		wr.pShardId.Text = fmt.Sprintf("ShardId: %v", shardId)
	}

	return
}

func (wr *WidgetsRender) prepareTxPoolLoadForDisplay(txPoolLoad int64) {
	wr.pTxPoolLoad.Text = fmt.Sprintf("Number of tx in pool: %v", txPoolLoad)

	return
}

func (wr *WidgetsRender) prepareNumConnectedPeersForDisplay(numConnectedPeers int64) {
	wr.pNumConnectedPeers.Text = fmt.Sprintf("Num connected peers: %v", numConnectedPeers)

	return
}

func (wr *WidgetsRender) prepareConsensusInformationsForDisplay(countConsensus uint64, countLeader uint64, acceptedBlocks uint64) {
	wr.pCountConsensus.Text = fmt.Sprintf("Consensus group participant count: %v", countConsensus)
	wr.pCountLeader.Text = fmt.Sprintf("Elected consensus leader count: %v", countLeader)
	wr.pCountAcceptedBlocks.Text = fmt.Sprintf("Consensus proposed & accepted blocks: %v", acceptedBlocks)

	return
}

func (wr *WidgetsRender) prepareSyncInfoForDisplay(nonce uint64, currentRound int64, synchronizedRound, syncStatus uint64) {
	isSyncingS := "Status: synchronized"
	wr.tSyncInfo.TextStyle = ui.NewStyle(ui.ColorWhite)
	wr.tSyncInfo.RowSeparator = true
	wr.tSyncInfo.BorderStyle = ui.NewStyle(ui.ColorWhite)
	wr.tSyncInfo.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorGreen, ui.ModifierBold)

	if syncStatus == 1 {
		isSyncingS = "Status: syncing"
		wr.tSyncInfo.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorRed, ui.ModifierBold)

	}

	nonceS := fmt.Sprintf("Current node block nonce: %v", nonce)
	currentRoundS := fmt.Sprintf("Current round: %v", currentRound)
	synchronizedRoundS := fmt.Sprintf("Syncronized round: %v", synchronizedRound)

	wr.tSyncInfo.Rows = [][]string{
		{isSyncingS, nonceS, synchronizedRoundS, currentRoundS},
	}
}

//TODO duplicate code next pull request refactor
func (wr *WidgetsRender) prepareLogLines(logData []string) []string {
	logDataLen := len(logData)

	if logDataLen > numLogLinesSmallLog {
		return logData[logDataLen-numLogLinesSmallLog : logDataLen]
	}
	return logData
}
