package termuiRenders

import (
	"fmt"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/cmd/termui/view"
	"github.com/multiversx/mx-chain-go/common"
)

const (
	statusSyncing       = "currently syncing"
	statusSynchronized  = "synchronized"
	statusNotApplicable = "N/A"
	invalidKey          = "invalid key"
)

// WidgetsRender will define termui widgets that need to display a termui console
type WidgetsRender struct {
	container    *DrawableContainer
	lLog         *widgets.List
	instanceInfo *widgets.Table
	chainInfo    *widgets.Table
	blockInfo    *widgets.Table

	epochLoad   *widgets.Gauge
	cpuLoad     *widgets.Gauge
	memoryLoad  *widgets.Gauge
	networkRecv *widgets.Gauge
	networkSent *widgets.Gauge

	networkBytesInEpoch *widgets.Gauge

	presenter view.Presenter
}

// NewWidgetsRender method will create new WidgetsRender that display termui console
func NewWidgetsRender(presenter view.Presenter, grid *DrawableContainer) (*WidgetsRender, error) {
	if presenter == nil || presenter.IsInterfaceNil() {
		return nil, view.ErrNilPresenterInterface
	}
	if grid == nil {
		return nil, view.ErrNilGrid
	}

	self := &WidgetsRender{
		presenter: presenter,
		container: grid,
	}
	self.initWidgets()
	self.setGrid()

	return self, nil
}

func (wr *WidgetsRender) initWidgets() {
	wr.instanceInfo = widgets.NewTable()
	wr.instanceInfo.Rows = [][]string{{""}}

	wr.chainInfo = widgets.NewTable()
	wr.chainInfo.Rows = [][]string{{"", "", "", ""}}

	wr.blockInfo = widgets.NewTable()
	wr.blockInfo.Rows = [][]string{{"", "", ""}}

	wr.epochLoad = widgets.NewGauge()
	wr.cpuLoad = widgets.NewGauge()
	wr.memoryLoad = widgets.NewGauge()
	wr.networkRecv = widgets.NewGauge()
	wr.networkSent = widgets.NewGauge()
	wr.networkBytesInEpoch = widgets.NewGauge()

	wr.lLog = widgets.NewList()
}

func (wr *WidgetsRender) setGrid() {
	gridLeft := ui.NewGrid()

	gridLeft.Set(
		ui.NewRow(10.0/22, wr.instanceInfo),
		ui.NewRow(12.0/22, wr.chainInfo))

	colNetworkRecv := ui.NewCol(1.0/2, wr.networkRecv)
	colNetworkSent := ui.NewCol(1.0/2, wr.networkSent)

	colCpuLoad := ui.NewCol(1.0/2, wr.cpuLoad)
	colMemoryLoad := ui.NewCol(1.0/2, wr.memoryLoad)

	gridRight := ui.NewGrid()
	gridRight.Set(
		ui.NewRow(10.0/22, wr.blockInfo),
		ui.NewRow(3.0/22, colCpuLoad, colMemoryLoad),
		ui.NewRow(3.0/22, wr.epochLoad),
		ui.NewRow(3.0/22, wr.networkBytesInEpoch),
		ui.NewRow(3.0/22, colNetworkSent, colNetworkRecv),
	)

	gridBottom := ui.NewGrid()
	gridBottom.Set(ui.NewRow(1.0, wr.lLog))

	wr.container.SetTopLeft(gridLeft)
	wr.container.SetTopRight(gridRight)
	wr.container.SetBottom(gridBottom)
}

// RefreshData method is used to prepare data that are displayed on container
func (wr *WidgetsRender) RefreshData(numMillisecondsRefreshTime int) {
	wr.prepareInstanceInfo()
	wr.prepareChainInfo(numMillisecondsRefreshTime)
	wr.prepareBlockInfo()
	wr.prepareListWithLogsForDisplay()
	wr.prepareLoads()
}

