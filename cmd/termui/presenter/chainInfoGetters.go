package presenter

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/cmd/termui/provider"
	"github.com/multiversx/mx-chain-go/common"
)

var maxSpeedHistorySaved = 2000

// GetNonce will return current nonce of node
func (psh *PresenterStatusHandler) GetNonce() uint64 {
	return psh.getFromCacheAsUint64(common.MetricNonce)
}

// GetLastExecutedNonce will return current last executed nonce of node
func (psh *PresenterStatusHandler) GetLastExecutedNonce() uint64 {
	return psh.getFromCacheAsUint64(common.MetricLastExecutedNonce)
}

// GetProposedNonce will return current proposed nonce of node
func (psh *PresenterStatusHandler) GetProposedNonce() uint64 {
	return psh.getFromCacheAsUint64(common.MetricProposedNonce)
}

// GetIsSyncing will return state of the node
func (psh *PresenterStatusHandler) GetIsSyncing() uint64 {
	return psh.getFromCacheAsUint64(common.MetricIsSyncing)
}

// GetTxPoolLoad will return how many transactions are in the pool
func (psh *PresenterStatusHandler) GetTxPoolLoad() uint64 {
	return psh.getFromCacheAsUint64(common.MetricTxPoolLoad)
}

// GetProbableHighestNonce will return the highest nonce of blockchain
func (psh *PresenterStatusHandler) GetProbableHighestNonce() uint64 {
	return psh.getFromCacheAsUint64(common.MetricProbableHighestNonce)
}

// GetSynchronizedRound will return number of synchronized round
func (psh *PresenterStatusHandler) GetSynchronizedRound() uint64 {
	return psh.getFromCacheAsUint64(common.MetricSynchronizedRound)
}

// GetRoundTime will return duration of a round
func (psh *PresenterStatusHandler) GetRoundTime() uint64 {
	return psh.getFromCacheAsUint64(common.MetricRoundTime)
}

// GetLiveValidatorNodes will return how many validator nodes are in blockchain and known by the current node to be active
func (psh *PresenterStatusHandler) GetLiveValidatorNodes() uint64 {
	return psh.getFromCacheAsUint64(common.MetricLiveValidatorNodes)
}

// GetConnectedNodes will return how many intra-shard nodes are connected
func (psh *PresenterStatusHandler) GetConnectedNodes() uint64 {
	return psh.getFromCacheAsUint64(common.MetricConnectedNodes)
}

// GetNumConnectedPeers will return how many peers are connected
func (psh *PresenterStatusHandler) GetNumConnectedPeers() uint64 {
	return psh.getFromCacheAsUint64(common.MetricNumConnectedPeers)
}

// GetIntraShardValidators will return how many intra-shard validator nodes are and known by the current node to be active
func (psh *PresenterStatusHandler) GetIntraShardValidators() uint64 {
	return psh.getFromCacheAsUint64(common.MetricNumIntraShardValidatorNodes)
}

// GetCurrentRound will return current round of node
func (psh *PresenterStatusHandler) GetCurrentRound() uint64 {
	return psh.getFromCacheAsUint64(common.MetricCurrentRound)
}

// CalculateTimeToSynchronize will calculate and return an estimation of
// the time required for synchronization in a human friendly format
func (psh *PresenterStatusHandler) CalculateTimeToSynchronize(numMillisecondsRefreshTime int) string {
	if numMillisecondsRefreshTime < 1 {
		return "N/A"
	}

	currentSynchronizedRound := psh.GetSynchronizedRound()

	speed := psh.calculateSpeedFromSpeedHistory(numMillisecondsRefreshTime)
	currentRound := psh.GetCurrentRound()
	if currentRound < currentSynchronizedRound || speed == 0 {
		return "Estimating..."
	}

	remainingRoundsToSynchronize := currentRound - currentSynchronizedRound
	timeEstimationSeconds := float64(remainingRoundsToSynchronize) / float64(speed)
	remainingTime := core.SecondsToHourMinSec(int(timeEstimationSeconds))

	return remainingTime
}

// CalculateSynchronizationSpeed will calculate and return speed of synchronization
// how many blocks per second are synchronized
func (psh *PresenterStatusHandler) CalculateSynchronizationSpeed(numMillisecondsRefreshTime int) uint64 {
	currentSynchronizedRound := psh.GetSynchronizedRound()

	if len(psh.synchronizationSpeedHistory) >= maxSpeedHistorySaved {
		psh.synchronizationSpeedHistory = psh.synchronizationSpeedHistory[1:]
	}
	psh.synchronizationSpeedHistory = append(psh.synchronizationSpeedHistory, currentSynchronizedRound)

	speed := psh.calculateSpeedFromSpeedHistory(numMillisecondsRefreshTime)

	return speed
}

