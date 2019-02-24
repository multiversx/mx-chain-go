package mdns

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/integrationTests/peerDiscovery"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/stretchr/testify/assert"
)

var durationBootstrapingTime = time.Duration(time.Second * 2)
var durationTopicAnnounceTime = time.Duration(time.Second * 2)
var durationMsgRecieved = time.Duration(time.Second * 2)

func TestPeerDiscoveryAndMessageSending(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	ti := peerDiscovery.TestInitializer{}

	basePort := 23000
	noOfPeers := 20

	//Step 1. Create noOfPeers instances of messenger type
	peers := make([]p2p.Messenger, noOfPeers)

	for i := 0; i < noOfPeers; i++ {
		peers[i] = ti.CreateMessenger(context.Background(), basePort+i, p2p.PeerDiscoveryMdns)
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

	//Step 4. run the test for a couple of times as peer discovering ant topic announcing
	// are not deterministic and instant processes

	noOfTests := 5
	for i := 0; i < noOfTests; i++ {
		testResult := runTest(peers, i, "test topic")

		if testResult {
			return
		}
	}

	assert.Fail(t, "test failed. Discovery/message passing are not validated")
}

func runTest(peers []p2p.Messenger, testIndex int, topic string) bool {
	fmt.Printf("Running test %v\n", testIndex)

	testMessage := "test " + strconv.Itoa(testIndex)
	messageProcessors := make([]p2p.MessageProcessor, len(peers))

	chanDone := make(chan struct{})
	chanMessageProcessor := make(chan struct{}, len(peers))

	//add a new message processor for each messenger
	for i, peer := range peers {
		if peer.HasTopicValidator(topic) {
			peer.UnregisterMessageProcessor(topic)
		}

		mp := peerDiscovery.NewMessageProcessor(chanMessageProcessor, []byte(testMessage))

		messageProcessors[i] = mp
		err := peer.RegisterMessageProcessor(topic, mp)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
	}

	var msgReceived int32 = 0

	go func() {

		for {
			<-chanMessageProcessor
			atomic.AddInt32(&msgReceived, 1)

			if atomic.LoadInt32(&msgReceived) == int32(len(peers)) {
				chanDone <- struct{}{}
				return
			}
		}
	}()

	//write the message on topic
	peers[0].Broadcast(topic, []byte(testMessage))

	select {
	case <-chanDone:
		return true
	case <-time.After(durationMsgRecieved):
		fmt.Printf("timeout fetching all messages. Got %d from %d\n",
			atomic.LoadInt32(&msgReceived), len(peers))
		return false
	}
}
