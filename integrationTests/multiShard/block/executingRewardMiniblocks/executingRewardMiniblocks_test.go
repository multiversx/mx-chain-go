package executingRewardMiniblocks

import (
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	testBlock "github.com/ElrondNetwork/elrond-go/integrationTests/multiShard/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
)

func getLeaderPercentage(node *integrationTests.TestProcessorNode) float64 {
	return node.EconomicsData.LeaderPercentage()
}

func TestExecuteBlocksWithTransactionsAndCheckRewards(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodesPerShard := 4
	nbMetaNodes := 2
	nbShards := 2
	consensusGroupSize := 2

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap(0)

	seedAddress := integrationTests.GetConnectableAddress(advertiser)

	// create map of shard - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinator(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		consensusGroupSize,
		seedAddress,
	)

	maxGasLimitPerBlock := uint64(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(100)
	valToTransfer := big.NewInt(100)
	nbTxsPerShard := uint32(100)
	mintValue := big.NewInt(1000000)

	for _, nodes := range nodesMap {
		integrationTests.SetEconomicsParameters(nodes, maxGasLimitPerBlock, gasPrice, gasLimit)
		integrationTests.DisplayAndStartNodes(nodes)
	}

	defer func() {
		_ = advertiser.Close()
		for _, nodes := range nodesMap {
			for _, n := range nodes {
				_ = n.Messenger.Close()
			}
		}
	}()

	integrationTests.GenerateIntraShardTransactions(nodesMap, nbTxsPerShard, mintValue, valToTransfer, gasPrice, gasLimit)

	round := uint64(1)
	nonce := uint64(1)
	nbBlocksProduced := 7

	var headers map[uint32]data.HeaderHandler
	var consensusNodes map[uint32][]*integrationTests.TestProcessorNode
	mapRewardsForShardAddresses := make(map[string]uint32)
	mapRewardsForMetachainAddresses := make(map[string]uint32)
	nbTxsForLeaderAddress := make(map[string]uint32)

	for i := 0; i < nbBlocksProduced; i++ {
		for _, nodes := range nodesMap {
			integrationTests.UpdateRound(nodes, round)
		}
		_, headers, consensusNodes = integrationTests.AllShardsProposeBlock(round, nonce, nodesMap)

		for shardId := range consensusNodes {
			addrRewards := make([]string, 0)
			updateExpectedRewards(mapRewardsForShardAddresses, addrRewards)
			nbTxs := getTransactionsFromHeaderInShard(t, headers, shardId)
			if len(addrRewards) > 0 {
				updateNumberTransactionsProposed(t, nbTxsForLeaderAddress, addrRewards[0], nbTxs)
			}
		}

		updateRewardsForMetachain(mapRewardsForMetachainAddresses, consensusNodes[0][0])

		indexesProposers := testBlock.GetBlockProposersIndexes(consensusNodes, nodesMap)
		integrationTests.SyncAllShardsWithRoundBlock(t, nodesMap, indexesProposers, round)
		round++
		nonce++
	}

	time.Sleep(5 * time.Second)

	verifyRewardsForShards(t, nodesMap, mapRewardsForShardAddresses, nbTxsForLeaderAddress, gasPrice, gasLimit)
	verifyRewardsForMetachain(t, mapRewardsForMetachainAddresses, nodesMap)
}

