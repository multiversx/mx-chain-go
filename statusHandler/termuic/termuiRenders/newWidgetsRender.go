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
const numLogLinesSmallLog = 10

//WidgetsRender will define termui widgets that need to display a termui console
type WidgetsRender2 struct {
	grid         *DrawableContainer
	lLog         *widgets.List
	canvas       *ui.Canvas
	instanceInfo *widgets.Paragraph
	chainInfo    *widgets.Paragraph

	cpuLoad     *widgets.Gauge
	memoryLoad  *widgets.Gauge
	networkLoad *widgets.Gauge
	txPoolLoad  *widgets.Gauge

	tSyncInfo *widgets.Table

	lInstanceInfo        *widgets.List
	lChainInfo           *widgets.List
	termuiConsoleMetrics *sync.Map
}

//NewWidgetsRender method will create new WidgetsRender that display termui console
func NewWidgetsRender2(metricData *sync.Map, grid *DrawableContainer) *WidgetsRender2 {
	self := &WidgetsRender2{
		termuiConsoleMetrics: metricData,
		grid:                 grid,
	}
	self.initWidgets()
	self.setGrid()

	return self
}

func (wr *WidgetsRender2) initWidgets() {
	wr.instanceInfo = widgets.NewParagraph()
	wr.instanceInfo.Text = ""

	wr.chainInfo = widgets.NewParagraph()
	wr.chainInfo.Text = ""

	wr.cpuLoad = widgets.NewGauge()
	wr.memoryLoad = widgets.NewGauge()
	wr.networkLoad = widgets.NewGauge()
	wr.txPoolLoad = widgets.NewGauge()

	wr.lLog = widgets.NewList()

	wr.tSyncInfo = widgets.NewTable()
	wr.tSyncInfo.Rows = [][]string{
		{"", "", "", ""},
	}

	wr.lInstanceInfo = widgets.NewList()
	wr.lChainInfo = widgets.NewList()

	wr.canvas = ui.NewCanvas()
}

func (wr *WidgetsRender2) setGrid() {

	gridLeft := ui.NewGrid()

	gridLeft.Set(
		ui.NewRow(1.0/2, wr.instanceInfo),
		ui.NewRow(1.0/2, wr.chainInfo))

	gridRight := ui.NewGrid()
	gridRight.Set(ui.NewRow(1.0/4, wr.cpuLoad),
		ui.NewRow(1.0/4, wr.memoryLoad),
		ui.NewRow(1.0/4, wr.networkLoad),
		ui.NewRow(1.0/4, wr.txPoolLoad))

	gridBottom := ui.NewGrid()
	gridBottom.Set(ui.NewRow(1.0, wr.lLog))

	wr.grid.SetTopLeft(gridLeft)
	wr.grid.SetTopRight(gridRight)
	wr.grid.SetBottom(gridBottom)
}

//RefreshData method is used to prepare data that are displayed on grid
func (wr *WidgetsRender2) RefreshData(logLines []string) {
	title, rows := wr.prepareInstanceInfo()
	wr.instanceInfo.Title = title
	wr.instanceInfo.WrapText = true
	wr.instanceInfo.Text = rows

	title, rows = wr.prepareChainInfo()
	wr.chainInfo.Title = title
	wr.instanceInfo.WrapText = false
	wr.chainInfo.Text = rows

	wr.prepareListWithLogsForDisplay(logLines)

	wr.prepareLoads()

	return
}

func (wr *WidgetsRender2) prepareInstanceInfo() (string, string) {
	rows := ""

	publicKeyI, _ := wr.termuiConsoleMetrics.Load(core.MetricPublicKeyTxSign)
	publicKey := publicKeyI.(string)
	rows = fmt.Sprintf("%s", fmt.Sprintf("Public key TxSign: "+publicKey))

	publicKeyI, _ = wr.termuiConsoleMetrics.Load(core.MetricPublicKeyBlockSign)
	publicKey = publicKeyI.(string)
	rows = fmt.Sprintf("%s\n%s", rows, fmt.Sprintf("%s", fmt.Sprintf("Public key BlockSign: "+publicKey)))

	shardIdI, _ := wr.termuiConsoleMetrics.Load(core.MetricShardId)
	shardId := shardIdI.(uint64)
	shardIdStr := ""
	if shardId == uint64(sharding.MetachainShardId) {
		shardIdStr = "meta"
	} else {
		shardIdStr = fmt.Sprintf("%v", shardId)
	}

	rows = fmt.Sprintf("%s\n%s", rows, fmt.Sprintf("ShardID: "+shardIdStr))

	rows = fmt.Sprintf("%s\n", rows)

	instanceTypeI, _ := wr.termuiConsoleMetrics.Load(core.MetricNodeType)
	instanceType := instanceTypeI.(string)
	rows = fmt.Sprintf("%s\n%s", rows, fmt.Sprintf("Instance type: %s", instanceType))

	countConsensusI, _ := wr.termuiConsoleMetrics.Load(core.MetricCountConsensus)
	countConsensus := countConsensusI.(uint64)
	rows = fmt.Sprintf("%s\n%s", rows, fmt.Sprintf("Consensus group participation count: %v", countConsensus))

	countLeaderI, _ := wr.termuiConsoleMetrics.Load(core.MetricCountLeader)
	countLeader := countLeaderI.(uint64)
	rows = fmt.Sprintf("%s\n%s", rows, fmt.Sprintf("Elected consensus leader count: %v", countLeader))

	countAcceptedBlocksI, _ := wr.termuiConsoleMetrics.Load(core.MetricCountAcceptedBlocks)
	countAcceptedBlocks := countAcceptedBlocksI.(uint64)
	rows = fmt.Sprintf("%s\n%s", rows, fmt.Sprintf("Consensus proposed & accepted blocks: %v", countAcceptedBlocks))

	return "Instance info", rows
}

