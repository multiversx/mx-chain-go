package transaction

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/addressConverters"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/stretchr/testify/assert"
	"github.com/whyrusleeping/go-logging"
)

func init() {
	logging.SetLevel(logging.ERROR, "pubsub")
}

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

	advertiser := createMessengerWithKadDht(context.Background(), "")
	advertiser.Bootstrap()

	nodes := createNodesWithNodeSkInShardExceptFirst(
		numOfShards,
		nodesPerShard,
		0,
		getConnectableAddress(advertiser),
	)
	displayAndStartNodes(nodes)

	defer func() {
		advertiser.Close()
		for _, n := range nodes {
			n.node.Stop()
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(time.Second * 5)

	txToSend := 100

	generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), 5)
	generateAddrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")

	fmt.Println("Generating and broadcasting transactions...")
	addrInShardFive := createDummyHexAddressInShard(generateCoordinator, generateAddrConverter)
	nodes[0].node.GenerateAndSendBulkTransactions(addrInShardFive, big.NewInt(1), uint64(txToSend))
	time.Sleep(time.Second * 10)

	//since there is a slight chance that some transactions get lost (peer to slow, queue full, validators throttling...)
	//we should get the max transactions received
	maxTxReceived := int32(0)
	for _, n := range nodes {
		txRecv := atomic.LoadInt32(&n.txRecv)

		if txRecv > maxTxReceived {
			maxTxReceived = txRecv
		}
	}

	assert.True(t, maxTxReceived > 0)

	//only sender shard (all 3 nodes from shard 0) have the transactions
	for _, n := range nodes {
		if n.shardId == 0 {
			assert.Equal(t, maxTxReceived, atomic.LoadInt32(&n.txRecv))
			continue
		}

		assert.Equal(t, int32(0), atomic.LoadInt32(&n.txRecv))
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

	firstSkInShard := uint32(4)

	advertiser := createMessengerWithKadDht(context.Background(), "")
	advertiser.Bootstrap()

	nodes := createNodesWithNodeSkInShardExceptFirst(
		numOfShards,
		nodesPerShard,
		firstSkInShard,
		getConnectableAddress(advertiser),
	)
	displayAndStartNodes(nodes)

	defer func() {
		advertiser.Close()
		for _, n := range nodes {
			n.node.Stop()
		}
	}()

	// delay for bootstrapping and topic a0nnouncement
	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(time.Second * 5)

	txToSend := 100

	generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), 5)
	generateAddrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")

	addrInShardFive := createDummyHexAddressInShard(generateCoordinator, generateAddrConverter)

	nodes[0].node.GenerateAndSendBulkTransactions(addrInShardFive, big.NewInt(1), uint64(txToSend))

	//display, can be removed
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)

		fmt.Println(makeDisplayTable(nodes))
	}

	//since there is a slight chance that some transactions get lost (peer to slow, queue full...)
	//we should get the max transactions received
	maxTxReceived := int32(0)
	for _, n := range nodes {
		txRecv := atomic.LoadInt32(&n.txRecv)

		if txRecv > maxTxReceived {
			maxTxReceived = txRecv
		}
	}

	assert.True(t, maxTxReceived > 0)

	//only sender shard (all 3 nodes from shard firstSkInShard) has the transactions
	for _, n := range nodes {
		if n.shardId == firstSkInShard {
			assert.Equal(t, atomic.LoadInt32(&n.txRecv), maxTxReceived)
			continue
		}

		assert.Equal(t, atomic.LoadInt32(&n.txRecv), int32(0))
	}
}

