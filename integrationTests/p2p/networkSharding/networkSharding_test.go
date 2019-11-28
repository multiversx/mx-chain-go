package networkSharding

import (
	"context"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
)

var p2pBootstrapStepDelay = 2 * time.Second

func TestConnectionsInNetworkSharding(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodesPerShard := 7
	nbMetaNodes := 7
	nbShards := 5
	consensusGroupSize := 2

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()
	seedAddress := integrationTests.GetConnectableAddress(advertiser)

	targetConnections := 12

	// create map of shard - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithTestP2PNodes(
		nodesPerShard,
		nbMetaNodes,
		nbShards,
		consensusGroupSize,
		consensusGroupSize,
		seedAddress,
		targetConnections,
	)

	defer func() {
		stopNodes(advertiser, nodesMap)
	}()

	createTestInterceptorForEachNode(nodesMap)

	time.Sleep(time.Second * 2)

	startNodes(nodesMap)

	t.Log("Delaying for node bootstrap and topic announcement...")
	time.Sleep(p2pBootstrapStepDelay)

	for i := 0; i < 15; i++ {
		t.Log("\n" + integrationTests.MakeDisplayTableForP2PNodes(nodesMap))

		time.Sleep(time.Second)
	}

	sendMessageOnGlobalTopic(t, nodesMap)
	sendMessagesOnIntraShardTopic(t, nodesMap)
	sendMessagesOnCrossShardTopic(t, nodesMap)

	for i := 0; i < 2; i++ {
		t.Log("\n" + integrationTests.MakeDisplayTableForP2PNodes(nodesMap))

		time.Sleep(time.Second)
	}

	testCounters(t, nodesMap, 1, 1, nbShards*2)
}

func stopNodes(advertiser p2p.Messenger, nodesMap map[uint32][]*integrationTests.TestP2PNode) {
	_ = advertiser.Close()
	for _, nodes := range nodesMap {
		for _, n := range nodes {
			_ = n.Node.Stop()
		}
	}
}

func startNodes(nodesMap map[uint32][]*integrationTests.TestP2PNode) {
	for _, nodes := range nodesMap {
		for _, n := range nodes {
			_ = n.Node.Start()
		}
	}
}

func createTestInterceptorForEachNode(nodesMap map[uint32][]*integrationTests.TestP2PNode) {
	for _, nodes := range nodesMap {
		for _, n := range nodes {
			n.CreateTestInterceptors()
		}
	}
}

func sendMessageOnGlobalTopic(t *testing.T, nodesMap map[uint32][]*integrationTests.TestP2PNode) {
	t.Log("sending a message on global topic")
	nodesMap[0][0].Messenger.Broadcast(integrationTests.GlobalTopic, []byte("global message"))
	time.Sleep(time.Second)
}

func sendMessagesOnIntraShardTopic(t *testing.T, nodesMap map[uint32][]*integrationTests.TestP2PNode) {
	t.Log("sending a message on intra shard topic")
	for _, nodes := range nodesMap {
		n := nodes[0]

		identifier := integrationTests.ShardTopic +
			n.ShardCoordinator.CommunicationIdentifier(n.ShardCoordinator.SelfId())
		nodes[0].Messenger.Broadcast(identifier, []byte("intra shard message"))
	}
	time.Sleep(time.Second)
}

func sendMessagesOnCrossShardTopic(t *testing.T, nodesMap map[uint32][]*integrationTests.TestP2PNode) {
	t.Log("sending messages on cross shard topics")

	for shardIdSrc, nodes := range nodesMap {
		n := nodes[0]

		for shardIdDest := range nodesMap {
			if shardIdDest == shardIdSrc {
				continue
			}

			identifier := integrationTests.ShardTopic +
				n.ShardCoordinator.CommunicationIdentifier(shardIdDest)
			nodes[0].Messenger.Broadcast(identifier, []byte("cross shard message"))
		}
	}
	time.Sleep(time.Second)
}

func testCounters(
	t *testing.T,
	nodesMap map[uint32][]*integrationTests.TestP2PNode,
	globalTopicMessagesCount int,
	intraTopicMessagesCount int,
	crossTopicMessagesCount int,
) {

	for _, nodes := range nodesMap {
		for _, n := range nodes {
			assert.Equal(t, globalTopicMessagesCount, n.CountGlobalMessages())
			assert.Equal(t, intraTopicMessagesCount, n.CountIntraShardMessages())
			assert.Equal(t, crossTopicMessagesCount, n.CountCrossShardMessages())
		}
	}
}
