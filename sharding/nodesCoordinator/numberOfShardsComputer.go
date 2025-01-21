package nodesCoordinator

import "github.com/multiversx/mx-chain-core-go/core"

type numberOfShardsWithMetaComputer struct {
}

func newNumberOfShardsWithMetaComputer() *numberOfShardsWithMetaComputer {
	return &numberOfShardsWithMetaComputer{}
}

// ComputeNumberOfShards computes the number of shards for the given nodes config
func (nsc *numberOfShardsWithMetaComputer) ComputeNumberOfShards(config *epochNodesConfig) (nbShards uint32, err error) {
	nbShards = uint32(len(config.eligibleMap))
	if nbShards < 2 {
		return 0, ErrInvalidNumberOfShards
	}
	if _, ok := config.eligibleMap[core.MetachainShardId]; !ok {
		return 0, ErrMetachainShardIdNotFound
	}
	// shards without metachain shard
	return nbShards - 1, nil
}

// ShardIdFromNodesConfig returns the shard id from nodes config
func (snsc *numberOfShardsWithMetaComputer) ShardIdFromNodesConfig(nodesConfig *epochNodesConfig) uint32 {
	return nodesConfig.shardID
}

// IsInterfaceNil returns true if there is no value under the interface
func (nsc *numberOfShardsWithMetaComputer) IsInterfaceNil() bool {
	return nsc == nil
}
