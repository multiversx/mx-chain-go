package presenter

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

var maxSpeedHistorySaved = 2000

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
func (psh *PresenterStatusHandler) CalculateTimeToSynchronize(numMillisecondsRefreshTime int) string {
	if numMillisecondsRefreshTime < 1 {
		return "N/A"
	}

	currentSynchronizedRound := psh.GetSynchronizedRound()

	numSynchronizationSpeedHistory := len(psh.synchronizationSpeedHistory)

	sum := uint64(0)
	for i := 0; i < len(psh.synchronizationSpeedHistory); i++ {
		sum += psh.synchronizationSpeedHistory[i]
	}

	speed := float64(0)
	if numSynchronizationSpeedHistory > 0 {
		speed = float64(sum*1000) / float64(numSynchronizationSpeedHistory*numMillisecondsRefreshTime)
	}

	currentRound := psh.GetCurrentRound()
	if currentRound < currentSynchronizedRound || speed == 0 {
		return ""
	}

	remainingRoundsToSynchronize := currentRound - currentSynchronizedRound
	timeEstimationSeconds := float64(remainingRoundsToSynchronize) / speed
	remainingTime := core.SecondsToHourMinSec(int(timeEstimationSeconds))

	return remainingTime
}

// CalculateSynchronizationSpeed will calculate and return speed of synchronization
// how many blocks per second are synchronized
func (psh *PresenterStatusHandler) CalculateSynchronizationSpeed(numMillisecondsRefreshTime int) uint64 {
	currentSynchronizedRound := psh.GetSynchronizedRound()
	if psh.oldRound == 0 {
		psh.oldRound = currentSynchronizedRound
		return 0
	}

	roundsPerSecond := int64(currentSynchronizedRound - psh.oldRound)
	if roundsPerSecond < 0 {
		roundsPerSecond = 0
	}

	if len(psh.synchronizationSpeedHistory) >= maxSpeedHistorySaved {
		psh.synchronizationSpeedHistory = psh.synchronizationSpeedHistory[1:len(psh.synchronizationSpeedHistory)]
	}
	psh.synchronizationSpeedHistory = append(psh.synchronizationSpeedHistory, uint64(roundsPerSecond))

	psh.oldRound = currentSynchronizedRound

	numSyncedBlocks := uint64(0)
	cumulatedTime := uint64(0)
	lastIndex := len(psh.synchronizationSpeedHistory) - 1
	millisecondsInASecond := uint64(1000)
	for {
		if lastIndex < 0 {
			break
		}
		if cumulatedTime >= millisecondsInASecond {
			break
		}

		numSyncedBlocks += psh.synchronizationSpeedHistory[lastIndex]
		lastIndex--
		cumulatedTime += uint64(numMillisecondsRefreshTime)
	}
	if cumulatedTime == 0 || numSyncedBlocks == 0 {
		return 0
	}

	timeAdjustment := float64(millisecondsInASecond) / float64(cumulatedTime)
	syncedBlocksAdjustment := timeAdjustment * float64(numSyncedBlocks)

	return uint64(syncedBlocksAdjustment)
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

// GetEpochInfo will return information about current epoch
func (psh *PresenterStatusHandler) GetEpochInfo() (uint64, uint64, int, string) {
	roundAtEpochStart := psh.getFromCacheAsUint64(core.MetricRoundAtEpochStart)
	roundsPerEpoch := psh.getFromCacheAsUint64(core.MetricRoundsPerEpoch)
	currentRound := psh.getFromCacheAsUint64(core.MetricCurrentRound)
	roundDuration := psh.getFromCacheAsUint64(core.MetricRoundDuration)

	epochFinishRound := roundAtEpochStart + roundsPerEpoch
	roundsRemained := epochFinishRound - currentRound
	if epochFinishRound < currentRound {
		roundsRemained = 0
	}
	if roundsRemained == 0 || roundsPerEpoch == 0 || roundDuration == 0 {
		return 0, 0, 0, ""
	}
	secondsRemainedInEpoch := roundsRemained * roundDuration / 1000

	remainingTime := core.SecondsToHourMinSec(int(secondsRemainedInEpoch))

	epochLoadPercent := 100 - int(float64(roundsRemained)/float64(roundsPerEpoch)*100.0)

	return currentRound, epochFinishRound, epochLoadPercent, remainingTime
}
