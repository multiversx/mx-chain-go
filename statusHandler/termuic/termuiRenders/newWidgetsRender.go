package termuiRenders

import (
	"fmt"
	"math"
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
	container    *DrawableContainer
	lLog         *widgets.List
	canvas       *ui.Canvas
	instanceInfo *widgets.Paragraph
	chainInfo    *widgets.Paragraph

	cpuLoad     *widgets.Gauge
	memoryLoad  *widgets.Gauge
	networkRecv *widgets.Gauge
	networkSent *widgets.Gauge

	tSyncInfo *widgets.Table

	lInstanceInfo        *widgets.List
	lChainInfo           *widgets.List
	termuiConsoleMetrics *sync.Map
}

//NewWidgetsRender method will create new WidgetsRender that display termui console
func NewWidgetsRender2(metricData *sync.Map, grid *DrawableContainer) *WidgetsRender2 {
	self := &WidgetsRender2{
		termuiConsoleMetrics: metricData,
		container:            grid,
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
	wr.networkRecv = widgets.NewGauge()
	wr.networkSent = widgets.NewGauge()

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
		ui.NewRow(10.0/22, wr.instanceInfo),
		ui.NewRow(12.0/22, wr.chainInfo))

	gridRight := ui.NewGrid()
	gridRight.Set(ui.NewRow(1.0/4, wr.cpuLoad),
		ui.NewRow(1.0/4, wr.memoryLoad),
		ui.NewRow(1.0/4, wr.networkRecv),
		ui.NewRow(1.0/4, wr.networkSent))

	gridBottom := ui.NewGrid()
	gridBottom.Set(ui.NewRow(1.0, wr.lLog))

	wr.container.SetTopLeft(gridLeft)
	wr.container.SetTopRight(gridRight)
	wr.container.SetBottom(gridBottom)
}

//RefreshData method is used to prepare data that are displayed on container
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

	syncStatus := wr.getFromCacheAsUint64(core.MetricIsSyncing)
	syncingStr := fmt.Sprintf("undefined %d", syncStatus)
	switch syncStatus {
	case 1:
		syncingStr = "currently synching"
	case 0:
		syncingStr = "synchronized"
	}
	rows = fmt.Sprintf("Status: %s\n", syncingStr)

	memTxPoolSize := wr.getFromCacheAsUint64(core.MetricTxPoolLoad)
	rows = fmt.Sprintf("%sNumber of transactions in pool: %d\n", rows, memTxPoolSize)

	rows = fmt.Sprintf("%s\n", rows)

	nonce := wr.getFromCacheAsUint64(core.MetricNonce)
	probableHighestNonce := wr.getFromCacheAsUint64(core.MetricProbableHighestNonce)
	rows = fmt.Sprintf("%sCurrent synchronized block nonce: %v / %v\n", rows, nonce, probableHighestNonce)

	synchronizedRound := wr.getFromCacheAsUint64(core.MetricSynchronizedRound)
	currentRound := wr.getFromCacheAsUint64(core.MetricCurrentRound)
	rows = fmt.Sprintf("%sCurrent consensus round: %v / %v\n", rows, synchronizedRound, currentRound)

	consensusRoundTime := wr.getFromCacheAsUint64(core.MetricRoundTime)
	rows = fmt.Sprintf("%sConsensus round time: %ds\n", rows, consensusRoundTime)

	rows = fmt.Sprintf("%s\n", rows)

	numLiveValidators := wr.getFromCacheAsUint64(core.MetricLiveValidatorNodes)
	rows = fmt.Sprintf("%sLive validator nodes: %v\n", rows, numLiveValidators)

	numConnectedNodes := wr.getFromCacheAsUint64(core.MetricConnectedNodes)
	rows = fmt.Sprintf("%sNetwork connected nodes: %v\n", rows, numConnectedNodes)

	numConnectedPeers := wr.getFromCacheAsUint64(core.MetricNumConnectedPeers)
	rows = fmt.Sprintf("%sThis node is connected to %v peers\n", rows, numConnectedPeers)

	return "Chain info", rows
}

func (wr *WidgetsRender2) prepareListWithLogsForDisplay(logData []string) {
	bordersHeight := 2
	wr.lLog.Title = "Log info"
	wr.lLog.TextStyle = ui.NewStyle(ui.ColorWhite)
	wr.lLog.Rows = wr.prepareLogLines(logData, wr.lLog.Size().Y-bordersHeight)
	wr.lLog.WrapText = true
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
	if size < 0 {
		size = 0
	}
	if logDataLen > size {
		return logData[logDataLen-size : logDataLen]
	}
	return logData
}

func (wr *WidgetsRender2) prepareLoads() {
	cpuLoadPercentI, _ := wr.termuiConsoleMetrics.Load(core.MetricCpuLoadPercent)
	cpuLoadPercent := cpuLoadPercentI.(uint64)
	wr.cpuLoad.Title = "CPU Load"
	wr.cpuLoad.Percent = int(cpuLoadPercent)

	memLoadPercentI, _ := wr.termuiConsoleMetrics.Load(core.MetricMemLoadPercent)
	memLoadPercent := memLoadPercentI.(uint64)
	memTotalMemoryBytesObj, _ := wr.termuiConsoleMetrics.Load(core.MetricTotalMem)
	memTotalMemoryBytes := memTotalMemoryBytesObj.(uint64)
	wr.memoryLoad.Title = "Memory load"
	wr.memoryLoad.Percent = int(memLoadPercent)
	wr.memoryLoad.Label = fmt.Sprintf("%d%% / T: %s", memLoadPercent, core.ConvertBytes(memTotalMemoryBytes))

	recvLoad := wr.getFromCacheAsUint64(core.MetricNetworkRecvPercent)
	recvBps := wr.getFromCacheAsUint64(core.MetricNetworkRecvBps)
	recvBpsPeak := wr.getFromCacheAsUint64(core.MetricNetworkRecvBpsPeak)
	wr.networkRecv.Title = "Network - recv:"
	wr.networkRecv.Percent = int(recvLoad)
	wr.networkRecv.Label = fmt.Sprintf("%d%% / rate: %s/s / peak rate: %s/s",
		recvLoad, core.ConvertBytes(recvBps), core.ConvertBytes(recvBpsPeak))

	sentLoad := wr.getFromCacheAsUint64(core.MetricNetworkSentPercent)
	sentBps := wr.getFromCacheAsUint64(core.MetricNetworkSentBps)
	sentBpsPeak := wr.getFromCacheAsUint64(core.MetricNetworkSentBpsPeak)
	wr.networkSent.Title = "Network - sent:"
	wr.networkSent.Percent = int(sentLoad)
	wr.networkSent.Label = fmt.Sprintf("%d%% / rate: %s/s / peak rate: %s/s",
		sentLoad, core.ConvertBytes(sentBps), core.ConvertBytes(sentBpsPeak))
}

func (wr *WidgetsRender2) getFromCacheAsUint64(metric string) uint64 {
	val, ok := wr.termuiConsoleMetrics.Load(metric)
	if !ok {
		return math.MaxUint64
	}

	valUint64, ok := val.(uint64)
	if !ok {
		return math.MaxUint64
	}

	return valUint64
}

func (wr *WidgetsRender2) getNetworkRecvStats() {

}
