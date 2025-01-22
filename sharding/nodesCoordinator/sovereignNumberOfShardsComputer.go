package nodesCoordinator

import "github.com/multiversx/mx-chain-core-go/core"

type sovereignNumberOfShardsComputer struct {
}

func newSovereignNumberOfShardsComputer() *sovereignNumberOfShardsComputer {
	return &sovereignNumberOfShardsComputer{}
}

// ComputeNumberOfShards computes the number of shards for the given nodes config
func (snsc *sovereignNumberOfShardsComputer) ComputeNumberOfShards(config *epochNodesConfig) (nbShards uint32, err error) {
	nbShards = uint32(len(config.eligibleMap))
	if nbShards != 1 {
		return 0, ErrInvalidNumberOfShards
	}
	if _, ok := config.eligibleMap[core.SovereignChainShardId]; !ok {
		return 0, ErrInvalidSovereignChainShardId
	}

	return nbShards, nil
}

// ShardIdFromNodesConfig always returns sovereign shard id
func (snsc *sovereignNumberOfShardsComputer) ShardIdFromNodesConfig(_ *epochNodesConfig) uint32 {
	return core.SovereignChainShardId
}

// IsInterfaceNil returns true if there is no value under the interface
func (snsc *sovereignNumberOfShardsComputer) IsInterfaceNil() bool {
	return snsc == nil
}
