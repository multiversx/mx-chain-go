package interceptedResolvedBulkTx

import (
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

// TestNode_InterceptorBulkTxsSentFromSameShardShouldRemainInSenderShard tests what happens when
// a network with 5 shards, each containing 3 nodes broadcast 100 transactions from node 0.
// Node 0 is part of the shard 0 and its public key is mapped also in shard 0.
// Transactions should spread only in shard 0.
func TestNode_InterceptorBulkTxsSentFromSameShardShouldRemainInSenderShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 6
	nodesPerShard := 3
	numMetachainNodes := 0
	shardId := uint32(5)

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)
	integrationTests.CreateAccountForNodes(nodes)
	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	txToSend := 100
	generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), shardId)

	fmt.Println("Generating and broadcasting transactions...")
	_, pkInShardFive, _ := integrationTests.GenerateSkAndPkInShard(generateCoordinator, shardId)
	pkBytes, _ := pkInShardFive.ToByteArray()
	addrInShardFive := integrationTests.TestAddressPubkeyConverter.Encode(pkBytes)

	idxSender := 0
	shardId = nodes[idxSender].ShardCoordinator.SelfId()
	balanceValue := big.NewInt(1000000000)
	transactionValue := big.NewInt(1)
	senderPrivateKeys := []crypto.PrivateKey{nodes[idxSender].OwnAccount.SkTxSign}
	integrationTests.CreateMintingForSenders(nodes, shardId, senderPrivateKeys, balanceValue)

	_ = nodes[idxSender].Node.GenerateAndSendBulkTransactions(
		addrInShardFive,
		transactionValue,
		uint64(txToSend),
		nodes[idxSender].OwnAccount.SkTxSign,
		nil,
		integrationTests.ChainID,
		integrationTests.MinTransactionVersion,
	)

	time.Sleep(time.Second * 10)

	//since there is a slight chance that some transactions get lost (peer to slow, queue full, validators throttling...)
	//we should get the max transactions received
	maxTxReceived := int32(0)
	for _, n := range nodes {
		txRecv := atomic.LoadInt32(&n.CounterTxRecv)

		if txRecv > maxTxReceived {
			maxTxReceived = txRecv
		}
	}

	assert.True(t, maxTxReceived > 0)

	//only sender shard (all 3 nodes from shard 0) have the transactions
	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() == 0 {
			assert.Equal(t, maxTxReceived, atomic.LoadInt32(&n.CounterTxRecv))
			continue
		}

		assert.Equal(t, int32(0), atomic.LoadInt32(&n.CounterTxRecv))
	}
}

// TestNode_InterceptorBulkTxsSentFromOtherShardShouldBeRoutedInSenderShard tests what happens when
// a network with 5 shards, each containing 3 nodes broadcast 100 transactions from node 0.
// Node 0 is part of the shard 0 and its public key is mapped in shard 4.
// Transactions should spread only in shard 4.
func TestNode_InterceptorBulkTxsSentFromOtherShardShouldBeRoutedInSenderShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 6
	nodesPerShard := 3
	numMetachainNodes := 0
	firstSkInShard := uint32(4)
	shardId := uint32(5)

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)
	nodes[0] = integrationTests.NewTestProcessorNode(uint32(numOfShards), 0, firstSkInShard)
	integrationTests.CreateAccountForNodes(nodes)
	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	txToSend := 100

	generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), shardId)

	_, pkInShardFive, _ := integrationTests.GenerateSkAndPkInShard(generateCoordinator, shardId)
	pkBytes, _ := pkInShardFive.ToByteArray()
	addrInShardFive := integrationTests.TestAddressPubkeyConverter.Encode(pkBytes)

	idxSender := 0
	shardId = uint32(4)
	mintingValue := big.NewInt(1000000000)
	txValue := big.NewInt(1)
	senderPrivateKeys := []crypto.PrivateKey{nodes[idxSender].OwnAccount.SkTxSign}
	integrationTests.CreateMintingForSenders(nodes, shardId, senderPrivateKeys, mintingValue)

	dispatcherNodeID := nodes[0].ShardCoordinator.ComputeId(nodes[idxSender].OwnAccount.Address)
	dispatcherNode := nodes[0]
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() == dispatcherNodeID {
			dispatcherNode = node
			break
		}
	}

	_ = dispatcherNode.Node.GenerateAndSendBulkTransactions(
		addrInShardFive,
		txValue,
		uint64(txToSend),
		nodes[idxSender].OwnAccount.SkTxSign,
		nil,
		integrationTests.ChainID,
		integrationTests.MinTransactionVersion,
	)

	//display, can be removed
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)

		fmt.Println(integrationTests.MakeDisplayTable(nodes))
	}

	//since there is a slight chance that some transactions get lost (peer to slow, queue full...)
	//we should get the max transactions received
	maxTxReceived := int32(0)
	for _, n := range nodes {
		txRecv := atomic.LoadInt32(&n.CounterTxRecv)

		if txRecv > maxTxReceived {
			maxTxReceived = txRecv
		}
	}

	assert.True(t, maxTxReceived > 0)

	//only sender shard (all 3 nodes from shard firstSkInShard) has the transactions
	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() == firstSkInShard {
			assert.Equal(t, atomic.LoadInt32(&n.CounterTxRecv), maxTxReceived)
			continue
		}

		assert.Equal(t, atomic.LoadInt32(&n.CounterTxRecv), int32(0))
	}
}