func (wr *WidgetsRender2) prepareChainInfo() (string, string) {
	rows := ""

	isSyncingI, _ := wr.termuiConsoleMetrics.Load(core.MetricIsSyncing)
	isSyncing := isSyncingI.(uint64)
	isSyncingStr := "synchronized"
	if isSyncing == 1 {
		isSyncingStr = "currently synching"
	}

	rows = fmt.Sprintf("%s", fmt.Sprintf("Status: %s", isSyncingStr))

	rows = fmt.Sprintf("%s\n", rows)

	nonceI, _ := wr.termuiConsoleMetrics.Load(core.MetricNonce)
	nonce := nonceI.(uint64)

	probableHighestNonceI, _ := wr.termuiConsoleMetrics.Load(core.MetricProbableHighestNonce)
	probableHighestNonce := probableHighestNonceI.(uint64)
	rows = fmt.Sprintf("%s\n%s", rows, fmt.Sprintf("Current synchronized block nonce: %v/%v", nonce, probableHighestNonce))

	synchronizedRoundI, _ := wr.termuiConsoleMetrics.Load(core.MetricSynchronizedRound)
	synchronizedRound := synchronizedRoundI.(uint64)

	currentRoundI, _ := wr.termuiConsoleMetrics.Load(core.MetricCurrentRound)
	currentRound := currentRoundI.(int64)
	rows = fmt.Sprintf("%s\n%s", rows, fmt.Sprintf("Current consensus round: %v/%v", synchronizedRound, currentRound))

	rows = fmt.Sprintf("%s\n", rows)

	numLiveNodesI, _ := wr.termuiConsoleMetrics.Load(core.MetricLiveValidatorNodes)
	numLiveNodes := numLiveNodesI.(int64)
	rows = fmt.Sprintf("%s\n%s", rows, fmt.Sprintf("Live validator nodes: %v", numLiveNodes))

	numConnectedPeersI, _ := wr.termuiConsoleMetrics.Load(core.MetricNumConnectedPeers)
	numConnectedPeers := numConnectedPeersI.(int64)
	rows = fmt.Sprintf("%s\n%s", rows, fmt.Sprintf("Connected peers: %v", numConnectedPeers))

	return "Chain info", rows
}

func (wr *WidgetsRender2) prepareListWithLogsForDisplay(logData []string) {
	wr.lLog.Title = "Log info"
	wr.lLog.TextStyle = ui.NewStyle(ui.ColorWhite)
	wr.lLog.Rows = wr.prepareLogLines(logData, wr.lLog.Size().Y)
	wr.lLog.WrapText = false
	return
}

func (wr *WidgetsRender2) prepareSyncInfoForDisplay(nonce uint64, currentRound int64, synchronizedRound,
	syncStatus uint64) {
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
func (wr *WidgetsRender2) prepareLogLines(logData []string, size int) []string {
	logDataLen := len(logData)

	if logDataLen > size {
		return logData[logDataLen-size : logDataLen]
	}
	return logData
}

func (wr *WidgetsRender2) prepareLoads() {
	cpuLoadPercentI, _ := wr.termuiConsoleMetrics.Load(core.MetricCpuLoadPercent)
	cpuLoadPercent := cpuLoadPercentI.(int64)
	wr.cpuLoad.Title = "CPU Load"
	wr.cpuLoad.Percent = int(cpuLoadPercent)

	memLoadPercentI, _ := wr.termuiConsoleMetrics.Load(core.MetricMemLoadPercent)
	memLoadPercent := memLoadPercentI.(int64)
	wr.memoryLoad.Title = "Memory Load"
	wr.memoryLoad.Percent = int(memLoadPercent)

	networkLoadPercentI, _ := wr.termuiConsoleMetrics.Load(core.MetricNetworkLoadPercent)
	networkLoadPercent := networkLoadPercentI.(int64)
	wr.networkLoad.Title = "Network Load"
	wr.networkLoad.Percent = int(networkLoadPercent)

	txPoolLoadI, _ := wr.termuiConsoleMetrics.Load(core.MetricTxPoolLoad)
	txPoolLoad := txPoolLoadI.(int64)
	wr.txPoolLoad.Title = "Transation Pool Load"
	wr.txPoolLoad.Percent = int(txPoolLoad)
	wr.txPoolLoad.Percent = 65
}
