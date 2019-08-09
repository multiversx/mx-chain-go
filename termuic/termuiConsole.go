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

	tc := TermuiConsole{}

	tc.initMetricsMap()

	tc.viewBigLog = false
	tc.numLogLines = 10

	return &tc
}

func (tc *TermuiConsole) Write(p []byte) (n int, err error) {

	go func() {
		logLine := string(p)

		tc.mutLogLineWrite.Lock()
		if len(tc.logLines) >= tc.numLogLines {
			tc.logLines = tc.logLines[1:tc.numLogLines]
		}
		tc.mutLogLineWrite.Unlock()

		logLine = strings.Replace(logLine, "\n", "", len(logLine))

		logLine = strings.Replace(logLine, "\r", "", len(logLine))

		if logLine != "" {
			tc.mutLogLineWrite.Lock()
			tc.logLines = append(tc.logLines, logLine)
			tc.mutLogLineWrite.Unlock()
		}
	}()

	return len(p), nil
}

// InitMetricsMap will init the map of prometheus metrics
func (tc *TermuiConsole) initMetricsMap() {
	tc.TermuiConsoleMetrics = sync.Map{}

	tc.TermuiConsoleMetrics.Store(core.MetricCurrentRound, 0)
	tc.TermuiConsoleMetrics.Store(core.MetricIsSyncing, 0)
	tc.TermuiConsoleMetrics.Store(core.MetricNonce, 0)
	tc.TermuiConsoleMetrics.Store(core.MetricNumConnectedPeers, 0)
	tc.TermuiConsoleMetrics.Store(core.MetricSynchronizedRound, 0)

	tc.TermuiConsoleMetrics.Store(core.MetricPublicKey, "")
	tc.TermuiConsoleMetrics.Store(core.MetricShardId, 0)
	tc.TermuiConsoleMetrics.Store(core.MetricTxPoolLoad, 0)

	tc.TermuiConsoleMetrics.Store(core.MetricCountConsensus, 0)
	tc.TermuiConsoleMetrics.Store(core.MetricCountLeader, 0)
	tc.TermuiConsoleMetrics.Store(core.MetricCountAcceptedBlocks, 0)
}

// SetInt64Value method - will update the value for a key
func (tc *TermuiConsole) SetInt64Value(key string, value int64) {
	if _, ok := tc.TermuiConsoleMetrics.Load(key); ok {
		tc.TermuiConsoleMetrics.Store(key, int(value))
	}
}

// SetUInt64Value method - will update the value for a key
func (tc *TermuiConsole) SetUInt64Value(key string, value uint64) {
	if _, ok := tc.TermuiConsoleMetrics.Load(key); ok {
		tc.TermuiConsoleMetrics.Store(key, int(value))
	}
}

// SetStringValue method - will update the value of a key
func (tc *TermuiConsole) SetStringValue(key string, value string) {
	if _, ok := tc.TermuiConsoleMetrics.Load(key); ok {
		tc.TermuiConsoleMetrics.Store(key, value)
	}
}

// Increment - will increment the value of a key
func (tc *TermuiConsole) Increment(key string) {
	if keyValueI, ok := tc.TermuiConsoleMetrics.Load(key); ok {

		keyValue := keyValueI.(int)
		keyValue++
		tc.TermuiConsoleMetrics.Store(key, keyValue)
	}
}

// Decrement method - will decrement the value for a key
func (tc *TermuiConsole) Decrement(key string) {
	return
}

// Close method - won't do anything
func (tc *TermuiConsole) Close() {
	return
}

// Start method - will start termui console
func (tc *TermuiConsole) Start() {
	go func() {
		if err := ui.Init(); err != nil {

		}
		defer ui.Close()

		tc.eventLoop()
	}()

}

func (tc *TermuiConsole) eventLoop() {

	tc.termuiConsoleWidgets = NewtermuiConsoleGrid()

	time.Sleep(1 * time.Second)

	grid := tc.configConsoleWithNormalTermuiWidgets()

	uiEvents := ui.PollEvents()
	// handles kill signal sent to gotop
	sigTerm := make(chan os.Signal, 2)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)

	drawTicker := time.NewTicker(time.Second).C

	for {
		select {
		case <-drawTicker:
			tc.refreshDataForConsole()

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
				grid = tc.changeConsoleDisplay()

			case "<C-c>":
				ui.Clear()
				_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)

				return
			}
		}
	}
}

