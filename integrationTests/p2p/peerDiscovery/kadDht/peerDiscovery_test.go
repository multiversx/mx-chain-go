package kadDht

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/p2p/peerDiscovery"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
)

var durationTopicAnnounceTime = 2 * time.Second

func TestPeerDiscoveryAndMessageSendingWithOneAdvertiser(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfPeers := 20

	//Step 1. Create advertiser
	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap(0)

	//Step 2. Create numOfPeers instances of messenger type and call bootstrap
	peers := make([]p2p.Messenger, numOfPeers)

	for i := 0; i < numOfPeers; i++ {
		peers[i] = integrationTests.CreateMessengerWithKadDht(integrationTests.GetConnectableAddress(advertiser))

		_ = peers[i].Bootstrap(0)
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

	integrationTests.WaitForBootstrapAndShowConnected(peers, integrationTests.P2pBootstrapDelay)

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
	advertisers[0] = integrationTests.CreateMessengerWithKadDht("")
	_ = advertisers[0].Bootstrap(0)

	for idx := 1; idx < numOfAdvertisers; idx++ {
		advertisers[idx] = integrationTests.CreateMessengerWithKadDht(integrationTests.GetConnectableAddress(advertisers[0]))
		_ = advertisers[idx].Bootstrap(0)
	}

	//Step 2. Create numOfPeers instances of messenger type and call bootstrap
	peers := make([]p2p.Messenger, numOfPeers)

	for i := 0; i < numOfPeers; i++ {
		peers[i] = integrationTests.CreateMessengerWithKadDht(integrationTests.GetConnectableAddress(advertisers[i%numOfAdvertisers]))
		_ = peers[i].Bootstrap(0)
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

	integrationTests.WaitForBootstrapAndShowConnected(peers, integrationTests.P2pBootstrapDelay)

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

func TestPeerDiscoveryAndMessageSendingWithOneAdvertiserAndProtocolID(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap(0)

	protocolID1 := "/erd/kad/1.0.0"
	protocolID2 := "/amony/kad/0.0.0"

	peer1 := integrationTests.CreateMessengerWithKadDhtAndProtocolID(
		integrationTests.GetConnectableAddress(advertiser),
		protocolID1,
	)
	peer2 := integrationTests.CreateMessengerWithKadDhtAndProtocolID(
		integrationTests.GetConnectableAddress(advertiser),
		protocolID1,
	)
	peer3 := integrationTests.CreateMessengerWithKadDhtAndProtocolID(
		integrationTests.GetConnectableAddress(advertiser),
		protocolID2,
	)

	peers := []p2p.Messenger{peer1, peer2, peer3}

	for _, peer := range peers {
		_ = peer.Bootstrap(0)
	}

	//cleanup function that closes all messengers
	defer func() {
		for i := 0; i < len(peers); i++ {
			if peers[i] != nil {
				_ = peers[i].Close()
			}
		}

		if advertiser != nil {
			_ = advertiser.Close()
		}
	}()

	integrationTests.WaitForBootstrapAndShowConnected(peers, integrationTests.P2pBootstrapDelay)

	createTestTopicAndWaitForAnnouncements(t, peers)

	topic := "test topic"
	message := []byte("message")
	messageProcessors := assignProcessors(peers, topic)

	peer1.Broadcast(topic, message)
	time.Sleep(time.Second * 2)

	assert.Equal(t, message, messageProcessors[0].GetLastMessage())
	assert.Equal(t, message, messageProcessors[1].GetLastMessage())
	assert.Nil(t, messageProcessors[2].GetLastMessage())

	assert.Equal(t, 2, len(peer1.ConnectedPeers()))
	assert.Equal(t, 2, len(peer2.ConnectedPeers()))
	assert.Equal(t, 1, len(peer3.ConnectedPeers()))
}

func assignProcessors(peers []p2p.Messenger, topic string) []*peerDiscovery.SimpleMessageProcessor {
	processors := make([]*peerDiscovery.SimpleMessageProcessor, 0, len(peers))
	for _, peer := range peers {
		proc := &peerDiscovery.SimpleMessageProcessor{}
		processors = append(processors, proc)

		err := peer.RegisterMessageProcessor(topic, "test", proc)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	return processors
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
