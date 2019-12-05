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

// TestAntifloodWithMessagesFromTheSamePeer tests what happens if a peer decide to send a large number of transactions
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

	checkFunctionForFloodedPeers := func(interceptor *messageProcessor) {
		assert.Equal(t, uint64(maxMumProcessMessages), interceptor.MessagesProcessed())
		//can not precisely determine how many message have been received
		assert.True(t, uint64(maxMumProcessMessages) < interceptor.MessagesReceived())
	}
	checkFunctionForProtectedPeers := func(interceptor *messageProcessor) {
		assert.Equal(t, uint64(maxMumProcessMessages), interceptor.MessagesProcessed())
		assert.Equal(t, uint64(maxMumProcessMessages), interceptor.MessagesReceived())
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
		interceptors[idx].CountersMap, _ = antiflood.NewCountersMap(antifloodPool, maxNumMessages)
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
