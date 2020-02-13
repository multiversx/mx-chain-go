package peer

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

func (vs *validatorStatistics) CheckForMissedBlocks(
	currentHeaderRound uint64,
	previousHeaderRound uint64,
	prevRandSeed []byte,
	shardId uint32,
) error {
	return vs.checkForMissedBlocks(currentHeaderRound, previousHeaderRound, prevRandSeed, shardId)
}

func (vs *validatorStatistics) SaveInitialState(in []*sharding.InitialNode, stakeValue *big.Int, initialRating uint32) error {
	return vs.saveInitialState(in, stakeValue, initialRating)
}

func (vs *validatorStatistics) GetMatchingPrevShardData(currentShardData block.ShardData, shardInfo []block.ShardData) *block.ShardData {
	return vs.getMatchingPrevShardData(currentShardData, shardInfo)
}

func (vs *validatorStatistics) GetLeaderDecreaseCount(key []byte) uint32 {
	vs.mutMissedBlocksCounters.RLock()
	defer vs.mutMissedBlocksCounters.RUnlock()

	return vs.missedBlocksCounters.get(key).leaderDecreaseCount
}
