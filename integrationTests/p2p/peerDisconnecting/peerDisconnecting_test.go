package peerDisconnecting

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/discovery"
	peer "github.com/libp2p/go-libp2p-peer"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/assert"
)

var durationBootstrapingTime = time.Duration(time.Second * 2)
var randezVous = "elrondRandezVous"

func TestPeerDisconnectionWithOneAdvertiser(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	noOfPeers := 20
	netw := mocknet.New(context.Background())

	//Step 1. Create advertiser
	advertiser, _ := libp2p.NewMemoryMessenger(
		context.Background(),
		netw,
		discovery.NewKadDhtPeerDiscoverer(time.Second, randezVous, nil),
	)

	//Step 2. Create noOfPeers instances of messenger type and call bootstrap
	peers := make([]p2p.Messenger, noOfPeers)
	for i := 0; i < noOfPeers; i++ {
		node, _ := libp2p.NewMemoryMessenger(
			context.Background(),
			netw,
			discovery.NewKadDhtPeerDiscoverer(
				time.Second,
				randezVous,
				[]string{chooseNonCircuitAddress(advertiser.Addresses())},
			),
		)
		peers[i] = node
	}

	//cleanup function that closes all messengers
	defer func() {
		for i := 0; i < noOfPeers; i++ {
			if peers[i] != nil {
				peers[i].Close()
			}
		}

		if advertiser != nil {
			advertiser.Close()
		}
	}()

	//link all peers so they can connect to each other
	netw.LinkAll()

	//Step 3. Call bootstrap on all peers
	advertiser.Bootstrap()
	for _, p := range peers {
		p.Bootstrap()
	}
	waitForBootstrapAndShowConnected(peers)

	//Step 4. Disconnect one peer
	disconnectedPeer := peers[5]
	fmt.Printf("--- Diconnecting peer: %v ---\n", disconnectedPeer.ID().Pretty())
	netw.UnlinkPeers(getPeerId(advertiser), getPeerId(disconnectedPeer))
	netw.DisconnectPeers(getPeerId(advertiser), getPeerId(disconnectedPeer))
	netw.DisconnectPeers(getPeerId(disconnectedPeer), getPeerId(advertiser))
	for _, p := range peers {
		if p != disconnectedPeer {
			netw.UnlinkPeers(getPeerId(p), getPeerId(disconnectedPeer))
			netw.DisconnectPeers(getPeerId(p), getPeerId(disconnectedPeer))
			netw.DisconnectPeers(getPeerId(disconnectedPeer), getPeerId(p))
		}
	}
	for i := 0; i < 5; i++ {
		waitForBootstrapAndShowConnected(peers)
	}

	//Step 4.1. Test that the peer is disconnected
	for _, p := range peers {
		if p != disconnectedPeer {
			assert.Equal(t, noOfPeers-1, len(p.ConnectedPeers()))
		} else {
			assert.Equal(t, 0, len(p.ConnectedPeers()))
		}
	}

	//Step 5. Re-link and test connections
	fmt.Println("--- Re-linking ---")
	netw.LinkAll()
	for i := 0; i < 5; i++ {
		waitForBootstrapAndShowConnected(peers)
	}

	//Step 5.1. Test that the peer is reconnected
	for _, p := range peers {
		assert.Equal(t, noOfPeers, len(p.ConnectedPeers()))
	}
}

func getPeerId(netMessenger p2p.Messenger) peer.ID {
	return peer.ID(netMessenger.ID().Bytes())
}

func waitForBootstrapAndShowConnected(peers []p2p.Messenger) {
	fmt.Printf("Waiting %v for peer discovery...\n", durationBootstrapingTime)
	time.Sleep(durationBootstrapingTime)

	fmt.Println("Connected peers:")
	for _, p := range peers {
		fmt.Printf("Peer %s is connected to %d peers\n", p.ID().Pretty(), len(p.ConnectedPeers()))
	}
}

func chooseNonCircuitAddress(addresses []string) string {
	for _, adr := range addresses {
		if strings.Contains(adr, "circuit") || strings.Contains(adr, "169.254") {
			continue
		}
		return adr
	}

	return ""
}
