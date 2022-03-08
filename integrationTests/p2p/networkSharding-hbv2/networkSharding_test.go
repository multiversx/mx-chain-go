package networkSharding

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
)

// todo remove this - tests only
type LatestKnownPeersHolder interface {
	GetLatestKnownPeers() map[string][]core.PeerID
}

var p2pBootstrapStepDelay = 2 * time.Second

func createDefaultConfig() config.P2PConfig {
	return config.P2PConfig{
		Node: config.NodeConfig{
			Port:                  "0",
			ConnectionWatcherType: "print",
		},
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
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
	p2pConfig := createDefaultConfig()
	p2pConfig.Sharding = config.ShardingConfig{
		TargetPeerCount:         12,
		MaxIntraShardValidators: 6,
		MaxCrossShardValidators: 1,
		MaxIntraShardObservers:  1,
		MaxCrossShardObservers:  1,
		MaxSeeders:              1,
		Type:                    p2p.ListsSharder,
		AdditionalConnections: config.AdditionalConnectionsConfig{
			MaxFullHistoryObservers: 1,
		},
	}

	testConnectionsInNetworkSharding(t, p2pConfig)
}

func testConnectionsInNetworkSharding(t *testing.T, p2pConfig config.P2PConfig) {
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
	initNodes(nodesMap)

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

func initNodes(nodesMap map[uint32][]*integrationTests.TestHeartbeatNode) {
	for _, nodes := range nodesMap {
		for _, n := range nodes {
			n.InitTestHeartbeatNode(0)
		}
	}
}

func createTestInterceptorForEachNode(nodesMap map[uint32][]*integrationTests.TestHeartbeatNode) {
	for _, nodes := range nodesMap {
		for _, n := range nodes {
			n.CreateTestInterceptors()
			n.CreateTxInterceptors()
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
			//assert.Equal(t, 0, len(n.Messenger.GetConnectedPeersInfo().UnknownPeers))
			assert.Equal(t, 1, len(n.Messenger.GetConnectedPeersInfo().Seeders))

			// todo remove this - tests only
			printDebugInfo(n)
		}
	}
}

func printDebugInfo(node *integrationTests.TestHeartbeatNode) {
	latestKnownPeers := node.CrossShardStatusProcessor.(LatestKnownPeersHolder).GetLatestKnownPeers()

	selfShard := node.ShardCoordinator.SelfId()
	selfPid := node.Messenger.ID()
	prettyPid := selfPid.Pretty()
	data := "----------\n"
	info := node.PeerShardMapper.GetPeerInfo(selfPid)
	data += fmt.Sprintf("PID: %s, shard: %d, PSM info: shard %d, type %s\n", prettyPid[len(prettyPid)-6:], node.ShardCoordinator.SelfId(), info.ShardID, info.PeerType)

	for topic, peers := range latestKnownPeers {
		data += fmt.Sprintf("topic: %s, connected crossshard pids:\n", topic)
		for _, peer := range peers {
			prettyPid = peer.Pretty()
			info = node.PeerShardMapper.GetPeerInfo(peer)
			data += fmt.Sprintf("   pid: %s, PSM info: shard %d, type %s\n", prettyPid[len(prettyPid)-6:], info.ShardID, info.PeerType)
		}
	}

	connectedPeersInfo := node.Messenger.GetConnectedPeersInfo()
	data += "connected peers from messenger...\n"
	if len(connectedPeersInfo.IntraShardValidators[selfShard]) > 0 {
		data += fmt.Sprintf("intraval %d:", len(connectedPeersInfo.IntraShardValidators[selfShard]))
		for _, val := range connectedPeersInfo.IntraShardValidators[selfShard] {
			data += fmt.Sprintf(" %s,", val[len(val)-6:])
		}
		data += "\n"
	}

	if len(connectedPeersInfo.IntraShardObservers[selfShard]) > 0 {
		data += fmt.Sprintf("intraobs %d:", len(connectedPeersInfo.IntraShardObservers[selfShard]))
		for _, obs := range connectedPeersInfo.IntraShardObservers[selfShard] {
			data += fmt.Sprintf(" %s,", obs[len(obs)-6:])
		}
		data += "\n"
	}

	if len(connectedPeersInfo.UnknownPeers) > 0 {
		data += fmt.Sprintf("unknown %d:", len(connectedPeersInfo.UnknownPeers))
		for _, unknown := range connectedPeersInfo.UnknownPeers {
			data += fmt.Sprintf(" %s,", unknown[len(unknown)-6:])
		}
		data += "\n"
	}

	data += "----------\n"
	println(data)
}
