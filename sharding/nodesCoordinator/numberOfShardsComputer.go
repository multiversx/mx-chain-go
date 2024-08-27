package nodesCoordinator

type numberOfShardsWithMetaComputer struct {
}

// NewNumberOfShardsWithMetaComputer creates a new instance of the number of shards computer with meta
func NewNumberOfShardsWithMetaComputer() *numberOfShardsWithMetaComputer {
	return &numberOfShardsWithMetaComputer{}
}

// ComputeNumberOfShards computes the number of shards for the given nodes config
func (nsc *numberOfShardsWithMetaComputer) ComputeNumberOfShards(config *epochNodesConfig) (nbShards uint32, err error) {
	nbShards = uint32(len(config.eligibleMap))
	if nbShards < 2 {
		return 0, ErrInvalidNumberOfShards
	}
	// shards without metachain shard
	return nbShards - 1, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (nsc *numberOfShardsWithMetaComputer) IsInterfaceNil() bool {
	return nsc == nil
}
