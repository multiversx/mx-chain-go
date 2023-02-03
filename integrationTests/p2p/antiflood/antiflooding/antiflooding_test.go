package antiflooding

import (
	"fmt"
	"math"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/p2p/antiflood"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
)

// TestAntifloodWithMessagesFromTheSamePeer tests what happens if a peer decide to send a large number of messages
// all originating from its peer ID
// All directed peers should prevent the flooding to the rest of the network and process only a limited number of messages
func TestAntifloodWithNumMessagesFromTheSamePeer(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	_ = logger.SetLogLevel("*:INFO,p2p:ERROR")
	defer func() {
		_ = logger.SetLogLevel("*:INFO")
	}()

	peers, err := integrationTests.CreateFixedNetworkOf8Peers()
	assert.Nil(t, err)

	defer func() {
		integrationTests.ClosePeers(peers)
	}()

	//node 3 is connected to 0, 2, 4 and 6 (check integrationTests.CreateFixedNetworkOf7Peers function)
	//large number of broadcast messages from 3 might flood above-mentioned peers but should not flood 5 and 7

	topic := "test_topic"
	broadcastMessageDuration := time.Second * 2
	peerMaxNumProcessMessages := uint32(5)
	maxMessageSize := uint64(1 << 20) //1MB
	interceptors, err := antiflood.CreateTopicsAndMockInterceptors(
		peers,
		nil,
		topic,
		peerMaxNumProcessMessages,
		maxMessageSize,
	)
	assert.Nil(t, err)

	fmt.Println("bootstrapping nodes")
	time.Sleep(antiflood.DurationBootstrapingTime)

	flooderIdx := 3
	floodedIdxes := []int{0, 2, 4, 6}
	protectedIdexes := []int{5, 7}

	//flooder will deactivate its flooding mechanism as to be able to flood the network
	interceptors[flooderIdx].FloodPreventer = nil

	fmt.Println("flooding the network")
	isFlooding := atomic.Value{}
	isFlooding.Store(true)
	go antiflood.FloodTheNetwork(peers[flooderIdx], topic, &isFlooding, 10)
	time.Sleep(broadcastMessageDuration)

	isFlooding.Store(false)

	checkMessagesOnPeers(t, peers, interceptors, peerMaxNumProcessMessages, floodedIdxes, protectedIdexes)
}

// TestAntifloodWithMessagesFromOtherPeers tests what happens if a peer decide to send a number of messages
// originating form other peer IDs. Since this is exceptionally hard to accomplish in integration tests because it needs
// 3-rd party library tweaking, the test is reduced to 10 peers generating 1 message through one peer that acts as a flooder
// All directed peers should prevent the flooding to the rest of the network and process only a limited number of messages
func TestAntifloodWithNumMessagesFromOtherPeers(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	peers, err := integrationTests.CreateFixedNetworkOf14Peers()
	assert.Nil(t, err)

	defer func() {
		integrationTests.ClosePeers(peers)
	}()

	//peer 2 acts as a flooder that propagates 10 messages from 10 different peers. Peer 1 should prevent flooding to peer 0
	// (check integrationTests.CreateFixedNetworkOf14Peers function)
	topic := "test_topic"
	broadcastMessageDuration := time.Second * 2
	peerMaxNumProcessMessages := uint32(5)
	maxMessageSize := uint64(1 << 20) //1MB
	interceptors, err := antiflood.CreateTopicsAndMockInterceptors(
		peers,
		nil,
		topic,
		peerMaxNumProcessMessages,
		maxMessageSize,
	)
	assert.Nil(t, err)

	fmt.Println("bootstrapping nodes")
	time.Sleep(antiflood.DurationBootstrapingTime)

	flooderIdxes := []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	floodedIdxes := []int{1}
	protectedIdexes := []int{0}

	//flooders will deactivate their flooding mechanism as to be able to flood the network
	for _, idx := range flooderIdxes {
		interceptors[idx].FloodPreventer = nil
	}

	//generate a message from connected peers of the main flooder (peer 2)
	fmt.Println("flooding the network")
	for i := 3; i <= 13; i++ {
		peers[i].Broadcast(topic, []byte("floodMessage"))
	}
	time.Sleep(broadcastMessageDuration)

	checkMessagesOnPeers(t, peers, interceptors, peerMaxNumProcessMessages, floodedIdxes, protectedIdexes)
}

