package presenter

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

var maxSpeedHistorySaved = 1000

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

// CalculateTimeToSynchronize will calculate and return an estimation of
// the time required for synchronization in a human friendly format
func (psh *PresenterStatusHandler) CalculateTimeToSynchronize() string {
	currentBlock := psh.GetNonce()

	numsynchronizationSpeedHistory := len(psh.synchronizationSpeedHistory)

	sum := uint64(0)
	for i := 0; i < len(psh.synchronizationSpeedHistory); i++ {
		sum += psh.synchronizationSpeedHistory[i]
	}

	speed := float64(0)
	if numsynchronizationSpeedHistory > 0 {
		speed = float64(sum) / float64(numsynchronizationSpeedHistory)
	}

	probableHighestNonce := psh.GetProbableHighestNonce()
	remainingBlocksToSynchronize := probableHighestNonce - currentBlock
	timeEstimationSeconds := float64(remainingBlocksToSynchronize) / speed
	remainingTime := core.SecondsToHumanFormat(int(timeEstimationSeconds))

	return remainingTime
}

// CalculateSynchronizationSpeed will calculate and return speed of synchronization
// how many block per second are synchronized
func (psh *PresenterStatusHandler) CalculateSynchronizationSpeed() uint64 {
	currentBlock := psh.GetNonce()
	if psh.oldNonce == 0 {
		psh.oldNonce = currentBlock
		return 0
	}

	blocksPerSecond := int64(currentBlock - psh.oldNonce)
	if blocksPerSecond < 0 {
		blocksPerSecond = 0
	}

	if len(psh.synchronizationSpeedHistory) >= maxSpeedHistorySaved {
		psh.synchronizationSpeedHistory = psh.synchronizationSpeedHistory[1:len(psh.synchronizationSpeedHistory)]
	}
	psh.synchronizationSpeedHistory = append(psh.synchronizationSpeedHistory, uint64(blocksPerSecond))

	psh.oldNonce = currentBlock

	return uint64(blocksPerSecond)
}

// GetNumTxProcessed will return number of processed transactions since node starts
func (psh *PresenterStatusHandler) GetNumTxProcessed() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNumProcessedTxs)
}

// GetNumShardHeadersInPool will return number of shard headers that are in pool
func (psh *PresenterStatusHandler) GetNumShardHeadersInPool() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNumShardHeadersFromPool)
}

// GetNumShardHeadersProcessed will return number of shard header processed until now
func (psh *PresenterStatusHandler) GetNumShardHeadersProcessed() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNumShardHeadersProcessed)
}
