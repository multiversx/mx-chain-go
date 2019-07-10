package metablock

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
	"github.com/whyrusleeping/go-logging"
)

func init() {
	logging.SetLevel(logging.ERROR, "pubsub")
}

// TestHeadersAreReceivedByMetachainAndShard tests if interceptors between shard and metachain work as expected
// (metachain receives shard headers and shard and metachain receive metablocks)
func TestHeadersAreReceivedByMetachainAndShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	advertiser := createMessengerWithKadDht(context.Background(), "")
	advertiser.Bootstrap()

	numMetaNodes := 10
	senderShard := uint32(0)

	nodes := createNodes(
		numMetaNodes,
		senderShard,
		getConnectableAddress(advertiser),
		testHasher,
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
	fmt.Println(makeDisplayTable(nodes))

	fmt.Println("Generating header and block body from shard 0 to metachain...")
	body, hdr := generateHeaderAndBody(senderShard)
	err := nodes[0].broadcastMessenger.BroadcastBlock(body, hdr)
	assert.Nil(t, err)
	err = nodes[0].broadcastMessenger.BroadcastHeader(hdr)
	assert.Nil(t, err)

	for i := 0; i < 5; i++ {
		fmt.Println(makeDisplayTable(nodes))

		time.Sleep(time.Second)
	}

	//all node should have received the shard header
	for _, n := range nodes {
		assert.Equal(t, int32(1), atomic.LoadInt32(&n.shardHdrRecv))
	}

	fmt.Println("Generating metaheader from metachain to any other shards...")
	metaHdr := generateMetaHeader()
	nodes[1].broadcastMessenger.BroadcastBlock(nil, metaHdr)

	for i := 0; i < 5; i++ {
		fmt.Println(makeDisplayTable(nodes))
		time.Sleep(time.Second)
	}

	//all node should have received the meta header
	for _, n := range nodes {
		assert.Equal(t, int32(1), atomic.LoadInt32(&n.metachainHdrRecv))
	}
}

