package kadDht

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/integrationTests/peerDiscovery"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/stretchr/testify/assert"
)

func TestPeerDiscoveryAndMessageSending3Advertisers(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	tr := peerDiscovery.TestRunner{}

	basePort := 23000
	noOfPeers := 20
	noOfAdvertisers := 3

	//Step 1. Create 3 advertisers and connect them together
	advertisers := make([]p2p.Messenger, noOfAdvertisers)
	for i := 0; i < noOfAdvertisers; i++ {
		advertisers[i] = tr.CreateMessenger(context.Background(), basePort, p2p.PeerDiscoveryKadDht)
		basePort++
		if i > 0 {
			advertisers[i].Bootstrap(time.Second, []string{chooseNonCircuitAddress(advertisers[0].Addresses())})
		}
	}

	//Step 2. Create noOfPeers instances of messenger type
	peers := make([]p2p.Messenger, noOfPeers)

	for i := 0; i < noOfPeers; i++ {
		peers[i] = tr.CreateMessenger(context.Background(), basePort+i, p2p.PeerDiscoveryMdns)
	}

	//Step 2. Call bootstrap to start the discovery process
	for i, peer := range peers {
		peer.Bootstrap(time.Second, []string{chooseNonCircuitAddress(advertisers[i%noOfAdvertisers].Addresses())})
	}

	//cleanup function that closes all messengers
	defer func() {
		for i := 0; i < noOfPeers; i++ {
			if peers[i] != nil {
				peers[i].Close()
			}
		}

		for i := 0; i < noOfAdvertisers; i++ {
			if advertisers[i] != nil {
				advertisers[i].Close()
			}
		}
	}()

	fmt.Printf("Waiting %v for peer discovery...\n", durationBootstrapingTime)
	time.Sleep(durationBootstrapingTime)

	fmt.Println("Connected peers:")
	for _, peer := range peers {
		fmt.Printf("Peer %s is connected to %d peers\n", peer.ID().Pretty(), len(peer.ConnectedPeers()))
	}

	//Step 3. Create a test topic, add receiving handlers
	for _, peer := range peers {
		err := peer.CreateTopic("test topic", true)
		if err != nil {
			assert.Fail(t, "test fail while creating topic")
		}
	}

	fmt.Printf("Waiting %v for topic announcement...\n", durationTopicAnnounceTime)
	time.Sleep(durationTopicAnnounceTime)

	//Step 4. run the test for a couple of times as peer discovering and topic announcing
	// are not deterministic nor instant processes

	noOfTests := 5
	for i := 0; i < noOfTests; i++ {
		testResult := tr.RunTest(peers, i, "test topic")

		if testResult {
			return
		}
	}

	assert.Fail(t, "test failed. Discovery/message passing are not validated")
}
