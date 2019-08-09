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
}

//NewTermuiConsole method is used to return a new TermuiConsole structure
func NewTermuiConsole() *TermuiConsole {

	tdm := TermuiConsole{}

	tdm.initMetricsMap()

	return &tdm
}

func (tdm *TermuiConsole) Write(p []byte) (n int, err error) {

	go func() {
		logLine := string(p)

		tdm.mutLogLineWrite.Lock()
		if len(tdm.logLines) >= 10 {
			tdm.logLines = tdm.logLines[1:10]
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

// SetString method - will update the value of a key
func (tdm *TermuiConsole) SetString(key string, value string) {
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

	termuiConsoleWidgets := NewtermuiConsoleGrid()

	grid := termuiConsoleWidgets.Grid()

	time.Sleep(1 * time.Second)

	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	uiEvents := ui.PollEvents()
	// handles kill signal sent to gotop
	sigTerm := make(chan os.Signal, 2)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM)

	drawTicker := time.NewTicker(time.Second).C

	for {
		select {
		case <-drawTicker:
			tdm.prepareData(termuiConsoleWidgets)

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
			case "?":
				ui.Render(grid)

			case "q":
				ui.Clear()
				return
			}
		}
	}
}

func (tdm *TermuiConsole) prepareData(termuiConsoleWidgets *termuiConsoleGrid) {
	nonceI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricNonce)
	nonce := nonceI.(int)
	termuiConsoleWidgets.PrepareNonceForDisplay(nonce)

	synchronizedRoundI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricSynchronizedRound)
	synchronizedRound := synchronizedRoundI.(int)
	termuiConsoleWidgets.PrepareSynchronizedRoundForDisplay(synchronizedRound)

	currentRoundI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricCurrentRound)
	currentRound := currentRoundI.(int)
	termuiConsoleWidgets.PrepareCurrentRoundForDisplay(currentRound)

	isSyncingI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricIsSyncing)
	isSyncing := isSyncingI.(int)
	termuiConsoleWidgets.PrepareIsSyncingForDisplay(isSyncing)

	publicKeyI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricPublicKey)
	publicKey := publicKeyI.(string)
	termuiConsoleWidgets.PreparePublicKeyForDisplay(publicKey)

	shardIdI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricShardId)
	shardId := shardIdI.(int)
	termuiConsoleWidgets.PrepareShardIdForDisplay(shardId)

	numConnectedPeersI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricNumConnectedPeers)
	numConnectedPeers := numConnectedPeersI.(int)

	txPoolLoadI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricTxPoolLoad)
	txPoolLoad := txPoolLoadI.(int)
	termuiConsoleWidgets.PrepareSparkLineGroupForDisplay(txPoolLoad, numConnectedPeers)

	countConsensusI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricCountConsensus)
	countConsensus := countConsensusI.(int)

	countLeaderI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricCountLeader)
	countLeader := countLeaderI.(int)

	countAcceptedBlocksI, _ := tdm.TermuiConsoleMetrics.Load(core.MetricCountAcceptedBlocks)
	countAcceptedBlocks := countAcceptedBlocksI.(int)

	termuiConsoleWidgets.PrepareConcensusInformationsForDisplay(countConsensus, countLeader, countAcceptedBlocks)

	termuiConsoleWidgets.PrepareListWithLogsForDisplay(tdm.logLines)
}
