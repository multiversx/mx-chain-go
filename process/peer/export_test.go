package peer

func (p *validatorStatistics) CheckForMissedBlocks(currentHeaderRound, previousHeaderRound uint64,
	prevRandSeed []byte, shardId uint32) error {
		return p.checkForMissedBlocks(currentHeaderRound, previousHeaderRound, prevRandSeed ,shardId)
}
