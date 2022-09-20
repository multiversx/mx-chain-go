package peerDisconnecting

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/p2p"
	p2pConfig "github.com/ElrondNetwork/elrond-go/p2p/config"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/libp2p/go-libp2p-core/peer"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDefaultConfig() p2pConfig.P2PConfig {
	return p2pConfig.P2PConfig{
		Node: p2pConfig.NodeConfig{},
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

func TestPeerDisconnectionWithOneAdvertiserWithShardingWithLists(t *testing.T) {
	p2pCfg := createDefaultConfig()
	p2pCfg.Sharding = p2pConfig.ShardingConfig{
		TargetPeerCount:         100,
		MaxIntraShardValidators: 40,
		MaxCrossShardValidators: 40,
		MaxIntraShardObservers:  1,
		MaxCrossShardObservers:  1,
		MaxSeeders:              1,
		Type:                    p2p.ListsSharder,
		AdditionalConnections: p2pConfig.AdditionalConnectionsConfig{
			MaxFullHistoryObservers: 1,
		},
	}
	p2pCfg.Node.ThresholdMinConnectedPeers = 3

	testPeerDisconnectionWithOneAdvertiser(t, p2pCfg)
}

func testPeerDisconnectionWithOneAdvertiser(t *testing.T, p2pConfig p2pConfig.P2PConfig) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfPeers := 20
	netw := mocknet.New()

	p2pConfigSeeder := p2pConfig
	argSeeder := p2p.ArgsNetworkMessenger{
		ListenAddress:         p2p.ListenLocalhostAddrWithIp4AndTcp,
		P2pConfig:             p2pConfigSeeder,
		PreferredPeersHolder:  &p2pmocks.PeersHolderStub{},
		NodeOperationMode:     p2p.NormalOperation,
		Marshalizer:           &testscommon.MarshalizerMock{},
		SyncTimer:             &testscommon.SyncTimerStub{},
		PeersRatingHandler:    &p2pmocks.PeersRatingHandlerStub{},
		ConnectionWatcherType: p2p.ConnectionWatcherTypePrint,
	}
	// Step 1. Create advertiser
	advertiser, err := p2p.NewMockMessenger(argSeeder, netw)
	require.Nil(t, err)
	p2pConfig.KadDhtPeerDiscovery.InitialPeerList = []string{integrationTests.GetConnectableAddress(advertiser)}

	// Step 2. Create noOfPeers instances of messenger type and call bootstrap
	peers := make([]p2p.Messenger, numOfPeers)
	for i := 0; i < numOfPeers; i++ {
		arg := p2p.ArgsNetworkMessenger{
			ListenAddress:         p2p.ListenLocalhostAddrWithIp4AndTcp,
			P2pConfig:             p2pConfig,
			PreferredPeersHolder:  &p2pmocks.PeersHolderStub{},
			NodeOperationMode:     p2p.NormalOperation,
			Marshalizer:           &testscommon.MarshalizerMock{},
			SyncTimer:             &testscommon.SyncTimerStub{},
			PeersRatingHandler:    &p2pmocks.PeersRatingHandlerStub{},
			ConnectionWatcherType: p2p.ConnectionWatcherTypePrint,
		}
		node, errCreate := p2p.NewMockMessenger(arg, netw)
		require.Nil(t, errCreate)
		peers[i] = node
	}

	// cleanup function that closes all messengers
	defer func() {
		for i := 0; i < numOfPeers; i++ {
			if peers[i] != nil {
				_ = peers[i].Close()
			}
		}

		if advertiser != nil {
			_ = advertiser.Close()
		}
	}()

	// link all peers so they can connect to each other
	_ = netw.LinkAll()

	// Step 3. Call bootstrap on all peers
	_ = advertiser.Bootstrap()
	for _, p := range peers {
		_ = p.Bootstrap()
	}
	integrationTests.WaitForBootstrapAndShowConnected(peers, integrationTests.P2pBootstrapDelay)

	// Step 4. Disconnect one peer
	disconnectedPeer := peers[5]
	fmt.Printf("--- Diconnecting peer: %v ---\n", disconnectedPeer.ID().Pretty())
	_ = netw.UnlinkPeers(getPeerId(advertiser), getPeerId(disconnectedPeer))
	_ = netw.DisconnectPeers(getPeerId(advertiser), getPeerId(disconnectedPeer))
	_ = netw.DisconnectPeers(getPeerId(disconnectedPeer), getPeerId(advertiser))
	for _, p := range peers {
		if p != disconnectedPeer {
			_ = netw.UnlinkPeers(getPeerId(p), getPeerId(disconnectedPeer))
			_ = netw.DisconnectPeers(getPeerId(p), getPeerId(disconnectedPeer))
			_ = netw.DisconnectPeers(getPeerId(disconnectedPeer), getPeerId(p))
		}
	}
	for i := 0; i < 5; i++ {
		integrationTests.WaitForBootstrapAndShowConnected(peers, integrationTests.P2pBootstrapDelay)
	}

	// Step 4.1. Test that the peer is disconnected
	for _, p := range peers {
		if p != disconnectedPeer {
			assert.Equal(t, numOfPeers-1, len(p.ConnectedPeers()))
		} else {
			assert.Equal(t, 0, len(p.ConnectedPeers()))
		}
	}

	// Step 5. Re-link and test connections
	fmt.Println("--- Re-linking ---")
	_ = netw.LinkAll()
	for i := 0; i < 5; i++ {
		integrationTests.WaitForBootstrapAndShowConnected(peers, integrationTests.P2pBootstrapDelay)
	}

	// Step 5.1. Test that the peer is reconnected
	for _, p := range peers {
		assert.Equal(t, numOfPeers, len(p.ConnectedPeers()))
	}
}

func getPeerId(netMessenger p2p.Messenger) peer.ID {
	return peer.ID(netMessenger.ID().Bytes())
}
