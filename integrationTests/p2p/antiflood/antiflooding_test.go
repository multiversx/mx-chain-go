package antiflood

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

var durationBootstrapingTime = 2 * time.Second
var maxSize = 1 << 20 //1MB

// TestAntifloodWithMessagesFromTheSamePeer tests what happens if a peer decide to send a large number of messages
// all originating from its peer ID
// All directed peers should prevent the flooding to the rest of the network and process only a limited number of messages
func TestAntifloodWithMessagesFromTheSamePeer(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	peers, err := integrationTests.CreateFixedNetworkOf7Peers()
	assert.Nil(t, err)

	defer func() {
		integrationTests.ClosePeers(peers)
	}()

	//node 3 is connected to 0, 2, 4 and 6 (check integrationTests.CreateFixedNetworkOf7Peers function)
	//large number of broadcast messages from 3 might flood above mentioned peers but should not flood 5 and 7

	topic := "test_topic"
	broadcastMessageDuration := time.Second * 2
	maxMumProcessMessages := 5
	interceptors, err := createTopicsAndMockInterceptors(peers, topic, maxMumProcessMessages)
	assert.Nil(t, err)

	fmt.Println("bootstrapping nodes")
	time.Sleep(durationBootstrapingTime)

	flooderIdx := 3
	floodedIdxes := []int{0, 2, 4, 6}
	protectedIdexes := []int{5, 7}

	//flooder will deactivate its flooding mechanism as to be able to flood the network
	interceptors[flooderIdx].CountersMap = nil

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

	checkMessagesOnPeers(t, peers, interceptors, uint64(maxMumProcessMessages), floodedIdxes, protectedIdexes)
}

// TestAntifloodWithMessagesFromOtherPeers tests what happens if a peer decide to send a number of messages
// originating form other peer IDs. Since this is exceptionally hard to accomplish in integration tests because it needs
// 3-rd party library tweaking, the test is reduced to 10 peers generating 1 message through one peer that acts as a flooder
// All directed peers should prevent the flooding to the rest of the network and process only a limited number of messages
func TestAntifloodWithMessagesFromOtherPeers(t *testing.T) {
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
	maxMumProcessMessages := 5
	interceptors, err := createTopicsAndMockInterceptors(peers, topic, maxMumProcessMessages)
	assert.Nil(t, err)

	fmt.Println("bootstrapping nodes")
	time.Sleep(durationBootstrapingTime)

	flooderIdxes := []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	floodedIdxes := []int{1}
	protectedIdexes := []int{0}

	//flooders will deactivate their flooding mechanism as to be able to flood the network
	for _, idx := range flooderIdxes {
		interceptors[idx].CountersMap = nil
	}

	//generate a message from connected peers of the main flooder (peer 2)
	fmt.Println("flooding the network")
	for i := 3; i <= 13; i++ {
		peers[i].Broadcast(topic, []byte("floodMessage"))
	}
	time.Sleep(broadcastMessageDuration)

	checkMessagesOnPeers(t, peers, interceptors, uint64(maxMumProcessMessages), floodedIdxes, protectedIdexes)
}

func checkMessagesOnPeers(
	t *testing.T,
	peers []p2p.Messenger,
	interceptors []*messageProcessor,
	maxMumProcessMessages uint64,
	floodedIdxes []int,
	protectedIdexes []int,
) {
	checkFunctionForFloodedPeers := func(interceptor *messageProcessor) {
		assert.Equal(t, maxMumProcessMessages, interceptor.MessagesProcessed())
		//can not precisely determine how many message have been received
		assert.True(t, maxMumProcessMessages < interceptor.MessagesReceived())
	}
	checkFunctionForProtectedPeers := func(interceptor *messageProcessor) {
		assert.Equal(t, maxMumProcessMessages, interceptor.MessagesProcessed())
		assert.Equal(t, maxMumProcessMessages, interceptor.MessagesReceived())
	}

	fmt.Println("checking flooded peers")
	checkPeers(peers, interceptors, floodedIdxes, checkFunctionForFloodedPeers)
	fmt.Println("checking protected peers")
	checkPeers(peers, interceptors, protectedIdexes, checkFunctionForProtectedPeers)
}

func createTopicsAndMockInterceptors(peers []p2p.Messenger, topic string, maxNumMessages int) ([]*messageProcessor, error) {
	interceptors := make([]*messageProcessor, len(peers))

	for idx, p := range peers {
		err := p.CreateTopic(topic, true)
		if err != nil {
			return nil, fmt.Errorf("%w, pid: %s", err, p.ID())
		}

		cacherCfg := storageUnit.CacheConfig{Size: 100, Type: storageUnit.LRUCache, Shards: 1}
		antifloodPool, _ := storageUnit.NewCache(cacherCfg.Type, cacherCfg.Size, cacherCfg.Shards)

		interceptors[idx] = newMessageProcessor()
		interceptors[idx].CountersMap, _ = antiflood.NewQuotaFloodPreventer(antifloodPool, uint32(maxNumMessages), uint64(maxSize))
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
		fmt.Printf("%s got %d total messages and processed %d\n",
			peer.ID().Pretty(),
			interceptor.MessagesReceived(),
			interceptor.MessagesProcessed(),
		)

		checkFunction(interceptor)
	}
}