func (wr *WidgetsRender) prepareInstanceInfo() {
	// 8 rows and one column
	numRows := 8
	rows := make([][]string, numRows)

	nodeName := wr.presenter.GetNodeName()
	shardId := wr.presenter.GetShardId()
	instanceType := wr.presenter.GetNodeType()
	peerType := wr.presenter.GetPeerType()
	peerSubType := wr.presenter.GetPeerSubType()
	chainID := wr.presenter.GetChainID()

	nodeTypeAndListDisplay := instanceType
	if peerType != string(common.ObserverList) && !strings.Contains(peerType, invalidKey) {
		nodeTypeAndListDisplay += fmt.Sprintf(" - %s", peerType)
	}
	if peerSubType == core.FullHistoryObserver.String() {
		nodeTypeAndListDisplay += " - full archive"
	}
	shardIdStr := fmt.Sprintf("%d", shardId)
	if shardId == uint64(core.MetachainShardId) {
		shardIdStr = "meta"
	}
	wr.instanceInfo.RowStyles[0] = ui.NewStyle(ui.ColorYellow)
	rows[0] = []string{
		fmt.Sprintf("Node name: %s (Shard %s - %s)",
			nodeName,
			shardIdStr,
			nodeTypeAndListDisplay,
		),
	}

	appVersion := wr.presenter.GetAppVersion()
	needUpdate, latestStableVersion := wr.presenter.CheckSoftwareVersion()
	rows[1] = []string{fmt.Sprintf("App version: %s", appVersion)}

	if needUpdate {
		wr.instanceInfo.RowStyles[1] = ui.NewStyle(ui.ColorRed, ui.ColorClear, ui.ModifierBold)
		rows[1][0] += fmt.Sprintf(" (version %s is available)", latestStableVersion)
	} else {
		wr.instanceInfo.RowStyles[1] = ui.NewStyle(ui.ColorGreen)
	}

	pkBlockSign := wr.presenter.GetPublicKeyBlockSign()
	rows[2] = []string{fmt.Sprintf("Public key BlockSign: %s", pkBlockSign)}

	countConsensus := wr.presenter.GetCountConsensus()
	countConsensusAcceptedBlocks := wr.presenter.GetCountConsensusAcceptedBlocks()

	rows[3] = []string{fmt.Sprintf("Validator signed blocks: %d | Blocks accepted: %d", countConsensus, countConsensusAcceptedBlocks)}

	countLeader := wr.presenter.GetCountLeader()
	countAcceptedBlocks := wr.presenter.GetCountAcceptedBlocks()
	rows[4] = []string{fmt.Sprintf("Blocks proposed: %d | Blocks accepted:  %d", countLeader, countAcceptedBlocks)}

	rows[5] = []string{computeRedundancyStr(wr.presenter.GetRedundancyLevel(), wr.presenter.GetRedundancyIsMainActive())}
	rows[6] = []string{fmt.Sprintf("Chain ID: %s", chainID)}

	wr.instanceInfo.Title = "MultiversX instance info:"
	wr.instanceInfo.RowSeparator = false
	wr.instanceInfo.Rows = rows
}

func (wr *WidgetsRender) getTrieSyncProgress() string {
	syncPercentageOut := statusNotApplicable

	syncPercentage, ok := wr.presenter.GetTrieSyncProcessedPercentage()
	if ok {
		syncPercentageOut = "~" + fmt.Sprint(syncPercentage) + "%"
	}

	return syncPercentageOut
}

