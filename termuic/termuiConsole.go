package termuic

import (
	"os"
	"os/signal"
	"runtime"
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
	termuiConsoleMetrics *sync.Map
	logLines             []string
	mutLogLineWrite      sync.RWMutex
	numLogLines          int
	termuiConsoleWidgets *termuiConsoleGrid
	grid                 *ui.Grid
}

//NewTermuiConsole method is used to return a new TermuiConsole structure
func NewTermuiConsole(metricData *sync.Map) *TermuiConsole {
	tc := TermuiConsole{
		termuiConsoleMetrics: metricData,
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

	case "<C-c>":
		ui.Close()
		if p, err := os.FindProcess(os.Getpid()); err != nil {
			return
		} else {
			if runtime.GOOS == "windows" {
				_ = p.Kill()
			} else {
				_ = p.Signal(syscall.SIGINT)
			}
		}
		return
	}
}

func (tc *TermuiConsole) refreshDataForConsole() {
	if tc.numLogLines == 10 {
		tc.prepareDataNormalWidgets()
	}
	tc.termuiConsoleWidgets.PrepareListWithLogsForDisplay(tc.logLines)

	ui.Clear()
	ui.Render(tc.grid)
}

func (tc *TermuiConsole) prepareDataNormalWidgets() {
	nonceI, _ := tc.termuiConsoleMetrics.Load(core.MetricNonce)
	nonce := nonceI.(int)

	synchronizedRoundI, _ := tc.termuiConsoleMetrics.Load(core.MetricSynchronizedRound)
	synchronizedRound := synchronizedRoundI.(int)

	currentRoundI, _ := tc.termuiConsoleMetrics.Load(core.MetricCurrentRound)
	currentRound := currentRoundI.(int)

	isSyncingI, _ := tc.termuiConsoleMetrics.Load(core.MetricIsSyncing)
	isSyncing := isSyncingI.(int)

	tc.termuiConsoleWidgets.PrepareSyncInfoForDisplay(nonce, currentRound, synchronizedRound, isSyncing)

	publicKeyI, _ := tc.termuiConsoleMetrics.Load(core.MetricPublicKey)
	publicKey := publicKeyI.(string)
	tc.termuiConsoleWidgets.PreparePublicKeyForDisplay(publicKey)

	shardIdI, _ := tc.termuiConsoleMetrics.Load(core.MetricShardId)
	shardId := shardIdI.(int)
	tc.termuiConsoleWidgets.PrepareShardIdForDisplay(shardId)

	numConnectedPeersI, _ := tc.termuiConsoleMetrics.Load(core.MetricNumConnectedPeers)
	numConnectedPeers := numConnectedPeersI.(int)

	txPoolLoadI, _ := tc.termuiConsoleMetrics.Load(core.MetricTxPoolLoad)
	txPoolLoad := txPoolLoadI.(int)
	tc.termuiConsoleWidgets.PrepareTxPoolLoadForDisplay(txPoolLoad)
	tc.termuiConsoleWidgets.PrepareNumConnectedPeersForDisplay(numConnectedPeers)

	countConsensusI, _ := tc.termuiConsoleMetrics.Load(core.MetricCountConsensus)
	countConsensus := countConsensusI.(int)

	countLeaderI, _ := tc.termuiConsoleMetrics.Load(core.MetricCountLeader)
	countLeader := countLeaderI.(int)

	countAcceptedBlocksI, _ := tc.termuiConsoleMetrics.Load(core.MetricCountAcceptedBlocks)
	countAcceptedBlocks := countAcceptedBlocksI.(int)

	tc.termuiConsoleWidgets.PrepareConsensusInformationsForDisplay(countConsensus, countLeader, countAcceptedBlocks)
}

func (tc *TermuiConsole) changeConsoleDisplay() *ui.Grid {
	if tc.numLogLines == 10 {
		tc.numLogLines = 50

		return tc.configConsoleWithBigLog()
	}
	tc.numLogLines = 10

	return tc.configConsoleWithNormalTermuiWidgets()
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
