package block

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
	"github.com/whyrusleeping/go-logging"
)

func init() {
	logging.SetLevel(logging.ERROR, "pubsub")
}

// TestHeaderAndMiniBlocksAreRoutedCorrectly tests what happens if a shard node broadcasts a header and a
// body with 3 miniblocks. One miniblock will be an intra-shard type and the other 2 will be cross-shard type.
func TestHeaderAndMiniBlocksAreRoutedCorrectly(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 6
	nodesPerShard := 3

	senderShard := uint32(0)
	recvShards := []uint32{1, 2}

	advertiser := createMessengerWithKadDht(context.Background(), "")
	advertiser.Bootstrap()

	nodes := createNodes(
		numOfShards,
		nodesPerShard,
		getConnectableAddress(advertiser),
	)
	displayAndStartNodes(nodes)

	defer func() {
		advertiser.Close()
		for _, nodeList := range nodes {
			for _, n := range nodeList {
				n.node.Stop()
			}
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(time.Second * 5)

	proposerNode := nodes[0][0]

	fmt.Println("Generating header and block body...")
	body, hdr := generateHeaderAndBody(senderShard, recvShards...)
	err := proposerNode.broadcastMessenger.BroadcastBlock(body, hdr)
	assert.Nil(t, err)
	miniBlocks, _, _ := proposerNode.blkProcessor.MarshalizedDataToBroadcast(hdr, body)
	err = proposerNode.broadcastMessenger.BroadcastMiniBlocks(miniBlocks)
	assert.Nil(t, err)

	time.Sleep(time.Second * 10)

	for shardId, nodeList := range nodes {
		for _, n := range nodeList {
			isSenderShard := shardId == senderShard
			isRecvShard := uint32InSlice(n.shardId, recvShards)
			isRecvMetachain := n.shardId == sharding.MetachainShardId

			assert.Equal(t, int32(0), atomic.LoadInt32(&n.metachainHdrRecv))

			if isSenderShard {
				assert.Equal(t, int32(1), atomic.LoadInt32(&n.headersRecv))

				shards := []uint32{senderShard}
				shards = append(shards, recvShards...)

				expectedMiniblocks := getMiniBlocksHashesFromShardIds(body.(block.Body), shards...)

				assert.True(t, equalSlices(expectedMiniblocks, n.miniblocksHashes))
			}

			if isRecvShard && !isSenderShard {
				assert.Equal(t, int32(0), atomic.LoadInt32(&n.headersRecv))
				expectedMiniblocks := getMiniBlocksHashesFromShardIds(body.(block.Body), n.shardId)
				assert.True(t, equalSlices(expectedMiniblocks, n.miniblocksHashes))
			}

			if !isSenderShard && !isRecvShard && !isRecvMetachain {
				//other nodes should have not received neither the header nor the miniblocks
				assert.Equal(t, int32(0), atomic.LoadInt32(&n.headersRecv))
				assert.Equal(t, int32(0), atomic.LoadInt32(&n.miniblocksRecv))
			}
		}
	}

	fmt.Println(makeDisplayTable(nodes))
}

func generateHeaderAndBody(senderShard uint32, recvShards ...uint32) (data.BodyHandler, data.HeaderHandler) {
	hdr := block.Header{
		Nonce:            0,
		PubKeysBitmap:    []byte{255, 0},
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
