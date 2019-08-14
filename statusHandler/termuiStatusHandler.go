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

// IsInterfaceNil return if there is no value under the interface
func (tsh *TermuiStatusHandler) IsInterfaceNil() bool {
	if tsh == nil {
		return true
	}

	return false
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

	tsh.termuiConsoleMetrics.Store(core.MetricPublicKeyTxSign, "")
	tsh.termuiConsoleMetrics.Store(core.MetricPublicKeyBlockSign, "")
	tsh.termuiConsoleMetrics.Store(core.MetricShardId, uint64(0))

	tsh.termuiConsoleMetrics.Store(core.MetricNodeType, "")
	tsh.termuiConsoleMetrics.Store(core.MetricCountConsensus, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricCountLeader, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricCountAcceptedBlocks, uint64(0))

	tsh.termuiConsoleMetrics.Store(core.MetricIsSyncing, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricNonce, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricProbableHighestNonce, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricCurrentRound, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricSynchronizedRound, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricRoundTime, uint64(0))

	tsh.termuiConsoleMetrics.Store(core.MetricLiveValidatorNodes, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricConnectedNodes, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricNumConnectedPeers, uint64(0))

	tsh.termuiConsoleMetrics.Store(core.MetricCpuLoadPercent, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricMemLoadPercent, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricTotalMem, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricNetworkRecvBps, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricNetworkRecvBpsPeak, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricNetworkRecvPercent, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricNetworkSentBps, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricNetworkSentBpsPeak, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricNetworkSentPercent, uint64(0))
	tsh.termuiConsoleMetrics.Store(core.MetricTxPoolLoad, uint64(0))

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
	keyValueI, ok := tsh.termuiConsoleMetrics.Load(key)
	if !ok {
		return
	}

	keyValue, ok := keyValueI.(uint64)
	if !ok {
		return
	}

	keyValue++
	tsh.termuiConsoleMetrics.Store(key, keyValue)

}

// Decrement - will decrement the value of a key
func (tsh *TermuiStatusHandler) Decrement(key string) {
}

// Close method - won't do anything
func (tsh *TermuiStatusHandler) Close() {
}
