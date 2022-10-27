package networkSharding

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/p2p"
	p2pConfig "github.com/ElrondNetwork/elrond-go/p2p/config"
	"github.com/stretchr/testify/assert"
)

var p2pBootstrapStepDelay = 2 * time.Second

func createDefaultConfig() p2pConfig.P2PConfig {
	return p2pConfig.P2PConfig{
		Node: p2pConfig.NodeConfig{
			Port: "0",
		},
		KadDhtPeerDiscovery: p2pConfig.KadDhtPeerDiscoveryConfig{
			Enabled:                          true,
			Type:                             "optimized",
			RefreshIntervalInSec:             1,
			RoutingTableRefreshIntervalInSec: 1,
			ProtocolID:                       "/erd/kad/1.0.0",
			InitialPeerList:                  nil,
			BucketSize:                       100,
		},
	}
}

func TestConnectionsInNetworkShardingWithShardingWithLists(t *testing.T) {
	p2pCfg := createDefaultConfig()
	p2pCfg.Sharding = p2pConfig.ShardingConfig{
		TargetPeerCount:         12,
		MaxIntraShardValidators: 6,
		MaxCrossShardValidators: 1,
		MaxIntraShardObservers:  1,
		MaxCrossShardObservers:  1,
		MaxSeeders:              1,
		Type:                    p2p.ListsSharder,
		AdditionalConnections: p2pConfig.AdditionalConnectionsConfig{
			MaxFullHistoryObservers: 1,
		},
	}

	testConnectionsInNetworkSharding(t, p2pCfg)
}

func testConnectionsInNetworkSharding(t *testing.T, p2pConfig p2pConfig.P2PConfig) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodesPerShard := 8
	numMetaNodes := 8
	numObserversOnShard := 2
	numShards := 2
	consensusGroupSize := 2

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()
	seedAddress := integrationTests.GetConnectableAddress(advertiser)

	p2pConfig.KadDhtPeerDiscovery.InitialPeerList = []string{seedAddress}

	// create map of shard - testHeartbeatNodes for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithTestHeartbeatNode(
		nodesPerShard,
		numMetaNodes,
		numShards,
		consensusGroupSize,
		numMetaNodes,
		numObserversOnShard,
		p2pConfig,
	)

	defer func() {
		stopNodes(advertiser, nodesMap)
	}()

	createTestInterceptorForEachNode(nodesMap)

	time.Sleep(time.Second * 2)

	startNodes(nodesMap)

	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(p2pBootstrapStepDelay)

	for i := 0; i < 5; i++ {
		fmt.Println("\n" + integrationTests.MakeDisplayTableForHeartbeatNodes(nodesMap))

		time.Sleep(time.Second)
	}

	fmt.Println("Initializing nodes components...")
	initNodes(t, nodesMap)

	for i := 0; i < 5; i++ {
		fmt.Println("\n" + integrationTests.MakeDisplayTableForHeartbeatNodes(nodesMap))

		time.Sleep(time.Second)
	}

	sendMessageOnGlobalTopic(nodesMap)
	sendMessagesOnIntraShardTopic(nodesMap)
	sendMessagesOnCrossShardTopic(nodesMap)

	for i := 0; i < 5; i++ {
		fmt.Println("\n" + integrationTests.MakeDisplayTableForHeartbeatNodes(nodesMap))

		time.Sleep(time.Second)
	}

	testCounters(t, nodesMap, 1, 1, numShards*2)
	testUnknownSeederPeers(t, nodesMap)
}

func stopNodes(advertiser p2p.Messenger, nodesMap map[uint32][]*integrationTests.TestHeartbeatNode) {
	_ = advertiser.Close()
	for _, nodes := range nodesMap {
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}
}

func startNodes(nodesMap map[uint32][]*integrationTests.TestHeartbeatNode) {
	for _, nodes := range nodesMap {
		for _, n := range nodes {
			_ = n.Messenger.Bootstrap()
		}
	}
}

func initNodes(tb testing.TB, nodesMap map[uint32][]*integrationTests.TestHeartbeatNode) {
	for _, nodes := range nodesMap {
		for _, n := range nodes {
			n.InitTestHeartbeatNode(tb, 0)
		}
	}
}

func createTestInterceptorForEachNode(nodesMap map[uint32][]*integrationTests.TestHeartbeatNode) {
	for _, nodes := range nodesMap {
		for _, n := range nodes {
			n.CreateTestInterceptors()
		}
	}
}

func sendMessageOnGlobalTopic(nodesMap map[uint32][]*integrationTests.TestHeartbeatNode) {
	fmt.Println("sending a message on global topic")
	nodesMap[0][0].Messenger.Broadcast(integrationTests.GlobalTopic, []byte("global message"))
	time.Sleep(time.Second)
}

func sendMessagesOnIntraShardTopic(nodesMap map[uint32][]*integrationTests.TestHeartbeatNode) {
	fmt.Println("sending a message on intra shard topic")
	for _, nodes := range nodesMap {
		n := nodes[0]

		identifier := integrationTests.ShardTopic +
			n.ShardCoordinator.CommunicationIdentifier(n.ShardCoordinator.SelfId())
		nodes[0].Messenger.Broadcast(identifier, []byte("intra shard message"))
	}
	time.Sleep(time.Second)
}

func sendMessagesOnCrossShardTopic(nodesMap map[uint32][]*integrationTests.TestHeartbeatNode) {
	fmt.Println("sending messages on cross shard topics")

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
	nodesMap map[uint32][]*integrationTests.TestHeartbeatNode,
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

func testUnknownSeederPeers(
	t *testing.T,
	nodesMap map[uint32][]*integrationTests.TestHeartbeatNode,
) {

	for _, nodes := range nodesMap {
		for _, n := range nodes {
			assert.Equal(t, 0, len(n.Messenger.GetConnectedPeersInfo().UnknownPeers))
			assert.Equal(t, 1, len(n.Messenger.GetConnectedPeersInfo().Seeders))
		}
	}
}