// TestShardHeadersAreResolvedByMetachain tests if resolvers between shard and metachain work as expected
// (metachain request and receives shard headers and shard and metachain request and receives metablocks)
func TestHeadersAreResolvedByMetachainAndShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	advertiser := createMessengerWithKadDht(context.Background(), "")
	advertiser.Bootstrap()

	numMetaNodes := 2
	senderShard := uint32(0)

	nodes := createNodes(
		numMetaNodes,
		senderShard,
		getConnectableAddress(advertiser),
		testHasher,
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
	fmt.Println(makeDisplayTable(nodes))

	fmt.Println("Generating header and block body in shard 0, save it in datapool and metachain creates a request for it...")
	_, hdr := generateHeaderAndBody(senderShard)
	shardHeaderBytes, _ := testMarshalizer.Marshal(hdr)
	shardHeaderHash := testHasher.Compute(string(shardHeaderBytes))
	nodes[0].shardDataPool.Headers().HasOrAdd(shardHeaderHash, hdr)

	for i := 0; i < 5; i++ {
		//continuously request
		for j := 0; j < numMetaNodes; j++ {
			resolver, err := nodes[j+1].resolvers.CrossShardResolver(factory.ShardHeadersForMetachainTopic, senderShard)
			assert.Nil(t, err)
			resolver.RequestDataFromHash(shardHeaderHash)
		}

		fmt.Println(makeDisplayTable(nodes))

		time.Sleep(time.Second)
	}

	//all node should have received the shard header
	for _, n := range nodes {
		assert.Equal(t, int32(1), atomic.LoadInt32(&n.shardHdrRecv))
	}

	fmt.Println("Generating meta header, save it in meta datapools and shard 0 node requests it after its hash...")
	metaHdr := generateMetaHeader()
	metaHeaderBytes, _ := testMarshalizer.Marshal(metaHdr)
	metaHeaderHash := testHasher.Compute(string(metaHeaderBytes))
	for i := 0; i < numMetaNodes; i++ {
		nodes[i+1].metaDataPool.MetaChainBlocks().HasOrAdd(metaHeaderHash, metaHdr)
	}

	for i := 0; i < 5; i++ {
		//continuously request
		resolver, err := nodes[0].resolvers.MetaChainResolver(factory.MetachainBlocksTopic)
		assert.Nil(t, err)
		resolver.RequestDataFromHash(metaHeaderHash)

		fmt.Println(makeDisplayTable(nodes))

		time.Sleep(time.Second)
	}

	//all node should have received the meta header
	for _, n := range nodes {
		assert.Equal(t, int32(1), atomic.LoadInt32(&n.metachainHdrRecv))
	}

	fmt.Println("Generating meta header, save it in meta datapools and shard 0 node requests it after its nonce...")
	metaHdr2 := generateMetaHeader()
	metaHdr2.SetNonce(64)
	metaHeaderBytes2, _ := testMarshalizer.Marshal(metaHdr2)
	metaHeaderHash2 := testHasher.Compute(string(metaHeaderBytes2))
	for i := 0; i < numMetaNodes; i++ {
		nodes[i+1].metaDataPool.MetaChainBlocks().HasOrAdd(metaHeaderHash2, metaHdr2)
		nodes[i+1].metaDataPool.MetaBlockNonces().HasOrAdd(metaHdr2.GetNonce(), metaHeaderHash2)
	}

	for i := 0; i < 5; i++ {
		//continuously request
		resolver, err := nodes[0].resolvers.MetaChainResolver(factory.MetachainBlocksTopic)
		assert.Nil(t, err)
		resolver.(*resolvers.HeaderResolver).RequestDataFromNonce(metaHdr2.GetNonce())

		fmt.Println(makeDisplayTable(nodes))

		time.Sleep(time.Second)
	}

	//all node should have received the meta header
	for _, n := range nodes {
		assert.Equal(t, int32(2), atomic.LoadInt32(&n.metachainHdrRecv))
	}
}

func generateHeaderAndBody(senderShard uint32, recvShards ...uint32) (data.BodyHandler, data.HeaderHandler) {
	hdr := block.Header{
		Nonce:            0,
		PubKeysBitmap:    []byte{1, 0, 0},
		Signature:        []byte("signature"),
		PrevHash:         []byte("prev hash"),
		TimeStamp:        uint64(time.Now().Unix()),
		Round:            1,
		Epoch:            2,
		ShardId:          senderShard,
		BlockBodyType:    block.TxBlock,
		RootHash:         []byte{255, 255},
		PrevRandSeed:     make([]byte, 0),
		RandSeed:         make([]byte, 0),
		MiniBlockHeaders: make([]block.MiniBlockHeader, 0),
	}

	body := block.Body{
		&block.MiniBlock{
			SenderShardID:   senderShard,
			ReceiverShardID: senderShard,
			TxHashes: [][]byte{
				testHasher.Compute("tx1"),
			},
		},
	}

	for i, recvShard := range recvShards {
		body = append(
			body,
			&block.MiniBlock{
				SenderShardID:   senderShard,
				ReceiverShardID: recvShard,
				TxHashes: [][]byte{
					testHasher.Compute(fmt.Sprintf("tx%d", i)),
				},
			},
		)
	}

	return body, &hdr
}

func generateMetaHeader() data.HeaderHandler {
	hdr := block.MetaBlock{
		Nonce:         0,
		PubKeysBitmap: []byte{1, 0, 0},
		Signature:     []byte("signature"),
		PrevHash:      []byte("prev hash"),
		TimeStamp:     uint64(time.Now().Unix()),
		Round:         1,
		Epoch:         2,
		RootHash:      []byte{255, 255},
		PrevRandSeed:  make([]byte, 0),
		RandSeed:      make([]byte, 0),
		ShardInfo:     make([]block.ShardData, 0),
	}

	return &hdr
}
