package kadDht

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/integrationTests/peerDiscovery"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/libp2p/discovery"
	"github.com/stretchr/testify/assert"
)

var durationBootstrapingTime = time.Duration(time.Second * 2)
var durationTopicAnnounceTime = time.Duration(time.Second * 2)
var randezVous = "elrondRandezVous"

func TestPeerDiscoveryAndMessageSendingWithOneAdvertiser(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	basePort := 25000
	noOfPeers := 20

	//Step 1. Create advertiser
	advertiser := peerDiscovery.CreateMessenger(
		context.Background(),
		basePort,
		discovery.NewKadDhtPeerDiscoverer(time.Second, randezVous, nil))
	basePort++
	advertiser.Bootstrap()

	//Step 2. Create noOfPeers instances of messenger type and call bootstrap
	peers := make([]p2p.Messenger, noOfPeers)

	for i := 0; i < noOfPeers; i++ {
		kadDht := discovery.NewKadDhtPeerDiscoverer(
			time.Second,
			randezVous,
			[]string{chooseNonCircuitAddress(advertiser.Addresses())})

		peers[i] = peerDiscovery.CreateMessenger(
			context.Background(),
			basePort+i,
			kadDht)

		peers[i].Bootstrap()
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

	waitForBootstrapAndShowConnected(peers)

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

func TestPeerDiscoveryAndMessageSendingWithThreeAdvertisers(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	basePort := 25100
	noOfPeers := 20
	noOfAdvertisers := 3

	//Step 1. Create 3 advertisers and connect them together
	advertisers := make([]p2p.Messenger, noOfAdvertisers)
	for i := 0; i < noOfAdvertisers; i++ {
		if i == 0 {
			advertisers[i] = peerDiscovery.CreateMessenger(
				context.Background(),
				basePort,
				discovery.NewKadDhtPeerDiscoverer(time.Second, randezVous, nil))
		} else {
			advertisers[i] = peerDiscovery.CreateMessenger(
				context.Background(),
				basePort,
				discovery.NewKadDhtPeerDiscoverer(
					time.Second,
					randezVous,
					[]string{chooseNonCircuitAddress(advertisers[0].Addresses())}))
		}

		basePort++
		advertisers[i].Bootstrap()
	}

	//Step 2. Create noOfPeers instances of messenger type and call bootstrap
	peers := make([]p2p.Messenger, noOfPeers)

	for i := 0; i < noOfPeers; i++ {
		kadDht := discovery.NewKadDhtPeerDiscoverer(
			time.Second,
			randezVous,
			[]string{chooseNonCircuitAddress(advertisers[i%noOfAdvertisers].Addresses())})

		peers[i] = peerDiscovery.CreateMessenger(
			context.Background(),
			basePort+i,
			kadDht)

		peers[i].Bootstrap()
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

	waitForBootstrapAndShowConnected(peers)

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

func waitForBootstrapAndShowConnected(peers []p2p.Messenger) {
	fmt.Printf("Waiting %v for peer discovery...\n", durationBootstrapingTime)
	time.Sleep(durationBootstrapingTime)

	fmt.Println("Connected peers:")
	for _, peer := range peers {
		fmt.Printf("Peer %s is connected to %d peers\n", peer.ID().Pretty(), len(peer.ConnectedPeers()))
	}
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

func chooseNonCircuitAddress(addresses []string) string {
	for _, adr := range addresses {
		if strings.Contains(adr, "circuit") || strings.Contains(adr, "169.254") {
			continue
		}

		return adr
	}

	return ""
}
