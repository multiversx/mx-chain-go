package staking

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
)

// NewTestMetaProcessor -
func NewTestMetaProcessor(
	numOfMetaNodes uint32,
	numOfShards uint32,
	numOfEligibleNodesPerShard uint32,
	numOfWaitingNodesPerShard uint32,
	numOfNodesToShufflePerShard uint32,
	shardConsensusGroupSize int,
	metaConsensusGroupSize int,
	numOfNodesInStakingQueue uint32,
) *TestMetaProcessor {
	coreComponents, dataComponents, bootstrapComponents, statusComponents, stateComponents := createComponentHolders(numOfShards)

	maxNodesConfig := createMaxNodesConfig(
		numOfMetaNodes,
		numOfShards,
		numOfEligibleNodesPerShard,
		numOfWaitingNodesPerShard,
		numOfNodesToShufflePerShard,
	)

	queue := createStakingQueue(
		numOfNodesInStakingQueue,
		maxNodesConfig[0].MaxNumNodes,
		coreComponents.InternalMarshalizer(),
		stateComponents.AccountsAdapter(),
	)

	eligibleMap, waitingMap := createGenesisNodes(
		numOfMetaNodes,
		numOfShards,
		numOfEligibleNodesPerShard,
		numOfWaitingNodesPerShard,
		coreComponents.InternalMarshalizer(),
		stateComponents,
	)

	nc := createNodesCoordinator(
		eligibleMap,
		waitingMap,
		numOfMetaNodes,
		numOfShards,
		numOfEligibleNodesPerShard,
		shardConsensusGroupSize,
		metaConsensusGroupSize,
		coreComponents,
		dataComponents.StorageService().GetStorer(dataRetriever.BootstrapUnit),
		bootstrapComponents.NodesCoordinatorRegistryFactory(),
		maxNodesConfig,
	)

	return newTestMetaProcessor(
		coreComponents,
		dataComponents,
		bootstrapComponents,
		statusComponents,
		stateComponents,
		nc,
		maxNodesConfig,
		queue,
	)
}

func createMaxNodesConfig(
	numOfMetaNodes uint32,
	numOfShards uint32,
	numOfEligibleNodesPerShard uint32,
	numOfWaitingNodesPerShard uint32,
	numOfNodesToShufflePerShard uint32,
) []config.MaxNodesChangeConfig {
	totalEligible := numOfMetaNodes + numOfShards*numOfEligibleNodesPerShard
	totalWaiting := (numOfShards + 1) * numOfWaitingNodesPerShard
	totalNodes := totalEligible + totalWaiting

	maxNodesConfig := make([]config.MaxNodesChangeConfig, 0)
	maxNodesConfig = append(maxNodesConfig, config.MaxNodesChangeConfig{
		EpochEnable:            0,
		MaxNumNodes:            totalNodes,
		NodesToShufflePerShard: numOfNodesToShufflePerShard,
	},
	)

	maxNodesConfig = append(maxNodesConfig, config.MaxNodesChangeConfig{
		EpochEnable:            stakingV4DistributeAuctionToWaitingEpoch,
		MaxNumNodes:            totalNodes - numOfNodesToShufflePerShard*(numOfShards+1),
		NodesToShufflePerShard: numOfNodesToShufflePerShard,
	},
	)

	return maxNodesConfig
}
