package components

import (
	"sort"

	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

// fixedShardNodesShuffler keeps genesis validators on the shard for which the simulator created
// their physical node. The regular intra-shard distributor already handles eligible/waiting
// shuffling, but staking-v4 auction re-entry uses a separate cross-shard distributor. A simulator
// node cannot restart in another shard, so genesis keys must remain pinned there as well. Keys added
// later through /simulator/add-keys are intentionally not pinned; every primary node manages them.
type fixedShardNodesShuffler struct {
	delegate    nodesCoordinator.NodesShuffler
	fixedShards map[string]uint32
}

func newFixedShardNodesShuffler(
	delegate nodesCoordinator.NodesShuffler,
	nodesSetup sharding.GenesisNodesSetupHandler,
) nodesCoordinator.NodesShuffler {
	fixedShards := make(map[string]uint32)
	eligible, waiting := nodesSetup.InitialNodesInfo()
	for shardID, validators := range eligible {
		for _, validator := range validators {
			fixedShards[string(validator.PubKeyBytes())] = shardID
		}
	}
	for shardID, validators := range waiting {
		for _, validator := range validators {
			fixedShards[string(validator.PubKeyBytes())] = shardID
		}
	}

	return &fixedShardNodesShuffler{
		delegate:    delegate,
		fixedShards: fixedShards,
	}
}

func (shuffler *fixedShardNodesShuffler) UpdateNodeLists(
	args nodesCoordinator.ArgsUpdateNodes,
) (*nodesCoordinator.ResUpdateNodes, error) {
	result, err := shuffler.delegate.UpdateNodeLists(args)
	if err != nil {
		return nil, err
	}

	result.Eligible = pinGenesisValidators(result.Eligible, shuffler.fixedShards)
	result.Waiting = pinGenesisValidators(result.Waiting, shuffler.fixedShards)
	result.ShuffledOut = pinGenesisValidators(result.ShuffledOut, shuffler.fixedShards)

	return result, nil
}

func pinGenesisValidators(
	validatorsByShard map[uint32][]nodesCoordinator.Validator,
	fixedShards map[string]uint32,
) map[uint32][]nodesCoordinator.Validator {
	pinned := make(map[uint32][]nodesCoordinator.Validator, len(validatorsByShard))
	shardIDs := make([]uint32, 0, len(validatorsByShard))
	for shardID := range validatorsByShard {
		shardIDs = append(shardIDs, shardID)
		pinned[shardID] = make([]nodesCoordinator.Validator, 0, len(validatorsByShard[shardID]))
	}
	sort.Slice(shardIDs, func(i, j int) bool {
		return shardIDs[i] < shardIDs[j]
	})

	// Nodes coordinator list order is consensus-critical. UpdateNodeLists runs independently on
	// every validator, so ranging over validatorsByShard directly would make validators moved back
	// from different source shards arrive in a process-local random order. That produces different
	// consensus groups after an epoch shuffle and stalls the chain.
	for _, shardID := range shardIDs {
		validators := validatorsByShard[shardID]
		for _, validator := range validators {
			targetShard := shardID
			if fixedShard, ok := fixedShards[string(validator.PubKey())]; ok {
				targetShard = fixedShard
			}
			pinned[targetShard] = append(pinned[targetShard], validator)
		}
	}

	return pinned
}

func (shuffler *fixedShardNodesShuffler) IsInterfaceNil() bool {
	return shuffler == nil
}