// TestNode_InterceptorBulkTxsSentFromOtherShardShouldBeRoutedInSenderShardAndRequestShouldWork tests what happens when
// a network with 5 shards, each containing 3 nodes broadcast 10 transactions from node 0 and the destination shard
// requests the same set of those generated transactions.
// Node 0 is part of the shard 0 and its public key is mapped in shard 4.
// Destination shard nodes (shard id 5) requests the same transactions set from nodes in shard 4 (senders).
// Transactions should spread only in shards 4 and 5.
// Transactions requested by another shard (2 for example) will not store the received transactions
// (interceptors will filter them out)
func TestNode_InterceptorBulkTxsSentFromOtherShardShouldBeRoutedInSenderShardAndRequestShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 6
	nodesPerShard := 1
	numMetachainNodes := 0
	firstSkInShard := uint32(4)

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)
	nodes[0] = integrationTests.NewTestProcessorNode(uint32(numOfShards), 0, firstSkInShard)
	integrationTests.CreateAccountForNodes(nodes)
	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	txToSend := 2

	shardRequester := uint32(5)
	randomShard := uint32(2)

	generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), shardRequester)

	_, pkInShardFive, _ := integrationTests.GenerateSkAndPkInShard(generateCoordinator, 5)
	pkBytes, _ := pkInShardFive.ToByteArray()
	addrInShardFive := integrationTests.TestAddressPubkeyConverter.Encode(pkBytes)

	mutGeneratedTxHashes := sync.Mutex{}
	generatedTxHashes := make([][]byte, 0)
	//wire a new hook for generated txs on a node in sender shard to populate tx hashes generated
	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() == firstSkInShard {
			n.DataPool.Transactions().RegisterOnAdded(func(key []byte, value interface{}) {
				mutGeneratedTxHashes.Lock()
				generatedTxHashes = append(generatedTxHashes, key)
				mutGeneratedTxHashes.Unlock()
			})
		}
	}

	idxSender := 0
	shardId := uint32(4)
	mintingValue := big.NewInt(1000000000)
	txValue := big.NewInt(1)
	senderPrivateKeys := []crypto.PrivateKey{nodes[idxSender].OwnAccount.SkTxSign}
	integrationTests.CreateMintingForSenders(nodes, shardId, senderPrivateKeys, mintingValue)

	whiteListTxs := func(txs []*transaction.Transaction) {
		integrationTests.WhiteListTxs(nodes, txs)
	}

	dispatcherNodeID := nodes[0].ShardCoordinator.ComputeId(nodes[idxSender].OwnAccount.Address)
	dispatcherNode := nodes[0]
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() == dispatcherNodeID {
			dispatcherNode = node
			break
		}
	}
	_ = dispatcherNode.Node.GenerateAndSendBulkTransactions(
		addrInShardFive,
		txValue,
		uint64(txToSend),
		nodes[0].OwnAccount.SkTxSign,
		whiteListTxs,
		integrationTests.ChainID,
		integrationTests.MinTransactionVersion,
	)

	fmt.Println("Waiting for senders to fetch generated transactions...")
	time.Sleep(time.Second * 10)

	fmt.Println("Request transactions by destination shard nodes...")
	//periodically compute and request missing transactions
	for i := 0; i < 10; i++ {
		integrationTests.ComputeAndRequestMissingTransactions(nodes, generatedTxHashes, firstSkInShard, shardRequester, randomShard)
		time.Sleep(time.Second)

		fmt.Println(integrationTests.MakeDisplayTable(nodes))
	}

	//since there is a slight chance that some transactions get lost (peer to slow, queue full...)
	//we should get the max transactions received
	maxTxReceived := int32(0)
	for _, n := range nodes {
		txRecv := atomic.LoadInt32(&n.CounterTxRecv)

		if txRecv > maxTxReceived {
			maxTxReceived = txRecv
		}
	}

	assert.True(t, maxTxReceived > 0)

	//only sender and destination shards have the transactions
	for _, n := range nodes {
		isSenderOrDestinationShard := n.ShardCoordinator.SelfId() == firstSkInShard || n.ShardCoordinator.SelfId() == shardRequester

		if isSenderOrDestinationShard {
			assert.Equal(t, atomic.LoadInt32(&n.CounterTxRecv), maxTxReceived)
			continue
		}

		assert.Equal(t, atomic.LoadInt32(&n.CounterTxRecv), int32(0))
	}
}