func (wr *WidgetsRender) prepareChainInfo(numMillisecondsRefreshTime int) {
	// 10 rows and one column
	numRows := 10
	rows := make([][]string, numRows)

	synchronizedRound := wr.presenter.GetSynchronizedRound()
	currentRound := wr.presenter.GetCurrentRound()

	nodesProcessed := wr.presenter.GetTrieSyncNumProcessedNodes()
	isNodeSyncingTrie := nodesProcessed != 0

	var syncingStr, statusMessage, blocksPerSecondMessage string
	switch {
	case isNodeSyncingTrie:
		syncingStr = statusSyncing
		bytesReceived := wr.presenter.GetTrieSyncNumBytesReceived()
		syncPercentageOut := wr.getTrieSyncProgress()

		statusMessage = fmt.Sprintf("Trie sync: %d nodes, progress %s, %s state size", nodesProcessed, syncPercentageOut, core.ConvertBytes(bytesReceived))
	case synchronizedRound < currentRound:
		syncingStr = statusSyncing

		remainingTime := wr.presenter.CalculateTimeToSynchronize(numMillisecondsRefreshTime)
		statusMessage = fmt.Sprintf("Synchronization time remaining: ~%s", remainingTime)

		blocksPerSecond := wr.presenter.CalculateSynchronizationSpeed(numMillisecondsRefreshTime)
		blocksPerSecondMessage = fmt.Sprintf("%d blocks/sec", blocksPerSecond)
	case currentRound == 0:
		syncingStr = statusNotApplicable
	default:
		syncingStr = statusSynchronized
	}
	rows[0] = []string{fmt.Sprintf("Status: %s %s", syncingStr, blocksPerSecondMessage)}

	if strings.Contains(syncingStr, statusSynchronized) {
		wr.chainInfo.RowStyles[0] = ui.NewStyle(ui.ColorGreen)
	} else {
		wr.chainInfo.RowStyles[0] = ui.NewStyle(ui.ColorYellow)
		wr.chainInfo.RowStyles[1] = ui.NewStyle(ui.ColorYellow)
	}

	rows[1] = []string{statusMessage}

	shardId := wr.presenter.GetShardId()
	if shardId == uint64(core.MetachainShardId) {
		numShardHeadersInPool := wr.presenter.GetNumShardHeadersInPool()
		rows[2] = []string{fmt.Sprintf("Number of shard headers in pool: %d", numShardHeadersInPool)}
		numShardHeaderProcessed := wr.presenter.GetNumShardHeadersProcessed()
		rows[3] = []string{fmt.Sprintf("Number of shard headers processed: %d", numShardHeaderProcessed)}
	} else {
		memTxPoolSize := wr.presenter.GetTxPoolLoad()
		rows[2] = []string{fmt.Sprintf("Number of transactions in pool: %d", memTxPoolSize)}

		numTxProcessed := wr.presenter.GetNumTxProcessed()
		rows[3] = []string{fmt.Sprintf("Number of transactions processed: %d", numTxProcessed)}
	}

	epoch := wr.presenter.GetEpochNumber()
	rows[4] = []string{fmt.Sprintf("Current epoch: %d", epoch)}

	nonce := wr.presenter.GetNonce()
	probableHighestNonce := wr.presenter.GetProbableHighestNonce()
	rows[5] = []string{fmt.Sprintf("Current synchronized block nonce: %d / %d",
		nonce, probableHighestNonce)}

	rows[6] = []string{fmt.Sprintf("Current consensus round: %d / %d",
		synchronizedRound, currentRound)}

	consensusRoundTime := wr.presenter.GetRoundTime()
	rows[7] = []string{fmt.Sprintf("Consensus round time: %ds", consensusRoundTime)}

	numConnectedPeers := wr.presenter.GetNumConnectedPeers()
	numLiveValidators := wr.presenter.GetLiveValidatorNodes()
	numConnectedNodes := wr.presenter.GetConnectedNodes()
	numIntraShardValidators := wr.presenter.GetIntraShardValidators()
	rows[8] = []string{fmt.Sprintf("Intra shard peers / validators / nodes: %d / %d / %d",
		numConnectedPeers, numIntraShardValidators, numConnectedNodes)}
	rows[9] = []string{fmt.Sprintf("All known validators: %d", numLiveValidators)}

	wr.chainInfo.Title = "Chain info:"
	wr.chainInfo.RowSeparator = false
	wr.chainInfo.Rows = rows
}

