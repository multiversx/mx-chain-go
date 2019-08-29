package statusHandler

import (
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
)

//maxLogLines is used to specify how many lines of logs need to store in slice
var maxLogLines = 100

// PresenterStatusHandler will be used when an AppStatusHandler is required, but another one isn't necessary or available
type PresenterStatusHandler struct {
	presenterMetrics *sync.Map
	logLines         []string
	mutLogLineWrite  sync.RWMutex
}

// NewPresenterStatusHandler will return an instance of the struct
func NewPresenterStatusHandler() *PresenterStatusHandler {
	psh := &PresenterStatusHandler{
		presenterMetrics: &sync.Map{},
	}
	return psh
}

// IsInterfaceNil returns true if there is no value under the interface
func (psh *PresenterStatusHandler) IsInterfaceNil() bool {
	if psh == nil {
		return true
	}
	return false
}

// SetInt64Value method - will update the value for a key
func (psh *PresenterStatusHandler) SetInt64Value(key string, value int64) {
	psh.presenterMetrics.Store(key, value)
}

// SetUInt64Value method - will update the value for a key
func (psh *PresenterStatusHandler) SetUInt64Value(key string, value uint64) {
	psh.presenterMetrics.Store(key, value)
}

// SetStringValue method - will update the value of a key
func (psh *PresenterStatusHandler) SetStringValue(key string, value string) {
	psh.presenterMetrics.Store(key, value)
}

// Increment - will increment the value of a key
func (psh *PresenterStatusHandler) Increment(key string) {
	keyValueI, ok := psh.presenterMetrics.Load(key)
	if !ok {
		return
	}

	keyValue, ok := keyValueI.(uint64)
	if !ok {
		return
	}

	keyValue++
	psh.presenterMetrics.Store(key, keyValue)
}

// Decrement - will decrement the value of a key
func (psh *PresenterStatusHandler) Decrement(key string) {
}

// Close method - won't do anything
func (psh *PresenterStatusHandler) Close() {
}

func (psh *PresenterStatusHandler) Write(p []byte) (n int, err error) {
	go func(p []byte) {
		logLine := string(p)
		stringSlice := strings.Split(logLine, "\n")

		psh.mutLogLineWrite.Lock()
		for _, line := range stringSlice {
			line = strings.Replace(line, "\r", "", len(line))
			if line != "" {
				psh.logLines = append(psh.logLines, line)
			}
		}

		startPos := len(psh.logLines) - maxLogLines
		if startPos < 0 {
			startPos = 0
		}
		psh.logLines = psh.logLines[startPos:len(psh.logLines)]

		psh.mutLogLineWrite.Unlock()
	}(p)

	return len(p), nil
}

// GetAppVersion will return application version
func (psh *PresenterStatusHandler) GetAppVersion() string {
	return psh.getFromCacheAsString(core.MetricAppVersion)
}

// GetPublicKeyTxSign will return node public key for sign transaction
func (psh *PresenterStatusHandler) GetPublicKeyTxSign() string {
	return psh.getFromCacheAsString(core.MetricPublicKeyTxSign)
}

// GetPublicKeyBlockSign will return node public key for sign blocks
func (psh *PresenterStatusHandler) GetPublicKeyBlockSign() string {
	return psh.getFromCacheAsString(core.MetricPublicKeyBlockSign)
}

// GetShardId will return shard Id of node
func (psh *PresenterStatusHandler) GetShardId() uint64 {
	return psh.getFromCacheAsUint64(core.MetricShardId)
}

// GetCountConsensus will return count of how many times node was in consensus group
func (psh *PresenterStatusHandler) GetCountConsensus() uint64 {
	return psh.getFromCacheAsUint64(core.MetricCountConsensus)
}

// GetCountLeader will return count of how many time node was leader in consensus group
func (psh *PresenterStatusHandler) GetCountLeader() uint64 {
	return psh.getFromCacheAsUint64(core.MetricCountLeader)
}

// GetCountAcceptedBlocks will return count of how many accepted blocks was proposed by the node
func (psh *PresenterStatusHandler) GetCountAcceptedBlocks() uint64 {
	return psh.getFromCacheAsUint64(core.MetricCountAcceptedBlocks)
}

// GetNonce will return current nonce of node
func (psh *PresenterStatusHandler) GetNonce() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNonce)
}

// GetIsSyncing will return state of the node
func (psh *PresenterStatusHandler) GetIsSyncing() uint64 {
	return psh.getFromCacheAsUint64(core.MetricIsSyncing)
}

// GetTxPoolLoad will return how many transactions are in the pool
func (psh *PresenterStatusHandler) GetTxPoolLoad() uint64 {
	return psh.getFromCacheAsUint64(core.MetricTxPoolLoad)
}

// GetProbableHighestNonce will return the highest nonce of blockchain
func (psh *PresenterStatusHandler) GetProbableHighestNonce() uint64 {
	return psh.getFromCacheAsUint64(core.MetricProbableHighestNonce)
}