func TestNode_InMultiShardEnvRequestTxsShouldRequireFromTheOtherShardAndSameShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes := make([]*integrationTests.TestProcessorNode, 0)
	maxShards := 2
	nodesPerShard := 2
	txGenerated := 10

	defer func() {
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	//shard 0, requesters
	recvTxs := make(map[int]map[string]struct{})
	mutRecvTxs := sync.Mutex{}
	connectableNodes := make([]integrationTests.Connectable, 0)
	for i := 0; i < nodesPerShard; i++ {
		dPool := integrationTests.CreateRequesterDataPool(recvTxs, &mutRecvTxs, i, 0)

		tn := integrationTests.NewTestProcessorNodeWithCustomDataPool(
			uint32(maxShards),
			0,
			0,
			dPool,
		)

		nodes = append(nodes, tn)
		connectableNodes = append(connectableNodes, tn)
	}

	senderShardId := uint32(0)
	recvShardId := uint32(1)
	balanceValue := big.NewInt(1000000000)
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(maxShards), senderShardId)
	dPool, txHashesGenerated, txsSndAddr := integrationTests.CreateResolversDataPool(t, txGenerated, senderShardId, recvShardId, shardCoordinator)

	integrationTests.CreateMintingFromAddresses(nodes, txsSndAddr, balanceValue)

	//shard 1, resolvers, same data pool, does not matter
	for i := 0; i < nodesPerShard; i++ {
		tn := integrationTests.NewTestProcessorNodeWithCustomDataPool(
			uint32(maxShards),
			1,
			1,
			dPool,
		)

		atomic.StoreInt32(&tn.CounterTxRecv, int32(txGenerated))

		nodes = append(nodes, tn)
		connectableNodes = append(connectableNodes, tn)
	}

	integrationTests.ConnectNodes(connectableNodes)
	integrationTests.DisplayAndStartNodes(nodes)
	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(time.Second * 5)

	fmt.Println(integrationTests.MakeDisplayTable(nodes))

	fmt.Println("Request nodes start asking the data...")
	reqShardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(maxShards), 0)
	for i := 0; i < nodesPerShard; i++ {
		resolver, _ := nodes[i].ResolverFinder.Get(factory.TransactionTopic + reqShardCoordinator.CommunicationIdentifier(1))
		txResolver, ok := resolver.(*resolvers.TxResolver)
		assert.True(t, ok)

		_ = txResolver.RequestDataFromHashArray(txHashesGenerated, 0)
	}

	time.Sleep(time.Second * 5)
	mutRecvTxs.Lock()
	defer mutRecvTxs.Unlock()
	for i := 0; i < nodesPerShard; i++ {
		mapTx := recvTxs[i]
		assert.NotNil(t, mapTx)

		txsReceived := len(recvTxs[i])
		assert.Equal(t, txGenerated, txsReceived)

		atomic.StoreInt32(&nodes[i].CounterTxRecv, int32(txsReceived))
	}

	fmt.Println(integrationTests.MakeDisplayTable(nodes))
}
