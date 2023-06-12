package sovereign

import (
	"math/big"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
)

func CreateGeneralSovereignSetup(initialBalance *big.Int) (
	[]*integrationTests.TestProcessorNode,
	[]int,
	[]*integrationTests.TestWalletAccount,
) {
	numOfShards := 1
	nodesPerShard := 3

	enableEpochs := config.EnableEpochs{
		OptimizeGasUsedInCrossMiniBlocksEnableEpoch: integrationTests.UnreachableEpoch,
		ScheduledMiniBlocksEnableEpoch:              integrationTests.UnreachableEpoch,
		MiniBlockPartialExecutionEnableEpoch:        integrationTests.UnreachableEpoch,
	}

	nodes := CreateNodesWithEnableEpochs(
		numOfShards,
		nodesPerShard,
		enableEpochs,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	integrationTests.MintAllNodes(nodes, initialBalance)

	numPlayers := 10
	players := make([]*integrationTests.TestWalletAccount, numPlayers)
	for i := 0; i < numPlayers; i++ {
		shardId := uint32(i) % nodes[0].ShardCoordinator.NumberOfShards()
		players[i] = integrationTests.CreateTestWalletAccount(nodes[0].ShardCoordinator, shardId)
	}

	integrationTests.MintAllPlayers(nodes, players, initialBalance)

	return nodes, idxProposers, players
}

// CreateNodesWithEnableEpochs creates multiple nodes with custom epoch config
func CreateNodesWithEnableEpochs(
	numOfShards int,
	nodesPerShard int,
	epochConfig config.EnableEpochs,
) []*integrationTests.TestProcessorNode {
	nodes := make([]*integrationTests.TestProcessorNode, numOfShards*nodesPerShard)
	connectableNodes := make([]integrationTests.Connectable, len(nodes))

	idx := 0
	for shardId := uint32(0); shardId < uint32(numOfShards); shardId++ {
		for j := 0; j < nodesPerShard; j++ {
			n := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
				MaxShards:            uint32(numOfShards),
				NodeShardId:          shardId,
				TxSignPrivKeyShardId: shardId,
				EpochsConfig:         &epochConfig,
				ChainRunType:         common.ChainRunTypeSovereign,
			})
			nodes[idx] = n
			connectableNodes[idx] = n
			idx++
		}
	}
	integrationTests.ConnectNodes(connectableNodes)

	return nodes
}
