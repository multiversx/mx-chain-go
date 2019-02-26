package kadDht

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/integrationTests/peerDiscovery"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/stretchr/testify/assert"
)

var durationBootstrapingTime = time.Duration(time.Second * 2)
var durationTopicAnnounceTime = time.Duration(time.Second * 2)

func TestPeerDiscoveryAndMessageSending1Advertiser(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	tr := peerDiscovery.TestRunner{}

	basePort := 23000
	noOfPeers := 20

	//Step 1. Create advertiser
	advertiser := tr.CreateMessenger(context.Background(), basePort, p2p.PeerDiscoveryKadDht)
	basePort++
	advertiser.Bootstrap(time.Second, nil)

	//Step 2. Create noOfPeers instances of messenger type
	peers := make([]p2p.Messenger, noOfPeers)

	for i := 0; i < noOfPeers; i++ {
		peers[i] = tr.CreateMessenger(context.Background(), basePort+i, p2p.PeerDiscoveryMdns)
	}

	//Step 2. Call bootstrap to start the discovery process
	for _, peer := range peers {
		peer.Bootstrap(time.Second, []string{chooseNonCircuitAddress(advertiser.Addresses())})
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

func chooseNonCircuitAddress(addresses []string) string {
	for _, adr := range addresses {
		if strings.Contains(adr, "circuit") {
			continue
		}

		return adr
	}

	return ""
}
