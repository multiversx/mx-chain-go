package mdns

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/integrationTests/peerDiscovery"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/discovery"
	"github.com/stretchr/testify/assert"
)

var durationBootstrapingTime = time.Duration(time.Second * 2)
var durationTopicAnnounceTime = time.Duration(time.Second * 2)

func TestPeerDiscoveryAndMessageSending(t *testing.T) {
	// TODO: this test still generates a race condition due to: https://github.com/whyrusleeping/mdns/pull/4
	// This pull request doesn't have a correct solution yet and it is not merged. When this is fixed,
	// we can add the short condition back so this test would be run on the nightly build.
	// Remember to skip this test when the short flag is set
	t.Skip("this is not a short test")

	basePort := 26000
	noOfPeers := 20

	//Step 1. Create noOfPeers instances of messenger type
	peers := make([]p2p.Messenger, noOfPeers)

	for i := 0; i < noOfPeers; i++ {
		peers[i] = peerDiscovery.CreateMessenger(
			context.Background(),
			basePort+i,
			discovery.NewMdnsPeerDiscoverer(time.Second, "subnet"))
	}

	//Step 2. Call bootstrap to start the discovery process
	for _, peer := range peers {
		peer.Bootstrap()
	}

	//cleanup function that closes all messengers
	defer func() {
		for i := 0; i < noOfPeers; i++ {
			if peers[i] != nil {
				peers[i].Close()
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
		testResult := peerDiscovery.RunTest(peers, i, "test topic")

		if testResult {
			return
		}
	}

	assert.Fail(t, "test failed. Discovery/message passing are not validated")
}
