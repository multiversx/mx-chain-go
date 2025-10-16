package deduplication

import (
	"fmt"
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/sharding"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
)

// TestNode_ShouldDeduplicateDuplicateMessages
// Scenario:
//
//	messenger1 -> (duplicates) -> node0(interceptor) -> (cross-shard) -> messenger2
func TestNode_ShouldDeduplicateMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("not a short test")
	}
	logger.SetLogLevel("*:DEBUG")
	defer logger.SetLogLevel("*:ERROR")

	// 2 shards: one sender, one receiver
	numShards := 2

	// node0 = the test node (interceptor)
	node0 := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            uint32(numShards),
		NodeShardId:          1,
		TxSignPrivKeyShardId: 1,
	})
	node0Proc := NewCountingMessageProcessor(node0.MainMessenger, "node0")
	node0.MainMessenger.RegisterMessageProcessor("transactions_0_1", "", node0Proc)

	// messenger1 = shard0 sender (connected to node0)
	messenger1 := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            uint32(numShards),
		NodeShardId:          1,
		TxSignPrivKeyShardId: 0,
	})

	messenger1Proc := NewCountingMessageProcessor(messenger1.MainMessenger, "messenger1")
	messenger1.MainMessenger.RegisterMessageProcessor("transactions_0_1", "", messenger1Proc)

	// messenger2 = shard1 receiver (connected only to node0)
	messenger2 := integrationTests.CreateMessengerWithNoDiscovery()
	messenger2.CreateTopic("transactions_0_1", true)
	messenger2Proc := NewCountingMessageProcessor(messenger2, "messenger2")
	messenger2.RegisterMessageProcessor("transactions_0_1", "", messenger2Proc)

	nodes := []*integrationTests.TestProcessorNode{node0, messenger1}
	integrationTests.CreateAccountForNodes(nodes)

	messenger2.ConnectToPeer(string(node0.MainMessenger.Addresses()[0]))
	messenger1.MainMessenger.ConnectToPeer(string(node0.MainMessenger.Addresses()[0]))

	integrationTests.DisplayAndStartNodes([]*integrationTests.TestProcessorNode{node0, messenger1})

	defer func() {
		for _, n := range nodes {
			logger.SetLogLevel("*:ERROR")
			n.Close()
		}
	}()

	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(time.Second * 8)

	// --- prepare a receiver address that maps to shard 1 (cross-shard)
	senderShard := uint32(0)
	receiverShard := uint32(1)
	coord, _ := sharding.NewMultiShardCoordinator(uint32(numShards), receiverShard)
	_, pkInShard1, _ := integrationTests.GenerateSkAndPkInShard(coord, receiverShard)
	pkBytes, _ := pkInShard1.ToByteArray()
	receiverAddr, _ := integrationTests.TestAddressPubkeyConverter.Encode(pkBytes)

	// mint balance for messenger1
	value := big.NewInt(1)
	balance := big.NewInt(1000000000)
	senderPrivKeys := []crypto.PrivateKey{messenger1.OwnAccount.SkTxSign}
	integrationTests.CreateMintingForSenders([]*integrationTests.TestProcessorNode{node0, messenger1}, senderShard, senderPrivKeys, balance)

	// --- send transactions from messenger1 to shard1
	txCount := 10
	whiteListTxs := func(txs []*transaction.Transaction) {
		integrationTests.WhiteListTxs(nodes, txs)
	}

	duplicatedTxCount := 5 // each batch sent 5 times
	numPackets, err := messenger1.Node.GenerateAndSendDuplicatedBulkTransactions(
		duplicatedTxCount,
		receiverAddr,
		value,
		uint64(txCount),
		messenger1.OwnAccount.SkTxSign,
		whiteListTxs,
		integrationTests.ChainID,
		integrationTests.MinTransactionVersion,
	)

	assert.Nil(t, err)
	time.Sleep(time.Second * 20)
	fmt.Println(integrationTests.MakeDisplayTable(nodes))

	// --- validate deduplication
	txRecvNode0 := atomic.LoadInt32(&node0.CounterTxRecv)
	msgsRecvNode0AfterDeduplication := node0Proc.NumMessagesReceived() //node0.MainMessenger.(*integrationTests.CountingMessenger).MsgCount
	msgsRecvMessenger2 := messenger2Proc.NumMessagesReceived()
	msgsRecvMessenger1 := messenger1Proc.NumMessagesReceived()

	assert.Equal(t, int32(txCount), txRecvNode0, "node0 should receive all txs")
	fmt.Println("node0 received total messages:", msgsRecvNode0AfterDeduplication)
	fmt.Println("messenger2 received total messages:", msgsRecvMessenger2)
	assert.Equal(t, int32(numPackets), int32(msgsRecvNode0AfterDeduplication), "interceptor should have deduplicated messages")
	assert.Equal(t, int32(numPackets), int32(msgsRecvMessenger2), "messenger2 should have received deduplicated messages")
	assert.Equal(t, int32(numPackets*duplicatedTxCount), int32(msgsRecvMessenger1), "messenger1 should have received all messages it sent")
}
