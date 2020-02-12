package longTests

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("integrationtests/singleshard/block")

// TestExecutingTransactionsFromGatheredRewards tests what happens in a setting where a proposer/validator
// periodically sends transactions to someone else. The tests starts with the node having 0 balance.
func TestExecutingTransactionsFromGatheredRewards(t *testing.T) {
	t.Skip("this is a long test")

	numOfNodes := 2
	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)

	_ = logger.SetLogLevel("*:DEBUG,process:TRACE")

	nodesMap := integrationTests.CreateProcessorNodesWithNodesCoordinator(numOfNodes, 1, 1, numOfNodes, 1, advertiserAddr)
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
	randomness := generateInitialRandomness(1)

	firstNode := nodesMap[0][0]

	transferValue := integrationTests.MinTxGasLimit * integrationTests.MinTxGasPrice / 2

	go func() {
		for {
			generateAndSendTxs(
				firstNode,
				transferValue,
				firstNode.OwnAccount.SkTxSign,
				firstNode.OwnAccount.Address,
				nodesMap[sharding.MetachainShardId][0].OwnAccount.PkTxSign)
			time.Sleep(time.Second)
		}
	}()

	numBlocksProduced := uint64(20)
	var consensusNodes map[uint32][]*integrationTests.TestProcessorNode
	for i := uint64(0); i < numBlocksProduced; i++ {
		printAccount(firstNode)

		for _, nodes := range nodesMap {
			integrationTests.UpdateRound(nodes, round)
		}
		_, _, consensusNodes, randomness = integrationTests.AllShardsProposeBlock(round, nonce, randomness, nodesMap)

		indexesProposers := getBlockProposersIndexes(consensusNodes, nodesMap)
		integrationTests.SyncAllShardsWithRoundBlock(t, nodesMap, indexesProposers, round)
		round++
		nonce++

		time.Sleep(time.Second)

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
			require.Equal(t, referenceBlock.GetNonce(), crtBlock.GetNonce())
		}
	}
}

func printAccount(node *integrationTests.TestProcessorNode) {
	accnt, _ := node.AccntState.GetExistingAccount(node.OwnAccount.Address)
	if accnt == nil {
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

func generateInitialRandomness(numShards uint32) map[uint32][]byte {
	randomness := make(map[uint32][]byte)

	for i := uint32(0); i < numShards; i++ {
		randomness[i] = []byte("root hash")
	}

	randomness[sharding.MetachainShardId] = []byte("root hash")

	return randomness
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

	numTxs := 2
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
