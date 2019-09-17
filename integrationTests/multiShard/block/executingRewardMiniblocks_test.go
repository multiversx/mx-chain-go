package block

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestExecuteBlocksWithOnlyRewards(t *testing.T) {
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

	randomness := generateInitialRandomness(uint32(nbShards))
	round := uint64(1)
	nonce := uint64(1)
	nbBlocksProduced := 7
	nbTxsPerShard := 100
	mintValue := big.NewInt(1000000)

	var headers map[uint32]data.HeaderHandler
	var consensusNodes map[uint32][]*integrationTests.TestProcessorNode
	mapRewardsForAddress := make(map[string]uint32)
	nbTxsForLeaderAddress := make(map[string]uint32)

	rewardValue := 1000
	gasPrice := uint64(10)
	gasLimit := uint64(100)
	feePerTxForLeader := gasPrice * gasLimit / 2
	valToTransfer := big.NewInt(100)

	sendersPrivateKeys := integrationTests.CreateAndSendIntraShardTransactions(
		nodesMap,
		nbTxsPerShard,
		gasPrice,
		gasLimit,
		valToTransfer,
	)

	for shardId, nodes := range nodesMap {
		if shardId == sharding.MetachainShardId {
			continue
		}

		fmt.Println("Minting sender addresses...")
		integrationTests.CreateMintingForSenders(
			nodes,
			shardId,
			sendersPrivateKeys[shardId],
			mintValue,
		)
	}

	for i := 0; i < nbBlocksProduced; i++ {
		_, headers, consensusNodes, randomness = integrationTests.AllShardsProposeBlock(round, nonce, randomness, nodesMap)

		for shardId, consensusGroup := range consensusNodes {
			addrRewards := consensusGroup[0].SpecialAddressHandler.ConsensusRewardAddresses()
			updateExpectedRewards(mapRewardsForAddress, addrRewards)
			nbTxs := transactionsFromHeaderInShard(t, headers, shardId)

			// without metachain nodes for now
			if len(addrRewards) > 0 {
				updateNbTransactionsProposed(t, nbTxsForLeaderAddress, addrRewards[0], nbTxs)
			}
		}

		indexesProposers := getBlockProposersIndexes(consensusNodes, nodesMap)
		integrationTests.VerifyNodesHaveHeaders(t, headers, nodesMap)
		integrationTests.SyncAllShardsWithRoundBlock(t, nodesMap, indexesProposers, round)
		round++
		nonce++
	}

	time.Sleep(time.Second)

	for address, nbRewards := range mapRewardsForAddress {
		addrContainer, _ := integrationTests.TestAddressConverter.CreateAddressFromPublicKeyBytes([]byte(address))
		shard := nodesMap[0][0].ShardCoordinator.ComputeId(addrContainer)

		for _, shardNode := range nodesMap[shard] {
			acc, err := shardNode.AccntState.GetExistingAccount(addrContainer)
			assert.Nil(t, err)

			nbProposedTxs := nbTxsForLeaderAddress[address]
			expectedBalance := int64(nbRewards)*int64(rewardValue) + int64(nbProposedTxs)*int64(feePerTxForLeader)
			fmt.Println(fmt.Sprintf("checking account %s has balance %d", core.ToB64(acc.AddressContainer().Bytes()), expectedBalance))
			assert.Equal(t, big.NewInt(expectedBalance), acc.(*state.Account).Balance)
		}
	}
}

func getBlockProposersIndexes(
	consensusMap map[uint32][]*integrationTests.TestProcessorNode,
	nodesMap map[uint32][]*integrationTests.TestProcessorNode,
) map[uint32]int {

	indexProposer := make(map[uint32]int)

	for sh, testNodeList := range nodesMap {
		for k, testNode := range testNodeList {
			if reflect.DeepEqual(consensusMap[sh][0], testNode) {
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

func transactionsFromHeaderInShard(t *testing.T, headers map[uint32]data.HeaderHandler, shardId uint32) uint32 {
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
		currentRewards, ok := rewardsForAddress[addresses[i]]
		if !ok {
			currentRewards = 0
		}

		rewardsForAddress[addresses[i]] = currentRewards + 1
	}
}

func updateNbTransactionsProposed(
	t *testing.T,
	transactionsForLeader map[string]uint32,
	addressProposer string,
	nbTransactions uint32,
) {
	if addressProposer == "" {
		assert.Error(t, errors.New("invalid address"))
	}

	proposedTransactions, ok := transactionsForLeader[addressProposer]
	if !ok {
		proposedTransactions = 0
	}

	transactionsForLeader[addressProposer] = proposedTransactions + nbTransactions
}