func computeRedundancyStr(redundancyLevel int64, redundancyIsMainActive string) string {
	if redundancyIsMainActive == statusNotApplicable {
		return ""
	}

	redundancyStr := "Redundancy: "
	if redundancyLevel < 0 {
		redundancyStr += "inactive"
	} else {
		if redundancyLevel == 0 {
			redundancyStr += "main machine"
		} else {
			redundancyStr += fmt.Sprintf("back-up #%d", redundancyLevel)
			redundancyStr += fmt.Sprintf(" (is main active: %s)", redundancyIsMainActive)
		}
	}

	return redundancyStr
}

func (wr *WidgetsRender) prepareBlockInfo() {
	// 7 rows and one column
	numRows := 8
	rows := make([][]string, numRows)

	currentBlockHeight := wr.presenter.GetNonce()
	blockSize := wr.presenter.GetBlockSize()
	rows[0] = []string{fmt.Sprintf("Current block height: %d, size: %s", currentBlockHeight, core.ConvertBytes(blockSize))}

	numTransactionInBlock := wr.presenter.GetNumTxInBlock()
	rows[1] = []string{fmt.Sprintf("Num transactions in block: %d", numTransactionInBlock)}

	numMiniBlocks := wr.presenter.GetNumMiniBlocks()
	rows[2] = []string{fmt.Sprintf("Num miniblocks in block: %d", numMiniBlocks)}

	currentBlockHash := wr.presenter.GetCurrentBlockHash()
	rows[3] = []string{fmt.Sprintf("Current block hash: %s", currentBlockHash)}

	crossCheckBlockHeight := wr.presenter.GetCrossCheckBlockHeight()
	rows[4] = []string{fmt.Sprintf("Cross check: %s", crossCheckBlockHeight)}

	shardId := wr.presenter.GetShardId()
	if shardId != uint64(core.MetachainShardId) {
		highestFinalBlock := wr.presenter.GetHighestFinalBlock()
		rows[4][0] += fmt.Sprintf(", final nonce: %d", highestFinalBlock)
	}

	consensusState := wr.presenter.GetConsensusState()
	rows[5] = []string{fmt.Sprintf("Consensus state: %s", consensusState)}

	syncStatus := wr.presenter.GetIsSyncing()
	switch syncStatus {
	case 1:
		rows[6] = []string{"Consensus round state: N/A (syncing)"}
	case 0:
		instanceType := wr.presenter.GetNodeType()
		if instanceType == string(core.NodeTypeObserver) {
			rows[6] = []string{fmt.Sprintf("Consensus round state: N/A (%s)", string(core.NodeTypeObserver))}
		} else {
			consensusRoundState := wr.presenter.GetConsensusRoundState()
			rows[6] = []string{fmt.Sprintf("Consensus round state: %s", consensusRoundState)}
		}
	}

	currentRoundTimestamp := wr.presenter.GetCurrentRoundTimestamp()
	rows[7] = []string{fmt.Sprintf("Current round timestamp: %d", currentRoundTimestamp)}

	wr.blockInfo.Title = "Block info:"
	wr.blockInfo.RowSeparator = false
	wr.blockInfo.Rows = rows
}

func (wr *WidgetsRender) prepareListWithLogsForDisplay() {
	wr.lLog.Title = "Log info:"
	wr.lLog.TextStyle = ui.NewStyle(ui.ColorWhite)

	logData := wr.presenter.GetLogLines()
	wr.lLog.Rows = wr.prepareLogLines(logData, wr.lLog.Size().Y)
	wr.lLog.WrapText = false
}

func (wr *WidgetsRender) prepareLogLines(logData []string, size int) []string {
	logDataLen := len(logData)
	maxSize := size - 2 // decrease 2 units as the total size of the log list includes also the header and the footer
	if maxSize <= 0 {
		return []string{} // there isn't place for any log line
	}

	if logDataLen > maxSize {
		return logData[(logDataLen - maxSize):]
	}

	return logData
}

