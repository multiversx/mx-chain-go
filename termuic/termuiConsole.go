package termuic

import (
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	ui "github.com/gizak/termui/v3"
)

// TermuiConsole data where is store data from handler
type TermuiConsole struct {
	TermuiConsoleMetrics sync.Map
	logLines             []string
	mutLogLineWrite      sync.RWMutex
	viewBigLog           bool
	numLogLines          int
	termuiConsoleWidgets *termuiConsoleGrid
}

//NewTermuiConsole method is used to return a new TermuiConsole structure
func NewTermuiConsole() *TermuiConsole {

	tdm := TermuiConsole{}

	tdm.initMetricsMap()

	tdm.viewBigLog = false
	tdm.numLogLines = 10

	return &tdm
}

func (tdm *TermuiConsole) Write(p []byte) (n int, err error) {

	go func() {
		logLine := string(p)

		tdm.mutLogLineWrite.Lock()
		if len(tdm.logLines) >= tdm.numLogLines {
			tdm.logLines = tdm.logLines[1:tdm.numLogLines]
		}
		tdm.mutLogLineWrite.Unlock()

		logLine = strings.Replace(logLine, "\n", "", len(logLine))

		logLine = strings.Replace(logLine, "\r", "", len(logLine))

		if logLine != "" {
			tdm.mutLogLineWrite.Lock()
			tdm.logLines = append(tdm.logLines, logLine)
			tdm.mutLogLineWrite.Unlock()
		}
	}()

	return len(p), nil
}

// InitMetricsMap will init the map of prometheus metrics
func (tdm *TermuiConsole) initMetricsMap() {
	tdm.TermuiConsoleMetrics = sync.Map{}

	tdm.TermuiConsoleMetrics.Store(core.MetricCurrentRound, 0)
	tdm.TermuiConsoleMetrics.Store(core.MetricIsSyncing, 0)
	tdm.TermuiConsoleMetrics.Store(core.MetricNonce, 0)
	tdm.TermuiConsoleMetrics.Store(core.MetricNumConnectedPeers, 0)
	tdm.TermuiConsoleMetrics.Store(core.MetricSynchronizedRound, 0)

	tdm.TermuiConsoleMetrics.Store(core.MetricPublicKey, "")
	tdm.TermuiConsoleMetrics.Store(core.MetricShardId, 0)
	tdm.TermuiConsoleMetrics.Store(core.MetricTxPoolLoad, 0)

	tdm.TermuiConsoleMetrics.Store(core.MetricCountConsensus, 0)
	tdm.TermuiConsoleMetrics.Store(core.MetricCountLeader, 0)
	tdm.TermuiConsoleMetrics.Store(core.MetricCountAcceptedBlocks, 0)
}

// SetInt64Value method - will update the value for a key
func (tdm *TermuiConsole) SetInt64Value(key string, value int64) {
	if _, ok := tdm.TermuiConsoleMetrics.Load(key); ok {
		tdm.TermuiConsoleMetrics.Store(key, int(value))
	}
}

// SetUInt64Value method - will update the value for a key
func (tdm *TermuiConsole) SetUInt64Value(key string, value uint64) {
	if _, ok := tdm.TermuiConsoleMetrics.Load(key); ok {
		tdm.TermuiConsoleMetrics.Store(key, int(value))
	}
}

// SetStringValue method - will update the value of a key
func (tdm *TermuiConsole) SetStringValue(key string, value string) {
	if _, ok := tdm.TermuiConsoleMetrics.Load(key); ok {
		tdm.TermuiConsoleMetrics.Store(key, value)
	}
}

// Increment - will increment the value of a key
func (tdm *TermuiConsole) Increment(key string) {
	if keyValueI, ok := tdm.TermuiConsoleMetrics.Load(key); ok {

		keyValue := keyValueI.(int)
		keyValue++
		tdm.TermuiConsoleMetrics.Store(key, keyValue)
	}
}

// Decrement method - will decrement the value for a key
func (tsh *TermuiConsole) Decrement(key string) {
	return
}

// Close method - won't do anything
func (tdm *TermuiConsole) Close() {
	return
}

// Start method - will start termui console
func (tdm *TermuiConsole) Start() {
	go func() {
		if err := ui.Init(); err != nil {

		}
		defer ui.Close()

		tdm.eventLoop()
	}()

}

