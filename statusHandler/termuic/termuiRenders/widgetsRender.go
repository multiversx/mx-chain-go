package termuiRenders

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/sharding"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const statusSynching = "currently synching"
const statusSynchronized = "synchronized"

//WidgetsRender will define termui widgets that need to display a termui console
type WidgetsRender struct {
	container    *DrawableContainer
	lLog         *widgets.List
	instanceInfo *widgets.Paragraph
	chainInfo    *widgets.Table

	cpuLoad     *widgets.Gauge
	memoryLoad  *widgets.Gauge
	networkRecv *widgets.Gauge
	networkSent *widgets.Gauge

	termuiConsoleMetrics *sync.Map
}

//NewWidgetsRender method will create new WidgetsRender that display termui console
func NewWidgetsRender(metricData *sync.Map, grid *DrawableContainer) *WidgetsRender {
	self := &WidgetsRender{
		termuiConsoleMetrics: metricData,
		container:            grid,
	}
	self.initWidgets()
	self.setGrid()

	return self
}

func (wr *WidgetsRender) initWidgets() {
	wr.instanceInfo = widgets.NewParagraph()
	wr.instanceInfo.Text = ""

	wr.chainInfo = widgets.NewTable()
	wr.chainInfo.Rows = [][]string{{"", "", "", ""}}

	wr.cpuLoad = widgets.NewGauge()
	wr.memoryLoad = widgets.NewGauge()
	wr.networkRecv = widgets.NewGauge()
	wr.networkSent = widgets.NewGauge()

	wr.lLog = widgets.NewList()
}

func (wr *WidgetsRender) setGrid() {
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
func (wr *WidgetsRender) RefreshData(logLines []string) {
	title, rows := wr.prepareInstanceInfo()
	wr.instanceInfo.Title = title
	wr.instanceInfo.WrapText = true
	wr.instanceInfo.Text = rows

	title2, rows2 := wr.prepareChainInfo()
	wr.chainInfo.Title = title2
	wr.instanceInfo.WrapText = false
	wr.chainInfo.RowSeparator = false
	wr.chainInfo.Rows = rows2

	status := rows2[0][0]
	if strings.Contains(status, statusSynchronized) {
		wr.chainInfo.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorGreen)
	} else {
		wr.chainInfo.RowStyles[0] = ui.NewStyle(ui.ColorWhite, ui.ColorYellow)
	}

	wr.prepareListWithLogsForDisplay(logLines)
	wr.prepareLoads()
}

func (wr *WidgetsRender) prepareInstanceInfo() (string, string) {
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

func (wr *WidgetsRender) prepareChainInfo() (string, [][]string) {
	rows := make([][]string, 0)

	syncStatus := wr.getFromCacheAsUint64(core.MetricIsSyncing)
	syncingStr := fmt.Sprintf("undefined %d", syncStatus)
	switch syncStatus {
	case 1:
		syncingStr = statusSynching
	case 0:
		syncingStr = statusSynchronized
	}

	rows = append(rows, append(make([]string, 0), fmt.Sprintf("Status: %s", syncingStr)))

	memTxPoolSize := wr.getFromCacheAsUint64(core.MetricTxPoolLoad)
	rows = append(rows, append(make([]string, 0), fmt.Sprintf("Number of transactions in pool: %d", memTxPoolSize)))

	rows = append(rows, append(make([]string, 0)))

	nonce := wr.getFromCacheAsUint64(core.MetricNonce)
	probableHighestNonce := wr.getFromCacheAsUint64(core.MetricProbableHighestNonce)
	rows = append(rows, append(make([]string, 0), fmt.Sprintf("Current synchronized block nonce: %v / %v",
		nonce, probableHighestNonce)))

	synchronizedRound := wr.getFromCacheAsUint64(core.MetricSynchronizedRound)
	currentRound := wr.getFromCacheAsUint64(core.MetricCurrentRound)
	rows = append(rows, append(make([]string, 0), fmt.Sprintf("Current consensus round: %v / %v",
		synchronizedRound, currentRound)))

	consensusRoundTime := wr.getFromCacheAsUint64(core.MetricRoundTime)
	rows = append(rows, append(make([]string, 0), fmt.Sprintf("Consensus round time: %ds", consensusRoundTime)))

	rows = append(rows, append(make([]string, 0)))

	numLiveValidators := wr.getFromCacheAsUint64(core.MetricLiveValidatorNodes)
	rows = append(rows, append(make([]string, 0), fmt.Sprintf("Live validator nodes: %v", numLiveValidators)))

	numConnectedNodes := wr.getFromCacheAsUint64(core.MetricConnectedNodes)
	rows = append(rows, append(make([]string, 0), fmt.Sprintf("Network connected nodes: %v", numConnectedNodes)))

	numConnectedPeers := wr.getFromCacheAsUint64(core.MetricNumConnectedPeers)
	rows = append(rows, append(make([]string, 0), fmt.Sprintf("This node is connected to %v peers", numConnectedPeers)))

	return "Chain info", rows
}

func (wr *WidgetsRender) prepareListWithLogsForDisplay(logData []string) {
	wr.lLog.Title = "Log info"
	wr.lLog.TextStyle = ui.NewStyle(ui.ColorWhite)
	wr.lLog.Rows = wr.prepareLogLines(logData, wr.lLog.Size().Y)
	wr.lLog.WrapText = true
}

//TODO duplicate code next pull request refactor
func (wr *WidgetsRender) prepareLogLines(logData []string, size int) []string {
	logDataLen := len(logData)
	if logDataLen > size {
		return logData[logDataLen-size : logDataLen]
	}

	return logData
}

func (wr *WidgetsRender) prepareLoads() {
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
	wr.memoryLoad.Label = fmt.Sprintf("%d%% / total: %s", memLoadPercent, core.ConvertBytes(memTotalMemoryBytes))

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

func (wr *WidgetsRender) getFromCacheAsUint64(metric string) uint64 {
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

func (wr *WidgetsRender) getNetworkRecvStats() {
}
