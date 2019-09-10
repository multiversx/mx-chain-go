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

    "github.com/ElrondNetwork/elrond-go/dataRetriever"
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

    advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
    _ = advertiser.Bootstrap()

    nodes := integrationTests.CreateNodes(
        numOfShards,
        nodesPerShard,
        numMetachainNodes,
        integrationTests.GetConnectableAddress(advertiser),
    )
    integrationTests.CreateAccountForNodes(nodes)
    integrationTests.DisplayAndStartNodes(nodes)

    defer func() {
        _ = advertiser.Close()
        for _, n := range nodes {
            _ = n.Node.Stop()
        }
    }()

    txToSend := 100

    generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), 5)

    fmt.Println("Generating and broadcasting transactions...")
    _, pkInShardFive, _ := integrationTests.GenerateSkAndPkInShard(generateCoordinator, 5)
    pkBytes, _ := pkInShardFive.ToByteArray()
    addrInShardFive := hex.EncodeToString(pkBytes)
    _ = nodes[0].Node.GenerateAndSendBulkTransactions(addrInShardFive, big.NewInt(1), uint64(txToSend))
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

    advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
    _ = advertiser.Bootstrap()

    nodes := integrationTests.CreateNodes(
        numOfShards,
        nodesPerShard,
        numMetachainNodes,
        integrationTests.GetConnectableAddress(advertiser),
    )
    nodes[0] = integrationTests.NewTestProcessorNode(uint32(numOfShards), 0, firstSkInShard, integrationTests.GetConnectableAddress(advertiser))
    integrationTests.CreateAccountForNodes(nodes)
    integrationTests.DisplayAndStartNodes(nodes)

    defer func() {
        _ = advertiser.Close()
        for _, n := range nodes {
            _ = n.Node.Stop()
        }
    }()

    txToSend := 100

    generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), 5)

    _, pkInShardFive, _ := integrationTests.GenerateSkAndPkInShard(generateCoordinator, 5)
    pkBytes, _ := pkInShardFive.ToByteArray()
    addrInShardFive := hex.EncodeToString(pkBytes)

    _ = nodes[0].Node.GenerateAndSendBulkTransactions(addrInShardFive, big.NewInt(1), uint64(txToSend))

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
    numMetachainNodes := 0
    firstSkInShard := uint32(4)

    advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
    _ = advertiser.Bootstrap()

    nodes := integrationTests.CreateNodes(
        numOfShards,
        nodesPerShard,
        numMetachainNodes,
        integrationTests.GetConnectableAddress(advertiser),
    )
    nodes[0] = integrationTests.NewTestProcessorNode(uint32(numOfShards), 0, firstSkInShard, integrationTests.GetConnectableAddress(advertiser))
    integrationTests.CreateAccountForNodes(nodes)
    integrationTests.DisplayAndStartNodes(nodes)

    defer func() {
        _ = advertiser.Close()
        for _, n := range nodes {
            _ = n.Node.Stop()
        }
    }()

    txToSend := 100

    shardRequester := uint32(5)
    randomShard := uint32(2)

    generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), shardRequester)

    _, pkInShardFive, _ := integrationTests.GenerateSkAndPkInShard(generateCoordinator, 5)
    pkBytes, _ := pkInShardFive.ToByteArray()
    addrInShardFive := hex.EncodeToString(pkBytes)

    mutGeneratedTxHashes := sync.Mutex{}
    generatedTxHashes := make([][]byte, 0)
    //wire a new hook for generated txs on a node in sender shard to populate tx hashes generated
    for _, n := range nodes {
        if n.ShardCoordinator.SelfId() == firstSkInShard {
            n.ShardDataPool.Transactions().RegisterHandler(func(key []byte) {
                mutGeneratedTxHashes.Lock()
                generatedTxHashes = append(generatedTxHashes, key)
                mutGeneratedTxHashes.Unlock()
            })
        }
    }

    _ = nodes[0].Node.GenerateAndSendBulkTransactions(addrInShardFive, big.NewInt(1), uint64(txToSend))

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

func TestNode_InMultiShardEnvRequestTxsShouldRequireOnlyFromTheOtherShard(t *testing.T) {
    if testing.Short() {
        t.Skip("this is not a short test")
    }

    advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
    _ = advertiser.Bootstrap()

    nodes := make([]*integrationTests.TestProcessorNode, 0)
    maxShards := 2
    nodesPerShard := 2
    txGenerated := 10

    defer func() {
        _ = advertiser.Close()
        for _, n := range nodes {
            _ = n.Node.Stop()
        }
    }()

    //shard 0, requesters
    recvTxs := make(map[int]map[string]struct{})
    mutRecvTxs := sync.Mutex{}
    for i := 0; i < nodesPerShard; i++ {
        dPool := integrationTests.CreateRequesterDataPool(t, recvTxs, &mutRecvTxs, i, uint32(maxShards))

        tn := integrationTests.NewTestProcessorNodeWithCustomDataPool(
            uint32(maxShards),
            0,
            0,
            integrationTests.GetConnectableAddress(advertiser),
            dPool,
        )

        nodes = append(nodes, tn)
    }

    var txHashesGenerated [][]byte
    var dPool dataRetriever.PoolsHolder
    shardCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(maxShards), 0)
    dPool, txHashesGenerated = integrationTests.CreateResolversDataPool(t, txGenerated, 0, 1, shardCoordinator)
    //shard 1, resolvers, same data pool, does not matter
    for i := 0; i < nodesPerShard; i++ {
        tn := integrationTests.NewTestProcessorNodeWithCustomDataPool(
            uint32(maxShards),
            1,
            1,
            integrationTests.GetConnectableAddress(advertiser),
            dPool,
        )

        atomic.StoreInt32(&tn.CounterTxRecv, int32(txGenerated))

        nodes = append(nodes, tn)
    }

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

        _ = txResolver.RequestDataFromHashArray(txHashesGenerated)
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
