package interceptedBlocks

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	testBlock "github.com/ElrondNetwork/elrond-go/integrationTests/multiShard/block"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)
	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	fmt.Println("Generating header and block body...")
	_, body, _ := integrationTests.ProposeBlockSignalsEmptyBlock(nodes[0], 1, 1)

	time.Sleep(time.Second * 10)

	for _, n := range nodes {
		isSenderShard := n.ShardCoordinator.SelfId() == senderShard
		isRecvShard := integrationTests.Uint32InSlice(n.ShardCoordinator.SelfId(), recvShards)
		isRecvMetachain := n.ShardCoordinator.SelfId() == core.MetachainShardId

		assert.Equal(t, int32(0), atomic.LoadInt32(&n.CounterMetaRcv))

		if isSenderShard {
			assert.Equal(t, int32(1), atomic.LoadInt32(&n.CounterHdrRecv))

			shards := []uint32{senderShard}
			shards = append(shards, recvShards...)

			expectedMiniblocks := integrationTests.GetMiniBlocksHashesFromShardIds(body.(*block.Body), shards...)
			assert.True(t, n.MiniBlocksPresent(expectedMiniblocks))
		}

		if isRecvShard && !isSenderShard {
			assert.Equal(t, int32(0), atomic.LoadInt32(&n.CounterHdrRecv))
			expectedMiniblocks := integrationTests.GetMiniBlocksHashesFromShardIds(body.(*block.Body), n.ShardCoordinator.SelfId())
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

// TestMetaHeadersAreRequsted tests the metaheader request to be made towards any peer
// The test will have 2 shards and meta, one shard node will request a metaheader and it should received it only from
// the meta node
func TestMetaHeadersAreRequsted(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	maxShards := uint32(2)

	node1Shard0Requester := integrationTests.NewTestProcessorNode(maxShards, 0, 0)
	node2Shard0 := integrationTests.NewTestProcessorNode(maxShards, 0, 0)
	node3Shard1 := integrationTests.NewTestProcessorNode(maxShards, 1, 0)
	node4Meta := integrationTests.NewTestProcessorNode(maxShards, core.MetachainShardId, 0)

	nodes := []*integrationTests.TestProcessorNode{node1Shard0Requester, node2Shard0, node3Shard1, node4Meta}
	connectableNodes := make([]integrationTests.Connectable, 0, len(nodes))
	for _, n := range nodes {
		connectableNodes = append(connectableNodes, n)
	}

	integrationTests.ConnectNodes(connectableNodes)

	defer func() {
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(integrationTests.P2pBootstrapDelay)

	metaHdrFromMetachain := &block.MetaBlock{
		Nonce:           1,
		Round:           1,
		Epoch:           0,
		ShardInfo:       nil,
		Signature:       []byte("signature"),
		PubKeysBitmap:   []byte{1},
		PrevHash:        []byte("prev hash"),
		PrevRandSeed:    []byte("prev rand seed"),
		RandSeed:        []byte("rand seed"),
		RootHash:        []byte("root hash"),
		TxCount:         0,
		ChainID:         integrationTests.ChainID,
		SoftwareVersion: integrationTests.SoftwareVersion,
	}
	metaHdrHashFromMetachain, _ := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, metaHdrFromMetachain)

	metaHdrFromShard := &block.MetaBlock{
		Nonce:           1,
		Round:           2,
		Epoch:           0,
		ShardInfo:       nil,
		Signature:       []byte("signature"),
		PubKeysBitmap:   []byte{1},
		PrevHash:        []byte("prev hash"),
		PrevRandSeed:    []byte("prev rand seed"),
		RandSeed:        []byte("rand seed"),
		RootHash:        []byte("root hash"),
		TxCount:         0,
		ChainID:         integrationTests.ChainID,
		SoftwareVersion: integrationTests.SoftwareVersion,
	}
	metaHdrFromShardHash, _ := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, metaHdrFromShard)

	for _, n := range nodes {
		n.DataPool.Headers().AddHeader(metaHdrFromShardHash, metaHdrFromShard)
		n.DataPool.Headers().AddHeader(metaHdrHashFromMetachain, metaHdrFromMetachain)
	}

	chanReceived := make(chan struct{}, 1000)
	node1Shard0Requester.DataPool.Headers().Clear()
	node1Shard0Requester.DataPool.Headers().RegisterHandler(func(header data.HeaderHandler, key []byte) {
		chanReceived <- struct{}{}
	})

	retrievedHeader := requestAndRetrieveMetaHeader(node1Shard0Requester, metaHdrHashFromMetachain, chanReceived)
	require.NotNil(t, retrievedHeader)
	assert.Equal(t, metaHdrFromMetachain.Nonce, retrievedHeader.Nonce)

	retrievedHeader = requestAndRetrieveMetaHeader(node1Shard0Requester, metaHdrFromShardHash, chanReceived)
	assert.Nil(t, retrievedHeader)
}

// TestMetaHeadersAreRequestedByAMetachainNode tests the metaheader request is served by a metachain node and
// by a shard node
func TestMetaHeadersAreRequestedByAMetachainNode(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	maxShards := uint32(2)

	node1Shard0 := integrationTests.NewTestProcessorNode(maxShards, 0, 0)
	node2MetaRequester := integrationTests.NewTestProcessorNode(maxShards, core.MetachainShardId, 0)
	node3MetaResolver := integrationTests.NewTestProcessorNode(maxShards, core.MetachainShardId, 0)

	nodes := []*integrationTests.TestProcessorNode{node1Shard0, node2MetaRequester, node3MetaResolver}
	connectableNodes := make([]integrationTests.Connectable, 0, len(nodes))
	for _, n := range nodes {
		connectableNodes = append(connectableNodes, n)
	}

	integrationTests.ConnectNodes(connectableNodes)

	//update to a "safe" round number
	integrationTests.UpdateRound(nodes, 3)

	defer func() {
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(integrationTests.P2pBootstrapDelay)

	metaBlock1 := &block.MetaBlock{
		Nonce:           1,
		Round:           1,
		Epoch:           0,
		ShardInfo:       nil,
		Signature:       []byte("signature"),
		PubKeysBitmap:   []byte{1},
		PrevHash:        []byte("prev hash"),
		PrevRandSeed:    []byte("prev rand seed"),
		RandSeed:        []byte("rand seed"),
		RootHash:        []byte("root hash"),
		TxCount:         0,
		ChainID:         integrationTests.ChainID,
		SoftwareVersion: integrationTests.SoftwareVersion,
	}
	metaBlock1Hash, _ := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, metaBlock1)

	metaBlock2 := &block.MetaBlock{
		Nonce:           2,
		Round:           2,
		Epoch:           0,
		ShardInfo:       nil,
		Signature:       []byte("signature"),
		PubKeysBitmap:   []byte{1},
		PrevHash:        []byte("prev hash"),
		PrevRandSeed:    []byte("prev rand seed"),
		RandSeed:        []byte("rand seed"),
		RootHash:        []byte("root hash"),
		TxCount:         0,
		ChainID:         integrationTests.ChainID,
		SoftwareVersion: integrationTests.SoftwareVersion,
	}
	metaBlock2Hash, _ := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, metaBlock2)

	node1Shard0.DataPool.Headers().AddHeader(metaBlock1Hash, metaBlock1)
	node3MetaResolver.DataPool.Headers().AddHeader(metaBlock2Hash, metaBlock2)

	retrievedHeaders := requestAndRetrieveMetaHeadersByNonce(node2MetaRequester, metaBlock1.Nonce)
	require.Equal(t, 1, len(retrievedHeaders))
	assert.Equal(t, metaBlock1.Nonce, retrievedHeaders[0].GetNonce())

	retrievedHeaders = requestAndRetrieveMetaHeadersByNonce(node2MetaRequester, metaBlock2.Nonce)
	require.Equal(t, 1, len(retrievedHeaders))
	assert.Equal(t, metaBlock2.Nonce, retrievedHeaders[0].GetNonce())
}

func requestAndRetrieveMetaHeader(
	node *integrationTests.TestProcessorNode,
	hash []byte,
	chanReceived chan struct{},
) *block.MetaBlock {

	resolver, _ := node.ResolverFinder.MetaChainResolver(factory.MetachainBlocksTopic)
	_ = resolver.RequestDataFromHash(hash, 0)

	select {
	case <-chanReceived:
	case <-time.After(testBlock.StepDelay):
		return nil
	}

	retrievedObject, _ := node.DataPool.Headers().GetHeaderByHash(hash)

	return retrievedObject.(*block.MetaBlock)
}

func requestAndRetrieveMetaHeadersByNonce(
	node *integrationTests.TestProcessorNode,
	nonce uint64,
) []data.HeaderHandler {

	node.RequestHandler.RequestMetaHeaderByNonce(nonce)
	time.Sleep(time.Second * 2)
	retrievedObject, _, _ := node.DataPool.Headers().GetHeadersByNonceAndShardId(nonce, core.MetachainShardId)

	return retrievedObject
}
