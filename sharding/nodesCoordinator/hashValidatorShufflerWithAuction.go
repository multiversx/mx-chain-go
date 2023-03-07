package nodesCoordinator

func getNumPubKeys(shardValidatorsMap map[uint32][]Validator) uint32 {
	numPubKeys := uint32(0)

	for _, validatorsInShard := range shardValidatorsMap {
		numPubKeys += uint32(len(validatorsInShard))
	}

	return numPubKeys
}
