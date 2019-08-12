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

//refreshInterval is used for a ticker that refresh termui console at a specific interval
const refreshInterval = time.Second

// TermuiConsole data where is store data from handler
type TermuiConsole struct {
	rermuiConsoleMetrics *sync.Map
	logLines             []string
	mutLogLineWrite      sync.RWMutex
	viewBigLog           bool
	numLogLines          int
	termuiConsoleWidgets *termuiConsoleGrid
	grid                 *ui.Grid
}

//NewTermuiConsole method is used to return a new TermuiConsole structure
func NewTermuiConsole(metricData *sync.Map) *TermuiConsole {

	tc := TermuiConsole{
		rermuiConsoleMetrics: metricData,
		viewBigLog:           false,
		numLogLines:          10,
	}

	return &tc
}

func (tc *TermuiConsole) Write(p []byte) (n int, err error) {

	go func() {
		logLine := string(p)
		logLine = strings.Replace(logLine, "\n", "", len(logLine))
		logLine = strings.Replace(logLine, "\r", "", len(logLine))

		tc.mutLogLineWrite.Lock()
		if len(tc.logLines) >= tc.numLogLines {
			tc.logLines = tc.logLines[1:tc.numLogLines]
		}
		if logLine != "" {
			tc.logLines = append(tc.logLines, logLine)
		}
		tc.mutLogLineWrite.Unlock()
	}()

	return len(p), nil
}

// Start method - will start termui console
func (tc *TermuiConsole) Start() error {

	if err := ui.Init(); err != nil {
		return err
	}

	go func() {
		defer ui.Close()

		tc.eventLoop()
	}()

	return nil
}

func (tc *TermuiConsole) eventLoop() {

	tc.termuiConsoleWidgets = NewtermuiConsoleGrid()

	time.Sleep(1 * time.Second)

	tc.grid = tc.configConsoleWithNormalTermuiWidgets()

	uiEvents := ui.PollEvents()
	// handles kill signal sent to gotop
	sigTerm := make(chan os.Signal, 2)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)

	drawTicker := time.NewTicker(refreshInterval).C

	for {
		select {
		case <-drawTicker:
			tc.refreshDataForConsole()

			ui.Clear()
			ui.Render(tc.grid)
		case <-sigTerm:
			ui.Clear()

			return
		case e := <-uiEvents:
			tc.processUiEvents(e)
		}
	}
}

func (tc *TermuiConsole) processUiEvents(e ui.Event) {
	switch e.ID {
	case "<Resize>":
		payload := e.Payload.(ui.Resize)
		tc.grid.SetRect(0, 0, payload.Width, payload.Height)
		ui.Clear()
		ui.Render(tc.grid)
	case "q":
		ui.Clear()
		tc.grid = tc.changeConsoleDisplay()
		ui.Render(tc.grid)

	case "<C-c>":
		ui.Clear()
		if p, err := os.FindProcess(os.Getpid()); err != nil {
			//Error check
			//TODO
		} else {
			p.Signal(syscall.SIGINT)
		}
		return
	}
}

func (tc *TermuiConsole) refreshDataForConsole() {
	if tc.viewBigLog == false {
		tc.prepareDataNormalWidgets()
	} else {
		tc.termuiConsoleWidgets.PrepareListWithLogsForDisplay(tc.logLines)
	}
}

func (tc *TermuiConsole) prepareDataNormalWidgets() {
	nonceI, _ := tc.rermuiConsoleMetrics.Load(core.MetricNonce)
	nonce := nonceI.(int)

	synchronizedRoundI, _ := tc.rermuiConsoleMetrics.Load(core.MetricSynchronizedRound)
	synchronizedRound := synchronizedRoundI.(int)

	currentRoundI, _ := tc.rermuiConsoleMetrics.Load(core.MetricCurrentRound)
	currentRound := currentRoundI.(int)

	isSyncingI, _ := tc.rermuiConsoleMetrics.Load(core.MetricIsSyncing)
	isSyncing := isSyncingI.(int)

	tc.termuiConsoleWidgets.PrepareSyncInfoForDisplay(nonce, currentRound, synchronizedRound, isSyncing)

	publicKeyI, _ := tc.rermuiConsoleMetrics.Load(core.MetricPublicKey)
	publicKey := publicKeyI.(string)
	tc.termuiConsoleWidgets.PreparePublicKeyForDisplay(publicKey)

	shardIdI, _ := tc.rermuiConsoleMetrics.Load(core.MetricShardId)
	shardId := shardIdI.(int)
	tc.termuiConsoleWidgets.PrepareShardIdForDisplay(shardId)

	numConnectedPeersI, _ := tc.rermuiConsoleMetrics.Load(core.MetricNumConnectedPeers)
	numConnectedPeers := numConnectedPeersI.(int)

	txPoolLoadI, _ := tc.rermuiConsoleMetrics.Load(core.MetricTxPoolLoad)
	txPoolLoad := txPoolLoadI.(int)
	tc.termuiConsoleWidgets.PrepareTxPoolLoadForDisplay(txPoolLoad)
	tc.termuiConsoleWidgets.PrepareNumConnectedPeersForDisplay(numConnectedPeers)

	countConsensusI, _ := tc.rermuiConsoleMetrics.Load(core.MetricCountConsensus)
	countConsensus := countConsensusI.(int)

	countLeaderI, _ := tc.rermuiConsoleMetrics.Load(core.MetricCountLeader)
	countLeader := countLeaderI.(int)

	countAcceptedBlocksI, _ := tc.rermuiConsoleMetrics.Load(core.MetricCountAcceptedBlocks)
	countAcceptedBlocks := countAcceptedBlocksI.(int)

	tc.termuiConsoleWidgets.PrepareConcensusInformationsForDisplay(countConsensus, countLeader, countAcceptedBlocks)

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
