package presenter

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
)

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

// GetSynchronizationEstimation will return an estimation of the time required for synchronization
func (psh *PresenterStatusHandler) GetSynchronizationEstimation() string {
	currentBlock := psh.GetNonce()

	psh.mutEstimationTime.Lock()
	diffTime := time.Now().Sub(psh.startTime).Seconds()
	blocksSynchronized := currentBlock - psh.startBlock
	psh.mutEstimationTime.Unlock()

	if blocksSynchronized == 0 {
		return ""
	}

	probableHighestNonce := psh.GetProbableHighestNonce()
	remainingBlocksToSynchronize := probableHighestNonce - currentBlock
	timeEstimationSeconds := uint64(diffTime) * remainingBlocksToSynchronize / blocksSynchronized
	remainingTime := secondsToHuman(int(timeEstimationSeconds))

	return remainingTime
}

// GetSynchronizationSpeed will return speed of synchronization how many block per second are synchronized
func (psh *PresenterStatusHandler) GetSynchronizationSpeed() uint64 {
	currentBlock := psh.GetNonce()
	psh.mutEstimationTime.Lock()
	blocksSynchronized := currentBlock - psh.startBlock
	psh.mutEstimationTime.Unlock()
	if blocksSynchronized == 0 {
		return 0
	}

	blocksPerSecond := int64(currentBlock - psh.oldNonce)
	if blocksPerSecond < 0 {
		blocksPerSecond = 0
	}

	psh.oldNonce = currentBlock

	return uint64(blocksPerSecond)
}

// PrepareForCalculationSynchronizationTime prepare information that are need to calculate synchronization time
func (psh *PresenterStatusHandler) PrepareForCalculationSynchronizationTime() {
	go func() {
		oldSyncStatus := 0
		// Wait until we receive a nonce from storage or a block was synchronized
		// This wait is needed because time estimation cannot be calculated until a block was synchronized or
		// block nonce was get from storage
		for psh.GetNonce() == 0 {
			time.Sleep(time.Second)
		}

		for {
			syncStatus := psh.GetIsSyncing()
			if syncStatus == 1 && oldSyncStatus == 0 {
				psh.mutEstimationTime.Lock()
				psh.startTime = time.Now()
				psh.startBlock = psh.GetNonce()
				oldSyncStatus = 1
				psh.mutEstimationTime.Unlock()
			}
			if syncStatus == 0 {
				oldSyncStatus = 0
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
}

// GetNumTxProcessed will return number of processed transactions since node starts
func (psh *PresenterStatusHandler) GetNumTxProcessed() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNumTxProcessed)
}

// GetNumShardHeadersInPool will return number of shard headers that are in pool
func (psh *PresenterStatusHandler) GetNumShardHeadersInPool() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNumShardHeadersInPool)
}

// GetNumShardHeadersProcessed will return number of shard header processed until now
func (psh *PresenterStatusHandler) GetNumShardHeadersProcessed() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNumShardHeadersProcessed)
}
