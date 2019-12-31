package antiflood

import (
	"fmt"
	"math"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

var durationBootstrapingTime = 2 * time.Second

// TestAntifloodWithMessagesFromTheSamePeer tests what happens if a peer decide to send a large number of messages
// all originating from its peer ID
// All directed peers should prevent the flooding to the rest of the network and process only a limited number of messages
func TestAntifloodWithNumMessagesFromTheSamePeer(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	peers, err := integrationTests.CreateFixedNetworkOf8Peers()
	assert.Nil(t, err)

	defer func() {
		integrationTests.ClosePeers(peers)
	}()

	//node 3 is connected to 0, 2, 4 and 6 (check integrationTests.CreateFixedNetworkOf7Peers function)
	//large number of broadcast messages from 3 might flood above mentioned peers but should not flood 5 and 7

	topic := "test_topic"
	broadcastMessageDuration := time.Second * 2
	peerMaxNumProcessMessages := uint32(5)
	maxNumProcessMessages := uint32(math.MaxUint32)
	maxMessageSize := uint64(1 << 20) //1MB
	interceptors, err := createTopicsAndMockInterceptors(
		peers,
		nil,
		topic,
		peerMaxNumProcessMessages,
		maxMessageSize,
		maxNumProcessMessages,
		maxMessageSize,
	)
	assert.Nil(t, err)

	fmt.Println("bootstrapping nodes")
	time.Sleep(durationBootstrapingTime)

	flooderIdx := 3
	floodedIdxes := []int{0, 2, 4, 6}
	protectedIdexes := []int{5, 7}

	//flooder will deactivate its flooding mechanism as to be able to flood the network
	interceptors[flooderIdx].floodPreventer = nil

	fmt.Println("flooding the network")
	isFlooding := atomic.Value{}
	isFlooding.Store(true)
	go func() {
		for {
			peers[flooderIdx].Broadcast(topic, []byte("floodMessage"))

			if !isFlooding.Load().(bool) {
				return
			}
		}
	}()
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
	maxNumProcessMessages := uint32(math.MaxUint32)
	maxMessageSize := uint64(1 << 20) //1MB
	interceptors, err := createTopicsAndMockInterceptors(
		peers,
		nil,
		topic,
		peerMaxNumProcessMessages,
		maxMessageSize,
		maxNumProcessMessages,
		maxMessageSize,
	)
	assert.Nil(t, err)

	fmt.Println("bootstrapping nodes")
	time.Sleep(durationBootstrapingTime)

	flooderIdxes := []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	floodedIdxes := []int{1}
	protectedIdexes := []int{0}

	//flooders will deactivate their flooding mechanism as to be able to flood the network
	for _, idx := range flooderIdxes {
		interceptors[idx].floodPreventer = nil
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

	peers, err := integrationTests.CreateFixedNetworkOf8Peers()
	assert.Nil(t, err)

	defer func() {
		integrationTests.ClosePeers(peers)
	}()

	//node 3 is connected to 0, 2, 4 and 6 (check integrationTests.CreateFixedNetworkOf7Peers function)
	//large number of broadcast messages from 3 might flood above mentioned peers but should not flood 5 and 7

	topic := "test_topic"
	broadcastMessageDuration := time.Second * 2
	maxNumProcessMessages := uint32(math.MaxUint32)
	maxMessageSize := uint64(math.MaxUint64)
	peerMaxMessageSize := uint64(1 << 10) //1KB
	interceptors, err := createTopicsAndMockInterceptors(
		peers,
		nil,
		topic,
		maxNumProcessMessages,
		peerMaxMessageSize,
		maxNumProcessMessages,
		maxMessageSize,
	)
	assert.Nil(t, err)

	fmt.Println("bootstrapping nodes")
	time.Sleep(durationBootstrapingTime)

	flooderIdx := 3
	floodedIdxes := []int{0, 2, 4, 6}
	protectedIdexes := []int{5, 7}

	//flooder will deactivate its flooding mechanism as to be able to flood the network
	interceptors[flooderIdx].floodPreventer = nil

	fmt.Println("flooding the network")
	isFlooding := atomic.Value{}
	isFlooding.Store(true)
	go func() {
		for {
			peers[flooderIdx].Broadcast(topic, make([]byte, peerMaxMessageSize+1))

			if !isFlooding.Load().(bool) {
				return
			}
		}
	}()
	time.Sleep(broadcastMessageDuration)

	isFlooding.Store(false)

	checkMessagesOnPeers(t, peers, interceptors, 1, floodedIdxes, protectedIdexes)
}

func checkMessagesOnPeers(
	t *testing.T,
	peers []p2p.Messenger,
	interceptors []*messageProcessor,
	maxNumProcessMessages uint32,
	floodedIdxes []int,
	protectedIdexes []int,
) {
	checkFunctionForFloodedPeers := func(interceptor *messageProcessor) {
		assert.Equal(t, maxNumProcessMessages, interceptor.NumMessagesProcessed())
		//can not precisely determine how many message have been received
		assert.True(t, maxNumProcessMessages < interceptor.NumMessagesReceived())
	}
	checkFunctionForProtectedPeers := func(interceptor *messageProcessor) {
		assert.Equal(t, maxNumProcessMessages, interceptor.NumMessagesProcessed())
		assert.Equal(t, maxNumProcessMessages, interceptor.NumMessagesReceived())
	}

	fmt.Println("checking flooded peers")
	checkPeers(peers, interceptors, floodedIdxes, checkFunctionForFloodedPeers)
	fmt.Println("checking protected peers")
	checkPeers(peers, interceptors, protectedIdexes, checkFunctionForProtectedPeers)
}

func createTopicsAndMockInterceptors(
	peers []p2p.Messenger,
	blacklistHandlers []antiflood.QuotaStatusHandler,
	topic string,
	peerMaxNumMessages uint32,
	peerMaxSize uint64,
	maxNumMessages uint32,
	maxSize uint64,
) ([]*messageProcessor, error) {

	interceptors := make([]*messageProcessor, len(peers))

	for idx, p := range peers {
		err := p.CreateTopic(topic, true)
		if err != nil {
			return nil, fmt.Errorf("%w, pid: %s", err, p.ID())
		}

		cacherCfg := storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache, Shards: 1}
		antifloodPool, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

		interceptors[idx] = newMessageProcessor()
		statusHandlers := []antiflood.QuotaStatusHandler{&nilQuotaStatusHandler{}}
		if len(blacklistHandlers) == len(peers) {
			statusHandlers = append(statusHandlers, blacklistHandlers[idx])
		}
		interceptors[idx].floodPreventer, _ = antiflood.NewQuotaFloodPreventer(
			antifloodPool,
			statusHandlers,
			peerMaxNumMessages,
			peerMaxSize,
			maxNumMessages,
			maxSize,
		)
		err = p.RegisterMessageProcessor(topic, interceptors[idx])
		if err != nil {
			return nil, fmt.Errorf("%w, pid: %s", err, p.ID())
		}
	}

	return interceptors, nil
}

func checkPeers(
	peers []p2p.Messenger,
	interceptors []*messageProcessor,
	indexes []int,
	checkFunction func(interceptor *messageProcessor),
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
