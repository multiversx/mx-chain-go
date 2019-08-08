package block

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
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

	// delay for bootstrapping and topic announcement
	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(time.Second * 5)

	fmt.Println("Generating header and block body...")
	body, hdr := integrationTests.GenerateDefaultHeaderAndBody(senderShard, recvShards...)
	err := nodes[0].BroadcastMessenger.BroadcastBlock(body, hdr)
	assert.Nil(t, err)
	miniBlocks, _, _ := nodes[0].BlockProcessor.MarshalizedDataToBroadcast(hdr, body)
	err = nodes[0].BroadcastMessenger.BroadcastMiniBlocks(miniBlocks)
	assert.Nil(t, err)

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

			assert.True(t, integrationTests.EqualSlices(expectedMiniblocks, n.MiniBlocksHashes))
		}

		if isRecvShard && !isSenderShard {
			assert.Equal(t, int32(0), atomic.LoadInt32(&n.CounterHdrRecv))
			expectedMiniblocks := integrationTests.GetMiniBlocksHashesFromShardIds(body.(block.Body), n.ShardCoordinator.SelfId())
			assert.True(t, integrationTests.EqualSlices(expectedMiniblocks, n.MiniBlocksHashes))
		}

		if !isSenderShard && !isRecvShard && !isRecvMetachain {
			//other nodes should have not received neither the header nor the miniblocks
			assert.Equal(t, int32(0), atomic.LoadInt32(&n.CounterHdrRecv))
			assert.Equal(t, int32(0), atomic.LoadInt32(&n.CounterMbRecv))
		}
	}

	fmt.Println(integrationTests.MakeDisplayTable(nodes))
}
