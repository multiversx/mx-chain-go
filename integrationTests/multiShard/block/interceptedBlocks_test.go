package block

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

// TestHeaderAndMiniBlocksAreRoutedCorrectly tests what happens if a shard node broadcasts a header and a
// body with 3 miniblocks. One miniblock will be an intra-shard type and the other 2 will be cross-shard type.
func TestHeaderAndMiniBlocksAreRoutedCorrectly(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 6
	nodesPerShard := 3
	numMetachainNodes := 1

	senderShard := uint32(0)
	recvShards := []uint32{1, 2}

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)
	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Node.Stop()
		}
	}()

	fmt.Println("Generating header and block body...")
	_, body, _ := integrationTests.ProposeBlockSignalsEmptyBlock(nodes[0], 1)

	time.Sleep(time.Second * 10)

	for _, n := range nodes {
		isSenderShard := n.ShardCoordinator.SelfId() == senderShard
		isRecvShard := integrationTests.Uint32InSlice(n.ShardCoordinator.SelfId(), recvShards)
		isRecvMetachain := n.ShardCoordinator.SelfId() == sharding.MetachainShardId

		assert.Equal(t, int32(0), atomic.LoadInt32(&n.CounterMetaRcv))

		if isSenderShard {
			assert.Equal(t, int32(1), atomic.LoadInt32(&n.CounterHdrRecv))

			shards := []uint32{senderShard}
			shards = append(shards, recvShards...)

			expectedMiniblocks := integrationTests.GetMiniBlocksHashesFromShardIds(body.(block.Body), shards...)
			assert.True(t, n.MiniBlocksPresent(expectedMiniblocks))
		}

		if isRecvShard && !isSenderShard {
			assert.Equal(t, int32(0), atomic.LoadInt32(&n.CounterHdrRecv))
			expectedMiniblocks := integrationTests.GetMiniBlocksHashesFromShardIds(body.(block.Body), n.ShardCoordinator.SelfId())
			assert.True(t, n.MiniBlocksPresent(expectedMiniblocks))
		}

		if !isSenderShard && !isRecvShard && !isRecvMetachain {
			//other nodes should have not received neither the header nor the miniblocks
			assert.Equal(t, int32(0), atomic.LoadInt32(&n.CounterHdrRecv))
			assert.Equal(t, int32(0), atomic.LoadInt32(&n.CounterMbRecv))
		}
	}

	fmt.Println(integrationTests.MakeDisplayTable(nodes))
}

// TestMetaHeadersAreRequstedOnlyFromMetachain tests the metaheader request to be made only from metachain nodes
// The test will have 2 shards and meta, one shard node will request a metaheader and it should received it only from
// the meta node
func TestMetaHeadersAreRequstedOnlyFromMetachain(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	maxShards := uint32(2)
	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)

	node1Shard0 := integrationTests.NewTestProcessorNode(maxShards, 0, 0, advertiserAddr)
	node2Shard0 := integrationTests.NewTestProcessorNode(maxShards, 0, 0, advertiserAddr)
	node3Shard1 := integrationTests.NewTestProcessorNode(maxShards, 1, 0, advertiserAddr)
	node4Meta := integrationTests.NewTestProcessorNode(maxShards, sharding.MetachainShardId, 0, advertiserAddr)

	nodes := []*integrationTests.TestProcessorNode{node1Shard0, node2Shard0, node3Shard1, node4Meta}

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	for _, n := range nodes {
		_ = n.Messenger.Bootstrap()
	}

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(time.Second * 5)

	metaHdrFromMetachain := &block.MetaBlock{
		Nonce:         110,
		Epoch:         0,
		ShardInfo:     make([]block.ShardData, 0),
		PeerInfo:      make([]block.PeerData, 0),
		Signature:     []byte("signature"),
		PubKeysBitmap: []byte{1},
		PrevHash:      []byte("prev hash"),
		PrevRandSeed:  []byte("prev rand seed"),
		RandSeed:      []byte("rand seed"),
		RootHash:      []byte("root hash"),
		TxCount:       0,
	}
	metaHdrHashFromMetachain, _ := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, metaHdrFromMetachain)

	metaHdrFromShard := &block.MetaBlock{
		Nonce:         220,
		Epoch:         0,
		ShardInfo:     make([]block.ShardData, 0),
		PeerInfo:      make([]block.PeerData, 0),
		Signature:     []byte("signature"),
		PubKeysBitmap: []byte{1},
		PrevHash:      []byte("prev hash"),
		PrevRandSeed:  []byte("prev rand seed"),
		RandSeed:      []byte("rand seed"),
		RootHash:      []byte("root hash"),
		TxCount:       0,
	}
	metaHdrFromShardHash, _ := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, metaHdrFromShard)

	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() != sharding.MetachainShardId {
			n.ShardDataPool.MetaBlocks().Put(metaHdrFromShardHash, metaHdrFromShard)
		}
	}

	chanReceived := make(chan struct{}, 1000)
	node4Meta.MetaDataPool.MetaChainBlocks().Put(metaHdrHashFromMetachain, metaHdrFromMetachain)
	node1Shard0.ShardDataPool.MetaBlocks().Clear()
	node1Shard0.ShardDataPool.MetaBlocks().RegisterHandler(func(key []byte) {
		chanReceived <- struct{}{}
	})

	retrievedHeader := requestAndRetrieveMetaHeader(node1Shard0, metaHdrHashFromMetachain, chanReceived)
	assert.NotNil(t, retrievedHeader)
	assert.Equal(t, metaHdrFromMetachain.Nonce, retrievedHeader.Nonce)

	retrievedHeader = requestAndRetrieveMetaHeader(node1Shard0, metaHdrFromShardHash, chanReceived)
	assert.Nil(t, retrievedHeader)
}

func requestAndRetrieveMetaHeader(
	node *integrationTests.TestProcessorNode,
	hash []byte,
	chanReceived chan struct{},
) *block.MetaBlock {

	resolver, _ := node.ResolverFinder.MetaChainResolver(factory.MetachainBlocksTopic)
	_ = resolver.RequestDataFromHash(hash)

	select {
	case <-chanReceived:
	case <-time.After(time.Second * 2):
		return nil
	}

	retrievedObject, _ := node.ShardDataPool.MetaBlocks().Get(hash)

	return retrievedObject.(*block.MetaBlock)
}
