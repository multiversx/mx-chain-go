package metablock

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

// TestHeadersAreReceivedByMetachainAndShard tests if interceptors between shard and metachain work as expected
// (metachain receives shard headers and shard and metachain receive metablocks)
func TestHeadersAreReceivedByMetachainAndShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
	nodesPerShard := 1
	numMetaNodes := 10

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetaNodes,
	)
	integrationTests.DisplayAndStartNodes(nodes)
	fmt.Println(integrationTests.MakeDisplayTable(nodes))

	defer func() {
		for _, n := range nodes {
			_ = n.Messenger.Close()
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
		assert.Equal(t, int32(2), atomic.LoadInt32(&n.CounterHdrRecv))
	}
}

// TestShardHeadersAreResolvedByMetachain tests if resolvers between shard and metachain work as expected
// (metachain request and receives shard headers and shard and metachain request and receives metablocks)
func TestHeadersAreResolvedByMetachainAndShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
	nodesPerShard := 1
	numMetaNodes := 2
	senderShard := uint32(0)

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetaNodes,
	)
	integrationTests.DisplayAndStartNodes(nodes)
	fmt.Println(integrationTests.MakeDisplayTable(nodes))

	defer func() {
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	round := uint64(1)
	nonce := uint64(1)
	fmt.Println("Generating header and block body in shard 0, save it in datapool and metachain creates a request for it...")
	integrationTests.UpdateRound(nodes, round)
	_, hdr, _ := nodes[0].ProposeBlock(round, nonce)
	shardHeaderBytes, _ := integrationTests.TestMarshalizer.Marshal(hdr)
	shardHeaderHash := integrationTests.TestHasher.Compute(string(shardHeaderBytes))
	nodes[0].DataPool.Headers().AddHeader(shardHeaderHash, hdr)

	maxNumRequests := 5
	for i := 0; i < maxNumRequests; i++ {
		for j := 0; j < numMetaNodes; j++ {
			resolver, err := nodes[j+1].ResolverFinder.CrossShardResolver(factory.ShardBlocksTopic, senderShard)
			assert.Nil(t, err)
			_ = resolver.RequestDataFromHash(shardHeaderHash, 0)
		}

		fmt.Println(integrationTests.MakeDisplayTable(nodes))

		time.Sleep(time.Second)
	}

	//all node should have received the shard header
	for _, n := range nodes {
		assert.Equal(t, int32(1), atomic.LoadInt32(&n.CounterHdrRecv))
	}

	fmt.Println("Generating meta header, save it in meta datapools and shard 0 node requests it after its hash...")
	_, metaHdr, _ := nodes[1].ProposeBlock(round, nonce)
	_ = nodes[1].BlockChain.SetCurrentBlockHeader(metaHdr)
	metaHeaderBytes, _ := integrationTests.TestMarshalizer.Marshal(metaHdr)
	metaHeaderHash := integrationTests.TestHasher.Compute(string(metaHeaderBytes))
	nodes[1].BlockChain.SetCurrentBlockHeaderHash(metaHeaderHash)
	_ = nodes[1].Storage.GetStorer(dataRetriever.MetaBlockUnit).Put(metaHeaderHash, metaHeaderBytes)
	for i := 0; i < numMetaNodes; i++ {
		nodes[i+1].DataPool.Headers().AddHeader(metaHeaderHash, metaHdr)
		_ = nodes[i+1].BlockChain.SetCurrentBlockHeader(metaHdr)
		nodes[i+1].BlockChain.SetCurrentBlockHeaderHash(metaHeaderHash)
	}

	for i := 0; i < maxNumRequests; i++ {
		resolver, err := nodes[0].ResolverFinder.MetaChainResolver(factory.MetachainBlocksTopic)
		assert.Nil(t, err)
		_ = resolver.RequestDataFromHash(metaHeaderHash, 0)

		fmt.Println(integrationTests.MakeDisplayTable(nodes))

		time.Sleep(time.Second)
	}

	//all node should have received the meta header
	for _, n := range nodes {
		assert.Equal(t, int32(2), atomic.LoadInt32(&n.CounterHdrRecv))
	}

	fmt.Println("Generating meta header, save it in meta datapools and shard 0 node requests it after its nonce...")
	round++
	nonce++
	integrationTests.UpdateRound(nodes, round)
	_, metaHdr2, _ := nodes[1].ProposeBlock(round, nonce)
	metaHeaderBytes2, _ := integrationTests.TestMarshalizer.Marshal(metaHdr2)
	metaHeaderHash2 := integrationTests.TestHasher.Compute(string(metaHeaderBytes2))
	for i := 0; i < numMetaNodes; i++ {
		nodes[i+1].DataPool.Headers().AddHeader(metaHeaderHash2, metaHdr2)
	}

	for i := 0; i < maxNumRequests; i++ {
		resolver, err := nodes[0].ResolverFinder.MetaChainResolver(factory.MetachainBlocksTopic)
		assert.Nil(t, err)
		_ = resolver.(*resolvers.HeaderResolver).RequestDataFromNonce(metaHdr2.GetNonce(), 0)

		fmt.Println(integrationTests.MakeDisplayTable(nodes))

		time.Sleep(time.Second)
	}

	//all node should have received the meta header
	for _, n := range nodes {
		assert.Equal(t, int32(3), atomic.LoadInt32(&n.CounterHdrRecv))
	}
}