func fitStringToWidth(original string, maxWidth int) string {
	suffixString := "..."
	numExtraPadding := 2

	if len(original)+numExtraPadding < maxWidth {
		return original
	}

	nothingToShow := maxWidth <= len(suffixString)+numExtraPadding ||
		len(original)-len(suffixString)-numExtraPadding < 0
	if nothingToShow {
		return ""
	}

	return original[:maxWidth-len(suffixString)-numExtraPadding] + suffixString
}

func (wr *WidgetsRender) prepareLoads() {
	cpuLoadPercent := wr.presenter.GetCpuLoadPercent()
	wr.cpuLoad.Title = "CPU load:"
	wr.cpuLoad.Percent = int(cpuLoadPercent)

	memLoadPercent := wr.presenter.GetMemLoadPercent()
	memTotalMemoryBytes := wr.presenter.GetTotalMem()
	memUsed := wr.presenter.GetMemUsedByNode()
	wr.memoryLoad.Title = "Memory load:"
	wr.memoryLoad.Percent = int(memLoadPercent)
	str := fmt.Sprintf("%d%% / used: %s / total: %s", memLoadPercent, core.ConvertBytes(memUsed), core.ConvertBytes(memTotalMemoryBytes))
	wr.memoryLoad.Label = fitStringToWidth(str, wr.memoryLoad.Size().X)

	recvLoad := wr.presenter.GetNetworkRecvPercent()
	recvBps := wr.presenter.GetNetworkRecvBps()
	recvBpsPeak := wr.presenter.GetNetworkRecvBpsPeak()
	wr.networkRecv.Title = "Network - received per host:"
	wr.networkRecv.Percent = int(recvLoad)
	str = fmt.Sprintf("%d%% / current: %s/s / peak: %s/s", recvLoad, core.ConvertBytes(recvBps), core.ConvertBytes(recvBpsPeak))
	wr.networkRecv.Label = fitStringToWidth(str, wr.networkRecv.Size().X)

	sentLoad := wr.presenter.GetNetworkSentPercent()
	sentBps := wr.presenter.GetNetworkSentBps()
	sentBpsPeak := wr.presenter.GetNetworkSentBpsPeak()
	wr.networkSent.Title = "Network - sent per host:"
	wr.networkSent.Percent = int(sentLoad)
	str = fmt.Sprintf("%d%% / current: %s/s / peak: %s/s", sentLoad, core.ConvertBytes(sentBps), core.ConvertBytes(sentBpsPeak))
	wr.networkSent.Label = fitStringToWidth(str, wr.networkSent.Size().X)

	// epoch info
	currentEpochRound, currentEpochFinishRound, epochLoadPercent, remainingTime := wr.presenter.GetEpochInfo()
	wr.epochLoad.Title = "Epoch - info:"
	wr.epochLoad.Percent = epochLoadPercent
	if len(remainingTime) > 0 {
		remainingTime = fmt.Sprintf(" (~%sremaining)", remainingTime)
	}

	str = fmt.Sprintf("%d / %d rounds%s", currentEpochRound, currentEpochFinishRound, remainingTime)
	wr.epochLoad.Label = fitStringToWidth(str, wr.epochLoad.Size().X)

	totalBytesSentInEpoch := wr.presenter.GetNetworkSentBytesInEpoch()
	totalBytesReceivedInEpoch := wr.presenter.GetNetworkReceivedBytesInEpoch()

	wr.networkBytesInEpoch.Title = "Epoch - traffic per host:"
	wr.networkBytesInEpoch.Percent = 0
	str = fmt.Sprintf("sent: %s / received: %s", core.ConvertBytes(totalBytesSentInEpoch), core.ConvertBytes(totalBytesReceivedInEpoch))
	wr.networkBytesInEpoch.Label = fitStringToWidth(str, wr.networkBytesInEpoch.Size().X)
}

// IsInterfaceNil returns true if there is no value under the interface
func (wr *WidgetsRender) IsInterfaceNil() bool {
	return wr == nil
}
