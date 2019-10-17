package metablock

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

// TestHeadersAreReceivedByMetachainAndShard tests if interceptors between shard and metachain work as expected
// (metachain receives shard headers and shard and metachain receive metablocks)
func TestHeadersAreReceivedByMetachainAndShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	numOfShards := 1
	nodesPerShard := 1
	numMetaNodes := 10

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetaNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)
	integrationTests.DisplayAndStartNodes(nodes)
	fmt.Println(integrationTests.MakeDisplayTable(nodes))

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Node.Stop()
		}
	}()

	fmt.Println("Generating header and block body from shard 0 to metachain...")
	integrationTests.ProposeBlockSignalsEmptyBlock(nodes[0], 1, 1)

	for i := 0; i < 5; i++ {
		fmt.Println(integrationTests.MakeDisplayTable(nodes))
		time.Sleep(time.Second)
	}

	//all node should have received the shard header
	for _, n := range nodes {
		assert.Equal(t, int32(1), atomic.LoadInt32(&n.CounterHdrRecv))
	}

	fmt.Println("Generating metaheader from metachain to any other shards...")
	integrationTests.ProposeBlockSignalsEmptyBlock(nodes[1], 1, 1)

	for i := 0; i < 5; i++ {
		fmt.Println(integrationTests.MakeDisplayTable(nodes))
		time.Sleep(time.Second)
	}

	//all node should have received the meta header
	for _, n := range nodes {
		assert.Equal(t, int32(1), atomic.LoadInt32(&n.CounterMetaRcv))
	}
}

// TestShardHeadersAreResolvedByMetachain tests if resolvers between shard and metachain work as expected
// (metachain request and receives shard headers and shard and metachain request and receives metablocks)
func TestHeadersAreResolvedByMetachainAndShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	numOfShards := 1
	nodesPerShard := 1
	numMetaNodes := 2
	senderShard := uint32(0)

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetaNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)
	integrationTests.DisplayAndStartNodes(nodes)
	fmt.Println(integrationTests.MakeDisplayTable(nodes))

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Node.Stop()
		}
	}()

	fmt.Println("Generating header and block body in shard 0, save it in datapool and metachain creates a request for it...")
	_, hdr, _ := nodes[0].ProposeBlock(1, 1)
	shardHeaderBytes, _ := integrationTests.TestMarshalizer.Marshal(hdr)
	shardHeaderHash := integrationTests.TestHasher.Compute(string(shardHeaderBytes))
	nodes[0].ShardDataPool.Headers().HasOrAdd(shardHeaderHash, hdr)

	maxNumRequests := 5
	for i := 0; i < maxNumRequests; i++ {
		for j := 0; j < numMetaNodes; j++ {
			resolver, err := nodes[j+1].ResolverFinder.CrossShardResolver(factory.ShardHeadersForMetachainTopic, senderShard)
			assert.Nil(t, err)
			_ = resolver.RequestDataFromHash(shardHeaderHash)
		}

		fmt.Println(integrationTests.MakeDisplayTable(nodes))

		time.Sleep(time.Second)
	}

	//all node should have received the shard header
	for _, n := range nodes {
		assert.Equal(t, int32(1), atomic.LoadInt32(&n.CounterHdrRecv))
	}

	fmt.Println("Generating meta header, save it in meta datapools and shard 0 node requests it after its hash...")
	_, metaHdr, _ := nodes[1].ProposeBlock(1, 1)
	metaHeaderBytes, _ := integrationTests.TestMarshalizer.Marshal(metaHdr)
	metaHeaderHash := integrationTests.TestHasher.Compute(string(metaHeaderBytes))
	for i := 0; i < numMetaNodes; i++ {
		nodes[i+1].MetaDataPool.MetaBlocks().HasOrAdd(metaHeaderHash, metaHdr)
	}

	for i := 0; i < maxNumRequests; i++ {
		resolver, err := nodes[0].ResolverFinder.MetaChainResolver(factory.MetachainBlocksTopic)
		assert.Nil(t, err)
		_ = resolver.RequestDataFromHash(metaHeaderHash)

		fmt.Println(integrationTests.MakeDisplayTable(nodes))

		time.Sleep(time.Second)
	}

	//all node should have received the meta header
	for _, n := range nodes {
		assert.Equal(t, int32(1), atomic.LoadInt32(&n.CounterMetaRcv))
	}

	fmt.Println("Generating meta header, save it in meta datapools and shard 0 node requests it after its nonce...")
	_, metaHdr2, _ := nodes[1].ProposeBlock(2, 2)
	metaHdr2.SetNonce(64)
	metaHeaderBytes2, _ := integrationTests.TestMarshalizer.Marshal(metaHdr2)
	metaHeaderHash2 := integrationTests.TestHasher.Compute(string(metaHeaderBytes2))
	for i := 0; i < numMetaNodes; i++ {
		nodes[i+1].MetaDataPool.MetaBlocks().HasOrAdd(metaHeaderHash2, metaHdr2)

		syncMap := &dataPool.ShardIdHashSyncMap{}
		syncMap.Store(sharding.MetachainShardId, metaHeaderHash2)
		nodes[i+1].MetaDataPool.HeadersNonces().Merge(metaHdr2.GetNonce(), syncMap)
	}

	for i := 0; i < maxNumRequests; i++ {
		resolver, err := nodes[0].ResolverFinder.MetaChainResolver(factory.MetachainBlocksTopic)
		assert.Nil(t, err)
		_ = resolver.(*resolvers.HeaderResolver).RequestDataFromNonce(metaHdr2.GetNonce())

		fmt.Println(integrationTests.MakeDisplayTable(nodes))

		time.Sleep(time.Second)
	}

	//all node should have received the meta header
	for _, n := range nodes {
		assert.Equal(t, int32(2), atomic.LoadInt32(&n.CounterMetaRcv))
	}
}
