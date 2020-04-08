package peer

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// CheckForMissedBlocks -
func (vs *validatorStatistics) CheckForMissedBlocks(
	currentHeaderRound uint64,
	previousHeaderRound uint64,
	prevRandSeed []byte,
	shardId uint32,
	epoch uint32,
) error {
	return vs.checkForMissedBlocks(currentHeaderRound, previousHeaderRound, prevRandSeed, shardId, epoch)
}

// GetMatchingPrevShardData -
func (vs *validatorStatistics) GetMatchingPrevShardData(currentShardData block.ShardData, shardInfo []block.ShardData) *block.ShardData {
	return vs.getMatchingPrevShardData(currentShardData, shardInfo)
}

// GetLeaderDecreaseCount -
func (vs *validatorStatistics) GetLeaderDecreaseCount(key []byte) uint32 {
	vs.mutMissedBlocksCounters.RLock()
	defer vs.mutMissedBlocksCounters.RUnlock()

	return vs.missedBlocksCounters.get(key).leaderDecreaseCount
}

func (vs *validatorStatistics) UpdateMissedBlocksCounters() error {
	return vs.updateMissedBlocksCounters()
}
