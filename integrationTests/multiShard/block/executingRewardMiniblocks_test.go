package block

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func getRewardValue(node *integrationTests.TestProcessorNode) *big.Int {
	return node.EconomicsData.RewardsValue()
}

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

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

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
				_ = n.Node.Stop()
			}
		}
	}()

	generateIntraShardTransactions(nodesMap, nbTxsPerShard, mintValue, valToTransfer, gasPrice, gasLimit)

	round := uint64(1)
	nonce := uint64(1)
	nbBlocksProduced := 7

	randomness := generateInitialRandomness(uint32(nbShards))
	var headers map[uint32]data.HeaderHandler
	var consensusNodes map[uint32][]*integrationTests.TestProcessorNode
	mapRewardsForShardAddresses := make(map[string]uint32)
	mapRewardsForMetachainAddresses := make(map[string]uint32)
	nbTxsForLeaderAddress := make(map[string]uint32)

	for i := 0; i < nbBlocksProduced; i++ {
		_, headers, consensusNodes, randomness = integrationTests.AllShardsProposeBlock(round, nonce, randomness, nodesMap)

		for shardId, consensusGroup := range consensusNodes {
			shardRewardData := consensusGroup[0].SpecialAddressHandler.ConsensusShardRewardData()
			addrRewards := shardRewardData.Addresses
			updateExpectedRewards(mapRewardsForShardAddresses, addrRewards)
			nbTxs := getTransactionsFromHeaderInShard(t, headers, shardId)
			if len(addrRewards) > 0 {
				updateNumberTransactionsProposed(t, nbTxsForLeaderAddress, addrRewards[0], nbTxs)
			}
		}

		updateRewardsForMetachain(mapRewardsForMetachainAddresses, consensusNodes[0][0])

		indexesProposers := getBlockProposersIndexes(consensusNodes, nodesMap)
		integrationTests.VerifyNodesHaveHeaders(t, headers, nodesMap)
		integrationTests.SyncAllShardsWithRoundBlock(t, nodesMap, indexesProposers, round)
		round++
		nonce++
	}

	time.Sleep(4 * time.Second)

	verifyRewardsForShards(t, nodesMap, mapRewardsForShardAddresses, nbTxsForLeaderAddress, gasPrice, gasLimit)
	verifyRewardsForMetachain(t, mapRewardsForMetachainAddresses, nodesMap)
}

func TestExecuteBlocksWithoutTransactionsAndCheckRewards(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodesPerShard := 4
	nbMetaNodes := 2
	nbShards := 2
	consensusGroupSize := 2

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

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
				_ = n.Node.Stop()
			}
		}
	}()

	round := uint64(1)
	nonce := uint64(1)
	nbBlocksProduced := 7

	randomness := generateInitialRandomness(uint32(nbShards))
	var headers map[uint32]data.HeaderHandler
	var consensusNodes map[uint32][]*integrationTests.TestProcessorNode
	mapRewardsForShardAddresses := make(map[string]uint32)
	mapRewardsForMetachainAddresses := make(map[string]uint32)
	nbTxsForLeaderAddress := make(map[string]uint32)

	for i := 0; i < nbBlocksProduced; i++ {
		_, headers, consensusNodes, randomness = integrationTests.AllShardsProposeBlock(round, nonce, randomness, nodesMap)

		for shardId, consensusGroup := range consensusNodes {
			if shardId == sharding.MetachainShardId {
				continue
			}

			shardRewardsData := consensusGroup[0].SpecialAddressHandler.ConsensusShardRewardData()
			if shardRewardsData == nil {
				shardRewardsData = &data.ConsensusRewardData{}
			}

			addrRewards := shardRewardsData.Addresses
			updateExpectedRewards(mapRewardsForShardAddresses, addrRewards)
		}

		updateRewardsForMetachain(mapRewardsForMetachainAddresses, consensusNodes[0][0])

		indexesProposers := getBlockProposersIndexes(consensusNodes, nodesMap)
		integrationTests.VerifyNodesHaveHeaders(t, headers, nodesMap)
		integrationTests.SyncAllShardsWithRoundBlock(t, nodesMap, indexesProposers, round)
		round++
		nonce++
	}

	time.Sleep(4 * time.Second)

	verifyRewardsForShards(t, nodesMap, mapRewardsForShardAddresses, nbTxsForLeaderAddress, 0, 0)
	verifyRewardsForMetachain(t, mapRewardsForMetachainAddresses, nodesMap)
}

