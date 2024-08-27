package nodesCoordinator

type sovereignNumberOfShardsComputer struct {
}

// NewSovereignNumberOfShardsComputer creates a new instance of the number of shards computer for sovereign
func NewSovereignNumberOfShardsComputer() *sovereignNumberOfShardsComputer {
	return &sovereignNumberOfShardsComputer{}
}

// ComputeNumberOfShards computes the number of shards for the given nodes config
func (snsc *sovereignNumberOfShardsComputer) ComputeNumberOfShards(config *epochNodesConfig) (nbShards uint32, err error) {
	nbShards = uint32(len(config.eligibleMap))
	if nbShards != 1 {
		return 0, ErrInvalidNumberOfShards
	}
	return nbShards, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (snsc *sovereignNumberOfShardsComputer) IsInterfaceNil() bool {
	return snsc == nil
}
