package peer

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