func generateIntraShardTransactions(
	nodesMap map[uint32][]*integrationTests.TestProcessorNode,
	nbTxsPerShard uint32,
	mintValue *big.Int,
	valToTransfer *big.Int,
	gasPrice uint64,
	gasLimit uint64,
) {
	sendersPrivateKeys := make(map[uint32][]crypto.PrivateKey)
	receiversPublicKeys := make(map[uint32][]crypto.PublicKey)

	for shardId, nodes := range nodesMap {
		if shardId == sharding.MetachainShardId {
			continue
		}

		sendersPrivateKeys[shardId], receiversPublicKeys[shardId] = integrationTests.CreateSendersAndReceiversInShard(
			nodes[0],
			nbTxsPerShard,
		)

		fmt.Println("Minting sender addresses...")
		integrationTests.CreateMintingForSenders(
			nodes,
			shardId,
			sendersPrivateKeys[shardId],
			mintValue,
		)
	}

	integrationTests.CreateAndSendTransactions(
		nodesMap,
		sendersPrivateKeys,
		receiversPublicKeys,
		gasPrice,
		gasLimit,
		valToTransfer,
	)
}

func getBlockProposersIndexes(
	consensusMap map[uint32][]*integrationTests.TestProcessorNode,
	nodesMap map[uint32][]*integrationTests.TestProcessorNode,
) map[uint32]int {

	indexProposer := make(map[uint32]int)

	for sh, testNodeList := range nodesMap {
		for k, testNode := range testNodeList {
			if consensusMap[sh][0] == testNode {
				indexProposer[sh] = k
			}
		}
	}

	return indexProposer
}

func generateInitialRandomness(nbShards uint32) map[uint32][]byte {
	randomness := make(map[uint32][]byte)

	for i := uint32(0); i < nbShards; i++ {
		randomness[i] = []byte("root hash")
	}

	randomness[sharding.MetachainShardId] = []byte("root hash")

	return randomness
}

func getTransactionsFromHeaderInShard(t *testing.T, headers map[uint32]data.HeaderHandler, shardId uint32) uint32 {
	if shardId == sharding.MetachainShardId {
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

func updateRewardsForMetachain(rewardsMap map[string]uint32, consensusNode *integrationTests.TestProcessorNode) {
	metaRewardDataSlice := consensusNode.SpecialAddressHandler.ConsensusMetaRewardData()
	if len(metaRewardDataSlice) > 0 {
		for _, metaRewardData := range metaRewardDataSlice {
			for _, addr := range metaRewardData.Addresses {
				rewardsMap[addr]++
			}
		}
	}
}

func verifyRewardsForMetachain(
	t *testing.T,
	mapRewardsForMeta map[string]uint32,
	nodes map[uint32][]*integrationTests.TestProcessorNode,
) {
	rewardValue := getRewardValue(nodes[0][0])

	for metaAddr, numOfTimesRewarded := range mapRewardsForMeta {
		addrContainer, _ := integrationTests.TestAddressConverter.CreateAddressFromPublicKeyBytes([]byte(metaAddr))
		acc, err := nodes[0][0].AccntState.GetExistingAccount(addrContainer)
		assert.Nil(t, err)

		expectedBalance := big.NewInt(0).SetUint64(uint64(numOfTimesRewarded))
		expectedBalance.Mul(expectedBalance, rewardValue)
		assert.Equal(t, expectedBalance, acc.(*state.Account).Balance)
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
	rewardValue := getRewardValue(nodesMap[0][0])
	feePerTxForLeader := float64(gasPrice) * float64(gasLimit) * getLeaderPercentage(nodesMap[0][0])

	for address, nbRewards := range mapRewardsForAddress {
		addrContainer, _ := integrationTests.TestAddressConverter.CreateAddressFromPublicKeyBytes([]byte(address))
		shard := nodesMap[0][0].ShardCoordinator.ComputeId(addrContainer)

		for _, shardNode := range nodesMap[shard] {
			acc, err := shardNode.AccntState.GetExistingAccount(addrContainer)
			assert.Nil(t, err)

			nbProposedTxs := nbTxsForLeaderAddress[address]
			expectedBalance := big.NewInt(0).SetUint64(uint64(nbRewards))
			expectedBalance.Mul(expectedBalance, rewardValue)
			totalFees := big.NewInt(0).SetUint64(uint64(nbProposedTxs))
			totalFees.Mul(totalFees, big.NewInt(0).SetUint64(uint64(feePerTxForLeader)))

			expectedBalance.Add(expectedBalance, totalFees)
			fmt.Println(fmt.Sprintf("checking account %s has balance %d", core.ToB64(acc.AddressContainer().Bytes()), expectedBalance))
			assert.Equal(t, expectedBalance, acc.(*state.Account).Balance)
		}
	}
}
