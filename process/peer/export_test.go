package peer

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

func (p *validatorStatistics) CheckForMissedBlocks(
	currentHeaderRound uint64,
	previousHeaderRound uint64,
	prevRandSeed []byte,
	shardId uint32,
) error {
	return p.checkForMissedBlocks(currentHeaderRound, previousHeaderRound, prevRandSeed, shardId)
}

func (p *validatorStatistics) SaveInitialState(in []*sharding.InitialNode, stakeValue *big.Int, initialRating uint32) error {
	return p.saveInitialState(in, stakeValue, initialRating)
}

func (p *validatorStatistics) GetMatchingPrevShardData(currentShardData block.ShardData, shardInfo []block.ShardData) *block.ShardData {
	return p.getMatchingPrevShardData(currentShardData, shardInfo)
}

func (p *validatorStatistics) LoadPreviousShardHeaders(currentHeader, previousHeader *block.MetaBlock) error {
	return p.loadPreviousShardHeaders(currentHeader, previousHeader)
}

func (p *validatorStatistics) LoadPreviousShardHeadersMeta(currentHeader, previousHeader *block.MetaBlock) error {
	return p.loadPreviousShardHeadersMeta(currentHeader)
}

func (p *validatorStatistics) PrevShardInfo() map[string]block.ShardData {
	p.mutPrevShardInfo.RLock()
	defer p.mutPrevShardInfo.RUnlock()
	return p.prevShardInfo
}

func (p *validatorStatistics) BuildShardDataKey(sh block.ShardData) string {
	return p.buildShardDataKey(sh)
}
