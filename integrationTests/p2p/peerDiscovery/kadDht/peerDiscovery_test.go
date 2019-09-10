package kadDht

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/p2p/peerDiscovery"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
)

var durationBootstrapingTime = 2 * time.Second
var durationTopicAnnounceTime = 2 * time.Second

func TestPeerDiscoveryAndMessageSendingWithOneAdvertiser(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfPeers := 20

	//Step 1. Create advertiser
	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	//Step 2. Create numOfPeers instances of messenger type and call bootstrap
	peers := make([]p2p.Messenger, numOfPeers)

	for i := 0; i < numOfPeers; i++ {
		peers[i] = integrationTests.CreateMessengerWithKadDht(context.Background(),
			integrationTests.GetConnectableAddress(advertiser))

		_ = peers[i].Bootstrap()
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

	integrationTests.WaitForBootstrapAndShowConnected(peers, durationBootstrapingTime)

	//Step 3. Create a test topic, add receiving handlers
	createTestTopicAndWaitForAnnouncements(t, peers)

	//Step 4. run the test for a couple of times as peer discovering and topic announcing
	// are not deterministic nor instant processes

	numOfTests := 5
	for i := 0; i < numOfTests; i++ {
		testResult := peerDiscovery.RunTest(peers, i, "test topic")

		if testResult {
			return
		}
	}

	assert.Fail(t, "test failed. Discovery/message passing are not validated")
}

func TestPeerDiscoveryAndMessageSendingWithThreeAdvertisers(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfPeers := 20
	numOfAdvertisers := 3

	//Step 1. Create 3 advertisers and connect them together
	advertisers := make([]p2p.Messenger, numOfAdvertisers)
	advertisers[0] = integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertisers[0].Bootstrap()

	for idx := 1; idx < numOfAdvertisers; idx++ {
		advertisers[idx] = integrationTests.CreateMessengerWithKadDht(context.Background(),
			integrationTests.GetConnectableAddress(advertisers[0]))
		_ = advertisers[idx].Bootstrap()
	}

	//Step 2. Create numOfPeers instances of messenger type and call bootstrap
	peers := make([]p2p.Messenger, numOfPeers)

	for i := 0; i < numOfPeers; i++ {
		peers[i] = integrationTests.CreateMessengerWithKadDht(context.Background(),
			integrationTests.GetConnectableAddress(advertisers[i%numOfAdvertisers]))
		_ = peers[i].Bootstrap()
	}

	//cleanup function that closes all messengers
	defer func() {
		for i := 0; i < numOfPeers; i++ {
			if peers[i] != nil {
				_ = peers[i].Close()
			}
		}

		for i := 0; i < numOfAdvertisers; i++ {
			if advertisers[i] != nil {
				_ = advertisers[i].Close()
			}
		}
	}()

	integrationTests.WaitForBootstrapAndShowConnected(peers, durationBootstrapingTime)

	//Step 3. Create a test topic, add receiving handlers
	createTestTopicAndWaitForAnnouncements(t, peers)

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

func createTestTopicAndWaitForAnnouncements(t *testing.T, peers []p2p.Messenger) {
	for _, peer := range peers {
		err := peer.CreateTopic("test topic", true)
		if err != nil {
			assert.Fail(t, "test fail while creating topic")
		}
	}

	fmt.Printf("Waiting %v for topic announcement...\n", durationTopicAnnounceTime)
	time.Sleep(durationTopicAnnounceTime)
}
