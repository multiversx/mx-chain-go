package termuiRenders

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const completRow = 1.0
const completCol = 1.0

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
	wr.grid.Set(
		ui.NewRow(completRow/7, wr.pPublicKey),
		ui.NewRow(completRow/10, wr.tSyncInfo),
		ui.NewRow(completRow/3,
			ui.NewCol(completCol/2,
				ui.NewRow(completRow/4, wr.pShardId),
				ui.NewRow(completRow/4, wr.pCountConsensus),
				ui.NewRow(completRow/4, wr.pCountLeader),
				ui.NewRow(completRow/4, wr.pCountAcceptedBlocks),
			),
			ui.NewCol(completCol/2,
				ui.NewRow(completRow/2, wr.pTxPoolLoad),
				ui.NewRow(completRow/2, wr.pNumConnectedPeers),
			),
		),
		ui.NewRow(completRow/3, wr.lLog),
	)
}

//RefreshData method is used to prepare data that are displayed on grid
func (wr *WidgetsRender) RefreshData(logLines []string) {
	nonceI, _ := wr.termuiConsoleMetrics.Load(core.MetricNonce)
	nonce := nonceI.(int)

	synchronizedRoundI, _ := wr.termuiConsoleMetrics.Load(core.MetricSynchronizedRound)
	synchronizedRound := synchronizedRoundI.(int)

	currentRoundI, _ := wr.termuiConsoleMetrics.Load(core.MetricCurrentRound)
	currentRound := currentRoundI.(int)

	isSyncingI, _ := wr.termuiConsoleMetrics.Load(core.MetricIsSyncing)
	isSyncing := isSyncingI.(int)

	wr.prepareSyncInfoForDisplay(nonce, currentRound, synchronizedRound, isSyncing)

	publicKeyI, _ := wr.termuiConsoleMetrics.Load(core.MetricPublicKey)
	publicKey := publicKeyI.(string)
	wr.preparePublicKeyForDisplay(publicKey)

	shardIdI, _ := wr.termuiConsoleMetrics.Load(core.MetricShardId)
	shardId := shardIdI.(int)
	wr.prepareShardIdForDisplay(shardId)

	numConnectedPeersI, _ := wr.termuiConsoleMetrics.Load(core.MetricNumConnectedPeers)
	numConnectedPeers := numConnectedPeersI.(int)

	txPoolLoadI, _ := wr.termuiConsoleMetrics.Load(core.MetricTxPoolLoad)
	txPoolLoad := txPoolLoadI.(int)
	wr.prepareTxPoolLoadForDisplay(txPoolLoad)
	wr.prepareNumConnectedPeersForDisplay(numConnectedPeers)

	countConsensusI, _ := wr.termuiConsoleMetrics.Load(core.MetricCountConsensus)
	countConsensus := countConsensusI.(int)

	countLeaderI, _ := wr.termuiConsoleMetrics.Load(core.MetricCountLeader)
	countLeader := countLeaderI.(int)

	countAcceptedBlocksI, _ := wr.termuiConsoleMetrics.Load(core.MetricCountAcceptedBlocks)
	countAcceptedBlocks := countAcceptedBlocksI.(int)

	wr.prepareConsensusInformationsForDisplay(countConsensus, countLeader, countAcceptedBlocks)

	wr.prepareListWithLogsForDisplay(logLines)

	return
}

func (wr *WidgetsRender) prepareListWithLogsForDisplay(logData []string) {
	wr.lLog.Title = "Log info"
	wr.lLog.Rows = wr.prepareLogLines(logData)
	wr.lLog.TextStyle = ui.NewStyle(ui.ColorYellow)
	wr.lLog.WrapText = false

	return
}

func (wr *WidgetsRender) preparePublicKeyForDisplay(publicKey string) {
	wr.pPublicKey.Title = fmt.Sprintf("PublicKey")
	wr.pPublicKey.Text = fmt.Sprintf(publicKey)

	return
}

func (wr *WidgetsRender) prepareShardIdForDisplay(shardId int) {
	//Check if node is in meta chain
	//Node is in meta chain if shardId is equals with max uint32 value
	if shardId == int(^uint32(0)) {
		wr.pShardId.Text = "ShardId: meta"
	} else {
		wr.pShardId.Text = fmt.Sprintf("ShardId: %v", shardId)
	}

	return
}

func (wr *WidgetsRender) prepareTxPoolLoadForDisplay(txPoolLoad int) {
	wr.pTxPoolLoad.Text = fmt.Sprintf("Tx pool load: %v", txPoolLoad)

	return
}

func (wr *WidgetsRender) prepareNumConnectedPeersForDisplay(numConnectedPeers int) {
	wr.pNumConnectedPeers.Text = fmt.Sprintf("Num connected peers: %v", numConnectedPeers)

	return
}

func (wr *WidgetsRender) prepareConsensusInformationsForDisplay(countConsensus int, countLeader int, acceptedBlocks int) {
	wr.pCountConsensus.Text = fmt.Sprintf("Count consensus group: %v", countConsensus)
	wr.pCountLeader.Text = fmt.Sprintf("Count leader: %v", countLeader)
	wr.pCountAcceptedBlocks.Text = fmt.Sprintf("Accepted blocks: %v", acceptedBlocks)

	return
}

func (wr *WidgetsRender) prepareSyncInfoForDisplay(nonce, currentRound, synchronizedRound, isSyncing int) {
	isSyncingS := "Synchronized"
	wr.tSyncInfo.TextStyle = ui.NewStyle(ui.ColorWhite)
	wr.tSyncInfo.RowSeparator = true
	wr.tSyncInfo.BorderStyle = ui.NewStyle(ui.ColorWhite)
	wr.tSyncInfo.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorGreen, ui.ModifierBold)

	if isSyncing == 1 {
		isSyncingS = "IsSyncing"
		wr.tSyncInfo.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorRed, ui.ModifierBold)

	}

	nonceS := fmt.Sprintf("Nonce: %v", nonce)
	currentRoundS := fmt.Sprintf("Current round: %v", currentRound)
	synchronizedRoundS := fmt.Sprintf("Syncronized round: %v", synchronizedRound)

	wr.tSyncInfo.Rows = [][]string{
		{isSyncingS, nonceS, synchronizedRoundS, currentRoundS},
	}
}

func (wr *WidgetsRender) prepareLogLines(logData []string) []string {
	logDataLen := len(logData)

	if logDataLen > numLogLinesSmallLog {
		return logData[logDataLen-numLogLinesSmallLog : logDataLen]
	}
	return logData
}