// TestAntifloodWithMessagesFromTheSamePeer tests what happens if a peer decide to send large messages
// all originating from its peer ID
// All directed peers should prevent the flooding to the rest of the network and process only a limited number of messages
func TestAntifloodWithLargeSizeMessagesFromTheSamePeer(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	_ = logger.SetLogLevel("*:INFO,p2p:ERROR")
	defer func() {
		_ = logger.SetLogLevel("*:INFO")
	}()

	peers, err := integrationTests.CreateFixedNetworkOf8Peers()
	assert.Nil(t, err)

	defer func() {
		integrationTests.ClosePeers(peers)
	}()

	//node 3 is connected to 0, 2, 4 and 6 (check integrationTests.CreateFixedNetworkOf7Peers function)
	//large number of broadcast messages from 3 might flood above-mentioned peers but should not flood 5 and 7

	topic := "test_topic"
	broadcastMessageDuration := time.Second * 2
	maxNumProcessMessages := uint32(math.MaxUint32)
	peerMaxMessageSize := uint64(1 << 10) //1KB
	interceptors, err := antiflood.CreateTopicsAndMockInterceptors(
		peers,
		nil,
		topic,
		maxNumProcessMessages,
		peerMaxMessageSize,
	)
	assert.Nil(t, err)

	fmt.Println("bootstrapping nodes")
	time.Sleep(antiflood.DurationBootstrapingTime)

	flooderIdx := 3
	floodedIdxes := []int{0, 2, 4, 6}
	protectedIdexes := []int{5, 7}

	//flooder will deactivate its flooding mechanism as to be able to flood the network
	interceptors[flooderIdx].FloodPreventer = nil

	fmt.Println("flooding the network")
	isFlooding := atomic.Value{}
	isFlooding.Store(true)
	go antiflood.FloodTheNetwork(peers[flooderIdx], topic, &isFlooding, peerMaxMessageSize+1)

	time.Sleep(broadcastMessageDuration)

	isFlooding.Store(false)

	checkMessagesOnPeers(t, peers, interceptors, 1, floodedIdxes, protectedIdexes)
}

func checkMessagesOnPeers(
	t *testing.T,
	peers []p2p.Messenger,
	interceptors []*antiflood.MessageProcessor,
	maxNumProcessMessages uint32,
	floodedIdxes []int,
	protectedIdexes []int,
) {
	checkFunctionForFloodedPeers := func(interceptor *antiflood.MessageProcessor) {
		assert.Equal(t, maxNumProcessMessages, interceptor.NumMessagesProcessed())
		//can not precisely determine how many message have been received
		assert.True(t, maxNumProcessMessages < interceptor.NumMessagesReceived())
	}
	checkFunctionForProtectedPeers := func(interceptor *antiflood.MessageProcessor) {
		assert.Equal(t, maxNumProcessMessages, interceptor.NumMessagesProcessed())
		assert.Equal(t, maxNumProcessMessages, interceptor.NumMessagesReceived())
	}

	fmt.Println("checking flooded peers")
	checkPeers(peers, interceptors, floodedIdxes, checkFunctionForFloodedPeers)
	fmt.Println("checking protected peers")
	checkPeers(peers, interceptors, protectedIdexes, checkFunctionForProtectedPeers)
}

func checkPeers(
	peers []p2p.Messenger,
	interceptors []*antiflood.MessageProcessor,
	indexes []int,
	checkFunction func(interceptor *antiflood.MessageProcessor),
) {

	for _, idx := range indexes {
		peer := peers[idx]
		interceptor := interceptors[idx]
		fmt.Printf("%s got %d (%s) total messages and processed %d (%s)\n",
			peer.ID().Pretty(),
			interceptor.NumMessagesReceived(),
			core.ConvertBytes(interceptor.SizeMessagesReceived()),
			interceptor.NumMessagesProcessed(),
			core.ConvertBytes(interceptor.SizeMessagesProcessed()),
		)

		checkFunction(interceptor)
	}
}
