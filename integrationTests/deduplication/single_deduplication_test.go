package deduplication

import (
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process/txsSender"
	"github.com/multiversx/mx-chain-go/sharding"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
)

type CountingMessenger struct {
	p2p.Messenger       // embed the real messenger
	MsgCount      int32 // atomic counter
}

func (cm *CountingMessenger) ProcessReceivedMessage(msg p2p.MessageP2P, from core.PeerID, handler p2p.MessageHandler) ([]byte, error) {
	atomic.AddInt32(&cm.MsgCount, 1)
	fmt.Println(">>> CountingMessenger saw message:", msg.Topic(), "from", from.Pretty())
	// forward to the actual messenger's processing
	return cm.Messenger.ProcessReceivedMessage(msg, from, handler)
}

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

	// messenger1 = shard0 sender (connected to node0)
	messenger1 := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            uint32(numShards),
		NodeShardId:          1,
		TxSignPrivKeyShardId: 0,
	})

	// messenger2 = shard1 receiver (connected only to node0)
	messenger2 := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            uint32(numShards),
		NodeShardId:          1,
		TxSignPrivKeyShardId: 1,
	})

	node0_wrappedMessenger := &CountingMessenger{Messenger: node0.MainMessenger}
	node0.MainMessenger = node0_wrappedMessenger
	node0.MainMessenger.RegisterMessageProcessor(txsSender.SendTransactionsPipe, "transactions_0_1", node0_wrappedMessenger)

	msg2_wrappedMessenger := &CountingMessenger{Messenger: messenger2.MainMessenger}
	messenger2.MainMessenger = msg2_wrappedMessenger
	messenger2.MainMessenger.RegisterMessageProcessor(txsSender.SendTransactionsPipe, "transactions_0_1", msg2_wrappedMessenger)

	nodes := []*integrationTests.TestProcessorNode{node0, messenger1, messenger2}
	integrationTests.CreateAccountForNodes(nodes)

	//connectables := []integrationTests.Connectable{node0, messenger1, messenger2}
	integrationTests.ConnectNodes([]integrationTests.Connectable{messenger1, node0})
	integrationTests.ConnectNodes([]integrationTests.Connectable{node0, messenger2})

	integrationTests.DisplayAndStartNodes([]*integrationTests.TestProcessorNode{node0, messenger1, messenger2})

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

	mutGeneratedTxHashes := sync.Mutex{}
	generatedTxHashes := make([][]byte, 0)
	//wire a new hook for generated txs on a node in sender shard to populate tx hashes generated
	messenger1.DataPool.Transactions().RegisterOnAdded(func(key []byte, value interface{}) {
		mutGeneratedTxHashes.Lock()
		generatedTxHashes = append(generatedTxHashes, key)
		mutGeneratedTxHashes.Unlock()
	})

	// mint balance for messenger1
	value := big.NewInt(1)
	balance := big.NewInt(1000000000)
	senderPrivKeys := []crypto.PrivateKey{messenger1.OwnAccount.SkTxSign}
	integrationTests.CreateMintingForSenders([]*integrationTests.TestProcessorNode{node0, messenger1, messenger2}, senderShard, senderPrivKeys, balance)

	// --- send 1 transaction from messenger1 to shard1
	txCount := 1
	whiteListTxs := func(txs []*transaction.Transaction) {
		integrationTests.WhiteListTxs(nodes, txs)
	}

	duplicatedTxCount := 5 // each tx sent 5 times
	err := messenger1.Node.GenerateAndSendDuplicatedBulkTransactions(
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

	time.Sleep(time.Second * 10)

	fmt.Println(integrationTests.MakeDisplayTable(nodes))

	// --- validate deduplication
	txRecvNode0 := atomic.LoadInt32(&node0.CounterTxRecv)
	txRecvMessenger2 := atomic.LoadInt32(&messenger2.CounterTxRecv)
	msgsRecvNode0 := node0_wrappedMessenger.MsgCount
	msgsRecvMessenger2 := msg2_wrappedMessenger.MsgCount

	assert.Equal(t, int32(txCount), txRecvNode0, "node0 should receive duplicated txs")
	assert.Equal(t, int32(txCount), txRecvMessenger2, "messenger2 should receive only 1 tx despite duplicates")
	fmt.Println("node0 received total messages:", atomic.LoadInt32(&node0.MainMessenger.(*CountingMessenger).MsgCount))
	fmt.Println("messenger2 received total messages:", atomic.LoadInt32(&msg2_wrappedMessenger.MsgCount))
	assert.Equal(t, duplicatedTxCount, msgsRecvNode0, "interceptor should have received all duplicated messages")
	assert.Equal(t, int32(1), msgsRecvMessenger2, "messenger2 should have received only one message")
}