// GetSynchronizedRound will return number of synchronized round
func (psh *PresenterStatusHandler) GetSynchronizedRound() uint64 {
	return psh.getFromCacheAsUint64(core.MetricSynchronizedRound)
}

// GetRoundTime will return duration of a round
func (psh *PresenterStatusHandler) GetRoundTime() uint64 {
	return psh.getFromCacheAsUint64(core.MetricRoundTime)
}

// GetLiveValidatorNodes will return how many validator nodes are in blockchain
func (psh *PresenterStatusHandler) GetLiveValidatorNodes() uint64 {
	return psh.getFromCacheAsUint64(core.MetricLiveValidatorNodes)
}

// GetConnectedNodes will return how many nodes are connected
func (psh *PresenterStatusHandler) GetConnectedNodes() uint64 {
	return psh.getFromCacheAsUint64(core.MetricConnectedNodes)
}

// GetNumConnectedPeers will return how many peers are connected
func (psh *PresenterStatusHandler) GetNumConnectedPeers() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNumConnectedPeers)
}

// GetCurrentRound will return current round of node
func (psh *PresenterStatusHandler) GetCurrentRound() uint64 {
	return psh.getFromCacheAsUint64(core.MetricCurrentRound)
}

// GetNodeType will return type of node
func (psh *PresenterStatusHandler) GetNodeType() string {
	return psh.getFromCacheAsString(core.MetricNodeType)
}

// GetNumTxInBlock will return how many transactions are in block
func (psh *PresenterStatusHandler) GetNumTxInBlock() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNumTxInBlock)
}

// GetNumMiniBlocks will return how many miniblocks are in a block
func (psh *PresenterStatusHandler) GetNumMiniBlocks() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNumMiniBlocks)
}

// GetCrossCheckBlockHeight will return cross block height
func (psh *PresenterStatusHandler) GetCrossCheckBlockHeight() string {
	return psh.getFromCacheAsString(core.MetricCrossCheckBlockHeight)
}

// GetConsensusState will return consensus state of node
func (psh *PresenterStatusHandler) GetConsensusState() string {
	return psh.getFromCacheAsString(core.MetricConsensusState)
}

// GetConsensusRoundState will return consensus round state
func (psh *PresenterStatusHandler) GetConsensusRoundState() string {
	return psh.getFromCacheAsString(core.MetricConsensusRoundState)
}

// GetLogLines will return log lines that's need to be displayed
func (psh *PresenterStatusHandler) GetLogLines() []string {
	return psh.logLines
}

// GetCpuLoadPercent wil return cpu load
func (psh *PresenterStatusHandler) GetCpuLoadPercent() uint64 {
	return psh.getFromCacheAsUint64(core.MetricCpuLoadPercent)
}

// GetMemLoadPercent will return mem percent usage of node
func (psh *PresenterStatusHandler) GetMemLoadPercent() uint64 {
	return psh.getFromCacheAsUint64(core.MetricMemLoadPercent)
}

// GetTotalMem will return how much memory does the computer have
func (psh *PresenterStatusHandler) GetTotalMem() uint64 {
	return psh.getFromCacheAsUint64(core.MetricTotalMem)
}

// GetMemUsedByNode will return how many memory node uses
func (psh *PresenterStatusHandler) GetMemUsedByNode() uint64 {
	return psh.getFromCacheAsUint64(core.MetricMemoryUsedByNode)
}

// GetNetworkRecvPercent will return network receive percent
func (psh *PresenterStatusHandler) GetNetworkRecvPercent() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNetworkRecvPercent)
}

// GetNetworkRecvBps will return network received bytes per second
func (psh *PresenterStatusHandler) GetNetworkRecvBps() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNetworkRecvBps)
}

// GetNetworkRecvBpsPeak will return received bytes per seconds peak
func (psh *PresenterStatusHandler) GetNetworkRecvBpsPeak() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNetworkRecvBpsPeak)
}

// GetNetworkSentPercent will return network sent percent
func (psh *PresenterStatusHandler) GetNetworkSentPercent() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNetworkSentPercent)
}

// GetNetworkSentBps will return network sent bytes per second
func (psh *PresenterStatusHandler) GetNetworkSentBps() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNetworkSentBps)
}

// GetNetworkSentBpsPeak will return sent bytes per seconds peak
func (psh *PresenterStatusHandler) GetNetworkSentBpsPeak() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNetworkSentBpsPeak)
}

func (psh *PresenterStatusHandler) getFromCacheAsUint64(metric string) uint64 {
	val, ok := psh.presenterMetrics.Load(metric)
	if !ok {
		return 0
	}

	valUint64, ok := val.(uint64)
	if !ok {
		return 0
	}

	return valUint64
}

func (psh *PresenterStatusHandler) getFromCacheAsString(metric string) string {
	val, ok := psh.presenterMetrics.Load(metric)
	if !ok {
		return "[invalid key]"
	}

	valStr, ok := val.(string)
	if !ok {
		return "[not a string]"
	}

	return valStr
}