func (tc *TermuiConsole) refreshDataForConsole() {
	if tc.viewBigLog == false {
		tc.prepareDataNormalWidgets()
	} else {
		tc.prepareDataBigLog()
	}
}

func (tc *TermuiConsole) prepareDataNormalWidgets() {
	nonceI, _ := tc.TermuiConsoleMetrics.Load(core.MetricNonce)
	nonce := nonceI.(int)
	tc.termuiConsoleWidgets.PrepareNonceForDisplay(nonce)

	synchronizedRoundI, _ := tc.TermuiConsoleMetrics.Load(core.MetricSynchronizedRound)
	synchronizedRound := synchronizedRoundI.(int)
	tc.termuiConsoleWidgets.PrepareSynchronizedRoundForDisplay(synchronizedRound)

	currentRoundI, _ := tc.TermuiConsoleMetrics.Load(core.MetricCurrentRound)
	currentRound := currentRoundI.(int)
	tc.termuiConsoleWidgets.PrepareCurrentRoundForDisplay(currentRound)

	isSyncingI, _ := tc.TermuiConsoleMetrics.Load(core.MetricIsSyncing)
	isSyncing := isSyncingI.(int)
	tc.termuiConsoleWidgets.PrepareIsSyncingForDisplay(isSyncing)

	publicKeyI, _ := tc.TermuiConsoleMetrics.Load(core.MetricPublicKey)
	publicKey := publicKeyI.(string)
	tc.termuiConsoleWidgets.PreparePublicKeyForDisplay(publicKey)

	shardIdI, _ := tc.TermuiConsoleMetrics.Load(core.MetricShardId)
	shardId := shardIdI.(int)
	tc.termuiConsoleWidgets.PrepareShardIdForDisplay(shardId)

	numConnectedPeersI, _ := tc.TermuiConsoleMetrics.Load(core.MetricNumConnectedPeers)
	numConnectedPeers := numConnectedPeersI.(int)

	txPoolLoadI, _ := tc.TermuiConsoleMetrics.Load(core.MetricTxPoolLoad)
	txPoolLoad := txPoolLoadI.(int)
	tc.termuiConsoleWidgets.PrepareSparkLineGroupForDisplay(txPoolLoad, numConnectedPeers)

	countConsensusI, _ := tc.TermuiConsoleMetrics.Load(core.MetricCountConsensus)
	countConsensus := countConsensusI.(int)

	countLeaderI, _ := tc.TermuiConsoleMetrics.Load(core.MetricCountLeader)
	countLeader := countLeaderI.(int)

	countAcceptedBlocksI, _ := tc.TermuiConsoleMetrics.Load(core.MetricCountAcceptedBlocks)
	countAcceptedBlocks := countAcceptedBlocksI.(int)

	tc.termuiConsoleWidgets.PrepareConcensusInformationsForDisplay(countConsensus, countLeader, countAcceptedBlocks)

	tc.termuiConsoleWidgets.PrepareListWithLogsForDisplay(tc.logLines)
}

func (tc *TermuiConsole) prepareDataBigLog() {
	tc.termuiConsoleWidgets.PrepareListWithLogsForDisplay(tc.logLines)
}

func (tc *TermuiConsole) changeConsoleDisplay() *ui.Grid {
	if tc.viewBigLog == false {
		tc.viewBigLog = true
		tc.numLogLines = 50

		return tc.configConsoleWithBigLog()

	} else {
		tc.viewBigLog = false
		tc.numLogLines = 10

		return tc.configConsoleWithNormalTermuiWidgets()
	}
}

func (tc *TermuiConsole) configConsoleWithBigLog() *ui.Grid {

	tc.termuiConsoleWidgets.SetupBigLogGrid()

	grid := tc.termuiConsoleWidgets.Grid()

	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	return grid
}

func (tc *TermuiConsole) configConsoleWithNormalTermuiWidgets() *ui.Grid {
	tc.termuiConsoleWidgets.setupGrid()

	grid := tc.termuiConsoleWidgets.Grid()

	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	return grid
}