func (psh *PresenterStatusHandler) calculateSpeedFromSpeedHistory(numMillisecondsRefreshTime int) uint64 {
	lastIndex := len(psh.synchronizationSpeedHistory) - 1
	firstIndex := 0

	if lastIndex <= firstIndex {
		return 0
	}

	numSyncedBlocks := psh.synchronizationSpeedHistory[lastIndex] - psh.synchronizationSpeedHistory[firstIndex]
	cumulatedTimeMs := uint64((lastIndex - firstIndex) * numMillisecondsRefreshTime)

	if cumulatedTimeMs == 0 || numSyncedBlocks == 0 {
		return 0
	}

	numMillisecondsInASecond := 1000.0
	speed := (float64(numSyncedBlocks) / float64(cumulatedTimeMs)) * numMillisecondsInASecond

	return uint64(speed)

}

// GetNumTxProcessed will return number of processed transactions since node starts
func (psh *PresenterStatusHandler) GetNumTxProcessed() uint64 {
	return psh.getFromCacheAsUint64(common.MetricNumProcessedTxs)
}

// GetNumShardHeadersInPool will return number of shard headers that are in pool
func (psh *PresenterStatusHandler) GetNumShardHeadersInPool() uint64 {
	return psh.getFromCacheAsUint64(common.MetricNumShardHeadersFromPool)
}

// GetNumShardHeadersProcessed will return number of shard header processed until now
func (psh *PresenterStatusHandler) GetNumShardHeadersProcessed() uint64 {
	return psh.getFromCacheAsUint64(common.MetricNumShardHeadersProcessed)
}

// GetEpochInfo will return information about current epoch
func (psh *PresenterStatusHandler) GetEpochInfo() (uint64, uint64, int, string) {
	roundAtEpochStart := psh.getFromCacheAsUint64(common.MetricRoundAtEpochStart)
	roundsPerEpoch := psh.getFromCacheAsUint64(common.MetricRoundsPerEpoch)
	currentRound := psh.getFromCacheAsUint64(common.MetricCurrentRound)
	roundDuration := psh.getFromCacheAsUint64(common.MetricRoundDuration)

	epochFinishRound := roundAtEpochStart + roundsPerEpoch
	roundsRemained := epochFinishRound - currentRound
	if epochFinishRound < currentRound {
		roundsRemained = 0
	}
	if roundsPerEpoch == 0 || roundDuration == 0 {
		return 0, 0, 0, ""
	}
	secondsRemainedInEpoch := roundsRemained * roundDuration / 1000

	remainingTime := core.SecondsToHourMinSec(int(secondsRemainedInEpoch))
	epochLoadPercent := 100 - int(float64(roundsRemained)/float64(roundsPerEpoch)*100.0)

	return currentRound, epochFinishRound, epochLoadPercent, remainingTime
}

// GetTrieSyncNumProcessedNodes will return the number of processed nodes during trie sync
func (psh *PresenterStatusHandler) GetTrieSyncNumProcessedNodes() uint64 {
	return psh.getFromCacheAsUint64(common.MetricTrieSyncNumProcessedNodes)
}

// GetTrieSyncProcessedPercentage will return the number of processed nodes during trie sync
func (psh *PresenterStatusHandler) GetTrieSyncProcessedPercentage() core.OptionalUint64 {
	numEstimatedNodes := psh.getFromCacheAsUint64(provider.AccountsSnapshotNumNodesMetric)
	if numEstimatedNodes <= 0 {
		return core.OptionalUint64{
			Value:    0,
			HasValue: false,
		}
	}

	numProcessedNodes := psh.GetTrieSyncNumProcessedNodes()

	percentage := (numProcessedNodes * 100) / numEstimatedNodes
	if percentage > 100 {
		return core.OptionalUint64{
			Value:    100,
			HasValue: true,
		}
	}

	return core.OptionalUint64{
		Value:    percentage,
		HasValue: true,
	}
}

// GetTrieSyncNumBytesReceived will return the number of bytes synced during trie sync
func (psh *PresenterStatusHandler) GetTrieSyncNumBytesReceived() uint64 {
	return psh.getFromCacheAsUint64(common.MetricTrieSyncNumReceivedBytes)
}
