package block

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("integrationTests/multishard/block")

// TestExecutingTransactionsFromRewardsFundsCrossShard tests the following scenario:
// A validator from shard 0 receives rewards from shard 1 (where it is assigned) and creates move balance
// transactions. All other shard peers can and will sync the blocks containing the move balance transactions.
func TestExecutingTransactionsFromRewardsFundsCrossShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)

	//it is important to have all combinations here as to test more edgecases
	mapAssignements := map[uint32][]uint32{
		0:                         {1, 0},
		1:                         {0, 1},
		sharding.MetachainShardId: {0, 1},
	}

	nodesMap, numShards := integrationTests.CreateProcessorNodesWithNodesCoordinator(mapAssignements, 1, 1, advertiserAddr)
	setRewardParametersToNodes(nodesMap)

	defer func() {
		_ = advertiser.Close()
		closeNodes(nodesMap)
	}()

	p2pBootstrapNodes(nodesMap)

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(stepDelay)

	round := uint64(0)
	nonce := uint64(1)
	round = integrationTests.IncrementAndPrintRound(round)
	randomness := generateInitialRandomness(numShards)

	senderShardID := uint32(0)
	receiver := nodesMap[sharding.MetachainShardId][0].OwnAccount.PkTxSign

	transferValue := integrationTests.MinTxGasLimit * integrationTests.MinTxGasPrice

	go func() {
		for {
			for _, n := range nodesMap[senderShardID] {
				generateAndSendTxs(
					n,
					transferValue,
					n.OwnAccount.SkTxSign,
					n.OwnAccount.Address,
					receiver)
			}
			time.Sleep(time.Second)
		}
	}()

	firstNode := nodesMap[senderShardID][0]
	numBlocksProduced := uint64(13)
	var consensusNodes map[uint32][]*integrationTests.TestProcessorNode
	for i := uint64(0); i < numBlocksProduced; i++ {
		printAccount(firstNode)

		for _, nodes := range nodesMap {
			integrationTests.UpdateRound(nodes, round)
		}
		_, _, consensusNodes, randomness = integrationTests.AllShardsProposeBlock(round, nonce, randomness, nodesMap)

		indexesProposers := getBlockProposersIndexes(consensusNodes, nodesMap)
		integrationTests.SyncAllShardsWithRoundBlock(t, nodesMap, indexesProposers, round)
		time.Sleep(stepDelay)

		round++
		nonce++

		checkSameBlockHeight(t, nodesMap)
	}

	printAccount(firstNode)
}

func setRewardParametersToNodes(nodesMap map[uint32][]*integrationTests.TestProcessorNode) {
	for _, nodes := range nodesMap {
		for _, n := range nodes {
			n.EconomicsData.SetRewards(big.NewInt(0).SetUint64(2 * integrationTests.MinTxGasLimit * integrationTests.MinTxGasPrice))
		}
	}
}

func closeNodes(nodesMap map[uint32][]*integrationTests.TestProcessorNode) {
	for _, shards := range nodesMap {
		for _, n := range shards {
			_ = n.Messenger.Close()
		}
	}
}

func p2pBootstrapNodes(nodesMap map[uint32][]*integrationTests.TestProcessorNode) {
	for _, shards := range nodesMap {
		for _, n := range shards {
			_ = n.Messenger.Bootstrap()
		}
	}
}

func checkSameBlockHeight(t *testing.T, nodesMap map[uint32][]*integrationTests.TestProcessorNode) {
	for _, nodes := range nodesMap {
		referenceBlock := nodes[0].BlockChain.GetCurrentBlockHeader()
		for _, n := range nodes {
			crtBlock := n.BlockChain.GetCurrentBlockHeader()
			//(crtBlock == nil) != (blkc == nil) actually does a XOR operation between the 2 conditions
			//as if the reference is nil, the same must be all other nodes. Same if the reference is not nil.
			require.False(t, (referenceBlock == nil) != (crtBlock == nil))
			if referenceBlock != nil {
				require.Equal(t, referenceBlock.GetNonce(), crtBlock.GetNonce())
			}
		}
	}
}

func printAccount(node *integrationTests.TestProcessorNode) {
	accnt, _ := node.AccntState.GetExistingAccount(node.OwnAccount.Address)
	if check.IfNil(accnt) {
		log.Info("account",
			"address", node.OwnAccount.Address.Bytes(),
			"nonce", "-",
			"balance", "-",
		)

		return
	}
	log.Info("account",
		"address", node.OwnAccount.Address.Bytes(),
		"nonce", accnt.GetNonce(),
		"balance", accnt.(*state.Account).Balance,
	)
}

func generateAndSendTxs(
	n *integrationTests.TestProcessorNode,
	transferValue uint64,
	sk crypto.PrivateKey,
	addr state.AddressContainer,
	pkReceiver crypto.PublicKey,
) {
	accnt, _ := n.AccntState.GetExistingAccount(addr)
	startingNonce := uint64(0)
	if accnt != nil {
		startingNonce = accnt.GetNonce()
	}

	numTxs := 1
	for i := 0; i < numTxs; i++ {
		nonce := startingNonce + uint64(i)

		tx := integrationTests.GenerateTransferTx(
			nonce,
			sk,
			pkReceiver,
			big.NewInt(0).SetUint64(transferValue),
			integrationTests.MinTxGasPrice,
			integrationTests.MinTxGasLimit,
		)

		txHash, _ := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, tx)

		_, _ = n.SendTransaction(tx)
		log.Info("send tx",
			"hash", txHash,
			"nonce", nonce,
			"value", transferValue,
		)
	}
}