// TestNode_InterceptorBulkTxsSentFromOtherShardShouldBeRoutedInSenderShardAndRequestShouldWork tests what happens when
// a network with 5 shards, each containing 3 nodes broadcast 100 transactions from node 0 and the destination shard
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
	nodesPerShard := 3

	firstSkInShard := uint32(4)

	advertiser := createMessengerWithKadDht(context.Background(), "")
	advertiser.Bootstrap()

	nodes := createNodesWithNodeSkInShardExceptFirst(
		numOfShards,
		nodesPerShard,
		firstSkInShard,
		getConnectableAddress(advertiser),
	)
	displayAndStartNodes(nodes)

	defer func() {
		advertiser.Close()
		for _, n := range nodes {
			n.node.Stop()
		}
	}()

	// delay for bootstrapping and topic a0nnouncement
	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(time.Second * 5)

	txToSend := 100

	shardRequestor := uint32(5)
	randomShard := uint32(2)

	generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), shardRequestor)
	generateAddrConverter, _ := addressConverters.NewPlainAddressConverter(32, "0x")

	addrInShardFive := createDummyHexAddressInShard(generateCoordinator, generateAddrConverter)

	mutGeneratedTxHashes := sync.Mutex{}
	generatedTxHashes := make([][]byte, 0)
	//wire a new hook for generated txs on a node in sender shard to populate tx hashes generated
	for _, n := range nodes {
		if n.shardId == firstSkInShard {
			n.dPool.Transactions().RegisterHandler(func(key []byte) {
				mutGeneratedTxHashes.Lock()
				generatedTxHashes = append(generatedTxHashes, key)
				mutGeneratedTxHashes.Unlock()
			})
		}
	}

	nodes[0].node.GenerateAndSendBulkTransactions(addrInShardFive, big.NewInt(1), uint64(txToSend))

	fmt.Println("Waiting for senders to fetch generated transactions...")
	time.Sleep(time.Second * 10)

	//right now all 3 nodes from sender shard have the transactions
	//nodes from shardRequestor should ask and receive all generated transactions
	mutGeneratedTxHashes.Lock()
	copyNeededTransactions(nodes, generatedTxHashes)
	mutGeneratedTxHashes.Unlock()

	fmt.Println("Request transactions by destination shard nodes...")
	//periodically compute and request missing transactions
	for i := 0; i < 10; i++ {
		computeAndRequestMissingTransactions(nodes, firstSkInShard, shardRequestor, randomShard)
		time.Sleep(time.Second)

		fmt.Println(makeDisplayTable(nodes))
	}

	//since there is a slight chance that some transactions get lost (peer to slow, queue full...)
	//we should get the max transactions received
	maxTxReceived := int32(0)
	for _, n := range nodes {
		txRecv := atomic.LoadInt32(&n.txRecv)

		if txRecv > maxTxReceived {
			maxTxReceived = txRecv
		}
	}

	assert.True(t, maxTxReceived > 0)

	//only sender and destination shards have the transactions
	for _, n := range nodes {
		isSenderOrDestinationShard := n.shardId == firstSkInShard || n.shardId == shardRequestor

		if isSenderOrDestinationShard {
			assert.Equal(t, atomic.LoadInt32(&n.txRecv), maxTxReceived)
			continue
		}

		assert.Equal(t, atomic.LoadInt32(&n.txRecv), int32(0))
	}
}

func TestNode_InMultiShardEnvRequestTxsShouldRequireOnlyFromTheOtherShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	advertiser := createMessengerWithKadDht(context.Background(), "")
	advertiser.Bootstrap()

	nodes := make([]*testNode, 0)
	maxShards := 2
	nodesPerShard := 2
	txGenerated := 10

	defer func() {
		advertiser.Close()
		for _, n := range nodes {
			n.node.Stop()
		}
	}()

	//shard 0, requestors
	recvTxs := make(map[int]map[string]struct{})
	mutRecvTxs := sync.Mutex{}
	for i := 0; i < nodesPerShard; i++ {
		dPool := createRequestorDataPool(t, recvTxs, mutRecvTxs, i)

		tn := createNode(
			0,
			maxShards,
			dPool,
			0,
			getConnectableAddress(advertiser),
		)

		nodes = append(nodes, tn)
	}

	var txHashesGenerated [][]byte
	var dPool dataRetriever.PoolsHolder
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(maxShards), 0)
	dPool, txHashesGenerated = createResolversDataPool(t, txGenerated, 0, 1, shardCoordinator)
	//shard 1, resolvers, same data pool, does not matter
	for i := 0; i < nodesPerShard; i++ {
		tn := createNode(
			1,
			maxShards,
			dPool,
			1,
			getConnectableAddress(advertiser),
		)

		atomic.StoreInt32(&tn.txRecv, int32(txGenerated))

		nodes = append(nodes, tn)
	}

	displayAndStartNodes(nodes)
	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(time.Second * 5)

	fmt.Println(makeDisplayTable(nodes))

	fmt.Println("Request nodes start asking the data...")
	reqShardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(maxShards), 0)
	for i := 0; i < nodesPerShard; i++ {
		resolver, _ := nodes[i].resFinder.Get(factory.TransactionTopic + reqShardCoordinator.CommunicationIdentifier(1))
		txResolver, ok := resolver.(*resolvers.TxResolver)
		assert.True(t, ok)

		txResolver.RequestDataFromHashArray(txHashesGenerated)
	}

	time.Sleep(time.Second * 5)
	mutRecvTxs.Lock()
	defer mutRecvTxs.Unlock()
	for i := 0; i < nodesPerShard; i++ {
		mapTx := recvTxs[i]
		assert.NotNil(t, mapTx)

		txsReceived := len(recvTxs[i])
		assert.Equal(t, txGenerated, txsReceived)

		atomic.StoreInt32(&nodes[i].txRecv, int32(txsReceived))
	}

	fmt.Println(makeDisplayTable(nodes))
}