func TestExecuteBlocksWithTransactionsWhichReachedGasLimitAndCheckRewards(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodesPerShard := 2
	nbMetaNodes := 2
	nbShards := 1
	consensusGroupSize := 2

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap(0)

	seedAddress := integrationTests.GetConnectableAddress(advertiser)

	// create map of shard - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinator(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		consensusGroupSize,
		seedAddress,
	)

	maxGasLimitPerBlock := uint64(100000)
	gasPrice := uint64(10)
	gasLimit := uint64(99990)
	valToTransfer := big.NewInt(100)
	nbTxsPerShard := uint32(2)
	mintValue := big.NewInt(1000000)

	for _, nodes := range nodesMap {
		integrationTests.SetEconomicsParameters(nodes, maxGasLimitPerBlock, gasPrice, gasLimit)
		integrationTests.DisplayAndStartNodes(nodes)
	}

	defer func() {
		_ = advertiser.Close()
		for _, nodes := range nodesMap {
			for _, n := range nodes {
				_ = n.Messenger.Close()
			}
		}
	}()

	integrationTests.GenerateIntraShardTransactions(nodesMap, nbTxsPerShard, mintValue, valToTransfer, gasPrice, gasLimit)

	round := uint64(1)
	nonce := uint64(1)
	nbBlocksProduced := 2

	var headers map[uint32]data.HeaderHandler
	var consensusNodes map[uint32][]*integrationTests.TestProcessorNode
	mapRewardsForShardAddresses := make(map[string]uint32)
	nbTxsForLeaderAddress := make(map[string]uint32)

	for i := 0; i < nbBlocksProduced; i++ {
		_, headers, consensusNodes = integrationTests.AllShardsProposeBlock(round, nonce, nodesMap)

		for shardId := range consensusNodes {
			addrRewards := make([]string, 0)
			updateExpectedRewards(mapRewardsForShardAddresses, addrRewards)
			nbTxs := getTransactionsFromHeaderInShard(t, headers, shardId)
			if len(addrRewards) > 0 {
				updateNumberTransactionsProposed(t, nbTxsForLeaderAddress, addrRewards[0], nbTxs)
			}
		}

		for _, nodes := range nodesMap {
			integrationTests.UpdateRound(nodes, round)
		}
		indexesProposers := testBlock.GetBlockProposersIndexes(consensusNodes, nodesMap)
		integrationTests.SyncAllShardsWithRoundBlock(t, nodesMap, indexesProposers, round)
		round++
		nonce++
	}

	verifyRewardsForShards(t, nodesMap, mapRewardsForShardAddresses, nbTxsForLeaderAddress, gasPrice, gasLimit)
}

func TestExecuteBlocksWithoutTransactionsAndCheckRewards(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodesPerShard := 4
	nbMetaNodes := 2
	nbShards := 2
	consensusGroupSize := 2

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap(0)

	seedAddress := integrationTests.GetConnectableAddress(advertiser)

	// create map of shard - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinator(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		consensusGroupSize,
		seedAddress,
	)

	for _, nodes := range nodesMap {
		integrationTests.DisplayAndStartNodes(nodes)
	}

	defer func() {
		_ = advertiser.Close()
		for _, nodes := range nodesMap {
			for _, n := range nodes {
				_ = n.Messenger.Close()
			}
		}
	}()

	round := uint64(1)
	nonce := uint64(1)
	nbBlocksProduced := 7

	var consensusNodes map[uint32][]*integrationTests.TestProcessorNode
	mapRewardsForShardAddresses := make(map[string]uint32)
	mapRewardsForMetachainAddresses := make(map[string]uint32)
	nbTxsForLeaderAddress := make(map[string]uint32)

	for i := 0; i < nbBlocksProduced; i++ {
		_, _, consensusNodes = integrationTests.AllShardsProposeBlock(round, nonce, nodesMap)

		for shardId := range consensusNodes {
			if shardId == core.MetachainShardId {
				continue
			}

			shardRewardsData := &data.ConsensusRewardData{}
			addrRewards := shardRewardsData.Addresses
			updateExpectedRewards(mapRewardsForShardAddresses, addrRewards)
		}

		updateRewardsForMetachain(mapRewardsForMetachainAddresses, consensusNodes[0][0])

		for _, nodes := range nodesMap {
			integrationTests.UpdateRound(nodes, round)
		}
		indexesProposers := testBlock.GetBlockProposersIndexes(consensusNodes, nodesMap)
		integrationTests.SyncAllShardsWithRoundBlock(t, nodesMap, indexesProposers, round)
		round++
		nonce++
	}

	time.Sleep(4 * time.Second)

	verifyRewardsForShards(t, nodesMap, mapRewardsForShardAddresses, nbTxsForLeaderAddress, 0, 0)
	verifyRewardsForMetachain(t, mapRewardsForMetachainAddresses, nodesMap)
}

