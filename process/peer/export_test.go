package peer

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/sharding"
)

type ShardMediator struct {
	shardMediator
}

func (p *validatorStatistics) CheckForMissedBlocks(
	currentHeaderRound uint64,
	previousHeaderRound uint64,
	prevRandSeed []byte,
	shardId uint32,
) error {
	return p.checkForMissedBlocks(currentHeaderRound, previousHeaderRound, prevRandSeed ,shardId)
}

func (p *validatorStatistics) SaveInitialState(in []*sharding.InitialNode, stakeValue *big.Int) error {
	return p.saveInitialState(in, stakeValue)
}
