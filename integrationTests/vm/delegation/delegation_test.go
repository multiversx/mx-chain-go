package delegation

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/multiShard/endOfEpoch"
	vm2 "github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/stretchr/testify/assert"
)

func TestDelegationSystemSCWithValidatorStatistics(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2
	shardConsensusGroupSize := 1
	metaConsensusGroupSize := 1

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	nodesMap := integrationTests.CreateNodesWithNodesCoordinatorAndTxKeys(
		nodesPerShard,
		numMetachainNodes,
		numOfShards,
		shardConsensusGroupSize,
		metaConsensusGroupSize,
		integrationTests.GetConnectableAddress(advertiser),
	)

	nodes := make([]*integrationTests.TestProcessorNode, 0)
	idxProposers := make([]int, numOfShards+1)

	for _, nds := range nodesMap {
		nodes = append(nodes, nds...)
	}

	for _, nds := range nodesMap {
		idx, err := vm2.GetNodeIndex(nodes, nds[0])
		assert.Nil(t, err)

		idxProposers = append(idxProposers, idx)
	}
	integrationTests.DisplayAndStartNodes(nodes)

	roundsPerEpoch := uint64(5)
	for _, node := range nodes {
		node.EpochStartTrigger.SetRoundsPerEpoch(roundsPerEpoch)
	}

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	for _, node := range nodesMap {
		fmt.Println(integrationTests.MakeDisplayTable(node))
	}

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	rewardAccount := integrationTests.CreateTestWalletAccount(nodes[0].ShardCoordinator, 0)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	var txData string
	for _, node := range nodes {
		txData = "changeRewardAddress" + "@" + hex.EncodeToString(rewardAccount.Address)
		integrationTests.CreateAndSendTransaction(node, big.NewInt(0), vm.AuctionSCAddress, txData, integrationTests.AdditionalGasLimit)
	}

	nbBlocksToProduce := roundsPerEpoch * 3
	var consensusNodes map[uint32][]*integrationTests.TestProcessorNode

	for i := uint64(0); i < nbBlocksToProduce; i++ {
		for _, nodesSlice := range nodesMap {
			integrationTests.UpdateRound(nodesSlice, round)
			integrationTests.AddSelfNotarizedHeaderByMetachain(nodesSlice)
		}

		_, _, consensusNodes = integrationTests.AllShardsProposeBlock(round, nonce, nodesMap)
		indexesProposers := endOfEpoch.GetBlockProposersIndexes(consensusNodes, nodesMap)
		integrationTests.SyncAllShardsWithRoundBlock(t, nodesMap, indexesProposers, round)
		round++
		nonce++

		time.Sleep(1 * time.Second)
	}
}