func createRequestorDataPool(
	t *testing.T,
	recvTxs map[int]map[string]struct{},
	mutRecvTxs sync.Mutex,
	nodeIndex int,
) dataRetriever.PoolsHolder {

	//not allowed to request data from the same shard
	return createTestDataPool(
		&mock.ShardedDataStub{
			SearchFirstDataCalled: func(key []byte) (value interface{}, ok bool) {
				assert.Fail(t, "same-shard requestors should not be queried")
				return nil, false
			},
			ShardDataStoreCalled: func(cacheId string) (c storage.Cacher) {
				assert.Fail(t, "same-shard requestors should not be queried")
				return nil
			},
			AddDataCalled: func(key []byte, data interface{}, cacheId string) {
				mutRecvTxs.Lock()
				defer mutRecvTxs.Unlock()

				txMap := recvTxs[nodeIndex]
				if txMap == nil {
					txMap = make(map[string]struct{})
					recvTxs[nodeIndex] = txMap
				}

				txMap[string(key)] = struct{}{}
			},
			RegisterHandlerCalled: func(i func(key []byte)) {
			},
		},
	)
}

func createResolversDataPool(
	t *testing.T,
	maxTxs int,
	senderShardID uint32,
	recvShardId uint32,
	shardCoordinator sharding.Coordinator,
) (dataRetriever.PoolsHolder, [][]byte) {

	txHashes := make([][]byte, maxTxs)

	txPool, _ := shardedData.NewShardedData(storage.CacheConfig{Size: 100, Type: storage.LRUCache})

	for i := 0; i < maxTxs; i++ {
		tx, txHash := generateValidTx(t, shardCoordinator, senderShardID, recvShardId)
		cacherIdentifier := process.ShardCacherIdentifier(1, 0)
		txPool.AddData(txHash, tx, cacherIdentifier)
		txHashes[i] = txHash
	}

	return createTestDataPool(txPool), txHashes
}

func generateValidTx(
	t *testing.T,
	shardCoordinator sharding.Coordinator,
	senderShardId uint32,
	receiverShardId uint32,
) (*transaction.Transaction, []byte) {

	skSender, pkSender := generateSkPkInShardAndCreateAccount(shardCoordinator, senderShardId, nil)
	pkSenderBuff, _ := pkSender.ToByteArray()
	_, pkRecv := generateSkPkInShardAndCreateAccount(shardCoordinator, receiverShardId, nil)
	pkRecvBuff, _ := pkRecv.ToByteArray()

	accnts := createAccountsDB()
	addrSender, _ := addrConverter.CreateAddressFromPublicKeyBytes(pkSenderBuff)
	accnts.GetAccountWithJournal(addrSender)
	accnts.Commit()

	mockNode, _ := node.NewNode(
		node.WithMarshalizer(marshalizer),
		node.WithHasher(hasher),
		node.WithAddressConverter(addrConverter),
		node.WithKeyGen(keyGen),
		node.WithTxSingleSigner(singleSigner),
		node.WithTxSignPrivKey(skSender),
		node.WithTxSignPubKey(pkSender),
		node.WithAccountsAdapter(accnts),
	)

	tx, err := mockNode.GenerateTransaction(
		hex.EncodeToString(pkSenderBuff),
		hex.EncodeToString(pkRecvBuff),
		big.NewInt(1),
		"",
	)
	assert.Nil(t, err)

	txBuff, _ := marshalizer.Marshal(tx)
	txHash := hasher.Compute(string(txBuff))

	return tx, txHash
}

func copyNeededTransactions(
	nodes []*testNode,
	generatedTxHashes [][]byte,
) {

	for _, n := range nodes {
		n.neededTxs = make([][]byte, len(generatedTxHashes))

		n.mutNeededTxs.Lock()
		for i := 0; i < len(generatedTxHashes); i++ {
			n.neededTxs[i] = make([]byte, len(generatedTxHashes[i]))
			copy(n.neededTxs[i], generatedTxHashes[i])
		}
		n.mutNeededTxs.Unlock()
	}
}

func computeAndRequestMissingTransactions(
	nodes []*testNode,
	shardResolver uint32,
	shardRequestors ...uint32,
) {
	for _, n := range nodes {
		if !isInList(n.shardId, shardRequestors) {
			continue
		}

		computeMissingTransactions(n)
		requestMissingTransactions(n, shardResolver)
	}
}

func isInList(searched uint32, list []uint32) bool {
	for _, val := range list {
		if val == searched {
			return true
		}
	}

	return false
}

func computeMissingTransactions(n *testNode) {
	n.mutNeededTxs.Lock()

	newNeededTxs := make([][]byte, 0)

	for i := 0; i < len(n.neededTxs); i++ {
		_, ok := n.dPool.Transactions().SearchFirstData(n.neededTxs[i])
		if !ok {
			//tx is still missing
			newNeededTxs = append(newNeededTxs, n.neededTxs[i])
		}
	}

	n.neededTxs = newNeededTxs

	n.mutNeededTxs.Unlock()
}

func requestMissingTransactions(n *testNode, shardResolver uint32) {
	n.mutNeededTxs.Lock()

	txResolver, _ := n.resFinder.CrossShardResolver(factory.TransactionTopic, shardResolver)

	for i := 0; i < len(n.neededTxs); i++ {
		txResolver.RequestDataFromHash(n.neededTxs[i])
	}

	n.mutNeededTxs.Unlock()
}
