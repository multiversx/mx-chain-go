package peerDisconnecting

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p/discovery/factory"
	"github.com/libp2p/go-libp2p-core/peer"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/assert"
)

var durationBootstrapingTime = 2 * time.Second
var randezVous = "elrondRandezVous"

func TestPeerDisconnectionWithOneAdvertiserWithShardingWithPrioBits(t *testing.T) {
	p2pConfig := config.P2PConfig{
		Node: config.NodeConfig{
			TargetPeerCount: 100,
		},
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled:              true,
			RefreshIntervalInSec: 1,
			RandezVous:           randezVous,
			InitialPeerList:      nil,
			Type:                 config.KadDhtVariantPrioBits,
		},
	}
	testPeerDisconnectionWithOneAdvertiser(t, p2pConfig)
}

func TestPeerDisconnectionWithOneAdvertiserWithShardingWithLists(t *testing.T) {
	p2pConfig := config.P2PConfig{
		Node: config.NodeConfig{
			TargetPeerCount: 100,
		},
		KadDhtPeerDiscovery: config.KadDhtPeerDiscoveryConfig{
			Enabled:              true,
			RefreshIntervalInSec: 1,
			RandezVous:           randezVous,
			InitialPeerList:      nil,
			Type:                 config.KadDhtVariantPrioBits,
		},
	}
	testPeerDisconnectionWithOneAdvertiser(t, p2pConfig)
}

func testPeerDisconnectionWithOneAdvertiser(t *testing.T, p2pConfig config.P2PConfig) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfPeers := 20
	netw := mocknet.New(context.Background())

	p2pConfigSeeder := p2pConfig
	//Step 1. Create advertiser
	advertiser, _ := libp2p.NewMemoryMessenger(
		context.Background(),
		netw,
		createPeerDiscoverer(p2pConfigSeeder),
	)
	p2pConfig.KadDhtPeerDiscovery.InitialPeerList = []string{integrationTests.GetConnectableAddress(advertiser)}

	//Step 2. Create noOfPeers instances of messenger type and call bootstrap
	peers := make([]p2p.Messenger, numOfPeers)
	for i := 0; i < numOfPeers; i++ {
		node, _ := libp2p.NewMemoryMessenger(
			context.Background(),
			netw,
			createPeerDiscoverer(p2pConfig),
		)
		peers[i] = node
	}

	//cleanup function that closes all messengers
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

	//link all peers so they can connect to each other
	_ = netw.LinkAll()

	//Step 3. Call bootstrap on all peers
	_ = advertiser.Bootstrap()
	for _, p := range peers {
		_ = p.Bootstrap()
	}
	integrationTests.WaitForBootstrapAndShowConnected(peers, durationBootstrapingTime)

	//Step 4. Disconnect one peer
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
		integrationTests.WaitForBootstrapAndShowConnected(peers, durationBootstrapingTime)
	}

	//Step 4.1. Test that the peer is disconnected
	for _, p := range peers {
		if p != disconnectedPeer {
			assert.Equal(t, numOfPeers-1, len(p.ConnectedPeers()))
		} else {
			assert.Equal(t, 0, len(p.ConnectedPeers()))
		}
	}

	//Step 5. Re-link and test connections
	fmt.Println("--- Re-linking ---")
	_ = netw.LinkAll()
	for i := 0; i < 5; i++ {
		integrationTests.WaitForBootstrapAndShowConnected(peers, durationBootstrapingTime)
	}

	//Step 5.1. Test that the peer is reconnected
	for _, p := range peers {
		assert.Equal(t, numOfPeers, len(p.ConnectedPeers()))
	}
}

func getPeerId(netMessenger p2p.Messenger) peer.ID {
	return peer.ID(netMessenger.ID().Bytes())
}

func createPeerDiscoverer(p2pConfig config.P2PConfig) p2p.PeerDiscoverer {
	peerDiscoveryFactory := factory.NewPeerDiscovererFactory(p2pConfig)
	peerDiscovery, _ := peerDiscoveryFactory.CreatePeerDiscoverer()

	return peerDiscovery
}
