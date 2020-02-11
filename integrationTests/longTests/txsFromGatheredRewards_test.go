package longTests

import (
	"bytes"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("integrationtests/singleshard/block")

// TestShardShouldNotProposeAndExecuteTwoBlocksInSameRound tests what happens in a setting where a proposer/validator
// periodically sends transactions to someone else. The tests starts with the node having 0 balance.
func TestExecutingTransactionsFromGatheredRewards(t *testing.T) {
	t.Skip("this is a long test")

	numOfNodes := 2
	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)

	_ = logger.SetLogLevel("*:DEBUG,process:TRACE")

	nodesMap := integrationTests.CreateProcessorNodesWithNodesCoordinator(numOfNodes, 1, 1, numOfNodes, 1, advertiserAddr)

	for _, nodes := range nodesMap {
		for _, n := range nodes {
			n.EconomicsData.SetRewards(big.NewInt(0).SetUint64(2 * integrationTests.MinTxGasLimit * integrationTests.MinTxGasPrice))
		}
	}

	defer func() {
		_ = advertiser.Close()
		for _, shards := range nodesMap {
			for _, n := range shards {
				_ = n.Messenger.Close()
			}
		}
	}()

	for _, shards := range nodesMap {
		for _, n := range shards {
			_ = n.Messenger.Bootstrap()
		}
	}

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(stepDelay)

	round := uint64(0)
	nonce := uint64(1)
	round = integrationTests.IncrementAndPrintRound(round)
	randomness := generateInitialRandomness(1)
	pubKeys, _ := nodesMap[0][0].NodesCoordinator.GetValidatorsPublicKeys(randomness[0], round, 0)

	var proposer *integrationTests.TestProcessorNode
	for _, shard := range nodesMap {
		for _, n := range shard {
			buffPk, _ := n.NodeKeys.Pk.ToByteArray()
			if bytes.Equal([]byte(pubKeys[0]), buffPk) {
				proposer = n
				break
			}
		}
	}
	if proposer == nil {
		assert.Fail(t, "failed to find the proposer")
		return
	}

	transferValue := integrationTests.MinTxGasLimit * integrationTests.MinTxGasPrice / 2

	go func() {
		for {
			generateAndSendTxs(
				proposer,
				transferValue,
				proposer.OwnAccount.SkTxSign,
				proposer.OwnAccount.Address,
				nodesMap[sharding.MetachainShardId][0].OwnAccount.PkTxSign)
			time.Sleep(time.Second)
		}
	}()

	nbBlocksProduced := uint64(20)
	var consensusNodes map[uint32][]*integrationTests.TestProcessorNode
	for i := uint64(0); i < nbBlocksProduced; i++ {
		log.Info("starting", "leader", proposer.OwnAccount.Address.Bytes())
		printBalance(proposer)

		for _, nodes := range nodesMap {
			integrationTests.UpdateRound(nodes, round)
		}
		_, _, consensusNodes, randomness = integrationTests.AllShardsProposeBlock(round, nonce, randomness, nodesMap)

		indexesProposers := getBlockProposersIndexes(consensusNodes, nodesMap)
		integrationTests.SyncAllShardsWithRoundBlock(t, nodesMap, indexesProposers, round)
		round++
		nonce++

		time.Sleep(time.Second)

		for _, nodes := range nodesMap {
			crtBlock := nodes[0].BlockChain.GetCurrentBlockHeader()
			for _, n := range nodes {
				blkc := n.BlockChain.GetCurrentBlockHeader()
				require.False(t, (crtBlock == nil) != (blkc == nil))
				require.Equal(t, crtBlock.GetNonce(), blkc.GetNonce())
			}
		}
	}

	printBalance(proposer)
}

func printBalance(node *integrationTests.TestProcessorNode) {
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

func generateInitialRandomness(nbShards uint32) map[uint32][]byte {
	randomness := make(map[uint32][]byte)

	for i := uint32(0); i < nbShards; i++ {
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
