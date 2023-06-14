package epochChangeWithNodesShuffling

import (
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/multiShard/endOfEpoch"
	"github.com/multiversx/mx-chain-go/testscommon/shardingmock"
	logger "github.com/multiversx/mx-chain-logger-go"
)

func TestEpochChangeWithCustomFinality(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	err := logger.SetLogLevel("*:DEBUG")
	if err != nil {
		return
	}
	nodesPerShard := 3
	nbMetaNodes := 2
	nbShards := 2
	consensusGroupSize := 2
	maxGasLimitPerBlock := uint64(100000)

	baseChainParameters := config.ChainParametersByEpochConfig{
		RoundDuration:               0,
		Hysteresis:                  0,
		EnableEpoch:                 0,
		ShardConsensusGroupSize:     uint32(consensusGroupSize),
		ShardMinNumNodes:            uint32(consensusGroupSize),
		MetachainConsensusGroupSize: uint32(consensusGroupSize),
		MetachainMinNumNodes:        uint32(consensusGroupSize),
		ShardFinality:               1,
		MetaFinality:                1,
	}

	chainParametersHandler := &shardingmock.ChainParametersHandlerStub{
		ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
			finality := baseChainParameters.ShardFinality
			if epoch < 1 {
				finality = 1
			} else if epoch < 2 {
				finality = 3
			} else {
				finality = 2
			}

			chainParamsCopy := baseChainParameters
			chainParamsCopy.ShardFinality = finality
			chainParamsCopy.MetaFinality = finality
			return chainParamsCopy, nil
		},
	}

	// create map of shard - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinatorAndChainParametersHandler(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		consensusGroupSize,
		chainParametersHandler,
	)

	gasPrice := uint64(10)
	gasLimit := uint64(100)
	valToTransfer := big.NewInt(100)
	nbTxsPerShard := uint32(100)
	mintValue := big.NewInt(1000000)

	defer func() {
		for _, nodes := range nodesMap {
			for _, n := range nodes {
				n.Close()
			}
		}
	}()

	roundsPerEpoch := uint64(10)
	for _, nodes := range nodesMap {
		integrationTests.SetEconomicsParameters(nodes, maxGasLimitPerBlock, gasPrice, gasLimit)
		integrationTests.DisplayAndStartNodes(nodes)
		for _, node := range nodes {
			node.EpochStartTrigger.SetRoundsPerEpoch(roundsPerEpoch)
		}
	}

	integrationTests.GenerateIntraShardTransactions(nodesMap, nbTxsPerShard, mintValue, valToTransfer, gasPrice, gasLimit)

	round := uint64(1)
	nonce := uint64(1)
	nbBlocksToProduce := uint64(37)
	expectedLastEpoch := uint32(nbBlocksToProduce / roundsPerEpoch)
	var consensusNodes map[uint32][]*integrationTests.TestProcessorNode

	for i := uint64(0); i < nbBlocksToProduce; i++ {
		for _, nodes := range nodesMap {
			integrationTests.UpdateRound(nodes, round)
		}

		_, _, consensusNodes = integrationTests.AllShardsProposeBlock(round, nonce, nodesMap)
		indexesProposers := endOfEpoch.GetBlockProposersIndexes(consensusNodes, nodesMap)
		integrationTests.SyncAllShardsWithRoundBlock(t, nodesMap, indexesProposers, round)
		round++
		nonce++

		time.Sleep(integrationTests.StepDelay)
	}

	for _, nodes := range nodesMap {
		endOfEpoch.VerifyThatNodesHaveCorrectEpoch(t, expectedLastEpoch, nodes)
	}
}