func getTransactionsFromHeaderInShard(t *testing.T, headers map[uint32]data.HeaderHandler, shardId uint32) uint32 {
	if shardId == core.MetachainShardId {
		return 0
	}

	header, ok := headers[shardId]
	if !ok {
		return 0
	}

	hdr, ok := header.(*block.Header)
	if !ok {
		assert.Error(t, process.ErrWrongTypeAssertion)
	}

	nbTxs := uint32(0)
	for _, mb := range hdr.MiniBlockHeaders {
		if mb.SenderShardID == shardId && mb.Type == block.TxBlock {
			nbTxs += mb.TxCount
		}
	}

	return nbTxs
}

func updateExpectedRewards(rewardsForAddress map[string]uint32, addresses []string) {
	for i := 0; i < len(addresses); i++ {
		if addresses[i] == "" {
			continue
		}

		rewardsForAddress[addresses[i]]++
	}
}

func updateNumberTransactionsProposed(
	t *testing.T,
	transactionsForLeader map[string]uint32,
	addressProposer string,
	nbTransactions uint32,
) {
	if addressProposer == "" {
		assert.Error(t, errors.New("invalid address"))
	}

	transactionsForLeader[addressProposer] += nbTransactions
}

func updateRewardsForMetachain(_ map[string]uint32, _ *integrationTests.TestProcessorNode) {
}

func verifyRewardsForMetachain(
	t *testing.T,
	mapRewardsForMeta map[string]uint32,
	nodes map[uint32][]*integrationTests.TestProcessorNode,
) {
	rewardValue := big.NewInt(0)

	for metaAddr, numOfTimesRewarded := range mapRewardsForMeta {
		acc, err := nodes[0][0].AccntState.GetExistingAccount([]byte(metaAddr))
		assert.Nil(t, err)

		expectedBalance := big.NewInt(0).SetUint64(uint64(numOfTimesRewarded))
		expectedBalance.Mul(expectedBalance, rewardValue)
		assert.Equal(t, expectedBalance, acc.(state.UserAccountHandler).GetBalance())
	}
}

func verifyRewardsForShards(
	t *testing.T,
	nodesMap map[uint32][]*integrationTests.TestProcessorNode,
	mapRewardsForAddress map[string]uint32,
	nbTxsForLeaderAddress map[string]uint32,
	gasPrice uint64,
	gasLimit uint64,
) {
	rewardValue := big.NewInt(0)
	feePerTxForLeader := float64(gasPrice) * float64(gasLimit) * getLeaderPercentage(nodesMap[0][0])

	for address, nbRewards := range mapRewardsForAddress {
		shard := nodesMap[0][0].ShardCoordinator.ComputeId([]byte(address))

		for _, shardNode := range nodesMap[shard] {
			acc, err := shardNode.AccntState.GetExistingAccount([]byte(address))
			assert.Nil(t, err)

			nbProposedTxs := nbTxsForLeaderAddress[address]
			expectedBalance := big.NewInt(0).SetUint64(uint64(nbRewards))
			expectedBalance.Mul(expectedBalance, rewardValue)
			totalFees := big.NewInt(0).SetUint64(uint64(nbProposedTxs))
			totalFees.Mul(totalFees, big.NewInt(0).SetUint64(uint64(feePerTxForLeader)))

			expectedBalance.Add(expectedBalance, totalFees)
			fmt.Println(fmt.Sprintf("checking account %s has balance %d", acc.AddressBytes(), expectedBalance))
			assert.Equal(t, expectedBalance, acc.(state.UserAccountHandler).GetBalance())
		}
	}
}