func (tdm *TermuiConsole) eventLoop() {

	tdm.termuiConsoleWidgets = NewtermuiConsoleGrid()

	time.Sleep(1 * time.Second)

	grid := tdm.configConsoleWithNormalTermuiWidgets()

	uiEvents := ui.PollEvents()
	// handles kill signal sent to gotop
	sigTerm := make(chan os.Signal, 2)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)

	drawTicker := time.NewTicker(time.Second).C

	for {
		select {
		case <-drawTicker:
			tdm.refreshDataForConsole()

			ui.Clear()
			ui.Render(grid)
		case <-sigTerm:
			ui.Clear()

			return
		case e := <-uiEvents:
			switch e.ID {
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				grid.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
				ui.Render(grid)
			case "q":
				ui.Clear()
				grid = tdm.changeConsoleDisplay()

			case "<C-c>":
				ui.Clear()
				_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)

				return
			}
		}
	}
}

func (tdm *TermuiConsole) refreshDataForConsole() {
	if tdm.viewBigLog == false {
		tdm.prepareDataNormalWidgets()
	} else {
		tdm.prepareDataBigLog()
	}
}

func (tdm *TermuiConsole) prepareDataNormalWidgets() {
	nonceI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricNonce)
	nonce := nonceI.(int)
	tdm.termuiConsoleWidgets.PrepareNonceForDisplay(nonce)

	synchronizedRoundI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricSynchronizedRound)
	synchronizedRound := synchronizedRoundI.(int)
	tdm.termuiConsoleWidgets.PrepareSynchronizedRoundForDisplay(synchronizedRound)

	currentRoundI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricCurrentRound)
	currentRound := currentRoundI.(int)
	tdm.termuiConsoleWidgets.PrepareCurrentRoundForDisplay(currentRound)

	isSyncingI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricIsSyncing)
	isSyncing := isSyncingI.(int)
	tdm.termuiConsoleWidgets.PrepareIsSyncingForDisplay(isSyncing)

	publicKeyI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricPublicKey)
	publicKey := publicKeyI.(string)
	tdm.termuiConsoleWidgets.PreparePublicKeyForDisplay(publicKey)

	shardIdI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricShardId)
	shardId := shardIdI.(int)
	tdm.termuiConsoleWidgets.PrepareShardIdForDisplay(shardId)

	numConnectedPeersI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricNumConnectedPeers)
	numConnectedPeers := numConnectedPeersI.(int)

	txPoolLoadI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricTxPoolLoad)
	txPoolLoad := txPoolLoadI.(int)
	tdm.termuiConsoleWidgets.PrepareSparkLineGroupForDisplay(txPoolLoad, numConnectedPeers)

	countConsensusI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricCountConsensus)
	countConsensus := countConsensusI.(int)

	countLeaderI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricCountLeader)
	countLeader := countLeaderI.(int)

	countAcceptedBlocksI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricCountAcceptedBlocks)
	countAcceptedBlocks := countAcceptedBlocksI.(int)

	tdm.termuiConsoleWidgets.PrepareConcensusInformationsForDisplay(countConsensus, countLeader, countAcceptedBlocks)

	tdm.termuiConsoleWidgets.PrepareListWithLogsForDisplay(tdm.logLines)
}

func (tdm *TermuiConsole) prepareDataBigLog() {
	tdm.termuiConsoleWidgets.PrepareListWithLogsForDisplay(tdm.logLines)
}

func (tdm *TermuiConsole) changeConsoleDisplay() *ui.Grid {
	if tdm.viewBigLog == false {
		tdm.viewBigLog = true
		tdm.numLogLines = 50

		return tdm.configConsoleWithBigLog()

	} else {
		tdm.viewBigLog = false
		tdm.numLogLines = 10

		return tdm.configConsoleWithNormalTermuiWidgets()
	}
}

func (tdm *TermuiConsole) configConsoleWithBigLog() *ui.Grid {
	grid := ui.NewGrid()
	grid.Set(
		ui.NewRow(1.0, tdm.termuiConsoleWidgets.lLog),
	)

	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	return grid
}

func (tdm *TermuiConsole) configConsoleWithNormalTermuiWidgets() *ui.Grid {
	tdm.termuiConsoleWidgets.setupGrid()

	grid := tdm.termuiConsoleWidgets.Grid()

	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	return grid
}
