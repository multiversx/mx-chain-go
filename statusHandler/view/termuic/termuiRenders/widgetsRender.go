package termuiRenders

import (
	"fmt"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/statusHandler/view"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const statusSyncing = "currently syncing"
const statusSynchronized = "synchronized"

//WidgetsRender will define termui widgets that need to display a termui console
type WidgetsRender struct {
	container    *DrawableContainer
	lLog         *widgets.List
	instanceInfo *widgets.Table
	chainInfo    *widgets.Table
	blockInfo    *widgets.Table

	cpuLoad     *widgets.Gauge
	memoryLoad  *widgets.Gauge
	networkRecv *widgets.Gauge
	networkSent *widgets.Gauge

	presenter view.Presenter
}

//NewWidgetsRender method will create new WidgetsRender that display termui console
func NewWidgetsRender(presenter view.Presenter, grid *DrawableContainer) (*WidgetsRender, error) {
	if presenter == nil || presenter.IsInterfaceNil() {
		return nil, statusHandler.ErrorNilPresenterInterface
	}
	if grid == nil {
		return nil, statusHandler.ErrorNilGrid
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
	gridRight.Set(
		ui.NewRow(10.0/22, wr.blockInfo),
		ui.NewRow(3.0/22, wr.cpuLoad),
		ui.NewRow(3.0/22, wr.memoryLoad),
		ui.NewRow(3.0/22, wr.networkRecv),
		ui.NewRow(3.0/22, wr.networkSent))

	gridBottom := ui.NewGrid()
	gridBottom.Set(ui.NewRow(1.0, wr.lLog))

	wr.container.SetTopLeft(gridLeft)
	wr.container.SetTopRight(gridRight)
	wr.container.SetBottom(gridBottom)
}

//RefreshData method is used to prepare data that are displayed on container
func (wr *WidgetsRender) RefreshData() {
	wr.prepareInstanceInfo()
	wr.prepareChainInfo()
	wr.prepareBlockInfo()
	wr.prepareListWithLogsForDisplay()
	wr.prepareLoads()
}

func (wr *WidgetsRender) prepareInstanceInfo() {
	//8 rows and one column
	numRows := 8
	rows := make([][]string, numRows)

	nodeName := wr.presenter.GetNodeName()
	shardId := wr.presenter.GetShardId()
	instanceType := wr.presenter.GetNodeType()
	shardIdStr := fmt.Sprintf("%d", shardId)
	if shardId == uint64(sharding.MetachainShardId) {
		shardIdStr = "meta"
	}
	wr.instanceInfo.RowStyles[0] = ui.NewStyle(ui.ColorYellow)
	rows[0] = []string{fmt.Sprintf("Node name: %s (Shard %s - %s)", nodeName, shardIdStr, strings.Title(instanceType))}

	appVersion := wr.presenter.GetAppVersion()
	needUpdate, latestStableVersion := wr.presenter.CheckSoftwareVersion()
	rows[1] = []string{fmt.Sprintf("App version: %s", appVersion)}

	if needUpdate {
		wr.instanceInfo.RowStyles[1] = ui.NewStyle(ui.ColorRed, ui.ColorWhite, ui.ModifierBold)
		rows[1][0] += fmt.Sprintf(" (version %s is available)", latestStableVersion)
	} else {
		wr.instanceInfo.RowStyles[1] = ui.NewStyle(ui.ColorGreen)
	}

	pkTxSign := wr.presenter.GetPublicKeyTxSign()
	rows[2] = []string{fmt.Sprintf("Public key TxSign: %s", pkTxSign)}

	pkBlockSign := wr.presenter.GetPublicKeyBlockSign()
	rows[3] = []string{fmt.Sprintf("Public key BlockSign: %s", pkBlockSign)}

	var consensusInfo string
	countConsensus := wr.presenter.GetCountConsensus()
	countConsensusAcceptedBlocks := wr.presenter.GetCountConsensusAcceptedBlocks()

	if shardId == uint64(sharding.MetachainShardId) {
		consensusInfo = fmt.Sprintf("Count consensus participant: %d | Signed blocks headers: %d", countConsensus, countConsensusAcceptedBlocks)

	} else {
		consensusInfo = fmt.Sprintf("Consensus accepted / signed blocks: %d / %d", countConsensusAcceptedBlocks, countConsensus)
	}

	rows[4] = []string{consensusInfo}

	countLeader := wr.presenter.GetCountLeader()
	countAcceptedBlocks := wr.presenter.GetCountAcceptedBlocks()
	rows[5] = []string{fmt.Sprintf("Consensus leader accepted / proposed blocks : %d / %d", countAcceptedBlocks, countLeader)}

	switch instanceType {
	case string(core.NodeTypeValidator):
		rewardsPerHour := wr.presenter.CalculateRewardsPerHour()
		rows[6] = []string{fmt.Sprintf("Rewards estimation: %s ERD/h (without fees)", rewardsPerHour)}

		var rewardsInfo []string
		totalRewardsValue, diffRewards := wr.presenter.GetTotalRewardsValue()
		if diffRewards != "0.00" {
			wr.instanceInfo.RowStyles[7] = ui.NewStyle(ui.ColorGreen)
			rewardsInfo = []string{fmt.Sprintf("Total rewards %s + %s ERD (without fees)", totalRewardsValue, diffRewards)}
		} else {
			wr.instanceInfo.RowStyles[7] = ui.NewStyle(ui.ColorWhite)
			rewardsInfo = []string{fmt.Sprintf("Total rewards %s ERD (without fees)", totalRewardsValue)}
		}
		rows[7] = rewardsInfo

	default:
		rows[6] = []string{""}
		rows[7] = []string{""}
	}

	wr.instanceInfo.Title = "Elrond instance info"
	wr.instanceInfo.RowSeparator = false
	wr.instanceInfo.Rows = rows
}

func (wr *WidgetsRender) prepareChainInfo() {
	//10 rows and one column
	numRows := 10
	rows := make([][]string, numRows)

	syncStatus := wr.presenter.GetIsSyncing()
	syncingStr := fmt.Sprintf("undefined %d", syncStatus)

	remainingTimeMessage := ""
	blocksPerSecondMessage := ""
	switch syncStatus {
	case 1:
		syncingStr = statusSyncing

		remainingTime := wr.presenter.CalculateTimeToSynchronize()
		remainingTimeMessage = fmt.Sprintf("Synchronization time remaining: ~%s", remainingTime)

		blocksPerSecond := wr.presenter.CalculateSynchronizationSpeed()
		blocksPerSecondMessage = fmt.Sprintf("%d blocks/sec", blocksPerSecond)
	case 0:
		syncingStr = statusSynchronized
	}
	rows[0] = []string{fmt.Sprintf("Status: %s %s", syncingStr, blocksPerSecondMessage)}

	if strings.Contains(syncingStr, statusSynchronized) {
		wr.chainInfo.RowStyles[0] = ui.NewStyle(ui.ColorGreen)
	} else {
		wr.chainInfo.RowStyles[0] = ui.NewStyle(ui.ColorYellow)
	}

	rows[1] = []string{fmt.Sprintf("%s", remainingTimeMessage)}

	shardId := wr.presenter.GetShardId()
	if shardId == uint64(sharding.MetachainShardId) {
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

	nonce := wr.presenter.GetNonce()
	probableHighestNonce := wr.presenter.GetProbableHighestNonce()
	rows[4] = []string{fmt.Sprintf("Current synchronized block nonce: %d / %d",
		nonce, probableHighestNonce)}

	synchronizedRound := wr.presenter.GetSynchronizedRound()
	currentRound := wr.presenter.GetCurrentRound()
	rows[5] = []string{fmt.Sprintf("Current consensus round: %d / %d",
		synchronizedRound, currentRound)}

	consensusRoundTime := wr.presenter.GetRoundTime()
	rows[6] = []string{fmt.Sprintf("Consensus round time: %ds", consensusRoundTime)}

	numLiveValidators := wr.presenter.GetLiveValidatorNodes()
	rows[7] = []string{fmt.Sprintf("Live validator nodes: %d", numLiveValidators)}

	numConnectedNodes := wr.presenter.GetConnectedNodes()
	rows[8] = []string{fmt.Sprintf("Network connected nodes: %d", numConnectedNodes)}

	numConnectedPeers := wr.presenter.GetNumConnectedPeers()
	rows[9] = []string{fmt.Sprintf("This node is connected to %d peers", numConnectedPeers)}

	wr.chainInfo.Title = "Chain info"
	wr.chainInfo.RowSeparator = false
	wr.chainInfo.Rows = rows
}

func (wr *WidgetsRender) prepareBlockInfo() {
	//7 rows and one column
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
	rows[3] = []string{fmt.Sprintf("Current block hash : %s", currentBlockHash)}

	crossCheckBlockHeight := wr.presenter.GetCrossCheckBlockHeight()
	rows[4] = []string{fmt.Sprintf("Cross check: %s", crossCheckBlockHeight)}

	shardId := wr.presenter.GetShardId()
	if shardId != uint64(sharding.MetachainShardId) {
		highestFinalBlockInShard := wr.presenter.GetHighestFinalBlockInShard()
		rows[4][0] += fmt.Sprintf(" ,final nonce: %d", highestFinalBlockInShard)
	}

	consensusState := wr.presenter.GetConsensusState()
	rows[5] = []string{fmt.Sprintf("Consensus state: %s", consensusState)}

	syncStatus := wr.presenter.GetIsSyncing()
	switch syncStatus {
	case 1:
		rows[6] = []string{fmt.Sprintf("Consensus round state: N/A (syncing)")}
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
	rows[7] = []string{fmt.Sprintf("Current round timestamp : %d", currentRoundTimestamp)}

	wr.blockInfo.Title = "Block info"
	wr.blockInfo.RowSeparator = false
	wr.blockInfo.Rows = rows
}

func (wr *WidgetsRender) prepareListWithLogsForDisplay() {
	wr.lLog.Title = "Log info"
	wr.lLog.TextStyle = ui.NewStyle(ui.ColorWhite)

	logData := wr.presenter.GetLogLines()
	wr.lLog.Rows = wr.prepareLogLines(logData, wr.lLog.Size().Y)
	wr.lLog.WrapText = true
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

func (wr *WidgetsRender) prepareLoads() {
	cpuLoadPercent := wr.presenter.GetCpuLoadPercent()
	wr.cpuLoad.Title = "CPU load"
	wr.cpuLoad.Percent = int(cpuLoadPercent)

	memLoadPercent := wr.presenter.GetMemLoadPercent()
	memTotalMemoryBytes := wr.presenter.GetTotalMem()
	memUsed := wr.presenter.GetMemUsedByNode()
	wr.memoryLoad.Title = "Memory load"
	wr.memoryLoad.Percent = int(memLoadPercent)
	wr.memoryLoad.Label = fmt.Sprintf("%d%% / used: %s / total: %s", memLoadPercent, core.ConvertBytes(memUsed), core.ConvertBytes(memTotalMemoryBytes))

	recvLoad := wr.presenter.GetNetworkRecvPercent()
	recvBps := wr.presenter.GetNetworkRecvBps()
	recvBpsPeak := wr.presenter.GetNetworkRecvBpsPeak()
	wr.networkRecv.Title = "Network - recv:"
	wr.networkRecv.Percent = int(recvLoad)
	wr.networkRecv.Label = fmt.Sprintf("%d%% / rate: %s/s / peak rate: %s/s",
		recvLoad, core.ConvertBytes(recvBps), core.ConvertBytes(recvBpsPeak))

	sentLoad := wr.presenter.GetNetworkSentPercent()
	sentBps := wr.presenter.GetNetworkSentBps()
	sentBpsPeak := wr.presenter.GetNetworkSentBpsPeak()
	wr.networkSent.Title = "Network - sent:"
	wr.networkSent.Percent = int(sentLoad)
	wr.networkSent.Label = fmt.Sprintf("%d%% / rate: %s/s / peak rate: %s/s",
		sentLoad, core.ConvertBytes(sentBps), core.ConvertBytes(sentBpsPeak))
}

func (wr *WidgetsRender) getNetworkRecvStats() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (wr *WidgetsRender) IsInterfaceNil() bool {
	if wr == nil {
		return true
	}
	return false
}
