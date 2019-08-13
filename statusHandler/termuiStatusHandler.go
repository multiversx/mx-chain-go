package statusHandler

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/statusHandler/termuic"
)

// TermuiStatusHandler will be used when an AppStatusHandler is required, but another one isn't necessary or available
type TermuiStatusHandler struct {
	tui                  *termuic.TermuiConsole
	termuiConsoleMetrics *sync.Map
}

// NewTermuiStatusHandler will return an instance of the struct
func NewTermuiStatusHandler() *TermuiStatusHandler {
	tsh := new(TermuiStatusHandler)
	tsh.initMetricsMap()
	tsh.tui = termuic.NewTermuiConsole(tsh.termuiConsoleMetrics)

	return tsh
}

// StartTermuiConsole method will start termui console
func (tsh *TermuiStatusHandler) StartTermuiConsole() error {
	err := tsh.tui.Start()

	return err
}

//Termui method - returns address of TermuiConsole structure from TermuiStatusHandler
func (tsh *TermuiStatusHandler) Termui() *termuic.TermuiConsole {
	return tsh.tui
}

// InitMetricsMap will init the map of prometheus metrics
func (tsh *TermuiStatusHandler) initMetricsMap() {
	tsh.termuiConsoleMetrics = &sync.Map{}

	tsh.termuiConsoleMetrics.Store(core.MetricNonce, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricCurrentRound, int64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricIsSyncing, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricNumConnectedPeers, int64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricSynchronizedRound, uint64(0))

	tsh.termuiConsoleMetrics.Store(core.MetricPublicKey, "")
	tsh.termuiConsoleMetrics.Store(core.MetricShardId, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricTxPoolLoad, int64(0))

	tsh.termuiConsoleMetrics.Store(core.MetricCountConsensus, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricCountLeader, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricCountAcceptedBlocks, uint64(0))
}

// SetInt64Value method - will update the value for a key
func (tsh *TermuiStatusHandler) SetInt64Value(key string, value int64) {
	if _, ok := tsh.termuiConsoleMetrics.Load(key); ok {
		tsh.termuiConsoleMetrics.Store(key, value)
	}
}

// SetUInt64Value method - will update the value for a key
func (tsh *TermuiStatusHandler) SetUInt64Value(key string, value uint64) {
	if _, ok := tsh.termuiConsoleMetrics.Load(key); ok {
		tsh.termuiConsoleMetrics.Store(key, value)
	}
}

// SetStringValue method - will update the value of a key
func (tsh *TermuiStatusHandler) SetStringValue(key string, value string) {
	if _, ok := tsh.termuiConsoleMetrics.Load(key); ok {
		tsh.termuiConsoleMetrics.Store(key, value)
	}
}

// Increment - will increment the value of a key
func (tsh *TermuiStatusHandler) Increment(key string) {
	if keyValueI, ok := tsh.termuiConsoleMetrics.Load(key); ok {

		keyValue := keyValueI.(uint64)
		keyValue++
		tsh.termuiConsoleMetrics.Store(key, keyValue)
	}
}

// Decrement - will decrement the value of a key
func (tsh *TermuiStatusHandler) Decrement(key string) {

}

// Close method - won't do anything
func (tsh *TermuiStatusHandler) Close() {
}
